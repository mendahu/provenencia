package identityclaims_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/claimconfidencegrades"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

type fixture struct {
	c                     *database.Catalog
	personType, placeType subjecttypes.Type
	james, jim, york      subjects.Subject
	per1, per2, plc1      canonicalentities.Entity
}

func TestIdentityClaims(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, f fixture)
	}{
		{
			name: "accepted claim copies subject type and makes a member",
			run: func(t *testing.T, f fixture) {
				cl, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{
					SubjectID: f.james.ID, EntityID: f.per1.ID, Status: identityclaims.StatusAccepted,
				})
				if err != nil {
					t.Fatal(err)
				}
				if string(cl.SubjectTypeID) != string(f.personType.ID) || cl.ConfidenceGradeID != nil {
					t.Fatalf("%+v", cl)
				}
				got, err := identityclaims.AcceptedEntityForSubject(f.c, f.james.ID)
				if err != nil || string(got.EntityID) != string(f.per1.ID) {
					t.Fatalf("%v %+v", err, got)
				}
				members, err := identityclaims.AcceptedMembers(f.c, f.per1.ID)
				if err != nil || len(members) != 1 || string(members[0].SubjectID) != string(f.james.ID) {
					t.Fatalf("%v %+v", err, members)
				}
			},
		},
		{
			name: "unpromoted subject has no membership",
			run: func(t *testing.T, f fixture) {
				if _, err := identityclaims.AcceptedEntityForSubject(f.c, f.jim.ID); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				members, err := identityclaims.AcceptedMembers(f.c, f.per1.ID)
				if err != nil || len(members) != 0 {
					t.Fatalf("%v %+v", err, members)
				}
			},
		},
		{
			name: "type mismatch rejected by composite FK",
			run: func(t *testing.T, f fixture) {
				_, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{
					SubjectID: f.james.ID, EntityID: f.plc1.ID, Status: identityclaims.StatusAccepted,
				})
				if !errors.Is(err, identityclaims.ErrTypeMismatch) {
					t.Fatalf("got %v", err)
				}
				// The FK itself, not only the store, refuses the pair.
				db, err := f.c.DB()
				if err != nil {
					t.Fatal(err)
				}
				id := uuid.Must(uuid.NewV7())
				_, err = db.Exec(`INSERT INTO identity_claims (id, subject_id, entity_id, subject_type_id, status)
					VALUES (?, ?, ?, ?, 'accepted')`, id[:], f.james.ID, f.plc1.ID, f.placeType.ID)
				if !database.IsConstraintViolation(err) {
					t.Fatalf("raw insert: %v", err)
				}
			},
		},
		{
			name: "second accepted claim for a subject rejected",
			run: func(t *testing.T, f fixture) {
				mustClaim(t, f, f.james, f.per1, identityclaims.StatusAccepted)
				_, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{
					SubjectID: f.james.ID, EntityID: f.per2.ID, Status: identityclaims.StatusAccepted,
				})
				if !errors.Is(err, identityclaims.ErrAlreadyMember) {
					t.Fatalf("got %v", err)
				}
				// The partial unique index enforces it below the store too.
				db, err := f.c.DB()
				if err != nil {
					t.Fatal(err)
				}
				id := uuid.Must(uuid.NewV7())
				_, err = db.Exec(`INSERT INTO identity_claims (id, subject_id, entity_id, subject_type_id, status)
					VALUES (?, ?, ?, ?, 'accepted')`, id[:], f.james.ID, f.per2.ID, f.personType.ID)
				if !database.IsUniqueConflict(err) {
					t.Fatalf("raw insert: %v", err)
				}
			},
		},
		{
			name: "provisional claim beside an accepted one is allowed",
			run: func(t *testing.T, f fixture) {
				mustClaim(t, f, f.james, f.per1, identityclaims.StatusAccepted)
				mustClaim(t, f, f.james, f.per2, identityclaims.StatusProvisional)
				got, err := identityclaims.AcceptedEntityForSubject(f.c, f.james.ID)
				if err != nil || string(got.EntityID) != string(f.per1.ID) {
					t.Fatalf("%v %+v", err, got)
				}
			},
		},
		{
			name: "duplicate subject entity pair rejected",
			run: func(t *testing.T, f fixture) {
				mustClaim(t, f, f.james, f.per1, identityclaims.StatusProvisional)
				_, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{
					SubjectID: f.james.ID, EntityID: f.per1.ID, Status: identityclaims.StatusRejected,
				})
				if !errors.Is(err, identityclaims.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "bad status rejected",
			run: func(t *testing.T, f fixture) {
				_, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{
					SubjectID: f.james.ID, EntityID: f.per1.ID, Status: "maybe",
				})
				if !errors.Is(err, identityclaims.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
				db, err := f.c.DB()
				if err != nil {
					t.Fatal(err)
				}
				id := uuid.Must(uuid.NewV7())
				_, err = db.Exec(`INSERT INTO identity_claims (id, subject_id, entity_id, subject_type_id, status)
					VALUES (?, ?, ?, ?, 'maybe')`, id[:], f.james.ID, f.per1.ID, f.personType.ID)
				if !database.IsConstraintViolation(err) {
					t.Fatalf("raw insert: %v", err)
				}
			},
		},
		{
			name: "confidence grade and argument stored",
			run: func(t *testing.T, f fixture) {
				if err := claimconfidencegrades.Install(f.c); err != nil {
					t.Fatal(err)
				}
				high, err := claimconfidencegrades.Lookup(f.c, "high_confidence", claimconfidencegrades.OriginProvenencia)
				if err != nil {
					t.Fatal(err)
				}
				cl, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{
					SubjectID: f.james.ID, EntityID: f.per1.ID, Status: identityclaims.StatusAccepted,
					ConfidenceGradeID: high.ID, Argument: " Same age and parish. ",
				})
				if err != nil {
					t.Fatal(err)
				}
				got, err := identityclaims.Get(f.c, cl.ID)
				if err != nil || string(got.ConfidenceGradeID) != string(high.ID) || got.Argument != "Same age and parish." {
					t.Fatalf("%v %+v", err, got)
				}
				_, err = identityclaims.Create(f.c, userID, identityclaims.CreateInput{
					SubjectID: f.jim.ID, EntityID: f.per1.ID, Status: identityclaims.StatusAccepted,
					ConfidenceGradeID: make([]byte, 16),
				})
				if !errors.Is(err, identityclaims.ErrInvalid) {
					t.Fatalf("unknown grade: %v", err)
				}
			},
		},
		{
			name: "create records an audit change",
			run: func(t *testing.T, f fixture) {
				cl := mustClaim(t, f, f.james, f.per1, identityclaims.StatusAccepted)
				db, err := f.c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var action, entityType string
				var entityID []byte
				if err := db.QueryRow(`SELECT t.action_type, ch.entity_type, ch.entity_id
					FROM audit_changes ch JOIN audit_transactions t ON t.id = ch.audit_transaction_id
					ORDER BY t.revision DESC LIMIT 1`).Scan(&action, &entityType, &entityID); err != nil {
					t.Fatal(err)
				}
				if action != "create_identity_claim" || entityType != "identity_claim" || string(entityID) != string(cl.ID) {
					t.Fatalf("%s %s", action, entityType)
				}
			},
		},
		{
			name: "subject delete cascades its claim, handle stays",
			run: func(t *testing.T, f fixture) {
				cl := mustClaim(t, f, f.james, f.per1, identityclaims.StatusAccepted)
				if err := subjects.Delete(f.c, userID, f.james.ID); err != nil {
					t.Fatal(err)
				}
				if _, err := identityclaims.Get(f.c, cl.ID); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("claim survived: %v", err)
				}
				if _, err := canonicalentities.Get(f.c, f.per1.ID); err != nil {
					t.Fatalf("handle gone: %v", err)
				}
			},
		},
		{
			name: "members listed both ways inside a transaction",
			run: func(t *testing.T, f fixture) {
				mustClaim(t, f, f.james, f.per1, identityclaims.StatusAccepted)
				mustClaim(t, f, f.jim, f.per1, identityclaims.StatusAccepted)
				db, err := f.c.DB()
				if err != nil {
					t.Fatal(err)
				}
				tx, err := db.Begin()
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback() }()
				members, err := identityclaims.AcceptedMembersTx(tx, f.per1.ID)
				if err != nil || len(members) != 2 {
					t.Fatalf("%v len=%d", err, len(members))
				}
				got, err := identityclaims.AcceptedEntityForSubjectTx(tx, f.jim.ID)
				if err != nil || string(got.EntityID) != string(f.per1.ID) {
					t.Fatalf("%v %+v", err, got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, newFixture(t))
		})
	}
}

