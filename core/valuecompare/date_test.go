package valuecompare_test

import (
	"math/rand"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

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

// dateValue wraps a date for Compare; nil is no date.
func dateValue(d *datevalue.Value) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeDate, Date: d}
}

// ofDate wraps two dates as a sample pair; a nil pointer is no date.
func ofDate(a, b *datevalue.Value) samplePair { return samplePair{dateValue(a), dateValue(b)} }

type dateCase struct {
	name     string
	a, b     *datevalue.Value
	rel      valuecompare.Relation
	sim      float64
	cause    valuecompare.Cause
	broader  valuecompare.Side
	distance float64
}

func TestDateCompare(t *testing.T) {
	bef, aft := datevalue.QualifierBEF, datevalue.QualifierAFT
	span := func(w float64) float64 { return 0.8 - 0.05*(w-1) }
	same, contains, overlaps, nearRel, different, unknown := valuecompare.Same, valuecompare.Contains, valuecompare.Overlaps, valuecompare.Near, valuecompare.Different, valuecompare.Unknown
	a, b, none := valuecompare.SideA, valuecompare.SideB, valuecompare.SideNone
	exact, precision, overlap, gap, beyond := valuecompare.CauseExact, valuecompare.CausePrecision, valuecompare.CauseOverlap, valuecompare.CauseGap, valuecompare.CauseBeyondTolerance
	cases := []dateCase{
		{"AC-1.1 year", dp(ymd(1817, 0, 0)), dp(ymd(1817, 0, 0)), same, 1, exact, none, 0},
		{"AC-1.1 before", dp(bound(bef, 1820)), dp(bound(bef, 1820)), same, 1, exact, none, 0},
		{"AC-1.1 about", dp(about(1817, 0, 0)), dp(about(1817, 0, 0)), same, 1, exact, none, 0},
		{"AC-1.1 range of one year", dp(between(1817, 0, 0, 1817, 0, 0)), dp(ymd(1817, 0, 0)), same, 1, exact, none, 0},
		{"AC-1.1 inverted range", dp(between(1820, 0, 0, 1815, 0, 0)), dp(between(1815, 0, 0, 1820, 0, 0)), same, 1, exact, none, 0},
		{"AC-1.2 missing day", dp(ymd(1817, 5, 0)), dp(ymd(1817, 5, 14)), contains, 0.9, precision, a, 0},
		{"AC-1.2 missing month", dp(ymd(1817, 0, 0)), dp(ymd(1817, 9, 0)), contains, 0.8, precision, a, 0},
		{"AC-1.3 about next year", dp(about(1817, 0, 0)), dp(ymd(1818, 0, 0)), contains, span(3), precision, a, 0},
		{"AC-1.3 about same year", dp(about(1817, 0, 0)), dp(ymd(1817, 0, 0)), contains, span(3), precision, a, 0},
		{"AC-1.4 before a wider before", dp(bound(bef, 1820)), dp(bound(bef, 1830)), contains, 0.2, precision, b, 0},
		{"AC-1.4 before a point", dp(bound(bef, 1820)), dp(ymd(1815, 0, 0)), contains, 0.2, precision, a, 0},
		{"AC-1.5 range holds a year", dp(between(1815, 0, 0, 1820, 0, 0)), dp(ymd(1818, 0, 0)), contains, span(6), precision, a, 0},
		{"AC-1.6 about month", dp(about(1817, 3, 0)), dp(ymd(1817, 4, 0)), contains, 0.8, precision, a, 0},
		{"AC-1.6 about day", dp(about(1817, 3, 3)), dp(ymd(1817, 3, 4)), contains, 0.8, precision, a, 0},
		{"AC-1.6 about month over a year end", dp(about(1817, 1, 0)), dp(ymd(1816, 12, 15)), contains, 0.8, precision, a, 0},
		{"AC-1.6 about day over a year end", dp(about(1817, 1, 1)), dp(ymd(1816, 12, 31)), contains, 0.8, precision, a, 0},
		{"AC-1.7 overlapping ranges", dp(between(1815, 0, 0, 1820, 0, 0)), dp(between(1819, 0, 0, 1825, 0, 0)), overlaps, span(7), overlap, none, 0},
		{"AC-1.8 other day", dp(ymd(1817, 5, 14)), dp(ymd(1817, 5, 20)), nearRel, 0.7, gap, none, 6},
		{"AC-1.8 other month", dp(ymd(1817, 5, 0)), dp(ymd(1817, 9, 0)), nearRel, 0.6, gap, none, 93},
		{"AC-1.9 next year", dp(ymd(1817, 0, 0)), dp(ymd(1818, 0, 0)), nearRel, 0.8 * (1 - 1.0/3), gap, none, 1},
		{"AC-1.9 across a year end", dp(ymd(1817, 12, 0)), dp(ymd(1818, 1, 0)), nearRel, 0.8 * (1 - (31.0/365)/3), gap, none, 1},
		{"AC-1.10 after a point", dp(bound(aft, 1820)), dp(ymd(1819, 0, 0)), nearRel, 0.2 * (1 - 1.0/3), gap, none, 1},
		{"AC-1.10 range and a point", dp(between(1815, 0, 0, 1820, 0, 0)), dp(ymd(1822, 0, 0)), nearRel, span(6) * (1 - 2.0/3), gap, none, 366},
		{"AC-1.10 two abouts", dp(about(1817, 0, 0)), dp(about(1820, 0, 0)), nearRel, span(3) * (1 - 1.0/3), gap, none, 1},
		{"AC-1.11 three years", dp(ymd(1817, 0, 0)), dp(ymd(1820, 0, 0)), different, 0, beyond, none, 731},
		{"AC-1.11 before and three years", dp(bound(bef, 1820)), dp(ymd(1823, 0, 0)), different, 0, beyond, none, 731},
		{"AC-1.11 before and after", dp(bound(bef, 1800)), dp(bound(aft, 1810)), different, 0, beyond, none, 3288},
		{"AC-1.12 phrase", dp(phrase("spring")), dp(ymd(1817, 0, 0)), unknown, 0, valuecompare.CausePhraseOnly, none, 0},
		{"AC-1.12 range with no year", dp(noYearRange()), dp(ymd(1817, 0, 0)), unknown, 0, valuecompare.CauseEmpty, none, 0},
		{"AC-1.12 no date", nil, dp(ymd(1817, 0, 0)), unknown, 0, valuecompare.CauseEmpty, none, 0},
		{"AC-1.13 day with no month", dp(datevalue.Value{Kind: datevalue.KindPoint, StartYear: optInt(1817), StartDay: optInt(14)}),
			dp(datevalue.Value{Kind: datevalue.KindPoint, StartYear: optInt(1817), StartDay: optInt(20)}), same, 1, exact, none, 0},
	}
	rules := matchregistry.Default().Value
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, swap := range []bool{false, true} {
				x, y, broader := tc.a, tc.b, tc.broader
				if swap {
					x, y = y, x
					if broader != none {
						broader = map[valuecompare.Side]valuecompare.Side{a: b, b: a}[broader]
					}
				}
				res, det := valuecompare.Compare(dateValue(x), dateValue(y), rules)
				if res.Relation != tc.rel || !near(res.Similarity, tc.sim) {
					t.Fatalf("swap=%v: %s at %v, want %s at %v", swap, res.Relation, res.Similarity, tc.rel, tc.sim)
				}
				// An unknown pair names its cause from the first side that failed.
				if tc.rel != unknown || !swap {
					if det.Cause != tc.cause {
						t.Fatalf("swap=%v: cause %s, want %s", swap, det.Cause, tc.cause)
					}
				}
				if det.Broader != broader {
					t.Fatalf("swap=%v: broader %q, want %q", swap, det.Broader, broader)
				}
				if det.Distance != tc.distance {
					t.Fatalf("swap=%v: distance %v, want %v", swap, det.Distance, tc.distance)
				}
				wantUnit := valuecompare.Unit("")
				if tc.rel == nearRel || tc.rel == different {
					wantUnit = valuecompare.UnitDays
				}
				if det.Unit != wantUnit || len(det.Parts) != 0 {
					t.Fatalf("swap=%v: unit %q, %d parts", swap, det.Unit, len(det.Parts))
				}
			}
		})
	}
	t.Run("AC-1.14 tolerance 0", func(t *testing.T) {
		r := matchregistry.Default().Value
		r.Dates.Tolerance = 0
		res, det := valuecompare.Compare(dateValue(dp(ymd(1817, 0, 0))), dateValue(dp(ymd(1818, 0, 0))), r)
		if res.Relation != different || det.Cause != beyond {
			t.Fatalf("%+v %+v", res, det)
		}
	})
	t.Run("below the near minimum", func(t *testing.T) {
		r := matchregistry.Default().Value
		r.Dates.NearMinimum = 0.6
		res, det := valuecompare.Compare(dateValue(dp(ymd(1817, 0, 0))), dateValue(dp(ymd(1818, 0, 0))), r)
		if res.Relation != different || res.Similarity != 0 || det.Cause != valuecompare.CauseBelowMinimum || det.Distance != 1 {
			t.Fatalf("%+v %+v", res, det)
		}
	})
}

