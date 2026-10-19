package effects_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// TestEvidenceGraphsResolve checks Resolve over the changes of real writes: a
// graph or link row flags the graph list and scopes its Source's activity
// without moving the Source's evidence; no other write flags the list.
func TestEvidenceGraphsResolve(t *testing.T) {
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
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	base, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID},
		func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Base"})
		})
	if err != nil {
		t.Fatal(err)
	}
	baseArt, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: base.ID, Label: "Scan"})
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

	src, srcChanges, err := sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Census"})
	if err != nil {
		t.Fatal(err)
	}
	_, artChanges, err := artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Page"})
	if err != nil {
		t.Fatal(err)
	}
	graphChange := onlyEntity(t, srcChanges, "evidence_graph")
	linkChange := onlyEntity(t, artChanges, "evidence_graph_artifact")

	set, err := effects.Resolve(tx.Tx, []rowchange.Change{graphChange, linkChange})
	if err != nil {
		t.Fatal(err)
	}
	if !set.Has(effects.ListEvidenceGraphs) {
		t.Fatal("graph and link changes did not flag the graph list")
	}
	if len(set.SourceScopes) != 1 || !bytes.Equal(set.SourceScopes[0], src.ID) {
		t.Fatalf("source scopes %x", set.SourceScopes)
	}
	if len(set.Sources) != 0 {
		t.Fatalf("graph and link rows moved Source evidence: %x", set.Sources)
	}

	_, subjChanges, err := subjects.Create(tx, userID, subjects.CreateInput{
		SourceID: base.ID, SubjectTypeID: personType.ID, Label: "Ada",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, citChanges, err := citations.CreateWithObservations(tx, userID, citations.CreateInput{
		ArtifactID: baseArt.ID, LocatorJSON: evidenceLocator,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	titleChanges, err := sources.Update(tx, userID, sources.Source{
		ID: base.ID, SourceTypeID: typeID, Title: "Renamed",
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, changes := range map[string][]rowchange.Change{
		"subject create":  subjChanges,
		"citation create": citChanges,
		"title edit":      titleChanges,
	} {
		got, err := effects.Resolve(tx.Tx, changes)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got.Has(effects.ListEvidenceGraphs) {
			t.Fatalf("%s flagged the graph list", name)
		}
	}
}
