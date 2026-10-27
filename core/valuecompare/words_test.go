package valuecompare_test

import (
	"math"
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

// The spelling-variant rule used by the name comparer (FuzzyFloor 0.8; one
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

var nameRoles = []valuecompare.NameRole{valuecompare.RoleGiven, valuecompare.RoleNick, valuecompare.RoleUntyped}

func aliasIndexOf(kind valuecompare.AliasKind, sets ...[]string) valuecompare.AliasIndex {
	return valuecompare.NewAliasIndex([]valuecompare.AliasTable{{Key: "t", Kind: kind, Roles: nameRoles, Sets: sets}}, matchregistry.Default().Value.Words)
}

// Names add initials and alias tables to the word table; text adds neither.
func TestWordSimilarityNames(t *testing.T) {
	words := matchregistry.Default().Value.Words
	abbreviations := aliasIndexOf(valuecompare.AliasAbbreviation, []string{"james", "jas"})
	nicknames := aliasIndexOf(valuecompare.AliasNickname, []string{"james", "jim"})
	given, family := valuecompare.RoleGiven, valuecompare.RoleFamily
	for _, tc := range []struct {
		name string
		x, y string
		opts valuecompare.WordOptionsType
		sim  float64
		rel  valuecompare.Relation
		why  valuecompare.Cause
	}{
		{"initial with initials", "j", "james", valuecompare.WordOptions(true, valuecompare.AliasIndex{}, given, given), words.InitialCredit, valuecompare.Contains, valuecompare.CauseInitial},
		{"initial without initials", "j", "james", valuecompare.WordOptions(false, valuecompare.AliasIndex{}, given, given), 0, valuecompare.Different, valuecompare.CauseNoMatch},
		{"initial of another word", "w", "james", valuecompare.WordOptions(true, valuecompare.AliasIndex{}, given, given), 0, valuecompare.Different, valuecompare.CauseNoMatch},
		{"abbreviation", "jas", "james", valuecompare.WordOptions(true, abbreviations, given, given), words.AliasCredit, valuecompare.Equivalent, valuecompare.CauseAbbreviation},
		{"nickname", "jim", "james", valuecompare.WordOptions(true, nicknames, given, given), words.AliasCredit, valuecompare.Equivalent, valuecompare.CauseAlias},
		{"nickname as surnames", "jim", "james", valuecompare.WordOptions(true, nicknames, family, family), 0, valuecompare.Different, valuecompare.CauseNoMatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, rel, why := valuecompare.WordSimilarity(tc.x, tc.y, words, tc.opts)
			if math.Abs(sim-tc.sim) > 1e-9 || rel != tc.rel || why != tc.why {
				t.Errorf("got %v %s %s, want %v %s %s", sim, rel, why, tc.sim, tc.rel, tc.why)
			}
		})
	}
}

// The index normalizes its tokens, drops empty ones, and the zero index hits
// nothing.
func TestAliasIndex(t *testing.T) {
	given := valuecompare.RoleGiven
	idx := aliasIndexOf(valuecompare.AliasAbbreviation, []string{"Jas.", " JIM ", ""}, []string{"", "  "})
	if kind, ok := idx.Hit("jas", "jim", given, given); !ok || kind != valuecompare.AliasAbbreviation {
		t.Errorf("normalized tokens: got %q %v", kind, ok)
	}
	if _, ok := idx.Hit("jas", "", given, given); ok {
		t.Error("an empty token hit")
	}
	if _, ok := idx.Hit("jas", "jim", valuecompare.RoleFamily, given); ok {
		t.Error("a role the table does not allow hit")
	}
	// A token in several sets: Hit finds the set that holds the other token,
	// and misses when no one set holds both.
	multi := valuecompare.NewAliasIndex([]valuecompare.AliasTable{{
		Key: "t", Kind: valuecompare.AliasNickname, Roles: []valuecompare.NameRole{given},
		Sets: [][]string{{"ben", "benji"}, {"ben", "benedict"}, {"benjamin", "jamie"}},
	}}, matchregistry.Default().Value.Words)
	if _, ok := multi.Hit("ben", "benedict", given, given); !ok {
		t.Error("a token in several sets missed the later set")
	}
	if _, ok := multi.Hit("benedict", "ben", given, given); !ok {
		t.Error("the hit is not symmetric")
	}
	if _, ok := multi.Hit("benji", "benedict", given, given); ok {
		t.Error("tokens in different sets hit")
	}
	if _, ok := multi.Hit("ben", "jamie", given, given); ok {
		t.Error("tokens in different sets hit")
	}
	if _, ok := (valuecompare.AliasIndex{}).Hit("jas", "jim", given, given); ok {
		t.Error("the zero index hit")
	}
}

