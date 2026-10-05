package match

import (
	"math"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/autoreconcile"
)

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
}

// wordRules are WordRules resolved against the registry.
type wordRules struct {
	floor, initialCredit float64
	shortMin             int
	initials             bool // whether lone letters match as initials
}

func (r WordRules) resolve(initials bool) wordRules {
	return wordRules{
		floor:         setting(r.FuzzyFloor, *DefaultWordRules.FuzzyFloor),
		initialCredit: setting(r.InitialCredit, *DefaultWordRules.InitialCredit),
		shortMin:      setting(r.ShortVariantMinLength, *DefaultWordRules.ShortVariantMinLength),
		initials:      initials,
	}
}

func splitWords(s string) []string {
	n := autoreconcile.NormalizeForm(s)
	if n == "" {
		return nil
	}
	return strings.Split(n, " ")
}

// compareForms is 1 for the same normalized form, else partial × the share
// of words the two forms have in common (wordDice). Empty forms are not
// comparable.
func compareForms(a, b string, partial float64, r wordRules) (float64, bool) {
	wa, wb := splitWords(a), splitWords(b)
	if len(wa) == 0 || len(wb) == 0 {
		return 0, false
	}
	if strings.Join(wa, " ") == strings.Join(wb, " ") {
		return 1, true
	}
	return partial * wordDice(wa, wb, r), true
}

// wordSimilarity: 1 for the same word, InitialCredit for an initial and a
// word it begins (when initials count), else the spelling-variant ratio
// (1 − edits / longer length, an adjacent swap being one edit) when it
// reaches FuzzyFloor — or, for words of ShortVariantMinLength or more, when
// they differ by one added or dropped letter (Ann ~ Anne, Jon ~ John: 0.75),
// since one edit is a larger share of a short word. A one-letter substitution in a short word is a
// different name (Mary ~ Mark). An initial is a lone cased letter ("J"); a
// lone character in an uncased script (蒋, 王) is a whole word.
func wordSimilarity(x, y string, r wordRules) float64 {
	if x == y {
		return 1
	}
	if autoreconcile.IsInitial(x) || autoreconcile.IsInitial(y) {
		if r.initials && []rune(x)[0] == []rune(y)[0] {
			return r.initialCredit
		}
		return 0
	}
	return autoreconcile.SpellingSimilarity(x, y, r.floor, r.shortMin)
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
			score[i][j] = wordSimilarity(x, y, r)
		}
	}
	return 2 * bestPairing(score) / float64(len(wa)+len(wb))
}

// bestPairing is the largest total over one-to-one pairings of rows with
// columns. Exact, so it does not depend on which side is the row.
func bestPairing(score [][]float64) float64 {
	if len(score) == 0 || len(score[0]) == 0 {
		return 0
	}
	rows, cols := len(score), len(score[0])
	at := func(i, j int) float64 { return score[i][j] }
	if rows < cols {
		rows, cols = cols, rows
		at = func(i, j int) float64 { return score[j][i] }
	}
	if cols > MaxExactPairing {
		return greedyPairing(rows, cols, at)
	}
	// best[mask] = best total with the shorter side's words in mask used.
	best := make([]float64, 1<<cols)
	for i := 0; i < rows; i++ {
		next := append([]float64(nil), best...)
		for mask, v := range best {
			for j := 0; j < cols; j++ {
				if s := at(i, j); mask&(1<<j) == 0 && s > 0 {
					if t := v + s; t > next[mask|1<<j] {
						next[mask|1<<j] = t
					}
				}
			}
		}
		best = next
	}
	var top float64
	for _, v := range best {
		top = math.Max(top, v)
	}
	return top
}

func greedyPairing(rows, cols int, at func(i, j int) float64) float64 {
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
	usedI, usedJ := map[int]bool{}, map[int]bool{}
	var total float64
	for _, p := range pairs {
		if !usedI[p.i] && !usedJ[p.j] {
			usedI[p.i], usedJ[p.j] = true, true
			total += p.s
		}
	}
	return total
}
