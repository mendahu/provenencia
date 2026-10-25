package matchregistry

import (
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The segment section's structure: a one-cardinality fact is a participation
// subject key holding from both ends, a gate names a registered bridge, and
// every support and credit is positive.
func TestSegmentSection(t *testing.T) {
	s := Default().Segment
	for _, k := range s.One {
		if k.Type != vocabkey.BridgeParticipation || k.Term != vocabkey.BridgeTerm(vocabkey.RoleSubject) || k.End != "" {
			t.Errorf("One key %s is not a participation subject key with an empty End", k)
		}
	}
	bridges := map[string]bool{}
	for _, b := range connectrules.BridgeTypeKeys() {
		bridges[b] = true
	}
	for k := range s.Gates {
		if !bridges[string(k.Type)] {
			t.Errorf("Gates key %s names no connectrules bridge type", k)
		}
	}
	for name, v := range map[string]float64{
		"OneSupport": s.OneSupport, "OneCredit": s.OneCredit, "OneConflict": s.OneConflict,
		"OtherSupport": s.OtherSupport, "OtherCredit": s.OtherCredit,
	} {
		if v <= 0 {
			t.Errorf("%s = %v, want positive", name, v)
		}
	}
}
