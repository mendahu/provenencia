package match

import (
	"math"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/resolve"
)

// Defaults for NameComparer's structured comparison.
const (
	defaultSurnameShare    = 0.6
	defaultGivenOnlyFactor = 0.5
	defaultSuffixConflict  = 0.3
	unknownRoleCredit      = 0.5
	firstGivenShare        = 0.7
)

// nameRoles is a NameValue's typed parts grouped by what they mean for
// matching. Words are normalized (resolve.NormalizeForm) and split.
type nameRoles struct {
	surnames []string // surname parts; surname_prefix ("van") is not compared
	given    []string // given names, then initials, in part order
	nicks    []string // nicknames: alternatives to the first given name
	suffixes []string // Jr., Sr., III — differing suffixes suggest different people
}

func (r nameRoles) structured() bool { return len(r.surnames) > 0 || len(r.given) > 0 }

// rolesOf groups typed parts by role. prefix (titles) and undetermined or
// untyped parts carry no role.
func rolesOf(n *namevalues.Value) nameRoles {
	var r nameRoles
	for _, p := range n.Parts {
		words := splitWords(p.Value)
		switch p.Type {
		case namevalues.PartTypeSurname:
			r.surnames = append(r.surnames, words...)
		case namevalues.PartTypeGiven, namevalues.PartTypeInitial:
			r.given = append(r.given, words...)
		case namevalues.PartTypeNick:
			r.nicks = append(r.nicks, words...)
		case namevalues.PartTypeSuffix:
			r.suffixes = append(r.suffixes, strings.Join(words, " "))
		}
	}
	return r
}

func splitWords(s string) []string {
	n := resolve.NormalizeForm(s)
	if n == "" {
		return nil
	}
	return strings.Split(n, " ")
}

// compareStructured scores two names by role when both have typed surname or
// given parts; ok is false otherwise (the caller falls back to form).
//
//   - Surname and given name are scored separately and blended by
//     surnameShare (default 0.6 surname, 0.4 given). A role only one side
//     has counts as half credit: "Robins" alone resembles "James Robins"
//     but less than "James Robins" does. A role neither has is left out,
//     so "James" against "James" is 1.
//   - Given names: the first given name (or a nickname, either side) counts
//     most; the rest, initials included, refine it. "J." ~ "James" is half
//     a match; "James K." ~ "James" is close.
//   - When both have surnames and they share nothing, the given-name match
//     counts at givenOnlyFactor (default half): a surname can change at
//     marriage, but a shared "James" alone is weak.
//   - Different suffixes on both sides (Jr. vs Sr.) multiply the result by
//     suffixConflict (default 0.3): likely father and son.
func (c NameComparer) compareStructured(a, b *namevalues.Value, floor float64) (float64, bool) {
	ra, rb := rolesOf(a), rolesOf(b)
	if !ra.structured() || !rb.structured() {
		return 0, false
	}
	surnameShare := orDefault(c.SurnameShare, defaultSurnameShare)
	surname := roleScore(ra.surnames, rb.surnames)
	given := roleScore(ra.given, rb.given)
	if surname.both {
		surname.sim = wordDice(ra.surnames, rb.surnames, floor, false)
	}
	if given.both {
		given.sim = givenSimilarity(ra, rb, floor)
		if surname.both && surname.sim == 0 {
			given.sim *= orDefault(c.GivenOnlyFactor, defaultGivenOnlyFactor)
		}
	}
	// A role neither name has is left out and the other takes its share.
	var sim float64
	switch {
	case surname.neither:
		sim = given.sim
	case given.neither:
		sim = surname.sim
	default:
		sim = surnameShare*surname.sim + (1-surnameShare)*given.sim
	}
	if len(ra.suffixes) > 0 && len(rb.suffixes) > 0 && !sharesAny(ra.suffixes, rb.suffixes) {
		sim *= orDefault(c.SuffixConflict, defaultSuffixConflict)
	}
	return sim, true
}

// role is one role's comparison: both names have it (sim to be computed),
// only one does (half credit), or neither (left out of the blend).
type role struct {
	sim           float64
	both, neither bool
}

func roleScore(a, b []string) role {
	switch {
	case len(a) > 0 && len(b) > 0:
		return role{both: true}
	case len(a) == 0 && len(b) == 0:
		return role{neither: true}
	}
	return role{sim: unknownRoleCredit}
}

// givenSimilarity weighs the first given name (or any nickname as an
// alternative to it) at firstGivenShare and the whole given set the rest.
func givenSimilarity(a, b nameRoles, floor float64) float64 {
	firsts := func(r nameRoles) []string { return append([]string{r.given[0]}, r.nicks...) }
	var first float64
	for _, x := range firsts(a) {
		for _, y := range firsts(b) {
			if s := wordSimilarity(x, y, floor, true); s > first {
				first = s
			}
		}
	}
	return firstGivenShare*first + (1-firstGivenShare)*wordDice(a.given, b.given, floor, true)
}

// wordDice is the Dice overlap of two word lists under their best one-to-one
// pairing: 2 × (sum of paired word similarities) / (total words). The
// pairing is the exact maximum, so the result is symmetric.
func wordDice(wa, wb []string, floor float64, initials bool) float64 {
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	return 2 * bestPairing(wa, wb, floor, initials) / float64(len(wa)+len(wb))
}

// maxPairedWords bounds the exact search (2^n states over the shorter list);
// longer lists pair greedily by best score first. Names never get close.
const maxPairedWords = 12

// bestPairing is the largest total similarity over one-to-one word pairings.
func bestPairing(wa, wb []string, floor float64, initials bool) float64 {
	if len(wa) < len(wb) {
		wa, wb = wb, wa
	}
	score := make([][]float64, len(wa))
	for i, x := range wa {
		score[i] = make([]float64, len(wb))
		for j, y := range wb {
			score[i][j] = wordSimilarity(x, y, floor, initials)
		}
	}
	if len(wb) > maxPairedWords {
		return greedyPairing(score)
	}
	// best[mask] = best total with the shorter list's words in mask used.
	best := make([]float64, 1<<len(wb))
	for i := range wa {
		next := append([]float64(nil), best...)
		for mask, v := range best {
			for j := range wb {
				if mask&(1<<j) == 0 && score[i][j] > 0 {
					if t := v + score[i][j]; t > next[mask|1<<j] {
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

func greedyPairing(score [][]float64) float64 {
	type pair struct {
		i, j int
		s    float64
	}
	var pairs []pair
	for i := range score {
		for j, s := range score[i] {
			if s > 0 {
				pairs = append(pairs, pair{i, j, s})
			}
		}
	}
	sort.Slice(pairs, func(a, b int) bool { return pairs[a].s > pairs[b].s })
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

func sharesAny(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}
