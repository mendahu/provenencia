package valuecompare

// boolean compares yes/no facts. A yes/no has no nearness: the values are the
// same or they disagree, and false is a recorded fact, not an absent one.
type boolean struct{}

func (boolean) compare(a, b Value, _ Rules) (Result, Detail) {
	if a.Boolean == b.Boolean {
		return Result{Relation: Same, Similarity: 1}, Detail{Cause: CauseExact}
	}
	return Result{Relation: Different}, Detail{Cause: CauseNoMatch}
}

// key is "true" or "false" at both levels; a boolean always has one.
func (boolean) key(v Value, _ Relation, _ Rules) (string, bool) {
	if v.Boolean {
		return "true", true
	}
	return "false", true
}
