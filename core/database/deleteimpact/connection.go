package deleteimpact

import (
	"database/sql"
)

const sqlSubjectObservations = `SELECT o.id, o.ref, o.value_date_id, o.value_name_id, st.key, p.key
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
				if !isConnectionFacet(r.typeKey, r.propKey) {
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
				if isConnectionFacet(r.typeKey, r.propKey) {
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
		if err := rows.Scan(&r.id, &r.ref, &dateID, &nameID, &r.typeKey, &r.propKey); err != nil {
			return nil, err
		}
		r.dateID = append([]byte(nil), dateID...)
		r.nameID = append([]byte(nil), nameID...)
		out = append(out, r)
	}
	return out, rows.Err()
}

func isConnectionFacet(typeKey, propKey string) bool {
	if isEdgePair(typeKey, propKey) {
		return true
	}
	return propKey == "role" || propKey == "relationship_type"
}

// Mirrors subjectvocab.EdgeEndpoint for seeded connect rules without importing
// that package (propertyterms → deleteimpact would cycle).
func isEdgePair(typeKey, propKey string) bool {
	switch typeKey {
	case "participation":
		return propKey == "person" || propKey == "event"
	case "relationship":
		return propKey == "person" || propKey == "related_to"
	case "location":
		return propKey == "event" || propKey == "place"
	default:
		return false
	}
}

func isEdgeObservation(tx *sql.Tx, observationID []byte) (bool, error) {
	var typeKey, propKey string
	err := tx.QueryRow(`
		SELECT st.key, p.key
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		JOIN subject_types st ON st.id = s.subject_type_id
		JOIN properties p ON p.id = o.property_id
		WHERE o.id = ?`, observationID,
	).Scan(&typeKey, &propKey)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return isEdgePair(typeKey, propKey), nil
}

// ReleaseConnectionFacets deletes edge + disambiguation Observations on a
// bridge subject, then releases their owned outbound values. Endpoints and the
// Citation stay. Official subjects.Delete calls this in S8-19.
func ReleaseConnectionFacets(tx *sql.Tx, subjectID []byte) error {
	if tx == nil || len(subjectID) != 16 {
		return ErrInvalid
	}
	rows, err := listSubjectObs(tx, subjectID)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if !isConnectionFacet(r.typeKey, r.propKey) {
			continue
		}
		if _, err := tx.Exec(`DELETE FROM observations WHERE id = ?`, r.id); err != nil {
			return err
		}
		if err := deleteValueRow(tx, "date_values", r.dateID); err != nil {
			return err
		}
		if err := deleteValueRow(tx, "name_values", r.nameID); err != nil {
			return err
		}
	}
	return nil
}
