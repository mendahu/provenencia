package graphcache

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/mendahu/provenencia/core/hops"
)

// EvidenceGraphLoader fills one Evidence graph. graphfill registers it. The
// graph does not import that package.
type EvidenceGraphLoader func(g *Graph, graphID []byte) (EvidenceGraph, error)

// EvidenceGraphCheck reloads one stored Evidence graph and compares the whole
// graph. graphfill registers it.
type EvidenceGraphCheck func(g *Graph, graphID []byte, stored EvidenceGraph) error

var (
	evidenceGraphLoad EvidenceGraphLoader
	evidenceGraphFn   EvidenceGraphCheck
)

// SetEvidenceGraphLoader registers the read that fills an Evidence graph.
func SetEvidenceGraphLoader(fn EvidenceGraphLoader) { evidenceGraphLoad = fn }

// SetEvidenceGraphCheck registers the comparison Verify runs for a loaded
// Evidence graph.
func SetEvidenceGraphCheck(fn EvidenceGraphCheck) { evidenceGraphFn = fn }

// EvidenceSubject is one subject on an Evidence graph. TypeKey and TypeOrigin are the
// subject type's identity. Labels of vocabulary are not stored.
type EvidenceSubject struct {
	ID            []byte
	SourceID      []byte
	SubjectTypeID []byte
	Ref           string
	Label         string
	Description   string
	TypeKey       string
	TypeOrigin    string
}

// EvidenceObservation is one observation on an Evidence graph. Property and term labels are applied
// when the row is shown. Date and Name are the stored value blobs.
type EvidenceObservation struct {
	ID             []byte
	Ref            string
	CitationID     []byte
	SubjectID      []byte
	PropertyID     []byte
	Polarity       string
	Text           string
	HasText        bool
	Integer        int64
	HasInteger     bool
	DateID         []byte
	NameID         []byte
	ValueSubjectID []byte
	TermID         []byte
	DepictionID    []byte
	// DepictionRegion and DepictionImageFileID are the depiction as
	// depictionvalue.Readable returns it: empty when the stored half is not
	// one to show.
	DepictionRegion      string
	DepictionImageFileID []byte
	Date                 []byte
	Name                 []byte
}

// EvidenceMember is an accepted claim on an Evidence graph: the subject, the claim, and the handle.
// The handle's name and label are read from its node.
type EvidenceMember struct {
	SubjectID []byte
	ClaimID   []byte
	EntityID  []byte
	Kind      string
}

// EvidenceEdge is one hop on an Evidence graph, kept so an event title can be chosen
// from the stored observations.
type EvidenceEdge struct {
	From           []byte
	Association    []byte
	AssociationRef string
	To             []byte
	ToRef          string
}

// EvidenceGraph is one loaded Evidence graph. SourceID and SourceTitle are the
// graph's Source, set by the loader; a graph whose row does not exist is not
// Found and has no SourceID, so no Source drop removes it.
type EvidenceGraph struct {
	GraphID      []byte
	SourceID     []byte
	SourceTitle  string
	Found        bool
	Subjects     []EvidenceSubject
	Observations []EvidenceObservation
	Members      []EvidenceMember
	EventPeople  []EvidenceEdge
	EventPlaces  []EvidenceEdge
}

// EvidenceGraph returns the loaded Evidence graph, filling it on first read.
func (g *Graph) EvidenceGraph(graphID []byte) (EvidenceGraph, error) {
	if g == nil || len(graphID) != 16 {
		return EvidenceGraph{}, fmt.Errorf("graphcache: evidence graph id is invalid")
	}
	if eg, ok := g.evidenceGraphs[string(graphID)]; ok {
		return eg, nil
	}
	if evidenceGraphLoad == nil {
		return EvidenceGraph{}, fmt.Errorf("graphcache: evidence graph loader is not registered")
	}
	loaded, err := evidenceGraphLoad(g, graphID)
	if err != nil {
		return EvidenceGraph{}, err
	}
	loaded.GraphID = append([]byte(nil), graphID...)
	g.evidenceGraphs[string(graphID)] = loaded
	return loaded, nil
}

// dropGraphsOfSources removes every loaded Evidence graph of each Source, with
// its memos. Set.Sources names Sources until effects carry graph ids.
func (g *Graph) dropGraphsOfSources(ids [][]byte) {
	if g == nil || len(ids) == 0 {
		return
	}
	for key, eg := range g.evidenceGraphs {
		for _, id := range ids {
			if len(id) == 16 && bytes.Equal(eg.SourceID, id) {
				delete(g.evidenceGraphs, key)
				delete(g.memos, key)
				break
			}
		}
	}
}

// EvidenceGraphMemo returns what a feature stored for a loaded Evidence graph
// under key. It is held only while the graph is loaded: a commit that names
// the graph's Source, Drop, a vocabulary commit and SetAliasTables clear it. A
// memoized value is shared, so every reader treats it as read-only.
func (g *Graph) EvidenceGraphMemo(graphID []byte, key string) (any, bool) {
	if g == nil {
		return nil, false
	}
	v, ok := g.memos[string(graphID)][key]
	return v, ok
}

