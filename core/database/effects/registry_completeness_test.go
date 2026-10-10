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
	return !e.Source.zero() || !e.Handles.zero() || !e.Structure.zero() || len(e.Search) > 0 || e.Vocabulary
}
