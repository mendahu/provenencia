package graphalign

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Candidates are the handles a Subject is offered on its own properties:
// entity compare evidence at least the weak band. Seeding passes it shortlisted
// handles (S11-15); top K replaces the limit (S11-16).
//
// It returns the handles of kind whose evidence against values is at least
// cfg.Alignment.Bands.Weak, best first: evidence, then similarity, then Ref, then ID.
// At most limit; all when limit <= 0.
func Candidates(kind string, values subjectcompare.Values, handles []Handle, metas subjectcompare.Metas, cfg Config, limit int) []Handle {
	type scored struct {
		h          Handle
		evidence   float64
		similarity float64
	}
	var found []scored
	for _, h := range handles {
		if h.Kind != kind {
			continue
		}
		r := subjectcompare.Compare(vocabkey.SubjectType(kind), values, h.Values, metas, cfg.Subject, cfg.Value)
		if r.Evidence < cfg.Alignment.Bands.Weak {
			continue
		}
		found = append(found, scored{h, r.Evidence, r.Similarity})
	}
	sort.SliceStable(found, func(i, j int) bool {
		a, b := found[i], found[j]
		if a.evidence != b.evidence {
			return a.evidence > b.evidence
		}
		if a.similarity != b.similarity {
			return a.similarity > b.similarity
		}
		if a.h.Ref != b.h.Ref {
			return a.h.Ref < b.h.Ref
		}
		return bytes.Compare(a.h.ID, b.h.ID) < 0
	})
	if limit > 0 && len(found) > limit {
		found = found[:limit]
	}
	var out []Handle
	for _, f := range found {
		out = append(out, f.h)
	}
	return out
}
