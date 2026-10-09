package effects

import (
	"testing"

	"github.com/mendahu/provenencia/core/database/catalogmodel"
)

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
