package promotealign_test

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database"
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
	"github.com/mendahu/provenencia/core/kinship"
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

// A Source that records a marriage reads as the spouses a tree files, and the
// other way round: one spouse relationship, checked, bundling every bridge
// behind it.
func TestEvidenceReadsASourceMarriageAsSpouses(t *testing.T) {
	t.Run("AC-1.2", func(t *testing.T) {
		sc := marriageEvidence(t, true)
		e := sc.e
		if len(e.Relationships) != 1 {
			t.Fatalf("relationships = %d, want 1: %+v", len(e.Relationships), e.Relationships)
		}
		r := e.Relationships[0]
		if r.Key.String() != "relationship·spouse" || !bytes.Equal(r.Neighbor, sc.ann.ID) || !r.Checked {
			t.Fatalf("relationship = %+v, want a checked relationship·spouse to Ann", r)
		}
		if r.Layer.Pattern != kinship.PatternMarriageSpouses || !bytes.Equal(r.Layer.Pivot, sc.recordMarriage.ID) || r.Canon.Pattern != "" {
			t.Errorf("Layer = %+v, Canon = %+v, want the record's marriage and direct", r.Layer, r.Canon)
		}
		if len(r.Incoming) != 7 || len(r.Members) != 1 {
			t.Fatalf("incoming %d, members %d, want 7 and 1", len(r.Incoming), len(r.Members))
		}
		if b := r.Members[0]; len(b.Bridges) != 1 || !bytes.Equal(b.Bridges[0], sc.treeRelationship.ID) {
			t.Errorf("bundle = %+v, want the tree relationship as its one bridge", b)
		}
	})
	t.Run("AC-1.1", func(t *testing.T) {
		sc := marriageEvidence(t, false)
		e := sc.e
		if len(e.Relationships) != 1 {
			t.Fatalf("relationships = %d, want 1: %+v", len(e.Relationships), e.Relationships)
		}
		r := e.Relationships[0]
		if r.Key.String() != "relationship·spouse" || !bytes.Equal(r.Neighbor, sc.ann.ID) || !r.Checked {
			t.Fatalf("relationship = %+v, want a checked relationship·spouse to Ann", r)
		}
		if r.Layer.Pattern != "" || r.Canon.Pattern != kinship.PatternMarriageSpouses || !bytes.Equal(r.Canon.Pivot, sc.treeMarriage.handle) {
			t.Errorf("Layer = %+v, Canon = %+v, want direct and the tree's marriage", r.Layer, r.Canon)
		}
		if len(r.Members) != 1 {
			t.Fatalf("members = %d, want 1", len(r.Members))
		}
		b := r.Members[0]
		if !bytes.Equal(b.Member, sc.treeBob.ID) || !bytes.Equal(b.Neighbor, sc.treeAnn.ID) ||
			len(b.Bridges) != 2 || !bytes.Equal(b.Bridges[0], sc.treeBobPart.ID) || !bytes.Equal(b.Bridges[1], sc.treeAnnPart.ID) {
			t.Fatalf("bundle = %+v, want Bob, Ann and their participations", b)
		}
		want := slices.Concat(sc.f.obsOn(sc.treeBobPart, ""), sc.f.obsOn(sc.treeAnnPart, ""), sc.f.obsOn(sc.treeMarriage.Subject, "event_type"))
		if !reflect.DeepEqual(b.Observations, want) {
			t.Errorf("Observations = %x, want %x", b.Observations, want)
		}
	})
}

// marriageScene is Bob's evidence on his handle with Ann placed on hers, and
// the catalog behind it.
type marriageScene struct {
	f                            *fixture
	e                            claimevidence.Evidence
	ann, bob                     subjects.Subject // the record's
	treeAnn, treeBob             subjects.Subject
	treeMarriage, recordMarriage marriageAt
	treeBobPart, treeAnnPart     subjects.Subject // the tree marriage's participations; zero when it has none
	treeRelationship             subjects.Subject // the tree's spouse relationship; zero when it has none
	annH, bobH                   []byte
	recordSource                 sources.Source
	recordArt                    artifacts.Artifact
}

// marriageAt is a marriage event Subject and, when promoted, its handle.
type marriageAt struct {
	subjects.Subject
	handle []byte
}

