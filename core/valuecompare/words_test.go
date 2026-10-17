package valuecompare_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestNormalizeForm(t *testing.T) {
	for in, want := range map[string]string{
		"  James   Robins ": "james robins",
		"J. Robins":         "j robins",
		"O'Brien, Mary":     "obrien mary",
		"ÉMILE\tZOLA":       "émile zola",
		"Smith-Jones":       "smith jones",
		"Mary–Ann / Polly":  "mary ann polly",
		"-Robins-":          "robins",
		"":                  "",
	} {
		if got := valuecompare.NormalizeForm(in); got != want {
			t.Errorf("NormalizeForm(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsInitial(t *testing.T) {
	for w, want := range map[string]bool{"j": true, "J": true, "é": true, "jo": false, "蒋": false, "": false, "1": false} {
		if valuecompare.IsInitial(w) != want {
			t.Errorf("IsInitial(%q) = %v", w, !want)
		}
	}
}

// The spelling-variant rule shared with core/match (FuzzyFloor 0.8; one
// added or dropped letter from ShortVariantMinLength letters up).
func TestSpellingSimilarity(t *testing.T) {
	words := matchregistry.Default().Value.Words
	for _, tc := range []struct {
		x, y    string
		variant bool
	}{
		{"robins", "robbins", true},
		{"robnis", "robins", true}, // a swap is one edit
		{"smith", "smyth", true},
		{"catherine", "katherine", true},
		{"johann", "johan", true},
		{"macdonald", "mcdonald", true},
		{"ann", "anne", true}, // short word, one added letter
		{"jon", "john", true},
		{"dan", "dean", true}, // a known false positive of the short-word rule
		{"mary", "mark", false},
		{"josé", "jose", false}, // accents kept
		{"smith", "smythe", false},
		{"james", "jake", false},
		{"james", "jim", false},
		{"kenneth", "kevin", false},
		{"jr", "sr", false},
	} {
		got := valuecompare.SpellingSimilarity(tc.x, tc.y, words.FuzzyFloor, words.ShortVariantMinLength) > 0
		if got != tc.variant {
			t.Errorf("%s ~ %s: variant %v, want %v", tc.x, tc.y, got, tc.variant)
		}
	}
	if valuecompare.SpellingSimilarity("robins", "robbins", 1, words.ShortVariantMinLength) != 0 {
		t.Error("a floor of 1 turns spelling variants off")
	}
}
