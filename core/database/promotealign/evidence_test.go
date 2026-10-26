package promotealign_test

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/observations/observationstest"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func (f *fixture) evidence(req promotealign.EvidenceRequest) claimevidence.Evidence {
	f.t.Helper()
	e, err := promotealign.Evidence(f.c, req)
	must(f.t, err)
	return e
}

// section returns the section with key, or fails.
func section(t testing.TB, ss []claimevidence.Section, key vocabkey.PropertyKey) claimevidence.Section {
	t.Helper()
	for _, s := range ss {
		if s.Key == key {
			return s
		}
	}
	t.Fatalf("no %s section in %+v", key, ss)
	return claimevidence.Section{}
}

// T-5: five accepted members who all say "John Smith" are one line with five
// pins, however many members the entity has.
func TestEvidenceCollapsesFiveNamedMembers(t *testing.T) {
	f := newFixture(t)
	first := f.person(f.source, f.artifact, "John Smith")
	handle := f.promote(first).Entity.ID
	for i := 0; i < 4; i++ {
		f.join(f.person(f.source, f.artifact, "John Smith"), handle)
	}
	srcB, artB := f.newSource("Obituary")
	incoming := f.person(srcB, artB, "John Smith")

	e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: incoming.ID, HandleID: handle})
	s := section(t, e.Sections, vocabkey.PropertyName)
	if len(s.Lines) != 1 {
		t.Fatalf("name lines = %d, want 1: %+v", len(s.Lines), s.Lines)
	}
	l := s.Lines[0]
	want := matchregistry.Default().Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	if l.Relation != valuecompare.Same || !l.Checked || len(l.Members) != 5 || l.Points != want {
		t.Errorf("line = %+v", l)
	}
	if len(l.Incoming.Ref) != 16 {
		t.Errorf("incoming ref = %x", l.Incoming.Ref)
	}
	for _, p := range l.Members {
		if len(p.Observation) != 16 || len(p.Subject) != 16 {
			t.Errorf("pin = %+v", p)
		}
	}
}

// T-6: a birth that corresponds is one relationship that bundles the
// bridges behind it; its dates are compared with that birth only.
func TestEvidenceBundlesAMatchingBirth(t *testing.T) {
	for _, tc := range []struct {
		name        string
		year        int
		wantState   propertycompare.Outcome
		checked     bool
		dateLine    propertycompare.Outcome
		dateChecked bool
	}{
		{"AC-2.1 the same year agrees", 1817, propertycompare.OutcomeAgree, true, propertycompare.OutcomeAgree, true},
		{"AC-2.2 a different year is partial", 1830, propertycompare.OutcomePartial, false, propertycompare.OutcomeConflict, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			aPerson := f.person(f.source, f.artifact, "John Smith")
			aBirth := f.datedBirth(f.source, f.artifact, 1817)
			f.participation(f.source, f.artifact, aPerson, aBirth, "subject")
			personHandle := f.promote(aPerson).Entity.ID
			birthHandle := f.promote(aBirth).Entity.ID

			srcB, artB := f.newSource("Obituary")
			bPerson := f.person(srcB, artB, "John Smith")
			bBirth := f.datedBirth(srcB, artB, tc.year)
			f.participation(srcB, artB, bPerson, bBirth, "subject")

			e := f.evidence(promotealign.EvidenceRequest{
				GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle,
				Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}},
			})
			if len(e.Relationships) != 1 {
				t.Fatalf("relationships = %d, want 1: %+v", len(e.Relationships), e.Relationships)
			}
			r := e.Relationships[0]
			wantKey := "participation·subject·person·birth"
			if r.Key.String() != wantKey || r.Outcome != tc.wantState || r.Checked != tc.checked || !bytes.Equal(r.Neighbor, bBirth.ID) {
				t.Errorf("relationship = %+v, want key %s outcome %s checked %v", r, wantKey, tc.wantState, tc.checked)
			}
			if len(r.Incoming) < 2 || len(r.Members) != 1 || len(r.Members[0].Observations) < 2 {
				t.Errorf("bundles: incoming %d, members %+v", len(r.Incoming), r.Members)
			}
			if !bytes.Equal(r.Members[0].Member, aPerson.ID) || !bytes.Equal(r.Members[0].Neighbor, aBirth.ID) {
				t.Errorf("bundle ends = %x %x", r.Members[0].Member, r.Members[0].Neighbor)
			}
			d := section(t, r.Sections, vocabkey.PropertyDate)
			if len(d.Lines) != 1 || d.Lines[0].Outcome != tc.dateLine || d.Lines[0].Checked != tc.dateChecked {
				t.Errorf("date lines = %+v", d.Lines)
			}
			// The defining event_type is bundled, never compared.
			for _, s := range r.Sections {
				if s.Key == vocabkey.PropertyEventType {
					t.Errorf("event_type compared under the relationship")
				}
			}
		})
	}
}

