package deleteimpact_test

import (
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"

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
		got := mustImpact(t, c, catalogmodel.KindSubjectType, place.ID)
		if got.Allowed {
			t.Fatalf("%+v", got)
		}
		g := findGroup(t, got, "canonical_entities.subject_type_id")
		if g.Kind != catalogmodel.KindCanonicalEntity || g.Total != 1 || g.Listed[0].Ref != plc.Ref {
			t.Fatalf("%+v", g)
		}
	})

	var otherClaimID []byte // a second handle's claim, for cross-claim pins

	t.Run("blocked promoted subject reports groups and the handle it leaves", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindSubject, york.ID)
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
		got := mustImpact(t, c, catalogmodel.KindSubject, leeds.ID)
		if !got.Allowed || len(got.Cascades) != 0 {
			t.Fatalf("unpromoted %+v", got)
		}
		res, err := promote.Save(c, userID, promote.Input{SubjectID: leeds.ID})
		if err != nil {
			t.Fatal(err)
		}
		got = mustImpact(t, c, catalogmodel.KindSubject, leeds.ID)
		if !got.Allowed || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
		assertLeaves(t, got, res.Entity.Ref)
		otherClaimID = res.Claim.ID
	})

	t.Run("pinned observation is allowed and names the handle its claims weaken", func(t *testing.T) {
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		// Pinned on its own member's claim and, as backfill would, on another claim.
		for _, cid := range [][]byte{claim.ID, otherClaimID} {
			if _, err := db.Exec(`INSERT INTO identity_claim_evidence (identity_claim_id, observation_id) VALUES (?, ?)`,
				cid, obsID); err != nil {
				t.Fatal(err)
			}
		}
		got := mustImpact(t, c, catalogmodel.KindObservation, obsID)
		if !got.Allowed || len(got.Groups) != 0 || len(got.Cascades) != 1 {
			t.Fatalf("%+v", got)
		}
		g := got.Cascades[0]
		if g.Via != "identity_claim_evidence.observation_id" || g.Kind != catalogmodel.KindCanonicalEntity ||
			g.Total != 2 || len(g.Listed) != 2 {
			t.Fatalf("%+v", g)
		}
	})

	t.Run("observation delete removes its pins explicitly and audits them", func(t *testing.T) {
		if err := observations.Delete(c, userID, obsID); err != nil {
			t.Fatal(err)
		}
		action, types := lastRevision(t, c)
		if action != "delete_observation" ||
			strings.Join(types, ",") != "identity_claim_evidence,identity_claim_evidence,observation" {
			t.Fatalf("%s %v", action, types)
		}
		if n := countRows(t, c, `SELECT COUNT(*) FROM identity_claim_evidence WHERE observation_id = ?`, obsID); n != 0 {
			t.Fatalf("pins left %d", n)
		}
		if _, err := identityclaims.Get(c, claim.ID); err != nil {
			t.Fatalf("claim should stay (weaker): %v", err)
		}
	})

	t.Run("subject delete removes its claim explicitly and audits it; handle stays", func(t *testing.T) {
		if err := subjects.Delete(c, userID, york.ID); err != nil {
			t.Fatal(err)
		}
		action, types := lastRevision(t, c)
		if action != "delete_subject" || strings.Join(types, ",") != "identity_claim,subject" {
			t.Fatalf("%s %v", action, types)
		}
		if _, err := canonicalentities.Get(c, plc.ID); err != nil {
			t.Fatalf("handle gone: %v", err)
		}
	})
}

// lastRevision returns the newest revision's action and its change entity types in insert order.
func lastRevision(t *testing.T, c *database.Catalog) (string, []string) {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT t.action_type, ch.entity_type
		FROM audit_changes ch JOIN audit_transactions t ON t.id = ch.audit_transaction_id
		WHERE t.revision = (SELECT MAX(revision) FROM audit_transactions)
		ORDER BY ch.rowid`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var action string
	var types []string
	for rows.Next() {
		var entityType string
		if err := rows.Scan(&action, &entityType); err != nil {
			t.Fatal(err)
		}
		types = append(types, entityType)
	}
	return action, types
}

func countRows(t *testing.T, c *database.Catalog, q string, args ...any) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func assertLeaves(t *testing.T, r deleteimpact.Report, handleRef string) {
	t.Helper()
	if len(r.Cascades) != 1 {
		t.Fatalf("cascades %+v", r.Cascades)
	}
	g := r.Cascades[0]
	if g.Via != "identity_claims.subject_id" || g.Kind != catalogmodel.KindCanonicalEntity ||
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
