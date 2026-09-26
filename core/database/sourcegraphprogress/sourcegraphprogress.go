// Package sourcegraphprogress aggregates Evidence-graph subject and
// observation counts per Source for the Sources list (S8-08).
package sourcegraphprogress

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
)

var ErrInvalid = apperr.New(apperr.CodeSourcesInvalid, apperr.KindUser)

const (
	canvasTypeSQL = `t.key != 'source'`

	sqlCountSubjectsBySource = `SELECT s.source_id, COUNT(*)
		FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		WHERE ` + canvasTypeSQL + `
		GROUP BY s.source_id`

	sqlCountObservationsBySource = `SELECT s.source_id, COUNT(*)
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		GROUP BY s.source_id`

	sqlCountSubjectsOne = `SELECT COUNT(*)
		FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		WHERE s.source_id = ? AND ` + canvasTypeSQL

	sqlCountObservationsOne = `SELECT COUNT(*)
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		WHERE s.source_id = ?`
)

// Progress is graph-progress counts for one Source.
type Progress struct {
	SourceID         []byte
	SubjectCount     int32
	ObservationCount int32
}

// List returns per-Source counts. Sources with all zeros are omitted.
func List(c *database.Catalog) ([]Progress, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	byID := map[string]*Progress{}
	if err := scanGrouped(db, sqlCountSubjectsBySource, func(p *Progress, n int32) {
		p.SubjectCount = n
	}, byID); err != nil {
		return nil, err
	}
	if err := scanGrouped(db, sqlCountObservationsBySource, func(p *Progress, n int32) {
		p.ObservationCount = n
	}, byID); err != nil {
		return nil, err
	}
	out := make([]Progress, 0, len(byID))
	for _, p := range byID {
		out = append(out, *p)
	}
	return out, nil
}

// Get returns counts for one Source. An unknown or empty graph is zeros.
func Get(c *database.Catalog, sourceID []byte) (Progress, error) {
	db, err := c.DB()
	if err != nil {
		return Progress{}, err
	}
	if len(sourceID) != 16 {
		return Progress{}, ErrInvalid
	}
	p := Progress{SourceID: append([]byte(nil), sourceID...)}
	if err := db.QueryRow(sqlCountSubjectsOne, sourceID).Scan(&p.SubjectCount); err != nil {
		return Progress{}, err
	}
	if err := db.QueryRow(sqlCountObservationsOne, sourceID).Scan(&p.ObservationCount); err != nil {
		return Progress{}, err
	}
	return p, nil
}

func scanGrouped(db *sql.DB, query string, set func(*Progress, int32), byID map[string]*Progress) error {
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sourceID []byte
		var n int32
		if err := rows.Scan(&sourceID, &n); err != nil {
			return err
		}
		key, err := uuid.FromBytes(sourceID)
		if err != nil {
			return err
		}
		p := byID[key.String()]
		if p == nil {
			p = &Progress{SourceID: append([]byte(nil), sourceID...)}
			byID[key.String()] = p
		}
		set(p, n)
	}
	return rows.Err()
}
