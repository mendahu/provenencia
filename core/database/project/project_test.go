package project

import (
	"bytes"
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
			name: "upsert mints uuid then get",
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
				if len(got.UUID) != 16 {
					t.Fatalf("expected minted uuid, got %v", got.UUID)
				}
			},
		},
		{
			name: "upsert preserves existing uuid",
			run: func(t *testing.T, c *database.Catalog) {
				r, err := ref.Mint(ref.PrefixUser)
				if err != nil {
					t.Fatal(err)
				}
				if err := users.Upsert(c, id, "Jake", r); err != nil {
					t.Fatal(err)
				}
				fixed := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
				now := NowUTC()
				if err := Upsert(c, Info{Label: "A", CreatedAt: now, UpdatedAt: now, UpdatedBy: id, UUID: fixed}); err != nil {
					t.Fatal(err)
				}
				other := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
				if err := Upsert(c, Info{Label: "B", CreatedAt: now, UpdatedAt: now, UpdatedBy: id, UUID: other}); err != nil {
					t.Fatal(err)
				}
				got, err := Get(c)
				if err != nil {
					t.Fatal(err)
				}
				if got.Label != "B" {
					t.Fatalf("label=%q", got.Label)
				}
				if !bytes.Equal(got.UUID, fixed) {
					t.Fatalf("uuid rewritten: got %v want %v", got.UUID, fixed)
				}
			},
		},
		{
			name: "ensure uuid heals null once",
			run: func(t *testing.T, c *database.Catalog) {
				r, err := ref.Mint(ref.PrefixUser)
				if err != nil {
					t.Fatal(err)
				}
				if err := users.Upsert(c, id, "Jake", r); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				now := NowUTC()
				// Insert without uuid to simulate a pre-000017 row after ALTER.
				if _, err := db.Exec(
					`INSERT INTO project (id, label, created_at, updated_at, updated_by, uuid) VALUES (1, ?, ?, ?, ?, NULL)`,
					"Old", now, now, id,
				); err != nil {
					t.Fatal(err)
				}
				got, err := Get(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(got.UUID) != 0 {
					t.Fatalf("expected null uuid, got %v", got.UUID)
				}
				if err := EnsureUUID(c); err != nil {
					t.Fatal(err)
				}
				first, err := Get(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(first.UUID) != 16 {
					t.Fatalf("expected healed uuid, got %v", first.UUID)
				}
				if err := EnsureUUID(c); err != nil {
					t.Fatal(err)
				}
				again, err := Get(c)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(again.UUID, first.UUID) {
					t.Fatalf("EnsureUUID not idempotent: %v -> %v", first.UUID, again.UUID)
				}
			},
		},
		{
			name: "ensure uuid no-op when missing row",
			run: func(t *testing.T, c *database.Catalog) {
				if err := EnsureUUID(c); err != nil {
					t.Fatal(err)
				}
				_, err := Get(c)
				if !errors.Is(err, ErrMissing) {
					t.Fatalf("got %v", err)
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
				if err := Upsert(c, Info{Label: "  ", CreatedAt: NowUTC(), UpdatedAt: NowUTC(), UpdatedBy: id}); !errors.Is(err, ErrInvalid) {
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
