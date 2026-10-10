package conclusionheaders

import (
	"bytes"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/datevalues"
)

// placeLink is one filed place_relationship edge with its membership dates.
type placeLink struct {
	from, to    []byte
	association []byte
	start, end  *datevalues.Value
}

// placeGraph is the part_of / succeeded_by neighborhood needed to compose
// chains for a set of Places at a query date.
type placeGraph struct {
	byID   map[string]PlaceHeader
	partOf []placeLink // from = part → to = whole
	succ   []placeLink // from = predecessor → to = successor
	period map[string]autoreconcile.Window
}

type assocDates struct {
	start, end *datevalues.Value
}

func (g *placeGraph) periodOf(id []byte) autoreconcile.Window {
	if w, ok := g.period[string(id)]; ok {
		return w
	}
	return autoreconcile.Always()
}

func (g *placeGraph) linkHolds(query autoreconcile.Window, l placeLink) bool {
	mem, ok := autoreconcile.LinkMembership(l.start, l.end, g.periodOf(l.from), g.periodOf(l.to))
	if !ok {
		return false
	}
	return autoreconcile.HoldsAt(query, mem)
}

func queryWindow(at *datevalues.Value) autoreconcile.Window {
	if at == nil {
		return autoreconcile.Always()
	}
	if w, ok := autoreconcile.WindowOfDate(*at); ok {
		return w
	}
	return autoreconcile.Always()
}

// ImmediateParents are the part_of wholes that hold (or are ambiguous) at at.
func (g *placeGraph) ImmediateParents(placeID []byte, at *datevalues.Value) [][]byte {
	q := queryWindow(at)
	seen := map[string]bool{}
	var out [][]byte
	for _, l := range g.partOf {
		if !bytes.Equal(l.from, placeID) || !g.linkHolds(q, l) {
			continue
		}
		if seen[string(l.to)] {
			continue
		}
		seen[string(l.to)] = true
		out = append(out, l.to)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return g.displayName(out[i]) < g.displayName(out[j])
	})
	return out
}

// ImmediateParts are the part_of parts that hold at at.
func (g *placeGraph) ImmediateParts(placeID []byte, at *datevalues.Value) [][]byte {
	q := queryWindow(at)
	seen := map[string]bool{}
	var out [][]byte
	for _, l := range g.partOf {
		if !bytes.Equal(l.to, placeID) || !g.linkHolds(q, l) {
			continue
		}
		if seen[string(l.from)] {
			continue
		}
		seen[string(l.from)] = true
		out = append(out, l.from)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return g.displayName(out[i]) < g.displayName(out[j])
	})
	return out
}

// Successors are succeeded_by targets (dates do not filter succession for
// lineage reads; membership spans never build a display chain).
func (g *placeGraph) Successors(placeID []byte) [][]byte {
	return g.succNeighbors(placeID, true)
}

// Predecessors are succeeded_by sources.
func (g *placeGraph) Predecessors(placeID []byte) [][]byte {
	return g.succNeighbors(placeID, false)
}

func (g *placeGraph) succNeighbors(placeID []byte, forward bool) [][]byte {
	seen := map[string]bool{}
	var out [][]byte
	for _, l := range g.succ {
		var other []byte
		if forward && bytes.Equal(l.from, placeID) {
			other = l.to
		} else if !forward && bytes.Equal(l.to, placeID) {
			other = l.from
		} else {
			continue
		}
		if seen[string(other)] {
			continue
		}
		seen[string(other)] = true
		out = append(out, other)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return g.displayName(out[i]) < g.displayName(out[j])
	})
	return out
}

func (g *placeGraph) displayName(id []byte) string {
	h, ok := g.byID[string(id)]
	if !ok {
		return ""
	}
	if t := strings.TrimSpace(placeTitle(h)); t != "" {
		return t
	}
	return h.Entity.Ref
}

// ParentChain walks hierarchical parents at at. One unambiguous path is
// nearest-first ancestors. Several parents at a level are every candidate
// (candidates=true; no nature preference). Cycle-guarded.
func (g *placeGraph) ParentChain(placeID []byte, at *datevalues.Value) (names []string, candidates bool) {
	return g.parentChain(placeID, at, map[string]bool{string(placeID): true})
}

// ParentChainNames is ParentChain's names only.
func (g *placeGraph) ParentChainNames(placeID []byte, at *datevalues.Value) []string {
	names, _ := g.ParentChain(placeID, at)
	return names
}

