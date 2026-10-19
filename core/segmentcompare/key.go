package segmentcompare

import "github.com/mendahu/provenencia/core/vocabkey"

// Link is a bridge as recorded, read from the end whose endpoint Property is
// End. Other is the Property of the far end. SameType is true when both
// endpoints hold the same Subject type.
type Link struct {
	Type       vocabkey.SubjectType
	Term       vocabkey.BridgeTerm
	End, Other vocabkey.PropertyKey
	EventType  vocabkey.EventType
	SameType   bool
}

// Term is the catalog's facts about one bridge term: whether it reads one way
// only, and the term it reads as from the other end.
type Term struct {
	Directed bool
	Inverse  vocabkey.BridgeTerm
}

// Terms are the catalog's term facts per bridge type. A type or term with no
// entry is the zero Term.
type Terms map[vocabkey.SubjectType]map[vocabkey.BridgeTerm]Term

// Key reads a link as one typed key. A term and its inverse read as the one
// that sorts first, with the ends swapped, so "Mary parent of John" and "John
// child of Mary" give the same key from the same end. A symmetric link
// between two Subjects of one type has no End. Type and EventType are copied.
func Key(l Link, terms Terms) vocabkey.BridgeKey {
	term, end, other := l.Term, l.End, l.Other
	if inv := terms[l.Type][term].Inverse; inv != "" && inv < term {
		term, end, other = inv, other, end
	}
	if l.SameType && !terms[l.Type][term].Directed {
		end = ""
	}
	return vocabkey.BridgeKey{Type: l.Type, Term: term, End: end, EventType: l.EventType}
}

// Corresponds reports whether two links mean the same thing read from the same
// end. It is the one correspondence rule; the walk and Compare call it.
func Corresponds(a, b vocabkey.BridgeKey) bool {
	return a == b
}
