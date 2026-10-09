// Package writes is the transaction around a catalog write: the row changes,
// the audit revision, and the derived data the effects registry names.
package writes

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/searchindex"
)

// Op is the audit action for one Run.
type Op struct {
	Action      string
	Description string
	UserID      []byte
}

// Result is the revision Run recorded and what the effects registry resolved.
// Revision is zero when the closure changed nothing.
type Result struct {
	Revision int64
	Effects  effects.Set
}

// Run begins a transaction, runs fn, and on a non-empty change list records
// audit, recomputes handles, reprojects search documents, commits, runs
// AfterCommit, then notifies listeners. A closure error or an empty change
// list rolls back and does not notify. A second Run on the same catalog,
// including from AfterCommit or a listener, returns database.ErrWriteReentry.
// Run does not take the catalog session lock.
func Run[T any](c *database.Catalog, op Op, fn func(tx *database.Tx) (T, []rowchange.Change, error)) (T, Result, error) {
	var zero T
	tx, err := c.BeginWrite()
	if err != nil {
		return zero, Result{}, err
	}
	defer c.EndWrite()
	defer func() { _ = tx.Rollback() }()

	value, changes, err := fn(tx)
	if err != nil {
		return zero, Result{}, err
	}
	if len(changes) == 0 {
		return value, Result{}, nil
	}

	rev, err := audit.Record(tx.Tx, audit.Revision{
		UserID:      op.UserID,
		ActionType:  op.Action,
		Description: op.Description,
		CreatedAt:   project.NowUTC(),
		Changes:     changes,
	})
	if err != nil {
		return zero, Result{}, err
	}
	set, err := effects.Resolve(tx.Tx, changes)
	if err != nil {
		return zero, Result{}, err
	}
	if len(set.Handles) > 0 {
		if err := autoreconciler.RecomputeTx(tx.Tx, set.Handles); err != nil {
			return zero, Result{}, err
		}
	}
	if len(set.Search) > 0 {
		if err := searchindex.Reproject(tx.Tx, set.Search); err != nil {
			return zero, Result{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return zero, Result{}, err
	}
	tx.RunAfterCommit()
	if err := c.Notify(rev, set); err != nil {
		return value, Result{Revision: rev, Effects: set}, err
	}
	return value, Result{Revision: rev, Effects: set}, nil
}
