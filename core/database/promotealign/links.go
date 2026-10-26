package promotealign

import (
	"bytes"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// seenEdge identifies one edge, whichever end listed it, by its ends and the
// key it reads from each.
type seenEdge struct {
	a, b       string
	keyA, keyB vocabkey.BridgeKey
}

// edgeSeenKey is the edge a and b share, with the ends in a fixed order.
func edgeSeenKey(a, b []byte, keyA, keyB vocabkey.BridgeKey) seenEdge {
	if string(a) > string(b) {
		a, b, keyA, keyB = b, a, keyB, keyA
	}
	return seenEdge{a: string(a), b: string(b), keyA: keyA, keyB: keyB}
}

// sideLink is one link on one side of a proposal: a recorded one, or one the
// kinship patterns imply, with how it is recorded.
type sideLink struct {
	kinship.Edge
	kinship.Recording
}

// linkSet is one side's links, one per pair of ends and keys, in the order they
// were first added. It is the one merge for both loaders, recorded links and
// implied ones.
type linkSet struct {
	links []sideLink
	at    map[seenEdge]int // a link's index in links
	// assocs holds each link's associations, parallel to links, so a merge does
	// not rescan them.
	assocs []map[string]struct{}
}

func newLinkSet() *linkSet { return &linkSet{at: map[seenEdge]int{}} }

// add merges l into the link with the same ends and keys, or appends it. The
// set holds its own copy of every end, association and pivot.
func (s *linkSet) add(l sideLink) {
	k := edgeSeenKey(l.A, l.B, l.KeyA, l.KeyB)
	if i, ok := s.at[k]; ok {
		have := &s.links[i]
		for _, a := range l.Associations {
			if _, dup := s.assocs[i][string(a)]; !dup {
				s.assocs[i][string(a)] = struct{}{}
				have.Associations = append(have.Associations, bytes.Clone(a))
			}
		}
		have.Recording.Merge(l.Recording)
		return
	}
	l.A, l.B = bytes.Clone(l.A), bytes.Clone(l.B)
	assocs := make([][]byte, len(l.Associations))
	have := make(map[string]struct{}, len(l.Associations))
	for i, a := range l.Associations {
		assocs[i] = bytes.Clone(a)
		have[string(a)] = struct{}{}
	}
	l.Associations = assocs
	var shapes []kinship.Shape
	for _, x := range l.Shapes {
		shapes = append(shapes, kinship.Shape{Pattern: x.Pattern, Pivot: bytes.Clone(x.Pivot)})
	}
	l.Shapes = shapes
	s.at[k] = len(s.links)
	s.links = append(s.links, l)
	s.assocs = append(s.assocs, have)
}

// addImplied adds a link a kinship pattern implied: the one place implied links
// enter a set.
func (s *linkSet) addImplied(imp kinship.Implied) {
	s.add(sideLink{Edge: imp.Edge, Recording: kinship.Recording{
		Shapes: []kinship.Shape{imp.Shape()}, OnlyImplied: true,
	}})
}

// withImplied adds the links the kinship patterns imply over the set's links.
func (s *linkSet) withImplied(terms segmentcompare.Terms, rules kinship.Rules) {
	edges := make([]kinship.Edge, len(s.links))
	for i, l := range s.links {
		edges[i] = l.Edge
	}
	for _, imp := range kinship.Derive(edges, terms, rules) {
		s.addImplied(imp)
	}
}
