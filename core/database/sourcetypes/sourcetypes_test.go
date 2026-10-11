package sourcetypes

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
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
			name: "upsert lookup and list",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{
					Key: "photograph", Origin: OriginProvenencia, Label: "Photograph", Description: "Photo",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, "photograph", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.Label != "Photograph" || got.Description != "Photo" {
					t.Fatalf("got %+v", got)
				}
				id2, err := Upsert(c, Type{
					Key: "photograph", Origin: OriginProvenencia, Label: "Photo", Description: "Updated",
				})
				if err != nil {
					t.Fatal(err)
				}
				if string(id2) != string(id) {
					t.Fatal("id should be stable on upsert")
				}
				got, err = Lookup(c, "photograph", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if got.Label != "Photo" || got.Description != "Updated" {
					t.Fatalf("got %+v", got)
				}
				all, err := List(c)
				if err != nil || len(all) != 1 {
					t.Fatalf("list %v %d", err, len(all))
				}
			},
		},
		{
			name: "same key different origins",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Type{Key: "book", Origin: OriginProvenencia, Label: "Book"}); err != nil {
					t.Fatal(err)
				}
				if _, err := Upsert(c, Type{Key: "book", Origin: OriginUser, Label: "My Book"}); err != nil {
					t.Fatal(err)
				}
				all, err := List(c)
				if err != nil || len(all) != 2 {
					t.Fatalf("list %v %d", err, len(all))
				}
			},
		},
		{
			name: "rejects blank key",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Type{Key: "  ", Origin: OriginUser, Label: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
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
				id, err := Upsert(c, Type{Key: "census", Origin: OriginProvenencia, Label: "Census"})
				if err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_source_type" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				_, err = Lookup(c, "census", OriginProvenencia)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "in use refuses",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				id, err := Upsert(c, Type{Key: "book", Origin: OriginUser, Label: "Book"})
				if err != nil {
					t.Fatal(err)
				}
				if _, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
					return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: id, Title: "T"})
				}); err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); !errors.Is(err, ErrInUse) {
					t.Fatalf("got %v", err)
				}
				if _, err := GetByID(c, id); err != nil {
					t.Fatalf("type gone: %v", err)
				}
			},
		},
		{
			name: "unused plugin is origin locked",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				id, err := Upsert(c, Type{
					Key: "memorial", Origin: "plugin:findagrave", Label: "Grave memorial",
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); !errors.Is(err, ErrOriginLocked) {
					t.Fatalf("got %v", err)
				}
				if _, err := GetByID(c, id); err != nil {
					t.Fatalf("type gone: %v", err)
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
			name: "unused type with suggestions still erases",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				id, err := Upsert(c, Type{Key: "book", Origin: OriginUser, Label: "Book"})
				if err != nil {
					t.Fatal(err)
				}
				fieldID, err := metadatafields.Upsert(c, metadatafields.Field{
					Key: "author", Origin: metadatafields.OriginUser, Label: "Author", DataType: metadatafields.DataTypeText,
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
					id, fieldID,
				); err != nil {
					t.Fatal(err)
				}
				if err := runDelete(c, userID, id); err != nil {
					t.Fatal(err)
				}
				if _, err := GetByID(c, id); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("type still there: %v", err)
				}
				var joins int
				if err := db.QueryRow(
					`SELECT COUNT(*) FROM source_type_metadata_fields WHERE source_type_id = ?`, id,
				).Scan(&joins); err != nil || joins != 0 {
					t.Fatalf("joins leftover %d %v", joins, err)
				}
				if _, err := metadatafields.GetByID(c, fieldID); err != nil {
					t.Fatalf("field should survive: %v", err)
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
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create mints kebab key from label",
			run: func(t *testing.T, c *database.Catalog) {
				got, err := runCreate(c, "Parish register", "Baptisms, marriages, burials", "type_scroll")
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != "parish-register" || got.Origin != OriginUser {
					t.Fatalf("got %+v", got)
				}
				if got.IconKey != "type_scroll" {
					t.Fatalf("icon_key %+v", got)
				}
			},
		},
		{
			name: "create refuses unknown icon_key",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := runCreate(c, "Deed", "", "type_not_a_real_icon"); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create empty icon_key defaults to type_evidence",
			run: func(t *testing.T, c *database.Catalog) {
				got, err := runCreate(c, "Loose note", "", "")
				if err != nil {
					t.Fatal(err)
				}
				if got.IconKey != DefaultIconKey {
					t.Fatalf("icon_key %+v", got)
				}
			},
		},
		{
			name: "create refuses an unslugifiable label",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := runCreate(c, "—", "", "type_evidence"); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create refuses a duplicate user key",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := runCreate(c, "Parish register", "", "type_scroll"); err != nil {
					t.Fatal(err)
				}
				_, err := runCreate(c, "Parish Register", "", "type_scroll")
				if !errors.Is(err, ErrDuplicateKey) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create may reuse a provenencia key under user origin",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Type{Key: "book", Origin: OriginProvenencia, Label: "Book"}); err != nil {
					t.Fatal(err)
				}
				got, err := runCreate(c, "Book", "", "type_book")
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != "book" || got.Origin != OriginUser {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "update patches label and description, keeping the key",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{Key: "book", Origin: OriginProvenencia, Label: "Book", Description: "Old"})
				if err != nil {
					t.Fatal(err)
				}
				got, err := runUpdate(c, id, "Printed book", "New", "type_book")
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != "book" || got.Label != "Printed book" || got.Description != "New" {
					t.Fatalf("got %+v", got)
				}
				if got.IconKey != "type_book" {
					t.Fatalf("icon_key %+v", got)
				}
			},
		},
		{
			name: "update empty icon_key preserves existing",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{
					Key: "scroll", Origin: OriginUser, Label: "Scroll", IconKey: "type_scroll",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := runUpdate(c, id, "Parish scroll", "", "")
				if err != nil {
					t.Fatal(err)
				}
				if got.IconKey != "type_scroll" {
					t.Fatalf("icon_key %+v", got)
				}
			},
		},
		{
			name: "update refuses a plugin type",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{Key: "memorial", Origin: "plugin:findagrave", Label: "Grave memorial"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runUpdate(c, id, "Mine now", "", "type_book"); !errors.Is(err, ErrLocked) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "update refuses a blank label",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{Key: "book", Origin: OriginUser, Label: "Book"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runUpdate(c, id, "   ", "", "type_evidence"); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "get by id round-trips, missing is ErrNoRows",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{Key: "book", Origin: OriginUser, Label: "Book"})
				if err != nil {
					t.Fatal(err)
				}
				got, err := GetByID(c, id)
				if err != nil || got.Key != "book" {
					t.Fatalf("got %+v %v", got, err)
				}
				missing := []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}
				if _, err := GetByID(c, missing); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "list and UsedBy count the sources of a type",
			run: func(t *testing.T, c *database.Catalog) {
				r, err := ref.Mint(ref.PrefixUser)
				if err != nil {
					t.Fatal(err)
				}
				if err := users.Upsert(c, userID, "Jake", r); err != nil {
					t.Fatal(err)
				}
				used, err := Upsert(c, Type{Key: "book", Origin: OriginUser, Label: "Book"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Upsert(c, Type{Key: "letter", Origin: OriginUser, Label: "Letter"}); err != nil {
					t.Fatal(err)
				}
				for _, title := range []string{"One", "Two"} {
					if _, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
						return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: used, Title: title})
					}); err != nil {
						t.Fatal(err)
					}
				}
				n, err := UsedBy(c, used)
				if err != nil || n != 2 {
					t.Fatalf("UsedBy %d %v", n, err)
				}
				all, err := List(c)
				if err != nil {
					t.Fatal(err)
				}
				byKey := map[string]int{}
				for _, tp := range all {
					byKey[tp.Key] = tp.UsedBy
				}
				if byKey["book"] != 2 || byKey["letter"] != 0 {
					t.Fatalf("got %v", byKey)
				}
			},
		},
		{
			name: "list counts the fields a type suggests",
			run: func(t *testing.T, c *database.Catalog) {
				typeID, err := Upsert(c, Type{Key: "book", Origin: OriginUser, Label: "Book"})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Upsert(c, Type{Key: "letter", Origin: OriginUser, Label: "Letter"}); err != nil {
					t.Fatal(err)
				}
				fieldID, err := metadatafields.Upsert(c, metadatafields.Field{
					Key: "author", Origin: metadatafields.OriginUser, Label: "Author", DataType: metadatafields.DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				// The join belongs to sourcevocab, which imports this
				// package — insert the row directly to avoid the cycle.
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO source_type_metadata_fields (source_type_id, field_id, sort_order) VALUES (?, ?, 0)`,
					typeID, fieldID,
				); err != nil {
					t.Fatal(err)
				}
				all, err := List(c)
				if err != nil {
					t.Fatal(err)
				}
				byKey := map[string]int{}
				for _, tp := range all {
					byKey[tp.Key] = tp.SuggestedFields
				}
				if byKey["book"] != 1 || byKey["letter"] != 0 {
					t.Fatalf("got %v", byKey)
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
				if _, err := Upsert(c, Type{Key: "book", Origin: OriginProvenencia, Label: "Book"}); err != nil {
					t.Fatal(err)
				}
				if _, err := Upsert(c, Type{Key: "letter", Origin: OriginUser, Label: "Letter"}); err != nil {
					t.Fatal(err)
				}
				if _, err := Upsert(c, Type{Key: "grave", Origin: "plugin:findagrave", Label: "Grave"}); err != nil {
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

func runCreate(c *database.Catalog, label, description, iconKey string) (Type, error) {
	got, _, err := writes.Run(c, writes.Op{Action: "create_source_type"},
		func(tx *database.Tx) (Type, []rowchange.Change, error) {
			return Create(tx, label, description, iconKey)
		})
	return got, err
}

func runUpdate(c *database.Catalog, id []byte, label, description, iconKey string) (Type, error) {
	got, _, err := writes.Run(c, writes.Op{Action: "update_source_type"},
		func(tx *database.Tx) (Type, []rowchange.Change, error) {
			return Update(tx, id, label, description, iconKey)
		})
	return got, err
}

func runDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_source_type", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}
