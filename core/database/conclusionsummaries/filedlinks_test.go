package conclusionsummaries_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
)

// promotedBirth promotes a person and their birth event, files the
// participation, and files the location at place unless place is the zero Subject.
func (f *fixture) promotedBirth(place subjects.Subject) (person, event subjects.Subject, handles [2][]byte) {
	f.T.Helper()
	person = subjectstest.On(f.T, f.C, userID, f.graph, "person", "")
	f.named(person, "Ann")
	f.at(person, 0, 20)
	event = subjectstest.On(f.T, f.C, userID, f.graph, "event", "")
	f.at(event, 6, 20)
	f.Cite(f.Artifact, event, observations.Input{PropertyID: f.Prop("event_type").ID, ValueTermID: f.Term("event_type", "birth").ID})
	handles = [2][]byte{f.promoteSubject(person), f.promoteSubject(event)}
	f.Participation(f.Source, f.Artifact, person, event, "subject")
	if place.ID != nil {
		f.Locate(f.Source, f.Artifact, event, place)
	}
	return person, event, handles
}

func countIn(ids [][]byte, want ...[]byte) int {
	n := 0
	for _, w := range want {
		if database.ContainsBlobID(ids, w) {
			n++
		}
	}
	return n
}

// A filed location names its event and its place, and reaches no other event
// at the place, however many there are: no row reads a link through the place end.
func TestFiledLocationReportsOnlyItsEvent(t *testing.T) {
	for _, siblings := range []int{3, 12} {
		f := newFixture(t)
		place := f.placeSubject("Leeds", nil, nil, 2, 0)
		placeID := f.promoteSubject(place)
		var siblingEvents [][]byte
		for i := 0; i < siblings; i++ {
			_, _, h := f.promotedBirth(place)
			siblingEvents = append(siblingEvents, h[1])
		}
		_, event, h := f.promotedBirth(subjects.Subject{})
		f.Locate(f.Source, f.Artifact, event, place)
		deps := f.Reported
		if countIn(deps, h[1], placeID) != 2 {
			t.Fatalf("%d siblings: deps lack an end of the link", siblings)
		}
		if n := countIn(deps, siblingEvents...); n != 0 {
			t.Fatalf("%d siblings: %d sibling events reprojected for a link on another event", siblings, n)
		}
	}
}

// A filed part_of names both places and walks from the part: the events at the
// part and at its descendants, and their birth or death people, are reached;
// the events at the whole are not.
func TestFiledPartOfReportsThePart(t *testing.T) {
	f := newFixture(t)
	whole := f.placeSubject("Ontario", nil, nil, 4, 0)
	part := f.placeSubject("Toronto", nil, nil, 2, 0)
	sub := f.placeSubject("Yorkville", nil, nil, 0, 0)
	wholeID, partID := f.promoteSubject(whole), f.promoteSubject(part)
	f.promoteSubject(sub)
	f.placeRel(sub, part, propertyterms.KeyPartOf, nil, nil)
	_, _, hPart := f.promotedBirth(part)
	_, _, hSub := f.promotedBirth(sub)
	_, _, hWhole := f.promotedBirth(whole)
	f.PlaceRel(f.Source, f.Artifact, part, whole, propertyterms.KeyPartOf)
	deps := f.Reported
	if countIn(deps, wholeID, partID) != 2 {
		t.Fatalf("deps lack an end of the link")
	}
	// The events at the part and at its descendant, and the people they are the birth of.
	want := [][]byte{hPart[0], hPart[1], hSub[0], hSub[1]}
	if countIn(deps, want...) != len(want) {
		t.Fatalf("deps lack the part's events or their people: %d of %d", countIn(deps, want...), len(want))
	}
	if countIn(deps, hWhole[0], hWhole[1]) != 0 {
		t.Fatalf("deps reach the events at the whole")
	}
}

// staleEventDocs marks every event search document, so a later write shows
// whether it rewrote one: a rewrite replaces the mark.
func (f *fixture) staleEventDocs() {
	f.T.Helper()
	db, err := f.C.DB()
	must(f.T, err)
	_, err = db.Exec(`UPDATE catalog_search_docs SET display_subtitle = 'stale' WHERE kind = 'event'`)
	must(f.T, err)
}

func (f *fixture) eventDocsStale() bool {
	f.T.Helper()
	db, err := f.C.DB()
	must(f.T, err)
	var n int
	must(f.T, db.QueryRow(`SELECT COUNT(*) FROM catalog_search_docs WHERE kind = 'event' AND display_subtitle = 'stale'`).Scan(&n))
	return n > 0
}

// The first part_of filed from a place moves the rows that read the place's
// parents, so the search document of an event at that place is rewritten. A
// second record of the same link joins the association and moves nothing, so
// that document is left alone.
func TestFiledPartOfReprojectsEventsOnlyWhenARowMoved(t *testing.T) {
	f := newFixture(t)
	whole := f.placeSubject("Ontario", nil, nil, 4, 0)
	part := f.placeSubject("Toronto", nil, nil, 2, 0)
	wholeID, partID := f.promoteSubject(whole), f.promoteSubject(part)
	f.promotedBirth(part)

	f.staleEventDocs()
	f.placeRel(part, whole, propertyterms.KeyPartOf, nil, nil)
	if f.eventDocsStale() {
		t.Fatal("the first part_of from a place left its event's search document unwritten")
	}

	join := func(name string, entityID []byte) subjects.Subject {
		s := f.placeSubject(name, nil, nil, 0, 0)
		promotetest.Save(t, f.C, userID, promote.Input{SubjectID: s.ID, EntityID: entityID})
		return s
	}
	part2, whole2 := join("Toronto", partID), join("Ontario", wholeID)
	f.staleEventDocs()
	f.placeRel(part2, whole2, propertyterms.KeyPartOf, nil, nil)
	if !f.eventDocsStale() {
		t.Fatal("a second record of the same part_of rewrote its event's search document")
	}
}
