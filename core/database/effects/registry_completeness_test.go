package effects

import (
	"github.com/mendahu/provenencia/core/database/rowchange"
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
		citation bool // CitationSource: the citation reads load it
	}{
		{"sources", false, true, false},
		{"source_notes", false, true, false},
		{"source_metadata", false, true, false},
		{"source_metadata_layout", false, true, false},
		{"source_credibility_assessments", false, true, false},
		{"evidence_graphs", false, true, false},
		{"evidence_graph_artifacts", false, true, false},
		{"files", false, true, false},
		{"citation_notes", false, true, false},
		{"observation_notes", false, true, false},
		{"artifacts", true, true, true},
		{"citations", true, true, true},
		{"observations", true, true, true},
		{"subjects", true, true, false},
		{"date_values", true, true, true},
		{"name_values", true, true, true},
		{"identity_claims", true, false, false},
		{"identity_claim_evidence", true, false, false},
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
			if got := !eff.CitationSource.zero(); got != tt.citation {
				t.Errorf("CitationSource set = %v, want %v", got, tt.citation)
			}
		})
	}
}

// TestPropertyUsage pins when an Observation change moves the Properties
// workspace's usedBy counts: a create, a delete, or a property_id edit.
func TestPropertyUsage(t *testing.T) {
	text := map[string]rowchange.FieldDiff{"value_text": {Old: "a", New: "b"}}
	prop := map[string]rowchange.FieldDiff{"property_id": {Old: []byte{1}, New: []byte{2}}}
	tests := []struct {
		name  string
		table string
		ch    rowchange.Change
		want  bool
	}{
		{"create", "observations", rowchange.Change{Action: rowchange.ActionCreate, Fields: text}, true},
		{"delete", "observations", rowchange.Change{Action: rowchange.ActionDelete, Fields: text}, true},
		{"text edit", "observations", rowchange.Change{Action: rowchange.ActionUpdate, Fields: text}, false},
		{"property edit", "observations", rowchange.Change{Action: rowchange.ActionUpdate, Fields: prop}, true},
		{"subject create", "subjects", rowchange.Change{Action: rowchange.ActionCreate, Fields: text}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := registry[tt.table].usesProperty(tt.ch); got != tt.want {
				t.Errorf("usesProperty = %v, want %v", got, tt.want)
			}
		})
	}
}
