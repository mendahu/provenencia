package searchindex

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
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
	if version < 18 {
		t.Fatalf("user_version %d, want >= 18", version)
	}
	for _, table := range []string{"catalog_search_docs", "catalog_search_fts", "catalog_search_meta"} {
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
}
