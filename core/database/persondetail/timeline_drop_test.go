package persondetail

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
)

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
	part := func(n byte) conclusionsummaries.EventParticipant {
		return conclusionsummaries.EventParticipant{ID: id(n)}
	}
	rows := []PersonEvent{
		{Event: id(1), Participants: []conclusionsummaries.EventParticipant{part(10)}},
		{Event: id(2), Participants: []conclusionsummaries.EventParticipant{part(10), part(11)}},
		{Event: id(3), Participants: []conclusionsummaries.EventParticipant{part(12)}},
	}
	// Event 1 and participant 11 were not read; participant 12 was not read.
	summaries := conclusionsummaries.NewSummaries(
		[]conclusionsummaries.PersonSummary{person(10)},
		[]conclusionsummaries.EventSummary{event(2), event(3)}, nil)
	got := dropUnreadEvents(rows, summaries)
	if len(got) != 2 || !bytes.Equal(got[0].Event, id(2)) || len(got[0].Participants) != 1 || !bytes.Equal(got[0].Participants[0].ID, id(10)) {
		t.Fatalf("got %+v", got)
	}
	if !bytes.Equal(got[1].Event, id(3)) || len(got[1].Participants) != 0 {
		t.Fatalf("an event with no readable participant: %+v", got[1])
	}
}
