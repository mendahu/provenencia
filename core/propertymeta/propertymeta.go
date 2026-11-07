// Package propertymeta holds what matching and the catalog share about a
// property: the value-type and cardinality constants and Meta. It is a pure
// value package: queries live in core/database/properties.
package propertymeta

// Meta is what matching reads about a property.
type Meta struct {
	ValueType   string // a ValueType* constant
	Cardinality string // a Cardinality* constant; empty reads as single
}

// CardinalityOrSingle is m's cardinality, with empty read as single.
func (m Meta) CardinalityOrSingle() string {
	if m.Cardinality == "" {
		return CardinalitySingle
	}
	return m.Cardinality
}

const (
	ValueTypeText    = "text"
	ValueTypeInteger = "integer"
	ValueTypeDate    = "date"
	ValueTypeName    = "name"
	ValueTypeSubject = "subject"
	ValueTypeTerm    = "term"
	// ValueTypeDepiction's value is a depiction_values row: a region of the
	// Citation's page and the image File Go stored from it.
	ValueTypeDepiction = "depiction"
	// ValueTypeBoolean is a yes/no fact, stored as observations.value_boolean.
	ValueTypeBoolean = "boolean"

	// CardinalitySingle aims for one value. Several survivors are mixed.
	CardinalitySingle = "single"
	// CardinalityMultiple keeps every distinct surviving value. Majority
	// does not crowd one out; weak evidence still drops.
	CardinalityMultiple = "multiple"
)

// TypeRule is what the catalog and the app share about one value type.
type TypeRule struct {
	ValueType string // a ValueType* constant
	// ResearcherCreatable is whether Create Property accepts the type. The
	// others are seeded or registry-only.
	ResearcherCreatable bool
	// SingleOnly is whether a Property of the type must be single-valued: two
	// answers to a yes/no are a conflict to show, not two values to keep.
	SingleOnly bool
}

// Types declares every value type once, in the order the app lists them.
// Boolean is not creatable until S14-37 adds its editor.
var Types = []TypeRule{
	{ValueType: ValueTypeText, ResearcherCreatable: true},
	{ValueType: ValueTypeInteger, ResearcherCreatable: true},
	{ValueType: ValueTypeDate, ResearcherCreatable: true},
	{ValueType: ValueTypeName, ResearcherCreatable: true},
	{ValueType: ValueTypeSubject, ResearcherCreatable: true},
	{ValueType: ValueTypeTerm},
	{ValueType: ValueTypeDepiction},
	{ValueType: ValueTypeBoolean, SingleOnly: true},
}

// ValueTypes lists every value type in Types order.
func ValueTypes() []string {
	out := make([]string, len(Types))
	for i, r := range Types {
		out[i] = r.ValueType
	}
	return out
}

// Rule returns the rule for a value type, or false when it is not one.
func Rule(valueType string) (TypeRule, bool) {
	for _, r := range Types {
		if r.ValueType == valueType {
			return r, true
		}
	}
	return TypeRule{}, false
}
