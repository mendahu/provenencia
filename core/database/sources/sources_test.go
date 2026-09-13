package sources

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/ref"
)

func TestSources(t *testing.T) {
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
	mustType := func(t *testing.T, c *database.Catalog) []byte {
		t.Helper()
		id, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		return id
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

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "title column is not null at current format",
			run: func(t *testing.T, c *database.Catalog) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 11 {
					t.Fatalf("user_version %d want >= 12", ver)
				}
				rows, err := db.Query(`PRAGMA table_info(sources)`)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				found := false
				for rows.Next() {
					var cid, notnull, pk int
					var name, ctype string
					var dflt any
					if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
						t.Fatal(err)
					}
					if name == "title" {
						found = true
						if notnull != 1 {
							t.Fatalf("title notnull=%d want 1", notnull)
						}
					}
				}
				if err := rows.Err(); err != nil {
					t.Fatal(err)
				}
				if !found {
					t.Fatal("title column missing")
				}
			},
		},
		{
			name: "create get and get by ref",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := Create(c, userID, CreateInput{
					SourceTypeID: typeID,
					Title:        "Family History",
					Description:  "A monograph",
				})
				if err != nil {
					t.Fatal(err)
				}
				if !ref.Valid(s.Ref) || s.Ref[:4] != "SRC-" {
					t.Fatalf("ref %q", s.Ref)
				}
				if latestAction(t, c) != "create_source" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				got, err := Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Title != "Family History" || got.Description != "A monograph" || got.Ref != s.Ref {
					t.Fatalf("got %+v", got)
				}
				byRef, err := GetByRef(c, s.Ref)
				if err != nil {
					t.Fatal(err)
				}
				if string(byRef.ID) != string(s.ID) {
					t.Fatal("id mismatch")
				}
			},
		},
		{
			name: "update records changed fields",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "Old"})
				if err != nil {
					t.Fatal(err)
				}
				s.Title = "New"
				s.Description = "Desc"
				if err := Update(c, userID, s); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_source" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				got, err := Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Title != "New" || got.Description != "Desc" {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "list multiple",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				if _, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "B"}); err != nil {
					t.Fatal(err)
				}
				if _, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "A"}); err != nil {
					t.Fatal(err)
				}
				all, err := List(c)
				if err != nil || len(all) != 2 {
					t.Fatalf("list %v %d", err, len(all))
				}
				if all[0].Title != "A" || all[1].Title != "B" {
					t.Fatalf("order %+v %+v", all[0], all[1])
				}
			},
		},
		{
			name: "notes add update delete list",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "Photo"})
				if err != nil {
					t.Fatal(err)
				}
				n, err := AddNote(c, userID, s.ID, "First thought")
				if err != nil {
					t.Fatal(err)
				}
				if n.AuthorDisplayName != "Jake" || n.CreatedAt == "" {
					t.Fatalf("add attribution %+v", n)
				}
				if latestAction(t, c) != "create_source_note" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				if err := UpdateNote(c, userID, n.ID, "Revised"); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_source_note" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				notes, err := ListNotes(c, s.ID)
				if err != nil || len(notes) != 1 || notes[0].Body != "Revised" {
					t.Fatalf("notes %v %+v", err, notes)
				}
				if notes[0].AuthorDisplayName != "Jake" || notes[0].CreatedAt == "" {
					t.Fatalf("list attribution %+v", notes[0])
				}
				got, err := GetNote(c, n.ID)
				if err != nil || got.Body != "Revised" || got.AuthorDisplayName != "Jake" {
					t.Fatalf("get %v %+v", err, got)
				}
				if err := DeleteNote(c, userID, n.ID); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_source_note" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				notes, err = ListNotes(c, s.ID)
				if err != nil || len(notes) != 0 {
					t.Fatalf("after delete %v %d", err, len(notes))
				}
			},
		},
		{
			name: "note cascades when source deleted",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "X"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := AddNote(c, userID, s.ID, "Keep"); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM sources WHERE id = ?`, s.ID); err != nil {
					t.Fatal(err)
				}
				notes, err := ListNotes(c, s.ID)
				if err != nil || len(notes) != 0 {
					t.Fatalf("cascade %v %d", err, len(notes))
				}
			},
		},
		{
			name: "count reports how many sources exist",
			run: func(t *testing.T, c *database.Catalog) {
				n, err := Count(c)
				if err != nil || n != 0 {
					t.Fatalf("empty %d %v", n, err)
				}
				mustUser(t, c)
				typeID := mustType(t, c)
				for _, title := range []string{"One", "Two", "Three"} {
					if _, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: title}); err != nil {
						t.Fatal(err)
					}
				}
				n, err = Count(c)
				if err != nil || n != 3 {
					t.Fatalf("got %d %v", n, err)
				}
			},
		},
		{
			name: "rejects bad type and blank title or note",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				badType := make([]byte, 16)
				badType[0] = 1
				if _, err := Create(c, userID, CreateInput{SourceTypeID: badType, Title: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad type %v", err)
				}
				typeID := mustType(t, c)
				if _, err := Create(c, userID, CreateInput{SourceTypeID: typeID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("blank title %v", err)
				}
				if _, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "  "}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("whitespace title %v", err)
				}
				s, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "Photo"})
				if err != nil {
					t.Fatal(err)
				}
				s.Title = ""
				if err := Update(c, userID, s); !errors.Is(err, ErrInvalid) {
					t.Fatalf("clear title %v", err)
				}
				if _, err := AddNote(c, userID, s.ID, "  "); !errors.Is(err, ErrInvalid) {
					t.Fatalf("blank note %v", err)
				}
			},
		},
		{
			name: "get by invalid ref",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := GetByRef(c, "not-a-ref"); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "closed catalog",
			run: func(t *testing.T, c *database.Catalog) {
				_ = c.Close()
				if _, err := List(c); !errors.Is(err, database.ErrClosed) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "missing get is no rows",
			run: func(t *testing.T, c *database.Catalog) {
				id := make([]byte, 16)
				id[0] = 9
				_, err := Get(c, id)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "cover columns default type_icon",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "Cover default"})
				if err != nil {
					t.Fatal(err)
				}
				if s.CoverMode != CoverModeTypeIcon {
					t.Fatalf("cover_mode %q", s.CoverMode)
				}
				if len(s.PrimaryArtifactID) != 0 {
					t.Fatalf("primary %v", s.PrimaryArtifactID)
				}
				got, err := Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.CoverMode != CoverModeTypeIcon || len(got.PrimaryArtifactID) != 0 {
					t.Fatalf("get cover %+v", got)
				}
			},
		},
		{
			name: "set cover artifact and revert to type icon",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "Pin cover"})
				if err != nil {
					t.Fatal(err)
				}
				path := t.TempDir() + "/scan.jpg"
				if err := os.WriteFile(path, []byte("not-really-jpeg"), 0o644); err != nil {
					t.Fatal(err)
				}
				fres, err := ingest.File(c, path, userID)
				if err != nil {
					t.Fatal(err)
				}
				a, err := artifacts.Create(c, userID, artifacts.CreateInput{
					SourceID: s.ID, FileID: fres.File.ID, Label: "Scan",
				})
				if err != nil {
					t.Fatal(err)
				}
				pinned, err := SetCover(c, userID, s.ID, CoverModeArtifact, a.ID)
				if err != nil {
					t.Fatal(err)
				}
				if pinned.CoverMode != CoverModeArtifact || !bytesEqual(pinned.PrimaryArtifactID, a.ID) {
					t.Fatalf("pinned %+v", pinned)
				}
				if latestAction(t, c) != "set_source_cover" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				fileless, err := artifacts.Create(c, userID, artifacts.CreateInput{
					SourceID: s.ID, Label: "Note only",
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := SetCover(c, userID, s.ID, CoverModeArtifact, fileless.ID); !errors.Is(err, ErrInvalid) {
					t.Fatalf("fileless pin %v", err)
				}
				reverted, err := SetCover(c, userID, s.ID, CoverModeTypeIcon, nil)
				if err != nil {
					t.Fatal(err)
				}
				if reverted.CoverMode != CoverModeTypeIcon || len(reverted.PrimaryArtifactID) != 0 {
					t.Fatalf("reverted %+v", reverted)
				}
			},
		},
		{
			name: "maybe pin first file cover once",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := Create(c, userID, CreateInput{SourceTypeID: typeID, Title: "Auto pin"})
				if err != nil {
					t.Fatal(err)
				}
				path := t.TempDir() + "/a.pdf"
				if err := os.WriteFile(path, []byte("%PDF"), 0o644); err != nil {
					t.Fatal(err)
				}
				f1, err := ingest.File(c, path, userID)
				if err != nil {
					t.Fatal(err)
				}
				a1, err := artifacts.Create(c, userID, artifacts.CreateInput{
					SourceID: s.ID, FileID: f1.File.ID, Label: "First",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, pinned, err := MaybePinFirstFileCover(c, userID, s.ID, a1.ID)
				if err != nil || !pinned {
					t.Fatalf("first pin got=%+v pinned=%v err=%v", got, pinned, err)
				}
				if got.CoverMode != CoverModeArtifact || !bytesEqual(got.PrimaryArtifactID, a1.ID) {
					t.Fatalf("got %+v", got)
				}
				path2 := t.TempDir() + "/b.pdf"
				if err := os.WriteFile(path2, []byte("%PDF-2"), 0o644); err != nil {
					t.Fatal(err)
				}
				f2, err := ingest.File(c, path2, userID)
				if err != nil {
					t.Fatal(err)
				}
				a2, err := artifacts.Create(c, userID, artifacts.CreateInput{
					SourceID: s.ID, FileID: f2.File.ID, Label: "Second",
				})
				if err != nil {
					t.Fatal(err)
				}
				again, pinned2, err := MaybePinFirstFileCover(c, userID, s.ID, a2.ID)
				if err != nil || pinned2 {
					t.Fatalf("second pin again=%+v pinned=%v err=%v", again, pinned2, err)
				}
				if !bytesEqual(again.PrimaryArtifactID, a1.ID) {
					t.Fatalf("stole cover %+v", again)
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
