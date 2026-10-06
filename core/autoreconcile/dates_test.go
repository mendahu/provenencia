package autoreconcile

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

func ymd(y int, m, d *int) datevalues.Value {
	return datevalues.Value{
		Kind: datevalues.KindPoint, Calendar: "gregorian",
		StartYear: ip(y), StartMonth: m, StartDay: d,
	}
}

func qualified(qual string, y int, m, d *int) datevalues.Value {
	v := ymd(y, m, d)
	v.Qualifier = qual
	return v
}

func between(y1 int, m1, d1 *int, y2 int, m2, d2 *int) datevalues.Value {
	return datevalues.Value{
		Kind: datevalues.KindRange, Calendar: "gregorian",
		StartYear: ip(y1), StartMonth: m1, StartDay: d1,
		EndYear: ip(y2), EndMonth: m2, EndDay: d2,
	}
}

func dateUnitOf(t *testing.T, d datevalues.Value) unit {
	t.Helper()
	units, ok := (dateModule{}).split(Value{Date: &d})
	if !ok {
		t.Fatalf("no window for %+v", d)
	}
	return units[wholeValue]
}

func TestDateFold(t *testing.T) {
	may := ip(5)
	apr := ip(4)
	jun := ip(6)
	aug := ip(8)
	mar := ip(3)
	day1 := ip(1)
	day2 := ip(2)
	day3 := ip(3)
	day14 := ip(14)
	hour := ip(15)

	year := ymd(1985, nil, nil)
	month := ymd(1985, may, nil)
	day := ymd(1985, may, day14)
	aprMonth := ymd(1985, apr, nil)
	junMonth := ymd(1985, jun, nil)
	day3May := ymd(1985, may, day3)
	day14Jun := ymd(1985, jun, day14)
	year84 := ymd(1984, nil, nil)
	year86 := ymd(1986, nil, nil)
	abtYear := qualified(datevalues.QualifierABT, 1985, nil, nil)
	abtMay := qualified(datevalues.QualifierABT, 1985, may, nil)
	bef1900 := qualified(datevalues.QualifierBEF, 1900, nil, nil)
	bef1890 := qualified(datevalues.QualifierBEF, 1890, nil, nil)
	befMay := qualified(datevalues.QualifierBEF, 1985, may, nil)
	aft1872 := qualified(datevalues.QualifierAFT, 1872, nil, nil)
	aft1880 := qualified(datevalues.QualifierAFT, 1880, nil, nil)
	aftMay := qualified(datevalues.QualifierAFT, 1985, may, nil)
	y1885 := ymd(1885, nil, nil)
	y1882 := ymd(1882, nil, nil)
	y1886 := ymd(1886, nil, nil)
	y1890 := ymd(1890, nil, nil)
	y1900 := ymd(1900, nil, nil)
	y1901 := ymd(1901, nil, nil)
	y1871 := ymd(1871, nil, nil)
	y1880 := ymd(1880, nil, nil)
	span := between(1880, nil, nil, 1885, nil, nil)
	overlapA := between(1880, ip(1), nil, 1880, jun, nil)
	overlapB := between(1880, mar, nil, 1880, ip(12), nil)
	mayDay := ymd(1880, may, day14)
	withHour := ymd(1985, may, day14)
	withHour.StartHour = hour
	leap := ymd(1984, ip(2), ip(29))
	feb1985 := ymd(1985, ip(2), nil)
	sameDayRange := between(1985, may, day14, 1985, may, day14)
	befDay := qualified(datevalues.QualifierBEF, 1985, may, day14)
	aftDay := qualified(datevalues.QualifierAFT, 1985, may, day14)
	otherCal := day
	otherCal.Calendar = "julian"
	maySpan := between(1985, mar, nil, 1985, jun, nil)
	june1985 := ymd(1985, jun, nil)
	april1985 := ymd(1985, apr, nil)

	// fold(from, into): from is wider and contains into.
	cases := []struct {
		name     string
		from, to datevalues.Value
		fold     bool
	}{
		{"year folds into its month", year, month, true},
		{"month does not fold into its year", month, year, false},
		{"month folds into its day", month, day, true},
		{"day does not fold into its month", day, month, false},
		{"year folds into its day", year, day, true},
		{"day does not fold into its year", day, year, false},
		{"April does not fold into May", aprMonth, month, false},
		{"May does not fold into April", month, aprMonth, false},
		{"3 May does not fold into 14 June", day3May, day14Jun, false},
		{"14 June does not fold into 3 May", day14Jun, day3May, false},
		{"about 1985 folds into 1985", abtYear, year, true},
		{"1985 does not fold into about 1985", year, abtYear, false},
		{"about 1985 folds into May 1985", abtYear, month, true},
		{"about 1985 does not fold into 1986", abtYear, year86, false},
		{"before 1900 folds into 1885", bef1900, y1885, true},
		{"before 1900 folds into 1900", bef1900, y1900, true},
		{"before 1900 does not fold into 1901", bef1900, y1901, false},
		{"1885 does not fold into before 1900", y1885, bef1900, false},
		{"after 1872 folds into 1880", aft1872, y1880, true},
		{"after 1872 does not fold into 1871", aft1872, y1871, false},
		{"before 1900 folds into before 1890", bef1900, bef1890, true},
		{"before 1890 does not fold into before 1900", bef1890, bef1900, false},
		{"after 1872 folds into after 1880", aft1872, aft1880, true},
		{"after 1880 does not fold into after 1872", aft1880, aft1872, false},
		{"1880–1885 folds into 1882", span, y1882, true},
		{"1880–1885 does not fold into 1886", span, y1886, false},
		{"overlapping ranges do not fold", overlapA, overlapB, false},
		{"the other overlap does not fold either", overlapB, overlapA, false},
		{"a day folds into its hour", day, withHour, true},
		{"an hour does not fold into its day", withHour, day, false},
		{"equal years do not fold", year, year, false},
		{"3 May does not fold into 14 May", day3May, day, false},
		{"14 May does not fold into 3 May", day, day3May, false},
		{"May folds into 3 May", month, day3May, true},
		{"May folds into 14 May", month, day, true},
		{"before 1900 contains 1885 and 1890", bef1900, y1890, true},
		{"1885 does not fold into 1890", y1885, y1890, false},
		{"1984 does not fold into 1986", year84, year86, false},
		{"about May folds into 14 May", abtMay, day, true},
		{"about May does not fold into June", abtMay, junMonth, false},
		{"March–June folds into 14 May", maySpan, mayDay, false}, // mayDay is 1880; span is 1985
		{"March–June 1985 folds into 14 May 1985", between(1985, mar, nil, 1985, jun, nil), day, true},
		{"March–June 1985 does not fold into August", between(1985, mar, nil, 1985, jun, nil), ymd(1985, aug, day14), false},
		{"a one-day range folds into that day", sameDayRange, day, true},
		{"that day does not fold into its one-day range", day, sameDayRange, false},
		{"before a day folds into that day", befDay, day, true},
		{"after a day folds into that day", aftDay, day, true},
		{"a different calendar does not fold", otherCal, day, false},
		{"before May contains 1 May", befMay, ymd(1985, may, day1), true},
		{"before May contains April", befMay, april1985, true},
		{"before May does not contain June", befMay, june1985, false},
		{"after May contains June", aftMay, june1985, true},
		{"after May does not contain April", aftMay, april1985, false},
		{"February 1985 does not contain a later year", feb1985, year86, false},
		{"a leap day does not fold into the next year", leap, ymd(1985, ip(2), day1), false},
		{"1984 contains its leap day", ymd(1984, nil, nil), leap, true},
		{"2 May does not fold into 1 May", ymd(1985, may, day2), ymd(1985, may, day1), false},
	}
	if len(cases) < 50 {
		t.Fatalf("date cases = %d, want at least 50", len(cases))
	}
	m := dateModule{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.fold("", dateUnitOf(t, tc.from), dateUnitOf(t, tc.to))
			if got != tc.fold {
				t.Fatalf("fold = %v, want %v", got, tc.fold)
			}
		})
	}
}

