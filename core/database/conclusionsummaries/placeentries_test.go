package conclusionsummaries_test

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// chainWorld is Town part_of City part_of Region, promoted, with three births
// located at City and one at Town.
type chainWorld struct {
	town, city, region []byte
	townRel, cityRel   subjects.Subject
	citySubject        subjects.Subject
	person             subjects.Subject
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
	w.citySubject = city
	p := f.lifePerson()
	w.person = p
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

// breakDates makes every start and end date stored on a link undecodable. A
// test that calls it dates no place, only links.
func (f *fixture) breakDates() {
	f.T.Helper()
	db, err := f.C.DB()
	must(f.T, err)
	res, err := db.Exec(`UPDATE auto_reconciler_values SET value_date = x'ff'
		WHERE value_date IS NOT NULL AND property_id IN (?, ?)`, f.Prop("start_date").ID, f.Prop("end_date").ID)
	must(f.T, err)
	if n, _ := res.RowsAffected(); n == 0 {
		f.T.Fatal("no stored date to break")
	}
}

// townChain reads the chain of the town, which holds the entries of all three
// places, and returns it.
func (f *fixture) townChain(w *chainWorld) []string {
	f.T.Helper()
	got, err := conclusionsummaries.ParentsAtDate(f.C, w.town, conclusionsummaries.TodayDate())
	must(f.T, err)
	return got
}

func TestChainsSkipLocationLinks(t *testing.T) {
	f := newFixture(t)
	w := f.chainWorld()
	g := f.C.Graph()
	g.Drop()
	if got := f.townChain(w); !slices.Equal(got, []string{"City", "Region"}) {
		t.Fatalf("parents %v, want City, Region", got)
	}
	for name, id := range map[string][]byte{"town": w.town, "city": w.city, "region": w.region} {
		if _, ok := g.Cached(id); ok {
			t.Errorf("the %s node was loaded to draw a chain", name)
		}
	}
	for _, e := range w.events {
		if _, ok := g.Cached(e); ok {
			t.Errorf("event %x was loaded to draw a chain", e[10:])
		}
	}

	db, err := f.C.DB()
	must(t, err)
	over := graphcache.Over(db)
	rows, err := conclusionsummaries.RowsForSearch(over, conclusionsummaries.Want{Events: [][]byte{w.events[3]}})
	must(t, err)
	row, ok := rows.Event(w.events[3])
	if !ok || len(row.Places) != 1 || !slices.Equal(row.Places[0].Names, []string{"Town"}) ||
		!slices.Equal(row.Places[0].Parents, []string{"City", "Region"}) {
		t.Fatalf("event row %+v", row.Places)
	}
	if _, ok := over.Cached(w.city); ok {
		t.Error("the write's graph loaded the city's node")
	}
}

func TestPlaceEntries(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, f *fixture, w *chainWorld)
	}{
		{"person write", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			list := func() int {
				n, err := conclusionsummaries.ListPlacesQueryCount(f.C)
				must(t, err)
				return n
			}
			list()
			warm := list()
			keys := g.KeyCounts()["conclusionsummaries.places"]
			f.named(w.person, "Bo")
			if got := g.KeyCounts()["conclusionsummaries.places"]; got != keys {
				t.Errorf("place entries %d after a person write, were %d", got, keys)
			}
			if n := list(); n > warm {
				t.Errorf("the list issued %d queries after the write, %d warm", n, warm)
			}
		}},
		{"city rename", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			f.citeToponyms(f.Artifact, w.citySubject.ID, "Big City")
			for name, want := range map[string]bool{"city": false, "town": true, "region": true} {
				id := map[string][]byte{"city": w.city, "town": w.town, "region": w.region}[name]
				if got := conclusionsummaries.HeldPlaceEntryForTest(g, id); got != want {
					t.Errorf("%s held %v, want %v", name, got, want)
				}
			}
			rows, err := conclusionsummaries.PlacesByIDs(f.C, [][]byte{w.city})
			must(t, err)
			if len(rows) != 1 || !slices.Contains(rows[0].Names, "Big City") {
				t.Fatalf("city names %+v", rows)
			}
			if got := f.townChain(w); len(got) == 0 || got[0] != rows[0].Names[0] {
				t.Errorf("chain %v starts with %q, want the city's first name %q", got, got, rows[0].Names[0])
			}
		}},
		{"remove link", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			must(t, conclusionsummaries.FillPlacePagesForTest(g, w.city))
			_, err := writes.Call(f.C, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				ch, err := subjects.Delete(tx, userID, w.townRel.ID)
				return struct{}{}, ch, err
			})
			must(t, err)
			for name, want := range map[string]bool{"town": false, "city": true, "region": true} {
				id := map[string][]byte{"city": w.city, "town": w.town, "region": w.region}[name]
				if got := conclusionsummaries.HeldPlaceEntryForTest(g, id); got != want {
					t.Errorf("%s held %v, want %v", name, got, want)
				}
			}
			if held, _ := conclusionsummaries.PlacePageForTest(g, w.city); held {
				t.Error("the city's page is held after its child link was removed")
			}
			if got := f.townChain(w); len(got) != 0 {
				t.Errorf("chain %v after the link was removed, want none", got)
			}
		}},
		{"self-loop", func(t *testing.T, _ *fixture, _ *chainWorld) {
			c := graphcachetest.OpenCatalog(t)
			k := graphcachetest.NewCanon(t, c)
			loop := k.Entity("place")
			k.PartOf(loop, loop)
			k.Commit(loop)
			g := c.Graph()
			g.Drop()
			if got, err := conclusionsummaries.ParentsAtDate(c, loop, conclusionsummaries.TodayDate()); err != nil || len(got) != 0 {
				t.Errorf("chain of a place that is part of itself: %v, %v", got, err)
			}
			h := conclusionsummaries.PlaceSummary{}
			h.Entity.ID = loop
			must(t, conclusionsummaries.AttachPlaceRelationships(c, &h))
			if len(h.PartOf) != 1 || len(h.Contains) != 1 {
				t.Errorf("part of %d, contains %d, want one each", len(h.PartOf), len(h.Contains))
			}
			db, err := c.DB()
			must(t, err)
			in := graphcache.Over(db).Inputs(nil)
			if _, err := in.Node(loop); err != nil {
				t.Fatal(err)
			}
			reads := func(hopList ...hops.Hop) []graphcache.InputKey {
				in := graphcache.Over(db).Inputs(nil)
				if _, err := in.Node(loop); err != nil {
					t.Fatal(err)
				}
				for _, hop := range hopList {
					links, err := in.Follow(loop, hop)
					must(t, err)
					for _, l := range links {
						if _, err := in.Node(l.Association); err != nil {
							t.Fatal(err)
						}
					}
				}
				return in.Keys()
			}
			_, _, chain := conclusionsummaries.PlaceEntryForTest(g, loop)
			_, page := conclusionsummaries.PlacePageForTest(g, loop)
			if want := reads(hops.ParentsOfPlace); !slices.EqualFunc(chain, want, graphcache.InputKey.Equal) {
				t.Errorf("chain inputs %+v, want %+v", chain, want)
			}
			if want := reads(hops.PartsOfPlace, hops.SuccessorsOfPlace, hops.PredecessorsOfPlace, hops.ParentsOfPlace); !slices.EqualFunc(page, want, graphcache.InputKey.Equal) {
				t.Errorf("page inputs %+v, want %+v", page, want)
			}
			if err := g.Verify(); err != nil {
				t.Errorf("Verify: %v", err)
			}
		}},
		{"vocabulary", func(t *testing.T, f *fixture, w *chainWorld) {
			graphcachetest.NewCanon(t, f.C).CommitVocabulary()
			if n := f.C.Graph().KeyCounts()["conclusionsummaries.places"]; n != 0 {
				t.Errorf("%d place entries after a vocabulary commit, want 0", n)
			}
		}},
		{"corrupt", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			conclusionsummaries.CorruptPlaceEntryForTest(g, w.city)
			err := g.Verify()
			if err == nil || !strings.Contains(err.Error(), "stored place entry") {
				t.Errorf("Verify error %v, want one naming the stored place entry", err)
			}
		}},
		{"bad succession date", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			start := 1800
			a := f.placeSubject("Alpha", nil, nil, 10, 0)
			b := f.placeSubject("Beta", nil, nil, 12, 0)
			f.placeRel(a, b, propertyterms.KeySucceededBy, &start, nil)
			aID := f.promoteSubject(a)
			f.promoteSubject(b)
			before, err := conclusionsummaries.PlacesByIDs(f.C, [][]byte{aID})
			must(t, err)
			f.breakDates()
			g.Drop()
			list, err := conclusionsummaries.ListPlaces(f.C)
			if err != nil {
				t.Fatalf("ListPlaces: %v", err)
			}
			after := 0
			for _, p := range list {
				if bytes.Equal(p.Entity.ID, aID) {
					after++
					if !slices.Equal(p.Names, before[0].Names) {
						t.Errorf("alpha reads %v, was %v", p.Names, before[0].Names)
					}
				}
			}
			if after != 1 {
				t.Fatalf("alpha listed %d times", after)
			}
			if _, _, err := conclusionsummaries.SuccessionNames(f.C, aID); err == nil {
				t.Error("SuccessionNames of alpha read a bad succession date without error")
			}
			h := conclusionsummaries.PlaceSummary{}
			h.Entity.ID = aID
			if err := conclusionsummaries.AttachPlaceRelationships(f.C, &h); err == nil {
				t.Error("AttachPlaceRelationships read a bad succession date without error")
			}
		}},
		{"bad part date", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			start := 1800
			hill := f.placeSubject("Hill", nil, nil, 14, 0)
			f.placeRel(hill, w.citySubject, propertyterms.KeyPartOf, &start, nil)
			hillID := f.promoteSubject(hill)
			f.breakDates()
			g.Drop()
			if _, err := conclusionsummaries.ParentsAtDate(f.C, w.city, conclusionsummaries.TodayDate()); err != nil {
				t.Errorf("the city's chain: %v", err)
			}
			if _, err := conclusionsummaries.ParentsAtDate(f.C, w.town, conclusionsummaries.TodayDate()); err != nil {
				t.Errorf("the town's chain does not climb the broken link: %v", err)
			}
			if _, err := conclusionsummaries.ParentsAtDate(f.C, hillID, conclusionsummaries.TodayDate()); err == nil {
				t.Error("the hill's own chain climbs the broken link and read it without error")
			}
		}},
		{"chain fills no page", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			g.Drop()
			_, err := conclusionsummaries.ListPlaces(f.C)
			must(t, err)
			if n := g.KeyCounts()["conclusionsummaries.placePages"]; n != 0 {
				t.Errorf("%d place pages held after a list, want 0", n)
			}
			if n := g.KeyCounts()["conclusionsummaries.places"]; n == 0 {
				t.Error("the list held no place entries")
			}
			h := conclusionsummaries.PlaceSummary{}
			h.Entity.ID = w.city
			must(t, conclusionsummaries.AttachPlaceRelationships(f.C, &h))
			if n := g.KeyCounts()["conclusionsummaries.placePages"]; n != 1 {
				t.Errorf("%d place pages held after the city's page, want 1", n)
			}
		}},
		{"missing place", func(t *testing.T, f *fixture, w *chainWorld) {
			g := f.C.Graph()
			unknown := bytes.Repeat([]byte{0xee}, 16)
			got, err := conclusionsummaries.ParentsAtDate(f.C, unknown, conclusionsummaries.TodayDate())
			must(t, err)
			if len(got) != 0 {
				t.Errorf("chain %v of an unknown place", got)
			}
			held, found, inputs := conclusionsummaries.PlaceEntryForTest(g, unknown)
			if !held || found || !slices.EqualFunc(inputs, []graphcache.InputKey{{ID: unknown}}, graphcache.InputKey.Equal) {
				t.Errorf("entry held %v found %v inputs %+v", held, found, inputs)
			}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			w := f.chainWorld()
			f.C.Graph().Drop()
			f.townChain(w)
			tc.run(t, f, w)
		})
	}
	t.Run("cap", func(t *testing.T) {
		f := newFixture(t)
		var subs []subjects.Subject
		for i := range 5 {
			subs = append(subs, f.placeSubject(fmt.Sprintf("P%d", i), nil, nil, int64(2*i), 0))
		}
		for i := 0; i+1 < len(subs); i++ {
			f.placeRel(subs[i], subs[i+1], propertyterms.KeyPartOf, nil, nil)
		}
		var ids [][]byte
		for _, s := range subs {
			ids = append(ids, f.promoteSubject(s))
		}
		g := f.C.Graph()
		g.Drop()
		conclusionsummaries.SetPlaceEntryCapForTest(g, 2)
		got, err := conclusionsummaries.ParentsAtDate(f.C, ids[0], conclusionsummaries.TodayDate())
		must(t, err)
		if len(got) != 4 {
			t.Errorf("chain %v, want four names", got)
		}
		if n := g.KeyCounts()["conclusionsummaries.places"]; n > 2 {
			t.Errorf("%d place entries held, want at most 2", n)
		}
	})
}

