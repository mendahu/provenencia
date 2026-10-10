package searchindex

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusionheaders"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// Handle kinds (Conclusion layer). Documents are match text from the header
// composers: the title the list shows, plus the secondary line (a Person's
// life, a Place's names and chain). autoreconciler.RecomputeTx reprojects
// the handles it rewrites and their header dependents, in the same transaction.
const (
	KindPerson = "person"
	KindEvent  = "event"
	KindPlace  = "place"
)

// handleBatch bounds the IN list per reprojection query.
const handleBatch = 500

// handleKinds are the Conclusion kinds that have a search document.
var handleKinds = map[string]bool{
	KindPerson: true,
	KindEvent:  true,
	KindPlace:  true,
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
		if handleKinds[typeKey] && typeOrigin == "provenencia" {
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
	headers, err := loadHandleHeaders(q, heads)
	if err != nil {
		return err
	}

	for _, id := range ids {
		h, ok := heads[string(id)]
		if !ok {
			// Gone: drop any document of any handle kind.
			for kind := range handleKinds {
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
		if err := Upsert(q, handleDocument(h, values[string(id)], headers)); err != nil {
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

type handleHeaders struct {
	persons map[string]conclusionheaders.PersonHeader
	events  map[string]conclusionheaders.EventHeader
	places  map[string]conclusionheaders.PlaceHeader
}

func loadHandleHeaders(q Querier, heads map[string]handleHead) (handleHeaders, error) {
	var persons, events, places [][]byte
	for key, h := range heads {
		if h.merged || h.kind == "" {
			continue
		}
		id := []byte(key)
		switch h.kind {
		case KindPerson:
			persons = append(persons, id)
		case KindEvent:
			events = append(events, id)
		case KindPlace:
			places = append(places, id)
		}
	}
	out := handleHeaders{
		persons: map[string]conclusionheaders.PersonHeader{},
		events:  map[string]conclusionheaders.EventHeader{},
		places:  map[string]conclusionheaders.PlaceHeader{},
	}
	ps, err := conclusionheaders.PersonsFromQuerier(q, persons)
	if err != nil {
		return handleHeaders{}, err
	}
	for _, h := range ps {
		out.persons[uuidString(h.Entity.ID)] = h
	}
	es, err := conclusionheaders.EventsFromQuerier(q, events)
	if err != nil {
		return handleHeaders{}, err
	}
	for _, h := range es {
		out.events[uuidString(h.Entity.ID)] = h
	}
	pl, err := conclusionheaders.PlacesFromQuerier(q, places)
	if err != nil {
		return handleHeaders{}, err
	}
	for _, h := range pl {
		out.places[uuidString(h.Entity.ID)] = h
	}
	return out, nil
}

// handleDocument is match text from the header the lists use. Other cached
// names and toponyms stay searchable when they are not the displayed title.
func handleDocument(h handleHead, v handleValues, headers handleHeaders) Document {
	title, secondary := matchText(h, v, headers)
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

func matchText(h handleHead, v handleValues, headers handleHeaders) (string, []string) {
	var title string
	var secondary []string
	switch h.kind {
	case KindPerson:
		if p, ok := headers.persons[h.id]; ok {
			if p.Name != nil {
				title = strings.TrimSpace(p.Name.Form)
			}
			secondary = append(secondary, personLifeMatch(p)...)
		}
		secondary = append(secondary, tagged(extraValues(v, "name", title), "name")...)
	case KindEvent:
		if e, ok := headers.events[h.id]; ok {
			title = eventMatchTitle(e)
			secondary = append(secondary, eventMatchSecondary(e)...)
		}
	case KindPlace:
		if p, ok := headers.places[h.id]; ok {
			if len(p.Names) > 0 {
				title = p.Names[0]
			}
			if len(p.Names) > 1 {
				secondary = append(secondary, tagged(p.Names[1:], "toponym")...)
			}
			secondary = append(secondary, p.Parents...)
		}
		secondary = append(secondary, tagged(extraValues(v, "toponym", title), "toponym")...)
	}
	if h.label != "" {
		if title == "" {
			title = h.label
		} else if h.label != title {
			secondary = append(secondary, h.label)
		}
	}
	return title, secondary
}

func tagged(vals []string, tag string) []string {
	if len(vals) == 0 {
		return nil
	}
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = tag + ":\t" + v
	}
	return out
}

func extraValues(v handleValues, key, title string) []string {
	var out []string
	for _, val := range v[key] {
		if val == title {
			continue
		}
		out = append(out, val)
	}
	return out
}

func personLifeMatch(h conclusionheaders.PersonHeader) []string {
	var out []string
	for _, life := range []conclusionheaders.LifeFacts{h.Birth, h.Death} {
		if life.Date != nil {
			if s := datevalues.CompactDisplay(*life.Date); s != "" {
				out = append(out, s)
			}
		}
		for _, p := range life.Places {
			out = append(out, p.Names...)
			out = append(out, p.Parents...)
		}
	}
	return out
}

func eventMatchSecondary(h conclusionheaders.EventHeader) []string {
	var out []string
	for _, d := range []*datevalues.Value{h.Date, h.StartDate, h.EndDate} {
		if d == nil {
			continue
		}
		if s := datevalues.CompactDisplay(*d); s != "" {
			out = append(out, s)
		}
	}
	for _, p := range h.Places {
		out = append(out, p.Names...)
		out = append(out, p.Parents...)
	}
	return out
}

// eventMatchTitle is the English list title, so a query for the words the
// row shows ("Birth of James Robins") finds the Event. The app still formats
// the visible row from the structured header.
func eventMatchTitle(h conclusionheaders.EventHeader) string {
	p := h.Title
	typeWord := strings.TrimSpace(p.Parts.TypeLabel)
	if typeWord == "" {
		typeWord = strings.TrimSpace(p.Parts.TypeKey)
	}
	if typeWord == "" {
		typeWord = "Event"
	}
	subject := func(i int) string {
		if i >= len(p.Parts.Subjects) || p.Parts.Subjects[i].Name == nil {
			return "unnamed person"
		}
		if form := strings.TrimSpace(p.Parts.Subjects[i].Name.Form); form != "" {
			return form
		}
		return "unnamed person"
	}
	switch p.Rule {
	case eventtitle.RuleRecordedName:
		return p.Parts.RecordedName
	case eventtitle.RuleSubject:
		return typeWord + " of " + subject(0)
	case eventtitle.RuleCouple:
		return "Marriage of " + subject(0) + " and " + subject(1)
	case eventtitle.RuleSubjects:
		return typeWord + " of " + subject(0) + " et al."
	case eventtitle.RuleLabel:
		return p.Parts.Label
	case eventtitle.RuleTypeAtPlace:
		return typeWord + " at " + p.Parts.Place
	case eventtitle.RuleType:
		return "Unspecified " + strings.ToLower(typeWord)
	default:
		return p.Parts.Ref
	}
}
