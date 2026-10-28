package segmentcompare_test

import (
	"math"
	"math/rand"
	"reflect"
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

func TestReadFrom(t *testing.T) {
	terms := segmentcomparetest.Terms()
	// A researcher's directed term with no inverse.
	terms[vocabkey.BridgeRelationship]["godparent"] = segmentcompare.Term{Directed: true}
	rel := func(term vocabkey.BridgeTerm, end, other vocabkey.PropertyKey) segmentcompare.Link {
		return segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: term, End: end, Other: other, SameType: true}
	}
	place := func(end, other vocabkey.PropertyKey) segmentcompare.Link {
		return segmentcompare.Link{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: end, Other: other, SameType: true}
	}
	partic := func(end, other vocabkey.PropertyKey) segmentcompare.Link {
		return segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: "subject", End: end, Other: other, EventType: vocabkey.EventBirth}
	}
	const (
		person    vocabkey.PropertyKey = vocabkey.PropertyPerson
		relatedTo vocabkey.PropertyKey = vocabkey.PropertyRelatedTo
		from      vocabkey.PropertyKey = vocabkey.PropertyFrom
		to        vocabkey.PropertyKey = vocabkey.PropertyTo
		event     vocabkey.PropertyKey = vocabkey.PropertyEvent
	)
	tests := []struct {
		name          string
		link          segmentcompare.Link
		first, second vocabkey.PropertyKey
		end           vocabkey.PropertyKey
		term          vocabkey.BridgeTerm
		names         bool
	}{
		{"row is the child", rel("parent", relatedTo, person), person, relatedTo, relatedTo, "parent", true},
		{"row is the parent", rel("parent", person, relatedTo), person, relatedTo, person, "child", true},
		{"a symmetric spouse link has no end", rel("spouse", person, relatedTo), person, relatedTo, "", "spouse", true},
		{"row is the part", place(from, to), from, to, to, vocabkey.TermPartOf, false},
		{"row is the whole", place(to, from), from, to, from, vocabkey.TermPartOf, true},
		{"row is the person in a birth", partic(person, event), person, event, event, "subject", false},
		{"row is the birth", partic(event, person), person, event, person, "subject", true},
		{"a researcher's term with no inverse", rel("godparent", person, relatedTo), person, relatedTo, relatedTo, "godparent", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			k := segmentcompare.Key(tc.link, terms)
			end, term, names := segmentcompare.ReadFrom(k, tc.first, tc.second, terms)
			if end != tc.end || term != tc.term || names != tc.names {
				t.Errorf("ReadFrom(%s) = (%q, %q, %v), want (%q, %q, %v)", k, end, term, names, tc.end, tc.term, tc.names)
			}
		})
	}
}

// compareMetas and compareReg are the metas and registry the Compare cases run under.
var (
	compareMetas = subjectcomparetest.Metas()
	compareReg   = matchregistry.Default()
)

// compareCase is one row of the Compare table, run by TestCompare and
// TestCompareNeighborMatchesCompare.
type compareCase struct {
	name     string
	a, b     segmentcompare.Segment
	placed   segmentcompare.Placement
	own      float64
	outcome  propertycompare.Outcome
	cause    segmentcompare.Cause
	evidence float64
	neighbor float64
}

