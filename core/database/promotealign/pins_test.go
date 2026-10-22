package promotealign_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/graphalign"
)

func handleRow(s subjects.Subject, handle []byte, defaultPins bool, ids ...string) promotealign.ApplyRow {
	return promotealign.ApplyRow{
		BatchRow:    promote.BatchRow{SubjectID: s.ID, Target: promote.TargetHandle, EntityID: handle},
		DefaultPins: defaultPins, PinIDs: ids,
	}
}

func (f *fixture) resolve(src sources.Source, rows ...promotealign.ApplyRow) []promote.BatchRow {
	f.t.Helper()
	out, err := promotealign.ResolvePins(f.c, src.ID, rows)
	must(f.t, err)
	return out
}

// sexPerson is a Source person who says (or denies) a sex at birth.
func (f *fixture) sexPerson(src sources.Source, art artifacts.Artifact, sex, polarity string) subjects.Subject {
	f.t.Helper()
	s := f.bare(src, "person")
	f.cite(art, s, observations.Input{
		PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", sex).ID, Polarity: polarity,
	})
	return s
}

// T-3: Apply pins what the evidence offers, from the page's final targets.
func TestResolvePins(t *testing.T) {
	f := newFixture(t)
	handle, members := f.namedMembers(5, "John Smith")
	srcB, artB := f.newSource("Obituary")
	incoming := f.person(srcB, artB, "John Smith")
	e := f.evidence(promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: incoming.ID, HandleID: handle})
	offers := promotealign.PinOffers(e)
	if len(offers) != 5 {
		t.Fatalf("offers = %d, want 5", len(offers))
	}
	for _, o := range offers {
		if !o.Checked {
			t.Fatalf("offer %s starts unchecked", o.ID)
		}
	}
	sets := func(os ...promotealign.PinOffer) []promote.PinSet {
		var out []promote.PinSet
		for _, o := range os {
			out = append(out, o.Set)
		}
		return out
	}
	// AC-1.1: one set per member, each the incoming and the member's Observation.
	for i, o := range offers {
		if !bytes.Equal(o.Set.Member, e.Sections[0].Lines[0].Members[i].Subject) || len(o.Set.Observations) != 2 {
			t.Fatalf("offer %d = %+v", i, o)
		}
	}
	for _, m := range members {
		found := false
		for _, o := range offers {
			found = found || bytes.Equal(o.Set.Member, m.ID)
		}
		if !found {
			t.Fatalf("no offer for member %x", m.ID)
		}
	}
	other := offers[0].ID + "x" // an id nothing offers

	t.Run("a row never opened pins every checked offer", func(t *testing.T) {
		got := f.resolve(srcB, handleRow(incoming, handle, true))
		if !reflect.DeepEqual(got[0].PinSets, sets(offers...)) {
			t.Fatalf("sets = %+v, want %+v", got[0].PinSets, sets(offers...))
		}
	})
	t.Run("a caller's pin sets are dropped", func(t *testing.T) {
		row := handleRow(incoming, handle, false)
		row.PinSets = []promote.PinSet{offers[0].Set}
		if got := f.resolve(srcB, row); len(got[0].PinSets) != 0 {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("an opened row pins the ids it kept", func(t *testing.T) {
		got := f.resolve(srcB, handleRow(incoming, handle, false, offers[2].ID))
		if !reflect.DeepEqual(got[0].PinSets, sets(offers[2])) {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("an id the evidence no longer offers is dropped", func(t *testing.T) {
		got := f.resolve(srcB, handleRow(incoming, handle, false, other, offers[1].ID, other))
		if !reflect.DeepEqual(got[0].PinSets, sets(offers[1])) {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("a row that kept nothing pins nothing", func(t *testing.T) {
		got := f.resolve(srcB, handleRow(incoming, handle, false))
		if len(got[0].PinSets) != 0 {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("a conflict line is dropped from the ids and from the defaults", func(t *testing.T) {
		g := newFixture(t)
		male := g.sexPerson(g.source, g.artifact, "male", observations.PolarityPositive)
		h := g.promote(male).Entity.ID
		src, art := g.newSource("Denial")
		denies := g.sexPerson(src, art, "male", observations.PolarityNegative)
		ev := g.evidence(promotealign.EvidenceRequest{SourceID: src.ID, SubjectID: denies.ID, HandleID: h})
		line := ev.Sections[0].Lines[0]
		if line.Outcome != "conflict" || line.Pinnable {
			t.Fatalf("line = %+v", line)
		}
		id := promotealign.LinePinID(line, line.Members[0])
		for _, row := range []promotealign.ApplyRow{handleRow(denies, h, true), handleRow(denies, h, false, id)} {
			if got := g.resolve(src, row); len(got[0].PinSets) != 0 {
				t.Fatalf("sets = %+v", got[0].PinSets)
			}
		}
	})
	t.Run("a New or Skip row with ids is invalid", func(t *testing.T) {
		for _, target := range []string{promote.TargetNew, promote.TargetSkip} {
			row := promotealign.ApplyRow{BatchRow: promote.BatchRow{SubjectID: incoming.ID, Target: target}, PinIDs: []string{offers[0].ID}}
			if _, err := promotealign.ResolvePins(f.c, srcB.ID, []promotealign.ApplyRow{row}); !errors.Is(err, promote.ErrInvalid) {
				t.Errorf("%s: err = %v", target, err)
			}
		}
	})
	t.Run("default pins with ids is invalid", func(t *testing.T) {
		if _, err := promotealign.ResolvePins(f.c, srcB.ID, []promotealign.ApplyRow{handleRow(incoming, handle, true, offers[0].ID)}); !errors.Is(err, promote.ErrInvalid) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("a row already filed on the Source gets no pins", func(t *testing.T) {
		f.join(incoming, handle)
		got := f.resolve(srcB, handleRow(incoming, handle, true))
		if len(got[0].PinSets) != 0 {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
}

// T-4: a relationship's bundle is one pin, with the Observations that say the
// link and the neighbor, and it follows the page's final targets.
func TestResolvePinsBundlesARelationship(t *testing.T) {
	w := newWarmFixture(t)
	f := w.f
	req := w.request()
	e := f.evidence(req)
	offers := promotealign.PinOffers(e)
	if len(e.Relationships) != 1 {
		t.Fatalf("relationships = %d", len(e.Relationships))
	}
	rel := e.Relationships[0]
	bundleID := promotealign.BundlePinID(rel, rel.Members[0])
	var bundle, date *promotealign.PinOffer
	for i, o := range offers {
		if !o.Checked {
			t.Errorf("offer %s starts unchecked", o.ID)
		}
		switch {
		case o.ID == bundleID:
			bundle = &offers[i]
		case bytes.Equal(o.Set.Member, rel.Members[0].Neighbor):
			date = &offers[i]
		}
	}
	if bundle == nil || date == nil {
		t.Fatalf("offers lack the bundle or the date line: %+v", offers)
	}
	want := append(append([][]byte{}, rel.Incoming...), rel.Members[0].Observations...)
	if !bytes.Equal(bundle.Set.Member, w.member.ID) || !reflect.DeepEqual(bundle.Set.Observations, want) {
		t.Errorf("bundle set = %+v, want member %x with %x", bundle.Set, w.member.ID, want)
	}

	rows := func(birth promotealign.ApplyRow) []promotealign.ApplyRow {
		return []promotealign.ApplyRow{handleRow(w.bPerson, w.personHandle, true), birth}
	}
	t.Run("both rows default: the person row pins the bundle and the date", func(t *testing.T) {
		got := f.resolve(w.srcB, rows(handleRow(w.bBirth, w.birthHandle, true))...)
		if !reflect.DeepEqual(got[0].PinSets, func() (out []promote.PinSet) {
			for _, o := range offers {
				out = append(out, o.Set)
			}
			return
		}()) {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("a birth filed New is no neighbor", func(t *testing.T) {
		birth := promotealign.ApplyRow{BatchRow: promote.BatchRow{SubjectID: w.bBirth.ID, Target: promote.TargetNew}}
		got := f.resolve(w.srcB, rows(birth)...)
		for _, set := range got[0].PinSets {
			if bytes.Equal(set.Member, w.member.ID) && len(set.Observations) > 2 {
				t.Fatalf("a bundle was pinned: %+v", set)
			}
			if bytes.Equal(set.Member, rel.Members[0].Neighbor) {
				t.Fatalf("the date was pinned: %+v", set)
			}
		}
		for _, target := range []graphalign.Target{graphalign.TargetNew, graphalign.TargetSkip} {
			r := req
			r.Neighbors = []graphalign.Fixed{{SubjectID: w.bBirth.ID, Target: target}}
			if e := f.evidence(r); len(e.Relationships) != 0 {
				t.Errorf("%s: relationships = %+v", target, e.Relationships)
			}
		}
		r := req
		r.Neighbors = nil
		if e := f.evidence(r); len(e.Relationships) != 0 {
			t.Errorf("left out: relationships = %+v", e.Relationships)
		}
	})
	t.Run("a birth already filed is a neighbor without being named", func(t *testing.T) {
		g := newWarmFixture(t)
		g.f.join(g.bBirth, g.birthHandle)
		r := g.request()
		r.Neighbors = nil
		got := g.f.evidence(r)
		if len(got.Relationships) != 1 || !bytes.Equal(got.Relationships[0].Neighbor, g.bBirth.ID) {
			t.Fatalf("relationships = %+v", got.Relationships)
		}
	})
	t.Run("the offers the RPC sends are the ids Apply resolves", func(t *testing.T) {
		var ids []string
		for _, o := range offers {
			ids = append(ids, o.ID)
		}
		got := f.resolve(w.srcB, handleRow(w.bPerson, w.personHandle, false, ids...), handleRow(w.bBirth, w.birthHandle, true))
		if len(got[0].PinSets) != len(offers) {
			t.Fatalf("sets = %d, want %d", len(got[0].PinSets), len(offers))
		}
	})
}

// T-5: Done reads the page once: it loads the Source once, and issues no graph
// query on a page Propose just read; its sets are what single-row Evidence
// calls offer.
func TestResolvePinsReadsThePageOnce(t *testing.T) {
	f := newFixture(t)
	srcB, artB := f.newSource("Obituary")
	var rows []promotealign.ApplyRow
	var want [][]promote.PinSet
	var subs []subjects.Subject
	var handles [][]byte
	for _, name := range []string{"Anne Brown", "Carl Davis", "Edith Fox"} {
		member := f.person(f.source, f.artifact, name)
		h := f.promote(member).Entity.ID
		in := f.person(srcB, artB, name)
		subs, handles = append(subs, in), append(handles, h)
		rows = append(rows, handleRow(in, h, true))
	}
	for i := range subs {
		req := promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: subs[i].ID, HandleID: handles[i]}
		for j := range subs {
			if j != i {
				req.Neighbors = append(req.Neighbors, graphalign.Fixed{SubjectID: subs[j].ID, HandleID: handles[j]})
			}
		}
		var sets []promote.PinSet
		for _, o := range promotealign.PinOffers(f.evidence(req)) {
			if o.Checked {
				sets = append(sets, o.Set)
			}
		}
		want = append(want, sets)
	}
	var fixed []graphalign.Fixed
	for i := range subs {
		fixed = append(fixed, graphalign.Fixed{SubjectID: subs[i].ID, HandleID: handles[i]})
	}
	f.propose(srcB.ID, fixed)
	n := 0
	f.c.Graph().SetQueryCounterForTest(func() { n++ })
	got := f.resolve(srcB, rows...)
	if n != 0 {
		t.Fatalf("ResolvePins issued %d graph queries", n)
	}
	for i := range got {
		if len(want[i]) == 0 || !reflect.DeepEqual(got[i].PinSets, want[i]) {
			t.Errorf("row %d sets = %+v, want %+v", i, got[i].PinSets, want[i])
		}
	}
}
