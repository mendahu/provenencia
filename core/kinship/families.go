package kinship

import (
	"bytes"
	"slices"

	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Role is a nuclear-family role of a relative, read from the person's end.
type Role int

const (
	RoleParent Role = iota + 1
	RoleChild
	RoleSpouse
	RoleSibling
)

// roleRow says which relationship term puts a link in a role, and which end of
// the term the person stands at. For an undirected term both ends read alike
// and the person may stand at either.
type roleRow struct {
	role Role
	term vocabkey.BridgeTerm
	at   termEnd
}

type termEnd int

const (
	atFirst termEnd = iota + 1
	atSecond
)

// nuclearRoles are the roles KinOf reads, in output order. A grandparent role
// is one more row.
var nuclearRoles = []roleRow{
	{RoleParent, vocabkey.TermParent, atSecond},
	{RoleChild, vocabkey.TermParent, atFirst},
	{RoleSpouse, vocabkey.TermSpouse, atFirst},
	{RoleSibling, vocabkey.TermSibling, atFirst},
}

// Roles returns the roles KinOf reads, in output order: the one list of them.
func Roles() []Role {
	roles := make([]Role, len(nuclearRoles))
	for i, row := range nuclearRoles {
		roles[i] = row.role
	}
	return roles
}

// Kin is a person's links by role, each in link order.
type Kin map[Role][]Link

// Around returns the distinct ids of person's parents then person's children, in
// order: the people whose own kin FamilyUnits reads.
func (k Kin) Around(person []byte) [][]byte {
	var ids [][]byte
	seen := map[string]bool{}
	for _, role := range []Role{RoleParent, RoleChild} {
		for _, l := range k[role] {
			other, _, _ := l.From(person)
			if !seen[string(other)] {
				seen[string(other)] = true
				ids = append(ids, other)
			}
		}
	}
	return ids
}

// KinOf reads person's links through nuclearRoles: each link is in every role
// whose key it carries from person's end. Any other link is in no role.
func KinOf(person []byte, links []Link, terms segmentcompare.Terms) Kin {
	kin := Kin{}
	for _, row := range nuclearRoles {
		keyFirst, keySecond, ok := relationshipKeys(row.term, terms)
		if !ok {
			continue
		}
		undirected := keyFirst == keySecond
		for _, l := range links {
			first, second, ok := readsTerm(l.Edge, row.term, terms)
			if !ok {
				continue
			}
			stands := first
			if row.at == atSecond {
				stands = second
			}
			if bytes.Equal(person, stands) || undirected && (bytes.Equal(person, first) || bytes.Equal(person, second)) {
				kin[row.role] = append(kin[row.role], l)
			}
		}
	}
	return kin
}

// UnitKind is what a family unit is to the person.
type UnitKind int

const (
	UnitParents   UnitKind = iota + 1 // Partners: a set of person's parents; Children: siblings through each of them
	UnitPartner                       // Partners: one partner; Children: person's children with them
	UnitOneParent                     // Children: person's children with no other known parent
	UnitSiblings                      // Children: siblings with no shown shared parent
)

// Relative is one person in a unit and person's own link to them; Link is nil
// for a co-parent known only through a child.
type Relative struct {
	ID   []byte
	Link *Link
}

// Unit is one family unit of a person.
type Unit struct {
	Kind     UnitKind
	Partners []Relative
	Children []Relative
}

// Empty reports whether the unit names nobody it is about: a parents or partner
// unit with no partner, a one parent or siblings unit with no child. A kind it
// does not know is empty.
func (u Unit) Empty() bool {
	switch u.Kind {
	case UnitParents, UnitPartner:
		return len(u.Partners) == 0
	case UnitOneParent, UnitSiblings:
		return len(u.Children) == 0
	}
	return true
}

// relativeOf is the relative a person's link names, with the link.
func relativeOf(person []byte, l Link) Relative {
	other, _, _ := l.From(person)
	return Relative{ID: other, Link: &l}
}

// FamilyUnits groups person's kin into units: the parent sets with the
// siblings through them, each partner with the children they share, the
// children with no other known parent, and the siblings with no shown shared
// parent, in that order. Every Relative.Link is person's own link to that
// relative. kinOf holds the kin of each of person's parents and children, keyed
// by string(id); a missing entry has none.
func FamilyUnits(person []byte, kin Kin, kinOf map[string]Kin, rules Rules) []Unit {
	units, siblings := parentUnits(person, kin, kinOf, rules)
	partners, oneParent := partnerUnits(person, kin, kinOf)
	units = append(units, partners...)
	for _, u := range []Unit{oneParent, siblings} {
		if !u.Empty() {
			units = append(units, u)
		}
	}
	return units
}

// parentUnits makes the parent sets and the siblings through them, and returns
// apart the unit of siblings with no shown shared parent.
//
// More parents than rules.MaxPivotEnd read as malformed data: none is paired
// with another, and a sibling is placed once, with the first parent it shares
// (in parent order), or in the siblings unit when it shares none. That keeps
// what is shown linear in the data.
func parentUnits(person []byte, kin Kin, kinOf map[string]Kin, rules Rules) ([]Unit, Unit) {
	parents := kin[RoleParent]
	ids := make([][]byte, len(parents))
	for i, l := range parents {
		ids[i], _, _ = l.From(person)
	}
	// paired: two parents are one set when either files a spouse link to the
	// other, or the person's links to both carry a birth_parents shape with the
	// same pivot.
	files := func(a, c []byte) bool { // a files a spouse link to c
		for _, s := range kinOf[string(a)][RoleSpouse] {
			if other, _, ok := s.From(a); ok && bytes.Equal(other, c) {
				return true
			}
		}
		return false
	}
	paired := func(i, j int) bool {
		if files(ids[i], ids[j]) || files(ids[j], ids[i]) {
			return true
		}
		for _, x := range parents[i].Shapes {
			if x.Pattern != PatternBirthParents {
				continue
			}
			for _, y := range parents[j].Shapes {
				if y.Pattern == PatternBirthParents && bytes.Equal(x.Pivot, y.Pivot) {
					return true
				}
			}
		}
		return false
	}
	var sets [][]int
	inPair := make([]bool, len(parents))
	pairing := len(parents) <= rules.MaxPivotEnd
	for i := range parents {
		for j := i + 1; pairing && j < len(parents); j++ {
			if paired(i, j) {
				sets = append(sets, []int{i, j})
				inPair[i], inPair[j] = true, true
			}
		}
		if !inPair[i] {
			sets = append(sets, []int{i})
		}
	}
	// A lone parent is emitted at its own place, after the pairs at lower i that
	// name it, so inPair for i is complete when i is reached.
	units := make([]Unit, len(sets))
	for k, set := range sets {
		units[k].Kind = UnitParents
		for _, i := range set {
			units[k].Partners = append(units[k].Partners, relativeOf(person, parents[i]))
		}
	}

	var unplaced Unit
	unplaced.Kind = UnitSiblings
	var extra []Unit
	var extraAt [][]int
	for _, l := range kin[RoleSibling] {
		sib := relativeOf(person, l)
		var shared []int // indices into parents, in parent order
		for i := range parents {
			if slices.ContainsFunc(l.Shapes, func(s Shape) bool {
				return s.Pattern == PatternSharedParent && bytes.Equal(s.Pivot, ids[i])
			}) {
				shared = append(shared, i)
			}
		}
		joined := false
		if !pairing {
			// Every set is one parent, at that parent's index.
			if len(shared) > 0 {
				units[shared[0]].Children = append(units[shared[0]].Children, sib)
				joined = true
			}
		} else {
			for k, set := range sets {
				if len(set) > 0 && subset(set, shared) {
					units[k].Children = append(units[k].Children, sib)
					joined = true
				}
			}
		}
		switch {
		case joined:
		case len(shared) == 0:
			unplaced.Children = append(unplaced.Children, sib)
		default:
			at := slices.IndexFunc(extraAt, func(x []int) bool { return slices.Equal(x, shared) })
			if at < 0 {
				at = len(extra)
				extraAt = append(extraAt, shared)
				u := Unit{Kind: UnitParents}
				for _, i := range shared {
					u.Partners = append(u.Partners, relativeOf(person, parents[i]))
				}
				extra = append(extra, u)
			}
			extra[at].Children = append(extra[at].Children, sib)
		}
	}
	return append(units, extra...), unplaced
}

// subset reports whether every index of set is in of.
func subset(set, of []int) bool {
	for _, i := range set {
		if !slices.Contains(of, i) {
			return false
		}
	}
	return true
}

// partnerUnits makes one unit per partner, and the unit of children with no
// other known parent.
func partnerUnits(person []byte, kin Kin, kinOf map[string]Kin) (partners []Unit, oneParent Unit) {
	oneParent.Kind = UnitOneParent
	at := map[string]int{}
	for _, l := range kin[RoleSpouse] {
		r := relativeOf(person, l)
		if _, dup := at[string(r.ID)]; dup {
			continue
		}
		at[string(r.ID)] = len(partners)
		partners = append(partners, Unit{Kind: UnitPartner, Partners: []Relative{r}})
	}
	for _, l := range kin[RoleChild] {
		child := relativeOf(person, l)
		others := 0
		for _, pl := range kinOf[string(child.ID)][RoleParent] {
			other, _, ok := pl.From(child.ID)
			if !ok || bytes.Equal(other, person) {
				continue
			}
			others++
			i, known := at[string(other)]
			if !known {
				i = len(partners)
				at[string(other)] = i
				partners = append(partners, Unit{Kind: UnitPartner, Partners: []Relative{{ID: other}}})
			}
			if !slices.ContainsFunc(partners[i].Children, func(c Relative) bool { return bytes.Equal(c.ID, child.ID) }) {
				partners[i].Children = append(partners[i].Children, child)
			}
		}
		if others == 0 {
			oneParent.Children = append(oneParent.Children, child)
		}
	}
	return partners, oneParent
}
