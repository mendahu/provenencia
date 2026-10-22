package subjectcompare_test

import (
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
	lifeWindow = vocabkey.PropertyLifespan
	lifeSpan   = vocabkey.PropertyLifespanSpan
)

// life is a person's window w1 to w2 and span s1 to s2, at year precision.
func life(w1, w2, s1, s2 int) Vals {
	return Vals{
		lifeWindow: {pt.Date(datevaluetest.Between(w1, 0, 0, w2, 0, 0))},
		lifeSpan:   {pt.Date(datevaluetest.Between(s1, 0, 0, s2, 0, 0))},
	}
}

// with adds the values of extra to v.
func with(v Vals, extra Vals) Vals {
	for k, x := range extra {
		v[k] = x
	}
	return v
}

func TestLifespan(t *testing.T) {
	robins := Vals{name: {pt.Name("James Robins")}}
	type row struct {
		name    string
		kind    vocabkey.SubjectType
		a, b    Vals
		sim, ev float64
		check   func(*testing.T, subjectcompare.Result)
	}
	rows := []row{
		{name: "AC-1.1 died before the other was born", kind: vocabkey.Person,
			a: life(1730, 1840, 1840, 1840), b: life(1850, 1960, 1850, 1850), sim: 0, ev: -20,
			check: is(lifeWindow, conflict, rel, -20)},
		{name: "AC-1.2 born before the other's span ended", kind: vocabkey.Person,
			a: life(1850, 1960, 1850, 1850), b: life(1880, 1990, 1990, 1990), sim: 0, ev: -20,
			check: is(lifeWindow, conflict, rel, -20)},
		{name: "AC-1.3 consistent lifespans agree", kind: vocabkey.Person,
			a: life(1850, 1960, 1850, 1850), b: life(1810, 1920, 1920, 1920), sim: 1, ev: 0.1,
			check: is(lifeWindow, agree, rel, 0.1)},
		{name: "AC-1.4 resided after the other died", kind: vocabkey.Person,
			a: life(1840, 2060, 1950, 1950), b: life(1790, 1900, 1900, 1900), sim: 0, ev: -20,
			check: is(lifeWindow, conflict, rel, -20)},
		{name: "AC-1.5 equal lifespans agree", kind: vocabkey.Person,
			a: life(1850, 1960, 1850, 1850), b: life(1850, 1960, 1850, 1850), sim: 1, ev: 0.1},
		{name: "AC-1.6 no lifespan on a side is unknown", kind: vocabkey.Person,
			a: life(1850, 1960, 1850, 1850), b: Vals{}, sim: 0, ev: 0,
			check: is(lifeWindow, unknown, noVals, 0)},
		{name: "AC-1.7 two years past the window is near", kind: vocabkey.Person,
			a: life(1850, 1960, 1850, 1850), b: life(1852, 1962, 1962, 1962), sim: 1, ev: 0.1,
			check: is(lifeWindow, agree, rel, 0.1)},
		{name: "AC-1.8 a conflicting lifespan weighs in similarity", kind: vocabkey.Person,
			a: with(life(1730, 1840, 1840, 1840), robins), b: with(life(1850, 1960, 1850, 1850), robins), sim: 2.5 / 22.5, ev: -17.5},
		{name: "AC-1.9 an event has no lifespan", kind: vocabkey.Event,
			a: with(life(1730, 1840, 1840, 1840), Vals{etype: {pt.Term("birth")}}),
			b: with(life(1850, 1960, 1850, 1850), Vals{etype: {pt.Term("birth")}}), sim: 1, ev: 0.5,
			check: func(t *testing.T, r subjectcompare.Result) {
				if !slices.Equal(keys(r), []vocabkey.PropertyKey{date, etype}) {
					t.Errorf("properties %v", keys(r))
				}
			}},
		{name: "AC-1.11 a span with no window to meet is unknown", kind: vocabkey.Person,
			a: life(1850, 1960, 1850, 1850),
			b: Vals{lifeSpan: {pt.Date(datevaluetest.Between(1850, 0, 0, 1850, 0, 0))}}, sim: 0, ev: 0,
			check: is(lifeWindow, unknown, noVals, 0)},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			got := subjectcompare.Compare(r.kind, r.a, r.b, subjectcomparetest.Metas(), matchregistry.Default().Subject, matchregistry.Default().Value)
			if !near(got.Similarity, r.sim) || !near(got.Evidence, r.ev) {
				t.Errorf("similarity %v evidence %v, want %v %v", got.Similarity, got.Evidence, r.sim, r.ev)
			}
			if r.check != nil {
				r.check(t, got)
			}
		})
	}
}

func TestFirstLifespanLead(t *testing.T) {
	got, ok := subjectcompare.FirstLifespanLead(map[vocabkey.PropertyKey]subjectcompare.Lifespan{"b": {}, "a": {}})
	if got != "a" || !ok {
		t.Fatalf("FirstLifespanLead = %q, %v; want a, true", got, ok)
	}
	if got, ok := subjectcompare.FirstLifespanLead(nil); got != "" || ok {
		t.Fatalf("FirstLifespanLead(nil) = %q, %v; want empty, false", got, ok)
	}
}
