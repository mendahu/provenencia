package graphcache_test

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
)

var (
	idA = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	idB = []byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	idC = []byte{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
)

func noApply[T any](*graphcache.Graph, graphcache.Keyed[T], [][]byte, effects.Set) error { return nil }

// five is the entries of AC-1.1: whole A (catalog lifetime), whole B derived
// from A, keyed C derived from B, keyed D derived from the vocabulary, keyed E
// derived from nothing.
type five struct {
	a, b graphcache.Whole[string]
	c, d graphcache.Keyed[string]
	e    graphcache.Keyed[string]
}

func registerFive() five {
	var f five
	f.a = graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.a", Lifetime: graphcache.LifeCatalog})
	f.b = graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.b", DerivedFrom: []graphcache.Source{f.a}})
	f.c = graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{Name: "t.c", DerivedFrom: []graphcache.Source{f.b}, Apply: noApply[string], Cap: 4})
	f.d = graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{Name: "t.d", DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}}, Apply: noApply[string], Cap: 4})
	f.e = graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{Name: "t.e", Apply: noApply[string], Cap: 4})
	return f
}

// fill makes every entry hold a value, A first because Set empties its readers.
func (f five) fill(g *graphcache.Graph) {
	f.a.Set(g, "a")
	f.b.Set(g, "b")
	f.c.Set(g, idA, "c")
	f.d.Set(g, idA, "d")
	f.e.Set(g, idA, "e")
}

// held names the entries that hold a value, in order.
func (f five) held(g *graphcache.Graph) string {
	var out []byte
	if _, ok, _ := f.a.Get(g); ok {
		out = append(out, 'A')
	}
	if _, ok, _ := f.b.Get(g); ok {
		out = append(out, 'B')
	}
	for i, k := range []graphcache.Keyed[string]{f.c, f.d, f.e} {
		if _, ok := k.Peek(g, idA); ok {
			out = append(out, "CDE"[i])
		}
	}
	return string(out)
}

func TestEntryDerivedFrom(t *testing.T) {
	defer graphcache.SnapshotRegistryForTest()()
	f := registerFive()
	g := graphcachetest.OpenCatalog(t).Graph()
	tests := []struct {
		name string
		do   func()
		want string
	}{
		{"A.Set empties B and C", func() { f.a.Set(g, "v") }, "ADE"},
		{"B.Drop empties C", func() { f.b.Drop(g) }, "ADE"},
		{"a vocabulary commit empties D", func() {
			if err := g.OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListPropertyVocabulary}}}); err != nil {
				t.Fatal(err)
			}
		}, "ABCE"},
		{"a Source vocabulary commit empties nothing", func() {
			if err := g.OnCommit(effects.Commit{Effects: effects.Set{Lists: []effects.List{effects.ListSourceVocabulary}}}); err != nil {
				t.Fatal(err)
			}
		}, "ABCDE"},
		{"Drop leaves the catalog entry", g.Drop, "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.fill(g)
			if got := f.held(g); got != "ABCDE" {
				t.Fatalf("held = %q after fill", got)
			}
			tt.do()
			if got := f.held(g); got != tt.want {
				t.Fatalf("held = %q, want %q", got, tt.want)
			}
		})
	}
	t.Run("A.Set(v) keeps v", func(t *testing.T) {
		f.fill(g)
		f.a.Set(g, "v")
		if v, ok, _ := f.a.Get(g); !ok || v != "v" {
			t.Fatalf("A = %q, %v", v, ok)
		}
	})
}

