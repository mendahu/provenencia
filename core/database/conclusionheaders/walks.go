package conclusionheaders

import (
	"bytes"
	"database/sql"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// walkBatch bounds an IN list. A list is still a fixed number of queries.
const walkBatch = 500

// A subject-role participation joins a person to an event. Role stays off
// the association's identity; the walk reads the reconciled rank-1 role.
const sqlLifeRows = `
SELECT pe.value_entity_id, tt.key, ev.id, ev.ref,
	d.value_date,
	(SELECT COUNT(*) FROM auto_reconciler_values c
		WHERE c.entity_id = ev.id AND c.property_id = dp.id AND c.reason = 'kept'),
	sd.value_date,
	(SELECT COUNT(*) FROM auto_reconciler_values c
		WHERE c.entity_id = ev.id AND c.property_id = sp.id AND c.reason = 'kept'),
	pl.id, pl.ref, tv.value_text, COALESCE(tv.rank, 0),
	(SELECT COUNT(*) FROM auto_reconciler_values c
		WHERE c.entity_id = pl.id AND c.property_id = tp.id AND c.reason = 'kept')
FROM auto_reconciler_values pe
JOIN properties pp ON pp.id = pe.property_id AND pp.key = 'person' AND pp.origin = 'provenencia'
JOIN canonical_entities assoc ON assoc.id = pe.entity_id
JOIN subject_types ast ON ast.id = assoc.subject_type_id
	AND ast.key = 'participation' AND ast.origin = 'provenencia'
JOIN auto_reconciler_values role ON role.entity_id = assoc.id AND role.rank = 1 AND role.reason = 'kept'
JOIN properties rp ON rp.id = role.property_id AND rp.key = 'role' AND rp.origin = 'provenencia'
JOIN property_terms rt ON rt.id = role.value_term_id AND rt.key = 'subject'
JOIN auto_reconciler_values ee ON ee.entity_id = assoc.id AND ee.rank = 1 AND ee.reason = 'kept'
JOIN properties ep ON ep.id = ee.property_id AND ep.key = 'event' AND ep.origin = 'provenencia'
JOIN canonical_entities ev ON ev.id = ee.value_entity_id AND ev.merged_into_id IS NULL
JOIN auto_reconciler_values et ON et.entity_id = ev.id AND et.rank = 1 AND et.reason = 'kept'
JOIN properties etp ON etp.id = et.property_id AND etp.key = 'event_type' AND etp.origin = 'provenencia'
JOIN property_terms tt ON tt.id = et.value_term_id AND tt.key IN ('birth', 'death')
LEFT JOIN properties dp ON dp.key = 'date' AND dp.origin = 'provenencia'
LEFT JOIN auto_reconciler_values d
	ON d.entity_id = ev.id AND d.property_id = dp.id AND d.rank = 1 AND d.reason = 'kept'
LEFT JOIN properties sp ON sp.key = 'start_date' AND sp.origin = 'provenencia'
LEFT JOIN auto_reconciler_values sd
	ON sd.entity_id = ev.id AND sd.property_id = sp.id AND sd.rank = 1 AND sd.reason = 'kept'
LEFT JOIN auto_reconciler_values lee
	ON lee.value_entity_id = ev.id AND lee.rank = 1 AND lee.reason = 'kept'
LEFT JOIN properties lep ON lep.id = lee.property_id AND lep.key = 'event' AND lep.origin = 'provenencia'
LEFT JOIN canonical_entities loc ON loc.id = lee.entity_id AND lep.id IS NOT NULL
LEFT JOIN subject_types lst ON lst.id = loc.subject_type_id
	AND lst.key = 'location' AND lst.origin = 'provenencia'
LEFT JOIN auto_reconciler_values lpe
	ON lst.id IS NOT NULL AND lpe.entity_id = loc.id AND lpe.rank = 1 AND lpe.reason = 'kept'
LEFT JOIN properties lpp ON lpp.id = lpe.property_id AND lpp.key = 'place' AND lpp.origin = 'provenencia'
LEFT JOIN canonical_entities pl ON pl.id = lpe.value_entity_id AND pl.merged_into_id IS NULL AND lpp.id IS NOT NULL
LEFT JOIN properties tp ON tp.key = 'toponym' AND tp.origin = 'provenencia'
LEFT JOIN auto_reconciler_values tv
	ON tv.entity_id = pl.id AND tv.property_id = tp.id AND tv.reason = 'kept'
WHERE pe.rank = 1 AND pe.reason = 'kept' AND pe.value_entity_id IN (`

const sqlEventSubjects = `
SELECT ee.value_entity_id, assoc.ref,
	per.id, per.subject_type_id, per.ref, COALESCE(per.argument, ''), COALESCE(per.label, ''),
	nv.value_name,
	(SELECT COUNT(*) FROM auto_reconciler_values c
		WHERE c.entity_id = per.id AND c.property_id = np.id AND c.reason = 'kept')
FROM auto_reconciler_values ee
JOIN properties ep ON ep.id = ee.property_id AND ep.key = 'event' AND ep.origin = 'provenencia'
JOIN canonical_entities assoc ON assoc.id = ee.entity_id
JOIN subject_types ast ON ast.id = assoc.subject_type_id
	AND ast.key = 'participation' AND ast.origin = 'provenencia'
JOIN auto_reconciler_values role ON role.entity_id = assoc.id AND role.rank = 1 AND role.reason = 'kept'
JOIN properties rp ON rp.id = role.property_id AND rp.key = 'role' AND rp.origin = 'provenencia'
JOIN property_terms rt ON rt.id = role.value_term_id AND rt.key = 'subject'
JOIN auto_reconciler_values pe ON pe.entity_id = assoc.id AND pe.rank = 1 AND pe.reason = 'kept'
JOIN properties pp ON pp.id = pe.property_id AND pp.key = 'person' AND pp.origin = 'provenencia'
JOIN canonical_entities per ON per.id = pe.value_entity_id AND per.merged_into_id IS NULL
LEFT JOIN properties np ON np.key = 'name' AND np.origin = 'provenencia'
LEFT JOIN auto_reconciler_values nv
	ON nv.entity_id = per.id AND nv.property_id = np.id AND nv.rank = 1 AND nv.reason = 'kept'
WHERE ee.rank = 1 AND ee.reason = 'kept' AND ee.value_entity_id IN (`

const sqlEventPlaces = `
SELECT ee.value_entity_id, pl.id, pl.ref, tv.value_text, COALESCE(tv.rank, 0),
	(SELECT COUNT(*) FROM auto_reconciler_values c
		WHERE c.entity_id = pl.id AND c.property_id = tp.id AND c.reason = 'kept')
FROM auto_reconciler_values ee
JOIN properties ep ON ep.id = ee.property_id AND ep.key = 'event' AND ep.origin = 'provenencia'
JOIN canonical_entities loc ON loc.id = ee.entity_id
JOIN subject_types lst ON lst.id = loc.subject_type_id
	AND lst.key = 'location' AND lst.origin = 'provenencia'
JOIN auto_reconciler_values lpe ON lpe.entity_id = loc.id AND lpe.rank = 1 AND lpe.reason = 'kept'
JOIN properties lpp ON lpp.id = lpe.property_id AND lpp.key = 'place' AND lpp.origin = 'provenencia'
JOIN canonical_entities pl ON pl.id = lpe.value_entity_id AND pl.merged_into_id IS NULL
LEFT JOIN properties tp ON tp.key = 'toponym' AND tp.origin = 'provenencia'
LEFT JOIN auto_reconciler_values tv
	ON tv.entity_id = pl.id AND tv.property_id = tp.id AND tv.reason = 'kept'
WHERE ee.rank = 1 AND ee.reason = 'kept' AND ee.value_entity_id IN (`

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
	places, err := loadEventPlaces(q, ids)
	if err != nil {
		return err
	}
	for i := range headers {
		headers[i].Subjects = subjects[string(headers[i].Entity.ID)]
		headers[i].Places = places[string(headers[i].Entity.ID)]
	}
	return nil
}

