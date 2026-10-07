package promotealign

import (
	"database/sql"
	"sync"

	"github.com/mendahu/provenencia/core/graphalign"
)

// Process-local stats cache keyed by the catalog's latest audit revision
// (design §8). Any write bumps revision and invalidates.
var (
	statsMu    sync.Mutex
	statsRev   int64
	statsCached graphalign.Stats
)

func loadStats(q Querier) (graphalign.Stats, error) {
	var rev int64
	if err := q.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		return graphalign.Stats{}, err
	}

	statsMu.Lock()
	defer statsMu.Unlock()
	if rev == statsRev && statsRev >= 0 {
		return cloneStats(statsCached), nil
	}

	st, err := computeStats(q)
	if err != nil {
		return graphalign.Stats{}, err
	}
	statsRev = rev
	statsCached = st
	return cloneStats(st), nil
}

// ResetStatsCacheForTest clears the process-local cache (tests only).
func ResetStatsCacheForTest() {
	statsMu.Lock()
	defer statsMu.Unlock()
	statsRev = -1
	statsCached = graphalign.Stats{}
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

	// Fan-out: average neighbors per handle through each product hop signature
	// class. Cold catalogs leave FanOut empty (Align uses EdgeSupportHigh path
	// via FanOutLowMax defaults when missing — see edgeSupport).
	for _, step := range []struct {
		sig graphalign.EdgeSignature
		sql string
	}{
		{
			sig: graphalign.EdgeSignature{BridgeType: "participation", RoleOrType: "subject", NeighborKind: "event"},
			sql: `SELECT AVG(cnt) FROM (
				SELECT COUNT(*) AS cnt FROM auto_reconciler_values a
				JOIN canonical_entities assoc ON assoc.id = a.entity_id AND assoc.merged_into_id IS NULL
				JOIN subject_types st ON st.id = assoc.subject_type_id AND st.key = 'participation' AND st.origin = 'provenencia'
				JOIN auto_reconciler_values role ON role.entity_id = a.entity_id AND role.rank = 1 AND role.reason = 'kept'
				JOIN properties rp ON rp.id = role.property_id AND rp.key = 'role' AND rp.origin = 'provenencia'
				JOIN property_terms rt ON rt.id = role.value_term_id AND rt.key = 'subject'
				WHERE a.rank = 1 AND a.reason = 'kept'
					AND a.property_id = (SELECT id FROM properties WHERE key = 'person' AND origin = 'provenencia')
				GROUP BY a.value_entity_id
			)`,
		},
		{
			sig: graphalign.EdgeSignature{BridgeType: "location", NeighborKind: "place"},
			sql: `SELECT AVG(cnt) FROM (
				SELECT COUNT(*) AS cnt FROM auto_reconciler_values a
				JOIN canonical_entities assoc ON assoc.id = a.entity_id AND assoc.merged_into_id IS NULL
				JOIN subject_types st ON st.id = assoc.subject_type_id AND st.key = 'location' AND st.origin = 'provenencia'
				WHERE a.rank = 1 AND a.reason = 'kept'
					AND a.property_id = (SELECT id FROM properties WHERE key = 'event' AND origin = 'provenencia')
				GROUP BY a.value_entity_id
			)`,
		},
		{
			sig: graphalign.EdgeSignature{BridgeType: "place_relationship", RoleOrType: "part_of", NeighborKind: "place"},
			sql: `SELECT AVG(cnt) FROM (
				SELECT COUNT(*) AS cnt FROM auto_reconciler_values a
				JOIN canonical_entities assoc ON assoc.id = a.entity_id AND assoc.merged_into_id IS NULL
				JOIN subject_types st ON st.id = assoc.subject_type_id AND st.key = 'place_relationship' AND st.origin = 'provenencia'
				JOIN auto_reconciler_values typ ON typ.entity_id = a.entity_id AND typ.rank = 1 AND typ.reason = 'kept'
				JOIN properties tp ON tp.id = typ.property_id AND tp.key = 'place_relationship_type' AND tp.origin = 'provenencia'
				JOIN property_terms tt ON tt.id = typ.value_term_id AND tt.key = 'part_of'
				WHERE a.rank = 1 AND a.reason = 'kept'
					AND a.property_id = (SELECT id FROM properties WHERE key = 'from' AND origin = 'provenencia')
				GROUP BY a.value_entity_id
			)`,
		},
	} {
		var avg sql.NullFloat64
		if err := q.QueryRow(step.sql).Scan(&avg); err != nil {
			return graphalign.Stats{}, err
		}
		if avg.Valid && avg.Float64 > 0 {
			st.FanOut[step.sig.Key()] = avg.Float64
		}
	}

	return st, nil
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
