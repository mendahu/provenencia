package subjectvocab

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/vocabkey"
)

// vocabkey names the seed's primary kinds and the Property keys code reads by
// name; the seed stays the source of truth for what exists.
func TestVocabKeysMatchSeed(t *testing.T) {
	var roots []string
	for _, st := range seedTypes {
		if st.Role == RoleRoot {
			roots = append(roots, st.Key)
		}
	}
	want := vocabkey.PrimaryKinds()
	slices.Sort(roots)
	slices.Sort(want)
	if !slices.Equal(roots, want) {
		t.Errorf("seed root types %v, vocabkey.PrimaryKinds %v", roots, want)
	}

	seeded := map[string]bool{}
	for _, p := range seedProperties {
		seeded[p.Key] = true
	}
	for _, key := range []string{
		vocabkey.PropertyName, vocabkey.PropertySexAtBirth, vocabkey.PropertyGender,
		vocabkey.PropertyEventType, vocabkey.PropertyDate, vocabkey.PropertyStartDate,
		vocabkey.PropertyEndDate, vocabkey.PropertyToponym,
	} {
		if !seeded[key] {
			t.Errorf("vocabkey Property %q is not a seeded Property", key)
		}
	}
}
