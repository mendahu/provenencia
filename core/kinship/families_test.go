package kinship_test

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// isEvent reports whether an id is one of e(n): a participation edge has one.
func isEvent(id []byte) bool { return id[0]&0xf0 == 0x20 }

// linksFor builds the LinkSet Relationships builds for person from pure edges:
// each relationship edge with person at an end, the links the patterns imply
// with person at an end, and the siblings a shared parent implies.
func linksFor(t testing.TB, person []byte, edges []kinship.Edge) []kinship.Link {
	t.Helper()
	set := kinship.NewLinkSet()
	var parentEdges []kinship.Edge
	for _, e := range edges {
		if isEvent(e.A) || isEvent(e.B) {
			continue
		}
		parentEdges = append(parentEdges, e)
		if bytes.Equal(e.A, person) || bytes.Equal(e.B, person) {
			set.Add(kinship.Link{Edge: e})
		}
	}
	implied := kinship.Derive(edges, terms, rules)
	for _, imp := range implied {
		if bytes.Equal(imp.A, person) || bytes.Equal(imp.B, person) {
			set.AddImplied(imp)
		}
		if imp.Pattern == kinship.PatternBirthParents {
			parentEdges = append(parentEdges, imp.Edge)
		}
	}
	for _, imp := range kinship.Siblings(person, parentEdges, terms, rules) {
		set.AddImplied(imp)
	}
	return set.Links()
}

// withoutPeople drops the links whose other end is one of drop.
func withoutPeople(person []byte, links []kinship.Link, drop []byte) []kinship.Link {
	var out []kinship.Link
	for _, l := range links {
		if other, _, _ := l.From(person); !bytes.Contains(drop, other) {
			out = append(out, l)
		}
	}
	return out
}

// unitsOf groups person's family over edges as the page read does, with the
// people in drop not shown.
func unitsOf(t testing.TB, person byte, edges []kinship.Edge, drop ...byte) []kinship.Unit {
	t.Helper()
	read := func(id []byte) kinship.Kin {
		return kinship.KinOf(id, withoutPeople(id, linksFor(t, id, edges), dropIDs(drop)), terms)
	}
	kin := read(p(person))
	kinOf := map[string]kinship.Kin{}
	for _, role := range []kinship.Role{kinship.RoleParent, kinship.RoleChild} {
		for _, l := range kin[role] {
			other, _, _ := l.From(p(person))
			kinOf[string(other)] = read(other)
		}
	}
	return kinship.FamilyUnits(p(person), kin, kinOf)
}

func dropIDs(drop []byte) []byte {
	out := make([]byte, len(drop))
	for i, d := range drop {
		out[i] = 0x10 + d
	}
	return out
}

func unitName(k kinship.UnitKind) string {
	switch k {
	case kinship.UnitParents:
		return "Parents"
	case kinship.UnitPartner:
		return "Partner"
	case kinship.UnitOneParent:
		return "OneParent"
	case kinship.UnitSiblings:
		return "Siblings"
	}
	return fmt.Sprintf("Unit%d", k)
}

// show writes units as Kind{partners;children}, people as their numbers, each
// side sorted: partners and children are sets.
func show(units []kinship.Unit) []string {
	side := func(rs []kinship.Relative) string {
		var ns []int
		for _, r := range rs {
			ns = append(ns, int(r.ID[0]-0x10))
		}
		sort.Ints(ns)
		parts := make([]string, len(ns))
		for i, n := range ns {
			parts[i] = fmt.Sprint(n)
		}
		return strings.Join(parts, ",")
	}
	var out []string
	for _, u := range units {
		out = append(out, fmt.Sprintf("%s{%s;%s}", unitName(u.Kind), side(u.Partners), side(u.Children)))
	}
	return out
}

func relativeIn(t testing.TB, units []kinship.Unit, kind kinship.UnitKind, side string, id byte) kinship.Relative {
	t.Helper()
	for _, u := range units {
		if u.Kind != kind {
			continue
		}
		rs := u.Partners
		if side == "children" {
			rs = u.Children
		}
		for _, r := range rs {
			if bytes.Equal(r.ID, p(id)) {
				return r
			}
		}
	}
	t.Fatalf("no P%d in a %s %s", id, unitName(kind), side)
	return kinship.Relative{}
}

