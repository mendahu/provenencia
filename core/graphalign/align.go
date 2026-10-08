package graphalign

import (
	"bytes"
	"container/heap"
	"sort"

	"github.com/mendahu/provenencia/core/match"
)

// Align proposes a handle, New, or Skip for every primary Subject on layer,
// holding fixed pairs as anchors. cfg nil uses DefaultConfig().
func Align(layer Layer, canon Canon, stats Stats, fixed []Fixed, cfg *Config) Proposal {
	c := DefaultConfig()
	if cfg != nil {
		c = *cfg
	}
	st := newState(layer, canon, stats, fixed, c)
	st.seed()
	st.walk()
	st.fallbackUnreachable()
	st.scoreFixed()
	return st.proposal()
}

type state struct {
	cfg   Config
	layer Layer
	canon Canon
	stats Stats
	metas []match.PropertyMeta

	subjects map[string]*Subject
	handles  map[string]*Handle
	// undirected neighbors on the layer: subjectID → []neighborID
	layerAdj map[string][]layerLink
	// canon adjacency: handleID → []canonLink
	canonAdj map[string][]canonLink

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
}

type layerLink struct {
	neighbor []byte
	sig      EdgeSignature
}

type canonLink struct {
	neighbor []byte
	sig      EdgeSignature
}

type scoredCand struct {
	handleID    []byte
	ref         string
	score       float64
	eval        match.Evaluation
	reasons     []string
	edge        float64
	viaNeighbor []byte
	viaSig      EdgeSignature
}

func newState(layer Layer, canon Canon, stats Stats, fixed []Fixed, cfg Config) *state {
	st := &state{
		cfg:         cfg,
		layer:       layer,
		canon:       canon,
		stats:       stats,
		metas:       layer.Metas,
		subjects:    map[string]*Subject{},
		handles:     map[string]*Handle{},
		layerAdj:    map[string][]layerLink{},
		canonAdj:    map[string][]canonLink{},
		assigned:    map[string][]byte{},
		fixedSet:    map[string]bool{},
		handleOwner: map[string]string{},
		held:        map[string]Target{},
		lostTo:      map[string]string{},
		best:        map[string][]scoredCand{},
		seen:        map[string]bool{},
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
		st.layerAdj[a] = append(st.layerAdj[a], layerLink{neighbor: b.B, sig: b.Signature})
		st.layerAdj[c] = append(st.layerAdj[c], layerLink{neighbor: b.A, sig: b.Signature})
	}
	for _, e := range canon.Edges {
		f, t := string(e.From), string(e.To)
		st.canonAdj[f] = append(st.canonAdj[f], canonLink{neighbor: e.To, sig: e.Signature})
		st.canonAdj[t] = append(st.canonAdj[t], canonLink{neighbor: e.From, sig: e.Signature})
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
			if ce.sig.Key() != link.sig.Key() {
				continue
			}
			h := st.handles[string(ce.neighbor)]
			if h == nil || h.Kind != ns.Kind {
				continue
			}
			st.enqueue(ns.ID, h, link.sig, subjectID)
		}
	}
}

func (st *state) enqueue(subjectID []byte, h *Handle, via EdgeSignature, viaNeighbor []byte) {
	key := string(subjectID) + "|" + string(h.ID)
	if st.seen[key] {
		return
	}
	sc := st.scoreCandidate(subjectID, h)
	sc.viaNeighbor = append([]byte(nil), viaNeighbor...)
	sc.viaSig = via
	st.recordBest(string(subjectID), sc)
	heap.Push(&st.queue, queueItem{
		subjectID: append([]byte(nil), subjectID...),
		handleID:  append([]byte(nil), h.ID...),
		ref:       h.Ref,
		score:     sc.score,
	})
}

func (st *state) recordBest(sk string, sc scoredCand) {
	list := st.best[sk]
	replaced := false
	for i := range list {
		if bytes.Equal(list[i].handleID, sc.handleID) {
			if sc.score > list[i].score {
				list[i] = sc
			}
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, sc)
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		if list[i].ref != list[j].ref {
			return list[i].ref < list[j].ref
		}
		return bytes.Compare(list[i].handleID, list[j].handleID) < 0
	})
	st.best[sk] = list
}

