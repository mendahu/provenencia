package promotealign

import (
	"database/sql"
	"strconv"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
)

type exhibitObs struct {
	id          []byte
	subjectID   []byte
	prop        match.Property
	valueType   string
	cardinality string
	display     string
	source      string
	value       match.Value
}

// attachExhibits fills pin lines on handle rows. The Subject's own records
// pair with records on the chosen handle's members. A neighbor's records pair
// only with records on the handle Align mapped that neighbor to, and only on
// members one hop from the chosen handle's members: "her birth date matches"
// compares this birth with that birth, never with some other event's date.
// A neighbor Align didn't map (New, Skip, unreachable) contributes nothing.
func attachExhibits(q Querier, layer graphalign.Layer, prop *graphalign.Proposal, cfg graphalign.Config, stats graphalign.Stats) error {
	if prop == nil || len(prop.Rows) == 0 {
		return nil
	}
	neighbors := map[string][][]byte{}
	for _, b := range layer.Bridges {
		neighbors[string(b.A)] = append(neighbors[string(b.A)], b.B)
		neighbors[string(b.B)] = append(neighbors[string(b.B)], b.A)
	}
	mappedTo := map[string][]byte{}
	for _, r := range prop.Rows {
		if r.Target == graphalign.TargetHandle && len(r.HandleID) == 16 {
			mappedTo[string(r.SubjectID)] = r.HandleID
		}
	}

	var handleIDs [][]byte
	for _, h := range mappedTo {
		handleIDs = append(handleIDs, h)
	}
	members, err := acceptedMembers(q, handleIDs) // every row's handle, neighbors' included
	if err != nil {
		return err
	}

	type neighborPair struct {
		neighbor []byte   // layer neighbor of the row's Subject
		sides    [][]byte // its handle's members one hop from the row's handle
	}
	type rowPlan struct {
		idx       int
		own       [][]byte // the row handle's members
		neighbors []neighborPair
	}
	var plans []rowPlan
	var memberIDs [][]byte
	for i := range prop.Rows {
		r := &prop.Rows[i]
		if r.Target != graphalign.TargetHandle || len(r.HandleID) != 16 {
			continue
		}
		own := members[string(r.HandleID)]
		memberIDs = append(memberIDs, own...)
		plans = append(plans, rowPlan{idx: i, own: own})
	}
	hops, err := bridgeNeighbors(q, memberIDs)
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
	for pi := range plans {
		plan := &plans[pi]
		r := &prop.Rows[plan.idx]
		add(r.SubjectID)
		for _, m := range plan.own {
			add(m)
		}
		near := map[string]bool{}
		for _, m := range plan.own {
			for _, n := range hops[string(m)] {
				near[string(n)] = true
			}
		}
		for _, n := range neighbors[string(r.SubjectID)] {
			g, ok := mappedTo[string(n)]
			if !ok {
				continue
			}
			var sides [][]byte
			for _, gm := range members[string(g)] {
				if near[string(gm)] {
					sides = append(sides, gm)
				}
			}
			if len(sides) == 0 {
				continue
			}
			add(n)
			for _, side := range sides {
				add(side)
			}
			plan.neighbors = append(plan.neighbors, neighborPair{neighbor: n, sides: sides})
		}
	}
	if len(want) == 0 {
		return nil
	}
	obs, err := loadExhibitObservations(q, want)
	if err != nil {
		return err
	}
	bySubject := map[string][]exhibitObs{}
	for _, o := range obs {
		bySubject[string(o.subjectID)] = append(bySubject[string(o.subjectID)], o)
	}
	collect := func(ids [][]byte) []exhibitObs {
		var out []exhibitObs
		for _, id := range ids {
			out = append(out, bySubject[string(id)]...)
		}
		return out
	}

	for _, plan := range plans {
		r := &prop.Rows[plan.idx]
		exhibits := pairExhibits(bySubject[string(r.SubjectID)], nil, collect(plan.own), cfg, stats)
		for _, np := range plan.neighbors {
			exhibits = append(exhibits, pairExhibits(bySubject[string(np.neighbor)], np.neighbor, collect(np.sides), cfg, stats)...)
		}
		r.Exhibits = exhibits
	}
	return nil
}