func shapesOf(l *kinship.Link) []string {
	var out []string
	for _, s := range l.Shapes {
		out = append(out, fmt.Sprintf("%s/%x", s.Pattern, s.Pivot))
	}
	return out
}

func TestKinOf(t *testing.T) {
	// One link of each of the eleven seeded relationship terms, filed from P1
	// to P2.
	terms11 := []vocabkey.BridgeTerm{
		vocabkey.TermParent, "child", vocabkey.TermSpouse, vocabkey.TermSibling,
		"cousin", vocabkey.TermGrandparent, "grandchild", "pibling", "nibling", "guardian", "ward",
	}
	type roles struct{ one, two []kinship.Role }
	want := map[vocabkey.BridgeTerm]roles{
		vocabkey.TermParent:  {one: []kinship.Role{kinship.RoleChild}, two: []kinship.Role{kinship.RoleParent}},
		"child":              {one: []kinship.Role{kinship.RoleParent}, two: []kinship.Role{kinship.RoleChild}},
		vocabkey.TermSpouse:  {one: []kinship.Role{kinship.RoleSpouse}, two: []kinship.Role{kinship.RoleSpouse}},
		vocabkey.TermSibling: {one: []kinship.Role{kinship.RoleSibling}, two: []kinship.Role{kinship.RoleSibling}},
	}
	if len(terms11) != 11 {
		t.Fatalf("%d terms", len(terms11))
	}
	for _, term := range terms11 {
		t.Run(string(term), func(t *testing.T) {
			l := kinship.Link{Edge: rel(t, 1, 2, term, 1)}
			for i, person := range []byte{1, 2} {
				kin := kinship.KinOf(p(person), []kinship.Link{l}, terms)
				w := want[term].one
				if i == 1 {
					w = want[term].two
				}
				var got []kinship.Role
				for _, role := range []kinship.Role{kinship.RoleParent, kinship.RoleChild, kinship.RoleSpouse, kinship.RoleSibling} {
					if len(kin[role]) > 0 {
						got = append(got, role)
					}
				}
				if !slices.Equal(got, w) {
					t.Fatalf("P%d reads %v as roles %v, want %v", person, term, got, w)
				}
			}
		})
	}
	t.Run("a link to self", func(t *testing.T) {
		l := kinship.Link{Edge: rel(t, 1, 1, vocabkey.TermParent, 1)}
		if kin := kinship.KinOf(p(1), []kinship.Link{l}, terms); len(kin) != 0 {
			t.Fatalf("got %v", kin)
		}
	})
	t.Run("a stranger", func(t *testing.T) {
		l := kinship.Link{Edge: rel(t, 1, 2, vocabkey.TermSpouse, 1)}
		if kin := kinship.KinOf(p(3), []kinship.Link{l}, terms); len(kin) != 0 {
			t.Fatalf("got %v", kin)
		}
	})
	t.Run("link order within a role", func(t *testing.T) {
		a := kinship.Link{Edge: rel(t, 2, 1, vocabkey.TermParent, 1)}
		b := kinship.Link{Edge: rel(t, 3, 1, vocabkey.TermParent, 2)}
		kin := kinship.KinOf(p(1), []kinship.Link{b, a}, terms)
		if !reflect.DeepEqual(kin[kinship.RoleParent], []kinship.Link{b, a}) {
			t.Fatalf("got %v", kin[kinship.RoleParent])
		}
	})
}

