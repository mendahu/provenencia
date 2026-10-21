// Package evidencegraphstest holds shared Evidence graph fixtures: a catalog
// whose Default graph has no ref, and a Source's one graph.
package evidencegraphstest

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// SeedRefless puts one Source in c, clears its Default graph's ref, closes c
// and returns the Source id.
func SeedRefless(t *testing.T, c *database.Catalog) []byte {
	t.Helper()
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	if err := sourcevocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typ, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typ.ID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE evidence_graphs SET ref = NULL`); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	return src.ID
}

// SoleGraph is the one graph of a Source; the test fails if it has none or
// more than one. A graph with no ref reads Ref "".
func SoleGraph(t *testing.T, c *database.Catalog, sourceID []byte) evidencegraphs.Graph {
	t.Helper()
	graphs, err := evidencegraphs.List(c, sourceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(graphs) != 1 {
		t.Fatalf("source %x has %d graphs, want 1", sourceID, len(graphs))
	}
	return graphs[0]
}

// Create makes a second graph of a Source linked to the given Artifacts and
// fails the test on error.
func Create(t testing.TB, c *database.Catalog, userID, sourceID []byte, name string, artifactIDs ...[]byte) evidencegraphs.Graph {
	t.Helper()
	g, err := writes.Call(c, writes.Op{Action: "create_evidence_graph", UserID: userID}, func(tx *database.Tx) (evidencegraphs.Graph, []rowchange.Change, error) {
		return evidencegraphs.Create(tx, userID, evidencegraphs.CreateInput{
			SourceID: sourceID, Name: name, ArtifactIDs: artifactIDs,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// Split moves the given Artifacts out of the Source's Default graph into a new
// graph of the given name, and returns both. The Source must have one graph
// when it is called; the test fails on error.
func Split(t *testing.T, c *database.Catalog, userID, sourceID []byte, name string, artifactIDs ...[]byte) (def, graph evidencegraphs.Graph) {
	t.Helper()
	def = SoleGraph(t, c, sourceID)
	graph = Create(t, c, userID, sourceID, name, artifactIDs...)
	for _, id := range artifactIDs {
		_, _, err := writes.Run(c, writes.Op{Action: "unlink_artifact", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := evidencegraphs.Unlink(tx, userID, def.ID, id)
			return struct{}{}, changes, err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return def, graph
}
