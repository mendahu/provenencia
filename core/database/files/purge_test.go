package files_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/files"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
)

func TestPurgeUnreferenced(t *testing.T) {
	sumKept := strings.Repeat("a", 64)
	sumGone := strings.Repeat("b", 64)
	sumTwice := strings.Repeat("c", 64)

	newCatalog := func(t *testing.T) *database.Catalog {
		t.Helper()
		c, err := database.Create(t.TempDir(), "t.provenencia")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
		return c
	}
	writeObject := func(t *testing.T, c *database.Catalog, sum string) string {
		t.Helper()
		rel, err := files.StorageRelPath(sum, "image/jpeg")
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(c.Dir(), filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("bytes"), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	exists := func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}
	insertRow := func(t *testing.T, c *database.Catalog, sum string) {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		id, err := files.NewID()
		if err != nil {
			t.Fatal(err)
		}
		if err := files.Insert(tx, files.File{ID: id, ChecksumSHA256: sum, MediaType: "image/jpeg", ByteSize: 5}); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
	obj := func(sum string) files.Object { return files.Object{ChecksumSHA256: sum, MediaType: "image/jpeg"} }

	tests := []struct {
		name     string
		rows     []string // checksums that keep a files row
		objects  []string // checksums written to disk
		purge    []files.Object
		wantGone []string
		wantKept []string
	}{
		{name: "no row removes the object", objects: []string{sumGone}, purge: []files.Object{obj(sumGone)}, wantGone: []string{sumGone}},
		{name: "a row keeps the object", rows: []string{sumKept}, objects: []string{sumKept}, purge: []files.Object{obj(sumKept)}, wantKept: []string{sumKept}},
		{name: "empty list", objects: []string{sumKept}, wantKept: []string{sumKept}},
		{name: "same object twice", objects: []string{sumTwice}, purge: []files.Object{obj(sumTwice), obj(sumTwice)}, wantGone: []string{sumTwice}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCatalog(t)
			paths := map[string]string{}
			for _, sum := range tt.objects {
				paths[sum] = writeObject(t, c, sum)
			}
			for _, sum := range tt.rows {
				insertRow(t, c, sum)
			}
			files.PurgeUnreferenced(c, tt.purge)
			for _, sum := range tt.wantGone {
				if exists(paths[sum]) {
					t.Fatalf("%s should be removed", sum[:4])
				}
			}
			for _, sum := range tt.wantKept {
				if !exists(paths[sum]) {
					t.Fatalf("%s should stay", sum[:4])
				}
			}
		})
	}

	t.Run("commit purges and rollback does not", func(t *testing.T) {
		c := newCatalog(t)
		path := writeObject(t, c, sumGone)
		boom := errors.New("boom")
		_, err := writes.Call(c, writes.Op{Action: "t"}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			files.PurgeAfterCommit(tx, []files.Object{obj(sumGone)})
			return struct{}{}, nil, boom
		})
		if !errors.Is(err, boom) {
			t.Fatalf("err %v", err)
		}
		if !exists(path) {
			t.Fatal("rolled-back write removed the object")
		}
	})

	t.Run("empty list on a Tx with no catalog", func(t *testing.T) {
		files.PurgeAfterCommit(&database.Tx{}, nil)
	})

	t.Run("non-empty list on a Tx with no catalog panics", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		files.PurgeAfterCommit(&database.Tx{}, []files.Object{obj(sumGone)})
	})
}
