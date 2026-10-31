package raster

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/image/tiff"

	"github.com/mendahu/provenencia/core/raster/rastertest"
)

// fakeConfigCalls counts calls to the fake format's config func.
var fakeConfigCalls atomic.Int32

func init() {
	// A registered-but-not-allowlisted format to exercise ErrUnsupportedFormat.
	image.RegisterFormat("fake", "FAKEIMG!",
		func(io.Reader) (image.Image, error) { return nil, errors.New("fake: no decode") },
		func(io.Reader) (image.Config, error) {
			fakeConfigCalls.Add(1)
			return image.Config{Width: 1, Height: 1}, nil
		},
	)
}

// gifHeaderOnly returns a GIF89a header declaring a w×h logical screen.
func gifHeaderOnly(w, h uint16) []byte {
	b := []byte("GIF89a")
	b = append(b, byte(w), byte(w>>8), byte(h), byte(h>>8), 0, 0, 0)
	return b
}

// goFixture reads a file from the Go distribution's image testdata copied
// under testdata/third_party/go.
func goFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "third_party", "go", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestEncodeJPEGRejects(t *testing.T) {
	tests := []struct {
		name    string
		src     func(t *testing.T) []byte
		opts    Options
		wantIs  error
		wantErr bool
	}{
		{
			name:   "gif bomb 65535x65535 header",
			src:    func(*testing.T) []byte { return gifHeaderOnly(65535, 65535) },
			opts:   Options{MaxEdge: 256},
			wantIs: ErrTooLarge,
		},
		{
			name:   "png bomb 100000x100000 header",
			src:    func(t *testing.T) []byte { return rastertest.HeaderOnlyPNG(t, 100000, 100000) },
			opts:   Options{MaxEdge: 256},
			wantIs: ErrTooLarge,
		},
		{
			name:   "png just over pixel budget",
			src:    func(t *testing.T) []byte { return rastertest.HeaderOnlyPNG(t, 8001, 8000) },
			opts:   Options{MaxEdge: 256},
			wantIs: ErrTooLarge,
		},
		{
			name:   "registered but not allowlisted format",
			src:    func(*testing.T) []byte { return []byte("FAKEIMG!....") },
			opts:   Options{MaxEdge: 256},
			wantIs: ErrUnsupportedFormat,
		},
		{
			name:    "MaxEdge above limit",
			src:     tinyPNG,
			opts:    Options{MaxEdge: MaxEdgeLimit + 1},
			wantErr: true,
		},
		{
			name:   "garbage bytes",
			src:    func(*testing.T) []byte { return []byte{0, 1, 2, 3} },
			opts:   Options{MaxEdge: 256},
			wantIs: ErrUnsupportedFormat,
		},
		{
			name:   "jpeg header 65535x65535 over even the reduced budget",
			src:    func(*testing.T) []byte { return rastertest.JPEGHeader(65535, 65535, 1, nil) },
			opts:   Options{MaxEdge: 256},
			wantIs: ErrTooLarge,
		},
		{
			name: "oversized jpeg cut at half",
			src: func(t *testing.T) []byte {
				b := rastertest.OversizedJPEG(t)
				return b[:len(b)/2]
			},
			opts:   Options{MaxEdge: 256},
			wantIs: ErrUndecodable,
		},
		{
			name: "truncated jpeg within budget",
			src: func(t *testing.T) []byte {
				b := goFixture(t, "video-001.jpeg")
				return b[:len(b)/2]
			},
			opts:   Options{MaxEdge: 256},
			wantIs: ErrUndecodable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := EncodeJPEG(tt.src(t), tt.opts)
			if err == nil {
				t.Fatalf("expected error, got %d bytes", len(out))
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Fatalf("got %v, want %v", err, tt.wantIs)
			}
		})
	}
}

func TestEncodeJPEGOversized(t *testing.T) {
	src := rastertest.OversizedJPEG(t)
	for _, tc := range []struct{ edge, w, h int }{{256, 256, 255}, {2048, 1001, 1000}} {
		start := time.Now()
		out, err := EncodeJPEG(src, Options{MaxEdge: tc.edge})
		elapsed := time.Since(start)
		t.Logf("EncodeJPEG MaxEdge=%d took %s", tc.edge, elapsed)
		if err != nil {
			t.Fatal(err)
		}
		if elapsed > 10*time.Second {
			t.Errorf("EncodeJPEG took %s, want under 10s", elapsed)
		}
		cfg, _, err := image.DecodeConfig(bytes.NewReader(out))
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Width != tc.w || cfg.Height != tc.h {
			t.Fatalf("MaxEdge %d: got %dx%d, want %dx%d", tc.edge, cfg.Width, cfg.Height, tc.w, tc.h)
		}
	}
}

