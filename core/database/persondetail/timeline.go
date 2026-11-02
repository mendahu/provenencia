package persondetail

import (
	"bytes"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// PersonEvent is one event a person took part in directly: the person's roles
// on it, the other canonical participants, and whether it is dated.
type PersonEvent struct {
	Event        []byte
	Roles        []conclusionsummaries.Term
	Participants []EventParticipant
	Dated        bool
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
		have := false
		for _, r := range out[i].roles {
			if bytes.Equal(r.ID, l.TermID) {
				have = true
				break
			}
		}
		if have {
			continue
		}
		if t, ok := label(l.TermID, l.TermKey); ok {
			out[i].roles = append(out[i].roles, t)
		}
	}
	return out
}

// PersonEvents is the Timeline section's read: every canonical event the
// person took part in, in any role, in event date order, and the summary of
// every event and participant it names. An unknown or merged handle, or one
// that is no Person, is conclusiondetails.ErrNotFound.
func PersonEvents(c *database.Catalog, entityID []byte) ([]PersonEvent, conclusionsummaries.Summaries, error) {
	none := conclusionsummaries.Summaries{}
	n, err := conclusiondetails.Handle(c, entityID)
	if err != nil {
		return nil, none, err
	}
	g := c.Graph()
	v, err := g.Vocabulary()
	if err != nil {
		return nil, none, err
	}
	personType, eventType := v.TypeIDsByKey()[vocabkey.Person], v.TypeIDsByKey()[vocabkey.Event]
	if !n.Is(personType) {
		return nil, none, conclusiondetails.ErrNotFound
	}
	label, err := conclusionsummaries.TermLabeler(c)
	if err != nil {
		return nil, none, err
	}
	keepAll := func([]byte) bool { return true }

	links, err := g.Follow(entityID, hops.EventsOfParticipant)
	if err != nil {
		return nil, none, err
	}
	mine := rolesByNeighbor(links, label, keepAll)
	eventIDs := make([][]byte, len(mine))
	for i, m := range mine {
		eventIDs[i] = m.id
	}
	if err := g.Nodes(eventIDs); err != nil {
		return nil, none, err
	}

	// The canonical events, each with the links to its participants.
	type walk struct {
		roles []conclusionsummaries.Term
		links []graphcache.Link
	}
	walks := map[string]walk{}
	var events, others [][]byte
	for _, m := range mine {
		if e, _ := g.Cached(m.id); !canonical(e, eventType) {
			continue
		}
		ls, err := g.Follow(m.id, hops.ParticipantsOfEvent)
		if err != nil {
			return nil, none, err
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
		return nil, none, err
	}
	keep := func(id []byte) bool {
		p, _ := g.Cached(id)
		return !bytes.Equal(id, entityID) && canonical(p, personType)
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
		return nil, none, err
	}
	personRows, err := conclusionsummaries.PersonsByIDs(c, people)
	if err != nil {
		return nil, none, err
	}
	summaries := conclusionsummaries.NewSummaries(personRows, eventRows, nil)

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
	for i := range rows {
		h, _ := summaries.Event(rows[i].Event)
		rows[i].Dated = conclusionsummaries.HasEventDate(h)
	}
	return rows, summaries, nil
}

// dropUnreadEvents drops each event whose summary was not read and each
// participant whose summary was not read (a node dropped mid-read: the next
// read is right, as dropUnread does). An event whose participants are all
// dropped stays. It neither orders nor sets Dated.
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
