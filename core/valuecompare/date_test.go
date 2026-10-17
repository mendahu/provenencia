package valuecompare_test

import "github.com/mendahu/provenencia/core/datevalue"

// Date builders shared by the date tests. A month or day of 0 is not named.

func optInt(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

func ymd(y, m, d int) datevalue.Value {
	return datevalue.Value{
		Kind: datevalue.KindPoint, Calendar: "gregorian",
		StartYear: optInt(y), StartMonth: optInt(m), StartDay: optInt(d),
	}
}

func about(y, m, d int) datevalue.Value {
	v := ymd(y, m, d)
	v.Qualifier = datevalue.QualifierABT
	return v
}

// bound is a year with a qualifier: datevalue.QualifierBEF or QualifierAFT.
func bound(qualifier string, y int) datevalue.Value {
	v := ymd(y, 0, 0)
	v.Qualifier = qualifier
	return v
}

func between(y1, m1, d1, y2, m2, d2 int) datevalue.Value {
	return datevalue.Value{
		Kind: datevalue.KindRange, Calendar: "gregorian",
		StartYear: optInt(y1), StartMonth: optInt(m1), StartDay: optInt(d1),
		EndYear: optInt(y2), EndMonth: optInt(m2), EndDay: optInt(d2),
	}
}

func phrase(s string) datevalue.Value {
	return datevalue.Value{Kind: datevalue.KindPoint, Phrase: s}
}

// noYearRange is a range with no year on either side.
func noYearRange() datevalue.Value {
	return datevalue.Value{Kind: datevalue.KindRange, StartMonth: optInt(3), EndMonth: optInt(6)}
}

func dp(v datevalue.Value) *datevalue.Value { return &v }
