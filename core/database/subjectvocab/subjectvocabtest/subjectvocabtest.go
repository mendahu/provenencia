// Package subjectvocabtest holds test helpers for binding writes.
package subjectvocabtest

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/writes"
)

// Assign binds a Property to the end of a type's order through writes.Call, as
// the assign RPC does, and fails the test on error.
func Assign(t testing.TB, c *database.Catalog, userID, typeID, propertyID []byte) {
	t.Helper()
	if _, err := writes.Call(c, writes.Op{Action: subjectvocab.ActionAssign, UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			changes, err := subjectvocab.AppendBinding(tx, userID, typeID, propertyID)
			return struct{}{}, changes, err
		}); err != nil {
		t.Fatal(err)
	}
}
