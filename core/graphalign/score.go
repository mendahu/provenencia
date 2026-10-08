package graphalign

import (
	"strconv"

	"github.com/mendahu/provenencia/core/match"
)

// Scored is the pairwise + edge result ScoreCandidate returns.
type Scored struct {
	Score float64
	Edge  float64
	Eval  match.Evaluation
}

// ScoreCandidate combines a pairwise Evaluation with edge support into one
// Fellegi–Sunter score. Align and any future caller share this path.
// provenance scales the total (≤0 treated as 1). edge is precomputed support
// from corresponding layer/canon bridges.
func ScoreCandidate(ev match.Evaluation, probe match.Values, edge, provenance float64, cfg Config, stats Stats) Scored {
	node := nodeScore(ev, probe, cfg, stats)
	// One exact agreement is at least a weak match. A small catalog's
	// frequency can push that log-odds under the accept bar. Lift the
	// log-odds to the bar and keep any PropertyAgreement weight on top, so
	// a toponym can clear medium while a name stays weak. A conflict is
	// not lifted.
	if unconflictedAgreement(ev) {
		bonus := agreementBonus(ev, cfg)
		if node-bonus < cfg.AcceptScore {
			node = cfg.AcceptScore + bonus
		}
	}
	if provenance <= 0 {
		provenance = 1
	}
	return Scored{
		Score: (node + edge) * provenance,
		Edge:  edge,
		Eval:  ev,
	}
}

// unconflictedAgreement is true when at least one Property agrees and none conflict.
func unconflictedAgreement(ev match.Evaluation) bool {
	agrees := 0
	for _, pc := range ev.Comparisons {
		switch pc.Outcome {
		case match.OutcomeAgree:
			agrees++
		case match.OutcomeConflict:
			return false
		}
	}
	return agrees > 0
}

func agreementBonus(ev match.Evaluation, cfg Config) float64 {
	var bonus float64
	for _, pc := range ev.Comparisons {
		if pc.Outcome == match.OutcomeAgree {
			bonus += cfg.agreementWeight(pc.Property)
		}
	}
	return bonus
}

// PropertyWeight is the log-odds nodeScore adds for one outcome.
func PropertyWeight(outcome match.Outcome, valueType string, property match.Property, probe match.Values, cfg Config, stats Stats) float64 {
	switch outcome {
	case match.OutcomeAgree:
		u := uFor(stats, property.Key, probe[property])
		return logOdds(cfg.mFor(valueType), cfg.uOr(u)) + cfg.agreementWeight(property)
	case match.OutcomeConflict:
		return -cfg.ConflictPenalty
	default:
		return 0
	}
}

func nodeScore(ev match.Evaluation, probe match.Values, cfg Config, stats Stats) float64 {
	var score float64
	for _, pc := range ev.Comparisons {
		score += PropertyWeight(pc.Outcome, pc.ValueType, pc.Property, probe, cfg, stats)
	}
	return score
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
