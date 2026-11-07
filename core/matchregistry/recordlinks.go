package matchregistry

import "github.com/mendahu/provenencia/core/recordlinks"

// defaultRecordLinks is the record-links comparison's shipped section.
func defaultRecordLinks() recordlinks.Rules {
	return recordlinks.Rules{
		// decided: user, 2026-11-07 (Spike 13 slice 7): two or more shared facts make a strong line
		StrongAt: 2,
	}
}
