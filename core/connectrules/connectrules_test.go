package connectrules

import "testing"

func TestProductMatrix(t *testing.T) {
	t.Cleanup(ResetForTest)

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
	if !IsDisambiguation("participation", "role", OriginProvenencia) {
		t.Fatal("participation role")
	}
	if !IsDisambiguation("relationship", "relationship_type", OriginProvenencia) {
		t.Fatal("relationship_type")
	}
	if IsDisambiguation("location", "role", OriginProvenencia) {
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

func TestPolicyLoopsAll(t *testing.T) {
	t.Cleanup(ResetForTest)

	for _, r := range All() {
		if r.Refuse {
			continue
		}
		for _, e := range r.Endpoints {
			if !IsConnectionFacet(r.BridgeTypeKey, e.PropertyKey, r.Origin) {
				t.Fatalf("endpoint %s on %s is not a facet", e.PropertyKey, r.BridgeTypeKey)
			}
			got, ok := Edge(r.BridgeTypeKey, e.PropertyKey)
			if !ok || got != e.TypeKey {
				t.Fatalf("Edge %s/%s = %q %v want %q", r.BridgeTypeKey, e.PropertyKey, got, ok, e.TypeKey)
			}
		}
		if !HasDisambiguation(r.Disambiguation) {
			continue
		}
		if !IsDisambiguation(r.BridgeTypeKey, r.Disambiguation, r.Origin) {
			t.Fatalf("disambiguation %s on %s", r.Disambiguation, r.BridgeTypeKey)
		}
		if !IsConnectionFacet(r.BridgeTypeKey, r.Disambiguation, r.Origin) {
			t.Fatalf("disambiguation facet %s on %s", r.Disambiguation, r.BridgeTypeKey)
		}
		if IsEdgePair(r.BridgeTypeKey, r.Disambiguation, r.Origin) {
			t.Fatalf("disambiguation %s must not be an edge", r.Disambiguation)
		}
	}
}

func TestRegisterPluginBridge(t *testing.T) {
	t.Cleanup(ResetForTest)

	const pluginOrigin = "plugin:test"
	Register(Bridge{
		Origin: pluginOrigin, BridgeTypeKey: "dna_match",
		Endpoints: []Endpoint{
			{PropertyKey: "person", TypeKey: "person"},
			{PropertyKey: "match", TypeKey: "person"},
		},
		Disambiguation: "confidence",
		Pairs:          [][2]string{{"person", "person"}},
	})

	got, ok := Edge("dna_match", "match")
	if !ok || got != "person" {
		t.Fatalf("plugin edge %q %v", got, ok)
	}
	if !IsDisambiguation("dna_match", "confidence", pluginOrigin) {
		t.Fatal("plugin disambiguation")
	}
	if !IsConnectionFacet("dna_match", "person", pluginOrigin) {
		t.Fatal("plugin facet")
	}
	if IsConnectionFacet("dna_match", "person", "user") {
		t.Fatal("user origin must not match plugin bridge")
	}
	if IsConnectionFacet("dna_match", "person", OriginProvenencia) {
		t.Fatal("provenencia origin must not match plugin bridge")
	}

	var saw bool
	for _, r := range All() {
		if r.BridgeTypeKey == "dna_match" && r.FromTypeKey == "person" && r.ToTypeKey == "person" && !r.Refuse {
			saw = true
			if r.Origin != pluginOrigin || r.Disambiguation != "confidence" {
				t.Fatalf("plugin rule %+v", r)
			}
		}
	}
	if !saw {
		t.Fatal("List/All missing plugin pair")
	}

	var lockedMatch, unlockedConf bool
	for _, b := range BridgeBindings() {
		if b.TypeKey != "dna_match" {
			continue
		}
		if b.PropertyKey == "match" && b.Locked {
			lockedMatch = true
		}
		if b.PropertyKey == "confidence" && !b.Locked {
			unlockedConf = true
		}
	}
	if !lockedMatch || !unlockedConf {
		t.Fatalf("derived bindings match=%v confidence=%v", lockedMatch, unlockedConf)
	}
}

func TestBridgeAuthoredOnce(t *testing.T) {
	t.Cleanup(ResetForTest)

	bs := Bridges()
	if len(bs) != 3 {
		t.Fatalf("want 3 product bridges, got %d", len(bs))
	}
	var participationPairs int
	for _, r := range All() {
		if r.BridgeTypeKey == "participation" && !r.Refuse {
			participationPairs++
		}
	}
	if participationPairs != 2 {
		t.Fatalf("participation should expand to 2 directed pairs, got %d", participationPairs)
	}
}
