package matchregistry

import "github.com/mendahu/provenencia/core/kinship"

// defaultKinship is the kinship patterns' shipped section.
func defaultKinship() kinship.Rules {
	return kinship.Rules{
		// decided: user, 2026-10-26 (S11-39 SEC-1): more than 2 spouses or parents, to allow for unusual families, but a bounded 10, so one malformed event cannot imply a million links
		MaxPivotEnd: 10,
		// decided: user, 2026-11-01 (S13-03 SEC-1): keep the first 10 steps of one parent–child pair, so duplicate links times children stays linear; real data files a handful
		MaxPairSteps: 10,
		// decided: user, 2026-11-07 (S13-26 SEC-1 rule, S13-29 SEC-2): one high total cap on the association copies one Derive's chains make, for a cost only a crafted catalog reaches; the largest Source in TestLargeSourceBudgets copies 4
		MaxChainWork: 100_000,
		// decided: user, 2026-11-08 (Spike 16 rule "Crafted fan-out is capped, not engineered"); a registry knob because kinship is a layer package (user, 2026-11-09): a real family holds a few dozen at most
		MaxKinPerRole: 100,
	}
}
