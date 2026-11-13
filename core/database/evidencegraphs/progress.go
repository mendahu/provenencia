package evidencegraphs

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/vocabkey"
)

const (
	// sqlCountSubjectsByType reads subjects: per graph and type, how many
	// subjects exist and how many carry an accepted claim (the unique
	// accepted-claim index makes the second count the filed one).
	sqlCountSubjectsByType = `SELECT s.graph_id, t.key, t.origin, COUNT(*), COUNT(ic.subject_id)
		FROM subjects s
		JOIN subject_types t ON t.id = s.subject_type_id
		LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'`
	sqlCountSubjectsGroup = ` GROUP BY s.graph_id, t.key, t.origin`

	sqlCountObservationsByGraph = `SELECT s.graph_id, COUNT(*)
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id`
	sqlCountObservationsGroup = ` GROUP BY s.graph_id`
)

// Progress is graph-progress counts for one Evidence graph.
type Progress struct {
	GraphID          []byte
	SubjectCount     int32
	ObservationCount int32
	// FiledCount is the subjects of a Promote primary kind (person, event,
	// place) with an accepted Identity Claim.
	FiledCount int32
}

// add folds one grouped subjects row into p.
func (p *Progress) add(key, origin string, n, filed int32) {
	if key != "source" {
		p.SubjectCount += n
	}
	if vocabkey.IsPrimaryType(key, origin) {
		p.FiledCount += filed
	}
}

// ListProgress returns per-graph counts in List's order. Graphs with all zeros
// are omitted.
func ListProgress(c *database.Catalog) ([]Progress, error) {
	idx, err := catalogIndex(c)
	if err != nil {
		return nil, err
	}
	return idx.AllProgress(), nil
}

// GetProgress returns counts for one Evidence graph. An unknown or empty graph
// is zeros.
func GetProgress(c *database.Catalog, graphID []byte) (Progress, error) {
	if _, err := c.DB(); err != nil {
		return Progress{}, err
	}
	if len(graphID) != 16 {
		return Progress{}, ErrInvalid
	}
	idx, err := IndexOf(c.Graph())
	if err != nil {
		return Progress{}, err
	}
	return idx.Progress(graphID), nil
}

// loadProgress counts the named graphs, or every graph when graphIDs is nil.
// A graph with no counted Subject and no Observation has no entry; a graph
// whose only Subject is its Source's reification has none either.
func loadProgress(q database.Querier, graphIDs [][]byte) (map[string]Progress, error) {
	byID := map[string]*Progress{}
	entry := func(graphID []byte) *Progress {
		p := byID[string(graphID)]
		if p == nil {
			p = &Progress{GraphID: append([]byte(nil), graphID...)}
			byID[string(graphID)] = p
		}
		return p
	}
	scanSubjects := func(rows *sql.Rows) error {
		var graphID []byte
		var key, origin string
		var n, filed int32
		if err := rows.Scan(&graphID, &key, &origin, &n, &filed); err != nil {
			return err
		}
		if key == "source" {
			return nil
		}
		entry(graphID).add(key, origin, n, filed)
		return nil
	}
	scanObservations := func(rows *sql.Rows) error {
		var graphID []byte
		var n int32
		if err := rows.Scan(&graphID, &n); err != nil {
			return err
		}
		entry(graphID).ObservationCount = n
		return nil
	}
	if graphIDs == nil {
		if err := sqlbatch.Scan(q, sqlCountSubjectsByType+sqlCountSubjectsGroup, nil, scanSubjects); err != nil {
			return nil, err
		}
		if err := sqlbatch.Scan(q, sqlCountObservationsByGraph+sqlCountObservationsGroup, nil, scanObservations); err != nil {
			return nil, err
		}
	} else {
		if err := sqlbatch.Query(q, sqlCountSubjectsByType+` WHERE s.graph_id IN (`, `)`+sqlCountSubjectsGroup, nil, graphIDs, scanSubjects); err != nil {
			return nil, err
		}
		if err := sqlbatch.Query(q, sqlCountObservationsByGraph+` WHERE s.graph_id IN (`, `)`+sqlCountObservationsGroup, nil, graphIDs, scanObservations); err != nil {
			return nil, err
		}
	}
	out := make(map[string]Progress, len(byID))
	for k, p := range byID {
		out[k] = *p
	}
	return out, nil
}
