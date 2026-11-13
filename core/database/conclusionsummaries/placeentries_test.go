package conclusionsummaries_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// meshWorld is the crafted part_of mesh BenchmarkPlaceMesh reads: a hub with
// 20,000 parts that is the lowest place of a chain of eight, and one event
// located at the hub per 10 parts.
type meshWorld struct {
	c    *database.Catalog
	k    *graphcachetest.Canon
	hub  []byte
	wide int
}

func newMeshWorld(tb testing.TB, parts int) *meshWorld {
	tb.Helper()
	c := graphcachetest.OpenCatalog(tb)
	w := &meshWorld{c: c, k: graphcachetest.NewCanon(tb, c), wide: parts}
	w.hub = w.k.Entity("place")
	w.k.Bulk(func() {
		below := w.hub
		for i := 1; i <= 7; i++ {
			whole := w.k.Entity("place")
			w.k.PartOf(below, whole)
			below = whole
		}
		for i := 0; i < parts; i++ {
			w.k.PartOf(w.k.Entity("place"), w.hub)
			if i%10 == 0 {
				w.k.Location(w.k.Event(vocabkey.EventBirth), w.hub)
			}
		}
	})
	w.k.Commit(w.hub)
	c.Graph().Drop()
	return w
}

// BenchmarkPlaceMesh reports, in milliseconds, what the S16-18 plan measures on
// a crafted part_of mesh: the Places list cold and warm, the hub's page
// relationships cold, the list after the hub is committed (what a write of its
// names commits), and after a new location at the hub is committed. Reported,
// not gated. It ignores b.N.
func BenchmarkPlaceMesh(b *testing.B) {
	w := newMeshWorld(b, 20000)
	ms := func(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }
	list := func() time.Duration {
		return databasetest.Since(func() {
			if _, err := conclusionsummaries.ListPlaces(w.c); err != nil {
				b.Fatal(err)
			}
		})
	}
	b.Run("list-cold", func(b *testing.B) {
		w.c.Graph().Drop()
		b.ReportMetric(ms(list()), "ms")
	})
	b.Run("list-warm", func(b *testing.B) {
		b.ReportMetric(ms(list()), "ms")
	})
	b.Run("page-cold", func(b *testing.B) {
		w.c.Graph().Drop()
		rows, err := conclusionsummaries.PlacesByIDs(w.c, [][]byte{w.hub})
		if err != nil || len(rows) != 1 {
			b.Fatalf("PlacesByIDs: %d rows, %v", len(rows), err)
		}
		w.c.Graph().Drop()
		b.ReportMetric(ms(databasetest.Since(func() {
			if err := conclusionsummaries.AttachPlaceRelationships(w.c, &rows[0]); err != nil {
				b.Fatal(err)
			}
		})), "ms")
	})
	b.Run("list-after-hub-commit", func(b *testing.B) {
		list()
		w.k.Commit(w.hub)
		b.ReportMetric(ms(list()), "ms")
	})
	b.Run("list-after-location", func(b *testing.B) {
		list()
		loc := w.k.Location(w.k.Event(vocabkey.EventBirth), w.hub)
		w.k.Commit(loc)
		b.ReportMetric(ms(list()), "ms")
	})
	b.Log(fmt.Sprintf("hub with %d parts", w.wide))
}
