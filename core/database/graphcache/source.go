package graphcache

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/mendahu/provenencia/core/hops"
)

// SourceLoader fills one Source. sourcefill registers it. The graph does
// not import that package.
type SourceLoader func(g *Graph, sourceID []byte) (SourceGraph, error)

// SourceCheck reloads one stored Source and compares the whole graph.
// sourcefill registers it.
type SourceCheck func(g *Graph, id []byte, stored SourceGraph) error

var (
	sourceLoad SourceLoader
	sourceFn   SourceCheck
)

// SetSourceLoader registers the read that fills a Source graph.
func SetSourceLoader(fn SourceLoader) { sourceLoad = fn }

// SetSourceCheck registers the comparison Verify runs for a loaded Source.
func SetSourceCheck(fn SourceCheck) { sourceFn = fn }

// SourceSubject is one subject on a Source. TypeKey and TypeOrigin are the
// subject type's identity. Labels of vocabulary are not stored.
type SourceSubject struct {
	ID            []byte
	SourceID      []byte
	SubjectTypeID []byte
	Ref           string
	Label         string
	Description   string
	TypeKey       string
	TypeOrigin    string
}

// SourceObservation is one observation. Property and term labels are applied
// when the row is shown. Date and Name are the stored value blobs.
type SourceObservation struct {
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

// SourceMember is an accepted claim: the subject, the claim, and the handle.
// The handle's name and label are read from its node.
type SourceMember struct {
	SubjectID []byte
	ClaimID   []byte
	EntityID  []byte
	Kind      string
}

// SourceEdge is one hop on the Source, kept so an event title can be chosen
// from the stored observations.
type SourceEdge struct {
	From           []byte
	Association    []byte
	AssociationRef string
	To             []byte
	ToRef          string
}

// SourceGraph is one loaded Source.
type SourceGraph struct {
	SourceID     []byte
	Subjects     []SourceSubject
	Observations []SourceObservation
	Members      []SourceMember
	EventPeople  []SourceEdge
	EventPlaces  []SourceEdge
}

// Source returns the loaded Source, filling it on first read.
func (g *Graph) Source(id []byte) (SourceGraph, error) {
	if g == nil || len(id) != 16 {
		return SourceGraph{}, fmt.Errorf("graphcache: source id is invalid")
	}
	if s, ok := g.sources[string(id)]; ok {
		return s, nil
	}
	if sourceLoad == nil {
		return SourceGraph{}, fmt.Errorf("graphcache: source loader is not registered")
	}
	loaded, err := sourceLoad(g, id)
	if err != nil {
		return SourceGraph{}, err
	}
	loaded.SourceID = append([]byte(nil), id...)
	g.sources[string(id)] = loaded
	return loaded, nil
}

func (g *Graph) dropSources(ids [][]byte) {
	if g == nil {
		return
	}
	for _, id := range ids {
		if len(id) == 16 {
			delete(g.sources, string(id))
		}
	}
}

// WalkSource follows h through one Source's Evidence graph from each Subject
// in from, in memory: Association is the bridge Subject. A bridge Subject is
// one of the hop's bridge type on sg. It cites a from Subject on the hop's
// from-endpoint property and a Subject on sg on its to-endpoint property,
// both positively, and the filter term on the filter property when the hop
// has one. Each edge appears once, ordered by from Subject id, then the
// bridge's ref, then the to Subject's ref. A name the vocabulary does not
// know reaches nothing.
func (g *Graph) WalkSource(sg SourceGraph, h hops.Hop, from [][]byte) ([]SourceEdge, error) {
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
	subjects := make(map[string]SourceSubject, len(sg.Subjects))
	for _, s := range sg.Subjects {
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
	for _, o := range sg.Observations {
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
	var out []SourceEdge
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
			out = append(out, SourceEdge{
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
