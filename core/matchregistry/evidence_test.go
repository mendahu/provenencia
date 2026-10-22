package matchregistry

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The evidence section: the shipped relations, with no relation both checked
// and merely pinnable.
func TestEvidenceSection(t *testing.T) {
	e := Default().Evidence
	if want := []valuecompare.Relation{valuecompare.Same, valuecompare.Equivalent}; !slices.Equal(e.Checked, want) {
		t.Errorf("Checked = %v, want %v", e.Checked, want)
	}
	if want := []valuecompare.Relation{valuecompare.Contains, valuecompare.Overlaps, valuecompare.Near}; !slices.Equal(e.Pinnable, want) {
		t.Errorf("Pinnable = %v, want %v", e.Pinnable, want)
	}
	if want := []valuecompare.Cause{valuecompare.CauseUnjudged}; !slices.Equal(e.ByEye, want) {
		t.Errorf("ByEye = %v, want %v", e.ByEye, want)
	}
	for _, r := range e.Checked {
		if slices.Contains(e.Pinnable, r) {
			t.Errorf("relation %s is both checked and pinnable", r)
		}
	}
	if want := []vocabkey.PropertyKey{vocabkey.PropertyEventType}; !slices.Equal(e.Defining[vocabkey.Event], want) {
		t.Errorf("Defining[event] = %v, want %v", e.Defining[vocabkey.Event], want)
	}
}
