package promotealign

import (
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type primarySubject struct {
	id   []byte
	ref  string
	kind string
}

// participation is one person taking part in one event on the layer, in a role.
type participation struct {
	person, event []byte
	role          string
}

// layerMemoKey is the Evidence graph memo key of a built layer, with the name pattern
// it was built under after it.
const layerMemoKey = "promotealign.layer\x00"

// layerMemo is a layer and its primary Subjects, as one Evidence graph memo.
type layerMemo struct {
	layer   graphalign.Layer
	primary map[string]primarySubject
}

// loadLayerMemo is the Evidence graph's layer: the one memoized on its graph,
// or loadLayer's, memoized. The layer reads the Evidence graph, the vocabulary,
// the terms, the name pattern and the static registry, and the memo is dropped
// with the graph (a commit that names its Source), a vocabulary change and new
// alias tables. A reader never changes
// a layer, so every caller shares it.
func loadLayerMemo(g *graphcache.Graph, graphID []byte, cfg graphalign.Config, namePattern string, terms segmentcompare.Terms) (graphalign.Layer, map[string]primarySubject, error) {
	key := layerMemoKey + namePattern
	if v, ok := g.EvidenceGraphMemo(graphID, key); ok {
		if m, ok := v.(layerMemo); ok {
			return m.layer, m.primary, nil
		}
	}
	layer, primary, err := loadLayer(g, graphID, cfg, namePattern, terms)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	g.SetEvidenceGraphMemo(graphID, key, layerMemo{layer: layer, primary: primary})
	return layer, primary, nil
}

// loadLayer builds the Evidence layer from a loaded Evidence graph. Each
// primary Subject's side is built here, once, with the profile builder;
// property metas stay a catalog read, since they are not part of one graph.
func loadLayer(g *graphcache.Graph, graphID []byte, cfg graphalign.Config, namePattern string, terms segmentcompare.Terms) (graphalign.Layer, map[string]primarySubject, error) {
	if g == nil {
		return graphalign.Layer{}, nil, promote.ErrInvalid
	}
	eg, err := g.EvidenceGraph(graphID)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	vocab, err := g.Vocabulary()
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	props := vocab.PropertiesByID()
	primary := map[string]primarySubject{}
	byKind := map[string][][]byte{}
	for _, s := range eg.Subjects {
		if !vocabkey.IsPrimaryType(s.TypeKey, s.TypeOrigin) {
			continue
		}
		if _, ok := cfg.Subject.Weights[vocabkey.SubjectType(s.TypeKey)]; !ok {
			continue
		}
		primary[string(s.ID)] = primarySubject{id: append([]byte(nil), s.ID...), ref: s.Ref, kind: s.TypeKey}
		byKind[s.TypeKey] = append(byKind[s.TypeKey], append([]byte(nil), s.ID...))
	}
	bridges, participations, err := loadLayerBridges(g, eg, props, vocab.TermsByID(), primary, byKind, terms)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	ids := make(map[string]bool, len(primary))
	for id := range primary {
		ids[id] = true
	}
	candidates, err := layerValues(eg, ids, props, vocab.TermsByID())
	if err != nil {
		return graphalign.Layer{}, nil, err
	}

	// Events first: a person's life is read from the events it takes part in.
	built := map[string]subjectcompare.Values{}
	forSubject := func(s primarySubject, life []subjectprofile.LifeEvent) subjectcompare.Values {
		vals := subjectprofile.ForSubject(subjectprofile.Subject{
			Values: candidates[string(s.id)], NamePattern: namePattern, Life: life,
		}, cfg.Subject.Lifespans[vocabkey.SubjectType(s.kind)])
		return vals
	}
	for id, s := range primary {
		if s.kind == vocabkey.Event {
			built[id] = forSubject(s, nil)
		}
	}
	life := map[string][]subjectprofile.LifeEvent{}
	for _, p := range participations {
		if ev, ok := built[string(p.event)]; ok {
			life[string(p.person)] = append(life[string(p.person)], lifeEvent(ev, vocabkey.Role(p.role)))
		}
	}
	for id, s := range primary {
		if s.kind != vocabkey.Event {
			built[id] = forSubject(s, life[id])
		}
	}
	var subjects []graphalign.Subject
	for _, s := range eg.Subjects {
		ps, ok := primary[string(s.ID)]
		if !ok {
			continue
		}
		subjects = append(subjects, graphalign.Subject{
			ID: ps.id, Ref: ps.ref, Kind: ps.kind, Values: built[string(s.ID)], Provenance: 1,
		})
	}
	sort.Slice(subjects, func(i, j int) bool { return subjects[i].Ref < subjects[j].Ref })

	metas, err := loadMetas(g)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	return graphalign.Layer{Subjects: subjects, Bridges: bridges, Metas: metas}, primary, nil
}

// neighborTypeProperty is the Property whose term types an event on a bridge
// key. Kinds without one key by bridge alone.
var neighborTypeProperty = map[string]string{vocabkey.Event: vocabkey.PropertyEventType}

// hopPersonToEvent follows a person to the events it takes part in, any role.
// The profile builder reads it.
var hopPersonToEvent = hops.MustHop(vocabkey.BridgeParticipation, vocabkey.PropertyPerson, vocabkey.PropertyEvent, nil)

// loadLayerBridges walks every bridge connectrules registers, with any term,
// and keys each edge from both ends. It also returns the participations whose
// ends are both primary, which the lifespans read. Links go through one
// linkSet, so links with the same ends and keys merge, and the links the
// kinship patterns imply follow the recorded ones (with no association of
// their own beyond those they were made from).
func loadLayerBridges(g *graphcache.Graph, eg graphcache.EvidenceGraph, props map[string]graphcache.VocabProperty, vocabTerms map[string]graphcache.VocabTerm, primary map[string]primarySubject, byKind map[string][][]byte, terms segmentcompare.Terms) ([]graphalign.Bridge, []participation, error) {
	var parts []participation
	set := newLinkSet()
	for _, b := range connectrules.Bridges() {
		if len(b.Endpoints) != 2 || len(byKind[b.Endpoints[0].TypeKey]) == 0 {
			continue
		}
		hop, err := hops.NewHop(b.BridgeTypeKey, b.Endpoints[0].PropertyKey, b.Endpoints[1].PropertyKey, nil)
		if err != nil {
			return nil, nil, err
		}
		edges, err := g.WalkEvidenceGraph(eg, hop, byKind[b.Endpoints[0].TypeKey])
		if err != nil {
			return nil, nil, err
		}
		assocs := make([][]byte, len(edges))
		for i, e := range edges {
			assocs[i] = e.Association
		}
		var termOf map[string]string
		if connectrules.HasDisambiguation(b.Disambiguation) {
			termOf = observedTerms(eg, props, vocabTerms, assocs, b.Disambiguation)
		}
		// The event type comes from whichever end is an event; `segmentcompare.Keys` alone
		// decides whether the key keeps it.
		eventFirst, hasEvent := segmentcompare.EventEndpoint(b)
		eventEnd := func(e graphcache.EvidenceEdge) []byte {
			if eventFirst {
				return e.From
			}
			return e.To
		}
		var eventTypes map[string]string
		if hasEvent {
			ids := make([][]byte, len(edges))
			for i, e := range edges {
				ids[i] = eventEnd(e)
			}
			eventTypes = observedTerms(eg, props, vocabTerms, ids, neighborTypeProperty[vocabkey.Event])
		}
		for _, e := range edges {
			term := termOf[string(e.Association)]
			_, aOK := primary[string(e.From)]
			_, bOK := primary[string(e.To)]
			if b.BridgeTypeKey == vocabkey.BridgeParticipation && aOK && bOK {
				parts = append(parts, participation{person: e.From, event: e.To, role: term})
			}
			if !aOK || !bOK {
				continue
			}
			var eventType string
			if hasEvent {
				eventType = eventTypes[string(eventEnd(e))]
			}
			keyA, keyB := segmentcompare.Keys(b, vocabkey.BridgeTerm(term), vocabkey.EventType(eventType), terms)
			set.add(sideLink{Edge: kinship.Edge{A: e.From, B: e.To, KeyA: keyA, KeyB: keyB, Associations: [][]byte{e.Association}}})
		}
	}
	set.withImplied(terms, matchregistry.Default().Kinship)
	var out []graphalign.Bridge
	for _, l := range set.links {
		out = append(out, graphalign.Bridge{
			A: l.A, B: l.B, KeyA: l.KeyA, KeyB: l.KeyB, Associations: l.Associations, Recording: l.Recording,
		})
	}
	return out, parts, nil
}
