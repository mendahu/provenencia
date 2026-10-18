// Package datevaluetest builds date values for test fixtures.
package datevaluetest

import "github.com/mendahu/provenencia/core/datevalue"

func optInt(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

// YMD is a Gregorian point date. A 0 month or day is left unset: YMD(1817, 0, 0)
// is the year 1817.
func YMD(y, m, d int) *datevalue.Value {
	return &datevalue.Value{
		Kind: datevalue.KindPoint, Calendar: "gregorian",
		StartYear: optInt(y), StartMonth: optInt(m), StartDay: optInt(d),
	}
}

// About is YMD with the ABT qualifier.
func About(y, m, d int) *datevalue.Value {
	v := YMD(y, m, d)
	v.Qualifier = datevalue.QualifierABT
	return v
}

// Phrase is a point date that carries only a phrase.
func Phrase(s string) *datevalue.Value {
	return &datevalue.Value{Kind: datevalue.KindPoint, Phrase: s}
}
