package canonlinks

import (
	"bytes"
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

// personAnswer is one person's stored answer, For's or the siblings'. It is
// shared and read-only.
type personAnswer struct {
	rules kinship.Rules
	links []kinship.Implied
}

// maxKeys caps each kinship entry: a crafted catalog cannot grow one past it,
// and a key not stored is derived again on read.
const maxKeys = 65_536

// pivots holds what each event, or each person as the parent between, implies.
// A commit drops a result for each handle it touches and for each neighbor of
// those: an event result reads the event's links, a chain result reads the
// person's links and the results of their events, and both are one hop from
// what they read.
var pivots graphcache.Keyed[pivotResult]

// forAnswers holds each person's full For answer. It reads nodes up to three
// hops from the person (a birth, the parent, the parent's events), past what
// the pivots' Apply can name, so any commit that names a handle drops all of
// it. S16-16 deletes this entry.
var forAnswers graphcache.Keyed[personAnswer]

// siblingAnswers holds each person's implied siblings, for the page. A sibling
// is two to four hops away, so any commit that names a handle drops all of it.
// S16-16 deletes this entry.
var siblingAnswers graphcache.Keyed[personAnswer]

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
	forAnswers = graphcache.RegisterKeyed(graphcache.KeyedEntry[personAnswer]{
		Name:        "canonlinks.for",
		DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}},
		Apply:       clearOnHandles[personAnswer],
		Check:       checkFor,
		Cap:         maxKeys,
	})
	siblingAnswers = graphcache.RegisterKeyed(graphcache.KeyedEntry[personAnswer]{
		Name:        "canonlinks.siblings",
		DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}},
		Apply:       clearOnHandles[personAnswer],
		Check:       checkSiblings,
		Cap:         maxKeys,
	})
}

// clearOnHandles empties the entry when a commit names any handle.
func clearOnHandles[T any](g *graphcache.Graph, k graphcache.Keyed[T], roots [][]byte, _ effects.Set) error {
	if len(roots) > 0 {
		k.Clear(g)
	}
	return nil
}

// checkPersonAnswer rebuilds a stored per-person answer from the cached nodes
// and compares. derive makes the answer for one id with a deriver that loads
// nothing; found is false when the id has no answer to make.
func checkPersonAnswer(derive func(d *deriver, id []byte) (links []kinship.Implied, found bool, err error)) func(g *graphcache.Graph, id []byte, res personAnswer) error {
	return func(g *graphcache.Graph, id []byte, res personAnswer) error {
		d, err := newDeriver(g, res.rules, cachedNodes(g))
		if err != nil {
			return err
		}
		links, found, err := derive(d, id)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("canonlinks: stored answer %x is for no cached person", id)
		}
		if !reflect.DeepEqual(res, personAnswer{rules: res.rules, links: links}) {
			return fmt.Errorf("canonlinks: stored answer %x does not match a fresh derivation", id)
		}
		return nil
	}
}

// checkFor checks a stored For answer: no pivot results are used.
var checkFor = checkPersonAnswer(func(d *deriver, id []byte) ([]kinship.Implied, bool, error) {
	links, err := d.forLinks(id, d.eventLinks, func(m []byte) ([]kinship.Implied, error) {
		return d.chainLinks(m, d.eventLinks)
	})
	return links, true, err
})

// checkSiblings checks a stored sibling answer.
var checkSiblings = checkPersonAnswer(func(d *deriver, id []byte) ([]kinship.Implied, bool, error) {
	answers, err := d.siblingsOf([][]byte{id}, d.eventLinks)
	if err != nil {
		return nil, false, err
	}
	links, found := answers[string(id)]
	return links, found, nil
})

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

