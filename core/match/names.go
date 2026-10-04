package match

import (
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
//     but less than "James Robins" does.
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

	surname, surnameKnown := unknownRoleCredit, false
	if len(ra.surnames) > 0 && len(rb.surnames) > 0 {
		surname, surnameKnown = wordDice(ra.surnames, rb.surnames, floor, false), true
	}
	given := unknownRoleCredit
	if len(ra.given) > 0 && len(rb.given) > 0 {
		given = givenSimilarity(ra, rb, floor)
		if surnameKnown && surname == 0 {
			given *= orDefault(c.GivenOnlyFactor, defaultGivenOnlyFactor)
		}
	}
	sim := surnameShare*surname + (1-surnameShare)*given
	if len(ra.suffixes) > 0 && len(rb.suffixes) > 0 && !sharesAny(ra.suffixes, rb.suffixes) {
		sim *= orDefault(c.SuffixConflict, defaultSuffixConflict)
	}
	return sim, true
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

// wordDice is the Dice overlap of two word lists over best pairings.
func wordDice(wa, wb []string, floor float64, initials bool) float64 {
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	used := make([]bool, len(wb))
	var shared float64
	for _, x := range wa {
		bestJ, bestS := -1, 0.0
		for j, y := range wb {
			if used[j] {
				continue
			}
			if s := wordSimilarity(x, y, floor, initials); s > bestS {
				bestJ, bestS = j, s
			}
		}
		if bestJ >= 0 {
			used[bestJ] = true
			shared += bestS
		}
	}
	return 2 * shared / float64(len(wa)+len(wb))
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
