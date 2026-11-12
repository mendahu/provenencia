package effects_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/metadatafields"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sourcemetadata"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// TestReleasedMetadataScopesThroughDeletedOwner runs the real deletes of an
// Artifact and of an Evidence graph that each hold one metadata value, and
// resolves every released metadata change on its own: it reaches its Source
// through the deleted owner's change, and nowhere without it.
func TestReleasedMetadataScopesThroughDeletedOwner(t *testing.T) {
	tests := []struct {
		name     string
		entities []string // the released metadata changes
		owner    string   // the owner's delete entity
		run      func(t *testing.T) (tx *database.Tx, changes []rowchange.Change, sourceID []byte)
	}{
		{
			name:     "artifact",
			entities: []string{"artifact_metadata", "artifact_metadata_layout"},
			owner:    "artifact",
			run: func(t *testing.T) (*database.Tx, []rowchange.Change, []byte) {
				tx, userID, _, _, art := readsFixture(t)
				setValue(t, tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelArtifact, ID: art.ID})
				changes, err := artifacts.Delete(tx, userID, art.ID)
				if err != nil {
					t.Fatal(err)
				}
				return tx, changes, art.SourceID
			},
		},
		{
			name:     "graph",
			entities: []string{"evidence_graph_metadata", "evidence_graph_metadata_layout"},
			owner:    "evidence_graph",
			run: func(t *testing.T) (*database.Tx, []rowchange.Change, []byte) {
				tx, userID, _, jones, _ := readsFixture(t)
				setValue(t, tx, userID, sourcemetadata.Owner{Level: vocabkey.LevelGraph, ID: jones.ID})
				changes, err := evidencegraphs.Delete(tx, userID, jones.ID)
				if err != nil {
					t.Fatal(err)
				}
				return tx, changes, jones.SourceID
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, changes, sourceID := tt.run(t)
			var filtered []rowchange.Change
			for _, ch := range changes {
				if ch.EntityType != tt.owner {
					filtered = append(filtered, ch)
				}
			}
			if len(filtered) == len(changes) {
				t.Fatalf("no %s delete among the changes", tt.owner)
			}
			for _, entity := range tt.entities {
				idx := indexOf(changes, entity)
				if len(idx) != 1 {
					t.Fatalf("%s changes = %d, want 1", entity, len(idx))
				}
				got, err := effects.SourceScopeOfChange(tx.Tx, changes, idx[0])
				if err != nil {
					t.Fatal(err)
				}
				if len(got) != 1 || !bytes.Equal(got[0], sourceID) {
					t.Fatalf("%s scope = %x, want the Source %x", entity, got, sourceID)
				}
				j := indexOf(filtered, entity)
				if len(j) != 1 {
					t.Fatalf("%s changes without the owner = %d, want 1", entity, len(j))
				}
				got, err = effects.SourceScopeOfChange(tx.Tx, filtered, j[0])
				if err != nil || len(got) != 0 {
					t.Fatalf("%s scope without the %s delete = %x, %v; want none", entity, tt.owner, got, err)
				}
			}
		})
	}
}

func indexOf(changes []rowchange.Change, entity string) []int {
	var out []int
	for i, ch := range changes {
		if ch.EntityType == entity {
			out = append(out, i)
		}
	}
	return out
}

func setValue(t *testing.T, tx *database.Tx, userID []byte, owner sourcemetadata.Owner) {
	t.Helper()
	field, _, err := metadatafields.Create(tx, "Provider", metadatafields.DataTypeText, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := sourcemetadata.Set(tx, userID, owner, sourcemetadata.Input{FieldID: field.ID, ValueText: "Ancestry"}); err != nil {
		t.Fatal(err)
	}
}
