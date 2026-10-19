package graphalign

import (
	"bytes"
	"container/heap"
	"slices"
	"sort"

	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Align proposes a handle, New, or Skip for every primary Subject on layer,
// holding fixed pairs as anchors.
func Align(layer Layer, canon Canon, fixed []Fixed, cfg Config) Proposal {
	st := newState(layer, canon, fixed, cfg)
	st.seed()
	st.walk()
	// Property matches that clear the bar seed the same walk. Candidates
	// only recorded are not anchors, and a pair this walk refuses
	// does not propagate further.
	anchored := st.assignedIDs()
	st.fallbackUnreachable()
	st.propagateNew(anchored)
	st.walk()
	st.refineOneHop()
	st.scoreFixed()
	return st.proposal()
}

type state struct {
	cfg   Config
	layer Layer
	canon Canon
	metas subjectcompare.Metas

	subjects map[string]*Subject
	handles  map[string]*Handle
	// undirected neighbors on the layer: subjectID → []neighborID
	layerAdj map[string][]adjLink
	// canon adjacency: handleID → []adjLink
	canonAdj map[string][]adjLink

	// mapping subjectID → handleID for fixed + accepted
	assigned map[string][]byte
	fixedSet map[string]bool
	// handleID → subjectID (one-to-one within the layer)
	handleOwner map[string]string
	// held New or Skip decisions: never mapped, never walked through
	held map[string]Target
	// subjectID → subjectID that took a handle this one cleared the bar for
	lostTo map[string]string

	// best scored candidates per subject (for alternatives / fallback)
	best map[string][]scoredCand

	queue candHeap
	seen  map[string]bool // subject|handle already accepted or rejected for queue

	kindCands map[string][]Handle // handlesOfKind, per kind
}

// An adjLink is one end's view of an edge: the neighbor, the bridge's key read
// from this end, and its key read from the neighbor's.
type adjLink struct {
	neighbor  []byte
	key, back vocabkey.BridgeKey
}

type scoredCand struct {
	handleID    []byte
	ref         string
	score       float64
	result      subjectcompare.Result
	edge        float64
	viaNeighbor []byte
	viaKey      vocabkey.BridgeKey
}

func newState(layer Layer, canon Canon, fixed []Fixed, cfg Config) *state {
	st := &state{
		cfg:         cfg,
		layer:       layer,
		canon:       canon,
		metas:       layer.Metas,
		subjects:    map[string]*Subject{},
		handles:     map[string]*Handle{},
		layerAdj:    map[string][]adjLink{},
		canonAdj:    map[string][]adjLink{},
		assigned:    map[string][]byte{},
		fixedSet:    map[string]bool{},
		handleOwner: map[string]string{},
		held:        map[string]Target{},
		lostTo:      map[string]string{},
		best:        map[string][]scoredCand{},
		seen:        map[string]bool{},
		kindCands:   map[string][]Handle{},
	}
	for i := range layer.Subjects {
		s := &layer.Subjects[i]
		st.subjects[string(s.ID)] = s
	}
	for i := range canon.Handles {
		h := &canon.Handles[i]
		st.handles[string(h.ID)] = h
	}
	for _, b := range layer.Bridges {
		a, c := string(b.A), string(b.B)
		st.layerAdj[a] = append(st.layerAdj[a], adjLink{neighbor: b.B, key: b.KeyA, back: b.KeyB})
		st.layerAdj[c] = append(st.layerAdj[c], adjLink{neighbor: b.A, key: b.KeyB, back: b.KeyA})
	}
	for _, e := range canon.Edges {
		f, t := string(e.From), string(e.To)
		st.canonAdj[f] = append(st.canonAdj[f], adjLink{neighbor: e.To, key: e.KeyFrom, back: e.KeyTo})
		st.canonAdj[t] = append(st.canonAdj[t], adjLink{neighbor: e.From, key: e.KeyTo, back: e.KeyFrom})
	}
	for _, f := range fixed {
		sk, hk := string(f.SubjectID), string(f.HandleID)
		if _, ok := st.subjects[sk]; !ok {
			continue
		}
		if f.Target == TargetNew || f.Target == TargetSkip {
			st.held[sk] = f.Target
			continue
		}
		if _, ok := st.handles[hk]; !ok {
			continue
		}
		st.assigned[sk] = append([]byte(nil), f.HandleID...)
		st.fixedSet[sk] = true
		st.handleOwner[hk] = sk
	}
	return st
}

func (st *state) seed() {
	for _, s := range st.sortedSubjects() {
		if hid, ok := st.assigned[string(s.ID)]; ok {
			st.pushFromAnchor(s.ID, hid)
		}
	}
}

// sortedSubjects is the layer in a stable order (ref, then id) so nothing
// Align decides depends on map iteration.
func (st *state) sortedSubjects() []*Subject {
	subs := make([]*Subject, 0, len(st.subjects))
	for _, s := range st.subjects {
		subs = append(subs, s)
	}
	sort.SliceStable(subs, func(i, j int) bool {
		if subs[i].Ref != subs[j].Ref {
			return subs[i].Ref < subs[j].Ref
		}
		return bytes.Compare(subs[i].ID, subs[j].ID) < 0
	})
	return subs
}

func (st *state) pushFromAnchor(subjectID, handleID []byte) {
	sk := string(subjectID)
	for _, link := range st.layerAdj[sk] {
		nk := string(link.neighbor)
		if _, taken := st.assigned[nk]; taken {
			continue
		}
		if _, decided := st.held[nk]; decided {
			continue
		}
		ns := st.subjects[nk]
		if ns == nil {
			continue
		}
		for _, ce := range st.canonAdj[string(handleID)] {
			if !segmentcompare.Corresponds(link.key, ce.key) {
				continue
			}
			h := st.handles[string(ce.neighbor)]
			if h == nil || h.Kind != ns.Kind {
				continue
			}
			st.enqueue(ns.ID, h, link.key, subjectID)
		}
	}
}

func (st *state) enqueue(subjectID []byte, h *Handle, via vocabkey.BridgeKey, viaNeighbor []byte) {
	key := string(subjectID) + "|" + string(h.ID)
	if st.seen[key] {
		return
	}
	sc := st.scoreCandidate(subjectID, h)
	sc.viaNeighbor = append([]byte(nil), viaNeighbor...)
	sc.viaKey = via
	st.recordBest(string(subjectID), sc)
	heap.Push(&st.queue, queueItem{
		subjectID: append([]byte(nil), subjectID...),
		handleID:  append([]byte(nil), h.ID...),
		ref:       h.Ref,
		score:     sc.score,
	})
}

// recordBest keeps the higher score for a handle already listed. During a
// walk the mapping only grows, so a later score for the same pair is never
// lower for a real reason; refinement replaces scores outright (rescoreBest).
func (st *state) recordBest(sk string, sc scoredCand) {
	list := st.best[sk]
	replaced := false
	for i := range list {
		if bytes.Equal(list[i].handleID, sc.handleID) {
			if sc.score > list[i].score {
				// A walk remembers which neighbor proposed this handle.
				// Rescoring raises the number and keeps that path.
				if len(sc.viaNeighbor) == 0 && len(list[i].viaNeighbor) > 0 {
					sc.viaNeighbor = append([]byte(nil), list[i].viaNeighbor...)
					sc.viaKey = list[i].viaKey
				}
				list[i] = sc
			}
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, sc)
	}
	sortCands(list)
	st.best[sk] = list
}

// sortCands orders candidates best first: score, then ref, then handle id.
func sortCands(list []scoredCand) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		if list[i].ref != list[j].ref {
			return list[i].ref < list[j].ref
		}
		return bytes.Compare(list[i].handleID, list[j].handleID) < 0
	})
}

