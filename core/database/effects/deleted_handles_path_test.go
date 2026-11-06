package effects

import "testing"

func TestDeletedPathEndsAtTheEntrysOwnTable(t *testing.T) {
	if err := (Effect{Entity: "canonical_entity", Deleted: onDelete(self())}).validate("canonical_entities"); err != nil {
		t.Fatalf("a path ending at its own table: %v", err)
	}
	if err := (Effect{Entity: "citation_note", Deleted: onDelete(up("citation_id", "artifact_id", "source_id"))}).validate("citation_notes"); err == nil {
		t.Fatal("a deleted path ending at another table validated")
	}
	if err := (Effect{None: true, Deleted: onDelete(self())}).validate("x"); err == nil {
		t.Fatal("a None entry with Deleted validated")
	}
	if registry["canonical_entities"].Deleted.zero() {
		t.Fatal("canonical_entities does not set Deleted")
	}
}
