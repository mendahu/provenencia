package searchindex

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// Handle kinds (Conclusion layer). Their documents are built from the
// auto-reconciler cache, so they stay current wherever the cache does:
// autoreconciler.RecomputeTx reprojects the handles it rewrites, in the same
// transaction. S9-34 replaces these documents with ones built from the
// header composers once they exist (S9-31).
const (
	KindPerson = "person"
	KindEvent  = "event"
	KindPlace  = "place"
)

// handleBatch bounds the IN list per reprojection query.
const handleBatch = 500

// Which cached Property titles each kind, and which feed its secondary text.
var handleText = map[string]struct {
	title     string   // rank-1 value of this Property is the title
	secondary []string // every value of these Properties (title Property's other ranks too)
}{
	KindPerson: {title: "name", secondary: []string{"name"}},
	KindPlace:  {title: "toponym", secondary: []string{"toponym"}},
	KindEvent:  {title: "event_type", secondary: []string{"event_type", "date", "start_date", "end_date"}},
}

const (
	sqlHandleHeads = `SELECT e.id, e.ref, COALESCE(e.label, ''), st.key, st.origin, e.merged_into_id IS NOT NULL
		FROM canonical_entities e JOIN subject_types st ON st.id = e.subject_type_id
		WHERE e.id IN (`
	// Cached values of the Properties handle documents read, terms by label.
	sqlHandleValues = `SELECT r.entity_id, p.key, r.rank, r.value_text, r.value_name, r.value_date, t.label
		FROM auto_reconciler_values r
		JOIN properties p ON p.id = r.property_id AND p.origin = 'provenencia'
		LEFT JOIN property_terms t ON t.id = r.value_term_id
		WHERE p.key IN ('name', 'toponym', 'event_type', 'date', 'start_date', 'end_date')
		  AND r.entity_id IN (`
	sqlAllHandles = `SELECT id FROM canonical_entities ORDER BY id`
)

// ReprojectHandles rewrites the search documents of the given handles from
// the auto-reconciler cache. Merged handles, unknown ids, and handles of
// kinds without a document are removed.
func ReprojectHandles(q Querier, entityIDs [][]byte) error {
	ids := database.UniqueBlobIDs(entityIDs)
	for start := 0; start < len(ids); start += handleBatch {
		end := min(start+handleBatch, len(ids))
		if err := reprojectHandleBatch(q, ids[start:end]); err != nil {
			return err
		}
	}
	return nil
}

// reprojectAllHandles reprojects every handle (RebuildAll).
func reprojectAllHandles(q Querier) error {
	ids, err := listIDs(q, sqlAllHandles)
	if err != nil {
		return err
	}
	return ReprojectHandles(q, ids)
}

type handleHead struct {
	id, ref, label, kind string
	merged               bool
}

// handleText values for one handle: Property key → values in rank order.
type handleValues map[string][]string

func reprojectHandleBatch(q Querier, ids [][]byte) error {
	in := database.SQLInPlaceholders(len(ids)) + `)`
	args := database.BlobArgs(ids)

	heads := map[string]handleHead{}
	rows, err := q.Query(sqlHandleHeads+in, args...)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			id                  []byte
			h                   handleHead
			typeKey, typeOrigin string
		)
		if err := rows.Scan(&id, &h.ref, &h.label, &typeKey, &typeOrigin, &h.merged); err != nil {
			_ = rows.Close()
			return err
		}
		if _, ok := handleText[typeKey]; ok && typeOrigin == "provenencia" {
			h.kind = typeKey
		}
		h.id = uuidString(id)
		heads[string(id)] = h
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return err
	}

	values, err := loadHandleValues(q, in, args)
	if err != nil {
		return err
	}

	for _, id := range ids {
		h, ok := heads[string(id)]
		if !ok {
			// Gone: drop any document of any handle kind.
			for kind := range handleText {
				if err := Delete(q, kind, uuidString(id)); err != nil {
					return err
				}
			}
			continue
		}
		if h.kind == "" {
			continue
		}
		if h.merged {
			if err := Delete(q, h.kind, h.id); err != nil {
				return err
			}
			continue
		}
		if err := Upsert(q, handleDocument(h, values[string(id)])); err != nil {
			return err
		}
	}
	return nil
}

func loadHandleValues(q Querier, in string, args []any) (map[string]handleValues, error) {
	rows, err := q.Query(sqlHandleValues+in+` ORDER BY r.entity_id, p.key, r.rank`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]handleValues{}
	for rows.Next() {
		var (
			entityID           []byte
			key                string
			rank               int
			text, termLabel    sql.NullString
			nameBlob, dateBlob []byte
		)
		if err := rows.Scan(&entityID, &key, &rank, &text, &nameBlob, &dateBlob, &termLabel); err != nil {
			return nil, err
		}
		var v string
		switch {
		case nameBlob != nil:
			n, err := valuecodec.UnmarshalName(nameBlob)
			if err != nil {
				return nil, err
			}
			v = n.Form
		case dateBlob != nil:
			d, err := valuecodec.UnmarshalDate(dateBlob)
			if err != nil {
				return nil, err
			}
			var years []string
			for _, y := range []*int{d.StartYear, d.EndYear} {
				if y != nil {
					years = append(years, strconv.Itoa(*y))
				}
			}
			v = strings.Join(years, " ")
		case termLabel.Valid:
			v = termLabel.String
		case text.Valid:
			v = text.String
		}
		if v = strings.TrimSpace(v); v == "" {
			continue
		}
		k := string(entityID)
		if out[k] == nil {
			out[k] = handleValues{}
		}
		out[k][key] = append(out[k][key], v)
	}
	return out, rows.Err()
}

// handleDocument: title is the kind's rank-1 value, else the working label,
// else the ref; secondary holds every other value. Values are data (names,
// toponyms, term labels, years), never composed sentences.
func handleDocument(h handleHead, v handleValues) Document {
	spec := handleText[h.kind]
	title := ""
	if vals := v[spec.title]; len(vals) > 0 {
		title = vals[0]
	}
	var secondary []string
	for _, key := range spec.secondary {
		for i, val := range v[key] {
			if key == spec.title && i == 0 {
				continue
			}
			secondary = append(secondary, val)
		}
	}
	if h.label != "" {
		if title == "" {
			title = h.label
		} else {
			secondary = append(secondary, h.label)
		}
	}
	display := title
	if display == "" {
		display = h.ref
	}
	return Document{
		Kind:         h.kind,
		EntityID:     h.id,
		DisplayRef:   h.ref,
		DisplayTitle: display,
		Title:        title,
		Ref:          h.ref,
		Secondary:    strings.Join(secondary, "\n"),
	}
}
