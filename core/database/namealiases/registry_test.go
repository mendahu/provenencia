package namealiases

import (
	"math"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// seedRules are the registry's rules with the seed read as the catalog would.
func seedRules() valuecompare.Rules {
	rules := matchregistry.Default().Value
	rules.Names.Aliases = valuecompare.NewAliasIndex(AliasTables(), rules.Words)
	return rules
}

func TestSharedNicknameDoesNotMergeCanonicalNames(t *testing.T) {
	rules := seedRules()
	res, det := valuecompare.Compare(given("Chris"), given("Christine"), rules)
	if res.Relation != valuecompare.Equivalent || math.Abs(res.Similarity-0.98) > 1e-9 || det.Cause != valuecompare.CauseAlias {
		t.Fatalf("Chris / Christine: %+v %+v", res, det)
	}
	res, _ = valuecompare.Compare(given("Tina"), given("Christopher"), rules)
	if res.Relation != valuecompare.Different {
		t.Fatalf("Tina / Christopher: %+v", res)
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
	if res, _ := valuecompare.Compare(given("Zz"), given("Zzzz"), seedRules()); res.Relation != valuecompare.Equivalent {
		t.Fatalf("the registered table did not hit: %+v", res)
	}
	if res, _ := valuecompare.Compare(given("Zz"), given("Zzzz"), matchregistry.Default().Value); res.Relation == valuecompare.Equivalent {
		t.Fatalf("default rules used the seed: %+v", res)
	}
}

// keptTwoWay are the pairs nicknameSets lists both ways and nicknameDirections
// leaves alone: spelling variants, translations, and nicknames of the same
// third name (S11-28 Appendix A).
var keptTwoWay = [][2]string{
	{"agnes", "inez"},
	{"allan", "allen"},
	{"amos", "moses"},
	{"bess", "bessie"},
	{"carla", "karla"},
	{"casper", "jasper"},
	{"chris", "kris"},
	{"ellen", "helen"},
	{"eloise", "heloise"},
	{"erwin", "irwin"},
	{"esther", "hester"},
	{"hailey", "haylee"},
	{"hailey", "hayley"},
	{"haylee", "hayley"},
	{"ian", "john"},
	{"jane", "jean"},
	{"jim", "jimmie"},
	{"joe", "joey"},
	{"john", "jon"},
	{"johnathan", "johnathon"},
	{"johnathan", "jonathan"},
	{"johnathan", "jonathon"},
	{"johnathon", "jonathan"},
	{"johnathon", "jonathon"},
	{"jonathan", "jonathon"},
	{"kimberley", "kimberly"},
	{"lil", "lilly"},
	{"lil", "lily"},
	{"lilly", "lily"},
	{"marie", "mary"},
	{"orlando", "roland"},
	{"rodger", "roger"},
	{"shaun", "shawn"},
	{"stacey", "stacie"},
	{"zach", "zack"},
}

func TestNicknameDirectionsCoverTwoWayPairs(t *testing.T) {
	heads := map[string]map[string]bool{} // head -> names its sets hold
	for _, set := range nicknameSets {
		if heads[set[0]] == nil {
			heads[set[0]] = map[string]bool{}
		}
		for _, name := range set[1:] {
			heads[set[0]][name] = true
		}
	}
	twoWay := map[[2]string]bool{}
	for a, held := range heads {
		for b := range held {
			if heads[b][a] {
				twoWay[[2]string{min(a, b), max(a, b)}] = true
			}
		}
	}
	pair := func(a, b string) [2]string { return [2]string{min(a, b), max(a, b)} }

	want := map[[2]string]bool{}
	for _, d := range nicknameDirections {
		if !heads[d.Formal][d.Nickname] || !heads[d.Nickname][d.Formal] {
			t.Errorf("%s <- %s is not listed both ways", d.Formal, d.Nickname)
		}
		want[pair(d.Formal, d.Nickname)] = true
	}
	for _, k := range keptTwoWay {
		want[pair(k[0], k[1])] = true
	}
	if !reflect.DeepEqual(twoWay, want) {
		for p := range twoWay {
			if !want[p] {
				t.Errorf("two-way pair %v has no direction and is not kept", p)
			}
		}
		for p := range want {
			if !twoWay[p] {
				t.Errorf("%v is directed or kept but not two-way", p)
			}
		}
	}
	if len(nicknameDirections) != 108 || len(keptTwoWay) != 35 || len(twoWay) != 143 {
		t.Errorf("directions %d, kept %d, two-way %d; want 108, 35, 143", len(nicknameDirections), len(keptTwoWay), len(twoWay))
	}
	if len(nicknameSets) != 1229 {
		t.Errorf("nicknameSets has %d sets, want 1229", len(nicknameSets))
	}
}
