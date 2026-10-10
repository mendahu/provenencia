// Package subjects accesses the subjects catalog table. Mutations return
// changes for the caller to record.
package subjects

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
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

// Placement is an optional grid cell written with the Subject (unaudited layout).
type Placement struct {
	GridX int64
	GridY int64
}

// Create inserts a Subject and mints a ref from the type's candidate_ref_prefix.
// A placement is written on the same transaction and returned with the subject
// change. The caller records the returned changes. Run stores the position and
// does not record it.
func Create(tx *database.Tx, userID []byte, in CreateInput, placement *Placement) (Subject, []rowchange.Change, error) {
	if tx == nil {
		return Subject{}, nil, ErrInvalid
	}
	in.Label = strings.TrimSpace(in.Label)
	in.Description = strings.TrimSpace(in.Description)
	if len(in.SourceID) != 16 || len(in.SubjectTypeID) != 16 {
		return Subject{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Subject{}, nil, err
	}

	s, change, err := InsertTx(tx.Tx, in)
	if err != nil {
		return Subject{}, nil, err
	}
	changes := []rowchange.Change{change}
	if placement != nil {
		_, pos, err := subjectpositions.Set(tx, s.ID, placement.GridX, placement.GridY)
		if err != nil {
			return Subject{}, nil, err
		}
		changes = append(changes, pos...)
	}
	return s, changes, nil
}

// InsertTx inserts a Subject on an open transaction (no commit, no revision).
func InsertTx(tx *sql.Tx, in CreateInput) (Subject, rowchange.Change, error) {
	in.Label = strings.TrimSpace(in.Label)
	in.Description = strings.TrimSpace(in.Description)
	if len(in.SourceID) != 16 || len(in.SubjectTypeID) != 16 {
		return Subject{}, rowchange.Change{}, ErrInvalid
	}
	if err := requireSource(tx, in.SourceID); err != nil {
		return Subject{}, rowchange.Change{}, err
	}
	prefix, err := requireTypePrefix(tx, in.SubjectTypeID)
	if err != nil {
		return Subject{}, rowchange.Change{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Subject{}, rowchange.Change{}, err
	}
	idBytes := id[:]

	var subjectRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		subjectRef, err = ref.Mint(prefix)
		if err != nil {
			return Subject{}, rowchange.Change{}, err
		}
		_, err = tx.Exec(sqlInsert, idBytes, subjectRef, in.SourceID, in.SubjectTypeID, nullStr(in.Label), nullStr(in.Description))
		if err == nil {
			break
		}
		if !database.IsUniqueConflict(err) {
			return Subject{}, rowchange.Change{}, mapConstraint(err)
		}
	}
	if err != nil {
		return Subject{}, rowchange.Change{}, ErrInvalid
	}

	change := rowchange.Change{
		EntityType: "subject",
		EntityID:   idBytes,
		Action:     rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{
			"id":              id.String(),
			"ref":             subjectRef,
			"source_id":       uuidJSON(in.SourceID),
			"subject_type_id": uuidJSON(in.SubjectTypeID),
			"label":           nullJSON(in.Label),
			"description":     nullJSON(in.Description),
		}),
	}
	return Subject{
		ID:            append([]byte(nil), idBytes...),
		Ref:           subjectRef,
		SourceID:      append([]byte(nil), in.SourceID...),
		SubjectTypeID: append([]byte(nil), in.SubjectTypeID...),
		Label:         in.Label,
		Description:   in.Description,
	}, change, nil
}

// Update changes label and/or description only. subject_type_id is immutable.
// An unchanged row returns no changes. The caller records the returned changes.
func Update(tx *database.Tx, userID, id []byte, label, description string) ([]rowchange.Change, error) {
	if tx == nil || len(id) != 16 {
		return nil, ErrInvalid
	}
	label = strings.TrimSpace(label)
	description = strings.TrimSpace(description)
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	prev, err := getTx(tx.Tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}

	fields := map[string]rowchange.FieldDiff{}
	if prev.Label != label {
		fields["label"] = rowchange.FieldDiff{Old: nullJSON(prev.Label), New: nullJSON(label)}
	}
	if prev.Description != description {
		fields["description"] = rowchange.FieldDiff{Old: nullJSON(prev.Description), New: nullJSON(description)}
	}
	if len(fields) == 0 {
		return nil, nil
	}

	if _, err := tx.Exec(sqlUpdate, nullStr(label), nullStr(description), id); err != nil {
		return nil, mapConstraint(err)
	}
	return []rowchange.Change{{
		EntityType: "subject",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}}, nil
}

