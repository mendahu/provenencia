package conclusionheaders

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
)

// PlaceRelationshipKind is how a related Place appears on a Place detail.
const (
	RelPartOf       = "part_of"
	RelContains     = "contains"
	RelPredecessor  = "predecessor"
	RelSuccessor    = "successor"
)

// PlaceRelationship is one related Place for a detail section (S9-40).
type PlaceRelationship struct {
	Entity    canonicalentities.Entity
	Title     string
	Kind      string
	StartDate *datevalues.Value // membership span, or the other place's period
	EndDate   *datevalues.Value
}

// PlaceHeader is one Place as a row. Names are the kept toponyms in rank
// order (a multi-valued Property: Montréal and Montreal are both names).
// StartDate / EndDate are the Place's period. Parents is today's hierarchical
// chain (nearest first); ParentsAreCandidates is true when several parents
// hold at once (join with "or"). Kind stays empty (place_nature deferred).
// PartOf / Contains / Predecessors / Successors are filled only for detail
// (AttachPlaceRelationships); list headers leave them empty.
type PlaceHeader struct {
	Entity               canonicalentities.Entity
	Names                []string
	StartDate            *datevalues.Value
	EndDate              *datevalues.Value
	Kind                 string
	Parents              []string
	ParentsAreCandidates bool
	PartOf               []PlaceRelationship
	Contains             []PlaceRelationship
	Predecessors         []PlaceRelationship
	Successors           []PlaceRelationship
}

// Unmerged Place handles. Order is sortByTitle's (R5), applied after the
// scan. Period dates are kept rank-1 start_date / end_date.
const (
	sqlPlacesSelect = `SELECT e.id, e.subject_type_id, e.ref, COALESCE(e.argument, ''), COALESCE(e.label, ''),
		(SELECT json_group_array(tn.value_text ORDER BY tn.rank)
			FROM auto_reconciler_values tn
			WHERE tn.entity_id = e.id AND tn.property_id = tp.id AND tn.reason = 'kept'),
		sd.value_date,
		ed.value_date
	FROM canonical_entities e
	JOIN subject_types st ON st.id = e.subject_type_id
	LEFT JOIN properties tp ON tp.key = 'toponym' AND tp.origin = 'provenencia'
	LEFT JOIN properties sp ON sp.key = 'start_date' AND sp.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values sd
		ON sd.entity_id = e.id AND sd.property_id = sp.id AND sd.rank = 1 AND sd.reason = 'kept'
	LEFT JOIN properties ep ON ep.key = 'end_date' AND ep.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values ed
		ON ed.entity_id = e.id AND ed.property_id = ep.id AND ed.rank = 1 AND ed.reason = 'kept'
	WHERE st.key = 'place' AND st.origin = 'provenencia' AND e.merged_into_id IS NULL`
)

// ListPlaces returns every Place's header in list order, in one query.
func ListPlaces(q Querier) ([]PlaceHeader, error) {
	return queryPlaces(q, sqlPlacesSelect)
}

// PlacesByIDs returns the headers of the given unmerged Places in list
// order, one query per database.InBatch ids. Unknown, merged, and non-Place
// ids are absent.
func PlacesByIDs(q Querier, ids [][]byte) ([]PlaceHeader, error) {
	out, err := placesWithoutChains(q, ids)
	if err != nil || len(out) == 0 {
		return nil, err
	}
	return withChains(q, out)
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
	return withChains(q, out)
}

// withChains attaches today's chains and applies the list order.
func withChains(q Querier, out []PlaceHeader) ([]PlaceHeader, error) {
	if err := attachPlaceChains(q, out, TodayDate()); err != nil {
		return nil, err
	}
	sortByTitle(out, placeTitle, func(h PlaceHeader) string { return h.Entity.Ref })
	return out, nil
}

func scanPlace(rows *sql.Rows) (PlaceHeader, error) {
	var (
		h                  PlaceHeader
		namesJSON          sql.NullString
		startBlob, endBlob []byte
	)
	e := &h.Entity
	if err := rows.Scan(
		&e.ID, &e.SubjectTypeID, &e.Ref, &e.Argument, &e.Label,
		&namesJSON, &startBlob, &endBlob,
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
	var err error
	if h.StartDate, err = unmarshalDate(startBlob); err != nil {
		return PlaceHeader{}, err
	}
	if h.EndDate, err = unmarshalDate(endBlob); err != nil {
		return PlaceHeader{}, err
	}
	return h, nil
}

// TodayDate is a Gregorian point for "today's chain" on Place rows.
func TodayDate() datevalues.Value {
	now := time.Now()
	y, m, d := now.Date()
	month, day := int(m), d
	return datevalues.Value{
		Kind: datevalues.KindPoint, Calendar: "gregorian",
		StartYear: &y, StartMonth: &month, StartDay: &day,
	}
}
