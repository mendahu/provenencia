// Package raster is the one bounded image codec: it decodes only allowlisted
// formats within a pixel budget, and re-encodes. Thumbnails and staged images
// both use it.
//
// Security: golang.org/x/image decoders (bmp, tiff, webp) parse fully
// untrusted input here and have a history of memory-exhaustion CVEs.
// Keep that dependency current (go list -m -u golang.org/x/image).
package raster

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"runtime"
	"slices"
	"strings"

	_ "golang.org/x/image/bmp"  // register bmp decoder
	_ "golang.org/x/image/tiff" // register tiff decoder
	_ "golang.org/x/image/webp" // register webp decoder
)

func init() {
	// Keep stdlib decoders linked for image.Decode.
	_ = jpeg.Decode
	_ = png.Decode
	_ = gif.Decode
}

// DefaultQuality is used when Options.Quality is zero.
const DefaultQuality = 85

// MaxEdgeLimit is the largest Options.MaxEdge EncodeJPEG accepts. It bounds
// output allocations even if callers become externally parameterized.
const MaxEdgeLimit = 4096

// MaxDecodePixels caps declared width*height before full decode. Decoders
// allocate pixel buffers from header dimensions, so this is the primary
// defense against decompression bombs (see image package security notes).
const MaxDecodePixels = 64_000_000 // 64 MP ≈ 256 MiB RGBA

// ErrTooLarge means the image declares more pixels than MaxDecodePixels
// allows, so it was rejected before full decode.
var ErrTooLarge = errors.New("raster: image dimensions exceed decode budget")

// ErrUnsupportedFormat means the bytes are not one of the allowlisted
// raster formats, regardless of the recorded media type.
var ErrUnsupportedFormat = errors.New("raster: unsupported image format")

var supportedMIME = map[string]struct{}{
	"image/jpeg": {},
	"image/jpg":  {},
	"image/png":  {},
	"image/gif":  {},
	"image/bmp":  {},
	"image/tiff": {},
	"image/tif":  {},
	"image/webp": {},
}

// supportedFormats are image.DecodeConfig format names allowed to fully
// decode. Keep in sync with the decoders registered by this package.
var supportedFormats = map[string]struct{}{
	"jpeg": {},
	"png":  {},
	"gif":  {},
	"bmp":  {},
	"tiff": {},
	"webp": {},
}

// Transform mutates or replaces an image in the encode pipeline (after scale).
// Add grayscale, sharpen, etc. as named Transform funcs later without changing EncodeJPEG.
type Transform func(image.Image) image.Image

// Options parameterizes raster encoding. Zero Quality means DefaultQuality.
type Options struct {
	MaxEdge    int // longest output edge in pixels; must be > 0
	Quality    int // JPEG quality 1–100; 0 → DefaultQuality (EncodeJPEG only)
	Transforms []Transform
	// Formats limits the image.DecodeConfig format names that may decode
	// (for example "png", "jpeg"). Nil allows every supported format.
	Formats []string
}

// Supported reports whether mediaType is an allowlisted raster format.
func Supported(mediaType string) bool {
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if i := strings.IndexByte(mediaType, ';'); i >= 0 {
		mediaType = strings.TrimSpace(mediaType[:i])
	}
	_, ok := supportedMIME[mediaType]
	return ok
}

// EncodeJPEG decodes src, scales to MaxEdge, applies Transforms, and encodes JPEG.
// Before full decode it checks the sniffed format against the package allowlist
// (ErrUnsupportedFormat) and the declared pixel count against MaxDecodePixels
// (ErrTooLarge), so header-declared decompression bombs never allocate.
func EncodeJPEG(src []byte, opts Options) ([]byte, error) {
	if err := checkMaxEdge(opts); err != nil {
		return nil, err
	}
	quality := opts.Quality
	if quality == 0 {
		quality = DefaultQuality
	}
	if quality < 1 || quality > 100 {
		return nil, fmt.Errorf("raster: Quality must be 1–100 or 0 for default")
	}
	img, err := prepare(src, opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodePNG decodes src with the same format allowlist and pixel budget as
// EncodeJPEG, scales to MaxEdge, applies Transforms, and encodes PNG. Quality
// is ignored.
func EncodePNG(src []byte, opts Options) ([]byte, error) {
	if err := checkMaxEdge(opts); err != nil {
		return nil, err
	}
	img, err := prepare(src, opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func checkMaxEdge(opts Options) error {
	if opts.MaxEdge < 1 {
		return fmt.Errorf("raster: MaxEdge must be positive")
	}
	if opts.MaxEdge > MaxEdgeLimit {
		return fmt.Errorf("raster: MaxEdge must be at most %d", MaxEdgeLimit)
	}
	return nil
}

// decode refuses a format outside formats (or outside supportedFormats) and an
// image declaring more than MaxDecodePixels before any full decode.
func decode(src []byte, formats []string) (image.Image, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	if _, ok := supportedFormats[format]; !ok {
		return nil, ErrUnsupportedFormat
	}
	if formats != nil && !slices.Contains(formats, format) {
		return nil, ErrUnsupportedFormat
	}
	if cfg.Width < 1 || cfg.Height < 1 ||
		int64(cfg.Width)*int64(cfg.Height) > MaxDecodePixels {
		return nil, ErrTooLarge
	}
	img, _, err := image.Decode(bytes.NewReader(src))
	return img, err
}

// prepare decodes src, scales it to MaxEdge, and applies Transforms. It does
// not check options.
func prepare(src []byte, opts Options) (image.Image, error) {
	img, err := decode(src, opts.Formats)
	if err != nil {
		return nil, err
	}
	img = scaleToMaxEdge(img, opts.MaxEdge)
	for _, t := range opts.Transforms {
		if t != nil {
			img = t(img)
		}
	}
	return img, nil
}

// decodeSem bounds concurrent read+decode+encode work so parallel callers
// cannot multiply peak image-decode memory.
var decodeSem = make(chan struct{}, decodeSlots())

func decodeSlots() int {
	n := runtime.GOMAXPROCS(0)
	if n > 4 {
		n = 4
	}
	if n < 1 {
		n = 1
	}
	return n
}

// Slot waits for a decode slot and returns the function that frees it. Hold it
// across the raw read, decode, and encode.
func Slot() (release func()) {
	decodeSem <- struct{}{}
	return func() { <-decodeSem }
}

func scaleToMaxEdge(src image.Image, maxEdge int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 1 || h < 1 {
		return src
	}
	long := w
	if h > long {
		long = h
	}
	if long <= maxEdge {
		return src
	}
	nw := w * maxEdge / long
	nh := h * maxEdge / long
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		sy := b.Min.Y + y*h/nh
		for x := 0; x < nw; x++ {
			sx := b.Min.X + x*w/nw
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
