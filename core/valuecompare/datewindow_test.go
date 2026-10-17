package valuecompare_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestDateEnds(t *testing.T) {
	const eps = 1e-12
	near := func(a, b float64) bool { return a-b < eps && b-a < eps }

	t.Run("year", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(ymd(1817, 0, 0))
		lo, hi, _ := datevalue.BoundDays(optInt(1817), nil, nil)
		if !ok || first != last || first.Open || first.Year != 1817 || first.Month != 0 || first.Day != 0 ||
			first.From != 1817 || first.To != 1818 || first.Lo != lo || first.Hi != hi ||
			!near(first.Mid, (first.From+first.To)/2) {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("month", func(t *testing.T) {
		first, _, ok := valuecompare.DateEnds(ymd(1817, 5, 0))
		// May is days 121 to 151 of 365.
		if !ok || first.Month != 5 || first.Day != 0 || !near(first.From, 1817+120.0/365) || !near(first.To, 1817+151.0/365) {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	t.Run("leap month", func(t *testing.T) {
		first, _, ok := valuecompare.DateEnds(ymd(1816, 2, 0))
		// February 1816 is days 32 to 60 of 366.
		if !ok || !near(first.From, 1816+31.0/366) || !near(first.To, 1816+60.0/366) {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	t.Run("day", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(ymd(1817, 5, 14))
		if !ok || first != last || first.Month != 5 || first.Day != 14 || first.Lo != first.Hi ||
			!near(first.From, 1817+133.0/365) || !near(first.To, 1817+134.0/365) {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	t.Run("range keeps its months", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(between(1817, 3, 0, 1817, 6, 0))
		if !ok || first.Month != 3 || last.Month != 6 || first.Open || last.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("before", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(bound(datevalue.QualifierBEF, 1820))
		if !ok || !first.Open || first.Year != 0 || last.Open || last.Year != 1820 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("after", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(bound(datevalue.QualifierAFT, 1820))
		if !ok || first.Open || !last.Open || last.Year != 0 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("range with no end", func(t *testing.T) {
		d := datevalue.Value{Kind: datevalue.KindRange, StartYear: optInt(1815)}
		first, last, ok := valuecompare.DateEnds(d)
		if !ok || first.Open || !last.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("range with no start", func(t *testing.T) {
		d := datevalue.Value{Kind: datevalue.KindRange, EndYear: optInt(1815)}
		first, last, ok := valuecompare.DateEnds(d)
		if !ok || !first.Open || last.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("inverted range swaps", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(between(1820, 0, 0, 1815, 0, 0))
		if !ok || first.Year != 1815 || last.Year != 1820 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("a month inside the end year does not swap", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(between(1817, 3, 0, 1817, 0, 0))
		if !ok || first.Month != 3 || last.Month != 0 || first.Lo >= last.Hi {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
		first, last, ok = valuecompare.DateEnds(between(1817, 5, 14, 1817, 0, 0))
		if !ok || first.Month != 5 || first.Day != 14 || last.Month != 0 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("about is a point", func(t *testing.T) {
		first, last, ok := valuecompare.DateEnds(about(1817, 0, 0))
		if !ok || first != last || first.Open {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("empty kind is a point", func(t *testing.T) {
		d := datevalue.Value{StartYear: optInt(1817)}
		first, last, ok := valuecompare.DateEnds(d)
		if !ok || first != last {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("point with only end fields", func(t *testing.T) {
		d := datevalue.Value{Kind: datevalue.KindPoint, EndYear: optInt(1817)}
		first, _, ok := valuecompare.DateEnds(d)
		if !ok || first.Year != 1817 {
			t.Fatalf("got %+v ok=%v", first, ok)
		}
	})
	for name, d := range map[string]datevalue.Value{
		"month 13":        ymd(1817, 13, 0),
		"day 31 in April": ymd(1817, 4, 31),
		"phrase only":     phrase("spring"),
		"range, no year":  {Kind: datevalue.KindRange},
		"range, bad end":  between(1815, 0, 0, 1817, 13, 0),
	} {
		t.Run("not ok: "+name, func(t *testing.T) {
			if _, _, ok := valuecompare.DateEnds(d); ok {
				t.Fatal("ok")
			}
		})
	}
}

func win(lo, hi *int) valuecompare.Window { return valuecompare.Window{Lo: lo, Hi: hi} }

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
		{"about a year", about(1817, 0, 0), day(1816, 1, 1), day(1818, 12, 31)},
		{"about a month", about(1817, 3, 0), day(1817, 2, 1), day(1817, 4, 30)},
		{"about January", about(1817, 1, 0), day(1816, 12, 1), day(1817, 2, 28)},
		{"about a day", about(1817, 1, 1), day(1816, 12, 31), day(1817, 1, 2)},
		{"a plain year", ymd(1817, 0, 0), day(1817, 1, 1), day(1817, 12, 31)},
		{"a range is not widened", between(1815, 0, 0, 1820, 0, 0), day(1815, 1, 1), day(1820, 12, 31)},
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
		first, last, ok := valuecompare.AboutEnds(about(1817, 0, 0), 2)
		if !ok || first.Year != 1815 || last.Year != 1819 {
			t.Fatalf("got %+v %+v ok=%v", first, last, ok)
		}
	})
	t.Run("a phrase has no ends", func(t *testing.T) {
		if _, _, ok := valuecompare.AboutEnds(phrase("spring"), 1); ok {
			t.Fatal("ok")
		}
	})
	t.Run("open ends", func(t *testing.T) {
		first, last, _ := valuecompare.AboutEnds(bound(datevalue.QualifierBEF, 1820), 1)
		w := valuecompare.EndsWindow(first, last)
		if w.Lo != nil || w.Hi == nil || *w.Hi != datevalue.DaysFromCivil(1820, 12, 31) {
			t.Fatalf("BEF window %+v", w)
		}
		first, last, _ = valuecompare.AboutEnds(bound(datevalue.QualifierAFT, 1820), 1)
		w = valuecompare.EndsWindow(first, last)
		if w.Hi != nil || w.Lo == nil || *w.Lo != datevalue.DaysFromCivil(1820, 1, 1) {
			t.Fatalf("AFT window %+v", w)
		}
	})
}
