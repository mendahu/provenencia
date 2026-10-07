package graphalign

import (
	"bytes"
	"container/heap"
	"sort"
	"strconv"

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
	return st.proposal()
}

type state struct {
	cfg    Config
	layer  Layer
	canon  Canon
	stats  Stats
	metas  []match.PropertyMeta

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
	handleID []byte
	ref      string
	score    float64
	eval     match.Evaluation
	reasons  []string
	edge     float64
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
	for sk, hid := range st.assigned {
		st.pushFromAnchor([]byte(sk), hid)
	}
}

func (st *state) pushFromAnchor(subjectID, handleID []byte) {
	sk := string(subjectID)
	for _, link := range st.layerAdj[sk] {
		nk := string(link.neighbor)
		if _, taken := st.assigned[nk]; taken {
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
			st.enqueue(ns.ID, h, link.sig)
		}
	}
}

func (st *state) enqueue(subjectID []byte, h *Handle, via EdgeSignature) {
	key := string(subjectID) + "|" + string(h.ID)
	if st.seen[key] {
		return
	}
	sc := st.scoreCandidate(subjectID, h, via)
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
		return list[i].ref < list[j].ref
	})
	st.best[sk] = list
}

func (st *state) scoreCandidate(subjectID []byte, h *Handle, via EdgeSignature) scoredCand {
	s := st.subjects[string(subjectID)]
	ev := match.Evaluate(s.Values, h.Values, st.metas)
	node, reasons := st.nodeScore(ev, s)
	edge := st.edgeSupport(subjectID, h.ID, via)
	if edge > 0 {
		reasons = append(reasons, "edge support")
	}
	prov := s.Provenance
	if prov <= 0 {
		prov = 1
	}
	return scoredCand{
		handleID: append([]byte(nil), h.ID...),
		ref:      h.Ref,
		score:    (node + edge) * prov,
		eval:     ev,
		reasons:  reasons,
		edge:     edge,
	}
}

func (st *state) nodeScore(ev match.Evaluation, s *Subject) (float64, []string) {
	var score float64
	var reasons []string
	for _, pc := range ev.Comparisons {
		switch pc.Outcome {
		case match.OutcomeAgree:
			u := st.uFor(pc.Property.Key, s.Values[pc.Property])
			m := st.cfg.mFor(pc.ValueType)
			w := logOdds(m, st.cfg.uOr(u))
			score += w
			reasons = append(reasons, "agree "+pc.Property.Key)
		case match.OutcomeConflict:
			score -= st.cfg.ConflictPenalty
			reasons = append(reasons, "conflict "+pc.Property.Key)
		}
	}
	return score, reasons
}

func (st *state) uFor(propertyKey string, vals []match.Value) float64 {
	if st.stats.ValueFreq == nil {
		return 0
	}
	byVal := st.stats.ValueFreq[propertyKey]
	if byVal == nil || len(vals) == 0 {
		return 0
	}
	// Use the first carrying text/term as the frequency key.
	for _, v := range vals {
		k := valueKey(v)
		if k == "" {
			continue
		}
		if u, ok := byVal[k]; ok {
			return u
		}
	}
	return 0
}

func valueKey(v match.Value) string {
	if v.HasText {
		return v.Text
	}
	if v.Term != "" {
		return v.Term
	}
	if v.HasInteger {
		return strconv.FormatInt(v.Integer, 10)
	}
	return ""
}

func (st *state) edgeSupport(subjectID, handleID []byte, via EdgeSignature) float64 {
	// Support from the seeding edge plus any other mapped neighbor that
	// lines up with a corresponding canon edge.
	var support float64
	add := func(sig EdgeSignature) {
		fan := 2.0
		if st.stats.FanOut != nil {
			if f, ok := st.stats.FanOut[sig.Key()]; ok {
				fan = f
			}
		}
		if fan <= st.cfg.FanOutLowMax {
			support += st.cfg.EdgeSupportLow
		} else {
			support += st.cfg.EdgeSupportHigh
		}
	}
	if via.BridgeType != "" {
		add(via)
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
		if owner, ok := st.handleOwner[hk]; ok && owner != sk {
			continue // one-to-one
		}
		if item.score < st.cfg.AcceptScore {
			continue
		}
		st.assigned[sk] = append([]byte(nil), item.handleID...)
		st.handleOwner[hk] = sk
		st.pushFromAnchor(item.subjectID, item.handleID)
	}
}

