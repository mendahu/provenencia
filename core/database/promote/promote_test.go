package promote_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/claimconfidencegrades"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
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

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

func TestSave(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog, mk func(key string) subjects.Subject)
	}{
		{
			name: "person mints a PER handle with one accepted member",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				james := mk("person")
				res, err := promote.Save(c, userID, promote.Input{SubjectID: james.ID})
				if err != nil {
					t.Fatal(err)
				}
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
				res, err := promote.Save(c, userID, promote.Input{SubjectID: mk("person").ID})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				rows, err := db.Query(`SELECT t.action_type, ch.entity_type, ch.entity_id
					FROM audit_changes ch JOIN audit_transactions t ON t.id = ch.audit_transaction_id
					WHERE t.revision = (SELECT MAX(revision) FROM audit_transactions)
					ORDER BY ch.entity_type`)
				if err != nil {
					t.Fatal(err)
				}
				defer rows.Close()
				var got []string
				for rows.Next() {
					var action, entityType string
					var id []byte
					if err := rows.Scan(&action, &entityType, &id); err != nil {
						t.Fatal(err)
					}
					if action != "promote_subject" {
						t.Fatalf("action %s", action)
					}
					want := res.Claim.ID
					if entityType == "canonical_entity" {
						want = res.Entity.ID
					}
					if string(id) != string(want) {
						t.Fatalf("%s id mismatch", entityType)
					}
					got = append(got, entityType)
				}
				if strings.Join(got, ",") != "canonical_entity,identity_claim" {
					t.Fatalf("changes %v", got)
				}
			},
		},
		{
			name: "second promote of a member refused without minting",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				james := mk("person")
				first, err := promote.Save(c, userID, promote.Input{SubjectID: james.ID})
				if err != nil {
					t.Fatal(err)
				}
				_, err = promote.Save(c, userID, promote.Input{SubjectID: james.ID})
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
					res, err := promote.Save(c, userID, promote.Input{SubjectID: mk(key).ID})
					if err != nil {
						t.Fatal(err)
					}
					if !strings.HasPrefix(res.Entity.Ref, prefix) {
						t.Fatalf("%s ref %q", key, res.Entity.Ref)
					}
				}
			},
		},
		{
			name: "join files a claim onto the existing handle without minting",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				first, err := promote.Save(c, userID, promote.Input{SubjectID: mk("person").ID})
				if err != nil {
					t.Fatal(err)
				}
				second := mk("person")
				res, err := promote.Save(c, userID, promote.Input{SubjectID: second.ID, EntityID: first.Entity.ID})
				if err != nil {
					t.Fatal(err)
				}
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
				if got := lastRevisionChanges(t, c); got != "promote_subject:identity_claim" {
					t.Fatalf("revision %s", got)
				}
			},
		},
		{
			name: "join onto another type's handle refused without writing",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				event, err := promote.Save(c, userID, promote.Input{SubjectID: mk("event").ID})
				if err != nil {
					t.Fatal(err)
				}
				james := mk("person")
				_, err = promote.Save(c, userID, promote.Input{SubjectID: james.ID, EntityID: event.Entity.ID})
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
				if _, err := promote.Save(c, userID, promote.Input{SubjectID: james.ID, EntityID: make([]byte, 16)}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("unknown: %v", err)
				}
				if _, err := promote.Save(c, userID, promote.Input{SubjectID: james.ID, EntityID: []byte{1}}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("short id: %v", err)
				}
				a, err := promote.Save(c, userID, promote.Input{SubjectID: mk("person").ID})
				if err != nil {
					t.Fatal(err)
				}
				b, err := promote.Save(c, userID, promote.Input{SubjectID: mk("person").ID})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, b.Entity.ID, a.Entity.ID); err != nil {
					t.Fatal(err)
				}
				if _, err := promote.Save(c, userID, promote.Input{SubjectID: james.ID, EntityID: a.Entity.ID}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("merged: %v", err)
				}
			},
		},
		{
			name: "confidence grade and argument land on the claim on both paths",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				if err := claimconfidencegrades.Install(c); err != nil {
					t.Fatal(err)
				}
				high, err := claimconfidencegrades.Lookup(c, "high_confidence", claimconfidencegrades.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				minted, err := promote.Save(c, userID, promote.Input{
					SubjectID: mk("person").ID, ConfidenceGradeID: high.ID, Argument: "  Grounding entry.  ",
				})
				if err != nil {
					t.Fatal(err)
				}
				joined, err := promote.Save(c, userID, promote.Input{
					SubjectID: mk("person").ID, EntityID: minted.Entity.ID, ConfidenceGradeID: high.ID, Argument: "Same name and age.",
				})
				if err != nil {
					t.Fatal(err)
				}
				for _, tc := range []struct {
					claim identityclaims.Claim
					arg   string
				}{{minted.Claim, "Grounding entry."}, {joined.Claim, "Same name and age."}} {
					got, err := identityclaims.Get(c, tc.claim.ID)
					if err != nil {
						t.Fatal(err)
					}
					if string(got.ConfidenceGradeID) != string(high.ID) || got.Argument != tc.arg {
						t.Fatalf("%+v", got)
					}
				}
				if _, err := promote.Save(c, userID, promote.Input{
					SubjectID: mk("person").ID, ConfidenceGradeID: make([]byte, 16),
				}); !errors.Is(err, identityclaims.ErrInvalid) {
					t.Fatalf("unknown grade: %v", err)
				}
			},
		},
		{
			name: "bridge kinds unsupported in v1",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				_, err := promote.Save(c, userID, promote.Input{SubjectID: mk("participation").ID})
				if !errors.Is(err, promote.ErrUnsupportedType) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "unknown subject or missing user invalid",
			run: func(t *testing.T, c *database.Catalog, mk func(string) subjects.Subject) {
				if _, err := promote.Save(c, userID, promote.Input{SubjectID: make([]byte, 16)}); !errors.Is(err, promote.ErrInvalid) {
					t.Fatalf("unknown subject: %v", err)
				}
				if _, err := promote.Save(c, nil, promote.Input{SubjectID: mk("person").ID}); !errors.Is(err, promote.ErrInvalid) {
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

// lastRevisionChanges is "action:entity_type,…" for the newest revision.
func lastRevisionChanges(t *testing.T, c *database.Catalog) string {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT t.action_type, ch.entity_type
		FROM audit_changes ch JOIN audit_transactions t ON t.id = ch.audit_transaction_id
		WHERE t.revision = (SELECT MAX(revision) FROM audit_transactions)
		ORDER BY ch.entity_type`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var action string
	var types []string
	for rows.Next() {
		var entityType string
		if err := rows.Scan(&action, &entityType); err != nil {
			t.Fatal(err)
		}
		types = append(types, entityType)
	}
	return action + ":" + strings.Join(types, ",")
}

func newCatalog(t *testing.T) (*database.Catalog, func(key string) subjects.Subject) {
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
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{
		Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, func(key string) subjects.Subject {
		t.Helper()
		st, err := subjecttypes.Lookup(c, key, subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		s, err := writes.Call(c, writes.Op{Action: "create_subject", UserID: userID}, func(tx *database.Tx) (subjects.Subject, []rowchange.Change, error) {
			return subjects.Create(tx, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: st.ID}, nil)
		})
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
}
