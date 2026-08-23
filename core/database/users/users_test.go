package users

import (
	"testing"

	"github.com/mendahu/provenance/core/database"
)

func TestUpsert(t *testing.T) {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "insert then update name",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Upsert(c, id, "Jake"); err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got != "Jake" {
					t.Fatalf("got %q", got)
				}
				if err := Upsert(c, id, "Jake R."); err != nil {
					t.Fatal(err)
				}
				got, err = Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got != "Jake R." {
					t.Fatalf("got %q", got)
				}
			},
		},
		{
			name: "rejects short id",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Upsert(c, []byte{1}, "Jake"); err != ErrInvalid {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects blank name",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Upsert(c, id, "  "); err != ErrInvalid {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects closed catalog",
			run: func(t *testing.T, c *database.Catalog) {
				c.Close()
				if err := Upsert(c, id, "Jake"); err != database.ErrClosed {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "list empty then two",
			run: func(t *testing.T, c *database.Catalog) {
				got, err := List(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 0 {
					t.Fatalf("got %d", len(got))
				}
				id2 := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
				if err := Upsert(c, id2, "Ann"); err != nil {
					t.Fatal(err)
				}
				if err := Upsert(c, id, "Jake"); err != nil {
					t.Fatal(err)
				}
				got, err = List(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 2 || got[0].DisplayName != "Ann" || got[1].DisplayName != "Jake" {
					t.Fatalf("%+v", got)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "T.provenance")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}
