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
	"github.com/mendahu/provenencia/core/database/subjectpositions"
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
		create := func(typeID []byte, label string, x, y int64) subjects.Subject {
			t.Helper()
			s, err := subjects.Create(c, userID, subjects.CreateInput{
				SourceID: src.ID, SubjectTypeID: typeID, Label: label,
			}, &subjects.Placement{GridX: x, GridY: y})
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
			person:     create(lookupType("person"), "Alice", 0, 0),
			personB:    create(lookupType("person"), "Bob", 4, 0),
			event:      create(lookupType("event"), "Wedding", 2, 4),
			place:      create(lookupType("place"), "Leeds", 6, 4),
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
		pos, err := subjectpositions.Get(s.c, res.Subject.ID)
		if err != nil {
			t.Fatal(err)
		}
		if pos.GridX != 1 || pos.GridY != 2 {
			t.Fatalf("midpoint (%d,%d)", pos.GridX, pos.GridY)
		}
		db, err := s.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actions int
		if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions WHERE action_type = 'create_cited_bridge'`).Scan(&actions); err != nil {
			t.Fatal(err)
		}
		if actions != 1 {
			t.Fatalf("revisions %d", actions)
		}
		var latest string
		if err := db.QueryRow(`SELECT action_type FROM audit_transactions ORDER BY rowid DESC LIMIT 1`).Scan(&latest); err != nil {
			t.Fatal(err)
		}
		if latest != "create_cited_bridge" {
			t.Fatalf("latest %q", latest)
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

	t.Run("location extra term invalid", func(t *testing.T) {
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
				{PropertyID: s.roleProp.ID, ValueTermID: s.roleTerm.ID},
			},
		})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v want invalid", err)
		}
	})

	t.Run("attach existing citation", func(t *testing.T) {
		s := mustSeed(t)
		cited, err := citations.CreateWithObservations(s.c, userID, citations.CreateInput{
			ArtifactID:  s.artifact.ID,
			LocatorJSON: validLocator,
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		res, err := CreateCitedBridge(s.c, userID, CreateInput{
			SourceID:      s.source.ID,
			FromSubjectID: s.event.ID,
			ToSubjectID:   s.place.ID,
			BridgeTypeKey: "location",
			CitationID:    cited.Citation.ID,
			Observations: []observations.Input{
				{PropertyID: s.eventProp.ID, ValueSubjectID: s.event.ID},
				{PropertyID: s.placeProp.ID, ValueSubjectID: s.place.ID},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if string(res.Citation.ID) != string(cited.Citation.ID) {
			t.Fatal("did not attach")
		}
		if res.Citation.LocatorJSON != validLocator {
			t.Fatalf("locator %q", res.Citation.LocatorJSON)
		}
	})

	t.Run("attach with citation fields invalid", func(t *testing.T) {
		s := mustSeed(t)
		cited, err := citations.CreateWithObservations(s.c, userID, citations.CreateInput{
			ArtifactID:  s.artifact.ID,
			LocatorJSON: validLocator,
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, err = CreateCitedBridge(s.c, userID, CreateInput{
			SourceID:      s.source.ID,
			FromSubjectID: s.event.ID,
			ToSubjectID:   s.place.ID,
			BridgeTypeKey: "location",
			CitationID:    cited.Citation.ID,
			Citation: citations.CreateInput{
				Transcription: "nope",
			},
			Observations: []observations.Input{
				{PropertyID: s.eventProp.ID, ValueSubjectID: s.event.ID},
				{PropertyID: s.placeProp.ID, ValueSubjectID: s.place.ID},
			},
		})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v want invalid", err)
		}
	})

	t.Run("missing position invalid", func(t *testing.T) {
		s := mustSeed(t)
		if err := subjectpositions.Clear(s.c, s.place.ID); err != nil {
			t.Fatal(err)
		}
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
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v want invalid", err)
		}
	})
}

func TestFloorDiv(t *testing.T) {
	cases := []struct {
		a, b, want int64
	}{
		{3, 2, 1},
		{4, 2, 2},
		{0, 2, 0},
		{-1, 2, -1},
		{-3, 2, -2},
		{-4, 2, -2},
		{5, 2, 2},
		{-5, 2, -3},
		{1, 2, 0},
	}
	for _, tc := range cases {
		if got := floorDiv(tc.a, tc.b); got != tc.want {
			t.Fatalf("floorDiv(%d,%d)=%d want %d", tc.a, tc.b, got, tc.want)
		}
	}
	if floorDiv(-3+1+1, 2) != -1 {
		t.Fatal("negative odd midpoint")
	}
}
