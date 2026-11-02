package matchregistry

import "github.com/mendahu/provenencia/core/kinship"

// defaultKinship is the kinship patterns' shipped section.
func defaultKinship() kinship.Rules {
	return kinship.Rules{
		// decided: user, 2026-10-26 (S11-39 SEC-1): more than 2 spouses or parents, to allow for unusual families, but a bounded 10, so one malformed event cannot imply a million links
		MaxPivotEnd: 10,
		// decided: user, 2026-11-01 (S13-03 SEC-1): keep the first 10 steps of one parent–child pair, so duplicate links times children stays linear; real data files a handful
		MaxPairSteps: 10,
	}
}
