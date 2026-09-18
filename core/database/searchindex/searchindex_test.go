package searchindex

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/objectpath"
)

func TestMigrationCreatesSearchTables(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version < 20 {
		t.Fatalf("user_version %d, want >= 20", version)
	}
	for _, table := range []string{
		"catalog_search_docs",
		"catalog_search_fts",
		"catalog_search_fts_trigram",
		"catalog_search_meta",
	} {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type IN ('table','view') AND name = ?`,
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("missing %s: %v", table, err)
		}
	}
	// FTS5 must accept a write (requires -tags fts5).
	if err := Upsert(db, Document{
		Kind:         KindSourceType,
		EntityID:     "00000000-0000-0000-0000-000000000001",
		DisplayTitle: "Smoke",
		Title:        "Smoke",
		Secondary:    "smoke-key",
	}); err != nil {
		t.Fatalf("fts upsert: %v", err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM catalog_search_fts WHERE catalog_search_fts MATCH 'smoke'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("match count %d", n)
	}
	// Trigram index is substring-oriented; three+ chars should hit.
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM catalog_search_fts_trigram WHERE catalog_search_fts_trigram MATCH 'smo'`,
	).Scan(&n); err != nil {
		t.Fatalf("trigram match: %v", err)
	}
	if n != 1 {
		t.Fatalf("trigram match count %d", n)
	}
}

func TestObjectsRelPathAgreesWithObjectpath(t *testing.T) {
	const sum = "8fce3b0000000000000000000000000000000000000000000000000000000000"
	for _, mt := range []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
		"image/gif",
		"application/pdf",
		"text/plain",
		"application/octet-stream",
		"",
	} {
		got, err := objectsRelPath(sum, mt)
		if err != nil {
			t.Fatalf("%s: %v", mt, err)
		}
		want, err := objectpath.Rel(sum, mt)
		if err != nil {
			t.Fatalf("%s objectpath: %v", mt, err)
		}
		if got != want {
			t.Fatalf("%s: searchindex %q != objectpath %q", mt, got, want)
		}
	}
}
