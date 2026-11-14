package promotealign

import (
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/shortlist"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// HandlesHeldForTest is how many handle profiles the graph's match index holds,
// 0 when it holds none.
func HandlesHeldForTest(g *graphcache.Graph) int {
	m, ok := MatchIndex.Peek(g)
	if !ok {
		return 0
	}
	return len(m.handles)
}

// HandleForTest is the held Subject of a handle, and false when the graph holds
// no match index or the index has no such handle.
func HandleForTest(g *graphcache.Graph, id []byte) (subjectcompare.Subject, bool) {
	m, ok := MatchIndex.Peek(g)
	if !ok {
		return subjectcompare.Subject{}, false
	}
	s, ok := m.handles[string(id)]
	return s, ok
}

// IndexedForTest is the held shortlist index's Lookup of probe, nil when the
// graph holds no match index.
func IndexedForTest(g *graphcache.Graph, kind string, probe subjectcompare.Values) [][]byte {
	m, ok := MatchIndex.Peek(g)
	if !ok {
		return nil
	}
	return m.shortlist.Lookup(vocabkey.SubjectType(kind), probe)
}

// EmptyShortlistForTest replaces the stored shortlist index with one that
// holds no handle, so Verify can fail a test.
func EmptyShortlistForTest(c *database.Catalog) error {
	g := c.Graph()
	p, err := newProfiler(g)
	if err != nil {
		return err
	}
	m, err := matchIndexOf(g)
	if err != nil {
		return err
	}
	MatchIndex.Set(g, &matchIndex{
		handles:   m.handles,
		shortlist: shortlist.NewIndex(p.cfg.Alignment.Shortlist, p.cfg.Subject, p.cfg.Value, p.metas),
	})
	return nil
}

// CandidatesHeldForTest reports whether the candidates entry holds the graph.
func CandidatesHeldForTest(g *graphcache.Graph, graphID []byte) bool {
	_, ok := candidates.Peek(g, graphID)
	return ok
}

// HoldCandidatesForTest stores an empty candidates memo for the graph.
func HoldCandidatesForTest(g *graphcache.Graph, graphID []byte) {
	candidates.Set(g, graphID, candidatesMemo{})
}

// SetCandidatesCapForTest overrides the candidates entry's Cap on g.
func SetCandidatesCapForTest(g *graphcache.Graph, n int) {
	candidates.SetCapForTest(g, n)
}
