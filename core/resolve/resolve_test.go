package resolve

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

func id(n byte) []byte { return []byte{0, 0, 0, n} }

func text(n byte, s string) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{Text: s, HasText: true}}
}

func name(n byte, form string) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{Name: &namevalues.Value{Form: form}}}
}

func ip(v int) *int { return &v }

func date(n byte, y int, m, d *int) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{Date: &datevalues.Value{
		Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: ip(y), StartMonth: m, StartDay: d,
	}}}
}

// shape is a compact view of a Result: each cluster's member ids.
func shape(r Result) [][]byte {
	var out [][]byte
	for _, c := range r.Clusters {
		var ids []byte
		for _, oid := range c.ObservationIDs {
			ids = append(ids, oid[3])
		}
		if c.Support > len(c.ObservationIDs) {
			panic("support counts more Sources than members")
		}
		out = append(out, ids)
	}
	return out
}

func TestResolve(t *testing.T) {
	cases := []struct {
		name       string
		valueType  string
		candidates []Candidate
		concluded  *Value
		want       [][]byte
		state      State
	}{
		{"empty", properties.ValueTypeText, nil, nil, nil, StateEmpty},
		{"single text", properties.ValueTypeText,
			[]Candidate{text(1, "York")}, nil, [][]byte{{1}}, StateSingle},
		{"identical text merges", properties.ValueTypeText,
			[]Candidate{text(3, "York"), text(1, "York"), text(2, " York ")}, nil, [][]byte{{1, 2, 3}}, StateMerged},
		{"text ignores case (S9-13)", properties.ValueTypeText,
			[]Candidate{text(1, "York"), text(2, "york")}, nil, [][]byte{{1, 2}}, StateMerged},
		{"names merge by normalized form", properties.ValueTypeName,
			[]Candidate{name(1, "JAMES ROBINS"), name(2, "James  Robins."), name(3, "james robins")}, nil,
			[][]byte{{1, 2, 3}}, StateMerged},
		{"different names stay apart", properties.ValueTypeName,
			[]Candidate{name(1, "James Robins"), name(2, "Jim Robins")}, nil, [][]byte{{1}, {2}}, StateMixed},
		{"tie broken by lowest id", properties.ValueTypeText,
			[]Candidate{text(5, "B"), text(2, "A")}, nil, [][]byte{{2}, {5}}, StateMixed},
		{"support beats id; two of three outvote the third (S9-13)", properties.ValueTypeText,
			[]Candidate{text(1, "A"), text(4, "B"), text(3, "B")}, nil, [][]byte{{3, 4}, {1}}, StateMerged},
		{"integers", properties.ValueTypeInteger,
			[]Candidate{
				{ObservationID: id(1), Value: Value{Integer: 7, HasInteger: true}},
				{ObservationID: id(2), Value: Value{Integer: 7, HasInteger: true}},
				{ObservationID: id(3), Value: Value{Integer: 0, HasInteger: true}},
			}, nil, [][]byte{{1, 2}, {3}}, StateMerged},
		{"terms by id", properties.ValueTypeTerm,
			[]Candidate{
				{ObservationID: id(1), Value: Value{TermID: []byte("birth")}},
				{ObservationID: id(2), Value: Value{TermID: []byte("birth")}},
			}, nil, [][]byte{{1, 2}}, StateMerged},
		{"subjects by id", properties.ValueTypeSubject,
			[]Candidate{
				{ObservationID: id(1), Value: Value{SubjectID: []byte("s1")}},
				{ObservationID: id(2), Value: Value{SubjectID: []byte("s2")}},
			}, nil, [][]byte{{1}, {2}}, StateMixed},
		{"equal dates merge", properties.ValueTypeDate,
			[]Candidate{date(1, 1985, ip(5), ip(14)), date(2, 1985, ip(5), ip(14))}, nil,
			[][]byte{{1, 2}}, StateMerged},
		{"contained dates stay apart in v1", properties.ValueTypeDate,
			[]Candidate{date(1, 1985, ip(5), nil), date(2, 1985, ip(5), ip(14))}, nil,
			[][]byte{{1}, {2}}, StateMixed},
		{"concluded joins its cluster at rank 1", properties.ValueTypeText,
			[]Candidate{text(1, "A"), text(2, "A"), text(3, "B")}, &Value{Text: "B", HasText: true},
			[][]byte{{3}, {1, 2}}, StateConcluded},
		{"concluded with no support leads", properties.ValueTypeText,
			[]Candidate{text(1, "A")}, &Value{Text: "C", HasText: true},
			[][]byte{nil, {1}}, StateConcluded},
		{"concluded with no candidates", properties.ValueTypeText,
			nil, &Value{Text: "C", HasText: true}, [][]byte{nil}, StateConcluded},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Resolve(tc.valueType, tc.candidates, tc.concluded)
			if err != nil {
				t.Fatal(err)
			}
			if s := shape(got); !reflect.DeepEqual(s, tc.want) {
				t.Fatalf("clusters %v, want %v", s, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
		})
	}
}