func TestEvidenceInvalidRequest(t *testing.T) {
	f := newFixture(t)
	aPerson := f.person(f.source, f.artifact, "John Smith")
	handle := f.promote(aPerson).Entity.ID
	srcB, artB := f.newSource("Obituary")
	b := f.person(srcB, artB, "John Smith")
	bBirth := f.datedBirth(srcB, artB, 1817)
	other := f.birth(f.source, f.artifact)
	birthHandle := f.promote(other).Entity.ID
	missing := bytes.Repeat([]byte{0xEE}, 16)

	for _, tc := range []struct {
		name string
		req  promotealign.EvidenceRequest
	}{
		{"no such graph", promotealign.EvidenceRequest{GraphID: missing, SubjectID: b.ID, HandleID: handle}},
		{"a Subject of another graph", promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: aPerson.ID, HandleID: handle}},
		{"an empty row handle", promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: b.ID}},
		{"a missing row handle", promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: b.ID, HandleID: missing}},
		{"a row handle of another kind", promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: b.ID, HandleID: birthHandle}},
		{"a missing neighbor handle", promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: b.ID, HandleID: handle,
			Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: missing}}}},
		{"a neighbor handle of another kind", promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: b.ID, HandleID: handle,
			Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: handle}}}},
		{"more neighbors than primary Subjects", promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: b.ID, HandleID: handle,
			Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}, {SubjectID: bBirth.ID, HandleID: birthHandle}, {SubjectID: bBirth.ID, HandleID: birthHandle}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := promotealign.Evidence(f.c, tc.req); !errors.Is(err, promote.ErrInvalid) {
				t.Errorf("err = %v, want promote.ErrInvalid", err)
			}
		})
	}
	if _, err := promotealign.Evidence(nil, promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: b.ID, HandleID: handle}); !errors.Is(err, promote.ErrInvalid) {
		t.Errorf("nil catalog: err = %v, want promote.ErrInvalid", err)
	}
}

// filedBirth files a person and his dated birth on src, joined by a subject
// participation (year 0: no date), joins them to the given entities or
// promotes them when an entity is nil, and returns them with their entities.
func (f *fixture) filedBirth(src sources.Source, art artifacts.Artifact, name string, year int, personHandle, birthHandle []byte) (p, b subjects.Subject, ph, bh []byte) {
	f.t.Helper()
	p = f.person(src, art, name)
	if year == 0 {
		b = f.birth(src, art)
	} else {
		b = f.datedBirth(src, art, year)
	}
	f.participation(src, art, p, b, "subject")
	ph, bh = personHandle, birthHandle
	if ph == nil {
		ph = f.promote(p).Entity.ID
	} else {
		f.join(p, ph)
	}
	if bh == nil {
		bh = f.promote(b).Entity.ID
	} else {
		f.join(b, bh)
	}
	return p, b, ph, bh
}

// T-7: a sibling in the record never meets the spouse the entity has.
func TestEvidenceSpouseNeverMeetsSibling(t *testing.T) {
	for _, tc := range []struct {
		name, typ string
		want      int
	}{
		{"AC-2.3 a sibling is not a spouse", "sibling", 0},
		{"control: a spouse is", "spouse", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			p := f.person(f.source, f.artifact, "John Smith")
			s1 := f.person(f.source, f.artifact, "Mary Jones")
			f.relationship(f.source, f.artifact, p, s1, "spouse")
			personHandle := f.promote(p).Entity.ID
			spouseHandle := f.promote(s1).Entity.ID

			srcB, artB := f.newSource("Obituary")
			bPerson := f.person(srcB, artB, "John Smith")
			bOther := f.person(srcB, artB, "Mary Jones")
			f.relationship(srcB, artB, bPerson, bOther, tc.typ)

			e := f.evidence(promotealign.EvidenceRequest{
				GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle,
				Neighbors: []graphalign.Fixed{{SubjectID: bOther.ID, HandleID: spouseHandle}},
			})
			if len(e.Relationships) != tc.want {
				t.Errorf("relationships = %+v, want %d", e.Relationships, tc.want)
			}
		})
	}
}

