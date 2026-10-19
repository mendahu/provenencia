package effects_test

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/rowchange"
)

// TestResolveLists pins which catalog-wide lists a batch reports: each list
// once, in the order of effects.Lists, and an Observation update only when it
// writes property_id.
func TestResolveLists(t *testing.T) {
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	id := func(b byte) []byte { return []byte{b, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15} }
	sourceType := rowchange.Change{
		EntityType: "source_type", EntityID: id(1), Action: rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{"label": "Book"}),
	}
	graph := rowchange.Change{
		EntityType: "evidence_graph", EntityID: id(2), Action: rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{"source_id": id(3)}),
	}
	obs := func(action string, fields map[string]rowchange.FieldDiff) rowchange.Change {
		return rowchange.Change{EntityType: "observation", EntityID: id(4), Action: action, Fields: fields}
	}
	text := map[string]rowchange.FieldDiff{"value_text": {Old: "a", New: "b"}}
	prop := map[string]rowchange.FieldDiff{"property_id": {Old: id(5), New: id(6)}}

	tests := []struct {
		name    string
		changes []rowchange.Change
		want    []effects.List
	}{
		{"graph then vocabulary", []rowchange.Change{graph, sourceType},
			[]effects.List{effects.ListVocabulary, effects.ListEvidenceGraphs}},
		{"repeated change adds one key", []rowchange.Change{graph, graph},
			[]effects.List{effects.ListEvidenceGraphs}},
		{"observation text edit", []rowchange.Change{obs(rowchange.ActionUpdate, text)}, nil},
		{"observation property edit", []rowchange.Change{obs(rowchange.ActionUpdate, prop)},
			[]effects.List{effects.ListPropertyUsage}},
		{"observation delete", []rowchange.Change{obs(rowchange.ActionDelete, text)},
			[]effects.List{effects.ListPropertyUsage}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := effects.Resolve(tx, tt.changes)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(set.Lists, tt.want) {
				t.Fatalf("Lists = %v, want %v", set.Lists, tt.want)
			}
			for _, l := range effects.Lists {
				want := false
				for _, w := range tt.want {
					want = want || w == l
				}
				if set.Has(l) != want {
					t.Errorf("Has(%s) = %v, want %v", l, set.Has(l), want)
				}
			}
		})
	}
}
