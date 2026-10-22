// Package graphprogress aggregates subject, observation, primary, and filed
// counts per Evidence graph for the Sources list.
package graphprogress

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/apperr"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var ErrInvalid = apperr.New(apperr.CodeSourcesInvalid, apperr.KindUser)

const (
	// sqlCountSubjectsByType reads subjects once per call: per graph and
	// type, how many subjects exist and how many carry an accepted claim
	// (the unique accepted-claim index makes the second count the filed one).
	// The composite foreign key makes s.source_id the graph's Source.
	sqlCountSubjectsByType = `SELECT s.graph_id, s.source_id, t.key, t.origin, COUNT(*), COUNT(ic.subject_id)
		FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		GROUP BY s.graph_id, s.source_id, t.key, t.origin`

	sqlCountSubjectsByTypeOne = `SELECT t.key, t.origin, COUNT(*), COUNT(ic.subject_id)
		FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		WHERE s.graph_id = ?
		GROUP BY t.key, t.origin`

	sqlCountObservationsByGraph = `SELECT s.graph_id, s.source_id, COUNT(*)
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		GROUP BY s.graph_id, s.source_id`

	sqlGraphSource = `SELECT source_id FROM evidence_graphs WHERE id = ?`

	sqlCountObservationsOne = `SELECT COUNT(*)
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		WHERE s.graph_id = ?`
)

// Progress is graph-progress counts for one Evidence graph.
type Progress struct {
	GraphID []byte
	// SourceID is the graph's Source; empty when the graph does not exist.
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
	if vocabkey.IsPrimaryType(key, origin) {
		p.PrimaryCount += n
		p.FiledCount += filed
	}
}

// List returns per-graph counts, in no fixed order. Graphs with all zeros are
// omitted.
func List(c *database.Catalog) ([]Progress, error) {
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	byID := map[string]*Progress{}
	if err := scanSubjects(db, byID); err != nil {
		return nil, err
	}
	if err := scanGrouped(db, sqlCountObservationsByGraph, func(p *Progress, n int32) {
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

// Get returns counts for one Evidence graph. An unknown or empty graph is zeros.
func Get(c *database.Catalog, graphID []byte) (Progress, error) {
	db, err := c.DB()
	if err != nil {
		return Progress{}, err
	}
	if len(graphID) != 16 {
		return Progress{}, ErrInvalid
	}
	p := Progress{GraphID: append([]byte(nil), graphID...)}
	if err := db.QueryRow(sqlGraphSource, graphID).Scan(&p.SourceID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Progress{}, err
	}
	if err := scanSubjectsOne(db, graphID, &p); err != nil {
		return Progress{}, err
	}
	if err := db.QueryRow(sqlCountObservationsOne, graphID).Scan(&p.ObservationCount); err != nil {
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
		var graphID, sourceID []byte
		var n int32
		if err := rows.Scan(&graphID, &sourceID, &n); err != nil {
			return err
		}
		key, err := uuid.FromBytes(graphID)
		if err != nil {
			return err
		}
		p := byID[key.String()]
		if p == nil {
			p = &Progress{GraphID: append([]byte(nil), graphID...), SourceID: append([]byte(nil), sourceID...)}
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
		var graphID, sourceID []byte
		var key, origin string
		var n, filed int32
		if err := rows.Scan(&graphID, &sourceID, &key, &origin, &n, &filed); err != nil {
			return err
		}
		// A graph whose only subject is its Source's reification has no entry.
		if key == "source" {
			continue
		}
		id, err := uuid.FromBytes(graphID)
		if err != nil {
			return err
		}
		p := byID[id.String()]
		if p == nil {
			p = &Progress{GraphID: append([]byte(nil), graphID...), SourceID: append([]byte(nil), sourceID...)}
			byID[id.String()] = p
		}
		p.add(key, origin, n, filed)
	}
	return rows.Err()
}

func scanSubjectsOne(db *sql.DB, graphID []byte, p *Progress) error {
	rows, err := db.Query(sqlCountSubjectsByTypeOne, graphID)
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
