package match

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/datevalue"
)

// name is a structured name. The form is a transcription and is not compared;
// the words become given names and a surname, the way a researcher files parts.
func name(form string) Value {
	n := &namevalues.Value{Form: form}
	words := strings.Fields(form)
	for i, w := range words {
		typ := namevalues.PartTypeGiven
		if i == len(words)-1 {
			typ = namevalues.PartTypeSurname
		}
		n.Parts = append(n.Parts, namevalues.Part{Idx: i, Value: w, Type: typ})
	}
	return Value{Name: n}
}
func text(s string) Value    { return Value{Text: s, HasText: true} }
func term(k string) Value    { return Value{Term: k} }
func neutral(k string) Value { return Value{Term: k, Neutral: true} } // flagged neutral in the catalog
func integer(n int64) Value  { return Value{Integer: n, HasInteger: true} }
func ip(n int) *int          { return &n }
func approx(near float64) func(float64) bool {
	return func(got float64) bool { return math.Abs(got-near) < 1e-9 }
}

func TestComparers(t *testing.T) {
	tests := []struct {
		name       string
		cmp        Comparer
		a, b       Value
		want       float64
		comparable bool
	}{

		{"text: same", TextComparer{}, text("York"), text("york"), 1, true},
		{"text: contained", TextComparer{}, text("York, Upper Canada"), text("York"), 0.35, true},
		{"text: lone letters are not initials", TextComparer{}, text("A"), text("Abbey Street"), 0, true},
		{"text: long text pairs greedily", TextComparer{}, text("a1 b2 c3 d4 e5 f6 g7 h8 i9 j10 k11 l12 m13 n14"), text("n14 m13 l12 k11 j10 i9 h8 g7 f6 e5 d4 c3 b2 a1"), 0.7, true},
		{"text: a lone letter matches itself", TextComparer{}, text("Lot A"), text("Lot A"), 1, true},
		{"text: different", TextComparer{}, text("York"), text("Toronto"), 0, true},

		{"term: same", TermComparer{}, term("male"), term("male"), 1, true},
		{"term: different", TermComparer{}, term("male"), term("female"), 0, true},
		{"term: neutral", TermComparer{}, neutral("unknown"), term("male"), 0, false},
		{"term: the key alone is not neutral", TermComparer{}, term("unknown"), term("male"), 0, true},

		{"integer: equal", IntegerComparer{}, integer(40), integer(40), 1, true},
		{"integer: off by one", IntegerComparer{}, integer(40), integer(41), 0, true},
		{"integer: within tolerance", IntegerComparer{Tolerance: Set(int64(1))}, integer(40), integer(41), 0.5, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.cmp.Compare(tt.a, tt.b)
			if ok != tt.comparable || !approx(tt.want)(got) {
				t.Fatalf("got %v %v, want %v %v", got, ok, tt.want, tt.comparable)
			}
			if back, _ := tt.cmp.Compare(tt.b, tt.a); !approx(got)(back) {
				t.Fatalf("not symmetric: %v then %v", got, back)
			}
		})
	}
}

func TestComparerFor(t *testing.T) {
	for _, vt := range []string{"name", "text", "term", "date", "integer"} {
		if ComparerFor(vt) == nil {
			t.Errorf("%s has no comparer", vt)
		}
	}
	if ComparerFor("subject") != nil {
		t.Error("subject values are not compared")
	}
}

func person(ref string, forms []string, sex string) Candidate {
	v := Values{}
	for _, f := range forms {
		v[product("name")] = append(v[product("name")], name(f))
	}
	switch sex {
	case "":
	case "unknown", "indeterminate": // seeded neutral
		v[product("sex_at_birth")] = []Value{neutral(sex)}
	default:
		v[product("sex_at_birth")] = []Value{term(sex)}
	}
	return Candidate{EntityID: []byte(ref), Ref: ref, Values: v}
}

func refs(ms []Match) string {
	var out []string
	for _, m := range ms {
		out = append(out, fmt.Sprintf("%s=%.1f", m.Ref, m.Score))
	}
	return fmt.Sprint(out)
}

