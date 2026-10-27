package claimevidence_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/claimevidence"
	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
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
	return s.buildWith(claimevidence.Inputs{Order: seedOrder()})
}

func (s scene) buildWith(in claimevidence.Inputs) claimevidence.Evidence {
	reg := matchregistry.Default()
	row := vals{vocabkey.PropertyName: {ref(name("John Smith"), 0x01)}}
	return claimevidence.Build(s.kind, row, row, s.members, s.neighbors, subjectcomparetest.Metas(),
		reg.Evidence, reg.Segment, reg.Subject, reg.Value, in)
}

func TestBuild(t *testing.T) {
	t.Run("AC-2.1 a birth in the same year agrees and bundles its Observations", func(t *testing.T) {
		e := newScene(1817, birthKey, birthKey).build()
		if len(e.Relationships) != 1 {
			t.Fatalf("relationships = %+v", e.Relationships)
		}
		r := e.Relationships[0]
		if r.Key != birthKey || r.Outcome != propertycompare.OutcomeAgree || !r.Checked || r.Points <= 0 || !bytes.Equal(r.Neighbor, []byte{0x30}) {
			t.Errorf("relationship = %+v", r)
		}
		// The layer link's Observations, then the defining event_type's.
		if want := [][]byte{{0x60}, {0x61}, {0x40}}; !equalIDs(r.Incoming, want) {
			t.Errorf("Incoming = %x, want %x", r.Incoming, want)
		}
		if len(r.Members) != 1 {
			t.Fatalf("bundles = %+v", r.Members)
		}
		if e.Kind != vocabkey.Person || r.Kind != vocabkey.Event || !equalIDs(r.Members[0].Bridges, [][]byte{{0x70}}) {
			t.Errorf("kind = %q, relationship kind = %q, bridges = %x", e.Kind, r.Kind, r.Members[0].Bridges)
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
		if r.Outcome != propertycompare.OutcomePartial || r.Checked {
			t.Errorf("relationship = %+v", r)
		}
		if l := only(t, r.Sections, vocabkey.PropertyDate); l.Outcome != propertycompare.OutcomeConflict {
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

// impliedScene is a person row (member 1) with a person neighbor (member 2)
// whose spouse link the entity records through a marriage: bob's participation
// 0x70, ann's 0x72, both reaching the marriage member 0xE0 whose event_type
// Observation is 0x90.
func impliedScene() scene {
	bob := member(1, vals{vocabkey.PropertyName: {ref(name("Bob Ames"), 0x10)}})
	ann := member(2, vals{vocabkey.PropertyName: {ref(name("Ann Ames"), 0x20)}})
	marriage := []byte{0xE0}
	shape := kinship.Shape{Pattern: kinship.PatternMarriageSpouses, Pivot: []byte{0xE1}}
	n := claimevidence.Neighbor{
		Subject: []byte{0x30}, Kind: vocabkey.Person,
		Values:  vals{vocabkey.PropertyName: {ref(name("Ann Ames"), 0x40)}},
		Handle:  vals{vocabkey.PropertyName: {ref(name("Ann Ames"), 0x50)}},
		Members: []claimevidence.Member{ann},
		Layer:   []claimevidence.LayerLink{{Key: spouseKey, Observations: [][]byte{{0x60}}}},
		Canon: []claimevidence.CanonLink{{
			Key:       spouseKey,
			Recording: kinship.Recording{Shapes: []kinship.Shape{shape}, OnlyImplied: true},
			Pivots:    []claimevidence.Pivot{{Subject: marriage, Kind: vocabkey.Event, Values: vals{vocabkey.PropertyEventType: {ref(term("marriage"), 0x90)}}}},
			Bridges: []claimevidence.Bridge{
				{Subject: []byte{0x70}, Source: []byte{0xB1}, Member: bob.Subject, Pivot: marriage, Observations: [][]byte{{0x71}}},
				{Subject: []byte{0x72}, Source: []byte{0xB1}, Neighbor: ann.Subject, Pivot: marriage, Observations: [][]byte{{0x73}}},
			},
		}},
	}
	return scene{kind: vocabkey.Person, members: []claimevidence.Member{bob}, neighbors: []claimevidence.Neighbor{n}}
}

func TestBuildImpliedLinks(t *testing.T) {
	pivot := kinship.Shape{Pattern: kinship.PatternMarriageSpouses, Pivot: []byte{0xE1}}

	t.Run("a link only implied on a side shows its shape, a direct one does not", func(t *testing.T) {
		s := impliedScene()
		s.neighbors[0].Layer[0].Recording = kinship.Recording{Shapes: []kinship.Shape{{Pattern: kinship.PatternMarriageSpouses, Pivot: []byte{0xF1}}}, OnlyImplied: true}
		r := s.build().Relationships[0]
		if r.Layer.Pattern != kinship.PatternMarriageSpouses || !bytes.Equal(r.Layer.Pivot, []byte{0xF1}) || !reflect.DeepEqual(r.Canon, pivot) {
			t.Errorf("Layer = %+v, Canon = %+v", r.Layer, r.Canon)
		}
		s = impliedScene()
		s.neighbors[0].Canon[0].Recording = kinship.Recording{}
		if r := s.build().Relationships[0]; !reflect.DeepEqual(r.Canon, kinship.Shape{}) {
			t.Errorf("Canon = %+v, want direct", r.Canon)
		}
	})

	t.Run("the bridges that reach one pivot are one bundle, the member's first", func(t *testing.T) {
		r := impliedScene().build().Relationships[0]
		if len(r.Members) != 1 {
			t.Fatalf("bundles = %+v", r.Members)
		}
		b := r.Members[0]
		if !bytes.Equal(b.Member, []byte{1}) || !bytes.Equal(b.Neighbor, []byte{2}) || !equalIDs(b.Bridges, [][]byte{{0x70}, {0x72}}) {
			t.Errorf("bundle = %+v", b)
		}
		if want := [][]byte{{0x71}, {0x73}, {0x90}}; !equalIDs(b.Observations, want) {
			t.Errorf("Observations = %x, want %x", b.Observations, want)
		}
		if got := len(r.Sections); got == 0 {
			t.Errorf("the neighbor member named by the bundle should be compared: %+v", r.Sections)
		}
	})

	t.Run("a bundle needs a bridge of a row member", func(t *testing.T) {
		s := impliedScene()
		s.neighbors[0].Canon[0].Bridges = s.neighbors[0].Canon[0].Bridges[1:]
		if r := s.build().Relationships[0]; len(r.Members) != 0 || len(r.Sections) != 0 {
			t.Errorf("bundles = %+v, sections = %+v", r.Members, r.Sections)
		}
	})

	t.Run("a member's bridge alone is a bundle with no neighbor, after a recorded one", func(t *testing.T) {
		s := impliedScene()
		c := &s.neighbors[0].Canon[0]
		c.Bridges = c.Bridges[:1]
		s.neighbors[0].Canon = append(s.neighbors[0].Canon, claimevidence.CanonLink{Key: spouseKey, Bridges: []claimevidence.Bridge{
			{Subject: []byte{0x80}, Source: []byte{0xB2}, Member: []byte{1}, Neighbor: []byte{2}, Observations: [][]byte{{0x81}}},
		}})
		r := s.build().Relationships[0]
		if len(r.Members) != 2 || !equalIDs(r.Members[0].Bridges, [][]byte{{0x80}}) || !equalIDs(r.Members[1].Bridges, [][]byte{{0x70}}) || r.Members[1].Neighbor != nil {
			t.Fatalf("bundles = %+v", r.Members)
		}
	})

	t.Run("both sides implied through paired pivots are counted once", func(t *testing.T) {
		s := impliedScene()
		s.neighbors[0].Layer[0].Recording = kinship.Recording{Shapes: []kinship.Shape{{Pattern: kinship.PatternMarriageSpouses, Pivot: []byte{0xF1}}}, OnlyImplied: true}
		in := claimevidence.Inputs{Order: seedOrder(), Placed: map[string][]byte{string([]byte{0xF1}): {0xE1}}}
		if e := s.buildWith(in); len(e.Relationships) != 0 {
			t.Errorf("relationships = %+v", e.Relationships)
		}
		in.Placed = nil
		if e := s.buildWith(in); len(e.Relationships) != 1 {
			t.Errorf("unpaired relationships = %+v", e.Relationships)
		}
	})

	t.Run("both sides recorded are never counted once", func(t *testing.T) {
		s := newScene(1817, birthKey, birthKey)
		in := claimevidence.Inputs{Order: seedOrder(), Placed: map[string][]byte{"x": {1}}}
		if e := s.buildWith(in); len(e.Relationships) != 1 {
			t.Errorf("relationships = %+v", e.Relationships)
		}
	})

	t.Run("a layer pivot's defining Observation ends the incoming list", func(t *testing.T) {
		s := impliedScene()
		l := &s.neighbors[0].Layer[0]
		l.Pivots = []claimevidence.Pivot{{Subject: []byte{0xF1}, Kind: vocabkey.Event, Values: vals{vocabkey.PropertyEventType: {ref(term("marriage"), 0x91)}}}}
		r := s.build().Relationships[0]
		if want := [][]byte{{0x60}, {0x91}}; !equalIDs(r.Incoming, want) {
			t.Errorf("Incoming = %x, want %x", r.Incoming, want)
		}
	})
}
