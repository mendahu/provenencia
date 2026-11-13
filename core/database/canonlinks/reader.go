package canonlinks

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"slices"

	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// pivotKind says what a cached result is a result of.
type pivotKind int

const (
	// pivotEvent is what one event implies: the event patterns over the
	// event's participation links.
	pivotEvent pivotKind = iota + 1
	// pivotChain is what one person implies as the parent between: the
	// parent-of-parent chain over their relationship links and the births
	// they stand in.
	pivotChain
)

// pivotResult is one stored result. Results are shared and read-only.
type pivotResult struct {
	kind  pivotKind
	rules kinship.Rules
	links []kinship.Implied
}

// maxKeys caps each kinship entry: a crafted catalog cannot grow one past it,
// and past it the least recently used key leaves and is derived again on read.
const maxKeys = 65_536

// pivots holds what each event, or each person as the parent between, implies.
// A commit drops a result for each handle it touches and for each neighbor of
// those: an event result reads the event's links, a chain result reads the
// person's links and the results of their events, and both are one hop from
// what they read.
var pivots graphcache.Keyed[pivotResult]

func init() {
	pivots = graphcache.RegisterKeyed(graphcache.KeyedEntry[pivotResult]{
		Name:        "canonlinks.kinship",
		DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}},
		Apply: func(g *graphcache.Graph, k graphcache.Keyed[pivotResult], roots [][]byte, _ effects.Set) error {
			k.Drop(g, roots)
			k.Drop(g, deps(g, roots))
			return nil
		},
		Check: check,
		Cap:   maxKeys,
	})
}

// deps names every neighbor of each root, and the root.
func deps(g *graphcache.Graph, roots [][]byte) [][]byte {
	var out [][]byte
	for _, id := range roots {
		out = append(out, id)
		if n, ok := g.Cached(id); ok {
			out = append(out, neighborsOf([]*graphcache.Node{n})...)
		}
	}
	return out
}

// check rebuilds a stored result from the cached nodes and compares.
func check(g *graphcache.Graph, id []byte, res pivotResult) error {
	// The check must not load: it reads only nodes already cached.
	d, err := newDeriver(g, res.rules, cachedNodes(g))
	if err != nil {
		return err
	}
	var links []kinship.Implied
	switch res.kind {
	case pivotEvent:
		links, err = d.eventLinks(id)
	case pivotChain:
		links, err = d.chainLinks(id, d.eventLinks)
	default:
		err = fmt.Errorf("canonlinks: stored result %x has kind %d", id, res.kind)
	}
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(res, pivotResult{kind: res.kind, rules: res.rules, links: links}) {
		return fmt.Errorf("canonlinks: stored result %x does not match a fresh derivation", id)
	}
	return nil
}

// nodeSource is the one load policy of a derivation: node reads one person's
// node, and load makes a set of nodes available to it.
type nodeSource struct {
	node func([]byte) (*graphcache.Node, error)
	load func([][]byte) error
}

// loadingNodes reads nodes from the graph, loading what is missing.
func loadingNodes(g *graphcache.Graph) nodeSource {
	return nodeSource{node: g.Node, load: g.Nodes}
}

// cachedNodes reads only nodes already cached, for a check, which must not
// load: a node that is not cached is an error, and a load does nothing.
func cachedNodes(g *graphcache.Graph) nodeSource {
	return nodeSource{
		node: func(id []byte) (*graphcache.Node, error) {
			n, ok := g.Cached(id)
			if !ok {
				return nil, fmt.Errorf("canonlinks: %x is not cached", id)
			}
			return n, nil
		},
		load: func([][]byte) error { return nil },
	}
}

// deriver derives one pivot's links from graph nodes, read through src.
type deriver struct {
	g             *graphcache.Graph
	src           nodeSource
	rules         kinship.Rules
	terms         segmentcompare.Terms
	eventTypeProp []byte
}

func newDeriver(g *graphcache.Graph, rules kinship.Rules, src nodeSource) (*deriver, error) {
	terms, err := LoadTerms(g)
	if err != nil {
		return nil, err
	}
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	return &deriver{g: g, src: src, rules: rules, terms: terms, eventTypeProp: v.PropertyID(vocabkey.PropertyEventType, vocabkey.OriginProvenencia)}, nil
}

