package kinship

import (
	"bytes"

	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// linkKey identifies one edge, whichever end listed it, by its ends and the
// key it reads from each.
type linkKey struct {
	a, b       string
	keyA, keyB vocabkey.BridgeKey
}

// linkKeyOf is the key of the edge a and b share, with the ends in a fixed order.
func linkKeyOf(a, b []byte, keyA, keyB vocabkey.BridgeKey) linkKey {
	if string(a) > string(b) {
		a, b, keyA, keyB = b, a, keyB, keyA
	}
	return linkKey{a: string(a), b: string(b), keyA: keyA, keyB: keyB}
}

// Link is one link on one side: a recorded one, or one the kinship patterns
// imply, with how it is recorded.
type Link struct {
	Edge
	Recording
}

// LinkSet is one side's links, one per pair of ends and keys, in the order they
// were first added. It is the one merge for every reader of a side's links,
// recorded links and implied ones.
type LinkSet struct {
	links []Link
	at    map[linkKey]int // a link's index in links
	// assocs holds each link's associations, parallel to links, so a merge does
	// not rescan them. A link with none (a canon link) has a nil set.
	assocs []map[string]struct{}
}

// NewLinkSet is an empty set.
func NewLinkSet() *LinkSet { return &LinkSet{at: map[linkKey]int{}} }

// Links is the set's links in first-added order. The slice is the set's own:
// read-only for callers. It is nil for an empty set.
func (s *LinkSet) Links() []Link { return s.links }

// Add merges l into the link with the same ends and keys, or appends it. The
// set holds its own copy of every end, association and pivot.
func (s *LinkSet) Add(l Link) {
	k := linkKeyOf(l.A, l.B, l.KeyA, l.KeyB)
	if i, ok := s.at[k]; ok {
		have := &s.links[i]
		for _, a := range l.Associations {
			if s.assocs[i] == nil {
				s.assocs[i] = map[string]struct{}{}
			}
			if _, dup := s.assocs[i][string(a)]; !dup {
				s.assocs[i][string(a)] = struct{}{}
				have.Associations = append(have.Associations, bytes.Clone(a))
			}
		}
		have.Recording.Merge(l.Recording)
		return
	}
	l.A, l.B = bytes.Clone(l.A), bytes.Clone(l.B)
	var have map[string]struct{}
	if len(l.Associations) > 0 {
		assocs := make([][]byte, len(l.Associations))
		have = make(map[string]struct{}, len(l.Associations))
		for i, a := range l.Associations {
			assocs[i] = bytes.Clone(a)
			have[string(a)] = struct{}{}
		}
		l.Associations = assocs
	}
	l.Recording = l.Recording.Clone()
	s.at[k] = len(s.links)
	s.links = append(s.links, l)
	s.assocs = append(s.assocs, have)
}

// AddImplied adds a link a kinship pattern implied: the one place implied links
// enter a set.
func (s *LinkSet) AddImplied(imp Implied) {
	s.Add(Link{Edge: imp.Edge, Recording: imp.Recording()})
}

// WithImplied adds the links the kinship patterns imply over the set's links.
func (s *LinkSet) WithImplied(terms segmentcompare.Terms, rules Rules) {
	edges := make([]Edge, len(s.links))
	for i, l := range s.links {
		edges[i] = l.Edge
	}
	for _, imp := range Derive(edges, terms, rules) {
		s.AddImplied(imp)
	}
}

// PlaceImplied returns where each implied link goes among s's links: the
// index of the link it merges into, or len(s.Links())+j for the j-th distinct
// new link. It does not change s.
func (s *LinkSet) PlaceImplied(implied []Implied) []int {
	at := make([]int, len(implied))
	var fresh map[linkKey]int
	for i, imp := range implied {
		k := linkKeyOf(imp.A, imp.B, imp.KeyA, imp.KeyB)
		if j, ok := s.at[k]; ok {
			at[i] = j
			continue
		}
		j, ok := fresh[k]
		if !ok {
			if fresh == nil {
				fresh = map[linkKey]int{}
			}
			j = len(s.links) + len(fresh)
			fresh[k] = j
		}
		at[i] = j
	}
	return at
}
