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
