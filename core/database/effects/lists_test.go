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
	source := func(action string, fields map[string]rowchange.FieldDiff) rowchange.Change {
		return rowchange.Change{EntityType: "source", EntityID: id(7), Action: action, Fields: fields}
	}
	entity := func(action string, fields map[string]rowchange.FieldDiff) rowchange.Change {
		return rowchange.Change{EntityType: "canonical_entity", EntityID: id(8), Action: action, Fields: fields}
	}
	title := map[string]rowchange.FieldDiff{"title": {Old: "a", New: "b"}}
	preferred := map[string]rowchange.FieldDiff{"preferred_likeness_observation_id": {Old: id(5), New: id(6)}}
	merged := map[string]rowchange.FieldDiff{"merged_into_id": {Old: nil, New: id(6)}}
	text := map[string]rowchange.FieldDiff{"value_text": {Old: "a", New: "b"}}
	prop := map[string]rowchange.FieldDiff{"property_id": {Old: id(5), New: id(6)}}

	tests := []struct {
		name    string
		changes []rowchange.Change
		want    []effects.List
	}{
		{"graph then vocabulary", []rowchange.Change{graph, sourceType},
			[]effects.List{effects.ListVocabulary}},
		{"repeated change adds one key", []rowchange.Change{sourceType, sourceType},
			[]effects.List{effects.ListVocabulary}},
		{"observation text edit", []rowchange.Change{obs(rowchange.ActionUpdate, text)}, nil},
		{"observation property edit", []rowchange.Change{obs(rowchange.ActionUpdate, prop)},
			[]effects.List{effects.ListPropertyUsage}},
		{"observation delete", []rowchange.Change{obs(rowchange.ActionDelete, text)},
			[]effects.List{effects.ListPropertyUsage}},
		{"source create", []rowchange.Change{source(rowchange.ActionCreate, rowchange.FullRow(map[string]any{"title": "a"}))},
			[]effects.List{effects.ListCounts}},
		{"source title edit", []rowchange.Change{source(rowchange.ActionUpdate, title)}, nil},
		{"source delete", []rowchange.Change{source(rowchange.ActionDelete, rowchange.FullRow(map[string]any{"title": "a"}))},
			[]effects.List{effects.ListCounts}},
		{"handle create", []rowchange.Change{entity(rowchange.ActionCreate, rowchange.FullRow(map[string]any{"ref": "PER-1"}))},
			[]effects.List{effects.ListCounts}},
		{"handle preferred picture edit", []rowchange.Change{entity(rowchange.ActionUpdate, preferred)}, nil},
		{"handle merge", []rowchange.Change{entity(rowchange.ActionUpdate, merged)},
			[]effects.List{effects.ListCounts}},
		{"source type then source keeps list order",
			[]rowchange.Change{source(rowchange.ActionCreate, rowchange.FullRow(map[string]any{"title": "a"})), sourceType},
			[]effects.List{effects.ListVocabulary, effects.ListCounts}},
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
