package match

import (
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/properties"
)

// Pairwise evaluation (S9-41) judges two value sets with the same sameness
// notion as reconciliation (`autoreconcile.Compatible`). Graph alignment
// calls Evaluate for structure-backed candidates. Rank remains the
// property-only bulk path (fuzzy Feature scores); unifying Rank onto
// Compatible is a later step so Promote never grows a second sameness story.

// Outcome is how one Property compares on two sides.
type Outcome string

const (
	OutcomeAgree    Outcome = "agree"
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
}

// Evaluation is the pairwise result for every Property listed in metas.
type Evaluation struct {
	Comparisons []PropertyComparison
}

// Evaluate compares probe and candidate property values using Compatible.
// metas supplies value type and cardinality for each Property.
//
// Agree: at least one Compatible pair. Conflict (single cardinality): both
// sides carry values and none are Compatible. Multiple cardinality never
// conflicts on a difference — non-agreeing populated sides are unknown.
// Unknown: one side has no carrying values for the Property.
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
		out.Comparisons = append(out.Comparisons, PropertyComparison{
			Property:    m.Property,
			ValueType:   m.ValueType,
			Cardinality: card,
			Outcome:     compareProperty(m.ValueType, card, probe[m.Property], candidate[m.Property]),
		})
	}
	return out
}

func compareProperty(valueType, cardinality string, a, b []Value) Outcome {
	aOK := hasCarrying(valueType, a)
	bOK := hasCarrying(valueType, b)
	if !aOK || !bOK {
		return OutcomeUnknown
	}
	if anyCompatible(valueType, a, b) {
		return OutcomeAgree
	}
	if cardinality == properties.CardinalityMultiple {
		return OutcomeUnknown
	}
	return OutcomeConflict
}

func hasCarrying(valueType string, vs []Value) bool {
	for _, v := range vs {
		av := toAuto(v)
		// Compatible with itself is true only when the value carries evidence.
		if autoreconcile.Compatible(valueType, av, av) {
			return true
		}
	}
	return false
}

func anyCompatible(valueType string, a, b []Value) bool {
	for _, x := range a {
		ax := toAuto(x)
		if !autoreconcile.Compatible(valueType, ax, ax) {
			continue
		}
		for _, y := range b {
			ay := toAuto(y)
			if !autoreconcile.Compatible(valueType, ay, ay) {
				continue
			}
			if autoreconcile.Compatible(valueType, ax, ay) {
				return true
			}
		}
	}
	return false
}

func toAuto(v Value) autoreconcile.Value {
	// Compatible's term module keys on TermID and refuses empty IDs. Pure
	// callers often only have the term key; use it as the id when unset.
	termID := append([]byte(nil), v.TermID...)
	if len(termID) == 0 && v.Term != "" {
		termID = []byte(v.Term)
	}
	return autoreconcile.Value{
		Text:       v.Text,
		HasText:    v.HasText,
		Integer:    v.Integer,
		HasInteger: v.HasInteger,
		TermKey:    v.Term,
		TermID:     termID,
		Date:       v.Date,
		Name:       v.Name,
	}
}
