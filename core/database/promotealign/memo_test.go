package promotealign_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/evidencelayer"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/namealiases"
	"github.com/mendahu/provenencia/core/database/nameformats"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/writes"
)

// warmFixture files a person with a dated birth on the fixture's Source and
// holds an incoming record of the same on another Source.
type warmFixture struct {
	f                         *fixture
	srcB                      sources.Source
	artB                      artifacts.Artifact
	member                    subjects.Subject
	bPerson, bBirth           subjects.Subject
	personHandle, birthHandle []byte
}

func newWarmFixture(t *testing.T) *warmFixture {
	t.Helper()
	f := newFixture(t)
	member, _, ph, bh := f.filedBirth(f.source, f.artifact, "John Smith", 1817, nil, nil)
	srcB, artB := f.NewSource("Obituary")
	bPerson := f.Person(srcB, artB, "John Smith")
	bBirth := f.DatedBirth(srcB, artB, 1817)
	f.Participation(srcB, artB, bPerson, bBirth, "subject")
	return &warmFixture{f: f, srcB: srcB, artB: artB, member: member, bPerson: bPerson, bBirth: bBirth, personHandle: ph, birthHandle: bh}
}

// T-18 (AC-4.2): a second Propose on one Source, with no commit between,
// issues no query through the graph. The first built the shortlist index.
func TestProposeWarmIssuesNoQuery(t *testing.T) {
	w := newWarmFixture(t)
	cold := w.f.propose(w.f.Graph(w.srcB), nil)
	n := 0
	w.f.c.Graph().SetQueryCounterForTest(func() { n++ })
	warm := w.f.propose(w.f.Graph(w.srcB), nil)
	if n != 0 {
		t.Fatalf("the second Propose issued %d graph queries", n)
	}
	if !reflect.DeepEqual(cold, warm) {
		t.Fatalf("warm Propose differs from cold:\ncold %+v\nwarm %+v", cold, warm)
	}
}

