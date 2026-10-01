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

type subjectObsRow struct {
	id, dateID, nameID []byte
	ref                string
	typeKey, propKey   string
	origin             string
}

func subjectResourceInbound() inboundEdge {
	memo := &subjectObsMemo{}
	return inboundEdge{
		Parent: KindSubject,
		Via:    "observations.subject_id",
		Child:  KindObservation,
		Bucket: BucketResource,
		Count: func(tx *sql.Tx, parentID []byte) (int, error) {
			rows, err := memo.list(tx, parentID)
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
			rows, err := memo.list(tx, parentID)
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

type subjectObsMemo struct {
	tx     *sql.Tx
	id     string
	rows   []subjectObsRow
	loaded bool
}

func (m *subjectObsMemo) list(tx *sql.Tx, parentID []byte) ([]subjectObsRow, error) {
	key := string(parentID)
	if m.loaded && m.tx == tx && m.id == key {
		return m.rows, nil
	}
	rows, err := listSubjectObs(tx, parentID)
	if err != nil {
		return nil, err
	}
	m.tx = tx
	m.id = key
	m.rows = rows
	m.loaded = true
	return rows, nil
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
