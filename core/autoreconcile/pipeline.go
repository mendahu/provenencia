package autoreconcile

import (
	"bytes"
	"slices"
	"sort"
	"strings"
)

// The shared reconciliation pipeline (conclusion-reconciliation.md §5). Every
// value type runs it; its module only says what "the same value" is, what
// folds into what, and how a group becomes one value.
//
//  0. Rank: strongest provenance first, then Observation id. Every tie
//     goes to the better-ranked candidate.
//  1. Admit: a value the module can't read is no evidence. A provisional
//     member's candidate is never displayed and casts no vote.
//  2. Deny: a negative eliminates the positives with the same value whose
//     provenance it is stronger than. They cast no vote.
//  3. Group each unit's values by key; fold less specific values into
//     fuller ones (the module decides).
//  4. Majority, per unit: a value with at least MinMajoritySupport Sources
//     and more than half the unit's support crowds out the rest — the ones
//     the module lets it outvote (names: spelling variants only).
//  5. Confidence, per unit: a surviving value carried only by weak
//     evidence drops when a value with non-weak evidence survived.
//  6. A candidate is displayed when every one of its units survived.
//     Displayed candidates that agree on every unit they share form one
//     value; one missing a unit joins the agreeing value with the most
//     members. A oneValue module (names) puts every displayed candidate in
//     one value instead, with every surviving value of each unit. The
//     module assembles each value.
//
// Nothing is dropped: every value that isn't displayed is still returned,
// with the reason, and every candidate gets an Outcome.

// MinMajoritySupport is the least support, in distinct Sources, a value needs
// to crowd out the others ("two out of three"). It must also hold more than
// half the unit's support.
const MinMajoritySupport = 2

// cardinality is fixed at single until Properties carry one (S9-36).
type cardinality int

const cardinalitySingle cardinality = iota

// entry is one candidate in the pipeline.
type entry struct {
	c        Candidate
	in       int // input index
	rank     int
	units    map[string]unit
	sig      string // its units, exactly
	reason   Reason // "" while still in the running
	deniedBy []byte
	vote     Vote                  // the first lost unit's vote, when outvoted
	roots    map[string]*unitValue // unit name → the value it settled on
	row      *row
}

// unitValue is one distinct value of one unit.
type unitValue struct {
	u       unit
	rank    int             // best-ranked carrier
	sources map[string]bool // Sources carrying exactly this value
	strong  bool            // some exact carrier is not weak
	into    *unitValue      // the fuller value it folds into; nil when maximal
	// Set on maximal values:
	total     map[string]bool // Sources once folded values are added
	anyStrong bool            // some carrier, folded ones included, is not weak
	fate      Reason          // ReasonKept, ReasonOutvoted or ReasonWeak
	vote      Vote            // the winner's Sources of all, when outvoted
}

// row is one auto-reconciled value under construction.
type row struct {
	members   []*entry              // supporters, in rank order
	attached  []*entry              // non-voters with this value (provisional, denied)
	settled   map[string]*unitValue // displayed rows: unit name → value
	sigs      map[string]bool       // every exact value that maps here
	displayed bool
	reason    Reason
	value     Value
	against   int
}

