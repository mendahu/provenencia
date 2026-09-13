package sourcevocab

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
)

func TestInstall(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "empty catalog gets trimmed seed",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				types, err := sourcetypes.List(c)
				if err != nil {
					t.Fatal(err)
				}
				fields, err := sourcefields.List(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(types) != 1 || len(fields) != 3 {
					t.Fatalf("types=%d fields=%d", len(types), len(fields))
				}
				cert, err := sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if cert.IconKey != "type_certificate" {
					t.Fatalf("seeded icon_key %+v", cert)
				}
				sugs, err := ListSuggestions(c, cert.ID)
				if err != nil {
					t.Fatal(err)
				}
				if len(sugs) != 3 || sugs[0].Field.Key != "document_number" || sugs[0].SortOrder != 0 {
					t.Fatalf("suggestions %+v", sugs)
				}
			},
		},
		{
			name: "install twice keeps same ids",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				cert, err := sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				id1 := append([]byte(nil), cert.ID...)
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				cert, err = sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if string(cert.ID) != string(id1) {
					t.Fatal("id changed on second install")
				}
				types, err := sourcetypes.List(c)
				if err != nil || len(types) != 1 {
					t.Fatalf("types %v %d", err, len(types))
				}
			},
		},
		{
			name: "deleted provenencia type stays deleted without reinstall",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				cert, err := sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := sourcetypes.Delete(c, cert.ID); err != nil {
					t.Fatal(err)
				}
				_, err = sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
				if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("want ErrNoRows got %v", err)
				}
			},
		},
		{
			name: "deleted suggestion join stays deleted without reinstall",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				cert, err := sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				field, err := sourcefields.Lookup(c, "issue_date", sourcefields.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := DeleteSuggestion(c, cert.ID, field.ID); err != nil {
					t.Fatal(err)
				}
				sugs, err := ListSuggestions(c, cert.ID)
				if err != nil || len(sugs) != 2 {
					t.Fatalf("after delete %v %d", err, len(sugs))
				}
			},
		},
		{
			name: "user origin twin left alone",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "birth_certificate", Origin: sourcetypes.OriginUser, Label: "My Birth Cert",
				}); err != nil {
					t.Fatal(err)
				}
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				userRow, err := sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginUser)
				if err != nil {
					t.Fatal(err)
				}
				if userRow.Label != "My Birth Cert" {
					t.Fatalf("user row mutated: %+v", userRow)
				}
				prov, err := sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if prov.Label != "Birth certificate" {
					t.Fatalf("provenencia %+v", prov)
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

func TestAppendSuggestion(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "appends in call order and is idempotent",
			run: func(t *testing.T, c *database.Catalog) {
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: sourcetypes.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				var fieldIDs [][]byte
				for _, label := range []string{"Author", "Publisher"} {
					f, err := sourcefields.Create(c, label, sourcefields.DataTypeText, "")
					if err != nil {
						t.Fatal(err)
					}
					fieldIDs = append(fieldIDs, f.ID)
				}
				// Assigned second first, so a plain label sort would disagree
				// with the stored order the UI has to preserve.
				for _, id := range [][]byte{fieldIDs[1], fieldIDs[0]} {
					if err := AppendSuggestion(c, typeID, id); err != nil {
						t.Fatal(err)
					}
				}
				// Re-assigning an existing pair must not move it to the end.
				if err := AppendSuggestion(c, typeID, fieldIDs[1]); err != nil {
					t.Fatal(err)
				}
				got, err := ListSuggestions(c, typeID)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 2 || got[0].Field.Label != "Publisher" || got[1].Field.Label != "Author" {
					t.Fatalf("got %+v", got)
				}
			},
		},
		{
			name: "removing a suggestion leaves the field in the vocabulary",
			run: func(t *testing.T, c *database.Catalog) {
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: sourcetypes.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				f, err := sourcefields.Create(c, "Author", sourcefields.DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if err := AppendSuggestion(c, typeID, f.ID); err != nil {
					t.Fatal(err)
				}
				if err := DeleteSuggestion(c, typeID, f.ID); err != nil {
					t.Fatal(err)
				}
				got, err := ListSuggestions(c, typeID)
				if err != nil || len(got) != 0 {
					t.Fatalf("suggestions %+v %v", got, err)
				}
				if _, err := sourcefields.GetByID(c, f.ID); err != nil {
					t.Fatalf("field should survive: %v", err)
				}
			},
		},
		{
			name: "bad ids are rejected",
			run: func(t *testing.T, c *database.Catalog) {
				if err := AppendSuggestion(c, []byte{1}, []byte{2}); !errors.Is(err, ErrInvalid) {
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