func (g *placeGraph) parentChain(placeID []byte, at *datevalues.Value, seen map[string]bool) ([]string, bool) {
	parents := g.ImmediateParents(placeID, at)
	if len(parents) == 0 {
		return nil, false
	}
	if len(parents) > 1 {
		names := make([]string, 0, len(parents))
		for _, p := range parents {
			if n := g.displayName(p); n != "" {
				names = append(names, n)
			}
		}
		return names, true
	}
	p := parents[0]
	if seen[string(p)] {
		return nil, false
	}
	seen[string(p)] = true
	name := g.displayName(p)
	rest, _ := g.parentChain(p, at, seen)
	if name == "" {
		return rest, false
	}
	return append([]string{name}, rest...), false
}

// isAncestor reports whether ancestor reaches descendant via holding
// part_of links at at (cycle-guarded BFS upward from descendant).
func (g *placeGraph) isAncestor(ancestor, descendant []byte, at *datevalues.Value) bool {
	if bytes.Equal(ancestor, descendant) {
		return false
	}
	seen := map[string]bool{string(descendant): true}
	queue := [][]byte{append([]byte(nil), descendant...)}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, p := range g.ImmediateParents(cur, at) {
			if bytes.Equal(p, ancestor) {
				return true
			}
			if seen[string(p)] {
				continue
			}
			seen[string(p)] = true
			queue = append(queue, p)
		}
	}
	return false
}

// FoldLocations collapses Places that share a part_of ancestry at at into
// one HeaderPlace (the deepest leaf). Disconnected Places stay competing.
// Succession never connects components.
func (g *placeGraph) FoldLocations(places []HeaderPlace, at *datevalues.Value) []HeaderPlace {
	if len(places) <= 1 {
		out := make([]HeaderPlace, len(places))
		copy(out, places)
		for i := range out {
			out[i].Parents, out[i].ParentsAreCandidates = g.ParentChain(out[i].Entity.ID, at)
		}
		return out
	}
	n := len(places)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			a, b := places[i].Entity.ID, places[j].Entity.ID
			if g.isAncestor(a, b, at) || g.isAncestor(b, a, at) {
				union(i, j)
			}
		}
	}
	groups := map[int][]int{}
	for i := range places {
		r := find(i)
		groups[r] = append(groups[r], i)
	}
	var out []HeaderPlace
	for _, idxs := range groups {
		leaf := idxs[0]
		for _, i := range idxs[1:] {
			// Prefer the place that is a descendant of the current leaf.
			if g.isAncestor(places[leaf].Entity.ID, places[i].Entity.ID, at) {
				leaf = i
			}
		}
		h := places[leaf]
		h.Parents, h.ParentsAreCandidates = g.ParentChain(h.Entity.ID, at)
		out = append(out, h)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Entity.Ref < out[j].Entity.Ref
	})
	return out
}

// chainQueryDate is today, or the Place's end when today is after its period
// (D7: show the last chain for a place whose period has ended).
func chainQueryDate(h PlaceHeader, today datevalues.Value) datevalues.Value {
	if h.EndDate == nil {
		return today
	}
	tw, tok := autoreconcile.WindowOfDate(today)
	ew, eok := autoreconcile.WindowOfDate(*h.EndDate)
	if !tok || !eok || tw.Lo == nil || ew.Hi == nil {
		return today
	}
	if *tw.Lo > *ew.Hi {
		return *h.EndDate
	}
	return today
}

// placeNeighborhood is the containment and succession neighborhood of ids,
// read from the open catalog's graph.
func placeNeighborhood(c *database.Catalog, ids [][]byte) (*placeGraph, error) {
	if c == nil || c.Graph() == nil {
		return &placeGraph{byID: map[string]PlaceHeader{}, period: map[string]autoreconcile.Window{}}, nil
	}
	g := c.Graph()
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	return buildPlaceGraph(g, v, ids)
}

