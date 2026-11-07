package promotealign

import (
	"slices"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/valuerules"
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

// MemoSnapshot is what the graph holds as promotealign's Evidence graph memos.
type MemoSnapshot struct {
	Layer evidencelayer.Layer
}

// MemoSnapshotForTest reads graphID's layer through evidencelayer.Load; it
// builds the layer on a miss, so call it after a Propose.
func MemoSnapshotForTest(c *database.Catalog, graphID []byte) MemoSnapshot {
	db, err := c.DB()
	if err != nil {
		return MemoSnapshot{}
	}
	_, namePattern, err := valuerules.Load(c.Graph(), db)
	if err != nil {
		return MemoSnapshot{}
	}
	layer, err := evidencelayer.Load(c.Graph(), graphID, namePattern)
	if err != nil {
		return MemoSnapshot{}
	}
	return MemoSnapshot{Layer: layer}
}

// ProfilesKeptForTest is how many handle profiles the graph's match index
// holds, and false when the graph has no match index.
func ProfilesKeptForTest(c *database.Catalog) (int, bool) {
	if _, ok := c.Graph().MatchIndex(); !ok {
		return 0, false
	}
	return len(matchIndexOf(c.Graph()).profiles), true
}

// MatchIndexGenForTest is the generation of the stored shortlist index, 0 when
// the graph has none.
func MatchIndexGenForTest(g *graphcache.Graph) uint64 {
	return matchIndexOf(g).gen
}
