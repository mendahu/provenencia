package match

import (
	"math"

	"github.com/mendahu/provenencia/core/namevalue"
)

// NameRole is what a name part means for matching. A name pattern maps part
// types to roles, so cultures differ in data, not in comparer code. Roles
// never stop two words from pairing; they only weigh and discount.
type NameRole string

const (
	RoleFamily     NameRole = "family"     // surname
	RoleGiven      NameRole = "given"      // given names and initials, in order
	RoleNick       NameRole = "nick"       // an alternative given name
	RoleUntyped    NameRole = "untyped"    // undetermined, untyped, or a form word
	RoleGeneration NameRole = "generation" // Jr., Sr., III: compared separately
	RoleIgnored    NameRole = "ignored"    // titles, surname particles
)

// NameWeights are what a matched word of each role earns and what an
// unmatched one costs. FirstGiven is the first given word in part order;
// OtherGiven every later one (middle names, initials).
type NameWeights struct {
	Family, FirstGiven, OtherGiven, Nick, Untyped float64
}

// NamePattern is everything culture-specific about comparing names: which
// part type plays which role, and how much each role weighs. The built-in
// Western pattern is in the registry; more patterns are data, not code.
type NamePattern struct {
	Key string
	// PartRoles maps part types to roles; a type missing here is untyped.
	PartRoles map[string]NameRole
	Weights   NameWeights
}

func (p NamePattern) role(partType string) NameRole {
	if r, ok := p.PartRoles[partType]; ok {
		return r
	}
	return RoleUntyped
}

func (p NamePattern) weight(role NameRole, firstGiven bool) float64 {
	switch role {
	case RoleFamily:
		return p.Weights.Family
	case RoleGiven:
		if firstGiven {
			return p.Weights.FirstGiven
		}
		return p.Weights.OtherGiven
	case RoleNick:
		return p.Weights.Nick
	}
	return p.Weights.Untyped
}

// NamePatterns supplies name patterns by key. Today it is BuiltinNamePatterns;
// it is the seam where a catalog read of name format profiles plugs in.
type NamePatterns interface {
	NamePattern(key string) (NamePattern, bool)
}

// NamePatternSet is a fixed set of patterns by key.
type NamePatternSet map[string]NamePattern

func (s NamePatternSet) NamePattern(key string) (NamePattern, bool) {
	p, ok := s[key]
	return p, ok
}

// NameComparer compares NameValues word by word, using part types as data
// rather than as gates. Each name's words take their roles and weights from
// its name pattern (Value.NamePattern, else Pattern); any word may pair with
// any word of the other name, discounted when their roles differ, so names
// entered in different formats, or under different patterns, still connect.
// Comparison reads parts only. form is the transcription and is never read;
// a name with no parts is not comparable.
//
// Nothing here is culture-specific; that is the pattern's job. Settings
// default to DefaultNames (registry.go).
type NameComparer struct {
	// Patterns supplies name patterns.
	Patterns NamePatterns
	// Pattern is the pattern key for names that don't name their own.
	Pattern *string

	// Affinities for pairs whose roles differ: a typed word against an
	// untyped one, a nickname against a given name, any other mismatch.
	UntypedAffinity, NickAffinity, CrossRole *float64
	// GivenOnlyFactor scales the score when both names have family words and
	// none resembles any word of the other name.
	GivenOnlyFactor *float64
	// SuffixConflict scales the score when both names carry generation words
	// and share none.
	SuffixConflict *float64

	// Frequency is how common each word is among the project's filed people
	// (Promote and merge hints supply it; see CountWords). It scales points,
	// never similarity: an identical name is still 1. FrequencyDiscount is
	// the most a word on every person can cost, as a share of points.
	// FrequencyMinPeople is the least population the counts are read against,
	// so a small project's first shared surname is not rare forever.
	Frequency          *WordCounts
	FrequencyDiscount  *float64
	FrequencyMinPeople *int

	Words WordRules
}

func (c NameComparer) Compare(a, b Value) (float64, bool) {
	s, _, ok := c.CompareDiscount(a, b)
	return s, ok
}

// CompareDiscount is Compare plus the share of points the matched words'
// frequency costs (0 with no counts). It makes NameComparer a
// DiscountComparer.
func (c NameComparer) CompareDiscount(a, b Value) (similarity, discount float64, comparable bool) {
	if a.Name == nil || b.Name == nil {
		return 0, 0, false
	}
	return c.compareNames(a.Name, c.pattern(a.NamePattern), b.Name, c.pattern(b.NamePattern))
}

// pattern resolves a name's pattern: its own key if the source has it, else
// the comparer's pattern, else the registry's default, else Western.
func (c NameComparer) pattern(key string) NamePattern {
	src := c.Patterns
	if src == nil {
		src = DefaultNames.Patterns
	}
	for _, k := range []string{key, setting(c.Pattern, *DefaultNames.Pattern), *DefaultNames.Pattern} {
		if k == "" {
			continue
		}
		if p, ok := src.NamePattern(k); ok {
			return p
		}
	}
	return WesternNamePattern
}

