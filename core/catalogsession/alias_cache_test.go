package catalogsession_test

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/onboarding"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestDoKeepsAliasTablesForTheOpenProject(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	ident := t.TempDir()
	parent := t.TempDir()
	created, err := onboarding.Complete(ident, parent, "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}

	err = catalogsession.Do(created.ProjectDir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		if _, err := db.Exec(`DELETE FROM name_alias_members WHERE token = 'jas'`); err != nil {
			return err
		}
		rules, err := namealiases.Rules(c.Graph(), db, matchregistry.Default().Value)
		if err != nil {
			return err
		}
		res, _ := valuecompare.Compare(aliasGiven("Jas."), aliasGiven("James"), rules)
		got, ok := res.Similarity, res.Relation != valuecompare.Unknown
		if !ok || math.Abs(got-matchregistry.Default().Value.Words.AliasCredit) > 1e-9 {
			t.Fatalf("open project missed the warmed seed: %.4f %v", got, ok)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := catalogsession.Close(created.ProjectDir); err != nil {
		t.Fatal(err)
	}
	err = catalogsession.Do(created.ProjectDir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		rules, err := namealiases.Rules(c.Graph(), db, matchregistry.Default().Value)
		if err != nil {
			return err
		}
		res, _ := valuecompare.Compare(aliasGiven("Jas."), aliasGiven("James"), rules)
		got, ok := res.Similarity, res.Relation != valuecompare.Unknown
		if ok && math.Abs(got-matchregistry.Default().Value.Words.AliasCredit) < 1e-9 {
			t.Fatal("reopen kept the previous project's alias cache")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func aliasGiven(text string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: namevaluetest.Given(text)}
}
