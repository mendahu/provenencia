package match

import (
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// Pairwise evaluation reports how close two value sets are. Each property's
// similarity comes from Compare, the same value compare Rank uses. The promote
// scale turns that similarity into points. Reconciliation stays on
// autoreconcile.Compatible, so two spellings remain two values on a handle.

// Outcome is how one Property compares on two sides.
type Outcome string

const (
	OutcomeAgree    Outcome = "agree"   // similarity 1
	OutcomePartial  Outcome = "partial" // resemblance above 0, below 1
	OutcomeConflict Outcome = "conflict"
	OutcomeUnknown  Outcome = "unknown"
)

// PropertyMeta names a Property and how pairwise evaluation treats it.
type PropertyMeta struct {
	Property    Property
	ValueType   string // propertymeta.ValueType*
	Cardinality string // propertymeta.Cardinality*; empty means single
}

// PropertyComparison is one Property's pairwise result.
type PropertyComparison struct {
	Property    Property
	Outcome     Outcome
	ValueType   string
	Cardinality string
	// Similarity is the value-compare result times both values' support weights,
	// 0…1, when Comparable; a minority value that matches exactly is below 1.
	// Comparable is false when either side has nothing value compare can judge.
	Similarity float64
	Comparable bool
	// Probe and Candidate are the two values behind the result.
	Probe, Candidate ValueRef
}

// Evaluation is the pairwise result for every Property listed in metas.
type Evaluation struct {
	Comparisons []PropertyComparison
}

// Evaluate reports a value-compare similarity for each property in metas,
// under rules. The outcome is OutcomeOf the similarity. Promote applies its
// own scale on top of Similarity and Comparable.
func Evaluate(probe, candidate Values, metas []PropertyMeta, rules valuecompare.Rules) Evaluation {
	var out Evaluation
	for _, m := range metas {
		if m.ValueType == "" {
			continue
		}
		card := m.Cardinality
		if card == "" {
			card = propertymeta.CardinalitySingle
		}
		a, b := poolPeriods(m.Property, probe, candidate)
		pr := pairValues(rules, m.ValueType, card, a, b)
		out.Comparisons = append(out.Comparisons, PropertyComparison{
			Property:    m.Property,
			ValueType:   m.ValueType,
			Cardinality: card,
			Similarity:  pr.Similarity,
			Comparable:  pr.Comparable,
			Probe:       pr.Probe,
			Candidate:   pr.Candidate,
			Outcome:     OutcomeOf(pr.Similarity, pr.Comparable, card),
		})
	}
	return out
}

// OutcomeOf is the outcome of a pair: not comparable is unknown; a similarity
// above 0 agrees at 1 or more and is partial below; otherwise a multiple
// Property is unknown and any other conflicts.
func OutcomeOf(similarity float64, comparable bool, cardinality string) Outcome {
	if !comparable {
		return OutcomeUnknown
	}
	if similarity > 0 {
		if similarity >= 1 {
			return OutcomeAgree
		}
		return OutcomePartial
	}
	if cardinality == propertymeta.CardinalityMultiple {
		return OutcomeUnknown
	}
	return OutcomeConflict
}
