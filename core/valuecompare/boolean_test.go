package valuecompare_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestBoolean(t *testing.T) {
	rules := matchregistry.Default().Value
	for _, tc := range []struct {
		name    string
		a, b    bool
		want    valuecompare.Relation
		sim     float64
		cause   valuecompare.Cause
		keyA    string
		keyB    string
		sameKey bool
	}{
		{"true and true", true, true, valuecompare.Same, 1, valuecompare.CauseExact, "true", "true", true},
		{"false and false", false, false, valuecompare.Same, 1, valuecompare.CauseExact, "false", "false", true},
		{"true and false", true, false, valuecompare.Different, 0, valuecompare.CauseNoMatch, "true", "false", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, det := valuecompare.Compare(boolValue(tc.a), boolValue(tc.b), rules)
			if res.Relation != tc.want || res.Similarity != tc.sim || det.Cause != tc.cause {
				t.Errorf("Compare = %+v %+v", res, det)
			}
			for _, level := range []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent} {
				ka, oka := valuecompare.Key(boolValue(tc.a), level, rules)
				kb, okb := valuecompare.Key(boolValue(tc.b), level, rules)
				if !oka || !okb || ka != tc.keyA || kb != tc.keyB {
					t.Errorf("Key at %s = %q/%v, %q/%v", level, ka, oka, kb, okb)
				}
			}
		})
	}
}
