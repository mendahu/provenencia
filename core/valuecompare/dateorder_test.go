package valuecompare_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestDateKeyOf(t *testing.T) {
	keyOf := func(t *testing.T, d *datevalue.Value) valuecompare.DateKey {
		t.Helper()
		k, _, _, ok := valuecompare.DateKeyOf(*d)
		if !ok {
			t.Fatalf("%+v has no place", *d)
		}
		return k
	}

	t.Run("ascending order", func(t *testing.T) {
		dates := []struct {
			name string
			d    *datevalue.Value
		}{
			{"1879", datevaluetest.YMD(1879, 0, 0)},
			{"bef. 1880", datevaluetest.Before(1880, 0, 0)},
			{"1880", datevaluetest.YMD(1880, 0, 0)},
			{"1881", datevaluetest.YMD(1881, 0, 0)},
			{"bef. 29 Mar 2011", datevaluetest.Before(2011, 3, 29)},
			{"29 Mar 2011", datevaluetest.YMD(2011, 3, 29)},
			{"aft. 29 Mar 2011", datevaluetest.After(2011, 3, 29)},
		}
		for i := 1; i < len(dates); i++ {
			a, b := keyOf(t, dates[i-1].d), keyOf(t, dates[i].d)
			if a.Compare(b) >= 0 || b.Compare(a) <= 0 || a.Int64() >= b.Int64() {
				t.Errorf("%s (%+v) not before %s (%+v)", dates[i-1].name, a, dates[i].name, b)
			}
		}
	})
	t.Run("equal keys", func(t *testing.T) {
		for _, c := range []struct {
			name string
			a, b *datevalue.Value
		}{
			{"about", datevaluetest.About(1880, 0, 0), datevaluetest.YMD(1880, 0, 0)},
			{"year is 1 Jan", datevaluetest.YMD(1880, 0, 0), datevaluetest.YMD(1880, 1, 1)},
			{"reversed range", datevaluetest.Between(1885, 0, 0, 1880, 0, 0), datevaluetest.Between(1880, 0, 0, 1885, 0, 0)},
			{"open start is BEF its end", &datevalue.Value{Kind: datevalue.KindRange, EndYear: datevaluetest.Part(1880)}, datevaluetest.Before(1880, 0, 0)},
		} {
			a, b := keyOf(t, c.a), keyOf(t, c.b)
			if a.Compare(b) != 0 || a.Int64() != b.Int64() {
				t.Errorf("%s: %+v != %+v", c.name, a, b)
			}
		}
	})
	t.Run("no place", func(t *testing.T) {
		for name, d := range map[string]*datevalue.Value{
			"phrase only":        datevaluetest.Phrase("in the reign of Victoria"),
			"both ends open":     {Kind: datevalue.KindRange},
			"year below window":  datevaluetest.YMD(datevalue.MinYear-1, 0, 0),
			"year above window":  datevaluetest.YMD(datevalue.MaxYear+1, 0, 0),
			"range end too late": datevaluetest.Between(1880, 0, 0, datevalue.MaxYear+1, 0, 0),
		} {
			if _, _, _, ok := valuecompare.DateKeyOf(*d); ok {
				t.Errorf("%s has a place", name)
			}
		}
	})
}