// A relationship carries the link read from the neighbor's side: two links to
// one neighbor are two relationships, each with its own term, and a directed
// link names the term the neighbor carries.
func TestEvidenceReadsEachLinkFromTheNeighbor(t *testing.T) {
	type read struct {
		end   vocabkey.PropertyKey
		term  vocabkey.BridgeTerm
		prop  vocabkey.PropertyKey // the property the term belongs to
		names bool
	}
	got := func(e claimevidence.Evidence) map[string]read {
		out := map[string]read{}
		for _, r := range e.Relationships {
			out[r.Key.String()] = read{r.NeighborEnd, r.NeighborTerm, r.NeighborTermProperty, r.NeighborTermNamesNeighbor}
		}
		return out
	}
	t.Run("spouses who are also cousins give two cards", func(t *testing.T) {
		f := newFixture(t)
		john := f.person(f.source, f.artifact, "John Smith")
		mary := f.person(f.source, f.artifact, "Mary Jones")
		f.relationship(f.source, f.artifact, john, mary, "spouse")
		f.relationship(f.source, f.artifact, john, mary, "cousin")
		johnHandle, maryHandle := f.promote(john).Entity.ID, f.promote(mary).Entity.ID

		srcB, artB := f.newSource("Obituary")
		bJohn, bMary := f.person(srcB, artB, "John Smith"), f.person(srcB, artB, "Mary Jones")
		f.relationship(srcB, artB, bJohn, bMary, "spouse")
		f.relationship(srcB, artB, bJohn, bMary, "cousin")

		e := f.evidence(promotealign.EvidenceRequest{
			GraphID: f.graph(srcB), SubjectID: bJohn.ID, HandleID: johnHandle,
			Neighbors: []graphalign.Fixed{{SubjectID: bMary.ID, HandleID: maryHandle}},
		})
		want := map[string]read{
			"relationship·spouse": {"", "spouse", "relationship_type", true},
			"relationship·cousin": {"", "cousin", "relationship_type", true},
		}
		if g := got(e); !reflect.DeepEqual(g, want) {
			t.Errorf("reads = %+v, want %+v", g, want)
		}
	})
	t.Run("a parent reads as a parent from the child's sheet and a child from the parent's", func(t *testing.T) {
		f := newFixture(t)
		john := f.person(f.source, f.artifact, "John Smith")
		mary := f.person(f.source, f.artifact, "Mary Jones")
		f.relationship(f.source, f.artifact, mary, john, "parent") // Mary is the parent of John
		johnHandle, maryHandle := f.promote(john).Entity.ID, f.promote(mary).Entity.ID

		srcB, artB := f.newSource("Obituary")
		bJohn, bMary := f.person(srcB, artB, "John Smith"), f.person(srcB, artB, "Mary Jones")
		f.relationship(srcB, artB, bMary, bJohn, "parent")

		asChild := f.evidence(promotealign.EvidenceRequest{
			GraphID: f.graph(srcB), SubjectID: bJohn.ID, HandleID: johnHandle,
			Neighbors: []graphalign.Fixed{{SubjectID: bMary.ID, HandleID: maryHandle}},
		})
		if g, want := got(asChild), map[string]read{"relationship·child·person": {vocabkey.PropertyRelatedTo, "parent", "relationship_type", true}}; !reflect.DeepEqual(g, want) {
			t.Errorf("from the child = %+v, want %+v", g, want)
		}
		asParent := f.evidence(promotealign.EvidenceRequest{
			GraphID: f.graph(srcB), SubjectID: bMary.ID, HandleID: maryHandle,
			Neighbors: []graphalign.Fixed{{SubjectID: bJohn.ID, HandleID: johnHandle}},
		})
		if g, want := got(asParent), map[string]read{"relationship·child·related_to": {vocabkey.PropertyPerson, "child", "relationship_type", true}}; !reflect.DeepEqual(g, want) {
			t.Errorf("from the parent = %+v, want %+v", g, want)
		}
	})
}

