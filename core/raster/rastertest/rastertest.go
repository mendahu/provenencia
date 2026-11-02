// Package rastertest holds image fixtures shared by the tests of the packages
// that decode through core/raster: images over the decode budget that stay a
// few bytes (or one cheap encode) long.
package rastertest

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// HeaderOnlyPNG returns a PNG signature plus a valid IHDR chunk declaring
// w×h. image.DecodeConfig needs only this, so a decompression bomb stays a
// few bytes long.
func HeaderOnlyPNG(t *testing.T, w, h uint32) []byte {
	t.Helper()
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})
	data := make([]byte, 13)
	binary.BigEndian.PutUint32(data[0:], w)
	binary.BigEndian.PutUint32(data[4:], h)
	data[8] = 8 // bit depth
	data[9] = 6 // color type RGBA
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], 13)
	buf.Write(length[:])
	buf.WriteString("IHDR")
	buf.Write(data)
	crc := crc32.NewIEEE()
	crc.Write([]byte("IHDR"))
	crc.Write(data)
	var sum [4]byte
	binary.BigEndian.PutUint32(sum[:], crc.Sum32())
	buf.Write(sum[:])
	return buf.Bytes()
}

// gradient is an image whose pixels are computed by At, so a very large one
// holds no pixel buffer.
type gradient struct{ w, h int }

// Gradient returns a w×h image whose At computes a color.YCbCr from the
// coordinates. It allocates no pixel buffer.
func Gradient(w, h int) image.Image { return gradient{w: w, h: h} }

func (g gradient) ColorModel() color.Model { return color.YCbCrModel }
func (g gradient) Bounds() image.Rectangle { return image.Rect(0, 0, g.w, g.h) }
func (g gradient) At(x, y int) color.Color {
	return color.YCbCr{
		Y:  uint8((x + y) / 64),
		Cb: uint8(x / 32),
		Cr: uint8(y / 32),
	}
}

var (
	oversizedOnce sync.Once
	oversizedJPEG []byte
	oversizedErr  error
)

// OversizedJPEG returns an 8,008 × 8,000 baseline 4:2:0 JPEG (64.06 MP, just
// over raster.MaxDecodePixels, about 1 MB). It is encoded once per test
// binary. Callers must not modify the returned slice.
func OversizedJPEG(t *testing.T) []byte {
	t.Helper()
	oversizedOnce.Do(func() {
		var buf bytes.Buffer
		oversizedErr = jpeg.Encode(&buf, Gradient(8008, 8000), &jpeg.Options{Quality: 50})
		oversizedJPEG = buf.Bytes()
	})
	if oversizedErr != nil {
		t.Fatal(oversizedErr)
	}
	return oversizedJPEG
}

// JPEGSegment returns one marker segment: 0xFF, marker, a big-endian length
// of len(payload)+2, then payload.
func JPEGSegment(marker byte, payload []byte) []byte {
	out := make([]byte, 0, len(payload)+4)
	out = append(out, 0xFF, marker)
	out = binary.BigEndian.AppendUint16(out, uint16(len(payload)+2))
	return append(out, payload...)
}

// JPEGHeader returns the start of a baseline JPEG declaring w×h with comps
// components (ids 1, 2, 3 in order, 1×1 sampling): SOI, a quantisation table
// of 64 ones, DC and AC Huffman tables 0 each holding one 1-bit code for
// value 0, SOF0, SOS over every component, then scanData. It has no EOI.
// DecodeConfig reads only the SOF, so a huge declared size stays a few bytes.
func JPEGHeader(w, h uint16, comps int, scanData []byte) []byte {
	out := []byte{0xFF, 0xD8}

	dqt := append([]byte{0x00}, bytes.Repeat([]byte{1}, 64)...)
	out = append(out, JPEGSegment(0xDB, dqt)...)

	table := func(class byte) []byte {
		p := make([]byte, 0, 18)
		p = append(p, class<<4)
		p = append(p, 1)                   // one code of length 1
		p = append(p, make([]byte, 15)...) // none of lengths 2–16
		return append(p, 0x00)             // its value
	}
	out = append(out, JPEGSegment(0xC4, table(0))...)
	out = append(out, JPEGSegment(0xC4, table(1))...)

	sof := []byte{8}
	sof = binary.BigEndian.AppendUint16(sof, h)
	sof = binary.BigEndian.AppendUint16(sof, w)
	sof = append(sof, byte(comps))
	for i := 1; i <= comps; i++ {
		sof = append(sof, byte(i), 0x11, 0)
	}
	out = append(out, JPEGSegment(0xC0, sof)...)

	sos := []byte{byte(comps)}
	for i := 1; i <= comps; i++ {
		sos = append(sos, byte(i), 0x00)
	}
	sos = append(sos, 0, 63, 0)
	out = append(out, JPEGSegment(0xDA, sos)...)

	return append(out, scanData...)
}

// GoFixture returns a file from the Go distribution's image testdata copied
// under core/raster/testdata/third_party/go. It finds the folder from this
// source file, so any package's tests can call it.
func GoFixture(tb testing.TB, name string) []byte {
	tb.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatal("rastertest: cannot locate its source file")
	}
	dir := filepath.Join(filepath.Dir(file), "..", "testdata", "third_party", "go")
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		tb.Fatal(err)
	}
	return b
}
