package deleteimpact_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/vocabkey"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/subjectvocab/subjectvocabtest"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

const testLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

func TestImpactGates(t *testing.T) {
	c, userID := testCatalog(t)
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	if err := sourcecredibilitygrades.Install(c); err != nil {
		t.Fatal(err)
	}

	t.Run("unknown kind", func(t *testing.T) {
		if _, err := deleteimpact.ParseKind("nope"); !errors.Is(err, deleteimpact.ErrInvalid) {
			t.Fatalf("parse %v", err)
		}
		mustImpactErr(t, c, catalogmodel.Kind("nope"), fakeID(t), deleteimpact.ErrInvalid)
	})

	t.Run("missing id", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindCitation, fakeID(t))
		if got.Allowed || got.Gate != deleteimpact.GateNotFound || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("infra kinds", func(t *testing.T) {
		for _, kind := range []catalogmodel.Kind{catalogmodel.KindUser, catalogmodel.KindProject} {
			got := mustImpact(t, c, kind, fakeID(t))
			if got.Allowed || got.Gate != deleteimpact.GateInfra {
				t.Fatalf("%s %+v", kind, got)
			}
		}
	})

	t.Run("plugin vocab origin_locked", func(t *testing.T) {
		id, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "plugbook", Origin: "plugin:acme", Label: "Plug book",
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindSourceType, id)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("seeded property origin_locked", func(t *testing.T) {
		p, err := properties.Lookup(c, "name", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindProperty, p.ID)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("seeded term origin_locked", func(t *testing.T) {
		p, err := properties.Lookup(c, "event_type", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		term, err := propertyterms.Lookup(c, p.ID, "birth", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindPropertyTerm, term.ID)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("seeded unused source type allowed", func(t *testing.T) {
		id, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "unused_book", Origin: vocabkey.OriginProvenencia, Label: "Unused book",
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindSourceType, id)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("get works for subject type grade user term", func(t *testing.T) {
		person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindSubjectType, person.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("person type %+v", got)
		}
		grade, err := sourcecredibilitygrades.Lookup(c, "standard", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got = mustImpact(t, c, catalogmodel.KindCredibilityGrade, grade.ID)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("grade %+v", got)
		}
		p, err := properties.Lookup(c, "event_type", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		term := createTerm(t, c, userID, p.ID, "Land Grant")
		got = mustImpact(t, c, catalogmodel.KindPropertyTerm, term.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("user term %+v", got)
		}
	})

	t.Run("property with term row inbound", func(t *testing.T) {
		propID, err := properties.Upsert(c, properties.Property{
			Key: "custom_event", Origin: vocabkey.OriginUser,
			Label: "Custom Event", ValueType: propertymeta.ValueTypeTerm,
		})
		if err != nil {
			t.Fatal(err)
		}
		createTerm(t, c, userID, propID, "Grant")
		got := mustImpact(t, c, catalogmodel.KindProperty, propID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("%+v", got)
		}
		if len(got.Groups) != 1 || got.Groups[0].Via != "property_terms.property_id" || got.Groups[0].Total != 1 {
			t.Fatalf("groups %+v", got.Groups)
		}
		// Term rows deep-link to the parent property's Properties inspector row.
		if len(got.Groups[0].Listed) != 1 {
			t.Fatalf("listed %+v", got.Groups[0].Listed)
		}
		loc := got.Groups[0].Listed[0].Location
		if loc.Section != "properties" || loc.PropertyID != idString(propID) || loc.SubjectTypeKey != "" {
			t.Fatalf("term location %+v", loc)
		}
		if loc.Title != "Grant" || got.Groups[0].Listed[0].Title != "Grant" {
			t.Fatalf("term title %+v", got.Groups[0].Listed[0])
		}
	})

	t.Run("plugin field origin_locked", func(t *testing.T) {
		id, err := metadatafields.Upsert(c, metadatafields.Field{
			Key: "plug_author", Origin: "plugin:acme", Label: "Plug author", DataType: metadatafields.DataTypeText,
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindMetadataField, id)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("%+v", got)
		}
	})
}

func TestImpactCitationAndSubject(t *testing.T) {
	c, userID := testCatalog(t)
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
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	placeType, err := subjecttypes.Lookup(c, "place", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	eventType, err := subjecttypes.Lookup(c, "event", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	place, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			GraphID: graph, SubjectTypeID: placeType.ID, Label: "Leeds",
		}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	alice, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			GraphID: graph, SubjectTypeID: personType.ID, Label: "Alice",
		}, &subjects.Placement{GridX: 0, GridY: 0})
	})
	if err != nil {
		t.Fatal(err)
	}
	wedding, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			GraphID: graph, SubjectTypeID: eventType.ID, Label: "Wedding",
		}, &subjects.Placement{GridX: 4, GridY: 4})
	})
	if err != nil {
		t.Fatal(err)
	}
	toponym, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("citation with observations", func(t *testing.T) {
		res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "Leeds",
			}, []observations.Input{{
				SubjectID: place.ID, PropertyID: toponym.ID, ValueText: "Leeds", HasText: true,
			}})
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindCitation, res.Citation.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound || len(got.Groups) != 1 {
			t.Fatalf("%+v", got)
		}
		g := got.Groups[0]
		if g.Via != "observations.citation_id" || g.Kind != catalogmodel.KindObservation || g.Total != 1 || len(g.Listed) != 1 {
			t.Fatalf("%+v", g)
		}
		listed := g.Listed[0]
		if listed.Ref != res.Observations[0].Ref {
			t.Fatalf("ref %q", listed.Ref)
		}
		if listed.Location.SourceSurface != "citationComposer" || listed.Location.SourceID != idString(src.ID) {
			t.Fatalf("loc %+v", listed.Location)
		}
		if listed.Location.CitationID != idString(res.Citation.ID) || listed.Location.ObservationID == "" {
			t.Fatalf("ids %+v", listed.Location)
		}
	})

	t.Run("propertyterms delete uses impact", func(t *testing.T) {
		eventTypeProp, err := properties.Lookup(c, "event_type", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		term := createTerm(t, c, userID, eventTypeProp.ID, "Land Grant")
		if _, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: testLocator,
			}, []observations.Input{{
				SubjectID: wedding.ID, PropertyID: eventTypeProp.ID, ValueTermID: term.ID,
			}})
		}); err != nil {
			t.Fatal(err)
		}
		_, _, err = writes.Run(c, writes.Op{Action: "delete_property_term", UserID: userID},
			func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := propertyterms.Delete(tx, userID, term.ID)
				return struct{}{}, changes, err
			})
		if !errors.Is(err, propertyterms.ErrInUse) {
			t.Fatalf("delete %v", err)
		}
	})

	t.Run("citation inbound cap 20", func(t *testing.T) {
		inputs := make([]observations.Input, 21)
		for i := range inputs {
			inputs[i] = observations.Input{
				SubjectID: place.ID, PropertyID: toponym.ID,
				ValueText: fmt.Sprintf("name-%02d", i), HasText: true,
			}
		}
		res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: testLocator,
			}, inputs)
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindCitation, res.Citation.ID)
		if got.Allowed || len(got.Groups) != 1 {
			t.Fatalf("%+v", got)
		}
		g := got.Groups[0]
		if g.Total != 21 || len(g.Listed) != 20 {
			t.Fatalf("total=%d listed=%d", g.Total, len(g.Listed))
		}
		for i := 1; i < len(g.Listed); i++ {
			if g.Listed[i].Ref < g.Listed[i-1].Ref {
				t.Fatalf("not ordered by ref %q %q", g.Listed[i-1].Ref, g.Listed[i].Ref)
			}
		}
	})

	t.Run("G2 value_subject_id", func(t *testing.T) {
		roleProp, err := properties.Lookup(c, "role", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		personProp, err := properties.Lookup(c, "person", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		eventProp, err := properties.Lookup(c, "event", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		roleTerm, err := propertyterms.Lookup(c, roleProp.ID, "witness", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		bridge, err := writes.Call(c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
			return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
				FromSubjectID: alice.ID, ToSubjectID: wedding.ID,
				BridgeTypeKey: "participation",
				Citation:      citations.CreateInput{ArtifactID: art.ID, LocatorJSON: testLocator},
				Observations: []observations.Input{
					{PropertyID: personProp.ID, ValueSubjectID: alice.ID},
					{PropertyID: eventProp.ID, ValueSubjectID: wedding.ID},
					{PropertyID: roleProp.ID, ValueTermID: roleTerm.ID},
				},
			})
		})
		got := mustImpact(t, c, catalogmodel.KindSubject, alice.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("alice %+v", got)
		}
		found := false
		for _, g := range got.Groups {
			if g.Via == "observations.value_subject_id" && g.Total > 0 {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing value_subject_id group %+v", got.Groups)
		}

		t.Run("bridge only connection facets allowed", func(t *testing.T) {
			got := mustImpact(t, c, catalogmodel.KindSubject, bridge.Subject.ID)
			if !got.Allowed || got.Gate != deleteimpact.GateOK {
				t.Fatalf("bridge %+v", got)
			}
			tx, err := c.BeginWrite()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = tx.Rollback()
				c.EndWrite()
			}()
			if _, err := tx.Exec(`DELETE FROM subjects WHERE id = ?`, bridge.Subject.ID); err == nil {
				t.Fatal("raw subject delete should fail while facets remain")
			}
			if _, err := deleteimpact.ReleaseFacets(tx, catalogmodel.KindSubject, bridge.Subject.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(`DELETE FROM subjects WHERE id = ?`, bridge.Subject.ID); err != nil {
				t.Fatalf("after release %v", err)
			}
		})

		t.Run("bridge extra observation inbound", func(t *testing.T) {
			subjectvocabtest.Assign(t, c, userID, bridge.Subject.SubjectTypeID, toponym.ID)
			if _, err := writes.Call(c, writes.Op{Action: "add_observations", UserID: userID}, func(tx *database.Tx) ([]observations.Observation, []rowchange.Change, error) {
				return observations.AddToCitation(tx, userID, bridge.Citation.ID, []observations.Input{{
					SubjectID: bridge.Subject.ID, PropertyID: toponym.ID, ValueText: "extra", HasText: true,
				}})
			}); err != nil {
				t.Fatal(err)
			}
			got := mustImpact(t, c, catalogmodel.KindSubject, bridge.Subject.ID)
			if got.Allowed || got.Gate != deleteimpact.GateInbound {
				t.Fatalf("%+v", got)
			}
			found := false
			for _, g := range got.Groups {
				if g.Via == "observations.subject_id" && g.Total >= 1 {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing subject_id group %+v", got.Groups)
			}
		})

		t.Run("edge observation edge_locked", func(t *testing.T) {
			var edgeID []byte
			for _, o := range bridge.Observations {
				if impact := mustImpact(t, c, catalogmodel.KindObservation, o.ID); impact.Gate == deleteimpact.GateEdgeLocked {
					edgeID = o.ID
					if impact.Allowed {
						t.Fatalf("allowed edge %+v", impact)
					}
					break
				}
			}
			if len(edgeID) == 0 {
				t.Fatal("no edge observation")
			}
			if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := observations.Delete(tx, userID, edgeID)
				return struct{}{}, changes, err
			}); !errors.Is(err, observations.ErrEdgeLocked) {
				t.Fatalf("delete edge %v", err)
			}
		})
	})

	t.Run("user role property on person is inbound not a facet", func(t *testing.T) {
		userRole, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
			return properties.Create(tx, userID, "Role", propertymeta.ValueTypeText, "", "")
		})
		if err != nil {
			t.Fatal(err)
		}
		if userRole.Key != "role" || userRole.Origin != vocabkey.OriginUser {
			t.Fatalf("user role %+v", userRole)
		}
		subjectvocabtest.Assign(t, c, userID, personType.ID, userRole.ID)
		res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: testLocator,
			}, []observations.Input{{
				SubjectID: alice.ID, PropertyID: userRole.ID, ValueText: "witness", HasText: true,
			}})
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindSubject, alice.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("alice %+v", got)
		}
		found := false
		for _, g := range got.Groups {
			if g.Via == "observations.subject_id" && g.Total >= 1 {
				found = true
				if len(g.Listed) == 0 || g.Listed[0].Ref != res.Observations[0].Ref {
					t.Fatalf("listed %+v want %q", g.Listed, res.Observations[0].Ref)
				}
			}
		}
		if !found {
			t.Fatalf("missing subject_id group %+v", got.Groups)
		}
		if _, err := writes.Call(c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjects.Delete(tx, userID, alice.ID)
			return struct{}{}, changes, err
		}); !errors.Is(err, subjects.ErrInUse) {
			t.Fatalf("delete %v", err)
		}
		listed, err := observations.ListBySubject(c, alice.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(listed) == 0 {
			t.Fatal("user role observation gone")
		}
	})

	t.Run("seeded role bound to person is inbound not a facet", func(t *testing.T) {
		roleProp, err := properties.Lookup(c, "role", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		subjectvocabtest.Assign(t, c, userID, personType.ID, roleProp.ID)
		roleTerm, err := propertyterms.Lookup(c, roleProp.ID, "witness", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
		bob, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{
				GraphID: graph, SubjectTypeID: personType.ID, Label: "Bob",
			}, &subjects.Placement{GridX: 8, GridY: 0})
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: testLocator,
			}, []observations.Input{{
				SubjectID: bob.ID, PropertyID: roleProp.ID, ValueTermID: roleTerm.ID,
			}})
		}); err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, catalogmodel.KindSubject, bob.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("bob %+v", got)
		}
		found := false
		for _, g := range got.Groups {
			if g.Via == "observations.subject_id" && g.Total >= 1 {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing subject_id group %+v", got.Groups)
		}
		if _, err := writes.Call(c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjects.Delete(tx, userID, bob.ID)
			return struct{}{}, changes, err
		}); !errors.Is(err, subjects.ErrInUse) {
			t.Fatalf("delete %v", err)
		}
		listed, err := observations.ListBySubject(c, bob.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(listed) == 0 {
			t.Fatal("seeded role observation gone")
		}
	})
}

