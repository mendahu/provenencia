package graphalign

import "github.com/mendahu/provenencia/core/subjectcompare"

// Scored is the pairwise + edge result ScoreCandidate returns.
type Scored struct {
	Score  float64
	Edge   float64
	Result subjectcompare.Result
}

// ScoreCandidate combines a pair's entity compare result with edge support
// into one score. Align and any future caller share this path. provenance
// scales the total (≤0 treated as 1). edge is EdgePoints over the
// corresponding links.
func ScoreCandidate(r subjectcompare.Result, edge, provenance float64) Scored {
	if provenance <= 0 {
		provenance = 1
	}
	return Scored{
		Score:  (r.Evidence + edge) * provenance,
		Edge:   edge,
		Result: r,
	}
}

// EdgeFact is one mapped neighbor under a signature. A match carries the
// fan-out and the neighbor's property total. A mismatch carries the
// signature's conflict factor. MinNode, ConflictFactor, and Mismatch are
// numbers copied from the signature scale; this type does not name a signature.
type EdgeFact struct {
	FanOut         float64
	NeighborNode   float64
	MinNode        float64
	ConflictFactor float64
	Mismatch       bool
}

// EdgePoints sums support and credit, gated by the fact's scale. nodeScore
// is the candidate's own property total. A mismatch subtracts
// ConflictFactor × nodeScore. A match whose MinNode is above nodeScore adds
// nothing. Otherwise Support(fan-out) + Credit(fan-out) × NeighborNode.
// Fan-out at or below FanOutLowMax uses the low pair. A negative neighbor
// total subtracts.
func EdgePoints(facts []EdgeFact, nodeScore float64, cfg Config) float64 {
	var total float64
	for _, f := range facts {
		if f.Mismatch {
			total -= f.ConflictFactor * nodeScore
			continue
		}
		if f.MinNode > 0 && nodeScore < f.MinNode {
			continue
		}
		if f.FanOut <= cfg.FanOutLowMax {
			total += cfg.EdgeSupportLow + cfg.NeighborCreditLow*f.NeighborNode
		} else {
			total += cfg.EdgeSupportHigh + cfg.NeighborCreditHigh*f.NeighborNode
		}
	}
	return total
}
