package audit

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

func TestRecord(t *testing.T) {
	entityID := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	userID := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	fixedAt := "2026-09-06T21:00:00Z"

	mustUser := func(t *testing.T, c *database.Catalog, id []byte) {
		t.Helper()
		r, err := ref.Mint(ref.PrefixUser)
		if err != nil {
			t.Fatal(err)
		}
		if err := users.Upsert(c, id, "Jake", r); err != nil {
			t.Fatal(err)
		}
	}

	begin := func(t *testing.T, c *database.Catalog) *sql.Tx {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		return tx
	}

	loadChangeJSON := func(t *testing.T, c *database.Catalog) (action string, raw string, txID []byte) {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		err = db.QueryRow(`SELECT action, changes_json, audit_transaction_id FROM audit_changes`).Scan(&action, &raw, &txID)
		if err != nil {
			t.Fatal(err)
		}
		return action, raw, txID
	}

	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create JSON old null new values",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c, userID)
				tx := begin(t, c)
				rev, err := Record(tx, Revision{
					UserID:     userID,
					ActionType: "create_source",
					CreatedAt:  fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   entityID,
						Action:     ActionCreate,
						Fields: map[string]FieldDiff{
							"id":    {Old: nil, New: "aaaaaaaa-bbbb-7ccc-8ddd-eeeeeeeeeeee"},
							"title": {Old: nil, New: "The Robins Family History"},
						},
					}},
				})
				if err != nil {
					t.Fatal(err)
				}
				if rev != 1 {
					t.Fatalf("revision %d want 1", rev)
				}
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
				action, raw, _ := loadChangeJSON(t, c)
				if action != ActionCreate {
					t.Fatalf("action %q", action)
				}
				var fields map[string]FieldDiff
				if err := json.Unmarshal([]byte(raw), &fields); err != nil {
					t.Fatal(err)
				}
				if fields["title"].Old != nil || fields["title"].New != "The Robins Family History" {
					t.Fatalf("fields %+v", fields)
				}
				if fields["id"].Old != nil {
					t.Fatalf("id.old want null got %#v", fields["id"].Old)
				}
			},
		},
		{
			name: "update JSON only changed fields",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c, userID)
				tx := begin(t, c)
				if _, err := Record(tx, Revision{
					UserID:     userID,
					ActionType: "update_source",
					CreatedAt:  fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   entityID,
						Action:     ActionUpdate,
						Fields: map[string]FieldDiff{
							"title": {Old: "Robins Family History", New: "The Robins Family History"},
						},
					}},
				}); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
				_, raw, _ := loadChangeJSON(t, c)
				var fields map[string]FieldDiff
				if err := json.Unmarshal([]byte(raw), &fields); err != nil {
					t.Fatal(err)
				}
				if len(fields) != 1 {
					t.Fatalf("want 1 field got %+v", fields)
				}
				if fields["title"].Old != "Robins Family History" || fields["title"].New != "The Robins Family History" {
					t.Fatalf("fields %+v", fields)
				}
			},
		},
		{
			name: "delete JSON new null",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c, userID)
				tx := begin(t, c)
				if _, err := Record(tx, Revision{
					UserID:     userID,
					ActionType: "delete_source_note",
					CreatedAt:  fixedAt,
					Changes: []Change{{
						EntityType: "source_note",
						EntityID:   entityID,
						Action:     ActionDelete,
						Fields: map[string]FieldDiff{
							"id":   {Old: "aaaaaaaa-bbbb-7ccc-8ddd-eeeeeeeeeeee", New: nil},
							"body": {Old: "Interview notes", New: nil},
						},
					}},
				}); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
				action, raw, _ := loadChangeJSON(t, c)
				if action != ActionDelete {
					t.Fatalf("action %q", action)
				}
				var fields map[string]FieldDiff
				if err := json.Unmarshal([]byte(raw), &fields); err != nil {
					t.Fatal(err)
				}
				if fields["body"].Old != "Interview notes" || fields["body"].New != nil {
					t.Fatalf("fields %+v", fields)
				}
			},
		},
		{
			name: "monotonic revision and shared transaction id",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c, userID)
				otherID := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}

				tx1 := begin(t, c)
				r1, err := Record(tx1, Revision{
					UserID:     userID,
					ActionType: "create_source",
					CreatedAt:  fixedAt,
					Changes: []Change{
						{
							EntityType: "source",
							EntityID:   entityID,
							Action:     ActionCreate,
							Fields:     map[string]FieldDiff{"title": {Old: nil, New: "A"}},
						},
						{
							EntityType: "source_metadata",
							EntityID:   otherID,
							Action:     ActionCreate,
							Fields:     map[string]FieldDiff{"value_text": {Old: nil, New: "x"}},
						},
					},
				})
				if err != nil {
					t.Fatal(err)
				}
				if r1 != 1 {
					t.Fatalf("r1=%d", r1)
				}
				if err := tx1.Commit(); err != nil {
					t.Fatal(err)
				}

				tx2 := begin(t, c)
				r2, err := Record(tx2, Revision{
					UserID:     userID,
					ActionType: "update_source",
					CreatedAt:  fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   entityID,
						Action:     ActionUpdate,
						Fields:     map[string]FieldDiff{"title": {Old: "A", New: "B"}},
					}},
				})
				if err != nil {
					t.Fatal(err)
				}
				if r2 != 2 {
					t.Fatalf("r2=%d", r2)
				}
				if err := tx2.Commit(); err != nil {
					t.Fatal(err)
				}

				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var nTx, nChanges int
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&nTx); err != nil {
					t.Fatal(err)
				}
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_changes`).Scan(&nChanges); err != nil {
					t.Fatal(err)
				}
				if nTx != 2 || nChanges != 3 {
					t.Fatalf("tx=%d changes=%d", nTx, nChanges)
				}
				var txIDA, txIDB []byte
				rows, err := db.Query(`SELECT audit_transaction_id FROM audit_changes WHERE entity_type IN ('source','source_metadata') AND action = 'create' ORDER BY entity_type`)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				ids := make([][]byte, 0, 2)
				for rows.Next() {
					var id []byte
					if err := rows.Scan(&id); err != nil {
						t.Fatal(err)
					}
					ids = append(ids, id)
				}
				if err := rows.Err(); err != nil {
					t.Fatal(err)
				}
				if len(ids) != 2 {
					t.Fatalf("create changes %d", len(ids))
				}
				txIDA, txIDB = ids[0], ids[1]
				if string(txIDA) != string(txIDB) {
					t.Fatal("create changes should share audit_transaction_id")
				}
			},
		},
		{
			name: "null user_id allowed",
			run: func(t *testing.T, c *database.Catalog) {
				tx := begin(t, c)
				if _, err := Record(tx, Revision{
					ActionType: "system_seed",
					CreatedAt:  fixedAt,
					Changes: []Change{{
						EntityType: "source_type",
						EntityID:   entityID,
						Action:     ActionCreate,
						Fields:     map[string]FieldDiff{"key": {Old: nil, New: "photograph"}},
					}},
				}); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var got []byte
				if err := db.QueryRow(`SELECT user_id FROM audit_transactions`).Scan(&got); err != nil {
					t.Fatal(err)
				}
				if got != nil {
					t.Fatalf("user_id want NULL got %v", got)
				}
			},
		},
		{
			name: "records user_id when provided",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c, userID)
				tx := begin(t, c)
				if _, err := Record(tx, Revision{
					UserID:     userID,
					ActionType: "create_source",
					CreatedAt:  fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   entityID,
						Action:     ActionCreate,
						Fields:     map[string]FieldDiff{"title": {Old: nil, New: "A"}},
					}},
				}); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var got []byte
				if err := db.QueryRow(`SELECT user_id FROM audit_transactions`).Scan(&got); err != nil {
					t.Fatal(err)
				}
				if string(got) != string(userID) {
					t.Fatalf("user_id mismatch")
				}
			},
		},
		{
			name: "rollback drops audit rows",
			run: func(t *testing.T, c *database.Catalog) {
				mustUser(t, c, userID)
				tx := begin(t, c)
				if _, err := Record(tx, Revision{
					UserID:     userID,
					ActionType: "create_source",
					CreatedAt:  fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   entityID,
						Action:     ActionCreate,
						Fields:     map[string]FieldDiff{"title": {Old: nil, New: "A"}},
					}},
				}); err != nil {
					t.Fatal(err)
				}
				if err := tx.Rollback(); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var n int
				if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&n); err != nil {
					t.Fatal(err)
				}
				if n != 0 {
					t.Fatalf("transactions after rollback: %d", n)
				}
			},
		},
		{
			name: "rejects bad action",
			run: func(t *testing.T, c *database.Catalog) {
				tx := begin(t, c)
				defer tx.Rollback()
				_, err := Record(tx, Revision{
					CreatedAt: fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   entityID,
						Action:     "merge",
						Fields:     map[string]FieldDiff{},
					}},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects short entity id",
			run: func(t *testing.T, c *database.Catalog) {
				tx := begin(t, c)
				defer tx.Rollback()
				_, err := Record(tx, Revision{
					CreatedAt: fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   []byte{1},
						Action:     ActionCreate,
						Fields:     map[string]FieldDiff{},
					}},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects empty entity type",
			run: func(t *testing.T, c *database.Catalog) {
				tx := begin(t, c)
				defer tx.Rollback()
				_, err := Record(tx, Revision{
					CreatedAt: fixedAt,
					Changes: []Change{{
						EntityType: "  ",
						EntityID:   entityID,
						Action:     ActionCreate,
						Fields:     map[string]FieldDiff{},
					}},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects nil fields map",
			run: func(t *testing.T, c *database.Catalog) {
				tx := begin(t, c)
				defer tx.Rollback()
				_, err := Record(tx, Revision{
					CreatedAt: fixedAt,
					Changes: []Change{{
						EntityType: "source",
						EntityID:   entityID,
						Action:     ActionCreate,
						Fields:     nil,
					}},
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects blank created_at",
			run: func(t *testing.T, c *database.Catalog) {
				tx := begin(t, c)
				defer tx.Rollback()
				_, err := Record(tx, Revision{
					CreatedAt: "  ",
					Changes:   nil,
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects short user id",
			run: func(t *testing.T, c *database.Catalog) {
				tx := begin(t, c)
				defer tx.Rollback()
				_, err := Record(tx, Revision{
					UserID:    []byte{1, 2},
					CreatedAt: fixedAt,
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "rejects nil tx",
			run: func(t *testing.T, c *database.Catalog) {
				_, err := Record(nil, Revision{CreatedAt: fixedAt})
				if !errors.Is(err, ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "create migrates audit tables",
			run: func(t *testing.T, c *database.Catalog) {
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var ver int
				if err := db.QueryRow(`PRAGMA user_version`).Scan(&ver); err != nil {
					t.Fatal(err)
				}
				if ver < 3 {
					t.Fatalf("user_version %d want >= 3", ver)
				}
				for _, table := range []string{"audit_transactions", "audit_changes"} {
					var name string
					err := db.QueryRow(
						`SELECT name FROM sqlite_master WHERE type='table' AND name=?`,
						table,
					).Scan(&name)
					if err != nil {
						t.Fatalf("%s: %v", table, err)
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
