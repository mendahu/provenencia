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
	}{
		// Exact
		{"exact", "single name",
			[]Candidate{nm(1, "given=James|surname=Robins")},
			[]cl{{[]byte{1}, "given=James|surname=Robins"}}, StateSingle},
		{"exact", "identical names merge",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged},
		{"exact", "case, punctuation and spacing are ignored; best rank's spelling kept",
			[]Candidate{nm(1, "given=JAMES|surname=Robins."), nm(2, "given=james|surname=  robins")},
			[]cl{{[]byte{1, 2}, "given=JAMES|surname=Robins."}}, StateMerged},
		{"exact", "a hyphenated part is one unit",
			[]Candidate{nm(1, "given=Ann|surname=Smith-Jones"), nm(2, "given=Ann|surname=Smith Jones")},
			[]cl{{[]byte{1, 2}, "given=Ann|surname=Smith-Jones"}}, StateMerged},
		{"exact", "one hyphenated part is not two parts",
			[]Candidate{nm(1, "given=Ann|surname=Smith-Jones"), nm(2, "given=Ann|surname=Smith|surname=Jones")},
			[]cl{{[]byte{1}, "given=Ann|surname=Smith-Jones"}, {[]byte{2}, "given=Ann|surname=Smith|surname=Jones"}}, StateMixed},
		{"exact", "parts of one type compare in idx order, wherever they sit",
			[]Candidate{nm(1, "given=James|surname=Robins|given=Kenneth"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins|given=Kenneth"}}, StateMerged},
		{"exact", "an initial with and without its period",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=J|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=J.|surname=Robins"}}, StateMerged},

		// Subsumption
		{"subsumption", "an initial expands: [J] into [James]",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged},
		{"subsumption", "the expansion does not depend on rank",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=J.|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged},
		{"subsumption", "[James] into [James, Kenneth]",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged},
		{"subsumption", "[J, K] into [James, Kenneth]",
			[]Candidate{nm(1, "given=J|given=K|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged},
		{"subsumption", "[James, K] into [James, Kenneth]",
			[]Candidate{nm(1, "given=James|given=K.|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged},
		{"subsumption", "a later part alone: [Kenneth] into [James, Kenneth]",
			[]Candidate{nm(1, "given=Kenneth|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged},
		{"subsumption", "a chain folds to the fullest: [J], [James], [James, Kenneth]",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1, 2, 3}, "given=James|given=Kenneth|surname=Robins"}}, StateMerged},
		{"subsumption", "order matters: [Kenneth, James] is not [James, Kenneth]",
			[]Candidate{nm(1, "given=Kenneth|given=James|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1}, "given=Kenneth|given=James|surname=Robins"}, {[]byte{2}, "given=James|given=Kenneth|surname=Robins"}}, StateMixed},
		{"subsumption", "different middle names do not fold",
			[]Candidate{nm(1, "given=James|given=Kevin|surname=Robins"), nm(2, "given=James|given=Kenneth|surname=Robins")},
			[]cl{{[]byte{1}, "given=James|given=Kevin|surname=Robins"}, {[]byte{2}, "given=James|given=Kenneth|surname=Robins"}}, StateMixed},
		{"subsumption", "an initial of another letter does not fold",
			[]Candidate{nm(1, "given=W.|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1}, "given=W.|surname=Robins"}, {[]byte{2}, "given=James|surname=Robins"}}, StateMixed},
		{"subsumption", "two initials of different letters stay apart",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=K.|surname=Robins")},
			[]cl{{[]byte{1}, "given=J.|surname=Robins"}, {[]byte{2}, "given=K.|surname=Robins"}}, StateMixed},
		{"subsumption", "a lone uncased character is a whole name, not an initial",
			[]Candidate{nm(1, "surname=蒋|given=浩"), nm(2, "surname=蒋|given=浩然")},
			[]cl{{[]byte{1}, "surname=蒋|given=浩"}, {[]byte{2}, "surname=蒋|given=浩然"}}, StateMixed},
		{"subsumption", "an ambiguous initial folds into the best ranked of equal support",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=John|surname=Robins"), nm(3, "given=J.|surname=Robins")},
			[]cl{{[]byte{1, 3}, "given=James|surname=Robins"}, {[]byte{2}, "given=John|surname=Robins"}}, StateMixed},
		{"subsumption", "an ambiguous initial folds into the best supported",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=John|surname=Robins"), nm(3, "given=John|surname=Robins"), nm(4, "given=J.|surname=Robins")},
			[]cl{{[]byte{2, 3, 4}, "given=John|surname=Robins"}, {[]byte{1}, "given=James|surname=Robins"}}, StateMixed},
		{"subsumption", "an initial never folds across types",
			[]Candidate{nm(1, "nick=J|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|nick=J|surname=Robins"}}, StateMerged},

		// Majority
		{"majority", "two of three crowd out a misspelling",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|surname=Robbins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}, {[]byte{3}, "given=James|surname=Robbins"}}, StateMixed},
		{"majority", "rank does not save an outlier",
			[]Candidate{nm(1, "given=James|surname=Robbins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|surname=Robins")},
			[]cl{{[]byte{2, 3}, "given=James|surname=Robins"}, {[]byte{1}, "given=James|surname=Robbins"}}, StateMixed},
		{"majority", "four of five",
			[]Candidate{nm(1, "surname=Robbins"), nm(2, "surname=Robins"), nm(3, "surname=Robins"), nm(4, "surname=Robins"), nm(5, "surname=Robins")},
			[]cl{{[]byte{2, 3, 4, 5}, "surname=Robins"}, {[]byte{1}, "surname=Robbins"}}, StateMixed},
		{"majority", "one to one keeps both",
			[]Candidate{nm(1, "given=Mary|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1}, "given=Mary|surname=Robins"}, {[]byte{2}, "given=James|surname=Robins"}}, StateMixed},
		{"majority", "two of four is not a majority",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=Mary|surname=Robins"), nm(4, "given=Ann|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}, {[]byte{3}, "given=Mary|surname=Robins"}, {[]byte{4}, "given=Ann|surname=Robins"}}, StateMixed},
		{"majority", "eliminated candidates group by exact parts",
			[]Candidate{nm(1, "surname=Robins"), nm(2, "surname=Robins"), nm(3, "surname=Robins"), nm(4, "surname=Robbins"), nm(5, "surname=ROBBINS")},
			[]cl{{[]byte{1, 2, 3}, "surname=Robins"}, {[]byte{4, 5}, "surname=Robbins"}}, StateMixed},
		{"majority", "each type is elected on its own",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robbins"), nm(3, "given=Jim|surname=Robins")},
			[]cl{{[]byte{1}, "given=James|surname=Robins"}, {[]byte{2}, "given=James|surname=Robbins"}, {[]byte{3}, "given=Jim|surname=Robins"}}, StateMixed},
		{"majority", "folded support counts, and the winner's parts can come from a loser",
			[]Candidate{nm(1, "given=J.|surname=Robins"), nm(2, "given=J.|surname=Robins"), nm(3, "given=James|surname=Robbins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}, {[]byte{3}, "given=James|surname=Robbins"}}, StateMixed},

		// No type is special
		{"types", "prefixes that differ stay apart",
			[]Candidate{nm(1, "prefix=Rev.|given=James|surname=Robins"), nm(2, "prefix=Dr.|given=James|surname=Robins")},
			[]cl{{[]byte{1}, "prefix=Rev.|given=James|surname=Robins"}, {[]byte{2}, "prefix=Dr.|given=James|surname=Robins"}}, StateMixed},
		{"types", "a prefix is elected like any part",
			[]Candidate{nm(1, "prefix=Rev.|given=James|surname=Robins"), nm(2, "prefix=Rev.|given=James|surname=Robins"), nm(3, "prefix=Dr.|given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "prefix=Rev.|given=James|surname=Robins"}, {[]byte{3}, "prefix=Dr.|given=James|surname=Robins"}}, StateMixed},
		{"types", "suffixes that differ stay apart",
			[]Candidate{nm(1, "given=James|surname=Robins|suffix=Jr."), nm(2, "given=James|surname=Robins|suffix=Sr.")},
			[]cl{{[]byte{1}, "given=James|surname=Robins|suffix=Jr."}, {[]byte{2}, "given=James|surname=Robins|suffix=Sr."}}, StateMixed},
		{"types", "James and Jim, both given, are different names",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=Jim|surname=Robins")},
			[]cl{{[]byte{1}, "given=James|surname=Robins"}, {[]byte{2}, "given=Jim|surname=Robins"}}, StateMixed},
		{"types", "a nickname beside a given name is kept",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "given=James|nick=Jim|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|nick=Jim|surname=Robins"}}, StateMerged},
		{"types", "made-up type keys work the same",
			[]Candidate{nm(1, "a=James|b=Robins"), nm(2, "a=James|b=Robins"), nm(3, "a=James|b=Robbins")},
			[]cl{{[]byte{1, 2}, "a=James|b=Robins"}, {[]byte{3}, "a=James|b=Robbins"}}, StateMixed},
		{"types", "untyped parts are a type like any other",
			[]Candidate{nm(1, "=J|=Robins"), nm(2, "=James|=Robins")},
			[]cl{{[]byte{1, 2}, "=James|=Robins"}}, StateMerged},
		{"types", "untyped parts do not meet typed ones",
			[]Candidate{nm(1, "=James|=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "=James|=Robins|given=James|surname=Robins"}}, StateMerged},
		{"types", "undetermined is a type like any other",
			[]Candidate{nm(1, "undetermined=Kendall|surname=Robins"), nm(2, "undetermined=Kendal|surname=Robins")},
			[]cl{{[]byte{1}, "undetermined=Kendall|surname=Robins"}, {[]byte{2}, "undetermined=Kendal|surname=Robins"}}, StateMixed},
		{"types", "accents are different names",
			[]Candidate{nm(1, "given=José|surname=Silva"), nm(2, "given=Jose|surname=Silva")},
			[]cl{{[]byte{1}, "given=José|surname=Silva"}, {[]byte{2}, "given=Jose|surname=Silva"}}, StateMixed},

		// Absence is not disagreement
		{"absence", "a surname alone joins the fuller name",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged},
		{"absence", "a given name alone joins",
			[]Candidate{nm(1, "given=James"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged},
		{"absence", "a surname alone joins the best-ranked of equal clusters",
			[]Candidate{nm(1, "given=Mary|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "surname=Robins")},
			[]cl{{[]byte{1, 3}, "given=Mary|surname=Robins"}, {[]byte{2}, "given=James|surname=Robins"}}, StateMixed},
		{"absence", "a fuller given name wins two of three, and the surname joins it",
			[]Candidate{nm(1, "given=Mary|surname=Robins"), nm(2, "given=James|surname=Robins"), nm(3, "given=James|given=K.|surname=Robins"), nm(4, "surname=Robins")},
			[]cl{{[]byte{2, 3, 4}, "given=James|given=K.|surname=Robins"}, {[]byte{1}, "given=Mary|surname=Robins"}}, StateMixed},
		{"absence", "a candidate sharing no type joins",
			[]Candidate{nm(1, "given=James|surname=Robins"), nm(2, "prefix=Rev.")},
			[]cl{{[]byte{1, 2}, "prefix=Rev.|given=James|surname=Robins"}}, StateMerged},

		// Form is a transcription
		{"form", "a name with no parts is dropped",
			[]Candidate{name(1, "James Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{2}, "given=James|surname=Robins"}}, StateSingle},
		{"form", "only names with no parts: empty",
			[]Candidate{name(1, "James Robins"), name(2, "James Robins")},
			nil, StateEmpty},
		{"form", "parts with nothing in them are no parts",
			[]Candidate{nm(1, "given=.|surname= "), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{2}, "given=James|surname=Robins"}}, StateSingle},
		{"form", "an empty part is skipped",
			[]Candidate{nm(1, "given=.|surname=Robins"), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged},
		{"form", "the same parts under different forms merge",
			[]Candidate{withNameForm(nm(1, "given=James|surname=Robins"), "Robins, James"), withNameForm(nm(2, "given=James|surname=Robins"), "Jas. Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins"}}, StateMerged},
		{"form", "the same form over different parts does not merge",
			[]Candidate{withNameForm(nm(1, "given=James|surname=Robins"), "J. Robins"), withNameForm(nm(2, "given=John|surname=Robins"), "J. Robins")},
			[]cl{{[]byte{1}, "given=James|surname=Robins"}, {[]byte{2}, "given=John|surname=Robins"}}, StateMixed},

		// Assembly
		{"assembly", "a type the base lacks goes after its predecessor",
			[]Candidate{nm(1, "prefix=Rev.|given=James|surname=Robins"), nm(2, "given=James|nick=Jim|surname=Robins")},
			[]cl{{[]byte{1, 2}, "prefix=Rev.|given=James|nick=Jim|surname=Robins"}}, StateMerged},
		{"assembly", "the base is the member with the most types",
			[]Candidate{nm(1, "surname=Robins|given=James"), nm(2, "given=James|surname=Robins|suffix=Jr.")},
			[]cl{{[]byte{1, 2}, "given=James|surname=Robins|suffix=Jr."}}, StateMerged},
		{"assembly", "each type's spelling comes from its best-ranked carrier",
			[]Candidate{nm(1, "given=J.|surname=ROBINS|suffix=Jr."), nm(2, "given=James|surname=Robins")},
			[]cl{{[]byte{1, 2}, "given=James|surname=ROBINS|suffix=Jr."}}, StateMerged},
		{"assembly", "a surname-first name keeps its order",
			[]Candidate{nm(1, "surname=蒋|given=浩"), nm(2, "surname=蒋")},
			[]cl{{[]byte{1, 2}, "surname=蒋|given=浩"}}, StateMerged},
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
			return Candidate{ObservationID: id(n), Value: Value{Name: v}}
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

			// Every structured candidate in exactly one cluster; no other.
			seen := map[byte]int{}
			for _, c := range got.Clusters {
				for _, oid := range c.ObservationIDs {
					seen[oid[3]]++
				}
			}
			for _, c := range in {
				want := 1
				if structuredName(c, 0) == nil {
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
				renamed[i] = Candidate{ObservationID: c.ObservationID, Value: Value{Name: &n}}
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
		out = append(out, fmt.Sprintf("%d:%q", c.ObservationID[3], spec(c.Value.Name)))
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

func signatureOf(n *namevalues.Value) string {
	nc := structuredName(Candidate{Value: Value{Name: n}}, 0)
	if nc == nil {
		return ""
	}
	return exactSignature(nc.values)
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
