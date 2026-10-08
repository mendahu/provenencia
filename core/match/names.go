package match

import (
	"math"

	"github.com/mendahu/provenencia/core/database/namevalues"
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

	Words WordRules
}

func (c NameComparer) Compare(a, b Value) (float64, bool) {
	if a.Name == nil || b.Name == nil {
		return 0, false
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
func nameWords(n *namevalues.Value, p NamePattern) (words []nameWord, generation []string) {
	firstGiven := true
	for _, part := range n.Parts {
		role := p.role(part.Type)
		for _, w := range splitWords(part.Value) {
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
func (c NameComparer) compareNames(a *namevalues.Value, pa NamePattern, b *namevalues.Value, pb NamePattern) (float64, bool) {
	r := c.Words.resolve(true)
	wa, ga := nameWords(a, pa)
	wb, gb := nameWords(b, pb)
	if len(wa) == 0 || len(wb) == 0 {
		return 0, false
	}
	sim := make([][]float64, len(wa))
	earned := make([][]float64, len(wa))
	for i, x := range wa {
		sim[i] = make([]float64, len(wb))
		earned[i] = make([]float64, len(wb))
		for j, y := range wb {
			sim[i][j] = wordSimilarity(x.text, y.text, r)
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
	if familyConflict(wa, wb, r) {
		s *= setting(c.GivenOnlyFactor, *DefaultNames.GivenOnlyFactor)
	}
	if len(ga) > 0 && len(gb) > 0 && !sharesAny(ga, gb) {
		s *= setting(c.SuffixConflict, *DefaultNames.SuffixConflict)
	}
	return math.Min(s, 1), true
}

// familyConflict: both names have family words, and no family word on
// either side resembles any word of the other name. An initial does not count
// as resembling here ("S." says nothing about "Smith").
func familyConflict(wa, wb []nameWord, r wordRules) bool {
	r.initials = false
	resembles := func(f nameWord, others []nameWord) bool {
		for _, o := range others {
			if wordSimilarity(f.text, o.text, r) > 0 {
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
