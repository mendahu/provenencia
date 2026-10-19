package graphalign

import (
	"bytes"
	"container/heap"
	"sort"

	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Align proposes a handle, New, or Skip for every primary Subject on layer,
// holding fixed pairs as anchors. It is the layer 6 procedure: fixed pairs are
// placed; each Subject's candidates (Canon.Candidates, plus the handles a
// placed pair's links nominate) enter one best-first queue; a placed pair
// re-scores only its neighbors; and when the queue empties, only the pairs
// whose neighbors moved are settled again.
func Align(layer Layer, canon Canon, fixed []Fixed, cfg Config) Proposal {
	st := newState(layer, canon, fixed, cfg)
	st.anchor()
	st.seed()
	st.walk()
	st.settle()
	st.rescoreAll()
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

	// placed maps subjectID → handleID for fixed and accepted pairs
	placed   map[string][]byte
	fixedSet map[string]bool
	// owner maps handleID → subjectID (one-to-one within the layer)
	owner map[string]string
	// held New or Skip decisions: never placed, never walked through
	held map[string]Target
	// subjectID → subjectID that took a handle this one cleared the bar for
	lostTo map[string]string

	// cands are each Subject's candidates by handle, with their latest score
	cands map[string]map[string]*cand
	queue candHeap

	// dirty are the placed Subjects whose neighbors changed since the last
	// settle round, with the first reason and neighbor each got
	dirty map[string]mark
	// releasedFrom is the handle settling last released each Subject from, and
	// move why
	releasedFrom map[string][]byte
	move         map[string]Move
}

// An adjLink is one end's view of an edge: the neighbor, the bridge's key read
// from this end, and its key read from the neighbor's.
type adjLink struct {
	neighbor  []byte
	key, back vocabkey.BridgeKey
}

// A cand is one Subject's candidate handle and its latest score. Every
// re-score replaces the earlier one, so a score never outlives its neighbors.
type cand struct {
	handleID []byte
	ref      string
	score    float64
	edge     float64
	result   subjectcompare.Result
	// links are the segment results behind edge, one per layer link that has one
	links []segmentcompare.Result
	// via is the placed neighbor whose link nominated this handle, and viaKey
	// the layer link's key
	via     []byte
	viaKey  vocabkey.BridgeKey
	version int
}

// A mark is why a placed Subject is dirty.
type mark struct {
	reason   Reason
	neighbor []byte
}

func newState(layer Layer, canon Canon, fixed []Fixed, cfg Config) *state {
	st := &state{
		cfg:          cfg,
		layer:        layer,
		canon:        canon,
		metas:        layer.Metas,
		subjects:     map[string]*Subject{},
		handles:      map[string]*Handle{},
		layerAdj:     map[string][]adjLink{},
		canonAdj:     map[string][]adjLink{},
		placed:       map[string][]byte{},
		fixedSet:     map[string]bool{},
		owner:        map[string]string{},
		held:         map[string]Target{},
		lostTo:       map[string]string{},
		cands:        map[string]map[string]*cand{},
		dirty:        map[string]mark{},
		releasedFrom: map[string][]byte{},
		move:         map[string]Move{},
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
		st.placed[sk] = append([]byte(nil), f.HandleID...)
		st.fixedSet[sk] = true
		st.owner[hk] = sk
	}
	return st
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

// anchor scores each fixed pair's handle as the Subject's first candidate,
// then places every fixed pair in order. The final re-score reads the mapping
// the run ends on, so the answer does not depend on the order.
func (st *state) anchor() {
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if !st.fixedSet[sk] {
			continue
		}
		st.candidate(sk, st.handles[string(st.placed[sk])])
	}
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if st.fixedSet[sk] {
			st.place(sk, st.placed[sk])
		}
	}
}

// seed gives every Subject neither placed nor held the candidates the loader
// listed for it, each scored and queued.
func (st *state) seed() {
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if _, ok := st.placed[sk]; ok {
			continue
		}
		if _, ok := st.held[sk]; ok {
			continue
		}
		if _, ok := st.cfg.Subject.Weights[vocabkey.SubjectType(s.Kind)]; !ok {
			continue
		}
		for _, hid := range st.canon.Candidates[sk] {
			h := st.handles[string(hid)]
			if h == nil || h.Kind != s.Kind {
				continue
			}
			if _, ok := st.cands[sk][string(hid)]; ok {
				continue
			}
			st.push(sk, st.candidate(sk, h))
		}
	}
}

// candidate returns the Subject's candidate for h, scored now; it is created
// on first use.
func (st *state) candidate(sk string, h *Handle) *cand {
	m := st.cands[sk]
	if m == nil {
		m = map[string]*cand{}
		st.cands[sk] = m
	}
	c := m[string(h.ID)]
	if c == nil {
		c = &cand{handleID: append([]byte(nil), h.ID...), ref: h.Ref}
		m[string(h.ID)] = c
	}
	st.score(sk, h, c)
	return c
}

