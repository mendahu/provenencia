// Package kinship reads one side's links through the kinship patterns: a
// shared marriage implies spouses, a birth's father or mother and subject
// imply a parent, and a parent of a parent implies a grandparent. Pure: no
// catalog, no I/O, no numbers; the catalog's bridge terms and the Rules come
// in as arguments. It returns only implied links; callers merge them with direct ones.
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
}

// Edge is one link between two primary Subjects: its two ends, the link read
// from each end, and the bridge Subjects behind it.
type Edge struct {
	A, B       []byte
	KeyA, KeyB vocabkey.BridgeKey // the link read from A's end and from B's
	// Associations are the bridge Subjects behind the link.
	Associations [][]byte
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
// person end: the parent or grandparent; for spouses, the lesser id.
type Implied struct {
	Edge
	Pattern Pattern
	Event   []byte // the marriage or birth; nil for a chain
	Through []byte // the parent between, for a chain; nil otherwise
	From    []Edge // the links it was made from
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
	rel, ok := connectrules.LookupBridge(vocabkey.BridgeRelationship)
	if !ok {
		return nil
	}
	parentFirst, parentSecond := segmentcompare.Keys(rel, vocabkey.TermParent, "", terms)

	var out []Implied
	var steps []step
	for _, e := range edges {
		if g, m, ok := e.reads(parentFirst, parentSecond); ok {
			steps = appendStep(steps, g, m, e)
		}
	}

	var implied []Implied
	for _, p := range eventPatterns {
		implied = append(implied, eventLinks(p, edges, part, rel, terms, rules.MaxPivotEnd)...)
	}
	out = append(out, implied...)
	for _, imp := range implied {
		if g, m, ok := imp.reads(parentFirst, parentSecond); ok {
			steps = appendStep(steps, g, m, imp.Edge)
		}
	}
	return append(out, chains(steps, rel, terms, rules.MaxPivotEnd)...)
}

// step is one parent step: g is m's parent, and edge is the link it was read
// from.
type step struct {
	g, m []byte
	edge Edge
}

// appendStep adds a parent step unless it joins a Subject to itself.
func appendStep(steps []step, g, m []byte, e Edge) []step {
	if bytes.Equal(g, m) {
		return steps
	}
	return append(steps, step{g, m, e})
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
func eventLinks(p eventPattern, edges []Edge, part, rel connectrules.Bridge, terms segmentcompare.Terms, maxPivotEnd int) []Implied {
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

	first, second := segmentcompare.Keys(rel, p.term, "", terms)
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
func chains(steps []step, rel connectrules.Bridge, terms segmentcompare.Terms, maxPivotEnd int) []Implied {
	first, second := segmentcompare.Keys(rel, vocabkey.TermGrandparent, "", terms)
	type triple struct{ g, m, c string }
	type pair struct{ g, m string }
	// Step indices by parent, and by parent and child, in step order.
	byG := map[string][]int{}
	byPair := map[pair][]int{}
	for i, s := range steps {
		byG[string(s.g)] = append(byG[string(s.g)], i)
		k := pair{string(s.g), string(s.m)}
		byPair[k] = append(byPair[k], i)
	}
	// Distinct parents of each middle person; one with too many implies nothing.
	parents := map[string]map[string]struct{}{}
	for _, s := range steps {
		if parents[string(s.m)] == nil {
			parents[string(s.m)] = map[string]struct{}{}
		}
		parents[string(s.m)][string(s.g)] = struct{}{}
	}
	made := map[triple]bool{}
	var out []Implied
	for _, up := range steps {
		if len(parents[string(up.m)]) > maxPivotEnd {
			continue
		}
		for _, i := range byG[string(up.m)] {
			down := steps[i]
			if bytes.Equal(up.g, down.m) {
				continue
			}
			t := triple{string(up.g), string(up.m), string(down.m)}
			if made[t] {
				continue
			}
			made[t] = true
			var from []Edge
			var assoc [][]byte
			seen := map[string]struct{}{}
			for _, j := range byPair[pair{t.g, t.m}] {
				from = append(from, steps[j].edge)
			}
			for _, j := range byPair[pair{t.m, t.c}] {
				from = append(from, steps[j].edge)
			}
			for _, e := range from {
				for _, a := range e.Associations {
					if _, dup := seen[string(a)]; !dup {
						seen[string(a)] = struct{}{}
						assoc = append(assoc, a)
					}
				}
			}
			out = append(out, Implied{
				Edge:    Edge{A: up.g, B: down.m, KeyA: first, KeyB: second, Associations: assoc},
				Pattern: PatternParentOfParent,
				Through: up.m,
				From:    from,
			})
		}
	}
	return out
}
