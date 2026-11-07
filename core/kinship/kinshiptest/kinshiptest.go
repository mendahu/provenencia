// Package kinshiptest holds kinship fixtures: the edge a participation or a
// relationship makes, built over the seeded terms.
package kinshiptest

import (
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Part is a participation edge with the person at A and the event at B, its
// keys built through segmentcomparetest.PartKeys, recorded by assocs.
func Part(person, event []byte, role vocabkey.BridgeTerm, eventType vocabkey.EventType, assocs ...[]byte) kinship.Edge {
	ka, kb := segmentcomparetest.PartKeys(role, eventType)
	return kinship.Edge{A: person, B: event, KeyA: ka, KeyB: kb, Associations: assocs}
}

// Rel is a relationship edge with the person end at A and the related_to end at
// B, its keys built through segmentcomparetest.RelKeys, recorded by assocs.
func Rel(person, relatedTo []byte, term vocabkey.BridgeTerm, assocs ...[]byte) kinship.Edge {
	ka, kb := segmentcomparetest.RelKeys(term)
	return kinship.Edge{A: person, B: relatedTo, KeyA: ka, KeyB: kb, Associations: assocs}
}
