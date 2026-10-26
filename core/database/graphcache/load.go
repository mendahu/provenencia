package graphcache

import (
	"bytes"
	"database/sql"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
)

type entityRow struct {
	id       []byte
	ref      string
	typeID   []byte
	merged   bool
	typeKey  string
	argument string
	label    string
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
	return sqlbatch.Each(ids, g.loadChunk)
}

func (g *Graph) loadChunk(ids [][]byte) error {
	nodes, _, err := g.readNodes(ids)
	if err != nil || len(nodes) == 0 {
		return err
	}
	present := make([][]byte, len(nodes))
	for i, n := range nodes {
		present[i] = n.ID
	}
	links, err := g.loadLinks(present)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		n.Links = links[string(n.ID)]
		g.install(n)
	}
	return nil
}

// readNodes reads the entity rows, values and members of ids and returns the
// nodes of the ids that exist, with no links, and their entity rows. An id with
// no row is removed from the graph. Nothing is installed.
func (g *Graph) readNodes(ids [][]byte) ([]*Node, map[string]entityRow, error) {
	rows, err := g.entityRows(ids)
	if err != nil {
		return nil, nil, err
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
		return nil, rows, nil
	}
	values, err := g.loadValues(present)
	if err != nil {
		return nil, nil, err
	}
	members, err := g.loadMembers(present)
	if err != nil {
		return nil, nil, err
	}
	nodes := make([]*Node, 0, len(present))
	for _, id := range present {
		row := rows[string(id)]
		n := &Node{
			ID:            cloneID(row.id),
			Ref:           row.ref,
			SubjectTypeID: cloneID(row.typeID),
			Argument:      row.argument,
			Label:         row.label,
			Merged:        row.merged,
			Values:        values[string(id)],
			Members:       members[string(id)],
		}
		if n.Values == nil {
			n.Values = map[string][]Value{}
		}
		nodes = append(nodes, n)
	}
	return nodes, rows, nil
}

