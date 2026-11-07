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
	"github.com/mendahu/provenencia/core/vocabkey"
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
					Key: "occupation", Origin: vocabkey.OriginProvenencia, Label: "Occupation", ValueType: propertymeta.ValueTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := properties.Lookup(c, "occupation", vocabkey.OriginProvenencia)
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
					Key: "x", Origin: vocabkey.OriginUser, Label: "X", ValueType: "real",
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
				if got.Key != "custom-fact" || got.Origin != vocabkey.OriginUser || got.Cardinality != propertymeta.CardinalitySingle {
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
					Key: "snapshot", Origin: vocabkey.OriginProvenencia, Label: "Snapshot", ValueType: propertymeta.ValueTypeDepiction,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := properties.Lookup(c, "snapshot", vocabkey.OriginProvenencia)
				if err != nil || got.ValueType != propertymeta.ValueTypeDepiction || string(got.ID) != string(id) {
					t.Fatalf("got %+v %v", got, err)
				}
			},
		},
		{
			name: "upsert allows proveniencia term",
			run: func(t *testing.T, c *database.Catalog, _ []byte) {
				id, err := properties.Upsert(c, properties.Property{
					Key: "event_type", Origin: vocabkey.OriginProvenencia, Label: "Event type", ValueType: propertymeta.ValueTypeTerm,
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := properties.Lookup(c, "event_type", vocabkey.OriginProvenencia)
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

func newPropertiesCatalog(t *testing.T) (*database.Catalog, []byte) {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", r); err != nil {
		t.Fatal(err)
	}
	return c, userID
}

func TestValueTypeRules(t *testing.T) {
	create := func(c *database.Catalog, userID []byte, label, valueType, cardinality string) error {
		_, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Create(tx, userID, label, valueType, "", cardinality)
		})
		return err
	}
	tests := []struct {
		name string
		run  func(c *database.Catalog, userID []byte) error
		want error
		// stored is the (value type, cardinality) read back for key, when set.
		key, storedType, storedCard string
	}{
		{
			name: "upsert boolean empty cardinality is single",
			run: func(c *database.Catalog, _ []byte) error {
				_, err := properties.Upsert(c, properties.Property{Key: "literate", Origin: vocabkey.OriginUser, Label: "Literate", ValueType: propertymeta.ValueTypeBoolean})
				return err
			},
			key: "literate", storedType: propertymeta.ValueTypeBoolean, storedCard: propertymeta.CardinalitySingle,
		},
		{
			name: "upsert boolean multiple is refused",
			run: func(c *database.Catalog, _ []byte) error {
				_, err := properties.Upsert(c, properties.Property{Key: "literate", Origin: vocabkey.OriginUser, Label: "Literate", ValueType: propertymeta.ValueTypeBoolean, Cardinality: propertymeta.CardinalityMultiple})
				return err
			},
			want: properties.ErrInvalid,
		},
		{
			name: "create boolean empty cardinality is single",
			run: func(c *database.Catalog, userID []byte) error {
				return create(c, userID, "Literate", propertymeta.ValueTypeBoolean, "")
			},
			key: "literate", storedType: propertymeta.ValueTypeBoolean, storedCard: propertymeta.CardinalitySingle,
		},
		{
			name: "create boolean multiple is refused",
			run: func(c *database.Catalog, userID []byte) error {
				return create(c, userID, "Literate", propertymeta.ValueTypeBoolean, propertymeta.CardinalityMultiple)
			},
			want: properties.ErrInvalid,
		},
		{
			name: "create integer multiple is allowed",
			run: func(c *database.Catalog, userID []byte) error {
				return create(c, userID, "Ages", propertymeta.ValueTypeInteger, propertymeta.CardinalityMultiple)
			},
			key: "ages", storedType: propertymeta.ValueTypeInteger, storedCard: propertymeta.CardinalityMultiple,
		},
		{
			name: "create term is refused",
			run: func(c *database.Catalog, userID []byte) error {
				return create(c, userID, "Kind", propertymeta.ValueTypeTerm, "")
			},
			want: properties.ErrInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, userID := newPropertiesCatalog(t)
			err := tt.run(c, userID)
			if tt.want != nil {
				if !errors.Is(err, tt.want) {
					t.Fatalf("got %v, want %v", err, tt.want)
				}
				list, lerr := properties.List(c)
				if lerr != nil || len(list) != 0 {
					t.Fatalf("rows after refusal: %v %d", lerr, len(list))
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got, err := properties.Lookup(c, tt.key, vocabkey.OriginUser)
			if err != nil {
				t.Fatal(err)
			}
			if got.ValueType != tt.storedType || got.Cardinality != tt.storedCard {
				t.Fatalf("got %s/%s, want %s/%s", got.ValueType, got.Cardinality, tt.storedType, tt.storedCard)
			}
		})
	}
}

func TestBooleanCardinalityStaysSingle(t *testing.T) {
	c, userID := newPropertiesCatalog(t)
	id, err := properties.Upsert(c, properties.Property{Key: "literate", Origin: vocabkey.OriginUser, Label: "Literate", ValueType: propertymeta.ValueTypeBoolean})
	if err != nil {
		t.Fatal(err)
	}
	_, err = writes.Call(c, writes.Op{Action: "update_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
		return properties.Update(tx, userID, id, "Literate", propertymeta.ValueTypeBoolean, "", propertymeta.CardinalityMultiple)
	})
	if !errors.Is(err, properties.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
	got, err := properties.GetByID(c, id)
	if err != nil || got.Cardinality != propertymeta.CardinalitySingle {
		t.Fatalf("row changed: %v %+v", err, got)
	}

	// The table holds the rule too, not only the Go write paths.
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE properties SET cardinality = 'multiple' WHERE id = ?`, id); err == nil {
		t.Fatal("the table accepted a multiple-valued boolean Property")
	}
}