func TestRankPersons(t *testing.T) {
	p, ok := DefaultProfile("person")
	if !ok {
		t.Fatal("no person profile")
	}
	probe := person("", []string{"James Robins"}, "male").Values
	candidates := []Candidate{
		person("PER-B", []string{"Jim Robins", "James Robins"}, "male"), // best cluster wins: 10 + 1
		person("PER-A", []string{"James Robins"}, ""),                   // no sex to compare: 10
		person("PER-C", []string{"Mary Robins"}, "unknown"),             // neutral sex: 6
		person("PER-D", []string{"James Robins"}, "female"),             // 10 − 2 = 8, still shown
		person("PER-E", []string{"Ada Lovelace"}, "male"),               // 0 + 1, below
		person("PER-F", nil, "male"),                                    // no name: 1, below
	}
	got := Rank(p, probe, candidates, 0)
	if want := "[PER-B=11.0 PER-A=10.0 PER-D=8.0 PER-C=6.0]"; refs(got) != want {
		t.Fatalf("got %s, want %s", refs(got), want)
	}
	if r := got[0].Reasons; len(r) != 2 || r[0].Property.Key != "name" || r[0].Similarity != 1 || r[1].Contribution != 1 {
		t.Fatalf("reasons %+v", r)
	}
	if r := got[3].Reasons; len(r) != 1 {
		t.Fatalf("a neutral term is not a reason: %+v", r)
	}
	if contra := Score(p, probe, candidates[3]); contra.Reasons[1].Contribution != -2 {
		t.Fatalf("contradiction %+v", contra.Reasons)
	}

	t.Run("limit", func(t *testing.T) {
		if got := Rank(p, probe, candidates, 1); refs(got) != "[PER-B=11.0]" {
			t.Fatalf("%s", refs(got))
		}
	})

	t.Run("ties break by ref", func(t *testing.T) {
		tied := []Candidate{person("PER-Z", []string{"James Robins"}, ""), person("PER-Y", []string{"James Robins"}, "")}
		if got := Rank(p, probe, tied, 0); refs(got) != "[PER-Y=10.0 PER-Z=10.0]" {
			t.Fatalf("%s", refs(got))
		}
	})

	t.Run("a probe with nothing comparable matches nothing", func(t *testing.T) {
		if got := Rank(p, Values{}, candidates, 0); len(got) != 0 {
			t.Fatalf("%s", refs(got))
		}
	})

	t.Run("tuned profile", func(t *testing.T) {
		lenient := p.With(Feature{Property: product("sex_at_birth"), Comparer: TermComparer{}, Weight: 1})
		lenient.MinScore = 1
		if len(lenient.Features) != 3 || len(p.Features) != 3 || p.Features[1].Contradiction != 2 {
			t.Fatal("With must replace in a copy")
		}
		got := Rank(lenient, probe, candidates, 0)
		if want := "[PER-B=11.0 PER-A=10.0 PER-D=10.0 PER-C=6.0 PER-E=1.0 PER-F=1.0]"; refs(got) != want {
			t.Fatalf("got %s, want %s", refs(got), want)
		}
	})
}

func TestRankEventsAndPlaces(t *testing.T) {
	ev, _ := DefaultProfile("event")
	event := func(ref, typ string, d Value) Candidate {
		v := Values{product("event_type"): {term(typ)}}
		if d.Date != nil {
			v[product("date")] = []Value{d}
		}
		return Candidate{EntityID: []byte(ref), Ref: ref, Values: v}
	}
	probe := event("", "birth", date(1817, ip(5), ip(14))).Values
	got := Rank(ev, probe, []Candidate{
		event("EVT-A", "birth", date(1817, ip(5), ip(14))), // 4 + 6
		event("EVT-B", "birth", date(1818, nil, nil)),      // 4 + 6 × 0.498
		event("EVT-C", "birth", Value{}),                   // type alone: 4, below
		event("EVT-D", "death", date(1817, ip(5), ip(14))), // −6 + 6 = 0
		event("EVT-E", "birth", date(1900, nil, nil)),      // 4 − 4 = 0
	}, 0)
	if want := "[EVT-A=10.0 EVT-B=7.0]"; refs(got) != want {
		t.Fatalf("events: got %s, want %s", refs(got), want)
	}

	pl, _ := DefaultProfile("place")
	place := func(ref, toponym string) Candidate {
		return Candidate{EntityID: []byte(ref), Ref: ref, Values: Values{product("toponym"): {text(toponym)}}}
	}
	got = Rank(pl, Values{product("toponym"): {text("York")}}, []Candidate{
		place("PLC-A", "York, Upper Canada"), place("PLC-B", "york"), place("PLC-C", "Toronto"),
	}, 0)
	if want := "[PLC-B=10.0 PLC-A=3.5]"; refs(got) != want {
		t.Fatalf("places: got %s, want %s", refs(got), want)
	}

	if _, ok := DefaultProfile("participation"); ok {
		t.Fatal("bridge kinds have no profile")
	}
}

