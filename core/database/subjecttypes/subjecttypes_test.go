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
			name: "upsert list lookup",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{
					Key: "person", Origin: OriginProvenencia, Label: "Person",
					RefPrefix: "PER", CandidateRefPrefix: "CPR",
				})
				if err != nil {
					t.Fatal(err)
				}
				person, err := Lookup(c, "person", OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(person.ID) != string(id) || person.RefPrefix != "PER" {
					t.Fatalf("%+v", person)
				}
				list, err := List(c)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "upsert twice keeps same id",
			run: func(t *testing.T, c *database.Catalog) {
				first, err := Upsert(c, Type{
					Key: "person", Origin: OriginProvenencia, Label: "Person",
					RefPrefix: "PER", CandidateRefPrefix: "CPR",
				})
				if err != nil {
					t.Fatal(err)
				}
				second, err := Upsert(c, Type{
					Key: "person", Origin: OriginProvenencia, Label: "Person",
					RefPrefix: "PER", CandidateRefPrefix: "CPR",
				})
				if err != nil {
					t.Fatal(err)
				}
				if string(first) != string(second) {
					t.Fatal("id changed")
				}
			},
		},
		{
			name: "reject empty key label",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Type{Key: "", Origin: OriginProvenencia, Label: "X", RefPrefix: "AAA", CandidateRefPrefix: "BBB"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty key %v", err)
				}
				if _, err := Upsert(c, Type{Key: "x", Origin: OriginProvenencia, Label: "", RefPrefix: "AAA", CandidateRefPrefix: "BBB"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty label %v", err)
				}
			},
		},
		{
			name: "reject user origin",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Upsert(c, Type{
					Key: "custom", Origin: "user", Label: "Custom",
					RefPrefix: "AAA", CandidateRefPrefix: "BBB",
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "accept plugin origin",
			run: func(t *testing.T, c *database.Catalog) {
				id, err := Upsert(c, Type{
					Key: "dna_match", Origin: "plugin:example", Label: "DNA match",
					RefPrefix: "DNA", CandidateRefPrefix: "CDM",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := Lookup(c, "dna_match", "plugin:example")
				if err != nil || string(got.ID) != string(id) {
					t.Fatalf("%+v %v", got, err)
				}
			},
		},
		{
			name: "reject reserved prefix",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Upsert(c, Type{
					Key: "bad", Origin: OriginProvenencia, Label: "Bad",
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
					Key: "dup", Origin: OriginProvenencia, Label: "Dup",
					RefPrefix: "ZZZ", CandidateRefPrefix: "zzz",
				})
				if !errors.Is(err, ErrDuplicatePrefix) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "reject cross-column collision with existing candidate",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := Upsert(c, Type{
					Key: "person", Origin: OriginProvenencia, Label: "Person",
					RefPrefix: "PER", CandidateRefPrefix: "CPR",
				}); err != nil {
					t.Fatal(err)
				}
				_, err := Upsert(c, Type{
					Key: "collide", Origin: OriginProvenencia, Label: "Collide",
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
