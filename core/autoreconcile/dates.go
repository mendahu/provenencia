package autoreconcile

import "github.com/mendahu/provenencia/core/datevalue"

// The date module (conclusion-reconciliation.md §7.1). A date is a window
// over its civil components. Missing finer fields are unknown, never zero.
//
// Equal structured dates share a key and merge. A wider window folds into a
// strictly narrower one it contains: MAY 1985 into 14 MAY 1985, BEF 1900
// into 1885, ABT 1985 into 1985. Disjoint points stay apart (APR and MAY,
// 3 MAY and 14 JUN): they are not widened into a range and not collapsed to
// the year they share. An open window folds into each closed date it
// contains and does not glue those dates together.
//
// A phrase with no year is no evidence. Time zones are not compared.

// dateSortOffset shifts a proleptic day number into a zero-padded sort key
// that still orders dates before year 1.
const dateSortOffset = 100_000_000

type dateModule struct{}

// dateWindow is an inclusive span of proleptic Gregorian days (1970-01-01
// is 0). An open end is unbounded on that side.
type dateWindow struct {
	lo, hi         int
	openLo, openHi bool
	precision      int  // 1 year, 2 month, 3 day, then clock fields
	qualified      bool // ABT, BEF, AFT, or a range
}

type dateUnit struct {
	date   datevalue.Value
	window dateWindow
}

func (dateModule) split(v Value) (map[string]unit, bool) {
	if v.Date == nil {
		return nil, false
	}
	w, ok := windowOf(*v.Date)
	if !ok {
		return nil, false
	}
	return map[string]unit{wholeValue: {key: dateKey(v.Date), data: dateUnit{date: *v.Date, window: w}}}, true
}

// fold: from is the wider date and into is the narrower one it contains.
func (dateModule) fold(_ string, from, into unit) bool {
	a := from.data.(dateUnit).window
	b := into.data.(dateUnit).window
	return a.contains(b) && b.moreSpecificThan(a)
}

// outvotes is false: a majority never crowds out a disjoint date.
func (dateModule) outvotes(string, unit, unit) bool { return false }

func (dateModule) oneValue() bool { return false }

func (dateModule) assemble(_ []Candidate, settled map[string][]unit) Value {
	u := settled[wholeValue]
	d := u[0].data.(dateUnit).date
	return Value{Date: &d}
}

// DateBounds returns the inclusive proleptic day numbers of d's window.
// A nil bound is open. ok is false when d has no year to place.
func DateBounds(d datevalue.Value) (lo, hi *int, ok bool) {
	w, ok := windowOf(d)
	if !ok {
		return nil, nil, false
	}
	if !w.openLo {
		v := w.lo
		lo = &v
	}
	if !w.openHi {
		v := w.hi
		hi = &v
	}
	return lo, hi, true
}

// DateEnd is one end of a date as written, for comparing dates by the days
// they cover. Month and Day are 0 when the end does not name them. Lo and Hi
// are its inclusive proleptic day numbers; From and To place it on a line of
// years, where a bound on day d of year y covers [y + (d-1)/n, y + d/n) with
// n the days in y, so a whole year is [y, y+1). An open end has Open set and
// every other field zero.
type DateEnd struct {
	Year, Month, Day int
	Lo, Hi           int
	From, To         float64
	Open             bool
}

// DateEnds returns the first and last end of d. A point has the same end
// twice, BEF an open first end, AFT an open last end; a range has its start
// bound first and its end bound last (swapped when the start begins after the
// end finishes), and an open end on a side with no year. ok is false when d
// has no year to place, or a month or day that does not exist. Civil
// components are read as Gregorian, as DateBounds does.
func DateEnds(d datevalue.Value) (first, last DateEnd, ok bool) {
	if d.Kind == datevalue.KindRange {
		if d.StartYear == nil && d.EndYear == nil {
			return DateEnd{}, DateEnd{}, false
		}
		first, last = DateEnd{Open: true}, DateEnd{Open: true}
		if d.StartYear != nil {
			if first, ok = boundEnd(d.StartYear, d.StartMonth, d.StartDay); !ok {
				return DateEnd{}, DateEnd{}, false
			}
		}
		if d.EndYear != nil {
			if last, ok = boundEnd(d.EndYear, d.EndMonth, d.EndDay); !ok {
				return DateEnd{}, DateEnd{}, false
			}
		}
		if !first.Open && !last.Open && first.Lo > last.Hi {
			first, last = last, first
		}
		return first, last, true
	}
	y, m, day := d.StartYear, d.StartMonth, d.StartDay
	if y == nil {
		y, m, day = d.EndYear, d.EndMonth, d.EndDay
	}
	e, ok := boundEnd(y, m, day)
	if !ok {
		return DateEnd{}, DateEnd{}, false
	}
	first, last = e, e
	switch d.Qualifier {
	case datevalue.QualifierBEF:
		first = DateEnd{Open: true}
	case datevalue.QualifierAFT:
		last = DateEnd{Open: true}
	}
	return first, last, true
}

