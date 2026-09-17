package subjecttypes

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/ref"
)

func TestSubjectTypes(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "install seeds seven provenencia types",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				list, err := List(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(list) != 7 {
					t.Fatalf("len=%d", len(list))
				}
				person, err := Lookup(c, "person", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if person.RefPrefix != "PER" || person.CandidateRefPrefix != "CPR" {
					t.Fatalf("prefixes %+v", person)
				}
				if person.Label != "Person" {
					t.Fatalf("label %q", person.Label)
				}
			},
		},
		{
			name: "install twice keeps same ids",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				first, err := Lookup(c, "person", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				second, err := Lookup(c, "person", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(first.ID) != string(second.ID) {
					t.Fatal("id changed on reinstall")
				}
				list, err := List(c)
				if err != nil || len(list) != 7 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "reject empty key label prefixes",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Type{Key: "", Origin: OriginUser, Label: "X", RefPrefix: "AAA", CandidateRefPrefix: "BBB"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty key %v", err)
				}
				if _, err := Upsert(c, Type{Key: "x", Origin: OriginUser, Label: "", RefPrefix: "AAA", CandidateRefPrefix: "BBB"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty label %v", err)
				}
			},
		},
		{
			name: "reject reserved prefix",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Upsert(c, Type{
					Key: "bad", Origin: OriginUser, Label: "Bad",
					RefPrefix: "SRC", CandidateRefPrefix: "BBB",
				})
				if !errors.Is(err, ref.ErrReservedPrefix) {
					t.Fatalf("got %v", err)
				}
				if apperr.From(err).Code() != apperr.CodeRefReservedPrefix {
					t.Fatalf("code %v", err)
				}
			},
		},
		{
			name: "reject same prefix on both columns",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Upsert(c, Type{
					Key: "dup", Origin: OriginUser, Label: "Dup",
					RefPrefix: "ZZZ", CandidateRefPrefix: "zzz",
				})
				if !errors.Is(err, ErrDuplicatePrefix) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "reject cross-column collision with seeded candidate",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				_, err := Upsert(c, Type{
					Key: "collide", Origin: OriginUser, Label: "Collide",
					RefPrefix: "CPR", CandidateRefPrefix: "QQQ",
				})
				if !errors.Is(err, ErrDuplicatePrefix) {
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
