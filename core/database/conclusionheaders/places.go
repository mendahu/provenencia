package conclusionheaders

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
)

// PlaceHeader is one Place as a row. Names are the kept toponyms in rank
// order (a multi-valued Property: Montréal and Montreal are both names).
// Period, kind, and parents stay empty until S9-38 and S9-39.
type PlaceHeader struct {
	Entity    canonicalentities.Entity
	Names     []string
	StartDate *datevalues.Value
	EndDate   *datevalues.Value
	Kind      string
	Parents   []string
}

// Unmerged Place handles. Order is sortByTitle's (R5), applied after the
// scan.
const (
	sqlPlacesSelect = `SELECT e.id, e.subject_type_id, e.ref, COALESCE(e.argument, ''), COALESCE(e.label, ''),
		(SELECT json_group_array(tn.value_text ORDER BY tn.rank)
			FROM auto_reconciler_values tn
			WHERE tn.entity_id = e.id AND tn.property_id = tp.id AND tn.reason = 'kept')
	FROM canonical_entities e
	JOIN subject_types st ON st.id = e.subject_type_id
	LEFT JOIN properties tp ON tp.key = 'toponym' AND tp.origin = 'provenencia'
	WHERE st.key = 'place' AND st.origin = 'provenencia' AND e.merged_into_id IS NULL`
)

// ListPlaces returns every Place's header in list order, in one query.
func ListPlaces(q Querier) ([]PlaceHeader, error) {
	return queryPlaces(q, sqlPlacesSelect)
}

// PlacesByIDs returns the headers of the given unmerged Places in list
// order, in one query. Unknown, merged, and non-Place ids are absent.
func PlacesByIDs(q Querier, ids [][]byte) ([]PlaceHeader, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	return queryPlaces(q, sqlPlacesSelect+` AND e.id IN (`+database.SQLInPlaceholders(len(ids))+`)`,
		database.BlobArgs(ids)...)
}

func queryPlaces(q Querier, query string, args ...any) ([]PlaceHeader, error) {
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PlaceHeader
	for rows.Next() {
		h, err := scanPlace(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortByTitle(out, placeTitle, func(h PlaceHeader) string { return h.Entity.Ref })
	return out, nil
}

func scanPlace(rows *sql.Rows) (PlaceHeader, error) {
	var (
		h         PlaceHeader
		namesJSON sql.NullString
	)
	e := &h.Entity
	if err := rows.Scan(
		&e.ID, &e.SubjectTypeID, &e.Ref, &e.Argument, &e.Label,
		&namesJSON,
	); err != nil {
		return PlaceHeader{}, err
	}
	if namesJSON.Valid && namesJSON.String != "" && namesJSON.String != "null" {
		if err := json.Unmarshal([]byte(namesJSON.String), &h.Names); err != nil {
			return PlaceHeader{}, err
		}
		trimmed := h.Names[:0]
		for _, name := range h.Names {
			name = strings.TrimSpace(name)
			if name != "" {
				trimmed = append(trimmed, name)
			}
		}
		h.Names = trimmed
	}
	return h, nil
}
