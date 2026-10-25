package graphcache_test

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type capStep struct {
	op  string // "read" (EvidenceGraph), "memo" (SetEvidenceGraphMemo(id, "k", "v")), "hit" (EvidenceGraphMemo, must be ok), "drop" (g.Drop())
	src int    // index into the four graph ids A, B, C, D
}

// capFixture is a catalog with four Sources, so four Evidence graphs, and a
// graph capped at two.
func capFixture(t *testing.T) (*graphcache.Graph, [][]byte) {
	t.Helper()
	f := newWalkFixture(t)
	typeID, err := sourcetypes.Upsert(f.c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	ids := [][]byte{evidencegraphstest.SoleGraph(t, f.c, f.src.ID).ID}
	for _, title := range []string{"Second", "Third", "Fourth"} {
		src, _ := f.newSource(typeID, title)
		ids = append(ids, evidencegraphstest.SoleGraph(t, f.c, src.ID).ID)
	}
	g := f.c.Graph()
	g.Drop()
	g.SetEvidenceGraphCapForTest(2)
	t.Cleanup(func() { g.SetEvidenceGraphCapForTest(0) })
	return g, ids
}

// The graph keeps at most the cap of Evidence graphs and drops the least
// recently used one, with its memos.
func TestEvidenceGraphCap(t *testing.T) {
	cases := []struct {
		name      string
		steps     []capStep
		loaded    []int
		notLoaded []int
	}{
		{"oldest is evicted", []capStep{{"read", 0}, {"memo", 0}, {"read", 1}, {"read", 2}}, []int{1, 2}, []int{0}},
		{"a read is a use", []capStep{{"read", 0}, {"read", 1}, {"read", 0}, {"read", 2}}, []int{0, 2}, []int{1}},
		{"a memo hit is a use", []capStep{{"read", 0}, {"memo", 0}, {"read", 1}, {"hit", 0}, {"read", 2}}, []int{0, 2}, []int{1}},
		{"a drop clears the loaded set", []capStep{{"read", 0}, {"read", 1}, {"drop", 0}, {"read", 2}, {"read", 3}}, []int{2, 3}, []int{0, 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, ids := capFixture(t)
			for _, s := range tc.steps {
				switch s.op {
				case "read":
					if _, err := g.EvidenceGraph(ids[s.src]); err != nil {
						t.Fatal(err)
					}
				case "memo":
					g.SetEvidenceGraphMemo(ids[s.src], "k", "v")
				case "hit":
					if _, ok := g.EvidenceGraphMemo(ids[s.src], "k"); !ok {
						t.Fatal("memo was not held")
					}
				case "drop":
					g.Drop()
				}
			}
			for _, i := range tc.loaded {
				if !g.LoadedEvidenceGraphForTest(ids[i]) {
					t.Errorf("graph %d is not loaded", i)
				}
			}
			for _, i := range tc.notLoaded {
				if g.LoadedEvidenceGraphForTest(ids[i]) {
					t.Errorf("graph %d is loaded", i)
				}
				if _, ok := g.EvidenceGraphMemo(ids[i], "k"); ok {
					t.Errorf("graph %d kept a memo", i)
				}
			}
		})
	}

	t.Run("an evicted graph reloads the same", func(t *testing.T) {
		g, ids := capFixture(t)
		first, err := g.EvidenceGraph(ids[0])
		if err != nil {
			t.Fatal(err)
		}
		g.SetEvidenceGraphMemo(ids[0], "k", "v")
		for _, i := range []int{1, 2} {
			if _, err := g.EvidenceGraph(ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		if g.LoadedEvidenceGraphForTest(ids[0]) {
			t.Fatal("graph 0 was not evicted")
		}
		again, err := g.EvidenceGraph(ids[0])
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(again, first) {
			t.Errorf("reloaded graph differs: %+v vs %+v", again, first)
		}
		if _, ok := g.EvidenceGraphMemo(ids[0], "k"); ok {
			t.Error("memo survived eviction")
		}
	})

	t.Run("verify skips an evicted graph", func(t *testing.T) {
		g, ids := capFixture(t)
		for i := 0; i < 3; i++ {
			if _, err := g.EvidenceGraph(ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.Verify(); err != nil {
			t.Fatal(err)
		}
		if g.LoadedEvidenceGraphForTest(ids[0]) {
			t.Error("Verify loaded the evicted graph")
		}
	})

	t.Run("a commit drop removes the entry", func(t *testing.T) {
		g, ids := capFixture(t)
		for _, i := range []int{0, 1} {
			if _, err := g.EvidenceGraph(ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 10}, Effects: effects.Set{EvidenceGraphs: [][]byte{ids[0]}}}); err != nil {
			t.Fatal(err)
		}
		for _, i := range []int{2, 3} {
			if _, err := g.EvidenceGraph(ids[i]); err != nil {
				t.Fatal(err)
			}
		}
		for _, i := range []int{2, 3} {
			if !g.LoadedEvidenceGraphForTest(ids[i]) {
				t.Errorf("graph %d is not loaded", i)
			}
		}
		for _, i := range []int{0, 1} {
			if g.LoadedEvidenceGraphForTest(ids[i]) {
				t.Errorf("graph %d is loaded", i)
			}
		}
	})
}
