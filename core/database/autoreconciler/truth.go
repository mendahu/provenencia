package autoreconciler

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/valuecodec"
)

func init() {
	graphcache.SetTruth(truthRows)
}

// rowQuerier adapts the graph's read surface to this package's Querier.
type rowQuerier struct{ q graphcache.Querier }

func (r rowQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	return r.q.Query(query, args...)
}

func (r rowQuerier) QueryRow(query string, args ...any) *sql.Row {
	return r.q.QueryRow(query, args...)
}

func (r rowQuerier) Exec(string, ...any) (sql.Result, error) {
	return nil, sql.ErrNoRows
}

// truthRows is the reconcile Verify compares to a loaded node. It does not write.
func truthRows(q graphcache.Querier, ids [][]byte) (map[string][]graphcache.TruthRow, error) {
	out := map[string][]graphcache.TruthRow{}
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return out, nil
	}
	rq := rowQuerier{q}
	for start := 0; start < len(ids); start += batchSize {
		chunk := ids[start:min(start+batchSize, len(ids))]
		groups, err := load(rq, chunk)
		if err != nil {
			return nil, err
		}
		for _, g := range groups {
			res, err := autoreconcile.Reconcile(g.valueType, g.candidates, nil, g.cardinality)
			if err != nil {
				return nil, err
			}
			key := string(g.entityID)
			for i, cl := range res.Values {
				row, err := truthRow(g, i+1, cl)
				if err != nil {
					return nil, err
				}
				out[key] = append(out[key], row)
			}
		}
	}
	return out, nil
}

func truthRow(g *group, rank int, cl autoreconcile.ReconciledValue) (graphcache.TruthRow, error) {
	v := cl.Value
	row := graphcache.TruthRow{
		PropertyID: append([]byte(nil), g.propertyID...),
		Rank:       rank,
		Reason:     string(cl.Reason),
		Text:       v.Text,
		HasText:    v.HasText,
		Integer:    v.Integer,
		HasInteger: v.HasInteger,
		TermID:     append([]byte(nil), v.TermID...),
		EntityID:   append([]byte(nil), v.EntityID...),
	}
	if v.Date != nil {
		b, err := valuecodec.MarshalDate(*v.Date)
		if err != nil {
			return graphcache.TruthRow{}, err
		}
		row.Date = b
	}
	if v.Name != nil {
		b, err := valuecodec.MarshalName(*v.Name)
		if err != nil {
			return graphcache.TruthRow{}, err
		}
		row.Name = b
	}
	return row, nil
}
