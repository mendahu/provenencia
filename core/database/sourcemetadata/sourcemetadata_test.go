package sourcemetadata

import (
	"bytes"
	"errors"
	"math/rand"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestSourceMetadata(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	mustField := func(t *testing.T, c *database.Catalog, key string) metadatafields.Field {
		t.Helper()
		f, err := metadatafields.Lookup(c, key, vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		return f
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
	workspaceKeys := func(t *testing.T, c *database.Catalog, sourceID []byte) []string {
		t.Helper()
		ws, err := ListWorkspace(c, sourceOwner(sourceID))
		if err != nil {
			t.Fatal(err)
		}
		keys := make([]string, 0, len(ws))
		for _, e := range ws {
			keys = append(keys, e.Field.Key)
		}
		return keys
	}
	hasKey := func(keys []string, key string) bool {
		for _, k := range keys {
			if k == key {
				return true
			}
		}
		return false
	}
	layoutOf := func(t *testing.T, c *database.Catalog, sourceID, fieldID []byte) (int, int) {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var sortOrder, dismissed int
		if err := db.QueryRow(
			`SELECT sort_order, dismissed FROM source_metadata_layout WHERE source_id = ? AND field_id = ?`,
			sourceID, fieldID,
		).Scan(&sortOrder, &dismissed); err != nil {
			t.Fatal(err)
		}
		return sortOrder, dismissed
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "source_metadata has no date_value_id",
			run: func(t *testing.T, c *database.Catalog) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				rows, err := db.Query(`PRAGMA table_info(source_metadata)`)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				for rows.Next() {
					var cid int
					var name, typ string
					var notnull, pk int
					var dflt any
					if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
						t.Fatal(err)
					}
					if name == "date_value_id" {
						t.Fatal("date_value_id still on source_metadata")
					}
				}
			},
		},
		{
			name: "set text metadata",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				row, err := runSet(c, userID, src.ID, Input{
					FieldID: doc.ID, ValueText: "12345",
				})
				if err != nil {
					t.Fatal(err)
				}
				if row.ValueText != "12345" {
					t.Fatalf("%+v", row)
				}
				if latestAction(t, c) != "update_source_metadata" {
					t.Fatalf("action %q", latestAction(t, c))
				}
			},
		},
		{
			name: "set url and reject date on url field",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				field, err := runFieldCreate(c, "Landing page", metadatafields.DataTypeURL, "")
				if err != nil {
					t.Fatal(err)
				}
				row, err := runSet(c, userID, src.ID, Input{
					FieldID: field.ID, ValueText: "https://example.com/record",
				})
				if err != nil {
					t.Fatal(err)
				}
				if row.ValueText != "https://example.com/record" {
					t.Fatalf("%+v", row)
				}
				if _, err := runSet(c, userID, src.ID, Input{
					FieldID: field.ID, ValueText: "www.url.com",
				}); err != nil {
					t.Fatal(err)
				}
				lowered, err := runSet(c, userID, src.ID, Input{
					FieldID: field.ID, ValueText: "HTTPS://Example.COM/Record",
				})
				if err != nil {
					t.Fatal(err)
				}
				if lowered.ValueText != "https://example.com/record" {
					t.Fatalf("canonical %q", lowered.ValueText)
				}
				if _, err := runSet(c, userID, src.ID, Input{
					FieldID: field.ID, ValueText: "file:/tmp",
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("file url %v", err)
				}
				if _, err := runSet(c, userID, src.ID, Input{
					FieldID: field.ID, ValueText: "not a url",
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("junk url %v", err)
				}
			},
		},
		{
			name: "date-named fields are text",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				rec := mustField(t, c, "record_date")
				if rec.DataType != metadatafields.DataTypeText {
					t.Fatalf("record_date type %q", rec.DataType)
				}
				row, err := runSet(c, userID, src.ID, Input{
					FieldID: rec.ID, ValueText: "about the year 1890",
				})
				if err != nil {
					t.Fatal(err)
				}
				if row.ValueText != "about the year 1890" {
					t.Fatalf("%+v", row)
				}
				kept, err := runSet(c, userID, src.ID, Input{
					FieldID:   rec.ID,
					ValueText: "circa 1890",
				})
				if err != nil {
					t.Fatal(err)
				}
				if kept.ValueText != "circa 1890" {
					t.Fatalf("%+v", kept)
				}
			},
		},
		{
			name: "clear and upsert same pair",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				first, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "111"})
				if err != nil {
					t.Fatal(err)
				}
				second, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "222"})
				if err != nil {
					t.Fatal(err)
				}
				if string(first.ID) != string(second.ID) || second.ValueText != "222" {
					t.Fatalf("first=%+v second=%+v", first, second)
				}
				list := values(t, c, src.ID)
				if len(list) != 1 {
					t.Fatalf("len=%d", len(list))
				}
				if err := runClear(c, userID, src.ID, doc.ID); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_source_metadata" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				list = values(t, c, src.ID)
				if len(list) != 0 {
					t.Fatalf("after clear len=%d", len(list))
				}
				if err := runClear(c, userID, src.ID, doc.ID); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "list workspace suggested and extras",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "12345"}); err != nil {
					t.Fatal(err)
				}
				extraID, err := metadatafields.Upsert(c, metadatafields.Field{
					Key: "shelf_mark", Origin: vocabkey.OriginUser, Label: "Shelf mark", DataType: metadatafields.DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runSet(c, userID, src.ID, Input{FieldID: extraID, ValueText: "A-12"}); err != nil {
					t.Fatal(err)
				}
				ws, err := ListWorkspace(c, sourceOwner(src.ID))
				if err != nil {
					t.Fatal(err)
				}
				var sawDoc, sawUnsetSuggested, sawExtra bool
				for _, e := range ws {
					switch e.Field.Key {
					case "document_number":
						sawDoc = true
						if !e.Suggested || e.Value == nil || e.Value.ValueText != "12345" {
							t.Fatalf("document_number %+v", e)
						}
					case "record_date", "issue_date":
						if e.Suggested && e.Value == nil {
							sawUnsetSuggested = true
						}
					case "shelf_mark":
						sawExtra = true
						if e.Suggested || e.Value == nil || e.Value.ValueText != "A-12" {
							t.Fatalf("extra %+v", e)
						}
					}
				}
				if !sawDoc || !sawUnsetSuggested || !sawExtra {
					t.Fatalf("doc=%v unset=%v extra=%v entries=%d", sawDoc, sawUnsetSuggested, sawExtra, len(ws))
				}
			},
		},
		{
			name: "list workspace suggests the source level only",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				fieldID, err := metadatafields.Upsert(c, metadatafields.Field{
					Key: "shelf_mark", Origin: vocabkey.OriginUser, Label: "Shelf mark", DataType: metadatafields.DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := sourcevocab.AppendSuggestion(c, src.SourceTypeID, fieldID, vocabkey.LevelGraph); err != nil {
					t.Fatal(err)
				}
				ws, err := ListWorkspace(c, sourceOwner(src.ID))
				if err != nil {
					t.Fatal(err)
				}
				suggested := 0
				for _, e := range ws {
					if e.Field.Key == "shelf_mark" {
						t.Fatalf("graph-level suggestion listed on the Source page: %+v", e)
					}
					if e.Suggested {
						suggested++
					}
				}
				if suggested != 3 {
					t.Fatalf("suggested = %d, want the 3 seeded source fields", suggested)
				}
			},
		},
		{
			name: "source delete cascades metadata",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "x"}); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM sources WHERE id = ?`, src.ID); err != nil {
					t.Fatal(err)
				}
				if countRows(t, c, "source_metadata") != 0 {
					t.Fatalf("count %d", countRows(t, c, "source_metadata"))
				}
			},
		},
		{
			name: "layout table shipped with catalog format",
			run: func(t *testing.T, c *database.Catalog) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 13 {
					t.Fatalf("user_version %d want >= 13", ver)
				}
				rows, err := db.Query(`PRAGMA table_info(source_metadata_layout)`)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				got := map[string]struct{}{}
				for rows.Next() {
					var cid int
					var name, colType string
					var notNull int
					var dflt any
					var pk int
					if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
						t.Fatal(err)
					}
					got[name] = struct{}{}
				}
				if err := rows.Err(); err != nil {
					t.Fatal(err)
				}
				for _, want := range []string{"source_id", "field_id", "sort_order", "dismissed"} {
					if _, ok := got[want]; !ok {
						t.Fatalf("column %q missing, got %v", want, got)
					}
				}
			},
		},
		{
			name: "dismiss hides empty suggestion and persists",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				issue := mustField(t, c, "issue_date")
				if err := runDismissSuggestion(c, userID, src.ID, issue.ID); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "dismiss_source_metadata_suggestion" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				if _, dismissed := layoutOf(t, c, src.ID, issue.ID); dismissed != 1 {
					t.Fatalf("dismissed %d", dismissed)
				}
				keys := workspaceKeys(t, c, src.ID)
				if hasKey(keys, "issue_date") || !hasKey(keys, "document_number") {
					t.Fatalf("keys %v", keys)
				}
				// Dismissing twice must stay a no-op rather than error.
				if err := runDismissSuggestion(c, userID, src.ID, issue.ID); err != nil {
					t.Fatal(err)
				}
				if keys := workspaceKeys(t, c, src.ID); hasKey(keys, "issue_date") {
					t.Fatalf("re-dismiss keys %v", keys)
				}
			},
		},
		{
			name: "dismiss with existing value keeps the value visible",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "12345"}); err != nil {
					t.Fatal(err)
				}
				if err := runDismissSuggestion(c, userID, src.ID, doc.ID); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_source_metadata" {
					t.Fatalf("dismiss audited a value field: %q", latestAction(t, c))
				}
				if _, dismissed := layoutOf(t, c, src.ID, doc.ID); dismissed != 0 {
					t.Fatalf("dismissed %d", dismissed)
				}
				keys := workspaceKeys(t, c, src.ID)
				if !hasKey(keys, "document_number") {
					t.Fatalf("keys %v", keys)
				}
			},
		},
		{
			name: "reorder rewrites workspace order and keeps dismissals",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				rec := mustField(t, c, "record_date")
				issue := mustField(t, c, "issue_date")
				if err := runDismissSuggestion(c, userID, src.ID, issue.ID); err != nil {
					t.Fatal(err)
				}
				if err := runReorder(c, userID, src.ID, [][]byte{rec.ID, doc.ID, issue.ID}); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "reorder_source_metadata" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				keys := workspaceKeys(t, c, src.ID)
				if len(keys) != 2 || keys[0] != "record_date" || keys[1] != "document_number" {
					t.Fatalf("keys %v", keys)
				}
				if order, dismissed := layoutOf(t, c, src.ID, issue.ID); order != 2 || dismissed != 1 {
					t.Fatalf("issue_date layout order=%d dismissed=%d", order, dismissed)
				}
			},
		},
		{
			name: "set gives a new field a layout slot at the end",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				extraID, err := metadatafields.Upsert(c, metadatafields.Field{
					Key: "shelf_mark", Origin: vocabkey.OriginUser, Label: "Shelf mark", DataType: metadatafields.DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := runSet(c, userID, src.ID, Input{FieldID: extraID, ValueText: "A-12"}); err != nil {
					t.Fatal(err)
				}
				if order, dismissed := layoutOf(t, c, src.ID, extraID); order != 0 || dismissed != 0 {
					t.Fatalf("extra layout order=%d dismissed=%d", order, dismissed)
				}
				doc := mustField(t, c, "document_number")
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "12345"}); err != nil {
					t.Fatal(err)
				}
				if order, _ := layoutOf(t, c, src.ID, doc.ID); order != 1 {
					t.Fatalf("document_number order %d", order)
				}
				// Updating an existing value must not append a second slot.
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "999"}); err != nil {
					t.Fatal(err)
				}
				if order, _ := layoutOf(t, c, src.ID, doc.ID); order != 1 {
					t.Fatalf("document_number order after update %d", order)
				}
				keys := workspaceKeys(t, c, src.ID)
				want := []string{"shelf_mark", "document_number", "record_date", "issue_date"}
				if len(keys) != len(want) {
					t.Fatalf("keys %v want %v", keys, want)
				}
				for i := range want {
					if keys[i] != want[i] {
						t.Fatalf("keys %v want %v", keys, want)
					}
				}
			},
		},
		{
			name: "reject bad layout arguments",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				missing := make([]byte, 16)
				missing[15] = 9
				if err := runDismissSuggestion(c, userID, missing, doc.ID); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if err := runDismissSuggestion(c, userID, src.ID, missing); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad field %v", err)
				}
				if err := runDismissSuggestion(c, nil, src.ID, doc.ID); !errors.Is(err, ErrInvalid) {
					t.Fatalf("nil user %v", err)
				}
				if err := runReorder(c, userID, src.ID, nil); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty list %v", err)
				}
				if err := runReorder(c, userID, src.ID, [][]byte{doc.ID, doc.ID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("duplicate field %v", err)
				}
				if err := runReorder(c, userID, src.ID, [][]byte{missing}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("missing field %v", err)
				}
				if err := runReorder(c, userID, missing, [][]byte{doc.ID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if err := runReorder(c, userID, src.ID, [][]byte{{1, 2, 3}}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("short field id %v", err)
				}
			},
		},
		{
			name: "source delete cascades layout",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "x"}); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM sources WHERE id = ?`, src.ID); err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM source_metadata_layout`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 0 {
					t.Fatalf("layout rows %d", n)
				}
			},
		},
		{
			name: "reject bad ids and empty text set",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				missing := make([]byte, 16)
				missing[15] = 9
				if _, err := runSet(c, userID, missing, Input{FieldID: doc.ID, ValueText: "x"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if _, err := runSet(c, userID, src.ID, Input{FieldID: missing, ValueText: "x"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad field %v", err)
				}
				if _, err := runSet(c, nil, src.ID, Input{FieldID: doc.ID, ValueText: "x"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("nil user %v", err)
				}
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty text %v", err)
				}
			},
		},
		{
			name: "list workspace reads several extras",
			run: func(t *testing.T, c *database.Catalog) {
				src := birthOfAlice(t, c, userID)
				doc := mustField(t, c, "document_number")
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "12345"}); err != nil {
					t.Fatal(err)
				}
				for _, label := range []string{"a", "b", "c"} {
					f, err := runFieldCreate(c, label, metadatafields.DataTypeText, "")
					if err != nil {
						t.Fatal(err)
					}
					if _, err := runSet(c, userID, src.ID, Input{FieldID: f.ID, ValueText: "v-" + label}); err != nil {
						t.Fatal(err)
					}
				}
				// Layout holds every filled field in the order it was filled; drop it
				// so the list shows suggestions in type order, then extras by field id.
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM source_metadata_layout`); err != nil {
					t.Fatal(err)
				}
				ws, err := ListWorkspace(c, sourceOwner(src.ID))
				if err != nil {
					t.Fatal(err)
				}
				var suggested, extras []string
				for _, e := range ws {
					if e.Suggested {
						suggested = append(suggested, e.Field.Key)
						continue
					}
					if e.Value == nil || e.Value.ValueText != "v-"+e.Field.Key || e.Field.Label != e.Field.Key {
						t.Fatalf("extra %+v", e)
					}
					extras = append(extras, e.Field.Key)
				}
				if len(suggested) == 0 || !slices.Contains(suggested, "document_number") {
					t.Fatalf("suggested %v", suggested)
				}
				if !slices.Equal(extras, []string{"a", "b", "c"}) {
					t.Fatalf("extras %v", extras)
				}
				list, err := sourcevocab.ListSuggestions(c, src.SourceTypeID, vocabkey.LevelSource)
				if err != nil {
					t.Fatal(err)
				}
				want := make([]string, 0, len(list)+3)
				for _, s := range list {
					want = append(want, s.Field.Key)
				}
				want = append(want, "a", "b", "c")
				if got := workspaceKeys(t, c, src.ID); !slices.Equal(got, want) {
					t.Fatalf("keys %v, want the type's suggestions then extras %v", got, want)
				}
			},
		},
		{
			name: "list workspace on a missing source is invalid",
			run: func(t *testing.T, c *database.Catalog) {
				birthOfAlice(t, c, userID)
				missing := make([]byte, 16)
				missing[15] = 9
				if _, err := ListWorkspace(c, sourceOwner(missing)); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := testWorld(t)
			tt.run(t, c)
		})
	}
}

