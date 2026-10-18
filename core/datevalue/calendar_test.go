package datevalue_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
)

func ip(v int) *int { return &v }

func TestCalendar(t *testing.T) {
	if got := datevalue.DaysFromCivil(1970, 1, 1); got != 0 {
		t.Errorf("DaysFromCivil(1970, 1, 1) = %d, want 0", got)
	}
	t.Run("AddDays round trip", func(t *testing.T) {
		for n := datevalue.DaysFromCivil(1600, 1, 1); n <= datevalue.DaysFromCivil(2100, 12, 31); n++ {
			y, m, d := datevalue.AddDays(1970, 1, 1, n)
			if got := datevalue.DaysFromCivil(y, m, d); got != n {
				t.Fatalf("day %d came back as %d-%d-%d (%d)", n, y, m, d, got)
			}
			y2, m2, d2 := datevalue.AddDays(y, m, d, 1)
			y3, m3, d3 := datevalue.AddDays(y2, m2, d2, -1)
			if y3 != y || m3 != m || d3 != d {
				t.Fatalf("%d-%d-%d +1 -1 = %d-%d-%d", y, m, d, y3, m3, d3)
			}
			if datevalue.DaysFromCivil(y2, m2, d2) != n+1 {
				t.Fatalf("%d-%d-%d +1 = %d-%d-%d", y, m, d, y2, m2, d2)
			}
		}
	})
	for _, tc := range []struct{ year, month, want int }{
		{1900, 2, 28}, {2000, 2, 29}, {1816, 2, 29}, {1817, 2, 28}, {1817, 4, 30}, {1817, 12, 31},
	} {
		if got := datevalue.DaysInMonth(tc.year, tc.month); got != tc.want {
			t.Errorf("DaysInMonth(%d, %d) = %d, want %d", tc.year, tc.month, got, tc.want)
		}
	}
	t.Run("BoundDays", func(t *testing.T) {
		for _, tc := range []struct {
			name             string
			year, month, day *int
			lo, hi           int
			ok               bool
		}{
			{"year", ip(1985), nil, nil, datevalue.DaysFromCivil(1985, 1, 1), datevalue.DaysFromCivil(1985, 12, 31), true},
			{"month", ip(1985), ip(5), nil, datevalue.DaysFromCivil(1985, 5, 1), datevalue.DaysFromCivil(1985, 5, 31), true},
			{"day", ip(1985), ip(5), ip(14), datevalue.DaysFromCivil(1985, 5, 14), datevalue.DaysFromCivil(1985, 5, 14), true},
			{"day with no month", ip(1985), nil, ip(14), datevalue.DaysFromCivil(1985, 1, 1), datevalue.DaysFromCivil(1985, 12, 31), true},
			{"no year", nil, ip(5), nil, 0, 0, false},
			{"month 13", ip(1985), ip(13), nil, 0, 0, false},
			{"30 February", ip(1985), ip(2), ip(30), 0, 0, false},
			{"day 0", ip(1985), ip(2), ip(0), 0, 0, false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				lo, hi, ok := datevalue.BoundDays(tc.year, tc.month, tc.day)
				if ok != tc.ok || (ok && (lo != tc.lo || hi != tc.hi)) {
					t.Fatalf("got %d %d %v, want %d %d %v", lo, hi, ok, tc.lo, tc.hi, tc.ok)
				}
			})
		}
	})
	t.Run("YearLine", func(t *testing.T) {
		const eps = 1e-12
		lo, hi, _ := datevalue.BoundDays(ip(1817), ip(5), nil)
		from, to, mid := datevalue.YearLine(lo, hi)
		// May is days 121 to 151 of 365.
		if d := from - (1817 + 120.0/365); d > eps || d < -eps {
			t.Errorf("from = %v", from)
		}
		if d := to - (1817 + 151.0/365); d > eps || d < -eps {
			t.Errorf("to = %v", to)
		}
		if d := mid - (from+to)/2; d > eps || d < -eps {
			t.Errorf("mid = %v", mid)
		}
	})
	if y, m := datevalue.AddMonths(1817, 1, -1); y != 1816 || m != 12 {
		t.Errorf("AddMonths(1817, 1, -1) = %d, %d", y, m)
	}
	if y, m := datevalue.AddMonths(1817, 12, 1); y != 1818 || m != 1 {
		t.Errorf("AddMonths(1817, 12, 1) = %d, %d", y, m)
	}
	if y, m := datevalue.AddMonths(1817, 3, -15); y != 1815 || m != 12 {
		t.Errorf("AddMonths(1817, 3, -15) = %d, %d", y, m)
	}
	if y, m, d := datevalue.AddDays(1817, 1, 1, -1); y != 1816 || m != 12 || d != 31 {
		t.Errorf("AddDays(1817, 1, 1, -1) = %d-%d-%d", y, m, d)
	}
}
