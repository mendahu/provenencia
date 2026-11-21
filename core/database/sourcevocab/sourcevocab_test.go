package sourcevocab

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/Jake-Robins-Studios/provenencia/core/citationformat"
	"github.com/Jake-Robins-Studios/provenencia/core/database"
	"github.com/Jake-Robins-Studios/provenencia/core/database/dbtest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/metadatafields"
	"github.com/Jake-Robins-Studios/provenencia/core/database/metadatafields/metadatafieldstest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/rowchange"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcetypes"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcetypes/sourcetypestest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/users"
	"github.com/Jake-Robins-Studios/provenencia/core/ref"
	"github.com/Jake-Robins-Studios/provenencia/core/vocabkey"
	"github.com/Jake-Robins-Studios/provenencia/core/writes"
)

// install runs Install the way createCatalog does: one writes.Call, no user.
func install(c *database.Catalog) error {
	_, err := writes.RunChanges(c, writes.Op{Action: ActionInstall}, func(tx *database.Tx) ([]rowchange.Change, error) {
		return Install(tx)
	})
	return err
}

// storedFormats reads a type's formats as stored (fields by id). It opens and
// closes its own transaction, so call it with none held.
func storedFormats(t *testing.T, c *database.Catalog, typeID []byte) sourcetypes.Formats {
	t.Helper()
	tx, err := dbtest.MustDB(t, c).Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	got, err := sourcetypes.StoredFormatsTx(tx, [][]byte{typeID})
	if err != nil {
		t.Fatal(err)
	}
	return got[string(typeID)]
}

func findSeedType(t *testing.T, key string) seedType {
	t.Helper()
	for _, st := range seedTypes {
		if st.Key == key {
			return st
		}
	}
	t.Fatalf("no seed type %s", key)
	return seedType{}
}

// catalogSnapshot is everything Install wrote that a second run must leave alone.
type catalogSnapshot struct {
	Types       map[string]string // key -> id, label, stored formats
	Fields      map[string]string // key -> id, label, unknown text
	Suggestions map[string][]string
}

