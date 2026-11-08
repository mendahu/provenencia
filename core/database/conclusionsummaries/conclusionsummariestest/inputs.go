// Package conclusionsummariestest checks the inputs the summary composers
// record: against the reads of each row composed alone, and against the
// hand-written Dependents walk. The scenario worlds and the large fixture
// share it.
package conclusionsummariestest

import (
	"bytes"
	"sort"
	"testing"

	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Kinds are the handle kinds that have summary rows.
var Kinds = []string{vocabkey.Person, vocabkey.Event, vocabkey.Place}

// Querier is the read surface the checks use.
type Querier = graphcache.Querier

// HandleIDs returns the ids of every unmerged canonical entity of a subject
// type key, in id order.
func HandleIDs(t testing.TB, q Querier, kind string) [][]byte {
	t.Helper()
	rows, err := q.Query(`SELECT e.id FROM canonical_entities e
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE st.key = ? AND e.merged_into_id IS NULL ORDER BY e.id`, kind)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func want(kind string, ids [][]byte) conclusionsummaries.Want {
	var w conclusionsummaries.Want
	for _, id := range ids {
		w.Add(kind, id)
	}
	return w
}

// isolator composes one row at a time on a single graph, so nodes load once.
// Its recorder is the graph's test recorder, not the scopes the composers
// record through; it is swapped for a fresh set before each row.
type isolator struct {
	t    testing.TB
	g    *graphcache.Graph
	id   []byte
	seen map[string]graphcache.InputKey
}

func newIsolator(t testing.TB, q Querier) *isolator {
	iso := &isolator{t: t, g: graphcache.Over(q)}
	iso.g.SetInputRecorderForTest(func(k graphcache.InputKey) {
		if k.Bridge == "" && bytes.Equal(k.ID, iso.id) {
			return
		}
		iso.seen[string(k.ID)+"\x00"+k.Bridge+"\x00"+k.Endpoint] = graphcache.InputKey{
			ID: append([]byte(nil), k.ID...), Bridge: k.Bridge, Endpoint: k.Endpoint,
		}
	})
	return iso
}

// row returns the keys read composing the row of id alone, without the row's
// own values.
func (iso *isolator) row(kind string, id []byte) []graphcache.InputKey {
	iso.t.Helper()
	iso.id, iso.seen = id, map[string]graphcache.InputKey{}
	if _, err := conclusionsummaries.RowsForSearch(iso.g, want(kind, [][]byte{id})); err != nil {
		iso.t.Fatal(err)
	}
	out := make([]graphcache.InputKey, 0, len(iso.seen))
	for _, k := range iso.seen {
		out = append(out, k)
	}
	graphcache.SortInputKeys(out)
	return out
}

func sameKeys(a, b []graphcache.InputKey) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i].ID, b[i].ID) || a[i].Bridge != b[i].Bridge || a[i].Endpoint != b[i].Endpoint {
			return false
		}
	}
	return true
}

// Stats is how many keys the rows of each kind hold.
type Stats struct {
	Rows, Keys, Max int
	Counts          []int
}

// Median is the median key count per row.
func (s Stats) Median() int {
	if len(s.Counts) == 0 {
		return 0
	}
	c := append([]int(nil), s.Counts...)
	sort.Ints(c)
	return c[len(c)/2]
}

// Rows composes every id of each kind in one batch over a fresh graph and
// returns the summaries, whose Inputs the checks read.
func Rows(t testing.TB, q Querier, ids map[string][][]byte) conclusionsummaries.Summaries {
	t.Helper()
	var w conclusionsummaries.Want
	for kind, list := range ids {
		for _, id := range list {
			w.Add(kind, id)
		}
	}
	s, err := conclusionsummaries.RowsForSearch(graphcache.Over(q), w)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// CheckMatchIsolated fails when the batch-composed Inputs of any id differ
// from the keys that row reads composed alone. It returns the key counts per
// kind.
func CheckMatchIsolated(t testing.TB, q Querier, ids map[string][][]byte) map[string]Stats {
	t.Helper()
	rows := Rows(t, q, ids)
	iso := newIsolator(t, q)
	stats := map[string]Stats{}
	for _, kind := range Kinds {
		st := stats[kind]
		for _, id := range ids[kind] {
			got, want := rows.Inputs(id), iso.row(kind, id)
			if !sameKeys(got, want) {
				t.Errorf("%s %x: batch inputs %s, composed alone %s", kind, id, show(got), show(want))
			}
			st.Rows++
			st.Keys += len(got)
			st.Counts = append(st.Counts, len(got))
			if len(got) > st.Max {
				st.Max = len(got)
			}
		}
		stats[kind] = st
	}
	return stats
}

// LargestBatch composes each kind's ids in calls of size rows and returns, per
// kind, the largest number of keys one call's rows hold together.
func LargestBatch(t testing.TB, q Querier, ids map[string][][]byte, size int) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, kind := range Kinds {
		list := ids[kind]
		for start := 0; start < len(list); start += size {
			chunk := list[start:min(start+size, len(list))]
			rows, err := conclusionsummaries.RowsForSearch(graphcache.Over(q), want(kind, chunk))
			if err != nil {
				t.Fatal(err)
			}
			total := 0
			for _, id := range chunk {
				total += len(rows.Inputs(id))
			}
			out[kind] = max(out[kind], total)
		}
	}
	return out
}

