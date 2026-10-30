package effects_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// TestCitationReadsPreferArtifactGraphs pins a Citation with no Observation:
// it names no graph for graphcache, and its citation reads follow the graphs
// its Artifact is linked to (here Jones only, not Default).
func TestCitationReadsPreferArtifactGraphs(t *testing.T) {
	userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	ur, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", ur); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlTx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sqlTx.Rollback() }()
	tx := &database.Tx{Tx: sqlTx}

	src, _, err := sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Census"})
	if err != nil {
		t.Fatal(err)
	}
	other, _, err := artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	jones, _, err := evidencegraphs.Create(tx, userID, evidencegraphs.CreateInput{
		SourceID: src.ID, Name: "Jones", ArtifactIDs: [][]byte{other.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	art, _, err := artifacts.Create(tx, userID, artifacts.CreateInput{
		SourceID: src.ID, Label: "Page", Graphs: evidencegraphs.GraphLinks{Given: true, IDs: [][]byte{jones.ID}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var linked int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM evidence_graph_artifacts WHERE artifact_id = ? AND graph_id = ?`,
		art.ID, jones.ID).Scan(&linked); err != nil || linked != 1 {
		t.Fatalf("Jones link: %d %v", linked, err)
	}
	if err := tx.QueryRow(`SELECT COUNT(*) FROM evidence_graph_artifacts WHERE artifact_id = ?`, art.ID).Scan(&linked); err != nil || linked != 1 {
		t.Fatalf("artifact links: %d %v, want Jones only", linked, err)
	}
	_, changes, err := citations.CreateWithObservations(tx, userID, citations.CreateInput{
		ArtifactID: art.ID, LocatorJSON: evidenceLocator,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var cit []rowchange.Change
	for _, ch := range changes {
		if ch.EntityType == "citation" {
			cit = append(cit, ch)
		}
	}
	set, err := effects.Resolve(tx.Tx, cit)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.EvidenceGraphs) != 0 {
		t.Fatalf("EvidenceGraphs %x, want none", set.EvidenceGraphs)
	}
	if len(set.Reads) != 1 || set.Reads[0].Read != effects.ReadCitations ||
		len(set.Reads[0].IDs) != 1 || !bytes.Equal(set.Reads[0].IDs[0], jones.ID) {
		t.Fatalf("Reads %+v, want citations -> Jones", set.Reads)
	}
}
