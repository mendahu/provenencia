package promotealign_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
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

	e := f.evidence(promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: incoming.ID, HandleID: handle})
	s := section(t, e.Sections, vocabkey.PropertyName)
	if len(s.Lines) != 1 {
		t.Fatalf("name lines = %d, want 1: %+v", len(s.Lines), s.Lines)
	}
	l := s.Lines[0]
	want := matchregistry.Default().Subject.Weights[vocabkey.Person][vocabkey.PropertyName].Agree
	if l.Relation != valuecompare.Same || !l.Checked || !l.Pinnable || len(l.Members) != 5 || l.Points != want {
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
		name       string
		year       int
		wantState  propertycompare.Outcome
		checked    bool
		dateLine   propertycompare.Outcome
		datePinned bool
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
				SourceID: srcB.ID, SubjectID: bPerson.ID, HandleID: personHandle,
				Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: birthHandle}},
			})
			if len(e.Relationships) != 1 {
				t.Fatalf("relationships = %d, want 1: %+v", len(e.Relationships), e.Relationships)
			}
			r := e.Relationships[0]
			wantKey := "participation·subject·person·birth"
			if r.Key.String() != wantKey || r.Outcome != tc.wantState || r.Checked != tc.checked || !r.Pinnable || !bytes.Equal(r.Neighbor, bBirth.ID) {
				t.Errorf("relationship = %+v, want key %s outcome %s checked %v", r, wantKey, tc.wantState, tc.checked)
			}
			if len(r.Incoming) < 2 || len(r.Members) != 1 || len(r.Members[0].Observations) < 2 {
				t.Errorf("bundles: incoming %d, members %+v", len(r.Incoming), r.Members)
			}
			if !bytes.Equal(r.Members[0].Member, aPerson.ID) || !bytes.Equal(r.Members[0].Neighbor, aBirth.ID) {
				t.Errorf("bundle ends = %x %x", r.Members[0].Member, r.Members[0].Neighbor)
			}
			d := section(t, r.Sections, vocabkey.PropertyDate)
			if len(d.Lines) != 1 || d.Lines[0].Outcome != tc.dateLine || d.Lines[0].Pinnable != tc.datePinned || d.Lines[0].Checked != tc.datePinned {
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
		{"no such Source", promotealign.EvidenceRequest{SourceID: missing, SubjectID: b.ID, HandleID: handle}},
		{"a Subject of another Source", promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: aPerson.ID, HandleID: handle}},
		{"an empty row handle", promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: b.ID}},
		{"a missing row handle", promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: b.ID, HandleID: missing}},
		{"a row handle of another kind", promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: b.ID, HandleID: birthHandle}},
		{"a missing neighbor handle", promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: b.ID, HandleID: handle,
			Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: missing}}}},
		{"a neighbor handle of another kind", promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: b.ID, HandleID: handle,
			Neighbors: []graphalign.Fixed{{SubjectID: bBirth.ID, HandleID: handle}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := promotealign.Evidence(f.c, tc.req); !errors.Is(err, promote.ErrInvalid) {
				t.Errorf("err = %v, want promote.ErrInvalid", err)
			}
		})
	}
	if _, err := promotealign.Evidence(nil, promotealign.EvidenceRequest{SourceID: srcB.ID, SubjectID: b.ID, HandleID: handle}); !errors.Is(err, promote.ErrInvalid) {
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
				SourceID: srcB.ID, SubjectID: bPerson.ID, HandleID: personHandle,
				Neighbors: []graphalign.Fixed{{SubjectID: bOther.ID, HandleID: spouseHandle}},
			})
			if len(e.Relationships) != tc.want {
				t.Errorf("relationships = %+v, want %d", e.Relationships, tc.want)
			}
		})
	}
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
				SourceID: srcB.ID, SubjectID: bPerson.ID, HandleID: personHandle, Neighbors: tc.neighbors,
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
				SourceID: srcB.ID, SubjectID: bPerson.ID, HandleID: personHandle,
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
		SourceID: srcB.ID, SubjectID: bPerson.ID, HandleID: personHandle,
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
