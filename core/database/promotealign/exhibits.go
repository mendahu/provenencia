package promotealign

import (
	"database/sql"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/propertymeta"
)

type exhibitObs struct {
	id          []byte
	subjectID   []byte
	prop        match.Property
	valueType   string
	cardinality string
	source      string
	value       match.Value
}

// attachExhibits fills pin lines on handle rows. The Subject's own records
// pair with records on the chosen handle's members. A neighbor's records pair
// only with records on the handle Align mapped that neighbor to, and only on
// members one hop from the chosen handle's members: "her birth date matches"
// compares this birth with that birth, never with some other event's date.
// A neighbor Align didn't map (New, Skip, unreachable) contributes nothing.
func attachExhibits(q Querier, g *graphcache.Graph, layer graphalign.Layer, prop *graphalign.Proposal, cfg graphalign.Config, stats graphalign.Stats) error {
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
	members, err := acceptedMembers(g, handleIDs)
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
	var fromMembers, towardMembers [][]byte
	for i := range prop.Rows {
		r := &prop.Rows[i]
		if r.Target != graphalign.TargetHandle || len(r.HandleID) != 16 {
			continue
		}
		own := members[string(r.HandleID)]
		plans = append(plans, rowPlan{idx: i, own: own})
		mapped := false
		for _, n := range neighbors[string(r.SubjectID)] {
			gID, ok := mappedTo[string(n)]
			if !ok {
				continue
			}
			mapped = true
			towardMembers = append(towardMembers, members[string(gID)]...)
		}
		if mapped {
			fromMembers = append(fromMembers, own...)
		}
	}
	hops, err := bridgeNeighbors(q, fromMembers, towardMembers)
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
	if len(want) > 0 {
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
	}
	for i := range prop.Rows {
		r := &prop.Rows[i]
		r.Exhibits = withRowLines(r.Exhibits, r.Comparisons)
	}
	return nil
}

// withRowLines completes a row's sheet lines from its comparisons. A row with
// no exhibit gets a line per comparison that spoke, so the sheet still shows
// the two values behind each outcome. A row with exhibits gets a line for each
// Property the row judges a conflict that the own-record exhibits do not
// already show as one: the majority disagreeing with a minority value, or a
// negative Observation, which the member picker never pairs. These lines
// carry no Observation, so they cannot be pinned.
func withRowLines(exhibits []graphalign.Exhibit, comparisons []graphalign.Comparison) []graphalign.Exhibit {
	none := len(exhibits) == 0
	for _, c := range comparisons {
		if none {
			if c.Outcome == match.OutcomeUnknown {
				continue
			}
		} else {
			if c.Outcome != match.OutcomeConflict {
				continue
			}
			shown := false
			for _, ex := range exhibits {
				shown = shown || len(ex.GroupSubjectID) == 0 && ex.Property == c.Property && ex.Outcome == match.OutcomeConflict
			}
			if shown {
				continue
			}
		}
		exhibits = append(exhibits, refLine(c))
	}
	return exhibits
}

// refLine is the sheet line for one row comparison: its two values and dates,
// with no Observation behind either.
func refLine(c graphalign.Comparison) graphalign.Exhibit {
	return graphalign.Exhibit{
		Property:         c.Property,
		Outcome:          c.Outcome,
		ValueType:        c.ValueType,
		Weight:           c.Weight,
		IncomingDisplay:  c.Probe.Label,
		IncomingDate:     cloneDate(c.Probe.Date),
		IncomingNegative: c.Probe.Negative,
		MemberDisplay:    c.Candidate.Label,
		MemberDate:       cloneDate(c.Candidate.Date),
	}
}

