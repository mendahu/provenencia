//go:build cgo

package databasetest_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/databasetest"
)

func TestCountSelects(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	sel := databasetest.CountSelects(t, c)

	one := func() {
		t.Helper()
		var n int
		if err := db.QueryRow(`SELECT 1`).Scan(&n); err != nil {
			t.Fatal(err)
		}
	}
	tests := []struct {
		name string
		do   func()
		want int
	}{
		{"one SELECT is counted", one, 1},
		{"a second adds one", one, 2},
		{"a write is not counted", func() {
			if _, err := db.Exec(`CREATE TABLE t (x INTEGER)`); err != nil {
				t.Fatal(err)
			}
		}, 2},
		{"Reset starts again from 0", sel.Reset, 0},
		{"a query after Reset is counted and does not block", one, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.do()
			if got := sel.N(); got != tt.want {
				t.Fatalf("N() = %d, want %d", got, tt.want)
			}
		})
	}
}
