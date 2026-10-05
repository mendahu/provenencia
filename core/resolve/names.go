package resolve

import (
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/mendahu/provenencia/core/database/namevalues"
)

// The name auto-reconciler (S9-13): a handle's name candidates in, the name a
// Reconciliation Claim would conclude out, built from structured parts by
// eliminating candidates in passes. It is display policy and may be bold: a
// researcher overrides it with a Reconciliation Claim.
//
// It is format-agnostic. A part type is only an identifier: parts are
// compared with parts of the same type, and no type behaves differently from
// another (a surname is no different from a given name or a prefix). `form`
// is never read; it is a transcription. A name with no parts carries nothing
// to reconcile and is dropped.
//
// For each part type, independently, a candidate's value is its ordered list
// of parts of that type, each part one normalized unit:
//
//  1. Exact: equal values group; support is how many candidates carry one.
//  2. Subsumption: a value folds into a fuller value when its parts map, in
//     order, onto a subsequence of the fuller one's, each equal or an initial
//     of it ([J] → [James]; [James] → [James, Kenneth]). Its support moves
//     there. When it fits several, it folds into the one with the most
//     support, then the best ranked.
//  3. Majority: a value with at least MinMajoritySupport and more than half
//     the type's support crowds out the rest. Otherwise every value survives.
//
// (S9-14 adds a confidence pass and swaps the rank order for provenance.)
//
// A candidate survives when every one of its values survived. Survivors
// group into clusters that agree on every type they share; a candidate
// missing a type joins the agreeing cluster with the most members. Each
// cluster's value is assembled from the fullest value of each type.
// Eliminated candidates are not hidden: they group by exact parts and rank
// after the survivors, so the result reads mixed when candidates disagreed.

// MinMajoritySupport is the least support a part value needs to crowd out
// the others ("two out of three"). It must also hold more than half the
// type's support.
const MinMajoritySupport = 2

// IsInitial reports whether a word is an initial: a lone cased letter ("J").
// A lone character in an uncased script (蒋, 王) is a whole word.
func IsInitial(word string) bool {
	r := []rune(word)
	return len(r) == 1 && unicode.ToUpper(r[0]) != unicode.ToLower(r[0])
}

// nameCandidate is one structured candidate, by type.
type nameCandidate struct {
	c      Candidate
	rank   int                 // position in rank order
	values map[string][]string // type → normalized parts, in idx order
	order  []string            // types in first-appearance order
	folded map[string]*partValue
}

// partValue is one distinct value of one part type.
type partValue struct {
	key     []string
	support int               // candidates carrying exactly this value
	rank    int               // best rank carrying it
	parts   []namevalues.Part // the best-ranked carrier's parts of this type
	into    *partValue        // the value it folds into; nil when maximal
	total   int               // support once folded values are added
	survive bool
}

// nameCluster is a cluster under construction.
type nameCluster struct {
	members []*nameCandidate
	sig     map[string]*partValue
}

// reconcileNames clusters name candidates already in rank order and returns
// the clusters in display order.
func reconcileNames(ranked []Candidate) []Cluster {
	var cands []*nameCandidate
	for _, c := range ranked {
		nc := structuredName(c, len(cands))
		if nc != nil {
			cands = append(cands, nc)
		}
	}
	if len(cands) == 0 {
		return nil
	}

	// Exact: one partValue per (type, value).
	byType := map[string][]*partValue{}
	var types []string
	for _, nc := range cands {
		nc.folded = map[string]*partValue{}
		for _, typ := range nc.order {
			key := nc.values[typ]
			var pv *partValue
			for _, v := range byType[typ] {
				if slices.Equal(v.key, key) {
					pv = v
					break
				}
			}
			if pv == nil {
				if byType[typ] == nil {
					types = append(types, typ)
				}
				pv = &partValue{key: key, rank: nc.rank, parts: partsOfType(nc.c.Value.Name, typ)}
				byType[typ] = append(byType[typ], pv)
			}
			pv.support++
			nc.folded[typ] = pv
		}
	}

	for _, typ := range types {
		values := byType[typ]
		foldValues(values)
		elect(values)
	}

	var survivors, eliminated []*nameCandidate
	for _, nc := range cands {
		ok := true
		for typ, pv := range nc.folded {
			pv = root(pv)
			nc.folded[typ] = pv
			ok = ok && pv.survive
		}
		if ok {
			survivors = append(survivors, nc)
		} else {
			eliminated = append(eliminated, nc)
		}
	}

	out := orderClusters(assemble(groupSurvivors(survivors)))
	return append(out, orderClusters(groupExact(eliminated))...)
}

