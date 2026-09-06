package identity

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestActive(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, dir string)
	}{
		{
			name: "round trip json",
			run: func(t *testing.T, dir string) {
				want := Active{ProjectDir: "/tmp/Smith Family.provenencia"}
				if err := SaveActive(dir, want); err != nil {
					t.Fatal(err)
				}
				got, err := LoadActive(dir)
				if err != nil {
					t.Fatal(err)
				}
				if got.ProjectDir != want.ProjectDir {
					t.Fatalf("got %+v want %+v", got, want)
				}
				raw, err := os.ReadFile(filepath.Join(dir, ActiveFileName))
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
			name: "load missing is ErrActiveNotFound",
			run: func(t *testing.T, dir string) {
				_, err := LoadActive(dir)
				if !errors.Is(err, ErrActiveNotFound) {
					t.Fatalf("got %v want %v", err, ErrActiveNotFound)
				}
			},
		},
		{
			name: "save rejects blank dir",
			run: func(t *testing.T, dir string) {
				if err := SaveActive(dir, Active{ProjectDir: "  \t"}); !errors.Is(err, ErrActiveInvalid) {
					t.Fatalf("got %v want %v", err, ErrActiveInvalid)
				}
			},
		},
		{
			name: "remove missing is ok",
			run: func(t *testing.T, dir string) {
				if err := RemoveActive(dir); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "remove then load is not found",
			run: func(t *testing.T, dir string) {
				if err := SaveActive(dir, Active{ProjectDir: "/tmp/A.provenencia"}); err != nil {
					t.Fatal(err)
				}
				if err := RemoveActive(dir); err != nil {
					t.Fatal(err)
				}
				if _, err := LoadActive(dir); !errors.Is(err, ErrActiveNotFound) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "corrupt json",
			run: func(t *testing.T, dir string) {
				if err := os.WriteFile(filepath.Join(dir, ActiveFileName), []byte("{"), 0o600); err != nil {
					t.Fatal(err)
				}
				if _, err := LoadActive(dir); err == nil {
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
