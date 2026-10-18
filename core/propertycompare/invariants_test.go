package propertycompare_test

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

const (
	invariantSeed  = 1
	invariantCases = 500
)

// invariantPools are the values each type draws from. Together they cover
// every relation the type's comparer returns.
var invariantPools = map[string][]V{
	propertymeta.ValueTypeTerm: {term("male"), term("female"), term("intersex")},
	propertymeta.ValueTypeName: {
		name("James Robins"), name("James Robbins"), name("J. Robins"),
		name("Mary Robins"), name("Robins"), name("Mary Smith"),
	},
	propertymeta.ValueTypeText: {
		text("York"), text("york"), text("Montréal"), text("Montreal"),
		text("York, Upper Canada"), text("London"),
	},
	propertymeta.ValueTypeDate: dates(y1817, y1818, y1830, about1817, spring),
}

// favor ranks outcomes from least to most favorable.
var favor = []propertycompare.Outcome{
	propertycompare.OutcomeConflict, propertycompare.OutcomeUnknown,
	propertycompare.OutcomePartial, propertycompare.OutcomeAgree,
}

type invariantCase struct {
	a, b []V
	meta propertymeta.Meta
}

func TestInvariants(t *testing.T) {
	rules := matchregistry.Default().Value
	types := []string{
		propertymeta.ValueTypeTerm, propertymeta.ValueTypeName,
		propertymeta.ValueTypeText, propertymeta.ValueTypeDate,
	}
	invariants := []struct {
		id    string
		check func(t *testing.T, r *rand.Rand, c invariantCase, pool []V, rules valuecompare.Rules) string
	}{
		{"AC-5.1", neutralChangesNothing},
		{"AC-5.2", negativeNeverHelps},
		{"AC-5.3", shuffleChangesNothing},
		{"AC-5.4", swapChangesNothing},
		{"AC-5.5", multipleNeverConflicts},
	}
	for _, inv := range invariants {
		for _, typ := range types {
			t.Run(inv.id+"/"+typ, func(t *testing.T) {
				r := rand.New(rand.NewSource(invariantSeed))
				pool := invariantPools[typ]
				for i := range invariantCases {
					cardinality := propertymeta.CardinalitySingle
					if i%2 == 1 {
						cardinality = propertymeta.CardinalityMultiple
					}
					negatives := inv.id != "AC-5.5"
					if inv.id == "AC-5.5" {
						cardinality = propertymeta.CardinalityMultiple
					}
					nextRef := byte(0)
					c := invariantCase{
						a:    draw(r, pool, negatives, &nextRef),
						b:    draw(r, pool, negatives, &nextRef),
						meta: propertymeta.Meta{ValueType: typ, Cardinality: cardinality},
					}
					if msg := inv.check(t, r, c, pool, rules); msg != "" {
						t.Errorf("case %d: a=%s b=%s: %s", i, labels(c.a), labels(c.b), msg)
					}
				}
			})
		}
	}
}

// draw takes 0 to 3 pool values, each negative one time in four when
// negatives is true, each with its own ref.
func draw(r *rand.Rand, pool []V, negatives bool, next *byte) []V {
	out := make([]V, r.Intn(4))
	for i := range out {
		v := pool[r.Intn(len(pool))]
		v.Negative = negatives && r.Intn(4) == 0
		*next++
		v.Ref = []byte{*next}
		out[i] = v
	}
	return out
}

// sameSpeaker reports how the speaking pair of x differs from y's, or "".
// keys also compares both Same keys.
func sameSpeaker(x, y propertycompare.Result, rules valuecompare.Rules, keys bool) string {
	if x.Outcome != y.Outcome || x.Cause != y.Cause {
		return fmt.Sprintf("outcome, cause %s, %s became %s, %s", x.Outcome, x.Cause, y.Outcome, y.Cause)
	}
	if (x.Speaking == nil) != (y.Speaking == nil) {
		return "speaking pair appeared or vanished"
	}
	if x.Speaking == nil {
		return ""
	}
	if x.Speaking.Result.Relation != y.Speaking.Result.Relation || math.Abs(x.Speaking.Result.Similarity-y.Speaking.Result.Similarity) > 1e-9 {
		return fmt.Sprintf("speaking pair %v became %v", x.Speaking.Result, y.Speaking.Result)
	}
	if keys && (sameKey(x.Speaking.A, rules) != sameKey(y.Speaking.A, rules) || sameKey(x.Speaking.B, rules) != sameKey(y.Speaking.B, rules)) {
		return "speaking pair keys changed"
	}
	return ""
}

