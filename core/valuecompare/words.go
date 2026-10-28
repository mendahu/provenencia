package valuecompare

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Word helpers: normalizing, splitting, folding and spelling variants. Shared
// by the text and name comparers and by the reconciler's name module
// (`SplitWords`, `SpellingSimilarity`), so each agrees on what a word and a
// misspelling is.

// NormalizeForm ignores case, punctuation, and whitespace differences: the
// name sort key and the text that text value compare reads. Dashes
// and slashes separate words ("Smith-Jones" reads as "smith jones"); other
// punctuation is dropped, so an apostrophe joins ("O'Brien" reads as
// "obrien", the same as "OBrien").
func NormalizeForm(form string) string {
	var b strings.Builder
	space := false
	for _, r := range form {
		keep, brk, ok := wordRune(r)
		switch {
		case brk:
			space = b.Len() > 0
		case ok:
			if space {
				b.WriteByte(' ')
				space = false
			}
			b.WriteRune(keep)
		}
	}
	return b.String()
}

// wordRune is the one per-rune rule for what a word is: a dash or slash or
// whitespace breaks a word (brk), other punctuation is dropped (neither brk
// nor ok), and any other rune is kept, lowercased (ok). NormalizeForm and the
// one-pass key read it, so they cannot disagree.
func wordRune(r rune) (keep rune, brk, ok bool) {
	if r < utf8.RuneSelf {
		switch {
		case r == '/' || r == '-' || r == ' ' || (r >= '\t' && r <= '\r'):
			return 0, true, false
		case 'a' <= r && r <= 'z' || '0' <= r && r <= '9':
			return r, false, true
		}
	}
	switch {
	case unicode.Is(unicode.Pd, r) || r == '/' || unicode.IsSpace(r):
		return 0, true, false
	case unicode.IsPunct(r):
		return 0, false, false
	}
	return unicode.ToLower(r), false, true
}

// IsInitial reports whether a word is an initial: a lone cased letter ("J").
// A lone character in an uncased script (蒋, 王) is a whole word.
func IsInitial(word string) bool {
	r := []rune(word)
	return len(r) == 1 && unicode.ToUpper(r[0]) != unicode.ToLower(r[0])
}

// PrefixRuleRole says whether a name part of this role takes part in the
// prefix rule: given, nickname and untyped words, never a surname. The
// shortlist, the comparer and the reconciler's fold read it.
func PrefixRuleRole(role NameRole) bool {
	return slices.Contains(AliasRoles(), role)
}

// prefixWord says whether a name word takes part in the prefix rule: the
// shortlist's prefix keys and the comparer's prefix hit read it, so the two
// never disagree.
func prefixWord(w nameWord, r Rules) bool {
	if r.Names.PrefixMinLength < 1 || !PrefixRuleRole(w.role) || IsInitial(w.text) {
		return false
	}
	n := utf8.RuneCountInString(w.fold)
	return n >= r.Names.PrefixMinLength && n <= r.Words.MaxSpellingLength
}

// prefixesOf are the strict prefixes of a folded word from PrefixMinLength
// runes up to one rune short of the word, cut on runes.
func prefixesOf(fold string, r Rules) []string {
	runes := []rune(fold)
	var out []string
	for k := r.Names.PrefixMinLength; k < len(runes); k++ {
		out = append(out, string(runes[:k]))
	}
	return out
}

