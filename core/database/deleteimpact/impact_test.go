package deleteimpact_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
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
		mustImpactErr(t, c, deleteimpact.Kind("nope"), fakeID(t), deleteimpact.ErrInvalid)
	})

	t.Run("missing id", func(t *testing.T) {
		got := mustImpact(t, c, deleteimpact.KindCitation, fakeID(t))
		if got.Allowed || got.Gate != deleteimpact.GateNotFound || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("infra kinds", func(t *testing.T) {
		for _, kind := range []deleteimpact.Kind{deleteimpact.KindUser, deleteimpact.KindProject} {
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
		got := mustImpact(t, c, deleteimpact.KindSourceType, id)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("seeded property origin_locked", func(t *testing.T) {
		p, err := properties.Lookup(c, "name", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindProperty, p.ID)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("seeded term origin_locked", func(t *testing.T) {
		p, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		term, err := propertyterms.Lookup(c, p.ID, "birth", propertyterms.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindPropertyTerm, term.ID)
		if got.Allowed || got.Gate != deleteimpact.GateOriginLocked {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("seeded unused source type allowed", func(t *testing.T) {
		id, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "unused_book", Origin: sourcetypes.OriginProvenencia, Label: "Unused book",
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindSourceType, id)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("get works for subject type grade user term", func(t *testing.T) {
		person, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindSubjectType, person.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("person type %+v", got)
		}
		grade, err := sourcecredibilitygrades.Lookup(c, "standard", sourcecredibilitygrades.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		got = mustImpact(t, c, deleteimpact.KindCredibilityGrade, grade.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("grade %+v", got)
		}
		p, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		term, err := propertyterms.Create(c, userID, p.ID, "Land Grant", "")
		if err != nil {
			t.Fatal(err)
		}
		got = mustImpact(t, c, deleteimpact.KindPropertyTerm, term.ID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK {
			t.Fatalf("user term %+v", got)
		}
	})

	t.Run("property with term row inbound", func(t *testing.T) {
		propID, err := properties.Upsert(c, properties.Property{
			Key: "custom_event", Origin: properties.OriginUser,
			Label: "Custom Event", ValueType: properties.ValueTypeTerm,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := propertyterms.Create(c, userID, propID, "Grant", ""); err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindProperty, propID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound {
			t.Fatalf("%+v", got)
		}
		if len(got.Groups) != 1 || got.Groups[0].Via != "property_terms.property_id" || got.Groups[0].Total != 1 {
			t.Fatalf("groups %+v", got.Groups)
		}
	})

	t.Run("plugin field origin_locked", func(t *testing.T) {
		id, err := sourcefields.Upsert(c, sourcefields.Field{
			Key: "plug_author", Origin: "plugin:acme", Label: "Plug author", DataType: sourcefields.DataTypeText,
		})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindSourceField, id)
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
	personType, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	eventType, err := subjecttypes.Lookup(c, "event", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	place, err := subjects.Create(c, userID, subjects.CreateInput{
		SourceID: src.ID, SubjectTypeID: placeType.ID, Label: "Leeds",
	}, nil)
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
	toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("citation with observations", func(t *testing.T) {
		res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "Leeds",
		}, []observations.Input{{
			SubjectID: place.ID, PropertyID: toponym.ID, ValueText: "Leeds", HasText: true,
		}})
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindCitation, res.Citation.ID)
		if got.Allowed || got.Gate != deleteimpact.GateInbound || len(got.Groups) != 1 {
			t.Fatalf("%+v", got)
		}
		g := got.Groups[0]
		if g.Via != "observations.citation_id" || g.Kind != deleteimpact.KindObservation || g.Total != 1 || len(g.Listed) != 1 {
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
		eventTypeProp, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		term, err := propertyterms.Create(c, userID, eventTypeProp.ID, "Land Grant", "")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: testLocator,
		}, []observations.Input{{
			SubjectID: wedding.ID, PropertyID: eventTypeProp.ID, ValueTermID: term.ID,
		}}); err != nil {
			t.Fatal(err)
		}
		if err := propertyterms.Delete(c, userID, term.ID); !errors.Is(err, propertyterms.ErrInUse) {
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
		res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: testLocator,
		}, inputs)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindCitation, res.Citation.ID)
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
		got := mustImpact(t, c, deleteimpact.KindSubject, alice.ID)
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
			got := mustImpact(t, c, deleteimpact.KindSubject, bridge.Subject.ID)
			if !got.Allowed || got.Gate != deleteimpact.GateOK {
				t.Fatalf("bridge %+v", got)
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
			if _, err := tx.Exec(`DELETE FROM subjects WHERE id = ?`, bridge.Subject.ID); err == nil {
				t.Fatal("raw subject delete should fail while facets remain")
			}
			if err := deleteimpact.ReleaseConnectionFacets(tx, bridge.Subject.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(`DELETE FROM subjects WHERE id = ?`, bridge.Subject.ID); err != nil {
				t.Fatalf("after release %v", err)
			}
		})

		t.Run("bridge extra observation inbound", func(t *testing.T) {
			if err := subjectvocab.AppendBinding(c, bridge.Subject.SubjectTypeID, toponym.ID); err != nil {
				t.Fatal(err)
			}
			if _, err := observations.AddToCitation(c, userID, bridge.Citation.ID, []observations.Input{{
				SubjectID: bridge.Subject.ID, PropertyID: toponym.ID, ValueText: "extra", HasText: true,
			}}); err != nil {
				t.Fatal(err)
			}
			got := mustImpact(t, c, deleteimpact.KindSubject, bridge.Subject.ID)
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
				if impact := mustImpact(t, c, deleteimpact.KindObservation, o.ID); impact.Gate == deleteimpact.GateEdgeLocked {
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
			if err := observations.Delete(c, userID, edgeID); !errors.Is(err, observations.ErrEdgeLocked) {
				t.Fatalf("delete edge %v", err)
			}
		})
	})
}

func mustImpact(t *testing.T, c *database.Catalog, kind deleteimpact.Kind, id []byte) deleteimpact.Report {
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

func mustImpactErr(t *testing.T, c *database.Catalog, kind deleteimpact.Kind, id []byte, want error) {
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
