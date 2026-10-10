// Package autoreconciler stores the auto-reconciler's output (Spike 9 R3):
// auto_reconciler_values holds every auto-reconciled value for every Property
// on every canonical handle, and auto_reconciler_outcomes what the
// auto-reconciler did with each Observation. core/autoreconcile computes
// them; this package loads its inputs, writes its output, and keeps it
// current. A Reconciliation Claim is never written here.
//
// The tables are derived and rebuildable. Every write that can change a
// handle's values recomputes the affected handles inside its own transaction;
// EnsureCatalog rebuilds everything on open when CacheVersion moves. Rebuild
// and upkeep share one batched loader, and the rebuild-equals-upkeep tests
// (fixed-seed sequences plus named scenarios) hold them equal — an upkeep miss
// is otherwise silent.
//
// Upkeep: RecomputeTx / RecomputeSubjectsTx for Observation, Promote and
// Subject writes; RecomputeSourceTx for a Source credibility change;
// RecomputeCitationTx for a Citation certainty change. A claim create of any
// status recomputes the claim's handle (identityclaims.Create). Claim
// confidence and status have no edit path yet; when one lands it recomputes
// the claim's handle the same way.
//
// Recompute is per handle: all of a handle's Properties are rewritten together.
// Narrowing to (handle, Property) waits for timings that need it.
//
// Every auto-reconciled value is cached, displayed or not: `reason` is 'kept'
// for a displayed value, else why it isn't displayed (outvoted, weak, denied,
// provisional). Readers that count displayed values filter on it; search and
// matching read them all.
//
// Inputs: every Observation on an accepted or provisional member (rejected
// members don't count), with its Source, polarity, Source credibility,
// transcription certainty, and claim confidence and status; each term's key
// and each name's parts. Subject-valued Properties wait for S9-28. Date
// windows are stored as date_lo / date_hi.
package autoreconciler

