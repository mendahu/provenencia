package subjectprofile_test

import (
	"math"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
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
	lifespanResult := func(outcome propertycompare.Outcome) func(*testing.T, subjectcompare.Result) {
		return func(t *testing.T, r subjectcompare.Result) {
			t.Helper()
			for _, p := range r.Properties {
				if p.Key == vocabkey.PropertyLifespan {
					if p.Result.Outcome != outcome {
						t.Errorf("lifespan: got %s, want %s", p.Result.Outcome, outcome)
					}
					return
				}
			}
			t.Error("no lifespan result")
		}
	}
	lifespanNot := func(outcome propertycompare.Outcome) func(*testing.T, subjectcompare.Result) {
		return func(t *testing.T, r subjectcompare.Result) {
			t.Helper()
			for _, p := range r.Properties {
				if p.Key == vocabkey.PropertyLifespan {
					if p.Result.Outcome == outcome {
						t.Errorf("lifespan: got %s, want anything else", p.Result.Outcome)
					}
					return
				}
			}
			t.Error("no lifespan result")
		}
	}
	james := func(t *testing.T) subjectprofile.Entity {
		return subjectprofile.Entity{Properties: props{
			nameProp: property(t, nameProp, cand(1, 1, 1, name("James Robins"))),
		}}
	}
	jamesSubject := observed{nameProp: pos(name("James Robins"))}
	subjectRole, birth, death, residence := vocabkey.RoleSubject, vocabkey.EventBirth, vocabkey.EventDeath, vocabkey.EventType("residence")
	life := func(eventType vocabkey.EventType, y int) []subjectprofile.LifeEvent {
		return []subjectprofile.LifeEvent{ev(eventType, subjectRole, datevaluetest.YMD(y, 0, 0))}
	}
	tests := []struct {
		name              string
		kind              vocabkey.SubjectType
		entity            func(t *testing.T) subjectprofile.Entity
		opts              subjectprofile.Options
		subject           observed
		life, subjectLife []subjectprofile.LifeEvent
		sim, ev           float64
		check             func(t *testing.T, r subjectcompare.Result)
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
			subject: observed{vocabkey.PropertyGender: pos(term("woman"))}, sim: 0, ev: -1,
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
		{name: "AC-3.10 an entity that died before the Subject was born", kind: vocabkey.Person, entity: james,
			life: life(death, 1840), subject: jamesSubject, subjectLife: life(birth, 1850), sim: 2.5 / 22.5, ev: -17.5,
			check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "AC-3.11 an entity born before the Subject's span ended", kind: vocabkey.Person, entity: james,
			life: life(birth, 1850), subject: jamesSubject, subjectLife: life(death, 1990), sim: 2.5 / 22.5, ev: -17.5,
			check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "AC-3.12 consistent lifespans agree", kind: vocabkey.Person, entity: james,
			life: life(birth, 1850), subject: jamesSubject, subjectLife: life(death, 1920), sim: 1, ev: 2.6,
			check: lifespanResult(propertycompare.OutcomeAgree)},
		{name: "AC-3.13 the Subject resided after the entity died", kind: vocabkey.Person, entity: james,
			life: life(death, 1900), subject: jamesSubject, subjectLife: life(residence, 1950), sim: 2.5 / 22.5, ev: -17.5,
			check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "AC-3.14 two births on one day agree", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1901, 5, 1))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1901, 5, 1))}, sim: 1, ev: 2.6,
			check: lifespanResult(propertycompare.OutcomeAgree)},
		{name: "aDeathBeforeALaterDayFitsAnEarlierDeath", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(death, subjectRole, datevaluetest.Before(2021, 1, 11))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.YMD(1932, 7, 27)), ev(death, subjectRole, datevaluetest.YMD(2001, 3, 31))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "aDeathBeforeTheOtherBirthConflicts", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(death, subjectRole, datevaluetest.Before(1900, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.YMD(1932, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(2001, 0, 0))},
			sim: 2.5 / 22.5, ev: -17.5, check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "aBirthAfterTheOtherDeathConflicts", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.After(2010, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.YMD(1932, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(2001, 0, 0))},
			sim: 2.5 / 22.5, ev: -17.5, check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "aBirthAfterAnEarlierYearFits", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.After(1900, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.YMD(1932, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(2001, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "aClosedDateStillBoundsAnOpenSpan", kind: vocabkey.Person, entity: james,
			life: []subjectprofile.LifeEvent{
				ev(vocabkey.EventType("baptism"), subjectRole, datevaluetest.YMD(1900, 0, 0)),
				ev(death, subjectRole, datevaluetest.Before(1980, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.YMD(1850, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(1870, 0, 0))},
			sim: 2.5 / 22.5, ev: -17.5, check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "aBirthBeforeAndADeathAfterStillFit", kind: vocabkey.Person, entity: james,
			life: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.Before(1900, 0, 0)), ev(death, subjectRole, datevaluetest.After(1950, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.YMD(1890, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(1960, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "twoOpenResidencesStillFit", kind: vocabkey.Person, entity: james,
			life: []subjectprofile.LifeEvent{
				ev(vocabkey.EventType("residence"), subjectRole, datevaluetest.After(1950, 0, 0)),
				ev(vocabkey.EventType("residence"), subjectRole, datevaluetest.Before(1900, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{
				ev(birth, subjectRole, datevaluetest.YMD(1890, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(1960, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "W1 a death before a year only fits a life ending earlier", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(death, subjectRole, datevaluetest.Before(1840, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1600, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(1650, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "W2 a birth after a year only fits a life beginning later", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.After(1800, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1950, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(2000, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "W3 a birth and a death after a year fit a residence within a life of the birth", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1900, 0, 0)), ev(death, subjectRole, datevaluetest.After(1950, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(residence, subjectRole, datevaluetest.YMD(1990, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "W4 a birth before and a death after still fit", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.Before(1900, 0, 0)), ev(death, subjectRole, datevaluetest.After(1950, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1890, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(1960, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "W5 a death before a year conflicts with a later residence", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(death, subjectRole, datevaluetest.Before(1840, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(residence, subjectRole, datevaluetest.YMD(1900, 0, 0))},
			sim: 2.5 / 22.5, ev: -17.5, check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "W6 a birth after a year conflicts with an earlier death", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.After(1800, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(death, subjectRole, datevaluetest.YMD(1790, 0, 0))},
			sim: 2.5 / 22.5, ev: -17.5, check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "W7 a residence beyond the longest life from the birth conflicts", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1900, 0, 0)), ev(death, subjectRole, datevaluetest.After(1950, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(residence, subjectRole, datevaluetest.YMD(2020, 0, 0))},
			sim: 2.5 / 22.5, ev: -17.5, check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "W8 a baptism and a residence before a year conflict with a death too early", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(vocabkey.EventType("baptism"), subjectRole, datevaluetest.YMD(1900, 0, 0)), ev(residence, subjectRole, datevaluetest.Before(1850, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.YMD(1700, 0, 0)), ev(death, subjectRole, datevaluetest.YMD(1785, 0, 0))},
			sim: 2.5 / 22.5, ev: -17.5, check: lifespanResult(propertycompare.OutcomeConflict)},
		{name: "W9 a birth before a year only fits a death in the distant past", kind: vocabkey.Person, entity: james,
			life:    []subjectprofile.LifeEvent{ev(birth, subjectRole, datevaluetest.Before(1900, 0, 0))},
			subject: jamesSubject, subjectLife: []subjectprofile.LifeEvent{ev(death, subjectRole, datevaluetest.YMD(1700, 0, 0))},
			sim: 1, ev: 2.6, check: lifespanNot(propertycompare.OutcomeConflict)},
		{name: "AC-5.5 two negatives never meet", kind: vocabkey.Person, entity: denied,
			subject: observed{sex: {vcand(term("female"), true)}}, sim: 0, ev: 0,
			check: sexResult(propertycompare.OutcomeUnknown, propertycompare.CauseNoValues)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg := matchregistry.Default()
			metas := subjectcomparetest.Metas()
			e := tc.entity(t)
			e.Life = tc.life
			lifespans := reg.Subject.Lifespans[vocabkey.Person]
			a, _, err := subjectprofile.Build(e, metas, reg.Value, lifespans, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			b, _ := subjectprofile.ForSubject(subjectprofile.Subject{Values: tc.subject, Life: tc.subjectLife}, lifespans)
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
