package subjectprofile_test

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

type observed = map[vocabkey.PropertyKey][]autoreconcile.Candidate

func pos(vs ...autoreconcile.Value) []autoreconcile.Candidate {
	var out []autoreconcile.Candidate
	for _, v := range vs {
		out = append(out, vcand(v, false))
	}
	return out
}

func TestEntityCompare(t *testing.T) {
	sexResult := func(outcome propertycompare.Outcome, cause propertycompare.Cause) func(*testing.T, subjectcompare.Result) {
		return func(t *testing.T, r subjectcompare.Result) {
			t.Helper()
			for _, p := range r.Properties {
				if p.Key == sex {
					if p.Result.Outcome != outcome || p.Result.Cause != cause {
						t.Errorf("sex_at_birth: got %s/%s, want %s/%s", p.Result.Outcome, p.Result.Cause, outcome, cause)
					}
					return
				}
			}
			t.Error("no sex_at_birth result")
		}
	}
	tests := []struct {
		name    string
		kind    vocabkey.SubjectType
		entity  func(t *testing.T) subjectprofile.Entity
		opts    subjectprofile.Options
		subject observed
		sim, ev float64
		check   func(t *testing.T, r subjectcompare.Result)
	}{
		{name: "AC-3.1 a Subject recording the outvoted value conflicts", kind: vocabkey.Person, entity: maleFemale,
			subject: observed{sex: pos(term("female"))}, sim: 0, ev: -1,
			check: sexResult(propertycompare.OutcomeConflict, propertycompare.CauseRelation)},
		{name: "AC-3.2 a Subject recording the displayed value agrees", kind: vocabkey.Person, entity: maleFemale,
			subject: observed{sex: pos(term("male"))}, sim: 1, ev: 0.6},
		{name: "AC-3.3 a Subject recording the denied value conflicts through the negative", kind: vocabkey.Person, entity: denied,
			subject: observed{sex: pos(term("female"))}, sim: 0, ev: -1,
			check: sexResult(propertycompare.OutcomeConflict, propertycompare.CauseNegative)},
		{name: "AC-3.4 the pool carries gender to the negative", kind: vocabkey.Person, entity: denied,
			subject: observed{vocabkey.PropertyGender: pos(term("female"))}, sim: 0, ev: -1,
			check: sexResult(propertycompare.OutcomeConflict, propertycompare.CauseNegative)},
		{name: "AC-3.5 a Subject recording another value is unknown", kind: vocabkey.Person, entity: denied,
			subject: observed{sex: pos(term("male"))}, sim: 0, ev: 0,
			check: sexResult(propertycompare.OutcomeUnknown, propertycompare.CauseNoValues)},
		{name: "AC-3.6 leave-one-out drops that member's support", kind: vocabkey.Person, entity: maleFemale,
			opts:    subjectprofile.Options{Excluding: [][]byte{mem(1)}},
			subject: observed{sex: pos(term("female"))}, sim: 1, ev: 0.6},
		{name: "AC-3.7 a person entity against a Subject", kind: vocabkey.Person, entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{
				nameProp: property(t, nameProp, cand(1, 1, 1, name("James Robins")), cand(2, 2, 2, name("James Robins"))),
				sex:      property(t, sex, cand(1, 1, 1, term("male")), cand(2, 2, 2, term("male"))),
			}}
		}, subject: observed{nameProp: pos(name("James Robins")), sex: pos(term("female"))}, sim: 0.7143, ev: 1.5},
		{name: "AC-3.8 an event entity against a Subject", kind: vocabkey.Event, entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{
				vocabkey.PropertyEventType: property(t, vocabkey.PropertyEventType, cand(1, 1, 1, term("birth"))),
				date:                       property(t, date, cand(1, 1, 1, year(1817)), cand(2, 2, 2, year(1830))),
			}}
		}, subject: observed{vocabkey.PropertyEventType: pos(term("birth")), date: pos(year(1830))}, sim: 1, ev: 2.5},
		{name: "AC-3.9 a place entity against a Subject", kind: vocabkey.Place, entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{toponym: property(t, toponym,
				cand(1, 1, 1, text("York")), cand(2, 2, 2, text("York")), cand(3, 3, 3, text("London")))}}
		}, subject: observed{toponym: pos(text("London"))}, sim: 1, ev: 3},
		{name: "AC-5.5 two negatives never meet", kind: vocabkey.Person, entity: denied,
			subject: observed{sex: {vcand(term("female"), true)}}, sim: 0, ev: 0,
			check: sexResult(propertycompare.OutcomeUnknown, propertycompare.CauseNoValues)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg := matchregistry.Default()
			metas := subjectcomparetest.Metas()
			a, err := subjectprofile.Build(tc.entity(t), metas, reg.Value, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			b := subjectprofile.ForSubject(tc.subject, "")
			r := subjectcompare.Compare(tc.kind, a, b, metas, reg.Subject, reg.Value)
			if math.Abs(r.Similarity-tc.sim) > 0.0001 || math.Abs(r.Evidence-tc.ev) > 0.0001 {
				t.Errorf("sim %.4f ev %.4f, want sim %.4f ev %.4f (a=%s b=%s)", r.Similarity, r.Evidence, tc.sim, tc.ev,
					subjectcomparetest.Labels(a), subjectcomparetest.Labels(b))
			}
			if tc.check != nil {
				tc.check(t, r)
			}
		})
	}
}
