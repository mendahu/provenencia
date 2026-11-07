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
	flags := []struct {
		valueType           string
		researcherCreatable bool
		singleOnly          bool
	}{
		{"text", true, false},
		{"integer", true, false},
		{"date", true, false},
		{"name", true, false},
		{"subject", true, false},
		{"term", false, false},
		{"depiction", false, false},
		{"boolean", false, true},
	}
	if len(flags) != len(want) {
		t.Fatalf("flags table has %d rows, want %d", len(flags), len(want))
	}
	for _, f := range flags {
		r, ok := propertymeta.Rule(f.valueType)
		if !ok || r.ResearcherCreatable != f.researcherCreatable || r.SingleOnly != f.singleOnly {
			t.Errorf("Rule(%s) = %+v, %v; want creatable %v, single-only %v", f.valueType, r, ok, f.researcherCreatable, f.singleOnly)
		}
	}
	if _, ok := propertymeta.Rule("nope"); ok {
		t.Error("Rule(nope) found")
	}
}
