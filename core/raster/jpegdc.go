package raster

import (
	"errors"
	"fmt"
	"image"
	"image/color"
)

// errJPEGDC is the base error of every refusal decodeJPEGDC makes itself, so
// a test can tell the reduced decoder from the standard library.
var errJPEGDC = errors.New("raster: jpeg reduced decode")

// jpegErr returns the error for a failed check of the reduced decoder.
func jpegErr(check string) error { return fmt.Errorf("%w: %s", errJPEGDC, check) }

// JPEG markers the reduced decoder reads (ITU-T T.81 Table B.1).
const (
	mSOF0 = 0xC0
	mSOF1 = 0xC1
	mSOF2 = 0xC2
	mDHT  = 0xC4
	mRST0 = 0xD0
	mRST7 = 0xD7
	mSOI  = 0xD8
	mEOI  = 0xD9
	mSOS  = 0xDA
	mDQT  = 0xDB
	mDRI  = 0xDD
	mAPP0 = 0xE0
	mAPP1 = 0xE1
	mAPPE = 0xEE
	mAPPF = 0xEF
	mCOM  = 0xFE
)

// The kinds of scan the decoder tells apart.
const (
	scanSequential = iota // all coefficients of a block, decoded and dropped but the DC
	scanDCFirst           // progressive DC, first pass for a component
	scanDCRefine          // progressive DC, one more bit per block
	scanACSkip            // progressive AC, passed over unread
)

// huffTable is one Huffman table with a 9-bit lookahead (T.81 Annex C).
type huffTable struct {
	look    [512]uint16 // length<<8 | value for a code of at most 9 bits; 0 if none
	count   [17]int     // codes of each length (index 0 unused)
	mincode [17]int32
	maxcode [17]int32
	valptr  [17]int
	vals    []byte
}

// newHuffTable builds a table from p, the 16 code-length counts followed by
// the values in code order, and returns the table and the bytes it used. It
// refuses a payload shorter than its counts, more than 256 codes, a table whose
// codes overflow a length (Annex C) and, when dc is set, a value over 11. That
// last check is the guard block relies on for the DC category.
func newHuffTable(p []byte, dc bool) (*huffTable, int, error) {
	if len(p) < 16 {
		return nil, 0, jpegErr("huffman table header length")
	}
	var counts [16]byte
	copy(counts[:], p[:16])
	total := 0
	for _, c := range counts {
		total += int(c)
	}
	if total > 256 {
		return nil, 0, jpegErr("huffman table has more than 256 codes")
	}
	if len(p) < 16+total {
		return nil, 0, jpegErr("huffman table values length")
	}
	vals := p[16 : 16+total]
	if dc {
		for _, v := range vals {
			if v > 11 {
				return nil, 0, jpegErr("dc huffman value over 11")
			}
		}
	}
	t := &huffTable{vals: append([]byte(nil), vals...)}
	code := int32(0)
	k := 0
	for l := 1; l <= 16; l++ {
		n := int(counts[l-1])
		t.count[l] = n
		t.valptr[l] = k
		t.mincode[l] = code
		for i := 0; i < n; i++ {
			if code >= 1<<uint(l) {
				return nil, 0, jpegErr("huffman code overflows its length")
			}
			if l <= 9 {
				shift := uint(9 - l)
				base := int(code) << shift
				for j := 0; j < 1<<shift; j++ {
					t.look[base+j] = uint16(l)<<8 | uint16(t.vals[k])
				}
			}
			code++
			k++
		}
		if code > 1<<uint(l) {
			return nil, 0, jpegErr("huffman code overflows its length")
		}
		t.maxcode[l] = code - 1
		code <<= 1
	}
	return t, 16 + total, nil
}

var errTruncated = jpegErr("scan data ends early")

// bitReader reads entropy-coded bits (T.81 B.1.1.5): a stuffed 0xFF00 is one
// 0xFF byte; any other 0xFF xx is a marker, where reading stops.
type bitReader struct {
	src    []byte
	pos    int    // next unread byte
	acc    uint64 // the low n bits are unread
	n      uint
	marker bool // fill stopped at a marker or at the end of src
}

func newBitReader(src []byte, pos int) *bitReader { return &bitReader{src: src, pos: pos} }

