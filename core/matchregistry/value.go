package matchregistry

import "github.com/mendahu/provenencia/core/valuecompare"

// defaultValue is the value section. Each number says where it came from.
func defaultValue() valuecompare.Rules {
	return valuecompare.Rules{
		Words: valuecompare.WordRules{
			// autoreconcile.SpellingFloor: Robins ~ Robbins (0.86) counts,
			// Mary ~ Mark (0.75) does not.
			FuzzyFloor: 0.8,
			// autoreconcile.SpellingShortMin: Ann ~ Anne, Jon ~ John.
			ShortVariantMinLength: 3,
			// match foldOnly: a match that exists only after folding.
			FoldCredit: 0.95,
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
		// match.DefaultText.Partial.
		Text: valuecompare.TextRules{Partial: 0.7},
		// match.DefaultIntegers.Tolerance: any difference disagrees.
		Integers: valuecompare.IntegerRules{Tolerance: 0},
	}
}
