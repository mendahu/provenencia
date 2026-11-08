package searchindex

import (
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
)

// catalog_search_inputs records what each person, event and place summary was
// composed from, and the handles at each association's ends, so a write can
// find the rows it moved. This package writes it with the documents, in the
// same transaction, and clears and rebuilds it with them.
const (
	sqlDeleteInputs     = `DELETE FROM catalog_search_inputs WHERE dependent_id IN (`
	sqlInsertInputsHead = `INSERT OR IGNORE INTO catalog_search_inputs (dependent_id, input_id, bridge, endpoint) VALUES `
	// inputRowsPerInsert keeps one statement under SQLite's variable limit.
	inputRowsPerInsert = 200
)

// writeInputs replaces the inputs of every id in the batch as a dependent: the
// composed keys of a person, event or place that is not merged, and the ends
// of an association handle. A gone or merged id keeps none.
func writeInputs(q Querier, g *graphcache.Graph, ids [][]byte, heads map[string]handleHead, s conclusionsummaries.Summaries) error {
	if _, err := q.Exec(sqlDeleteInputs+database.SQLInPlaceholders(len(ids))+`)`, database.BlobArgs(ids)...); err != nil {
		return err
	}
	var rows []graphcache.InputKey
	var deps [][]byte
	add := func(dep []byte, keys []graphcache.InputKey) {
		for _, k := range keys {
			rows = append(rows, k)
			deps = append(deps, dep)
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
		add(id, s.Inputs(id))
	}
	ends, err := g.AssociationInputs(assocs)
	if err != nil {
		return err
	}
	for _, id := range assocs {
		add(id, ends[string(id)])
	}
	for start := 0; start < len(rows); start += inputRowsPerInsert {
		end := min(start+inputRowsPerInsert, len(rows))
		marks := strings.TrimSuffix(strings.Repeat("(?, ?, ?, ?), ", end-start), ", ")
		args := make([]any, 0, 4*(end-start))
		for i := start; i < end; i++ {
			args = append(args, deps[i], rows[i].ID, rows[i].Bridge, rows[i].Endpoint)
		}
		if _, err := q.Exec(sqlInsertInputsHead+marks, args...); err != nil {
			return err
		}
	}
	return nil
}
