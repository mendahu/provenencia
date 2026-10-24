package promote_test

import (
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

const pinLocator = `{"version":1,"selectors":[{"type":"page","artifact_page":1}]}`

// pinFixture is a catalog with one Source and Artifact and the seeded
// vocabulary; each record is a person Subject with name Observations.
type pinFixture struct {
	t        *testing.T
	c        *database.Catalog
	artifact artifacts.Artifact
	person   subjecttypes.Type
	place    subjecttypes.Type
	name     properties.Property
	toponym  properties.Property
}

func newPinFixture(t *testing.T) *pinFixture {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	f := &pinFixture{t: t, c: c}
	r, err := ref.Mint(ref.PrefixUser)
	f.must(err)
	f.must(users.Upsert(c, userID, "Tester", r))
	f.must(subjectvocab.Install(c))
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	f.must(err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	f.must(err)
	f.artifact, err = artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	f.must(err)
	f.person, err = subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	f.must(err)
	f.place, err = subjecttypes.Lookup(c, "place", vocabkey.OriginProvenencia)
	f.must(err)
	f.name, err = properties.Lookup(c, "name", vocabkey.OriginProvenencia)
	f.must(err)
	f.toponym, err = properties.Lookup(c, "toponym", vocabkey.OriginProvenencia)
	f.must(err)
	return f
}

func (f *pinFixture) must(err error) {
	f.t.Helper()
	if err != nil {
		f.t.Fatal(err)
	}
}

// record is a new Subject of st with one name-like Observation per value, on
// prop, all on one Citation.
func (f *pinFixture) record(st subjecttypes.Type, prop properties.Property, values ...string) (subjects.Subject, []observations.Observation) {
	f.t.Helper()
	graph := evidencegraphstest.SoleGraph(f.t, f.c, f.artifact.SourceID).ID
	s, err := writes.Call(f.c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
		return subjects.Create(tx, userID, subjects.CreateInput{GraphID: graph, SubjectTypeID: st.ID}, nil)
	})
	f.must(err)
	var in []observations.Input
	for _, v := range values {
		if prop.ValueType == propertymeta.ValueTypeName {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: prop.ID, Name: &namevalue.Value{
				Form: v, Parts: []namevalue.Part{{Idx: 0, Type: "given", Value: v}},
			}})
		} else {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: prop.ID, ValueText: v, HasText: true})
		}
	}
	res, err := writes.Call(f.c, writes.Op{Action: "create_citation_with_observations", UserID: userID}, func(tx *database.Tx) (citations.CreateResult, []rowchange.Change, error) {
		return citations.CreateWithObservations(tx, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: pinLocator}, in)
	})
	f.must(err)
	return s, res.Observations
}

func (f *pinFixture) person1(values ...string) (subjects.Subject, []observations.Observation) {
	return f.record(f.person, f.name, values...)
}

