package match

import (
	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/database/properties"
)

// Pairwise evaluation (S9-41) judges two value sets with the same sameness
// notion as reconciliation (`autoreconcile.Compatible`). Free text that is
// not the same value can still be a partial resemblance (TextResemblance):
// a spelling variant or shared words. That outcome is for scoring only.
// Reconciliation stays exact, so two spellings remain two values on a handle.
// Rank remains the property-only bulk path (fuzzy Feature scores).

// Outcome is how one Property compares on two sides.
type Outcome string

const (
	OutcomeAgree    Outcome = "agree"
	OutcomePartial  Outcome = "partial" // text resembles, but is not the same value
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
	// Similarity is set for OutcomePartial: how close the text is, in 0…1.
	// An exact agreement leaves it 0; that weight is the full agreement.
	Similarity float64
}

// Evaluation is the pairwise result for every Property listed in metas.
type Evaluation struct {
	Comparisons []PropertyComparison
}

// Evaluate compares probe and candidate property values using Compatible.
// metas supplies value type and cardinality for each Property.
//
// Agree: at least one Compatible pair. Partial: free text that is not the
// same value but resembles (a spelling variant, or shared words). Conflict
// (single cardinality): both sides carry values and none agree or resemble.
// Multiple cardinality never conflicts on a difference — non-agreeing
// populated sides are unknown. Unknown: one side has no carrying values.
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
		outcome, sim := compareProperty(m.ValueType, card, probe[m.Property], candidate[m.Property])
		pc := PropertyComparison{
			Property:    m.Property,
			ValueType:   m.ValueType,
			Cardinality: card,
			Outcome:     outcome,
		}
		if outcome == OutcomePartial {
			pc.Similarity = sim
		}
		out.Comparisons = append(out.Comparisons, pc)
	}
	return out
}

func compareProperty(valueType, cardinality string, a, b []Value) (Outcome, float64) {
	aOK := hasCarrying(valueType, a)
	bOK := hasCarrying(valueType, b)
	if !aOK || !bOK {
		return OutcomeUnknown, 0
	}
	if anyCompatible(valueType, a, b) {
		return OutcomeAgree, 0
	}
	if valueType == properties.ValueTypeText {
		if sim := bestTextResemblance(a, b); sim >= 1 {
			return OutcomeAgree, 0
		} else if sim > 0 {
			return OutcomePartial, sim
		}
	}
	if cardinality == properties.CardinalityMultiple {
		return OutcomeUnknown, 0
	}
	return OutcomeConflict, 0
}

// bestTextResemblance is the closest carrying text pair. 0 when none resemble.
func bestTextResemblance(a, b []Value) float64 {
	var best float64
	for _, x := range a {
		if !x.HasText {
			continue
		}
		for _, y := range b {
			if !y.HasText {
				continue
			}
			if sim := TextResemblance(x.Text, y.Text); sim > best {
				best = sim
			}
		}
	}
	return best
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
