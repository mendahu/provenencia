package conclusionheaders_test

import (
	"bytes"
	"database/sql"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
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

	person := f.bare("person")
	f.named(person, "Ada Town")
	f.at(person, 8, 0)
	birth := f.bare("event")
	f.cite(birth,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(1901)},
	)
	f.at(birth, 8, 2)
	f.participation(person, birth, "subject")
	f.location(birth, town)

	countryID := f.promoteSubject(country)
	f.promoteSubject(region)
	f.promoteSubject(city)
	townID := f.promoteSubject(town)
	personID := f.promoteSubject(person)
	birthID := f.promoteSubject(birth)

	db, err := f.c.DB()
	must(t, err)
	deps, err := conclusionheaders.HeaderDependents(db, [][]byte{countryID})
	must(t, err)
	if !hasID(deps, townID) || !hasID(deps, birthID) || !hasID(deps, personID) {
		t.Fatalf("country dependents missing the town, the birth, or the person (%d ids)", len(deps))
	}

	before := personHeader(t, f, personID)
	if !sameSet(before.Birth.Places[0].Parents, []string{"City", "Region", "Country"}) {
		t.Fatalf("birth parents %v", before.Birth.Places[0].Parents)
	}
	if err := f.c.Graph().Verify(); err != nil {
		t.Fatal(err)
	}

	toponymID := f.prop("toponym").ID
	obsID := observationID(t, f.c, country.ID, toponymID)
	_, err = writes.Call(f.c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, userID, observations.Input{
			ID: obsID, SubjectID: country.ID, PropertyID: toponymID,
			ValueText: "Renamed", HasText: true,
		})
	})
	must(t, err)

	after := personHeader(t, f, personID)
	if !sameSet(after.Birth.Places[0].Parents, []string{"City", "Region", "Renamed"}) {
		t.Fatalf("after rename %v", after.Birth.Places[0].Parents)
	}
}

// A birth that becomes a marriage must still name its subject. The person
// row drops the birth, and the search document written in that save drops it too.
func TestBirthBecomingMarriageDropsThePerson(t *testing.T) {
	f := newFixture(t)
	person := f.bare("person")
	f.named(person, "Ada Birth")
	f.at(person, 0, 0)
	birth := f.bare("event")
	f.cite(birth,
		observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID},
		observations.Input{PropertyID: f.prop("date").ID, Date: pointYear(1901)},
	)
	f.at(birth, 2, 0)
	f.participation(person, birth, "subject")
	personID := f.promoteSubject(person)
	birthID := f.promoteSubject(birth)

	before := personHeader(t, f, personID)
	if before.Birth.Event == nil || !bytes.Equal(before.Birth.Event.ID, birthID) {
		t.Fatal("birth was not on the person row")
	}
	db, err := f.c.DB()
	must(t, err)
	if n := personDocsMentioning(t, db, "1901"); n != 1 {
		t.Fatalf("person search docs mentioning 1901: %d", n)
	}

	eventTypeID := f.prop("event_type").ID
	marriageID := f.term("event_type", "marriage").ID
	obsID := observationID(t, f.c, birth.ID, eventTypeID)
	_, err = writes.Call(f.c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, userID, observations.Input{
			ID: obsID, SubjectID: birth.ID, PropertyID: eventTypeID,
			ValueTermID: marriageID,
		})
	})
	must(t, err)

	deps, err := conclusionheaders.HeaderDependents(db, [][]byte{birthID})
	must(t, err)
	if !hasID(deps, personID) {
		t.Fatal("a marriage that was a birth did not reach its subject")
	}
	after := personHeader(t, f, personID)
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

func personHeader(t *testing.T, f *fixture, id []byte) conclusionheaders.PersonHeader {
	t.Helper()
	for _, h := range f.list() {
		if string(h.Entity.ID) == string(id) {
			return h
		}
	}
	t.Fatal("person missing from the list")
	return conclusionheaders.PersonHeader{}
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
