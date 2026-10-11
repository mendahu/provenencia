package effects_test

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

// TestConclusionHandleSets builds the change lists promote, claim create,
// credibility, and a cardinality edit actually produce, on a transaction that
// rolls back, and checks effects.Handles covers the set those writes recompute
// by hand. A short set is a registry bug. A strict superset is kept.
func TestConclusionHandleSets(t *testing.T) {
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
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
	if err := sourcecredibilitygrades.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID},
		func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "A"})
		})
	if err != nil {
		t.Fatal(err)
	}
	art, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
		})
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
	seenWith, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
		return properties.Create(tx, userID, "Seen with", properties.ValueTypeSubject, "", "")
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO subject_type_properties (subject_type_id, property_id, sort_order) VALUES (?, ?, 99)`,
		personType.ID, seenWith.ID,
	); err != nil {
		t.Fatal(err)
	}
	mk := func(typeID []byte, label string, x int64) subjects.Subject {
		t.Helper()
		s, _, err := writes.Run(c, writes.Op{Action: "create_subject", UserID: userID},
			func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
				return subjects.Create(tx, userID, subjects.CreateInput{
					SourceID: src.ID, SubjectTypeID: typeID, Label: label,
				}, &subjects.Placement{GridX: x, GridY: 0})
			})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	ada := mk(personType.ID, "Ada", 0)
	bob := mk(personType.ID, "Bob", 8)
	birth := mk(eventType.ID, "Birth", 16)
	nameProp, err := properties.Lookup(c, "name", properties.OriginProvenencia)
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
	roleProp, err := properties.Lookup(c, "role", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	role, err := propertyterms.Lookup(c, roleProp.ID, "subject", propertyterms.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := writes.Run(c, writes.Op{Action: "create_citation_with_observations", UserID: userID},
		func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: evidenceLocator,
			}, []observations.Input{
				{SubjectID: ada.ID, PropertyID: nameProp.ID, Name: &namevalues.Value{Form: "Ada"}},
				{SubjectID: ada.ID, PropertyID: seenWith.ID, ValueSubjectID: bob.ID},
			})
		}); err != nil {
		t.Fatal(err)
	}
	adaPromoted, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, promote.Input{SubjectID: ada.ID})
	})
	if err != nil {
		t.Fatal(err)
	}
	bridge := func(from, to subjects.Subject) {
		t.Helper()
		_, _, err := writes.Run(c, writes.Op{Action: "create_cited_bridge", UserID: userID},
			func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
				return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
					SourceID:      src.ID,
					FromSubjectID: from.ID,
					ToSubjectID:   to.ID,
					Citation:      citations.CreateInput{ArtifactID: art.ID, LocatorJSON: evidenceLocator},
					Observations: []observations.Input{
						{PropertyID: personProp.ID, ValueSubjectID: from.ID},
						{PropertyID: eventProp.ID, ValueSubjectID: to.ID},
						{PropertyID: roleProp.ID, ValueTermID: role.ID},
					},
				})
			})
		if err != nil {
			t.Fatal(err)
		}
	}
	bridge(ada, birth)
	bridge(bob, birth)

	grades, err := sourcecredibilitygrades.List(c)
	if err != nil || len(grades) < 2 {
		t.Fatalf("grades: %v len=%d", err, len(grades))
	}
	open := func() (*database.Tx, func()) {
		t.Helper()
		sqlTx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		return &database.Tx{Tx: sqlTx}, func() { _ = sqlTx.Rollback() }
	}
	// covers fails when got is missing an id the hand recompute would touch.
	// Extras are a superset and are logged, not failed.
	covers := func(got, want [][]byte) {
		t.Helper()
		for _, id := range want {
			ok := false
			for _, g := range got {
				if bytes.Equal(id, g) {
					ok = true
					break
				}
			}
			if !ok {
				t.Fatalf("handles %d, hand set %d, missing one", len(got), len(want))
			}
		}
		if len(got) > len(dedupeIDs(want)) {
			t.Logf("registry superset: handles %d, hand set %d", len(got), len(dedupeIDs(want)))
		}
	}
	claim := func(tx *sql.Tx, subject subjects.Subject, entityID []byte, status string, file bool) (canonicalentities.Entity, [][]byte, []rowchange.Change) {
		t.Helper()
		var changes []rowchange.Change
		var entity canonicalentities.Entity
		if len(entityID) == 0 {
			var ch rowchange.Change
			var err error
			entity, ch, err = canonicalentities.InsertTx(tx, canonicalentities.CreateInput{SubjectTypeID: subject.SubjectTypeID})
			if err != nil {
				t.Fatal(err)
			}
			entityID = entity.ID
			changes = append(changes, ch)
		} else {
			entity.ID = append([]byte(nil), entityID...)
		}
		_, ch, err := identityclaims.InsertTx(tx, identityclaims.CreateInput{
			SubjectID: subject.ID, EntityID: entityID, Status: status,
		})
		if err != nil {
			t.Fatal(err)
		}
		changes = append(changes, ch)
		var assocs [][]byte
		if file {
			var err error
			assocs, changes, err = identityclaims.AppendFiling(tx, subject.ID, changes)
			if err != nil {
				t.Fatal(err)
			}
		}
		return entity, assocs, changes
	}

	t.Run("accepted mint covers claim handle, filings, and observers", func(t *testing.T) {
		tx, done := open()
		defer done()
		entity, assocs, changes := claim(tx.Tx, bob, nil, identityclaims.StatusAccepted, true)
		if len(assocs) != 0 {
			t.Fatalf("unpromoted other end filed %d associations", len(assocs))
		}
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		obs, err := autoreconciler.HandlesObservingSubject(tx.Tx, bob.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(obs) == 0 {
			t.Fatal("expected an observer of the claimed subject")
		}
		want := append([][]byte{entity.ID}, obs...)
		covers(got, want)
	})

	t.Run("join covers the existing handle", func(t *testing.T) {
		tx, done := open()
		defer done()
		entity, _, changes := claim(tx.Tx, bob, adaPromoted.Entity.ID, identityclaims.StatusAccepted, true)
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		obs, err := autoreconciler.HandlesObservingSubject(tx.Tx, bob.ID)
		if err != nil {
			t.Fatal(err)
		}
		covers(got, append([][]byte{entity.ID}, obs...))
	})

	t.Run("provisional covers only its handle", func(t *testing.T) {
		tx, done := open()
		defer done()
		entity, _, changes := claim(tx.Tx, bob, nil, identityclaims.StatusProvisional, false)
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		if !sameIDs(got, [][]byte{entity.ID}) {
			t.Fatalf("provisional handles = %d, want 1", len(got))
		}
	})

	t.Run("filing the other end covers the association", func(t *testing.T) {
		tx, done := open()
		defer done()
		entity, assocs, changes := claim(tx.Tx, birth, nil, identityclaims.StatusAccepted, true)
		if len(assocs) == 0 {
			t.Fatal("expected the participation to file")
		}
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		obs, err := autoreconciler.HandlesObservingSubject(tx.Tx, birth.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := append([][]byte{entity.ID}, assocs...)
		want = append(want, obs...)
		covers(got, want)
	})

	t.Run("batch files both ends", func(t *testing.T) {
		tx, done := open()
		defer done()
		bobEntity, bobChange, err := canonicalentities.InsertTx(tx.Tx, canonicalentities.CreateInput{SubjectTypeID: bob.SubjectTypeID})
		if err != nil {
			t.Fatal(err)
		}
		_, bobClaim, err := identityclaims.InsertTx(tx.Tx, identityclaims.CreateInput{
			SubjectID: bob.ID, EntityID: bobEntity.ID, Status: identityclaims.StatusAccepted,
		})
		if err != nil {
			t.Fatal(err)
		}
		birthEntity, birthChange, err := canonicalentities.InsertTx(tx.Tx, canonicalentities.CreateInput{SubjectTypeID: birth.SubjectTypeID})
		if err != nil {
			t.Fatal(err)
		}
		_, birthClaim, err := identityclaims.InsertTx(tx.Tx, identityclaims.CreateInput{
			SubjectID: birth.ID, EntityID: birthEntity.ID, Status: identityclaims.StatusAccepted,
		})
		if err != nil {
			t.Fatal(err)
		}
		assocs, filed, err := identityclaims.FileSourceBridgesTx(tx.Tx, src.ID, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(assocs) == 0 {
			t.Fatal("expected both ends to file")
		}
		changes := []rowchange.Change{bobChange, bobClaim, birthChange, birthClaim}
		changes = append(changes, filed...)
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		bobObs, err := autoreconciler.HandlesObservingSubject(tx.Tx, bob.ID)
		if err != nil {
			t.Fatal(err)
		}
		birthObs, err := autoreconciler.HandlesObservingSubject(tx.Tx, birth.ID)
		if err != nil {
			t.Fatal(err)
		}
		want := append([][]byte{bobEntity.ID, birthEntity.ID}, assocs...)
		want = append(want, bobObs...)
		want = append(want, birthObs...)
		covers(got, want)
	})

	t.Run("grade change matches source handles", func(t *testing.T) {
		tx, done := open()
		defer done()
		id, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		changes := []rowchange.Change{{
			EntityType: "source_credibility_assessment",
			EntityID:   id[:],
			Action:     rowchange.ActionUpdate,
			Fields: map[string]rowchange.FieldDiff{
				"source_id":            {New: mustUUID(t, src.ID)},
				"credibility_grade_id": {New: mustUUID(t, grades[0].ID)},
			},
		}}
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		want, err := autoreconciler.HandlesForSource(tx.Tx, src.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(want) == 0 {
			t.Fatal("expected a handle for the source")
		}
		if !sameIDs(got, want) {
			t.Fatalf("grade handles %d, source set %d", len(got), len(want))
		}
	})

	t.Run("argument only resolves no handles", func(t *testing.T) {
		tx, done := open()
		defer done()
		id, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		got, err := effects.Handles(tx.Tx, []rowchange.Change{{
			EntityType: "source_credibility_assessment",
			EntityID:   id[:],
			Action:     rowchange.ActionUpdate,
			Fields: map[string]rowchange.FieldDiff{
				"source_id": {New: mustUUID(t, src.ID)},
				"argument":  {New: "A note"},
			},
		}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("argument handles = %d", len(got))
		}
	})

	t.Run("cardinality matches property handles", func(t *testing.T) {
		tx, done := open()
		defer done()
		got, err := effects.Handles(tx.Tx, []rowchange.Change{{
			EntityType: "property",
			EntityID:   append([]byte(nil), nameProp.ID...),
			Action:     rowchange.ActionUpdate,
			Fields: map[string]rowchange.FieldDiff{
				"cardinality": {Old: properties.CardinalitySingle, New: properties.CardinalityMultiple},
			},
		}})
		if err != nil {
			t.Fatal(err)
		}
		want, err := autoreconciler.HandlesForProperty(tx.Tx, nameProp.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(want) == 0 {
			t.Fatal("expected a handle for the property")
		}
		if !sameIDs(got, want) {
			t.Fatalf("cardinality handles %d, property set %d", len(got), len(want))
		}
	})

	t.Run("label edit resolves no handles", func(t *testing.T) {
		tx, done := open()
		defer done()
		got, err := effects.Handles(tx.Tx, []rowchange.Change{{
			EntityType: "property",
			EntityID:   append([]byte(nil), nameProp.ID...),
			Action:     rowchange.ActionUpdate,
			Fields: map[string]rowchange.FieldDiff{
				"label": {Old: "Name", New: "Full name"},
			},
		}})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("label handles = %d", len(got))
		}
	})
}

func mustUUID(t *testing.T, id []byte) string {
	t.Helper()
	u, err := uuid.FromBytes(id)
	if err != nil {
		t.Fatal(err)
	}
	return u.String()
}

func dedupeIDs(ids [][]byte) [][]byte {
	seen := map[string]struct{}{}
	var out [][]byte
	for _, id := range ids {
		if len(id) != 16 {
			continue
		}
		if _, ok := seen[string(id)]; ok {
			continue
		}
		seen[string(id)] = struct{}{}
		out = append(out, id)
	}
	return out
}
