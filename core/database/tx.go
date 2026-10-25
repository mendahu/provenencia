package database

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/apperr"
)

// ErrWriteReentry is a second writes.Run on a catalog that is already inside
// one, including from AfterCommit or a commit listener.
var ErrWriteReentry = apperr.New(apperr.CodeWritesReentry, apperr.KindInternal)

// Tx is the transaction writes.Run owns. It embeds *sql.Tx so callers can
// query and pass the inner transaction to packages that still take *sql.Tx.
// Run commits and rolls back. A write function does not.
type Tx struct {
	*sql.Tx
	after   []func()
	catalog *Catalog
}

// Querier is the read surface that *sql.DB and *sql.Tx share. A package writes
// one read function over it and exposes a Catalog entry point and a Tx twin, so
// the SQL and scan exist once. A read inside writes.Run must use the Tx: the
// catalog has one connection.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Catalog is the catalog this write began on. It is nil for a Tx built outside
// BeginWrite.
func (t *Tx) Catalog() *Catalog {
	if t == nil {
		return nil
	}
	return t.catalog
}

// AfterCommit registers work that runs only after Run commits. A rollback
// drops it. The callback runs while Run still holds the catalog, so a nested
// Run is refused.
func (t *Tx) AfterCommit(fn func()) {
	if t == nil || fn == nil {
		return
	}
	t.after = append(t.after, fn)
}

// RunAfterCommit runs the registered callbacks. writes.Run calls it after commit.
func (t *Tx) RunAfterCommit() {
	if t == nil {
		return
	}
	for _, fn := range t.after {
		fn()
	}
}

// BeginWrite starts the transaction Run owns and marks the catalog busy.
// Only writes.Run calls this.
func (c *Catalog) BeginWrite() (*Tx, error) {
	if c == nil {
		return nil, ErrWriteReentry
	}
	c.mu.Lock()
	if c.writing {
		c.mu.Unlock()
		return nil, ErrWriteReentry
	}
	c.writing = true
	c.mu.Unlock()

	db, err := c.DB()
	if err != nil {
		c.EndWrite()
		return nil, err
	}
	sqlTx, err := db.Begin()
	if err != nil {
		c.EndWrite()
		return nil, err
	}
	return &Tx{Tx: sqlTx, catalog: c}, nil
}

// EndWrite clears the busy flag. Run defers it so AfterCommit and listeners
// still count as inside the write.
func (c *Catalog) EndWrite() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.writing = false
	c.mu.Unlock()
}
