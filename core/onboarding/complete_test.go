package onboarding

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/identity"
	"github.com/mendahu/provenencia/core/installstate"
)

func TestFolderName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{name: "plain", in: "Robins Family", want: "Robins Family.provenencia"},
		{name: "slash", in: "Robins/Family", want: "Robins-Family.provenencia"},
		{name: "backslash", in: "Robins\\Family", want: "Robins-Family.provenencia"},
		{name: "nul", in: "Robins\x00Family", want: "Robins-Family.provenencia"},
		{name: "colon", in: "Robins:Family", want: "Robins-Family.provenencia"},
		{name: "star", in: "Robins*Family", want: "Robins-Family.provenencia"},
		{name: "question", in: "Robins?Family", want: "Robins-Family.provenencia"},
		{name: "quote", in: "Robins\"Family", want: "Robins-Family.provenencia"},
		{name: "lt gt", in: "Robins<Family>", want: "Robins-Family-.provenencia"},
		{name: "pipe", in: "Robins|Family", want: "Robins-Family.provenencia"},
		{name: "leading dots", in: "...Robins", want: "Robins.provenencia"},
		{name: "already suffix", in: "Robins Family.provenencia", want: "Robins Family.provenencia"},
		{name: "only dots", in: "...", wantErr: ErrInvalidFamilyName},
		{name: "blank", in: "  ", wantErr: ErrInvalidFamilyName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FolderName(tt.in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, ident, parent string)
	}{
		{
			name: "first run writes identity project and users",
			run: func(t *testing.T, ident, parent string) {
				res, err := Complete(ident, parent, "Jake Robins", "Robins Family")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(filepath.Join(ident, identity.FileName)); err != nil {
					t.Fatal(err)
				}
				loaded, err := identity.Load(ident)
				if err != nil {
					t.Fatal(err)
				}
				if loaded.DisplayName != "Jake Robins" || loaded.UserID != res.Identity.UserID {
					t.Fatalf("identity mismatch")
				}
				p, err := database.Open(res.ProjectDir)
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				name, err := users.Lookup(p, res.Identity.UserID[:])
				if err != nil {
					t.Fatal(err)
				}
				if name != "Jake Robins" {
					t.Fatalf("users %q", name)
				}
				if filepath.Base(res.ProjectDir) != "Robins Family"+database.Suffix {
					t.Fatalf("dir %s", res.ProjectDir)
				}
				raw, err := os.ReadFile(filepath.Join(ident, identity.FileName))
				if err != nil {
					t.Fatal(err)
				}
				var m map[string]string
				if err := json.Unmarshal(raw, &m); err != nil {
					t.Fatal(err)
				}
				if _, ok := m["family_name"]; ok {
					t.Fatal("identity.json must not store family name")
				}
				if _, ok := m["project"]; ok {
					t.Fatal("identity.json must not store project")
				}
				if m["user_id"] == "" || m["display_name"] != "Jake Robins" {
					t.Fatalf("json %v", m)
				}
				if len(m) != 2 {
					t.Fatalf("unexpected identity keys %v", m)
				}
				act, err := installstate.Load(ident)
				if err != nil {
					t.Fatal(err)
				}
				if act.ProjectDir != res.ProjectDir {
					t.Fatalf("active %s want %s", act.ProjectDir, res.ProjectDir)
				}
			},
		},
		{
			name: "second project same identity uuid",
			run: func(t *testing.T, ident, parent string) {
				a, err := Complete(ident, parent, "Jake", "One")
				if err != nil {
					t.Fatal(err)
				}
				b, err := Complete(ident, parent, "Jake R.", "Two")
				if err != nil {
					t.Fatal(err)
				}
				if a.Identity.UserID != b.Identity.UserID {
					t.Fatal("uuid changed")
				}
				loaded, err := identity.Load(ident)
				if err != nil {
					t.Fatal(err)
				}
				if loaded.DisplayName != "Jake R." {
					t.Fatalf("cache name %q", loaded.DisplayName)
				}
				p, err := database.Open(b.ProjectDir)
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				name, err := users.Lookup(p, b.Identity.UserID[:])
				if err != nil {
					t.Fatal(err)
				}
				if name != "Jake R." {
					t.Fatalf("users %q", name)
				}
				act, err := installstate.Load(ident)
				if err != nil {
					t.Fatal(err)
				}
				if act.ProjectDir != b.ProjectDir {
					t.Fatalf("active %s", act.ProjectDir)
				}
			},
		},
		{
			name: "already exists keeps identity",
			run: func(t *testing.T, ident, parent string) {
				_, err := Complete(ident, parent, "Jake", "Dup")
				if err != nil {
					t.Fatal(err)
				}
				_, err = Complete(ident, parent, "Jake", "Dup")
				if !errors.Is(err, database.ErrAlreadyExists) {
					t.Fatalf("got %v", err)
				}
				if _, err := identity.Load(ident); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "sanitize slash in family name",
			run: func(t *testing.T, ident, parent string) {
				res, err := Complete(ident, parent, "Jake", "A/B")
				if err != nil {
					t.Fatal(err)
				}
				if filepath.Base(res.ProjectDir) != "A-B.provenencia" {
					t.Fatalf("got %s", res.ProjectDir)
				}
			},
		},
		{
			name: "dots-only family does not write identity",
			run: func(t *testing.T, ident, parent string) {
				_, err := Complete(ident, parent, "Jake", "...")
				if !errors.Is(err, ErrInvalidFamilyName) {
					t.Fatalf("got %v", err)
				}
				ents, err := os.ReadDir(ident)
				if err != nil {
					t.Fatal(err)
				}
				if len(ents) != 0 {
					t.Fatalf("identity dir touched: %v", ents)
				}
			},
		},
		{
			name: "corrupt identity json",
			run: func(t *testing.T, ident, parent string) {
				if err := os.WriteFile(filepath.Join(ident, identity.FileName), []byte("{"), 0o600); err != nil {
					t.Fatal(err)
				}
				_, err := Complete(ident, parent, "Jake", "Family")
				if err == nil {
					t.Fatal("expected error")
				}
				ents, err := os.ReadDir(parent)
				if err != nil {
					t.Fatal(err)
				}
				if len(ents) != 0 {
					t.Fatalf("parent touched: %v", ents)
				}
			},
		},
		{
			name: "blank names do not write",
			run: func(t *testing.T, ident, parent string) {
				_, err := Complete(ident, parent, "  ", "Family")
				if !errors.Is(err, ErrBlankName) {
					t.Fatalf("got %v", err)
				}
				_, err = Complete(ident, parent, "Jake", "  ")
				if !errors.Is(err, ErrBlankName) {
					t.Fatalf("got %v", err)
				}
				ents, err := os.ReadDir(ident)
				if err != nil {
					t.Fatal(err)
				}
				if len(ents) != 0 {
					t.Fatalf("identity dir touched: %v", ents)
				}
				ents, err = os.ReadDir(parent)
				if err != nil {
					t.Fatal(err)
				}
				if len(ents) != 0 {
					t.Fatalf("parent touched: %v", ents)
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
