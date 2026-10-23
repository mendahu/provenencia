package promotealign

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The profile loader: both sides of a pair are built here, once per proposal,
// with the profile builder (layer 4), so graphalign stays pure.

// loadMetas is the catalog's value type and cardinality of every product
// property.
func loadMetas(g *graphcache.Graph) (subjectcompare.Metas, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	out := subjectcompare.Metas{}
	for _, p := range v.Properties {
		if p.Origin != vocabkey.OriginProvenencia {
			continue
		}
		out[vocabkey.PropertyKey(p.Key)] = propertymeta.Meta{ValueType: p.ValueType, Cardinality: p.Cardinality}
	}
	return out, nil
}

// propertiesByID maps each property id to its catalog row.
func propertiesByID(g *graphcache.Graph) (map[string]graphcache.VocabProperty, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	out := make(map[string]graphcache.VocabProperty, len(v.Properties))
	for _, p := range v.Properties {
		out[string(p.ID)] = p
	}
	return out, nil
}

// decode is the one row decoder for both sides: a Source Observation and a
// cached handle value carry the same value columns.
func decode(text string, hasText bool, integer int64, hasInteger bool, termID, date, name, image []byte, terms map[string]graphcache.VocabTerm) (autoreconcile.Value, error) {
	v := autoreconcile.Value{Text: text, HasText: hasText, Integer: integer, HasInteger: hasInteger}
	if len(termID) > 0 {
		t := terms[string(termID)]
		v.TermID = bytes.Clone(termID)
		v.TermKey = t.Key
		v.Neutral = t.Neutral
	}
	if len(date) > 0 {
		d, err := valuecodec.UnmarshalDate(date)
		if err != nil {
			return autoreconcile.Value{}, err
		}
		v.Date = &d
	}
	if len(name) > 0 {
		n, err := valuecodec.UnmarshalName(name)
		if err != nil {
			return autoreconcile.Value{}, err
		}
		v.Name = &n
	}
	if len(image) > 0 {
		v.Depiction = &depictionvalue.Value{ImageFileID: bytes.Clone(image)}
	}
	return v, nil
}

// reconciled is a cached handle value as the reconciler left it.
func reconciled(row graphcache.Value, terms map[string]graphcache.VocabTerm) (autoreconcile.ReconciledValue, error) {
	v, err := decode(row.Text, row.HasText, row.Integer, row.HasInteger, row.TermID, row.Date, row.Name, row.ImageFileID, terms)
	if err != nil {
		return autoreconcile.ReconciledValue{}, err
	}
	return autoreconcile.ReconciledValue{
		Value: v, Support: row.Support, Against: row.Against, Reason: autoreconcile.Reason(row.Reason),
	}, nil
}

// layerValues is each listed Subject's own product Observations, both
// polarities, by property key and in Observation id order. subjects is the set
// of Subject ids to read.
func layerValues(eg graphcache.EvidenceGraph, subjects map[string]bool, props map[string]graphcache.VocabProperty, terms map[string]graphcache.VocabTerm) (map[string]map[vocabkey.PropertyKey][]autoreconcile.Candidate, error) {
	obs := slices.Clone(eg.Observations)
	slices.SortStableFunc(obs, func(a, b graphcache.EvidenceObservation) int { return bytes.Compare(a.ID, b.ID) })
	out := map[string]map[vocabkey.PropertyKey][]autoreconcile.Candidate{}
	for _, o := range obs {
		if !subjects[string(o.SubjectID)] || len(o.ValueSubjectID) > 0 {
			continue
		}
		prop, ok := props[string(o.PropertyID)]
		if !ok || prop.Origin != vocabkey.OriginProvenencia {
			continue
		}
		v, err := decode(o.Text, o.HasText, o.Integer, o.HasInteger, o.TermID, o.Date, o.Name, o.DepictionImageFileID, terms)
		if err != nil {
			return nil, err
		}
		key := vocabkey.PropertyKey(prop.Key)
		if out[string(o.SubjectID)] == nil {
			out[string(o.SubjectID)] = map[vocabkey.PropertyKey][]autoreconcile.Candidate{}
		}
		out[string(o.SubjectID)][key] = append(out[string(o.SubjectID)][key], autoreconcile.Candidate{
			ObservationID: bytes.Clone(o.ID),
			Value:         v,
			Negative:      o.Polarity == observations.PolarityNegative,
		})
	}
	return out, nil
}

// lifeEvent reads one event an entity takes part in from the event's built
// profile: its type, and every date it records.
func lifeEvent(event subjectcompare.Values, role vocabkey.Role) subjectprofile.LifeEvent {
	le := subjectprofile.LifeEvent{Role: role}
	for _, v := range event[vocabkey.PropertyEventType] {
		if v.Positive() {
			le.EventType = vocabkey.EventType(v.Term)
			break
		}
	}
	for _, key := range []vocabkey.PropertyKey{vocabkey.PropertyDate, vocabkey.PropertyStartDate, vocabkey.PropertyEndDate} {
		for _, v := range event[key] {
			if v.Positive() && v.Date != nil {
				le.Dates = append(le.Dates, *v.Date)
			}
		}
	}
	return le
}

// profile is a built handle: its values.
type profile struct {
	values subjectcompare.Values
}

// profiler builds handle profiles from the canon graph, each once for as long
// as the graph's match index lives. done is the index's profile map. A caller
// keeps the profiles with keep(); nothing else stores them.
type profiler struct {
	g           *graphcache.Graph
	props       map[string]graphcache.VocabProperty
	terms       map[string]graphcache.VocabTerm
	metas       subjectcompare.Metas
	types       map[string][]byte
	rules       subjectcompare.Rules
	reconcile   valuecompare.Rules
	namePattern string
	done        map[string]profile
}