// score replaces c's score with the pair's score against the mapping as it
// stands, and bumps its version.
func (st *state) score(sk string, h *Handle, c *cand) {
	s := st.subjects[sk]
	r := st.evaluate(s.Kind, s.Values, h.Values)
	links := st.segments(s.ID, h, r.Evidence, st.placedOf)
	var edge float64
	for _, l := range links {
		edge += l.Evidence
	}
	sc := ScoreCandidate(r, edge, s.Provenance)
	c.score, c.edge, c.result, c.links = sc.Score, sc.Edge, sc.Result, links
	c.version++
}

// rescore re-scores every candidate of sk, pushing each when push is true.
func (st *state) rescore(sk string, push bool) {
	for _, c := range st.sortedCands(sk) {
		st.score(sk, st.handles[string(c.handleID)], c)
		if push {
			st.push(sk, c)
		}
	}
}

func (st *state) push(sk string, c *cand) {
	heap.Push(&st.queue, queueItem{
		subjectID: []byte(sk),
		handleID:  append([]byte(nil), c.handleID...),
		ref:       c.ref,
		score:     c.score,
		version:   c.version,
	})
}

// sortedCands lists a Subject's candidates best first: score, then ref, then
// handle id.
func (st *state) sortedCands(sk string) []*cand {
	out := make([]*cand, 0, len(st.cands[sk]))
	for _, c := range st.cands[sk] {
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].score != out[j].score {
			return out[i].score > out[j].score
		}
		if out[i].ref != out[j].ref {
			return out[i].ref < out[j].ref
		}
		return bytes.Compare(out[i].handleID, out[j].handleID) < 0
	})
	return out
}

func (st *state) placedOf(subjectID []byte) ([]byte, bool) {
	h, ok := st.placed[string(subjectID)]
	return h, ok
}

// place puts s on h and tells its neighbors. Each canon neighbor of h through
// a corresponding link becomes a candidate of the layer neighbor of its kind,
// at any score, with s as the one that nominated it. A neighbor that is
// already placed (and not fixed) is marked dirty; one that is not placed has
// every candidate re-scored and queued.
func (st *state) place(sk string, hid []byte) {
	st.placed[sk] = append([]byte(nil), hid...)
	st.owner[string(hid)] = sk
	reason := ReasonNeighborPlaced
	if from, ok := st.releasedFrom[sk]; ok && !bytes.Equal(from, hid) {
		reason = ReasonNeighborMoved
	}
	s := st.subjects[sk]
	done := map[string]bool{}
	for _, link := range st.layerAdj[sk] {
		nk := string(link.neighbor)
		ns := st.subjects[nk]
		if ns == nil {
			continue
		}
		if _, held := st.held[nk]; held {
			continue
		}
		_, nPlaced := st.placed[nk]
		for _, ce := range st.canonAdj[string(hid)] {
			if !segmentcompare.Corresponds(link.key, ce.key) {
				continue
			}
			g := st.handles[string(ce.neighbor)]
			if g == nil || g.Kind != ns.Kind {
				continue
			}
			c := st.cands[nk][string(g.ID)]
			if c == nil {
				c = st.candidate(nk, g)
				// An unplaced neighbor is re-scored below, with the rest.
				c.via, c.viaKey = append([]byte(nil), s.ID...), link.key
			} else if st.fixedSet[nk] && len(c.via) == 0 {
				c.via, c.viaKey = append([]byte(nil), s.ID...), link.key
			}
		}
		if st.fixedSet[nk] || done[nk] {
			continue
		}
		done[nk] = true
		if nPlaced {
			if _, ok := st.dirty[nk]; !ok {
				st.dirty[nk] = mark{reason: reason, neighbor: s.ID}
			}
		} else {
			st.rescore(nk, true)
		}
	}
}

// release takes sk off its handle: it remembers where it was and why it
// moved, marks its placed neighbors dirty, and re-scores and queues every
// candidate of sk and of its unplaced neighbors.
func (st *state) release(sk string, why mark) {
	hid := st.placed[sk]
	delete(st.placed, sk)
	if st.owner[string(hid)] == sk {
		delete(st.owner, string(hid))
	}
	st.releasedFrom[sk] = hid
	st.move[sk] = Move{Reason: why.reason, NeighborSubjectID: why.neighbor}
	st.rescore(sk, true)
	done := map[string]bool{}
	for _, link := range st.layerAdj[sk] {
		nk := string(link.neighbor)
		if done[nk] || st.subjects[nk] == nil {
			continue
		}
		done[nk] = true
		if _, held := st.held[nk]; held {
			continue
		}
		if _, ok := st.placed[nk]; ok {
			if !st.fixedSet[nk] {
				if _, ok := st.dirty[nk]; !ok {
					st.dirty[nk] = mark{reason: ReasonNeighborReleased, neighbor: []byte(sk)}
				}
			}
			continue
		}
		st.rescore(nk, true)
	}
}

// walk pops the queue best first and places the pairs that clear the accept
// bar and the one-to-one rule.
func (st *state) walk() {
	for st.queue.Len() > 0 {
		item := heap.Pop(&st.queue).(queueItem)
		sk, hk := string(item.subjectID), string(item.handleID)
		c := st.cands[sk][hk]
		if c == nil || c.version != item.version {
			continue
		}
		if _, ok := st.placed[sk]; ok {
			continue
		}
		if _, ok := st.held[sk]; ok {
			continue
		}
		if c.score < st.cfg.Alignment.AcceptBar {
			continue
		}
		if st.lostContest(sk, hk) {
			continue
		}
		st.place(sk, item.handleID)
	}
}

