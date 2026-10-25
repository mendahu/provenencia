// Package artifacts accesses the artifacts catalog table.
// Create, Update, and Delete run inside writes.Run and return the row changes
// that Run records.
package artifacts

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = apperr.New(apperr.CodeArtifactsInvalid, apperr.KindUser)

// ErrInUse is returned by Delete when Citations still point at the Artifact.
var ErrInUse = apperr.New(apperr.CodeArtifactsInUse, apperr.KindConflict)

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
	sqlGetMany = `SELECT id, ref, source_id, file_id, label, COALESCE(description, '')
		FROM artifacts WHERE id IN (`
	sqlGetByRef = `SELECT id, ref, source_id, file_id, label, COALESCE(description, '')
		FROM artifacts WHERE ref = ?`
	sqlListBySource = `SELECT id, ref, source_id, file_id, label, COALESCE(description, '')
		FROM artifacts WHERE source_id = ?
		ORDER BY ref COLLATE NOCASE`
	sqlDelete         = `DELETE FROM artifacts WHERE id = ?`
	sqlExistsBySource = `SELECT 1 FROM artifacts WHERE source_id = ? LIMIT 1`
	sqlExistsByFile   = `SELECT 1 FROM artifacts WHERE file_id = ? LIMIT 1`
	sqlSourceExists   = `SELECT 1 FROM sources WHERE id = ?`
	sqlFileExists     = `SELECT 1 FROM files WHERE id = ?`
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
	// GraphIDs are the graphs of the Source the Artifact is linked to on
	// create. Empty = every graph of the Source.
	GraphIDs [][]byte
}

