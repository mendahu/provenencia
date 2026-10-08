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
	exact, partial := nodeScore(ev, probe, cfg, stats)
	// One exact agreement is at least a weak match. A small catalog's
	// frequency can push that log-odds under the accept bar. Lift the
	// log-odds to the bar and keep any PropertyAgreement weight on top, so
	// a toponym can clear medium while a name stays weak. A conflict is
	// not lifted. A partial text resemblance is already a fraction of that
	// floored weight, so it is not lifted again.
	if unconflictedAgreement(ev) {
		bonus := agreementBonus(ev, cfg)
		if exact-bonus < cfg.AcceptScore {
			exact = cfg.AcceptScore + bonus
		}
	}
	if provenance <= 0 {
		provenance = 1
	}
	return Scored{
		Score: (exact + partial + edge) * provenance,
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

// PropertyWeight is the log-odds one outcome adds. similarity scales a
// partial text resemblance; exact agreements ignore it.
func PropertyWeight(outcome match.Outcome, valueType string, property match.Property, similarity float64, probe match.Values, cfg Config, stats Stats) float64 {
	switch outcome {
	case match.OutcomeAgree:
		u := uFor(stats, property.Key, probe[property])
		return logOdds(cfg.mFor(valueType), cfg.uOr(u)) + cfg.agreementWeight(property)
	case match.OutcomePartial:
		if similarity <= 0 {
			return 0
		}
		if similarity > 1 {
			similarity = 1
		}
		return similarity * flooredAgreementWeight(valueType, property, probe, cfg, stats)
	case match.OutcomeConflict:
		return -cfg.ConflictPenalty
	default:
		return 0
	}
}

// flooredAgreementWeight is what one exact agreement on this Property would
// contribute on its own: log-odds, lifted to the accept bar when a small
// catalog makes it look cheap, plus any registry bonus.
func flooredAgreementWeight(valueType string, property match.Property, probe match.Values, cfg Config, stats Stats) float64 {
	raw := logOdds(cfg.mFor(valueType), cfg.uOr(uFor(stats, property.Key, probe[property])))
	bonus := cfg.agreementWeight(property)
	if raw < cfg.AcceptScore {
		return cfg.AcceptScore + bonus
	}
	return raw + bonus
}

// nodeScore splits exact outcomes from partial text resemblances. The accept
// floor applies only to the exact side.
func nodeScore(ev match.Evaluation, probe match.Values, cfg Config, stats Stats) (exact, partial float64) {
	for _, pc := range ev.Comparisons {
		w := PropertyWeight(pc.Outcome, pc.ValueType, pc.Property, pc.Similarity, probe, cfg, stats)
		if pc.Outcome == match.OutcomePartial {
			partial += w
			continue
		}
		exact += w
	}
	return exact, partial
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
