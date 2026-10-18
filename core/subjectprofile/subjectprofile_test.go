package subjectprofile_test

import (
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

const (
	sex      = vocabkey.PropertySexAtBirth
	date     = vocabkey.PropertyDate
	toponym  = vocabkey.PropertyToponym
	nameProp = vocabkey.PropertyName
)

type props = map[vocabkey.PropertyKey]subjectprofile.Property

func TestBuild(t *testing.T) {
	var named subjectprofile.Entity // the entity AC-1.9 built, for its check
	maleFemale := func(t *testing.T) subjectprofile.Entity {
		return subjectprofile.Entity{Properties: props{sex: property(t, sex,
			cand(1, 1, 1, term("male")), cand(2, 2, 2, term("male")), cand(3, 3, 3, term("female")))}}
	}
	denied := func(t *testing.T) subjectprofile.Entity {
		return subjectprofile.Entity{Properties: props{sex: property(t, sex,
			weak(cand(1, 1, 1, term("female"))), negative(cand(2, 2, 2, term("female"))))}}
	}
	ex := func(ms ...byte) subjectprofile.Options {
		var o subjectprofile.Options
		for _, m := range ms {
			o.Excluding = append(o.Excluding, mem(m))
		}
		return o
	}
	tests := []struct {
		name    string
		entity  func(t *testing.T) subjectprofile.Entity
		metas   subjectcompare.Metas
		opts    subjectprofile.Options
		want    string
		wantErr error
		check   func(t *testing.T, p subjectcompare.Values)
	}{
		{name: "AC-1.1 outvoted value is left out", entity: maleFemale, want: "sex_at_birth={male}"},
		{name: "AC-1.2 denied value is a negative", entity: denied, want: "sex_at_birth={¬female}"},
		{name: "AC-1.3 mixed dates are both positive", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{date: property(t, date,
				cand(1, 1, 1, year(1817)), cand(2, 2, 2, year(1830)))}}
		}, want: "date={1817, 1830}"},
		{name: "AC-1.4 multiple keeps every value", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{toponym: property(t, toponym,
				cand(1, 1, 1, text("York")), cand(2, 2, 2, text("York")), cand(3, 3, 3, text("London")))}}
		}, want: "toponym={York, London}"},
		{name: "AC-1.5 weak value is left out", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{sex: property(t, sex,
				cand(1, 1, 1, term("male")), weak(cand(2, 2, 2, term("female"))))}}
		}, want: "sex_at_birth={male}"},
		{name: "AC-1.6 provisional value is left out", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{sex: property(t, sex,
				cand(1, 1, 1, term("male")), provisional(cand(2, 2, 2, term("female"))))}}
		}, want: "sex_at_birth={male}"},
		{name: "AC-1.7 denied date beside a kept one", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{date: property(t, date,
				cand(1, 1, 1, year(1817)), weak(cand(2, 2, 2, year(1830))), negative(cand(3, 3, 3, year(1830))))}}
		}, want: "date={1817, ¬1830}"},
		{name: "AC-1.8 a property with nothing left is absent", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{sex: property(t, sex, provisional(cand(1, 1, 1, term("female"))))}}
		}, check: func(t *testing.T, p subjectcompare.Values) {
			if p == nil || len(p) != 0 {
				t.Errorf("want non-nil empty profile, got %v", p)
			}
			if _, ok := p[sex]; ok {
				t.Error("sex_at_birth is a key of the profile")
			}
		}},
		{name: "AC-1.9 a name carries the entity's pattern", entity: func(t *testing.T) subjectprofile.Entity {
			named = subjectprofile.Entity{NamePattern: "pattern-1", Properties: props{nameProp: property(t, nameProp,
				cand(1, 1, 1, name("James Robins")))}}
			return named
		}, check: func(t *testing.T, p subjectcompare.Values) {
			if len(p[nameProp]) != 1 {
				t.Fatalf("want one name value, got %v", p[nameProp])
			}
			v := p[nameProp][0]
			if v.Name == nil || v.Name.Form != "James Robins" || v.NamePattern != "pattern-1" ||
				v.Type != "" || v.Ref != nil || v.Negative || v.Neutral {
				t.Errorf("got %+v", v)
			}
			if v.Name != named.Properties[nameProp].Values[0].Value.Name {
				t.Error("the profile's Name is not the entity's own pointer")
			}
		}},
		{name: "AC-1.10 a term carries only its key", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{NamePattern: "pattern-1", Properties: props{sex: property(t, sex, cand(1, 1, 1, term("male")))}}
		}, check: func(t *testing.T, p subjectcompare.Values) {
			v := p[sex][0]
			if v.Term != "male" || v.Type != "" || v.Ref != nil || v.NamePattern != "" {
				t.Errorf("got %+v", v)
			}
		}},
		{name: "AC-1.11 a neutral value stays neutral", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{sex: {Values: []autoreconcile.ReconciledValue{{
				Value:  autoreconcile.Value{TermKey: "unknown", Neutral: true},
				Reason: autoreconcile.ReasonKept,
			}}}}}
		}, check: func(t *testing.T, p subjectcompare.Values) {
			if len(p[sex]) != 1 || p[sex][0].Term != "unknown" || !p[sex][0].Neutral {
				t.Errorf("got %v", p[sex])
			}
		}},
		{name: "AC-1.12 the zero entity", entity: func(*testing.T) subjectprofile.Entity { return subjectprofile.Entity{} },
			check: func(t *testing.T, p subjectcompare.Values) {
				if p == nil || len(p) != 0 {
					t.Errorf("want non-nil empty profile, got %v", p)
				}
			}},

		{name: "AC-2.1 leaving out a voter removes the majority", entity: maleFemale, opts: ex(1), want: "sex_at_birth={male, female}"},
		{name: "AC-2.2 leaving out the outvoted member", entity: maleFemale, opts: ex(3), want: "sex_at_birth={male}"},
		{name: "AC-2.3 a member with no candidate changes nothing", entity: maleFemale, opts: ex(9), want: "sex_at_birth={male}"},
		{name: "AC-2.4 a property the member never touched keeps its values", entity: func(t *testing.T) subjectprofile.Entity {
			e := maleFemale(t)
			e.Properties[toponym] = subjectprofile.Property{
				Values: reconcile(t, toponym, []subjectprofile.Candidate{cand(2, 2, 2, text("York"))}),
			}
			return e
		}, opts: ex(1), want: "sex_at_birth={male, female}; toponym={York}"},
		{name: "AC-2.5 leaving out the negative restores the weak value", entity: denied, opts: ex(2), want: "sex_at_birth={female}"},
		{name: "AC-2.6 leaving out two members", entity: maleFemale, opts: ex(1, 2), want: "sex_at_birth={female}"},
		{name: "AC-2.7 leaving out the only member empties the profile", entity: func(t *testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{date: property(t, date, cand(1, 1, 1, year(1817)))}}
		}, opts: ex(1), check: func(t *testing.T, p subjectcompare.Values) {
			if p == nil || len(p) != 0 {
				t.Errorf("want non-nil empty profile, got %v", p)
			}
		}},
		{name: "AC-2.8 a mismatched candidate is Reconcile's error", entity: func(*testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{date: {Candidates: []subjectprofile.Candidate{
				cand(1, 1, 1, year(1817)), cand(2, 2, 2, text("x"))}}}}
		}, opts: ex(1), wantErr: autoreconcile.ErrValueMismatch, check: func(t *testing.T, p subjectcompare.Values) {
			if p != nil {
				t.Errorf("want nil profile, got %v", p)
			}
		}},
		{name: "AC-2.9 an unknown property is Reconcile's error", entity: func(*testing.T) subjectprofile.Entity {
			return subjectprofile.Entity{Properties: props{"occupation": {Candidates: []subjectprofile.Candidate{
				cand(1, 1, 1, text("x"))}}}}
		}, opts: ex(1), wantErr: autoreconcile.ErrUnknownValueType, check: func(t *testing.T, p subjectcompare.Values) {
			if p != nil {
				t.Errorf("want nil profile, got %v", p)
			}
		}},
		{name: "AC-2.10 empty cardinality reads as single", entity: maleFemale,
			metas: subjectcompare.Metas{sex: {ValueType: propertymeta.ValueTypeTerm}}, opts: ex(1),
			want: "sex_at_birth={male, female}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			metas := tc.metas
			if metas == nil {
				metas = subjectcomparetest.Metas()
			}
			got, err := subjectprofile.Build(tc.entity(t), metas, matchregistry.Default().Value, tc.opts)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if tc.check != nil {
				tc.check(t, got)
				return
			}
			if g := subjectcomparetest.Labels(got); g != tc.want {
				t.Errorf("got %s, want %s", g, tc.want)
			}
		})
	}
}

