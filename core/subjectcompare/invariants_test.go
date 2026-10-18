package subjectcompare_test

import (
	"fmt"
	"maps"
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

const (
	invariantSeed  = 1
	invariantCases = 500
)

// group is a lead property and every property whose values reach it. pinned
// is the lead and its lifespan's span, the properties AC-3.5 fills.
type group struct {
	lead   vocabkey.PropertyKey
	props  []vocabkey.PropertyKey
	pinned []vocabkey.PropertyKey
}

// groups builds the lead groups of a kind from the registry: each lead of
// Weights[kind] in key order with the lead, then its pool, its period and its
// lifespan's span, and the pinned lead and span. A lead or property with no propertyPools entry fails.
func groups(t *testing.T, kind vocabkey.SubjectType) []group {
	t.Helper()
	s := matchregistry.Default().Subject
	var out []group
	for _, lead := range slices.Sorted(maps.Keys(s.Weights[kind])) {
		props := []vocabkey.PropertyKey{lead}
		pinned := []vocabkey.PropertyKey{lead}
		props = append(props, s.Pools[kind][lead].With...)
		if p, ok := s.Periods[kind][lead]; ok {
			props = append(props, p.Start, p.End)
		}
		if l, ok := s.Lifespans[kind][lead]; ok {
			props = append(props, l.Span)
			pinned = append(pinned, l.Span)
		}
		for _, p := range props {
			if len(propertyPools[p]) == 0 {
				t.Fatalf("%s.%s has no propertyPools entry", kind, p)
			}
		}
		out = append(out, group{lead: lead, props: props, pinned: pinned})
	}
	return out
}

var propertyPools = map[vocabkey.PropertyKey][]V{
	name:   {pt.Name("James Robins"), pt.Name("James Robbins"), pt.Name("J. Robins"), pt.Name("Mary Robins")},
	sex:    {pt.Term("male"), pt.Term("female"), pt.Neutral(pt.Term("unknown"))},
	gender: {pt.Term("male"), pt.Term("female"), pt.Neutral(pt.Term("unknown"))},
	etype:  {pt.Term("birth"), pt.Term("death"), pt.Term("baptism")},
	date:   {year(1817), year(1818), year(1830), pt.Date(datevaluetest.About(1817, 0, 0))},
	start:  {year(1815), year(1816)},
	end:    {year(1820)},
	topo:   {pt.Text("York"), pt.Text("york"), pt.Text("Montréal"), pt.Text("Montreal"), pt.Text("London")},
	lifeWindow: {pt.Date(datevaluetest.Between(1850, 0, 0, 1960, 0, 0)), pt.Date(datevaluetest.Between(1730, 0, 0, 1840, 0, 0)),
		pt.Date(datevaluetest.Between(1790, 0, 0, 1900, 0, 0))},
	lifeSpan: {pt.Date(datevaluetest.Between(1850, 0, 0, 1850, 0, 0)), pt.Date(datevaluetest.Between(1840, 0, 0, 1840, 0, 0)),
		pt.Date(datevaluetest.Between(1950, 0, 0, 1950, 0, 0))},
}

func properties(gs []group) []vocabkey.PropertyKey {
	var out []vocabkey.PropertyKey
	for _, g := range gs {
		out = append(out, g.props...)
	}
	return out
}

// draw gives each property of the kind 0 to 2 pool values, each negative one
// time in four.
func draw(r *rand.Rand, gs []group) Vals {
	vals := Vals{}
	for _, p := range properties(gs) {
		var vs []V
		for range r.Intn(3) {
			v := propertyPools[p][r.Intn(len(propertyPools[p]))]
			v.Negative = r.Intn(4) == 0
			vs = append(vs, v)
		}
		if len(vs) > 0 {
			vals[p] = vs
		}
	}
	return vals
}

func clone(vals Vals) Vals {
	out := Vals{}
	for k, v := range vals {
		out[k] = slices.Clone(v)
	}
	return out
}

func shuffled(r *rand.Rand, vals Vals) Vals {
	out := clone(vals)
	for _, vs := range out {
		r.Shuffle(len(vs), func(i, j int) { vs[i], vs[j] = vs[j], vs[i] })
	}
	return out
}

// empty reports whether the group records nothing in vals.
func empty(g group, vals Vals) bool {
	for _, p := range g.props {
		if len(vals[p]) > 0 {
			return false
		}
	}
	return true
}

func compare(kind vocabkey.SubjectType, a, b Vals) subjectcompare.Result {
	reg := matchregistry.Default()
	return subjectcompare.Compare(kind, a, b, subjectcomparetest.Metas(), reg.Subject, reg.Value)
}

func same(x, y subjectcompare.Result) string {
	if math.Abs(x.Similarity-y.Similarity) > 1e-9 || math.Abs(x.Evidence-y.Evidence) > 1e-9 {
		return fmt.Sprintf("similarity, evidence %v, %v became %v, %v", x.Similarity, x.Evidence, y.Similarity, y.Evidence)
	}
	if len(x.Properties) != len(y.Properties) {
		return "property count changed"
	}
	for i := range x.Properties {
		if x.Properties[i].Key != y.Properties[i].Key || x.Properties[i].Result.Outcome != y.Properties[i].Result.Outcome {
			return fmt.Sprintf("%s outcome %s became %s", x.Properties[i].Key, x.Properties[i].Result.Outcome, y.Properties[i].Result.Outcome)
		}
	}
	return ""
}

// invariant builds a case and returns what it found wrong, or "". ok is false
// when the case does not apply and must be drawn again.
type invariant func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (a, b Vals, msg string, ok bool)

func TestInvariants(t *testing.T) {
	kinds := []vocabkey.SubjectType{vocabkey.Person, vocabkey.Event, vocabkey.Place}
	byKind := map[vocabkey.SubjectType][]group{}
	for _, kind := range kinds {
		byKind[kind] = groups(t, kind)
	}
	invariants := []struct {
		id    string
		kinds []vocabkey.SubjectType
		run   invariant
	}{
		{"AC-3.1", kinds, func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (Vals, Vals, string, bool) {
			a, b := draw(r, gs), draw(r, gs)
			return a, b, same(compare(kind, a, b), compare(kind, b, a)), true
		}},
		{"AC-3.2", kinds, func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (Vals, Vals, string, bool) {
			a, b := draw(r, gs), draw(r, gs)
			return a, b, same(compare(kind, a, b), compare(kind, shuffled(r, a), shuffled(r, b))), true
		}},
		{"AC-3.3", kinds, func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (Vals, Vals, string, bool) {
			a, b := draw(r, gs), draw(r, gs)
			if s := compare(kind, a, b).Similarity; s < 0 || s > 1 {
				return a, b, fmt.Sprintf("similarity %v", s), true
			}
			return a, b, "", true
		}},
		{"AC-3.4", kinds, func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (Vals, Vals, string, bool) {
			a, b := draw(r, gs), draw(r, gs)
			var open []group
			for _, g := range gs {
				if empty(g, b) {
					open = append(open, g)
				}
			}
			if len(open) == 0 {
				return nil, nil, "", false
			}
			g := open[r.Intn(len(open))]
			p := g.props[r.Intn(len(g.props))]
			var v V
			for {
				v = propertyPools[p][r.Intn(len(propertyPools[p]))]
				if v.Positive() {
					break
				}
			}
			with := clone(a)
			with[p] = append(with[p], v)
			return a, b, same(compare(kind, a, b), compare(kind, with, b)), true
		}},
		{"AC-3.5", kinds, func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (Vals, Vals, string, bool) {
			a, b := draw(r, gs), draw(r, gs)
			var open []group
			for _, g := range gs {
				if empty(g, a) && empty(g, b) {
					open = append(open, g)
				}
			}
			if len(open) == 0 {
				return nil, nil, "", false
			}
			g := open[r.Intn(len(open))]
			wa, wb := clone(a), clone(b)
			for _, p := range g.pinned {
				wa[p], wb[p] = []V{propertyPools[p][0]}, []V{propertyPools[p][0]}
			}
			before, after := compare(kind, a, b), compare(kind, wa, wb)
			switch {
			case after.Evidence < before.Evidence-1e-9:
				return a, b, fmt.Sprintf("evidence %v went down to %v with %s", before.Evidence, after.Evidence, g.lead), true
			case after.Similarity < before.Similarity-1e-9:
				return a, b, fmt.Sprintf("similarity %v went down to %v with %s", before.Similarity, after.Similarity, g.lead), true
			}
			return a, b, "", true
		}},
		{"AC-3.6", []vocabkey.SubjectType{vocabkey.Person, vocabkey.Event}, func(r *rand.Rand, kind vocabkey.SubjectType, gs []group) (Vals, Vals, string, bool) {
			a, b := draw(r, gs), draw(r, gs)
			if kind == vocabkey.Person {
				a[name], b[name] = []V{pt.Name("James Robins")}, []V{pt.Name("Mary Robins")}
			} else {
				a[etype], b[etype] = []V{pt.Term("birth")}, []V{pt.Term("death")}
			}
			if ev := compare(kind, a, b).Evidence; ev >= 0 {
				return a, b, fmt.Sprintf("evidence %v, want below 0", ev), true
			}
			return a, b, "", true
		}},
	}
	for _, inv := range invariants {
		for _, kind := range inv.kinds {
			t.Run(inv.id+"/"+string(kind), func(t *testing.T) {
				r := rand.New(rand.NewSource(invariantSeed))
				counted := 0
				for i := 0; counted < invariantCases; i++ {
					if i > invariantCases*1000 {
						t.Fatalf("only %d applicable cases in %d draws", counted, i)
					}
					a, b, msg, ok := inv.run(r, kind, byKind[kind])
					if !ok {
						continue
					}
					counted++
					if msg != "" {
						t.Errorf("case %d: a=%s b=%s: %s", i, subjectcomparetest.Labels(a), subjectcomparetest.Labels(b), msg)
					}
				}
			})
		}
	}
}
