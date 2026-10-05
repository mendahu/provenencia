// Package conclusionheaders composes the row-shaped header of each canonical
// Person (and later Event and Place) from the resolved-values cache (Spike 9
// R4). One header composer per kind serves every surface that shows a handle
// as a row: lists, Promote's target picker, omnibar hits, later tree nodes.
//
// Headers are composed at read time, never stored, and set-based: a whole
// list is one query whatever its length. Go returns structures (the rank-1
// NameValue, cluster counts, label, ref); the app formats text, including
// the name → label → ref fallback.
package conclusionheaders

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// PersonHeader is one Person as a row.
type PersonHeader struct {
	Entity canonicalentities.Entity
	// Name is the rank-1 resolved name cluster; nil when no member names
	// the Person.
	Name *namevalues.Value
	// NameClusterCount is the number of distinct resolved names. More than
	// one means the members disagree (mixed); the extra clusters are the +N.
	NameClusterCount int
}

// Unmerged Person handles with their rank-1 name row and name cluster count,
// named Persons first by name sort key, then by ref (R5).
const (
	sqlPersonsSelect = `SELECT e.id, e.subject_type_id, e.ref, COALESCE(e.argument, ''), COALESCE(e.label, ''),
		r.value_name,
		(SELECT COUNT(*) FROM conclusion_resolved_values c
			WHERE c.entity_id = e.id AND c.property_id = np.id AND c.reason = 'kept') AS clusters
	FROM canonical_entities e
	JOIN subject_types st ON st.id = e.subject_type_id
	LEFT JOIN properties np ON np.key = 'name' AND np.origin = 'provenencia'
	LEFT JOIN conclusion_resolved_values r
		ON r.entity_id = e.id AND r.property_id = np.id AND r.rank = 1
	WHERE st.key = 'person' AND st.origin = 'provenencia' AND e.merged_into_id IS NULL`
	sqlPersonsOrder = ` ORDER BY r.sort_key IS NULL, r.sort_key, e.ref COLLATE NOCASE`
	sqlListPersons  = sqlPersonsSelect + sqlPersonsOrder
)

// ListPersons returns every Person's header in list order, in one query.
func ListPersons(q Querier) ([]PersonHeader, error) {
	return queryPersons(q, sqlListPersons)
}

// PersonsByIDs returns the headers of the given unmerged Persons in list
// order, in one query. Unknown, merged, and non-Person ids are absent.
func PersonsByIDs(q Querier, ids [][]byte) ([]PersonHeader, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	return queryPersons(q, sqlPersonsSelect+` AND e.id IN (`+database.SQLInPlaceholders(len(ids))+`)`+sqlPersonsOrder,
		database.BlobArgs(ids)...)
}

func queryPersons(q Querier, query string, args ...any) ([]PersonHeader, error) {
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PersonHeader
	for rows.Next() {
		var (
			h        PersonHeader
			nameBlob []byte
		)
		e := &h.Entity
		if err := rows.Scan(&e.ID, &e.SubjectTypeID, &e.Ref, &e.Argument, &e.Label, &nameBlob, &h.NameClusterCount); err != nil {
			return nil, err
		}
		if nameBlob != nil {
			n, err := valuecodec.UnmarshalName(nameBlob)
			if err != nil {
				return nil, err
			}
			h.Name = &n
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
