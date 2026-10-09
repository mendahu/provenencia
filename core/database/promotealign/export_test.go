package promotealign

import "github.com/mendahu/provenencia/core/graphalign"

// LoadStatsForTest exposes the cached stats read.
var LoadStatsForTest = loadStats

// SetCanonLimitsForTest overrides the expansion bounds and returns a restore.
func SetCanonLimitsForTest(edgesPerStep, handles int) func() {
	oldEdges, oldHandles := maxEdgesPerStep, maxCanonHandles
	maxEdgesPerStep, maxCanonHandles = edgesPerStep, handles
	return func() { maxEdgesPerStep, maxCanonHandles = oldEdges, oldHandles }
}

// CanonForTest is the canon Propose would align a Source against.
func CanonForTest(q Querier, sourceID []byte, fixed []graphalign.Fixed) (graphalign.Canon, error) {
	layer, primary, err := loadLayer(q, sourceID)
	if err != nil {
		return graphalign.Canon{}, err
	}
	anchors, err := mergeFixed(q, sourceID, primary, fixed)
	if err != nil {
		return graphalign.Canon{}, err
	}
	return loadCanon(q, layer, anchors)
}
