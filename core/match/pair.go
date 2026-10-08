package match

import (
	"github.com/mendahu/provenencia/core/database/properties"
)

// Pairwise evaluation reports how close two value sets are. Each property's
// similarity comes from ComparerFor, the same comparer Rank uses. The promote
// scale turns that similarity into points. Reconciliation stays on
// autoreconcile.Compatible, so two spellings remain two values on a handle.

// Outcome is how one Property compares on two sides.
type Outcome string

const (
	OutcomeAgree    Outcome = "agree"   // similarity 1
	OutcomePartial  Outcome = "partial" // resemblance above the scale floor
	OutcomeConflict Outcome = "conflict"
	OutcomeUnknown  Outcome = "unknown"
)

// PropertyMeta names a Property and how pairwise evaluation treats it.
type PropertyMeta struct {
	Property    Property
	ValueType   string // properties.ValueType*
	Cardinality string // properties.Cardinality*; empty means single
}

// PropertyComparison is one Property's pairwise result.
type PropertyComparison struct {
	Property    Property
	Outcome     Outcome
	ValueType   string
	Cardinality string
	// Similarity is the comparer result, 0…1, when Comparable.
	// Comparable is false when either side has nothing the comparer can judge.
	Similarity float64
	Comparable bool
}

// Evaluation is the pairwise result for every Property listed in metas.
type Evaluation struct {
	Comparisons []PropertyComparison
}

// Evaluate reports a comparer similarity for each property in metas.
// The outcome here uses a floor of zero: similarity 1 agrees, a positive
// resemblance is partial, and a comparable zero conflicts (or is unknown
// when the property holds several values). Promote applies its own scale
// on top of Similarity and Comparable, which can raise that floor.
func Evaluate(probe, candidate Values, metas []PropertyMeta) Evaluation {
	var out Evaluation
	for _, m := range metas {
		if m.ValueType == "" {
			continue
		}
		card := m.Cardinality
		if card == "" {
			card = properties.CardinalitySingle
		}
		sim, ok := compareProperty(m.ValueType, probe[m.Property], candidate[m.Property])
		pc := PropertyComparison{
			Property:    m.Property,
			ValueType:   m.ValueType,
			Cardinality: card,
			Similarity:  sim,
			Comparable:  ok,
			Outcome:     outcomeAtFloor(sim, ok, card, 0),
		}
		out.Comparisons = append(out.Comparisons, pc)
	}
	return out
}

func compareProperty(valueType string, a, b []Value) (float64, bool) {
	cmp := ComparerFor(valueType)
	if cmp == nil {
		return 0, false
	}
	return best(cmp, a, b)
}

// outcomeAtFloor is the outcome a scale with this floor would give.
// Promote's ApplyScale is the one that scores; this mirrors a floor of zero
// so a caller without a scale still sees agree, partial, or conflict.
func outcomeAtFloor(similarity float64, comparable bool, cardinality string, floor float64) Outcome {
	if !comparable {
		return OutcomeUnknown
	}
	if similarity > 0 && similarity >= floor {
		if similarity >= 1 {
			return OutcomeAgree
		}
		return OutcomePartial
	}
	if cardinality == properties.CardinalityMultiple {
		return OutcomeUnknown
	}
	return OutcomeConflict
}
