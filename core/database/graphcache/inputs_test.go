package graphcache_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func keyOf(id []byte, bridge, endpoint string) graphcache.InputKey {
	return graphcache.InputKey{ID: id, Bridge: bridge, Endpoint: endpoint}
}

func sameKeys(t *testing.T, what string, got, want []graphcache.InputKey) {
	t.Helper()
	graphcache.SortInputKeys(want)
	if !slices.EqualFunc(got, want, graphcache.InputKey.Equal) {
		t.Fatalf("%s: got %+v, want %+v", what, got, want)
	}
}

func TestInputScope(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	canon := graphcachetest.NewCanon(t, c)
	person, other := canon.Person(), canon.Person()
	birth := canon.Event(vocabkey.EventBirth)
	subject := canon.Join(person, birth, "subject")
	witness := canon.Join(other, birth, "witness")
	canon.Commit(person, other, birth, subject, witness)
	g := c.Graph()
	missing := bytes.Repeat([]byte{0xee}, 16)

	t.Run("the recorder hears Add", func(t *testing.T) {
		var heard []graphcache.InputKey
		g.SetInputRecorderForTest(func(k graphcache.InputKey) { heard = append(heard, k) })
		defer g.SetInputRecorderForTest(nil)
		k := keyOf(other, "", "")
		g.Inputs(person).Add(k)
		g.NoInputs().Add(k)
		if len(heard) != 2 || !heard[0].Equal(k) || !heard[1].Equal(k) {
			t.Fatalf("heard %+v, want the key twice", heard)
		}
		heard = nil
		if _, err := g.Inputs(person).Node(other); err != nil {
			t.Fatal(err)
		}
		if len(heard) != 1 {
			t.Fatalf("Node was heard %d times, want once", len(heard))
		}
	})
	t.Run("Node records the node's values", func(t *testing.T) {
		in := g.Inputs(person)
		if _, err := in.Node(other); err != nil {
			t.Fatal(err)
		}
		sameKeys(t, "other", in.Keys(), []graphcache.InputKey{keyOf(other, "", "")})
	})
	t.Run("Node of self records nothing", func(t *testing.T) {
		in := g.Inputs(person)
		if _, err := in.Node(person); err != nil {
			t.Fatal(err)
		}
		sameKeys(t, "self", in.Keys(), nil)
	})
	t.Run("Node of a missing id records it", func(t *testing.T) {
		in := g.Inputs(person)
		n, err := in.Node(missing)
		if err != nil || n != nil {
			t.Fatalf("node %v err %v", n, err)
		}
		sameKeys(t, "missing", in.Keys(), []graphcache.InputKey{keyOf(missing, "", "")})
	})
	t.Run("Follow records the bridge end and each association", func(t *testing.T) {
		in := g.Inputs(birth)
		links, err := in.Follow(person, hops.EventsOfSubject)
		if err != nil || len(links) != 1 {
			t.Fatalf("links %+v err %v", links, err)
		}
		sameKeys(t, "follow", in.Keys(), []graphcache.InputKey{
			keyOf(person, "participation", "person"), keyOf(subject, "", ""),
		})
	})
	t.Run("Follow of the dependent's own bridge end is kept", func(t *testing.T) {
		in := g.Inputs(person)
		if _, err := in.Follow(person, hops.EventsOfSubject); err != nil {
			t.Fatal(err)
		}
		sameKeys(t, "own end", in.Keys(), []graphcache.InputKey{
			keyOf(person, "participation", "person"), keyOf(subject, "", ""),
		})
	})
	t.Run("a filtered-out link is covered by the bridge end", func(t *testing.T) {
		in := g.Inputs(birth)
		links, err := in.Follow(other, hops.EventsOfSubject)
		if err != nil || len(links) != 0 {
			t.Fatalf("links %+v err %v", links, err)
		}
		sameKeys(t, "witness", in.Keys(), []graphcache.InputKey{keyOf(other, "participation", "person")})
	})
	t.Run("Nodes and NoInputs record nothing", func(t *testing.T) {
		in := g.Inputs(person)
		if err := in.Nodes([][]byte{other, birth}); err != nil {
			t.Fatal(err)
		}
		sameKeys(t, "nodes", in.Keys(), nil)
		none := g.NoInputs()
		if _, err := none.Node(other); err != nil {
			t.Fatal(err)
		}
		if _, err := none.Follow(person, hops.EventsOfSubject); err != nil {
			t.Fatal(err)
		}
		none.Add(keyOf(other, "", ""))
		sameKeys(t, "no inputs", none.Keys(), nil)
	})
	t.Run("Inputs(nil) keeps self keys", func(t *testing.T) {
		in := g.Inputs(nil)
		if _, err := in.Node(person); err != nil {
			t.Fatal(err)
		}
		sameKeys(t, "nil dependent", in.Keys(), []graphcache.InputKey{keyOf(person, "", "")})
	})
	t.Run("the recorder hears Node and Follow and no node values from Follow", func(t *testing.T) {
		rg := graphcache.Over(dbtest.MustDB(t, c))
		var heard []graphcache.InputKey
		rg.SetInputRecorderForTest(func(k graphcache.InputKey) { heard = append(heard, k) })
		if _, err := rg.Follow(person, hops.EventsOfSubject); err != nil {
			t.Fatal(err)
		}
		sameKeys(t, "follow", heard, []graphcache.InputKey{
			keyOf(person, "participation", "person"), keyOf(subject, "", ""),
		})
		heard = nil
		if _, err := rg.Node(person); err != nil {
			t.Fatal(err)
		}
		sameKeys(t, "node", heard, []graphcache.InputKey{keyOf(person, "", "")})
	})
}

func TestAssociationInputs(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	canon := graphcachetest.NewCanon(t, c)
	person, other := canon.Person(), canon.Person()
	birth := canon.Event(vocabkey.EventBirth)
	subject := canon.Join(person, birth, "subject")
	moved := canon.Join(other, birth, "witness")
	canon.Commit(person, other, birth, subject, moved)
	g := c.Graph()
	got, err := g.AssociationInputs([][]byte{subject, person})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got[string(person)]; ok || len(got) != 1 {
		t.Fatalf("a Person has no association inputs: %+v", got)
	}
	sameKeys(t, "participation", got[string(subject)], []graphcache.InputKey{
		keyOf(person, "participation", "person"), keyOf(birth, "participation", "event"),
	})

	// An end that is not rank 1 gives no key.
	db := dbtest.MustDB(t, c)
	if _, err := db.Exec(`UPDATE auto_reconciler_values SET rank = 2
		WHERE entity_id = ? AND value_entity_id = ?`, moved, other); err != nil {
		t.Fatal(err)
	}
	g.Drop()
	got, err = g.AssociationInputs([][]byte{moved})
	if err != nil {
		t.Fatal(err)
	}
	sameKeys(t, "rank 2 end", got[string(moved)], []graphcache.InputKey{keyOf(birth, "participation", "event")})
}
