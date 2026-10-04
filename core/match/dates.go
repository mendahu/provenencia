package match

import "github.com/mendahu/provenencia/core/database/datevalues"

// Default DateComparer tolerance, in years.
const defaultDateTolerance = 2

// DateComparer compares dates as spans of years.
//
//   - Identical dates score 1 at any precision ("1817" against "1817",
//     "BEF 1820" against "BEF 1820"); only differing precision costs.
//   - Two exact points (or ABT points) compare by year, refined by month and
//     day: the same month with a day missing on one side 0.9; the same year
//     with a month missing on one side 0.8; the same month but different known days
//     0.7; different known months 0.6. Years apart within Tolerance fall off
//     linearly from 0.8; beyond it is a disagreement (0). ABT on either side
//     doubles the tolerance.
//   - When either side is a range (FROM / TO / BET) or a bound (BEF, AFT),
//     identical spans are 1. Otherwise overlapping spans are consistent, and
//     how much that says depends on the wider span: 0.8 for one year, 0.05
//     less per extra year, never below 0.2; an open span (BEF, AFT, FROM
//     without TO) says little and is 0.2. Spans apart fall off from that
//     score by the years between them within Tolerance; beyond it is a
//     disagreement. "BEF 1820" against 1823 is 3 years apart, not an
//     approximate match.
//
// A date with no year (a phrase) is not comparable.
type DateComparer struct {
	// Tolerance is how many years apart still resemble. Default 2; Set(0)
	// requires the same year (or overlapping spans).
	Tolerance *int
}

// yearSpan is a date's years, lo…hi; an open end is unbounded (BEF, AFT,
// FROM without TO). point marks an exact or ABT point, whose month and day
// refine the comparison.
type yearSpan struct {
	lo, hi         int
	openLo, openHi bool
	point, approx  bool
	month, day     *int
}

func spanOf(d *datevalues.Value) (yearSpan, bool) {
	if d.Kind == datevalues.KindRange {
		if d.StartYear == nil && d.EndYear == nil {
			return yearSpan{}, false
		}
		s := yearSpan{openLo: d.StartYear == nil, openHi: d.EndYear == nil}
		if d.StartYear != nil {
			s.lo = *d.StartYear
		}
		if d.EndYear != nil {
			s.hi = *d.EndYear
		}
		if !s.openLo && !s.openHi && s.lo > s.hi {
			s.lo, s.hi = s.hi, s.lo // an inverted range reads as its years
		}
		return s, true
	}
	y, m, day := d.StartYear, d.StartMonth, d.StartDay
	if y == nil {
		y, m, day = d.EndYear, d.EndMonth, d.EndDay
	}
	if y == nil {
		return yearSpan{}, false
	}
	switch d.Qualifier {
	case datevalues.QualifierBEF:
		return yearSpan{hi: *y, openLo: true}, true
	case datevalues.QualifierAFT:
		return yearSpan{lo: *y, openHi: true}, true
	}
	return yearSpan{lo: *y, hi: *y, point: true, approx: d.Qualifier != "", month: m, day: day}, true
}

// Span scoring: a one-year span is as good as a same-year point (0.8); each
// extra year of width costs spanStep, down to spanFloor, which open spans get.
const (
	spanBest  = 0.8
	spanStep  = 0.05
	spanFloor = 0.2
)

// spanScore is what overlapping spans are worth, judged by the wider one.
func spanScore(a, b yearSpan) float64 {
	width := func(s yearSpan) (int, bool) {
		if s.openLo || s.openHi {
			return 0, false
		}
		return s.hi - s.lo + 1, true
	}
	wa, okA := width(a)
	wb, okB := width(b)
	if !okA || !okB {
		return spanFloor
	}
	w := max(wa, wb)
	return max(spanBest-spanStep*float64(w-1), spanFloor)
}

// sameOptional: both missing, or both present and equal.
func sameOptional(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// yearsApart is the gap between two spans, 0 when they overlap.
func yearsApart(a, b yearSpan) int {
	switch {
	case !a.openHi && !b.openLo && a.hi < b.lo:
		return b.lo - a.hi
	case !b.openHi && !a.openLo && b.hi < a.lo:
		return a.lo - b.hi
	}
	return 0
}

func (c DateComparer) Compare(a, b Value) (float64, bool) {
	if a.Date == nil || b.Date == nil {
		return 0, false
	}
	sa, okA := spanOf(a.Date)
	sb, okB := spanOf(b.Date)
	if !okA || !okB {
		return 0, false
	}
	tol := setting(c.Tolerance, defaultDateTolerance)
	gap := yearsApart(sa, sb)

	if !sa.point || !sb.point {
		if sa == sb {
			return 1, true
		}
		if gap > tol {
			return 0, true
		}
		return spanScore(sa, sb) * (1 - float64(gap)/float64(tol+1)), true
	}
	if sa.approx || sb.approx {
		tol *= 2
	}
	sameMonth := sa.month != nil && sb.month != nil && *sa.month == *sb.month
	switch {
	case gap == 0 && sameOptional(sa.month, sb.month) && sameOptional(sa.day, sb.day):
		return 1, true
	case gap == 0 && sameMonth && sa.day != nil && sb.day != nil:
		return 0.7, true
	case gap == 0 && sameMonth:
		return 0.9, true
	case gap == 0 && sa.month != nil && sb.month != nil:
		return 0.6, true
	case gap == 0:
		return 0.8, true
	case gap <= tol:
		return 0.8 * (1 - float64(gap)/float64(tol+1)), true
	}
	return 0, true
}
