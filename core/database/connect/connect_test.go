package connect

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/ref"
)

func TestCreateCitedBridge(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	validLocator := `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

	type seed struct {
		c          *database.Catalog
		source     sources.Source
		artifact   artifacts.Artifact
		person     subjects.Subject
		personB    subjects.Subject
		event      subjects.Subject
		place      subjects.Subject
		personProp properties.Property
		eventProp  properties.Property
		placeProp  properties.Property
		related    properties.Property
		roleProp   properties.Property
		relType    properties.Property
		roleTerm   propertyterms.Term
		relTerm    propertyterms.Term
	}

	mustSeed := func(t *testing.T) seed {
		t.Helper()
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
		ur, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Tester", ur); err != nil {
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
		src, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: typeID, Title: "Register",
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := artifacts.Create(c, userID, artifacts.CreateInput{
			SourceID: src.ID, Label: "Scan",
		})
		if err != nil {
			t.Fatal(err)
		}
		lookupType := func(key string) []byte {
			t.Helper()
			typ, err := subjecttypes.Lookup(c, key, subjecttypes.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			return typ.ID
		}
		lookupProp := func(key string) properties.Property {
			t.Helper()
			p, err := properties.Lookup(c, key, properties.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			return p
		}
		create := func(typeID []byte, label string) subjects.Subject {
			t.Helper()
			s, err := subjects.Create(c, userID, subjects.CreateInput{
				SourceID: src.ID, SubjectTypeID: typeID, Label: label,
			})
			if err != nil {
				t.Fatal(err)
			}
			return s
		}
		personProp := lookupProp("person")
		eventProp := lookupProp("event")
		placeProp := lookupProp("place")
		related := lookupProp("related_to")
		roleProp := lookupProp("role")
		relType := lookupProp("relationship_type")
		roleTerm, err := propertyterms.Lookup(c, roleProp.ID, "witness", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		relTerm, err := propertyterms.Lookup(c, relType.ID, "spouse", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		return seed{
			c: c, source: src, artifact: art,
			person: create(lookupType("person"), "Alice"),
			personB: create(lookupType("person"), "Bob"),
			event: create(lookupType("event"), "Wedding"),
			place: create(lookupType("place"), "Leeds"),
			personProp: personProp, eventProp: eventProp, placeProp: placeProp,
			related: related, roleProp: roleProp, relType: relType,
			roleTerm: roleTerm, relTerm: relTerm,
		}
	}

	countBridges := func(t *testing.T, c *database.Catalog, typeKey string) int {
		t.Helper()
		typ, err := subjecttypes.Lookup(c, typeKey, subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM subjects WHERE subject_type_id = ?`, typ.ID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	t.Run("person event participation", func(t *testing.T) {
		s := mustSeed(t)
		res, err := CreateCitedBridge(s.c, userID, CreateInput{
			SourceID:      s.source.ID,
			FromSubjectID: s.person.ID,
			ToSubjectID:   s.event.ID,
			BridgeTypeKey: "participation",
			Label:         "",
			GridX:         2,
			GridY:         3,
			Citation: citations.CreateInput{
				ArtifactID:  s.artifact.ID,
				LocatorJSON: validLocator,
			},
			Observations: []observations.Input{
				{PropertyID: s.personProp.ID, ValueSubjectID: s.person.ID},
				{PropertyID: s.eventProp.ID, ValueSubjectID: s.event.ID},
				{PropertyID: s.roleProp.ID, ValueTermID: s.roleTerm.ID},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Observations) != 3 {
			t.Fatalf("observations %d", len(res.Observations))
		}
		got, err := subjects.Get(s.c, res.Subject.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Ref == "" {
			t.Fatal("missing bridge ref")
		}
	})

	t.Run("person person relationship", func(t *testing.T) {
		s := mustSeed(t)
		_, err := CreateCitedBridge(s.c, userID, CreateInput{
			SourceID:      s.source.ID,
			FromSubjectID: s.person.ID,
			ToSubjectID:   s.personB.ID,
			BridgeTypeKey: "relationship",
			Citation: citations.CreateInput{
				ArtifactID:  s.artifact.ID,
				LocatorJSON: validLocator,
			},
			Observations: []observations.Input{
				{PropertyID: s.personProp.ID, ValueSubjectID: s.person.ID},
				{PropertyID: s.related.ID, ValueSubjectID: s.personB.ID},
				{PropertyID: s.relType.ID, ValueTermID: s.relTerm.ID},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("event place location", func(t *testing.T) {
		s := mustSeed(t)
		_, err := CreateCitedBridge(s.c, userID, CreateInput{
			SourceID:      s.source.ID,
			FromSubjectID: s.event.ID,
			ToSubjectID:   s.place.ID,
			BridgeTypeKey: "location",
			Citation: citations.CreateInput{
				ArtifactID:  s.artifact.ID,
				LocatorJSON: validLocator,
			},
			Observations: []observations.Input{
				{PropertyID: s.eventProp.ID, ValueSubjectID: s.event.ID},
				{PropertyID: s.placeProp.ID, ValueSubjectID: s.place.ID},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("person place refused", func(t *testing.T) {
		s := mustSeed(t)
		before := countBridges(t, s.c, "location")
		_, err := CreateCitedBridge(s.c, userID, CreateInput{
			SourceID:      s.source.ID,
			FromSubjectID: s.person.ID,
			ToSubjectID:   s.place.ID,
			BridgeTypeKey: "location",
			Citation: citations.CreateInput{
				ArtifactID:  s.artifact.ID,
				LocatorJSON: validLocator,
			},
			Observations: []observations.Input{
				{PropertyID: s.personProp.ID, ValueSubjectID: s.person.ID},
				{PropertyID: s.placeProp.ID, ValueSubjectID: s.place.ID},
			},
		})
		if !errors.Is(err, ErrRefused) {
			t.Fatalf("got %v want refused", err)
		}
		if countBridges(t, s.c, "location") != before {
			t.Fatal("refused pair wrote a bridge")
		}
	})

	t.Run("citation failure rolls back subject", func(t *testing.T) {
		s := mustSeed(t)
		before := countBridges(t, s.c, "participation")
		_, err := CreateCitedBridge(s.c, userID, CreateInput{
			SourceID:      s.source.ID,
			FromSubjectID: s.person.ID,
			ToSubjectID:   s.event.ID,
			BridgeTypeKey: "participation",
			Citation: citations.CreateInput{
				ArtifactID:  s.artifact.ID,
				LocatorJSON: `{}`,
			},
			Observations: []observations.Input{
				{PropertyID: s.personProp.ID, ValueSubjectID: s.person.ID},
				{PropertyID: s.eventProp.ID, ValueSubjectID: s.event.ID},
				{PropertyID: s.roleProp.ID, ValueTermID: s.roleTerm.ID},
			},
		})
		if !errors.Is(err, locator.ErrInvalid) {
			t.Fatalf("got %v want locator invalid", err)
		}
		if countBridges(t, s.c, "participation") != before {
			t.Fatal("failed citation left a bridge subject")
		}
	})
}