func TestResolveRepresentative(t *testing.T) {
	got, err := Resolve(properties.ValueTypeName,
		[]Candidate{name(9, "james robins"), name(4, "James Robins")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if f := got.Clusters[0].Value.Name.Form; f != "James Robins" {
		t.Fatalf("representative %q, want the lowest-id member's", f)
	}

	concluded := &Value{Name: &namevalues.Value{Form: "JAMES ROBINS"}}
	got, err = Resolve(properties.ValueTypeName,
		[]Candidate{name(9, "james robins"), name(4, "James Robins")}, concluded)
	if err != nil {
		t.Fatal(err)
	}
	if got.Clusters[0].Value.Name != concluded.Name || got.Clusters[0].Support != 2 {
		t.Fatalf("concluded cluster %+v", got.Clusters[0])
	}
}

func TestResolveOrderIndependent(t *testing.T) {
	in := []Candidate{text(1, "A"), text(2, "B"), text(3, "B"), text(4, "C"), text(5, "A"), text(6, "D")}
	want, err := Resolve(properties.ValueTypeText, in, nil)
	if err != nil {
		t.Fatal(err)
	}
	perms := [][]int{{5, 4, 3, 2, 1, 0}, {2, 0, 5, 1, 4, 3}, {3, 1, 4, 0, 5, 2}}
	for _, p := range perms {
		shuffled := make([]Candidate, len(in))
		for i, j := range p {
			shuffled[i] = in[j]
		}
		got, err := Resolve(properties.ValueTypeText, shuffled, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("order %v: %v, want %v", p, shape(got), shape(want))
		}
	}
}

func TestResolveDoesNotMutateInput(t *testing.T) {
	in := []Candidate{text(3, "A"), text(1, "B")}
	if _, err := Resolve(properties.ValueTypeText, in, nil); err != nil {
		t.Fatal(err)
	}
	if in[0].ObservationID[3] != 3 {
		t.Fatal("input reordered")
	}
}

func TestResolveErrors(t *testing.T) {
	if _, err := Resolve("colour", nil, nil); !errors.Is(err, ErrUnknownValueType) {
		t.Fatalf("unknown type: %v", err)
	}
	if _, err := Resolve(properties.ValueTypeName, []Candidate{text(1, "A")}, nil); !errors.Is(err, ErrValueMismatch) {
		t.Fatalf("mismatched candidate: %v", err)
	}
	if _, err := Resolve(properties.ValueTypeText, nil, &Value{Integer: 1, HasInteger: true}); !errors.Is(err, ErrValueMismatch) {
		t.Fatalf("mismatched concluded: %v", err)
	}
}

func TestNormalizeForm(t *testing.T) {
	for in, want := range map[string]string{
		"  James   Robins ": "james robins",
		"J. Robins":         "j robins",
		"O'Brien, Mary":     "obrien mary",
		"ÉMILE\tZOLA":       "émile zola",
		"Smith-Jones":       "smith jones",
		"Mary–Ann / Polly":  "mary ann polly",
		"-Robins-":          "robins",
		"":                  "",
	} {
		if got := NormalizeForm(in); got != want {
			t.Errorf("NormalizeForm(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSortKey(t *testing.T) {
	k, ok := SortKey(properties.ValueTypeName, Value{Name: &namevalues.Value{Form: "  O'Brien,  Mary "}})
	if !ok || k != "obrien mary" {
		t.Fatalf("name %q %v", k, ok)
	}
	k, ok = SortKey(properties.ValueTypeText, Value{Text: " Upper Canada ", HasText: true})
	if !ok || k != "upper canada" {
		t.Fatalf("text %q %v", k, ok)
	}
	ints := []int64{-1 << 63, -100, -1, 0, 1, 9, 10, 1<<63 - 1}
	prev := ""
	for i, n := range ints {
		k, ok := SortKey(properties.ValueTypeInteger, Value{Integer: n, HasInteger: true})
		if !ok || (i > 0 && k <= prev) {
			t.Fatalf("integer %d key %q not after %q", n, k, prev)
		}
		prev = k
	}
	for _, vt := range []string{properties.ValueTypeDate, properties.ValueTypeTerm, properties.ValueTypeSubject} {
		if _, ok := SortKey(vt, Value{TermID: []byte("x"), SubjectID: []byte("x"), Date: &datevalues.Value{}}); ok {
			t.Fatalf("%s has no sort key yet", vt)
		}
	}
}
