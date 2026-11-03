package connectrules

import (
	"errors"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/vocabkey"
)

func TestLinksClosesCycle(t *testing.T) {
	chain := []Link{{"a", "b"}, {"b", "c"}}
	split := []Link{{"a", "b"}, {"a", "c"}}
	tests := []struct {
		name     string
		edges    []Link
		from, to string
		want     bool
	}{
		{"closing a chain", chain, "c", "a", true},
		{"extending a chain", chain, "a", "c", false},
		{"a split", split, "b", "c", false},
		{"a link to itself", nil, "a", "a", true},
		{"edges that already loop", []Link{{"a", "b"}, {"b", "a"}}, "c", "d", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewLinks(tc.edges).ClosesCycle(tc.from, tc.to); got != tc.want {
				t.Fatalf("ClosesCycle = %v, want %v", got, tc.want)
			}
		})
	}
	t.Run("an added link counts", func(t *testing.T) {
		s := NewLinks(nil)
		s.Add(Link{"a", "b"})
		if !s.ClosesCycle("b", "a") {
			t.Fatal("a→b was added, so b→a closes a loop")
		}
	})
}

func TestDecide(t *testing.T) {
	place, _ := LookupBridge("place_relationship")
	rel, _ := LookupBridge("relationship")
	part, _ := LookupBridge("participation")
	rec := func(r string) []Value { return []Value{{Record: r}} }
	partOf := []vocabkey.BridgeTerm{vocabkey.TermPartOf}
	boom := errors.New("boom")
	edges := func(e ...Link) func() (*Links, error) {
		return func() (*Links, error) { return NewLinks(e), nil }
	}
	never := func() (*Links, error) {
		t.Fatal("filed must not be called")
		return nil, nil
	}
	tests := []struct {
		name  string
		b     Bridge
		terms []vocabkey.BridgeTerm
		ends  [][]Value
		filed func() (*Links, error)
		want  Result
		err   error
	}{
		{"term order is free", place, []vocabkey.BridgeTerm{vocabkey.TermSucceededBy, vocabkey.TermPartOf},
			[][]Value{rec("a"), rec("b")}, edges(Link{"b", "a"}), Result{Decision: DecisionCycle}, nil},
		{"term order is free, other way", place, []vocabkey.BridgeTerm{vocabkey.TermPartOf, vocabkey.TermSucceededBy},
			[][]Value{rec("a"), rec("b")}, edges(Link{"b", "a"}), Result{Decision: DecisionCycle}, nil},
		{"an endpoint with no value", place, partOf,
			[][]Value{rec("a"), nil}, never, Result{Decision: DecisionIncomplete}, nil},
		{"first waiting value", place, partOf,
			[][]Value{rec("a"), {{Record: "b"}, {Wait: "skipped"}, {Wait: "unset"}}}, never,
			Result{Decision: DecisionWaits, Endpoint: 1, Value: 1, Wait: "skipped"}, nil},
		{"endpoint 0 on two records decides before a wait on 1", place, partOf,
			[][]Value{{{Record: "a"}, {Record: "b"}}, {{Record: "c"}}}, never,
			Result{Decision: DecisionIncomplete}, nil},
		{"one record", place, partOf, [][]Value{rec("a"), rec("a")}, never,
			Result{Decision: DecisionSelfLink}, nil},
		{"self-link before term", place, nil, [][]Value{rec("a"), rec("a")}, never,
			Result{Decision: DecisionSelfLink}, nil},
		{"typed link with no term", rel, nil, [][]Value{rec("a"), rec("b")}, never,
			Result{Decision: DecisionIncomplete}, nil},
		{"role needs no term", part, nil, [][]Value{rec("a"), rec("b")}, never,
			Result{Decision: DecisionFiles, Records: []string{"a", "b"}}, nil},
		{"a filing carries its ends in endpoint order and the link it checked", place, partOf,
			[][]Value{{{Record: "a"}, {Record: "a"}}, rec("b")}, edges(Link{"c", "d"}),
			Result{Decision: DecisionFiles, Records: []string{"a", "b"}, Link: &Link{From: "a", To: "b"}}, nil},
		{"closes part_of", place, partOf, [][]Value{rec("a"), rec("b")}, edges(Link{"b", "a"}),
			Result{Decision: DecisionCycle}, nil},
		{"succeeded_by beside part_of", place, []vocabkey.BridgeTerm{vocabkey.TermSucceededBy},
			[][]Value{rec("a"), rec("b")}, edges(), Result{Decision: DecisionFiles, Records: []string{"a", "b"}, Link: &Link{From: "a", To: "b"}}, nil},
		{"spouse is not loop-checked", rel, []vocabkey.BridgeTerm{vocabkey.TermSpouse},
			[][]Value{rec("a"), rec("b")}, never, Result{Decision: DecisionFiles, Records: []string{"a", "b"}}, nil},
		{"filed fails", place, partOf, [][]Value{rec("a"), rec("b")},
			func() (*Links, error) { return nil, boom }, Result{}, boom},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Decide(tc.b, tc.terms, tc.ends, tc.filed)
			if !errors.Is(err, tc.err) {
				t.Fatalf("err = %v, want %v", err, tc.err)
			}
			if !sameResult(got, tc.want) {
				t.Fatalf("Decide = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func sameResult(a, b Result) bool {
	if a.Decision != b.Decision || a.Endpoint != b.Endpoint || a.Value != b.Value || a.Wait != b.Wait ||
		!slices.Equal(a.Records, b.Records) || (a.Link == nil) != (b.Link == nil) {
		return false
	}
	return a.Link == nil || *a.Link == *b.Link
}

func TestTermIsTheLowestKey(t *testing.T) {
	got := Term([]vocabkey.BridgeTerm{vocabkey.TermSucceededBy, "", vocabkey.TermPartOf})
	if got != vocabkey.TermPartOf {
		t.Fatalf("Term = %q", got)
	}
	if Term(nil) != "" {
		t.Fatal("no terms, no term")
	}
}
