package conclusionheaders

import (
	"strings"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// EventType is the displayed (kept rank-1) event_type term.
type EventType struct {
	ID     []byte
	Key    string
	Origin string // the term's origin; Key is a product type's only when it is the product's
	Label  string
}

// EventHeader is one Event as a row. A rank-1 date wins over a start/end
// span; the span's counts are still reported. Subjects are the subject-role
// persons, participation ref then person ref. Places are every location's
// kept names with hierarchical parents at the event's date.
type EventHeader struct {
	Entity         canonicalentities.Entity
	EventName      string
	EventNameCount int
	EventType      *EventType
	EventTypeCount int
	Date           *datevalue.Value
	DateCount      int
	StartDate      *datevalue.Value
	StartDateCount int
	EndDate        *datevalue.Value
	EndDateCount   int
	Subjects       []EventSubject
	Places         []HeaderPlace
	// Title is the naming-matrix rule and parts, chosen from the above.
	Title eventtitle.Plan
	// Updated is the handle's latest canonical_entity revision and its time.
	// Filled by List only; zero means none.
	Updated effects.Activity
}

// EventSubject is one person on an event through a subject-role participation.
type EventSubject struct {
	Entity         canonicalentities.Entity
	Name           *namevalue.Value
	NameValueCount int
}

// eventTitle chooses the header's title from its own parts: subjects in
// header order, and the first named place.
func eventTitle(h EventHeader) eventtitle.Plan {
	parts := eventtitle.Parts{
		RecordedName: h.EventName,
		Label:        h.Entity.Label,
		Ref:          h.Entity.Ref,
		Place:        FirstPlaceName(h.Places),
	}
	if h.EventType != nil {
		parts.TypeKey, parts.TypeOrigin, parts.TypeLabel = h.EventType.Key, h.EventType.Origin, h.EventType.Label
	}
	for _, s := range h.Subjects {
		parts.Subjects = append(parts.Subjects, eventtitle.Subject{Name: s.Name})
	}
	return eventtitle.Choose(parts)
}

// FirstPlaceName is the first kept name of the first named Place.
func FirstPlaceName(places []HeaderPlace) string {
	for _, p := range places {
		for _, n := range p.Names {
			if n = strings.TrimSpace(n); n != "" {
				return n
			}
		}
	}
	return ""
}

func unmarshalDate(b []byte) (*datevalue.Value, error) {
	if len(b) == 0 {
		return nil, nil
	}
	v, err := valuecodec.UnmarshalDate(b)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
