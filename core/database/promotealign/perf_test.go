package promotealign_test

import (
	"math/rand/v2"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/graphalign"
)

// The budgets of S11-21 D-1. They assume go test runs packages in parallel on
// four cores, as scripts/check-pr.sh and CI do. The warm figure was raised from
// 250 ms by the user on 2026-10-26 when the fixture gained the 12-generation
// tree and matching read kinship links (S11-39), then from 350 ms to 375 ms by
// the user on 2026-11-05, when the CI runner measured 350.57 ms (S14-32) and
// S14-20 had measured 360 ms as a best of five. The after-commit figure was
// raised from 1.5 s by the user on 2026-11-01: the CI runner measured 1.6–2.05 s
// where a Mac measures 0.6 s (S13-02, S13-03).
const (
	budgetWarmOnOpen   = 2500 * time.Millisecond
	budgetCold         = 2500 * time.Millisecond
	budgetWarm         = 375 * time.Millisecond
	budgetAfterCommit  = 2500 * time.Millisecond
	budgetOnePlaceCold = 5 * time.Second
	budgetOnePlaceWarm = 3 * time.Second
	// budgetJoinOntoCity: joining a second "City" toponym onto the city moves no
	// shown value, so only the city's document is reprojected, not its 6,000
	// events and their people. B-10 measured 6.06 s before S14-21.
	budgetJoinOntoCity = 1 * time.Second
	// budgetLocateAtCity: filing a location between a new event and the city
	// moves no row of the city's other 6,000 events, so none is reprojected.
	// S14-33 measured 3.2 s before it walked from the event alone.
	budgetLocateAtCity = 500 * time.Millisecond
	// budgetSecondPartOfAtCity: a second record's part_of from the city joins
	// the association already filed, which moves no row, so the city's 6,000
	// events are not reprojected. S14-33 measured 2.75 s before movedRoots
	// followed only the associations that moved.
	budgetSecondPartOfAtCity = 500 * time.Millisecond
)

// largeFixture is a realistic large catalog (largeShape) and the Sources the
// budgets run on.
type largeFixture struct {
	*fixture
	// a is Source A: 100 primary Subjects, none located.
	a sources.Source
	// onePlace is the Source whose 30 births are all located at one place, the
	// city, which has 6,000 filed events.
	onePlace sources.Source
	// pageArt is the artifact the one-place Source's births are cited on.
	pageArt artifacts.Artifact
	other   sources.Source
	otherA  artifacts.Artifact
	// city is the city's handle id and citySubject the Subject it was made from.
	city        []byte
	citySubject subjects.Subject
}

func newLargeFixture(t *testing.T) *largeFixture {
	t.Helper()
	f := newFixture(t)
	must(t, nameformats.Install(f.c))
	must(t, namealiases.Install(f.c))
	must(t, catalogsession.WarmOnOpen(f.c))

	placeSrc, placeArt := f.NewSource("Births at one place")
	city := f.Place(placeSrc, placeArt, "City")
	cityID := f.promote(city).Entity.ID

	r := rand.New(rand.NewPCG(1, 2))
	names := newZipfNames(r)

	filed := make([]string, 20)
	for n := 1; n <= 20; n++ {
		filed[n-1] = names.next()
		src, art := f.NewSource("Census " + letters(n))
		f.filedBirth(src, art, filed[n-1], 1850+n, nil, nil)
	}

	db, err := f.c.DB()
	must(t, err)
	must(t, bulkCanon(db, cityID, largeShape(r, names, filed[0])))
	f.c.Graph().Drop()

	// Source A: 40 persons, 40 births and 20 deaths, none located.
	filedOn := func(name string, birthYear, deathYear int) {
		p := f.Person(f.source, f.artifact, name)
		b := f.EventOn(f.source, f.artifact, "birth", fixtureDate(r, birthYear))
		f.Participation(f.source, f.artifact, p, b, "subject")
		if deathYear != 0 {
			d := f.EventOn(f.source, f.artifact, "death", fixtureDate(r, deathYear))
			f.Participation(f.source, f.artifact, p, d, "subject")
		}
	}
	for n := 1; n <= 20; n++ {
		filedOn(filed[n-1], 1850+n, 1900+n)
	}
	for i := 0; i < 20; i++ {
		filedOn(names.next(), 1800+r.IntN(100), 0)
	}

	// The one-place Source: 30 births at the city, on an artifact of its own.
	pageArt, err := artifactstest.Create(f.c, userID, artifacts.CreateInput{SourceID: placeSrc.ID, Label: "Page"})
	must(t, err)
	for i := 0; i < 30; i++ {
		p := f.Person(placeSrc, pageArt, names.next())
		b := f.EventOn(placeSrc, pageArt, "birth", fixtureDate(r, 1800+r.IntN(100)))
		f.Participation(placeSrc, pageArt, p, b, "subject")
		f.Locate(placeSrc, pageArt, b, city)
	}

	otherSrc, otherArt := f.NewSource("Other")
	return &largeFixture{fixture: f, a: f.source, onePlace: placeSrc, pageArt: pageArt, other: otherSrc, otherA: otherArt, city: cityID, citySubject: city}
}