func (st *state) fallbackUnreachable() {
	for sk, s := range st.subjects {
		if _, ok := st.assigned[sk]; ok {
			continue
		}
		profile, ok := match.DefaultProfile(s.Kind)
		if !ok {
			continue
		}
		var cands []match.Candidate
		for _, h := range st.handles {
			if h.Kind != s.Kind {
				continue
			}
			if owner, taken := st.handleOwner[string(h.ID)]; taken && owner != sk {
				continue
			}
			cands = append(cands, match.Candidate{
				EntityID: h.ID, Ref: h.Ref, Values: h.Values,
			})
		}
		ranked := match.Rank(profile, s.Values, cands, st.cfg.AlternativeLimit)
		for _, m := range ranked {
			// Rank uses point weights; store as-is for alternatives. Assignment
			// uses profile.MinScore (not Align AcceptScore).
			st.recordBest(sk, scoredCand{
				handleID: append([]byte(nil), m.EntityID...),
				ref:      m.Ref,
				score:    m.Score,
				reasons:  []string{"property-only match"},
			})
		}
		if len(ranked) == 0 || ranked[0].Score < profile.MinScore {
			continue
		}
		top := ranked[0]
		if owner, taken := st.handleOwner[string(top.EntityID)]; taken && owner != sk {
			continue
		}
		st.assigned[sk] = append([]byte(nil), top.EntityID...)
		st.handleOwner[string(top.EntityID)] = sk
	}
}

func (st *state) proposal() Proposal {
	var rows []Row
	// Stable subject order by Ref then ID.
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
	for _, s := range subs {
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
			row.Comparisons = comparisonsFrom(sc.eval)
			row.Assessment = st.band(sc.score, st.fixedSet[sk], true)
			row.Alternatives = st.alts(sk, hid)
			if st.fixedSet[sk] {
				if best := st.bestNonAssigned(sk, hid); best != nil && best.score > sc.score+st.cfg.ConflictPenalty {
					row.Flags.ConflictWithFixed = true
				}
			}
		} else {
			row.Assessment = AssessNone
			if best := st.bestNonAssigned(sk, nil); best != nil {
				// A candidate existed but was not accepted (weak Rank or
				// one-to-one loss): Skip, never force a merge.
				row.Target = TargetSkip
				row.Score = best.score
				row.Alternatives = st.alts(sk, nil)
				row.Reasons = best.reasons
				if best.score >= st.cfg.WeakScore {
					row.Assessment = st.band(best.score, false, false)
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
	// Duplicate flag: two rows wanting same handle shouldn't happen if
	// one-to-one held; mark if best lists collide.
	ownerCount := map[string]int{}
	for _, r := range rows {
		if r.Target == TargetHandle && len(r.HandleID) > 0 {
			ownerCount[string(r.HandleID)]++
		}
	}
	for i := range rows {
		if rows[i].Target == TargetHandle && ownerCount[string(rows[i].HandleID)] > 1 {
			rows[i].Flags.PossibleDuplicate = true
		}
	}
	return Proposal{Rows: rows}
}

func comparisonsFrom(ev match.Evaluation) []Comparison {
	out := make([]Comparison, 0, len(ev.Comparisons))
	for _, pc := range ev.Comparisons {
		out = append(out, Comparison{
			Property:  pc.Property,
			Outcome:   pc.Outcome,
			ValueType: pc.ValueType,
			Pinned:    pc.Outcome == match.OutcomeAgree,
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

func (st *state) band(score float64, fixed, assigned bool) Assessment {
	if fixed {
		return AssessStrong
	}
	if !assigned {
		if score >= st.cfg.StrongScore {
			return AssessStrong
		}
		if score >= st.cfg.WeakScore {
			return AssessWeak
		}
		return AssessNone
	}
	if score >= st.cfg.StrongScore {
		return AssessStrong
	}
	if score >= st.cfg.WeakScore {
		return AssessWeak
	}
	return AssessNone
}

// Priority queue of candidates (best score first; ties by ref then id).
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
	return bytes.Compare(h[i].handleID, h[j].handleID) < 0
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
