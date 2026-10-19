package subjectvocab

import (
	"testing"

	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/segmentcompare/segmentcomparetest"
	"github.com/mendahu/provenencia/core/vocabkey"
)

// The registry's segment section names only registered bridge types, seeded
// terms and endpoint properties, and the segment fixture's terms are the
// seed's.
func TestSegmentSectionMatchesSeed(t *testing.T) {
	bridges := map[vocabkey.SubjectType]connectrules.Bridge{}
	for _, b := range connectrules.Bridges() {
		bridges[vocabkey.SubjectType(b.BridgeTypeKey)] = b
	}
	terms := map[string]map[string]bool{}
	for _, tm := range seedTerms {
		if terms[tm.PropertyKey] == nil {
			terms[tm.PropertyKey] = map[string]bool{}
		}
		terms[tm.PropertyKey][tm.Key] = true
	}
	s := matchregistry.Default().Segment

	check := func(k vocabkey.BridgeKey, perEnd bool) {
		b, ok := bridges[k.Type]
		if !ok {
			t.Errorf("%s: %q is not a registered bridge type", k, k.Type)
			return
		}
		if !connectrules.HasDisambiguation(b.Disambiguation) || !terms[b.Disambiguation][string(k.Term)] {
			t.Errorf("%s: %q is not a seeded %s term", k, k.Term, b.Disambiguation)
		}
		if k.EventType != "" && !terms[vocabkey.PropertyEventType][string(k.EventType)] {
			t.Errorf("%s: %q is not a seeded event_type term", k, k.EventType)
		}
		if !perEnd {
			if k.End != "" {
				t.Errorf("%s: a One key has an empty End", k)
			}
			return
		}
		found := false
		for _, e := range b.Endpoints {
			found = found || e.PropertyKey == string(k.End)
		}
		if !found {
			t.Errorf("%s: %q is not an endpoint of %s", k, k.End, k.Type)
		}
	}
	for _, k := range s.One {
		check(k, false)
	}
	for k := range s.Gates {
		check(k, true)
	}

	got := segmentcomparetest.Terms()
	for _, b := range connectrules.Bridges() {
		if !connectrules.HasDisambiguation(b.Disambiguation) {
			continue
		}
		typ := vocabkey.SubjectType(b.BridgeTypeKey)
		seen := 0
		for _, tm := range seedTerms {
			if tm.PropertyKey != b.Disambiguation {
				continue
			}
			seen++
			want := struct {
				directed bool
				inverse  vocabkey.BridgeTerm
			}{tm.Directed, vocabkey.BridgeTerm(tm.Inverse)}
			g, ok := got[typ][vocabkey.BridgeTerm(tm.Key)]
			if !ok {
				t.Errorf("segmentcomparetest.Terms has no %s term %q", typ, tm.Key)
				continue
			}
			if g.Directed != want.directed || g.Inverse != want.inverse {
				t.Errorf("segmentcomparetest.Terms %s %q = %+v, seeded directed %v inverse %q", typ, tm.Key, g, want.directed, want.inverse)
			}
		}
		if len(got[typ]) != seen {
			t.Errorf("segmentcomparetest.Terms has %d %s terms, the seed has %d", len(got[typ]), typ, seen)
		}
	}
}
