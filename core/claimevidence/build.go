package claimevidence

import (
	"bytes"
	"cmp"
	"maps"
	"slices"
	"strings"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Build decides one row's claim evidence: the property lines of the row's
// values against the entity's accepted members, and, for each layer neighbor
// the page placed on an entity, a relationship for every link of the row's
// Subject that corresponds, through segment compare, to a link between the
// entity's members and the neighbor entity's. handle is the entity's profile,
// whose evidence against the row (own) gates the segments. in carries the
// catalog's property order for each kind. Sources is left
// for the loader to fill.
func Build(kind vocabkey.SubjectType, row, handle subjectcompare.Values, members []Member, neighbors []Neighbor,
	metas subjectcompare.Metas, rules Rules, segment segmentcompare.Rules, subject subjectcompare.Rules,
	value valuecompare.Rules, in Inputs) Evidence {
	out := Evidence{Kind: kind, Sections: Sections(kind, in.Order[kind], row, members, metas, rules, subject, value)}
	compared := subjectcompare.Compare(kind, row, handle, metas, subject, value)
	own := compared.Evidence
	for _, n := range neighbors {
		out.Relationships = append(out.Relationships, relationships(n, members, own, metas, rules, segment, subject, value, in)...)
	}
	slices.SortStableFunc(out.Relationships, func(a, b Relationship) int {
		if c := cmp.Compare(boolRank(!isOne(segment, a.Key)), boolRank(!isOne(segment, b.Key))); c != 0 {
			return c
		}
		if c := strings.Compare(a.Key.String(), b.Key.String()); c != 0 {
			return c
		}
		return bytes.Compare(a.Neighbor, b.Neighbor)
	})
	return out
}

func boolRank(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isOne reports whether a key, read from either end, is one of segment's
// one-cardinality keys.
func isOne(segment segmentcompare.Rules, k vocabkey.BridgeKey) bool {
	k.End = ""
	return slices.Contains(segment.One, k)
}

// relationships are one neighbor's relationships: one for each distinct layer
// key that a link between the entities corresponds to, unless both sides
// record that fact through the same pattern with paired pivots (counted once).
func relationships(n Neighbor, members []Member, own float64, metas subjectcompare.Metas, rules Rules,
	segment segmentcompare.Rules, subject subjectcompare.Rules, value valuecompare.Rules, in Inputs) []Relationship {
	type layerSide struct {
		observations [][]byte
		recording    kinship.Recording
		pivots       []Pivot
	}
	var keys []vocabkey.BridgeKey
	layer := map[vocabkey.BridgeKey]*layerSide{}
	for _, l := range n.Layer {
		side, seen := layer[l.Key]
		if !seen {
			keys = append(keys, l.Key)
			side = &layerSide{recording: l.Recording.Clone()}
			layer[l.Key] = side
		} else {
			side.recording.Merge(l.Recording)
		}
		side.observations = append(side.observations, l.Observations...)
		for _, p := range l.Pivots {
			if !slices.ContainsFunc(side.pivots, func(x Pivot) bool { return bytes.Equal(x.Subject, p.Subject) }) {
				side.pivots = append(side.pivots, p)
			}
		}
	}
	paired := func(p []byte) ([]byte, bool) {
		h, ok := in.Placed[string(p)]
		return h, ok
	}
	defining := rules.Defining[n.Kind]
	var out []Relationship
	for _, key := range keys {
		side := layer[key]
		var canons []CanonLink
		for _, c := range n.Canon {
			if segmentcompare.Corresponds(key, c.Key) && !kinship.CountedOnce(side.recording, c.Recording, paired) {
				canons = append(canons, c)
			}
		}
		if len(canons) == 0 {
			continue
		}
		res := segmentcompare.Compare(
			segmentcompare.Segment{Key: key, NeighborKind: n.Kind, Neighbor: n.Values},
			segmentcompare.Segment{Key: canons[0].Key, NeighborKind: n.Kind, Neighbor: n.Handle},
			segmentcompare.Together, own, metas, segment, subject, value)
		if res.Outcome != propertycompare.OutcomeAgree && res.Outcome != propertycompare.OutcomePartial {
			continue
		}
		incoming := appendNew(slices.Clone(side.observations), refsOf(n.Values, defining)...)
		for _, p := range side.pivots {
			incoming = appendNew(incoming, refsOf(p.Values, rules.Defining[p.Kind])...)
		}
		canonRec := canons[0].Recording.Clone()
		for _, c := range canons[1:] {
			canonRec.Merge(c.Recording)
		}
		rel := Relationship{
			Key:      key,
			Neighbor: n.Subject,
			Kind:     n.Kind,
			Outcome:  res.Outcome,
			Cause:    res.Cause,
			Points:   res.Evidence,
			Checked:  res.Outcome == propertycompare.OutcomeAgree,
			Incoming: incoming,
			Layer:    side.recording.Shown(),
			Canon:    canonRec.Shown(),
		}
		rel.Members, rel.Sections = bundles(n, canons, members, defining, metas, rules, subject, value, in)
		out = append(out, rel)
	}
	return out
}

// bundles are the member bridges of the corresponding links, each with the
// bridge's Observations and the defining values of the neighbor member it
// reaches, and the sections that compare the neighbor's other properties with
// only those neighbor members. Bridges behind an implied link make one bundle
// per pivot (one Source's recording of the fact): the bridges that reach the
// pivot, the member's first, and the pivot's defining values. A pivot with no
// bridge of a row member makes none.
func bundles(n Neighbor, canons []CanonLink, members []Member, defining []vocabkey.PropertyKey,
	metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules, in Inputs) ([]Bundle, []Section) {
	// entry is a recorded bridge, or the bridges of one pivot, at the position
	// of its first bridge.
	type entry struct {
		pivot   []byte
		bridges []Bridge
	}
	var walk []entry
	pivots := map[string]Pivot{}
	for _, c := range canons {
		for _, p := range c.Pivots {
			pivots[string(p.Subject)] = p
		}
		for _, b := range c.Bridges {
			if b.Pivot == nil {
				walk = append(walk, entry{bridges: []Bridge{b}})
				continue
			}
			if i := slices.IndexFunc(walk, func(e entry) bool { return e.pivot != nil && bytes.Equal(e.pivot, b.Pivot) }); i >= 0 {
				walk[i].bridges = append(walk[i].bridges, b)
				continue
			}
			walk = append(walk, entry{pivot: b.Pivot, bridges: []Bridge{b}})
		}
	}
	memberAt := func(id []byte) int {
		return slices.IndexFunc(members, func(m Member) bool { return bytes.Equal(m.Subject, id) })
	}
	neighborAt := func(id []byte) int {
		if len(id) == 0 {
			return -1
		}
		return slices.IndexFunc(n.Members, func(m Member) bool { return bytes.Equal(m.Subject, id) })
	}
	type placed struct {
		bundle         Bundle
		member, nbhood int
	}
	var found []placed
	named := map[int]bool{}
	for _, e := range walk {
		if e.pivot == nil {
			b := e.bridges[0]
			mi, ni := memberAt(b.Member), neighborAt(b.Neighbor)
			if mi < 0 || ni < 0 {
				continue
			}
			named[ni] = true
			found = append(found, placed{Bundle{
				Member: b.Member, Neighbor: b.Neighbor, Source: b.Source, Bridges: [][]byte{b.Subject},
				Observations: appendNew(slices.Clone(b.Observations), refsOf(n.Members[ni].Values, defining)...),
			}, mi, ni})
			continue
		}
		first := slices.IndexFunc(e.bridges, func(b Bridge) bool { return memberAt(b.Member) >= 0 })
		if first < 0 {
			continue
		}
		lead := e.bridges[first]
		ordered := []Bridge{lead}
		ni := -1
		for i, b := range e.bridges {
			if i != first {
				ordered = append(ordered, b)
			}
			if j := neighborAt(b.Neighbor); ni < 0 && j >= 0 {
				ni = j
			}
		}
		bundle := Bundle{Member: lead.Member, Source: lead.Source}
		var obs [][]byte
		for _, b := range ordered {
			bundle.Bridges = append(bundle.Bridges, b.Subject)
			obs = appendNew(obs, b.Observations...)
		}
		if ni >= 0 {
			named[ni] = true
			bundle.Neighbor = n.Members[ni].Subject
			obs = appendNew(obs, refsOf(n.Members[ni].Values, defining)...)
		}
		if p, ok := pivots[string(e.pivot)]; ok {
			obs = appendNew(obs, refsOf(p.Values, rules.Defining[p.Kind])...)
		}
		bundle.Observations = obs
		found = append(found, placed{bundle, memberAt(lead.Member), ni})
	}
	slices.SortStableFunc(found, func(a, b placed) int {
		if c := cmp.Compare(a.member, b.member); c != 0 {
			return c
		}
		ai, bi := a.nbhood, b.nbhood
		if ai < 0 {
			ai = len(n.Members)
		}
		if bi < 0 {
			bi = len(n.Members)
		}
		return cmp.Compare(ai, bi)
	})
	var out []Bundle
	for _, f := range found {
		out = append(out, f.bundle)
	}
	var theirs []Member
	for i, m := range n.Members {
		if named[i] {
			theirs = append(theirs, m)
		}
	}
	other := maps.Clone(n.Values)
	for _, p := range defining {
		delete(other, p)
	}
	return out, Sections(n.Kind, in.Order[n.Kind], other, theirs, metas, rules, subject, value)
}

// refsOf is the Ref of every value of the listed properties, in property order.
func refsOf(v subjectcompare.Values, props []vocabkey.PropertyKey) [][]byte {
	var out [][]byte
	for _, p := range props {
		for _, x := range v[p] {
			if len(x.Ref) > 0 {
				out = append(out, x.Ref)
			}
		}
	}
	return out
}

// appendNew appends each id not already in dst, in order.
func appendNew(dst [][]byte, ids ...[]byte) [][]byte {
	for _, id := range ids {
		if !slices.ContainsFunc(dst, func(d []byte) bool { return bytes.Equal(d, id) }) {
			dst = append(dst, id)
		}
	}
	return dst
}
