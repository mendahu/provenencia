package evidencegraphs_test

import (
	"encoding/binary"
	"strconv"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func benchID(kind byte, n int) []byte {
	id := make([]byte, 16)
	id[0] = kind
	binary.BigEndian.PutUint64(id[8:], uint64(n))
	return id
}

// BenchmarkGraphIndex: Build of the index at 3,000 Sources, 10,000 graphs each
// linked to an Artifact and holding a Citation, 20,000 person Subjects and
// 50,000 Observations; then Apply of one commit that recounts a graph of
// 10,000 Observations and one that re-lists a Source of 10,000 graphs.
func BenchmarkGraphIndex(b *testing.B) {
	const sourcesN, graphsN, subjectsN, observationsN, crafted = 3000, 10000, 20000, 50000, 10000
	c := graphcachetest.OpenCatalog(b)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		b.Fatal(err)
	}
	personType := graphcachetest.SubjectTypeID(b, c, "person")
	db, err := c.DB()
	if err != nil {
		b.Fatal(err)
	}
	var propID []byte
	if err := db.QueryRow(`SELECT id FROM properties WHERE key = 'remark' AND origin = 'provenencia'`).Scan(&propID); err != nil {
		b.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		b.Fatal(err)
	}
	exec := func(q string, args ...any) {
		if _, err := tx.Exec(q, args...); err != nil {
			b.Fatal(err)
		}
	}
	for s := 1; s <= sourcesN+2; s++ {
		exec(`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, ?, ?, ?)`, benchID(1, s), "SRC-B"+strconv.Itoa(s), typeID, "Source "+strconv.Itoa(s))
		exec(`INSERT INTO artifacts (id, ref, source_id, label) VALUES (?, ?, ?, ?)`, benchID(2, s), "ART-B"+strconv.Itoa(s), benchID(1, s), "Page")
	}
	graph := func(i, source int) {
		exec(`INSERT INTO evidence_graphs (id, ref, source_id, name) VALUES (?, ?, ?, ?)`, benchID(3, i), "GPH-B"+strconv.Itoa(i), benchID(1, source), "Graph "+strconv.Itoa(i))
		exec(`INSERT INTO evidence_graph_artifacts (graph_id, artifact_id) VALUES (?, ?)`, benchID(3, i), benchID(2, source))
		exec(`INSERT INTO citations (id, ref, artifact_id, graph_id, locator_json) VALUES (?, ?, ?, ?, '{}')`, benchID(4, i), "CIT-B"+strconv.Itoa(i), benchID(2, source), benchID(3, i))
	}
	subject := func(i, graphN, source int) {
		exec(`INSERT INTO subjects (id, ref, source_id, graph_id, subject_type_id) VALUES (?, ?, ?, ?, ?)`, benchID(5, i), "SUB-B"+strconv.Itoa(i), benchID(1, source), benchID(3, graphN), personType)
	}
	obs := func(i, subjectN, graphN int) {
		exec(`INSERT INTO observations (id, ref, citation_id, subject_id, property_id, value_text) VALUES (?, ?, ?, ?, ?, 'x')`, benchID(6, i), "OBS-B"+strconv.Itoa(i), benchID(4, graphN), benchID(5, subjectN), propID)
	}
	for i := 1; i <= graphsN; i++ {
		graph(i, (i-1)%sourcesN+1)
	}
	for i := 1; i <= subjectsN; i++ {
		g := (i-1)%graphsN + 1
		subject(i, g, (g-1)%sourcesN+1)
	}
	for i := 1; i <= observationsN; i++ {
		obs(i, (i-1)%subjectsN+1, ((i-1)%subjectsN)%graphsN+1)
	}
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}

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
	if tx, err = db.Begin(); err != nil {
		b.Fatal(err)
	}
	heavy := graphsN + 1
	graph(heavy, sourcesN+1)
	subject(subjectsN+1, heavy, sourcesN+1)
	for i := 1; i <= crafted; i++ {
		obs(observationsN+i, subjectsN+1, heavy)
	}
	wide := sourcesN + 2
	for i := 1; i <= crafted; i++ {
		exec(`INSERT INTO evidence_graphs (id, ref, source_id, name) VALUES (?, ?, ?, ?)`, benchID(3, heavy+i), "GPH-B"+strconv.Itoa(heavy+i), benchID(1, wide), "Wide "+strconv.Itoa(i))
	}
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}
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
	apply("apply-progress-ms", effects.ReadEvidenceGraphProgress, benchID(3, 7))
	apply("apply-progress-10k-ms", effects.ReadEvidenceGraphProgress, benchID(3, heavy))
	apply("apply-list-ms", effects.ReadEvidenceGraphs, benchID(1, 7))
	apply("apply-list-10k-ms", effects.ReadEvidenceGraphs, benchID(1, wide))
}
