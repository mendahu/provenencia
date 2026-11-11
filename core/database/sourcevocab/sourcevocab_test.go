package sourcevocab

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
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
				fields, err := metadatafields.List(c)
				if err != nil {
					t.Fatal(err)
				}
				if len(types) != 1 || len(fields) != 3 {
					t.Fatalf("types=%d fields=%d", len(types), len(fields))
				}
				cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if cert.IconKey != "type_certificate" {
					t.Fatalf("seeded icon_key %+v", cert)
				}
				sugs, err := ListSuggestions(c, cert.ID, vocabkey.LevelSource)
				if err != nil {
					t.Fatal(err)
				}
				wantKeys := []string{"document_number", "record_date", "issue_date"}
				if len(sugs) != 3 {
					t.Fatalf("suggestions %+v", sugs)
				}
				for i, want := range wantKeys {
					if sugs[i].Field.Key != want || sugs[i].SortOrder != i {
						t.Fatalf("suggestion %d = %+v, want %s at %d", i, sugs[i], want, i)
					}
				}
				for _, l := range []vocabkey.Level{vocabkey.LevelArtifact, vocabkey.LevelGraph} {
					other, err := ListSuggestions(c, cert.ID, l)
					if err != nil || len(other) != 0 {
						t.Fatalf("%s suggestions %+v %v", l, other, err)
					}
				}
			},
		},
		{
			name: "a seed entry naming an unknown type is rejected even with no fields",
			run: func(t *testing.T, c *database.Catalog) {
				saved := seedSuggestions
				t.Cleanup(func() { seedSuggestions = saved })
				seedSuggestions = []seedSuggestion{{TypeKey: "no_such_type", Level: vocabkey.LevelSource}}
				if err := Install(c); !errors.Is(err, ErrInvalid) {
					t.Fatalf("Install = %v, want ErrInvalid", err)
				}
			},
		},
		{
			name: "install twice keeps same ids",
			run: func(t *testing.T, c *database.Catalog) {
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				id1 := append([]byte(nil), cert.ID...)
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				cert, err = sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
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
				cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				userID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
				userRef, err := ref.Mint(ref.PrefixUser)
				if err != nil {
					t.Fatal(err)
				}
				if err := users.Upsert(c, userID, "Tester", userRef); err != nil {
					t.Fatal(err)
				}
				if err := runSourceTypeDelete(c, userID, cert.ID); err != nil {
					t.Fatal(err)
				}
				_, err = sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
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
				cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				field, err := metadatafields.Lookup(c, "issue_date", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := DeleteSuggestion(c, cert.ID, field.ID, vocabkey.LevelSource); err != nil {
					t.Fatal(err)
				}
				sugs, err := ListSuggestions(c, cert.ID, vocabkey.LevelSource)
				if err != nil || len(sugs) != 2 {
					t.Fatalf("after delete %v %d", err, len(sugs))
				}
			},
		},
		{
			name: "user origin twin left alone",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "birth_certificate", Origin: vocabkey.OriginUser, Label: "My Birth Cert",
				}); err != nil {
					t.Fatal(err)
				}
				if err := Install(c); err != nil {
					t.Fatal(err)
				}
				userRow, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginUser)
				if err != nil {
					t.Fatal(err)
				}
				if userRow.Label != "My Birth Cert" {
					t.Fatalf("user row mutated: %+v", userRow)
				}
				prov, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
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
					Key: "book", Origin: vocabkey.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				var fieldIDs [][]byte
				for _, label := range []string{"Author", "Publisher"} {
					f, err := runFieldCreate(c, label, metadatafields.DataTypeText, "")
					if err != nil {
						t.Fatal(err)
					}
					fieldIDs = append(fieldIDs, f.ID)
				}
				// Assigned second first, so a plain label sort would disagree
				// with the stored order the UI has to preserve.
				for _, id := range [][]byte{fieldIDs[1], fieldIDs[0]} {
					if err := AppendSuggestion(c, typeID, id, vocabkey.LevelSource); err != nil {
						t.Fatal(err)
					}
				}
				// Re-assigning an existing pair must not move it to the end.
				if err := AppendSuggestion(c, typeID, fieldIDs[1], vocabkey.LevelSource); err != nil {
					t.Fatal(err)
				}
				got, err := ListSuggestions(c, typeID, vocabkey.LevelSource)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 2 || got[0].Field.Label != "Publisher" || got[1].Field.Label != "Author" {
					t.Fatalf("got %+v", got)
				}
				// Each level is its own ordered list: Author appended at
				// artifact starts at 0 and the source list does not move.
				if err := AppendSuggestion(c, typeID, fieldIDs[0], vocabkey.LevelArtifact); err != nil {
					t.Fatal(err)
				}
				art, err := ListSuggestions(c, typeID, vocabkey.LevelArtifact)
				if err != nil || len(art) != 1 || art[0].Field.Label != "Author" || art[0].SortOrder != 0 {
					t.Fatalf("artifact suggestions %+v %v", art, err)
				}
				again, err := ListSuggestions(c, typeID, vocabkey.LevelSource)
				if err != nil || len(again) != 2 || again[0].Field.Label != "Publisher" || again[1].Field.Label != "Author" {
					t.Fatalf("source suggestions moved: %+v %v", again, err)
				}
			},
		},
		{
			name: "removing a suggestion leaves the field in the vocabulary",
			run: func(t *testing.T, c *database.Catalog) {
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: vocabkey.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				f, err := runFieldCreate(c, "Author", metadatafields.DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				if err := AppendSuggestion(c, typeID, f.ID, vocabkey.LevelSource); err != nil {
					t.Fatal(err)
				}
				if err := AppendSuggestion(c, typeID, f.ID, vocabkey.LevelGraph); err != nil {
					t.Fatal(err)
				}
				if before, err := ListSuggestions(c, typeID, vocabkey.LevelGraph); err != nil || len(before) != 1 {
					t.Fatalf("graph suggestions before the delete %+v %v", before, err)
				}
				if err := DeleteSuggestion(c, typeID, f.ID, vocabkey.LevelGraph); err != nil {
					t.Fatal(err)
				}
				graph, err := ListSuggestions(c, typeID, vocabkey.LevelGraph)
				if err != nil || len(graph) != 0 {
					t.Fatalf("graph suggestions %+v %v", graph, err)
				}
				got, err := ListSuggestions(c, typeID, vocabkey.LevelSource)
				if err != nil || len(got) != 1 || got[0].Field.Label != "Author" {
					t.Fatalf("source suggestions %+v %v", got, err)
				}
				if err := DeleteSuggestion(c, typeID, f.ID, vocabkey.LevelSource); err != nil {
					t.Fatal(err)
				}
				got, err = ListSuggestions(c, typeID, vocabkey.LevelSource)
				if err != nil || len(got) != 0 {
					t.Fatalf("suggestions %+v %v", got, err)
				}
				if _, err := metadatafields.GetByID(c, f.ID); err != nil {
					t.Fatalf("field should survive: %v", err)
				}
			},
		},
		{
			name: "bad ids are rejected",
			run: func(t *testing.T, c *database.Catalog) {
				if err := AppendSuggestion(c, []byte{1}, []byte{2}, vocabkey.LevelSource); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "ListSuggestionsTx reads what ListSuggestions reads",
			run: func(t *testing.T, c *database.Catalog) {
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: vocabkey.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				for _, label := range []string{"Author", "Publisher"} {
					f, err := runFieldCreate(c, label, metadatafields.DataTypeText, "")
					if err != nil {
						t.Fatal(err)
					}
					if err := AppendSuggestion(c, typeID, f.ID, vocabkey.LevelSource); err != nil {
						t.Fatal(err)
					}
				}
				want, err := ListSuggestions(c, typeID, vocabkey.LevelSource)
				if err != nil || len(want) != 2 {
					t.Fatalf("ListSuggestions = %+v, %v", want, err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				tx, err := db.Begin()
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback() }()
				got, err := ListSuggestionsTx(tx, typeID, vocabkey.LevelSource)
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != len(want) {
					t.Fatalf("got %d suggestions, want %d", len(got), len(want))
				}
				for i := range got {
					if !bytes.Equal(got[i].Field.ID, want[i].Field.ID) || got[i].SortOrder != want[i].SortOrder {
						t.Fatalf("suggestion %d = %+v, want %+v", i, got[i], want[i])
					}
				}
				if _, err := ListSuggestionsTx(tx, typeID, vocabkey.LevelCitation); !errors.Is(err, ErrInvalid) {
					t.Fatalf("citation level = %v", err)
				}
			},
		},
		{
			name: "ListSuggestionsTx with a nil tx is invalid",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := ListSuggestionsTx(nil, make([]byte, 16), vocabkey.LevelSource); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "levels that hold no metadata are rejected",
			run: func(t *testing.T, c *database.Catalog) {
				typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
					Key: "book", Origin: vocabkey.OriginUser, Label: "Book",
				})
				if err != nil {
					t.Fatal(err)
				}
				f, err := runFieldCreate(c, "Author", metadatafields.DataTypeText, "")
				if err != nil {
					t.Fatal(err)
				}
				for _, l := range []vocabkey.Level{"page", vocabkey.LevelCitation, "", "Source"} {
					if err := AppendSuggestion(c, typeID, f.ID, l); !errors.Is(err, ErrInvalid) {
						t.Fatalf("AppendSuggestion(%q) = %v", l, err)
					}
					if err := EnsureSuggestion(c, typeID, f.ID, l, 0); !errors.Is(err, ErrInvalid) {
						t.Fatalf("EnsureSuggestion(%q) = %v", l, err)
					}
					if _, err := ListSuggestions(c, typeID, l); !errors.Is(err, ErrInvalid) {
						t.Fatalf("ListSuggestions(%q) = %v", l, err)
					}
					if err := DeleteSuggestion(c, typeID, f.ID, l); !errors.Is(err, ErrInvalid) {
						t.Fatalf("DeleteSuggestion(%q) = %v", l, err)
					}
				}
				for _, l := range vocabkey.MetadataLevels() {
					if got, err := ListSuggestions(c, typeID, l); err != nil || len(got) != 0 {
						t.Fatalf("a refused write left %s rows: %+v %v", l, got, err)
					}
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

func runSourceTypeDelete(c *database.Catalog, userID, id []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "delete_source_type", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := sourcetypes.Delete(tx, userID, id)
			return struct{}{}, changes, err
		})
	return err
}

func runFieldCreate(c *database.Catalog, label, dataType, description string) (metadatafields.Field, error) {
	got, _, err := writes.Run(c, writes.Op{Action: "create_metadata_field"},
		func(tx *database.Tx) (metadatafields.Field, []rowchange.Change, error) {
			return metadatafields.Create(tx, label, dataType, description)
		})
	return got, err
}
