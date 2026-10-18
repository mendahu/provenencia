package valuecompare_test

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func termValue(key string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeTerm, Term: key}
}

func TestTermCompare(t *testing.T) {
	rules := matchregistry.Default().Value
	for _, tc := range []struct {
		name  string
		a, b  string
		want  valuecompare.Relation
		sim   float64
		cause valuecompare.Cause
	}{
		{"same", "male", "male", valuecompare.Same, 1, valuecompare.CauseExact},
		{"different", "male", "female", valuecompare.Different, 0, valuecompare.CauseNoMatch},
		{"empty", "", "male", valuecompare.Unknown, 0, valuecompare.CauseEmpty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, det := valuecompare.Compare(termValue(tc.a), termValue(tc.b), rules)
			if res.Relation != tc.want || math.Abs(res.Similarity-tc.sim) > 1e-9 || det.Cause != tc.cause {
				t.Errorf("Compare(%q, %q) = %+v %+v, want %s %v %s", tc.a, tc.b, res, det, tc.want, tc.sim, tc.cause)
			}
		})
	}
	for _, level := range []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent} {
		if k, ok := valuecompare.Key(termValue("male"), level, rules); !ok || k != "male" {
			t.Errorf("Key(male, %s) = %q, %v", level, k, ok)
		}
	}
}
