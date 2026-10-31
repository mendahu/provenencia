package effects

import (
	"slices"
	"sort"
	"testing"
)

// TestReadsDeclare pins which tables move each keyed read, that every read
// path ends at its key table, and that a bad entry does not validate.
func TestReadsDeclare(t *testing.T) {
	want := map[Read][]string{
		ReadCitations: {
			"citations", "date_values", "depiction_values", "name_values", "observations",
		},
		ReadEvidenceGraphProgress: {
			"identity_claims", "observations", "subjects",
		},
	}
	got := map[Read][]string{}
	for table, eff := range registry {
		for _, r := range eff.Reads {
			got[r.Read] = append(got[r.Read], table)
			end, err := r.Path.check(table)
			if err != nil {
				t.Errorf("%s %s: %v", table, r.Read, err)
			} else if end != r.Read.KeyTable() {
				t.Errorf("%s %s ends at %s, want %s", table, r.Read, end, r.Read.KeyTable())
			}
		}
	}
	for _, read := range Reads {
		if read.KeyTable() == "" {
			t.Errorf("%s has no key table", read)
		}
		sort.Strings(got[read])
		if !slices.Equal(got[read], want[read]) {
			t.Errorf("tables with %s = %v, want %v", read, got[read], want[read])
		}
	}
	if eff := registry["artifacts"]; len(eff.Reads) != 0 || !eff.EvidenceGraph.zero() {
		t.Errorf("artifacts name graphs or reads: %+v", eff)
	}
	if err := (Effect{Reads: []ReadPath{{"nope", self()}}}).validate("x"); err == nil {
		t.Fatal("an undeclared read validated")
	}
	if err := (Effect{Reads: []ReadPath{{ReadCitations, Path{}}}}).validate("x"); err == nil {
		t.Fatal("a read with no path validated")
	}
	if err := (Effect{None: true, Reads: []ReadPath{{ReadCitations, self()}}}).validate("x"); err == nil {
		t.Fatal("a None entry with Reads validated")
	}
}