// supports reports whether neighbor, as mapped now, reaches handleID through
// a canon edge that corresponds to its layer link with subjectID.
func (st *state) supports(subjectID, handleID, neighbor []byte) bool {
	g, ok := st.assigned[string(neighbor)]
	if !ok {
		return false
	}
	for _, link := range st.layerAdj[string(subjectID)] {
		if !bytes.Equal(link.neighbor, neighbor) {
			continue
		}
		for _, ce := range st.canonAdj[string(handleID)] {
			if segmentcompare.Corresponds(link.key, ce.key) && bytes.Equal(ce.neighbor, g) {
				return true
			}
		}
	}
	return false
}

func (st *state) evaluate(kind string, a, b subjectcompare.Values) subjectcompare.Result {
	return subjectcompare.Compare(vocabkey.SubjectType(kind), a, b, st.metas, st.cfg.Subject, st.cfg.Value)
}

func (st *state) scoreCandidate(subjectID []byte, h *Handle) scoredCand {
	s := st.subjects[string(subjectID)]
	r := st.evaluate(s.Kind, s.Values, h.Values)
	edge := st.segments(subjectID, h, r.Evidence)
	sc := ScoreCandidate(r, edge, s.Provenance)
	return scoredCand{
		handleID: append([]byte(nil), h.ID...),
		ref:      h.Ref,
		score:    sc.Score,
		result:   sc.Result,
		edge:     sc.Edge,
	}
}

