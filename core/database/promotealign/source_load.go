package promotealign

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/valuecodec"
)

func init() {
	graphcache.SetSourceLoader(loadSource)
	graphcache.SetSourceCheck(checkSource)
}

func loadSource(q graphcache.Querier, sourceID []byte) (graphcache.SourceGraph, error) {
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
	bridges, err := loadSourceBridges(q, sourceID, subjects)
	if err != nil {
		return graphcache.SourceGraph{}, err
	}
	people, places, err := loadTitleEdges(q, sourceID, subjects)
	if err != nil {
		return graphcache.SourceGraph{}, err
	}
	return graphcache.SourceGraph{
		Subjects:     subjects,
		Observations: observations,
		Members:      members,
		Bridges:      bridges,
		EventPeople:  people,
		EventPlaces:  places,
	}, nil
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
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_subject_id, o.value_term_id
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		WHERE s.source_id = ?
		ORDER BY o.ref COLLATE NOCASE`, sourceID)
	if err != nil {
		return nil, err
	}
	var out []graphcache.SourceObservation
	var dateIDs, nameIDs [][]byte
	for rows.Next() {
		var (
			o       graphcache.SourceObservation
			text    sql.NullString
			integer sql.NullInt64
		)
		if err := rows.Scan(&o.ID, &o.Ref, &o.CitationID, &o.SubjectID, &o.PropertyID, &o.Polarity,
			&text, &integer, &o.DateID, &o.NameID, &o.ValueSubjectID, &o.TermID); err != nil {
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

func loadSourceBridges(q graphcache.Querier, sourceID []byte, subjects []graphcache.SourceSubject) ([]graphcache.SourceBridge, error) {
	primary := map[string]primarySubject{}
	byKind := map[string][][]byte{}
	for _, s := range subjects {
		if !promotePrimary(s.TypeKey, s.TypeOrigin) {
			continue
		}
		if _, ok := match.DefaultProfile(s.TypeKey); !ok {
			continue
		}
		primary[string(s.ID)] = primarySubject{id: s.ID, ref: s.Ref, kind: s.TypeKey}
		byKind[s.TypeKey] = append(byKind[s.TypeKey], s.ID)
	}
	bridges, err := loadLayerBridges(q, sourceID, primary, byKind)
	if err != nil {
		return nil, err
	}
	out := make([]graphcache.SourceBridge, 0, len(bridges))
	for _, b := range bridges {
		out = append(out, graphcache.SourceBridge{
			A: append([]byte(nil), b.A...), B: append([]byte(nil), b.B...),
			BridgeType: b.Signature.BridgeType, Term: b.Signature.RoleOrType,
			NeighborKind: b.Signature.NeighborKind, NeighborTypeTerm: b.Signature.NeighborTypeTerm,
			Directed: b.Signature.Directed,
		})
	}
	return out, nil
}

func loadTitleEdges(q graphcache.Querier, sourceID []byte, subjects []graphcache.SourceSubject) ([]graphcache.SourceEdge, []graphcache.SourceEdge, error) {
	var events [][]byte
	for _, s := range subjects {
		if s.TypeKey == "event" && s.TypeOrigin == "provenencia" {
			events = append(events, s.ID)
		}
	}
	if len(events) == 0 {
		return nil, nil, nil
	}
	people, err := canonicalgraph.WalkSource(q, sourceID, hops.SubjectsOfEvent, events)
	if err != nil {
		return nil, nil, err
	}
	places, err := canonicalgraph.WalkSource(q, sourceID, hops.PlacesOfEvent, events)
	if err != nil {
		return nil, nil, err
	}
	return sourceEdges(people), sourceEdges(places), nil
}

func sourceEdges(edges []canonicalgraph.Edge) []graphcache.SourceEdge {
	out := make([]graphcache.SourceEdge, len(edges))
	for i, e := range edges {
		out[i] = graphcache.SourceEdge{
			From: append([]byte(nil), e.From...), To: append([]byte(nil), e.To...), ToRef: e.ToRef,
		}
	}
	return out
}

func checkSource(g *graphcache.Graph, id []byte, stored graphcache.SourceGraph) error {
	fresh, err := loadSource(g, id)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(stored.Members, fresh.Members) || !reflect.DeepEqual(stored.Bridges, fresh.Bridges) {
		return fmt.Errorf("promotealign: stored source %x members or bridges do not match a fresh read", id)
	}
	return nil
}
