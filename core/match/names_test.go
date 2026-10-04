package match

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/namevalues"
)

// nv builds a name from "type=value" parts separated by "|", with a form
// that says nothing, so structured comparison has to carry the test.
//
//	g given · i initial · n nick · s surname · sp surname_prefix ·
//	x suffix · p prefix · u undetermined · _ untyped
func nv(spec string) Value {
	n := &namevalues.Value{Form: "(as written)"}
	for i, item := range strings.Split(spec, "|") {
		code, val, ok := strings.Cut(item, "=")
		if !ok {
			panic("bad part " + item)
		}
		typ, known := partCodes[code]
		if !known {
			panic("bad part code " + code)
		}
		n.Parts = append(n.Parts, namevalues.Part{Idx: i, Type: typ, Value: val})
	}
	return Value{Name: n}
}

var partCodes = map[string]string{
	"g": namevalues.PartTypeGiven, "i": namevalues.PartTypeInitial, "n": namevalues.PartTypeNick,
	"s": namevalues.PartTypeSurname, "sp": namevalues.PartTypeSurnamePrefix, "x": namevalues.PartTypeSuffix,
	"p": namevalues.PartTypePrefix, "u": namevalues.PartTypeUndetermined, "_": "",
}

// formOnly is a name with no parts.
func formOnly(form string) Value { return Value{Name: &namevalues.Value{Form: form}} }

// withForm replaces a name's form.
func withForm(v Value, form string) Value {
	n := *v.Name
	n.Form = form
	return Value{Name: &n}
}

// Edit-distance ratios used below.
const (
	robbins = 6.0 / 7 // robins ~ robbins
	smyth   = 4.0 / 5 // smith ~ smyth
	johan   = 5.0 / 6 // johann ~ johan
	robnis  = 5.0 / 6 // robnis ~ robins: one adjacent swap
	anne    = 3.0 / 4 // ann ~ anne: one added letter in a short word
)

// Default word weights: family F, first given G1, other given G2, nick N,
// untyped U. A pair earns similarity × affinity × (wa + wb); a score is the
// earned share of both names' total weight. Expectations below are written
// as that arithmetic.
const (
	F  = 1.5
	G1 = 1.0
	G2 = 0.5
	N  = 0.3
	U  = 1.0
)

type nameCase struct {
	group, name string
	cmp         NameComparer
	a, b        Value
	want        float64
}

