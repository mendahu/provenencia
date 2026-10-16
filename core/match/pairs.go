package match

import (
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

// Pair selection: which one value on each side speaks for a Feature, and how
// much it counts. A handle's cached values carry the support the
// auto-reconciler counted for them, so a value seen on five Sources outweighs
// one seen on one; a negative Observation counts against the value it denies.
// Score and compareProperty both read their pair through pairValues.

// ValueRef names one value that took part in a comparison.
type ValueRef struct {
	Label string
	// ObservationID is the Observation a probe value was loaded from; empty
	// for a cached handle value, which is a cluster of Observations.
	ObservationID []byte
	Negative      bool
}

// PairResult is the outcome of comparing two sides' values of one Property.
type PairResult struct {
	// Similarity is the resemblance times both values' weights, 0…1; 0 is a
	// contradiction.
	Similarity float64
	// Discount is the share of points the matched values' frequency cost;
	// 0 when the values contradict.
	Discount float64
	// Comparable is false when no pair is comparable (a side is empty or
	// holds only values that cannot speak).
	Comparable bool
	// Probe and Candidate are the two values behind the result.
	Probe, Candidate ValueRef
}

// SetEvidence fills the fields the auto-reconciler cached for a handle value.
func (v *Value) SetEvidence(support, against, rank int, reason string) {
	v.Counted = true
	v.Support = support
	v.Against = against
	v.Rank = rank
	v.Reason = reason
	v.Denied = reason == string(autoreconcile.ReasonDenied)
}

// HasPositive reports whether any value is positive.
func (vs Values) HasPositive() bool {
	for _, list := range vs {
		for _, v := range list {
			if !v.Negative {
				return true
			}
		}
	}
	return false
}

// valueType is the autoreconcile value type the fields a Value sets imply.
func (v Value) valueType() string {
	switch {
	case v.Name != nil:
		return properties.ValueTypeName
	case v.Date != nil:
		return properties.ValueTypeDate
	case v.Term != "":
		return properties.ValueTypeTerm
	case v.HasInteger:
		return properties.ValueTypeInteger
	case v.HasText:
		return properties.ValueTypeText
	}
	return ""
}

// Auto is the value in the shape autoreconcile.Compatible reads. A term is
// keyed by its key: loaders differ on whether they carry the catalog id.
func (v Value) Auto() autoreconcile.Value {
	a := autoreconcile.Value{
		Text: v.Text, HasText: v.HasText, Integer: v.Integer, HasInteger: v.HasInteger,
		TermKey: v.Term, Neutral: v.Neutral, Date: v.Date, Name: v.Name,
	}
	if v.Term != "" {
		a.TermID = []byte(v.Term)
	} else {
		a.TermID = append([]byte(nil), v.TermID...)
	}
	return a
}

// Label is the value as text, for naming it in a Reason.
func (v Value) Label() string {
	switch {
	case v.Name != nil:
		if v.Name.Form != "" {
			return v.Name.Form
		}
		parts := make([]string, 0, len(v.Name.Parts))
		for _, p := range v.Name.Parts {
			parts = append(parts, p.Value)
		}
		return strings.Join(parts, " ")
	case v.Date != nil:
		return datevalues.CompactDisplay(*v.Date)
	case v.Term != "":
		return v.Term
	case v.HasInteger:
		return strconv.FormatInt(v.Integer, 10)
	}
	return v.Text
}

func (v Value) ref() ValueRef {
	return ValueRef{Label: v.Label(), ObservationID: v.ObservationID, Negative: v.Negative}
}

func compatible(a, b Value) bool {
	t := a.valueType()
	return t != "" && t == b.valueType() && autoreconcile.Compatible(t, a.Auto(), b.Auto())
}

// effective is the support a value has left after the negatives against it.
// A displayed value is floored at 1 so a handle's only displayed value
// always votes; a value with no cache evidence counts 1.
func (v Value) effective() int {
	if !v.Counted {
		return 1
	}
	floor := 0
	if v.Reason == string(autoreconcile.ReasonKept) {
		floor = 1
	}
	return max(v.Support-v.Against, floor)
}

func (v Value) displayed() bool {
	return !v.Counted || v.Reason == string(autoreconcile.ReasonKept)
}

// votes is whether the value may speak: positive, standing, evidence-bearing.
func (v Value) votes() bool {
	return !v.Negative && !v.Denied && v.Reason != string(autoreconcile.ReasonProvisional) &&
		!v.Neutral && v.effective() > 0
}

func (v Value) plain() bool { return !v.Counted && !v.Negative && !v.Denied }

func plainValues(vs []Value) bool {
	for i := range vs {
		if !vs[i].plain() {
			return false
		}
	}
	return true
}

type voteInfo struct {
	eff       int
	voting    bool
	displayed bool
	leader    bool
	weight    float64
}

// weigh finds a side's voting values, its leaders, and each value's weight:
// effective support over the leaders' (D-3), at most 1.
func weigh(vs []Value) []voteInfo {
	info := make([]voteInfo, len(vs))
	var top, topShown int
	anyShown := false
	for i := range vs {
		if !vs[i].votes() {
			continue
		}
		e, shown := vs[i].effective(), vs[i].displayed()
		info[i] = voteInfo{eff: e, voting: true, displayed: shown}
		top = max(top, e)
		if shown {
			anyShown = true
			topShown = max(topShown, e)
		}
	}
	leaderEff := top
	if anyShown {
		leaderEff = topShown
	}
	for i := range info {
		if !info[i].voting {
			continue
		}
		info[i].leader = info[i].eff == leaderEff && (info[i].displayed || !anyShown)
		info[i].weight = min(1, float64(info[i].eff)/float64(leaderEff))
	}
	return info
}

// order compares two values of one side for the equal-key tiebreak: displayed
// first, more support, label, then rank (cache) or Observation id. Negative
// when i comes before j; 0 leaves input order.
func order(vs []Value, info []voteInfo, i, j int) int {
	if info[i].displayed != info[j].displayed {
		if info[i].displayed {
			return -1
		}
		return 1
	}
	if info[i].eff != info[j].eff {
		if info[i].eff > info[j].eff {
			return -1
		}
		return 1
	}
	if c := strings.Compare(vs[i].Label(), vs[j].Label()); c != 0 {
		return c
	}
	if vs[i].Counted && vs[j].Counted {
		if vs[i].Rank != vs[j].Rank {
			if vs[i].Rank < vs[j].Rank {
				return -1
			}
			return 1
		}
		return 0
	}
	return strings.Compare(string(vs[i].ObservationID), string(vs[j].ObservationID))
}

type pairIdx struct{ i, j int }

// pairSet picks, among candidate pairs offered to it, the one that sorts
// first by the tiebreak (the probe value decides first, then the candidate's).
type pairSet struct {
	probe, cand  []Value
	pinfo, cinfo []voteInfo
	best         pairIdx
	found        bool
}

func (s *pairSet) offer(i, j int) {
	if !s.found || s.before(i, j) {
		s.best, s.found = pairIdx{i, j}, true
	}
}

func (s *pairSet) before(i, j int) bool {
	if i != s.best.i {
		if c := order(s.probe, s.pinfo, i, s.best.i); c != 0 {
			return c < 0
		}
	}
	if j != s.best.j {
		return order(s.cand, s.cinfo, j, s.best.j) < 0
	}
	return false
}

// pairValues compares a probe's values of a Property with a candidate's. In
// order: a negative that is compatible with a leader, a value compatible with
// a denied value nothing else on that side supports, or leaders that all
// disagree, is a contradiction; otherwise the best weighted resemblance
// among voting pairs speaks, and with none the result is not comparable.
func pairValues(cmp Comparer, probe, candidate []Value) PairResult {
	if plainValues(probe) && plainValues(candidate) {
		return plainPair(cmp, probe, candidate)
	}
	s := &pairSet{probe: probe, cand: candidate, pinfo: weigh(probe), cinfo: weigh(candidate)}

	// R1: a negative against a leader on the other side.
	for i := range probe {
		for j := range candidate {
			if (probe[i].Negative && s.cinfo[j].leader || candidate[j].Negative && s.pinfo[i].leader) &&
				compatible(probe[i], candidate[j]) {
				s.offer(i, j)
			}
		}
	}
	if s.found {
		return s.contradiction()
	}

	// R2: a value the other side's evidence refutes and nothing there backs.
	for i := range probe {
		if !probe[i].Negative {
			for _, j := range refutedBy(probe[i], candidate, s.cinfo) {
				s.offer(i, j)
			}
		}
	}
	for j := range candidate {
		if !candidate[j].Negative {
			for _, i := range refutedBy(candidate[j], probe, s.pinfo) {
				s.offer(i, j)
			}
		}
	}
	if s.found {
		return s.contradiction()
	}

	// R3 and the credit, over the voting pairs.
	var (
		leaderComparable, leaderPositive bool
		leaders                          = &pairSet{probe: probe, cand: candidate, pinfo: s.pinfo, cinfo: s.cinfo}
		credit                           = &pairSet{probe: probe, cand: candidate, pinfo: s.pinfo, cinfo: s.cinfo}
		top, topSim, topDisc             float64
	)
	for i := range probe {
		if !s.pinfo[i].voting {
			continue
		}
		for j := range candidate {
			if !s.cinfo[j].voting {
				continue
			}
			sim, disc, ok := CompareDiscount(cmp, probe[i], candidate[j])
			if !ok {
				continue
			}
			if s.pinfo[i].leader && s.cinfo[j].leader {
				leaderComparable = true
				if sim > 0 {
					leaderPositive = true
				} else {
					leaders.offer(i, j)
				}
			}
			w := s.pinfo[i].weight * s.cinfo[j].weight
			score := sim * w * (1 - disc)
			switch {
			case !credit.found || score > top:
				credit.best, credit.found = pairIdx{i, j}, true
				top, topSim, topDisc = score, sim*w, disc
			case score == top && credit.before(i, j):
				credit.best = pairIdx{i, j}
				topSim, topDisc = sim*w, disc
			}
		}
	}
	if leaderComparable && !leaderPositive {
		return leaders.contradiction()
	}
	if !credit.found {
		return PairResult{}
	}
	if topSim == 0 {
		topDisc = 0
	}
	return PairResult{
		Similarity: topSim, Discount: topDisc, Comparable: true,
		Probe: probe[credit.best.i].ref(), Candidate: candidate[credit.best.j].ref(),
	}
}

// refutedBy lists the denied values of a side that v is compatible with,
// or nothing when a voting value of that side is compatible with v too.
func refutedBy(v Value, others []Value, info []voteInfo) []int {
	var hits []int
	denied := false
	for k := range others {
		denied = denied || others[k].Denied
	}
	if !denied {
		return nil
	}
	for k := range others {
		if !compatible(v, others[k]) {
			continue
		}
		if info[k].voting {
			return nil
		}
		if others[k].Denied {
			hits = append(hits, k)
		}
	}
	return hits
}

func (s *pairSet) contradiction() PairResult {
	return PairResult{
		Comparable: true,
		Probe:      s.probe[s.best.i].ref(), Candidate: s.cand[s.best.j].ref(),
	}
}

// plainPair is pairValues for values with no cache evidence and no
// negatives: every value weighs 1 and every pair is a leader pair, so the
// best pair is the comparable one with the highest similarity × (1 − discount),
// the first on a tie. No allocation until the reported pair is labeled.
func plainPair(cmp Comparer, probe, candidate []Value) PairResult {
	var (
		top, sim, disc float64
		bi, bj         int
		ok             bool
	)
	for i := range probe {
		for j := range candidate {
			s, d, comparable := CompareDiscount(cmp, probe[i], candidate[j])
			if !comparable {
				continue
			}
			if p := s * (1 - d); !ok || p > top {
				top, sim, disc, ok, bi, bj = p, s, d, true, i, j
			}
		}
	}
	if !ok {
		return PairResult{}
	}
	if sim == 0 {
		disc = 0
	}
	return PairResult{Similarity: sim, Discount: disc, Comparable: true, Probe: probe[bi].ref(), Candidate: candidate[bj].ref()}
}
