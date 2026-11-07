package propertymeta_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/propertymeta"
)

func TestCardinalityOrSingle(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"", propertymeta.CardinalitySingle},
		{propertymeta.CardinalitySingle, propertymeta.CardinalitySingle},
		{propertymeta.CardinalityMultiple, propertymeta.CardinalityMultiple},
	} {
		got := propertymeta.Meta{Cardinality: tc.in}.CardinalityOrSingle()
		if got != tc.want {
			t.Errorf("%q: got %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTypes(t *testing.T) {
	want := []string{"text", "integer", "date", "name", "subject", "term", "depiction", "boolean"}
	got := propertymeta.ValueTypes()
	if len(got) != len(want) {
		t.Fatalf("ValueTypes() = %v, want %v", got, want)
	}
	seen := map[string]bool{}
	for i, vt := range got {
		if vt != want[i] {
			t.Errorf("ValueTypes()[%d] = %q, want %q", i, vt, want[i])
		}
		if seen[vt] {
			t.Errorf("duplicate value type %q", vt)
		}
		seen[vt] = true
	}
	b, ok := propertymeta.Rule("boolean")
	if !ok || b.ResearcherCreatable || !b.SingleOnly {
		t.Errorf("Rule(boolean) = %+v, %v; want not creatable, single-only", b, ok)
	}
	if _, ok := propertymeta.Rule("nope"); ok {
		t.Error("Rule(nope) found")
	}
}
