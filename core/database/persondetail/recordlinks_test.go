package persondetail_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/conclusiondetails"
	"github.com/mendahu/provenencia/core/database/evidencegraphs"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/persondetail"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/recordlinks"
	"github.com/mendahu/provenencia/core/writes"
)

// register is one Source with a son, and optionally a father, a birth the
// father took part in as its subject, and a mother, each cited with its name.
type register struct {
	son, father, birth, mother subjects.Subject
}

// registerOf files the son and then the people named, one Source titled title.
// A blank name leaves that person out; birth adds the birth the father is the
// subject of.
func (f *fixture) registerOf(title, father, mother string, birth bool) register {
	f.T.Helper()
	src := f.source(title)
	art := f.arts[string(src.ID)]
	r := register{son: f.Person(src, art, "Thomas Smith")}
	if father != "" {
		r.father = f.Person(src, art, father)
	}
	if birth {
		r.birth = f.Birth(src, art)
		f.Participation(src, art, r.father, r.birth, "subject")
	}
	if mother != "" {
		r.mother = f.Person(src, art, mother)
	}
	return r
}

// summary reads a result as text: the Source titles in order, then each pair as
// "A-B: entity/facts/strength ...", entities named by names.
func summary(rl persondetail.PersonRecordLinks, names map[string]string) (graphs []string, pairs []string) {
	for _, g := range rl.Graphs {
		graphs = append(graphs, g.SourceTitle)
	}
	for _, p := range rl.Pairs {
		var ls []string
		for _, l := range p.Lines {
			ls = append(ls, fmt.Sprintf("%s/%d/%s", names[string(l.Entity)], l.Facts, l.Strength))
		}
		pairs = append(pairs, fmt.Sprintf("%s-%s: %s", rl.Graphs[p.A].SourceTitle, rl.Graphs[p.B].SourceTitle, strings.Join(ls, " ")))
	}
	return graphs, pairs
}

func (f *fixture) recordLinks(id []byte) persondetail.PersonRecordLinks {
	f.T.Helper()
	rl, err := persondetail.RecordLinks(f.C, id)
	must(f.T, err)
	return rl
}

// world is the base of every TestRecordLinks row: the four registers, the son
// promoted on all of them, the father on A and B, the mother on A and C, and the
// birth on A and B.
type world struct {
	s, father, mother, birth canonicalentities.Entity
	// names reads an entity in a summary.
	names map[string]string
}

// newWorld files the four registers on f.
func newWorld(f *fixture) world {
	a := f.registerOf("A register", "William Smith", "Mary Ann Jones", true)
	b := f.registerOf("B register", "William Smith", "", true)
	c := f.registerOf("C register", "", "Mary Jones", false)
	d := f.registerOf("D register", "", "", false)
	w := world{s: f.promote(a.son, nil)}
	for _, r := range []register{b, c, d} {
		f.promote(r.son, w.s.ID)
	}
	w.father = f.promote(a.father, nil)
	f.promote(b.father, w.father.ID)
	w.birth = f.promote(a.birth, nil)
	f.promote(b.birth, w.birth.ID)
	w.mother = f.promote(a.mother, nil)
	f.promote(c.mother, w.mother.ID)
	w.names = map[string]string{string(w.s.ID): "S", string(w.father.ID): "F", string(w.mother.ID): "M"}
	return w
}

