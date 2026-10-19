package matchregistry

import (
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// defaultValue is the value section. Each number says where it came from.
func defaultValue() valuecompare.Rules {
	return valuecompare.Rules{
		Words: valuecompare.WordRules{
			// the reconciler's old spelling floor: Robins ~ Robbins (0.86) counts,
			// Mary ~ Mark (0.75) does not.
			FuzzyFloor: 0.8,
			// the reconciler's old short-word minimum: Ann ~ Anne, Jon ~ John.
			ShortVariantMinLength: 3,
			// match DefaultWordRules.InitialCredit: an initial against a word
			// it begins.
			InitialCredit: 0.5,
			// match.AliasCredit: a nickname or abbreviation table hit.
			AliasCredit: 0.98,
			// match.MaxExactPairing: longer word lists pair greedily.
			MaxExactPairing: 12,
			// match foldOnly: a match that exists only after folding.
			FoldCredit: 0.95,
			// An estimate: a word this long is a run-on, not a spelling variant.
			MaxSpellingLength: 32,
			// match.SpacedParticles. A typed surname prefix (van, de) is a
			// part role, not a particle.
			Particles: []string{"o", "mc", "mac", "st", "d", "fitz"},
			// match.LetterFolds: letters that survive accent stripping.
			Folds: map[rune]string{
				'ß': "ss",
				'æ': "ae",
				'œ': "oe",
				'ø': "o",
				'ł': "l",
				'đ': "d",
				'þ': "th",
				'ı': "i",
			},
		},
		Names: valuecompare.NameRules{
			// carried from graphalign's name scale floor.
			NearMinimum: 0.75,
			// match.DefaultNames affinities.
			UntypedAffinity: 0.8,
			NickAffinity:    0.9,
			CrossRole:       0.5,
			// estimate (S11-03 user call 2026-10-17): far above any real name;
			// bounds BestAssignment's exact route.
			MaxWords: 32,
			// An estimate: a Jr. on one side only says about as much as a
			// middle name (OtherGiven 0.5).
			Generation: 0.5,
			// match.WesternNamePattern: given names then surname, with
			// GEDCOM-style part types. Every product part type
			// (namevalue.PartTypes) must have a role.
			Western: valuecompare.NamePattern{
				PartRoles: map[string]valuecompare.NameRole{
					namevalue.PartTypeSurname:       valuecompare.RoleFamily,
					namevalue.PartTypeGiven:         valuecompare.RoleGiven,
					namevalue.PartTypeNick:          valuecompare.RoleNick,
					namevalue.PartTypeSuffix:        valuecompare.RoleGeneration,
					namevalue.PartTypePrefix:        valuecompare.RoleIgnored,
					namevalue.PartTypeSurnamePrefix: valuecompare.RoleIgnored,
					namevalue.PartTypeUndetermined:  valuecompare.RoleUntyped,
					"":                              valuecompare.RoleUntyped,
				},
				// What a matched word of each role earns and an unmatched one
				// costs: match.WesternNamePattern.Weights.
				Weights: valuecompare.NameWeights{
					Family:     1.5,
					FirstGiven: 1,
					OtherGiven: 0.5, // middle names and initials
					Nick:       0.3,
					Untyped:    1,
				},
			},
		},
		// match.DefaultText.Partial; MaxWords is an estimate, well above a toponym or
		// a short remark.
		Text: valuecompare.TextRules{Partial: 0.7, MaxWords: 64},
		// match.DefaultIntegers.Tolerance: any difference disagrees.
		Integers: valuecompare.IntegerRules{Tolerance: 0},
		Dates: valuecompare.DateRules{
			// DefaultDates.Tolerance: civil years apart that still resemble.
			Tolerance: 2,
			// carried from graphalign's date scale: no floor.
			NearMinimum: 0,
			// DefaultDates.MissingDay: same month, a day missing on one side.
			MissingDay: 0.9,
			// DefaultDates.MissingMonth: a month missing on one side.
			MissingMonth: 0.8,
			// DefaultDates.OtherDay: same month, different known days.
			OtherDay: 0.7,
			// DefaultDates.OtherMonth: different known months.
			OtherMonth: 0.6,
			// DefaultDates.YearsApartFrom: start of the fall-off across years.
			YearsApartFrom: 0.8,
			// DefaultDates.SpanOneYear: a span of a year or less.
			SpanOneYear: 0.8,
			// DefaultDates.SpanPerExtraYear: less for each extra year of span.
			SpanPerExtraYear: 0.05,
			// DefaultDates.SpanFloor: lowest span similarity; open bounds.
			SpanFloor: 0.2,
			// An estimate: the design's ABT rule, one unit of the most precise
			// part on each side.
			AboutUnits: 1,
		},
	}
}
