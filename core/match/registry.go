package match

import "github.com/mendahu/provenencia/core/database/namevalues"

// The registry: every number the matching algorithm uses, every built-in name
// pattern, and every default profile, in one file. Tune here.
//
// Comparers read their defaults from this file: a nil setting on any
// comparer falls back to the value below, so a bare NameComparer{} and
// DefaultNames score identically, and no number is written twice. Callers
// override a setting per call with Set (match.NameComparer{CrossRole:
// match.Set(0.0)}) or pass a whole Profile (matching.Options.Profile).
// registry_test.go fails if a setting is added without a registry value.
//
// Later, per-project profiles and data-driven name patterns replace these
// values through the same shapes: a stored Profile, and a NamePatterns
// source backed by name_format_profiles instead of BuiltinNamePatterns.

// ---------------------------------------------------------------------------
// Words (names and free text)

// DefaultWordRules: how two words compare.
var DefaultWordRules = WordRules{
	// Least edit-distance ratio (1 − edits / longer length) for a spelling
	// variant: Robins ~ Robbins (0.86) counts, Mary ~ Mark (0.75) does not.
	// 1 turns fuzzy matching off.
	FuzzyFloor: Set(0.8),
	// An initial against a word it begins: "J." ~ "James".
	InitialCredit: Set(0.5),
	// Words at least this long may differ by one added or dropped letter
	// below FuzzyFloor (Ann ~ Anne, Jon ~ John).
	ShortVariantMinLength: Set(3),
}

// MaxExactPairing bounds the exact best-pairing search (2^n states over the
// shorter word list); longer lists pair greedily. A performance limit, not a
// scoring weight: names never get close.
const MaxExactPairing = 12

// ---------------------------------------------------------------------------
// Names

// DefaultNames: how two names compare once their words have roles and weights
// (which come from the name pattern, below).
var DefaultNames = NameComparer{
	Patterns: BuiltinNamePatterns,
	// The pattern for names that don't name one (today: every name).
	// Later: the project default name format.
	Pattern: Set(WesternPattern),

	// Affinities: how much a pair of words counts when their roles differ.
	UntypedAffinity: Set(0.8), // a typed word against an untyped one
	NickAffinity:    Set(0.9), // a nickname against a given name
	CrossRole:       Set(0.5), // any other mismatch: surname against given

	// Both names have family words and none resembles any word of the other.
	GivenOnlyFactor: Set(0.5),
	// Both names carry generation words (Jr., Sr.) and share none.
	SuffixConflict: Set(0.3),

	Words: DefaultWordRules,
}

// WesternPattern is the key of the built-in Western name pattern.
const WesternPattern = "western"

// WesternNamePattern: given names then surname, with GEDCOM-style part types.
// The first given name leads; middle names and initials count less; titles
// and surname particles are ignored. Every product part type (namevalues
// registry) must have a role here.
var WesternNamePattern = NamePattern{
	Key: WesternPattern,
	PartRoles: map[string]NameRole{
		namevalues.PartTypeSurname:       RoleFamily,
		namevalues.PartTypeGiven:         RoleGiven,
		namevalues.PartTypeInitial:       RoleGiven,
		namevalues.PartTypeNick:          RoleNick,
		namevalues.PartTypeSuffix:        RoleGeneration,
		namevalues.PartTypePrefix:        RoleIgnored,
		namevalues.PartTypeSurnamePrefix: RoleIgnored,
		namevalues.PartTypeUndetermined:  RoleUntyped,
		"":                               RoleUntyped,
	},
	// What a matched word of each role earns and an unmatched one costs.
	Weights: NameWeights{
		Family:     1.5,
		FirstGiven: 1,
		OtherGiven: 0.5, // middle names and initials
		Nick:       0.3,
		Untyped:    1,
	},
}

// BuiltinNamePatterns are the patterns that ship in code. A later NamePatterns
// source reads name_format_profiles instead; nothing else changes.
var BuiltinNamePatterns = NamePatternSet{
	WesternPattern: WesternNamePattern,
}

// ---------------------------------------------------------------------------
// Dates

// DefaultDates: how two dates compare as spans of years.
var DefaultDates = DateComparer{
	// Years apart that still resemble; beyond is a disagreement.
	Tolerance: Set(2),
	// ABT on either point multiplies the tolerance.
	ApproxToleranceFactor: Set(2),

	// Two points in the same year (identical dates are always 1):
	MissingDay:   Set(0.9), // same month, a day missing on one side
	MissingMonth: Set(0.8), // a month missing on one side
	OtherDay:     Set(0.7), // same month, different known days
	OtherMonth:   Set(0.6), // different known months
	// Points years apart fall off linearly from this, within the tolerance.
	YearsApartFrom: Set(0.8),

	// A range or bound on either side, overlapping: judged by the wider span.
	SpanOneYear:      Set(0.8),  // a one-year span
	SpanPerExtraYear: Set(0.05), // less for each extra year of width
	SpanFloor:        Set(0.2),  // never below; open spans (BEF, AFT, FROM)
}

// ---------------------------------------------------------------------------
// Text, terms, integers

// DefaultText: free text (toponyms, remarks). The same normalized text is 1;
// otherwise shared words, capped at Partial.
var DefaultText = TextComparer{
	Partial: Set(0.7),
	Words:   DefaultWordRules,
}

// DefaultIntegers: equal is 1; Tolerance 0 means any difference disagrees.
var DefaultIntegers = IntegerComparer{
	Tolerance: Set(int64(0)),
}

// SexAtBirthNeutral are sex_at_birth terms that are no evidence either way.
var SexAtBirthNeutral = map[string]bool{"unknown": true, "indeterminate": true}

// ---------------------------------------------------------------------------
// Profiles and consumers

// DefaultSuggestionLimit caps Promote's target suggestions when the caller
// passes no limit.
const DefaultSuggestionLimit = 10

// DefaultProfile is the shipped algorithm for a Subject type key, or false
// when the type has none. Every consumer (Promote suggestions, merge hints)
// reads these unless it passes its own Profile.
//
// The weights are points. As a guide: a candidate worth showing scores at
// least MinScore; the same name alone clears it, a shared surname alone
// clears it barely, and one hard contradiction sinks any name match.
func DefaultProfile(kind string) (Profile, bool) {
	switch kind {
	case "person":
		return Profile{
			Kind: "person",
			Features: []Feature{
				{Property: product("name"), Comparer: DefaultNames, Weight: 10},
				{Property: product("sex_at_birth"), Comparer: TermComparer{Neutral: SexAtBirthNeutral}, Weight: 1, Contradiction: 8},
			},
			// "Mary Robins" ~ "James Robins" is 10 × 0.55 = 5.5 typed (5 as
			// forms): shown, low. "James Smith" is 10 × 0.2 = 2: hidden.
			MinScore: 3,
		}, true
	case "event":
		return Profile{
			Kind: "event",
			Features: []Feature{
				{Property: product("event_type"), Comparer: TermComparer{}, Weight: 4, Contradiction: 6},
				{Property: product("date"), Comparer: DefaultDates, Weight: 6, Contradiction: 4},
				{Property: product("start_date"), Comparer: DefaultDates, Weight: 3, Contradiction: 2},
				{Property: product("end_date"), Comparer: DefaultDates, Weight: 3, Contradiction: 2},
			},
			// The same type alone (every birth) is not enough; a date is.
			MinScore: 5,
		}, true
	case "place":
		return Profile{
			Kind: "place",
			Features: []Feature{
				{Property: product("toponym"), Comparer: DefaultText, Weight: 10},
			},
			MinScore: 3,
		}, true
	}
	return Profile{}, false
}
