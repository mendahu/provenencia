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
	first.Subject.Lifespans[vocabkey.Person][vocabkey.PropertyLifespan].Starts[0] = subjectcompare.EventRole{}
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

// Every lead shortlisting keys by is a catalog property the subject section
// weighs, not a derived lead: the index builds no life.
func TestShortlistSection(t *testing.T) {
	reg := Default()
	rules := reg.Alignment.Shortlist
	if rules.YearBucket <= 0 {
		t.Errorf("YearBucket = %d, want > 0", rules.YearBucket)
	}
	check := func(kind vocabkey.SubjectType, lead vocabkey.PropertyKey) {
		t.Helper()
		if _, ok := reg.Subject.Weights[kind][lead]; !ok {
			t.Errorf("%s lead %s has no weights in the subject section", kind, lead)
		}
		if _, derived := reg.Subject.Lifespans[kind][lead]; derived {
			t.Errorf("%s lead %s is derived (a lifespan): the shortlist builds no life", kind, lead)
		}
	}
	for kind, leads := range rules.Words {
		for _, lead := range leads {
			check(kind, lead)
		}
	}
	for kind, d := range rules.Dated {
		check(kind, d.Type)
		check(kind, d.Date)
	}
}