func (st *state) scoreCandidate(subjectID []byte, h *Handle) scoredCand {
	s := st.subjects[string(subjectID)]
	ev := match.Evaluate(s.Values, h.Values, st.metas)
	edge := st.edgeSupport(subjectID, h.ID)
	sc := ScoreCandidate(ev, s.Values, edge, s.Provenance, st.cfg, st.stats)
	return scoredCand{
		handleID: append([]byte(nil), h.ID...),
		ref:      h.Ref,
		score:    sc.Score,
		eval:     sc.Eval,
		reasons:  sc.Reasons,
		edge:     sc.Edge,
	}
}

// edgeSupport sums one term per mapped layer neighbor whose handle the
// candidate reaches through a canon edge with the same signature. The edge
// that seeded the candidate is one of those neighbors, so it counts once.
func (st *state) edgeSupport(subjectID, handleID []byte) float64 {
	var support float64
	add := func(sig EdgeSignature) {
		fan, ok := st.stats.FanOut[sig.Key()]
		if !ok {
			fan = st.cfg.FanOutUnknown
		}
		if fan <= st.cfg.FanOutLowMax {
			support += st.cfg.EdgeSupportLow
		} else {
			support += st.cfg.EdgeSupportHigh
		}
	}
	sk := string(subjectID)
	for _, link := range st.layerAdj[sk] {
		nk := string(link.neighbor)
		hid, ok := st.assigned[nk]
		if !ok {
			continue
		}
		// Does canon connect handleID to hid with the same signature?
		for _, ce := range st.canonAdj[string(handleID)] {
			if ce.sig.Key() == link.sig.Key() && bytes.Equal(ce.neighbor, hid) {
				add(link.sig)
				break
			}
		}
	}
	return support
}

func (st *state) walk() {
	for st.queue.Len() > 0 {
		item := heap.Pop(&st.queue).(queueItem)
		sk, hk := string(item.subjectID), string(item.handleID)
		seenKey := sk + "|" + hk
		if st.seen[seenKey] {
			continue
		}
		st.seen[seenKey] = true
		if _, taken := st.assigned[sk]; taken {
			continue
		}
		if item.score < st.cfg.AcceptScore {
			continue
		}
		if st.lostContest(sk, hk) {
			continue // one-to-one
		}
		st.assigned[sk] = append([]byte(nil), item.handleID...)
		st.handleOwner[hk] = sk
		st.pushFromAnchor(item.subjectID, item.handleID)
	}
}

// fallbackUnreachable handles Subjects no anchor reached. Property-only Rank
// picks each one's candidates; each is then scored by the same pairwise path
// the walk uses, so assignment and assessment share one scale. Assignment is
// best-first across every unreachable Subject, so a contest over one handle
// goes to the stronger match, not to whichever Subject came first.
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
		profile, ok := match.DefaultProfile(s.Kind)
		if !ok {
			continue
		}
		for _, m := range match.Rank(profile, s.Values, st.candidatesOfKind(s.Kind), st.cfg.AlternativeLimit) {
			h := st.handles[string(m.EntityID)]
			sc := st.scoreCandidate(s.ID, h)
			sc.reasons = append([]string{"property-only match"}, sc.reasons...)
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
		sk, hk := string(item.subjectID), string(item.handleID)
		if _, taken := st.assigned[sk]; taken {
			continue
		}
		if item.score < st.cfg.AcceptScore {
			continue
		}
		if st.lostContest(sk, hk) {
			continue
		}
		st.assigned[sk] = append([]byte(nil), item.handleID...)
		st.handleOwner[hk] = sk
	}
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
			for _, ce := range st.canonAdj[string(g)] {
				if ce.sig.Key() != link.sig.Key() || bytes.Equal(ce.neighbor, hid) {
					continue
				}
				h := st.handles[string(ce.neighbor)]
				if h == nil || h.Kind != s.Kind {
					continue
				}
				sc := st.scoreCandidate(s.ID, h)
				sc.viaNeighbor = append([]byte(nil), link.neighbor...)
				sc.viaSig = link.sig
				st.recordBest(sk, sc)
			}
		}
	}
}