// edge reads one link of n as an edge that stands for that link alone.
func (d *deriver) edge(n *graphcache.Node, l graphcache.Link) (kinship.Edge, bool) {
	from, to, keyFrom, keyTo, ok := EdgeOf(n.ID, n, l, nil, d.eventTypeProp, d.terms)
	if !ok {
		return kinship.Edge{}, false
	}
	return kinship.Edge{A: from, B: to, KeyA: keyFrom, KeyB: keyTo, Associations: [][]byte{l.Association}}, true
}

// relationshipEdges reads each relationship link of n that edge reads, in
// n.Links order: the links that are kin.
func (d *deriver) relationshipEdges(n *graphcache.Node) []kinship.Edge {
	var out []kinship.Edge
	for _, l := range n.Links {
		if l.BridgeTypeKey != vocabkey.BridgeRelationship {
			continue
		}
		if e, ok := d.edge(n, l); ok {
			out = append(out, e)
		}
	}
	return out
}

// kinEdges is n's parent links for a canon read: each relationship link filed
// on n, then the edge of each birth link in atEnd (n's event links).
func (d *deriver) kinEdges(n *graphcache.Node, atEnd []kinship.Implied) []kinship.Edge {
	edges := d.relationshipEdges(n)
	for _, imp := range atEnd {
		if imp.Pattern == kinship.PatternBirthParents {
			edges = append(edges, imp.Edge)
		}
	}
	return edges
}

// parentEdges is n's parent links for a canon read: the links of the events n
// stands in (atEnd), then n's kin edges. event gives an event's links.
func (d *deriver) parentEdges(n *graphcache.Node, event func([]byte) ([]kinship.Implied, error)) (atEnd []kinship.Implied, edges []kinship.Edge, err error) {
	atEnd, err = d.eventLinksAt(n, event)
	if err != nil {
		return nil, nil, err
	}
	return atEnd, d.kinEdges(n, atEnd), nil
}

// eventLinks is what the event patterns imply over one event's links.
func (d *deriver) eventLinks(id []byte) ([]kinship.Implied, error) {
	n, err := d.src.node(id)
	if err != nil || n == nil {
		return nil, err
	}
	var edges []kinship.Edge
	for _, l := range n.Links {
		if e, ok := d.edge(n, l); ok {
			edges = append(edges, e)
		}
	}
	// Derive also chains parents among the event's own links; a chain is a
	// person's result, not an event's.
	var out []kinship.Implied
	for _, imp := range kinship.Derive(edges, d.terms, d.rules) {
		if imp.Event != nil {
			out = append(out, imp)
		}
	}
	return out, nil
}

// eventLinksAt returns the links of each event n stands in, each event once,
// that have n as an end. eventResult gives one event's links.
func (d *deriver) eventLinksAt(n *graphcache.Node, eventResult func([]byte) ([]kinship.Implied, error)) ([]kinship.Implied, error) {
	var out []kinship.Implied
	seen := map[string]bool{}
	for _, l := range n.Links {
		if l.NeighborTypeKey != vocabkey.Event || seen[string(l.Neighbor)] {
			continue
		}
		seen[string(l.Neighbor)] = true
		links, err := eventResult(l.Neighbor)
		if err != nil {
			return nil, err
		}
		for _, imp := range links {
			if bytes.Equal(imp.A, n.ID) || bytes.Equal(imp.B, n.ID) {
				out = append(out, imp)
			}
		}
	}
	return out, nil
}

// chainLinks is what the parent-of-parent chain implies with one person as the
// parent between. It reads the person's relationship links and the birth links
// with the person at an end from the results of the events they stand in.
func (d *deriver) chainLinks(id []byte, eventResult func([]byte) ([]kinship.Implied, error)) ([]kinship.Implied, error) {
	n, err := d.src.node(id)
	if err != nil || n == nil {
		return nil, err
	}
	_, edges, err := d.parentEdges(n, eventResult)
	if err != nil {
		return nil, err
	}
	return kinship.Derive(edges, d.terms, d.rules), nil
}

// Reader reads the links the kinship patterns imply, from one cached result
// per pivot. A pivot is an event, or the parent between in a chain.
type Reader struct {
	g *graphcache.Graph
	d *deriver
}

// NewReader loads the catalog's bridge terms and event type property once. It
// reads with the registry's kinship rules.
func NewReader(g *graphcache.Graph) (*Reader, error) {
	d, err := newDeriver(g, matchregistry.Default().Kinship, loadingNodes(g))
	if err != nil {
		return nil, err
	}
	return &Reader{g: g, d: d}, nil
}