// obsOn is every Observation on a Subject, in the order its Evidence graph
// lists them; only is the property key that narrows them ("" for all).
func (f *fixture) obsOn(s subjects.Subject, only string) [][]byte {
	f.t.Helper()
	eg, err := f.c.Graph().EvidenceGraph(s.GraphID)
	must(f.t, err)
	var out [][]byte
	for _, o := range eg.Observations {
		if bytes.Equal(o.SubjectID, s.ID) && (only == "" || bytes.Equal(o.PropertyID, f.prop(only).ID)) {
			out = append(out, o.ID)
		}
	}
	return out
}

// marriageEvidence is Bob's evidence on his handle with Ann placed on hers:
// the tree files a spouse relationship and the Source a marriage, or, when
// sourceMarries is false, the other way round. Both marriages are dated.
func marriageEvidence(t *testing.T, sourceMarries bool) marriageScene {
	t.Helper()
	f := newFixture(t)
	sc := marriageScene{f: f}
	sc.treeAnn = f.person(f.source, f.artifact, "Ann Ames")
	sc.treeBob = f.person(f.source, f.artifact, "Bob Ames")
	if sourceMarries {
		sc.treeRelationship = f.relationship(f.source, f.artifact, sc.treeAnn, sc.treeBob, "spouse")
	} else {
		sc.treeMarriage.Subject = f.datedOn(f.source, f.artifact, "marriage", 1842, 5, 12)
		sc.treeAnnPart = f.participation(f.source, f.artifact, sc.treeAnn, sc.treeMarriage.Subject, "subject")
		sc.treeBobPart = f.participation(f.source, f.artifact, sc.treeBob, sc.treeMarriage.Subject, "subject")
	}
	sc.annH = f.promote(sc.treeAnn).Entity.ID
	sc.bobH = f.promote(sc.treeBob).Entity.ID
	if !sourceMarries {
		sc.treeMarriage.handle = f.promote(sc.treeMarriage.Subject).Entity.ID
	}

	sc.recordSource, sc.recordArt = f.newSource("Marriage register")
	srcB, artB := sc.recordSource, sc.recordArt
	sc.ann = f.person(srcB, artB, "Ann Ames")
	sc.bob = f.person(srcB, artB, "Bob Ames")
	if sourceMarries {
		sc.recordMarriage.Subject = f.datedOn(srcB, artB, "marriage", 1842, 5, 12)
		f.participation(srcB, artB, sc.ann, sc.recordMarriage.Subject, "subject")
		f.participation(srcB, artB, sc.bob, sc.recordMarriage.Subject, "subject")
	} else {
		f.relationship(srcB, artB, sc.ann, sc.bob, "spouse")
	}
	sc.e = f.evidence(promotealign.EvidenceRequest{
		GraphID: f.graph(srcB), SubjectID: sc.bob.ID, HandleID: sc.bobH,
		Neighbors: []graphalign.Fixed{{SubjectID: sc.ann.ID, HandleID: sc.annH}},
	})
	return sc
}

// oneRelationship is the one relationship line of e with key, or fails.
func oneRelationship(t *testing.T, e claimevidence.Evidence, key string) claimevidence.Relationship {
	t.Helper()
	var found []claimevidence.Relationship
	for _, r := range e.Relationships {
		if r.Key.String() == key {
			found = append(found, r)
		}
	}
	if len(found) != 1 {
		t.Fatalf("relationships %q = %+v, want one in %+v", key, found, e.Relationships)
	}
	return found[0]
}

// nameSectionPins are the Subjects the relationship's name section pins.
func nameSectionPins(r claimevidence.Relationship) [][]byte {
	var out [][]byte
	for _, s := range r.Sections {
		if s.Key != "name" {
			continue
		}
		for _, l := range s.Lines {
			for _, p := range l.Members {
				out = append(out, p.Subject)
			}
		}
	}
	return out
}