func TestSnapshotOwnedThenRelease(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Deed"})
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}

	fileA, err := files.NewID()
	if err != nil {
		t.Fatal(err)
	}
	const sum = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := files.Insert(tx, files.File{
		ID: fileA, ChecksumSHA256: sum, OriginalFilename: "share.jpg", MediaType: "image/jpeg", ByteSize: 4,
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	art1, err := artifactstest.Create(c, userID, artifacts.CreateInput{
		SourceID: src.ID, FileID: fileA, Label: "Share 1",
	})
	if err != nil {
		t.Fatal(err)
	}
	art2, err := artifactstest.Create(c, userID, artifacts.CreateInput{
		SourceID: src.ID, FileID: fileA, Label: "Share 2",
	})
	if err != nil {
		t.Fatal(err)
	}

	wtx, err := c.BeginWrite()
	if err != nil {
		t.Fatal(err)
	}
	snap1, err := deleteimpact.SnapshotOwned(wtx.Tx, catalogmodel.KindArtifact, art1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wtx.Exec(`DELETE FROM artifacts WHERE id = ?`, art1.ID); err != nil {
		t.Fatal(err)
	}
	if err := deleteimpact.ReleaseSnapshot(wtx, snap1); err != nil {
		t.Fatal(err)
	}
	var left int
	if err := wtx.QueryRow(`SELECT COUNT(*) FROM files WHERE id = ?`, fileA).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 1 {
		t.Fatalf("shared file removed early left=%d", left)
	}
	if err := wtx.Commit(); err != nil {
		t.Fatal(err)
	}
	c.EndWrite()

	wtx, err = c.BeginWrite()
	if err != nil {
		t.Fatal(err)
	}
	snap2, err := deleteimpact.SnapshotOwned(wtx.Tx, catalogmodel.KindArtifact, art2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wtx.Exec(`DELETE FROM artifacts WHERE id = ?`, art2.ID); err != nil {
		t.Fatal(err)
	}
	if err := deleteimpact.ReleaseSnapshot(wtx, snap2); err != nil {
		t.Fatal(err)
	}
	if err := wtx.QueryRow(`SELECT COUNT(*) FROM files WHERE id = ?`, fileA).Scan(&left); err != nil {
		t.Fatal(err)
	}
	if left != 0 {
		t.Fatalf("last pointer left files=%d", left)
	}
	if err := wtx.Commit(); err != nil {
		t.Fatal(err)
	}
	c.EndWrite()
}

func TestImpactSourceAndArtifact(t *testing.T) {
	c, userID := testCatalog(t)
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
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty source allowed", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindSource, src.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
	})

	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("source blocked by artifact", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindSource, src.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("%+v", got)
		}
		if len(got.Groups) != 1 || got.Groups[0].Via != "artifacts.source_id" || got.Groups[0].Total != 1 {
			t.Fatalf("groups %+v", got.Groups)
		}
		if got.Groups[0].Listed[0].Ref != art.Ref {
			t.Fatalf("listed %+v", got.Groups[0].Listed)
		}
	})

	t.Run("empty artifact allowed", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindArtifact, art.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("%+v", got)
		}
	})

	cit, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: testLocator,
			Transcription: "John Hartley, worsted weaver, of 14 Back Nile Street",
		}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("artifact blocked by citation", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindArtifact, art.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("%+v", got)
		}
		if len(got.Groups) != 1 || got.Groups[0].Via != "citations.artifact_id" || got.Groups[0].Total != 1 {
			t.Fatalf("groups %+v", got.Groups)
		}
		if got.Groups[0].Listed[0].Ref != cit.Citation.Ref {
			t.Fatalf("listed %+v", got.Groups[0].Listed)
		}
		if got.Groups[0].Listed[0].Title != "John Hartley, worsted weaver, of 14 Back Nile Street" {
			t.Fatalf("title %q", got.Groups[0].Listed[0].Title)
		}
		loc := got.Groups[0].Listed[0].Location
		if loc.Section != "sources" || loc.SourceSurface != "citationComposer" || loc.SourceID == "" || loc.CitationID == "" || loc.ArtifactID == "" {
			t.Fatalf("citation location %+v", loc)
		}
	})

	empty, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Uncited"})
	})
	if err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	graph := evidencegraphstest.SoleGraph(t, c, empty.ID).ID
	sub, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			GraphID: graph, SubjectTypeID: personType.ID, Label: "Alice",
		}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("source blocked by subject", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindSource, empty.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("%+v", got)
		}
		if len(got.Groups) != 1 || got.Groups[0].Via != "subjects.source_id" || got.Groups[0].Total != 1 {
			t.Fatalf("groups %+v", got.Groups)
		}
		if got.Groups[0].Listed[0].Ref != sub.Ref {
			t.Fatalf("listed %+v", got.Groups[0].Listed)
		}
	})
}

