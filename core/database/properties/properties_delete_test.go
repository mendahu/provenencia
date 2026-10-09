package properties_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

const testLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

func TestDelete(t *testing.T) {
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "unused user property erases and audits",
			run: func(t *testing.T, c *database.Catalog) {
				got, err := properties.Create(c, userID, "Burial Ground", properties.ValueTypeText, "", "")
				if err != nil {
					t.Fatal(err)
				}
				if err := properties.Delete(c, userID, got.ID); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "delete_property" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				if _, err := properties.GetByID(c, got.ID); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("still there: %v", err)
				}
			},
		},
		{
			name: "bound only to a type still erases",
			run: func(t *testing.T, c *database.Catalog) {
				got, err := properties.Create(c, userID, "Maiden Name", properties.ValueTypeText, "", "")
				if err != nil {
					t.Fatal(err)
				}
				typeID, err := subjecttypes.Upsert(c, subjecttypes.Type{
					Key: "person", Origin: subjecttypes.OriginProvenencia, Label: "Person",
					RefPrefix: "PER", CandidateRefPrefix: "CPR",
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(
					`INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order) VALUES (?, ?, 0)`,
					typeID, got.ID,
				); err != nil {
					t.Fatal(err)
				}
				n, err := properties.UsedBy(c, got.ID)
				if err != nil || n != 0 {
					t.Fatalf("usedBy want 0 got %d %v", n, err)
				}
				if err := properties.Delete(c, userID, got.ID); err != nil {
					t.Fatal(err)
				}
				if _, err := properties.GetByID(c, got.ID); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("still there: %v", err)
				}
				var joins int
				if err := db.QueryRow(
					`SELECT COUNT(*) FROM subject_type_properties WHERE property_id = ?`, got.ID,
				).Scan(&joins); err != nil || joins != 0 {
					t.Fatalf("binding leftover %d %v", joins, err)
				}
			},
		},
		{
			name: "observation inbound is in use",
			run: func(t *testing.T, c *database.Catalog) {
				if err := subjectvocab.Install(c); err != nil {
					t.Fatal(err)
				}
				prop, err := properties.Create(c, userID, "Outcome", properties.ValueTypeText, "", "")
				if err != nil {
					t.Fatal(err)
				}
				obsID := insertObservationOn(t, c, userID, prop.ID)
				n, err := properties.UsedBy(c, prop.ID)
				if err != nil || n != 1 {
					t.Fatalf("usedBy want 1 got %d %v", n, err)
				}
				if err := properties.Delete(c, userID, prop.ID); !errors.Is(err, properties.ErrInUse) {
					t.Fatalf("got %v", err)
				}
				if _, err := properties.GetByID(c, prop.ID); err != nil {
					t.Fatalf("property gone: %v", err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var leftover int
				if err := db.QueryRow(`SELECT COUNT(*) FROM observations WHERE id = ?`, obsID).Scan(&leftover); err != nil || leftover != 1 {
					t.Fatalf("observation leftover %d %v", leftover, err)
				}
			},
		},
		{
			name: "unused terms are in use",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := properties.Upsert(c, properties.Property{
					Key: "residence_status", Origin: properties.OriginUser,
					Label: "Residence status", ValueType: properties.ValueTypeTerm,
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, _, err := writes.Run(c, writes.Op{Action: "create_property_term", UserID: userID},
					func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
						return propertyterms.Create(tx, userID, id, "Lodger", "")
					}); err != nil {
					t.Fatal(err)
				}
				if err := properties.Delete(c, userID, id); !errors.Is(err, properties.ErrInUse) {
					t.Fatalf("got %v", err)
				}
				if _, err := properties.GetByID(c, id); err != nil {
					t.Fatalf("property gone: %v", err)
				}
			},
		},
		{
			name: "seeded is origin locked",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := properties.Upsert(c, properties.Property{
					Key: "occupation", Origin: properties.OriginProvenencia,
					Label: "Occupation", ValueType: properties.ValueTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := properties.Delete(c, userID, id); !errors.Is(err, properties.ErrOriginLocked) {
					t.Fatalf("got %v", err)
				}
				if _, err := properties.GetByID(c, id); err != nil {
					t.Fatalf("property gone: %v", err)
				}
			},
		},
		{
			name: "unused plugin is origin locked",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := properties.Upsert(c, properties.Property{
					Key: "plugin_fact", Origin: "plugin:acme",
					Label: "Plugin fact", ValueType: properties.ValueTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := properties.Delete(c, userID, id); !errors.Is(err, properties.ErrOriginLocked) {
					t.Fatalf("got %v", err)
				}
				if _, err := properties.GetByID(c, id); err != nil {
					t.Fatalf("property gone: %v", err)
				}
			},
		},
		{
			name: "missing is invalid",
			run: func(t *testing.T, c *database.Catalog) {
				missing := make([]byte, 16)
				missing[15] = 9
				if err := properties.Delete(c, userID, missing); !errors.Is(err, properties.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "bad id",
			run: func(t *testing.T, c *database.Catalog) {
				if err := properties.Delete(c, userID, []byte{1}); !errors.Is(err, properties.ErrInvalid) {
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
			r, err := ref.Mint(ref.PrefixUser)
			if err != nil {
				t.Fatal(err)
			}
			if err := users.Upsert(c, userID, "Tester", r); err != nil {
				t.Fatal(err)
			}
			tt.run(t, c)
		})
	}
}

func latestAction(t *testing.T, c *database.Catalog) string {
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

func insertObservationOn(t *testing.T, c *database.Catalog, userID, propertyID []byte) []byte {
	t.Helper()
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifacts.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	placeType, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT OR IGNORE INTO subject_type_properties (subject_type_id, property_id, sort_order) VALUES (?, ?, 0)`,
		placeType.ID, propertyID,
	); err != nil {
		t.Fatal(err)
	}
	place, err := subjects.Create(c, userID, subjects.CreateInput{
		SourceID: src.ID, SubjectTypeID: placeType.ID, Label: "Leeds",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
		ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "Leeds",
	}, []observations.Input{{
		SubjectID: place.ID, PropertyID: propertyID, ValueText: "Leeds", HasText: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Observations) != 1 {
		t.Fatalf("observations %d", len(res.Observations))
	}
	return res.Observations[0].ID
}

func TestUsedByCountsObservationsNotBindings(t *testing.T) {
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
	prop, err := properties.Create(c, userID, "Custom Fact", properties.ValueTypeText, "", "")
	if err != nil {
		t.Fatal(err)
	}
	typeID, err := subjecttypes.Upsert(c, subjecttypes.Type{
		Key: "person", Origin: subjecttypes.OriginProvenencia, Label: "Person",
		RefPrefix: "PER", CandidateRefPrefix: "CPR",
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order) VALUES (?, ?, 0)`,
		typeID, prop.ID,
	); err != nil {
		t.Fatal(err)
	}
	n, err := properties.UsedBy(c, prop.ID)
	if err != nil || n != 0 {
		t.Fatalf("binding must not count: %d %v", n, err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	insertObservationOn(t, c, userID, prop.ID)
	n, err = properties.UsedBy(c, prop.ID)
	if err != nil || n != 1 {
		t.Fatalf("want 1 observation got %d %v", n, err)
	}
	list, err := properties.List(c)
	if err != nil {
		t.Fatal(err)
	}
	var listed int
	for _, row := range list {
		if string(row.ID) == string(prop.ID) {
			listed = row.UsedBy
		}
	}
	if listed != 1 {
		t.Fatalf("list usedBy %d", listed)
	}
}

func TestDeleteRejectsEmptyUserID(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	id := make([]byte, 16)
	id[0] = 1
	if err := properties.Delete(c, nil, id); !errors.Is(err, properties.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
