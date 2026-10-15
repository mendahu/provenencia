package graphalign

import (
	"math"

	"github.com/mendahu/provenencia/core/match"
)

// The registry: every walk and band tunable graph alignment reads, in one
// file. Tune here while dogfooding. Pairwise / Rank tunables stay in
// core/match. Tests pass Config overrides without editing these defaults.

// DefaultConfig is the shipped walk/band configuration.
func DefaultConfig() Config {
	return Config{
		AcceptScore:     2.0,
		StrongScore:     4.0,
		MediumScore:     3.0,
		WeakScore:       2.0,
		ConflictPenalty: 3.0,
		// Added when that Property agrees, on top of its log-odds. A small
		// catalog can make a real agreement look cheap; this is the finger
		// on the scale. Toponym +1 clears medium (3) and stays short of
		// strong (4) until an edge adds support.
		PropertyAgreement: map[match.Property]float64{
			{Key: "toponym", Origin: "provenencia"}: 1,
		},
		// Value-type defaults. A property entry replaces one of these.
		// Frequency weights are filled when the score is computed.
		ValueTypeScale: map[string]Scale{
			"text":    {Floor: 0, Frequency: true},
			"term":    {Floor: 1, Frequency: true},
			"integer": {Floor: 1, Frequency: true},
			// Any resemblance the date comparer reports is partial. A
			// disagreement scores 0 and conflicts. Open bounds that overlap
			// (before 2001, before 1985) resemble; they are not the same date.
			"date":    {Floor: 0, Frequency: true},
			"name":    {Floor: 1, Frequency: true},
			"subject": {Floor: 1, Frequency: true},
		},
		// Name: a different given name is about 0.6 and conflicts. A word
		// that exists only on one side does not lower the score, so Jake
		// Robins against James Kenneth Jake Robins agrees. An initial
		// (J. ~ James, 0.8) is 2.5 × 0.8, exactly the weak bar. The floor
		// sits above the given-name clash.
		// Sex and gender: agreement is a small nudge and a mismatch subtracts
		// 1 rather than the global penalty.
		PropertyScale: map[match.Property]Scale{
			{Key: "name", Origin: "provenencia"}: {
				Floor: 0.75, Weight: 2.5, Contradiction: 3,
			},
			{Key: "sex_at_birth", Origin: "provenencia"}: {
				Floor: 1, Weight: 0.6, Contradiction: 1,
			},
			{Key: "gender", Origin: "provenencia"}: {
				Floor: 1, Weight: 0.6, Contradiction: 1,
			},
		},
		EdgeSupportLow:     1.5, // fan-out ≈1 correspondence
		EdgeSupportHigh:    0.25,
		NeighborCreditLow:  0.5,  // fraction of the neighbor's property score
		NeighborCreditHigh: 0.25, // a common link passes less of that score
		FanOutLowMax:       1.5,  // signatures with fan-out ≤ this get the low pair
		FanOutUnknown:      2.0,  // a signature the catalog has no fan-out for yet
		AlternativeLimit:   3,
		MPrior: map[string]float64{
			"text":    0.9,
			"term":    0.95,
			"integer": 0.9,
			"date":    0.85,
			"name":    0.8,
			"subject": 0.9,
		},
		UPrior: 0.1, // cold-start agreement rate among unrelated
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
	AcceptScore       float64 // minimum score to accept a structure-backed candidate
	StrongScore       float64 // assessment ≥ this → strong
	MediumScore       float64 // assessment ≥ this (and < Strong) → medium
	WeakScore         float64 // assessment ≥ this (and < Medium) → weak; else no match
	ConflictPenalty   float64
	PropertyAgreement map[match.Property]float64 // extra weight when this Property agrees
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
	MPrior             map[string]float64 // value type → m
	UPrior             float64
	// Names scores name values when set. Nil keeps match.ComparerFor.
	Names *match.NameComparer
	// SignatureScale is how one edge signature combines with the candidate's
	// own property score. The walk looks a signature up and copies the
	// numbers onto the fact. A missing key is the zero value.
	SignatureScale map[string]SignatureScale
}

func (c Config) agreementWeight(p match.Property) float64 {
	if c.PropertyAgreement == nil {
		return 0
	}
	return c.PropertyAgreement[p]
}

func (c Config) mFor(valueType string) float64 {
	if c.MPrior != nil {
		if m, ok := c.MPrior[valueType]; ok && m > 0 && m < 1 {
			return m
		}
	}
	return 0.85
}

func (c Config) uOr(u float64) float64 {
	if u > 0 && u < 1 {
		return u
	}
	if c.UPrior > 0 && c.UPrior < 1 {
		return c.UPrior
	}
	return 0.1
}

// Scale is how one property's similarity becomes points. The scorer looks
// a property up here and does not name it.
type Scale struct {
	// Floor is the least similarity that earns points. Below it, a
	// single-value property conflicts and a multiple-value one is unknown.
	Floor float64
	// Weight is the points at similarity 1. The contribution is Weight ×
	// similarity. Ignored when Frequency is set; the log-odds fill it in.
	Weight float64
	// Contradiction is the penalty below Floor. Zero on a frequency scale
	// means the global ConflictPenalty.
	Contradiction float64
	// Frequency derives Weight from catalog frequency (m, u) plus any
	// agreement bonus, and lifts a lone exact agreement to the weak bar.
	// A fixed Weight is not lifted.
	Frequency bool
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

func logOdds(m, u float64) float64 {
	if m <= 0 || m >= 1 || u <= 0 || u >= 1 {
		return 0
	}
	return math.Log(m / u)
}
