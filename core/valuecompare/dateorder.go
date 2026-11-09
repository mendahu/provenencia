package valuecompare

import "github.com/mendahu/provenencia/core/datevalue"

// The one date order: where a date sits among other dates. Timelines, birth
// order, a Source's date and the auto-reconciler's stored sort key all read it.

// DateKey is a date's place in date order. Day is the first day number of the
// date's place: its first end, or its last end when the first is open. Side
// breaks a tie on the same day: -1 for a date open at its start (BEF), 0 for a
// closed one, 1 for one open at its end (AFT), so "bef. 29 Mar 2011" sorts
// before "29 Mar 2011" and "aft. 29 Mar 2011" after.
type DateKey struct {
	Day, Side int
}

// DateKeyOf is d's place in date order, with the two ends it comes from (one
// DateEnds call). ok is false when d has no year to place, has both ends open,
// or has a closed end whose year is outside datevalue.MinYear…MaxYear.
func DateKeyOf(d datevalue.Value) (key DateKey, first, last DateEnd, ok bool) {
	first, last, ok = DateEnds(d)
	if !ok || (first.Open && last.Open) {
		return DateKey{}, DateEnd{}, DateEnd{}, false
	}
	for _, end := range []DateEnd{first, last} {
		if !end.Open && !datevalue.YearInWindow(end.Year) {
			return DateKey{}, DateEnd{}, DateEnd{}, false
		}
	}
	switch {
	case first.Open:
		key = DateKey{Day: last.Lo, Side: -1}
	case last.Open:
		key = DateKey{Day: first.Lo, Side: 1}
	default:
		key = DateKey{Day: first.Lo}
	}
	return key, first, last, true
}

// Compare orders k against o by Day, then Side: -1, 0 or 1.
func (k DateKey) Compare(o DateKey) int {
	switch {
	case k.Day != o.Day:
		if k.Day < o.Day {
			return -1
		}
		return 1
	case k.Side != o.Side:
		if k.Side < o.Side {
			return -1
		}
		return 1
	}
	return 0
}

// Int64 is k as one integer whose order is Compare's (datevalue.OrderInt).
func (k DateKey) Int64() int64 { return datevalue.OrderInt(k.Day, k.Side) }
