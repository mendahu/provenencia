package conclusionsummaries

import (
	"bytes"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/hops"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// summaryChainDepth is how far a place row's parent chain climbs. The place
// neighborhood, closure, ParentChain and isAncestor (so FoldLocations) use the
// same bound, so a row composed in a batch shows no ancestor its recorded
// inputs lack.
const summaryChainDepth = 8

// maxInputKeysPerCall caps the place-chain keys one RowsForSearch call
// records across all its rows, counted before the keys are copied. On the large
// fixture the largest 500-row call counts 1,000 and holds 3,500 keys in all (a
// whole-fixture call counts 16,102); the cap is above 50 times the larger. Only a
// crafted catalog reaches it, such as a dense part_of mesh whose every row
// copies its whole place closure. Over the cap the call counts nothing: no
// row of that batch has inputs, so its index entries are empty and an edit to
// its places does not invalidate it. S14-42 inherits that miss for a crafted
// catalog only.
const maxInputKeysPerCall = 250000

// inputKeyCap is maxInputKeysPerCall; a test lowers it.
var inputKeyCap = maxInputKeysPerCall

// placeLink is one filed place_relationship edge with its membership dates.
type placeLink struct {
	from, to    []byte
	association []byte
	start, end  *datevalue.Value
}

// placeGraph is the part_of / succeeded_by neighborhood needed to compose
// chains for a set of Places at a query date. Links are indexed by endpoint
// so a parent question does not scan every other place's edges.
type placeGraph struct {
	byID     map[string]PlaceSummary
	partFrom map[string][]placeLink // part id → wholes
	partTo   map[string][]placeLink // whole id → parts
	succFrom map[string][]placeLink // predecessor id → successors
	succTo   map[string][]placeLink // successor id → predecessors
	period   map[string]valuecompare.Window
	// inputs are the keys the build recorded reading each place it expanded,
	// by place id: its values, its part_of links at the part end, and those
	// links' associations.
	inputs map[string][]graphcache.InputKey
}

func newPlaceGraph() *placeGraph {
	return &placeGraph{
		byID:     map[string]PlaceSummary{},
		partFrom: map[string][]placeLink{},
		partTo:   map[string][]placeLink{},
		succFrom: map[string][]placeLink{},
		succTo:   map[string][]placeLink{},
		period:   map[string]valuecompare.Window{},
		inputs:   map[string][]graphcache.InputKey{},
	}
}

// reach visits, once each, the places a walk up part_of from seeds expands:
// the same rounds buildPlaceGraph runs, over the links it kept.
func (g *placeGraph) reach(seeds [][]byte, visit func(id []byte)) {
	seeds = database.UniqueBlobIDs(seeds)
	frontier := append([][]byte(nil), seeds...)
	seen := map[string]bool{}
	for _, id := range seeds {
		seen[string(id)] = true
	}
	for round := 0; round < summaryChainDepth && len(frontier) > 0; round++ {
		var next [][]byte
		for _, id := range frontier {
			visit(id)
			for _, l := range g.partFrom[string(id)] {
				if !seen[string(l.to)] {
					seen[string(l.to)] = true
					next = append(next, l.to)
				}
			}
		}
		frontier = next
	}
}

// closure is the union of the keys recorded for every place a walk up part_of
// from seeds expands. It records nothing of its own, so a read added to the
// build reaches every row that walks through it.
func (g *placeGraph) closure(seeds [][]byte) []graphcache.InputKey {
	var keys []graphcache.InputKey
	g.reach(seeds, func(id []byte) { keys = append(keys, g.inputs[string(id)]...) })
	return keys
}

// closureSize is len(closure(seeds)) without building it.
func (g *placeGraph) closureSize(seeds [][]byte) int {
	n := 0
	g.reach(seeds, func(id []byte) { n += len(g.inputs[string(id)]) })
	return n
}

// addChains follows each event's locations through in and records the closure
// of their places, through the call's cap.
func (g *placeGraph) addChains(in *graphcache.InputScope, rec *inputBudget, eventIDs [][]byte) error {
	seeds, err := placesOfEvents(in, eventIDs)
	if err != nil {
		return err
	}
	rec.addClosure(in, g, seeds)
	return nil
}

// inputBudget counts the keys one RowsForSearch call records for the place
// chains of all its rows. Over inputKeyCap the call counts nothing: it stops
// recording and hands back no inputs for any row. A nil budget is a read that
// records nothing: every method then does nothing and its scopes record
// nothing.
type inputBudget struct {
	used int
	over bool
}

// scope is the scope a row for dependent is composed through: recording with a
// budget, and recording nothing without one.
func (b *inputBudget) scope(g *graphcache.Graph, dependent []byte) *graphcache.InputScope {
	if b == nil {
		return g.NoInputs()
	}
	return g.Inputs(dependent)
}

// keep files the keys in recorded for the row id, when the call records.
func (b *inputBudget) keep(into map[string][]graphcache.InputKey, id []byte, in *graphcache.InputScope) {
	if b != nil {
		into[string(id)] = in.Keys()
	}
}

// addClosure records the closure of seeds on in, unless that would take the
// call past the cap. It counts the closure before it builds it.
func (b *inputBudget) addClosure(in *graphcache.InputScope, pg *placeGraph, seeds [][]byte) {
	if b == nil || b.over {
		return
	}
	if b.used += pg.closureSize(seeds); b.used > inputKeyCap {
		b.over = true
		return
	}
	in.Add(pg.closure(seeds)...)
}

type assocDates struct {
	start, end *datevalue.Value
}

func (g *placeGraph) periodOf(id []byte) valuecompare.Window {
	if w, ok := g.period[string(id)]; ok {
		return w
	}
	return valuecompare.Always()
}

func (g *placeGraph) linkHolds(query valuecompare.Window, l placeLink) bool {
	mem, ok := linkMembership(l.start, l.end, g.periodOf(l.from), g.periodOf(l.to))
	if !ok {
		return false
	}
	return query.Overlaps(mem)
}

func queryWindow(at *datevalue.Value) valuecompare.Window {
	if at == nil {
		return valuecompare.Always()
	}
	if w, ok := valuecompare.DateWindow(*at); ok {
		return w
	}
	return valuecompare.Always()
}

// ImmediateParents are the part_of wholes that hold (or are ambiguous) at at.
func (g *placeGraph) ImmediateParents(placeID []byte, at *datevalue.Value) [][]byte {
	q := queryWindow(at)
	seen := map[string]bool{}
	var out [][]byte
	for _, l := range g.partFrom[string(placeID)] {
		if !g.linkHolds(q, l) {
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
func (g *placeGraph) ImmediateParts(placeID []byte, at *datevalue.Value) [][]byte {
	q := queryWindow(at)
	seen := map[string]bool{}
	var out [][]byte
	for _, l := range g.partTo[string(placeID)] {
		if !g.linkHolds(q, l) {
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
	links := g.succTo[string(placeID)]
	if forward {
		links = g.succFrom[string(placeID)]
	}
	for _, l := range links {
		other := l.from
		if forward {
			other = l.to
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
func (g *placeGraph) ParentChain(placeID []byte, at *datevalue.Value) (names []string, candidates bool) {
	return g.parentChain(placeID, at, map[string]bool{string(placeID): true}, 0)
}

// ParentChainNames is ParentChain's names only.
func (g *placeGraph) ParentChainNames(placeID []byte, at *datevalue.Value) []string {
	names, _ := g.ParentChain(placeID, at)
	return names
}

// parentChain shows ancestors only as far as the build expanded places: the
// build reads the shells of summaryChainDepth places up from a seed, the seed
// first, so the highest ancestor with a name is summaryChainDepth-1 levels
// above it. A name the build did not read must not appear because another row
// of the batch happened to read it.
func (g *placeGraph) parentChain(placeID []byte, at *datevalue.Value, seen map[string]bool, depth int) ([]string, bool) {
	if depth+1 >= summaryChainDepth {
		return nil, false
	}
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
	rest, _ := g.parentChain(p, at, seen, depth+1)
	if name == "" {
		return rest, false
	}
	return append([]string{name}, rest...), false
}

// isAncestor reports whether ancestor reaches descendant via holding
// part_of links at at (cycle-guarded BFS upward from descendant), through
// links of at most summaryChainDepth places: the ones the build expanded.
func (g *placeGraph) isAncestor(ancestor, descendant []byte, at *datevalue.Value) bool {
	if bytes.Equal(ancestor, descendant) {
		return false
	}
	seen := map[string]bool{string(descendant): true}
	level := [][]byte{append([]byte(nil), descendant...)}
	for depth := 0; depth < summaryChainDepth && len(level) > 0; depth++ {
		var next [][]byte
		for _, cur := range level {
			for _, p := range g.ImmediateParents(cur, at) {
				if bytes.Equal(p, ancestor) {
					return true
				}
				if seen[string(p)] {
					continue
				}
				seen[string(p)] = true
				next = append(next, p)
			}
		}
		level = next
	}
	return false
}

// FoldLocations collapses Places that lie on one part_of chain at at into
// the deepest of them: "Ontario" and "Toronto, Ontario" are one reading,
// shown as Toronto. A Place is dropped only when another listed Place is
// below it. Places that are not on one chain stay competing, even when a
// vaguer listed Place is above both: Toronto and Hamilton under Ontario are
// two readings, not one. Succession never connects Places.
func (g *placeGraph) FoldLocations(places []ChainedPlace, at *datevalue.Value) []ChainedPlace {
	var out []ChainedPlace
	seen := map[string]bool{}
	for i := range places {
		id := places[i].Entity.ID
		if seen[string(id)] {
			continue
		}
		covered := false
		for j := range places {
			other := places[j].Entity.ID
			if bytes.Equal(other, id) {
				continue
			}
			// A link back up (a cycle at this date) is not "below".
			if g.isAncestor(id, other, at) && !g.isAncestor(other, id, at) {
				covered = true
				break
			}
		}
		if covered {
			continue
		}
		seen[string(id)] = true
		h := places[i]
		h.Parents, h.ParentsAreCandidates = g.ParentChain(id, at)
		out = append(out, h)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Entity.Ref < out[j].Entity.Ref
	})
	return out
}

// chainQueryDate is today, or the Place's end when today is after its period
// (D7: show the last chain for a place whose period has ended).
func chainQueryDate(h PlaceSummary, today datevalue.Value) datevalue.Value {
	if h.EndDate == nil {
		return today
	}
	tw, tok := valuecompare.DateWindow(today)
	ew, eok := valuecompare.DateWindow(*h.EndDate)
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
		return newPlaceGraph(), nil
	}
	g := c.Graph()
	v, err := vocabulary(g)
	if err != nil {
		return nil, err
	}
	return buildPlaceGraph(g, v, ids, nil)
}

// AttachPlaceRelationships fills PartOf / Contains / Predecessors /
// Successors on a Place detail header (every part_of parent with its
// membership span; children that hold today; succession both ways).
func AttachPlaceRelationships(c *database.Catalog, h *PlaceSummary) error {
	if h == nil || len(h.Entity.ID) == 0 {
		return nil
	}
	g, err := placeStep(c, h.Entity.ID, hops.ParentsOfPlace, hops.PartsOfPlace, hops.PredecessorsOfPlace, hops.SuccessorsOfPlace)
	if err != nil {
		return err
	}
	g.byID[string(h.Entity.ID)] = *h
	g.period[string(h.Entity.ID)] = periodWindow(h.StartDate, h.EndDate)
	today := TodayDate()

	seenPart := map[string]bool{}
	for _, l := range g.partFrom[string(h.Entity.ID)] {
		if _, ok := linkMembership(l.start, l.end, g.periodOf(l.from), g.periodOf(l.to)); !ok {
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
		for _, l := range g.partTo[string(h.Entity.ID)] {
			if bytes.Equal(l.from, id) {
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
func membershipDates(l placeLink, from, to PlaceSummary) (start, end *datevalue.Value) {
	if l.start != nil || l.end != nil {
		return l.start, l.end
	}
	return laterStart(from.StartDate, to.StartDate), earlierEnd(from.EndDate, to.EndDate)
}

// laterStart is the later of two starts by day. A nil side yields the other;
// an unreadable date (phrase-only) never beats a readable one, and an open
// start (BEF) loses to a set one. A tie keeps a.
func laterStart(a, b *datevalue.Value) *datevalue.Value {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	wa, okA := valuecompare.DateWindow(*a)
	wb, okB := valuecompare.DateWindow(*b)
	switch {
	case !okB:
		return a
	case !okA:
		return b
	case wb.Lo != nil && (wa.Lo == nil || *wb.Lo > *wa.Lo):
		return b
	default:
		return a
	}
}

// earlierEnd is the earlier of two ends by day, mirroring laterStart on the
// last day.
func earlierEnd(a, b *datevalue.Value) *datevalue.Value {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	wa, okA := valuecompare.DateWindow(*a)
	wb, okB := valuecompare.DateWindow(*b)
	switch {
	case !okB:
		return a
	case !okA:
		return b
	case wb.Hi != nil && (wa.Hi == nil || *wb.Hi < *wa.Hi):
		return b
	default:
		return a
	}
}

// periodWindow is a Place's existence span from optional start and end
// dates. Missing sides are open; both missing is Always.
func periodWindow(start, end *datevalue.Value) valuecompare.Window {
	w := valuecompare.Always()
	if start != nil {
		if sw, ok := valuecompare.DateWindow(*start); ok {
			w.Lo = sw.Lo
		}
	}
	if end != nil {
		if ew, ok := valuecompare.DateWindow(*end); ok {
			w.Hi = ew.Hi
		}
	}
	return w
}

// linkMembership is when a hierarchical place relationship holds.
// An undated link (nil start and end) holds on the overlap of the two
// places' periods. A dated link holds inside its own span and both periods.
func linkMembership(linkStart, linkEnd *datevalue.Value, fromPeriod, toPeriod valuecompare.Window) (valuecompare.Window, bool) {
	base, ok := fromPeriod.Intersect(toPeriod)
	if !ok {
		return valuecompare.Window{}, false
	}
	if linkStart == nil && linkEnd == nil {
		return base, true
	}
	return base.Intersect(periodWindow(linkStart, linkEnd))
}

// ParentsAtDate is the hierarchical parent display names for placeID at at
// (nearest first; several parents → every candidate).
func ParentsAtDate(c *database.Catalog, placeID []byte, at datevalue.Value) ([]string, error) {
	g, err := placeNeighborhood(c, [][]byte{placeID})
	if err != nil {
		return nil, err
	}
	return g.ParentChainNames(placeID, &at), nil
}

// PartsAtDate is the display names of Places that are part_of placeID at at.
func PartsAtDate(c *database.Catalog, placeID []byte, at datevalue.Value) ([]string, error) {
	g, err := placeStep(c, placeID, hops.PartsOfPlace)
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
	g, err := placeStep(c, placeID, hops.PredecessorsOfPlace, hops.SuccessorsOfPlace)
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
