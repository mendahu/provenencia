package conclusionsummaries_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
)

// filedLinkDeps recomputes the association handle of the last bridge of type
// kind, as the write that filed it does, and returns what the walk reported.
func (f *fixture) filedLinkDeps(kind string) [][]byte {
	f.t.Helper()
	db, err := f.c.DB()
	must(f.t, err)
	var assoc []byte
	must(f.t, db.QueryRow(`SELECT ic.entity_id FROM identity_claims ic
		JOIN subjects s ON s.id = ic.subject_id
		JOIN subject_types st ON st.id = s.subject_type_id AND st.key = ?
		WHERE ic.status = 'accepted' ORDER BY s.id DESC LIMIT 1`, kind).Scan(&assoc))
	deps, err := autoreconciler.RecomputeReported(db, [][]byte{assoc})
	must(f.t, err)
	return deps
}

// promotedBirth promotes a person and their birth event, files the
// participation, and files the location at place unless place is the zero Subject.
func (f *fixture) promotedBirth(place subjects.Subject) (person, event subjects.Subject, handles [2][]byte) {
	f.t.Helper()
	person = subjectstest.On(f.t, f.c, userID, f.graph, "person", "")
	f.named(person, "Ann")
	f.at(person, 0, 20)
	event = subjectstest.On(f.t, f.c, userID, f.graph, "event", "")
	f.at(event, 6, 20)
	f.cite(event, observations.Input{PropertyID: f.prop("event_type").ID, ValueTermID: f.term("event_type", "birth").ID})
	handles = [2][]byte{f.promoteSubject(person), f.promoteSubject(event)}
	f.participation(person, event, "subject")
	if place.ID != nil {
		f.location(event, place)
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
func TestFiledLocationWalksOnlyFromItsEvent(t *testing.T) {
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
		f.location(event, place)
		deps := f.filedLinkDeps("location")
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
func TestFiledPartOfWalksFromThePart(t *testing.T) {
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
	f.placeRel(part, whole, propertyterms.KeyPartOf, nil, nil)
	deps := f.filedLinkDeps("place_relationship")
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
