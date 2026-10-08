package graphalign

import (
	"strconv"

	"github.com/mendahu/provenencia/core/match"
)

// Scored is the pairwise + edge result ScoreCandidate returns.
type Scored struct {
	Score   float64
	Reasons []string
	Edge    float64
	Eval    match.Evaluation
}

// ScoreCandidate combines a pairwise Evaluation with edge support into one
// Fellegi–Sunter score. Align and any future caller share this path.
// provenance scales the total (≤0 treated as 1). edge is precomputed support
// from corresponding layer/canon bridges.
func ScoreCandidate(ev match.Evaluation, probe match.Values, edge, provenance float64, cfg Config, stats Stats) Scored {
	node, reasons := nodeScore(ev, probe, cfg, stats)
	if edge > 0 {
		reasons = append(reasons, "edge support")
	}
	if provenance <= 0 {
		provenance = 1
	}
	return Scored{
		Score:   (node + edge) * provenance,
		Reasons: reasons,
		Edge:    edge,
		Eval:    ev,
	}
}

// PropertyWeight is the log-odds nodeScore adds for one outcome.
func PropertyWeight(outcome match.Outcome, valueType string, property match.Property, probe match.Values, cfg Config, stats Stats) float64 {
	switch outcome {
	case match.OutcomeAgree:
		u := uFor(stats, property.Key, probe[property])
		return logOdds(cfg.mFor(valueType), cfg.uOr(u))
	case match.OutcomeConflict:
		return -cfg.ConflictPenalty
	default:
		return 0
	}
}

func nodeScore(ev match.Evaluation, probe match.Values, cfg Config, stats Stats) (float64, []string) {
	var score float64
	var reasons []string
	for _, pc := range ev.Comparisons {
		w := PropertyWeight(pc.Outcome, pc.ValueType, pc.Property, probe, cfg, stats)
		switch pc.Outcome {
		case match.OutcomeAgree:
			score += w
			reasons = append(reasons, "agree "+pc.Property.Key)
		case match.OutcomeConflict:
			score += w
			reasons = append(reasons, "conflict "+pc.Property.Key)
		}
	}
	return score, reasons
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
