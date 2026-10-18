package autoreconcile

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/depictionvalue"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func term(n byte, termID, key string) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{TermID: []byte(termID), TermKey: key}}
}

// depiction is a likeness holding image; an empty image is one whose File
// depictionvalue.Readable refused.
func depiction(n byte, image string) Candidate {
	return Candidate{ObservationID: id(n), Value: Value{Depiction: &depictionvalue.Value{ImageFileID: []byte(image)}}}
}

// neutralTerm is a term the catalog flags as no evidence ("unknown").
func neutralTerm(n byte, termID, key string) Candidate {
	c := term(n, termID, key)
	c.Value.Neutral = true
	return c
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
		want         [][]byte // displayed and hidden values, by member ids
		state        State
		displayed    string // the first value's text, when checked
	}{
		{"text", "trimmed and case-insensitive", propertymeta.ValueTypeText,
			[]Candidate{text(1, "  York "), text(2, "YORK"), text(3, "york")}, [][]byte{{1, 2, 3}}, StateMerged, "  York "},
		{"text", "non-ASCII case folds", propertymeta.ValueTypeText,
			[]Candidate{text(1, "ÉMILE"), text(2, "émile")}, [][]byte{{1, 2}}, StateMerged, "ÉMILE"},
		{"text", "inner spacing is ignored; U.C. is the word uc", propertymeta.ValueTypeText,
			[]Candidate{text(1, "Upper Canada"), text(2, "Upper  Canada"), text(3, "U.C.")}, [][]byte{{1, 2}, {3}}, StateMerged, ""},
		{"text", "punctuation and spacing between words are ignored", propertymeta.ValueTypeText,
			[]Candidate{text(1, "Sainte-Anne"), text(2, "Sainte Anne")}, [][]byte{{1, 2}}, StateMerged, "Sainte-Anne"},
		{"text", "word breaks are ignored", propertymeta.ValueTypeText,
			[]Candidate{text(1, "Yorktown"), text(2, "York Town")}, [][]byte{{1, 2}}, StateMerged, "Yorktown"},
		{"text", "an accent fold is a different value", propertymeta.ValueTypeText,
			[]Candidate{text(1, "Montréal"), text(2, "Montreal")}, [][]byte{{1}, {2}}, StateMixed, ""},
		{"text", "text never folds", propertymeta.ValueTypeText,
			[]Candidate{text(1, "York"), text(2, "York, Upper Canada")}, [][]byte{{1}, {2}}, StateMixed, ""},
		{"text", "the best-ranked member's spelling is displayed", propertymeta.ValueTypeText,
			[]Candidate{text(1, "york"), highTrust(text(2, "York"))}, [][]byte{{1, 2}}, StateMerged, "York"},
		{"text", "blank is no evidence", propertymeta.ValueTypeText,
			[]Candidate{text(1, ""), text(2, "  ")}, nil, StateEmpty, ""},
		{"integer", "equal values", propertymeta.ValueTypeInteger,
			[]Candidate{integer(1, 34), integer(2, 34), integer(3, 0)}, [][]byte{{1, 2}, {3}}, StateMerged, ""},
		{"integer", "negative and zero are values", propertymeta.ValueTypeInteger,
			[]Candidate{integer(1, -1), integer(2, 0)}, [][]byte{{1}, {2}}, StateMixed, ""},
		{"term", "the same term", propertymeta.ValueTypeTerm,
			[]Candidate{term(1, "m", "male"), term(2, "m", "male")}, [][]byte{{1, 2}}, StateMerged, ""},
		{"term", "a neutral term is no evidence", propertymeta.ValueTypeTerm,
			[]Candidate{term(1, "m", "male"), neutralTerm(2, "u", "unknown"), neutralTerm(3, "i", "indeterminate")}, [][]byte{{1}}, StateSingle, ""},
		{"term", "only neutral terms: empty", propertymeta.ValueTypeTerm,
			[]Candidate{neutralTerm(1, "u", "unknown")}, nil, StateEmpty, ""},
		{"term", "a term with no key still counts", propertymeta.ValueTypeTerm,
			[]Candidate{term(1, "f", ""), term(2, "m", "")}, [][]byte{{1}, {2}}, StateMixed, ""},
		{"name", "a name with no parts is no evidence, whatever its form", propertymeta.ValueTypeName,
			[]Candidate{name(1, "James Robins"), nm(2, "given=James|surname=Robins")}, [][]byte{{2}}, StateSingle, ""},
		{"depiction", "the same image is one value", propertymeta.ValueTypeDepiction,
			[]Candidate{depiction(1, "a"), depiction(2, "a")}, [][]byte{{1, 2}}, StateMerged, ""},
		{"depiction", "two images", propertymeta.ValueTypeDepiction,
			[]Candidate{depiction(1, "a"), depiction(2, "b")}, [][]byte{{1}, {2}}, StateMixed, ""},
		{"depiction", "an unreadable image is no evidence", propertymeta.ValueTypeDepiction,
			[]Candidate{depiction(1, "")}, nil, StateEmpty, ""},
	}
	for _, tc := range cases {
		t.Run(tc.module+"/"+tc.name, func(t *testing.T) {
			got, err := Reconcile(testRules, tc.valueType, tc.in, nil, propertymeta.CardinalitySingle)
			if err != nil {
				t.Fatal(err)
			}
			if s := shape(got); !reflect.DeepEqual(s, tc.want) {
				t.Fatalf("values %v, want %v", s, tc.want)
			}
			if got.State() != tc.state {
				t.Fatalf("state %q, want %q", got.State(), tc.state)
			}
			if tc.displayed != "" && got.Values[0].Value.Text != tc.displayed {
				t.Fatalf("displayed %q, want %q", got.Values[0].Value.Text, tc.displayed)
			}
		})
	}
}