func TestImpactEvidenceGraph(t *testing.T) {
	c, userID := testCatalog(t)
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
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	sub, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{
			GraphID: graph, SubjectTypeID: personType.ID, Label: "Alice",
		}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	def := evidencegraphstest.SoleGraph(t, c, src.ID)

	t.Run("a Source's only graph is last_graph", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindEvidenceGraph, def.ID)
		if got.Allowed || got.Gate != deleteimpact.GateLastGraph || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
	})

	jones := evidencegraphstest.Create(t, c, userID, src.ID, "Jones", art.ID)

	t.Run("a graph with Subjects is blocked by them", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindEvidenceGraph, def.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("%+v", got)
		}
		if len(got.Groups) != 1 || got.Groups[0].Via != "subjects.graph_id" || got.Groups[0].Total != 1 {
			t.Fatalf("groups %+v", got.Groups)
		}
		listed := got.Groups[0].Listed
		if len(listed) != 1 || listed[0].Ref != sub.Ref || listed[0].Location.Section == "" {
			t.Fatalf("listed %+v", listed)
		}
	})

	t.Run("an empty second graph is allowed", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindEvidenceGraph, jones.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("unknown graph", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindEvidenceGraph, fakeID(t))
		if got.Allowed || got.Gate != deleteimpact.GateNotFound {
			t.Fatalf("%+v", got)
		}
	})
}

