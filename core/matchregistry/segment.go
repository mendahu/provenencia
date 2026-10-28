package matchregistry

import (
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// defaultSegment is layer 5's shipped section: what a matching or differing
// link counts.
func defaultSegment() segmentcompare.Rules {
	subject := vocabkey.BridgeTerm(vocabkey.RoleSubject)
	partOf := func(end vocabkey.PropertyKey) vocabkey.BridgeKey {
		return vocabkey.BridgeKey{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: end}
	}
	return segmentcompare.Rules{
		// decided: a person has one birth and one death, and each has one subject
		One: []vocabkey.BridgeKey{
			{Type: vocabkey.BridgeParticipation, Term: subject, EventType: vocabkey.EventBirth},
			{Type: vocabkey.BridgeParticipation, Term: subject, EventType: vocabkey.EventDeath},
		},
		OneSupport:   1.5,  // carried from graphalign EdgeSupportLow
		OneCredit:    0.5,  // carried from graphalign NeighborCreditLow
		OneConflict:  3,    // estimate: a differing birth or death is as strong as a differing name or date (subject.Weights.person.name.Conflict)
		// decided: user, 2026-10-28 (S14-08): dogfood Jake + grandparent under Weak at 0.25; 0.75 clears Weak (~2.43) without auto-placing kinship-alone pairs (still under AcceptBar 2)
		OtherSupport: 0.75,
		OtherCredit:  0.25, // carried from graphalign NeighborCreditHigh
		// estimate: the accept bar, a neighbor that could be placed on its own properties
		UnplacedAlike: 2,
		Gates: map[vocabkey.BridgeKey]segmentcompare.Gate{
			// carried from graphalign's old part_of scale, from end
			partOf(vocabkey.PropertyFrom): {MinOwnEvidence: 2, ConflictFactor: 1},
			// carried from graphalign's old part_of scale, to end; decided: the whole's end had no conflict factor
			partOf(vocabkey.PropertyTo): {MinOwnEvidence: 2, ConflictFactor: 0},
		},
	}
}
