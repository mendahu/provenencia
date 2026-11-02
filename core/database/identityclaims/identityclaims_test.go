package identityclaims_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/connectrules"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	_ "github.com/mendahu/provenencia/core/database/graphfill"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
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
				cl, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
					return identityclaims.Create(tx, userID, identityclaims.CreateInput{
						SubjectID: f.james.ID, EntityID: f.per1.ID, Status: identityclaims.StatusAccepted,
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				if string(cl.SubjectTypeID) != string(f.personType.ID) {
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
				_, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
					return identityclaims.Create(tx, userID, identityclaims.CreateInput{
						SubjectID: f.james.ID, EntityID: f.plc1.ID, Status: identityclaims.StatusAccepted,
					})
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
				_, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
					return identityclaims.Create(tx, userID, identityclaims.CreateInput{
						SubjectID: f.james.ID, EntityID: f.per2.ID, Status: identityclaims.StatusAccepted,
					})
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
				_, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
					return identityclaims.Create(tx, userID, identityclaims.CreateInput{
						SubjectID: f.james.ID, EntityID: f.per1.ID, Status: identityclaims.StatusRejected,
					})
				})
				if !errors.Is(err, identityclaims.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "bad status rejected",
			run: func(t *testing.T, f fixture) {
				_, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
					return identityclaims.Create(tx, userID, identityclaims.CreateInput{
						SubjectID: f.james.ID, EntityID: f.per1.ID, Status: "maybe",
					})
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
				if _, err := writes.Call(f.c, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
					changes, err := subjects.Delete(tx, userID, f.james.ID)
					return struct{}{}, changes, err
				}); err != nil {
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
	cl, err := writes.Call(f.c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
		return identityclaims.Create(tx, userID, identityclaims.CreateInput{
			SubjectID: s.ID, EntityID: e.ID, Status: status,
		})
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
		Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book",
	})
	if err != nil {
		t.Fatal(err)
	}
	src, _, err := writes.Run(c, writes.Op{Action: "create_source", UserID: userID}, func(tx *database.Tx) (sources.Source, []rowchange.Change, error) {
		return sources.Create(tx, userID, sources.CreateInput{SourceTypeID: typeID, Title: "Census"})
	})
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
	st, err := subjecttypes.Lookup(c, key, vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func mustSubject(t *testing.T, c *database.Catalog, sourceID, typeID []byte, label string) subjects.Subject {
	t.Helper()
	graph := evidencegraphstest.SoleGraph(t, c, sourceID).ID
	return subjectstest.OfType(t, c, userID, graph, typeID, label)
}

func mustEntity(t *testing.T, c *database.Catalog, typeID []byte) canonicalentities.Entity {
	t.Helper()
	e, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: typeID})
	})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestAcceptedClaimsAreSourceMembers(t *testing.T) {
	f := newFixture(t)
	mustClaim(t, f, f.james, f.per1, identityclaims.StatusAccepted)
	mustClaim(t, f, f.jim, f.per1, identityclaims.StatusAccepted)
	mustClaim(t, f, f.york, f.plc1, identityclaims.StatusProvisional)

	eg, err := f.c.Graph().EvidenceGraph(f.james.GraphID)
	if err != nil {
		t.Fatal(err)
	}
	bySubject := map[string][]byte{}
	for _, m := range eg.Members {
		bySubject[string(m.SubjectID)] = m.EntityID
	}
	for _, s := range []subjects.Subject{f.james, f.jim} {
		if string(bySubject[string(s.ID)]) != string(f.per1.ID) {
			t.Fatalf("%s member %x", s.Label, bySubject[string(s.ID)])
		}
	}
	if _, ok := bySubject[string(f.york.ID)]; ok || len(eg.Members) != 2 {
		t.Fatalf("provisional subject listed: %+v", eg.Members)
	}
}

// TestBridgeHeadQueryPlans: each bridge-head query seeks its Subjects by id
// (or graph) instead of starting at the subject type index, which would read
// every bridge Subject in the catalog.
func TestBridgeHeadQueryPlans(t *testing.T) {
	f := newFixture(t)
	db, err := f.c.DB()
	if err != nil {
		t.Fatal(err)
	}
	keys := connectrules.BridgeTypeKeys()
	id := make([]byte, 16)
	tests := []struct {
		name, cond string
		nArgs      int
		seek       string
	}{
		{"file", identityclaims.FileCond, 2, "sqlite_autoindex_subjects_1"},
		{"graph", identityclaims.GraphCond, 1, "subjects_graph_id_idx"},
		{"unfile", identityclaims.UnfileCond, 1, "sqlite_autoindex_subjects_1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := make([]any, 0, len(keys)+tc.nArgs)
			for _, k := range keys {
				args = append(args, k)
			}
			for range tc.nArgs {
				args = append(args, id)
			}
			seeks := false
			for _, detail := range dbtest.QueryPlan(t, db, identityclaims.BridgeHeadsSQL(tc.cond, len(keys)), args...) {
				t.Log(detail)
				if strings.Contains(detail, "subjects_subject_type_id_idx") || strings.HasPrefix(detail, "SCAN s") {
					t.Errorf("plan starts at the type index or scans: %s", detail)
				}
				if strings.HasPrefix(detail, "SEARCH s ") && strings.Contains(detail, tc.seek) {
					seeks = true
				}
			}
			if !seeks {
				t.Errorf("subjects step does not seek %s", tc.seek)
			}
		})
	}
}
