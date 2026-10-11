// Package hops is the vocabulary of named graph steps. A Hop names one bridge
// from the connectrules registry and the two endpoint Properties it crosses,
// checked against that registry when the Hop is made. The SQL walks and the
// in-memory graph both follow these hops, so a parent place is declared once.
package hops

import (
	"fmt"
	"strings"

	"github.com/mendahu/provenencia/core/connectrules"
)

// TermFilter keeps an association only when its kept rank-1 value of the
// bridge's disambiguation Property is the given term (role = subject).
type TermFilter struct {
	PropertyKey string
	TermKey     string
}

// Hop crosses one bridge from one endpoint Property to the other.
type Hop struct {
	origin, bridge string
	from, to       string
	filter         *TermFilter
}

// NewHop checks a hop against the registered bridge: both Properties must be
// its endpoints, and a filter must be on its disambiguation Property.
func NewHop(bridgeTypeKey, fromProperty, toProperty string, filter *TermFilter) (Hop, error) {
	b, ok := connectrules.LookupBridge(bridgeTypeKey)
	if !ok {
		return Hop{}, fmt.Errorf("hops: no bridge %q", bridgeTypeKey)
	}
	if fromProperty == toProperty || !isEndpoint(b, fromProperty) || !isEndpoint(b, toProperty) {
		return Hop{}, fmt.Errorf("hops: %s does not join %s to %s", bridgeTypeKey, fromProperty, toProperty)
	}
	if filter != nil && (b.Disambiguation != filter.PropertyKey || filter.TermKey == "") {
		return Hop{}, fmt.Errorf("hops: %s does not disambiguate by %s", bridgeTypeKey, filter.PropertyKey)
	}
	origin := strings.TrimSpace(b.Origin)
	if origin == "" {
		origin = connectrules.OriginProvenencia
	}
	h := Hop{origin: origin, bridge: b.BridgeTypeKey, from: fromProperty, to: toProperty}
	if filter != nil {
		f := *filter
		h.filter = &f
	}
	return h, nil
}

// The product hops. A subject-role participation joins a person to an event
// (role stays off the association's identity; walks read the role). A
// location joins an event to a place. Place relationships join two places
// (part_of for containment chains; succeeded_by for lineage, never chains).
var (
	EventsOfSubject = MustHop("participation", "person", "event",
		&TermFilter{PropertyKey: connectrules.DisambiguationRole, TermKey: "subject"})
	SubjectsOfEvent = EventsOfSubject.Reverse()
	PlacesOfEvent   = MustHop("location", "event", "place", nil)
	EventsAtPlace   = PlacesOfEvent.Reverse()
	// ParentsOfPlace: from = part, to = whole (part_of).
	ParentsOfPlace = MustHop("place_relationship", "from", "to",
		&TermFilter{PropertyKey: connectrules.DisambiguationPlaceRelationshipType, TermKey: "part_of"})
	PartsOfPlace = ParentsOfPlace.Reverse()
	// SuccessorsOfPlace: from = predecessor, to = successor (succeeded_by).
	SuccessorsOfPlace = MustHop("place_relationship", "from", "to",
		&TermFilter{PropertyKey: connectrules.DisambiguationPlaceRelationshipType, TermKey: "succeeded_by"})
	PredecessorsOfPlace = SuccessorsOfPlace.Reverse()
)

// MustHop is NewHop for package-level hops over product bridges.
func MustHop(bridgeTypeKey, fromProperty, toProperty string, filter *TermFilter) Hop {
	h, err := NewHop(bridgeTypeKey, fromProperty, toProperty, filter)
	if err != nil {
		panic(err)
	}
	return h
}

// Reverse is the same hop walked the other way.
func (h Hop) Reverse() Hop {
	h.from, h.to = h.to, h.from
	return h
}

// Origin is the bridge vocabulary origin the SQL walk matches.
func (h Hop) Origin() string { return h.origin }

// Bridge is the bridge type key.
func (h Hop) Bridge() string { return h.bridge }

// From is the endpoint Property this hop starts on.
func (h Hop) From() string { return h.from }

// To is the endpoint Property this hop arrives on.
func (h Hop) To() string { return h.to }

// Filter is the disambiguation term, when the hop has one.
func (h Hop) Filter() (propertyKey, termKey string, ok bool) {
	if h.filter == nil {
		return "", "", false
	}
	return h.filter.PropertyKey, h.filter.TermKey, true
}

func isEndpoint(b connectrules.Bridge, propertyKey string) bool {
	for _, e := range b.Endpoints {
		if e.PropertyKey == propertyKey {
			return true
		}
	}
	return false
}
