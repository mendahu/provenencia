package sourcecredibility

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

func TestAssessments(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	mustUser := func(t *testing.T, c *database.Catalog) {
		t.Helper()
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, userID, "Jake", r); err != nil {
			t.Fatal(err)
		}
	}
	mustSource := func(t *testing.T, c *database.Catalog) sources.Source {
		t.Helper()
		typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
			Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		s, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
			return sources.Create(tx, userID, sources.CreateInput{
				SourceTypeID: typeID,
				Title:        "Family Bible",
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	mustGrades := func(t *testing.T, c *database.Catalog) []sourcecredibilitygrades.Grade {
		t.Helper()
		if err := sourcecredibilitygrades.Install(c); err != nil {
			t.Fatal(err)
		}
		list, err := sourcecredibilitygrades.List(c)
		if err != nil || len(list) < 3 {
			t.Fatalf("%v len=%d", err, len(list))
		}
		return list
	}
	latestAction := func(t *testing.T, c *database.Catalog) string {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var actionType string
		if err := db.QueryRow(`SELECT action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`).Scan(&actionType); err != nil {
			t.Fatal(err)
		}
		return actionType
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "get missing returns ErrNoRows",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				_, err := GetBySource(c, src.ID)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "upsert create then update",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				grades := mustGrades(t, c)
				low := grades[0]
				high := grades[2]

				a, err := Upsert(c, userID, UpsertInput{
					SourceID:           src.ID,
					CredibilityGradeID: low.ID,
					Argument:           "hearsay copy",
				})
				if err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "create_source_credibility_assessment" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				got, err := GetBySource(c, src.ID)
				if err != nil || string(got.ID) != string(a.ID) || got.Argument != "hearsay copy" {
					t.Fatalf("%v %+v", err, got)
				}

				a2, err := Upsert(c, userID, UpsertInput{
					SourceID:           src.ID,
					CredibilityGradeID: high.ID,
					Argument:           "original register",
				})
				if err != nil {
					t.Fatal(err)
				}
				if string(a2.ID) != string(a.ID) {
					t.Fatalf("id changed %+v vs %+v", a2, a)
				}
				if latestAction(t, c) != "update_source_credibility_assessment" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				got, err = GetBySource(c, src.ID)
				if err != nil || string(got.CredibilityGradeID) != string(high.ID) || got.Argument != "original register" {
					t.Fatalf("%v %+v", err, got)
				}
			},
		},
		{
			name: "reject unknown grade and source",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSource(t, c)
				grades := mustGrades(t, c)
				missing := make([]byte, 16)
				missing[15] = 9
				if _, err := Upsert(c, userID, UpsertInput{SourceID: missing, CredibilityGradeID: grades[0].ID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if _, err := Upsert(c, userID, UpsertInput{SourceID: src.ID, CredibilityGradeID: missing}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad grade %v", err)
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