func TestDateKey(t *testing.T) {
	rules := matchregistry.Default().Value
	key := func(d datevalue.Value, level valuecompare.Relation) (string, bool) {
		return valuecompare.Key(dateValue(&d), level, rules)
	}
	// AC-2.3
	k1, ok1 := key(ymd(1817, 0, 0), valuecompare.Same)
	k2, ok2 := key(between(1817, 1, 0, 1817, 12, 0), valuecompare.Same)
	if !ok1 || !ok2 || k1 != k2 {
		t.Errorf("1817 keys %q (%v) and %q (%v)", k1, ok1, k2, ok2)
	}
	if k, ok := key(ymd(1817, 0, 0), valuecompare.Equivalent); !ok || k != k1 {
		t.Errorf("Equivalent key = %q, %v", k, ok)
	}
	if k, ok := key(bound(datevalue.QualifierBEF, 1820), valuecompare.Same); !ok || k != "*..-54422" {
		t.Errorf("BEF 1820 key = %q, %v", k, ok)
	}
	if k, ok := key(phrase("spring"), valuecompare.Same); ok || k != "" {
		t.Errorf("phrase key = %q, %v", k, ok)
	}
	if k, ok := key(ymd(1817, 0, 0), valuecompare.Near); ok || k != "" {
		t.Errorf("Near key = %q, %v", k, ok)
	}
	if k, ok := valuecompare.Key(dateValue(nil), valuecompare.Same, rules); ok || k != "" {
		t.Errorf("no date key = %q, %v", k, ok)
	}
}

