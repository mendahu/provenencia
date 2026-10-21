package claimevidence

import (
	"bytes"
	"cmp"
	"maps"
	"slices"

	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// route says what an incoming value of one property is compared with: the
// section it counts under, the meta that types the comparison, and the weight
// its points come from. pooled marks a lead or pooled property, whose two
// sides are subjectcompare.Side; any other property compares the member's own
// values of it.
type route struct {
	lead     vocabkey.PropertyKey
	section  vocabkey.PropertyKey
	meta     vocabkey.PropertyKey
	pooled   bool
	weighted bool
	weights  subjectcompare.Weights
}

// routeOf decides, for a property of a kind, where its values are compared.
// A lead or a pooled property is compared through Side under the lead's meta;
// a period end is compared with the member's own values under its own meta but
// counts under the lead; any other property has a section of its own and no
// weight.
func routeOf(kind vocabkey.SubjectType, p vocabkey.PropertyKey, subject subjectcompare.Rules) route {
	lead, role, ok := subject.Role(kind, p)
	if !ok {
		return route{section: p, meta: p}
	}
	r := route{lead: lead, section: lead, weighted: true, weights: subject.Weights[kind][lead]}
	if role == subjectcompare.RolePeriodEnd {
		r.meta = p
		return r
	}
	r.pooled, r.meta = true, lead
	return r
}

// found is one incoming value's line against one member, before collapse.
type found struct {
	line   Line
	member int    // index of the member
	key    string // the member value's Same key; empty when it has none
}

// Sections compares each incoming value that names an Observation (a derived
// value has no Ref and gives no line) with each member's values of the same
// property, one line per member that says something. Members whose lines say
// the same thing about an incoming value are one line. A pooled or period
// property counts under its lead. Sections are in ascending key order, lines
// by incoming Ref then member order.
func Sections(kind vocabkey.SubjectType, incoming subjectcompare.Values, members []Member, metas subjectcompare.Metas,
	rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) []Section {
	bySection := map[vocabkey.PropertyKey][]found{}
	for _, p := range slices.Sorted(maps.Keys(incoming)) {
		r := routeOf(kind, p, subject)
		meta, ok := metas[r.meta]
		if !ok {
			continue
		}
		for _, v := range incoming[p] {
			if len(v.Ref) == 0 {
				continue
			}
			side := []propertycompare.Value{v}
			if r.pooled {
				side = subjectcompare.Side(kind, r.lead, subjectcompare.Values{p: side}, subject)
			}
			for i, m := range members {
				theirs := m.Values[p]
				if r.pooled {
					theirs = subjectcompare.Side(kind, r.lead, m.Values, subject)
				}
				res := propertycompare.Compare(side, theirs, meta, value)
				if res.Outcome == propertycompare.OutcomeUnknown || res.Speaking == nil || len(res.Speaking.Detail.RefB) == 0 {
					continue
				}
				bySection[r.section] = append(bySection[r.section], found{
					line:   line(p, v, res, m, r, rules),
					member: i,
					key:    memberKey(res.Speaking.B, value),
				})
			}
		}
	}
	var out []Section
	for _, key := range slices.Sorted(maps.Keys(bySection)) {
		out = append(out, Section{Key: key, Lines: collapse(bySection[key])})
	}
	return out
}

// line is one incoming value's verdict against one member, with the state the
// evidence section gives it.
func line(p vocabkey.PropertyKey, v propertycompare.Value, res propertycompare.Result, m Member, r route, rules Rules) Line {
	sp := res.Speaking
	l := Line{
		Property:      p,
		Incoming:      v,
		Outcome:       res.Outcome,
		PropertyCause: res.Cause,
		Relation:      sp.Result.Relation,
		Cause:         sp.Detail.Cause,
		Members:       []Pin{{Subject: m.Subject, Observation: sp.Detail.RefB, Source: m.Source, Value: sp.B}},
	}
	if r.weighted {
		l.Points = subjectcompare.Points(res, r.weights)
	}
	l.Checked = res.Outcome == propertycompare.OutcomeAgree && slices.Contains(rules.Checked, l.Relation)
	l.Pinnable = l.Checked || (res.Outcome == propertycompare.OutcomePartial && slices.Contains(rules.Pinnable, l.Relation))
	return l
}

// memberKey is the member value's Same key: two values with one key are
// identical. A value with no key gives "", and does not collapse.
func memberKey(v propertycompare.Value, value valuecompare.Rules) string {
	k, ok := valuecompare.Key(v.Value, valuecompare.Same, value)
	if !ok {
		return ""
	}
	return k
}

// collapse joins lines with the same incoming Observation, verdict and member
// value key into one, and sorts the result.
func collapse(fs []found) []Line {
	type first struct {
		line   Line
		member int
	}
	var firsts []first
	seen := map[string]int{}
	for _, f := range fs {
		if f.key != "" {
			id := string(f.line.Incoming.Ref) + "\x00" + string(f.line.Relation) + "\x00" + string(f.line.Cause) + "\x00" +
				string(f.line.Outcome) + "\x00" + string(f.line.PropertyCause) + "\x00" + f.key
			if at, ok := seen[id]; ok {
				firsts[at].line.Members = append(firsts[at].line.Members, f.line.Members...)
				continue
			}
			seen[id] = len(firsts)
		}
		firsts = append(firsts, first{f.line, f.member})
	}
	slices.SortStableFunc(firsts, func(a, b first) int {
		if c := bytes.Compare(a.line.Incoming.Ref, b.line.Incoming.Ref); c != 0 {
			return c
		}
		return cmp.Compare(a.member, b.member)
	})
	out := make([]Line, len(firsts))
	for i, f := range firsts {
		out[i] = f.line
	}
	return out
}