// T-8: only a neighbor the page placed on an entity, and that is a layer
// neighbor of the row's Subject, can give a relationship.
func TestEvidenceOnlyMappedNeighbors(t *testing.T) {
	f := newFixture(t)
	_, _, personHandle, birthHandle := f.filedBirth(f.source, f.artifact, "John Smith", 1817, nil, nil)
	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "John Smith")
	bBirth := f.datedBirth(srcB, artB, 1817)
	stranger := f.datedBirth(srcB, artB, 1817) // on the Source, joined to no one
	f.participation(srcB, artB, bPerson, bBirth, "subject")

	mapped := graphalign.Fixed{SubjectID: bBirth.ID, HandleID: birthHandle}
	for _, tc := range []struct {
		name      string
		neighbors []graphalign.Fixed
		want      int
	}{
		{"control: mapped", []graphalign.Fixed{mapped}, 1},
		{"New", []graphalign.Fixed{{SubjectID: bBirth.ID, Target: graphalign.TargetNew}}, 0},
		{"Skip", []graphalign.Fixed{{SubjectID: bBirth.ID, Target: graphalign.TargetSkip}}, 0},
		{"absent", nil, 0},
		{"not a layer neighbor", []graphalign.Fixed{{SubjectID: stranger.ID, HandleID: birthHandle}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := promotealign.Evidence(f.c, promotealign.EvidenceRequest{
				GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle, Neighbors: tc.neighbors,
			})
			must(t, err)
			if len(e.Relationships) != tc.want {
				t.Errorf("relationships = %d, want %d", len(e.Relationships), tc.want)
			}
		})
	}
}

// T-9: two members of the entity, each the subject of a birth that is a member
// of the neighbor's entity, give two bundles under one relationship; a birth
// with no date has nothing to compare, and the relationship stands alone.
func TestEvidenceRelationshipBundles(t *testing.T) {
	for _, tc := range []struct {
		name         string
		year         int
		bundles      int
		wantSections bool
	}{
		{"AC-2.5 two members, two bundles", 1817, 2, true},
		{"AC-2.6 no dates, no section", 0, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			_, _, personHandle, birthHandle := f.filedBirth(f.source, f.artifact, "John Smith", tc.year, nil, nil)
			if tc.bundles == 2 {
				f.filedBirth(f.source, f.artifact, "John Smith", tc.year, personHandle, birthHandle)
			}
			srcB, artB := f.newSource("Obituary")
			bPerson := f.person(srcB, artB, "John Smith")
			var bBirth subjects.Subject
			if tc.year == 0 {
				bBirth = f.birth(srcB, artB)
			} else {
				bBirth = f.datedBirth(srcB, artB, tc.year)
			}
			f.participation(srcB, artB, bPerson, bBirth, "subject")

			e := f.evidence(promotealign.EvidenceRequest{
				GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle,
				Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}},
			})
			if len(e.Relationships) != 1 {
				t.Fatalf("relationships = %d, want 1", len(e.Relationships))
			}
			r := e.Relationships[0]
			if len(r.Members) != tc.bundles || r.Outcome != propertycompare.OutcomeAgree || !r.Checked {
				t.Errorf("relationship = %+v, want %d bundles, agree, checked", r, tc.bundles)
			}
			if got := len(r.Sections) > 0; got != tc.wantSections {
				t.Errorf("sections = %+v, want some: %v", r.Sections, tc.wantSections)
			}
		})
	}
}

// T-14: every Source a line or bundle names is in Sources, with its title.
func TestEvidenceNamesEverySource(t *testing.T) {
	f := newFixture(t)
	_, _, personHandle, birthHandle := f.filedBirth(f.source, f.artifact, "John Smith", 1817, nil, nil)
	census, censusArt := f.newSource("Census 1851")
	f.join(f.person(census, censusArt, "John Smith"), personHandle)
	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "John Smith")
	bBirth := f.datedBirth(srcB, artB, 1817)
	f.participation(srcB, artB, bPerson, bBirth, "subject")

	e := f.evidence(promotealign.EvidenceRequest{
		GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle,
		Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}},
	})
	named := map[string]bool{}
	var walk func(ss []claimevidence.Section)
	walk = func(ss []claimevidence.Section) {
		for _, s := range ss {
			for _, l := range s.Lines {
				for _, p := range l.Members {
					named[string(p.Source)] = true
				}
			}
		}
	}
	walk(e.Sections)
	for _, r := range e.Relationships {
		for _, b := range r.Members {
			named[string(b.Source)] = true
		}
		walk(r.Sections)
	}
	want := map[string]string{string(f.source.ID): "Register", string(census.ID): "Census 1851"}
	for id := range named {
		if want[id] == "" {
			t.Errorf("a line names Source %x that the test did not expect", id)
		}
	}
	for id, title := range want {
		if !named[id] {
			t.Errorf("no line or bundle names %q", title)
		}
		if e.Sources[id] != title {
			t.Errorf("Sources[%x] = %q, want %q", id, e.Sources[id], title)
		}
	}
}

