package graphalign

import "github.com/mendahu/provenencia/core/subjectcompare"

// Scored is the pairwise + edge result ScoreCandidate returns.
type Scored struct {
	Score  float64
	Edge   float64
	Result subjectcompare.Result
}

// ScoreCandidate combines a pair's entity compare result with edge support
// into one score. Align and any future caller share this path. edge is the
// summed segment evidence of the corresponding links.
func ScoreCandidate(r subjectcompare.Result, edge float64) Scored {
	return Scored{
		Score:  r.Evidence + edge,
		Edge:   edge,
		Result: r,
	}
}