// TestPlaceEntriesFollowLinkMoves holds the three cases whose commits move a
// place's links without writing the place. The city's page is held too, and a
// link at it drops the page.
func TestPlaceEntriesFollowLinkMoves(t *testing.T) {
	tests := []struct {
		name string
		move func(k *graphcachetest.Canon, town, city, region []byte)
		held [3]bool // town, city, region
		page bool    // the city's page is held
	}{
		{"a location at the city", func(k *graphcachetest.Canon, town, city, region []byte) {
			k.Commit(k.Location(k.Event(vocabkey.EventBirth), city))
		}, [3]bool{true, true, true}, true},
		{"a new part_of link at the city", func(k *graphcachetest.Canon, town, city, region []byte) {
			k.Commit(k.PartOf(k.Entity("place"), city))
		}, [3]bool{true, false, true}, false},
		{"a merge of the region", func(k *graphcachetest.Canon, town, city, region []byte) {
			into := k.Entity("place")
			k.Merge(region, into)
			k.Commit(region, into)
		}, [3]bool{true, false, false}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := graphcachetest.OpenCatalog(t)
			k := graphcachetest.NewCanon(t, c)
			town, city, region := k.Entity("place"), k.Entity("place"), k.Entity("place")
			k.PartOf(town, city)
			k.PartOf(city, region)
			k.Commit(town, city, region)
			g := c.Graph()
			must(t, conclusionsummaries.FillPlaceEntriesForTest(g, town, city, region))
			must(t, conclusionsummaries.FillPlacePagesForTest(g, city))
			tc.move(k, town, city, region)
			for i, id := range [][]byte{town, city, region} {
				if got := conclusionsummaries.HeldPlaceEntryForTest(g, id); got != tc.held[i] {
					t.Errorf("%s held %v, want %v", []string{"town", "city", "region"}[i], got, tc.held[i])
				}
			}
			if held, _ := conclusionsummaries.PlacePageForTest(g, city); held != tc.page {
				t.Errorf("the city's page held %v, want %v", held, tc.page)
			}
		})
	}
}

