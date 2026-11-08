package conclusionsummaries_test

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// A country three steps above a town is a dependent of that town, of a birth
// located there, and of the person whose birth it is. Saving the country's
// name drops the stored person row, so the next People list shows the new name.
func TestDisplayCacheFollowsAPlaceThreeStepsUp(t *testing.T) {
	f := newFixture(t)
	country := f.placeSubject("Country", nil, nil, 0, 0)
	region := f.placeSubject("Region", nil, nil, 2, 0)
	city := f.placeSubject("City", nil, nil, 4, 0)
	town := f.placeSubject("Town", nil, nil, 6, 0)
	f.placeRel(region, country, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(city, region, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(town, city, propertyterms.KeyPartOf, nil, nil)

	person := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(person, "Ada Town")
	f.at(person, 8, 0)
	birth := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, birth,
		observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "birth").ID},
		observations.Input{PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(1901, 0, 0)},
	)
	f.at(birth, 8, 2)
	f.Participation(f.Source, f.Artifact, person, birth, "subject")
	f.Locate(f.Source, f.Artifact, birth, town)

	f.promoteSubject(country)
	f.promoteSubject(region)
	f.promoteSubject(city)
	townID := f.promoteSubject(town)
	personID := f.promoteSubject(person)
	birthID := f.promoteSubject(birth)

	before := personSummary(t, f, personID)
	if !sameSet(before.Birth.Places[0].Parents, []string{"City", "Region", "Country"}) {
		t.Fatalf("birth parents %v", before.Birth.Places[0].Parents)
	}
	if err := f.C.Graph().Verify(); err != nil {
		t.Fatal(err)
	}

	toponymID := f.Prop("toponym").ID
	obsID := observationID(t, f.C, country.ID, toponymID)
	deps := f.Run("update_observation", func(tx *database.Tx) ([]rowchange.Change, error) {
		_, ch, err := observations.Update(tx, userID, observations.Input{
			ID: obsID, SubjectID: country.ID, PropertyID: toponymID,
			Scalar: propertymeta.TextScalar("Renamed"),
		})
		return ch, err
	})
	if !database.ContainsBlobID(deps, townID) || !database.ContainsBlobID(deps, birthID) || !database.ContainsBlobID(deps, personID) {
		t.Fatalf("the country rename reported %d ids, missing the town, the birth, or the person", len(deps))
	}

	after := personSummary(t, f, personID)
	if !sameSet(after.Birth.Places[0].Parents, []string{"City", "Region", "Renamed"}) {
		t.Fatalf("after rename %v", after.Birth.Places[0].Parents)
	}

	towns, err := conclusionsummaries.PlacesByIDs(f.C, [][]byte{townID})
	must(t, err)
	if len(towns) != 1 || !sameSet(towns[0].Parents, []string{"City", "Region", "Renamed"}) {
		t.Fatalf("town chain %v", towns)
	}
	must(t, conclusionsummaries.AttachPlaceRelationships(f.C, &towns[0]))
	if len(towns[0].PartOf) != 1 || towns[0].PartOf[0].Title != "City" {
		t.Fatalf("town part of %+v", towns[0].PartOf)
	}
}

// A person row reads its birth, so merging the birth into another drops the
// stored person row, though only the birth's own handle was committed.
func TestKitMergeDropsReadersDisplays(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	k := graphcachetest.NewCanon(t, c)
	person := k.Person()
	birth, _ := k.Birth(person, nil, nil)
	other := k.Event(vocabkey.EventBirth)
	k.Commit(person, birth, other)

	if _, err := conclusionsummaries.PersonsByIDs(c, [][]byte{person}); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Graph().Display(person); !ok {
		t.Fatal("listing the person stored no display row")
	}
	k.Merge(birth, other)
	k.Commit(birth)
	if _, ok := c.Graph().Display(person); ok {
		t.Fatal("the person's stored display survived a merge of the birth it read")
	}
}

// A birth that becomes a marriage must still name its subject. The person
// row drops the birth, and the search document written in that save drops it too.
func TestBirthBecomingMarriageDropsThePerson(t *testing.T) {
	f := newFixture(t)
	person := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(person, "Ada Birth")
	f.at(person, 0, 0)
	birth := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, birth,
		observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "birth").ID},
		observations.Input{PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(1901, 0, 0)},
	)
	f.at(birth, 2, 0)
	f.Participation(f.Source, f.Artifact, person, birth, "subject")
	personID := f.promoteSubject(person)
	birthID := f.promoteSubject(birth)

	before := personSummary(t, f, personID)
	if before.Birth.Event == nil || !bytes.Equal(before.Birth.Event.ID, birthID) {
		t.Fatal("birth was not on the person row")
	}
	db, err := f.C.DB()
	must(t, err)
	if n := personDocsMentioning(t, db, "1901"); n != 1 {
		t.Fatalf("person search docs mentioning 1901: %d", n)
	}

	eventTypeID := f.Prop("event_type").ID
	marriageID := f.Term("event_type", "marriage").ID
	obsID := observationID(t, f.C, birth.ID, eventTypeID)
	deps := f.Run("update_observation", func(tx *database.Tx) ([]rowchange.Change, error) {
		_, ch, err := observations.Update(tx, userID, observations.Input{
			ID: obsID, SubjectID: birth.ID, PropertyID: eventTypeID,
			ValueTermID: marriageID,
		})
		return ch, err
	})
	if !database.ContainsBlobID(deps, personID) {
		t.Fatal("a marriage that was a birth did not report its subject")
	}
	after := personSummary(t, f, personID)
	if after.Birth.Event != nil {
		t.Fatal("person row still shows the birth")
	}
	if n := personDocsMentioning(t, db, "1901"); n != 0 {
		t.Fatalf("person search docs still mentioning 1901: %d", n)
	}
}

func personDocsMentioning(t *testing.T, db *sql.DB, text string) int {
	t.Helper()
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM catalog_search_docs WHERE kind = 'person' AND secondary LIKE ?`, "%"+text+"%").Scan(&n)
	must(t, err)
	return n
}

func personSummary(t *testing.T, f *fixture, id []byte) conclusionsummaries.PersonSummary {
	t.Helper()
	for _, h := range f.list() {
		if string(h.Entity.ID) == string(id) {
			return h
		}
	}
	t.Fatal("person missing from the list")
	return conclusionsummaries.PersonSummary{}
}

func observationID(t *testing.T, c *database.Catalog, subjectID, propertyID []byte) []byte {
	t.Helper()
	db, err := c.DB()
	must(t, err)
	var id []byte
	err = db.QueryRow(`SELECT id FROM observations WHERE subject_id = ? AND property_id = ? LIMIT 1`, subjectID, propertyID).Scan(&id)
	must(t, err)
	return id
}
