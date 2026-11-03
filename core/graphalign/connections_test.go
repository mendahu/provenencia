package graphalign_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func TestConnections(t *testing.T) {
	handle := func(h string) graphalign.EndTarget {
		return graphalign.EndTarget{Target: graphalign.TargetHandle, HandleID: id(h)}
	}
	isNew := graphalign.EndTarget{Target: graphalign.TargetNew}
	skip := graphalign.EndTarget{Waits: graphalign.ConnectionSkipped}
	unset := graphalign.EndTarget{Waits: graphalign.ConnectionUnset}
	partOf := []vocabkey.BridgeTerm{vocabkey.TermPartOf}
	place := func(bid string, from, to string, terms []vocabkey.BridgeTerm) graphalign.PageBridge {
		return graphalign.PageBridge{
			ID: id(bid), TypeKey: vocabkey.BridgePlaceRelationship, Terms: terms,
			Ends: [][][]byte{{id(from)}, {id(to)}},
		}
	}
	// A canon edge York part_of Ontario, read from York's end (from).
	york := graphalign.CanonEdge{
		From: id("H-york"), To: id("H-ont"),
		KeyFrom: vocabkey.BridgeKey{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: vocabkey.PropertyFrom},
		KeyTo:   vocabkey.BridgeKey{Type: vocabkey.BridgePlaceRelationship, Term: vocabkey.TermPartOf, End: vocabkey.PropertyTo},
	}
	// The same edge stored the other way round.
	yorkReversed := graphalign.CanonEdge{
		From: id("H-ont"), To: id("H-york"), KeyFrom: york.KeyTo, KeyTo: york.KeyFrom,
	}
	person := graphalign.PageBridge{
		ID: id("b1"), TypeKey: vocabkey.BridgeParticipation,
		Ends: [][][]byte{{id("ann")}, {id("birth")}},
	}
	withEnds := func(b graphalign.PageBridge, ends ...[][]byte) graphalign.PageBridge {
		b.Ends = ends
		return b
	}
	rel := graphalign.PageBridge{
		ID: id("b1"), TypeKey: vocabkey.BridgeRelationship,
		Ends: [][][]byte{{id("ann")}, {id("bob")}},
	}
	ends := func(kv ...any) map[string]graphalign.EndTarget {
		m := map[string]graphalign.EndTarget{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1].(graphalign.EndTarget)
		}
		return m
	}
	tests := []struct {
		name    string
		bridges []graphalign.PageBridge
		ends    map[string]graphalign.EndTarget
		canon   []graphalign.CanonEdge
		want    []graphalign.Connection
	}{
		{"both ends on records", []graphalign.PageBridge{person},
			ends("ann", handle("H-ann"), "birth", handle("H-birth")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionFiles}}},
		{"both ends anchors, no edges", []graphalign.PageBridge{place("b1", "a", "b", partOf)},
			ends("a", handle("H-a"), "b", handle("H-b")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionFiles}}},
		{"a Subject not on the page", []graphalign.PageBridge{person},
			ends("ann", handle("H-ann")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionNotOnPage, Waiting: id("birth")}}},
		{"an end set to Skip", []graphalign.PageBridge{person},
			ends("ann", handle("H-ann"), "birth", skip), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionSkipped, Waiting: id("birth")}}},
		{"an end with no target", []graphalign.PageBridge{person},
			ends("ann", unset, "birth", handle("H-birth")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionUnset, Waiting: id("ann")}}},
		{"the first waiting end in endpoint order", []graphalign.PageBridge{person},
			ends("ann", skip, "birth", unset), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionSkipped, Waiting: id("ann")}}},
		{"an endpoint with no Subject", []graphalign.PageBridge{withEnds(person, [][]byte{id("ann")}, nil)},
			ends("ann", handle("H-ann")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionNotOnPage}}},
		{"Subjects on two records at one end", []graphalign.PageBridge{withEnds(person, [][]byte{id("ann"), id("ann2")}, [][]byte{id("birth")})},
			ends("ann", handle("H-ann"), "ann2", handle("H-ann2"), "birth", handle("H-birth")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionNotOnPage}}},
		{"a typed link with no type", []graphalign.PageBridge{rel},
			ends("ann", handle("H-ann"), "bob", handle("H-bob")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionNotOnPage}}},
		{"both ends one record", []graphalign.PageBridge{person},
			ends("ann", handle("H-x"), "birth", handle("H-x")), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionOutcome("self_link")}}},
		{"a loop with the canon", []graphalign.PageBridge{place("b1", "ont", "york", partOf)},
			ends("ont", handle("H-ont"), "york", handle("H-york")), []graphalign.CanonEdge{york},
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionOutcome("cycle")}}},
		{"a canon edge stored the other way", []graphalign.PageBridge{place("b1", "ont", "york", partOf)},
			ends("ont", handle("H-ont"), "york", handle("H-york")), []graphalign.CanonEdge{yorkReversed},
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionOutcome("cycle")}}},
		{"succeeded_by over the same places", []graphalign.PageBridge{place("b1", "ont", "york", []vocabkey.BridgeTerm{vocabkey.TermSucceededBy})},
			ends("ont", handle("H-ont"), "york", handle("H-york")), []graphalign.CanonEdge{york},
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionFiles}}},
		{"page links in id order", []graphalign.PageBridge{place("b2", "b", "a", partOf), place("b1", "a", "b", partOf)},
			ends("a", isNew, "b", isNew), nil,
			[]graphalign.Connection{
				{BridgeID: id("b1"), Outcome: graphalign.ConnectionFiles},
				{BridgeID: id("b2"), Outcome: graphalign.ConnectionOutcome("cycle")},
			}},
		{"a skipped end frees the loop", []graphalign.PageBridge{place("b1", "a", "b", partOf), place("b2", "b", "a", partOf)},
			ends("a", skip, "b", isNew), nil,
			[]graphalign.Connection{
				{BridgeID: id("b1"), Outcome: graphalign.ConnectionSkipped, Waiting: id("a")},
				{BridgeID: id("b2"), Outcome: graphalign.ConnectionSkipped, Waiting: id("a")},
			}},
		{"New and New", []graphalign.PageBridge{place("b1", "a", "b", partOf)},
			ends("a", isNew, "b", isNew), nil,
			[]graphalign.Connection{{BridgeID: id("b1"), Outcome: graphalign.ConnectionFiles}}},
		{"an unknown bridge type is left out", []graphalign.PageBridge{{ID: id("b1"), TypeKey: "nope"}},
			nil, nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := graphalign.Connections(tc.bridges, tc.ends, tc.canon)
			if len(got) != len(tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
			for i, w := range tc.want {
				g := got[i]
				if string(g.BridgeID) != string(w.BridgeID) || g.Outcome != w.Outcome || string(g.Waiting) != string(w.Waiting) {
					t.Fatalf("connection %d = %+v, want %+v", i, g, w)
				}
			}
		})
	}
}

func TestConnectionOutcomesListEveryRefusal(t *testing.T) {
	got := map[graphalign.ConnectionOutcome]bool{}
	for _, o := range graphalign.ConnectionOutcomes() {
		got[o] = true
	}
	for _, want := range []graphalign.ConnectionOutcome{"files", "skipped", "unset", "not_on_page", "self_link", "cycle"} {
		if !got[want] {
			t.Fatalf("missing outcome %q", want)
		}
	}
	if len(got) != 6 {
		t.Fatalf("outcomes %v", got)
	}
}
