package identityclaims

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// FileBridgesTx files every unfiled bridge whose ends touch subjectID, in
// this transaction. A bridge whose ends are both handles joins the association
// those ends already share, or a new one is minted and the bridge is claimed
// onto it. An unpromoted end is left for a later claim. A self-link is not
// filed and is not an error: the write that triggered filing still stands.
// Role stays off the key (it is reconciled on the association). A relationship
// type, and any later non-role disambiguation such as place_relationship_type,
// is part of the key; a directed term keeps order and a symmetric one does not.
//
// The returned ids are the handles filing touched. The caller discards them:
// writes.Run recomputes every handle the changes name. Changes belong on the
// caller's revision. Does not commit. subjectID is a new member (the bridges
// touching it file) or a new bridge (it files itself).
func FileBridgesTx(tx *sql.Tx, subjectID []byte) (assocIDs [][]byte, changes []rowchange.Change, err error) {
	if len(subjectID) != 16 {
		return nil, nil, nil
	}
	bridges, err := bridgeHeads(tx, fileCond, subjectID, subjectID)
	if err != nil {
		return nil, nil, err
	}
	return fileHeads(tx, bridges)
}

// FileEvidenceGraphBridgesTx files every unfiled bridge on the Evidence graph
// whose ends are both handles, including a bridge whose ends were claimed
// before this call. Does not commit.
func FileEvidenceGraphBridgesTx(tx *sql.Tx, graphID []byte) (assocIDs [][]byte, changes []rowchange.Change, err error) {
	if len(graphID) != 16 {
		return nil, nil, nil
	}
	heads, err := bridgeHeads(tx, graphCond, graphID)
	if err != nil {
		return nil, nil, err
	}
	return fileHeads(tx, heads)
}

