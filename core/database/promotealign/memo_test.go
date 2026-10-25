package promotealign_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promotealign"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/graphalign"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
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
	srcB, artB := f.newSource("Obituary")
	bPerson := f.person(srcB, artB, "John Smith")
	bBirth := f.datedBirth(srcB, artB, 1817)
	f.participation(srcB, artB, bPerson, bBirth, "subject")
	return &warmFixture{f: f, srcB: srcB, artB: artB, member: member, bPerson: bPerson, bBirth: bBirth, personHandle: ph, birthHandle: bh}
}

func (w *warmFixture) request() promotealign.EvidenceRequest {
	return promotealign.EvidenceRequest{
		GraphID: w.f.graph(w.srcB), SubjectID: w.bPerson.ID, HandleID: w.personHandle,
		Neighbors: []graphalign.Fixed{{SubjectID: w.bBirth.ID, HandleID: w.birthHandle}},
	}
}

// T-18 (AC-4.1): a second Evidence for one row, with no commit between, issues
// no query through the graph.
func TestEvidenceWarmIssuesNoQuery(t *testing.T) {
	w := newWarmFixture(t)
	cold := w.f.evidence(w.request())
	if len(cold.Relationships) != 1 {
		t.Fatalf("relationships = %d, want 1", len(cold.Relationships))
	}
	n := 0
	w.f.c.Graph().SetQueryCounterForTest(func() { n++ })
	warm := w.f.evidence(w.request())
	if n != 0 {
		t.Fatalf("the second Evidence issued %d graph queries", n)
	}
	if !reflect.DeepEqual(cold, warm) {
		t.Fatalf("warm Evidence differs from cold:\ncold %+v\nwarm %+v", cold, warm)
	}
}

// T-18 (AC-4.2): a second Propose on one Source, with no commit between,
// issues no query through the graph. The first built the shortlist index.
func TestProposeWarmIssuesNoQuery(t *testing.T) {
	w := newWarmFixture(t)
	cold := w.f.propose(w.f.graph(w.srcB), nil)
	n := 0
	w.f.c.Graph().SetQueryCounterForTest(func() { n++ })
	warm := w.f.propose(w.f.graph(w.srcB), nil)
	if n != 0 {
		t.Fatalf("the second Propose issued %d graph queries", n)
	}
	if !reflect.DeepEqual(cold, warm) {
		t.Fatalf("warm Propose differs from cold:\ncold %+v\nwarm %+v", cold, warm)
	}
}

