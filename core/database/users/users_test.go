package users

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/ref"
)

func TestUpsert(t *testing.T) {
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	mustRef := func(t *testing.T) string {
		t.Helper()
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "insert then update name",
			run: func(t *testing.T, c *database.Catalog) {
				r := mustRef(t)
				if err := Upsert(c, id, "Jake", r); err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.DisplayName != "Jake" || got.Ref != r {
					t.Fatalf("got %+v", got)
				}
				if err := Upsert(c, id, "Jake R.", r); err != nil {
					t.Fatal(err)
				}
				got, err = Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.DisplayName != "Jake R." {
					t.Fatalf("got %q", got.DisplayName)
				}
			},
		},
		{
			name: "rejects short id",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Upsert(c, []byte{1}, "Jake", mustRef(t)); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects blank name",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Upsert(c, id, "  ", mustRef(t)); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects bad ref",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Upsert(c, id, "Jake", "nope"); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects closed catalog",
			run: func(t *testing.T, c *database.Catalog) {
				c.Close()
				if err := Upsert(c, id, "Jake", mustRef(t)); !errors.Is(err, database.ErrClosed) {
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
				if err := Upsert(c, id2, "Ann", mustRef(t)); err != nil {
					t.Fatal(err)
				}
				if err := Upsert(c, id, "Jake", mustRef(t)); err != nil {
					t.Fatal(err)
				}
				got, err = List(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 2 || got[0].DisplayName != "Ann" || got[1].DisplayName != "Jake" {
					t.Fatalf("%+v", got)
				}
				if got[0].Ref == "" || got[1].Ref == "" {
					t.Fatalf("missing refs %+v", got)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := database.Create(t.TempDir(), "T.provenencia")
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			tt.run(t, c)
		})
	}
}

func TestEnsureRefs(t *testing.T) {
	idNull := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	idEmpty := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	c, err := database.Create(t.TempDir(), "T.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, display_name, ref) VALUES (?, ?, NULL)`, idNull, "Null Ref"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, display_name, ref) VALUES (?, ?, '')`, idEmpty, "Empty Ref"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureRefs(c); err != nil {
		t.Fatal(err)
	}
	gotNull, err := Lookup(c, idNull)
	if err != nil {
		t.Fatal(err)
	}
	gotEmpty, err := Lookup(c, idEmpty)
	if err != nil {
		t.Fatal(err)
	}
	if gotNull.Ref == "" || ref.Validate(gotNull.Ref) != nil {
		t.Fatalf("null backfill: %+v", gotNull)
	}
	if gotEmpty.Ref == "" || ref.Validate(gotEmpty.Ref) != nil {
		t.Fatalf("empty backfill: %+v", gotEmpty)
	}
	if gotNull.Ref == gotEmpty.Ref {
		t.Fatalf("expected distinct refs, both %q", gotNull.Ref)
	}
	first := gotNull.Ref
	if err := EnsureRefs(c); err != nil {
		t.Fatal(err)
	}
	again, err := Lookup(c, idNull)
	if err != nil {
		t.Fatal(err)
	}
	if again.Ref != first {
		t.Fatalf("EnsureRefs should be idempotent: %q -> %q", first, again.Ref)
	}
}
