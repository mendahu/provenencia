package audit

import (
	"database/sql"
	"testing"

	"github.com/mendahu/provenencia/core/database/effects"
)

// recordResolved is what writes.Run does: resolve the change list once, then
// record it. A change list effects cannot resolve fails the test; a test of
// Record's own validation of an entity type effects does not know calls Record
// with an empty Set instead.
func recordResolved(t *testing.T, tx *sql.Tx, rev Revision) (int64, error) {
	t.Helper()
	var set effects.Set
	if tx != nil {
		var err error
		set, err = effects.Resolve(tx, rev.Changes)
		if err != nil {
			t.Fatalf("effects.Resolve: %v", err)
		}
	}
	return Record(tx, rev, set)
}
