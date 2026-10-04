package match

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
)

func span(lo, hi *int) Value {
	return Value{Date: &datevalues.Value{Kind: datevalues.KindRange, StartYear: lo, EndYear: hi}}
}

func bound(q string, y int) Value {
	return Value{Date: &datevalues.Value{Kind: datevalues.KindPoint, Qualifier: q, StartYear: &y}}
}

func date(y int, m, d *int) Value {
	return Value{Date: &datevalues.Value{StartYear: &y, StartMonth: m, StartDay: d}}
}

func TestDateComparerScores(t *testing.T) {
	tests := []struct {
		name       string
		cmp        DateComparer
		a, b       Value
		want       float64
		comparable bool
	}{
		{"same year, both year-only", DateComparer{}, date(1817, nil, nil), date(1817, nil, nil), 1, true},
		{"same month, both without day", DateComparer{}, date(1817, ip(5), nil), date(1817, ip(5), nil), 1, true},
		{"same day", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), ip(14)), 1, true},
		{"same month", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), nil), 0.9, true},
		{"same year, other month", DateComparer{}, date(1817, ip(5), nil), date(1817, ip(9), nil), 0.6, true},
		{"same year", DateComparer{}, date(1817, nil, nil), date(1817, ip(9), nil), 0.8, true},
		{"a year apart", DateComparer{}, date(1817, nil, nil), date(1818, nil, nil), 0.8 * 2 / 3, true},
		{"beyond tolerance", DateComparer{}, date(1817, nil, nil), date(1820, nil, nil), 0, true},
		{"wider tolerance", DateComparer{Tolerance: Set(5)}, date(1817, nil, nil), date(1820, nil, nil), 0.8 * 3 / 6, true},
		{"about widens", DateComparer{}, Value{Date: &datevalues.Value{StartYear: ip(1817), Qualifier: "ABT"}}, date(1820, nil, nil), 0.8 * 2 / 5, true},
		{"same month, different days", DateComparer{}, date(1817, ip(5), ip(14)), date(1817, ip(5), ip(20)), 0.7, true},
		{"range contains the point", DateComparer{}, span(ip(1815), ip(1820)), date(1818, nil, nil), 0.8 - 0.05*5, true},
		{"tight range contains the point", DateComparer{}, span(ip(1817), ip(1818)), date(1817, nil, nil), 0.8 - 0.05, true},
		{"wide range contains the point", DateComparer{}, span(ip(1700), ip(1900)), date(1817, nil, nil), 0.2, true},
		{"point just outside a range", DateComparer{}, span(ip(1815), ip(1820)), date(1822, nil, nil), (0.8 - 0.05*5) * (1 - 2.0/3), true},
		{"point far outside a range", DateComparer{}, span(ip(1815), ip(1820)), date(1830, nil, nil), 0, true},
		{"identical ranges", DateComparer{}, span(ip(1815), ip(1820)), span(ip(1815), ip(1820)), 1, true},
		{"overlapping ranges, judged by the wider", DateComparer{}, span(ip(1815), ip(1820)), span(ip(1819), ip(1825)), 0.8 - 0.05*6, true},
		{"open-ended range", DateComparer{}, span(ip(1815), nil), date(1900, nil, nil), 0.2, true},
		{"identical bounds", DateComparer{}, bound("BEF", 1820), bound("BEF", 1820), 1, true},
		{"before, consistent", DateComparer{}, bound("BEF", 1820), date(1815, nil, nil), 0.2, true},
		{"before, impossible", DateComparer{}, bound("BEF", 1820), date(1823, nil, nil), 0, true},
		{"after, a year short", DateComparer{}, bound("AFT", 1820), date(1819, nil, nil), 0.2 * (1 - 1.0/3), true},
		{"before vs after, disjoint", DateComparer{}, bound("BEF", 1800), bound("AFT", 1810), 0, true},
		{"before vs before, different years", DateComparer{}, bound("BEF", 1800), bound("BEF", 1810), 0.2, true},
		{"tolerance zero", DateComparer{Tolerance: Set(0)}, date(1817, nil, nil), date(1818, nil, nil), 0, true},
		{"inverted range reads as its years", DateComparer{}, span(ip(1820), ip(1815)), span(ip(1815), ip(1820)), 1, true},
		{"empty range", DateComparer{}, span(nil, nil), date(1818, nil, nil), 0, false},
		{"phrase only", DateComparer{}, Value{Date: &datevalues.Value{Phrase: "spring"}}, date(1817, nil, nil), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.cmp.Compare(tt.a, tt.b)
			if ok != tt.comparable || !approx(tt.want)(got) {
				t.Fatalf("got %v %v, want %v %v", got, ok, tt.want, tt.comparable)
			}
			if back, _ := tt.cmp.Compare(tt.b, tt.a); !approx(got)(back) {
				t.Fatalf("not symmetric: %v then %v", got, back)
			}
		})
	}
}

