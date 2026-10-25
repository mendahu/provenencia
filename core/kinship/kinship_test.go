package kinship_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var terms = segmentcomparetest.Terms()

func bridge(t testing.TB, key string) connectrules.Bridge {
	t.Helper()
	b, ok := connectrules.LookupBridge(key)
	if !ok {
		t.Fatalf("no %s bridge", key)
	}
	return b
}

// Subject ids: people P1.. = 0x11.., events E1.. = 0x21.., bridges b1.. = 0x31...
func p(n byte) []byte { return []byte{0x10 + n} }
func e(n byte) []byte { return []byte{0x20 + n} }
func b(n byte) []byte { return []byte{0x30 + n} }

// part is a participation edge with the person at A.
func part(t testing.TB, person, event byte, role vocabkey.Role, eventType vocabkey.EventType, bridgeID byte) kinship.Edge {
	ka, kb := segmentcompare.Keys(bridge(t, vocabkey.BridgeParticipation), vocabkey.BridgeTerm(role), eventType, terms)
	return kinship.Edge{A: p(person), B: e(event), KeyA: ka, KeyB: kb, Associations: [][]byte{b(bridgeID)}}
}

// rel is a relationship edge with the person endpoint at A.
func rel(t testing.TB, person, relatedTo byte, term vocabkey.BridgeTerm, bridgeID byte) kinship.Edge {
	ka, kb := segmentcompare.Keys(bridge(t, vocabkey.BridgeRelationship), term, "", terms)
	return kinship.Edge{A: p(person), B: p(relatedTo), KeyA: ka, KeyB: kb, Associations: [][]byte{b(bridgeID)}}
}

func flip(x kinship.Edge) kinship.Edge {
	x.A, x.B, x.KeyA, x.KeyB = x.B, x.A, x.KeyB, x.KeyA
	return x
}

// link is the Edge an implied relationship link has.
func link(t testing.TB, a, c byte, term vocabkey.BridgeTerm, assoc ...byte) kinship.Edge {
	x := rel(t, a, c, term, 0)
	x.Associations = nil
	for _, n := range assoc {
		x.Associations = append(x.Associations, b(n))
	}
	return x
}

