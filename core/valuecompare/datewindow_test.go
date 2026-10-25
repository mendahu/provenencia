package valuecompare_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestDateEnds(t *testing.T) {
	const eps = 1e-12
	near := func(a, b float64) bool { return a-b < eps && b-a < eps }

	t.Run("year", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*datevaluetest.YMD(1817, 0, 0))
		lo, hi, _ := datevalue.BoundDays(datevaluetest.Part(1817), nil, nil)
		if !ok || first != last || first.Open || first.Year != 1817 || first.Month != 0 || first.Day != 0 ||
			first.From != 1817 || first.To != 1818 || first.Lo != lo || first.Hi != hi ||
			!near(first.Mid, (first.From+first.To)/2) {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("month", func(t *testing.T) {
		first, _, ok := valuecompare.DateEnds(*datevaluetest.YMD(1817, 5, 0))
		// May is days 121 to 151 of 365.
		if !ok || first.Month != 5 || first.Day != 0 || !near(first.From, 1817+120.0/365) || !near(first.To, 1817+151.0/365) {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	t.Run("leap month", func(t *testing.T) {
		first, _, ok := valuecompare.DateEnds(*datevaluetest.YMD(1816, 2, 0))
		// February 1816 is days 32 to 60 of 366.
		if !ok || !near(first.From, 1816+31.0/366) || !near(first.To, 1816+60.0/366) {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	t.Run("day", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*datevaluetest.YMD(1817, 5, 14))
		if !ok || first != last || first.Month != 5 || first.Day != 14 || first.Lo != first.Hi ||
			!near(first.From, 1817+133.0/365) || !near(first.To, 1817+134.0/365) {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	t.Run("range keeps its months", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*datevaluetest.Between(1817, 3, 0, 1817, 6, 0))
		if !ok || first.Month != 3 || last.Month != 6 || first.Open || last.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("before", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*bound(datevalue.QualifierBEF, 1820))
		if !ok || !first.Open || first.Year != 0 || last.Open || last.Year != 1820 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("after", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*bound(datevalue.QualifierAFT, 1820))
		if !ok || first.Open || !last.Open || last.Year != 0 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("range with no end", func(t *testing.T) {
		d := datevalue.Value{Kind: datevalue.KindRange, StartYear: datevaluetest.Part(1815)}
		first, last, ok := valuecompare.DateEnds(d)
		if !ok || first.Open || !last.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("range with no start", func(t *testing.T) {
		d := datevalue.Value{Kind: datevalue.KindRange, EndYear: datevaluetest.Part(1815)}
		first, last, ok := valuecompare.DateEnds(d)
		if !ok || !first.Open || last.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("inverted range swaps", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*datevaluetest.Between(1820, 0, 0, 1815, 0, 0))
		if !ok || first.Year != 1815 || last.Year != 1820 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("a month inside the end year does not swap", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*datevaluetest.Between(1817, 3, 0, 1817, 0, 0))
		if !ok || first.Month != 3 || last.Month != 0 || first.Lo >= last.Hi {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
		first, last, ok = valuecompare.DateEnds(*datevaluetest.Between(1817, 5, 14, 1817, 0, 0))
		if !ok || first.Month != 5 || first.Day != 14 || last.Month != 0 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("about is a point", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(*datevaluetest.About(1817, 0, 0))
		if !ok || first != last || first.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("empty kind is a point", func(t *testing.T) {
		d := datevalue.Value{StartYear: datevaluetest.Part(1817)}
		first, last, ok := valuecompare.DateEnds(d)
		if !ok || first != last {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("point with only end fields", func(t *testing.T) {
		d := datevalue.Value{Kind: datevalue.KindPoint, EndYear: datevaluetest.Part(1817)}
		first, _, ok := valuecompare.DateEnds(d)
		if !ok || first.Year != 1817 {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	for name, d := range map[string]datevalue.Value{
		"month 13":        *datevaluetest.YMD(1817, 13, 0),
		"day 31 in April": *datevaluetest.YMD(1817, 4, 31),
		"phrase only":     *datevaluetest.Phrase("spring"),
		"range, no year":  {Kind: datevalue.KindRange},
		"range, bad end":  *datevaluetest.Between(1815, 0, 0, 1817, 13, 0),
	} {
		t.Run("not ok: "+name, func(t *testing.T) {
			if _, _, ok := valuecompare.DateEnds(d); ok {
				t.Fatal("ok")
			}
		})
	}
}

func win(lo, hi *int) valuecompare.Window { return valuecompare.Window{Lo: lo, Hi: hi} }

func TestEndsDate(t *testing.T) {
	for name, d := range map[string]*datevalue.Value{
		"year":            datevaluetest.YMD(1817, 0, 0),
		"day":             datevaluetest.YMD(1817, 3, 3),
		"before":          bound(datevalue.QualifierBEF, 1820),
		"after":           bound(datevalue.QualifierAFT, 1815),
		"range":           datevaluetest.Between(1815, 0, 0, 1820, 0, 0),
		"year zero":       yearZero(nil),
		"range from zero": yearZero(datevaluetest.Part(5)),
	} {
		t.Run("round trip "+name, func(t *testing.T) {
			first, last, ok := valuecompare.DateEnds(*d)
			if !ok {
				t.Fatalf("no ends for %s", name)
			}
			got, ok := valuecompare.EndsDate(first, last)
			if !ok {
				t.Fatalf("EndsDate(%+v, %+v) not ok", first, last)
			}
			f2, l2, ok := valuecompare.DateEnds(got)
			if !ok || f2 != first || l2 != last {
				t.Fatalf("ends %+v %+v, want %+v %+v", f2, l2, first, last)
			}
		})
	}
	t.Run("both open", func(t *testing.T) {
		if _, ok := valuecompare.EndsDate(valuecompare.DateEnd{Open: true}, valuecompare.DateEnd{Open: true}); ok {
			t.Fatal("both open is ok")
		}
	})
	t.Run("first after last", func(t *testing.T) {
		first, _, _ := valuecompare.DateEnds(*datevaluetest.YMD(1821, 0, 0))
		_, last, _ := valuecompare.DateEnds(*datevaluetest.YMD(1820, 0, 0))
		if _, ok := valuecompare.EndsDate(first, last); ok {
			t.Fatal("first after last is ok")
		}
	})
}

func TestWindow(t *testing.T) {
	d := func(n int) *int { return &n }
	a, b, c := win(d(1), d(10)), win(d(5), d(20)), win(d(30), d(40))
	always := valuecompare.Always()
	if !a.Overlaps(b) || !b.Overlaps(a) {
		t.Error("overlapping windows do not overlap")
	}
	if a.Overlaps(c) || c.Overlaps(a) {
		t.Error("disjoint windows overlap")
	}
	if !always.Overlaps(c) || !c.Overlaps(always) {
		t.Error("an open window does not overlap every window")
	}
	if got, ok := a.Intersect(b); !ok || got.Lo == nil || *got.Lo != 5 || got.Hi == nil || *got.Hi != 10 {
		t.Errorf("a ∩ b = %+v, %v", got, ok)
	}
	if _, ok := a.Intersect(c); ok {
		t.Error("disjoint windows intersect")
	}
	if got, ok := always.Intersect(c); !ok || *got.Lo != 30 || *got.Hi != 40 {
		t.Errorf("always ∩ c = %+v, %v", got, ok)
	}
	if !a.Contains(a) {
		t.Error("a window does not contain itself")
	}
	if a.Contains(b) || b.Contains(a) {
		t.Error("overlapping windows contain each other")
	}
	if !always.Contains(a) || a.Contains(always) {
		t.Error("an open window should contain a closed one and not the reverse")
	}
	if !win(nil, d(10)).Contains(win(nil, d(5))) || win(nil, d(5)).Contains(win(nil, d(10))) {
		t.Error("open-low windows contain by their high end")
	}
	if !win(d(1), nil).Contains(win(d(5), nil)) || win(d(5), nil).Contains(win(d(1), nil)) {
		t.Error("open-high windows contain by their low end")
	}
}

func TestAboutEnds(t *testing.T) {
	day := datevalue.DaysFromCivil
	for _, tc := range []struct {
		name   string
		d      datevalue.Value
		lo, hi int
	}{
		{"about a year", *datevaluetest.About(1817, 0, 0), day(1816, 1, 1), day(1818, 12, 31)},
		{"about a month", *datevaluetest.About(1817, 3, 0), day(1817, 2, 1), day(1817, 4, 30)},
		{"about January", *datevaluetest.About(1817, 1, 0), day(1816, 12, 1), day(1817, 2, 28)},
		{"about a day", *datevaluetest.About(1817, 1, 1), day(1816, 12, 31), day(1817, 1, 2)},
		{"a plain year", *datevaluetest.YMD(1817, 0, 0), day(1817, 1, 1), day(1817, 12, 31)},
		{"a range is not widened", *datevaluetest.Between(1815, 0, 0, 1820, 0, 0), day(1815, 1, 1), day(1820, 12, 31)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			first, last, ok := valuecompare.AboutEnds(tc.d, 1)
			if !ok || first.Lo != tc.lo || last.Hi != tc.hi {
				t.Fatalf("got %d..%d ok=%v, want %d..%d", first.Lo, last.Hi, ok, tc.lo, tc.hi)
			}
			w := valuecompare.EndsWindow(first, last)
			if w.Lo == nil || w.Hi == nil || *w.Lo != tc.lo || *w.Hi != tc.hi {
				t.Fatalf("window %+v", w)
			}
		})
	}
	t.Run("two units", func(t *testing.T) {
		first, last, ok := valuecompare.AboutEnds(*datevaluetest.About(1817, 0, 0), 2)
		if !ok || first.Year != 1815 || last.Year != 1819 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("about a year keeps a day with no month", func(t *testing.T) {
		day := datevalue.DaysFromCivil
		first, last, ok := valuecompare.AboutEnds(*datevaluetest.About(1817, 0, 5), 1)
		if !ok || first.Year != 1816 || first.Month != 0 || first.Day != 5 || last.Year != 1818 || last.Day != 5 ||
			first.Lo != day(1816, 1, 1) || last.Hi != day(1818, 12, 31) {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("a phrase has no ends", func(t *testing.T) {
		if _, _, ok := valuecompare.AboutEnds(*datevaluetest.Phrase("spring"), 1); ok {
			t.Fatal("ok")
		}
	})
	t.Run("open ends", func(t *testing.T) {
		first, last, _ := valuecompare.AboutEnds(*bound(datevalue.QualifierBEF, 1820), 1)
		w := valuecompare.EndsWindow(first, last)
		if w.Lo != nil || w.Hi == nil || *w.Hi != datevalue.DaysFromCivil(1820, 12, 31) {
			t.Fatalf("BEF window %+v", w)
		}
		first, last, _ = valuecompare.AboutEnds(*bound(datevalue.QualifierAFT, 1820), 1)
		w = valuecompare.EndsWindow(first, last)
		if w.Hi != nil || w.Lo == nil || *w.Lo != datevalue.DaysFromCivil(1820, 1, 1) {
			t.Fatalf("AFT window %+v", w)
		}
	})
}

func TestDateWindow(t *testing.T) {
	year, _ := valuecompare.DateWindow(*datevaluetest.YMD(1985, 0, 0))
	plain, _ := valuecompare.DateWindow(*datevaluetest.YMD(1817, 0, 0))
	closed := func(w valuecompare.Window) bool { return w.Lo != nil && w.Hi != nil }
	for _, tc := range []struct {
		name  string
		d     datevalue.Value
		ok    bool
		check func(w valuecompare.Window) bool
	}{
		{"epoch", *datevaluetest.YMD(1970, 1, 1), true, func(w valuecompare.Window) bool {
			return closed(w) && *w.Lo == 0 && *w.Hi == 0
		}},
		{"year", *datevaluetest.YMD(1985, 0, 0), true, func(w valuecompare.Window) bool {
			return closed(w) && *w.Lo < *w.Hi
		}},
		{"day inside the year", *datevaluetest.YMD(1985, 5, 14), true, func(w valuecompare.Window) bool {
			return closed(w) && *w.Lo == *w.Hi && year.Contains(w)
		}},
		{"before", *bound(datevalue.QualifierBEF, 1900), true, func(w valuecompare.Window) bool {
			return w.Lo == nil && w.Hi != nil
		}},
		{"after", *bound(datevalue.QualifierAFT, 1872), true, func(w valuecompare.Window) bool {
			return w.Lo != nil && w.Hi == nil
		}},
		{"range with only an end", datevalue.Value{Kind: datevalue.KindRange, EndYear: datevaluetest.Part(1900)}, true, func(w valuecompare.Window) bool {
			return w.Lo == nil && w.Hi != nil
		}},
		{"range with only a start", datevalue.Value{Kind: datevalue.KindRange, StartYear: datevaluetest.Part(1885)}, true, func(w valuecompare.Window) bool {
			return w.Lo != nil && w.Hi == nil
		}},
		{"about is the plain year", *datevaluetest.About(1817, 0, 0), true, func(w valuecompare.Window) bool {
			return closed(w) && *w.Lo == *plain.Lo && *w.Hi == *plain.Hi
		}},
		{"phrase-only has no window", *datevaluetest.Phrase("Christmas"), false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, ok := valuecompare.DateWindow(tc.d)
			if ok != tc.ok {
				t.Fatalf("ok=%v want %v (%+v)", ok, tc.ok, w)
			}
			if ok && !tc.check(w) {
				t.Fatalf("got %+v", w)
			}
		})
	}
}

// yearZero is a point at year 0, or a range from year 0 to end when end is
// set; datevaluetest.Part reads 0 as unset, so the zero is built here.
func yearZero(end *int) *datevalue.Value {
	zero := 0
	v := datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &zero}
	if end != nil {
		v.Kind, v.EndYear = datevalue.KindRange, end
	}
	return &v
}

func endOf(t *testing.T, y, m, d int) valuecompare.DateEnd {
	t.Helper()
	e, _, ok := valuecompare.DateEnds(*datevaluetest.YMD(y, m, d))
	if !ok {
		t.Fatalf("no ends for %d-%d-%d", y, m, d)
	}
	return e
}

func TestShiftYears(t *testing.T) {
	t.Run("year precision", func(t *testing.T) {
		got, ok := valuecompare.ShiftYears(endOf(t, 1850, 0, 0), 110)
		if !ok || got.Year != 1960 || got.Lo != datevalue.DaysFromCivil(1960, 1, 1) {
			t.Errorf("got %+v ok=%v", got, ok)
		}
	})
	t.Run("day precision", func(t *testing.T) {
		got, ok := valuecompare.ShiftYears(endOf(t, 1901, 5, 1), 110)
		want := datevalue.DaysFromCivil(2011, 5, 1)
		if !ok || got.Lo != want || got.Hi != want {
			t.Errorf("got %+v ok=%v", got, ok)
		}
	})
	t.Run("cannot be placed", func(t *testing.T) {
		if _, ok := valuecompare.ShiftYears(valuecompare.DateEnd{Year: 1850, Month: 13}, 5); ok {
			t.Error("ok")
		}
	})
	t.Run("open end", func(t *testing.T) {
		open := valuecompare.DateEnd{Open: true}
		got, ok := valuecompare.ShiftYears(open, 5)
		if !ok || got != open {
			t.Errorf("got %+v ok=%v", got, ok)
		}
	})
}

func TestEarliestLatest(t *testing.T) {
	y1850, y1850d := endOf(t, 1850, 0, 0), endOf(t, 1850, 1, 1)
	y1900 := endOf(t, 1900, 0, 0)
	open := valuecompare.DateEnd{Open: true}
	t.Run("earliest lower Lo", func(t *testing.T) {
		got, ok := valuecompare.Earliest([]valuecompare.DateEnd{y1900, y1850})
		if !ok || got != y1850 {
			t.Errorf("got %+v ok=%v", got, ok)
		}
	})
	t.Run("earliest equal Lo lower Hi", func(t *testing.T) {
		got, ok := valuecompare.Earliest([]valuecompare.DateEnd{y1850, y1850d})
		if !ok || got != y1850d {
			t.Errorf("got %+v ok=%v", got, ok)
		}
	})
	t.Run("latest higher Hi", func(t *testing.T) {
		got, ok := valuecompare.Latest([]valuecompare.DateEnd{y1850, y1900})
		if !ok || got != y1900 {
			t.Errorf("got %+v ok=%v", got, ok)
		}
	})
	t.Run("latest equal Hi higher Lo", func(t *testing.T) {
		dec31 := endOf(t, 1850, 12, 31)
		got, ok := valuecompare.Latest([]valuecompare.DateEnd{y1850, dec31})
		if !ok || got != dec31 {
			t.Errorf("got %+v ok=%v", got, ok)
		}
	})
	t.Run("open ends skipped", func(t *testing.T) {
		got, ok := valuecompare.Earliest([]valuecompare.DateEnd{open, y1900})
		if !ok || got != y1900 {
			t.Errorf("earliest got %+v ok=%v", got, ok)
		}
		got, ok = valuecompare.Latest([]valuecompare.DateEnd{y1900, open})
		if !ok || got != y1900 {
			t.Errorf("latest got %+v ok=%v", got, ok)
		}
	})
	t.Run("none closed", func(t *testing.T) {
		if _, ok := valuecompare.Earliest([]valuecompare.DateEnd{open}); ok {
			t.Error("Earliest of open ends ok")
		}
		if _, ok := valuecompare.Latest(nil); ok {
			t.Error("Latest of nothing ok")
		}
	})
}
