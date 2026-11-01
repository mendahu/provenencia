package conclusionsummaries

import (
	"strings"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecodec"
)

// Term is a term value of a summary: the displayed (kept rank-1) term of a
// Property. Key names a product term only when Origin is the product's.
type Term struct {
	ID     []byte
	Key    string
	Origin string // the term's origin; Key is a product type's only when it is the product's
	Label  string
}

// EventSummary is one Event as a row. A rank-1 date wins over a start/end
// span; the span's counts are still reported. Subjects are the subject-role
// persons, participation ref then person ref. Places are every location's
// kept names with hierarchical parents at the event's date.
type EventSummary struct {
	Entity         canonicalentities.Entity
	EventName      string
	EventNameCount int
	// Type is the event_type term, with how many values are displayed.
	Type           TermFacts
	Date           *datevalue.Value
	DateCount      int
	StartDate      *datevalue.Value
	StartDateCount int
	EndDate        *datevalue.Value
	EndDateCount   int
	Subjects       []EventSubject
	Places         []ChainedPlace
	// PictureImageFileID is the image of the picture the Event leads with;
	// nil when none. Preferred when set and in the gallery, else earliest.
	PictureImageFileID []byte
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

// eventTitle chooses the summary's title from its own parts: subjects in
// summary order, and the first named place.
func eventTitle(h EventSummary) eventtitle.Plan {
	parts := eventtitle.Parts{
		RecordedName: h.EventName,
		Label:        h.Entity.Label,
		Ref:          h.Entity.Ref,
		Place:        FirstPlaceName(h.Places),
	}
	if t := h.Type.Term; t != nil {
		parts.TypeKey, parts.TypeOrigin, parts.TypeLabel = t.Key, t.Origin, t.Label
	}
	for _, s := range h.Subjects {
		parts.Subjects = append(parts.Subjects, eventtitle.Subject{Name: s.Name})
	}
	return eventtitle.Choose(parts)
}

// FirstPlaceName is the first kept name of the first named Place.
func FirstPlaceName(places []ChainedPlace) string {
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
