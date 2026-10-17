package valuecompare

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Word helpers: normalizing, splitting, folding and spelling variants. Shared
// by the text comparer and, until the callers move, by core/match and the
// reconciler's name module, so each agrees on what a word and a misspelling is.

// NormalizeForm ignores case, punctuation, and whitespace differences: the
// name sort key, and the text core/match tokenizes. Dashes
// and slashes separate words ("Smith-Jones" reads as "smith jones"); other
// punctuation is dropped, so an apostrophe joins ("O'Brien" reads as
// "obrien", the same as "OBrien").
func NormalizeForm(form string) string {
	var b strings.Builder
	space := false
	for _, r := range form {
		switch {
		case unicode.Is(unicode.Pd, r) || r == '/':
			space = b.Len() > 0
			continue
		case unicode.IsPunct(r):
			continue
		case unicode.IsSpace(r):
			space = b.Len() > 0
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// IsInitial reports whether a word is an initial: a lone cased letter ("J").
// A lone character in an uncased script (蒋, 王) is a whole word.
func IsInitial(word string) bool {
	r := []rune(word)
	return len(r) == 1 && unicode.ToUpper(r[0]) != unicode.ToLower(r[0])
}

// SplitWords is s as comparison tokens: split on whitespace, dashes and
// slashes, each token normalized, and a particle rejoined with the word that
// follows it. particles is the set of particle tokens.
func SplitWords(s string, particles map[string]bool) []string {
	raw := rawTokens(s)
	var out []string
	for i := 0; i < len(raw); i++ {
		if i+1 < len(raw) && spacedParticle(raw[i], particles) {
			if n := NormalizeForm(raw[i] + raw[i+1]); n != "" {
				out = append(out, n)
			}
			i++
			continue
		}
		if n := NormalizeForm(raw[i]); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// rawTokens splits on whitespace and on the dash and slash breaks
// NormalizeForm uses, and keeps other punctuation so a period initial
// can be told from a particle.
func rawTokens(s string) []string {
	var b strings.Builder
	var out []string
	flush := func() {
		if b.Len() == 0 {
			return
		}
		out = append(out, b.String())
		b.Reset()
	}
	for _, r := range s {
		if unicode.Is(unicode.Pd, r) || r == '/' || unicode.IsSpace(r) {
			flush()
			continue
		}
		b.WriteRune(r)
	}
	flush()
	return out
}

func spacedParticle(raw string, particles map[string]bool) bool {
	if periodInitial(raw) {
		return false
	}
	return particles[NormalizeForm(raw)]
}

// periodInitial is a lone cased letter written with a period ("O.").
func periodInitial(raw string) bool {
	if !strings.ContainsRune(raw, '.') {
		return false
	}
	var letters []rune
	for _, r := range raw {
		if unicode.IsLetter(r) {
			letters = append(letters, r)
		}
	}
	return len(letters) == 1 && unicode.ToUpper(letters[0]) != unicode.ToLower(letters[0])
}

// FoldWord drops accents, then applies folds. The word is already the
// normalized comparison token. A nil or empty map still strips accents.
func FoldWord(s string, folds map[rune]string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if repl, ok := folds[r]; ok {
			b.WriteString(repl)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

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
			pi, pj := i-1, j-1
			if pi > 0 && pj > 0 && a[pi] == b[pj-1] && a[pi-1] == b[pj] {
				d[i][j] = min(d[i][j], d[pi-1][pj-1]+1)
			}
		}
	}
	return d[len(a)][len(b)]
}

// wordSimilarity is how two comparison tokens relate, in this order: byte-equal
// tokens are Same, 1 (exact); an initial against anything else is no pair
// (text has no initials); tokens equal after folding are Equivalent at
// FoldCredit (fold); a word over MaxSpellingLength runes has no spelling pair; a spelling variant is Near at its ratio (spelling); else
// no pair, 0. A relation of Different means no pair.
func wordSimilarity(x, y string, w WordRules) (float64, Relation, Cause) {
	if x == y {
		return 1, Same, CauseExact
	}
	if IsInitial(x) || IsInitial(y) {
		return 0, Different, CauseNoMatch
	}
	fx, fy := FoldWord(x, w.Folds), FoldWord(y, w.Folds)
	if fx == fy {
		return w.FoldCredit, Equivalent, CauseFold
	}
	// The edit-distance table is the product of the lengths: bound it.
	if utf8.RuneCountInString(fx) > w.MaxSpellingLength || utf8.RuneCountInString(fy) > w.MaxSpellingLength {
		return 0, Different, CauseNoMatch
	}
	if sim := SpellingSimilarity(fx, fy, w.FuzzyFloor, w.ShortVariantMinLength); sim > 0 {
		return sim, Near, CauseSpelling
	}
	return 0, Different, CauseNoMatch
}

// wordPair is one aligned pair, by index into each side's words.
type wordPair struct {
	i, j int
	sim  float64
	rel  Relation
	why  Cause
}

// alignWords pairs the words of a with the words of b in order: the
// order-preserving pairing with the highest total word similarity, a pair
// counting only above 0. Among equal totals it takes the most pairs; then,
// walking back from the last cell, the diagonal before skipping a word of a,
// and skipping a word of a before skipping a word of b. It returns the pairs in
// a's order and the indexes left unpaired on each side.
func alignWords(a, b []string, w WordRules) (pairs []wordPair, onlyA, onlyB []int) {
	type cell struct {
		total  float64
		pairs  int
		choice int
	}
	const (
		diagonal = iota
		skipA
		skipB
	)
	better := func(x, y cell) bool {
		return x.total > y.total || (x.total == y.total && x.pairs > y.pairs)
	}
	sims := make([][]wordPair, len(a))
	for i := range a {
		sims[i] = make([]wordPair, len(b))
		for j := range b {
			sim, rel, why := wordSimilarity(a[i], b[j], w)
			sims[i][j] = wordPair{i: i, j: j, sim: sim, rel: rel, why: why}
		}
	}
	grid := make([][]cell, len(a)+1)
	for i := range grid {
		grid[i] = make([]cell, len(b)+1)
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			best := cell{total: -1}
			if p := sims[i-1][j-1]; p.sim > 0 {
				diag := grid[i-1][j-1]
				best = cell{total: diag.total + p.sim, pairs: diag.pairs + 1, choice: diagonal}
			}
			if up := grid[i-1][j]; better(cell{total: up.total, pairs: up.pairs}, best) {
				best = cell{total: up.total, pairs: up.pairs, choice: skipA}
			}
			if left := grid[i][j-1]; better(cell{total: left.total, pairs: left.pairs}, best) {
				best = cell{total: left.total, pairs: left.pairs, choice: skipB}
			}
			grid[i][j] = best
		}
	}
	i, j := len(a), len(b)
	for i > 0 && j > 0 {
		switch grid[i][j].choice {
		case diagonal:
			pairs = append(pairs, sims[i-1][j-1])
			i--
			j--
		case skipA:
			onlyA = append(onlyA, i-1)
			i--
		default:
			onlyB = append(onlyB, j-1)
			j--
		}
	}
	for ; i > 0; i-- {
		onlyA = append(onlyA, i-1)
	}
	for ; j > 0; j-- {
		onlyB = append(onlyB, j-1)
	}
	slices.Reverse(pairs)
	slices.Reverse(onlyA)
	slices.Reverse(onlyB)
	return pairs, onlyA, onlyB
}