func newProfiler(g *graphcache.Graph, metas subjectcompare.Metas, rules subjectcompare.Rules, namePattern string) (*profiler, error) {
	props, err := propertiesByID(g)
	if err != nil {
		return nil, err
	}
	vocab, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	types, err := typeByKey(g)
	if err != nil {
		return nil, err
	}
	// Profiles are kept with the match index, for one name pattern at a time:
	// this profiler shares the index's map when it has one for this pattern,
	// and starts its own otherwise. keep stores a map into the index.
	m := matchIndexOf(g)
	done := m.profiles
	if done == nil || m.profilesPattern != namePattern {
		done = map[string]profile{}
	}
	return &profiler{
		g: g, props: props, terms: vocab.TermsByID(), metas: metas, types: types, rules: rules,
		reconcile: matchregistry.Default().Value, namePattern: namePattern,
		done: done,
	}, nil
}

// keep stores this profiler's profiles in the graph's match index, creating the
// index slot when there is none. The profiles then live as long as the index.
// A map already stored for this name pattern is this profiler's own, so it is
// left alone.
func (p *profiler) keep() {
	m := matchIndexOf(p.g)
	if m.profiles != nil && m.profilesPattern == p.namePattern {
		return
	}
	m.profiles, m.profilesPattern = p.done, p.namePattern
	p.g.SetMatchIndex(m)
}

func (p *profiler) kind(n *graphcache.Node) vocabkey.SubjectType {
	return vocabkey.SubjectType(typeKey(p.types, n.SubjectTypeID))
}

// profile builds every node not built yet. A node whose kind has a lifespan
// reads its life from the events it takes part in, so all of those events are
// loaded with one call and built first.
func (p *profiler) profile(nodes []*graphcache.Node) error {
	var todo []*graphcache.Node
	queued := map[string]bool{}
	for _, n := range nodes {
		if n == nil || n.Merged || queued[string(n.ID)] {
			continue
		}
		if _, ok := p.done[string(n.ID)]; ok {
			continue
		}
		queued[string(n.ID)] = true
		todo = append(todo, n)
	}
	if len(todo) == 0 {
		return nil
	}
	var neighborIDs [][]byte
	for _, n := range todo {
		if len(p.rules.Lifespans[p.kind(n)]) == 0 {
			continue
		}
		links, err := p.g.Follow(n.ID, hopPersonToEvent)
		if err != nil {
			return err
		}
		for _, l := range links {
			neighborIDs = append(neighborIDs, l.Neighbor)
		}
	}
	if err := p.g.Nodes(neighborIDs); err != nil {
		return err
	}
	var events []*graphcache.Node
	for _, id := range neighborIDs {
		n, err := p.g.Node(id)
		if err != nil {
			return err
		}
		events = append(events, n)
	}
	for _, group := range [][]*graphcache.Node{events, todo} {
		for _, n := range group {
			if n == nil || n.Merged {
				continue
			}
			if _, ok := p.done[string(n.ID)]; ok {
				continue
			}
			if err := p.build(n); err != nil {
				return err
			}
		}
	}
	return nil
}

// entity is a node's reconciled properties as the profile builder reads them,
// with no life.
func (p *profiler) entity(n *graphcache.Node) (subjectprofile.Entity, error) {
	props := map[vocabkey.PropertyKey]subjectprofile.Property{}
	for propID, rows := range n.Values {
		prop, ok := p.props[propID]
		if !ok || prop.Origin != vocabkey.OriginProvenencia {
			continue
		}
		vals := make([]autoreconcile.ReconciledValue, 0, len(rows))
		for _, row := range rows {
			rv, err := reconciled(row, p.terms)
			if err != nil {
				return subjectprofile.Entity{}, err
			}
			vals = append(vals, rv)
		}
		props[vocabkey.PropertyKey(prop.Key)] = subjectprofile.Property{Values: vals}
	}
	return subjectprofile.Entity{Properties: props, NamePattern: p.namePattern}, nil
}

// keyValues is a node's values for shortlist keys: the profile with no derived
// lead, so no life event is read. It is not kept in done, which holds only full
// profiles.
func (p *profiler) keyValues(n *graphcache.Node) (subjectcompare.Values, error) {
	e, err := p.entity(n)
	if err != nil {
		return nil, err
	}
	vals, err := subjectprofile.Build(e, p.metas, p.reconcile, nil, subjectprofile.Options{})
	return vals, err
}

// build makes one node's profile. A kind with a lifespan reads its life from
// the already built events it is linked to.
func (p *profiler) build(n *graphcache.Node) error {
	kind := p.kind(n)
	e, err := p.entity(n)
	if err != nil {
		return err
	}
	if len(p.rules.Lifespans[kind]) > 0 {
		links, err := p.g.Follow(n.ID, hopPersonToEvent)
		if err != nil {
			return err
		}
		for _, l := range links {
			if ev, ok := p.done[string(l.Neighbor)]; ok {
				e.Life = append(e.Life, lifeEvent(ev.values, vocabkey.Role(l.TermKey)))
			}
		}
	}
	vals, err := subjectprofile.Build(e, p.metas, p.reconcile, p.rules.Lifespans[kind], subjectprofile.Options{})
	if err != nil {
		return err
	}
	p.done[string(n.ID)] = profile{values: vals}
	return nil
}

// values is a built node's profile.
func (p *profiler) values(id []byte) subjectcompare.Values {
	if v, ok := p.done[string(id)]; ok {
		return v.values
	}
	return subjectcompare.Values{}
}
