package sources_test

import (
	"encoding/binary"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func benchID(kind byte, n int) []byte {
	id := make([]byte, 16)
	id[0] = kind
	binary.BigEndian.PutUint64(id[8:], uint64(n))
	return id
}

// BenchmarkSourceRecords: Build of the records at 3,000 Sources of 10
// Artifacts (every tenth assessed), and Apply of one commit that names an
// ordinary Source and one that names a Source of 10,000 Artifacts.
func BenchmarkSourceRecords(b *testing.B) {
	const sourcesN, perSource, crafted = 3000, 10, 10000
	c, err := database.Create(b.TempDir(), "t.provenencia")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = c.Close() })
	if err := sourcecredibilitygrades.Install(c); err != nil {
		b.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		b.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		b.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}
	var gradeID []byte
	if err := tx.QueryRow(`SELECT id FROM source_credibility_grades WHERE key = 'low_trust'`).Scan(&gradeID); err != nil {
		b.Fatal(err)
	}
	n := 0
	seed := func(src, artifacts int) {
		sid := benchID(1, src)
		if _, err := tx.Exec(`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, ?, ?, ?)`, sid, "SRC-B"+string(rune('A'+src%26))+strconv.Itoa(src), typeID, "Source "+strconv.Itoa(src)); err != nil {
			b.Fatal(err)
		}
		if src%10 == 0 {
			if _, err := tx.Exec(`INSERT INTO source_credibility_assessments (id, source_id, credibility_grade_id) VALUES (?, ?, ?)`, benchID(3, src), sid, gradeID); err != nil {
				b.Fatal(err)
			}
		}
		for i := 0; i < artifacts; i++ {
			n++
			if _, err := tx.Exec(`INSERT INTO artifacts (id, ref, source_id, label) VALUES (?, ?, ?, ?)`, benchID(2, n), "ART-B"+strconv.Itoa(n), sid, "Page "+strconv.Itoa(i)); err != nil {
				b.Fatal(err)
			}
		}
	}
	for s := 1; s <= sourcesN; s++ {
		seed(s, perSource)
	}
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}

	g := c.Graph()
	g.Drop()
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	if _, err := sources.RecordsOf(g); err != nil {
		b.Fatal(err)
	}
	build := time.Since(start)
	runtime.GC()
	runtime.ReadMemStats(&after)
	b.ReportMetric(float64(build.Microseconds())/1000, "build-ms")
	b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc)/(1<<20), "heap-MB")

	// The crafted Source joins after the Build measure, which is the 3,000 x 10 shape.
	if tx, err = db.Begin(); err != nil {
		b.Fatal(err)
	}
	seed(sourcesN+1, crafted)
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}
	g.Drop()
	if _, err := sources.RecordsOf(g); err != nil {
		b.Fatal(err)
	}

	apply := func(metric string, id []byte) {
		start := time.Now()
		for i := 0; i < b.N; i++ {
			if err := g.OnCommit(effects.Commit{Effects: effects.Set{SourceScopes: [][]byte{id}}}); err != nil {
				b.Fatal(err)
			}
		}
		b.ReportMetric(float64(time.Since(start).Microseconds())/1000/float64(b.N), metric)
	}
	apply("apply-ms", benchID(1, 7))
	apply("apply-10k-ms", benchID(1, sourcesN+1))
}
