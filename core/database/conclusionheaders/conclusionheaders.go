// Package conclusionheaders composes the row-shaped header of each canonical
// Person, Event, and Place from the auto-reconciler cache (Spike 9
// R4). One header composer per kind serves every surface that shows a handle
// as a row: lists, Promote's target picker, omnibar hits, later tree nodes.
//
// Headers are composed at read time, never stored, and set-based: a whole
// list is a fixed number of queries whatever its length. Go returns
// structures, including which naming-matrix rule titles an Event
// (eventtitle); the app formats text.
package conclusionheaders

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
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
	// Name is the displayed auto-reconciled name; nil when no member names
	// the Person.
	Name *namevalues.Value
	// NameValueCount is the number of displayed name values: names are one
	// structure, so at most 1. More would read as mixed, the extras as +N.
	NameValueCount int
	// Birth and Death are read off subject-role participations in a birth or
	// death event. Place names are the linked Places' kept toponyms; the
	// parent chain stays empty until S9-39.
	Birth LifeFacts
	Death LifeFacts
}

// LifeFacts is a birth or a death composed from the canonical graph.
type LifeFacts struct {
	// Event is the birth or death event read; nil when none is linked.
	// Its page owns the date's Why. EventCount is how many such events
	// survive; more than one is a disagreement.
	Event      *canonicalentities.Entity
	EventCount int
	// Date is the event's date, else its start date.
	Date      *datevalues.Value
	DateCount int
	Places    []HeaderPlace
}

// HeaderPlace is one Place a walk reached. Names are kept toponyms in rank
// order. No chain. Entity is the Place, whose page owns the names' Why.
type HeaderPlace struct {
	Entity canonicalentities.Entity
	Names  []string
}

// Unmerged Person handles with their displayed (kept rank-1) name and name
// count. Order is sortByTitle's (R5), applied after the scan.
const (
	sqlPersonsSelect = `SELECT e.id, e.subject_type_id, e.ref, COALESCE(e.argument, ''), COALESCE(e.label, ''),
		r.value_name,
		(SELECT COUNT(*) FROM auto_reconciler_values c
			WHERE c.entity_id = e.id AND c.property_id = np.id AND c.reason = 'kept') AS name_values
	FROM canonical_entities e
	JOIN subject_types st ON st.id = e.subject_type_id
	LEFT JOIN properties np ON np.key = 'name' AND np.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values r
		ON r.entity_id = e.id AND r.property_id = np.id AND r.rank = 1 AND r.reason = 'kept'
	WHERE st.key = 'person' AND st.origin = 'provenencia' AND e.merged_into_id IS NULL`
)

// ListPersons returns every Person's header in list order.
func ListPersons(q Querier) ([]PersonHeader, error) {
	return queryPersons(q, sqlPersonsSelect)
}

// PersonsByIDs returns the headers of the given unmerged Persons in list
// order, in one query. Unknown, merged, and non-Person ids are absent.
func PersonsByIDs(q Querier, ids [][]byte) ([]PersonHeader, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	return queryPersons(q, sqlPersonsSelect+` AND e.id IN (`+database.SQLInPlaceholders(len(ids))+`)`,
		database.BlobArgs(ids)...)
}

// personRowsByIDs is PersonsByIDs without the birth and death walk, for an
// Event's subjects.
func personRowsByIDs(q Querier, ids [][]byte) ([]PersonHeader, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	return scanPersons(q, sqlPersonsSelect+` AND e.id IN (`+database.SQLInPlaceholders(len(ids))+`)`,
		database.BlobArgs(ids)...)
}

func queryPersons(q Querier, query string, args ...any) ([]PersonHeader, error) {
	out, err := scanPersons(q, query, args...)
	if err != nil {
		return nil, err
	}
	if err := attachLives(q, out); err != nil {
		return nil, err
	}
	sortByTitle(out, personTitle, func(h PersonHeader) string { return h.Entity.Ref })
	return out, nil
}

func scanPersons(q Querier, query string, args ...any) ([]PersonHeader, error) {
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
		if err := rows.Scan(&e.ID, &e.SubjectTypeID, &e.Ref, &e.Argument, &e.Label, &nameBlob, &h.NameValueCount); err != nil {
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
