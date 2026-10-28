package valuecompare

// Rules are the layer's tunables, one group per kind of value. The numbers
// live in core/matchregistry; this package declares only the shape, so a
// layer takes its section as an argument and never reads a package variable.

// Rules is everything Compare and Key read.
type Rules struct {
	Words    WordRules
	Names    NameRules
	Text     TextRules
	Integers IntegerRules
	Dates    DateRules
}

// WordRules are how two words compare.
type WordRules struct {
	// FuzzyFloor is the least edit-distance ratio (1 − edits / longer length)
	// for a spelling variant; 1 turns fuzzy matching off, short-word variants
	// included.
	FuzzyFloor float64
	// ShortVariantMinLength: words at least this long may differ by one added
	// or dropped letter below FuzzyFloor (Ann ~ Anne).
	ShortVariantMinLength int
	// FoldCredit is the similarity of two words equal only after accent and
	// letter folding (Montréal ~ Montreal).
	FoldCredit float64
	// InitialCredit is an initial against a word it begins ("J." ~ "James").
	InitialCredit float64
	// AliasCredit is a nickname or abbreviation table hit.
	AliasCredit float64
	// MaxExactPairing bounds the exact best-pairing search (2^n states over
	// the shorter word list); longer lists pair greedily. A performance
	// limit, not a scoring weight.
	MaxExactPairing int
	// MaxSpellingLength is the longest word, in runes, that may pair as a
	// spelling variant. A longer word pairs only when equal or equal after
	// folding, which keeps the edit-distance table small.
	MaxSpellingLength int
	// Particles are comparison tokens that rejoin with the following word
	// before comparing (o, mc, mac). A period initial never rejoins.
	Particles []string
	// Folds replaces letters that do not decompose when accents are
	// stripped. An empty map still strips accents.
	Folds map[rune]string
}

// NameRules are how two names compare once their words have roles and
// weights. Patterns and Aliases are catalog data passed in at run time, not
// knobs: Default() leaves both empty.
type NameRules struct {
	// NearMinimum is the least paired mean for a Near name; under it the
	// names are Different.
	NearMinimum float64
	// UntypedAffinity scores a typed word against an untyped one.
	UntypedAffinity float64
	// NickAffinity scores a nickname against a given name.
	NickAffinity float64
	// CrossRole scores any other role mismatch (surname against given).
	CrossRole float64
	// MaxWords is the most words, generation words included, either name may
	// have for the pairing to run. Beyond it names that are not the same are
	// Unknown (too_long), so the exact pairing's cost stays bounded.
	MaxWords int
	// Generation is the weight of a generation word (Jr., III) on either side.
	Generation float64
	// PrefixMinLength is the fewest letters a given, nickname or untyped word
	// needs to take part in the prefix rule: a shortlist prefix key and a
	// comparer prefix hit (Des for Desiree). 0 turns both off.
	PrefixMinLength int
	// PrefixCredit is what a given, nickname or untyped word earns when it is
	// the start of the other's word (Des for Desiree), below an alias and above
	// an initial.
	PrefixCredit float64
	// Western is the pattern for a name that names no pattern, or one this
	// catalog does not have.
	Western NamePattern
	// Patterns are the catalog's name formats by key. Not a knob.
	Patterns map[string]NamePattern
	// Aliases are the catalog's abbreviation and nickname tables, indexed.
	// Not a knob.
	Aliases AliasIndex
}

// TextRules are how two free-text values compare.
type TextRules struct {
	// Partial scales the similarity of a value that contains, or overlaps,
	// another: the share of words in common is multiplied by it.
	Partial float64
	// MaxWords is the most words either side may have for the word alignment
	// to run. Beyond it, texts that are not Same or Equivalent are Unknown
	// (too_long), so the alignment's cost stays bounded.
	MaxWords int
}

// IntegerRules are how two integers compare.
type IntegerRules struct {
	// Tolerance is the largest gap that still resembles (Near). 0 makes any
	// difference a disagreement.
	Tolerance int64
}

// DateRules are how two dates compare.
type DateRules struct {
	// Tolerance is the most civil years apart two dates may be and still be Near.
	Tolerance int
	// NearMinimum is the least similarity of a Near date; below it the pair is Different.
	NearMinimum float64
	// MissingDay is a point inside a month that names no day (May 1817 / 14 May 1817).
	MissingDay float64
	// MissingMonth is a point inside a year that names no month (1817 / Sep 1817).
	MissingMonth float64
	// OtherDay is two points in one month on different days.
	OtherDay float64
	// OtherMonth is two points in one year in different months.
	OtherMonth float64
	// YearsApartFrom is where two points in different years start their fall-off.
	YearsApartFrom float64
	// SpanOneYear is a range, bound or ABT on either side spanning a year or less.
	SpanOneYear float64
	// SpanPerExtraYear is taken off SpanOneYear for each further year of span.
	SpanPerExtraYear float64
	// SpanFloor is the least span score, and the score of an open span (BEF, AFT).
	SpanFloor float64
	// AboutUnits is how many units of its most precise part ABT widens each side.
	AboutUnits int
}
