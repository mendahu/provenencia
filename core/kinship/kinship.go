// Package kinship reads one side's links through the kinship patterns: a
// shared marriage implies spouses, a birth's father or mother and subject
// imply a parent, and a parent of a parent implies a grandparent. Pure: no
// catalog, no I/O, no numbers; the catalog's bridge terms come in as an
// argument. It returns only implied links; callers merge them with direct ones.
package kinship

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Edge is one link between two primary Subjects: its two ends, the link read
// from each end, and the bridge Subjects behind it.
type Edge struct {
	A, B       []byte
	KeyA, KeyB vocabkey.BridgeKey // the link read from A's end and from B's
	// Associations are the bridge Subjects behind the link.
	Associations [][]byte
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
func Derive(edges []Edge, terms segmentcompare.Terms) []Implied {
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
		switch {
		case e.KeyA == parentFirst && e.KeyB == parentSecond:
			steps = appendStep(steps, e.A, e.B, e)
		case e.KeyA == parentSecond && e.KeyB == parentFirst:
			steps = appendStep(steps, e.B, e.A, e)
		}
	}

	var implied []Implied
	for _, p := range eventPatterns {
		implied = append(implied, eventLinks(p, edges, part, rel, terms)...)
	}
	out = append(out, implied...)
	for _, imp := range implied {
		if imp.KeyA == parentFirst && imp.KeyB == parentSecond {
			steps = appendStep(steps, imp.A, imp.B, imp.Edge)
		}
	}
	return append(out, chains(steps, rel, terms)...)
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
func eventLinks(p eventPattern, edges []Edge, part, rel connectrules.Bridge, terms segmentcompare.Terms) []Implied {
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
			var x participant
			switch {
			case e.KeyA == r.person && e.KeyB == r.event:
				x = participant{person: e.A, event: e.B}
			case e.KeyA == r.event && e.KeyB == r.person:
				x = participant{person: e.B, event: e.A}
			default:
				continue
			}
			x.edge, x.atA, x.atB = e, r.atA, r.atB
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
func chains(steps []step, rel connectrules.Bridge, terms segmentcompare.Terms) []Implied {
	first, second := segmentcompare.Keys(rel, vocabkey.TermGrandparent, "", terms)
	type triple struct{ g, m, c string }
	made := map[triple]bool{}
	var out []Implied
	for _, up := range steps {
		for _, down := range steps {
			t := triple{string(up.g), string(up.m), string(down.m)}
			if !bytes.Equal(up.m, down.g) || bytes.Equal(up.g, down.m) || made[t] {
				continue
			}
			made[t] = true
			var from []Edge
			var assoc [][]byte
			for _, s := range steps {
				if bytes.Equal(s.g, up.g) && bytes.Equal(s.m, up.m) {
					from = append(from, s.edge)
				}
			}
			for _, s := range steps {
				if bytes.Equal(s.g, up.m) && bytes.Equal(s.m, down.m) {
					from = append(from, s.edge)
				}
			}
			for _, e := range from {
				for _, a := range e.Associations {
					if !slices.ContainsFunc(assoc, func(b []byte) bool { return bytes.Equal(a, b) }) {
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
