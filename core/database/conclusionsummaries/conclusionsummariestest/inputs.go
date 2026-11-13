// Package conclusionsummariestest checks the inputs the summary composers
// record against the reads of each row composed alone. The scenario worlds and
// the large fixture share it.
package conclusionsummariestest

import (
	"bytes"
	"slices"
	"sort"
	"testing"

	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
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
	seen []graphcache.InputKey
}

func newIsolator(t testing.TB, q Querier) *isolator {
	iso := &isolator{t: t, g: graphcache.Over(q)}
	iso.g.SetInputRecorderForTest(func(k graphcache.InputKey) {
		if k.Bridge == "" && bytes.Equal(k.ID, iso.id) {
			return
		}
		iso.seen = append(iso.seen, graphcache.InputKey{
			ID: append([]byte(nil), k.ID...), Bridge: k.Bridge, Endpoint: k.Endpoint,
		})
	})
	return iso
}

// row returns the keys read composing the row of id alone, without the row's
// own values.
func (iso *isolator) row(kind string, id []byte) []graphcache.InputKey {
	iso.t.Helper()
	iso.id, iso.seen = id, nil
	if _, err := conclusionsummaries.RowsForSearch(iso.g, want(kind, [][]byte{id})); err != nil {
		iso.t.Fatal(err)
	}
	return graphcache.UniqueInputKeys(iso.seen)
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
			if !slices.EqualFunc(got, want, graphcache.InputKey.Equal) {
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
