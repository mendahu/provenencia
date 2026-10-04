package match

import (
	"math"

	"github.com/mendahu/provenencia/core/database/namevalues"
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

// NameComparer compares NameValues word by word, using part types as data
// rather than as gates (compareNames in names.go). Each word carries a role
// (from PartRoles) and a weight; any word may pair with any word of the other
// name, discounted when their roles differ, so names entered in different
// formats still connect. A name with no typed parts is read from its form as
// untyped words.
//
// Word matching: an equal word is 1, an initial against a word it begins
// ("J." ~ "James") 0.5, a near spelling ("Robins" ~ "Robbins") its
// edit-distance ratio when at least FuzzyFloor.
//
// Nil settings take their defaults; set one with Set.
type NameComparer struct {
	// PartRoles maps part types to roles. Default WesternPartRoles; a name
	// format profile may supply its own.
	PartRoles map[string]NameRole

	// Word weights by role. Defaults: family 1.5, first given 1, other given
	// (middle names, initials) 0.5, nick 0.3, untyped 1.
	FamilyWeight, FirstGivenWeight, OtherGivenWeight, NickWeight, UntypedWeight *float64

	// Role affinities for pairs whose roles differ. Defaults: a typed word
	// against an untyped one 0.8, a nickname against a given name 0.9, any
	// other mismatch (a surname against a given name) 0.5.
	UntypedAffinity, NickAffinity, CrossRole *float64

	// GivenOnlyFactor scales the score when both names have family words and
	// none resembles any word of the other name. Default 0.5.
	GivenOnlyFactor *float64
	// SuffixConflict scales the score when both names carry generation words
	// (Jr., Sr.) and share none. Default 0.3.
	SuffixConflict *float64
	// FuzzyFloor is the least edit-distance ratio two words need to count as
	// a spelling variant, 0…1. Default 0.8 (Robins ~ Robbins, not Mary ~ Mark);
	// 1 turns fuzzy matching off. Short words may also differ by one added or
	// dropped letter (Ann ~ Anne, Jon ~ John); see wordSimilarity.
	FuzzyFloor *float64
}

func (c NameComparer) Compare(a, b Value) (float64, bool) {
	if a.Name == nil || b.Name == nil {
		return 0, false
	}
	return c.compareNames(a.Name, b.Name)
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
		return setting(c.FamilyWeight, defaultFamilyWeight)
	case RoleGiven:
		if firstGiven {
			return setting(c.FirstGivenWeight, defaultFirstGivenWeight)
		}
		return setting(c.OtherGivenWeight, defaultOtherGivenWeight)
	case RoleNick:
		return setting(c.NickWeight, defaultNickWeight)
	}
	return setting(c.UntypedWeight, defaultUntypedWeight)
}

// affinity is how far two words' roles agree: 1 for the same role, less when
// the types differ.
func (c NameComparer) affinity(a, b NameRole) float64 {
	switch {
	case a == b:
		return 1
	case a == RoleUntyped || b == RoleUntyped:
		return setting(c.UntypedAffinity, defaultUntypedAffinity)
	case (a == RoleNick && b == RoleGiven) || (a == RoleGiven && b == RoleNick):
		return setting(c.NickAffinity, defaultNickAffinity)
	}
	return setting(c.CrossRole, defaultCrossRole)
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
	floor := setting(c.FuzzyFloor, defaultFuzzyFloor)
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
	if familyConflict(wa, wb, floor) {
		s *= setting(c.GivenOnlyFactor, defaultGivenOnlyFactor)
	}
	if len(ga) > 0 && len(gb) > 0 && !sharesAny(ga, gb) {
		s *= setting(c.SuffixConflict, defaultSuffixConflict)
	}
	return math.Min(s, 1), true
}

// familyConflict: both names have family words, and no family word on
// either side resembles any word of the other name. An initial does not count
// as resembling here ("S." says nothing about "Smith").
func familyConflict(wa, wb []nameWord, floor float64) bool {
	resembles := func(f nameWord, others []nameWord) bool {
		for _, o := range others {
			if wordSimilarity(f.text, o.text, floor, false) > 0 {
				return true
			}
		}
		return false
	}
	var hasA, hasB bool
	for _, x := range wa {
		if x.role == RoleFamily {
			hasA = true
			if resembles(x, wb) {
				return false
			}
		}
	}
	for _, y := range wb {
		if y.role == RoleFamily {
			hasB = true
			if resembles(y, wa) {
				return false
			}
		}
	}
	return hasA && hasB
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
