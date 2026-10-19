package valuecompare

import (
	"strconv"

	"github.com/mendahu/provenencia/core/datevalue"
)

// date compares dates as spans of days (comparison-layers.md §Layer 1). Each
// date is read by AboutEnds, so an ABT point is a range of AboutUnits of its
// most precise part. Equal spans are Same; one inside the other is Contains;
// partly shared is Overlaps; apart within Tolerance civil years is Near, and
// beyond it Different. A date with no year is Unknown.
type date struct{}

// dateSide is one date read for comparing.
type dateSide struct {
	first, last DateEnd
	window      Window
	plain       bool // a point with no qualifier
}

func readDate(d *datevalue.Value, r Rules) (dateSide, bool) {
	if d == nil {
		return dateSide{}, false
	}
	first, last, ok := AboutEnds(*d, r.Dates.AboutUnits)
	if !ok {
		return dateSide{}, false
	}
	return dateSide{
		first: first, last: last, window: EndsWindow(first, last),
		plain: d.Kind != datevalue.KindRange && d.Qualifier == "",
	}, true
}

func (date) compare(a, b Value, r Rules) (Result, Detail) {
	sa, okA := readDate(a.Date, r)
	sb, okB := readDate(b.Date, r)
	if !okA || !okB {
		failed := a.Date
		if okA {
			failed = b.Date
		}
		cause := CauseEmpty
		if failed != nil && failed.Phrase != "" && failed.StartYear == nil && failed.EndYear == nil {
			cause = CausePhraseOnly
		}
		return Result{Relation: Unknown}, Detail{Cause: cause}
	}
	d := r.Dates

	if sa.window.Contains(sb.window) && sb.window.Contains(sa.window) {
		return Result{Relation: Same, Similarity: 1}, Detail{Cause: CauseExact}
	}

	if sa.window.Overlaps(sb.window) {
		det := Detail{Cause: CauseOverlap}
		rel := Overlaps
		broader := sa
		switch {
		case sa.window.Contains(sb.window):
			rel, det.Cause, det.Broader = Contains, CausePrecision, SideA
		case sb.window.Contains(sa.window):
			rel, det.Cause, det.Broader = Contains, CausePrecision, SideB
			broader = sb
		}
		if sa.plain && sb.plain {
			if broader.first.Month == 0 {
				return Result{Relation: rel, Similarity: d.MissingMonth}, det
			}
			return Result{Relation: rel, Similarity: d.MissingDay}, det
		}
		return Result{Relation: rel, Similarity: d.spanScore(sa, sb)}, det
	}

	earlier, later := sa, sb
	if !sa.first.Open && !sb.last.Open && sa.first.Lo > sb.last.Hi {
		earlier, later = sb, sa
	}
	det := Detail{
		Distance: float64(later.first.Lo - earlier.last.Hi),
		Unit:     UnitDays,
	}
	if later.first.Year-earlier.last.Year > d.Tolerance {
		det.Cause = CauseBeyondTolerance
		return Result{Relation: Different}, det
	}
	det.Cause = CauseGap
	gap := later.first.Mid - earlier.last.Mid
	var sim float64
	switch {
	case sa.plain && sb.plain && earlier.last.Year == later.first.Year:
		sim = d.OtherMonth
		if earlier.last.Month == later.first.Month {
			sim = d.OtherDay
		}
	case sa.plain && sb.plain:
		sim = d.YearsApartFrom * falloff(gap, float64(d.Tolerance))
	default:
		sim = d.spanScore(sa, sb) * falloff(gap, float64(d.Tolerance))
	}
	if sim < d.NearMinimum {
		det.Cause = CauseBelowMinimum
		return Result{Relation: Different}, det
	}
	return Result{Relation: Near, Similarity: sim}, det
}

// spanScore is what two dates sharing a day are worth, judged by the wider
// span in years (at least one); an open span is SpanFloor.
func (d DateRules) spanScore(a, b dateSide) float64 {
	if a.first.Open || a.last.Open || b.first.Open || b.last.Open {
		return d.SpanFloor
	}
	w := max(a.last.To-a.first.From, b.last.To-b.first.From, 1)
	return max(d.SpanOneYear-d.SpanPerExtraYear*(w-1), d.SpanFloor)
}

// key is the date's window, lo..hi in day numbers, * for an open end. Equal
// windows are exactly Same; dates have no Equivalent.
func (date) key(v Value, _ Relation, r Rules) (string, bool) {
	s, ok := readDate(v.Date, r)
	if !ok {
		return "", false
	}
	end := func(n *int) string {
		if n == nil {
			return "*"
		}
		return strconv.Itoa(*n)
	}
	return end(s.window.Lo) + ".." + end(s.window.Hi), true
}

// YearReach is the civil years a date can still be Near to. Two dates Compare
// relates other than Different or Unknown have overlapping reaches, or one is
// open. ok is false when the date cannot be placed; open when either end is
// open.
func YearReach(v Value, r Rules) (lo, hi int, open, ok bool) {
	s, ok := readDate(v.Date, r)
	if !ok {
		return 0, 0, false, false
	}
	if s.first.Open || s.last.Open {
		return 0, 0, true, true
	}
	return s.first.Year - r.Dates.Tolerance, s.last.Year + r.Dates.Tolerance, false, true
}
