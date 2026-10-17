package valuecompare_test

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

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
		parts        int
	}{
		{name: "exact", a: "York", b: "York", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseExact, parts: 1},
		{name: "case", a: "York", b: "york", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: 1},
		{name: "particle", a: "O Brien", b: "O'Brien", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: 1},
		{name: "dash", a: "Sainte-Anne", b: "Sainte Anne", want: valuecompare.Same, sim: 1, cause: valuecompare.CauseCase, parts: 2},
		{name: "fold", a: "Montréal", b: "Montreal", want: valuecompare.Equivalent, sim: 0.95, cause: valuecompare.CauseFold, parts: 1},
		{
			name: "contains", a: "York", b: "York, Upper Canada",
			want: valuecompare.Contains, sim: 0.7 * (1 + 1) / (1 + 3), cause: valuecompare.CauseMissingPart,
			broader: valuecompare.SideA, onlyB: []string{"upper", "canada"}, parts: 1,
		},
		{
			name: "contains from the other side", a: "York, Upper Canada", b: "York",
			want: valuecompare.Contains, sim: 0.7 * (1 + 1) / (1 + 3), cause: valuecompare.CauseMissingPart,
			broader: valuecompare.SideB, onlyA: []string{"upper", "canada"}, parts: 1,
		},
		{
			name: "overlaps", a: "York, Ontario", b: "Toronto, York",
			want: valuecompare.Overlaps, sim: 0.7 * (1 + 1) / (2 + 2), cause: valuecompare.CauseOverlap,
			onlyA: []string{"ontario"}, onlyB: []string{"toronto"}, parts: 1,
		},
		{name: "order", a: "Upper Canada, York", b: "York, Upper Canada", want: valuecompare.Different, sim: 0, cause: valuecompare.CauseOrder, onlyA: []string{"york"}, onlyB: []string{"york"}, parts: 2},
		{name: "spelling one word", a: "Provenance", b: "Provenanced", want: valuecompare.Near, sim: 1 - 1.0/11, cause: valuecompare.CauseSpelling, parts: 1},
		{name: "spelling two words", a: "Upper Canada", b: "Upper Kanada", want: valuecompare.Near, sim: (1 + 1 + 5.0/6 + 5.0/6) / 4, cause: valuecompare.CauseSpelling, parts: 2},
		{name: "short variant", a: "Ann", b: "Anne", want: valuecompare.Near, sim: 0.75, cause: valuecompare.CauseSpelling, parts: 1},
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
			if len(det.Parts) != tc.parts {
				t.Errorf("Parts = %+v, want %d", det.Parts, tc.parts)
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
