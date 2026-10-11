// Package sqlbatch reads a list of ids of any length in fixed-size IN lists.
// SQLite binds at most 32,766 values per statement, so every by-id read goes
// through here. It imports nothing from the catalog, so the graph cache, which
// the database package depends on, can use it too.
package sqlbatch

import (
	"database/sql"
	"strings"
)

// Size bounds one IN list. A batch stays far under SQLite's bind limit, with
// room for a query's other arguments, so a read of any size is one query per
// batch.
const Size = 500

// Each calls fn with consecutive slices of ids, each at most Size long. It
// stops at fn's first error.
func Each(ids [][]byte, fn func(batch [][]byte) error) error {
	for start := 0; start < len(ids); start += Size {
		if err := fn(ids[start:min(start+Size, len(ids))]); err != nil {
			return err
		}
	}
	return nil
}

// Querier runs a query: *sql.DB, *sql.Tx, or a wrapper around either.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Query runs head + placeholders + tail once per batch of ids, binding args
// first and the batch after them, and calls scan for each row. head ends
// inside the IN list ("… IN (") and tail closes it (")", or ") ORDER BY …").
// Rows arrive batch by batch, so a caller that needs one order across
// batches sorts afterwards.
func Query(q Querier, head, tail string, args []any, ids [][]byte, scan func(rows *sql.Rows) error) error {
	return Each(ids, func(batch [][]byte) error {
		bound := make([]any, 0, len(args)+len(batch))
		bound = append(bound, args...)
		for _, id := range batch {
			bound = append(bound, id)
		}
		rows, err := q.Query(head+strings.Repeat("?,", len(batch)-1)+"?"+tail, bound...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			if err := scan(rows); err != nil {
				return err
			}
		}
		return rows.Err()
	})
}
