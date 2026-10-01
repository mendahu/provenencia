// Package claimconfidencegrades accesses claim_confidence_grades vocabulary:
// the Claim confidence scale on Identity (and later Reconciliation) Claims.
// Not Source credibility — same shape, different judgment.
package claimconfidencegrades

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeClaimConfidenceInvalid, apperr.KindUser)

const (
	OriginProvenencia = "provenencia"
	OriginUser        = "user"

	sqlUpsert = `INSERT INTO claim_confidence_grades (id, key, origin, label, sort_order)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(key, origin) DO UPDATE SET
			label = excluded.label,
			sort_order = excluded.sort_order`
	sqlLookup = `SELECT id, key, origin, label, sort_order
		FROM claim_confidence_grades WHERE key = ? AND origin = ?`
	sqlGetByID = `SELECT id, key, origin, label, sort_order
		FROM claim_confidence_grades WHERE id = ?`
	sqlList = `SELECT id, key, origin, label, sort_order
		FROM claim_confidence_grades
		ORDER BY sort_order ASC, label COLLATE NOCASE, key`
)

// Grade is one claim_confidence_grades row.
type Grade struct {
	ID        []byte
	Key       string
	Origin    string
	Label     string
	SortOrder int
}

// Upsert inserts or updates by (key, origin). Mints a UUIDv7 id when ID is empty on insert.
func Upsert(c *database.Catalog, g Grade) ([]byte, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	g.Key = strings.TrimSpace(g.Key)
	g.Origin = strings.TrimSpace(g.Origin)
	g.Label = strings.TrimSpace(g.Label)
	if g.Key == "" || g.Label == "" || !originOK(g.Origin) {
		return nil, ErrInvalid
	}
	id := g.ID
	if len(id) == 0 {
		existing, err := Lookup(c, g.Key, g.Origin)
		if err == nil {
			id = existing.ID
		} else if errors.Is(err, sql.ErrNoRows) {
			uid, err := uuid.NewV7()
			if err != nil {
				return nil, err
			}
			id = uid[:]
		} else {
			return nil, err
		}
	} else if len(id) != 16 {
		return nil, ErrInvalid
	}
	if _, err := db.Exec(sqlUpsert, id, g.Key, g.Origin, g.Label, g.SortOrder); err != nil {
		return nil, err
	}
	return append([]byte(nil), id...), nil
}

// Lookup returns the grade for (key, origin), or sql.ErrNoRows.
func Lookup(c *database.Catalog, key, origin string) (Grade, error) {
	db, err := c.DB()
	if err != nil {
		return Grade{}, err
	}
	key = strings.TrimSpace(key)
	origin = strings.TrimSpace(origin)
	if key == "" || origin == "" {
		return Grade{}, ErrInvalid
	}
	return scanGrade(db.QueryRow(sqlLookup, key, origin))
}

// GetByID returns a grade by id, or sql.ErrNoRows.
func GetByID(c *database.Catalog, id []byte) (Grade, error) {
	db, err := c.DB()
	if err != nil {
		return Grade{}, err
	}
	if len(id) != 16 {
		return Grade{}, ErrInvalid
	}
	return scanGrade(db.QueryRow(sqlGetByID, id))
}

// List returns all grades ordered by sort_order.
func List(c *database.Catalog) ([]Grade, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(sqlList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Grade
	for rows.Next() {
		g, err := scanGrade(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanGrade(row rowScanner) (Grade, error) {
	var g Grade
	if err := row.Scan(&g.ID, &g.Key, &g.Origin, &g.Label, &g.SortOrder); err != nil {
		return Grade{}, err
	}
	return g, nil
}

func originOK(origin string) bool {
	if origin == OriginProvenencia || origin == OriginUser {
		return true
	}
	return strings.HasPrefix(origin, "plugin:") && len(origin) > len("plugin:")
}
