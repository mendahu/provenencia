package promotealign

import (
	"bytes"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// loadTerms reads the catalog's facts about every term of each registered
// bridge's disambiguation Property: whether it keeps direction and which term
// reads it from the other end. It is the one source of both.
func loadTerms(g *graphcache.Graph) (segmentcompare.Terms, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	out := segmentcompare.Terms{}
	for _, b := range connectrules.Bridges() {
		if !connectrules.HasDisambiguation(b.Disambiguation) {
			continue
		}
		typ := vocabkey.SubjectType(b.BridgeTypeKey)
		for _, t := range v.Terms {
			if t.PropertyKey != b.Disambiguation {
				continue
			}
			if out[typ] == nil {
				out[typ] = map[vocabkey.BridgeTerm]segmentcompare.Term{}
			}
			out[typ][vocabkey.BridgeTerm(t.Key)] = segmentcompare.Term{
				Directed: t.Directed, Inverse: vocabkey.BridgeTerm(t.InverseKey),
			}
		}
	}
	return out, nil
}

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
// kinship patterns imply. Shapes are the patterns that imply it; OnlyImplied is
// true when no recorded bridge stands behind it.
type sideLink struct {
	kinship.Edge
	Shapes      []kinship.Shape
	OnlyImplied bool
}

// linkSet is one side's links, one per pair of ends and keys, in the order they
// were first added. It is the one merge for both loaders, recorded links and
// implied ones.
type linkSet struct {
	links []sideLink
	at    map[seenEdge]int // a link's index in links
	// seen holds each link's associations, parallel to links, so a merge does
	// not rescan them.
	seen []map[string]struct{}
}

func newLinkSet() *linkSet { return &linkSet{at: map[seenEdge]int{}} }

// add merges l into the link with the same ends and keys, or appends it. The
// set holds its own copy of every end, association and pivot.
func (s *linkSet) add(l sideLink) {
	k := edgeSeenKey(l.A, l.B, l.KeyA, l.KeyB)
	if i, ok := s.at[k]; ok {
		have := &s.links[i]
		for _, a := range l.Associations {
			if _, dup := s.seen[i][string(a)]; !dup {
				s.seen[i][string(a)] = struct{}{}
				have.Associations = append(have.Associations, bytes.Clone(a))
			}
		}
		have.Shapes = append(have.Shapes, cloneShapes(l.Shapes)...)
		have.OnlyImplied = have.OnlyImplied && l.OnlyImplied
		return
	}
	l.A, l.B = bytes.Clone(l.A), bytes.Clone(l.B)
	assocs := make([][]byte, len(l.Associations))
	have := make(map[string]struct{}, len(l.Associations))
	for i, a := range l.Associations {
		assocs[i] = bytes.Clone(a)
		have[string(a)] = struct{}{}
	}
	s.seen = append(s.seen, have)
	l.Associations = assocs
	l.Shapes = cloneShapes(l.Shapes)
	s.at[k] = len(s.links)
	s.links = append(s.links, l)
}

func cloneShapes(in []kinship.Shape) []kinship.Shape {
	var out []kinship.Shape
	for _, x := range in {
		out = append(out, kinship.Shape{Pattern: x.Pattern, Pivot: bytes.Clone(x.Pivot)})
	}
	return out
}

// withImplied adds the links the kinship patterns imply to the set, through add.
func (s *linkSet) withImplied(terms segmentcompare.Terms, rules kinship.Rules) {
	edges := make([]kinship.Edge, len(s.links))
	for i, l := range s.links {
		edges[i] = l.Edge
	}
	for _, imp := range kinship.Derive(edges, terms, rules) {
		s.add(sideLink{Edge: imp.Edge, Shapes: []kinship.Shape{imp.Shape()}, OnlyImplied: true})
	}
}

// observedTerms maps each Subject to the key of a positive Observation's term
// on the Property, among those the Evidence graph records for the listed
// Subjects. The first key in order wins.
func observedTerms(eg graphcache.EvidenceGraph, props map[string]graphcache.VocabProperty, terms map[string]graphcache.VocabTerm, subjectIDs [][]byte, propertyKey string) map[string]string {
	out := map[string]string{}
	if propertyKey == "" {
		return out
	}
	want := make(map[string]bool, len(subjectIDs))
	for _, id := range subjectIDs {
		want[string(id)] = true
	}
	for _, o := range eg.Observations {
		if o.Polarity != observations.PolarityPositive || !want[string(o.SubjectID)] || len(o.TermID) == 0 {
			continue
		}
		if p, ok := props[string(o.PropertyID)]; !ok || p.Key != propertyKey || p.Origin != vocabkey.OriginProvenencia {
			continue
		}
		t, ok := terms[string(o.TermID)]
		if !ok {
			continue
		}
		if cur, ok := out[string(o.SubjectID)]; !ok || t.Key < cur {
			out[string(o.SubjectID)] = t.Key
		}
	}
	return out
}
