package match

import (
	"fmt"
	"math"
	"math/rand"
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

func span(lo, hi *int) Value {
	return Value{Date: &datevalues.Value{Kind: datevalues.KindRange, StartYear: lo, EndYear: hi}}
}

func bound(q string, y int) Value {
	return Value{Date: &datevalues.Value{Kind: datevalues.KindPoint, Qualifier: q, StartYear: &y}}
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
		{"name: same words reordered", NameComparer{}, name("Robins, James"), name("James Robins"), 1, true},
		{"name: shared surname", NameComparer{}, name("Mary Robins"), name("James Robins"), 0.5, true},
		{"name: initial is half a word", NameComparer{}, name("J. Robins"), name("James Robins"), 2 * 1.5 / 4, true},
		{"name: spelling variant", NameComparer{}, name("James Robbins"), name("James Robins"), 2 * (1 + 6.0/7) / 4, true},
		{"name: not a variant", NameComparer{}, name("Mary"), name("Mark"), 0, true},
		{"name: fuzzy off", NameComparer{FuzzyFloor: Set(1.0)}, name("Robbins"), name("Robins"), 0, true},
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

		{"date: same year, both year-only", DateComparer{}, date(1817, nil, nil), date(1817, nil, nil), 1, true},
		{"date: same month, both without day", DateComparer{}, date(1817, ip(5), nil), date(1817, ip(5), nil), 1, true},
		{"date: same day", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), ip(14)), 1, true},
		{"date: same month", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), nil), 0.9, true},
		{"date: same year, other month", DateComparer{}, date(1817, ip(5), nil), date(1817, ip(9), nil), 0.6, true},
		{"date: same year", DateComparer{}, date(1817, nil, nil), date(1817, ip(9), nil), 0.8, true},
		{"date: a year apart", DateComparer{}, date(1817, nil, nil), date(1818, nil, nil), 0.8 * 2 / 3, true},
		{"date: beyond tolerance", DateComparer{}, date(1817, nil, nil), date(1820, nil, nil), 0, true},
		{"date: wider tolerance", DateComparer{Tolerance: Set(5)}, date(1817, nil, nil), date(1820, nil, nil), 0.8 * 3 / 6, true},
		{"date: about widens", DateComparer{}, Value{Date: &datevalues.Value{StartYear: ip(1817), Qualifier: "ABT"}}, date(1820, nil, nil), 0.8 * 2 / 5, true},
		{"date: same month, different days", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), ip(20)), 0.7, true},
		{"date: range contains the point", DateComparer{}, span(ip(1815), ip(1820)), date(1818, nil, nil), 0.8 - 0.05*5, true},
		{"date: tight range contains the point", DateComparer{}, span(ip(1817), ip(1818)), date(1817, nil, nil), 0.8 - 0.05, true},
		{"date: wide range contains the point", DateComparer{}, span(ip(1700), ip(1900)), date(1817, nil, nil), 0.2, true},
		{"date: point just outside a range", DateComparer{}, span(ip(1815), ip(1820)), date(1822, nil, nil), (0.8 - 0.05*5) * (1 - 2.0/3), true},
		{"date: point far outside a range", DateComparer{}, span(ip(1815), ip(1820)), date(1830, nil, nil), 0, true},
		{"date: identical ranges", DateComparer{}, span(ip(1815), ip(1820)), span(ip(1815), ip(1820)), 1, true},
		{"date: overlapping ranges, judged by the wider", DateComparer{}, span(ip(1815), ip(1820)), span(ip(1819), ip(1825)), 0.8 - 0.05*6, true},
		{"date: open-ended range", DateComparer{}, span(ip(1815), nil), date(1900, nil, nil), 0.2, true},
		{"date: identical bounds", DateComparer{}, bound("BEF", 1820), bound("BEF", 1820), 1, true},
		{"date: before, consistent", DateComparer{}, bound("BEF", 1820), date(1815, nil, nil), 0.2, true},
		{"date: before, impossible", DateComparer{}, bound("BEF", 1820), date(1823, nil, nil), 0, true},
		{"date: after, a year short", DateComparer{}, bound("AFT", 1820), date(1819, nil, nil), 0.2 * (1 - 1.0/3), true},
		{"date: before vs after, disjoint", DateComparer{}, bound("BEF", 1800), bound("AFT", 1810), 0, true},
		{"date: before vs before, different years", DateComparer{}, bound("BEF", 1800), bound("BEF", 1810), 0.2, true},
		{"date: tolerance zero", DateComparer{Tolerance: Set(0)}, date(1817, nil, nil), date(1818, nil, nil), 0, true},
		{"date: inverted range reads as its years", DateComparer{}, span(ip(1820), ip(1815)), span(ip(1815), ip(1820)), 1, true},
		{"date: empty range", DateComparer{}, span(nil, nil), date(1818, nil, nil), 0, false},
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
		person("PER-C", []string{"Mary Robins"}, "unknown"),             // neutral sex: 5
		person("PER-D", []string{"James Robins"}, "female"),             // 10 − 8 = 2, below MinScore
		person("PER-E", []string{"Ada Lovelace"}, "male"),               // 0 + 1, below
		person("PER-F", nil, "male"),                                    // no name: 1, below
	}
	got := Rank(p, probe, candidates, 0)
	if want := "[PER-B=11.0 PER-A=10.0 PER-C=5.0]"; refs(got) != want {
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
		if want := "[PER-B=11.0 PER-A=10.0 PER-D=10.0 PER-C=5.0 PER-E=1.0 PER-F=1.0]"; refs(got) != want {
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

// Date ladder: each date must resemble 14 May 1817 strictly more than the next.
func TestDateComparerLadder(t *testing.T) {
	probe := date(1817, ip(5), ip(14))
	order := []struct {
		name string
		v    Value
	}{
		{"same day", date(1817, ip(5), ip(14))},
		{"same month, no day", date(1817, ip(5), nil)},
		{"same year, no month", date(1817, nil, nil)},
		{"same month, other day", date(1817, ip(5), ip(20))},
		{"same year, other month", date(1817, ip(9), nil)},
		{"inside a six-year range", span(ip(1815), ip(1820))},
		{"a year later", date(1818, nil, nil)},
		{"two years later", date(1819, nil, nil)},
		{"before a later year", bound("BEF", 1820)},
		{"three years later", date(1820, nil, nil)},
	}
	prev := math.Inf(1)
	for i, o := range order {
		got, ok := DateComparer{}.Compare(probe, o.v)
		if !ok {
			t.Fatalf("%s: not comparable", o.name)
		}
		if !(got < prev) && !(i == len(order)-1 && got == 0) {
			t.Errorf("#%d %s scored %.4f, not below %.4f", i, o.name, got, prev)
		}
		prev = got
	}
}

// Generated dates from fixed seeds: symmetry, range, and self-similarity.
func TestDateComparerInvariants(t *testing.T) {
	quals := []string{"", "", "ABT", "BEF", "AFT"}
	opt := func(rng *rand.Rand, lo, n int) *int {
		if rng.Intn(3) == 0 {
			return nil
		}
		return ip(lo + rng.Intn(n))
	}
	for _, seed := range []int64{1, 2, 3, 4} {
		rng := rand.New(rand.NewSource(seed))
		gen := func() Value {
			if rng.Intn(3) == 0 {
				return span(opt(rng, 1810, 10), opt(rng, 1815, 10))
			}
			d := &datevalues.Value{Kind: datevalues.KindPoint, Qualifier: quals[rng.Intn(len(quals))], StartYear: opt(rng, 1810, 15)}
			if d.StartYear != nil {
				d.StartMonth = opt(rng, 1, 12)
				if d.StartMonth != nil {
					d.StartDay = opt(rng, 1, 28)
				}
			}
			return Value{Date: d}
		}
		for _, cmp := range []DateComparer{{}, {Tolerance: Set(0)}, {Tolerance: Set(5)}} {
			for i := 0; i < 2000; i++ {
				a, b := gen(), gen()
				got, ok := cmp.Compare(a, b)
				back, okBack := cmp.Compare(b, a)
				where := fmt.Sprintf("seed %d pair %d: %+v vs %+v", seed, i, *a.Date, *b.Date)
				if ok != okBack || math.Abs(got-back) > 1e-9 {
					t.Fatalf("%s: not symmetric: %.4f then %.4f", where, got, back)
				}
				if ok && (got < 0 || got > 1) {
					t.Fatalf("%s: out of range %.4f", where, got)
				}
				if self, okSelf := cmp.Compare(a, a); okSelf && self != 1 {
					t.Fatalf("%s: self-similarity %.4f", where, self)
				}
			}
		}
	}
}
