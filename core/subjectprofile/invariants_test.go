package subjectprofile_test

import (
	"math/rand"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/autoreconcile"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

const (
	invariantSeed  = 1
	invariantCases = 500
)

// pools are the values each generated property draws from.
var pools = map[vocabkey.PropertyKey][]autoreconcile.Value{
	sex:     {term("male"), term("female")},
	date:    {year(1817), year(1830), {Date: datevaluetest.About(1817, 0, 0)}},
	toponym: {text("York"), text("London")},
}

var generated = []vocabkey.PropertyKey{sex, date, toponym}

// genCandidates is one entity's candidates by property.
type genCandidates map[vocabkey.PropertyKey][]subjectprofile.Candidate

func generate(r *rand.Rand) (genCandidates, []byte) {
	members := 1 + r.Intn(4)
	out := genCandidates{}
	next := byte(1)
	for _, key := range generated {
		for m := 1; m <= members; m++ {
			for n := r.Intn(3); n > 0; n-- {
				pool := pools[key]
				c := cand(byte(m), next, byte(1+r.Intn(3)), pool[r.Intn(len(pool))])
				next++
				if r.Intn(4) == 0 {
					c = negative(c)
				}
				if r.Intn(4) == 0 {
					c = weak(c)
				}
				if r.Intn(6) == 0 {
					c = provisional(c)
				}
				out[key] = append(out[key], c)
			}
		}
	}
	return out, mem(byte(1 + r.Intn(members)))
}

func entityOf(t testing.TB, cs genCandidates) subjectprofile.Entity {
	e := subjectprofile.Entity{Properties: map[vocabkey.PropertyKey]subjectprofile.Property{}}
	for key, list := range cs {
		e.Properties[key] = property(t, key, list...)
	}
	return e
}

func TestInvariants(t *testing.T) {
	metas := subjectcomparetest.Metas()
	rules := matchregistry.Default().Value
	build := func(t *testing.T, e subjectprofile.Entity, ex ...[]byte) subjectcompare.Values {
		t.Helper()
		p, _, err := subjectprofile.Build(e, metas, rules, nil, subjectprofile.Options{Excluding: ex})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	run := func(name string, check func(t *testing.T, i int, r *rand.Rand, cs genCandidates, member []byte)) {
		t.Run(name, func(t *testing.T) {
			r := rand.New(rand.NewSource(invariantSeed))
			for i := 0; i < invariantCases; i++ {
				cs, member := generate(r)
				check(t, i, r, cs, member)
			}
		})
	}
	same := func(t *testing.T, i int, what string, got, want subjectcompare.Values) {
		t.Helper()
		if g, w := subjectcomparetest.Labels(got), subjectcomparetest.Labels(want); g != w {
			t.Errorf("case %d: %s: got %s, want %s", i, what, g, w)
		}
	}

	run("AC-4.1", func(t *testing.T, i int, _ *rand.Rand, cs genCandidates, _ []byte) {
		e := entityOf(t, cs)
		same(t, i, "excluding a member with no candidate", build(t, e, mem(9)), build(t, e))
	})
	run("AC-4.2", func(t *testing.T, i int, _ *rand.Rand, cs genCandidates, member []byte) {
		rest := genCandidates{}
		for key, list := range cs {
			for _, c := range list {
				if string(c.Member) != string(member) {
					rest[key] = append(rest[key], c)
				}
			}
		}
		same(t, i, "leave-one-out equals a group without the member",
			build(t, entityOf(t, cs), member), build(t, entityOf(t, rest)))
	})
	run("AC-4.3", func(t *testing.T, i int, r *rand.Rand, cs genCandidates, member []byte) {
		e := entityOf(t, cs)
		shuffled := subjectprofile.Entity{Properties: map[vocabkey.PropertyKey]subjectprofile.Property{}}
		for key, p := range e.Properties {
			c := append([]subjectprofile.Candidate(nil), p.Candidates...)
			r.Shuffle(len(c), func(a, b int) { c[a], c[b] = c[b], c[a] })
			shuffled.Properties[key] = subjectprofile.Property{Values: p.Values, Candidates: c}
		}
		same(t, i, "candidate order", build(t, shuffled, member), build(t, e, member))
	})
	run("AC-4.4", func(t *testing.T, i int, _ *rand.Rand, cs genCandidates, _ []byte) {
		e := entityOf(t, cs)
		p := build(t, e)
		for key, prop := range e.Properties {
			var wantPos, wantNeg int
			for _, v := range prop.Values {
				if v.Displayed() {
					wantPos++
				}
				if v.Reason == autoreconcile.ReasonDenied {
					wantNeg++
				}
			}
			var gotPos, gotNeg int
			for _, v := range p[key] {
				if v.Negative {
					gotNeg++
				} else {
					gotPos++
				}
			}
			if gotPos != wantPos || gotNeg != wantNeg {
				t.Errorf("case %d: %s: got %d positive %d negative, want %d and %d", i, key, gotPos, gotNeg, wantPos, wantNeg)
			}
		}
	})

	t.Run("AC-4.7", func(t *testing.T) {
		r := rand.New(rand.NewSource(invariantSeed))
		lifespans := matchregistry.Default().Subject.Lifespans[vocabkey.Person]
		types := []vocabkey.EventType{vocabkey.EventBirth, vocabkey.EventDeath, vocabkey.EventBurial, "residence", "marriage"}
		roles := []vocabkey.Role{vocabkey.RoleSubject, "spouse", "father"}
		built := 0
		for i := 0; i < invariantCases; i++ {
			var events []subjectprofile.LifeEvent
			for n := 1 + r.Intn(4); n > 0; n-- {
				events = append(events, ev(types[r.Intn(len(types))], roles[r.Intn(len(roles))], datevaluetest.YMD(1800+r.Intn(200), 0, 0)))
			}
			got, _ := subjectprofile.ForSubject(subjectprofile.Subject{Life: events}, lifespans)
			shuffled := slices.Clone(events)
			r.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })
			again, _ := subjectprofile.ForSubject(subjectprofile.Subject{Life: shuffled}, lifespans)
			same(t, i, "event order", again, got)
			window := got[vocabkey.PropertyLifespan]
			if len(window) == 0 {
				continue
			}
			built++
			w, ok := valuecompare.DateWindow(*window[0].Date)
			if !ok {
				t.Fatalf("case %d: window %s has no days", i, subjectcomparetest.Labels(got))
			}
			for _, e := range events {
				d, _ := valuecompare.DateWindow(e.Dates[0])
				if !w.Overlaps(d) {
					t.Errorf("case %d: window %s misses an event date", i, subjectcomparetest.Labels(got))
				}
			}
		}
		if built < 100 {
			t.Errorf("only %d of %d cases built a lifespan", built, invariantCases)
		}
	})
}
