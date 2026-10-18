package promote

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Unpromote is the inverse of Save for one primary Subject: it deletes the
// Subject's accepted Identity Claim and, in the same transaction, the accepted
// claims of the bridges on its Source that have it as an end, each with its
// pins. It does not reject a claim and deletes no handle; other members'
// claims and their pins stay. A Subject with no accepted claim returns no
// changes and no error (a stale second click). The caller records the changes
// (action unpromote_subject); writes.Run recomputes the affected handles.
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
	subject, err := subjects.GetTx(q, subjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	st, err := subjecttypes.GetByIDTx(q, subject.SubjectTypeID)
	if err != nil {
		return nil, err
	}
	if !vocabkey.IsPrimaryType(st.Key, st.Origin) {
		return nil, ErrUnsupportedType
	}
	claim, err := identityclaims.AcceptedEntityForSubjectTx(q, subject.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	released, err := deleteimpact.ReleaseClaim(tx, claim.ID)
	if err != nil {
		return nil, err
	}
	bridges, err := identityclaims.UnfileBridges(tx, subject.SourceID, subject.ID)
	if err != nil {
		return nil, err
	}
	return append(released, bridges...), nil
}
