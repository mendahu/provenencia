package propertycompare_test

import (
	"math"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

type V = propertycompare.Value

var (
	single   = propertymeta.CardinalitySingle
	multiple = propertymeta.CardinalityMultiple
	none     = valuecompare.SideNone
)

// wantPair is the expected speaking pair. sim < 0 leaves the similarity unchecked.
type wantPair struct {
	a, b       string
	aNeg, bNeg bool
	rel        valuecompare.Relation
	sim        float64
	detCause   valuecompare.Cause // checked when set
	broader    valuecompare.Side  // checked when set
}

type wantExcluded struct {
	side  valuecompare.Side
	value string
}

type compareCase struct {
	name     string
	valType  string
	card     string
	a, b     []V
	outcome  propertycompare.Outcome
	cause    propertycompare.Cause
	pair     *wantPair
	excluded []wantExcluded
}

func TestCompare(t *testing.T) {
	const (
		tTerm = propertymeta.ValueTypeTerm
		tName = propertymeta.ValueTypeName
		tText = propertymeta.ValueTypeText
		tDate = propertymeta.ValueTypeDate
	)
	agree, partial, conflict, unknown := propertycompare.OutcomeAgree, propertycompare.OutcomePartial, propertycompare.OutcomeConflict, propertycompare.OutcomeUnknown
	rel, negative, noValues, card := propertycompare.CauseRelation, propertycompare.CauseNegative, propertycompare.CauseNoValues, propertycompare.CauseCardinality
	same, equiv, contains, nearR, different, unk := valuecompare.Same, valuecompare.Equivalent, valuecompare.Contains, valuecompare.Near, valuecompare.Different, valuecompare.Unknown

	cases := []compareCase{
		// US-1
		{name: "AC-1.1 term same", valType: tTerm, card: single, a: []V{term("male")}, b: []V{term("male")},
			outcome: agree, cause: rel, pair: &wantPair{a: "male", b: "male", rel: same, sim: 1}},
		{name: "AC-1.2 term different", valType: tTerm, card: single, a: []V{term("male")}, b: []V{term("female")},
			outcome: conflict, cause: rel, pair: &wantPair{a: "male", b: "female", rel: different, sim: 0}},
		{name: "AC-1.3 name near", valType: tName, card: single, a: []V{name("James Robins")}, b: []V{name("James Robbins")},
			outcome: partial, cause: rel, pair: &wantPair{a: "James Robins", b: "James Robbins", rel: nearR, sim: 0.9143, detCause: valuecompare.CauseSpelling}},
		{name: "AC-1.4 name contains", valType: tName, card: single, a: []V{name("J. Robins")}, b: []V{name("James Robins")},
			outcome: partial, cause: rel, pair: &wantPair{a: "J. Robins", b: "James Robins", rel: contains, sim: 0.8, broader: valuecompare.SideA}},
		{name: "AC-1.5 text equivalent", valType: tText, card: single, a: []V{text("Montréal")}, b: []V{text("Montreal")},
			outcome: agree, cause: rel, pair: &wantPair{a: "Montréal", b: "Montreal", rel: equiv, sim: 0.95}},
		{name: "AC-1.6 multiple difference", valType: tText, card: multiple, a: []V{text("York")}, b: []V{text("London")},
			outcome: unknown, cause: card, pair: &wantPair{a: "York", b: "London", rel: different, sim: -1}},
		{name: "AC-1.7 multiple shared value", valType: tText, card: multiple, a: []V{text("York"), text("Eboracum")}, b: []V{text("Eboracum")},
			outcome: agree, cause: rel, pair: &wantPair{a: "Eboracum", b: "Eboracum", rel: same, sim: 1}},
		{name: "AC-1.8 phrase-only date", valType: tDate, card: single, a: dates(spring), b: dates(y1817),
			outcome: unknown, cause: rel, pair: &wantPair{a: "spring", b: "1817", rel: unk, sim: 0, detCause: valuecompare.CausePhraseOnly}},
		{name: "AC-1.9 no values on one side", valType: tTerm, card: single, a: nil, b: []V{term("male")},
			outcome: unknown, cause: noValues},
		{name: "AC-1.9 no values on either side", valType: tTerm, card: single, a: nil, b: nil,
			outcome: unknown, cause: noValues},
		{name: "AC-1.10 unsupported value type", valType: propertymeta.ValueTypeSubject, card: single, a: []V{text("x")}, b: []V{text("y")},
			outcome: unknown, cause: rel, pair: &wantPair{a: "x", b: "y", rel: unk, sim: 0, detCause: valuecompare.CauseUnsupported}},
		{name: "AC-1.11 empty cardinality reads as single", valType: tTerm, card: "", a: []V{term("male")}, b: []V{term("female")},
			outcome: conflict, cause: rel, pair: &wantPair{a: "male", b: "female", rel: different, sim: 0}},
		{name: "AC-1.12 the property sets the type", valType: tText, card: single, a: []V{text("York")}, b: []V{text("york")},
			outcome: agree, cause: rel, pair: &wantPair{a: "York", b: "york", rel: same, sim: 1}},
		{name: "D-4 a caller's type is overwritten", valType: tText, card: single,
			a: []V{{Value: valuecompare.Value{Text: "York", Type: propertymeta.ValueTypeName}}}, b: []V{text("york")},
			outcome: agree, cause: rel, pair: &wantPair{a: "York", b: "york", rel: same, sim: 1}},

		// US-2
		{name: "AC-2.1 a matching value speaks", valType: tName, card: single, a: []V{name("Mary Smith"), name("James Robins")}, b: []V{name("James Robins")},
			outcome: agree, cause: rel, pair: &wantPair{a: "James Robins", b: "James Robins", rel: same, sim: 1}},
		{name: "AC-2.2 higher similarity speaks", valType: tName, card: single, a: []V{name("James Robins")}, b: []V{name("J. Robins"), name("James Robbins")},
			outcome: partial, cause: rel, pair: &wantPair{a: "James Robins", b: "James Robbins", rel: nearR, sim: 0.9143}},
		{name: "AC-2.3 a different pair does not hide a near one", valType: tName, card: single, a: []V{name("James Robbins")}, b: []V{name("J. Robins"), name("James Robins")},
			outcome: partial, cause: rel, pair: &wantPair{a: "James Robbins", b: "James Robins", rel: nearR, sim: 0.9143}},
		{name: "AC-2.4 a different pair speaks before an unknown one", valType: tDate, card: single, a: dates(spring, y1817), b: dates(y1830),
			outcome: conflict, cause: rel, pair: &wantPair{a: "1817", b: "1830", rel: different, sim: 0, detCause: valuecompare.CauseBeyondTolerance}},

		// US-3
		{name: "AC-3.1 negative against its value", valType: tTerm, card: single, a: []V{neg(term("male"))}, b: []V{term("male")},
			outcome: conflict, cause: negative, pair: &wantPair{a: "male", b: "male", aNeg: true, rel: same, sim: 1}},
		{name: "AC-3.2 a positive does not rescue the negative", valType: tTerm, card: single, a: []V{term("female"), neg(term("male"))}, b: []V{term("male")},
			outcome: conflict, cause: negative, pair: &wantPair{a: "male", b: "male", aNeg: true, rel: same, sim: 1}},
		{name: "AC-3.3 negative against a contained name", valType: tName, card: single, a: []V{neg(name("J. Robins"))}, b: []V{name("James Robins")},
			outcome: conflict, cause: negative, pair: &wantPair{a: "J. Robins", b: "James Robins", aNeg: true, rel: contains, sim: -1}},
		{name: "AC-3.4 negative against an equivalent text", valType: tText, card: single, a: []V{neg(text("Montréal"))}, b: []V{text("Montreal")},
			outcome: conflict, cause: negative, pair: &wantPair{a: "Montréal", b: "Montreal", aNeg: true, rel: equiv, sim: -1}},
		{name: "AC-3.5 near does not contradict", valType: tName, card: single, a: []V{neg(name("James Robbins"))}, b: []V{name("James Robins")},
			outcome: unknown, cause: noValues},
		{name: "AC-3.6 different does not contradict", valType: tTerm, card: single, a: []V{neg(term("female"))}, b: []V{term("male")},
			outcome: unknown, cause: noValues},
		{name: "AC-3.7 two negatives never meet", valType: tTerm, card: single, a: []V{neg(term("male"))}, b: []V{neg(term("male"))},
			outcome: unknown, cause: noValues},
		{name: "AC-3.8 negative on side b", valType: tTerm, card: single, a: []V{term("male")}, b: []V{neg(term("male")), term("female")},
			outcome: conflict, cause: negative, pair: &wantPair{a: "male", b: "male", bNeg: true, rel: same, sim: 1}},
		{name: "AC-3.9 negative on a multiple-value property", valType: tText, card: multiple, a: []V{neg(text("Eboracum"))}, b: []V{text("York"), text("Eboracum")},
			outcome: conflict, cause: negative, pair: &wantPair{a: "Eboracum", b: "Eboracum", aNeg: true, rel: same, sim: 1}},

		{name: "D-6 tie on contradicting pairs keeps A's input order", valType: tTerm, card: single, a: []V{neg(term("male")), term("male")}, b: []V{term("male"), neg(term("male"))},
			outcome: conflict, cause: negative, pair: &wantPair{a: "male", b: "male", aNeg: true, rel: same, sim: 1}},

		// US-4
		{name: "AC-4.1 a neutral value is no evidence", valType: tTerm, card: single, a: []V{neutral(term("unknown"))}, b: []V{term("male")},
			outcome: unknown, cause: noValues, excluded: []wantExcluded{{valuecompare.SideA, "unknown"}}},
		{name: "AC-4.2 the rest still compares", valType: tTerm, card: single, a: []V{neutral(term("unknown")), term("female")}, b: []V{term("male")},
			outcome: conflict, cause: rel, pair: &wantPair{a: "female", b: "male", rel: different, sim: 0},
			excluded: []wantExcluded{{valuecompare.SideA, "unknown"}}},
		{name: "AC-4.3 neutral negatives drop too", valType: tTerm, card: single, a: []V{neg(neutral(term("unknown")))}, b: []V{neutral(term("unknown"))},
			outcome: unknown, cause: noValues, excluded: []wantExcluded{{valuecompare.SideA, "unknown"}, {valuecompare.SideB, "unknown"}}},
	}

	// AC-2.5: the same pair speaks in each of the six orders of side A.
	smiths := []V{name("Anna Smith"), name("Bea Smith"), name("Cora Smith")}
	for i, order := range permutations(smiths) {
		cases = append(cases, compareCase{
			name: "AC-2.5 order " + string(rune('1'+i)), valType: tName, card: single, a: order, b: []V{name("Zed Jones")},
			outcome: conflict, cause: rel, pair: &wantPair{a: "Anna Smith", b: "Zed Jones", rel: different, sim: -1},
		})
	}

	rules := matchregistry.Default().Value
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := propertycompare.Compare(tc.a, tc.b, propertymeta.Meta{ValueType: tc.valType, Cardinality: tc.card}, rules)
			if got.Outcome != tc.outcome || got.Cause != tc.cause {
				t.Errorf("outcome, cause = %s, %s; want %s, %s", got.Outcome, got.Cause, tc.outcome, tc.cause)
			}
			checkPair(t, got.Speaking, tc.pair, tc.valType)
			if len(got.Excluded) != len(tc.excluded) {
				t.Fatalf("Excluded = %d entries, want %d", len(got.Excluded), len(tc.excluded))
			}
			for i, e := range tc.excluded {
				x := got.Excluded[i]
				if x.Side != e.side || label(x.Value) != e.value || x.Reason != propertycompare.ExcludeNeutral || x.Value.Type != tc.valType {
					t.Errorf("Excluded[%d] = %v %q %s type %q; want %v %q neutral", i, x.Side, label(x.Value), x.Reason, x.Value.Type, e.side, e.value)
				}
			}
			if tc.excluded == nil && got.Excluded != nil {
				t.Errorf("Excluded = %v, want nil", got.Excluded)
			}
		})
	}
}

