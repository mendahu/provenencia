package graphalign

import (
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare"
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
		Value:            reg.Value,
		AcceptScore:      accept,
		StrongScore:      4.0,
		MediumScore:      3.0,
		WeakScore:        2.0,
		ConflictPenalty:  3.0, // margin for ConflictWithFixed (S11-16)
		Subject:          reg.Subject,
		AlternativeLimit: 3,
		Segment:          reg.Segment,
	}
}

// Config holds every graph-alignment tunable Align reads.
type Config struct {
	AcceptScore      float64 // minimum score to accept a structure-backed candidate
	StrongScore      float64 // assessment ≥ this → strong
	MediumScore      float64 // assessment ≥ this (and < Strong) → medium
	WeakScore        float64 // assessment ≥ this (and < Medium) → weak; else no match
	ConflictPenalty  float64
	AlternativeLimit int
	// Value is the value section, with the catalog's name patterns and alias
	// tables. DefaultConfig holds the registry's; valuerules.Load fills a
	// catalog's.
	Value valuecompare.Rules
	// Subject is the registry's subject section: what each property counts.
	// DefaultConfig holds the registry's.
	Subject subjectcompare.Rules
	// Segment is the registry's segment section: what a link counts.
	// DefaultConfig holds the registry's.
	Segment segmentcompare.Rules
}
