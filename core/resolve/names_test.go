package resolve

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

// nm builds a name candidate from "type=value" parts separated by "|", with a
// form that says nothing, so only parts can carry the test. Type keys are
// written out in full: the reconciler knows no type, so "a=x" works as well.
func nm(n byte, spec string) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{Name: parse(spec)}}
}

func parse(spec string) *namevalues.Value {
	v := &namevalues.Value{Form: "(as written)"}
	for i, item := range strings.Split(spec, "|") {
		typ, val, ok := strings.Cut(item, "=")
		if !ok {
			panic("bad part " + item)
		}
		v.Parts = append(v.Parts, namevalues.Part{Idx: i, Type: typ, Value: val})
	}
	return v
}

// withNameForm replaces a candidate's form.
func withNameForm(c Candidate, form string) Candidate {
	n := *c.Value.Name
	n.Form = form
	c.Value.Name = &n
	return c
}

// spec renders a name's parts back to "type=value|…".
func spec(n *namevalues.Value) string {
	var out []string
	for _, p := range n.Parts {
		out = append(out, p.Type+"="+p.Value)
	}
	return strings.Join(out, "|")
}

// cl is one expected cluster: member ids and the reconciled parts.
type cl struct {
	ids   []byte
	parts string
}

func TestReconcileNames(t *testing.T) {
	cases := []struct {
		group, name string
		in          []Candidate
		want        []cl
		state       State
		reasons     string // each row's Reason, in order
	}{
		// Exact
		{"exact", "single name",
			[]Candidate{nm(1, "given=James|surname=Robins")},
			[]cl{{[]byte{1}, "given=James|surname=Robins"}}, StateSingle, "kept"},
		{"exact", "identical names merge",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged, "kept"},
		{"exact", "case, punctuation and spacing are ignored; best rank's spelling kept",
			[]Candidate{nm(1, "given=JAMES|surname=Robins."), nm(2, "given=james|surname=  robins")},
			[]cl{{[]byte{1, 2}, "given=JAMES|surname=Robins."}}, StateMerged, "kept"},
		{"exact", "a hyphenated part is one unit",
			[]Candidate{nm(1, "given=Ann|surname=Smith-Jones"), nm(2, "given=Ann|surname=Smith Jones")},
			[]cl{{[]byte{1, 2}, "given=Ann|surname=Smith-Jones"}}, StateMerged, "kept"},
		{"exact", "one hyphenated part equals two parts (words are compared)",
			[]Candidate{nm(1, "given=Ann|surname=Smith-Jones"), nm(2, "given=Ann|surname=Smith|surname=Jones")},
			[]cl{{[]byte{1, 2}, "given=Ann|surname=Smith-Jones"}}, StateMerged, "kept"},
		{"exact", "parts of one type compare in idx order, wherever they sit",
			[]Candidate{nm(1, "given=James|surname=Robins|given=Kenneth"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins|given=Kenneth"}}, StateMerged, "kept"},
		{"exact", "an initial with and without its period",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=J|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=J.|surname=Robins"}}, StateMerged, "kept"},

		// Subsumption
		{"subsumption", "an initial expands: [J] into [James]",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "the expansion does not depend on rank",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=J.|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "[James] into [James, Kenneth]",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "[J, K] into [James, Kenneth]",
			[]Candidate{nm(1, "given=J|given=K|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "[James, K] into [James, Kenneth]",
			[]Candidate{nm(1, "given=James|given=K.|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "a later part alone: [Kenneth] into [James, Kenneth]",
			[]Candidate{nm(1, "given=Kenneth|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "a chain folds to the fullest: [J], [James], [James, Kenneth]",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2, 3}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "order matters: [Kenneth, James] does not fold into [James, Kenneth]; each part appears once",
			[]Candidate{nm(1, "given=Kenneth|given=James|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=Kenneth|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "different middle names do not fold; both stay, each part once",
			[]Candidate{nm(1, "given=James|given=Kevin|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kevin|given=Kenneth|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "an initial of another letter does not fold",
			[]Candidate{nm(1, "given=W.|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=W.|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "two initials of different letters both stay",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=K.|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=J.|given=K.|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "a lone uncased character is a whole name, not an initial",
			[]Candidate{nm(1, "surname=蒋|given=浩"), nm(2, "surname=蒋|given=浩然")},
			[]cl{{[]byte{1, 2}, "surname=蒋|given=浩|given=浩然"}}, StateMerged, "kept"},
		{"subsumption", "an ambiguous initial folds into the best ranked of equal support",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=John|surname=Robins"), nm(3, "given=J.|surname=Robins")},
			[]cl{{[]byte{1, 2, 3}, "given=James|given=John|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "an ambiguous initial folds into the best supported",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=John|surname=Robins"), nm(3, "given=John|surname=Robins"), nm(4, "given=J.|surname=Robins")},
			[]cl{{[]byte{1, 2, 3, 4}, "given=John|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"subsumption", "an initial never folds across types",
			[]Candidate{nm(1, "nick=J|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|nick=J|surname=Robins"}}, StateMerged, "kept"},

		// Majority
		{"majority", "two of three crowd out a misspelling",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|surname=Robbins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}, {[]byte{3}, "given=James|surname=Robbins"}}, StateMerged, "kept outvoted"},
		{"majority", "rank does not save an outlier",
			[]Candidate{nm(1, "given=James|surname=Robbins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|surname=Robins")},
			[]cl{{[]byte{2, 3}, "given=James|surname=Robins"}, {[]byte{1}, "given=James|surname=Robbins"}}, StateMerged, "kept outvoted"},
		{"majority", "four of five",
			[]Candidate{nm(1, "surname=Robbins"), nm(2, "surname=Robins"), nm(3, "surname=Robins"), nm(4, "surname=Robins"), nm(5, "surname=Robins")},
			[]cl{{[]byte{2, 3, 4, 5}, "surname=Robins"}, {[]byte{1}, "surname=Robbins"}}, StateMerged, "kept outvoted"},
		{"majority", "one to one keeps both in the name",
			[]Candidate{nm(1, "given=Mary|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=Mary|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"majority", "two of four is not a majority",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=Mary|surname=Robins"), nm(4, "given=Ann|surname=Robins")},
			[]cl{{[]byte{1, 2, 3, 4}, "given=James|given=Mary|given=Ann|surname=Robins"}}, StateMerged, "kept"},
		{"majority", "eliminated candidates group by exact parts",
			[]Candidate{nm(1, "surname=Robins"), nm(2, "surname=Robins"), nm(3, "surname=Robins"), nm(4, "surname=Robbins"), nm(5, "surname=ROBBINS")},
			[]cl{{[]byte{1, 2, 3}, "surname=Robins"}, {[]byte{4, 5}, "surname=Robbins"}}, StateMerged, "kept outvoted"},
		{"majority", "each type is elected on its own",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robbins"), nm(3, "given=Jim|surname=Robins")},
			[]cl{{[]byte{1, 3}, "given=James|given=Jim|surname=Robins"}, {[]byte{2}, "given=James|surname=Robbins"}}, StateMerged, "kept outvoted"},
		{"majority", "folded support counts, and the winner's parts can come from a loser",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=J.|surname=Robins"), nm(3, "given=James|surname=Robbins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}, {[]byte{3}, "given=James|surname=Robbins"}}, StateMerged, "kept outvoted"},

		// No type is special
		{"types", "prefixes that differ both stay in the name",
			[]Candidate{nm(1, "prefix=Rev.|given=James|surname=Robins"), nm(2, "prefix=Dr.|given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "prefix=Rev.|prefix=Dr.|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"types", "a different prefix is not outvoted",
			[]Candidate{nm(1, "prefix=Rev.|given=James|surname=Robins"), nm(2, "prefix=Rev.|given=James|surname=Robins"), nm(3, "prefix=Dr.|given=James|surname=Robins")},
			[]cl{{[]byte{1, 2, 3}, "prefix=Rev.|prefix=Dr.|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"types", "suffixes that differ both stay in the name",
			[]Candidate{nm(1, "given=James|surname=Robins|suffix=Jr."), nm(2, "given=James|surname=Robins|suffix=Sr.")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins|suffix=Jr.|suffix=Sr."}}, StateMerged, "kept"},
		{"types", "James and Jim, both given, both stay in the name",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=Jim|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Jim|surname=Robins"}}, StateMerged, "kept"},
		{"types", "a nickname beside a given name is kept",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|nick=Jim|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|nick=Jim|surname=Robins"}}, StateMerged, "kept"},
		{"types", "made-up type keys work the same",
			[]Candidate{nm(1, "a=James|b=Robins"), nm(2, "a=James|b=Robins"), nm(3, "a=James|b=Robbins")},
			[]cl{{[]byte{1, 2}, "a=James|b=Robins"}, {[]byte{3}, "a=James|b=Robbins"}}, StateMerged, "kept outvoted"},
		{"types", "untyped parts are a type like any other",
			[]Candidate{nm(1, "=J|=Robins"), nm(2, "=James|=Robins")},
			[]cl{{[]byte{1, 2}, "=James|=Robins"}}, StateMerged, "kept"},
		{"types", "untyped parts do not meet typed ones",
			[]Candidate{nm(1, "=James|=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "=James|=Robins|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"types", "undetermined is a type like any other",
			[]Candidate{nm(1, "undetermined=Kendall|surname=Robins"), nm(2, "undetermined=Kendal|surname=Robins")},
			[]cl{{[]byte{1, 2}, "undetermined=Kendall|undetermined=Kendal|surname=Robins"}}, StateMerged, "kept"},
		{"types", "accented and plain spellings both stay in the name",
			[]Candidate{nm(1, "given=José|surname=Silva"), nm(2, "given=Jose|surname=Silva")},
			[]cl{{[]byte{1, 2}, "given=José|given=Jose|surname=Silva"}}, StateMerged, "kept"},

		// Absence is not disagreement
		{"absence", "a surname alone joins the fuller name",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged, "kept"},
		{"absence", "a given name alone joins",
			[]Candidate{nm(1, "given=James"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged, "kept"},
		{"absence", "a surname alone joins the one name",
			[]Candidate{nm(1, "given=Mary|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "surname=Robins")},
			[]cl{{[]byte{1, 2, 3}, "given=Mary|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"absence", "a different given name is not outvoted by a fuller one",
			[]Candidate{nm(1, "given=Mary|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|given=K.|surname=Robins"), nm(4, "surname=Robins")},
			[]cl{{[]byte{1, 2, 3, 4}, "given=James|given=K.|given=Mary|surname=Robins"}}, StateMerged, "kept"},
		{"absence", "a candidate sharing no type joins",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "prefix=Rev.")},
			[]cl{{[]byte{1, 2}, "prefix=Rev.|given=James|surname=Robins"}}, StateMerged, "kept"},

		// Form is a transcription
		{"form", "a name with no parts is dropped",
			[]Candidate{name(1, "James Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{2}, "given=James|surname=Robins"}}, StateSingle, "kept"},
		{"form", "only names with no parts: empty",
			[]Candidate{name(1, "James Robins"), name(2, "James Robins")},
			nil, StateEmpty, ""},
		{"form", "parts with nothing in them are no parts",
			[]Candidate{nm(1, "given=.|surname= "), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{2}, "given=James|surname=Robins"}}, StateSingle, "kept"},
		{"form", "an empty part is skipped",
			[]Candidate{nm(1, "given=.|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged, "kept"},
		{"form", "the same parts under different forms merge",
			[]Candidate{withNameForm(nm(1, "given=James|surname=Robins"), "Robins, James"), withNameForm(nm(2, "given=James|surname=Robins"), "Jas. Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged, "kept"},
		{"form", "different parts under one form both stay in the name",
			[]Candidate{withNameForm(nm(1, "given=James|surname=Robins"), "J. Robins"), withNameForm(nm(2, "given=John|surname=Robins"), "J. Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=John|surname=Robins"}}, StateMerged, "kept"},

		// Assembly
		{"assembly", "a type the base lacks goes after its predecessor",
			[]Candidate{nm(1, "prefix=Rev.|given=James|surname=Robins"), nm(2, "given=James|nick=Jim|surname=Robins")},
			[]cl{{[]byte{1, 2}, "prefix=Rev.|given=James|nick=Jim|surname=Robins"}}, StateMerged, "kept"},
		{"assembly", "the base is the member with the most types",
			[]Candidate{nm(1, "surname=Robins|given=James"), nm(2, "given=James|surname=Robins|suffix=Jr.")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins|suffix=Jr."}}, StateMerged, "kept"},
		{"assembly", "each type's spelling comes from its best-ranked carrier",
			[]Candidate{nm(1, "given=J.|surname=ROBINS|suffix=Jr."), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=ROBINS|suffix=Jr."}}, StateMerged, "kept"},
		{"assembly", "a surname-first name keeps its order",
			[]Candidate{nm(1, "surname=蒋|given=浩"), nm(2, "surname=蒋")},
			[]cl{{[]byte{1, 2}, "surname=蒋|given=浩"}}, StateMerged, "kept"},
	}
	for _, tc := range cases {
		t.Run(tc.group+"/"+tc.name, func(t *testing.T) {
			got, err := Resolve(properties.ValueTypeName, tc.in, nil)
			if err != nil {
				t.Fatal(err)
			}
			var have []cl
			for i, c := range got.Clusters {
				have = append(have, cl{shape(got)[i], spec(c.Value.Name)})
			}
			if !reflect.DeepEqual(have, tc.want) {
				t.Fatalf("got  %v\nwant %v", have, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
			if r := reasonsOf(got); r != tc.reasons {
				t.Fatalf("reasons %q, want %q", r, tc.reasons)
			}
		})
	}
}

// Provenance helpers: weak evidence each way, high trust, and a negative.

// clAgainst is an expected cluster with its count of negatives against it.
type clAgainst struct {
	ids     []byte
	parts   string
	against int
}

func TestReconcileNamesProvenance(t *testing.T) {
	cases := []struct {
		group, name string
		in          []Candidate
		want        []clAgainst
		state       State
		reasons     string // each row's Reason, in order
	}{
		// Confidence: weak values drop when a non-weak value disagrees.
		{"confidence", "a low-trust outlier drops one to one",
			[]Candidate{lowTrust(nm(1, "given=James|surname=Robbins")), nm(2, "given=James|surname=Robins")},
			[]clAgainst{{[]byte{2}, "given=James|surname=Robins", 0}, {[]byte{1}, "given=James|surname=Robbins", 0}}, StateSingle, "kept weak"},
		{"confidence", "an uncertain transcription drops",
			[]Candidate{uncertain(nm(1, "given=Mary|surname=Robins")), nm(2, "given=James|surname=Robins")},
			[]clAgainst{{[]byte{2}, "given=James|surname=Robins", 0}, {[]byte{1}, "given=Mary|surname=Robins", 0}}, StateSingle, "kept weak"},
		{"confidence", "a low-confidence claim drops",
			[]Candidate{lowClaim(nm(1, "given=Mary|surname=Robins")), nm(2, "given=James|surname=Robins")},
			[]clAgainst{{[]byte{2}, "given=James|surname=Robins", 0}, {[]byte{1}, "given=Mary|surname=Robins", 0}}, StateSingle, "kept weak"},
		{"confidence", "two weak values both stay when nothing stronger disagrees",
			[]Candidate{lowTrust(nm(1, "given=Mary|surname=Robins")), uncertain(nm(2, "given=James|surname=Robins"))},
			[]clAgainst{{[]byte{1, 2}, "given=James|given=Mary|surname=Robins", 0}}, StateMerged, "kept"},
		{"confidence", "all weak and agreeing still reconcile",
			[]Candidate{lowTrust(nm(1, "given=J.|surname=Robins")), lowClaim(nm(2, "given=James|surname=Robins"))},
			[]clAgainst{{[]byte{1, 2}, "given=James|surname=Robins", 0}}, StateMerged, "kept"},
		{"confidence", "majority runs first: a weak two of three still wins",
			[]Candidate{lowTrust(nm(1, "surname=Robbins")), lowTrust(nm(2, "surname=Robbins")), nm(3, "surname=Robins")},
			[]clAgainst{{[]byte{1, 2}, "surname=Robbins", 0}, {[]byte{3}, "surname=Robins", 0}}, StateMerged, "kept outvoted"},
		{"confidence", "a folded value is strong if any carrier is",
			[]Candidate{lowTrust(nm(1, "given=Mary|surname=Robins")), lowTrust(nm(2, "given=Mary|surname=Robins")),
				nm(3, "given=J.|surname=Robins"), lowTrust(nm(4, "given=James|surname=Robins"))},
			[]clAgainst{{[]byte{3, 4}, "given=James|surname=Robins", 0}, {[]byte{1, 2}, "given=Mary|surname=Robins", 0}}, StateMerged, "kept weak"},
		{"confidence", "a weak candidate's rival in one type takes its other parts with it",
			[]Candidate{lowTrust(nm(1, "prefix=Dr.|given=James|surname=Robins")), nm(2, "prefix=Rev.|given=James")},
			[]clAgainst{{[]byte{2}, "prefix=Rev.|given=James", 0}, {[]byte{1}, "prefix=Dr.|given=James|surname=Robins", 0}}, StateSingle, "kept weak"},
		{"confidence", "high trust against standard is not weak against strong",
			[]Candidate{nm(1, "given=Mary|surname=Robins"), highTrust(nm(2, "given=James|surname=Robins"))},
			[]clAgainst{{[]byte{1, 2}, "given=James|given=Mary|surname=Robins", 0}}, StateMerged, "kept"},

		// Negatives deny weaker positives with the same parts.
		{"negatives", "a stronger negative eliminates the same name",
			[]Candidate{nm(1, "given=James|surname=Robbins"), highTrust(neg(nm(2, "given=James|surname=Robbins"))), nm(3, "given=James|surname=Robins")},
			[]clAgainst{{[]byte{3}, "given=James|surname=Robins", 0}, {[]byte{1}, "given=James|surname=Robbins", 1}}, StateSingle, "kept denied"},
		{"negatives", "an equally strong negative only counts against",
			[]Candidate{nm(1, "given=James|surname=Robins"), neg(nm(2, "given=James|surname=Robins"))},
			[]clAgainst{{[]byte{1}, "given=James|surname=Robins", 1}}, StateSingle, "kept"},
		{"negatives", "a weaker negative only counts against",
			[]Candidate{nm(1, "given=James|surname=Robins"), lowTrust(neg(nm(2, "given=James|surname=Robins")))},
			[]clAgainst{{[]byte{1}, "given=James|surname=Robins", 1}}, StateSingle, "kept"},
		{"negatives", "it eliminates only the weaker of two equal names",
			[]Candidate{lowTrust(nm(1, "given=James|surname=Robins")), highTrust(nm(2, "given=James|surname=Robins")), neg(nm(3, "given=James|surname=Robins"))},
			[]clAgainst{{[]byte{2}, "given=James|surname=Robins", 1}}, StateSingle, "kept"},
		{"negatives", "a negative of other parts does nothing",
			[]Candidate{nm(1, "given=James|surname=Robins"), highTrust(neg(nm(2, "given=Mary|surname=Robins")))},
			[]clAgainst{{[]byte{1}, "given=James|surname=Robins", 0}}, StateSingle, "kept"},
		{"negatives", "a negative with no parts is ignored",
			[]Candidate{lowTrust(nm(1, "given=James|surname=Robins")), neg(name(2, "James Robins"))},
			[]clAgainst{{[]byte{1}, "given=James|surname=Robins", 0}}, StateSingle, "kept"},
		{"negatives", "a negative counts against the reconciled name too",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=James"), lowTrust(neg(nm(3, "given=James|surname=Robins")))},
			[]clAgainst{{[]byte{1, 2}, "given=James|surname=Robins", 1}}, StateMerged, "kept"},
		{"negatives", "a denied candidate casts no vote",
			[]Candidate{lowTrust(nm(1, "surname=Robbins")), lowTrust(nm(2, "surname=Robbins")), nm(3, "surname=Robins"), neg(nm(4, "surname=Robbins"))},
			[]clAgainst{{[]byte{3}, "surname=Robins", 0}, {[]byte{1, 2}, "surname=Robbins", 1}}, StateSingle, "kept denied"},
		{"negatives", "only negatives: empty",
			[]Candidate{neg(nm(1, "given=James|surname=Robins"))},
			nil, StateEmpty, ""},

		// Rank order: provenance before id.
		{"rank", "the stronger record's spelling is kept",
			[]Candidate{nm(1, "given=JAMES|surname=ROBINS"), highTrust(nm(2, "given=James|surname=Robins"))},
			[]clAgainst{{[]byte{1, 2}, "given=James|surname=Robins", 0}}, StateMerged, "kept"},
		{"rank", "an ambiguous initial folds toward the stronger record",
			[]Candidate{nm(1, "given=James|surname=Robins"), highTrust(nm(2, "given=John|surname=Robins")), nm(3, "given=J.|surname=Robins")},
			[]clAgainst{{[]byte{1, 2, 3}, "given=John|given=James|surname=Robins", 0}}, StateMerged, "kept"},
	}
	for _, tc := range cases {
		t.Run(tc.group+"/"+tc.name, func(t *testing.T) {
			got, err := Resolve(properties.ValueTypeName, tc.in, nil)
			if err != nil {
				t.Fatal(err)
			}
			var have []clAgainst
			for i, c := range got.Clusters {
				have = append(have, clAgainst{shape(got)[i], spec(c.Value.Name), c.Against})
			}
			if !reflect.DeepEqual(have, tc.want) {
				t.Fatalf("got  %v\nwant %v", have, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
			if r := reasonsOf(got); r != tc.reasons {
				t.Fatalf("reasons %q, want %q", r, tc.reasons)
			}
		})
	}
}

// A cluster whose reconciled parts are exactly one member's is that member's
// own value, form and all; otherwise the form is the parts in order.
func TestReconcileNamesValue(t *testing.T) {
	a := withNameForm(nm(1, "given=J.|surname=Robins"), "J. Robins")
	b := withNameForm(nm(2, "given=James|surname=Robins"), "Robins, James")
	got, err := Resolve(properties.ValueTypeName, []Candidate{a, b}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Clusters[0].Value.Name != b.Value.Name {
		t.Fatalf("want member 2's own value, got %+v", got.Clusters[0].Value.Name)
	}

	c := withNameForm(nm(3, "given=James|surname=Robins|suffix=Jr."), "x")
	d := withNameForm(nm(4, "given=James|nick=Jim|surname=Robins"), "y")
	got, err = Resolve(properties.ValueTypeName, []Candidate{c, d}, nil)
	if err != nil {
		t.Fatal(err)
	}
	n := got.Clusters[0].Value.Name
	if n.Form != "James Jim Robins Jr." || len(n.ID) != 0 {
		t.Fatalf("assembled %+v", n)
	}
	for i, p := range n.Parts {
		if p.Idx != i || len(p.ID) != 0 {
			t.Fatalf("part %d: %+v", i, p)
		}
	}
}

// Generated name lists from fixed seeds: every list must satisfy the
// reconciler's invariants. A failure names its seed and list; shrink it into
// TestReconcileNames.
func TestReconcileNamesInvariants(t *testing.T) {
	pools := map[string][]string{
		"given":   {"James", "J.", "Jim", "Mary", "M", "Kenneth", "K."},
		"surname": {"Robins", "Robbins", "ROBINS", "Smith"},
		"prefix":  {"Rev.", "Dr."},
		"suffix":  {"Jr.", "Sr."},
		"nick":    {"Jim"},
		"":        {"James", "Robins"},
	}
	types := []string{"given", "given", "surname", "surname", "prefix", "suffix", "nick", ""}
	forms := []string{"James Robins", "J. Robins", ""}

	for _, seed := range []int64{1, 2, 3, 4} {
		rng := rand.New(rand.NewSource(seed))
		gen := func(n byte) Candidate {
			v := &namevalues.Value{Form: forms[rng.Intn(len(forms))]}
			for i, k := 0, rng.Intn(5); i < k; i++ {
				typ := types[rng.Intn(len(types))]
				pool := pools[typ]
				v.Parts = append(v.Parts, namevalues.Part{Idx: i, Type: typ, Value: pool[rng.Intn(len(pool))]})
			}
			return Candidate{ObservationID: id(n), Value: Value{Name: v}, Negative: rng.Intn(8) == 0,
				Provenance: Provenance{Credibility: rng.Intn(3) - 1, Uncertain: rng.Intn(5) == 0, ClaimConfidence: rng.Intn(3) - 1}}
		}
		for list := 0; list < 500; list++ {
			var in []Candidate
			for i, k := 0, rng.Intn(8); i < k; i++ {
				in = append(in, gen(byte(i+1)))
			}
			where := fmt.Sprintf("seed %d list %d: %s", seed, list, describe(in))
			got, err := Resolve(properties.ValueTypeName, in, nil)
			if err != nil {
				t.Fatalf("%s: %v", where, err)
			}

			// Order independent.
			shuffled := append([]Candidate(nil), in...)
			rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
			if again, _ := Resolve(properties.ValueTypeName, shuffled, nil); !reflect.DeepEqual(again, got) {
				t.Fatalf("%s: input order changed the result", where)
			}

			// Every structured positive candidate in exactly one cluster; no
			// negative in any.
			seen := map[byte]int{}
			for _, c := range got.Clusters {
				for _, oid := range c.ObservationIDs {
					seen[oid[3]]++
				}
			}
			for _, c := range in {
				want := 1
				if c.Negative || signatureOf(c.Value.Name) == "" {
					want = 0
				}
				if seen[c.ObservationID[3]] != want {
					t.Fatalf("%s: candidate %d in %d clusters, want %d", where, c.ObservationID[3], seen[c.ObservationID[3]], want)
				}
			}

			// Every reconciled part is some candidate's part of that type.
			had := map[string]bool{}
			for _, c := range in {
				for _, p := range c.Value.Name.Parts {
					had[p.Type+"="+NormalizeForm(p.Value)] = true
				}
			}
			for _, c := range got.Clusters {
				for _, p := range c.Value.Name.Parts {
					if !had[p.Type+"="+NormalizeForm(p.Value)] {
						t.Fatalf("%s: invented part %s=%s", where, p.Type, p.Value)
					}
				}
			}

			// Forms never matter.
			reformed := make([]Candidate, len(in))
			for i, c := range in {
				reformed[i] = withNameForm(c, "anything "+strings.Repeat("x", i))
			}
			if again, _ := Resolve(properties.ValueTypeName, reformed, nil); !reflect.DeepEqual(specs(again), specs(got)) {
				t.Fatalf("%s: forms changed the result", where)
			}

			// Renaming types one-to-one renames the result and nothing else.
			rename := map[string]string{"given": "x", "surname": "given", "prefix": "", "suffix": "nick", "nick": "suffix", "": "surname"}
			renamed := make([]Candidate, len(in))
			for i, c := range in {
				n := *c.Value.Name
				n.Parts = append([]namevalues.Part(nil), n.Parts...)
				for j := range n.Parts {
					n.Parts[j].Type = rename[n.Parts[j].Type]
				}
				renamed[i] = c
				renamed[i].Value = Value{Name: &n}
			}
			if again, _ := Resolve(properties.ValueTypeName, renamed, nil); !reflect.DeepEqual(shape(again), shape(got)) {
				t.Fatalf("%s: renaming types changed the clusters", where)
			}
		}
	}
}

func describe(in []Candidate) string {
	var out []string
	for _, c := range in {
		out = append(out, fmt.Sprintf("%d:%q%+v neg=%v", c.ObservationID[3], spec(c.Value.Name), c.Provenance, c.Negative))
	}
	return strings.Join(out, " ")
}

// specs is each cluster's ids and parts, without forms.
func specs(r Result) []string {
	var out []string
	for i, c := range r.Clusters {
		out = append(out, fmt.Sprint(shape(r)[i], signatureOf(c.Value.Name)))
	}
	return out
}

func TestIsInitial(t *testing.T) {
	for w, want := range map[string]bool{"j": true, "J": true, "é": true, "jo": false, "蒋": false, "": false, "1": false} {
		if IsInitial(w) != want {
			t.Errorf("IsInitial(%q) = %v", w, !want)
		}
	}
}

// BenchmarkReconcileNames: one handle with 200 name candidates, the shape a
// heavily researched Person reaches (performance-ledger.md).
func BenchmarkReconcileNames(b *testing.B) {
	given := []string{"James", "J.", "Jim", "James|given=Kenneth", "J.|given=K."}
	surname := []string{"Robins", "Robins", "Robins", "Robbins"}
	var in []Candidate
	for i := 0; i < 200; i++ {
		c := nm(0, "given="+given[i%len(given)]+"|surname="+surname[i%len(surname)])
		c.ObservationID = []byte{0, 0, byte(i >> 8), byte(i)}
		in = append(in, c)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Resolve(properties.ValueTypeName, in, nil); err != nil {
			b.Fatal(err)
		}
	}
}

// signatureOf is a name's parts by type, as the name module compares them;
// "" for a name with no parts.
func signatureOf(n *namevalues.Value) string {
	units, ok := nameModule{}.split(Value{Name: n})
	if !ok {
		return ""
	}
	return signature(units)
}

func reasonsOf(r Result) string {
	var out []string
	for _, c := range r.Clusters {
		out = append(out, string(c.Reason))
	}
	return strings.Join(out, " ")
}

// One name per Person: values that disagree within a part type combine into
// it, unless one is a spelling variant of a majority winner or only weak
// evidence backs it.
func TestReconcileNamesCombine(t *testing.T) {
	cases := []struct {
		name    string
		in      []Candidate
		want    []cl
		state   State
		reasons string
	}{
		{"a nickname and a given name from different records combine",
			[]Candidate{nm(1, "nick=Jake|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|nick=Jake|surname=Robins"}}, StateMerged, "kept"},
		{"a nickname recorded as a given name combines with the given name",
			[]Candidate{nm(1, "given=Jake|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=Jake|given=James|surname=Robins"}}, StateMerged, "kept"},
		{"a different given name in the minority is not outvoted",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=Jake|surname=Robins")},
			[]cl{{[]byte{1, 2, 3}, "given=James|given=Jake|surname=Robins"}}, StateMerged, "kept"},
		{"a married surname stays beside the birth surname",
			[]Candidate{nm(1, "given=Mary|surname=Smith"), nm(2, "given=Mary|surname=Robins"), nm(3, "given=Mary|surname=Robins")},
			[]cl{{[]byte{1, 2, 3}, "given=Mary|surname=Robins|surname=Smith"}}, StateMerged, "kept"},
		{"a misspelt surname is outvoted, a different given name is not",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|surname=Robbins"), nm(4, "given=Jim|surname=Robins")},
			[]cl{{[]byte{1, 2, 4}, "given=James|given=Jim|surname=Robins"}, {[]byte{3}, "given=James|surname=Robbins"}}, StateMerged, "kept outvoted"},
		{"a misspelling ties: both stay",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robbins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins|surname=Robbins"}}, StateMerged, "kept"},
		{"weak evidence for a different name still drops",
			[]Candidate{lowTrust(nm(1, "given=Jake|surname=Robins")), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{2}, "given=James|surname=Robins"}, {[]byte{1}, "given=Jake|surname=Robins"}}, StateSingle, "kept weak"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(properties.ValueTypeName, tc.in, nil)
			if err != nil {
				t.Fatal(err)
			}
			var have []cl
			for i, c := range got.Clusters {
				have = append(have, cl{shape(got)[i], spec(c.Value.Name)})
			}
			if !reflect.DeepEqual(have, tc.want) {
				t.Fatalf("got  %v\nwant %v", have, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
			if r := reasonsOf(got); r != tc.reasons {
				t.Fatalf("reasons %q, want %q", r, tc.reasons)
			}
		})
	}
}

// Parts compare as words: a hyphen or a space inside a part is the same as
// separate parts. The recorded parts are what's displayed.
func TestReconcileNamesWords(t *testing.T) {
	cases := []struct {
		name    string
		in      []Candidate
		want    []cl
		state   State
		reasons string
	}{
		{"hyphenated, spaced and separate given names are one value",
			[]Candidate{nm(1, "given=Mary-Ann|surname=Lee"), nm(2, "given=Mary Ann|surname=Lee"), nm(3, "given=Mary|given=Ann|surname=Lee")},
			[]cl{{[]byte{1, 2, 3}, "given=Mary-Ann|surname=Lee"}}, StateMerged, "kept"},
		{"the best-ranked record's spelling is displayed",
			[]Candidate{nm(1, "given=Mary|given=Ann|surname=Lee"), nm(2, "given=Mary-Ann|surname=Lee")},
			[]cl{{[]byte{1, 2}, "given=Mary|given=Ann|surname=Lee"}}, StateMerged, "kept"},
		{"a maiden and a hyphenated married surname fold into one",
			[]Candidate{nm(1, "given=Mary|surname=Smith"), nm(2, "given=Mary|surname=Smith-Jones"), nm(3, "given=Mary|surname=Jones")},
			[]cl{{[]byte{1, 2, 3}, "given=Mary|surname=Smith-Jones"}}, StateMerged, "kept"},
		{"the same words in another order add nothing new to show",
			[]Candidate{nm(1, "given=Mary|surname=Jones-Smith"), nm(2, "given=Mary|surname=Smith-Jones")},
			[]cl{{[]byte{1, 2}, "given=Mary|surname=Jones-Smith"}}, StateMerged, "kept"},
		{"a misspelt word in a hyphenated surname is outvoted",
			[]Candidate{nm(1, "given=Mary|surname=Smith-Jones"), nm(2, "given=Mary|surname=Smith-Jones"), nm(3, "given=Mary|surname=Smyth-Jones")},
			[]cl{{[]byte{1, 2}, "given=Mary|surname=Smith-Jones"}, {[]byte{3}, "given=Mary|surname=Smyth-Jones"}}, StateMerged, "kept outvoted"},
		{"a different second surname stays",
			[]Candidate{nm(1, "given=Mary|surname=Smith-Jones"), nm(2, "given=Mary|surname=Smith-Jones"), nm(3, "given=Mary|surname=Smith-Brown")},
			[]cl{{[]byte{1, 2, 3}, "given=Mary|surname=Smith-Jones|surname=Smith-Brown"}}, StateMerged, "kept"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(properties.ValueTypeName, tc.in, nil)
			if err != nil {
				t.Fatal(err)
			}
			var have []cl
			for i, c := range got.Clusters {
				have = append(have, cl{shape(got)[i], spec(c.Value.Name)})
			}
			if !reflect.DeepEqual(have, tc.want) {
				t.Fatalf("got  %v\nwant %v", have, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
			if r := reasonsOf(got); r != tc.reasons {
				t.Fatalf("reasons %q, want %q", r, tc.reasons)
			}
		})
	}
}
