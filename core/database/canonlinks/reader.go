package canonlinks

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/kinship"
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

// kinshipSlot is where results are stored on the graph. A commit drops a
// result for each handle it touches and for each neighbor of those: an event
// result reads the event's links, a chain result reads the person's links and
// the results of their events, and both are one hop from what they read.
var kinshipSlot graphcache.MemoSlot

func init() {
	kinshipSlot = graphcache.RegisterNodeMemo(graphcache.NodeMemo{
		Name:             "canonlinks.kinship",
		Check:            check,
		Deps:             deps,
		DropOnVocabulary: true,
	})
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
	cached := func(id []byte) (*graphcache.Node, error) {
		n, ok := g.Cached(id)
		if !ok {
			return nil, fmt.Errorf("canonlinks: %x is not cached", id)
		}
		return n, nil
	}
	d, err := newDeriver(g, res.rules, cached)
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

// deriver derives one pivot's links from graph nodes, read through node.
type deriver struct {
	g             *graphcache.Graph
	node          func([]byte) (*graphcache.Node, error)
	rules         kinship.Rules
	terms         segmentcompare.Terms
	eventTypeProp []byte
}

func newDeriver(g *graphcache.Graph, rules kinship.Rules, node func([]byte) (*graphcache.Node, error)) (*deriver, error) {
	terms, err := LoadTerms(g)
	if err != nil {
		return nil, err
	}
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	return &deriver{g: g, node: node, rules: rules, terms: terms, eventTypeProp: v.PropertyID(vocabkey.PropertyEventType, vocabkey.OriginProvenencia)}, nil
}

// edge reads one link of n as an edge that stands for that link alone.
func (d *deriver) edge(n *graphcache.Node, l graphcache.Link) (kinship.Edge, bool) {
	from, to, keyFrom, keyTo, ok := EdgeOf(n.ID, n, l, nil, d.eventTypeProp, d.terms)
	if !ok {
		return kinship.Edge{}, false
	}
	return kinship.Edge{A: from, B: to, KeyA: keyFrom, KeyB: keyTo, Associations: [][]byte{l.Association}}, true
}

// eventLinks is what the event patterns imply over one event's links.
func (d *deriver) eventLinks(id []byte) ([]kinship.Implied, error) {
	n, err := d.node(id)
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
	n, err := d.node(id)
	if err != nil || n == nil {
		return nil, err
	}
	var edges []kinship.Edge
	for _, l := range n.Links {
		if l.BridgeTypeKey == vocabkey.BridgeRelationship {
			if e, ok := d.edge(n, l); ok {
				edges = append(edges, e)
			}
		}
	}
	atEnd, err := d.eventLinksAt(n, eventResult)
	if err != nil {
		return nil, err
	}
	for _, imp := range atEnd {
		if imp.Pattern == kinship.PatternBirthParents {
			edges = append(edges, imp.Edge)
		}
	}
	return kinship.Derive(edges, d.terms, d.rules), nil
}

// Reader reads the links the kinship patterns imply, from one cached result
// per pivot. A pivot is an event, or the parent between in a chain.
type Reader struct {
	g *graphcache.Graph
	d *deriver
}

// NewReader loads the catalog's bridge terms and event type property once.
func NewReader(g *graphcache.Graph, rules kinship.Rules) (*Reader, error) {
	d, err := newDeriver(g, rules, g.Node)
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

// loadAround loads a person and their neighbors.
func (r *Reader) loadAround(id []byte) (*graphcache.Node, error) {
	n, err := r.g.Node(id)
	if err != nil || n == nil {
		return nil, err
	}
	ids := make([][]byte, 0, len(n.Links))
	for _, l := range n.Links {
		ids = append(ids, l.Neighbor)
	}
	return n, r.g.Nodes(ids)
}

// For returns every spouse, parent and grandparent link the patterns imply
// with id as an end: the links of each event id stands in, then the chain links
// through each parent and child of id. The links are the cache's, shared and
// read-only; the slice is new. Links from different pivots differ in Event or
// Through, so none is repeated.
func (r *Reader) For(id []byte) ([]kinship.Implied, error) {
	n, err := r.loadAround(id)
	if err != nil || n == nil {
		return nil, err
	}
	atEnd := func(imp kinship.Implied) bool { return bytes.Equal(imp.A, id) || bytes.Equal(imp.B, id) }
	out, err := r.d.eventLinksAt(n, r.event)
	if err != nil {
		return nil, err
	}
	var births []kinship.Implied
	for _, imp := range out {
		if imp.Pattern == kinship.PatternBirthParents {
			births = append(births, imp)
		}
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
	for _, l := range n.Links {
		if l.BridgeTypeKey != vocabkey.BridgeRelationship {
			continue
		}
		e, ok := r.d.edge(n, l)
		if !ok {
			continue
		}
		parent, child, ok := kinship.ParentStep(e, r.d.terms)
		if !ok {
			continue
		}
		if bytes.Equal(parent, id) {
			add(child)
		} else {
			add(parent)
		}
	}
	for _, imp := range births {
		if bytes.Equal(imp.A, id) {
			add(imp.B)
		} else {
			add(imp.A)
		}
	}
	for _, m := range people {
		if _, err := r.loadAround(m); err != nil {
			return nil, err
		}
		links, err := r.chain(m)
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
