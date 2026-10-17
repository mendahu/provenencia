package valuecompare

// WesternPattern is the key catalogs store for the built-in Western name
// pattern.
const WesternPattern = "western"

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
	// PartRoles maps part types to roles; a type missing here is untyped.
	PartRoles map[string]NameRole
	Weights   NameWeights
}

// Role is the role of a part type under the pattern.
func (p NamePattern) Role(partType string) NameRole {
	if r, ok := p.PartRoles[partType]; ok {
		return r
	}
	return RoleUntyped
}

// Weight is what a word of the role weighs; firstGiven is whether it is the
// first given word in part order.
func (p NamePattern) Weight(role NameRole, firstGiven bool) float64 {
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
