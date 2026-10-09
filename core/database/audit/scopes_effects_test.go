package audit_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/rowchange"
)

// seenEffectsRevisions are audit transactions already compared to effects.Sources.
// Each id is checked once, while the rows that write left behind still match
// what Record resolved. A later delete must not replay an earlier update whose
// foreign key was only on the live row.
var seenEffectsRevisions = map[string]struct{}{}

func checkNewEffects(t *testing.T, c *database.Catalog) {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.Query(`SELECT id, revision FROM audit_transactions ORDER BY revision`)
	if err != nil {
		t.Fatal(err)
	}
	type rev struct {
		id       []byte
		revision int64
	}
	var revs []rev
	for rows.Next() {
		var r rev
		if err := rows.Scan(&r.id, &r.revision); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		r.id = append([]byte(nil), r.id...)
		revs = append(revs, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatal(err)
	}
	rows.Close()

	for _, r := range revs {
		if _, ok := seenEffectsRevisions[string(r.id)]; ok {
			continue
		}
		changes, err := loadRevisionChanges(tx, r.id)
		if err != nil {
			t.Fatalf("revision %d: %v", r.revision, err)
		}
		want, err := loadSourceScopes(tx, r.id)
		if err != nil {
			t.Fatalf("revision %d: %v", r.revision, err)
		}
		got, err := effects.Sources(tx, changes)
		if err != nil {
			t.Fatalf("revision %d: %v", r.revision, err)
		}
		legacy, err := audit.LegacySourceIDs(tx, changes)
		if err != nil {
			t.Fatalf("revision %d: %v", r.revision, err)
		}
		if !sameIDSet(got, legacy) {
			t.Fatalf("revision %d: effects.Sources = %d ids, old resolvers = %d", r.revision, len(got), len(legacy))
		}
		if !sameIDSet(got, want) {
			t.Fatalf("revision %d: effects.Sources = %d ids, stored source scopes = %d", r.revision, len(got), len(want))
		}
		seenEffectsRevisions[string(r.id)] = struct{}{}
	}
}

func loadRevisionChanges(tx *sql.Tx, id []byte) ([]rowchange.Change, error) {
	rows, err := tx.Query(`SELECT entity_type, entity_id, action, changes_json
		FROM audit_changes WHERE audit_transaction_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []rowchange.Change
	for rows.Next() {
		var ch rowchange.Change
		var raw string
		if err := rows.Scan(&ch.EntityType, &ch.EntityID, &ch.Action, &raw); err != nil {
			return nil, err
		}
		ch.EntityID = append([]byte(nil), ch.EntityID...)
		if err := json.Unmarshal([]byte(raw), &ch.Fields); err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

func loadSourceScopes(tx *sql.Tx, id []byte) ([][]byte, error) {
	rows, err := tx.Query(`SELECT scope_id FROM audit_transaction_scopes
		WHERE audit_transaction_id = ? AND scope_type = ?`, id, audit.ScopeSource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var scopeID []byte
		if err := rows.Scan(&scopeID); err != nil {
			return nil, err
		}
		out = append(out, append([]byte(nil), scopeID...))
	}
	return out, rows.Err()
}

func sameIDSet(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for _, id := range a {
		found := false
		for _, other := range b {
			if len(id) == len(other) && string(id) == string(other) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
