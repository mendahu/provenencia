package subjectprofile_test

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func TestLifespanFromEvents(t *testing.T) {
	const (
		birth, death, burial = vocabkey.EventBirth, vocabkey.EventDeath, vocabkey.EventBurial
		subj                 = vocabkey.RoleSubject
	)
	var (
		baptism   = vocabkey.EventType("baptism")
		residence = vocabkey.EventType("residence")
		marriage  = vocabkey.EventType("marriage")
		y         = func(n int) *datevalue.Value { return datevaluetest.YMD(n, 0, 0) }
	)
	empty := func(t *testing.T, p subjectcompare.Values) {
		t.Helper()
		if p == nil || len(p) != 0 {
			t.Errorf("want non-nil empty profile, got %s", subjectcomparetest.Labels(p))
		}
	}
	ends := func(t *testing.T, p subjectcompare.Values, key vocabkey.PropertyKey) (first, last valuecompare.DateEnd) {
		t.Helper()
		if len(p[key]) != 1 || p[key][0].Date == nil {
			t.Fatalf("want one %s value, got %s", key, subjectcomparetest.Labels(p))
		}
		first, _, ok := valuecompare.DateEnds(*p[key][0].Date)
		_, last, ok2 := valuecompare.DateEnds(*p[key][0].Date)
		if !ok || !ok2 {
			t.Fatalf("no ends for %s", key)
		}
		return first, last
	}
	bef1840 := datevalue.Value{Kind: datevalue.KindPoint, Qualifier: datevalue.QualifierBEF, StartYear: datevaluetest.Part(1840)}
	tests := []struct {
		name    string
		events  []subjectprofile.LifeEvent
		noRules bool
		values  map[vocabkey.PropertyKey][]autoreconcile.Candidate
		want    string
		check   func(t *testing.T, p subjectcompare.Values)
		// missing is why ForSubject returns no lifespan; first and last, when
		// set, are the dates as written that set the span.
		missing     subjectcompare.LifespanMissing
		first, last *datevalue.Value
	}{
		{name: "AC-2.1 a birth starts the life", events: []subjectprofile.LifeEvent{ev(birth, subj, y(1850))},
			want: "lifespan={Between 1840 and 1970}; lifespan_span={Between 1850 and 1850}", first: y(1850), last: y(1850)},
		{name: "AC-2.2 a death ends the life", events: []subjectprofile.LifeEvent{ev(death, subj, y(1840))},
			want: "lifespan={Between 1720 and 1850}; lifespan_span={Between 1840 and 1840}"},
		{name: "AC-2.3 a birth and a death", events: []subjectprofile.LifeEvent{ev(birth, subj, y(1850)), ev(death, subj, y(1920))},
			want: "lifespan={Between 1840 and 1930}; lifespan_span={Between 1850 and 1920}"},
		{name: "AC-2.4 any event counts", events: []subjectprofile.LifeEvent{ev(residence, subj, y(1950))},
			want: "lifespan={Between 1830 and 2070}; lifespan_span={Between 1950 and 1950}"},
		{name: "AC-2.5 a burial ends the life", events: []subjectprofile.LifeEvent{ev(baptism, subj, y(1830)), ev(burial, subj, y(1901))},
			want: "lifespan={Between 1710 and 1911}; lifespan_span={Between 1830 and 1901}"},
		{name: "AC-2.6 a child's birth is not the person's", events: []subjectprofile.LifeEvent{ev(birth, vocabkey.Role("father"), y(1880))},
			want: "lifespan={Between 1760 and 2000}; lifespan_span={Between 1880 and 1880}"},
		{name: "AC-2.7 order does not matter", events: []subjectprofile.LifeEvent{
			ev(residence, subj, y(1900)), ev(birth, subj, y(1850)), ev(marriage, vocabkey.Role("spouse"), y(1875))},
			want: "lifespan={Between 1840 and 2020}; lifespan_span={Between 1850 and 1900}"},
		{name: "AC-2.8 the earliest birth starts the life", events: []subjectprofile.LifeEvent{ev(birth, subj, y(1852)), ev(birth, subj, y(1850))},
			want: "lifespan={Between 1840 and 1972}; lifespan_span={Between 1850 and 1852}"},
		{name: "AC-2.9 a death before a year is skipped and leaves the side undated", events: []subjectprofile.LifeEvent{ev(death, subj, &bef1840)},
			check: empty, missing: subjectcompare.LifespanUndated},
		{name: "AC-2.10 a leap day moves to the last day of February", events: []subjectprofile.LifeEvent{ev(birth, subj, datevaluetest.YMD(1904, 2, 29))},
			check: func(t *testing.T, p subjectcompare.Values) {
				first, last := ends(t, p, vocabkey.PropertyLifespan)
				if first.Year != 1894 || first.Month != 2 || first.Day != 28 || last.Year != 2024 || last.Month != 2 || last.Day != 28 {
					t.Errorf("window ends %+v %+v", first, last)
				}
			}},
		{name: "AC-2.11 undated events give no lifespan", events: []subjectprofile.LifeEvent{
			{EventType: residence, Role: subj},
			ev(residence, subj, datevaluetest.Phrase("spring"))},
			check: empty, missing: subjectcompare.LifespanUndated},
		{name: "AC-2.12 a window that would end before it starts", events: []subjectprofile.LifeEvent{ev(birth, subj, y(1900)), ev(death, subj, y(1850))},
			check: empty, missing: subjectcompare.LifespanContradicts},
		{name: "AC-2.13 no rules give no lifespan", noRules: true, events: []subjectprofile.LifeEvent{ev(birth, subj, y(1850))},
			check: empty},
		{name: "AC-2.15 a plain Subject keeps its values beside the lifespan",
			events: []subjectprofile.LifeEvent{ev(birth, subj, y(1850))},
			values: map[vocabkey.PropertyKey][]autoreconcile.Candidate{sex: {vcand(term("male"), false)}},
			want:   "lifespan={Between 1840 and 1970}; lifespan_span={Between 1850 and 1850}; sex_at_birth={male}"},
		{name: "AC-2.16 a residence wholly before the birth", events: []subjectprofile.LifeEvent{ev(birth, subj, y(1900)), ev(residence, subj, y(1850))},
			check: empty, missing: subjectcompare.LifespanContradicts},
		{name: "AC-2.17 a residence wholly after the death", events: []subjectprofile.LifeEvent{ev(death, subj, y(1850)), ev(residence, subj, y(1900))},
			check: empty, missing: subjectcompare.LifespanContradicts},
		{name: "AC-2.20 an open-ended date beside an exact one is skipped, not a contradiction", events: []subjectprofile.LifeEvent{ev(baptism, subj, y(1900)), ev(residence, subj, datevaluetest.Before(1780, 0, 0))},
			want: "lifespan={Between 1780 and 2020}; lifespan_span={Between 1900 and 1900}", first: y(1900), last: y(1900)},
		{name: "AC-2.18 a coarser date in the same year is no contradiction", events: []subjectprofile.LifeEvent{
			ev(birth, subj, datevaluetest.YMD(1850, 5, 1)), ev(baptism, subj, y(1850))},
			first: y(1850), last: y(1850),
			check: func(t *testing.T, p subjectcompare.Values) {
				first, last := ends(t, p, vocabkey.PropertyLifespan)
				if first.Year != 1840 || first.Month != 5 || first.Day != 1 || last.Year != 1970 || last.Month != 0 || last.Day != 0 {
					t.Errorf("window ends %+v %+v", first, last)
				}
				sFirst, _ := ends(t, p, vocabkey.PropertyLifespanSpan)
				if sFirst.Year != 1850 || sFirst.Month != 0 || sFirst.Day != 0 {
					t.Errorf("span first end %+v", sFirst)
				}
			}},
		{name: "a date written about stays as written", events: []subjectprofile.LifeEvent{
			ev(birth, subj, datevaluetest.About(1815, 0, 0)), ev(death, subj, y(1890))},
			want:  "lifespan={Between 1805 and 1900}; lifespan_span={Between 1815 and 1890}",
			first: datevaluetest.About(1815, 0, 0), last: y(1890)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rules := matchregistry.Default().Subject.Lifespans[vocabkey.Person]
			if tc.noRules {
				rules = nil
			}
			got, side := subjectprofile.ForSubject(subjectprofile.Subject{Values: tc.values, Life: tc.events}, rules)
			if side.Missing != tc.missing || (tc.missing != "" || tc.noRules) && (side.First != nil || side.Last != nil) {
				t.Errorf("side = %+v, want missing %q and no dates", side, tc.missing)
			}
			if tc.first != nil && (!reflect.DeepEqual(side.First, tc.first) || !reflect.DeepEqual(side.Last, tc.last)) {
				t.Errorf("side first %v last %v, want %v and %v", side.First, side.Last, tc.first, tc.last)
			}
			if tc.check != nil {
				tc.check(t, got)
				return
			}
			if g := subjectcomparetest.Labels(got); g != tc.want {
				t.Errorf("got %s, want %s", g, tc.want)
			}
		})
	}
}
