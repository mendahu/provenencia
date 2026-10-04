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
)

// giv is the given-role score from the first-name similarity and the
// whole-set Dice overlap; str blends surname and given at the default 60 / 40.
func giv(first, dice float64) float64 { return 0.7*first + 0.3*dice }
func str(surname, g float64) float64  { return 0.6*surname + 0.4*g }

type nameCase struct {
	group, name string
	cmp         NameComparer
	a, b        Value
	want        float64
}

// nameCases pin every rule to an exact score. When a default is retuned,
// update the helpers or the expressions here, not the expectations by hand.
var nameCases = []nameCase{
	// Identity and normalization: case, punctuation, spacing, part order
	// across roles, and the form never matter when parts are typed.
	{"identity", "same parts", NameComparer{}, nv("g=James|s=Robins"), nv("g=James|s=Robins"), 1},
	{"identity", "case and punctuation", NameComparer{}, nv("g=JAMES|s=Robins."), nv("g=james|s=robins"), 1},
	{"identity", "spacing inside a part", NameComparer{}, nv("g=James|s= Silva   Costa "), nv("g=James|s=Silva Costa"), 1},
	{"identity", "surname listed first", NameComparer{}, nv("s=Robins|g=James"), nv("g=James|s=Robins"), 1},
	{"identity", "forms disagree, parts agree", NameComparer{}, withForm(nv("g=James|s=Robins"), "ROBINS, Jas."), withForm(nv("g=James|s=Robins"), "Jim Robbins"), 1},
	{"identity", "forms agree, parts disagree", NameComparer{}, withForm(nv("g=Mary|s=Robins"), "James Robins"), withForm(nv("g=James|s=Robins"), "James Robins"), str(1, 0)},
	{"identity", "initial with and without period", NameComparer{}, nv("i=J.|s=Robins"), nv("i=J|s=Robins"), 1},
	{"identity", "non-Latin script", NameComparer{}, nv("s=蒋|g=浩"), nv("g=浩|s=蒋"), 1},
	{"identity", "accented letters match themselves", NameComparer{}, nv("g=José|s=Núñez"), nv("g=josé|s=NÚÑEZ"), 1},

	// Surname role.
	{"surname", "spelling variant", NameComparer{}, nv("g=James|s=Robbins"), nv("g=James|s=Robins"), str(robbins, 1)},
	{"surname", "spelling variant at the floor", NameComparer{}, nv("g=John|s=Smyth"), nv("g=John|s=Smith"), str(smyth, 1)},
	{"surname", "unrelated", NameComparer{}, nv("g=James|s=Lovelace"), nv("g=James|s=Robins"), str(0, 0.5)},
	{"surname", "dual surname vs one of them", NameComparer{}, nv("g=Maria|s=Silva Costa"), nv("g=Maria|s=Costa"), str(2.0/3, 1)},
	{"surname", "dual surnames as separate parts", NameComparer{}, nv("g=Maria|s=Silva|s=Costa"), nv("g=Maria|s=Silva Costa"), 1},
	{"surname", "dual surnames reversed", NameComparer{}, nv("g=Maria|s=Costa Silva"), nv("g=Maria|s=Silva Costa"), 1},
	{"surname", "dual surnames share one", NameComparer{}, nv("g=Maria|s=Silva Costa"), nv("g=Maria|s=Costa Pereira"), str(0.5, 1)},
	{"surname", "particle on one side only", NameComparer{}, nv("g=Vincent|sp=van|s=Gogh"), nv("g=Vincent|s=Gogh"), 1},
	{"surname", "different particles", NameComparer{}, nv("g=Ludwig|sp=van|s=Beethoven"), nv("g=Ludwig|sp=von|s=Beethoven"), 1},
	{"surname", "missing on one side", NameComparer{}, nv("g=James"), nv("g=James|s=Robins"), str(0.5, 1)},
	{"surname", "missing on both sides", NameComparer{}, nv("g=James"), nv("g=James"), 1},
	{"surname", "missing on both, given differs", NameComparer{}, nv("g=James"), nv("g=Mary"), 0},
	{"surname", "single-character surname", NameComparer{}, nv("s=王|g=芳"), nv("s=王|g=伟"), str(1, 0)},
	{"given", "single-character given is a word, not an initial", NameComparer{}, nv("s=蒋|g=浩"), nv("s=蒋|g=浩然"), str(1, 0)},
	{"given", "given only on both sides, initial", NameComparer{}, nv("i=J."), nv("g=James"), giv(0.5, 0.5)},

	// Given role: the first given name leads; the rest refine.
	{"given", "different first names", NameComparer{}, nv("g=Mary|s=Robins"), nv("g=James|s=Robins"), str(1, 0)},
	{"given", "middle name on one side", NameComparer{}, nv("g=James|g=William|s=Robins"), nv("g=James|s=Robins"), str(1, giv(1, 2.0/3))},
	{"given", "names in one part", NameComparer{}, nv("g=James William|s=Robins"), nv("g=James|g=William|s=Robins"), 1},
	{"given", "first and middle swapped", NameComparer{}, nv("g=William|g=James|s=Robins"), nv("g=James|g=William|s=Robins"), str(1, giv(0, 1))},
	{"given", "middle name used as first", NameComparer{}, nv("g=William|s=Robins"), nv("g=James|g=William|s=Robins"), str(1, giv(0, 2.0/3))},
	{"given", "initial for the first name", NameComparer{}, nv("i=J.|s=Robins"), nv("g=James|s=Robins"), str(1, giv(0.5, 0.5))},
	{"given", "initial for the middle name", NameComparer{}, nv("g=James|i=W.|s=Robins"), nv("g=James|g=William|s=Robins"), str(1, giv(1, 0.75))},
	{"given", "initials on both sides", NameComparer{}, nv("i=J.|i=W.|s=Robins"), nv("i=J|i=W|s=Robins"), 1},
	{"given", "initials vs full names", NameComparer{}, nv("i=J.|i=W.|s=Robins"), nv("g=James|g=William|s=Robins"), str(1, giv(0.5, 0.5))},
	{"given", "initial that does not fit", NameComparer{}, nv("i=W.|s=Robins"), nv("g=James|s=Robins"), str(1, 0)},
	{"given", "first-name spelling variant", NameComparer{}, nv("g=Johann|s=Bach"), nv("g=Johan|s=Bach"), str(1, 5.0/6)},
	{"given", "missing on one side", NameComparer{}, nv("s=Robins"), nv("g=James|s=Robins"), str(1, 0.5)},
	{"given", "nickname on one side matches given", NameComparer{}, nv("g=James|n=Jim|s=Robins"), nv("g=Jim|s=Robins"), str(1, giv(1, 0))},
	{"given", "nicknames on both sides", NameComparer{}, nv("g=James|n=Jim|s=Robins"), nv("g=Jacob|n=Jim|s=Robins"), str(1, giv(1, 0))},
	{"given", "nickname does not fit", NameComparer{}, nv("g=James|n=Jim|s=Robins"), nv("g=Mary|s=Robins"), str(1, 0)},
	{"given", "nickname alone is not structured", NameComparer{}, withForm(nv("n=Jim"), "Jim"), withForm(nv("g=Jim|s=Robins"), "Jim Robins"), 0.8 * 2 * 1 / 3},

	// A shared given name with unrelated surnames is weak.
	{"surname conflict", "same given, unrelated surname", NameComparer{}, nv("g=James|s=Smith"), nv("g=James|s=Robins"), str(0, 0.5)},
	{"surname conflict", "initial only, unrelated surname", NameComparer{}, nv("i=J.|s=Smith"), nv("g=James|s=Robins"), str(0, giv(0.5, 0.5)*0.5)},
	{"surname conflict", "factor off", NameComparer{GivenOnlyFactor: 1}, nv("g=James|s=Smith"), nv("g=James|s=Robins"), str(0, 1)},
	{"surname conflict", "variant is not a conflict", NameComparer{}, nv("g=James|s=Robbins"), nv("g=James|s=Robins"), str(robbins, 1)},

	// Suffixes tell generations apart.
	{"suffix", "same suffix", NameComparer{}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins|x=jr"), 1},
	{"suffix", "different suffixes", NameComparer{}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins|x=Sr."), 0.3},
	{"suffix", "different numerals", NameComparer{}, nv("g=James|s=Robins|x=III"), nv("g=James|s=Robins|x=II"), 0.3},
	{"suffix", "suffix on one side only", NameComparer{}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins"), 1},
	{"suffix", "conflict on a partial match", NameComparer{}, nv("g=Mary|s=Robins|x=Jr."), nv("g=James|s=Robins|x=Sr."), str(1, 0) * 0.3},
	{"suffix", "penalty off", NameComparer{SuffixConflict: 1}, nv("g=James|s=Robins|x=Jr."), nv("g=James|s=Robins|x=Sr."), 1},

	// Titles and untyped parts carry no role.
	{"no role", "different titles", NameComparer{}, nv("p=Rev.|g=James|s=Robins"), nv("p=Dr.|g=James|s=Robins"), 1},
	{"no role", "undetermined part ignored", NameComparer{}, nv("g=James|u=Kendall|s=Robins"), nv("g=James|s=Robins"), 1},
	{"no role", "untyped part ignored", NameComparer{}, nv("g=James|_=Kendall|s=Robins"), nv("g=James|s=Robins"), 1},

	// The form fallback, when either side has no typed surname or given part.
	{"fallback", "both form-only, same", NameComparer{}, formOnly("James Robins"), formOnly("james  robins."), 1},
	{"fallback", "both form-only, shared surname", NameComparer{}, formOnly("Mary Robins"), formOnly("James Robins"), 0.8 * 0.5},
	{"fallback", "structured vs form-only", NameComparer{}, withForm(nv("g=James|s=Robins"), "James Robins"), formOnly("James Robins"), 1},
	{"fallback", "only untyped parts", NameComparer{}, withForm(nv("_=Mary|u=Robins"), "Mary Robins"), formOnly("James Robins"), 0.8 * 0.5},
	{"fallback", "only a title and suffix", NameComparer{}, withForm(nv("p=Rev.|x=Jr."), "Rev. Robins Jr."), formOnly("Robins"), 0.8 * 2 * 1 / 4},
	{"fallback", "initial in the form", NameComparer{}, formOnly("J. Robins"), formOnly("James Robins"), 0.8 * 2 * 1.5 / 4},
	{"fallback", "custom partial", NameComparer{Partial: 0.5}, formOnly("Mary Robins"), formOnly("James Robins"), 0.5 * 0.5},

	// Tuning knobs.
	{"tuning", "surname share 0.5", NameComparer{SurnameShare: 0.5}, nv("g=Mary|s=Robins"), nv("g=James|s=Robins"), 0.5},
	{"tuning", "surname only", NameComparer{SurnameShare: 1}, nv("g=Mary|s=Robins"), nv("g=James|s=Robins"), 1},
	{"tuning", "fuzzy off", NameComparer{FuzzyFloor: 1}, nv("g=James|s=Robbins"), nv("g=James|s=Robins"), str(0, 0.5)},
	{"tuning", "looser fuzzy floor", NameComparer{FuzzyFloor: 0.7}, nv("g=Mary|s=Robins"), nv("g=Mark|s=Robins"), str(1, 0.75)},
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
			"g=James|s=Robins",
			"g=James|i=K.|s=Robbins",
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
	}
	for _, l := range ladders {
		probe := nv(l.probe)
		prev := math.Inf(1)
		for i, spec := range l.order {
			got, ok := (NameComparer{}).Compare(probe, nv(spec))
			if !ok {
				t.Fatalf("%s vs %s: not comparable", l.probe, spec)
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
	cmps := []NameComparer{{}, {SurnameShare: 0.3, GivenOnlyFactor: 1, SuffixConflict: 0.9}, {FuzzyFloor: 1}}

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

			if rolesOf(a.Name).structured() && rolesOf(b.Name).structured() {
				// The form is irrelevant once both sides are typed.
				if again, _ := cmp.Compare(withForm(a, "anything else"), withForm(b, "")); math.Abs(again-got) > 1e-9 {
					t.Fatalf("%s: form changed a structured score: %.4f vs %.4f", where, got, again)
				}
				// Moving surname parts to the front keeps the score.
				if again, _ := cmp.Compare(surnamesFirst(a), b); math.Abs(again-got) > 1e-9 {
					t.Fatalf("%s: part order across roles changed the score: %.4f vs %.4f", where, got, again)
				}
				// Giving two suffix-less names different suffixes never raises it.
				if len(rolesOf(a.Name).suffixes) > 0 || len(rolesOf(b.Name).suffixes) > 0 {
					// skip: an added suffix can match an existing one
				} else if again, _ := cmp.Compare(addPart(a, "x", "Jr."), addPart(b, "x", "Sr.")); again > got+1e-9 {
					t.Fatalf("%s: a suffix conflict raised the score: %.4f vs %.4f", where, got, again)
				}
				// Titles and untyped parts never move it.
				if again, _ := cmp.Compare(addPart(a, "p", "Rev."), addPart(b, "u", "Kendall")); math.Abs(again-got) > 1e-9 {
					t.Fatalf("%s: a role-less part moved the score: %.4f vs %.4f", where, got, again)
				}
			}
		}
	}
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
