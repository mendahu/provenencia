package match

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

// cached is a handle value as the cache loader fills it.
func cached(v Value, support, against, rank int, reason string) Value {
	v.SetEvidence(support, against, rank, reason)
	return v
}

func kept(v Value, support int) Value { return cached(v, support, 0, 1, "kept") }

func negative(v Value, id byte) Value {
	v.Negative = true
	v.ObservationID = []byte{id}
	return v
}

func TestPairValues(t *testing.T) {
	male, female := term("male"), term("female")
	tests := []struct {
		name         string
		cmp          Comparer
		probe, cand  []Value
		sim          float64
		probeLabel   string
		candLabel    string
		notComparble bool
	}{
		{"AC-1.1 one dissenting reading", TermComparer{}, []Value{female}, []Value{kept(male, 3), kept(female, 1)}, 0, "female", "male", false},
		{"AC-1.2 probe is the majority", TermComparer{}, []Value{male}, []Value{kept(male, 3), kept(female, 1)}, 1, "male", "male", false},
		{"AC-1.3 equal support keeps both", TermComparer{}, []Value{female}, []Value{kept(male, 2), kept(female, 2)}, 1, "female", "female", false},
		{"AC-1.8 displayed value leads, male", TermComparer{}, []Value{male}, []Value{kept(male, 1), cached(female, 3, 0, 2, "weak")}, 1, "male", "male", false},
		{"AC-1.8 displayed value leads, female", TermComparer{}, []Value{female}, []Value{kept(male, 1), cached(female, 3, 0, 2, "weak")}, 0, "female", "male", false},
		{"AC-1.9 neutral never leads", TermComparer{}, []Value{female}, []Value{cached(neutral("unknown"), 5, 0, 1, "kept"), cached(male, 1, 0, 2, "outvoted")}, 0, "female", "male", false},
		{"AC-1.10 sole displayed value with against, male", TermComparer{}, []Value{male}, []Value{cached(male, 1, 3, 1, "kept")}, 1, "male", "male", false},
		{"AC-1.10 sole displayed value with against, female", TermComparer{}, []Value{female}, []Value{cached(male, 1, 3, 1, "kept")}, 0, "female", "male", false},
		{"AC-1.10 concluded value support 0, male", TermComparer{}, []Value{male}, []Value{cached(male, 0, 0, 1, "kept")}, 1, "male", "male", false},
		{"AC-1.10 concluded value support 0, female", TermComparer{}, []Value{female}, []Value{cached(male, 0, 0, 1, "kept")}, 0, "female", "male", false},
		{"AC-1.7 no cache evidence", TermComparer{}, []Value{female}, []Value{male, female}, 1, "female", "female", false},
		{"AC-1.7 no cache evidence, none shared", TermComparer{}, []Value{term("x")}, []Value{male, female}, 0, "x", "male", false},
		{"AC-1.6 minority exact match is weaker than the leader's near match", DefaultNames, []Value{name("Robbins")}, []Value{kept(name("Robins"), 3), kept(name("Robbins"), 1)}, 0.857, "Robbins", "Robins", false},
		{"AC-1.5 equal support keeps both names", DefaultNames, []Value{name("Mary Jones")}, []Value{kept(name("Mary Smith"), 1), kept(name("Mary Jones"), 1)}, 1, "Mary Jones", "Mary Jones", false},
		{"AC-1.4 stray name", DefaultNames, []Value{name("James Robins")}, []Value{kept(name("Mary Smith"), 5), kept(name("James Robins"), 1)}, 0, "James Robins", "Mary Smith", false},
		{"empty probe", TermComparer{}, nil, []Value{kept(male, 1)}, 0, "", "", true},
		{"only neutral", TermComparer{}, []Value{neutral("unknown")}, []Value{kept(male, 1)}, 0, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pairValues(tt.cmp, "", tt.probe, tt.cand)
			if got.Comparable == tt.notComparble {
				t.Fatalf("comparable = %v, want %v", got.Comparable, !tt.notComparble)
			}
			if tt.notComparble {
				return
			}
			if math.Abs(got.Similarity-tt.sim) > 0.001 {
				t.Errorf("similarity = %v, want %v", got.Similarity, tt.sim)
			}
			if got.Probe.Label != tt.probeLabel || got.Candidate.Label != tt.candLabel {
				t.Errorf("pair = (%q, %q), want (%q, %q)", got.Probe.Label, got.Candidate.Label, tt.probeLabel, tt.candLabel)
			}
		})
	}
}

