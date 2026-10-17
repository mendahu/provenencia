package valuecompare_test

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func intValue(n int64) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeInteger, Integer: n}
}

func TestIntegerCompare(t *testing.T) {
	for _, tc := range []struct {
		name      string
		a, b      int64
		tolerance int64
		want      valuecompare.Relation
		sim       float64
		cause     valuecompare.Cause
		distance  float64
	}{
		{"equal", 3, 3, 0, valuecompare.Same, 1, valuecompare.CauseExact, 0},
		{"gap at tolerance 0", 3, 4, 0, valuecompare.Different, 0, valuecompare.CauseBeyondTolerance, 1},
		{"gap within tolerance", 3, 4, 2, valuecompare.Near, 1 - 1.0/3, valuecompare.CauseGap, 1},
		{"gap beyond tolerance", 3, 6, 2, valuecompare.Different, 0, valuecompare.CauseBeyondTolerance, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rules := matchregistry.Default().Value
			rules.Integers.Tolerance = tc.tolerance
			res, det := valuecompare.Compare(intValue(tc.a), intValue(tc.b), rules)
			if res.Relation != tc.want || math.Abs(res.Similarity-tc.sim) > 1e-9 || det.Cause != tc.cause ||
				det.Distance != tc.distance || det.Unit != valuecompare.UnitCount {
				t.Errorf("Compare(%d, %d) = %+v %+v", tc.a, tc.b, res, det)
			}
		})
	}
}
