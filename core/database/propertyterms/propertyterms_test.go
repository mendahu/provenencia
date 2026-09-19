package propertyterms

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestPropertyTerms(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog, userID, propID []byte)
	}{
		{
			name: "migration creates property_terms table",
			run: func(t *testing.T, c *database.Catalog, _, _ []byte) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var name string
				if err := db.QueryRow(
					`SELECT name FROM sqlite_schema WHERE type='table' AND name='property_terms'`,
				).Scan(&name); err != nil || name != "property_terms" {
					t.Fatalf("table %q %v", name, err)
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 24 {
					t.Fatalf("user_version=%d", ver)
				}
			},
		},
		{
			name: "upsert lookup list",
			run: func(t *testing.T, c *database.Catalog, _, propID []byte) {
				id, err := Upsert(c, Term{
					PropertyID: propID, Key: "birth", Origin: OriginProvenencia, Label: "Birth",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, propID, "birth", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.Label != "Birth" {
					t.Fatalf("got %+v", got)
				}
				list, err := ListByProperty(c, propID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "create update delete user term",
			run: func(t *testing.T, c *database.Catalog, userID, propID []byte) {
				got, err := Create(c, userID, propID, "Land Grant", "custom")
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != "land-grant" || got.Origin != OriginUser {
					t.Fatalf("got %+v", got)
				}
				updated, err := Update(c, userID, got.ID, "Land Grant 2", "")
				if err != nil {
					t.Fatal(err)
				}
				if updated.Label != "Land Grant 2" {
					t.Fatalf("label %q", updated.Label)
				}
				if err := Delete(c, userID, got.ID); err != nil {
					t.Fatal(err)
				}
				_, err = GetByID(c, got.ID)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("after delete %v", err)
				}
			},
		},
		{
			name: "refuse mutate product term",
			run: func(t *testing.T, c *database.Catalog, userID, propID []byte) {
				id, err := Upsert(c, Term{
					PropertyID: propID, Key: "death", Origin: OriginProvenencia, Label: "Death",
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Update(c, userID, id, "X", ""); !errors.Is(err, ErrLocked) {
					t.Fatalf("update %v", err)
				}
				if err := Delete(c, userID, id); !errors.Is(err, ErrLocked) {
					t.Fatalf("delete %v", err)
				}
			},
		},
		{
			name: "duplicate user key",
			run: func(t *testing.T, c *database.Catalog, userID, propID []byte) {
				if _, err := Create(c, userID, propID, "Dup", ""); err != nil {
					t.Fatal(err)
				}
				_, err := Create(c, userID, propID, "Dup", "")
				if !errors.Is(err, ErrDuplicateKey) {
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
			userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
			r, err := ref.Mint(ref.PrefixUser)
			if err != nil {
				t.Fatal(err)
			}
			if err := users.Upsert(c, userID, "Tester", r); err != nil {
				t.Fatal(err)
			}
			propID, err := properties.Upsert(c, properties.Property{
				Key: "event_type", Origin: properties.OriginProvenencia,
				Label: "Event type", ValueType: properties.ValueTypeTerm,
			})
			if err != nil {
				t.Fatal(err)
			}
			tt.run(t, c, userID, propID)
		})
	}
}
