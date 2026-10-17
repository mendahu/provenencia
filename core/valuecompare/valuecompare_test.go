package valuecompare_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
)

func TestCompareUnsupported(t *testing.T) {
	want := []string{propertymeta.ValueTypeDate, propertymeta.ValueTypeInteger, propertymeta.ValueTypeTerm, propertymeta.ValueTypeText}
	if got := valuecompare.ComparerTypes(); !slices.Equal(got, want) {
		t.Errorf("ComparerTypes() = %v, want %v", got, want)
	}
	rules := matchregistry.Default().Value
	for _, typ := range []string{
		propertymeta.ValueTypeName,
		propertymeta.ValueTypeSubject, propertymeta.ValueTypeDepiction,
	} {
		v := valuecompare.Value{Type: typ, Text: "x"}
		res, det := valuecompare.Compare(v, v, rules)
		if res.Relation != valuecompare.Unknown || res.Similarity != 0 || det.Cause != valuecompare.CauseUnsupported {
			t.Errorf("%s: %+v %+v", typ, res, det)
		}
		if k, ok := valuecompare.Key(v, valuecompare.Same, rules); ok || k != "" {
			t.Errorf("%s: Key = %q, %v", typ, k, ok)
		}
	}
	mixed := []valuecompare.Value{
		{Type: propertymeta.ValueTypeText, Text: "male"},
		{Type: propertymeta.ValueTypeTerm, Term: "male"},
	}
	res, det := valuecompare.Compare(mixed[0], mixed[1], rules)
	if res.Relation != valuecompare.Unknown || det.Cause != valuecompare.CauseUnsupported {
		t.Errorf("text against term: %+v %+v", res, det)
	}
}

func TestCompareEchoesRefs(t *testing.T) {
	rules := matchregistry.Default().Value
	r1, r2 := []byte("r1"), []byte("r2")
	term := func(key string, ref []byte) valuecompare.Value {
		return valuecompare.Value{Type: propertymeta.ValueTypeTerm, Term: key, Ref: ref}
	}
	for name, pair := range map[string][2]valuecompare.Value{
		"same":      {term("male", r1), term("male", r2)},
		"different": {term("male", r1), term("female", r2)},
		"unknown":   {term("", r1), term("male", r2)},
	} {
		_, det := valuecompare.Compare(pair[0], pair[1], rules)
		if !bytes.Equal(det.RefA, r1) || !bytes.Equal(det.RefB, r2) {
			t.Errorf("%s: refs %q %q", name, det.RefA, det.RefB)
		}
	}
}
