// Package promotealign loads catalog inputs for Promote graph alignment and
// calls pure graphalign.Align (S9-42). It does not score pairs or walk policy.
package promotealign

import (
	"bytes"
	"database/sql"
	"errors"

	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// Propose loads the Evidence layer for sourceID, expands a bounded canon
// neighborhood, merges caller fixed with already-promoted anchors, and runs
// graphalign.Align. The int64 is MAX(audit_transactions.revision) at load
// time, the stamp Done sends back. An unknown or empty sourceID is
// promote.ErrInvalid.
func Propose(q Querier, sourceID []byte, fixed []graphalign.Fixed) (graphalign.Proposal, int64, error) {
	if len(sourceID) != 16 {
		return graphalign.Proposal{}, 0, promote.ErrInvalid
	}
	var rev int64
	if err := q.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		return graphalign.Proposal{}, 0, err
	}
	var exists int
	if err := q.QueryRow(`SELECT 1 FROM sources WHERE id = ?`, sourceID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return graphalign.Proposal{}, 0, promote.ErrInvalid
		}
		return graphalign.Proposal{}, 0, err
	}

	layer, primary, err := loadLayer(q, sourceID)
	if err != nil {
		return graphalign.Proposal{}, 0, err
	}
	if len(layer.Subjects) == 0 {
		return graphalign.Proposal{Rows: nil}, rev, nil
	}

	anchors, err := mergeFixed(q, sourceID, primary, fixed)
	if err != nil {
		return graphalign.Proposal{}, 0, err
	}

	canon, err := loadCanon(q, primary, anchors)
	if err != nil {
		return graphalign.Proposal{}, 0, err
	}

	stats, err := loadStats(q)
	if err != nil {
		return graphalign.Proposal{}, 0, err
	}

	cfg := graphalign.DefaultConfig()
	return graphalign.Align(layer, canon, stats, anchors, &cfg), rev, nil
}

// mergeFixed validates caller fixed pairs and adds already-promoted Subjects
// as free anchors. Caller fixed wins when both name the same Subject.
func mergeFixed(q Querier, sourceID []byte, primary map[string]primarySubject, caller []graphalign.Fixed) ([]graphalign.Fixed, error) {
	out := make([]graphalign.Fixed, 0, len(caller)+len(primary))
	seen := map[string]bool{}

	for _, f := range caller {
		if len(f.SubjectID) != 16 || len(f.HandleID) != 16 {
			return nil, promote.ErrInvalid
		}
		sk := string(f.SubjectID)
		ps, ok := primary[sk]
		if !ok {
			return nil, promote.ErrInvalid
		}
		var kind string
		err := q.QueryRow(`SELECT st.key FROM canonical_entities e
			JOIN subject_types st ON st.id = e.subject_type_id
			WHERE e.id = ? AND e.merged_into_id IS NULL`, f.HandleID).Scan(&kind)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, promote.ErrInvalid
		}
		if err != nil {
			return nil, err
		}
		if kind != ps.kind {
			return nil, promote.ErrInvalid
		}
		out = append(out, graphalign.Fixed{
			SubjectID: append([]byte(nil), f.SubjectID...),
			HandleID:  append([]byte(nil), f.HandleID...),
		})
		seen[sk] = true
	}

	rows, err := q.Query(`SELECT s.id, ic.entity_id
		FROM subjects s
		JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		WHERE s.source_id = ?`, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var sid, eid []byte
		if err := rows.Scan(&sid, &eid); err != nil {
			return nil, err
		}
		sk := string(sid)
		if seen[sk] {
			continue
		}
		if _, ok := primary[sk]; !ok {
			continue
		}
		out = append(out, graphalign.Fixed{
			SubjectID: append([]byte(nil), sid...),
			HandleID:  append([]byte(nil), eid...),
		})
		seen[sk] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Stable order: by subject id bytes.
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if bytes.Compare(out[j].SubjectID, out[i].SubjectID) < 0 {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}
