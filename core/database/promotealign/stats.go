package promotealign

import (
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/graphalign"
)

// loadStats returns this catalog's scan. A commit that names handles drops it.
// A vocabulary-only write does not.
func loadStats(c *database.Catalog) (graphalign.Stats, error) {
	db, err := c.DB()
	if err != nil {
		return graphalign.Stats{}, err
	}
	if cached, ok := c.Graph().Stats(); ok {
		return cloneStats(cached.(graphalign.Stats)), nil
	}
	st, err := computeStats(c.Graph(), db)
	if err != nil {
		return graphalign.Stats{}, err
	}
	c.Graph().SetStats(st)
	return cloneStats(st), nil
}

func cloneStats(s graphalign.Stats) graphalign.Stats {
	out := graphalign.Stats{FanOut: map[string]float64{}}
	for k, v := range s.FanOut {
		out.FanOut[k] = v
	}
	return out
}

func computeStats(g *graphcache.Graph, q Querier) (graphalign.Stats, error) {
	st := graphalign.Stats{FanOut: map[string]float64{}}

	kin, err := loadKinship(g)
	if err != nil {
		return graphalign.Stats{}, err
	}
	sums := map[string]fanOutSum{}
	for _, b := range connectrules.Bridges() {
		if err := addFanOut(q, b, kin, sums); err != nil {
			return graphalign.Stats{}, err
		}
	}
	for key, s := range sums {
		if s.ends > 0 {
			st.FanOut[key] = s.edges / s.ends
		}
	}
	return st, nil
}

// sqlFanOut averages, per (disambiguation term, neighbor type term), how many
// filed associations of one bridge type each from-end handle has. Those two
// terms plus the bridge type and neighbor kind are the edge signature the
// layer and canon loaders build, so the keys line up exactly.
const sqlFanOut = `SELECT COALESCE(dt.key, ''), COALESCE(nt.key, ''),
		COUNT(*), COUNT(DISTINCT f.value_entity_id)
	FROM canonical_entities assoc
	JOIN subject_types st ON st.id = assoc.subject_type_id AND st.key = ? AND st.origin = ?
	JOIN auto_reconciler_values f ON f.entity_id = assoc.id AND f.rank = 1 AND f.reason = 'kept'
		AND f.property_id = (SELECT id FROM properties WHERE key = ? AND origin = ?)
	JOIN auto_reconciler_values n ON n.entity_id = assoc.id AND n.rank = 1 AND n.reason = 'kept'
		AND n.property_id = (SELECT id FROM properties WHERE key = ? AND origin = ?)
	LEFT JOIN auto_reconciler_values d ON d.entity_id = assoc.id AND d.rank = 1 AND d.reason = 'kept'
		AND d.property_id = (SELECT id FROM properties WHERE key = ? AND origin = ?)
	LEFT JOIN property_terms dt ON dt.id = d.value_term_id
	LEFT JOIN auto_reconciler_values nv ON nv.entity_id = n.value_entity_id AND nv.rank = 1 AND nv.reason = 'kept'
		AND nv.property_id = (SELECT id FROM properties WHERE key = ? AND origin = 'provenencia')
	LEFT JOIN property_terms nt ON nt.id = nv.value_term_id
	WHERE assoc.merged_into_id IS NULL
	GROUP BY dt.key, nt.key`

// fanOutSum is one signature's association count over its distinct
// from-end handles.
type fanOutSum struct{ edges, ends float64 }

// addFanOut adds one bridge type's counts to sums. A relationship term and
// its inverse share one signature key, so their counts add up under it.
func addFanOut(q Querier, b connectrules.Bridge, kin kinship, sums map[string]fanOutSum) error {
	if len(b.Endpoints) != 2 {
		return nil
	}
	from, to := b.Endpoints[0], b.Endpoints[1]
	disamb := ""
	if connectrules.HasDisambiguation(b.Disambiguation) {
		disamb = b.Disambiguation
	}
	rows, err := q.Query(sqlFanOut,
		b.BridgeTypeKey, b.Origin,
		from.PropertyKey, b.Origin,
		to.PropertyKey, b.Origin,
		disamb, b.Origin,
		neighborTypeProperty[to.TypeKey],
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var term, neighborType string
		var edges, ends float64
		if err := rows.Scan(&term, &neighborType, &edges, &ends); err != nil {
			return err
		}
		if ends <= 0 {
			continue
		}
		if disamb != "" {
			term, _ = kin.canonical(disamb, term)
		}
		sig := graphalign.EdgeSignature{
			BridgeType: b.BridgeTypeKey, RoleOrType: term,
			NeighborKind: to.TypeKey, NeighborTypeTerm: neighborType,
		}
		sum := sums[sig.Key()]
		sum.edges += edges
		sum.ends += ends
		sums[sig.Key()] = sum
	}
	return rows.Err()
}