func (b *bitReader) fill() {
	for b.n <= 56 && !b.marker {
		if b.pos >= len(b.src) {
			b.marker = true
			return
		}
		c := b.src[b.pos]
		if c == 0xFF {
			if b.pos+1 >= len(b.src) || b.src[b.pos+1] != 0 {
				b.marker = true
				return
			}
			b.pos += 2
		} else {
			b.pos++
		}
		b.acc = b.acc<<8 | uint64(c)
		b.n += 8
	}
}

// peek returns the next k (at most 16) bits without consuming them, padded
// with zeros past the end of the data.
func (b *bitReader) peek(k uint) uint32 {
	if b.n < k {
		b.fill()
	}
	mask := uint64(1)<<k - 1
	if b.n >= k {
		return uint32(b.acc >> (b.n - k) & mask)
	}
	return uint32(b.acc << (k - b.n) & mask)
}

// bits consumes and returns k (at most 16) bits.
func (b *bitReader) bits(k uint) (uint32, error) {
	if k == 0 {
		return 0, nil
	}
	if b.n < k {
		b.fill()
		if b.n < k {
			return 0, errTruncated
		}
	}
	v := uint32(b.acc >> (b.n - k) & (uint64(1)<<k - 1))
	b.n -= k
	return v, nil
}

// huff decodes one symbol.
func (b *bitReader) huff(t *huffTable) (byte, error) {
	if e := t.look[b.peek(9)]; e != 0 {
		l := uint(e >> 8)
		if l > b.n {
			return 0, errTruncated
		}
		b.n -= l
		return byte(e), nil
	}
	var code int32
	for l := 1; l <= 16; l++ {
		bit, err := b.bits(1)
		if err != nil {
			return 0, err
		}
		code = code<<1 | int32(bit)
		if t.count[l] > 0 && code >= t.mincode[l] && code <= t.maxcode[l] {
			return t.vals[t.valptr[l]+int(code-t.mincode[l])], nil
		}
	}
	return 0, jpegErr("invalid huffman code")
}

// discard drops the unread bits (the padding before a marker) and returns the
// position to search for the next marker from.
func (b *bitReader) discard() int {
	b.acc, b.n = 0, 0
	return b.pos
}

// markerAt returns the index of the first marker at or after pos: an 0xFF
// followed, past any further 0xFF fill bytes, by a byte that is not 0x00 (a
// stuffed 0xFF) and, when skipRST is set, not RST0–RST7. It returns the index
// of the first 0xFF, the marker code and the index just past the marker; ok
// is false when there is none.
func markerAt(src []byte, pos int, skipRST bool) (idx int, code byte, end int, ok bool) {
	for i := pos; i < len(src); i++ {
		if src[i] != 0xFF {
			continue
		}
		j := i + 1
		for j < len(src) && src[j] == 0xFF {
			j++
		}
		if j >= len(src) {
			return 0, 0, 0, false
		}
		c := src[j]
		if c == 0 || (skipRST && c >= mRST0 && c <= mRST7) {
			i = j
			continue
		}
		return i, c, j + 1, true
	}
	return 0, 0, 0, false
}

// dcComponent is one frame component and its DC coefficients.
type dcComponent struct {
	id           byte
	h, v         int
	tq           int
	gridW, gridH int     // blocks held: whole MCUs
	scanW, scanH int     // blocks a single-component scan walks
	dc           []int16 // one DC value per block, gridW per row
	step         int32   // DC quantisation step, latched at the first scan
	covered      bool    // a sequential or DC-first scan has decoded it
	lastAl       int     // Al of its last progressive DC scan
}

// dcScan is one parsed scan header.
type dcScan struct {
	kind  int
	comps []*dcComponent
	dcTab []*huffTable
	acTab []*huffTable
	al    uint
}

// dcDecoder is the state of one decodeJPEGDC call.
type dcDecoder struct {
	src       []byte
	pos       int
	maxPixels int64

	huff [2][4]*huffTable // [0] DC, [1] AC
	qDef [4]bool
	q0   [4]int32 // the DC (first) entry of each quantisation table
	ri   int      // restart interval

	jfif       bool
	adobeValid bool
	adobeT     byte

	haveFrame   bool
	progressive bool
	w, h        int
	w8, h8      int
	hmax, vmax  int
	mcusX       int
	mcusY       int
	comps       []*dcComponent
	allocated   bool
	scans       int
}