func TestRecordLinks(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T, f *fixture, w world)
	}{
		{"AC-1.1 and AC-1.2", func(t *testing.T, f *fixture, w world) {
			graphs, pairs := summary(f.recordLinks(w.s.ID), w.names)
			if want := []string{"A register", "B register", "C register", "D register"}; !reflect.DeepEqual(graphs, want) {
				t.Fatalf("graphs %v, want %v", graphs, want)
			}
			if want := []string{"A register-B register: F/2/strong", "A register-C register: M/1/weak"}; !reflect.DeepEqual(pairs, want) {
				t.Fatalf("pairs %v, want %v (no line names S)", pairs, want)
			}
		}},
		{"AC-1.3 a member with no accepted claim draws no line", func(t *testing.T, f *fixture, w world) {
			e1 := f.registerOf("E register", "William Smith", "", false)
			f.promote(e1.son, w.s.ID)
			f.provisional(e1.father, w.father)
			e2 := f.registerOf("F register", "William Smith", "", false)
			f.promote(e2.son, w.s.ID)
			graphs, pairs := summary(f.recordLinks(w.s.ID), w.names)
			if len(graphs) != 6 {
				t.Fatalf("graphs %v, want six", graphs)
			}
			if want := []string{"A register-B register: F/2/strong", "A register-C register: M/1/weak"}; !reflect.DeepEqual(pairs, want) {
				t.Fatalf("pairs %v, want %v", pairs, want)
			}
		}},
		{"AC-1.5 two Subjects of one graph list it once", func(t *testing.T, f *fixture, w world) {
			src := f.source("G register")
			second := f.Person(src, f.arts[string(src.ID)], "Tom Smith")
			first := f.Person(src, f.arts[string(src.ID)], "Thomas Smith")
			f.promote(first, w.s.ID)
			f.promote(second, w.s.ID)
			n := 0
			for _, g := range f.recordLinks(w.s.ID).Graphs {
				if g.SourceTitle == "G register" {
					n++
				}
			}
			if n != 1 {
				t.Fatalf("G register listed %d times", n)
			}
		}},
		{"AC-1.6 an id that is no Person", func(t *testing.T, f *fixture, w world) {
			for name, id := range map[string][]byte{
				"event handle":  w.birth.ID,
				"random id":     {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 99},
				"short id":      {1, 2, 3},
				"unknown nil":   nil,
				"merged person": mergedHandle(t, f),
			} {
				if _, err := persondetail.RecordLinks(f.C, id); !errors.Is(err, conclusiondetails.ErrNotFound) {
					t.Errorf("%s: error %v, want ErrNotFound", name, err)
				}
			}
		}},
		{"AC-2.6 a warm repeat issues no graph-cache query", func(t *testing.T, f *fixture, w world) {
			f.recordLinks(w.s.ID)
			var n int
			f.C.Graph().SetQueryCounterForTest(func() { n++ })
			defer f.C.Graph().SetQueryCounterForTest(nil)
			f.recordLinks(w.s.ID)
			if n != 0 {
				t.Fatalf("the warm read issued %d queries", n)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFixture(t)
			c.run(t, f, newWorld(f))
		})
	}
}

// mergedHandle is a Person handle merged into another.
func mergedHandle(t *testing.T, f *fixture) []byte {
	t.Helper()
	src := f.source("Merged register")
	art := f.arts[string(src.ID)]
	from := f.promote(f.Person(src, art, "Gone Person"), nil)
	into := f.promote(f.Person(src, art, "Kept Person"), nil)
	db, err := f.C.DB()
	must(t, err)
	_, err = db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, into.ID, from.ID)
	must(t, err)
	f.C.Graph().Drop()
	return from.ID
}

func TestRecordLinksGraphOrder(t *testing.T) {
	f := newFixture(t)
	var s canonicalentities.Entity
	onto := func(sub subjects.Subject) {
		if s.ID == nil {
			s = f.promote(sub, nil)
		} else {
			f.promote(sub, s.ID)
		}
	}
	for _, title := range []string{"Census 1851", "Baptisms", "Census 1841"} {
		src := f.source(title)
		g := evidencegraphstest.SoleGraph(t, f.C, src.ID)
		if title == "Baptisms" {
			_, _, err := writes.Run(f.C, writes.Op{Action: "update_evidence_graph", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				ch, err := evidencegraphs.Update(tx, userID, g.ID, evidencegraphs.UpdateInput{Name: "Register"})
				return struct{}{}, ch, err
			})
			must(t, err)
			index := evidencegraphstest.Create(t, f.C, userID, src.ID, "Index", f.arts[string(src.ID)].ID)
			onto(f.person(index.ID, "Thomas"))
		}
		onto(f.person(g.ID, "Thomas"))
	}
	var got []string
	for _, g := range f.recordLinks(s.ID).Graphs {
		got = append(got, g.SourceTitle+" / "+g.GraphName)
	}
	want := []string{"Baptisms / Index", "Baptisms / Register", "Census 1841 / Default", "Census 1851 / Default"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("graphs %q, want %q", got, want)
	}
}

