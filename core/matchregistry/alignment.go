package matchregistry

import (
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Alignment is layer 6's section. It is declared here only until S11-16 deletes
// core/graphalign/registry.go, which is what keeps graphalign from declaring
// it; S11-16 replaces it with a layer-owned type and adds the walk's knobs.
type Alignment struct {
	// Shortlist is the keys a Subject is looked up by.
	Shortlist shortlist.Rules
}

// defaultAlignment is layer 6's shipped section.
func defaultAlignment() Alignment {
	return Alignment{Shortlist: shortlist.Rules{
		Words: map[vocabkey.SubjectType][]vocabkey.PropertyKey{
			vocabkey.Person: {vocabkey.PropertyName},    // decided: comparison-layers layer 6 procedure
			vocabkey.Place:  {vocabkey.PropertyToponym}, // decided: comparison-layers layer 6 procedure
		},
		Dated: map[vocabkey.SubjectType]shortlist.Dated{
			// decided: comparison-layers layer 6 procedure
			vocabkey.Event: {Type: vocabkey.PropertyEventType, Date: vocabkey.PropertyDate},
		},
		YearBucket: 2, // estimate: the date tolerance, so a Near date shares a bucket
	}}
}