// TestKeyedEvictsLeastRecentlyUsed: a keyed entry at its Cap drops the key used
// least recently; a Get hit, a Peek, a Set and a store are uses; a Check and
// KeyCounts are not.
func TestKeyedEvictsLeastRecentlyUsed(t *testing.T) {
	defer graphcache.SnapshotRegistryForTest()()
	checked := 0
	k := graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{
		Name: "t.lru", Apply: noApply[string], Cap: 2,
		Check: func(*graphcache.Graph, []byte, string) error { checked++; return nil },
	})
	g := graphcachetest.OpenCatalog(t).Graph()
	fill := func(v string) func() (string, error) { return func() (string, error) { return v, nil } }
	held := func(ids ...[]byte) string {
		t.Helper()
		var out []byte
		for _, id := range ids {
			if _, ok := k.Peek(g, id); ok {
				out = append(out, 'A'+id[0]-1)
			} else {
				out = append(out, '-')
			}
		}
		return string(out)
	}
	cases := []struct {
		name string
		run  func()
		want string // idA, idB, idC held
	}{
		{"a Get hit is a use", func() {
			k.Set(g, idA, "1")
			k.Set(g, idB, "2")
			if v, err := k.Get(g, idA, fill("x")); err != nil || v != "1" {
				t.Fatalf("Get hit = %q, %v", v, err)
			}
			k.Set(g, idC, "3")
		}, "A-C"},
		{"a Peek is a use", func() {
			k.Set(g, idA, "1")
			k.Set(g, idB, "2")
			k.Peek(g, idA)
			k.Set(g, idC, "3")
		}, "A-C"},
		{"a Set of a held key is a use", func() {
			k.Set(g, idA, "1")
			k.Set(g, idB, "2")
			k.Set(g, idA, "again")
			k.Set(g, idC, "3")
			if v, _ := k.Peek(g, idA); v != "again" {
				t.Errorf("a held key = %q, want it replaced", v)
			}
		}, "A-C"},
		{"a Get fill stores and evicts", func() {
			k.Set(g, idA, "1")
			k.Set(g, idB, "2")
			if v, err := k.Get(g, idC, fill("filled")); err != nil || v != "filled" {
				t.Fatalf("Get fill = %q, %v", v, err)
			}
		}, "-BC"},
		{"GetMany touches held ids, then stores in the order asked", func() {
			k.Set(g, idA, "1")
			got, err := k.GetMany(g, [][]byte{idA, idB, idC}, func(missing [][]byte) (map[string]string, error) {
				if len(missing) != 2 || !bytes.Equal(missing[0], idB) || !bytes.Equal(missing[1], idC) {
					t.Errorf("GetMany asked for %x, want B then C", missing)
				}
				return map[string]string{string(idB): "b", string(idC): "c", string(idA[:1]): "stray"}, nil
			})
			if err != nil || len(got) != 3 || got[string(idA)] != "1" || got[string(idB)] != "b" || got[string(idC)] != "c" {
				t.Errorf("GetMany = %v, %v", got, err)
			}
		}, "-BC"},
		{"a fill error stores nothing", func() {
			k.Set(g, idA, "1")
			boom := errors.New("boom")
			if _, err := k.GetMany(g, [][]byte{idB}, func([][]byte) (map[string]string, error) { return nil, boom }); !errors.Is(err, boom) {
				t.Errorf("GetMany error = %v, want boom", err)
			}
			if _, err := k.Get(g, idB, func() (string, error) { return "", boom }); !errors.Is(err, boom) {
				t.Errorf("Get error = %v, want boom", err)
			}
		}, "A--"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k.Clear(g)
			tc.run()
			if got := held(idA, idB, idC); got != tc.want {
				t.Errorf("held = %q, want %q", got, tc.want)
			}
			if n := g.KeyCounts()["t.lru"]; n > 2 {
				t.Errorf("KeyCounts = %d, want at most the Cap", n)
			}
		})
	}

	t.Run("a Check and KeyCounts are no use", func(t *testing.T) {
		// 64 keys, so a Check that touched keys in map order would leave key 1
		// the least recently used only one time in 64.
		const n = 64
		id := func(i int) []byte { return bytes.Repeat([]byte{byte(i)}, 16) }
		k.Clear(g)
		k.SetCapForTest(g, n)
		defer k.SetCapForTest(g, 0)
		for i := 1; i <= n; i++ {
			k.Set(g, id(i), "v")
		}
		checked = 0
		if err := g.Verify(); err != nil {
			t.Fatal(err)
		}
		if checked == 0 {
			t.Error("Verify ran no Check")
		}
		if got := g.KeyCounts()["t.lru"]; got != n {
			t.Errorf("KeyCounts = %d, want %d", got, n)
		}
		k.Set(g, id(n+1), "v")
		if _, ok := k.Peek(g, id(1)); ok {
			t.Error("key 1 is held, want it evicted as the least recently used")
		}
		for i := 2; i <= n+1; i++ {
			if _, ok := k.Peek(g, id(i)); !ok {
				t.Errorf("key %d is not held", i)
			}
		}
	})

	t.Run("SetCapForTest evicts at once and survives Drop", func(t *testing.T) {
		k.Clear(g)
		k.Set(g, idA, "1")
		k.Set(g, idB, "2")
		k.SetCapForTest(g, 1)
		if got := held(idA, idB); got != "-B" {
			t.Errorf("after SetCapForTest(1) held = %q, want -B", got)
		}
		g.Drop()
		k.Set(g, idA, "1")
		k.Set(g, idB, "2")
		if n := g.KeyCounts()["t.lru"]; n != 1 {
			t.Errorf("KeyCounts after Drop = %d, want the overridden Cap 1", n)
		}
		k.SetCapForTest(g, 0)
		k.Set(g, idC, "3")
		if n := g.KeyCounts()["t.lru"]; n != 2 {
			t.Errorf("KeyCounts after restoring the Cap = %d, want 2", n)
		}
	})
}