// AC-2.5: effective supports are 1 and 2, so female is the consensus.
func TestPairValuesAgainstReducesSupport(t *testing.T) {
	male, female := term("male"), term("female")
	cand := []Value{cached(male, 3, 2, 1, "kept"), cached(female, 2, 0, 2, "kept")}
	if got := pairValues(TermComparer{}, "", []Value{female}, cand); got.Similarity != 1 {
		t.Errorf("female: %+v", got)
	}
	if got := pairValues(TermComparer{}, "", []Value{male}, cand); got.Similarity != 0 || got.Candidate.Label != "female" {
		t.Errorf("male: %+v", got)
	}
}

// AC-3.3: the same pair is reported whatever order the rows arrive in.
func TestPairValuesOrderIndependent(t *testing.T) {
	a := cached(name("Anna Smith"), 2, 0, 1, "kept")
	b := cached(name("Bea Smith"), 2, 0, 2, "kept")
	c := cached(name("Cora Smith"), 2, 0, 3, "kept")
	probe := []Value{name("Zed Jones")}
	want := pairValues(DefaultNames, "", probe, []Value{a, b, c})
	for _, cand := range [][]Value{{c, b, a}, {b, a, c}, {c, a, b}} {
		got := pairValues(DefaultNames, "", probe, cand)
		if got.Candidate.Label != want.Candidate.Label || got.Probe.Label != want.Probe.Label {
			t.Errorf("order changed the pair: %+v vs %+v", got, want)
		}
	}
	if want.Candidate.Label != "Anna Smith" {
		t.Errorf("equal support breaks the tie by label: %q", want.Candidate.Label)
	}
}

func TestPairValuesNegatives(t *testing.T) {
	male, female := term("male"), term("female")
	handle := []Value{kept(male, 3), kept(female, 1)}
	t.Run("AC-2.1 negative against the leader", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{negative(male, 7)}, handle)
		if !got.Comparable || got.Similarity != 0 {
			t.Fatalf("%+v", got)
		}
		if !got.Probe.Negative || got.Probe.Label != "male" || got.Candidate.Label != "male" || string(got.Probe.ObservationID) != "\x07" {
			t.Errorf("refs: %+v", got)
		}
	})
	t.Run("negative against a minority value is not comparable", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{negative(female, 7)}, handle)
		if got.Comparable {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("AC-2.2 negative alone against a female-led handle", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{negative(male, 7)}, []Value{kept(female, 3), kept(male, 1)})
		if got.Comparable {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("AC-2.3 a positive does not rescue the negative", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{female, negative(male, 7)}, handle)
		if !got.Comparable || got.Similarity != 0 {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("AC-2.4 a denied value", func(t *testing.T) {
		denied := []Value{cached(male, 3, 0, 1, "denied"), kept(female, 1)}
		got := pairValues(TermComparer{}, "", []Value{male}, denied)
		if !got.Comparable || got.Similarity != 0 || got.Probe.Label != "male" || got.Candidate.Label != "male" {
			t.Fatalf("%+v", got)
		}
		if got := pairValues(TermComparer{}, "", []Value{female}, denied); got.Similarity != 1 {
			t.Errorf("female agrees with the standing value: %+v", got)
		}
	})
	t.Run("a denied value backed by a standing one is no contradiction", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{male}, []Value{cached(male, 3, 0, 1, "denied"), kept(male, 1)})
		if got.Similarity != 1 {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("a weak value with against at its support is silent", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{female}, []Value{kept(male, 2), cached(female, 1, 1, 2, "weak")})
		if got.Similarity != 0 || got.Candidate.Label != "male" {
			t.Fatalf("%+v", got)
		}
		got = pairValues(TermComparer{}, "", []Value{female}, []Value{cached(female, 1, 1, 1, "weak")})
		if got.Comparable {
			t.Fatalf("a weight-0 value neither credits nor contradicts: %+v", got)
		}
	})
	t.Run("a negative never raises similarity", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{female, negative(female, 7)}, handle)
		if got.Similarity != 0 {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("compatible folds count", func(t *testing.T) {
		got := pairValues(DefaultNames, "", []Value{negative(name("J. Robins"), 9)}, []Value{kept(name("James Robins"), 3)})
		if !got.Comparable || got.Similarity != 0 {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("a candidate negative against a probe leader", func(t *testing.T) {
		got := pairValues(TermComparer{}, "", []Value{male}, []Value{negative(male, 3), kept(female, 1)})
		if got.Similarity != 0 || !got.Candidate.Negative {
			t.Fatalf("%+v", got)
		}
	})
}

func TestValueLabel(t *testing.T) {
	year := 1890
	d := datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year}
	tests := []struct {
		v    Value
		want string
	}{
		{text("York"), "York"},
		{term("male"), "male"},
		{Value{Term: "two-spirit", TermLabel: "Two-Spirit"}, "Two-Spirit"},
		{integer(42), "42"},
		{Value{Date: &d}, "1890"},
		{name("James Robins"), "James Robins"},
		{Value{Name: &namevalues.Value{Parts: []namevalues.Part{{Value: "James"}, {Value: "Robins"}}}}, "James Robins"},
	}
	for _, tt := range tests {
		if got := tt.v.Label(); got != tt.want {
			t.Errorf("Label() = %q, want %q", got, tt.want)
		}
	}
}

func TestSetEvidence(t *testing.T) {
	var v Value
	v.SetEvidence(3, 1, 2, "denied")
	if !v.Counted || !v.Denied || v.Support != 3 || v.Against != 1 || v.Rank != 2 || v.Reason != "denied" {
		t.Errorf("%+v", v)
	}
	v.SetEvidence(1, 0, 1, "kept")
	if !v.Counted || v.Denied {
		t.Errorf("only denied sets Denied: %+v", v)
	}
}

func TestValuesHasPositive(t *testing.T) {
	sex := product("sex_at_birth")
	if (Values{sex: {negative(term("male"), 1)}}).HasPositive() {
		t.Error("a negative alone is not positive")
	}
	if !(Values{sex: {negative(term("male"), 1), term("female")}}).HasPositive() {
		t.Error("one positive is enough")
	}
	if (Values{}).HasPositive() {
		t.Error("empty")
	}
}

func BenchmarkPairValues(b *testing.B) {
	plainA := []Value{term("male"), term("female")}
	plainB := []Value{term("female"), term("other")}
	b.Run("plain", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pairValues(TermComparer{}, "", plainA, plainB)
		}
	})
	counted := []Value{kept(term("male"), 3), kept(term("female"), 1)}
	b.Run("counted", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pairValues(TermComparer{}, "", plainA, counted)
		}
	})
}

