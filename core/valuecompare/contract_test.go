package valuecompare_test

import (
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// The contract every comparer keeps (comparison-layers.md §Tests). A value
// type that gains a comparer needs a sample row and a generator here.

const pairsPerType = 500

type samplePair struct{ a, b valuecompare.Value }

func ofText(a, b string) samplePair { return samplePair{textValue(a), textValue(b)} }
func ofTerm(a, b string) samplePair { return samplePair{termValue(a), termValue(b)} }
func ofInt(a, b int64) samplePair   { return samplePair{intValue(a), intValue(b)} }

// samples are hand-picked pairs per value type, read before the generated ones.
var samples = map[string][]samplePair{
	propertymeta.ValueTypeText: {
		ofText("York", "York"), ofText("York", "york"), ofText("Montréal", "Montreal"),
		ofText("York", "York, Upper Canada"), ofText("York, Ontario", "Toronto, York"),
		ofText("Upper Canada, York", "York, Upper Canada"), ofText("Upper Canada", "Upper Kanada"),
		ofText("Ann", "Anne"), ofText("Mary", "Mark"), ofText("!!!", "York"), ofText("", ""),
		ofText("É", "E"), ofText("J. Smith", "J Smith"),
	},
	propertymeta.ValueTypeTerm: {
		ofTerm("male", "male"), ofTerm("male", "female"), ofTerm("", "male"), ofTerm("", ""),
	},
	propertymeta.ValueTypeDate: {
		// AC-1.x pairs, one family per row.
		ofDate(dp(ymd(1817, 0, 0)), dp(ymd(1817, 0, 0))), ofDate(dp(bound(datevalue.QualifierBEF, 1820)), dp(bound(datevalue.QualifierBEF, 1820))),
		ofDate(dp(about(1817, 0, 0)), dp(about(1817, 0, 0))), ofDate(dp(between(1817, 0, 0, 1817, 0, 0)), dp(ymd(1817, 0, 0))),
		ofDate(dp(between(1820, 0, 0, 1815, 0, 0)), dp(between(1815, 0, 0, 1820, 0, 0))),
		ofDate(dp(ymd(1817, 5, 0)), dp(ymd(1817, 5, 14))), ofDate(dp(ymd(1817, 0, 0)), dp(ymd(1817, 9, 0))),
		ofDate(dp(about(1817, 0, 0)), dp(ymd(1818, 0, 0))), ofDate(dp(about(1817, 0, 0)), dp(ymd(1817, 0, 0))),
		ofDate(dp(bound(datevalue.QualifierBEF, 1820)), dp(bound(datevalue.QualifierBEF, 1830))), ofDate(dp(bound(datevalue.QualifierBEF, 1820)), dp(ymd(1815, 0, 0))),
		ofDate(dp(between(1815, 0, 0, 1820, 0, 0)), dp(ymd(1818, 0, 0))),
		ofDate(dp(about(1817, 3, 0)), dp(ymd(1817, 4, 0))), ofDate(dp(about(1817, 3, 3)), dp(ymd(1817, 3, 4))),
		ofDate(dp(about(1817, 1, 0)), dp(ymd(1816, 12, 15))), ofDate(dp(about(1817, 1, 1)), dp(ymd(1816, 12, 31))),
		ofDate(dp(between(1815, 0, 0, 1820, 0, 0)), dp(between(1819, 0, 0, 1825, 0, 0))),
		ofDate(dp(ymd(1817, 5, 14)), dp(ymd(1817, 5, 20))), ofDate(dp(ymd(1817, 5, 0)), dp(ymd(1817, 9, 0))),
		ofDate(dp(ymd(1817, 0, 0)), dp(ymd(1818, 0, 0))), ofDate(dp(ymd(1817, 12, 0)), dp(ymd(1818, 1, 0))),
		ofDate(dp(bound(datevalue.QualifierAFT, 1820)), dp(ymd(1819, 0, 0))), ofDate(dp(between(1815, 0, 0, 1820, 0, 0)), dp(ymd(1822, 0, 0))),
		ofDate(dp(about(1817, 0, 0)), dp(about(1820, 0, 0))),
		ofDate(dp(ymd(1817, 0, 0)), dp(ymd(1820, 0, 0))), ofDate(dp(bound(datevalue.QualifierBEF, 1820)), dp(ymd(1823, 0, 0))),
		ofDate(dp(bound(datevalue.QualifierBEF, 1800)), dp(bound(datevalue.QualifierAFT, 1810))),
		ofDate(dp(phrase("spring")), dp(ymd(1817, 0, 0))), ofDate(dp(phrase("spring")), dp(phrase("spring"))),
		ofDate(dp(noYearRange()), dp(ymd(1817, 0, 0))), ofDate(nil, dp(ymd(1817, 0, 0))), ofDate(nil, nil),
	},
	propertymeta.ValueTypeInteger: {
		ofInt(3, 3), ofInt(3, 4), ofInt(3, 6), ofInt(0, 0), ofInt(-2, 2),
		ofInt(math.MaxInt64, math.MaxInt64-1), ofInt(math.MaxInt64, math.MinInt64), ofInt(math.MinInt64, math.MinInt64+2),
	},
}

var wordPool = []string{
	"York", "york", "YORK", "Montréal", "Montreal", "Upper", "Canada", "Kanada", "O'Brien", "O Brien",
	"Mc Donald", "McDonald", "Ann", "Anne", "Provenance", "Provenanced", "É", "E", "J.", "J", "Ontario",
	"Toronto", "Sainte-Anne", "Sainte Anne", "Groß", "Gross", "!!!",
}

func randomText(rng *rand.Rand) string {
	n := 1 + rng.Intn(4)
	words := make([]string, n)
	for i := range words {
		words[i] = wordPool[rng.Intn(len(wordPool))]
	}
	sep := []string{" ", ", ", "-"}[rng.Intn(3)]
	return strings.Join(words, sep)
}

// variantText is s changed the way two records of one place differ.
func variantText(rng *rand.Rand, s string) string {
	words := strings.Fields(strings.ReplaceAll(s, ",", " "))
	if len(words) == 0 {
		return s
	}
	switch rng.Intn(6) {
	case 0:
		return strings.ToLower(s)
	case 1:
		rng.Shuffle(len(words), func(i, j int) { words[i], words[j] = words[j], words[i] })
	case 2:
		i := rng.Intn(len(words))
		words = slices.Delete(words, i, i+1)
	case 3:
		words = append(words, wordPool[rng.Intn(len(wordPool))])
	case 4:
		words[rng.Intn(len(words))] = wordPool[rng.Intn(len(wordPool))]
	}
	return strings.Join(words, ", ")
}

var generators = map[string]func(rng *rand.Rand) samplePair{
	propertymeta.ValueTypeText: func(rng *rand.Rand) samplePair {
		a := randomText(rng)
		if rng.Intn(3) == 0 {
			return ofText(a, randomText(rng))
		}
		return ofText(a, variantText(rng, a))
	},
	propertymeta.ValueTypeTerm: func(rng *rand.Rand) samplePair {
		keys := []string{"male", "female", "unknown", "other", ""}
		return ofTerm(keys[rng.Intn(len(keys))], keys[rng.Intn(len(keys))])
	},
	propertymeta.ValueTypeDate: func(rng *rand.Rand) samplePair {
		a, b := randomDate(rng), randomDate(rng)
		return ofDate(&a, &b)
	},
	propertymeta.ValueTypeInteger: func(rng *rand.Rand) samplePair {
		return ofInt(int64(rng.Intn(10)), int64(rng.Intn(10)))
	},
}

// randomDate draws one side of a date pair: a point, ABT, BEF or AFT point, a
// range, or a phrase with no year.
func randomDate(rng *rand.Rand) datevalue.Value {
	draw := func() (y, m, d int) {
		y = 1815 + rng.Intn(7)
		if rng.Intn(2) == 0 {
			return y, 0, 0
		}
		m = 1 + rng.Intn(12)
		if rng.Intn(2) == 0 {
			return y, m, 0
		}
		return y, m, 1 + rng.Intn(28)
	}
	switch n := rng.Intn(10); {
	case n < 4:
		return ymd(draw())
	case n == 4:
		return about(draw())
	case n == 5:
		y, m, d := draw()
		v := ymd(y, m, d)
		v.Qualifier = datevalue.QualifierBEF
		return v
	case n == 6:
		y, m, d := draw()
		v := ymd(y, m, d)
		v.Qualifier = datevalue.QualifierAFT
		return v
	case n < 9:
		y1, m1, d1 := draw()
		y2, m2, d2 := draw()
		if datevalue.DaysFromCivil(y1, max(m1, 1), max(d1, 1)) > datevalue.DaysFromCivil(y2, max(m2, 1), max(d2, 1)) {
			y1, m1, d1, y2, m2, d2 = y2, m2, d2, y1, m1, d1
		}
		return between(y1, m1, d1, y2, m2, d2)
	}
	return phrase("spring")
}

// contractPairs are a type's sample rows, then its generated pairs.
func contractPairs(typ string, seed int64) []samplePair {
	rng := rand.New(rand.NewSource(seed))
	pairs := slices.Clone(samples[typ])
	for range pairsPerType {
		pairs = append(pairs, generators[typ](rng))
	}
	return pairs
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestComparerContract(t *testing.T) {
	allTypes := []string{
		propertymeta.ValueTypeText, propertymeta.ValueTypeInteger, propertymeta.ValueTypeDate,
		propertymeta.ValueTypeName, propertymeta.ValueTypeSubject, propertymeta.ValueTypeTerm,
		propertymeta.ValueTypeDepiction,
	}
	supported := valuecompare.ComparerTypes()
	for _, typ := range allTypes {
		if !slices.Contains(supported, typ) {
			continue
		}
		t.Run(typ, func(t *testing.T) {
			if len(samples[typ]) == 0 || generators[typ] == nil {
				t.Fatalf("value type %s has a comparer but no sample row and generator in the contract test", typ)
			}
			for seed, tolerance := range []int64{0, 1, 2, 3} {
				rules := matchregistry.Default().Value
				rules.Integers.Tolerance = tolerance
				shuffled := matchregistry.Default().Value
				shuffled.Integers.Tolerance = tolerance
				rand.New(rand.NewSource(int64(seed))).Shuffle(len(shuffled.Words.Particles), func(i, j int) {
					shuffled.Words.Particles[i], shuffled.Words.Particles[j] = shuffled.Words.Particles[j], shuffled.Words.Particles[i]
				})
				for _, p := range contractPairs(typ, int64(seed)) {
					res, det := valuecompare.Compare(p.a, p.b, rules)
					label := func() string { return typ + ": " + p.a.Text + p.a.Term + " / " + p.b.Text + p.b.Term }
					if res.Similarity < 0 || res.Similarity > 1 {
						t.Fatalf("%s: similarity %v out of range", label(), res.Similarity)
					}
					if (res.Relation == valuecompare.Same) != (res.Similarity == 1) {
						t.Fatalf("%s: %s with similarity %v", label(), res.Relation, res.Similarity)
					}
					if res.Relation != valuecompare.Unknown && (res.Relation == valuecompare.Different) != (res.Similarity == 0) {
						t.Fatalf("%s: %s with similarity %v", label(), res.Relation, res.Similarity)
					}
					switch res.Relation {
					case valuecompare.Equivalent:
						if res.Similarity < 0.9 {
							t.Fatalf("%s: equivalent at %v", label(), res.Similarity)
						}
					case valuecompare.Contains, valuecompare.Overlaps, valuecompare.Near:
						if res.Similarity <= 0 || res.Similarity >= 1 {
							t.Fatalf("%s: %s at %v", label(), res.Relation, res.Similarity)
						}
					case valuecompare.Unknown:
						if res.Similarity != 0 {
							t.Fatalf("%s: unknown at %v", label(), res.Similarity)
						}
					}
					// Swapped sides.
					swapped, swappedDet := valuecompare.Compare(p.b, p.a, rules)
					if swapped.Relation != res.Relation || !near(swapped.Similarity, res.Similarity) {
						t.Fatalf("%s: swapped %+v, want %+v", label(), swapped, res)
					}
					if res.Relation == valuecompare.Contains && swappedDet.Broader == det.Broader {
						t.Fatalf("%s: Broader %q did not flip", label(), det.Broader)
					}
					// Particles in another order.
					other, _ := valuecompare.Compare(p.a, p.b, shuffled)
					if other.Relation != res.Relation || !near(other.Similarity, res.Similarity) {
						t.Fatalf("%s: shuffled particles %+v, want %+v", label(), other, res)
					}
					// A value against itself.
					for _, v := range []valuecompare.Value{p.a, p.b} {
						self, selfDet := valuecompare.Compare(v, v, rules)
						if self.Relation != valuecompare.Same && !(self.Relation == valuecompare.Unknown && (selfDet.Cause == valuecompare.CauseEmpty || selfDet.Cause == valuecompare.CausePhraseOnly)) {
							t.Fatalf("%s: against itself %+v %+v", label(), self, selfDet)
						}
					}
				}
			}
		})
	}
}

// Key agrees with Compare: equal at Same exactly when Same, equal at
// Equivalent exactly when Same or Equivalent.
func TestKeyContract(t *testing.T) {
	supported := valuecompare.ComparerTypes()
	for _, typ := range supported {
		t.Run(typ, func(t *testing.T) {
			if len(samples[typ]) == 0 || generators[typ] == nil {
				t.Fatalf("value type %s has a comparer but no sample row and generator in the contract test", typ)
			}
			rules := matchregistry.Default().Value
			for _, p := range contractPairs(typ, 1) {
				res, _ := valuecompare.Compare(p.a, p.b, rules)
				for _, level := range []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent} {
					ka, oka := valuecompare.Key(p.a, level, rules)
					kb, okb := valuecompare.Key(p.b, level, rules)
					equal := oka && okb && ka == kb
					want := res.Relation == valuecompare.Same
					if level == valuecompare.Equivalent {
						want = want || res.Relation == valuecompare.Equivalent
					}
					if equal != want {
						t.Fatalf("%s at %s: keys %q/%q (%v/%v), relation %s", typ, level, ka, kb, oka, okb, res.Relation)
					}
				}
				for _, level := range []valuecompare.Relation{valuecompare.Near, valuecompare.Contains, valuecompare.Different, valuecompare.Unknown} {
					if k, ok := valuecompare.Key(p.a, level, rules); ok || k != "" {
						t.Fatalf("%s: Key at %s = %q, %v", typ, level, k, ok)
					}
				}
			}
		})
	}
	// An empty value has no key.
	rules := matchregistry.Default().Value
	for _, v := range []valuecompare.Value{textValue(""), termValue("")} {
		if k, ok := valuecompare.Key(v, valuecompare.Same, rules); ok || k != "" {
			t.Errorf("Key of an empty %s = %q, %v", v.Type, k, ok)
		}
	}
}