// segments sums what each mapped layer neighbor's link adds to the pair of
// subjectID and h, whose own evidence is own. Each link of the Subject stands
// on its own: the first canon link of h that corresponds and lands on the
// neighbor's handle is compared together; failing that, the first that
// corresponds is compared apart; none adds nothing. A neighbor not yet mapped
// adds nothing: a missing neighbor is not a disagreement.
func (st *state) segments(subjectID []byte, h *Handle, own float64) float64 {
	var total float64
	canon := st.canonAdj[string(h.ID)]
	for _, link := range st.layerAdj[string(subjectID)] {
		hid, ok := st.assigned[string(link.neighbor)]
		if !ok {
			continue
		}
		ns := st.subjects[string(link.neighbor)]
		if ns == nil {
			continue
		}
		var together, apart *adjLink
		for i := range canon {
			ce := &canon[i]
			if !segmentcompare.Corresponds(link.key, ce.key) {
				continue
			}
			if bytes.Equal(ce.neighbor, hid) {
				together = ce
				break
			}
			if apart == nil {
				apart = ce
			}
		}
		placed, ce := segmentcompare.Together, together
		if ce == nil {
			placed, ce = segmentcompare.Apart, apart
		}
		if ce == nil {
			continue
		}
		nh := st.handles[string(ce.neighbor)]
		if nh == nil {
			continue
		}
		total += st.compareSegments(link, ns, *ce, nh, placed, own).Evidence
	}
	return total
}

// compareSegments is layer 5 on one layer link and one canon link.
func (st *state) compareSegments(l adjLink, ns *Subject, ce adjLink, nh *Handle, placed segmentcompare.Placement, own float64) segmentcompare.Result {
	return st.compare(l.key, ns.Kind, ns.Values, ce.key, nh.Kind, nh.Values, placed, own)
}

// compare is the one call to layer 5: two segments, each a key, the neighbor's
// kind and its values.
func (st *state) compare(keyA vocabkey.BridgeKey, kindA string, valuesA subjectcompare.Values, keyB vocabkey.BridgeKey, kindB string, valuesB subjectcompare.Values, placed segmentcompare.Placement, own float64) segmentcompare.Result {
	return segmentcompare.Compare(
		segmentcompare.Segment{Key: keyA, NeighborKind: vocabkey.SubjectType(kindA), Neighbor: valuesA},
		segmentcompare.Segment{Key: keyB, NeighborKind: vocabkey.SubjectType(kindB), Neighbor: valuesB},
		placed, own, st.metas, st.cfg.Segment, st.cfg.Subject, st.cfg.Value)
}

func (st *state) walk() {
	for st.queue.Len() > 0 {
		item := heap.Pop(&st.queue).(queueItem)
		seenKey := string(item.subjectID) + "|" + string(item.handleID)
		if st.seen[seenKey] {
			continue
		}
		st.seen[seenKey] = true
		if st.accept(item.subjectID, item.handleID, item.score) {
			st.pushFromAnchor(item.subjectID, item.handleID)
		}
	}
}