func TestPairValuesMultiple(t *testing.T) {
	york, eboracum := text("York"), text("Eboracum")
	handle := []Value{kept(york, 3), kept(eboracum, 1)}

	t.Run("minority value speaks at full weight", func(t *testing.T) {
		got := pairValues(DefaultText, "multiple", []Value{text("Eboracum")}, handle)
		if got.Similarity != 1 || got.Candidate.Label != "Eboracum" {
			t.Errorf("got %+v", got)
		}
	})
	t.Run("no shared value is not a leader contradiction", func(t *testing.T) {
		got := pairValues(DefaultText, "multiple", []Value{text("London")}, handle)
		if got.Comparable && got.Similarity != 0 {
			t.Errorf("got %+v", got)
		}
	})
	t.Run("a compatible negative is still a contradiction (R1)", func(t *testing.T) {
		got := pairValues(DefaultText, "multiple", []Value{negative(text("Eboracum"), 7)}, handle)
		if !got.Comparable || got.Similarity != 0 || got.Probe.Label != "Eboracum" || !got.Probe.Negative {
			t.Errorf("got %+v", got)
		}
	})
}

func TestOrderIgnoresTermLabel(t *testing.T) {
	vs := []Value{{Term: "a", TermLabel: "Zed"}, {Term: "b", TermLabel: "Alpha"}}
	info := []voteInfo{{eff: 1}, {eff: 1}}
	if got := order(vs, info, 0, 1); got >= 0 {
		t.Errorf("order(a, b) = %d, want negative", got)
	}
	if got := order(vs, info, 1, 0); got <= 0 {
		t.Errorf("order(b, a) = %d, want positive", got)
	}
}

func TestPairValuesRefDate(t *testing.T) {
	probe, cand := date(1890, nil, nil), date(1890, nil, nil)
	r := pairValues(ComparerFor(properties.ValueTypeDate), "", []Value{probe}, []Value{cand})
	if r.Probe.Date != probe.Date || r.Candidate.Date != cand.Date || r.Probe.Date == nil {
		t.Errorf("refs = %+v / %+v, want the input date pointers", r.Probe, r.Candidate)
	}
}