func TestEntryBuild(t *testing.T) {
	defer graphcache.SnapshotRegistryForTest()()
	builds, catalogBuilds, graphBuilds := 0, 0, 0
	fail := false
	w := graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.w", Build: func(*graphcache.Graph) (string, error) {
		builds++
		if fail {
			return "", errors.New("boom")
		}
		return "built", nil
	}})
	r := graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.r", DerivedFrom: []graphcache.Source{w}})
	cat := graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.catalog", Lifetime: graphcache.LifeCatalog, Build: func(g *graphcache.Graph) (string, error) {
		catalogBuilds++
		var one int
		if err := g.QueryRow(`SELECT 1`).Scan(&one); err != nil {
			return "", err
		}
		return "catalog", nil
	}})
	gr := graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.graph", Build: func(*graphcache.Graph) (string, error) {
		graphBuilds++
		return "graph", nil
	}})

	c := graphcachetest.OpenCatalog(t)
	g := c.Graph()
	r.Set(g, "reader")
	if _, ok := w.Peek(g); ok || builds != 0 {
		t.Fatalf("Peek on an empty entry with a Build returned %v after %d builds, want false and none", ok, builds)
	}
	for i := 0; i < 2; i++ {
		if v, ok, err := w.Get(g); err != nil || !ok || v != "built" {
			t.Fatalf("Get = %q, %v, %v", v, ok, err)
		}
	}
	if builds != 1 {
		t.Errorf("Build ran %d times, want 1", builds)
	}
	if v, ok := w.Peek(g); !ok || v != "built" {
		t.Errorf("Peek of a held entry = %q, %v, want built and true", v, ok)
	}
	if _, ok, _ := r.Get(g); !ok {
		t.Error("a first fill by Build emptied a reader")
	}
	w.Set(g, "set")
	if _, ok, _ := r.Get(g); ok {
		t.Error("Set left a reader")
	}

	w.Drop(g)
	fail = true
	if _, ok, err := w.Get(g); err == nil || ok {
		t.Errorf("a failing Build returned %v, %v", ok, err)
	}
	fail = false
	if _, ok, _ := w.Get(g); !ok {
		t.Error("a Build after a failed one did not run")
	}

	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("a write's graph", func(t *testing.T) {
		over := graphcache.Over(db)
		if v, ok, err := cat.Get(over); err != nil || !ok || v != "catalog" {
			t.Fatalf("catalog Get = %q, %v, %v", v, ok, err)
		}
		if _, _, err := cat.Get(over); err != nil || catalogBuilds != 1 {
			t.Fatalf("catalog Build ran %d times, want 1 (%v)", catalogBuilds, err)
		}
		if _, ok, err := gr.Get(over); err != nil || ok || graphBuilds != 0 {
			t.Fatalf("graph Get = %v, %v after %d builds, want false with no Build", ok, err, graphBuilds)
		}
		gr.Set(over, "stored")
		if _, ok, _ := gr.Get(over); ok {
			t.Error("Set of a graph entry stored on a write's graph")
		}
	})
}

