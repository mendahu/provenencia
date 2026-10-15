package audit

import (
	"database/sql"
	"testing"

	"github.com/mendahu/provenencia/core/database/effects"
)

// recordResolved is what writes.Run does: resolve the change list once, then
// record it. A change list effects cannot resolve (an unknown entity type)
// goes to Record with an empty Set so Record's own validation answers.
func recordResolved(t *testing.T, tx *sql.Tx, rev Revision) (int64, error) {
	t.Helper()
	var set effects.Set
	if tx != nil {
		if s, err := effects.Resolve(tx, rev.Changes); err == nil {
			set = s
		}
	}
	return Record(tx, rev, set)
}