// Every value type has a module, and mismatched values are refused.
func TestModuleFor(t *testing.T) {
	for _, vt := range []string{propertymeta.ValueTypeText, propertymeta.ValueTypeInteger, propertymeta.ValueTypeTerm,
		propertymeta.ValueTypeSubject, propertymeta.ValueTypeName, propertymeta.ValueTypeDate, propertymeta.ValueTypeDepiction} {
		if moduleFor(testRules, vt) == nil {
			t.Errorf("%s has no module", vt)
		}
	}
	if _, err := Reconcile(testRules, propertymeta.ValueTypeName, []Candidate{{ObservationID: id(1), Value: Value{Name: &namevalue.Value{}}}}, nil, propertymeta.CardinalitySingle); err != nil {
		t.Fatalf("an empty name is no evidence, not an error: %v", err)
	}
}

// A person's likenesses are a gallery: every distinct image stays.
func TestReconcileDepictionMultiple(t *testing.T) {
	t.Run("distinct images are distinct values", func(t *testing.T) {
		got, err := Reconcile(testRules, propertymeta.ValueTypeDepiction,
			[]Candidate{depiction(1, "a"), depiction(2, "a"), depiction(3, "b")}, nil, propertymeta.CardinalityMultiple)
		if err != nil {
			t.Fatal(err)
		}
		if s := shape(got); !reflect.DeepEqual(s, [][]byte{{1, 2}, {3}}) {
			t.Fatalf("values %v", s)
		}
		for _, v := range got.Values {
			if v.Reason != ReasonKept {
				t.Fatalf("reason %q, want kept", v.Reason)
			}
		}
		if got.State() != StateMultiple {
			t.Fatalf("state %q", got.State())
		}
	})
	t.Run("no depiction is a mismatch", func(t *testing.T) {
		_, err := Reconcile(testRules, propertymeta.ValueTypeDepiction,
			[]Candidate{{ObservationID: id(1)}}, nil, propertymeta.CardinalityMultiple)
		if !errors.Is(err, ErrValueMismatch) {
			t.Fatalf("err %v", err)
		}
	})
	t.Run("an unreadable image is no evidence", func(t *testing.T) {
		got, err := Reconcile(testRules, propertymeta.ValueTypeDepiction,
			[]Candidate{depiction(1, "")}, nil, propertymeta.CardinalityMultiple)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Values) != 0 || got.Outcomes[0].Reason != ReasonNoEvidence {
			t.Fatalf("values %v, outcomes %+v", got.Values, got.Outcomes)
		}
	})
}

