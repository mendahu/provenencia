// Package matchregistry holds every matching tunable, one section per layer
// group. Layers take their section as an argument; Default() is the shipped
// set.
//
// Layers declare their section's types and never import this package, so the
// registry can name a layer's types without a cycle. Numbers are written only
// here (guard_test.go keeps the layers free of them).
package matchregistry

import (
	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// Registry is the whole set of matching tunables.
type Registry struct {
	// Value is layer 1: how two values of one type compare.
	Value valuecompare.Rules
	// Subject is layer 3: what each property counts when two Subjects of one
	// type are compared.
	Subject subjectcompare.Rules
	// Segment is layer 5: what a matching or differing link counts.
	Segment segmentcompare.Rules
	// Alignment is layer 6: candidates, the accept bar, bands and settling.
	Alignment graphalign.Rules
	// Kinship is the bound on what one pivot of a kinship pattern implies.
	Kinship kinship.Rules
	// Evidence is the claim evidence view: which lines start checked, and what
	// a relationship line bundles. It is not evidence points; no comparison
	// layer reads it.
	Evidence claimevidence.Rules
}

// Default is the shipped registry. Each call builds new slices and maps, so a
// caller that changes its copy changes no one else's.
func Default() Registry {
	return Registry{Value: defaultValue(), Subject: defaultSubject(), Segment: defaultSegment(), Alignment: defaultAlignment(), Kinship: defaultKinship(), Evidence: defaultEvidence()}
}

// AlignConfig is the registry as graph alignment reads it.
func (r Registry) AlignConfig() graphalign.Config {
	return graphalign.Config{Value: r.Value, Subject: r.Subject, Segment: r.Segment, Alignment: r.Alignment}
}