func TestImpactAgreesWithSQLite(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Deed"})
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty source raw delete matches Impact", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindSource, src.ID)
		if !got.Allowed {
			t.Fatalf("%+v", got)
		}
		if !rawDeleteOK(t, c, "sources", src.ID) {
			t.Fatal("raw DELETE of unused source failed")
		}
	})

	if _, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"}); err != nil {
		t.Fatal(err)
	}

	t.Run("inbound source raw delete matches Impact", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindSource, src.ID)
		if got.Allowed {
			t.Fatalf("%+v", got)
		}
		if rawDeleteOK(t, c, "sources", src.ID) {
			t.Fatal("raw DELETE succeeded with artifact inbound")
		}
	})
}

func rawDeleteOK(t *testing.T, c *database.Catalog, table string, id []byte) bool {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(`DELETE FROM `+table+` WHERE id = ?`, id)
	return err == nil
}

func mustImpact(t *testing.T, c *database.Catalog, kind catalogmodel.Kind, id []byte) deleteimpact.Report {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	got, err := deleteimpact.Impact(tx, kind, id)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func mustImpactErr(t *testing.T, c *database.Catalog, kind catalogmodel.Kind, id []byte, want error) {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	_, err = deleteimpact.Impact(tx, kind, id)
	if !errors.Is(err, want) {
		t.Fatalf("got %v want %v", err, want)
	}
}

func fakeID(t *testing.T) []byte {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return id[:]
}

func idString(id []byte) string {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}

func createTerm(t *testing.T, c *database.Catalog, userID, propertyID []byte, label string) propertyterms.Term {
	t.Helper()
	term, _, err := writes.Run(c, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, propertyID, label, "")
		})
	if err != nil {
		t.Fatal(err)
	}
	return term
}