// accept stores the pair when the subject is free, the score clears the bar,
// and one-to-one still holds. Callers propagate; this only records the mapping.
func (st *state) accept(subjectID, handleID []byte, score float64) bool {
	sk, hk := string(subjectID), string(handleID)
	if _, taken := st.assigned[sk]; taken {
		return false
	}
	if score < st.cfg.Alignment.AcceptBar {
		return false
	}
	if st.lostContest(sk, hk) {
		return false
	}
	st.assigned[sk] = append([]byte(nil), handleID...)
	st.handleOwner[hk] = sk
	return true
}

// fallbackUnreachable handles Subjects no anchor reached. Candidates picks
// each one's handles on its own properties; each is scored by the same pairwise path the
// walk uses. Assignment is best-first, so a contest over one handle goes to
// the stronger match. propagateNew then walks from the pairs this pass
// accepts. refineOneHop rescores the recorded set; the band published for
// these rows is that later score.
func (st *state) fallbackUnreachable() {
	var queue candHeap
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if _, ok := st.assigned[sk]; ok {
			continue
		}
		if _, decided := st.held[sk]; decided {
			continue
		}
		if _, ok := st.cfg.Subject.Weights[vocabkey.SubjectType(s.Kind)]; !ok {
			continue
		}
		for _, cand := range Candidates(s.Kind, s.Values, st.handlesOfKind(s.Kind), st.metas, st.cfg) {
			h := st.handles[string(cand.ID)]
			sc := st.scoreCandidate(s.ID, h)
			st.recordBest(sk, sc)
			heap.Push(&queue, queueItem{
				subjectID: append([]byte(nil), s.ID...),
				handleID:  append([]byte(nil), h.ID...),
				ref:       h.Ref,
				score:     sc.score,
			})
		}
	}
	for queue.Len() > 0 {
		item := heap.Pop(&queue).(queueItem)
		st.accept(item.subjectID, item.handleID, item.score)
	}
}

// assignedIDs is the subjects that already have a handle, so a later pass
// can tell its own accepts from anchors the walk already propagated.
func (st *state) assignedIDs() map[string]bool {
	out := map[string]bool{}
	for sk := range st.assigned {
		out[sk] = true
	}
	return out
}

// propagateNew seeds the walk from accepts that were not already anchors.
// pushFromAnchor nominates an unmapped neighbor only when a canon edge
// carries the same signature as the layer bridge.
func (st *state) propagateNew(already map[string]bool) {
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if already[sk] {
			continue
		}
		hid, ok := st.assigned[sk]
		if !ok {
			continue
		}
		st.pushFromAnchor(s.ID, hid)
	}
}

// refineOneHop rescores every candidate from the mapping the property pass
// just published, then reassigns unfixed rows. A neighbor's handle counts;
// that neighbor's own edge bonus does not, so support stays one hop.
// Repeat until the mapping settles, at most one round per subject; when the
// rounds run out, score once more so every row reads the mapping it ends on.
func (st *state) refineOneHop() {
	rounds := len(st.subjects)
	for range rounds {
		items := st.rescoreBest()
		if !st.reassignUnfixed(items) {
			return
		}
	}
	st.rescoreBest()
}

