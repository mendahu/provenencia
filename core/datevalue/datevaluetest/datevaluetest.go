// Package datevaluetest builds date values for test fixtures.
package datevaluetest

import "github.com/mendahu/provenencia/core/datevalue"

// Part is a date part for a datevalue.Value literal: nil for 0.
func Part(n int) *int {
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
		StartYear: Part(y), StartMonth: Part(m), StartDay: Part(d),
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

// Between is a Gregorian range date. A 0 month or day is left unset.
func Between(y1, m1, d1, y2, m2, d2 int) *datevalue.Value {
	return &datevalue.Value{
		Kind: datevalue.KindRange, Calendar: "gregorian",
		StartYear: Part(y1), StartMonth: Part(m1), StartDay: Part(d1),
		EndYear: Part(y2), EndMonth: Part(m2), EndDay: Part(d2),
	}
}
