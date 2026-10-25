// Package vocabkey holds the product vocabulary keys code reads by name:
// vocabulary origins, the primary Subject kinds, the seeded Property keys on
// them, the seeded term keys it reads, and the derived Property keys matching
// builds. Pure: the seed and the queries live under core/database.
//
// A kind is the key of a primary Subject type. IsPrimaryKind takes a key
// alone; IsPrimaryType takes a catalog type's key and origin.
package vocabkey

import (
	"slices"
	"strings"
)

// Vocabulary origins. A plugin origin is "plugin:<id>"; see IsPluginOrigin.
const (
	OriginProvenencia  = "provenencia"
	OriginUser         = "user"
	originPluginPrefix = "plugin:"
)

// SubjectType is a Subject type key, typed for registry maps.
type SubjectType string

// PropertyKey is a Property key, typed for registry maps.
type PropertyKey string

// CompareByOrder ranks two Property keys against a kind's binding order: a key
// in order before one that is not, bound keys by their index in order, the rest
// by key. It is the one ranking for the conclusion page and the evidence
// sheet.
func CompareByOrder(order []PropertyKey, a, b PropertyKey) int {
	ia, ib := slices.Index(order, a), slices.Index(order, b)
	switch {
	case ia >= 0 && ib >= 0:
		if ia != ib {
			return ia - ib
		}
	case ia >= 0:
		return -1
	case ib >= 0:
		return 1
	}
	return strings.Compare(string(a), string(b))
}

// Primary Subject kinds.
const (
	Person = "person"
	Event  = "event"
	Place  = "place"
)

// Seeded Property keys that matching and search read by name.
const (
	PropertyName       = "name"
	PropertySexAtBirth = "sex_at_birth"
	PropertyGender     = "gender"
	PropertyEventType  = "event_type"
	PropertyDate       = "date"
	PropertyStartDate  = "start_date"
	PropertyEndDate    = "end_date"
	PropertyToponym    = "toponym"
	PropertyLikeness   = "likeness"

	PropertyPerson    = "person"
	PropertyEvent     = "event"
	PropertyPlace     = "place"
	PropertyRelatedTo = "related_to"
	PropertyFrom      = "from"
	PropertyTo        = "to"
)

// Bridge types: the seeded Subject types that join two primary Subjects.
const (
	BridgeParticipation     = "participation"
	BridgeRelationship      = "relationship"
	BridgeLocation          = "location"
	BridgePlaceRelationship = "place_relationship"
)

// EventType is an event_type term key, typed for registry rules.
type EventType string

// Role is a participation role term key, typed for registry rules.
type Role string

// Seeded term keys matching reads by name.
const (
	EventBirth    EventType = "birth"
	EventDeath    EventType = "death"
	EventBurial   EventType = "burial"
	EventMarriage EventType = "marriage"
	RoleSubject   Role      = "subject"
	RoleFather    Role      = "father"
	RoleMother    Role      = "mother"
	RoleSpouse    Role      = "spouse"

	SexMale     = "male"   // sex_at_birth
	SexFemale   = "female" // sex_at_birth
	GenderMan   = "man"    // gender
	GenderWoman = "woman"  // gender
)

// BridgeTerm is a bridge's disambiguation term key: a participation role, a
// relationship type or a place relationship type, typed for registry rules.
type BridgeTerm string

const (
	// TermPartOf is the place relationship type of a place within a larger one.
	TermPartOf BridgeTerm = "part_of"

	// Relationship types of a person to related_to.
	TermSpouse      BridgeTerm = "spouse"
	TermParent      BridgeTerm = "parent"
	TermGrandparent BridgeTerm = "grandparent"
)

// BridgeKey is one bridge read from one end: its bridge type, its term in
// canonical direction, the endpoint Property of the end it is read from (empty
// for a symmetric link between two Subjects of one type), and the type of the
// event on a participation.
type BridgeKey struct {
	Type      SubjectType
	Term      BridgeTerm
	End       PropertyKey
	EventType EventType
}

// String joins the non-empty fields with a middle dot, in field order.
func (k BridgeKey) String() string {
	parts := make([]string, 0, 4)
	for _, f := range []string{string(k.Type), string(k.Term), string(k.End), string(k.EventType)} {
		if f != "" {
			parts = append(parts, f)
		}
	}
	return strings.Join(parts, "·")
}

// Derived Property keys the profile builder adds (never seeded).
const (
	PropertyLifespan     = "lifespan"
	PropertyLifespanSpan = "lifespan_span"
)

// PrimaryKinds returns a new slice of the primary kinds on each call.
func PrimaryKinds() []string {
	return []string{Person, Event, Place}
}

// IsPrimaryKind reports whether kind is a primary kind. Use it for keys whose
// origin is already known to be provenencia.
func IsPrimaryKind(kind string) bool {
	switch kind {
	case Person, Event, Place:
		return true
	}
	return false
}

// IsPrimaryType reports whether a catalog type row is a primary kind.
func IsPrimaryType(typeKey, origin string) bool {
	return origin == OriginProvenencia && IsPrimaryKind(typeKey)
}

// IsPluginOrigin reports whether origin is "plugin:<id>" with a non-empty id.
func IsPluginOrigin(origin string) bool {
	return strings.HasPrefix(origin, originPluginPrefix) && len(origin) > len(originPluginPrefix)
}

// ValidOrigin reports whether origin is provenencia, user, or a plugin origin.
func ValidOrigin(origin string) bool {
	return origin == OriginProvenencia || origin == OriginUser || IsPluginOrigin(origin)
}
