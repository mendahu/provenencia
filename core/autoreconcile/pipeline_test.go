package autoreconcile

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/properties"
)

// Evidence helpers: a Source, weak evidence each way, high trust, a negative,
// a provisional member.
func from(src string, c Candidate) Candidate { c.SourceID = []byte(src); return c }
func lowTrust(c Candidate) Candidate         { c.Provenance.Credibility = -1; return c }
func uncertain(c Candidate) Candidate        { c.Provenance.Uncertain = true; return c }
func lowClaim(c Candidate) Candidate         { c.Provenance.ClaimConfidence = -1; return c }
func highTrust(c Candidate) Candidate        { c.Provenance.Credibility = 1; return c }
func neg(c Candidate) Candidate              { c.Negative = true; return c }
func prov(c Candidate) Candidate             { c.Provisional = true; return c }

// wantRow is an expected value: members, reason, support, against.
type wantRow struct {
	ids     []byte
	reason  Reason
	support int
	against int
}

func rowsOf(r Result) []wantRow {
	var out []wantRow
	for i, c := range r.Values {
		out = append(out, wantRow{shape(r)[i], c.Reason, c.Support, c.Against})
	}
	return out
}

// The shared passes, carried by text values (the text module never folds).
func TestPipeline(t *testing.T) {
	cases := []struct {
		group, name string
		in          []Candidate
		want        []wantRow
		state       State
	}{
		// Admit
		{"admit", "a provisional member's value keeps a row but isn't displayed",
			[]Candidate{text(1, "York"), prov(text(2, "Toronto"))},
			[]wantRow{{[]byte{1}, ReasonKept, 1, 0}, {[]byte{2}, ReasonProvisional, 1, 0}}, StateSingle},
		{"admit", "a provisional candidate with a displayed value joins its row, not its support",
			[]Candidate{text(1, "York"), prov(text(2, "York"))},
			[]wantRow{{[]byte{1}, ReasonKept, 1, 0}}, StateSingle},
		{"admit", "only provisional: rows but nothing displayed",
			[]Candidate{prov(text(1, "York"))},
			[]wantRow{{[]byte{1}, ReasonProvisional, 1, 0}}, StateEmpty},
		{"admit", "blank text is no evidence",
			[]Candidate{text(1, "   "), text(2, "York")},
			[]wantRow{{[]byte{2}, ReasonKept, 1, 0}}, StateSingle},

		// Deny
		{"deny", "a stronger negative denies the same value",
			[]Candidate{text(1, "Robbins"), highTrust(neg(text(2, "Robbins"))), text(3, "Robins")},
			[]wantRow{{[]byte{3}, ReasonKept, 1, 0}, {[]byte{1}, ReasonDenied, 1, 1}}, StateSingle},
		{"deny", "an equally strong negative only counts against",
			[]Candidate{text(1, "York"), neg(text(2, "York"))},
			[]wantRow{{[]byte{1}, ReasonKept, 1, 1}}, StateSingle},
		{"deny", "a weaker negative only counts against",
			[]Candidate{text(1, "York"), lowTrust(neg(text(2, "york")))},
			[]wantRow{{[]byte{1}, ReasonKept, 1, 1}}, StateSingle},
		{"deny", "a denied candidate casts no vote",
			[]Candidate{text(1, "B"), text(2, "B"), text(3, "A"), highTrust(neg(text(4, "B")))},
			[]wantRow{{[]byte{3}, ReasonKept, 1, 0}, {[]byte{1, 2}, ReasonDenied, 2, 1}}, StateSingle},
		{"deny", "a denied candidate with a displayed value joins its row",
			[]Candidate{lowTrust(text(1, "York")), highTrust(text(2, "York")), neg(text(3, "York"))},
			[]wantRow{{[]byte{2}, ReasonKept, 1, 1}}, StateSingle},
		{"deny", "a negative alone: nothing",
			[]Candidate{neg(text(1, "York"))},
			nil, StateEmpty},

		// Majority (Sources)
		{"majority", "two of three Sources outvote the third",
			[]Candidate{from("a", text(1, "York")), from("b", text(2, "york")), from("c", text(3, "Toronto"))},
			[]wantRow{{[]byte{1, 2}, ReasonKept, 2, 0}, {[]byte{3}, ReasonOutvoted, 1, 0}}, StateMerged},
		{"majority", "two Observations from one Source are one vote",
			[]Candidate{from("a", text(1, "York")), from("a", text(2, "York")), from("b", text(3, "Toronto"))},
			[]wantRow{{[]byte{1, 2}, ReasonKept, 1, 0}, {[]byte{3}, ReasonKept, 1, 0}}, StateMixed},
		{"majority", "two of four is not a majority",
			[]Candidate{text(1, "A"), text(2, "A"), text(3, "B"), text(4, "C")},
			[]wantRow{{[]byte{1, 2}, ReasonKept, 2, 0}, {[]byte{3}, ReasonKept, 1, 0}, {[]byte{4}, ReasonKept, 1, 0}}, StateMixed},
		{"majority", "four of five",
			[]Candidate{text(1, "B"), text(2, "A"), text(3, "A"), text(4, "A"), text(5, "A")},
			[]wantRow{{[]byte{2, 3, 4, 5}, ReasonKept, 4, 0}, {[]byte{1}, ReasonOutvoted, 1, 0}}, StateMerged},

		// Confidence
		{"confidence", "a low-trust value drops one to one",
			[]Candidate{lowTrust(text(1, "Robbins")), text(2, "Robins")},
			[]wantRow{{[]byte{2}, ReasonKept, 1, 0}, {[]byte{1}, ReasonWeak, 1, 0}}, StateSingle},
		{"confidence", "an uncertain transcription is weak",
			[]Candidate{uncertain(text(1, "Robbins")), text(2, "Robins")},
			[]wantRow{{[]byte{2}, ReasonKept, 1, 0}, {[]byte{1}, ReasonWeak, 1, 0}}, StateSingle},
		{"confidence", "a low-confidence claim is weak",
			[]Candidate{lowClaim(text(1, "Robbins")), text(2, "Robins")},
			[]wantRow{{[]byte{2}, ReasonKept, 1, 0}, {[]byte{1}, ReasonWeak, 1, 0}}, StateSingle},
		{"confidence", "all weak: nothing stronger disagrees (credibility outranks certainty)",
			[]Candidate{lowTrust(text(1, "A")), uncertain(text(2, "B"))},
			[]wantRow{{[]byte{2}, ReasonKept, 1, 0}, {[]byte{1}, ReasonKept, 1, 0}}, StateMixed},
		{"confidence", "majority runs first: a weak two of three still wins",
			[]Candidate{lowTrust(text(1, "B")), lowTrust(text(2, "B")), text(3, "A")},
			[]wantRow{{[]byte{1, 2}, ReasonKept, 2, 0}, {[]byte{3}, ReasonOutvoted, 1, 0}}, StateMerged},
		{"confidence", "a value is strong if any carrier is",
			[]Candidate{lowTrust(text(1, "A")), text(2, "A"), lowTrust(text(3, "B")), lowTrust(text(4, "C"))},
			[]wantRow{{[]byte{1, 2}, ReasonKept, 2, 0}, {[]byte{3}, ReasonWeak, 1, 0}, {[]byte{4}, ReasonWeak, 1, 0}}, StateMerged},

		// Order
		{"order", "displayed values first, even when a hidden one has more support",
			[]Candidate{prov(text(1, "B")), prov(text(2, "B")), text(3, "A")},
			[]wantRow{{[]byte{3}, ReasonKept, 1, 0}, {[]byte{1, 2}, ReasonProvisional, 2, 0}}, StateSingle},
		{"order", "equal support: the stronger record's value first",
			[]Candidate{text(1, "A"), highTrust(text(2, "B"))},
			[]wantRow{{[]byte{2}, ReasonKept, 1, 0}, {[]byte{1}, ReasonKept, 1, 0}}, StateMixed},
	}
	for _, tc := range cases {
		t.Run(tc.group+"/"+tc.name, func(t *testing.T) {
			got, err := Reconcile(properties.ValueTypeText, tc.in, nil, properties.CardinalitySingle)
			if err != nil {
				t.Fatal(err)
			}
			if have := rowsOf(got); !reflect.DeepEqual(have, tc.want) {
				t.Fatalf("got  %+v\nwant %+v", have, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
		})
	}
}

func TestPipelineMultipleCardinality(t *testing.T) {
	cases := []struct {
		name  string
		in    []Candidate
		want  []wantRow
		state State
	}{
		{"majority does not crowd out a distinct value",
			[]Candidate{from("a", text(1, "York")), from("b", text(2, "york")), from("c", text(3, "Toronto"))},
			[]wantRow{{[]byte{1, 2}, ReasonKept, 2, 0}, {[]byte{3}, ReasonKept, 1, 0}}, StateMultiple},
		{"a weak distinct value still drops",
			[]Candidate{text(1, "York"), lowTrust(text(2, "Toronto"))},
			[]wantRow{{[]byte{1}, ReasonKept, 1, 0}, {[]byte{2}, ReasonWeak, 1, 0}}, StateSingle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Reconcile(properties.ValueTypeText, tc.in, nil, properties.CardinalityMultiple)
			if err != nil {
				t.Fatal(err)
			}
			if have := rowsOf(got); !reflect.DeepEqual(have, tc.want) {
				t.Fatalf("got  %+v\nwant %+v", have, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
		})
	}
}

func TestPipelineOutcomes(t *testing.T) {
	in := []Candidate{
		text(5, "   "),                             // no evidence
		lowTrust(text(1, "York")),                  // denied, attached to the displayed York
		highTrust(text(2, "York")),                 // kept
		neg(text(3, "York")),                       // against
		prov(text(4, "Toronto")),                   // provisional, its own row
		uncertain(text(6, "U.C.")),                 // weak
		text(7, "york"),                            // kept, same value as 2
		from("x", text(8, "Upper Canada")),         // kept
		from("x", lowTrust(text(9, "Muddy York"))), // weak
	}
	got, err := Reconcile(properties.ValueTypeText, in, nil, properties.CardinalitySingle)
	if err != nil {
		t.Fatal(err)
	}
	want := map[byte]Reason{1: ReasonDenied, 2: ReasonKept, 3: ReasonAgainst, 4: ReasonProvisional, 5: ReasonNoEvidence,
		6: ReasonWeak, 7: ReasonKept, 8: ReasonKept, 9: ReasonWeak}
	if len(got.Outcomes) != len(in) {
		t.Fatalf("%d outcomes for %d candidates", len(got.Outcomes), len(in))
	}
	for i, o := range got.Outcomes {
		n := o.ObservationID[3]
		if i > 0 && got.Outcomes[i-1].ObservationID[3] >= n {
			t.Fatalf("outcomes not in id order")
		}
		if o.Reason != want[n] {
			t.Errorf("candidate %d: %q, want %q", n, o.Reason, want[n])
		}
		switch {
		case o.Reason == ReasonNoEvidence && o.Value != -1:
			t.Errorf("candidate %d: no evidence has value %d", n, o.Value)
		case o.Reason != ReasonNoEvidence && (o.Value < 0 || o.Value >= len(got.Values)):
			t.Errorf("candidate %d: value %d out of range", n, o.Value)
		}
	}
	york := got.Values[got.Outcomes[1].Value]
	if york.Value.Text != "York" || !york.Displayed() || york.Against != 1 {
		t.Fatalf("York value %+v", york)
	}
	if !reflect.DeepEqual(got.Outcomes[0].DeniedBy, id(3)) {
		t.Fatalf("denied by %v", got.Outcomes[0].DeniedBy)
	}
}

func TestPipelineConcluded(t *testing.T) {
	in := []Candidate{text(1, "A"), text(2, "A"), text(3, "B")}
	got, err := Reconcile(properties.ValueTypeText, in, &Value{Text: "b", HasText: true}, properties.CardinalitySingle)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Concluded || got.State() != StateConcluded {
		t.Fatalf("not concluded: %+v", got)
	}
	if want := []wantRow{{[]byte{3}, ReasonKept, 1, 0}, {[]byte{1, 2}, ReasonKept, 2, 0}}; !reflect.DeepEqual(rowsOf(got), want) {
		t.Fatalf("rows %+v", rowsOf(got))
	}
	if got.Values[0].Value.Text != "b" {
		t.Fatalf("concluded value %q", got.Values[0].Value.Text)
	}
	for _, o := range got.Outcomes {
		wantCluster := 1
		if o.ObservationID[3] == 3 {
			wantCluster = 0
		}
		if o.Value != wantCluster {
			t.Fatalf("candidate %d points at %d after the concluded value moved", o.ObservationID[3], o.Value)
		}
	}
}

// Generated candidate lists from fixed seeds: every list must satisfy the
// pipeline's invariants. A failure names its seed and list; shrink it into
// TestPipeline.
func TestPipelineInvariants(t *testing.T) {
	texts := []string{"York", "york", "Toronto", "U.C.", "  "}
	sources := []string{"", "a", "b", "c"}
	for _, seed := range []int64{1, 2, 3, 4} {
		rng := rand.New(rand.NewSource(seed))
		for list := 0; list < 1000; list++ {
			var in []Candidate
			for i, k := 0, rng.Intn(8); i < k; i++ {
				c := text(byte(i+1), texts[rng.Intn(len(texts))])
				c.SourceID = []byte(sources[rng.Intn(len(sources))])
				c.Negative = rng.Intn(8) == 0
				c.Provisional = rng.Intn(8) == 0
				c.Provenance = Provenance{Credibility: rng.Intn(3) - 1, Uncertain: rng.Intn(5) == 0, ClaimConfidence: rng.Intn(3) - 1}
				in = append(in, c)
			}
			where := fmt.Sprintf("seed %d list %d: %+v", seed, list, in)
			got, err := Reconcile(properties.ValueTypeText, in, nil, properties.CardinalitySingle)
			if err != nil {
				t.Fatalf("%s: %v", where, err)
			}

			shuffled := append([]Candidate(nil), in...)
			rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
			if again, _ := Reconcile(properties.ValueTypeText, shuffled, nil, properties.CardinalitySingle); !reflect.DeepEqual(again, got) {
				t.Fatalf("%s: input order changed the result", where)
			}

			if len(got.Outcomes) != len(in) {
				t.Fatalf("%s: %d outcomes", where, len(got.Outcomes))
			}
			members := map[byte]bool{}
			for _, c := range got.Values {
				if c.Support > len(c.ObservationIDs) {
					t.Fatalf("%s: support %d over %d members", where, c.Support, len(c.ObservationIDs))
				}
				for _, oid := range c.ObservationIDs {
					members[oid[3]] = true
				}
			}
			voters := false
			for _, c := range in {
				if c.Negative && members[c.ObservationID[3]] {
					t.Fatalf("%s: negative %d is a member", where, c.ObservationID[3])
				}
			}
			for _, o := range got.Outcomes {
				n := o.ObservationID[3]
				admitted := o.Reason != ReasonNoEvidence && o.Reason != ReasonAgainst
				if admitted && (o.Value < 0 || o.Value >= len(got.Values)) {
					t.Fatalf("%s: candidate %d has no value", where, n)
				}
				voters = voters || o.Reason == ReasonKept || o.Reason == ReasonOutvoted || o.Reason == ReasonWeak
			}
			if voters && got.State() == StateEmpty {
				t.Fatalf("%s: votes but nothing displayed", where)
			}
		}
	}
}

func TestProvenance(t *testing.T) {
	std := Provenance{}
	for _, tc := range []struct {
		name     string
		p, q     Provenance
		stronger bool
	}{
		{"equal is not stronger", std, std, false},
		{"credibility leads", Provenance{Credibility: 1, Uncertain: true, ClaimConfidence: -1}, std, true},
		{"lower credibility loses whatever follows", Provenance{Credibility: -1, ClaimConfidence: 1}, std, false},
		{"then certainty", std, Provenance{Uncertain: true, ClaimConfidence: 1}, true},
		{"then claim confidence", Provenance{ClaimConfidence: 1}, std, true},
		{"lower claim confidence", Provenance{ClaimConfidence: -1}, std, false},
	} {
		if got := tc.p.Stronger(tc.q); got != tc.stronger {
			t.Errorf("%s: Stronger = %v", tc.name, got)
		}
	}
	for p, weak := range map[Provenance]bool{
		std:                                  false,
		{Credibility: 1}:                     false,
		{Credibility: -1}:                    true,
		{Uncertain: true}:                    true,
		{ClaimConfidence: -1}:                true,
		{Credibility: 1, ClaimConfidence: 1}: false,
	} {
		if p.Weak() != weak {
			t.Errorf("%+v: Weak = %v", p, !weak)
		}
	}
}

// An outvoted candidate carries the vote that beat it; nothing else does.
func TestOutvotedVote(t *testing.T) {
	votes := func(r Result) map[byte]Vote {
		out := map[byte]Vote{}
		for _, o := range r.Outcomes {
			out[o.ObservationID[len(o.ObservationID)-1]] = o.Vote
		}
		return out
	}
	cases := []struct {
		name      string
		valueType string
		in        []Candidate
		want      map[byte]Vote
	}{
		{"two of three Sources", properties.ValueTypeText,
			[]Candidate{from("a", text(1, "York")), from("b", text(2, "york")), from("c", text(3, "Toronto"))},
			map[byte]Vote{1: {}, 2: {}, 3: {Support: 2, Of: 3}}},
		{"four of five, one Source each", properties.ValueTypeText,
			[]Candidate{text(1, "B"), text(2, "A"), text(3, "A"), text(4, "A"), text(5, "A")},
			map[byte]Vote{1: {Support: 4, Of: 5}, 2: {}, 3: {}, 4: {}, 5: {}}},
		{"a Source's records are one vote", properties.ValueTypeText,
			[]Candidate{from("a", text(1, "A")), from("a", text(2, "A")), from("b", text(3, "A")), from("c", text(4, "B"))},
			map[byte]Vote{1: {}, 2: {}, 3: {}, 4: {Support: 2, Of: 3}}},
		{"weak and kept carry none", properties.ValueTypeText,
			[]Candidate{lowTrust(text(1, "Robbins")), text(2, "Robins")},
			map[byte]Vote{1: {}, 2: {}}},
		{"a name outvoted on its surname carries that part's vote", properties.ValueTypeName,
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|surname=Robbins"), nm(4, "given=Jim|surname=Robins")},
			map[byte]Vote{1: {}, 2: {}, 3: {Support: 3, Of: 4}, 4: {}}},
	}
	for _, tc := range cases {
		got, err := Reconcile(tc.valueType, tc.in, nil, properties.CardinalitySingle)
		if err != nil {
			t.Fatal(err)
		}
		if g := votes(got); !reflect.DeepEqual(g, tc.want) {
			t.Errorf("%s: votes %+v, want %+v", tc.name, g, tc.want)
		}
		for _, o := range got.Outcomes {
			if (o.Reason == ReasonOutvoted) != (o.Vote != Vote{}) {
				t.Errorf("%s: %+v", tc.name, o)
			}
		}
	}
}
