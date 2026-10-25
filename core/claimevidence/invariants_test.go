package claimevidence_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// randomMembers builds n members, each with a random name and date whose
// Observation ids are unique across the set.
func randomMembers(r *rand.Rand, n int) []claimevidence.Member {
	names := []string{"John Smith", "Jon Smith", "J. Smith", "Mary Jones", "José Smith"}
	var ms []claimevidence.Member
	for i := 0; i < n; i++ {
		id := byte(i + 1)
		ms = append(ms, member(id, vals{
			"name": {ref(name(names[r.Intn(len(names))]), 0x40+id)},
			"date": {ref(propertycomparetest.Date(datevaluetest.YMD(1810+r.Intn(3)*10, 0, 0)), 0x80+id)},
		}))
	}
	return ms
}

// shape is a section list reduced to what reordering members must not change:
// each line's incoming ref, verdict and the set of member Subjects it holds.
func shape(ss []claimevidence.Section) []string {
	var out []string
	for _, s := range ss {
		for _, l := range s.Lines {
			subjects := pins(l)
			slices.Sort(subjects)
			out = append(out, fmt.Sprint(s.Key, l.Incoming.Ref, l.Relation, l.Cause, l.Outcome, subjects))
		}
	}
	slices.Sort(out)
	return out
}

func TestSectionsInvariants(t *testing.T) {
	r := rand.New(rand.NewSource(11))
	for i := 0; i < 200; i++ {
		var kind vocabkey.SubjectType = vocabkey.Event
		in := vals{"date": {ref(propertycomparetest.Date(datevaluetest.YMD(1810+r.Intn(3)*10, 0, 0)), 1)}}
		ms := randomMembers(r, 3+r.Intn(4))
		if i%2 == 0 {
			kind = vocabkey.Person
			in = vals{"name": {ref(name("John Smith"), 1)}}
		}
		got := sections(kind, in, ms...)

		// Reordering the members changes only which line holds which member
		// first: the lines are the same set.
		shuffled := slices.Clone(ms)
		r.Shuffle(len(shuffled), func(a, b int) { shuffled[a], shuffled[b] = shuffled[b], shuffled[a] })
		if a, b := shape(got), shape(sections(kind, in, shuffled...)); !slices.Equal(a, b) {
			t.Fatalf("run %d: reordering members changed the lines:\n%v\n%v", i, a, b)
		}

		// A member with no values changes nothing.
		withEmpty := append(slices.Clone(ms), member(0xEE, vals{}))
		if a, b := shape(got), shape(sections(kind, in, withEmpty...)); !slices.Equal(a, b) {
			t.Fatalf("run %d: an empty member changed the lines", i)
		}

		held := map[string]bool{}
		for _, m := range ms {
			for _, vs := range m.Values {
				for _, v := range vs {
					held[string(v.Ref)] = true
				}
			}
		}
		for _, s := range got {
			for _, l := range s.Lines {
				for _, p := range l.Members {
					if !held[string(p.Observation)] {
						t.Fatalf("run %d: pin %x is no member's ref", i, p.Observation)
					}
				}
				if l.Outcome == propertycompare.OutcomeConflict && l.Checked {
					t.Fatalf("run %d: a conflict starts checked: %+v", i, l)
				}
			}
			if !slices.IsSortedFunc(s.Lines, func(a, b claimevidence.Line) int { return bytes.Compare(a.Incoming.Ref, b.Incoming.Ref) }) {
				t.Fatalf("run %d: lines are not in incoming ref order", i)
			}
		}
	}
}
