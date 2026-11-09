// Package recordlinks is the record-links comparison: given the Evidence
// graphs of one person, Pairs counts, for each pair of them, what the two
// records share about each other person both hold, and weighs that line weak
// or strong. It judges two members
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
	// MaxEntityMembers is the most members of any kind either graph may file on
	// one canonical entity before the comparison reads it as malformed: a
	// person over it keeps its line at 0 facts, and links to an entity over it
	// count nothing. It bounds the member pairs and link pairs one pair of graphs
	// compares.
	MaxEntityMembers int
	// MaxUnitValues is the most values either member may hold for one property
	// that counts, its pooled properties included, before that property counts
	// nothing. It bounds the value pairs property compare reads.
	MaxUnitValues int
	// MaxPairWork is the most work one pair of graphs may do before it reads
	// the pair as malformed and compares nothing: every shared person
	// keeps a line at 0 facts. Work is counted before any compare: for each
	// member pair, every link and every counting property key either member
	// holds (a link to an entity the other graph lacks included), then the
	// words of one side times the words of the other for each property that
	// counts, and for each entity both link to, the shapes of one side's links
	// times the other's, a link with no shape counting one.
	MaxPairWork int
	// MaxValueBytes is the most bytes one value may hold (valuecompare.Size)
	// before the property holding it counts nothing for the member pair: it
	// is not compared and charges no work. It bounds what one value costs to
	// read, so a crafted value cannot hang a pair's count.
	MaxValueBytes int
	// MaxReadWork is the most work one Pairs call may do before it compares
	// nothing: every line keeps 0 facts, and when the number of lines alone
	// passes it, no line is listed. Work is the number of lines plus each
	// pair's work as MaxPairWork counts it, capped at MaxPairWork.
	MaxReadWork int
}

// Config is the registry as the record-links comparison reads it.
type Config struct {
	Value       valuecompare.Rules
	Subject     subjectcompare.Rules
	RecordLinks Rules
}

// Member is one Subject of an Evidence graph.
type Member struct {
	subjectcompare.Subject
	// Entity is the canonical entity of the member's accepted Identity Claim,
	// nil when it has none; the caller sets it.
	Entity []byte
}

