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
