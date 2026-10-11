package metadatafields

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

func TestUpsertLookupList(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "upsert lookup refresh data type",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "author", Origin: OriginProvenencia, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, "author", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.DataType != DataTypeText {
					t.Fatalf("got %+v", got)
				}
				if _, err := Upsert(c, Field{
					Key: "author", Origin: OriginProvenencia, Label: "Author", DataType: DataTypeURL,
				}); err != nil {
					t.Fatal(err)
				}
				got, err = Lookup(c, "author", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if got.DataType != DataTypeURL {
					t.Fatalf("data_type %q", got.DataType)
				}
			},
		},
		{
			name: "rejects bad data type",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Field{
					Key: "x", Origin: OriginUser, Label: "X", DataType: "number",
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "upsert accepts url data type",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "landing-page", Origin: OriginUser, Label: "Landing page", DataType: DataTypeURL,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, "landing-page", OriginUser)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.DataType != DataTypeURL {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "lookup missing",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Lookup(c, "nope", OriginProvenencia)
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

func TestDelete(t *testing.T) {
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

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
	latestAction := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actionType string
		if err := db.QueryRow(
			`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`,
		).Scan(&actionType); err != nil {
			t.Fatal(err)
		}
		return actionType
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "unused provenencia ok",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				id, err := Upsert(c, Field{
					Key: "author", Origin: OriginProvenencia, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_metadata_field" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				_, err = Lookup(c, "author", OriginProvenencia)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "in use refuses",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: sourcetypes.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "T"})
				})
				if err != nil {
					t.Fatal(err)
				}
				field, err := runCreate(c, "Folio", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				metaID, err := uuid.NewV7()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO source_metadata (id, source_id, field_id, value_text) VALUES (?, ?, ?, ?)`,
					metaID[:], src.ID, field.ID, "12",
				); err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, field.ID); !errors.Is(err, ErrInUse) {
					t.Fatalf("got %v", err)
				}
				if _, err := GetByID(c, field.ID); err != nil {
					t.Fatalf("field gone: %v", err)
				}
			},
		},
		{
			name: "unused plugin is origin locked",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				id, err := Upsert(c, Field{
					Key: "memorial_id", Origin: "plugin:findagrave", Label: "Memorial id", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); !errors.Is(err, ErrOriginLocked) {
					t.Fatalf("got %v", err)
				}
				if _, err := GetByID(c, id); err != nil {
					t.Fatalf("field gone: %v", err)
				}
			},
		},
		{
			name: "missing is invalid",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				missing := make([]byte, 16)
				missing[15] = 9
				if err := runDelete(c, userID, missing); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "unused field with suggestion and layout still erases",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: sourcetypes.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "T"})
				})
				if err != nil {
					t.Fatal(err)
				}
				id, err := Upsert(c, Field{
					Key: "author", Origin: OriginUser, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO source_type_metadata_fields (source_type_id, field_id, sort_order) VALUES (?, ?, 0)`,
					typeID, id,
				); err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO source_metadata_layout (source_id, field_id, sort_order) VALUES (?, ?, 0)`,
					src.ID, id,
				); err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); err != nil {
					t.Fatal(err)
				}
				if _, err := GetByID(c, id); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("field still there: %v", err)
				}
				var joins int
				if err := db.QueryRow(
					`SELECT COUNT(*) FROM source_type_metadata_fields WHERE field_id = ?`, id,
				).Scan(&joins); err != nil || joins != 0 {
					t.Fatalf("suggestion leftover %d %v", joins, err)
				}
				if err := db.QueryRow(
					`SELECT COUNT(*) FROM source_metadata_layout WHERE field_id = ?`, id,
				).Scan(&joins); err != nil || joins != 0 {
					t.Fatalf("layout leftover %d %v", joins, err)
				}
				if _, err := sourcetypes.GetByID(c, typeID); err != nil {
					t.Fatalf("type should survive: %v", err)
				}
			},
		},
		{
			name: "bad id",
			run: func(t *testing.T, c *database.Catalog) {
				if err := runDelete(c, userID, []byte{1}); !errors.Is(err, ErrInvalid) {
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

func TestCreateUpdateGetByID(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create mints slug key",
			run: func(t *testing.T, c *database.Catalog) {
				f, err := runCreate(c, "Grandma's album code", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if f.Key != "grandmas-album-code" || f.Origin != OriginUser {
					t.Fatalf("got %+v", f)
				}
			},
		},
		{
			name: "create rejects unslugifiable label",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := runCreate(c, "...", DataTypeText, ""); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create rejects duplicate key under user origin",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := runCreate(c, "Album code", DataTypeText, ""); err != nil {
					t.Fatal(err)
				}
				_, err := runCreate(c, "Album code", DataTypeText, "")
				if !errors.Is(err, ErrDuplicateKey) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "same key different origin does not collide",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Field{
					Key: "photographer", Origin: OriginProvenencia, Label: "Photographer", DataType: DataTypeText,
				}); err != nil {
					t.Fatal(err)
				}
				f, err := runCreate(c, "Photographer", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if f.Key != "photographer" || f.Origin != OriginUser {
					t.Fatalf("got %+v", f)
				}
			},
		},
		{
			name: "update patches label and description but not key or data type",
			run: func(t *testing.T, c *database.Catalog) {
				created, err := runCreate(c, "Album code", DataTypeText, "old")
				if err != nil {
					t.Fatal(err)
				}
				updated, err := runUpdate(c, created.ID, "Album Code", DataTypeText, "new")
				if err != nil {
					t.Fatal(err)
				}
				if updated.Key != created.Key || updated.Label != "Album Code" ||
					updated.DataType != DataTypeText || updated.Description != "new" {
					t.Fatalf("got %+v", updated)
				}
				got, err := GetByID(c, created.ID)
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != updated.Key || got.Label != updated.Label {
					t.Fatalf("got %+v want %+v", got, updated)
				}
			},
		},
		{
			name: "update rejects data type change",
			run: func(t *testing.T, c *database.Catalog) {
				created, err := runCreate(c, "Album code", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runUpdate(c, created.ID, "Album code", DataTypeURL, ""); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "update provenencia field",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "author", Origin: OriginProvenencia, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				updated, err := runUpdate(c, id, "Author renamed", DataTypeText, "edited")
				if err != nil {
					t.Fatal(err)
				}
				if updated.Key != "author" || updated.Origin != OriginProvenencia ||
					updated.Label != "Author renamed" || updated.DataType != DataTypeText ||
					updated.Description != "edited" {
					t.Fatalf("got %+v", updated)
				}
			},
		},
		{
			name: "update rejects plugin field",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "memorial_id", Origin: "plugin:findagrave", Label: "Memorial id", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runUpdate(c, id, "Memorial", DataTypeText, ""); !errors.Is(err, ErrLocked) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "get by id missing",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := GetByID(c, make([]byte, 16)); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "CountByOrigin splits seeded user and plugin",
			run: func(t *testing.T, c *database.Catalog) {
				got, err := CountByOrigin(c)
				if err != nil || got != (OriginCounts{}) {
					t.Fatalf("empty %+v %v", got, err)
				}
				if _, err := Upsert(c, Field{
					Key: "author", Origin: OriginProvenencia, Label: "Author", DataType: DataTypeText,
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := Upsert(c, Field{
					Key: "notes", Origin: OriginUser, Label: "Notes", DataType: DataTypeText,
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := Upsert(c, Field{
					Key: "memorial_id", Origin: "plugin:findagrave", Label: "Memorial", DataType: DataTypeText,
				}); err != nil {
					t.Fatal(err)
				}
				got, err = CountByOrigin(c)
				if err != nil {
					t.Fatal(err)
				}
				want := OriginCounts{Total: 3, Seeded: 1, User: 1, Plugin: 1}
				if got != want {
					t.Fatalf("got %+v want %+v", got, want)
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

func runCreate(c *database.Catalog, label, dataType, description string) (Field, error) {
	got, _, err := writes.Run(c, writes.Op{Action: "create_metadata_field"},
		func(tx *database.Tx) (Field, []rowchange.Change, error) {
			return Create(tx, label, dataType, description)
		})
	return got, err
}

func runUpdate(c *database.Catalog, id []byte, label, dataType, description string) (Field, error) {
	got, _, err := writes.Run(c, writes.Op{Action: "update_metadata_field"},
		func(tx *database.Tx) (Field, []rowchange.Change, error) {
			return Update(tx, id, label, dataType, description)
		})
	return got, err
}

func runDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_metadata_field", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}
