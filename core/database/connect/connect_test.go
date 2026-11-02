package connect

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/locator"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
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
			Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
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
		art, err := artifactstest.Create(c, userID, artifacts.CreateInput{
			SourceID: src.ID, Label: "Scan",
		})
		if err != nil {
			t.Fatal(err)
		}
		lookupType := func(key string) []byte {
			t.Helper()
			typ, err := subjecttypes.Lookup(c, key, vocabkey.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			return typ.ID
		}
		lookupProp := func(key string) properties.Property {
			t.Helper()
			p, err := properties.Lookup(c, key, vocabkey.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			return p
		}
		create := func(typeID []byte, label string, x, y int64) subjects.Subject {
			t.Helper()
			graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
			return subjectstest.At(t, c, userID, graph, typeID, label, subjects.Placement{GridX: x, GridY: y})
		}
		personProp := lookupProp("person")
		eventProp := lookupProp("event")
		placeProp := lookupProp("place")
		related := lookupProp("related_to")
		roleProp := lookupProp("role")
		relType := lookupProp("relationship_type")
		roleTerm, err := propertyterms.Lookup(c, roleProp.ID, "witness", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		relTerm, err := propertyterms.Lookup(c, relType.ID, "spouse", vocabkey.OriginProvenencia)
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
		typ, err := subjecttypes.Lookup(c, typeKey, vocabkey.OriginProvenencia)
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
		res, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
		if !bytes.Equal(got.GraphID, evidencegraphstest.SoleGraph(t, s.c, s.source.ID).ID) {
			t.Fatalf("bridge graph %x is not the Source's Default graph", got.GraphID)
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
		_, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("event place location", func(t *testing.T) {
		s := mustSeed(t)
		_, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("person place refused", func(t *testing.T) {
		s := mustSeed(t)
		before := countBridges(t, s.c, "location")
		_, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
		_, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
		_, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
		})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v want invalid", err)
		}
	})

	t.Run("attach existing citation", func(t *testing.T) {
		s := mustSeed(t)
		graph := evidencegraphstest.SoleGraph(t, s.c, s.source.ID).ID
		cited, _, err := observationstest.Cite(s.c, userID, graph, s.artifact.ID, validLocator)
		if err != nil {
			t.Fatal(err)
		}
		res, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
				FromSubjectID: s.event.ID,
				ToSubjectID:   s.place.ID,
				BridgeTypeKey: "location",
				CitationID:    cited.Citation.ID,
				Observations: []observations.Input{
					{PropertyID: s.eventProp.ID, ValueSubjectID: s.event.ID},
					{PropertyID: s.placeProp.ID, ValueSubjectID: s.place.ID},
				},
			})
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
		graph := evidencegraphstest.SoleGraph(t, s.c, s.source.ID).ID
		cited, _, err := observationstest.Cite(s.c, userID, graph, s.artifact.ID, validLocator)
		if err != nil {
			t.Fatal(err)
		}
		_, err = writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
		})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("got %v want invalid", err)
		}
	})

	// twoGraphs adds Artifact P2 and a Jones graph linked to P2 only (unlinked
	// from Default), with a Jones person and event.
	type twoGraph struct {
		seed
		p2          artifacts.Artifact
		jones       evidencegraphs.Graph
		jonesPerson subjects.Subject
		jonesEvent  subjects.Subject
	}
	mustTwoGraphs := func(t *testing.T) twoGraph {
		t.Helper()
		s := mustSeed(t)
		_, jones, p2 := artifactstest.Split(t, s.c, userID, s.source.ID, "Jones")
		create := func(key, label string, x, y int64) subjects.Subject {
			t.Helper()
			typ, err := subjecttypes.Lookup(s.c, key, vocabkey.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			return subjectstest.At(t, s.c, userID, jones.ID, typ.ID, label, subjects.Placement{GridX: x, GridY: y})
		}
		return twoGraph{
			seed: s, p2: p2, jones: jones,
			jonesPerson: create("person", "Ben", 8, 0),
			jonesEvent:  create("event", "Jones wedding", 10, 4),
		}
	}
	bridge := func(s seed, from, to subjects.Subject, in CreateInput) (Result, error) {
		in.FromSubjectID, in.ToSubjectID, in.BridgeTypeKey = from.ID, to.ID, "participation"
		in.Observations = []observations.Input{
			{PropertyID: s.personProp.ID, ValueSubjectID: s.person.ID},
			{PropertyID: s.eventProp.ID, ValueSubjectID: s.event.ID},
			{PropertyID: s.roleProp.ID, ValueTermID: s.roleTerm.ID},
		}
		return writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, in)
		})
	}

	t.Run("bridge joins its endpoints' graph", func(t *testing.T) {
		g := mustTwoGraphs(t)
		s := g.seed
		s.person, s.event = g.jonesPerson, g.jonesEvent
		res, err := bridge(s, g.jonesPerson, g.jonesEvent, CreateInput{
			Citation: citations.CreateInput{ArtifactID: g.p2.ID, LocatorJSON: validLocator},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(res.Subject.GraphID, g.jones.ID) || !bytes.Equal(res.Subject.SourceID, g.source.ID) {
			t.Fatalf("bridge graph %x source %x", res.Subject.GraphID, res.Subject.SourceID)
		}
		if !bytes.Equal(res.Citation.GraphID, g.jones.ID) {
			t.Fatalf("new Citation graph %x, want the endpoints' graph %x", res.Citation.GraphID, g.jones.ID)
		}
		db, err := g.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var stored []byte
		if err := db.QueryRow(`SELECT graph_id FROM citations WHERE id = ?`, res.Citation.ID).Scan(&stored); err != nil || !bytes.Equal(stored, g.jones.ID) {
			t.Fatalf("stored Citation graph %x, err %v", stored, err)
		}
	})

	t.Run("endpoints in two graphs refused", func(t *testing.T) {
		g := mustTwoGraphs(t)
		before := countBridges(t, g.c, "participation")
		s := g.seed
		s.event = g.jonesEvent
		_, err := bridge(s, g.person, g.jonesEvent, CreateInput{
			Citation: citations.CreateInput{ArtifactID: g.artifact.ID, LocatorJSON: validLocator},
		})
		if !errors.Is(err, observations.ErrOtherGraph) {
			t.Fatalf("got %v want ErrOtherGraph", err)
		}
		if n := countBridges(t, g.c, "participation"); n != before {
			t.Fatalf("bridges %d want %d", n, before)
		}
	})

	t.Run("attached citation in another graph refused", func(t *testing.T) {
		g := mustTwoGraphs(t)
		cited, _, err := observationstest.Cite(g.c, userID, g.jones.ID, g.p2.ID, validLocator)
		if err != nil {
			t.Fatal(err)
		}
		_, err = bridge(g.seed, g.person, g.event, CreateInput{CitationID: cited.Citation.ID})
		if !errors.Is(err, observations.ErrOtherGraph) {
			t.Fatalf("got %v want ErrOtherGraph", err)
		}
	})

	t.Run("missing position invalid", func(t *testing.T) {
		s := mustSeed(t)
		if _, err := writes.Call(s.c, writes.Op{}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjectpositions.Clear(tx, s.place.ID)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		_, err := writes.Call(s.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (Result, []rowchange.Change, error) {
			return CreateCitedBridge(tx, userID, CreateInput{
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
