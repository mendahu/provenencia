package graphalign

import (
	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The registry: every walk and band tunable graph alignment reads, in one
// file. Tune here while dogfooding. Value compare's numbers come from
// core/matchregistry as Config.Value. Tests pass Config overrides without
// editing these defaults.

// DefaultConfig is the shipped walk/band configuration.
func DefaultConfig() Config {
	reg := matchregistry.Default()
	// subject is a property's weights in the registry's subject section, as
	// a scale: the walk scores these properties on the same numbers.
	subject := func(kind vocabkey.SubjectType, key vocabkey.PropertyKey) Scale {
		w := reg.Subject.Weights[kind][key]
		return Scale{Weight: w.Agree, Contradiction: w.Conflict}
	}
	// accept is the least evidence to place a pair. plain, every value
	// type's weight, equals it, so one exact agreement alone reaches the
	// weak bar.
	accept := 2.0
	plain := accept
	// conflict is the points one disagreement costs.
	conflict := 3.0
	return Config{
		Value:           reg.Value,
		AcceptScore:     accept,
		StrongScore:     4.0,
		MediumScore:     3.0,
		WeakScore:       2.0,
		ConflictPenalty: conflict,
		// Value-type defaults. A property entry replaces one of these.
		// Any resemblance value compare reports is partial. A disagreement
		// scores 0 and conflicts. Open bounds that overlap (before 2001,
		// before 1985) resemble; they are not the same date.
		ValueTypeScale: map[string]Scale{
			"text":    {Weight: plain, Contradiction: conflict},
			"term":    {Weight: plain, Contradiction: conflict},
			"integer": {Weight: plain, Contradiction: conflict},
			"date":    {Weight: plain, Contradiction: conflict},
			"name":    {Weight: plain, Contradiction: conflict},
		},
		// Name: value compare decides a clash. A different first given name
		// or surname on each side is Different (0) and conflicts; a word on
		// one side only is a resemblance below a full match, so Jake Robins
		// against James Kenneth Jake Robins is 2.5 × 0.75. An initial
		// (J. ~ James, 0.8) is 2.5 × 0.8, exactly the weak bar.
		// Sex and gender: agreement is a small nudge and a mismatch subtracts
		// 1 rather than the global penalty.
		// Toponym: an exact toponym alone is 3, medium, and short of strong (4) until an edge adds support.
		PropertyScale: map[match.Property]Scale{
			{Key: "name", Origin: vocabkey.OriginProvenencia}:         subject(vocabkey.Person, vocabkey.PropertyName),
			{Key: "sex_at_birth", Origin: vocabkey.OriginProvenencia}: subject(vocabkey.Person, vocabkey.PropertySexAtBirth),
			{Key: "gender", Origin: vocabkey.OriginProvenencia}:       subject(vocabkey.Person, vocabkey.PropertySexAtBirth), // pooled with sex_at_birth (matchregistry subject.Pools)
			{Key: "toponym", Origin: vocabkey.OriginProvenencia}:      subject(vocabkey.Place, vocabkey.PropertyToponym),
		},
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
	AcceptScore     float64 // minimum score to accept a structure-backed candidate
	StrongScore     float64 // assessment ≥ this → strong
	MediumScore     float64 // assessment ≥ this (and < Strong) → medium
	WeakScore       float64 // assessment ≥ this (and < Medium) → weak; else no match
	ConflictPenalty float64
	// PropertyScale replaces the value-type scale for that Property.
	PropertyScale map[match.Property]Scale
	// ValueTypeScale is the scale for a value type with no property entry.
	ValueTypeScale     map[string]Scale
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
	// SignatureScale is how one edge signature combines with the candidate's
	// own property score. The walk looks a signature up and copies the
	// numbers onto the fact. A missing key is the zero value.
	SignatureScale map[string]SignatureScale
}

// Scale is how one property's similarity becomes points. The scorer looks
// a property up here and does not name it.
type Scale struct {
	// Weight is the points at similarity 1. The contribution is Weight × similarity.
	Weight float64
	// Contradiction is the penalty when value compare finds the values different (similarity 0).
	Contradiction float64
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