type lifeRow struct {
	personID              []byte
	kind                  string
	eventID               []byte
	eventRef              string
	dateBlob, startBlob   []byte
	dateCount, startCount int
	placeID               []byte
	placeRef, name        string
	nameRank, placeCount  int
}

func loadLives(q Querier, personIDs [][]byte) (map[string]map[string]LifeFacts, error) {
	rows, err := queryBatches(q, sqlLifeRows, personIDs, func(sc scanner) (lifeRow, error) {
		var r lifeRow
		var name, placeRef sql.NullString
		if err := sc.Scan(&r.personID, &r.kind, &r.eventID, &r.eventRef,
			&r.dateBlob, &r.dateCount, &r.startBlob, &r.startCount,
			&r.placeID, &placeRef, &name, &r.nameRank, &r.placeCount); err != nil {
			return lifeRow{}, err
		}
		r.placeRef = placeRef.String
		r.personID = append([]byte(nil), r.personID...)
		r.eventID = append([]byte(nil), r.eventID...)
		r.dateBlob = append([]byte(nil), r.dateBlob...)
		r.startBlob = append([]byte(nil), r.startBlob...)
		r.placeID = append([]byte(nil), r.placeID...)
		if name.Valid {
			r.name = strings.TrimSpace(name.String)
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	// person → kind → event ref → accumulator. The lowest event ref wins
	// when several births (or deaths) survive.
	type ev struct {
		ref    string
		date   *datevalues.Value
		count  int
		places map[string]*placeBuild
	}
	byPerson := map[string]map[string]map[string]*ev{}
	for _, r := range rows {
		kind := byPerson[string(r.personID)]
		if kind == nil {
			kind = map[string]map[string]*ev{}
			byPerson[string(r.personID)] = kind
		}
		events := kind[r.kind]
		if events == nil {
			events = map[string]*ev{}
			kind[r.kind] = events
		}
		e := events[string(r.eventID)]
		if e == nil {
			dateBlob, count := r.dateBlob, r.dateCount
			if len(dateBlob) == 0 {
				dateBlob, count = r.startBlob, r.startCount
			}
			date, err := unmarshalDate(dateBlob)
			if err != nil {
				return nil, err
			}
			e = &ev{ref: r.eventRef, date: date, count: count, places: map[string]*placeBuild{}}
			events[string(r.eventID)] = e
		}
		if len(r.placeID) == 0 {
			continue
		}
		e.places[string(r.placeID)] = addPlaceName(e.places[string(r.placeID)], r.placeRef, r.name, r.nameRank, r.placeCount)
	}
	out := map[string]map[string]LifeFacts{}
	for person, kinds := range byPerson {
		out[person] = map[string]LifeFacts{}
		for kind, events := range kinds {
			var best *ev
			for _, e := range events {
				if best == nil || e.ref < best.ref {
					best = e
				}
			}
			out[person][kind] = LifeFacts{Date: best.date, DateCount: best.count, Places: placesOf(best.places)}
		}
	}
	return out, nil
}

type subjectRow struct {
	eventID []byte
	partRef string
	subject EventSubject
}

func loadEventSubjects(q Querier, eventIDs [][]byte) (map[string][]EventSubject, error) {
	rows, err := queryBatches(q, sqlEventSubjects, eventIDs, func(sc scanner) (subjectRow, error) {
		var (
			r        subjectRow
			nameBlob []byte
		)
		e := &r.subject.Entity
		if err := sc.Scan(&r.eventID, &r.partRef, &e.ID, &e.SubjectTypeID, &e.Ref, &e.Argument, &e.Label,
			&nameBlob, &r.subject.NameValueCount); err != nil {
			return subjectRow{}, err
		}
		r.eventID = append([]byte(nil), r.eventID...)
		e.ID = append([]byte(nil), e.ID...)
		e.SubjectTypeID = append([]byte(nil), e.SubjectTypeID...)
		if len(nameBlob) > 0 {
			n, err := valuecodec.UnmarshalName(nameBlob)
			if err != nil {
				return subjectRow{}, err
			}
			r.subject.Name = &n
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].partRef != rows[j].partRef {
			return rows[i].partRef < rows[j].partRef
		}
		return rows[i].subject.Entity.Ref < rows[j].subject.Entity.Ref
	})
	out := map[string][]EventSubject{}
	seen := map[string]bool{}
	for _, r := range rows {
		key := string(r.eventID) + "\x00" + string(r.subject.Entity.ID)
		if seen[key] {
			continue
		}
		seen[key] = true
		out[string(r.eventID)] = append(out[string(r.eventID)], r.subject)
	}
	return out, nil
}

type placeRow struct {
	eventID []byte
	placeID []byte
	ref     string
	name    string
	rank    int
	count   int
}

func loadEventPlaces(q Querier, eventIDs [][]byte) (map[string][]HeaderPlace, error) {
	rows, err := queryBatches(q, sqlEventPlaces, eventIDs, func(sc scanner) (placeRow, error) {
		var r placeRow
		var name *string
		if err := sc.Scan(&r.eventID, &r.placeID, &r.ref, &name, &r.rank, &r.count); err != nil {
			return placeRow{}, err
		}
		r.eventID = append([]byte(nil), r.eventID...)
		r.placeID = append([]byte(nil), r.placeID...)
		if name != nil {
			r.name = strings.TrimSpace(*name)
		}
		return r, nil
	})
	if err != nil {
		return nil, err
	}
	byEvent := map[string]map[string]*placeBuild{}
	for _, r := range rows {
		places := byEvent[string(r.eventID)]
		if places == nil {
			places = map[string]*placeBuild{}
			byEvent[string(r.eventID)] = places
		}
		places[string(r.placeID)] = addPlaceName(places[string(r.placeID)], r.ref, r.name, r.rank, r.count)
	}
	out := map[string][]HeaderPlace{}
	for event, places := range byEvent {
		out[event] = placesOf(places)
	}
	return out, nil
}

type named struct {
	rank int
	text string
}

type placeBuild struct {
	ref   string
	count int
	names []named
}

func addPlaceName(p *placeBuild, ref, name string, rank, count int) *placeBuild {
	if p == nil {
		p = &placeBuild{ref: ref, count: count}
	}
	if name == "" {
		return p
	}
	for _, n := range p.names {
		if n.rank == rank && n.text == name {
			return p
		}
	}
	p.names = append(p.names, named{rank: rank, text: name})
	return p
}

func placesOf(places map[string]*placeBuild) []HeaderPlace {
	list := make([]*placeBuild, 0, len(places))
	for _, p := range places {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ref < list[j].ref })
	var out []HeaderPlace
	for _, p := range list {
		sort.Slice(p.names, func(i, j int) bool { return p.names[i].rank < p.names[j].rank })
		h := HeaderPlace{Count: p.count}
		for _, n := range p.names {
			h.Names = append(h.Names, n.text)
		}
		out = append(out, h)
	}
	return out
}

// HeaderDependents is the reverse of the header walks, for a later search
// reprojection (S9-34). A Place reaches its locations' events and those
// events' subject persons. A Person reaches the events of their subject-role
// participations. Association handles are not returned. This does not
// reproject.
func HeaderDependents(q Querier, ids [][]byte) ([][]byte, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	events, err := dependentEvents(q, sqlPlaceEvents, ids)
	if err != nil {
		return nil, err
	}
	persons, err := dependentEvents(q, sqlEventPersons, events)
	if err != nil {
		return nil, err
	}
	fromPerson, err := dependentEvents(q, sqlPersonEvents, ids)
	if err != nil {
		return nil, err
	}
	var out [][]byte
	seen := map[string]bool{}
	add := func(id []byte) {
		if len(id) != 16 || seen[string(id)] {
			return
		}
		seen[string(id)] = true
		out = append(out, id)
	}
	for _, id := range events {
		add(id)
	}
	for _, id := range persons {
		add(id)
	}
	for _, id := range fromPerson {
		add(id)
	}
	sort.Slice(out, func(i, j int) bool { return bytes.Compare(out[i], out[j]) < 0 })
	return out, nil
}

const sqlPlaceEvents = `
SELECT DISTINCT ee.value_entity_id
FROM auto_reconciler_values pe
JOIN properties pp ON pp.id = pe.property_id AND pp.key = 'place' AND pp.origin = 'provenencia'
JOIN canonical_entities loc ON loc.id = pe.entity_id
JOIN subject_types lst ON lst.id = loc.subject_type_id
	AND lst.key = 'location' AND lst.origin = 'provenencia'
JOIN auto_reconciler_values ee ON ee.entity_id = loc.id AND ee.rank = 1 AND ee.reason = 'kept'
JOIN properties ep ON ep.id = ee.property_id AND ep.key = 'event' AND ep.origin = 'provenencia'
WHERE pe.rank = 1 AND pe.reason = 'kept' AND pe.value_entity_id IN (`

const sqlEventPersons = `
SELECT DISTINCT pe.value_entity_id
FROM auto_reconciler_values ee
JOIN properties ep ON ep.id = ee.property_id AND ep.key = 'event' AND ep.origin = 'provenencia'
JOIN canonical_entities assoc ON assoc.id = ee.entity_id
JOIN subject_types ast ON ast.id = assoc.subject_type_id
	AND ast.key = 'participation' AND ast.origin = 'provenencia'
JOIN auto_reconciler_values role ON role.entity_id = assoc.id AND role.rank = 1 AND role.reason = 'kept'
JOIN properties rp ON rp.id = role.property_id AND rp.key = 'role' AND rp.origin = 'provenencia'
JOIN property_terms rt ON rt.id = role.value_term_id AND rt.key = 'subject'
JOIN auto_reconciler_values pe ON pe.entity_id = assoc.id AND pe.rank = 1 AND pe.reason = 'kept'
JOIN properties pp ON pp.id = pe.property_id AND pp.key = 'person' AND pp.origin = 'provenencia'
WHERE ee.rank = 1 AND ee.reason = 'kept' AND ee.value_entity_id IN (`

const sqlPersonEvents = `
SELECT DISTINCT ee.value_entity_id
FROM auto_reconciler_values pe
JOIN properties pp ON pp.id = pe.property_id AND pp.key = 'person' AND pp.origin = 'provenencia'
JOIN canonical_entities assoc ON assoc.id = pe.entity_id
JOIN subject_types ast ON ast.id = assoc.subject_type_id
	AND ast.key = 'participation' AND ast.origin = 'provenencia'
JOIN auto_reconciler_values role ON role.entity_id = assoc.id AND role.rank = 1 AND role.reason = 'kept'
JOIN properties rp ON rp.id = role.property_id AND rp.key = 'role' AND rp.origin = 'provenencia'
JOIN property_terms rt ON rt.id = role.value_term_id AND rt.key = 'subject'
JOIN auto_reconciler_values ee ON ee.entity_id = assoc.id AND ee.rank = 1 AND ee.reason = 'kept'
JOIN properties ep ON ep.id = ee.property_id AND ep.key = 'event' AND ep.origin = 'provenencia'
WHERE pe.rank = 1 AND pe.reason = 'kept' AND pe.value_entity_id IN (`

func dependentEvents(q Querier, query string, ids [][]byte) ([][]byte, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return queryBatches(q, query, ids, func(sc scanner) ([]byte, error) {
		var id []byte
		if err := sc.Scan(&id); err != nil {
			return nil, err
		}
		return append([]byte(nil), id...), nil
	})
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
