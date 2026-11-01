package kinship_test

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// sib is a sibling link as a test reads it: ends and pivot as person numbers
// (ids are 0x10+n), associations as bridge numbers.
type sib struct {
	A, B, Through byte
	Assoc         []byte
}

func sibsOf(imps []kinship.Implied) []sib {
	var out []sib
	for _, i := range imps {
		s := sib{A: i.A[0] - 0x10, B: i.B[0] - 0x10, Through: i.Through[0] - 0x10}
		for _, a := range i.Associations {
			s.Assoc = append(s.Assoc, a[0]-0x30)
		}
		out = append(out, s)
	}
	return out
}

func TestSiblings(t *testing.T) {
	parent := func(g, m, br byte) kinship.Edge { return rel(t, g, m, vocabkey.TermParent, br) }
	birth := func(father, subject, event, b1, b2 byte) []kinship.Edge {
		return []kinship.Edge{
			part(t, father, event, vocabkey.RoleFather, vocabkey.EventBirth, b1),
			part(t, subject, event, vocabkey.RoleSubject, vocabkey.EventBirth, b2),
		}
	}
	check := func(t *testing.T, person byte, edges []kinship.Edge, want ...sib) {
		t.Helper()
		got := sibsOf(kinship.Siblings(p(person), edges, terms, rules))
		if len(got) == 0 && len(want) == 0 {
			return
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Siblings(P%d):\n got %+v\nwant %+v", person, got, want)
		}
	}

	t.Run("AC-1.2", func(t *testing.T) {
		edges := []kinship.Edge{parent(1, 2, 1), parent(1, 3, 2)}
		check(t, 2, edges, sib{2, 3, 1, []byte{1, 2}})
		check(t, 3, edges, sib{2, 3, 1, []byte{2, 1}})
		imps := kinship.Siblings(p(2), edges, terms, rules)
		if imps[0].Pattern != kinship.PatternSharedParent || imps[0].Event != nil {
			t.Fatalf("got pattern %q event %v", imps[0].Pattern, imps[0].Event)
		}
		if got := imps[0].Shape(); got.Pattern != kinship.PatternSharedParent || !slices.Equal(got.Pivot, p(1)) {
			t.Fatalf("shape %+v", got)
		}
		if len(imps[0].From) != 2 {
			t.Fatalf("From has %d edges, want 2", len(imps[0].From))
		}
		ka, kb := imps[0].KeyA, imps[0].KeyB
		if ka != kb || ka.Term != vocabkey.TermSibling {
			t.Fatalf("keys %v %v", ka, kb)
		}
	})
	t.Run("AC-1.3", func(t *testing.T) {
		// The birth links Derive returns for two births with one father.
		edges := slices.Concat(
			[]kinship.Edge{part(t, 1, 1, vocabkey.RoleFather, vocabkey.EventBirth, 1), part(t, 2, 1, vocabkey.RoleSubject, vocabkey.EventBirth, 2)},
			[]kinship.Edge{part(t, 1, 2, vocabkey.RoleFather, vocabkey.EventBirth, 3), part(t, 3, 2, vocabkey.RoleSubject, vocabkey.EventBirth, 4)},
		)
		var births []kinship.Edge
		for _, imp := range kinship.Derive(edges, terms, rules) {
			if imp.Pattern == kinship.PatternBirthParents {
				births = append(births, imp.Edge)
			}
		}
		if len(births) != 2 {
			t.Fatalf("Derive made %d birth links, want 2", len(births))
		}
		check(t, 2, births, sib{2, 3, 1, []byte{1, 2, 3, 4}})
	})
	t.Run("AC-1.5", func(t *testing.T) {
		edges := []kinship.Edge{parent(1, 2, 1), parent(1, 3, 2), parent(4, 2, 3), parent(4, 3, 4)}
		check(t, 2, edges, sib{2, 3, 1, []byte{1, 2}}, sib{2, 3, 4, []byte{3, 4}})
		if got := kinship.Parents(p(2), edges, terms, rules); !reflect.DeepEqual(got, [][]byte{p(1), p(4)}) {
			t.Fatalf("Parents = %v", got)
		}
	})
	t.Run("AC-1.6", func(t *testing.T) {
		edges := []kinship.Edge{parent(1, 2, 1), parent(1, 3, 2), parent(4, 3, 3)}
		check(t, 2, edges, sib{2, 3, 1, []byte{1, 2}})
		check(t, 3, edges, sib{2, 3, 1, []byte{2, 1}})
	})
	t.Run("AC-1.7", func(t *testing.T) {
		edges := []kinship.Edge{
			parent(1, 2, 1), rel(t, 1, 4, vocabkey.TermSpouse, 2), parent(4, 5, 3), parent(2, 3, 4),
		}
		check(t, 2, edges)
		check(t, 3, edges)
		check(t, 5, edges)
	})
	t.Run("AC-1.1", func(t *testing.T) {
		// A recorded sibling link is Relationships' job, not Siblings'.
		check(t, 2, []kinship.Edge{rel(t, 2, 3, vocabkey.TermSibling, 1)})
	})
	t.Run("AC-1.8", func(t *testing.T) {
		edges := []kinship.Edge{parent(1, 2, 1)}
		var want []sib
		for n := byte(3); n < 17; n++ {
			edges = append(edges, parent(1, n, n))
			want = append(want, sib{2, n, 1, []byte{1, n}})
		}
		check(t, 2, edges, want...)
	})
	t.Run("AC-1.9", func(t *testing.T) {
		check(t, 2, nil)
		check(t, 2, []kinship.Edge{parent(2, 2, 1)})
		if got := kinship.Parents(p(2), nil, terms, rules); len(got) != 0 {
			t.Fatalf("Parents of nothing = %v", got)
		}
	})
	t.Run("AC-1.10", func(t *testing.T) {
		edges := []kinship.Edge{parent(1, 2, 1), parent(1, 3, 2), parent(3, 2, 3)}
		check(t, 2, edges, sib{2, 3, 1, []byte{1, 2}})
		if got := kinship.Parents(p(2), edges, terms, rules); !reflect.DeepEqual(got, [][]byte{p(1), p(3)}) {
			t.Fatalf("Parents = %v", got)
		}
	})
	t.Run("duplicate steps", func(t *testing.T) {
		// The same step passed twice adds no association.
		edges := []kinship.Edge{parent(1, 2, 1), parent(1, 2, 1), parent(1, 3, 2)}
		check(t, 2, edges, sib{2, 3, 1, []byte{1, 2}})
		if got := kinship.Parents(p(2), edges, terms, rules); !reflect.DeepEqual(got, [][]byte{p(1)}) {
			t.Fatalf("Parents = %v", got)
		}
	})
	t.Run("a birth step", func(t *testing.T) {
		edges := slices.Concat(birth(1, 2, 1, 1, 2), []kinship.Edge{parent(1, 3, 3)})
		var steps []kinship.Edge
		for _, imp := range kinship.Derive(edges, terms, rules) {
			if imp.Pattern == kinship.PatternBirthParents {
				steps = append(steps, imp.Edge)
			}
		}
		check(t, 2, append(steps, parent(1, 3, 3)), sib{2, 3, 1, []byte{1, 2, 3}})
	})
	t.Run("AC-3.1", func(t *testing.T) {
		edges := slices.Concat(
			[]kinship.Edge{parent(1, 2, 1), parent(1, 3, 2)},
			birth(1, 2, 1, 3, 4), birth(1, 3, 2, 5, 6),
		)
		for _, imp := range kinship.Derive(edges, terms, rules) {
			if imp.Pattern == kinship.PatternSharedParent {
				t.Fatalf("Derive made %+v", imp)
			}
			if imp.Edge.KeyA.Term == vocabkey.TermSibling {
				t.Fatalf("Derive made a sibling link %+v", imp)
			}
		}
		if slices.Contains(kinship.Patterns(), kinship.PatternSharedParent) {
			t.Fatal("Patterns lists PatternSharedParent")
		}
	})
}