func TestEvidenceImpliedCanonLinks(t *testing.T) {
	// spouses: Bob's evidence where the tree records both a spouse relationship
	// and a marriage, the record a spouse relationship.
	t.Run("AC-1.3 both shapes", func(t *testing.T) {
		f := newFixture(t)
		ann := f.person(f.source, f.artifact, "Ann Ames")
		bob := f.person(f.source, f.artifact, "Bob Ames")
		rel := f.relationship(f.source, f.artifact, ann, bob, "spouse")
		m := f.datedOn(f.source, f.artifact, "marriage", 1842, 5, 12)
		bobPart := f.participation(f.source, f.artifact, bob, m, "subject")
		annPart := f.participation(f.source, f.artifact, ann, m, "subject")
		annH, bobH := f.promote(ann).Entity.ID, f.promote(bob).Entity.ID
		f.promote(m)
		srcB, artB := f.newSource("Register")
		bAnn, bBob := f.person(srcB, artB, "Ann Ames"), f.person(srcB, artB, "Bob Ames")
		f.relationship(srcB, artB, bAnn, bBob, "spouse")
		e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: bBob.ID, HandleID: bobH,
			Neighbors: []graphalign.Fixed{{SubjectID: bAnn.ID, HandleID: annH}}})
		r := oneRelationship(t, e, "relationship·spouse")
		if r.Canon.Pattern != "" || len(r.Members) != 2 {
			t.Fatalf("Canon = %+v, bundles = %+v, want direct and two bundles", r.Canon, r.Members)
		}
		for _, b := range r.Members {
			if !bytes.Equal(b.Member, bob.ID) || !bytes.Equal(b.Neighbor, ann.ID) {
				t.Errorf("bundle = %+v, want tree Bob and tree Ann", b)
			}
		}
		if got := r.Members[0].Bridges; len(got) != 1 || !bytes.Equal(got[0], rel.ID) {
			t.Errorf("first bundle bridges = %x, want the relationship", got)
		}
		if got := r.Members[1].Bridges; len(got) != 2 || !bytes.Equal(got[0], bobPart.ID) || !bytes.Equal(got[1], annPart.ID) {
			t.Errorf("second bundle bridges = %x, want Bob's then Ann's participation", got)
		}
	})

	birth := func(t *testing.T) (*fixture, claimevidence.Evidence, subjects.Subject, subjects.Subject, subjects.Subject, subjects.Subject, subjects.Subject, []byte) {
		f := newFixture(t)
		ann := f.person(f.source, f.artifact, "Ann Ames")
		cal := f.person(f.source, f.artifact, "Cal Ames")
		b := f.birth(f.source, f.artifact)
		annPart := f.participation(f.source, f.artifact, ann, b, "father")
		calPart := f.participation(f.source, f.artifact, cal, b, "subject")
		annH, calH := f.promote(ann).Entity.ID, f.promote(cal).Entity.ID
		bH := f.promote(b).Entity.ID
		srcB, artB := f.newSource("Register")
		bAnn, bCal := f.person(srcB, artB, "Ann Ames"), f.person(srcB, artB, "Cal Ames")
		f.relationship(srcB, artB, bAnn, bCal, "parent")
		e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: bCal.ID, HandleID: calH,
			Neighbors: []graphalign.Fixed{{SubjectID: bAnn.ID, HandleID: annH}}})
		return f, e, cal, ann, annPart, calPart, b, bH
	}
	t.Run("AC-1.4 birth parents", func(t *testing.T) {
		f, e, cal, ann, annPart, calPart, birthSubject, bH := birth(t)
		if len(e.Relationships) != 1 {
			t.Fatalf("relationships = %+v", e.Relationships)
		}
		r := e.Relationships[0]
		if r.Canon.Pattern != kinship.PatternBirthParents || !bytes.Equal(r.Canon.Pivot, bH) || r.Layer.Pattern != "" {
			t.Errorf("Layer = %+v, Canon = %+v", r.Layer, r.Canon)
		}
		if len(r.Members) != 1 {
			t.Fatalf("bundles = %+v", r.Members)
		}
		b := r.Members[0]
		if !bytes.Equal(b.Member, cal.ID) || !bytes.Equal(b.Neighbor, ann.ID) ||
			len(b.Bridges) != 2 || !bytes.Equal(b.Bridges[0], calPart.ID) || !bytes.Equal(b.Bridges[1], annPart.ID) {
			t.Fatalf("bundle = %+v", b)
		}
		if last := b.Observations[len(b.Observations)-1]; !slices.ContainsFunc(f.obsOn(birthSubject, "event_type"), func(o []byte) bool { return bytes.Equal(o, last) }) {
			t.Errorf("Observations %x do not end with the birth's event_type", b.Observations)
		}
	})

	chain := func(t *testing.T, third bool) (claimevidence.Evidence, subjects.Subject, subjects.Subject, subjects.Subject, subjects.Subject, subjects.Subject, []byte) {
		f := newFixture(t)
		ann := f.person(f.source, f.artifact, "Ann Ames")
		bea := f.person(f.source, f.artifact, "Bea Ames")
		cal := f.person(f.source, f.artifact, "Cal Ames")
		beaCal := f.relationship(f.source, f.artifact, bea, cal, "parent")
		var annBea subjects.Subject
		if !third {
			annBea = f.relationship(f.source, f.artifact, ann, bea, "parent")
		}
		annH, beaH, calH := f.promote(ann).Entity.ID, f.promote(bea).Entity.ID, f.promote(cal).Entity.ID
		if third {
			srcC, artC := f.newSource("Third")
			cAnn, cBea := f.person(srcC, artC, "Ann Ames"), f.person(srcC, artC, "Bea Ames")
			annBea = f.relationship(srcC, artC, cAnn, cBea, "parent")
			f.join(cAnn, annH)
			f.join(cBea, beaH)
		}
		srcB, artB := f.newSource("Register")
		bAnn, bCal := f.person(srcB, artB, "Ann Ames"), f.person(srcB, artB, "Cal Ames")
		f.relationship(srcB, artB, bAnn, bCal, "grandparent")
		e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: bCal.ID, HandleID: calH,
			Neighbors: []graphalign.Fixed{{SubjectID: bAnn.ID, HandleID: annH}}})
		return e, ann, cal, beaCal, annBea, subjects.Subject{ID: beaH}, annH
	}
	t.Run("AC-1.5 chain", func(t *testing.T) {
		e, ann, cal, beaCal, annBea, bea, _ := chain(t, false)
		if len(e.Relationships) != 1 {
			t.Fatalf("relationships = %+v", e.Relationships)
		}
		r := e.Relationships[0]
		if r.Canon.Pattern != kinship.PatternParentOfParent || !bytes.Equal(r.Canon.Pivot, bea.ID) {
			t.Errorf("Canon = %+v", r.Canon)
		}
		if len(r.Members) != 1 {
			t.Fatalf("bundles = %+v", r.Members)
		}
		b := r.Members[0]
		if !bytes.Equal(b.Member, cal.ID) || !bytes.Equal(b.Neighbor, ann.ID) ||
			len(b.Bridges) != 2 || !bytes.Equal(b.Bridges[0], beaCal.ID) || !bytes.Equal(b.Bridges[1], annBea.ID) {
			t.Fatalf("bundle = %+v", b)
		}
		if got := nameSectionPins(r); len(got) == 0 || !bytes.Equal(got[0], ann.ID) {
			t.Errorf("AC-1.7: name section pins %x, want the tree Ann", got)
		}
	})
	t.Run("AC-1.6 chain across Sources", func(t *testing.T) {
		e, ann, cal, beaCal, _, _, _ := chain(t, true)
		r := e.Relationships[0]
		if len(e.Relationships) != 1 || len(r.Members) != 1 {
			t.Fatalf("relationships = %+v", e.Relationships)
		}
		b := r.Members[0]
		if !bytes.Equal(b.Member, cal.ID) || b.Neighbor != nil || len(b.Bridges) != 1 || !bytes.Equal(b.Bridges[0], beaCal.ID) {
			t.Fatalf("bundle = %+v, want Cal and no neighbor", b)
		}
		for _, pin := range nameSectionPins(r) {
			if bytes.Equal(pin, ann.ID) {
				t.Errorf("AC-1.7: name section pins the tree Ann: %x", pin)
			}
		}
	})

	t.Run("AC-1.7 the neighbor is named", func(t *testing.T) {
		sc := marriageEvidence(t, false)
		r := sc.e.Relationships[0]
		pins := nameSectionPins(r)
		if len(pins) != 1 || !bytes.Equal(pins[0], sc.treeAnn.ID) {
			t.Fatalf("name section pins %x, want the tree Ann %x", pins, sc.treeAnn.ID)
		}
	})

	t.Run("AC-1.10 recorded both sides", func(t *testing.T) {
		f := newFixture(t)
		ann := f.person(f.source, f.artifact, "Ann Ames")
		bob := f.person(f.source, f.artifact, "Bob Ames")
		rel := f.relationship(f.source, f.artifact, ann, bob, "spouse")
		annH, bobH := f.promote(ann).Entity.ID, f.promote(bob).Entity.ID
		srcB, artB := f.newSource("Register")
		bAnn, bBob := f.person(srcB, artB, "Ann Ames"), f.person(srcB, artB, "Bob Ames")
		f.relationship(srcB, artB, bAnn, bBob, "spouse")
		e := f.evidence(promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: bBob.ID, HandleID: bobH,
			Neighbors: []graphalign.Fixed{{SubjectID: bAnn.ID, HandleID: annH}}})
		r := oneRelationship(t, e, "relationship·spouse")
		if r.Layer.Pattern != "" || r.Canon.Pattern != "" || len(r.Members) != 1 || len(r.Members[0].Bridges) != 1 || !bytes.Equal(r.Members[0].Bridges[0], rel.ID) {
			t.Fatalf("relationship = %+v", r)
		}
		want := "b:" + database.UUIDString(bAnn.ID) + ":relationship·spouse:" + database.UUIDString(rel.ID)
		if got := promotealign.BundlePinID(r, r.Members[0]); got != want {
			t.Errorf("BundlePinID = %q, want %q", got, want)
		}
	})
}

