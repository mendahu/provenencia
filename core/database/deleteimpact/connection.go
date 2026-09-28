package deleteimpact

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/connectrules"
)

const sqlSubjectObservations = `SELECT o.id, o.ref, o.value_date_id, o.value_name_id, st.key, p.key, p.origin
	FROM observations o
	JOIN subjects s ON s.id = o.subject_id
	JOIN subject_types st ON st.id = s.subject_type_id
	JOIN properties p ON p.id = o.property_id
	WHERE o.subject_id = ?
	ORDER BY o.ref COLLATE NOCASE`

const sqlObservationNotes = `SELECT id, body FROM observation_notes WHERE observation_id = ? ORDER BY rowid`

type subjectObsRow struct {
	id, dateID, nameID []byte
	ref                string
	typeKey, propKey   string
	origin             string
}

// ReleasedFacet is one connection-facet Observation deleted by
// ReleaseConnectionFacets, plus notes that CASCADE with it.
type ReleasedFacet struct {
	ID, DateID, NameID []byte
	Ref                string
	Notes              []ReleasedNote
}

// ReleasedNote is one observation_notes row captured before CASCADE.
type ReleasedNote struct {
	ID   []byte
	Body string
}

func subjectResourceInbound() inboundEdge {
	return inboundEdge{
		Parent: KindSubject,
		Via:    "observations.subject_id",
		Child:  KindObservation,
		Bucket: BucketResource,
		Count: func(tx *sql.Tx, parentID []byte) (int, error) {
			rows, err := listSubjectObs(tx, parentID)
			if err != nil {
				return 0, err
			}
			n := 0
			for _, r := range rows {
				if !isConnectionFacet(r) {
					n++
				}
			}
			return n, nil
		},
		List: func(tx *sql.Tx, parentID []byte, limit int) ([]probeRow, error) {
			rows, err := listSubjectObs(tx, parentID)
			if err != nil {
				return nil, err
			}
			var out []probeRow
			for _, r := range rows {
				if isConnectionFacet(r) {
					continue
				}
				out = append(out, probeRow{ID: r.id, Ref: r.ref})
				if len(out) == limit {
					break
				}
			}
			return out, nil
		},
	}
}

func listSubjectObs(tx *sql.Tx, subjectID []byte) ([]subjectObsRow, error) {
	rows, err := tx.Query(sqlSubjectObservations, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []subjectObsRow
	for rows.Next() {
		var r subjectObsRow
		var dateID, nameID []byte
		if err := rows.Scan(&r.id, &r.ref, &dateID, &nameID, &r.typeKey, &r.propKey, &r.origin); err != nil {
			return nil, err
		}
		r.dateID = append([]byte(nil), dateID...)
		r.nameID = append([]byte(nil), nameID...)
		out = append(out, r)
	}
	return out, rows.Err()
}

func isConnectionFacet(r subjectObsRow) bool {
	return connectrules.IsConnectionFacet(r.typeKey, r.propKey, r.origin)
}

func isEdgeObservation(tx *sql.Tx, observationID []byte) (bool, error) {
	var typeKey, propKey, origin string
	err := tx.QueryRow(`
		SELECT st.key, p.key, p.origin
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		JOIN subject_types st ON st.id = s.subject_type_id
		JOIN properties p ON p.id = o.property_id
		WHERE o.id = ?`, observationID,
	).Scan(&typeKey, &propKey, &origin)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return connectrules.IsEdgePair(typeKey, propKey, origin), nil
}

// ReleaseConnectionFacets deletes edge + disambiguation Observations on a
// bridge subject, then releases their owned outbound values. Endpoints and the
// Citation stay. Official subjects.Delete calls this before DELETE.
func ReleaseConnectionFacets(tx *sql.Tx, subjectID []byte) ([]ReleasedFacet, error) {
	if tx == nil || len(subjectID) != 16 {
		return nil, ErrInvalid
	}
	rows, err := listSubjectObs(tx, subjectID)
	if err != nil {
		return nil, err
	}
	var released []ReleasedFacet
	for _, r := range rows {
		if !isConnectionFacet(r) {
			continue
		}
		notes, err := listObservationNotes(tx, r.id)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(`DELETE FROM observations WHERE id = ?`, r.id); err != nil {
			return nil, err
		}
		if err := deleteValueRow(tx, "date_values", r.dateID); err != nil {
			return nil, err
		}
		if err := deleteValueRow(tx, "name_values", r.nameID); err != nil {
			return nil, err
		}
		released = append(released, ReleasedFacet{
			ID:     append([]byte(nil), r.id...),
			DateID: append([]byte(nil), r.dateID...),
			NameID: append([]byte(nil), r.nameID...),
			Ref:    r.ref,
			Notes:  notes,
		})
	}
	return released, nil
}

func listObservationNotes(tx *sql.Tx, observationID []byte) ([]ReleasedNote, error) {
	rows, err := tx.Query(sqlObservationNotes, observationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReleasedNote
	for rows.Next() {
		var n ReleasedNote
		if err := rows.Scan(&n.ID, &n.Body); err != nil {
			return nil, err
		}
		n.ID = append([]byte(nil), n.ID...)
		out = append(out, n)
	}
	return out, rows.Err()
}
