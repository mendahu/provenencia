package segmentcompare

import (
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/vocabkey"
)

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

// LinkOf is a bridge as recorded, read from its first endpoint's end or its
// second. It is the one builder of a link: every reader of a bridge goes
// through it, or through Keys, so their keys cannot drift. b has two
// endpoints, and only a participation carries an event type.
func LinkOf(b connectrules.Bridge, term vocabkey.BridgeTerm, eventType vocabkey.EventType, firstEnd bool) Link {
	first, second := b.Endpoints[0], b.Endpoints[1]
	l := Link{
		Type:     vocabkey.SubjectType(b.BridgeTypeKey),
		Term:     term,
		End:      vocabkey.PropertyKey(first.PropertyKey),
		Other:    vocabkey.PropertyKey(second.PropertyKey),
		SameType: first.TypeKey == second.TypeKey,
	}
	if !firstEnd {
		l.End, l.Other = l.Other, l.End
	}
	if b.BridgeTypeKey == vocabkey.BridgeParticipation {
		l.EventType = eventType
	}
	return l
}

// Keys are a bridge's keys read from its first endpoint's end and from its
// second.
func Keys(b connectrules.Bridge, term vocabkey.BridgeTerm, eventType vocabkey.EventType, terms Terms) (first, second vocabkey.BridgeKey) {
	return Key(LinkOf(b, term, eventType, true), terms), Key(LinkOf(b, term, eventType, false), terms)
}

// EventEndpoint says which end of a bridge is an event: first when it is the
// bridge's first endpoint, and ok false when neither is. It is the one rule
// for where a link reads its event type.
func EventEndpoint(b connectrules.Bridge) (first, ok bool) {
	switch {
	case b.Endpoints[0].TypeKey == vocabkey.Event:
		return true, true
	case b.Endpoints[1].TypeKey == vocabkey.Event:
		return false, true
	}
	return false, false
}

// Corresponds reports whether two links mean the same thing read from the same
// end. It is the one correspondence rule; the walk and Compare call it.
func Corresponds(a, b vocabkey.BridgeKey) bool {
	return a == b
}
