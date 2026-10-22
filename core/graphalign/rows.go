package graphalign

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// contested reports whether a comparison behind the pair conflicts, whatever
// its band (RowFlags.Contested).
func (c *cand) contested() bool {
	return c.result.Conflicts() || slices.ContainsFunc(c.links, func(l segmentcompare.Result) bool {
		return l.Outcome == propertycompare.OutcomeConflict
	})
}

func (st *state) proposal() Proposal {
	var rows []Row
	for _, s := range st.sortedSubjects() {
		sk := string(s.ID)
		row := Row{SubjectID: append([]byte(nil), s.ID...), Kind: s.Kind}
		var reported *cand
		if hid, ok := st.placed[sk]; ok {
			row.Target = TargetHandle
			row.HandleID = append([]byte(nil), hid...)
			sc := st.cands[sk][string(hid)]
			reported = sc
			row.HandleRef = st.handles[string(hid)].Ref
			row.Score = sc.score
			row.Comparisons = comparisonsFrom(sc.result, s.Kind, st.cfg.Subject)
			via, viaKey := st.viaOf(sk, sc)
			switch {
			case st.fixedSet[sk]:
				row.Reason = ReasonDecided
			case len(via) > 0:
				row.Reason = ReasonVia
				row.Via = &Via{
					NeighborSubjectID: append([]byte(nil), via...),
					Key:               viaKey,
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
				reported = best
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
		if reported != nil {
			row.Flags.Contested = reported.contested()
		}
		if from, ok := st.releasedFrom[sk]; ok && !(row.Target == TargetHandle && bytes.Equal(row.HandleID, from)) {
			mv := st.move[sk]
			mv.NeighborSubjectID = append([]byte(nil), mv.NeighborSubjectID...)
			row.Move = &mv
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
// disagrees when no neighbor pair is the same subject or lined up as unplaced
// neighbors (layer 5 reads their own properties). A directed key carries its end, so a man's son and
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
				placed := segmentcompare.Together
				if !same {
					placed = segmentcompare.Unplaced
				}
				res := st.compare(la.key, sa.Kind, sa.Values, lb.key, sb.Kind, sb.Values, placed, own)
				if !same && res.Outcome != propertycompare.OutcomeAgree && res.Outcome != propertycompare.OutcomePartial {
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
func comparisonsFrom(r subjectcompare.Result, kind string, rules subjectcompare.Rules) []Comparison {
	out := make([]Comparison, 0, len(r.Properties))
	for _, pr := range r.Properties {
		out = append(out, Comparison{
			PropertyResult: pr,
			Derived:        rules.Derived(vocabkey.SubjectType(kind), pr.Key),
		})
	}
	return out
}

func (st *state) bestNonAssigned(sk string, except []byte) *cand {
	for _, c := range st.sortedCands(sk) {
		if except != nil && bytes.Equal(c.handleID, except) {
			continue
		}
		return c
	}
	return nil
}

func (st *state) alts(sk string, except []byte) []Alternative {
	kind := st.subjects[sk].Kind
	limit := st.cfg.Alignment.Alternatives
	var out []Alternative
	for _, c := range st.sortedCands(sk) {
		if except != nil && bytes.Equal(c.handleID, except) {
			continue
		}
		out = append(out, st.alternative(sk, c, kind))
		if len(out) >= limit {
			break
		}
	}
	return out
}

// alternative is one selectable record: its band, and the sentence that
// explains that band. A via walk wins; otherwise the strongest agreeing
// property; otherwise there was too little to name.
func (st *state) alternative(sk string, c *cand, kind string) Alternative {
	alt := Alternative{
		HandleID:   append([]byte(nil), c.handleID...),
		Ref:        c.ref,
		Score:      c.score,
		Assessment: st.band(c.score),
		Reason:     ReasonWeak,
	}
	if via, _ := st.viaOf(sk, c); len(via) > 0 {
		alt.Reason = ReasonVia
		alt.ViaNeighbor = append([]byte(nil), via...)
		return alt
	}
	// Below the weak bar there is not enough to name an agreement.
	if alt.Assessment == AssessNone {
		return alt
	}
	prop := strongestAgreement(comparisonsFrom(c.result, kind, st.cfg.Subject))
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
