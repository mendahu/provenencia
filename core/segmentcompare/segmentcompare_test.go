package segmentcompare_test

import (
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/datevalue/datevaluetest"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/propertycompare"
	pt "github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func partOf(end vocabkey.PropertyKey) vocabkey.BridgeKey {
	return vocabkey.BridgeKey{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: end}
}

func partKey(end vocabkey.PropertyKey, et vocabkey.EventType) vocabkey.BridgeKey {
	return vocabkey.BridgeKey{Type: vocabkey.BridgeParticipation, Term: vocabkey.BridgeTerm(vocabkey.RoleSubject), End: end, EventType: et}
}

func term(k string) propertycompare.Value { return pt.Term(k) }
func date(y int) propertycompare.Value {
	return pt.Date(datevaluetest.YMD(y, 0, 0))
}

func event(t string, y int) subjectcompare.Values {
	v := subjectcompare.Values{vocabkey.PropertyEventType: {term(t)}}
	if y != 0 {
		v[vocabkey.PropertyDate] = []propertycompare.Value{date(y)}
	}
	return v
}

func seg(k vocabkey.BridgeKey, kind vocabkey.SubjectType, v subjectcompare.Values) segmentcompare.Segment {
	return segmentcompare.Segment{Key: k, NeighborKind: kind, Neighbor: v}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestKey(t *testing.T) {
	terms := segmentcomparetest.Terms()
	rel := func(term vocabkey.BridgeTerm, end, other vocabkey.PropertyKey) segmentcompare.Link {
		return segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: term, End: end, Other: other, SameType: true}
	}
	tests := []struct {
		name string
		link segmentcompare.Link
		want string
	}{
		{"AC-1.1 spouse from person", rel("spouse", vocabkey.PropertyPerson, vocabkey.PropertyRelatedTo), "relationship·spouse"},
		{"AC-1.1 spouse from related_to", rel("spouse", vocabkey.PropertyRelatedTo, vocabkey.PropertyPerson), "relationship·spouse"},
		{"AC-1.1 sibling", rel("sibling", vocabkey.PropertyPerson, vocabkey.PropertyRelatedTo), "relationship·sibling"},
		{"AC-1.2 parent, Mary's end", rel("parent", vocabkey.PropertyPerson, vocabkey.PropertyRelatedTo), "relationship·child·related_to"},
		{"AC-1.2 parent, John's end", rel("parent", vocabkey.PropertyRelatedTo, vocabkey.PropertyPerson), "relationship·child·person"},
		{"AC-1.2 child, Mary's end", rel("child", vocabkey.PropertyRelatedTo, vocabkey.PropertyPerson), "relationship·child·related_to"},
		{"AC-1.2 child, John's end", rel("child", vocabkey.PropertyPerson, vocabkey.PropertyRelatedTo), "relationship·child·person"},
		{"AC-1.3 birth from the person", segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: "subject", End: vocabkey.PropertyPerson, Other: vocabkey.PropertyEvent, EventType: vocabkey.EventBirth}, "participation·subject·person·birth"},
		{"AC-1.3 birth from the event", segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: "subject", End: vocabkey.PropertyEvent, Other: vocabkey.PropertyPerson, EventType: vocabkey.EventBirth}, "participation·subject·event·birth"},
		{"AC-1.4 part_of from the part", segmentcompare.Link{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: vocabkey.PropertyFrom, Other: vocabkey.PropertyTo, SameType: true}, "place_relationship·part_of·from"},
		{"AC-1.4 part_of from the whole", segmentcompare.Link{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: vocabkey.PropertyTo, Other: vocabkey.PropertyFrom, SameType: true}, "place_relationship·part_of·to"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := segmentcompare.Key(tc.link, terms).String(); got != tc.want {
				t.Errorf("Key = %q, want %q", got, tc.want)
			}
		})
	}
	spouse := segmentcompare.Key(rel("spouse", vocabkey.PropertyPerson, vocabkey.PropertyRelatedTo), terms)
	sibling := segmentcompare.Key(rel("sibling", vocabkey.PropertyPerson, vocabkey.PropertyRelatedTo), terms)
	if segmentcompare.Corresponds(spouse, sibling) {
		t.Error("a spouse corresponds to a sibling")
	}
	if !segmentcompare.Corresponds(spouse, spouse) {
		t.Error("a link does not correspond to itself")
	}
}

