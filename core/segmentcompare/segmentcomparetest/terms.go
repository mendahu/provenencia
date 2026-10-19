// Package segmentcomparetest holds segment compare fixtures.
package segmentcomparetest

import (
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
		"subject", "father", "mother", "spouse", "child", "witness", "informant", "pallbearer",
	} {
		role[k] = segmentcompare.Term{}
	}
	return segmentcompare.Terms{
		vocabkey.BridgeParticipation: role,
		vocabkey.BridgeRelationship:  rel,
		vocabkey.BridgePlaceRelationship: {
			vocabkey.TermPartOf: {Directed: true},
			"succeeded_by":      {Directed: true},
		},
	}
}