// decodeJPEGDC decodes a JPEG at 1/8 scale from its DC coefficients only.
// It returns an image of ⌈W/8⌉ × ⌈H/8⌉ pixels with Min at (0, 0), or an
// error and no image. It allocates nothing sized by the frame until the
// frame is within maxPixels and the first scan has enough bytes (D-12).
func decodeJPEGDC(src []byte, maxPixels int64) (image.Image, error) {
	d := &dcDecoder{src: src, maxPixels: maxPixels}
	if len(src) < 2 || src[0] != 0xFF || src[1] != mSOI {
		return nil, jpegErr("missing start of image")
	}
	d.pos = 2
	for {
		m, err := d.readMarker()
		if err != nil {
			return nil, err
		}
		switch {
		case m == mEOI:
			return d.finish()
		case m == mSOF0 || m == mSOF1 || m == mSOF2:
			p, err := d.segment()
			if err != nil {
				return nil, err
			}
			err = d.parseSOF(m, p)
			if err != nil {
				return nil, err
			}
		case m == mDHT:
			p, err := d.segment()
			if err == nil {
				err = d.parseDHT(p)
			}
			if err != nil {
				return nil, err
			}
		case m == mDQT:
			p, err := d.segment()
			if err == nil {
				err = d.parseDQT(p)
			}
			if err != nil {
				return nil, err
			}
		case m == mDRI:
			p, err := d.segment()
			if err != nil {
				return nil, err
			}
			if len(p) != 2 {
				return nil, jpegErr("restart interval length")
			}
			d.ri = int(p[0])<<8 | int(p[1])
		case m == mAPP0:
			p, err := d.segment()
			if err != nil {
				return nil, err
			}
			d.jfif = len(p) >= 5 && string(p[:5]) == "JFIF\x00"
		case m == mAPPE:
			p, err := d.segment()
			if err != nil {
				return nil, err
			}
			if len(p) >= 12 && string(p[:5]) == "Adobe" {
				d.adobeValid, d.adobeT = true, p[11]
			}
		case (m >= mAPP1 && m <= mAPPF) || m == mCOM:
			if _, err := d.segment(); err != nil {
				return nil, err
			}
		case m == mSOS:
			p, err := d.segment()
			if err != nil {
				return nil, err
			}
			sc, err := d.parseSOS(p)
			if err != nil {
				return nil, err
			}
			if err := d.decodeScan(sc); err != nil {
				return nil, err
			}
		default:
			return nil, jpegErr(fmt.Sprintf("unsupported marker 0x%02X", m))
		}
	}
}

// readMarker returns the next marker, skipping 0xFF fill bytes before it.
func (d *dcDecoder) readMarker() (byte, error) {
	if d.pos >= len(d.src) {
		return 0, jpegErr("data ends before end of image")
	}
	if d.src[d.pos] != 0xFF {
		return 0, jpegErr("expected a marker")
	}
	for d.pos < len(d.src) && d.src[d.pos] == 0xFF {
		d.pos++
	}
	if d.pos >= len(d.src) {
		return 0, jpegErr("data ends inside a marker")
	}
	m := d.src[d.pos]
	d.pos++
	if m == 0 {
		return 0, jpegErr("stuffed byte outside scan data")
	}
	return m, nil
}

// segment returns the payload of the marker segment at pos, after checking
// its length against the bytes left.
func (d *dcDecoder) segment() ([]byte, error) {
	if d.pos+2 > len(d.src) {
		return nil, jpegErr("data ends inside a segment length")
	}
	n := int(d.src[d.pos])<<8 | int(d.src[d.pos+1])
	if n < 2 || d.pos+n > len(d.src) {
		return nil, jpegErr("segment length out of range")
	}
	p := d.src[d.pos+2 : d.pos+n]
	d.pos += n
	return p, nil
}

func validSampling(f int) bool { return f == 1 || f == 2 || f == 4 }

