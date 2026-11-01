// Package kinship reads one side's links through the kinship patterns: a
// shared marriage implies spouses, a birth's father or mother and subject
// imply a parent, and a parent of a parent implies a grandparent. Pure: no
// catalog, no I/O, no numbers; the catalog's bridge terms and the Rules come
// in as arguments. It returns only implied links; `LinkSet` merges them with direct ones.
// `Siblings` reads the siblings a shared parent implies, for the page; matching
// never calls it. `KinOf` and `FamilyUnits` read a person's nuclear family for
// the page, from links `LinkSet` has merged; matching never calls them.
package kinship

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Rules is what Derive reads from the registry.
type Rules struct {
	// MaxPivotEnd bounds what one pivot can imply: a pivot (an event, or the
	// parent between in a chain) with more participants than this at either
	// end reads as malformed data and implies nothing. It keeps Derive linear
	// in its input.
	MaxPivotEnd int
	// MaxPairSteps is the most steps kept for one parent-child pair; later
	// duplicates of the same pair are not read. It keeps Siblings and the
	// parent-of-parent chain linear in their input.
	MaxPairSteps int
}

// Edge is one link between two primary Subjects: its two ends, the link read
// from each end, and the bridge Subjects behind it.
type Edge struct {
	A, B       []byte
	KeyA, KeyB vocabkey.BridgeKey // the link read from A's end and from B's
	// Associations are the bridge Subjects behind the link.
	Associations [][]byte
}

// From reads the edge from one end: the Subject at the other end and the link
// as seen from id. ok is false when id is neither end.
func (e Edge) From(id []byte) (other []byte, key vocabkey.BridgeKey, ok bool) {
	switch {
	case bytes.Equal(e.A, id):
		return e.B, e.KeyA, true
	case bytes.Equal(e.B, id):
		return e.A, e.KeyB, true
	}
	return nil, vocabkey.BridgeKey{}, false
}

// Clone returns a copy of e that shares no bytes with it: its ends and every
// association are copied.
func (e Edge) Clone() Edge {
	out := e
	out.A, out.B = bytes.Clone(e.A), bytes.Clone(e.B)
	if e.Associations != nil {
		out.Associations = make([][]byte, len(e.Associations))
		for i, a := range e.Associations {
			out.Associations[i] = bytes.Clone(a)
		}
	}
	return out
}

// reads returns the edge's ends in the order of a key pair, whichever end the
// pair is read from: x is the end read by first, y the end read by second.
func (e Edge) reads(first, second vocabkey.BridgeKey) (x, y []byte, ok bool) {
	switch {
	case e.KeyA == first && e.KeyB == second:
		return e.A, e.B, true
	case e.KeyA == second && e.KeyB == first:
		return e.B, e.A, true
	}
	return nil, nil, false
}

// Pattern is how an implied link was made.
type Pattern string

const (
	PatternMarriageSpouses Pattern = "marriage_spouses"
	PatternBirthParents    Pattern = "birth_parents"
	PatternParentOfParent  Pattern = "parent_of_parent"
)

// Implied is a relationship link a pattern found. A is the relationship's
// person end: the parent or grandparent; for spouses and siblings, the lesser id.
type Implied struct {
	Edge
	Pattern Pattern
	Event   []byte // the marriage or birth; nil for a chain
	Through []byte // the parent between, for a chain; nil otherwise
	From    []Edge // the links it was made from
}

// Shape is one way a link is recorded through a pattern. Pivot is
// Implied.Event for an event pattern and Implied.Through for a chain: the event
// or middle parent the link passes through.
type Shape struct {
	Pattern Pattern
	Pivot   []byte
}

// Clone returns a copy of s that shares no pivot bytes with it.
func (s Shape) Clone() Shape { return Shape{s.Pattern, bytes.Clone(s.Pivot)} }

// Shape is the way the pattern made this link.
func (i Implied) Shape() Shape {
	if i.Event != nil {
		return Shape{i.Pattern, i.Event}
	}
	return Shape{i.Pattern, i.Through}
}

// Recording is how this implied link records itself on a side: its one shape,
// and implied only.
func (i Implied) Recording() Recording {
	return Recording{Shapes: []Shape{i.Shape()}, OnlyImplied: true}
}

// Recording is how one link is recorded on one side: the shapes that implied
// it, and whether they are all it has (it is recorded directly otherwise). The
// zero value is a link recorded only.
type Recording struct {
	Shapes      []Shape
	OnlyImplied bool
}

// Clone returns a copy of r that shares no pivot bytes with it.
func (r Recording) Clone() Recording {
	var out Recording
	out.OnlyImplied = r.OnlyImplied
	for _, s := range r.Shapes {
		out.Shapes = append(out.Shapes, s.Clone())
	}
	return out
}

