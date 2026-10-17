package autoreconcile

import (
	"github.com/mendahu/provenencia/core/datevalue"
)

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
