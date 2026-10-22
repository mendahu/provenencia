// Package graphfill loads one Evidence graph onto the catalog graph: subjects,
// observations, members, and the edges an event title reads. Promote builds
// its own bridge signatures when it reads the layer.
package graphfill

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
	graphcache.SetEvidenceGraphLoader(loadEvidenceGraph)
	graphcache.SetEvidenceGraphCheck(checkEvidenceGraph)
}

func loadEvidenceGraph(g *graphcache.Graph, graphID []byte) (graphcache.EvidenceGraph, error) {
	q := graphcache.Querier(g)
	sourceID, title, found, err := loadGraphSource(q, graphID)
	if err != nil {
		return graphcache.EvidenceGraph{}, err
	}
	if !found {
		return graphcache.EvidenceGraph{Found: false}, nil
	}
	subjects, err := loadGraphSubjects(q, graphID)
	if err != nil {
		return graphcache.EvidenceGraph{}, err
	}
	observations, err := loadGraphObservations(q, graphID)
	if err != nil {
		return graphcache.EvidenceGraph{}, err
	}
	members, err := loadGraphMembers(q, graphID)
	if err != nil {
		return graphcache.EvidenceGraph{}, err
	}
	people, places, err := loadTitleEdges(g, graphcache.EvidenceGraph{GraphID: graphID, SourceID: sourceID, Subjects: subjects, Observations: observations})
	if err != nil {
		return graphcache.EvidenceGraph{}, err
	}
	return graphcache.EvidenceGraph{
		SourceID:     sourceID,
		SourceTitle:  title,
		Found:        true,
		Subjects:     subjects,
		Observations: observations,
		Members:      members,
		EventPeople:  people,
		EventPlaces:  places,
	}, nil
}

// loadGraphSource reads the Source of an Evidence graph and its title, and
// whether the graph exists.
func loadGraphSource(q graphcache.Querier, graphID []byte) (sourceID []byte, title string, found bool, err error) {
	err = q.QueryRow(`SELECT g.source_id, s.title FROM evidence_graphs g JOIN sources s ON s.id = g.source_id WHERE g.id = ?`, graphID).Scan(&sourceID, &title)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", false, nil
	}
	if err != nil {
		return nil, "", false, err
	}
	return append([]byte(nil), sourceID...), title, true, nil
}

func loadGraphSubjects(q graphcache.Querier, graphID []byte) ([]graphcache.EvidenceSubject, error) {
	rows, err := q.Query(`SELECT s.id, s.ref, s.source_id, s.subject_type_id,
			COALESCE(s.label, ''), COALESCE(s.description, ''), st.key, st.origin
		FROM subjects s
		JOIN subject_types st ON st.id = s.subject_type_id
		WHERE s.graph_id = ?
		ORDER BY s.label COLLATE NOCASE, s.ref COLLATE NOCASE`, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []graphcache.EvidenceSubject
	for rows.Next() {
		var s graphcache.EvidenceSubject
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

func loadGraphObservations(q graphcache.Querier, graphID []byte) ([]graphcache.EvidenceObservation, error) {
	rows, err := q.Query(`SELECT o.id, o.ref, o.citation_id, o.subject_id, o.property_id, o.polarity,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_subject_id, o.value_term_id, o.value_depiction_id,
			dpv.region_json, dpv.image_file_id, dc.locator_json, dpf.media_type
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		LEFT JOIN depiction_values dpv ON dpv.id = o.value_depiction_id
		LEFT JOIN citations dc ON dc.id = o.citation_id
		LEFT JOIN files dpf ON dpf.id = dpv.image_file_id
		WHERE s.graph_id = ?
		ORDER BY o.ref COLLATE NOCASE`, graphID)
	if err != nil {
		return nil, err
	}
	var out []graphcache.EvidenceObservation
	var dateIDs, nameIDs [][]byte
	for rows.Next() {
		var (
			o                               graphcache.EvidenceObservation
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

func loadGraphMembers(q graphcache.Querier, graphID []byte) ([]graphcache.EvidenceMember, error) {
	rows, err := q.Query(`SELECT s.id, ic.id, e.id, st.key
		FROM subjects s
		JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
		JOIN canonical_entities e ON e.id = ic.entity_id
		JOIN subject_types st ON st.id = e.subject_type_id
		WHERE s.graph_id = ?
		ORDER BY s.ref COLLATE NOCASE`, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []graphcache.EvidenceMember
	for rows.Next() {
		var m graphcache.EvidenceMember
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

func loadTitleEdges(g *graphcache.Graph, eg graphcache.EvidenceGraph) ([]graphcache.EvidenceEdge, []graphcache.EvidenceEdge, error) {
	var events [][]byte
	for _, s := range eg.Subjects {
		if s.TypeKey == vocabkey.Event && s.TypeOrigin == vocabkey.OriginProvenencia {
			events = append(events, s.ID)
		}
	}
	if len(events) == 0 {
		return nil, nil, nil
	}
	people, err := g.WalkEvidenceGraph(eg, hops.SubjectsOfEvent, events)
	if err != nil {
		return nil, nil, err
	}
	places, err := g.WalkEvidenceGraph(eg, hops.PlacesOfEvent, events)
	if err != nil {
		return nil, nil, err
	}
	return people, places, nil
}

func checkEvidenceGraph(g *graphcache.Graph, graphID []byte, stored graphcache.EvidenceGraph) error {
	fresh, err := loadEvidenceGraph(g, graphID)
	if err != nil {
		return err
	}
	fresh.GraphID = append([]byte(nil), graphID...)
	if !reflect.DeepEqual(stored, fresh) {
		return fmt.Errorf("graphfill: stored evidence graph %x does not match a fresh read", graphID)
	}
	return nil
}