// Merge folds another recording of the same link into r: it appends each shape
// of o that r lacks (same Pattern and pivot), copying its pivot, and the link
// stays only implied when both were.
func (r *Recording) Merge(o Recording) {
	for _, s := range o.Shapes {
		if !slices.ContainsFunc(r.Shapes, func(x Shape) bool {
			return x.Pattern == s.Pattern && bytes.Equal(x.Pivot, s.Pivot)
		}) {
			r.Shapes = append(r.Shapes, s.Clone())
		}
	}
	r.OnlyImplied = r.OnlyImplied && o.OnlyImplied
}

// CountedOnce reports whether two corresponding links stand for one fact both
// sides record through the same kinship pattern, already counted because the
// shape's pivots are paired: paired maps a pivot of a, the layer side, to its
// partner on b's side. Two links that are both recorded never are.
func CountedOnce(a, b Recording, paired func(pivot []byte) ([]byte, bool)) bool {
	if !a.OnlyImplied && !b.OnlyImplied {
		return false
	}
	for _, x := range a.Shapes {
		partner, ok := paired(x.Pivot)
		if !ok {
			continue
		}
		for _, y := range b.Shapes {
			if x.Pattern == y.Pattern && bytes.Equal(partner, y.Pivot) {
				return true
			}
		}
	}
	return false
}

// Derive returns the relationship links the edges imply through the event
// patterns and the parent-of-parent chain, event patterns first. An edge is
// recognized by its keys alone, from either end. It returns only implied
// links, never an input edge, and does not modify edges; the returned slices
// share the input's byte slices. It is nil when nothing is implied.
func Derive(edges []Edge, terms segmentcompare.Terms, rules Rules) []Implied {
	part, ok := connectrules.LookupBridge(vocabkey.BridgeParticipation)
	if !ok {
		return nil
	}

	var out []Implied
	steps := stepsOf(edges, terms)

	var implied []Implied
	for _, p := range eventPatterns {
		implied = append(implied, eventLinks(p, edges, part, terms, rules.MaxPivotEnd)...)
	}
	out = append(out, implied...)
	impliedEdges := make([]Edge, len(implied))
	for i, imp := range implied {
		impliedEdges[i] = imp.Edge
	}
	steps = append(steps, stepsOf(impliedEdges, terms)...)
	return append(out, chains(indexSteps(steps, rules.MaxPairSteps), terms, rules.MaxPivotEnd)...)
}

// step is one parent step: g is m's parent, and edge is the link it was read
// from.
type step struct {
	g, m []byte
	edge Edge
}

// stepsOf reads each edge that is a parent step, in input order.
func stepsOf(edges []Edge, terms segmentcompare.Terms) []step {
	var steps []step
	for _, e := range edges {
		if g, m, ok := ParentStep(e, terms); ok {
			steps = append(steps, step{g, m, e})
		}
	}
	return steps
}

// stepPair is a parent and a child, as the index keys their steps.
type stepPair struct{ g, m string }

// parentSteps indexes parent steps by parent and child. Duplicate steps (the
// same pair read from several links) never pair with each other, so a loop over
// the distinct pairs is linear.
type parentSteps struct {
	steps []step
	// byPair holds the step indices of each parent and child; order the
	// distinct pairs in first-seen order.
	byPair map[stepPair][]int
	order  []stepPair
	// childrenOf holds, for each parent, the index of the first step of each of
	// its distinct children, first-seen; parentsOf the distinct parents of each
	// child, first-seen.
	childrenOf map[string][]int
	parentsOf  map[string][][]byte
}

// indexSteps builds the index of steps, keeping the first maxPairSteps steps of
// each parent and child pair, in input order; the pair itself is always indexed.
func indexSteps(steps []step, maxPairSteps int) parentSteps {
	idx := parentSteps{
		steps:      steps,
		byPair:     map[stepPair][]int{},
		childrenOf: map[string][]int{},
		parentsOf:  map[string][][]byte{},
	}
	for i, s := range steps {
		k := stepPair{string(s.g), string(s.m)}
		kept, seen := idx.byPair[k]
		if !seen {
			idx.order = append(idx.order, k)
			idx.childrenOf[k.g] = append(idx.childrenOf[k.g], i)
			idx.parentsOf[k.m] = append(idx.parentsOf[k.m], s.g)
		}
		if len(kept) < max(maxPairSteps, 1) {
			kept = append(kept, i)
		}
		idx.byPair[k] = kept
	}
	return idx
}

