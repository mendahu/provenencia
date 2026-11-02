package persondetail_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/conclusionsummaries"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func yearValue(y int) datevalue.Value {
	return datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &y}
}

// event adds an event of the type, names it for the failure messages and gives
// it a date when year is not zero.
func (w *kinWorld) event(name string, typ vocabkey.EventType, year int) []byte {
	id := w.k.Event(typ)
	w.names[string(id)] = name
	if year != 0 {
		w.k.Date(id, vocabkey.PropertyDate, yearValue(year))
	}
	return id
}

// byRef orders ids by the ref of the entity they name, as the graph does for
// ties: refs compare as strings, so creation order is not their order.
func (w *kinWorld) byRef(ids ...[]byte) [][]byte {
	w.t.Helper()
	ref := func(id []byte) string {
		n, err := w.c.Graph().Node(id)
		if err != nil || n == nil {
			w.t.Fatalf("no node %x: %v", id, err)
		}
		return strings.ToLower(n.Ref)
	}
	out := append([][]byte(nil), ids...)
	sort.SliceStable(out, func(i, j int) bool { return ref(out[i]) < ref(out[j]) })
	return out
}

func keys(ts []conclusionsummaries.Term) string {
	var ks []string
	for _, t := range ts {
		ks = append(ks, t.Key)
	}
	return strings.Join(ks, "+")
}

// line writes a row as Event(roles;Person:roles,...).
func (w *kinWorld) line(r persondetail.PersonEvent) string {
	var ps []string
	for _, p := range r.Participants {
		ps = append(ps, w.name(p.Person)+":"+keys(p.Roles))
	}
	return fmt.Sprintf("%s(%s;%s)", w.name(r.Event), keys(r.Roles), strings.Join(ps, ","))
}

func (w *kinWorld) events(id []byte) ([]persondetail.PersonEvent, conclusionsummaries.Summaries) {
	w.t.Helper()
	rows, summaries, err := persondetail.PersonEvents(w.c, id)
	if err != nil {
		w.t.Fatal(err)
	}
	for _, r := range rows {
		if _, ok := summaries.Event(r.Event); !ok {
			w.t.Errorf("no summary for event %s", w.name(r.Event))
		}
		for _, p := range r.Participants {
			if _, ok := summaries.Person(p.Person); !ok {
				w.t.Errorf("no summary for participant %s", w.name(p.Person))
			}
		}
	}
	return rows, summaries
}

func (w *kinWorld) expectEvents(id []byte, want ...string) []persondetail.PersonEvent {
	w.t.Helper()
	rows, _ := w.events(id)
	var got []string
	for _, r := range rows {
		got = append(got, w.line(r))
	}
	if !reflect.DeepEqual(got, append([]string(nil), want...)) && (len(got) != 0 || len(want) != 0) {
		w.t.Fatalf("events of %s:\n got %v\nwant %v", w.name(id), got, want)
	}
	return rows
}

// ids names events and their ids for an expected order.
func (w *kinWorld) order(ids ...[]byte) []string {
	var out []string
	for _, id := range ids {
		out = append(out, w.name(id))
	}
	return out
}

func eventOrder(w *kinWorld, rows []persondetail.PersonEvent) []string {
	var out []string
	for _, r := range rows {
		out = append(out, w.name(r.Event))
	}
	return out
}

