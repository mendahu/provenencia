package effects_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// TestEvidenceGraphsResolve checks Resolve over the changes of real writes: a
// graph or link row flags the graph list and scopes its Source's activity;
// each write names the Evidence graphs whose loaded rows it moves (a Subject
// and its Observations their one graph, a Citation the graphs of its
// Observations, a Source title every graph of the Source, an Artifact and a
// link row none); and the keyed reads name the same graphs (a Citation create
// moves citations, a Subject create moves graph progress, an Artifact label
// moves neither); no other write flags the list.
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
	baseGraph := evidencegraphstest.SoleGraph(t, c, base.ID).ID
	nameProp, err := properties.Lookup(c, "name", vocabkey.OriginProvenencia)
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

	// sameIDs fails unless got holds exactly the want ids, in any order.
	sameIDs := func(what string, got [][]byte, want ...[]byte) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s: evidence graphs %x, want %x", what, got, want)
		}
		for _, w := range want {
			if !database.ContainsBlobID(got, w) {
				t.Fatalf("%s: evidence graphs %x, want %x", what, got, want)
			}
		}
	}
	only := func(changes []rowchange.Change, entity string) []rowchange.Change {
		var out []rowchange.Change
		for _, ch := range changes {
			if ch.EntityType == entity {
				out = append(out, ch)
			}
		}
		if len(out) == 0 {
			t.Fatalf("no %s change", entity)
		}
		return out
	}
	resolve := func(what string, changes []rowchange.Change) effects.Set {
		t.Helper()
		got, err := effects.Resolve(tx.Tx, changes)
		if err != nil {
			t.Fatalf("%s: %v", what, err)
		}
		return got
	}

	src, srcChanges, err := sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Census"})
	if err != nil {
		t.Fatal(err)
	}
	srcArt, artChanges, err := artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Page"})
	if err != nil {
		t.Fatal(err)
	}
	graphChange := onlyEntity(t, srcChanges, "evidence_graph")
	linkChange := onlyEntity(t, artChanges, "evidence_graph_artifact")
	def := graphChange.EntityID

	// A Source create reports its Default graph; the link row alone reports
	// none and flags the list.
	set := resolve("graph and link", []rowchange.Change{graphChange, linkChange})
	if !set.Has(effects.ListEvidenceGraphs) {
		t.Fatal("graph and link changes did not flag the graph list")
	}
	if len(set.SourceScopes) != 1 || !bytes.Equal(set.SourceScopes[0], src.ID) {
		t.Fatalf("source scopes %x", set.SourceScopes)
	}
	sameIDs("source create", set.EvidenceGraphs, def)
	linkSet := resolve("link", []rowchange.Change{linkChange})
	if !linkSet.Has(effects.ListEvidenceGraphs) || len(linkSet.EvidenceGraphs) != 0 {
		t.Fatalf("a link row alone: list %v, graphs %x", linkSet.Has(effects.ListEvidenceGraphs), linkSet.EvidenceGraphs)
	}

	// A graph write names that graph. A rename resolves from the live row; a
	// delete resolves after its rows are gone.
	jones, jonesChanges, err := evidencegraphs.Create(tx, userID, evidencegraphs.CreateInput{
		SourceID: src.ID, Name: "Jones", ArtifactIDs: [][]byte{srcArt.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	smith, _, err := evidencegraphs.Create(tx, userID, evidencegraphs.CreateInput{
		SourceID: src.ID, Name: "Smith", ArtifactIDs: [][]byte{srcArt.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	renameChanges, err := evidencegraphs.Update(tx, userID, jones.ID, evidencegraphs.UpdateInput{Name: "Jones household"})
	if err != nil {
		t.Fatal(err)
	}
	deleteChanges, err := evidencegraphs.Delete(tx, userID, smith.ID)
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		changes []rowchange.Change
		want    []byte
	}{
		"graph create": {jonesChanges, jones.ID},
		"graph rename": {renameChanges, jones.ID},
		"graph delete": {deleteChanges, smith.ID},
	} {
		got := resolve(name, tc.changes)
		if !got.Has(effects.ListEvidenceGraphs) {
			t.Fatalf("%s did not flag the graph list", name)
		}
		if len(got.SourceScopes) != 1 || !bytes.Equal(got.SourceScopes[0], src.ID) {
			t.Fatalf("%s source scopes %x", name, got.SourceScopes)
		}
		sameIDs(name, got.EvidenceGraphs, tc.want)
	}

	// A Subject and an Observation on Jones name Jones alone.
	subjChange, err := func() ([]rowchange.Change, error) {
		_, ch, err := subjects.Create(tx, userID, subjects.CreateInput{
			GraphID: jones.ID, SubjectTypeID: personType.ID, Label: "Ada",
		}, nil)
		return ch, err
	}()
	if err != nil {
		t.Fatal(err)
	}
	sameIDs("subject on Jones", resolve("subject on Jones", subjChange).EvidenceGraphs, jones.ID)
	subjectID := only(subjChange, "subject")[0].EntityID
	cited, citChanges, err := citations.CreateWithObservations(tx, userID, citations.CreateInput{
		ArtifactID: srcArt.ID, LocatorJSON: evidenceLocator,
	}, []observations.Input{{SubjectID: subjectID, PropertyID: nameProp.ID, Name: &namevalue.Value{Form: "Ada"}}})
	if err != nil {
		t.Fatal(err)
	}
	_ = cited
	sameIDs("observation on Jones", resolve("observation on Jones", only(citChanges, "observation")).EvidenceGraphs, jones.ID)
	sameIDs("citation", resolve("citation", only(citChanges, "citation")).EvidenceGraphs, jones.ID)

	// readIDs is the graphs a batch moved one keyed read for.
	readIDs := func(set effects.Set, read effects.Read) [][]byte {
		for _, r := range set.Reads {
			if r.Read == read {
				return r.IDs
			}
		}
		return nil
	}
	citSet := resolve("citation create", citChanges)
	sameIDs("citation create reads", readIDs(citSet, effects.ReadCitations), jones.ID)
	sameIDs("citation create graphs", citSet.EvidenceGraphs, jones.ID)
	sameIDs("subject create progress", readIDs(resolve("subject", subjChange), effects.ReadEvidenceGraphProgress), jones.ID)
	claim := rowchange.Change{
		EntityType: "identity_claim", EntityID: []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
		Action: rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{"subject_id": subjectID}),
	}
	sameIDs("claim create progress", readIDs(resolve("claim", []rowchange.Change{claim}), effects.ReadEvidenceGraphProgress), jones.ID)

	// A Source title edit names every graph of the Source; an Artifact create
	// and label edit name none and move no progress.
	titleChanges, err := sources.Update(tx, userID, sources.Source{ID: src.ID, SourceTypeID: typeID, Title: "Census 1851"})
	if err != nil {
		t.Fatal(err)
	}
	sameIDs("title edit", resolve("title edit", titleChanges).EvidenceGraphs, def, jones.ID)
	_, newArtChanges, err := artifacts.Create(tx, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Page 2"})
	if err != nil {
		t.Fatal(err)
	}
	artSet := resolve("artifact create", only(newArtChanges, "artifact"))
	sameIDs("artifact create", artSet.EvidenceGraphs)
	sameIDs("artifact create progress", readIDs(artSet, effects.ReadEvidenceGraphProgress))
	sameIDs("artifact create reads", readIDs(artSet, effects.ReadCitations))
	labelChanges, err := artifacts.Update(tx, userID, artifacts.Artifact{
		ID: srcArt.ID, SourceID: src.ID, Label: "Page one",
	}, evidencegraphs.GraphLinks{})
	if err != nil {
		t.Fatal(err)
	}
	labelSet := resolve("artifact label", only(labelChanges, "artifact"))
	sameIDs("artifact label", labelSet.EvidenceGraphs)
	if len(labelSet.Reads) != 0 {
		t.Fatalf("artifact label moved reads %v", labelSet.Reads)
	}

	// No other write flags the list.
	_, otherChanges, err := subjects.Create(tx, userID, subjects.CreateInput{
		GraphID: baseGraph, SubjectTypeID: personType.ID, Label: "Ada",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, baseCit, err := citations.CreateWithObservations(tx, userID, citations.CreateInput{
		ArtifactID: baseArt.ID, LocatorJSON: evidenceLocator,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	baseTitle, err := sources.Update(tx, userID, sources.Source{
		ID: base.ID, SourceTypeID: typeID, Title: "Renamed",
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, changes := range map[string][]rowchange.Change{
		"subject create":  otherChanges,
		"citation create": baseCit,
		"title edit":      baseTitle,
	} {
		if resolve(name, changes).Has(effects.ListEvidenceGraphs) {
			t.Fatalf("%s flagged the graph list", name)
		}
	}
}