import (
	"bytes"
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// CacheVersion is bumped whenever resolution changes what the table would
// hold; a stored version that differs rebuilds on open.
//
//	2: dashes and slashes separate words in name keys (S9-10).
//	5: one reconciler pipeline; text case-insensitive; majority; every value
//	   kept with a reason (S9-13). 3 and 4 were stamped by closed PRs' builds.
//	6: names reconcile by part type; names with no parts are no evidence (S9-13b).
//	7: one name per Person; values combine per part type; majority only
//	   outvotes spelling variants in names (S9-13b).
//	8: name parts compare as words: Smith-Jones = Smith + Jones (S9-13b).
//	9: evidence loaded (Sources, provenance, negatives, provisional members);
//	   each Observation's outcome cached (S9-14).
//	10: an outvoted outcome keeps the vote that beat it (S9-16).
//	11: dates reconcile by window; date_lo / date_hi and the date sort_key
//	   are stored (S9-21).
//	12: multiple cardinality keeps every distinct surviving value (S9-36).
//	13: subject-valued Properties map to the accepted handle (S9-28).
const CacheVersion = 13

// batchSize bounds the handles per loader batch (and so the IN list length).
const batchSize = 500

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

const (
	sqlDeleteFor         = `DELETE FROM auto_reconciler_values WHERE entity_id IN (`
	sqlDeleteOutcomesFor = `DELETE FROM auto_reconciler_outcomes WHERE entity_id IN (`

	sqlInsertOutcome = `INSERT INTO auto_reconciler_outcomes
		(entity_id, property_id, observation_id, reason, value_rank, denied_by,
		 vote_support, vote_total)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	sqlInsert = `INSERT INTO auto_reconciler_values
		(entity_id, property_id, rank, value_text, value_integer, value_term_id,
		 value_entity_id, value_date, value_name, date_lo, date_hi, sort_key, support, against, reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// One row per Observation on an accepted or provisional member of the
	// batch (rejected members don't count), with its evidence: polarity, the
	// Source it is cited under, and provenance as grade sort_order relative to
	// the provenencia default grade (standard, moderate); no assessment, no
	// grade, or no vocabulary reads as 0.
	sqlLoadCandidates = `SELECT ic.entity_id, o.id, o.property_id, p.value_type, p.cardinality,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_term_id,
			COALESCE(t.key, ''), o.value_subject_id,
			ic.status = 'provisional', o.polarity = 'negative', a.source_id,
			COALESCE(sg.sort_order - (SELECT sort_order FROM source_credibility_grades
				WHERE key = 'standard' AND origin = 'provenencia'), 0),
			c.transcription_uncertain,
			COALESCE(cg.sort_order - (SELECT sort_order FROM claim_confidence_grades
				WHERE key = 'moderate' AND origin = 'provenencia'), 0)
		FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		JOIN properties p ON p.id = o.property_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		LEFT JOIN source_credibility_assessments sca ON sca.source_id = a.source_id
		LEFT JOIN source_credibility_grades sg ON sg.id = sca.credibility_grade_id
		LEFT JOIN claim_confidence_grades cg ON cg.id = ic.confidence_grade_id
		WHERE ic.status IN ('accepted', 'provisional')
		  AND ic.entity_id IN (`

	sqlAcceptedEntities = `SELECT subject_id, entity_id FROM identity_claims
		WHERE status = 'accepted' AND subject_id IN (`

	// Handles whose members cite subjectID as a subject-valued end.
	sqlHandlesObservingSubject = `SELECT DISTINCT ic.entity_id
		FROM observations o
		JOIN identity_claims ic ON ic.subject_id = o.subject_id
			AND ic.status IN ('accepted', 'provisional')
		WHERE o.value_subject_id = ?`

	sqlHandlesForSubjects = `SELECT DISTINCT entity_id FROM identity_claims
		WHERE status IN ('accepted', 'provisional') AND subject_id IN (`

	// Handles with an accepted or provisional member that has an Observation
	// citing the Source / the Citation.
	sqlHandlesForSource = `SELECT DISTINCT ic.entity_id FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		WHERE ic.status IN ('accepted', 'provisional') AND a.source_id = ?`

	sqlHandlesForCitation = `SELECT DISTINCT ic.entity_id FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		WHERE ic.status IN ('accepted', 'provisional') AND o.citation_id = ?`

	sqlHandlesForProperty = `SELECT DISTINCT ic.entity_id FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		WHERE ic.status IN ('accepted', 'provisional') AND o.property_id = ?`

	sqlAllEntities = `SELECT id FROM canonical_entities ORDER BY id`
)

// RecomputeTx rewrites the cached rows of the given handles from truth tables.
// A handle with no members (or no cacheable values) ends with no rows. The
// handles' search documents, and the documents of handles whose headers read
// them, are reprojected in the same transaction. Deletes pass Released.Handles
// through the same call, so a removed member updates the rows that named it.
func RecomputeTx(q Querier, entityIDs [][]byte) error {
	ids := database.UniqueBlobIDs(entityIDs)
	for start := 0; start < len(ids); start += batchSize {
		end := min(start+batchSize, len(ids))
		if err := recomputeBatch(q, ids[start:end]); err != nil {
			return err
		}
	}
	deps, err := conclusionheaders.HeaderDependents(q, ids)
	if err != nil {
		return err
	}
	return searchindex.ReprojectHandles(q, append(ids, deps...))
}

// HandlesObservingSubject returns handles whose members' Observations point
// at subjectID. A claim create or a subject delete recomputes them so a
// subject-valued end tracks the handle it resolved to.
func HandlesObservingSubject(q Querier, subjectID []byte) ([][]byte, error) {
	if len(subjectID) != 16 {
		return nil, nil
	}
	return listIDs(q, sqlHandlesObservingSubject, subjectID)
}

// RecomputeTouchingTx recomputes entityIDs plus every handle whose members'
// Observations point at subjectID.
func RecomputeTouchingTx(q Querier, entityIDs [][]byte, subjectID []byte) error {
	extra, err := HandlesObservingSubject(q, subjectID)
	if err != nil {
		return err
	}
	return RecomputeTx(q, append(entityIDs, extra...))
}

// HandlesForSubjects is the handle list RecomputeSubjectsTx recomputes.
// Unpromoted Subjects are ignored.
func HandlesForSubjects(q Querier, subjectIDs [][]byte) ([][]byte, error) {
	ids := database.UniqueBlobIDs(subjectIDs)
	var handles [][]byte
	for start := 0; start < len(ids); start += batchSize {
		end := min(start+batchSize, len(ids))
		batch := ids[start:end]
		got, err := listIDs(q, sqlHandlesForSubjects+database.SQLInPlaceholders(len(batch))+`)`, database.BlobArgs(batch)...)
		if err != nil {
			return nil, err
		}
		handles = append(handles, got...)
	}
	return handles, nil
}

// RecomputeSubjectsTx recomputes the handles the given Subjects are accepted
// members of. Unpromoted Subjects are ignored.
func RecomputeSubjectsTx(q Querier, subjectIDs [][]byte) error {
	handles, err := HandlesForSubjects(q, subjectIDs)
	if err != nil {
		return err
	}
	return RecomputeTx(q, handles)
}

// HandlesForSource is the handle list RecomputeSourceTx recomputes.
func HandlesForSource(q Querier, sourceID []byte) ([][]byte, error) {
	return listIDs(q, sqlHandlesForSource, sourceID)
}

// RecomputeSourceTx recomputes the handles whose members have Observations
// citing the Source: its credibility is part of their evidence.
func RecomputeSourceTx(q Querier, sourceID []byte) error {
	ids, err := HandlesForSource(q, sourceID)
	if err != nil {
		return err
	}
	return RecomputeTx(q, ids)
}

// HandlesForCitation is the handle list RecomputeCitationTx recomputes.
func HandlesForCitation(q Querier, citationID []byte) ([][]byte, error) {
	return listIDs(q, sqlHandlesForCitation, citationID)
}

// RecomputeCitationTx recomputes the handles whose members have Observations
// on the Citation: its transcription certainty is part of their evidence.
func RecomputeCitationTx(q Querier, citationID []byte) error {
	ids, err := HandlesForCitation(q, citationID)
	if err != nil {
		return err
	}
	return RecomputeTx(q, ids)
}

// HandlesForProperty is the handle list RecomputePropertyTx recomputes.
func HandlesForProperty(q Querier, propertyID []byte) ([][]byte, error) {
	return listIDs(q, sqlHandlesForProperty, propertyID)
}

// RecomputePropertyTx recomputes every handle that has an Observation on the
// Property. A cardinality change resolves those handles differently.
func RecomputePropertyTx(q Querier, propertyID []byte) error {
	ids, err := HandlesForProperty(q, propertyID)
	if err != nil {
		return err
	}
	return RecomputeTx(q, ids)
}

// Rebuild clears the table, recomputes every handle, and stores CacheVersion.
func Rebuild(q Querier) error {
	if _, err := q.Exec(`DELETE FROM auto_reconciler_values`); err != nil {
		return err
	}
	if _, err := q.Exec(`DELETE FROM auto_reconciler_outcomes`); err != nil {
		return err
	}
	ids, err := listIDs(q, sqlAllEntities)
	if err != nil {
		return err
	}
	if err := RecomputeTx(q, ids); err != nil {
		return err
	}
	_, err = q.Exec(`UPDATE auto_reconciler_meta SET cache_version = ? WHERE id = 1`, CacheVersion)
	return err
}

// StoredVersion returns the cache version the catalog was last built at.
func StoredVersion(q Querier) (int, error) {
	var v int
	err := q.QueryRow(`SELECT cache_version FROM auto_reconciler_meta WHERE id = 1`).Scan(&v)
	return v, err
}

// NeedsRebuild reports whether the stored version differs from CacheVersion.
func NeedsRebuild(q Querier) (bool, error) {
	v, err := StoredVersion(q)
	if err != nil {
		return false, err
	}
	return v != CacheVersion, nil
}

// EnsureCatalog rebuilds in one transaction when the cache is stale. Call it
// on open before searchindex.EnsureCatalog, whose handle documents read the
// cache.
func EnsureCatalog(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	need, err := NeedsRebuild(db)
	if err != nil || !need {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := Rebuild(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func recomputeBatch(q Querier, ids [][]byte) error {
	in := database.SQLInPlaceholders(len(ids)) + `)`
	if _, err := q.Exec(sqlDeleteFor+in, database.BlobArgs(ids)...); err != nil {
		return err
	}
	if _, err := q.Exec(sqlDeleteOutcomesFor+in, database.BlobArgs(ids)...); err != nil {
		return err
	}
	groups, err := load(q, ids)
	if err != nil {
		return err
	}
	for _, g := range groups {
		res, err := autoreconcile.Reconcile(g.valueType, g.candidates, nil, g.cardinality)
		if err != nil {
			return err
		}
		for i, cl := range res.Values {
			if err := insertRow(q, g, i+1, cl); err != nil {
				return err
			}
		}
		for _, o := range res.Outcomes {
			var rank any
			if o.Value >= 0 {
				rank = o.Value + 1
			}
			var deniedBy any
			if len(o.DeniedBy) > 0 {
				deniedBy = o.DeniedBy
			}
			var voteSupport, voteTotal any
			if o.Vote != (autoreconcile.Vote{}) {
				voteSupport, voteTotal = o.Vote.Support, o.Vote.Of
			}
			if _, err := q.Exec(sqlInsertOutcome, g.entityID, g.propertyID, o.ObservationID, string(o.Reason), rank, deniedBy,
				voteSupport, voteTotal); err != nil {
				return err
			}
		}
	}
	return nil
}

// group is one (handle, Property) and its candidates.
type group struct {
	entityID    []byte
	propertyID  []byte
	valueType   string
	cardinality string
	candidates  []autoreconcile.Candidate
}

// load reads every candidate for the batch in a fixed number of queries: the
// candidates, then their date values, name values, and name parts. Groups
// come back ordered by (entity id, property id).
func load(q Querier, ids [][]byte) ([]*group, error) {
	rows, err := q.Query(sqlLoadCandidates+database.SQLInPlaceholders(len(ids))+`)`, database.BlobArgs(ids)...)
	if err != nil {
		return nil, err
	}
	type pending struct {
		g      *group
		c      autoreconcile.Candidate
		dateID []byte
		nameID []byte
	}
	byKey := map[string]*group{}
	var groups []*group
	var all []pending
	var dateIDs, nameIDs [][]byte
	for rows.Next() {
		var (
			entityID, obsID, propertyID       []byte
			valueType, cardinality            string
			text                              sql.NullString
			integer                           sql.NullInt64
			dateID, nameID, termID, subjectID []byte
			termKey                           string
			provisional, negative             bool
			sourceID                          []byte
			credibility, confidence           int
			uncertain                         bool
		)
		if err := rows.Scan(&entityID, &obsID, &propertyID, &valueType, &cardinality, &text, &integer, &dateID, &nameID, &termID, &termKey, &subjectID,
			&provisional, &negative, &sourceID, &credibility, &uncertain, &confidence); err != nil {
			_ = rows.Close()
			return nil, err
		}
		k := string(entityID) + "\x00" + string(propertyID)
		g, ok := byKey[k]
		if !ok {
			g = &group{entityID: entityID, propertyID: propertyID, valueType: valueType, cardinality: cardinality}
			byKey[k] = g
			groups = append(groups, g)
		}
		c := autoreconcile.Candidate{ObservationID: obsID, Value: autoreconcile.Value{
			Text: text.String, HasText: text.Valid,
			Integer: integer.Int64, HasInteger: integer.Valid,
			TermID: termID, TermKey: termKey, SubjectID: subjectID,
		}, SourceID: sourceID, Negative: negative, Provisional: provisional,
			Provenance: autoreconcile.Provenance{Credibility: credibility, Uncertain: uncertain, ClaimConfidence: confidence}}
		all = append(all, pending{g: g, c: c, dateID: dateID, nameID: nameID})
		if len(dateID) > 0 {
			dateIDs = append(dateIDs, dateID)
		}
		if len(nameID) > 0 {
			nameIDs = append(nameIDs, nameID)
		}
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}

	var subjectIDs [][]byte
	for i := range all {
		if len(all[i].c.Value.SubjectID) > 0 {
			subjectIDs = append(subjectIDs, all[i].c.Value.SubjectID)
		}
	}
	entities, err := acceptedEntities(q, subjectIDs)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if eid, ok := entities[string(all[i].c.Value.SubjectID)]; ok {
			all[i].c.Value.EntityID = eid
		}
	}

	dates, err := datevalues.LookupManyTx(q, dateIDs)
	if err != nil {
		return nil, err
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return nil, err
	}
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.c.Value.Date = &d
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.c.Value.Name = &n
		}
		// An Observation missing its type's value has nothing to autoreconcile.
		if !hasValue(p.g.valueType, p.c.Value) {
			continue
		}
		p.g.candidates = append(p.g.candidates, p.c)
	}

	sort.Slice(groups, func(i, j int) bool {
		if c := bytes.Compare(groups[i].entityID, groups[j].entityID); c != 0 {
			return c < 0
		}
		return bytes.Compare(groups[i].propertyID, groups[j].propertyID) < 0
	})
	return groups, nil
}

func hasValue(valueType string, v autoreconcile.Value) bool {
	switch valueType {
	case properties.ValueTypeText:
		return v.HasText
	case properties.ValueTypeInteger:
		return v.HasInteger
	case properties.ValueTypeTerm:
		return len(v.TermID) > 0
	case properties.ValueTypeDate:
		return v.Date != nil
	case properties.ValueTypeName:
		return v.Name != nil
	case properties.ValueTypeSubject:
		return len(v.SubjectID) > 0
	}
	return false
}

func acceptedEntities(q Querier, ids [][]byte) (map[string][]byte, error) {
	out := map[string][]byte{}
	ids = database.UniqueBlobIDs(ids)
	for start := 0; start < len(ids); start += batchSize {
		end := min(start+batchSize, len(ids))
		batch := ids[start:end]
		rows, err := q.Query(sqlAcceptedEntities+database.SQLInPlaceholders(len(batch))+`)`, database.BlobArgs(batch)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var subjectID, entityID []byte
			if err := rows.Scan(&subjectID, &entityID); err != nil {
				_ = rows.Close()
				return nil, err
			}
			out[string(subjectID)] = append([]byte(nil), entityID...)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return nil, err
		}
		_ = rows.Close()
	}
	return out, nil
}

func insertRow(q Querier, g *group, rank int, cl autoreconcile.ReconciledValue) error {
	v := cl.Value
	var (
		text, sortKey     any
		integer           any
		termID            any
		valueEntity       any
		dateBlob, nameBlb any
		dateLo, dateHi    any
	)
	if v.HasText {
		text = v.Text
	}
	if v.HasInteger {
		integer = v.Integer
	}
	if len(v.TermID) > 0 {
		termID = v.TermID
	}
	if len(v.EntityID) > 0 {
		valueEntity = v.EntityID
	}
	if v.Date != nil {
		b, err := valuecodec.MarshalDate(*v.Date)
		if err != nil {
			return err
		}
		dateBlob = b
		if lo, hi, ok := autoreconcile.DateBounds(*v.Date); ok {
			if lo != nil {
				dateLo = *lo
			}
			if hi != nil {
				dateHi = *hi
			}
		}
	}
	if v.Name != nil {
		b, err := valuecodec.MarshalName(*v.Name)
		if err != nil {
			return err
		}
		nameBlb = b
	}
	if k, ok := autoreconcile.SortKey(g.valueType, v); ok {
		sortKey = k
	}
	_, err := q.Exec(sqlInsert, g.entityID, g.propertyID, rank, text, integer, termID, valueEntity, dateBlob, nameBlb, dateLo, dateHi, sortKey,
		cl.Support, cl.Against, string(cl.Reason))
	return err
}

func listIDs(q Querier, query string, args ...any) ([][]byte, error) {
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
		out = append(out, id)
	}
	return out, rows.Err()
}