// Delete erases a Subject when Impact allows it. Connection facets (edge +
// disambiguation rows on a bridge) are released first. Positions CASCADE.
// Handles the released changes name, handles that observed this Subject, linked
// handles, and their header dependents are recomputed here in one call.
// Header documents are written at the end of that call, so they have to see
// the released values already gone. Run recomputes the returned changes again.
func Delete(tx *database.Tx, userID, id []byte) ([]rowchange.Change, error) {
	if tx == nil || len(id) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	prev, err := getTx(tx.Tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}

	report, err := deleteimpact.Impact(tx.Tx, catalogmodel.KindSubject, id)
	if err != nil {
		return nil, err
	}
	if err := deleteimpact.Refuse(report, deleteimpact.Codes{
		InUse: ErrInUse, NotFound: ErrInvalid,
	}); err != nil {
		return nil, err
	}
	// Header dependents are snapshotted while the link still exists. The
	// recompute itself runs after the rows are gone, and it includes the
	// handles effects.Handles reads off the released changes: RecomputeTx
	// writes header search documents after it clears values, so those
	// documents have to be in the same call.
	inbound, err := autoreconciler.HandlesObservingSubject(tx.Tx, id)
	if err != nil {
		return nil, err
	}
	ends, err := linkedHandles(tx.Tx, id)
	if err != nil {
		return nil, err
	}
	deps, err := conclusionheaders.HeaderDependents(tx.Tx, ends)
	if err != nil {
		return nil, err
	}
	released, err := deleteimpact.ReleaseFacets(tx.Tx, catalogmodel.KindSubject, id)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDelete, id); err != nil {
		return nil, err
	}
	releasedHandles, err := effects.Handles(tx.Tx, released.Changes)
	if err != nil {
		return nil, err
	}
	touched := append(append(releasedHandles, inbound...), ends...)
	if err := autoreconciler.RecomputeTx(tx.Tx, append(touched, deps...)); err != nil {
		return nil, err
	}
	fields := map[string]rowchange.FieldDiff{
		"id":              {Old: uuidString(id), New: nil},
		"ref":             {Old: prev.Ref, New: nil},
		"source_id":       {Old: uuidString(prev.SourceID), New: nil},
		"subject_type_id": {Old: uuidString(prev.SubjectTypeID), New: nil},
	}
	if prev.Label != "" {
		fields["label"] = rowchange.FieldDiff{Old: prev.Label, New: nil}
	}
	if prev.Description != "" {
		fields["description"] = rowchange.FieldDiff{Old: prev.Description, New: nil}
	}
	return append(released.Changes, rowchange.Change{
		EntityType: "subject",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionDelete,
		Fields:     fields,
	}), nil
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

// GetTx returns a Subject by id on an open transaction, or sql.ErrNoRows.
func GetTx(tx *sql.Tx, id []byte) (Subject, error) {
	if len(id) != 16 {
		return Subject{}, ErrInvalid
	}
	return getTx(tx, id)
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

// linkedHandles are accepted handles of Subjects this one points at.
func linkedHandles(tx *sql.Tx, subjectID []byte) ([][]byte, error) {
	rows, err := tx.Query(`
		SELECT DISTINCT ic.entity_id
		FROM observations o
		JOIN identity_claims ic ON ic.subject_id = o.value_subject_id AND ic.status = 'accepted'
		WHERE o.subject_id = ?`, subjectID)
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

func uuidString(id []byte) string {
	u, err := uuid.FromBytes(id)
	if err != nil {
		return ""
	}
	return u.String()
}

func uuidJSON(id []byte) any {
	if len(id) != 16 {
		return nil
	}
	s := uuidString(id)
	if s == "" {
		return nil
	}
	return s
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
