package objectstore_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/objectstore"
)

func TestRemoveObject(t *testing.T) {
	const rel = "objects/ab/cd/" + "abcd.jpg"

	write := func(t *testing.T, path string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("bytes"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	exists := func(path string) bool {
		_, err := os.Lstat(path)
		return err == nil
	}

	tests := []struct {
		name string
		rel  string
		// setup returns the project folder and a path that must still exist.
		setup   func(t *testing.T) (project, mustStay string)
		wantErr bool
		gone    bool // the object at rel must be gone
	}{
		{
			name: "regular object is removed",
			rel:  rel,
			setup: func(t *testing.T) (string, string) {
				p := t.TempDir()
				write(t, filepath.Join(p, filepath.FromSlash(rel)))
				return p, ""
			},
			gone: true,
		},
		{
			name: "missing object is not an error",
			rel:  rel,
			setup: func(t *testing.T) (string, string) {
				p := t.TempDir()
				if err := os.MkdirAll(filepath.Join(p, "objects"), 0o755); err != nil {
					t.Fatal(err)
				}
				return p, ""
			},
		},
		{
			name: "objects subfolder symlinked outside",
			rel:  rel,
			setup: func(t *testing.T) (string, string) {
				p, out := t.TempDir(), t.TempDir()
				outside := filepath.Join(out, "cd", "abcd.jpg")
				write(t, outside)
				if err := os.MkdirAll(filepath.Join(p, "objects"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(out, filepath.Join(p, "objects", "ab")); err != nil {
					t.Fatal(err)
				}
				return p, outside
			},
			wantErr: true,
		},
		{
			name: "objects folder symlinked outside",
			rel:  rel,
			setup: func(t *testing.T) (string, string) {
				p, out := t.TempDir(), t.TempDir()
				outside := filepath.Join(out, "ab", "cd", "abcd.jpg")
				write(t, outside)
				if err := os.Symlink(out, filepath.Join(p, "objects")); err != nil {
					t.Fatal(err)
				}
				return p, outside
			},
			wantErr: true,
		},
		{
			name: "object path is a symlink to an outside file",
			rel:  rel,
			setup: func(t *testing.T) (string, string) {
				p, out := t.TempDir(), t.TempDir()
				outside := filepath.Join(out, "secret.jpg")
				write(t, outside)
				link := filepath.Join(p, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, link); err != nil {
					t.Fatal(err)
				}
				return p, outside
			},
		},
		{
			name: "rel outside objects",
			rel:  "notes/abcd.jpg",
			setup: func(t *testing.T) (string, string) {
				p := t.TempDir()
				keep := filepath.Join(p, "notes", "abcd.jpg")
				write(t, keep)
				return p, keep
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, mustStay := tt.setup(t)
			err := objectstore.RemoveObject(project, tt.rel)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err %v, wantErr %v", err, tt.wantErr)
			}
			if mustStay != "" && !exists(mustStay) {
				t.Fatalf("%s was removed", strings.TrimPrefix(mustStay, project))
			}
			if tt.gone && exists(filepath.Join(project, filepath.FromSlash(tt.rel))) {
				t.Fatal("object still exists")
			}
		})
	}
}
