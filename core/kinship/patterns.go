package kinship

import "github.com/mendahu/provenencia/core/vocabkey"

// eventPattern is one fact an event implies between two of its participants:
// the event type, the roles at each end of the implied link, and the
// relationship term it implies. The implied link's A is the relationship's
// person end.
type eventPattern struct {
	code  Pattern
	event vocabkey.EventType
	a, b  []vocabkey.Role
	term  vocabkey.BridgeTerm
}

// eventPatterns are the product's event patterns, in output order. A catalog's
// own roles and event types imply nothing.
var eventPatterns = []eventPattern{
	{PatternMarriageSpouses, vocabkey.EventMarriage, []vocabkey.Role{vocabkey.RoleSubject}, []vocabkey.Role{vocabkey.RoleSubject}, vocabkey.TermSpouse},
	{PatternBirthParents, vocabkey.EventBirth, []vocabkey.Role{vocabkey.RoleFather, vocabkey.RoleMother}, []vocabkey.Role{vocabkey.RoleSubject}, vocabkey.TermParent},
}

// PivotKind is the Subject type a pattern's pivot is: the event for an event
// pattern, the parent between for a chain, the shared parent for a sibling, and
// "" for a code that is no pattern.
func (p Pattern) PivotKind() vocabkey.SubjectType {
	for _, ep := range eventPatterns {
		if ep.code == p {
			return vocabkey.Event
		}
	}
	if p == PatternParentOfParent || p == PatternSharedParent {
		return vocabkey.Person
	}
	return ""
}

// Patterns are the product's pattern codes: each event pattern in order, then
// the parent-of-parent chain. Page-only codes (PatternSharedParent) are not in
// it.
func Patterns() []Pattern {
	out := make([]Pattern, 0, len(eventPatterns)+1)
	for _, ep := range eventPatterns {
		out = append(out, ep.code)
	}
	return append(out, PatternParentOfParent)
}
