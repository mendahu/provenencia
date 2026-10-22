package promotealign_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// T-6: a sheet value carries its catalog type and, for a term, the label the
// pool rule finds, so the client words it.
func TestEvidenceValues(t *testing.T) {
	f := newFixture(t)
	vocab, err := f.c.Graph().Vocabulary()
	must(t, err)
	label := func(property, term string) string {
		for _, x := range vocab.Terms {
			if x.PropertyKey == property && x.Key == term {
				return x.Label
			}
		}
		t.Fatalf("no %s term %s", property, term)
		return ""
	}
	values, err := promotealign.NewEvidenceValues(f.c.Graph())
	must(t, err)

	val := func(v valuecompare.Value) propertycompare.Value { return propertycompare.Value{Value: v} }
	term := func(key string) propertycompare.Value {
		return val(valuecompare.Value{Type: propertymeta.ValueTypeTerm, Term: key})
	}
	section := func(key vocabkey.PropertyKey) claimevidence.Section { return claimevidence.Section{Key: key} }
	line := func(p vocabkey.PropertyKey, in propertycompare.Value) claimevidence.Line {
		return claimevidence.Line{Property: p, Incoming: in}
	}

	for _, tc := range []struct {
		name      string
		got       promotealign.EvidenceValue
		wantType  string
		wantLabel string
	}{
		{"a name", values.Incoming(vocabkey.Person, section("name"),
			line("name", val(valuecompare.Value{Type: propertymeta.ValueTypeName, Name: namevaluetest.Western("John Smith")}))),
			propertymeta.ValueTypeName, ""},
		{"a gender term under the sex at birth lead", values.Incoming(vocabkey.Person, section(vocabkey.PropertySexAtBirth),
			line(vocabkey.PropertyGender, term("man"))),
			propertymeta.ValueTypeTerm, label("gender", "man")},
		{"a term with no label", values.Incoming(vocabkey.Person, section(vocabkey.PropertySexAtBirth),
			line(vocabkey.PropertyGender, term("no_such_term"))),
			propertymeta.ValueTypeTerm, ""},
		{"a member term under a pooled lead", values.Member(vocabkey.Person, section(vocabkey.PropertySexAtBirth),
			line(vocabkey.PropertyGender, term("man")), claimevidence.Pin{Value: term("male")}),
			propertymeta.ValueTypeTerm, label("sex_at_birth", "male")},
		{"a date", values.Incoming(vocabkey.Event, section(vocabkey.PropertyDate),
			line(vocabkey.PropertyDate, val(valuecompare.Value{Type: propertymeta.ValueTypeDate, Date: datevaluetest.YMD(1817, 0, 0)}))),
			propertymeta.ValueTypeDate, ""},
		{"a start date counted under the date section", values.Incoming(vocabkey.Event, section(vocabkey.PropertyDate),
			line(vocabkey.PropertyStartDate, val(valuecompare.Value{Type: propertymeta.ValueTypeDate, Date: datevaluetest.YMD(1817, 0, 0)}))),
			propertymeta.ValueTypeDate, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.Type != tc.wantType || tc.got.TermLabel != tc.wantLabel {
				t.Errorf("got type %q label %q, want %q %q", tc.got.Type, tc.got.TermLabel, tc.wantType, tc.wantLabel)
			}
		})
	}
}

// ResolvePaths gives each likeness the path of its image File, and two images
// different paths.
func TestEvidenceValuesResolvePaths(t *testing.T) {
	l := newLikenessFixture(t, 90)
	e := l.evidence()
	values, err := promotealign.NewEvidenceValues(l.f.c.Graph())
	must(t, err)
	must(t, values.ResolvePaths(l.f.c, e))
	s := section(t, e.Sections, vocabkey.PropertyLikeness)
	paths := map[string]bool{}
	for _, line := range s.Lines {
		in := values.Incoming(vocabkey.Person, s, line)
		if in.Type != propertymeta.ValueTypeDepiction || in.ImageRelPath == "" {
			t.Fatalf("incoming = %+v", in)
		}
		for _, p := range line.Members {
			m := values.Member(vocabkey.Person, s, line, p)
			if m.ImageRelPath == "" {
				t.Fatalf("member = %+v", m)
			}
			paths[m.ImageRelPath] = true
		}
	}
	if len(paths) < 2 {
		t.Fatalf("member image paths = %v, want at least two distinct", paths)
	}
}