// nameCases pin every rule to an exact score.
var nameCases = []nameCase{
	// Identity and normalization.
	{"identity", "same parts", NameComparer{}, nv("g=James|s=Robins"), nv("g=James|s=Robins"), 1},
	{"identity", "case and punctuation", NameComparer{}, nv("g=JAMES|s=Robins."), nv("g=james|s=robins"), 1},
	{"identity", "spacing inside a part", NameComparer{}, nv("g=James|s= Silva   Costa "), nv("g=James|s=Silva Costa"), 1},
	{"identity", "surname listed first", NameComparer{}, nv("s=Robins|g=James"), nv("g=James|s=Robins"), 1},
	{"identity", "forms disagree, parts agree", NameComparer{}, withForm(nv("g=James|s=Robins"), "ROBINS, Jas."), withForm(nv("g=James|s=Robins"), "Jim Robbins"), 1},
	{"identity", "forms agree, parts disagree", NameComparer{}, withForm(nv("g=Mary|s=Robins"), "James Robins"), withForm(nv("g=James|s=Robins"), "James Robins"), 2 * F / (2 * (G1 + F))},
	{"identity", "initial with and without period", NameComparer{}, nv("i=J.|s=Robins"), nv("i=J|s=Robins"), 1},
	{"identity", "non-Latin script", NameComparer{}, nv("s=蒋|g=浩"), nv("g=浩|s=蒋"), 1},
	{"identity", "accented letters match themselves", NameComparer{}, nv("g=José|s=Núñez"), nv("g=josé|s=NÚÑEZ"), 1},

	// Family words.
	{"surname", "spelling variant", NameComparer{}, nv("g=James|s=Robbins"), nv("g=James|s=Robins"), (2*G1 + robbins*2*F) / (2 * (G1 + F))},
	{"surname", "spelling variant at the floor", NameComparer{}, nv("g=John|s=Smyth"), nv("g=John|s=Smith"), (2*G1 + smyth*2*F) / (2 * (G1 + F))},
	{"surname", "unrelated", NameComparer{}, nv("g=James|s=Lovelace"), nv("g=James|s=Robins"), 2 * G1 / (2 * (G1 + F)) * 0.5},
	{"surname", "dual surname vs one of them", NameComparer{}, nv("g=Maria|s=Silva Costa"), nv("g=Maria|s=Costa"), (2*G1 + 2*F) / (G1 + 2*F + G1 + F)},
	{"surname", "dual surnames as separate parts", NameComparer{}, nv("g=Maria|s=Silva|s=Costa"), nv("g=Maria|s=Silva Costa"), 1},
	{"surname", "dual surnames reversed", NameComparer{}, nv("g=Maria|s=Costa Silva"), nv("g=Maria|s=Silva Costa"), 1},
	{"surname", "dual surnames share one", NameComparer{}, nv("g=Maria|s=Silva Costa"), nv("g=Maria|s=Costa Pereira"), (2*G1 + 2*F) / (2 * (G1 + 2*F))},
	{"surname", "hyphenated vs spaced", NameComparer{}, nv("g=Ann|s=Smith-Jones"), nv("g=Ann|s=Smith Jones"), 1},
	{"surname", "hyphenated vs one half", NameComparer{}, nv("g=Ann|s=Smith-Jones"), nv("g=Ann|s=Jones"), (2*G1 + 2*F) / (G1 + 2*F + G1 + F)},
	{"surname", "en dash joins like a hyphen", NameComparer{}, nv("g=Ann|s=Smith–Jones"), nv("g=Ann|s=Smith-Jones"), 1},
	{"surname", "apostrophe joins", NameComparer{}, nv("g=Pat|s=O'Brien"), nv("g=Pat|s=OBrien"), 1},
	{"surname", "adjacent letters swapped", NameComparer{}, nv("g=Ann|s=Robnis"), nv("g=Ann|s=Robins"), (2*G1 + robnis*2*F) / (2 * (G1 + F))},
	{"surname", "particle on one side only", NameComparer{}, nv("g=Vincent|sp=van|s=Gogh"), nv("g=Vincent|s=Gogh"), 1},
	{"surname", "different particles", NameComparer{}, nv("g=Ludwig|sp=van|s=Beethoven"), nv("g=Ludwig|sp=von|s=Beethoven"), 1},
	{"surname", "missing on one side", NameComparer{}, nv("g=James"), nv("g=James|s=Robins"), 2 * G1 / (G1 + G1 + F)},
	{"surname", "missing on both sides", NameComparer{}, nv("g=James"), nv("g=James"), 1},
	{"surname", "missing on both, given differs", NameComparer{}, nv("g=James"), nv("g=Mary"), 0},
	{"surname", "single-character surname", NameComparer{}, nv("s=王|g=芳"), nv("s=王|g=伟"), 2 * F / (2 * (G1 + F))},

	// Given words: the first given name weighs most; middle names and
	// initials less.
	{"given", "different first names", NameComparer{}, nv("g=Mary|s=Robins"), nv("g=James|s=Robins"), 2 * F / (2 * (G1 + F))},
	{"given", "middle name on one side", NameComparer{}, nv("g=James|g=William|s=Robins"), nv("g=James|s=Robins"), (2*G1 + 2*F) / (G1 + G2 + F + G1 + F)},
	{"given", "names in one part", NameComparer{}, nv("g=James William|s=Robins"), nv("g=James|g=William|s=Robins"), 1},
	{"given", "first and middle swapped still pair", NameComparer{}, nv("g=William|g=James|s=Robins"), nv("g=James|g=William|s=Robins"), 1},
	{"given", "middle name used as first", NameComparer{}, nv("g=William|s=Robins"), nv("g=James|g=William|s=Robins"), ((G1 + G2) + 2*F) / (G1 + F + G1 + G2 + F)},
	{"given", "initial for the first name", NameComparer{}, nv("i=J.|s=Robins"), nv("g=James|s=Robins"), (0.5*2*G1 + 2*F) / (2 * (G1 + F))},
	{"given", "initial for the middle name", NameComparer{}, nv("g=James|i=W.|s=Robins"), nv("g=James|g=William|s=Robins"), (2*G1 + 0.5*2*G2 + 2*F) / (2 * (G1 + G2 + F))},
	{"given", "initials on both sides", NameComparer{}, nv("i=J.|i=W.|s=Robins"), nv("i=J|i=W|s=Robins"), 1},
	{"given", "initials vs full names", NameComparer{}, nv("i=J.|i=W.|s=Robins"), nv("g=James|g=William|s=Robins"), (0.5*2*G1 + 0.5*2*G2 + 2*F) / (2 * (G1 + G2 + F))},
	{"given", "initial that does not fit", NameComparer{}, nv("i=W.|s=Robins"), nv("g=James|s=Robins"), 2 * F / (2 * (G1 + F))},
	{"given", "first-name spelling variant", NameComparer{}, nv("g=Johann|s=Bach"), nv("g=Johan|s=Bach"), (johan*2*G1 + 2*F) / (2 * (G1 + F))},
	{"given", "short name, one added letter", NameComparer{}, nv("g=Ann|s=Robins"), nv("g=Anne|s=Robins"), (anne*2*G1 + 2*F) / (2 * (G1 + F))},
	{"given", "short name, one dropped letter", NameComparer{}, nv("g=John|s=Robins"), nv("g=Jon|s=Robins"), (anne*2*G1 + 2*F) / (2 * (G1 + F))},
	{"given", "short name, one substituted letter", NameComparer{}, nv("g=Jane|s=Robins"), nv("g=June|s=Robins"), 2 * F / (2 * (G1 + F))},
	{"given", "two-letter name, one added letter", NameComparer{}, nv("g=Al|s=Robins"), nv("g=Ali|s=Robins"), 2 * F / (2 * (G1 + F))},
	{"given", "abbreviation is not yet a match", NameComparer{}, nv("g=Jas.|s=Robins"), nv("g=James|s=Robins"), 2 * F / (2 * (G1 + F))},
	{"given", "missing on one side", NameComparer{}, nv("s=Robins"), nv("g=James|s=Robins"), 2 * F / (F + G1 + F)},
	{"given", "given only, initial", NameComparer{}, nv("i=J."), nv("g=James"), 0.5},
	{"given", "single character is a word, not an initial", NameComparer{}, nv("s=蒋|g=浩"), nv("s=蒋|g=浩然"), 2 * F / (2 * (G1 + F))},
	{"given", "nickname matches a given name", NameComparer{}, nv("g=James|n=Jim|s=Robins"), nv("g=Jim|s=Robins"), (0.9*(N+G1) + 2*F) / (G1 + N + F + G1 + F)},
	{"given", "nicknames on both sides", NameComparer{}, nv("g=James|n=Jim|s=Robins"), nv("g=Jacob|n=Jim|s=Robins"), (2*N + 2*F) / (2 * (G1 + N + F))},
	{"given", "nickname does not fit", NameComparer{}, nv("g=James|n=Jim|s=Robins"), nv("g=Mary|s=Robins"), 2 * F / (G1 + N + F + G1 + F)},
	{"given", "nickname alone", NameComparer{}, nv("n=Jim"), nv("g=Jim|s=Robins"), 0.9 * (N + G1) / (N + G1 + F)},

	// Types are data: words of different types still pair, discounted.
	{"cross-format", "surname and given swapped", NameComparer{}, nv("g=Robins|s=James"), nv("g=James|s=Robins"), (0.5*(G1+F) + 0.5*(F+G1)) / (2 * (G1 + F))},
	{"cross-format", "typed vs form-only", NameComparer{}, withForm(nv("g=James|s=Robins"), "x"), formOnly("James Robins"), (0.8*(G1+U) + 0.8*(F+U)) / (G1 + F + 2*U)},
	{"cross-format", "typed vs untyped parts", NameComparer{}, nv("g=James|s=Robins"), nv("_=James|u=Robins"), (0.8*(G1+U) + 0.8*(F+U)) / (G1 + F + 2*U)},
	{"cross-format", "an extra untyped word", NameComparer{}, nv("g=James|u=Kendall|s=Robins"), nv("g=James|s=Robins"), (2*G1 + 2*F) / (G1 + U + F + G1 + F)},
	{"cross-format", "surname-first entry typed by position", NameComparer{}, nv("g=Wang|s=Fang"), nv("s=Wang|g=Fang"), 0.5},
	{"cross-format", "a profile's own roles", NameComparer{PartRoles: map[string]NameRole{"given": RoleGiven, "undetermined": RoleFamily}}, nv("g=Anders|u=Jonsson"), nv("g=Anders|u=Jonsson"), 1},
	{"cross-format", "a profile's role for an unknown type", NameComparer{PartRoles: map[string]NameRole{"given": RoleGiven, "undetermined": RoleFamily}}, nv("g=Anders|u=Jonsson"), nv("g=Anders|s=Jonsson"), (2*G1 + 0.8*(F+U)) / (G1 + F + G1 + U)},

	// A shared given name with unrelated surnames is weak.
	{"surname conflict", "same given, unrelated surname", NameComparer{}, nv("g=James|s=Smith"), nv("g=James|s=Robins"), 2 * G1 / (2 * (G1 + F)) * 0.5},
	{"surname conflict", "initial only, unrelated surname", NameComparer{}, nv("i=J.|s=Smith"), nv("g=James|s=Robins"), 0.5 * 2 * G1 / (2 * (G1 + F)) * 0.5},
	{"surname conflict", "an initial does not hide it", NameComparer{}, nv("i=J.|s=Smith"), nv("i=S.|s=Robins"), 0.5 * 0.5 * (F + G1) / (2 * (G1 + F)) * 0.5},
	{"surname conflict", "factor zero", NameComparer{GivenOnlyFactor: Set(0.0)}, nv("g=James|s=Smith"), nv("g=James|s=Robins"), 0},
	{"surname conflict", "factor off", NameComparer{GivenOnlyFactor: Set(1.0)}, nv("g=James|s=Smith"), nv("g=James|s=Robins"), 2 * G1 / (2 * (G1 + F))},
	{"surname conflict", "variant is not a conflict", NameComparer{}, nv("g=James|s=Robbins"), nv("g=James|s=Robins"), (2*G1 + robbins*2*F) / (2 * (G1 + F))},
	{"surname conflict", "swapped types are not a conflict", NameComparer{}, nv("g=Robins|s=James"), nv("g=James|s=Robins"), 0.5},

	// Suffixes tell generations apart.
	{"suffix", "same suffix", NameComparer{}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins|x=jr"), 1},
	{"suffix", "different suffixes", NameComparer{}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins|x=Sr."), 0.3},
	{"suffix", "different numerals", NameComparer{}, nv("g=James|s=Robins|x=III"), nv("g=James|s=Robins|x=II"), 0.3},
	{"suffix", "suffix on one side only", NameComparer{}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins"), 1},
	{"suffix", "conflict on a partial match", NameComparer{}, nv("g=Mary|s=Robins|x=Jr."), nv("g=James|s=Robins|x=Sr."), 2 * F / (2 * (G1 + F)) * 0.3},
	{"suffix", "penalty off", NameComparer{SuffixConflict: Set(1.0)}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins|x=Sr."), 1},

	// Titles and surname particles carry no weight.
	{"ignored", "different titles", NameComparer{}, nv("p=Rev.|g=James|s=Robins"), nv("p=Dr.|g=James|s=Robins"), 1},

	// Names with no typed parts are read from their form as untyped words.
	{"form", "spelling variant", NameComparer{}, formOnly("James Robbins"), formOnly("James Robins"), (2*U + robbins*2*U) / (4 * U)},
	{"form", "not a variant", NameComparer{}, formOnly("Mary"), formOnly("Mark"), 0},
	{"form", "fuzzy off", NameComparer{FuzzyFloor: Set(1.0)}, formOnly("Robbins"), formOnly("Robins"), 0},
	{"form", "both form-only, same", NameComparer{}, formOnly("James Robins"), formOnly("james  robins."), 1},
	{"form", "both form-only, reordered", NameComparer{}, formOnly("Robins, James"), formOnly("James Robins"), 1},
	{"form", "both form-only, shared surname", NameComparer{}, formOnly("Mary Robins"), formOnly("James Robins"), 0.5},
	{"form", "initial in the form", NameComparer{}, formOnly("J. Robins"), formOnly("James Robins"), (0.5*2*U + 2*U) / (4 * U)},
	{"form", "only a title and suffix parts", NameComparer{}, withForm(nv("p=Rev.|x=Jr."), "Rev. Robins Jr."), formOnly("Robins"), 2 * U / (4 * U)},

	// Tuning knobs.
	{"tuning", "equal weights", NameComparer{FamilyWeight: Set(1.0)}, nv("g=Mary|s=Robins"), nv("g=James|s=Robins"), 0.5},
	{"tuning", "cross-role at full", NameComparer{CrossRole: Set(1.0)}, nv("g=Robins|s=James"), nv("g=James|s=Robins"), 1},
	{"tuning", "cross-role zero: types are strict", NameComparer{CrossRole: Set(0.0)}, nv("g=Robins|s=James"), nv("g=James|s=Robins"), 0},
	{"tuning", "fuzzy off also stops short-word variants", NameComparer{FuzzyFloor: Set(1.0)}, nv("g=Ann|s=Robins"), nv("g=Anne|s=Robins"), 2 * F / (2 * (G1 + F))},
	{"tuning", "untyped at full", NameComparer{UntypedAffinity: Set(1.0)}, withForm(nv("g=James|s=Robins"), "x"), formOnly("James Robins"), 1},
	{"tuning", "fuzzy off", NameComparer{FuzzyFloor: Set(1.0)}, nv("g=James|s=Robbins"), nv("g=James|s=Robins"), 2 * G1 / (2 * (G1 + F)) * 0.5},
	{"tuning", "looser fuzzy floor", NameComparer{FuzzyFloor: Set(0.7)}, nv("g=Mary|s=Robins"), nv("g=Mark|s=Robins"), (0.75*2*G1 + 2*F) / (2 * (G1 + F))},
}

func TestNameComparerScores(t *testing.T) {
	for _, tt := range nameCases {
		t.Run(tt.group+"/"+tt.name, func(t *testing.T) {
			got, ok := tt.cmp.Compare(tt.a, tt.b)
			if !ok {
				t.Fatal("not comparable")
			}
			if math.Abs(got-tt.want) > 1e-9 {
				t.Fatalf("got %.4f, want %.4f", got, tt.want)
			}
			if back, _ := tt.cmp.Compare(tt.b, tt.a); math.Abs(back-got) > 1e-9 {
				t.Fatalf("not symmetric: %.4f then %.4f", got, back)
			}
		})
	}
}

func TestNameComparerNotComparable(t *testing.T) {
	for name, pair := range map[string][2]Value{
		"no name":           {Value{}, nv("g=James")},
		"empty form":        {formOnly(""), formOnly("James")},
		"punctuation form":  {formOnly("—"), formOnly("James")},
		"blank parts, form": {withForm(nv("g= |s=."), ""), formOnly("James")},
	} {
		if s, ok := (NameComparer{}).Compare(pair[0], pair[1]); ok {
			t.Errorf("%s: comparable (%.3f)", name, s)
		}
	}
}

// Ladders assert order rather than exact scores, so they keep their meaning
// when weights are retuned: each name must resemble the probe strictly more
// than the next.
func TestNameComparerLadders(t *testing.T) {
	ladders := []struct {
		probe string
		order []string
	}{
		{"g=James|i=K.|s=Robins", []string{
			"g=James|i=K.|s=Robins",
			"g=James|i=K.|s=Robbins",
			"g=James|s=Robins",
			"i=J.|i=K.|s=Robins",
			"s=Robins",
			"g=Mary|s=Robins",
			"g=James|i=K.|s=Smith",
			"g=Ada|s=Lovelace",
		}},
		{"g=James|s=Robins|x=Jr.", []string{
			"g=James|s=Robins|x=Jr.",
			"g=Jim|n=James|s=Robbins",
			"g=Mary|s=Robins",
			"g=James|s=Robins|x=Sr.",
			"g=Mary|s=Robins|x=Sr.",
		}},
		{"g=Maria|g=Luisa|s=Silva Costa", []string{
			"g=Maria|g=Luisa|s=Silva|s=Costa",
			"g=Maria|s=Silva Costa",
			"g=Maria|g=Luisa|s=Costa",
			"i=M.|s=Costa",
			"g=Luisa|s=Pereira",
		}},
		// Across formats: the same words typed differently still rank above
		// a different person typed the same way.
		{"g=James|s=Robins", []string{
			"g=James|s=Robins",
			"_=James|_=Robins",
			"s=James|g=Robins",
			"_=Mary|_=Robins",
			"g=Ada|s=Lovelace",
		}},
	}
	for _, l := range ladders {
		probe := nv(l.probe)
		prev := math.Inf(1)
		for i, spec := range l.order {
			got, ok := (NameComparer{}).Compare(probe, nv(spec))
			if !ok {
				t.Fatalf("%s vs %s: not comparable", l.probe, spec)
			}
			if i == len(l.order)-1 && got == 0 && prev > 0 {
				prev = got
				continue
			}
			if !(got < prev) {
				t.Errorf("%s: #%d %s scored %.4f, not below %.4f", l.probe, i, spec, got, prev)
			}
			prev = got
		}
	}
}

// Generated permutations from fixed seeds (deterministic, like the cache's
// rebuild-equals-upkeep sequences): every pair must satisfy the comparer's
// invariants. A failure names its seed and pair; shrink it into
// nameCases.
func TestNameComparerInvariants(t *testing.T) {
	pools := map[string][]string{
		"g":  {"James", "Jim", "Mary", "William", "Johann", "Johan", "José", "Maria Luisa", "浩"},
		"i":  {"J.", "K.", "W", "M."},
		"n":  {"Jim", "Bill", "Polly"},
		"s":  {"Robins", "Robbins", "Smith", "Smyth", "Silva Costa", "Costa", "Gogh", "蒋"},
		"sp": {"van", "von", "de"},
		"x":  {"Jr.", "Sr.", "III"},
		"p":  {"Rev.", "Dr."},
		"u":  {"Kendall"},
		"_":  {"Robins", "James"},
	}
	codes := []string{"g", "g", "i", "n", "s", "s", "sp", "x", "p", "u", "_"}
	forms := []string{"James Robins", "J. Robins", "Mary Smith", "(as written)", ""}
	cmps := []NameComparer{{}, {FamilyWeight: Set(3.0), CrossRole: Set(0.2), GivenOnlyFactor: Set(1.0), SuffixConflict: Set(0.9)}, {FuzzyFloor: Set(1.0)}}

	for _, seed := range []int64{1, 2, 3, 4} {
		rng := rand.New(rand.NewSource(seed))
		gen := func() Value {
			n := &namevalues.Value{Form: forms[rng.Intn(len(forms))]}
			for i, k := 0, rng.Intn(5); i < k; i++ {
				code := codes[rng.Intn(len(codes))]
				pool := pools[code]
				n.Parts = append(n.Parts, namevalues.Part{Idx: i, Type: partCodes[code], Value: pool[rng.Intn(len(pool))]})
			}
			return Value{Name: n}
		}
		for i := 0; i < 2000; i++ {
			a, b := gen(), gen()
			cmp := cmps[rng.Intn(len(cmps))]
			where := fmt.Sprintf("seed %d pair %d: %+v vs %+v (%+v)", seed, i, *a.Name, *b.Name, cmp)

			got, ok := cmp.Compare(a, b)
			back, okBack := cmp.Compare(b, a)
			if ok != okBack || math.Abs(got-back) > 1e-9 {
				t.Fatalf("%s: not symmetric: %.4f %v then %.4f %v", where, got, ok, back, okBack)
			}
			if ok && (got < 0 || got > 1+1e-9 || math.IsNaN(got)) {
				t.Fatalf("%s: out of range %.4f", where, got)
			}
			if self, okSelf := cmp.Compare(a, a); okSelf && math.Abs(self-1) > 1e-9 {
				t.Fatalf("%s: self-similarity %.4f", where, self)
			}

			if partWords(a) && partWords(b) {
				// The form is irrelevant once both sides are typed.
				if again, _ := cmp.Compare(withForm(a, "anything else"), withForm(b, "")); math.Abs(again-got) > 1e-9 {
					t.Fatalf("%s: form changed a structured score: %.4f vs %.4f", where, got, again)
				}
				// Moving surname parts to the front keeps the score.
				if again, _ := cmp.Compare(surnamesFirst(a), b); math.Abs(again-got) > 1e-9 {
					t.Fatalf("%s: part order across roles changed the score: %.4f vs %.4f", where, got, again)
				}
				// Giving two suffix-less names different suffixes never raises it.
				if hasCode(a, "x") || hasCode(b, "x") {
					// skip: an added suffix can match an existing one
				} else if again, _ := cmp.Compare(addPart(a, "x", "Jr."), addPart(b, "x", "Sr.")); again > got+1e-9 {
					t.Fatalf("%s: a suffix conflict raised the score: %.4f vs %.4f", where, got, again)
				}
				// Titles and surname particles never move it.
				if again, _ := cmp.Compare(addPart(a, "p", "Rev."), addPart(b, "sp", "van")); math.Abs(again-got) > 1e-9 {
					t.Fatalf("%s: an ignored part moved the score: %.4f vs %.4f", where, got, again)
				}
				// The same words in another format still connect, never
				// above the name itself.
				if again, ok := cmp.Compare(a, untyped(a)); !ok || again <= 0 || again > 1+1e-9 {
					t.Fatalf("%s: retyped copy scored %.4f %v", where, again, ok)
				}
			}
		}
	}
}

// partWords reports whether a name's parts yield comparable words (so its
// form is not read).
func partWords(v Value) bool {
	words, _ := (NameComparer{}).nameWords(withForm(v, "").Name)
	return len(words) > 0
}

func hasCode(v Value, code string) bool {
	for _, p := range v.Name.Parts {
		if p.Type == partCodes[code] {
			return true
		}
	}
	return false
}

// untyped clears every comparable part's type, as if entered with no format.
func untyped(v Value) Value {
	n := *v.Name
	n.Parts = nil
	for _, p := range v.Name.Parts {
		if r := WesternPartRoles[p.Type]; r != RoleIgnored && r != RoleGeneration {
			p.Type = ""
		}
		n.Parts = append(n.Parts, p)
	}
	return Value{Name: &n}
}

// surnamesFirst moves surname parts ahead of the others, keeping each role's
// own order.
func surnamesFirst(v Value) Value {
	n := *v.Name
	var front, rest []namevalues.Part
	for _, p := range v.Name.Parts {
		if p.Type == namevalues.PartTypeSurname {
			front = append(front, p)
		} else {
			rest = append(rest, p)
		}
	}
	n.Parts = nil
	for i, p := range append(front, rest...) {
		p.Idx = i
		n.Parts = append(n.Parts, p)
	}
	return Value{Name: &n}
}

func addPart(v Value, code, value string) Value {
	n := *v.Name
	n.Parts = append(append([]namevalues.Part(nil), v.Name.Parts...), namevalues.Part{Idx: len(v.Name.Parts), Type: partCodes[code], Value: value})
	return Value{Name: &n}
}