func sameKey(v propertycompare.Value, rules valuecompare.Rules) string {
	k, ok := valuecompare.Key(v.Value, valuecompare.Same, rules)
	if !ok {
		return ""
	}
	return k
}

func neutralValue(r *rand.Rand, pool []V) V {
	v := pool[r.Intn(len(pool))]
	if v.Term != "" {
		v = term("unknown") // a term is neutral as "unknown", never as a pool term
	}
	v.Neutral = true
	v.Ref = []byte{0xff}
	return v
}

func neutralChangesNothing(_ *testing.T, r *rand.Rand, c invariantCase, pool []V, rules valuecompare.Rules) string {
	base := propertycompare.Compare(c.a, c.b, c.meta, rules)
	extra := neutralValue(r, pool)
	a, b := c.a, c.b
	if r.Intn(2) == 0 {
		a = append(slices.Clone(a), extra)
	} else {
		b = append(slices.Clone(b), extra)
	}
	if msg := sameSpeaker(base, propertycompare.Compare(a, b, c.meta, rules), rules, true); msg != "" {
		return "adding " + labels([]V{extra}) + ": " + msg
	}
	return ""
}

func negativeNeverHelps(_ *testing.T, r *rand.Rand, c invariantCase, pool []V, rules valuecompare.Rules) string {
	base := propertycompare.Compare(c.a, c.b, c.meta, rules)
	extra := pool[r.Intn(len(pool))]
	extra.Negative = true
	a, b := c.a, c.b
	if r.Intn(2) == 0 {
		a = append(slices.Clone(a), extra)
	} else {
		b = append(slices.Clone(b), extra)
	}
	got := propertycompare.Compare(a, b, c.meta, rules)
	if slices.Index(favor, got.Outcome) > slices.Index(favor, base.Outcome) {
		return fmt.Sprintf("adding %s: outcome %s became %s", labels([]V{extra}), base.Outcome, got.Outcome)
	}
	return ""
}

func shuffleChangesNothing(_ *testing.T, r *rand.Rand, c invariantCase, _ []V, rules valuecompare.Rules) string {
	base := propertycompare.Compare(c.a, c.b, c.meta, rules)
	a, b := slices.Clone(c.a), slices.Clone(c.b)
	r.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	r.Shuffle(len(b), func(i, j int) { b[i], b[j] = b[j], b[i] })
	if msg := sameSpeaker(base, propertycompare.Compare(a, b, c.meta, rules), rules, true); msg != "" {
		return "shuffled to a=" + labels(a) + " b=" + labels(b) + ": " + msg
	}
	return ""
}

func swapChangesNothing(_ *testing.T, _ *rand.Rand, c invariantCase, _ []V, rules valuecompare.Rules) string {
	base := propertycompare.Compare(c.a, c.b, c.meta, rules)
	if msg := sameSpeaker(base, propertycompare.Compare(c.b, c.a, c.meta, rules), rules, false); msg != "" {
		return "swapped: " + msg
	}
	return ""
}

func multipleNeverConflicts(_ *testing.T, r *rand.Rand, c invariantCase, pool []V, rules valuecompare.Rules) string {
	if got := propertycompare.Compare(c.a, c.b, c.meta, rules); got.Outcome == propertycompare.OutcomeConflict {
		return "conflict on a multiple-value property"
	}
	extra := pool[r.Intn(len(pool))]
	a, b := c.a, c.b
	if r.Intn(2) == 0 {
		a = append(slices.Clone(a), extra)
	} else {
		b = append(slices.Clone(b), extra)
	}
	if got := propertycompare.Compare(a, b, c.meta, rules); got.Outcome == propertycompare.OutcomeConflict {
		return "conflict after adding " + labels([]V{extra})
	}
	return ""
}
