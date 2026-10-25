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

// ReadFrom reads the link a key stands for from the neighbor's side: the
// endpoint the neighbor is at, the term that endpoint carries, and whether that
// term names the neighbor ("the neighbor is the parent") rather than the row.
// k is the key from the row's end. first and second are the bridge type's
// endpoint properties, in the order the connect rules list them.
//
// k.End is the row's end in the key's reading, which is not always the end the
// row is stored at: Key swaps a term with its inverse when the inverse sorts
// first, and the ends with it. A symmetric same-type link has no End and reads
// the same from both sides. Otherwise the neighbor is at the other endpoint.
// At the first endpoint the key's term names the neighbor. At the second it
// names the row, so the neighbor's term is the inverse when the catalog has
// one; with none, the term still names the row and the caller says so.
func ReadFrom(k vocabkey.BridgeKey, first, second vocabkey.PropertyKey, terms Terms) (neighborEnd vocabkey.PropertyKey, term vocabkey.BridgeTerm, namesNeighbor bool) {
	if k.End == "" {
		return "", k.Term, true
	}
	if k.End != first {
		return first, k.Term, true
	}
	if inv := terms[k.Type][k.Term].Inverse; inv != "" {
		return second, inv, true
	}
	return second, k.Term, false
}
