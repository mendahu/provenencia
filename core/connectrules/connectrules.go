// Package connectrules is the seeded Interpretation connect matrix.
//
// This is the only product table of “this property on this bridge type is an
// endpoint or a disambiguation.” subjectvocab Installs the same keys; Impact,
// edge-lock, and Connect loop Seed — they do not name role / relationship_type.
package connectrules

import "strings"

const (
	DisambiguationNone             = "none"
	DisambiguationRole             = "role"
	DisambiguationRelationshipType = "relationship_type"

	OriginProvenencia = "provenencia"
)

// Endpoint binds a Property key on a bridge to the node type it points at.
type Endpoint struct {
	PropertyKey string
	TypeKey     string
}

// Rule is one allowed (or refused) connect endpoint pair.
type Rule struct {
	FromTypeKey    string
	ToTypeKey      string
	BridgeTypeKey  string
	Endpoints      []Endpoint
	Disambiguation string
	Refuse         bool
}

// Seed is the product connect matrix. Omitted pairs refuse by default.
// Add a bridge property here (Endpoints or Disambiguation) and Impact /
// edge-lock / Connect pick it up. Bind the same key in subjectvocab seedBindings.
var Seed = []Rule{
	{
		FromTypeKey: "person", ToTypeKey: "event",
		BridgeTypeKey: "participation",
		Endpoints: []Endpoint{
			{PropertyKey: "person", TypeKey: "person"},
			{PropertyKey: "event", TypeKey: "event"},
		},
		Disambiguation: DisambiguationRole,
	},
	{
		FromTypeKey: "event", ToTypeKey: "person",
		BridgeTypeKey: "participation",
		Endpoints: []Endpoint{
			{PropertyKey: "person", TypeKey: "person"},
			{PropertyKey: "event", TypeKey: "event"},
		},
		Disambiguation: DisambiguationRole,
	},
	{
		FromTypeKey: "person", ToTypeKey: "person",
		BridgeTypeKey: "relationship",
		Endpoints: []Endpoint{
			{PropertyKey: "person", TypeKey: "person"},
			{PropertyKey: "related_to", TypeKey: "person"},
		},
		Disambiguation: DisambiguationRelationshipType,
	},
	{
		FromTypeKey: "event", ToTypeKey: "place",
		BridgeTypeKey: "location",
		Endpoints: []Endpoint{
			{PropertyKey: "event", TypeKey: "event"},
			{PropertyKey: "place", TypeKey: "place"},
		},
		Disambiguation: DisambiguationNone,
	},
	{
		FromTypeKey: "place", ToTypeKey: "event",
		BridgeTypeKey: "location",
		Endpoints: []Endpoint{
			{PropertyKey: "event", TypeKey: "event"},
			{PropertyKey: "place", TypeKey: "place"},
		},
		Disambiguation: DisambiguationNone,
	},
	{FromTypeKey: "person", ToTypeKey: "place", Refuse: true},
	{FromTypeKey: "place", ToTypeKey: "person", Refuse: true},
	{FromTypeKey: "event", ToTypeKey: "event", Refuse: true},
	{FromTypeKey: "place", ToTypeKey: "place", Refuse: true},
}

// EdgePropertyKeys is the endpoint property list on a rule (FFI / UI order).
func (r Rule) EdgePropertyKeys() []string {
	out := make([]string, 0, len(r.Endpoints))
	for _, e := range r.Endpoints {
		out = append(out, e.PropertyKey)
	}
	return out
}

// Edge returns the endpoint type a bridge edge property binds, if the pair
// is a non-refused connect rule.
func Edge(bridgeTypeKey, propertyKey string) (endpointTypeKey string, ok bool) {
	bridgeTypeKey = strings.TrimSpace(bridgeTypeKey)
	propertyKey = strings.TrimSpace(propertyKey)
	if bridgeTypeKey == "" || propertyKey == "" {
		return "", false
	}
	for _, r := range Seed {
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

// IsEdgePair reports whether propertyKey is a seeded edge on bridge typeKey.
func IsEdgePair(typeKey, propKey string) bool {
	_, ok := Edge(typeKey, propKey)
	return ok
}

// IsDisambiguation reports whether propKey is the seeded disambiguation on typeKey.
func IsDisambiguation(typeKey, propKey string) bool {
	typeKey = strings.TrimSpace(typeKey)
	propKey = strings.TrimSpace(propKey)
	if typeKey == "" || propKey == "" {
		return false
	}
	for _, r := range Seed {
		if r.Refuse || r.BridgeTypeKey != typeKey {
			continue
		}
		if r.Disambiguation == DisambiguationNone || r.Disambiguation == "" {
			continue
		}
		if r.Disambiguation == propKey {
			return true
		}
	}
	return false
}

// IsConnectionFacet reports a seeded (origin provenencia) edge or
// disambiguation Observation on a bridge type.
func IsConnectionFacet(typeKey, propKey, origin string) bool {
	if strings.TrimSpace(origin) != OriginProvenencia {
		return false
	}
	return IsEdgePair(typeKey, propKey) || IsDisambiguation(typeKey, propKey)
}
