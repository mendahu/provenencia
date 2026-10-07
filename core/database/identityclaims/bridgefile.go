package identityclaims

import (
	"bytes"
	"database/sql"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
)

// FileBridgesTx files every unfiled bridge whose ends touch subjectID, in
// this transaction. A bridge whose ends are both handles joins the association
// those ends already share, or a new one is minted and the bridge is claimed
// onto it. An unpromoted end is left for a later claim. A self-link is not
// filed and is not an error: the claim that triggered filing still stands.
// Role stays off the key (it is reconciled on the association). A relationship
// type, and any later non-role disambiguation such as place_relationship_type,
// is part of the key; a directed term keeps order and a symmetric one does not.
//
// Returned association ids are the handles whose cache the caller recomputes.
// Changes belong on the caller's revision. Does not commit.
func FileBridgesTx(tx *sql.Tx, subjectID []byte) (assocIDs [][]byte, changes []audit.Change, err error) {
	if len(subjectID) != 16 {
		return nil, nil, nil
	}
	bridges, err := bridgesTouching(tx, subjectID)
	if err != nil {
		return nil, nil, err
	}
	seen := map[string]bool{}
	for _, b := range bridges {
		id, ch, filed, err := fileBridge(tx, b)
		if err != nil {
			return nil, nil, err
		}
		if !filed {
			continue
		}
		changes = append(changes, ch...)
		if !seen[string(id)] {
			seen[string(id)] = true
			assocIDs = append(assocIDs, id)
		}
	}
	return assocIDs, changes, nil
}

type bridgeHead struct {
	id      []byte
	typeKey string
	typeID  []byte
}

func bridgesTouching(tx *sql.Tx, subjectID []byte) ([]bridgeHead, error) {
	keys := connectrules.BridgeTypeKeys()
	if len(keys) == 0 {
		return nil, nil
	}
	args := make([]any, 0, len(keys)+2)
	for _, k := range keys {
		args = append(args, k)
	}
	args = append(args, subjectID, subjectID)
	q := `SELECT s.id, st.key, s.subject_type_id
		FROM subjects s
		JOIN subject_types st ON st.id = s.subject_type_id AND st.origin = 'provenencia'
		WHERE st.key IN (` + database.SQLInPlaceholders(len(keys)) + `)
		  AND (
		    s.id = ?
		    OR EXISTS (
		      SELECT 1 FROM observations o
		      WHERE o.subject_id = s.id AND o.value_subject_id = ?
		    )
		  )`
	rows, err := tx.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bridgeHead
	for rows.Next() {
		var b bridgeHead
		if err := rows.Scan(&b.id, &b.typeKey, &b.typeID); err != nil {
			return nil, err
		}
		b.id = append([]byte(nil), b.id...)
		b.typeID = append([]byte(nil), b.typeID...)
		out = append(out, b)
	}
	return out, rows.Err()
}

type endpoint struct {
	propertyKey string
	entityID    []byte
}

