// Package evidencelayer loads one Evidence graph's primary Subjects with their
// comparable values and its links, recorded and implied, kept in the Layers
// entry.
// Promote and the Person page share it. Decode, the one row decoder, also
// serves Promote's canon side, since a cached handle value carries the same
// value columns.
package evidencelayer

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// ErrNoGraph is returned for a graph id the cache does not find.
var ErrNoGraph = errors.New("evidencelayer: no such Evidence graph")

var errNilGraph = errors.New("evidencelayer: nil graph")

// Layer is one Evidence graph's primary Subjects and links.
type Layer struct {
	Subjects []subjectcompare.Subject // primary Subjects, by Ref
	Links    []kinship.Link           // recorded and implied, one per ends and keys
	byID     map[string]int
}

// Subject is the primary Subject with the id, and false when the graph has
// none.
func (l Layer) Subject(id []byte) (subjectcompare.Subject, bool) {
	i, ok := l.byID[string(id)]
	if !ok {
		return subjectcompare.Subject{}, false
	}
	return l.Subjects[i], true
}

// Layers is the matching layer of each Evidence graph, keyed by graph id. A
// layer reads the Evidence graph, the vocabulary, the terms, the name pattern
// and the static registry, so a vocabulary commit or a name format change
// empties the entry, and a commit that moves the graph's rows
// (Set.EvidenceGraphs) drops that graph's layer. A held layer needs no loaded
// graph. A reader never changes a layer, so every caller shares it.
var Layers graphcache.Keyed[Layer]

func init() {
	Layers = graphcache.RegisterKeyed(graphcache.KeyedEntry[Layer]{
		Name:        "evidencelayer.layers",
		DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}, nameformats.Default},
		Apply:       graphcache.DropNamedGraphs[Layer],
		Check:       checkLayer,
		Cap:         graphfill.MaxEvidenceGraphs,
	})
}

// Load is the Evidence graph's layer: the one held in Layers, or the one built
// from the loaded graph, held. A graph id the cache does not find is
// ErrNoGraph, and nothing is stored.
func Load(g *graphcache.Graph, graphID []byte) (Layer, error) {
	if g == nil {
		return Layer{}, errNilGraph
	}
	return Layers.Get(g, graphID, func() (Layer, error) {
		eg, err := graphfill.EvidenceGraph(g, graphID)
		if err != nil {
			return Layer{}, err
		}
		if !eg.Found {
			return Layer{}, ErrNoGraph
		}
		loaded, err := nameformats.LoadFor(g)
		if err != nil {
			return Layer{}, err
		}
		return Build(g, eg, loaded.Key)
	})
}

// checkLayer builds the graph's layer again from a fresh read of the graph,
// storing nothing, and compares.
func checkLayer(g *graphcache.Graph, graphID []byte, stored Layer) error {
	eg, err := graphfill.LoadEvidenceGraph(g, graphID)
	if err != nil {
		return err
	}
	if !eg.Found {
		return fmt.Errorf("evidencelayer: a layer is held for %x, which is no Evidence graph", graphID)
	}
	loaded, err := nameformats.LoadFor(g)
	if err != nil {
		return err
	}
	fresh, err := Build(g, eg, loaded.Key)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(stored, fresh) {
		return fmt.Errorf("evidencelayer: the stored layer of %x does not match a fresh build", graphID)
	}
	return nil
}

// participation is one person taking part in one event on the layer, in a role.
type participation struct {
	person, event []byte
	role          string
}

