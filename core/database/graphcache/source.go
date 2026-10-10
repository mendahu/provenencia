package graphcache

import (
	"fmt"
)

// SourceLoader fills one Source. promotealign registers it. The graph does
// not import that package.
type SourceLoader func(q Querier, sourceID []byte) (SourceGraph, error)

// SourceCheck reloads one stored Source and compares the membership pointers
// and bridge signatures. promotealign registers it.
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
	Date           []byte
	Name           []byte
}

// SourceMember is an accepted claim: the subject, the claim, and the handle.
// The handle's name and label are read from its node.
type SourceMember struct {
	SubjectID []byte
	ClaimID   []byte
	EntityID  []byte
	Kind      string
}

// SourceBridge is one evidence bridge between two primary subjects. The
// align layer builds an edge signature from these fields.
type SourceBridge struct {
	A, B             []byte
	BridgeType       string
	Term             string
	NeighborKind     string
	NeighborTypeTerm string
	Directed         bool
}

// SourceEdge is one hop on the Source, kept so an event title can be chosen
// from the stored observations.
type SourceEdge struct {
	From  []byte
	To    []byte
	ToRef string
}

// SourceGraph is one loaded Source.
type SourceGraph struct {
	SourceID     []byte
	Subjects     []SourceSubject
	Observations []SourceObservation
	Members      []SourceMember
	Bridges      []SourceBridge
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