func fileHeads(tx *sql.Tx, bridges []bridgeHead) (assocIDs [][]byte, changes []rowchange.Change, err error) {
	seen := map[string]bool{}
	pass := &filePass{tx: tx, links: map[string]*connectrules.Links{}, terms: map[string]*bridgeTerm{}}
	for _, b := range bridges {
		id, ch, filed, err := pass.fileBridge(b)
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
	// claimID is the head's accepted claim and entityID the handle that claim
	// is a member of; both nil when it has none.
	claimID, entityID []byte
}

// touchesEnd is true for a bridge Subject s (alias) that has subject ? as an
// end: one of its Observations names that Subject as its value. It takes one
// argument and applies no polarity filter, as filing reads none. The IN
// subquery lets SQLite seek the Observations by value and then s by id.
const touchesEnd = `s.id IN (SELECT o.subject_id FROM observations o WHERE o.value_subject_id = ?)`

// fileCond is subject ? itself or a bridge touching it (two arguments, the
// same id twice). The UNION ALL keeps the seek on s.id.
const fileCond = `s.id IN (SELECT ? UNION ALL SELECT o.subject_id FROM observations o WHERE o.value_subject_id = ?)`

// graphCond is a bridge on Evidence graph ? (one argument).
const graphCond = `s.graph_id = ?`

// FiledBridge is a bridge Subject with an accepted claim: its Subject id, that
// claim's id and the handle the claim is a member of.
type FiledBridge struct {
	SubjectID, ClaimID, EntityID []byte
}

// FiledBridgesOf lists the filed bridges that have subjectID as an end, in
// bridge Subject id order. It is the one read of what an Unpromote takes back:
// UnfileBridges releases these claims and the Unpromote preflight reports them.
// A bridge that waits (no accepted claim) is not listed.
func FiledBridgesOf(tx *sql.Tx, subjectID []byte) ([]FiledBridge, error) {
	if len(subjectID) != 16 {
		return nil, nil
	}
	heads, err := bridgeHeads(tx, touchesEnd, subjectID)
	if err != nil {
		return nil, err
	}
	var out []FiledBridge
	for _, head := range heads {
		if head.claimID == nil {
			continue
		}
		out = append(out, FiledBridge{SubjectID: head.id, ClaimID: head.claimID, EntityID: head.entityID})
	}
	return out, nil
}

// UnpromoteBridgeHandlesDeleted lists association handles UnfileBridges would
// delete: those whose every accepted member is among the filed bridges, in the
// order a handle first appears in filed. UnfileBridges releases all of them, so
// a handle several of them share goes when the last is released.
func UnpromoteBridgeHandlesDeleted(tx *sql.Tx, filed []FiledBridge) ([]deleteimpact.Listed, error) {
	var order [][]byte
	leaving := map[string][][]byte{}
	for _, f := range filed {
		key := string(f.EntityID)
		if _, ok := leaving[key]; !ok {
			order = append(order, f.EntityID)
		}
		leaving[key] = append(leaving[key], f.SubjectID)
	}
	var out []deleteimpact.Listed
	for _, entityID := range order {
		goes, err := deleteimpact.HandleGoesWithout(tx, entityID, leaving[string(entityID)])
		if err != nil {
			return nil, err
		}
		if !goes {
			continue
		}
		var ref string
		if err := tx.QueryRow(`SELECT ref FROM canonical_entities WHERE id = ?`, entityID).Scan(&ref); err != nil {
			return nil, err
		}
		out = append(out, deleteimpact.Listed{ID: append([]byte(nil), entityID...), Ref: ref})
	}
	return out, nil
}

// UnfileBridges is the inverse of FileBridgesTx: it deletes, audited, the
// accepted claim of every bridge that has subjectID as an end, in
// bridge Subject id order. A bridge is on its ends' Evidence graph, so this is
// per graph. A bridge handle whose last member goes is deleted in releaseClaim.
// Does not commit.
func UnfileBridges(tx *database.Tx, subjectID []byte) ([]rowchange.Change, error) {
	filed, err := FiledBridgesOf(tx.Tx, subjectID)
	if err != nil {
		return nil, err
	}
	var changes []rowchange.Change
	for _, f := range filed {
		released, err := deleteimpact.ReleaseClaim(tx, f.ClaimID)
		if err != nil {
			return nil, err
		}
		changes = append(changes, released...)
	}
	return changes, nil
}

// bridgeHeads lists the bridge Subjects (product-seeded bridge types) that
// satisfy cond, in Subject id order. cond is a SQL condition over the alias s.
func bridgeHeads(tx *sql.Tx, cond string, condArgs ...any) ([]bridgeHead, error) {
	keys := connectrules.BridgeTypeKeys()
	if len(keys) == 0 {
		return nil, nil
	}
	args := make([]any, 0, len(keys)+len(condArgs))
	for _, k := range keys {
		args = append(args, k)
	}
	args = append(args, condArgs...)
	rows, err := tx.Query(bridgeHeadsSQL(cond, len(keys)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []bridgeHead
	for rows.Next() {
		var b bridgeHead
		if err := rows.Scan(&b.id, &b.typeKey, &b.typeID, &b.claimID, &b.entityID); err != nil {
			return nil, err
		}
		b.id = append([]byte(nil), b.id...)
		b.typeID = append([]byte(nil), b.typeID...)
		b.claimID = append([]byte(nil), b.claimID...)
		b.entityID = append([]byte(nil), b.entityID...)
		out = append(out, b)
	}
	return out, rows.Err()
}

type endpoint struct {
	propertyKey string
	entityID    []byte
}

// unpromoted is the wait code for a value Subject with no accepted claim.
const unpromoted = "unpromoted"

// readCount, when set, is called once for each read of filed links or term
// rows. Tests only.
var readCount func()

// SetReadCounterForTest counts the filed-link and term reads a filing pass
// makes. fn nil stops counting. Tests only.
func SetReadCounterForTest(fn func()) { readCount = fn }

func countRead() {
	if readCount != nil {
		readCount()
	}
}

// filePass is the reads one filing pass shares: each type and term's filed
// links, read once on first need and added to as the pass files, and each
// term row by id.
type filePass struct {
	tx    *sql.Tx
	links map[string]*connectrules.Links
	// terms holds a term row by id; nil for an id with no row.
	terms map[string]*bridgeTerm
}

// fileBridge claims one bridge onto its association. filed is false when the
// bridge waits (an end is unpromoted, or it is already filed) or when
// connectrules.Decide refuses it (a self-link, a loop, an incomplete bridge).
func (p *filePass) fileBridge(b bridgeHead) (assocID []byte, changes []rowchange.Change, filed bool, err error) {
	tx := p.tx
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
	values := make([][]connectrules.Value, len(rule.Endpoints))
	for i, ep := range rule.Endpoints {
		values[i], err = endpointValues(tx, obs[ep.PropertyKey])
		if err != nil {
			return nil, nil, false, err
		}
	}
	terms, err := p.bridgeTerms(obs, rule.Disambiguation)
	if err != nil {
		return nil, nil, false, err
	}
	keys := make([]vocabkey.BridgeTerm, 0, len(terms))
	for _, t := range terms {
		keys = append(keys, vocabkey.BridgeTerm(t.key))
	}
	term := connectrules.Term(keys)

	linkKey := b.typeKey + "|" + string(term)
	res, err := connectrules.Decide(rule, keys, values, func() (*connectrules.Links, error) {
		if p.links[linkKey] == nil {
			filed, err := filedLinks(tx, rule, string(term))
			if err != nil {
				return nil, err
			}
			p.links[linkKey] = connectrules.NewLinks(filed)
		}
		return p.links[linkKey], nil
	})
	if err != nil {
		return nil, nil, false, err
	}
	if res.Decision != connectrules.DecisionFiles {
		return nil, nil, false, nil
	}
	if res.Link != nil {
		// A later bridge of this pass is checked against the link filed here.
		p.links[linkKey].Add(*res.Link)
	}

	ends := make([]endpoint, 0, len(rule.Endpoints))
	for i, ep := range rule.Endpoints {
		ends = append(ends, endpoint{propertyKey: ep.PropertyKey, entityID: []byte(res.Records[i])})
	}
	var termID []byte
	directed := false
	if rule.KeysOnType() {
		for _, t := range terms {
			if vocabkey.BridgeTerm(t.key) == term {
				termID, directed = t.id, t.directed
				break
			}
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

// filedLinks reads every filed link of the rule's bridge type and term as
// from→to entity edges, for the loop check. The first endpoint is the from end.
func filedLinks(tx *sql.Tx, rule connectrules.Bridge, termKey string) ([]connectrules.Link, error) {
	countRead()
	rows, err := tx.Query(`
		SELECT ic_from.entity_id, ic_to.entity_id
		FROM identity_claims ic
		JOIN subjects s ON s.id = ic.subject_id
		JOIN subject_types st ON st.id = s.subject_type_id
			AND st.key = ? AND st.origin = 'provenencia'
		JOIN observations o_type ON o_type.subject_id = s.id
		JOIN properties p_type ON p_type.id = o_type.property_id
			AND p_type.key = ? AND p_type.origin = 'provenencia'
		JOIN property_terms pt ON pt.id = o_type.value_term_id AND pt.key = ?
		JOIN observations o_from ON o_from.subject_id = s.id
		JOIN properties p_from ON p_from.id = o_from.property_id
			AND p_from.key = ? AND p_from.origin = 'provenencia'
		JOIN observations o_to ON o_to.subject_id = s.id
		JOIN properties p_to ON p_to.id = o_to.property_id
			AND p_to.key = ? AND p_to.origin = 'provenencia'
		JOIN identity_claims ic_from ON ic_from.subject_id = o_from.value_subject_id
			AND ic_from.status = 'accepted'
		JOIN identity_claims ic_to ON ic_to.subject_id = o_to.value_subject_id
			AND ic_to.status = 'accepted'
		WHERE ic.status = 'accepted'`,
		rule.BridgeTypeKey, rule.Disambiguation, termKey,
		rule.Endpoints[0].PropertyKey, rule.Endpoints[1].PropertyKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []connectrules.Link
	for rows.Next() {
		var from, to []byte
		if err := rows.Scan(&from, &to); err != nil {
			return nil, err
		}
		out = append(out, connectrules.Link{From: string(from), To: string(to)})
	}
	return out, rows.Err()
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

// endpointValues maps the Subjects one endpoint property names to their
// records: the entity of the accepted claim, or the unpromoted wait. A value
// that is not a 16-byte id is skipped.
func endpointValues(tx *sql.Tx, vals []obsValue) ([]connectrules.Value, error) {
	var out []connectrules.Value
	for _, v := range vals {
		if len(v.subjectID) != 16 {
			continue
		}
		var id []byte
		err := tx.QueryRow(`SELECT entity_id FROM identity_claims
			WHERE subject_id = ? AND status = 'accepted'`, v.subjectID).Scan(&id)
		if err == sql.ErrNoRows {
			out = append(out, connectrules.Value{Wait: unpromoted})
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, connectrules.Value{Record: string(id)})
	}
	return out, nil
}

type bridgeTerm struct {
	id       []byte
	key      string
	directed bool
}

// bridgeTerms reads every disambiguation term the bridge records, any
// polarity, with its key and directed bit. A term id with no property_terms
// row is skipped.
func (p *filePass) bridgeTerms(obs map[string][]obsValue, propertyKey string) ([]bridgeTerm, error) {
	if !connectrules.HasDisambiguation(propertyKey) {
		return nil, nil
	}
	var out []bridgeTerm
	for _, v := range obs[propertyKey] {
		if len(v.termID) != 16 {
			continue
		}
		t, err := p.term(v.termID)
		if err != nil {
			return nil, err
		}
		if t == nil {
			continue
		}
		out = append(out, *t)
	}
	return out, nil
}

// term reads one property_terms row once per pass; nil when the id has none.
func (p *filePass) term(id []byte) (*bridgeTerm, error) {
	if t, ok := p.terms[string(id)]; ok {
		return t, nil
	}
	countRead()
	t := bridgeTerm{id: id}
	var bit int
	err := p.tx.QueryRow(`SELECT directed, key FROM property_terms WHERE id = ?`, id).Scan(&bit, &t.key)
	if err == sql.ErrNoRows {
		p.terms[string(id)] = nil
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.directed = bit != 0
	p.terms[string(id)] = &t
	return &t, nil
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
	q, args := associationSQL(typeKey, ends, termKey, termID)
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

// associationSQL is the queryAssociation query and its arguments. It seeks
// from ends[0], the from end of every bridge (the event of a location, the
// child of a part_of), through the Observations that name its Subjects, to
// the bridge Subject. CROSS JOIN pins that order: left to choose, SQLite
// starts at the subject type index and tests every bridge of the type, and a
// place can be the other end of thousands of links. The other ends and the
// term stay EXISTS tests on the bridge's own observations; the unary plus
// keeps those on observations_subject_id_idx.
func associationSQL(typeKey string, ends []endpoint, termKey string, termID []byte) (string, []any) {
	q := `SELECT ic.entity_id
		FROM identity_claims ec0
		CROSS JOIN observations o0 ON o0.value_subject_id = ec0.subject_id
		CROSS JOIN properties p0 ON p0.id = o0.property_id AND p0.key = ? AND p0.origin = 'provenencia'
		CROSS JOIN subjects s ON s.id = o0.subject_id
		CROSS JOIN subject_types st ON st.id = s.subject_type_id
			AND st.key = ? AND st.origin = 'provenencia'
		CROSS JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		WHERE ec0.entity_id = ? AND ec0.status = 'accepted'`
	args := []any{ends[0].propertyKey, typeKey, ends[0].entityID}
	for _, e := range ends[1:] {
		q += ` AND EXISTS (
			SELECT 1 FROM observations o
			JOIN properties p ON p.id = +o.property_id AND p.key = ? AND p.origin = 'provenencia'
			JOIN identity_claims ec ON ec.subject_id = o.value_subject_id
				AND ec.status = 'accepted' AND ec.entity_id = ?
			WHERE o.subject_id = s.id)`
		args = append(args, e.propertyKey, e.entityID)
	}
	if len(termID) == 16 {
		q += ` AND EXISTS (
			SELECT 1 FROM observations o
			JOIN properties p ON p.id = +o.property_id AND p.key = ? AND p.origin = 'provenencia'
			WHERE o.subject_id = s.id AND +o.value_term_id = ?)`
		args = append(args, termKey, termID)
	}
	q += ` LIMIT 1`
	return q, args
}

// bridgeHeadsSQL is the bridgeHeads query for nKeys bridge type keys. CROSS
// JOIN keeps SQLite starting at s through cond's selective seek; left to
// order the join, it starts at the type index and reads every bridge Subject
// in the catalog (see graphcache/load.go linkLoadSQL).
func bridgeHeadsSQL(cond string, nKeys int) string {
	return `SELECT s.id, st.key, s.subject_type_id, ic.id, ic.entity_id
		FROM subjects s
		CROSS JOIN subject_types st ON st.id = s.subject_type_id AND st.origin = 'provenencia'
		LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		WHERE st.key IN (` + database.SQLInPlaceholders(nKeys) + `)
		  AND (` + cond + `)
		ORDER BY s.id`
}
