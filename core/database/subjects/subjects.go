// Package subjects accesses the subjects catalog table with audited mutations.
package subjects

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/ref"
)

var (
	ErrInvalid = apperr.New(apperr.CodeSubjectsInvalid, apperr.KindUser)
	ErrInUse   = apperr.New(apperr.CodeSubjectsInUse, apperr.KindConflict)
)

const (
	sqlInsert = `INSERT INTO subjects (id, ref, source_id, subject_type_id, label, description)
		VALUES (?, ?, ?, ?, ?, ?)`
	sqlUpdate = `UPDATE subjects SET label = ?, description = ? WHERE id = ?`
	sqlDelete = `DELETE FROM subjects WHERE id = ?`
	sqlGet    = `SELECT id, ref, source_id, subject_type_id, COALESCE(label, ''), COALESCE(description, '')
		FROM subjects WHERE id = ?`
	sqlGetByRef = `SELECT id, ref, source_id, subject_type_id, COALESCE(label, ''), COALESCE(description, '')
		FROM subjects WHERE ref = ?`
	sqlListBySource = `SELECT id, ref, source_id, subject_type_id, COALESCE(label, ''), COALESCE(description, '')
		FROM subjects WHERE source_id = ?
		ORDER BY label COLLATE NOCASE, ref COLLATE NOCASE`
	sqlSourceExists = `SELECT 1 FROM sources WHERE id = ?`
	sqlTypePrefix   = `SELECT candidate_ref_prefix FROM subject_types WHERE id = ?`
	sqlObsRefs      = `SELECT COUNT(*) FROM observations WHERE subject_id = ? OR value_subject_id = ?`
	maxRefRetries   = 8
)

// Subject is one subjects row.
type Subject struct {
	ID            []byte
	Ref           string
	SourceID      []byte
	SubjectTypeID []byte
	Label         string
	Description   string
}

// CreateInput is the mutable fields for a new Subject.
type CreateInput struct {
	SourceID      []byte
	SubjectTypeID []byte
	Label         string
	Description   string
}

// Create inserts a Subject, mints a ref from the type's candidate_ref_prefix, and records create_subject.
func Create(c *database.Catalog, userID []byte, in CreateInput) (Subject, error) {
	db, err := c.DB()
	if err != nil {
		return Subject{}, err
	}
	in.Label = strings.TrimSpace(in.Label)
	in.Description = strings.TrimSpace(in.Description)
	if len(in.SourceID) != 16 || len(in.SubjectTypeID) != 16 {
		return Subject{}, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Subject{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Subject{}, err
	}
	defer func() { _ = tx.Rollback() }()

	s, err := InsertTx(tx, userID, in)
	if err != nil {
		return Subject{}, err
	}
	if err := tx.Commit(); err != nil {
		return Subject{}, err
	}
	return s, nil
}

// InsertTx inserts a Subject on an open transaction (no commit).
func InsertTx(tx *sql.Tx, userID []byte, in CreateInput) (Subject, error) {
	in.Label = strings.TrimSpace(in.Label)
	in.Description = strings.TrimSpace(in.Description)
	if len(in.SourceID) != 16 || len(in.SubjectTypeID) != 16 {
		return Subject{}, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Subject{}, err
	}
	if err := requireSource(tx, in.SourceID); err != nil {
		return Subject{}, err
	}
	prefix, err := requireTypePrefix(tx, in.SubjectTypeID)
	if err != nil {
		return Subject{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Subject{}, err
	}
	idBytes := id[:]

	var subjectRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		subjectRef, err = ref.Mint(prefix)
		if err != nil {
			return Subject{}, err
		}
		_, err = tx.Exec(sqlInsert, idBytes, subjectRef, in.SourceID, in.SubjectTypeID, nullStr(in.Label), nullStr(in.Description))
		if err == nil {
			break
		}
		if !database.IsUniqueConflict(err) {
			return Subject{}, mapConstraint(err)
		}
	}
	if err != nil {
		return Subject{}, ErrInvalid
	}

	fields := map[string]audit.FieldDiff{
		"id":              {Old: nil, New: id.String()},
		"ref":             {Old: nil, New: subjectRef},
		"source_id":       {Old: nil, New: uuidString(in.SourceID)},
		"subject_type_id": {Old: nil, New: uuidString(in.SubjectTypeID)},
	}
	if in.Label != "" {
		fields["label"] = audit.FieldDiff{Old: nil, New: in.Label}
	}
	if in.Description != "" {
		fields["description"] = audit.FieldDiff{Old: nil, New: in.Description}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_subject",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "subject",
			EntityID:   idBytes,
			Action:     audit.ActionCreate,
			Fields:     fields,
		}},
	}); err != nil {
		return Subject{}, err
	}
	return Subject{
		ID:            append([]byte(nil), idBytes...),
		Ref:           subjectRef,
		SourceID:      append([]byte(nil), in.SourceID...),
		SubjectTypeID: append([]byte(nil), in.SubjectTypeID...),
		Label:         in.Label,
		Description:   in.Description,
	}, nil
}

