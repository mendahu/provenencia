package artifacts

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ingest"
	"github.com/mendahu/provenencia/core/ref"
)

func TestArtifacts(t *testing.T) {
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
		s, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID,
			Title:        "Family Bible",
		})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	mustFile := func(t *testing.T, c *database.Catalog, name string, data []byte) files.File {
		t.Helper()
		path := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		res, err := ingest.File(c, path, userID)
		if err != nil {
			t.Fatal(err)
		}
		return res.File
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
	latestChangesJSON := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var raw string
		if err := db.QueryRow(`
			SELECT c.changes_json FROM audit_changes c
			JOIN audit_transactions t ON t.id = c.audit_transaction_id
			ORDER BY t.revision DESC LIMIT 1`).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		return raw
	}
	artifactCount := func(t *testing.T, c *database.Catalog) int {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM artifacts`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create fileless and get by ref",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				a, err := Create(c, userID, CreateInput{
					SourceID:    src.ID,
					Label:       "Physical copy",
					Description: "held by Mary",
				})
				if err != nil {
					t.Fatal(err)
				}
				if !ref.Valid(a.Ref) || a.Ref[:4] != "ART-" {
					t.Fatalf("ref %q", a.Ref)
				}
				if a.FileID != nil {
					t.Fatalf("expected nil file_id %+v", a)
				}
				if a.Label != "Physical copy" {
					t.Fatalf("label %q", a.Label)
				}
				if latestAction(t, c) != "create_artifact" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				got, err := Get(c, a.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Label != "Physical copy" || got.Description != "held by Mary" || got.FileID != nil {
					t.Fatalf("got %+v", got)
				}
				byRef, err := getByRef(c, a.Ref)
				if err != nil || string(byRef.ID) != string(a.ID) {
					t.Fatalf("%v %+v", err, byRef)
				}
			},
		},
		{
			name: "create with file_id",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				f := mustFile(t, c, "scan.jpg", []byte("jpeg-bytes"))
				a, err := Create(c, userID, CreateInput{SourceID: src.ID, FileID: f.ID, Label: "Scan"})
				if err != nil {
					t.Fatal(err)
				}
				if string(a.FileID) != string(f.ID) {
					t.Fatalf("file_id %+v", a)
				}
				got, err := Get(c, a.ID)
				if err != nil || string(got.FileID) != string(f.ID) {
					t.Fatalf("%v %+v", err, got)
				}
			},
		},
		{
			name: "reject empty label",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				if _, err := Create(c, userID, CreateInput{SourceID: src.ID, Label: "  "}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				a, err := Create(c, userID, CreateInput{SourceID: src.ID, Label: "Keep"})
				if err != nil {
					t.Fatal(err)
				}
				a.Label = ""
				if err := Update(c, userID, a); !errors.Is(err, ErrInvalid) {
					t.Fatalf("update empty label %v", err)
				}
			},
		},
		{
			name: "attach first file then reject replace",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				a, err := Create(c, userID, CreateInput{SourceID: src.ID, Label: "Photo", Description: "photo"})
				if err != nil {
					t.Fatal(err)
				}
				f1 := mustFile(t, c, "a.bin", []byte("file-one"))
				a.FileID = f1.ID
				if err := Update(c, userID, a); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_artifact" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				var diffs map[string]struct {
					Old any `json:"old"`
					New any `json:"new"`
				}
				if err := json.Unmarshal([]byte(latestChangesJSON(t, c)), &diffs); err != nil {
					t.Fatal(err)
				}
				if diffs["file_id"].Old != nil || diffs["file_id"].New == nil {
					t.Fatalf("attach diff %+v", diffs["file_id"])
				}

				f2 := mustFile(t, c, "b.bin", []byte("file-two"))
				a.FileID = f2.ID
				if err := Update(c, userID, a); !errors.Is(err, ErrFileAlreadyAttached) {
					t.Fatalf("got %v", err)
				}
				got, err := Get(c, a.ID)
				if err != nil || string(got.FileID) != string(f1.ID) {
					t.Fatalf("%v %+v", err, got)
				}
				if _, err := files.Lookup(c, f1.ID); err != nil {
					t.Fatalf("old file retained: %v", err)
				}
				if _, err := files.Lookup(c, f2.ID); err != nil {
					t.Fatalf("new file: %v", err)
				}
			},
		},
		{
			name: "reject clearing file_id",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				f := mustFile(t, c, "x.bin", []byte("x"))
				a, err := Create(c, userID, CreateInput{SourceID: src.ID, FileID: f.ID, Label: "X"})
				if err != nil {
					t.Fatal(err)
				}
				a.FileID = nil
				if err := Update(c, userID, a); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "reject unknown source file and user",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				missing := make([]byte, 16)
				missing[15] = 9
				if _, err := Create(c, userID, CreateInput{SourceID: missing, Label: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if _, err := Create(c, userID, CreateInput{SourceID: src.ID, FileID: missing, Label: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad file %v", err)
				}
				if _, err := Create(c, nil, CreateInput{SourceID: src.ID, Label: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("nil user %v", err)
				}
				if _, err := Create(c, []byte{1}, CreateInput{SourceID: src.ID, Label: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("short user %v", err)
				}
			},
		},
		{
			name: "list by source only",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "photo", Origin: sourcetypes.OriginProvenencia, Label: "Photo",
				})
				if err != nil {
					t.Fatal(err)
				}
				s1, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "One"})
				if err != nil {
					t.Fatal(err)
				}
				s2, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Two"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Create(c, userID, CreateInput{SourceID: s1.ID, Label: "a", Description: "a"}); err != nil {
					t.Fatal(err)
				}
				if _, err := Create(c, userID, CreateInput{SourceID: s1.ID, Label: "b", Description: "b"}); err != nil {
					t.Fatal(err)
				}
				if _, err := Create(c, userID, CreateInput{SourceID: s2.ID, Label: "c", Description: "c"}); err != nil {
					t.Fatal(err)
				}
				list, err := ListBySource(c, s1.ID)
				if err != nil || len(list) != 2 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				for _, a := range list {
					if string(a.SourceID) != string(s1.ID) {
						t.Fatalf("wrong source %+v", a)
					}
				}
			},
		},
		{
			name: "source delete refused while artifact exists",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				f := mustFile(t, c, "keep.bin", []byte("keep"))
				if _, err := Create(c, userID, CreateInput{SourceID: src.ID, FileID: f.ID, Label: "Keep"}); err != nil {
					t.Fatal(err)
				}
				if artifactCount(t, c) != 1 {
					t.Fatalf("count %d", artifactCount(t, c))
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM sources WHERE id = ?`, src.ID); err == nil {
					t.Fatal("expected source delete to fail while an artifact remains")
				}
				if artifactCount(t, c) != 1 {
					t.Fatalf("artifact count %d", artifactCount(t, c))
				}
				if _, err := files.Lookup(c, f.ID); err != nil {
					t.Fatalf("file retained: %v", err)
				}
			},
		},
		{
			name: "artifact without citation deletes file remains",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				f := mustFile(t, c, "scan.bin", []byte("scan"))
				a, err := Create(c, userID, CreateInput{SourceID: src.ID, FileID: f.ID, Label: "Scan"})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM artifacts WHERE id = ?`, a.ID); err != nil {
					t.Fatal(err)
				}
				if artifactCount(t, c) != 0 {
					t.Fatalf("count %d", artifactCount(t, c))
				}
				if _, err := files.Lookup(c, f.ID); err != nil {
					t.Fatalf("file retained: %v", err)
				}
			},
		},
		{
			name: "update label and description",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				a, err := Create(c, userID, CreateInput{SourceID: src.ID, Label: "Old label", Description: "old"})
				if err != nil {
					t.Fatal(err)
				}
				a.Label = "New label"
				a.Description = "new"
				if err := Update(c, userID, a); err != nil {
					t.Fatal(err)
				}
				got, err := Get(c, a.ID)
				if err != nil || got.Label != "New label" || got.Description != "new" {
					t.Fatalf("%v %+v", err, got)
				}
			},
		},
		{
			name: "get missing returns ErrNoRows",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Get(c, make([]byte, 16))
				if !errors.Is(err, sql.ErrNoRows) {
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
