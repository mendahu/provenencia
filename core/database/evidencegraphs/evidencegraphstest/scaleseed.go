package evidencegraphstest

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/databasetest"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The batch's Scale (spike 16): what Base files.
const (
	ScaleSources      = 3_000
	ScaleGraphs       = 10_000
	ScaleSubjects     = 20_000
	ScaleObservations = 50_000
)

// ScaleSeed writes the Scale catalog the graph benchmarks measure, one
// transaction of direct inserts. Ids are databasetest.BenchID of kind 1 (Source),
// 2 (Artifact), 3 (graph), 4 (Citation), 5 (Subject) and 6 (Observation), and
// refs follow from the number.
type ScaleSeed struct {
	tb testing.TB
	tx *sql.Tx
	// PersonTypeID is the person subject type, PropID the remark Property and
	// TypeID the book source type every Source takes.
	PersonTypeID, PropID, TypeID []byte
}

// NewScaleSeed looks up the person type and the remark property, upserts the
// book source type and opens the transaction. It seeds nothing.
func NewScaleSeed(tb testing.TB, c *database.Catalog) *ScaleSeed {
	tb.Helper()
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		tb.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		tb.Fatal(err)
	}
	s := &ScaleSeed{tb: tb, TypeID: typeID, PersonTypeID: graphcachetest.SubjectTypeID(tb, c, "person")}
	if err := db.QueryRow(`SELECT id FROM properties WHERE key = 'remark' AND origin = 'provenencia'`).Scan(&s.PropID); err != nil {
		tb.Fatal(err)
	}
	if s.tx, err = db.Begin(); err != nil {
		tb.Fatal(err)
	}
	return s
}

// Exec runs one statement in the transaction.
func (s *ScaleSeed) Exec(q string, args ...any) {
	s.tb.Helper()
	if _, err := s.tx.Exec(q, args...); err != nil {
		s.tb.Fatal(err)
	}
}

// Source files Source i and its one Artifact.
func (s *ScaleSeed) Source(i int) {
	s.tb.Helper()
	s.Exec(`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, ?, ?, ?)`, databasetest.BenchID(1, i), "SRC-B"+strconv.Itoa(i), s.TypeID, "Source "+strconv.Itoa(i))
	s.Exec(`INSERT INTO artifacts (id, ref, source_id, label) VALUES (?, ?, ?, ?)`, databasetest.BenchID(2, i), "ART-B"+strconv.Itoa(i), databasetest.BenchID(1, i), "Page")
}

// Graph files graph i in Source number source, linked to that Source's
// Artifact, with one Citation (id kind 4, number i).
func (s *ScaleSeed) Graph(i, source int) {
	s.tb.Helper()
	s.Exec(`INSERT INTO evidence_graphs (id, ref, source_id, name) VALUES (?, ?, ?, ?)`, databasetest.BenchID(3, i), "GPH-B"+strconv.Itoa(i), databasetest.BenchID(1, source), "Graph "+strconv.Itoa(i))
	s.Exec(`INSERT INTO evidence_graph_artifacts (graph_id, artifact_id) VALUES (?, ?)`, databasetest.BenchID(3, i), databasetest.BenchID(2, source))
	s.Exec(`INSERT INTO citations (id, ref, artifact_id, graph_id, locator_json) VALUES (?, ?, ?, ?, '{}')`, databasetest.BenchID(4, i), "CIT-B"+strconv.Itoa(i), databasetest.BenchID(2, source), databasetest.BenchID(3, i))
}

// Subject files person Subject i on graph number graph, in Source number source.
func (s *ScaleSeed) Subject(i, graph, source int) {
	s.tb.Helper()
	s.Exec(`INSERT INTO subjects (id, ref, source_id, graph_id, subject_type_id) VALUES (?, ?, ?, ?, ?)`, databasetest.BenchID(5, i), "SUB-B"+strconv.Itoa(i), databasetest.BenchID(1, source), databasetest.BenchID(3, graph), s.PersonTypeID)
}

// Obs files remark Observation i on Subject number subject, cited by the
// Citation of graph number graph.
func (s *ScaleSeed) Obs(i, subject, graph int) {
	s.tb.Helper()
	s.Exec(`INSERT INTO observations (id, ref, citation_id, subject_id, property_id, value_text) VALUES (?, ?, ?, ?, ?, 'x')`, databasetest.BenchID(6, i), "OBS-B"+strconv.Itoa(i), databasetest.BenchID(4, graph), databasetest.BenchID(5, subject), s.PropID)
}

// Base files the Scale: Sources 1 to ScaleSources+2 (the last two are left for
// the crafted rows of a benchmark), ScaleGraphs graphs spread over the first
// ScaleSources, ScaleSubjects Subjects spread over the graphs and
// ScaleObservations Observations spread over the Subjects.
func (s *ScaleSeed) Base() {
	s.tb.Helper()
	for i := 1; i <= ScaleSources+2; i++ {
		s.Source(i)
	}
	for i := 1; i <= ScaleGraphs; i++ {
		s.Graph(i, (i-1)%ScaleSources+1)
	}
	for i := 1; i <= ScaleSubjects; i++ {
		g := (i-1)%ScaleGraphs + 1
		s.Subject(i, g, (g-1)%ScaleSources+1)
	}
	for i := 1; i <= ScaleObservations; i++ {
		s.Obs(i, (i-1)%ScaleSubjects+1, ((i-1)%ScaleSubjects)%ScaleGraphs+1)
	}
}

// Commit commits the transaction.
func (s *ScaleSeed) Commit() {
	s.tb.Helper()
	if err := s.tx.Commit(); err != nil {
		s.tb.Fatal(err)
	}
}