// T-19 (AC-4.3): after a commit that changes a member, a title or a handle,
// Evidence and Propose follow it: the warm output is a fresh read's.
func TestMemoFollowsCommit(t *testing.T) {
	// compare warms the caches, applies the commit, and checks that the
	// output changed, and equals what a cold graph reads.
	compare := func(t *testing.T, w *warmFixture, commit func()) {
		t.Helper()
		beforeE, errE := promotealign.Evidence(w.f.c, w.request())
		beforeP := w.f.propose(w.f.graph(w.srcB), nil)
		if errE != nil {
			t.Fatal(errE)
		}
		commit()
		warmE, warmErr := promotealign.Evidence(w.f.c, w.request())
		warmP, _, warmPErr := promotealign.Propose(w.f.c, w.f.graph(w.srcB), nil)
		w.f.c.Graph().Drop()
		freshE, freshErr := promotealign.Evidence(w.f.c, w.request())
		freshP, _, freshPErr := promotealign.Propose(w.f.c, w.f.graph(w.srcB), nil)
		if !sameErr(warmErr, freshErr) || !sameErr(warmPErr, freshPErr) {
			t.Fatalf("errors differ: warm %v / %v, fresh %v / %v", warmErr, warmPErr, freshErr, freshPErr)
		}
		if !reflect.DeepEqual(warmE, freshE) {
			t.Errorf("warm Evidence follows the commit wrongly:\nwarm  %+v\nfresh %+v", warmE, freshE)
		}
		if !reflect.DeepEqual(warmP, freshP) {
			t.Errorf("warm Propose follows the commit wrongly:\nwarm  %+v\nfresh %+v", warmP, freshP)
		}
		if reflect.DeepEqual(beforeE, freshE) && reflect.DeepEqual(beforeP, freshP) && freshErr == nil {
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
				if l.PropertyID == nil || string(l.PropertyID) != string(w.f.prop("name").ID) {
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

	t.Run("a Source retitled", func(t *testing.T) {
		w := newWarmFixture(t)
		compare(t, w, func() {
			src := w.f.source
			src.Title = "Parish register"
			_, err := writes.Call(w.f.c, writes.Op{Action: "update_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
				ch, err := sources.Update(tx, userID, src)
				return src, ch, err
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("a handle merged", func(t *testing.T) {
		w := newWarmFixture(t)
		compare(t, w, func() {
			other := w.f.person(w.f.source, w.f.artifact, "Somebody Else")
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
// reader: Propose and Evidence leave them as they were stored.
func TestMemoIsNotMutated(t *testing.T) {
	w := newWarmFixture(t)
	w.f.evidence(w.request())
	w.f.propose(w.f.graph(w.srcB), nil)
	w.f.evidence(promotealign.EvidenceRequest{GraphID: w.f.graph(w.srcB), SubjectID: w.bPerson.ID, HandleID: w.personHandle})
	first := w.snapshot(t)
	if !first.LayerOK || len(first.Layer.Subjects) == 0 || len(first.Members) == 0 {
		t.Fatalf("nothing was memoized: %+v", first)
	}
	before := memoJSON(t, w)
	for i := 0; i < 2; i++ {
		w.f.propose(w.f.graph(w.srcB), []graphalign.Fixed{{SubjectID: w.bPerson.ID, HandleID: w.personHandle}})
		w.f.evidence(w.request())
		if after := memoJSON(t, w); after != before {
			t.Fatalf("a read changed the memo:\nbefore %s\nafter  %s", before, after)
		}
		// The same layer, not a rebuilt copy.
		again := w.snapshot(t)
		if &again.Layer.Subjects[0] != &first.Layer.Subjects[0] {
			t.Fatal("the layer was rebuilt, not read from the memo")
		}
	}
}

// snapshot is the memoized layer of the incoming Source's graph and the member
// values of the fixture Source's graph.
func (w *warmFixture) snapshot(t *testing.T) promotealign.MemoSnapshot {
	t.Helper()
	return promotealign.MemoSnapshotForTest(w.f.c,
		w.f.graph(w.srcB), w.f.graph(w.f.source))
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

// D-22: handle profiles live as long as the match index: kept by an Evidence,
// shared by the next one, and dropped by a handle commit.
func TestProfilesLiveWithTheMatchIndex(t *testing.T) {
	w := newWarmFixture(t)
	w.f.c.Graph().Drop()
	if _, ok := promotealign.ProfilesKeptForTest(w.f.c); ok {
		t.Fatal("a dropped graph holds a match index")
	}
	w.f.evidence(w.request())
	kept, ok := promotealign.ProfilesKeptForTest(w.f.c)
	if !ok || kept == 0 {
		t.Fatalf("after Evidence the index holds %d profiles (index %v)", kept, ok)
	}
	w.f.evidence(w.request())
	if again, _ := promotealign.ProfilesKeptForTest(w.f.c); again != kept {
		t.Fatalf("a second Evidence changed the kept profiles from %d to %d", kept, again)
	}
	w.f.join(w.f.person(w.f.source, w.f.artifact, "John Smith"), w.personHandle)
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
			srcC, artC := w.f.newSource("Parish register")
			_, _, ph, _ := w.f.filedBirth(srcC, artC, "John Smith", 1817, nil, nil)
			return ph
		}},
		{"a Source edit", func(w *warmFixture) []byte {
			w.f.cite(w.artB, w.bPerson, observations.Input{
				PropertyID: w.f.prop("name").ID, Name: namevaluetest.Western("Johnny Smith"),
			})
			return nil
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newWarmFixture(t)
			graph := w.f.graph(w.srcB)
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
