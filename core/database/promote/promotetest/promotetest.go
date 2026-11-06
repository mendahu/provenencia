// Package promotetest holds the shared test helpers that promote or claim a
// Subject in a write, so tests do not each spell the writes.Call block.
package promotetest

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
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

// ClaimWithoutFiling claims s onto a new handle without filing bridges, as a
// catalog written before S14-33 left a bridge drawn between two handles. It
// returns the handle id and fails the test on error.
func ClaimWithoutFiling(t testing.TB, c *database.Catalog, userID []byte, s subjects.Subject) []byte {
	t.Helper()
	id, err := writes.Call(c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) ([]byte, []rowchange.Change, error) {
		entity, entityChange, err := canonicalentities.InsertTx(tx.Tx, canonicalentities.CreateInput{SubjectTypeID: s.SubjectTypeID})
		if err != nil {
			return nil, nil, err
		}
		_, claimChange, err := identityclaims.InsertTx(tx.Tx, identityclaims.CreateInput{
			SubjectID: s.ID, EntityID: entity.ID, Status: identityclaims.StatusAccepted,
		})
		if err != nil {
			return nil, nil, err
		}
		return entity.ID, []rowchange.Change{entityChange, claimChange}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}
