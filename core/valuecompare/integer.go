package valuecompare

import (
	"fmt"
	"math"
)

// integer compares counts and other whole numbers. Equal is Same; a gap within
// Tolerance is Near and falls off with the gap; beyond it they disagree.
type integer struct{}

func (integer) compare(a, b Value, r Rules) (Result, Detail) {
	gap := math.Abs(float64(a.Integer) - float64(b.Integer))
	det := Detail{Distance: gap, Unit: UnitCount}
	switch {
	case a.Integer == b.Integer:
		det.Cause = CauseExact
		return Result{Relation: Same, Similarity: 1}, det
	case gap <= float64(r.Integers.Tolerance):
		det.Cause = CauseGap
		return Result{Relation: Near, Similarity: falloff(gap, float64(r.Integers.Tolerance))}, det
	}
	det.Cause = CauseBeyondTolerance
	return Result{Relation: Different}, det
}

// key is the integer's decimal text at both levels.
func (integer) key(v Value, _ Relation, _ Rules) (string, bool) {
	return fmt.Sprint(v.Integer), true
}