func TestRecordLinksFewGraphs(t *testing.T) {
	f := newFixture(t)
	one := f.registerOf("One register", "William Smith", "", false)
	s := f.promote(one.son, nil)
	f.promote(one.father, nil)
	if rl := f.recordLinks(s.ID); len(rl.Graphs) != 1 || len(rl.Pairs) != 0 {
		t.Fatalf("one accepted member: %d graphs, %d pairs", len(rl.Graphs), len(rl.Pairs))
	}
	two := f.registerOf("Two register", "Peter Jones", "", false)
	f.promote(two.son, s.ID)
	f.promote(two.father, nil)
	if rl := f.recordLinks(s.ID); len(rl.Graphs) != 2 || len(rl.Pairs) != 0 {
		t.Fatalf("two graphs naming no one else: %d graphs, %d pairs", len(rl.Graphs), len(rl.Pairs))
	}
}

// TestRecordLinksSaysWhenPartial: the read says when it did not count or
// dropped lines (S16-01 AC-2.4, AC-2.5), on four household graphs.
func TestRecordLinksSaysWhenPartial(t *testing.T) {
	f := newFixture(t)
	son := households(f, 4) // 6 pairs of 3 strong lines: 18 lines
	read := func(maxReadWork, most int) persondetail.PersonRecordLinks {
		t.Helper()
		cfg := matchregistry.Default().RecordLinksConfig()
		if maxReadWork > 0 {
			cfg.RecordLinks.MaxReadWork = maxReadWork
		}
		rl, err := persondetail.RecordLinksWith(f.C, son, cfg, most)
		must(t, err)
		return rl
	}
	lines := func(rl persondetail.PersonRecordLinks) (n, strong int) {
		for _, p := range rl.Pairs {
			for _, l := range p.Lines {
				n++
				if l.Strength == recordlinks.StrengthStrong {
					strong++
				}
			}
		}
		return n, strong
	}
	t.Run("an ordinary read", func(t *testing.T) {
		rl := read(0, 75_000)
		if n, strong := lines(rl); rl.NotCounted || rl.LinesDropped || n != 18 || strong != 18 {
			t.Fatalf("not counted %v, dropped %v, %d lines, %d strong", rl.NotCounted, rl.LinesDropped, n, strong)
		}
	})
	t.Run("past the work bound by pair work", func(t *testing.T) {
		rl := read(18, 75_000)
		if !rl.NotCounted || rl.LinesDropped || len(rl.Pairs) != 0 || len(rl.Graphs) != 4 {
			t.Fatalf("not counted %v, dropped %v, %d pairs, %d graphs", rl.NotCounted, rl.LinesDropped, len(rl.Pairs), len(rl.Graphs))
		}
	})
	t.Run("past the work bound by the lines alone", func(t *testing.T) {
		rl := read(1, 75_000)
		if !rl.NotCounted || rl.LinesDropped || len(rl.Pairs) != 0 {
			t.Fatalf("not counted %v, dropped %v, %d pairs", rl.NotCounted, rl.LinesDropped, len(rl.Pairs))
		}
	})
	t.Run("past the line cap", func(t *testing.T) {
		rl := read(0, 5)
		if n, _ := lines(rl); rl.NotCounted || !rl.LinesDropped || n != 5 {
			t.Fatalf("not counted %v, dropped %v, %d lines", rl.NotCounted, rl.LinesDropped, n)
		}
	})
}
