// Package propertycompare is layer 2 of matching: whether two Subjects agree
// on one property, from every value each one records for it. Pure: no I/O, no
// catalog, no numbers; layer 1's rules come in as an argument.
//
// Compare drops values that are no evidence, lets a negative value contradict
// a value it matches, runs layer 1 on every pair of positive values, picks the
// pair that speaks, and reads the outcome from that pair's relation.
package propertycompare

import (
	"slices"

	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// Value is one value a Subject records for the property. Type is ignored:
// Compare compares every value as the property's value type.
type Value struct {
	valuecompare.Value
	Negative bool // the Subject says the property does not have this value
	Neutral  bool // a term that is no evidence (sex "unknown")
}

// Positive reports whether v speaks for the property: neither negative nor
// neutral.
func (v Value) Positive() bool { return !v.Negative && !v.Neutral }

// Outcome is how two Subjects stand on a property.
type Outcome string

const (
	OutcomeAgree    Outcome = "agree"
	OutcomePartial  Outcome = "partial"
	OutcomeConflict Outcome = "conflict"
	OutcomeUnknown  Outcome = "unknown"
)

// Cause is why an outcome came out as it did.
type Cause string

const (
	// CauseRelation: the speaking pair's relation decided it.
	CauseRelation Cause = "relation"
	// CauseNegative: a negative value contradicted a value it matches.
	CauseNegative Cause = "negative"
	// CauseNoValues: a side has no positive value.
	CauseNoValues Cause = "no_values"
	// CauseCardinality: a difference on a multiple-value property.
	CauseCardinality Cause = "cardinality"
)

// ExcludeReason is why a value was dropped.
type ExcludeReason string

const (
	// ExcludeNeutral: the value is no evidence.
	ExcludeNeutral ExcludeReason = "neutral"
)

// Result is what Compare decides.
type Result struct {
	Outcome  Outcome
	Cause    Cause
	Speaking *SpeakingPair // the pair that spoke; nil when a side has no positive value
	Excluded []Excluded    // values dropped as non-evidence, side A's then side B's
}

// SpeakingPair is the one pair of values that decided the outcome.
type SpeakingPair struct {
	A, B   Value // side A's and side B's value; one may be negative
	Result valuecompare.Result
	Detail valuecompare.Detail
}

// Excluded is a value dropped as non-evidence.
type Excluded struct {
	Side   valuecompare.Side
	Value  Value
	Reason ExcludeReason
}

// speakingOrder is the order outcomes speak in: a pair that agrees speaks
// before one that partly agrees, then one that conflicts, then one that cannot
// say.
var speakingOrder = []Outcome{OutcomeAgree, OutcomePartial, OutcomeConflict, OutcomeUnknown}

// outcomeOf is the one relation to outcome mapping.
func outcomeOf(r valuecompare.Relation) Outcome {
	switch r {
	case valuecompare.Same, valuecompare.Equivalent:
		return OutcomeAgree
	case valuecompare.Contains, valuecompare.Overlaps, valuecompare.Near:
		return OutcomePartial
	case valuecompare.Different:
		return OutcomeConflict
	}
	return OutcomeUnknown
}

// keyed is a value with its Same key, computed once.
type keyed struct {
	v      Value
	key    string
	hasKey bool
}

// pair is two values compared once, side A's first.
type pair struct {
	a, b keyed
	res  valuecompare.Result
	det  valuecompare.Detail
}

// Compare decides whether the Subjects whose values are a and b agree on a
// property of the given meta. Each pair is compared with layer 1 under rules,
// once, in one pass that keeps the best contradiction and the best positive
// pair so far.
func Compare(a, b []Value, meta propertymeta.Meta, rules valuecompare.Rules) Result {
	var out Result
	va, posA := split(a, valuecompare.SideA, meta, rules, &out)
	vb, posB := split(b, valuecompare.SideB, meta, rules, &out)

	// Side A's values in input order and, for each, side B's: the first pair
	// offered wins a tie.
	var contra, positive *pair
	for _, x := range va {
		for _, y := range vb {
			if x.v.Negative && y.v.Negative {
				continue
			}
			res, det := valuecompare.Compare(x.v.Value, y.v.Value, rules)
			p := pair{a: x, b: y, res: res, det: det}
			switch {
			case x.v.Negative || y.v.Negative:
				if contradicts(p) && (contra == nil || better(p, *contra)) {
					contra = &p
				}
			case positive == nil || better(p, *positive):
				positive = &p
			}
		}
	}

	if contra != nil {
		out.Outcome, out.Cause, out.Speaking = OutcomeConflict, CauseNegative, contra.speaking()
		return out
	}

	if posA == 0 || posB == 0 {
		out.Outcome, out.Cause = OutcomeUnknown, CauseNoValues
		return out
	}
	out.Speaking = positive.speaking()
	out.Outcome, out.Cause = outcomeOf(positive.res.Relation), CauseRelation
	if out.Outcome == OutcomeConflict && meta.Cardinality == propertymeta.CardinalityMultiple {
		out.Outcome, out.Cause = OutcomeUnknown, CauseCardinality
	}
	return out
}

// split sets each value's type, drops neutral values into out.Excluded and
// returns the rest in input order with the count of positive ones.
func split(values []Value, side valuecompare.Side, meta propertymeta.Meta, rules valuecompare.Rules, out *Result) (kept []keyed, positives int) {
	for _, v := range values {
		v.Type = meta.ValueType
		if v.Neutral {
			out.Excluded = append(out.Excluded, Excluded{Side: side, Value: v, Reason: ExcludeNeutral})
			continue
		}
		k, ok := valuecompare.Key(v.Value, valuecompare.Same, rules)
		kept = append(kept, keyed{v: v, key: k, hasKey: ok})
		if v.Positive() {
			positives++
		}
	}
	return kept, positives
}

// contradicts reports whether a negative value matches the value it denies.
func contradicts(p pair) bool {
	switch p.res.Relation {
	case valuecompare.Same, valuecompare.Equivalent, valuecompare.Contains:
		return true
	}
	return false
}

func (p *pair) speaking() *SpeakingPair {
	return &SpeakingPair{A: p.a.v, B: p.b.v, Result: p.res, Detail: p.det}
}

// better reports whether x speaks before y: by outcome, then similarity, then
// the stronger relation, then each side's Same key. Refs are never read.
func better(x, y pair) bool {
	ox := slices.Index(speakingOrder, outcomeOf(x.res.Relation))
	oy := slices.Index(speakingOrder, outcomeOf(y.res.Relation))
	if ox != oy {
		return ox < oy
	}
	if x.res.Similarity != y.res.Similarity {
		return x.res.Similarity > y.res.Similarity
	}
	if x.res.Relation != y.res.Relation {
		return valuecompare.Stronger(x.res.Relation, y.res.Relation)
	}
	if x.a.key != y.a.key || x.a.hasKey != y.a.hasKey {
		return keyBefore(x.a, y.a)
	}
	if x.b.key != y.b.key || x.b.hasKey != y.b.hasKey {
		return keyBefore(x.b, y.b)
	}
	return false
}

// keyBefore sorts keys ascending, a value with no key after any value with one.
func keyBefore(x, y keyed) bool {
	if x.hasKey != y.hasKey {
		return x.hasKey
	}
	return x.key < y.key
}
