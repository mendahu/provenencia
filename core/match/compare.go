package match

import (
	"strings"

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

// Defaults applied when a comparer's field is zero.
const (
	defaultNamePartial   = 0.8
	defaultTextPartial   = 0.7
	defaultFuzzyFloor    = 0.8
	defaultDateTolerance = 2
)

// NameComparer compares NameValues by their typed parts when both have
// them (see compareStructured in names.go): surname against surname, given
// names (with initials and nicknames) against given names, suffix against
// suffix. form is a reading of the name as written, closer to a
// transcription, so it is only the fallback for a name with no typed surname
// or given part: the same normalized form is 1, otherwise shared words
// (initials and near spellings included) scaled by Partial.
//
// Word matching throughout: an equal word is 1, an initial against a word it
// begins ("J." ~ "James") 0.5, a near spelling ("Robins" ~ "Robbins") its
// edit-distance ratio when at least FuzzyFloor.
type NameComparer struct {
	// SurnameShare is the surname's part of a structured score; the given
	// name has the rest. Default 0.6.
	SurnameShare float64
	// GivenOnlyFactor scales the given-name match when both surnames are
	// known and share nothing. Default 0.5.
	GivenOnlyFactor float64
	// SuffixConflict scales a structured score when both names carry
	// suffixes and they differ (Jr. vs Sr.). Default 0.3.
	SuffixConflict float64
	// Partial caps a non-identical form in the fallback, 0…1. Default 0.8.
	Partial float64
	// FuzzyFloor is the least edit-distance ratio two words need to count as
	// a spelling variant, 0…1. Default 0.8 (Robins ~ Robbins, not Mary ~ Mark);
	// 1 turns fuzzy matching off.
	FuzzyFloor float64
}

func (c NameComparer) Compare(a, b Value) (float64, bool) {
	if a.Name == nil || b.Name == nil {
		return 0, false
	}
	floor := orDefault(c.FuzzyFloor, defaultFuzzyFloor)
	if sim, ok := c.compareStructured(a.Name, b.Name, floor); ok {
		return sim, true
	}
	return compareForms(a.Name.Form, b.Name.Form, orDefault(c.Partial, defaultNamePartial), floor, true)
}

// TextComparer compares free text the way names are compared, without
// initials: the same normalized text is 1, shared words are partial
// ("York, Upper Canada" ~ "York").
type TextComparer struct {
	// Partial caps non-identical text, 0…1. Default 0.7.
	Partial float64
	// FuzzyFloor as for NameComparer. Default 0.8.
	FuzzyFloor float64
}

func (c TextComparer) Compare(a, b Value) (float64, bool) {
	if !a.HasText || !b.HasText {
		return 0, false
	}
	return compareForms(a.Text, b.Text, orDefault(c.Partial, defaultTextPartial),
		orDefault(c.FuzzyFloor, defaultFuzzyFloor), false)
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

// DateComparer compares dates by their years, refined by month and day.
// The same day is 1, the same month 0.9, the same year 0.8 (0.6 when the
// months are known and differ). Years apart within Tolerance fall off
// linearly from 0.8; beyond it is a disagreement (0). An approximate date
// (ABT / BEF / AFT) on either side doubles the tolerance. A date with no
// year (a phrase) is not comparable.
type DateComparer struct {
	// Tolerance is how many years apart still resemble. Default 2.
	Tolerance int
}

func (c DateComparer) Compare(a, b Value) (float64, bool) {
	if a.Date == nil || b.Date == nil {
		return 0, false
	}
	ya, ma, da := dateParts(a)
	yb, mb, db := dateParts(b)
	if ya == nil || yb == nil {
		return 0, false
	}
	tol := c.Tolerance
	if tol == 0 {
		tol = defaultDateTolerance
	}
	if a.Date.Qualifier != "" || b.Date.Qualifier != "" {
		tol *= 2
	}
	gap := *ya - *yb
	if gap < 0 {
		gap = -gap
	}
	switch {
	case gap == 0 && ma != nil && mb != nil && *ma == *mb && da != nil && db != nil && *da == *db:
		return 1, true
	case gap == 0 && ma != nil && mb != nil && *ma == *mb:
		return 0.9, true
	case gap == 0 && ma != nil && mb != nil:
		return 0.6, true
	case gap == 0:
		return 0.8, true
	case gap <= tol:
		return 0.8 * (1 - float64(gap)/float64(tol+1)), true
	}
	return 0, true
}

// dateParts is the date's leading year, month, and day: the start of a
// range, else its end.
func dateParts(v Value) (y, m, d *int) {
	if v.Date.StartYear != nil {
		return v.Date.StartYear, v.Date.StartMonth, v.Date.StartDay
	}
	return v.Date.EndYear, v.Date.EndMonth, v.Date.EndDay
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
// begins, the edit-distance ratio for spelling variants at or above floor,
// else 0. A lone letter matches only as an initial.
func wordSimilarity(x, y string, floor float64, initials bool) float64 {
	if x == y {
		if len([]rune(x)) == 1 && !initials {
			return 0
		}
		return 1
	}
	rx, ry := []rune(x), []rune(y)
	if len(rx) == 1 || len(ry) == 1 {
		if initials && rx[0] == ry[0] {
			return 0.5
		}
		return 0
	}
	if floor >= 1 {
		return 0
	}
	longest := len(rx)
	if len(ry) > longest {
		longest = len(ry)
	}
	ratio := 1 - float64(levenshtein(rx, ry))/float64(longest)
	if ratio < floor {
		return 0
	}
	return ratio
}

func levenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func orDefault(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}