func TestDerive(t *testing.T) {
	const (
		marriage = vocabkey.EventMarriage
		birth    = vocabkey.EventBirth
	)
	spouse := func(person, event, br byte) kinship.Edge {
		return part(t, person, event, vocabkey.RoleSpouse, marriage, br)
	}
	parent := func(g, m, br byte) kinship.Edge { return rel(t, g, m, vocabkey.TermParent, br) }
	birthParents := func(father, subject, event, b1, b2 byte) kinship.Implied {
		pa := part(t, father, event, vocabkey.RoleFather, birth, b1)
		ch := part(t, subject, event, vocabkey.RoleSubject, birth, b2)
		return kinship.Implied{
			Edge:    link(t, father, subject, vocabkey.TermParent, b1, b2),
			Pattern: kinship.PatternBirthParents, Event: e(event), From: []kinship.Edge{pa, ch},
		}
	}
	marriedPair := func(event, b1, b2 byte) kinship.Implied {
		return kinship.Implied{
			Edge:    link(t, 1, 2, vocabkey.TermSpouse, b1, b2),
			Pattern: kinship.PatternMarriageSpouses, Event: e(event),
			From: []kinship.Edge{spouse(1, event, b1), spouse(2, event, b2)},
		}
	}
	grand := func(g, c, through byte, from []kinship.Edge, assoc ...byte) kinship.Implied {
		return kinship.Implied{
			Edge:    link(t, g, c, vocabkey.TermGrandparent, assoc...),
			Pattern: kinship.PatternParentOfParent, Through: p(through), From: from,
		}
	}

	birthP2 := birthParents(2, 3, 1, 2, 3)
	cases := []struct {
		name  string
		edges []kinship.Edge
		want  []kinship.Implied
	}{
		{"AC-1.1", []kinship.Edge{spouse(1, 1, 1), spouse(2, 1, 2)}, []kinship.Implied{marriedPair(1, 1, 2)}},
		{"AC-1.2", []kinship.Edge{flip(spouse(1, 1, 1)), flip(spouse(2, 1, 2))}, []kinship.Implied{
			{
				Edge:    link(t, 1, 2, vocabkey.TermSpouse, 1, 2),
				Pattern: kinship.PatternMarriageSpouses, Event: e(1),
				From: []kinship.Edge{flip(spouse(1, 1, 1)), flip(spouse(2, 1, 2))},
			},
		}},
		{"AC-1.3", []kinship.Edge{
			part(t, 1, 1, vocabkey.RoleFather, birth, 1),
			part(t, 2, 1, vocabkey.RoleMother, birth, 2),
			part(t, 3, 1, vocabkey.RoleSubject, birth, 3),
		}, []kinship.Implied{
			{
				Edge:    link(t, 1, 3, vocabkey.TermParent, 1, 3),
				Pattern: kinship.PatternBirthParents, Event: e(1),
				From: []kinship.Edge{part(t, 1, 1, vocabkey.RoleFather, birth, 1), part(t, 3, 1, vocabkey.RoleSubject, birth, 3)},
			},
			{
				Edge:    link(t, 2, 3, vocabkey.TermParent, 2, 3),
				Pattern: kinship.PatternBirthParents, Event: e(1),
				From: []kinship.Edge{part(t, 2, 1, vocabkey.RoleMother, birth, 2), part(t, 3, 1, vocabkey.RoleSubject, birth, 3)},
			},
		}},
		{"AC-1.4", []kinship.Edge{parent(1, 2, 1), parent(2, 3, 2)}, []kinship.Implied{
			grand(1, 3, 2, []kinship.Edge{parent(1, 2, 1), parent(2, 3, 2)}, 1, 2),
		}},
		{"AC-1.5", []kinship.Edge{
			parent(1, 2, 1),
			part(t, 2, 1, vocabkey.RoleFather, birth, 2),
			part(t, 3, 1, vocabkey.RoleSubject, birth, 3),
		}, []kinship.Implied{
			birthP2,
			grand(1, 3, 2, []kinship.Edge{parent(1, 2, 1), birthP2.Edge}, 1, 2, 3),
		}},
		{"AC-1.6", []kinship.Edge{rel(t, 2, 1, "child", 1), parent(2, 3, 2)}, []kinship.Implied{
			grand(1, 3, 2, []kinship.Edge{rel(t, 2, 1, "child", 1), parent(2, 3, 2)}, 1, 2),
		}},
		{"AC-1.7", []kinship.Edge{rel(t, 1, 3, vocabkey.TermGrandparent, 1)}, nil},
		{"AC-1.8", []kinship.Edge{parent(1, 2, 1), parent(2, 3, 2), parent(3, 4, 3)}, []kinship.Implied{
			grand(1, 3, 2, []kinship.Edge{parent(1, 2, 1), parent(2, 3, 2)}, 1, 2),
			grand(2, 4, 3, []kinship.Edge{parent(2, 3, 2), parent(3, 4, 3)}, 2, 3),
		}},
		{"AC-1.9 spouse", []kinship.Edge{rel(t, 1, 2, vocabkey.TermSpouse, 1)}, nil},
		{"AC-1.9 parent", []kinship.Edge{parent(1, 2, 1)}, nil},
		{"AC-1.10", []kinship.Edge{
			part(t, 1, 1, "groom", marriage, 1), part(t, 2, 1, "groom", marriage, 2),
			part(t, 1, 2, vocabkey.RoleSpouse, "wedding", 3), part(t, 2, 2, vocabkey.RoleSpouse, "wedding", 4),
		}, nil},
		{"AC-1.11 child role", []kinship.Edge{
			part(t, 1, 1, vocabkey.RoleFather, birth, 1), part(t, 2, 1, "child", birth, 2),
		}, nil},
		{"AC-1.11 witness", []kinship.Edge{
			part(t, 1, 1, "witness", marriage, 1), spouse(2, 1, 2),
		}, nil},
		{"AC-1.11 father in marriage", []kinship.Edge{
			part(t, 1, 1, vocabkey.RoleFather, marriage, 1), part(t, 2, 1, vocabkey.RoleSubject, marriage, 2),
		}, nil},
		{"AC-1.12 birth", []kinship.Edge{
			part(t, 1, 1, vocabkey.RoleFather, birth, 1), part(t, 1, 1, vocabkey.RoleSubject, birth, 2),
		}, nil},
		{"AC-1.12 cycle", []kinship.Edge{parent(1, 2, 1), parent(2, 1, 2)}, nil},
		{"AC-1.13", []kinship.Edge{
			parent(1, 2, 1),
			part(t, 1, 1, vocabkey.RoleFather, birth, 2),
			part(t, 2, 1, vocabkey.RoleSubject, birth, 3),
			parent(2, 3, 4),
		}, []kinship.Implied{
			birthParents(1, 2, 1, 2, 3),
			grand(1, 3, 2, []kinship.Edge{parent(1, 2, 1), birthParents(1, 2, 1, 2, 3).Edge, parent(2, 3, 4)}, 1, 2, 3, 4),
		}},
		{"AC-1.14 empty", []kinship.Edge{}, nil},
		{"AC-1.14 nil", nil, nil},
		{"AC-1.15", []kinship.Edge{spouse(2, 1, 2), spouse(1, 1, 1)}, []kinship.Implied{
			{
				Edge:    link(t, 1, 2, vocabkey.TermSpouse, 1, 2),
				Pattern: kinship.PatternMarriageSpouses, Event: e(1),
				From: []kinship.Edge{spouse(1, 1, 1), spouse(2, 1, 2)},
			},
		}},
		{"AC-1.16", []kinship.Edge{spouse(1, 1, 1), spouse(2, 1, 2), spouse(1, 2, 3), spouse(2, 2, 4)}, []kinship.Implied{
			marriedPair(1, 1, 2), marriedPair(2, 3, 4),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := slices.Clone(tc.edges)
			got := kinship.Derive(tc.edges, terms)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Derive =\n%+v\nwant\n%+v", got, tc.want)
			}
			if !reflect.DeepEqual(tc.edges, before) {
				t.Errorf("Derive changed its input")
			}
		})
	}
}