// boundEnd is the end one bound covers at its precision.
func boundEnd(year, month, day *int) (DateEnd, bool) {
	lo, hi, ok := datevalue.BoundDays(year, month, day)
	if !ok {
		return DateEnd{}, false
	}
	y := *year
	e := DateEnd{Year: y, Lo: lo, Hi: hi}
	if month != nil {
		e.Month = *month
	}
	if day != nil {
		e.Day = *day
	}
	e.From, e.To, _ = datevalue.YearLine(lo, hi)
	return e, true
}

// Window is an inclusive proleptic-day span. A nil bound is open.
// A fully open window is "always" (no place period, undated place).
type Window struct {
	Lo, Hi *int
}

// HoldKind is how a query date relates to a membership span.
type HoldKind int

const (
	// HoldMiss: query and membership are disjoint.
	HoldMiss HoldKind = iota
	// HoldYes: query lies entirely inside membership.
	HoldYes
	// HoldAmbiguous: they overlap but the query is not fully inside
	// (approximate or ranged date straddling a change).
	HoldAmbiguous
)

// Always is a fully open window (no period).
func Always() Window { return Window{} }

// WindowOfDate is DateBounds as a Window. ok is false when d has no year.
func WindowOfDate(d datevalue.Value) (Window, bool) {
	lo, hi, ok := DateBounds(d)
	if !ok {
		return Window{}, false
	}
	return Window{Lo: lo, Hi: hi}, true
}

// PeriodWindow is a Place's existence span from optional start and end
// dates. Missing sides are open; both missing is Always.
func PeriodWindow(start, end *datevalue.Value) Window {
	w := Always()
	if start != nil {
		if sw, ok := WindowOfDate(*start); ok {
			w.Lo = sw.Lo
		}
	}
	if end != nil {
		if ew, ok := WindowOfDate(*end); ok {
			w.Hi = ew.Hi
		}
	}
	return w
}

// LinkMembership is when a hierarchical place relationship holds.
// An undated link (nil start and end) holds on the overlap of the two
// places' periods. A dated link holds inside its own span and both periods.
func LinkMembership(linkStart, linkEnd *datevalue.Value, fromPeriod, toPeriod Window) (Window, bool) {
	base, ok := Intersect(fromPeriod, toPeriod)
	if !ok {
		return Window{}, false
	}
	if linkStart == nil && linkEnd == nil {
		return base, true
	}
	link := PeriodWindow(linkStart, linkEnd)
	return Intersect(base, link)
}

// Overlaps reports whether a and b share any day (open ends included).
func Overlaps(a, b Window) bool {
	if a.Lo != nil && b.Hi != nil && *a.Lo > *b.Hi {
		return false
	}
	if b.Lo != nil && a.Hi != nil && *b.Lo > *a.Hi {
		return false
	}
	return true
}

// Intersect is the overlap of a and b. ok is false when empty.
func Intersect(a, b Window) (Window, bool) {
	if !Overlaps(a, b) {
		return Window{}, false
	}
	out := Window{}
	out.Lo = laterBound(a.Lo, b.Lo)
	out.Hi = earlierBound(a.Hi, b.Hi)
	if out.Lo != nil && out.Hi != nil && *out.Lo > *out.Hi {
		return Window{}, false
	}
	return out, true
}

// Contains reports whether outer fully covers inner.
func Contains(outer, inner Window) bool {
	if inner.Lo == nil {
		if outer.Lo != nil {
			return false
		}
	} else if outer.Lo != nil && *inner.Lo < *outer.Lo {
		return false
	}
	if inner.Hi == nil {
		if outer.Hi != nil {
			return false
		}
	} else if outer.Hi != nil && *inner.Hi > *outer.Hi {
		return false
	}
	return true
}

// Relate says how query sits against membership for chain composition.
func Relate(query, membership Window) HoldKind {
	if !Overlaps(query, membership) {
		return HoldMiss
	}
	if Contains(membership, query) {
		return HoldYes
	}
	return HoldAmbiguous
}

