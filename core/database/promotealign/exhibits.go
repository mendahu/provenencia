package promotealign

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

type exhibitObs struct {
	id        []byte
	subjectID []byte
	prop      match.Property
	valueType string
	display   string
	source    string
	value     match.Value
}

// attachExhibits fills pin lines on handle rows: the Subject's own records
// and one hop through a bridge, paired with a member observation of the
// chosen handle (or one hop from that member) on the same Property.
func attachExhibits(q Querier, layer graphalign.Layer, prop *graphalign.Proposal, cfg graphalign.Config, stats graphalign.Stats) error {
	if prop == nil || len(prop.Rows) == 0 {
		return nil
	}
	neighbors := map[string][][]byte{}
	for _, b := range layer.Bridges {
		neighbors[string(b.A)] = append(neighbors[string(b.A)], b.B)
		neighbors[string(b.B)] = append(neighbors[string(b.B)], b.A)
	}
	memberHops, err := bridgeNeighbors(q)
	if err != nil {
		return err
	}

	var want [][]byte
	seen := map[string]bool{}
	add := func(id []byte) {
		if len(id) != 16 || seen[string(id)] {
			return
		}
		seen[string(id)] = true
		want = append(want, id)
	}

	type rowMembers struct {
		idx     int
		members [][]byte
	}
	var rows []rowMembers
	for i := range prop.Rows {
		r := &prop.Rows[i]
		if r.Target != graphalign.TargetHandle || len(r.HandleID) != 16 {
			continue
		}
		members, err := acceptedMembers(q, r.HandleID)
		if err != nil {
			return err
		}
		add(r.SubjectID)
		for _, n := range neighbors[string(r.SubjectID)] {
			add(n)
		}
		for _, m := range members {
			add(m)
			for _, n := range memberHops[string(m)] {
				add(n)
			}
		}
		rows = append(rows, rowMembers{idx: i, members: members})
	}
	if len(want) == 0 {
		return nil
	}
	obs, labels, err := loadExhibitObservations(q, want)
	if err != nil {
		return err
	}
	bySubject := map[string][]exhibitObs{}
	for _, o := range obs {
		bySubject[string(o.subjectID)] = append(bySubject[string(o.subjectID)], o)
	}

	for _, rm := range rows {
		r := &prop.Rows[rm.idx]
		var incoming []exhibitObs
		incoming = append(incoming, bySubject[string(r.SubjectID)]...)
		hopLabel := map[string]string{}
		for _, n := range neighbors[string(r.SubjectID)] {
			hopLabel[string(n)] = labels[string(n)]
			incoming = append(incoming, bySubject[string(n)]...)
		}
		var memberObs []exhibitObs
		for _, m := range rm.members {
			memberObs = append(memberObs, bySubject[string(m)]...)
			for _, n := range memberHops[string(m)] {
				memberObs = append(memberObs, bySubject[string(n)]...)
			}
		}
		r.Exhibits = pairExhibits(r.SubjectID, incoming, hopLabel, memberObs, cfg, stats)
	}
	return nil
}

func pairExhibits(subjectID []byte, incoming []exhibitObs, hopLabel map[string]string, memberObs []exhibitObs, cfg graphalign.Config, stats graphalign.Stats) []graphalign.Exhibit {
	var out []graphalign.Exhibit
	used := map[string]bool{}
	for _, in := range incoming {
		if in.valueType == "" || in.valueType == "subject" {
			continue
		}
		var agree, clash *exhibitObs
		for i := range memberObs {
			mem := &memberObs[i]
			if mem.prop != in.prop || mem.valueType != in.valueType {
				continue
			}
			if exhibitCompatible(in.valueType, in.value, mem.value) {
				agree = mem
				break
			}
			if clash == nil && carries(in.valueType, in.value) && carries(mem.valueType, mem.value) {
				clash = mem
			}
		}
		var mem *exhibitObs
		outcome := match.OutcomeUnknown
		pinned := false
		switch {
		case agree != nil:
			mem = agree
			outcome = match.OutcomeAgree
			pinned = true
		case clash != nil:
			mem = clash
			outcome = match.OutcomeConflict
		default:
			continue
		}
		key := string(in.id) + "|" + string(mem.id)
		if used[key] {
			continue
		}
		used[key] = true
		group := ""
		if string(in.subjectID) != string(subjectID) {
			label := hopLabel[string(in.subjectID)]
			if label == "" {
				label = "Neighbor"
			}
			group = label + " · " + strings.ReplaceAll(in.prop.Key, "_", " ")
		}
		probe := match.Values{in.prop: {in.value}}
		out = append(out, graphalign.Exhibit{
			Property:              in.prop,
			Outcome:               outcome,
			ValueType:             in.valueType,
			Pinned:                pinned,
			Weight:                graphalign.PropertyWeight(outcome, in.valueType, in.prop, probe, cfg, stats),
			GroupLabel:            group,
			IncomingObservationID: append([]byte(nil), in.id...),
			IncomingDisplay:       in.display,
			IncomingSource:        in.source,
			MemberObservationID:   append([]byte(nil), mem.id...),
			MemberDisplay:         mem.display,
			MemberSource:          mem.source,
		})
	}
	return out
}

