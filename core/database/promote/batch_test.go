package promote_test

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/autoreconciler/autoreconcilertest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/writes"
)

type batchFixture struct {
	*evidencelayertest.Records
	graph []byte // the Source's one Evidence graph
}

func newBatchFixture(t *testing.T) *batchFixture {
	t.Helper()
	r := evidencelayertest.New(t)
	return &batchFixture{Records: r, graph: r.Graph(r.Source)}
}

func (f *batchFixture) revision() int64 {
	f.T.Helper()
	db, err := f.C.DB()
	if err != nil {
		f.T.Fatal(err)
	}
	var rev int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		f.T.Fatal(err)
	}
	return rev
}

func (f *batchFixture) save(in promote.Batch) {
	f.T.Helper()
	_, err := writes.Call(f.C, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		return promote.SaveBatch(tx, userID, in)
	})
	if err != nil {
		f.T.Fatal(err)
	}
}

func (f *batchFixture) bridgeClaims(kind string) int {
	f.T.Helper()
	db, err := f.C.DB()
	if err != nil {
		f.T.Fatal(err)
	}
	var n int
	err = db.QueryRow(`SELECT COUNT(*) FROM identity_claims ic
		JOIN subjects s ON s.id = ic.subject_id AND s.source_id = ?
		JOIN subject_types st ON st.id = s.subject_type_id AND st.key = ? AND st.origin = 'provenencia'
		WHERE ic.status = 'accepted'`, f.Source.ID, kind).Scan(&n)
	if err != nil {
		f.T.Fatal(err)
	}
	return n
}

func (f *batchFixture) hasClaim(subjectID []byte) bool {
	f.T.Helper()
	_, err := identityclaims.AcceptedEntityForSubject(f.C, subjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		f.T.Fatal(err)
	}
	return true
}

