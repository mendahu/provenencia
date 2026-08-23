package onboarding

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenance/core/database"
	"github.com/mendahu/provenance/core/database/users"
	"github.com/mendahu/provenance/core/identity"
	"github.com/mendahu/provenance/core/session"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, ident, parent string)
	}{
		{
			name: "opens existing and upserts users",
			run: func(t *testing.T, ident, parent string) {
				created, err := Complete(ident, parent, "Jake", "One")
				if err != nil {
					t.Fatal(err)
				}
				if err := session.Remove(ident); err != nil {
					t.Fatal(err)
				}
				res, err := Open(ident, created.ProjectDir, "")
				if err != nil {
					t.Fatal(err)
				}
				if res.Identity.UserID != created.Identity.UserID {
					t.Fatal("uuid changed")
				}
				act, err := session.Load(ident)
				if err != nil {
					t.Fatal(err)
				}
				if act.ProjectDir != created.ProjectDir {
					t.Fatalf("active %s", act.ProjectDir)
				}
				p, err := database.Open(created.ProjectDir)
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				name, err := users.Lookup(p, res.Identity.UserID[:])
				if err != nil {
					t.Fatal(err)
				}
				if name != "Jake" {
					t.Fatalf("users %q", name)
				}
			},
		},
		{
			name: "mints when no identity and name given",
			run: func(t *testing.T, ident, parent string) {
				other := t.TempDir()
				created, err := Complete(other, parent, "Other", "Shared")
				if err != nil {
					t.Fatal(err)
				}
				res, err := Open(ident, created.ProjectDir, "Jake")
				if err != nil {
					t.Fatal(err)
				}
				if res.Identity.DisplayName != "Jake" {
					t.Fatalf("name %q", res.Identity.DisplayName)
				}
				loaded, err := identity.Load(ident)
				if err != nil {
					t.Fatal(err)
				}
				if loaded.UserID != res.Identity.UserID {
					t.Fatal("identity mismatch")
				}
			},
		},
		{
			name: "no identity blank name",
			run: func(t *testing.T, ident, parent string) {
				_, err := Open(ident, parent, "")
				if !errors.Is(err, ErrBlankName) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "not a project",
			run: func(t *testing.T, ident, parent string) {
				id, err := identity.Mint("Jake")
				if err != nil {
					t.Fatal(err)
				}
				if err := identity.Save(ident, id); err != nil {
					t.Fatal(err)
				}
				_, err = Open(ident, parent, "")
				if !errors.Is(err, database.ErrNotAProject) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "corrupt identity json",
			run: func(t *testing.T, ident, parent string) {
				created, err := Complete(t.TempDir(), parent, "Jake", "One")
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(ident, identity.FileName), []byte("{"), 0o600); err != nil {
					t.Fatal(err)
				}
				_, err = Open(ident, created.ProjectDir, "Jake")
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			name: "blank project dir",
			run: func(t *testing.T, ident, parent string) {
				_, err := Open(ident, "  ", "Jake")
				if !errors.Is(err, database.ErrNotAProject) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, t.TempDir(), t.TempDir())
		})
	}
}
