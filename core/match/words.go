package match

import (
	"strings"

	"github.com/mendahu/provenencia/core/valuecompare"
)

// foldOnly is a match that exists only after accent and letter folding.
// An exact word still scores 1.
var foldOnly = value.Words.FoldCredit

// Word-level comparison shared by NameComparer and TextComparer: splitting,
// word similarity (initials, spelling variants), and best one-to-one pairing.

// WordRules are how two words compare. Nil settings take DefaultWordRules.
type WordRules struct {
	// FuzzyFloor is the least edit-distance ratio for a spelling variant,
	// 0…1; 1 turns fuzzy matching off (short-word variants included).
	FuzzyFloor *float64
	// InitialCredit is an initial against a word it begins ("J." ~ "James").
	InitialCredit *float64
	// ShortVariantMinLength: words at least this long may differ by one
	// added or dropped letter below FuzzyFloor (Ann ~ Anne).
	ShortVariantMinLength *int
	// Particles rejoin with the following word. Nil uses SpacedParticles.
	// An empty list rejoins nothing. A period initial never rejoins.
	Particles *[]string
	// Folds replaces letters that do not decompose. Nil uses LetterFolds.
	// An empty map strips accents and substitutes nothing.
	Folds *map[rune]string
	// Aliases are abbreviation and nickname tables loaded from the catalog.
	// Nil and empty both hit nothing. Every table in the slice is consulted.
	Aliases *[]valuecompare.AliasTable
}

// wordRules are WordRules resolved against the registry.
type wordRules struct {
	floor, initialCredit float64
	shortMin             int
	initials             bool // whether lone letters match as initials
	particles            map[string]bool
	folds                map[rune]string
	aliases              valuecompare.AliasIndex
}

func (r WordRules) resolve(initials bool) wordRules {
	list := setting(r.Particles, *DefaultWordRules.Particles)
	particles := make(map[string]bool, len(list))
	for _, w := range list {
		particles[w] = true
	}
	return wordRules{
		floor:         setting(r.FuzzyFloor, *DefaultWordRules.FuzzyFloor),
		initialCredit: setting(r.InitialCredit, *DefaultWordRules.InitialCredit),
		shortMin:      setting(r.ShortVariantMinLength, *DefaultWordRules.ShortVariantMinLength),
		initials:      initials,
		particles:     particles,
		folds:         setting(r.Folds, *DefaultWordRules.Folds),
		aliases:       valuecompare.NewAliasIndex(aliasTablesOf(r.Aliases)),
	}
}

// TextResemblance is how close two free-text strings are for scoring when
// they are not the same value. 1 is the same normalized form. A single-word
// spelling variant is the edit-distance ratio: Rank's Partial cap on every
// non-identical string would leave a one-letter typo under the weak bar.
// Longer text that shares words is that cap times the word overlap. 0 means
// no resemblance.
func TextResemblance(a, b string) float64 {
	sim, ok := compareForms(a, b, setting(DefaultText.Partial, 0), DefaultText.Words.resolve(false))
	if !ok {
		return 0
	}
	return sim
}

// compareForms is 1 for the same normalized form, else partial × the share
// of words the two forms have in common (wordDice). Empty forms are not
// comparable.
func compareForms(a, b string, partial float64, r wordRules) (float64, bool) {
	wa, wb := valuecompare.SplitWords(a, r.particles), valuecompare.SplitWords(b, r.particles)
	if len(wa) == 0 || len(wb) == 0 {
		return 0, false
	}
	if strings.Join(wa, " ") == strings.Join(wb, " ") {
		return 1, true
	}
	// A one-word spelling variant is the edit-distance ratio. Capping it
	// would leave a one-letter typo under the promote weak bar.
	if len(wa) == 1 && len(wb) == 1 {
		return wordSimilarity(wa[0], wb[0], "", "", r), true
	}
	return partial * wordDice(wa, wb, r), true
}

// wordSimilarity: 1 for the same word, InitialCredit for an initial and a
// word it begins (when initials count), AliasCredit when both tokens share
// an alias set and both roles are allowed, else the spelling-variant ratio
// (1 − edits / longer length, an adjacent swap being one edit) when it
// reaches FuzzyFloor — or, for words of ShortVariantMinLength or more, when
// they differ by one added or dropped letter (Ann ~ Anne, Jon ~ John: 0.75),
// since one edit is a larger share of a short word. A one-letter substitution in a short word is a
// different name (Mary ~ Mark). An initial is a lone cased letter ("J"); a
// lone character in an uncased script (蒋, 王) is a whole word. Empty roles
// skip the alias tables, so free text does not consult them.
func wordSimilarity(x, y string, roleX, roleY valuecompare.NameRole, r wordRules) float64 {
	if x == y {
		return 1
	}
	if valuecompare.IsInitial(x) || valuecompare.IsInitial(y) {
		if r.initials && []rune(x)[0] == []rune(y)[0] {
			return r.initialCredit
		}
		return 0
	}
	if _, ok := r.aliases.Hit(x, y, roleX, roleY); ok {
		return AliasCredit
	}
	fx, fy := valuecompare.FoldWord(x, r.folds), valuecompare.FoldWord(y, r.folds)
	if fx == fy {
		return foldOnly
	}
	return valuecompare.SpellingSimilarity(fx, fy, r.floor, r.shortMin)
}

// wordDice is the Dice overlap of two unweighted word lists under their best
// one-to-one pairing (free text).
func wordDice(wa, wb []string, r wordRules) float64 {
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	score := make([][]float64, len(wa))
	for i, x := range wa {
		score[i] = make([]float64, len(wb))
		for j, y := range wb {
			score[i][j] = wordSimilarity(x, y, "", "", r)
		}
	}
	return 2 * bestPairing(score) / float64(len(wa)+len(wb))
}

// bestPairing is the largest total over one-to-one pairings of rows with
// columns (valuecompare.BestAssignment).
func bestPairing(score [][]float64) float64 {
	total, _ := valuecompare.BestAssignment(score, MaxExactPairing)
	return total
}

// aliasTablesOf is the list on the comparer. Nil and empty both hit nothing.
func aliasTablesOf(p *[]valuecompare.AliasTable) []valuecompare.AliasTable {
	if p == nil {
		return nil
	}
	out := make([]valuecompare.AliasTable, len(*p))
	copy(out, *p)
	return out
}
