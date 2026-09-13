// Package sources accesses the sources catalog table with audited mutations.
package sources

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

var ErrInvalid = apperr.New(apperr.CodeSourcesInvalid, apperr.KindUser)

// Cover modes persisted on sources.cover_mode.
const (
	CoverModeTypeIcon = "type_icon"
	CoverModeArtifact = "artifact"
)

const (
	sqlInsert = `INSERT INTO sources (id, ref, source_type_id, title, description)
		VALUES (?, ?, ?, ?, ?)`
	sqlUpdate = `UPDATE sources SET source_type_id = ?, title = ?, description = ?
		WHERE id = ?`
	sqlSetCover = `UPDATE sources SET cover_mode = ?, primary_artifact_id = ?
		WHERE id = ?`
	sqlGet = `SELECT id, ref, source_type_id, title, COALESCE(description, ''),
		cover_mode, primary_artifact_id
		FROM sources WHERE id = ?`
	sqlGetByRef = `SELECT id, ref, source_type_id, title, COALESCE(description, ''),
		cover_mode, primary_artifact_id
		FROM sources WHERE ref = ?`
	sqlList = `SELECT id, ref, source_type_id, title, COALESCE(description, ''),
		cover_mode, primary_artifact_id
		FROM sources
		ORDER BY title COLLATE NOCASE, ref COLLATE NOCASE`
	sqlCount      = `SELECT COUNT(*) FROM sources`
	sqlTypeExists = `SELECT 1 FROM source_types WHERE id = ?`
	sqlArtifactForCover = `SELECT id, source_id, file_id FROM artifacts WHERE id = ?`
	maxRefRetries       = 8
)

// Source is one sources row.
type Source struct {
	ID                []byte
	Ref               string
	SourceTypeID      []byte
	Title             string
	Description       string
	CoverMode         string
	PrimaryArtifactID []byte // nil when CoverModeTypeIcon
}

// CreateInput is the mutable fields for a new Source.
type CreateInput struct {
	SourceTypeID []byte
	Title        string
	Description  string
}

// Create inserts a Source, mints SRC-…, and records create_source.
func Create(c *database.Catalog, userID []byte, in CreateInput) (Source, error) {
	db, err := c.DB()
	if err != nil {
		return Source{}, err
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if len(in.SourceTypeID) != 16 || in.Title == "" {
		return Source{}, ErrInvalid
	}
	if err := requireUserID(userID); err != nil {
		return Source{}, err
	}

	tx, err := db.Begin()
	if err != nil {
		return Source{}, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := requireType(tx, in.SourceTypeID); err != nil {
		return Source{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Source{}, err
	}
	idBytes := id[:]

	var sourceRef string
	for attempt := 0; attempt < maxRefRetries; attempt++ {
		sourceRef, err = ref.Mint(ref.PrefixSource)
		if err != nil {
			return Source{}, err
		}
		_, err = tx.Exec(sqlInsert, idBytes, sourceRef, in.SourceTypeID, in.Title, nullStr(in.Description))
		if err == nil {
			break
		}
		if !isUniqueConflict(err) {
			return Source{}, mapConstraint(err)
		}
	}
	if err != nil {
		return Source{}, ErrInvalid
	}

	fields := map[string]audit.FieldDiff{
		"id":             {Old: nil, New: id.String()},
		"ref":            {Old: nil, New: sourceRef},
		"source_type_id": {Old: nil, New: uuidString(in.SourceTypeID)},
		"title":          {Old: nil, New: in.Title},
	}
	if in.Description != "" {
		fields["description"] = audit.FieldDiff{Old: nil, New: in.Description}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "create_source",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "source",
			EntityID:   idBytes,
			Action:     audit.ActionCreate,
			Fields:     fields,
		}},
	}); err != nil {
		return Source{}, err
	}
	if err := tx.Commit(); err != nil {
		return Source{}, err
	}
	return Source{
		ID:           append([]byte(nil), idBytes...),
		Ref:          sourceRef,
		SourceTypeID: append([]byte(nil), in.SourceTypeID...),
		Title:        in.Title,
		Description:  in.Description,
		CoverMode:    CoverModeTypeIcon,
	}, nil
}