// TestAliasIndexKeysTokens: the index keys a token as a name part produces it,
// so a dash or a particle space in the table meets the joined comparison token.
func TestAliasIndexKeysTokens(t *testing.T) {
	words := matchregistry.Default().Value.Words
	for _, tc := range []struct {
		name string
		role valuecompare.NameRole
		set  []string
		x, y string
	}{
		{"dash", valuecompare.RoleGiven, []string{"Mary-Ann", "Polly"}, "maryann", "polly"},
		{"particle", valuecompare.RoleUntyped, []string{"Mc Donald", "MacDonald"}, "mcdonald", "macdonald"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			idx := valuecompare.NewAliasIndex([]valuecompare.AliasTable{{
				Key: "t", Kind: valuecompare.AliasNickname, Roles: []valuecompare.NameRole{tc.role}, Sets: [][]string{tc.set},
			}}, words)
			sim, _, _ := valuecompare.WordSimilarity(tc.x, tc.y, words, valuecompare.WordOptions(true, idx, tc.role, tc.role))
			if math.Abs(sim-words.AliasCredit) > 1e-9 {
				t.Errorf("%s ~ %s scored %v, want AliasCredit %v", tc.x, tc.y, sim, words.AliasCredit)
			}
		})
	}
}

func TestAliasIndexExpands(t *testing.T) {
	roles := []valuecompare.NameRole{valuecompare.RoleGiven, valuecompare.RoleNick, valuecompare.RoleUntyped}
	table := valuecompare.AliasTable{
		Key: "t", Kind: valuecompare.AliasNickname, Roles: roles, FormalFirst: true,
		Sets: [][]string{
			{"christopher", "chris", "kit"}, {"chris", "kris"}, {"abel", "ebbie"}, {"raphael", "ralph"},
			{"william", "bill"}, {"bill", "william"}, {"allan", "allen"}, {"allen", "allan"},
		},
		Directions: []valuecompare.AliasDirection{{Formal: "william", Nickname: "bill"}},
	}
	w := matchregistry.Default().Value.Words
	idx := valuecompare.NewAliasIndex([]valuecompare.AliasTable{table}, w)
	plain := table
	plain.FormalFirst = false
	flat := valuecompare.NewAliasIndex([]valuecompare.AliasTable{plain}, w)

	for _, c := range []struct {
		from, into string
		role       valuecompare.NameRole
		want       bool
	}{
		{"chris", "christopher", valuecompare.RoleGiven, true},
		{"christopher", "chris", valuecompare.RoleGiven, false},
		{"kit", "christopher", valuecompare.RoleGiven, true},
		{"chris", "kit", valuecompare.RoleGiven, false},
		{"kris", "chris", valuecompare.RoleGiven, true},
		{"ebbie", "abel", valuecompare.RoleGiven, true},
		{"ralph", "raphael", valuecompare.RoleGiven, true},
		{"bill", "william", valuecompare.RoleGiven, true},
		{"william", "bill", valuecompare.RoleGiven, false},
		{"allan", "allen", valuecompare.RoleGiven, false},
		{"allen", "allan", valuecompare.RoleGiven, false},
		{"chris", "christopher", valuecompare.RoleFamily, false},
	} {
		if got := idx.Expands(c.from, c.into, c.role); got != c.want {
			t.Errorf("Expands(%s, %s, %s) = %v, want %v", c.from, c.into, c.role, got, c.want)
		}
		if flat.Expands(c.from, c.into, c.role) {
			t.Errorf("without FormalFirst, Expands(%s, %s, %s) = true", c.from, c.into, c.role)
		}
	}
}
