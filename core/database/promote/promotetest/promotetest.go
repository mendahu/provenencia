// Package promotetest holds the shared test helper that promotes a Subject in
// a write, so tests do not each spell the writes.Call block.
package promotetest

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
)

// Save runs promote.Save in one audited write as userID and fails the test on
// error.
func Save(t testing.TB, c *database.Catalog, userID []byte, in promote.Input) promote.Result {
	t.Helper()
	res, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
		return promote.Save(tx, userID, in)
	})
	if err != nil {
		t.Fatal(err)
	}
	return res
}
