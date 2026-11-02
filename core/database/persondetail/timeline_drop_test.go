package persondetail

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
)

// TestRolesByNeighbor covers D-5: the one collapse rule of both walks.
func TestRolesByNeighbor(t *testing.T) {
	id := func(n byte) []byte { return []byte{n} }
	link := func(neighbor, term byte) graphcache.Link {
		l := graphcache.Link{Neighbor: id(neighbor)}
		if term != 0 {
			l.TermID, l.TermKey = id(term), "t"
		}
		return l
	}
	label := func(termID []byte, key string) (conclusionsummaries.Term, bool) {
		if bytes.Equal(termID, id(9)) {
			return conclusionsummaries.Term{}, false
		}
		return conclusionsummaries.Term{ID: termID, Key: key}, true
	}
	// A subject (1), A informant (2), A subject again (1), B no term, C a term
	// the labeller does not know (9), D not kept.
	links := []graphcache.Link{link('A', 1), link('A', 2), link('A', 1), link('B', 0), link('C', 9), link('D', 1)}
	got := rolesByNeighbor(links, label, func(n []byte) bool { return !bytes.Equal(n, id('D')) })
	if len(got) != 3 {
		t.Fatalf("%d neighbors, want 3: %+v", len(got), got)
	}
	if !bytes.Equal(got[0].id, id('A')) || len(got[0].roles) != 2 ||
		!bytes.Equal(got[0].roles[0].ID, id(1)) || !bytes.Equal(got[0].roles[1].ID, id(2)) {
		t.Fatalf("A: %+v", got[0])
	}
	if !bytes.Equal(got[1].id, id('B')) || len(got[1].roles) != 0 {
		t.Fatalf("B: %+v", got[1])
	}
	if !bytes.Equal(got[2].id, id('C')) || len(got[2].roles) != 0 {
		t.Fatalf("C: %+v", got[2])
	}
}

// TestDropUnreadEvents covers D-7: a summary missing mid-read cannot be built
// from the catalog, so the summaries come from the rows given.
func TestDropUnreadEvents(t *testing.T) {
	id := func(n byte) []byte { return []byte{n} }
	person := func(n byte) conclusionsummaries.PersonSummary {
		return conclusionsummaries.PersonSummary{Entity: canonicalentities.Entity{ID: id(n)}}
	}
	event := func(n byte) conclusionsummaries.EventSummary {
		return conclusionsummaries.EventSummary{Entity: canonicalentities.Entity{ID: id(n)}}
	}
	part := func(n byte) EventParticipant { return EventParticipant{Person: id(n)} }
	rows := []PersonEvent{
		{Event: id(1), Participants: []EventParticipant{part(10)}},
		{Event: id(2), Participants: []EventParticipant{part(10), part(11)}},
		{Event: id(3), Participants: []EventParticipant{part(12)}},
	}
	// Event 1 and participant 11 were not read; participant 12 was not read.
	summaries := conclusionsummaries.NewSummaries(
		[]conclusionsummaries.PersonSummary{person(10)},
		[]conclusionsummaries.EventSummary{event(2), event(3)}, nil)
	got := dropUnreadEvents(rows, summaries)
	if len(got) != 2 || !bytes.Equal(got[0].Event, id(2)) || len(got[0].Participants) != 1 || !bytes.Equal(got[0].Participants[0].Person, id(10)) {
		t.Fatalf("got %+v", got)
	}
	if !bytes.Equal(got[1].Event, id(3)) || len(got[1].Participants) != 0 {
		t.Fatalf("an event with no readable participant: %+v", got[1])
	}
}
