package connectrules

import "testing"

func TestSeedMatrix(t *testing.T) {
	endpoint, ok := Edge("participation", "person")
	if !ok || endpoint != "person" {
		t.Fatalf("participation/person %q %v", endpoint, ok)
	}
	endpoint, ok = Edge("relationship", "related_to")
	if !ok || endpoint != "person" {
		t.Fatalf("relationship/related_to %q %v", endpoint, ok)
	}
	if _, ok := Edge("participation", "role"); ok {
		t.Fatal("role is disambiguation, not an edge")
	}
	if !IsDisambiguation("participation", "role") {
		t.Fatal("participation role")
	}
	if !IsDisambiguation("relationship", "relationship_type") {
		t.Fatal("relationship_type")
	}
	if IsDisambiguation("location", "role") {
		t.Fatal("location has no role")
	}
	if !IsConnectionFacet("participation", "person", OriginProvenencia) {
		t.Fatal("seeded edge")
	}
	if !IsConnectionFacet("participation", "role", OriginProvenencia) {
		t.Fatal("seeded role")
	}
	if IsConnectionFacet("participation", "person", "user") {
		t.Fatal("user-origin edge is not a connection facet")
	}
	if IsConnectionFacet("person", "role", OriginProvenencia) {
		t.Fatal("role on person is not a bridge facet")
	}
}

func TestPolicyLoopsSeed(t *testing.T) {
	for _, r := range Seed {
		if r.Refuse {
			continue
		}
		for _, e := range r.Endpoints {
			if !IsConnectionFacet(r.BridgeTypeKey, e.PropertyKey, OriginProvenencia) {
				t.Fatalf("endpoint %s on %s is not a facet", e.PropertyKey, r.BridgeTypeKey)
			}
			got, ok := Edge(r.BridgeTypeKey, e.PropertyKey)
			if !ok || got != e.TypeKey {
				t.Fatalf("Edge %s/%s = %q %v want %q", r.BridgeTypeKey, e.PropertyKey, got, ok, e.TypeKey)
			}
		}
		if r.Disambiguation == DisambiguationNone || r.Disambiguation == "" {
			continue
		}
		if !IsDisambiguation(r.BridgeTypeKey, r.Disambiguation) {
			t.Fatalf("disambiguation %s on %s", r.Disambiguation, r.BridgeTypeKey)
		}
		if !IsConnectionFacet(r.BridgeTypeKey, r.Disambiguation, OriginProvenencia) {
			t.Fatalf("disambiguation facet %s on %s", r.Disambiguation, r.BridgeTypeKey)
		}
		if IsEdgePair(r.BridgeTypeKey, r.Disambiguation) {
			t.Fatalf("disambiguation %s must not be an edge", r.Disambiguation)
		}
	}
}