// nameWord is one comparable word of a name.
type nameWord struct {
	text   string
	role   NameRole
	weight float64
}

// nameWords splits a name into weighted, role-tagged words under its
// pattern. Only parts are read. Generation words come back separately. A
// name whose parts yield no comparable word is not comparable.
func nameWords(n *namevalue.Value, p NamePattern, particles map[string]bool) (words []nameWord, generation []string) {
	firstGiven := true
	for _, part := range n.Parts {
		role := p.role(part.Type)
		for _, w := range splitWords(part.Value, particles) {
			switch role {
			case RoleIgnored:
			case RoleGeneration:
				generation = append(generation, w)
			default:
				words = append(words, nameWord{text: w, role: role, weight: p.weight(role, firstGiven)})
				if role == RoleGiven {
					firstGiven = false
				}
			}
		}
	}
	return words, generation
}

// affinity is how far two words' roles agree: 1 for the same role, less when
// the types differ.
func (c NameComparer) affinity(a, b NameRole) float64 {
	switch {
	case a == b:
		return 1
	case a == RoleUntyped || b == RoleUntyped:
		return setting(c.UntypedAffinity, *DefaultNames.UntypedAffinity)
	case (a == RoleNick && b == RoleGiven) || (a == RoleGiven && b == RoleNick):
		return setting(c.NickAffinity, *DefaultNames.NickAffinity)
	}
	return setting(c.CrossRole, *DefaultNames.CrossRole)
}

// compareNames scores two names over all their words, whatever their types.
//
// Every word may pair with any word of the other name (one-to-one, the best
// pairing). A pair earns word similarity × role affinity × the two words'
// weights. The score is that earned total over the weight of the words that
// bear on it. A word left over when the other name is fully paired exists
// only on one side: it adds nothing and takes nothing away. Words left over
// on both sides did not meet, and that disagreement costs their weight.
//
// Two rules sit on top:
//   - When both names have family words and none of them resembles any word
//     of the other name, the score is scaled by GivenOnlyFactor: a shared
//     "James" alone is weak, though a surname can change at marriage.
//   - When both carry generation words (Jr., Sr.) and share none, the score
//     is scaled by SuffixConflict: likely father and son.
//
// The discount is separate from the score: the role-weighted mean of how
// common the paired words are in the project (see discount). It never lowers
// the similarity.
func (c NameComparer) compareNames(a *namevalue.Value, pa NamePattern, b *namevalue.Value, pb NamePattern) (float64, float64, bool) {
	r := c.Words.resolve(true)
	wa, ga := nameWords(a, pa, r.particles)
	wb, gb := nameWords(b, pb, r.particles)
	if len(wa) == 0 || len(wb) == 0 {
		return 0, 0, false
	}
	earned := make([][]float64, len(wa))
	for i, x := range wa {
		earned[i] = make([]float64, len(wb))
		for j, y := range wb {
			s := wordSimilarity(x.text, y.text, x.role, y.role, r)
			earned[i][j] = s * c.affinity(x.role, y.role) * (x.weight + y.weight)
		}
	}
	totalEarned, usedA, usedB := bestAssignment(earned)
	total := countedWeight(wa, usedA, usedB) + countedWeight(wb, usedB, usedA)
	if total == 0 {
		return 0, 0, false
	}
	s := totalEarned / total
	if familyConflict(wa, wb, r) {
		s *= setting(c.GivenOnlyFactor, *DefaultNames.GivenOnlyFactor)
	}
	if len(ga) > 0 && len(gb) > 0 && !sharesAny(ga, gb) {
		s *= setting(c.SuffixConflict, *DefaultNames.SuffixConflict)
	}
	return math.Min(s, 1), c.discount(wa, usedA, wb, usedB, r), true
}

// countedWeight is the weight of words that bear on the score. Paired words
// always count. A leftover counts only when the other name also has a
// leftover: those two failed to meet. A leftover on a name whose other side
// is fully paired exists only on one side, so it is neutral.
func countedWeight(words []nameWord, used []bool, otherUsed []bool) float64 {
	otherLeft := false
	for _, u := range otherUsed {
		if !u {
			otherLeft = true
			break
		}
	}
	var sum float64
	for i, w := range words {
		paired := i < len(used) && used[i]
		if paired || otherLeft {
			sum += w.weight
		}
	}
	return sum
}

// familyConflict: both names have family words, and no family word on
// either side resembles any word of the other name. An initial does not count
// as resembling here ("S." says nothing about "Smith").
func familyConflict(wa, wb []nameWord, r wordRules) bool {
	r.initials = false
	resembles := func(f nameWord, others []nameWord) bool {
		for _, o := range others {
			if wordSimilarity(f.text, o.text, f.role, o.role, r) > 0 {
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
