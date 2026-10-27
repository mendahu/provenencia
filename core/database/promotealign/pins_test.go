package promotealign_test

import (
	"bytes"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/database"
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

func (f *fixture) resolve(graphID []byte, rows ...promotealign.ApplyRow) []promote.BatchRow {
	f.t.Helper()
	out, err := promotealign.ResolvePins(f.c, graphID, rows)
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
	e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: incoming.ID, HandleID: handle})
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
		got := f.resolve(f.graph(srcB), handleRow(incoming, handle, true))
		if !reflect.DeepEqual(got[0].PinSets, sets(offers...)) {
			t.Fatalf("sets = %+v, want %+v", got[0].PinSets, sets(offers...))
		}
	})
	t.Run("a caller's pin sets are dropped", func(t *testing.T) {
		row := handleRow(incoming, handle, false)
		row.PinSets = []promote.PinSet{offers[0].Set}
		if got := f.resolve(f.graph(srcB), row); len(got[0].PinSets) != 0 {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("an opened row pins the ids it kept", func(t *testing.T) {
		got := f.resolve(f.graph(srcB), handleRow(incoming, handle, false, offers[2].ID))
		if !reflect.DeepEqual(got[0].PinSets, sets(offers[2])) {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("an id the evidence no longer offers is dropped", func(t *testing.T) {
		got := f.resolve(f.graph(srcB), handleRow(incoming, handle, false, other, offers[1].ID, other))
		if !reflect.DeepEqual(got[0].PinSets, sets(offers[1])) {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("a row that kept nothing pins nothing", func(t *testing.T) {
		got := f.resolve(f.graph(srcB), handleRow(incoming, handle, false))
		if len(got[0].PinSets) != 0 {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	conflict := func(t *testing.T) (*fixture, sources.Source, subjects.Subject, []byte, promotealign.PinOffer) {
		g := newFixture(t)
		male := g.sexPerson(g.source, g.artifact, "male", observations.PolarityPositive)
		h := g.promote(male).Entity.ID
		src, art := g.newSource("Denial")
		denies := g.sexPerson(src, art, "male", observations.PolarityNegative)
		ev := g.evidence(promotealign.EvidenceRequest{GraphID: g.graph(src), SubjectID: denies.ID, HandleID: h})
		line := ev.Sections[0].Lines[0]
		if line.Outcome != "conflict" || line.Checked {
			t.Fatalf("line = %+v", line)
		}
		offers := promotealign.PinOffers(ev)
		if len(offers) == 0 {
			t.Fatalf("a conflicting line offers no pin")
		}
		return g, src, denies, h, offers[0]
	}
	t.Run("a conflict line is offered, unchecked", func(t *testing.T) {
		_, _, _, _, offer := conflict(t)
		if offer.Checked {
			t.Fatalf("offer = %+v", offer)
		}
	})
	t.Run("a checked conflict id is kept", func(t *testing.T) {
		g, src, denies, h, offer := conflict(t)
		got := g.resolve(g.graph(src), handleRow(denies, h, false, offer.ID))
		if !reflect.DeepEqual(got[0].PinSets, []promote.PinSet{offer.Set}) {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("the default pins leave a conflict unchecked", func(t *testing.T) {
		g, src, denies, h, _ := conflict(t)
		if got := g.resolve(g.graph(src), handleRow(denies, h, true)); len(got[0].PinSets) != 0 {
			t.Fatalf("sets = %+v", got[0].PinSets)
		}
	})
	t.Run("a New or Skip row with ids is invalid", func(t *testing.T) {
		for _, target := range []string{promote.TargetNew, promote.TargetSkip} {
			row := promotealign.ApplyRow{BatchRow: promote.BatchRow{SubjectID: incoming.ID, Target: target}, PinIDs: []string{offers[0].ID}}
			if _, err := promotealign.ResolvePins(f.c, f.graph(srcB), []promotealign.ApplyRow{row}); !errors.Is(err, promote.ErrInvalid) {
				t.Errorf("%s: err = %v", target, err)
			}
		}
	})
	t.Run("an unknown graph is invalid", func(t *testing.T) {
		row := handleRow(incoming, handle, true)
		if _, err := promotealign.ResolvePins(f.c, make([]byte, 16), []promotealign.ApplyRow{row}); !errors.Is(err, promote.ErrInvalid) {
			t.Errorf("err = %v, want promote.ErrInvalid", err)
		}
	})
	t.Run("default pins with ids is invalid", func(t *testing.T) {
		if _, err := promotealign.ResolvePins(f.c, f.graph(srcB), []promotealign.ApplyRow{handleRow(incoming, handle, true, offers[0].ID)}); !errors.Is(err, promote.ErrInvalid) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("a row already filed on the graph gets no pins", func(t *testing.T) {
		f.join(incoming, handle)
		got := f.resolve(f.graph(srcB), handleRow(incoming, handle, true))
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
		got := f.resolve(f.graph(w.srcB), rows(handleRow(w.bBirth, w.birthHandle, true))...)
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
		got := f.resolve(f.graph(w.srcB), rows(birth)...)
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
		got := f.resolve(f.graph(w.srcB), handleRow(w.bPerson, w.personHandle, false, ids...), handleRow(w.bBirth, w.birthHandle, true))
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
		req := promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: subs[i].ID, HandleID: handles[i]}
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
	f.propose(f.graph(srcB), fixed)
	n := 0
	f.c.Graph().SetQueryCounterForTest(func() { n++ })
	got := f.resolve(f.graph(srcB), rows...)
	if n != 0 {
		t.Fatalf("ResolvePins issued %d graph queries", n)
	}
	for i := range got {
		if len(want[i]) == 0 || !reflect.DeepEqual(got[i].PinSets, want[i]) {
			t.Errorf("row %d sets = %+v, want %+v", i, got[i].PinSets, want[i])
		}
	}
}

// AC-1.2, AC-1.3: likeness lines offer a pin for the same-image pair (checked)
// and the other-image pair (unchecked); a row never opened pins only the first.
func TestResolvePinsLikeness(t *testing.T) {
	l := newLikenessFixture(t, 90)
	e := l.evidence()
	var offers []promotealign.PinOffer
	for _, o := range promotealign.PinOffers(e) {
		for _, sec := range e.Sections {
			if sec.Key != "likeness" {
				continue
			}
			for _, line := range sec.Lines {
				for _, p := range line.Members {
					if o.ID == promotealign.LinePinID(line, p) {
						offers = append(offers, o)
					}
				}
			}
		}
	}
	if len(offers) != 3 {
		t.Fatalf("likeness offers = %d, want 3 (one same-image, two other-image)", len(offers))
	}
	checked := 0
	for _, o := range offers {
		if o.Checked {
			checked++
		}
	}
	if checked != 1 {
		t.Fatalf("checked likeness offers = %d, want 1", checked)
	}
	likeness := section(t, e.Sections, "likeness").Lines[0].Incoming.Ref
	got := l.f.resolve(l.f.graph(l.srcB), handleRow(l.incoming, l.handle, true))
	pinned := map[string]bool{}
	for _, set := range got[0].PinSets {
		for _, o := range set.Observations {
			if bytes.Equal(o, likeness) {
				pinned[string(set.Member)] = true
			}
		}
	}
	if !pinned[string(l.a.ID)] || pinned[string(l.b.ID)] || pinned[string(l.third.ID)] {
		t.Fatalf("default pins of the likeness, by member: %v, want only the image-A member", len(pinned))
	}
}

// The offer of a spouse relationship a Source's marriage implies is checked and
// starts with the relationship's incoming Observations.
func TestPinOffersOfAnImpliedLayerLink(t *testing.T) {
	e := marriageEvidence(t, true).e
	if len(e.Relationships) != 1 || len(e.Relationships[0].Members) != 1 {
		t.Fatalf("relationships %+v, want one spouse relationship with one bundle", e.Relationships)
	}
	rel := e.Relationships[0]
	var offer *promotealign.PinOffer
	for _, o := range promotealign.PinOffers(e) {
		if o.ID == promotealign.BundlePinID(rel, rel.Members[0]) {
			offer = &o
		}
	}
	if offer == nil || !offer.Checked {
		t.Fatalf("no checked offer for the spouse bundle in %+v", promotealign.PinOffers(e))
	}
	if len(offer.Set.Observations) < len(rel.Incoming) || !reflect.DeepEqual(offer.Set.Observations[:len(rel.Incoming)], rel.Incoming) {
		t.Fatalf("offer observations %x, want to start with %x", offer.Set.Observations, rel.Incoming)
	}
}

// Done on a row whose sheet was never opened pins an implied line's bundle like
// any checked relationship.
func TestResolvePinsImpliedCanonLink(t *testing.T) {
	sc := marriageEvidence(t, false)
	f := sc.f
	r := sc.e.Relationships[0]
	want := append(slices.Clone(r.Incoming), r.Members[0].Observations...)
	graphID := f.graph(sc.recordSource)

	t.Run("AC-2.1 default pins", func(t *testing.T) {
		got := f.resolve(graphID, handleRow(sc.bob, sc.bobH, true), handleRow(sc.ann, sc.annH, true))
		var set *promote.PinSet
		for i, s := range got[0].PinSets {
			if bytes.Equal(s.Member, sc.treeBob.ID) && len(s.Observations) == len(want) {
				set = &got[0].PinSets[i]
			}
		}
		if set == nil || !reflect.DeepEqual(set.Observations, want) {
			t.Fatalf("Bob's pin sets = %+v, want a set of tree Bob with %x", got[0].PinSets, want)
		}
	})

	id := promotealign.BundlePinID(r, r.Members[0])
	t.Run("AC-2.2 the offer", func(t *testing.T) {
		wantID := "b:" + database.UUIDString(sc.ann.ID) + ":relationship·spouse:" + database.UUIDString(sc.treeBobPart.ID)
		if id != wantID {
			t.Fatalf("BundlePinID = %q, want %q", id, wantID)
		}
		var offers int
		for _, o := range promotealign.PinOffers(sc.e) {
			if o.ID == id {
				offers++
				if !o.Checked || !reflect.DeepEqual(o.Set.Observations, want) || !bytes.Equal(o.Set.Member, sc.treeBob.ID) {
					t.Errorf("offer = %+v", o)
				}
			}
		}
		if offers != 1 {
			t.Fatalf("offers with id %s = %d, want 1", id, offers)
		}
		got := f.resolve(graphID, handleRow(sc.bob, sc.bobH, false, id), handleRow(sc.ann, sc.annH, true))
		var found bool
		for _, s := range got[0].PinSets {
			found = found || (bytes.Equal(s.Member, sc.treeBob.ID) && reflect.DeepEqual(s.Observations, want))
		}
		if !found {
			t.Fatalf("pin sets %+v lack the bundle's", got[0].PinSets)
		}
	})
}
