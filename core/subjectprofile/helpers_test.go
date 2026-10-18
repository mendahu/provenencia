package subjectprofile_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func obs(n byte) []byte { return []byte{0, 0, 0, n} }
func src(n byte) []byte { return []byte{0, 1, 0, n} }
func mem(n byte) []byte { return []byte{0, 2, 0, n} }

func term(key string) autoreconcile.Value {
	return autoreconcile.Value{TermID: []byte(key), TermKey: key}
}

func text(s string) autoreconcile.Value {
	return autoreconcile.Value{Text: s, HasText: true}
}

func year(y int) autoreconcile.Value {
	return autoreconcile.Value{Date: datevaluetest.YMD(y, 0, 0)}
}

func name(form string) autoreconcile.Value {
	return autoreconcile.Value{Name: namevaluetest.Western(form)}
}

// cand is a candidate on member m, Observation o, under Source s.
func cand(m, o, s byte, v autoreconcile.Value) subjectprofile.Candidate {
	return subjectprofile.Candidate{
		Member: mem(m),
		Candidate: autoreconcile.Candidate{
			ObservationID: obs(o),
			SourceID:      src(s),
			Value:         v,
		},
	}
}

func weak(c subjectprofile.Candidate) subjectprofile.Candidate {
	c.Provenance.Credibility = -1
	return c
}

func provisional(c subjectprofile.Candidate) subjectprofile.Candidate {
	c.Provisional = true
	return c
}

func negative(c subjectprofile.Candidate) subjectprofile.Candidate {
	c.Negative = true
	return c
}

// reconcile runs the reconciler over cs the way the cache does for the
// property: the property's value type and cardinality, no concluded value.
func reconcile(t testing.TB, key vocabkey.PropertyKey, cs []subjectprofile.Candidate) []autoreconcile.ReconciledValue {
	t.Helper()
	meta := subjectcomparetest.Metas()[key]
	var in []autoreconcile.Candidate
	for _, c := range cs {
		in = append(in, c.Candidate)
	}
	res, err := autoreconcile.Reconcile(matchregistry.Default().Value, meta.ValueType, in, nil, meta.CardinalityOrSingle())
	if err != nil {
		t.Fatal(err)
	}
	return res.Values
}

// property is key's Property for cs, reconciled.
func property(t testing.TB, key vocabkey.PropertyKey, cs ...subjectprofile.Candidate) subjectprofile.Property {
	t.Helper()
	return subjectprofile.Property{Values: reconcile(t, key, cs), Candidates: cs}
}

// vcand is a bare Subject-side candidate for v, negative when neg.
func vcand(v autoreconcile.Value, neg bool) autoreconcile.Candidate {
	return autoreconcile.Candidate{Value: v, Negative: neg}
}

// maleFemale is an entity whose members 1 and 2 say male and 3 says female.
func maleFemale(t *testing.T) subjectprofile.Entity {
	return subjectprofile.Entity{Properties: props{sex: property(t, sex,
		cand(1, 1, 1, term("male")), cand(2, 2, 2, term("male")), cand(3, 3, 3, term("female")))}}
}

// denied is an entity whose weak female is denied by member 2's negative.
func denied(t *testing.T) subjectprofile.Entity {
	return subjectprofile.Entity{Properties: props{sex: property(t, sex,
		weak(cand(1, 1, 1, term("female"))), negative(cand(2, 2, 2, term("female"))))}}
}

// ev is a life event of the given type and role on one date.
func ev(eventType vocabkey.EventType, role vocabkey.Role, d *datevalue.Value) subjectprofile.LifeEvent {
	return subjectprofile.LifeEvent{EventType: eventType, Role: role, Dates: []datevalue.Value{*d}}
}
