package subjectcompare_test

import (
	"math"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type (
	V    = propertycompare.Value
	Vals = subjectcompare.Values
)

const (
	name   = vocabkey.PropertyName
	sex    = vocabkey.PropertySexAtBirth
	gender = vocabkey.PropertyGender
	etype  = vocabkey.PropertyEventType
	date   = vocabkey.PropertyDate
	start  = vocabkey.PropertyStartDate
	end    = vocabkey.PropertyEndDate
	topo   = vocabkey.PropertyToponym
)

func year(y int) V { return pt.Date(datevaluetest.YMD(y, 0, 0)) }

func keys(r subjectcompare.Result) []vocabkey.PropertyKey {
	var out []vocabkey.PropertyKey
	for _, p := range r.Properties {
		out = append(out, p.Key)
	}
	return out
}

func prop(t *testing.T, r subjectcompare.Result, key vocabkey.PropertyKey) subjectcompare.PropertyResult {
	t.Helper()
	for _, p := range r.Properties {
		if p.Key == key {
			return p
		}
	}
	t.Fatalf("no result for %s in %v", key, keys(r))
	return subjectcompare.PropertyResult{}
}

func near(a, b float64) bool { return math.Abs(a-b) < 0.0001 }

// is checks a property's outcome, cause and points.
func is(key vocabkey.PropertyKey, out propertycompare.Outcome, cause propertycompare.Cause, points float64) func(*testing.T, subjectcompare.Result) {
	return func(t *testing.T, r subjectcompare.Result) {
		t.Helper()
		p := prop(t, r, key)
		if p.Result.Outcome != out || p.Result.Cause != cause || !near(p.Points, points) {
			t.Errorf("%s: outcome %s cause %s points %v, want %s %s %v", key, p.Result.Outcome, p.Result.Cause, p.Points, out, cause, points)
		}
	}
}

func all(checks ...func(*testing.T, subjectcompare.Result)) func(*testing.T, subjectcompare.Result) {
	return func(t *testing.T, r subjectcompare.Result) {
		t.Helper()
		for _, c := range checks {
			c(t, r)
		}
	}
}

const (
	agree    = propertycompare.OutcomeAgree
	partial  = propertycompare.OutcomePartial
	conflict = propertycompare.OutcomeConflict
	unknown  = propertycompare.OutcomeUnknown
	rel      = propertycompare.CauseRelation
	neg      = propertycompare.CauseNegative
	noVals   = propertycompare.CauseNoValues
)

func TestCompare(t *testing.T) {
	noName := subjectcomparetest.Metas()
	delete(noName, name)
	type row struct {
		name     string
		kind     vocabkey.SubjectType
		a, b     Vals
		metas    subjectcompare.Metas
		sim, ev  float64
		check    func(*testing.T, subjectcompare.Result)
		nilProps bool
	}
	robins, mary := pt.Name("James Robins"), pt.Name("Mary Robins")
	rows := []row{
		{name: "AC-1.1 name agrees", kind: vocabkey.Person,
			a: Vals{name: {robins}}, b: Vals{name: {robins}}, sim: 1, ev: 2.5,
			check: all(is(name, agree, rel, 2.5), is(sex, unknown, noVals, 0))},
		{name: "AC-1.2 gender counts as sex at birth", kind: vocabkey.Person,
			a: Vals{name: {robins}, gender: {pt.Term("woman")}}, b: Vals{name: {robins}, gender: {pt.Term("woman")}}, sim: 1, ev: 3.1},
		{name: "AC-1.3 name agrees, sex conflicts", kind: vocabkey.Person,
			a: Vals{name: {robins}, sex: {pt.Term("male")}}, b: Vals{name: {robins}, sex: {pt.Term("female")}}, sim: 2.5 / 3.5, ev: 1.5},
		{name: "AC-1.4 name partial", kind: vocabkey.Person,
			a: Vals{name: {pt.Name("J. Robins")}}, b: Vals{name: {robins}}, sim: 0.8, ev: 2.0,
			check: is(name, partial, rel, 2.0)},
		{name: "AC-1.5 name conflict", kind: vocabkey.Person,
			a: Vals{name: {robins}}, b: Vals{name: {mary}}, sim: 0, ev: -3,
			check: is(name, conflict, rel, -3)},
		{name: "AC-1.6 event type agrees", kind: vocabkey.Event,
			a: Vals{etype: {pt.Term("birth")}}, b: Vals{etype: {pt.Term("birth")}}, sim: 1, ev: 0.5},
		{name: "AC-1.7 birth against death on one date", kind: vocabkey.Event,
			a: Vals{etype: {pt.Term("birth")}, date: {year(1817)}}, b: Vals{etype: {pt.Term("death")}, date: {year(1817)}}, sim: 2.0 / 22, ev: -18},
		{name: "AC-1.8 birth against death", kind: vocabkey.Event,
			a: Vals{etype: {pt.Term("birth")}}, b: Vals{etype: {pt.Term("death")}}, sim: 0, ev: -20},
		{name: "AC-1.9 toponym agrees at full weight", kind: vocabkey.Place,
			a: Vals{topo: {pt.Text("Montréal")}}, b: Vals{topo: {pt.Text("Montreal")}}, sim: 1, ev: 3,
			check: is(topo, agree, rel, 3)},
		{name: "AC-1.10 a different toponym is unknown", kind: vocabkey.Place,
			a: Vals{topo: {pt.Text("York")}}, b: Vals{topo: {pt.Text("London")}}, sim: 0, ev: 0,
			check: is(topo, unknown, propertycompare.CauseCardinality, 0)},
		{name: "AC-1.11 an unrecorded name counts for nothing", kind: vocabkey.Person,
			a: Vals{name: {robins}, sex: {pt.Term("male")}}, b: Vals{sex: {pt.Term("male")}}, sim: 1, ev: 0.6,
			check: is(name, unknown, noVals, 0)},
		{name: "AC-1.12 both sides empty", kind: vocabkey.Person, sim: 0, ev: 0,
			check: func(t *testing.T, r subjectcompare.Result) {
				if !slices.Equal(keys(r), []vocabkey.PropertyKey{vocabkey.PropertyLifespan, name, sex}) {
					t.Errorf("properties %v", keys(r))
				}
				is(vocabkey.PropertyLifespan, unknown, noVals, 0)(t, r)
				is(name, unknown, noVals, 0)(t, r)
				is(sex, unknown, noVals, 0)(t, r)
			}},
		{name: "AC-1.13 a kind with no entry", kind: "relationship",
			a: Vals{name: {robins}}, b: Vals{name: {robins}}, nilProps: true},
		{name: "AC-1.14 a lead with no meta is unsupported", kind: vocabkey.Person, metas: noName,
			a: Vals{name: {robins}}, b: Vals{name: {robins}}, sim: 0, ev: 0,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(name, unknown, rel, 0)(t, r)
				p := prop(t, r, name)
				if p.Result.Speaking == nil || p.Result.Speaking.Result.Relation != valuecompare.Unknown ||
					p.Result.Speaking.Detail.Cause != valuecompare.CauseUnsupported {
					t.Errorf("speaking %+v", p.Result.Speaking)
				}
			}},
		{name: "AC-1.15 an unlisted property is never read", kind: vocabkey.Event,
			a: Vals{"event_name": {pt.Text("Great Fire")}}, b: Vals{"event_name": {pt.Text("Great Fire")}}, sim: 0, ev: 0,
			check: func(t *testing.T, r subjectcompare.Result) {
				if !slices.Equal(keys(r), []vocabkey.PropertyKey{date, etype}) {
					t.Errorf("properties %v", keys(r))
				}
				is(date, unknown, noVals, 0)(t, r)
				is(etype, unknown, noVals, 0)(t, r)
			}},

		{name: "AC-2.1 gender against sex at birth agrees", kind: vocabkey.Person,
			a: Vals{gender: {pt.Term("woman")}}, b: Vals{sex: {pt.Term("female")}}, sim: 1, ev: 0.6,
			check: is(sex, agree, rel, 0.6)},
		{name: "AC-2.1 the gender man against the sex male agrees", kind: vocabkey.Person,
			a: Vals{gender: {pt.Term("man")}}, b: Vals{sex: {pt.Term("male")}}, sim: 1, ev: 0.6,
			check: is(sex, agree, rel, 0.6)},
		{name: "AC-2.1 a side recording both sex and gender agrees with sex alone", kind: vocabkey.Person,
			a: Vals{sex: {pt.Term("male")}, gender: {pt.Term("man")}}, b: Vals{sex: {pt.Term("male")}}, sim: 1, ev: 0.6,
			check: is(sex, agree, rel, 0.6)},
		{name: "AC-2.1 a gender the pool does not map is left out", kind: vocabkey.Person,
			a: Vals{gender: {pt.Term("non_binary")}}, b: Vals{sex: {pt.Term("male")}}, sim: 0, ev: 0,
			check: is(sex, unknown, noVals, 0)},
		{name: "AC-2.2 sex at birth against gender conflicts", kind: vocabkey.Person,
			a: Vals{sex: {pt.Term("male")}}, b: Vals{gender: {pt.Term("woman")}}, sim: 0, ev: -1,
			check: is(sex, conflict, rel, -1)},
		{name: "AC-2.3 an unknown sex is excluded", kind: vocabkey.Person,
			a: Vals{sex: {pt.Neutral(pt.Term("unknown"))}}, b: Vals{sex: {pt.Term("male")}}, sim: 0, ev: 0,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(sex, unknown, noVals, 0)(t, r)
				ex := prop(t, r, sex).Result.Excluded
				if len(ex) != 1 || ex[0].Value.Term != "unknown" {
					t.Errorf("excluded %+v", ex)
				}
			}},
		{name: "AC-2.4 a negative sex contradicts gender", kind: vocabkey.Person,
			a: Vals{sex: {pt.Neg(pt.Term("male"))}}, b: Vals{gender: {pt.Term("man")}}, sim: 0, ev: -1,
			check: is(sex, conflict, neg, -1)},
		{name: "AC-2.5 a date against a period", kind: vocabkey.Event,
			a: Vals{date: {year(1817)}}, b: Vals{start: {year(1815)}, end: {year(1820)}}, sim: 0.55, ev: 1.1,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(date, partial, rel, 1.1)(t, r)
				s := prop(t, r, date).Result.Speaking
				if got := datevalue.CompactDisplay(*s.B.Date); got != "Between 1815 and 1820" || s.B.Ref != nil ||
					s.Result.Relation != valuecompare.Contains || !near(s.Result.Similarity, 0.55) {
					t.Errorf("speaking B %q ref %v relation %s %v", got, s.B.Ref, s.Result.Relation, s.Result.Similarity)
				}
			}},
		{name: "AC-2.6 two periods agree", kind: vocabkey.Event,
			a: Vals{start: {year(1815)}, end: {year(1820)}}, b: Vals{start: {year(1815)}, end: {year(1820)}}, sim: 1, ev: 2,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(date, agree, rel, 2)(t, r)
				if rl := prop(t, r, date).Result.Speaking.Result.Relation; rl != valuecompare.Same {
					t.Errorf("relation %s", rl)
				}
			}},
		{name: "AC-2.7 a recorded date speaks for itself", kind: vocabkey.Event,
			a: Vals{date: {year(1817)}, start: {year(1700)}, end: {year(1701)}}, b: Vals{date: {year(1817)}}, sim: 1, ev: 2,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(date, agree, rel, 2)(t, r)
				if rl := prop(t, r, date).Result.Speaking.Result.Relation; rl != valuecompare.Same {
					t.Errorf("relation %s", rl)
				}
			}},
		{name: "AC-2.7 a recorded date keeps its period out", kind: vocabkey.Event,
			a: Vals{date: {year(1817)}, start: {year(1815)}, end: {year(1820)}}, b: Vals{date: {pt.Neg(year(1818))}}, sim: 0, ev: 0,
			check: is(date, unknown, noVals, 0)},
		{name: "AC-2.8 a denied start is skipped", kind: vocabkey.Event,
			a: Vals{start: {pt.Neg(year(1815))}, end: {year(1820)}}, b: Vals{date: {year(1817)}}, sim: 0.2, ev: 0.4,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(date, partial, rel, 0.4)(t, r)
				if got := datevalue.CompactDisplay(*prop(t, r, date).Result.Speaking.A.Date); got != "Before 1820" {
					t.Errorf("speaking A %q", got)
				}
			}},
		{name: "AC-2.8 a start alone reads as after", kind: vocabkey.Event,
			a: Vals{start: {year(1815)}}, b: Vals{date: {year(1817)}}, sim: 0.2, ev: 0.4,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(date, partial, rel, 0.4)(t, r)
				s := prop(t, r, date).Result.Speaking
				if got := datevalue.CompactDisplay(*s.A.Date); got != "After 1815" || s.Result.Relation != valuecompare.Contains || !near(s.Result.Similarity, 0.2) {
					t.Errorf("speaking A %q %s %v", got, s.Result.Relation, s.Result.Similarity)
				}
			}},
		{name: "AC-2.9 an end alone reads as before", kind: vocabkey.Event,
			a: Vals{end: {year(1820)}}, b: Vals{date: {year(1817)}}, sim: 0.2, ev: 0.4,
			check: func(t *testing.T, r subjectcompare.Result) {
				is(date, partial, rel, 0.4)(t, r)
				s := prop(t, r, date).Result.Speaking
				if got := datevalue.CompactDisplay(*s.A.Date); got != "Before 1820" || s.Result.Relation != valuecompare.Contains || !near(s.Result.Similarity, 0.2) {
					t.Errorf("speaking A %q %s %v", got, s.Result.Relation, s.Result.Similarity)
				}
			}},
		{name: "AC-2.10 an end before its start builds no period", kind: vocabkey.Event,
			a: Vals{start: {year(1821)}, end: {year(1820)}}, b: Vals{date: {year(1817)}}, sim: 0, ev: 0,
			check: is(date, unknown, noVals, 0)},
		{name: "AC-2.11 the earliest start is used", kind: vocabkey.Event,
			a: Vals{start: {year(1816), year(1815)}, end: {year(1820)}}, b: Vals{date: {year(1815)}}, sim: 0.55, ev: 1.1,
			check: is(date, partial, rel, 1.1)},
		{name: "AC-2.11 the earliest start is used in either order", kind: vocabkey.Event,
			a: Vals{start: {year(1815), year(1816)}, end: {year(1820)}}, b: Vals{date: {year(1815)}}, sim: 0.55, ev: 1.1,
			check: is(date, partial, rel, 1.1)},
		{name: "AC-2.12 a denied date still contradicts", kind: vocabkey.Event,
			a: Vals{date: {pt.Neg(year(1817))}, start: {year(1815)}, end: {year(1820)}}, b: Vals{date: {year(1817)}}, sim: 0, ev: -3,
			check: is(date, conflict, neg, -3)},
		{name: "AC-2.13 a period against a different date", kind: vocabkey.Event,
			a: Vals{start: {year(1815)}, end: {year(1820)}}, b: Vals{date: {year(1830)}}, sim: 0, ev: -3,
			check: is(date, conflict, rel, -3)},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			metas := r.metas
			if metas == nil {
				metas = subjectcomparetest.Metas()
			}
			got := subjectcompare.Compare(r.kind, r.a, r.b, metas, matchregistry.Default().Subject, matchregistry.Default().Value)
			if !near(got.Similarity, r.sim) || !near(got.Evidence, r.ev) {
				t.Errorf("similarity %v evidence %v, want %v %v", got.Similarity, got.Evidence, r.sim, r.ev)
			}
			if r.nilProps && got.Properties != nil {
				t.Errorf("properties %v, want nil", got.Properties)
			}
			if r.check != nil {
				r.check(t, got)
			}
		})
	}
}