func exhibitCompatible(valueType string, a, b match.Value) bool {
	return autoreconcile.Compatible(valueType, exhibitAuto(a), exhibitAuto(b))
}

func carries(valueType string, v match.Value) bool {
	av := exhibitAuto(v)
	return autoreconcile.Compatible(valueType, av, av)
}

func exhibitAuto(v match.Value) autoreconcile.Value {
	termID := append([]byte(nil), v.TermID...)
	if len(termID) == 0 && v.Term != "" {
		termID = []byte(v.Term)
	}
	return autoreconcile.Value{
		Text: v.Text, HasText: v.HasText, Integer: v.Integer, HasInteger: v.HasInteger,
		TermKey: v.Term, TermID: termID, Date: v.Date, Name: v.Name,
	}
}

func acceptedMembers(q Querier, entityID []byte) ([][]byte, error) {
	rows, err := q.Query(`SELECT subject_id FROM identity_claims
		WHERE entity_id = ? AND status = 'accepted' ORDER BY subject_id`, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, append([]byte(nil), id...))
	}
	return out, rows.Err()
}

func bridgeNeighbors(q Querier) (map[string][][]byte, error) {
	rows, err := q.Query(`SELECT e1.value_subject_id, e2.value_subject_id
		FROM observations e1
		JOIN observations e2 ON e2.subject_id = e1.subject_id AND e2.id != e1.id
			AND e2.polarity = 'positive' AND e2.value_subject_id IS NOT NULL
		JOIN subjects bs ON bs.id = e1.subject_id
		JOIN subject_types st ON st.id = bs.subject_type_id
			AND st.origin = 'provenencia'
			AND st.key IN ('participation', 'relationship', 'location', 'place_relationship')
		WHERE e1.polarity = 'positive' AND e1.value_subject_id IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][][]byte{}
	seen := map[string]bool{}
	for rows.Next() {
		var a, b []byte
		if err := rows.Scan(&a, &b); err != nil {
			return nil, err
		}
		if len(a) != 16 || len(b) != 16 || string(a) == string(b) {
			continue
		}
		key := string(a) + "|" + string(b)
		if seen[key] {
			continue
		}
		seen[key] = true
		out[string(a)] = append(out[string(a)], append([]byte(nil), b...))
	}
	return out, rows.Err()
}

func loadExhibitObservations(q Querier, subjectIDs [][]byte) ([]exhibitObs, map[string]string, error) {
	placeholders := make([]string, len(subjectIDs))
	args := make([]any, len(subjectIDs))
	for i, id := range subjectIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `SELECT o.id, o.subject_id, s.ref, p.key, p.origin, p.value_type,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, t.key, src.title
		FROM observations o
		JOIN subjects s ON s.id = o.subject_id
		JOIN properties p ON p.id = o.property_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		WHERE o.polarity = 'positive' AND o.subject_id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	type pending struct {
		o              exhibitObs
		dateID, nameID []byte
		term           string
	}
	var all []pending
	var dateIDs, nameIDs [][]byte
	labels := map[string]string{}
	for rows.Next() {
		var (
			p              pending
			ref, valueType string
			text, term     sql.NullString
			integer        sql.NullInt64
		)
		if err := rows.Scan(&p.o.id, &p.o.subjectID, &ref, &p.o.prop.Key, &p.o.prop.Origin, &valueType,
			&text, &integer, &p.dateID, &p.nameID, &term, &p.o.source); err != nil {
			_ = rows.Close()
			return nil, nil, err
		}
		p.o.valueType = valueType
		p.o.value = match.Value{
			Text: text.String, HasText: text.Valid,
			Integer: integer.Int64, HasInteger: integer.Valid,
			Term: term.String,
		}
		p.term = term.String
		if _, ok := labels[string(p.o.subjectID)]; !ok {
			labels[string(p.o.subjectID)] = ref
		}
		all = append(all, p)
		if len(p.dateID) > 0 {
			dateIDs = append(dateIDs, p.dateID)
		}
		if len(p.nameID) > 0 {
			nameIDs = append(nameIDs, p.nameID)
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, nil, err
	}
	_ = rows.Close()

	dates, err := datevalues.LookupManyTx(q, dateIDs)
	if err != nil {
		return nil, nil, err
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return nil, nil, err
	}
	out := make([]exhibitObs, 0, len(all))
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.o.value.Date = &d
			if d.StartYear != nil {
				p.o.display = strconv.Itoa(*d.StartYear)
			}
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.o.value.Name = &n
			if n.Form != "" {
				p.o.display = n.Form
				labels[string(p.o.subjectID)] = n.Form
			}
		}
		if p.o.display == "" && p.o.value.HasText && p.o.value.Text != "" {
			p.o.display = p.o.value.Text
			if p.o.prop.Key == "toponym" {
				labels[string(p.o.subjectID)] = p.o.value.Text
			}
		}
		if p.o.display == "" && p.term != "" {
			p.o.display = p.term
			if p.o.prop.Key == "event_type" {
				labels[string(p.o.subjectID)] = p.term
			}
		}
		if p.o.display == "" && p.o.value.HasInteger {
			p.o.display = strconv.FormatInt(p.o.value.Integer, 10)
		}
		out = append(out, p.o)
	}
	return out, labels, nil
}
