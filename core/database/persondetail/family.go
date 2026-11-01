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

// PersonFamily is the page read: one handle's family and the summary of every
// person its units name. An unknown or merged handle, or one that is no Person,
// is conclusiondetails.ErrNotFound. The family and the summaries agree: a
// relative with no summary is dropped, and a unit left empty with it.
func PersonFamily(c *database.Catalog, entityID []byte) (Family, conclusionsummaries.Summaries, error) {
	if _, err := conclusiondetails.Handle(c, entityID); err != nil {
		return Family{}, conclusionsummaries.Summaries{}, err
	}
	families, err := Families(c, [][]byte{entityID})
	if err != nil {
		return Family{}, conclusionsummaries.Summaries{}, err
	}
	if len(families) == 0 {
		return Family{}, conclusionsummaries.Summaries{}, conclusiondetails.ErrNotFound
	}
	family := families[0]
	var want conclusionsummaries.Want
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
		return Family{}, conclusionsummaries.Summaries{}, err
	}
	return dropUnread(family, summaries), summaries, nil
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
