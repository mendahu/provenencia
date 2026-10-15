package effects

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

// TestEveryNonSkipTableHasAnEntry checks the registry against catalogmodel.
// Every table outside the skip bucket has an entry, including an explicit
// empty one. Skip tables stay out: project, audit, search docs, and the
// auto-reconciler. A path that names a missing foreign key panics in init,
// so this test only runs when every declared edge exists.
func TestEveryNonSkipTableHasAnEntry(t *testing.T) {
	seen := map[string]struct{}{}
	for _, table := range catalogmodel.Tables {
		if table.Bucket == catalogmodel.BucketSkip {
			if _, ok := registry[table.Name]; ok {
				t.Errorf("skip table %s has an effects entry", table.Name)
			}
			continue
		}
		if _, ok := registry[table.Name]; !ok {
			t.Errorf("no effects entry for %s", table.Name)
		}
		seen[table.Name] = struct{}{}
	}
	for name := range registry {
		if _, ok := seen[name]; !ok {
			t.Errorf("effects entry for unknown table %s", name)
		}
	}
}

// TestCascadeIntoNonEmptyEffectIsAudited checks that a CASCADE into a table
// with effects is Audited. Official deletes release and audit those rows
// first. A silent cascade is allowed only into an empty effect, such as
// layout or name-value parts. This pairs with deleteimpact's facet-release
// check, which requires a release for each audited CASCADE.
func TestCascadeIntoNonEmptyEffectIsAudited(t *testing.T) {
	for _, fk := range catalogmodel.FKs {
		if fk.OnDelete != "CASCADE" {
			continue
		}
		eff, ok := registry[fk.From]
		if !ok || !hasWork(eff) {
			continue
		}
		if !fk.Audited {
			t.Errorf("%s.%s cascades into a table with effects and is not Audited", fk.From, fk.Column)
		}
	}
}

// TestGradeTablesHaveNoHandles checks that credibility and confidence grades
// are an explicit empty effect. They are seed-stable: a grade edit must not
// fan out to every handle weighed by sort order. Source credibility grades
// are seed-locked in deleteimpact; claim confidence grades have no delete.
func TestGradeTablesHaveNoHandles(t *testing.T) {
	for _, name := range []string{"source_credibility_grades", "claim_confidence_grades"} {
		eff, ok := registry[name]
		if !ok {
			t.Fatalf("missing %s", name)
		}
		if !eff.None || !eff.Handles.zero() || eff.Vocabulary {
			t.Fatalf("%s should be an empty effect, got %+v", name, eff)
		}
	}
}

func hasWork(e Effect) bool {
	for _, p := range e.paths() {
		if !p.zero() {
			return true
		}
	}
	return len(e.Search) > 0 || e.Vocabulary
}

// TestSourceEvidenceIsNotActivity pins which tables move a Source's evidence
// (Effect.Source: what the Source graph, graph progress, citation counts, and
// citations by artifact load) and which move its last activity
// (Effect.SourceScope: the audit source scope). Identity claims change the
// evidence graph but have never been audited under a Source.
func TestSourceEvidenceIsNotActivity(t *testing.T) {
	tests := []struct {
		table    string
		evidence bool
		activity bool
	}{
		{"sources", false, true},
		{"source_notes", false, true},
		{"source_metadata", false, true},
		{"source_metadata_layout", false, true},
		{"source_credibility_assessments", false, true},
		{"files", false, true},
		{"citation_notes", false, true},
		{"observation_notes", false, true},
		{"artifacts", true, true},
		{"citations", true, true},
		{"observations", true, true},
		{"subjects", true, true},
		{"date_values", true, true},
		{"name_values", true, true},
		{"identity_claims", true, false},
		{"identity_claim_evidence", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.table, func(t *testing.T) {
			eff, ok := registry[tt.table]
			if !ok {
				t.Fatalf("no entry for %s", tt.table)
			}
			if got := !eff.Source.zero(); got != tt.evidence {
				t.Errorf("Source set = %v, want %v", got, tt.evidence)
			}
			if got := !eff.SourceScope.zero(); got != tt.activity {
				t.Errorf("SourceScope set = %v, want %v", got, tt.activity)
			}
		})
	}
}
