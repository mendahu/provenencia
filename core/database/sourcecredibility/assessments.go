// Package sourcecredibility accesses source_credibility_assessments with audited upserts.
package sourcecredibility

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/sourcecredibilitygrades"
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

// Upsert inserts or updates the single assessment row for a Source and audits
// create_source_credibility_assessment or update_source_credibility_assessment.
func Upsert(c *database.Catalog, userID []byte, in UpsertInput) (Assessment, error) {
	db, err := c.DB()
	if err != nil {
		return Assessment{}, err
	}
	in.Argument = strings.TrimSpace(in.Argument)
	if len(in.SourceID) != 16 || len(in.CredibilityGradeID) != 16 {
		return Assessment{}, ErrInvalid
	}
	if len(userID) != 16 {
		return Assessment{}, ErrInvalid
	}
	if _, err := sourcecredibilitygrades.GetByID(c, in.CredibilityGradeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Assessment{}, ErrInvalid
		}
		return Assessment{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Assessment{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := requireSource(tx, in.SourceID); err != nil {
		return Assessment{}, err
	}

	prev, err := scanAssessment(tx.QueryRow(sqlGetBySource, in.SourceID))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Assessment{}, err
	}
	creating := errors.Is(err, sql.ErrNoRows)

	var id []byte
	var actionType string
	var action string
	fields := map[string]audit.FieldDiff{}

	if creating {
		uid, err := uuid.NewV7()
		if err != nil {
			return Assessment{}, err
		}
		id = uid[:]
		if _, err := tx.Exec(sqlInsert, id, in.SourceID, in.CredibilityGradeID, nullStr(in.Argument)); err != nil {
			return Assessment{}, mapConstraint(err)
		}
		actionType = "create_source_credibility_assessment"
		action = audit.ActionCreate
		fields["id"] = audit.FieldDiff{Old: nil, New: uid.String()}
		fields["source_id"] = audit.FieldDiff{Old: nil, New: uuidString(in.SourceID)}
		fields["credibility_grade_id"] = audit.FieldDiff{Old: nil, New: uuidString(in.CredibilityGradeID)}
		if in.Argument != "" {
			fields["argument"] = audit.FieldDiff{Old: nil, New: in.Argument}
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
			}, tx.Commit()
		}
		if _, err := tx.Exec(sqlUpdate, in.CredibilityGradeID, nullStr(in.Argument), id); err != nil {
			return Assessment{}, mapConstraint(err)
		}
		actionType = "update_source_credibility_assessment"
		action = audit.ActionUpdate
		if gradeChanged {
			fields["credibility_grade_id"] = audit.FieldDiff{
				Old: uuidString(prev.CredibilityGradeID),
				New: uuidString(in.CredibilityGradeID),
			}
		}
		if argChanged {
			fields["argument"] = audit.FieldDiff{
				Old: nullJSON(prev.Argument),
				New: nullJSON(in.Argument),
			}
		}
	}

	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: actionType,
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "source_credibility_assessment",
			EntityID:   id,
			Action:     action,
			Fields:     fields,
		}},
	}); err != nil {
		return Assessment{}, err
	}
	if err := tx.Commit(); err != nil {
		return Assessment{}, err
	}
	return Assessment{
		ID:                 append([]byte(nil), id...),
		SourceID:           append([]byte(nil), in.SourceID...),
		CredibilityGradeID: append([]byte(nil), in.CredibilityGradeID...),
		Argument:           in.Argument,
	}, nil
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
