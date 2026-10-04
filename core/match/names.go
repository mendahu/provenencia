package match

import (
	"math"
	"sort"
	"strings"

	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/resolve"
)

// NameRole is what a name part means for matching. Part types map to roles
// through NameComparer.PartRoles, so a name format profile can supply its
// own mapping (a patronymic, a maternal surname) without new comparer code.
// Roles never stop two words from pairing; they only weigh and discount.
type NameRole string

const (
	RoleFamily     NameRole = "family"     // surname
	RoleGiven      NameRole = "given"      // given names and initials, in order
	RoleNick       NameRole = "nick"       // an alternative given name
	RoleUntyped    NameRole = "untyped"    // undetermined, untyped, or a form word
	RoleGeneration NameRole = "generation" // Jr., Sr., III: compared separately
	RoleIgnored    NameRole = "ignored"    // titles, surname particles
)

// WesternPartRoles maps the product part types (namevalues) to roles. Types
// missing from a map are untyped.
var WesternPartRoles = map[string]NameRole{
	namevalues.PartTypeSurname:       RoleFamily,
	namevalues.PartTypeGiven:         RoleGiven,
	namevalues.PartTypeInitial:       RoleGiven,
	namevalues.PartTypeNick:          RoleNick,
	namevalues.PartTypeSuffix:        RoleGeneration,
	namevalues.PartTypePrefix:        RoleIgnored,
	namevalues.PartTypeSurnamePrefix: RoleIgnored,
	namevalues.PartTypeUndetermined:  RoleUntyped,
	"":                               RoleUntyped,
}

// Defaults for NameComparer. Weights are relative: what a matched word of
// that role earns, and what an unmatched one costs.
const (
	defaultFamilyWeight     = 1.5
	defaultFirstGivenWeight = 1
	defaultOtherGivenWeight = 0.5
	defaultNickWeight       = 0.3
	defaultUntypedWeight    = 1
	defaultCrossRole        = 0.5
	defaultUntypedAffinity  = 0.8
	defaultNickAffinity     = 0.9
	defaultGivenOnlyFactor  = 0.5
	defaultSuffixConflict   = 0.3
)

// nameWord is one comparable word of a name.
type nameWord struct {
	text   string
	role   NameRole
	weight float64
}

// nameWords splits a name into weighted, role-tagged words. Parts give the
// roles; a name whose parts yield no comparable word is read from its form
// as untyped words. Generation words come back separately.
func (c NameComparer) nameWords(n *namevalues.Value) (words []nameWord, generation []string) {
	roles := c.PartRoles
	if roles == nil {
		roles = WesternPartRoles
	}
	firstGiven := true
	for _, p := range n.Parts {
		role, ok := roles[p.Type]
		if !ok {
			role = RoleUntyped
		}
		for _, w := range splitWords(p.Value) {
			switch role {
			case RoleIgnored:
			case RoleGeneration:
				generation = append(generation, w)
			default:
				words = append(words, nameWord{text: w, role: role, weight: c.weight(role, firstGiven)})
				if role == RoleGiven {
					firstGiven = false
				}
			}
		}
	}
	if len(words) == 0 {
		for _, w := range splitWords(n.Form) {
			words = append(words, nameWord{text: w, role: RoleUntyped, weight: c.weight(RoleUntyped, false)})
		}
	}
	return words, generation
}

func (c NameComparer) weight(role NameRole, firstGiven bool) float64 {
	switch role {
	case RoleFamily:
		return orDefault(c.FamilyWeight, defaultFamilyWeight)
	case RoleGiven:
		if firstGiven {
			return orDefault(c.FirstGivenWeight, defaultFirstGivenWeight)
		}
		return orDefault(c.OtherGivenWeight, defaultOtherGivenWeight)
	case RoleNick:
		return orDefault(c.NickWeight, defaultNickWeight)
	}
	return orDefault(c.UntypedWeight, defaultUntypedWeight)
}

