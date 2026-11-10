package persondetail

import "github.com/mendahu/provenencia/core/vocabkey"

// ParticipantRoleOrder is participantRoleOrder, for the seed check.
func ParticipantRoleOrder() []vocabkey.Role { return participantRoleOrder }

// RecordLinksWith is recordLinks, for the tests that set the bounds.
var RecordLinksWith = recordLinks

// MarkerBefore, MarkerAfter and SameMarkerPlace are the marker-place helpers of
// lifeline_test.go, for the external timeline tests.
var (
	MarkerBefore    = markerBefore
	MarkerAfter     = markerAfter
	SameMarkerPlace = sameMarkerPlace
)
