package valuecompare

import "github.com/mendahu/provenencia/core/datevalue"

// Date ends and day windows: a date read as the days it covers. Pure calendar
// arithmetic lives in core/datevalue.

// DateEnd is one end of a date as written, for comparing dates by the days
// they cover. Month and Day are 0 when the end does not name them. Lo and Hi
// are its inclusive proleptic day numbers; From and To place it on a line of
// years, where a bound on day d of year y covers [y + (d-1)/n, y + d/n) with
// n the days in y, so a whole year is [y, y+1), and Mid is its centre. An open
// end has Open set and every other field zero.
type DateEnd struct {
	Year, Month, Day int
	Lo, Hi           int
	From, To, Mid    float64
	Open             bool
}

// DateEnds returns the first and last end of d. A point has the same end
// twice, BEF an open first end, AFT an open last end; a range has its start
// bound first and its end bound last (swapped when the start begins after the
// end finishes), and an open end on a side with no year. ok is false when d
// has no year to place, or a month or day that does not exist. Civil
// components are read as Gregorian.
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

// EndsDate is the date whose DateEnds are first and last: a range when both
// are closed, BEF last when first is open, AFT first when last is open. ok is
// false when both are open or first begins after last ends.
func EndsDate(first, last DateEnd) (datevalue.Value, bool) {
	switch {
	case first.Open && last.Open:
		return datevalue.Value{}, false
	case first.Open:
		return endsPoint(last, datevalue.QualifierBEF), true
	case last.Open:
		return endsPoint(first, datevalue.QualifierAFT), true
	case first.Lo > last.Hi:
		return datevalue.Value{}, false
	}
	return datevalue.Value{
		Kind: datevalue.KindRange, Calendar: "gregorian",
		StartYear: &first.Year, StartMonth: endPart(first.Month), StartDay: endPart(first.Day),
		EndYear: &last.Year, EndMonth: endPart(last.Month), EndDay: endPart(last.Day),
	}, true
}

// endsPoint is the point date at e with a BEF or AFT qualifier.
func endsPoint(e DateEnd, qualifier string) datevalue.Value {
	return datevalue.Value{
		Kind: datevalue.KindPoint, Calendar: "gregorian", Qualifier: qualifier,
		StartYear: &e.Year, StartMonth: endPart(e.Month), StartDay: endPart(e.Day),
	}
}

// endPart is a month or day for a datevalue.Value: nil for 0. A closed end
// always has a year, so the year never goes through it.
func endPart(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
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
	e.From, e.To, e.Mid = datevalue.YearLine(lo, hi)
	return e, true
}

// ShiftYears moves a closed end by whole civil years (datevalue.AddYears). An
// open end is returned as is. ok is false when the moved end cannot be placed.
func ShiftYears(e DateEnd, years int) (DateEnd, bool) {
	if e.Open {
		return e, true
	}
	y, m, d := datevalue.AddYears(e.Year, e.Month, e.Day, years)
	return boundEnd(&y, endPart(m), endPart(d))
}

// Earliest is the closed end that starts first: lowest Lo, then lowest Hi.
// Open ends are skipped; ok is false when none is closed.
func Earliest(ends []DateEnd) (best DateEnd, ok bool) {
	for _, e := range ends {
		if e.Open {
			continue
		}
		if !ok || e.Lo < best.Lo || (e.Lo == best.Lo && e.Hi < best.Hi) {
			best, ok = e, true
		}
	}
	return best, ok
}

// Latest is the closed end that finishes last: highest Hi, then highest Lo.
// Open ends are skipped; ok is false when none is closed.
func Latest(ends []DateEnd) (best DateEnd, ok bool) {
	for _, e := range ends {
		if e.Open {
			continue
		}
		if !ok || e.Hi > best.Hi || (e.Hi == best.Hi && e.Lo > best.Lo) {
			best, ok = e, true
		}
	}
	return best, ok
}

// Window is an inclusive proleptic-day span. A nil bound is open.
// A fully open window is "always" (no place period, undated place).
type Window struct {
	Lo, Hi *int
}

// Always is a fully open window (no period).
func Always() Window { return Window{} }

// Overlaps reports whether a and b share any day (open ends included).
func (a Window) Overlaps(b Window) bool {
	if a.Lo != nil && b.Hi != nil && *a.Lo > *b.Hi {
		return false
	}
	if b.Lo != nil && a.Hi != nil && *b.Lo > *a.Hi {
		return false
	}
	return true
}

// Intersect is the overlap of a and b. ok is false when empty.
func (a Window) Intersect(b Window) (Window, bool) {
	if !a.Overlaps(b) {
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
func (outer Window) Contains(inner Window) bool {
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

// AboutEnds reads d as compared: DateEnds, then an ABT point widens by units of
// its most precise part on each side. Year precision (no month; a day with no
// month reads as its year) widens by years, month precision by months, day
// precision by days. DateEnds stays as written.
func AboutEnds(d datevalue.Value, units int) (first, last DateEnd, ok bool) {
	first, last, ok = DateEnds(d)
	if !ok || d.Kind == datevalue.KindRange || d.Qualifier != datevalue.QualifierABT {
		return first, last, ok
	}
	y, m, day := first.Year, first.Month, first.Day
	switch {
	case m == 0:
		first, ok = boundEnd(intPtr(y-units), nil, nil)
		if ok {
			last, ok = boundEnd(intPtr(y+units), nil, nil)
		}
	case day == 0:
		y1, m1 := datevalue.AddMonths(y, m, -units)
		y2, m2 := datevalue.AddMonths(y, m, units)
		first, ok = boundEnd(intPtr(y1), intPtr(m1), nil)
		if ok {
			last, ok = boundEnd(intPtr(y2), intPtr(m2), nil)
		}
	default:
		y1, m1, d1 := datevalue.AddDays(y, m, day, -units)
		y2, m2, d2 := datevalue.AddDays(y, m, day, units)
		first, ok = boundEnd(intPtr(y1), intPtr(m1), intPtr(d1))
		if ok {
			last, ok = boundEnd(intPtr(y2), intPtr(m2), intPtr(d2))
		}
	}
	if !ok {
		return DateEnd{}, DateEnd{}, false
	}
	return first, last, true
}

func intPtr(v int) *int { return &v }

// EndsWindow is the days from the first end's start to the last end's finish;
// an open end leaves that side of the window open.
func EndsWindow(first, last DateEnd) Window {
	var w Window
	if !first.Open {
		w.Lo = intPtr(first.Lo)
	}
	if !last.Open {
		w.Hi = intPtr(last.Hi)
	}
	return w
}

// DateWindow is the days d covers as written: DateEnds as a Window, with an
// ABT point not widened. ok is false when DateEnds is not ok. Sort keys, stored
// date bounds and place periods read a date this way; comparing reads
// AboutEnds.
func DateWindow(d datevalue.Value) (Window, bool) {
	first, last, ok := DateEnds(d)
	if !ok {
		return Window{}, false
	}
	return EndsWindow(first, last), true
}
