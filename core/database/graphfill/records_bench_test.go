package graphfill_test

import (
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/graphfill"
)

// Budgets of S16-13 D-15 at the map's Scale: every graph's records filled once.
const (
	maxRecordsBuild  = 2 * time.Second
	maxRecordsHeapMB = 35
)

// BenchmarkGraphRecords: the records of all 10,000 graphs of the Scale (3,000
// Sources, 20,000 Subjects, 50,000 Observations) filled from the rows, one
// graph at a time as the claims read asks for them. The entry holds at most
// 4,096 of them, so the heap is that many.
func BenchmarkGraphRecords(b *testing.B) {
	c := graphcachetest.OpenCatalog(b)
	seed := evidencegraphstest.NewScaleSeed(b, c)
	seed.Base()
	seed.Commit()
	g := c.Graph()
	g.Drop()

	var build time.Duration
	heap := databasetest.HeapDelta(func() {
		start := time.Now()
		for i := 1; i <= evidencegraphstest.ScaleGraphs; i++ {
			if _, err := graphfill.RecordsOf(g, [][]byte{databasetest.BenchID(3, i)}); err != nil {
				b.Fatal(err)
			}
		}
		build = time.Since(start)
	})
	b.ReportMetric(float64(build.Microseconds())/1000, "build-ms")
	b.ReportMetric(heap/(1<<20), "heap-MB")
	b.ReportMetric(float64(g.KeyCounts()["graphfill.records"]), "records-keys")
	if build > maxRecordsBuild {
		b.Errorf("records of %d graphs took %v, over %v", evidencegraphstest.ScaleGraphs, build, maxRecordsBuild)
	}
	if heap/(1<<20) > maxRecordsHeapMB {
		b.Errorf("records of %d graphs hold %.1f MB, over %d MB", evidencegraphstest.ScaleGraphs, heap/(1<<20), maxRecordsHeapMB)
	}
}