// structuredName reads a candidate's parts by type; nil when it has none.
func structuredName(c Candidate, rank int) *nameCandidate {
	nc := &nameCandidate{c: c, rank: rank, values: map[string][]string{}}
	for _, p := range sortedParts(c.Value.Name) {
		v := NormalizeForm(p.Value)
		if v == "" {
			continue
		}
		typ := strings.TrimSpace(p.Type)
		if _, seen := nc.values[typ]; !seen {
			nc.order = append(nc.order, typ)
		}
		nc.values[typ] = append(nc.values[typ], v)
	}
	if len(nc.order) == 0 {
		return nil
	}
	return nc
}

func sortedParts(n *namevalues.Value) []namevalues.Part {
	parts := append([]namevalues.Part(nil), n.Parts...)
	sort.SliceStable(parts, func(i, j int) bool { return parts[i].Idx < parts[j].Idx })
	return parts
}

// partsOfType is a name's non-empty parts of one type, in idx order.
func partsOfType(n *namevalues.Value, typ string) []namevalues.Part {
	var out []namevalues.Part
	for _, p := range sortedParts(n) {
		if strings.TrimSpace(p.Type) == typ && NormalizeForm(p.Value) != "" {
			out = append(out, p)
		}
	}
	return out
}

// subsumes reports whether short maps, in order, onto a subsequence of long,
// each part equal to or an initial of its match, and the two differ.
func subsumes(long, short []string) bool {
	if slices.Equal(long, short) || len(short) > len(long) {
		return false
	}
	j := 0
	for _, s := range short {
		for j < len(long) && !partFits(s, long[j]) {
			j++
		}
		if j == len(long) {
			return false
		}
		j++
	}
	return true
}

// partFits: the same part, or an initial of it.
func partFits(short, long string) bool {
	if short == long {
		return true
	}
	if !IsInitial(short) || IsInitial(long) {
		return false
	}
	return []rune(short)[0] == []rune(long)[0]
}

// foldValues points each non-maximal value at the maximal value it folds
// into and totals support on the maximal ones.
func foldValues(values []*partValue) {
	var maximal []*partValue
	for _, v := range values {
		isMax := true
		for _, u := range values {
			if u != v && subsumes(u.key, v.key) {
				isMax = false
				break
			}
		}
		if isMax {
			maximal = append(maximal, v)
		}
	}
	for _, v := range values {
		v.total = 0
	}
	for _, v := range values {
		target := v
		for _, u := range maximal {
			if u == v || !subsumes(u.key, v.key) {
				continue
			}
			if target == v || u.support > target.support ||
				(u.support == target.support && u.rank < target.rank) {
				target = u
			}
		}
		if target != v {
			v.into = target
		}
		target.total += v.support
	}
}

// elect marks the surviving maximal values of one type.
func elect(values []*partValue) {
	var all int
	var best *partValue
	for _, v := range values {
		if v.into != nil {
			continue
		}
		all += v.total
		if best == nil || v.total > best.total {
			best = v
		}
	}
	winner := best.total >= MinMajoritySupport && 2*best.total > all
	for _, v := range values {
		if v.into == nil {
			v.survive = !winner || v == best
		}
	}
}

func root(v *partValue) *partValue {
	for v.into != nil {
		v = v.into
	}
	return v
}

// groupSurvivors groups surviving candidates that agree on every type they
// share. Fuller candidates place first; one missing a type joins the
// agreeing cluster with the most members, then the best ranked.
func groupSurvivors(cands []*nameCandidate) []*nameCluster {
	order := append([]*nameCandidate(nil), cands...)
	sort.SliceStable(order, func(i, j int) bool { return len(order[i].order) > len(order[j].order) })
	var clusters []*nameCluster
	for _, nc := range order {
		var into *nameCluster
		for _, cl := range clusters {
			if !agrees(cl.sig, nc.folded) {
				continue
			}
			if into == nil || len(cl.members) > len(into.members) ||
				(len(cl.members) == len(into.members) && cl.members[0].rank < into.members[0].rank) {
				into = cl
			}
		}
		if into == nil {
			into = &nameCluster{sig: map[string]*partValue{}}
			clusters = append(clusters, into)
		}
		into.members = append(into.members, nc)
		sort.SliceStable(into.members, func(i, j int) bool { return into.members[i].rank < into.members[j].rank })
		for typ, pv := range nc.folded {
			into.sig[typ] = pv
		}
	}
	return clusters
}

