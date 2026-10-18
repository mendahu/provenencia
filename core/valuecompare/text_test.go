package valuecompare_test

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func textValue(s string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeText, Text: s}
}

func words(parts []valuecompare.Part) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p.Text)
	}
	return out
}

func pm(a, b string, rel valuecompare.Relation, sim float64, cause valuecompare.Cause) valuecompare.PartMatch {
	return valuecompare.PartMatch{
		A: valuecompare.Part{Kind: valuecompare.PartWord, Text: a}, B: valuecompare.Part{Kind: valuecompare.PartWord, Text: b},
		Relation: rel, Similarity: sim, Cause: cause,
	}
}

func TestTextCompare(t *testing.T) {
	rules := matchregistry.Default().Value
	for _, tc := range []struct {
		name         string
		a, b         string
		want         valuecompare.Relation
		sim          float64
		cause        valuecompare.Cause
		broader      valuecompare.Side
		onlyA, onlyB []string
		parts        []valuecompare.PartMatch
	}{
		{name: "exact", a: "York", b: "York", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseExact, parts: []valuecompare.PartMatch{pm("york", "york", valuecompare.Same, 1, valuecompare.CauseExact)}},
		{name: "case", a: "York", b: "york", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: []valuecompare.PartMatch{pm("york", "york", valuecompare.Same, 1, valuecompare.CauseExact)}},
		{name: "particle", a: "O Brien", b: "O'Brien", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: []valuecompare.PartMatch{pm("obrien", "obrien", valuecompare.Same, 1, valuecompare.CauseExact)}},
		{name: "dash", a: "Sainte-Anne", b: "Sainte Anne", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: []valuecompare.PartMatch{pm("sainte", "sainte", valuecompare.Same, 1, valuecompare.CauseExact), pm("anne", "anne", valuecompare.Same, 1, valuecompare.CauseExact)}},
		{name: "word break dash", a: "god-father", b: "godfather", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: []valuecompare.PartMatch{pm("god father", "godfather", valuecompare.Same, 1, valuecompare.CauseCase)}},
		{name: "word break space", a: "York Town", b: "Yorktown", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: []valuecompare.PartMatch{pm("york town", "yorktown", valuecompare.Same, 1, valuecompare.CauseCase)}},
		{name: "word break and fold", a: "Mont-réal", b: "Montreal", want: valuecompare.Equivalent, sim: 0.95, cause: valuecompare.CauseFold, parts: []valuecompare.PartMatch{pm("mont réal", "montreal", valuecompare.Equivalent, 0.95, valuecompare.CauseFold)}},
		{name: "word break same accents", a: "Mont-réal", b: "montréal", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: []valuecompare.PartMatch{pm("mont réal", "montréal", valuecompare.Same, 1, valuecompare.CauseCase)}},
		{name: "word break equal counts", a: "ab c", b: "a bc", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: []valuecompare.PartMatch{pm("ab c", "a bc", valuecompare.Same, 1, valuecompare.CauseCase)}},
		{name: "hangul composed against decomposed", a: "서울 한국", b: norm.NFD.String("서울 한국"), want: valuecompare.Equivalent, sim: 0.95, cause: valuecompare.CauseFold, parts: []valuecompare.PartMatch{pm("서울", norm.NFD.String("서울"), valuecompare.Equivalent, 0.95, valuecompare.CauseFold), pm("한국", norm.NFD.String("한국"), valuecompare.Equivalent, 0.95, valuecompare.CauseFold)}},
		{name: "fold", a: "Montréal", b: "Montreal", want: valuecompare.Equivalent, sim: 0.95, cause: valuecompare.CauseFold, parts: []valuecompare.PartMatch{pm("montréal", "montreal", valuecompare.Equivalent, 0.95, valuecompare.CauseFold)}},
		{
			name: "contains", a: "York", b: "York, Upper Canada",
			want: valuecompare.Contains, sim: 0.7 * (1 + 1) / (1 + 3), cause: valuecompare.CauseMissingPart,
			broader: valuecompare.SideA, onlyB: []string{"upper", "canada"}, parts: []valuecompare.PartMatch{pm("york", "york", valuecompare.Same, 1, valuecompare.CauseExact)},
		},
		{
			name: "contains from the other side", a: "York, Upper Canada", b: "York",
			want: valuecompare.Contains, sim: 0.7 * (1 + 1) / (1 + 3), cause: valuecompare.CauseMissingPart,
			broader: valuecompare.SideB, onlyA: []string{"upper", "canada"}, parts: []valuecompare.PartMatch{pm("york", "york", valuecompare.Same, 1, valuecompare.CauseExact)},
		},
		{
			name: "overlaps", a: "York, Ontario", b: "Toronto, York",
			want: valuecompare.Overlaps, sim: 0.7 * (1 + 1) / (2 + 2), cause: valuecompare.CauseOverlap,
			onlyA: []string{"ontario"}, onlyB: []string{"toronto"}, parts: []valuecompare.PartMatch{pm("york", "york", valuecompare.Same, 1, valuecompare.CauseExact)},
		},
		{name: "order", a: "Upper Canada, York", b: "York, Upper Canada", want: valuecompare.Different, sim: 0, cause: valuecompare.CauseOrder, onlyA: []string{"york"}, onlyB: []string{"york"}, parts: []valuecompare.PartMatch{pm("upper", "upper", valuecompare.Same, 1, valuecompare.CauseExact), pm("canada", "canada", valuecompare.Same, 1, valuecompare.CauseExact)}},
		{name: "spelling one word", a: "Provenance", b: "Provenanced", want: valuecompare.Near, sim: 1 - 1.0/11, cause: valuecompare.CauseSpelling, parts: []valuecompare.PartMatch{pm("provenance", "provenanced", valuecompare.Near, 1-1.0/11, valuecompare.CauseSpelling)}},
		{name: "spelling two words", a: "Upper Canada", b: "Upper Kanada", want: valuecompare.Near, sim: (1 + 1 + 5.0/6 + 5.0/6) / 4, cause: valuecompare.CauseSpelling, parts: []valuecompare.PartMatch{pm("upper", "upper", valuecompare.Same, 1, valuecompare.CauseExact), pm("canada", "kanada", valuecompare.Near, 5.0/6, valuecompare.CauseSpelling)}},
		{name: "short variant", a: "Ann", b: "Anne", want: valuecompare.Near, sim: 0.75, cause: valuecompare.CauseSpelling, parts: []valuecompare.PartMatch{pm("ann", "anne", valuecompare.Near, 0.75, valuecompare.CauseSpelling)}},
		{name: "substitution", a: "Mary", b: "Mark", want: valuecompare.Different, sim: 0, cause: valuecompare.CauseNoMatch, onlyA: []string{"mary"}, onlyB: []string{"mark"}},
		{name: "empty", a: "!!!", b: "York", want: valuecompare.Unknown, sim: 0, cause: valuecompare.CauseEmpty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, det := valuecompare.Compare(textValue(tc.a), textValue(tc.b), rules)
			if res.Relation != tc.want || math.Abs(res.Similarity-tc.sim) > 1e-9 || det.Cause != tc.cause {
				t.Fatalf("Compare(%q, %q) = %+v cause %s, want %s %v %s", tc.a, tc.b, res, det.Cause, tc.want, tc.sim, tc.cause)
			}
			if det.Broader != tc.broader {
				t.Errorf("Broader = %q, want %q", det.Broader, tc.broader)
			}
			if !slices.Equal(words(det.OnlyA), tc.onlyA) || !slices.Equal(words(det.OnlyB), tc.onlyB) {
				t.Errorf("OnlyA %v OnlyB %v, want %v %v", words(det.OnlyA), words(det.OnlyB), tc.onlyA, tc.onlyB)
			}
			if !slices.Equal(det.Parts, tc.parts) {
				t.Errorf("Parts = %+v, want %+v", det.Parts, tc.parts)
			}
		})
	}
}

