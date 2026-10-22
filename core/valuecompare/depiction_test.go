package valuecompare_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestCompareDepiction(t *testing.T) {
	rules := matchregistry.Default().Value
	for _, tc := range []struct {
		name  string
		a, b  byte
		want  valuecompare.Relation
		sim   float64
		cause valuecompare.Cause
	}{
		{"same image", 1, 1, valuecompare.Same, 1, valuecompare.CauseExact},
		{"two images", 1, 2, valuecompare.Unknown, 0, valuecompare.CauseUnjudged},
		{"empty side", 0, 1, valuecompare.Unknown, 0, valuecompare.CauseEmpty},
		{"both empty", 0, 0, valuecompare.Unknown, 0, valuecompare.CauseEmpty},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, det := valuecompare.Compare(pic(tc.a), pic(tc.b), rules)
			if res.Relation != tc.want || res.Similarity != tc.sim || det.Cause != tc.cause {
				t.Errorf("Compare = %+v %+v, want %s %v %s", res, det, tc.want, tc.sim, tc.cause)
			}
		})
	}
	ka, oka := valuecompare.Key(pic(1), valuecompare.Same, rules)
	kb, okb := valuecompare.Key(pic(1), valuecompare.Equivalent, rules)
	kc, _ := valuecompare.Key(pic(2), valuecompare.Same, rules)
	if !oka || !okb || ka != kb || ka == kc {
		t.Errorf("keys %q %q %q", ka, kb, kc)
	}
	if k, ok := valuecompare.Key(pic(0), valuecompare.Same, rules); ok || k != "" {
		t.Errorf("empty depiction key = %q, %v", k, ok)
	}
}
