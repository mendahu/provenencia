package objectstore_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
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

func TestWriteIfAbsent(t *testing.T) {
	data := []byte("object bytes")
	sum := sha256.Sum256(data)
	checksum := hex.EncodeToString(sum[:])
	const rel = "objects/ab/cd/obj.png"

	noTemps := func(t *testing.T, project string) {
		t.Helper()
		entries, err := os.ReadDir(filepath.Join(project, "objects"))
		if err != nil {
			return
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".write-") {
				t.Fatalf("temp %s left behind", e.Name())
			}
		}
	}

	t.Run("write, repeat, and replace", func(t *testing.T) {
		project := t.TempDir()
		path := filepath.Join(project, filepath.FromSlash(rel))

		wrote, err := objectstore.WriteIfAbsent(project, rel, checksum, data)
		if err != nil || !wrote {
			t.Fatalf("absent: wrote=%v err=%v", wrote, err)
		}
		if got, _ := os.ReadFile(path); string(got) != string(data) {
			t.Fatalf("content = %q", got)
		}

		wrote, err = objectstore.WriteIfAbsent(project, rel, checksum, data)
		if err != nil || wrote {
			t.Fatalf("same bytes again: wrote=%v err=%v", wrote, err)
		}

		if err := os.WriteFile(path, []byte("other"), 0o644); err != nil {
			t.Fatal(err)
		}
		wrote, err = objectstore.WriteIfAbsent(project, rel, checksum, data)
		if err != nil || !wrote {
			t.Fatalf("different file: wrote=%v err=%v", wrote, err)
		}
		if got, _ := os.ReadFile(path); string(got) != string(data) {
			t.Fatalf("content after replace = %q", got)
		}
		noTemps(t, project)
	})

	t.Run("mismatched checksum writes nothing", func(t *testing.T) {
		project := t.TempDir()
		if wrote, err := objectstore.WriteIfAbsent(project, "objects/x.png", strings.Repeat("0", 64), data); err == nil || wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		if _, err := os.Lstat(filepath.Join(project, "objects")); err == nil {
			t.Fatal("mismatched checksum created objects/")
		}
	})

	t.Run("creates objects when absent", func(t *testing.T) {
		project := t.TempDir()
		wrote, err := objectstore.WriteIfAbsent(project, rel, checksum, data)
		if err != nil || !wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		noTemps(t, project)
	})

	t.Run("objects subfolder symlinked outside", func(t *testing.T) {
		project, out := t.TempDir(), t.TempDir()
		if err := os.MkdirAll(filepath.Join(project, "objects"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(out, filepath.Join(project, "objects", "ab")); err != nil {
			t.Fatal(err)
		}
		if wrote, err := objectstore.WriteIfAbsent(project, rel, checksum, data); err == nil || wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		if entries, _ := os.ReadDir(out); len(entries) != 0 {
			t.Fatalf("outside folder holds %d entries", len(entries))
		}
		noTemps(t, project)
	})

	t.Run("objects folder symlinked outside", func(t *testing.T) {
		project, out := t.TempDir(), t.TempDir()
		if err := os.Symlink(out, filepath.Join(project, "objects")); err != nil {
			t.Fatal(err)
		}
		if wrote, err := objectstore.WriteIfAbsent(project, rel, checksum, data); !errors.Is(err, objectstore.ErrUnsafeObjects) || wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		if entries, _ := os.ReadDir(out); len(entries) != 0 {
			t.Fatalf("outside folder holds %d entries", len(entries))
		}
	})

	t.Run("object path symlinked to an outside file", func(t *testing.T) {
		project, out := t.TempDir(), t.TempDir()
		outside := filepath.Join(out, "secret.png")
		if err := os.WriteFile(outside, data, 0o644); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(project, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, link); err != nil {
			t.Fatal(err)
		}
		wrote, err := objectstore.WriteIfAbsent(project, rel, checksum, data)
		if err != nil || !wrote {
			t.Fatalf("wrote=%v err=%v", wrote, err)
		}
		st, err := os.Lstat(link)
		if err != nil || !st.Mode().IsRegular() {
			t.Fatalf("object path is not a regular file: %v %v", st, err)
		}
		if got, _ := os.ReadFile(link); string(got) != string(data) {
			t.Fatalf("content = %q", got)
		}
		if got, _ := os.ReadFile(outside); string(got) != string(data) {
			t.Fatalf("outside file changed: %q", got)
		}
		noTemps(t, project)
	})
}

func TestOpenObject(t *testing.T) {
	const rel = "objects/ab/cd/abcd.png"

	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns the project folder
		wantErr error
	}{
		{
			name: "regular file opens",
			setup: func(t *testing.T) string {
				p := t.TempDir()
				path := filepath.Join(p, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("bytes"), 0o644); err != nil {
					t.Fatal(err)
				}
				return p
			},
		},
		{
			name: "symlink to an outside file",
			setup: func(t *testing.T) string {
				p, out := t.TempDir(), t.TempDir()
				outside := filepath.Join(out, "secret.png")
				if err := os.WriteFile(outside, []byte("bytes"), 0o644); err != nil {
					t.Fatal(err)
				}
				link := filepath.Join(p, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, link); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: objectstore.ErrNotRegular,
		},
		{
			name: "missing object",
			setup: func(t *testing.T) string {
				p := t.TempDir()
				if err := os.MkdirAll(filepath.Join(p, "objects"), 0o755); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: fs.ErrNotExist,
		},
		{
			name: "object path is a folder",
			setup: func(t *testing.T) string {
				p := t.TempDir()
				if err := os.MkdirAll(filepath.Join(p, filepath.FromSlash(rel)), 0o755); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: objectstore.ErrNotRegular,
		},
		{
			name: "objects folder symlinked outside",
			setup: func(t *testing.T) string {
				p, out := t.TempDir(), t.TempDir()
				if err := os.Symlink(out, filepath.Join(p, "objects")); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: objectstore.ErrUnsafeObjects,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := objectstore.OpenObject(tt.setup(t), rel)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			got, err := io.ReadAll(f)
			if err != nil || string(got) != "bytes" {
				t.Fatalf("read %q, %v", got, err)
			}
		})
	}
}
