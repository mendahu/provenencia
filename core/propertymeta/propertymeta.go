// Package propertymeta holds the property value-type and cardinality
// constants. It is a pure value package: queries live in
// core/database/properties.
package propertymeta

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
