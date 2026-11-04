package promote

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database/subjects"
)

// loadPrimary loads a Subject and refuses non-primary types (associations).
func loadPrimary(q *sql.Tx, subjectID []byte) (subjects.Subject, error) {
	subject, err := subjects.GetTx(q, subjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return subjects.Subject{}, ErrInvalid
	}
	if err != nil {
		return subjects.Subject{}, err
	}
	if err := requirePrimary(q, subject); err != nil {
		return subjects.Subject{}, err
	}
	return subject, nil
}
