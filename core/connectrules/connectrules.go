// Package connectrules is the Interpretation connect matrix.
//
// This is the only product table of “this property on this bridge type is an
// endpoint or a disambiguation.” subjectvocab Install derives bridge bindings
// from All(); Impact, edge-lock, and Connect loop All() — they do not name
// role / relationship_type. Future plugins append via Register; they do not
// patch policy functions.
package connectrules

import (
	"strings"
	"sync"
)

const (
	DisambiguationNone                  = "none"
	DisambiguationRole                  = "role"
	DisambiguationRelationshipType      = "relationship_type"
	DisambiguationPlaceRelationshipType = "place_relationship_type"

	OriginProvenencia = "provenencia"
)

// Endpoint binds a Property key on a bridge to the node type it points at.
type Endpoint struct {
	PropertyKey string
	TypeKey     string
}

// Bridge is one connectable bridge type authored once (directions in Pairs).
type Bridge struct {
	Origin         string
	BridgeTypeKey  string
	Endpoints      []Endpoint
	Disambiguation string // property key; "" or "none" = none
	Pairs          [][2]string
}

// Rule is one allowed (or refused) connect endpoint pair after expansion.
type Rule struct {
	Origin         string
	FromTypeKey    string
	ToTypeKey      string
	BridgeTypeKey  string
	Endpoints      []Endpoint
	Disambiguation string
	Refuse         bool
}

// Product bridges and refusals. Omitted pairs refuse by default.
var productBridges = []Bridge{
	{
		Origin: OriginProvenencia, BridgeTypeKey: "participation",
		Endpoints: []Endpoint{
			{PropertyKey: "person", TypeKey: "person"},
			{PropertyKey: "event", TypeKey: "event"},
		},
		Disambiguation: DisambiguationRole,
		Pairs:          [][2]string{{"person", "event"}, {"event", "person"}},
	},
	{
		Origin: OriginProvenencia, BridgeTypeKey: "relationship",
		Endpoints: []Endpoint{
			{PropertyKey: "person", TypeKey: "person"},
			{PropertyKey: "related_to", TypeKey: "person"},
		},
		Disambiguation: DisambiguationRelationshipType,
		Pairs:          [][2]string{{"person", "person"}},
	},
	{
		Origin: OriginProvenencia, BridgeTypeKey: "location",
		Endpoints: []Endpoint{
			{PropertyKey: "event", TypeKey: "event"},
			{PropertyKey: "place", TypeKey: "place"},
		},
		Disambiguation: DisambiguationNone,
		Pairs:          [][2]string{{"event", "place"}, {"place", "event"}},
	},
	{
		Origin: OriginProvenencia, BridgeTypeKey: "place_relationship",
		Endpoints: []Endpoint{
			{PropertyKey: "from", TypeKey: "place"},
			{PropertyKey: "to", TypeKey: "place"},
		},
		Disambiguation: DisambiguationPlaceRelationshipType,
		Pairs:          [][2]string{{"place", "place"}},
	},
}

var productRefusals = [][2]string{
	{"person", "place"},
	{"place", "person"},
	{"event", "event"},
}

var (
	mu       sync.RWMutex
	bridges  []Bridge
	refusals [][2]string
	rules    []Rule
)

func init() {
	loadProduct()
}

func loadProduct() {
	bridges = append([]Bridge(nil), productBridges...)
	refusals = append([][2]string(nil), productRefusals...)
	rebuild()
}

func rebuild() {
	out := make([]Rule, 0, len(bridges)*2+len(refusals))
	for _, b := range bridges {
		origin := strings.TrimSpace(b.Origin)
		if origin == "" {
			origin = OriginProvenencia
		}
		disamb := b.Disambiguation
		if disamb == "" {
			disamb = DisambiguationNone
		}
		eps := append([]Endpoint(nil), b.Endpoints...)
		for _, p := range b.Pairs {
			out = append(out, Rule{
				Origin:         origin,
				FromTypeKey:    p[0],
				ToTypeKey:      p[1],
				BridgeTypeKey:  b.BridgeTypeKey,
				Endpoints:      eps,
				Disambiguation: disamb,
			})
		}
	}
	for _, p := range refusals {
		out = append(out, Rule{
			FromTypeKey: p[0],
			ToTypeKey:   p[1],
			Refuse:      true,
		})
	}
	rules = out
}

// All returns the expanded directed connect matrix. Consumers loop this.
func All() []Rule {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Rule, len(rules))
	copy(out, rules)
	return out
}

// Bridges returns registered bridge definitions (not refusals).
func Bridges() []Bridge {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Bridge, len(bridges))
	copy(out, bridges)
	return out
}

// Register appends bridges and rebuilds the directed table.
// Future plugins call this; they do not patch Impact or edge-lock.
func Register(extra ...Bridge) {
	mu.Lock()
	defer mu.Unlock()
	bridges = append(bridges, extra...)
	rebuild()
}

// RegisterRefusals appends explicit refuse pairs and rebuilds.
func RegisterRefusals(pairs ...[2]string) {
	mu.Lock()
	defer mu.Unlock()
	refusals = append(refusals, pairs...)
	rebuild()
}