// Compatible is the pipeline's "same value or fold", one pair at a time.
func TestCompatible(t *testing.T) {
	cases := []struct {
		name      string
		valueType string
		a, b      Candidate
		want      bool
	}{
		{"text: trimmed and case-insensitive", propertymeta.ValueTypeText, text(1, " York"), text(2, "york"), true},
		{"text: different", propertymeta.ValueTypeText, text(1, "York"), text(2, "Toronto"), false},
		{"text: punctuation and spacing are ignored", propertymeta.ValueTypeText, text(1, "Sainte-Anne"), text(2, "Sainte Anne"), true},
		{"text: word breaks are ignored", propertymeta.ValueTypeText, text(1, "Yorktown"), text(2, "York Town"), true},
		{"text: an accent fold is a different value", propertymeta.ValueTypeText, text(1, "Montréal"), text(2, "Montreal"), false},
		{"text: never folds", propertymeta.ValueTypeText, text(1, "York"), text(2, "York, Upper Canada"), false},
		{"text: blank is no evidence", propertymeta.ValueTypeText, text(1, ""), text(2, ""), false},
		{"integer: equal", propertymeta.ValueTypeInteger, integer(1, 34), integer(2, 34), true},
		{"integer: different", propertymeta.ValueTypeInteger, integer(1, 34), integer(2, 35), false},
		{"term: the same term", propertymeta.ValueTypeTerm, term(1, "m", "male"), term(2, "m", "male"), true},
		{"term: different terms", propertymeta.ValueTypeTerm, term(1, "m", "male"), term(2, "f", "female"), false},
		{"term: a neutral term is no evidence", propertymeta.ValueTypeTerm, neutralTerm(1, "u", "unknown"), neutralTerm(2, "u", "unknown"), false},
		{"name: the same parts", propertymeta.ValueTypeName,
			nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robins"), true},
		{"name: an initial folds into the full given name", propertymeta.ValueTypeName,
			nm(1, "given=J.|surname=Robins"), nm(2, "given=James|surname=Robins"), true},
		{"name: fold either way round", propertymeta.ValueTypeName,
			nm(1, "given=James|surname=Robins"), nm(2, "given=J.|surname=Robins"), true},
		{"name: a part only one carries does not count against", propertymeta.ValueTypeName,
			nm(1, "given=James"), nm(2, "given=James|surname=Robins"), true},
		{"name: a spelling variant differs", propertymeta.ValueTypeName,
			nm(1, "given=James|surname=Robins"), nm(2, "given=James|surname=Robbins"), false},
		{"name: no shared part", propertymeta.ValueTypeName,
			nm(1, "given=James"), nm(2, "surname=Robins"), false},
		{"name: no parts is no evidence", propertymeta.ValueTypeName, name(1, "James Robins"), name(2, "James Robins"), false},
		{"depiction: the same image", propertymeta.ValueTypeDepiction, depiction(1, "a"), depiction(2, "a"), true},
		{"depiction: different images", propertymeta.ValueTypeDepiction, depiction(1, "a"), depiction(2, "b"), false},
		{"depiction: an unreadable image is no evidence", propertymeta.ValueTypeDepiction, depiction(1, ""), depiction(2, ""), false},
		{"mismatched value", propertymeta.ValueTypeInteger, text(1, "34"), integer(2, 34), false},
		{"unknown value type", "colour", text(1, "red"), text(2, "red"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Compatible(testRules, tc.valueType, tc.a.Value, tc.b.Value); got != tc.want {
				t.Fatalf("Compatible = %v, want %v", got, tc.want)
			}
		})
	}
}

// An alias never makes two names compatible: value compare finds the longer
// name contains the shorter, but only as an alias, which does not fold.
func TestCompatibleIgnoresAliases(t *testing.T) {
	rules := matchregistry.Default().Value
	rules.Names.Aliases = valuecompare.NewAliasIndex([]valuecompare.AliasTable{{
		Key: "t", Kind: valuecompare.AliasNickname,
		Roles: []valuecompare.NameRole{valuecompare.RoleGiven, valuecompare.RoleNick, valuecompare.RoleUntyped},
		Sets:  [][]string{{"jim", "james"}},
	}}, rules.Words)
	a := nm(1, "given=Jim|surname=Robins").Value
	b := nm(2, "given=James|given=Kenneth|surname=Robins").Value
	if Compatible(rules, propertymeta.ValueTypeName, a, b) {
		t.Fatal("Jim Robins is compatible with James Kenneth Robins through an alias")
	}
	res, _ := valuecompare.Compare(
		valuecompare.Value{Type: propertymeta.ValueTypeName, Name: a.Name},
		valuecompare.Value{Type: propertymeta.ValueTypeName, Name: b.Name}, rules)
	if res.Relation != valuecompare.Contains {
		t.Fatalf("whole names relate as %q, want contains", res.Relation)
	}
}
