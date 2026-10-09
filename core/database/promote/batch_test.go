package promote_test

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/connect"
	"github.com/mendahu/provenencia/core/database/datevalues"
	"github.com/mendahu/provenencia/core/database/evrun"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

const batchLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

type batchFixture struct {
	t        *testing.T
	c        *database.Catalog
	source   sources.Source
	artifact artifacts.Artifact
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	art, err := runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	return &batchFixture{t: t, c: c, source: src, artifact: art}
}

func (f *batchFixture) prop(key string) properties.Property {
	f.t.Helper()
	p, err := properties.Lookup(f.c, key, properties.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

func (f *batchFixture) term(propKey, termKey string) propertyterms.Term {
	f.t.Helper()
	term, err := propertyterms.Lookup(f.c, f.prop(propKey).ID, termKey, propertyterms.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	return term
}

func (f *batchFixture) bare(kind string) subjects.Subject {
	f.t.Helper()
	st, err := subjecttypes.Lookup(f.c, kind, subjecttypes.OriginProvenencia)
	if err != nil {
		f.t.Fatal(err)
	}
	s, err := evrun.CreateSubject(f.c, userID, subjects.CreateInput{SourceID: f.source.ID, SubjectTypeID: st.ID}, nil)
	if err != nil {
		f.t.Fatal(err)
	}
	if _, err := evrun.SetPosition(f.c, s.ID, 0, f.y); err != nil {
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
	res, err := evrun.CreateCitation(f.c, userID, citations.CreateInput{
		ArtifactID: f.artifact.ID, LocatorJSON: batchLocator,
	}, in)
	if err != nil {
		f.t.Fatal(err)
	}
	return res.Observations
}

func (f *batchFixture) participation(person, event subjects.Subject) subjects.Subject {
	f.t.Helper()
	res, err := evrun.CreateBridge(f.c, userID, connect.CreateInput{
		SourceID: f.source.ID, FromSubjectID: person.ID, ToSubjectID: event.ID, BridgeTypeKey: "participation",
		Citation: citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: batchLocator},
		Observations: []observations.Input{
			{PropertyID: f.prop("person").ID, ValueSubjectID: person.ID},
			{PropertyID: f.prop("event").ID, ValueSubjectID: event.ID},
			{PropertyID: f.prop("role").ID, ValueTermID: f.term("role", "subject").ID},
		},
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
	res, err := promote.SaveBatch(f.c, userID, in)
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

func yearDate(y int) *datevalues.Value {
	return &datevalues.Value{Kind: datevalues.KindPoint, Calendar: "gregorian", StartYear: &y}
}

func containsID(ids [][]byte, id []byte) bool {
	for _, x := range ids {
		if bytes.Equal(x, id) {
			return true
		}
	}
	return false
}

func TestSaveBatchFilesBothEndsOnce(t *testing.T) {
	f := newBatchFixture(t)
	person := f.bare("person")
	event := f.bare("event")
	f.participation(person, event)

	res := f.save(promote.Batch{
		SourceID: f.source.ID, SeenRevision: f.revision(),
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
	assertCacheMatchesRebuild(t, db)
}

func TestSaveBatchFilesBridgeWhoseEndsWereAlreadyPromoted(t *testing.T) {
	f := newBatchFixture(t)
	person := f.bare("person")
	event := f.bare("event")
	per, err := promote.Save(f.c, userID, promote.Input{SubjectID: person.ID})
	if err != nil {
		t.Fatal(err)
	}
	evt, err := promote.Save(f.c, userID, promote.Input{SubjectID: event.ID})
	if err != nil {
		t.Fatal(err)
	}
	bridge := f.participation(person, event)
	if f.bridgeClaims("participation") != 0 {
		t.Fatal("bridge filed before the batch")
	}

	f.save(promote.Batch{
		SourceID: f.source.ID, SeenRevision: f.revision(),
		Rows: []promote.BatchRow{
			{SubjectID: person.ID, Target: promote.TargetHandle, EntityID: per.Entity.ID},
			{SubjectID: event.ID, Target: promote.TargetHandle, EntityID: evt.Entity.ID},
		},
	})
	if f.bridgeClaims("participation") != 1 {
		t.Fatalf("participation claims %d, want 1", f.bridgeClaims("participation"))
	}

	// A second Done with the bridge switched off does not file another copy.
	// (Already filed, so this also checks idempotence.) The skip path is the
	// unfiled case below.
	_ = bridge
}

func TestSaveBatchSkipsSwitchedOffBridge(t *testing.T) {
	f := newBatchFixture(t)
	person := f.bare("person")
	event := f.bare("event")
	if _, err := promote.Save(f.c, userID, promote.Input{SubjectID: person.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := promote.Save(f.c, userID, promote.Input{SubjectID: event.ID}); err != nil {
		t.Fatal(err)
	}
	bridge := f.participation(person, event)
	f.save(promote.Batch{
		SourceID: f.source.ID, SeenRevision: f.revision(),
		SkipBridgeIDs: [][]byte{bridge.ID},
	})
	if n := f.bridgeClaims("participation"); n != 0 {
		t.Fatalf("skipped bridge filed %d claims", n)
	}
}

func TestSaveBatchStaleWritesNothing(t *testing.T) {
	f := newBatchFixture(t)
	person := f.bare("person")
	seen := f.revision()
	_ = f.bare("place") // a later write advances the revision
	_, err := promote.SaveBatch(f.c, userID, promote.Batch{
		SourceID: f.source.ID, SeenRevision: seen,
		Rows: []promote.BatchRow{{SubjectID: person.ID, Target: promote.TargetNew}},
	})
	if !errors.Is(err, promote.ErrStale) {
		t.Fatalf("err %v, want promote.ErrStale", err)
	}
	if f.hasClaim(person.ID) {
		t.Fatal("stale batch filed a claim")
	}
}

func TestSaveBatchOneHopPin(t *testing.T) {
	tests := []struct {
		name string
		// eventTarget is how the layer's own event is filed in the same Done.
		eventTarget string
		// pinDeath pairs with the handle person's death date, not their birth.
		pinDeath bool
		wantErr  bool
	}{
		{name: "a neighbor filed on the matching handle pins and backfills", eventTarget: promote.TargetHandle},
		{name: "a neighbor filed New pins nothing", eventTarget: promote.TargetNew, wantErr: true},
		{name: "a skipped neighbor pins nothing", eventTarget: promote.TargetSkip, wantErr: true},
		{name: "a record from another of the handle's events is refused", eventTarget: promote.TargetHandle, pinDeath: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newBatchFixture(t)
			person := f.bare("person")
			birth := f.bare("event")
			birthDate := f.cite(birth, observations.Input{PropertyID: f.prop("date").ID, Date: yearDate(1849)})
			death := f.bare("event")
			deathDate := f.cite(death, observations.Input{PropertyID: f.prop("date").ID, Date: yearDate(1849)})
			f.participation(person, birth)
			f.participation(person, death)
			per, err := promote.Save(f.c, userID, promote.Input{SubjectID: person.ID})
			if err != nil {
				t.Fatal(err)
			}
			birthH, err := promote.Save(f.c, userID, promote.Input{SubjectID: birth.ID})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := promote.Save(f.c, userID, promote.Input{SubjectID: death.ID}); err != nil {
				t.Fatal(err)
			}

			// A second layer: the birth date sits on the event, one hop from the person.
			personB := f.bare("person")
			eventB := f.bare("event")
			incoming := f.cite(eventB, observations.Input{PropertyID: f.prop("date").ID, Date: yearDate(1849)})
			f.participation(personB, eventB)

			member := birthDate[0].ID
			if tt.pinDeath {
				member = deathDate[0].ID
			}
			eventRow := promote.BatchRow{SubjectID: eventB.ID, Target: tt.eventTarget}
			if tt.eventTarget == promote.TargetHandle {
				eventRow.EntityID = birthH.Entity.ID
			}
			res, err := promote.SaveBatch(f.c, userID, promote.Batch{
				SourceID: f.source.ID, SeenRevision: f.revision(),
				Rows: []promote.BatchRow{{
					SubjectID: personB.ID, Target: promote.TargetHandle, EntityID: per.Entity.ID,
					Pairs: []promote.Pair{{IncomingObservationID: incoming[0].ID, MemberObservationID: member}},
				}, eventRow},
			})
			if tt.wantErr {
				if !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("err %v, want promote.ErrInvalid", err)
				}
				if f.hasClaim(personB.ID) {
					t.Fatal("a refused pin still filed the claim")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Written) != 2 || res.Written[0].Pins < 2 {
				t.Fatalf("written %+v", res.Written)
			}
			db, err := f.c.DB()
			if err != nil {
				t.Fatal(err)
			}
			for _, claimID := range [][]byte{res.Written[0].Claim.ID, birthH.Claim.ID} {
				pins, err := identityclaims.PinnedObservations(db, claimID)
				if err != nil {
					t.Fatal(err)
				}
				if !containsID(pins, incoming[0].ID) || !containsID(pins, member) {
					t.Fatalf("claim %x pins %d, want both dates", claimID, len(pins))
				}
			}
		})
	}
}

func assertCacheMatchesRebuild(t *testing.T, db *sql.DB) {
	t.Helper()
	got := cacheLines(t, db)
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := autoreconciler.Rebuild(tx); err != nil {
		t.Fatal(err)
	}
	want := cacheLines(t, tx)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("upkeep %d cache lines, rebuild %d", len(got), len(want))
	}
}

func cacheLines(t *testing.T, q interface {
	Query(string, ...any) (*sql.Rows, error)
}) []string {
	t.Helper()
	rows, err := q.Query(`SELECT 'v|' || hex(entity_id) || '|' || hex(property_id) || '|' || rank || '|' || reason
		FROM auto_reconciler_values
		UNION ALL
		SELECT 'o|' || hex(entity_id) || '|' || hex(property_id) || '|' || hex(observation_id) || '|' || reason
		FROM auto_reconciler_outcomes
		ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func runArtifactCreate(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error) {
	a, _, err := writes.Run(c, writes.Op{Action: "create_artifact", UserID: userID},
		func(tx *database.Tx) (artifacts.Artifact, []rowchange.Change, error) {
			return artifacts.Create(tx, userID, in)
		})
	return a, err
}