// rescoreBest scores the candidates the fallback and the walk already recorded,
// against the mapping as it stands, and replaces their earlier scores: a
// score that leaned on a neighbor's old handle drops when that neighbor
// moves. A candidate keeps the neighbor that nominated it only while that
// neighbor still supports it. It does not add handles.
func (st *state) rescoreBest() []queueItem {
	var items []queueItem
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if st.fixedSet[sk] {
			continue
		}
		if _, held := st.held[sk]; held {
			continue
		}
		fresh := make([]scoredCand, 0, len(st.best[sk]))
		for _, cand := range st.best[sk] {
			h := st.handles[string(cand.handleID)]
			if h == nil {
				continue
			}
			sc := st.scoreCandidate(s.ID, h)
			if len(cand.viaNeighbor) > 0 && st.supports(s.ID, h.ID, cand.viaNeighbor) {
				sc.viaNeighbor = cand.viaNeighbor
				sc.viaKey = cand.viaKey
			}
			fresh = append(fresh, sc)
		}
		sortCands(fresh)
		st.best[sk] = fresh
		for _, cand := range st.best[sk] {
			items = append(items, queueItem{
				subjectID: append([]byte(nil), s.ID...),
				handleID:  append([]byte(nil), cand.handleID...),
				ref:       cand.ref,
				score:     cand.score,
			})
		}
	}
	return items
}

// reassignUnfixed drops every unfixed pair and accepts the rescored
// candidates. changed is false when that mapping is the one just scored.
func (st *state) reassignUnfixed(items []queueItem) bool {
	before := map[string]string{}
	for sk, hid := range st.assigned {
		if st.fixedSet[sk] {
			continue
		}
		before[sk] = string(hid)
		if st.handleOwner[string(hid)] == sk {
			delete(st.handleOwner, string(hid))
		}
		delete(st.assigned, sk)
	}
	st.lostTo = map[string]string{}

	var queue candHeap
	for _, item := range items {
		heap.Push(&queue, item)
	}
	for queue.Len() > 0 {
		item := heap.Pop(&queue).(queueItem)
		st.accept(item.subjectID, item.handleID, item.score)
	}

	after := map[string]string{}
	for sk, hid := range st.assigned {
		if !st.fixedSet[sk] {
			after[sk] = string(hid)
		}
	}
	if len(before) != len(after) {
		return true
	}
	for sk, hid := range before {
		if after[sk] != hid {
			return true
		}
	}
	return false
}

// lostContest reports whether another Subject already holds the handle, and
// remembers the first such rival: a row that cleared the bar for a handle
// another row took may be that row's duplicate.
func (st *state) lostContest(sk, hk string) bool {
	owner, ok := st.handleOwner[hk]
	if !ok || owner == sk {
		return false
	}
	if _, seen := st.lostTo[sk]; !seen {
		st.lostTo[sk] = owner
	}
	return true
}

// scoreFixed scores each held handle with its now-mapped neighbors, and
// scores the handles those neighbors point at instead, so a decision the
// rest of the page contradicts can carry a warning (design §3.1).
func (st *state) scoreFixed() {
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if !st.fixedSet[sk] {
			continue
		}
		hid := st.assigned[sk]
		if h := st.handles[string(hid)]; h != nil {
			st.recordBest(sk, st.scoreCandidate(s.ID, h))
		}
		for _, link := range st.layerAdj[sk] {
			g, ok := st.assigned[string(link.neighbor)]
			if !ok {
				continue
			}
			// canonAdj[g] is seen from the neighbor's end, so compare the
			// neighbor's view of this link.
			for _, ce := range st.canonAdj[string(g)] {
				if !segmentcompare.Corresponds(link.back, ce.key) || bytes.Equal(ce.neighbor, hid) {
					continue
				}
				h := st.handles[string(ce.neighbor)]
				if h == nil || h.Kind != s.Kind {
					continue
				}
				sc := st.scoreCandidate(s.ID, h)
				sc.viaNeighbor = append([]byte(nil), link.neighbor...)
				sc.viaKey = link.key
				st.recordBest(sk, sc)
			}
		}
	}
}

// handlesOfKind lists the canon handles of one kind in a stable order, built
// once per kind.
func (st *state) handlesOfKind(kind string) []Handle {
	if hs, ok := st.kindCands[kind]; ok {
		return hs
	}
	var out []Handle
	for i := range st.canon.Handles {
		if h := st.canon.Handles[i]; h.Kind == kind {
			out = append(out, h)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Ref != out[j].Ref {
			return out[i].Ref < out[j].Ref
		}
		return bytes.Compare(out[i].ID, out[j].ID) < 0
	})
	st.kindCands[kind] = out
	return out
}

