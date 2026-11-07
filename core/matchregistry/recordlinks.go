package matchregistry

import "github.com/mendahu/provenencia/core/recordlinks"

// defaultRecordLinks is the record-links comparison's shipped section.
func defaultRecordLinks() recordlinks.Rules {
	return recordlinks.Rules{
		// decided: user, 2026-11-07 (Spike 13 slice 7): two or more shared facts make a strong line
		StrongAt: 2,
		// decided: user, 2026-11-07 (S13-26 SEC-1): an entity holding more than ten members in a graph is malformed; real records file a handful
		MaxEntityMembers: 10,
		// decided: user, 2026-11-07 (S13-26 SEC-1): more than twenty values for one property is malformed; real records hold a few
		MaxUnitValues: 20,
		// decided: user, 2026-11-07 (S13-26 SEC-1): a high cap a researcher never reaches, so a crafted catalog cannot hang a read
		MaxPairWork: 25_000,
	}
}
