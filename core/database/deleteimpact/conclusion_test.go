package deleteimpact_test

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestImpactConclusion(t *testing.T) {
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
	place, err := subjecttypes.Lookup(c, "place", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	york := subjectstest.OfType(t, c, userID, graph, place.ID, "York")
	toponym, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			GraphID:    graph,
			ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "York",
		}, []observations.Input{{
			SubjectID: york.ID, PropertyID: toponym.ID, ValueText: "York", HasText: true,
		}})
	})
	if err != nil {
		t.Fatal(err)
	}
	obsID := res.Observations[0].ID
	plc, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: place.ID})
	})
	if err != nil {
		t.Fatal(err)
	}
	claim, err := writes.Call(c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
		return identityclaims.Create(tx, userID, identityclaims.CreateInput{
			SubjectID: york.ID, EntityID: plc.ID, Status: identityclaims.StatusAccepted,
		})
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

	t.Run("blocked promoted subject reports groups and the handle it leaves", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindSubject, york.ID)
		if got.Allowed || len(got.Groups) == 0 {
			t.Fatalf("%+v", got)
		}
		assertLeaves(t, got, plc.Ref)
	})

	t.Run("allowed promoted subject still names the handle it leaves", func(t *testing.T) {
		graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
		leeds := subjectstest.OfType(t, c, userID, graph, place.ID, "Leeds")
		got := mustImpact(t, c, catalogmodel.KindSubject, leeds.ID)
		if !got.Allowed || len(got.Cascades) != 0 {
			t.Fatalf("unpromoted %+v", got)
		}
		res := promotetest.Save(t, c, userID, promote.Input{SubjectID: leeds.ID})
		got = mustImpact(t, c, catalogmodel.KindSubject, leeds.ID)
		if !got.Allowed || len(got.Groups) != 0 {
			t.Fatalf("%+v", got)
		}
		assertLeaves(t, got, res.Entity.Ref)
	})

	t.Run("a member's observation names no handle", func(t *testing.T) {
		got := mustImpact(t, c, catalogmodel.KindObservation, obsID)
		if !got.Allowed || got.Gate != deleteimpact.GateOK || len(got.Groups) != 0 || len(got.Cascades) != 0 {
			t.Fatalf("%+v", got)
		}
	})

	t.Run("observation delete audits only the observation", func(t *testing.T) {
		if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := observations.Delete(tx, userID, obsID)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		action, types := lastRevision(t, c)
		if action != "delete_observation" || strings.Join(types, ",") != "observation" {
			t.Fatalf("%s %v", action, types)
		}
		if _, err := identityclaims.Get(c, claim.ID); err != nil {
			t.Fatalf("claim should stay as written: %v", err)
		}
	})

	t.Run("subject delete removes its claim explicitly and audits it; handle stays", func(t *testing.T) {
		if _, err := writes.Call(c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjects.Delete(tx, userID, york.ID)
			return struct{}{}, changes, err
		}); err != nil {
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
