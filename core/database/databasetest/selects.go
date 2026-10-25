//go:build cgo

// Package databasetest holds test helpers that watch a catalog's SQLite
// connection.
package databasetest

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/mattn/go-sqlite3"

	"github.com/mendahu/provenencia/core/database"
)

// Selects counts the SELECT statements a catalog prepares.
type Selects struct{ n atomic.Int64 }

// N is the number of SELECTs prepared since CountSelects or the last Reset.
func (s *Selects) N() int { return int(s.n.Load()) }

// Reset sets the count to 0.
func (s *Selects) Reset() { s.n.Store(0) }

// CountSelects counts every SELECT the catalog prepares from now until the test
// ends, whoever issues it: the graph, a query package, or a raw *sql.DB call.
// It relies on the catalog holding exactly one connection
// (SetMaxOpenConns(1) in core/database/catalog.go), so the authorizer it
// registers sees every statement.
func CountSelects(t testing.TB, c *database.Catalog) *Selects {
	t.Helper()
	s := &Selects{}
	set := func(cb func(int, string, string, string) int) {
		t.Helper()
		db, err := c.DB()
		if err != nil {
			t.Fatal(err)
		}
		conn, err := db.Conn(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		// Close the *sql.Conn before returning: an open one blocks every
		// later query on the single connection. The authorizer stays on the
		// driver connection.
		defer conn.Close()
		if err := conn.Raw(func(driverConn any) error {
			driverConn.(*sqlite3.SQLiteConn).RegisterAuthorizer(cb)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	set(func(op int, _, _, _ string) int {
		if op == sqlite3.SQLITE_SELECT {
			s.n.Add(1)
		}
		return sqlite3.SQLITE_OK
	})
	t.Cleanup(func() {
		if db, err := c.DB(); err == nil && db != nil {
			set(nil)
		}
	})
	return s
}