func reconcile(m module, candidates []Candidate, concluded *Value, _ cardinality) Result {
	res := Result{Outcomes: make([]Outcome, len(candidates))}

	order := make([]int, len(candidates))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return bytes.Compare(candidates[order[a]].ObservationID, candidates[order[b]].ObservationID) < 0
	})
	sort.SliceStable(order, func(a, b int) bool {
		return candidates[order[a]].Provenance.Stronger(candidates[order[b]].Provenance)
	})

	// 1. Admit.
	var pos, neg []*entry
	for rank, i := range order {
		c := candidates[i]
		units, ok := m.split(c.Value)
		if !ok {
			res.Outcomes[i] = Outcome{ObservationID: c.ObservationID, Reason: ReasonNoEvidence, Value: -1}
			continue
		}
		e := &entry{c: c, in: i, rank: rank, units: units, sig: signature(units)}
		if c.Negative {
			neg = append(neg, e)
			continue
		}
		pos = append(pos, e)
	}

	// 2. Deny; provisional members don't vote either.
	var voters []*entry
	for _, e := range pos {
		for _, n := range neg {
			if n.sig == e.sig && n.c.Provenance.Stronger(e.c.Provenance) {
				e.reason, e.deniedBy = ReasonDenied, n.c.ObservationID
				break
			}
		}
		if e.reason == "" && e.c.Provisional {
			e.reason = ReasonProvisional
		}
		if e.reason == "" {
			voters = append(voters, e)
		}
	}

	// 3–5. Per unit: group, fold, majority, confidence.
	byName := map[string][]*unitValue{}
	for _, e := range voters {
		e.roots = map[string]*unitValue{}
		for _, name := range unitNames(e.units) {
			u := e.units[name]
			var v *unitValue
			for _, w := range byName[name] {
				if w.u.key == u.key {
					v = w
					break
				}
			}
			if v == nil {
				v = &unitValue{u: u, rank: e.rank, sources: map[string]bool{}}
				byName[name] = append(byName[name], v)
			}
			v.sources[sourceKey(e.c)] = true
			v.strong = v.strong || !e.c.Provenance.Weak()
			e.roots[name] = v
		}
	}
	for _, name := range sortedKeys(byName) {
		vals := byName[name]
		foldValues(m, name, vals)
		elect(m, name, vals)
		dropWeak(vals)
	}

	// 6. Each voter's fate: displayed when every unit survived.
	for _, e := range voters {
		folded, fate := false, ReasonKept
		for _, name := range unitNames(e.units) {
			v := e.roots[name]
			if v.into != nil {
				folded = true
				v = root(v)
				e.roots[name] = v
			}
			switch {
			case v.fate == ReasonOutvoted:
				if fate != ReasonOutvoted {
					e.vote = v.vote
				}
				fate = ReasonOutvoted
			case v.fate == ReasonWeak && fate == ReasonKept:
				fate = ReasonWeak
			}
		}
		if fate == ReasonKept && folded {
			fate = ReasonFolded
		}
		e.reason = fate
	}

	// Displayed rows.
	var shown []*entry
	for _, e := range voters {
		if e.reason == ReasonKept || e.reason == ReasonFolded {
			shown = append(shown, e)
		}
	}
	var rows []*row
	if m.oneValue() {
		rows = groupAll(shown)
	} else {
		rows = groupDisplayed(shown)
	}
	bySig := map[string]*row{}
	for _, r := range rows {
		var members []Candidate
		for _, e := range r.members {
			members = append(members, e.c)
		}
		r.value = m.assemble(members, r.settledUnits())
		r.reason = ReasonKept
		r.displayed = true
		if units, ok := m.split(r.value); ok {
			r.sigs[signature(units)] = true
		}
		for s := range r.sigs {
			bySig[s] = r
		}
	}

	// Everything else keeps a row: attached to the displayed row with its
	// value, or grouped by exact value with its best-ranked member's reason.
	var hidden []*row
	for _, e := range pos {
		if e.row != nil {
			continue
		}
		if r, ok := bySig[e.sig]; ok && r.displayed && (e.reason == ReasonDenied || e.reason == ReasonProvisional) {
			r.attached = append(r.attached, e)
			e.row = r
			continue
		}
		r, ok := bySig[e.sig]
		if !ok || r.displayed {
			r = &row{sigs: map[string]bool{e.sig: true}, reason: e.reason, value: e.c.Value}
			bySig[e.sig] = r
			hidden = append(hidden, r)
		}
		r.members = append(r.members, e)
		e.row = r
	}

	// Negatives count against every row with their value.
	for _, n := range neg {
		res.Outcomes[n.in] = Outcome{ObservationID: n.c.ObservationID, Reason: ReasonAgainst, Value: -1}
		for _, r := range append(append([]*row(nil), rows...), hidden...) {
			if r.sigs[n.sig] {
				r.against++
			}
		}
	}

	ordered := append(orderRows(rows), orderRows(hidden)...)
	index := map[*row]int{}
	for i, r := range ordered {
		index[r] = i
		res.Values = append(res.Values, r.reconciled())
	}
	for _, e := range pos {
		res.Outcomes[e.in] = Outcome{ObservationID: e.c.ObservationID, Reason: e.reason, Value: index[e.row], DeniedBy: e.deniedBy}
		if e.reason == ReasonOutvoted {
			res.Outcomes[e.in].Vote = e.vote
		}
	}
	for _, n := range neg {
		for _, r := range ordered {
			if r.sigs[n.sig] {
				res.Outcomes[n.in].Value = index[r]
				break
			}
		}
	}

	if concluded != nil {
		applyConcluded(m, &res, ordered, *concluded)
	}
	sort.SliceStable(res.Outcomes, func(i, j int) bool {
		return bytes.Compare(res.Outcomes[i].ObservationID, res.Outcomes[j].ObservationID) < 0
	})
	return res
}

