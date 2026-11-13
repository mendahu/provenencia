package conclusionsummaries

import (
	"strings"

	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/eventtitle"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecodec"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Term is a term value of a summary: the displayed (kept rank-1) term of a
// Property. Key names a product term only when Origin is the product's.
type Term struct {
	ID     []byte
	Key    string
	Origin string // the term's origin; Key is a product type's only when it is the product's
	Label  string
}

// ProductKey is the key of t when t is the product's own term, else "": a key
// means the product's only within the product's origin. A nil Term has none.
func (t *Term) ProductKey() string {
	if t == nil || t.Origin != vocabkey.OriginProvenencia {
		return ""
	}
	return t.Key
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
	// Participants are every person on the event through any role, by id and
	// roles, in the order their participations are linked. The roles are not
	// finished: a reader that shows them labels them (TermLabeler).
	Participants []EventParticipant
	Places       []ChainedPlace
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

// EventParticipant is one entity on an event through its participations and the
// distinct roles those carry, in link order.
type EventParticipant struct {
	ID    []byte
	Roles []Term
}

// rolesByNeighbor is the one collapse rule for an event's participants: links
// in input order; neighbors are de-duplicated in first-seen order; each link
// adds its role term to its neighbor's roles unless that term is already
// there. A link with no term, or whose term label does not know, adds no role.
// The ids are copies, so the result does not alias the links.
func rolesByNeighbor(links []graphcache.Link, label func(id []byte, key string) (Term, bool)) []EventParticipant {
	var out []EventParticipant
	index := map[string]int{}
	seen := map[string]map[string]struct{}{}
	for _, l := range links {
		i, ok := index[string(l.Neighbor)]
		if !ok {
			i = len(out)
			index[string(l.Neighbor)] = i
			out = append(out, EventParticipant{ID: append([]byte(nil), l.Neighbor...)})
		}
		if len(l.TermID) == 0 {
			continue
		}
		if _, have := seen[string(l.Neighbor)][string(l.TermID)]; have {
			continue
		}
		if t, ok := label(l.TermID, l.TermKey); ok {
			if seen[string(l.Neighbor)] == nil {
				seen[string(l.Neighbor)] = map[string]struct{}{}
			}
			seen[string(l.Neighbor)][string(l.TermID)] = struct{}{}
			out[i].Roles = append(out[i].Roles, t)
		}
	}
	return out
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