// Pairwise evaluation treats a term the catalog flags neutral as nothing to
// compare, as Rank's person profile and the reconciler do: "unknown" neither
// conflicts with "male" nor agrees with another "unknown".
func TestEvaluateNeutralTermIsNotComparable(t *testing.T) {
	sex := Property{Key: "sex_at_birth", Origin: "provenencia"}
	metas := []PropertyMeta{{Property: sex, ValueType: "term"}}
	for _, other := range []Value{term("male"), neutral("unknown"), neutral("indeterminate")} {
		ev := Evaluate(Values{sex: {neutral("unknown")}}, Values{sex: {other}}, metas)
		if c := ev.Comparisons[0]; c.Comparable || c.Outcome != OutcomeUnknown {
			t.Fatalf("unknown vs %s: comparable=%v outcome=%s, want not comparable", other.Term, c.Comparable, c.Outcome)
		}
	}
	ev := Evaluate(Values{sex: {{Term: "female"}}}, Values{sex: {{Term: "male"}}}, metas)
	if c := ev.Comparisons[0]; !c.Comparable || c.Outcome != OutcomeConflict {
		t.Fatalf("female vs male: comparable=%v outcome=%s, want a conflict", c.Comparable, c.Outcome)
	}
}

func TestDiscountedScoring(t *testing.T) {
	prof, _ := DefaultProfile("person")
	names := NameComparer{Frequency: counts100()}
	prof = prof.WithNameComparer(names)
	probe := func(f string) Values { return Values{product("name"): {name(f)}} }
	cand := func(f string) Candidate { return Candidate{Ref: f, Values: probe(f)} }

	for _, tt := range []struct {
		form     string
		score    float64
		discount float64
	}{{"John Quirk", 9.60, 0.040}, {"John Smith", 8.64, 0.136}} {
		t.Run(tt.form, func(t *testing.T) {
			got := Rank(prof, probe(tt.form), []Candidate{cand(tt.form)}, 0)
			if len(got) != 1 {
				t.Fatalf("got %d matches", len(got))
			}
			r := got[0].Reasons[0]
			if math.Abs(got[0].Score-tt.score) > 0.005 || math.Abs(r.Discount-tt.discount) > 0.0005 || r.Similarity != 1 {
				t.Fatalf("score %.3f discount %.4f similarity %v", got[0].Score, r.Discount, r.Similarity)
			}
		})
	}

	t.Run("evaluate still agrees and reports the discount", func(t *testing.T) {
		metas := []PropertyMeta{{Property: product("name"), ValueType: "name"}}
		ev := EvaluateWithNames(probe("John Smith"), probe("John Smith"), metas, &names)
		pc := ev.Comparisons[0]
		if pc.Outcome != OutcomeAgree || math.Abs(pc.Discount-0.136) > 0.0005 {
			t.Fatalf("%+v", pc)
		}
	})

	t.Run("a contradiction is not discounted", func(t *testing.T) {
		p := Profile{Features: []Feature{{Property: product("name"), Comparer: names, Weight: 10, Contradiction: 2}}}
		m := Score(p, probe("John Smith"), cand("Mary Jones"))
		if r := m.Reasons[0]; r.Contribution != -2 || r.Discount != 0 {
			t.Fatalf("%+v", r)
		}
	})

	t.Run("the cut sees the discounted score", func(t *testing.T) {
		p := Profile{MinScore: 3, Features: []Feature{{Property: product("name"), Comparer: names, Weight: 3.2}}}
		m := Score(p, probe("John Smith"), cand("John Smith"))
		if math.Abs(m.Score-2.765) > 0.005 {
			t.Fatalf("score %.3f", m.Score)
		}
		if got := Rank(p, probe("John Smith"), []Candidate{cand("John Smith")}, 0); len(got) != 0 {
			t.Fatalf("returned %d", len(got))
		}
	})
}