func TestFamilies(t *testing.T) {
	const (
		marriage = vocabkey.EventMarriage
		birth    = vocabkey.EventBirth
	)
	parent := func(g, m, br byte) kinship.Edge { return rel(t, g, m, vocabkey.TermParent, br) }
	spouse := func(a, c, br byte) kinship.Edge { return rel(t, a, c, vocabkey.TermSpouse, br) }
	sibling := func(a, c, br byte) kinship.Edge { return rel(t, a, c, vocabkey.TermSibling, br) }
	married := func(person, event, br byte) kinship.Edge {
		return part(t, person, event, vocabkey.RoleSubject, marriage, br)
	}
	born := func(person, event byte, role vocabkey.Role, br byte) kinship.Edge {
		return part(t, person, event, role, birth, br)
	}
	// johnMaryPat is birth 1: P6 father, P7 mother, P1 subject.
	johnMaryPat := []kinship.Edge{
		born(6, 1, vocabkey.RoleFather, 21), born(7, 1, vocabkey.RoleMother, 22), born(1, 1, vocabkey.RoleSubject, 23),
	}
	// quin is birth 2: the same parents, P9 subject.
	quin := []kinship.Edge{
		born(6, 2, vocabkey.RoleFather, 24), born(7, 2, vocabkey.RoleMother, 25), born(9, 2, vocabkey.RoleSubject, 26),
	}

	cases := []struct {
		name  string
		edges []kinship.Edge
		drop  []byte
		want  []string
	}{
		{"AC-1.1", []kinship.Edge{
			married(1, 1, 1), married(2, 1, 2), married(1, 2, 3), married(3, 2, 4),
			parent(1, 4, 5), parent(2, 4, 6),
			born(1, 3, vocabkey.RoleFather, 7), born(3, 3, vocabkey.RoleMother, 8), born(5, 3, vocabkey.RoleSubject, 9),
		}, nil, []string{"Partner{2;4}", "Partner{3;5}"}},
		{"AC-1.2", append(slices.Clone(johnMaryPat), parent(8, 1, 1)), nil, []string{"Parents{8;}", "Parents{6,7;}"}},
		{"AC-1.3 with a spouse link", []kinship.Edge{parent(6, 1, 1), parent(7, 1, 2), spouse(6, 7, 3)}, nil, []string{"Parents{6,7;}"}},
		{"AC-1.3 without it", []kinship.Edge{parent(6, 1, 1), parent(7, 1, 2)}, nil, []string{"Parents{6;}", "Parents{7;}"}},
		{"AC-1.4", []kinship.Edge{spouse(1, 2, 1), married(1, 1, 2), married(2, 1, 3)}, nil, []string{"Partner{2;}"}},
		{"AC-1.5", slices.Concat(johnMaryPat, quin, []kinship.Edge{parent(7, 10, 1)}), nil,
			[]string{"Parents{6,7;9}", "Parents{7;10}"}},
		{"AC-1.6", slices.Concat(johnMaryPat, quin, []kinship.Edge{sibling(1, 11, 1), sibling(1, 9, 2)}), nil,
			[]string{"Parents{6,7;9}", "Siblings{;11}"}},
		{"AC-1.7", []kinship.Edge{parent(1, 12, 1), parent(13, 12, 2)}, nil, []string{"Partner{13;12}"}},
		{"AC-1.8", []kinship.Edge{parent(1, 14, 1), spouse(1, 2, 2)}, nil, []string{"Partner{2;}", "OneParent{;14}"}},
		{"AC-1.10", []kinship.Edge{parent(6, 1, 1), spouse(1, 6, 2)}, nil, []string{"Parents{6;}", "Partner{6;}"}},
		{"AC-1.11", []kinship.Edge{
			rel(t, 8, 1, vocabkey.TermGrandparent, 1), rel(t, 1, 9, "cousin", 2), rel(t, 10, 1, "guardian", 3),
			parent(6, 7, 4), parent(7, 1, 5),
		}, nil, []string{"Parents{7;}"}},
		{"a parent not shown, a sibling only through them", []kinship.Edge{parent(6, 1, 1), parent(6, 9, 2)}, []byte{6},
			[]string{"Siblings{;9}"}},
		{"a child whose only other parent is not shown", []kinship.Edge{parent(1, 4, 1), parent(2, 4, 2)}, []byte{2},
			[]string{"OneParent{;4}"}},
		{"a child with two other parents", []kinship.Edge{parent(1, 4, 1), parent(2, 4, 2), parent(3, 4, 3)}, nil,
			[]string{"Partner{2;4}", "Partner{3;4}"}},
		{"a birth with two fathers and a mother", []kinship.Edge{
			born(6, 1, vocabkey.RoleFather, 1), born(7, 1, vocabkey.RoleFather, 2), born(8, 1, vocabkey.RoleMother, 3), born(1, 1, vocabkey.RoleSubject, 4),
		}, nil, []string{"Parents{6,7;}", "Parents{6,8;}", "Parents{7,8;}"}},
		{"ten recorded parents with no pairing evidence", func() []kinship.Edge {
			var es []kinship.Edge
			for n := byte(2); n <= 11; n++ {
				es = append(es, parent(n, 1, n))
			}
			return es
		}(), nil, func() []string {
			var out []string
			for n := 2; n <= 11; n++ {
				out = append(out, fmt.Sprintf("Parents{%d;}", n))
			}
			return out
		}()},
		{"empty kin", nil, nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := show(unitsOf(t, 1, tc.edges, tc.drop...))
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got  %v\nwant %v", got, tc.want)
			}
		})
	}

	t.Run("links say why", func(t *testing.T) {
		// AC-1.1: Sam is implied through the marriage; Uma is recorded.
		units := unitsOf(t, 1, []kinship.Edge{
			married(1, 1, 1), married(2, 1, 2), parent(1, 4, 5), parent(2, 4, 6),
		})
		sam := relativeIn(t, units, kinship.UnitPartner, "partners", 2)
		if sam.Link == nil || !sam.Link.OnlyImplied || !slices.Equal(shapesOf(sam.Link), []string{"marriage_spouses/21"}) {
			t.Fatalf("Sam's link %+v", sam.Link)
		}
		uma := relativeIn(t, units, kinship.UnitPartner, "children", 4)
		if uma.Link == nil || uma.Link.OnlyImplied || !reflect.DeepEqual(uma.Link.Associations, [][]byte{b(5)}) {
			t.Fatalf("Uma's link %+v", uma.Link)
		}
		// AC-1.4: a recorded spouse link also implied by the marriage.
		units = unitsOf(t, 1, []kinship.Edge{spouse(1, 2, 1), married(1, 1, 2), married(2, 1, 3)})
		sam = relativeIn(t, units, kinship.UnitPartner, "partners", 2)
		if sam.Link == nil || sam.Link.OnlyImplied || !slices.Equal(shapesOf(sam.Link), []string{"marriage_spouses/21"}) || !slices.ContainsFunc(sam.Link.Associations, func(a []byte) bool { return bytes.Equal(a, b(1)) }) {
			t.Fatalf("Sam's recorded link %+v", sam.Link)
		}
		// AC-1.5: Quin's link has a shared_parent shape through each parent, Rae's
		// through one.
		units = unitsOf(t, 1, slices.Concat(johnMaryPat, quin, []kinship.Edge{parent(7, 10, 1)}))
		q := relativeIn(t, units, kinship.UnitParents, "children", 9)
		if !slices.Equal(shapesOf(q.Link), []string{"shared_parent/16", "shared_parent/17"}) {
			t.Fatalf("Quin's shapes %v", shapesOf(q.Link))
		}
		r := relativeIn(t, units, kinship.UnitParents, "children", 10)
		if !slices.Equal(shapesOf(r.Link), []string{"shared_parent/17"}) {
			t.Fatalf("Rae's shapes %v", shapesOf(r.Link))
		}
		// AC-1.6: a recorded sibling who also shares parents is recorded and implied.
		units = unitsOf(t, 1, slices.Concat(johnMaryPat, quin, []kinship.Edge{sibling(1, 9, 2)}))
		q = relativeIn(t, units, kinship.UnitParents, "children", 9)
		if q.Link.OnlyImplied || len(q.Link.Shapes) == 0 {
			t.Fatalf("Quin's link %+v", q.Link)
		}
		// AC-1.7: a co-parent known only through a child has no link.
		units = unitsOf(t, 1, []kinship.Edge{parent(1, 12, 1), parent(13, 12, 2)})
		if xan := relativeIn(t, units, kinship.UnitPartner, "partners", 13); xan.Link != nil {
			t.Fatalf("Xan has link %+v", xan.Link)
		}
		if wren := relativeIn(t, units, kinship.UnitPartner, "children", 12); wren.Link == nil {
			t.Fatal("Wren has no link")
		}
	})
}
