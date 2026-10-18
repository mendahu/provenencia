package matchregistry

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Default builds a fresh registry each call (D-2): one caller's edits reach no
// one else.
func TestDefaultIsFresh(t *testing.T) {
	first := Default()
	want := Default()
	first.Value.Words.Particles[0] = "changed"
	first.Value.Words.Folds['x'] = "added"
	first.Value.Names.Western.PartRoles["x"] = valuecompare.RoleFamily
	first.Subject.Weights[vocabkey.Person][vocabkey.PropertyName] = subjectcompare.Weights{}
	first.Subject.Pools[vocabkey.Person][vocabkey.PropertySexAtBirth].With[0] = "changed"
	first.Subject.Periods[vocabkey.Event][vocabkey.PropertyDate] = subjectcompare.Period{}
	if got := Default(); !reflect.DeepEqual(got, want) {
		t.Errorf("second Default() = %+v, want %+v", got, want)
	}
}

// The Western pattern gives every product part type a role, so a new part
// type forces a decision here.
func TestWesternPatternCoversEveryPartType(t *testing.T) {
	western := Default().Value.Names.Western
	for _, pt := range namevalue.PartTypes() {
		if _, ok := western.PartRoles[pt.Key]; !ok {
			t.Errorf("part type %q has no role in the Western pattern", pt.Key)
		}
	}
	if _, ok := western.PartRoles[""]; !ok {
		t.Error("untyped parts have no role in the Western pattern")
	}
}
