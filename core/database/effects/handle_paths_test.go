package effects_test

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

const locator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// TestHandlePaths resolves Handles against a catalog. No write calls Handles
// yet; these checks are the rules the registry states.
//
// membersOf keeps accepted and provisional claims and drops rejected. A
// citation edit recomputes handles only when transcription_uncertain is on
// the diff. onStatus runs when the old or the new status is accepted, and a
// status that is never accepted does not recompute observers. A date value
// walks inbound to the observation that points at it. value_date_id is
// unique, so that is one observation today.
func TestHandlePaths(t *testing.T) {
	userID := []byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 9, 8, 7, 6, 5, 4, 3}
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
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "A"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	placeType, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	mk := func(typeID []byte, label string, x int64) subjects.Subject {
		t.Helper()
		s, err := subjects.Create(c, userID, subjects.CreateInput{
			SourceID: src.ID, SubjectTypeID: typeID, Label: label,
		}, &subjects.Placement{GridX: x, GridY: 0})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	alice := mk(personType.ID, "Alice", 0)
	bob := mk(personType.ID, "Bob", 4)
	cara := mk(personType.ID, "Cara", 8)
	drew := mk(personType.ID, "Drew", 12)
	eve := mk(personType.ID, "Eve", 14)
	boston := mk(placeType.ID, "Boston", 16)

	aliceRes, err := promote.Save(c, userID, promote.Input{SubjectID: alice.ID})
	if err != nil {
		t.Fatal(err)
	}
	bobRes, err := promote.Save(c, userID, promote.Input{SubjectID: bob.ID})
	if err != nil {
		t.Fatal(err)
	}
	eveRes, err := promote.Save(c, userID, promote.Input{SubjectID: eve.ID})
	if err != nil {
		t.Fatal(err)
	}
	bostonRes, err := promote.Save(c, userID, promote.Input{SubjectID: boston.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := identityclaims.Create(c, userID, identityclaims.CreateInput{
		SubjectID: cara.ID, EntityID: aliceRes.Entity.ID, Status: identityclaims.StatusRejected,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := identityclaims.Create(c, userID, identityclaims.CreateInput{
		SubjectID: drew.ID, EntityID: aliceRes.Entity.ID, Status: identityclaims.StatusProvisional,
	}); err != nil {
		t.Fatal(err)
	}

	personProp, err := properties.Lookup(c, "person", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	dateProp, err := properties.Lookup(c, "start_date", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	year := 1842
	created, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
		ArtifactID: art.ID, LocatorJSON: locator,
	}, []observations.Input{{
		SubjectID: boston.ID, PropertyID: dateProp.ID,
		Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year},
	}})
	if err != nil {
		t.Fatal(err)
	}
	dateID := created.Observations[0].ValueDateID
	if len(dateID) != 16 {
		t.Fatal("date value was not stored")
	}

	// Bob's observation points at Alice. Person subjects are not bound to a
	// subject-valued property, so this row is inserted directly: the registry
	// reads observations, and the binding rule is the writer's.
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	insertPointing := func(subjectID []byte) {
		t.Helper()
		id, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		r, err := ref.Mint(ref.PrefixObservation)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO observations (
			id, ref, citation_id, subject_id, property_id, polarity, value_subject_id
		) VALUES (?, ?, ?, ?, ?, 'positive', ?)`,
			id[:], r, created.Citation.ID, subjectID, personProp.ID, alice.ID); err != nil {
			t.Fatal(err)
		}
	}
	insertPointing(bob.ID)
	insertPointing(eve.ID)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	t.Run("members keep accepted and provisional", func(t *testing.T) {
		got := mustHandles(t, tx, observationSubject(alice.ID))
		if !idSet(got, aliceRes.Entity.ID) {
			t.Fatalf("alice handles %d", len(got))
		}
		got = mustHandles(t, tx, observationSubject(drew.ID))
		if !idSet(got, aliceRes.Entity.ID) {
			t.Fatalf("provisional handles %d", len(got))
		}
		got = mustHandles(t, tx, observationSubject(cara.ID))
		if len(got) != 0 {
			t.Fatalf("rejected claim resolved %d handles", len(got))
		}
	})

	t.Run("transcription_uncertain only", func(t *testing.T) {
		note := rowchange.Change{
			EntityType: "citation", EntityID: created.Citation.ID, Action: rowchange.ActionUpdate,
			Fields: map[string]rowchange.FieldDiff{
				"transcription_note": {Old: "", New: "ink faded"},
			},
		}
		if got := mustHandles(t, tx, note); len(got) != 0 {
			t.Fatalf("note edit resolved %d handles", len(got))
		}
		uncertain := note
		uncertain.Fields = map[string]rowchange.FieldDiff{
			"transcription_uncertain": {Old: false, New: true},
		}
		got := mustHandles(t, tx, uncertain)
		if !idSet(got, bostonRes.Entity.ID, bobRes.Entity.ID, eveRes.Entity.ID) {
			t.Fatalf("uncertain edit resolved %d handles", len(got))
		}
	})

	t.Run("accepted on old or new status", func(t *testing.T) {
		base := rowchange.Change{
			EntityType: "identity_claim", EntityID: aliceRes.Claim.ID, Action: rowchange.ActionUpdate,
		}
		argument := base
		argument.Fields = map[string]rowchange.FieldDiff{"argument": {Old: "", New: "sure"}}
		if got := mustHandles(t, tx, argument); !idSet(got, aliceRes.Entity.ID) {
			t.Fatalf("argument edit resolved %d handles", len(got))
		}
		acceptedToRejected := base
		acceptedToRejected.Fields = map[string]rowchange.FieldDiff{
			"status": {Old: identityclaims.StatusAccepted, New: identityclaims.StatusRejected},
		}
		if got := mustHandles(t, tx, acceptedToRejected); !idSet(got, aliceRes.Entity.ID, bobRes.Entity.ID, eveRes.Entity.ID) {
			t.Fatalf("accepted→rejected resolved %d handles", len(got))
		}
		rejectedToAccepted := base
		rejectedToAccepted.Fields = map[string]rowchange.FieldDiff{
			"status": {Old: identityclaims.StatusRejected, New: identityclaims.StatusAccepted},
		}
		if got := mustHandles(t, tx, rejectedToAccepted); !idSet(got, aliceRes.Entity.ID, bobRes.Entity.ID, eveRes.Entity.ID) {
			t.Fatalf("rejected→accepted resolved %d handles", len(got))
		}
		neither := base
		neither.Fields = map[string]rowchange.FieldDiff{
			"status": {Old: identityclaims.StatusProvisional, New: identityclaims.StatusRejected},
		}
		if got := mustHandles(t, tx, neither); !idSet(got, aliceRes.Entity.ID) {
			t.Fatalf("status that is never accepted resolved %d handles", len(got))
		}
	})

	t.Run("date value reaches its observation", func(t *testing.T) {
		ch := rowchange.Change{
			EntityType: "date_value", EntityID: dateID, Action: rowchange.ActionUpdate,
			Fields: map[string]rowchange.FieldDiff{"start_year": {Old: 1842, New: 1843}},
		}
		got := mustHandles(t, tx, ch)
		if !idSet(got, bostonRes.Entity.ID) {
			t.Fatalf("date value resolved %d handles", len(got))
		}
	})
}

func dashed(id []byte) string {
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}

func observationSubject(subjectID []byte) rowchange.Change {
	return rowchange.Change{
		EntityType: "observation",
		EntityID:   []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Action:     rowchange.ActionUpdate,
		Fields: map[string]rowchange.FieldDiff{
			"subject_id": {New: dashed(subjectID)},
		},
	}
}

func mustHandles(t *testing.T, tx *sql.Tx, changes ...rowchange.Change) [][]byte {
	t.Helper()
	got, err := effects.Handles(tx, changes)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func idSet(got [][]byte, want ...[]byte) bool {
	if len(got) != len(want) {
		return false
	}
	for _, id := range want {
		found := false
		for _, other := range got {
			if bytes.Equal(id, other) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
