package persondetail

import (
	"bytes"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// PersonEvent is one event a person took part in directly: the person's roles
// on it, the other canonical participants (most important role first), and its
// span on the line of years, nil when it is undated, and the person's age on
// it, nil when no age can be told.
type PersonEvent struct {
	Event        []byte
	Roles        []conclusionsummaries.Term
	Participants []EventParticipant
	Span         *conclusionsummaries.TimelineSpan
	Age          *conclusionsummaries.AgeRange
}

// Timeline is the Timeline section's read: the person's events in date order
// and where the birth and death markers go.
type Timeline struct {
	Events       []PersonEvent
	Birth, Death LifeMarker
}

// EventParticipant is one other canonical person on the event and their roles.
type EventParticipant struct {
	Person []byte
	Roles  []conclusionsummaries.Term
}

// neighborRoles is one neighbor of a walk and the distinct roles its links
// carry.
type neighborRoles struct {
	id    []byte
	roles []conclusionsummaries.Term
	seen  map[string]struct{}
}

// rolesByNeighbor is the one collapse rule for both timeline walks: links in
// input order; a link whose neighbor fails keep is skipped; neighbors are
// de-duplicated in first-seen order; each link adds its role term to its
// neighbor's roles unless that term is already there. A link with no term, or
// whose term label does not know, adds no role.
func rolesByNeighbor(links []graphcache.Link, label func(id []byte, key string) (conclusionsummaries.Term, bool), keep func(id []byte) bool) []neighborRoles {
	var out []neighborRoles
	index := map[string]int{}
	for _, l := range links {
		if !keep(l.Neighbor) {
			continue
		}
		i, ok := index[string(l.Neighbor)]
		if !ok {
			i = len(out)
			index[string(l.Neighbor)] = i
			out = append(out, neighborRoles{id: l.Neighbor})
		}
		if len(l.TermID) == 0 {
			continue
		}
		if _, have := out[i].seen[string(l.TermID)]; have {
			continue
		}
		if t, ok := label(l.TermID, l.TermKey); ok {
			if out[i].seen == nil {
				out[i].seen = map[string]struct{}{}
			}
			out[i].seen[string(l.TermID)] = struct{}{}
			out[i].roles = append(out[i].roles, t)
		}
	}
	return out
}

// PersonEvents is the Timeline section's read: every canonical event the
// person took part in, in any role, in event date order, each with its span on
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
	keepAll := func([]byte) bool { return true }

	links, err := g.Follow(entityID, hops.EventsOfParticipant)
	if err != nil {
		return Timeline{}, none, err
	}
	mine := rolesByNeighbor(links, label, keepAll)
	eventIDs := make([][]byte, len(mine))
	for i, m := range mine {
		eventIDs[i] = m.id
	}
	if err := g.Nodes(eventIDs); err != nil {
		return Timeline{}, none, err
	}

	// The canonical events, each with the links to its participants.
	type walk struct {
		roles []conclusionsummaries.Term
		links []graphcache.Link
	}
	walks := map[string]walk{}
	var events, others [][]byte
	for _, m := range mine {
		if e, _ := g.Cached(m.id); !e.Canonical(eventType) {
			continue
		}
		ls, err := g.Follow(m.id, hops.ParticipantsOfEvent)
		if err != nil {
			return Timeline{}, none, err
		}
		walks[string(m.id)] = walk{roles: m.roles, links: ls}
		events = append(events, m.id)
		for _, l := range ls {
			if !bytes.Equal(l.Neighbor, entityID) {
				others = append(others, l.Neighbor)
			}
		}
	}
	if err := g.Nodes(others); err != nil {
		return Timeline{}, none, err
	}
	keep := func(id []byte) bool {
		p, _ := g.Cached(id)
		return !bytes.Equal(id, entityID) && p.Canonical(personType)
	}
	participants := map[string][]neighborRoles{}
	var people [][]byte
	for _, id := range events {
		ps := rolesByNeighbor(walks[string(id)].links, label, keep)
		participants[string(id)] = ps
		for _, p := range ps {
			people = append(people, p.id)
		}
	}

	eventRows, err := conclusionsummaries.EventsByIDs(c, events)
	if err != nil {
		return Timeline{}, none, err
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

	rows := make([]PersonEvent, 0, len(eventRows))
	for _, h := range eventRows {
		id := h.Entity.ID
		row := PersonEvent{Event: id, Roles: walks[string(id)].roles}
		for _, p := range participants[string(id)] {
			row.Participants = append(row.Participants, EventParticipant{Person: p.id, Roles: p.roles})
		}
		rows = append(rows, row)
	}
	rows = dropUnreadEvents(rows, summaries)
	me, haveMe := summaries.Person(entityID)
	var birthDate *datevalue.Value
	if haveMe {
		birthDate = me.Birth.Date
	}
	placedBirth, hasBirth := conclusionsummaries.PlaceBirth(birthDate)
	for i := range rows {
		h, _ := summaries.Event(rows[i].Event)
		span, dated, age, aged := conclusionsummaries.EventSpanAndAge(h, placedBirth, hasBirth)
		if dated {
			rows[i].Span = &span
		}
		if aged {
			rows[i].Age = &age
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
	t := Timeline{Events: rows}
	t.Birth, t.Death = lifeline.markers(rows, typeOf, birth, death)
	return t, summaries, nil
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
		var kept []EventParticipant
		for _, p := range e.Participants {
			if _, ok := summaries.Person(p.Person); ok {
				kept = append(kept, p)
			}
		}
		e.Participants = kept
		out = append(out, e)
	}
	return out
}
