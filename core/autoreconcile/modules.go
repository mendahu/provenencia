package autoreconcile

import (
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

// A module is what one value type supplies to the shared pipeline
// (conclusion-reconciliation.md §7): when two values are the same, what folds
// into what, and how a group becomes one displayed value. Everything else —
// denial, majority, confidence, ordering, reasons — is the pipeline's.

// unit is one comparable piece of a value. A simple module splits a value
// into one unit; a structured one (names, S9-13b) into one per part type, and
// the pipeline votes on each unit separately.
type unit struct {
	key  string // equal keys are the same value
	data any    // the module's own payload, for fold and assemble
}

type module interface {
	// split reads a value's units by unit name. ok is false when the value
	// carries nothing to compare (no evidence).
	split(v Value) (units map[string]unit, ok bool)
	// fold reports whether from, a less specific value, folds into into, a
	// fuller one, both units of name. Simple modules never fold.
	fold(name string, from, into unit) bool
	// outvotes reports whether majority may crowd other out when winner wins
	// unit name. Simple modules let majority outvote anything; names only
	// outvote spelling variants of the winner.
	outvotes(name string, winner, other unit) bool
	// oneValue reports whether every displayed candidate forms a single
	// value (names: one structure, several values per part type), rather
	// than one value per agreeing group.
	oneValue() bool
	// assemble builds a value's displayed form from its supporting members
	// in rank order and, per unit name, the values it settled on (best
	// supported first). A module without oneValue gets one value per name.
	assemble(members []Candidate, settled map[string][]unit) Value
}

// wholeValue is the unit name of a module with one unit per value.
const wholeValue = ""

// moduleFor is valueType's module under rules, nil for an unknown type.
// Values are the same, and fold, by value compare's relations: Same groups,
// Contains folds, and no other relation does (D-2 of S11-06).
func moduleFor(rules valuecompare.Rules, valueType string) module {
	switch valueType {
	case propertymeta.ValueTypeText, propertymeta.ValueTypeInteger:
		return keyModule{keyOf: func(v Value) (string, bool) {
			return valuecompare.Key(v.compareValue(valueType), valuecompare.Same, rules)
		}}
	case propertymeta.ValueTypeTerm:
		return keyModule{keyOf: termKey}
	case propertymeta.ValueTypeSubject:
		// The accepted handle. The loader fills EntityID; an unpromoted or
		// missing end has none and is no evidence.
		return keyModule{keyOf: func(v Value) (string, bool) { return idKey(v.EntityID) }}
	case propertymeta.ValueTypeDepiction:
		// The same image File, so the same pixels; no readable image is no evidence.
		return keyModule{keyOf: func(v Value) (string, bool) {
			if v.Depiction == nil {
				return "", false
			}
			return idKey(v.Depiction.ImageFileID)
		}}
	case propertymeta.ValueTypeName:
		particles := make(map[string]bool, len(rules.Words.Particles))
		for _, p := range rules.Words.Particles {
			particles[p] = true
		}
		return nameModule{rules: rules, particles: particles}
	case propertymeta.ValueTypeDate:
		return dateModule{rules: rules}
	}
	return nil
}

// compareValue is v in the shape value compare reads: the text, integer or
// date of a value of valueType. Names go through nameUnit.compareValue.
func (v Value) compareValue(valueType string) valuecompare.Value {
	return valuecompare.Value{Type: valueType, Text: v.Text, Integer: v.Integer, Date: v.Date}
}

// keyModule is a value type whose values are the same exactly when their
// keys are equal: one unit, no folding, and the best-ranked member's value
// displayed.
type keyModule struct {
	keyOf func(Value) (key string, ok bool)
}

func (m keyModule) split(v Value) (map[string]unit, bool) {
	k, ok := m.keyOf(v)
	if !ok {
		return nil, false
	}
	return map[string]unit{wholeValue: {key: k}}, true
}

func (keyModule) fold(string, unit, unit) bool { return false }

func (keyModule) outvotes(string, unit, unit) bool { return true }

func (keyModule) oneValue() bool { return false }

func (keyModule) assemble(members []Candidate, _ map[string][]unit) Value {
	return members[0].Value
}

// termKey: the same term. A neutral term is no evidence: an "unknown" sex at
// birth never wins or outvotes a recorded one.
func termKey(v Value) (string, bool) {
	if v.Neutral {
		return "", false
	}
	return string(v.TermID), true
}

// idKey: an id is its own key; no id is no evidence.
func idKey(id []byte) (string, bool) { return string(id), len(id) > 0 }

// Compatible reports whether two values of valueType are the same value as
// the pipeline would see it: every unit both carry is equal, or one folds
// into the other (J. Robins and James Robins). A unit only one carries does
// not count against them, as in the pipeline's grouping; values that share
// no unit, or where either is no evidence, are not compatible. Promote
// graph alignment counts a comparison as an agreement by this test (S9-41).
func Compatible(rules valuecompare.Rules, valueType string, a, b Value) bool {
	m := moduleFor(rules, valueType)
	if m == nil || !Carries(valueType, a) || !Carries(valueType, b) {
		return false
	}
	ua, ok := m.split(a)
	if !ok {
		return false
	}
	ub, ok := m.split(b)
	if !ok {
		return false
	}
	shared := false
	for name, x := range ua {
		y, ok := ub[name]
		if !ok {
			continue
		}
		shared = true
		if x.key != y.key && !m.fold(name, x, y) && !m.fold(name, y, x) {
			return false
		}
	}
	return shared
}