func TestEncodeJPEGOversizedHeaderAllocation(t *testing.T) {
	src := rastertest.JPEGHeader(60000, 60000, 1, make([]byte, 16))
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err := EncodeJPEG(src, Options{MaxEdge: 256})
	runtime.ReadMemStats(&after)
	if !errors.Is(err, ErrUndecodable) || !errors.Is(err, errJPEGDC) {
		t.Fatalf("err = %v, want ErrUndecodable wrapping errJPEGDC", err)
	}
	if grew := after.TotalAlloc - before.TotalAlloc; grew >= 16<<20 {
		t.Fatalf("allocated %d bytes, want under 16 MiB", grew)
	}
}

func TestEncodeJPEGAtBudgetBoundaries(t *testing.T) {
	// 8000x8000 is exactly the 64 MP budget and must still decode.
	// Header-only would fail at full decode, so use a real (small) image
	// for the success path and MaxEdgeLimit for the cap boundary.
	out, err := EncodeJPEG(tinyPNG(t), Options{MaxEdge: MaxEdgeLimit})
	if err != nil || len(out) == 0 {
		t.Fatalf("MaxEdge=%d should be accepted: %v", MaxEdgeLimit, err)
	}
	// 8000x8000 JPEG keeps the standard library: a header with no scan fails
	// there, not in the reduced decoder.
	_, err = decode(rastertest.JPEGHeader(8000, 8000, 1, nil), nil, true)
	if !errors.Is(err, ErrUndecodable) || errors.Is(err, errJPEGDC) {
		t.Fatalf("at-budget jpeg err = %v, want ErrUndecodable from the standard library", err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(rastertest.HeaderOnlyPNG(t, 8000, 8000)))
	if err != nil {
		t.Fatal(err)
	}
	if int64(cfg.Width)*int64(cfg.Height) > MaxDecodePixels {
		t.Fatalf("8000x8000 must sit inside the pixel budget (%d)", MaxDecodePixels)
	}
}

func TestSupported(t *testing.T) {
	cases := []struct {
		media string
		want  bool
	}{
		{"image/jpeg", true},
		{"image/jpg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/bmp", true},
		{"image/tiff", true},
		{"image/tif", true},
		{"image/webp", true},
		{"image/jpeg; charset=binary", true},
		{"IMAGE/PNG", true},
		{"image/heic", false},
		{"image/heif", false},
		{"image/svg+xml", false},
		{"application/pdf", false},
		{"application/octet-stream", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := Supported(tc.media); got != tc.want {
			t.Fatalf("%q: got %v want %v", tc.media, got, tc.want)
		}
	}
}

func TestEncodeJPEGOptionsAndTransform(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	var src bytes.Buffer
	if err := png.Encode(&src, img); err != nil {
		t.Fatal(err)
	}

	called := false
	out, err := EncodeJPEG(src.Bytes(), Options{
		MaxEdge: 40,
		Quality: 70,
		Transforms: []Transform{
			func(in image.Image) image.Image {
				called = true
				return in
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected transform to run")
	}
	decoded, _, err := image.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	b := decoded.Bounds()
	if b.Dx() != 40 || b.Dy() != 20 {
		t.Fatalf("got %dx%d", b.Dx(), b.Dy())
	}
}

func encodeImage(t *testing.T, w, h int, enc func(io.Writer, image.Image) error) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 7, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := enc(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func pngBytes(t *testing.T, w, h int) []byte {
	return encodeImage(t, w, h, func(wr io.Writer, m image.Image) error { return png.Encode(wr, m) })
}

func jpegBytes(t *testing.T, w, h int) []byte {
	return encodeImage(t, w, h, func(wr io.Writer, m image.Image) error { return jpeg.Encode(wr, m, nil) })
}

func tiffBytes(t *testing.T, w, h int) []byte {
	return encodeImage(t, w, h, func(wr io.Writer, m image.Image) error { return tiff.Encode(wr, m, nil) })
}

func TestDecode(t *testing.T) {
	both := []string{"png", "jpeg"}
	tests := []struct {
		name    string
		src     []byte
		formats []string
		want    error
	}{
		{"png in list", pngBytes(t, 4, 4), both, nil},
		{"tiff outside list", tiffBytes(t, 4, 4), both, ErrUnsupportedFormat},
		{"fake format outside list, no full decode", []byte("FAKEIMG!rest"), both, ErrUnsupportedFormat},
		{"tiff with nil list", tiffBytes(t, 4, 4), nil, nil},
		{"bomb header", rastertest.HeaderOnlyPNG(t, 100000, 100000), both, ErrTooLarge},
	}
	t.Run("oversized jpeg reduces only when asked", func(t *testing.T) {
		src := rastertest.JPEGHeader(65535, 65535, 1, nil)
		if _, err := decode(src, both, false); !errors.Is(err, ErrTooLarge) {
			t.Fatalf("reduce=false err = %v", err)
		}
		if _, err := decode(src, both, true); !errors.Is(err, ErrTooLarge) {
			t.Fatalf("reduce=true err = %v", err)
		}
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decode(tt.src, tt.formats, false)
			if !errors.Is(err, tt.want) || (tt.want == nil && err != nil) {
				t.Fatalf("decode err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestDecodeByteLimit(t *testing.T) {
	old := maxDecodeBytes
	maxDecodeBytes = 64
	t.Cleanup(func() { maxDecodeBytes = old })

	enc := map[string]func([]byte, Options) ([]byte, error){
		"EncodeJPEG": EncodeJPEG,
		"EncodePNG":  EncodePNG,
	}
	over := append([]byte("FAKEIMG!"), make([]byte, 57)...)
	at := over[:64]
	for name, fn := range enc {
		t.Run(name+"/over the limit", func(t *testing.T) {
			fakeConfigCalls.Store(0)
			_, err := fn(over, Options{MaxEdge: 256})
			if !errors.Is(err, ErrTooLarge) {
				t.Fatalf("err = %v, want ErrTooLarge", err)
			}
			if n := fakeConfigCalls.Load(); n != 0 {
				t.Fatalf("config func called %d times, want 0", n)
			}
		})
		t.Run(name+"/at the limit", func(t *testing.T) {
			fakeConfigCalls.Store(0)
			_, err := fn(at, Options{MaxEdge: 256})
			if !errors.Is(err, ErrUnsupportedFormat) {
				t.Fatalf("err = %v, want ErrUnsupportedFormat", err)
			}
			if n := fakeConfigCalls.Load(); n != 1 {
				t.Fatalf("config func called %d times, want 1", n)
			}
		})
	}
}

func TestEncodePNG(t *testing.T) {
	size := func(t *testing.T, b []byte) (int, int) {
		t.Helper()
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatal(err)
		}
		return img.Bounds().Dx(), img.Bounds().Dy()
	}
	t.Run("jpeg scaled to max edge", func(t *testing.T) {
		out, err := EncodePNG(jpegBytes(t, 4000, 1000), Options{MaxEdge: 2048})
		if err != nil {
			t.Fatal(err)
		}
		if w, h := size(t, out); w != 2048 || h != 512 {
			t.Fatalf("size = %dx%d, want 2048x512", w, h)
		}
	})
	t.Run("oversized jpeg never reduces", func(t *testing.T) {
		_, err := EncodePNG(rastertest.OversizedJPEG(t), Options{MaxEdge: 2048})
		if !errors.Is(err, ErrTooLarge) {
			t.Fatalf("err = %v, want ErrTooLarge", err)
		}
	})
	t.Run("small png keeps its size", func(t *testing.T) {
		out, err := EncodePNG(pngBytes(t, 300, 200), Options{MaxEdge: 2048})
		if err != nil {
			t.Fatal(err)
		}
		if w, h := size(t, out); w != 300 || h != 200 {
			t.Fatalf("size = %dx%d, want 300x200", w, h)
		}
	})
	t.Run("zero max edge", func(t *testing.T) {
		if _, err := EncodePNG(pngBytes(t, 4, 4), Options{}); err == nil {
			t.Fatal("want error")
		}
	})
	t.Run("gif outside formats", func(t *testing.T) {
		var buf bytes.Buffer
		if err := gif.Encode(&buf, image.NewPaletted(image.Rect(0, 0, 4, 4), color.Palette{color.Black, color.White}), nil); err != nil {
			t.Fatal(err)
		}
		_, err := EncodePNG(buf.Bytes(), Options{MaxEdge: 64, Formats: []string{"png"}})
		if !errors.Is(err, ErrUnsupportedFormat) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("EncodeJPEG honors formats", func(t *testing.T) {
		_, err := EncodeJPEG(jpegBytes(t, 8, 8), Options{MaxEdge: 64, Formats: []string{"png"}})
		if !errors.Is(err, ErrUnsupportedFormat) {
			t.Fatalf("err = %v", err)
		}
	})
}
