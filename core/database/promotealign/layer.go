package promotealign

import (
	"bytes"
	"slices"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/hops"
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

// layerMemoKey is the Source memo key of a built layer, with the name pattern
// it was built under after it.
const layerMemoKey = "promotealign.layer\x00"

// layerMemo is a layer and its primary Subjects, as one Source memo.
type layerMemo struct {
	layer   graphalign.Layer
	primary map[string]primarySubject
}

// loadLayerMemo is the Source's layer: the one memoized on its graph, or
// loadLayer's, memoized. The layer reads the Source graph, the vocabulary, the
// terms, the name pattern and the static registry, and the memo is dropped with
// the Source, a vocabulary change and new alias tables. A reader never changes
// a layer, so every caller shares it.
func loadLayerMemo(g *graphcache.Graph, sourceID []byte, cfg graphalign.Config, namePattern string, terms segmentcompare.Terms) (graphalign.Layer, map[string]primarySubject, error) {
	key := layerMemoKey + namePattern
	if v, ok := g.SourceMemo(sourceID, key); ok {
		m := v.(layerMemo)
		return m.layer, m.primary, nil
	}
	layer, primary, err := loadLayer(g, sourceID, cfg, namePattern, terms)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	g.SetSourceMemo(sourceID, key, layerMemo{layer: layer, primary: primary})
	return layer, primary, nil
}

// loadLayer builds the Evidence layer from a loaded Source. Each primary
// Subject's side is built here, once, with the profile builder; property
// metas stay a catalog read, since they are not part of one Source.
func loadLayer(g *graphcache.Graph, sourceID []byte, cfg graphalign.Config, namePattern string, terms segmentcompare.Terms) (graphalign.Layer, map[string]primarySubject, error) {
	if g == nil {
		return graphalign.Layer{}, nil, promote.ErrInvalid
	}
	sg, err := g.Source(sourceID)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	props, err := propertiesByID(g)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	vocab, err := g.Vocabulary()
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	primary := map[string]primarySubject{}
	byKind := map[string][][]byte{}
	for _, s := range sg.Subjects {
		if !vocabkey.IsPrimaryType(s.TypeKey, s.TypeOrigin) {
			continue
		}
		if _, ok := cfg.Subject.Weights[vocabkey.SubjectType(s.TypeKey)]; !ok {
			continue
		}
		primary[string(s.ID)] = primarySubject{id: append([]byte(nil), s.ID...), ref: s.Ref, kind: s.TypeKey}
		byKind[s.TypeKey] = append(byKind[s.TypeKey], append([]byte(nil), s.ID...))
	}
	bridges, participations, err := loadLayerBridges(g, sg, props, vocab.TermsByID(), primary, byKind, terms)
	if err != nil {
		return graphalign.Layer{}, nil, err
	}
	ids := make(map[string]bool, len(primary))
	for id := range primary {
		ids[id] = true
	}
	candidates, err := layerValues(sg, ids, props, vocab.TermsByID())
	if err != nil {
		return graphalign.Layer{}, nil, err
	}

	// Events first: a person's life is read from the events it takes part in.
	built := map[string]subjectcompare.Values{}
	forSubject := func(s primarySubject, life []subjectprofile.LifeEvent) subjectcompare.Values {
		return subjectprofile.ForSubject(subjectprofile.Subject{
			Values: candidates[string(s.id)], NamePattern: namePattern, Life: life,
		}, cfg.Subject.Lifespans[vocabkey.SubjectType(s.kind)])
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
	for _, s := range sg.Subjects {
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
// ends are both primary, which the lifespans read.
func loadLayerBridges(g *graphcache.Graph, sg graphcache.SourceGraph, props map[string]graphcache.VocabProperty, vocabTerms map[string]graphcache.VocabTerm, primary map[string]primarySubject, byKind map[string][][]byte, terms segmentcompare.Terms) ([]graphalign.Bridge, []participation, error) {
	var out []graphalign.Bridge
	var parts []participation
	seen := map[seenEdge]int{} // the edge's index in out
	for _, b := range connectrules.Bridges() {
		if len(b.Endpoints) != 2 || len(byKind[b.Endpoints[0].TypeKey]) == 0 {
			continue
		}
		hop, err := hops.NewHop(b.BridgeTypeKey, b.Endpoints[0].PropertyKey, b.Endpoints[1].PropertyKey, nil)
		if err != nil {
			return nil, nil, err
		}
		edges, err := g.WalkSource(sg, hop, byKind[b.Endpoints[0].TypeKey])
		if err != nil {
			return nil, nil, err
		}
		assocs := make([][]byte, len(edges))
		for i, e := range edges {
			assocs[i] = e.Association
		}
		var termOf map[string]string
		if connectrules.HasDisambiguation(b.Disambiguation) {
			termOf = observedTerms(sg, props, vocabTerms, assocs, b.Disambiguation)
		}
		// The event type comes from whichever end is an event; linkOf alone
		// decides whether the key keeps it.
		eventFirst, hasEvent := eventEndpoint(b)
		eventEnd := func(e graphcache.SourceEdge) []byte {
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
			eventTypes = observedTerms(sg, props, vocabTerms, ids, neighborTypeProperty[vocabkey.Event])
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
			keyA := segmentcompare.Key(linkOf(b, term, eventType, true), terms)
			keyB := segmentcompare.Key(linkOf(b, term, eventType, false), terms)
			ek := edgeSeenKey(e.From, e.To, keyA, keyB)
			if at, ok := seen[ek]; ok {
				if !slices.ContainsFunc(out[at].Associations, func(a []byte) bool { return bytes.Equal(a, e.Association) }) {
					out[at].Associations = append(out[at].Associations, bytes.Clone(e.Association))
				}
				continue
			}
			seen[ek] = len(out)
			out = append(out, graphalign.Bridge{
				A: append([]byte(nil), e.From...), B: append([]byte(nil), e.To...), KeyA: keyA, KeyB: keyB,
				Associations: [][]byte{bytes.Clone(e.Association)},
			})
		}
	}
	return out, parts, nil
}
