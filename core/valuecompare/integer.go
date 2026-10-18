package valuecompare

import "fmt"

// integer compares counts and other whole numbers. Equal is Same; a gap within
// Tolerance is Near and falls off with the gap; beyond it they disagree.
type integer struct{}

func (integer) compare(a, b Value, r Rules) (Result, Detail) {
	// The gap as a uint64 is exact for any pair of int64s; float64 would lose
	// it past 2^53.
	lo, hi := a.Integer, b.Integer
	if lo > hi {
		lo, hi = hi, lo
	}
	gap := uint64(hi) - uint64(lo)
	det := Detail{Distance: float64(gap), Unit: UnitCount}
	switch {
	case gap == 0:
		det.Cause = CauseExact
		return Result{Relation: Same, Similarity: 1}, det
	case gap <= uint64(r.Integers.Tolerance):
		det.Cause = CauseGap
		return Result{Relation: Near, Similarity: falloff(det.Distance, float64(r.Integers.Tolerance))}, det
	}
	det.Cause = CauseBeyondTolerance
	return Result{Relation: Different}, det
}

// key is the integer's decimal text at both levels.
func (integer) key(v Value, _ Relation, _ Rules) (string, bool) {
	return fmt.Sprint(v.Integer), true
}
