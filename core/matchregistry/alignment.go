package matchregistry

import (
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// defaultAlignment is layer 6's shipped section.
func defaultAlignment() graphalign.Rules {
	return graphalign.Rules{
		Shortlist: shortlist.Rules{
			Words: map[vocabkey.SubjectType][]vocabkey.PropertyKey{
				vocabkey.Person: {vocabkey.PropertyName},    // decided: comparison-layers layer 6 procedure
				vocabkey.Place:  {vocabkey.PropertyToponym}, // decided: comparison-layers layer 6 procedure
			},
			Dated: map[vocabkey.SubjectType]shortlist.Dated{
				// decided: comparison-layers layer 6 procedure
				vocabkey.Event: {Type: vocabkey.PropertyEventType, Date: vocabkey.PropertyDate},
			},
			YearBucket: 2,   // estimate: the date tolerance, so a Near date shares a bucket
			MaxBuckets: 100, // estimate: two centuries at YearBucket 2; a wider reach keys as open
		},
		MinSimilarity: 0.5,                                             // estimate: at least half alike; a name with a sex conflict is 0.71, a birth against a death 0.09
		TopK:          5,                                               // carried from promotealign seedCandidateLimit
		TieCap:        20,                                              // estimate: room for a common name's ties without unbounded work
		AcceptBar:     2,                                               // carried from graphalign AcceptScore
		Bands:         graphalign.Bands{Strong: 4, Medium: 3, Weak: 2}, // carried from graphalign StrongScore, MediumScore, WeakScore
		Alternatives:  3,                                               // carried from graphalign AlternativeLimit
		SettleMargin:  1,                                               // estimate: one point, a third of a band
		SettleRounds:  10,                                              // estimate: a safety cap; dirty pairs normally settle in two or three
	}
}
