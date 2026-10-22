// Package subjectcomparetest holds Subject compare fixtures.
package subjectcomparetest

import (
	"maps"
	"slices"
	"strings"

	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Metas returns a new map of the seeded value type and cardinality of every
// product property the matching tests use. The subjectvocab seed test pins it
// to the seed.
func Metas() subjectcompare.Metas {
	single := propertymeta.CardinalitySingle
	return subjectcompare.Metas{
		vocabkey.PropertyName:       {ValueType: propertymeta.ValueTypeName, Cardinality: single},
		vocabkey.PropertySexAtBirth: {ValueType: propertymeta.ValueTypeTerm, Cardinality: single},
		vocabkey.PropertyGender:     {ValueType: propertymeta.ValueTypeTerm, Cardinality: single},
		vocabkey.PropertyEventType:  {ValueType: propertymeta.ValueTypeTerm, Cardinality: single},
		vocabkey.PropertyDate:       {ValueType: propertymeta.ValueTypeDate, Cardinality: single},
		vocabkey.PropertyStartDate:  {ValueType: propertymeta.ValueTypeDate, Cardinality: single},
		vocabkey.PropertyEndDate:    {ValueType: propertymeta.ValueTypeDate, Cardinality: single},
		vocabkey.PropertyToponym:    {ValueType: propertymeta.ValueTypeText, Cardinality: propertymeta.CardinalityMultiple},
		vocabkey.PropertyLikeness:   {ValueType: propertymeta.ValueTypeDepiction, Cardinality: propertymeta.CardinalityMultiple},
		"interests":                 {ValueType: propertymeta.ValueTypeText, Cardinality: propertymeta.CardinalityMultiple},
		"event_name":                {ValueType: propertymeta.ValueTypeText, Cardinality: single},
	}
}

// Labels names a Subject's values for a failure message and an expected value:
// each property in ascending key order as `key=` and its values'
// propertycomparetest.Labels, joined with `; `.
func Labels(v subjectcompare.Values) string {
	var out []string
	for _, p := range slices.Sorted(maps.Keys(v)) {
		out = append(out, string(p)+"="+pt.Labels(v[p]))
	}
	return strings.Join(out, "; ")
}
