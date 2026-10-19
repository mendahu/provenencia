package graphalign_test

import (
	"slices"
	"testing"

	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/subjectcompare/subjectcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

func TestCandidates(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	metas := subjectcomparetest.Metas()
	person := func(ref, form string) graphalign.Handle {
		return graphalign.Handle{ID: id("h-" + ref), Ref: ref, Kind: "person", Values: vals(nameProp(form))}
	}
	handles := []graphalign.Handle{
		person("PER-3", "James Robins"),
		person("PER-1", "James Robins"),
		person("PER-2", "Robins"),
		person("PER-4", "Mary Robins"),
	}
	probe := vals(nameProp("James Robins"))
	refs := func(hs []graphalign.Handle) []string {
		var out []string
		for _, h := range hs {
			out = append(out, h.Ref)
		}
		return out
	}

	// The precondition: the look-alikes are under the weak band, one by far.
	evidence := func(h graphalign.Handle) float64 {
		return subjectcompare.Compare(vocabkey.Person, probe, h.Values, metas, cfg.Subject, cfg.Value).Evidence
	}
	if e := evidence(handles[2]); e >= cfg.Alignment.Bands.Weak {
		t.Fatalf("PER-2 evidence %.3f, want under %.1f", e, cfg.Alignment.Bands.Weak)
	}
	if e := evidence(handles[3]); e >= 0 {
		t.Fatalf("PER-4 evidence %.3f, want negative", e)
	}

	t.Run("AC-2.1", func(t *testing.T) {
		got := refs(graphalign.Candidates("person", probe, handles, metas, cfg, 0))
		if want := []string{"PER-1", "PER-3"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("AC-2.2", func(t *testing.T) {
		got := refs(graphalign.Candidates("person", probe, handles, metas, cfg, 1))
		if want := []string{"PER-1"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("AC-2.3", func(t *testing.T) {
		withPlace := append(slices.Clone(handles), graphalign.Handle{ID: id("h-plc"), Ref: "PLC-1", Kind: "place", Values: probe})
		got := refs(graphalign.Candidates("person", probe, withPlace, metas, cfg, 0))
		if slices.Contains(got, "PLC-1") {
			t.Fatalf("candidates %v hold a place", got)
		}
	})
	t.Run("AC-2.4", func(t *testing.T) {
		if got := graphalign.Candidates("person", subjectcompare.Values{}, handles, metas, cfg, 0); len(got) != 0 {
			t.Fatalf("empty probe offered %v", refs(got))
		}
	})
}

// A pair that shares no shortlist key is never offered, so the evidence a kind
// can earn on leads that are not keyed must stay under the weak band.
func TestUnkeyedEvidenceStaysUnderTheWeakBand(t *testing.T) {
	cfg := matchregistry.Default().AlignConfig()
	for kind, leads := range cfg.Subject.Weights {
		keyed := map[vocabkey.PropertyKey]bool{}
		for _, lead := range cfg.Alignment.Shortlist.Words[kind] {
			keyed[lead] = true
		}
		if d, ok := cfg.Alignment.Shortlist.Dated[kind]; ok {
			keyed[d.Date] = true
		}
		var most float64
		for lead, w := range leads {
			if !keyed[lead] {
				most += w.Agree
			}
		}
		if most >= cfg.Alignment.Bands.Weak {
			t.Errorf("%s: unkeyed leads can earn %.2f, want under %.2f", kind, most, cfg.Alignment.Bands.Weak)
		}
	}
}