// pairExhibits pairs each incoming record with the closest record on the
// same Property. A same-value pair (Compatible) is preferred and pinned.
// The line's outcome and weight come from the promote scale. neighbor is
// nil for the Subject's own records and names the layer neighbor for a
// one-hop group.
func pairExhibits(incoming []exhibitObs, neighbor []byte, memberObs []exhibitObs, cfg graphalign.Config, stats graphalign.Stats) []graphalign.Exhibit {
	var out []graphalign.Exhibit
	used := map[string]bool{}
	for _, in := range incoming {
		if in.valueType == "" || in.valueType == "subject" {
			continue
		}
		cmp := match.ComparerFor(in.valueType)
		var chosen *exhibitObs
		var sim float64
		var comparable bool
		pinned := false
		for i := range memberObs {
			mem := &memberObs[i]
			if mem.prop != in.prop || mem.valueType != in.valueType {
				continue
			}
			if exhibitCompatible(in.valueType, in.value, mem.value) {
				chosen = mem
				pinned = true
				sim, comparable = comparerSimilarity(cmp, in.value, mem.value)
				break
			}
			s, ok := comparerSimilarity(cmp, in.value, mem.value)
			if !ok {
				continue
			}
			if chosen == nil || s > sim {
				chosen = mem
				sim = s
				comparable = true
			}
		}
		if chosen == nil {
			continue
		}
		if pinned && !comparable {
			sim, comparable = 1, true
		}
		card := in.cardinality
		if card == "" {
			card = properties.CardinalitySingle
		}
		outcome, weight := graphalign.ScoreSimilarity(sim, comparable, card, in.prop, in.valueType, []match.Value{in.value}, cfg, stats)
		if outcome == match.OutcomeUnknown {
			continue
		}
		key := string(in.id) + "|" + string(chosen.id)
		if used[key] {
			continue
		}
		used[key] = true
		out = append(out, graphalign.Exhibit{
			Property:  in.prop,
			Outcome:   outcome,
			ValueType: in.valueType,
			// Compatible prefers the pair. A conflict is shown and left
			// unpinned; the sheet says a disagreement is not evidence.
			Pinned:                pinned && outcome == match.OutcomeAgree,
			Weight:                weight,
			GroupSubjectID:        append([]byte(nil), neighbor...),
			IncomingObservationID: append([]byte(nil), in.id...),
			IncomingDisplay:       in.display,
			IncomingSource:        in.source,
			IncomingDate:          cloneDate(in.value.Date),
			MemberObservationID:   append([]byte(nil), chosen.id...),
			MemberDisplay:         chosen.display,
			MemberSource:          chosen.source,
			MemberDate:            cloneDate(chosen.value.Date),
		})
	}
	return out
}

func cloneDate(d *datevalues.Value) *datevalues.Value {
	if d == nil {
		return nil
	}
	c := *d
	return &c
}

func comparerSimilarity(cmp match.Comparer, a, b match.Value) (float64, bool) {
	if cmp == nil {
		return 0, false
	}
	return cmp.Compare(a, b)
}

