package deleteimpact_test

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/catalogmodel"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibility"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/writes"
)

// Facets that were silently CASCADEd before facet release are now removed
// explicitly and land in the parent's delete revision.
func TestFacetReleaseAuditsSourceSide(t *testing.T) {
	c, userID := testCatalog(t)
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcecredibilitygrades.Install(c); err != nil {
		t.Fatal(err)
	}
	grade, err := sourcecredibilitygrades.Lookup(c, "standard", sourcecredibilitygrades.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	fieldID, err := metadatafields.Upsert(c, metadatafields.Field{
		Key: "shelfmark", Origin: metadatafields.OriginUser, Label: "Shelfmark", DataType: "text",
	})
	if err != nil {
		t.Fatal(err)
	}
	other, err := metadatafields.Upsert(c, metadatafields.Field{
		Key: "box", Origin: metadatafields.OriginUser, Label: "Box", DataType: "text",
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
		if _, err := runMetadataSet(c, userID, sourcemetadata.Input{SourceID: src.ID, FieldID: fieldID, ValueText: "MS 12"}); err != nil {
			t.Fatal(err)
		}
		if err := runDismissMetadata(c, userID, src.ID, other); err != nil {
			t.Fatal(err)
		}
		if _, err := sourcecredibility.Upsert(c, userID, sourcecredibility.UpsertInput{
			SourceID: src.ID, CredibilityGradeID: grade.ID,
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, err := writes.Run(c, writes.Op{Action: "delete_source", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sources.Delete(tx, userID, src.ID)
			return struct{}{}, changes, err
		}); err != nil {
			t.Fatal(err)
		}
		action, types := lastRevision(t, c)
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
		action, types := lastRevision(t, c)
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
		art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
		if err != nil {
			t.Fatal(err)
		}
		res, err := citations.CreateWithObservations(c, userID, citations.CreateInput{
			ArtifactID: art.ID, LocatorJSON: testLocator, Notes: []string{"Faded ink", "Second hand"},
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := citations.Delete(c, userID, res.Citation.ID); err != nil {
			t.Fatal(err)
		}
		action, types := lastRevision(t, c)
		if action != "delete_citation" || strings.Join(types, ",") != "citation_note,citation_note,citation" {
			t.Fatalf("%s %v", action, types)
		}
	})
}

// Released.Handles is the seam cache upkeep and search reprojection will read.
func TestReleaseFacetsReportsHandles(t *testing.T) {
	f := newConclusionFixture(t)
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	// obsA is pinned on both claims of one handle: one handle, reported once.
	got, err := deleteimpact.ReleaseFacets(tx, catalogmodel.KindObservation, f.obsA)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Handles) != 1 || !bytes.Equal(got.Handles[0], f.entityID) {
		t.Fatalf("handles %v", got.Handles)
	}
	var types []string
	for _, ch := range got.Changes {
		types = append(types, ch.EntityType)
	}
	sort.Strings(types)
	if strings.Join(types, ",") != "identity_claim_evidence,identity_claim_evidence" {
		t.Fatalf("changes %v", types)
	}
}

func runMetadataSet(c *database.Catalog, userID []byte, in sourcemetadata.Input) (sourcemetadata.Row, error) {
	row, _, err := writes.Run(c, writes.Op{Action: "update_source_metadata", UserID: userID},
		func(tx *database.Tx) (sourcemetadata.Row, []rowchange.Change, error) {
			return sourcemetadata.Set(tx, userID, in)
		})
	return row, err
}

func runDismissMetadata(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "dismiss_source_metadata_suggestion", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sourcemetadata.DismissSuggestion(tx, userID, sourceID, fieldID)
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
