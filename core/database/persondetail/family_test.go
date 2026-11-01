package persondetail_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/kinship"
)

// kin is a catalog of canonical people named for a test, and the Canon that
// writes them.
type kinWorld struct {
	t     *testing.T
	c     *database.Catalog
	k     *graphcachetest.Canon
	names map[string]string
}

func newKinWorld(t *testing.T) *kinWorld {
	t.Helper()
	c := graphcachetest.OpenCatalog(t)
	// A project session warms the default name format; a summary read needs it.
	if err := nameformats.Warm(c); err != nil {
		t.Fatal(err)
	}
	return &kinWorld{t: t, c: c, k: graphcachetest.NewCanon(t, c), names: map[string]string{}}
}

func (w *kinWorld) person(name string) []byte {
	id := w.k.Person()
	w.names[string(id)] = name
	return id
}

func (w *kinWorld) name(id []byte) string {
	if n, ok := w.names[string(id)]; ok {
		return n
	}
	return fmt.Sprintf("%x", id)
}

func unitKind(k kinship.UnitKind) string {
	return map[kinship.UnitKind]string{
		kinship.UnitParents: "Parents", kinship.UnitPartner: "Partner",
		kinship.UnitOneParent: "OneParent", kinship.UnitSiblings: "Siblings",
	}[k]
}

// show writes units as Kind{partners;children}, each side sorted by name.
func (w *kinWorld) show(units []kinship.Unit) []string {
	side := func(rs []kinship.Relative) string {
		var ns []string
		for _, r := range rs {
			ns = append(ns, w.name(r.ID))
		}
		sort.Strings(ns)
		return strings.Join(ns, ",")
	}
	var out []string
	for _, u := range units {
		out = append(out, fmt.Sprintf("%s{%s;%s}", unitKind(u.Kind), side(u.Partners), side(u.Children)))
	}
	return out
}

func (w *kinWorld) family(id []byte) (persondetail.Family, error) {
	w.t.Helper()
	f, _, err := persondetail.PersonFamily(w.c, id)
	return f, err
}

func (w *kinWorld) expect(id []byte, want ...string) persondetail.Family {
	w.t.Helper()
	f, summaries, err := persondetail.PersonFamily(w.c, id)
	if err != nil {
		w.t.Fatal(err)
	}
	got := w.show(f.Units)
	if len(got) != 0 || len(want) != 0 {
		if !reflect.DeepEqual(got, want) {
			w.t.Fatalf("family of %s:\n got %v\nwant %v", w.name(id), got, want)
		}
	}
	for _, u := range f.Units {
		for _, r := range slices.Concat(u.Partners, u.Children) {
			if _, ok := summaries.Person(r.ID); !ok {
				w.t.Errorf("no summary for %s", w.name(r.ID))
			}
		}
	}
	return f
}

func relative(t *testing.T, f persondetail.Family, kind kinship.UnitKind, children bool, id []byte) kinship.Relative {
	t.Helper()
	for _, u := range f.Units {
		rs := u.Partners
		if children {
			rs = u.Children
		}
		if u.Kind != kind {
			continue
		}
		for _, r := range rs {
			if bytes.Equal(r.ID, id) {
				return r
			}
		}
	}
	t.Fatalf("%x is in no such unit", id)
	return kinship.Relative{}
}

