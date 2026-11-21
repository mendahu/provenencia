package sourcemetadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/Jake-Robins-Studios/provenencia/core/apperr"
	"github.com/Jake-Robins-Studios/provenencia/core/citationformat"
	"github.com/Jake-Robins-Studios/provenencia/core/database"
	"github.com/Jake-Robins-Studios/provenencia/core/database/artifacts"
	"github.com/Jake-Robins-Studios/provenencia/core/database/artifacts/artifactstest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/catalogmodel"
	"github.com/Jake-Robins-Studios/provenencia/core/database/dbtest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/metadatafields"
	"github.com/Jake-Robins-Studios/provenencia/core/database/metadatafields/metadatafieldstest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/rowchange"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sources"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcetypes"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcetypes/sourcetypestest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcevocab"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sourcevocab/sourcevocabtest"
	"github.com/Jake-Robins-Studios/provenencia/core/database/sqlbatch"
	"github.com/Jake-Robins-Studios/provenencia/core/database/users"
	"github.com/Jake-Robins-Studios/provenencia/core/ref"
	"github.com/Jake-Robins-Studios/provenencia/core/valuemark"
	"github.com/Jake-Robins-Studios/provenencia/core/vocabkey"
	"github.com/Jake-Robins-Studios/provenencia/core/writes"
	"github.com/google/uuid"
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
				if latestAction(t, c) != "update_metadata" {
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
				if latestAction(t, c) != "update_metadata" {
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
				extraID, err := metadatafieldstest.Upsert(c, metadatafields.Field{
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
				fieldID, err := metadatafieldstest.Upsert(c, metadatafields.Field{
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
				if latestAction(t, c) != "dismiss_metadata_suggestion" {
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
				if latestAction(t, c) != "update_metadata" {
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
				if latestAction(t, c) != "reorder_metadata" {
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
				extraID, err := metadatafieldstest.Upsert(c, metadatafields.Field{
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
				link, err := runFieldCreate(c, "Web page", metadatafields.DataTypeURL, "")
				if err != nil {
					t.Fatal(err)
				}
				_, err = runSet(c, userID, src.ID, Input{FieldID: link.ID, ValueText: "not a link"})
				if !errors.Is(err, ErrInvalid) || !reflect.DeepEqual(errParams(t, err), []string{database.UUIDString(link.ID)}) {
					t.Fatalf("bad url %v %v", err, errParams(t, err))
				}
				_, err = runSet(c, userID, src.ID, Input{FieldID: missing, ValueText: "x"})
				if !errors.Is(err, ErrInvalid) || len(errParams(t, err)) != 0 {
					t.Fatalf("missing field %v", err)
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
	row, _, err := runSetOwner(c, userID, sourceOwner(sourceID), in)
	return row, err
}

func runClear(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	_, err := runClearOwner(c, userID, sourceOwner(sourceID), fieldID)
	return err
}

// runSetOwner runs Set on any owner in its own transaction.
func runSetOwner(c *database.Catalog, userID []byte, owner Owner, in Input) (Row, []rowchange.Change, error) {
	var changes []rowchange.Change
	row, _, err := writes.Run(c, writes.Op{Action: "update_metadata", UserID: userID},
		func(tx *database.Tx) (Row, []rowchange.Change, error) {
			r, ch, err := Set(tx, userID, owner, in)
			changes = ch
			return r, ch, err
		})
	return row, changes, err
}

// runClearOwner runs Clear on any owner in its own transaction.
func runClearOwner(c *database.Catalog, userID []byte, owner Owner, fieldID []byte) ([]rowchange.Change, error) {
	var changes []rowchange.Change
	_, err := writes.RunChanges(c, writes.Op{Action: "update_metadata", UserID: userID},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			var err error
			changes, err = Clear(tx, userID, owner, fieldID)
			return changes, err
		})
	return changes, err
}

func runDismissSuggestion(c *database.Catalog, userID, sourceID, fieldID []byte) error {
	_, err := writes.RunChanges(c, writes.Op{Action: "dismiss_metadata_suggestion", UserID: userID},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			return DismissSuggestion(tx, userID, sourceOwner(sourceID), fieldID)
		})
	return err
}

func runReorder(c *database.Catalog, userID, sourceID []byte, fieldIDs [][]byte) error {
	_, err := writes.RunChanges(c, writes.Op{Action: "reorder_metadata", UserID: userID},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			return Reorder(tx, userID, sourceOwner(sourceID), fieldIDs)
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
	_, err := writes.RunChanges(c, writes.Op{Action: "copy_source", UserID: userID},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			var err error
			changes, err = CreateAll(tx, userID, sourceOwner(sourceID), d)
			return changes, err
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
		{
			name: "CreateAll stores a supplied value",
			run: func(t *testing.T) {
				c, userID, src, text, _, _, _ := applyWorld(t)
				if _, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: text.ID, ValueText: "Topeka", Mark: valuemark.Mark{Supplied: true}}}}); err != nil {
					t.Fatal(err)
				}
				rows := values(t, c, src.ID)
				if len(rows) != 1 || rows[0].ValueText != "Topeka" || rows[0].Mark != (valuemark.Mark{Supplied: true}) {
					t.Fatalf("rows %+v", rows)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

// seedBirthCertificate gives the catalog its own small vocabulary, not the
// product seed: a provenencia birth certificate type suggesting three Source
// fields, which these tests are about.
func seedBirthCertificate(t *testing.T, c *database.Catalog) {
	t.Helper()
	typeID, err := sourcetypestest.Upsert(c, sourcetypes.Type{
		Key: "birth_certificate", Origin: vocabkey.OriginProvenencia, Label: "Birth certificate", IconKey: "type_certificate",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []struct{ key, label string }{
		{"document_number", "Document number"},
		{"record_date", "Record date"},
		{"issue_date", "Issue date"},
	} {
		fieldID, err := metadatafieldstest.Upsert(c, metadatafields.Field{
			Key: f.key, Origin: vocabkey.OriginProvenencia, Label: f.label, DataType: metadatafields.DataTypeText,
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := sourcevocab.AppendSuggestion(c, typeID, fieldID, vocabkey.LevelSource); err != nil {
			t.Fatal(err)
		}
	}
}

// testWorld is a catalog with a user and a Source of a birth certificate type.
func testWorld(t *testing.T) (c *database.Catalog, userID []byte) {
	t.Helper()
	c, userID = emptyWorld(t)
	seedBirthCertificate(t, c)
	return c, userID
}

// seededWorld is a catalog with a user and the product seed (sourcevocabtest.Install).
// It never calls seedBirthCertificate, whose format-less type the seed's would collide with.
func seededWorld(t *testing.T) (c *database.Catalog, userID []byte) {
	t.Helper()
	c, userID = emptyWorld(t)
	if err := sourcevocabtest.Install(c); err != nil {
		t.Fatal(err)
	}
	return c, userID
}

// emptyWorld is a catalog with a user and no vocabulary.
func emptyWorld(t *testing.T) (c *database.Catalog, userID []byte) {
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

// birthOfAlice is a Source of testWorld's birth certificate type.
func birthOfAlice(t *testing.T, c *database.Catalog, userID []byte) sources.Source {
	t.Helper()
	cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return newSource(t, c, userID, cert.ID, "Birth of Alice")
}

func TestStoresAreRegistered(t *testing.T) {
	var want, got []string
	for _, l := range vocabkey.MetadataLevels() {
		want = append(want, string(l))
	}
	for l := range stores {
		got = append(got, string(l))
	}
	sort.Strings(want)
	sort.Strings(got)
	if !slices.Equal(got, want) {
		t.Fatalf("store levels = %v, want vocabkey.MetadataLevels %v", got, want)
	}
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
	owners := map[string]Owner{
		"short id":          {Level: vocabkey.LevelSource, ID: src.ID[:15]},
		"short artifact id": {Level: vocabkey.LevelArtifact, ID: src.ID[:15]},
		"short graph id":    {Level: vocabkey.LevelGraph, ID: src.ID[:15]},
		"citation":          {Level: vocabkey.LevelCitation, ID: src.ID},
	}
	owners["empty level"] = Owner{Level: "", ID: src.ID}
	owners["page"] = Owner{Level: "page", ID: src.ID}

	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	write := func(f func(tx *database.Tx, o Owner) ([]rowchange.Change, error)) func(Owner) error {
		return func(o Owner) error {
			_, err := writes.RunChanges(c, writes.Op{Action: "update_metadata", UserID: userID},
				func(tx *database.Tx) ([]rowchange.Change, error) {
					return f(tx, o)
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
			n := 0
			for _, st := range stores {
				n += countRows(t, c, st.values) + countRows(t, c, st.layout)
			}
			if n != 0 {
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
			return sourcetypes.Create(tx, sourcetypes.CreateInput{Label: "Ledger"})
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
	if _, err := runSet(c, userID, a.ID, Input{FieldID: doc.ID, ValueText: "No. 7", Mark: valuemark.Mark{Supplied: true}}); err != nil {
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
		if !reflect.DeepEqual(got[sourceOwner(src.ID).key()].Entries, alone[src.Title]) {
			t.Fatalf("%s: batch\n%+v\nalone\n%+v", src.Title, got[sourceOwner(src.ID).key()].Entries, alone[src.Title])
		}
	}
	// B has no layout: its entries follow the type's suggestion order.
	gotB := got[sourceOwner(b.ID).key()].Entries
	if len(gotB) != 2 || !bytes.Equal(gotB[0].Field.ID, f2.ID) || !bytes.Equal(gotB[1].Field.ID, f1.ID) {
		t.Fatalf("B entries %+v", gotB)
	}
	// A has a layout row for document_number: it leads.
	gotA := got[sourceOwner(a.ID).key()].Entries
	if len(gotA) == 0 || !bytes.Equal(gotA[0].Field.ID, doc.ID) || gotA[0].Value == nil {
		t.Fatalf("A entries %+v", gotA)
	}
	if gotA[0].Value.Mark != (valuemark.Mark{Supplied: true}) {
		t.Fatalf("A's value mark %+v, want supplied", gotA[0].Value.Mark)
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
	_, err := writes.RunChanges(c, writes.Op{Action: "update_metadata", UserID: userID},
		func(tx *database.Tx) ([]rowchange.Change, error) {
			var err error
			changes, err = Apply(tx, userID, sourceOwner(sourceID), d)
			return changes, err
		})
	return changes, err
}

// readOwnerDraft is readDraft for any owner.
func readOwnerDraft(t *testing.T, c *database.Catalog, o Owner) Draft {
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
	d, err := ReadDraftTx(tx, o)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func readDraft(t *testing.T, c *database.Catalog, sourceID []byte) Draft {
	t.Helper()
	return readOwnerDraft(t, c, sourceOwner(sourceID))
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
				if !slices.Equal(fieldKeys(changes[0]), []string{"field_id", "id", "is_supplied", "is_unknown", "source_id", "value_text"}) {
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
				_, err := writes.RunChanges(c, writes.Op{Action: "dismiss_metadata_suggestion", UserID: userID},
					func(tx *database.Tx) ([]rowchange.Change, error) {
						var err error
						changes, err = DismissSuggestion(tx, userID, sourceOwner(src.ID), f1.ID)
						return changes, err
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
					_, err := runApply(c, userID, tc.owner, tc.d)
					if !errors.Is(err, ErrInvalid) {
						t.Fatalf("%s: got %v", name, err)
					}
					if name != "blank value" && len(errParams(t, err)) != 0 {
						t.Fatalf("%s: params %v", name, errParams(t, err))
					}
					if n := rows(); n != 0 {
						t.Fatalf("%s wrote %d rows", name, n)
					}
				}
			},
		},
		{
			name: "a refused value names its field",
			run: func(t *testing.T) {
				c, userID, src, doc, link, f1, _ := applyWorld(t)
				rows := func() int {
					return countRows(t, c, "source_metadata") + countRows(t, c, "source_metadata_layout")
				}
				want := func(f metadatafields.Field) []string { return []string{database.UUIDString(f.ID)} }
				cases := []struct {
					name string
					run  func() error
					want []string
				}{
					{"blank text", func() error {
						_, err := runApply(c, userID, src.ID, Draft{Values: []Input{{FieldID: f1.ID, ValueText: "a"}, {FieldID: doc.ID, ValueText: "   "}}})
						return err
					}, want(doc)},
					{"bad url", func() error {
						_, err := runApply(c, userID, src.ID, Draft{Values: []Input{{FieldID: link.ID, ValueText: "not a link"}}})
						return err
					}, want(link)},
					{"first refused in draft order", func() error {
						_, err := runApply(c, userID, src.ID, Draft{Values: []Input{{FieldID: link.ID, ValueText: "x"}, {FieldID: doc.ID, ValueText: " "}}})
						return err
					}, want(link)},
					{"CreateAll", func() error {
						_, err := runCreateAll(c, userID, src.ID, Draft{Values: []Input{{FieldID: link.ID, ValueText: "x"}}})
						return err
					}, want(link)},
				}
				for _, tc := range cases {
					err := tc.run()
					if !errors.Is(err, ErrInvalid) {
						t.Fatalf("%s: got %v", tc.name, err)
					}
					if got := errParams(t, err); !reflect.DeepEqual(got, tc.want) {
						t.Fatalf("%s: params %v, want %v", tc.name, got, tc.want)
					}
					if n := rows(); n != 0 {
						t.Fatalf("%s wrote %d rows", tc.name, n)
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
		{
			name: "a draft carries marks, and a mark-only edit is one update",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, f2 := applyWorld(t)
				o := newLevelOwner(t, c, userID, src, vocabkey.LevelArtifact)
				apply := func(d Draft) ([]rowchange.Change, error) {
					var changes []rowchange.Change
					_, err := writes.RunChanges(c, writes.Op{Action: "update_metadata", UserID: userID},
						func(tx *database.Tx) ([]rowchange.Change, error) {
							var err error
							changes, err = Apply(tx, userID, o, d)
							return changes, err
						})
					return changes, err
				}
				d := Draft{Values: []Input{
					{FieldID: f1.ID, Mark: valuemark.Mark{Unknown: true}},
					{FieldID: f2.ID, ValueText: "https://x.org/1", Mark: valuemark.Mark{Supplied: true}},
				}}
				if _, err := apply(d); err != nil {
					t.Fatal(err)
				}
				got := readOwnerDraft(t, c, o)
				if len(got.Values) != 2 || got.Values[0].Mark != d.Values[0].Mark || got.Values[1].Mark != d.Values[1].Mark ||
					got.Values[0].ValueText != "" || got.Values[1].ValueText != "https://x.org/1" {
					t.Fatalf("draft %+v", got.Values)
				}
				// Only the mark of the second value changes.
				d.Values[1].Mark = valuemark.Mark{}
				changes, err := apply(d)
				if err != nil {
					t.Fatal(err)
				}
				if len(changes) != 1 || changes[0].Action != rowchange.ActionUpdate || !slices.Equal(fieldKeys(changes[0]), []string{"is_supplied"}) {
					t.Fatalf("changes %+v", changes)
				}
			},
		},
		{
			name: "a draft with a refused mark names its field and writes nothing",
			run: func(t *testing.T) {
				c, userID, src, _, _, f1, f2 := applyWorld(t)
				_, err := runApply(c, userID, src.ID, Draft{Values: []Input{
					{FieldID: f1.ID, ValueText: "ok"},
					{FieldID: f2.ID, ValueText: "text", Mark: valuemark.Mark{Unknown: true}},
				}})
				if !errors.Is(err, valuemark.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				if got := errParams(t, err); !slices.Equal(got, []string{database.UUIDString(f2.ID)}) {
					t.Fatalf("params %v", got)
				}
				if n := countRows(t, c, "source_metadata"); n != 0 {
					t.Fatalf("wrote %d rows", n)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

// errParams is the params of the *apperr.Error in err's chain.
func errParams(t *testing.T, err error) []string {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("not an apperr: %v", err)
	}
	return ae.Params()
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
	if _, err := ListByOwnersTx(tx, []Owner{sourceOwner(a.ID), {Level: vocabkey.LevelCitation, ID: a.ID}}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("citation owner = %v", err)
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
	_, err = writes.RunChanges(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) ([]rowchange.Change, error) {
		var changes []rowchange.Change
		for i := 0; i < n; i++ {
			src, ch, err := sources.Create(tx, userID, sources.CreateInput{SourceTypeID: cert.ID, Title: "S"})
			if err != nil {
				return nil, err
			}
			changes = append(changes, ch...)
			o := sourceOwner(src.ID)
			for _, f := range []metadatafields.Field{f2, f1} {
				_, ch, err := Set(tx, userID, o, Input{FieldID: f.ID, ValueText: "v"})
				if err != nil {
					return nil, err
				}
				changes = append(changes, ch...)
			}
			owners = append(owners, o)
		}
		return changes, nil
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
			if len(got[o.key()].Entries) == 0 {
				t.Fatalf("owner %x has no entries", o.ID)
			}
		}
	})
}

// newLevelOwner makes one more owner at the level on the Source: an Artifact
// (linked to its graphs) or a bare graph inserted by SQL, since a graph made
// through evidencegraphs.Create needs an Artifact and keeps its last one.
func newLevelOwner(t *testing.T, c *database.Catalog, userID []byte, src sources.Source, level vocabkey.Level) Owner {
	t.Helper()
	switch level {
	case vocabkey.LevelArtifact:
		a, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Page"})
		if err != nil {
			t.Fatal(err)
		}
		return Owner{Level: level, ID: a.ID}
	case vocabkey.LevelGraph:
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		id, err := uuid.NewV7()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO evidence_graphs (id, source_id, name) VALUES (?, ?, ?)`, id[:], src.ID, "Household"); err != nil {
			t.Fatal(err)
		}
		return Owner{Level: level, ID: id[:]}
	}
	t.Fatalf("no fixture for level %s", level)
	return Owner{}
}

// TestOwnerTables runs every operation on an Artifact owner and a graph owner:
// each writes its own level's tables and nothing else.
func TestOwnerTables(t *testing.T) {
	tables := map[vocabkey.Level]struct{ values, layout, column string }{
		vocabkey.LevelArtifact: {"artifact_metadata", "artifact_metadata_layout", "artifact_id"},
		vocabkey.LevelGraph:    {"evidence_graph_metadata", "evidence_graph_metadata_layout", "graph_id"},
	}
	for _, level := range []vocabkey.Level{vocabkey.LevelArtifact, vocabkey.LevelGraph} {
		t.Run(string(level), func(t *testing.T) {
			tb := tables[level]
			c, userID, src, doc, _, f1, f2 := applyWorld(t)
			cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			if err := sourcevocab.AppendSuggestion(c, cert.ID, doc.ID, level); err != nil {
				t.Fatal(err)
			}
			// A field suggested at each other level must not reach this owner.
			for _, other := range []vocabkey.Level{vocabkey.LevelSource, vocabkey.LevelArtifact, vocabkey.LevelGraph} {
				if other == level {
					continue
				}
				if err := sourcevocab.AppendSuggestion(c, cert.ID, f2.ID, other); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Anchor"}); err != nil {
				t.Fatal(err)
			}
			o := newLevelOwner(t, c, userID, src, level)

			// Only this level's suggestion, not the Source level's.
			ws, err := ListWorkspace(c, o)
			if err != nil {
				t.Fatal(err)
			}
			if len(ws) != 1 || !bytes.Equal(ws[0].Field.ID, doc.ID) || !ws[0].Suggested || ws[0].Value != nil {
				t.Fatalf("workspace = %+v, want document_number suggested only", ws)
			}

			run := func(f func(tx *database.Tx) ([]rowchange.Change, error)) []rowchange.Change {
				t.Helper()
				var changes []rowchange.Change
				_, err := writes.RunChanges(c, writes.Op{Action: "update_metadata", UserID: userID},
					func(tx *database.Tx) ([]rowchange.Change, error) {
						var err error
						changes, err = f(tx)
						return changes, err
					})
				if err != nil {
					t.Fatal(err)
				}
				return changes
			}
			ownerIs := func(ch rowchange.Change) bool {
				return ch.Fields[tb.column].New == database.UUIDString(o.ID) || ch.Fields[tb.column].Old == database.UUIDString(o.ID)
			}

			changes := run(func(tx *database.Tx) ([]rowchange.Change, error) {
				_, ch, err := Set(tx, userID, o, Input{FieldID: doc.ID, ValueText: "No. 7"})
				return ch, err
			})
			if len(changes) != 1 || changes[0].EntityType != tb.values || !ownerIs(changes[0]) {
				t.Fatalf("Set changes = %+v, want one %s create carrying %s", changes, tb.values, tb.column)
			}
			if n := countRows(t, c, tb.values); n != 1 {
				t.Fatalf("%s rows = %d", tb.values, n)
			}
			if n := countRows(t, c, "source_metadata") + countRows(t, c, "source_metadata_layout"); n != 0 {
				t.Fatalf("Source tables hold %d rows", n)
			}

			run(func(tx *database.Tx) ([]rowchange.Change, error) { return Clear(tx, userID, o, doc.ID) })
			if n := countRows(t, c, tb.values); n != 0 {
				t.Fatalf("Clear left %d values", n)
			}
			changes = run(func(tx *database.Tx) ([]rowchange.Change, error) { return DismissSuggestion(tx, userID, o, doc.ID) })
			if len(changes) != 1 || changes[0].EntityType != tb.layout || !ownerIs(changes[0]) {
				t.Fatalf("DismissSuggestion changes = %+v, want one %s change carrying %s", changes, tb.layout, tb.column)
			}
			run(func(tx *database.Tx) ([]rowchange.Change, error) {
				return Reorder(tx, userID, o, [][]byte{f2.ID, f1.ID, doc.ID})
			})
			draft := Draft{
				Values: []Input{{FieldID: f1.ID, ValueText: "a"}},
				Layout: []LayoutRow{{FieldID: f1.ID, SortOrder: 0}, {FieldID: doc.ID, SortOrder: 1, Dismissed: true}},
			}
			run(func(tx *database.Tx) ([]rowchange.Change, error) { return Apply(tx, userID, o, draft) })

			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			got, err := ReadDraftTx(tx, o)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Values) != 1 || !bytes.Equal(got.Values[0].FieldID, f1.ID) || got.Values[0].ValueText != "a" {
				t.Fatalf("draft values = %+v", got.Values)
			}
			rows, err := ListByOwnersTx(tx, []Owner{o})
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].Owner.Level != level || !bytes.Equal(rows[0].Owner.ID, o.ID) {
				t.Fatalf("ListByOwnersTx = %+v", rows)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}

			// CreateAll on a fresh owner of the level.
			o2 := newLevelOwner(t, c, userID, src, level)
			run(func(tx *database.Tx) ([]rowchange.Change, error) { return CreateAll(tx, userID, o2, draft) })
			if n := countRows(t, c, tb.values); n != 2 {
				t.Fatalf("%s rows after CreateAll = %d", tb.values, n)
			}

			// A missing owner is refused.
			missing := Owner{Level: level, ID: bytes.Repeat([]byte{9}, 16)}
			_, _, err = writes.Run(c, writes.Op{Action: "update_metadata", UserID: userID},
				func(tx *database.Tx) (Row, []rowchange.Change, error) {
					return Set(tx, userID, missing, Input{FieldID: doc.ID, ValueText: "x"})
				})
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("Set on missing owner = %v, want ErrInvalid", err)
			}
		})
	}
}

// TestListWorkspacesExported reads many owners at once: the lists are
// index-aligned, a repeated owner gets the same list, and each equals the
// owner read alone.
func TestListWorkspacesExported(t *testing.T) {
	c, userID, src, doc, _, _, _ := applyWorld(t)
	cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if err := sourcevocab.AppendSuggestion(c, cert.ID, doc.ID, vocabkey.LevelArtifact); err != nil {
		t.Fatal(err)
	}
	s := sourceOwner(src.ID)
	a1 := newLevelOwner(t, c, userID, src, vocabkey.LevelArtifact)
	a2 := newLevelOwner(t, c, userID, src, vocabkey.LevelArtifact)
	if _, _, err := writes.Run(c, writes.Op{Action: "update_metadata", UserID: userID}, func(tx *database.Tx) (Row, []rowchange.Change, error) {
		return Set(tx, userID, a1, Input{FieldID: doc.ID, ValueText: "No. 7"})
	}); err != nil {
		t.Fatal(err)
	}

	owners := []Owner{s, a1, a2, a1}
	got, err := ListWorkspaces(c, owners)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(owners) {
		t.Fatalf("lists = %d, want %d", len(got), len(owners))
	}
	for i, o := range owners {
		alone, err := ListWorkspace(c, o)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got[i].Entries, alone) {
			t.Errorf("list %d differs from the owner read alone", i)
		}
	}
	if !reflect.DeepEqual(got[1].Entries, got[3].Entries) {
		t.Error("a repeated owner got different lists")
	}
	if len(got[1].Entries) != 1 || got[1].Entries[0].Value == nil || got[1].Entries[0].Value.ValueText != "No. 7" {
		t.Errorf("A1's list = %+v", got[1].Entries)
	}
	if len(got[2].Entries) != 1 || got[2].Entries[0].Value != nil || !got[2].Entries[0].Suggested {
		t.Errorf("A2's list = %+v, want the suggestion only", got[2].Entries)
	}

	if none, err := ListWorkspaces(c, nil); err != nil || none != nil {
		t.Errorf("no owners = %v, %v; want nil", none, err)
	}
	missing := Owner{Level: vocabkey.LevelGraph, ID: bytes.Repeat([]byte{9}, 16)}
	if _, err := ListWorkspaces(c, []Owner{s, missing}); !errors.Is(err, ErrInvalid) {
		t.Errorf("missing graph owner = %v, want ErrInvalid", err)
	}
}

// TestReadDraftsTx reads many owners' drafts in one pass, index-aligned.
func TestReadDraftsTx(t *testing.T) {
	c, userID, src, doc, _, f1, _ := applyWorld(t)
	a1 := newLevelOwner(t, c, userID, src, vocabkey.LevelArtifact)
	g1 := newLevelOwner(t, c, userID, src, vocabkey.LevelGraph)
	if _, err := writes.RunChanges(c, writes.Op{Action: "update_metadata", UserID: userID}, func(tx *database.Tx) ([]rowchange.Change, error) {
		_, ch, err := Set(tx, userID, a1, Input{FieldID: doc.ID, ValueText: "No. 7", Mark: valuemark.Mark{Supplied: true}})
		if err != nil {
			return nil, err
		}
		_, more, err := Set(tx, userID, g1, Input{FieldID: f1.ID, Mark: valuemark.Mark{Unknown: true}})
		return append(ch, more...), err
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

	got, err := ReadDraftsTx(tx, []Owner{a1, g1, a1})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("drafts = %d", len(got))
	}
	for i, o := range []Owner{a1, g1, a1} {
		alone, err := ReadDraftTx(tx, o)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got[i], alone) {
			t.Errorf("draft %d = %+v, alone %+v", i, got[i], alone)
		}
	}
	if len(got[0].Values) != 1 || !bytes.Equal(got[0].Values[0].FieldID, doc.ID) ||
		len(got[1].Values) != 1 || !bytes.Equal(got[1].Values[0].FieldID, f1.ID) {
		t.Errorf("drafts = %+v", got)
	}
	// Marks travel into the draft.
	if got[0].Values[0].Mark != (valuemark.Mark{Supplied: true}) || got[0].Values[0].ValueText != "No. 7" ||
		got[1].Values[0].Mark != (valuemark.Mark{Unknown: true}) || got[1].Values[0].ValueText != "" {
		t.Errorf("marks in drafts = %+v", got)
	}
	if none, err := ReadDraftsTx(tx, nil); err != nil || none != nil {
		t.Errorf("no owners = %v, %v; want nil", none, err)
	}
	if _, err := ReadDraftsTx(nil, []Owner{a1}); !errors.Is(err, ErrInvalid) {
		t.Errorf("nil tx = %v, want ErrInvalid", err)
	}
	if _, err := ReadDraftsTx(tx, []Owner{a1, {Level: vocabkey.LevelCitation, ID: a1.ID}}); !errors.Is(err, ErrInvalid) {
		t.Errorf("citation owner = %v, want ErrInvalid", err)
	}
}

// draftBoardFile is testdata/draft_board_scenarios.json, shared with the Swift
// MetadataDraftHostParityTests: the rules the client's draft board mirrors.
type draftBoardFile struct {
	Suggestions []string        `json:"suggestions"`
	Needed      []string        `json:"needed"`
	Extras      []string        `json:"extras"`
	Scenarios   []draftScenario `json:"scenarios"`
}

type draftScenario struct {
	Name            string        `json:"name"`
	KnownDivergence string        `json:"known_divergence"`
	Seed            []draftAction `json:"seed"`
	Start           draftBoard    `json:"start"`
	Actions         []draftAction `json:"actions"`
	Board           draftBoard    `json:"board"`
	Draft           *struct {
		Values []struct {
			Field string `json:"field"`
			Value string `json:"value"`
		} `json:"values"`
		Layout []struct {
			Field     string `json:"field"`
			SortOrder int    `json:"sortOrder"`
			Dismissed bool   `json:"dismissed"`
		} `json:"layout"`
	} `json:"draft"`
}

type draftAction struct {
	Op     string  `json:"op"`
	Field  string  `json:"field"`
	Value  string  `json:"value"`
	Before *string `json:"before"`
	// Refused is an action the engine refuses with sourcemetadata.invalid.
	Refused bool `json:"refused"`
}

// draftBoard is a board as it is drawn: the saved rows in order, then the empty suggestions.
type draftBoard struct {
	Saved []draftSaved `json:"saved"`
	Empty []string     `json:"empty"`
}

type draftSaved struct {
	Field     string `json:"field"`
	Value     string `json:"value"`
	Suggested bool   `json:"suggested"`
}

func TestDraftBoardScenarios(t *testing.T) {
	raw, err := os.ReadFile("testdata/draft_board_scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	var file draftBoardFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	if len(file.Scenarios) == 0 {
		t.Fatal("no scenarios")
	}

	// world is a Source of the seeded type whose fields are named by key.
	type world struct {
		c      *database.Catalog
		userID []byte
		src    sources.Source
		byKey  map[string]metadatafields.Field
	}
	newWorld := func(t *testing.T) world {
		t.Helper()
		c, userID := testWorld(t)
		w := world{c: c, userID: userID, src: birthOfAlice(t, c, userID), byKey: map[string]metadatafields.Field{}}
		// The type's list format reads each needed field outside any group.
		var list string
		for _, key := range file.Needed {
			list += "{source.meta." + key + "}."
		}
		cert, err := sourcetypes.Lookup(c, "birth_certificate", vocabkey.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := sourcetypestest.SetFormats(c, userID, cert.ID, sourcetypes.Formats{List: list}); err != nil {
			t.Fatal(err)
		}
		for _, key := range file.Suggestions {
			f, err := metadatafields.Lookup(c, key, vocabkey.OriginProvenencia)
			if err != nil {
				t.Fatal(err)
			}
			w.byKey[key] = f
		}
		for _, key := range file.Extras {
			f, err := runFieldCreate(c, key, metadatafields.DataTypeText, "")
			if err != nil {
				t.Fatal(err)
			}
			if f.Key != key {
				t.Fatalf("extra %q got key %q", key, f.Key)
			}
			w.byKey[key] = f
		}
		return w
	}
	boardOf := func(t *testing.T, w world) draftBoard {
		t.Helper()
		entries, err := ListWorkspace(w.c, sourceOwner(w.src.ID))
		if err != nil {
			t.Fatal(err)
		}
		board := draftBoard{Saved: []draftSaved{}, Empty: []string{}}
		for _, e := range entries {
			switch {
			case e.Value != nil:
				board.Saved = append(board.Saved, draftSaved{Field: e.Field.Key, Value: e.Value.ValueText, Suggested: e.Suggested})
			case e.Suggested:
				board.Empty = append(board.Empty, e.Field.Key)
			}
		}
		return board
	}
	run := func(t *testing.T, w world, actions []draftAction) {
		t.Helper()
		for _, a := range actions {
			id := w.byKey[a.Field].ID
			var err error
			switch a.Op {
			case "set":
				_, err = runSet(w.c, w.userID, w.src.ID, Input{FieldID: id, ValueText: a.Value})
			case "clear":
				err = runClear(w.c, w.userID, w.src.ID, id)
			case "dismiss":
				err = runDismissSuggestion(w.c, w.userID, w.src.ID, id)
			case "move":
				// The client's drag: saved rows with the moved one before another, then the empty suggestions.
				board := boardOf(t, w)
				var order []string
				for _, s := range board.Saved {
					if s.Field != a.Field {
						order = append(order, s.Field)
					}
				}
				at := len(order)
				if a.Before != nil {
					at = slices.Index(order, *a.Before)
				}
				order = slices.Insert(order, at, a.Field)
				order = append(order, board.Empty...)
				ids := make([][]byte, len(order))
				for i, key := range order {
					ids[i] = w.byKey[key].ID
				}
				err = runReorder(w.c, w.userID, w.src.ID, ids)
			default:
				t.Fatalf("unknown op %q", a.Op)
			}
			if a.Refused {
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("%s %s: %v, want sourcemetadata.invalid", a.Op, a.Field, err)
				}
				continue
			}
			if err != nil {
				t.Fatalf("%s %s: %v", a.Op, a.Field, err)
			}
		}
	}
	// ordered is the board with its saved rows sorted, for a scenario whose order the draft host
	// is known to differ on.
	ordered := func(b draftBoard) draftBoard {
		b.Saved = slices.Clone(b.Saved)
		slices.SortFunc(b.Saved, func(x, y draftSaved) int { return bytes.Compare([]byte(x.Field), []byte(y.Field)) })
		return b
	}

	for _, sc := range file.Scenarios {
		t.Run(sc.Name, func(t *testing.T) {
			w := newWorld(t)
			run(t, w, sc.Seed)
			if got := boardOf(t, w); !reflect.DeepEqual(got, sc.Start) {
				t.Fatalf("start board\n got %+v\nwant %+v", got, sc.Start)
			}
			run(t, w, sc.Actions)
			if got := boardOf(t, w); !reflect.DeepEqual(got, sc.Board) {
				t.Fatalf("live board\n got %+v\nwant %+v", got, sc.Board)
			}

			// The draft the client sends from the same actions gives the same board.
			w = newWorld(t)
			run(t, w, sc.Seed)
			want := sc.Start
			if sc.Draft != nil {
				var d Draft
				for _, v := range sc.Draft.Values {
					d.Values = append(d.Values, Input{FieldID: w.byKey[v.Field].ID, ValueText: v.Value})
				}
				for _, l := range sc.Draft.Layout {
					d.Layout = append(d.Layout, LayoutRow{FieldID: w.byKey[l.Field].ID, SortOrder: l.SortOrder, Dismissed: l.Dismissed})
				}
				if _, err := runApply(w.c, w.userID, w.src.ID, d); err != nil {
					t.Fatal(err)
				}
				want = sc.Board
			}
			got := boardOf(t, w)
			if sc.KnownDivergence == "order" {
				got, want = ordered(got), ordered(want)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("applied draft board\n got %+v\nwant %+v", got, want)
			}
		})
	}
}

// ownerOps runs the value writes on any owner, each in its own transaction.
type ownerOps struct {
	t      *testing.T
	c      *database.Catalog
	userID []byte
	owner  Owner
}

func (o ownerOps) set(in Input) (Row, []rowchange.Change, error) {
	o.t.Helper()
	return runSetOwner(o.c, o.userID, o.owner, in)
}

func (o ownerOps) clear(fieldID []byte) []rowchange.Change {
	o.t.Helper()
	changes, err := runClearOwner(o.c, o.userID, o.owner, fieldID)
	if err != nil {
		o.t.Fatal(err)
	}
	return changes
}

// stored is a value row's columns as SQLite holds them.
func (o ownerOps) stored(fieldID []byte) (textNull bool, unknown, supplied, n int) {
	o.t.Helper()
	db, err := o.c.DB()
	if err != nil {
		o.t.Fatal(err)
	}
	st := stores[o.owner.Level]
	err = db.QueryRow(`SELECT COUNT(*) FROM `+st.values+` WHERE field_id = ?`, fieldID).Scan(&n)
	if err != nil {
		o.t.Fatal(err)
	}
	if n == 0 {
		return
	}
	err = db.QueryRow(`SELECT value_text IS NULL, is_unknown, is_supplied FROM `+st.values+` WHERE field_id = ?`, fieldID).
		Scan(&textNull, &unknown, &supplied)
	if err != nil {
		o.t.Fatal(err)
	}
	return
}

func TestValueMarks(t *testing.T) {
	for _, level := range []vocabkey.Level{vocabkey.LevelSource, vocabkey.LevelArtifact, vocabkey.LevelGraph} {
		t.Run(string(level), func(t *testing.T) {
			c, userID, src, doc, link, f1, f2 := applyWorld(t)
			owner := sourceOwner(src.ID)
			if level != vocabkey.LevelSource {
				owner = newLevelOwner(t, c, userID, src, level)
			}
			o := ownerOps{t: t, c: c, userID: userID, owner: owner}
			unknown, supplied := valuemark.Mark{Unknown: true}, valuemark.Mark{Supplied: true}

			// The steps on f1 run in order: each starts from the row the one before left.
			t.Run("unknown, no text, is stored with a NULL text", func(t *testing.T) {
				row, changes, err := o.set(Input{FieldID: f1.ID, Mark: unknown})
				if err != nil || row.Mark != unknown || row.ValueText != "" || len(changes) != 1 {
					t.Fatalf("unknown = %+v, %v, %d changes", row, err, len(changes))
				}
				if textNull, u, s, n := o.stored(f1.ID); n != 1 || !textNull || u != 1 || s != 0 {
					t.Fatalf("stored unknown = %v %d %d %d", textNull, u, s, n)
				}
			})
			t.Run("the same unknown again changes nothing, blank text counting as no text", func(t *testing.T) {
				if _, changes, err := o.set(Input{FieldID: f1.ID, ValueText: "   ", Mark: unknown}); err != nil || len(changes) != 0 {
					t.Fatalf("repeat = %v, %d changes", err, len(changes))
				}
			})
			t.Run("text and no mark over unknown is an update to a plain value", func(t *testing.T) {
				row, changes, err := o.set(Input{FieldID: f1.ID, ValueText: "Lawrence"})
				if err != nil || row.Mark != (valuemark.Mark{}) || row.ValueText != "Lawrence" || len(changes) != 1 || changes[0].Action != rowchange.ActionUpdate {
					t.Fatalf("plain over unknown = %+v, %v, %+v", row, err, changes)
				}
			})
			t.Run("unknown over text again", func(t *testing.T) {
				if _, _, err := o.set(Input{FieldID: f1.ID, Mark: unknown}); err != nil {
					t.Fatal(err)
				}
			})
			t.Run("dismissing a field that holds an unknown value is a no-op", func(t *testing.T) {
				_, err := writes.RunChanges(c, writes.Op{Action: "dismiss_metadata_suggestion", UserID: userID},
					func(tx *database.Tx) ([]rowchange.Change, error) {
						ch, err := DismissSuggestion(tx, userID, owner, f1.ID)
						if len(ch) != 0 {
							t.Errorf("dismiss changes %+v", ch)
						}
						return ch, err
					})
				if err != nil {
					t.Fatal(err)
				}
			})
			t.Run("the field is in use while an unknown value holds it", func(t *testing.T) {
				_, err := writes.RunChanges(c, writes.Op{Action: "delete_metadata_field", UserID: userID},
					func(tx *database.Tx) ([]rowchange.Change, error) {
						return metadatafields.Delete(tx, userID, f1.ID)
					})
				if !errors.Is(err, metadatafields.ErrInUse) {
					t.Fatalf("delete field = %v, want ErrInUse", err)
				}
			})
			t.Run("clearing an unknown value deletes its row", func(t *testing.T) {
				if ch := o.clear(f1.ID); len(ch) != 1 || ch[0].Action != rowchange.ActionDelete {
					t.Fatalf("clear = %+v", ch)
				}
				if _, _, _, n := o.stored(f1.ID); n != 0 {
					t.Fatalf("row left after clear: %d", n)
				}
			})

			// Refused marks name the field and write nothing; an empty value
			// with no mark is still the old refusal.
			refused := []struct {
				name     string
				in       Input
				wantMark bool
			}{
				{"unknown with text", Input{FieldID: f2.ID, ValueText: "Topeka", Mark: unknown}, true},
				{"both marks", Input{FieldID: f2.ID, Mark: valuemark.Mark{Unknown: true, Supplied: true}}, true},
				{"supplied with blank text", Input{FieldID: f2.ID, ValueText: "  ", Mark: supplied}, true},
				{"empty with no mark", Input{FieldID: f2.ID}, false},
			}
			for _, tt := range refused {
				t.Run("refused: "+tt.name, func(t *testing.T) {
					_, changes, err := o.set(tt.in)
					if tt.wantMark {
						if !errors.Is(err, valuemark.ErrInvalid) || len(changes) != 0 {
							t.Fatalf("set %+v = %v", tt.in, err)
						}
						if got := errParams(t, err); !slices.Equal(got, []string{database.UUIDString(f2.ID)}) {
							t.Fatalf("params %v", got)
						}
					} else if !errors.Is(err, ErrInvalid) || errors.Is(err, valuemark.ErrInvalid) {
						t.Fatalf("set %+v = %v, want ErrInvalid", tt.in, err)
					}
					if _, _, _, n := o.stored(f2.ID); n != 0 {
						t.Fatalf("refused set wrote a row")
					}
				})
			}

			t.Run("supplied text is stored with its mark, and a repeat changes nothing", func(t *testing.T) {
				row, _, err := o.set(Input{FieldID: f2.ID, ValueText: "Topeka", Mark: supplied})
				if err != nil || row.Mark != supplied || row.ValueText != "Topeka" {
					t.Fatalf("supplied = %+v, %v", row, err)
				}
				if _, changes, err := o.set(Input{FieldID: f2.ID, ValueText: "Topeka", Mark: supplied}); err != nil || len(changes) != 0 {
					t.Fatalf("repeat supplied = %v, %d changes", err, len(changes))
				}
			})

			// A url field: unknown has no text; supplied is canonicalized as any url is; a non-link is refused.
			urls := []struct {
				name     string
				in       Input
				wantErr  bool
				wantText string
				wantHref string
				wantU    int
				wantS    int
			}{
				{name: "unknown has no text and no href", in: Input{FieldID: link.ID, Mark: unknown}, wantU: 1},
				{name: "supplied is canonicalized for the href", in: Input{FieldID: link.ID, ValueText: "ancestry.com/x", Mark: supplied},
					wantText: "ancestry.com/x", wantHref: "https://ancestry.com/x", wantS: 1},
				{name: "a supplied non-link is refused", in: Input{FieldID: link.ID, ValueText: "not a link", Mark: supplied}, wantErr: true},
			}
			for _, tt := range urls {
				t.Run("url: "+tt.name, func(t *testing.T) {
					row, _, err := o.set(tt.in)
					if tt.wantErr {
						if !errors.Is(err, ErrInvalid) {
							t.Fatalf("set = %v, want ErrInvalid", err)
						}
						return
					}
					if err != nil || row.ValueText != tt.wantText || row.Mark != tt.in.Mark {
						t.Fatalf("set = %+v, %v", row, err)
					}
					if _, u, s, _ := o.stored(link.ID); u != tt.wantU || s != tt.wantS {
						t.Fatalf("is_unknown %d is_supplied %d", u, s)
					}
					if href := (WorkspaceEntry{Field: link, Value: &row}).OpenHref(); href != tt.wantHref {
						t.Fatalf("href %q, want %q", href, tt.wantHref)
					}
				})
			}

			// The table refuses what Go refuses, for a crafted write.
			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			st := stores[level]
			crafted := []struct {
				name string
				args []any
			}{
				{"unknown with text", []any{"text", 1, 0}},
				{"supplied with no text", []any{nil, 0, 1}},
				{"both marks", []any{nil, 1, 1}},
				{"is_unknown out of range", []any{"text", 2, 0}},
			}
			for _, tt := range crafted {
				t.Run("crafted insert: "+tt.name, func(t *testing.T) {
					id, err := uuid.NewV7()
					if err != nil {
						t.Fatal(err)
					}
					_, err = db.Exec(st.insert, id[:], owner.ID, doc.ID, tt.args[0], tt.args[1], tt.args[2])
					if err == nil || !database.IsConstraintViolation(err) {
						t.Fatalf("crafted insert %v = %v, want a constraint error", tt.args, err)
					}
				})
			}
		})
	}
}

// markDiff is one field diff a value change is expected to carry.
type markDiff struct {
	key      string
	old, new any
}

// checkChange fails unless ch has exactly the field keys and the diffs listed.
func checkChange(t *testing.T, ch rowchange.Change, keys []string, diffs []markDiff) {
	t.Helper()
	if !slices.Equal(fieldKeys(ch), keys) {
		t.Fatalf("%s keys %v, want %v", ch.Action, fieldKeys(ch), keys)
	}
	for _, w := range diffs {
		if d := ch.Fields[w.key]; d.Old != w.old || d.New != w.new {
			t.Fatalf("%s %s = %+v, want old %v new %v", ch.Action, w.key, d, w.old, w.new)
		}
	}
}

func TestValueMarkChanges(t *testing.T) {
	unknown, supplied := valuemark.Mark{Unknown: true}, valuemark.Mark{Supplied: true}
	// Each case sets prior (if any), then in, and wants the one change in's set
	// makes: its action's field keys and the diffs listed.
	tests := []struct {
		name     string
		prior    []Input // set first, on f2
		in       Input
		onF1     bool // run on f1 instead of f2
		wantKeys []string
		wantDiff []markDiff
	}{
		{name: "create unknown", in: Input{Mark: unknown}, onF1: true,
			wantKeys: []string{"field_id", "id", "is_supplied", "is_unknown", "source_id", "value_text"},
			wantDiff: []markDiff{{"value_text", nil, nil}, {"is_unknown", nil, true}, {"is_supplied", nil, false}}},
		{name: "text over unknown", prior: []Input{{Mark: unknown}}, in: Input{ValueText: "Lawrence"}, onF1: true,
			wantKeys: []string{"is_unknown", "value_text"},
			wantDiff: []markDiff{{"value_text", nil, "Lawrence"}, {"is_unknown", true, false}}},
		{name: "new text on a supplied value", prior: []Input{{ValueText: "A", Mark: supplied}}, in: Input{ValueText: "B", Mark: supplied},
			wantKeys: []string{"value_text"}},
		{name: "supplied to unknown", prior: []Input{{ValueText: "B", Mark: supplied}}, in: Input{Mark: unknown},
			wantKeys: []string{"is_supplied", "is_unknown", "value_text"},
			wantDiff: []markDiff{{"value_text", "B", nil}, {"is_unknown", false, true}, {"is_supplied", true, false}}},
		{name: "text with no mark to unknown has no is_supplied diff", prior: []Input{{ValueText: "A"}}, in: Input{Mark: unknown},
			wantKeys: []string{"is_unknown", "value_text"},
			wantDiff: []markDiff{{"value_text", "A", nil}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, userID, src, _, _, f1, f2 := applyWorld(t)
			o := ownerOps{t: t, c: c, userID: userID, owner: sourceOwner(src.ID)}
			field := f2
			if tt.onF1 {
				field = f1
			}
			for _, p := range tt.prior {
				p.FieldID = field.ID
				if _, _, err := o.set(p); err != nil {
					t.Fatal(err)
				}
			}
			tt.in.FieldID = field.ID
			_, changes, err := o.set(tt.in)
			if err != nil || len(changes) != 1 {
				t.Fatalf("changes %+v, %v", changes, err)
			}
			checkChange(t, changes[0], tt.wantKeys, tt.wantDiff)
		})
	}

	t.Run("delete a supplied value", func(t *testing.T) {
		c, userID, src, _, _, _, f2 := applyWorld(t)
		o := ownerOps{t: t, c: c, userID: userID, owner: sourceOwner(src.ID)}
		if _, _, err := o.set(Input{FieldID: f2.ID, ValueText: "A", Mark: supplied}); err != nil {
			t.Fatal(err)
		}
		changes := o.clear(f2.ID)
		if len(changes) != 1 {
			t.Fatalf("changes %+v", changes)
		}
		checkChange(t, changes[0],
			[]string{"field_id", "id", "is_supplied", "is_unknown", "source_id", "value_text"},
			[]markDiff{{"is_unknown", false, nil}, {"is_supplied", true, nil}, {"value_text", "A", nil}})
	})
}

// TestDamagedMarksReadSettled damages stored mark pairs the table checks refuse,
// as a crafted catalog holds them. Every read returns the pair settled for the
// value (valuemark.Settle); a write over the row diffs against the stored pair.
func TestDamagedMarksReadSettled(t *testing.T) {
	unknown, supplied := valuemark.Mark{Unknown: true}, valuemark.Mark{Supplied: true}
	cases := []struct {
		name   string
		in     Input
		damage string
		want   Input
	}{
		{"unknown beside text reads unmarked", Input{ValueText: "Lawrence"}, "is_unknown = 1", Input{ValueText: "Lawrence"}},
		{"both beside text reads supplied", Input{ValueText: "Lawrence", Mark: supplied}, "is_unknown = 1", Input{ValueText: "Lawrence", Mark: supplied}},
		{"both without text reads unknown", Input{Mark: unknown}, "is_supplied = 1", Input{Mark: unknown}},
		{"supplied without text reads unmarked", Input{Mark: unknown}, "is_unknown = 0, is_supplied = 1", Input{}},
		{"supplied beside blank text reads unmarked", Input{ValueText: "Lawrence", Mark: supplied}, "value_text = '  '", Input{ValueText: "  "}},
	}
	for _, level := range []vocabkey.Level{vocabkey.LevelSource, vocabkey.LevelArtifact, vocabkey.LevelGraph} {
		for _, tc := range cases {
			t.Run(string(level)+"/"+tc.name, func(t *testing.T) {
				c, userID, src, _, _, f1, _ := applyWorld(t)
				owner := sourceOwner(src.ID)
				if level != vocabkey.LevelSource {
					owner = newLevelOwner(t, c, userID, src, level)
				}
				o := ownerOps{t: t, c: c, userID: userID, owner: owner}
				in := tc.in
				in.FieldID = f1.ID
				if _, _, err := o.set(in); err != nil {
					t.Fatal(err)
				}
				db := dbtest.MustDB(t, c)
				dbtest.Damage(t, db, `UPDATE `+stores[level].values+` SET `+tc.damage+` WHERE field_id = ?`, f1.ID)

				want := tc.want
				want.FieldID = f1.ID
				tx, err := db.Begin()
				if err != nil {
					t.Fatal(err)
				}
				d, err := ReadDraftTx(tx, owner)
				if err != nil {
					t.Fatal(err)
				}
				rows, err := ListByOwnersTx(tx, []Owner{owner})
				if err != nil {
					t.Fatal(err)
				}
				_ = tx.Rollback()
				if len(d.Values) != 1 || !reflect.DeepEqual(d.Values[0], want) {
					t.Errorf("draft = %+v, want %+v", d.Values, want)
				}
				if len(rows) != 1 || rows[0].ValueText != want.ValueText || rows[0].Mark != want.Mark {
					t.Errorf("rows = %+v, want %+v", rows, want)
				}
				ws, err := ListWorkspace(c, owner)
				if err != nil {
					t.Fatal(err)
				}
				var found bool
				for _, e := range ws {
					if e.Value != nil && bytes.Equal(e.Value.FieldID, f1.ID) {
						found = true
						if e.Value.ValueText != want.ValueText || e.Value.Mark != want.Mark {
							t.Errorf("workspace value = %+v, want %+v", e.Value, want)
						}
					}
				}
				if !found {
					t.Errorf("workspace has no value for the field: %+v", ws)
				}
			})
		}
	}

	t.Run("a write over a damaged row diffs against the stored pair", func(t *testing.T) {
		c, userID, src, _, _, f1, _ := applyWorld(t)
		o := ownerOps{t: t, c: c, userID: userID, owner: sourceOwner(src.ID)}
		if _, _, err := o.set(Input{FieldID: f1.ID, ValueText: "Lawrence"}); err != nil {
			t.Fatal(err)
		}
		dbtest.Damage(t, dbtest.MustDB(t, c), `UPDATE source_metadata SET is_unknown = 1 WHERE field_id = ?`, f1.ID)
		_, changes, err := o.set(Input{FieldID: f1.ID, ValueText: "Lawrence"})
		if err != nil {
			t.Fatal(err)
		}
		if len(changes) != 1 || changes[0].Action != rowchange.ActionUpdate ||
			!reflect.DeepEqual(changes[0].Fields, map[string]rowchange.FieldDiff{"is_unknown": {Old: true, New: false}}) {
			t.Fatalf("changes = %+v, want is_unknown true → false", changes)
		}
		if _, u, s, n := o.stored(f1.ID); n != 1 || u != 0 || s != 0 {
			t.Fatalf("stored = %d %d %d", u, s, n)
		}
	})
}

// seededType is the product-seeded Source type with the key.
func seededType(t *testing.T, c *database.Catalog, key string) sourcetypes.Type {
	t.Helper()
	ty, err := sourcetypes.Lookup(c, key, vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return ty
}

// seededField is the product-seeded metadata field with the key.
func seededField(t *testing.T, c *database.Catalog, key string) metadatafields.Field {
	t.Helper()
	f, err := metadatafields.Lookup(c, key, vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// entryOf is the entry for the field key, and whether the board holds one.
func entryOf(entries []WorkspaceEntry, key string) (WorkspaceEntry, bool) {
	for _, e := range entries {
		if e.Field.Key == key {
			return e, true
		}
	}
	return WorkspaceEntry{}, false
}

// boardOf reads one owner's board.
func boardOf(t *testing.T, c *database.Catalog, o Owner) Board {
	t.Helper()
	boards, err := ListWorkspaces(c, []Owner{o})
	if err != nil {
		t.Fatal(err)
	}
	return boards[0]
}

// ledger is a custom type whose full format reads line_number from a graph, else the Source.
type ledger struct {
	typeID []byte
	line   metadatafields.Field
	src    sources.Source
	graph  Owner
}

// newLedger makes the line-number type (suggesting line_number at graph) in a seeded
// world, a Source of it and one graph.
func newLedger(t *testing.T, c *database.Catalog, userID []byte) ledger {
	t.Helper()
	l := ledger{line: seededField(t, c, "line_number")}
	ty, _, err := writes.Run(c, writes.Op{Action: "create_source_type", UserID: userID},
		func(tx *database.Tx) (sourcetypes.Type, []rowchange.Change, error) {
			return sourcetypes.Create(tx, sourcetypes.CreateInput{Label: "Line ledger"})
		})
	if err != nil {
		t.Fatal(err)
	}
	l.typeID = ty.ID
	if err := sourcevocab.AppendSuggestion(c, ty.ID, l.line.ID, vocabkey.LevelGraph); err != nil {
		t.Fatal(err)
	}
	l.setFormats(t, c, userID, sourcetypes.Formats{Full: "{graph.meta.line_number|source.meta.line_number}."})
	l.src = newSource(t, c, userID, ty.ID, "Ledger 1")
	l.graph = newLevelOwner(t, c, userID, l.src, vocabkey.LevelGraph)
	return l
}

func (l ledger) setFormats(t *testing.T, c *database.Catalog, userID []byte, f sourcetypes.Formats) {
	t.Helper()
	if _, err := sourcetypestest.SetFormats(c, userID, l.typeID, f); err != nil {
		t.Fatal(err)
	}
}

func TestListWorkspacesAnnotations(t *testing.T) {
	c, userID := seededWorld(t)
	needed := citationformat.NeedNeeded

	t.Run("census", func(t *testing.T) {
		src := newSource(t, c, userID, seededType(t, c, "census").ID, "1900 census")
		art := newLevelOwner(t, c, userID, src, vocabkey.LevelArtifact)
		boards, err := ListWorkspaces(c, []Owner{sourceOwner(src.ID), art})
		if err != nil {
			t.Fatal(err)
		}
		jur, ok := entryOf(boards[0].Entries, "jurisdiction")
		if !ok || !reflect.DeepEqual(jur.Annotation, Annotation{Need: needed}) {
			t.Errorf("jurisdiction = %+v, %v; want needed, not met, no alternatives", jur.Annotation, ok)
		}
		if e, ok := entryOf(boards[0].Entries, "creator"); !ok || e.Annotation.Need != citationformat.NeedUsed {
			t.Errorf("creator = %+v, %v; want used", e.Annotation, ok)
		}
		if e, ok := entryOf(boards[1].Entries, "url"); !ok || e.Annotation.Need != citationformat.NeedUsed {
			t.Errorf("artifact url = %+v, %v; want used", e.Annotation, ok)
		}
		for _, key := range []string{"item", "line_number", "folio", "dwelling_number", "family_number", "schedule_number"} {
			if _, ok := entryOf(boards[0].Entries, key); ok {
				t.Errorf("%s is on the census Source board; a later option shows only where suggested or valued", key)
			}
		}
	})

	t.Run("photograph shows an unsuggested head after the suggestions", func(t *testing.T) {
		src := newSource(t, c, userID, seededType(t, c, "photograph").ID, "Portrait of Alice")
		entries := boardOf(t, c, sourceOwner(src.ID)).Entries
		last := entries[len(entries)-1]
		if last.Field.Key != "short_title" || last.Suggested || last.Value != nil || last.Annotation.Need != citationformat.NeedUsed {
			t.Fatalf("last entry = %+v, want the empty unsuggested used short title", last)
		}
		for _, e := range entries[:len(entries)-1] {
			if !e.Suggested {
				t.Errorf("%s comes before the suggestions end", e.Field.Key)
			}
		}
	})

	t.Run("a valued field the formats read is met elsewhere", func(t *testing.T) {
		src := newSource(t, c, userID, seededType(t, c, "book").ID, "A book")
		if _, err := runSet(c, userID, src.ID, Input{FieldID: seededField(t, c, "author").ID, ValueText: "Ada Robins"}); err != nil {
			t.Fatal(err)
		}
		e, ok := entryOf(boardOf(t, c, sourceOwner(src.ID)).Entries, "author_sort")
		if !ok || e.Suggested || e.Annotation.Need != needed || !e.Annotation.MetElsewhere {
			t.Fatalf("author_sort = %+v, %v; want unsuggested, needed, met elsewhere", e, ok)
		}
	})

	t.Run("a valued extra is annotated once", func(t *testing.T) {
		src := newSource(t, c, userID, seededType(t, c, "census").ID, "1910 census")
		if _, err := runSet(c, userID, src.ID, Input{FieldID: seededField(t, c, "item").ID, ValueText: "Roll 4"}); err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, e := range boardOf(t, c, sourceOwner(src.ID)).Entries {
			if e.Field.Key == "item" {
				n++
				if e.Value == nil || e.Suggested || e.Annotation.Need != citationformat.NeedUsed {
					t.Errorf("item = %+v, want a valued extra, used", e)
				}
			}
		}
		if n != 1 {
			t.Errorf("item comes out %d times", n)
		}
	})

	t.Run("the line-number type", func(t *testing.T) {
		l := newLedger(t, c, userID)
		note := func() Annotation {
			e, ok := entryOf(boardOf(t, c, l.graph).Entries, "line_number")
			if !ok {
				t.Fatal("no line_number entry on the graph board")
			}
			return e.Annotation
		}
		alt := []Alternative{{Level: vocabkey.LevelSource, Field: &l.line}}
		check := func(name string, got Annotation, met bool) {
			t.Helper()
			want := Annotation{Need: needed, MetElsewhere: met, Alternatives: alt}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("%s = %+v, want %+v", name, got, want)
			}
		}
		check("neither holds it", note(), false)
		if _, err := runSet(c, userID, l.src.ID, Input{FieldID: l.line.ID, ValueText: "14"}); err != nil {
			t.Fatal(err)
		}
		check("the Source holds it", note(), true)
		if _, err := runSet(c, userID, l.src.ID, Input{FieldID: l.line.ID, Mark: valuemark.Mark{Unknown: true}}); err != nil {
			t.Fatal(err)
		}
		check("the Source's is unknown", note(), true)
	})

	t.Run("a type with no formats", func(t *testing.T) {
		ty, _, err := writes.Run(c, writes.Op{Action: "create_source_type", UserID: userID},
			func(tx *database.Tx) (sourcetypes.Type, []rowchange.Change, error) {
				return sourcetypes.Create(tx, sourcetypes.CreateInput{Label: "Plain ledger"})
			})
		if err != nil {
			t.Fatal(err)
		}
		if err := sourcevocab.AppendSuggestion(c, ty.ID, seededField(t, c, "folio").ID, vocabkey.LevelSource); err != nil {
			t.Fatal(err)
		}
		src := newSource(t, c, userID, ty.ID, "Plain")
		b := boardOf(t, c, sourceOwner(src.ID))
		if len(b.Entries) != 1 || !reflect.DeepEqual(b.Entries[0].Annotation, Annotation{}) {
			t.Errorf("entries = %+v, want the one suggestion with no annotation", b.Entries)
		}
		if b.Date.Need != "" || b.Page.Need != "" {
			t.Errorf("date %+v, page %+v; want none", b.Date, b.Page)
		}
	})
}

func TestDismissNeeded(t *testing.T) {
	c, userID := seededWorld(t)
	layoutRows := func() int { return countRows(t, c, "source_metadata_layout") }

	cert := newSource(t, c, userID, seededType(t, c, "birth_certificate").ID, "Birth of Alice")
	jur := seededField(t, c, "jurisdiction")
	err := runDismissSuggestion(c, userID, cert.ID, jur.ID)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("dismissing a needed field = %v, want ErrInvalid", err)
	}
	if got := errParams(t, err); !reflect.DeepEqual(got, []string{database.UUIDString(jur.ID)}) {
		t.Errorf("params = %v, want the field's id", got)
	}
	if n := layoutRows(); n != 0 {
		t.Errorf("%d layout rows after a refused dismissal", n)
	}

	t.Run("a dismissed row for a needed field still shows, in its place", func(t *testing.T) {
		if _, err := runApply(c, userID, cert.ID, Draft{Layout: []LayoutRow{{FieldID: jur.ID, SortOrder: 0, Dismissed: true}}}); err != nil {
			t.Fatal(err)
		}
		entries := boardOf(t, c, sourceOwner(cert.ID)).Entries
		if len(entries) == 0 || entries[0].Field.Key != "jurisdiction" {
			t.Fatalf("entries lead with %+v, want jurisdiction at its layout position", entries)
		}
	})

	t.Run("a used unsuggested field hides when dismissed", func(t *testing.T) {
		reg := newSource(t, c, userID, seededType(t, c, "church_register").ID, "Register")
		cov := seededField(t, c, "coverage")
		if e, ok := entryOf(boardOf(t, c, sourceOwner(reg.ID)).Entries, "coverage"); !ok || e.Suggested || e.Annotation.Need != citationformat.NeedUsed {
			t.Fatalf("coverage = %+v, %v; want unsuggested, used", e, ok)
		}
		if err := runDismissSuggestion(c, userID, reg.ID, cov.ID); err != nil {
			t.Fatal(err)
		}
		if _, ok := entryOf(boardOf(t, c, sourceOwner(reg.ID)).Entries, "coverage"); ok {
			t.Error("a dismissed used field is still on the board")
		}
	})

	t.Run("a used suggested field is dismissed as before", func(t *testing.T) {
		census := newSource(t, c, userID, seededType(t, c, "census").ID, "1900 census")
		creator := seededField(t, c, "creator")
		before := layoutRows()
		if err := runDismissSuggestion(c, userID, census.ID, creator.ID); err != nil {
			t.Fatal(err)
		}
		if layoutRows() != before+1 {
			t.Errorf("layout rows %d, want %d", layoutRows(), before+1)
		}
		if _, ok := entryOf(boardOf(t, c, sourceOwner(census.ID)).Entries, "creator"); ok {
			t.Error("a dismissed suggestion is still on the board")
		}
	})
}

func TestAnnotationsFollowFormats(t *testing.T) {
	c, userID := seededWorld(t)
	l := newLedger(t, c, userID)
	line := func() (WorkspaceEntry, bool) { return entryOf(boardOf(t, c, l.graph).Entries, "line_number") }
	if _, err := writes.RunChanges(c, writes.Op{Action: "update_metadata", UserID: userID}, func(tx *database.Tx) ([]rowchange.Change, error) {
		return Apply(tx, userID, l.graph, Draft{Layout: []LayoutRow{{FieldID: l.line.ID, SortOrder: 0, Dismissed: true}}})
	}); err != nil {
		t.Fatal(err)
	}
	if e, ok := line(); !ok || e.Annotation.Need != citationformat.NeedNeeded {
		t.Fatalf("with the format, line_number = %+v, %v; want shown and needed despite its dismissed row", e, ok)
	}

	l.setFormats(t, c, userID, sourcetypes.Formats{})
	if e, ok := line(); ok {
		t.Errorf("formats cleared: the dismissed suggestion = %+v is back", e)
	}
	// The kept layout row hides it again; clearing the dismissal's row is the editor's job.
	b := boardOf(t, c, l.graph)
	for _, e := range b.Entries {
		if e.Annotation.Need != "" {
			t.Errorf("formats cleared: %s still annotated %+v", e.Field.Key, e.Annotation)
		}
	}
}

func TestAnnotationsDanglingField(t *testing.T) {
	c, userID := seededWorld(t)
	l := newLedger(t, c, userID)
	dangling := uuid.New()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	full := "{graph.meta.#" + uuid.UUID(l.line.ID).String() + "|source.meta.#" + dangling.String() + "}.{source.meta.#" + dangling.String() + "}"
	if _, err := db.Exec(`UPDATE source_types SET citation_full = ? WHERE id = ?`, full, l.typeID); err != nil {
		t.Fatal(err)
	}
	if entries := boardOf(t, c, sourceOwner(l.src.ID)).Entries; len(entries) != 0 {
		t.Errorf("source board = %+v, want no entry for a dangling head", entries)
	}
	e, ok := entryOf(boardOf(t, c, l.graph).Entries, "line_number")
	if !ok || e.Annotation.Need != citationformat.NeedNeeded || len(e.Annotation.Alternatives) != 0 || e.Annotation.MetElsewhere {
		t.Errorf("graph line_number = %+v, %v; want needed with no alternative", e.Annotation, ok)
	}
}

func TestSourceBoardBuiltins(t *testing.T) {
	c, userID := seededWorld(t)
	census := newSource(t, c, userID, seededType(t, c, "census").ID, "1900 census")
	court := newSource(t, c, userID, seededType(t, c, "court_report").ID, "Report")
	plain := sourcetypesLedger(t, c, userID)

	if b := boardOf(t, c, sourceOwner(census.ID)); b.Date.Need != citationformat.NeedNeeded || b.Date.MetElsewhere || b.Page.Need != citationformat.NeedUsed {
		t.Errorf("census date %+v, page %+v; want date needed and unmet, page used", b.Date, b.Page)
	}
	b := boardOf(t, c, sourceOwner(court.ID))
	pages := seededField(t, c, "page_range")
	if b.Page.Need != citationformat.NeedNeeded || b.Page.MetElsewhere || len(b.Page.Alternatives) != 1 ||
		b.Page.Alternatives[0].Level != vocabkey.LevelSource || b.Page.Alternatives[0].Field == nil ||
		!bytes.Equal(b.Page.Alternatives[0].Field.ID, pages.ID) {
		t.Errorf("court report page = %+v, want needed, unmet, alternative [source page_range]", b.Page)
	}
	if _, err := runSet(c, userID, court.ID, Input{FieldID: pages.ID, ValueText: "12-14"}); err != nil {
		t.Fatal(err)
	}
	if got := boardOf(t, c, sourceOwner(court.ID)).Page; !got.MetElsewhere {
		t.Errorf("court report page with page_range = %+v, want met elsewhere", got)
	}
	art := newLevelOwner(t, c, userID, census, vocabkey.LevelArtifact)
	if ab := boardOf(t, c, art); ab.Date.Need != "" || ab.Page.Need != "" {
		t.Errorf("artifact board date %+v, page %+v; want none", ab.Date, ab.Page)
	}
	if lb := boardOf(t, c, sourceOwner(plain.ID)); lb.Date.Need != "" || lb.Page.Need != "" {
		t.Errorf("ledger date %+v, page %+v; want none", lb.Date, lb.Page)
	}
}

// sourcetypesLedger is a Source of a new type with no formats.
func sourcetypesLedger(t *testing.T, c *database.Catalog, userID []byte) sources.Source {
	t.Helper()
	ty, _, err := writes.Run(c, writes.Op{Action: "create_source_type", UserID: userID},
		func(tx *database.Tx) (sourcetypes.Type, []rowchange.Change, error) {
			return sourcetypes.Create(tx, sourcetypes.CreateInput{Label: "Ledger"})
		})
	if err != nil {
		t.Fatal(err)
	}
	return newSource(t, c, userID, ty.ID, "Ledger")
}

// seedExpect is what the seeded formats ask of one Source type's Source board,
// measured over full, short and list at planning.
type seedExpect struct {
	needed  []string // Source fields needed, valued or not
	unsug   []string // empty unsuggested heads, "key:need"
	omitted []string // later options the board leaves off
	date    citationformat.Need
	page    citationformat.Need
}

var seedExpectations = map[string]seedExpect{
	"birth_certificate":    {needed: []string{"jurisdiction"}, omitted: []string{"entry_number", "item"}, date: "used", page: "used"},
	"book":                 {needed: []string{"author", "author_sort"}, unsug: []string{"short_title:used", "author_sort:needed"}, date: "used", page: "used"},
	"broadcast":            {needed: []string{"author", "published_in", "author_sort"}, unsug: []string{"short_title:used", "author_sort:needed"}, date: "used", page: "used"},
	"case_file":            {needed: []string{"jurisdiction"}, unsug: []string{"short_title:used", "collection:used"}, omitted: []string{"item"}},
	"cemetery_marker":      {needed: []string{"creator", "cemetery_location"}, omitted: []string{"item"}},
	"census":               {needed: []string{"jurisdiction"}, omitted: []string{"dwelling_number", "family_number", "folio", "item", "line_number", "schedule_number"}, date: "needed", page: "used"},
	"church_register":      {needed: []string{"creator"}, unsug: []string{"short_title:used", "coverage:used"}, omitted: []string{"entry_number", "folio", "item"}, page: "used"},
	"correspondence":       {needed: []string{"creator", "recipient"}, unsug: []string{"coverage:used"}, date: "needed"},
	"court_report":         {needed: []string{"published_in"}, unsug: []string{"short_title:used"}, date: "used", page: "needed"},
	"death_certificate":    {needed: []string{"jurisdiction"}, omitted: []string{"entry_number", "item"}, date: "used", page: "used"},
	"dna_match_report":     {needed: []string{"testing_company"}, unsug: []string{"short_title:used"}, omitted: []string{"item"}, date: "used"},
	"family_bible":         {unsug: []string{"short_title:used", "coverage:used"}, omitted: []string{"item"}, date: "used"},
	"gedcom_file":          {unsug: []string{"short_title:used"}, omitted: []string{"item"}, date: "used"},
	"government_file":      {needed: []string{"repository"}, unsug: []string{"short_title:used"}, omitted: []string{"item"}, date: "used"},
	"interview":            {needed: []string{"interviewee", "interviewer"}, omitted: []string{"item"}, date: "needed"},
	"journal_article":      {needed: []string{"author", "published_in", "author_sort"}, unsug: []string{"short_title:used", "author_sort:needed"}, date: "used", page: "used"},
	"manuscript":           {needed: []string{"repository"}, unsug: []string{"short_title:used", "coverage:used"}, omitted: []string{"item"}, date: "used"},
	"marriage_certificate": {needed: []string{"jurisdiction"}, omitted: []string{"entry_number", "item"}, date: "used", page: "used"},
	"newspaper_item":       {needed: []string{"published_in"}, unsug: []string{"short_title:used"}, omitted: []string{"column"}, date: "needed", page: "used"},
	"oral_testimony":       {needed: []string{"interviewee"}, omitted: []string{"item"}, date: "needed"},
	"passenger_list":       {unsug: []string{"short_title:used", "creator:used", "collection:used", "repository:used"}, omitted: []string{"item", "line_number"}, page: "used"},
	"photograph":           {unsug: []string{"short_title:used"}, date: "used"},
	"private_papers":       {unsug: []string{"short_title:used", "coverage:used"}, omitted: []string{"item"}, date: "used"},
	"record_book":          {needed: []string{"jurisdiction"}, unsug: []string{"coverage:used"}, omitted: []string{"entry_number", "item"}, page: "used"},
	"statute":              {needed: []string{"jurisdiction", "published_in"}, date: "used", page: "used"},
	"website":              {needed: []string{"site_name"}, unsug: []string{"short_title:used"}, omitted: []string{"item"}},
}

func TestSeedAnnotations(t *testing.T) {
	c, userID := seededWorld(t)
	list, err := sourcetypes.List(c)
	if err != nil {
		t.Fatal(err)
	}
	var seeded []sourcetypes.Type
	for _, ty := range list {
		if ty.Origin == vocabkey.OriginProvenencia {
			seeded = append(seeded, ty)
		}
	}
	if len(seeded) != len(seedExpectations) {
		t.Fatalf("%d seeded types, the table holds %d", len(seeded), len(seedExpectations))
	}
	var neededAtSource, neededBelow, unsuggested, dates, pages, omittedPairs int
	for _, ty := range seeded {
		want, ok := seedExpectations[ty.Key]
		if !ok {
			t.Errorf("no expectation for %s", ty.Key)
			continue
		}
		t.Run(ty.Key, func(t *testing.T) {
			src := newSource(t, c, userID, ty.ID, ty.Key)
			art := newLevelOwner(t, c, userID, src, vocabkey.LevelArtifact)
			gr := newLevelOwner(t, c, userID, src, vocabkey.LevelGraph)
			boards, err := ListWorkspaces(c, []Owner{sourceOwner(src.ID), art, gr})
			if err != nil {
				t.Fatal(err)
			}
			var needed, unsug []string
			for _, e := range boards[0].Entries {
				if e.Annotation.Need == citationformat.NeedNeeded {
					needed = append(needed, e.Field.Key)
				}
				if !e.Suggested && e.Value == nil && e.Annotation.Need != "" {
					unsug = append(unsug, e.Field.Key+":"+string(e.Annotation.Need))
				}
			}
			for _, key := range want.omitted {
				if _, ok := entryOf(boards[0].Entries, key); ok {
					t.Errorf("%s is on the Source board; a later option shows only where suggested or valued", key)
				}
			}
			for _, b := range boards[1:] {
				for _, e := range b.Entries {
					if e.Annotation.Need == citationformat.NeedNeeded {
						neededBelow++
						t.Errorf("%s is needed on an Artifact or graph board", e.Field.Key)
					}
				}
			}
			slices.Sort(needed)
			wantNeeded := slices.Clone(want.needed)
			slices.Sort(wantNeeded)
			if !slices.Equal(needed, wantNeeded) {
				t.Errorf("needed = %v, want %v", needed, wantNeeded)
			}
			if !slices.Equal(unsug, want.unsug) {
				t.Errorf("unsuggested heads = %v, want %v", unsug, want.unsug)
			}
			if boards[0].Date.Need != want.date || boards[0].Page.Need != want.page {
				t.Errorf("date %q, page %q; want %q, %q", boards[0].Date.Need, boards[0].Page.Need, want.date, want.page)
			}
			neededAtSource += len(needed)
			unsuggested += len(unsug)
			omittedPairs += len(want.omitted)
			if boards[0].Date.Need == citationformat.NeedNeeded {
				dates++
			}
			if boards[0].Page.Need == citationformat.NeedNeeded {
				pages++
			}
		})
	}
	if neededAtSource != 30 || neededBelow != 0 || unsuggested != 29 || dates != 5 || pages != 1 || omittedPairs != 31 {
		t.Errorf("needed %d (want 30), needed below %d (0), unsuggested heads %d (29), needed dates %d (5), needed pages %d (1), omitted %d (31)",
			neededAtSource, neededBelow, unsuggested, dates, pages, omittedPairs)
	}
}

// TestListWorkspacesAnnotationsScale reads a census Source with 60 Artifacts and
// 60 graphs in one call: every board comes back annotated.
func TestListWorkspacesAnnotationsScale(t *testing.T) {
	c, userID := seededWorld(t)
	src := newSource(t, c, userID, seededType(t, c, "census").ID, "1900 census")
	owners := []Owner{sourceOwner(src.ID)}
	for i := 0; i < 60; i++ {
		owners = append(owners, newLevelOwner(t, c, userID, src, vocabkey.LevelArtifact))
	}
	for i := 0; i < 60; i++ {
		owners = append(owners, newLevelOwner(t, c, userID, src, vocabkey.LevelGraph))
	}
	boards, err := ListWorkspaces(c, owners)
	if err != nil {
		t.Fatal(err)
	}
	if len(boards) != 121 {
		t.Fatalf("%d boards, want 121", len(boards))
	}
	key := map[vocabkey.Level]string{vocabkey.LevelSource: "jurisdiction", vocabkey.LevelArtifact: "url", vocabkey.LevelGraph: "item"}
	want := map[vocabkey.Level]citationformat.Need{
		vocabkey.LevelSource: citationformat.NeedNeeded, vocabkey.LevelArtifact: citationformat.NeedUsed, vocabkey.LevelGraph: citationformat.NeedUsed,
	}
	for i, o := range owners {
		e, ok := entryOf(boards[i].Entries, key[o.Level])
		if !ok || e.Annotation.Need != want[o.Level] {
			t.Fatalf("board %d (%s) %s = %+v, %v; want %q", i, o.Level, key[o.Level], e.Annotation, ok, want[o.Level])
		}
	}
}
