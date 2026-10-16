package graphalign

import (
	"strconv"

	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/match"
)

// Scored is the pairwise + edge result ScoreCandidate returns.
type Scored struct {
	Score float64
	Edge  float64
	Eval  match.Evaluation
}

// ScoreCandidate combines a pairwise Evaluation with edge support into one
// score. Align and any future caller share this path. provenance scales the
// total (≤0 treated as 1). edge is EdgePoints over the corresponding links.
func ScoreCandidate(ev match.Evaluation, probe match.Values, edge, provenance float64, cfg Config, stats Stats) Scored {
	total := propertyPoints(&ev, probe, cfg, stats)
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
func propertyPoints(ev *match.Evaluation, probe match.Values, cfg Config, stats Stats) float64 {
	var total float64
	for i := range ev.Comparisons {
		pc := &ev.Comparisons[i]
		outcome, points := ScoreSimilarity(pc.Similarity, pc.Discount, pc.Comparable, pc.Cardinality, pc.Property, pc.ValueType, probe[pc.Property], cfg, stats)
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
// type beyond using them as lookup keys. Positive points are multiplied by
// 1 − discount, the share the matched values' frequency in the project costs.
// The outcome never reads the discount, and a conflict is not scaled.
func ScoreSimilarity(similarity, discount float64, comparable bool, cardinality string, property match.Property, valueType string, vals []match.Value, cfg Config, stats Stats) (match.Outcome, float64) {
	outcome, points := ApplyScale(similarity, comparable, cardinality, cfg.resolvedScale(property, valueType, vals, stats))
	if points > 0 {
		points *= 1 - discount
	}
	return outcome, points
}

// ApplyScale turns one similarity into an outcome and points. It has no
// property names, value types, or catalog lookups.
func ApplyScale(similarity float64, comparable bool, cardinality string, scale Scale) (match.Outcome, float64) {
	if !comparable {
		return match.OutcomeUnknown, 0
	}
	if similarity > 1 {
		similarity = 1
	}
	if similarity > 0 && similarity >= scale.Floor {
		points := scale.Weight * similarity
		if similarity >= 1 {
			return match.OutcomeAgree, points
		}
		return match.OutcomePartial, points
	}
	if cardinality == properties.CardinalityMultiple {
		return match.OutcomeUnknown, 0
	}
	return match.OutcomeConflict, -scale.Contradiction
}

// resolvedScale is the property's scale, else its value type's, with a
// frequency weight filled in when the scale asks for one.
func (c Config) resolvedScale(property match.Property, valueType string, vals []match.Value, stats Stats) Scale {
	scale, ok := c.PropertyScale[property]
	if !ok && c.ValueTypeScale != nil {
		scale = c.ValueTypeScale[valueType]
	}
	if scale.Frequency {
		if scale.Contradiction == 0 {
			scale.Contradiction = c.ConflictPenalty
		}
		scale.Weight = flooredAgreementWeight(valueType, property, vals, c, stats)
	}
	return scale
}

// flooredAgreementWeight is what one exact agreement on this Property would
// contribute on its own: log-odds, lifted to the accept bar when a small
// catalog makes it look cheap, plus any registry bonus.
func flooredAgreementWeight(valueType string, property match.Property, vals []match.Value, cfg Config, stats Stats) float64 {
	raw := logOdds(cfg.mFor(valueType), cfg.uOr(uFor(stats, property.Key, vals)))
	bonus := cfg.agreementWeight(property)
	if raw < cfg.AcceptScore {
		return cfg.AcceptScore + bonus
	}
	return raw + bonus
}

func uFor(stats Stats, propertyKey string, vals []match.Value) float64 {
	if stats.ValueFreq == nil {
		return 0
	}
	byVal := stats.ValueFreq[propertyKey]
	if byVal == nil || len(vals) == 0 {
		return 0
	}
	for _, v := range vals {
		if v.Negative {
			continue
		}
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
