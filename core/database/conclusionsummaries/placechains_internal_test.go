package conclusionsummaries

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database/graphcache"
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
		from, to PlaceSummary
	}{
		{"part first", PlaceSummary{StartDate: march, EndDate: oct31}, PlaceSummary{StartDate: sept, EndDate: jul1}},
		{"whole first", PlaceSummary{StartDate: sept, EndDate: jul1}, PlaceSummary{StartDate: march, EndDate: oct31}},
	} {
		t.Run(c.name, func(t *testing.T) {
			start, end := membershipDates(placeLink{}, c.from, c.to)
			if start != sept || end != jul1 {
				t.Fatalf("got %+v %+v want %+v %+v", start, end, sept, jul1)
			}
		})
	}
}

func TestMembershipDatesByDayBranches(t *testing.T) {
	sept := ptrDate(ymd(1850, ip(9), ip(1)))
	spring := ptrDate(datevalue.Value{Kind: datevalue.KindPoint, Phrase: "spring"})
	before1850 := ptrDate(qualified(datevalue.QualifierBEF, 1850, nil, nil))
	y1840 := ptrDate(ymd(1840, nil, nil))
	after1867 := ptrDate(qualified(datevalue.QualifierAFT, 1867, nil, nil))
	jul1 := ptrDate(ymd(1867, ip(7), ip(1)))
	for _, c := range []struct {
		name       string
		a, b, want *datevalue.Value
		start      bool
	}{
		{"a phrase-only start never beats a readable one", spring, sept, sept, true},
		{"an open start loses to a set one", before1850, y1840, y1840, true},
		{"an open end loses to a set one", after1867, jul1, jul1, false},
	} {
		for _, swap := range []bool{false, true} {
			a, b := c.a, c.b
			if swap {
				a, b = b, a
			}
			t.Run(c.name, func(t *testing.T) {
				start, end := membershipDates(placeLink{}, PlaceSummary{StartDate: a, EndDate: a}, PlaceSummary{StartDate: b, EndDate: b})
				got := end
				if c.start {
					got = start
				}
				if got != c.want {
					t.Fatalf("swap=%v got %+v want %+v", swap, got, c.want)
				}
			})
		}
	}
	t.Run("a tie keeps the first", func(t *testing.T) {
		a, b := ptrDate(ymd(1850, nil, nil)), ptrDate(ymd(1850, nil, nil))
		start, end := membershipDates(placeLink{}, PlaceSummary{StartDate: a, EndDate: a}, PlaceSummary{StartDate: b, EndDate: b})
		if start != a || end != a {
			t.Fatalf("got %p %p want %p", start, end, a)
		}
	})
}

func TestRolesByNeighbor(t *testing.T) {
	id := func(n byte) []byte { return []byte{n} }
	link := func(neighbor, term byte) graphcache.Link {
		l := graphcache.Link{Neighbor: id(neighbor)}
		if term != 0 {
			l.TermID, l.TermKey = id(term), "t"
		}
		return l
	}
	label := func(termID []byte, key string) (Term, bool) {
		if bytes.Equal(termID, id(9)) {
			return Term{}, false
		}
		return Term{ID: termID, Key: key}, true
	}
	// A subject (1), A informant (2), A subject again (1), B no term, C a term
	// the labeller does not know (9), D a neighbor like any other.
	links := []graphcache.Link{link('A', 1), link('A', 2), link('A', 1), link('B', 0), link('C', 9), link('D', 1)}
	got := rolesByNeighbor(links, label)
	if len(got) != 4 {
		t.Fatalf("%d neighbors, want 4: %+v", len(got), got)
	}
	if !bytes.Equal(got[0].id, id('A')) || len(got[0].roles) != 2 ||
		!bytes.Equal(got[0].roles[0].ID, id(1)) || !bytes.Equal(got[0].roles[1].ID, id(2)) {
		t.Fatalf("A: %+v", got[0])
	}
	if !bytes.Equal(got[1].id, id('B')) || len(got[1].roles) != 0 {
		t.Fatalf("B: %+v", got[1])
	}
	if !bytes.Equal(got[2].id, id('C')) || len(got[2].roles) != 0 {
		t.Fatalf("C: %+v", got[2])
	}
	if !bytes.Equal(got[3].id, id('D')) || len(got[3].roles) != 1 || !bytes.Equal(got[3].roles[0].ID, id(1)) {
		t.Fatalf("D: %+v", got[3])
	}
}

// TestRolesByNeighborManyLinks pins the contract for one neighbor reached by
// many links: each distinct role once, in first-seen order.
func TestRolesByNeighborManyLinks(t *testing.T) {
	id := func(n byte) []byte { return []byte{n} }
	label := func(termID []byte, key string) (Term, bool) {
		return Term{ID: termID, Key: key}, true
	}
	var links []graphcache.Link
	for i := 0; i < 50; i++ {
		for _, term := range []byte{3, 1, 2, 3, 1} {
			links = append(links, graphcache.Link{Neighbor: id('A'), TermID: id(term), TermKey: "t"})
		}
	}
	got := rolesByNeighbor(links, label)
	if len(got) != 1 || len(got[0].roles) != 3 {
		t.Fatalf("got %+v", got)
	}
	for i, want := range []byte{3, 1, 2} {
		if !bytes.Equal(got[0].roles[i].ID, id(want)) {
			t.Fatalf("role %d: %+v, want %d", i, got[0].roles[i], want)
		}
	}
}
