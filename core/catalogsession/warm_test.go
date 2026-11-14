package catalogsession_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/valuerules"
	"github.com/mendahu/provenencia/core/onboarding"
)

// A project opened through Do has its default name format warm on the graph,
// so value rules load without a query.
func TestDoWarmsNameFormat(t *testing.T) {
	dir := newProject(t)
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		if _, ok, _ := nameformats.Default.Get(c.Graph()); !ok {
			t.Fatal("the open project has no warm name format")
		}
		selects := databasetest.CountSelects(t, c)
		if _, _, err := valuerules.Load(c.Graph()); err != nil {
			return err
		}
		if _, err := nameformats.LoadFor(c.Graph()); err != nil {
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

func newProject(t *testing.T) string {
	t.Helper()
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}
	return created.ProjectDir
}

// A project opened through Do has matching warm: the match index is built and
// the shortlist kinds are loaded. An empty project opens the same way.
func TestDoWarmsMatching(t *testing.T) {
	dir := newProject(t)
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		g := c.Graph()
		if _, ok := promotealign.MatchIndex.Peek(g); !ok {
			t.Error("the open project has no match index")
		}
		graphcachetest.AssertKindsLoaded(t, c, "person", "event", "place")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// The alias index and name format live with the catalog connection: a vocabulary
// commit, a Drop, a handle commit and a failed call leave them, and no call
// warms them again. The match index follows its declarations: a vocabulary
// commit and a Drop empty it, and a handle commit (which patches it) and a
// Source vocabulary commit do not.
func TestDoKeepsWarmEntries(t *testing.T) {
	handle := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	cases := []struct {
		name      string
		fn        func(*database.Catalog) error
		wantErr   bool
		wantIndex bool // the match index is present afterwards
		kinds     bool // the shortlist kinds are not loaded afterwards
	}{
		{"vocabulary commit", func(c *database.Catalog) error {
			return c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListPropertyVocabulary}}})
		}, false, false, false},
		{"drop", func(c *database.Catalog) error { c.Graph().Drop(); return nil }, false, false, true},
		{"handle commit", func(c *database.Catalog) error {
			return c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Handles: [][]byte{handle}}})
		}, false, true, false},
		{"source vocabulary commit", func(c *database.Catalog) error {
			return c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListSourceVocabulary}}})
		}, false, true, false},
		{"failed call", func(c *database.Catalog) error { c.Graph().Drop(); return errors.New("boom") }, true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := newProject(t)
			if err := catalogsession.Do(dir, func(*database.Catalog) error { return nil }); err != nil {
				t.Fatal(err)
			}
			err := catalogsession.Do(dir, tc.fn)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Do returned %v, want error %v", err, tc.wantErr)
			}
			err = catalogsession.Do(dir, func(c *database.Catalog) error {
				g := c.Graph()
				if _, ok := promotealign.MatchIndex.Peek(g); ok != tc.wantIndex {
					t.Errorf("match index present = %v, want %v", ok, tc.wantIndex)
				}
				selects := databasetest.CountSelects(t, c)
				if _, _, err := valuerules.Load(g); err != nil {
					return err
				}
				if n := selects.N(); n != 0 {
					t.Errorf("valuerules.Load issued %d SELECTs, want 0 (the alias index and name format survive)", n)
				}
				if tc.kinds {
					// A drop empties the kinds; the next proposal loads them.
					typeID := graphcachetest.SubjectTypeID(t, c, "person")
					selects := databasetest.CountSelects(t, c)
					if _, err := g.Kind(typeID); err != nil {
						return err
					}
					if selects.N() == 0 {
						t.Error("the person kind is loaded after a drop, want it left to the next proposal")
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Setting the name format empties exactly what read it: the match index goes;
// the alias index stays.
func TestNameFormatDropsItsReaders(t *testing.T) {
	dir := newProject(t)
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		g := c.Graph()
		if _, ok := promotealign.MatchIndex.Peek(g); !ok {
			t.Fatal("the open project has no match index")
		}
		loaded, ok, err := nameformats.Default.Get(g)
		if err != nil || !ok {
			t.Fatalf("name format = %v, %v", ok, err)
		}
		nameformats.Default.Set(g, loaded)
		if _, ok := promotealign.MatchIndex.Peek(g); ok {
			t.Error("the match index survived a new name format")
		}
		selects := databasetest.CountSelects(t, c)
		if _, _, err := valuerules.Load(g); err != nil {
			return err
		}
		if n := selects.N(); n != 0 {
			t.Errorf("valuerules.Load issued %d SELECTs, want 0", n)
		}
		if _, ok, _ := namealiases.Index.Get(g); !ok {
			t.Error("a new name format dropped the alias index")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// A value that does not decode does not fail the open or a call: the shortlist
// index stays unbuilt and the first proposal reports the error.
func TestOpenSurvivesABadValue(t *testing.T) {
	dir := newProject(t)
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		typeID := graphcachetest.SubjectTypeID(t, c, "event")
		var propID []byte
		if err := db.QueryRow(`SELECT id FROM properties WHERE key = 'date' AND origin = 'provenencia'`).Scan(&propID); err != nil {
			return err
		}
		entity := []byte{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
		if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, 'TST-BAD')`, entity, typeID); err != nil {
			return err
		}
		_, err = db.Exec(`INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_date, support) VALUES (?, ?, 1, ?, 1)`,
			entity, propID, []byte{0xff, 0xff, 0xff})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := catalogsession.Close(dir); err != nil {
		t.Fatal(err)
	}
	err = catalogsession.Do(dir, func(c *database.Catalog) error {
		if _, ok := promotealign.MatchIndex.Peek(c.Graph()); ok {
			t.Error("the shortlist index was built over a value that does not decode")
		}
		if err := promotealign.Warm(c); err == nil {
			t.Error("promotealign.Warm succeeded over a value that does not decode")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("the project did not open over a bad value: %v", err)
	}
}
