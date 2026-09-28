// Package connectrules is the seeded Interpretation connect matrix.
// subjectvocab Installs it; deleteimpact classifies connection facets from it.
package connectrules

import "strings"

const (
	DisambiguationNone             = "none"
	DisambiguationRole             = "role"
	DisambiguationRelationshipType = "relationship_type"

	OriginProvenencia = "provenencia"
)

// Rule is one allowed (or refused) connect endpoint pair.
type Rule struct {
	FromTypeKey      string
	ToTypeKey        string
	BridgeTypeKey    string
	EdgePropertyKeys []string
	Disambiguation   string
	Refuse           bool
}

// Seed is the product connect matrix. Omitted pairs refuse by default.
var Seed = []Rule{
	{
		FromTypeKey: "person", ToTypeKey: "event",
		BridgeTypeKey:    "participation",
		EdgePropertyKeys: []string{"person", "event"},
		Disambiguation:   DisambiguationRole,
	},
	{
		FromTypeKey: "event", ToTypeKey: "person",
		BridgeTypeKey:    "participation",
		EdgePropertyKeys: []string{"person", "event"},
		Disambiguation:   DisambiguationRole,
	},
	{
		FromTypeKey: "person", ToTypeKey: "person",
		BridgeTypeKey:    "relationship",
		EdgePropertyKeys: []string{"person", "related_to"},
		Disambiguation:   DisambiguationRelationshipType,
	},
	{
		FromTypeKey: "event", ToTypeKey: "place",
		BridgeTypeKey:    "location",
		EdgePropertyKeys: []string{"event", "place"},
		Disambiguation:   DisambiguationNone,
	},
	{
		FromTypeKey: "place", ToTypeKey: "event",
		BridgeTypeKey:    "location",
		EdgePropertyKeys: []string{"event", "place"},
		Disambiguation:   DisambiguationNone,
	},
	{FromTypeKey: "person", ToTypeKey: "place", Refuse: true},
	{FromTypeKey: "place", ToTypeKey: "person", Refuse: true},
	{FromTypeKey: "event", ToTypeKey: "event", Refuse: true},
	{FromTypeKey: "place", ToTypeKey: "place", Refuse: true},
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
		for _, e := range edgesFromKeys(r.EdgePropertyKeys) {
			if e.propertyKey == propertyKey {
				return e.endpointTypeKey, true
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

type edge struct {
	propertyKey     string
	endpointTypeKey string
}

func edgesFromKeys(keys []string) []edge {
	out := make([]edge, 0, len(keys))
	for _, key := range keys {
		endpoint := key
		if key == "related_to" {
			endpoint = "person"
		}
		out = append(out, edge{propertyKey: key, endpointTypeKey: endpoint})
	}
	return out
}
