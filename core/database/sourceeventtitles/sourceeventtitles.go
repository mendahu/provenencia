// Package sourceeventtitles chooses the title of every Event Subject on one
// Source's Evidence graph, from what that Source cites: the event's own
// event_name and event_type, the people its subject-role participations name,
// and the places its locations name. The rule is eventtitle's, the same one
// canonical Event headers use, so a card and the Event it is promoted to
// title themselves by one matrix.
//
// Set-based: a fixed number of queries for the whole Source.
package sourceeventtitles

import (
	"database/sql"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/eventtitle"
)

// Querier is *sql.Tx or *sql.DB.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

// Title is one Event Subject's chosen title.
type Title struct {
	SubjectID []byte
	Plan      eventtitle.Plan
}

const (
	sqlEvents = `SELECT s.id, s.ref, COALESCE(s.label, '')
		FROM subjects s
		JOIN subject_types st ON st.id = s.subject_type_id AND st.key = 'event' AND st.origin = 'provenencia'
		WHERE s.source_id = ?
		ORDER BY s.ref`

	// The positive Observations a title reads, first-cited (by ref) first.
	sqlNaming = `SELECT o.subject_id, p.key, o.value_text, o.value_name_id, COALESCE(t.key, ''), COALESCE(t.label, '')
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id AND s.source_id = ?
		JOIN properties p ON p.id = o.property_id AND p.origin = 'provenencia'
			AND p.key IN ('event_name', 'event_type', 'name', 'toponym')
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		WHERE o.polarity = 'positive'
		ORDER BY o.ref`
)

type naming struct {
	text     string
	nameID   []byte
	termKey  string
	termText string
}

// ForSource returns a title for every Event Subject on the Source, by ref.
func ForSource(q Querier, sourceID []byte) ([]Title, error) {
	type event struct {
		id         []byte
		ref, label string
	}
	var events []event
	rows, err := q.Query(sqlEvents, sourceID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var e event
		if err := rows.Scan(&e.id, &e.ref, &e.label); err != nil {
			_ = rows.Close()
			return nil, err
		}
		events = append(events, e)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil || len(events) == 0 {
		return nil, err
	}

	first, err := firstNamings(q, sourceID)
	if err != nil {
		return nil, err
	}
	ids := make([][]byte, len(events))
	for i, e := range events {
		ids[i] = e.id
	}
	people, err := canonicalgraph.WalkSource(q, sourceID, canonicalgraph.SubjectsOfEvent, ids)
	if err != nil {
		return nil, err
	}
	places, err := canonicalgraph.WalkSource(q, sourceID, canonicalgraph.PlacesOfEvent, ids)
	if err != nil {
		return nil, err
	}
	var nameIDs [][]byte
	for _, e := range people {
		if n, ok := first[key(e.To, "name")]; ok {
			nameIDs = append(nameIDs, n.nameID)
		}
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return nil, err
	}

	peopleOf := map[string][]eventtitle.Subject{}
	seen := map[string]bool{}
	for _, e := range people {
		pair := string(e.From) + "\x00" + string(e.To)
		if seen[pair] {
			continue
		}
		seen[pair] = true
		var subject eventtitle.Subject
		if n, ok := first[key(e.To, "name")]; ok {
			if v, ok := names[string(n.nameID)]; ok {
				subject.Name = &v
			}
		}
		peopleOf[string(e.From)] = append(peopleOf[string(e.From)], subject)
	}
	// A place's name is its first-cited toponym; the event's place is the
	// first named one by place ref.
	sort.SliceStable(places, func(i, j int) bool { return places[i].ToRef < places[j].ToRef })
	placeOf := map[string]string{}
	for _, e := range places {
		if _, done := placeOf[string(e.From)]; done {
			continue
		}
		if n, ok := first[key(e.To, "toponym")]; ok && n.text != "" {
			placeOf[string(e.From)] = n.text
		}
	}

	out := make([]Title, 0, len(events))
	for _, e := range events {
		parts := eventtitle.Parts{
			Label:    e.label,
			Ref:      e.ref,
			Subjects: peopleOf[string(e.id)],
			Place:    placeOf[string(e.id)],
		}
		if n, ok := first[key(e.id, "event_name")]; ok {
			parts.RecordedName = n.text
		}
		if n, ok := first[key(e.id, "event_type")]; ok {
			parts.TypeKey, parts.TypeLabel = n.termKey, n.termText
		}
		out = append(out, Title{SubjectID: e.id, Plan: eventtitle.Choose(parts)})
	}
	return out, nil
}

func key(subjectID []byte, propertyKey string) string {
	return string(subjectID) + "\x00" + propertyKey
}

// firstNamings is the first-cited usable value of each naming Property for
// every Subject on the Source.
func firstNamings(q Querier, sourceID []byte) (map[string]naming, error) {
	rows, err := q.Query(sqlNaming, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]naming{}
	for rows.Next() {
		var (
			subjectID, nameID []byte
			propertyKey       string
			text              sql.NullString
			n                 naming
		)
		if err := rows.Scan(&subjectID, &propertyKey, &text, &nameID, &n.termKey, &n.termText); err != nil {
			return nil, err
		}
		n.text = strings.TrimSpace(text.String)
		n.nameID = append([]byte(nil), nameID...)
		usable := n.text != "" || len(n.nameID) > 0 || n.termKey != ""
		k := key(subjectID, propertyKey)
		if _, done := out[k]; usable && !done {
			out[k] = n
		}
	}
	return out, rows.Err()
}
