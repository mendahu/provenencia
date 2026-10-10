package conclusionheaders

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
)

// headerChainDepth is how far a place row's parent chain climbs, and how far
// the reverse of that read walks back down. loadPlaceGraph uses the same bound.
const headerChainDepth = 8

// Dependents is the reverse of the header reads. A place reaches its
// descendants to headerChainDepth, the events located at those places, and
// the people whose birth or death that event is. A person reaches the events
// that list them as a subject. An event that is a birth or a death reaches
// its subject people. Succession, one hop both ways, is included so a later
// place detail is covered. Association handles are not returned. ids
// themselves are not returned; the caller already rewrites those.
func Dependents(q Querier, ids [][]byte) ([][]byte, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	places, err := walkDepth(q, canonicalgraph.PartsOfPlace, ids, headerChainDepth)
	if err != nil {
		return nil, err
	}
	parents, err := walkDepth(q, canonicalgraph.ParentsOfPlace, ids, 1)
	if err != nil {
		return nil, err
	}
	succTo, err := walkDepth(q, canonicalgraph.SuccessorsOfPlace, ids, 1)
	if err != nil {
		return nil, err
	}
	succFrom, err := walkDepth(q, canonicalgraph.PredecessorsOfPlace, ids, 1)
	if err != nil {
		return nil, err
	}
	located, err := walkDepth(q, canonicalgraph.EventsAtPlace, append(append(append([][]byte(nil), ids...), places...), parents...), 1)
	if err != nil {
		return nil, err
	}
	shown, err := walkDepth(q, canonicalgraph.EventsOfSubject, ids, 1)
	if err != nil {
		return nil, err
	}
	life, err := birthOrDeath(q, append(append([][]byte(nil), ids...), located...))
	if err != nil {
		return nil, err
	}
	people, err := walkDepth(q, canonicalgraph.SubjectsOfEvent, life, 1)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, id := range ids {
		seen[string(id)] = true
	}
	var out [][]byte
	for _, group := range [][][]byte{places, parents, succTo, succFrom, located, shown, people} {
		for _, id := range group {
			if seen[string(id)] {
				continue
			}
			seen[string(id)] = true
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return bytes.Compare(out[i], out[j]) < 0 })
	return out, nil
}

// AssociationEndpoints reads the kept entity values of the association handles
// in ids. A save calls it before rewriting those rows and again after, and
// both sets are roots of Dependents. The cache's link index is empty until
// something has loaded the association, so the rows are the source.
func AssociationEndpoints(q Querier, ids [][]byte) ([][]byte, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	bridges := connectrules.Bridges()
	if len(bridges) == 0 {
		return nil, nil
	}
	var where string
	args := make([]any, 0, len(bridges)*2+len(ids))
	for i, b := range bridges {
		if i > 0 {
			where += " OR "
		}
		origin := b.Origin
		if origin == "" {
			origin = connectrules.OriginProvenencia
		}
		where += "(st.origin = ? AND st.key = ?)"
		args = append(args, origin, b.BridgeTypeKey)
	}
	args = append(args, database.BlobArgs(ids)...)
	query := `SELECT DISTINCT r.value_entity_id
		FROM auto_reconciler_values r
		JOIN canonical_entities e ON e.id = r.entity_id
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE r.rank = 1 AND r.reason = 'kept' AND r.value_entity_id IS NOT NULL
			AND (` + where + `)
			AND r.entity_id IN (` + database.SQLInPlaceholders(len(ids)) + `)`
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, append([]byte(nil), id...))
	}
	return out, rows.Err()
}

func walkDepth(q Querier, hop canonicalgraph.Hop, roots [][]byte, depth int) ([][]byte, error) {
	roots = database.UniqueBlobIDs(roots)
	if len(roots) == 0 || depth < 1 {
		return nil, nil
	}
	seen := map[string]bool{}
	for _, id := range roots {
		seen[string(id)] = true
	}
	frontier := roots
	var out [][]byte
	for round := 0; round < depth && len(frontier) > 0; round++ {
		edges, err := canonicalgraph.Walk(q, hop, frontier)
		if err != nil {
			return nil, err
		}
		frontier = nil
		for _, id := range canonicalgraph.Targets(edges) {
			if seen[string(id)] {
				continue
			}
			seen[string(id)] = true
			cp := append([]byte(nil), id...)
			out = append(out, cp)
			frontier = append(frontier, cp)
		}
	}
	return out, nil
}

const sqlBirthOrDeath = `SELECT ev.id
FROM canonical_entities ev
JOIN auto_reconciler_values et ON et.entity_id = ev.id AND et.rank = 1 AND et.reason = 'kept'
JOIN properties etp ON etp.id = et.property_id AND etp.key = 'event_type' AND etp.origin = 'provenencia'
JOIN property_terms tt ON tt.id = et.value_term_id AND tt.key IN ('birth', 'death')
WHERE ev.id IN (`

func birthOrDeath(q Querier, ids [][]byte) ([][]byte, error) {
	ids = database.UniqueBlobIDs(ids)
	var out [][]byte
	for start := 0; start < len(ids); start += walkBatch {
		chunk := ids[start:min(start+walkBatch, len(ids))]
		rows, err := q.Query(sqlBirthOrDeath+database.SQLInPlaceholders(len(chunk))+`)`, database.BlobArgs(chunk)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id []byte
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return nil, err
			}
			out = append(out, append([]byte(nil), id...))
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