// TestLargeSourceBudgets holds Promote to the times of S11-21 D-1 on a catalog
// of 4,000 persons and 8,000 events (one city with 6,000), so a change that
// makes it slow fails here instead of in a dogfood session. Subtests share one
// fixture and run in order.
func TestLargeSourceBudgets(t *testing.T) {
	lf := newLargeFixture(t)
	f := lf.fixture
	aGraph := f.Graph(lf.a)
	propose := func() graphalign.Proposal { return f.propose(aGraph, nil) }
	t.Run("fixture", func(t *testing.T) {
		p := propose()
		var handles, strong int
		for _, r := range p.Rows {
			if r.Target == graphalign.TargetHandle {
				handles++
			}
			if r.Assessment == graphalign.AssessStrong {
				strong++
			}
		}
		t.Logf("rows %d, handle rows %d, strong rows %d", len(p.Rows), handles, strong)
		if len(p.Rows) != 100 || handles < 40 {
			t.Fatalf("Source A has %d rows, %d with a handle; want 100 and at least 40", len(p.Rows), handles)
		}
	})

	// What the open spinner covers: the alias index, name format, shortlist
	// kinds and name index.
	t.Run("warm at open", func(t *testing.T) {
		d := databasetest.Fastest(2, func() time.Duration {
			f.c.Graph().Drop()
			return databasetest.Since(func() { must(t, catalogsession.WarmOnOpen(f.c)) })
		})
		databasetest.WithinBudget(t, "warm at open", d, budgetWarmOnOpen)
	})

	// The first proposal after an open.
	t.Run("cold", func(t *testing.T) {
		d := databasetest.Fastest(2, func() time.Duration {
			f.c.Graph().Drop()
			must(t, catalogsession.WarmOnOpen(f.c))
			return databasetest.Since(func() { propose() })
		})
		databasetest.WithinBudget(t, "cold proposal", d, budgetCold)
	})

	// A warm proposal rebuilds the canon each time, about 120 MB of short-lived
	// allocations over a heap of about 170 MB, so each run sets off one or two
	// collections. Collecting before each timed run bills a run for its own
	// garbage, not for what the subtests before it left; five runs take the
	// noise of a loaded CI runner (2026-11-03, a best of three measured 388 ms
	// there against 110 – 190 ms on a Mac). The cost itself is
	// docs/ideas/promote-ranking.md#narrower-canon-walk.
	t.Run("warm", func(t *testing.T) {
		propose()
		d := databasetest.Fastest(5, func() time.Duration {
			runtime.GC()
			return databasetest.Since(func() { propose() })
		})
		databasetest.WithinBudget(t, "warm proposal", d, budgetWarm)
	})

	t.Run("memo matches a fresh run", func(t *testing.T) {
		propose()
		warm := propose()
		f.c.Graph().Drop()
		if fresh := propose(); !reflect.DeepEqual(warm, fresh) {
			t.Fatal("the warm proposal differs from a proposal on a dropped graph")
		}
	})

	t.Run("after a handle commit", func(t *testing.T) {
		n := 0
		d := databasetest.Fastest(2, func() time.Duration {
			n++
			f.promote(f.Person(lf.other, lf.otherA, "Added "+letters(n)+" Person"))
			return databasetest.Since(func() { propose() })
		})
		databasetest.WithinBudget(t, "proposal after a handle commit", d, budgetAfterCommit)
	})

	t.Run("warm queries", func(t *testing.T) {
		propose()
		n := 0
		f.c.Graph().SetQueryCounterForTest(func() { n++ })
		defer f.c.Graph().SetQueryCounterForTest(nil)
		selects := databasetest.CountSelects(t, f.c)
		propose()
		if n != 0 {
			t.Errorf("a warm Propose issued %d graph queries, want 0", n)
		}
		if selects.N() != 1 {
			t.Errorf("a warm Propose issued %d SELECTs, want 1 (MAX(revision))", selects.N())
		}
	})

	onePlaceGraph := f.Graph(lf.onePlace)
	proposeOnePlace := func() { f.propose(onePlaceGraph, nil) }

	t.Run("births at one place cold", func(t *testing.T) {
		d := databasetest.Fastest(2, func() time.Duration {
			f.c.Graph().Drop()
			must(t, catalogsession.WarmOnOpen(f.c))
			return databasetest.Since(proposeOnePlace)
		})
		databasetest.WithinBudget(t, "cold proposal on the one-place Source", d, budgetOnePlaceCold)
	})

	t.Run("births at one place warm", func(t *testing.T) {
		proposeOnePlace()
		d := databasetest.Fastest(3, func() time.Duration { return databasetest.Since(proposeOnePlace) })
		databasetest.WithinBudget(t, "warm proposal on the one-place Source", d, budgetOnePlaceWarm)
	})

	t.Run("join a place onto the city", func(t *testing.T) {
		var joining []subjects.Subject
		for i := 0; i < 3; i++ {
			joining = append(joining, f.Place(lf.other, lf.otherA, "City"))
		}
		i := 0
		d := databasetest.Fastest(3, func() time.Duration {
			s := joining[i]
			i++
			return databasetest.Since(func() { f.join(s, lf.city) })
		})
		databasetest.WithinBudget(t, "join a place onto the city", d, budgetJoinOntoCity)
	})

	t.Run("locate an event at the city", func(t *testing.T) {
		var events []subjects.Subject
		for i := 0; i < 3; i++ {
			e := f.Birth(lf.onePlace, lf.pageArt)
			f.promote(e)
			events = append(events, e)
		}
		i := 0
		d := databasetest.Fastest(3, func() time.Duration {
			e := events[i]
			i++
			return databasetest.Since(func() { f.Locate(lf.onePlace, lf.pageArt, e, lf.citySubject) })
		})
		databasetest.WithinBudget(t, "file one location at the city", d, budgetLocateAtCity)
	})
	t.Run("second part_of from the city", func(t *testing.T) {
		region := f.Place(lf.other, lf.otherA, "Region")
		regionID := f.promote(region).Entity.ID
		// The first part_of mints the association; it moves rows and is not timed.
		first := f.Place(lf.other, lf.otherA, "City")
		f.join(first, lf.city)
		f.PlaceRel(lf.other, lf.otherA, first, region, "part_of")
		type pair struct{ from, to subjects.Subject }
		var pairs []pair
		for i := 0; i < 3; i++ {
			from := f.Place(lf.other, lf.otherA, "City")
			f.join(from, lf.city)
			to := f.Place(lf.other, lf.otherA, "Region")
			f.join(to, regionID)
			pairs = append(pairs, pair{from, to})
		}
		i := 0
		d := databasetest.Fastest(3, func() time.Duration {
			p := pairs[i]
			i++
			return databasetest.Since(func() { f.PlaceRel(lf.other, lf.otherA, p.from, p.to, "part_of") })
		})
		databasetest.WithinBudget(t, "file a second part_of from the city", d, budgetSecondPartOfAtCity)
	})
}
