package graphcache

import (
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
)

type entityRow struct {
	id      []byte
	ref     string
	typeID  []byte
	merged  bool
	typeKey string
}

// loadIDs loads ids that are not already cached. replace reloads ids that are.
func (g *Graph) loadIDs(ids [][]byte, replace bool) error {
	ids = uniqueIDs(ids)
	if !replace {
		var fresh [][]byte
		for _, id := range ids {
			if _, ok := g.nodes[string(id)]; !ok {
				fresh = append(fresh, id)
			}
		}
		ids = fresh
	}
	if len(ids) == 0 {
		return nil
	}
	for start := 0; start < len(ids); start += batch {
		chunk := ids[start:min(start+batch, len(ids))]
		if err := g.loadChunk(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) loadChunk(ids [][]byte) error {
	rows, err := g.entityRows(ids)
	if err != nil {
		return err
	}
	present := make([][]byte, 0, len(rows))
	for _, id := range ids {
		if _, ok := rows[string(id)]; ok {
			present = append(present, id)
		} else {
			g.removeNode(id)
		}
	}
	if len(present) == 0 {
		return nil
	}
	values, err := g.loadValues(present)
	if err != nil {
		return err
	}
	members, err := g.loadMembers(present)
	if err != nil {
		return err
	}
	links, err := g.loadLinks(present)
	if err != nil {
		return err
	}
	for _, id := range present {
		row := rows[string(id)]
		n := &Node{
			ID:            cloneID(row.id),
			Ref:           row.ref,
			SubjectTypeID: cloneID(row.typeID),
			Merged:        row.merged,
			Values:        values[string(id)],
			Links:         links[string(id)],
			Members:       members[string(id)],
		}
		if n.Values == nil {
			n.Values = map[string][]Value{}
		}
		g.install(n)
	}
	return nil
}

func (g *Graph) entityRows(ids [][]byte) (map[string]entityRow, error) {
	out := map[string]entityRow{}
	ids = uniqueIDs(ids)
	for start := 0; start < len(ids); start += batch {
		chunk := ids[start:min(start+batch, len(ids))]
		q := `SELECT e.id, e.ref, e.subject_type_id, e.merged_into_id IS NOT NULL, st.key
			FROM ` + g.entityTable + ` e
			JOIN subject_types st ON st.id = e.subject_type_id
			WHERE e.id IN (` + placeholders(len(chunk)) + `)`
		rows, err := g.query(q, blobArgs(chunk)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var row entityRow
			if err := rows.Scan(&row.id, &row.ref, &row.typeID, &row.merged, &row.typeKey); err != nil {
				_ = rows.Close()
				return nil, err
			}
			row.id = cloneID(row.id)
			row.typeID = cloneID(row.typeID)
			out[string(row.id)] = row
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (g *Graph) loadValues(ids [][]byte) (map[string]map[string][]Value, error) {
	out := map[string]map[string][]Value{}
	q := `SELECT r.entity_id, r.property_id, r.rank, r.reason,
			r.value_text, r.value_integer, r.value_term_id, COALESCE(t.key, ''),
			r.value_entity_id, r.value_date, r.value_name
		FROM ` + g.valueTable + ` r
		LEFT JOIN property_terms t ON t.id = r.value_term_id
		WHERE r.entity_id IN (` + placeholders(len(ids)) + `)
		ORDER BY r.entity_id, r.property_id, r.rank`
	rows, err := g.query(q, blobArgs(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			entityID, propID []byte
			rank             int
			reason           string
			text             sql.NullString
			integer          sql.NullInt64
			termID           []byte
			termKey          string
			entityVal        []byte
			dateBlob         []byte
			nameBlob         []byte
		)
		if err := rows.Scan(&entityID, &propID, &rank, &reason, &text, &integer, &termID, &termKey, &entityVal, &dateBlob, &nameBlob); err != nil {
			return nil, err
		}
		ek, pk := string(entityID), string(propID)
		if out[ek] == nil {
			out[ek] = map[string][]Value{}
		}
		out[ek][pk] = append(out[ek][pk], Value{
			Rank: rank, Reason: reason,
			Text: text.String, HasText: text.Valid,
			Integer: integer.Int64, HasInteger: integer.Valid,
			TermID: cloneID(termID), TermKey: termKey,
			EntityID: cloneID(entityVal),
			Date:     append([]byte(nil), dateBlob...),
			Name:     append([]byte(nil), nameBlob...),
		})
	}
	return out, rows.Err()
}

func (g *Graph) loadMembers(ids [][]byte) (map[string][]Member, error) {
	out := map[string][]Member{}
	q := `SELECT entity_id, subject_id, status FROM ` + g.claimTable + `
		WHERE status IN ('accepted', 'provisional') AND entity_id IN (` + placeholders(len(ids)) + `)
		ORDER BY id`
	rows, err := g.query(q, blobArgs(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entityID, subjectID []byte
		var status string
		if err := rows.Scan(&entityID, &subjectID, &status); err != nil {
			return nil, err
		}
		out[string(entityID)] = append(out[string(entityID)], Member{
			SubjectID: cloneID(subjectID),
			Accepted:  status == "accepted",
		})
	}
	return out, rows.Err()
}

type linkRaw struct {
	self, assoc, neighbor        []byte
	selfProp, neighborType       []byte
	assocRef, bridgeKey          string
	selfPropKey, otherPropKey    string
	neighborRef, neighborTypeKey string
}

func (g *Graph) loadLinks(ids [][]byte) (map[string][]Link, error) {
	q := `SELECT a.value_entity_id, a.entity_id, assoc.ref, ast.key, a.property_id, ap.key, bp.key,
			b.value_entity_id, n.ref, n.subject_type_id, nt.key
		FROM ` + g.valueTable + ` a
		JOIN ` + g.entityTable + ` assoc ON assoc.id = a.entity_id AND assoc.merged_into_id IS NULL
		JOIN subject_types ast ON ast.id = assoc.subject_type_id AND ast.origin = 'provenencia'
		JOIN properties ap ON ap.id = a.property_id
		JOIN ` + g.valueTable + ` b ON b.entity_id = a.entity_id AND b.rank = 1 AND b.reason = 'kept'
			AND b.value_entity_id IS NOT NULL AND b.property_id != a.property_id
		JOIN properties bp ON bp.id = b.property_id
		JOIN ` + g.entityTable + ` n ON n.id = b.value_entity_id AND n.merged_into_id IS NULL
		JOIN subject_types nt ON nt.id = n.subject_type_id
		WHERE a.rank = 1 AND a.reason = 'kept' AND a.value_entity_id IS NOT NULL
			AND a.value_entity_id IN (` + placeholders(len(ids)) + `)`
	rows, err := g.query(q, blobArgs(ids)...)
	if err != nil {
		return nil, err
	}
	var assocIDs [][]byte
	var raws []linkRaw
	for rows.Next() {
		var r linkRaw
		if err := rows.Scan(&r.self, &r.assoc, &r.assocRef, &r.bridgeKey, &r.selfProp, &r.selfPropKey, &r.otherPropKey,
			&r.neighbor, &r.neighborRef, &r.neighborType, &r.neighborTypeKey); err != nil {
			_ = rows.Close()
			return nil, err
		}
		b, ok := connectrules.LookupBridge(r.bridgeKey)
		if !ok || !bridgeEndpoints(b, r.selfPropKey, r.otherPropKey) {
			continue
		}
		r.self = cloneID(r.self)
		r.assoc = cloneID(r.assoc)
		r.selfProp = cloneID(r.selfProp)
		r.neighbor = cloneID(r.neighbor)
		r.neighborType = cloneID(r.neighborType)
		raws = append(raws, r)
		assocIDs = append(assocIDs, r.assoc)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	_ = rows.Close()
	return g.finishLinks(raws, assocIDs)
}

func bridgeEndpoints(b connectrules.Bridge, selfKey, otherKey string) bool {
	if len(b.Endpoints) != 2 {
		return false
	}
	first, second := b.Endpoints[0].PropertyKey, b.Endpoints[1].PropertyKey
	return (selfKey == first && otherKey == second) || (selfKey == second && otherKey == first)
}

type termMark struct {
	propID     []byte
	termID     []byte
	key        string
	inverseKey string
	directed   bool
}

func (g *Graph) finishLinks(raws []linkRaw, assocIDs [][]byte) (map[string][]Link, error) {
	marks, err := g.disambiguation(uniqueIDs(assocIDs))
	if err != nil {
		return nil, err
	}
	out := map[string][]Link{}
	seen := map[string]bool{}
	for _, r := range raws {
		b, ok := connectrules.LookupBridge(r.bridgeKey)
		if !ok {
			continue
		}
		mark := marks[string(r.assoc)+"|"+b.Disambiguation]
		key := string(r.self) + "|" + string(r.assoc) + "|" + string(r.neighbor)
		if seen[key] {
			continue
		}
		seen[key] = true
		out[string(r.self)] = append(out[string(r.self)], Link{
			Neighbor:                 cloneID(r.neighbor),
			NeighborRef:              r.neighborRef,
			Association:              cloneID(r.assoc),
			AssociationRef:           r.assocRef,
			FromEnd:                  r.selfPropKey == b.Endpoints[0].PropertyKey,
			EndpointPropertyID:       cloneID(r.selfProp),
			DisambiguationPropertyID: cloneID(mark.propID),
			TermID:                   cloneID(mark.termID),
			TermKey:                  mark.key,
			InverseKey:               mark.inverseKey,
			Directed:                 mark.directed,
			NeighborTypeID:           cloneID(r.neighborType),
			BridgeTypeKey:            r.bridgeKey,
			NeighborTypeKey:          r.neighborTypeKey,
		})
	}
	for id, links := range out {
		sort.Slice(links, func(i, j int) bool {
			if links[i].AssociationRef != links[j].AssociationRef {
				return links[i].AssociationRef < links[j].AssociationRef
			}
			return links[i].NeighborRef < links[j].NeighborRef
		})
		out[id] = links
	}
	return out, nil
}

func (g *Graph) disambiguation(assocIDs [][]byte) (map[string]termMark, error) {
	out := map[string]termMark{}
	if len(assocIDs) == 0 {
		return out, nil
	}
	for start := 0; start < len(assocIDs); start += batch {
		chunk := assocIDs[start:min(start+batch, len(assocIDs))]
		q := `SELECT r.entity_id, r.property_id, r.value_term_id, COALESCE(t.key, ''),
				COALESCE(t.inverse_key, ''), COALESCE(t.directed, 0), p.key
			FROM ` + g.valueTable + ` r
			JOIN properties p ON p.id = r.property_id
			LEFT JOIN property_terms t ON t.id = r.value_term_id
			WHERE r.rank = 1 AND r.reason = 'kept' AND r.value_term_id IS NOT NULL
				AND r.entity_id IN (` + placeholders(len(chunk)) + `)
			ORDER BY r.entity_id, t.key`
		rows, err := g.query(q, blobArgs(chunk)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var entityID, propID, termID []byte
			var key, inverse, propKey string
			var directed int
			if err := rows.Scan(&entityID, &propID, &termID, &key, &inverse, &directed, &propKey); err != nil {
				_ = rows.Close()
				return nil, err
			}
			mk := string(entityID) + "|" + propKey
			if _, taken := out[mk]; taken || !disambiguationProperty(propKey) {
				continue
			}
			out[mk] = termMark{
				propID: cloneID(propID), termID: cloneID(termID),
				key: key, inverseKey: inverse, directed: directed != 0,
			}
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func disambiguationProperty(propKey string) bool {
	for _, b := range connectrules.Bridges() {
		if connectrules.HasDisambiguation(b.Disambiguation) && b.Disambiguation == propKey {
			return true
		}
	}
	return false
}

func (g *Graph) unmergedIDs(typeID []byte) ([][]byte, error) {
	rows, err := g.query(`SELECT id FROM `+g.entityTable+`
		WHERE subject_type_id = ? AND merged_into_id IS NULL
		ORDER BY ref, id`, typeID)
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
		out = append(out, cloneID(id))
	}
	return out, rows.Err()
}

func (g *Graph) endpointIDs(assoc []byte) ([][]byte, error) {
	rows, err := g.query(`SELECT value_entity_id FROM `+g.valueTable+`
		WHERE entity_id = ? AND rank = 1 AND reason = 'kept' AND value_entity_id IS NOT NULL`, assoc)
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
		out = append(out, cloneID(id))
	}
	return out, rows.Err()
}

func (g *Graph) install(n *Node) {
	if old, ok := g.nodes[string(n.ID)]; ok {
		g.detach(old)
	}
	g.nodes[string(n.ID)] = n
	g.attach(n)
	g.syncKind(n)
}

func (g *Graph) attach(n *Node) {
	seen := map[string]bool{}
	for _, l := range n.Links {
		k := string(l.Association)
		if seen[k] {
			continue
		}
		seen[k] = true
		g.holders[k] = append(g.holders[k], cloneID(n.ID))
	}
}

func (g *Graph) detach(n *Node) {
	seen := map[string]bool{}
	for _, l := range n.Links {
		k := string(l.Association)
		if seen[k] {
			continue
		}
		seen[k] = true
		g.holders[k] = dropID(g.holders[k], n.ID)
		if len(g.holders[k]) == 0 {
			delete(g.holders, k)
		}
	}
}

func (g *Graph) removeNode(id []byte) {
	if n, ok := g.nodes[string(id)]; ok {
		g.detach(n)
		delete(g.nodes, string(id))
	}
	for k, set := range g.kindSet {
		if !set[string(id)] {
			continue
		}
		delete(set, string(id))
		g.kinds[k] = dropID(g.kinds[k], id)
	}
}

func (g *Graph) syncKind(n *Node) {
	k := string(n.SubjectTypeID)
	if !g.kindOn[k] {
		return
	}
	in := g.kindSet[k][string(n.ID)]
	if n.Merged {
		if in {
			delete(g.kindSet[k], string(n.ID))
			g.kinds[k] = dropID(g.kinds[k], n.ID)
		}
		return
	}
	if !in {
		if g.kindSet[k] == nil {
			g.kindSet[k] = map[string]bool{}
		}
		g.kindSet[k][string(n.ID)] = true
		g.kinds[k] = append(g.kinds[k], cloneID(n.ID))
	}
}

func dropID(ids [][]byte, id []byte) [][]byte {
	out := ids[:0]
	for _, cur := range ids {
		if !sameID(cur, id) {
			out = append(out, cur)
		}
	}
	return out
}
