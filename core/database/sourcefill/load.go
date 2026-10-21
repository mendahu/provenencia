// Package sourcefill loads one Source onto the catalog graph: subjects,
// observations, members, and the edges an event title reads. Promote builds
// its own bridge signatures when it reads the layer.
package sourcefill

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/depictionvalues"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func init() {
	graphcache.SetSourceLoader(loadSource)
	graphcache.SetSourceCheck(checkSource)
}

func loadSource(g *graphcache.Graph, sourceID []byte) (graphcache.SourceGraph, error) {
	q := graphcache.Querier(g)
	title, found, err := loadSourceTitle(q, sourceID)
	if err != nil {
		return graphcache.SourceGraph{}, err
	}
	subjects, err := loadSourceSubjects(q, sourceID)
	if err != nil {
		return graphcache.SourceGraph{}, err
	}
	observations, err := loadSourceObservations(q, sourceID)
	if err != nil {
		return graphcache.SourceGraph{}, err
	}
	members, err := loadSourceMembers(q, sourceID)
	if err != nil {
		return graphcache.SourceGraph{}, err
	}
	people, places, err := loadTitleEdges(g, graphcache.SourceGraph{SourceID: sourceID, Subjects: subjects, Observations: observations}, subjects)
	if err != nil {
		return graphcache.SourceGraph{}, err
	}
	return graphcache.SourceGraph{
		Title:        title,
		Found:        found,
		Subjects:     subjects,
		Observations: observations,
		Members:      members,
		EventPeople:  people,
		EventPlaces:  places,
	}, nil
}

// loadSourceTitle reads the Source's title, and whether the Source exists.
func loadSourceTitle(q graphcache.Querier, sourceID []byte) (string, bool, error) {
	var title string
	err := q.QueryRow(`SELECT title FROM sources WHERE id = ?`, sourceID).Scan(&title)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return title, err == nil, err
}

func loadSourceSubjects(q graphcache.Querier, sourceID []byte) ([]graphcache.SourceSubject, error) {
	rows, err := q.Query(`SELECT s.id, s.ref, s.source_id, s.subject_type_id,
			COALESCE(s.label, ''), COALESCE(s.description, ''), st.key, st.origin
		FROM subjects s
		JOIN subject_types st ON st.id = s.subject_type_id
		WHERE s.source_id = ?
		ORDER BY s.label COLLATE NOCASE, s.ref COLLATE NOCASE`, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []graphcache.SourceSubject
	for rows.Next() {
		var s graphcache.SourceSubject
		if err := rows.Scan(&s.ID, &s.Ref, &s.SourceID, &s.SubjectTypeID, &s.Label, &s.Description, &s.TypeKey, &s.TypeOrigin); err != nil {
			return nil, err
		}
		s.ID = append([]byte(nil), s.ID...)
		s.SourceID = append([]byte(nil), s.SourceID...)
		s.SubjectTypeID = append([]byte(nil), s.SubjectTypeID...)
		out = append(out, s)
	}
	return out, rows.Err()
}

func loadSourceObservations(q graphcache.Querier, sourceID []byte) ([]graphcache.SourceObservation, error) {
	rows, err := q.Query(`SELECT o.id, o.ref, o.citation_id, o.subject_id, o.property_id, o.polarity,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_subject_id, o.value_term_id, o.value_depiction_id,
			dpv.region_json, dpv.image_file_id, dc.locator_json, dpf.media_type
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		LEFT JOIN depiction_values dpv ON dpv.id = o.value_depiction_id
		LEFT JOIN citations dc ON dc.id = o.citation_id
		LEFT JOIN files dpf ON dpf.id = dpv.image_file_id
		WHERE s.source_id = ?
		ORDER BY o.ref COLLATE NOCASE`, sourceID)
	if err != nil {
		return nil, err
	}
	var out []graphcache.SourceObservation
	var dateIDs, nameIDs [][]byte
	for rows.Next() {
		var (
			o                               graphcache.SourceObservation
			text                            sql.NullString
			integer                         sql.NullInt64
			depictionRegion                 sql.NullString
			depictionImage                  []byte
			citationLocator, imageMediaType sql.NullString
		)
		if err := rows.Scan(&o.ID, &o.Ref, &o.CitationID, &o.SubjectID, &o.PropertyID, &o.Polarity,
			&text, &integer, &o.DateID, &o.NameID, &o.ValueSubjectID, &o.TermID, &o.DepictionID,
			&depictionRegion, &depictionImage, &citationLocator, &imageMediaType); err != nil {
			_ = rows.Close()
			return nil, err
		}
		o.Text, o.HasText = text.String, text.Valid
		o.Integer, o.HasInteger = integer.Int64, integer.Valid
		o.ID = append([]byte(nil), o.ID...)
		o.CitationID = append([]byte(nil), o.CitationID...)
		o.SubjectID = append([]byte(nil), o.SubjectID...)
		o.PropertyID = append([]byte(nil), o.PropertyID...)
		o.DateID = append([]byte(nil), o.DateID...)
		o.NameID = append([]byte(nil), o.NameID...)
		o.ValueSubjectID = append([]byte(nil), o.ValueSubjectID...)
		o.TermID = append([]byte(nil), o.TermID...)
		o.DepictionID = append([]byte(nil), o.DepictionID...)
		if dv := depictionvalues.FromJoin(o.DepictionID, depictionRegion, depictionImage, citationLocator, imageMediaType); dv != nil {
			o.DepictionRegion, o.DepictionImageFileID = dv.RegionJSON, dv.ImageFileID
		}
		if len(o.DateID) > 0 {
			dateIDs = append(dateIDs, o.DateID)
		}
		if len(o.NameID) > 0 {
			nameIDs = append(nameIDs, o.NameID)
		}
		out = append(out, o)
	}
	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}
	dates, err := datevalues.LookupManyTx(q, dateIDs)
	if err != nil {
		return nil, err
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if d, ok := dates[string(out[i].DateID)]; ok {
			b, err := valuecodec.MarshalDate(d)
			if err != nil {
				return nil, err
			}
			out[i].Date = b
		}
		if n, ok := names[string(out[i].NameID)]; ok {
			b, err := valuecodec.MarshalName(n)
			if err != nil {
				return nil, err
			}
			out[i].Name = b
		}
	}
	return out, nil
}

