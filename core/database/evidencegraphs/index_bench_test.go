package evidencegraphs_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
)

// BenchmarkGraphIndex: Build of the index at 3,000 Sources, 10,000 graphs each
// linked to an Artifact and holding a Citation, 20,000 person Subjects and
// 50,000 Observations; then Apply of one commit that recounts a graph of
// 10,000 Observations and one that re-lists a Source of 10,000 graphs.
func BenchmarkGraphIndex(b *testing.B) {
	const crafted = 10000
	c := graphcachetest.OpenCatalog(b)
	seed := evidencegraphstest.NewScaleSeed(b, c)
	seed.Base()
	seed.Commit()

	g := c.Graph()
	g.Drop()
	var build time.Duration
	heap := databasetest.HeapDelta(func() {
		start := time.Now()
		if _, err := evidencegraphs.IndexOf(g); err != nil {
			b.Fatal(err)
		}
		build = time.Since(start)
	})
	b.ReportMetric(float64(build.Microseconds())/1000, "build-ms")
	b.ReportMetric(heap/(1<<20), "heap-MB")

	// The crafted graph and Source join after the Build measure, which is the
	// ordinary shape.
	craft := evidencegraphstest.NewScaleSeed(b, c)
	heavy := evidencegraphstest.ScaleGraphs + 1
	craft.Graph(heavy, evidencegraphstest.ScaleSources+1)
	craft.Subject(evidencegraphstest.ScaleSubjects+1, heavy, evidencegraphstest.ScaleSources+1)
	for i := 1; i <= crafted; i++ {
		craft.Obs(evidencegraphstest.ScaleObservations+i, evidencegraphstest.ScaleSubjects+1, heavy)
	}
	wide := evidencegraphstest.ScaleSources + 2
	for i := 1; i <= crafted; i++ {
		craft.Exec(`INSERT INTO evidence_graphs (id, ref, source_id, name) VALUES (?, ?, ?, ?)`, databasetest.BenchID(3, heavy+i), "GPH-B"+strconv.Itoa(heavy+i), databasetest.BenchID(1, wide), "Wide "+strconv.Itoa(i))
	}
	craft.Commit()
	g.Drop()
	if _, err := evidencegraphs.IndexOf(g); err != nil {
		b.Fatal(err)
	}

	apply := func(metric string, read effects.Read, id []byte) {
		start := time.Now()
		for i := 0; i < b.N; i++ {
			if err := g.OnCommit(effects.Commit{Effects: effects.Set{Reads: []effects.ReadIDs{{Read: read, IDs: [][]byte{id}}}}}); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(time.Since(start).Microseconds())/1000/float64(b.N), metric)
	}
	apply("apply-progress-ms", effects.ReadEvidenceGraphProgress, databasetest.BenchID(3, 7))
	apply("apply-progress-10k-ms", effects.ReadEvidenceGraphProgress, databasetest.BenchID(3, heavy))
	apply("apply-list-ms", effects.ReadEvidenceGraphs, databasetest.BenchID(1, 7))
	apply("apply-list-10k-ms", effects.ReadEvidenceGraphs, databasetest.BenchID(1, wide))
}