func TestPersonEvents(t *testing.T) {
	t.Run("AC-1.1", func(t *testing.T) {
		w := newKinWorld(t)
		pat, sam, kid, ned := w.person("Pat"), w.person("Sam"), w.person("Kid"), w.person("Ned")
		b1 := w.event("B1", vocabkey.EventBirth, 1820)
		w.k.Join(pat, b1, "subject")
		m := w.event("M", vocabkey.EventMarriage, 1845)
		w.k.Join(pat, m, "subject")
		w.k.Join(sam, m, "subject")
		c := w.event("C", vocabkey.EventType("census"), 1851)
		w.k.Join(pat, c, "subject")
		w.k.Join(kid, c, "subject")
		u := w.event("U", vocabkey.EventBurial, 1890)
		w.k.Join(ned, u, "subject")
		w.k.Join(pat, u, "witness")
		rows := w.expectEvents(pat, "B1(subject;)", "M(subject;Sam:subject)", "C(subject;Kid:subject)", "U(witness;Ned:subject)")
		if rows[3].Roles[0].Label != "Witness" {
			t.Fatalf("role label %q", rows[3].Roles[0].Label)
		}
	})
	t.Run("AC-1.2", func(t *testing.T) {
		w := newKinWorld(t)
		pat, sam, kid := w.person("Pat"), w.person("Sam"), w.person("Kid")
		w.k.Parent(pat, kid)
		b2 := w.event("B2", vocabkey.EventBirth, 1850)
		w.k.Join(kid, b2, "subject")
		w.k.Join(sam, b2, "mother")
		w.expectEvents(pat)
	})
	t.Run("AC-1.3", func(t *testing.T) {
		w := newKinWorld(t)
		pat, sam, kid := w.person("Pat"), w.person("Sam"), w.person("Kid")
		b2 := w.event("B2", vocabkey.EventBirth, 1850)
		kj := w.k.Join(kid, b2, "subject")
		w.k.Join(pat, b2, "father")
		sj := w.k.Join(sam, b2, "mother")
		names := map[string]string{string(kj): "Kid:subject", string(sj): "Sam:mother"}
		var want []string
		for _, id := range w.byRef(kj, sj) {
			want = append(want, names[string(id)])
		}
		w.expectEvents(pat, "B2(father;"+strings.Join(want, ",")+")")
	})
	t.Run("AC-1.4", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		b1 := w.event("B1", vocabkey.EventBirth, 1820)
		r := w.event("R", vocabkey.EventType("residence"), 0)
		tr := w.event("T", vocabkey.EventType("travel"), 0)
		w.k.Date(tr, vocabkey.PropertyStartDate, yearValue(1860))
		v := w.event("V", vocabkey.EventType("migration"), 0)
		w.k.Date(v, vocabkey.PropertyEndDate, yearValue(1870))
		for _, e := range [][]byte{r, v, tr, b1} {
			w.k.Join(pat, e, "subject")
		}
		rows, _ := w.events(pat)
		want := append(w.order(b1, tr), w.order(w.byRef(r, v)...)...)
		if got := eventOrder(w, rows); !reflect.DeepEqual(got, want) {
			t.Fatalf("order %v, want %v", got, want)
		}
		dated := map[string]bool{"B1": true, "T": true, "R": false, "V": false}
		for _, row := range rows {
			if row.Dated != dated[w.name(row.Event)] {
				t.Errorf("%s dated = %v", w.name(row.Event), row.Dated)
			}
		}
	})
	t.Run("AC-1.5", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		var undated, dated [][]byte
		for i := 0; i < 2; i++ {
			e := w.event(fmt.Sprintf("U%d", i), vocabkey.EventType("residence"), 0)
			undated = append(undated, e)
			w.k.Join(pat, e, "subject")
			d := w.event(fmt.Sprintf("D%d", i), vocabkey.EventType("residence"), 1850)
			dated = append(dated, d)
			w.k.Join(pat, d, "subject")
		}
		rows, _ := w.events(pat)
		want := append(w.order(w.byRef(dated...)...), w.order(w.byRef(undated...)...)...)
		if got := eventOrder(w, rows); !reflect.DeepEqual(got, want) {
			t.Fatalf("order %v, want %v", got, want)
		}
	})
	t.Run("AC-1.6", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		c := w.event("C", vocabkey.EventType("census"), 1851)
		sj := w.k.Join(pat, c, "subject")
		ij := w.k.Join(pat, c, "informant")
		w.k.Join(pat, c, "subject")
		names := map[string]string{string(sj): "subject", string(ij): "informant"}
		var want []string
		for _, id := range w.byRef(sj, ij) {
			want = append(want, names[string(id)])
		}
		// A third link repeats subject at the end of the link order.
		rows := w.expectEvents(pat, "C("+strings.Join(want, "+")+";)")
		if len(rows) != 1 {
			t.Fatalf("%d rows", len(rows))
		}
	})
	t.Run("AC-1.7", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		c := w.event("C", vocabkey.EventType("census"), 1851)
		w.k.Link("participation", map[string][]byte{"person": pat, "event": c}, "", "")
		w.expectEvents(pat, "C(;)")
	})
	t.Run("AC-1.8", func(t *testing.T) {
		w := newKinWorld(t)
		pat, kid, ann, bea := w.person("Pat"), w.person("Kid"), w.person("Ann"), w.person("Bea")
		c := w.event("C", vocabkey.EventType("census"), 1851)
		for _, p := range [][]byte{pat, kid, ann} {
			w.k.Join(p, c, "subject")
		}
		w.k.Provisional(kid)
		w.k.Merge(ann, bea)
		w.k.Commit(kid, ann, bea)
		w.expectEvents(pat, "C(subject;)")
	})
	t.Run("AC-1.9", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		e1 := w.event("E1", vocabkey.EventType("census"), 1851)
		e2 := w.event("E2", vocabkey.EventType("census"), 1852)
		e3 := w.event("E3", vocabkey.EventType("census"), 1853)
		b1 := w.event("B1", vocabkey.EventBirth, 1820)
		for _, e := range [][]byte{e1, e2, b1} {
			w.k.Join(pat, e, "subject")
		}
		w.k.Provisional(e1)
		w.k.Merge(e2, e3)
		w.k.Commit(e1, e2, e3)
		w.expectEvents(pat, "B1(subject;)")
	})
	t.Run("AC-1.10", func(t *testing.T) {
		w := newKinWorld(t)
		pat, bea := w.person("Pat"), w.person("Bea")
		b1 := w.event("B1", vocabkey.EventBirth, 1820)
		w.k.Join(pat, b1, "subject")
		w.k.Merge(bea, pat)
		w.k.Commit(bea, pat)
		for name, id := range map[string][]byte{
			"unknown":   bytes.Repeat([]byte{0x7e}, 16),
			"event":     b1,
			"merged":    bea,
			"malformed": {1, 2, 3},
		} {
			if _, _, err := persondetail.PersonEvents(w.c, id); !errors.Is(err, conclusiondetails.ErrNotFound) {
				t.Errorf("%s: err = %v", name, err)
			}
		}
	})
	t.Run("AC-1.11", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		w.expectEvents(pat)
	})
	t.Run("AC-1.12", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		b1 := w.event("B1", vocabkey.EventBirth, 1820)
		w.k.Join(pat, b1, "subject")
		w.k.Provisional(pat)
		w.k.Commit(pat)
		w.expectEvents(pat, "B1(subject;)")
	})
}

