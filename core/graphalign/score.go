package graphalign

import "github.com/mendahu/provenencia/core/match"

// Scored is the pairwise + edge result ScoreCandidate returns.
type Scored struct {
	Score float64
	Edge  float64
	Eval  match.Evaluation
}

// ScoreCandidate combines a pairwise Evaluation with edge support into one
// score. Align and any future caller share this path. provenance scales the
// total (≤0 treated as 1). edge is EdgePoints over the corresponding links.
func ScoreCandidate(ev match.Evaluation, edge, provenance float64, cfg Config) Scored {
	total := propertyPoints(&ev, cfg)
	if provenance <= 0 {
		provenance = 1
	}
	return Scored{
		Score: (total + edge) * provenance,
		Edge:  edge,
		Eval:  ev,
	}
}

// propertyPoints is the pairwise total. It writes each comparison's outcome.
// Callers that credit a neighbor use this and leave that neighbor's links out.
func propertyPoints(ev *match.Evaluation, cfg Config) float64 {
	var total float64
	for i := range ev.Comparisons {
		pc := &ev.Comparisons[i]
		outcome, points := ScoreSimilarity(pc.Similarity, pc.Comparable, pc.Cardinality, pc.Property, pc.ValueType, cfg)
		pc.Outcome = outcome
		total += points
	}
	return total
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

// ScoreSimilarity looks up the scale for a property and applies it. The
// scale is registry data; this function does not name a property or a value
// type beyond using them as lookup keys.
func ScoreSimilarity(similarity float64, comparable bool, cardinality string, property match.Property, valueType string, cfg Config) (match.Outcome, float64) {
	return ApplyScale(similarity, comparable, cardinality, cfg.resolvedScale(property, valueType))
}

// ApplyScale turns one similarity into an outcome and points. The outcome is
// match.OutcomeOf; an agreement or partial earns Weight × similarity, a
// conflict costs Contradiction, unknown nothing. It has no property names,
// value types, or catalog lookups.
func ApplyScale(similarity float64, comparable bool, cardinality string, scale Scale) (match.Outcome, float64) {
	similarity = min(similarity, 1)
	outcome := match.OutcomeOf(similarity, comparable, cardinality)
	switch outcome {
	case match.OutcomeAgree, match.OutcomePartial:
		return outcome, scale.Weight * similarity
	case match.OutcomeConflict:
		return outcome, -scale.Contradiction
	}
	return outcome, 0
}

// resolvedScale is the property's scale, else its value type's, else the zero scale.
func (c Config) resolvedScale(property match.Property, valueType string) Scale {
	scale, ok := c.PropertyScale[property]
	if !ok && c.ValueTypeScale != nil {
		scale = c.ValueTypeScale[valueType]
	}
	return scale
}