func TestForSubject(t *testing.T) {
	c := func(v autoreconcile.Value) autoreconcile.Candidate { return autoreconcile.Candidate{Value: v} }
	neg := func(v autoreconcile.Value) autoreconcile.Candidate {
		return autoreconcile.Candidate{Value: v, Negative: true}
	}
	james := name("James Robins")
	tests := []struct {
		name        string
		values      map[vocabkey.PropertyKey][]autoreconcile.Candidate
		namePattern string
		want        string
		check       func(t *testing.T, p subjectcompare.Values)
	}{
		{name: "AC-5.1 positives and negatives in input order",
			values: map[vocabkey.PropertyKey][]autoreconcile.Candidate{sex: {c(term("male")), neg(term("female"))}},
			want:   "sex_at_birth={male, ¬female}"},
		{name: "AC-5.2 a name carries the pattern",
			values:      map[vocabkey.PropertyKey][]autoreconcile.Candidate{nameProp: {c(james)}},
			namePattern: "pattern-1",
			check: func(t *testing.T, p subjectcompare.Values) {
				if len(p[nameProp]) != 1 || p[nameProp][0].NamePattern != "pattern-1" || p[nameProp][0].Name != james.Name {
					t.Errorf("got %+v", p[nameProp])
				}
			}},
		{name: "AC-5.3 provenance and provisional are not read",
			values: map[vocabkey.PropertyKey][]autoreconcile.Candidate{sex: {{
				Value: term("female"), Provisional: true, Provenance: autoreconcile.Provenance{Credibility: -1},
			}}},
			want: "sex_at_birth={female}"},
		{name: "AC-5.4 empty input gives an empty non-nil profile",
			values: map[vocabkey.PropertyKey][]autoreconcile.Candidate{date: {}},
			check: func(t *testing.T, p subjectcompare.Values) {
				if p == nil || len(p) != 0 {
					t.Errorf("want non-nil empty profile, got %v", p)
				}
				if n := subjectprofile.ForSubject(nil, ""); n == nil || len(n) != 0 {
					t.Errorf("nil map: want non-nil empty profile, got %v", n)
				}
			}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := subjectprofile.ForSubject(tc.values, tc.namePattern)
			if tc.check != nil {
				tc.check(t, got)
				return
			}
			if g := subjectcomparetest.Labels(got); g != tc.want {
				t.Errorf("got %s, want %s", g, tc.want)
			}
		})
	}
}