// settle re-scores only the pairs whose neighbors moved, until none are
// dirty or SettleRounds run out. A pair is released when it falls under the
// accept bar or a free rival leads it by SettleMargin; the released Subjects
// then walk again.
func (st *state) settle() {
	for range st.cfg.Alignment.SettleRounds {
		if len(st.dirty) == 0 {
			return
		}
		marks := st.dirty
		st.dirty = map[string]mark{}
		for _, s := range st.sortedSubjects() {
			sk := string(s.ID)
			why, ok := marks[sk]
			if !ok {
				continue
			}
			hid, ok := st.placed[sk]
			if !ok || st.fixedSet[sk] {
				continue
			}
			st.rescore(sk, false)
			cur := st.cands[sk][string(hid)]
			release := cur == nil || cur.score < st.cfg.Alignment.AcceptBar
			if !release {
				for _, c := range st.sortedCands(sk) {
					if _, taken := st.owner[string(c.handleID)]; taken {
						continue
					}
					if c.score > cur.score+st.cfg.Alignment.SettleMargin {
						release = true
					}
					break
				}
			}
			if release {
				st.release(sk, why)
			}
		}
		st.walk()
	}
}

// rescoreAll scores every candidate of every Subject once more, without
// moving anything, so rows read the mapping they end on.
func (st *state) rescoreAll() {
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		if _, held := st.held[sk]; held {
			continue
		}
		st.rescore(sk, false)
	}
}

// supports reports whether neighbor, as placed now, reaches handleID through
// a canon edge that corresponds to its layer link with subjectID.
func (st *state) supports(subjectID, handleID, neighbor []byte) bool {
	g, ok := st.placed[string(neighbor)]
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

// viaOf is the neighbor that nominated c while it still supports the pair.
func (st *state) viaOf(sk string, c *cand) ([]byte, vocabkey.BridgeKey) {
	if len(c.via) > 0 && st.supports([]byte(sk), c.handleID, c.via) {
		return c.via, c.viaKey
	}
	return nil, vocabkey.BridgeKey{}
}

func (st *state) evaluate(kind string, a, b subjectcompare.Values) subjectcompare.Result {
	return subjectcompare.Compare(vocabkey.SubjectType(kind), a, b, st.metas, st.cfg.Subject, st.cfg.Value)
}

// segments is the one placement rule: what each layer link of subjectID adds
// to the pair with h, whose own evidence is own. Each link stands on its own,
// and placedOf says where a layer neighbor is placed. A neighbor placed on
// the canon neighbor of a corresponding link is compared together; one placed
// elsewhere, apart on the first corresponding link; one not placed yet,
// unplaced on each corresponding link whose canon neighbor no Subject holds,
// keeping the result with the most evidence. A held neighbor, or a link with
// no corresponding canon link, has no result.
func (st *state) segments(subjectID []byte, h *Handle, own float64, placedOf func(subjectID []byte) ([]byte, bool)) []segmentcompare.Result {
	var out []segmentcompare.Result
	canon := st.canonAdj[string(h.ID)]
	for _, link := range st.layerAdj[string(subjectID)] {
		ns := st.subjects[string(link.neighbor)]
		if ns == nil {
			continue
		}
		if _, held := st.held[string(link.neighbor)]; held {
			continue
		}
		hid, isPlaced := placedOf(link.neighbor)
		if !isPlaced {
			var best *segmentcompare.Result
			for i := range canon {
				ce := &canon[i]
				if !segmentcompare.Corresponds(link.key, ce.key) {
					continue
				}
				nh := st.handles[string(ce.neighbor)]
				if nh == nil || nh.Kind != ns.Kind {
					continue
				}
				if _, held := st.owner[string(ce.neighbor)]; held {
					continue
				}
				r := st.compareSegments(link, ns, *ce, nh, segmentcompare.Unplaced, own)
				if best == nil || r.Evidence > best.Evidence {
					best = &r
				}
			}
			if best != nil {
				out = append(out, *best)
			}
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
		out = append(out, st.compareSegments(link, ns, *ce, nh, placed, own))
	}
	return out
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

// lostContest reports whether another Subject already holds the handle, and
// remembers the first such rival: a row that cleared the bar for a handle
// another row took may be that row's duplicate.
func (st *state) lostContest(sk, hk string) bool {
	owner, ok := st.owner[hk]
	if !ok || owner == sk {
		return false
	}
	if _, seen := st.lostTo[sk]; !seen {
		st.lostTo[sk] = owner
	}
	return true
}

// Priority queue of candidates (best score first; ties by ref, handle id,
// then subject id, so equal candidates pop in one order every run).
type queueItem struct {
	subjectID []byte
	handleID  []byte
	ref       string
	score     float64
	version   int // the candidate's version when this entry was pushed
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
