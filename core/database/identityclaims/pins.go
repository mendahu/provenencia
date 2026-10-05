package identityclaims

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database/audit"
)

const (
	sqlPin = `INSERT INTO identity_claim_evidence (identity_claim_id, observation_id)
		VALUES (?, ?) ON CONFLICT DO NOTHING`
	sqlPinnedObservations = `SELECT observation_id FROM identity_claim_evidence
		WHERE identity_claim_id = ? ORDER BY observation_id`
)

// PinTx pins an Observation on a claim's exhibit on an open transaction (no
// commit, no revision). A pin the claim already carries is left alone: ok is
// false and there is no change. A new pin's change is recorded under the
// claim's id, the mirror of deleteimpact's audited release, so replaying
// (identity_claim_evidence, claim id) yields that claim's exhibit history.
func PinTx(tx *sql.Tx, claimID, observationID []byte) (change audit.Change, ok bool, err error) {
	if len(claimID) != 16 || len(observationID) != 16 {
		return audit.Change{}, false, ErrInvalid
	}
	res, err := tx.Exec(sqlPin, claimID, observationID)
	if err != nil {
		return audit.Change{}, false, err
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return audit.Change{}, false, err
	}
	return audit.Change{
		EntityType: "identity_claim_evidence",
		EntityID:   append([]byte(nil), claimID...),
		Action:     audit.ActionCreate,
		Fields: audit.FullRow(map[string]any{
			"identity_claim_id": uuidJSON(claimID),
			"observation_id":    uuidJSON(observationID),
		}),
	}, true, nil
}

// PinnedObservations returns the Observations pinned on a claim, by id.
func PinnedObservations(q interface {
	Query(string, ...any) (*sql.Rows, error)
}, claimID []byte) ([][]byte, error) {
	if len(claimID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := q.Query(sqlPinnedObservations, claimID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
