package deleteimpact_test

import (
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"

	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
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

// A Subject emptied of its own Observations can be deleted while its claim
// still pins another member's evidence (a confirmed comparison). Those pins
// are removed explicitly and audited; the other member's claim keeps its own.
type conclusionFixture struct {
	c                *database.Catalog
	userID           []byte
	a, b             subjects.Subject
	obsA, obsB       []byte
	entityID, ca, cb []byte
}

// newConclusionFixture: Subjects A and B on one Place, each with a toponym
// Observation, and a confirmed comparison with backfill — both Observations
// pinned on both claims.
func newConclusionFixture(t *testing.T) conclusionFixture {
	t.Helper()
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
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Gazetteer"})
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
	toponym, err := properties.Lookup(c, "toponym", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	mkSubject := func(label string) subjects.Subject {
		t.Helper()
		s, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: place.ID, Label: label,
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	a, b := mkSubject("York"), mkSubject("York (U.C.)")
	res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
		ArtifactID: art.ID, LocatorJSON: testLocator, Transcription: "York",
	}, []observations.Input{
		{SubjectID: a.ID, PropertyID: toponym.ID, ValueText: "York", HasText: true},
		{SubjectID: b.ID, PropertyID: toponym.ID, ValueText: "York", HasText: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	obsA, obsB := res.Observations[0].ID, res.Observations[1].ID

	grounding, err := promote.Save(c, userID, promote.Input{SubjectID: a.ID})
	if err != nil {
		t.Fatal(err)
	}
	joined, err := identityclaims.Create(c, userID, identityclaims.CreateInput{
		SubjectID: b.ID, EntityID: grounding.Entity.ID, Status: identityclaims.StatusAccepted,
	})
	if err != nil {
		t.Fatal(err)
	}
	ca, cb := grounding.Claim.ID, joined.ID

	// A confirmed comparison with backfill: both Observations pinned on both claims.
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	for _, claimID := range [][]byte{ca, cb} {
		for _, obsID := range [][]byte{obsA, obsB} {
			if _, err := db.Exec(`INSERT INTO identity_claim_evidence (identity_claim_id, observation_id) VALUES (?, ?)`,
				claimID, obsID); err != nil {
				t.Fatal(err)
			}
		}
	}
	return conclusionFixture{c: c, userID: userID, a: a, b: b, obsA: obsA, obsB: obsB,
		entityID: grounding.Entity.ID, ca: ca, cb: cb}
}

func TestSubjectDeleteReleasesOwnClaimPins(t *testing.T) {
	f := newConclusionFixture(t)
	c, userID, b, obsA, obsB, ca, cb := f.c, f.userID, f.b, f.obsA, f.obsB, f.ca, f.cb
	pins := func(claimID, obsID []byte) int {
		return countRows(t, c, `SELECT COUNT(*) FROM identity_claim_evidence
			WHERE identity_claim_id = ? AND observation_id = ?`, claimID, obsID)
	}

	// B's own Observation blocks deleting B, so it goes first, taking its pins off both claims.
	if err := observations.Delete(c, userID, obsB); err != nil {
		t.Fatal(err)
	}
	if pins(ca, obsB)+pins(cb, obsB) != 0 {
		t.Fatal("obsB pins left")
	}

	// B is now empty; its claim CB still pins A's Observation.
	if pins(cb, obsA) != 1 {
		t.Fatal("setup: CB should still pin obsA")
	}
	if err := subjects.Delete(c, userID, b.ID); err != nil {
		t.Fatal(err)
	}
	action, types := lastRevision(t, c)
	if action != "delete_subject" || strings.Join(types, ",") != "identity_claim_evidence,identity_claim,subject" {
		t.Fatalf("%s %v", action, types)
	}
	if pins(cb, obsA) != 0 {
		t.Fatal("CB pin on obsA left")
	}
	if pins(ca, obsA) != 1 {
		t.Fatal("A's claim lost its own pin on obsA")
	}
	if n := countRows(t, c, `SELECT COUNT(*) FROM observations WHERE id = ?`, obsA); n != 1 {
		t.Fatal("obsA should be untouched")
	}
	members, err := identityclaims.AcceptedMembers(c, f.entityID)
	if err != nil || len(members) != 1 || string(members[0].ID) != string(ca) {
		t.Fatalf("%v %+v", err, members)
	}
}
