package graphcache

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/propertymeta"
)

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

// SubjectLabels maps each Subject's id (as a string) to its label: the one
// label rule, for an Observation whose value is another Subject.
func SubjectLabels(subjects []EvidenceSubject) map[string]string {
	labels := make(map[string]string, len(subjects))
	for _, s := range subjects {
		labels[string(s.ID)] = s.Label
	}
	return labels
}

// EvidenceObservation is one observation on an Evidence graph. Property and term labels are applied
// when the row is shown. Date and Name are the stored value blobs.
type EvidenceObservation struct {
	ID         []byte
	Ref        string
	CitationID []byte
	SubjectID  []byte
	PropertyID []byte
	Polarity   string
	propertymeta.Scalar
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

// EvidenceGraph is one loaded Evidence graph (graphfill.EvidenceGraph). SourceID is the graph's Source
// (its words live in the Source records, sources.Records), and GraphRef and
// GraphName the graph's own, set by the loader. GraphRef is "" for a graph with
// no ref. A graph whose row does not exist is not Found and has no SourceID, so
// no Source drop removes it.
type EvidenceGraph struct {
	GraphID      []byte
	GraphRef     string
	GraphName    string
	SourceID     []byte
	Found        bool
	Subjects     []EvidenceSubject
	Observations []EvidenceObservation
	Members      []EvidenceMember
	EventPeople  []EvidenceEdge
	EventPlaces  []EvidenceEdge
}

// maxWalkEdges is the most edges one walk builds. A graph over it is
// malformed and the walk reaches nothing. A real graph holds about one edge
// per bridge Subject; the largest walk in the 100-Subject budget fixture builds 60.
const maxWalkEdges = 50_000

// WalkEvidenceGraph follows h through one Evidence graph from each Subject
// in from, in memory: Association is the bridge Subject. A bridge Subject is
// one of the hop's bridge type on eg. It cites a from Subject on the hop's
// from-endpoint property and a Subject on eg on its to-endpoint property,
// both positively, and the filter term on the filter property when the hop
// has one. Each edge appears once, ordered by from Subject id, then the
// bridge's ref, then the to Subject's ref. A name the vocabulary does not
// know reaches nothing. A walk that would build more than maxWalkEdges edges
// reaches nothing, counted before any edge is built.
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
	fromProp, toProp := v.PropertyID(h.From(), h.Origin()), v.PropertyID(h.To(), h.Origin())
	if len(bridgeType) == 0 || len(fromProp) == 0 || len(toProp) == 0 {
		return nil, nil
	}
	var filterProp []byte
	filterTerms := map[string]bool{}
	filterKey, termKey, filtered := h.Filter()
	if filtered {
		if filterProp = v.PropertyID(filterKey, h.Origin()); len(filterProp) == 0 {
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
	total := 0
	for _, r := range reaches {
		if c := bridges[string(r.bridge)]; !filtered || c.filtered {
			total += len(c.tos)
		}
	}
	if total > maxWalkEdges {
		return nil, nil
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