// SetCover pins an Artifact as cover or reverts to the Source type icon.
// mode must be CoverModeArtifact (with a file-bearing Artifact under this Source)
// or CoverModeTypeIcon (clears primary_artifact_id).
func SetCover(c *database.Catalog, userID []byte, sourceID []byte, mode string, primaryArtifactID []byte) (Source, error) {
	db, err := c.DB()
	if err != nil {
		return Source{}, err
	}
	if len(sourceID) != 16 {
		return Source{}, ErrInvalid
	}
	if err := requireUserID(userID); err != nil {
		return Source{}, err
	}
	mode = strings.TrimSpace(mode)
	switch mode {
	case CoverModeTypeIcon:
		primaryArtifactID = nil
	case CoverModeArtifact:
		if len(primaryArtifactID) != 16 {
			return Source{}, ErrInvalid
		}
	default:
		return Source{}, ErrInvalid
	}

	tx, err := db.Begin()
	if err != nil {
		return Source{}, err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := getTx(tx, sourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return Source{}, ErrInvalid
	}
	if err != nil {
		return Source{}, err
	}

	if mode == CoverModeArtifact {
		if err := requireFileBearingArtifact(tx, sourceID, primaryArtifactID); err != nil {
			return Source{}, err
		}
	}

	if prev.CoverMode == mode && bytesEqual(prev.PrimaryArtifactID, primaryArtifactID) {
		if err := tx.Commit(); err != nil {
			return Source{}, err
		}
		return prev, nil
	}

	if _, err := tx.Exec(sqlSetCover, mode, nullBlob(primaryArtifactID), sourceID); err != nil {
		return Source{}, mapConstraint(err)
	}

	fields := map[string]audit.FieldDiff{
		"cover_mode": {Old: prev.CoverMode, New: mode},
	}
	if !bytesEqual(prev.PrimaryArtifactID, primaryArtifactID) {
		fields["primary_artifact_id"] = audit.FieldDiff{
			Old: uuidJSON(prev.PrimaryArtifactID),
			New: uuidJSON(primaryArtifactID),
		}
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "set_source_cover",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "source",
			EntityID:   sourceID,
			Action:     audit.ActionUpdate,
			Fields:     fields,
		}},
	}); err != nil {
		return Source{}, err
	}
	if err := tx.Commit(); err != nil {
		return Source{}, err
	}
	return Get(c, sourceID)
}

// MaybePinFirstFileCover pins artifactID as the Source cover when the Source
// still has no primary (type_icon). Same end state as SetCover(artifact, id).
// No-op when a cover Artifact is already set.
func MaybePinFirstFileCover(c *database.Catalog, userID, sourceID, artifactID []byte) (Source, bool, error) {
	s, err := Get(c, sourceID)
	if err != nil {
		return Source{}, false, err
	}
	if s.CoverMode == CoverModeArtifact && len(s.PrimaryArtifactID) == 16 {
		return s, false, nil
	}
	got, err := SetCover(c, userID, sourceID, CoverModeArtifact, artifactID)
	if err != nil {
		return Source{}, false, err
	}
	return got, true, nil
}

