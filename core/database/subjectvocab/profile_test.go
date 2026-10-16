package subjectvocab

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/match"
)

// A default matching Feature carries the cardinality the seeded Property
// has, so the two registries cannot drift.
func TestDefaultProfileCardinality(t *testing.T) {
	card := map[string]string{}
	for _, p := range seedProperties {
		c := p.Cardinality
		if c == "" {
			c = properties.CardinalitySingle
		}
		card[p.Key] = c
	}
	for _, kind := range []string{"person", "event", "place"} {
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
				got = properties.CardinalitySingle
			}
			if got != want {
				t.Errorf("%s.%s: Feature cardinality %q, seeded Property %q", kind, f.Property.Key, got, want)
			}
		}
	}
}
