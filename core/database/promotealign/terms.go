package promotealign

import (
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// observedTerms maps each Subject to the key of a positive Observation's term
// on the Property, among those the Evidence graph records for the listed
// Subjects. The first key in order wins.
func observedTerms(eg graphcache.EvidenceGraph, props map[string]graphcache.VocabProperty, terms map[string]graphcache.VocabTerm, subjectIDs [][]byte, propertyKey string) map[string]string {
	out := map[string]string{}
	if propertyKey == "" {
		return out
	}
	want := make(map[string]bool, len(subjectIDs))
	for _, id := range subjectIDs {
		want[string(id)] = true
	}
	for _, o := range eg.Observations {
		if o.Polarity != observations.PolarityPositive || !want[string(o.SubjectID)] || len(o.TermID) == 0 {
			continue
		}
		if p, ok := props[string(o.PropertyID)]; !ok || p.Key != propertyKey || p.Origin != vocabkey.OriginProvenencia {
			continue
		}
		t, ok := terms[string(o.TermID)]
		if !ok {
			continue
		}
		if cur, ok := out[string(o.SubjectID)]; !ok || t.Key < cur {
			out[string(o.SubjectID)] = t.Key
		}
	}
	return out
}
