package evidencelayer_test

import (
	"crypto/rand"
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// links is the Evidence layer's links of src's graph, built under the
// catalog's name pattern.
func links(t *testing.T, r *evidencelayertest.Records, graph []byte) []kinship.Link {
	t.Helper()
	return r.Layer(graph).Links
}

// The loader reads a Source's links from the connect rules: one link for
// each registered bridge type, whatever its term.
func TestLoadCoversEveryConnectRule(t *testing.T) {
	r := evidencelayertest.New(t)
	ann := r.Person(r.Source, r.Artifact, "Ann Ames")
	bob := r.Person(r.Source, r.Artifact, "Bob Ames")
	birth := r.Birth(r.Source, r.Artifact)
	york := r.Place(r.Source, r.Artifact, "York")
	ontario := r.Place(r.Source, r.Artifact, "Ontario")
	r.Participation(r.Source, r.Artifact, ann, birth, "subject")
	r.Locate(r.Source, r.Artifact, birth, york)
	r.Relationship(r.Source, r.Artifact, ann, bob, "parent")
	r.PlaceRel(r.Source, r.Artifact, york, ontario, "part_of")

	var got []string
	for _, l := range links(t, r, r.Graph(r.Source)) {
		got = append(got, string(l.KeyA.Type))
	}
	want := connectrules.BridgeTypeKeys()
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("bridge types %v, want %v", got, want)
	}
}

// A Source that records a spouse relationship and a marriage has one spouse
// link, with every association behind it; a marriage alone has an implied one.
func TestLoadMergesKinshipShapes(t *testing.T) {
	spouses := func(ls []kinship.Link) []kinship.Link {
		var out []kinship.Link
		for _, l := range ls {
			if l.KeyA.Type == vocabkey.BridgeRelationship && l.KeyA.Term == "spouse" {
				out = append(out, l)
			}
		}
		return out
	}
	build := func(withRelationship bool) []kinship.Link {
		r := evidencelayertest.New(t)
		ann := r.Person(r.Source, r.Artifact, "Ann Ames")
		bob := r.Person(r.Source, r.Artifact, "Bob Ames")
		r.Marriage(r.Source, r.Artifact, ann, bob, "subject")
		if withRelationship {
			r.Relationship(r.Source, r.Artifact, ann, bob, "spouse")
		}
		return links(t, r, r.Graph(r.Source))
	}
	both := spouses(build(true))
	if len(both) != 1 || len(both[0].Associations) != 3 || len(both[0].Shapes) != 1 || both[0].OnlyImplied {
		t.Fatalf("recorded and implied: %+v, want one recorded link with 3 associations and one shape", both)
	}
	only := spouses(build(false))
	if len(only) != 1 || len(only[0].Associations) != 2 || !only[0].OnlyImplied {
		t.Fatalf("marriage only: %+v, want one implied link with 2 associations", only)
	}
}

// TestLayersEntry: the Layers entry holds a layer per graph and follows the
// signals it declares; a held layer needs no loaded graph, and a graph the
// cache does not find stores nothing.
func TestLayersEntry(t *testing.T) {
	type world struct {
		r    *evidencelayertest.Records
		g    *graphcache.Graph
		a, b []byte
	}
	newWorld := func(t *testing.T) world {
		t.Helper()
		r := evidencelayertest.New(t)
		r.Person(r.Source, r.Artifact, "Ann Ames")
		srcB, artB := r.NewSource("Second")
		r.Person(srcB, artB, "Bob Ames")
		w := world{r: r, g: r.C.Graph(), a: r.Graph(r.Source), b: r.Graph(srcB)}
		for _, id := range [][]byte{w.a, w.b} {
			if _, err := evidencelayer.Load(w.g, id); err != nil {
				t.Fatal(err)
			}
		}
		if n := w.g.KeyCounts()["evidencelayer.layers"]; n != 2 {
			t.Fatalf("setup holds %d layers, want 2", n)
		}
		return w
	}
	held := func(w world, id []byte) bool {
		_, ok := evidencelayer.Layers.Peek(w.g, id)
		return ok
	}
	commit := func(t *testing.T, w world, set effects.Set) {
		t.Helper()
		if err := w.g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1 << 40}, Effects: set}); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("a vocabulary commit empties it", func(t *testing.T) {
		w := newWorld(t)
		commit(t, w, effects.Set{Lists: []effects.List{effects.ListPropertyVocabulary}})
		if n := w.g.KeyCounts()["evidencelayer.layers"]; n != 0 {
			t.Fatalf("%d layers held after a vocabulary commit", n)
		}
	})
	t.Run("a name format change empties it", func(t *testing.T) {
		w := newWorld(t)
		if err := nameformats.Invalidate(w.r.C); err != nil {
			t.Fatal(err)
		}
		if n := w.g.KeyCounts()["evidencelayer.layers"]; n != 0 {
			t.Fatalf("%d layers held after Invalidate", n)
		}
	})
	t.Run("a write that names a graph drops that graph's layer", func(t *testing.T) {
		w := newWorld(t)
		w.r.Person(w.r.Source, w.r.Artifact, "Cy Ames")
		if held(w, w.a) || !held(w, w.b) {
			t.Fatalf("after a write on A: A held %v, B held %v, want false and true", held(w, w.a), held(w, w.b))
		}
		layer, err := evidencelayer.Load(w.g, w.a)
		if err != nil {
			t.Fatal(err)
		}
		if len(layer.Subjects) != 2 {
			t.Fatalf("the layer after the write has %d Subjects, want 2", len(layer.Subjects))
		}
	})
	t.Run("a held layer needs no loaded graph", func(t *testing.T) {
		w := newWorld(t)
		graphfill.Graphs.SetCapForTest(w.g, 1)
		t.Cleanup(func() { graphfill.Graphs.SetCapForTest(w.g, 0) })
		if _, err := graphfill.EvidenceGraph(w.g, w.b); err != nil {
			t.Fatal(err)
		}
		if _, ok := graphfill.PeekEvidenceGraph(w.g, w.a); ok {
			t.Fatal("graph A is still loaded")
		}
		queries := 0
		w.g.SetQueryCounterForTest(func() { queries++ })
		if _, err := evidencelayer.Load(w.g, w.a); err != nil {
			t.Fatal(err)
		}
		if queries != 0 {
			t.Fatalf("Load of a held layer issued %d queries", queries)
		}
		if _, ok := graphfill.PeekEvidenceGraph(w.g, w.a); ok {
			t.Fatal("Load of a held layer loaded the Evidence graph")
		}
	})
	t.Run("an unknown graph is ErrNoGraph and stores nothing", func(t *testing.T) {
		w := newWorld(t)
		w.g.Drop()
		id := make([]byte, 16)
		if _, err := rand.Read(id); err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 2; i++ {
			if _, err := evidencelayer.Load(w.g, id); !errors.Is(err, evidencelayer.ErrNoGraph) {
				t.Fatalf("call %d: err %v, want ErrNoGraph", i+1, err)
			}
		}
		if n := w.g.KeyCounts()["evidencelayer.layers"]; n != 0 {
			t.Fatalf("%d layers held for an unknown graph", n)
		}
	})
}
