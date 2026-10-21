package matchregistry

import (
	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// defaultEvidence is the claim evidence view's shipped section: which lines
// start checked, which can be pinned, and what a relationship line bundles.
func defaultEvidence() claimevidence.Rules {
	return claimevidence.Rules{
		// decided: a line whose values are the same starts checked (claim-evidence.md, Lines)
		Checked: []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent},
		// decided: a line whose values overlap in part can be pinned, unchecked (claim-evidence.md, Lines)
		Pinnable: []valuecompare.Relation{valuecompare.Contains, valuecompare.Overlaps, valuecompare.Near},
		Defining: map[vocabkey.SubjectType][]vocabkey.PropertyKey{
			// decided: an event is told apart by its type (claim-evidence.md, Evidence through relationships)
			vocabkey.Event: {vocabkey.PropertyEventType},
		},
	}
}