// One Evidence request reads each Evidence graph once, whatever the cap on
// loaded graphs: the row's graph and both member graphs are read by the
// reader's first pass, and the Source titles read the member graphs again.
func TestEvidenceReadsEachGraphOnce(t *testing.T) {
	f := newFixture(t)
	_, _, personHandle, birthHandle := f.filedBirth(f.source, f.artifact, "John Smith", 1817, nil, nil)
	census, censusArt := f.newSource("Census 1851")
	f.join(f.person(census, censusArt, "John Smith"), personHandle)
	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "John Smith")
	bBirth := f.datedBirth(srcB, artB, 1817)
	f.participation(srcB, artB, bPerson, bBirth, "subject")
	req := promotealign.EvidenceRequest{
		GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle,
		Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}},
	}

	run := func() (int, claimevidence.Evidence) {
		f.c.Graph().Drop()
		n := 0
		f.c.Graph().SetQueryCounterForTest(func() { n++ })
		defer f.c.Graph().SetQueryCounterForTest(nil)
		e := f.evidence(req)
		return n, e
	}
	defaultQueries, want := run()
	f.c.Graph().SetEvidenceGraphCapForTest(1)
	defer f.c.Graph().SetEvidenceGraphCapForTest(0)
	oneQueries, got := run()
	if oneQueries != defaultQueries {
		t.Errorf("a cap of one issued %d graph queries, the default cap %d: a graph was read twice", oneQueries, defaultQueries)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("a cap of one changed the response:\n got %+v\nwant %+v", got, want)
	}
}

// A filed bridge gives a bundle only when each end is one positive Observation
// naming one Subject. The catalog's writes refuse the other shapes, so the
// test rewrites the filed bridge's endpoint rows.
func TestEvidenceBridgeEndpoints(t *testing.T) {
	for _, tc := range []struct {
		name    string
		bundles int
		mutate  func(t *testing.T, db *sql.DB, f *fixture, person, event, other subjects.Subject)
	}{
		{"a positive single-valued bridge gives a bundle", 2, func(*testing.T, *sql.DB, *fixture, subjects.Subject, subjects.Subject, subjects.Subject) {}},
		{"a neighbor end that is not positive gives no bundle", 1, func(t *testing.T, db *sql.DB, f *fixture, _, event, _ subjects.Subject) {
			_, err := db.Exec(`UPDATE observations SET polarity = 'negative' WHERE value_subject_id = ? AND property_id = ?`, event.ID, f.prop("event").ID)
			must(t, err)
		}},
		{"a member end that names two Subjects gives no bundle", 1, func(t *testing.T, db *sql.DB, f *fixture, person, _, other subjects.Subject) {
			_, err := db.Exec(`INSERT INTO observations (id, ref, citation_id, subject_id, property_id, polarity, value_subject_id)
				SELECT x'ffffffffffffffffffffffffffffffff', 'OBS-' || hex(randomblob(5)), citation_id, subject_id, property_id, polarity, ?
				FROM observations WHERE value_subject_id = ? AND property_id = ?`, other.ID, person.ID, f.prop("person").ID)
			must(t, err)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			// The rewritten bridge is filed before anything reads the Source into
			// the graph; a second, good one keeps the link between the entities.
			p := f.person(f.source, f.artifact, "John Smith")
			b := f.datedBirth(f.source, f.artifact, 1817)
			other := f.person(f.source, f.artifact, "Jane Doe")
			f.participation(f.source, f.artifact, p, b, "subject")
			db, err := f.c.DB()
			must(t, err)
			tc.mutate(t, db, f, p, b, other)
			_, _, personHandle, birthHandle := f.filedBirth(f.source, f.artifact, "John Smith", 1817, nil, nil)
			f.join(p, personHandle)
			f.join(b, birthHandle)

			srcB, artB := f.newSource("Obituary")
			bPerson := f.person(srcB, artB, "John Smith")
			bBirth := f.datedBirth(srcB, artB, 1817)
			f.participation(srcB, artB, bPerson, bBirth, "subject")

			e := f.evidence(promotealign.EvidenceRequest{
				GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle,
				Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}},
			})
			got := 0
			for _, r := range e.Relationships {
				got += len(r.Members)
			}
			if got != tc.bundles {
				t.Errorf("bundles = %d, want %d: %+v", got, tc.bundles, e.Relationships)
			}
		})
	}
}