// around reads a person's node and makes their neighbors available.
func (d *deriver) around(id []byte) (*graphcache.Node, error) {
	n, err := d.src.node(id)
	if err != nil || n == nil {
		return nil, err
	}
	return n, d.src.load(neighborsOf([]*graphcache.Node{n}))
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
// stands in, then n's kin edges. event gives an event's links.
func (d *deriver) parentEdges(n *graphcache.Node, event func([]byte) ([]kinship.Implied, error)) ([]kinship.Edge, error) {
	atEnd, err := d.eventLinksAt(n, event)
	if err != nil {
		return nil, err
	}
	return d.kinEdges(n, atEnd), nil
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
	edges, err := d.parentEdges(n, eventResult)
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
// through each parent and child of id. The answer is stored until the next
// commit that names a handle; the slice and its links are the cache's, shared
// and read-only. Links from different pivots differ in Event or Through, so
// none is repeated.
func (r *Reader) For(id []byte) ([]kinship.Implied, error) {
	got, err := forAnswers.GetMany(r.g, [][]byte{id}, func(missing [][]byte) (map[string]personAnswer, error) {
		n, err := r.d.around(id)
		if err != nil || n == nil {
			// An id that is no node has no answer to keep.
			return map[string]personAnswer{}, err
		}
		out, err := r.d.forLinks(id, r.event, r.chain)
		if err != nil {
			return nil, err
		}
		return map[string]personAnswer{string(id): {rules: r.d.rules, links: out}}, nil
	})
	if err != nil {
		return nil, err
	}
	return got[string(id)].links, nil
}

// Options are what Relationships returns beyond the person's recorded and
// For-implied links. Siblings adds the siblings a shared parent implies.
type Options struct{ Siblings bool }

// Relationships returns id's relationship links, recorded and implied, one per
// pair of ends and keys: each relationship link filed on id, then For's links,
// then, with opts.Siblings, the siblings a shared parent implies, merged through
// one kinship.LinkSet. The caller owns the slice and its bytes. Nothing is
// filtered: a link to a merged person is returned, as For returns implied links
// to one.
func (r *Reader) Relationships(id []byte, opts Options) ([]kinship.Link, error) {
	n, err := r.d.around(id)
	if err != nil || n == nil {
		return nil, err
	}
	set := kinship.NewLinkSet()
	for _, e := range r.d.relationshipEdges(n) {
		set.Add(kinship.Link{Edge: e})
	}
	implied, err := r.For(id)
	if err != nil {
		return nil, err
	}
	for _, imp := range implied {
		set.AddImplied(imp)
	}
	if opts.Siblings {
		answers, err := r.siblings([][]byte{id})
		if err != nil {
			return nil, err
		}
		for _, imp := range answers[string(id)].links {
			set.AddImplied(imp)
		}
	}
	return set.Links(), nil
}

// LoadRelationships loads what Relationships(id, opts) reads for each id in a
// fixed number of batches, however many ids and relatives: the ids, their
// neighbors, the people one parent step from each, and those people's
// neighbors. It then derives and stores each id's For answer, and with
// opts.Siblings the siblings'. After it, Relationships for each id issues no
// query. An id that is no node stores nothing.
func (r *Reader) LoadRelationships(ids [][]byte, opts Options) error {
	if _, err := r.d.loadSteps(ids, r.event); err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := r.For(id); err != nil {
			return err
		}
	}
	if opts.Siblings {
		return r.LoadSiblings(ids)
	}
	return nil
}

// LoadSiblings derives and stores the implied siblings of each id that has none
// stored, loading every id's nodes in a fixed number of batches. An id that is
// no node stores nothing.
func (r *Reader) LoadSiblings(ids [][]byte) error {
	_, err := r.siblings(ids)
	return err
}

// siblings is each id's stored sibling answer, derived for those without one.
// An id that is no node has no entry.
func (r *Reader) siblings(ids [][]byte) (map[string]personAnswer, error) {
	return siblingAnswers.GetMany(r.g, ids, func(missing [][]byte) (map[string]personAnswer, error) {
		answers, err := r.d.siblingsOf(missing, r.event)
		if err != nil {
			return nil, err
		}
		out := make(map[string]personAnswer, len(answers))
		for key, links := range answers {
			out[key] = personAnswer{rules: r.d.rules, links: links}
		}
		return out, nil
	})
}

// stepped is one id's node, its parent links for a canon read, and the people
// one parent step away over them.
type stepped struct {
	node   *graphcache.Node
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
		edges, err := d.parentEdges(n, event)
		if err != nil {
			return nil, err
		}
		out[i] = stepped{node: n, edges: edges, people: stepPeople(n.ID, edges, d.terms)}
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

// siblingsOf derives the siblings of each id that is a node, from loadSteps'
// four batched loads however many ids and siblings there are. An id with no
// entry is no node; a node with no siblings has an empty answer. event gives an
// event's links. A parent that is linked but not cached is skipped.
func (d *deriver) siblingsOf(ids [][]byte, event func([]byte) ([]kinship.Implied, error)) (map[string][]kinship.Implied, error) {
	steps, err := d.loadSteps(ids, event)
	if err != nil {
		return nil, err
	}
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
			es, err := d.parentEdges(n, event)
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

// edgeKey names the link an edge was read from: its ends and associations. Two
// distinct links between the same people differ in their associations.
func edgeKey(e kinship.Edge) string {
	return string(e.A) + string(e.B) + string(bytes.Join(e.Associations, nil))
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

// forLinks is For's derivation. event and chain give a pivot's links.
func (d *deriver) forLinks(id []byte, event, chain func([]byte) ([]kinship.Implied, error)) ([]kinship.Implied, error) {
	n, err := d.around(id)
	if err != nil || n == nil {
		return nil, err
	}
	atEnd := func(imp kinship.Implied) bool { return bytes.Equal(imp.A, id) || bytes.Equal(imp.B, id) }
	out, err := d.eventLinksAt(n, event)
	if err != nil {
		return nil, err
	}

	// The people one parent step from id: filed parents and children first,
	// then the birth links above.
	people := stepPeople(id, d.kinEdges(n, out), d.terms)
	for _, m := range people {
		if _, err := d.around(m); err != nil {
			return nil, err
		}
		links, err := chain(m)
		if err != nil {
			return nil, err
		}
		for _, imp := range links {
			if atEnd(imp) {
				out = append(out, imp)
			}
		}
	}
	return out, nil
}
