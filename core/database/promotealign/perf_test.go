package promotealign_test

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession"
	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/graphalign"
)

// The budgets of S11-21 D-1. They assume go test runs packages in parallel on
// four cores, as scripts/check-pr.sh and CI do. The warm figure was raised from
// 250 ms by the user on 2026-10-26 when the fixture gained the 12-generation
// tree and matching read kinship links (S11-39).
const (
	budgetWarmOnOpen   = 2500 * time.Millisecond
	budgetCold         = 2500 * time.Millisecond
	budgetWarm         = 350 * time.Millisecond
	budgetAfterCommit  = 1500 * time.Millisecond
	budgetEvidence     = 100 * time.Millisecond
	budgetOnePlaceCold = 5 * time.Second
	budgetOnePlaceWarm = 3 * time.Second
	// budgetPlaceEvidence is a place row's evidence on a cold graph: without
	// kinship links to derive it reads two nodes, not the place's 6,000 events.
	budgetPlaceEvidence = 10 * time.Millisecond
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
	other    sources.Source
	otherA   artifacts.Artifact
}

func newLargeFixture(t *testing.T) *largeFixture {
	t.Helper()
	f := newFixture(t)
	must(t, nameformats.Install(f.c))
	must(t, namealiases.Install(f.c))
	must(t, catalogsession.WarmOnOpen(f.c))

	placeSrc, placeArt := f.newSource("Births at one place")
	city := f.place(placeSrc, placeArt, "City")
	cityID := f.promote(city).Entity.ID

	r := rand.New(rand.NewPCG(1, 2))
	names := newZipfNames(r)

	filed := make([]string, 20)
	for n := 1; n <= 20; n++ {
		filed[n-1] = names.next()
		src, art := f.newSource("Census " + letters(n))
		f.filedBirth(src, art, filed[n-1], 1850+n, nil, nil)
	}

	db, err := f.c.DB()
	must(t, err)
	must(t, bulkCanon(db, cityID, largeShape(r, names, filed[0])))
	f.c.Graph().Drop()

	// Source A: 40 persons, 40 births and 20 deaths, none located.
	filedOn := func(name string, birthYear, deathYear int) {
		p := f.person(f.source, f.artifact, name)
		b := f.eventOn(f.source, f.artifact, "birth", fixtureDate(r, birthYear))
		f.participation(f.source, f.artifact, p, b, "subject")
		if deathYear != 0 {
			d := f.eventOn(f.source, f.artifact, "death", fixtureDate(r, deathYear))
			f.participation(f.source, f.artifact, p, d, "subject")
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
		p := f.person(placeSrc, pageArt, names.next())
		b := f.eventOn(placeSrc, pageArt, "birth", fixtureDate(r, 1800+r.IntN(100)))
		f.participation(placeSrc, pageArt, p, b, "subject")
		f.locate(placeSrc, pageArt, b, city)
	}

	otherSrc, otherArt := f.newSource("Other")
	return &largeFixture{fixture: f, a: f.source, onePlace: placeSrc, other: otherSrc, otherA: otherArt}
}

// fastest is the least of n timed runs of run; each run's own setup stays
// outside the time it returns.
func fastest(n int, run func() time.Duration) time.Duration {
	best := time.Duration(-1)
	for i := 0; i < n; i++ {
		if d := run(); best < 0 || d < best {
			best = d
		}
	}
	return best
}

// since times one call.
func since(call func()) time.Duration {
	start := time.Now()
	call()
	return time.Since(start)
}

// TestLargeSourceBudgets holds Promote to the times of S11-21 D-1 on a catalog
// of 4,000 persons and 8,000 events (one city with 6,000), so a change that
// makes it slow fails here instead of in a dogfood session. Subtests share one
// fixture and run in order.
func TestLargeSourceBudgets(t *testing.T) {
	lf := newLargeFixture(t)
	f := lf.fixture
	aGraph := f.graph(lf.a)
	propose := func() graphalign.Proposal { return f.propose(aGraph, nil) }
	over := func(t *testing.T, what string, got, budget time.Duration) {
		t.Helper()
		t.Logf("%s: %v (budget %v)", what, got, budget)
		if got > budget {
			t.Errorf("%s took %v, over the %v budget", what, got, budget)
		}
	}

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
		d := fastest(2, func() time.Duration {
			f.c.Graph().Drop()
			return since(func() { must(t, catalogsession.WarmOnOpen(f.c)) })
		})
		over(t, "warm at open", d, budgetWarmOnOpen)
	})

	// The first proposal after an open.
	t.Run("cold", func(t *testing.T) {
		d := fastest(2, func() time.Duration {
			f.c.Graph().Drop()
			must(t, catalogsession.WarmOnOpen(f.c))
			return since(func() { propose() })
		})
		over(t, "cold proposal", d, budgetCold)
	})

	t.Run("warm", func(t *testing.T) {
		propose()
		d := fastest(3, func() time.Duration { return since(func() { propose() }) })
		over(t, "warm proposal", d, budgetWarm)
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
		d := fastest(2, func() time.Duration {
			n++
			f.promote(f.person(lf.other, lf.otherA, "Added "+letters(n)+" Person"))
			return since(func() { propose() })
		})
		over(t, "proposal after a handle commit", d, budgetAfterCommit)
	})

	// evidenceRequests is one request for each handle row, naming every handle
	// row as a neighbor.
	evidenceRequests := func() []promotealign.EvidenceRequest {
		var neighbors []graphalign.Fixed
		for _, r := range propose().Rows {
			if r.Target == graphalign.TargetHandle {
				neighbors = append(neighbors, graphalign.Fixed{SubjectID: r.SubjectID, HandleID: r.HandleID})
			}
		}
		var reqs []promotealign.EvidenceRequest
		for _, n := range neighbors {
			reqs = append(reqs, promotealign.EvidenceRequest{
				GraphID: aGraph, SubjectID: n.SubjectID, HandleID: n.HandleID, Neighbors: neighbors,
			})
		}
		return reqs
	}
	pass := func(reqs []promotealign.EvidenceRequest) {
		for _, req := range reqs {
			f.evidence(req)
		}
	}

	t.Run("evidence warm", func(t *testing.T) {
		reqs := evidenceRequests()
		pass(reqs)
		d := fastest(3, func() time.Duration { return since(func() { pass(reqs) }) })
		over(t, fmt.Sprintf("warm Evidence on %d handle rows", len(reqs)), d, budgetEvidence)
	})

	t.Run("warm queries", func(t *testing.T) {
		reqs := evidenceRequests()
		pass(reqs)
		propose()
		n := 0
		f.c.Graph().SetQueryCounterForTest(func() { n++ })
		defer f.c.Graph().SetQueryCounterForTest(nil)
		selects := databasetest.CountSelects(t, f.c)
		pass(reqs)
		if n != 0 || selects.N() != 0 {
			t.Errorf("warm Evidence issued %d graph queries and %d SELECTs, want 0 and 0", n, selects.N())
		}
		selects.Reset()
		propose()
		if n != 0 {
			t.Errorf("a warm Propose issued %d graph queries, want 0", n)
		}
		if selects.N() != 1 {
			t.Errorf("a warm Propose issued %d SELECTs, want 1 (MAX(revision))", selects.N())
		}
	})

	onePlaceGraph := f.graph(lf.onePlace)
	proposeOnePlace := func() { f.propose(onePlaceGraph, nil) }

	t.Run("births at one place cold", func(t *testing.T) {
		d := fastest(2, func() time.Duration {
			f.c.Graph().Drop()
			must(t, catalogsession.WarmOnOpen(f.c))
			return since(proposeOnePlace)
		})
		over(t, "cold proposal on the one-place Source", d, budgetOnePlaceCold)
	})

	t.Run("births at one place warm", func(t *testing.T) {
		proposeOnePlace()
		d := fastest(3, func() time.Duration { return since(proposeOnePlace) })
		over(t, "warm proposal on the one-place Source", d, budgetOnePlaceWarm)
	})

	// A place has no kinship links, so evidence on its row must not walk its
	// 6,000 events for them.
	t.Run("evidence on a place row", func(t *testing.T) {
		var city *graphalign.Row
		p := f.propose(onePlaceGraph, nil)
		for i, r := range p.Rows {
			if r.Target == graphalign.TargetHandle && r.Kind == "place" {
				city = &p.Rows[i]
			}
		}
		if city == nil {
			t.Fatal("the one-place Source has no place row with a handle")
		}
		req := promotealign.EvidenceRequest{GraphID: onePlaceGraph, SubjectID: city.SubjectID, HandleID: city.HandleID}
		d := fastest(3, func() time.Duration {
			f.c.Graph().Drop()
			must(t, catalogsession.WarmOnOpen(f.c))
			return since(func() { f.evidence(req) })
		})
		over(t, "evidence on the place row, cold", d, budgetPlaceEvidence)
	})

	// A pass over Source A touches 21 Evidence graphs; a cap of 8 evicts along
	// the way and every response must still match the default cap's.
	t.Run("evidence under a small cap", func(t *testing.T) {
		const small = 8
		reqs := evidenceRequests()
		f.c.Graph().Drop()
		want := make([]claimevidence.Evidence, len(reqs))
		for i, req := range reqs {
			want[i] = f.evidence(req)
		}
		f.c.Graph().SetEvidenceGraphCapForTest(small)
		defer f.c.Graph().SetEvidenceGraphCapForTest(0)
		f.c.Graph().Drop()
		for i, req := range reqs {
			got := f.evidence(req)
			if n := f.c.Graph().LoadedEvidenceGraphCountForTest(); n > small {
				t.Fatalf("request %d left %d Evidence graphs loaded, want at most %d", i, n, small)
			}
			if !reflect.DeepEqual(got, want[i]) {
				t.Fatalf("request %d under cap %d differs from the default cap's response", i, small)
			}
		}
	})
}
