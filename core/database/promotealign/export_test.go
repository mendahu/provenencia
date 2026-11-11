package promotealign

import (
	"slices"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/shortlist"
)

// ProfiledRefsForTest proposes for the graph with no fixed rows and returns,
// sorted, the Ref of every handle whose full profile was built.
func ProfiledRefsForTest(c *database.Catalog, graphID []byte) ([]string, error) {
	_, _, prof, err := propose(c, graphID, nil, nil)
	if err != nil || prof == nil {
		return nil, err
	}
	var refs []string
	for id := range prof.done {
		n, err := c.Graph().Node([]byte(id))
		if err != nil {
			return nil, err
		}
		if n != nil {
			refs = append(refs, n.Ref)
		}
	}
	slices.Sort(refs)
	return refs, nil
}

// ProfilesKeptForTest is how many handle profiles the graph's match index
// holds, and false when the graph has no match index.
func ProfilesKeptForTest(c *database.Catalog) (int, bool) {
	if _, ok, _ := MatchIndex.Get(c.Graph()); !ok {
		return 0, false
	}
	return len(matchIndexOf(c.Graph()).profiles), true
}

// MatchIndexGenForTest is the generation of the stored shortlist index, 0 when
// the graph has none.
func MatchIndexGenForTest(g *graphcache.Graph) uint64 {
	return matchIndexOf(g).gen
}

// EmptyShortlistForTest replaces the stored shortlist index with one that
// holds no handle, so Verify can fail a test.
func EmptyShortlistForTest(c *database.Catalog) error {
	g := c.Graph()
	cfg, _, p, err := shortlistInputs(g, false)
	if err != nil {
		return err
	}
	m := matchIndexOf(g)
	m.shortlist = shortlist.NewIndex(cfg.Alignment.Shortlist, cfg.Subject, cfg.Value, p.metas)
	MatchIndex.Set(g, m)
	return nil
}