// pivot returns the stored result of kind for id, deriving and storing it when
// there is none. An id held under the other kind is derived again and not
// stored: one handle is both an event and, as a parent between, a person.
func (r *Reader) pivot(kind pivotKind, id []byte, derive func() ([]kinship.Implied, error)) ([]kinship.Implied, error) {
	result := func() (pivotResult, error) {
		links, err := derive()
		return pivotResult{kind: kind, rules: r.d.rules, links: links}, err
	}
	if held, ok := pivots.Peek(r.g, id); ok {
		if held.kind == kind {
			return held.links, nil
		}
		res, err := result()
		return res.links, err
	}
	res, err := pivots.Get(r.g, id, result)
	return res.links, err
}

func (r *Reader) event(id []byte) ([]kinship.Implied, error) {
	return r.pivot(pivotEvent, id, func() ([]kinship.Implied, error) { return r.d.eventLinks(id) })
}

func (r *Reader) chain(id []byte) ([]kinship.Implied, error) {
	return r.pivot(pivotChain, id, func() ([]kinship.Implied, error) { return r.d.chainLinks(id, r.event) })
}

// For returns every spouse, parent and grandparent link the patterns imply
// with id as an end: the links of each event id stands in, then the chain links
// through each parent and child of id. It composes the answer from the pivot
// results at each call and stores none; the slice is the caller's to read and
// must not be changed, since its links are the pivots'. Links from different
// pivots differ in Event or Through, so none is repeated. An id that is no node
// has no links.
func (r *Reader) For(id []byte) ([]kinship.Implied, error) {
	steps, err := r.d.loadSteps([][]byte{id}, r.event)
	if err != nil || len(steps) == 0 {
		return nil, err
	}
	return r.d.forLinks(steps[0], r.chain)
}

// Options are what Relationships returns beyond the person's recorded and
// For-implied links. Siblings adds the siblings a shared parent implies.
type Options struct{ Siblings bool }

// Relationships returns each id's relationship links, recorded and implied, one
// per pair of ends and keys: each relationship link filed on the id, then For's
// links, then, with opts.Siblings, the siblings a shared parent implies, merged
// through one kinship.LinkSet. It loads what it reads in a fixed number of
// batches however many ids and relatives there are. An id that is no node has no
// entry. The caller owns each slice and its bytes. Nothing is filtered: a link to
// a merged person is returned, as For returns implied links to one.
func (r *Reader) Relationships(ids [][]byte, opts Options) (map[string][]kinship.Link, error) {
	steps, err := r.d.loadSteps(ids, r.event)
	if err != nil {
		return nil, err
	}
	var siblings map[string][]kinship.Implied
	if opts.Siblings {
		siblings, err = r.d.siblingsOf(steps, r.event)
		if err != nil {
			return nil, err
		}
	}
	out := make(map[string][]kinship.Link, len(steps))
	for _, s := range steps {
		set := kinship.NewLinkSet()
		for _, e := range r.d.relationshipEdges(s.node) {
			set.Add(kinship.Link{Edge: e})
		}
		implied, err := r.d.forLinks(s, r.chain)
		if err != nil {
			return nil, err
		}
		for _, imp := range implied {
			set.AddImplied(imp)
		}
		for _, imp := range siblings[string(s.node.ID)] {
			set.AddImplied(imp)
		}
		out[string(s.node.ID)] = set.Links()
	}
	return out, nil
}

// stepped is one id's node, the links of the events it stands in that have it
// at an end, its parent links for a canon read, and the people one parent step
// away over them.
type stepped struct {
	node   *graphcache.Node
	atEnd  []kinship.Implied
	edges  []kinship.Edge
	people [][]byte
}

// stepPeople returns the distinct people one parent step from id over edges,
// first-seen, never id: each edge's child when id is its parent, else its
// parent. An edge that is no parent step is skipped.
func stepPeople(id []byte, edges []kinship.Edge, terms segmentcompare.Terms) [][]byte {
	seen := map[string]bool{string(id): true}
	var people [][]byte
	add := func(person []byte) {
		if !seen[string(person)] {
			seen[string(person)] = true
			people = append(people, person)
		}
	}
	for _, e := range edges {
		parent, child, ok := kinship.ParentStep(e, terms)
		if !ok {
			continue
		}
		if bytes.Equal(parent, id) {
			add(child)
		} else {
			add(parent)
		}
	}
	return people
}