type applied struct {
	roots [][]byte
	set   effects.Set
}

func TestEntryApply(t *testing.T) {
	defer graphcache.SnapshotRegistryForTest()()
	var calls []applied
	k := graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{Name: "t.apply", Cap: 4,
		Apply: func(_ *graphcache.Graph, _ graphcache.Keyed[string], roots [][]byte, set effects.Set) error {
			calls = append(calls, applied{roots: roots, set: set})
			return nil
		}})
	c := graphcachetest.OpenCatalog(t)
	canon := graphcachetest.NewCanon(t, c)
	g := c.Graph()
	person, other := canon.Person(), canon.Person()
	event, joins := canon.Birth(person, nil, nil)
	canon.Commit(joins...)
	g.Drop()

	has := func(call applied, ids ...[]byte) bool {
		for _, id := range ids {
			if !slices.ContainsFunc(call.roots, func(r []byte) bool { return bytes.Equal(r, id) }) {
				return false
			}
		}
		return true
	}
	hold := func() { k.Set(g, idC, "x") }
	last := func(t *testing.T) applied {
		t.Helper()
		if len(calls) != 1 {
			t.Fatalf("Apply ran %d times, want 1", len(calls))
		}
		call := calls[0]
		calls = nil
		return call
	}

	t.Run("a first load", func(t *testing.T) {
		hold()
		if _, err := g.Node(event); err != nil {
			t.Fatal(err)
		}
		if len(calls) != 0 {
			t.Fatalf("a first load applied %d times", len(calls))
		}
	})
	t.Run("an association commit", func(t *testing.T) {
		hold()
		canon.Commit(joins[0])
		if call := last(t); !has(call, person, event) {
			t.Errorf("roots %x lack the association's ends", call.roots)
		}
	})
	t.Run("a merge", func(t *testing.T) {
		hold()
		canon.Merge(person, other)
		canon.Commit(person, other)
		if call := last(t); !has(call, person, event) {
			t.Errorf("roots %x lack the merged handle and its neighbor", call.roots)
		}
	})
	t.Run("a delete of a loaded node", func(t *testing.T) {
		gone := canon.Person()
		ev, links := canon.Birth(gone, nil, nil)
		canon.Commit(links...)
		calls = nil
		if _, err := g.Node(gone); err != nil {
			t.Fatal(err)
		}
		for _, l := range links {
			canon.Remove(l)
		}
		canon.Remove(gone)
		hold()
		canon.Commit(append([][]byte{gone}, links...)...)
		if call := last(t); !has(call, gone, ev) {
			t.Errorf("roots %x lack the deleted node and its link neighbor", call.roots)
		}
	})
	t.Run("a Source-only commit", func(t *testing.T) {
		hold()
		src := [][]byte{idA}
		if err := g.OnCommit(effects.Commit{Effects: effects.Set{SourceScopes: src}}); err != nil {
			t.Fatal(err)
		}
		call := last(t)
		if len(call.roots) != 0 || len(call.set.SourceScopes) != 1 {
			t.Errorf("roots %x, set %+v, want no roots and the commit's set", call.roots, call.set)
		}
	})
}