// cachedPerson is a handle whose values carry cache support, in rank order.
func cachedPerson(ref string, names map[string]int, sex map[string]int) Candidate {
	v := Values{}
	for _, f := range sortedKeys(names) {
		n := name(f)
		n.SetEvidence(names[f], 0, len(v[product("name")])+1, "kept")
		v[product("name")] = append(v[product("name")], n)
	}
	for _, k := range sortedKeys(sex) {
		t := term(k)
		t.SetEvidence(sex[k], 0, len(v[product("sex_at_birth")])+1, "kept")
		v[product("sex_at_birth")] = append(v[product("sex_at_birth")], t)
	}
	return Candidate{EntityID: []byte(ref), Ref: ref, Values: v}
}

func sortedKeys(m map[string]int) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func reasonFor(t *testing.T, m Match, key string) Reason {
	t.Helper()
	for _, r := range m.Reasons {
		if r.Property.Key == key {
			return r
		}
	}
	t.Fatalf("no %s reason in %+v", key, m.Reasons)
	return Reason{}
}

func TestScoreSupport(t *testing.T) {
	p, _ := DefaultProfile("person")
	sexProbe := func(k string) Values { return Values{product("sex_at_birth"): {term(k)}} }
	handle := cachedPerson("PER-A", nil, map[string]int{"male": 3, "female": 1})

	if r := reasonFor(t, Score(p, sexProbe("female"), handle), "sex_at_birth"); r.Contribution != -2 || r.Probe.Label != "female" || r.Candidate.Label != "male" {
		t.Errorf("AC-1.1: %+v", r)
	}
	if r := reasonFor(t, Score(p, sexProbe("male"), handle), "sex_at_birth"); r.Contribution != 1 || r.Probe.Label != "male" || r.Candidate.Label != "male" {
		t.Errorf("AC-1.2: %+v", r)
	}

	stray := cachedPerson("PER-B", map[string]int{"Mary Smith": 5, "James Robins": 1}, nil)
	probe := Values{product("name"): {name("James Robins")}}
	r := reasonFor(t, Score(p, probe, stray), "name")
	if r.Similarity != 0 || r.Contribution != 0 || r.Probe.Label != "James Robins" || r.Candidate.Label != "Mary Smith" {
		t.Errorf("AC-1.4: %+v", r)
	}

	equal := cachedPerson("PER-C", map[string]int{"Mary Smith": 1, "Mary Jones": 1}, nil)
	probe = Values{product("name"): {name("Mary Jones")}}
	if r := reasonFor(t, Score(p, probe, equal), "name"); r.Contribution != 10 {
		t.Errorf("AC-1.5: %+v", r)
	}

	near := cachedPerson("PER-D", map[string]int{"Robins": 3, "Robbins": 1}, nil)
	probe = Values{product("name"): {name("Robbins")}}
	r = reasonFor(t, Score(p, probe, near), "name")
	if !approx(0.857)(math.Round(r.Similarity*1000)/1000) || r.Probe.Label != "Robbins" || r.Candidate.Label != "Robins" {
		t.Errorf("AC-1.6: %+v", r)
	}
}

