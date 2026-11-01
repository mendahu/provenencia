package canonlinks

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"

	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// idLen is the length of a Subject id; graphcache stores no answer for any other.
const idLen = 16

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

// kinshipSlot is where results are stored on the graph. A commit drops a
// result for each handle it touches and for each neighbor of those: an event
// result reads the event's links, a chain result reads the person's links and
// the results of their events, and both are one hop from what they read.
var kinshipSlot graphcache.MemoSlot

// siblingSlot is where each person's implied siblings are stored, for the page.
// A sibling is two to four hops away, past what Deps can name, so any commit that
// names a handle drops the whole slot.
var siblingSlot graphcache.MemoSlot

// forSlot is where each person's full For answer is stored. It reads nodes up
// to three hops from the person (a birth, the parent, the parent's events), so
// any commit that names a handle drops the whole slot.
var forSlot graphcache.MemoSlot

// personAnswer is one person's stored answer, For's or the siblings'. It is
// shared and read-only.
type personAnswer struct {
	rules kinship.Rules
	links []kinship.Implied
}

func init() {
	kinshipSlot = graphcache.RegisterNodeMemo(graphcache.NodeMemo{
		Name:             "canonlinks.kinship",
		Check:            check,
		Deps:             deps,
		DropOnVocabulary: true,
	})
	forSlot = graphcache.RegisterNodeMemo(graphcache.NodeMemo{
		Name:               "canonlinks.for",
		Check:              checkFor,
		DropOnVocabulary:   true,
		DropOnHandleCommit: true,
	})
	siblingSlot = graphcache.RegisterNodeMemo(graphcache.NodeMemo{
		Name:               "canonlinks.siblings",
		Check:              checkSiblings,
		DropOnVocabulary:   true,
		DropOnHandleCommit: true,
	})
}

// checkFor rebuilds a stored For answer from the cached nodes, with no pivot
// results, and compares.
func checkFor(g *graphcache.Graph, id []byte, stored any) error {
	res, ok := stored.(personAnswer)
	if !ok {
		return fmt.Errorf("canonlinks: stored answer %x has type %T", id, stored)
	}
	d, err := newDeriver(g, res.rules, cachedNodes(g))
	if err != nil {
		return err
	}
	links, err := d.forLinks(id, d.eventLinks, func(m []byte) ([]kinship.Implied, error) {
		return d.chainLinks(m, d.eventLinks)
	})
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(res, personAnswer{rules: res.rules, links: links}) {
		return fmt.Errorf("canonlinks: stored answer %x does not match a fresh derivation", id)
	}
	return nil
}

// checkSiblings rebuilds a stored sibling answer from the cached nodes and
// compares.
func checkSiblings(g *graphcache.Graph, id []byte, stored any) error {
	res, ok := stored.(personAnswer)
	if !ok {
		return fmt.Errorf("canonlinks: stored answer %x has type %T", id, stored)
	}
	d, err := newDeriver(g, res.rules, cachedNodes(g))
	if err != nil {
		return err
	}
	answers, err := d.siblingsOf([][]byte{id}, d.eventLinks)
	if err != nil {
		return err
	}
	links, ok := answers[string(id)]
	if !ok {
		return fmt.Errorf("canonlinks: stored answer %x is for no cached person", id)
	}
	if !reflect.DeepEqual(res, personAnswer{rules: res.rules, links: links}) {
		return fmt.Errorf("canonlinks: stored answer %x does not match a fresh derivation", id)
	}
	return nil
}

// deps names every neighbor of each root, and the root.
func deps(g *graphcache.Graph, roots [][]byte) ([][]byte, error) {
	var out [][]byte
	for _, id := range roots {
		out = append(out, id)
		if n, ok := g.Cached(id); ok {
			for _, l := range n.Links {
				out = append(out, l.Neighbor)
			}
		}
	}
	return out, nil
}

// check rebuilds a stored result from the cached nodes and compares.
func check(g *graphcache.Graph, id []byte, stored any) error {
	res, ok := stored.(pivotResult)
	if !ok {
		return fmt.Errorf("canonlinks: stored result %x has type %T", id, stored)
	}
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
	ids := make([][]byte, 0, len(n.Links))
	for _, l := range n.Links {
		ids = append(ids, l.Neighbor)
	}
	return n, d.src.load(ids)
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
	atEnd, err := d.eventLinksAt(n, eventResult)
	if err != nil {
		return nil, err
	}
	return kinship.Derive(d.kinEdges(n, atEnd), d.terms, d.rules), nil
}

// Reader reads the links the kinship patterns imply, from one cached result
// per pivot. A pivot is an event, or the parent between in a chain.
type Reader struct {
	g *graphcache.Graph
	d *deriver
}

// NewReader loads the catalog's bridge terms and event type property once.
func NewReader(g *graphcache.Graph, rules kinship.Rules) (*Reader, error) {
	d, err := newDeriver(g, rules, loadingNodes(g))
	if err != nil {
		return nil, err
	}
	return &Reader{g: g, d: d}, nil
}

// pivot returns the stored result of kind for id, deriving and storing it when
// there is none or it was made under other rules.
func (r *Reader) pivot(kind pivotKind, id []byte, derive func() ([]kinship.Implied, error)) ([]kinship.Implied, error) {
	if v, ok := r.g.Memo(kinshipSlot, id); ok {
		if res, ok := v.(pivotResult); ok && res.kind == kind && res.rules == r.d.rules {
			return res.links, nil
		}
	}
	links, err := derive()
	if err != nil {
		return nil, err
	}
	r.g.SetMemo(kinshipSlot, id, pivotResult{kind: kind, rules: r.d.rules, links: links})
	return links, nil
}