// fileBridge claims one bridge onto its association. filed is false when the
// bridge waits (an end is unpromoted, or it is already filed) or when a
// self-link is refused.
func fileBridge(tx *sql.Tx, b bridgeHead) (assocID []byte, changes []audit.Change, filed bool, err error) {
	rule, ok := connectrules.LookupBridge(b.typeKey)
	if !ok || len(rule.Endpoints) == 0 {
		return nil, nil, false, nil
	}
	var exists int
	err = tx.QueryRow(`SELECT 1 FROM identity_claims
		WHERE subject_id = ? AND status = 'accepted'`, b.id).Scan(&exists)
	if err == nil {
		return nil, nil, false, nil
	}
	if err != sql.ErrNoRows {
		return nil, nil, false, err
	}

	obs, err := bridgeObservations(tx, b.id)
	if err != nil {
		return nil, nil, false, err
	}
	ends := make([]endpoint, 0, len(rule.Endpoints))
	seenEntity := map[string]bool{}
	for _, ep := range rule.Endpoints {
		entityID, ready, err := resolveEndpoint(tx, obs, ep.PropertyKey)
		if err != nil {
			return nil, nil, false, err
		}
		if !ready {
			return nil, nil, false, nil
		}
		ends = append(ends, endpoint{propertyKey: ep.PropertyKey, entityID: entityID})
		seenEntity[string(entityID)] = true
	}
	// Both ends on one handle: a duplicate on the graph, or a wrong match.
	if len(seenEntity) < len(ends) {
		return nil, nil, false, nil
	}

	var termID []byte
	directed := false
	var category string
	if keysOnType(rule) {
		termID = firstTerm(obs, rule.Disambiguation)
		if len(termID) != 16 {
			return nil, nil, false, nil
		}
		var bit int
		if err := tx.QueryRow(`SELECT directed, COALESCE(category, '') FROM property_terms WHERE id = ?`, termID).Scan(&bit, &category); err != nil {
			return nil, nil, false, err
		}
		directed = bit != 0
	}

	if b.typeKey == "place_relationship" && (category == "hierarchical" || category == "temporal") {
		fromID, toID := ends[0].entityID, ends[1].entityID
		if rule.Endpoints[0].PropertyKey != "from" {
			for _, e := range ends {
				if e.propertyKey == "from" {
					fromID = e.entityID
				}
				if e.propertyKey == "to" {
					toID = e.entityID
				}
			}
		}
		closes, err := placeRelationshipWouldCycle(tx, category, fromID, toID)
		if err != nil {
			return nil, nil, false, err
		}
		if closes {
			return nil, nil, false, nil
		}
	}

	assocID, err = findAssociation(tx, b.typeKey, ends, rule.Disambiguation, termID, !directed && sameEndpointType(rule))
	if err != nil {
		return nil, nil, false, err
	}
	if assocID == nil {
		minted, change, err := canonicalentities.InsertTx(tx, canonicalentities.CreateInput{
			SubjectTypeID: b.typeID,
		})
		if err != nil {
			return nil, nil, false, err
		}
		assocID = minted.ID
		changes = append(changes, change)
	}
	_, claimChange, err := InsertTx(tx, CreateInput{
		SubjectID: b.id,
		EntityID:  assocID,
		Status:    StatusAccepted,
	})
	if err != nil {
		return nil, nil, false, err
	}
	changes = append(changes, claimChange)
	return assocID, changes, true, nil
}

// keysOnType reports whether the bridge's disambiguation is the association's
// identity (relationship type, place relationship type). Role is a reconciled
// value, not part of the key.
func keysOnType(b connectrules.Bridge) bool {
	return connectrules.HasDisambiguation(b.Disambiguation) && b.Disambiguation != connectrules.DisambiguationRole
}

// placeRelationshipWouldCycle reports whether adding from→to of the given
// category would close a loop among already filed place relationships.
func placeRelationshipWouldCycle(tx *sql.Tx, category string, fromID, toID []byte) (bool, error) {
	rows, err := tx.Query(`
		SELECT ic_from.entity_id, ic_to.entity_id
		FROM identity_claims ic
		JOIN subjects s ON s.id = ic.subject_id
		JOIN subject_types st ON st.id = s.subject_type_id
			AND st.key = 'place_relationship' AND st.origin = 'provenencia'
		JOIN observations o_type ON o_type.subject_id = s.id
		JOIN properties p_type ON p_type.id = o_type.property_id
			AND p_type.key = 'place_relationship_type' AND p_type.origin = 'provenencia'
		JOIN property_terms pt ON pt.id = o_type.value_term_id AND pt.category = ?
		JOIN observations o_from ON o_from.subject_id = s.id
		JOIN properties p_from ON p_from.id = o_from.property_id
			AND p_from.key = 'from' AND p_from.origin = 'provenencia'
		JOIN observations o_to ON o_to.subject_id = s.id
		JOIN properties p_to ON p_to.id = o_to.property_id
			AND p_to.key = 'to' AND p_to.origin = 'provenencia'
		JOIN identity_claims ic_from ON ic_from.subject_id = o_from.value_subject_id
			AND ic_from.status = 'accepted'
		JOIN identity_claims ic_to ON ic_to.subject_id = o_to.value_subject_id
			AND ic_to.status = 'accepted'
		WHERE ic.status = 'accepted'`, category)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	adj := map[string][][]byte{}
	for rows.Next() {
		var fromEnt, toEnt []byte
		if err := rows.Scan(&fromEnt, &toEnt); err != nil {
			return false, err
		}
		adj[string(fromEnt)] = append(adj[string(fromEnt)], append([]byte(nil), toEnt...))
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	// Reach fromID starting at toID along existing edges?
	seen := map[string]bool{string(toID): true}
	queue := [][]byte{append([]byte(nil), toID...)}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if bytes.Equal(cur, fromID) {
			return true, nil
		}
		for _, next := range adj[string(cur)] {
			if seen[string(next)] {
				continue
			}
			seen[string(next)] = true
			queue = append(queue, next)
		}
	}
	return false, nil
}