func TestDateBounds(t *testing.T) {
	lo, hi, ok := DateBounds(ymd(1970, ip(1), ip(1)))
	if !ok || lo == nil || hi == nil || *lo != 0 || *hi != 0 {
		t.Fatalf("epoch %+v %+v ok=%v", lo, hi, ok)
	}
	lo, hi, ok = DateBounds(ymd(1985, nil, nil))
	if !ok || lo == nil || hi == nil || *lo >= *hi {
		t.Fatalf("year bounds %+v %+v", lo, hi)
	}
	dayLo, dayHi, ok := DateBounds(ymd(1985, ip(5), ip(14)))
	if !ok || dayLo == nil || dayHi == nil || *dayLo != *dayHi || *dayLo <= *lo || *dayHi >= *hi {
		t.Fatalf("day %+v %+v outside year %+v %+v", dayLo, dayHi, lo, hi)
	}
	lo, hi, ok = DateBounds(qualified(datevalues.QualifierBEF, 1900, nil, nil))
	if !ok || lo != nil || hi == nil {
		t.Fatalf("before %+v %+v ok=%v", lo, hi, ok)
	}
	lo, hi, ok = DateBounds(qualified(datevalues.QualifierAFT, 1872, nil, nil))
	if !ok || lo == nil || hi != nil {
		t.Fatalf("after %+v %+v ok=%v", lo, hi, ok)
	}
	phrase := datevalues.Value{Kind: datevalues.KindPoint, Phrase: "Christmas"}
	if _, _, ok := DateBounds(phrase); ok {
		t.Fatal("phrase-only has a window")
	}
	if _, ok := (dateModule{}).split(Value{Date: &phrase}); ok {
		t.Fatal("phrase-only is evidence")
	}
}

