package promote_test

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/autoreconciler/autoreconcilertest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectpositions"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/datevalue"
	"github.com/mendahu/provenencia/core/namevalue/namevaluetest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

const batchLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type batchFixture struct {
	t        *testing.T
	c        *database.Catalog
	source   sources.Source
	artifact artifacts.Artifact
	graph    []byte // the Source's one Evidence graph
	y        int64
}

func newBatchFixture(t *testing.T) *batchFixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	if err := users.Upsert(c, userID, "Tester", r); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	return &batchFixture{t: t, c: c, source: src, artifact: art, graph: evidencegraphstest.SoleGraph(t, c, src.ID).ID}
}

func (f *batchFixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, vocabkey.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

func (f *batchFixture) term(propKey, termKey string) propertyterms.Term {
	f.t.Helper()
	term, err := propertyterms.Lookup(f.c, f.prop(propKey).ID, termKey, vocabkey.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return term
}

func (f *batchFixture) bare(kind string) subjects.Subject {
	f.t.Helper()
	s := subjectstest.On(f.t, f.c, userID, f.graph, kind, "")
	if _, err := writes.Call(f.c, writes.Op{}, func(tx *database.Tx) (subjectpositions.Position, []rowchange.Change, error) {
		return subjectpositions.Set(tx, s.ID, 0, f.y)
	}); err != nil {
		f.t.Fatal(err)
	}
	f.y += 2
	return s
}

func (f *batchFixture) cite(s subjects.Subject, in ...observations.Input) []observations.Observation {
	f.t.Helper()
	for i := range in {
		in[i].SubjectID = s.ID
	}
	res, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{
			ArtifactID: f.artifact.ID, LocatorJSON: batchLocator,
		}, in)
	})
	if err != nil {
		f.t.Fatal(err)
	}
	return res.Observations
}

func (f *batchFixture) participation(person, event subjects.Subject) subjects.Subject {
	f.t.Helper()
	personProp := f.prop("person").ID
	eventProp := f.prop("event").ID
	roleProp := f.prop("role").ID
	role := f.term("role", "subject").ID
	res, err := writes.Call(f.c, writes.Op{Action: "create_cited_bridge", UserID: userID}, func(tx *database.Tx) (connect.Result, []rowchange.Change, error) {
		return connect.CreateCitedBridge(tx, userID, connect.CreateInput{
			FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
			Citation: citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: batchLocator},
			Observations: []observations.Input{
				{PropertyID: personProp, ValueSubjectID: person.ID},
				{PropertyID: eventProp, ValueSubjectID: event.ID},
				{PropertyID: roleProp, ValueTermID: role},
			},
		})
	})
	if err != nil {
		f.t.Fatal(err)
	}
	return res.Subject
}

func (f *batchFixture) revision() int64 {
	f.t.Helper()
	db, err := f.c.DB()
	if err != nil {
		f.t.Fatal(err)
	}
	var rev int64
	if err := db.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev); err != nil {
		f.t.Fatal(err)
	}
	return rev
}

func (f *batchFixture) save(in promote.Batch) promote.BatchResult {
	f.t.Helper()
	res, err := writes.Call(f.c, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (promote.BatchResult, []rowchange.Change, error) {
		return promote.SaveBatch(tx, userID, in)
	})
	if err != nil {
		f.t.Fatal(err)
	}
	return res
}

func (f *batchFixture) bridgeClaims(kind string) int {
	f.t.Helper()
	db, err := f.c.DB()
	if err != nil {
		f.t.Fatal(err)
	}
	var n int
	err = db.QueryRow(`SELECT COUNT(*) FROM identity_claims ic
		JOIN subjects s ON s.id = ic.subject_id AND s.source_id = ?
		JOIN subject_types st ON st.id = s.subject_type_id AND st.key = ? AND st.origin = 'provenencia'
		WHERE ic.status = 'accepted'`, f.source.ID, kind).Scan(&n)
	if err != nil {
		f.t.Fatal(err)
	}
	return n
}

func (f *batchFixture) hasClaim(subjectID []byte) bool {
	f.t.Helper()
	_, err := identityclaims.AcceptedEntityForSubject(f.c, subjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		f.t.Fatal(err)
	}
	return true
}

func yearDate(y int) *datevalue.Value {
	return &datevalue.Value{Kind: datevalue.KindPoint, Calendar: "gregorian", StartYear: &y}
}

