// Package artifacts accesses the artifacts catalog table with audited mutations.
package artifacts

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/audit"
	"github.com/mendahu/provenencia/core/database/project"
	"github.com/mendahu/provenencia/core/database/searchindex"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = apperr.New(apperr.CodeArtifactsInvalid, apperr.KindUser)

// ErrFileAlreadyAttached is returned when attaching a File to an Artifact that
// already has a primary File (first-attach only; better scan = new Artifact).
var ErrFileAlreadyAttached = apperr.New(apperr.CodeArtifactsFileAlreadyAttached, apperr.KindConflict)

const (
	sqlInsert = `INSERT INTO artifacts (id, ref, source_id, file_id, label, description)
		VALUES (?, ?, ?, ?, ?, ?)`
	sqlUpdate = `UPDATE artifacts SET file_id = ?, label = ?, description = ?
		WHERE id = ?`
	sqlGet = `SELECT id, ref, source_id, file_id, label, COALESCE(description, '')
		FROM artifacts WHERE id = ?`
	sqlGetByRef = `SELECT id, ref, source_id, file_id, label, COALESCE(description, '')
		FROM artifacts WHERE ref = ?`
	sqlListBySource = `SELECT id, ref, source_id, file_id, label, COALESCE(description, '')
		FROM artifacts WHERE source_id = ?
		ORDER BY ref COLLATE NOCASE`
	sqlExistsBySource = `SELECT 1 FROM artifacts WHERE source_id = ? LIMIT 1`
	sqlExistsByFile   = `SELECT 1 FROM artifacts WHERE file_id = ? LIMIT 1`
	sqlSourceExists = `SELECT 1 FROM sources WHERE id = ?`
	sqlFileExists   = `SELECT 1 FROM files WHERE id = ?`
	maxRefRetries   = 8
)

// Artifact is one artifacts row. FileID is nil when fileless.
type Artifact struct {
	ID          []byte
	Ref         string
	SourceID    []byte
	FileID      []byte
	Label       string
	Description string
}

// CreateInput is the mutable fields for a new Artifact.
type CreateInput struct {
	SourceID    []byte
	FileID      []byte // nil/empty = fileless
	Label       string
	Description string
}

// Create inserts an Artifact, mints ART-…, and records create_artifact.
func Create(c *database.Catalog, userID []byte, in CreateInput) (Artifact, error) {
	db, err := c.DB()
	if err != nil {
		return Artifact{}, err
	}
	in.Label = strings.TrimSpace(in.Label)
	in.Description = strings.TrimSpace(in.Description)
	if len(in.SourceID) != 16 || in.Label == "" {
		return Artifact{}, ErrInvalid
	}
	if len(in.FileID) != 0 && len(in.FileID) != 16 {
		return Artifact{}, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Artifact{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Artifact{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := requireSource(tx, in.SourceID); err != nil {
		return Artifact{}, err
	}
	if len(in.FileID) == 16 {
		if err := requireFile(tx, in.FileID); err != nil {
			return Artifact{}, err
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Artifact{}, err
	}
	idBytes := id[:]

	var artRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		artRef, err = ref.Mint(ref.PrefixArtifact)
		if err != nil {
			return Artifact{}, err
		}
		_, err = tx.Exec(sqlInsert, idBytes, artRef, in.SourceID, nullBlob(in.FileID), in.Label, nullStr(in.Description))
		if err == nil {
			break
		}
		if !database.IsUniqueConflict(err) {
			return Artifact{}, mapConstraint(err)
		}
	}
	if err != nil {
		return Artifact{}, ErrInvalid
	}

	fields := map[string]audit.FieldDiff{
		"id":        {Old: nil, New: id.String()},
		"ref":       {Old: nil, New: artRef},
		"source_id": {Old: nil, New: uuidString(in.SourceID)},
		"label":     {Old: nil, New: in.Label},
	}
	if len(in.FileID) == 16 {
		fields["file_id"] = audit.FieldDiff{Old: nil, New: uuidString(in.FileID)}
	}
	if in.Description != "" {
		fields["description"] = audit.FieldDiff{Old: nil, New: in.Description}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_artifact",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "artifact",
			EntityID:   idBytes,
			Action:     audit.ActionCreate,
			Fields:     fields,
		}},
	}); err != nil {
		return Artifact{}, err
	}
	if err := searchindex.ReprojectSource(tx, in.SourceID); err != nil {
		return Artifact{}, err
	}
	if err := tx.Commit(); err != nil {
		return Artifact{}, err
	}
	return Artifact{
		ID:          append([]byte(nil), idBytes...),
		Ref:         artRef,
		SourceID:    append([]byte(nil), in.SourceID...),
		FileID:      copyBlob(in.FileID),
		Label:       in.Label,
		Description: in.Description,
	}, nil
}