func show(keys []graphcache.InputKey) string {
	var b bytes.Buffer
	b.WriteString("[")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(hex8(k.ID))
		if k.Bridge != "" {
			b.WriteString("/" + k.Bridge + "/" + k.Endpoint)
		}
	}
	b.WriteString("]")
	return b.String()
}

func hex8(id []byte) string {
	const digits = "0123456789abcdef"
	if len(id) > 4 {
		id = id[:4]
	}
	out := make([]byte, 0, 8)
	for _, c := range id {
		out = append(out, digits[c>>4], digits[c&15])
	}
	return string(out)
}

// CheckCoversDependents fails when, for a handle X in probe, the rows whose
// inputs hold X (its values or any of its link keys) miss an id
// conclusionsummaries.Dependents([X]) returns, except X's immediate part_of
// parents and its succession neighbours, the hops a detail page reads. The
// extras are returned, counted per kind. A missing person is excused (D-13)
// when every birth or death event in Dependents([X]) that it is a subject of
// has no accepted member: Dependents reads the kept event type alone, and a
// person row reads only a canonical event's places. The excused count is
// returned apart from the extras. ids is every row composed, by kind.
func CheckCoversDependents(t testing.TB, q Querier, ids map[string][][]byte, probe [][]byte) (extras map[string]int, excused int) {
	t.Helper()
	rows := Rows(t, q, ids)
	kindOf := map[string]string{}
	holders := map[string]map[string]bool{}
	for _, kind := range Kinds {
		for _, id := range ids[kind] {
			kindOf[string(id)] = kind
			for _, k := range rows.Inputs(id) {
				if holders[string(k.ID)] == nil {
					holders[string(k.ID)] = map[string]bool{}
				}
				holders[string(k.ID)][string(id)] = true
			}
		}
	}
	extras = map[string]int{}
	lifeEvent := map[string]bool{}
	canonical := map[string]bool{}
	// excusable reports whether person d, missing for a dependent set deps,
	// is reached only through birth or death events that are not canonical.
	excusable := func(d []byte, isDep map[string]bool) bool {
		edges, err := canonicalgraph.Walk(q, hops.EventsOfSubject, [][]byte{d})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, e := range canonicalgraph.Targets(edges) {
			if !isDep[string(e)] {
				continue
			}
			life, ok := lifeEvent[string(e)]
			if !ok {
				var n int
				if err := q.QueryRow(`SELECT COUNT(*) FROM auto_reconciler_values v
					JOIN properties p ON p.id = v.property_id AND p.key = 'event_type' AND p.origin = 'provenencia'
					JOIN property_terms tt ON tt.id = v.value_term_id AND tt.key IN ('birth', 'death')
					WHERE v.entity_id = ? AND v.rank = 1 AND v.reason = 'kept'`, e).Scan(&n); err != nil {
					t.Fatal(err)
				}
				life = n > 0
				lifeEvent[string(e)] = life
			}
			if !life {
				continue
			}
			can, ok := canonical[string(e)]
			if !ok {
				var n int
				if err := q.QueryRow(`SELECT COUNT(*) FROM identity_claims WHERE entity_id = ? AND status = 'accepted'`, e).Scan(&n); err != nil {
					t.Fatal(err)
				}
				can = n > 0
				canonical[string(e)] = can
			}
			if can {
				return false
			}
			found = true
		}
		return found
	}
	for _, x := range probe {
		deps, err := conclusionsummaries.Dependents(q, [][]byte{x})
		if err != nil {
			t.Fatal(err)
		}
		isDep := map[string]bool{}
		for _, d := range deps {
			isDep[string(d)] = true
		}
		skip := map[string]bool{}
		for _, hop := range []hops.Hop{hops.ParentsOfPlace, hops.SuccessorsOfPlace, hops.PredecessorsOfPlace} {
			edges, err := canonicalgraph.Walk(q, hop, [][]byte{x})
			if err != nil {
				t.Fatal(err)
			}
			for _, to := range canonicalgraph.Targets(edges) {
				skip[string(to)] = true
			}
		}
		for _, d := range deps {
			if skip[string(d)] || kindOf[string(d)] == "" {
				continue
			}
			if !holders[string(x)][string(d)] {
				if kindOf[string(d)] == vocabkey.Person && excusable(d, isDep) {
					excused++
					continue
				}
				t.Errorf("%s %x is a dependent of %x by Dependents, and its inputs do not hold it", kindOf[string(d)], d, x)
			}
		}
		for d := range holders[string(x)] {
			if !isDep[d] {
				extras[kindOf[d]]++
			}
		}
	}
	return extras, excused
}
