package resolve

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/database/properties"
)

func term(n byte, termID, key string) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{TermID: []byte(termID), TermKey: key}}
}

func integer(n byte, v int64) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{Integer: v, HasInteger: true}}
}

// Each module's "same value" and "no evidence", through the pipeline.
func TestModules(t *testing.T) {
	cases := []struct {
		module, name string
		valueType    string
		in           []Candidate
		want         [][]byte // displayed and hidden clusters, by member ids
		state        State
		displayed    string // the first cluster's text, when checked
	}{
		{"text", "trimmed and case-insensitive", properties.ValueTypeText,
			[]Candidate{text(1, "  York "), text(2, "YORK"), text(3, "york")}, [][]byte{{1, 2, 3}}, StateMerged, "  York "},
		{"text", "non-ASCII case folds", properties.ValueTypeText,
			[]Candidate{text(1, "ÉMILE"), text(2, "émile")}, [][]byte{{1, 2}}, StateMerged, "ÉMILE"},
		{"text", "inner spacing and punctuation still matter", properties.ValueTypeText,
			[]Candidate{text(1, "Upper Canada"), text(2, "Upper  Canada"), text(3, "U.C.")}, [][]byte{{1}, {2}, {3}}, StateMixed, ""},
		{"text", "the best-ranked member's spelling is displayed", properties.ValueTypeText,
			[]Candidate{text(1, "york"), highTrust(text(2, "York"))}, [][]byte{{1, 2}}, StateMerged, "York"},
		{"text", "blank is no evidence", properties.ValueTypeText,
			[]Candidate{text(1, ""), text(2, "  ")}, nil, StateEmpty, ""},
		{"integer", "equal values", properties.ValueTypeInteger,
			[]Candidate{integer(1, 34), integer(2, 34), integer(3, 0)}, [][]byte{{1, 2}, {3}}, StateMerged, ""},
		{"integer", "negative and zero are values", properties.ValueTypeInteger,
			[]Candidate{integer(1, -1), integer(2, 0)}, [][]byte{{1}, {2}}, StateMixed, ""},
		{"term", "the same term", properties.ValueTypeTerm,
			[]Candidate{term(1, "m", "male"), term(2, "m", "male")}, [][]byte{{1, 2}}, StateMerged, ""},
		{"term", "a neutral term is no evidence", properties.ValueTypeTerm,
			[]Candidate{term(1, "m", "male"), term(2, "u", "unknown"), term(3, "i", "indeterminate")}, [][]byte{{1}}, StateSingle, ""},
		{"term", "only neutral terms: empty", properties.ValueTypeTerm,
			[]Candidate{term(1, "u", "unknown")}, nil, StateEmpty, ""},
		{"term", "a term with no key still counts", properties.ValueTypeTerm,
			[]Candidate{term(1, "f", ""), term(2, "m", "")}, [][]byte{{1}, {2}}, StateMixed, ""},
		{"name", "a name with no parts is no evidence, whatever its form", properties.ValueTypeName,
			[]Candidate{name(1, "James Robins"), nm(2, "given=James|surname=Robins")}, [][]byte{{2}}, StateSingle, ""},
	}
	for _, tc := range cases {
		t.Run(tc.module+"/"+tc.name, func(t *testing.T) {
			got, err := Resolve(tc.valueType, tc.in, nil)
			if err != nil {
				t.Fatal(err)
			}
			if s := shape(got); !reflect.DeepEqual(s, tc.want) {
				t.Fatalf("clusters %v, want %v", s, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
			if tc.displayed != "" && got.Clusters[0].Value.Text != tc.displayed {
				t.Fatalf("displayed %q, want %q", got.Clusters[0].Value.Text, tc.displayed)
			}
		})
	}
}

// Every value type has a module, and mismatched values are refused.
func TestModuleFor(t *testing.T) {
	for _, vt := range []string{properties.ValueTypeText, properties.ValueTypeInteger, properties.ValueTypeTerm,
		properties.ValueTypeSubject, properties.ValueTypeName, properties.ValueTypeDate} {
		if moduleFor(vt) == nil {
			t.Errorf("%s has no module", vt)
		}
	}
	if _, err := Resolve(properties.ValueTypeName, []Candidate{{ObservationID: id(1), Value: Value{Name: &namevalues.Value{}}}}, nil); err != nil {
		t.Fatalf("an empty name is no evidence, not an error: %v", err)
	}
}

// The spelling-variant rule shared with core/match (SpellingFloor 0.8; one
// added or dropped letter from SpellingShortMin letters up).
func TestSpellingSimilarity(t *testing.T) {
	for _, tc := range []struct {
		x, y    string
		variant bool
	}{
		{"robins", "robbins", true},
		{"robnis", "robins", true}, // a swap is one edit
		{"smith", "smyth", true},
		{"catherine", "katherine", true},
		{"johann", "johan", true},
		{"macdonald", "mcdonald", true},
		{"ann", "anne", true}, // short word, one added letter
		{"jon", "john", true},
		{"dan", "dean", true}, // a known false positive of the short-word rule
		{"mary", "mark", false},
		{"josé", "jose", false}, // accents kept
		{"smith", "smythe", false},
		{"james", "jake", false},
		{"james", "jim", false},
		{"kenneth", "kevin", false},
		{"jr", "sr", false},
	} {
		got := SpellingSimilarity(tc.x, tc.y, SpellingFloor, SpellingShortMin) > 0
		if got != tc.variant {
			t.Errorf("%s ~ %s: variant %v, want %v", tc.x, tc.y, got, tc.variant)
		}
	}
	if SpellingSimilarity("robins", "robbins", 1, SpellingShortMin) != 0 {
		t.Error("a floor of 1 turns spelling variants off")
	}
}
