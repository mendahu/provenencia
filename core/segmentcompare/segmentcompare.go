// Package segmentcompare is layer 5 of matching: whether a segment, two
// primary Subjects joined by a bridge, lines up with a research segment. Pure:
// no I/O, no catalog, no numbers; the segment section, layer 3's rules and the
// bridge terms come in as arguments.
package segmentcompare

import (
	"slices"

	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Segment is one link of a primary Subject as the walk sees it: the bridge's
// key from this Subject's end, and the Subject at the other end.
type Segment struct {
	Key          vocabkey.BridgeKey
	NeighborKind vocabkey.SubjectType
	Neighbor     subjectcompare.Values
}

// Placement says where the layer neighbor stands against the research one.
type Placement string

const (
	// Together: the layer neighbor is placed on this research neighbor.
	Together Placement = "together"
	// Apart: the layer neighbor is placed on another handle.
	Apart Placement = "apart"
	// Unplaced: the layer neighbor is placed nowhere yet: entity compare
	// decides.
	Unplaced Placement = "unplaced"
)

// Cause is why a segment adds nothing or subtracts; empty when it supports.
type Cause string

const (
	CauseNotCorresponding Cause = "not_corresponding"
	CauseGated            Cause = "gated"
	CausePlacedApart      Cause = "placed_apart"
	CauseUnplacedDiffers  Cause = "unplaced_differs"
	CauseUnplacedUnsure   Cause = "unplaced_unsure"
)

// Gate is a per-key condition on a link: a minimum own evidence before
// support counts, and a factor of own evidence lost when placed apart.
type Gate struct{ MinOwnEvidence, ConflictFactor float64 }

// Rules is layer 5's section of the registry.
type Rules struct {
	// One lists the links where a differing neighbor conflicts, written with
	// an empty End: each holds from both ends.
	One []vocabkey.BridgeKey
	// OneSupport is earned by a one-cardinality link whose neighbor lines up.
	OneSupport float64
	// OneCredit is the share of the neighbor's own evidence a one-cardinality
	// link earns.
	OneCredit float64
	// OneConflict is lost when a one-cardinality link's neighbor is placed
	// elsewhere.
	OneConflict float64
	// OtherSupport is earned by any other link whose neighbor lines up.
	OtherSupport float64
	// OtherCredit is the share of the neighbor's own evidence any other link
	// earns.
	OtherCredit float64
	// Gates is, per key, a minimum own evidence before support counts and a
	// factor of own evidence lost when placed apart; a missing key is the
	// zero gate.
	Gates map[vocabkey.BridgeKey]Gate
	// UnplacedAlike is the least neighbor evidence for an unplaced neighbor
	// to count as lined up.
	UnplacedAlike float64
}

// Result is what Compare decides about one pair of segments.
type Result struct {
	Outcome  propertycompare.Outcome
	Cause    Cause
	Evidence float64
	Neighbor subjectcompare.Result // layer 3's verdict on the neighbors; zero when placed apart
}

// Compare says what one research segment adds to a pair, given the layer
// segment, where its neighbor is placed and the pair's own evidence.
// Segments whose keys do not correspond add nothing. Placed apart, a link
// declared one-cardinality loses OneConflict and a gated key loses its factor
// of own evidence; any other link is unknown. Placed together, the neighbors
// are compared by layer 3 and the link earns its support plus its credit of
// the neighbor's evidence, unless a gate wants more own evidence than the
// pair has. Unplaced, the neighbors are compared by layer 3 too: a neighbor
// with at least UnplacedAlike evidence counts as placed together, one with
// negative evidence as placed apart, and any other adds nothing.
func Compare(a, b Segment, placed Placement, own float64, metas subjectcompare.Metas, rules Rules, subject subjectcompare.Rules, value valuecompare.Rules) Result {
	return CompareNeighbor(a, b, placed, own, func() subjectcompare.Result {
		return subjectcompare.Compare(a.NeighborKind, a.Neighbor, b.Neighbor, metas, subject, value)
	}, rules)
}

// CompareNeighbor is Compare with the neighbors' comparison handed in as a
// function, called at most once and only when the link needs it: not for keys
// that do not correspond, nor for a neighbor placed apart. A caller that has
// compared the neighbors already returns that result.
func CompareNeighbor(a, b Segment, placed Placement, own float64, neighbor func() subjectcompare.Result, rules Rules) Result {
	if !Corresponds(a.Key, b.Key) {
		return Result{Outcome: propertycompare.OutcomeUnknown, Cause: CauseNotCorresponding}
	}
	g := rules.Gates[a.Key]
	both := a.Key
	both.End = ""
	one := slices.Contains(rules.One, both)
	if placed == Apart {
		out := apart(one, g, own, rules)
		out.Cause = CausePlacedApart
		return out
	}
	nr := neighbor()
	if placed == Unplaced {
		switch {
		case nr.Evidence >= rules.UnplacedAlike:
		case nr.Evidence < 0:
			out := apart(one, g, own, rules)
			out.Cause, out.Neighbor = CauseUnplacedDiffers, nr
			return out
		default:
			return Result{Outcome: propertycompare.OutcomeUnknown, Cause: CauseUnplacedUnsure, Neighbor: nr}
		}
	}
	out := Result{Neighbor: nr}
	if g.MinOwnEvidence > 0 && own < g.MinOwnEvidence {
		out.Outcome, out.Cause = propertycompare.OutcomeUnknown, CauseGated
		return out
	}
	support, credit := rules.OtherSupport, rules.OtherCredit
	if one {
		support, credit = rules.OneSupport, rules.OneCredit
	}
	out.Evidence = support + credit*out.Neighbor.Evidence
	out.Outcome = propertycompare.OutcomeAgree
	if out.Neighbor.Conflicts() {
		out.Outcome = propertycompare.OutcomePartial
	}
	return out
}

// apart scores a link whose neighbor is somewhere else: a one-cardinality link
// loses OneConflict and a gated key loses its factor of own evidence. The
// caller sets Cause and Neighbor.
func apart(one bool, g Gate, own float64, rules Rules) Result {
	out := Result{Outcome: propertycompare.OutcomeUnknown}
	if one {
		out.Evidence -= rules.OneConflict
	}
	out.Evidence -= g.ConflictFactor * own
	if one || g.ConflictFactor > 0 {
		out.Outcome = propertycompare.OutcomeConflict
	}
	return out
}
