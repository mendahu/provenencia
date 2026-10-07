package conclusionheaders

import (
	"bytes"
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/datevalues"
)

// walkBatch bounds an IN list. A list is still a fixed number of queries.
const walkBatch = 500

// The hops headers read. A subject-role participation joins a person to an
// event (role stays off the association's identity; the walk reads the
// kept rank-1 role). A location joins an event to a place.
var (
	eventsOfPerson = canonicalgraph.MustHop("participation", "person", "event",
		&canonicalgraph.TermFilter{PropertyKey: connectrules.DisambiguationRole, TermKey: "subject"})
	personsOfEvent = eventsOfPerson.Reverse()
	placesOfEvent  = canonicalgraph.MustHop("location", "event", "place", nil)
	eventsAtPlace  = placesOfEvent.Reverse()
)

// Birth and death events among the given handles, with the date a life row
// shows (the date, else the start date) and its sort key.
const sqlLifeEvents = `SELECT ev.id, ev.subject_type_id, ev.ref, COALESCE(ev.argument, ''), COALESCE(ev.label, ''),
	tt.key,
	d.value_date,
	(SELECT COUNT(*) FROM auto_reconciler_values c
		WHERE c.entity_id = ev.id AND c.property_id = dp.id AND c.reason = 'kept'),
	sd.value_date,
	(SELECT COUNT(*) FROM auto_reconciler_values c
		WHERE c.entity_id = ev.id AND c.property_id = sp.id AND c.reason = 'kept'),
	COALESCE(d.sort_key, sd.sort_key)
FROM canonical_entities ev
JOIN auto_reconciler_values et ON et.entity_id = ev.id AND et.rank = 1 AND et.reason = 'kept'
JOIN properties etp ON etp.id = et.property_id AND etp.key = 'event_type' AND etp.origin = 'provenencia'
JOIN property_terms tt ON tt.id = et.value_term_id AND tt.key IN ('birth', 'death')
LEFT JOIN properties dp ON dp.key = 'date' AND dp.origin = 'provenencia'
LEFT JOIN auto_reconciler_values d
	ON d.entity_id = ev.id AND d.property_id = dp.id AND d.rank = 1 AND d.reason = 'kept'
LEFT JOIN properties sp ON sp.key = 'start_date' AND sp.origin = 'provenencia'
LEFT JOIN auto_reconciler_values sd
	ON sd.entity_id = ev.id AND sd.property_id = sp.id AND sd.rank = 1 AND sd.reason = 'kept'
WHERE ev.merged_into_id IS NULL AND ev.id IN (`

func attachLives(q Querier, headers []PersonHeader) error {
	ids := make([][]byte, len(headers))
	for i := range headers {
		ids[i] = headers[i].Entity.ID
	}
	lives, err := loadLives(q, ids)
	if err != nil {
		return err
	}
	for i := range headers {
		if life, ok := lives[string(headers[i].Entity.ID)]; ok {
			headers[i].Birth = life["birth"]
			headers[i].Death = life["death"]
		}
	}
	return nil
}

func attachEventGraph(q Querier, headers []EventHeader) error {
	ids := make([][]byte, len(headers))
	for i := range headers {
		ids[i] = headers[i].Entity.ID
	}
	subjects, err := loadEventSubjects(q, ids)
	if err != nil {
		return err
	}
	places, err := loadPlacesOf(q, ids)
	if err != nil {
		return err
	}
	for i := range headers {
		headers[i].Subjects = subjects[string(headers[i].Entity.ID)]
		headers[i].Places = places[string(headers[i].Entity.ID)]
	}
	return nil
}

type lifeEvent struct {
	entity canonicalentities.Entity
	kind   string
	date   *datevalues.Value
	count  int
	sort   sql.NullString
}