func loadSourceMembers(q graphcache.Querier, sourceID []byte) ([]graphcache.SourceMember, error) {
	rows, err := q.Query(`SELECT s.id, ic.id, e.id, st.key
		FROM subjects s
		JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		JOIN canonical_entities e ON e.id = ic.entity_id
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE s.source_id = ?
		ORDER BY s.ref COLLATE NOCASE`, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []graphcache.SourceMember
	for rows.Next() {
		var m graphcache.SourceMember
		if err := rows.Scan(&m.SubjectID, &m.ClaimID, &m.EntityID, &m.Kind); err != nil {
			return nil, err
		}
		m.SubjectID = append([]byte(nil), m.SubjectID...)
		m.ClaimID = append([]byte(nil), m.ClaimID...)
		m.EntityID = append([]byte(nil), m.EntityID...)
		out = append(out, m)
	}
	return out, rows.Err()
}

func loadTitleEdges(g *graphcache.Graph, sg graphcache.SourceGraph, subjects []graphcache.SourceSubject) ([]graphcache.SourceEdge, []graphcache.SourceEdge, error) {
	var events [][]byte
	for _, s := range subjects {
		if s.TypeKey == vocabkey.Event && s.TypeOrigin == vocabkey.OriginProvenencia {
			events = append(events, s.ID)
		}
	}
	if len(events) == 0 {
		return nil, nil, nil
	}
	people, err := g.WalkSource(sg, hops.SubjectsOfEvent, events)
	if err != nil {
		return nil, nil, err
	}
	places, err := g.WalkSource(sg, hops.PlacesOfEvent, events)
	if err != nil {
		return nil, nil, err
	}
	return people, places, nil
}

func checkSource(g *graphcache.Graph, id []byte, stored graphcache.SourceGraph) error {
	fresh, err := loadSource(g, id)
	if err != nil {
		return err
	}
	fresh.SourceID = append([]byte(nil), id...)
	if !reflect.DeepEqual(stored, fresh) {
		return fmt.Errorf("sourcefill: stored source %x does not match a fresh read", id)
	}
	return nil
}
