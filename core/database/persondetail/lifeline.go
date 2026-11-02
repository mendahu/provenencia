package persondetail

import (
	"bytes"

	"github.com/mendahu/provenencia/core/vocabkey"
)

// lifeline names the event types that may sit outside the years of a life. It
// places the birth and death markers of a timeline when the records name no
// event for them (lifeMarkers); it judges no event, and flags none.
var lifeline = struct{ afterDeath, beforeBirth []vocabkey.EventType }{
	afterDeath: []vocabkey.EventType{vocabkey.EventBurial, vocabkey.EventFuneral},
}

// LifeMarker is where a timeline puts a birth or a death. Event is the
// recorded event, nil when the records name none (a placeholder). Beside is set
// whenever the marker has no span of its own, to the dated event it sits next
// to; no date is invented.
type LifeMarker struct {
	Event  []byte
	Beside []byte
}

// lifeMarkers places the birth and death markers on events, which are in event
// date order. birth and death are the recorded events (nil for none). A marker
// whose event has a span needs no Beside. Otherwise a birth sits beside the
// first row with a span whose type may come after birth, and a death beside
// the last row with a span whose type may come before death; Beside is nil when
// no row qualifies. types gives an event's type.
func lifeMarkers(events []PersonEvent, types func(event []byte) vocabkey.EventType, birth, death []byte) (Birth, Death LifeMarker) {
	hasSpan := func(id []byte) bool {
		for _, e := range events {
			if bytes.Equal(e.Event, id) {
				return e.Span != nil
			}
		}
		return false
	}
	skip := func(list []vocabkey.EventType, id []byte) bool {
		t := types(id)
		for _, x := range list {
			if x == t {
				return true
			}
		}
		return false
	}
	Birth, Death = LifeMarker{Event: birth}, LifeMarker{Event: death}
	if birth == nil || !hasSpan(birth) {
		for _, e := range events {
			if e.Span != nil && !skip(lifeline.beforeBirth, e.Event) {
				Birth.Beside = e.Event
				break
			}
		}
	}
	if death == nil || !hasSpan(death) {
		for i := len(events) - 1; i >= 0; i-- {
			e := events[i]
			if e.Span != nil && !skip(lifeline.afterDeath, e.Event) {
				Death.Beside = e.Event
				break
			}
		}
	}
	return Birth, Death
}
