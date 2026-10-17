// Package valuecompare is layer 1 of matching: how two values of one value
// type relate. Pure: no I/O, no catalog; rules come in as an argument from
// core/matchregistry.
//
// Compare returns a relation, a similarity and a structured detail. Key gives
// a value a normalized key that agrees with Compare: two values share a key at
// a level exactly when Compare relates them at that level.
package valuecompare

import (
	"sort"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
)

// Relation is how two values relate.
type Relation string

const (
	// Same: the same value, up to case and spacing.
	Same Relation = "same"
	// Equivalent: the same value written differently (accents, aliases).
	Equivalent Relation = "equivalent"
	// Contains: one side says the other and more.
	Contains Relation = "contains"
	// Overlaps: each side has something the other lacks, and they share some.
	Overlaps Relation = "overlaps"
	// Near: close but not the same (a spelling variant, a small gap).
	Near Relation = "near"
	// Different: the values disagree.
	Different Relation = "different"
	// Unknown: the layer cannot say (empty value, unsupported type).
	Unknown Relation = "unknown"
)

// Cause is the rule that decided a relation.
type Cause string

const (
	CauseExact              Cause = "exact"
	CauseCase               Cause = "case"
	CauseFold               Cause = "fold"
	CauseAlias              Cause = "alias"
	CauseAbbreviation       Cause = "abbreviation"
	CauseInitial            Cause = "initial"
	CauseSpelling           Cause = "spelling"
	CausePrecision          Cause = "precision"
	CauseMissingPart        Cause = "missing_part"
	CauseOverlap            Cause = "overlap"
	CauseGap                Cause = "gap"
	CauseBeyondTolerance    Cause = "beyond_tolerance"
	CauseGenerationConflict Cause = "generation_conflict"
	CauseEmpty              Cause = "empty"
	CausePhraseOnly         Cause = "phrase_only"
	// CauseBelowMinimum: a pair close enough to be Near but under its type's
	// minimum, so Different.
	CauseBelowMinimum Cause = "below_minimum"
	// CauseNoMatch: nothing in common.
	CauseNoMatch Cause = "no_match"
	// CauseTooLong: text with more words than the alignment may compare.
	CauseTooLong Cause = "too_long"
	// CauseOrder: text whose shared words are out of order.
	CauseOrder Cause = "order"
	// CauseUnsupported: no comparer for the type, or two value types.
	CauseUnsupported Cause = "unsupported"
)

// Side names one of the two compared values.
type Side string

const (
	SideNone Side = ""
	SideA    Side = "a"
	SideB    Side = "b"
)

// PartKind is the kind of a compared part.
type PartKind string

const PartWord PartKind = "word"

// Unit is the unit of Detail.Distance.
type Unit string

const (
	UnitCount Unit = "count"
	UnitDays  Unit = "days"
)

// Result is what Compare decides.
type Result struct {
	Relation   Relation
	Similarity float64 // 0…1; 0 for Unknown
}

// Detail is why.
type Detail struct {
	Parts      []PartMatch // what paired with what, in A's order
	OnlyA      []Part      // parts on one side only
	OnlyB      []Part
	Broader    Side    // for Contains: the less specific side
	Distance   float64 // for curve types: the raw gap …
	Unit       Unit    // … and its unit (UnitCount, UnitDays)
	Cause      Cause   // the rule that decided the relation
	RefA, RefB []byte  // the callers' value refs, echoed
}

// Part is one comparison token.
type Part struct {
	Kind PartKind // word
	Text string   // the comparison token
}

// PartMatch is one pair of parts and how they relate.
type PartMatch struct {
	A, B       Part
	Relation   Relation
	Similarity float64
	Cause      Cause
}

// Value is one value to compare. Type is a propertymeta.ValueType* constant.
// Text is present when NormalizeForm leaves a word, a term when its key is
// non-empty; an integer is always present; a date is present when not nil.
// Ref is the caller's id for the value, echoed in the Detail.
type Value struct {
	Type    string
	Text    string
	Term    string
	Integer int64
	Date    *datevalue.Value
	Ref     []byte
}

// comparer compares and keys the values of one value type.
type comparer interface {
	compare(a, b Value, r Rules) (Result, Detail)
	key(v Value, level Relation, r Rules) (string, bool)
}

// comparers are the supported value types.
var comparers = map[string]comparer{
	propertymeta.ValueTypeText:    text{},
	propertymeta.ValueTypeTerm:    term{},
	propertymeta.ValueTypeInteger: integer{},
	propertymeta.ValueTypeDate:    date{},
}

// Compare relates a to b under rules. Values of different types, and types
// without a comparer, are Unknown with cause unsupported.
func Compare(a, b Value, rules Rules) (Result, Detail) {
	var res Result
	var det Detail
	c, ok := comparers[a.Type]
	if !ok || a.Type != b.Type {
		res, det = Result{Relation: Unknown}, Detail{Cause: CauseUnsupported}
	} else {
		res, det = c.compare(a, b, rules)
	}
	det.RefA, det.RefB = a.Ref, b.Ref
	return res, det
}

// Key is v's normalized key at level: two values of one type have the same key
// exactly when Compare relates them at that level or closer (Same for level
// Same; Same or Equivalent for level Equivalent). Any other level, an empty
// value, or an unsupported type has no key.
func Key(v Value, level Relation, rules Rules) (string, bool) {
	c, ok := comparers[v.Type]
	if !ok || (level != Same && level != Equivalent) {
		return "", false
	}
	return c.key(v, level, rules)
}

// comparerTypes are the sorted value types that have a comparer.
func comparerTypes() []string {
	out := make([]string, 0, len(comparers))
	for t := range comparers {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// falloff is the shared fall-off curve for a type with a distance: 1 at gap 0,
// down to just above 0 at the tolerance, 1 − gap / (tolerance + 1).
func falloff(gap, tolerance float64) float64 {
	return 1 - gap/(tolerance+1)
}