func TestEntryVerify(t *testing.T) {
	t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")
	defer graphcache.SnapshotRegistryForTest()()
	reject := errors.New("rejected")
	k := graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{Name: "t.vk", Cap: 2, Apply: noApply[string],
		Check: func(*graphcache.Graph, []byte, string) error { return reject }})
	w := graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.vw",
		Check: func(*graphcache.Graph, string) error { return reject }})
	g := graphcachetest.OpenCatalog(t).Graph()
	if err := g.OnCommit(effects.Commit{}); err != nil {
		t.Fatalf("a commit over empty entries failed: %v", err)
	}
	k.Set(g, idA, "v")
	if err := g.OnCommit(effects.Commit{}); !errors.Is(err, reject) {
		t.Errorf("keyed Check: error = %v, want it returned", err)
	}
	k.Clear(g)
	w.Set(g, "v")
	if err := g.OnCommit(effects.Commit{}); !errors.Is(err, reject) {
		t.Errorf("whole Check: error = %v, want it returned", err)
	}
}

func TestRegisterRejects(t *testing.T) {
	defer graphcache.SnapshotRegistryForTest()()
	graph := graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.graphwhole"})
	graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.dup"})
	tests := []struct {
		name string
		do   func()
	}{
		{"keyed with no Cap", func() {
			graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{Name: "t.nocap", Apply: noApply[string]})
		}},
		{"keyed with neither Apply nor Inputs", func() {
			defer func() {
				r := recover()
				if msg, _ := r.(string); !strings.Contains(msg, "needs an Apply or Inputs") {
					t.Errorf("panic %v, want one naming Apply or Inputs", r)
				}
				panic(r)
			}()
			graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{Name: "t.noapply", Cap: 1})
		}},
		{"derived from an unregistered entry", func() {
			graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.unreg", DerivedFrom: []graphcache.Source{graphcache.Whole[int]{}}})
		}},
		{"catalog entry derived from a graph entry", func() {
			graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.cg", Lifetime: graphcache.LifeCatalog, DerivedFrom: []graphcache.Source{graph}})
		}},
		{"catalog entry derived from the vocabulary", func() {
			graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.cv", Lifetime: graphcache.LifeCatalog, DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}}})
		}},
		{"a repeated Name", func() {
			graphcache.RegisterWhole(graphcache.WholeEntry[string]{Name: "t.dup"})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("registration did not panic")
				}
			}()
			tt.do()
		})
	}
}

// inputsOfKeys is the Inputs of a test entry whose value names the keys it was
// built from, by id.
func inputsOfKeys(from map[string][]graphcache.InputKey) func(id []byte, v string) []graphcache.InputKey {
	return func(id []byte, v string) []graphcache.InputKey { return from[v] }
}

func handlesCommit(t *testing.T, g *graphcache.Graph, ids ...[]byte) {
	t.Helper()
	if err := g.OnCommit(effects.Commit{Effects: effects.Set{Handles: ids}}); err != nil {
		t.Fatal(err)
	}
}

