package matchregistry

import (
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// defaultSubject is layer 3's shipped section: what each property counts when
// two Subjects of one type are compared.
func defaultSubject() subjectcompare.Rules {
	return subjectcompare.Rules{
		Weights: map[vocabkey.SubjectType]map[vocabkey.PropertyKey]subjectcompare.Weights{
			vocabkey.Person: {
				vocabkey.PropertyName:       {Agree: 2.5, Conflict: 3},  // carried from graphalign's name scale
				vocabkey.PropertySexAtBirth: {Agree: 0.6, Conflict: 1},  // carried from graphalign's sex / gender scale; pools gender
				vocabkey.PropertyLifespan:   {Agree: 0.1, Conflict: 20}, // estimate: near zero, the dates count through their segments; impossible together never recovers
			},
			vocabkey.Event: {
				vocabkey.PropertyEventType: {Agree: 0.5, Conflict: 20}, // estimate: generic alone; a birth is never a death
				vocabkey.PropertyDate:      {Agree: 2, Conflict: 3},    // estimate: the cold-start ln(0.85 / 0.1) S11-24 removed, rounded; carried from graphalign ConflictPenalty
			},
			vocabkey.Place: {
				vocabkey.PropertyToponym: {Agree: 3, Conflict: 3}, // estimate: the cold-start + toponym bonus S11-24 removed, rounded; carried from graphalign ConflictPenalty
			},
		},
		Pools: map[vocabkey.SubjectType]map[vocabkey.PropertyKey]subjectcompare.Pool{
			vocabkey.Person: {vocabkey.PropertySexAtBirth: {
				With: []vocabkey.PropertyKey{vocabkey.PropertyGender},
				// decided (user, 2026-10-21): a gender reads as the sex at birth it
				// names; one with no entry (non_binary) is left out, never a conflict
				Terms: map[vocabkey.PropertyKey]map[string]string{
					vocabkey.PropertyGender: {vocabkey.GenderMan: vocabkey.SexMale, vocabkey.GenderWoman: vocabkey.SexFemale},
				},
			}},
		},
		Periods: map[vocabkey.SubjectType]map[vocabkey.PropertyKey]subjectcompare.Period{
			vocabkey.Event: {vocabkey.PropertyDate: {Start: vocabkey.PropertyStartDate, End: vocabkey.PropertyEndDate}},
		},
		Lifespans: map[vocabkey.SubjectType]map[vocabkey.PropertyKey]subjectcompare.Lifespan{
			vocabkey.Person: {vocabkey.PropertyLifespan: {
				Span:       vocabkey.PropertyLifespanSpan,
				MaxYears:   110, // estimate: longest plausible human life
				SlackYears: 10,  // estimate: forgiving margin for misdated records
				Starts:     []subjectcompare.EventRole{{EventType: vocabkey.EventBirth, Role: vocabkey.RoleSubject}},
				Ends: []subjectcompare.EventRole{
					{EventType: vocabkey.EventDeath, Role: vocabkey.RoleSubject},
					{EventType: vocabkey.EventBurial, Role: vocabkey.RoleSubject},
				},
			}},
		},
	}
}