func TestDeriveInvariants(t *testing.T) {
	relB := bridge(t, vocabkey.BridgeRelationship)
	keysOf := func(term vocabkey.BridgeTerm) [2]vocabkey.BridgeKey {
		a, c := segmentcompare.Keys(relB, term, "", terms)
		return [2]vocabkey.BridgeKey{a, c}
	}
	allowed := [][2]vocabkey.BridgeKey{keysOf(vocabkey.TermSpouse), keysOf(vocabkey.TermParent), keysOf(vocabkey.TermGrandparent)}
	flatten := func(imps []kinship.Implied) []string {
		var out []string
		for _, i := range imps {
			out = append(out, fmt.Sprintf("%x %x %v %v %s %x %x", i.A, i.B, i.KeyA, i.KeyB, i.Pattern, i.Event, i.Through))
		}
		slices.Sort(out)
		return out
	}
	eventTypes := []vocabkey.EventType{vocabkey.EventBirth, vocabkey.EventMarriage, "wedding"}
	roles := []vocabkey.Role{vocabkey.RoleSubject, vocabkey.RoleFather, vocabkey.RoleMother, vocabkey.RoleSpouse, "witness", "groom"}
	relTerms := []vocabkey.BridgeTerm{vocabkey.TermParent, "child", vocabkey.TermSpouse, vocabkey.TermGrandparent}

	r := rand.New(rand.NewSource(1))
	for i := 0; i < 200; i++ {
		people := 2 + r.Intn(5)
		events := 1 + r.Intn(3)
		evType := make([]vocabkey.EventType, events+1)
		for n := 1; n <= events; n++ {
			evType[n] = eventTypes[r.Intn(len(eventTypes))]
		}
		person := func() byte { return byte(1 + r.Intn(people)) }
		var edges []kinship.Edge
		bridgeID := byte(0)
		for n := r.Intn(11); n > 0; n-- {
			bridgeID++
			var x kinship.Edge
			if r.Intn(2) == 0 {
				ev := byte(1 + r.Intn(events))
				x = part(t, person(), ev, roles[r.Intn(len(roles))], evType[ev], bridgeID)
			} else {
				x = rel(t, person(), person(), relTerms[r.Intn(len(relTerms))], bridgeID)
			}
			if r.Intn(2) == 0 {
				x = flip(x)
			}
			edges = append(edges, x)
		}

		got := kinship.Derive(edges, terms)
		var assocs [][]byte
		for _, x := range edges {
			assocs = append(assocs, x.Associations...)
		}
		for _, imp := range got {
			keys := [2]vocabkey.BridgeKey{imp.KeyA, imp.KeyB}
			if !slices.Contains(allowed, keys) || slices.Equal(imp.A, imp.B) {
				t.Fatalf("case %d: implied link %v has keys %v or joins a Subject to itself", i, imp, keys)
			}
			for _, a := range imp.Associations {
				if !slices.ContainsFunc(assocs, func(x []byte) bool { return slices.Equal(x, a) }) {
					t.Fatalf("case %d: association %x is in no input edge", i, a)
				}
			}
		}

		reversed := slices.Clone(edges)
		slices.Reverse(reversed)
		if a, c := flatten(got), flatten(kinship.Derive(reversed, terms)); !slices.Equal(a, c) {
			t.Fatalf("case %d: reversed input gives %s, want %s", i, strings.Join(c, "; "), strings.Join(a, "; "))
		}

		catalogOnly := append(slices.Clone(edges),
			part(t, person(), 1, "groom", evType[1], 100), rel(t, person(), person(), "godparent", 101))
		if again := kinship.Derive(catalogOnly, terms); !reflect.DeepEqual(again, got) {
			t.Fatalf("case %d: a catalog's own role and term changed the output", i)
		}

		withDirect := append(slices.Clone(edges), rel(t, 1, 2, vocabkey.TermGrandparent, 102))
		if again := kinship.Derive(withDirect, terms); !reflect.DeepEqual(again, got) {
			t.Fatalf("case %d: a direct grandparent link changed the output", i)
		}
	}
}
