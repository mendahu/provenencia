package identityclaims

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database/audit"
)

// Pins and claims are removed explicitly so each removal is audited. The
// schema's ON DELETE CASCADE (pins → claim, pins → Observation, claim →
// Subject) is only a backstop for writers that forget.

const (
	sqlPinsByObservation = `SELECT identity_claim_id, observation_id FROM identity_claim_evidence
		WHERE observation_id = ? ORDER BY identity_claim_id`
	sqlPinsBySubjectObservations = `SELECT ev.identity_claim_id, ev.observation_id FROM identity_claim_evidence ev
		JOIN observations o ON o.id = ev.observation_id
		WHERE o.subject_id = ? ORDER BY ev.identity_claim_id, ev.observation_id`
	sqlPinsByClaim = `SELECT identity_claim_id, observation_id FROM identity_claim_evidence
		WHERE identity_claim_id = ? ORDER BY observation_id`
	sqlClaimsBySubject = `SELECT ` + sqlColumns + ` FROM identity_claims WHERE subject_id = ? ORDER BY id`
	sqlDeletePin       = `DELETE FROM identity_claim_evidence WHERE identity_claim_id = ? AND observation_id = ?`
	sqlDeleteClaim     = `DELETE FROM identity_claims WHERE id = ?`
)

type pin struct {
	claimID, observationID []byte
}

// ReleaseObservationPinsTx removes every pin on an Observation about to be
// deleted. The claims that pinned it stay, with a weaker exhibit (§5.2).
func ReleaseObservationPinsTx(tx *sql.Tx, observationID []byte) ([]audit.Change, error) {
	if len(observationID) != 16 {
		return nil, ErrInvalid
	}
	pins, err := queryPins(tx, sqlPinsByObservation, observationID)
	if err != nil {
		return nil, err
	}
	return deletePins(tx, pins)
}

// ReleaseSubjectTx removes everything Conclusion holds on a Subject about to be
// deleted: pins on its Observations (on any claim), then its own claims with
// their remaining pins. The handles stay. A Subject delete is refused while it
// owns non-connection Observations, so the first step only ever finds pins on
// bridge edge Observations (reachable once bridges are filed, S9-28).
func ReleaseSubjectTx(tx *sql.Tx, subjectID []byte) ([]audit.Change, error) {
	if len(subjectID) != 16 {
		return nil, ErrInvalid
	}
	pins, err := queryPins(tx, sqlPinsBySubjectObservations, subjectID)
	if err != nil {
		return nil, err
	}
	changes, err := deletePins(tx, pins)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(sqlClaimsBySubject, subjectID)
	if err != nil {
		return nil, err
	}
	claims, err := scanClaims(rows)
	if err != nil {
		return nil, err
	}
	for _, cl := range claims {
		own, err := queryPins(tx, sqlPinsByClaim, cl.ID)
		if err != nil {
			return nil, err
		}
		pinChanges, err := deletePins(tx, own)
		if err != nil {
			return nil, err
		}
		changes = append(changes, pinChanges...)
		if _, err := tx.Exec(sqlDeleteClaim, cl.ID); err != nil {
			return nil, err
		}
		changes = append(changes, audit.Change{
			EntityType: "identity_claim",
			EntityID:   cl.ID,
			Action:     audit.ActionDelete,
			Fields: audit.DeletedRow(map[string]any{
				"id":                  uuidJSON(cl.ID),
				"subject_id":          uuidJSON(cl.SubjectID),
				"entity_id":           uuidJSON(cl.EntityID),
				"subject_type_id":     uuidJSON(cl.SubjectTypeID),
				"status":              cl.Status,
				"confidence_grade_id": uuidJSON(cl.ConfidenceGradeID),
				"argument":            nullStr(cl.Argument),
			}),
		})
	}
	return changes, nil
}

func queryPins(tx *sql.Tx, q string, id []byte) ([]pin, error) {
	rows, err := tx.Query(q, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []pin
	for rows.Next() {
		var p pin
		if err := rows.Scan(&p.claimID, &p.observationID); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// deletePins deletes each pin and audits it under its claim's id, so replaying
// (identity_claim_evidence, claim id) yields that claim's exhibit history.
func deletePins(tx *sql.Tx, pins []pin) ([]audit.Change, error) {
	changes := make([]audit.Change, 0, len(pins))
	for _, p := range pins {
		if _, err := tx.Exec(sqlDeletePin, p.claimID, p.observationID); err != nil {
			return nil, err
		}
		changes = append(changes, audit.Change{
			EntityType: "identity_claim_evidence",
			EntityID:   p.claimID,
			Action:     audit.ActionDelete,
			Fields: audit.DeletedRow(map[string]any{
				"identity_claim_id": uuidJSON(p.claimID),
				"observation_id":    uuidJSON(p.observationID),
			}),
		})
	}
	return changes, nil
}
