package match

import (
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// DateComparer compares dates by the days they cover. Settings default to
// DefaultDates (registry.go), where each one is described. Tolerance counts
// civil years: the cutoff reads the year numbers of the two dates. The
// fall-off inside it counts days, as a gap between the dates' centres in
// years, so December 1817 and January 1818 are a month apart, not a year.
//
//   - Identical dates score 1 at any precision ("1817" against "1817",
//     "BEF 1820" against "BEF 1820"); only differing precision costs.
//   - Two exact points (or ABT points) in one civil year take a tier by month
//     and day: MissingDay, MissingMonth, OtherDay, OtherMonth. In different
//     years they fall off linearly from YearsApartFrom within Tolerance;
//     beyond it is a disagreement (0). ABT on either side multiplies the
//     tolerance by ApproxToleranceFactor. The score is not continuous at the
//     year boundary: 31 December against 1 January is about 0.799, two days
//     in one month 0.7.
//   - When either side is a range (FROM / TO / BET) or a bound (BEF, AFT),
//     dates sharing a day are consistent, judged by the wider span:
//     SpanOneYear, less SpanPerExtraYear per extra year (a span under a year
//     counts as one), never below SpanFloor; an open span (BEF, AFT, FROM
//     without TO) is SpanFloor. Dates apart fall off from that within
//     Tolerance; beyond it is a disagreement. "BEF 1820" against 1823 is 3
//     years apart, not an approximate match.
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

// dateSpan is a date's first and last end. An open end is unbounded (BEF,
// AFT, FROM without TO). point marks an exact or ABT point, whose month and
// day refine the comparison. It has no pointer fields, so == means equal by
// value.
type dateSpan struct {
	first, last   valuecompare.DateEnd
	point, approx bool
}

func spanOf(d *datevalue.Value) (dateSpan, bool) {
	first, last, ok := valuecompare.DateEnds(*d)
	if !ok {
		return dateSpan{}, false
	}
	s := dateSpan{first: first, last: last}
	if d.Kind != datevalue.KindRange && d.Qualifier != datevalue.QualifierBEF && d.Qualifier != datevalue.QualifierAFT {
		s.point, s.approx = true, d.Qualifier != ""
	}
	return s, true
}

func centre(e valuecompare.DateEnd) float64 { return (e.From + e.To) / 2 }

// width is the years a closed span covers; ok is false for an open span.
func (s dateSpan) width() (float64, bool) {
	if s.first.Open || s.last.Open {
		return 0, false
	}
	return s.last.To - s.first.From, true
}

// shares reports whether two spans have a day in common. An open end reaches
// past every date.
func shares(a, b dateSpan) bool {
	return (a.last.Open || b.first.Open || b.first.Lo <= a.last.Hi) &&
		(b.last.Open || a.first.Open || a.first.Lo <= b.last.Hi)
}

// spanScore is what overlapping spans are worth, judged by the wider one; a
// span under a year counts as one year.
func (r dateRules) spanScore(a, b dateSpan) float64 {
	wa, okA := a.width()
	wb, okB := b.width()
	if !okA || !okB {
		return r.spanFloor
	}
	w := max(wa, wb, 1)
	return max(r.spanOne-r.spanStep*(w-1), r.spanFloor)
}

// apart is the gap between two spans that share no day: the later one's first
// end less the earlier one's last end, as civil years and as years between
// their centres. Both are 0 when the spans share a day.
func apart(a, b dateSpan) (years int, gap float64) {
	if shares(a, b) {
		return 0, 0
	}
	if !a.last.Open && !b.first.Open && a.last.Hi < b.first.Lo {
		return b.first.Year - a.last.Year, centre(b.first) - centre(a.last)
	}
	return a.first.Year - b.last.Year, centre(a.first) - centre(b.last)
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

	if !sa.point || !sb.point {
		if sa == sb {
			return 1, true
		}
		years, gap := apart(sa, sb)
		if years > tol {
			return 0, true
		}
		return r.spanScore(sa, sb) * (1 - gap/float64(tol+1)), true
	}
	if sa.approx || sb.approx {
		tol *= r.approx
	}
	ea, eb := sa.first, sb.first
	if ea.Year != eb.Year {
		years := max(ea.Year-eb.Year, eb.Year-ea.Year)
		if years > tol {
			return 0, true
		}
		gap := max(centre(ea)-centre(eb), centre(eb)-centre(ea))
		return r.apartFrom * (1 - gap/float64(tol+1)), true
	}
	sameMonth := ea.Month != 0 && ea.Month == eb.Month
	switch {
	case ea.Month == eb.Month && ea.Day == eb.Day:
		return 1, true
	case sameMonth && ea.Day != 0 && eb.Day != 0:
		return r.otherDay, true
	case sameMonth:
		return r.missingDay, true
	case ea.Month != 0 && eb.Month != 0:
		return r.otherMonth, true
	}
	return r.missingMonth, true
}