// A neighbor entity's members are read only when a filed bridge reaches them.
func TestEvidenceLoadsOnlyBridgedNeighborMembers(t *testing.T) {
	a, b, c := bytes.Repeat([]byte{1}, 16), bytes.Repeat([]byte{2}, 16), bytes.Repeat([]byte{3}, 16)
	links := []claimevidence.CanonLink{{Bridges: []claimevidence.Bridge{{Neighbor: c}, {Neighbor: a}}}, {}}
	got := promotealign.BridgedOnly([][]byte{a, b, c}, links)
	if len(got) != 2 || !bytes.Equal(got[0], a) || !bytes.Equal(got[1], c) {
		t.Errorf("bridgedOnly = %x, want [a c] in member order", got)
	}
	if got := promotealign.BridgedOnly([][]byte{a, b}, nil); len(got) != 0 {
		t.Errorf("no links: bridgedOnly = %x, want none", got)
	}

	// A second, unbridged birth in the neighbor entity changes nothing.
	f := newFixture(t)
	_, _, personHandle, birthHandle := f.filedBirth(f.source, f.artifact, "John Smith", 1817, nil, nil)
	f.join(f.datedBirth(f.source, f.artifact, 1900), birthHandle)
	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "John Smith")
	bBirth := f.datedBirth(srcB, artB, 1817)
	f.participation(srcB, artB, bPerson, bBirth, "subject")
	e := f.evidence(promotealign.EvidenceRequest{
		GraphID: f.graph(srcB), SubjectID: bPerson.ID, HandleID: personHandle,
		Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}},
	})
	if len(e.Relationships) != 1 || len(e.Relationships[0].Members) != 1 {
		t.Fatalf("relationships = %+v, want one with one bundle", e.Relationships)
	}
	for _, s := range e.Relationships[0].Sections {
		for _, l := range s.Lines {
			if len(l.Members) != 1 {
				t.Errorf("%s line pins %d members, want only the bridged birth", s.Key, len(l.Members))
			}
		}
	}
}

// A member of the handle may sit on any graph of its Source: the evidence reads
// the values the member's own graph records, and names the graph's Source.
func TestEvidenceReadsMemberGraph(t *testing.T) {
	f := newFixture(t)
	census, _ := f.newSource("Census 1851")
	_, jones, page2 := artifactstest.Split(t, f.c, userID, census.ID, "Jones")
	member := subjectstest.On(t, f.c, userID, jones.ID, "person", "")
	f.cite(page2, member, observations.Input{PropertyID: f.prop("name").ID, Name: namevaluetest.Western("John Smith")})
	handle := f.promote(member).Entity.ID

	srcB, artB := f.newSource("Obituary")
	incoming := f.person(srcB, artB, "John Smith")
	e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: incoming.ID, HandleID: handle})

	s := section(t, e.Sections, vocabkey.PropertyName)
	if len(s.Lines) != 1 || len(s.Lines[0].Members) != 1 {
		t.Fatalf("name lines %+v", s.Lines)
	}
	l := s.Lines[0]
	if l.Relation != valuecompare.Same {
		t.Errorf("relation %v, want Same: the member's own graph records John Smith", l.Relation)
	}
	p := l.Members[0]
	if !bytes.Equal(p.Subject, member.ID) || !bytes.Equal(p.Source, census.ID) {
		t.Errorf("pin %+v, want Subject %x on Source %x", p, member.ID, census.ID)
	}
	if got := e.Sources[string(census.ID)]; got != "Census 1851" {
		t.Errorf("Sources[census] = %q", got)
	}
}

