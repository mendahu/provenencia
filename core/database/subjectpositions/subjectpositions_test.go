package subjectpositions

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestSubjectPositions(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	mustUser := func(t *testing.T, c *database.Catalog) {
		t.Helper()
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Jake", r); err != nil {
			t.Fatal(err)
		}
	}
	mustSource := func(t *testing.T, c *database.Catalog, title string) sources.Source {
		t.Helper()
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book_" + title, Origin: sourcetypes.OriginProvenencia, Label: title,
		})
		if err != nil {
			t.Fatal(err)
		}
		src, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID, Title: title,
		})
		if err != nil {
			t.Fatal(err)
		}
		return src
	}
	mustPerson := func(t *testing.T, c *database.Catalog, sourceID []byte, label string) subjects.Subject {
		t.Helper()
		if err := subjecttypes.Install(c); err != nil {
			t.Fatal(err)
		}
		person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		s, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: sourceID, SubjectTypeID: person.ID, Label: label,
		})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	auditCount := func(t *testing.T, c *database.Catalog) int {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "set get round-trip including negatives",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				before := auditCount(t, c)
				p, err := Set(c, sub.ID, -3, 7)
				if err != nil {
					t.Fatal(err)
				}
				if p.GridX != -3 || p.GridY != 7 {
					t.Fatalf("%+v", p)
				}
				got, err := Get(c, sub.ID)
				if err != nil || got.GridX != -3 || got.GridY != 7 {
					t.Fatalf("%v %+v", err, got)
				}
				if auditCount(t, c) != before {
					t.Fatal("set wrote audit")
				}
			},
		},
		{
			name: "set overwrites same subject",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				if _, err := Set(c, sub.ID, 1, 1); err != nil {
					t.Fatal(err)
				}
				if _, err := Set(c, sub.ID, 2, 3); err != nil {
					t.Fatal(err)
				}
				got, err := Get(c, sub.ID)
				if err != nil || got.GridX != 2 || got.GridY != 3 {
					t.Fatalf("%v %+v", err, got)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM subject_positions`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 1 {
					t.Fatalf("rows %d", n)
				}
			},
		},
		{
			name: "get missing is tray; clear returns to tray",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				_, err := Get(c, sub.ID)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				if _, err := Set(c, sub.ID, 0, 0); err != nil {
					t.Fatal(err)
				}
				before := auditCount(t, c)
				if err := Clear(c, sub.ID); err != nil {
					t.Fatal(err)
				}
				if err := Clear(c, sub.ID); err != nil {
					t.Fatal(err)
				}
				_, err = Get(c, sub.ID)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				if auditCount(t, c) != before {
					t.Fatal("clear wrote audit")
				}
			},
		},
		{
			name: "list by source scoped; unplaced omitted",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				srcA := mustSource(t, c, "A")
				srcB := mustSource(t, c, "B")
				a1 := mustPerson(t, c, srcA.ID, "A1")
				a2 := mustPerson(t, c, srcA.ID, "A2")
				b1 := mustPerson(t, c, srcB.ID, "B1")
				if _, err := Set(c, a1.ID, 1, 0); err != nil {
					t.Fatal(err)
				}
				if _, err := Set(c, b1.ID, 9, 9); err != nil {
					t.Fatal(err)
				}
				// a2 stays unplaced
				_ = a2
				list, err := ListBySource(c, srcA.ID)
				if err != nil || len(list) != 1 || string(list[0].SubjectID) != string(a1.ID) {
					t.Fatalf("%v %+v", err, list)
				}
			},
		},
		{
			name: "set unknown subject is invalid",
			run: func(t *testing.T, c *database.Catalog) {
				missing := make([]byte, 16)
				missing[0] = 1
				if _, err := Set(c, missing, 0, 0); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				if _, err := Set(c, []byte{1}, 0, 0); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "reopen persists position",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c, "Deed")
				sub := mustPerson(t, c, src.ID, "Alice")
				if _, err := Set(c, sub.ID, 4, -1); err != nil {
					t.Fatal(err)
				}
				dir := c.Dir()
				if err := c.Close(); err != nil {
					t.Fatal(err)
				}
				reopened, err := database.Open(dir)
				if err != nil {
					t.Fatal(err)
				}
				defer reopened.Close()
				got, err := Get(reopened, sub.ID)
				if err != nil || got.GridX != 4 || got.GridY != -1 {
					t.Fatalf("%v %+v", err, got)
				}
				list, err := ListBySource(reopened, src.ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			closedEarly := false
			defer func() {
				if !closedEarly {
					_ = c.Close()
				}
			}()
			if tt.name == "reopen persists position" {
				closedEarly = true
			}
			tt.run(t, c)
		})
	}
}
