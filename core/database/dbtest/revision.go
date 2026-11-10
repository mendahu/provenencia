package dbtest

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/mendahu/provenencia/core/database/rowchange"
)

// LatestRevision reads the newest audit revision: its action_type and every
// change in write order, with changes_json decoded into Fields.
func LatestRevision(t testing.TB, db *sql.DB) (string, []rowchange.Change) {
	t.Helper()
	var txID []byte
	var action string
	if err := db.QueryRow(`SELECT id, action_type FROM audit_transactions ORDER BY revision DESC LIMIT 1`).Scan(&txID, &action); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT entity_type, entity_id, action, changes_json FROM audit_changes
		WHERE audit_transaction_id = ? ORDER BY rowid`, txID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []rowchange.Change
	for rows.Next() {
		var ch rowchange.Change
		var raw string
		if err := rows.Scan(&ch.EntityType, &ch.EntityID, &ch.Action, &raw); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal([]byte(raw), &ch.Fields); err != nil {
			t.Fatal(err)
		}
		out = append(out, ch)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return action, out
}

// Count runs a single-integer query (SELECT COUNT(*) …) on a *sql.DB or *sql.Tx.
func Count(t testing.TB, q interface {
	QueryRow(string, ...any) *sql.Row
}, query string, args ...any) int {
	t.Helper()
	var n int
	if err := q.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// EntityTypes lists the entity_type of each change, in order.
func EntityTypes(changes []rowchange.Change) []string {
	types := make([]string, len(changes))
	for i, ch := range changes {
		types[i] = ch.EntityType
	}
	return types
}

// MustDB opens c's *sql.DB, failing the test on error.
func MustDB(t testing.TB, c interface{ DB() (*sql.DB, error) }) *sql.DB {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	return db
}