// Update changes label and/or description only. subject_type_id is immutable.
// No-op when nothing changed (commits without a revision).
func Update(c *database.Catalog, userID, id []byte, label, description string) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if len(id) != 16 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := getTx(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}

	fields := map[string]audit.FieldDiff{}
	if prev.Label != label {
		fields["label"] = audit.FieldDiff{Old: nullJSON(prev.Label), New: nullJSON(label)}
	}
	if prev.Description != description {
		fields["description"] = audit.FieldDiff{Old: nullJSON(prev.Description), New: nullJSON(description)}
	}
	if len(fields) == 0 {
		return tx.Commit()
	}

	if _, err := tx.Exec(sqlUpdate, nullStr(label), nullStr(description), id); err != nil {
		return mapConstraint(err)
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_subject",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "subject",
			EntityID:   id,
			Action:     audit.ActionUpdate,
			Fields:     fields,
		}},
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// Delete removes a Subject and records delete_subject. Positions CASCADE.
func Delete(c *database.Catalog, userID, id []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(id) != 16 {
		return ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := getTx(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	var obsCount int
	if err := tx.QueryRow(sqlObsRefs, id, id).Scan(&obsCount); err != nil {
		return err
	}
	if obsCount > 0 {
		return ErrInUse
	}
	if _, err := tx.Exec(sqlDelete, id); err != nil {
		return err
	}
	fields := map[string]audit.FieldDiff{
		"id":              {Old: uuidString(id), New: nil},
		"ref":             {Old: prev.Ref, New: nil},
		"source_id":       {Old: uuidString(prev.SourceID), New: nil},
		"subject_type_id": {Old: uuidString(prev.SubjectTypeID), New: nil},
	}
	if prev.Label != "" {
		fields["label"] = audit.FieldDiff{Old: prev.Label, New: nil}
	}
	if prev.Description != "" {
		fields["description"] = audit.FieldDiff{Old: prev.Description, New: nil}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "delete_subject",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "subject",
			EntityID:   id,
			Action:     audit.ActionDelete,
			Fields:     fields,
		}},
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// Get returns a Subject by id, or sql.ErrNoRows.
func Get(c *database.Catalog, id []byte) (Subject, error) {
	db, err := c.DB()
	if err != nil {
		return Subject{}, err
	}
	if len(id) != 16 {
		return Subject{}, ErrInvalid
	}
	return scanSubject(db.QueryRow(sqlGet, id))
}

// getByRef returns a Subject by ref, or sql.ErrNoRows.
func getByRef(c *database.Catalog, subjectRef string) (Subject, error) {
	db, err := c.DB()
	if err != nil {
		return Subject{}, err
	}
	subjectRef = strings.TrimSpace(subjectRef)
	if ref.Validate(subjectRef) != nil {
		return Subject{}, ErrInvalid
	}
	return scanSubject(db.QueryRow(sqlGetByRef, subjectRef))
}

// ListBySource returns Subjects homed to sourceID.
func ListBySource(c *database.Catalog, sourceID []byte) ([]Subject, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := db.Query(sqlListBySource, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subject
	for rows.Next() {
		s, err := scanSubject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func getTx(tx *sql.Tx, id []byte) (Subject, error) {
	return scanSubject(tx.QueryRow(sqlGet, id))
}

func scanSubject(row rowScanner) (Subject, error) {
	var s Subject
	if err := row.Scan(&s.ID, &s.Ref, &s.SourceID, &s.SubjectTypeID, &s.Label, &s.Description); err != nil {
		return Subject{}, err
	}
	return s, nil
}

func requireSource(tx *sql.Tx, sourceID []byte) error {
	var one int
	err := tx.QueryRow(sqlSourceExists, sourceID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	return err
}

func requireTypePrefix(tx *sql.Tx, typeID []byte) (string, error) {
	var prefix string
	err := tx.QueryRow(sqlTypePrefix, typeID).Scan(&prefix)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrInvalid
	}
	if err != nil {
		return "", err
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "", ErrInvalid
	}
	return prefix, nil
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