func testCatalog(t *testing.T) (*database.Catalog, []byte) {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
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

func TestListedGroup(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	var rows []deleteimpact.Listed
	for i := 0; i < 25; i++ {
		res, err := writes.Call(c, writes.Op{Action: "create_citation", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: testLocator,
			}, nil)
		})
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, deleteimpact.Listed{ID: res.Citation.ID, Ref: res.Citation.Ref})
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	t.Run("three rows are projected", func(t *testing.T) {
		g, err := deleteimpact.ListedGroup(tx, "x.via", catalogmodel.KindCitation, rows[:3])
		if err != nil {
			t.Fatal(err)
		}
		if g.Via != "x.via" || g.Kind != catalogmodel.KindCitation || g.Total != 3 || len(g.Listed) != 3 {
			t.Fatalf("%+v", g)
		}
		for i, l := range g.Listed {
			if l.Ref != rows[i].Ref || string(l.ID) != string(rows[i].ID) {
				t.Fatalf("row %d: %+v", i, l)
			}
			if l.Title == "" || l.Location.SourceSurface != "citationComposer" || l.Location.CitationID != idString(rows[i].ID) {
				t.Fatalf("row %d projection: %+v", i, l)
			}
		}
	})
	t.Run("more than the cap lists the first 20", func(t *testing.T) {
		g, err := deleteimpact.ListedGroup(tx, "x.via", catalogmodel.KindCitation, rows)
		if err != nil {
			t.Fatal(err)
		}
		if g.Total != 25 || len(g.Listed) != 20 {
			t.Fatalf("total %d listed %d", g.Total, len(g.Listed))
		}
		for i, l := range g.Listed {
			if l.Ref != rows[i].Ref {
				t.Fatalf("row %d out of order: %q want %q", i, l.Ref, rows[i].Ref)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		g, err := deleteimpact.ListedGroup(tx, "x.via", catalogmodel.KindCitation, nil)
		if err != nil {
			t.Fatal(err)
		}
		if g.Total != 0 || len(g.Listed) != 0 || g.Via != "x.via" {
			t.Fatalf("%+v", g)
		}
	})
	t.Run("nil tx", func(t *testing.T) {
		if _, err := deleteimpact.ListedGroup(nil, "x.via", catalogmodel.KindCitation, rows); !errors.Is(err, deleteimpact.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})
}

// A Subject, an Observation, a Citation and a graph name their Evidence graph
// in the location the delete report links to.
func TestLocationsNameTheirGraph(t *testing.T) {
	c, userID := testCatalog(t)
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
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"}); err != nil {
		t.Fatal(err)
	}
	def, jones, page := artifactstest.Split(t, c, userID, src.ID, "Jones")
	subject := artifactstest.SubjectOn(t, c, userID, jones.ID, "place", "Leeds")
	toponym, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	cited, err := writes.Call(c, writes.Op{Action: "create_citation", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: page.ID, LocatorJSON: testLocator,
		}, []observations.Input{{SubjectID: subject.ID, PropertyID: toponym.ID, ValueText: "Leeds", HasText: true}})
	})
	if err != nil {
		t.Fatal(err)
	}
	bare, err := writes.Call(c, writes.Op{Action: "create_citation", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: page.ID, LocatorJSON: testLocator,
		}, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	loc := func(kind catalogmodel.Kind, id []byte, ref string) deleteimpact.Location {
		t.Helper()
		g, err := deleteimpact.ListedGroup(tx, "x.via", kind, []deleteimpact.Listed{{ID: id, Ref: ref}})
		if err != nil {
			t.Fatal(err)
		}
		return g.Listed[0].Location
	}
	for _, tt := range []struct {
		name     string
		kind     catalogmodel.Kind
		id       []byte
		ref      string
		surface  string
		graphID  []byte
		sourceID []byte
	}{
		{"Subject", catalogmodel.KindSubject, subject.ID, subject.Ref, "graph", jones.ID, src.ID},
		{"Observation", catalogmodel.KindObservation, cited.Observations[0].ID, cited.Observations[0].Ref, "citationComposer", jones.ID, src.ID},
		{"Citation with an Observation", catalogmodel.KindCitation, cited.Citation.ID, cited.Citation.Ref, "citationComposer", jones.ID, src.ID},
		{"Citation with no Observation", catalogmodel.KindCitation, bare.Citation.ID, bare.Citation.Ref, "citationComposer", jones.ID, src.ID},
		{"graph", catalogmodel.KindEvidenceGraph, jones.ID, jones.Ref, "graph", jones.ID, src.ID},
		{"Default graph", catalogmodel.KindEvidenceGraph, def.ID, def.Ref, "graph", def.ID, src.ID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := loc(tt.kind, tt.id, tt.ref)
			if got.SourceSurface != tt.surface || got.GraphID != idString(tt.graphID) || got.SourceID != idString(tt.sourceID) {
				t.Fatalf("location %+v, want %s on graph %x of %x", got, tt.surface, tt.graphID, tt.sourceID)
			}
		})
	}
}
