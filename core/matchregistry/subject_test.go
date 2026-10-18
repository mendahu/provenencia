package matchregistry

import (
	"testing"

	"github.com/mendahu/provenencia/core/vocabkey"
)

// The subject section's structure: a pair belongs to a weighted lead, and no
// property is in two pairs.
func TestSubjectSection(t *testing.T) {
	s := Default().Subject
	inPair := map[vocabkey.SubjectType]map[vocabkey.PropertyKey]bool{}
	claim := func(kind vocabkey.SubjectType, key vocabkey.PropertyKey, what string) {
		if inPair[kind] == nil {
			inPair[kind] = map[vocabkey.PropertyKey]bool{}
		}
		if inPair[kind][key] {
			t.Errorf("%s.%s is in more than one pair (%s)", kind, key, what)
		}
		inPair[kind][key] = true
	}
	for kind, pools := range s.Pools {
		for lead, pool := range pools {
			if _, ok := s.Weights[kind][lead]; !ok {
				t.Errorf("pool lead %s.%s has no Weights entry", kind, lead)
			}
			if _, ok := s.Periods[kind][lead]; ok {
				t.Errorf("%s.%s has both a pool and a period", kind, lead)
			}
			claim(kind, lead, "pool lead")
			for _, with := range pool.With {
				if _, ok := s.Weights[kind][with]; ok {
					t.Errorf("pooled property %s.%s has a Weights entry of its own", kind, with)
				}
				claim(kind, with, "pooled")
			}
		}
	}
	for kind, periods := range s.Periods {
		for lead, p := range periods {
			if _, ok := s.Weights[kind][lead]; !ok {
				t.Errorf("period lead %s.%s has no Weights entry", kind, lead)
			}
			claim(kind, lead, "period lead")
			for _, end := range []vocabkey.PropertyKey{p.Start, p.End} {
				if _, ok := s.Weights[kind][end]; ok {
					t.Errorf("period property %s.%s has a Weights entry of its own", kind, end)
				}
				claim(kind, end, "period")
			}
		}
	}
}
