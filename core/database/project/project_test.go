package project

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestProject(t *testing.T) {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "upsert then get",
			run: func(t *testing.T, c *database.Catalog) {
				r, err := ref.Mint(ref.PrefixUser)
				if err != nil {
					t.Fatal(err)
				}
				if err := users.Upsert(c, id, "Jake", r); err != nil {
					t.Fatal(err)
				}
				now := NowUTC()
				info := Info{Label: "Robins Family", CreatedAt: now, UpdatedAt: now, UpdatedBy: id}
				if err := Upsert(c, info); err != nil {
					t.Fatal(err)
				}
				got, err := Get(c)
				if err != nil {
					t.Fatal(err)
				}
				if got.Label != "Robins Family" || got.CreatedAt != now || string(got.UpdatedBy) != string(id) {
					t.Fatalf("%+v", got)
				}
			},
		},
		{
			name: "get missing",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Get(c)
				if !errors.Is(err, ErrMissing) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects blank label",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Upsert(c, Info{Label: "  ", CreatedAt: NowUTC(), UpdatedAt: NowUTC(), UpdatedBy: id}); err != ErrInvalid {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "label from dir",
			run: func(t *testing.T, c *database.Catalog) {
				dir := filepath.Join(t.TempDir(), "Robins Family.provenencia")
				if got := LabelFromDir(dir); got != "Robins Family" {
					t.Fatalf("got %q", got)
				}
				if got := LabelFromDir(filepath.Join(t.TempDir(), "robins-family.provenencia")); got != "robins-family" {
					t.Fatalf("got %q", got)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "t.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}
