package valuerules_test

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func given(text string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeName, NamePattern: valuecompare.WesternPattern,
		Name: &namevalue.Value{Parts: []namevalue.Part{{Value: text, Type: namevalue.PartTypeGiven}}}}
}

func newCatalog(t *testing.T) *database.Catalog {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestLoad(t *testing.T) {
	t.Run("an installed catalog", func(t *testing.T) {
		c := newCatalog(t)
		if err := nameformats.Install(c); err != nil {
			t.Fatal(err)
		}
		if err := namealiases.Install(c); err != nil {
			t.Fatal(err)
		}
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		rules, key, err := valuerules.Load(nil, db)
		if err != nil {
			t.Fatal(err)
		}
		if key != valuecompare.WesternPattern {
			t.Errorf("key %q", key)
		}
		loaded, err := nameformats.Load(db)
		if err != nil {
			t.Fatal(err)
		}
		if got, ok := rules.Names.Patterns[valuecompare.WesternPattern]; !ok || !reflect.DeepEqual(got, loaded.Profile.Pattern) {
			t.Errorf("Names.Patterns[western] = %+v, want the stored pattern", rules.Names.Patterns)
		}
		res, _ := valuecompare.Compare(given("Jas."), given("James"), rules)
		if res.Relation != valuecompare.Equivalent {
			t.Errorf("Jas. / James is %s, want equivalent", res.Relation)
		}
	})
	t.Run("no default name format", func(t *testing.T) {
		c := newCatalog(t)
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		rules, key, err := valuerules.Load(nil, db)
		if err != nil {
			t.Fatal(err)
		}
		if key != valuecompare.WesternPattern {
			t.Errorf("key %q", key)
		}
		if len(rules.Names.Patterns) != 0 {
			t.Errorf("Names.Patterns %+v, want none", rules.Names.Patterns)
		}
		if !reflect.DeepEqual(rules.Names.Western, matchregistry.Default().Value.Names.Western) {
			t.Errorf("Names.Western differs from the registry's")
		}
	})
}
