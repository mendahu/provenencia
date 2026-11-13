package persondetail_test

import (
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database/databasetest"
)

// Budgets of S16-13 D-15, on a person in 80 Evidence graphs (past the 64 the
// loaded-graph store keeps): the claims read composes from cached records, so
// a cold read fills each graph's records once and a warm read is lookups.
const (
	budgetClaimsCold = time.Second
	budgetClaimsWarm = 50 * time.Millisecond
)

func TestClaimsBudgets(t *testing.T) {
	f := newFixture(t)
	son := households(f, 80)
	read := func() { f.claims(son) }

	t.Run("fixture", func(t *testing.T) {
		groups := f.claims(son)
		if len(groups) != 80 {
			t.Fatalf("%d groups, want 80", len(groups))
		}
		for _, g := range groups {
			if len(g.Claims) != 1 {
				t.Fatalf("group %q has %d claims, want 1", g.Source.Title, len(g.Claims))
			}
		}
	})

	t.Run("cold", func(t *testing.T) {
		d := catalogsessiontest.Cold(t, f.C, 2, read)
		databasetest.WithinBudget(t, "cold claims read at 80 graphs", d, budgetClaimsCold)
	})

	t.Run("warm", func(t *testing.T) {
		d := databasetest.Warm(read)
		databasetest.WithinBudget(t, "warm claims read at 80 graphs", d, budgetClaimsWarm)
	})
}