// pinned is the claim's pinned Observation ids as a sorted, comparable string.
func (f *pinFixture) pinned(claimID []byte) string {
	f.t.Helper()
	db, err := f.c.DB()
	f.must(err)
	ids, err := identityclaims.PinnedObservations(db, claimID)
	f.must(err)
	var out []string
	for _, id := range ids {
		out = append(out, string(id))
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

// revision is the catalog's latest audit revision, the stamp a batch is read at.
func (f *pinFixture) revision() int64 {
	f.t.Helper()
	db, err := f.c.DB()
	f.must(err)
	var rev int64
	f.must(db.QueryRow(`SELECT COALESCE(MAX(revision), 0) FROM audit_transactions`).Scan(&rev))
	return rev
}

// batch files one Done of the fixture's Source, the one pin write.
func (f *pinFixture) batch(rows ...promote.BatchRow) (promote.BatchResult, error) {
	f.t.Helper()
	seen := f.revision()
	return writes.Call(f.c, writes.Op{Action: "promote_batch", UserID: userID}, func(tx *database.Tx) (promote.BatchResult, []rowchange.Change, error) {
		return promote.SaveBatch(tx, userID, promote.Batch{SourceID: f.artifact.SourceID, SeenRevision: seen, Rows: rows})
	})
}

// promote files subject on a new handle.
func (f *pinFixture) promote(subject subjects.Subject, argument string) promote.Written {
	f.t.Helper()
	res, err := f.batch(promote.BatchRow{SubjectID: subject.ID, Target: promote.TargetNew, Argument: argument})
	f.must(err)
	return res.Written[0]
}

// join files subject on entityID with the pin sets.
func (f *pinFixture) join(subject subjects.Subject, entityID []byte, sets ...promote.PinSet) (promote.Written, error) {
	f.t.Helper()
	res, err := f.batch(promote.BatchRow{SubjectID: subject.ID, Target: promote.TargetHandle, EntityID: entityID, PinSets: sets})
	if err != nil {
		return promote.Written{}, err
	}
	return res.Written[0], nil
}

// set pins observations on the new claim and on member's claim.
func set(member subjects.Subject, obs ...observations.Observation) promote.PinSet {
	ps := promote.PinSet{Member: member.ID}
	for _, o := range obs {
		ps.Observations = append(ps.Observations, o.ID)
	}
	return ps
}

func idsOf(obs ...observations.Observation) string {
	var out []string
	for _, o := range obs {
		out = append(out, string(o.ID))
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func TestSavePins(t *testing.T) {
	t.Run("a pin set pins its Observations on both claims", func(t *testing.T) {
		f := newPinFixture(t)
		birth, birthObs := f.person1("James")
		first := f.promote(birth, "The birth record.")
		census, censusObs := f.person1("James")
		res, err := f.join(census, first.Entity.ID, set(birth, censusObs[0], birthObs[0]))
		f.must(err)
		want := idsOf(birthObs[0], censusObs[0])
		if got := f.pinned(res.Claim.ID); got != want {
			t.Fatalf("new claim pins differ")
		}
		if got := f.pinned(first.Claim.ID); got != want {
			t.Fatalf("member claim pins differ (backfill)")
		}
		if res.Pins != 2 {
			t.Fatalf("Pins = %d, want 2", res.Pins)
		}
		member, err := identityclaims.Get(f.c, first.Claim.ID)
		f.must(err)
		if member.Argument != "The birth record." {
			t.Fatalf("member argument rewritten: %q", member.Argument)
		}
	})

	t.Run("pins are audited per claim in the promote_batch revision", func(t *testing.T) {
		f := newPinFixture(t)
		birth, birthObs := f.person1("James")
		first := f.promote(birth, "")
		census, censusObs := f.person1("James")
		res, err := f.join(census, first.Entity.ID, set(birth, censusObs[0], birthObs[0]))
		f.must(err)
		got := lastRevisionChanges(t, f.c)
		want := "promote_batch:identity_claim,identity_claim_evidence,identity_claim_evidence,identity_claim_evidence,identity_claim_evidence"
		if got != want {
			t.Fatalf("changes %s", got)
		}
		for claimID, n := range map[string]int{string(res.Claim.ID): 2, string(first.Claim.ID): 2} {
			if c := evidenceChanges(t, f.c, []byte(claimID), "create"); c != n {
				t.Fatalf("claim has %d pin changes, want %d", c, n)
			}
		}
	})

	t.Run("pins already carried are not written or audited twice", func(t *testing.T) {
		f := newPinFixture(t)
		a, aObs := f.person1("James", "Jim")
		first := f.promote(a, "")
		b, bObs := f.person1("James")
		_, err := f.join(b, first.Entity.ID, set(a, bObs[0], aObs[0]), set(a, bObs[0], aObs[0]))
		f.must(err)
		c, cObs := f.person1("James")
		// The member's Observation is already pinned on its claim by b's join.
		res, err := f.join(c, first.Entity.ID, set(a, cObs[0], aObs[0]))
		f.must(err)
		if got, want := f.pinned(first.Claim.ID), idsOf(aObs[0], bObs[0], cObs[0]); got != want {
			t.Fatalf("member pins differ")
		}
		if res.Pins != 2 {
			t.Fatalf("Pins = %d, want 2", res.Pins)
		}
		if n := evidenceChanges(t, f.c, first.Claim.ID, "create"); n != 3 {
			t.Fatalf("member claim has %d pin changes, want 3", n)
		}
	})

	t.Run("sets against several members backfill each", func(t *testing.T) {
		f := newPinFixture(t)
		a, aObs := f.person1("James")
		first := f.promote(a, "")
		b, bObs := f.person1("James")
		second, err := f.join(b, first.Entity.ID)
		f.must(err)
		c, cObs := f.person1("James")
		res, err := f.join(c, first.Entity.ID, set(a, cObs[0], aObs[0]), set(b, cObs[0], bObs[0]))
		f.must(err)
		if got, want := f.pinned(res.Claim.ID), idsOf(aObs[0], bObs[0], cObs[0]); got != want {
			t.Fatalf("new claim pins differ")
		}
		if got, want := f.pinned(first.Claim.ID), idsOf(aObs[0], cObs[0]); got != want {
			t.Fatalf("first member pins differ")
		}
		if got, want := f.pinned(second.Claim.ID), idsOf(bObs[0], cObs[0]); got != want {
			t.Fatalf("second member pins differ")
		}
	})

	t.Run("invalid pin sets are refused and write nothing", func(t *testing.T) {
		f := newPinFixture(t)
		a, aObs := f.person1("James")
		first := f.promote(a, "")
		outsider, _ := f.person1("James")
		provisional, _ := f.person1("James")
		_, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
			return identityclaims.Create(tx, userID, identityclaims.CreateInput{
				SubjectID: provisional.ID, EntityID: first.Entity.ID, Status: identityclaims.StatusProvisional})
		})
		f.must(err)
		in, inObs := f.person1("James")
		join := func(sets ...promote.PinSet) promote.BatchRow {
			return promote.BatchRow{SubjectID: in.ID, Target: promote.TargetHandle, EntityID: first.Entity.ID, PinSets: sets}
		}
		cases := map[string]promote.BatchRow{
			"pin sets on a New row": {SubjectID: in.ID, Target: promote.TargetNew,
				PinSets: []promote.PinSet{set(a, inObs[0], aObs[0])}},
			"pin sets on a Skip row": {SubjectID: in.ID, Target: promote.TargetSkip,
				PinSets: []promote.PinSet{set(a, inObs[0], aObs[0])}},
			"member with no claim":     join(set(outsider, inObs[0], aObs[0])),
			"provisional member":       join(set(provisional, inObs[0], aObs[0])),
			"malformed Observation id": join(promote.PinSet{Member: a.ID, Observations: [][]byte{inObs[0].ID[:4]}}),
			"malformed member id":      join(promote.PinSet{Member: a.ID[:4], Observations: [][]byte{inObs[0].ID}}),
		}
		for name, row := range cases {
			if _, err := f.batch(row); !errors.Is(err, promote.ErrInvalid) {
				t.Fatalf("%s: got %v", name, err)
			}
		}
		if _, err := identityclaims.AcceptedEntityForSubject(f.c, in.ID); err == nil {
			t.Fatal("a refused step filed a claim")
		}
		if got := f.pinned(first.Claim.ID); got != "" {
			t.Fatal("a refused step pinned the member")
		}
	})

	t.Run("deleting a pinned Observation releases it from both claims, audited", func(t *testing.T) {
		f := newPinFixture(t)
		a, aObs := f.person1("James", "Jim")
		first := f.promote(a, "")
		b, bObs := f.person1("James")
		res, err := f.join(b, first.Entity.ID, set(a, bObs[0], aObs[0], aObs[1]))
		f.must(err)
		{
			_, err := writes.Call(f.c, writes.Op{Action: "delete_observation", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
				changes, err := observations.Delete(tx, userID, aObs[0].ID)
				return struct{}{}, changes, err
			})
			f.must(err)
		}
		for _, claimID := range [][]byte{res.Claim.ID, first.Claim.ID} {
			if got, want := f.pinned(claimID), idsOf(aObs[1], bObs[0]); got != want {
				t.Fatalf("pins after delete differ")
			}
			if n := evidenceChanges(t, f.c, claimID, "delete"); n != 1 {
				t.Fatalf("claim has %d pin removals, want 1", n)
			}
		}
	})
}

// evidenceChanges counts the audited identity_claim_evidence changes of one
// action recorded under a claim's id.
func evidenceChanges(t *testing.T, c *database.Catalog, claimID []byte, action string) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_changes
		WHERE entity_type = 'identity_claim_evidence' AND entity_id = ? AND action = ?`, claimID, action).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
