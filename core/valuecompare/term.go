package valuecompare

// term compares catalog vocabulary keys: slugs, so byte-equal or different,
// with no normalization.
type term struct{}

func (term) compare(a, b Value, _ Rules) (Result, Detail) {
	switch {
	case a.Term == "" || b.Term == "":
		return Result{Relation: Unknown}, Detail{Cause: CauseEmpty}
	case a.Term == b.Term:
		return Result{Relation: Same, Similarity: 1}, Detail{Cause: CauseExact}
	}
	return Result{Relation: Different}, Detail{Cause: CauseNoMatch}
}

// key is the term key at both levels: a term has no looser form.
func (term) key(v Value, _ Relation, _ Rules) (string, bool) {
	return v.Term, v.Term != ""
}