func TestPersonFamily(t *testing.T) {
	t.Run("AC-1.1", func(t *testing.T) {
		w := newKinWorld(t)
		pat, sam, tess, uma, vic := w.person("Pat"), w.person("Sam"), w.person("Tess"), w.person("Uma"), w.person("Vic")
		m1, _ := w.k.Marriage(pat, sam)
		w.k.Marriage(pat, tess)
		patUma := w.k.Parent(pat, uma)
		w.k.Parent(sam, uma)
		b2, _ := w.k.Birth(vic, pat, tess)
		f := w.expect(pat, "Partner{Sam;Uma}", "Partner{Tess;Vic}")

		samLink := relative(t, f, kinship.UnitPartner, false, sam).Link
		if samLink == nil || !samLink.OnlyImplied || len(samLink.Shapes) != 1 ||
			samLink.Shapes[0].Pattern != kinship.PatternMarriageSpouses || !bytes.Equal(samLink.Shapes[0].Pivot, m1) {
			t.Fatalf("Sam's link %+v", samLink)
		}
		vicLink := relative(t, f, kinship.UnitPartner, true, vic).Link
		if vicLink == nil || !vicLink.OnlyImplied || len(vicLink.Shapes) != 1 ||
			vicLink.Shapes[0].Pattern != kinship.PatternBirthParents || !bytes.Equal(vicLink.Shapes[0].Pivot, b2) {
			t.Fatalf("Vic's link %+v", vicLink)
		}
		umaLink := relative(t, f, kinship.UnitPartner, true, uma).Link
		if umaLink == nil || umaLink.OnlyImplied || !slices.ContainsFunc(umaLink.Associations, func(a []byte) bool { return bytes.Equal(a, patUma) }) {
			t.Fatalf("Uma's link %+v", umaLink)
		}
	})
	t.Run("AC-1.2", func(t *testing.T) {
		w := newKinWorld(t)
		pat, john, mary, james := w.person("Pat"), w.person("John"), w.person("Mary"), w.person("James")
		w.k.Birth(pat, john, mary)
		w.k.Parent(james, pat)
		w.expect(pat, "Parents{James;}", "Parents{John,Mary;}")
	})
	t.Run("AC-1.5", func(t *testing.T) {
		w := newKinWorld(t)
		pat, john, mary, quin, rae := w.person("Pat"), w.person("John"), w.person("Mary"), w.person("Quin"), w.person("Rae")
		w.k.Birth(pat, john, mary)
		w.k.Birth(quin, john, mary)
		w.k.Parent(mary, rae)
		f := w.expect(pat, "Parents{John,Mary;Quin}", "Parents{Mary;Rae}")
		shapes := func(id []byte) int {
			l := relative(t, f, kinship.UnitParents, true, id).Link
			n := 0
			for _, s := range l.Shapes {
				if s.Pattern == kinship.PatternSharedParent {
					n++
				}
			}
			return n
		}
		if q, r := shapes(quin), shapes(rae); q != 2 || r != 1 {
			t.Fatalf("shared_parent shapes: Quin %d, Rae %d", q, r)
		}
	})
	t.Run("AC-1.7", func(t *testing.T) {
		w := newKinWorld(t)
		pat, wren, xan := w.person("Pat"), w.person("Wren"), w.person("Xan")
		w.k.Parent(pat, wren)
		w.k.Parent(xan, wren)
		f := w.expect(pat, "Partner{Xan;Wren}")
		if l := relative(t, f, kinship.UnitPartner, false, xan).Link; l != nil {
			t.Fatalf("Xan has link %+v", l)
		}
	})
	t.Run("AC-1.9", func(t *testing.T) {
		w := newKinWorld(t)
		pat, john, uma, other := w.person("Pat"), w.person("John"), w.person("Uma"), w.person("Other")
		w.k.Parent(john, pat)
		w.k.Parent(pat, uma)
		w.k.Provisional(john)
		w.k.Merge(uma, other)
		w.k.Commit(john, uma, other)
		w.expect(pat)
	})
	t.Run("AC-1.12", func(t *testing.T) {
		w := newKinWorld(t)
		pat, other := w.person("Pat"), w.person("Other")
		ev := w.k.Event("birth")
		merged := w.person("Merged")
		w.k.Merge(merged, other)
		w.k.Commit(merged, other)
		for name, id := range map[string][]byte{"unknown": bytes.Repeat([]byte{7}, 16), "event": ev, "merged": merged, "malformed": {1, 2, 3}} {
			if _, _, err := persondetail.PersonFamily(w.c, id); !errors.Is(err, conclusiondetails.ErrNotFound) {
				t.Errorf("%s: err %v, want ErrNotFound", name, err)
			}
		}
		got, err := persondetail.Families(w.c, [][]byte{bytes.Repeat([]byte{7}, 16), ev, merged, {1, 2, 3}, pat})
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || !bytes.Equal(got[0].Person, pat) {
			t.Fatalf("Families kept %d, want only Pat", len(got))
		}
	})
	t.Run("AC-1.13", func(t *testing.T) {
		w := newKinWorld(t)
		w.expect(w.person("Lone"))
	})
	t.Run("AC-1.14", func(t *testing.T) {
		w := newKinWorld(t)
		pat, sam, uma, tess := w.person("Pat"), w.person("Sam"), w.person("Uma"), w.person("Tess")
		w.k.Marriage(pat, sam)
		w.k.Marriage(pat, tess)
		w.k.Parent(pat, uma)
		w.k.Parent(sam, uma)
		ids := [][]byte{pat, sam, uma}
		got, err := persondetail.Families(w.c, ids)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(ids) {
			t.Fatalf("%d families, want %d", len(got), len(ids))
		}
		for i, id := range ids {
			if !bytes.Equal(got[i].Person, id) {
				t.Fatalf("family %d is for %s, want %s", i, w.name(got[i].Person), w.name(id))
			}
			one, _, err := persondetail.PersonFamily(w.c, id)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got[i], one) {
				t.Errorf("Families(%s) differs from PersonFamily:\n%v\n%v", w.name(id), w.show(got[i].Units), w.show(one.Units))
			}
		}
		dup, err := persondetail.Families(w.c, [][]byte{pat, pat})
		if err != nil || len(dup) != 1 {
			t.Fatalf("a repeated id gave %d families, err %v", len(dup), err)
		}
	})
	t.Run("AC-2.5", func(t *testing.T) {
		w := newKinWorld(t)
		pat, ned := w.person("Pat"), w.person("Ned")
		w.expect(pat)
		link := w.k.Parent(pat, ned)
		w.k.Commit(pat, ned, link)
		w.expect(pat, "OneParent{;Ned}")
		w.k.Remove(link)
		w.k.Commit(pat, ned, link)
		w.expect(pat)
	})
}
