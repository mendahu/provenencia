package persondetail

import (
	"bytes"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// PersonEvent is one event a person took part in directly: the person's roles
// on it, the other canonical participants (most important role first), and its
// span on the line of years, nil when it is undated, and the person's age on
// it, nil when no age can be told.
type PersonEvent struct {
	Event        []byte
	Roles        []conclusionsummaries.Term
	Participants []conclusionsummaries.EventParticipant
	Span         *conclusionsummaries.TimelineSpan
	Age          *conclusionsummaries.AgeRange
}

// Timeline is the Timeline section's read: the person's events in the
// Timeline's order (date order, then lifelineTypes.order's moves) and where the
// birth and death markers go.
type Timeline struct {
	Events       []PersonEvent
	Birth, Death LifeMarker
}

// PersonEvents is the Timeline section's read: every canonical event the
// person took part in, in any role, in the Timeline's order (event date order,
// then lifelineTypes.order's moves), each with its span on
// the line of years, the birth and death markers (the header's birth and death
// when they are among the events, else placed after or before a dated event), and the
// summary of every event and participant it names. An unknown or merged handle, or one
// that is no Person, is conclusiondetails.ErrNotFound.
func PersonEvents(c *database.Catalog, entityID []byte) (Timeline, conclusionsummaries.Summaries, error) {
	none := conclusionsummaries.Summaries{}
	if _, err := personNode(c, entityID); err != nil {
		return Timeline{}, none, err
	}
	g := c.Graph()
	v, err := g.Vocabulary()
	if err != nil {
		return Timeline{}, none, err
	}
	types := v.TypeIDsByKey()
	personType, eventType := types[vocabkey.Person], types[vocabkey.Event]
	label, err := conclusionsummaries.TermLabeler(c)
	if err != nil {
		return Timeline{}, none, err
	}
	links, err := g.Follow(entityID, hops.EventsOfParticipant)
	if err != nil {
		return Timeline{}, none, err
	}
	var eventIDs [][]byte
	seenEvent := map[string]bool{}
	for _, l := range links {
		if !seenEvent[string(l.Neighbor)] {
			seenEvent[string(l.Neighbor)] = true
			eventIDs = append(eventIDs, l.Neighbor)
		}
	}
	if err := g.Nodes(eventIDs); err != nil {
		return Timeline{}, none, err
	}
	var events [][]byte
	for _, id := range eventIDs {
		if e, _ := g.Cached(id); e.Canonical(eventType) {
			events = append(events, id)
		}
	}
	eventRows, err := conclusionsummaries.EventsByIDs(c, events)
	if err != nil {
		return Timeline{}, none, err
	}
	// The other participants come from the event rows, read in one load.
	var others [][]byte
	for _, h := range eventRows {
		for _, p := range h.Participants {
			if !bytes.Equal(p.ID, entityID) {
				others = append(others, p.ID)
			}
		}
	}
	if err := g.Nodes(others); err != nil {
		return Timeline{}, none, err
	}
	canonicalPerson := func(id []byte) bool {
		p, _ := g.Cached(id)
		return p.Canonical(personType)
	}
	// Roles are labelled on copies: the stored row's slices stay as composed.
	labelled := func(roles []conclusionsummaries.Term) []conclusionsummaries.Term {
		var out []conclusionsummaries.Term
		for _, r := range roles {
			if t, ok := label(r.ID, r.Key); ok {
				out = append(out, t)
			}
		}
		return out
	}
	rows := make([]PersonEvent, 0, len(eventRows))
	var people [][]byte
	for _, h := range eventRows {
		row := PersonEvent{Event: h.Entity.ID}
		for _, p := range h.Participants {
			if bytes.Equal(p.ID, entityID) {
				row.Roles = labelled(p.Roles)
				continue
			}
			if !canonicalPerson(p.ID) {
				continue
			}
			people = append(people, p.ID)
			row.Participants = append(row.Participants, conclusionsummaries.EventParticipant{ID: p.ID, Roles: labelled(p.Roles)})
		}
		rows = append(rows, row)
	}
	personRows, err := conclusionsummaries.PersonsByIDs(c, append(people, entityID))
	if err != nil {
		return Timeline{}, none, err
	}
	summaries := conclusionsummaries.NewSummaries(personRows, eventRows, nil)
	index := make(map[string]int, len(personRows))
	for i, p := range personRows {
		index[string(p.Entity.ID)] = i
	}

	rows = dropUnreadEvents(rows, summaries)
	me, haveMe := summaries.Person(entityID)
	var birthDate *datevalue.Value
	if haveMe {
		birthDate = me.Birth.Date
	}
	placedBirth, hasBirth := conclusionsummaries.PlaceBirth(birthDate)
	days := map[string]valuecompare.Window{}
	for i := range rows {
		h, _ := summaries.Event(rows[i].Event)
		placing := conclusionsummaries.EventSpanAndAge(h, placedBirth, hasBirth)
		if placing.Dated {
			rows[i].Span = &placing.Span
			days[string(rows[i].Event)] = placing.Days
		}
		if placing.Aged {
			rows[i].Age = &placing.Age
		}
	}
	for i := range rows {
		sortParticipants(rows[i].Participants, index)
	}

	var birth, death []byte
	if haveMe {
		birth, death = recorded(rows, me.Birth.Event), recorded(rows, me.Death.Event)
	}
	typeOf := func(event []byte) vocabkey.EventType {
		if h, ok := summaries.Event(event); ok {
			return vocabkey.EventType(h.Type.Term.ProductKey())
		}
		return ""
	}
	daysOf := func(id []byte) (valuecompare.Window, bool) { w, ok := days[string(id)]; return w, ok }
	return lifeline.timeline(rows, typeOf, daysOf, birth, death), summaries, nil
}

// recorded is the id of pick when it is one of the timeline's rows, else nil:
// the header can name an event the timeline drops.
func recorded(rows []PersonEvent, pick *canonicalentities.Entity) []byte {
	if pick == nil {
		return nil
	}
	for _, r := range rows {
		if bytes.Equal(r.Event, pick.ID) {
			return r.Event
		}
	}
	return nil
}

// dropUnreadEvents drops each event whose summary was not read and each
// participant whose summary was not read (a node dropped mid-read: the next
// read is right, as dropUnread does). An event whose participants are all
// dropped stays. It neither orders nor sets Span or Age.
func dropUnreadEvents(events []PersonEvent, summaries conclusionsummaries.Summaries) []PersonEvent {
	var out []PersonEvent
	for _, e := range events {
		if _, ok := summaries.Event(e.Event); !ok {
			continue
		}
		var kept []conclusionsummaries.EventParticipant
		for _, p := range e.Participants {
			if _, ok := summaries.Person(p.ID); ok {
				kept = append(kept, p)
			}
		}
		e.Participants = kept
		out = append(out, e)
	}
	return out
}
