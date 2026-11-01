package ingest

import (
	"github.com/mendahu/provenencia/core/database/observations"
)

// StageDepictions stages the image of every Observation input that carries a
// Depiction with no staged Image, and sets Image on it. Call it before the
// catalog session, never inside it, and defer the returned stages' Discard
// outside it and Settle inside it (withStagedDepictions in api/ffi/handlers
// does both):
//
//	stages, err := ingest.StageDepictions(dir, inputs)
//	if err != nil { ... }
//	defer stages.Discard()
//
// It does not read the Property: a Depiction on another value type is staged,
// then refused by the write, and Settle removes it. An input whose Image is
// already set is left as it is. On the first error every stage so far is
// discarded and the returned Stages is nil.
func StageDepictions(projectDir string, inputs []observations.Input) (Stages, error) {
	var staged Stages
	for i := range inputs {
		d := inputs[i].Depiction
		if d == nil || d.Image != nil {
			continue
		}
		s, err := StageImage(projectDir, d.ImagePath)
		if err != nil {
			staged.Discard()
			return nil, err
		}
		staged = append(staged, s)
		d.Image = s
	}
	return staged, nil
}
