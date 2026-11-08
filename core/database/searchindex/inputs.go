package searchindex

import (
	"bytes"
	"database/sql"
	"slices"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
)

// catalog_search_inputs records what each person, event and place summary was
// composed from, and the handles at each association's ends, so a write can
// find the rows it moved. This package writes it with the documents, in the
// same transaction, and clears and rebuilds it with them.
const (
	sqlDeleteInputs     = `DELETE FROM catalog_search_inputs WHERE dependent_id IN (`
	sqlInsertInputsHead = `INSERT OR IGNORE INTO catalog_search_inputs (dependent_id, input_id, bridge, endpoint, dependent_is_association) VALUES `
	// inputRowsPerInsert keeps one statement under SQLite's variable limit.
	inputRowsPerInsert = 160

	// dependentsOf reads the exact keys; oldEnds and dependentsOfHandles read by
	// one column through sqlbatch.
	sqlDependentsOfKeys = `SELECT DISTINCT dependent_id FROM catalog_search_inputs
		WHERE (input_id, bridge, endpoint) IN (VALUES `
	sqlDependentsOfHandles = `SELECT DISTINCT dependent_id FROM catalog_search_inputs WHERE input_id IN (`
	sqlOldEnds             = `SELECT dependent_id, input_id, bridge, endpoint FROM catalog_search_inputs
		WHERE dependent_is_association = 1 AND dependent_id IN (`
	// keysPerLookup keeps one dependentsOf statement at 750 binds.
	keysPerLookup = 250
)

// writeInputs replaces the inputs of every id in the batch as a dependent: the
// composed keys of a person, event or place that is not merged, and the ends
// of an association handle. A gone or merged id keeps none.
func writeInputs(q Querier, g *graphcache.Graph, ids [][]byte, heads map[string]handleHead, s conclusionsummaries.Summaries) error {
	if _, err := q.Exec(sqlDeleteInputs+database.SQLInPlaceholders(len(ids))+`)`, database.BlobArgs(ids)...); err != nil {
		return err
	}
	var rows []inputRow
	add := func(dep []byte, keys []graphcache.InputKey, association bool) {
		for _, k := range keys {
			rows = append(rows, inputRow{dep: dep, key: k, association: association})
		}
	}
	var assocs [][]byte
	for _, id := range ids {
		h, ok := heads[string(id)]
		if !ok || h.merged {
			continue
		}
		if h.kind == "" {
			assocs = append(assocs, id)
			continue
		}
		add(id, s.Inputs(id), false)
	}
	ends, err := g.AssociationInputs(assocs)
	if err != nil {
		return err
	}
	for _, id := range assocs {
		add(id, ends[string(id)], true)
	}
	for start := 0; start < len(rows); start += inputRowsPerInsert {
		end := min(start+inputRowsPerInsert, len(rows))
		marks := strings.TrimSuffix(strings.Repeat("(?, ?, ?, ?, ?), ", end-start), ", ")
		args := make([]any, 0, 5*(end-start))
		for _, r := range rows[start:end] {
			flag := 0
			if r.association {
				flag = 1
			}
			args = append(args, r.dep, r.key.ID, r.key.Bridge, r.key.Endpoint, flag)
		}
		if _, err := q.Exec(sqlInsertInputsHead+marks, args...); err != nil {
			return err
		}
	}
	return nil
}

// inputRow is one row of the index: dep was composed from key. association
// marks the rows that are an association's ends, which a summary's link keys
// look like.
type inputRow struct {
	dep         []byte
	key         graphcache.InputKey
	association bool
}