func TestSaveBatchFilesBothEndsOnce(t *testing.T) {
	f := newBatchFixture(t)
	person := f.bare("person")
	event := f.bare("event")
	f.participation(person, event)

	res := f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: person.ID, Target: promote.TargetNew},
			{SubjectID: event.ID, Target: promote.TargetNew},
		},
	})
	if len(res.Written) != 2 {
		t.Fatalf("written %d", len(res.Written))
	}
	if f.bridgeClaims("participation") != 1 {
		t.Fatalf("participation claims %d, want 1", f.bridgeClaims("participation"))
	}
	db, err := f.c.DB()
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
	autoreconcilertest.AssertCacheMatchesRebuild(t, f.c)
}

func TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted(t *testing.T) {
	f := newBatchFixture(t)
	person := f.bare("person")
	event := f.bare("event")
	per := promotetest.Save(t, f.c, userID, promote.Input{SubjectID: person.ID})
	evt := promotetest.Save(t, f.c, userID, promote.Input{SubjectID: event.ID})
	f.participation(person, event)
	if f.bridgeClaims("participation") != 0 {
		t.Fatal("bridge filed before the batch")
	}

	f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: person.ID, Target: promote.TargetHandle, EntityID: per.Entity.ID},
			{SubjectID: event.ID, Target: promote.TargetHandle, EntityID: evt.Entity.ID},
		},
	})
	if f.bridgeClaims("participation") != 1 {
		t.Fatalf("participation claims %d, want 1", f.bridgeClaims("participation"))
	}
}

func TestSaveBatchLeavesBridgeWithASkippedEndUnfiled(t *testing.T) {
	f := newBatchFixture(t)
	person := f.bare("person")
	event := f.bare("event")
	f.participation(person, event)
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
	person := f.bare("person")
	seen := f.revision()
	_ = f.bare("place") // a later write advances the revision
	_, err := writes.Call(f.c, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (promote.BatchResult, []rowchange.Change, error) {
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
	b := f.bare("person")
	f.cite(b, observations.Input{PropertyID: f.prop("name").ID, Name: namevaluetest.Western("Mary Smith")})
	e1 := promotetest.Save(t, f.c, userID, promote.Input{SubjectID: b.ID})
	a := f.bare("person")
	f.cite(a, observations.Input{PropertyID: f.prop("name").ID, Name: namevaluetest.Western("Mary Smith")})
	c := f.bare("event")
	d := f.bare("place")
	f.participation(a, c)
	memberBefore, err := identityclaims.Get(f.c, e1.Claim.ID)
	if err != nil {
		t.Fatal(err)
	}

	res := f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: a.ID, Target: promote.TargetHandle, EntityID: e1.Entity.ID},
			{SubjectID: c.ID, Target: promote.TargetNew},
			{SubjectID: d.ID, Target: promote.TargetSkip},
		},
	})
	if len(res.Written) != 2 {
		t.Fatalf("written %d, want 2", len(res.Written))
	}
	for _, w := range res.Written {
		if w.Claim.Status != identityclaims.StatusAccepted {
			t.Fatalf("claim status %s", w.Claim.Status)
		}
	}
	if n := f.bridgeClaims("participation"); n != 1 {
		t.Fatalf("participation claims %d, want 1", n)
	}
	memberAfter, err := identityclaims.Get(f.c, e1.Claim.ID)
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
	a := f.bare("person")
	other := f.bare("person")
	e1 := promotetest.Save(t, f.c, userID, promote.Input{SubjectID: a.ID})
	e2 := promotetest.Save(t, f.c, userID, promote.Input{SubjectID: other.ID})

	res := f.save(promote.Batch{
		GraphID: f.graph, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{{SubjectID: a.ID, Target: promote.TargetHandle, EntityID: e1.Entity.ID}},
	})
	if len(res.Written) != 0 {
		t.Fatalf("an anchor wrote %+v", res.Written)
	}

	try := func(row promote.BatchRow) error {
		seen := f.revision()
		_, err := writes.Call(f.c, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (promote.BatchResult, []rowchange.Change, error) {
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
	person := f.bare("person")
	event := f.bare("event")
	promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: person.ID})
	promotetest.Save(f.t, f.c, userID, promote.Input{SubjectID: event.ID})
	f.participation(person, event)
	if n := f.bridgeClaims("participation"); n != 0 {
		t.Fatalf("participation claims %d before any Done, want 0", n)
	}
	def, jones, _ := artifactstest.Split(t, f.c, userID, f.source.ID, "Jones")
	if !bytes.Equal(def.ID, f.graph) {
		t.Fatalf("Default graph %x, want %x", def.ID, f.graph)
	}
	j := subjectstest.On(t, f.c, userID, jones.ID, "person", "")

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
		_, err := writes.Call(f.c, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (promote.BatchResult, []rowchange.Change, error) {
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
		smith := f.bare("person")
		_, jones, _ := artifactstest.Split(t, f.c, userID, f.source.ID, "Jones")
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
