package deleteimpact_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
)

func TestImpactConclusion(t *testing.T) {
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
	place, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	york, err := subjects.Create(c, userID, subjects.CreateInput{
		SourceID: src.ID, SubjectTypeID: place.ID, Label: "York",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
		ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "York",
	}, []observations.Input{{
		SubjectID: york.ID, PropertyID: toponym.ID, ValueText: "York", HasText: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	obsID := res.Observations[0].ID
	plc, err := canonicalentities.Create(c, userID, canonicalentities.CreateInput{SubjectTypeID: place.ID})
	if err != nil {
		t.Fatal(err)
	}
	claim, err := identityclaims.Create(c, userID, identityclaims.CreateInput{
		SubjectID: york.ID, EntityID: plc.ID, Status: identityclaims.StatusAccepted,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("subject type in use names its handle", func(t *testing.T) {
		got := mustImpact(t, c, deleteimpact.KindSubjectType, place.ID)
		if got.Allowed {
			t.Fatalf("%+v", got)
		}
		g := findGroup(t, got, "canonical_entities.subject_type_id")
		if g.Kind != deleteimpact.KindCanonicalEntity || g.Total != 1 || g.Listed[0].Ref != plc.Ref {
			t.Fatalf("%+v", g)
		}
	})

	t.Run("blocked promoted subject reports groups and the handle it leaves", func(t *testing.T) {
		got := mustImpact(t, c, deleteimpact.KindSubject, york.ID)
		if got.Allowed || len(got.Groups) == 0 {
			t.Fatalf("%+v", got)
		}
		assertLeaves(t, got, plc.Ref)
	})

	t.Run("allowed promoted subject still names the handle it leaves", func(t *testing.T) {
		leeds, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: place.ID, Label: "Leeds",
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindSubject, leeds.ID)
		if !got.Allowed || len(got.Cascades) != 0 {
			t.Fatalf("unpromoted %+v", got)
		}
		res, err := promote.Save(c, userID, promote.Input{SubjectID: leeds.ID})
		if err != nil {
			t.Fatal(err)
		}
		got = mustImpact(t, c, deleteimpact.KindSubject, leeds.ID)
		if !got.Allowed || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
		assertLeaves(t, got, res.Entity.Ref)
	})

	t.Run("pinned observation names the handle its claim files onto", func(t *testing.T) {
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO identity_claim_evidence (identity_claim_id, observation_id) VALUES (?, ?)`,
			claim.ID, obsID); err != nil {
			t.Fatal(err)
		}
		got := mustImpact(t, c, deleteimpact.KindObservation, obsID)
		if got.Allowed {
			t.Fatalf("%+v", got)
		}
		g := findGroup(t, got, "identity_claim_evidence.observation_id")
		if g.Kind != deleteimpact.KindCanonicalEntity || g.Total != 1 || g.Listed[0].Ref != plc.Ref {
			t.Fatalf("%+v", g)
		}
		if rawDeleteOK(t, c, "observations", obsID) {
			t.Fatal("raw DELETE succeeded with a pin inbound")
		}
	})
}

func assertLeaves(t *testing.T, r deleteimpact.Report, handleRef string) {
	t.Helper()
	if len(r.Cascades) != 1 {
		t.Fatalf("cascades %+v", r.Cascades)
	}
	g := r.Cascades[0]
	if g.Via != "identity_claims.subject_id" || g.Kind != deleteimpact.KindCanonicalEntity ||
		g.Total != 1 || len(g.Listed) != 1 || g.Listed[0].Ref != handleRef {
		t.Fatalf("%+v", g)
	}
}

func findGroup(t *testing.T, r deleteimpact.Report, via string) deleteimpact.Group {
	t.Helper()
	for _, g := range r.Groups {
		if g.Via == via {
			return g
		}
	}
	t.Fatalf("no group %s in %+v", via, r)
	return deleteimpact.Group{}
}
