package valuecompare

import "bytes"

// depiction compares likenesses by their image File: the same File is the same
// pixels. Two different crops can show one face and this layer cannot tell, so
// they are Unknown (cause unjudged) and the researcher judges them by eye.
type depiction struct{}

func (depiction) compare(a, b Value, _ Rules) (Result, Detail) {
	ia, ib := imageOf(a), imageOf(b)
	switch {
	case len(ia) == 0 || len(ib) == 0:
		return Result{Relation: Unknown}, Detail{Cause: CauseEmpty}
	case bytes.Equal(ia, ib):
		return Result{Relation: Same, Similarity: 1}, Detail{Cause: CauseExact}
	}
	return Result{Relation: Unknown}, Detail{Cause: CauseUnjudged}
}

// key is the image File id at both levels: a depiction has no looser form.
func (depiction) key(v Value, _ Relation, _ Rules) (string, bool) {
	id := imageOf(v)
	return string(id), len(id) > 0
}

func imageOf(v Value) []byte {
	if v.Depiction == nil {
		return nil
	}
	return v.Depiction.ImageFileID
}