func checkPair(t *testing.T, got *propertycompare.SpeakingPair, want *wantPair, valType string) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Errorf("Speaking = %+v, want nil", got)
		}
		return
	}
	if got == nil {
		t.Fatal("Speaking = nil")
	}
	if label(got.A) != want.a || label(got.B) != want.b || got.A.Negative != want.aNeg || got.B.Negative != want.bNeg {
		t.Errorf("pair = %q (neg %v) / %q (neg %v), want %q (neg %v) / %q (neg %v)",
			label(got.A), got.A.Negative, label(got.B), got.B.Negative, want.a, want.aNeg, want.b, want.bNeg)
	}
	if got.A.Type != valType || got.B.Type != valType {
		t.Errorf("pair types = %q, %q; want %q", got.A.Type, got.B.Type, valType)
	}
	if got.Result.Relation != want.rel {
		t.Errorf("relation = %s, want %s", got.Result.Relation, want.rel)
	}
	if want.sim >= 0 && math.Abs(got.Result.Similarity-want.sim) > 0.001 {
		t.Errorf("similarity = %v, want %v", got.Result.Similarity, want.sim)
	}
	if want.detCause != "" && got.Detail.Cause != want.detCause {
		t.Errorf("detail cause = %s, want %s", got.Detail.Cause, want.detCause)
	}
	if want.broader != none && got.Detail.Broader != want.broader {
		t.Errorf("broader = %q, want %q", got.Detail.Broader, want.broader)
	}
}

