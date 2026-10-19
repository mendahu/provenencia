package graphalign

import (
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// The registry: every walk and band tunable graph alignment reads, in one
// file. Tune here while dogfooding. Value compare's numbers come from
// core/matchregistry as Config.Value. Tests pass Config overrides without
// editing these defaults.

// DefaultConfig is the shipped walk/band configuration.
func DefaultConfig() Config {
	reg := matchregistry.Default()
	// accept is the least evidence to place a pair.
	accept := 2.0
	return Config{
		Value:              reg.Value,
		AcceptScore:        accept,
		StrongScore:        4.0,
		MediumScore:        3.0,
		WeakScore:          2.0,
		ConflictPenalty:    3.0, // margin for ConflictWithFixed (S11-16)
		Subject:            reg.Subject,
		EdgeSupportLow:     1.5, // fan-out ≈1 correspondence
		EdgeSupportHigh:    0.25,
		NeighborCreditLow:  0.5,  // fraction of the neighbor's property score
		NeighborCreditHigh: 0.25, // a common link passes less of that score
		FanOutLowMax:       1.5,  // signatures with fan-out ≤ this get the low pair
		FanOutUnknown:      2.0,  // a signature the catalog has no fan-out for yet
		AlternativeLimit:   3,
		// Keyed by EdgeSignature.ScaleKey. A missing entry is the zero
		// scale: support is added even when the candidate's own properties
		// contribute nothing, and a different neighbor is silence.
		// part_of is one-to-many from the whole. Both ends require a property
		// score that already clears the accept bar. Only the part cancels
		// that score when the parent is a different handle: cancelling the
		// whole while its part is still settling never lets the chain agree.
		SignatureScale: map[string]SignatureScale{
			"place_relationship|part_of|place||from": {MinNode: 2, ConflictFactor: 1},
			"place_relationship|part_of|place||to":   {MinNode: 2},
		},
	}
}

// Config holds every graph-alignment tunable Align reads.
type Config struct {
	AcceptScore        float64 // minimum score to accept a structure-backed candidate
	StrongScore        float64 // assessment ≥ this → strong
	MediumScore        float64 // assessment ≥ this (and < Strong) → medium
	WeakScore          float64 // assessment ≥ this (and < Medium) → weak; else no match
	ConflictPenalty    float64
	EdgeSupportLow     float64
	EdgeSupportHigh    float64
	NeighborCreditLow  float64 // times the neighbor's property score when fan-out is low
	NeighborCreditHigh float64
	FanOutLowMax       float64
	FanOutUnknown      float64
	AlternativeLimit   int
	// Value is the value section, with the catalog's name patterns and alias
	// tables. DefaultConfig holds the registry's; valuerules.Load fills a
	// catalog's.
	Value valuecompare.Rules
	// Subject is the registry's subject section: what each property counts.
	// DefaultConfig holds the registry's.
	Subject subjectcompare.Rules
	// SignatureScale is how one edge signature combines with the candidate's
	// own property score. The walk looks a signature up and copies the
	// numbers onto the fact. A missing key is the zero value.
	SignatureScale map[string]SignatureScale
}

// SignatureScale is how one edge combines with the candidate's own property
// score. The zero value adds support even when that score is nothing, and
// leaves a mismatched neighbor silent.
type SignatureScale struct {
	// MinNode is the property score required before this signature's support
	// and credit are added. Zero adds them even when the candidate's own
	// properties contribute nothing.
	MinNode float64
	// ConflictFactor times the property score is subtracted when this
	// candidate has a same-end neighbor under the signature and it is a
	// different handle. Zero leaves a mismatch silent.
	ConflictFactor float64
}
