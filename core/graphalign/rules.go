package graphalign

import (
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// Config holds every section Align reads.
type Config struct {
	// Value is the registry's value section, with the catalog's name patterns
	// and alias tables (valuerules.Load fills a catalog's).
	Value valuecompare.Rules
	// Subject is the registry's subject section.
	Subject subjectcompare.Rules
	// Segment is the registry's segment section.
	Segment segmentcompare.Rules
	// Alignment is the registry's alignment section.
	Alignment Rules
}

// Rules is layer 6's section of the registry: candidates, the accept bar,
// bands and settling. The registry fills it; this package holds no numbers.
type Rules struct {
	// Shortlist is the keys the loader shortlists by.
	Shortlist shortlist.Rules
	// MinSimilarity is the least similarity for a shortlisted handle to be a
	// candidate.
	MinSimilarity float64
	// TopK is the candidates kept per Subject from its shortlist, and per layer
	// neighbor from one placed pair's links.
	TopK int
	// TieCap is the most candidates kept when ties at the K-th place pass K, per
	// Subject from its shortlist and per layer neighbor from one placed pair's
	// links.
	TieCap int
	// AcceptBar is the least evidence to place a pair.
	AcceptBar float64
	// Bands is the evidence a row's band starts at.
	Bands Bands
	// Alternatives is the alternatives listed per row.
	Alternatives int
	// SettleMargin is the lead a free rival needs over a placed pair to
	// release it, and over a decided row's handle to flag it.
	SettleMargin float64
	// SettleRounds is the most settle rounds.
	SettleRounds int
}

// Bands are the evidence a row's band starts at: strong at Strong or more,
// medium at Medium, weak at Weak, none below.
type Bands struct{ Strong, Medium, Weak float64 }
