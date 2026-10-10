// Package sourcecredibility accesses source_credibility_assessments with audited upserts.
package sourcecredibility

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeSourceCredibilityInvalid, apperr.KindUser)

const (
	sqlInsert = `INSERT INTO source_credibility_assessments (id, source_id, credibility_grade_id, argument)
		VALUES (?, ?, ?, ?)`
	sqlUpdate = `UPDATE source_credibility_assessments
		SET credibility_grade_id = ?, argument = ?
		WHERE id = ?`
	sqlGetBySource = `SELECT id, source_id, credibility_grade_id, COALESCE(argument, '')
		FROM source_credibility_assessments WHERE source_id = ?`
	sqlSourceExists = `SELECT 1 FROM sources WHERE id = ?`
)

// Assessment is one source_credibility_assessments row.
type Assessment struct {
	ID                 []byte
	SourceID           []byte
	CredibilityGradeID []byte
	Argument           string
}

// UpsertInput is the mutable fields for an assessment upsert.
type UpsertInput struct {
	SourceID           []byte
	CredibilityGradeID []byte
	Argument           string
}

// GetBySource returns the assessment for a Source, or sql.ErrNoRows when none.
func GetBySource(c *database.Catalog, sourceID []byte) (Assessment, error) {
	db, err := c.DB()
	if err != nil {
		return Assessment{}, err
	}
	if len(sourceID) != 16 {
		return Assessment{}, ErrInvalid
	}
	return scanAssessment(db.QueryRow(sqlGetBySource, sourceID))
}

// Upsert inserts or updates the single assessment row for a Source. The caller
// records create_source_credibility_assessment or update_source_credibility_assessment.
// A grade change recomputes the Source's handles; an argument-only edit and a
// no-op do not. A no-op returns no changes.
func Upsert(tx *database.Tx, userID []byte, in UpsertInput) (Assessment, []rowchange.Change, error) {
	if tx == nil {
		return Assessment{}, nil, ErrInvalid
	}
	in.Argument = strings.TrimSpace(in.Argument)
	if len(in.SourceID) != 16 || len(in.CredibilityGradeID) != 16 {
		return Assessment{}, nil, ErrInvalid
	}
	if len(userID) != 16 {
		return Assessment{}, nil, ErrInvalid
	}
	q := tx.Tx
	var grade int
	if err := q.QueryRow(`SELECT 1 FROM source_credibility_grades WHERE id = ?`, in.CredibilityGradeID).Scan(&grade); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Assessment{}, nil, ErrInvalid
		}
		return Assessment{}, nil, err
	}
	if err := requireSource(q, in.SourceID); err != nil {
		return Assessment{}, nil, err
	}

	prev, err := scanAssessment(q.QueryRow(sqlGetBySource, in.SourceID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Assessment{}, nil, err
	}
	creating := errors.Is(err, sql.ErrNoRows)

	var id []byte
	var action string
	fields := map[string]rowchange.FieldDiff{}

	if creating {
		uid, err := uuid.NewV7()
		if err != nil {
			return Assessment{}, nil, err
		}
		id = uid[:]
		if _, err := q.Exec(sqlInsert, id, in.SourceID, in.CredibilityGradeID, nullStr(in.Argument)); err != nil {
			return Assessment{}, nil, mapConstraint(err)
		}
		action = rowchange.ActionCreate
		fields["id"] = rowchange.FieldDiff{Old: nil, New: uid.String()}
		fields["source_id"] = rowchange.FieldDiff{Old: nil, New: uuidString(in.SourceID)}
		fields["credibility_grade_id"] = rowchange.FieldDiff{Old: nil, New: uuidString(in.CredibilityGradeID)}
		if in.Argument != "" {
			fields["argument"] = rowchange.FieldDiff{Old: nil, New: in.Argument}
		}
	} else {
		id = prev.ID
		gradeChanged := !bytes.Equal(prev.CredibilityGradeID, in.CredibilityGradeID)
		argChanged := prev.Argument != in.Argument
		if !gradeChanged && !argChanged {
			return Assessment{
				ID:                 append([]byte(nil), prev.ID...),
				SourceID:           append([]byte(nil), prev.SourceID...),
				CredibilityGradeID: append([]byte(nil), prev.CredibilityGradeID...),
				Argument:           prev.Argument,
			}, nil, nil
		}
		if _, err := q.Exec(sqlUpdate, in.CredibilityGradeID, nullStr(in.Argument), id); err != nil {
			return Assessment{}, nil, mapConstraint(err)
		}
		action = rowchange.ActionUpdate
		if gradeChanged {
			fields["credibility_grade_id"] = rowchange.FieldDiff{
				Old: uuidString(prev.CredibilityGradeID),
				New: uuidString(in.CredibilityGradeID),
			}
		}
		if argChanged {
			fields["argument"] = rowchange.FieldDiff{
				Old: nullJSON(prev.Argument),
				New: nullJSON(in.Argument),
			}
		}
	}

	return Assessment{
		ID:                 append([]byte(nil), id...),
		SourceID:           append([]byte(nil), in.SourceID...),
		CredibilityGradeID: append([]byte(nil), in.CredibilityGradeID...),
		Argument:           in.Argument,
	}, []rowchange.Change{{
		EntityType: "source_credibility_assessment",
		EntityID:   append([]byte(nil), id...),
		Action:     action,
		Fields:     fields,
	}}, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAssessment(row rowScanner) (Assessment, error) {
	var a Assessment
	if err := row.Scan(&a.ID, &a.SourceID, &a.CredibilityGradeID, &a.Argument); err != nil {
		return Assessment{}, err
	}
	return a, nil
}

func requireSource(tx *sql.Tx, sourceID []byte) error {
	var one int
	err := tx.QueryRow(sqlSourceExists, sourceID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullJSON(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func uuidString(id []byte) string {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