// SplitWords is s as comparison tokens: split on whitespace, dashes and
// slashes, each token normalized, and a particle (w.Particles) rejoined with
// the word that follows it. It is the one split every comparer and the
// reconciler share.
func SplitWords(s string, w WordRules) []string {
	isParticle := func(token string) bool { return slices.Contains(w.Particles, token) }
	raw := rawTokens(s)
	var out []string
	for i := 0; i < len(raw); i++ {
		if i+1 < len(raw) && spacedParticle(raw[i], isParticle) {
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

func spacedParticle(raw string, isParticle func(token string) bool) bool {
	if periodInitial(raw) {
		return false
	}
	return isParticle(NormalizeForm(raw))
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
	b.Grow(len(s))
	for _, r := range s {
		appendFolded(&b, r, folds)
	}
	return b.String()
}

// appendFolded writes r as FoldWord folds it: its full decomposition (the
// whole NFD of the rune where the decomposition is algorithmic, as for Hangul)
// without the accents, each remaining rune replaced by its fold. It is the one
// per-rune fold rule, shared by FoldWord and the one-pass key.
func appendFolded(b *strings.Builder, r rune, folds map[rune]string) {
	if r < utf8.RuneSelf {
		if repl, ok := folds[r]; ok {
			b.WriteString(repl)
			return
		}
		b.WriteByte(byte(r))
		return
	}
	var rbuf [utf8.UTFMax]byte
	enc := utf8.AppendRune(rbuf[:0], r)
	dec := norm.NFD.Properties(enc).Decomposition()
	if dec == nil {
		dec = norm.NFD.AppendString(nil, string(r))
	}
	for _, d := range string(dec) {
		if d >= utf8.RuneSelf && unicode.Is(unicode.Mn, d) {
			continue
		}
		if repl, ok := folds[d]; ok {
			b.WriteString(repl)
			continue
		}
		b.WriteRune(d)
	}
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

// wordOptions are what a caller adds to the word table: names read initials,
// alias tables and prefixes, text none.
type wordOptions struct {
	initials     bool       // whether lone letters match as initials
	aliases      AliasIndex // abbreviation and nickname tables
	roleX, roleY NameRole   // the roles an alias hit must allow
	prefixCredit float64    // what a prefix hit earns (Names.PrefixCredit); 0 when either word is not a prefix word, so there is no prefix rung
}

// wordSimilarity is how two comparison tokens relate, in this order: byte-equal
// tokens are Same, 1 (exact); an initial against anything else is a pair at
// InitialCredit, Contains (initial), when initials count and the first letters
// agree, else no pair; two tokens in one alias set whose tables allow both
// roles are Equivalent at AliasCredit (abbreviation for an abbreviation table,
// alias for any other); tokens equal after folding are Equivalent at
// FoldCredit (fold); a word over MaxSpellingLength runes has no spelling pair;
// a spelling variant is Near at its ratio (spelling); a word that is the start
// of the other after folding, when the options carry a prefix credit, is a pair
// at that credit, Contains (prefix); else no pair, 0. broader is the side of
// the initial or the prefix on a Contains pair, else SideNone. A relation of
// Different means no pair. fx and fy are x and y folded by the caller, once per
// word.
func wordSimilarity(x, y, fx, fy string, w WordRules, o wordOptions) (float64, Relation, Cause, Side) {
	if x == y {
		return 1, Same, CauseExact, SideNone
	}
	if IsInitial(x) || IsInitial(y) {
		if o.initials && []rune(x)[0] == []rune(y)[0] {
			if IsInitial(x) {
				return w.InitialCredit, Contains, CauseInitial, SideA
			}
			return w.InitialCredit, Contains, CauseInitial, SideB
		}
		return 0, Different, CauseNoMatch, SideNone
	}
	if kind, ok := o.aliases.Hit(x, y, o.roleX, o.roleY); ok {
		if kind == AliasAbbreviation {
			return w.AliasCredit, Equivalent, CauseAbbreviation, SideNone
		}
		return w.AliasCredit, Equivalent, CauseAlias, SideNone
	}
	if fx == fy {
		return w.FoldCredit, Equivalent, CauseFold, SideNone
	}
	// The edit-distance table is the product of the lengths: bound it.
	if utf8.RuneCountInString(fx) > w.MaxSpellingLength || utf8.RuneCountInString(fy) > w.MaxSpellingLength {
		return 0, Different, CauseNoMatch, SideNone
	}
	if sim := SpellingSimilarity(fx, fy, w.FuzzyFloor, w.ShortVariantMinLength); sim > 0 {
		return sim, Near, CauseSpelling, SideNone
	}
	if o.prefixCredit > 0 {
		if strings.HasPrefix(fy, fx) {
			return o.prefixCredit, Contains, CausePrefix, SideA
		}
		if strings.HasPrefix(fx, fy) {
			return o.prefixCredit, Contains, CausePrefix, SideB
		}
	}
	return 0, Different, CauseNoMatch, SideNone
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
// a's order and the indexes left unpaired on each side, and every cell's
// similarity for a caller that needs another pair. fa and fb are a and b folded.
func alignWords(a, b, fa, fb []string, w WordRules) (pairs []wordPair, onlyA, onlyB []int, sims [][]wordPair) {
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
	sims = make([][]wordPair, len(a))
	for i := range a {
		sims[i] = make([]wordPair, len(b))
		for j := range b {
			sim, rel, why, _ := wordSimilarity(a[i], b[j], fa[i], fb[j], w, wordOptions{})
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
	return pairs, onlyA, onlyB, sims
}
