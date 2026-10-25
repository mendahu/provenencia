package vocabkey

import (
	"slices"
	"testing"
)

func TestIsPrimaryType(t *testing.T) {
	tests := []struct {
		key, origin string
		want        bool
	}{
		{"person", "provenencia", true},
		{"event", "provenencia", true},
		{"place", "provenencia", true},
		{"person", "user", false},
		{"person", "plugin:x", false},
		{"participation", "provenencia", false},
		{"source", "provenencia", false},
		{"Person", "provenencia", false},
	}
	for _, tc := range tests {
		if got := IsPrimaryType(tc.key, tc.origin); got != tc.want {
			t.Errorf("IsPrimaryType(%q, %q) = %v, want %v", tc.key, tc.origin, got, tc.want)
		}
	}
}

func TestPrimaryKinds(t *testing.T) {
	got := PrimaryKinds()
	if want := []string{"person", "event", "place"}; !slices.Equal(got, want) {
		t.Fatalf("PrimaryKinds() = %v, want %v", got, want)
	}
	got[0] = "x"
	if PrimaryKinds()[0] != "person" {
		t.Fatal("PrimaryKinds shares its backing slice between calls")
	}
	for _, k := range PrimaryKinds() {
		if !IsPrimaryKind(k) {
			t.Errorf("IsPrimaryKind(%q) = false", k)
		}
	}
	if IsPrimaryKind("location") {
		t.Error("IsPrimaryKind(location) = true")
	}
}

func TestValidOrigin(t *testing.T) {
	tests := []struct {
		origin        string
		valid, plugin bool
	}{
		{"provenencia", true, false},
		{"user", true, false},
		{"plugin:dna", true, true},
		{"plugin:", false, false},
		{"", false, false},
		{"Provenencia", false, false},
		{"plugins:x", false, false},
	}
	for _, tc := range tests {
		if got := ValidOrigin(tc.origin); got != tc.valid {
			t.Errorf("ValidOrigin(%q) = %v, want %v", tc.origin, got, tc.valid)
		}
		if got := IsPluginOrigin(tc.origin); got != tc.plugin {
			t.Errorf("IsPluginOrigin(%q) = %v, want %v", tc.origin, got, tc.plugin)
		}
	}
}

func TestBridgeKeyString(t *testing.T) {
	tests := []struct {
		key  BridgeKey
		want string
	}{
		{BridgeKey{BridgeParticipation, "subject", PropertyPerson, EventBirth}, "participation·subject·person·birth"},
		{BridgeKey{Type: BridgeRelationship, Term: "spouse"}, "relationship·spouse"},
		{BridgeKey{Type: BridgeLocation, End: PropertyEvent}, "location·event"},
	}
	for _, tc := range tests {
		if got := tc.key.String(); got != tc.want {
			t.Errorf("%+v.String() = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestCompareByOrder(t *testing.T) {
	order := []PropertyKey{"name", "likeness", "gender"}
	tests := []struct {
		a, b PropertyKey
		want int
	}{
		{"name", "likeness", -1},
		{"gender", "likeness", 1},
		{"name", "name", 0},
		{"gender", "alpha", -1},
		{"alpha", "gender", 1},
		{"alpha", "beta", -1},
		{"beta", "alpha", 1},
	}
	for _, tc := range tests {
		got := CompareByOrder(order, tc.a, tc.b)
		if (got < 0) != (tc.want < 0) || (got > 0) != (tc.want > 0) {
			t.Errorf("CompareByOrder(%q, %q) = %d, want sign %d", tc.a, tc.b, got, tc.want)
		}
	}
	if CompareByOrder(nil, "b", "a") <= 0 {
		t.Error("with no order, keys rank by key")
	}
}
