package effects

import "testing"

func TestDeletedHandlesPathEndsAtHandles(t *testing.T) {
	if err := (Effect{DeletedHandles: onDelete(self())}).validate("canonical_entities"); err != nil {
		t.Fatalf("a path ending at canonical_entities: %v", err)
	}
	if err := (Effect{DeletedHandles: onDelete(self())}).validate("sources"); err == nil {
		t.Fatal("a deleted-handles path ending at sources validated")
	}
	if err := (Effect{None: true, DeletedHandles: onDelete(self())}).validate("x"); err == nil {
		t.Fatal("a None entry with DeletedHandles validated")
	}
	for table, eff := range registry {
		if table != "canonical_entities" && !eff.DeletedHandles.zero() {
			t.Fatalf("%s sets DeletedHandles; only canonical_entities names handles", table)
		}
	}
	if registry["canonical_entities"].DeletedHandles.zero() {
		t.Fatal("canonical_entities does not set DeletedHandles")
	}
}
