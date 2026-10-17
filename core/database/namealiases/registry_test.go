package namealiases

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestSharedNicknameDoesNotMergeCanonicalNames(t *testing.T) {
	tables := AliasTables()
	c := match.NameComparer{Words: match.WordRules{Aliases: &tables}}
	got, ok := c.Compare(given("Chris"), given("Christine"))
	if !ok || math.Abs(got-match.AliasCredit) > 1e-9 {
		t.Fatalf("Chris / Christine: %.4f ok=%v", got, ok)
	}
	got, ok = c.Compare(given("Tina"), given("Christopher"))
	if ok && got > 0 {
		t.Fatalf("Tina / Christopher scored %.4f", got)
	}
}

func TestRegisterAddsAnAliasTable(t *testing.T) {
	t.Cleanup(ResetAliasTablesForTest)
	Register(valuecompare.AliasTable{
		Key: "zz", Origin: "plugin:example", Kind: valuecompare.AliasNickname,
		Roles: aliasPartRoles(), Sets: [][]string{{"zz", "zzzz"}},
	})
	var found bool
	for _, table := range AliasTables() {
		if table.Key == "zz" {
			found = true
		}
	}
	if !found {
		t.Fatal("register did not append to the seed")
	}
	got, ok := (match.NameComparer{}).Compare(given("Zz"), given("Zzzz"))
	if !ok || math.Abs(got-match.AliasCredit) < 1e-9 {
		t.Fatalf("bare comparer used the seed: %.4f %v", got, ok)
	}
}