func mustClaim(t *testing.T, f fixture, s subjects.Subject, e canonicalentities.Entity, status string) identityclaims.Claim {
	t.Helper()
	cl, err := identityclaims.Create(f.c, userID, identityclaims.CreateInput{
		SubjectID: s.ID, EntityID: e.ID, Status: status,
	})
	if err != nil {
		t.Fatal(err)
	}
	return cl
}

func newFixture(t *testing.T) fixture {
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
	src, err := sources.Create(c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Census"})
	if err != nil {
		t.Fatal(err)
	}
	f := fixture{c: c}
	f.personType = lookupType(t, c, "person")
	f.placeType = lookupType(t, c, "place")
	f.james = mustSubject(t, c, src.ID, f.personType.ID, "James")
	f.jim = mustSubject(t, c, src.ID, f.personType.ID, "Jim")
	f.york = mustSubject(t, c, src.ID, f.placeType.ID, "York")
	f.per1 = mustEntity(t, c, f.personType.ID)
	f.per2 = mustEntity(t, c, f.personType.ID)
	f.plc1 = mustEntity(t, c, f.placeType.ID)
	return f
}

func lookupType(t *testing.T, c *database.Catalog, key string) subjecttypes.Type {
	t.Helper()
	st, err := subjecttypes.Lookup(c, key, subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func mustSubject(t *testing.T, c *database.Catalog, sourceID, typeID []byte, label string) subjects.Subject {
	t.Helper()
	s, err := subjects.Create(c, userID, subjects.CreateInput{SourceID: sourceID, SubjectTypeID: typeID, Label: label}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustEntity(t *testing.T, c *database.Catalog, typeID []byte) canonicalentities.Entity {
	t.Helper()
	e, err := canonicalentities.Create(c, userID, canonicalentities.CreateInput{SubjectTypeID: typeID})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestMembershipsBySource(t *testing.T) {
	f := newFixture(t)
	claimByLabel := map[string][]byte{
		"James": mustClaim(t, f, f.james, f.per1, identityclaims.StatusAccepted).ID,
		"Jim":   mustClaim(t, f, f.jim, f.per1, identityclaims.StatusAccepted).ID,
	}
	mustClaim(t, f, f.york, f.plc1, identityclaims.StatusProvisional)

	got, err := identityclaims.MembershipsBySource(f.c, f.james.SourceID)
	if err != nil {
		t.Fatal(err)
	}
	bySubject := map[string]identityclaims.Membership{}
	for _, m := range got {
		bySubject[string(m.SubjectID)] = m
	}
	for _, s := range []subjects.Subject{f.james, f.jim} {
		m, ok := bySubject[string(s.ID)]
		if !ok || string(m.Entity.ID) != string(f.per1.ID) || m.Entity.Ref != f.per1.Ref || m.Kind != "person" ||
			string(m.ClaimID) != string(claimByLabel[s.Label]) {
			t.Fatalf("%s: %+v", s.Label, m)
		}
	}
	if _, ok := bySubject[string(f.york.ID)]; ok || len(got) != 2 {
		t.Fatalf("provisional-only subject listed: %+v", got)
	}

	t.Run("other source and bad id", func(t *testing.T) {
		typeID, err := sourcetypes.Upsert(f.c, sourcetypes.Type{
			Key: "book", Origin: sourcetypes.OriginProvenencia, Label: "Book",
		})
		if err != nil {
			t.Fatal(err)
		}
		other, err := sources.Create(f.c, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Other"})
		if err != nil {
			t.Fatal(err)
		}
		list, err := identityclaims.MembershipsBySource(f.c, other.ID)
		if err != nil || len(list) != 0 {
			t.Fatalf("%v %+v", err, list)
		}
		if _, err := identityclaims.MembershipsBySource(f.c, []byte{1}); !errors.Is(err, identityclaims.ErrInvalid) {
			t.Fatalf("got %v", err)
		}
	})

	t.Run("query uses indexes, no full scans", func(t *testing.T) {
		db, err := f.c.DB()
		if err != nil {
			t.Fatal(err)
		}
		rows, err := db.Query(`EXPLAIN QUERY PLAN SELECT s.id FROM subjects s
			JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'
			JOIN canonical_entities e ON e.id = ic.entity_id
			JOIN subject_types st ON st.id = e.subject_type_id
			WHERE s.source_id = ?`, f.james.SourceID)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var id, parent, notUsed int
			var detail string
			if err := rows.Scan(&id, &parent, &notUsed, &detail); err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(detail, "SCAN ") {
				t.Fatalf("full scan: %s", detail)
			}
		}
	})
}
