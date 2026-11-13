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
// holds no handle, so Verify can fail a test. Like any index loadShortlist
// builds, it has a new generation, so candidates ranked against the old one are
// stale and Verify leaves them alone.
func EmptyShortlistForTest(c *database.Catalog) error {
	g := c.Graph()
	cfg, _, p, err := shortlistInputs(g, false)
	if err != nil {
		return err
	}
	m := matchIndexOf(g)
	m.shortlist = shortlist.NewIndex(cfg.Alignment.Shortlist, cfg.Subject, cfg.Value, p.metas)
	m.gen = indexGen.Add(1)
	MatchIndex.Set(g, m)
	return nil
}

// CandidatesHeldForTest reports whether the candidates entry holds the graph.
func CandidatesHeldForTest(g *graphcache.Graph, graphID []byte) bool {
	_, ok := candidates.Peek(g, graphID)
	return ok
}

// HoldCandidatesForTest stores an empty candidates memo for the graph.
func HoldCandidatesForTest(g *graphcache.Graph, graphID []byte) {
	candidates.Set(g, graphID, candidatesMemo{gen: matchIndexOf(g).gen})
}

// SetCandidatesCapForTest overrides the candidates entry's Cap on g.
func SetCandidatesCapForTest(g *graphcache.Graph, n int) {
	candidates.SetCapForTest(g, n)
}
