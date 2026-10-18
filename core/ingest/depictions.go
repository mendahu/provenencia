package ingest

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/observations"
)

// StageDepictions stages the image of every Observation input that carries a
// Depiction with no staged Image, and sets Image on it. Call it before
// writes.Run, never inside it, and defer the returned discard:
//
//	discard, err := ingest.StageDepictions(c, inputs)
//	if err != nil { ... }
//	defer discard()
//
// It does not read the Property: a Depiction on another value type is staged,
// then refused by the write, and discard removes it. An input whose Image is
// already set is left as it is. On the first error every stage so far is
// discarded and the returned discard is a no-op.
func StageDepictions(c *database.Catalog, inputs []observations.Input) (discard func(), err error) {
	var staged []*StagedImage
	discardAll := func() {
		for _, s := range staged {
			s.Discard()
		}
	}
	for i := range inputs {
		d := inputs[i].Depiction
		if d == nil || d.Image != nil {
			continue
		}
		s, err := StageImage(c, d.ImagePath)
		if err != nil {
			discardAll()
			return func() {}, err
		}
		staged = append(staged, s)
		d.Image = s
	}
	return discardAll, nil
}
