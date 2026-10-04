package match

import (
	"fmt"
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
)

func name(form string) Value { return Value{Name: &namevalues.Value{Form: form}} }
func text(s string) Value    { return Value{Text: s, HasText: true} }
func term(k string) Value    { return Value{Term: k} }
func integer(n int64) Value  { return Value{Integer: n, HasInteger: true} }
func ip(n int) *int          { return &n }
func approx(near float64) func(float64) bool {
	return func(got float64) bool { return math.Abs(got-near) < 1e-9 }
}

func date(y int, m, d *int) Value {
	return Value{Date: &datevalues.Value{StartYear: &y, StartMonth: m, StartDay: d}}
}

func TestComparers(t *testing.T) {
	tests := []struct {
		name       string
		cmp        Comparer
		a, b       Value
		want       float64
		comparable bool
	}{
		{"name: same normalized form", NameComparer{}, name("James Robins"), name("james  robins."), 1, true},
		{"name: same words reordered", NameComparer{}, name("Robins, James"), name("James Robins"), 0.8, true},
		{"name: shared surname", NameComparer{}, name("Mary Robins"), name("James Robins"), 0.4, true},
		{"name: initial is half a word", NameComparer{}, name("J. Robins"), name("James Robins"), 0.8 * 2 * 1.5 / 4, true},
		{"name: spelling variant", NameComparer{}, name("James Robbins"), name("James Robins"), 0.8 * 2 * (1 + 6.0/7) / 4, true},
		{"name: not a variant", NameComparer{}, name("Mary"), name("Mark"), 0, true},
		{"name: fuzzy off", NameComparer{FuzzyFloor: 1}, name("Robbins"), name("Robins"), 0, true},
		{"name: custom partial", NameComparer{Partial: 0.5}, name("Mary Robins"), name("James Robins"), 0.25, true},
		{"name: missing", NameComparer{}, name("James"), Value{}, 0, false},

		{"text: same", TextComparer{}, text("York"), text("york"), 1, true},
		{"text: contained", TextComparer{}, text("York, Upper Canada"), text("York"), 0.35, true},
		{"text: lone letters are not initials", TextComparer{}, text("A"), text("Abbey Street"), 0, true},
		{"text: long text pairs greedily", TextComparer{}, text("a1 b2 c3 d4 e5 f6 g7 h8 i9 j10 k11 l12 m13 n14"), text("n14 m13 l12 k11 j10 i9 h8 g7 f6 e5 d4 c3 b2 a1"), 0.7, true},
		{"text: a lone letter matches itself", TextComparer{}, text("Lot A"), text("Lot A"), 1, true},
		{"text: different", TextComparer{}, text("York"), text("Toronto"), 0, true},

		{"term: same", TermComparer{}, term("male"), term("male"), 1, true},
		{"term: different", TermComparer{}, term("male"), term("female"), 0, true},
		{"term: neutral", TermComparer{Neutral: map[string]bool{"unknown": true}}, term("unknown"), term("male"), 0, false},

		{"date: same day", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), ip(14)), 1, true},
		{"date: same month", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), nil), 0.9, true},
		{"date: same year, other month", DateComparer{}, date(1817, ip(5), nil), date(1817, ip(9), nil), 0.6, true},
		{"date: same year", DateComparer{}, date(1817, nil, nil), date(1817, ip(9), nil), 0.8, true},
		{"date: a year apart", DateComparer{}, date(1817, nil, nil), date(1818, nil, nil), 0.8 * 2 / 3, true},
		{"date: beyond tolerance", DateComparer{}, date(1817, nil, nil), date(1820, nil, nil), 0, true},
		{"date: wider tolerance", DateComparer{Tolerance: 5}, date(1817, nil, nil), date(1820, nil, nil), 0.8 * 3 / 6, true},
		{"date: about widens", DateComparer{}, Value{Date: &datevalues.Value{StartYear: ip(1817), Qualifier: "ABT"}}, date(1820, nil, nil), 0.8 * 2 / 5, true},
		{"date: phrase only", DateComparer{}, Value{Date: &datevalues.Value{Phrase: "spring"}}, date(1817, nil, nil), 0, false},

		{"integer: equal", IntegerComparer{}, integer(40), integer(40), 1, true},
		{"integer: off by one", IntegerComparer{}, integer(40), integer(41), 0, true},
		{"integer: within tolerance", IntegerComparer{Tolerance: 1}, integer(40), integer(41), 0.5, true},
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
	if sex != "" {
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
		person("PER-C", []string{"Mary Robins"}, "unknown"),             // neutral sex: 4
		person("PER-D", []string{"James Robins"}, "female"),             // 10 − 8 = 2, below MinScore
		person("PER-E", []string{"Ada Lovelace"}, "male"),               // 0 + 1, below
		person("PER-F", nil, "male"),                                    // no name: 1, below
	}
	got := Rank(p, probe, candidates, 0)
	if want := "[PER-B=11.0 PER-A=10.0 PER-C=4.0]"; refs(got) != want {
		t.Fatalf("got %s, want %s", refs(got), want)
	}
	if r := got[0].Reasons; len(r) != 2 || r[0].Property.Key != "name" || r[0].Similarity != 1 || r[1].Contribution != 1 {
		t.Fatalf("reasons %+v", r)
	}
	if r := got[2].Reasons; len(r) != 1 {
		t.Fatalf("a neutral term is not a reason: %+v", r)
	}
	if contra := Score(p, probe, candidates[3]); contra.Reasons[1].Contribution != -8 {
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
		if len(lenient.Features) != 2 || len(p.Features) != 2 || p.Features[1].Contradiction != 8 {
			t.Fatal("With must replace in a copy")
		}
		got := Rank(lenient, probe, candidates, 0)
		if want := "[PER-B=11.0 PER-A=10.0 PER-D=10.0 PER-C=4.0 PER-E=1.0 PER-F=1.0]"; refs(got) != want {
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
		event("EVT-B", "birth", date(1818, nil, nil)),      // 4 + 6 × 0.533
		event("EVT-C", "birth", Value{}),                   // type alone: 4, below
		event("EVT-D", "death", date(1817, ip(5), ip(14))), // −6 + 6 = 0
		event("EVT-E", "birth", date(1900, nil, nil)),      // 4 − 4 = 0
	}, 0)
	if want := "[EVT-A=10.0 EVT-B=7.2]"; refs(got) != want {
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
