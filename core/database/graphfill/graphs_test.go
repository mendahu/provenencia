package graphfill_test

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphfill"
)

type capStep struct {
	op  string // "read" (EvidenceGraph) or "drop" (g.Drop())
	src int    // index into the four graph ids A, B, C, D
}

// capFixture is a catalog with four Sources, so four Evidence graphs, and a
// loaded-graph store capped at two.
func capFixture(t *testing.T) (*evidencelayertest.Records, *graphcache.Graph, [][]byte) {
	t.Helper()
	r := evidencelayertest.New(t)
	ids := [][]byte{evidencegraphstest.SoleGraph(t, r.C, r.Source.ID).ID}
	for _, title := range []string{"Second", "Third", "Fourth"} {
		src, _ := r.NewSource(title)
		ids = append(ids, evidencegraphstest.SoleGraph(t, r.C, src.ID).ID)
	}
	g := r.C.Graph()
	g.Drop()
	graphfill.Graphs.SetCapForTest(g, 2)
	t.Cleanup(func() { graphfill.Graphs.SetCapForTest(g, 0) })
	return r, g, ids
}

// The store keeps at most its Cap of Evidence graphs and drops the least
// recently used one.
func TestEvidenceGraphCap(t *testing.T) {
	cases := []struct {
		name      string
		steps     []capStep
		loaded    []int
		notLoaded []int
	}{
		{"oldest is evicted", []capStep{{"read", 0}, {"read", 1}, {"read", 2}}, []int{1, 2}, []int{0}},
		{"a read is a use", []capStep{{"read", 0}, {"read", 1}, {"read", 0}, {"read", 2}}, []int{0, 2}, []int{1}},
		{"a drop clears the loaded set", []capStep{{"read", 0}, {"read", 1}, {"drop", 0}, {"read", 2}, {"read", 3}}, []int{2, 3}, []int{0, 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, g, ids := capFixture(t)
			for _, s := range tc.steps {
				switch s.op {
				case "read":
					if _, err := graphfill.EvidenceGraph(g, ids[s.src]); err != nil {
						t.Fatal(err)
					}
				case "drop":
					g.Drop()
				}
			}
			for _, i := range tc.loaded {
				if !loaded(g, ids[i]) {
					t.Errorf("graph %d is not loaded", i)
				}
			}
			for _, i := range tc.notLoaded {
				if loaded(g, ids[i]) {
					t.Errorf("graph %d is loaded", i)
				}
			}
		})
	}

	t.Run("an evicted graph reloads the same", func(t *testing.T) {
		_, g, ids := capFixture(t)
		first, err := graphfill.EvidenceGraph(g, ids[0])
		if err != nil {
			t.Fatal(err)
		}
		for _, i := range []int{1, 2} {
			if _, err := graphfill.EvidenceGraph(g, ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		if loaded(g, ids[0]) {
			t.Fatal("graph 0 was not evicted")
		}
		again, err := graphfill.EvidenceGraph(g, ids[0])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(again, first) {
			t.Errorf("reloaded graph differs: %+v vs %+v", again, first)
		}
	})

	t.Run("verify skips an evicted graph", func(t *testing.T) {
		_, g, ids := capFixture(t)
		for i := 0; i < 3; i++ {
			if _, err := graphfill.EvidenceGraph(g, ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.Verify(); err != nil {
			t.Fatal(err)
		}
		if loaded(g, ids[0]) {
			t.Error("Verify loaded the evicted graph")
		}
	})

	t.Run("a commit drop removes the entry", func(t *testing.T) {
		_, g, ids := capFixture(t)
		for _, i := range []int{0, 1} {
			if _, err := graphfill.EvidenceGraph(g, ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 10}, Effects: effects.Set{EvidenceGraphs: [][]byte{ids[0]}}}); err != nil {
			t.Fatal(err)
		}
		for _, i := range []int{2, 3} {
			if _, err := graphfill.EvidenceGraph(g, ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		for _, i := range []int{2, 3} {
			if !loaded(g, ids[i]) {
				t.Errorf("graph %d is not loaded", i)
			}
		}
		for _, i := range []int{0, 1} {
			if loaded(g, ids[i]) {
				t.Errorf("graph %d is loaded", i)
			}
		}
	})
}