func TestTextLadder(t *testing.T) {
	rules := matchregistry.Default().Value
	pairs := [][2]string{
		{"York", "York"},
		{"Montréal", "Montreal"},
		{"Provenance", "Provenanced"},
		{"York", "York, Upper Canada"},
		{"York", "York, Upper Canada, North America"},
	}
	prev := math.Inf(1)
	for _, p := range pairs {
		res, _ := valuecompare.Compare(textValue(p[0]), textValue(p[1]), rules)
		if res.Similarity >= prev {
			t.Errorf("%q / %q = %v, not below the row above (%v)", p[0], p[1], res.Similarity, prev)
		}
		prev = res.Similarity
	}
	res, _ := valuecompare.Compare(textValue("York"), textValue("York, Upper Canada, North America"), rules)
	if want := 0.7 * 2.0 / 6; math.Abs(res.Similarity-want) > 1e-9 {
		t.Errorf("York / York, Upper Canada, North America = %v, want %v", res.Similarity, want)
	}
}

func TestTextKey(t *testing.T) {
	rules := matchregistry.Default().Value
	key := func(s string, level valuecompare.Relation) (string, bool) {
		return valuecompare.Key(textValue(s), level, rules)
	}
	for _, in := range []string{"god-father", "God Father", "godfather"} {
		if k, _ := valuecompare.TextKey(in, valuecompare.Same, rules.Words); k != "godfather" {
			t.Errorf("Same TextKey(%q) = %q, want godfather", in, k)
		}
	}
	for _, in := range []string{"Mont-réal", "Montreal", "Montréal"} {
		if k, _ := valuecompare.TextKey(in, valuecompare.Equivalent, rules.Words); k != "montreal" {
			t.Errorf("Equivalent TextKey(%q) = %q, want montreal", in, k)
		}
	}
	a, _ := key("Montréal", valuecompare.Same)
	b, _ := key("Montreal", valuecompare.Same)
	if a == b {
		t.Errorf("Same keys of Montréal / Montreal are equal: %q", a)
	}
	a, _ = key("Montréal", valuecompare.Equivalent)
	b, _ = key("Montreal", valuecompare.Equivalent)
	if a != b || a != "montreal" {
		t.Errorf("Equivalent keys %q %q, want montreal", a, b)
	}
	for _, level := range []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent} {
		x, _ := key("Upper Canada, York", level)
		y, _ := key("York, Upper Canada", level)
		if x == y {
			t.Errorf("%s keys of a reordered toponym are equal: %q", level, x)
		}
	}
	if k, ok := key("!!!", valuecompare.Same); ok || k != "" {
		t.Errorf("Key(!!!) = %q, %v", k, ok)
	}
	if k, ok := key("York", valuecompare.Near); ok || k != "" {
		t.Errorf("Key at Near = %q, %v", k, ok)
	}
}