// pairExhibits pairs each incoming record with the closest record on the
// same Property. A same-value pair (Compatible) is preferred and pinned.
// The line's outcome and weight come from the promote scale. neighbor is
// nil for the Subject's own records and names the layer neighbor for a
// one-hop group.
func pairExhibits(incoming []exhibitObs, neighbor []byte, memberObs []exhibitObs, cfg graphalign.Config, stats graphalign.Stats) []graphalign.Exhibit {
	var out []graphalign.Exhibit
	for _, in := range incoming {
		if in.valueType == "" || in.valueType == "subject" {
			continue
		}
		cmp := match.ComparerFor(in.valueType)
		if in.valueType == propertymeta.ValueTypeName && cfg.Names != nil {
			cmp = *cfg.Names
		}
		var chosen *exhibitObs
		var sim, discount float64
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
				sim, discount, comparable = comparerDiscount(cmp, in.value, mem.value)
				break
			}
			s, d, ok := comparerDiscount(cmp, in.value, mem.value)
			if !ok {
				continue
			}
			if chosen == nil || s*(1-d) > sim*(1-discount) {
				chosen = mem
				sim, discount = s, d
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
			card = propertymeta.CardinalitySingle
		}
		outcome, weight := graphalign.ScoreSimilarity(sim, discount, comparable, card, in.prop, in.valueType, []match.Value{in.value}, cfg, stats)
		if outcome == match.OutcomeUnknown {
			continue
		}
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
			IncomingDisplay:       in.value.Label(),
			IncomingSource:        in.source,
			IncomingDate:          cloneDate(in.value.Date),
			MemberObservationID:   append([]byte(nil), chosen.id...),
			MemberDisplay:         chosen.value.Label(),
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

// comparerDiscount is the comparer's similarity and the share of points the
// matched values' frequency costs.
func comparerDiscount(cmp match.Comparer, a, b match.Value) (similarity, discount float64, comparable bool) {
	if cmp == nil {
		return 0, 0, false
	}
	return match.CompareDiscount(cmp, a, b)
}

func exhibitCompatible(valueType string, a, b match.Value) bool {
	return autoreconcile.Compatible(valueType, a.Auto(), b.Auto())
}

// acceptedMembers maps each handle to its accepted member Subjects, in claim
// id order, from nodes the proposal already loaded.
func acceptedMembers(g *graphcache.Graph, entityIDs [][]byte) (map[string][][]byte, error) {
	out := map[string][][]byte{}
	if g == nil {
		return out, nil
	}
	for _, id := range database.UniqueBlobIDs(entityIDs) {
		n, err := g.Node(id)
		if err != nil {
			return nil, err
		}
		if n == nil {
			continue
		}
		for _, m := range n.Members {
			if m.Accepted {
				out[string(n.ID)] = append(out[string(n.ID)], m.SubjectID)
			}
		}
	}
	return out, nil
}

// bridgeNeighbors maps each subject in from to the subjects in toward that
// sit one bridge away. toward limits the read to the neighbor handle's
// members, so a hub's members do not pull every bridge.
func bridgeNeighbors(q Querier, from, toward [][]byte) (map[string][][]byte, error) {
	out := map[string][][]byte{}
	keys := connectrules.BridgeTypeKeys()
	from = database.UniqueBlobIDs(from)
	toward = database.UniqueBlobIDs(toward)
	if len(from) == 0 || len(toward) == 0 || len(keys) == 0 {
		return out, nil
	}
	seen := map[string]bool{}
	// Both lists can be long, so each pair of batches is one query.
	err := sqlbatch.Each(from, func(fromBatch [][]byte) error {
		args := make([]any, 0, len(keys)+len(fromBatch))
		for _, k := range keys {
			args = append(args, k)
		}
		args = append(args, database.BlobArgs(fromBatch)...)
		head := `SELECT e1.value_subject_id, e2.value_subject_id
			FROM observations e1
			JOIN observations e2 ON e2.subject_id = e1.subject_id AND e2.id != e1.id
				AND e2.polarity = 'positive' AND e2.value_subject_id IS NOT NULL
			JOIN subjects bs ON bs.id = e1.subject_id
			JOIN subject_types st ON st.id = bs.subject_type_id
				AND st.origin = 'provenencia'
				AND st.key IN (` + database.SQLInPlaceholders(len(keys)) + `)
			WHERE e1.polarity = 'positive'
				AND e1.value_subject_id IN (` + database.SQLInPlaceholders(len(fromBatch)) + `)
				AND e2.value_subject_id IN (`
		return sqlbatch.Query(q, head, `)`, args, toward, func(rows *sql.Rows) error {
			var a, b []byte
			if err := rows.Scan(&a, &b); err != nil {
				return err
			}
			if len(a) != 16 || len(b) != 16 || string(a) == string(b) {
				return nil
			}
			key := string(a) + "|" + string(b)
			if seen[key] {
				return nil
			}
			seen[key] = true
			out[string(a)] = append(out[string(a)], append([]byte(nil), b...))
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func loadExhibitObservations(q Querier, subjectIDs [][]byte) ([]exhibitObs, error) {
	query := `SELECT o.id, o.subject_id, p.key, p.origin, p.value_type, p.cardinality,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, t.key, t.label, COALESCE(t.neutral, 0), src.title
		FROM observations o
		JOIN properties p ON p.id = o.property_id
		JOIN citations c ON c.id = o.citation_id
		JOIN artifacts a ON a.id = c.artifact_id
		JOIN sources src ON src.id = a.source_id
		LEFT JOIN property_terms t ON t.id = o.value_term_id
		WHERE o.polarity = 'positive' AND o.subject_id IN (`
	type pending struct {
		o              exhibitObs
		dateID, nameID []byte
	}
	var all []pending
	var dateIDs, nameIDs [][]byte
	err := sqlbatch.Query(q, query, `)`, nil, subjectIDs, func(rows *sql.Rows) error {
		var (
			p                     pending
			valueType             string
			text, term, termLabel sql.NullString
			integer               sql.NullInt64
		)
		if err := rows.Scan(&p.o.id, &p.o.subjectID, &p.o.prop.Key, &p.o.prop.Origin, &valueType, &p.o.cardinality,
			&text, &integer, &p.dateID, &p.nameID, &term, &termLabel, &p.o.value.Neutral, &p.o.source); err != nil {
			return err
		}
		p.o.valueType = valueType
		p.o.value = match.Value{
			Text: text.String, HasText: text.Valid,
			Integer: integer.Int64, HasInteger: integer.Valid,
			Term: term.String, Neutral: p.o.value.Neutral,
			TermLabel: termLabel.String,
		}
		all = append(all, p)
		if len(p.dateID) > 0 {
			dateIDs = append(dateIDs, p.dateID)
		}
		if len(p.nameID) > 0 {
			nameIDs = append(nameIDs, p.nameID)
		}
		return nil
	})
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
	loaded, err := nameformats.Load(q)
	if err != nil {
		return nil, err
	}
	out := make([]exhibitObs, 0, len(all))
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.o.value.Date = &d
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.o.value.Name = &n
			p.o.value.NamePattern = loaded.Key
		}
		out = append(out, p.o)
	}
	return out, nil
}
