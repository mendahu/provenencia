package match

import (
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

// Period pooling: an Event's date is compared with another Event's start and
// end when that is all the other has. Score and evaluate call poolPeriods on
// each Feature or meta before pairValues, so a pooled side speaks for the
// date Property with one period value and everything downstream is unchanged.

// poolPeriods returns the probe's and the candidate's values to compare for
// prop. For a Property with no DatePeriods entry, or when neither side pools,
// they are the values as given. A side pools exactly when the other has a
// voting prop value and no voting start or end, and it has no voting prop value
// and a voting start or end. At most one side can. The period is appended to
// that side's values; its own non-voting values stay where they were.
func poolPeriods(prop Property, probe, candidate Values) (a, b []Value) {
	a, b = probe[prop], candidate[prop]
	pp, ok := DatePeriods[prop]
	if !ok {
		return a, b
	}
	switch {
	case pools(prop, pp, candidate, probe):
		if p, ok := period(pp, candidate); ok {
			b = append(append(make([]Value, 0, len(b)+1), b...), p)
		}
	case pools(prop, pp, probe, candidate):
		if p, ok := period(pp, probe); ok {
			a = append(append(make([]Value, 0, len(a)+1), a...), p)
		}
	}
	return a, b
}

// pools reports whether the side with values this pools against other.
func pools(prop Property, pp PeriodProperties, this, other Values) bool {
	return hasVoting(other[prop]) && !hasVoting(other[pp.Start]) && !hasVoting(other[pp.End]) &&
		!hasVoting(this[prop]) && (hasVoting(this[pp.Start]) || hasVoting(this[pp.End]))
}

func hasVoting(vs []Value) bool {
	for i := range vs {
		if vs[i].votes() {
			return true
		}
	}
	return false
}

// firstLeader is the index of the value that speaks for a single-cardinality
// Property: the first leader in tie order; -1 when none votes.
func firstLeader(vs []Value) int {
	info := weigh(vs, properties.CardinalitySingle)
	best := -1
	for i := range vs {
		if info[i].leader && (best < 0 || order(vs, info, i, best) < 0) {
			best = i
		}
	}
	return best
}

// period builds the one date value a side's start and end make: a range when
// it has both, AFT the start or BEF the end when it has one. An open start
// (BEF) or end (AFT), a value with no date, and a start after the end count as
// absent. The qualifiers on the sources are dropped.
func period(pp PeriodProperties, vs Values) (Value, bool) {
	var start, end autoreconcile.DateEnd
	var hasStart, hasEnd bool
	if i := firstLeader(vs[pp.Start]); i >= 0 && vs[pp.Start][i].Date != nil {
		if first, _, ok := autoreconcile.DateEnds(*vs[pp.Start][i].Date); ok && !first.Open {
			start, hasStart = first, true
		}
	}
	if i := firstLeader(vs[pp.End]); i >= 0 && vs[pp.End][i].Date != nil {
		if _, last, ok := autoreconcile.DateEnds(*vs[pp.End][i].Date); ok && !last.Open {
			end, hasEnd = last, true
		}
	}
	d := datevalues.Value{Calendar: "gregorian"}
	switch {
	case hasStart && hasEnd:
		if start.Lo > end.Hi {
			return Value{}, false
		}
		d.Kind = datevalues.KindRange
		d.StartYear, d.StartMonth, d.StartDay = endFields(start)
		d.EndYear, d.EndMonth, d.EndDay = endFields(end)
	case hasStart:
		d.Kind, d.Qualifier = datevalues.KindPoint, datevalues.QualifierAFT
		d.StartYear, d.StartMonth, d.StartDay = endFields(start)
	case hasEnd:
		d.Kind, d.Qualifier = datevalues.KindPoint, datevalues.QualifierBEF
		d.StartYear, d.StartMonth, d.StartDay = endFields(end)
	default:
		return Value{}, false
	}
	return Value{Date: &d}, true
}

// endFields is a date end as the year, month, and day of a value; a month or
// day the end does not name is nil.
func endFields(e autoreconcile.DateEnd) (year, month, day *int) {
	year = &e.Year
	if e.Month != 0 {
		month = &e.Month
	}
	if e.Day != 0 {
		day = &e.Day
	}
	return year, month, day
}