// TestTextBounds: the alignment is quadratic, so the registry bounds the words
// of a text and the length of a word that may pair as a spelling variant.
func TestTextBounds(t *testing.T) {
	rules := matchregistry.Default().Value
	gen := func(prefix string, n int) string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("%s%c%c", prefix, 'a'+i/26, 'a'+i%26)
		}
		return strings.Join(out, " ")
	}
	long := rules.Text.MaxWords + 1
	t.Run("too long", func(t *testing.T) {
		res, det := valuecompare.Compare(textValue(gen("x", long)), textValue(gen("y", long)), rules)
		if res.Relation != valuecompare.Unknown || res.Similarity != 0 || det.Cause != valuecompare.CauseTooLong ||
			len(det.Parts) != 0 || len(det.OnlyA) != 0 || len(det.OnlyB) != 0 {
			t.Errorf("= %+v %+v", res, det)
		}
	})
	t.Run("over-limit same", func(t *testing.T) {
		s := gen("x", long)
		if res, det := valuecompare.Compare(textValue(s), textValue(s), rules); res.Relation != valuecompare.Same || det.Cause != valuecompare.CauseExact {
			t.Errorf("= %+v %+v", res, det)
		}
	})
	t.Run("over-limit fold", func(t *testing.T) {
		a := gen("x", long)
		if res, det := valuecompare.Compare(textValue(a+" é"), textValue(a+" e"), rules); res.Relation != valuecompare.Equivalent || det.Cause != valuecompare.CauseFold {
			t.Errorf("= %+v %+v", res, det)
		}
	})
	t.Run("spelling length", func(t *testing.T) {
		at := strings.Repeat("a", rules.Words.MaxSpellingLength)
		res, det := valuecompare.Compare(textValue(at), textValue(at[1:]), rules)
		if res.Relation != valuecompare.Near || det.Cause != valuecompare.CauseSpelling {
			t.Errorf("at the limit = %+v %+v", res, det)
		}
		res, det = valuecompare.Compare(textValue(at), textValue(at+"a"), rules)
		if res.Relation != valuecompare.Different || det.Cause != valuecompare.CauseNoMatch {
			t.Errorf("over the limit = %+v %+v", res, det)
		}
	})
}

