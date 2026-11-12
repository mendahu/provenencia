//go:build cgo

// Package databasetest holds catalog test helpers: watching a catalog's SQLite
// connection and timing work against a budget.
package databasetest

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mattn/go-sqlite3"

	"github.com/mendahu/provenencia/core/database"
)

// Selects counts the SELECT statements a catalog prepares and remembers which
// tables they read.
type Selects struct {
	n atomic.Int64

	mu    sync.Mutex
	reads map[string]bool
}

// N is the number of SELECTs prepared since CountSelects or the last Reset.
func (s *Selects) N() int { return int(s.n.Load()) }

// Read reports whether any statement since CountSelects or the last Reset read
// a column of the table.
func (s *Selects) Read(table string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reads[table]
}

// Reset sets the count to 0 and forgets the tables read.
func (s *Selects) Reset() {
	s.n.Store(0)
	s.mu.Lock()
	s.reads = nil
	s.mu.Unlock()
}

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
	set(func(op int, table, _, _ string) int {
		switch op {
		case sqlite3.SQLITE_SELECT:
			s.n.Add(1)
		case sqlite3.SQLITE_READ:
			s.mu.Lock()
			if s.reads == nil {
				s.reads = map[string]bool{}
			}
			s.reads[table] = true
			s.mu.Unlock()
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
