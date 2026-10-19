package promotealign

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/sqlbatch"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type exhibitObs struct {
	id, subjectID []byte
	key, origin   string // the Observation's property
	valueType     string
	cardinality   string
	source        string
	termLabel     string
	auto          autoreconcile.Value
	value         propertycompare.Value
}

// attachExhibits fills pin lines on handle rows. The Subject's own records
// pair with records on the chosen handle's members. A neighbor's records pair
// only with records on the handle Align mapped that neighbor to, and only on
// members one hop from the chosen handle's members: "her birth date matches"
// compares this birth with that birth, never with some other event's date.
// A neighbor Align didn't map (New, Skip, unreachable) contributes nothing.
func attachExhibits(q Querier, g *graphcache.Graph, layer graphalign.Layer, prop *graphalign.Proposal, cfg graphalign.Config, namePattern string) error {
	if prop == nil || len(prop.Rows) == 0 {
		return nil
	}
	vocab, err := g.Vocabulary()
	if err != nil {
		return err
	}
	termLabels := map[string]string{}
	for _, t := range vocab.Terms {
		termLabels[t.PropertyKey+"|"+t.Key] = t.Label
	}
	kindOf := map[string]vocabkey.SubjectType{}
	for _, s := range layer.Subjects {
		kindOf[string(s.ID)] = vocabkey.SubjectType(s.Kind)
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
		obs, err := loadExhibitObservations(q, want, namePattern)
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
			exhibits := pairExhibits(bySubject[string(r.SubjectID)], nil, vocabkey.SubjectType(r.Kind), collect(plan.own), cfg)
			for _, np := range plan.neighbors {
				exhibits = append(exhibits, pairExhibits(bySubject[string(np.neighbor)], np.neighbor, kindOf[string(np.neighbor)], collect(np.sides), cfg)...)
			}
			r.Exhibits = exhibits
		}
	}
	for i := range prop.Rows {
		r := &prop.Rows[i]
		r.Exhibits = withRowLines(r.Exhibits, r.Comparisons, rowLabeler(termLabels, vocabkey.SubjectType(r.Kind), cfg.Subject))
	}
	return nil
}

// withRowLines completes a row's sheet lines from its comparisons. A row with
// no exhibit gets a line per comparison that spoke, so the sheet still shows
// the two values behind each outcome. A row with exhibits gets a line for each
// Property the row judges a conflict that the own-record exhibits do not
// already show as one: the majority disagreeing with a minority value, or a
// negative Observation, which the member picker never pairs. These lines
// carry no Observation, so they cannot be pinned. A derived comparison (the
// lifespan) has no copy yet and gets no line (S11-20).
func withRowLines(exhibits []graphalign.Exhibit, comparisons []graphalign.Comparison, label func(graphalign.Comparison, propertycompare.Value) string) []graphalign.Exhibit {
	none := len(exhibits) == 0
	for _, c := range comparisons {
		if c.Derived {
			continue
		}
		if none {
			if c.Result.Outcome == propertycompare.OutcomeUnknown {
				continue
			}
		} else {
			if c.Result.Outcome != propertycompare.OutcomeConflict {
				continue
			}
			shown := false
			for _, ex := range exhibits {
				shown = shown || len(ex.GroupSubjectID) == 0 && ex.PropertyOrigin == vocabkey.OriginProvenencia &&
					ex.PropertyKey == string(c.Key) && ex.Outcome == propertycompare.OutcomeConflict
			}
			if shown {
				continue
			}
		}
		if line, ok := refLine(c, label); ok {
			exhibits = append(exhibits, line)
		}
	}
	return exhibits
}

// refLine is the sheet line for one row comparison: its two values and dates,
// with no Observation behind either. A comparison with no speaking pair has
// no values to show.
func refLine(c graphalign.Comparison, label func(graphalign.Comparison, propertycompare.Value) string) (graphalign.Exhibit, bool) {
	sp := c.Result.Speaking
	if sp == nil {
		return graphalign.Exhibit{}, false
	}
	return graphalign.Exhibit{
		PropertyKey:      string(c.Key),
		PropertyOrigin:   vocabkey.OriginProvenencia,
		Outcome:          c.Result.Outcome,
		ValueType:        c.ValueType,
		Weight:           c.Points,
		IncomingDisplay:  label(c, sp.A),
		IncomingDate:     cloneDate(sp.A.Date),
		IncomingNegative: sp.A.Negative,
		MemberDisplay:    label(c, sp.B),
		MemberDate:       cloneDate(sp.B.Date),
	}, true
}

// rowLabeler words a ref line's values. A term shows its catalog label: the
// lead property's, else each pooled property's in order, else its key.
func rowLabeler(termLabels map[string]string, kind vocabkey.SubjectType, rules subjectcompare.Rules) func(graphalign.Comparison, propertycompare.Value) string {
	return func(c graphalign.Comparison, v propertycompare.Value) string {
		var label string
		if v.Type == propertymeta.ValueTypeTerm {
			props := append([]vocabkey.PropertyKey{c.Key}, rules.Pools[kind][c.Key].With...)
			for _, p := range props {
				if l, ok := termLabels[string(p)+"|"+v.Term]; ok {
					label = l
					break
				}
			}
		}
		return valueLabel(v, label)
	}
}