// TestTextKeyMatchesWords: the one-pass key is the words joined with no
// separator, folded word by word at Equivalent.
func TestTextKeyMatchesWords(t *testing.T) {
	rules := matchregistry.Default().Value
	texts := []string{
		"Mc Donald", "O. Brien", "god-father", "Mont-réal", "Sainte-Anne register",
		"Straße ß", "!!!", "", "Upper Canada, York", "van der Berg", "e\u0301cole", "Ｆull ŁÓDŹ",
		"서울 한국", norm.NFD.String("서울 한국"),
		strings.Repeat("Le père Noël a mangé des crêpes à l'église de Saint-Étienne. ", 35),
	}
	for _, p := range samples[propertymeta.ValueTypeText] {
		texts = append(texts, p.a.Text, p.b.Text)
	}
	rng := rand.New(rand.NewSource(1))
	for range 200 {
		p := generators[propertymeta.ValueTypeText](rng)
		texts = append(texts, p.a.Text, p.b.Text)
	}
	for _, in := range texts {
		words := valuecompare.SplitWords(in, rules.Words)
		for _, level := range []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent} {
			want := slices.Clone(words)
			if level == valuecompare.Equivalent {
				for i, w := range want {
					want[i] = valuecompare.FoldWord(w, rules.Words.Folds)
				}
			}
			got, ok := valuecompare.TextKey(in, level, rules.Words)
			if ok != (len(words) > 0) || got != strings.Join(want, "") {
				t.Errorf("%s TextKey(%.40q) = %q, %v; want %q", level, in, got, ok, strings.Join(want, ""))
			}
		}
	}
}

// TestTextKeyMatchesWordsEveryRune: for every valid rune, alone, after a
// letter and before a combining mark, the key is the words joined, folded word
// by word at Equivalent. The per-rune rules in the key and in the word helpers
// cannot drift apart.
func TestTextKeyMatchesWordsEveryRune(t *testing.T) {
	rules := matchregistry.Default().Value
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if !utf8.ValidRune(r) {
			continue
		}
		for _, in := range []string{string(r), "a" + string(r), string(r) + "\u0301"} {
			words := valuecompare.SplitWords(in, rules.Words)
			for _, level := range []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent} {
				want := slices.Clone(words)
				if level == valuecompare.Equivalent {
					for i, w := range want {
						want[i] = valuecompare.FoldWord(w, rules.Words.Folds)
					}
				}
				got, ok := valuecompare.TextKey(in, level, rules.Words)
				if ok != (len(words) > 0) || got != strings.Join(want, "") {
					t.Fatalf("%s TextKey(%+q) = %q, %v; want %q", level, in, got, ok, strings.Join(want, ""))
				}
			}
		}
	}
}
