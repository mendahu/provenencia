// Package identityclaims accesses the identity_claims catalog table: one
// Interpretation Subject is a reading of one canonical handle. Accepted claims
// are the handle's members. Mutations are audited.
package identityclaims

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/project"
)

var (
	ErrInvalid       = apperr.New(apperr.CodeIdentityClaimsInvalid, apperr.KindUser)
	ErrAlreadyMember = apperr.New(apperr.CodeIdentityClaimsAlreadyMember, apperr.KindConflict)
	ErrTypeMismatch  = apperr.New(apperr.CodeIdentityClaimsTypeMismatch, apperr.KindUser)
)

// Claim workflow status. Only accepted claims are members.
const (
	StatusProvisional = "provisional"
	StatusAccepted    = "accepted"
	StatusRejected    = "rejected"
)

const (
	sqlInsert = `INSERT INTO identity_claims
		(id, subject_id, entity_id, subject_type_id, status, confidence_grade_id, argument)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	sqlColumns            = `id, subject_id, entity_id, subject_type_id, status, confidence_grade_id, COALESCE(argument, '')`
	sqlGet                = `SELECT ` + sqlColumns + ` FROM identity_claims WHERE id = ?`
	sqlAcceptedForSubject = `SELECT ` + sqlColumns + ` FROM identity_claims
		WHERE subject_id = ? AND status = 'accepted'`
	sqlAcceptedMembers = `SELECT ` + sqlColumns + ` FROM identity_claims
		WHERE entity_id = ? AND status = 'accepted'
		ORDER BY id`
	sqlSubjectType = `SELECT subject_type_id FROM subjects WHERE id = ?`
	sqlEntityType  = `SELECT subject_type_id FROM canonical_entities WHERE id = ?`
)

// Claim is one identity_claims row.
type Claim struct {
	ID                []byte
	SubjectID         []byte
	EntityID          []byte
	SubjectTypeID     []byte
	Status            string
	ConfidenceGradeID []byte // nil = no grade
	Argument          string
}

// CreateInput is the fields for a new claim. subject_type_id is read from the subject.
type CreateInput struct {
	SubjectID         []byte
	EntityID          []byte
	Status            string
	ConfidenceGradeID []byte
	Argument          string
}

// Create inserts a claim and records create_identity_claim.
func Create(c *database.Catalog, userID []byte, in CreateInput) (Claim, error) {
	db, err := c.DB()
	if err != nil {
		return Claim{}, err
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Claim{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Claim{}, err
	}
	defer func() { _ = tx.Rollback() }()

	cl, change, err := InsertTx(tx, in)
	if err != nil {
		return Claim{}, err
	}
	changes := []audit.Change{change}
	var assocs [][]byte
	if in.Status == StatusAccepted {
		assocs, changes, err = AppendFiling(tx, cl.SubjectID, changes)
		if err != nil {
			return Claim{}, err
		}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_identity_claim",
		CreatedAt:  project.NowUTC(),
		Changes:    changes,
	}); err != nil {
		return Claim{}, err
	}
	if in.Status == StatusAccepted {
		if err := autoreconciler.RecomputeTouchingTx(tx, append([][]byte{cl.EntityID}, assocs...), cl.SubjectID); err != nil {
			return Claim{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return Claim{}, err
	}
	return cl, nil
}

// AppendFiling files bridges the new member completes and returns the
// association handles plus the revision's changes.
func AppendFiling(tx *sql.Tx, subjectID []byte, changes []audit.Change) ([][]byte, []audit.Change, error) {
	assocs, filed, err := FileBridgesTx(tx, subjectID)
	if err != nil {
		return nil, nil, err
	}
	return assocs, append(changes, filed...), nil
}

// InsertTx inserts a claim on an open transaction (no commit, no revision).
// The subject's type is copied onto the claim, so the composite FK rejects a
// handle of another type (ErrTypeMismatch). A second accepted claim for one
// subject is ErrAlreadyMember.
func InsertTx(tx *sql.Tx, in CreateInput) (Claim, audit.Change, error) {
	in.Status = strings.TrimSpace(in.Status)
	in.Argument = strings.TrimSpace(in.Argument)
	if len(in.SubjectID) != 16 || len(in.EntityID) != 16 || !statusOK(in.Status) {
		return Claim{}, audit.Change{}, ErrInvalid
	}
	if len(in.ConfidenceGradeID) != 0 && len(in.ConfidenceGradeID) != 16 {
		return Claim{}, audit.Change{}, ErrInvalid
	}
	var subjectTypeID []byte
	err := tx.QueryRow(sqlSubjectType, in.SubjectID).Scan(&subjectTypeID)
	if errors.Is(err, sql.ErrNoRows) {
		return Claim{}, audit.Change{}, ErrInvalid
	}
	if err != nil {
		return Claim{}, audit.Change{}, err
	}
	if in.Status == StatusAccepted {
		if _, err := acceptedForSubject(tx, in.SubjectID); err == nil {
			return Claim{}, audit.Change{}, ErrAlreadyMember
		} else if !errors.Is(err, sql.ErrNoRows) {
			return Claim{}, audit.Change{}, err
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Claim{}, audit.Change{}, err
	}
	idBytes := id[:]
	if _, err := tx.Exec(sqlInsert, idBytes, in.SubjectID, in.EntityID, subjectTypeID,
		in.Status, nullBytes(in.ConfidenceGradeID), nullStr(in.Argument)); err != nil {
		return Claim{}, audit.Change{}, mapInsert(tx, err, in.EntityID, subjectTypeID)
	}

	change := audit.Change{
		EntityType: "identity_claim",
		EntityID:   idBytes,
		Action:     audit.ActionCreate,
		Fields: audit.FullRow(map[string]any{
			"id":                  id.String(),
			"subject_id":          uuidJSON(in.SubjectID),
			"entity_id":           uuidJSON(in.EntityID),
			"subject_type_id":     uuidJSON(subjectTypeID),
			"status":              in.Status,
			"confidence_grade_id": uuidJSON(in.ConfidenceGradeID),
			"argument":            nullStr(in.Argument),
		}),
	}
	return Claim{
		ID:                append([]byte(nil), idBytes...),
		SubjectID:         append([]byte(nil), in.SubjectID...),
		EntityID:          append([]byte(nil), in.EntityID...),
		SubjectTypeID:     subjectTypeID,
		Status:            in.Status,
		ConfidenceGradeID: append([]byte(nil), in.ConfidenceGradeID...),
		Argument:          in.Argument,
	}, change, nil
}

// Get returns a claim by id, or sql.ErrNoRows.
func Get(c *database.Catalog, id []byte) (Claim, error) {
	db, err := c.DB()
	if err != nil {
		return Claim{}, err
	}
	if len(id) != 16 {
		return Claim{}, ErrInvalid
	}
	return scanClaim(db.QueryRow(sqlGet, id))
}

// AcceptedEntityForSubject returns the subject's accepted claim (its handle
// membership), or sql.ErrNoRows when the subject is unpromoted.
func AcceptedEntityForSubject(c *database.Catalog, subjectID []byte) (Claim, error) {
	db, err := c.DB()
	if err != nil {
		return Claim{}, err
	}
	if len(subjectID) != 16 {
		return Claim{}, ErrInvalid
	}
	return scanClaim(db.QueryRow(sqlAcceptedForSubject, subjectID))
}

// AcceptedEntityForSubjectTx is AcceptedEntityForSubject on an open transaction.
func AcceptedEntityForSubjectTx(tx *sql.Tx, subjectID []byte) (Claim, error) {
	if len(subjectID) != 16 {
		return Claim{}, ErrInvalid
	}
	return acceptedForSubject(tx, subjectID)
}

// AcceptedMembers returns the handle's accepted claims (its members).
func AcceptedMembers(c *database.Catalog, entityID []byte) ([]Claim, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(entityID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlAcceptedMembers, entityID)
	if err != nil {
		return nil, err
	}
	return scanClaims(rows)
}

// AcceptedMembersTx is AcceptedMembers on an open transaction.
func AcceptedMembersTx(tx *sql.Tx, entityID []byte) ([]Claim, error) {
	if len(entityID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := tx.Query(sqlAcceptedMembers, entityID)
	if err != nil {
		return nil, err
	}
	return scanClaims(rows)
}

func acceptedForSubject(tx *sql.Tx, subjectID []byte) (Claim, error) {
	return scanClaim(tx.QueryRow(sqlAcceptedForSubject, subjectID))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanClaim(row rowScanner) (Claim, error) {
	var cl Claim
	if err := row.Scan(&cl.ID, &cl.SubjectID, &cl.EntityID, &cl.SubjectTypeID,
		&cl.Status, &cl.ConfidenceGradeID, &cl.Argument); err != nil {
		return Claim{}, err
	}
	return cl, nil
}

func scanClaims(rows *sql.Rows) ([]Claim, error) {
	defer rows.Close()
	var out []Claim
	for rows.Next() {
		cl, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cl)
	}
	return out, rows.Err()
}

// mapInsert names a type mismatch the composite FK caught; other constraint
// failures (missing handle or grade, duplicate pair) are ErrInvalid.
func mapInsert(tx *sql.Tx, err error, entityID, subjectTypeID []byte) error {
	if !database.IsConstraintViolation(err) {
		return err
	}
	var entityTypeID []byte
	if qerr := tx.QueryRow(sqlEntityType, entityID).Scan(&entityTypeID); qerr == nil &&
		string(entityTypeID) != string(subjectTypeID) {
		return ErrTypeMismatch
	}
	return ErrInvalid
}

func statusOK(s string) bool {
	switch s {
	case StatusProvisional, StatusAccepted, StatusRejected:
		return true
	}
	return false
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

func uuidJSON(id []byte) any {
	if len(id) != 16 {
		return nil
	}
	u, err := uuid.FromBytes(id)
	if err != nil {
		return nil
	}
	return u.String()
}
