package properties_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

func TestProperties(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog, userID []byte)
	}{
		{
			name: "upsert lookup list",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				id, err := properties.Upsert(c, properties.Property{
					Key: "occupation", Origin: properties.OriginProvenencia, Label: "Occupation", ValueType: propertymeta.ValueTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := properties.Lookup(c, "occupation", properties.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(got.ID) != string(id) || got.ValueType != propertymeta.ValueTypeText {
					t.Fatalf("got %+v", got)
				}
				list, err := properties.List(c)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "rejects bad value type",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				if _, err := properties.Upsert(c, properties.Property{
					Key: "x", Origin: properties.OriginUser, Label: "X", ValueType: "boolean",
				}); !errors.Is(err, properties.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create update delete audited user property",
			run: func(t *testing.T, c *database.Catalog, userID []byte) {
				got, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Custom Fact", propertymeta.ValueTypeText, "note", "")
				})
				if err != nil {
					t.Fatal(err)
				}
				if got.Key != "custom-fact" || got.Origin != properties.OriginUser || got.Cardinality != propertymeta.CardinalitySingle {
					t.Fatalf("got %+v", got)
				}
				updated, err := writes.Call(c, writes.Op{Action: "update_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Update(tx, userID, got.ID, "Custom Fact 2", propertymeta.ValueTypeText, "", "")
				})
				if err != nil {
					t.Fatal(err)
				}
				if updated.Label != "Custom Fact 2" || updated.Cardinality != propertymeta.CardinalitySingle {
					t.Fatalf("label %q", updated.Label)
				}
				if _, err := writes.Call(c, writes.Op{Action: "delete_property", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := properties.Delete(tx, userID, got.ID)
					return struct{}{}, changes, err
				}); err != nil {
					t.Fatal(err)
				}
				_, err = properties.GetByID(c, got.ID)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("after delete %v", err)
				}
			},
		},
		{
			name: "duplicate user key",
			run: func(t *testing.T, c *database.Catalog, userID []byte) {
				if _, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Dup", propertymeta.ValueTypeText, "", "")
				}); err != nil {
					t.Fatal(err)
				}
				_, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Dup", propertymeta.ValueTypeText, "", "")
				})
				if !errors.Is(err, properties.ErrDuplicateKey) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create stores cardinality",
			run: func(t *testing.T, c *database.Catalog, userID []byte) {
				got, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Languages Spoken", propertymeta.ValueTypeText, "", propertymeta.CardinalityMultiple)
				})
				if err != nil {
					t.Fatal(err)
				}
				if got.Cardinality != propertymeta.CardinalityMultiple {
					t.Fatalf("cardinality %q", got.Cardinality)
				}
				if _, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
					return properties.Create(tx, userID, "Bad Holds", propertymeta.ValueTypeText, "", "triple")
				}); !errors.Is(err, properties.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create refuses registry-only and unknown value types",
			run: func(t *testing.T, c *database.Catalog, userID []byte) {
				for _, vt := range []string{propertymeta.ValueTypeTerm, propertymeta.ValueTypeDepiction, "not_a_type"} {
					_, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
						return properties.Create(tx, userID, "Event Kind", vt, "", "")
					})
					if !errors.Is(err, properties.ErrInvalid) {
						t.Fatalf("%s: got %v", vt, err)
					}
				}
			},
		},
		{
			name: "upsert accepts depiction",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				id, err := properties.Upsert(c, properties.Property{
					Key: "snapshot", Origin: properties.OriginProvenencia, Label: "Snapshot", ValueType: propertymeta.ValueTypeDepiction,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := properties.Lookup(c, "snapshot", properties.OriginProvenencia)
				if err != nil || got.ValueType != propertymeta.ValueTypeDepiction || string(got.ID) != string(id) {
					t.Fatalf("got %+v %v", got, err)
				}
			},
		},
		{
			name: "upsert allows proveniencia term",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				id, err := properties.Upsert(c, properties.Property{
					Key: "event_type", Origin: properties.OriginProvenencia, Label: "Event type", ValueType: propertymeta.ValueTypeTerm,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
				if err != nil || got.ValueType != propertymeta.ValueTypeTerm || string(got.ID) != string(id) {
					t.Fatalf("got %+v %v", got, err)
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
				if ver < 24 {
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
