package effects_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/rowchange"
)

func id16(b byte) []byte { return bytes.Repeat([]byte{b}, 16) }

func handleDelete(id []byte) rowchange.Change {
	return rowchange.Change{
		EntityType: "canonical_entity", EntityID: id, Action: rowchange.ActionDelete,
		Fields: rowchange.DeletedRow(map[string]any{"ref": "PER-X", "merged_into_id": nil}),
	}
}

func TestResolveReportsDeletedHandles(t *testing.T) {
	h1, h2, h3 := id16(1), id16(2), id16(3)
	claimDelete := rowchange.Change{
		EntityType: "identity_claim", EntityID: id16(9), Action: rowchange.ActionDelete,
		Fields: rowchange.DeletedRow(map[string]any{"subject_id": id16(8), "entity_id": h1, "status": "accepted"}),
	}
	merge := rowchange.Change{
		EntityType: "canonical_entity", EntityID: h2, Action: rowchange.ActionUpdate,
		Fields: map[string]rowchange.FieldDiff{"merged_into_id": {Old: nil, New: h1}},
	}
	create := rowchange.Change{
		EntityType: "canonical_entity", EntityID: h3, Action: rowchange.ActionCreate,
		Fields: rowchange.FullRow(map[string]any{"ref": "PER-Y"}),
	}
	tests := []struct {
		name    string
		changes []rowchange.Change
		want    [][]byte
	}{
		{"a handle delete among other changes", []rowchange.Change{handleDelete(h1), claimDelete, merge}, [][]byte{h1}},
		{"two deletes in change order", []rowchange.Change{handleDelete(h1), handleDelete(h3)}, [][]byte{h1, h3}},
		{"the same handle twice is listed once", []rowchange.Change{handleDelete(h1), handleDelete(h1)}, [][]byte{h1}},
		{"a claim delete, a merge and a create delete no handle", []rowchange.Change{claimDelete, merge, create}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			set, err := effects.Resolve(tx, tt.changes)
			if err != nil {
				t.Fatal(err)
			}
			if len(set.DeletedHandles) != len(tt.want) {
				t.Fatalf("deleted handles %x, want %x", set.DeletedHandles, tt.want)
			}
			for i, id := range tt.want {
				if !bytes.Equal(set.DeletedHandles[i], id) {
					t.Fatalf("deleted handles %x, want %x", set.DeletedHandles, tt.want)
				}
			}
		})
	}
}