func (d *dcDecoder) parseSOF(marker byte, p []byte) error {
	if d.haveFrame {
		return jpegErr("repeated frame header")
	}
	if len(p) < 6 {
		return jpegErr("frame header length")
	}
	if p[0] != 8 {
		return jpegErr("sample precision is not 8")
	}
	h := int(p[1])<<8 | int(p[2])
	w := int(p[3])<<8 | int(p[4])
	nf := int(p[5])
	if nf != 1 && nf != 3 {
		return jpegErr("component count is not 1 or 3")
	}
	if len(p) != 6+3*nf {
		return jpegErr("frame header length")
	}
	if w < 1 || h < 1 {
		return jpegErr("frame has no size")
	}
	comps := make([]*dcComponent, nf)
	hmax, vmax := 1, 1
	for i := range comps {
		c := &dcComponent{id: p[6+3*i], h: int(p[7+3*i] >> 4), v: int(p[7+3*i] & 15), tq: int(p[8+3*i]), lastAl: -1}
		if nf == 1 {
			c.h, c.v = 1, 1
		}
		if !validSampling(c.h) || !validSampling(c.v) {
			return jpegErr("sampling factor is not 1, 2 or 4")
		}
		if c.tq > 3 {
			return jpegErr("quantisation table id over 3")
		}
		for _, o := range comps[:i] {
			if o.id == c.id {
				return jpegErr("repeated component id")
			}
		}
		comps[i] = c
		hmax, vmax = max(hmax, c.h), max(vmax, c.v)
	}
	if nf == 3 {
		for _, c := range comps {
			if hmax%c.h != 0 || vmax%c.v != 0 {
				return jpegErr("sampling factors do not divide")
			}
		}
	}
	w8, h8 := (w+7)/8, (h+7)/8
	if int64(w8)*int64(h8) > d.maxPixels {
		return ErrTooLarge
	}
	d.haveFrame = true
	d.progressive = marker == mSOF2
	d.w, d.h, d.w8, d.h8 = w, h, w8, h8
	d.hmax, d.vmax = hmax, vmax
	d.mcusX = (w + 8*hmax - 1) / (8 * hmax)
	d.mcusY = (h + 8*vmax - 1) / (8 * vmax)
	for _, c := range comps {
		c.gridW, c.gridH = d.mcusX*c.h, d.mcusY*c.v
		c.scanW = ((w*c.h+hmax-1)/hmax + 7) / 8
		c.scanH = ((h*c.v+vmax-1)/vmax + 7) / 8
	}
	d.comps = comps
	return nil
}

func (d *dcDecoder) parseDHT(p []byte) error {
	for len(p) > 0 {
		tc, th := int(p[0]>>4), int(p[0]&15)
		if tc > 1 || th > 3 {
			return jpegErr("huffman table class or id")
		}
		t, n, err := newHuffTable(p[1:], tc == 0)
		if err != nil {
			return err
		}
		d.huff[tc][th] = t
		p = p[1+n:]
	}
	return nil
}

func (d *dcDecoder) parseDQT(p []byte) error {
	for len(p) > 0 {
		pq, tq := int(p[0]>>4), int(p[0]&15)
		if pq > 1 || tq > 3 {
			return jpegErr("quantisation table precision or id")
		}
		size := 1 + 64*(pq+1)
		if len(p) < size {
			return jpegErr("quantisation table length")
		}
		if pq == 0 {
			d.q0[tq] = int32(p[1])
		} else {
			d.q0[tq] = int32(p[1])<<8 | int32(p[2])
		}
		d.qDef[tq] = true
		p = p[size:]
	}
	return nil
}

func (d *dcDecoder) parseSOS(p []byte) (*dcScan, error) {
	if !d.haveFrame {
		return nil, jpegErr("scan before frame header")
	}
	if len(p) < 1 {
		return nil, jpegErr("scan header length")
	}
	ns := int(p[0])
	if len(p) != 4+2*ns {
		return nil, jpegErr("scan header length")
	}
	if ns < 1 || ns > len(d.comps) {
		return nil, jpegErr("scan component count")
	}
	sc := &dcScan{}
	tds, tas := make([]int, ns), make([]int, ns)
	for i := 0; i < ns; i++ {
		var c *dcComponent
		for _, f := range d.comps {
			if f.id == p[1+2*i] {
				c = f
			}
		}
		if c == nil {
			return nil, jpegErr("scan names a component not in the frame")
		}
		for _, o := range sc.comps {
			if o == c {
				return nil, jpegErr("scan repeats a component")
			}
		}
		sc.comps = append(sc.comps, c)
		tds[i], tas[i] = int(p[2+2*i]>>4), int(p[2+2*i]&15)
	}
	ss, se := int(p[1+2*ns]), int(p[2+2*ns])
	ah, al := int(p[3+2*ns]>>4), int(p[3+2*ns]&15)

	needDC := func() error {
		for i := range sc.comps {
			if tds[i] > 3 || d.huff[0][tds[i]] == nil {
				return jpegErr("scan uses an undefined dc huffman table")
			}
			sc.dcTab = append(sc.dcTab, d.huff[0][tds[i]])
		}
		return nil
	}
	if !d.progressive {
		if ss != 0 || se != 63 || ah != 0 || al != 0 {
			return nil, jpegErr("sequential scan parameters")
		}
		if err := needDC(); err != nil {
			return nil, err
		}
		for i := range sc.comps {
			if tas[i] > 3 || d.huff[1][tas[i]] == nil {
				return nil, jpegErr("scan uses an undefined ac huffman table")
			}
			sc.acTab = append(sc.acTab, d.huff[1][tas[i]])
		}
		sc.kind = scanSequential
	} else if ss == 0 {
		if se != 0 {
			return nil, jpegErr("progressive dc scan spectral end")
		}
		if ah == 0 {
			if al > 13 {
				return nil, jpegErr("progressive successive approximation")
			}
			if err := needDC(); err != nil {
				return nil, err
			}
			sc.kind = scanDCFirst
		} else {
			if ah != al+1 {
				return nil, jpegErr("progressive dc refinement bit positions")
			}
			for _, c := range sc.comps {
				if c.lastAl != ah {
					return nil, jpegErr("dc refinement out of order")
				}
			}
			sc.kind = scanDCRefine
		}
		sc.al = uint(al)
	} else {
		if ss > se || se > 63 || ns != 1 {
			return nil, jpegErr("progressive ac scan parameters")
		}
		if (ah != 0 && ah != al+1) || al > 13 {
			return nil, jpegErr("progressive successive approximation")
		}
		sc.kind = scanACSkip
	}
	if sc.kind == scanSequential || sc.kind == scanDCFirst {
		for _, c := range sc.comps {
			if c.covered {
				return nil, jpegErr("component decoded twice")
			}
		}
	}
	return sc, nil
}