// Update patches title, description, and source_type_id; records update_source for changed fields.
func Update(c *database.Catalog, userID []byte, s Source) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	s.Title = strings.TrimSpace(s.Title)
	s.Description = strings.TrimSpace(s.Description)
	if len(s.ID) != 16 || len(s.SourceTypeID) != 16 || s.Title == "" {
		return ErrInvalid
	}
	if err := requireUserID(userID); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	prev, err := getTx(tx, s.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	if err := requireType(tx, s.SourceTypeID); err != nil {
		return err
	}

	fields := map[string]audit.FieldDiff{}
	if !bytesEqual(prev.SourceTypeID, s.SourceTypeID) {
		fields["source_type_id"] = audit.FieldDiff{
			Old: uuidString(prev.SourceTypeID),
			New: uuidString(s.SourceTypeID),
		}
	}
	if prev.Title != s.Title {
		fields["title"] = audit.FieldDiff{Old: prev.Title, New: s.Title}
	}
	if prev.Description != s.Description {
		fields["description"] = audit.FieldDiff{Old: nullJSON(prev.Description), New: nullJSON(s.Description)}
	}
	if len(fields) == 0 {
		return tx.Commit()
	}

	if _, err := tx.Exec(sqlUpdate, s.SourceTypeID, s.Title, nullStr(s.Description), s.ID); err != nil {
		return mapConstraint(err)
	}
	if _, err := audit.Record(tx, audit.Revision{
		UserID:     userID,
		ActionType: "update_source",
		CreatedAt:  project.NowUTC(),
		Changes: []audit.Change{{
			EntityType: "source",
			EntityID:   s.ID,
			Action:     audit.ActionUpdate,
			Fields:     fields,
		}},
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// Get returns a Source by id, or sql.ErrNoRows.
func Get(c *database.Catalog, id []byte) (Source, error) {
	db, err := c.DB()
	if err != nil {
		return Source{}, err
	}
	if len(id) != 16 {
		return Source{}, ErrInvalid
	}
	return scanSource(db.QueryRow(sqlGet, id))
}

// GetByRef returns a Source by SRC-… ref, or sql.ErrNoRows.
func GetByRef(c *database.Catalog, sourceRef string) (Source, error) {
	db, err := c.DB()
	if err != nil {
		return Source{}, err
	}
	sourceRef = strings.TrimSpace(sourceRef)
	if ref.Validate(sourceRef) != nil {
		return Source{}, ErrInvalid
	}
	return scanSource(db.QueryRow(sqlGetByRef, sourceRef))
}

// List returns Sources ordered for a catalog list.
func List(c *database.Catalog) ([]Source, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sqlList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Source
	for rows.Next() {
		s, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Count returns how many Sources are in the catalog — used by the workspace
// sidebar badge, not the Sources list itself.
func Count(c *database.Catalog) (int, error) {
	db, err := c.DB()
	if err != nil {
		return 0, err
	}
	var n int
	if err := db.QueryRow(sqlCount).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSource(row rowScanner) (Source, error) {
	var s Source
	var primary []byte
	if err := row.Scan(&s.ID, &s.Ref, &s.SourceTypeID, &s.Title, &s.Description, &s.CoverMode, &primary); err != nil {
		return Source{}, err
	}
	s.PrimaryArtifactID = primary
	if s.CoverMode == "" {
		s.CoverMode = CoverModeTypeIcon
	}
	return s, nil
}

func getTx(tx *sql.Tx, id []byte) (Source, error) {
	return scanSource(tx.QueryRow(sqlGet, id))
}

func requireType(tx *sql.Tx, typeID []byte) error {
	var one int
	err := tx.QueryRow(sqlTypeExists, typeID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	return err
}

func requireFileBearingArtifact(tx *sql.Tx, sourceID, artifactID []byte) error {
	var id, artSource, fileID []byte
	err := tx.QueryRow(sqlArtifactForCover, artifactID).Scan(&id, &artSource, &fileID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	if !bytesEqual(artSource, sourceID) || len(fileID) != 16 {
		return ErrInvalid
	}
	return nil
}

func nullBlob(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

func uuidJSON(id []byte) any {
	if len(id) != 16 {
		return nil
	}
	return uuidString(id)
}

func requireUserID(userID []byte) error {
	if len(userID) == 0 {
		return nil
	}
	if len(userID) != 16 {
		return ErrInvalid
	}
	return nil
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

func isUniqueConflict(err error) bool {
	return database.IsUniqueConflict(err)
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
