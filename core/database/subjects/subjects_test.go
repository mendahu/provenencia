package subjects

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestSubjects(t *testing.T) {
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
	mustSource := func(t *testing.T, c *database.Catalog) sources.Source {
		t.Helper()
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		src, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID,
			Title:        "Deed",
		})
		if err != nil {
			t.Fatal(err)
		}
		return src
	}
	mustSubjectTypes := func(t *testing.T, c *database.Catalog) {
		t.Helper()
		if err := subjecttypes.Install(c); err != nil {
			t.Fatal(err)
		}
	}
	latestAction := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actionType string
		if err := db.QueryRow(`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`).Scan(&actionType); err != nil {
			t.Fatal(err)
		}
		return actionType
	}
	latestEntityType := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var entityType string
		if err := db.QueryRow(`
			SELECT entity_type FROM audit_changes
			ORDER BY rowid DESC LIMIT 1`).Scan(&entityType); err != nil {
			t.Fatal(err)
		}
		return entityType
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create rejects empty user id",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				_, err = Create(c, nil, CreateInput{
					SourceID:      src.ID,
					SubjectTypeID: person.ID,
					Label:         "Alice",
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "create person mints CPR ref and audits",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := Create(c, userID, CreateInput{
					SourceID:      src.ID,
					SubjectTypeID: person.ID,
					Label:         "Alice",
				})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(s.Ref, "CPR-") {
					t.Fatalf("ref %q", s.Ref)
				}
				if latestAction(t, c) != "create_subject" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				if latestEntityType(t, c) != "subject" {
					t.Fatalf("entity %s", latestEntityType(t, c))
				}
				got, err := Get(c, s.ID)
				if err != nil || got.Label != "Alice" {
					t.Fatalf("%v %+v", err, got)
				}
				byRef, err := getByRef(c, s.Ref)
				if err != nil || string(byRef.ID) != string(s.ID) {
					t.Fatalf("%v %+v", err, byRef)
				}
			},
		},
		{
			name: "create event mints CEV ref",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				event, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := Create(c, userID, CreateInput{
					SourceID:      src.ID,
					SubjectTypeID: event.ID,
				})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(s.Ref, "CEV-") {
					t.Fatalf("ref %q", s.Ref)
				}
			},
		},
		{
			name: "update label records update_subject",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := Create(c, userID, CreateInput{
					SourceID: src.ID, SubjectTypeID: person.ID, Label: "Alice",
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := Update(c, userID, s.ID, "Alicia", ""); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_subject" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				got, err := Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Label != "Alicia" || string(got.SubjectTypeID) != string(person.ID) {
					t.Fatalf("%+v", got)
				}
				if err := Update(c, userID, s.ID, "Alicia", ""); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "delete records delete_subject and cascades positions",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				src := mustSource(t, c)
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				s, err := Create(c, userID, CreateInput{
					SourceID: src.ID, SubjectTypeID: person.ID, Label: "Bob",
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO subject_positions (subject_id, grid_x, grid_y) VALUES (?, 1, 2)`,
					s.ID,
				); err != nil {
					t.Fatal(err)
				}
				if err := Delete(c, userID, s.ID); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_subject" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				_, err = Get(c, s.ID)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM subject_positions WHERE subject_id = ?`, s.ID).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 0 {
					t.Fatalf("positions left %d", n)
				}
			},
		},
		{
			name: "list by source is scoped",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				mustSubjectTypes(t, c)
				srcA := mustSource(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "census", Origin: sourcetypes.OriginProvenencia, Label: "Census",
				})
				if err != nil {
					t.Fatal(err)
				}
				srcB, err := sources.Create(c, userID, sources.CreateInput{
					SourceTypeID: typeID, Title: "Other",
				})
				if err != nil {
					t.Fatal(err)
				}
				person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Create(c, userID, CreateInput{SourceID: srcA.ID, SubjectTypeID: person.ID, Label: "A"}); err != nil {
					t.Fatal(err)
				}
				if _, err := Create(c, userID, CreateInput{SourceID: srcB.ID, SubjectTypeID: person.ID, Label: "B"}); err != nil {
					t.Fatal(err)
				}
				list, err := ListBySource(c, srcA.ID)
				if err != nil || len(list) != 1 || list[0].Label != "A" {
					t.Fatalf("%v %+v", err, list)
				}
			},
		},
		{
			name: "reject invalid ids",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Create(c, userID, CreateInput{}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				if _, err := Get(c, []byte{1}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				if _, err := ListBySource(c, nil); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
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
			defer c.Close()
			tt.run(t, c)
		})
	}
}