// TestSiblingsManyDuplicatesTimesChildren files the person's parent link 2,000
// times and 2,000 children of the parent: the work stays linear, and each link
// keeps at most MaxPairSteps steps of each pair.
func TestSiblingsManyDuplicatesTimesChildren(t *testing.T) {
	const n = 2000
	rel := segmentcomparetest.Bridge(t, vocabkey.BridgeRelationship)
	first, second := segmentcompare.Keys(rel, vocabkey.TermParent, "", terms)
	var edges []kinship.Edge
	for i := 0; i < n; i++ {
		a := []byte(fmt.Sprintf("a%d", i))
		edges = append(edges,
			kinship.Edge{A: []byte("q"), B: []byte("person"), KeyA: first, KeyB: second, Associations: [][]byte{a}},
			kinship.Edge{A: []byte("q"), B: []byte(fmt.Sprintf("c%d", i)), KeyA: first, KeyB: second, Associations: [][]byte{a}})
	}
	start := time.Now()
	got := kinship.Siblings([]byte("person"), edges, terms, rules)
	if len(got) != n {
		t.Fatalf("Siblings made %d links, want %d", len(got), n)
	}
	for _, imp := range got {
		if len(imp.From) > rules.MaxPairSteps+1 {
			t.Fatalf("a link's From holds %d edges, want at most %d", len(imp.From), rules.MaxPairSteps+1)
		}
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Siblings took %v, want under a second", d)
	}
}
