package graphalign_test

import (
	"fmt"
	"math"
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
	sexed := func(ref, form, sex string) graphalign.Handle {
		k, v := termProp("sex_at_birth", sex)
		nk, nv := nameProp(form)
		return graphalign.Handle{ID: id("h-" + ref), Ref: ref, Kind: "person", Values: vals(nk, nv, k, v)}
	}
	refs := func(hs []graphalign.Handle) []string {
		var out []string
		for _, h := range hs {
			out = append(out, h.Ref)
		}
		return out
	}

	t.Run("a handle under the similarity floor is not a candidate", func(t *testing.T) {
		handles := []graphalign.Handle{
			person("PER-3", "James Robins"),
			person("PER-1", "James Robins"),
			person("PER-2", "Mary Smith"),
		}
		probe := vals(nameProp("James Robins"))
		got := refs(graphalign.Candidates("person", probe, handles, metas, cfg))
		if want := []string{"PER-1", "PER-3"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("only the probe's kind", func(t *testing.T) {
		probe := vals(nameProp("James Robins"))
		handles := []graphalign.Handle{person("PER-1", "James Robins"), {ID: id("h-plc"), Ref: "PLC-1", Kind: "place", Values: probe}}
		got := refs(graphalign.Candidates("person", probe, handles, metas, cfg))
		if want := []string{"PER-1"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("an empty probe is offered nothing", func(t *testing.T) {
		if got := graphalign.Candidates("person", subjectcompare.Values{}, []graphalign.Handle{person("PER-1", "James Robins")}, metas, cfg); len(got) != 0 {
			t.Fatalf("empty probe offered %v", refs(got))
		}
	})
	t.Run("a name with a sex conflict is kept", func(t *testing.T) {
		k, v := termProp("sex_at_birth", "male")
		nk, nv := nameProp("Ann Ames")
		probe := vals(nk, nv, k, v)
		got := refs(graphalign.Candidates("person", probe, []graphalign.Handle{sexed("PER-1", "Ann Ames", "female")}, metas, cfg))
		if want := []string{"PER-1"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("AC-2.3 every handle tied at the K-th place is kept", func(t *testing.T) {
		k, v := termProp("sex_at_birth", "male")
		nk, nv := nameProp("Ann Ames")
		probe := vals(nk, nv, k, v)
		var handles []graphalign.Handle
		for _, ref := range []string{"PER-1", "PER-2", "PER-3"} {
			handles = append(handles, sexed(ref, "Ann Ames", "male"))
		}
		for _, ref := range []string{"PER-4", "PER-5", "PER-6", "PER-7"} {
			handles = append(handles, sexed(ref, "Ann Ames", "female"))
		}
		got := refs(graphalign.Candidates("person", probe, handles, metas, cfg))
		if want := []string{"PER-1", "PER-2", "PER-3", "PER-4", "PER-5", "PER-6", "PER-7"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("AC-2.4 ties stop at the cap, first by Ref", func(t *testing.T) {
		var handles []graphalign.Handle
		for i := 30; i >= 1; i-- {
			handles = append(handles, person(fmt.Sprintf("PER-%02d", i), "John Smith"))
		}
		got := refs(graphalign.Candidates("person", vals(nameProp("John Smith")), handles, metas, cfg))
		if len(got) != cfg.Alignment.TieCap || got[0] != "PER-01" || got[len(got)-1] != fmt.Sprintf("PER-%02d", cfg.Alignment.TieCap) {
			t.Fatalf("candidates %v, want the first %d by Ref", got, cfg.Alignment.TieCap)
		}
	})
	t.Run("AC-2.5 a death is not offered for a birth", func(t *testing.T) {
		event := func(ref, kind string, year int) graphalign.Handle {
			return graphalign.Handle{ID: id("h-" + ref), Ref: ref, Kind: "event", Values: eventVals(kind, year)}
		}
		got := refs(graphalign.Candidates("event", eventVals("birth", 1850), []graphalign.Handle{event("EVT-1", "death", 1850), event("EVT-2", "birth", 1850)}, metas, cfg))
		if want := []string{"EVT-2"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("AC-1.1 one given name reaches a record with three more names", func(t *testing.T) {
		amy := vals(givenProp("Amy"))
		handles := []graphalign.Handle{
			person("PER-1", "Amy Lee Anne Long"),
			person("PER-2", "Beth Amy Long"),
		}
		sim := subjectcompare.Compare(vocabkey.Person, amy, handles[0].Values, metas, cfg.Subject, cfg.Value).Similarity
		if math.Abs(sim-2/4.5) > 1e-9 {
			t.Fatalf("similarity %v, want %v", sim, 2/4.5)
		}
		if sim < cfg.Alignment.MinSimilarity {
			t.Errorf("similarity %v is under the gate %v", sim, cfg.Alignment.MinSimilarity)
		}
		got := refs(graphalign.Candidates("person", amy, handles, metas, cfg))
		if want := []string{"PER-1"}; !slices.Equal(got, want) {
			t.Fatalf("candidates %v, want %v", got, want)
		}
	})
	t.Run("the best K by similarity when nothing ties", func(t *testing.T) {
		small := cfg
		small.Alignment.TopK = 1
		got := refs(graphalign.Candidates("person", vals(nameProp("James Robins")), []graphalign.Handle{person("PER-1", "James Robins"), person("PER-2", "J Robins")}, metas, small))
		if len(got) == 0 || got[0] != "PER-1" || (len(got) > 1 && got[1] == "PER-1") {
			t.Fatalf("candidates %v, want PER-1 first", got)
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
