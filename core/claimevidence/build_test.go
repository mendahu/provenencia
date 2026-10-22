package claimevidence_test

import (
	"bytes"
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/subjectprofile"
	"github.com/mendahu/provenencia/core/vocabkey"
)

var (
	birthKey = vocabkey.BridgeKey{
		Type: vocabkey.BridgeParticipation, Term: vocabkey.BridgeTerm(vocabkey.RoleSubject),
		End: vocabkey.PropertyPerson, EventType: vocabkey.EventBirth,
	}
	spouseKey  = vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "spouse"}
	siblingKey = vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "sibling"}
)

// birthValues is an event cited as a birth in year (0: undated), refs from base.
func birthValues(base byte, year int) vals {
	v := vals{vocabkey.PropertyEventType: {ref(term("birth"), base)}}
	if year != 0 {
		v[vocabkey.PropertyDate] = []pv{ref(date(year), base+1)}
	}
	return v
}

// scene is a person with one member, and a birth neighbor with one member
// joined to it by one bridge.
type scene struct {
	kind      vocabkey.SubjectType
	members   []claimevidence.Member
	neighbors []claimevidence.Neighbor
}

func newScene(year int, layerKey, canonKey vocabkey.BridgeKey) scene {
	person := member(1, vals{vocabkey.PropertyName: {ref(name("John Smith"), 0x10)}})
	theirs := member(2, birthValues(0x50, 1817))
	if year == 0 {
		theirs = member(2, birthValues(0x50, 0))
	}
	n := claimevidence.Neighbor{
		Subject: []byte{0x30}, Kind: vocabkey.Event,
		Values:  birthValues(0x40, year),
		Handle:  birthValues(0x50, 1817),
		Members: []claimevidence.Member{theirs},
		Layer:   []claimevidence.LayerLink{{Key: layerKey, Observations: [][]byte{{0x60}, {0x61}}}},
		Canon: []claimevidence.CanonLink{{Key: canonKey, Bridges: []claimevidence.Bridge{{
			Subject: []byte{0x70}, Source: []byte{0xB1}, Member: person.Subject, Neighbor: theirs.Subject,
			Observations: [][]byte{{0x71}, {0x72}},
		}}}},
	}
	return scene{kind: vocabkey.Person, members: []claimevidence.Member{person}, neighbors: []claimevidence.Neighbor{n}}
}

func (s scene) build() claimevidence.Evidence {
	reg := matchregistry.Default()
	row := vals{vocabkey.PropertyName: {ref(name("John Smith"), 0x01)}}
	return claimevidence.Build(s.kind, row, row, s.members, s.neighbors, subjectcomparetest.Metas(),
		reg.Evidence, reg.Segment, reg.Subject, reg.Value, claimevidence.Inputs{Order: seedOrder()})
}

