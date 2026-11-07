package deleteimpact_test

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/autoreconciler/autoreconcilertest"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
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
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// A Subject emptied of its own Observations can be deleted while its claim
// still files it onto a handle with another member. The claim is removed
// explicitly and audited; the other member's claim keeps the handle.
type conclusionFixture struct {
	c         *database.Catalog
	userID    []byte
	a, b      subjects.Subject
	obsB      []byte
	entityID  []byte
	entityRef string
	ca        []byte
}

// newConclusionFixture: Subjects A and B on one Place, each with a toponym
// Observation; A grounds a handle and B joins it through Promote.
func newConclusionFixture(t *testing.T) conclusionFixture {
	t.Helper()
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
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Gazetteer"})
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
	toponym, err := properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	mkSubject := func(label string) subjects.Subject {
		t.Helper()
		graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
		s := subjectstest.OfType(t, c, userID, graph, place.ID, label)
		return s
	}
	a, b := mkSubject("York"), mkSubject("York (U.C.)")
	graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			GraphID:    graph,
			ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "York",
		}, []observations.Input{
			{SubjectID: a.ID, PropertyID: toponym.ID, Scalar: propertymeta.TextScalar("York")},
			{SubjectID: b.ID, PropertyID: toponym.ID, Scalar: propertymeta.TextScalar("York")},
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	obsB := res.Observations[1].ID

	grounding := promotetest.Save(t, c, userID, promote.Input{SubjectID: a.ID})
	promotetest.Save(t, c, userID, promote.Input{SubjectID: b.ID, EntityID: grounding.Entity.ID})

	return conclusionFixture{c: c, userID: userID, a: a, b: b, obsB: obsB,
		entityID: grounding.Entity.ID, entityRef: grounding.Entity.Ref, ca: grounding.Claim.ID}
}

func TestMemberDeleteEndToEnd(t *testing.T) {
	f := newConclusionFixture(t)
	c, userID, b, obsB, ca := f.c, f.userID, f.b, f.obsB, f.ca

	// B's own Observation blocks deleting B, so it goes first.
	report := mustImpact(t, c, catalogmodel.KindObservation, obsB)
	if !report.Allowed || len(report.Cascades) != 0 {
		t.Fatalf("%+v", report)
	}
	if _, err := writes.Call(c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := observations.Delete(tx, userID, obsB)
		return struct{}{}, changes, err
	}); err != nil {
		t.Fatal(err)
	}

	// B is now empty; its confirm names the handle it leaves.
	report = mustImpact(t, c, catalogmodel.KindSubject, b.ID)
	if !report.Allowed {
		t.Fatalf("%+v", report)
	}
	assertLeaves(t, report, f.entityRef)
	if _, err := writes.Call(c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := subjects.Delete(tx, userID, b.ID)
		return struct{}{}, changes, err
	}); err != nil {
		t.Fatal(err)
	}
	action, types := lastRevision(t, c)
	if action != "delete_subject" || strings.Join(types, ",") != "identity_claim,subject" {
		t.Fatalf("%s %v", action, types)
	}
	members, err := identityclaims.AcceptedMembers(c, f.entityID)
	if err != nil || len(members) != 1 || string(members[0].ID) != string(ca) {
		t.Fatalf("%v %+v", err, members)
	}
	autoreconcilertest.AssertCacheMatchesRebuild(t, c)
}

func TestLastMemberDeleteEndToEnd(t *testing.T) {
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
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	james := subjectstest.OfType(t, c, userID, graph, personType.ID, "James")
	promoted := promotetest.Save(t, c, userID, promote.Input{SubjectID: james.ID})
	entityID := promoted.Entity.ID
	entityUUID := database.UUIDString(entityID)

	persons, err := conclusionsummaries.ListPersons(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(persons) != 1 || persons[0].Entity.Ref != promoted.Entity.Ref {
		t.Fatalf("%+v", persons)
	}

	if _, err := writes.Call(c, writes.Op{Action: "unpromote_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		ch, err := promote.Unpromote(tx, userID, james.ID)
		return struct{}{}, ch, err
	}); err != nil {
		t.Fatal(err)
	}

	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var searchDocs int
	if err := db.QueryRow(`SELECT COUNT(*) FROM catalog_search_docs WHERE entity_id = ?`, entityUUID).Scan(&searchDocs); err != nil {
		t.Fatal(err)
	}
	if searchDocs != 0 {
		t.Fatalf("search docs %d", searchDocs)
	}
	node, err := c.Graph().Node(entityID)
	if err != nil {
		t.Fatal(err)
	}
	if node != nil {
		t.Fatalf("graph node still cached: %+v", node)
	}
	persons, err = conclusionsummaries.ListPersons(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(persons) != 0 {
		t.Fatalf("persons list %+v", persons)
	}
	var scopeCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transaction_scopes sc
		JOIN audit_transactions t ON t.id = sc.audit_transaction_id
		WHERE sc.scope_type = ? AND sc.scope_id = ? AND t.action_type = 'unpromote_subject'`,
		effects.ScopeCanonicalEntity, entityID).Scan(&scopeCount); err != nil {
		t.Fatal(err)
	}
	if scopeCount != 1 {
		t.Fatalf("scope rows %d", scopeCount)
	}
}
