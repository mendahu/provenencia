package match

import (
	"sort"
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
	aliases              []aliasSet
}

// aliasSet is one equivalence set and the roles that may use it.
type aliasSet struct {
	tokens map[string]bool
	roles  map[valuecompare.NameRole]bool
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
		aliases:       indexAliases(aliasTablesOf(r.Aliases)),
	}
}

func indexAliases(tables []valuecompare.AliasTable) []aliasSet {
	var sets []aliasSet
	for _, table := range tables {
		roles := make(map[valuecompare.NameRole]bool, len(table.Roles))
		for _, role := range table.Roles {
			roles[role] = true
		}
		for _, raw := range table.Sets {
			tokens := make(map[string]bool, len(raw))
			for _, token := range raw {
				if token != "" {
					tokens[token] = true
				}
			}
			if len(tokens) == 0 {
				continue
			}
			sets = append(sets, aliasSet{tokens: tokens, roles: roles})
		}
	}
	return sets
}

func aliasHit(x, y string, roleX, roleY valuecompare.NameRole, sets []aliasSet) bool {
	for _, set := range sets {
		if set.tokens[x] && set.tokens[y] && set.roles[roleX] && set.roles[roleY] {
			return true
		}
	}
	return false
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
	if aliasHit(x, y, roleX, roleY, r.aliases) {
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
// columns. Exact, so it does not depend on which side is the row. A pair
// scores only when it is positive, so a word with nothing in common stays
// unpaired.
func bestPairing(score [][]float64) float64 {
	total, _, _ := bestAssignment(score)
	return total
}

// bestAssignment is bestPairing plus which cells took part. usedRow and
// usedCol are in the score matrix's own orientation.
func bestAssignment(score [][]float64) (total float64, usedRow, usedCol []bool) {
	rows := len(score)
	if rows == 0 || len(score[0]) == 0 {
		return 0, make([]bool, rows), nil
	}
	cols := len(score[0])
	usedRow = make([]bool, rows)
	usedCol = make([]bool, cols)
	transposed := false
	at := func(i, j int) float64 { return score[i][j] }
	if rows < cols {
		transposed = true
		rows, cols = cols, rows
		at = func(i, j int) float64 { return score[j][i] }
	}
	rowUsed := make([]bool, rows)
	colUsed := make([]bool, cols)
	if cols > MaxExactPairing {
		total = greedyAssign(rows, cols, at, rowUsed, colUsed)
	} else {
		total = exactAssign(rows, cols, at, rowUsed, colUsed)
	}
	if transposed {
		for i := range usedRow {
			usedRow[i] = colUsed[i]
		}
		for j := range usedCol {
			usedCol[j] = rowUsed[j]
		}
		return total, usedRow, usedCol
	}
	return total, rowUsed, colUsed
}

func exactAssign(rows, cols int, at func(i, j int) float64, rowUsed, colUsed []bool) float64 {
	type step struct {
		col  int
		mask int
	}
	best := make([]float64, 1<<cols)
	steps := make([][]step, rows)
	for i := 0; i < rows; i++ {
		next := append([]float64(nil), best...)
		ch := make([]step, 1<<cols)
		for mask := range ch {
			ch[mask] = step{col: -1, mask: mask}
		}
		for mask, v := range best {
			for j := 0; j < cols; j++ {
				s := at(i, j)
				if mask&(1<<j) != 0 || s <= 0 {
					continue
				}
				nm := mask | (1 << j)
				if t := v + s; t > next[nm] {
					next[nm] = t
					ch[nm] = step{col: j, mask: mask}
				}
			}
		}
		best = next
		steps[i] = ch
	}
	topMask := 0
	var top float64
	for mask, v := range best {
		if v > top {
			top = v
			topMask = mask
		}
	}
	mask := topMask
	for i := rows - 1; i >= 0; i-- {
		ch := steps[i][mask]
		if ch.col >= 0 {
			rowUsed[i] = true
			colUsed[ch.col] = true
		}
		mask = ch.mask
	}
	return top
}

func greedyAssign(rows, cols int, at func(i, j int) float64, rowUsed, colUsed []bool) float64 {
	type pair struct {
		i, j int
		s    float64
	}
	var pairs []pair
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if s := at(i, j); s > 0 {
				pairs = append(pairs, pair{i, j, s})
			}
		}
	}
	sort.SliceStable(pairs, func(a, b int) bool { return pairs[a].s > pairs[b].s })
	var total float64
	for _, p := range pairs {
		if !rowUsed[p.i] && !colUsed[p.j] {
			rowUsed[p.i], colUsed[p.j] = true, true
			total += p.s
		}
	}
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
