package graphalign

import "github.com/mendahu/provenencia/core/subjectcompare"

// Scored is the pairwise + edge result ScoreCandidate returns.
type Scored struct {
	Score  float64
	Edge   float64
	Result subjectcompare.Result
}

// ScoreCandidate combines a pair's entity compare result with edge support
// into one score. Align and any future caller share this path. provenance
// scales the total (≤0 treated as 1). edge is the summed segment evidence of
// the corresponding links.
func ScoreCandidate(r subjectcompare.Result, edge, provenance float64) Scored {
	if provenance <= 0 {
		provenance = 1
	}
	return Scored{
		Score:  (r.Evidence + edge) * provenance,
		Edge:   edge,
		Result: r,
	}
}