// edgesOf returns the edges of the kept steps of each pair, in argument order.
func (s parentSteps) edgesOf(pairs ...stepPair) []Edge {
	var from []Edge
	for _, p := range pairs {
		for _, j := range s.byPair[p] {
			from = append(from, s.steps[j].edge)
		}
	}
	return from
}

// associationsOf returns the associations of edges, each once, first-seen.
func associationsOf(edges []Edge) [][]byte {
	var assoc [][]byte
	seen := map[string]struct{}{}
	for _, e := range edges {
		for _, a := range e.Associations {
			if _, dup := seen[string(a)]; !dup {
				seen[string(a)] = struct{}{}
				assoc = append(assoc, a)
			}
		}
	}
	return assoc
}

// ParentStep reads an edge as one parent step: parent is child's parent, from
// whichever end the link is read. It is false for any other link, and for a
// link that joins a Subject to itself. It is the one rule for what a parent
// step is, shared by Derive and the readers of its inputs.
func ParentStep(e Edge, terms segmentcompare.Terms) (parent, child []byte, ok bool) {
	return readsTerm(e, vocabkey.TermParent, terms)
}

// participant is one person's participation edge in an event, and the ends of
// the pattern it can stand at.
type participant struct {
	person, event []byte
	edge          Edge
	atA, atB      bool
}

// eventLinks makes the links one event pattern implies, one per event and
// pair of participation edges.
func eventLinks(p eventPattern, edges []Edge, part connectrules.Bridge, terms segmentcompare.Terms, maxPivotEnd int) []Implied {
	type roleKeys struct {
		person, event vocabkey.BridgeKey
		atA, atB      bool
	}
	var roles []roleKeys
	for _, role := range slices.Concat(p.a, p.b) {
		personKey, eventKey := segmentcompare.Keys(part, vocabkey.BridgeTerm(role), p.event, terms)
		roles = append(roles, roleKeys{personKey, eventKey, slices.Contains(p.a, role), slices.Contains(p.b, role)})
	}
	// Participants grouped by event, events and participants in input order.
	var order [][]byte
	byEvent := map[string][]participant{}
	for _, e := range edges {
		for _, r := range roles {
			person, event, ok := e.reads(r.person, r.event)
			if !ok {
				continue
			}
			x := participant{person: person, event: event, edge: e, atA: r.atA, atB: r.atB}
			if _, seen := byEvent[string(x.event)]; !seen {
				order = append(order, x.event)
			}
			byEvent[string(x.event)] = append(byEvent[string(x.event)], x)
			break
		}
	}

	first, second, ok := relationshipKeys(p.term, terms)
	if !ok {
		return nil
	}
	symmetric := first == second
	var out []Implied
	for _, event := range order {
		group := byEvent[string(event)]
		var atA, atB int
		for _, x := range group {
			if x.atA {
				atA++
			}
			if x.atB {
				atB++
			}
		}
		if atA > maxPivotEnd || atB > maxPivotEnd {
			continue
		}
		for _, x := range group {
			for _, y := range group {
				if !x.atA || !y.atB || bytes.Equal(x.person, y.person) {
					continue
				}
				if symmetric && bytes.Compare(x.person, y.person) > 0 {
					continue
				}
				out = append(out, Implied{
					Edge: Edge{
						A: x.person, B: y.person, KeyA: first, KeyB: second,
						Associations: slices.Concat(x.edge.Associations, y.edge.Associations),
					},
					Pattern: p.code,
					Event:   event,
					From:    []Edge{x.edge, y.edge},
				})
			}
		}
	}
	return out
}

// chains makes one grandparent link per grandparent, middle and grandchild.
func chains(idx parentSteps, terms segmentcompare.Terms, maxPivotEnd int) []Implied {
	first, second, ok := relationshipKeys(vocabkey.TermGrandparent, terms)
	if !ok {
		return nil
	}
	var out []Implied
	for _, up := range idx.order {
		// A middle person with too many parents implies nothing.
		if len(idx.parentsOf[up.m]) > maxPivotEnd {
			continue
		}
		upStep := idx.steps[idx.byPair[up][0]]
		for _, i := range idx.childrenOf[up.m] {
			down := idx.steps[i]
			if bytes.Equal(upStep.g, down.m) {
				continue
			}
			from := idx.edgesOf(up, stepPair{up.m, string(down.m)})
			assoc := associationsOf(from)
			out = append(out, Implied{
				Edge:    Edge{A: upStep.g, B: down.m, KeyA: first, KeyB: second, Associations: assoc},
				Pattern: PatternParentOfParent,
				Through: upStep.m,
				From:    from,
			})
		}
	}
	return out
}
