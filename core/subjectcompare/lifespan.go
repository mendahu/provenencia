package subjectcompare

import (
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// dateMeta is the fixed meta of a derived lifespan: a derived key has no
// catalog row, so its meta sits beside the rule that reads it.
var dateMeta = propertymeta.Meta{ValueType: propertymeta.ValueTypeDate, Cardinality: propertymeta.CardinalitySingle}

// lifespan checks each side's span against the other side's window. A
// conflict from either direction is the result, then an unknown; otherwise the
// lifespans are consistent. A lifespan's agree means consistent: its speaking
// pair keeps the relation that decided it, which may be partial (`contains`,
// `near`).
func lifespan(window vocabkey.PropertyKey, a, b Values, l Lifespan, value valuecompare.Rules) propertycompare.Result {
	ab := propertycompare.Compare(a[window], b[l.Span], dateMeta, value)
	ba := propertycompare.Compare(a[l.Span], b[window], dateMeta, value)
	switch {
	case ab.Outcome == propertycompare.OutcomeConflict:
		return ab
	case ba.Outcome == propertycompare.OutcomeConflict:
		return ba
	case ab.Outcome == propertycompare.OutcomeUnknown:
		return ab
	case ba.Outcome == propertycompare.OutcomeUnknown:
		return ba
	}
	ab.Outcome = propertycompare.OutcomeAgree
	return ab
}
