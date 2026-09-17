// Package subjectpositions accesses unaudited graph layout for subjects.
// Arranging bubbles is UI state, not a research assertion — no audit.Record.
package subjectpositions

import (
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeSubjectPositionsInvalid, apperr.KindUser)

const (
	sqlSet = `INSERT INTO subject_positions (subject_id, grid_x, grid_y) VALUES (?, ?, ?)
		ON CONFLICT(subject_id) DO UPDATE SET
			grid_x = excluded.grid_x,
			grid_y = excluded.grid_y`
	sqlGet          = `SELECT subject_id, grid_x, grid_y FROM subject_positions WHERE subject_id = ?`
	sqlListBySource = `SELECT p.subject_id, p.grid_x, p.grid_y
		FROM subject_positions p
		INNER JOIN subjects s ON s.id = p.subject_id
		WHERE s.source_id = ?
		ORDER BY p.grid_y, p.grid_x, p.subject_id`
	sqlClear         = `DELETE FROM subject_positions WHERE subject_id = ?`
	sqlSubjectExists = `SELECT 1 FROM subjects WHERE id = ?`
)

// Position is one subject_positions row (placed on the grid).
type Position struct {
	SubjectID []byte
	GridX     int64
	GridY     int64
}

// Set places or moves a subject on the grid. No audit.
func Set(c *database.Catalog, subjectID []byte, gridX, gridY int64) (Position, error) {
	db, err := c.DB()
	if err != nil {
		return Position{}, err
	}
	if len(subjectID) != 16 {
		return Position{}, ErrInvalid
	}
	var one int
	err = db.QueryRow(sqlSubjectExists, subjectID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return Position{}, ErrInvalid
	}
	if err != nil {
		return Position{}, err
	}
	if _, err := db.Exec(sqlSet, subjectID, gridX, gridY); err != nil {
		return Position{}, err
	}
	return Position{
		SubjectID: append([]byte(nil), subjectID...),
		GridX:     gridX,
		GridY:     gridY,
	}, nil
}

// Get returns a stored position, or sql.ErrNoRows when unplaced (tray).
func Get(c *database.Catalog, subjectID []byte) (Position, error) {
	db, err := c.DB()
	if err != nil {
		return Position{}, err
	}
	if len(subjectID) != 16 {
		return Position{}, ErrInvalid
	}
	return scanPosition(db.QueryRow(sqlGet, subjectID))
}

// ListBySource returns placed positions for subjects homed to sourceID.
// Unplaced subjects are omitted (tray is the set difference client-side).
func ListBySource(c *database.Catalog, sourceID []byte) ([]Position, error) {
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
	var out []Position
	for rows.Next() {
		p, err := scanPosition(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Clear removes a position (back to tray). Missing row is success.
func Clear(c *database.Catalog, subjectID []byte) error {
	db, err := c.DB()
	if err != nil {
		return err
	}
	if len(subjectID) != 16 {
		return ErrInvalid
	}
	_, err = db.Exec(sqlClear, subjectID)
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPosition(row rowScanner) (Position, error) {
	var p Position
	if err := row.Scan(&p.SubjectID, &p.GridX, &p.GridY); err != nil {
		return Position{}, err
	}
	return p, nil
}
