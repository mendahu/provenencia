// Package vocabkey holds the product vocabulary keys code reads by name:
// vocabulary origins, the primary Subject kinds, and the seeded Property keys
// on them. Pure: the seed and the queries live under core/database.
//
// A kind is the key of a primary Subject type. IsPrimaryKind takes a key
// alone; IsPrimaryType takes a catalog type's key and origin.
package vocabkey

import "strings"

// Vocabulary origins. A plugin origin is "plugin:<id>"; see IsPluginOrigin.
const (
	OriginProvenencia  = "provenencia"
	OriginUser         = "user"
	originPluginPrefix = "plugin:"
)

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
