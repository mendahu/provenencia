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
