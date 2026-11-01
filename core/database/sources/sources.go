// Package sources accesses the sources catalog table with audited mutations.
package sources

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
	"github.com/mendahu/provenencia/core/database/rowchange"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/filederivatives"
	"github.com/mendahu/provenencia/core/ref"
)

var ErrInvalid = apperr.New(apperr.CodeSourcesInvalid, apperr.KindUser)

// ErrInUse is returned by Delete when Artifacts or Subjects still belong to the Source.
var ErrInUse = apperr.New(apperr.CodeSourcesInUse, apperr.KindConflict)

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
	// Newest-first by UUIDv7 id (create time). The Source's last activity is not
	// here: List reads it from the display cache (Graph.Activity), so no list
	// load scans the audit history.
	// has_artifact gates the Evidence graph zone without a per-row query.
	sqlList = `SELECT s.id, s.ref, s.source_type_id, s.title, COALESCE(s.description, ''),
		s.cover_mode, s.primary_artifact_id,
		EXISTS (SELECT 1 FROM artifacts a WHERE a.source_id = s.id)
		FROM sources s
		ORDER BY s.id DESC`
	sqlCount      = `SELECT COUNT(*) FROM sources`
	sqlDelete     = `DELETE FROM sources WHERE id = ?`
	sqlTypeExists = `SELECT 1 FROM source_types WHERE id = ?`
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
	// Updated is the latest audit revision of any work scoped to this Source
	// (its row, children, and evidence graph) and when it was recorded.
	// Filled by List and by AttachActivity; zero means unset.
	Updated effects.Activity
	// HasArtifact is true when at least one artifacts row exists for this Source.
	// Filled by List only; other reads leave it false and callers query separately.
	HasArtifact bool
}

// CreateInput is the mutable fields for a new Source.
type CreateInput struct {
	SourceTypeID []byte
	Title        string
	Description  string
}

