package sourcemetadata

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sourcefields"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/sourcevocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestSourceMetadata(t *testing.T) {
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
	mustSeededSource := func(t *testing.T, c *database.Catalog) sources.Source {
		t.Helper()
		if err := sourcevocab.Install(c); err != nil {
			t.Fatal(err)
		}
		st, err := sourcetypes.Lookup(c, "birth_certificate", sourcetypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		s, err := sources.Create(c, userID, sources.CreateInput{
			SourceTypeID: st.ID,
			Title:        "Birth of Alice",
		})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	mustField := func(t *testing.T, c *database.Catalog, key string) sourcefields.Field {
		t.Helper()
		f, err := sourcefields.Lookup(c, key, sourcefields.OriginProvenencia)
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
	metaCount := func(t *testing.T, c *database.Catalog) int {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM source_metadata`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	workspaceKeys := func(t *testing.T, c *database.Catalog, sourceID []byte) []string {
		t.Helper()
		ws, err := ListWorkspace(c, sourceID)
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				row, err := Set(c, userID, Input{
					SourceID: src.ID, FieldID: doc.ID, ValueText: "12345",
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				field, err := sourcefields.Create(c, "Landing page", sourcefields.DataTypeURL, "")
				if err != nil {
					t.Fatal(err)
				}
				row, err := Set(c, userID, Input{
					SourceID: src.ID, FieldID: field.ID, ValueText: "https://example.com/record",
				})
				if err != nil {
					t.Fatal(err)
				}
				if row.ValueText != "https://example.com/record" {
					t.Fatalf("%+v", row)
				}
				if _, err := Set(c, userID, Input{
					SourceID: src.ID, FieldID: field.ID, ValueText: "www.url.com",
				}); err != nil {
					t.Fatal(err)
				}
				if _, err := Set(c, userID, Input{
					SourceID: src.ID, FieldID: field.ID, ValueText: "file:/tmp",
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("file url %v", err)
				}
				if _, err := Set(c, userID, Input{
					SourceID: src.ID, FieldID: field.ID, ValueText: "not a url",
				}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("junk url %v", err)
				}
			},
		},
		{
			name: "date-named fields are text",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSeededSource(t, c)
				rec := mustField(t, c, "record_date")
				if rec.DataType != sourcefields.DataTypeText {
					t.Fatalf("record_date type %q", rec.DataType)
				}
				row, err := Set(c, userID, Input{
					SourceID: src.ID, FieldID: rec.ID, ValueText: "about the year 1890",
				})
				if err != nil {
					t.Fatal(err)
				}
				if row.ValueText != "about the year 1890" {
					t.Fatalf("%+v", row)
				}
				kept, err := Set(c, userID, Input{
					SourceID: src.ID, FieldID: rec.ID,
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				first, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "111"})
				if err != nil {
					t.Fatal(err)
				}
				second, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "222"})
				if err != nil {
					t.Fatal(err)
				}
				if string(first.ID) != string(second.ID) || second.ValueText != "222" {
					t.Fatalf("first=%+v second=%+v", first, second)
				}
				list, err := ListBySource(c, src.ID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				if err := Clear(c, userID, src.ID, doc.ID); err != nil {
					t.Fatal(err)
				}
				if latestAction(t, c) != "update_source_metadata" {
					t.Fatalf("action %q", latestAction(t, c))
				}
				list, err = ListBySource(c, src.ID)
				if err != nil || len(list) != 0 {
					t.Fatalf("after clear %v len=%d", err, len(list))
				}
				if err := Clear(c, userID, src.ID, doc.ID); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "list workspace suggested and extras",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "12345"}); err != nil {
					t.Fatal(err)
				}
				extraID, err := sourcefields.Upsert(c, sourcefields.Field{
					Key: "shelf_mark", Origin: sourcefields.OriginUser, Label: "Shelf mark", DataType: sourcefields.DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: extraID, ValueText: "A-12"}); err != nil {
					t.Fatal(err)
				}
				ws, err := ListWorkspace(c, src.ID)
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
			name: "source delete cascades metadata",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "x"}); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`DELETE FROM sources WHERE id = ?`, src.ID); err != nil {
					t.Fatal(err)
				}
				if metaCount(t, c) != 0 {
					t.Fatalf("count %d", metaCount(t, c))
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				issue := mustField(t, c, "issue_date")
				if err := DismissSuggestion(c, userID, src.ID, issue.ID); err != nil {
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
				if err := DismissSuggestion(c, userID, src.ID, issue.ID); err != nil {
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "12345"}); err != nil {
					t.Fatal(err)
				}
				if err := DismissSuggestion(c, userID, src.ID, doc.ID); err != nil {
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				rec := mustField(t, c, "record_date")
				issue := mustField(t, c, "issue_date")
				if err := DismissSuggestion(c, userID, src.ID, issue.ID); err != nil {
					t.Fatal(err)
				}
				if err := Reorder(c, userID, src.ID, [][]byte{rec.ID, doc.ID, issue.ID}); err != nil {
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				extraID, err := sourcefields.Upsert(c, sourcefields.Field{
					Key: "shelf_mark", Origin: sourcefields.OriginUser, Label: "Shelf mark", DataType: sourcefields.DataTypeText,
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: extraID, ValueText: "A-12"}); err != nil {
					t.Fatal(err)
				}
				if order, dismissed := layoutOf(t, c, src.ID, extraID); order != 0 || dismissed != 0 {
					t.Fatalf("extra layout order=%d dismissed=%d", order, dismissed)
				}
				doc := mustField(t, c, "document_number")
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "12345"}); err != nil {
					t.Fatal(err)
				}
				if order, _ := layoutOf(t, c, src.ID, doc.ID); order != 1 {
					t.Fatalf("document_number order %d", order)
				}
				// Updating an existing value must not append a second slot.
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "999"}); err != nil {
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				missing := make([]byte, 16)
				missing[15] = 9
				if err := DismissSuggestion(c, userID, missing, doc.ID); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if err := DismissSuggestion(c, userID, src.ID, missing); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad field %v", err)
				}
				if err := DismissSuggestion(c, nil, src.ID, doc.ID); !errors.Is(err, ErrInvalid) {
					t.Fatalf("nil user %v", err)
				}
				if err := Reorder(c, userID, src.ID, nil); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty list %v", err)
				}
				if err := Reorder(c, userID, src.ID, [][]byte{doc.ID, doc.ID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("duplicate field %v", err)
				}
				if err := Reorder(c, userID, src.ID, [][]byte{missing}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("missing field %v", err)
				}
				if err := Reorder(c, userID, missing, [][]byte{doc.ID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if err := Reorder(c, userID, src.ID, [][]byte{{1, 2, 3}}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("short field id %v", err)
				}
			},
		},
		{
			name: "source delete cascades layout",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "x"}); err != nil {
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
				mustUser(t, c)
				src := mustSeededSource(t, c)
				doc := mustField(t, c, "document_number")
				missing := make([]byte, 16)
				missing[15] = 9
				if _, err := Set(c, userID, Input{SourceID: missing, FieldID: doc.ID, ValueText: "x"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad source %v", err)
				}
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: missing, ValueText: "x"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("bad field %v", err)
				}
				if _, err := Set(c, nil, Input{SourceID: src.ID, FieldID: doc.ID, ValueText: "x"}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("nil user %v", err)
				}
				if _, err := Set(c, userID, Input{SourceID: src.ID, FieldID: doc.ID}); !errors.Is(err, ErrInvalid) {
					t.Fatalf("empty text %v", err)
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
