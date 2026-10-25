package graphcache_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
)

// T-20: an Evidence graph's memo lives as long as the graph is loaded, and a
// drop of the graph, Drop, a vocabulary change and SetAliasIndex each clear it.
func TestEvidenceGraphMemoDroppedWithSource(t *testing.T) {
	f := newWalkFixture(t)
	g := f.c.Graph()
	id := evidencegraphstest.SoleGraph(t, f.c, f.src.ID).ID
	load := func(t *testing.T) {
		t.Helper()
		if _, err := g.EvidenceGraph(id); err != nil {
			t.Fatal(err)
		}
		g.SetEvidenceGraphMemo(id, "k", "layer")
		if v, ok := g.EvidenceGraphMemo(id, "k"); !ok || v != "layer" {
			t.Fatalf("memo = %v, %v after SetEvidenceGraphMemo", v, ok)
		}
	}
	held := func() bool {
		_, ok := g.EvidenceGraphMemo(id, "k")
		return ok
	}

	t.Run("a Source that is not loaded holds nothing", func(t *testing.T) {
		g.Drop()
		g.SetEvidenceGraphMemo(id, "k", "layer")
		if held() {
			t.Error("a memo was kept for a Source that is not loaded")
		}
	})
	t.Run("another key and another Source are separate", func(t *testing.T) {
		g.Drop()
		load(t)
		if _, ok := g.EvidenceGraphMemo(id, "other"); ok {
			t.Error("a key nobody set was held")
		}
		if _, ok := g.EvidenceGraphMemo(make([]byte, 16), "k"); ok {
			t.Error("another Source's memo was held")
		}
	})
	t.Run("a graph drop", func(t *testing.T) {
		g.Drop()
		load(t)
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1}, Effects: effects.Set{EvidenceGraphs: [][]byte{id}}}); err != nil {
			t.Fatal(err)
		}
		if held() {
			t.Error("a commit that names the graph left its memo")
		}
	})
	t.Run("a commit that names another graph", func(t *testing.T) {
		g.Drop()
		load(t)
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 2}, Effects: effects.Set{EvidenceGraphs: [][]byte{make([]byte, 16)}}}); err != nil {
			t.Fatal(err)
		}
		if !held() {
			t.Error("a commit on another graph dropped this memo")
		}
	})
	t.Run("Drop", func(t *testing.T) {
		load(t)
		g.Drop()
		if held() {
			t.Error("Drop left the memo")
		}
	})
	t.Run("a vocabulary change", func(t *testing.T) {
		load(t)
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 3}, Effects: effects.Set{Lists: []effects.List{effects.ListVocabulary}}}); err != nil {
			t.Fatal(err)
		}
		if held() {
			t.Error("a vocabulary commit left the memo")
		}
	})
	t.Run("SetAliasIndex", func(t *testing.T) {
		load(t)
		g.SetAliasIndex(nil)
		if held() {
			t.Error("SetAliasIndex left the memo")
		}
	})
}
