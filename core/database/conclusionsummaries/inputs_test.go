package conclusionsummaries_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries/conclusionsummariestest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// inputsWorld is the scenario the input tests share. It holds the handle ids
// the assertions name.
type inputsWorld struct {
	james, mary, witness, loner []byte
	birth, marriage, death      []byte
	census, elsewhere           []byte
	country, region, city, town []byte
	oldTown, newShire           []byte
	york, toronto               []byte
}

func (f *fixture) eventOf(kind string, year int) subjects.Subject {
	f.T.Helper()
	e := subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.Cite(f.Artifact, e,
		observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", kind).ID},
		observations.Input{PropertyID: f.Prop("date").ID, Date: datevaluetest.YMD(year, 0, 0)},
	)
	f.at(e, int64(year%7), int64(year%11))
	return e
}

func (f *fixture) personNamed(name string) subjects.Subject {
	f.T.Helper()
	p := subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(p, name)
	f.at(p, 0, int64(len(name)))
	return p
}

// buildInputsWorld files: James, born in Town (part_of City, Region, Country)
// and dead at Old Town, whose part_of link ended before his death; his
// marriage at a place with no parents; a witness at his birth; a census at two
// places, York and its successor Toronto; and a person with no event.
func buildInputsWorld(f *fixture) inputsWorld {
	f.T.Helper()
	country := f.placeSubject("Country", nil, nil, 0, 0)
	region := f.placeSubject("Region", nil, nil, 2, 0)
	city := f.placeSubject("City", nil, nil, 4, 0)
	town := f.placeSubject("Town", nil, nil, 6, 0)
	f.placeRel(region, country, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(city, region, propertyterms.KeyPartOf, nil, nil)
	f.placeRel(town, city, propertyterms.KeyPartOf, nil, nil)
	oldTown := f.placeSubject("Old Town", nil, nil, 8, 0)
	newShire := f.placeSubject("New Shire", nil, nil, 10, 0)
	from, to := 1700, 1750
	f.placeRel(oldTown, newShire, propertyterms.KeyPartOf, &from, &to)
	elsewhere := f.placeSubject("Elsewhere", nil, nil, 12, 0)
	york := f.placeSubject("York", nil, nil, 14, 0)
	toronto := f.placeSubject("Toronto", nil, nil, 16, 0)
	f.placeRel(york, toronto, propertyterms.KeySucceededBy, nil, nil)

	james := f.personNamed("James Robins")
	mary := f.personNamed("Mary Smith")
	witness := f.personNamed("Witness Webb")
	loner := f.personNamed("Loner Lee")
	birth := f.eventOf("birth", 1901)
	death := f.eventOf("death", 1980)
	marriage := f.eventOf("marriage", 1925)
	census := f.eventOf("census", 1911)
	f.Participation(f.Source, f.Artifact, james, birth, "subject")
	f.Participation(f.Source, f.Artifact, witness, birth, "witness")
	f.Participation(f.Source, f.Artifact, james, death, "subject")
	f.Participation(f.Source, f.Artifact, james, marriage, "subject")
	f.Participation(f.Source, f.Artifact, mary, marriage, "subject")
	f.Participation(f.Source, f.Artifact, mary, census, "subject")
	f.Locate(f.Source, f.Artifact, birth, town)
	f.Locate(f.Source, f.Artifact, death, oldTown)
	f.Locate(f.Source, f.Artifact, marriage, elsewhere)
	f.Locate(f.Source, f.Artifact, census, york)
	f.Locate(f.Source, f.Artifact, census, toronto)

	return inputsWorld{
		country: f.promoteSubject(country), region: f.promoteSubject(region),
		city: f.promoteSubject(city), town: f.promoteSubject(town),
		oldTown: f.promoteSubject(oldTown), newShire: f.promoteSubject(newShire),
		elsewhere: f.promoteSubject(elsewhere), york: f.promoteSubject(york), toronto: f.promoteSubject(toronto),
		james: f.promoteSubject(james), mary: f.promoteSubject(mary),
		witness: f.promoteSubject(witness), loner: f.promoteSubject(loner),
		birth: f.promoteSubject(birth), death: f.promoteSubject(death),
		marriage: f.promoteSubject(marriage), census: f.promoteSubject(census),
	}
}

func handleIDs(t *testing.T, f *fixture) map[string][][]byte {
	t.Helper()
	db, err := f.C.DB()
	must(t, err)
	ids := map[string][][]byte{}
	for _, kind := range conclusionsummariestest.Kinds {
		ids[kind] = conclusionsummariestest.HandleIDs(t, db, kind)
	}
	return ids
}

func inputsOf(t *testing.T, f *fixture, kind string, id []byte) []graphcache.InputKey {
	t.Helper()
	db, err := f.C.DB()
	must(t, err)
	var w conclusionsummaries.Want
	w.Add(kind, id)
	s, err := conclusionsummaries.RowsForSearch(graphcache.Over(db), w)
	must(t, err)
	return s.Inputs(id)
}

func holds(keys []graphcache.InputKey, id []byte, bridge, endpoint string) bool {
	for _, k := range keys {
		if bytes.Equal(k.ID, id) && k.Bridge == bridge && k.Endpoint == endpoint {
			return true
		}
	}
	return false
}

func holdsAny(keys []graphcache.InputKey, id []byte) bool {
	for _, k := range keys {
		if bytes.Equal(k.ID, id) {
			return true
		}
	}
	return false
}

func TestInputsMatchIsolatedComposition(t *testing.T) {
	f := newFixture(t)
	w := buildInputsWorld(f)
	db, err := f.C.DB()
	must(t, err)
	ids := handleIDs(t, f)
	stats := conclusionsummariestest.CheckMatchIsolated(t, db, ids)
	for kind, s := range stats {
		t.Logf("%s: %d rows, %d keys, median %d, max %d", kind, s.Rows, s.Keys, s.Median(), s.Max)
		if s.Rows == 0 {
			t.Errorf("the world has no %s", kind)
		}
	}

	// AC-1.2: a person holds its birth and death events' places and no other
	// event's, and not its own values.
	got := inputsOf(t, f, vocabkey.Person, w.james)
	if holds(got, w.james, "", "") {
		t.Error("James holds his own values")
	}
	for _, c := range []struct {
		name         string
		id           []byte
		bridge, end  string
		wantPresence bool
	}{
		{"birth values", w.birth, "", "", true},
		{"birth location end", w.birth, "location", "event", true},
		{"town values", w.town, "", "", true},
		{"city values", w.city, "", "", true},
		{"region values", w.region, "", "", true},
		{"country values", w.country, "", "", true},
		{"town part_of end", w.town, "place_relationship", "from", true},
		{"death location end", w.death, "location", "event", true},
		{"old town values", w.oldTown, "", "", true},
		{"new shire values", w.newShire, "", "", true},
		{"marriage values", w.marriage, "", "", true},
		{"marriage location end", w.marriage, "location", "event", false},
		{"marriage place values", w.elsewhere, "", "", false},
	} {
		if holds(got, c.id, c.bridge, c.end) != c.wantPresence {
			t.Errorf("James %s: held %v, want %v", c.name, !c.wantPresence, c.wantPresence)
		}
	}

	// A part_of link whose dates exclude it at the event's date is still read.
	death := inputsOf(t, f, vocabkey.Event, w.death)
	if !holdsAny(death, w.newShire) {
		t.Error("the death's inputs miss the part_of link that ended before it")
	}
	// An event at two places reads both and their succession is not followed.
	census := inputsOf(t, f, vocabkey.Event, w.census)
	if !holds(census, w.york, "", "") || !holds(census, w.toronto, "", "") {
		t.Error("the census's inputs miss one of its two places")
	}
	if !holds(census, w.york, "place_relationship", "from") {
		t.Error("the census's inputs miss York's part_of end")
	}
}

func TestInputsCoverDependents(t *testing.T) {
	f := newFixture(t)
	buildInputsWorld(f)
	db, err := f.C.DB()
	must(t, err)
	ids := handleIDs(t, f)
	var probe [][]byte
	for _, kind := range conclusionsummariestest.Kinds {
		probe = append(probe, ids[kind]...)
	}
	extras, excused := conclusionsummariestest.CheckCoversDependents(t, db, ids, probe)
	t.Logf("rows the index reaches beyond Dependents, by kind: %v; persons excused (non-canonical birth or death): %d", extras, excused)
}
