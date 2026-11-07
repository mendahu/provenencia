package segmentcomparetest_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func TestKeysMatchSeededBridges(t *testing.T) {
	terms := segmentcomparetest.Terms()
	for _, c := range []struct {
		name      string
		role      vocabkey.BridgeTerm
		eventType vocabkey.EventType
	}{
		{"subject birth", "subject", vocabkey.EventBirth},
		{"subject any event", "subject", ""},
		{"father birth", "father", vocabkey.EventBirth},
	} {
		t.Run("PartKeys/"+c.name, func(t *testing.T) {
			gp, ge := segmentcomparetest.PartKeys(c.role, c.eventType)
			wp, we := segmentcompare.Keys(segmentcomparetest.Bridge(t, vocabkey.BridgeParticipation), c.role, c.eventType, terms)
			if gp != wp || ge != we {
				t.Errorf("PartKeys = %v, %v; seeded bridge = %v, %v", gp, ge, wp, we)
			}
		})
	}
	for _, term := range []vocabkey.BridgeTerm{vocabkey.TermParent, vocabkey.TermSpouse} {
		t.Run("RelKeys/"+string(term), func(t *testing.T) {
			gp, gr := segmentcomparetest.RelKeys(term)
			wp, wr := segmentcompare.Keys(segmentcomparetest.Bridge(t, vocabkey.BridgeRelationship), term, "", terms)
			if gp != wp || gr != wr {
				t.Errorf("RelKeys = %v, %v; seeded bridge = %v, %v", gp, gr, wp, wr)
			}
		})
	}
}