// A row Subject must be a primary Subject of the request's graph, not just of
// its Source.
func TestEvidenceRefusesASubjectOfAnotherGraph(t *testing.T) {
	f := newFixture(t)
	h := f.promote(f.person(f.source, f.artifact, "John Smith")).Entity.ID
	census, censusArt := f.newSource("Census")
	def, jones, _ := artifactstest.Split(t, f.c, userID, census.ID, "Jones")
	smith := subjectstest.On(t, f.c, userID, def.ID, "person", "")
	f.cite(censusArt, smith, observations.Input{PropertyID: f.prop("name").ID, Name: namevaluetest.Western("John Smith")})

	if _, err := promotealign.Evidence(f.c, promotealign.EvidenceRequest{GraphID: jones.ID, SubjectID: smith.ID, HandleID: h}); !errors.Is(err, promote.ErrInvalid) {
		t.Errorf("Jones graph: err = %v, want promote.ErrInvalid", err)
	}
	if _, err := promotealign.Evidence(f.c, promotealign.EvidenceRequest{GraphID: def.ID, SubjectID: smith.ID, HandleID: h}); err != nil {
		t.Errorf("Default graph: err = %v", err)
	}
}

// likenessFixture is a Source person with image A (shade 40) and a handle with
// three members: image A, image B (shade 200) and a third saved under shade.
type likenessFixture struct {
	f                *fixture
	srcB             sources.Source
	incoming         subjects.Subject
	handle           []byte
	a, b, third      subjects.Subject
	thirdObservation []byte
}

func newLikenessFixture(t *testing.T, third uint8) likenessFixture {
	t.Helper()
	f := newFixture(t)
	l := likenessFixture{f: f}
	l.a = f.person(f.source, f.artifact, "John Smith")
	observationstest.SaveLikeness(t, f.c, userID, f.source.ID, l.a.ID, 40)
	l.handle = f.promote(l.a).Entity.ID

	srcM2, artM2 := f.newSource("Album")
	l.b = f.person(srcM2, artM2, "John Smith")
	observationstest.SaveLikeness(t, f.c, userID, srcM2.ID, l.b.ID, 200)
	f.join(l.b, l.handle)

	srcM3, artM3 := f.newSource("Card")
	l.third = f.person(srcM3, artM3, "John Smith")
	l.thirdObservation = observationstest.SaveLikeness(t, f.c, userID, srcM3.ID, l.third.ID, third).ID
	f.join(l.third, l.handle)

	var art artifacts.Artifact
	l.srcB, art = f.newSource("Obituary")
	l.incoming = f.person(l.srcB, art, "John Smith")
	observationstest.SaveLikeness(t, f.c, userID, l.srcB.ID, l.incoming.ID, 40)
	return l
}

func (l likenessFixture) evidence() claimevidence.Evidence {
	return l.f.evidence(promotealign.EvidenceRequest{GraphID: l.f.graph(l.srcB), SubjectID: l.incoming.ID, HandleID: l.handle})
}

// T-6: a Source person's likeness meets each member's: the same image is a
// checked line, another image is judged by eye, and a refused image is no line.
func TestEvidenceLikenesses(t *testing.T) {
	t.Run("AC-1.1 and AC-1.4 same image, other image, collapse", func(t *testing.T) {
		l := newLikenessFixture(t, 40)
		s := section(t, l.evidence().Sections, vocabkey.PropertyLikeness)
		if len(s.Lines) != 2 {
			t.Fatalf("likeness lines = %+v", s.Lines)
		}
		same, other := s.Lines[0], s.Lines[1]
		if same.Relation != valuecompare.Same || same.Cause != valuecompare.CauseExact || same.Outcome != propertycompare.OutcomeAgree || !same.Checked {
			t.Errorf("same-image line = %+v", same)
		}
		if len(same.Members) != 2 {
			t.Errorf("same-image pins = %d, want the two image-A members collapsed", len(same.Members))
		}
		if other.Relation != valuecompare.Unknown || other.Cause != valuecompare.CauseUnjudged || other.Outcome != propertycompare.OutcomeUnknown || other.Checked {
			t.Errorf("other-image line = %+v", other)
		}
		if len(other.Members) != 1 || !bytes.Equal(other.Members[0].Subject, l.b.ID) {
			t.Errorf("other-image pins = %+v", other.Members)
		}
	})

	t.Run("AC-1.5 a refused image gives no line for that member", func(t *testing.T) {
		l := newLikenessFixture(t, 90)
		db, err := l.f.c.DB()
		must(t, err)
		_, err = db.Exec(`UPDATE files SET media_type = 'image/jpeg' WHERE id = ?`, observationstest.ImageOf(t, l.f.c, l.thirdObservation))
		must(t, err)
		l.f.c.Graph().Drop()
		for _, line := range section(t, l.evidence().Sections, vocabkey.PropertyLikeness).Lines {
			for _, p := range line.Members {
				if bytes.Equal(p.Subject, l.third.ID) {
					t.Errorf("a refused image gave a line: %+v", line)
				}
			}
		}
	})
}