func agrees(sig, values map[string]*partValue) bool {
	for typ, pv := range values {
		if have, ok := sig[typ]; ok && have != pv {
			return false
		}
	}
	return true
}

// groupExact groups eliminated candidates by their exact parts; each
// cluster's value is its best-ranked member's name.
func groupExact(cands []*nameCandidate) []Cluster {
	var out []Cluster
	index := map[string]int{}
	for _, nc := range cands {
		k := exactSignature(nc.values)
		i, ok := index[k]
		if !ok {
			i = len(out)
			index[k] = i
			out = append(out, Cluster{Value: nc.c.Value})
		}
		out[i].ObservationIDs = append(out[i].ObservationIDs, nc.c.ObservationID)
		out[i].Support++
	}
	return out
}

func exactSignature(values map[string][]string) string {
	types := make([]string, 0, len(values))
	for typ := range values {
		types = append(types, typ)
	}
	sort.Strings(types)
	var b strings.Builder
	for _, typ := range types {
		b.WriteString(typ)
		b.WriteByte(0)
		b.WriteString(strings.Join(values[typ], "\x01"))
		b.WriteByte(0)
	}
	return b.String()
}

// assemble builds each cluster's name: per type, the fullest value's parts
// from its best-ranked carrier; types in the order of the member with the
// most types, others placed after the type that precedes them where they
// came from. A member whose parts are exactly the result is returned as is.
func assemble(clusters []*nameCluster) []Cluster {
	out := make([]Cluster, 0, len(clusters))
	for _, cl := range clusters {
		base := cl.members[0]
		for _, m := range cl.members {
			if len(m.order) > len(base.order) {
				base = m
			}
		}
		order := append([]string(nil), base.order...)
		for _, m := range cl.members {
			for i, typ := range m.order {
				if indexOf(order, typ) >= 0 {
					continue
				}
				at := 0
				for k := i - 1; k >= 0; k-- {
					if p := indexOf(order, m.order[k]); p >= 0 {
						at = p + 1
						break
					}
				}
				order = append(order[:at], append([]string{typ}, order[at:]...)...)
			}
		}

		values := map[string][]string{}
		name := &namevalues.Value{}
		var words []string
		for _, typ := range order {
			pv := cl.sig[typ]
			values[typ] = pv.key
			for _, p := range pv.parts {
				name.Parts = append(name.Parts, namevalues.Part{Idx: len(name.Parts), Value: p.Value, Type: p.Type})
				words = append(words, strings.TrimSpace(p.Value))
			}
		}
		name.Form = strings.Join(words, " ")

		c := Cluster{Value: Value{Name: name}}
		sig := exactSignature(values)
		for _, m := range cl.members {
			if exactSignature(m.values) == sig && slices.Equal(m.order, order) {
				c.Value = m.c.Value
				break
			}
		}
		for _, m := range cl.members {
			c.ObservationIDs = append(c.ObservationIDs, m.c.ObservationID)
		}
		c.Support = len(c.ObservationIDs)
		out = append(out, c)
	}
	return out
}

func indexOf(xs []string, x string) int {
	for i, y := range xs {
		if y == x {
			return i
		}
	}
	return -1
}

// orderClusters sorts by support descending, then lowest Observation id.
func orderClusters(cs []Cluster) []Cluster {
	sort.SliceStable(cs, func(i, j int) bool {
		if cs[i].Support != cs[j].Support {
			return cs[i].Support > cs[j].Support
		}
		return string(cs[i].ObservationIDs[0]) < string(cs[j].ObservationIDs[0])
	})
	return cs
}

// sameName reports whether two names carry the same parts by type, compared
// as the reconciler compares them; names with no parts match nothing.
func sameName(a, b *namevalues.Value) bool {
	na, nb := structuredName(Candidate{Value: Value{Name: a}}, 0), structuredName(Candidate{Value: Value{Name: b}}, 0)
	if na == nil || nb == nil {
		return false
	}
	return exactSignature(na.values) == exactSignature(nb.values)
}