// candidatesOfKind lists the canon handles of one kind in a stable order.
func (st *state) candidatesOfKind(kind string) []match.Candidate {
	var out []match.Candidate
	for i := range st.canon.Handles {
		h := &st.canon.Handles[i]
		if h.Kind != kind {
			continue
		}
		out = append(out, match.Candidate{EntityID: h.ID, Ref: h.Ref, Values: h.Values})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Ref != out[j].Ref {
			return out[i].Ref < out[j].Ref
		}
		return bytes.Compare(out[i].EntityID, out[j].EntityID) < 0
	})
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
				if len(sc.eval.Comparisons) == 0 {
					sc.eval = match.Evaluate(s.Values, h.Values, st.metas)
					if len(sc.reasons) == 0 {
						sc.reasons = []string{"fixed"}
					}
				}
			}
			row.Score = sc.score
			row.Reasons = sc.reasons
			row.Comparisons = comparisonsFrom(sc.eval, s.Values, st.cfg, st.stats)
			if len(sc.viaNeighbor) > 0 {
				row.Via = &Via{
					NeighborSubjectID: append([]byte(nil), sc.viaNeighbor...),
					Signature:         sc.viaSig,
				}
			}
			row.Assessment = st.band(sc.score)
			row.Alternatives = st.alts(sk, hid)
			if st.fixedSet[sk] {
				if best := st.bestNonAssigned(sk, hid); best != nil && best.score > sc.score+st.cfg.ConflictPenalty {
					row.Flags.ConflictWithFixed = true
				}
			}
		} else if decided, ok := st.held[sk]; ok {
			row.Target = decided
			row.Assessment = AssessNone
			row.Reasons = []string{"decided"}
		} else {
			row.Assessment = AssessNone
			if owner, ok := st.lostTo[sk]; ok {
				row.Flags.PossibleDuplicate = true
				row.Flags.DuplicateOf = []byte(owner)
			}
			if best := st.bestNonAssigned(sk, nil); best != nil {
				// A candidate existed but was not accepted (weak Rank or
				// one-to-one loss): Skip, never force a merge.
				row.Target = TargetSkip
				row.Score = best.score
				row.Alternatives = st.alts(sk, nil)
				row.Reasons = best.reasons
				if best.score >= st.cfg.WeakScore {
					row.Assessment = st.band(best.score)
				} else {
					row.Reasons = append([]string{"weak match"}, best.reasons...)
				}
			} else if subjectHasValues(s) {
				row.Target = TargetNew
				row.Reasons = []string{"no matching handle"}
			} else {
				row.Target = TargetSkip
				row.Reasons = []string{"unreachable or empty"}
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
		if len(idx) < 2 {
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

// flagAlikeNewRows marks two New rows of one kind whose values would clear
// the accept bar against each other: filing both mints two handles for what
// may be one entity. It only warns, so the bar is acceptance, not strong.
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
			ev := match.Evaluate(a.Values, b.Values, st.metas)
			if ScoreCandidate(ev, a.Values, 0, 1, st.cfg, st.stats).Score < st.cfg.AcceptScore {
				continue
			}
			rows[i].Flags.PossibleDuplicate = true
			rows[i].Flags.DuplicateOf = append([]byte(nil), rows[j].SubjectID...)
			break
		}
	}
}

func comparisonsFrom(ev match.Evaluation, probe match.Values, cfg Config, stats Stats) []Comparison {
	out := make([]Comparison, 0, len(ev.Comparisons))
	for _, pc := range ev.Comparisons {
		out = append(out, Comparison{
			Property:  pc.Property,
			Outcome:   pc.Outcome,
			ValueType: pc.ValueType,
			Pinned:    pc.Outcome == match.OutcomeAgree,
			Weight:    PropertyWeight(pc.Outcome, pc.ValueType, pc.Property, probe, cfg, stats),
		})
	}
	return out
}

func (st *state) pickScore(sk string, hid []byte) scoredCand {
	for _, sc := range st.best[sk] {
		if bytes.Equal(sc.handleID, hid) {
			return sc
		}
	}
	return scoredCand{handleID: hid, score: st.cfg.StrongScore, reasons: []string{"fixed"}}
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
	limit := st.cfg.AlternativeLimit
	if limit <= 0 {
		limit = 3
	}
	var out []Alternative
	for _, sc := range st.best[sk] {
		if except != nil && bytes.Equal(sc.handleID, except) {
			continue
		}
		out = append(out, Alternative{
			HandleID: append([]byte(nil), sc.handleID...),
			Ref:      sc.ref,
			Score:    sc.score,
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

func subjectHasValues(s *Subject) bool {
	for _, vs := range s.Values {
		if len(vs) > 0 {
			return true
		}
	}
	return false
}

func (st *state) band(score float64) Assessment {
	if score >= st.cfg.StrongScore {
		return AssessStrong
	}
	if score >= st.cfg.WeakScore {
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
