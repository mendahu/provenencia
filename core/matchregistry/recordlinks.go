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
		// decided: user, 2026-11-07 (S13-27 SEC-1): the walk charge counts each member's links and properties, so the cap keeps 20 times an obituary's work (2,215)
		MaxPairWork: 45_000,
		// decided: user, 2026-11-07 (S13-26 SEC-3): a high limit on one value, so a crafted value cannot hang a read; a longer note loses that one fact
		MaxValueBytes: 256,
		// decided: S13-27 plan D-3, the rule the user set for MaxPairWork (2026-11-07, S13-26 SEC-1): a high cap a researcher never reaches, so a crafted catalog cannot hang a read; 20 times the 53,900 work of a person in 50 household graphs
		MaxReadWork: 1_100_000,
	}
}