func TestPoints(t *testing.T) {
	w := subjectcompare.Weights{Agree: 2.5, Conflict: 3}
	speaking := func(sim float64) *propertycompare.SpeakingPair {
		return &propertycompare.SpeakingPair{Result: valuecompare.Result{Similarity: sim}}
	}
	for _, c := range []struct {
		name string
		res  propertycompare.Result
		want float64
	}{
		{"agree", propertycompare.Result{Outcome: propertycompare.OutcomeAgree, Speaking: speaking(0.4)}, 2.5},
		{"partial", propertycompare.Result{Outcome: propertycompare.OutcomePartial, Speaking: speaking(0.75)}, 1.875},
		{"conflict", propertycompare.Result{Outcome: propertycompare.OutcomeConflict}, -3},
		{"unknown", propertycompare.Result{Outcome: propertycompare.OutcomeUnknown}, 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := subjectcompare.Points(c.res, w); !near(got, c.want) {
				t.Errorf("Points = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDerived(t *testing.T) {
	rules := matchregistry.Default().Subject
	for _, c := range []struct {
		kind vocabkey.SubjectType
		key  vocabkey.PropertyKey
		want bool
	}{
		{vocabkey.Person, vocabkey.PropertyLifespan, true},
		{vocabkey.Person, vocabkey.PropertyLifespanSpan, true},
		{vocabkey.Person, name, false},
		{vocabkey.Event, vocabkey.PropertyLifespan, false},
	} {
		if got := rules.Derived(c.kind, c.key); got != c.want {
			t.Errorf("Derived(%s, %s) = %v, want %v", c.kind, c.key, got, c.want)
		}
	}
}

func TestLead(t *testing.T) {
	rules := matchregistry.Default().Subject
	for _, c := range []struct {
		kind vocabkey.SubjectType
		key  vocabkey.PropertyKey
		want vocabkey.PropertyKey
		ok   bool
	}{
		{vocabkey.Person, name, name, true},
		{vocabkey.Person, gender, sex, true},
		{vocabkey.Event, start, date, true},
		{vocabkey.Event, end, date, true},
		{vocabkey.Person, vocabkey.PropertyLifespan, vocabkey.PropertyLifespan, true},
		{vocabkey.Person, vocabkey.PropertyLifespanSpan, "", false},
		{vocabkey.Person, vocabkey.PropertyKey("occupation"), "", false},
		{vocabkey.Event, name, "", false},
		{vocabkey.SubjectType("source"), name, "", false},
	} {
		got, ok := rules.Lead(c.kind, c.key)
		if got != c.want || ok != c.ok {
			t.Errorf("Lead(%s, %s) = %q, %v; want %q, %v", c.kind, c.key, got, ok, c.want, c.ok)
		}
	}
}
