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
		if _, ok := g.MatchIndex(); !ok {
			t.Error("the open project has no match index")
		}
		if g.WarmNeeded() {
			t.Error("the open project still needs a warm")
		}
		graphcachetest.AssertKindsLoaded(t, c, "person", "event", "place")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// A call that throws the warm state away warms the alias index and name format
// again before it returns and leaves the shortlist index to the next proposal;
// a handle commit warms nothing, and a failed call does not.
func TestDoRewarms(t *testing.T) {
	handle := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	cases := []struct {
		name      string
		fn        func(*database.Catalog) error
		wantErr   bool
		wantIndex bool
		wantWarm  bool // WarmNeeded afterwards
		aliases   bool // the alias index and name format are present afterwards
		kinds     bool // the shortlist kinds are not loaded afterwards
	}{
		{"AC-2.1 vocabulary commit", func(c *database.Catalog) error {
			return c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListVocabulary}}})
		}, false, false, false, true, false},
		{"AC-2.2 drop", func(c *database.Catalog) error { c.Graph().Drop(); return nil }, false, false, false, true, true},
		{"AC-2.3 alias warm", func(c *database.Catalog) error { return namealiases.Warm(c) }, false, false, false, true, false},
		{"AC-2.4 handle commit", func(c *database.Catalog) error {
			return c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Handles: [][]byte{handle}}})
		}, false, false, false, true, false},
		{"AC-2.5 failed call", func(c *database.Catalog) error { c.Graph().Drop(); return errors.New("boom") }, true, false, true, true, false},
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
			// fn of a second Do runs before that Do's own warm, so it sees what the
			// first left.
			err = catalogsession.Do(dir, func(c *database.Catalog) error {
				g := c.Graph()
				if _, ok := g.MatchIndex(); ok != tc.wantIndex {
					t.Errorf("match index present = %v, want %v", ok, tc.wantIndex)
				}
				if g.WarmNeeded() != tc.wantWarm {
					t.Errorf("WarmNeeded = %v, want %v", g.WarmNeeded(), tc.wantWarm)
				}
				_, aliases := g.AliasIndex()
				_, format := g.NameFormat()
				if aliases != tc.aliases || format != tc.aliases {
					t.Errorf("alias index present = %v, name format present = %v, want %v", aliases, format, tc.aliases)
				}
				if tc.kinds {
					// A drop empties the kinds; the next proposal loads them.
					db, err := c.DB()
					if err != nil {
						return err
					}
					var typeID []byte
					if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = 'person' AND origin = 'provenencia'`).Scan(&typeID); err != nil {
						return err
					}
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

// A value that does not decode does not fail the open or a call: the shortlist
// index stays unbuilt and the first proposal reports the error.
func TestOpenSurvivesABadValue(t *testing.T) {
	dir := newProject(t)
	err := catalogsession.Do(dir, func(c *database.Catalog) error {
		db, err := c.DB()
		if err != nil {
			return err
		}
		var typeID, propID []byte
		if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = 'event' AND origin = 'provenencia'`).Scan(&typeID); err != nil {
			return err
		}
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
		if _, ok := c.Graph().MatchIndex(); ok {
			t.Error("the shortlist index was built over a value that does not decode")
		}
		if err := promotealign.Warm(c); err == nil {
			t.Error("promotealign.Warm succeeded over a value that does not decode")
		}
		_, _, err := promotealign.Propose(c, make([]byte, 16), nil)
		if err == nil {
			t.Error("a proposal succeeded over a value that does not decode")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("the project did not open over a bad value: %v", err)
	}
}