// located gives an event its own Place.
func (w *kinWorld) located(event []byte) {
	w.k.Location(event, w.k.Entity("place"))
}

// guest adds a co-participant of the event: a person with a dated birth at a
// Place.
func (w *kinWorld) guest(event []byte, name string, year int) {
	g := w.person(name)
	b, _ := w.k.Birth(g, nil, nil)
	w.k.Date(b, vocabkey.PropertyDate, yearValue(year))
	w.located(b)
	w.k.Join(g, event, "subject")
}

func TestPersonEventsQueryCount(t *testing.T) {
	t.Run("AC-2.1", func(t *testing.T) {
		w := newKinWorld(t)
		pat, ned := w.person("Pat"), w.person("Ned")
		b1 := w.event("B1", vocabkey.EventBirth, 1820)
		w.k.Join(pat, b1, "subject")
		u := w.event("U", vocabkey.EventBurial, 1890)
		w.k.Join(ned, u, "subject")
		w.k.Join(pat, u, "witness")
		first, _ := w.events(pat)
		var queries int
		w.c.Graph().SetQueryCounterForTest(func() { queries++ })
		second, _ := w.events(pat)
		if queries != 0 {
			t.Fatalf("a second read issued %d graph queries", queries)
		}
		if len(first) != 2 || !reflect.DeepEqual(first, second) {
			t.Fatalf("second read differs:\n%v\n%v", first, second)
		}
	})
	t.Run("AC-2.2", func(t *testing.T) {
		count := func(events, perEvent, census int) int {
			w := newKinWorld(t)
			pat := w.person("Pat")
			for i := 0; i < events; i++ {
				e := w.event(fmt.Sprintf("E%d", i), vocabkey.EventType("census"), 1800+i)
				w.k.Join(pat, e, "subject")
				w.located(e)
				n := perEvent
				if i == 0 {
					n = census
				}
				for j := 0; j < n; j++ {
					w.guest(e, fmt.Sprintf("G%d.%d", i, j), 1750+j)
				}
			}
			var queries int
			w.c.Graph().SetQueryCounterForTest(func() { queries++ })
			rows, _ := w.events(pat)
			if len(rows) != events {
				t.Fatalf("%d rows, want %d", len(rows), events)
			}
			return queries
		}
		a, b := count(1, 1, 1), count(12, 3, 41)
		if a == 0 || a != b {
			t.Fatalf("queries: small timeline %d, large timeline %d", a, b)
		}
	})
	t.Run("AC-2.3", func(t *testing.T) {
		w := newKinWorld(t)
		pat := w.person("Pat")
		b1 := w.event("B1", vocabkey.EventBirth, 1820)
		w.k.Join(pat, b1, "subject")
		w.expectEvents(pat, "B1(subject;)")
		e := w.event("W", vocabkey.EventBurial, 1890)
		j := w.k.Join(pat, e, "witness")
		w.k.Commit(pat, e, j)
		w.expectEvents(pat, "B1(subject;)", "W(witness;)")
	})
}