// HoldsAt is true when the link should appear in a chain at query
// (HoldYes or HoldAmbiguous).
func HoldsAt(query, membership Window) bool {
	k := Relate(query, membership)
	return k == HoldYes || k == HoldAmbiguous
}

func laterBound(a, b *int) *int {
	switch {
	case a == nil:
		return copyBound(b)
	case b == nil:
		return copyBound(a)
	case *a >= *b:
		return copyBound(a)
	default:
		return copyBound(b)
	}
}

func earlierBound(a, b *int) *int {
	switch {
	case a == nil:
		return copyBound(b)
	case b == nil:
		return copyBound(a)
	case *a <= *b:
		return copyBound(a)
	default:
		return copyBound(b)
	}
}

func copyBound(v *int) *int {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}

func (w dateWindow) contains(inner dateWindow) bool {
	if !w.openLo && (inner.openLo || inner.lo < w.lo) {
		return false
	}
	if !w.openHi && (inner.openHi || inner.hi > w.hi) {
		return false
	}
	return true
}

// moreSpecificThan reports whether w is a strictly narrower statement than o.
func (w dateWindow) moreSpecificThan(o dateWindow) bool {
	if w.precision != o.precision {
		return w.precision > o.precision
	}
	if openEnds(w) != openEnds(o) {
		return openEnds(w) < openEnds(o)
	}
	if w.openLo && !w.openHi && o.openLo && !o.openHi && w.hi != o.hi {
		return w.hi < o.hi
	}
	if w.openHi && !w.openLo && o.openHi && !o.openLo && w.lo != o.lo {
		return w.lo > o.lo
	}
	if !w.openLo && !w.openHi && !o.openLo && !o.openHi && w.hi-w.lo != o.hi-o.lo {
		return w.hi-w.lo < o.hi-o.lo
	}
	if w.qualified != o.qualified {
		return !w.qualified && o.qualified
	}
	return false
}

func openEnds(w dateWindow) int {
	n := 0
	if w.openLo {
		n++
	}
	if w.openHi {
		n++
	}
	return n
}

func windowOf(d datevalue.Value) (dateWindow, bool) {
	switch d.Kind {
	case datevalue.KindRange:
		return rangeWindow(d)
	default:
		return pointWindow(d)
	}
}

func pointWindow(d datevalue.Value) (dateWindow, bool) {
	lo, hi, prec, ok := spanOf(d.StartYear, d.StartMonth, d.StartDay, d.StartHour, d.StartMinute, d.StartSecond, d.StartMillisecond)
	if !ok {
		return dateWindow{}, false
	}
	w := dateWindow{lo: lo, hi: hi, precision: prec}
	switch d.Qualifier {
	case datevalue.QualifierBEF:
		w.openLo = true
		w.qualified = true
	case datevalue.QualifierAFT:
		w.openHi = true
		w.qualified = true
	case datevalue.QualifierABT:
		w.qualified = true
	case "":
	default:
		return dateWindow{}, false
	}
	return w, true
}

func rangeWindow(d datevalue.Value) (dateWindow, bool) {
	sLo, _, sPrec, sOK := spanOf(d.StartYear, d.StartMonth, d.StartDay, d.StartHour, d.StartMinute, d.StartSecond, d.StartMillisecond)
	_, eHi, ePrec, eOK := spanOf(d.EndYear, d.EndMonth, d.EndDay, d.EndHour, d.EndMinute, d.EndSecond, d.EndMillisecond)
	if !sOK || !eOK {
		return dateWindow{}, false
	}
	w := dateWindow{lo: sLo, hi: eHi, precision: min(sPrec, ePrec), qualified: true}
	if w.lo > w.hi {
		w.lo, w.hi = w.hi, w.lo
	}
	return w, true
}

// spanOf is the inclusive day span a bound covers at its precision, and that
// precision: 1 year, 2 month, 3 day, then one per clock field.
func spanOf(year, month, day, hour, minute, second, ms *int) (lo, hi, prec int, ok bool) {
	lo, hi, ok = datevalue.BoundDays(year, month, day)
	if !ok {
		return 0, 0, 0, false
	}
	prec = 1
	if month == nil {
		return lo, hi, prec, true
	}
	prec = 2
	if day == nil {
		return lo, hi, prec, true
	}
	prec = 3
	for _, c := range []*int{hour, minute, second, ms} {
		if c != nil {
			prec++
		}
	}
	return lo, hi, prec, true
}
