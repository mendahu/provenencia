package segmentcompare_test

import (
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
)

type registry struct {
	Value   valuecompare.Rules
	Subject subjectcompare.Rules
	Segment segmentcompare.Rules
}

// newRegistry is the shipped layers 1 and 3, with the segment rules the test
// file holds until STEP-3 moves them onto the registry.
func newRegistry() registry {
	d := matchregistry.Default()
	return registry{Value: d.Value, Subject: d.Subject, Segment: rules()}
}