// Graph is one Evidence graph indexed once for comparison. The zero Graph is
// valid and holds nothing.
type Graph struct {
	// entityOf is each member's entity by member id, nil when it has none.
	entityOf map[string][]byte
	// persons holds the person members that have an entity, by entity.
	persons map[string][]Member
	// members counts the members of any kind that have an entity, by entity.
	members map[string]int
	// links holds each member's links read from its end, by member id and then
	// by the far end's entity.
	links map[string]map[string][]memberLink
	// linkCount is the number of links each member holds, by member id: every
	// memberLink under every far entity.
	linkCount map[string]int
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
		entityOf:  map[string][]byte{},
		persons:   map[string][]Member{},
		members:   map[string]int{},
		links:     map[string]map[string][]memberLink{},
		linkCount: map[string]int{},
	}
	for _, m := range members {
		if _, dup := g.entityOf[string(m.ID)]; dup {
			continue
		}
		g.entityOf[string(m.ID)] = m.Entity
		if m.Entity != nil {
			g.members[string(m.Entity)]++
		}
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
	g.linkCount[string(end)]++
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

// Pair is the lines between two graphs of a read; A < B index the graphs
// Pairs was given.
type Pair struct {
	A, B  int
	Lines []Line
}

// Pairs is the lines between every pair of graphs that share a person other
// than the page's own, ordered by A and then B; a pair with no line is absent.
// Each pair's lines are what the one-pair count gives over the two graphs. The
// whole read is bounded by MaxReadWork, counted before any compare: past it
// every pair is listed and every line has 0 facts, and when the number of lines
// alone passes it, no pair is returned. counted is false exactly then, when the
// read's work passes MaxReadWork, and true otherwise, a read of no graph
// included. A pair over MaxPairWork or an entity over MaxEntityMembers keeps
// its lines at 0 facts and does not change counted.
func Pairs(page []byte, graphs []Graph, metas subjectcompare.Metas, cfg Config) (pairs []Pair, counted bool) {
	h := holdings{}
	limit := cfg.RecordLinks.MaxReadWork
	shared, work := readPlan(string(page), graphs, h, cfg, limit)
	if shared == nil {
		return nil, false
	}
	skip := work > limit
	var out []Pair
	for _, p := range shared {
		out = append(out, Pair{A: p.a, B: p.b, Lines: linesOf(p.entities, graphs[p.a], graphs[p.b], metas, cfg, h, h, skip || p.over)})
	}
	return out, !skip
}

// CapLines keeps at most most lines (most is at least 1) of pairs. With at
// most most lines in all it returns pairs as they are and false. Otherwise it
// keeps the first most strong lines in read order (pair order, then line
// order), then, while fewer than most are kept, the first weak lines in read
// order; each kept line stays in its pair at its place, kept pairs keep their
// order, a pair left with no line is dropped, and it returns true. It builds
// new Lines slices and does not modify its input.
func CapLines(pairs []Pair, most int) ([]Pair, bool) {
	total, strong := 0, 0
	for _, p := range pairs {
		total += len(p.Lines)
		for _, l := range p.Lines {
			if l.Strength == StrengthStrong {
				strong++
			}
		}
	}
	if total <= most {
		return pairs, false
	}
	strongLeft := min(strong, most)
	weakLeft := most - strongLeft
	var out []Pair
	for _, p := range pairs {
		var lines []Line
		for _, l := range p.Lines {
			switch {
			case l.Strength == StrengthStrong && strongLeft > 0:
				strongLeft--
			case l.Strength != StrengthStrong && weakLeft > 0:
				weakLeft--
			default:
				continue
			}
			lines = append(lines, l)
		}
		if len(lines) > 0 {
			out = append(out, Pair{A: p.A, B: p.B, Lines: lines})
		}
	}
	return out, true
}

// pairKey is the indexes of two graphs, lower first.
type pairKey struct{ a, b int }

// sharedPair is the person entities, ascending, that graphs a < b both hold.
type sharedPair struct {
	a, b     int
	entities []string
	// over reports that the pair's work passes MaxPairWork.
	over bool
}

// readPlan indexes the persons of graphs other than the page and returns the
// pairs that share one, ordered, with the read's work: the number of lines plus
// each pair's work capped at MaxPairWork. It counts the lines before it builds
// any pair and stops counting once past limit; past it, the work is returned
// as counted so far, and the pairs are nil when the lines alone passed limit.
func readPlan(page string, graphs []Graph, h holdings, cfg Config, limit int) ([]sharedPair, int) {
	holders := map[string][]int{}
	for i, g := range graphs {
		for e := range g.persons {
			if e != page {
				holders[e] = append(holders[e], i)
			}
		}
	}
	entities := slices.Sorted(maps.Keys(holders))
	var n int
	for _, e := range entities {
		for i := range holders[e] {
			if n += i; n > limit {
				return nil, n
			}
		}
	}
	byPair := map[pairKey]*sharedPair{}
	for _, e := range entities {
		hs := holders[e]
		for j, y := range hs {
			for _, x := range hs[:j] {
				p := byPair[pairKey{x, y}]
				if p == nil {
					p = &sharedPair{a: x, b: y}
					byPair[pairKey{x, y}] = p
				}
				p.entities = append(p.entities, e)
			}
		}
	}
	out := make([]sharedPair, 0, len(byPair))
	for _, p := range byPair {
		out = append(out, *p)
	}
	slices.SortFunc(out, func(p, q sharedPair) int {
		if p.a != q.a {
			return p.a - q.a
		}
		return p.b - q.b
	})
	for i := range out {
		p := &out[i]
		w := pairWork(p.entities, graphs[p.a], graphs[p.b], h, h, cfg, cfg.RecordLinks.MaxPairWork)
		p.over = w > cfg.RecordLinks.MaxPairWork
		if n += min(w, cfg.RecordLinks.MaxPairWork); n > limit {
			return out, n
		}
	}
	return out, n
}

// sharedOf is the person entities, ascending, that a and b both hold, the
// page's own excluded.
func sharedOf(page []byte, a, b Graph) []string {
	var out []string
	for _, e := range slices.Sorted(maps.Keys(a.persons)) {
		if e != string(page) && len(b.persons[e]) != 0 {
			out = append(out, e)
		}
	}
	return out
}

// pairLines is one line for each person both graphs hold on one canonical
// entity, other than the page's own: the facts the two records share about
// that person and the line's weight. Lines are sorted by entity. A pair of
// graphs whose work passes MaxPairWork compares nothing: every line has 0
// facts.
func pairLines(page []byte, a, b Graph, metas subjectcompare.Metas, cfg Config) []Line {
	shared := sharedOf(page, a, b)
	return linesOf(shared, a, b, metas, cfg, holdings{}, holdings{}, overWork(shared, a, b, holdings{}, holdings{}, cfg))
}

// linesOf is the lines for the shared entities. capped compares nothing.
func linesOf(shared []string, a, b Graph, metas subjectcompare.Metas, cfg Config, ha, hb holdings, capped bool) []Line {
	var out []Line
	for _, e := range shared {
		var best int
		if !capped && !over(e, a, b, cfg) {
			for _, x := range a.persons[e] {
				for _, y := range b.persons[e] {
					best = max(best, len(properties(ha.of(x, cfg), hb.of(y, cfg), x, y, metas, cfg))+len(links(x, y, a, b, cfg)))
				}
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

// over reports whether either graph files more members of any kind on entity e
// than the registry allows.
func over(e string, a, b Graph, cfg Config) bool {
	most := cfg.RecordLinks.MaxEntityMembers
	return a.members[e] > most || b.members[e] > most
}

// properties returns the keys, ascending, whose values the two members agree
// or partly agree on. A property counts once under its key, with its pooled
// properties; one that is skipped for either member (see tally.skipped)
// counts nothing. heldX and heldY are the members' holdings.
func properties(heldX, heldY map[vocabkey.PropertyKey]tally, x, y Member, metas subjectcompare.Metas, cfg Config) []vocabkey.PropertyKey {
	keys := map[vocabkey.PropertyKey]struct{}{}
	for key := range heldX {
		keys[key] = struct{}{}
	}
	for key := range heldY {
		keys[key] = struct{}{}
	}
	var out []vocabkey.PropertyKey
	for _, key := range slices.Sorted(maps.Keys(keys)) {
		if heldX[key].skipped(cfg) || heldY[key].skipped(cfg) {
			continue
		}
		res := subjectcompare.CompareProperty(vocabkey.Person, key, x.Values, y.Values, metas, cfg.Subject, cfg.Value)
		if res.Outcome.Matches() {
			out = append(out, key)
		}
	}
	return out
}

// tally is what a member holds for one property that counts: its values, the
// words of the values within MaxValueBytes, and whether any value is over it.
type tally struct {
	values, words int
	big           bool
}

// skipped reports whether the property counts nothing for the member: more than
// MaxUnitValues values, or a value over MaxValueBytes.
func (t tally) skipped(cfg Config) bool {
	return t.big || t.values > cfg.RecordLinks.MaxUnitValues
}

// held tallies a member's values by the key their property counts under. A
// value over MaxValueBytes is sized, never tokenized.
func held(v subjectcompare.Values, cfg Config) map[vocabkey.PropertyKey]tally {
	out := map[vocabkey.PropertyKey]tally{}
	for p, vs := range v {
		if key, ok := cfg.Subject.CountsUnder(vocabkey.Person, p); ok {
			t := out[key]
			t.values += len(vs)
			for _, val := range vs {
				if valuecompare.Size(val.Value) > cfg.RecordLinks.MaxValueBytes {
					t.big = true
					continue
				}
				t.words += valuecompare.WordCount(val.Value)
			}
			out[key] = t
		}
	}
	return out
}

// holdings are the members' tallies of one graph by member id, each read once.
type holdings map[string]map[vocabkey.PropertyKey]tally

// of is m's tally, read the first time it is asked for.
func (h holdings) of(m Member, cfg Config) map[vocabkey.PropertyKey]tally {
	t, ok := h[string(m.ID)]
	if !ok {
		t = held(m.Values, cfg)
		h[string(m.ID)] = t
	}
	return t
}

// cost is what one member brings to the work count: its words by property
// that counts and, by far entity, the shapes of its links.
type cost struct {
	words  map[vocabkey.PropertyKey]int
	shapes map[string]int
}

// costOf reads a member's cost. A skipped property and a far entity over
// MaxEntityMembers compare nothing, so they cost nothing.
func costOf(m Member, g Graph, h holdings, a, b Graph, cfg Config) cost {
	c := cost{words: map[vocabkey.PropertyKey]int{}, shapes: map[string]int{}}
	for key, t := range h.of(m, cfg) {
		if !t.skipped(cfg) {
			c.words[key] = t.words
		}
	}
	for e, ls := range g.links[string(m.ID)] {
		if over(e, a, b, cfg) {
			continue
		}
		for _, l := range ls {
			c.shapes[e] += max(1, len(l.rec.Shapes))
		}
	}
	return c
}

// against is the work of comparing two members: the product of their words for
// each property and of their link shapes for each entity both link to.
func (x cost) against(y cost) int {
	var n int
	for key, w := range x.words {
		n += w * y.words[key]
	}
	for e, s := range x.shapes {
		n += s * y.shapes[e]
	}
	return n
}

// overWork reports whether the work of comparing every member pair of every
// shared person passes MaxPairWork.
func overWork(shared []string, a, b Graph, ha, hb holdings, cfg Config) bool {
	return pairWork(shared, a, b, ha, hb, cfg, cfg.RecordLinks.MaxPairWork) > cfg.RecordLinks.MaxPairWork
}

// walk is what comparing m against any member walks before a product is
// formed: every link it holds and every property key it holds that counts. A
// link to an entity the other graph lacks is walked all the same.
func walk(m Member, g Graph, h holdings, cfg Config) int {
	return g.linkCount[string(m.ID)] + len(h.of(m, cfg))
}

// pairWork is the sum, over every member pair of every shared person (a sorted
// list that excludes the page), of the pair's work: the links and property keys
// both members walk, then the pair's products. A sum does not depend on the
// order of the graphs or of their members. It stops counting once past limit,
// so the result is the work only when it is within limit. An entity's walks
// are charged before any member's cost is built, so a member whose links alone
// pass limit is never walked per pair.
func pairWork(shared []string, a, b Graph, ha, hb holdings, cfg Config, limit int) int {
	var n int
	for _, e := range shared {
		xs, ys := a.persons[e], b.persons[e]
		if len(ys) == 0 || over(e, a, b, cfg) {
			continue
		}
		for _, x := range xs {
			wx := walk(x, a, ha, cfg)
			for _, y := range ys {
				if n += wx + walk(y, b, hb, cfg); n > limit {
					return n
				}
			}
		}
		costY := make([]cost, len(ys))
		for i, y := range ys {
			costY[i] = costOf(y, b, hb, a, b, cfg)
		}
		for _, x := range xs {
			cx := costOf(x, a, ha, a, b, cfg)
			for _, cy := range costY {
				if n += cx.against(cy); n > limit {
					return n
				}
			}
		}
	}
	return n
}

// linkFact is one counted link: the link as the member reads it, and the
// entity at its far end.
type linkFact struct {
	key    vocabkey.BridgeKey
	entity string
}

// links returns the link facts the two members share. For each far entity both
// have links to, a pair of links that correspond counts once when any such pair
// is not skipped. A pair is skipped when kinship.SameShape matches their shapes
// at pivots filed on one entity E in both graphs and the two members' links to
// E hold a corresponding pair: the pivot's own fact is then counted, and this
// one counts through it. A pivot link that is itself skipped counts through its
// own pivot, so shapes end at a recorded participation or parent link. Links to
// an entity over the member bound count nothing, so a pivot filed on one is not
// counted and its links are not skipped.
func links(x, y Member, a, b Graph, cfg Config) []linkFact {
	byX, byY := a.links[string(x.ID)], b.links[string(y.ID)]
	counted := map[string]bool{} // by pivot entity: the members' links to it correspond
	same := func(pa, pb []byte) bool {
		ea, eb := a.entityOf[string(pa)], b.entityOf[string(pb)]
		if ea == nil || !bytes.Equal(ea, eb) || over(string(ea), a, b, cfg) {
			return false
		}
		v, ok := counted[string(ea)]
		if !ok {
			v = anyCorresponds(byX[string(ea)], byY[string(ea)])
			counted[string(ea)] = v
		}
		return v
	}
	var out []linkFact
	for _, e := range slices.Sorted(maps.Keys(byX)) {
		if over(e, a, b, cfg) {
			continue
		}
		seen := map[vocabkey.BridgeKey]bool{}
		for _, lx := range byX[e] {
			if seen[lx.key] {
				continue
			}
			for _, ly := range byY[e] {
				if segmentcompare.Corresponds(lx.key, ly.key) && !kinship.SameShape(lx.rec, ly.rec, same) {
					seen[lx.key] = true
					out = append(out, linkFact{lx.key, e})
					break
				}
			}
		}
	}
	return out
}

// anyCorresponds reports whether a link of xs and a link of ys correspond.
func anyCorresponds(xs, ys []memberLink) bool {
	return slices.ContainsFunc(xs, func(lx memberLink) bool {
		return slices.ContainsFunc(ys, func(ly memberLink) bool { return segmentcompare.Corresponds(lx.key, ly.key) })
	})
}
