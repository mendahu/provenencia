package catalogsession_test

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/onboarding"
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
		cmp, err := namealiases.Comparer(c.Graph(), db, match.DefaultNames)
		if err != nil {
			return err
		}
		got, ok := cmp.Compare(aliasGiven("Jas."), aliasGiven("James"))
		if !ok || math.Abs(got-match.AliasCredit) > 1e-9 {
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
		cmp, err := namealiases.Comparer(c.Graph(), db, match.DefaultNames)
		if err != nil {
			return err
		}
		got, ok := cmp.Compare(aliasGiven("Jas."), aliasGiven("James"))
		if ok && math.Abs(got-match.AliasCredit) < 1e-9 {
			t.Fatal("reopen kept the previous project's alias cache")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func aliasGiven(text string) match.Value {
	return match.Value{Name: &namevalue.Value{Parts: []namevalue.Part{{
		Value: text, Type: namevalue.PartTypeGiven,
	}}}}
}