func TestCompareRefs(t *testing.T) {
	// AC-2.6: refs are echoed, never read.
	got := propertycompare.Compare(
		[]V{ref(term("male"), 0x07)}, []V{ref(term("male"), 0x09)},
		propertymeta.Meta{ValueType: propertymeta.ValueTypeTerm, Cardinality: single}, matchregistry.Default().Value)
	p := got.Speaking
	if p == nil {
		t.Fatal("Speaking = nil")
	}
	if !slices.Equal(p.A.Ref, []byte{0x07}) || !slices.Equal(p.B.Ref, []byte{0x09}) ||
		!slices.Equal(p.Detail.RefA, []byte{0x07}) || !slices.Equal(p.Detail.RefB, []byte{0x09}) {
		t.Errorf("refs = %x %x / %x %x", p.A.Ref, p.B.Ref, p.Detail.RefA, p.Detail.RefB)
	}
}

// permutations returns every order of vs.
func permutations(vs []V) [][]V {
	if len(vs) <= 1 {
		return [][]V{slices.Clone(vs)}
	}
	var out [][]V
	for i := range vs {
		rest := slices.Concat(vs[:i], vs[i+1:])
		for _, p := range permutations(rest) {
			out = append(out, append([]V{vs[i]}, p...))
		}
	}
	return out
}