func (st *state) proposal() Proposal {
	var rows []Row
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		row := Row{SubjectID: append([]byte(nil), s.ID...), Kind: s.Kind}
		if hid, ok := st.assigned[sk]; ok {
			h := st.handles[string(hid)]
			row.Target = TargetHandle
			row.HandleID = append([]byte(nil), hid...)
			sc := st.pickScore(sk, hid)
			if h != nil {
				row.HandleRef = h.Ref
				if len(sc.result.Properties) == 0 {
					sc.result = st.evaluate(s.Kind, s.Values, h.Values)
				}
			}
			row.Score = sc.score
			row.Comparisons = comparisonsFrom(sc.result, s.Kind, st.metas, st.cfg.Subject)
			switch {
			case st.fixedSet[sk]:
				row.Reason = ReasonDecided
			case len(sc.viaNeighbor) > 0:
				row.Reason = ReasonVia
				row.Via = &Via{
					NeighborSubjectID: append([]byte(nil), sc.viaNeighbor...),
					Key:               sc.viaKey,
				}
			default:
				row.Reason = ReasonAgrees
				row.ReasonProperty = strongestAgreement(row.Comparisons)
			}
			row.Assessment = st.band(sc.score)
			row.Alternatives = st.alts(sk, hid)
			if st.fixedSet[sk] {
				if best := st.bestNonAssigned(sk, hid); best != nil && best.score > sc.score+st.cfg.Alignment.SettleMargin {
					row.Flags.ConflictWithFixed = true
				}
			}
		} else if decided, ok := st.held[sk]; ok {
			row.Target = decided
			row.Assessment = AssessNone
			row.Reason = ReasonDecided
		} else {
			row.Assessment = AssessNone
			if best := st.bestNonAssigned(sk, nil); best != nil {
				// A candidate existed but was not accepted (below the bar or
				// lost one-to-one): Skip, never force a merge.
				row.Target = TargetSkip
				row.Score = best.score
				row.Alternatives = st.alts(sk, nil)
				row.Reason = ReasonWeak
				if best.score >= st.cfg.Alignment.Bands.Weak {
					row.Assessment = st.band(best.score)
				}
			} else if st.subjectHasValues(s) {
				row.Target = TargetNew
				row.Reason = ReasonNoMatch
			} else {
				row.Target = TargetSkip
				row.Reason = ReasonEmpty
			}
			if owner, ok := st.lostTo[sk]; ok {
				row.Reason = ReasonTaken
				row.Flags.PossibleDuplicate = true
				row.Flags.DuplicateOf = []byte(owner)
			}
		}
		rows = append(rows, row)
	}
	st.flagSharedHandles(rows)
	st.flagAlikeNewRows(rows)
	return Proposal{Rows: rows}
}

// flagSharedHandles marks rows held on one handle (one-to-one keeps the walk
// from doing this; two decisions can).
func (st *state) flagSharedHandles(rows []Row) {
	byHandle := map[string][]int{}
	for i, r := range rows {
		if r.Target == TargetHandle && len(r.HandleID) > 0 {
			byHandle[string(r.HandleID)] = append(byHandle[string(r.HandleID)], i)
		}
	}
	for _, idx := range byHandle {
		if len(idx) <= 1 {
			continue
		}
		for _, i := range idx {
			other := idx[0]
			if other == i {
				other = idx[1]
			}
			rows[i].Flags.PossibleDuplicate = true
			rows[i].Flags.DuplicateOf = append([]byte(nil), rows[other].SubjectID...)
		}
	}
}

