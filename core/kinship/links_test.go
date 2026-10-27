package kinship_test

import (
	"bytes"
	"reflect"
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/kinship"
	"github.com/mendahu/provenencia/core/segmentcompare"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The link set merges a recorded link and an implied one of the same ends and
// keys into one, and merges the links of two marriages of one couple.
func TestLinkSet(t *testing.T) {
	linkSetOf := func(links []kinship.Link, terms segmentcompare.Terms, rules kinship.Rules) []kinship.Link {
		s := kinship.NewLinkSet()
		for _, l := range links {
			s.Add(l)
		}
		s.WithImplied(terms, rules)
		return s.Links()
	}
	person := func(n byte) []byte { return []byte{0x10 + n} }
	event := func(n byte) []byte { return []byte{0x20 + n} }
	bridgeID := func(n byte) []byte { return []byte{0x30 + n} }
	part := func(p, e, b byte) kinship.Link {
		ka, kb := segmentcompare.Keys(segmentcomparetest.Bridge(t, vocabkey.BridgeParticipation), "subject", "marriage", terms)
		return kinship.Link{Edge: kinship.Edge{A: person(p), B: event(e), KeyA: ka, KeyB: kb, Associations: [][]byte{bridgeID(b)}}}
	}
	spouse := func(p, q, b byte) kinship.Link {
		ka, kb := segmentcompare.Keys(segmentcomparetest.Bridge(t, vocabkey.BridgeRelationship), "spouse", "", terms)
		return kinship.Link{Edge: kinship.Edge{A: person(p), B: person(q), KeyA: ka, KeyB: kb, Associations: [][]byte{bridgeID(b)}}}
	}
	flip := func(l kinship.Link) kinship.Link {
		l.A, l.B, l.KeyA, l.KeyB = l.B, l.A, l.KeyB, l.KeyA
		return l
	}
	t.Run("the same recorded link from each end is one", func(t *testing.T) {
		got := linkSetOf([]kinship.Link{spouse(1, 2, 1), flip(spouse(1, 2, 2))}, terms, rules)
		if len(got) != 1 || len(got[0].Associations) != 2 || got[0].OnlyImplied || len(got[0].Shapes) != 0 {
			t.Fatalf("links %+v, want one recorded link with 2 associations", got)
		}
	})
	t.Run("a recorded spouse link and a marriage are one link", func(t *testing.T) {
		got := linkSetOf([]kinship.Link{spouse(1, 2, 1), part(1, 1, 2), part(2, 1, 3)}, terms, rules)
		var spouses []kinship.Link
		for _, l := range got {
			if l.KeyA.Type == vocabkey.BridgeRelationship {
				spouses = append(spouses, l)
			}
		}
		if len(spouses) != 1 || len(spouses[0].Associations) != 3 || len(spouses[0].Shapes) != 1 || spouses[0].OnlyImplied {
			t.Fatalf("spouse links %+v, want one with 3 associations, one shape, recorded", spouses)
		}
	})
	t.Run("two marriages of one couple are one implied link", func(t *testing.T) {
		got := linkSetOf([]kinship.Link{part(1, 1, 1), part(2, 1, 2), part(1, 2, 3), part(2, 2, 4)}, terms, rules)
		var spouses []kinship.Link
		for _, l := range got {
			if l.KeyA.Type == vocabkey.BridgeRelationship {
				spouses = append(spouses, l)
			}
		}
		if len(spouses) != 1 || len(spouses[0].Associations) != 4 || len(spouses[0].Shapes) != 2 || !spouses[0].OnlyImplied {
			t.Fatalf("spouse links %+v, want one implied link with 4 associations and 2 shapes", spouses)
		}
	})
	t.Run("the same implied link added twice is one shape", func(t *testing.T) {
		imp := spouse(1, 2, 1)
		imp.Recording = kinship.Recording{Shapes: []kinship.Shape{{Pattern: kinship.PatternMarriageSpouses, Pivot: event(1)}}, OnlyImplied: true}
		got := linkSetOf([]kinship.Link{imp, imp}, terms, rules)
		if len(got) != 1 || len(got[0].Shapes) != 1 || !got[0].OnlyImplied {
			t.Fatalf("links %+v, want one implied link with one shape", got)
		}
	})
	t.Run("an implied link does not alias the input", func(t *testing.T) {
		in := []kinship.Link{part(1, 1, 1), part(2, 1, 2)}
		got := linkSetOf(in, terms, rules)
		var imp kinship.Link
		for _, l := range got {
			if l.KeyA.Type == vocabkey.BridgeRelationship {
				imp = l
			}
		}
		if imp.A == nil {
			t.Fatalf("no implied link in %+v", got)
		}
		want := bytes.Clone(imp.A)
		in[0].A[0], in[1].A[0] = 0xff, 0xff
		if !bytes.Equal(imp.A, want) {
			t.Fatalf("implied A %x changed with the input, want %x", imp.A, want)
		}
	})
	t.Run("PlaceImplied places without growing the set", func(t *testing.T) {
		impliedSpouse := func(p, q, m byte) kinship.Implied {
			return kinship.Implied{Edge: spouse(p, q, m).Edge, Pattern: kinship.PatternMarriageSpouses, Event: event(m)}
		}
		s := kinship.NewLinkSet()
		s.Add(spouse(1, 2, 1))
		got := s.PlaceImplied([]kinship.Implied{impliedSpouse(1, 2, 1), impliedSpouse(3, 4, 2), impliedSpouse(4, 3, 3)})
		if !slices.Equal(got, []int{0, 1, 1}) || len(s.Links()) != 1 {
			t.Fatalf("placed %v over %d links, want [0 1 1] over 1", got, len(s.Links()))
		}
	})
	t.Run("an added implied link records as Implied.Recording", func(t *testing.T) {
		imp := kinship.Implied{Edge: spouse(1, 2, 1).Edge, Pattern: kinship.PatternMarriageSpouses, Event: event(1)}
		s := kinship.NewLinkSet()
		s.AddImplied(imp)
		got := s.Links()
		if len(got) != 1 || !reflect.DeepEqual(got[0].Recording, imp.Recording()) {
			t.Fatalf("links %+v, want one carrying %+v", got, imp.Recording())
		}
	})
}
