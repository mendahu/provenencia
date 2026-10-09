package conclusionheaders

import (
	"bytes"
	"sort"
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/canonicalgraph"
	"github.com/mendahu/provenencia/core/database/datevalues"
)

// placeLink is one filed place_relationship edge with its membership dates.
type placeLink struct {
	from, to    []byte
	association []byte
	start, end  *datevalues.Value
}

// placeGraph is the part_of / succeeded_by neighborhood needed to compose
// chains for a set of Places at a query date. Links are indexed by endpoint,
// so a lookup reads one Place's links, not every link loaded.
type placeGraph struct {
	byID   map[string]PlaceHeader
	partOf []placeLink // from = part → to = whole
	succ   []placeLink // from = predecessor → to = successor
	period map[string]autoreconcile.Window

	partUp   map[string][]int // part id → its partOf links
	partDown map[string][]int // whole id → its partOf links
	succOut  map[string][]int // predecessor id → its succ links
	succIn   map[string][]int // successor id → its succ links
	linkSeen map[string]bool
	chains   map[string]parentChainResult // ParentChain memo, by place and window
}

type parentChainResult struct {
	names      []string
	candidates bool
}

func newPlaceGraph() *placeGraph {
	return &placeGraph{
		byID:     map[string]PlaceHeader{},
		period:   map[string]autoreconcile.Window{},
		partUp:   map[string][]int{},
		partDown: map[string][]int{},
		succOut:  map[string][]int{},
		succIn:   map[string][]int{},
		linkSeen: map[string]bool{},
		chains:   map[string]parentChainResult{},
	}
}

// placeReach is how much of the hierarchy a caller needs.
type placeReach int

const (
	// reachAncestors climbs part_of to the top: enough for chains and
	// folding, which only ever look up.
	reachAncestors placeReach = iota
	// reachDetail is the ancestors plus one hop of parts and of succession
	// both ways, for a Place page.
	reachDetail
)

// maxClimb bounds the part_of climb; filing refuses cycles, so a real
// hierarchy ends long before it.
const maxClimb = 64

// loadPlaceGraph loads headers, periods, and part_of links for seeds and
// every ancestor (reachAncestors), plus one hop of parts and succession for
// reachDetail. Each step is one batched walk over the new Places, never
// their siblings or descendants.
func loadPlaceGraph(q Querier, seeds [][]byte, reach placeReach) (*placeGraph, error) {
	seeds = database.UniqueBlobIDs(seeds)
	g := newPlaceGraph()
	if len(seeds) == 0 {
		return g, nil
	}
	if err := g.ensurePlaces(q, seeds); err != nil {
		return nil, err
	}
	if reach == reachDetail {
		partRev, err := canonicalgraph.Walk(q, canonicalgraph.PartsOfPlace, seeds)
		if err != nil {
			return nil, err
		}
		succEdges, err := canonicalgraph.Walk(q, canonicalgraph.SuccessorsOfPlace, seeds)
		if err != nil {
			return nil, err
		}
		predEdges, err := canonicalgraph.Walk(q, canonicalgraph.PredecessorsOfPlace, seeds)
		if err != nil {
			return nil, err
		}
		dates, err := loadAssociationDates(q, associationIDs(partRev, succEdges, predEdges))
		if err != nil {
			return nil, err
		}
		g.addLinks(false, reverseEdges(partRev), dates)
		g.addLinks(true, succEdges, dates)
		g.addLinks(true, reverseEdges(predEdges), dates)
		var near [][]byte
		for _, e := range append(append(partRev, succEdges...), predEdges...) {
			near = append(near, e.To)
		}
		if err := g.ensurePlaces(q, database.UniqueBlobIDs(near)); err != nil {
			return nil, err
		}
	}

	seen := map[string]bool{}
	for _, id := range seeds {
		seen[string(id)] = true
	}
	frontier := seeds
	for climb := 0; climb < maxClimb && len(frontier) > 0; climb++ {
		edges, err := canonicalgraph.Walk(q, canonicalgraph.ParentsOfPlace, frontier)
		if err != nil {
			return nil, err
		}
		dates, err := loadAssociationDates(q, associationIDs(edges))
		if err != nil {
			return nil, err
		}
		g.addLinks(false, edges, dates)
		var next [][]byte
		for _, e := range edges {
			if !seen[string(e.To)] {
				seen[string(e.To)] = true
				next = append(next, e.To)
			}
		}
		if err := g.ensurePlaces(q, next); err != nil {
			return nil, err
		}
		frontier = next
	}
	return g, nil
}

