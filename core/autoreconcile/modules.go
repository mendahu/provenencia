package autoreconcile

import (
	"strconv"
	"strings"

	"github.com/mendahu/provenencia/core/database/properties"
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

// NeutralTermKeys are term keys that are no evidence either way: an
// "unknown" sex at birth never wins or outvotes a recorded one.
var NeutralTermKeys = map[string]bool{"unknown": true, "indeterminate": true}

func moduleFor(valueType string) module {
	switch valueType {
	case properties.ValueTypeText:
		return keyModule{keyOf: textKey}
	case properties.ValueTypeInteger:
		return keyModule{keyOf: func(v Value) (string, bool) { return strconv.FormatInt(v.Integer, 10), true }}
	case properties.ValueTypeTerm:
		return keyModule{keyOf: termKey}
	case properties.ValueTypeSubject:
		// Interim: the subject itself. S9-28 maps it to its handle.
		return keyModule{keyOf: func(v Value) (string, bool) { return string(v.SubjectID), true }}
	case properties.ValueTypeName:
		return nameModule{}
	case properties.ValueTypeDate:
		// Interim: every structured field, as S9-05. S9-21 reconciles windows.
		return keyModule{keyOf: func(v Value) (string, bool) { return dateKey(v.Date), true }}
	}
	return nil
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

// textKey: trimmed and case-insensitive (York = york). Blank text is no
// evidence.
func textKey(v Value) (string, bool) {
	k := strings.ToLower(strings.TrimSpace(v.Text))
	return k, k != ""
}

// termKey: the same term. A neutral term is no evidence.
func termKey(v Value) (string, bool) {
	if NeutralTermKeys[v.TermKey] {
		return "", false
	}
	return string(v.TermID), true
}

// Compatible reports whether two values of valueType are the same value as
// the pipeline would see it: every unit both carry is equal, or one folds
// into the other (J. Robins and James Robins). A unit only one carries does
// not count against them, as in the pipeline's grouping; values that share
// no unit, or where either is no evidence, are not compatible. Promote
// alignment counts a comparison as an agreement by this test (S9-41).
func Compatible(valueType string, a, b Value) bool {
	if !knownValueType(valueType) || !carries(valueType, a) || !carries(valueType, b) {
		return false
	}
	m := moduleFor(valueType)
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
