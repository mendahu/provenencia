package promotealign

import (
	"bytes"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// loadStats returns this catalog's scan. A commit that names handles drops it.
// A vocabulary-only write does not. The property list is the match profiles,
// not a fixed key list.
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
	// Words is never mutated after it is counted, so copies share its map.
	out.Words = s.Words
	return out
}

func computeStats(g *graphcache.Graph, q Querier) (graphalign.Stats, error) {
	st := graphalign.Stats{
		ValueFreq: map[string]map[string]float64{},
		FanOut:    map[string]float64{},
	}

	// Value frequencies: fraction of unmerged handles carrying each text/term
	// value for profile Properties. Approximates u among unrelated handles.
	keys := profilePropertyKeys()
	args := make([]any, len(keys))
	for i, k := range keys {
		args[i] = k
	}
	rows, err := q.Query(`SELECT p.key,
			COALESCE(r.value_text, ''),
			COALESCE(t.key, ''),
			COUNT(DISTINCT r.entity_id)
		FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id AND p.origin = 'provenencia'
		JOIN canonical_entities e ON e.id = r.entity_id AND e.merged_into_id IS NULL
		LEFT JOIN property_terms t ON t.id = r.value_term_id
		WHERE r.rank = 1 AND r.reason = 'kept'
			AND p.key IN (`+database.SQLInPlaceholders(len(keys))+`)
		GROUP BY p.key, r.value_text, t.key`, args...)
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

	words, err := countNameWords(q)
	if err != nil {
		return graphalign.Stats{}, err
	}
	st.Words = words

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

// sqlNameWords reads every reconciled name value of an unmerged Person
// handle, at every rank. Rows of one handle are adjacent.
const sqlNameWords = `SELECT r.entity_id, r.value_name
	FROM auto_reconciler_values r
	JOIN properties p ON p.id = r.property_id AND p.key = 'name' AND p.origin = 'provenencia'
	JOIN canonical_entities e ON e.id = r.entity_id AND e.merged_into_id IS NULL
	JOIN subject_types st ON st.id = e.subject_type_id AND st.key = 'person' AND st.origin = 'provenencia'
	WHERE r.value_name IS NOT NULL
	ORDER BY r.entity_id`

// countNameWords counts, per word, the filed people who carry it. A handle
// counts once however many name values it holds.
func countNameWords(q Querier) (match.WordCounts, error) {
	rows, err := q.Query(sqlNameWords)
	if err != nil {
		return match.WordCounts{}, err
	}
	defer rows.Close()
	var people [][]namevalue.Value
	var last []byte
	for rows.Next() {
		var id, blob []byte
		if err := rows.Scan(&id, &blob); err != nil {
			return match.WordCounts{}, err
		}
		n, err := valuecodec.UnmarshalName(blob)
		if err != nil {
			return match.WordCounts{}, err
		}
		if people == nil || !bytes.Equal(id, last) {
			people = append(people, nil)
			last = id
		}
		people[len(people)-1] = append(people[len(people)-1], n)
	}
	if err := rows.Err(); err != nil {
		return match.WordCounts{}, err
	}
	return match.DefaultNames.CountWords(people), nil
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

func profilePropertyKeys() []string {
	seen := map[string]bool{}
	var keys []string
	for _, kind := range vocabkey.PrimaryKinds() {
		profile, ok := match.DefaultProfile(kind)
		if !ok {
			continue
		}
		for _, prop := range profile.Properties() {
			if prop.Origin != vocabkey.OriginProvenencia || seen[prop.Key] {
				continue
			}
			seen[prop.Key] = true
			keys = append(keys, prop.Key)
		}
	}
	return keys
}