// SetEvidenceGraphMemo stores a feature's value for a loaded Evidence graph
// under key. A graph that is not loaded keeps nothing.
func (g *Graph) SetEvidenceGraphMemo(graphID []byte, key string, v any) {
	if g == nil {
		return
	}
	if _, loaded := g.evidenceGraphs[string(graphID)]; !loaded {
		return
	}
	if g.memos[string(graphID)] == nil {
		g.memos[string(graphID)] = map[string]any{}
	}
	g.memos[string(graphID)][key] = v
}

// WalkEvidenceGraph follows h through one Evidence graph from each Subject
// in from, in memory: Association is the bridge Subject. A bridge Subject is
// one of the hop's bridge type on eg. It cites a from Subject on the hop's
// from-endpoint property and a Subject on eg on its to-endpoint property,
// both positively, and the filter term on the filter property when the hop
// has one. Each edge appears once, ordered by from Subject id, then the
// bridge's ref, then the to Subject's ref. A name the vocabulary does not
// know reaches nothing.
func (g *Graph) WalkEvidenceGraph(eg EvidenceGraph, h hops.Hop, from [][]byte) ([]EvidenceEdge, error) {
	if h.Bridge() == "" {
		return nil, fmt.Errorf("graphcache: zero Hop")
	}
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	var bridgeType []byte
	for _, t := range v.Types {
		if t.Key == h.Bridge() && t.Origin == h.Origin() {
			bridgeType = t.ID
		}
	}
	propID := func(key string) []byte {
		for _, p := range v.Properties {
			if p.Key == key && p.Origin == h.Origin() {
				return p.ID
			}
		}
		return nil
	}
	fromProp, toProp := propID(h.From()), propID(h.To())
	if len(bridgeType) == 0 || len(fromProp) == 0 || len(toProp) == 0 {
		return nil, nil
	}
	var filterProp []byte
	filterTerms := map[string]bool{}
	filterKey, termKey, filtered := h.Filter()
	if filtered {
		if filterProp = propID(filterKey); len(filterProp) == 0 {
			return nil, nil
		}
		for _, t := range v.Terms {
			if t.Key == termKey && t.Origin == h.Origin() {
				filterTerms[string(t.ID)] = true
			}
		}
	}

	wanted := make(map[string]bool, len(from))
	for _, id := range from {
		wanted[string(id)] = true
	}
	subjects := make(map[string]EvidenceSubject, len(eg.Subjects))
	for _, s := range eg.Subjects {
		subjects[string(s.ID)] = s
	}
	// What each bridge Subject cites, positively.
	type cited struct {
		tos      [][]byte
		filtered bool
	}
	bridges := map[string]*cited{}
	type reach struct{ from, bridge []byte }
	var reaches []reach
	for _, o := range eg.Observations {
		if o.Polarity != "positive" {
			continue
		}
		s, ok := subjects[string(o.SubjectID)]
		if !ok || !bytes.Equal(s.SubjectTypeID, bridgeType) {
			continue
		}
		c := bridges[string(o.SubjectID)]
		if c == nil {
			c = &cited{}
			bridges[string(o.SubjectID)] = c
		}
		switch {
		case bytes.Equal(o.PropertyID, fromProp) && wanted[string(o.ValueSubjectID)]:
			reaches = append(reaches, reach{from: o.ValueSubjectID, bridge: o.SubjectID})
		case bytes.Equal(o.PropertyID, toProp) && len(o.ValueSubjectID) > 0:
			if _, on := subjects[string(o.ValueSubjectID)]; on {
				c.tos = append(c.tos, o.ValueSubjectID)
			}
		}
		if filtered && bytes.Equal(o.PropertyID, filterProp) && filterTerms[string(o.TermID)] {
			c.filtered = true
		}
	}
	var out []EvidenceEdge
	seen := map[string]bool{}
	for _, r := range reaches {
		c := bridges[string(r.bridge)]
		if filtered && !c.filtered {
			continue
		}
		bridge := subjects[string(r.bridge)]
		for _, to := range c.tos {
			k := string(r.from) + "\x00" + string(r.bridge) + "\x00" + string(to)
			if seen[k] {
				continue
			}
			seen[k] = true
			dst := subjects[string(to)]
			out = append(out, EvidenceEdge{
				From: bytes.Clone(r.from), Association: bytes.Clone(r.bridge), AssociationRef: bridge.Ref,
				To: bytes.Clone(to), ToRef: dst.Ref,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if c := bytes.Compare(a.From, b.From); c != 0 {
			return c < 0
		}
		if a.AssociationRef != b.AssociationRef {
			return a.AssociationRef < b.AssociationRef
		}
		return a.ToRef < b.ToRef
	})
	return out, nil
}