func sourceOwner(id []byte) Owner { return Owner{Level: vocabkey.LevelSource, ID: id} }

// values reads the owner's stored value rows in a read transaction.
func values(t *testing.T, c *database.Catalog, sourceID []byte) []Row {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := listByOwners(tx, []Owner{sourceOwner(sourceID)})
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func runSet(c *database.Catalog, userID, sourceID []byte, in Input) (Row, error) {
	row, _, err := writes.Run(c, writes.Op{Action: "update_source_metadata", UserID: userID},
		func(tx *database.Tx) (Row, []rowchange.Change, error) {
			return Set(tx, userID, sourceOwner(sourceID), in)
		})
	return row, err
}

func runClear(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "update_source_metadata", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Clear(tx, userID, sourceOwner(sourceID), fieldID)
			return struct{}{}, changes, err
		})
	return err
}

func runDismissSuggestion(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "dismiss_source_metadata_suggestion", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := DismissSuggestion(tx, userID, sourceOwner(sourceID), fieldID)
			return struct{}{}, changes, err
		})
	return err
}

func runReorder(c *database.Catalog, userID, sourceID []byte, fieldIDs [][]byte) error {
	_, _, err := writes.Run(c, writes.Op{Action: "reorder_source_metadata", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := Reorder(tx, userID, sourceOwner(sourceID), fieldIDs)
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

func runCreateAll(c *database.Catalog, userID, sourceID []byte, d Draft) ([]rowchange.Change, error) {
	var changes []rowchange.Change
	_, _, err := writes.Run(c, writes.Op{Action: "copy_source", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			var err error
			changes, err = CreateAll(tx, userID, sourceOwner(sourceID), d)
			return struct{}{}, changes, err
		})
	return changes, err
}

func TestWorkspaceEntryOpenHref(t *testing.T) {
	entry := func(dataType, value string) WorkspaceEntry {
		return WorkspaceEntry{
			Field: metadatafields.Field{DataType: dataType},
			Value: &Row{ValueText: value},
		}
	}
	tests := []struct {
		name  string
		entry WorkspaceEntry
		want  string
	}{
		{name: "url field, host-shaped value", entry: entry(metadatafields.DataTypeURL, "www.example.org"), want: "https://www.example.org"},
		{name: "url field, https value", entry: entry(metadatafields.DataTypeURL, "https://example.org/records"), want: "https://example.org/records"},
		{name: "url field, damaged file value", entry: entry(metadatafields.DataTypeURL, "file:///System/Applications/Calculator.app"), want: ""},
		{name: "url field, damaged smb value", entry: entry(metadatafields.DataTypeURL, "smb://server.example.org/share"), want: ""},
		{name: "text field, host-shaped value", entry: entry(metadatafields.DataTypeText, "www.example.org"), want: ""},
		{name: "url field, suggested and unset", entry: WorkspaceEntry{Field: metadatafields.Field{DataType: metadatafields.DataTypeURL}, Suggested: true}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.entry.OpenHref(); got != tt.want {
				t.Fatalf("OpenHref() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateAll(t *testing.T) {
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "values with no layout rows write no layout",
			run: func(t *testing.T) {
				c, _, src, text, _, _, _ := applyWorld(t)
				changes, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: text.ID, ValueText: "No. 7"}}})
				if err != nil {
					t.Fatal(err)
				}
				if len(changes) != 1 || changes[0].EntityType != "source_metadata" {
					t.Fatalf("changes %+v", changes)
				}
				rows := values(t, c, src.ID)
				if len(rows) != 1 || rows[0].ValueText != "No. 7" {
					t.Fatalf("rows %+v", rows)
				}
				if n := countRows(t, c, "source_metadata_layout"); n != 0 {
					t.Fatalf("layout rows %d", n)
				}
			},
		},
		{
			name: "dismissed layout row on a valued field stays dismissed",
			run: func(t *testing.T) {
				c, _, src, text, link, _, _ := applyWorld(t)
				changes, err := runCreateAll(c, userID, src.ID, Draft{
					Values: []Input{{FieldID: text.ID, ValueText: "No. 7"}},
					Layout: []LayoutRow{{FieldID: link.ID, SortOrder: 0}, {FieldID: text.ID, SortOrder: 1, Dismissed: true}},
				})
				if err != nil {
					t.Fatal(err)
				}
				if len(changes) != 3 || changes[0].EntityType != "source_metadata" ||
					changes[1].EntityType != "source_metadata_layout" || changes[2].EntityType != "source_metadata_layout" {
					t.Fatalf("changes %+v", changes)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var dismissed int
				if err := db.QueryRow(`SELECT dismissed FROM source_metadata_layout WHERE source_id = ? AND field_id = ?`, src.ID, text.ID).Scan(&dismissed); err != nil || dismissed != 1 {
					t.Fatalf("dismissed %d err %v", dismissed, err)
				}
				// A layout row for a field with no value is allowed.
				if n := countRows(t, c, "source_metadata_layout"); n != 2 {
					t.Fatalf("layout rows %d", n)
				}
			},
		},
		{
			name: "URL value takes the form Set stores",
			run: func(t *testing.T) {
				c, _, src, _, link, _, _ := applyWorld(t)
				if _, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: link.ID, ValueText: "HTTPS://Example.org/a"}}}); err != nil {
					t.Fatal(err)
				}
				rows := values(t, c, src.ID)
				if len(rows) != 1 {
					t.Fatalf("rows %+v", rows)
				}
				if rows[0].ValueText != "https://example.org/a" {
					t.Fatalf("stored %q", rows[0].ValueText)
				}
			},
		},
		{
			name: "source that already has a value is refused",
			run: func(t *testing.T) {
				c, _, src, text, link, _, _ := applyWorld(t)
				if _, err := runSet(c, userID, src.ID, Input{FieldID: text.ID, ValueText: "x"}); err != nil {
					t.Fatal(err)
				}
				_, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: link.ID, ValueText: "https://example.org"}}})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "source that already has a layout row is refused",
			run: func(t *testing.T) {
				c, _, src, text, link, _, _ := applyWorld(t)
				if err := runDismissSuggestion(c, userID, src.ID, text.ID); err != nil {
					t.Fatal(err)
				}
				_, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: link.ID, ValueText: "https://example.org"}}})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "unknown field and empty text are refused",
			run: func(t *testing.T) {
				c, _, src, text, _, _, _ := applyWorld(t)
				missing := make([]byte, 16)
				missing[15] = 9
				if _, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: missing, ValueText: "x"}}}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("value field: %v", err)
				}
				if _, err := runCreateAll(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: missing}}}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("layout field: %v", err)
				}
				if _, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: text.ID}}}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty text: %v", err)
				}
				if n := countRows(t, c, "source_metadata") + countRows(t, c, "source_metadata_layout"); n != 0 {
					t.Fatalf("rows %d", n)
				}
			},
		},
		{
			name: "empty inputs write nothing",
			run: func(t *testing.T) {
				c, _, src, _, _, _, _ := applyWorld(t)
				changes, err := runCreateAll(c, userID, src.ID, Draft{})
				if err != nil || len(changes) != 0 {
					t.Fatalf("changes %+v err %v", changes, err)
				}
			},
		},
		{
			name: "ReadDraftTx returns layout rows by sort order",
			run: func(t *testing.T) {
				c, _, src, text, link, _, _ := applyWorld(t)
				if _, err := runCreateAll(c, userID, src.ID, Draft{
					Layout: []LayoutRow{{FieldID: text.ID, SortOrder: 1}, {FieldID: link.ID, SortOrder: 0, Dismissed: true}},
				}); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				tx, err := db.Begin()
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback()
				d, err := ReadDraftTx(tx, sourceOwner(src.ID))
				if err != nil {
					t.Fatal(err)
				}
				got := d.Layout
				if len(got) != 2 || !bytes.Equal(got[0].FieldID, link.ID) || !got[0].Dismissed ||
					!bytes.Equal(got[1].FieldID, text.ID) || got[1].Dismissed {
					t.Fatalf("layout %+v", got)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

// testWorld is a catalog with a user and a Source of a seeded type.
func testWorld(t *testing.T) (c *database.Catalog, userID []byte) {
	t.Helper()
	userID = []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Jake", r); err != nil {
		t.Fatal(err)
	}
	if err := sourcevocab.Install(c); err != nil {
		t.Fatal(err)
	}
	return c, userID
}

func newSource(t *testing.T, c *database.Catalog, userID, typeID []byte, title string) sources.Source {
	t.Helper()
	s, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: title})
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// birthOfAlice is a Source of the seeded birth certificate type.
func birthOfAlice(t *testing.T, c *database.Catalog, userID []byte) sources.Source {
	t.Helper()
	cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return newSource(t, c, userID, cert.ID, "Birth of Alice")
}

func TestStoresAreRegistered(t *testing.T) {
	tables := map[string]bool{}
	for _, tb := range catalogmodel.Tables {
		tables[tb.Name] = true
	}
	for level, st := range stores {
		if st.level != level {
			t.Fatalf("store at %q carries level %q", level, st.level)
		}
		for _, name := range []string{st.values, st.layout} {
			if !tables[name] {
				t.Fatalf("%s: %q is not in catalogmodel.Tables", level, name)
			}
			found := false
			for _, fk := range catalogmodel.FKs {
				if fk.From == name && fk.Column == st.ownerColumn {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s: catalogmodel.FKs has no FK from %s.%s", level, name, st.ownerColumn)
			}
		}
	}
}

func TestOwnerLevels(t *testing.T) {
	c, userID := testWorld(t)
	src := birthOfAlice(t, c, userID)
	doc, err := metadatafields.Lookup(c, "document_number", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	owners := map[string]Owner{"short id": {Level: vocabkey.LevelSource, ID: src.ID[:15]}}
	for _, l := range vocabkey.Levels() {
		if l != vocabkey.LevelSource {
			owners[string(l)] = Owner{Level: l, ID: src.ID}
		}
	}
	owners["empty level"] = Owner{Level: "", ID: src.ID}
	owners["page"] = Owner{Level: "page", ID: src.ID}

	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	write := func(f func(tx *database.Tx, o Owner) ([]rowchange.Change, error)) func(Owner) error {
		return func(o Owner) error {
			_, _, err := writes.Run(c, writes.Op{Action: "update_source_metadata", UserID: userID},
				func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					ch, err := f(tx, o)
					return struct{}{}, ch, err
				})
			return err
		}
	}
	ops := map[string]func(Owner) error{
		"Set": write(func(tx *database.Tx, o Owner) ([]rowchange.Change, error) {
			_, ch, err := Set(tx, userID, o, Input{FieldID: doc.ID, ValueText: "x"})
			return ch, err
		}),
		"Clear": write(func(tx *database.Tx, o Owner) ([]rowchange.Change, error) {
			return Clear(tx, userID, o, doc.ID)
		}),
		"DismissSuggestion": write(func(tx *database.Tx, o Owner) ([]rowchange.Change, error) {
			return DismissSuggestion(tx, userID, o, doc.ID)
		}),
		"Reorder": write(func(tx *database.Tx, o Owner) ([]rowchange.Change, error) {
			return Reorder(tx, userID, o, [][]byte{doc.ID})
		}),
		"CreateAll": write(func(tx *database.Tx, o Owner) ([]rowchange.Change, error) {
			return CreateAll(tx, userID, o, Draft{Values: []Input{{FieldID: doc.ID, ValueText: "x"}}})
		}),
		"Apply": write(func(tx *database.Tx, o Owner) ([]rowchange.Change, error) {
			return Apply(tx, userID, o, Draft{Values: []Input{{FieldID: doc.ID, ValueText: "x"}}})
		}),
		"ListByOwnersTx": func(o Owner) error {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()
			_, err = ListByOwnersTx(tx, []Owner{o})
			return err
		},
		"ListWorkspace": func(o Owner) error { _, err := ListWorkspace(c, o); return err },
		"ReadDraftTx": func(o Owner) error {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()
			_, err = ReadDraftTx(tx, o)
			return err
		},
	}
	for opName, op := range ops {
		for ownerName, o := range owners {
			if err := op(o); !errors.Is(err, ErrInvalid) {
				t.Fatalf("%s on %s owner = %v, want ErrInvalid", opName, ownerName, err)
			}
			if n := countRows(t, c, "source_metadata") + countRows(t, c, "source_metadata_layout"); n != 0 {
				t.Fatalf("%s on %s owner wrote %d rows", opName, ownerName, n)
			}
		}
	}
	good := sourceOwner(src.ID)
	if _, err := ReadDraftTx(nil, good); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ReadDraftTx(nil tx) = %v", err)
	}
	if _, err := ListByOwnersTx(nil, []Owner{good}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ListByOwnersTx(nil tx) = %v", err)
	}
}

func TestListWorkspaces(t *testing.T) {
	c, userID := testWorld(t)
	cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	other, _, err := writes.Run(c, writes.Op{Action: "create_source_type", UserID: userID},
		func(tx *database.Tx) (sourcetypes.Type, []rowchange.Change, error) {
			return sourcetypes.Create(tx, "Ledger", "", "")
		})
	if err != nil {
		t.Fatal(err)
	}
	f1, err := runFieldCreate(c, "Shelf mark", metadatafields.DataTypeText, "")
	if err != nil {
		t.Fatal(err)
	}
	f2, err := runFieldCreate(c, "Volume", metadatafields.DataTypeText, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []metadatafields.Field{f2, f1} {
		if err := sourcevocab.AppendSuggestion(c, other.ID, f.ID, vocabkey.LevelSource); err != nil {
			t.Fatal(err)
		}
	}
	a := newSource(t, c, userID, cert.ID, "Birth of Alice")
	b := newSource(t, c, userID, other.ID, "Ledger 1")
	doc, err := metadatafields.Lookup(c, "document_number", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runSet(c, userID, a.ID, Input{FieldID: doc.ID, ValueText: "No. 7"}); err != nil {
		t.Fatal(err)
	}

	// The catalog has one connection: read each Source alone before the batch
	// read transaction opens.
	alone := map[string][]WorkspaceEntry{}
	for _, src := range []sources.Source{a, b} {
		ws, err := ListWorkspace(c, sourceOwner(src.ID))
		if err != nil {
			t.Fatal(err)
		}
		alone[src.Title] = ws
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	got, err := listWorkspaces(tx, []Owner{sourceOwner(a.ID), sourceOwner(b.ID)})
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range []sources.Source{a, b} {
		if !reflect.DeepEqual(got[sourceOwner(src.ID).key()], alone[src.Title]) {
			t.Fatalf("%s: batch\n%+v\nalone\n%+v", src.Title, got[sourceOwner(src.ID).key()], alone[src.Title])
		}
	}
	// B has no layout: its entries follow the type's suggestion order.
	gotB := got[sourceOwner(b.ID).key()]
	if len(gotB) != 2 || !bytes.Equal(gotB[0].Field.ID, f2.ID) || !bytes.Equal(gotB[1].Field.ID, f1.ID) {
		t.Fatalf("B entries %+v", gotB)
	}
	// A has a layout row for document_number: it leads.
	gotA := got[sourceOwner(a.ID).key()]
	if len(gotA) == 0 || !bytes.Equal(gotA[0].Field.ID, doc.ID) || gotA[0].Value == nil {
		t.Fatalf("A entries %+v", gotA)
	}
}

// applyWorld is a catalog with a Source and four fields: doc and f1, f2 hold
// text and link holds a URL.
func applyWorld(t *testing.T) (c *database.Catalog, userID []byte, src sources.Source, doc, link, f1, f2 metadatafields.Field) {
	t.Helper()
	c, userID = testWorld(t)
	src = birthOfAlice(t, c, userID)
	var err error
	if doc, err = metadatafields.Lookup(c, "document_number", vocabkey.OriginProvenencia); err != nil {
		t.Fatal(err)
	}
	if link, err = runFieldCreate(c, "Web page", metadatafields.DataTypeURL, ""); err != nil {
		t.Fatal(err)
	}
	if f1, err = runFieldCreate(c, "Field one", metadatafields.DataTypeText, ""); err != nil {
		t.Fatal(err)
	}
	if f2, err = runFieldCreate(c, "Field two", metadatafields.DataTypeText, ""); err != nil {
		t.Fatal(err)
	}
	return c, userID, src, doc, link, f1, f2
}

func runApply(c *database.Catalog, userID, sourceID []byte, d Draft) ([]rowchange.Change, error) {
	var changes []rowchange.Change
	_, _, err := writes.Run(c, writes.Op{Action: "update_source_metadata", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			var err error
			changes, err = Apply(tx, userID, sourceOwner(sourceID), d)
			return struct{}{}, changes, err
		})
	return changes, err
}

func readDraft(t *testing.T, c *database.Catalog, sourceID []byte) Draft {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	d, err := ReadDraftTx(tx, sourceOwner(sourceID))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func changeShape(ch []rowchange.Change) []string {
	out := make([]string, len(ch))
	for i, c := range ch {
		out[i] = c.Action + " " + c.EntityType
	}
	return out
}

func fieldKeys(c rowchange.Change) []string {
	keys := make([]string, 0, len(c.Fields))
	for k := range c.Fields {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func TestApply(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "an empty owner gets what CreateAll writes",
			run: func(t *testing.T) {
				c, userID, src, doc, link, _, _ := applyWorld(t)
				cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				twin := newSource(t, c, userID, cert.ID, "Twin")
				draft := Draft{
					Values: []Input{{FieldID: doc.ID, ValueText: "No. 7"}},
					Layout: []LayoutRow{{FieldID: link.ID, SortOrder: 0}, {FieldID: doc.ID, SortOrder: 1, Dismissed: true}},
				}
				got, err := runApply(c, userID, src.ID, draft)
				if err != nil {
					t.Fatal(err)
				}
				want, err := runCreateAll(c, userID, twin.ID, draft)
				if err != nil {
					t.Fatal(err)
				}
				wantShape := []string{"create source_metadata", "create source_metadata_layout", "create source_metadata_layout"}
				if !slices.Equal(changeShape(got), wantShape) || !slices.Equal(changeShape(want), wantShape) {
					t.Fatalf("apply %v, create-all %v, want %v", changeShape(got), changeShape(want), wantShape)
				}
				if a, b := readDraft(t, c, src.ID), readDraft(t, c, twin.ID); !reflect.DeepEqual(a, b) || len(a.Layout) != 2 || !a.Layout[1].Dismissed {
					t.Fatalf("apply %+v, create-all %+v", a, b)
				}
			},
		},
		{
			name: "a value equal after trimming writes nothing",
			run: func(t *testing.T) {
				c, userID, src, doc, _, _, _ := applyWorld(t)
				if _, err := runSet(c, userID, src.ID, Input{FieldID: doc.ID, ValueText: "No. 7"}); err != nil {
					t.Fatal(err)
				}
				changes, err := runApply(c, userID, src.ID, Draft{Values: []Input{{FieldID: doc.ID, ValueText: "  No. 7  "}}})
				if err != nil || len(changes) != 0 {
					t.Fatalf("changes %+v err %v", changes, err)
				}
			},
		},
		{
			name: "a URL equal after canonicalization writes nothing",
			run: func(t *testing.T) {
				c, userID, src, _, link, _, _ := applyWorld(t)
				if _, err := runSet(c, userID, src.ID, Input{FieldID: link.ID, ValueText: "https://example.org/a"}); err != nil {
					t.Fatal(err)
				}
				changes, err := runApply(c, userID, src.ID, Draft{Values: []Input{{FieldID: link.ID, ValueText: "HTTPS://Example.org/a"}}})
				if err != nil || len(changes) != 0 {
					t.Fatalf("changes %+v err %v", changes, err)
				}
			},
		},
		{
			name: "a value left out is deleted, a changed one updated, its layout row stays",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, f2 := applyWorld(t)
				for _, in := range []Input{{FieldID: f1.ID, ValueText: "x"}, {FieldID: f2.ID, ValueText: "y"}} {
					if _, err := runSet(c, userID, src.ID, in); err != nil {
						t.Fatal(err)
					}
				}
				changes, err := runApply(c, userID, src.ID, Draft{Values: []Input{{FieldID: f1.ID, ValueText: "z"}}})
				if err != nil {
					t.Fatal(err)
				}
				if want := []string{"delete source_metadata", "update source_metadata"}; !slices.Equal(changeShape(changes), want) {
					t.Fatalf("changes %v, want %v", changeShape(changes), want)
				}
				if !slices.Equal(fieldKeys(changes[0]), []string{"field_id", "id", "source_id", "value_text"}) {
					t.Fatalf("delete fields %v", fieldKeys(changes[0]))
				}
				if !slices.Equal(fieldKeys(changes[1]), []string{"value_text"}) {
					t.Fatalf("update fields %v", fieldKeys(changes[1]))
				}
				d := readDraft(t, c, src.ID)
				if len(d.Values) != 1 || d.Values[0].ValueText != "z" || len(d.Layout) != 2 {
					t.Fatalf("draft %+v", d)
				}
			},
		},
		{
			name: "a reorder updates sort_order only",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, f2 := applyWorld(t)
				if _, err := runCreateAll(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: f1.ID, SortOrder: 0}, {FieldID: f2.ID, SortOrder: 1}}}); err != nil {
					t.Fatal(err)
				}
				changes, err := runApply(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: f2.ID, SortOrder: 0}, {FieldID: f1.ID, SortOrder: 1}}})
				if err != nil {
					t.Fatal(err)
				}
				if want := []string{"update source_metadata_layout", "update source_metadata_layout"}; !slices.Equal(changeShape(changes), want) {
					t.Fatalf("changes %v", changeShape(changes))
				}
				for _, ch := range changes {
					if !slices.Equal(fieldKeys(ch), []string{"field_id", "sort_order", "source_id"}) {
						t.Fatalf("fields %v", fieldKeys(ch))
					}
				}
				if d := changes[0].Fields["sort_order"]; d.Old != 1 || d.New != 0 {
					t.Fatalf("sort_order %+v", d)
				}
				if f := changes[0].Fields["field_id"]; f.Old != f.New {
					t.Fatalf("field_id %+v", f)
				}
			},
		},
		{
			name: "a dismissal updates dismissed only",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, _ := applyWorld(t)
				if _, err := runCreateAll(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: f1.ID, SortOrder: 0}}}); err != nil {
					t.Fatal(err)
				}
				changes, err := runApply(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: f1.ID, SortOrder: 0, Dismissed: true}}})
				if err != nil {
					t.Fatal(err)
				}
				if len(changes) != 1 || changes[0].Action != "update" || !slices.Equal(fieldKeys(changes[0]), []string{"dismissed", "field_id", "source_id"}) {
					t.Fatalf("changes %+v", changes)
				}
				if d := changes[0].Fields["dismissed"]; d.Old != false || d.New != true {
					t.Fatalf("dismissed %+v", d)
				}
			},
		},
		{
			name: "a layout row the draft leaves out stays",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, f2 := applyWorld(t)
				if _, err := runCreateAll(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: f1.ID, SortOrder: 0}, {FieldID: f2.ID, SortOrder: 1, Dismissed: true}}}); err != nil {
					t.Fatal(err)
				}
				changes, err := runApply(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: f1.ID, SortOrder: 0}}})
				if err != nil || len(changes) != 0 {
					t.Fatalf("changes %+v err %v", changes, err)
				}
				if d := readDraft(t, c, src.ID); len(d.Layout) != 2 || !d.Layout[1].Dismissed {
					t.Fatalf("draft %+v", d)
				}
			},
		},
		{
			name: "an empty draft clears every value and keeps the layout",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, f2 := applyWorld(t)
				for _, in := range []Input{{FieldID: f1.ID, ValueText: "x"}, {FieldID: f2.ID, ValueText: "y"}} {
					if _, err := runSet(c, userID, src.ID, in); err != nil {
						t.Fatal(err)
					}
				}
				changes, err := runApply(c, userID, src.ID, Draft{})
				if err != nil {
					t.Fatal(err)
				}
				if want := []string{"delete source_metadata", "delete source_metadata"}; !slices.Equal(changeShape(changes), want) {
					t.Fatalf("changes %v", changeShape(changes))
				}
				if d := readDraft(t, c, src.ID); len(d.Values) != 0 || len(d.Layout) != 2 {
					t.Fatalf("draft %+v", d)
				}
			},
		},
		{
			name: "ReadDraftTx on a missing source is an empty draft",
			run: func(t *testing.T) {
				c, _, _, _, _, _, _ := applyWorld(t)
				missing := make([]byte, 16)
				missing[15] = 9
				if d := readDraft(t, c, missing); len(d.Values) != 0 || len(d.Layout) != 0 {
					t.Fatalf("draft %+v", d)
				}
			},
		},
		{
			name: "dismissing a field with no layout row creates the row with an owner key",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, _ := applyWorld(t)
				var changes []rowchange.Change
				_, _, err := writes.Run(c, writes.Op{Action: "dismiss_source_metadata_suggestion", UserID: userID},
					func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
						var err error
						changes, err = DismissSuggestion(tx, userID, sourceOwner(src.ID), f1.ID)
						return struct{}{}, changes, err
					})
				if err != nil {
					t.Fatal(err)
				}
				if len(changes) != 1 || changes[0].Action != "create" {
					t.Fatalf("changes %+v", changes)
				}
				for _, k := range []string{"source_id", "field_id"} {
					if d := changes[0].Fields[k]; d.Old != nil || d.New == nil {
						t.Fatalf("%s %+v", k, d)
					}
				}
			},
		},
		{
			name: "refused drafts write nothing",
			run: func(t *testing.T) {
				c, userID, src, doc, _, f1, _ := applyWorld(t)
				missing := make([]byte, 16)
				missing[15] = 9
				rows := func() int {
					return countRows(t, c, "source_metadata") + countRows(t, c, "source_metadata_layout")
				}
				bad := map[string]struct {
					owner []byte
					d     Draft
				}{
					"repeated value field": {src.ID, Draft{Values: []Input{{FieldID: doc.ID, ValueText: "a"}, {FieldID: doc.ID, ValueText: "b"}}}},
					"repeated layout field": {src.ID, Draft{Values: []Input{{FieldID: f1.ID, ValueText: "a"}},
						Layout: []LayoutRow{{FieldID: doc.ID}, {FieldID: doc.ID, SortOrder: 1}}}},
					"blank value":          {src.ID, Draft{Values: []Input{{FieldID: f1.ID, ValueText: "a"}, {FieldID: doc.ID, ValueText: "   "}}}},
					"unknown value field":  {src.ID, Draft{Values: []Input{{FieldID: f1.ID, ValueText: "a"}, {FieldID: missing, ValueText: "b"}}}},
					"unknown layout field": {src.ID, Draft{Values: []Input{{FieldID: f1.ID, ValueText: "a"}}, Layout: []LayoutRow{{FieldID: missing}}}},
					"owner with no row":    {missing, Draft{Values: []Input{{FieldID: f1.ID, ValueText: "a"}}}},
				}
				for name, tc := range bad {
					if _, err := runApply(c, userID, tc.owner, tc.d); !errors.Is(err, ErrInvalid) {
						t.Fatalf("%s: got %v", name, err)
					}
					if n := rows(); n != 0 {
						t.Fatalf("%s wrote %d rows", name, n)
					}
				}
			},
		},
		{
			name: "CreateAll refuses an owner that holds a layout row",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, f2 := applyWorld(t)
				if _, err := runCreateAll(c, userID, src.ID, Draft{Layout: []LayoutRow{{FieldID: f1.ID}}}); err != nil {
					t.Fatal(err)
				}
				if _, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: f2.ID, ValueText: "x"}}}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func countRows(t *testing.T, c *database.Catalog, table string) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestListByOwnersTx(t *testing.T) {
	c, userID, a, _, _, f1, f2 := applyWorld(t)
	cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	b := newSource(t, c, userID, cert.ID, "B")
	empty := newSource(t, c, userID, cert.ID, "C")
	for _, set := range []struct {
		src sources.Source
		in  Input
	}{
		{a, Input{FieldID: f2.ID, ValueText: "a2"}},
		{a, Input{FieldID: f1.ID, ValueText: "a1"}},
		{b, Input{FieldID: f1.ID, ValueText: "b1"}},
	} {
		if _, err := runSet(c, userID, set.src.ID, set.in); err != nil {
			t.Fatal(err)
		}
	}
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	rows, err := ListByOwnersTx(tx, []Owner{sourceOwner(b.ID), sourceOwner(a.ID), sourceOwner(empty.ID), sourceOwner(a.ID)})
	if err != nil {
		t.Fatal(err)
	}
	first, second := a, b
	if bytes.Compare(a.ID, b.ID) > 0 {
		first, second = b, a
	}
	if len(rows) != 3 {
		t.Fatalf("rows %+v", rows)
	}
	if !bytes.Equal(rows[0].Owner.ID, first.ID) || rows[0].Owner.Level != vocabkey.LevelSource {
		t.Fatalf("first row owner %+v, want %x", rows[0].Owner, first.ID)
	}
	if !bytes.Equal(rows[2].Owner.ID, second.ID) {
		t.Fatalf("last row owner %+v, want %x", rows[2].Owner, second.ID)
	}
	if bytes.Equal(first.ID, a.ID) {
		if !bytes.Equal(rows[0].FieldID, f1.ID) || !bytes.Equal(rows[1].FieldID, f2.ID) || rows[0].ValueText != "a1" {
			t.Fatalf("A rows out of field-id order: %+v", rows[:2])
		}
	}

	if none, err := ListByOwnersTx(tx, nil); err != nil || none != nil {
		t.Fatalf("nil owners = %+v, %v", none, err)
	}
	if _, err := ListByOwnersTx(tx, []Owner{sourceOwner(a.ID), {Level: vocabkey.LevelGraph, ID: a.ID}}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("graph owner = %v", err)
	}

	listByOwnersBatches(t)
}

