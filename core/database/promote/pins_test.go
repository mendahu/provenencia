package promote_test

import (
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/citations"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/namevalues"
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
	"github.com/mendahu/provenencia/core/ref"
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book"})
	f.must(err)
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	f.must(err)
	f.artifact, err = runArtifactCreate(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	f.must(err)
	f.person, err = subjecttypes.Lookup(c, "person", subjecttypes.OriginProvenencia)
	f.must(err)
	f.place, err = subjecttypes.Lookup(c, "place", subjecttypes.OriginProvenencia)
	f.must(err)
	f.name, err = properties.Lookup(c, "name", properties.OriginProvenencia)
	f.must(err)
	f.toponym, err = properties.Lookup(c, "toponym", properties.OriginProvenencia)
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
	s, err := subjects.Create(f.c, userID, subjects.CreateInput{SourceID: f.artifact.SourceID, SubjectTypeID: st.ID}, nil)
	f.must(err)
	var in []observations.Input
	for _, v := range values {
		if prop.ValueType == properties.ValueTypeName {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: prop.ID, Name: &namevalues.Value{
				Form: v, Parts: []namevalues.Part{{Idx: 0, Type: "given", Value: v}},
			}})
		} else {
			in = append(in, observations.Input{SubjectID: s.ID, PropertyID: prop.ID, ValueText: v, HasText: true})
		}
	}
	res, err := citations.CreateWithObservations(f.c, userID, citations.CreateInput{ArtifactID: f.artifact.ID, LocatorJSON: pinLocator}, in)
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

func idsOf(obs ...observations.Observation) string {
	var out []string
	for _, o := range obs {
		out = append(out, string(o.ID))
	}
	sort.Strings(out)
	return strings.Join(out, ",")
}

func TestSavePins(t *testing.T) {
	t.Run("a confirmed pair pins both Observations on both claims", func(t *testing.T) {
		f := newPinFixture(t)
		birth, birthObs := f.person1("James")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: birth.ID, Argument: "The birth record."})
		f.must(err)
		census, censusObs := f.person1("James")
		res, err := promote.Save(f.c, userID, promote.Input{SubjectID: census.ID, EntityID: first.Entity.ID,
			Pairs: []promote.Pair{{IncomingObservationID: censusObs[0].ID, MemberObservationID: birthObs[0].ID}}})
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

	t.Run("pins are audited per claim in the promote_subject revision", func(t *testing.T) {
		f := newPinFixture(t)
		birth, birthObs := f.person1("James")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: birth.ID})
		f.must(err)
		census, censusObs := f.person1("James")
		res, err := promote.Save(f.c, userID, promote.Input{SubjectID: census.ID, EntityID: first.Entity.ID,
			Pairs: []promote.Pair{{IncomingObservationID: censusObs[0].ID, MemberObservationID: birthObs[0].ID}}})
		f.must(err)
		got := lastRevisionChanges(t, f.c)
		want := "promote_subject:identity_claim,identity_claim_evidence,identity_claim_evidence,identity_claim_evidence,identity_claim_evidence"
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
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		b, bObs := f.person1("James")
		pair := promote.Pair{IncomingObservationID: bObs[0].ID, MemberObservationID: aObs[0].ID}
		_, err = promote.Save(f.c, userID, promote.Input{SubjectID: b.ID, EntityID: first.Entity.ID,
			Pairs: []promote.Pair{pair, pair}})
		f.must(err)
		c, cObs := f.person1("James")
		// The member's Observation is already pinned on its claim by b's join.
		res, err := promote.Save(f.c, userID, promote.Input{SubjectID: c.ID, EntityID: first.Entity.ID,
			Pairs: []promote.Pair{{IncomingObservationID: cObs[0].ID, MemberObservationID: aObs[0].ID}}})
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

	t.Run("pairs against several members backfill each", func(t *testing.T) {
		f := newPinFixture(t)
		a, aObs := f.person1("James")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		b, bObs := f.person1("James")
		second, err := promote.Save(f.c, userID, promote.Input{SubjectID: b.ID, EntityID: first.Entity.ID})
		f.must(err)
		c, cObs := f.person1("James")
		res, err := promote.Save(f.c, userID, promote.Input{SubjectID: c.ID, EntityID: first.Entity.ID, Pairs: []promote.Pair{
			{IncomingObservationID: cObs[0].ID, MemberObservationID: aObs[0].ID},
			{IncomingObservationID: cObs[0].ID, MemberObservationID: bObs[0].ID},
		}})
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

	t.Run("invalid pairs are refused and write nothing", func(t *testing.T) {
		f := newPinFixture(t)
		a, aObs := f.person1("James")
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		_, outsiderObs := f.person1("James")
		provisional, provisionalObs := f.person1("James")
		_, err = identityclaims.Create(f.c, userID, identityclaims.CreateInput{
			SubjectID: provisional.ID, EntityID: first.Entity.ID, Status: identityclaims.StatusProvisional})
		f.must(err)
		_, otherPropObs := f.record(f.place, f.toponym, "York")
		in, inObs := f.person1("James")
		cases := map[string]promote.Input{
			"pairs on a mint": {SubjectID: in.ID,
				Pairs: []promote.Pair{{IncomingObservationID: inObs[0].ID, MemberObservationID: aObs[0].ID}}},
			"incoming Observation not the Subject's": {SubjectID: in.ID, EntityID: first.Entity.ID,
				Pairs: []promote.Pair{{IncomingObservationID: outsiderObs[0].ID, MemberObservationID: aObs[0].ID}}},
			"member Observation's Subject not a member": {SubjectID: in.ID, EntityID: first.Entity.ID,
				Pairs: []promote.Pair{{IncomingObservationID: inObs[0].ID, MemberObservationID: outsiderObs[0].ID}}},
			"provisional member": {SubjectID: in.ID, EntityID: first.Entity.ID,
				Pairs: []promote.Pair{{IncomingObservationID: inObs[0].ID, MemberObservationID: provisionalObs[0].ID}}},
			"different Property": {SubjectID: in.ID, EntityID: first.Entity.ID,
				Pairs: []promote.Pair{{IncomingObservationID: inObs[0].ID, MemberObservationID: otherPropObs[0].ID}}},
			"malformed id": {SubjectID: in.ID, EntityID: first.Entity.ID,
				Pairs: []promote.Pair{{IncomingObservationID: inObs[0].ID[:4], MemberObservationID: aObs[0].ID}}},
		}
		for name, input := range cases {
			if _, err := promote.Save(f.c, userID, input); !errors.Is(err, promote.ErrInvalid) {
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
		first, err := promote.Save(f.c, userID, promote.Input{SubjectID: a.ID})
		f.must(err)
		b, bObs := f.person1("James")
		res, err := promote.Save(f.c, userID, promote.Input{SubjectID: b.ID, EntityID: first.Entity.ID, Pairs: []promote.Pair{
			{IncomingObservationID: bObs[0].ID, MemberObservationID: aObs[0].ID},
			{IncomingObservationID: bObs[0].ID, MemberObservationID: aObs[1].ID},
		}})
		f.must(err)
		f.must(observations.Delete(f.c, userID, aObs[0].ID))
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
