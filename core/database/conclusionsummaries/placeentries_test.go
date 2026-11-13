package conclusionsummaries_test

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// chainWorld is Town part_of City part_of Region, promoted, with three births
// located at City and one at Town.
type chainWorld struct {
	town, city, region []byte
	townRel, cityRel   subjects.Subject
	events             [][]byte
}

func (f *fixture) chainWorld() *chainWorld {
	f.T.Helper()
	town := f.placeSubject("Town", nil, nil, 0, 0)
	city := f.placeSubject("City", nil, nil, 2, 0)
	region := f.placeSubject("Region", nil, nil, 4, 0)
	w := &chainWorld{
		townRel: f.placeRel(town, city, propertyterms.KeyPartOf, nil, nil),
		cityRel: f.placeRel(city, region, propertyterms.KeyPartOf, nil, nil),
	}
	p := f.lifePerson()
	var evs []subjects.Subject
	for i, at := range []subjects.Subject{city, city, city, town} {
		evs = append(evs, f.lifeEvent(p, "birth", "date", 1800+i, int64(2+2*i), at))
	}
	w.town, w.city, w.region = f.promoteSubject(town), f.promoteSubject(city), f.promoteSubject(region)
	f.promoteSubject(p)
	for _, e := range evs {
		w.events = append(w.events, f.promoteSubject(e))
	}
	return w
}

func TestPlaceEntryInputsMatchReads(t *testing.T) {
	f := newFixture(t)
	w := f.chainWorld()
	unknown := bytes.Repeat([]byte{0xee}, 16)
	g := f.C.Graph()
	g.Drop()
	must(t, conclusionsummaries.FillPlaceEntriesForTest(g, w.town, w.city, w.region, unknown))
	db, err := f.C.DB()
	must(t, err)
	// reads is what a scope records reading id and following hopList.
	reads := func(id []byte, hopList ...hops.Hop) []graphcache.InputKey {
		in := graphcache.Over(db).Inputs(nil)
		n, err := in.Node(id)
		if err != nil {
			t.Fatal(err)
		}
		if n == nil {
			return in.Keys() // a missing place follows nothing
		}
		for _, hop := range hopList {
			links, err := in.Follow(id, hop)
			if err != nil {
				t.Fatal(err)
			}
			for _, l := range links {
				if _, err := in.Node(l.Association); err != nil {
					t.Fatal(err)
				}
			}
		}
		return in.Keys()
	}
	for name, id := range map[string][]byte{"town": w.town, "city": w.city, "region": w.region, "unknown": unknown} {
		t.Run(name, func(t *testing.T) {
			held, _, inputs, dropBy := conclusionsummaries.PlaceEntryForTest(g, id)
			if !held {
				t.Fatal("the entry is not held")
			}
			same := func(what string, got, want []graphcache.InputKey) {
				t.Helper()
				if !slices.EqualFunc(got, want, graphcache.InputKey.Equal) {
					t.Errorf("%s: got %+v, want %+v", what, got, want)
				}
			}
			same("inputs", inputs, reads(id, hops.ParentsOfPlace))
			same("drop inputs", dropBy, reads(id, hops.ParentsOfPlace, hops.PartsOfPlace, hops.SuccessorsOfPlace, hops.PredecessorsOfPlace))
		})
	}
}

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