func TestDateReconcile(t *testing.T) {
	may, day14 := ip(5), ip(14)
	cases := []struct {
		name   string
		cands  []Candidate
		want   [][]byte
		state  State
		shown  *datevalues.Value // rank-1 displayed date, when one value
		folded []byte            // observation ids whose reason is folded
		noEvid []byte
	}{
		{
			name: "equal days merge",
			cands: []Candidate{
				date(1, 1985, may, day14),
				date(2, 1985, may, day14),
			},
			want: [][]byte{{1, 2}}, state: StateMerged,
			shown: datePtr(ymd(1985, may, day14)),
		},
		{
			name: "May folds into 14 May",
			cands: []Candidate{
				date(1, 1985, may, nil),
				date(2, 1985, may, day14),
			},
			want: [][]byte{{1, 2}}, state: StateMerged,
			shown: datePtr(ymd(1985, may, day14)), folded: []byte{1},
		},
		{
			name: "April and May stay mixed",
			cands: []Candidate{
				date(1, 1985, ip(4), nil),
				date(2, 1985, may, nil),
			},
			want: [][]byte{{1}, {2}}, state: StateMixed,
		},
		{
			name: "3 May and 14 June stay mixed",
			cands: []Candidate{
				date(1, 1985, may, ip(3)),
				date(2, 1985, ip(6), day14),
			},
			want: [][]byte{{1}, {2}}, state: StateMixed,
		},
		{
			name: "before 1900 does not glue 1885 and 1890",
			cands: []Candidate{
				{ObservationID: id(1), Value: Value{Date: datePtr(qualified(datevalues.QualifierBEF, 1900, nil, nil))}},
				date(2, 1885, nil, nil),
				date(3, 1890, nil, nil),
			},
			want: [][]byte{{1, 2}, {3}}, state: StateMixed, folded: []byte{1},
		},
		{
			name: "a phrase is no evidence",
			cands: []Candidate{
				{ObservationID: id(1), Value: Value{Date: &datevalues.Value{Kind: datevalues.KindPoint, Phrase: "Christmas"}}},
				date(2, 1985, may, day14),
			},
			want: [][]byte{{2}}, state: StateSingle, noEvid: []byte{1},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Reconcile(properties.ValueTypeDate, tc.cands, nil, properties.CardinalitySingle)
			if err != nil {
				t.Fatal(err)
			}
			if s := shape(got); !equalIDs(s, tc.want) {
				t.Fatalf("values %v, want %v", s, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
			if tc.shown != nil {
				gotDate := got.Values[0].Value.Date
				if gotDate == nil || gotDate.StartYear == nil || *gotDate.StartYear != *tc.shown.StartYear {
					t.Fatalf("displayed %+v, want %+v", gotDate, tc.shown)
				}
				if (gotDate.StartDay == nil) != (tc.shown.StartDay == nil) ||
					(gotDate.StartDay != nil && *gotDate.StartDay != *tc.shown.StartDay) {
					t.Fatalf("displayed day %+v, want %+v", gotDate.StartDay, tc.shown.StartDay)
				}
			}
			for _, n := range tc.folded {
				if reasonOf(got, n) != ReasonFolded {
					t.Fatalf("obs %d reason %q, want folded", n, reasonOf(got, n))
				}
			}
			for _, n := range tc.noEvid {
				if reasonOf(got, n) != ReasonNoEvidence {
					t.Fatalf("obs %d reason %q, want no_evidence", n, reasonOf(got, n))
				}
			}
		})
	}
}

func datePtr(d datevalues.Value) *datevalues.Value { return &d }

func equalIDs(got, want [][]byte) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if len(got[i]) != len(want[i]) {
			return false
		}
		for j := range got[i] {
			if got[i][j] != want[i][j] {
				return false
			}
		}
	}
	return true
}

func reasonOf(r Result, n byte) Reason {
	for _, o := range r.Outcomes {
		if len(o.ObservationID) > 0 && o.ObservationID[len(o.ObservationID)-1] == n {
			return o.Reason
		}
	}
	return ""
}