// Update patches label, description, and/or first-attach file_id; records
// update_artifact for changed fields. Clearing a set file_id back to null is
// rejected. Replacing a set file_id with a different File is rejected
// (first-attach only).
func Update(c *database.Catalog, userID []byte, a Artifact) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	a.Label = strings.TrimSpace(a.Label)
	a.Description = strings.TrimSpace(a.Description)
	if len(a.ID) != 16 || a.Label == "" {
		return ErrInvalid
	}
	if len(a.FileID) != 0 && len(a.FileID) != 16 {
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

	prev, err := getTx(tx, a.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	if len(prev.FileID) == 16 && len(a.FileID) == 0 {
		return ErrInvalid
	}
	if len(prev.FileID) == 16 && len(a.FileID) == 16 && !bytesEqual(prev.FileID, a.FileID) {
		return ErrFileAlreadyAttached
	}
	if len(a.FileID) == 16 {
		if err := requireFile(tx, a.FileID); err != nil {
			return err
		}
	}

	fields := map[string]audit.FieldDiff{}
	if !bytesEqual(prev.FileID, a.FileID) {
		fields["file_id"] = audit.FieldDiff{
			Old: uuidJSON(prev.FileID),
			New: uuidJSON(a.FileID),
		}
	}
	if prev.Label != a.Label {
		fields["label"] = audit.FieldDiff{Old: prev.Label, New: a.Label}
	}
	if prev.Description != a.Description {
		fields["description"] = audit.FieldDiff{Old: nullJSON(prev.Description), New: nullJSON(a.Description)}
	}
	if len(fields) == 0 {
		return tx.Commit()
	}

	if _, err := tx.Exec(sqlUpdate, nullBlob(a.FileID), a.Label, nullStr(a.Description), a.ID); err != nil {
		return mapConstraint(err)
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_artifact",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "artifact",
			EntityID:   a.ID,
			Action:     audit.ActionUpdate,
			Fields:     fields,
		}},
	}); err != nil {
		return err
	}
	if err := searchindex.ReprojectSource(tx, prev.SourceID); err != nil {
		return err
	}
	return tx.Commit()
}

// Get returns an Artifact by id, or sql.ErrNoRows.
func Get(c *database.Catalog, id []byte) (Artifact, error) {
	db, err := c.DB()
	if err != nil {
		return Artifact{}, err
	}
	if len(id) != 16 {
		return Artifact{}, ErrInvalid
	}
	return scanArtifact(db.QueryRow(sqlGet, id))
}

// GetByRef returns an Artifact by ART-… ref, or sql.ErrNoRows.
func GetByRef(c *database.Catalog, artRef string) (Artifact, error) {
	db, err := c.DB()
	if err != nil {
		return Artifact{}, err
	}
	artRef = strings.TrimSpace(artRef)
	if ref.Validate(artRef) != nil {
		return Artifact{}, ErrInvalid
	}
	return scanArtifact(db.QueryRow(sqlGetByRef, artRef))
}

// ListBySource returns Artifacts for a Source, ordered by ref.
func ListBySource(c *database.Catalog, sourceID []byte) ([]Artifact, error) {
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
	var out []Artifact
	for rows.Next() {
		a, err := scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// HasAnyForSource reports whether the Source has at least one Artifact (fileless counts).
func HasAnyForSource(c *database.Catalog, sourceID []byte) (bool, error) {
	db, err := c.DB()
	if err != nil {
		return false, err
	}
	if len(sourceID) != 16 {
		return false, ErrInvalid
	}
	var one int
	err = db.QueryRow(sqlExistsBySource, sourceID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// HasPrimaryFile reports whether any Artifact uses fileID as its primary File.
func HasPrimaryFile(c *database.Catalog, fileID []byte) (bool, error) {
	db, err := c.DB()
	if err != nil {
		return false, err
	}
	if len(fileID) != 16 {
		return false, ErrInvalid
	}
	var one int
	err = db.QueryRow(sqlExistsByFile, fileID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanArtifact(row rowScanner) (Artifact, error) {
	var a Artifact
	var fileID []byte
	if err := row.Scan(&a.ID, &a.Ref, &a.SourceID, &fileID, &a.Label, &a.Description); err != nil {
		return Artifact{}, err
	}
	a.FileID = fileID
	return a, nil
}

func getTx(tx *sql.Tx, id []byte) (Artifact, error) {
	return scanArtifact(tx.QueryRow(sqlGet, id))
}

func requireSource(tx *sql.Tx, sourceID []byte) error {
	var one int
	err := tx.QueryRow(sqlSourceExists, sourceID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	return err
}

func requireFile(tx *sql.Tx, fileID []byte) error {
	var one int
	err := tx.QueryRow(sqlFileExists, fileID).Scan(&one)
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

func nullBlob(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
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

func uuidJSON(id []byte) any {
	if len(id) == 0 {
		return nil
	}
	return uuidString(id)
}

func copyBlob(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	return append([]byte(nil), b...)
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