// Create inserts a Source and mints SRC-…. The caller records the returned changes.
func Create(tx *database.Tx, userID []byte, in CreateInput) (Source, []rowchange.Change, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if tx == nil || len(in.SourceTypeID) != 16 || in.Title == "" {
		return Source{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Source{}, nil, err
	}
	if err := requireType(tx.Tx, in.SourceTypeID); err != nil {
		return Source{}, nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Source{}, nil, err
	}
	idBytes := id[:]

	sourceRef, err := database.InsertWithRef(database.Minter(ref.PrefixSource), func(storedRef string) error {
		_, err := tx.Exec(sqlInsert, idBytes, storedRef, in.SourceTypeID, in.Title, database.NullString(in.Description))
		return err
	})
	if errors.Is(err, database.ErrRefsExhausted) {
		return Source{}, nil, ErrInvalid
	}
	if err != nil {
		return Source{}, nil, mapConstraint(err)
	}

	fields := map[string]rowchange.FieldDiff{
		"id":             {Old: nil, New: id.String()},
		"ref":            {Old: nil, New: sourceRef},
		"source_type_id": {Old: nil, New: database.UUIDString(in.SourceTypeID)},
		"title":          {Old: nil, New: in.Title},
	}
	if in.Description != "" {
		fields["description"] = rowchange.FieldDiff{Old: nil, New: in.Description}
	}
	_, graphChanges, err := evidencegraphs.CreateDefault(tx, idBytes)
	if err != nil {
		return Source{}, nil, err
	}
	changes := append([]rowchange.Change{{
		EntityType: "source",
		EntityID:   append([]byte(nil), idBytes...),
		Action:     rowchange.ActionCreate,
		Fields:     fields,
	}}, graphChanges...)
	return Source{
		ID:           append([]byte(nil), idBytes...),
		Ref:          sourceRef,
		SourceTypeID: append([]byte(nil), in.SourceTypeID...),
		Title:        in.Title,
		Description:  in.Description,
		CoverMode:    CoverModeTypeIcon,
	}, changes, nil
}

// NormalizeCover checks a cover choice. Artifact mode needs the cover's
// thumbnail to exist already (SetSourceCover makes it before its session), so
// call this before SetCover's Run.
func NormalizeCover(c *database.Catalog, sourceID []byte, mode string, primaryArtifactID []byte) (string, []byte, error) {
	if c == nil || len(sourceID) != 16 {
		return "", nil, ErrInvalid
	}
	mode = strings.TrimSpace(mode)
	switch mode {
	case CoverModeTypeIcon:
		return mode, nil, nil
	case CoverModeArtifact:
		if len(primaryArtifactID) != 16 {
			return "", nil, ErrInvalid
		}
		if err := requireRasterCoverArtifact(c, sourceID, primaryArtifactID); err != nil {
			return "", nil, err
		}
		return mode, append([]byte(nil), primaryArtifactID...), nil
	default:
		return "", nil, ErrInvalid
	}
}

// SetCover pins an Artifact as cover or reverts to the Source type icon.
// Call NormalizeCover first when mode is CoverModeArtifact. An unchanged cover
// returns the existing Source and no changes.
func SetCover(tx *database.Tx, userID, sourceID []byte, mode string, primaryArtifactID []byte) (Source, []rowchange.Change, error) {
	if tx == nil || len(sourceID) != 16 {
		return Source{}, nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return Source{}, nil, err
	}
	mode = strings.TrimSpace(mode)
	switch mode {
	case CoverModeTypeIcon:
		primaryArtifactID = nil
	case CoverModeArtifact:
		if len(primaryArtifactID) != 16 {
			return Source{}, nil, ErrInvalid
		}
	default:
		return Source{}, nil, ErrInvalid
	}

	prev, err := getTx(tx.Tx, sourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return Source{}, nil, ErrInvalid
	}
	if err != nil {
		return Source{}, nil, err
	}
	if prev.CoverMode == mode && bytes.Equal(prev.PrimaryArtifactID, primaryArtifactID) {
		return prev, nil, nil
	}
	if _, err := tx.Exec(sqlSetCover, mode, database.NullBlob(primaryArtifactID), sourceID); err != nil {
		return Source{}, nil, mapConstraint(err)
	}
	fields := map[string]rowchange.FieldDiff{
		"cover_mode": {Old: prev.CoverMode, New: mode},
	}
	if !bytes.Equal(prev.PrimaryArtifactID, primaryArtifactID) {
		fields["primary_artifact_id"] = rowchange.FieldDiff{
			Old: database.UUIDJSON(prev.PrimaryArtifactID),
			New: database.UUIDJSON(primaryArtifactID),
		}
	}
	prev.CoverMode = mode
	if len(primaryArtifactID) == 16 {
		prev.PrimaryArtifactID = append([]byte(nil), primaryArtifactID...)
	} else {
		prev.PrimaryArtifactID = nil
	}
	return prev, []rowchange.Change{{
		EntityType: "source",
		EntityID:   append([]byte(nil), sourceID...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}}, nil
}

// Update patches title, description, and source_type_id. An unchanged row
// returns no changes.
func Update(tx *database.Tx, userID []byte, s Source) ([]rowchange.Change, error) {
	if tx == nil {
		return nil, ErrInvalid
	}
	s.Title = strings.TrimSpace(s.Title)
	s.Description = strings.TrimSpace(s.Description)
	if len(s.ID) != 16 || len(s.SourceTypeID) != 16 || s.Title == "" {
		return nil, ErrInvalid
	}
	if err := database.RequireUserID(userID, ErrInvalid); err != nil {
		return nil, err
	}
	prev, err := getTx(tx.Tx, s.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if err := requireType(tx.Tx, s.SourceTypeID); err != nil {
		return nil, err
	}

	fields := map[string]rowchange.FieldDiff{}
	if !bytes.Equal(prev.SourceTypeID, s.SourceTypeID) {
		fields["source_type_id"] = rowchange.FieldDiff{
			Old: database.UUIDString(prev.SourceTypeID),
			New: database.UUIDString(s.SourceTypeID),
		}
	}
	if prev.Title != s.Title {
		fields["title"] = rowchange.FieldDiff{Old: prev.Title, New: s.Title}
	}
	if prev.Description != s.Description {
		fields["description"] = rowchange.FieldDiff{Old: database.NullString(prev.Description), New: database.NullString(s.Description)}
	}
	if len(fields) == 0 {
		return nil, nil
	}
	if _, err := tx.Exec(sqlUpdate, s.SourceTypeID, s.Title, database.NullString(s.Description), s.ID); err != nil {
		return nil, mapConstraint(err)
	}
	return []rowchange.Change{{
		EntityType: "source",
		EntityID:   append([]byte(nil), s.ID...),
		Action:     rowchange.ActionUpdate,
		Fields:     fields,
	}}, nil
}

// Delete erases a Source when inbound resources are empty. Notes, metadata,
// credibility, and layout CASCADE. Artifacts or Subjects block with ErrInUse.
// The returned changes include the released facet rows.
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

	report, err := deleteimpact.Impact(tx.Tx, catalogmodel.KindSource, id)
	if err != nil {
		return nil, err
	}
	if err := deleteimpact.Refuse(report, deleteimpact.Codes{
		InUse: ErrInUse, NotFound: ErrInvalid,
	}); err != nil {
		return nil, err
	}
	released, err := deleteimpact.ReleaseFacets(tx, catalogmodel.KindSource, id)
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sqlDelete, id); err != nil {
		return nil, err
	}

	fields := map[string]rowchange.FieldDiff{
		"id":             {Old: database.UUIDString(id), New: nil},
		"ref":            {Old: existing.Ref, New: nil},
		"source_type_id": {Old: database.UUIDString(existing.SourceTypeID), New: nil},
		"title":          {Old: existing.Title, New: nil},
	}
	if existing.Description != "" {
		fields["description"] = rowchange.FieldDiff{Old: existing.Description, New: nil}
	}
	return append(released, rowchange.Change{
		EntityType: "source",
		EntityID:   append([]byte(nil), id...),
		Action:     rowchange.ActionDelete,
		Fields:     fields,
	}), nil
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

// getByRef returns a Source by SRC-… ref, or sql.ErrNoRows.
func getByRef(c *database.Catalog, sourceRef string) (Source, error) {
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

// List returns Sources newest-first by UUIDv7 id, with HasArtifact filled from
// a single list query and Updated read from the display cache.
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
		s, err := scanSourceList(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// The scan is finished: the catalog has one connection, so the cache's
	// first load cannot run while rows is open.
	for i := range out {
		if out[i].Updated, err = c.Graph().Activity(effects.ScopeSource, out[i].ID); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// AttachActivity sets Updated from the display cache's last activity.
func AttachActivity(c *database.Catalog, s *Source) error {
	if s == nil || len(s.ID) != 16 {
		return ErrInvalid
	}
	if s.Updated.Revision > 0 {
		return nil
	}
	var err error
	s.Updated, err = c.Graph().Activity(effects.ScopeSource, s.ID)
	return err
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

func scanSourceList(row rowScanner) (Source, error) {
	var s Source
	var primary []byte
	if err := row.Scan(
		&s.ID, &s.Ref, &s.SourceTypeID, &s.Title, &s.Description, &s.CoverMode, &primary,
		&s.HasArtifact,
	); err != nil {
		return Source{}, err
	}
	s.PrimaryArtifactID = primary
	if s.CoverMode == "" {
		s.CoverMode = CoverModeTypeIcon
	}
	return s, nil
}

// GetTx returns a Source by id on an open transaction, or sql.ErrNoRows.
func GetTx(tx *sql.Tx, id []byte) (Source, error) {
	if len(id) != 16 {
		return Source{}, ErrInvalid
	}
	return getTx(tx, id)
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

// CoverArtifactFile returns the File id of the Artifact a cover would pin: the
// Artifact must exist, belong to the Source and hold a File, else ErrInvalid.
// It decodes nothing; it is the one place the cover rule lives, for
// SetSourceCover's stage and NormalizeCover alike.
func CoverArtifactFile(c *database.Catalog, sourceID, artifactID []byte) ([]byte, error) {
	art, err := artifacts.Get(c, artifactID)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, artifacts.ErrInvalid) {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(art.SourceID, sourceID) || len(art.FileID) != 16 {
		return nil, ErrInvalid
	}
	return art.FileID, nil
}

// requireRasterCoverArtifact accepts a cover Artifact whose File already has a
// thumbnail. It never decodes: the thumbnail is made before, outside the
// session, so an Artifact with no thumbnail (a PDF, an image that failed to
// decode, one deleted meanwhile) is ErrInvalid.
func requireRasterCoverArtifact(c *database.Catalog, sourceID, artifactID []byte) error {
	fileID, err := CoverArtifactFile(c, sourceID, artifactID)
	if err != nil {
		return err
	}
	link, err := filederivatives.Lookup(c, fileID, filederivatives.TypeThumbnail)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalid
	}
	if err != nil {
		return err
	}
	if len(link.DerivedFileID) != 16 {
		return ErrInvalid
	}
	return nil
}

func mapConstraint(err error) error {
	if database.IsConstraintViolation(err) {
		return ErrInvalid
	}
	return err
}