func (g *placeGraph) ensurePlaces(q Querier, ids [][]byte) error {
	missing := make([][]byte, 0, len(ids))
	for _, id := range ids {
		if _, ok := g.byID[string(id)]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	// PlacesByIDs attaches today's chains; avoid recursion by scanning raw.
	headers, err := placesWithoutChains(q, missing)
	if err != nil {
		return err
	}
	for _, h := range headers {
		g.setPlace(h)
	}
	return nil
}

// setPlace records a Place's header and period.
func (g *placeGraph) setPlace(h PlaceHeader) {
	g.byID[string(h.Entity.ID)] = h
	g.period[string(h.Entity.ID)] = autoreconcile.PeriodWindow(h.StartDate, h.EndDate)
}

// addLinks adds walked edges as succession (succession) or part_of links,
// once each.
func (g *placeGraph) addLinks(succession bool, edges []canonicalgraph.Edge, dates map[string]assocDates) {
	for _, e := range edges {
		key := string(e.From) + "\x00" + string(e.To) + "\x00" + string(e.Association)
		if succession {
			key = "s\x00" + key
		}
		if g.linkSeen[key] {
			continue
		}
		g.linkSeen[key] = true
		d := dates[string(e.Association)]
		l := placeLink{
			from: append([]byte(nil), e.From...), to: append([]byte(nil), e.To...),
			association: append([]byte(nil), e.Association...),
			start:       d.start, end: d.end,
		}
		if succession {
			g.succ = append(g.succ, l)
			i := len(g.succ) - 1
			g.succOut[string(l.from)] = append(g.succOut[string(l.from)], i)
			g.succIn[string(l.to)] = append(g.succIn[string(l.to)], i)
			continue
		}
		g.partOf = append(g.partOf, l)
		i := len(g.partOf) - 1
		g.partUp[string(l.from)] = append(g.partUp[string(l.from)], i)
		g.partDown[string(l.to)] = append(g.partDown[string(l.to)], i)
	}
}

func reverseEdges(edges []canonicalgraph.Edge) []canonicalgraph.Edge {
	out := make([]canonicalgraph.Edge, len(edges))
	for i, e := range edges {
		out[i] = canonicalgraph.Edge{
			From: e.To, To: e.From, Association: e.Association, AssociationRef: e.AssociationRef, ToRef: e.ToRef,
		}
	}
	return out
}

func associationIDs(groups ...[]canonicalgraph.Edge) [][]byte {
	seen := map[string]bool{}
	var out [][]byte
	for _, edges := range groups {
		for _, e := range edges {
			if !seen[string(e.Association)] {
				seen[string(e.Association)] = true
				out = append(out, e.Association)
			}
		}
	}
	return out
}

type assocDates struct {
	start, end *datevalues.Value
}

const sqlAssocDates = `SELECT e.id, sd.value_date, ed.value_date
FROM canonical_entities e
LEFT JOIN properties sp ON sp.key = 'start_date' AND sp.origin = 'provenencia'
LEFT JOIN auto_reconciler_values sd
	ON sd.entity_id = e.id AND sd.property_id = sp.id AND sd.rank = 1 AND sd.reason = 'kept'
LEFT JOIN properties ep ON ep.key = 'end_date' AND ep.origin = 'provenencia'
LEFT JOIN auto_reconciler_values ed
	ON ed.entity_id = e.id AND ed.property_id = ep.id AND ed.rank = 1 AND ed.reason = 'kept'
WHERE e.id IN (`

func loadAssociationDates(q Querier, ids [][]byte) (map[string]assocDates, error) {
	ids = database.UniqueBlobIDs(ids)
	out := map[string]assocDates{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := queryBatches(q, sqlAssocDates, ids, func(sc scanner) (struct {
		id         []byte
		start, end []byte
	}, error) {
		var id, start, end []byte
		if err := sc.Scan(&id, &start, &end); err != nil {
			return struct {
				id         []byte
				start, end []byte
			}{}, err
		}
		return struct {
			id         []byte
			start, end []byte
		}{id: append([]byte(nil), id...), start: start, end: end}, nil
	})
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		start, err := unmarshalDate(r.start)
		if err != nil {
			return nil, err
		}
		end, err := unmarshalDate(r.end)
		if err != nil {
			return nil, err
		}
		out[string(r.id)] = assocDates{start: start, end: end}
	}
	return out, nil
}

// placesWithoutChains is PlacesByIDs without attaching Parents (avoids
// recursion when the graph loader needs raw headers).
func placesWithoutChains(q Querier, ids [][]byte) ([]PlaceHeader, error) {
	var out []PlaceHeader
	err := database.ForEachBatch(database.UniqueBlobIDs(ids), func(batch [][]byte) error {
		rows, err := q.Query(sqlPlacesSelect+` AND e.id IN (`+database.SQLInPlaceholders(len(batch))+`)`,
			database.BlobArgs(batch)...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			h, err := scanPlace(rows)
			if err != nil {
				return err
			}
			out = append(out, h)
		}
		return rows.Err()
	})
	return out, err
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
	return g.partNeighbors(g.partUp[string(placeID)], at, true)
}

// ImmediateParts are the part_of parts that hold at at.
func (g *placeGraph) ImmediateParts(placeID []byte, at *datevalues.Value) [][]byte {
	return g.partNeighbors(g.partDown[string(placeID)], at, false)
}

// partNeighbors is the other end of each holding link, once, by name.
func (g *placeGraph) partNeighbors(links []int, at *datevalues.Value, up bool) [][]byte {
	q := queryWindow(at)
	seen := map[string]bool{}
	var out [][]byte
	for _, i := range links {
		l := g.partOf[i]
		if !g.linkHolds(q, l) {
			continue
		}
		other := l.from
		if up {
			other = l.to
		}
		if seen[string(other)] {
			continue
		}
		seen[string(other)] = true
		out = append(out, other)
	}
	g.sortByName(out)
	return out
}

// Successors are succeeded_by targets (dates do not filter succession for
// lineage reads; membership spans never build a display chain).
func (g *placeGraph) Successors(placeID []byte) [][]byte {
	return g.succNeighbors(g.succOut[string(placeID)], true)
}

// Predecessors are succeeded_by sources.
func (g *placeGraph) Predecessors(placeID []byte) [][]byte {
	return g.succNeighbors(g.succIn[string(placeID)], false)
}

func (g *placeGraph) succNeighbors(links []int, forward bool) [][]byte {
	seen := map[string]bool{}
	var out [][]byte
	for _, i := range links {
		l := g.succ[i]
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
	g.sortByName(out)
	return out
}

func (g *placeGraph) sortByName(ids [][]byte) {
	sort.SliceStable(ids, func(i, j int) bool {
		return g.displayName(ids[i]) < g.displayName(ids[j])
	})
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
// (candidates=true; no nature preference). Cycle-guarded, and remembered per
// Place and query window: a list asks for the same chain many times.
func (g *placeGraph) ParentChain(placeID []byte, at *datevalues.Value) (names []string, candidates bool) {
	key := string(placeID) + "\x00" + windowKey(queryWindow(at))
	if r, ok := g.chains[key]; ok {
		return append([]string(nil), r.names...), r.candidates
	}
	names, candidates = g.parentChain(placeID, at, map[string]bool{string(placeID): true})
	g.chains[key] = parentChainResult{names: append([]string(nil), names...), candidates: candidates}
	return names, candidates
}

func windowKey(w autoreconcile.Window) string {
	bound := func(p *int) string {
		if p == nil {
			return "*"
		}
		return strconv.Itoa(*p)
	}
	return bound(w.Lo) + ".." + bound(w.Hi)
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

// FoldLocations collapses Places that lie on one part_of chain at at into
// the deepest of them: "Ontario" and "Toronto, Ontario" are one reading,
// shown as Toronto. A Place is dropped only when another listed Place is
// below it. Places that are not on one chain stay competing, even when a
// vaguer listed Place is above both: Toronto and Hamilton under Ontario are
// two readings, not one. Succession never connects Places.
func (g *placeGraph) FoldLocations(places []HeaderPlace, at *datevalues.Value) []HeaderPlace {
	var out []HeaderPlace
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

// attachPlaceChains fills Parents for today's (or last) hierarchical chain.
func attachPlaceChains(q Querier, headers []PlaceHeader, today datevalues.Value) error {
	if len(headers) == 0 {
		return nil
	}
	ids := make([][]byte, len(headers))
	for i := range headers {
		ids[i] = headers[i].Entity.ID
	}
	g, err := loadPlaceGraph(q, ids, reachAncestors)
	if err != nil {
		return err
	}
	for i := range headers {
		// Prefer the header we already scanned (has Names/period); merge into graph.
		g.setPlace(headers[i])
	}
	for i := range headers {
		at := chainQueryDate(headers[i], today)
		headers[i].Parents, headers[i].ParentsAreCandidates = g.ParentChain(headers[i].Entity.ID, &at)
	}
	return nil
}

// PlaceDetail is one Place's header for its page: today's (or last) chain,
// and PartOf / Contains / Predecessors / Successors. One graph load serves
// both. found is false for an unknown, merged, or non-Place id.
func PlaceDetail(q Querier, id []byte) (h PlaceHeader, found bool, err error) {
	headers, err := placesWithoutChains(q, [][]byte{id})
	if err != nil || len(headers) != 1 {
		return PlaceHeader{}, false, err
	}
	h = headers[0]
	g, err := loadPlaceGraph(q, [][]byte{id}, reachDetail)
	if err != nil {
		return PlaceHeader{}, false, err
	}
	g.setPlace(h)
	today := TodayDate()
	at := chainQueryDate(h, today)
	h.Parents, h.ParentsAreCandidates = g.ParentChain(id, &at)
	g.attachRelationships(&h, today)
	return h, true, nil
}

// AttachPlaceRelationships fills PartOf / Contains / Predecessors /
// Successors on a Place detail header (every part_of parent with its
// membership span; children that hold today; succession both ways).
func AttachPlaceRelationships(q Querier, h *PlaceHeader) error {
	if h == nil || len(h.Entity.ID) == 0 {
		return nil
	}
	g, err := loadPlaceGraph(q, [][]byte{h.Entity.ID}, reachDetail)
	if err != nil {
		return err
	}
	g.setPlace(*h)
	g.attachRelationships(h, TodayDate())
	return nil
}

func (g *placeGraph) attachRelationships(h *PlaceHeader, today datevalues.Value) {

	seenPart := map[string]bool{}
	for _, i := range g.partUp[string(h.Entity.ID)] {
		l := g.partOf[i]
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
		for _, i := range g.partDown[string(h.Entity.ID)] {
			if l := g.partOf[i]; bytes.Equal(l.from, id) {
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

// ChildPlaceIDs returns Places that name id as a part_of parent (any span).
// Used by HeaderDependents so a parent rename can reproject children.
func ChildPlaceIDs(q Querier, ids [][]byte) ([][]byte, error) {
	edges, err := canonicalgraph.Walk(q, canonicalgraph.PartsOfPlace, ids)
	if err != nil {
		return nil, err
	}
	return canonicalgraph.Targets(edges), nil
}

// ParentsAtDate is the hierarchical parent display names for placeID at at
// (nearest first; several parents → every candidate).
func ParentsAtDate(q Querier, placeID []byte, at datevalues.Value) ([]string, error) {
	g, err := loadPlaceGraph(q, [][]byte{placeID}, reachAncestors)
	if err != nil {
		return nil, err
	}
	return g.ParentChainNames(placeID, &at), nil
}

// PartsAtDate is the display names of Places that are part_of placeID at at.
func PartsAtDate(q Querier, placeID []byte, at datevalues.Value) ([]string, error) {
	g, err := loadPlaceGraph(q, [][]byte{placeID}, reachDetail)
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
func SuccessionNames(q Querier, placeID []byte) (predecessors, successors []string, err error) {
	g, err := loadPlaceGraph(q, [][]byte{placeID}, reachDetail)
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