// ReprojectChanged reprojects what a write moved. written is every handle the
// write recomputed and moved the ones among them whose displayed rows
// changed. It finds the rows composed from what changed in the index, before
// anything is reprojected, because a reprojection deletes a gone or rewritten
// row's inputs:
//
//   - every row composed from a gone or merged handle, whatever its key;
//   - every row composed from a moved handle's own values;
//   - every row composed from a link at an end of an association that moved or
//     went, by the ends the index recorded and the ends it has now.
//
// A handle that was written but neither moved nor gone adds no key. It then
// reprojects written and those rows on one graph and returns the rows it found
// plus the ends of the changed links, which are the handles whose pages the
// app reloads. An end that no row reads through is reported, not reprojected.
func ReprojectChanged(q Querier, written, moved [][]byte) ([][]byte, error) {
	written = database.UniqueBlobIDs(written)
	if len(written) == 0 {
		return nil, nil
	}
	g := newGraph(q)
	if err := g.Nodes(written); err != nil {
		return nil, err
	}
	var gone [][]byte
	isGone := map[string]bool{}
	for _, id := range written {
		// Nodes loaded every entity that exists; Cached never loads, so an
		// absent id is gone without a query of its own.
		n, _ := g.Cached(id)
		if n == nil || n.Merged {
			gone = append(gone, id)
			isGone[string(id)] = true
		}
	}
	var keys []graphcache.InputKey
	var movedLive [][]byte
	for _, id := range database.UniqueBlobIDs(moved) {
		if !isGone[string(id)] {
			keys = append(keys, graphcache.InputKey{ID: id})
			movedLive = append(movedLive, id)
		}
	}
	changed := append(append([][]byte(nil), gone...), movedLive...)
	old, err := oldEnds(q, changed)
	if err != nil {
		return nil, err
	}
	fresh, err := g.AssociationInputs(movedLive)
	if err != nil {
		return nil, err
	}
	var ends [][]byte
	for _, set := range []map[string][]graphcache.InputKey{old, fresh} {
		for _, ks := range set {
			for _, k := range ks {
				keys = append(keys, k)
				ends = append(ends, k.ID)
			}
		}
	}
	byKey, err := dependentsOf(q, keys)
	if err != nil {
		return nil, err
	}
	byHandle, err := dependentsOfHandles(q, gone)
	if err != nil {
		return nil, err
	}
	dependents := database.UniqueBlobIDs(append(byKey, byHandle...))
	if err := reprojectOn(q, g, append(append([][]byte(nil), written...), dependents...)); err != nil {
		return nil, err
	}
	out := database.UniqueBlobIDs(append(dependents, ends...))
	slices.SortFunc(out, bytes.Compare)
	return out, nil
}

// dependentsOf is the rows whose inputs hold any of the exact keys.
func dependentsOf(q Querier, keys []graphcache.InputKey) ([][]byte, error) {
	seen := map[string]graphcache.InputKey{}
	for _, k := range keys {
		seen[k.Key()] = k
	}
	unique := make([]graphcache.InputKey, 0, len(seen))
	for _, k := range seen {
		unique = append(unique, k)
	}
	var out [][]byte
	for start := 0; start < len(unique); start += keysPerLookup {
		batch := unique[start:min(start+keysPerLookup, len(unique))]
		marks := strings.TrimSuffix(strings.Repeat("(?, ?, ?), ", len(batch)), ", ")
		args := make([]any, 0, 3*len(batch))
		for _, k := range batch {
			args = append(args, k.ID, k.Bridge, k.Endpoint)
		}
		ids, err := listIDs(q, sqlDependentsOfKeys+marks+`)`, args...)
		if err != nil {
			return nil, err
		}
		out = append(out, ids...)
	}
	return database.UniqueBlobIDs(out), nil
}

// dependentsOfHandles is the rows whose inputs hold any key of the handles.
func dependentsOfHandles(q Querier, ids [][]byte) ([][]byte, error) {
	var out [][]byte
	err := sqlbatch.Query(q, sqlDependentsOfHandles, `)`, nil, database.UniqueBlobIDs(ids), func(rows *sql.Rows) error {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return err
		}
		out = append(out, id)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return database.UniqueBlobIDs(out), nil
}

// oldEnds is the ends the index recorded for each association in ids, by raw
// id. A row that is not an association has none.
func oldEnds(q Querier, ids [][]byte) (map[string][]graphcache.InputKey, error) {
	out := map[string][]graphcache.InputKey{}
	err := sqlbatch.Query(q, sqlOldEnds, `)`, nil, database.UniqueBlobIDs(ids), func(rows *sql.Rows) error {
		var dep, input []byte
		var k graphcache.InputKey
		if err := rows.Scan(&dep, &input, &k.Bridge, &k.Endpoint); err != nil {
			return err
		}
		k.ID = input
		out[string(dep)] = append(out[string(dep)], k)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, ks := range out {
		graphcache.SortInputKeys(ks)
	}
	return out, nil
}
