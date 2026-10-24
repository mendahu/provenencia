package promotealign_test

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession"
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
// four cores, as scripts/check-pr.sh and CI do.
const (
	budgetCold        = 2500 * time.Millisecond
	budgetWarm        = 250 * time.Millisecond
	budgetAfterCommit = 1500 * time.Millisecond
	budgetEvidence    = 100 * time.Millisecond
	budgetHubCold     = 5 * time.Second
	budgetHubWarm     = 3 * time.Second
)

// largeFixture is a realistic large catalog (largeShape) and the Sources the
// budgets run on.
type largeFixture struct {
	*fixture
	// a is Source A: 100 primary Subjects, none located.
	a sources.Source
	// hubPage is the Source whose 30 births are all located at the hub place.
	hubPage sources.Source
	other   sources.Source
	otherA  artifacts.Artifact
}

func newLargeFixture(t *testing.T) *largeFixture {
	t.Helper()
	f := newFixture(t)
	must(t, nameformats.Install(f.c))
	must(t, namealiases.Install(f.c))
	must(t, catalogsession.WarmOnOpen(f.c))

	hubSrc, hubArt := f.newSource("Hub page")
	hubPlace := f.place(hubSrc, hubArt, "Hub")
	hub := f.promote(hubPlace).Entity.ID

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
	must(t, bulkCanon(db, hub, largeShape(r, names)))
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

	// The hub page: 30 births at the city, on an artifact of its own.
	pageArt, err := artifactstest.Create(f.c, userID, artifacts.CreateInput{SourceID: hubSrc.ID, Label: "Page"})
	must(t, err)
	for i := 0; i < 30; i++ {
		p := f.person(hubSrc, pageArt, names.next())
		b := f.eventOn(hubSrc, pageArt, "birth", fixtureDate(r, 1800+r.IntN(100)))
		f.participation(hubSrc, pageArt, p, b, "subject")
		f.locate(hubSrc, pageArt, b, hubPlace)
	}

	otherSrc, otherArt := f.newSource("Other")
	return &largeFixture{fixture: f, a: f.source, hubPage: hubSrc, other: otherSrc, otherA: otherArt}
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

	t.Run("cold", func(t *testing.T) {
		d := fastest(2, func() time.Duration {
			f.c.Graph().Drop()
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

	hubGraph := f.graph(lf.hubPage)
	proposeHub := func() { f.propose(hubGraph, nil) }

	t.Run("hub page cold", func(t *testing.T) {
		d := fastest(2, func() time.Duration {
			f.c.Graph().Drop()
			return since(proposeHub)
		})
		over(t, "cold proposal on the hub page", d, budgetHubCold)
	})

	t.Run("hub page warm", func(t *testing.T) {
		proposeHub()
		d := fastest(3, func() time.Duration { return since(proposeHub) })
		over(t, "warm proposal on the hub page", d, budgetHubWarm)
	})
}
