package properties

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestProperties(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog, userID []byte)
	}{
		{
			name: "upsert lookup list",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				id, err := Upsert(c, Property{
					Key: "occupation", Origin: OriginProvenencia, Label: "Occupation", ValueType: ValueTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, "occupation", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.ValueType != ValueTypeText {
					t.Fatalf("got %+v", got)
				}
				list, err := List(c)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "rejects bad value type",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				if _, err := Upsert(c, Property{
					Key: "x", Origin: OriginUser, Label: "X", ValueType: "boolean",
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create update delete audited user property",
			run: func(t *testing.T, c *database.Catalog, userID []byte) {
				got, err := Create(c, userID, "Custom Fact", ValueTypeText, "note")
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != "custom-fact" || got.Origin != OriginUser {
					t.Fatalf("got %+v", got)
				}
				updated, err := Update(c, userID, got.ID, "Custom Fact 2", ValueTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if updated.Label != "Custom Fact 2" {
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
			name: "duplicate user key",
			run: func(t *testing.T, c *database.Catalog, userID []byte) {
				if _, err := Create(c, userID, "Dup", ValueTypeText, ""); err != nil {
					t.Fatal(err)
				}
				_, err := Create(c, userID, "Dup", ValueTypeText, "")
				if !errors.Is(err, ErrDuplicateKey) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "migration creates properties table",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var name string
				if err := db.QueryRow(
					`SELECT name FROM sqlite_schema WHERE type='table' AND name='properties'`,
				).Scan(&name); err != nil || name != "properties" {
					t.Fatalf("table %q %v", name, err)
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 23 {
					t.Fatalf("user_version=%d", ver)
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
			tt.run(t, c, userID)
		})
	}
}
