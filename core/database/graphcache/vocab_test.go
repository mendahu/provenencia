package graphcache_test

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/vocabkey"
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