var compareCases = func() []compareCase {
	birthP := partKey(vocabkey.PropertyPerson, vocabkey.EventBirth)
	birthE := partKey(vocabkey.PropertyEvent, vocabkey.EventBirth)
	resid := partKey(vocabkey.PropertyPerson, "residence")
	from, to := partOf(vocabkey.PropertyFrom), partOf(vocabkey.PropertyTo)
	york := subjectcompare.Values{vocabkey.PropertyToponym: {pt.Text("York")}}
	spouse := vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "spouse"}
	sibling := vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "sibling"}
	ann := subjectcompare.Values{vocabkey.PropertyName: {pt.Name("Ann Ames")}}
	bob := subjectcompare.Values{vocabkey.PropertyName: {pt.Name("John Smith")}}

	return []compareCase{
		{"AC-1.5", seg(spouse, vocabkey.Person, nil), seg(sibling, vocabkey.Person, nil), segmentcompare.Together, 0, propertycompare.OutcomeUnknown, segmentcompare.CauseNotCorresponding, 0, 0},
		{"AC-2.1", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1817)), segmentcompare.Together, 2.5, propertycompare.OutcomeAgree, "", 2.75, 2.5},
		{"AC-2.2", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1817)), segmentcompare.Apart, 2.5, propertycompare.OutcomeConflict, segmentcompare.CausePlacedApart, -3, 0},
		{"AC-2.3", seg(birthE, vocabkey.Person, nil), seg(birthE, vocabkey.Person, nil), segmentcompare.Apart, 0, propertycompare.OutcomeConflict, segmentcompare.CausePlacedApart, -3, 0},
		{"AC-2.4", seg(resid, vocabkey.Event, event("residence", 0)), seg(resid, vocabkey.Event, event("residence", 0)), segmentcompare.Apart, 2.5, propertycompare.OutcomeUnknown, segmentcompare.CausePlacedApart, 0, 0},
		{"AC-2.5", seg(resid, vocabkey.Event, event("residence", 0)), seg(resid, vocabkey.Event, event("residence", 0)), segmentcompare.Together, 2.5, propertycompare.OutcomeAgree, "", 0.875, 0.5},
		{"AC-2.6", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1830)), segmentcompare.Together, 2.5, propertycompare.OutcomePartial, "", 0.25, -2.5},
		{"AC-2.7", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Together, 1.9, propertycompare.OutcomeUnknown, segmentcompare.CauseGated, 0, 3},
		{"AC-2.8", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Together, 3, propertycompare.OutcomeAgree, "", 1.5, 3},
		{"AC-2.9", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Apart, 3, propertycompare.OutcomeConflict, segmentcompare.CausePlacedApart, -3, 0},
		{"unplaced alike", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1817)), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeAgree, "", 2.75, 2.5},
		{"unplaced differs", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 1830)), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeConflict, segmentcompare.CauseUnplacedDiffers, -3, -2.5},
		{"unplaced unsure", seg(birthP, vocabkey.Event, event("birth", 1817)), seg(birthP, vocabkey.Event, event("birth", 0)), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeUnknown, segmentcompare.CauseUnplacedUnsure, 0, 0.5},
		{"unplaced spouse differs", seg(spouse, vocabkey.Person, ann), seg(spouse, vocabkey.Person, bob), segmentcompare.Unplaced, 2.5, propertycompare.OutcomeUnknown, segmentcompare.CauseUnplacedDiffers, 0, -3},
		{"unplaced gated", seg(from, vocabkey.Place, york), seg(from, vocabkey.Place, york), segmentcompare.Unplaced, 1, propertycompare.OutcomeUnknown, segmentcompare.CauseGated, 0, 3},
		{"AC-2.10", seg(to, vocabkey.Place, york), seg(to, vocabkey.Place, york), segmentcompare.Apart, 3, propertycompare.OutcomeUnknown, segmentcompare.CausePlacedApart, 0, 0},
	}
}()

func TestCompare(t *testing.T) {
	reg, metas := compareReg, compareMetas
	for _, tc := range compareCases {
		t.Run(tc.name, func(t *testing.T) {
			got := segmentcompare.Compare(tc.a, tc.b, tc.placed, tc.own, metas, reg.Segment, reg.Subject, reg.Value)
			if got.Outcome != tc.outcome || got.Cause != tc.cause || !near(got.Evidence, tc.evidence) || !near(got.Neighbor.Evidence, tc.neighbor) {
				t.Errorf("got %s/%q evidence %v neighbor %v; want %s/%q evidence %v neighbor %v",
					got.Outcome, got.Cause, got.Evidence, got.Neighbor.Evidence, tc.outcome, tc.cause, tc.evidence, tc.neighbor)
			}
		})
	}
}

