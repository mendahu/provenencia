package search

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestEmptyQueryReturnsNoHits(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	hits, err := DefaultEngine().Search(context.Background(), c, Query{Text: "   "})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("got %d hits", len(hits))
	}
}

func TestSearchRanksTitleOverDescriptionAndMapsLocation(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	userID := seedUser(t, c)
	typeID := seedType(t, c, "Book", "a monograph")

	titleHit, err := sources.Create(c, userID, sources.CreateInput{
		SourceTypeID: typeID,
		Title:        "Ilminster parish register",
		Description:  "misc notes",
	})
	if err != nil {
		t.Fatal(err)
	}
	descHit, err := sources.Create(c, userID, sources.CreateInput{
		SourceTypeID: typeID,
		Title:        "Other deed",
		Description:  "mentions the Ilminster area only here",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = descHit

	hits, err := DefaultEngine().Search(context.Background(), c, Query{
		Text:     "Ilminster",
		Location: WorkspaceLocation{Section: SectionSources},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 2 {
		t.Fatalf("want >= 2 hits, got %d", len(hits))
	}
	if hits[0].Kind != KindSource || hits[0].ID != uuidString(titleHit.ID) {
		t.Fatalf("want title Source first, got %+v", hits[0])
	}
	if hits[0].MatchReason != "title" {
		t.Fatalf("match_reason %q", hits[0].MatchReason)
	}
	if hits[0].Location.Section != SectionSources || hits[0].Location.SourceID != hits[0].ID {
		t.Fatalf("location %+v", hits[0].Location)
	}
	if hits[0].Score <= hits[1].Score {
		t.Fatalf("title score %v should beat description %v", hits[0].Score, hits[1].Score)
	}
}

func TestSearchFindsTypesAndFields(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	seedUser(t, c)
	ty, err := sourcetypes.Create(c, "Birth certificate", "civil record", "")
	if err != nil {
		t.Fatal(err)
	}
	field, err := sourcefields.Create(c, "Publication date", "date", "when published")
	if err != nil {
		t.Fatal(err)
	}

	typeHits, err := DefaultEngine().Search(context.Background(), c, Query{
		Text:     "birth certificate",
		Location: WorkspaceLocation{Section: SectionSourceTypes},
	})
	if err != nil {
		t.Fatal(err)
	}
	foundType := false
	for _, h := range typeHits {
		if h.Kind == KindSourceType && h.ID == uuidString(ty.ID) {
			foundType = true
			if h.Location.TypeID != h.ID || h.Location.Section != SectionSourceTypes {
				t.Fatalf("type location %+v", h.Location)
			}
		}
	}
	if !foundType {
		t.Fatalf("missing type hit in %+v", typeHits)
	}

	fieldHits, err := DefaultEngine().Search(context.Background(), c, Query{Text: "publication"})
	if err != nil {
		t.Fatal(err)
	}
	foundField := false
	for _, h := range fieldHits {
		if h.Kind == KindSourceField && h.ID == uuidString(field.ID) {
			foundField = true
			if h.Location.FieldID != h.ID {
				t.Fatalf("field location %+v", h.Location)
			}
		}
	}
	if !foundField {
		t.Fatalf("missing field hit in %+v", fieldHits)
	}
}

func TestContextBoostPrefersMatchingSection(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	seedUser(t, c)
	ty, err := sourcetypes.Create(c, "SharedToken Type", "x", "")
	if err != nil {
		t.Fatal(err)
	}
	field, err := sourcefields.Create(c, "SharedToken Field", "text", "")
	if err != nil {
		t.Fatal(err)
	}
	_ = ty
	_ = field

	onTypes, err := DefaultEngine().Search(context.Background(), c, Query{
		Text:     "SharedToken",
		Location: WorkspaceLocation{Section: SectionSourceTypes},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(onTypes) < 2 {
		t.Fatalf("hits %d", len(onTypes))
	}
	if onTypes[0].Kind != KindSourceType {
		t.Fatalf("expected type first under types context, got %s", onTypes[0].Kind)
	}

	onFields, err := DefaultEngine().Search(context.Background(), c, Query{
		Text:     "SharedToken",
		Location: WorkspaceLocation{Section: SectionSourceFields},
	})
	if err != nil {
		t.Fatal(err)
	}
	if onFields[0].Kind != KindSourceField {
		t.Fatalf("expected field first under fields context, got %s", onFields[0].Kind)
	}
}

func TestNoteBodyRollsIntoSourceHit(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	userID := seedUser(t, c)
	typeID := seedType(t, c, "Book", "")
	src, err := sources.Create(c, userID, sources.CreateInput{
		SourceTypeID: typeID,
		Title:        "Quiet title",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sources.AddNote(c, userID, src.ID, "mentions Zemblanity only in the note"); err != nil {
		t.Fatal(err)
	}
	hits, err := DefaultEngine().Search(context.Background(), c, Query{Text: "Zemblanity"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Kind != KindSource || hits[0].ID != uuidString(src.ID) {
		t.Fatalf("want Source hit for note text, got %+v", hits)
	}
	if hits[0].MatchReason != "notes" {
		t.Fatalf("match_reason %q", hits[0].MatchReason)
	}
}

func TestEnsureIndexRebuildsAfterWipe(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	userID := seedUser(t, c)
	typeID := seedType(t, c, "Book", "")
	src, err := sources.Create(c, userID, sources.CreateInput{
		SourceTypeID: typeID,
		Title:        "RebuildMe Parish",
	})
	if err != nil {
		t.Fatal(err)
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := searchindex.ClearAll(db); err != nil {
		t.Fatal(err)
	}
	if err := searchindex.SetProjectionVersion(db, 0); err != nil {
		t.Fatal(err)
	}
	hits, err := DefaultEngine().Search(context.Background(), c, Query{Text: "RebuildMe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("expected empty after wipe, got %+v", hits)
	}
	if err := EnsureIndex(c); err != nil {
		t.Fatal(err)
	}
	hits, err = DefaultEngine().Search(context.Background(), c, Query{Text: "RebuildMe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ID != uuidString(src.ID) {
		t.Fatalf("after heal got %+v", hits)
	}
}

func TestTokenize(t *testing.T) {
	got := tokenize(`  John's "birth"  `)
	if len(got) != 2 || got[0] != "john's" && got[0] != "johns" {
		if !strings.Contains(strings.Join(got, " "), "john") {
			t.Fatalf("%v", got)
		}
	}
	if tokenize("") != nil {
		t.Fatal("empty")
	}
}

func seedUser(t *testing.T, c *database.Catalog) []byte {
	t.Helper()
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	return userID
}

func seedType(t *testing.T, c *database.Catalog, label, desc string) []byte {
	t.Helper()
	id, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: label, Description: desc,
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func uuidString(id []byte) string {
	if len(id) != 16 {
		return ""
	}
	var u uuid.UUID
	copy(u[:], id)
	return u.String()
}