// loadLives walks each person's subject-role events, keeps the births and
// deaths, and reads their places. When several births (or deaths) survive,
// the header leads with the earliest dated one (undated last, then by ref)
// and counts them all: more than one is a disagreement the page shows as
// mixed.
func loadLives(q Querier, personIDs [][]byte) (map[string]map[string]LifeFacts, error) {
	edges, err := canonicalgraph.Walk(q, eventsOfPerson, personIDs)
	if err != nil {
		return nil, err
	}
	events, err := loadLifeEvents(q, canonicalgraph.Targets(edges))
	if err != nil {
		return nil, err
	}
	lifeIDs := make([][]byte, 0, len(events))
	for _, e := range events {
		lifeIDs = append(lifeIDs, e.entity.ID)
	}
	places, err := loadPlacesOf(q, lifeIDs)
	if err != nil {
		return nil, err
	}

	// person → kind → lead event and count
	type pick struct {
		lead  *lifeEvent
		count int
	}
	byPerson := map[string]map[string]*pick{}
	for _, edge := range edges {
		e, ok := events[string(edge.To)]
		if !ok {
			continue
		}
		kinds := byPerson[string(edge.From)]
		if kinds == nil {
			kinds = map[string]*pick{}
			byPerson[string(edge.From)] = kinds
		}
		p := kinds[e.kind]
		if p == nil {
			p = &pick{}
			kinds[e.kind] = p
		}
		p.count++
		if p.lead == nil || leadsLife(e, p.lead) {
			p.lead = e
		}
	}
	out := map[string]map[string]LifeFacts{}
	for person, kinds := range byPerson {
		out[person] = map[string]LifeFacts{}
		for kind, p := range kinds {
			event := p.lead.entity
			out[person][kind] = LifeFacts{
				Event: &event, EventCount: p.count,
				Date: p.lead.date, DateCount: p.lead.count,
				Places: places[string(event.ID)],
			}
		}
	}
	return out, nil
}

// leadsLife orders competing births or deaths: dated before undated, then
// by the date's sort key, then by ref so the choice is stable.
func leadsLife(a, b *lifeEvent) bool {
	if a.sort.Valid != b.sort.Valid {
		return a.sort.Valid
	}
	if a.sort.String != b.sort.String {
		return a.sort.String < b.sort.String
	}
	return a.entity.Ref < b.entity.Ref
}

func loadLifeEvents(q Querier, eventIDs [][]byte) (map[string]*lifeEvent, error) {
	rows, err := queryBatches(q, sqlLifeEvents, eventIDs, func(sc scanner) (*lifeEvent, error) {
		var (
			e                   lifeEvent
			dateBlob, startBlob []byte
			dateCount, startCnt int
		)
		ent := &e.entity
		if err := sc.Scan(&ent.ID, &ent.SubjectTypeID, &ent.Ref, &ent.Argument, &ent.Label,
			&e.kind, &dateBlob, &dateCount, &startBlob, &startCnt, &e.sort); err != nil {
			return nil, err
		}
		e.entity = ownEntity(e.entity)
		if len(dateBlob) == 0 {
			dateBlob, dateCount = startBlob, startCnt
		}
		date, err := unmarshalDate(dateBlob)
		if err != nil {
			return nil, err
		}
		e.date, e.count = date, dateCount
		return &e, nil
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string]*lifeEvent, len(rows))
	for _, e := range rows {
		out[string(e.entity.ID)] = e
	}
	return out, nil
}

// loadEventSubjects walks each event's subject-role persons, participation
// ref then person ref, each person once.
func loadEventSubjects(q Querier, eventIDs [][]byte) (map[string][]EventSubject, error) {
	edges, err := canonicalgraph.Walk(q, personsOfEvent, eventIDs)
	if err != nil {
		return nil, err
	}
	persons, err := personRowsByIDs(q, canonicalgraph.Targets(edges))
	if err != nil {
		return nil, err
	}
	byID := make(map[string]PersonHeader, len(persons))
	for _, p := range persons {
		byID[string(p.Entity.ID)] = p
	}
	out := map[string][]EventSubject{}
	seen := map[string]bool{}
	for _, edge := range edges {
		p, ok := byID[string(edge.To)]
		key := string(edge.From) + "\x00" + string(edge.To)
		if !ok || seen[key] {
			continue
		}
		seen[key] = true
		out[string(edge.From)] = append(out[string(edge.From)], EventSubject{
			Entity: p.Entity, Name: p.Name, NameValueCount: p.NameValueCount,
		})
	}
	return out, nil
}

