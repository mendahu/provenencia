package catalogsession_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	_ "github.com/mendahu/provenencia/core/database/canonlinks"
	_ "github.com/mendahu/provenencia/core/database/conclusionsummaries"
	_ "github.com/mendahu/provenencia/core/database/evidencegraphs"
	_ "github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
	_ "github.com/mendahu/provenencia/core/database/namealiases"
	_ "github.com/mendahu/provenencia/core/database/nameformats"
	_ "github.com/mendahu/provenencia/core/database/promotealign"
	_ "github.com/mendahu/provenencia/core/database/sources"
)

// TestCacheDocListsEveryEntry holds docs/catalog-caches.md to the cache
// registry both ways: every registered entry has a row in the "Registry
// entries" table, and every row names a registered entry. This package imports
// every owner (TestCacheDocSeesEveryOwner proves it), so its registry is the
// app's. A PR that registers a cache adds its row in the same change.
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

// TestCacheDocSeesEveryOwner holds this file's imports to the packages that
// register a cache: a package under core/ or api/ that calls
// graphcache.RegisterKeyed or RegisterWhole is imported here, so the registry
// this package reads is the app's.
func TestCacheDocSeesEveryOwner(t *testing.T) {
	const module = "github.com/mendahu/provenencia/"
	owners := map[string]bool{}
	for _, root := range []string{"core", "api"} {
		err := filepath.WalkDir(filepath.Join("../..", root), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(src), "graphcache.RegisterKeyed(") || strings.Contains(string(src), "graphcache.RegisterWhole(") {
				dir, err := filepath.Rel("../..", filepath.Dir(path))
				if err != nil {
					return err
				}
				owners[module+filepath.ToSlash(dir)] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(owners) == 0 {
		t.Fatal("no package registers a cache")
	}
	file, err := parser.ParseFile(token.NewFileSet(), "cache_doc_test.go", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	imported := map[string]bool{}
	for _, imp := range file.Imports {
		imported[strings.Trim(imp.Path.Value, `"`)] = true
	}
	for owner := range owners {
		if !imported[owner] {
			t.Errorf("cache_doc_test.go does not import %s, which registers a cache", owner)
		}
	}
}
