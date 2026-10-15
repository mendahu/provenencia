// Package sourcegraphprogress aggregates Evidence-graph subject, observation,
// primary, and filed counts per Source for the Sources list.
package sourcegraphprogress

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/promote"
)

var ErrInvalid = apperr.New(apperr.CodeSourcesInvalid, apperr.KindUser)

const (
	// sqlCountSubjectsByType reads subjects once per call: per Source and
	// type, how many subjects exist and how many carry an accepted claim
	// (the unique accepted-claim index makes the second count the filed one).
	sqlCountSubjectsByType = `SELECT s.source_id, t.key, t.origin, COUNT(*), COUNT(ic.subject_id)
		FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		GROUP BY s.source_id, t.key, t.origin`

	sqlCountSubjectsByTypeOne = `SELECT t.key, t.origin, COUNT(*), COUNT(ic.subject_id)
		FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		WHERE s.source_id = ?
		GROUP BY t.key, t.origin`

	sqlCountObservationsBySource = `SELECT s.source_id, COUNT(*)
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		GROUP BY s.source_id`

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
	// PrimaryCount is the subjects of a Promote primary kind (person, event,
	// place); FiledCount is those with an accepted Identity Claim.
	PrimaryCount int32
	FiledCount   int32
}

// add folds one grouped subjects row into p.
func (p *Progress) add(key, origin string, n, filed int32) {
	if key != "source" {
		p.SubjectCount += n
	}
	if promote.PrimaryKind(key, origin) {
		p.PrimaryCount += n
		p.FiledCount += filed
	}
}

// List returns per-Source counts. Sources with all zeros are omitted.
func List(c *database.Catalog) ([]Progress, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	byID := map[string]*Progress{}
	if err := scanSubjects(db, byID); err != nil {
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
	if err := scanSubjectsOne(db, sourceID, &p); err != nil {
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

func scanSubjects(db *sql.DB, byID map[string]*Progress) error {
	rows, err := db.Query(sqlCountSubjectsByType)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sourceID []byte
		var key, origin string
		var n, filed int32
		if err := rows.Scan(&sourceID, &key, &origin, &n, &filed); err != nil {
			return err
		}
		// A Source whose only subject is its own reification has no entry.
		if key == "source" {
			continue
		}
		id, err := uuid.FromBytes(sourceID)
		if err != nil {
			return err
		}
		p := byID[id.String()]
		if p == nil {
			p = &Progress{SourceID: append([]byte(nil), sourceID...)}
			byID[id.String()] = p
		}
		p.add(key, origin, n, filed)
	}
	return rows.Err()
}

func scanSubjectsOne(db *sql.DB, sourceID []byte, p *Progress) error {
	rows, err := db.Query(sqlCountSubjectsByTypeOne, sourceID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var key, origin string
		var n, filed int32
		if err := rows.Scan(&key, &origin, &n, &filed); err != nil {
			return err
		}
		p.add(key, origin, n, filed)
	}
	return rows.Err()
}