// A multiple-cardinality Feature keeps every displayed value as a leader, so
// a handle's minority toponym still matches in full.
func TestScoreMultipleToponyms(t *testing.T) {
	p, _ := DefaultProfile("place")
	york, eboracum := text("York"), text("Eboracum")
	york.SetEvidence(3, 0, 1, "kept")
	eboracum.SetEvidence(1, 0, 2, "kept")
	cand := Candidate{EntityID: []byte("PLC-A"), Ref: "PLC-A", Values: Values{product("toponym"): {york, eboracum}}}
	probe := Values{product("toponym"): {text("Eboracum")}}

	if r := reasonFor(t, Score(p, probe, cand), "toponym"); r.Contribution != 10 || r.Candidate.Label != "Eboracum" {
		t.Errorf("minority toponym: %+v", r)
	}
	if got := Rank(p, probe, []Candidate{cand}, 10); len(got) != 1 || got[0].Ref != "PLC-A" {
		t.Errorf("Rank = %+v", got)
	}
}

func TestScoreDatePeriod(t *testing.T) {
	ev, _ := DefaultProfile("event")
	period := Candidate{Ref: "EVT-P", Values: Values{pStart: {year(1815)}, pEnd: {year(1820)}}}
	dated := func(y int) Candidate { return Candidate{Ref: "EVT-D", Values: Values{pDate: {year(y)}}} }
	reasonFor := func(m Match, p Property) (Reason, bool) {
		for _, r := range m.Reasons {
			if r.Property == p {
				return r, true
			}
		}
		return Reason{}, false
	}

	t.Run("a date inside a period", func(t *testing.T) {
		m := Score(ev, Values{pDate: {year(1817)}}, period)
		r, ok := reasonFor(m, pDate)
		if !ok || !approx(0.55)(r.Similarity) || !approx(3.3)(r.Contribution) || len(m.Reasons) != 1 {
			t.Fatalf("%+v", m.Reasons)
		}
		if r.Probe.Label != "1817" || r.Candidate.Label != "Between 1815 and 1820" || r.Candidate.Date == nil {
			t.Fatalf("labels %+v", r)
		}
	})
	t.Run("swapped", func(t *testing.T) {
		m := Score(ev, period.Values, dated(1817))
		r, ok := reasonFor(m, pDate)
		if !ok || !approx(3.3)(r.Contribution) || r.Probe.Label != "Between 1815 and 1820" || r.Candidate.Label != "1817" {
			t.Fatalf("%+v", m.Reasons)
		}
	})
	t.Run("a date beyond the tolerance", func(t *testing.T) {
		m := Score(ev, Values{pDate: {year(1823)}}, period)
		if r, ok := reasonFor(m, pDate); !ok || r.Contribution != -4 {
			t.Fatalf("%+v", m.Reasons)
		}
	})
	t.Run("only a start", func(t *testing.T) {
		m := Score(ev, Values{pDate: {year(1817)}}, Candidate{Values: Values{pStart: {year(1815)}}})
		r, ok := reasonFor(m, pDate)
		if !ok || !approx(0.2)(r.Similarity) || r.Candidate.Label != "After 1815" {
			t.Fatalf("%+v", m.Reasons)
		}
	})
	t.Run("both with a date", func(t *testing.T) {
		m := Score(ev, Values{pDate: {year(1817)}}, dated(1817))
		if r, ok := reasonFor(m, pDate); !ok || r.Similarity != 1 || r.Candidate.Date.Kind == datevalue.KindRange {
			t.Fatalf("%+v", m.Reasons)
		}
	})
	t.Run("start and end only", func(t *testing.T) {
		m := Score(ev, period.Values, period)
		if _, ok := reasonFor(m, pDate); ok || len(m.Reasons) != 2 {
			t.Fatalf("%+v", m.Reasons)
		}
		if m.Reasons[0].Property != pStart || m.Reasons[1].Property != pEnd {
			t.Fatalf("%+v", m.Reasons)
		}
	})
	t.Run("a date, a start, and an end against a start and an end", func(t *testing.T) {
		probe := Values{pDate: {year(1817)}, pStart: {year(1815)}, pEnd: {year(1820)}}
		m := Score(ev, probe, period)
		if _, ok := reasonFor(m, pDate); ok || len(m.Reasons) != 2 {
			t.Fatalf("%+v", m.Reasons)
		}
	})
}