// flagAlikeNewRows marks two New rows of one kind that would clear the accept
// bar against each other: filing both mints two handles for what may be one
// entity. A signature both rows have neighbors for suppresses the warning
// when none of those neighbors are the same subject or a match. It only
// warns, so the bar is acceptance, not strong.
func (st *state) flagAlikeNewRows(rows []Row) {
	for i := range rows {
		if rows[i].Target != TargetNew || rows[i].Flags.PossibleDuplicate {
			continue
		}
		a := st.subjects[string(rows[i].SubjectID)]
		for j := range rows {
			if j == i || rows[j].Target != TargetNew || rows[j].Kind != rows[i].Kind {
				continue
			}
			b := st.subjects[string(rows[j].SubjectID)]
			if !st.alikeNew(a, b) {
				continue
			}
			rows[i].Flags.PossibleDuplicate = true
			rows[i].Flags.DuplicateOf = append([]byte(nil), rows[j].SubjectID...)
			break
		}
	}
}

// alikeNew reports whether two New subjects score as one entity. Properties
// are scored as usual. Each link key both have neighbors for adds one link
// when a neighbor matches, and rejects the pair when none do.
func (st *state) alikeNew(a, b *Subject) bool {
	r := st.evaluate(a.Kind, a.Values, b.Values)
	segs, disagree := st.newPairSegments(a, b, r.Evidence)
	if disagree {
		return false
	}
	return ScoreCandidate(r, segs, 1).Score >= st.cfg.Alignment.AcceptBar
}

// newPairSegments compares the layer neighbors of two subjects, whose own
// evidence is own. A key only one of them has stays out. A key both have
// disagrees when no neighbor pair is the same subject or clears the accept bar
// on its own properties. A directed key carries its end, so a man's son and
// another man's father are not neighbors to compare. It returns the sum of
// the best segment evidence per key.
func (st *state) newPairSegments(a, b *Subject, own float64) (float64, bool) {
	gb := map[vocabkey.BridgeKey][]adjLink{}
	for _, link := range st.layerAdj[string(b.ID)] {
		gb[link.key] = append(gb[link.key], link)
	}
	seen := map[vocabkey.BridgeKey]bool{}
	var total float64
	for _, linkA := range st.layerAdj[string(a.ID)] {
		if seen[linkA.key] {
			continue
		}
		nb := gb[linkA.key]
		if len(nb) == 0 {
			continue
		}
		seen[linkA.key] = true
		best, ok := 0.0, false
		for _, la := range st.layerAdj[string(a.ID)] {
			if la.key != linkA.key {
				continue
			}
			for _, lb := range nb {
				same := bytes.Equal(la.neighbor, lb.neighbor)
				sa, sb := st.subjects[string(la.neighbor)], st.subjects[string(lb.neighbor)]
				if sa == nil || sb == nil || (!same && sa.Kind != sb.Kind) {
					continue
				}
				res := st.compare(la.key, sa.Kind, sa.Values, lb.key, sb.Kind, sb.Values, segmentcompare.Together, own)
				if !same && res.Neighbor.Evidence < st.cfg.Alignment.AcceptBar {
					continue
				}
				if !ok || res.Evidence > best {
					best, ok = res.Evidence, true
				}
			}
		}
		if !ok {
			return 0, true
		}
		total += best
	}
	return total, false
}

// strongestAgreement is the agreeing or resembling property that added the
// most points. A derived property (the lifespan) is never named: it counts in
// the score but has no copy on the page.
func strongestAgreement(cs []Comparison) vocabkey.PropertyKey {
	var top vocabkey.PropertyKey
	best := 0.0
	for _, c := range cs {
		if c.Derived {
			continue
		}
		if (c.Result.Outcome == propertycompare.OutcomeAgree || c.Result.Outcome == propertycompare.OutcomePartial) && c.Points > best {
			top, best = c.Key, c.Points
		}
	}
	return top
}

// comparisonsFrom lists one Comparison per lead property r compared.
func comparisonsFrom(r subjectcompare.Result, kind string, metas subjectcompare.Metas, rules subjectcompare.Rules) []Comparison {
	out := make([]Comparison, 0, len(r.Properties))
	for _, pr := range r.Properties {
		out = append(out, Comparison{
			PropertyResult: pr,
			ValueType:      metas[pr.Key].ValueType,
			Derived:        rules.Derived(vocabkey.SubjectType(kind), pr.Key),
		})
	}
	return out
}

