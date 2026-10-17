package conclusionheaders

import (
	"testing"

	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func ip(v int) *int { return &v }

func ymd(y int, m, d *int) datevalue.Value {
	return datevalue.Value{
		Kind: datevalue.KindPoint, Calendar: "gregorian",
		StartYear: ip(y), StartMonth: m, StartDay: d,
	}
}

func qualified(qual string, y int, m, d *int) datevalue.Value {
	v := ymd(y, m, d)
	v.Qualifier = qual
	return v
}

func between(y1 int, m1, d1 *int, y2 int, m2, d2 *int) datevalue.Value {
	return datevalue.Value{
		Kind: datevalue.KindRange, Calendar: "gregorian",
		StartYear: ip(y1), StartMonth: m1, StartDay: d1,
		EndYear: ip(y2), EndMonth: m2, EndDay: d2,
	}
}

func ptrDate(d datevalue.Value) *datevalue.Value { return &d }

func mustWindow(t *testing.T, d datevalue.Value) valuecompare.Window {
	t.Helper()
	w, ok := valuecompare.DateWindow(d)
	if !ok {
		t.Fatalf("no window for %+v", d)
	}
	return w
}

func TestPeriodWindowAndLinkMembership(t *testing.T) {
	uc := periodWindow(ptrDate(ymd(1791, nil, nil)), ptrDate(ymd(1841, nil, nil)))
	pc := periodWindow(ptrDate(ymd(1841, nil, nil)), ptrDate(ymd(1867, nil, nil)))
	ontario := periodWindow(ptrDate(ymd(1867, nil, nil)), nil)
	always := periodWindow(nil, nil)
	if !always.Overlaps(uc) || !valuecompare.Always().Overlaps(always) {
		t.Fatal("undated place always holds")
	}
	overlap, ok := always.Intersect(uc)
	if !ok || overlap.Lo == nil || *overlap.Lo != *uc.Lo || overlap.Hi == nil || *overlap.Hi != *uc.Hi {
		t.Fatalf("always ∩ uc = %+v ok=%v", overlap, ok)
	}
	// Undated Toronto→UC holds on UC's period when Toronto has no period.
	mem, ok := linkMembership(nil, nil, always, uc)
	if !ok || !mustWindow(t, ymd(1820, nil, nil)).Overlaps(mem) {
		t.Fatalf("1820 in undated UC membership %+v ok=%v", mem, ok)
	}
	if mustWindow(t, ymd(1850, nil, nil)).Overlaps(mem) {
		t.Fatal("1850 should miss UC membership")
	}
	// Dated Ireland⊂UK until 1922: both places continue.
	ireland := always
	uk := always
	end1922 := ymd(1922, nil, nil)
	dated, ok := linkMembership(nil, &end1922, ireland, uk)
	if !ok {
		t.Fatal("dated membership")
	}
	if !mustWindow(t, ymd(1900, nil, nil)).Overlaps(dated) {
		t.Fatal("1900 should hold in Ireland⊂UK until 1922")
	}
	if mustWindow(t, ymd(1923, nil, nil)).Overlaps(dated) {
		t.Fatal("1923 should miss Ireland⊂UK until 1922")
	}
	// About 1841 overlaps both UC and PC undated memberships.
	abt := mustWindow(t, qualified(datevalue.QualifierABT, 1841, nil, nil))
	ucMem, _ := linkMembership(nil, nil, always, uc)
	pcMem, _ := linkMembership(nil, nil, always, pc)
	if !abt.Overlaps(ucMem) || !abt.Overlaps(pcMem) {
		t.Fatalf("about 1841 should hold in both UC and PC")
	}
	// Range straddling UC→PC is ambiguous on each.
	span := mustWindow(t, between(1840, nil, nil, 1842, nil, nil))
	if !(span.Overlaps(ucMem) && !ucMem.Contains(span)) {
		t.Fatal("1840–1842 vs UC should overlap without being inside")
	}
	if !(span.Overlaps(pcMem) && !pcMem.Contains(span)) {
		t.Fatal("1840–1842 vs PC should overlap without being inside")
	}
	ontMem, _ := linkMembership(nil, nil, always, ontario)
	if !mustWindow(t, ymd(1950, nil, nil)).Overlaps(ontMem) {
		t.Fatal("1950 should hold in Ontario")
	}
}

func TestMembershipDatesByDay(t *testing.T) {
	day := func(y, m, d int) *datevalue.Value { return ptrDate(ymd(y, ip(m), ip(d))) }
	march, sept := day(1850, 3, 1), day(1850, 9, 1)
	oct31, jul1 := day(1867, 10, 31), day(1867, 7, 1)
	for _, c := range []struct {
		name     string
		from, to PlaceHeader
	}{
		{"part first", PlaceHeader{StartDate: march, EndDate: oct31}, PlaceHeader{StartDate: sept, EndDate: jul1}},
		{"whole first", PlaceHeader{StartDate: sept, EndDate: jul1}, PlaceHeader{StartDate: march, EndDate: oct31}},
	} {
		t.Run(c.name, func(t *testing.T) {
			start, end := membershipDates(placeLink{}, c.from, c.to)
			if start != sept || end != jul1 {
				t.Fatalf("got %+v %+v want %+v %+v", start, end, sept, jul1)
			}
		})
	}
}
