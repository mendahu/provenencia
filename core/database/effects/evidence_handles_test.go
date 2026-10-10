package effects_test

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/namevalues"
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

const evidenceLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// TestEvidenceHandleSets builds the change an evidence write actually returns
// and checks effects.Handles against HandlesForSubjects and HandlesForCitation.
// A short set is a registry bug: Run would recompute less than those sets.
func TestEvidenceHandleSets(t *testing.T) {
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
	placeType, err := subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	seenWith, err := writes.Call(c, writes.Op{Action: "create_property", UserID: userID}, func(tx *database.Tx) (properties.Property, []rowchange.Change, error) {
		return properties.Create(tx, userID, "Seen with", properties.ValueTypeSubject, "", "")
	})
	if err != nil {
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
	person := mk(personType.ID, "Ada", 0)
	place := mk(placeType.ID, "York", 4)
	alice := mk(personType.ID, "Alice", 12)
	bob := mk(personType.ID, "Bob", 16)
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
	for _, s := range []subjects.Subject{person, place} {
		if _, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
			return promote.Save(tx, userID, promote.Input{SubjectID: s.ID})
		}); err != nil {
			t.Fatal(err)
		}
	}

	nameProp, err := properties.Lookup(c, "name", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	dateProp, err := properties.Lookup(c, "start_date", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	year := 1842
	created, _, err := writes.Run(c, writes.Op{Action: "create_citation_with_observations", UserID: userID},
		func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				ArtifactID: art.ID, LocatorJSON: evidenceLocator,
			}, []observations.Input{
				{SubjectID: person.ID, PropertyID: nameProp.ID, Name: &namevalues.Value{Form: "Ada"}},
				{SubjectID: place.ID, PropertyID: dateProp.ID, Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &year}},
				{SubjectID: person.ID, PropertyID: seenWith.ID, ValueSubjectID: alice.ID},
			})
		})
	if err != nil {
		t.Fatal(err)
	}
	var nameObs, dateObs, seenObs observations.Observation
	for _, o := range created.Observations {
		switch {
		case bytes.Equal(o.PropertyID, nameProp.ID):
			nameObs = o
		case bytes.Equal(o.PropertyID, dateProp.ID):
			dateObs = o
		case bytes.Equal(o.PropertyID, seenWith.ID):
			seenObs = o
		}
	}
	open := func() (*database.Tx, func()) {
		t.Helper()
		sqlTx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		return &database.Tx{Tx: sqlTx}, func() { _ = sqlTx.Rollback() }
	}
	expectSubjects := func(tx *sql.Tx, changes []rowchange.Change, subjectID []byte) {
		t.Helper()
		got, err := effects.Handles(tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		want, err := autoreconciler.HandlesForSubjects(tx, [][]byte{subjectID})
		if err != nil {
			t.Fatal(err)
		}
		if len(want) == 0 {
			t.Fatal("expected a handle for the subject")
		}
		if !sameIDs(got, want) {
			t.Fatalf("handles %d, recompute set %d", len(got), len(want))
		}
	}

	t.Run("name", func(t *testing.T) {
		tx, done := open()
		defer done()
		_, changes, err := observations.Update(tx, userID, observations.Input{
			ID: nameObs.ID, SubjectID: person.ID, PropertyID: nameProp.ID,
			Name: &namevalues.Value{Form: "Adelaide"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !hasEntity(changes, "name_value") {
			t.Fatalf("name edit changes = %#v", changes)
		}
		expectSubjects(tx.Tx, changes, person.ID)
	})

	t.Run("date", func(t *testing.T) {
		tx, done := open()
		defer done()
		nextYear := 1843
		_, changes, err := observations.Update(tx, userID, observations.Input{
			ID: dateObs.ID, SubjectID: place.ID, PropertyID: dateProp.ID,
			Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &nextYear},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !hasEntity(changes, "date_value") {
			t.Fatalf("date edit changes = %#v", changes)
		}
		expectSubjects(tx.Tx, changes, place.ID)
	})

	t.Run("polarity", func(t *testing.T) {
		tx, done := open()
		defer done()
		_, changes, err := observations.Update(tx, userID, observations.Input{
			ID: nameObs.ID, SubjectID: person.ID, PropertyID: nameProp.ID,
			Polarity: observations.PolarityNegative, ValueNameID: nameObs.ValueNameID,
		})
		if err != nil {
			t.Fatal(err)
		}
		ch := onlyEntity(t, changes, "observation")
		if _, ok := ch.Fields["subject_id"]; ok {
			t.Fatal("polarity edit included subject_id")
		}
		if _, ok := ch.Fields["polarity"]; !ok {
			t.Fatal("polarity edit omitted polarity")
		}
		expectSubjects(tx.Tx, changes, person.ID)
	})

	t.Run("value subject", func(t *testing.T) {
		tx, done := open()
		defer done()
		_, changes, err := observations.Update(tx, userID, observations.Input{
			ID: seenObs.ID, SubjectID: person.ID, PropertyID: seenWith.ID,
			ValueSubjectID: bob.ID,
		})
		if err != nil {
			t.Fatal(err)
		}
		ch := onlyEntity(t, changes, "observation")
		if _, ok := ch.Fields["subject_id"]; ok {
			t.Fatal("value subject move included subject_id")
		}
		diff, ok := ch.Fields["value_subject_id"]
		if !ok || diff.Old == nil || diff.New == nil {
			t.Fatalf("value_subject_id diff = %#v", ch.Fields)
		}
		expectSubjects(tx.Tx, changes, person.ID)
	})

	t.Run("citation certainty", func(t *testing.T) {
		tx, done := open()
		defer done()
		_, changes, err := citations.Update(tx, userID, created.Citation.ID, citations.CitationFieldsInput{
			LocatorJSON: evidenceLocator, TranscriptionUncertain: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		want, err := autoreconciler.HandlesForCitation(tx.Tx, created.Citation.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(want) == 0 || !sameIDs(got, want) {
			t.Fatalf("citation handles %d, recompute set %d", len(got), len(want))
		}
	})

	t.Run("locator only", func(t *testing.T) {
		tx, done := open()
		defer done()
		_, changes, err := citations.Update(tx, userID, created.Citation.ID, citations.CitationFieldsInput{
			LocatorJSON: `{"version":1,"selectors":[{"type":"page","artifact_page":2}]}`,
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := effects.Handles(tx.Tx, changes)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 0 {
			t.Fatalf("locator-only handles = %d", len(got))
		}
	})
}

func hasEntity(changes []rowchange.Change, entity string) bool {
	for _, ch := range changes {
		if ch.EntityType == entity {
			return true
		}
	}
	return false
}

func onlyEntity(t *testing.T, changes []rowchange.Change, entity string) rowchange.Change {
	t.Helper()
	var found rowchange.Change
	n := 0
	for _, ch := range changes {
		if ch.EntityType == entity {
			found = ch
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%s changes = %d (%#v)", entity, n, changes)
	}
	return found
}

func sameIDs(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for _, id := range a {
		ok := false
		for _, other := range b {
			if bytes.Equal(id, other) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}
