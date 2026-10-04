package match

import (
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

// Defaults applied when a comparer's setting is nil. Name and date defaults
// live with their comparers (names.go, dates.go).
const (
	defaultTextPartial = 0.7
	defaultFuzzyFloor  = 0.8
)

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
