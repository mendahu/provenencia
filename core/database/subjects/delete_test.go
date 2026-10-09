package subjects_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
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

func TestSubjectDelete(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	setup := func(t *testing.T) (*database.Catalog, sources.Source, artifacts.Artifact, subjects.Subject, subjects.Subject) {
		t.Helper()
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
		if err := subjectvocab.Install(c); err != nil {
			t.Fatal(err)
		}
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID, Title: "Register",
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := runArtifactCreate(c, userID, artifacts.CreateInput{
			SourceID: src.ID, Label: "Scan",
		})
		if err != nil {
			t.Fatal(err)
		}
		personType, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		eventType, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		alice, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: personType.ID, Label: "Alice",
		}, &subjects.Placement{GridX: 0, GridY: 0})
		if err != nil {
			t.Fatal(err)
		}
		wedding, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: eventType.ID, Label: "Wedding",
		}, &subjects.Placement{GridX: 4, GridY: 4})
		if err != nil {
			t.Fatal(err)
		}
		return c, src, art, alice, wedding
	}

	mustBridge := func(t *testing.T, c *database.Catalog, src sources.Source, art artifacts.Artifact, alice, wedding subjects.Subject) connect.Result {
		t.Helper()
		roleProp, err := properties.Lookup(c, "role", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		personProp, err := properties.Lookup(c, "person", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		eventProp, err := properties.Lookup(c, "event", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		roleTerm, err := propertyterms.Lookup(c, roleProp.ID, "witness", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		bridge, err := connect.CreateCitedBridge(c, userID, connect.CreateInput{
			SourceID: src.ID, FromSubjectID: alice.ID, ToSubjectID: wedding.ID,
			BridgeTypeKey: "participation",
			Citation:      citations.CreateInput{ArtifactID: art.ID, LocatorJSON: testLocator},
			Observations: []observations.Input{
				{PropertyID: personProp.ID, ValueSubjectID: alice.ID},
				{PropertyID: eventProp.ID, ValueSubjectID: wedding.ID},
				{PropertyID: roleProp.ID, ValueTermID: roleTerm.ID},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return bridge
	}

	t.Run("G2 endpoint refuses", func(t *testing.T) {
		c, src, art, alice, wedding := setup(t)
		mustBridge(t, c, src, art, alice, wedding)
		if err := subjects.Delete(c, userID, alice.ID); !errors.Is(err, subjects.ErrInUse) {
			t.Fatalf("delete %v", err)
		}
		if _, err := subjects.Get(c, alice.ID); err != nil {
			t.Fatalf("alice gone: %v", err)
		}
	})

	t.Run("G6 facets-only bridge erases", func(t *testing.T) {
		c, src, art, alice, wedding := setup(t)
		bridge := mustBridge(t, c, src, art, alice, wedding)
		if err := subjects.Delete(c, userID, bridge.Subject.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := subjects.Get(c, bridge.Subject.ID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("bridge still there: %v", err)
		}
		if _, err := subjects.Get(c, alice.ID); err != nil {
			t.Fatalf("alice %v", err)
		}
		if _, err := subjects.Get(c, wedding.ID); err != nil {
			t.Fatalf("wedding %v", err)
		}
		if _, err := citations.Get(c, bridge.Citation.ID); err != nil {
			t.Fatalf("citation %v", err)
		}
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM observations WHERE subject_id = ?`, bridge.Subject.ID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("facet rows left %d", n)
		}
		var deleted int
		if err := db.QueryRow(
			`SELECT COUNT(*) FROM audit_changes WHERE entity_type = 'observation' AND action = 'delete'`,
		).Scan(&deleted); err != nil {
			t.Fatal(err)
		}
		if deleted != 3 {
			t.Fatalf("facet observation audit %d", deleted)
		}
	})

	t.Run("G6 extra observation refuses", func(t *testing.T) {
		c, src, art, alice, wedding := setup(t)
		bridge := mustBridge(t, c, src, art, alice, wedding)
		toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		if err := subjectvocab.AppendBinding(c, bridge.Subject.SubjectTypeID, toponym.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := observations.AddToCitation(c, userID, bridge.Citation.ID, []observations.Input{{
			SubjectID: bridge.Subject.ID, PropertyID: toponym.ID, ValueText: "extra", HasText: true,
		}}); err != nil {
			t.Fatal(err)
		}
		if err := subjects.Delete(c, userID, bridge.Subject.ID); !errors.Is(err, subjects.ErrInUse) {
			t.Fatalf("delete %v", err)
		}
		if _, err := subjects.Get(c, bridge.Subject.ID); err != nil {
			t.Fatalf("bridge gone: %v", err)
		}
	})

	t.Run("edge observation stays edge_locked", func(t *testing.T) {
		c, src, art, alice, wedding := setup(t)
		bridge := mustBridge(t, c, src, art, alice, wedding)
		var edgeID []byte
		for _, o := range bridge.Observations {
			if err := observations.Delete(c, userID, o.ID); errors.Is(err, observations.ErrEdgeLocked) {
				edgeID = o.ID
				break
			}
		}
		if len(edgeID) == 0 {
			t.Fatal("no edge_locked observation")
		}
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM observations WHERE id = ?`, edgeID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("edge gone: %d", n)
		}
	})

	t.Run("missing is invalid", func(t *testing.T) {
		c, _, _, _, _ := setup(t)
		missing := make([]byte, 16)
		missing[15] = 9
		if err := subjects.Delete(c, userID, missing); !errors.Is(err, subjects.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