// Date ladder: each date must resemble 14 May 1817 strictly more than the next.
func TestDateComparerLadder(t *testing.T) {
	probe := date(1817, ip(5), ip(14))
	order := []struct {
		name string
		v    Value
	}{
		{"same day", date(1817, ip(5), ip(14))},
		{"same month, no day", date(1817, ip(5), nil)},
		{"same year, no month", date(1817, nil, nil)},
		{"same month, other day", date(1817, ip(5), ip(20))},
		{"same year, other month", date(1817, ip(9), nil)},
		{"inside a six-year range", span(ip(1815), ip(1820))},
		{"a year later", date(1818, nil, nil)},
		{"two years later", date(1819, nil, nil)},
		{"before a later year", bound("BEF", 1820)},
		{"three years later", date(1820, nil, nil)},
	}
	prev := math.Inf(1)
	for i, o := range order {
		got, ok := DateComparer{}.Compare(probe, o.v)
		if !ok {
			t.Fatalf("%s: not comparable", o.name)
		}
		if !(got < prev) && !(i == len(order)-1 && got == 0) {
			t.Errorf("#%d %s scored %.4f, not below %.4f", i, o.name, got, prev)
		}
		prev = got
	}
}

// Generated dates from fixed seeds: symmetry, range, and self-similarity.
func TestDateComparerInvariants(t *testing.T) {
	quals := []string{"", "", "ABT", "BEF", "AFT"}
	opt := func(rng *rand.Rand, lo, n int) *int {
		if rng.Intn(3) == 0 {
			return nil
		}
		return ip(lo + rng.Intn(n))
	}
	for _, seed := range []int64{1, 2, 3, 4} {
		rng := rand.New(rand.NewSource(seed))
		gen := func() Value {
			if rng.Intn(3) == 0 {
				return span(opt(rng, 1810, 10), opt(rng, 1815, 10))
			}
			d := &datevalues.Value{Kind: datevalues.KindPoint, Qualifier: quals[rng.Intn(len(quals))], StartYear: opt(rng, 1810, 15)}
			if d.StartYear != nil {
				d.StartMonth = opt(rng, 1, 12)
				if d.StartMonth != nil {
					d.StartDay = opt(rng, 1, 28)
				}
			}
			return Value{Date: d}
		}
		for _, cmp := range []DateComparer{{}, {Tolerance: Set(0)}, {Tolerance: Set(5)}} {
			for i := 0; i < 2000; i++ {
				a, b := gen(), gen()
				got, ok := cmp.Compare(a, b)
				back, okBack := cmp.Compare(b, a)
				where := fmt.Sprintf("seed %d pair %d: %+v vs %+v", seed, i, *a.Date, *b.Date)
				if ok != okBack || math.Abs(got-back) > 1e-9 {
					t.Fatalf("%s: not symmetric: %.4f then %.4f", where, got, back)
				}
				if ok && (got < 0 || got > 1) {
					t.Fatalf("%s: out of range %.4f", where, got)
				}
				if self, okSelf := cmp.Compare(a, a); okSelf && self != 1 {
					t.Fatalf("%s: self-similarity %.4f", where, self)
				}
			}
		}
	}
}
