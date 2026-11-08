package searchindex

import "github.com/mendahu/provenencia/core/database/graphcache"

// SetNewGraphForTest replaces the graph builder ReprojectHandles uses and
// returns a function that restores it.
func SetNewGraphForTest(fn func(graphcache.Querier) *graphcache.Graph) (restore func()) {
	old := newGraph
	newGraph = fn
	return func() { newGraph = old }
}

// DependentsOfForTest is dependentsOf.
func DependentsOfForTest(q Querier, keys []graphcache.InputKey) ([][]byte, error) {
	return dependentsOf(q, keys)
}

// DependentsOfHandlesForTest is dependentsOfHandles.
func DependentsOfHandlesForTest(q Querier, ids [][]byte) ([][]byte, error) {
	return dependentsOfHandles(q, ids)
}

// OldEndsForTest is oldEnds.
func OldEndsForTest(q Querier, ids [][]byte) (map[string][]graphcache.InputKey, error) {
	return oldEnds(q, ids)
}
