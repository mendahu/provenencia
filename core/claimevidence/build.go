package claimevidence

import (
	"bytes"
	"cmp"
	"maps"
	"slices"
	"strings"

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
// catalog's property order for each kind and each side's life. Sources is left
// for the loader to fill.
func Build(kind vocabkey.SubjectType, row, handle subjectcompare.Values, members []Member, neighbors []Neighbor,
	metas subjectcompare.Metas, rules Rules, segment segmentcompare.Rules, subject subjectcompare.Rules,
	value valuecompare.Rules, in Inputs) Evidence {
	out := Evidence{Kind: kind, Sections: Sections(kind, in.Order[kind], row, members, metas, rules, subject, value)}
	compared := subjectcompare.Compare(kind, row, handle, metas, subject, value)
	own := compared.Evidence
	out.Lifespan = lifespanOf(kind, compared, subject, in)
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

// lifespanOf is the kind's lifespan check: the first lead, in key order, that
// the kind derives a lifespan for. Nil for a kind with none, and when neither
// side has dates and neither contradicts itself.
func lifespanOf(kind vocabkey.SubjectType, compared subjectcompare.Result, subject subjectcompare.Rules, in Inputs) *Lifespan {
	lead, ok := subjectcompare.FirstLifespanLead(subject.Lifespans[kind])
	if !ok {
		return nil
	}
	if in.Row.First == nil && in.Record.First == nil &&
		in.Row.Missing != subjectcompare.LifespanContradicts && in.Record.Missing != subjectcompare.LifespanContradicts {
		return nil
	}
	out := &Lifespan{Row: in.Row, Record: in.Record}
	for _, p := range compared.Properties {
		if p.Key == lead {
			out.Outcome, out.PropertyCause, out.Relation, out.Cause = verdictOf(p.Result)
		}
	}
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
// key that a link between the entities corresponds to.
func relationships(n Neighbor, members []Member, own float64, metas subjectcompare.Metas, rules Rules,
	segment segmentcompare.Rules, subject subjectcompare.Rules, value valuecompare.Rules, in Inputs) []Relationship {
	var keys []vocabkey.BridgeKey
	observations := map[vocabkey.BridgeKey][][]byte{}
	for _, l := range n.Layer {
		if _, seen := observations[l.Key]; !seen {
			keys = append(keys, l.Key)
		}
		observations[l.Key] = append(observations[l.Key], l.Observations...)
	}
	defining := rules.Defining[n.Kind]
	var out []Relationship
	for _, key := range keys {
		var canons []CanonLink
		for _, c := range n.Canon {
			if segmentcompare.Corresponds(key, c.Key) {
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
		rel := Relationship{
			Key:      key,
			Neighbor: n.Subject,
			Kind:     n.Kind,
			Outcome:  res.Outcome,
			Cause:    res.Cause,
			Points:   res.Evidence,
			Checked:  res.Outcome == propertycompare.OutcomeAgree,
			Pinnable: true,
			Incoming: appendNew(slices.Clone(observations[key]), refsOf(n.Values, defining)...),
		}
		rel.Members, rel.Sections = bundles(n, canons, members, defining, metas, rules, subject, value, in)
		out = append(out, rel)
	}
	return out
}

// bundles are the member bridges of the corresponding links, each with the
// bridge's Observations and the defining values of the neighbor member it
// reaches, and the sections that compare the neighbor's other properties with
// only those neighbor members.
func bundles(n Neighbor, canons []CanonLink, members []Member, defining []vocabkey.PropertyKey,
	metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules, in Inputs) ([]Bundle, []Section) {
	type placed struct {
		bundle         Bundle
		member, nbhood int
	}
	var found []placed
	named := map[int]bool{}
	for _, c := range canons {
		for _, b := range c.Bridges {
			mi := slices.IndexFunc(members, func(m Member) bool { return bytes.Equal(m.Subject, b.Member) })
			ni := slices.IndexFunc(n.Members, func(m Member) bool { return bytes.Equal(m.Subject, b.Neighbor) })
			if mi < 0 || ni < 0 {
				continue
			}
			named[ni] = true
			found = append(found, placed{Bundle{
				Member: b.Member, Neighbor: b.Neighbor, Source: b.Source, Bridge: b.Subject,
				Observations: appendNew(slices.Clone(b.Observations), refsOf(n.Members[ni].Values, defining)...),
			}, mi, ni})
		}
	}
	slices.SortStableFunc(found, func(a, b placed) int {
		if c := cmp.Compare(a.member, b.member); c != 0 {
			return c
		}
		return cmp.Compare(a.nbhood, b.nbhood)
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
