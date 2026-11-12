package deleteimpact_test

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

// Facets that were silently CASCADEd before facet release are now removed
// explicitly and land in the parent's delete revision.
func TestFacetReleaseAuditsSourceSide(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcecredibilitygrades.Install(c); err != nil {
		t.Fatal(err)
	}
	grade, err := sourcecredibilitygrades.Lookup(c, "standard", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	fieldID, err := metadatafields.Upsert(c, metadatafields.Field{
		Key: "shelfmark", Origin: vocabkey.OriginUser, Label: "Shelfmark", DataType: "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	other, err := metadatafields.Upsert(c, metadatafields.Field{
		Key: "box", Origin: vocabkey.OriginUser, Label: "Box", DataType: "text",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("source delete audits notes, metadata, credibility, layout", func(t *testing.T) {
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Deed"})
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := writes.Run(c, writes.Op{Action: "create_source_note", UserID: userID}, func(tx *database.Tx) (sources.Note, []rowchange.Change, error) {
			return sources.AddNote(tx, userID, src.ID, "Found in the attic")
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := runMetadataSet(c, userID, src.ID, sourcemetadata.Input{FieldID: fieldID, ValueText: "MS 12"}); err != nil {
			t.Fatal(err)
		}
		if err := runDismissMetadata(c, userID, src.ID, other); err != nil {
			t.Fatal(err)
		}
		p6in2 := sourcecredibility.UpsertInput{
			SourceID: src.ID, CredibilityGradeID: grade.ID,
		}
		p6act1 := "create_source_credibility_assessment"
		if _, p6look3 := sourcecredibility.GetBySource(c, p6in2.SourceID); p6look3 == nil {
			p6act1 = "update_source_credibility_assessment"
		}
		if _, err := writes.Call(c, writes.Op{Action: p6act1, UserID: userID}, func(tx *database.Tx) (sourcecredibility.Assessment, []rowchange.Change, error) {
			return sourcecredibility.Upsert(tx, userID, p6in2)
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, err := writes.Run(c, writes.Op{Action: "delete_source", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sources.Delete(tx, userID, src.ID)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		db := dbtest.MustDB(t, c)
		action, revision := dbtest.LatestRevision(t, db)
		types := dbtest.EntityTypes(revision)
		if action != "delete_source" || types[len(types)-1] != "source" {
			t.Fatalf("%s %v", action, types)
		}
		got := map[string]bool{}
		for _, ty := range types[:len(types)-1] {
			got[ty] = true
		}
		for _, want := range []string{"source_note", "source_metadata", "source_credibility_assessment", "source_metadata_layout"} {
			if !got[want] {
				t.Errorf("missing %s in %v", want, types)
			}
		}
	})

	// A Source with no Artifacts keeps its default graph; the graph's metadata
	// is deleted audited, ahead of the graph's own delete (AC-3.3).
	graphSource := func(t *testing.T, title string) (sources.Source, sourcemetadata.Owner) {
		t.Helper()
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
		})
		if err != nil {
			t.Fatal(err)
		}
		return src, sourcemetadata.Owner{Level: vocabkey.LevelGraph, ID: evidencegraphstest.SoleGraph(t, c, src.ID).ID}
	}
	deleteSource := func(t *testing.T, id []byte) []string {
		t.Helper()
		if _, _, err := writes.Run(c, writes.Op{Action: "delete_source", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sources.Delete(tx, userID, id)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		_, revision := dbtest.LatestRevision(t, dbtest.MustDB(t, c))
		return dbtest.EntityTypes(revision)
	}

	t.Run("source delete audits its graph's metadata", func(t *testing.T) {
		src, graph := graphSource(t, "Census")
		runOwnerSet(t, c, userID, graph, fieldID, "14")
		got := deleteSource(t, src.ID)
		// The Source's own default-graph row and the value, layout and graph
		// deletes, in that order, then the Source.
		want := []string{"evidence_graph_metadata", "evidence_graph_metadata_layout", "evidence_graph", "source"}
		if !strings.HasSuffix(strings.Join(got, ","), strings.Join(want, ",")) {
			t.Fatalf("types %v, want a tail of %v", got, want)
		}
		for _, table := range []string{"evidence_graph_metadata", "evidence_graph_metadata_layout"} {
			if n := dbtest.Count(t, dbtest.MustDB(t, c), `SELECT COUNT(*) FROM `+table); n != 0 {
				t.Fatalf("%s left %d rows", table, n)
			}
		}
	})

	t.Run("source delete releases two graphs in id order", func(t *testing.T) {
		src, g1 := graphSource(t, "Census")
		g2ID, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := dbtest.MustDB(t, c).Exec(`INSERT INTO evidence_graphs (id, source_id, name) VALUES (?, ?, 'Second')`, g2ID[:], src.ID); err != nil {
			t.Fatal(err)
		}
		g2 := sourcemetadata.Owner{Level: vocabkey.LevelGraph, ID: g2ID[:]}
		runOwnerSet(t, c, userID, g1, fieldID, "1")
		runOwnerSet(t, c, userID, g2, fieldID, "2")
		first, second := g1, g2
		if bytes.Compare(g1.ID, g2.ID) > 0 {
			first, second = g2, g1
		}
		if _, _, err := writes.Run(c, writes.Op{Action: "delete_source", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sources.Delete(tx, userID, src.ID)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		_, revision := dbtest.LatestRevision(t, dbtest.MustDB(t, c))
		var owners []string
		for _, ch := range revision {
			switch ch.EntityType {
			case "evidence_graph_metadata":
				owners = append(owners, "value:"+fmt.Sprint(ch.Fields["graph_id"].Old))
			case "evidence_graph":
				owners = append(owners, "graph:"+database.UUIDString(ch.EntityID))
			}
		}
		want := []string{
			"value:" + database.UUIDString(first.ID), "graph:" + database.UUIDString(first.ID),
			"value:" + database.UUIDString(second.ID), "graph:" + database.UUIDString(second.ID),
		}
		if !slices.Equal(owners, want) {
			t.Fatalf("order %v, want %v", owners, want)
		}
	})

	t.Run("metadata field delete audits per-source layout rows", func(t *testing.T) {
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Will"})
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := runDismissMetadata(c, userID, src.ID, other); err != nil {
			t.Fatal(err)
		}
		if err := runFieldDelete(c, userID, other); err != nil {
			t.Fatal(err)
		}
		db := dbtest.MustDB(t, c)
		action, revision := dbtest.LatestRevision(t, db)
		types := dbtest.EntityTypes(revision)
		if action != "delete_metadata_field" || strings.Join(types, ",") != "source_metadata_layout,metadata_field" {
			t.Fatalf("%s %v", action, types)
		}
	})

	t.Run("citation delete audits its notes", func(t *testing.T) {
		src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
		})
		if err != nil {
			t.Fatal(err)
		}
		art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
		if err != nil {
			t.Fatal(err)
		}
		graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
		res, err := writes.Call(c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
			return citations.CreateWithObservations(tx, userID, citations.CreateInput{
				GraphID:    graph,
				ArtifactID: art.ID, LocatorJSON: testLocator, Notes: []string{"Faded ink", "Second hand"},
			}, nil)
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := writes.Call(c, writes.Op{Action: "delete_citation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := citations.Delete(tx, userID, res.Citation.ID)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		db := dbtest.MustDB(t, c)
		action, revision := dbtest.LatestRevision(t, db)
		types := dbtest.EntityTypes(revision)
		if action != "delete_citation" || strings.Join(types, ",") != "citation_note,citation_note,citation" {
			t.Fatalf("%s %v", action, types)
		}
	})
}

// effects.Handles on the released changes is the set subject delete recomputes.
func TestReleaseFacetsReportsHandles(t *testing.T) {
	f := newConclusionFixture(t)
	tx, err := f.c.BeginWrite()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
		f.c.EndWrite()
	}()
	// B's claim is released: one claim change, one handle.
	got, err := deleteimpact.ReleaseFacets(tx, catalogmodel.KindSubject, f.b.ID)
	if err != nil {
		t.Fatal(err)
	}
	handles, err := effects.Handles(tx.Tx, got)
	if err != nil {
		t.Fatal(err)
	}
	if len(handles) != 1 || !bytes.Equal(handles[0], f.entityID) {
		t.Fatalf("handles %v", handles)
	}
	var types []string
	for _, ch := range got {
		types = append(types, ch.EntityType)
	}
	sort.Strings(types)
	if strings.Join(types, ",") != "identity_claim" {
		t.Fatalf("changes %v", types)
	}
}

func runMetadataSet(c *database.Catalog, userID, sourceID []byte, in sourcemetadata.Input) (sourcemetadata.Row, error) {
	row, _, err := writes.Run(c, writes.Op{Action: "update_metadata", UserID: userID},
		func(tx *database.Tx) (sourcemetadata.Row, []rowchange.Change, error) {
			return sourcemetadata.Set(tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: sourceID}, in)
		})
	return row, err
}

func runDismissMetadata(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "dismiss_metadata_suggestion", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sourcemetadata.DismissSuggestion(tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: sourceID}, fieldID)
			return struct{}{}, changes, err
		})
	return err
}

func runFieldDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_metadata_field", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := metadatafields.Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}

// releaseCitation lists the Citation's columns by hand (the import cycle keeps
// it from calling citations); its delete change must name every column, so a
// new column fails here until the release carries it.
func TestReleaseCitationChangeNamesEveryColumn(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	graph := evidencegraphstest.SoleGraph(t, c, src.ID).ID
	res, _, err := observationstest.Cite(c, userID, graph, art.ID, testLocator)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := c.BeginWrite()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
		c.EndWrite()
	}()
	changes, err := deleteimpact.ReleaseCitation(tx, res.Citation.ID)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, ch := range changes {
		if ch.EntityType != "citation" || ch.Action != rowchange.ActionDelete {
			continue
		}
		for k := range ch.Fields {
			got = append(got, k)
		}
	}
	rows, err := tx.Query(`SELECT name FROM pragma_table_info('citations')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var want []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		want = append(want, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("delete change fields %v, citations columns %v", got, want)
	}
}

// ownerFixture is a Source of a book type holding one Artifact, with its sole
// graph, for the Artifact and graph metadata cases.
func ownerFixture(t *testing.T, c *database.Catalog, userID []byte, title string) (src sources.Source, artifact, graph sourcemetadata.Owner) {
	t.Helper()
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err = writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	return src,
		sourcemetadata.Owner{Level: vocabkey.LevelArtifact, ID: art.ID},
		sourcemetadata.Owner{Level: vocabkey.LevelGraph, ID: evidencegraphstest.SoleGraph(t, c, src.ID).ID}
}

func runOwnerSet(t *testing.T, c *database.Catalog, userID []byte, o sourcemetadata.Owner, fieldID []byte, text string) {
	t.Helper()
	if _, _, err := writes.Run(c, writes.Op{Action: "update_metadata", UserID: userID},
		func(tx *database.Tx) (sourcemetadata.Row, []rowchange.Change, error) {
			return sourcemetadata.Set(tx, userID, o, sourcemetadata.Input{FieldID: fieldID, ValueText: text})
		}); err != nil {
		t.Fatal(err)
	}
}

func runOwnerDismiss(t *testing.T, c *database.Catalog, userID []byte, o sourcemetadata.Owner, fieldID []byte) {
	t.Helper()
	if _, _, err := writes.Run(c, writes.Op{Action: "dismiss_metadata_suggestion", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sourcemetadata.DismissSuggestion(tx, userID, o, fieldID)
			return struct{}{}, changes, err
		}); err != nil {
		t.Fatal(err)
	}
}

// TestImpactMetadataFieldHolders pins the field delete's inbound groups: a
// value on any of the three levels blocks, once per holder, in level order.
func TestImpactMetadataFieldHolders(t *testing.T) {
	c, userID := testCatalog(t)
	fieldID, err := metadatafields.Upsert(c, metadatafields.Field{Key: "shelfmark", Origin: vocabkey.OriginUser, Label: "Shelfmark", DataType: "text"})
	if err != nil {
		t.Fatal(err)
	}
	s1, a1, g1 := ownerFixture(t, c, userID, "S1")

	// AC-2.1: a graph alone.
	runOwnerSet(t, c, userID, g1, fieldID, "14")
	got := mustImpact(t, c, catalogmodel.KindMetadataField, fieldID)
	if got.Allowed || got.Gate != deleteimpact.GateInbound || len(got.Groups) != 1 {
		t.Fatalf("graph only: %+v", got)
	}
	g := got.Groups[0]
	if g.Via != "evidence_graph_metadata.field_id" || g.Kind != catalogmodel.KindEvidenceGraph || g.Total != 1 || len(g.Listed) != 1 ||
		!bytes.Equal(g.Listed[0].ID, g1.ID) {
		t.Fatalf("graph group %+v", g)
	}

	// AC-2.2: two Sources, an Artifact and a graph.
	s2, _, _ := ownerFixture(t, c, userID, "S2")
	runOwnerSet(t, c, userID, a1, fieldID, "p")
	runOwnerSet(t, c, userID, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: s1.ID}, fieldID, "x")
	runOwnerSet(t, c, userID, sourcemetadata.Owner{Level: vocabkey.LevelSource, ID: s2.ID}, fieldID, "y")
	got = mustImpact(t, c, catalogmodel.KindMetadataField, fieldID)
	var vias []string
	var totals []int
	for _, g := range got.Groups {
		vias = append(vias, g.Via)
		totals = append(totals, g.Total)
	}
	if strings.Join(vias, ",") != "source_metadata.field_id,artifact_metadata.field_id,evidence_graph_metadata.field_id" ||
		len(totals) != 3 || totals[0] != 2 || totals[1] != 1 || totals[2] != 1 {
		t.Fatalf("groups %v %v", vias, totals)
	}
}
