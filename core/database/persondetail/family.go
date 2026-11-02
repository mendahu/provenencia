package persondetail

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonlinks"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// Family is one person's family units.
type Family struct {
	Person []byte
	Units  []kinship.Unit
}

// shownPerson is the package's one canonical-relative rule: an unmerged Person
// with at least one accepted member. A section that names other people calls it
// before it groups them.
func shownPerson(n *graphcache.Node, personTypeID []byte) bool {
	return n.Is(personTypeID) && len(n.AcceptedMembers()) > 0
}

// personTypeID is the id of the provenencia Person type.
func personTypeID(g *graphcache.Graph) ([]byte, error) {
	v, err := g.Vocabulary()
	if err != nil {
		return nil, err
	}
	return v.TypeIDsByKey()[vocabkey.Person], nil
}

// Families reads the family of each id that is an unmerged Person, in ids
// order, from the graph cache: canonical relatives only, a fixed number of
// batched loads cold, no query warm. An id that is not a Person, is merged, or
// is not 16 bytes is skipped, and an id repeated gives one Family.
func Families(c *database.Catalog, ids [][]byte) ([]Family, error) {
	if c == nil || c.Graph() == nil {
		return nil, nil
	}
	g := c.Graph()
	rules := matchregistry.Default().Kinship
	r, err := canonlinks.NewReader(g, rules)
	if err != nil {
		return nil, err
	}
	terms, err := canonlinks.LoadTerms(g)
	if err != nil {
		return nil, err
	}
	typeID, err := personTypeID(g)
	if err != nil {
		return nil, err
	}
	var want [][]byte
	seen := map[string]bool{}
	for _, id := range ids {
		if len(id) == 16 && !seen[string(id)] {
			seen[string(id)] = true
			want = append(want, id)
		}
	}
	if err := r.LoadRelationships(want, canonlinks.Options{Siblings: true}); err != nil {
		return nil, err
	}

	// The unfiltered kin of each person that is read, and every relative they name.
	var people [][]byte
	var relatives [][]byte
	named := map[string]bool{}
	for _, id := range want {
		n, _ := g.Cached(id)
		if !n.Is(typeID) {
			continue
		}
		people = append(people, id)
		links, err := r.Relationships(id, canonlinks.Options{Siblings: true})
		if err != nil {
			return nil, err
		}
		kin := kinship.KinOf(id, links, terms)
		for _, role := range kinship.Roles() {
			for _, l := range kin[role] {
				if other, _, _ := l.From(id); !named[string(other)] {
					named[string(other)] = true
					relatives = append(relatives, other)
				}
			}
		}
	}
	if err := r.LoadRelationships(relatives, canonlinks.Options{}); err != nil {
		return nil, err
	}

	// kinOf reads id's relationship links, drops those to people who are not
	// shown, and reads the kin of what is left.
	kinOf := func(id []byte, opts canonlinks.Options) (kinship.Kin, error) {
		links, err := r.Relationships(id, opts)
		if err != nil {
			return nil, err
		}
		var shown []kinship.Link
		for _, l := range links {
			other, _, _ := l.From(id)
			if n, _ := g.Cached(other); shownPerson(n, typeID) {
				shown = append(shown, l)
			}
		}
		return kinship.KinOf(id, shown, terms), nil
	}
	out := make([]Family, 0, len(people))
	for _, id := range people {
		kin, err := kinOf(id, canonlinks.Options{Siblings: true})
		if err != nil {
			return nil, err
		}
		around := map[string]kinship.Kin{}
		for _, other := range kin.Around(id) {
			if around[string(other)], err = kinOf(other, canonlinks.Options{}); err != nil {
				return nil, err
			}
		}
		out = append(out, Family{Person: id, Units: kinship.FamilyUnits(id, kin, around, rules)})
	}
	return out, nil
}

// OriginSibling is one child of the person's parents, or a sibling recorded
// with no shared parent. Relative carries the person's own link to them (nil
// for the person); ParentSets holds the indexes, among the person's parents
// units, of the parent sets they are a child of (empty for a sibling recorded
// only).
type OriginSibling struct {
	Relative   kinship.Relative
	ParentSets []int
}

