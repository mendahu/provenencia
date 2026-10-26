package catalogsession_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/namealiases"
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

func newProject(t *testing.T) string {
	t.Helper()
	t.Cleanup(func() { _ = catalogsession.CloseAll() })
	created, err := onboarding.Complete(t.TempDir(), t.TempDir(), "Jake", "Robins")
	if err != nil {
		t.Fatal(err)
	}
	return created.ProjectDir
}

// assertKindsLoaded fails when loading the person, event or place kind issues
// a SELECT.
func assertKindsLoaded(t *testing.T, c *database.Catalog) {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	selects := databasetest.CountSelects(t, c)
	for _, key := range []string{"person", "event", "place"} {
		var typeID []byte
		if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = ? AND origin = 'provenencia'`, key).Scan(&typeID); err != nil {
			t.Fatal(err)
		}
		before := selects.N()
		if _, err := c.Graph().Kind(typeID); err != nil {
			t.Fatal(err)
		}
		if n := selects.N() - before; n != 0 {
			t.Errorf("Kind(%s) issued %d SELECTs, want 0", key, n)
		}
	}
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
		assertKindsLoaded(t, c)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// A call that throws the warm state away warms it again before it returns, a
// handle commit does not, and a failed call does not.
func TestDoRewarms(t *testing.T) {
	handle := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	cases := []struct {
		name      string
		fn        func(*database.Catalog) error
		wantErr   bool
		wantIndex bool
		wantWarm  bool // WarmNeeded afterwards
		kinds     bool
	}{
		{"AC-2.1 vocabulary commit", func(c *database.Catalog) error {
			return c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListVocabulary}}})
		}, false, true, false, false},
		{"AC-2.2 drop", func(c *database.Catalog) error { c.Graph().Drop(); return nil }, false, true, false, true},
		{"AC-2.3 alias warm", func(c *database.Catalog) error { return namealiases.Warm(c) }, false, true, false, false},
		{"AC-2.4 handle commit", func(c *database.Catalog) error {
			return c.Graph().OnCommit(effects.Commit{Effects: effects.Set{Handles: [][]byte{handle}}})
		}, false, false, false, false},
		{"AC-2.5 failed call", func(c *database.Catalog) error { c.Graph().Drop(); return errors.New("boom") }, true, false, true, false},
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
				if tc.kinds {
					assertKindsLoaded(t, c)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
