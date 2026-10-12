package match

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"golang.org/x/text/unicode/norm"
)

// foldOnly is a match that exists only after accent and letter folding.
// An exact word still scores 1.
const foldOnly = 0.95

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
	Aliases *[]AliasTable
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
	roles  map[NameRole]bool
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

func indexAliases(tables []AliasTable) []aliasSet {
	var sets []aliasSet
	for _, table := range tables {
		roles := make(map[NameRole]bool, len(table.Roles))
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

func aliasHit(x, y string, roleX, roleY NameRole, sets []aliasSet) bool {
	for _, set := range sets {
		if set.tokens[x] && set.tokens[y] && set.roles[roleX] && set.roles[roleY] {
			return true
		}
	}
	return false
}

func splitWords(s string, particles map[string]bool) []string {
	raw := rawTokens(s)
	var out []string
	for i := 0; i < len(raw); i++ {
		if i+1 < len(raw) && spacedParticle(raw[i], particles) {
			if n := autoreconcile.NormalizeForm(raw[i] + raw[i+1]); n != "" {
				out = append(out, n)
			}
			i++
			continue
		}
		if n := autoreconcile.NormalizeForm(raw[i]); n != "" {
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
	return particles[autoreconcile.NormalizeForm(raw)]
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

// foldWord drops accents, then applies folds. The word is already the
// normalized comparison token. A nil or empty map still strips accents.
func foldWord(s string, folds map[rune]string) string {
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
	wa, wb := splitWords(a, r.particles), splitWords(b, r.particles)
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
func wordSimilarity(x, y string, roleX, roleY NameRole, r wordRules) float64 {
	if x == y {
		return 1
	}
	if autoreconcile.IsInitial(x) || autoreconcile.IsInitial(y) {
		if r.initials && []rune(x)[0] == []rune(y)[0] {
			return r.initialCredit
		}
		return 0
	}
	if aliasHit(x, y, roleX, roleY, r.aliases) {
		return AliasCredit
	}
	fx, fy := foldWord(x, r.folds), foldWord(y, r.folds)
	if fx == fy {
		return foldOnly
	}
	return autoreconcile.SpellingSimilarity(fx, fy, r.floor, r.shortMin)
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
