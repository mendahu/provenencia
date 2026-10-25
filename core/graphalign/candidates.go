package graphalign

import (
	"bytes"
	"sort"

	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Candidates are the shortlisted handles a Subject is compared with:
// similarity at least MinSimilarity, the best TopK by similarity, and every
// handle tied with the TopK-th, up to TieCap. Evidence decides everything
// after.
//
// It returns the handles of kind in that order: similarity (high first), then
// evidence (high first), then Ref, then ID. A placed pair's nominations are cut
// by the same rule, with no similarity gate (keepTop).
func Candidates(kind string, values subjectcompare.Values, handles []Handle, metas subjectcompare.Metas, cfg Config) []Handle {
	var found []scored
	for _, h := range handles {
		if h.Kind != kind {
			continue
		}
		r := subjectcompare.Compare(vocabkey.SubjectType(kind), values, h.Values, metas, cfg.Subject, cfg.Value)
		if r.Similarity < cfg.Alignment.MinSimilarity {
			continue
		}
		found = append(found, scored{h, r.Evidence, r.Similarity})
	}
	return keepTop(found, cfg.Alignment)
}

// scored is a handle with the subject compare of a Subject against it.
type scored struct {
	h          Handle
	evidence   float64
	similarity float64
}

// keepTop sorts found in place by similarity (high first), then evidence (high
// first), then Ref, then ID, and keeps the best TopK, every handle tied with
// the TopK-th, up to TieCap.
func keepTop(found []scored, r Rules) []Handle {
	sort.SliceStable(found, func(i, j int) bool {
		a, b := found[i], found[j]
		if a.similarity != b.similarity {
			return a.similarity > b.similarity
		}
		if a.evidence != b.evidence {
			return a.evidence > b.evidence
		}
		if a.h.Ref != b.h.Ref {
			return a.h.Ref < b.h.Ref
		}
		return bytes.Compare(a.h.ID, b.h.ID) < 0
	})
	keep := min(r.TopK, len(found))
	for keep > 0 && keep < len(found) && found[keep].similarity == found[keep-1].similarity {
		keep++
	}
	keep = min(keep, r.TieCap)
	var out []Handle
	for _, f := range found[:keep] {
		out = append(out, f.h)
	}
	return out
}