// AttachPlaceRelationships fills PartOf / Contains / Predecessors /
// Successors on a Place detail header (every part_of parent with its
// membership span; children that hold today; succession both ways).
func AttachPlaceRelationships(c *database.Catalog, h *PlaceHeader) error {
	if h == nil || len(h.Entity.ID) == 0 {
		return nil
	}
	g, err := placeNeighborhood(c, [][]byte{h.Entity.ID})
	if err != nil {
		return err
	}
	g.byID[string(h.Entity.ID)] = *h
	g.period[string(h.Entity.ID)] = autoreconcile.PeriodWindow(h.StartDate, h.EndDate)
	today := TodayDate()

	seenPart := map[string]bool{}
	for _, l := range g.partOf {
		if !bytes.Equal(l.from, h.Entity.ID) {
			continue
		}
		if _, ok := autoreconcile.LinkMembership(l.start, l.end, g.periodOf(l.from), g.periodOf(l.to)); !ok {
			continue
		}
		key := string(l.to)
		if seenPart[key] {
			continue
		}
		seenPart[key] = true
		rel := g.relationship(l.to, RelPartOf)
		rel.StartDate, rel.EndDate = membershipDates(l, g.byID[string(l.from)], g.byID[string(l.to)])
		h.PartOf = append(h.PartOf, rel)
	}
	sort.SliceStable(h.PartOf, func(i, j int) bool {
		return h.PartOf[i].Title < h.PartOf[j].Title
	})

	for _, id := range g.ImmediateParts(h.Entity.ID, &today) {
		rel := g.relationship(id, RelContains)
		for _, l := range g.partOf {
			if bytes.Equal(l.to, h.Entity.ID) && bytes.Equal(l.from, id) {
				rel.StartDate, rel.EndDate = membershipDates(l, g.byID[string(l.from)], g.byID[string(l.to)])
				break
			}
		}
		h.Contains = append(h.Contains, rel)
	}

	for _, id := range g.Predecessors(h.Entity.ID) {
		rel := g.relationship(id, RelPredecessor)
		if other, ok := g.byID[string(id)]; ok {
			rel.StartDate, rel.EndDate = other.StartDate, other.EndDate
		}
		h.Predecessors = append(h.Predecessors, rel)
	}
	for _, id := range g.Successors(h.Entity.ID) {
		rel := g.relationship(id, RelSuccessor)
		if other, ok := g.byID[string(id)]; ok {
			rel.StartDate, rel.EndDate = other.StartDate, other.EndDate
		}
		h.Successors = append(h.Successors, rel)
	}
	return nil
}

func (g *placeGraph) relationship(id []byte, kind string) PlaceRelationship {
	title := g.displayName(id)
	rel := PlaceRelationship{Kind: kind, Title: title}
	if h, ok := g.byID[string(id)]; ok {
		rel.Entity = h.Entity
	} else {
		rel.Entity = canonicalentities.Entity{ID: append([]byte(nil), id...)}
	}
	return rel
}

// membershipDates is the span a hierarchical link shows: the link's own
// dates when present, else the overlap of the two places' periods.
func membershipDates(l placeLink, from, to PlaceHeader) (start, end *datevalues.Value) {
	if l.start != nil || l.end != nil {
		return l.start, l.end
	}
	return laterStart(from.StartDate, to.StartDate), earlierEnd(from.EndDate, to.EndDate)
}

func yearOf(d *datevalues.Value) *int {
	if d == nil || d.StartYear == nil {
		return nil
	}
	return d.StartYear
}

func laterStart(a, b *datevalues.Value) *datevalues.Value {
	ya, yb := yearOf(a), yearOf(b)
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	case ya != nil && yb != nil && *yb > *ya:
		return b
	default:
		return a
	}
}

func earlierEnd(a, b *datevalues.Value) *datevalues.Value {
	ya, yb := yearOf(a), yearOf(b)
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	case ya != nil && yb != nil && *yb < *ya:
		return b
	default:
		return a
	}
}

// ParentsAtDate is the hierarchical parent display names for placeID at at
// (nearest first; several parents → every candidate).
func ParentsAtDate(c *database.Catalog, placeID []byte, at datevalues.Value) ([]string, error) {
	g, err := placeNeighborhood(c, [][]byte{placeID})
	if err != nil {
		return nil, err
	}
	return g.ParentChainNames(placeID, &at), nil
}

// PartsAtDate is the display names of Places that are part_of placeID at at.
func PartsAtDate(c *database.Catalog, placeID []byte, at datevalues.Value) ([]string, error) {
	g, err := placeNeighborhood(c, [][]byte{placeID})
	if err != nil {
		return nil, err
	}
	ids := g.ImmediateParts(placeID, &at)
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if n := g.displayName(id); n != "" {
			out = append(out, n)
		}
	}
	return out, nil
}

// SuccessionNames is predecessor and successor display names via
// succeeded_by (dates do not filter; never used for display chains).
func SuccessionNames(c *database.Catalog, placeID []byte) (predecessors, successors []string, err error) {
	g, err := placeNeighborhood(c, [][]byte{placeID})
	if err != nil {
		return nil, nil, err
	}
	for _, id := range g.Predecessors(placeID) {
		if n := g.displayName(id); n != "" {
			predecessors = append(predecessors, n)
		}
	}
	for _, id := range g.Successors(placeID) {
		if n := g.displayName(id); n != "" {
			successors = append(successors, n)
		}
	}
	return predecessors, successors, nil
}
