package promote_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/promote"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
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
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Register"})
	if err != nil {
		t.Fatal(err)
	}
	return c, func(key string) subjects.Subject {
		t.Helper()
		st, err := subjecttypes.Lookup(c, key, subjecttypes.OriginProvenencia)
		if err != nil {
			t.Fatal(err)
		}
		s, err := subjects.Create(c, userID, subjects.CreateInput{SourceID: src.ID, SubjectTypeID: st.ID}, nil)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
}