// listByOwnersBatches is TestListByOwnersTx's cases over more owners than one
// sqlbatch IN list holds: the rows still come back by owner id then field id,
// and the workspace read finds every owner's type, values and layout.
func listByOwnersBatches(t *testing.T) {
	t.Helper()
	c, userID, _, _, _, f1, f2 := applyWorld(t)
	cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	n := sqlbatch.Size*2 + 201
	owners := make([]Owner, 0, n)
	_, _, err = writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		var changes []rowchange.Change
		for i := 0; i < n; i++ {
			src, ch, err := sources.Create(tx, userID, sources.CreateInput{SourceTypeID: cert.ID, Title: "S"})
			if err != nil {
				return struct{}{}, nil, err
			}
			changes = append(changes, ch...)
			o := sourceOwner(src.ID)
			for _, f := range []metadatafields.Field{f2, f1} {
				_, ch, err := Set(tx, userID, o, Input{FieldID: f.ID, ValueText: "v"})
				if err != nil {
					return struct{}{}, nil, err
				}
				changes = append(changes, ch...)
			}
			owners = append(owners, o)
		}
		return struct{}{}, changes, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	rand.New(rand.NewSource(1)).Shuffle(len(owners), func(i, j int) { owners[i], owners[j] = owners[j], owners[i] })
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	t.Run("order holds across batches", func(t *testing.T) {
		rows, err := ListByOwnersTx(tx, owners)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != 2*n {
			t.Fatalf("got %d rows, want %d", len(rows), 2*n)
		}
		for i := 1; i < len(rows); i++ {
			prev, cur := rows[i-1], rows[i]
			byOwner := bytes.Compare(prev.Owner.ID, cur.Owner.ID)
			if byOwner > 0 || (byOwner == 0 && bytes.Compare(prev.FieldID, cur.FieldID) >= 0) {
				t.Fatalf("row %d out of order: %x/%x then %x/%x", i, prev.Owner.ID, prev.FieldID, cur.Owner.ID, cur.FieldID)
			}
		}
	})
	t.Run("workspaces read every owner", func(t *testing.T) {
		got, err := listWorkspaces(tx, owners)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != n {
			t.Fatalf("got %d owners, want %d", len(got), n)
		}
		for _, o := range owners {
			if len(got[o.key()]) == 0 {
				t.Fatalf("owner %x has no entries", o.ID)
			}
		}
	})
}
