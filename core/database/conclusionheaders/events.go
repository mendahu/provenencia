package conclusionheaders

import (
	"database/sql"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// EventType is the displayed (kept rank-1) event_type term.
type EventType struct {
	ID    []byte
	Key   string
	Label string
}

// EventHeader is one Event as a row. A rank-1 date wins over a start/end
// span; the span's counts are still reported. Subjects are the subject-role
// persons, participation ref then person ref. Places are every location's
// kept names; chains stay empty until S9-39.
type EventHeader struct {
	Entity         canonicalentities.Entity
	EventName      string
	EventNameCount int
	EventType      *EventType
	EventTypeCount int
	Date           *datevalues.Value
	DateCount      int
	StartDate      *datevalues.Value
	StartDateCount int
	EndDate        *datevalues.Value
	EndDateCount   int
	Subjects       []EventSubject
	Places         []HeaderPlace
}

// EventSubject is one person on an event through a subject-role participation.
type EventSubject struct {
	Entity         canonicalentities.Entity
	Name           *namevalues.Value
	NameValueCount int
}

// Unmerged Event handles. Dated events first by the date window's sort key,
// else the start date's, then undated by ref (R5).
const (
	sqlEventsSelect = `SELECT e.id, e.subject_type_id, e.ref, COALESCE(e.argument, ''), COALESCE(e.label, ''),
		en.value_text,
		(SELECT COUNT(*) FROM auto_reconciler_values c
			WHERE c.entity_id = e.id AND c.property_id = enp.id AND c.reason = 'kept'),
		et.value_term_id, COALESCE(t.key, ''), COALESCE(t.label, ''),
		(SELECT COUNT(*) FROM auto_reconciler_values c
			WHERE c.entity_id = e.id AND c.property_id = etp.id AND c.reason = 'kept'),
		d.value_date,
		(SELECT COUNT(*) FROM auto_reconciler_values c
			WHERE c.entity_id = e.id AND c.property_id = dp.id AND c.reason = 'kept'),
		sd.value_date,
		(SELECT COUNT(*) FROM auto_reconciler_values c
			WHERE c.entity_id = e.id AND c.property_id = sp.id AND c.reason = 'kept'),
		ed.value_date,
		(SELECT COUNT(*) FROM auto_reconciler_values c
			WHERE c.entity_id = e.id AND c.property_id = ep.id AND c.reason = 'kept')
	FROM canonical_entities e
	JOIN subject_types st ON st.id = e.subject_type_id
	LEFT JOIN properties enp ON enp.key = 'event_name' AND enp.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values en
		ON en.entity_id = e.id AND en.property_id = enp.id AND en.rank = 1 AND en.reason = 'kept'
	LEFT JOIN properties etp ON etp.key = 'event_type' AND etp.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values et
		ON et.entity_id = e.id AND et.property_id = etp.id AND et.rank = 1 AND et.reason = 'kept'
	LEFT JOIN property_terms t ON t.id = et.value_term_id
	LEFT JOIN properties dp ON dp.key = 'date' AND dp.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values d
		ON d.entity_id = e.id AND d.property_id = dp.id AND d.rank = 1 AND d.reason = 'kept'
	LEFT JOIN properties sp ON sp.key = 'start_date' AND sp.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values sd
		ON sd.entity_id = e.id AND sd.property_id = sp.id AND sd.rank = 1 AND sd.reason = 'kept'
	LEFT JOIN properties ep ON ep.key = 'end_date' AND ep.origin = 'provenencia'
	LEFT JOIN auto_reconciler_values ed
		ON ed.entity_id = e.id AND ed.property_id = ep.id AND ed.rank = 1 AND ed.reason = 'kept'
	WHERE st.key = 'event' AND st.origin = 'provenencia' AND e.merged_into_id IS NULL`
	sqlEventsOrder = ` ORDER BY COALESCE(d.sort_key, sd.sort_key) IS NULL, COALESCE(d.sort_key, sd.sort_key), e.ref COLLATE NOCASE`
	sqlListEvents  = sqlEventsSelect + sqlEventsOrder
)

// ListEvents returns every Event's header in list order, in one query.
func ListEvents(q Querier) ([]EventHeader, error) {
	return queryEvents(q, sqlListEvents)
}

// EventsByIDs returns the headers of the given unmerged Events in list
// order, in one query. Unknown, merged, and non-Event ids are absent.
func EventsByIDs(q Querier, ids [][]byte) ([]EventHeader, error) {
	ids = database.UniqueBlobIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	return queryEvents(q, sqlEventsSelect+` AND e.id IN (`+database.SQLInPlaceholders(len(ids))+`)`+sqlEventsOrder,
		database.BlobArgs(ids)...)
}

func queryEvents(q Querier, query string, args ...any) ([]EventHeader, error) {
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EventHeader
	for rows.Next() {
		h, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := attachEventGraph(q, out); err != nil {
		return nil, err
	}
	return out, nil
}

func scanEvent(rows *sql.Rows) (EventHeader, error) {
	var (
		h                            EventHeader
		name                         sql.NullString
		termID                       []byte
		termKey, termLabel           string
		dateBlob, startBlob, endBlob []byte
	)
	e := &h.Entity
	if err := rows.Scan(
		&e.ID, &e.SubjectTypeID, &e.Ref, &e.Argument, &e.Label,
		&name, &h.EventNameCount,
		&termID, &termKey, &termLabel, &h.EventTypeCount,
		&dateBlob, &h.DateCount,
		&startBlob, &h.StartDateCount,
		&endBlob, &h.EndDateCount,
	); err != nil {
		return EventHeader{}, err
	}
	h.EventName = strings.TrimSpace(name.String)
	if termKey != "" {
		h.EventType = &EventType{ID: append([]byte(nil), termID...), Key: termKey, Label: termLabel}
	}
	var err error
	if h.Date, err = unmarshalDate(dateBlob); err != nil {
		return EventHeader{}, err
	}
	if h.StartDate, err = unmarshalDate(startBlob); err != nil {
		return EventHeader{}, err
	}
	if h.EndDate, err = unmarshalDate(endBlob); err != nil {
		return EventHeader{}, err
	}
	if h.Date != nil {
		h.StartDate = nil
		h.EndDate = nil
	}
	return h, nil
}

func unmarshalDate(b []byte) (*datevalues.Value, error) {
	if len(b) == 0 {
		return nil, nil
	}
	v, err := valuecodec.UnmarshalDate(b)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
