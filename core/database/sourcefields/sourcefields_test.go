package sourcefields

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
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
					Key: "author", Origin: OriginProvenencia, Label: "Author", DataType: DataTypeDate,
				}); err != nil {
					t.Fatal(err)
				}
				got, err = Lookup(c, "author", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if got.DataType != DataTypeDate {
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

func TestCreateUpdateGetByID(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create mints slug key",
			run: func(t *testing.T, c *database.Catalog) {
				f, err := Create(c, "Grandma's album code", DataTypeText, "")
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
				if _, err := Create(c, "...", DataTypeText, ""); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create rejects duplicate key under user origin",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Create(c, "Album code", DataTypeText, ""); err != nil {
					t.Fatal(err)
				}
				_, err := Create(c, "Album code", DataTypeText, "")
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
				f, err := Create(c, "Photographer", DataTypeText, "")
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
				created, err := Create(c, "Album code", DataTypeText, "old")
				if err != nil {
					t.Fatal(err)
				}
				updated, err := Update(c, created.ID, "Album Code", DataTypeText, "new")
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
				created, err := Create(c, "Album code", DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Update(c, created.ID, "Album code", DataTypeDate, ""); !errors.Is(err, ErrInvalid) {
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
				updated, err := Update(c, id, "Author renamed", DataTypeText, "edited")
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
				if _, err := Update(c, id, "Memorial", DataTypeText, ""); !errors.Is(err, ErrLocked) {
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
			name: "delete unused provenencia ok",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Field{
					Key: "author", Origin: OriginProvenencia, Label: "Author", DataType: DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := Delete(c, id); err != nil {
					t.Fatal(err)
				}
				_, err = Lookup(c, "author", OriginProvenencia)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "delete in use refuses",
			run: func(t *testing.T, c *database.Catalog) {
				userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
				r, err := ref.Mint(ref.PrefixUser)
				if err != nil {
					t.Fatal(err)
				}
				if err := users.Upsert(c, userID, "Jake", r); err != nil {
					t.Fatal(err)
				}
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: sourcetypes.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "T"})
				if err != nil {
					t.Fatal(err)
				}
				field, err := Create(c, "Folio", DataTypeText, "")
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
					`INSERT INTO source_metadata (id, source_id, field_id, value_text, date_value_id) VALUES (?, ?, ?, ?, NULL)`,
					metaID[:], src.ID, field.ID, "12",
				); err != nil {
					t.Fatal(err)
				}
				if err := Delete(c, field.ID); !errors.Is(err, ErrInUse) {
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
