package valuecompare

// Rules are the layer's tunables, one group per kind of value. The numbers
// live in core/matchregistry; this package declares only the shape, so a
// layer takes its section as an argument and never reads a package variable.

// Rules is everything Compare and Key read.
type Rules struct {
	Words    WordRules
	Text     TextRules
	Integers IntegerRules
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
