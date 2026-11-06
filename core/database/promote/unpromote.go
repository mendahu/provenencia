package promote

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
)

// Unpromote is the inverse of Save for one primary Subject: it deletes the
// Subject's accepted Identity Claim and, in the same transaction, the accepted
// claims of the bridges that have it as an end. It does not reject a claim. When
// the claim was a handle's last accepted member, the handle is deleted too.
// A Subject with no accepted claim returns no changes and no error (a stale
// second click). The caller records the changes (action unpromote_subject);
// writes.Run recomputes the affected handles.
func Unpromote(tx *database.Tx, userID, subjectID []byte) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	if len(subjectID) != 16 {
		return nil, ErrInvalid
	}
	q := tx.Tx
	subject, claim, filed, err := loadFiledPrimary(q, subjectID)
	if err != nil || !filed {
		return nil, err
	}
	released, err := deleteimpact.ReleaseClaim(tx, claim.ID)
	if err != nil {
		return nil, err
	}
	bridges, err := identityclaims.UnfileBridges(tx, subject.ID)
	if err != nil {
		return nil, err
	}
	return append(released, bridges...), nil
}

// loadFiledPrimary loads a primary Subject and its accepted Identity Claim.
// filed is false when the Subject has no accepted claim: an Unpromote and its
// preflight have nothing to do then.
func loadFiledPrimary(q *sql.Tx, subjectID []byte) (subjects.Subject, identityclaims.Claim, bool, error) {
	subject, err := loadPrimary(q, subjectID)
	if err != nil {
		return subjects.Subject{}, identityclaims.Claim{}, false, err
	}
	claim, err := identityclaims.AcceptedEntityForSubjectTx(q, subject.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return subject, identityclaims.Claim{}, false, nil
	}
	if err != nil {
		return subjects.Subject{}, identityclaims.Claim{}, false, err
	}
	return subject, claim, true, nil
}
