package subjectcompare

import (
	"slices"

	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Side is the values Compare reads for lead on one Subject: the lead's own,
// its pooled properties', or, when the lead has a period and no positive value,
// the period built from its start and end. It copies.
func Side(kind vocabkey.SubjectType, lead vocabkey.PropertyKey, values Values, rules Rules) []propertycompare.Value {
	out := slices.Clone(values[lead])
	for _, with := range rules.Pools[kind][lead].With {
		out = append(out, values[with]...)
	}
	if p, ok := rules.Periods[kind][lead]; ok && !slices.ContainsFunc(out, propertycompare.Value.Positive) {
		if v, ok := period(values[p.Start], values[p.End]); ok {
			out = append(out, v)
		}
	}
	return out
}

// period is the widest span the positive start and end values record: the
// earliest start and the latest end, each picked by day number, so the order
// of the values never matters. A missing start or end is open. ok is false
// when neither is recorded or the span would end before it begins.
func period(starts, ends []propertycompare.Value) (propertycompare.Value, bool) {
	var firsts, lasts []valuecompare.DateEnd
	for _, v := range starts {
		if !v.Positive() || v.Date == nil {
			continue
		}
		if e, _, ok := valuecompare.DateEnds(*v.Date); ok {
			firsts = append(firsts, e)
		}
	}
	for _, v := range ends {
		if !v.Positive() || v.Date == nil {
			continue
		}
		if _, e, ok := valuecompare.DateEnds(*v.Date); ok {
			lasts = append(lasts, e)
		}
	}
	first, hasFirst := valuecompare.Earliest(firsts)
	last, hasLast := valuecompare.Latest(lasts)
	if !hasFirst && !hasLast {
		return propertycompare.Value{}, false
	}
	if !hasFirst {
		first = valuecompare.DateEnd{Open: true}
	}
	if !hasLast {
		last = valuecompare.DateEnd{Open: true}
	}
	d, ok := valuecompare.EndsDate(first, last)
	if !ok {
		return propertycompare.Value{}, false
	}
	return propertycompare.Value{Value: valuecompare.Value{Date: &d}}, true
}
