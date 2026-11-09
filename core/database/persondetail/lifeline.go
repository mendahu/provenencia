package persondetail

import (
	"bytes"

	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// lifelineTypes names the event-and-role pairs that may sit outside the years
// of a life: a person's own burial or funeral comes after the death. It places
// the birth and death markers of a timeline when the records name no event for
// them (markers); it judges no event, and flags none.
type lifelineTypes struct{ afterDeath, beforeBirth []subjectcompare.EventRole }

var lifeline = lifelineTypes{
	afterDeath: []subjectcompare.EventRole{
		{EventType: vocabkey.EventBurial, Role: vocabkey.RoleSubject},
		{EventType: vocabkey.EventFuneral, Role: vocabkey.RoleSubject},
	},
}

// MarkerSide is which side of its row a marker sits on.
type MarkerSide int

const (
	// SideAfter puts the marker just after its row.
	SideAfter MarkerSide = iota + 1
	// SideBefore puts the marker just before its row.
	SideBefore
)

// MarkerPlace is the one dated row a marker sits next to, and on which side.
type MarkerPlace struct {
	Row  []byte
	Side MarkerSide
}

// LifeMarker is where a timeline puts a birth or a death. Event is the
// recorded event, nil when the records name none (a placeholder). Place is set
// whenever the marker has no span of its own and a row is dated, to the dated
// event it sits next to; no date is invented. It is nil when the marker's
// event has a span, or when no row is dated.
type LifeMarker struct {
	Event []byte
	Place *MarkerPlace
}

// matches reports whether row, of type t, is one of list: the type and one of
// the person's roles on the row make an entry. A role matches when it is the
// product's own term with that key.
func (l lifelineTypes) matches(list []subjectcompare.EventRole, row PersonEvent, t vocabkey.EventType) bool {
	for _, x := range list {
		if x.EventType != t {
			continue
		}
		for i := range row.Roles {
			if row.Roles[i].ProductKey() == string(x.Role) {
				return true
			}
		}
	}
	return false
}

// markers places the birth and death markers on events, which are in event
// date order. birth and death are the recorded events (nil for none). A marker
// whose event has a span needs no Place. Otherwise a birth sits after the last
// dated row that may come before birth when there is one, else before the first
// dated row; a death sits before the first dated row that may come after death
// when there is one, else after the last dated row. Place is nil when no row is
// dated. types gives an event's type.
func (l lifelineTypes) markers(events []PersonEvent, types func(event []byte) vocabkey.EventType, birth, death []byte) (Birth, Death LifeMarker) {
	hasSpan := func(id []byte) bool {
		for _, e := range events {
			if bytes.Equal(e.Event, id) {
				return e.Span != nil
			}
		}
		return false
	}
	Birth, Death = LifeMarker{Event: birth}, LifeMarker{Event: death}
	if birth == nil || !hasSpan(birth) {
		var first, lastBefore []byte
		for _, e := range events {
			if e.Span == nil {
				continue
			}
			if first == nil {
				first = e.Event
			}
			if l.matches(l.beforeBirth, e, types(e.Event)) {
				lastBefore = e.Event
			}
		}
		switch {
		case lastBefore != nil:
			Birth.Place = &MarkerPlace{Row: lastBefore, Side: SideAfter}
		case first != nil:
			Birth.Place = &MarkerPlace{Row: first, Side: SideBefore}
		}
	}
	if death == nil || !hasSpan(death) {
		var last, firstAfter []byte
		for _, e := range events {
			if e.Span == nil {
				continue
			}
			last = e.Event
			if firstAfter == nil && l.matches(l.afterDeath, e, types(e.Event)) {
				firstAfter = e.Event
			}
		}
		switch {
		case firstAfter != nil:
			Death.Place = &MarkerPlace{Row: firstAfter, Side: SideBefore}
		case last != nil:
			Death.Place = &MarkerPlace{Row: last, Side: SideAfter}
		}
	}
	return Birth, Death
}