// PersonFamily is the page read: one handle's family, the summary of every
// person its units name and of the person, and the origin siblings. An unknown
// or merged handle, or one that is no Person, is conclusiondetails.ErrNotFound.
// The family and the summaries agree: a relative with no summary is dropped,
// and a unit left empty with it. Each unit's children are in birth order,
// undated last. The origin siblings are every child of the parents and siblings
// units once, with the person among them, in the order the page draws them
// (see originSiblings); the person is left out only when its summary was not
// read.
func PersonFamily(c *database.Catalog, entityID []byte) (Family, []OriginSibling, conclusionsummaries.Summaries, error) {
	if _, err := conclusiondetails.Handle(c, entityID); err != nil {
		return Family{}, nil, conclusionsummaries.Summaries{}, err
	}
	families, err := Families(c, [][]byte{entityID})
	if err != nil {
		return Family{}, nil, conclusionsummaries.Summaries{}, err
	}
	if len(families) == 0 {
		return Family{}, nil, conclusionsummaries.Summaries{}, conclusiondetails.ErrNotFound
	}
	family := families[0]
	var want conclusionsummaries.Want
	want.Add(vocabkey.Person, entityID)
	for _, u := range family.Units {
		for _, r := range u.Partners {
			want.Add(vocabkey.Person, r.ID)
		}
		for _, r := range u.Children {
			want.Add(vocabkey.Person, r.ID)
		}
	}
	summaries, err := conclusionsummaries.ByIDs(c, want)
	if err != nil {
		return Family{}, nil, conclusionsummaries.Summaries{}, err
	}
	family = dropUnread(family, summaries)
	byBirth := func(rs []kinship.Relative) {
		conclusionsummaries.SortByBirth(rs, func(r kinship.Relative) conclusionsummaries.PersonSummary {
			h, _ := summaries.Person(r.ID)
			return h
		})
	}
	for _, u := range family.Units {
		byBirth(u.Children)
	}
	return family, originSiblings(family, byBirth, summaries), summaries, nil
}

// originSiblings lists every child of the person's parents units and siblings
// unit once, with the person among them, in the order the page draws them: the
// siblings in every parent set; then, walking the parent sets from last to
// first, each set's siblings not yet placed; then the siblings in no parent
// set. Each group is in birth order. The person is a child of every parent set.
func originSiblings(family Family, byBirth func([]kinship.Relative), summaries conclusionsummaries.Summaries) []OriginSibling {
	var sets int
	byID := map[string]*OriginSibling{}
	var all []*OriginSibling
	add := func(r kinship.Relative) *OriginSibling {
		if o, ok := byID[string(r.ID)]; ok {
			return o
		}
		o := &OriginSibling{Relative: r}
		byID[string(r.ID)] = o
		all = append(all, o)
		return o
	}
	for _, u := range family.Units {
		switch u.Kind {
		case kinship.UnitParents:
			for _, r := range u.Children {
				o := add(r)
				o.ParentSets = append(o.ParentSets, sets)
			}
			sets++
		case kinship.UnitSiblings:
			for _, r := range u.Children {
				add(r)
			}
		}
	}
	if _, ok := summaries.Person(family.Person); ok {
		o := add(kinship.Relative{ID: family.Person})
		for k := 0; k < sets; k++ {
			o.ParentSets = append(o.ParentSets, k)
		}
	}
	// One pass buckets each entry: the full bucket, else the bucket of its
	// largest parent set (the walk from last to first places it at the first
	// set it is in), else the recorded bucket.
	full := make([]kinship.Relative, 0)
	perSet := make([][]kinship.Relative, sets)
	var recorded []kinship.Relative
	for _, o := range all {
		switch {
		case len(o.ParentSets) == 0:
			recorded = append(recorded, o.Relative)
		case len(o.ParentSets) == sets:
			full = append(full, o.Relative)
		default:
			k := o.ParentSets[len(o.ParentSets)-1]
			perSet[k] = append(perSet[k], o.Relative)
		}
	}
	var out []OriginSibling
	emit := func(group []kinship.Relative) {
		byBirth(group)
		for _, r := range group {
			out = append(out, *byID[string(r.ID)])
		}
	}
	emit(full)
	for k := sets - 1; k >= 0; k-- {
		emit(perSet[k])
	}
	emit(recorded)
	return out
}

// dropUnread drops each relative whose summary was not read (a node dropped
// mid-read: the next read is right, as memberGraphs skips) and then each unit
// that is left empty: a parents or partner unit with no partner, and a one
// parent or siblings unit with no child.
func dropUnread(family Family, summaries conclusionsummaries.Summaries) Family {
	read := func(rs []kinship.Relative) []kinship.Relative {
		var kept []kinship.Relative
		for _, r := range rs {
			if _, ok := summaries.Person(r.ID); ok {
				kept = append(kept, r)
			}
		}
		return kept
	}
	var units []kinship.Unit
	for _, u := range family.Units {
		u.Partners, u.Children = read(u.Partners), read(u.Children)
		if u.Empty() {
			continue
		}
		units = append(units, u)
	}
	family.Units = units
	return family
}
