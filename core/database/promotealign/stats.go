package promotealign

import (
	"sync"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/graphalign"
)

// Process-local stats cache keyed by the catalog file and its latest audit
// revision (design §8). Any write bumps the revision and invalidates; a
// different project never reads another's frequencies.
var (
	statsMu     sync.Mutex
	statsKey    statsStamp
	statsCached graphalign.Stats
	statsValid  bool
)

type statsStamp struct {
	file string
	rev  int64
}

// catalogStamp names the catalog file and its latest audit revision: a
// process-local cache keyed by it is fresh until any write.
func catalogStamp(q Querier) (statsStamp, error) {
	var stamp statsStamp
	if err := q.QueryRow(`SELECT file FROM pragma_database_list WHERE name = 'main'`).Scan(&stamp.file); err != nil {
		return statsStamp{}, err
	}
	if err := q.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&stamp.rev); err != nil {
		return statsStamp{}, err
	}
	return stamp, nil
}

func loadStats(q Querier) (graphalign.Stats, error) {
	stamp, err := catalogStamp(q)
	if err != nil {
		return graphalign.Stats{}, err
	}

	statsMu.Lock()
	defer statsMu.Unlock()
	// An in-memory catalog has no file, so it can't be told apart: never cache it.
	if statsValid && stamp.file != "" && stamp == statsKey {
		return cloneStats(statsCached), nil
	}

	st, err := computeStats(q)
	if err != nil {
		return graphalign.Stats{}, err
	}
	statsKey, statsCached, statsValid = stamp, st, true
	return cloneStats(st), nil
}

// ResetStatsCacheForTest clears the process-local caches (tests only).
func ResetStatsCacheForTest() {
	statsMu.Lock()
	statsKey, statsCached, statsValid = statsStamp{}, graphalign.Stats{}, false
	statsMu.Unlock()
	candMu.Lock()
	candKey, candCached = statsStamp{}, nil
	candMu.Unlock()
}

func computeStats(q Querier) (graphalign.Stats, error) {
	st := graphalign.Stats{
		ValueFreq: map[string]map[string]float64{},
		FanOut:    map[string]float64{},
	}

	// Value frequencies: fraction of unmerged handles carrying each text/term
	// value for profile Properties. Approximates u among unrelated handles.
	rows, err := q.Query(`SELECT p.key,
			COALESCE(r.value_text, ''),
			COALESCE(t.key, ''),
			COUNT(DISTINCT r.entity_id)
		FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id AND p.origin = 'provenencia'
		JOIN canonical_entities e ON e.id = r.entity_id AND e.merged_into_id IS NULL
		LEFT JOIN property_terms t ON t.id = r.value_term_id
		WHERE r.rank = 1 AND r.reason = 'kept'
			AND p.key IN ('name', 'sex_at_birth', 'event_type', 'toponym')
		GROUP BY p.key, r.value_text, t.key`)
	if err != nil {
		return graphalign.Stats{}, err
	}
	defer rows.Close()

	totals := map[string]float64{}
	type cell struct {
		prop, key string
		n         float64
	}
	var cells []cell
	for rows.Next() {
		var prop, text, term string
		var n float64
		if err := rows.Scan(&prop, &text, &term, &n); err != nil {
			return graphalign.Stats{}, err
		}
		k := text
		if k == "" {
			k = term
		}
		if k == "" {
			continue
		}
		cells = append(cells, cell{prop: prop, key: k, n: n})
		totals[prop] += n
	}
	if err := rows.Err(); err != nil {
		return graphalign.Stats{}, err
	}
	_ = rows.Close()
	for _, c := range cells {
		den := totals[c.prop]
		if den <= 0 {
			continue
		}
		if st.ValueFreq[c.prop] == nil {
			st.ValueFreq[c.prop] = map[string]float64{}
		}
		st.ValueFreq[c.prop][c.key] = c.n / den
	}

	kin, err := loadKinship(q)
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

func cloneStats(s graphalign.Stats) graphalign.Stats {
	out := graphalign.Stats{
		ValueFreq: map[string]map[string]float64{},
		FanOut:    map[string]float64{},
	}
	for k, m := range s.ValueFreq {
		cp := map[string]float64{}
		for vk, v := range m {
			cp[vk] = v
		}
		out.ValueFreq[k] = cp
	}
	for k, v := range s.FanOut {
		out.FanOut[k] = v
	}
	return out
}
