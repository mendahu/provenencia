package claimconfidencegrades

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
)

func TestGrades(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "install seeds three provenencia grades",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				list, err := List(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(list) != 3 {
					t.Fatalf("len=%d", len(list))
				}
				if list[0].Key != "low_confidence" || list[1].Key != "moderate" || list[2].Key != "high_confidence" {
					t.Fatalf("order %+v", list)
				}
				std, err := Lookup(c, "moderate", OriginProvenencia)
				if err != nil || std.Label != "Moderate" || std.SortOrder != 2 {
					t.Fatalf("%v %+v", err, std)
				}
			},
		},
		{
			name: "install is idempotent",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				list, err := List(c)
				if err != nil || len(list) != 3 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "reject empty key label",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Grade{Key: "", Origin: OriginUser, Label: "X"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				if _, err := Upsert(c, Grade{Key: "x", Origin: OriginUser, Label: ""}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "get missing returns ErrNoRows",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := GetByID(c, make([]byte, 16))
				if !errors.Is(err, sql.ErrNoRows) {
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