// Build makes the layer from a loaded Evidence graph under a name pattern. Each
// primary Subject's side is built here, once, with the profile builder.
func Build(g *graphcache.Graph, eg graphcache.EvidenceGraph, namePattern string) (Layer, error) {
	rules := matchregistry.Default()
	terms, err := canonlinks.LoadTerms(g)
	if err != nil {
		return Layer{}, err
	}
	vocab, err := g.Vocabulary()
	if err != nil {
		return Layer{}, err
	}
	props := vocab.PropertiesByID()
	primary := map[string]subjectcompare.Subject{}
	byKind := map[string][][]byte{}
	for _, s := range eg.Subjects {
		if !vocabkey.IsPrimaryType(s.TypeKey, s.TypeOrigin) {
			continue
		}
		if _, ok := rules.Subject.Weights[vocabkey.SubjectType(s.TypeKey)]; !ok {
			continue
		}
		primary[string(s.ID)] = subjectcompare.Subject{ID: append([]byte(nil), s.ID...), Ref: s.Ref, Kind: s.TypeKey}
		byKind[s.TypeKey] = append(byKind[s.TypeKey], append([]byte(nil), s.ID...))
	}
	recorded, participations, err := links(g, eg, props, vocab.TermsByID(), primary, byKind, terms)
	if err != nil {
		return Layer{}, err
	}
	ids := make(map[string]bool, len(primary))
	for id := range primary {
		ids[id] = true
	}
	candidates, err := layerValues(eg, ids, props, vocab.TermsByID())
	if err != nil {
		return Layer{}, err
	}

	// Events first: a person's life is read from the events it takes part in.
	built := map[string]subjectcompare.Values{}
	forSubject := func(s subjectcompare.Subject, life []subjectprofile.LifeEvent) subjectcompare.Values {
		return subjectprofile.ForSubject(subjectprofile.Subject{
			Values: candidates[string(s.ID)], NamePattern: namePattern, Life: life,
		}, rules.Subject.Lifespans[vocabkey.SubjectType(s.Kind)])
	}
	for id, s := range primary {
		if s.Kind == vocabkey.Event {
			built[id] = forSubject(s, nil)
		}
	}
	life := map[string][]subjectprofile.LifeEvent{}
	for _, p := range participations {
		if ev, ok := built[string(p.event)]; ok {
			life[string(p.person)] = append(life[string(p.person)], subjectprofile.LifeEventOf(ev, vocabkey.Role(p.role)))
		}
	}
	for id, s := range primary {
		if s.Kind != vocabkey.Event {
			built[id] = forSubject(s, life[id])
		}
	}
	var subjects []subjectcompare.Subject
	for _, s := range eg.Subjects {
		ps, ok := primary[string(s.ID)]
		if !ok {
			continue
		}
		subjects = append(subjects, subjectcompare.Subject{
			ID: ps.ID, Ref: ps.Ref, Kind: ps.Kind, Values: built[string(s.ID)],
		})
	}
	sort.Slice(subjects, func(i, j int) bool { return subjects[i].Ref < subjects[j].Ref })
	byID := make(map[string]int, len(subjects))
	for i, s := range subjects {
		byID[string(s.ID)] = i
	}
	return Layer{Subjects: subjects, Links: recorded, byID: byID}, nil
}

// neighborTypeProperty is the Property whose term types an event on a bridge
// key. Kinds without one key by bridge alone.
var neighborTypeProperty = map[string]string{vocabkey.Event: vocabkey.PropertyEventType}

// links walks every bridge connectrules registers, with any term, and keys
// each edge from both ends. It also returns the participations whose ends are
// both primary, which the lifespans read. Links go through one
// kinship.LinkSet, so links with the same ends and keys merge, and the links
// the kinship patterns imply follow the recorded ones (with no association of
// their own beyond those they were made from).
func links(g *graphcache.Graph, eg graphcache.EvidenceGraph, props map[string]graphcache.VocabProperty, vocabTerms map[string]graphcache.VocabTerm, primary map[string]subjectcompare.Subject, byKind map[string][][]byte, terms segmentcompare.Terms) ([]kinship.Link, []participation, error) {
	var parts []participation
	set := kinship.NewLinkSet()
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
			set.Add(kinship.Link{Edge: kinship.Edge{A: e.From, B: e.To, KeyA: keyA, KeyB: keyB, Associations: [][]byte{e.Association}}})
		}
	}
	set.WithImplied(terms, matchregistry.Default().Kinship)
	return set.Links(), parts, nil
}
