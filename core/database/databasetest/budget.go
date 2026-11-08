//go:build cgo

package databasetest

import (
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/database/graphcache"
)

// Fastest is the least of n timed runs of run; each run's own setup stays
// outside the time it returns.
func Fastest(n int, run func() time.Duration) time.Duration {
	best := time.Duration(-1)
	for i := 0; i < n; i++ {
		if d := run(); best < 0 || d < best {
			best = d
		}
	}
	return best
}

// Since times one call.
func Since(call func()) time.Duration {
	start := time.Now()
	call()
	return time.Since(start)
}

// WithinBudget logs what took got and fails the test when it is over budget.
// Verify mode rebuilds the graph cache on every commit, so budgets are
// measured without it and the check is skipped there.
func WithinBudget(t testing.TB, what string, got, budget time.Duration) {
	t.Helper()
	t.Logf("%s: %v (budget %v)", what, got, budget)
	if graphcache.VerifyMode() {
		return
	}
	if got > budget {
		t.Errorf("%s took %v, over the %v budget", what, got, budget)
	}
}
