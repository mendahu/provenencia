// Package recordlinks is the record-links comparison: given two Evidence
// graphs of one person, it counts what the two records share about each other
// person both hold, and weighs that line weak or strong. It judges two members
// the researcher already joined to one canonical entity; it is not a layer
// under graph alignment and never reads its scoring, weights, points or band.
// Properties compare through subjectcompare.CompareProperty and links through
// segmentcompare.Corresponds, with kinship.SameShape keeping a fact recorded
// through a pivot one fact. Pure: no catalog, no I/O, no numbers; the rules
// come in as arguments.
package recordlinks

import (
	"bytes"
	"maps"
	"slices"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Rules is the match registry's record-links section.
type Rules struct {
	// StrongAt is the fewest facts two records share about a person for their
	// line to be strong; fewer is weak.
	StrongAt int
}

// Config is the registry as the record-links comparison reads it.
type Config struct {
	Value       valuecompare.Rules
	Subject     subjectcompare.Rules
	RecordLinks Rules
}

// Member is one Subject of an Evidence graph.
type Member struct {
	ID []byte
	// Entity is the canonical entity of the member's accepted Identity Claim,
	// nil when it has none; the caller sets it.
	Entity []byte
	Kind   vocabkey.SubjectType
	Values subjectcompare.Values
}

// Graph is one Evidence graph indexed once for comparison. The zero Graph is
// valid and holds nothing.
type Graph struct {
	// entityOf is each member's entity by member id, nil when it has none.
	entityOf map[string][]byte
	// persons holds the person members that have an entity, by entity.
	persons map[string][]Member
	// links holds each member's links read from its end, by member id and then
	// by the far end's entity.
	links map[string]map[string][]memberLink
}

// memberLink is one link read from a member's end.
type memberLink struct {
	key vocabkey.BridgeKey
	rec kinship.Recording
}

// NewGraph indexes a graph's members and links, links being its
// kinship.LinkSet's Links after WithImplied. A later member with an id already
// seen is ignored; a link with neither end a member is ignored; a link whose
// far end is no member, or has no entity, never counts. It keeps references to
// its inputs and changes neither.
func NewGraph(members []Member, links []kinship.Link) Graph {
	g := Graph{
		entityOf: map[string][]byte{},
		persons:  map[string][]Member{},
		links:    map[string]map[string][]memberLink{},
	}
	for _, m := range members {
		if _, dup := g.entityOf[string(m.ID)]; dup {
			continue
		}
		g.entityOf[string(m.ID)] = m.Entity
		if m.Kind == vocabkey.Person && m.Entity != nil {
			g.persons[string(m.Entity)] = append(g.persons[string(m.Entity)], m)
		}
	}
	for _, l := range links {
		g.addLink(l, l.A)
		if !bytes.Equal(l.A, l.B) {
			g.addLink(l, l.B)
		}
	}
	return g
}

// addLink files l under the member at one end, by the far end's entity.
func (g Graph) addLink(l kinship.Link, end []byte) {
	if _, ok := g.entityOf[string(end)]; !ok {
		return
	}
	far, key, _ := l.Edge.From(end)
	e := g.entityOf[string(far)]
	if e == nil {
		return
	}
	byEntity := g.links[string(end)]
	if byEntity == nil {
		byEntity = map[string][]memberLink{}
		g.links[string(end)] = byEntity
	}
	byEntity[string(e)] = append(byEntity[string(e)], memberLink{key, l.Recording})
}

// Strength is how heavily a line weighs.
type Strength string

const (
	StrengthWeak   Strength = "weak"
	StrengthStrong Strength = "strong"
)

// Line is one person both graphs hold: the facts the two records share about
// that person and the line's weight.
type Line struct {
	Entity   []byte
	Facts    int
	Strength Strength
}

// Lines is one line for each person both graphs hold on one canonical entity,
// other than the page's own: the facts the two records share about that
// person and the line's weight. Lines are sorted by entity.
func Lines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line {
	var out []Line
	for _, e := range slices.Sorted(maps.Keys(a.persons)) {
		if e == string(page) || len(b.persons[e]) == 0 {
			continue
		}
		var best int
		for _, x := range a.persons[e] {
			for _, y := range b.persons[e] {
				best = max(best, len(properties(x, y, metas, cfg))+len(links(x, y, a, b)))
			}
		}
		s := StrengthWeak
		if best >= cfg.RecordLinks.StrongAt {
			s = StrengthStrong
		}
		out = append(out, Line{Entity: bytes.Clone([]byte(e)), Facts: best, Strength: s})
	}
	return out
}

// properties returns the keys, ascending, whose values the two members agree
// or partly agree on. A property counts once under its key, with its pooled
// properties.
func properties(x, y Member, metas subjectcompare.Metas, cfg Config) []vocabkey.PropertyKey {
	keys := map[vocabkey.PropertyKey]struct{}{}
	for _, v := range []subjectcompare.Values{x.Values, y.Values} {
		for p := range v {
			if key, ok := cfg.Subject.CountsUnder(vocabkey.Person, p); ok {
				keys[key] = struct{}{}
			}
		}
	}
	var out []vocabkey.PropertyKey
	for _, key := range slices.Sorted(maps.Keys(keys)) {
		res := subjectcompare.CompareProperty(vocabkey.Person, key, x.Values, y.Values, metas, cfg.Subject, cfg.Value)
		if res.Outcome.Matches() {
			out = append(out, key)
		}
	}
	return out
}

// linkFact is one counted link: the link as the member reads it, and the
// entity at its far end.
type linkFact struct {
	key    vocabkey.BridgeKey
	entity string
}

// links returns the link facts the two members share. For each far entity both
// have links to, a pair of links that correspond counts once when any such pair
// is not skipped; a pair is skipped when kinship.SameShape matches their shapes
// at pivots filed on one entity in both graphs, because that fact counts
// through the pivot.
func links(x, y Member, a, b Graph) []linkFact {
	same := func(pa, pb []byte) bool {
		ea, eb := a.entityOf[string(pa)], b.entityOf[string(pb)]
		return ea != nil && eb != nil && bytes.Equal(ea, eb)
	}
	byX, byY := a.links[string(x.ID)], b.links[string(y.ID)]
	var out []linkFact
	for _, e := range slices.Sorted(maps.Keys(byX)) {
		seen := map[vocabkey.BridgeKey]bool{}
		for _, lx := range byX[e] {
			for _, ly := range byY[e] {
				if seen[lx.key] || !segmentcompare.Corresponds(lx.key, ly.key) || kinship.SameShape(lx.rec, ly.rec, same) {
					continue
				}
				seen[lx.key] = true
				out = append(out, linkFact{lx.key, e})
			}
		}
	}
	return out
}