func TestCompare(t *testing.T) {
	metas := subjectcomparetest.Metas()
	reg := matchregistry.Default()
	birthP := partKey(vocabkey.PropertyPerson, vocabkey.EventBirth)
	birthE := partKey(vocabkey.PropertyEvent, vocabkey.EventBirth)
	resid := partKey(vocabkey.PropertyPerson, "residence")
	from, to := partOf(vocabkey.PropertyFrom), partOf(vocabkey.PropertyTo)
	york := subjectcompare.Values{vocabkey.PropertyToponym: {pt.Text("York")}}
	spouse := vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "spouse"}
	sibling := vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "sibling"}
	ann := subjectcompare.Values{vocabkey.PropertyName: {pt.Name("Ann Ames")}}
	bob := subjectcompare.Values{vocabkey.PropertyName: {pt.Name("John Smith")}}

	tests := []struct {
		name     string
		a, b     segmentcompare.Segment
		placed   segmentcompare.Placement
		own      float64
		outcome  propertycompare.Outcome
		cause    segmentcompare.Cause
		evidence float64
		neighbor float64
	}{
		{"AC-1.5", seg(spouse, vocabkey.Person, nil), seg(sibling, vocabkey.Person, nil), segmentcompare.Together, 0, propertycompare.OutcomeUnknown, segmentcompare.CauseNotCorresponding, 0, 0},
		{"AC-2.1", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1817)), segmentcompare.Together, 2.5, propertycompare.OutcomeAgree, "", 2.75, 2.5},
		{"AC-2.2", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1817)), segmentcompare.Apart, 2.5, propertycompare.OutcomeConflict, segmentcompare.CausePlacedApart, -3, 0},
		{"AC-2.3", seg(birthE, vocabkey.Person, nil), seg(birthE, vocabkey.Person, nil), segmentcompare.Apart, 0, propertycompare.OutcomeConflict, segmentcompare.CausePlacedApart, -3, 0},
		{"AC-2.4", seg(resid, vocabkey.Event, event("residence", 0)), seg(resid, vocabkey.Event, event("residence", 0)), segmentcompare.Apart, 2.5, propertycompare.OutcomeUnknown, segmentcompare.CausePlacedApart, 0, 0},
		{"AC-2.5", seg(resid, vocabkey.Event, event("residence", 0)), seg(resid, vocabkey.Event, event("residence", 0)), segmentcompare.Together, 2.5, propertycompare.OutcomeAgree, "", 0.375, 0.5},
		{"AC-2.6", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1830)), segmentcompare.Together, 2.5, propertycompare.OutcomePartial, "", 0.25, -2.5},
		{"AC-2.7", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Together, 1.9, propertycompare.OutcomeUnknown, segmentcompare.CauseGated, 0, 3},
		{"AC-2.8", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Together, 3, propertycompare.OutcomeAgree, "", 1, 3},
		{"AC-2.9", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Apart, 3, propertycompare.OutcomeConflict, segmentcompare.CausePlacedApart, -3, 0},
		{"unplaced alike", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1817)), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeAgree, "", 2.75, 2.5},
		{"unplaced differs", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1830)), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeConflict, segmentcompare.CauseUnplacedDiffers, -3, -2.5},
		{"unplaced unsure", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 0)), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeUnknown, segmentcompare.CauseUnplacedUnsure, 0, 0.5},
		{"unplaced spouse differs", seg(spouse, vocabkey.Person, ann), seg(spouse, vocabkey.Person, bob), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeUnknown, segmentcompare.CauseUnplacedDiffers, 0, -3},
		{"unplaced gated", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Unplaced, 1, propertycompare.OutcomeUnknown, segmentcompare.CauseGated, 0, 3},
		{"AC-2.10", seg(to, vocabkey.Place, york), seg(to, vocabkey.Place, york), segmentcompare.Apart, 3, propertycompare.OutcomeUnknown, segmentcompare.CausePlacedApart, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := segmentcompare.Compare(tc.a, tc.b, tc.placed, tc.own, metas, reg.Segment, reg.Subject, reg.Value)
			if got.Outcome != tc.outcome || got.Cause != tc.cause || !near(got.Evidence, tc.evidence) || !near(got.Neighbor.Evidence, tc.neighbor) {
				t.Errorf("got %s/%q evidence %v neighbor %v; want %s/%q evidence %v neighbor %v",
					got.Outcome, got.Cause, got.Evidence, got.Neighbor.Evidence, tc.outcome, tc.cause, tc.evidence, tc.neighbor)
			}
		})
	}
}