func TestPlaceEntryInputsMatchReads(t *testing.T) {
	f := newFixture(t)
	w := f.chainWorld()
	unknown := bytes.Repeat([]byte{0xee}, 16)
	g := f.C.Graph()
	g.Drop()
	must(t, conclusionsummaries.FillPlaceEntriesForTest(g, w.town, w.city, w.region, unknown))
	must(t, conclusionsummaries.FillPlacePagesForTest(g, w.town, w.city, w.region, unknown))
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
			held, _, inputs := conclusionsummaries.PlaceEntryForTest(g, id)
			pageHeld, pageInputs := conclusionsummaries.PlacePageForTest(g, id)
			if !held || !pageHeld {
				t.Fatalf("entry held %v, page held %v", held, pageHeld)
			}
			same := func(what string, got, want []graphcache.InputKey) {
				t.Helper()
				if !slices.EqualFunc(got, want, graphcache.InputKey.Equal) {
					t.Errorf("%s: got %+v, want %+v", what, got, want)
				}
			}
			same("inputs", inputs, reads(id, hops.ParentsOfPlace))
			same("page inputs", pageInputs, reads(id, hops.PartsOfPlace, hops.SuccessorsOfPlace, hops.PredecessorsOfPlace))
		})
	}
}

// meshWorld is the crafted part_of mesh BenchmarkPlaceMesh reads: a hub with
// 20,000 parts that is the lowest place of a chain of eight, and one event
// located at the hub per 10 parts.
type meshWorld struct {
	c         *database.Catalog
	k         *graphcachetest.Canon
	hub       []byte
	part      []byte // the first part of the hub
	partEvent []byte // an event located at it
	wide      int
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
			part := w.k.Entity("place")
			if i == 0 {
				w.part, w.partEvent = part, w.k.Event(vocabkey.EventBirth)
				w.k.Location(w.partEvent, part)
			}
			w.k.PartOf(part, w.hub)
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
	b.Run("write-under-hub", func(b *testing.B) {
		// A write's one-shot graph composes the rows of an event located at a part
		// of the hub and of that part: its chain climbs into the hub, which has
		// 20,000 parts.
		db, err := w.c.DB()
		if err != nil {
			b.Fatal(err)
		}
		want := conclusionsummaries.Want{Events: [][]byte{w.partEvent}, Places: [][]byte{w.part}}
		b.ReportMetric(ms(databasetest.Since(func() {
			if _, err := conclusionsummaries.RowsForSearch(graphcache.Over(db), want); err != nil {
				b.Fatal(err)
			}
		})), "ms")
	})
	b.Log(fmt.Sprintf("hub with %d parts", w.wide))
}