func TestKeyedInputs(t *testing.T) {
	defer graphcache.SnapshotRegistryForTest()()
	idX := bytes.Repeat([]byte{0xa1}, 16)
	idY := bytes.Repeat([]byte{0xa2}, 16)
	idZ := bytes.Repeat([]byte{0xa3}, 16)
	k := func(id []byte) graphcache.InputKey { return graphcache.InputKey{ID: id} }
	k1 := graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{
		Name: "t.inputs", Cap: 4,
		Inputs: inputsOfKeys(map[string][]graphcache.InputKey{
			"a": {k(idA), k(idX)}, "b": {k(idB)}, "a2": {k(idA), k(idZ)},
		}),
	})
	held := func(g *graphcache.Graph) string {
		var out string
		for i, id := range [][]byte{idA, idB} {
			if _, ok := k1.Peek(g, id); ok {
				out += "AB"[i : i+1]
			}
		}
		return out
	}
	fill := func(g *graphcache.Graph) {
		k1.Set(g, idA, "a")
		k1.Set(g, idB, "b")
	}
	tests := []struct {
		name    string
		handles [][]byte
		want    string
	}{
		{"a written input drops its key", [][]byte{idX}, "B"},
		{"an unrelated handle drops nothing", [][]byte{idY}, "AB"},
		{"a key's own handle drops it", [][]byte{idB}, "A"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := graphcachetest.OpenCatalog(t).Graph()
			fill(g)
			handlesCommit(t, g, tc.handles...)
			if got := held(g); got != tc.want {
				t.Fatalf("held %q, want %q", got, tc.want)
			}
		})
	}
	t.Run("an overwrite replaces the inputs", func(t *testing.T) {
		g := graphcachetest.OpenCatalog(t).Graph()
		fill(g)
		k1.Set(g, idA, "a2")
		handlesCommit(t, g, idX)
		if got := held(g); got != "AB" {
			t.Fatalf("held %q after the old input moved, want AB", got)
		}
		handlesCommit(t, g, idZ)
		if got := held(g); got != "B" {
			t.Fatalf("held %q after the new input moved, want B", got)
		}
	})
	t.Run("the index follows every mutation", func(t *testing.T) {
		g := graphcachetest.OpenCatalog(t).Graph()
		fill(g)
		if n := graphcache.InputIndexSizeForTest(g, k1); n != 3 {
			t.Fatalf("index holds %d pairs, want 3", n)
		}
		k1.Set(g, idA, "a2")
		if n := graphcache.InputIndexSizeForTest(g, k1); n != 3 {
			t.Fatalf("after an overwrite %d pairs, want 3", n)
		}
		k1.SetCapForTest(g, 1)
		if n := graphcache.InputIndexSizeForTest(g, k1); n != 2 {
			t.Fatalf("after eviction %d pairs, want 2", n)
		}
		fill(g)
		k1.Drop(g, [][]byte{idA, idB})
		if n := graphcache.InputIndexSizeForTest(g, k1); n != 0 {
			t.Fatalf("after Drop %d pairs, want 0", n)
		}
		k1.SetCapForTest(g, 0)
		fill(g)
		k1.Clear(g)
		if n := graphcache.InputIndexSizeForTest(g, k1); n != 0 {
			t.Fatalf("after Clear %d pairs, want 0", n)
		}
	})
}

func TestKeyedInputLinks(t *testing.T) {
	defer graphcache.SnapshotRegistryForTest()()
	c := graphcachetest.OpenCatalog(t)
	g := c.Graph()
	canon := graphcachetest.NewCanon(t, c)
	event, place := canon.Event("birth"), canon.Entity("place")
	person := canon.Person()
	canon.Commit(event, place, person)

	var link []byte
	kv := graphcache.RegisterKeyed(graphcache.KeyedEntry[string]{
		Name: "t.links", Cap: 4,
		Inputs: func(id []byte, v string) []graphcache.InputKey {
			if bytes.Equal(id, event) {
				return []graphcache.InputKey{{ID: event, Bridge: "location", Endpoint: "event"}}
			}
			keys := []graphcache.InputKey{{ID: place, Bridge: "location", Endpoint: "place"}}
			if link != nil {
				keys = append(keys, graphcache.InputKey{ID: link})
			}
			return keys
		},
	})
	set := func() {
		kv.Set(g, event, "e")
		kv.Set(g, place, "p")
	}
	// held names the keys still held: "e" for the event's, "p" for the place's.
	held := func() string {
		var out string
		for _, h := range []struct {
			id   []byte
			name string
		}{{event, "e"}, {place, "p"}} {
			if _, ok := kv.Peek(g, h.id); ok {
				out += h.name
			}
		}
		return out
	}

	set()
	link = canon.Location(event, place)
	canon.Commit(link)
	if got := held(); got != "" {
		t.Fatalf("a new location left %q held, want none", got)
	}
	set()
	join := canon.Join(person, event, "subject")
	canon.Commit(join)
	if got := held(); got != "ep" {
		t.Fatalf("a participation left %q held, want ep", got)
	}
	canon.Remove(link)
	canon.Commit(link)
	if got := held(); got != "e" { // the event's inputs do not list the association
		t.Fatalf("a removed location left %q held, want e (the event)", got)
	}
	set()
	second := canon.Event("birth")
	canon.Commit(second, canon.Location(second, place))
	set()
	canon.Merge(second, event)
	canon.Commit(second, event)
	if got := held(); got != "e" {
		t.Fatalf("merging a handle located at the place left %q held, want e (the event)", got)
	}
}
