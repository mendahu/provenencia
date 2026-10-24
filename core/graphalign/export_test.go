package graphalign

import (
	"bytes"
	"slices"
)

// NominatedForTest runs Align's procedure and returns the ids of the handles
// the Subject ended with as candidates, sorted.
func NominatedForTest(layer Layer, canon Canon, fixed []Fixed, cfg Config, subjectID []byte) [][]byte {
	st := run(layer, canon, fixed, cfg)
	var out [][]byte
	for _, c := range st.cands[string(subjectID)] {
		out = append(out, c.handleID)
	}
	slices.SortFunc(out, bytes.Compare)
	return out
}