func (r *Reader) event(id []byte) ([]kinship.Implied, error) {
	return r.pivot(pivotEvent, id, func() ([]kinship.Implied, error) { return r.d.eventLinks(id) })
}

func (r *Reader) chain(id []byte) ([]kinship.Implied, error) {
	return r.pivot(pivotChain, id, func() ([]kinship.Implied, error) { return r.d.chainLinks(id, r.event) })
}

// answer returns the stored answer of slot for id, when there is one made under
// the reader's rules.
func (r *Reader) answer(slot graphcache.MemoSlot, id []byte) ([]kinship.Implied, bool) {
	if v, ok := r.g.Memo(slot, id); ok {
		if res, ok := v.(personAnswer); ok && res.rules == r.d.rules {
			return res.links, true
		}
	}
	return nil, false
}

// For returns every spouse, parent and grandparent link the patterns imply
// with id as an end: the links of each event id stands in, then the chain links
// through each parent and child of id. The answer is stored until the next
// commit that names a handle; the slice and its links are the cache's, shared
// and read-only. Links from different pivots differ in Event or Through, so
// none is repeated.
func (r *Reader) For(id []byte) ([]kinship.Implied, error) {
	if links, ok := r.answer(forSlot, id); ok {
		return links, nil
	}
	n, err := r.d.around(id)
	if err != nil || n == nil {
		return nil, err
	}
	out, err := r.d.forLinks(id, r.event, r.chain)
	if err != nil {
		return nil, err
	}
	r.g.SetMemo(forSlot, id, personAnswer{rules: r.d.rules, links: out})
	return out, nil
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
		if err := r.LoadSiblings([][]byte{id}); err != nil {
			return nil, err
		}
		siblings, _ := r.answer(siblingSlot, id)
		for _, imp := range siblings {
			set.AddImplied(imp)
		}
	}
	return set.Links(), nil
}

// LoadSiblings derives and stores the implied siblings of each id that has none
// stored, loading every id's nodes in a fixed number of batches. An id that is
// no node stores nothing.
func (r *Reader) LoadSiblings(ids [][]byte) error {
	var need [][]byte
	for _, id := range ids {
		if _, ok := r.answer(siblingSlot, id); !ok {
			need = append(need, id)
		}
	}
	if len(need) == 0 {
		return nil
	}
	answers, err := r.d.siblingsOf(need, r.event)
	if err != nil {
		return err
	}
	for _, id := range need {
		if links, ok := answers[string(id)]; ok {
			r.g.SetMemo(siblingSlot, id, personAnswer{rules: r.d.rules, links: links})
		}
	}
	return nil
}

// siblingsOf derives the siblings of each id that is a node, in four batched
// loads however many ids and siblings there are: the ids, their neighbors, their
// parents, and the parents' neighbors. An id with no entry is no node (or not 16
// bytes); a node with no siblings has an empty answer. event gives an event's
// links. A parent that is linked but not cached is skipped.
func (d *deriver) siblingsOf(ids [][]byte, event func([]byte) ([]kinship.Implied, error)) (map[string][]kinship.Implied, error) {
	var want [][]byte
	seen := map[string]bool{}
	for _, id := range ids {
		if len(id) == idLen && !seen[string(id)] {
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
	edges := make([][]kinship.Edge, len(nodes))
	parents := make([][][]byte, len(nodes))
	var allParents [][]byte
	for i, n := range nodes {
		atEnd, err := d.eventLinksAt(n, event)
		if err != nil {
			return nil, err
		}
		edges[i] = d.kinEdges(n, atEnd)
		parents[i] = kinship.Parents(n.ID, edges[i], d.terms)
		allParents = append(allParents, parents[i]...)
	}
	if err := d.src.load(allParents); err != nil {
		return nil, err
	}
	var parentNodes []*graphcache.Node
	cachedParents := map[string]*graphcache.Node{}
	for _, q := range allParents {
		if n, ok := d.g.Cached(q); ok && cachedParents[string(q)] == nil {
			cachedParents[string(q)] = n
			parentNodes = append(parentNodes, n)
		}
	}
	if err := d.src.load(neighborsOf(parentNodes)); err != nil {
		return nil, err
	}
	parentEdges := map[string][]kinship.Edge{}
	for _, n := range parentNodes {
		atEnd, err := d.eventLinksAt(n, event)
		if err != nil {
			return nil, err
		}
		parentEdges[string(n.ID)] = d.kinEdges(n, atEnd)
	}
	out := make(map[string][]kinship.Implied, len(nodes))
	for i, n := range nodes {
		all := edges[i]
		for _, q := range parents[i] {
			all = slices.Concat(all, parentEdges[string(q)])
		}
		out[string(n.ID)] = kinship.Siblings(n.ID, all, d.terms)
	}
	return out, nil
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
	seen := map[string]bool{string(id): true}
	var people [][]byte
	add := func(person []byte) {
		if !seen[string(person)] {
			seen[string(person)] = true
			people = append(people, person)
		}
	}
	for _, e := range d.kinEdges(n, out) {
		parent, child, ok := kinship.ParentStep(e, d.terms)
		if !ok {
			continue
		}
		if bytes.Equal(parent, id) {
			add(child)
		} else {
			add(parent)
		}
	}
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
