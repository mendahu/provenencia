package raster

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/raster/rastertest"
)

// segment is one marker segment found by jpegSegments.
type segment struct {
	marker  byte
	off     int // index of the segment's 0xFF
	payload []byte
}

// jpegSegments walks src from SOI by segment length to the first SOS, which
// it returns as the last segment.
func jpegSegments(src []byte) ([]segment, error) {
	if len(src) < 2 || src[0] != 0xFF || src[1] != 0xD8 {
		return nil, errors.New("no SOI")
	}
	var out []segment
	for pos := 2; ; {
		if pos+4 > len(src) || src[pos] != 0xFF {
			return nil, fmt.Errorf("no marker at %d", pos)
		}
		m := src[pos+1]
		n := int(src[pos+2])<<8 | int(src[pos+3])
		if n < 2 || pos+2+n > len(src) {
			return nil, fmt.Errorf("bad length at %d", pos)
		}
		out = append(out, segment{marker: m, off: pos, payload: src[pos+4 : pos+2+n]})
		if m == 0xDA {
			return out, nil
		}
		pos += 2 + n
	}
}

func find(t *testing.T, src []byte, marker byte) segment {
	t.Helper()
	segs, err := jpegSegments(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range segs {
		if s.marker == marker {
			return s
		}
	}
	t.Fatalf("no marker 0x%02X", marker)
	return segment{}
}

// decodableFixtures are the Go distribution JPEGs the reduced decoder reads.
var decodableFixtures = []string{
	"video-001.jpeg",
	"video-001.progressive.jpeg",
	"video-001.progressive.truncated.jpeg",
	"video-001.q50.410.jpeg",
	"video-001.q50.411.jpeg",
	"video-001.q50.420.jpeg",
	"video-001.q50.422.jpeg",
	"video-001.q50.440.jpeg",
	"video-001.q50.444.jpeg",
	"video-001.q50.420.progressive.jpeg",
	"video-001.q50.444.progressive.jpeg",
	"video-001.221212.jpeg",
	"video-001.restart2.jpeg",
	"video-001.rgb.jpeg",
	"video-001.separate.dc.progression.progressive.jpeg",
	"video-005.gray.jpeg",
	"video-005.gray.q50.2x2.jpeg",
	"video-005.gray.q50.2x2.progressive.jpeg",
}

func TestDecodeJPEGDCMatchesStdlib(t *testing.T) {
	for _, name := range decodableFixtures {
		t.Run(name, func(t *testing.T) {
			src := rastertest.GoFixture(t, name)
			got, err := decodeJPEGDC(src, MaxDecodePixels)
			if err != nil {
				t.Fatal(err)
			}
			if want := image.Rect(0, 0, 19, 13); got.Bounds() != want {
				t.Fatalf("bounds = %v, want %v", got.Bounds(), want)
			}
			gray := bytes.Contains([]byte(name), []byte("gray"))
			if _, ok := got.(*image.Gray); ok != gray {
				t.Fatalf("got %T, gray fixture = %v", got, gray)
			}
			if _, ok := got.(*image.RGBA); ok == gray {
				t.Fatalf("got %T, gray fixture = %v", got, gray)
			}
			full, err := jpeg.Decode(bytes.NewReader(src))
			if err != nil {
				t.Fatal(err)
			}
			sof := sofPayload(t, src)
			nf := int(sof[5])
			hs, vs := make([]int, nf), make([]int, nf)
			hmax, vmax := 1, 1
			for i := 0; i < nf; i++ {
				hs[i], vs[i] = int(sof[7+3*i]>>4), int(sof[7+3*i]&15)
				hmax, vmax = max(hmax, hs[i]), max(vmax, vs[i])
			}
			fb := full.Bounds()
			// compSample is the mean of component k of the full decode over
			// the pixel rectangle that pixel (x, y)'s block covers (D-22).
			compSample := func(k, x, y int) uint8 {
				bx, by := x*hs[k]/hmax, y*vs[k]/vmax
				x0, x1 := bx*8*hmax/hs[k], (bx+1)*8*hmax/hs[k]
				y0, y1 := by*8*vmax/vs[k], (by+1)*8*vmax/vs[k]
				x1, y1 = min(x1, fb.Dx()), min(y1, fb.Dy())
				var sum, n int
				for py := y0; py < y1; py++ {
					for px := x0; px < x1; px++ {
						var v uint8
						switch f := full.(type) {
						case *image.Gray:
							v = f.GrayAt(px, py).Y
						case *image.YCbCr:
							c := f.YCbCrAt(px, py)
							v = [3]uint8{c.Y, c.Cb, c.Cr}[k]
						case *image.RGBA:
							c := f.RGBAAt(px, py)
							v = [3]uint8{c.R, c.G, c.B}[k]
						default:
							t.Fatalf("unexpected result type %T", full)
						}
						sum += int(v)
						n++
					}
				}
				return uint8((2*sum + n) / (2 * n))
			}
			var diffSum [3]float64
			cells := 0
			for cy := 0; cy < 12; cy++ {
				for cx := 0; cx < 18; cx++ {
					var ref [3]uint8
					switch full.(type) {
					case *image.Gray:
						ref[0] = compSample(0, cx, cy)
					case *image.YCbCr:
						ref[0], ref[1], ref[2] = color.YCbCrToRGB(compSample(0, cx, cy), compSample(1, cx, cy), compSample(2, cx, cy))
					default:
						for k := range ref {
							ref[k] = compSample(k, cx, cy)
						}
					}
					c := color.RGBAModel.Convert(got.At(cx, cy)).(color.RGBA)
					for k, g := range [3]uint8{c.R, c.G, c.B} {
						diffSum[k] += math.Abs(float64(g) - float64(ref[k]))
					}
					cells++
				}
			}
			chans := 3
			if gray {
				chans = 1
			}
			for k := 0; k < chans; k++ {
				if m := diffSum[k] / float64(cells); m > 3.0 {
					t.Errorf("channel %d mean abs diff = %.2f, want <= 3.0", k, m)
				}
				t.Logf("channel %d mean abs diff = %.2f", k, diffSum[k]/float64(cells))
			}
		})
	}
}

// sofPayload returns the SOF0–2 payload of src.
func sofPayload(t *testing.T, src []byte) []byte {
	t.Helper()
	segs, err := jpegSegments(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range segs {
		if s.marker >= 0xC0 && s.marker <= 0xC2 {
			return s.payload
		}
	}
	t.Fatal("no SOF")
	return nil
}

// reducedPix returns the pixels of src decoded by the reduced decoder.
func reducedPix(t *testing.T, src []byte) []byte {
	t.Helper()
	img, err := decodeJPEGDC(src, MaxDecodePixels)
	if err != nil {
		t.Fatal(err)
	}
	return img.(*image.RGBA).Pix
}

// headerParts are the segments of rastertest.JPEGHeader, so a row can
// replace the one it changes.
type headerParts struct {
	dqt, dhtDC, dhtAC, sof, sos []byte
	sofMarker                   byte
}

func defaultParts(comps int) headerParts {
	segs, err := jpegSegments(rastertest.JPEGHeader(8, 8, comps, nil))
	if err != nil {
		panic(err)
	}
	return headerParts{
		dqt: segs[0].payload, dhtDC: segs[1].payload, dhtAC: segs[2].payload,
		sof: segs[3].payload, sos: segs[4].payload,
		sofMarker: 0xC0,
	}
}

func (p headerParts) build(scan []byte) []byte {
	out := []byte{0xFF, 0xD8}
	out = append(out, rastertest.JPEGSegment(0xDB, p.dqt)...)
	out = append(out, rastertest.JPEGSegment(0xC4, p.dhtDC)...)
	out = append(out, rastertest.JPEGSegment(0xC4, p.dhtAC)...)
	out = append(out, rastertest.JPEGSegment(p.sofMarker, p.sof)...)
	out = append(out, rastertest.JPEGSegment(0xDA, p.sos)...)
	return append(out, scan...)
}

func clone(b []byte) []byte { return append([]byte(nil), b...) }

func TestDecodeJPEGDCRefuses(t *testing.T) {
	scan := bytes.Repeat([]byte{0}, 32)
	patch := func(f func(p *headerParts)) []byte {
		p := defaultParts(3)
		p.dqt, p.dhtDC, p.dhtAC = clone(p.dqt), clone(p.dhtDC), clone(p.dhtAC)
		p.sof, p.sos = clone(p.sof), clone(p.sos)
		f(&p)
		return p.build(scan)
	}
	fixturePatch := func(name string, f func(t *testing.T, b []byte)) func(t *testing.T) []byte {
		return func(t *testing.T) []byte {
			b := clone(rastertest.GoFixture(t, name))
			f(t, b)
			return b
		}
	}
	rows := []struct {
		name string
		src  func(t *testing.T) []byte
	}{
		{"header/cmyk", func(t *testing.T) []byte { return rastertest.GoFixture(t, "video-001.cmyk.jpeg") }},
		{"header/arithmetic sof", fixturePatch("video-001.jpeg", func(t *testing.T, b []byte) {
			b[find(t, b, 0xC0).off+1] = 0xC9
		})},
		{"header/precision 12", func(*testing.T) []byte { return patch(func(p *headerParts) { p.sof[0] = 12 }) }},
		{"header/two components", func(*testing.T) []byte { return rastertest.JPEGHeader(8, 8, 2, scan) }},
		{"header/sampling 5", func(*testing.T) []byte { return patch(func(p *headerParts) { p.sof[7] = 0x51 }) }},
		{"header/sampling 3", func(*testing.T) []byte { return patch(func(p *headerParts) { p.sof[7] = 0x31 }) }},
		{"header/sampling does not divide", func(*testing.T) []byte {
			// Cb 4×4 over Y 2×2: Hmax % Hc ≠ 0
			return patch(func(p *headerParts) { p.sof[7] = 0x22; p.sof[10] = 0x44 })
		}},
		{"header/repeated component id", func(*testing.T) []byte { return patch(func(p *headerParts) { p.sof[9] = 1 }) }},
		{"header/dht counts over 256", func(*testing.T) []byte {
			return patch(func(p *headerParts) {
				p.dhtDC = append([]byte{0x00, 255, 2}, make([]byte, 14)...)
			})
		}},
		{"header/dht dc value 12", func(*testing.T) []byte {
			return patch(func(p *headerParts) { p.dhtDC[len(p.dhtDC)-1] = 12 })
		}},
		{"header/dqt id 4", func(*testing.T) []byte { return patch(func(p *headerParts) { p.dqt[0] = 0x04 }) }},
		{"header/sos names component 9", func(*testing.T) []byte { return patch(func(p *headerParts) { p.sos[1] = 9 }) }},
		{"header/progressive ac with two components", func(*testing.T) []byte {
			p := defaultParts(3)
			p.sos = clone(p.sos)
			p.sos[0] = 2
			p.sos = append(p.sos[:5], 1, 63, 0)
			p.sofMarker = 0xC2
			return p.build(scan)
		}},
		{"header/sos one byte short", func(*testing.T) []byte {
			return patch(func(p *headerParts) { p.sos = p.sos[:len(p.sos)-1] })
		}},
		{"scan/wrong restart marker", fixturePatch("video-001.restart2.jpeg", func(t *testing.T, b []byte) {
			sos := find(t, b, 0xDA)
			i := bytes.Index(b[sos.off+4+len(sos.payload):], []byte{0xFF, 0xD0})
			if i < 0 {
				t.Fatal("no RST0")
			}
			b[sos.off+4+len(sos.payload)+i+1] = 0xD1
		})},
		{"scan/component decoded twice", func(t *testing.T) []byte {
			b := clone(rastertest.GoFixture(t, "video-001.jpeg"))
			id := find(t, b, 0xC0).payload[6]
			extra := rastertest.JPEGSegment(0xDA, []byte{1, id, 0x00, 0, 63, 0})
			end := len(b) - 2
			return append(append(b[:end:end], extra...), 0x00, 0xFF, 0xD9)
		}},
		{"truncated/progressive-half", func(t *testing.T) []byte {
			b := rastertest.GoFixture(t, "video-001.progressive.jpeg")
			return b[:len(b)/2]
		}},
		{"truncated/baseline-half", func(t *testing.T) []byte {
			b := rastertest.GoFixture(t, "video-001.jpeg")
			return b[:len(b)/2]
		}},
		{"truncated/no-eoi", func(t *testing.T) []byte {
			b := rastertest.GoFixture(t, "video-001.jpeg")
			return b[:len(b)-2]
		}},
	}
	for _, tt := range rows {
		t.Run(tt.name, func(t *testing.T) {
			img, err := decodeJPEGDC(tt.src(t), MaxDecodePixels)
			if err == nil {
				t.Fatalf("decoded %v, want an error", img.Bounds())
			}
			if img != nil {
				t.Fatalf("image returned with error %v", err)
			}
			if !errors.Is(err, errJPEGDC) {
				t.Fatalf("err = %v, want errJPEGDC", err)
			}
		})
	}

	t.Run("header/frame-over-budget", func(t *testing.T) {
		_, err := decodeJPEGDC(rastertest.JPEGHeader(65535, 65535, 1, nil), MaxDecodePixels)
		if !errors.Is(err, ErrTooLarge) || errors.Is(err, errJPEGDC) {
			t.Fatalf("err = %v, want ErrTooLarge only", err)
		}
	})

	t.Run("scan/sof1 decodes like sof0", func(t *testing.T) {
		b := clone(rastertest.GoFixture(t, "video-001.jpeg"))
		want := reducedPix(t, b)
		b[find(t, b, 0xC0).off+1] = 0xC1
		if got := reducedPix(t, b); !bytes.Equal(got, want) {
			t.Fatal("SOF1 pixels differ from SOF0")
		}
	})
}

func TestJPEGDCHuffman(t *testing.T) {
	tab, _, err := newHuffTable(append([]byte{1, 2}, append(make([]byte, 14), 5, 6, 7)...), false)
	if err != nil {
		t.Fatal(err)
	}
	br := newBitReader([]byte{0b01011000}, 0)
	for i, want := range []byte{5, 6, 7} {
		got, err := br.huff(tab)
		if err != nil || got != want {
			t.Fatalf("symbol %d = %d, %v; want %d", i, got, err, want)
		}
	}

	if _, _, err := newHuffTable(append([]byte{3}, append(make([]byte, 15), 1, 2, 3)...), false); !errors.Is(err, errJPEGDC) {
		t.Fatalf("three 1-bit codes: err = %v, want overflow", err)
	}

	src := rastertest.GoFixture(t, "video-001.jpeg")
	dht := find(t, src, 0xC4).payload
	if _, _, err := newHuffTable(dht[1:], dht[0]>>4 == 0); err != nil {
		t.Fatalf("first DHT of video-001.jpeg: %v", err)
	}
}

// FuzzDecodeJPEGDC feeds arbitrary bytes to the reduced decoder: it must not
// panic, and a success must be the 1/8-scale size the frame header declares.
func FuzzDecodeJPEGDC(f *testing.F) {
	for _, name := range append([]string{"video-001.cmyk.jpeg"}, decodableFixtures...) {
		f.Add(rastertest.GoFixture(f, name))
	}
	grad := image.NewRGBA(image.Rect(0, 0, 37, 21))
	for y := 0; y < 21; y++ {
		for x := 0; x < 37; x++ {
			grad.Set(x, y, color.RGBA{R: uint8(x * 7), G: uint8(y * 11), B: uint8(x + y), A: 255})
		}
	}
	for _, q := range []int{10, 90} {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, grad, &jpeg.Options{Quality: q}); err != nil {
			f.Fatal(err)
		}
		f.Add(buf.Bytes())
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, 16, 16)), nil); err != nil {
		f.Fatal(err)
	}
	f.Add(buf.Bytes())

	f.Fuzz(func(t *testing.T, data []byte) {
		img, err := decodeJPEGDC(data, 1<<20)
		if err != nil {
			if img != nil {
				t.Fatal("image returned with an error")
			}
			return
		}
		b := img.Bounds()
		if int64(b.Dx())*int64(b.Dy()) > 1<<20 {
			t.Fatalf("bounds %v over the budget", b)
		}
		segs, err := jpegSegments(data)
		if err != nil {
			return
		}
		var sof []segment
		for _, s := range segs {
			if s.marker >= 0xC0 && s.marker <= 0xC2 {
				sof = append(sof, s)
			}
		}
		if len(sof) != 1 || len(sof[0].payload) < 5 {
			return
		}
		p := sof[0].payload
		h, w := int(p[1])<<8|int(p[2]), int(p[3])<<8|int(p[4])
		if b.Dx() != (w+7)/8 || b.Dy() != (h+7)/8 {
			t.Fatalf("bounds %v, want %dx%d", b, (w+7)/8, (h+7)/8)
		}
	})
}