// valueLabel is a value as text for the sheet. termLabel is a term's catalog
// label; without one a term reads as its key.
func valueLabel(v propertycompare.Value, termLabel string) string {
	switch v.Type {
	case propertymeta.ValueTypeName:
		if v.Name == nil {
			return ""
		}
		if v.Name.Form != "" {
			return v.Name.Form
		}
		parts := make([]string, 0, len(v.Name.Parts))
		for _, p := range v.Name.Parts {
			parts = append(parts, p.Value)
		}
		return strings.Join(parts, " ")
	case propertymeta.ValueTypeDate:
		if v.Date == nil {
			return ""
		}
		return datevalue.CompactDisplay(*v.Date)
	case propertymeta.ValueTypeTerm:
		if termLabel != "" {
			return termLabel
		}
		return v.Term
	case propertymeta.ValueTypeInteger:
		return strconv.FormatInt(v.Integer, 10)
	}
	return v.Text
}

// pairExhibits pairs each incoming record with the closest record on the
// same Property. A same-value pair (Compatible) is preferred and pinned.
// The line's outcome comes from property compare, and its weight is what the
// record's property counts in the Subject's score (zero for a property the
// subject section does not weigh). neighbor is nil for the Subject's own
// records and names the layer neighbor for a one-hop group; kind is the
// Subject type of whichever Subject the incoming records are on.
func pairExhibits(incoming []exhibitObs, neighbor []byte, kind vocabkey.SubjectType, memberObs []exhibitObs, cfg graphalign.Config) []graphalign.Exhibit {
	var out []graphalign.Exhibit
	for _, in := range incoming {
		if in.valueType == "" || in.valueType == propertymeta.ValueTypeSubject {
			continue
		}
		meta := propertymeta.Meta{ValueType: in.valueType, Cardinality: in.cardinality}
		var chosen *exhibitObs
		var chosenRes propertycompare.Result
		var sim float64
		pinned := false
		for i := range memberObs {
			mem := &memberObs[i]
			if mem.key != in.key || mem.origin != in.origin || mem.valueType != in.valueType {
				continue
			}
			res := propertycompare.Compare([]propertycompare.Value{in.value}, []propertycompare.Value{mem.value}, meta, cfg.Value)
			if autoreconcile.Compatible(cfg.Value, in.valueType, in.auto, mem.auto) {
				chosen, chosenRes, pinned = mem, res, true
				sim = similarity(res)
				break
			}
			if res.Outcome == propertycompare.OutcomeUnknown {
				continue
			}
			if s := similarity(res); chosen == nil || s > sim {
				chosen, chosenRes, sim = mem, res, s
			}
		}
		if chosen == nil || chosenRes.Outcome == propertycompare.OutcomeUnknown {
			continue
		}
		var weight float64
		if in.origin == vocabkey.OriginProvenencia {
			if lead, ok := cfg.Subject.Lead(kind, vocabkey.PropertyKey(in.key)); ok {
				weight = subjectcompare.Points(chosenRes.Outcome, sim, cfg.Subject.Weights[kind][lead])
			}
		}
		out = append(out, graphalign.Exhibit{
			PropertyKey:    in.key,
			PropertyOrigin: in.origin,
			Outcome:        chosenRes.Outcome,
			ValueType:      in.valueType,
			// Compatible prefers the pair. A conflict is shown and left
			// unpinned; the sheet says a disagreement is not evidence.
			Pinned:                pinned && chosenRes.Outcome == propertycompare.OutcomeAgree,
			Weight:                weight,
			GroupSubjectID:        append([]byte(nil), neighbor...),
			IncomingObservationID: append([]byte(nil), in.id...),
			IncomingDisplay:       valueLabel(in.value, in.termLabel),
			IncomingSource:        in.source,
			IncomingDate:          cloneDate(in.value.Date),
			MemberObservationID:   append([]byte(nil), chosen.id...),
			MemberDisplay:         valueLabel(chosen.value, chosen.termLabel),
			MemberSource:          chosen.source,
			MemberDate:            cloneDate(chosen.value.Date),
		})
	}
	return out
}

// similarity is the speaking pair's similarity; 0 when no pair spoke.
func similarity(r propertycompare.Result) float64 {
	if r.Speaking == nil {
		return 0
	}
	return r.Speaking.Result.Similarity
}

func cloneDate(d *datevalue.Value) *datevalue.Value {
	if d == nil {
		return nil
	}
	c := *d
	return &c
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

func loadExhibitObservations(q Querier, subjectIDs [][]byte, namePattern string) ([]exhibitObs, error) {
	query := `SELECT o.id, o.subject_id, p.key, p.origin, p.value_type, p.cardinality,
			o.value_text, o.value_integer, o.value_date_id, o.value_name_id, o.value_term_id, t.key, t.label, COALESCE(t.neutral, 0), src.title
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
			text, term, termLabel sql.NullString
			integer               sql.NullInt64
			termID                []byte
		)
		if err := rows.Scan(&p.o.id, &p.o.subjectID, &p.o.key, &p.o.origin, &p.o.valueType, &p.o.cardinality,
			&text, &integer, &p.dateID, &p.nameID, &termID, &term, &termLabel, &p.o.auto.Neutral, &p.o.source); err != nil {
			return err
		}
		p.o.termLabel = termLabel.String
		p.o.auto.Text, p.o.auto.HasText = text.String, text.Valid
		p.o.auto.Integer, p.o.auto.HasInteger = integer.Int64, integer.Valid
		p.o.auto.TermID, p.o.auto.TermKey = termID, term.String
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
	out := make([]exhibitObs, 0, len(all))
	for _, p := range all {
		if d, ok := dates[string(p.dateID)]; ok {
			p.o.auto.Date = &d
		}
		if n, ok := names[string(p.nameID)]; ok {
			p.o.auto.Name = &n
		}
		p.o.value = subjectprofile.Value(p.o.auto, false, namePattern)
		p.o.value.Type = p.o.valueType
		out = append(out, p.o)
	}
	return out, nil
}