// loadSteps loads, in four batches however many ids there are, the ids, their
// neighbors, the people one parent step from each, and those people's
// neighbors. It returns, per distinct id that is a node, its stepped; an id that
// is no node is skipped. event gives an event's links. A person that is linked
// but not cached is skipped by the caller that reads it.
func (d *deriver) loadSteps(ids [][]byte, event func([]byte) ([]kinship.Implied, error)) ([]stepped, error) {
	var want [][]byte
	seen := map[string]bool{}
	for _, id := range ids {
		if !seen[string(id)] {
			seen[string(id)] = true
			want = append(want, id)
		}
	}
	if err := d.src.load(want); err != nil {
		return nil, err
	}
	var nodes []*graphcache.Node
	for _, id := range want {
		if n, ok := d.g.Cached(id); ok {
			nodes = append(nodes, n)
		}
	}
	if err := d.src.load(neighborsOf(nodes)); err != nil {
		return nil, err
	}
	out := make([]stepped, len(nodes))
	var allPeople [][]byte
	for i, n := range nodes {
		atEnd, edges, err := d.parentEdges(n, event)
		if err != nil {
			return nil, err
		}
		out[i] = stepped{node: n, atEnd: atEnd, edges: edges, people: stepPeople(n.ID, edges, d.terms)}
		allPeople = append(allPeople, out[i].people...)
	}
	if err := d.src.load(allPeople); err != nil {
		return nil, err
	}
	var peopleNodes []*graphcache.Node
	have := map[string]bool{}
	for _, q := range allPeople {
		if n, ok := d.g.Cached(q); ok && !have[string(q)] {
			have[string(q)] = true
			peopleNodes = append(peopleNodes, n)
		}
	}
	if err := d.src.load(neighborsOf(peopleNodes)); err != nil {
		return nil, err
	}
	return out, nil
}

// siblingsOf derives the siblings of each stepped node, from the steps loadSteps
// loaded. A node with no siblings has an empty answer. event gives an event's
// links. A parent that is linked but not cached is skipped.
func (d *deriver) siblingsOf(steps []stepped, event func([]byte) ([]kinship.Implied, error)) (map[string][]kinship.Implied, error) {
	parents := make([][][]byte, len(steps))
	edgesOfParent := map[string][]kinship.Edge{}
	for i, s := range steps {
		parents[i] = kinship.Parents(s.node.ID, s.edges, d.terms, d.rules)
		for _, q := range parents[i] {
			if _, done := edgesOfParent[string(q)]; done {
				continue
			}
			n, ok := d.g.Cached(q)
			if !ok {
				continue
			}
			_, es, err := d.parentEdges(n, event)
			if err != nil {
				return nil, err
			}
			edgesOfParent[string(q)] = es
		}
	}
	out := make(map[string][]kinship.Implied, len(steps))
	for i, s := range steps {
		all := slices.Clone(s.edges)
		read := map[string]bool{}
		for _, e := range all {
			read[edgeKey(e)] = true
		}
		for _, q := range parents[i] {
			for _, e := range edgesOfParent[string(q)] {
				if k := edgeKey(e); !read[k] {
					read[k] = true
					all = append(all, e)
				}
			}
		}
		out[string(s.node.ID)] = kinship.Siblings(s.node.ID, all, d.terms, d.rules)
	}
	return out, nil
}

// edgeKey names the link an edge was read from: its ends and associations, each
// length-prefixed so no two distinct links share a key. Two distinct links
// between the same people differ in their associations.
func edgeKey(e kinship.Edge) string {
	var key []byte
	part := func(b []byte) {
		key = binary.AppendUvarint(key, uint64(len(b)))
		key = append(key, b...)
	}
	part(e.A)
	part(e.B)
	for _, a := range e.Associations {
		part(a)
	}
	return string(key)
}

// neighborsOf names every neighbor of each node.
func neighborsOf(nodes []*graphcache.Node) [][]byte {
	var ids [][]byte
	for _, n := range nodes {
		for _, l := range n.Links {
			ids = append(ids, l.Neighbor)
		}
	}
	return ids
}

// forLinks is For's derivation over one stepped node: the links of the events
// it stands in, then the chain links with the node at an end through each person
// one parent step away. chain gives a person's chain links.
func (d *deriver) forLinks(s stepped, chain func([]byte) ([]kinship.Implied, error)) ([]kinship.Implied, error) {
	out := slices.Clone(s.atEnd)
	for _, m := range s.people {
		links, err := chain(m)
		if err != nil {
			return nil, err
		}
		for _, imp := range links {
			if bytes.Equal(imp.A, s.node.ID) || bytes.Equal(imp.B, s.node.ID) {
				out = append(out, imp)
			}
		}
	}
	return out, nil
}