// foldValues points each non-maximal value at the maximal value it folds
// into (the best supported, then the best ranked) and totals each maximal
// value's Sources and strength.
func foldValues(m module, name string, vals []*unitValue) {
	var maximal []*unitValue
	for _, v := range vals {
		isMax := true
		for _, u := range vals {
			if u != v && m.fold(name, v.u, u.u) {
				isMax = false
				break
			}
		}
		if isMax {
			maximal = append(maximal, v)
		}
	}
	for _, v := range vals {
		target := v
		for _, u := range maximal {
			if u == v || !m.fold(name, v.u, u.u) {
				continue
			}
			if target == v || len(u.sources) > len(target.sources) ||
				(len(u.sources) == len(target.sources) && u.rank < target.rank) {
				target = u
			}
		}
		if target != v {
			v.into = target
		}
	}
	for _, v := range maximal {
		v.total = map[string]bool{}
	}
	for _, v := range vals {
		r := root(v)
		for s := range v.sources {
			r.total[s] = true
		}
		r.anyStrong = r.anyStrong || v.strong
	}
}

// elect: a maximal value with at least MinMajoritySupport Sources and more
// than half the unit's support wins; the others the module lets it outvote
// are outvoted. Otherwise every value stays.
func elect(m module, name string, vals []*unitValue) {
	var all int
	var best *unitValue
	for _, v := range vals {
		if v.into != nil {
			continue
		}
		all += len(v.total)
		if best == nil || len(v.total) > len(best.total) {
			best = v
		}
	}
	win := best != nil && len(best.total) >= MinMajoritySupport && 2*len(best.total) > all
	for _, v := range vals {
		if v.into != nil {
			continue
		}
		v.fate = ReasonKept
		if win && v != best && m.outvotes(name, best.u, v.u) {
			v.fate = ReasonOutvoted
			v.vote = Vote{Support: len(best.total), Of: all}
		}
	}
}

// dropWeak: a surviving value carried only by weak evidence is weak when a
// value with non-weak evidence survived.
func dropWeak(vals []*unitValue) {
	strong := false
	for _, v := range vals {
		strong = strong || (v.into == nil && v.fate == ReasonKept && v.anyStrong)
	}
	if !strong {
		return
	}
	for _, v := range vals {
		if v.into == nil && v.fate == ReasonKept && !v.anyStrong {
			v.fate = ReasonWeak
		}
	}
}

func root(v *unitValue) *unitValue {
	for v.into != nil {
		v = v.into
	}
	return v
}

// groupDisplayed groups displayed candidates that agree on every unit they
// share. Fuller candidates place first; one missing a unit joins the
// agreeing row with the most members, then the best ranked.
func groupDisplayed(shown []*entry) []*row {
	order := append([]*entry(nil), shown...)
	sort.SliceStable(order, func(i, j int) bool { return len(order[i].units) > len(order[j].units) })
	var rows []*row
	for _, e := range order {
		var into *row
		for _, r := range rows {
			if !agrees(r.settled, e.roots) {
				continue
			}
			if into == nil || len(r.members) > len(into.members) ||
				(len(r.members) == len(into.members) && r.members[0].rank < into.members[0].rank) {
				into = r
			}
		}
		if into == nil {
			into = &row{settled: map[string]*unitValue{}, sigs: map[string]bool{}}
			rows = append(rows, into)
		}
		into.members = append(into.members, e)
		sort.SliceStable(into.members, func(i, j int) bool { return into.members[i].rank < into.members[j].rank })
		for name, v := range e.roots {
			into.settled[name] = v
		}
		into.sigs[e.sig] = true
		e.row = into
	}
	return rows
}

