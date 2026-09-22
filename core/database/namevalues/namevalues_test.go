package namevalues

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
)

func TestInsertLookup(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "migration creates name tables",
			run: func(t *testing.T, c *database.Catalog) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				for _, table := range []string{"name_values", "name_value_parts"} {
					var name string
					if err := db.QueryRow(
						`SELECT name FROM sqlite_schema WHERE type='table' AND name=?`, table,
					).Scan(&name); err != nil || name != table {
						t.Fatalf("table %q %v", name, err)
					}
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 25 {
					t.Fatalf("user_version=%d", ver)
				}
			},
		},
		{
			name: "form only round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{Form: "蒋浩"})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Form != "蒋浩" || len(got.Parts) != 0 {
					t.Fatalf("got %+v", got)
				}
				if len(got.ID) != 16 {
					t.Fatalf("id len %d", len(got.ID))
				}
			},
		},
		{
			name: "form with typed parts round trip",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Form: "James K. Robins",
					Parts: []Part{
						{Idx: 0, Value: "James", Type: PartTypeGiven},
						{Idx: 1, Value: "K.", Type: PartTypeInitial},
						{Idx: 2, Value: "Robins", Type: PartTypeSurname},
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Form != "James K. Robins" || len(got.Parts) != 3 {
					t.Fatalf("got %+v", got)
				}
				if got.Parts[0].Value != "James" || got.Parts[0].Type != PartTypeGiven ||
					got.Parts[1].Value != "K." || got.Parts[1].Type != PartTypeInitial ||
					got.Parts[2].Value != "Robins" || got.Parts[2].Type != PartTypeSurname {
					t.Fatalf("parts %+v", got.Parts)
				}
				for _, p := range got.Parts {
					if len(p.ID) != 16 {
						t.Fatalf("part id len %d", len(p.ID))
					}
				}
			},
		},
		{
			name: "untyped part and gapped idx",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Form: "Maria da Silva Costa",
					Parts: []Part{
						{Idx: 0, Value: "Maria"},
						{Idx: 2, Value: "Costa", Type: PartTypeSurname},
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if len(got.Parts) != 2 ||
					got.Parts[0].Idx != 0 || got.Parts[0].Type != "" ||
					got.Parts[1].Idx != 2 || got.Parts[1].Type != PartTypeSurname {
					t.Fatalf("parts %+v", got.Parts)
				}
			},
		},
		{
			name: "trims form and part fields",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Insert(c, Value{
					Form: "  James Robins  ",
					Parts: []Part{
						{Idx: 0, Value: "  James  ", Type: "  given  "},
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, id)
				if err != nil {
					t.Fatal(err)
				}
				if got.Form != "James Robins" ||
					got.Parts[0].Value != "James" || got.Parts[0].Type != "given" {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "rejects unknown part type",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Form:  "James",
					Parts: []Part{{Idx: 0, Value: "James", Type: "first_name"}},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects blank form",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{Form: "   "})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects blank part value",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Form:  "James",
					Parts: []Part{{Idx: 0, Value: "  "}},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects negative idx",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Form:  "James",
					Parts: []Part{{Idx: -1, Value: "James"}},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects duplicate idx",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Insert(c, Value{
					Form: "James Robins",
					Parts: []Part{
						{Idx: 0, Value: "James"},
						{Idx: 0, Value: "Robins"},
					},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "lookup rejects short id",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Lookup(c, []byte{1})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "lookup missing row",
			run: func(t *testing.T, c *database.Catalog) {
				id := make([]byte, 16)
				_, err := Lookup(c, id)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects closed catalog",
			run: func(t *testing.T, c *database.Catalog) {
				c.Close()
				_, err := Insert(c, Value{Form: "James"})
				if !errors.Is(err, database.ErrClosed) {
					t.Fatalf("got %v", err)
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