// alloc allocates the DC buffers for every component after checking that the
// data left can hold one bit per block (D-12).
func (d *dcDecoder) alloc() error {
	var blocks int64
	for _, c := range d.comps {
		blocks += int64(c.gridW) * int64(c.gridH)
	}
	if int64(len(d.src)-d.pos)*8 < blocks {
		return jpegErr("scan data too short for the frame")
	}
	for _, c := range d.comps {
		c.dc = make([]int16, c.gridW*c.gridH)
	}
	d.allocated = true
	return nil
}

func (d *dcDecoder) decodeScan(sc *dcScan) error {
	d.scans++
	if sc.kind == scanSequential || sc.kind == scanDCFirst {
		if !d.allocated {
			if err := d.alloc(); err != nil {
				return err
			}
		}
		for _, c := range sc.comps {
			if !d.qDef[c.tq] {
				return jpegErr("scan uses an undefined quantisation table")
			}
			c.step = d.q0[c.tq]
		}
	} else if !d.allocated {
		return jpegErr("scan before the first dc scan")
	}
	if sc.kind == scanACSkip {
		idx, _, _, ok := markerAt(d.src, d.pos, true)
		if !ok {
			return errTruncated
		}
		d.pos = idx
		return nil
	}

	br := newBitReader(d.src, d.pos)
	var pred [3]int32
	single := len(sc.comps) == 1
	var cols, mcus int
	if single {
		cols = sc.comps[0].scanW
		mcus = cols * sc.comps[0].scanH
	} else {
		cols = d.mcusX
		mcus = d.mcusX * d.mcusY
	}
	rst := 0
	for m := 0; m < mcus; m++ {
		if d.ri > 0 && m > 0 && m%d.ri == 0 {
			pos := br.discard()
			_, code, end, ok := markerAt(d.src, pos, false)
			if !ok || code != byte(mRST0+rst&7) {
				return jpegErr("missing or out-of-order restart marker")
			}
			rst++
			br = newBitReader(d.src, end)
			pred = [3]int32{}
		}
		mx, my := m%cols, m/cols
		for i, c := range sc.comps {
			if single {
				if err := d.block(sc, br, i, c, mx, my, &pred[i]); err != nil {
					return err
				}
				continue
			}
			for v := 0; v < c.v; v++ {
				for h := 0; h < c.h; h++ {
					if err := d.block(sc, br, i, c, mx*c.h+h, my*c.v+v, &pred[i]); err != nil {
						return err
					}
				}
			}
		}
	}
	idx, _, _, ok := markerAt(d.src, br.discard(), false)
	if !ok {
		return jpegErr("data ends before end of image")
	}
	d.pos = idx
	for _, c := range sc.comps {
		switch sc.kind {
		case scanSequential:
			c.covered = true
		case scanDCFirst:
			c.covered = true
			c.lastAl = int(sc.al)
		case scanDCRefine:
			c.lastAl = int(sc.al)
		}
	}
	return nil
}

