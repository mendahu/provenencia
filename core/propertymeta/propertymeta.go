// Package propertymeta holds what matching and the catalog share about a
// property: the value-type and cardinality constants and Meta. It is a pure
// value package: queries live in core/database/properties.
package propertymeta

// Meta is what matching reads about a property.
type Meta struct {
	ValueType   string // a ValueType* constant
	Cardinality string // a Cardinality* constant; empty reads as single
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

	// CardinalitySingle aims for one value. Several survivors are mixed.
	CardinalitySingle = "single"
	// CardinalityMultiple keeps every distinct surviving value. Majority
	// does not crowd one out; weak evidence still drops.
	CardinalityMultiple = "multiple"
)
