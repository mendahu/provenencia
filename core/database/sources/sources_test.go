package sources

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/derivatives"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func runCreate(c *database.Catalog, userID []byte, in CreateInput) (Source, error) {
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID},
		func(tx *database.Tx) (Source, []rowchange.Change, error) {
			return Create(tx, userID, in)
		})
	return src, err
}

func runUpdate(c *database.Catalog, userID []byte, s Source) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_source", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Update(tx, userID, s)
			return struct{}{}, changes, err
		})
	return err
}

func runDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_source", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}

func runSetCover(c *database.Catalog, userID, sourceID []byte, mode string, primaryArtifactID []byte) (Source, error) {
	if mode == CoverModeArtifact {
		// The thumbnail is made before the cover is chosen (SetSourceCover does
		// it outside the session). A File that gets none is NormalizeCover's to refuse.
		if fileID, err := CoverArtifactFile(c, sourceID, primaryArtifactID); err == nil {
			_, _ = derivatives.EnsureThumbnail(c, fileID)
		}
	}
	mode, artID, err := NormalizeCover(c, sourceID, mode, primaryArtifactID)
	if err != nil {
		return Source{}, err
	}
	src, _, err := writes.Run(c, writes.Op{Action: "set_source_cover", UserID: userID},
		func(tx *database.Tx) (Source, []rowchange.Change, error) {
			return SetCover(tx, userID, sourceID, mode, artID)
		})
	return src, err
}

func runAddNote(c *database.Catalog, userID, sourceID []byte, body string) (Note, error) {
	n, _, err := writes.Run(c, writes.Op{Action: "create_source_note", UserID: userID},
		func(tx *database.Tx) (Note, []rowchange.Change, error) {
			return AddNote(tx, userID, sourceID, body)
		})
	if err != nil {
		return Note{}, err
	}
	author, err := users.Lookup(c, userID)
	if err != nil {
		return Note{}, err
	}
	n.AuthorDisplayName = author.DisplayName
	return n, nil
}

func runUpdateNote(c *database.Catalog, userID, noteID []byte, body string) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_source_note", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := UpdateNote(tx, userID, noteID, body)
			return struct{}{}, changes, err
		})
	return err
}

func runDeleteNote(c *database.Catalog, userID, noteID []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_source_note", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := DeleteNote(tx, userID, noteID)
			return struct{}{}, changes, err
		})
	return err
}

func writeTinyPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

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
			Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
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
			name: "create rejects empty user id",
			run: func(t *testing.T, c *database.Catalog) {
				typeID := mustType(t, c)
				_, err := runCreate(c, nil, CreateInput{SourceTypeID: typeID, Title: "No user"})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v want ErrInvalid", err)
				}
			},
		},
		{
			name: "create get and get by ref",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := runCreate(c, userID, CreateInput{
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
				byRef, err := getByRef(c, s.Ref)
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
				s, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Old"})
				if err != nil {
					t.Fatal(err)
				}
				s.Title = "New"
				s.Description = "Desc"
				if err := runUpdate(c, userID, s); err != nil {
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
			name: "list and attach read the activity cache",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				src, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "One"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := List(c); err != nil {
					t.Fatal(err)
				}
				_, res, err := writes.Run(c, writes.Op{Action: "update_source", UserID: userID},
					func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						changes, err := Update(tx, userID, Source{ID: src.ID, Ref: src.Ref, SourceTypeID: typeID, Title: "Edited"})
						return struct{}{}, changes, err
					})
				if err != nil || res.Revision == 0 || res.At == "" {
					t.Fatalf("write %+v %v", res, err)
				}
				want := res.Activity
				var queries int
				c.Graph().SetQueryCounterForTest(func() { queries++ })
				all, err := List(c)
				if err != nil || len(all) != 1 || all[0].Updated != want {
					t.Fatalf("List %+v %v, want %+v", all, err, want)
				}
				attached := Source{ID: src.ID}
				if err := AttachActivity(c, &attached); err != nil || attached.Updated != want {
					t.Fatalf("Attach %+v %v, want %+v", attached.Updated, err, want)
				}
				if queries != 0 {
					t.Fatalf("%d scope queries after a write; the cache should have been patched", queries)
				}
				c.Graph().Drop()
				if _, err := List(c); err != nil || queries != 1 {
					t.Fatalf("after Drop: %v, %d queries, want 1", err, queries)
				}
			},
		},
		{
			name: "list newest first by id with updated revision",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				first, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Older"})
				if err != nil {
					t.Fatal(err)
				}
				second, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Newer"})
				if err != nil {
					t.Fatal(err)
				}
				all, err := List(c)
				if err != nil || len(all) != 2 {
					t.Fatalf("list %v %d", err, len(all))
				}
				if all[0].Title != "Newer" || all[1].Title != "Older" {
					t.Fatalf("want newest-first got %+v %+v", all[0], all[1])
				}
				if all[0].Updated.Revision <= 0 || all[1].Updated.Revision <= 0 {
					t.Fatalf("want create revisions got %d %d", all[0].Updated.Revision, all[1].Updated.Revision)
				}
				if all[0].Updated.Revision <= all[1].Updated.Revision {
					t.Fatalf("newer create should have higher revision: %+v", all)
				}
				if err := runUpdate(c, userID, Source{
					ID: first.ID, Ref: first.Ref, SourceTypeID: typeID,
					Title: "Older edited", Description: first.Description,
				}); err != nil {
					t.Fatal(err)
				}
				all, err = List(c)
				if err != nil || len(all) != 2 {
					t.Fatalf("list after update %v %d", err, len(all))
				}
				var older, newer Source
				for _, s := range all {
					switch s.Title {
					case "Older edited":
						older = s
					case "Newer":
						newer = s
					}
				}
				if older.ID == nil || newer.ID == nil {
					t.Fatalf("missing rows %+v", all)
				}
				if older.Updated.Revision <= newer.Updated.Revision {
					t.Fatalf("updated source should lead on revision: older=%d newer=%d", older.Updated.Revision, newer.Updated.Revision)
				}
				// Default list order is still create-time (id DESC), not updated.
				if !bytes.Equal(all[0].ID, second.ID) || !bytes.Equal(all[1].ID, first.ID) {
					t.Fatalf("list order should stay newest-created-first got %+v %+v", all[0].Title, all[1].Title)
				}
				if all[0].HasArtifact || all[1].HasArtifact {
					t.Fatal("want has_artifact false with no artifacts")
				}
			},
		},
		{
			name: "list multiple",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				if _, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "B"}); err != nil {
					t.Fatal(err)
				}
				if _, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "A"}); err != nil {
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
				s, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Photo"})
				if err != nil {
					t.Fatal(err)
				}
				n, err := runAddNote(c, userID, s.ID, "First thought")
				if err != nil {
					t.Fatal(err)
				}
				if n.AuthorDisplayName != "Jake" || n.CreatedAt == "" {
					t.Fatalf("add attribution %+v", n)
				}
				if latestAction(t, c) != "create_source_note" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				if err := runUpdateNote(c, userID, n.ID, "Revised"); err != nil {
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
				if err := runDeleteNote(c, userID, n.ID); err != nil {
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
				s, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "X"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runAddNote(c, userID, s.ID, "Keep"); err != nil {
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
					if _, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: title}); err != nil {
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
				if _, err := runCreate(c, userID, CreateInput{SourceTypeID: badType, Title: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad type %v", err)
				}
				typeID := mustType(t, c)
				if _, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("blank title %v", err)
				}
				if _, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "  "}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("whitespace title %v", err)
				}
				s, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Photo"})
				if err != nil {
					t.Fatal(err)
				}
				s.Title = ""
				if err := runUpdate(c, userID, s); !errors.Is(err, ErrInvalid) {
					t.Fatalf("clear title %v", err)
				}
				if _, err := runAddNote(c, userID, s.ID, "  "); !errors.Is(err, ErrInvalid) {
					t.Fatalf("blank note %v", err)
				}
			},
		},
		{
			name: "get by invalid ref",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := getByRef(c, "not-a-ref"); !errors.Is(err, ErrInvalid) {
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
			name: "missing get is not found",
			run: func(t *testing.T, c *database.Catalog) {
				id := make([]byte, 16)
				id[0] = 9
				_, err := Get(c, id)
				if !errors.Is(err, ErrNotFound) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "cover columns default type_icon",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Cover default"})
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
				s, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Pin cover"})
				if err != nil {
					t.Fatal(err)
				}
				path := t.TempDir() + "/scan.png"
				writeTinyPNG(t, path)
				fres, err := ingest.File(c, path, userID)
				if err != nil {
					t.Fatal(err)
				}
				a, err := runArtifactCreate(c, userID, artifacts.CreateInput{
					SourceID: s.ID, FileID: fres.File.ID, Label: "Scan",
				})
				if err != nil {
					t.Fatal(err)
				}
				pinned, err := runSetCover(c, userID, s.ID, CoverModeArtifact, a.ID)
				if err != nil {
					t.Fatal(err)
				}
				if pinned.CoverMode != CoverModeArtifact || !bytes.Equal(pinned.PrimaryArtifactID, a.ID) {
					t.Fatalf("pinned %+v", pinned)
				}
				if latestAction(t, c) != "set_source_cover" {
					t.Fatalf("action %s", latestAction(t, c))
				}
				fileless, err := runArtifactCreate(c, userID, artifacts.CreateInput{
					SourceID: s.ID, Label: "Note only",
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runSetCover(c, userID, s.ID, CoverModeArtifact, fileless.ID); !errors.Is(err, ErrInvalid) {
					t.Fatalf("fileless pin %v", err)
				}
				pdfPath := t.TempDir() + "/deed.pdf"
				if err := os.WriteFile(pdfPath, []byte("%PDF-1.1\n%%EOF\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				pdfFile, err := ingest.File(c, pdfPath, userID)
				if err != nil {
					t.Fatal(err)
				}
				pdfArt, err := runArtifactCreate(c, userID, artifacts.CreateInput{
					SourceID: s.ID, FileID: pdfFile.File.ID, Label: "PDF",
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runSetCover(c, userID, s.ID, CoverModeArtifact, pdfArt.ID); !errors.Is(err, ErrInvalid) {
					t.Fatalf("pdf pin %v", err)
				}
				reverted, err := runSetCover(c, userID, s.ID, CoverModeTypeIcon, nil)
				if err != nil {
					t.Fatal(err)
				}
				if reverted.CoverMode != CoverModeTypeIcon || len(reverted.PrimaryArtifactID) != 0 {
					t.Fatalf("reverted %+v", reverted)
				}
			},
		},
		{
			name: "ingest leaves cover as type icon until explicit pin",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID := mustType(t, c)
				s, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "No auto pin"})
				if err != nil {
					t.Fatal(err)
				}
				path := t.TempDir() + "/a.png"
				writeTinyPNG(t, path)
				f1, err := ingest.File(c, path, userID)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runArtifactCreate(c, userID, artifacts.CreateInput{
					SourceID: s.ID, FileID: f1.File.ID, Label: "First",
				}); err != nil {
					t.Fatal(err)
				}
				got, err := Get(c, s.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.CoverMode != CoverModeTypeIcon || len(got.PrimaryArtifactID) != 0 {
					t.Fatalf("want type_icon after ingest, got %+v", got)
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

func TestSourceDelete(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}

	empty, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Notes only"})
	if err != nil {
		t.Fatal(err)
	}
	note, err := runAddNote(c, userID, empty.ID, "Film is faded")
	if err != nil {
		t.Fatal(err)
	}
	if err := runDelete(c, userID, empty.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := Get(c, empty.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("source left %v", err)
	}
	if _, err := GetNote(c, note.ID); !errors.Is(err, ErrInvalid) {
		t.Fatalf("note left %v", err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var action string
	if err := db.QueryRow(`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`).Scan(&action); err != nil {
		t.Fatal(err)
	}
	if action != "delete_source" {
		t.Fatalf("audit %q", action)
	}
	var docs int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM catalog_search_docs WHERE kind = 'source' AND entity_id = ?`,
		database.UUIDString(empty.ID),
	).Scan(&docs); err != nil {
		t.Fatal(err)
	}
	if docs != 0 {
		t.Fatalf("search docs left=%d", docs)
	}

	blocked, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Has artifact"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: blocked.ID, Label: "Scan"}); err != nil {
		t.Fatal(err)
	}
	if err := runDelete(c, userID, blocked.ID); !errors.Is(err, ErrInUse) {
		t.Fatalf("artifact inbound %v", err)
	}

	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	uncited, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Has subject"})
	if err != nil {
		t.Fatal(err)
	}
	graphs, err := evidencegraphs.List(c, uncited.ID)
	if err != nil || len(graphs) != 1 {
		t.Fatalf("source %x has %d graphs, want 1", uncited.ID, len(graphs))
	}
	subjectstest.OfType(t, c, userID, graphs[0].ID, personType.ID, "Alice")
	if err := runDelete(c, userID, uncited.ID); !errors.Is(err, ErrInUse) {
		t.Fatalf("subject inbound %v", err)
	}

	t.Run("releases the Default graph", func(t *testing.T) {
		graphed, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Graph only"})
		if err != nil {
			t.Fatal(err)
		}
		var got []rowchange.Change
		if _, err := writes.Call(c, writes.Op{Action: "delete_source", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, graphed.ID)
			got = changes
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM evidence_graphs WHERE source_id = ?`, graphed.ID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("graph rows left=%d", n)
		}
		var types []string
		for _, ch := range got {
			types = append(types, ch.EntityType+" "+string(ch.Action))
		}
		want := []string{"evidence_graph delete", "source delete"}
		if len(types) != len(want) {
			t.Fatalf("changes %v", types)
		}
		for i := range want {
			if types[i] != want[i] {
				t.Fatalf("changes %v want %v", types, want)
			}
		}
	})

	t.Run("dated source releases its date", func(t *testing.T) {
		dated, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Dated"})
		if err != nil {
			t.Fatal(err)
		}
		withDate, err := runSetDate(c, userID, dated.ID, datevaluetest.YMD(1880, 0, 0))
		if err != nil {
			t.Fatal(err)
		}
		var got []rowchange.Change
		if _, err := writes.Call(c, writes.Op{Action: "delete_source", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, dated.ID)
			got = changes
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM date_values WHERE id = ?`, withDate.DateValueID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("date rows left=%d", n)
		}
		last := got[len(got)-1]
		if last.EntityType != "source" || last.Fields["date_value_id"].Old != database.UUIDString(withDate.DateValueID) {
			t.Fatalf("delete change %+v", last)
		}
	})

	missing := make([]byte, 16)
	if err := runDelete(c, userID, missing); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing %v", err)
	}
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}

func TestNormalizeCoverNeedsThumbnail(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	mine, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Mine"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/scan.png"
	writeTinyPNG(t, path)
	fres, err := ingest.File(c, path, userID)
	if err != nil {
		t.Fatal(err)
	}
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: mine.ID, FileID: fres.File.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := NormalizeCover(c, mine.ID, CoverModeArtifact, art.ID); !errors.Is(err, ErrInvalid) {
		t.Fatalf("no thumbnail: err = %v, want ErrInvalid", err)
	}
	if _, err := derivatives.EnsureThumbnail(c, fres.File.ID); err != nil {
		t.Fatal(err)
	}
	mode, id, err := NormalizeCover(c, mine.ID, CoverModeArtifact, art.ID)
	if err != nil || mode != CoverModeArtifact || !bytes.Equal(id, art.ID) {
		t.Fatalf("with thumbnail: %q %x %v", mode, id, err)
	}
	if _, _, err := NormalizeCover(c, other.ID, CoverModeArtifact, art.ID); !errors.Is(err, ErrInvalid) {
		t.Fatalf("another Source's Artifact: err = %v, want ErrInvalid", err)
	}
}

func TestCoverArtifactID(t *testing.T) {
	id := bytes.Repeat([]byte{7}, 16)
	for _, tc := range []struct {
		name string
		mode string
		id   []byte
		want bool
	}{
		{"artifact with an id", CoverModeArtifact, id, true},
		{"mode is trimmed", " artifact ", id, true},
		{"artifact with no id", CoverModeArtifact, nil, false},
		{"artifact with a short id", CoverModeArtifact, []byte{1}, false},
		{"type icon", CoverModeTypeIcon, id, false},
		{"unknown mode", "pinned", id, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := CoverArtifactID(tc.mode, tc.id)
			if ok != tc.want {
				t.Fatalf("ok = %v, want %v", ok, tc.want)
			}
			if ok && !bytes.Equal(got, id) {
				t.Fatalf("id = %x, want %x", got, id)
			}
		})
	}
}

func runSetDate(c *database.Catalog, userID, sourceID []byte, d *datevalue.Value) (Source, error) {
	src, _, err := writes.Run(c, writes.Op{Action: "set_source_date", UserID: userID},
		func(tx *database.Tx) (Source, []rowchange.Change, error) {
			return SetDate(tx, userID, sourceID, d)
		})
	return src, err
}

func runSetPublic(c *database.Catalog, userID, sourceID []byte, isPublic bool) (Source, []rowchange.Change, error) {
	var changes []rowchange.Change
	src, _, err := writes.Run(c, writes.Op{Action: "set_source_public", UserID: userID},
		func(tx *database.Tx) (Source, []rowchange.Change, error) {
			s, ch, err := SetPublic(tx, userID, sourceID, isPublic)
			changes = ch
			return s, ch, err
		})
	return src, changes, err
}

// sourceFixture is a catalog with one user and one Source type.
func sourceFixture(t *testing.T) (c *database.Catalog, userID, typeID []byte) {
	t.Helper()
	userID = []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	typeID, err = sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, userID, typeID
}

func TestSourceDate(t *testing.T) {
	c, userID, typeID := sourceFixture(t)
	src, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Parish register"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Census"})
	if err != nil {
		t.Fatal(err)
	}
	if src.Date != nil || src.DateValueID != nil {
		t.Fatalf("new source is dated: %+v", src)
	}

	var (
		firstID []byte
		changes []rowchange.Change
	)
	set := func(d *datevalue.Value) (Source, error) {
		changes = nil
		s, _, err := writes.Run(c, writes.Op{Action: "set_source_date", UserID: userID},
			func(tx *database.Tx) (Source, []rowchange.Change, error) {
				s, ch, err := SetDate(tx, userID, src.ID, d)
				changes = ch
				return s, ch, err
			})
		return s, err
	}

	t.Run("year only", func(t *testing.T) {
		got, err := set(datevaluetest.YMD(1880, 0, 0))
		if err != nil {
			t.Fatal(err)
		}
		if got.Date == nil || got.Date.Kind != datevalue.KindPoint || got.Date.StartYear == nil || *got.Date.StartYear != 1880 || len(got.DateValueID) != 16 {
			t.Fatalf("got %+v", got)
		}
		firstID = got.DateValueID
		read, err := Get(c, src.ID)
		if err != nil || read.Date == nil || !bytes.Equal(read.Date.ID, firstID) {
			t.Fatalf("Get %+v err=%v", read, err)
		}
		if dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM date_values`) != 1 {
			t.Fatalf("date rows %d", dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM date_values`))
		}
		if len(changes) != 2 || changes[0].EntityType != "date_value" || changes[0].Action != rowchange.ActionCreate ||
			changes[1].EntityType != "source" || changes[1].Fields["date_value_id"].New != database.UUIDString(firstID) {
			t.Fatalf("changes %+v", changes)
		}
	})
	t.Run("edit in place", func(t *testing.T) {
		got, err := set(datevaluetest.About(1880, 3, 12))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got.DateValueID, firstID) || got.Date.Qualifier != datevalue.QualifierABT || *got.Date.StartDay != 12 {
			t.Fatalf("got %+v", got)
		}
		if len(changes) != 1 || changes[0].EntityType != "date_value" || changes[0].Action != rowchange.ActionUpdate {
			t.Fatalf("changes %+v", changes)
		}
		if dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM date_values`) != 1 {
			t.Fatalf("date rows %d", dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM date_values`))
		}
	})
	t.Run("same date writes nothing", func(t *testing.T) {
		if _, err := set(datevaluetest.About(1880, 3, 12)); err != nil {
			t.Fatal(err)
		}
		if len(changes) != 0 {
			t.Fatalf("changes %+v", changes)
		}
	})
	t.Run("invalid date keeps the stored one", func(t *testing.T) {
		bad := &datevalue.Value{Kind: datevalue.KindPoint, StartMonth: datevaluetest.Part(13)}
		if _, err := set(bad); !errors.Is(err, datevalue.ErrInvalid) {
			t.Fatalf("err %v", err)
		}
		read, err := Get(c, src.ID)
		if err != nil || read.Date == nil || *read.Date.StartYear != 1880 || !bytes.Equal(read.DateValueID, firstID) {
			t.Fatalf("Get %+v err=%v", read, err)
		}
	})
	t.Run("clear releases the row", func(t *testing.T) {
		got, err := set(nil)
		if err != nil {
			t.Fatal(err)
		}
		if got.Date != nil || got.DateValueID != nil || dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM date_values`) != 0 {
			t.Fatalf("got %+v, date rows %d", got, dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM date_values`))
		}
		if len(changes) != 2 || changes[0].EntityType != "source" || changes[0].Fields["date_value_id"].New != nil ||
			changes[1].EntityType != "date_value" || changes[1].Action != rowchange.ActionDelete {
			t.Fatalf("changes %+v", changes)
		}
		if _, err := set(nil); err != nil || len(changes) != 0 {
			t.Fatalf("clearing an undated source: err=%v changes=%+v", err, changes)
		}
	})
	t.Run("unknown source", func(t *testing.T) {
		_, err := runSetDate(c, userID, make([]byte, 16), datevaluetest.YMD(1880, 0, 0))
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("err %v", err)
		}
		if _, err := runSetDate(c, userID, []byte{1}, nil); !errors.Is(err, ErrInvalid) {
			t.Fatalf("short id err %v", err)
		}
	})
	t.Run("list carries each date", func(t *testing.T) {
		if _, err := set(datevaluetest.YMD(1881, 0, 0)); err != nil {
			t.Fatal(err)
		}
		all, err := List(c)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range all {
			switch {
			case bytes.Equal(s.ID, src.ID):
				if s.Date == nil || *s.Date.StartYear != 1881 || len(s.Date.ID) != 16 {
					t.Fatalf("dated source %+v", s)
				}
			case bytes.Equal(s.ID, other.ID):
				if s.Date != nil {
					t.Fatalf("undated source has %+v", s.Date)
				}
			}
		}
	})
	t.Run("dangling date row reads undated", func(t *testing.T) {
		dangle := func() []byte {
			ctx := context.Background()
			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			conn, err := db.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys=OFF`); err != nil {
				t.Fatal(err)
			}
			defer conn.ExecContext(ctx, `PRAGMA foreign_keys=ON`)
			id := make([]byte, 16)
			if _, err := rand.Read(id); err != nil {
				t.Fatal(err)
			}
			if _, err := conn.ExecContext(ctx, `UPDATE sources SET date_value_id = ? WHERE id = ?`, id, other.ID); err != nil {
				t.Fatal(err)
			}
			return id
		}
		danglingID := dangle()
		got, err := Get(c, other.ID)
		if err != nil || got.Date != nil || !bytes.Equal(got.DateValueID, danglingID) {
			t.Fatalf("Get %+v err=%v", got, err)
		}
		set, err := runSetDate(c, userID, other.ID, datevaluetest.YMD(1900, 0, 0))
		if err != nil || set.Date == nil || *set.Date.StartYear != 1900 || bytes.Equal(set.DateValueID, danglingID) {
			t.Fatalf("SetDate %+v err=%v", set, err)
		}
		dangle()
		cleared, err := runSetDate(c, userID, other.ID, nil)
		if err != nil || cleared.DateValueID != nil || cleared.Date != nil {
			t.Fatalf("clear %+v err=%v", cleared, err)
		}
	})
	t.Run("one date row per source", func(t *testing.T) {
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		cur, err := Get(c, src.ID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`UPDATE sources SET date_value_id = ? WHERE id = ?`, cur.DateValueID, other.ID); err == nil {
			t.Fatal("two sources shared one date row")
		}
	})
}

func TestSourcePublic(t *testing.T) {
	c, userID, typeID := sourceFixture(t)
	src, err := runCreate(c, userID, CreateInput{SourceTypeID: typeID, Title: "Parish register"})
	if err != nil {
		t.Fatal(err)
	}
	if !src.IsPublic {
		t.Fatal("Create returned a not-public source")
	}
	all, err := List(c)
	if err != nil || len(all) != 1 || !all[0].IsPublic {
		t.Fatalf("List %+v err=%v", all, err)
	}

	got, changes, err := runSetPublic(c, userID, src.ID, false)
	if err != nil || got.IsPublic || len(changes) != 1 || changes[0].Fields["is_public"].Old != 1 || changes[0].Fields["is_public"].New != 0 {
		t.Fatalf("set false: %+v changes=%+v err=%v", got, changes, err)
	}
	read, err := Get(c, src.ID)
	if err != nil || read.IsPublic {
		t.Fatalf("Get %+v err=%v", read, err)
	}
	if _, changes, err := runSetPublic(c, userID, src.ID, false); err != nil || len(changes) != 0 {
		t.Fatalf("set false again: changes=%+v err=%v", changes, err)
	}
	got, _, err = runSetPublic(c, userID, src.ID, true)
	if err != nil || !got.IsPublic {
		t.Fatalf("set true: %+v err=%v", got, err)
	}
	if _, _, err := runSetPublic(c, userID, make([]byte, 16), false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown source err=%v", err)
	}
}

func TestDateSortKey(t *testing.T) {
	ordered := []*datevalue.Value{
		datevaluetest.YMD(1879, 0, 0),
		datevaluetest.Before(1880, 0, 0),
		datevaluetest.YMD(1880, 0, 0),
		datevaluetest.YMD(1881, 0, 0),
		datevaluetest.Before(2011, 3, 29),
		datevaluetest.YMD(2011, 3, 29),
		datevaluetest.After(2011, 3, 29),
	}
	var prev int64
	for i, d := range ordered {
		k, ok := DateSortKey(Source{Date: d})
		if !ok || (i > 0 && k <= prev) {
			t.Fatalf("date %d key %d ok=%v after %d", i, k, ok, prev)
		}
		prev = k
	}
	about, _ := DateSortKey(Source{Date: datevaluetest.About(1880, 0, 0)})
	plain, _ := DateSortKey(Source{Date: datevaluetest.YMD(1880, 0, 0)})
	if about != plain {
		t.Fatalf("ABT 1880 key %d, 1880 key %d", about, plain)
	}
	if _, ok := DateSortKey(Source{}); ok {
		t.Fatal("undated source has a key")
	}
	if _, ok := DateSortKey(Source{Date: datevaluetest.Phrase("in the reign of Victoria")}); ok {
		t.Fatal("phrase-only date has a key")
	}
}

func TestMissingSourceIsNotFound(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	unused := make([]byte, 16)
	unused[15] = 9
	short := []byte{1, 2, 3}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		run  func(id []byte) error
	}{
		{"SetCover", func(id []byte) error {
			_, err := runSetCover(c, userID, id, CoverModeTypeIcon, nil)
			return err
		}},
		{"SetDate", func(id []byte) error { _, err := runSetDate(c, userID, id, nil); return err }},
		{"SetPublic", func(id []byte) error { _, _, err := runSetPublic(c, userID, id, false); return err }},
		{"Update", func(id []byte) error { return runUpdate(c, userID, Source{ID: id, SourceTypeID: unused, Title: "x"}) }},
		{"Delete", func(id []byte) error { return runDelete(c, userID, id) }},
		{"Get", func(id []byte) error { _, err := Get(c, id); return err }},
		{"GetTx", func(id []byte) error {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()
			_, err = GetTx(tx, id)
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(unused); !errors.Is(err, ErrNotFound) {
				t.Fatalf("unused id: %v, want ErrNotFound", err)
			}
			if err := tc.run(short); !errors.Is(err, ErrInvalid) {
				t.Fatalf("3-byte id: %v, want ErrInvalid", err)
			}
		})
	}
}
