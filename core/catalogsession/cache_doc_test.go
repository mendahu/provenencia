package catalogsession_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	_ "github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/graphcache"
	_ "github.com/mendahu/provenencia/core/database/namealiases"
	_ "github.com/mendahu/provenencia/core/database/nameformats"
	_ "github.com/mendahu/provenencia/core/database/promotealign"
)

// TestCacheDocListsEveryEntry holds docs/catalog-caches.md to the cache
// registry both ways: every registered entry has a row in the "Registry
// entries" table, and every row names a registered entry. This package imports
// every owner, so its registry is the app's. A PR that registers a cache adds
// its row in the same change.
func TestCacheDocListsEveryEntry(t *testing.T) {
	raw, err := os.ReadFile("../../docs/catalog-caches.md")
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	inSection := false
	for _, line := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(line, "## "):
			inSection = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Registry entries"
		case inSection && strings.HasPrefix(line, "|"):
			cells := strings.Split(line, "|")
			if len(cells) < 3 {
				continue
			}
			first := strings.TrimSpace(cells[1])
			if strings.HasPrefix(first, "`") && strings.HasSuffix(first, "`") {
				rows = append(rows, strings.Trim(first, "`"))
			}
		}
	}
	names := graphcache.EntryNames()
	if len(names) == 0 {
		t.Fatal("no cache entry is registered")
	}
	for _, name := range names {
		if !slices.Contains(rows, name) {
			t.Errorf("cache entry %s has no row in docs/catalog-caches.md (## Registry entries)", name)
		}
	}
	for _, row := range rows {
		if !slices.Contains(names, row) {
			t.Errorf("docs/catalog-caches.md names cache entry %s, which nothing registers", row)
		}
	}
}