// block decodes one block of component i in scan sc.
func (d *dcDecoder) block(sc *dcScan, br *bitReader, i int, c *dcComponent, bx, by int, pred *int32) error {
	at := by*c.gridW + bx
	if sc.kind == scanDCRefine {
		bit, err := br.bits(1)
		if err != nil {
			return err
		}
		c.dc[at] |= int16(bit << sc.al)
		return nil
	}
	s, err := br.huff(sc.dcTab[i])
	if err != nil {
		return err
	}
	var diff int32
	if s > 0 {
		v, err := br.bits(uint(s))
		if err != nil {
			return err
		}
		diff = int32(v)
		if diff < 1<<(s-1) {
			diff -= 1<<s - 1
		}
	}
	p := int64(*pred) + int64(diff)
	p = min(max(p, -1<<31), 1<<31-1)
	*pred = int32(p)
	shift := uint(0)
	if sc.kind == scanDCFirst {
		shift = sc.al
	}
	c.dc[at] = int16(min(max(p<<shift, -1<<15), 1<<15-1))
	if sc.kind == scanDCFirst {
		return nil
	}
	// Sequential: decode the 63 AC coefficients and drop them.
	ac := sc.acTab[i]
	for k := 1; k < 64; {
		rs, err := br.huff(ac)
		if err != nil {
			return err
		}
		r, s := int(rs>>4), uint(rs&15)
		if s == 0 {
			if r != 15 {
				break
			}
			k += 16
			continue
		}
		k += r
		if k > 63 {
			return jpegErr("ac coefficient past the block")
		}
		if s > 10 {
			return jpegErr("ac category over 10")
		}
		if _, err := br.bits(s); err != nil {
			return err
		}
		k++
	}
	return nil
}

// isRGB mirrors the standard library: three samples are R, G, B unless a
// JFIF header says YCbCr; an Adobe transform of 0 or component ids R, G, B
// say RGB.
func (d *dcDecoder) isRGB() bool {
	if d.jfif {
		return false
	}
	if d.adobeValid && d.adobeT == 0 {
		return true
	}
	return d.comps[0].id == 'R' && d.comps[1].id == 'G' && d.comps[2].id == 'B'
}

// sample turns a block's DC value into the block's mean sample: DC×Q/8 plus
// the level shift, rounded half away from zero and clamped to a byte.
func sample(dc int16, step int32) uint8 {
	v := int64(dc) * int64(step)
	if v >= 0 {
		v = (v + 4) >> 3
	} else {
		v = -((-v + 4) >> 3)
	}
	return uint8(min(max(v+128, 0), 255))
}

// finish checks that the stream is complete and builds the image.
func (d *dcDecoder) finish() (image.Image, error) {
	if !d.haveFrame || d.scans == 0 {
		return nil, jpegErr("no image data")
	}
	for _, c := range d.comps {
		if !c.covered {
			return nil, jpegErr("a component was never decoded")
		}
	}
	w8, h8 := d.w8, d.h8
	if len(d.comps) == 1 {
		c := d.comps[0]
		img := image.NewGray(image.Rect(0, 0, w8, h8))
		for y := 0; y < h8; y++ {
			row := img.Pix[y*img.Stride : y*img.Stride+w8]
			for x := range row {
				row[x] = sample(c.dc[y*c.gridW+x], c.step)
			}
		}
		return img, nil
	}
	img := image.NewRGBA(image.Rect(0, 0, w8, h8))
	rgb := d.isRGB()
	var xi [3][]int
	for k, c := range d.comps {
		xi[k] = make([]int, w8)
		for x := range xi[k] {
			xi[k][x] = x * c.h / d.hmax
		}
	}
	for y := 0; y < h8; y++ {
		var rows [3]int
		for k, c := range d.comps {
			rows[k] = (y * c.v / d.vmax) * c.gridW
		}
		out := img.Pix[y*img.Stride:]
		for x := 0; x < w8; x++ {
			a := sample(d.comps[0].dc[rows[0]+xi[0][x]], d.comps[0].step)
			b := sample(d.comps[1].dc[rows[1]+xi[1][x]], d.comps[1].step)
			c := sample(d.comps[2].dc[rows[2]+xi[2][x]], d.comps[2].step)
			if !rgb {
				a, b, c = color.YCbCrToRGB(a, b, c)
			}
			out[4*x], out[4*x+1], out[4*x+2], out[4*x+3] = a, b, c, 0xFF
		}
	}
	return img, nil
}
