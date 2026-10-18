package namealiases

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// givenName is a name value of one given part.
func givenName(text string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: &namevalue.Value{Parts: []namevalue.Part{{
		Value: text, Type: namevalue.PartTypeGiven,
	}}}}
}

// seedRules are the registry's rules with the seed read as the catalog would.
func seedRules() valuecompare.Rules {
	rules := matchregistry.Default().Value
	rules.Names.Aliases = valuecompare.NewAliasIndex(AliasTables(), rules.Words)
	return rules
}

func TestSharedNicknameDoesNotMergeCanonicalNames(t *testing.T) {
	rules := seedRules()
	res, det := valuecompare.Compare(givenName("Chris"), givenName("Christine"), rules)
	if res.Relation != valuecompare.Equivalent || math.Abs(res.Similarity-0.98) > 1e-9 || det.Cause != valuecompare.CauseAlias {
		t.Fatalf("Chris / Christine: %+v %+v", res, det)
	}
	res, _ = valuecompare.Compare(givenName("Tina"), givenName("Christopher"), rules)
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
	if res, _ := valuecompare.Compare(givenName("Zz"), givenName("Zzzz"), seedRules()); res.Relation != valuecompare.Equivalent {
		t.Fatalf("the registered table did not hit: %+v", res)
	}
	if res, _ := valuecompare.Compare(givenName("Zz"), givenName("Zzzz"), matchregistry.Default().Value); res.Relation == valuecompare.Equivalent {
		t.Fatalf("default rules used the seed: %+v", res)
	}
}
