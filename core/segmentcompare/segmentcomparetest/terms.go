// Package segmentcomparetest holds segment compare fixtures.
package segmentcomparetest

import (
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Terms returns a new map of every seeded term of role, relationship_type and
// place_relationship_type under its bridge type, with the seed's directed and
// inverse marks. The subjectvocab seed test pins it to the seed.
func Terms() segmentcompare.Terms {
	pair := func(a, b vocabkey.BridgeTerm) map[vocabkey.BridgeTerm]segmentcompare.Term {
		return map[vocabkey.BridgeTerm]segmentcompare.Term{
			a: {Directed: true, Inverse: b},
			b: {Directed: true, Inverse: a},
		}
	}
	rel := map[vocabkey.BridgeTerm]segmentcompare.Term{"spouse": {}, "sibling": {}, "cousin": {}}
	for _, p := range [][2]vocabkey.BridgeTerm{
		{"parent", "child"}, {"grandparent", "grandchild"}, {"pibling", "nibling"}, {"guardian", "ward"},
	} {
		for k, v := range pair(p[0], p[1]) {
			rel[k] = v
		}
	}
	role := map[vocabkey.BridgeTerm]segmentcompare.Term{}
	for _, k := range []vocabkey.BridgeTerm{
		"subject", "father", "mother", "child", "witness", "informant", "pallbearer",
	} {
		role[k] = segmentcompare.Term{}
	}
	return segmentcompare.Terms{
		vocabkey.BridgeParticipation: role,
		vocabkey.BridgeRelationship:  rel,
		vocabkey.BridgePlaceRelationship: {
			vocabkey.TermPartOf:      {Directed: true},
			vocabkey.TermSucceededBy: {Directed: true},
		},
	}
}

// Bridge returns the seeded bridge type of key, failing t when there is none.
func Bridge(t testing.TB, key string) connectrules.Bridge {
	t.Helper()
	b, ok := connectrules.LookupBridge(key)
	if !ok {
		t.Fatalf("no %s bridge", key)
	}
	return b
}

// PartKeys are a participation read from the person's end and the event's,
// built through segmentcompare.Key over the seeded terms. It takes no testing.TB;
// TestKeysMatchSeededBridges holds it equal to the seeded bridge's keys.
func PartKeys(role vocabkey.BridgeTerm, eventType vocabkey.EventType) (person, event vocabkey.BridgeKey) {
	terms := Terms()
	person = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: role,
		End: vocabkey.PropertyPerson, Other: vocabkey.PropertyEvent, EventType: eventType}, terms)
	event = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: role,
		End: vocabkey.PropertyEvent, Other: vocabkey.PropertyPerson, EventType: eventType}, terms)
	return person, event
}

// RelKeys are a relationship read from the person's end and the related_to
// end, built the way PartKeys builds a participation.
func RelKeys(term vocabkey.BridgeTerm) (person, relatedTo vocabkey.BridgeKey) {
	terms := Terms()
	person = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: term,
		End: vocabkey.PropertyPerson, Other: vocabkey.PropertyRelatedTo, SameType: true}, terms)
	relatedTo = segmentcompare.Key(segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: term,
		End: vocabkey.PropertyRelatedTo, Other: vocabkey.PropertyPerson, SameType: true}, terms)
	return person, relatedTo
}
