package persondetail_test

import (
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/persondetail"
)

// shape builds one family around a Person: a pair of parents from a birth, the
// siblings, a marriage, a second marriage when more is set, and children each
// with a birth. It returns the Person.
func (w *kinWorld) shape(tag string, siblings, children int, more bool) []byte {
	pat := w.person(tag + "Pat")
	john, mary := w.person(tag+"John"), w.person(tag+"Mary")
	w.k.Birth(pat, john, mary)
	for i := 0; i < siblings; i++ {
		w.k.Birth(w.person(tag+"Sib"+string(rune('a'+i))), john, mary)
	}
	sam := w.person(tag + "Sam")
	w.k.Marriage(pat, sam)
	for i := 0; i < children; i++ {
		w.k.Birth(w.person(tag+"Kid"+string(rune('a'+i))), pat, sam)
	}
	if more {
		w.k.Marriage(pat, w.person(tag+"Tess"))
	}
	return pat
}

func TestPersonFamilyQueryCount(t *testing.T) {
	t.Run("AC-2.1", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.shape("", 2, 2, true)
		first, _, _, err := persondetail.PersonFamily(w.c, pat)
		if err != nil {
			t.Fatal(err)
		}
		var queries int
		w.c.Graph().SetQueryCounterForTest(func() { queries++ })
		second, _, _, err := persondetail.PersonFamily(w.c, pat)
		if err != nil {
			t.Fatal(err)
		}
		if queries != 0 {
			t.Fatalf("a second read issued %d graph queries", queries)
		}
		if !reflect.DeepEqual(first, second) {
			t.Fatalf("second read differs:\n%v\n%v", w.show(first.Units), w.show(second.Units))
		}
		if len(first.Units) == 0 {
			t.Fatal("the family is empty")
		}
	})
	t.Run("AC-2.2", func(t *testing.T) {
		count := func(siblings, children int, more bool) int {
			w := newKinWorld(t)
			pat := w.shape("", siblings, children, more)
			var queries int
			w.c.Graph().SetQueryCounterForTest(func() { queries++ })
			if _, _, _, err := persondetail.PersonFamily(w.c, pat); err != nil {
				t.Fatal(err)
			}
			return queries
		}
		a, b := count(1, 1, false), count(7, 7, true)
		if a == 0 || a != b {
			t.Fatalf("queries: small family %d, large family %d", a, b)
		}
	})
	t.Run("AC-2.3", func(t *testing.T) {
		count := func(copies int) int {
			w := newKinWorld(t)
			var ids [][]byte
			for i := 0; i < copies; i++ {
				ids = append(ids, w.shape(string(rune('x'+i)), 1, 1, false))
			}
			var queries int
			w.c.Graph().SetQueryCounterForTest(func() { queries++ })
			got, err := persondetail.Families(w.c, ids)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != copies {
				t.Fatalf("%d families, want %d", len(got), copies)
			}
			return queries
		}
		a, b := count(1), count(3)
		if a == 0 || a != b {
			t.Fatalf("queries: one person %d, three people %d", a, b)
		}
	})
}
