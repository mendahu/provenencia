package onboarding

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/identity"
	"github.com/mendahu/provenencia/core/ref"
)

func TestFolderName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{name: "plain", in: "Robins Family", want: "robins-family.provenencia"},
		{name: "slash", in: "Robins/Family", want: "robins-family.provenencia"},
		{name: "backslash", in: "Robins\\Family", want: "robins-family.provenencia"},
		{name: "nul", in: "Robins\x00Family", want: "robins-family.provenencia"},
		{name: "colon", in: "Robins:Family", want: "robins-family.provenencia"},
		{name: "star", in: "Robins*Family", want: "robins-family.provenencia"},
		{name: "question", in: "Robins?Family", want: "robins-family.provenencia"},
		{name: "quote", in: "Robins\"Family", want: "robins-family.provenencia"},
		{name: "lt gt", in: "Robins<Family>", want: "robins-family.provenencia"},
		{name: "pipe", in: "Robins|Family", want: "robins-family.provenencia"},
		{name: "leading dots", in: "...Robins", want: "robins.provenencia"},
		{name: "already suffix", in: "Robins Family.provenencia", want: "robins-family.provenencia"},
		{name: "collapse spaces", in: "  Robins   Family  ", want: "robins-family.provenencia"},
		{name: "unicode", in: "García Family", want: "garcía-family.provenencia"},
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
				if !ref.Valid(loaded.Ref) || loaded.Ref != res.Identity.Ref {
					t.Fatalf("ref %q", loaded.Ref)
				}
				p, err := database.Open(res.ProjectDir)
				if err != nil {
					t.Fatal(err)
				}
				defer p.Close()
				u, err := users.Lookup(p, res.Identity.UserID[:])
				if err != nil {
					t.Fatal(err)
				}
				if u.DisplayName != "Jake Robins" || u.Ref != loaded.Ref {
					t.Fatalf("users %+v", u)
				}
				info, err := project.Get(p)
				if err != nil {
					t.Fatal(err)
				}
				if info.Label != "Robins Family" {
					t.Fatalf("label %q", info.Label)
				}
				if filepath.Base(res.ProjectDir) != "robins-family"+database.Suffix {
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
				if m["user_id"] == "" || m["display_name"] != "Jake Robins" || !ref.Valid(m["ref"]) {
					t.Fatalf("json %v", m)
				}
				if len(m) != 3 {
					t.Fatalf("unexpected identity keys %v", m)
				}
				act, err := identity.LoadActive(ident)
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
				if a.Identity.Ref != b.Identity.Ref {
					t.Fatal("ref changed")
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
				u, err := users.Lookup(p, b.Identity.UserID[:])
				if err != nil {
					t.Fatal(err)
				}
				if u.DisplayName != "Jake R." {
					t.Fatalf("users %q", u.DisplayName)
				}
				act, err := identity.LoadActive(ident)
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
			name: "slug slash in family name",
			run: func(t *testing.T, ident, parent string) {
				res, err := Complete(ident, parent, "Jake", "A/B")
				if err != nil {
					t.Fatal(err)
				}
				if filepath.Base(res.ProjectDir) != "a-b.provenencia" {
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