func exhibitCompatible(valueType string, a, b match.Value) bool {
	return autoreconcile.Compatible(valueType, exhibitAuto(a), exhibitAuto(b))
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

// acceptedMembers maps each handle to its accepted member Subjects.
func acceptedMembers(q Querier, entityIDs [][]byte) (map[string][][]byte, error) {
	out := map[string][][]byte{}
	ids := database.UniqueBlobIDs(entityIDs)
	for start := 0; start < len(ids); start += inBatch {
		chunk := ids[start:min(start+inBatch, len(ids))]
		rows, err := q.Query(`SELECT entity_id, subject_id FROM identity_claims
			WHERE status = 'accepted' AND entity_id IN (`+database.SQLInPlaceholders(len(chunk))+`)
			ORDER BY entity_id, subject_id`, database.BlobArgs(chunk)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var entity, subject []byte
			if err := rows.Scan(&entity, &subject); err != nil {
				_ = rows.Close()
				return nil, err
			}
			out[string(entity)] = append(out[string(entity)], subject)
		}
		err = rows.Err()
		_ = rows.Close()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// bridgeNeighbors maps each of subjectIDs to the Subjects one bridge away.
func bridgeNeighbors(q Querier, subjectIDs [][]byte) (map[string][][]byte, error) {
	out := map[string][][]byte{}
	keys := connectrules.BridgeTypeKeys()
	if len(subjectIDs) == 0 || len(keys) == 0 {
		return out, nil
	}
	args := make([]any, 0, len(keys)+len(subjectIDs))
	for _, k := range keys {
		args = append(args, k)
	}
	for _, id := range subjectIDs {
		args = append(args, id)
	}
	rows, err := q.Query(`SELECT e1.value_subject_id, e2.value_subject_id
		FROM observations e1
		JOIN observations e2 ON e2.subject_id = e1.subject_id AND e2.id != e1.id
			AND e2.polarity = 'positive' AND e2.value_subject_id IS NOT NULL
		JOIN subjects bs ON bs.id = e1.subject_id
		JOIN subject_types st ON st.id = bs.subject_type_id
			AND st.origin = 'provenencia'
			AND st.key IN (`+database.SQLInPlaceholders(len(keys))+`)
		WHERE e1.polarity = 'positive'
			AND e1.value_subject_id IN (`+database.SQLInPlaceholders(len(subjectIDs))+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func loadExhibitObservations(q Querier, subjectIDs [][]byte) ([]exhibitObs, error) {
	query := `SELECT o.id, o.subject_id, p.key, p.origin, p.value_type, p.cardinality,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, t.key, t.label, src.title
		FROM observations o
		JOIN properties p ON p.id = o.property_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		WHERE o.polarity = 'positive' AND o.subject_id IN (` + database.SQLInPlaceholders(len(subjectIDs)) + `)`
	args := make([]any, len(subjectIDs))
	for i, id := range subjectIDs {
		args[i] = id
	}
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	type pending struct {
		o              exhibitObs
		dateID, nameID []byte
		termLabel      string
	}
	var all []pending
	var dateIDs, nameIDs [][]byte
	for rows.Next() {
		var (
			p                     pending
			valueType             string
			text, term, termLabel sql.NullString
			integer               sql.NullInt64
		)
		if err := rows.Scan(&p.o.id, &p.o.subjectID, &p.o.prop.Key, &p.o.prop.Origin, &valueType, &p.o.cardinality,
			&text, &integer, &p.dateID, &p.nameID, &term, &termLabel, &p.o.source); err != nil {
			_ = rows.Close()
			return nil, err
		}
		p.o.valueType = valueType
		p.o.value = match.Value{
			Text: text.String, HasText: text.Valid,
			Integer: integer.Int64, HasInteger: integer.Valid,
			Term: term.String,
		}
		p.termLabel = termLabel.String
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
		return nil, err
	}
	_ = rows.Close()

	dates, err := datevalues.LookupManyTx(q, dateIDs)
	if err != nil {
		return nil, err
	}
	names, err := namevalues.LookupManyTx(q, nameIDs)
	if err != nil {
		return nil, err
	}
	out := make([]exhibitObs, 0, len(all))
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.o.value.Date = &d
			p.o.display = datevalues.CompactDisplay(d)
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.o.value.Name = &n
			p.o.display = n.Form
		}
		if p.o.display == "" && p.o.value.HasText {
			p.o.display = p.o.value.Text
		}
		if p.o.display == "" && p.termLabel != "" {
			p.o.display = p.termLabel
		}
		if p.o.display == "" && p.o.value.HasInteger {
			p.o.display = strconv.FormatInt(p.o.value.Integer, 10)
		}
		out = append(out, p.o)
	}
	return out, nil
}