func TestDateLadder(t *testing.T) {
	rules := matchregistry.Default().Value
	base := dp(ymd(1817, 5, 14))
	// 14 May 1817 is day 134 of 365; its centre is 133.5 days in.
	centre := 133.5 / 365
	ladder := []struct {
		name string
		d    *datevalue.Value
		rel  valuecompare.Relation
		sim  float64
	}{
		{"14 MAY 1817", dp(ymd(1817, 5, 14)), valuecompare.Same, 1},
		{"MAY 1817", dp(ymd(1817, 5, 0)), valuecompare.Contains, 0.9},
		{"1817", dp(ymd(1817, 0, 0)), valuecompare.Contains, 0.8},
		{"20 MAY 1817", dp(ymd(1817, 5, 20)), valuecompare.Near, 0.7},
		{"SEP 1817", dp(ymd(1817, 9, 0)), valuecompare.Near, 0.6},
		{"BET 1815 AND 1820", dp(between(1815, 0, 0, 1820, 0, 0)), valuecompare.Contains, 0.8 - 0.05*5},
		{"1818", dp(ymd(1818, 0, 0)), valuecompare.Near, 0.8 * (1 - (1.5-centre)/3)},
		{"1819", dp(ymd(1819, 0, 0)), valuecompare.Near, 0.8 * (1 - (2.5-centre)/3)},
		{"BEF 1820", dp(bound(datevalue.QualifierBEF, 1820)), valuecompare.Contains, 0.2},
		{"1820", dp(ymd(1820, 0, 0)), valuecompare.Different, 0},
	}
	prev := 2.0
	for _, step := range ladder {
		res, _ := valuecompare.Compare(dateValue(base), dateValue(step.d), rules)
		if res.Relation != step.rel || !near(res.Similarity, step.sim) {
			t.Errorf("%s: %s at %v, want %s at %v", step.name, res.Relation, res.Similarity, step.rel, step.sim)
		}
		if res.Similarity >= prev {
			t.Errorf("%s: %v does not fall below %v", step.name, res.Similarity, prev)
		}
		prev = res.Similarity
	}
}

func TestDateInvariants(t *testing.T) {
	for seed := int64(1); seed <= 4; seed++ {
		rules := matchregistry.Default().Value
		rules.Dates.Tolerance = int(seed - 1)
		wider := rules
		wider.Dates.AboutUnits = rules.Dates.AboutUnits + 1
		looser := rules
		looser.Dates.Tolerance = rules.Dates.Tolerance + 1
		rng := rand.New(rand.NewSource(seed))
		isAbout := func(d *datevalue.Value) bool {
			return d != nil && d.Kind != datevalue.KindRange && d.Qualifier == datevalue.QualifierABT
		}
		for range 2000 {
			p := generators[propertymeta.ValueTypeDate](rng)
			res, det := valuecompare.Compare(p.a, p.b, rules)
			// Widening ABT keeps what it contained.
			if res.Relation == valuecompare.Contains &&
				((det.Broader == valuecompare.SideA && isAbout(p.a.Date)) || (det.Broader == valuecompare.SideB && isAbout(p.b.Date))) {
				res2, det2 := valuecompare.Compare(p.a, p.b, wider)
				if res2.Relation != valuecompare.Contains || det2.Broader != det.Broader {
					t.Fatalf("seed %d: %+v / %+v: %s (broader %q) became %s (broader %q) with wider ABT", seed, *p.a.Date, *p.b.Date, res.Relation, det.Broader, res2.Relation, det2.Broader)
				}
			}
			// More tolerance never turns Near into Different.
			if res.Relation == valuecompare.Near {
				if res3, _ := valuecompare.Compare(p.a, p.b, looser); res3.Relation == valuecompare.Different {
					t.Fatalf("seed %d: %+v / %+v: Near became Different with more tolerance", seed, *p.a.Date, *p.b.Date)
				}
			}
			// A plain point is never broader than the same point under ABT.
			for _, d := range []*datevalue.Value{p.a.Date, p.b.Date} {
				if d.Kind == datevalue.KindRange || d.Qualifier != "" || d.StartYear == nil {
					continue
				}
				about := *d
				about.Qualifier = datevalue.QualifierABT
				res4, det4 := valuecompare.Compare(dateValue(d), dateValue(&about), rules)
				if res4.Relation != valuecompare.Contains || det4.Broader != valuecompare.SideB {
					t.Fatalf("seed %d: %+v against its ABT: %s, broader %q", seed, *d, res4.Relation, det4.Broader)
				}
			}
		}
	}
}