// affinity is how far two words' roles agree: 1 for the same role, less when
// the types differ.
func (c NameComparer) affinity(a, b NameRole) float64 {
	switch {
	case a == b:
		return 1
	case a == RoleUntyped || b == RoleUntyped:
		return orDefault(c.UntypedAffinity, defaultUntypedAffinity)
	case (a == RoleNick && b == RoleGiven) || (a == RoleGiven && b == RoleNick):
		return orDefault(c.NickAffinity, defaultNickAffinity)
	}
	return orDefault(c.CrossRole, defaultCrossRole)
}

func splitWords(s string) []string {
	n := resolve.NormalizeForm(s)
	if n == "" {
		return nil
	}
	return strings.Split(n, " ")
}

// compareNames scores two names over all their words, whatever their types.
//
// Every word may pair with any word of the other name (one-to-one, the best
// pairing). A pair earns word similarity × role affinity × the two words'
// weights, and the score is the earned share of both names' total weight. So
// the same words in the same roles score 1; the same words typed differently
// (a surname against a given name, a typed part against a form word) still
// connect, discounted; and an unmatched word costs its weight.
//
// Two rules sit on top:
//   - When both names have family words and none of them resembles any word
//     of the other name, the score is scaled by GivenOnlyFactor: a shared
//     "James" alone is weak, though a surname can change at marriage.
//   - When both carry generation words (Jr., Sr.) and share none, the score
//     is scaled by SuffixConflict: likely father and son.
func (c NameComparer) compareNames(a, b *namevalues.Value) (float64, bool) {
	floor := orDefault(c.FuzzyFloor, defaultFuzzyFloor)
	wa, ga := c.nameWords(a)
	wb, gb := c.nameWords(b)
	if len(wa) == 0 || len(wb) == 0 {
		return 0, false
	}
	sim := make([][]float64, len(wa))
	earned := make([][]float64, len(wa))
	for i, x := range wa {
		sim[i] = make([]float64, len(wb))
		earned[i] = make([]float64, len(wb))
		for j, y := range wb {
			sim[i][j] = wordSimilarity(x.text, y.text, floor, true)
			earned[i][j] = sim[i][j] * c.affinity(x.role, y.role) * (x.weight + y.weight)
		}
	}
	var total float64
	for _, w := range wa {
		total += w.weight
	}
	for _, w := range wb {
		total += w.weight
	}
	s := bestPairing(earned) / total
	if familyConflict(wa, wb, sim) {
		s *= orDefault(c.GivenOnlyFactor, defaultGivenOnlyFactor)
	}
	if len(ga) > 0 && len(gb) > 0 && !sharesAny(ga, gb) {
		s *= orDefault(c.SuffixConflict, defaultSuffixConflict)
	}
	return math.Min(s, 1), true
}

// familyConflict: both names have family words, and no family word on
// either side resembles any word of the other name.
func familyConflict(wa, wb []nameWord, sim [][]float64) bool {
	var hasA, hasB bool
	for i, x := range wa {
		if x.role != RoleFamily {
			continue
		}
		hasA = true
		for j := range wb {
			if sim[i][j] > 0 {
				return false
			}
		}
	}
	for j, y := range wb {
		if y.role != RoleFamily {
			continue
		}
		hasB = true
		for i := range wa {
			if sim[i][j] > 0 {
				return false
			}
		}
	}
	return hasA && hasB
}

// wordDice is the Dice overlap of two unweighted word lists under their best
// one-to-one pairing (free text).
func wordDice(wa, wb []string, floor float64, initials bool) float64 {
	if len(wa) == 0 || len(wb) == 0 {
		return 0
	}
	score := make([][]float64, len(wa))
	for i, x := range wa {
		score[i] = make([]float64, len(wb))
		for j, y := range wb {
			score[i][j] = wordSimilarity(x, y, floor, initials)
		}
	}
	return 2 * bestPairing(score) / float64(len(wa)+len(wb))
}

// maxPairedWords bounds the exact search (2^n states over the shorter side);
// longer lists pair greedily by best score first. Names never get close.
const maxPairedWords = 12

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
	if cols > maxPairedWords {
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
