package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSession(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, dir string)
	}{
		{
			name: "round trip json",
			run: func(t *testing.T, dir string) {
				want := Active{ProjectDir: "/tmp/Smith Family.provenance"}
				if err := Save(dir, want); err != nil {
					t.Fatal(err)
				}
				got, err := Load(dir)
				if err != nil {
					t.Fatal(err)
				}
				if got.ProjectDir != want.ProjectDir {
					t.Fatalf("got %+v want %+v", got, want)
				}
				raw, err := os.ReadFile(filepath.Join(dir, FileName))
				if err != nil {
					t.Fatal(err)
				}
				var m map[string]string
				if err := json.Unmarshal(raw, &m); err != nil {
					t.Fatal(err)
				}
				if m["project_dir"] != want.ProjectDir {
					t.Fatalf("json %v", m)
				}
			},
		},
		{
			name: "load missing is ErrNotFound",
			run: func(t *testing.T, dir string) {
				_, err := Load(dir)
				if err != ErrNotFound {
					t.Fatalf("got %v want %v", err, ErrNotFound)
				}
			},
		},
		{
			name: "save rejects blank dir",
			run: func(t *testing.T, dir string) {
				if err := Save(dir, Active{ProjectDir: "  \t"}); err != ErrInvalid {
					t.Fatalf("got %v want %v", err, ErrInvalid)
				}
			},
		},
		{
			name: "remove missing is ok",
			run: func(t *testing.T, dir string) {
				if err := Remove(dir); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "remove then load is not found",
			run: func(t *testing.T, dir string) {
				if err := Save(dir, Active{ProjectDir: "/tmp/A.provenance"}); err != nil {
					t.Fatal(err)
				}
				if err := Remove(dir); err != nil {
					t.Fatal(err)
				}
				if _, err := Load(dir); err != ErrNotFound {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "corrupt json",
			run: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, FileName), []byte("{"), 0o600); err != nil {
					t.Fatal(err)
				}
				if _, err := Load(dir); err == nil {
					t.Fatal("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, t.TempDir())
		})
	}
}