func TestBuild(t *testing.T) {
	t.Run("AC-2.1 a birth in the same year agrees and bundles its Observations", func(t *testing.T) {
		e := newScene(1817, birthKey, birthKey).build()
		if len(e.Relationships) != 1 {
			t.Fatalf("relationships = %+v", e.Relationships)
		}
		r := e.Relationships[0]
		if r.Key != birthKey || r.Outcome != propertycompare.OutcomeAgree || !r.Checked || !r.Pinnable || r.Points <= 0 || !bytes.Equal(r.Neighbor, []byte{0x30}) {
			t.Errorf("relationship = %+v", r)
		}
		// The layer link's Observations, then the defining event_type's.
		if want := [][]byte{{0x60}, {0x61}, {0x40}}; !equalIDs(r.Incoming, want) {
			t.Errorf("Incoming = %x, want %x", r.Incoming, want)
		}
		if len(r.Members) != 1 {
			t.Fatalf("bundles = %+v", r.Members)
		}
		if e.Kind != vocabkey.Person || r.Kind != vocabkey.Event || !bytes.Equal(r.Members[0].Bridge, []byte{0x70}) {
			t.Errorf("kind = %q, relationship kind = %q, bridge = %x", e.Kind, r.Kind, r.Members[0].Bridge)
		}
		if want := [][]byte{{0x71}, {0x72}, {0x50}}; !equalIDs(r.Members[0].Observations, want) {
			t.Errorf("bundle Observations = %x, want %x", r.Members[0].Observations, want)
		}
		if b := r.Members[0]; !bytes.Equal(b.Member, []byte{1}) || !bytes.Equal(b.Neighbor, []byte{2}) || !bytes.Equal(b.Source, []byte{0xB1}) {
			t.Errorf("bundle = %+v", b)
		}
		l := only(t, r.Sections, vocabkey.PropertyDate)
		if !l.Checked || !bytes.Equal(l.Members[0].Observation, []byte{0x51}) || !bytes.Equal(l.Incoming.Ref, []byte{0x41}) {
			t.Errorf("date line = %+v", l)
		}
	})

	t.Run("AC-2.2 a birth in another year is partial, and its date a conflict", func(t *testing.T) {
		r := newScene(1830, birthKey, birthKey).build().Relationships[0]
		if r.Outcome != propertycompare.OutcomePartial || r.Checked || !r.Pinnable {
			t.Errorf("relationship = %+v", r)
		}
		if l := only(t, r.Sections, vocabkey.PropertyDate); l.Outcome != propertycompare.OutcomeConflict || l.Pinnable {
			t.Errorf("date line = %+v", l)
		}
	})

	t.Run("AC-2.3 a sibling does not correspond to a spouse", func(t *testing.T) {
		if e := newScene(1817, siblingKey, spouseKey).build(); len(e.Relationships) != 0 {
			t.Errorf("relationships = %+v", e.Relationships)
		}
	})

	t.Run("AC-2.4 no neighbor, no relationship", func(t *testing.T) {
		s := newScene(1817, birthKey, birthKey)
		s.neighbors = nil
		if e := s.build(); len(e.Relationships) != 0 {
			t.Errorf("relationships = %+v", e.Relationships)
		}
	})

	t.Run("AC-2.5 two bridges are two bundles", func(t *testing.T) {
		s := newScene(1817, birthKey, birthKey)
		second := member(3, vals{vocabkey.PropertyName: {ref(name("John Smith"), 0x11)}})
		theirs2 := member(4, birthValues(0x52, 1817))
		s.members = append(s.members, second)
		n := &s.neighbors[0]
		n.Members = append(n.Members, theirs2)
		n.Canon[0].Bridges = append(n.Canon[0].Bridges, claimevidence.Bridge{
			Subject: []byte{0x73}, Source: []byte{0xB2}, Member: second.Subject, Neighbor: theirs2.Subject,
			Observations: [][]byte{{0x74}},
		})
		r := s.build().Relationships[0]
		if len(r.Members) != 2 || !bytes.Equal(r.Members[0].Member, []byte{1}) || !bytes.Equal(r.Members[1].Member, []byte{3}) {
			t.Fatalf("bundles = %+v", r.Members)
		}
		// Both neighbor members' dates are compared with the row's birth.
		if l := only(t, r.Sections, vocabkey.PropertyDate); len(l.Members) != 2 {
			t.Errorf("date line members = %+v", l.Members)
		}
	})

	t.Run("a bridge from a Subject that is not a member is skipped", func(t *testing.T) {
		s := newScene(1817, birthKey, birthKey)
		s.neighbors[0].Canon[0].Bridges[0].Member = []byte{9}
		r := s.build().Relationships[0]
		if len(r.Members) != 0 || len(r.Sections) != 0 {
			t.Errorf("relationship = %+v, want no bundles and no sections", r)
		}
	})

	t.Run("AC-2.6 no dates, no section", func(t *testing.T) {
		r := newScene(0, birthKey, birthKey).build().Relationships[0]
		if r.Outcome != propertycompare.OutcomeAgree || !r.Checked || len(r.Members) != 1 || len(r.Sections) != 0 {
			t.Errorf("relationship = %+v", r)
		}
	})
}