// groupAll puts every displayed candidate in one row, for a module whose
// value is one structure.
func groupAll(shown []*entry) []*row {
	if len(shown) == 0 {
		return nil
	}
	r := &row{sigs: map[string]bool{}}
	for _, e := range shown {
		r.members = append(r.members, e)
		r.sigs[e.sig] = true
		e.row = r
	}
	return []*row{r}
}

// settledUnits is, per unit name, the distinct values the row's members
// settled on: best supported first, then best ranked.
func (r *row) settledUnits() map[string][]unit {
	byName := map[string][]*unitValue{}
	for _, e := range r.members {
		for name, v := range e.roots {
			if !slices.Contains(byName[name], v) {
				byName[name] = append(byName[name], v)
			}
		}
	}
	out := make(map[string][]unit, len(byName))
	for name, vals := range byName {
		sort.SliceStable(vals, func(i, j int) bool {
			if len(vals[i].total) != len(vals[j].total) {
				return len(vals[i].total) > len(vals[j].total)
			}
			return vals[i].rank < vals[j].rank
		})
		for _, v := range vals {
			out[name] = append(out[name], v.u)
		}
	}
	return out
}

func agrees(settled, roots map[string]*unitValue) bool {
	for name, v := range roots {
		if have, ok := settled[name]; ok && have != v {
			return false
		}
	}
	return true
}

// orderRows: support descending, then the best-ranked member.
func orderRows(rows []*row) []*row {
	sort.SliceStable(rows, func(i, j int) bool {
		si, sj := rows[i].support(), rows[j].support()
		if si != sj {
			return si > sj
		}
		return rows[i].bestRank() < rows[j].bestRank()
	})
	return rows
}

func (r *row) support() int {
	s := map[string]bool{}
	for _, e := range r.members {
		s[sourceKey(e.c)] = true
	}
	return len(s)
}

func (r *row) bestRank() int {
	best := -1
	for _, e := range r.members {
		if best < 0 || e.rank < best {
			best = e.rank
		}
	}
	return best
}

func (r *row) reconciled() ReconciledValue {
	c := ReconciledValue{Value: r.value, Support: r.support(), Against: r.against, Reason: r.reason}
	for _, e := range r.members {
		c.ObservationIDs = append(c.ObservationIDs, e.c.ObservationID)
	}
	sort.Slice(c.ObservationIDs, func(i, j int) bool {
		return bytes.Compare(c.ObservationIDs[i], c.ObservationIDs[j]) < 0
	})
	return c
}

// applyConcluded puts a concluded value at rank 1: it takes over the value
// with the same value, or leads alone with no support.
func applyConcluded(m module, res *Result, ordered []*row, concluded Value) {
	top := ReconciledValue{Value: concluded, Reason: ReasonKept}
	at := -1
	if units, ok := m.split(concluded); ok {
		sig := signature(units)
		for i, r := range ordered {
			if r.sigs[sig] {
				at = i
				break
			}
		}
	}
	shift := func(i int) int { return i + 1 }
	if at >= 0 {
		c := res.Values[at]
		top.ObservationIDs, top.Support, top.Against = c.ObservationIDs, c.Support, c.Against
		res.Values = append(res.Values[:at:at], res.Values[at+1:]...)
		shift = func(i int) int {
			switch {
			case i == at:
				return 0
			case i < at:
				return i + 1
			}
			return i
		}
	}
	res.Values = append([]ReconciledValue{top}, res.Values...)
	for i := range res.Outcomes {
		if res.Outcomes[i].Value >= 0 {
			res.Outcomes[i].Value = shift(res.Outcomes[i].Value)
		}
	}
	res.Concluded = true
}

func sourceKey(c Candidate) string {
	if len(c.SourceID) > 0 {
		return "s" + string(c.SourceID)
	}
	return "o" + string(c.ObservationID)
}

// signature is a set of units, exactly: two candidates with equal signatures
// carry the same value.
func signature(units map[string]unit) string {
	var b strings.Builder
	for _, name := range unitNames(units) {
		b.WriteString(name)
		b.WriteByte(0)
		b.WriteString(units[name].key)
		b.WriteByte(0)
	}
	return b.String()
}

func unitNames(units map[string]unit) []string {
	names := make([]string, 0, len(units))
	for n := range units {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
