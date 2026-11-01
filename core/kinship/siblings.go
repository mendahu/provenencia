package kinship

import (
	"bytes"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// PatternSharedParent is how a sibling is implied through a parent both have.
// It is page-only: Derive never makes it and Patterns does not list it.
const PatternSharedParent Pattern = "shared_parent"

// Parents returns person's distinct parents among edges, read through
// ParentStep, in first-seen order.
func Parents(person []byte, edges []Edge, terms segmentcompare.Terms) [][]byte {
	return indexSteps(stepsOf(edges, terms)).parentsOf[string(person)]
}

// Siblings returns the siblings person's parents imply among edges: one link
// per parent and other child of that parent, through the parent. It returns
// only implied links; LinkSet merges them with a recorded sibling link. It is
// for the page; matching never calls it.
//
// It has no MaxPivotEnd bound: a family of eleven children is real data, and
// the work is linear in the edges.
func Siblings(person []byte, edges []Edge, terms segmentcompare.Terms) []Implied {
	rel, ok := connectrules.LookupBridge(vocabkey.BridgeRelationship)
	if !ok {
		return nil
	}
	first, second := segmentcompare.Keys(rel, vocabkey.TermSibling, "", terms)
	idx := indexSteps(stepsOf(edges, terms))
	var out []Implied
	for _, q := range idx.parentsOf[string(person)] {
		for _, i := range idx.childrenOf[string(q)] {
			c := idx.steps[i].m
			if bytes.Equal(c, person) {
				continue
			}
			from, assoc := idx.through(stepPair{string(q), string(person)}, stepPair{string(q), string(c)})
			a, b := person, c
			if bytes.Compare(a, b) > 0 {
				a, b = b, a
			}
			out = append(out, Implied{
				Edge:    Edge{A: a, B: b, KeyA: first, KeyB: second, Associations: assoc},
				Pattern: PatternSharedParent,
				Through: q,
				From:    from,
			})
		}
	}
	return out
}
