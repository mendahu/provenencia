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
		EdgeSupportLow:   1.5, // fan-out ≈1 correspondence
		EdgeSupportHigh:  0.25,
		FanOutLowMax:     1.5, // signatures with fan-out ≤ this get EdgeSupportLow
		FanOutUnknown:    2.0, // a signature the catalog has no fan-out for yet
		AlternativeLimit: 3,
		MPrior: map[string]float64{
			"text":    0.9,
			"term":    0.95,
			"integer": 0.9,
			"date":    0.85,
			"name":    0.8,
			"subject": 0.9,
		},
		UPrior: 0.1, // cold-start agreement rate among unrelated
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
	EdgeSupportLow    float64
	EdgeSupportHigh   float64
	FanOutLowMax      float64
	FanOutUnknown     float64
	AlternativeLimit  int
	MPrior            map[string]float64 // value type → m
	UPrior            float64
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

func logOdds(m, u float64) float64 {
	if m <= 0 || m >= 1 || u <= 0 || u >= 1 {
		return 0
	}
	return math.Log(m / u)
}
