package searchindex

import "github.com/mendahu/provenencia/core/database/graphcache"

// SetNewGraphForTest replaces the graph builder ReprojectHandles uses and
// returns a function that restores it.
func SetNewGraphForTest(fn func(graphcache.Querier) *graphcache.Graph) (restore func()) {
	old := newGraph
	newGraph = fn
	return func() { newGraph = old }
}
