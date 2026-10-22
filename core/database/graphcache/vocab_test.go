package graphcache_test

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestVocabularyBindings(t *testing.T) {
	c := openCatalog(t)
	v, err := c.Graph().Vocabulary()
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Bindings) == 0 {
		t.Fatal("the seeded vocabulary carries no bindings")
	}
	got := v.BindingOrder()[vocabkey.Person]
	want := []vocabkey.PropertyKey{"name", "likeness", "sex_at_birth", "gender", "interests"}
	if !slices.Equal(got, want) {
		t.Fatalf("BindingOrder()[person] = %v, want %v", got, want)
	}
	if ev := v.BindingOrder()[vocabkey.Event]; len(ev) < 3 || ev[0] != "event_type" {
		t.Fatalf("BindingOrder()[event] = %v, want event_type first", ev)
	}
}

// A plugin type that shares a key with a provenencia type does not take over
// that key's binding order.
func TestVocabularyBindingsPluginTypeSeededKey(t *testing.T) {
	c := openCatalog(t)
	typeID, err := subjecttypes.Upsert(c, subjecttypes.Type{
		Key: "person", Origin: "plugin:x", Label: "Person (plugin)",
		RefPrefix: "XPR", CandidateRefPrefix: "XPC",
	})
	if err != nil {
		t.Fatal(err)
	}
	prop, err := properties.Lookup(c, "interests", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.EnsureBinding(c, typeID, prop.ID, 0); err != nil {
		t.Fatal(err)
	}
	v, err := c.Graph().Vocabulary()
	if err != nil {
		t.Fatal(err)
	}
	got := v.BindingOrder()[vocabkey.Person]
	want := []vocabkey.PropertyKey{"name", "likeness", "sex_at_birth", "gender", "interests"}
	if !slices.Equal(got, want) {
		t.Fatalf("BindingOrder()[person] = %v, want only the provenencia person's %v", got, want)
	}
	if order := v.TypeOrder(typeID); !slices.Equal(order, []vocabkey.PropertyKey{"interests"}) {
		t.Fatalf("TypeOrder(plugin person) = %v", order)
	}
}

// A binding reorder committed through writes.Run drops the cached vocabulary,
// so the next read sees the new order with no restart.
func TestVocabularyDropsOnBindingReorder(t *testing.T) {
	c := openCatalog(t)
	userID := []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}
	person, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Graph().Vocabulary(); err != nil {
		t.Fatal(err)
	}
	bs, err := subjectvocab.ListBindings(c, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([][]byte, len(bs))
	for i, b := range bs {
		ids[len(bs)-1-i] = b.Property.ID
	}
	if _, err := writes.Call(c, writes.Op{Action: "reorder_subject_type_properties", UserID: userID},
		func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			ch, err := subjectvocab.ReorderBindings(tx, userID, person.ID, ids)
			return struct{}{}, ch, err
		}); err != nil {
		t.Fatal(err)
	}
	v, err := c.Graph().Vocabulary()
	if err != nil {
		t.Fatal(err)
	}
	got := v.BindingOrder()[vocabkey.Person]
	if len(got) != len(bs) || got[0] != "interests" || got[len(got)-1] != "name" {
		t.Fatalf("order after reorder = %v", got)
	}
}