// loadPlacesOf walks each event's locations to their Places, by place ref.
// A Place's names are its kept toponyms.
func loadPlacesOf(q Querier, eventIDs [][]byte) (map[string][]HeaderPlace, error) {
	edges, err := canonicalgraph.Walk(q, placesOfEvent, eventIDs)
	if err != nil {
		return nil, err
	}
	headers, err := PlacesByIDs(q, canonicalgraph.Targets(edges))
	if err != nil {
		return nil, err
	}
	byID := make(map[string]PlaceHeader, len(headers))
	for _, h := range headers {
		byID[string(h.Entity.ID)] = h
	}
	out := map[string][]HeaderPlace{}
	seen := map[string]bool{}
	for _, edge := range edges {
		h, ok := byID[string(edge.To)]
		key := string(edge.From) + "\x00" + string(edge.To)
		if !ok || seen[key] {
			continue
		}
		seen[key] = true
		out[string(edge.From)] = append(out[string(edge.From)], HeaderPlace{
			Entity: h.Entity, Names: h.Names, Count: len(h.Names),
		})
	}
	for event, places := range out {
		sort.SliceStable(places, func(i, j int) bool { return places[i].Entity.Ref < places[j].Entity.Ref })
		out[event] = places
	}
	return out, nil
}

// HeaderDependents is the reverse of the header walks, for a later search
// reprojection (S9-34): the handles whose header embeds one of ids. A Place
// reaches its events and those events' subject persons. An Event reaches its
// subject persons (their birth or death). A Person reaches the events of
// their subject-role participations. Association handles are not returned.
// This does not reproject.
func HeaderDependents(q Querier, ids [][]byte) ([][]byte, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	placeEdges, err := canonicalgraph.Walk(q, eventsAtPlace, ids)
	if err != nil {
		return nil, err
	}
	events := canonicalgraph.Targets(placeEdges)
	personEdges, err := canonicalgraph.Walk(q, personsOfEvent, append(append([][]byte(nil), events...), ids...))
	if err != nil {
		return nil, err
	}
	eventEdges, err := canonicalgraph.Walk(q, eventsOfPerson, ids)
	if err != nil {
		return nil, err
	}
	var out [][]byte
	seen := map[string]bool{}
	for _, id := range ids {
		seen[string(id)] = true
	}
	for _, group := range [][][]byte{events, canonicalgraph.Targets(personEdges), canonicalgraph.Targets(eventEdges)} {
		for _, id := range group {
			if !seen[string(id)] {
				seen[string(id)] = true
				out = append(out, id)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return bytes.Compare(out[i], out[j]) < 0 })
	return out, nil
}

// ownEntity copies an entity's byte fields off the driver's scan buffers.
func ownEntity(e canonicalentities.Entity) canonicalentities.Entity {
	if len(e.ID) == 0 {
		return canonicalentities.Entity{}
	}
	e.ID = append([]byte(nil), e.ID...)
	e.SubjectTypeID = append([]byte(nil), e.SubjectTypeID...)
	return e
}

type scanner interface {
	Scan(dest ...any) error
}

func queryBatches[T any](q Querier, query string, ids [][]byte, scan func(scanner) (T, error)) ([]T, error) {
	ids = database.UniqueBlobIDs(ids)
	var out []T
	for start := 0; start < len(ids); start += walkBatch {
		end := min(start+walkBatch, len(ids))
		batch := ids[start:end]
		rows, err := q.Query(query+database.SQLInPlaceholders(len(batch))+`)`, database.BlobArgs(batch)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			item, err := scan(rows)
			if err != nil {
				_ = rows.Close()
				return nil, err
			}
			out = append(out, item)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return nil, err
		}
		_ = rows.Close()
	}
	return out, nil
}
