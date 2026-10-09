package promotealign

import (
	"sync"

	"github.com/mendahu/provenencia/core/database/matching"
	"github.com/mendahu/provenencia/core/match"
)

// Every handle of a kind, with its profile values, is what Rank seeds the
// canon from. Reading it is a scan of the kind, and Promote proposes again on
// each decision without writing, so the scan is cached per catalog file and
// audit revision like the stats. A write misses the cache.
var (
	candMu     sync.Mutex
	candKey    statsStamp
	candCached map[string][]match.Candidate
)

// candidatesOfType is matching.CandidatesOfType, cached by catalog stamp.
// The slices are shared between proposals: callers only read them.
func candidatesOfType(q Querier, kind string, profile match.Profile) ([]match.Candidate, error) {
	stamp, err := catalogStamp(q)
	if err != nil {
		return nil, err
	}
	candMu.Lock()
	defer candMu.Unlock()
	// An in-memory catalog has no file, so it can't be told apart: never cache it.
	if stamp.file == "" || stamp != candKey {
		candKey, candCached = stamp, map[string][]match.Candidate{}
	}
	if cands, ok := candCached[kind]; ok && stamp.file != "" {
		return cands, nil
	}
	cands, err := matching.CandidatesOfType(q, kind, "provenencia", profile)
	if err != nil {
		return nil, err
	}
	if stamp.file != "" {
		candCached[kind] = cands
	}
	return cands, nil
}
