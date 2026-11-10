package promote_test

import (
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/evidencelayer/evidencelayertest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/promote/promotetest"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = evidencelayertest.UserID

func TestSave(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog, mk func(key string) subjects.Subject)
	}{
		{
			name: "person mints a PER handle with one accepted member",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				james := mk("person")
				res := promotetest.Save(t, c, userID, promote.Input{SubjectID: james.ID})
				if !ref.Valid(res.Entity.Ref) || !strings.HasPrefix(res.Entity.Ref, "PER-") {
					t.Fatalf("ref %q", res.Entity.Ref)
				}
				if res.Claim.Status != identityclaims.StatusAccepted || string(res.Claim.EntityID) != string(res.Entity.ID) {
					t.Fatalf("%+v", res.Claim)
				}
				got, err := identityclaims.AcceptedEntityForSubject(c, james.ID)
				if err != nil || string(got.EntityID) != string(res.Entity.ID) {
					t.Fatalf("%v %+v", err, got)
				}
				members, err := identityclaims.AcceptedMembers(c, res.Entity.ID)
				if err != nil || len(members) != 1 {
					t.Fatalf("%v %+v", err, members)
				}
			},
		},
		{
			name: "one promote_subject revision with two changes",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				res := promotetest.Save(t, c, userID, promote.Input{SubjectID: mk("person").ID})
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				action, revision := dbtest.LatestRevision(t, db)
				if action != "promote_subject" {
					t.Fatalf("action %s", action)
				}
				var got []string
				for _, ch := range revision {
					want := res.Claim.ID
					if ch.EntityType == "canonical_entity" {
						want = res.Entity.ID
					}
					if string(ch.EntityID) != string(want) {
						t.Fatalf("%s id mismatch", ch.EntityType)
					}
					got = append(got, ch.EntityType)
				}
				sort.Strings(got)
				if strings.Join(got, ",") != "canonical_entity,identity_claim" {
					t.Fatalf("changes %v", got)
				}
			},
		},
		{
			name: "second promote of a member refused without minting",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				james := mk("person")
				first := promotetest.Save(t, c, userID, promote.Input{SubjectID: james.ID})
				_, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, userID, promote.Input{SubjectID: james.ID})
				})
				if !errors.Is(err, identityclaims.ErrAlreadyMember) {
					t.Fatalf("got %v", err)
				}
				list, err := canonicalentities.ListByType(c, first.Entity.SubjectTypeID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v len=%d", err, len(list))
				}
			},
		},
		{
			name: "event and place mint their own prefixes",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				for key, prefix := range map[string]string{"event": "EVT-", "place": "PLC-"} {
					res := promotetest.Save(t, c, userID, promote.Input{SubjectID: mk(key).ID})
					if !strings.HasPrefix(res.Entity.Ref, prefix) {
						t.Fatalf("%s ref %q", key, res.Entity.Ref)
					}
				}
			},
		},
		{
			name: "join files a claim onto the existing handle without minting",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				first := promotetest.Save(t, c, userID, promote.Input{SubjectID: mk("person").ID})
				second := mk("person")
				res := promotetest.Save(t, c, userID, promote.Input{SubjectID: second.ID, EntityID: first.Entity.ID})
				if string(res.Entity.ID) != string(first.Entity.ID) || res.Entity.Ref != first.Entity.Ref {
					t.Fatalf("joined %+v, want %+v", res.Entity, first.Entity)
				}
				if string(res.Claim.SubjectID) != string(second.ID) || res.Claim.Status != identityclaims.StatusAccepted {
					t.Fatalf("%+v", res.Claim)
				}
				members, err := identityclaims.AcceptedMembers(c, first.Entity.ID)
				if err != nil || len(members) != 2 {
					t.Fatalf("%v members=%d", err, len(members))
				}
				list, err := canonicalentities.ListByType(c, first.Entity.SubjectTypeID)
				if err != nil || len(list) != 1 {
					t.Fatalf("%v handles=%d", err, len(list))
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				action, revision := dbtest.LatestRevision(t, db)
				if action != "promote_subject" || len(revision) != 1 || revision[0].EntityType != "identity_claim" {
					t.Fatalf("revision %s %+v", action, revision)
				}
			},
		},
		{
			name: "join onto another type's handle refused without writing",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				event := promotetest.Save(t, c, userID, promote.Input{SubjectID: mk("event").ID})
				james := mk("person")
				_, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, userID, promote.Input{SubjectID: james.ID, EntityID: event.Entity.ID})
				})
				if !errors.Is(err, identityclaims.ErrTypeMismatch) {
					t.Fatalf("got %v", err)
				}
				if _, err := identityclaims.AcceptedEntityForSubject(c, james.ID); err == nil {
					t.Fatal("claim written")
				}
			},
		},
		{
			name: "join onto an unknown or merged handle invalid",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				james := mk("person")
				in := promote.Input{SubjectID: james.ID, EntityID: make([]byte, 16)}
				if _, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, userID, in)
				}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("unknown: %v", err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, userID, promote.Input{SubjectID: james.ID, EntityID: []byte{1}})
				}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("short id: %v", err)
				}
				a := promotetest.Save(t, c, userID, promote.Input{SubjectID: mk("person").ID})
				b := promotetest.Save(t, c, userID, promote.Input{SubjectID: mk("person").ID})
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, b.Entity.ID, a.Entity.ID); err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, userID, promote.Input{SubjectID: james.ID, EntityID: a.Entity.ID})
				}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("merged: %v", err)
				}
			},
		},
		{
			name: "bridge kinds unsupported in v1",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				in := promote.Input{SubjectID: mk("participation").ID}
				_, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, userID, in)
				})
				if !errors.Is(err, promote.ErrUnsupportedType) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "unknown subject or missing user invalid",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				in := promote.Input{SubjectID: make([]byte, 16)}
				if _, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, userID, in)
				}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("unknown subject: %v", err)
				}
				in = promote.Input{SubjectID: mk("person").ID}
				if _, err := writes.Call(c, writes.Op{Action: "promote_subject", UserID: nil}, func(tx *database.Tx) (promote.Result, []rowchange.Change, error) {
					return promote.Save(tx, nil, in)
				}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("missing user: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, mk := newCatalog(t)
			tt.run(t, c, mk)
		})
	}
}

func newCatalog(t *testing.T) (*database.Catalog, func(key string) subjects.Subject) {
	t.Helper()
	r := evidencelayertest.New(t)
	graph := r.Graph(r.Source)
	return r.C, func(key string) subjects.Subject {
		t.Helper()
		return subjectstest.On(t, r.C, userID, graph, key, "")
	}
}
