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
	sqlInsertInputsHead = `INSERT OR IGNORE INTO catalog_search_inputs (dependent_id, input_id, bridge, endpoint, dependent_is_association) VALUES `
	// inputRowsPerInsert keeps one statement under SQLite's variable limit.
	inputRowsPerInsert = 160
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
