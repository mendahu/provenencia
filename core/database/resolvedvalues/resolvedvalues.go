// Package resolvedvalues maintains conclusion_resolved_values: the resolver's
// ranked clusters for every Property on every canonical handle (Spike 9 R3).
//
// The table is derived and rebuildable. Every write that can change a handle's
// values recomputes the affected handles inside its own transaction
// (RecomputeTx / RecomputeSubjectsTx); EnsureCatalog rebuilds everything on
// open when CacheVersion moves. Rebuild and upkeep share one batched loader,
// and a randomized test holds them equal — an upkeep miss is otherwise silent.
//
// Recompute is per handle: all of a handle's Properties are rewritten together.
// Narrowing to (handle, Property) waits for timings that need it.
//
// v1 scope: positive Observations on accepted members, clustered by
// core/resolve. Subject-valued Properties are not cached until S9-28 maps them
// to handles; provenance arrives with S9-14; date_lo / date_hi with S9-21.
package resolvedvalues

import (
	"bytes"
	"database/sql"
	"sort"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/resolve"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// CacheVersion is bumped whenever resolution changes what the table would
// hold; a stored version that differs rebuilds on open.
const CacheVersion = 1

// batchSize bounds the handles per loader batch (and so the IN list length).
const batchSize = 500

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

const (
	sqlDeleteFor = `DELETE FROM conclusion_resolved_values WHERE entity_id IN (`

	sqlInsert = `INSERT INTO conclusion_resolved_values
		(entity_id, property_id, rank, value_text, value_integer, value_term_id,
		 value_date, value_name, sort_key, support)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// One row per positive Observation on an accepted member of the batch.
	sqlLoadCandidates = `SELECT ic.entity_id, o.id, o.property_id, p.value_type,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_term_id
		FROM identity_claims ic
		JOIN observations o ON o.subject_id = ic.subject_id
		JOIN properties p ON p.id = o.property_id
		WHERE ic.status = 'accepted'
		  AND o.polarity = 'positive'
		  AND p.value_type <> 'subject'
		  AND ic.entity_id IN (`

	sqlHandlesForSubjects = `SELECT DISTINCT entity_id FROM identity_claims
		WHERE status = 'accepted' AND subject_id IN (`

	sqlAllEntities = `SELECT id FROM canonical_entities ORDER BY id`
)

// RecomputeTx rewrites the cached rows of the given handles from truth tables.
// A handle with no members (or no cacheable values) ends with no rows.
func RecomputeTx(q Querier, entityIDs [][]byte) error {
	ids := database.UniqueBlobIDs(entityIDs)
	for start := 0; start < len(ids); start += batchSize {
		end := min(start+batchSize, len(ids))
		if err := recomputeBatch(q, ids[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// RecomputeSubjectsTx recomputes the handles the given Subjects are accepted
// members of. Unpromoted Subjects are ignored.
func RecomputeSubjectsTx(q Querier, subjectIDs [][]byte) error {
	ids := database.UniqueBlobIDs(subjectIDs)
	var handles [][]byte
	for start := 0; start < len(ids); start += batchSize {
		end := min(start+batchSize, len(ids))
		batch := ids[start:end]
		got, err := listIDs(q, sqlHandlesForSubjects+database.SQLInPlaceholders(len(batch))+`)`, database.BlobArgs(batch)...)
		if err != nil {
			return err
		}
		handles = append(handles, got...)
	}
	return RecomputeTx(q, handles)
}

// Rebuild clears the table, recomputes every handle, and stores CacheVersion.
func Rebuild(q Querier) error {
	if _, err := q.Exec(`DELETE FROM conclusion_resolved_values`); err != nil {
		return err
	}
	ids, err := listIDs(q, sqlAllEntities)
	if err != nil {
		return err
	}
	if err := RecomputeTx(q, ids); err != nil {
		return err
	}
	_, err = q.Exec(`UPDATE conclusion_resolved_meta SET cache_version = ? WHERE id = 1`, CacheVersion)
	return err
}

// StoredVersion returns the cache version the catalog was last built at.
func StoredVersion(q Querier) (int, error) {
	var v int
	err := q.QueryRow(`SELECT cache_version FROM conclusion_resolved_meta WHERE id = 1`).Scan(&v)
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
// on open, beside searchindex.EnsureCatalog.
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
	if _, err := q.Exec(sqlDeleteFor+database.SQLInPlaceholders(len(ids))+`)`, database.BlobArgs(ids)...); err != nil {
		return err
	}
	groups, err := load(q, ids)
	if err != nil {
		return err
	}
	for _, g := range groups {
		res, err := resolve.Resolve(g.valueType, g.candidates, nil)
		if err != nil {
			return err
		}
		for i, cl := range res.Clusters {
			if err := insertRow(q, g, i+1, cl); err != nil {
				return err
			}
		}
	}
	return nil
}

// group is one (handle, Property) and its candidates.
type group struct {
	entityID   []byte
	propertyID []byte
	valueType  string
	candidates []resolve.Candidate
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
		c      resolve.Candidate
		dateID []byte
		nameID []byte
	}
	byKey := map[string]*group{}
	var groups []*group
	var all []pending
	var dateIDs, nameIDs [][]byte
	for rows.Next() {
		var (
			entityID, obsID, propertyID []byte
			valueType                   string
			text                        sql.NullString
			integer                     sql.NullInt64
			dateID, nameID, termID      []byte
		)
		if err := rows.Scan(&entityID, &obsID, &propertyID, &valueType, &text, &integer, &dateID, &nameID, &termID); err != nil {
			_ = rows.Close()
			return nil, err
		}
		k := string(entityID) + "\x00" + string(propertyID)
		g, ok := byKey[k]
		if !ok {
			g = &group{entityID: entityID, propertyID: propertyID, valueType: valueType}
			byKey[k] = g
			groups = append(groups, g)
		}
		c := resolve.Candidate{ObservationID: obsID, Value: resolve.Value{
			Text: text.String, HasText: text.Valid,
			Integer: integer.Int64, HasInteger: integer.Valid,
			TermID: termID,
		}}
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
		// An Observation missing its type's value has nothing to resolve.
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

func hasValue(valueType string, v resolve.Value) bool {
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
	}
	return false
}

func insertRow(q Querier, g *group, rank int, cl resolve.Cluster) error {
	v := cl.Value
	var (
		text, sortKey     any
		integer           any
		termID            any
		dateBlob, nameBlb any
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
	if v.Date != nil {
		b, err := valuecodec.MarshalDate(*v.Date)
		if err != nil {
			return err
		}
		dateBlob = b
	}
	if v.Name != nil {
		b, err := valuecodec.MarshalName(*v.Name)
		if err != nil {
			return err
		}
		nameBlb = b
	}
	if k, ok := resolve.SortKey(g.valueType, v); ok {
		sortKey = k
	}
	_, err := q.Exec(sqlInsert, g.entityID, g.propertyID, rank, text, integer, termID, dateBlob, nameBlb, sortKey, cl.Support)
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