func TestSaveBatchFilesBothEndsOnce(t *testing.T) {
	f := newBatchFixture(t)
	person := f.Bare(f.Source, "person")
	event := f.Bare(f.Source, "event")
	f.Participation(f.Source, f.Artifact, person, event, "subject")

	f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: person.ID, Target: promote.TargetNew},
			{SubjectID: event.ID, Target: promote.TargetNew},
		},
	})
	if !f.hasClaim(person.ID) || !f.hasClaim(event.ID) {
		t.Fatal("both ends were not filed")
	}
	if f.bridgeClaims("participation") != 1 {
		t.Fatalf("participation claims %d, want 1", f.bridgeClaims("participation"))
	}
	db, err := f.C.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions WHERE action_type = 'promote_batch'`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("promote_batch revisions %d", n)
	}
	autoreconcilertest.AssertCacheMatchesRebuild(t, f.C)
}

func TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted(t *testing.T) {
	f := newBatchFixture(t)
	person := f.Bare(f.Source, "person")
	event := f.Bare(f.Source, "event")
	f.Participation(f.Source, f.Artifact, person, event, "subject")
	per := promotetest.ClaimWithoutFiling(t, f.C, userID, person)
	evt := promotetest.ClaimWithoutFiling(t, f.C, userID, event)
	if f.bridgeClaims("participation") != 0 {
		t.Fatal("bridge filed before the batch")
	}

	f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: person.ID, Target: promote.TargetHandle, EntityID: per},
			{SubjectID: event.ID, Target: promote.TargetHandle, EntityID: evt},
		},
	})
	if f.bridgeClaims("participation") != 1 {
		t.Fatalf("participation claims %d, want 1", f.bridgeClaims("participation"))
	}
}

func TestSaveBatchLeavesBridgeWithASkippedEndUnfiled(t *testing.T) {
	f := newBatchFixture(t)
	person := f.Bare(f.Source, "person")
	event := f.Bare(f.Source, "event")
	f.Participation(f.Source, f.Artifact, person, event, "subject")
	f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: person.ID, Target: promote.TargetNew},
			{SubjectID: event.ID, Target: promote.TargetSkip},
		},
	})
	if n := f.bridgeClaims("participation"); n != 0 {
		t.Fatalf("bridge with a skipped end filed %d claims", n)
	}
	if !f.hasClaim(person.ID) {
		t.Fatal("person was not filed")
	}
}

func TestSaveBatchStaleWritesNothing(t *testing.T) {
	f := newBatchFixture(t)
	person := f.Bare(f.Source, "person")
	seen := f.revision()
	_ = f.Bare(f.Source, "place") // a later write advances the revision
	_, err := writes.Call(f.C, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		return promote.SaveBatch(tx, userID, promote.Batch{
			GraphID: f.graph, SeenRevision: seen,
			Rows: []promote.BatchRow{{SubjectID: person.ID, Target: promote.TargetNew}},
		})
	})
	if !errors.Is(err, promote.ErrStale) {
		t.Fatalf("err %v, want promote.ErrStale", err)
	}
	if f.hasClaim(person.ID) {
		t.Fatal("stale batch filed a claim")
	}
}

// A Done files the chosen rows' claims and their connections and nothing else:
// no other member's claim changes.
func TestSaveBatchFilesClaimsAndBridgesOnly(t *testing.T) {
	f := newBatchFixture(t)
	b := f.Bare(f.Source, "person")
	f.Cite(f.Artifact, b, observations.Input{PropertyID: f.Prop("name").ID, Name: namevaluetest.Western("Mary Smith")})
	e1 := promotetest.Save(t, f.C, userID, promote.Input{SubjectID: b.ID})
	a := f.Bare(f.Source, "person")
	f.Cite(f.Artifact, a, observations.Input{PropertyID: f.Prop("name").ID, Name: namevaluetest.Western("Mary Smith")})
	c := f.Bare(f.Source, "event")
	d := f.Bare(f.Source, "place")
	f.Participation(f.Source, f.Artifact, a, c, "subject")
	memberBefore, err := identityclaims.Get(f.C, e1.Claim.ID)
	if err != nil {
		t.Fatal(err)
	}

	f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: a.ID, Target: promote.TargetHandle, EntityID: e1.Entity.ID},
			{SubjectID: c.ID, Target: promote.TargetNew},
			{SubjectID: d.ID, Target: promote.TargetSkip},
		},
	})
	joined, err := identityclaims.AcceptedEntityForSubject(f.C, a.ID)
	if err != nil || !bytes.Equal(joined.EntityID, e1.Entity.ID) {
		t.Fatalf("a joined %v on %x, want %x", err, joined.EntityID, e1.Entity.ID)
	}
	if !f.hasClaim(c.ID) {
		t.Fatal("c was not filed")
	}
	if f.hasClaim(d.ID) {
		t.Fatal("the skipped row was filed")
	}
	if n := f.bridgeClaims("participation"); n != 1 {
		t.Fatalf("participation claims %d, want 1", n)
	}
	memberAfter, err := identityclaims.Get(f.C, e1.Claim.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(memberBefore, memberAfter) {
		t.Fatalf("the member's claim changed:\nbefore %+v\nafter  %+v", memberBefore, memberAfter)
	}
}

// A row on the handle its Subject already belongs to is an anchor and writes
// nothing; on any other handle it is refused. A Skip carries no record.
func TestSaveBatchAnchorsARowOnItsOwnHandle(t *testing.T) {
	f := newBatchFixture(t)
	a := f.Bare(f.Source, "person")
	other := f.Bare(f.Source, "person")
	e1 := promotetest.Save(t, f.C, userID, promote.Input{SubjectID: a.ID})
	e2 := promotetest.Save(t, f.C, userID, promote.Input{SubjectID: other.ID})

	before := f.revision()
	f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: before,
		Rows: []promote.BatchRow{{SubjectID: a.ID, Target: promote.TargetHandle, EntityID: e1.Entity.ID}},
	})
	if got := f.revision(); got != before {
		t.Fatalf("an anchor wrote: revision %d, want %d", got, before)
	}

	try := func(row promote.BatchRow) error {
		seen := f.revision()
		_, err := writes.Call(f.C, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			return promote.SaveBatch(tx, userID, promote.Batch{GraphID: f.graph, SeenRevision: seen, Rows: []promote.BatchRow{row}})
		})
		return err
	}
	if err := try(promote.BatchRow{SubjectID: a.ID, Target: promote.TargetHandle, EntityID: e2.Entity.ID}); !errors.Is(err, identityclaims.ErrAlreadyMember) {
		t.Fatalf("another handle: err %v, want identityclaims.ErrAlreadyMember", err)
	}
	if err := try(promote.BatchRow{SubjectID: a.ID, Target: promote.TargetSkip, EntityID: e1.Entity.ID}); !errors.Is(err, promote.ErrInvalid) {
		t.Fatalf("skip with a record: err %v, want promote.ErrInvalid", err)
	}
}

// A Done files the Subjects of its own graph and the bridges between them; a
// waiting bridge on another graph of the Source stays unfiled until that
// graph's own Done.
func TestSaveBatchFilesOnlyItsGraph(t *testing.T) {
	f := newBatchFixture(t)
	person := f.Bare(f.Source, "person")
	event := f.Bare(f.Source, "event")
	f.Participation(f.Source, f.Artifact, person, event, "subject")
	promotetest.ClaimWithoutFiling(f.T, f.C, userID, person)
	promotetest.ClaimWithoutFiling(f.T, f.C, userID, event)
	if n := f.bridgeClaims("participation"); n != 0 {
		t.Fatalf("participation claims %d before any Done, want 0", n)
	}
	def, jones, _ := artifactstest.Split(t, f.C, userID, f.Source.ID, "Jones")
	if !bytes.Equal(def.ID, f.graph) {
		t.Fatalf("Default graph %x, want %x", def.ID, f.graph)
	}
	j := subjectstest.On(t, f.C, userID, jones.ID, "person", "")

	f.save(promote.Batch{
		GraphID: jones.ID, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{{SubjectID: j.ID, Target: promote.TargetNew}},
	})
	if !f.hasClaim(j.ID) {
		t.Fatal("the Jones person was not filed")
	}
	if n := f.bridgeClaims("participation"); n != 0 {
		t.Fatalf("participation claims %d after the Jones Done, want 0", n)
	}

	f.save(promote.Batch{GraphID: f.graph, SeenRevision: f.revision()})
	if n := f.bridgeClaims("participation"); n != 1 {
		t.Fatalf("participation claims %d after the Default Done, want 1", n)
	}

	refused := func(t *testing.T, f *batchFixture, in promote.Batch) {
		t.Helper()
		before := f.revision()
		_, err := writes.Call(f.C, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			return promote.SaveBatch(tx, userID, in)
		})
		if !errors.Is(err, promote.ErrInvalid) {
			t.Fatalf("err %v, want promote.ErrInvalid", err)
		}
		if got := f.revision(); got != before {
			t.Fatalf("revision %d, want %d: nothing is written", got, before)
		}
	}

	t.Run("a row on another graph is invalid", func(t *testing.T) {
		f := newBatchFixture(t)
		smith := f.Bare(f.Source, "person")
		_, jones, _ := artifactstest.Split(t, f.C, userID, f.Source.ID, "Jones")
		refused(t, f, promote.Batch{
			GraphID: jones.ID, SeenRevision: f.revision(),
			Rows: []promote.BatchRow{{SubjectID: smith.ID, Target: promote.TargetNew}},
		})
		if f.hasClaim(smith.ID) {
			t.Fatal("the Default person was filed")
		}
	})

	t.Run("a graph that does not exist is invalid", func(t *testing.T) {
		f := newBatchFixture(t)
		refused(t, f, promote.Batch{GraphID: make([]byte, 16), SeenRevision: f.revision()})
		refused(t, f, promote.Batch{GraphID: []byte{1, 2, 3}, SeenRevision: f.revision()})
	})
}

// A row naming an association Subject is unsupported even once its bridge is
// filed: the type is checked before membership.
func TestSaveBatchRefusesAnAssociationRowAlreadyFiled(t *testing.T) {
	f := newBatchFixture(t)
	person := f.Bare(f.Source, "person")
	event := f.Bare(f.Source, "event")
	bridge := f.Participation(f.Source, f.Artifact, person, event, "subject")
	promotetest.Save(t, f.C, userID, promote.Input{SubjectID: person.ID})
	promotetest.Save(t, f.C, userID, promote.Input{SubjectID: event.ID})
	entity, err := identityclaims.AcceptedEntityForSubject(f.C, bridge.ID)
	if err != nil {
		t.Fatal(err)
	}
	seen := f.revision()
	_, err = writes.Call(f.C, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		return promote.SaveBatch(tx, userID, promote.Batch{
			GraphID: f.graph, SeenRevision: seen,
			Rows: []promote.BatchRow{{SubjectID: bridge.ID, Target: promote.TargetHandle, EntityID: entity.EntityID}},
		})
	})
	if !errors.Is(err, promote.ErrUnsupportedType) {
		t.Fatalf("err %v, want promote.ErrUnsupportedType", err)
	}
}

// A bridge filed when it was drawn stays filed, once, through a Done.
func TestSaveBatchKeepsABridgeFiledOnCreate(t *testing.T) {
	f := newBatchFixture(t)
	person := f.Bare(f.Source, "person")
	event := f.Bare(f.Source, "event")
	promotetest.Save(t, f.C, userID, promote.Input{SubjectID: person.ID})
	promotetest.Save(t, f.C, userID, promote.Input{SubjectID: event.ID})
	f.Participation(f.Source, f.Artifact, person, event, "subject")
	if n := f.bridgeClaims("participation"); n != 1 {
		t.Fatalf("participation claims %d after the connect, want 1", n)
	}

	f.save(promote.Batch{GraphID: f.graph, SeenRevision: f.revision()})
	if n := f.bridgeClaims("participation"); n != 1 {
		t.Fatalf("participation claims %d after the Done, want 1", n)
	}
}
