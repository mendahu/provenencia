package valuecompare_test

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
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
func ofName(a, b string) samplePair { return samplePair{nv(a), nv(b)} }

// pic is a depiction value whose image File id is 16 bytes of id; 0 is none.
func pic(id byte) valuecompare.Value {
	v := valuecompare.Value{Type: propertymeta.ValueTypeDepiction}
	if id != 0 {
		v.Depiction = &depictionvalue.Value{ImageFileID: bytes.Repeat([]byte{id}, 16)}
	}
	return v
}

func ofPic(a, b byte) samplePair { return samplePair{pic(a), pic(b)} }

// samples are hand-picked pairs per value type, read before the generated ones.
var samples = map[string][]samplePair{
	propertymeta.ValueTypeText: {
		ofText("York", "York"), ofText("York", "york"), ofText("Montréal", "Montreal"),
		ofText("York", "York, Upper Canada"), ofText("York, Ontario", "Toronto, York"),
		ofText("Upper Canada, York", "York, Upper Canada"), ofText("Upper Canada", "Upper Kanada"),
		ofText("Ann", "Anne"), ofText("Mary", "Mark"), ofText("!!!", "York"), ofText("", ""),
		ofText("É", "E"), ofText("J. Smith", "J Smith"),
		ofText("god-father", "godfather"), ofText("ab c", "a bc"), ofText("Mont-réal", "Montreal"),
	},
	propertymeta.ValueTypeTerm: {
		ofTerm("male", "male"), ofTerm("male", "female"), ofTerm("", "male"), ofTerm("", ""),
	},
	propertymeta.ValueTypeName: {
		ofName("g=James|s=Robins", "g=James|s=Robins"), ofName("g=JAMES|s=Robins.", "g=james|s=robins"),
		ofName("s=Robins|g=James", "g=James|s=Robins"), ofName("p=Rev.|g=James|s=Robins", "p=Dr.|g=James|s=Robins"),
		ofName("g=Vincent|sp=van|s=Gogh", "g=Vincent|s=Gogh"), ofName("s=蒋|g=浩", "g=浩|s=蒋"),
		ofName("g=José|s=Núñez", "g=Jose|s=Nunez"), ofName("g=Jas.|s=Robins", "g=James|s=Robins"),
		ofName("g=Jim", "g=James"), ofName("s=Bill", "s=William"),
		ofName("s=Robins", "g=James|s=Robins"), ofName("g=James", "g=James|s=Robins"),
		ofName("g=James|s=Robins", "g=James|g=William|s=Robins"), ofName("g=James|s=Robins|x=Jr.", "g=James|s=Robins"),
		ofName("g=J.|s=Robins", "g=James|s=Robins"), ofName("g=James|s=Robbins", "g=James|s=Robins"),
		ofName("g=Ann|s=Robins", "g=Anne|s=Robins"), ofName("g=James|s=Robins", "u=James|u=Robins"),
		ofName("g=James|g=Kenneth|s=Robins", "g=James|g=William|s=Robins"), ofName("g=James|g=K.|s=Robins", "g=James|g=W.|s=Robins"),
		ofName("g=Ann", "g=Anne|s=Robins"), ofName("g=Mary|s=Robins", "g=James|s=Robins"),
		ofName("g=Mary|s=Smith", "g=Mary|s=Robins"), ofName("g=Maria|s=Silva Costa", "g=Maria|s=Costa Pereira"),
		ofName("g=W.|s=Robins", "g=James|s=Robins"), ofName("g=Ada|s=Lovelace", "g=James|s=Robins"),
		ofName("g=Robins|s=James", "g=James|s=Robins"), ofName("g=J.|s=Robbins", "g=James|s=Robins"),
		ofName("g=James|s=Robins|x=Jr.", "g=James|s=Robins|x=Sr."), ofName("p=Rev.|x=Jr.", "g=James"),
		ofName("g=Des|s=Long", "g=Desiree|s=Long"), ofName("g=Desiree", "g=Des"), ofName("n=Des", "g=Desiree"),
		ofName("g=Mary|s=Mac", "g=Mary|s=MacDonald"), ofName("g=James|g=Des", "g=J.|g=Desiree"),
		{formOnly("James Robins"), nv("g=James")}, {valuecompare.Value{Type: propertymeta.ValueTypeName}, nv("g=James")},
	},
	propertymeta.ValueTypeDate: {
		// AC-1.x pairs, one family per row.
		ofDate(datevaluetest.YMD(1817, 0, 0), datevaluetest.YMD(1817, 0, 0)), ofDate(bound(datevalue.QualifierBEF, 1820), bound(datevalue.QualifierBEF, 1820)),
		ofDate(datevaluetest.About(1817, 0, 0), datevaluetest.About(1817, 0, 0)), ofDate(datevaluetest.Between(1817, 0, 0, 1817, 0, 0), datevaluetest.YMD(1817, 0, 0)),
		ofDate(datevaluetest.Between(1820, 0, 0, 1815, 0, 0), datevaluetest.Between(1815, 0, 0, 1820, 0, 0)),
		ofDate(datevaluetest.YMD(1817, 5, 0), datevaluetest.YMD(1817, 5, 14)), ofDate(datevaluetest.YMD(1817, 0, 0), datevaluetest.YMD(1817, 9, 0)),
		ofDate(datevaluetest.About(1817, 0, 0), datevaluetest.YMD(1818, 0, 0)), ofDate(datevaluetest.About(1817, 0, 0), datevaluetest.YMD(1817, 0, 0)),
		ofDate(bound(datevalue.QualifierBEF, 1820), bound(datevalue.QualifierBEF, 1830)), ofDate(bound(datevalue.QualifierBEF, 1820), datevaluetest.YMD(1815, 0, 0)),
		ofDate(datevaluetest.Between(1815, 0, 0, 1820, 0, 0), datevaluetest.YMD(1818, 0, 0)),
		ofDate(datevaluetest.About(1817, 3, 0), datevaluetest.YMD(1817, 4, 0)), ofDate(datevaluetest.About(1817, 3, 3), datevaluetest.YMD(1817, 3, 4)),
		ofDate(datevaluetest.About(1817, 1, 0), datevaluetest.YMD(1816, 12, 15)), ofDate(datevaluetest.About(1817, 1, 1), datevaluetest.YMD(1816, 12, 31)),
		ofDate(datevaluetest.Between(1815, 0, 0, 1820, 0, 0), datevaluetest.Between(1819, 0, 0, 1825, 0, 0)),
		ofDate(datevaluetest.YMD(1817, 5, 14), datevaluetest.YMD(1817, 5, 20)), ofDate(datevaluetest.YMD(1817, 5, 0), datevaluetest.YMD(1817, 9, 0)),
		ofDate(datevaluetest.YMD(1817, 0, 0), datevaluetest.YMD(1818, 0, 0)), ofDate(datevaluetest.YMD(1817, 12, 0), datevaluetest.YMD(1818, 1, 0)),
		ofDate(bound(datevalue.QualifierAFT, 1820), datevaluetest.YMD(1819, 0, 0)), ofDate(datevaluetest.Between(1815, 0, 0, 1820, 0, 0), datevaluetest.YMD(1822, 0, 0)),
		ofDate(datevaluetest.About(1817, 0, 0), datevaluetest.About(1820, 0, 0)),
		ofDate(datevaluetest.YMD(1817, 0, 0), datevaluetest.YMD(1820, 0, 0)), ofDate(bound(datevalue.QualifierBEF, 1820), datevaluetest.YMD(1823, 0, 0)),
		ofDate(bound(datevalue.QualifierBEF, 1800), bound(datevalue.QualifierAFT, 1810)),
		ofDate(datevaluetest.Phrase("spring"), datevaluetest.YMD(1817, 0, 0)), ofDate(datevaluetest.Phrase("spring"), datevaluetest.Phrase("spring")),
		ofDate(noYearRange(), datevaluetest.YMD(1817, 0, 0)), ofDate(nil, datevaluetest.YMD(1817, 0, 0)), ofDate(nil, nil),
	},
	propertymeta.ValueTypeDepiction: {
		ofPic(1, 1), ofPic(1, 2), ofPic(0, 1), ofPic(0, 0),
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

// Name parts come from pools per part code (the codes of nv).
var namePools = map[string][]string{
	"g":  {"James", "Jim", "Mary", "William", "Johann", "Johan", "Jon", "John", "José", "Jose", "Maria Luisa", "浩", "J.", "K.", "W", "M.", "Des", "Desiree", "Al", "Albert"},
	"n":  {"Jim", "Bill", "Polly"},
	"s":  {"Robins", "Robbins", "Smith", "Smyth", "Silva Costa", "Costa", "Gogh", "Núñez", "Nunez", "蒋", "Mac", "MacDonald"},
	"sp": {"van", "von", "de"},
	"x":  {"Jr.", "Sr.", "III"},
	"p":  {"Rev.", "Dr."},
	"u":  {"Kendall"},
	"_":  {"Robins", "Robbins", "Smith", "James"},
}

var namePoolCodes = []string{"g", "g", "g", "n", "s", "s", "sp", "x", "p", "u", "_"}

// randomName is a name of 0 to 4 parts from the pools.
func randomName(rng *rand.Rand) valuecompare.Value {
	var spec []string
	k := rng.Intn(5)
	if k == 0 && rng.Intn(3) > 0 {
		k = 1 + rng.Intn(4) // an empty name is the rare case
	}
	for i := 0; i < k; i++ {
		code := namePoolCodes[rng.Intn(len(namePoolCodes))]
		pool := namePools[code]
		spec = append(spec, code+"="+pool[rng.Intn(len(pool))])
	}
	return namePartsOf(spec)
}

// namePartsOf is a name of "code=value" parts; none makes a name with only a form.
func namePartsOf(spec []string) valuecompare.Value {
	if len(spec) == 0 {
		return formOnly("(as written)")
	}
	return nv(strings.Join(spec, "|"))
}

// variantName is v changed the way two records of one person differ: case, a
// dropped part, an added part, a spelling from the pool, or parts reordered.
func variantName(rng *rand.Rand, v valuecompare.Value) valuecompare.Value {
	parts := slices.Clone(v.Name.Parts)
	if len(parts) == 0 {
		return v
	}
	switch rng.Intn(6) {
	case 0:
		for i := range parts {
			parts[i].Value = strings.ToUpper(parts[i].Value)
		}
	case 5:
		unaccent := strings.NewReplacer("é", "e", "ñ", "n", "ú", "u", "É", "E")
		for i := range parts {
			parts[i].Value = unaccent.Replace(parts[i].Value)
		}
	case 1:
		i := rng.Intn(len(parts))
		parts = slices.Delete(parts, i, i+1)
	case 2:
		code := namePoolCodes[rng.Intn(len(namePoolCodes))]
		pool := namePools[code]
		parts = append(parts, namevalue.Part{Idx: len(parts), Type: partCodes[code], Value: pool[rng.Intn(len(pool))]})
	case 3:
		i := rng.Intn(len(parts))
		for code, typ := range partCodes {
			if typ == parts[i].Type {
				pool := namePools[code]
				parts[i].Value = pool[rng.Intn(len(pool))]
				break
			}
		}
	default:
		rng.Shuffle(len(parts), func(i, j int) { parts[i], parts[j] = parts[j], parts[i] })
	}
	n := *v.Name
	n.Parts = parts
	v.Name = &n
	return v
}

var generators = map[string]func(rng *rand.Rand) samplePair{
	propertymeta.ValueTypeName: func(rng *rand.Rand) samplePair {
		a := randomName(rng)
		if rng.Intn(2) == 0 {
			return samplePair{a, randomName(rng)}
		}
		return samplePair{a, variantName(rng, a)}
	},
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
	propertymeta.ValueTypeDepiction: func(rng *rand.Rand) samplePair {
		return ofPic(byte(rng.Intn(4)), byte(rng.Intn(4)))
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
		return *datevaluetest.YMD(draw())
	case n == 4:
		return *datevaluetest.About(draw())
	case n == 5:
		y, m, d := draw()
		v := *datevaluetest.YMD(y, m, d)
		v.Qualifier = datevalue.QualifierBEF
		return v
	case n == 6:
		y, m, d := draw()
		v := *datevaluetest.YMD(y, m, d)
		v.Qualifier = datevalue.QualifierAFT
		return v
	case n < 9:
		y1, m1, d1 := draw()
		y2, m2, d2 := draw()
		if datevalue.DaysFromCivil(y1, max(m1, 1), max(d1, 1)) > datevalue.DaysFromCivil(y2, max(m2, 1), max(d2, 1)) {
			y1, m1, d1, y2, m2, d2 = y2, m2, d2, y1, m1, d1
		}
		return *datevaluetest.Between(y1, m1, d1, y2, m2, d2)
	}
	return *datevaluetest.Phrase("spring")
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

// mirrored reports whether b is a's detail from the other side: the same
// pairs with A and B exchanged, and OnlyA and OnlyB exchanged, in any order.
func mirrored(a, b valuecompare.Detail) bool {
	flip := func(parts []valuecompare.PartMatch) []string {
		out := make([]string, len(parts))
		for i, p := range parts {
			out[i] = fmt.Sprintf("%v|%v|%s|%.9f|%s", p.B, p.A, p.Relation, p.Similarity, p.Cause)
		}
		slices.Sort(out)
		return out
	}
	same := func(parts []valuecompare.PartMatch) []string {
		out := make([]string, len(parts))
		for i, p := range parts {
			out[i] = fmt.Sprintf("%v|%v|%s|%.9f|%s", p.A, p.B, p.Relation, p.Similarity, p.Cause)
		}
		slices.Sort(out)
		return out
	}
	sortedParts := func(parts []valuecompare.Part) []string {
		out := make([]string, len(parts))
		for i, p := range parts {
			out[i] = fmt.Sprintf("%v", p)
		}
		slices.Sort(out)
		return out
	}
	return slices.Equal(flip(a.Parts), same(b.Parts)) &&
		slices.Equal(sortedParts(a.OnlyA), sortedParts(b.OnlyB)) &&
		slices.Equal(sortedParts(a.OnlyB), sortedParts(b.OnlyA))
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
					// Names keep the same pairs from both sides; text alignment is not mirrored.
					if swappedDet.Cause != det.Cause {
						t.Fatalf("%s: swapped cause %q, want %q", label(), swappedDet.Cause, det.Cause)
					}
					if typ == propertymeta.ValueTypeName && !mirrored(det, swappedDet) {
						t.Fatalf("%s: swapped parts %+v, want the mirror of %+v", label(), swappedDet, det)
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