func sameEndpointType(b connectrules.Bridge) bool {
	if len(b.Endpoints) != 2 {
		return false
	}
	return b.Endpoints[0].TypeKey == b.Endpoints[1].TypeKey
}

type obsValue struct {
	subjectID []byte
	termID    []byte
}

func bridgeObservations(tx *sql.Tx, bridgeID []byte) (map[string][]obsValue, error) {
	rows, err := tx.Query(`SELECT p.key, o.value_subject_id, o.value_term_id
		FROM observations o
		JOIN properties p ON p.id = o.property_id AND p.origin = 'provenencia'
		WHERE o.subject_id = ?`, bridgeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]obsValue{}
	for rows.Next() {
		var key string
		var v obsValue
		if err := rows.Scan(&key, &v.subjectID, &v.termID); err != nil {
			return nil, err
		}
		v.subjectID = append([]byte(nil), v.subjectID...)
		v.termID = append([]byte(nil), v.termID...)
		out[key] = append(out[key], v)
	}
	return out, rows.Err()
}

// resolveEndpoint maps one endpoint property to a single accepted handle.
// ready is false when the end is missing, unpromoted, or points at two handles.
func resolveEndpoint(tx *sql.Tx, obs map[string][]obsValue, propertyKey string) (entityID []byte, ready bool, err error) {
	var resolved [][]byte
	for _, v := range obs[propertyKey] {
		if len(v.subjectID) != 16 {
			continue
		}
		var id []byte
		err := tx.QueryRow(`SELECT entity_id FROM identity_claims
			WHERE subject_id = ? AND status = 'accepted'`, v.subjectID).Scan(&id)
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		if !containsID(resolved, id) {
			resolved = append(resolved, append([]byte(nil), id...))
		}
	}
	if len(resolved) != 1 {
		return nil, false, nil
	}
	return append([]byte(nil), resolved[0]...), true, nil
}

func firstTerm(obs map[string][]obsValue, propertyKey string) []byte {
	for _, v := range obs[propertyKey] {
		if len(v.termID) == 16 {
			return v.termID
		}
	}
	return nil
}

func containsID(ids [][]byte, id []byte) bool {
	for _, existing := range ids {
		if bytes.Equal(existing, id) {
			return true
		}
	}
	return false
}

func findAssociation(tx *sql.Tx, typeKey string, ends []endpoint, termKey string, termID []byte, symmetric bool) ([]byte, error) {
	id, err := queryAssociation(tx, typeKey, ends, termKey, termID)
	if err != nil || id != nil || !symmetric || len(ends) != 2 {
		return id, err
	}
	swapped := []endpoint{
		{propertyKey: ends[0].propertyKey, entityID: ends[1].entityID},
		{propertyKey: ends[1].propertyKey, entityID: ends[0].entityID},
	}
	return queryAssociation(tx, typeKey, swapped, termKey, termID)
}

func queryAssociation(tx *sql.Tx, typeKey string, ends []endpoint, termKey string, termID []byte) ([]byte, error) {
	q := `SELECT ic.entity_id
		FROM identity_claims ic
		JOIN subjects s ON s.id = ic.subject_id
		JOIN subject_types st ON st.id = s.subject_type_id
			AND st.key = ? AND st.origin = 'provenencia'
		WHERE ic.status = 'accepted'`
	args := []any{typeKey}
	for _, e := range ends {
		q += ` AND EXISTS (
			SELECT 1 FROM observations o
			JOIN properties p ON p.id = o.property_id AND p.key = ? AND p.origin = 'provenencia'
			JOIN identity_claims ec ON ec.subject_id = o.value_subject_id
				AND ec.status = 'accepted' AND ec.entity_id = ?
			WHERE o.subject_id = s.id)`
		args = append(args, e.propertyKey, e.entityID)
	}
	if len(termID) == 16 {
		q += ` AND EXISTS (
			SELECT 1 FROM observations o
			JOIN properties p ON p.id = o.property_id AND p.key = ? AND p.origin = 'provenencia'
			WHERE o.subject_id = s.id AND o.value_term_id = ?)`
		args = append(args, termKey, termID)
	}
	q += ` LIMIT 1`
	var id []byte
	err := tx.QueryRow(q, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}