func equalIDs(a, b [][]byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

// side builds one side's values and life from its life events, as the loader does.
func side(events ...subjectprofile.LifeEvent) (vals, subjectcompare.LifeSide) {
	return subjectprofile.ForSubject(subjectprofile.Subject{Life: events}, matchregistry.Default().Subject.Lifespans[vocabkey.Person])
}

func life(eventType vocabkey.EventType, d *datevalue.Value) subjectprofile.LifeEvent {
	return subjectprofile.LifeEvent{EventType: eventType, Role: vocabkey.RoleSubject, Dates: []datevalue.Value{*d}}
}

func buildLife(kind vocabkey.SubjectType, row, rec vals, rowLife, recLife subjectcompare.LifeSide) claimevidence.Evidence {
	reg := matchregistry.Default()
	return claimevidence.Build(kind, row, rec, nil, nil, subjectcomparetest.Metas(),
		reg.Evidence, reg.Segment, reg.Subject, reg.Value,
		claimevidence.Inputs{Order: seedOrder(), Row: rowLife, Record: recLife})
}

func TestBuildLifespan(t *testing.T) {
	ymd := datevaluetest.YMD
	person := vocabkey.SubjectType(vocabkey.Person)
	rowVals, rowLife := side(life(vocabkey.EventBirth, ymd(1817, 5, 14)), life(vocabkey.EventDeath, ymd(1890, 6, 2)))

	t.Run("AC-2.1 two dated sides agree and carry both lives", func(t *testing.T) {
		recVals, recLife := side(life(vocabkey.EventBirth, ymd(1817, 5, 14)), life("census", ymd(1871, 4, 2)))
		l := buildLife(person, rowVals, recVals, rowLife, recLife).Lifespan
		if l == nil || l.Outcome != propertycompare.OutcomeAgree {
			t.Fatalf("lifespan = %+v", l)
		}
		if l.Row.Missing != "" || l.Row.First == nil || l.Row.Last == nil || l.Record.First == nil || l.Record.Last == nil {
			t.Errorf("sides = %+v", l)
		}
		if l.Record.Last.StartYear == nil || *l.Record.Last.StartYear != 1871 {
			t.Errorf("record last = %+v", l.Record.Last)
		}
	})

	t.Run("AC-2.2 a side that contradicts itself has no lifespan and the check is unknown", func(t *testing.T) {
		badVals, badLife := side(life(vocabkey.EventBirth, ymd(1817, 5, 14)), life(vocabkey.EventDeath, ymd(1809, 6, 2)))
		recVals, recLife := side(life(vocabkey.EventBirth, ymd(1817, 5, 14)))
		l := buildLife(person, badVals, recVals, badLife, recLife).Lifespan
		if l == nil || l.Row.Missing != subjectcompare.LifespanContradicts || l.Row.First != nil {
			t.Fatalf("lifespan = %+v", l)
		}
		if l.Outcome != propertycompare.OutcomeUnknown || l.Relation != "" || l.Cause != "" {
			t.Errorf("verdict = %+v", l)
		}
	})

	t.Run("AC-2.3 an undated record is missing, undated", func(t *testing.T) {
		recVals, recLife := side()
		l := buildLife(person, rowVals, recVals, rowLife, recLife).Lifespan
		if l == nil || l.Record.Missing != subjectcompare.LifespanUndated || l.Outcome != propertycompare.OutcomeUnknown {
			t.Fatalf("lifespan = %+v", l)
		}
	})

	t.Run("AC-2.4 neither side dated gives no lifespan", func(t *testing.T) {
		a, al := side()
		b, bl := side()
		if l := buildLife(person, a, b, al, bl).Lifespan; l != nil {
			t.Errorf("lifespan = %+v", l)
		}
	})

	t.Run("AC-2.5 a record whose life ended before the row was born conflicts", func(t *testing.T) {
		recVals, recLife := side(life("baptism", ymd(1801, 3, 3)), life(vocabkey.EventBurial, ymd(1810, 1, 9)))
		l := buildLife(person, rowVals, recVals, rowLife, recLife).Lifespan
		if l == nil || l.Outcome != propertycompare.OutcomeConflict {
			t.Fatalf("lifespan = %+v", l)
		}
	})

	t.Run("AC-2.6 an event has no lifespan", func(t *testing.T) {
		if l := buildLife(vocabkey.Event, rowVals, rowVals, rowLife, rowLife).Lifespan; l != nil {
			t.Errorf("lifespan = %+v", l)
		}
	})
}
