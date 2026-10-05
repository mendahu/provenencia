package resolve

// Spelling variants: when two normalized words are the same name spelled
// differently (Robins ~ Robbins, Jon ~ John). Shared by the name module,
// which lets majority outvote only spelling variants, and by core/match's
// word comparison, so reconciling and matching agree on what a misspelling is.

// SpellingFloor is the least similarity (1 − edits / longer length) for a
// spelling variant: Robins ~ Robbins (0.86) counts, Mary ~ Mark (0.75) does
// not. The default for core/match's FuzzyFloor.
const SpellingFloor = 0.8

// SpellingShortMin: words at least this long may differ by one added or
// dropped letter below SpellingFloor (Ann ~ Anne, Jon ~ John: 0.75). The
// default for core/match's ShortVariantMinLength.
const SpellingShortMin = 3

// SpellingSimilarity is x and y's similarity when they are spelling variants
// under floor and shortMin, else 0. Equal words are 1. A one-letter
// substitution in a short word is a different name (Mary ~ Mark). Callers
// normalize first; initials are not handled here.
func SpellingSimilarity(x, y string, floor float64, shortMin int) float64 {
	if x == y {
		return 1
	}
	if floor >= 1 {
		return 0
	}
	rx, ry := []rune(x), []rune(y)
	longest, shortest := len(rx), len(ry)
	if shortest > longest {
		longest, shortest = shortest, longest
	}
	if longest == 0 {
		return 0
	}
	edits := EditDistance(rx, ry)
	ratio := 1 - float64(edits)/float64(longest)
	if ratio >= floor {
		return ratio
	}
	if edits == 1 && longest == shortest+1 && shortest >= shortMin {
		return ratio
	}
	return 0
}

// EditDistance is the optimal-string-alignment distance: insertions,
// deletions, substitutions, and swaps of adjacent letters ("Robnis" ~
// "Robins") each cost one.
func EditDistance(a, b []rune) int {
	d := make([][]int, len(a)+1)
	for i := range d {
		d[i] = make([]int, len(b)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(a)][len(b)]
}
