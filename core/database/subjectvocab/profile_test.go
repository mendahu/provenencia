package subjectvocab

import (
	"testing"

	"github.com/mendahu/provenencia/core/match"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// A default matching Feature carries the cardinality and value type the
// seeded Property has, so the two registries cannot drift.
func TestDefaultProfileCardinality(t *testing.T) {
	card := map[string]string{}
	valueType := map[string]string{}
	for _, p := range seedProperties {
		c := p.Cardinality
		if c == "" {
			c = propertymeta.CardinalitySingle
		}
		card[p.Key] = c
		valueType[p.Key] = p.ValueType
	}
	for _, kind := range vocabkey.PrimaryKinds() {
		prof, ok := match.DefaultProfile(kind)
		if !ok {
			t.Fatalf("no default profile for %s", kind)
		}
		for _, f := range prof.Features {
			want, ok := card[f.Property.Key]
			if !ok {
				t.Errorf("%s: Feature %s is not a seeded Property", kind, f.Property.Key)
				continue
			}
			got := f.Cardinality
			if got == "" {
				got = propertymeta.CardinalitySingle
			}
			if got != want {
				t.Errorf("%s.%s: Feature cardinality %q, seeded Property %q", kind, f.Property.Key, got, want)
			}
			if f.ValueType != valueType[f.Property.Key] {
				t.Errorf("%s.%s: Feature value type %q, seeded Property %q", kind, f.Property.Key, f.ValueType, valueType[f.Property.Key])
			}
		}
	}
}
