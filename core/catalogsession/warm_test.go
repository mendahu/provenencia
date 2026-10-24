package catalogsession_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/onboarding"
)

// A project opened through Do has its default name format warm on the graph,
// so value rules load without a query.
func TestDoWarmsNameFormat(t *testing.T) {
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}
	err = catalogsession.Do(created.ProjectDir, func(c *database.Catalog) error {
		raw, ok := c.Graph().NameFormat()
		if !ok {
			t.Fatal("the open project has no warm name format")
		}
		if _, ok := raw.(nameformats.Loaded); !ok {
			t.Fatalf("the warm name format is a %T", raw)
		}
		db, err := c.DB()
		if err != nil {
			return err
		}
		selects := databasetest.CountSelects(t, c)
		if _, _, err := valuerules.Load(c.Graph(), db); err != nil {
			return err
		}
		if _, err := nameformats.LoadFor(c.Graph(), db); err != nil {
			return err
		}
		if n := selects.N(); n != 0 {
			t.Errorf("valuerules.Load and nameformats.LoadFor issued %d SELECTs, want 0", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
