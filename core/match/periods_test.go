package match

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
)

var (
	pDate  = product("date")
	pStart = product("start_date")
	pEnd   = product("end_date")
)

func year(y int) Value {
	return Value{Date: &datevalues.Value{Kind: datevalues.KindPoint, StartYear: &y}}
}

func denied(v Value) Value {
	v = kept(v, 1)
	v.Denied = true
	return v
}

func TestPoolPeriods(t *testing.T) {
	label := func(vs []Value) string {
		if len(vs) == 0 {
			return ""
		}
		return vs[len(vs)-1].Label()
	}
	tests := []struct {
		name          string
		probe, cand   Values
		prop          Property
		wantProbe     int // values on the probe side after pooling
		wantCand      int
		wantPeriod    string // label of the appended period; "" for none
		periodOnProbe bool
	}{
		{name: "the candidate has a period", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {year(1815)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Between 1815 and 1820"},
		{name: "the probe has a period", probe: Values{pStart: {year(1815)}, pEnd: {year(1820)}}, cand: Values{pDate: {year(1817)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Between 1815 and 1820", periodOnProbe: true},
		{name: "both have a date", probe: Values{pDate: {year(1817)}}, cand: Values{pDate: {year(1818)}, pStart: {year(1815)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1},
		{name: "neither has a date", probe: Values{pStart: {year(1815)}}, cand: Values{pStart: {year(1815)}, pEnd: {year(1820)}},
			prop: pDate},
		{name: "the dated side has a voting start", probe: Values{pDate: {year(1817)}, pStart: {year(1816)}}, cand: Values{pStart: {year(1815)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1},
		{name: "the dated side has a voting end", probe: Values{pDate: {year(1817)}, pEnd: {year(1816)}}, cand: Values{pStart: {year(1815)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1},
		{name: "a provisional start on the dated side does not block", probe: Values{pDate: {year(1817)}, pStart: {cached(year(1816), 1, 0, 1, "provisional")}}, cand: Values{pStart: {year(1815)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Between 1815 and 1820"},
		{name: "a denied start on the dated side does not block", probe: Values{pDate: {year(1817)}, pStart: {denied(year(1816))}}, cand: Values{pStart: {year(1815)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Between 1815 and 1820"},
		{name: "only a start", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {year(1815)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "After 1815"},
		{name: "only an end", probe: Values{pDate: {year(1817)}}, cand: Values{pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Before 1820"},
		{name: "start after end", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {year(1822)}, pEnd: {year(1815)}},
			prop: pDate, wantProbe: 1},
		{name: "the first leader's start", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {kept(year(1815), 1), kept(year(1812), 3)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Between 1812 and 1820"},
		{name: "an open start counts as absent", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {bound("BEF", 1815)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Before 1820"},
		{name: "an open end counts as absent", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {year(1815)}, pEnd: {bound("AFT", 1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "After 1815"},
		{name: "a negative start is skipped", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {negative(year(1815), 1)}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Before 1820"},
		{name: "a denied start is skipped", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {denied(year(1815))}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Before 1820"},
		{name: "a value with no date counts as absent", probe: Values{pDate: {year(1817)}}, cand: Values{pStart: {{Text: "x", HasText: true}}, pEnd: {year(1820)}},
			prop: pDate, wantProbe: 1, wantCand: 1, wantPeriod: "Before 1820"},
		{name: "a side whose only date is negative pools", probe: Values{pDate: {negative(year(1801), 1)}, pStart: {year(1815)}, pEnd: {year(1820)}}, cand: Values{pDate: {year(1817)}},
			prop: pDate, wantProbe: 2, wantCand: 1, wantPeriod: "Between 1815 and 1820", periodOnProbe: true},
		{name: "a property with no period", probe: Values{pStart: {year(1815)}}, cand: Values{pEnd: {year(1820)}},
			prop: pStart, wantProbe: 1, wantCand: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := poolPeriods(tt.prop, tt.probe, tt.cand)
			if len(a) != tt.wantProbe || len(b) != tt.wantCand {
				t.Fatalf("probe %d values, candidate %d; want %d, %d", len(a), len(b), tt.wantProbe, tt.wantCand)
			}
			if tt.wantPeriod == "" {
				return
			}
			side, orig := b, tt.cand[tt.prop]
			if tt.periodOnProbe {
				side, orig = a, tt.probe[tt.prop]
			}
			if got := label(side); got != tt.wantPeriod {
				t.Fatalf("period %q, want %q", got, tt.wantPeriod)
			}
			if len(side) != len(orig)+1 {
				t.Fatalf("period not appended: %d values from %d", len(side), len(orig))
			}
			if side[len(side)-1].Counted {
				t.Fatal("the period carries cache evidence")
			}
		})
	}
}

// A pooled side keeps its non-voting date values, in order, ahead of the period.
func TestPoolPeriodsKeepsNonVoting(t *testing.T) {
	neg, den, prov := negative(year(1801), 1), denied(year(1802)), cached(year(1803), 1, 0, 1, "provisional")
	probe := Values{pDate: {year(1817)}}
	cand := Values{pDate: {neg, den, prov}, pStart: {year(1815)}, pEnd: {year(1820)}}
	_, b := poolPeriods(pDate, probe, cand)
	if len(b) != 4 || !b[0].Negative || !b[1].Denied || b[2].Reason != prov.Reason || b[3].Label() != "Between 1815 and 1820" {
		t.Fatalf("%+v", b)
	}
	if len(cand[pDate]) != 3 {
		t.Fatal("the input slice grew")
	}
}

func TestPoolPeriodsMonthsAndDays(t *testing.T) {
	probe := Values{pDate: {year(1817)}}
	cand := Values{pStart: {date(1815, ip(3), ip(2))}, pEnd: {date(1820, ip(6), nil)}}
	_, b := poolPeriods(pDate, probe, cand)
	d := b[len(b)-1].Date
	if d == nil || d.Kind != datevalues.KindRange || *d.StartMonth != 3 || *d.StartDay != 2 || *d.EndMonth != 6 || d.EndDay != nil {
		t.Fatalf("%+v", d)
	}
}

func BenchmarkPoolPeriods(b *testing.B) {
	notKey := Values{pStart: {year(1815)}}
	dated := Values{pDate: {year(1817)}}
	same := Values{pDate: {year(1818)}}
	period := Values{pStart: {year(1815)}, pEnd: {year(1820)}}
	cases := []struct {
		name        string
		prop        Property
		probe, cand Values
	}{
		{"not a key", pStart, notKey, notKey},
		{"both dated", pDate, dated, same},
		{"neither dated", pDate, period, period},
		{"pooling", pDate, dated, period},
	}
	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				poolPeriods(c.prop, c.probe, c.cand)
			}
		})
	}
}
