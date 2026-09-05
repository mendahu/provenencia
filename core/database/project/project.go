package project

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mendahu/provenencia/core/database"
)

var (
	ErrInvalid = errors.New("Invalid project metadata.")
	ErrMissing = errors.New("Project metadata missing.")
)

const (
	sqlUpsert = `INSERT INTO project (id, label, created_at, updated_at, updated_by)
		VALUES (1, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			label = excluded.label,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			updated_by = excluded.updated_by`
	sqlGet = `SELECT label, created_at, updated_at, updated_by FROM project WHERE id = 1`
)

// Info is the singleton project bookkeeping row.
type Info struct {
	Label     string
	CreatedAt string // RFC3339 UTC
	UpdatedAt string // RFC3339 UTC
	UpdatedBy []byte // users.id
}

// Upsert writes the singleton project row.
func Upsert(c *database.Catalog, info Info) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	info.Label = strings.TrimSpace(info.Label)
	if info.Label == "" || len(info.UpdatedBy) != 16 || info.CreatedAt == "" || info.UpdatedAt == "" {
		return ErrInvalid
	}
	_, err = db.Exec(sqlUpsert, info.Label, info.CreatedAt, info.UpdatedAt, info.UpdatedBy)
	return err
}

// Get returns the singleton project row, or ErrMissing.
func Get(c *database.Catalog) (Info, error) {
	db, err := c.DB()
	if err != nil {
		return Info{}, err
	}
	var info Info
	err = db.QueryRow(sqlGet).Scan(&info.Label, &info.CreatedAt, &info.UpdatedAt, &info.UpdatedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return Info{}, ErrMissing
	}
	if err != nil {
		return Info{}, err
	}
	return info, nil
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
