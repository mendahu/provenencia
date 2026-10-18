package subjectvocab

import (
	"testing"

	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The registry's subject section names only seeded types and the properties
// the seed binds to them, and a pair joins properties of one value type.
func TestSubjectSectionMatchesSeed(t *testing.T) {
	types := map[string]bool{}
	for _, st := range seedTypes {
		types[st.Key] = true
	}
	valueType := map[string]string{}
	for _, p := range seedProperties {
		valueType[p.Key] = p.ValueType
	}
	bound := map[string]map[string]bool{}
	for _, b := range allBindings() {
		if bound[b.TypeKey] == nil {
			bound[b.TypeKey] = map[string]bool{}
		}
		bound[b.TypeKey][b.PropertyKey] = true
	}
	s := matchregistry.Default().Subject

	checkKind := func(kind vocabkey.SubjectType) {
		if !types[string(kind)] {
			t.Errorf("kind %q is not a seeded type", kind)
		}
	}
	checkBound := func(kind vocabkey.SubjectType, key vocabkey.PropertyKey) {
		if !bound[string(kind)][string(key)] {
			t.Errorf("%s.%s is not bound by the seed", kind, key)
		}
	}
	for kind, weights := range s.Weights {
		checkKind(kind)
		for key := range weights {
			checkBound(kind, key)
		}
	}
	for kind, pools := range s.Pools {
		checkKind(kind)
		for lead, pool := range pools {
			checkBound(kind, lead)
			for _, with := range pool.With {
				checkBound(kind, with)
				if valueType[string(with)] != valueType[string(lead)] {
					t.Errorf("%s.%s pools %s of value type %q, lead is %q", kind, lead, with, valueType[string(with)], valueType[string(lead)])
				}
			}
		}
	}
	for kind, periods := range s.Periods {
		checkKind(kind)
		for lead, p := range periods {
			for _, key := range []vocabkey.PropertyKey{lead, p.Start, p.End} {
				checkBound(kind, key)
				if valueType[string(key)] != propertymeta.ValueTypeDate {
					t.Errorf("%s.%s is in a period but has value type %q, not date", kind, key, valueType[string(key)])
				}
			}
		}
	}

	card := map[string]string{}
	for _, p := range seedProperties {
		c := p.Cardinality
		if c == "" {
			c = propertymeta.CardinalitySingle
		}
		card[p.Key] = c
	}
	for key, meta := range subjectcomparetest.Metas() {
		k := string(key)
		if _, ok := valueType[k]; !ok {
			t.Errorf("fixture meta %s is not a seeded Property", k)
			continue
		}
		if meta.ValueType != valueType[k] {
			t.Errorf("fixture meta %s: value type %q, seeded %q", k, meta.ValueType, valueType[k])
		}
		got := meta.Cardinality
		if got == "" {
			got = propertymeta.CardinalitySingle
		}
		if got != card[k] {
			t.Errorf("fixture meta %s: cardinality %q, seeded %q", k, got, card[k])
		}
	}
}
