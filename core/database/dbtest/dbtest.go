// Package dbtest holds shared test helpers for the catalog database, so query
// tests do not each spell the same boilerplate.
package dbtest

import (
	"database/sql"
	"testing"
)

// QueryPlan runs EXPLAIN QUERY PLAN on query with args and returns the detail
// line of each plan row. It fails the test on error.
func QueryPlan(t testing.TB, db *sql.DB, query string, args ...any) []string {
	t.Helper()
	rows, err := db.Query("EXPLAIN QUERY PLAN "+query, args...)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var details []string
	for rows.Next() {
		var id, parent, notUsed int
		var detail string
		if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
			t.Fatal(err)
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return details
}
