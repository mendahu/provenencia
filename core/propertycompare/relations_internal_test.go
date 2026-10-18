package propertycompare

import (
	"testing"

	"github.com/mendahu/provenencia/core/valuecompare"
)

// TestEveryRelationHasAnOutcome fails until a new relation has an outcome in
// outcomeOf and a contradiction rule in contradicts.
func TestEveryRelationHasAnOutcome(t *testing.T) {
	want := map[valuecompare.Relation]struct {
		outcome     Outcome
		contradicts bool
	}{
		valuecompare.Same:       {OutcomeAgree, true},
		valuecompare.Equivalent: {OutcomeAgree, true},
		valuecompare.Contains:   {OutcomePartial, true},
		valuecompare.Overlaps:   {OutcomePartial, false},
		valuecompare.Near:       {OutcomePartial, false},
		valuecompare.Different:  {OutcomeConflict, false},
		valuecompare.Unknown:    {OutcomeUnknown, false},
	}
	for _, r := range valuecompare.Relations() {
		w, ok := want[r]
		if !ok {
			t.Errorf("relation %s has no row: decide its outcome and whether a negative contradicts it", r)
			continue
		}
		if got := outcomeOf(r); got != w.outcome {
			t.Errorf("outcomeOf(%s) = %s, want %s", r, got, w.outcome)
		}
		if got := contradicts(pair{res: valuecompare.Result{Relation: r}}); got != w.contradicts {
			t.Errorf("contradicts(%s) = %v, want %v", r, got, w.contradicts)
		}
	}
	if len(want) != len(valuecompare.Relations()) {
		t.Errorf("table has %d rows, Relations() has %d", len(want), len(valuecompare.Relations()))
	}
}