// ResetForTest restores the product-only matrix. Call from tests only.
func ResetForTest() {
	mu.Lock()
	defer mu.Unlock()
	loadProduct()
}

// EdgePropertyKeys is the endpoint property list on a rule (FFI / UI order).
func (r Rule) EdgePropertyKeys() []string {
	out := make([]string, 0, len(r.Endpoints))
	for _, e := range r.Endpoints {
		out = append(out, e.PropertyKey)
	}
	return out
}

// HasDisambiguation reports whether the rule names a disambiguation property.
func HasDisambiguation(key string) bool {
	key = strings.TrimSpace(key)
	return key != "" && key != DisambiguationNone
}

// Edge returns the endpoint type a bridge edge property binds, if any
// non-refused registered bridge defines the pair.
func Edge(bridgeTypeKey, propertyKey string) (endpointTypeKey string, ok bool) {
	bridgeTypeKey = strings.TrimSpace(bridgeTypeKey)
	propertyKey = strings.TrimSpace(propertyKey)
	if bridgeTypeKey == "" || propertyKey == "" {
		return "", false
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, r := range rules {
		if r.Refuse || r.BridgeTypeKey != bridgeTypeKey {
			continue
		}
		for _, e := range r.Endpoints {
			if e.PropertyKey == propertyKey {
				return e.TypeKey, true
			}
		}
	}
	return "", false
}

// IsEdgePair reports whether propertyKey is a registered edge on bridge typeKey
// for the given property origin.
func IsEdgePair(typeKey, propKey, origin string) bool {
	typeKey = strings.TrimSpace(typeKey)
	propKey = strings.TrimSpace(propKey)
	origin = strings.TrimSpace(origin)
	if typeKey == "" || propKey == "" || origin == "" {
		return false
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, r := range rules {
		if r.Refuse || r.BridgeTypeKey != typeKey || r.Origin != origin {
			continue
		}
		for _, e := range r.Endpoints {
			if e.PropertyKey == propKey {
				return true
			}
		}
	}
	return false
}

// IsDisambiguation reports whether propKey is a registered disambiguation on
// typeKey for the given property origin.
func IsDisambiguation(typeKey, propKey, origin string) bool {
	typeKey = strings.TrimSpace(typeKey)
	propKey = strings.TrimSpace(propKey)
	origin = strings.TrimSpace(origin)
	if typeKey == "" || propKey == "" || origin == "" {
		return false
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, r := range rules {
		if r.Refuse || r.BridgeTypeKey != typeKey || r.Origin != origin {
			continue
		}
		if !HasDisambiguation(r.Disambiguation) {
			continue
		}
		if r.Disambiguation == propKey {
			return true
		}
	}
	return false
}

// IsConnectionFacet reports a registered-origin edge or disambiguation
// Observation on a bridge type.
func IsConnectionFacet(typeKey, propKey, origin string) bool {
	return IsEdgePair(typeKey, propKey, origin) || IsDisambiguation(typeKey, propKey, origin)
}

// Binding is one type↔property Install row derived from a bridge.
type Binding struct {
	Origin               string
	TypeKey, PropertyKey string
	SortOrder            int
	Locked               bool
}

// LookupBridge returns the registered bridge for a type key, or false.
// Refusals are not bridges. The filer keys whatever this returns: ends plus a
// non-role disambiguation (relationship_type, place_relationship_type).
func LookupBridge(typeKey string) (Bridge, bool) {
	typeKey = strings.TrimSpace(typeKey)
	if typeKey == "" {
		return Bridge{}, false
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, b := range bridges {
		if b.BridgeTypeKey == typeKey {
			return b, true
		}
	}
	return Bridge{}, false
}

// BridgeTypeKeys returns each registered bridge type once, in registry order.
func BridgeTypeKeys() []string {
	mu.RLock()
	defer mu.RUnlock()
	seen := make(map[string]bool)
	var out []string
	for _, b := range bridges {
		if seen[b.BridgeTypeKey] {
			continue
		}
		seen[b.BridgeTypeKey] = true
		out = append(out, b.BridgeTypeKey)
	}
	return out
}

// BridgeBindings returns Install bindings derived from registered bridges:
// endpoints locked (endpoint order). Product disambiguation is locked
// (role, relationship_type); a plugin's disambiguation stays unlocked.
// Dedupes by BridgeTypeKey (first wins).
func BridgeBindings() []Binding {
	mu.RLock()
	defer mu.RUnlock()
	seen := make(map[string]bool)
	var out []Binding
	for _, b := range bridges {
		if seen[b.BridgeTypeKey] {
			continue
		}
		seen[b.BridgeTypeKey] = true
		origin := strings.TrimSpace(b.Origin)
		if origin == "" {
			origin = OriginProvenencia
		}
		order := 0
		for _, e := range b.Endpoints {
			out = append(out, Binding{
				Origin: origin, TypeKey: b.BridgeTypeKey, PropertyKey: e.PropertyKey,
				SortOrder: order, Locked: true,
			})
			order++
		}
		if HasDisambiguation(b.Disambiguation) {
			out = append(out, Binding{
				Origin: origin, TypeKey: b.BridgeTypeKey, PropertyKey: b.Disambiguation,
				SortOrder: order, Locked: origin == OriginProvenencia,
			})
		}
	}
	return out
}