// pickScore is the recorded score for the row's handle. Every pass records
// the handle it assigns, so a miss scores it now rather than inventing one.
func (st *state) pickScore(sk string, hid []byte) scoredCand {
	for _, sc := range st.best[sk] {
		if bytes.Equal(sc.handleID, hid) {
			return sc
		}
	}
	if h := st.handles[string(hid)]; h != nil {
		return st.scoreCandidate([]byte(sk), h)
	}
	return scoredCand{handleID: append([]byte(nil), hid...)}
}

func (st *state) bestNonAssigned(sk string, except []byte) *scoredCand {
	for i := range st.best[sk] {
		sc := &st.best[sk][i]
		if except != nil && bytes.Equal(sc.handleID, except) {
			continue
		}
		return sc
	}
	return nil
}

func (st *state) alts(sk string, except []byte) []Alternative {
	kind := st.subjects[sk].Kind
	limit := st.cfg.Alignment.Alternatives
	var out []Alternative
	for _, sc := range st.best[sk] {
		if except != nil && bytes.Equal(sc.handleID, except) {
			continue
		}
		out = append(out, st.alternative(sc, kind))
		if len(out) >= limit {
			break
		}
	}
	return out
}

// alternative is one selectable record: its band, and the sentence that
// explains that band. A via walk wins; otherwise the strongest agreeing
// property; otherwise there was too little to name.
func (st *state) alternative(sc scoredCand, kind string) Alternative {
	alt := Alternative{
		HandleID:   append([]byte(nil), sc.handleID...),
		Ref:        sc.ref,
		Score:      sc.score,
		Assessment: st.band(sc.score),
		Reason:     ReasonWeak,
	}
	if len(sc.viaNeighbor) > 0 {
		alt.Reason = ReasonVia
		alt.ViaNeighbor = append([]byte(nil), sc.viaNeighbor...)
		return alt
	}
	// Below the weak bar there is not enough to name an agreement.
	if alt.Assessment == AssessNone {
		return alt
	}
	prop := strongestAgreement(comparisonsFrom(sc.result, kind, st.metas, st.cfg.Subject))
	if prop != "" {
		alt.Reason = ReasonAgrees
		alt.ReasonProperty = prop
	}
	return alt
}

// subjectHasValues reports whether the Subject records anything matching
// reads: a positive value under a lead property the profile builder does not
// derive.
func (st *state) subjectHasValues(s *Subject) bool {
	kind := vocabkey.SubjectType(s.Kind)
	for key, vs := range s.Values {
		if _, ok := st.cfg.Subject.Lead(kind, key); !ok || st.cfg.Subject.Derived(kind, key) {
			continue
		}
		if slices.ContainsFunc(vs, propertycompare.Value.Positive) {
			return true
		}
	}
	return false
}

func (st *state) band(score float64) Assessment {
	if score >= st.cfg.Alignment.Bands.Strong {
		return AssessStrong
	}
	if st.cfg.Alignment.Bands.Medium > 0 && score >= st.cfg.Alignment.Bands.Medium {
		return AssessMedium
	}
	if score >= st.cfg.Alignment.Bands.Weak {
		return AssessWeak
	}
	return AssessNone
}

// Priority queue of candidates (best score first; ties by ref, handle id,
// then subject id, so equal candidates pop in one order every run).
type queueItem struct {
	subjectID []byte
	handleID  []byte
	ref       string
	score     float64
	index     int
}

type candHeap []queueItem

func (h candHeap) Len() int { return len(h) }

func (h candHeap) Less(i, j int) bool {
	if h[i].score != h[j].score {
		return h[i].score > h[j].score
	}
	if h[i].ref != h[j].ref {
		return h[i].ref < h[j].ref
	}
	if c := bytes.Compare(h[i].handleID, h[j].handleID); c != 0 {
		return c < 0
	}
	return bytes.Compare(h[i].subjectID, h[j].subjectID) < 0
}

func (h candHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *candHeap) Push(x any) {
	item := x.(queueItem)
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *candHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = queueItem{}
	*h = old[:n-1]
	return item
}
