package match

import "github.com/mendahu/provenencia/core/database/datevalues"

// DateComparer compares dates as spans of years. Settings default to
// DefaultDates (registry.go), where each one is described.
//
//   - Identical dates score 1 at any precision ("1817" against "1817",
//     "BEF 1820" against "BEF 1820"); only differing precision costs.
//   - Two exact points (or ABT points) compare by year, refined by month and
//     day: MissingDay, MissingMonth, OtherDay, OtherMonth within the same
//     year; years apart fall off linearly from YearsApartFrom within
//     Tolerance; beyond it is a disagreement (0). ABT on either side
//     multiplies the tolerance by ApproxToleranceFactor.
//   - When either side is a range (FROM / TO / BET) or a bound (BEF, AFT),
//     overlapping spans are consistent, judged by the wider span:
//     SpanOneYear, less SpanPerExtraYear per extra year, never below
//     SpanFloor; an open span (BEF, AFT, FROM without TO) is SpanFloor. Spans
//     apart fall off from that within Tolerance; beyond it is a disagreement.
//     "BEF 1820" against 1823 is 3 years apart, not an approximate match.
//
// A date with no year (a phrase) is not comparable.
type DateComparer struct {
	Tolerance, ApproxToleranceFactor *int

	MissingDay, MissingMonth, OtherDay, OtherMonth, YearsApartFrom *float64

	SpanOneYear, SpanPerExtraYear, SpanFloor *float64
}

// dateRules are DateComparer settings resolved against the registry.
type dateRules struct {
	tol, approx                                               int
	missingDay, missingMonth, otherDay, otherMonth, apartFrom float64
	spanOne, spanStep, spanFloor                              float64
}

func (c DateComparer) resolve() dateRules {
	d := DefaultDates
	return dateRules{
		tol:          setting(c.Tolerance, *d.Tolerance),
		approx:       setting(c.ApproxToleranceFactor, *d.ApproxToleranceFactor),
		missingDay:   setting(c.MissingDay, *d.MissingDay),
		missingMonth: setting(c.MissingMonth, *d.MissingMonth),
		otherDay:     setting(c.OtherDay, *d.OtherDay),
		otherMonth:   setting(c.OtherMonth, *d.OtherMonth),
		apartFrom:    setting(c.YearsApartFrom, *d.YearsApartFrom),
		spanOne:      setting(c.SpanOneYear, *d.SpanOneYear),
		spanStep:     setting(c.SpanPerExtraYear, *d.SpanPerExtraYear),
		spanFloor:    setting(c.SpanFloor, *d.SpanFloor),
	}
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

// spanScore is what overlapping spans are worth, judged by the wider one.
func (r dateRules) spanScore(a, b yearSpan) float64 {
	width := func(s yearSpan) (int, bool) {
		if s.openLo || s.openHi {
			return 0, false
		}
		return s.hi - s.lo + 1, true
	}
	wa, okA := width(a)
	wb, okB := width(b)
	if !okA || !okB {
		return r.spanFloor
	}
	w := max(wa, wb)
	return max(r.spanOne-r.spanStep*float64(w-1), r.spanFloor)
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
	r := c.resolve()
	tol := r.tol
	gap := yearsApart(sa, sb)

	if !sa.point || !sb.point {
		if sa == sb {
			return 1, true
		}
		if gap > tol {
			return 0, true
		}
		return r.spanScore(sa, sb) * (1 - float64(gap)/float64(tol+1)), true
	}
	if sa.approx || sb.approx {
		tol *= r.approx
	}
	sameMonth := sa.month != nil && sb.month != nil && *sa.month == *sb.month
	switch {
	case gap == 0 && sameOptional(sa.month, sb.month) && sameOptional(sa.day, sb.day):
		return 1, true
	case gap == 0 && sameMonth && sa.day != nil && sb.day != nil:
		return r.otherDay, true
	case gap == 0 && sameMonth:
		return r.missingDay, true
	case gap == 0 && sa.month != nil && sb.month != nil:
		return r.otherMonth, true
	case gap == 0:
		return r.missingMonth, true
	case gap <= tol:
		return r.apartFrom * (1 - float64(gap)/float64(tol+1)), true
	}
	return 0, true
}
