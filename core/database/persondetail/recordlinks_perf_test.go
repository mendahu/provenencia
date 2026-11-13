package persondetail_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/recordlinks"
)

// Budgets of S13-27 D-15, on a person in 50 household graphs (the map's Scale).
const (
	budgetLinksCold   = 1500 * time.Millisecond
	budgetLinksRepeat = 300 * time.Millisecond
	budgetLinksPast   = 1500 * time.Millisecond // 65 graphs reload on every read
	// budgetLinks80: S16-13 D-15, a repeat read at 80 graphs. The planning base
	// read in 113 ms; 1.25 times that rounds up to 200 ms, below budgetLinksPast,
	// so the budget is budgetLinksPast.
	budgetLinks80 = budgetLinksPast // measured: S16-13 planning base
)

// households files n Sources "Household 1" to "Household n", each with the son
// "Thomas Smith", the father "William Smith", the mother "Mary Ann Jones" and
// the sister "Ann Smith", each cited with its name only, and four parent
// links (father and mother to the son, father and mother to the sister). Each
// graph's four Subjects are promoted onto the same four entities, so each pair
// of graphs has three strong lines: father, mother, sister (1,225 pairs at
// 50). It returns the son's entity. householdGraphs in core/recordlinks builds
// the same shape in memory and asserts the same numbers.
func households(f *fixture, n int) []byte {
	f.T.Helper()
	var son, father, mother, sister []byte
	for i := 1; i <= n; i++ {
		src := f.source(fmt.Sprintf("Household %d", i))
		art := f.arts[string(src.ID)]
		s := f.Person(src, art, "Thomas Smith")
		fa := f.Person(src, art, "William Smith")
		m := f.Person(src, art, "Mary Ann Jones")
		si := f.Person(src, art, "Ann Smith")
		f.Relationship(src, art, fa, s, "parent")
		f.Relationship(src, art, m, s, "parent")
		f.Relationship(src, art, fa, si, "parent")
		f.Relationship(src, art, m, si, "parent")
		if i == 1 {
			son, father, mother, sister = f.promote(s, nil).ID, f.promote(fa, nil).ID, f.promote(m, nil).ID, f.promote(si, nil).ID
			continue
		}
		f.promote(s, son)
		f.promote(fa, father)
		f.promote(m, mother)
		f.promote(si, sister)
	}
	return son
}

// threeStrong checks the read has n graphs and a pair of three strong lines
// for each of them.
func threeStrong(t *testing.T, rl persondetail.PersonRecordLinks, n int) {
	t.Helper()
	if len(rl.Graphs) != n || len(rl.Pairs) != n*(n-1)/2 {
		t.Fatalf("%d graphs and %d pairs, want %d and %d", len(rl.Graphs), len(rl.Pairs), n, n*(n-1)/2)
	}
	for _, p := range rl.Pairs {
		if len(p.Lines) != 3 {
			t.Fatalf("pair %d-%d has %d lines, want 3", p.A, p.B, len(p.Lines))
		}
		for _, l := range p.Lines {
			if l.Strength != recordlinks.StrengthStrong {
				t.Fatalf("pair %d-%d line %x: %d facts, %s", p.A, p.B, l.Entity, l.Facts, l.Strength)
			}
		}
	}
}

func TestRecordLinksBudgets(t *testing.T) {
	f := newFixture(t)
	son := households(f, 50)
	read := func() persondetail.PersonRecordLinks { return f.recordLinks(son) }

	t.Run("fixture", func(t *testing.T) { threeStrong(t, read(), 50) })

	t.Run("cold", func(t *testing.T) {
		d := catalogsessiontest.Cold(t, f.C, 2, func() { read() })
		databasetest.WithinBudget(t, "cold read at 50 graphs", d, budgetLinksCold)
	})

	t.Run("repeat", func(t *testing.T) {
		read()
		d := databasetest.Fastest(5, func() time.Duration {
			runtime.GC()
			return databasetest.Since(func() { read() })
		})
		databasetest.WithinBudget(t, "repeat read at 50 graphs", d, budgetLinksRepeat)
	})

	t.Run("past the cache", func(t *testing.T) {
		g := newFixture(t)
		son := households(g, 65)
		first := g.recordLinks(son)
		g.C.Graph().Drop()
		cold := g.recordLinks(son)
		if !reflect.DeepEqual(first.Pairs, cold.Pairs) || !reflect.DeepEqual(graphIDs(first), graphIDs(cold)) {
			t.Fatal("the read at 65 graphs differs from the read after dropping the cache")
		}
		threeStrong(t, first, 65)
		d := databasetest.Fastest(2, func() time.Duration {
			runtime.GC()
			return databasetest.Since(func() { g.recordLinks(son) })
		})
		databasetest.WithinBudget(t, "repeat read at 65 graphs", d, budgetLinksPast)
	})

	t.Run("80 graphs", func(t *testing.T) {
		g := newFixture(t)
		son := households(g, 80)
		threeStrong(t, g.recordLinks(son), 80)
		d := databasetest.Fastest(2, func() time.Duration {
			runtime.GC()
			return databasetest.Since(func() { g.recordLinks(son) })
		})
		databasetest.WithinBudget(t, "repeat read at 80 graphs", d, budgetLinks80)
	})
}

// graphIDs is the id of each graph of a read, in the order the read lists them.
func graphIDs(rl persondetail.PersonRecordLinks) [][]byte {
	out := make([][]byte, len(rl.Graphs))
	for i, g := range rl.Graphs {
		out[i] = g.Graph.GraphID
	}
	return out
}
