package graphalign

import (
	"testing"

	"github.com/mendahu/provenencia/core/propertycompare/propertycomparetest"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/subjectcompare"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// A row whose handle has no recorded score is scored against that handle,
// not handed the strong band.
func TestPickScoreScoresAMissingPick(t *testing.T) {
	vals := func(s string) subjectcompare.Values {
		return subjectcompare.Values{vocabkey.PropertyName: {propertycomparetest.Text(s)}}
	}
	layer := Layer{
		Subjects: []Subject{{ID: []byte("s"), Ref: "S", Kind: "person", Values: vals("miller")}},
		Metas:    subjectcompare.Metas{vocabkey.PropertyName: {ValueType: propertymeta.ValueTypeText, Cardinality: propertymeta.CardinalitySingle}},
	}
	canon := Canon{Handles: []Handle{
		{ID: []byte("same"), Ref: "A", Kind: "person", Values: vals("miller")},
		{ID: []byte("other"), Ref: "B", Kind: "person", Values: vals("weaver")},
	}}
	cfg := DefaultConfig()
	st := newState(layer, canon, Stats{}, nil, cfg)

	same := st.pickScore("s", []byte("same"))
	if want := st.scoreCandidate([]byte("s"), st.handles["same"]).score; same.score != want || want >= cfg.StrongScore {
		t.Fatalf("agreeing pick scored %.2f, want its own score %.2f (under strong %.2f)", same.score, want, cfg.StrongScore)
	}
	if other := st.pickScore("s", []byte("other")); other.score >= 0 {
		t.Fatalf("conflicting pick scored %.2f, want a negative score", other.score)
	}
	if gone := st.pickScore("s", []byte("gone")); gone.score != 0 {
		t.Fatalf("a handle outside the canon scored %.2f, want 0", gone.score)
	}
}