// Create inserts an Artifact and mints ART-…. The caller records the returned
// changes. FileID empty is fileless.
func Create(tx *database.Tx, userID []byte, in CreateInput) (Artifact, []rowchange.Change, error) {
	if tx == nil {
		return Artifact{}, nil, ErrInvalid
	}
	in.Label = strings.TrimSpace(in.Label)
	in.Description = strings.TrimSpace(in.Description)
	if len(in.SourceID) != 16 || in.Label == "" {
		return Artifact{}, nil, ErrInvalid
	}
	if len(in.FileID) != 0 && len(in.FileID) != 16 {
		return Artifact{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Artifact{}, nil, err
	}

	if err := requireSource(tx.Tx, in.SourceID); err != nil {
		return Artifact{}, nil, err
	}
	if len(in.FileID) == 16 {
		if err := requireFile(tx.Tx, in.FileID); err != nil {
			return Artifact{}, nil, err
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Artifact{}, nil, err
	}
	idBytes := id[:]

	artRef, err := database.InsertWithRef(database.Minter(ref.PrefixArtifact), func(storedRef string) error {
		_, err := tx.Exec(sqlInsert, idBytes, storedRef, in.SourceID, database.NullBlob(in.FileID), in.Label, database.NullString(in.Description))
		return err
	})
	if errors.Is(err, database.ErrRefsExhausted) {
		return Artifact{}, nil, ErrInvalid
	}
	if err != nil {
		return Artifact{}, nil, mapConstraint(err)
	}

	fields := map[string]rowchange.FieldDiff{
		"id":        {Old: nil, New: id.String()},
		"ref":       {Old: nil, New: artRef},
		"source_id": {Old: nil, New: database.UUIDString(in.SourceID)},
		"label":     {Old: nil, New: in.Label},
	}
	if len(in.FileID) == 16 {
		fields["file_id"] = rowchange.FieldDiff{Old: nil, New: database.UUIDString(in.FileID)}
	}
	if in.Description != "" {
		fields["description"] = rowchange.FieldDiff{Old: nil, New: in.Description}
	}
	linked, err := evidencegraphs.LinkArtifact(tx, userID, in.SourceID, idBytes, in.GraphIDs)
	if err != nil {
		return Artifact{}, nil, err
	}
	changes := append([]rowchange.Change{{
		EntityType: "artifact",
		EntityID:   append([]byte(nil), idBytes...),
		Action:     rowchange.ActionCreate,
		Fields:     fields,
	}}, linked...)
	return Artifact{
		ID:          append([]byte(nil), idBytes...),
		Ref:         artRef,
		SourceID:    append([]byte(nil), in.SourceID...),
		FileID:      copyBlob(in.FileID),
		Label:       in.Label,
		Description: in.Description,
	}, changes, nil
}

// Update patches label, description, and/or first-attach file_id. The caller
// records the returned changes. An unchanged row returns no changes. Clearing
// a set file_id back to null is rejected. Replacing a set file_id with a
// different File is rejected (first-attach only).
func Update(tx *database.Tx, userID []byte, a Artifact) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	a.Label = strings.TrimSpace(a.Label)
	a.Description = strings.TrimSpace(a.Description)
	if len(a.ID) != 16 || a.Label == "" {
		return nil, ErrInvalid
	}
	if len(a.FileID) != 0 && len(a.FileID) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}

	prev, err := getTx(tx.Tx, a.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if len(prev.FileID) == 16 && len(a.FileID) == 0 {
		return nil, ErrInvalid
	}
	if len(prev.FileID) == 16 && len(a.FileID) == 16 && !bytes.Equal(prev.FileID, a.FileID) {
		return nil, ErrFileAlreadyAttached
	}
	if len(a.FileID) == 16 {
		if err := requireFile(tx.Tx, a.FileID); err != nil {
			return nil, err
		}
	}

	fields := map[string]rowchange.FieldDiff{}
	if !bytes.Equal(prev.FileID, a.FileID) {
		fields["file_id"] = rowchange.FieldDiff{
			Old: database.UUIDJSON(prev.FileID),
			New: database.UUIDJSON(a.FileID),
		}
	}
	if prev.Label != a.Label {
		fields["label"] = rowchange.FieldDiff{Old: prev.Label, New: a.Label}
	}
	if prev.Description != a.Description {
		fields["description"] = rowchange.FieldDiff{Old: database.NullString(prev.Description), New: database.NullString(a.Description)}
	}
	if len(fields) == 0 {
		return nil, nil
	}

	if _, err := tx.Exec(sqlUpdate, database.NullBlob(a.FileID), a.Label, database.NullString(a.Description), a.ID); err != nil {
		return nil, mapConstraint(err)
	}
	return []rowchange.Change{{
		EntityType: "artifact",
		EntityID:   append([]byte(nil), a.ID...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}}, nil
}

// Delete erases an Artifact when no Citation points at it. File releases
// ifUnused (snapshot file_id before DELETE). Orphan object bytes are unlinked
// from AfterCommit, after Run commits, so a rollback leaves the files in place.
// Cover SET NULL is SQLite's job. The caller records the returned changes.
func Delete(tx *database.Tx, userID, id []byte) ([]rowchange.Change, error) {
	if tx == nil || len(id) != 16 {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	existing, err := getTx(tx.Tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}

	report, err := deleteimpact.Impact(tx.Tx, catalogmodel.KindArtifact, id)
	if err != nil {
		return nil, err
	}
	if err := deleteimpact.Refuse(report, deleteimpact.Codes{
		InUse: ErrInUse, NotFound: ErrInvalid,
	}); err != nil {
		return nil, err
	}

	released, err := deleteimpact.ReleaseFacets(tx, catalogmodel.KindArtifact, id)
	if err != nil {
		return nil, err
	}
	snap, err := deleteimpact.SnapshotOwned(tx.Tx, catalogmodel.KindArtifact, id)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDelete, id); err != nil {
		return nil, err
	}
	err = deleteimpact.ReleaseSnapshot(tx, snap)
	if err != nil {
		return nil, err
	}

	fields := map[string]rowchange.FieldDiff{
		"id":        {Old: database.UUIDString(id), New: nil},
		"ref":       {Old: existing.Ref, New: nil},
		"source_id": {Old: database.UUIDString(existing.SourceID), New: nil},
		"label":     {Old: existing.Label, New: nil},
	}
	if len(existing.FileID) == 16 {
		fields["file_id"] = rowchange.FieldDiff{Old: database.UUIDString(existing.FileID), New: nil}
	}
	if existing.Description != "" {
		fields["description"] = rowchange.FieldDiff{Old: existing.Description, New: nil}
	}
	return append(released, rowchange.Change{
		EntityType: "artifact",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionDelete,
		Fields:     fields,
	}), nil
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

// GetMany returns Artifacts keyed by id string. Missing ids are omitted.
func GetMany(c *database.Catalog, ids [][]byte) (map[string]Artifact, error) {
	ids = database.UniqueBlobIDs(ids)
	out := make(map[string]Artifact, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	err = sqlbatch.Query(db, sqlGetMany, `)`, nil, ids, func(rows *sql.Rows) error {
		a, err := scanArtifact(rows)
		if err != nil {
			return err
		}
		out[string(a.ID)] = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// getByRef returns an Artifact by ART-… ref, or sql.ErrNoRows.
func getByRef(c *database.Catalog, artRef string) (Artifact, error) {
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
	return listBySource(db, sourceID)
}

// ListBySourceTx is ListBySource on an open transaction.
func ListBySourceTx(tx *sql.Tx, sourceID []byte) ([]Artifact, error) {
	if len(sourceID) != 16 {
		return nil, ErrInvalid
	}
	return listBySource(tx, sourceID)
}

func listBySource(q database.Querier, sourceID []byte) ([]Artifact, error) {
	rows, err := q.Query(sqlListBySource, sourceID)
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

func copyBlob(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	return append([]byte(nil), b...)
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