func (g *Graph) entityRows(ids [][]byte) (map[string]entityRow, error) {
	out := map[string]entityRow{}
	ids = uniqueIDs(ids)
	head := `SELECT e.id, e.ref, e.subject_type_id, e.merged_into_id IS NOT NULL, st.key,
			COALESCE(e.argument, ''), COALESCE(e.label, '')
		FROM ` + g.entityTable + ` e
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE e.id IN (`
	err := sqlbatch.Query(g, head, `)`, nil, ids, func(rows *sql.Rows) error {
		var row entityRow
		if err := rows.Scan(&row.id, &row.ref, &row.typeID, &row.merged, &row.typeKey, &row.argument, &row.label); err != nil {
			return err
		}
		row.id = cloneID(row.id)
		row.typeID = cloneID(row.typeID)
		out[string(row.id)] = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (g *Graph) loadValues(ids [][]byte) (map[string]map[string][]Value, error) {
	out := map[string]map[string][]Value{}
	q := `SELECT r.entity_id, r.property_id, r.rank, r.reason,
			r.value_text, r.value_integer, r.value_term_id, COALESCE(t.key, ''),
			r.value_entity_id, r.value_date, r.value_name, r.value_image_file_id, r.support, r.against
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
			imageID          []byte
			support, against int
		)
		if err := rows.Scan(&entityID, &propID, &rank, &reason, &text, &integer, &termID, &termKey, &entityVal, &dateBlob, &nameBlob, &imageID, &support, &against); err != nil {
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
			EntityID:    cloneID(entityVal),
			Date:        append([]byte(nil), dateBlob...),
			Name:        append([]byte(nil), nameBlob...),
			ImageFileID: cloneID(imageID),
			Support:     support,
			Against:     against,
		})
	}
	return out, rows.Err()
}

func (g *Graph) loadMembers(ids [][]byte) (map[string][]Member, error) {
	out := map[string][]Member{}
	q := `SELECT c.entity_id, c.subject_id, s.source_id, s.graph_id, c.status FROM ` + g.claimTable + ` c
		JOIN subjects s ON s.id = c.subject_id
		WHERE c.status IN ('accepted', 'provisional') AND c.entity_id IN (` + placeholders(len(ids)) + `)
		ORDER BY c.id`
	rows, err := g.query(q, blobArgs(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entityID, subjectID, sourceID, graphID []byte
		var status string
		if err := rows.Scan(&entityID, &subjectID, &sourceID, &graphID, &status); err != nil {
			return nil, err
		}
		out[string(entityID)] = append(out[string(entityID)], Member{
			SubjectID: cloneID(subjectID),
			SourceID:  cloneID(sourceID),
			GraphID:   cloneID(graphID),
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
	q, args := linkLoadSQL(g.valueTable, g.entityTable, ids)
	rows, err := g.query(q, args...)
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
		if !keepRaw(r) {
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

// keepRaw is whether a read link is one the graph keeps: its association is a
// registered bridge and the two property keys are that bridge's two ends. Both
// link loaders keep exactly these.
func keepRaw(r linkRaw) bool {
	b, ok := connectrules.LookupBridge(r.bridgeKey)
	return ok && bridgeEndpoints(b, r.selfPropKey, r.otherPropKey)
}

// LinkLoadSQL is the association-edge query loadLinks runs.
func LinkLoadSQL(ids [][]byte) (string, []any) {
	g := New(nil)
	return linkLoadSQL(g.valueTable, g.entityTable, ids)
}

// linkLoadSQL constrains the association type and endpoint property to the
// registered bridges, each at its own origin. CROSS JOIN and INDEXED BY keep
// SQLite on auto_reconciler_values_edge_idx (property_id, value_entity_id)
// instead of starting at every unmerged entity.
func linkLoadSQL(valueTable, entityTable string, ids [][]byte) (string, []any) {
	types, props := bridgePredicates()
	q := `SELECT a.value_entity_id, a.entity_id, assoc.ref, ast.key, a.property_id, ap.key, bp.key,
			b.value_entity_id, n.ref, n.subject_type_id, nt.key
		FROM ` + valueTable + ` a INDEXED BY auto_reconciler_values_edge_idx
		CROSS JOIN ` + entityTable + ` assoc ON assoc.id = a.entity_id AND assoc.merged_into_id IS NULL
		CROSS JOIN subject_types ast ON ast.id = assoc.subject_type_id AND ` + pairIn("ast.key", "ast.origin", len(types)) + `
		CROSS JOIN properties ap ON ap.id = a.property_id
		CROSS JOIN ` + valueTable + ` b ON b.entity_id = a.entity_id AND b.rank = 1 AND b.reason = 'kept'
			AND b.value_entity_id IS NOT NULL AND b.property_id != a.property_id
		CROSS JOIN properties bp ON bp.id = b.property_id
		CROSS JOIN ` + entityTable + ` n ON n.id = b.value_entity_id AND n.merged_into_id IS NULL
		CROSS JOIN subject_types nt ON nt.id = n.subject_type_id
		WHERE a.rank = 1 AND a.reason = 'kept' AND a.value_entity_id IS NOT NULL
			AND a.value_entity_id IN (` + placeholders(len(ids)) + `)
			AND a.property_id IN (SELECT id FROM properties WHERE ` + pairIn("key", "origin", len(props)) + `)`
	args := make([]any, 0, len(types)*2+len(ids)+len(props)*2)
	args = append(args, pairArgs(types)...)
	args = append(args, blobArgs(ids)...)
	args = append(args, pairArgs(props)...)
	return q, args
}

// LinkScanSQL is the three catalog-wide queries scanLinks runs: the
// associations, their endpoints and their disambiguation terms.
func LinkScanSQL() (assocQ, endsQ, termsQ string, assocArgs, endsArgs, termsArgs []any) {
	g := New(nil)
	return linkScanSQL(g.valueTable, g.entityTable)
}

// linkScanSQL reads the links of every handle at once, so no query carries a
// list of ids. Associations are every unmerged entity of a registered bridge
// type (id, ref, type key). Endpoints are the kept rank-1 entity values of the
// bridges' endpoint properties, at any origin: the Go side checks the self end
// against the registered (key, origin) pairs, as bridgeEndpoints filters the
// other end. Terms are the kept disambiguation terms. CROSS JOIN and INDEXED BY
// keep SQLite on auto_reconciler_values_edge_idx (property_id, value_entity_id)
// instead of starting at every row.
func linkScanSQL(valueTable, entityTable string) (assocQ, endsQ, termsQ string, assocArgs, endsArgs, termsArgs []any) {
	types, _ := bridgePredicates()
	endKeys, disKeys := bridgePropertyKeys()
	assocQ = `SELECT e.id, e.ref, st.key
		FROM subject_types st
		CROSS JOIN ` + entityTable + ` e ON e.subject_type_id = st.id AND e.merged_into_id IS NULL
		WHERE st.id IN (SELECT id FROM subject_types WHERE ` + pairIn("key", "origin", len(types)) + `)`
	assocArgs = pairArgs(types)
	endsQ = `SELECT a.entity_id, a.property_id, p.key, p.origin, a.value_entity_id
		FROM ` + valueTable + ` a INDEXED BY auto_reconciler_values_edge_idx
		CROSS JOIN properties p ON p.id = a.property_id
		WHERE a.rank = 1 AND a.reason = 'kept' AND a.value_entity_id IS NOT NULL
			AND a.property_id IN (SELECT id FROM properties WHERE key IN (` + placeholders(len(endKeys)) + `))`
	endsArgs = stringArgs(endKeys)
	termsQ = `SELECT r.entity_id, r.property_id, r.value_term_id, COALESCE(t.key, ''), p.key
		FROM ` + valueTable + ` r
		JOIN properties p ON p.id = r.property_id
		LEFT JOIN property_terms t ON t.id = r.value_term_id
		WHERE r.rank = 1 AND r.reason = 'kept' AND r.value_term_id IS NOT NULL
			AND r.property_id IN (SELECT id FROM properties WHERE key IN (` + placeholders(len(disKeys)) + `))`
	termsArgs = stringArgs(disKeys)
	return assocQ, endsQ, termsQ, assocArgs, endsArgs, termsArgs
}

// bridgePropertyKeys are the endpoint property keys of every registered
// bridge and the disambiguation property keys, each once, at any origin.
func bridgePropertyKeys() (ends, disambiguation []string) {
	seenE := map[string]bool{}
	seenD := map[string]bool{}
	for _, b := range connectrules.Bridges() {
		for _, e := range b.Endpoints {
			if !seenE[e.PropertyKey] {
				seenE[e.PropertyKey] = true
				ends = append(ends, e.PropertyKey)
			}
		}
		if connectrules.HasDisambiguation(b.Disambiguation) && !seenD[b.Disambiguation] {
			seenD[b.Disambiguation] = true
			disambiguation = append(disambiguation, b.Disambiguation)
		}
	}
	if len(ends) == 0 {
		ends = []string{""}
	}
	if len(disambiguation) == 0 {
		disambiguation = []string{""}
	}
	return ends, disambiguation
}

func stringArgs(keys []string) []any {
	args := make([]any, len(keys))
	for i, k := range keys {
		args[i] = k
	}
	return args
}

// scanEnd is one endpoint row of an association read by scanLinks.
type scanEnd struct {
	propID          []byte
	propKey, origin string
	entity          []byte
}

// scanLinks is the links of ids (loadLinks's, for every id at once): it reads
// every association in the catalog and keeps the ends that touch ids. known is
// the entity rows already read; neighbors outside it are read here.
func (g *Graph) scanLinks(ids [][]byte, known map[string]entityRow) (map[string][]Link, error) {
	assocQ, endsQ, termsQ, assocArgs, endsArgs, termsArgs := linkScanSQL(g.valueTable, g.entityTable)
	type assocRow struct{ ref, typeKey string }
	assocs := map[string]assocRow{}
	rows, err := g.query(assocQ, assocArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id []byte
		var r assocRow
		if err := rows.Scan(&id, &r.ref, &r.typeKey); err != nil {
			_ = rows.Close()
			return nil, err
		}
		assocs[string(id)] = r
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	ends := map[string][]scanEnd{}
	rows, err = g.query(endsQ, endsArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var assoc, propID, entity []byte
		var e scanEnd
		if err := rows.Scan(&assoc, &propID, &e.propKey, &e.origin, &entity); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if _, ok := assocs[string(assoc)]; !ok {
			continue
		}
		e.propID, e.entity = cloneID(propID), cloneID(entity)
		ends[string(assoc)] = append(ends[string(assoc)], e)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	// The mark of an association and property key is its least term key, as
	// disambiguation's ORDER BY t.key takes the first.
	marks := map[string]termMark{}
	rows, err = g.query(termsQ, termsArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var entityID, propID, termID []byte
		var key, propKey string
		if err := rows.Scan(&entityID, &propID, &termID, &key, &propKey); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if _, ok := assocs[string(entityID)]; !ok {
			continue
		}
		mk := string(entityID) + "|" + propKey
		if cur, taken := marks[mk]; !taken || key < cur.key {
			marks[mk] = termMark{propID: cloneID(propID), termID: cloneID(termID), key: key}
		}
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	_, registered := bridgePredicates()
	isRegistered := make(map[[2]string]bool, len(registered))
	for _, p := range registered {
		isRegistered[p] = true
	}
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[string(id)] = true
	}
	var missing [][]byte
	for _, list := range ends {
		touches := false
		for _, e := range list {
			touches = touches || want[string(e.entity)]
		}
		if !touches {
			continue
		}
		for _, e := range list {
			if _, ok := known[string(e.entity)]; !ok {
				missing = append(missing, e.entity)
			}
		}
	}
	more, err := g.entityRows(missing)
	if err != nil {
		return nil, err
	}
	entityOf := func(id []byte) (entityRow, bool) {
		if r, ok := known[string(id)]; ok {
			return r, true
		}
		r, ok := more[string(id)]
		return r, ok
	}
	var raws []linkRaw
	for assoc, list := range ends {
		ar := assocs[assoc]
		for _, self := range list {
			if !want[string(self.entity)] || !isRegistered[[2]string{self.propKey, self.origin}] {
				continue
			}
			for _, other := range list {
				if bytes.Equal(other.propID, self.propID) {
					continue
				}
				n, ok := entityOf(other.entity)
				if !ok || n.merged {
					continue
				}
				r := linkRaw{
					self: self.entity, assoc: []byte(assoc), neighbor: other.entity,
					selfProp: self.propID, neighborType: n.typeID,
					assocRef: ar.ref, bridgeKey: ar.typeKey,
					selfPropKey: self.propKey, otherPropKey: other.propKey,
					neighborRef: n.ref, neighborTypeKey: n.typeKey,
				}
				if keepRaw(r) {
					raws = append(raws, r)
				}
			}
		}
	}
	return assembleLinks(raws, marks), nil
}

func bridgePredicates() (types, props [][2]string) {
	seenT := map[string]bool{}
	seenP := map[string]bool{}
	for _, b := range connectrules.Bridges() {
		tk := b.BridgeTypeKey + "\x00" + b.Origin
		if !seenT[tk] {
			seenT[tk] = true
			types = append(types, [2]string{b.BridgeTypeKey, b.Origin})
		}
		for _, e := range b.Endpoints {
			pk := e.PropertyKey + "\x00" + b.Origin
			if seenP[pk] {
				continue
			}
			seenP[pk] = true
			props = append(props, [2]string{e.PropertyKey, b.Origin})
		}
	}
	if len(types) == 0 {
		types = [][2]string{{"", ""}}
	}
	if len(props) == 0 {
		props = [][2]string{{"", ""}}
	}
	return types, props
}

func pairIn(a, b string, n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "(?, ?)"
	}
	return "(" + a + ", " + b + ") IN (" + strings.Join(parts, ", ") + ")"
}

func pairArgs(pairs [][2]string) []any {
	args := make([]any, 0, len(pairs)*2)
	for _, p := range pairs {
		args = append(args, p[0], p[1])
	}
	return args
}

func bridgeEndpoints(b connectrules.Bridge, selfKey, otherKey string) bool {
	if len(b.Endpoints) != 2 {
		return false
	}
	first, second := b.Endpoints[0].PropertyKey, b.Endpoints[1].PropertyKey
	return (selfKey == first && otherKey == second) || (selfKey == second && otherKey == first)
}

type termMark struct {
	propID []byte
	termID []byte
	key    string
}

func (g *Graph) finishLinks(raws []linkRaw, assocIDs [][]byte) (map[string][]Link, error) {
	marks, err := g.disambiguation(uniqueIDs(assocIDs))
	if err != nil {
		return nil, err
	}
	return assembleLinks(raws, marks), nil
}

// assembleLinks is each node's links from the raw links read for it: one link
// for each (self, association, neighbor), the first of them in (self,
// association, self property, neighbor) order, so a relationship of a person to
// themselves reads the same whatever order the rows came in. Links are ordered
// by association ref, then neighbor ref. marks is each association's
// disambiguation by association id and property key.
func assembleLinks(raws []linkRaw, marks map[string]termMark) map[string][]Link {
	sort.SliceStable(raws, func(i, j int) bool {
		a, b := raws[i], raws[j]
		if c := bytes.Compare(a.self, b.self); c != 0 {
			return c < 0
		}
		if c := bytes.Compare(a.assoc, b.assoc); c != 0 {
			return c < 0
		}
		if c := bytes.Compare(a.selfProp, b.selfProp); c != 0 {
			return c < 0
		}
		return bytes.Compare(a.neighbor, b.neighbor) < 0
	})
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
	return out
}

func (g *Graph) disambiguation(assocIDs [][]byte) (map[string]termMark, error) {
	out := map[string]termMark{}
	if len(assocIDs) == 0 {
		return out, nil
	}
	head := `SELECT r.entity_id, r.property_id, r.value_term_id, COALESCE(t.key, ''), p.key
		FROM ` + g.valueTable + ` r
		JOIN properties p ON p.id = r.property_id
		LEFT JOIN property_terms t ON t.id = r.value_term_id
		WHERE r.rank = 1 AND r.reason = 'kept' AND r.value_term_id IS NOT NULL
			AND r.entity_id IN (`
	err := sqlbatch.Query(g, head, `)
		ORDER BY r.entity_id, t.key`, nil, assocIDs, func(rows *sql.Rows) error {
		var entityID, propID, termID []byte
		var key, propKey string
		if err := rows.Scan(&entityID, &propID, &termID, &key, &propKey); err != nil {
			return err
		}
		mk := string(entityID) + "|" + propKey
		if _, taken := out[mk]; taken || !disambiguationProperty(propKey) {
			return nil
		}
		out[mk] = termMark{
			propID: cloneID(propID), termID: cloneID(termID),
			key: key,
		}
		return nil
	})
	if err != nil {
		return nil, err
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
	g.dropDisplay(n.ID)
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
	g.dropDisplay(id)
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