func snapshot(t *testing.T, c *database.Catalog) catalogSnapshot {
	t.Helper()
	snap := catalogSnapshot{Types: map[string]string{}, Fields: map[string]string{}, Suggestions: map[string][]string{}}
	fields, err := metadatafields.List(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fields {
		snap.Fields[f.Key] = string(f.ID) + "|" + f.Label + "|" + f.UnknownText
	}
	types, err := sourcetypes.List(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, ty := range types {
		f := storedFormats(t, c, ty.ID)
		snap.Types[ty.Key] = string(ty.ID) + "|" + ty.Label + "|" + f.Full + "|" + f.Short + "|" + f.List
		for _, level := range vocabkey.MetadataLevels() {
			sugs, err := ListSuggestions(c, ty.ID, level)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range sugs {
				k := ty.Key + "/" + string(level)
				snap.Suggestions[k] = append(snap.Suggestions[k], fmt.Sprintf("%s@%d", s.Field.Key, s.SortOrder))
			}
		}
	}
	return snap
}

func TestInstall(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "empty catalog gets the seed",
			run: func(t *testing.T, c *database.Catalog) {
				if err := install(c); err != nil {
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
				// The map's counts, asserted once; every other check reads the registry.
				if len(seedTypes) != 26 || len(seedFields) != 67 {
					t.Fatalf("registry holds %d types and %d fields, the map says 26 and 67", len(seedTypes), len(seedFields))
				}
				if len(types) != len(seedTypes) || len(fields) != len(seedFields) {
					t.Fatalf("types=%d fields=%d", len(types), len(fields))
				}
				byKey := map[string]metadatafields.Field{}
				byID := map[string]metadatafields.Field{}
				for _, f := range fields {
					byKey[f.Key] = f
					byID[string(f.ID)] = f
				}
				for _, want := range seedFields {
					got, ok := byKey[want.Key]
					if !ok || got.Origin != vocabkey.OriginProvenencia || got.Label != want.Label ||
						got.DataType != want.DataType || got.Description != want.Description || got.UnknownText != want.UnknownText {
						t.Fatalf("field %s = %+v, want %+v", want.Key, got, want)
					}
				}
				for _, want := range seedTypes {
					got, err := sourcetypes.Lookup(c, want.Key, vocabkey.OriginProvenencia)
					if err != nil {
						t.Fatalf("type %s: %v", want.Key, err)
					}
					if got.Label != want.Label || got.Description != want.Description || got.IconKey != want.IconKey {
						t.Fatalf("type %s = %+v, want %+v", want.Key, got, want)
					}
				}

				census, err := sourcetypes.Lookup(c, "census", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				stored := storedFormats(t, c, census.ID)
				for _, form := range citationformat.Forms() {
					text := stored.Text(form)
					tpl, diags := citationformat.Parse(text, form)
					if len(diags) > 0 {
						t.Fatalf("stored %s: %v", form, diags)
					}
					want, _ := citationformat.Parse(findSeedType(t, "census").Formats.Text(form), form)
					var wantKeys, gotKeys []string
					for _, f := range want.Fields() {
						wantKeys = append(wantKeys, f.Key)
					}
					for _, f := range tpl.Fields() {
						field, ok := byID[string(f.ID[:])]
						if !f.ByID() || !ok || field.Origin != vocabkey.OriginProvenencia {
							t.Fatalf("stored %s names %+v, want a provenencia field id", form, f)
						}
						gotKeys = append(gotKeys, field.Key)
					}
					if !slices.Equal(gotKeys, wantKeys) {
						t.Fatalf("stored %s reads %v, want %v", form, gotKeys, wantKeys)
					}
				}
				view, err := sourcetypes.GetFormats(c, census.ID)
				if err != nil {
					t.Fatal(err)
				}
				if view.Formats != findSeedType(t, "census").Formats || len(view.Choices) != 0 || len(view.Diagnostics) != 0 {
					t.Fatalf("GetFormats = %+v", view)
				}

				// Every suggestion list is written in registry order, and a
				// (type, level) the registry does not list has no rows.
				listed := map[string]bool{}
				for _, want := range seedSuggestions {
					ty, err := sourcetypes.Lookup(c, want.TypeKey, vocabkey.OriginProvenencia)
					if err != nil {
						t.Fatalf("type %s: %v", want.TypeKey, err)
					}
					listed[want.TypeKey+"/"+string(want.Level)] = true
					sugs, err := ListSuggestions(c, ty.ID, want.Level)
					if err != nil {
						t.Fatal(err)
					}
					if len(sugs) != len(want.FieldKeys) {
						t.Fatalf("%s/%s lists %d fields, want %d", want.TypeKey, want.Level, len(sugs), len(want.FieldKeys))
					}
					for i, s := range sugs {
						if s.Field.Key != want.FieldKeys[i] || s.Field.Origin != vocabkey.OriginProvenencia || s.SortOrder != i {
							t.Fatalf("%s/%s[%d] = %s origin %s sort %d, want %s origin provenencia sort %d",
								want.TypeKey, want.Level, i, s.Field.Key, s.Field.Origin, s.SortOrder, want.FieldKeys[i], i)
						}
					}
				}
				for _, ty := range types {
					for _, level := range vocabkey.MetadataLevels() {
						if listed[ty.Key+"/"+string(level)] {
							continue
						}
						sugs, err := ListSuggestions(c, ty.ID, level)
						if err != nil || len(sugs) != 0 {
							t.Fatalf("%s/%s = %+v, %v, want no suggestions", ty.Key, level, sugs, err)
						}
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
				if err := install(c); !errors.Is(err, ErrInvalid) {
					t.Fatalf("Install = %v, want ErrInvalid", err)
				}
				// The valid fields and type ahead of the bad suggestion are rolled back.
				fields, err := metadatafields.List(c)
				if err != nil || len(fields) != 0 {
					t.Fatalf("fields after a failed install: %v %v", fields, err)
				}
				types, err := sourcetypes.List(c)
				if err != nil || len(types) != 0 {
					t.Fatalf("types after a failed install: %v %v", types, err)
				}
			},
		},
		{
			name: "install twice changes nothing",
			run: func(t *testing.T, c *database.Catalog) {
				if err := install(c); err != nil {
					t.Fatal(err)
				}
				before := snapshot(t, c)
				if err := install(c); err != nil {
					t.Fatal(err)
				}
				if after := snapshot(t, c); !reflect.DeepEqual(before, after) {
					t.Fatalf("a second install changed the catalog:\nbefore %+v\nafter  %+v", before, after)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if got := dbtest.Revisions(t, db); got != 1 {
					t.Fatalf("revisions after a second install = %d, want 1", got)
				}
			},
		},
		{
			name: "a second install does not restore a deleted suggestion",
			run: func(t *testing.T, c *database.Catalog) {
				if err := install(c); err != nil {
					t.Fatal(err)
				}
				census, err := sourcetypes.Lookup(c, "census", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				field, err := metadatafields.Lookup(c, "piece", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := DeleteSuggestion(c, census.ID, field.ID, vocabkey.LevelSource); err != nil {
					t.Fatal(err)
				}
				if err := install(c); err != nil {
					t.Fatal(err)
				}
				sugs, err := ListSuggestions(c, census.ID, vocabkey.LevelSource)
				if err != nil || len(sugs) != 4 {
					t.Fatalf("after second install %v %d", err, len(sugs))
				}
			},
		},
		{
			name: "install records one revision",
			run: func(t *testing.T, c *database.Catalog) {
				if err := install(c); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if got := dbtest.Revisions(t, db); got != 1 {
					t.Fatalf("revisions = %d, want 1", got)
				}
				action, changes := dbtest.LatestRevision(t, db)
				if action != ActionInstall {
					t.Fatalf("action = %q", action)
				}
				counts := map[string]int{}
				for _, ch := range changes {
					if ch.Action != rowchange.ActionCreate {
						t.Fatalf("change %+v", ch)
					}
					counts[ch.EntityType]++
				}
				if len(counts) != 2 || counts["source_type"] != len(seedTypes) || counts["metadata_field"] != len(seedFields) {
					t.Fatalf("changes by entity = %v", counts)
				}
				docs := map[string]int{}
				rows, err := db.Query(`SELECT kind, COUNT(*) FROM catalog_search_docs GROUP BY kind`)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				for rows.Next() {
					var kind string
					var n int
					if err := rows.Scan(&kind, &n); err != nil {
						t.Fatal(err)
					}
					docs[kind] = n
				}
				if err := rows.Err(); err != nil {
					t.Fatal(err)
				}
				if len(docs) != 2 || docs["source_type"] != len(seedTypes) || docs["metadata_field"] != len(seedFields) {
					t.Fatalf("search docs by kind = %v", docs)
				}
				var nullUser int
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions WHERE user_id IS NULL`).Scan(&nullUser); err != nil || nullUser != 1 {
					t.Fatalf("null-user revisions = %d %v", nullUser, err)
				}
			},
		},
		{
			name: "deleted provenencia type stays deleted without reinstall",
			run: func(t *testing.T, c *database.Catalog) {
				if err := install(c); err != nil {
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
				if err := install(c); err != nil {
					t.Fatal(err)
				}
				census, err := sourcetypes.Lookup(c, "census", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				field, err := metadatafields.Lookup(c, "piece", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				if err := DeleteSuggestion(c, census.ID, field.ID, vocabkey.LevelSource); err != nil {
					t.Fatal(err)
				}
				sugs, err := ListSuggestions(c, census.ID, vocabkey.LevelSource)
				if err != nil || len(sugs) != 4 {
					t.Fatalf("after delete %v %d", err, len(sugs))
				}
			},
		},
		{
			name: "a user field with a seeded key does not capture a format",
			run: func(t *testing.T, c *database.Catalog) {
				userAuthor, err := metadatafieldstest.Upsert(c, metadatafields.Field{
					Key: "author", Origin: vocabkey.OriginUser, Label: "My author", DataType: metadatafields.DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := install(c); err != nil {
					t.Fatal(err)
				}
				book, err := sourcetypes.Lookup(c, "book", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				fields, err := metadatafields.List(c)
				if err != nil {
					t.Fatal(err)
				}
				origin := map[string]string{}
				for _, f := range fields {
					origin[string(f.ID)] = f.Origin
				}
				stored := storedFormats(t, c, book.ID)
				named := 0
				for _, form := range citationformat.Forms() {
					tpl, diags := citationformat.Parse(stored.Text(form), form)
					if len(diags) > 0 {
						t.Fatal(diags)
					}
					for _, f := range tpl.Fields() {
						named++
						if bytes.Equal(f.ID[:], userAuthor) {
							t.Fatalf("book's %s format names the user's author", form)
						}
						if origin[string(f.ID[:])] != vocabkey.OriginProvenencia {
							t.Fatalf("book's %s format names %+v, not a provenencia field", form, f)
						}
					}
				}
				if named == 0 {
					t.Fatal("book's formats name no field")
				}
			},
		},
		{
			name: "user origin twin left alone",
			run: func(t *testing.T, c *database.Catalog) {
				if _, err := sourcetypestest.Upsert(c, sourcetypes.Type{
					Key: "birth_certificate", Origin: vocabkey.OriginUser, Label: "My Birth Cert",
				}); err != nil {
					t.Fatal(err)
				}
				if err := install(c); err != nil {
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
				typeID, err := sourcetypestest.Upsert(c, sourcetypes.Type{
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
				typeID, err := sourcetypestest.Upsert(c, sourcetypes.Type{
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
				typeID, err := sourcetypestest.Upsert(c, sourcetypes.Type{
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
					if label == "Publisher" {
						unknownText := "n.p."
						if _, _, err := writes.Run(c, writes.Op{Action: "update_metadata_field"},
							func(tx *database.Tx) (metadatafields.Field, []rowchange.Change, error) {
								return metadatafields.Update(tx, f.ID, f.Label, f.DataType, "", &unknownText)
							}); err != nil {
							t.Fatal(err)
						}
					}
				}
				want, err := ListSuggestions(c, typeID, vocabkey.LevelSource)
				if err != nil || len(want) != 2 || want[1].Field.UnknownText != "n.p." {
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
					if !bytes.Equal(got[i].Field.ID, want[i].Field.ID) || got[i].SortOrder != want[i].SortOrder || got[i].Field.UnknownText != want[i].Field.UnknownText {
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
				typeID, err := sourcetypestest.Upsert(c, sourcetypes.Type{
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
					_, err := writes.RunChanges(c, writes.Op{Action: "ensure_suggestion"}, func(tx *database.Tx) ([]rowchange.Change, error) {
						return nil, ensureSuggestionTx(tx, typeID, f.ID, l, 0)
					})
					if !errors.Is(err, ErrInvalid) {
						t.Fatalf("ensureSuggestionTx(%q) = %v", l, err)
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
	_, err := writes.RunChanges(c, writes.Op{Action: "delete_source_type", UserID: userID},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			return sourcetypes.Delete(tx, userID, id)
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