func TestInvariants(t *testing.T) {
	metas := subjectcomparetest.Metas()
	reg := matchregistry.Default()
	keys := []vocabkey.BridgeKey{
		partKey(vocabkey.PropertyPerson, vocabkey.EventBirth), partKey(vocabkey.PropertyEvent, vocabkey.EventBirth),
		partKey(vocabkey.PropertyPerson, vocabkey.EventDeath), partKey(vocabkey.PropertyEvent, vocabkey.EventDeath),
		partKey(vocabkey.PropertyPerson, "residence"),
		{Type: vocabkey.BridgeRelationship, Term: "spouse"}, {Type: vocabkey.BridgeRelationship, Term: "sibling"},
		partOf(vocabkey.PropertyFrom), partOf(vocabkey.PropertyTo),
	}
	neighbors := []subjectcompare.Values{
		event("birth", 1817), event("birth", 1830), event("residence", 0), event("death", 1890),
		{vocabkey.PropertyToponym: {pt.Text("York")}}, {vocabkey.PropertyToponym: {pt.Text("London")}},
		{vocabkey.PropertyName: {pt.Name("Ann Ames")}}, {vocabkey.PropertyName: {pt.Name("Bob Ames")}},
	}
	owns := []float64{0, 1.9, 3}
	r := rand.New(rand.NewSource(1))
	isOne := func(k vocabkey.BridgeKey) bool {
		k.End = ""
		return slices.Contains(reg.Segment.One, k)
	}
	for i := 0; i < 500; i++ {
		ka, kb := keys[r.Intn(len(keys))], keys[r.Intn(len(keys))]
		if r.Intn(2) == 0 {
			kb = ka
		}
		a := seg(ka, vocabkey.Event, neighbors[r.Intn(len(neighbors))])
		b := seg(kb, vocabkey.Event, neighbors[r.Intn(len(neighbors))])
		placed := segmentcompare.Together
		if r.Intn(2) == 0 {
			placed = segmentcompare.Apart
		}
		own := owns[r.Intn(len(owns))]
		got := segmentcompare.Compare(a, b, placed, own, metas, reg.Segment, reg.Subject, reg.Value)
		g := reg.Segment.Gates[ka]
		switch {
		case ka != kb:
			if got.Evidence != 0 || got.Cause != segmentcompare.CauseNotCorresponding {
				t.Fatalf("case %d: non-corresponding %v/%v gave %+v", i, ka, kb, got)
			}
		case placed == segmentcompare.Together && own >= g.MinOwnEvidence:
			support, credit := reg.Segment.OtherSupport, reg.Segment.OtherCredit
			if isOne(ka) {
				support, credit = reg.Segment.OneSupport, reg.Segment.OneCredit
			}
			if want := support + credit*got.Neighbor.Evidence; !near(got.Evidence, want) {
				t.Fatalf("case %d: %v together gave %v, want %v", i, ka, got.Evidence, want)
			}
		case placed == segmentcompare.Apart && !isOne(ka) && g == (segmentcompare.Gate{}):
			if got.Evidence != 0 {
				t.Fatalf("case %d: %v apart gave %v, want 0", i, ka, got.Evidence)
			}
		}
		if ka == kb && placed == segmentcompare.Apart && isOne(ka) {
			other := ka
			if ka.End == vocabkey.PropertyPerson {
				other.End = vocabkey.PropertyEvent
			} else {
				other.End = vocabkey.PropertyPerson
			}
			back := segmentcompare.Compare(seg(other, vocabkey.Event, a.Neighbor), seg(other, vocabkey.Event, b.Neighbor), placed, own, metas, reg.Segment, reg.Subject, reg.Value)
			if !near(back.Evidence, got.Evidence) {
				t.Fatalf("case %d: a One fact apart reads %v from one end and %v from the other", i, got.Evidence, back.Evidence)
			}
		}
	}
}