// Both sides record a marriage: the spouse line is counted once, through the
// marriages, and only when the request pairs them.
func TestEvidenceCountsASharedMarriageOnce(t *testing.T) {
	f := newFixture(t)
	ann := f.person(f.source, f.artifact, "Ann Ames")
	bob := f.person(f.source, f.artifact, "Bob Ames")
	m := f.datedOn(f.source, f.artifact, "marriage", 1842, 5, 12)
	f.participation(f.source, f.artifact, ann, m, "subject")
	f.participation(f.source, f.artifact, bob, m, "subject")
	annH, bobH := f.promote(ann).Entity.ID, f.promote(bob).Entity.ID
	mH := f.promote(m).Entity.ID
	srcB, artB := f.newSource("Register")
	bAnn, bBob := f.person(srcB, artB, "Ann Ames"), f.person(srcB, artB, "Bob Ames")
	m2 := f.datedOn(srcB, artB, "marriage", 1842, 5, 12)
	f.participation(srcB, artB, bAnn, m2, "subject")
	f.participation(srcB, artB, bBob, m2, "subject")
	req := promotealign.EvidenceRequest{GraphID: f.graph(srcB), SubjectID: bBob.ID, HandleID: bobH,
		Neighbors: []graphalign.Fixed{{SubjectID: bAnn.ID, HandleID: annH}}}

	t.Run("AC-1.8 paired marriages", func(t *testing.T) {
		r := req
		r.Neighbors = append(slices.Clone(req.Neighbors), graphalign.Fixed{SubjectID: m2.ID, HandleID: mH})
		e := f.evidence(r)
		for _, rel := range e.Relationships {
			if rel.Key.String() == "relationship·spouse" {
				t.Fatalf("spouse line %+v, want it counted once through the marriage", rel)
			}
		}
		var participation bool
		for _, rel := range e.Relationships {
			if rel.Key.Type == vocabkey.BridgeParticipation && rel.Key.EventType == vocabkey.EventMarriage && bytes.Equal(rel.Neighbor, m2.ID) {
				participation = true
			}
		}
		if !participation {
			t.Errorf("no participation line to the record's marriage in %+v", e.Relationships)
		}
	})
	t.Run("AC-1.9 unpaired marriages", func(t *testing.T) {
		rel := oneRelationship(t, f.evidence(req), "relationship·spouse")
		if rel.Layer.Pattern != kinship.PatternMarriageSpouses || !bytes.Equal(rel.Layer.Pivot, m2.ID) ||
			rel.Canon.Pattern != kinship.PatternMarriageSpouses || !bytes.Equal(rel.Canon.Pivot, mH) {
			t.Errorf("Layer = %+v, Canon = %+v", rel.Layer, rel.Canon)
		}
	})
}
