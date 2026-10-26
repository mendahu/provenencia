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
	{PatternMarriageSpouses, vocabkey.EventMarriage, []vocabkey.Role{vocabkey.RoleSpouse}, []vocabkey.Role{vocabkey.RoleSpouse}, vocabkey.TermSpouse},
	{PatternBirthParents, vocabkey.EventBirth, []vocabkey.Role{vocabkey.RoleFather, vocabkey.RoleMother}, []vocabkey.Role{vocabkey.RoleSubject}, vocabkey.TermParent},
}