// CompareNeighbor with the neighbors compared by the caller gives Compare's
// result, and asks for the comparison only when the link needs it.
func TestCompareNeighborMatchesCompare(t *testing.T) {
	reg, metas := compareReg, compareMetas
	for _, tc := range compareCases {
		t.Run(tc.name, func(t *testing.T) {
			want := segmentcompare.Compare(tc.a, tc.b, tc.placed, tc.own, metas, reg.Segment, reg.Subject, reg.Value)
			calls := 0
			got := segmentcompare.CompareNeighbor(tc.a, tc.b, tc.placed, tc.own, func() subjectcompare.Result {
				calls++
				return subjectcompare.Compare(tc.a.NeighborKind, tc.a.Neighbor, tc.b.Neighbor, metas, reg.Subject, reg.Value)
			}, reg.Segment)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("CompareNeighbor = %+v, Compare = %+v", got, want)
			}
			wantCalls := 1
			if tc.cause == segmentcompare.CauseNotCorresponding || tc.cause == segmentcompare.CausePlacedApart {
				wantCalls = 0
			}
			if calls != wantCalls {
				t.Errorf("the neighbor comparison ran %d times, want %d", calls, wantCalls)
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

func TestLinkOf(t *testing.T) {
	subject := vocabkey.BridgeTerm(vocabkey.RoleSubject)
	tests := []struct {
		name      string
		bridge    string
		term      vocabkey.BridgeTerm
		eventType vocabkey.EventType
		firstEnd  bool
		want      segmentcompare.Link
	}{
		{"AC-1.1 participation first end", vocabkey.BridgeParticipation, subject, vocabkey.EventBirth, true,
			segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: subject, End: vocabkey.PropertyPerson, Other: vocabkey.PropertyEvent, EventType: vocabkey.EventBirth}},
		{"AC-1.2 participation second end", vocabkey.BridgeParticipation, subject, vocabkey.EventBirth, false,
			segmentcompare.Link{Type: vocabkey.BridgeParticipation, Term: subject, End: vocabkey.PropertyEvent, Other: vocabkey.PropertyPerson, EventType: vocabkey.EventBirth}},
		{"AC-1.3 relationship first end", vocabkey.BridgeRelationship, "parent", vocabkey.EventBirth, true,
			segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: "parent", End: vocabkey.PropertyPerson, Other: vocabkey.PropertyRelatedTo, SameType: true}},
		{"relationship second end", vocabkey.BridgeRelationship, "parent", vocabkey.EventBirth, false,
			segmentcompare.Link{Type: vocabkey.BridgeRelationship, Term: "parent", End: vocabkey.PropertyRelatedTo, Other: vocabkey.PropertyPerson, SameType: true}},
		{"place relationship first end", vocabkey.BridgePlaceRelationship, vocabkey.TermPartOf, "", true,
			segmentcompare.Link{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: vocabkey.PropertyFrom, Other: vocabkey.PropertyTo, SameType: true}},
		{"AC-1.6 location first end", vocabkey.BridgeLocation, "", vocabkey.EventBirth, true,
			segmentcompare.Link{Type: vocabkey.BridgeLocation, End: vocabkey.PropertyEvent, Other: vocabkey.PropertyPlace}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := segmentcompare.LinkOf(segmentcomparetest.Bridge(t, tt.bridge), tt.term, tt.eventType, tt.firstEnd)
			if got != tt.want {
				t.Errorf("LinkOf = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestEventEndpoint(t *testing.T) {
	tests := []struct {
		bridge    string
		first, ok bool
	}{
		{vocabkey.BridgeParticipation, false, true},
		{vocabkey.BridgeRelationship, false, false},
		{vocabkey.BridgePlaceRelationship, false, false},
		{vocabkey.BridgeLocation, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.bridge, func(t *testing.T) {
			first, ok := segmentcompare.EventEndpoint(segmentcomparetest.Bridge(t, tt.bridge))
			if first != tt.first || ok != tt.ok {
				t.Errorf("EventEndpoint = (%v, %v), want (%v, %v)", first, ok, tt.first, tt.ok)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	terms := segmentcomparetest.Terms()
	subject := vocabkey.BridgeTerm(vocabkey.RoleSubject)
	tests := []struct {
		name          string
		bridge        string
		term          vocabkey.BridgeTerm
		eventType     vocabkey.EventType
		first, second vocabkey.BridgeKey
	}{
		{"AC-1.7 parent reads as child with the ends swapped", vocabkey.BridgeRelationship, "parent", "",
			vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "child", End: vocabkey.PropertyRelatedTo},
			vocabkey.BridgeKey{Type: vocabkey.BridgeRelationship, Term: "child", End: vocabkey.PropertyPerson}},
		{"AC-1.8 participation", vocabkey.BridgeParticipation, subject, vocabkey.EventBirth,
			partKey(vocabkey.PropertyPerson, vocabkey.EventBirth), partKey(vocabkey.PropertyEvent, vocabkey.EventBirth)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := segmentcomparetest.Bridge(t, tt.bridge)
			first, second := segmentcompare.Keys(b, tt.term, tt.eventType, terms)
			if first != tt.first || second != tt.second {
				t.Errorf("Keys = (%v, %v), want (%v, %v)", first, second, tt.first, tt.second)
			}
			if first != segmentcompare.Key(segmentcompare.LinkOf(b, tt.term, tt.eventType, true), terms) ||
				second != segmentcompare.Key(segmentcompare.LinkOf(b, tt.term, tt.eventType, false), terms) {
				t.Error("Keys differs from Key(LinkOf(...))")
			}
		})
	}
}
