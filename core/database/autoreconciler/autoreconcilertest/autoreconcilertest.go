// Package autoreconcilertest holds the shared test check that the
// auto-reconciler cache upkeep left equals a rebuild from truth tables.
package autoreconcilertest

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
)

// AssertCacheMatchesRebuild fails the test when the cache the writes upkept
// differs from autoreconciler.Rebuild's (run on a rolled-back transaction, so
// it changes nothing).
func AssertCacheMatchesRebuild(t testing.TB, c *database.Catalog) {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	before := dump(t, db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := autoreconciler.Rebuild(tx); err != nil {
		t.Fatal(err)
	}
	if after := dump(t, tx); after != before {
		t.Fatalf("upkeep differs from a rebuild\nupkeep:\n%s\nrebuild:\n%s", before, after)
	}
}

func dump(t testing.TB, q interface {
	Query(string, ...any) (*sql.Rows, error)
}) string {
	t.Helper()
	var b strings.Builder
	for _, table := range []string{"auto_reconciler_values", "auto_reconciler_outcomes"} {
		rows, err := q.Query(`SELECT * FROM ` + table + ` ORDER BY 1, 2, 3`)
		if err != nil {
			t.Fatal(err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&b, "%s %v\n", table, vals)
		}
		_ = rows.Close()
	}
	return b.String()
}