// T-7: the sections follow the catalog's order.
func TestEvidenceSectionOrder(t *testing.T) {
	t.Run("AC-3.1 sections follow the seeded order", func(t *testing.T) {
		f := newFixture(t)
		sex := func(src sources.Source, art artifacts.Artifact) subjects.Subject {
			p := f.person(src, art, "John Smith")
			f.cite(art, p, observations.Input{PropertyID: f.prop("sex_at_birth").ID, ValueTermID: f.term("sex_at_birth", "male").ID})
			return p
		}
		member := sex(f.source, f.artifact)
		handle := f.promote(member).Entity.ID
		srcB, artB := f.newSource("Obituary")
		incoming := sex(srcB, artB)
		e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: incoming.ID, HandleID: handle})
		var keys []vocabkey.PropertyKey
		for _, s := range e.Sections {
			keys = append(keys, s.Key)
		}
		if !slices.Equal(keys, []vocabkey.PropertyKey{"name", "sex_at_birth"}) {
			t.Errorf("sections = %v", keys)
		}
	})
}

// A Source that records a marriage reads as the spouses a tree files: one
// spouse relationship, checked, bundling both participations' Observations.
// The other way round, the filed marriage derives on the canon side later.
func TestEvidenceReadsASourceMarriageAsSpouses(t *testing.T) {
	t.Run("AC-3.1", func(t *testing.T) {
		e, ann, _ := marriageEvidence(t, true)
		if len(e.Relationships) != 1 {
			t.Fatalf("relationships = %d, want 1: %+v", len(e.Relationships), e.Relationships)
		}
		r := e.Relationships[0]
		if r.Key.String() != "relationship·spouse" || !bytes.Equal(r.Neighbor, ann.ID) || !r.Checked {
			t.Fatalf("relationship = %+v, want a checked relationship·spouse to Ann", r)
		}
		if len(r.Incoming) != 6 || len(r.Members) != 1 {
			t.Fatalf("incoming %d, members %d, want 6 and 1", len(r.Incoming), len(r.Members))
		}
	})
	t.Run("AC-3.3", func(t *testing.T) {
		e, _, _ := marriageEvidence(t, false)
		if len(e.Relationships) != 0 {
			t.Fatalf("relationships %+v, want none", e.Relationships)
		}
	})
}

// marriageEvidence is Bob's evidence on his handle with Ann placed on hers:
// the tree files a spouse relationship and the Source a marriage, or, when
// sourceMarries is false, the other way round.
func marriageEvidence(t *testing.T, sourceMarries bool) (claimevidence.Evidence, subjects.Subject, subjects.Subject) {
	t.Helper()
	f := newFixture(t)
	marry := func(src sources.Source, art artifacts.Artifact, a, b subjects.Subject) subjects.Subject {
		m := f.eventOn(src, art, "marriage", nil)
		f.participation(src, art, a, m, "spouse")
		f.participation(src, art, b, m, "spouse")
		return m
	}
	aAnn := f.person(f.source, f.artifact, "Ann Ames")
	aBob := f.person(f.source, f.artifact, "Bob Ames")
	if sourceMarries {
		f.relationship(f.source, f.artifact, aAnn, aBob, "spouse")
	}
	var treeMarriage *subjects.Subject
	if !sourceMarries {
		m := marry(f.source, f.artifact, aAnn, aBob)
		treeMarriage = &m
	}
	annH := f.promote(aAnn).Entity.ID
	bobH := f.promote(aBob).Entity.ID
	if treeMarriage != nil {
		f.promote(*treeMarriage)
	}

	srcB, artB := f.newSource("Marriage register")
	bAnn := f.person(srcB, artB, "Ann Ames")
	bBob := f.person(srcB, artB, "Bob Ames")
	if sourceMarries {
		marry(srcB, artB, bAnn, bBob)
	} else {
		f.relationship(srcB, artB, bAnn, bBob, "spouse")
	}
	e := f.evidence(promotealign.EvidenceRequest{
		GraphID: f.graph(srcB), SubjectID: bBob.ID, HandleID: bobH,
		Neighbors: []graphalign.Fixed{{SubjectID: bAnn.ID, HandleID: annH}},
	})
	return e, bAnn, bBob
}
