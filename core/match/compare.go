package match

import (
	"strings"
	"unicode"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

// Comparer judges two values of one Property. Similarity is 0…1: 1 is the
// same value, 0 a clear disagreement (which a Feature may penalize).
// comparable is false when either value carries nothing to judge (no year,
// a neutral term), so the pair neither helps nor hurts.
type Comparer interface {
	Compare(a, b Value) (similarity float64, comparable bool)
}

// ComparerFor is the default Comparer for a Property's value type, so a
// profile can weigh any Property, including researcher-defined ones. nil for
// value types that are not compared (subject).
func ComparerFor(valueType string) Comparer {
	switch valueType {
	case properties.ValueTypeName:
		return NameComparer{}
	case properties.ValueTypeText:
		return TextComparer{}
	case properties.ValueTypeTerm:
		return TermComparer{}
	case properties.ValueTypeDate:
		return DateComparer{}
	case properties.ValueTypeInteger:
		return IntegerComparer{}
	}
	return nil
}

// Set returns a pointer to v, for comparer settings. A nil setting takes its
// default, so an explicit zero is a real value: CrossRole: Set(0.0) makes
// mismatched name types never pair; Tolerance: Set(0) requires the same year.
func Set[T any](v T) *T { return &v }

func setting[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// Defaults applied when a comparer's setting is nil.
const (
	defaultTextPartial   = 0.7
	defaultFuzzyFloor    = 0.8
	defaultDateTolerance = 2
)

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

// TextComparer compares free text the way names are compared, without
// initials: the same normalized text is 1, shared words are partial
// ("York, Upper Canada" ~ "York").
type TextComparer struct {
	// Partial caps non-identical text, 0…1. Default 0.7.
	Partial *float64
	// FuzzyFloor as for NameComparer. Default 0.8.
	FuzzyFloor *float64
}

func (c TextComparer) Compare(a, b Value) (float64, bool) {
	if !a.HasText || !b.HasText {
		return 0, false
	}
	return compareForms(a.Text, b.Text, setting(c.Partial, defaultTextPartial),
		setting(c.FuzzyFloor, defaultFuzzyFloor), false)
}

// TermComparer compares vocabulary terms by key: the same term is 1, any
// other 0. Neutral terms ("unknown") are not evidence either way.
type TermComparer struct {
	Neutral map[string]bool
}

func (c TermComparer) Compare(a, b Value) (float64, bool) {
	if a.Term == "" || b.Term == "" || c.Neutral[a.Term] || c.Neutral[b.Term] {
		return 0, false
	}
	if a.Term == b.Term {
		return 1, true
	}
	return 0, true
}

// DateComparer compares dates as spans of years.
//
//   - Identical dates score 1 at any precision ("1817" against "1817",
//     "BEF 1820" against "BEF 1820"); only differing precision costs.
//   - Two exact points (or ABT points) compare by year, refined by month and
//     day: the same month with a day missing on one side 0.9; the same year
//     with a month missing on one side 0.8; the same month but different known days
//     0.7; different known months 0.6. Years apart within Tolerance fall off
//     linearly from 0.8; beyond it is a disagreement (0). ABT on either side
//     doubles the tolerance.
//   - When either side is a range (FROM / TO / BET) or a bound (BEF, AFT),
//     identical spans are 1. Otherwise overlapping spans are consistent, and
//     how much that says depends on the wider span: 0.8 for one year, 0.05
//     less per extra year, never below 0.2; an open span (BEF, AFT, FROM
//     without TO) says little and is 0.2. Spans apart fall off from that
//     score by the years between them within Tolerance; beyond it is a
//     disagreement. "BEF 1820" against 1823 is 3 years apart, not an
//     approximate match.
//
// A date with no year (a phrase) is not comparable.
type DateComparer struct {
	// Tolerance is how many years apart still resemble. Default 2; Set(0)
	// requires the same year (or overlapping spans).
	Tolerance *int
}

// yearSpan is a date's years, lo…hi; an open end is unbounded (BEF, AFT,
// FROM without TO). point marks an exact or ABT point, whose month and day
// refine the comparison.
type yearSpan struct {
	lo, hi         int
	openLo, openHi bool
	point, approx  bool
	month, day     *int
}

func spanOf(d *datevalues.Value) (yearSpan, bool) {
	if d.Kind == datevalues.KindRange {
		if d.StartYear == nil && d.EndYear == nil {
			return yearSpan{}, false
		}
		s := yearSpan{openLo: d.StartYear == nil, openHi: d.EndYear == nil}
		if d.StartYear != nil {
			s.lo = *d.StartYear
		}
		if d.EndYear != nil {
			s.hi = *d.EndYear
		}
		if !s.openLo && !s.openHi && s.lo > s.hi {
			s.lo, s.hi = s.hi, s.lo // an inverted range reads as its years
		}
		return s, true
	}
	y, m, day := d.StartYear, d.StartMonth, d.StartDay
	if y == nil {
		y, m, day = d.EndYear, d.EndMonth, d.EndDay
	}
	if y == nil {
		return yearSpan{}, false
	}
	switch d.Qualifier {
	case datevalues.QualifierBEF:
		return yearSpan{hi: *y, openLo: true}, true
	case datevalues.QualifierAFT:
		return yearSpan{lo: *y, openHi: true}, true
	}
	return yearSpan{lo: *y, hi: *y, point: true, approx: d.Qualifier != "", month: m, day: day}, true
}

// Span scoring: a one-year span is as good as a same-year point (0.8); each
// extra year of width costs spanStep, down to spanFloor, which open spans get.
const (
	spanBest  = 0.8
	spanStep  = 0.05
	spanFloor = 0.2
)

// spanScore is what overlapping spans are worth, judged by the wider one.
func spanScore(a, b yearSpan) float64 {
	width := func(s yearSpan) (int, bool) {
		if s.openLo || s.openHi {
			return 0, false
		}
		return s.hi - s.lo + 1, true
	}
	wa, okA := width(a)
	wb, okB := width(b)
	if !okA || !okB {
		return spanFloor
	}
	w := max(wa, wb)
	return max(spanBest-spanStep*float64(w-1), spanFloor)
}

// sameOptional: both missing, or both present and equal.
func sameOptional(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// yearsApart is the gap between two spans, 0 when they overlap.
func yearsApart(a, b yearSpan) int {
	switch {
	case !a.openHi && !b.openLo && a.hi < b.lo:
		return b.lo - a.hi
	case !b.openHi && !a.openLo && b.hi < a.lo:
		return a.lo - b.hi
	}
	return 0
}

func (c DateComparer) Compare(a, b Value) (float64, bool) {
	if a.Date == nil || b.Date == nil {
		return 0, false
	}
	sa, okA := spanOf(a.Date)
	sb, okB := spanOf(b.Date)
	if !okA || !okB {
		return 0, false
	}
	tol := setting(c.Tolerance, defaultDateTolerance)
	gap := yearsApart(sa, sb)

	if !sa.point || !sb.point {
		if sa == sb {
			return 1, true
		}
		if gap > tol {
			return 0, true
		}
		return spanScore(sa, sb) * (1 - float64(gap)/float64(tol+1)), true
	}
	if sa.approx || sb.approx {
		tol *= 2
	}
	sameMonth := sa.month != nil && sb.month != nil && *sa.month == *sb.month
	switch {
	case gap == 0 && sameOptional(sa.month, sb.month) && sameOptional(sa.day, sb.day):
		return 1, true
	case gap == 0 && sameMonth && sa.day != nil && sb.day != nil:
		return 0.7, true
	case gap == 0 && sameMonth:
		return 0.9, true
	case gap == 0 && sa.month != nil && sb.month != nil:
		return 0.6, true
	case gap == 0:
		return 0.8, true
	case gap <= tol:
		return 0.8 * (1 - float64(gap)/float64(tol+1)), true
	}
	return 0, true
}

// IntegerComparer compares integers: equal is 1; within Tolerance falls off
// linearly; beyond it is a disagreement.
type IntegerComparer struct {
	Tolerance int64
}

func (c IntegerComparer) Compare(a, b Value) (float64, bool) {
	if !a.HasInteger || !b.HasInteger {
		return 0, false
	}
	gap := a.Integer - b.Integer
	if gap < 0 {
		gap = -gap
	}
	if gap > c.Tolerance {
		return 0, true
	}
	return 1 - float64(gap)/float64(c.Tolerance+1), true
}

// compareForms is 1 for the same normalized form, else partial × the share
// of words the two forms have in common (wordDice). Empty forms are not
// comparable.
func compareForms(a, b string, partial, fuzzyFloor float64, initials bool) (float64, bool) {
	wa, wb := splitWords(a), splitWords(b)
	if len(wa) == 0 || len(wb) == 0 {
		return 0, false
	}
	if strings.Join(wa, " ") == strings.Join(wb, " ") {
		return 1, true
	}
	return partial * wordDice(wa, wb, fuzzyFloor, initials), true
}

// wordSimilarity: 1 for the same word, 0.5 for an initial and a word it
// begins (when initials count), else the spelling-variant ratio
// (1 − edits / longer length, an adjacent swap being one edit) when it
// reaches floor — or, for short words, when they differ by one added or
// dropped letter (Ann ~ Anne, Jon ~ John: 0.75), since one edit is a larger
// share of a short word. A one-letter substitution in a short word is a
// different name (Mary ~ Mark). An initial is a lone cased letter ("J"); a
// lone character in an uncased script (蒋, 王) is a whole word.
func wordSimilarity(x, y string, floor float64, initials bool) float64 {
	if x == y {
		return 1
	}
	rx, ry := []rune(x), []rune(y)
	if isInitial(rx) || isInitial(ry) {
		if initials && rx[0] == ry[0] {
			return 0.5
		}
		return 0
	}
	if floor >= 1 {
		return 0
	}
	longest, shortest := len(rx), len(ry)
	if shortest > longest {
		longest, shortest = shortest, longest
	}
	edits := editDistance(rx, ry)
	ratio := 1 - float64(edits)/float64(longest)
	if ratio >= floor {
		return ratio
	}
	if edits == 1 && longest == shortest+1 && shortest >= 3 {
		return ratio
	}
	return 0
}

func isInitial(r []rune) bool {
	return len(r) == 1 && unicode.ToUpper(r[0]) != unicode.ToLower(r[0])
}

// editDistance is the optimal-string-alignment distance: insertions,
// deletions, substitutions, and swaps of adjacent letters ("Robnis" ~
// "Robins") each cost one.
func editDistance(a, b []rune) int {
	d := make([][]int, len(a)+1)
	for i := range d {
		d[i] = make([]int, len(b)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(a)][len(b)]
}