// T-19 (AC-4.3): after a commit that changes a member, a title or a handle,
// Propose follows it: the warm output is a fresh read's.
func TestMemoFollowsCommit(t *testing.T) {
	// compare warms the caches, applies the commit, and checks that the
	// output changed, and equals what a cold graph reads.
	compare := func(t *testing.T, w *warmFixture, commit func()) {
		t.Helper()
		beforeP := w.f.propose(w.f.Graph(w.srcB), nil)
		commit()
		warmP, _, warmPErr := promotealign.Propose(w.f.c, w.f.Graph(w.srcB), nil, nil)
		w.f.c.Graph().Drop()
		freshP, _, freshPErr := promotealign.Propose(w.f.c, w.f.Graph(w.srcB), nil, nil)
		if !sameErr(warmPErr, freshPErr) {
			t.Fatalf("errors differ: warm %v, fresh %v", warmPErr, freshPErr)
		}
		if !reflect.DeepEqual(warmP, freshP) {
			t.Errorf("warm Propose follows the commit wrongly:\nwarm  %+v\nfresh %+v", warmP, freshP)
		}
		if reflect.DeepEqual(beforeP, freshP) && freshPErr == nil {
			t.Errorf("the commit changed nothing the test can see")
		}
	}

	t.Run("a member's Observation edited", func(t *testing.T) {
		w := newWarmFixture(t)
		compare(t, w, func() {
			listed, err := observations.ListBySubject(w.f.c, w.member.ID)
			if err != nil {
				t.Fatal(err)
			}
			for _, l := range listed {
				if l.PropertyID == nil || string(l.PropertyID) != string(w.f.Prop("name").ID) {
					continue
				}
				_, err := writes.Call(w.f.c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
					return observations.Update(tx, userID, observations.Input{
						ID: l.ID, SubjectID: w.member.ID, PropertyID: l.PropertyID, Name: namevaluetest.Western("Jon Smyth"),
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			t.Fatal("the member has no name Observation")
		})
	})

	t.Run("a handle merged", func(t *testing.T) {
		w := newWarmFixture(t)
		compare(t, w, func() {
			other := w.f.Person(w.f.source, w.f.artifact, "Somebody Else")
			into := w.f.promote(other).Entity.ID
			db, err := w.f.c.DB()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, into, w.personHandle); err != nil {
				t.Fatal(err)
			}
			if err := w.f.c.Graph().OnCommit(effects.Commit{
				Activity: effects.Activity{Revision: 1 << 40}, Effects: effects.Set{Handles: [][]byte{w.personHandle}},
			}); err != nil {
				t.Fatal(err)
			}
		})
	})
}

func sameErr(a, b error) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return errors.Is(a, promote.ErrInvalid) == errors.Is(b, promote.ErrInvalid)
}

// T-19 (D-21): a memoized layer and member values are never changed by a
// reader: Propose leaves them as they were stored.
func TestMemoIsNotMutated(t *testing.T) {
	w := newWarmFixture(t)
	w.f.propose(w.f.Graph(w.srcB), nil)
	first := w.snapshot(t)
	if len(first.Subjects) == 0 {
		t.Fatalf("nothing was memoized: %+v", first)
	}
	before := memoJSON(t, w)
	for i := 0; i < 2; i++ {
		w.f.propose(w.f.Graph(w.srcB), []graphalign.Fixed{{SubjectID: w.bPerson.ID, HandleID: w.personHandle}})
		if after := memoJSON(t, w); after != before {
			t.Fatalf("a read changed the memo:\nbefore %s\nafter  %s", before, after)
		}
		// The same layer, not a rebuilt copy.
		again := w.snapshot(t)
		if &again.Subjects[0] != &first.Subjects[0] {
			t.Fatal("the layer was rebuilt, not read from the memo")
		}
	}
}

// T-3 (AC-1.1): Propose holds the layer in the Layers entry, and
// evidencelayer.Load reads that layer with no query; a second Load shares it.
func TestLoadSharesPromotesMemo(t *testing.T) {
	w := newWarmFixture(t)
	graph := w.f.Graph(w.srcB)
	w.f.propose(graph, nil)
	stored, ok := evidencelayer.Layers.Peek(w.f.c.Graph(), graph)
	if !ok || len(stored.Subjects) == 0 {
		t.Fatalf("Propose held no layer with Subjects: %v, %v", stored, ok)
	}
	n := 0
	w.f.c.Graph().SetQueryCounterForTest(func() { n++ })
	got, err := evidencelayer.Load(w.f.c.Graph(), graph)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("Load issued %d graph queries after a Propose", n)
	}
	if &got.Subjects[0] != &stored.Subjects[0] {
		t.Fatal("Load built its own layer, not the one Propose held")
	}
	again, err := evidencelayer.Load(w.f.c.Graph(), graph)
	if err != nil {
		t.Fatal(err)
	}
	if &again.Subjects[0] != &got.Subjects[0] {
		t.Fatal("a second Load did not share the held layer")
	}
}

// entriesWorld is two incoming graphs, each proposed, so the Layers and
// candidates entries hold both.
type entriesWorld struct {
	w    *warmFixture
	a, b []byte
	artB artifacts.Artifact
}

func newEntriesWorld(t *testing.T) entriesWorld {
	t.Helper()
	w := newWarmFixture(t)
	srcC, artC := w.f.NewSource("Register")
	w.f.Person(srcC, artC, "Mary Jones")
	e := entriesWorld{w: w, a: w.f.Graph(w.srcB), b: w.f.Graph(srcC), artB: w.artB}
	w.f.propose(e.a, nil)
	w.f.propose(e.b, nil)
	for _, id := range [][]byte{e.a, e.b} {
		if !e.layer(id) || !e.candidates(id) {
			t.Fatalf("after Propose graph %x holds layer %v, candidates %v", id, e.layer(id), e.candidates(id))
		}
	}
	return e
}

func (e entriesWorld) layer(id []byte) bool {
	_, ok := evidencelayer.Layers.Peek(e.w.f.c.Graph(), id)
	return ok
}

func (e entriesWorld) candidates(id []byte) bool {
	return promotealign.CandidatesHeldForTest(e.w.f.c.Graph(), id)
}

// AC-4.1, AC-4.2, AC-4.3, AC-4.8: the Layers and candidates entries empty on
// a vocabulary commit and a name format change, drop only the graph a write
// names, and pass Verify after a proposal.
func TestEntriesFollowSignals(t *testing.T) {
	none := func(t *testing.T, e entriesWorld) {
		t.Helper()
		counts := e.w.f.c.Graph().KeyCounts()
		if counts["evidencelayer.layers"] != 0 || counts["promotealign.candidates"] != 0 {
			t.Fatalf("held %d layers and %d candidates, want none", counts["evidencelayer.layers"], counts["promotealign.candidates"])
		}
	}
	t.Run("a vocabulary commit", func(t *testing.T) {
		e := newEntriesWorld(t)
		if err := e.w.f.c.Graph().OnCommit(effects.Commit{
			Activity: effects.Activity{Revision: 1 << 40}, Effects: effects.Set{Lists: []effects.List{effects.ListPropertyVocabulary}},
		}); err != nil {
			t.Fatal(err)
		}
		none(t, e)
	})
	t.Run("a name format change", func(t *testing.T) {
		e := newEntriesWorld(t)
		if err := nameformats.Invalidate(e.w.f.c); err != nil {
			t.Fatal(err)
		}
		none(t, e)
	})
	t.Run("a write that names a graph", func(t *testing.T) {
		e := newEntriesWorld(t)
		e.w.f.Cite(e.artB, e.w.bPerson, observations.Input{
			PropertyID: e.w.f.Prop("name").ID, Name: namevaluetest.Western("Johnny Smith"),
		})
		if e.layer(e.a) || e.candidates(e.a) {
			t.Errorf("the written graph kept layer %v, candidates %v", e.layer(e.a), e.candidates(e.a))
		}
		if !e.layer(e.b) || !e.candidates(e.b) {
			t.Errorf("the other graph lost layer %v, candidates %v", e.layer(e.b), e.candidates(e.b))
		}
	})
	t.Run("Verify after a proposal", func(t *testing.T) {
		t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")
		e := newEntriesWorld(t)
		if err := e.w.f.c.Graph().OnCommit(effects.Commit{
			Activity: effects.Activity{Revision: 2000}, Effects: effects.Set{SourceScopes: [][]byte{e.w.f.source.ID}},
		}); err != nil {
			t.Fatalf("Verify after a proposal: %v", err)
		}
	})
}

// AC-4.9: Verify reads the entries without counting a use, so a commit with
// Verify on leaves every entry's use order as one with Verify off: the key
// that leaves next is the same.
func TestVerifyLeavesEntriesAlone(t *testing.T) {
	for _, verify := range []string{"1", "0"} {
		t.Run("verify="+verify, func(t *testing.T) {
			t.Setenv("PROVENENCIA_GRAPH_VERIFY", verify)
			e := newEntriesWorld(t)
			g := e.w.f.c.Graph()
			// Six graphs, so a Check that touched keys in map order would
			// leave a least recently used only one time in six.
			ids := [][]byte{e.a, e.b}
			for i := 0; i < 4; i++ {
				src, art := e.w.f.NewSource(fmt.Sprintf("Extra %d", i))
				e.w.f.Person(src, art, "Mary Jones")
				id := e.w.f.Graph(src)
				e.w.f.propose(id, nil)
				ids = append(ids, id)
			}
			// A fresh read of every graph's records and loaded graph, in order, so a is the least recently used of each entry.
			for _, id := range ids {
				if _, err := graphfill.EvidenceGraph(g, id); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := graphfill.RecordsOf(g, ids); err != nil {
				t.Fatal(err)
			}
			for _, id := range ids {
				graphfill.Records.Peek(g, id)
				if _, err := evidencelayer.Load(g, id); err != nil {
					t.Fatal(err)
				}
				if !e.candidates(id) {
					t.Fatalf("graph %x holds no candidates", id)
				}
			}
			// Full caps on every entry, then a commit naming none of the graphs.
			graphfill.Graphs.SetCapForTest(g, len(ids))
			graphfill.Records.SetCapForTest(g, len(ids))
			evidencelayer.Layers.SetCapForTest(g, len(ids))
			promotealign.SetCandidatesCapForTest(g, len(ids))
			t.Cleanup(func() {
				graphfill.Graphs.SetCapForTest(g, 0)
				graphfill.Records.SetCapForTest(g, 0)
				evidencelayer.Layers.SetCapForTest(g, 0)
				promotealign.SetCandidatesCapForTest(g, 0)
			})
			if err := g.OnCommit(effects.Commit{
				Activity: effects.Activity{Revision: 3000}, Effects: effects.Set{SourceScopes: [][]byte{e.w.f.source.ID}},
			}); err != nil {
				t.Fatal(err)
			}
			// Another key arrives at each full entry: a, the least recently
			// used, leaves and the rest stay.
			c := bytes.Repeat([]byte{9}, 16)
			evidencelayer.Layers.Set(g, c, evidencelayer.Layer{})
			graphfill.Graphs.Set(g, c, graphcache.EvidenceGraph{})
			graphfill.Records.Set(g, c, nil)
			promotealign.HoldCandidatesForTest(g, c)
			for name, got := range map[string][2]bool{
				"graphs":     held(ids, func(id []byte) bool { return peekGraph(g, id) }),
				"records":    held(ids, func(id []byte) bool { return peekRecords(g, id) }),
				"layers":     held(ids, e.layer),
				"candidates": held(ids, e.candidates),
			} {
				if got[0] || !got[1] {
					t.Errorf("%s: a held %v, the rest held %v, want a evicted and the rest held", name, got[0], got[1])
				}
			}
		})
	}
}

// held is whether the first id is held, and whether all the others are.
func held(ids [][]byte, has func([]byte) bool) [2]bool {
	rest := true
	for _, id := range ids[1:] {
		rest = rest && has(id)
	}
	return [2]bool{has(ids[0]), rest}
}

func peekGraph(g *graphcache.Graph, id []byte) bool {
	_, ok := graphfill.PeekEvidenceGraph(g, id)
	return ok
}

func peekRecords(g *graphcache.Graph, id []byte) bool {
	_, ok := graphfill.Records.Peek(g, id)
	return ok
}

// snapshot is the layer of the incoming Source's graph, as the memo holds it
// once a Propose has run.
func (w *warmFixture) snapshot(t *testing.T) evidencelayer.Layer {
	t.Helper()
	return w.f.Layer(w.f.Graph(w.srcB))
}

func memoJSON(t *testing.T, w *warmFixture) string {
	t.Helper()
	snap := w.snapshot(t)
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// D-22: handle profiles live as long as the match index: kept by a Propose,
// shared by the next one, and dropped by a handle commit.
func TestProfilesLiveWithTheMatchIndex(t *testing.T) {
	w := newWarmFixture(t)
	w.f.c.Graph().Drop()
	if _, ok := promotealign.ProfilesKeptForTest(w.f.c); ok {
		t.Fatal("a dropped graph holds a match index")
	}
	w.f.propose(w.f.Graph(w.srcB), nil)
	kept, ok := promotealign.ProfilesKeptForTest(w.f.c)
	if !ok || kept == 0 {
		t.Fatalf("after Propose the index holds %d profiles (index %v)", kept, ok)
	}
	w.f.propose(w.f.Graph(w.srcB), nil)
	if again, _ := promotealign.ProfilesKeptForTest(w.f.c); again != kept {
		t.Fatalf("a second Propose changed the kept profiles from %d to %d", kept, again)
	}
	w.f.join(w.f.Person(w.f.source, w.f.artifact, "John Smith"), w.personHandle)
	if _, ok := promotealign.ProfilesKeptForTest(w.f.c); ok {
		t.Fatal("a handle commit left the match index and its profiles")
	}
}

// S11-21 D-8 (AC-4.1, AC-4.2, AC-4.3): a Subject's remembered candidates never
// outlive the index or the Source edit they were ranked against: the next
// proposal equals one on a dropped graph.
func TestCandidatesMemoFollowsCommits(t *testing.T) {
	tests := []struct {
		name   string
		change func(w *warmFixture) (newHandle []byte)
	}{
		{"no commit", func(*warmFixture) []byte { return nil }},
		{"a handle commit", func(w *warmFixture) []byte {
			srcC, artC := w.f.NewSource("Parish register")
			_, _, ph, _ := w.f.filedBirth(srcC, artC, "John Smith", 1817, nil, nil)
			return ph
		}},
		{"a vocabulary commit", func(w *warmFixture) []byte {
			if err := w.f.c.Graph().OnCommit(effects.Commit{
				Activity: effects.Activity{Revision: 1 << 40}, Effects: effects.Set{Lists: []effects.List{effects.ListPropertyVocabulary}},
			}); err != nil {
				t.Fatal(err)
			}
			return nil
		}},
		{"a name format change", func(w *warmFixture) []byte {
			if err := nameformats.Invalidate(w.f.c); err != nil {
				t.Fatal(err)
			}
			return nil
		}},
		{"a Source edit", func(w *warmFixture) []byte {
			w.f.Cite(w.artB, w.bPerson, observations.Input{
				PropertyID: w.f.Prop("name").ID, Name: namevaluetest.Western("Johnny Smith"),
			})
			return nil
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newWarmFixture(t)
			graph := w.f.Graph(w.srcB)
			w.f.propose(graph, nil)
			newHandle := tc.change(w)
			warm := w.f.propose(graph, nil)
			w.f.c.Graph().Drop()
			fresh := w.f.propose(graph, nil)
			if !reflect.DeepEqual(warm, fresh) {
				t.Fatalf("the proposal after the change differs from a fresh one:\nwarm  %+v\nfresh %+v", warm, fresh)
			}
			if newHandle == nil {
				return
			}
			row := rowFor(fresh, w.bPerson.ID)
			found := row.Target == graphalign.TargetHandle && bytes.Equal(row.HandleID, newHandle)
			for _, a := range row.Alternatives {
				found = found || bytes.Equal(a.HandleID, newHandle)
			}
			if !found {
				t.Errorf("the new handle %x is neither the row's target nor an alternative: %+v", newHandle, row)
			}
		})
	}
}

// T-2 (AC-1.2): Warm loads the shortlist kinds and builds the index, so a
// proposal on another Source builds no new one and loads no kind.
func TestWarmBuildsTheIndex(t *testing.T) {
	w := newWarmFixture(t)
	c := w.f.c
	g := c.Graph()
	g.Drop()
	if err := promotealign.Warm(c); err != nil {
		t.Fatal(err)
	}
	gen := promotealign.MatchIndexGenForTest(g)
	if gen == 0 {
		t.Fatal("Warm built no index")
	}
	graphcachetest.AssertKindsLoaded(t, c, "person", "event")
	w.f.propose(w.f.Graph(w.srcB), nil)
	if got := promotealign.MatchIndexGenForTest(g); got != gen {
		t.Errorf("a proposal rebuilt the index: generation %d, want %d", got, gen)
	}
}

// The match index drops as it is declared: a Source-only commit keeps it; a
// handle commit, a vocabulary commit, a new alias index and Drop empty it.
func TestMatchIndexDrops(t *testing.T) {
	w := newWarmFixture(t)
	c := w.f.c
	g := c.Graph()
	warm := func() {
		t.Helper()
		g.Drop()
		if err := promotealign.Warm(c); err != nil {
			t.Fatal(err)
		}
		if !stored(g) {
			t.Fatal("Warm stored no match index")
		}
	}
	rev := int64(1000)
	commit := func(set effects.Set) {
		t.Helper()
		rev++
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: rev}, Effects: set}); err != nil {
			t.Fatal(err)
		}
	}
	warm()
	commit(effects.Set{SourceScopes: [][]byte{w.f.source.ID}})
	if !stored(g) {
		t.Error("a Source-only commit dropped the match index")
	}
	commit(effects.Set{Handles: [][]byte{w.personHandle}})
	if stored(g) {
		t.Error("a handle commit left the match index")
	}
	warm()
	commit(effects.Set{Lists: []effects.List{effects.ListPropertyVocabulary}})
	if stored(g) {
		t.Error("a vocabulary commit left the match index")
	}
	warm()
	commit(effects.Set{Lists: []effects.List{effects.ListSourceVocabulary}})
	if !stored(g) {
		t.Error("a Source vocabulary commit dropped the match index")
	}
	namealiases.Index.Set(g, valuecompare.AliasIndex{})
	if stored(g) {
		t.Error("a new alias index left the match index")
	}
	warm()
	g.Drop()
	if stored(g) {
		t.Error("Drop left the match index")
	}
}

func stored(g *graphcache.Graph) bool {
	_, ok, _ := promotealign.MatchIndex.Get(g)
	return ok
}

// With PROVENENCIA_GRAPH_VERIFY=1 a commit rebuilds the stored shortlist index
// and profiles and compares them.
func TestVerifyChecksMatchIndex(t *testing.T) {
	t.Setenv("PROVENENCIA_GRAPH_VERIFY", "1")
	w := newWarmFixture(t)
	c := w.f.c
	g := c.Graph()
	w.f.propose(w.f.Graph(w.srcB), nil)
	if !stored(g) {
		t.Fatal("setup: no match index")
	}
	if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 2000}, Effects: effects.Set{SourceScopes: [][]byte{w.f.source.ID}}}); err != nil {
		t.Fatalf("a fresh match index failed Verify: %v", err)
	}
	if err := promotealign.EmptyShortlistForTest(c); err != nil {
		t.Fatal(err)
	}
	err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 2001}, Effects: effects.Set{SourceScopes: [][]byte{w.f.source.ID}}})
	if err == nil || !strings.Contains(err.Error(), "promotealign.matchindex") {
		t.Fatalf("a shortlist lacking a handle: error = %v, want one naming promotealign.matchindex", err)
	}
}
