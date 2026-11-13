package persondetail_test

import (
	"testing"
	"time"

	"github.com/mendahu/provenencia/core/catalogsession/catalogsessiontest"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/databasetest"
)

// Budgets of S16-30 D-13, on a person in 80 Evidence graphs (past the 64 the
// loaded-graph store keeps): the detail read composes a record from the node's
// outcome row, the graph's Subject record and the Source record.
const (
	budgetDetailCold = time.Second
	budgetDetailWarm = 50 * time.Millisecond
)

func TestDetailBudgets(t *testing.T) {
	f := newFixture(t)
	son := households(f, 80)
	read := func() conclusiondetails.Detail {
		d, err := conclusiondetails.ForEntity(f.C, son)
		if err != nil {
			f.T.Fatal(err)
		}
		return d
	}

	t.Run("fixture", func(t *testing.T) {
		for _, fld := range read().Fields {
			if fld.PropertyKey == "name" {
				if len(fld.Outcomes) != 80 {
					t.Fatalf("%d name outcomes, want 80", len(fld.Outcomes))
				}
				return
			}
		}
		t.Fatal("no name field")
	})

	t.Run("cold", func(t *testing.T) {
		d := catalogsessiontest.Cold(t, f.C, 2, func() { read() })
		databasetest.WithinBudget(t, "cold detail read at 80 graphs", d, budgetDetailCold)
	})

	t.Run("warm", func(t *testing.T) {
		d := databasetest.Warm(func() { read() })
		databasetest.WithinBudget(t, "warm detail read at 80 graphs", d, budgetDetailWarm)
	})
}
