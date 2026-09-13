package project

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var (
	ErrInvalid = apperr.New(apperr.CodeProjectInvalidMetadata, apperr.KindUser)
	ErrMissing = apperr.New(apperr.CodeProjectMissingMetadata, apperr.KindNotFound)
)

const (
	sqlUpsert = `INSERT INTO project (id, label, created_at, updated_at, updated_by, uuid)
		VALUES (1, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			label = excluded.label,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			updated_by = excluded.updated_by,
			uuid = COALESCE(project.uuid, excluded.uuid)`
	sqlGet = `SELECT label, created_at, updated_at, updated_by, uuid FROM project WHERE id = 1`
	sqlListMissingUUID = `SELECT 1 FROM project WHERE id = 1 AND (uuid IS NULL OR length(uuid) = 0)`
	sqlSetUUID         = `UPDATE project SET uuid = ? WHERE id = 1 AND (uuid IS NULL OR length(uuid) = 0)`
)

// Info is the singleton project bookkeeping row.
type Info struct {
	Label     string
	CreatedAt string // RFC3339 UTC
	UpdatedAt string // RFC3339 UTC
	UpdatedBy []byte // users.id
	UUID      []byte // UUIDv7; durable project identity (immutable after mint)
}

// NewID mints a UUIDv7 for project.uuid.
func NewID() ([]byte, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return id[:], nil
}

// Upsert writes the singleton project row. Empty UUID is minted once; an
// existing uuid column is never overwritten on conflict.
func Upsert(c *database.Catalog, info Info) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	info.Label = strings.TrimSpace(info.Label)
	if info.Label == "" || len(info.UpdatedBy) != 16 || info.CreatedAt == "" || info.UpdatedAt == "" {
		return ErrInvalid
	}
	if len(info.UUID) == 0 {
		id, err := NewID()
		if err != nil {
			return err
		}
		info.UUID = id
	} else if len(info.UUID) != 16 {
		return ErrInvalid
	}
	_, err = db.Exec(sqlUpsert, info.Label, info.CreatedAt, info.UpdatedAt, info.UpdatedBy, info.UUID)
	return err
}

// Get returns the singleton project row, or ErrMissing.
func Get(c *database.Catalog) (Info, error) {
	db, err := c.DB()
	if err != nil {
		return Info{}, err
	}
	var info Info
	err = db.QueryRow(sqlGet).Scan(&info.Label, &info.CreatedAt, &info.UpdatedAt, &info.UpdatedBy, &info.UUID)
	if errors.Is(err, sql.ErrNoRows) {
		return Info{}, ErrMissing
	}
	if err != nil {
		return Info{}, err
	}
	return info, nil
}

// EnsureUUID mints project.uuid when the singleton row exists but uuid is NULL.
// Idempotent. No-op when the project row is missing.
func EnsureUUID(c *database.Catalog) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	var one int
	err = db.QueryRow(sqlListMissingUUID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	id, err := NewID()
	if err != nil {
		return err
	}
	_, err = db.Exec(sqlSetUUID, id)
	return err
}

// NowUTC returns an RFC3339 UTC timestamp suitable for created_at/updated_at.
func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// LabelFromDir derives a display label from a *.provenencia directory path when
// the catalog has no project row yet (strip folder suffix; keep historical casing).
func LabelFromDir(projectDir string) string {
	base := projectDir
	if i := strings.LastIndexAny(projectDir, `/\`); i >= 0 {
		base = projectDir[i+1:]
	}
	if strings.HasSuffix(strings.ToLower(base), database.Suffix) {
		base = base[:len(base)-len(database.Suffix)]
	}
	return strings.TrimSpace(base)
}
