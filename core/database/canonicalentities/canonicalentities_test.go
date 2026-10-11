package canonicalentities_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

var userID = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

func TestCanonicalEntities(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, c *database.Catalog)
	}{
		{
			name: "create mints ref from ref_prefix, not candidate prefix",
			run: func(t *testing.T, c *database.Catalog) {
				person := lookupType(t, c, "person")
				e, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{
						SubjectTypeID: person.ID, Label: " Mother of James ",
					})
				})
				if err != nil {
					t.Fatal(err)
				}
				if !ref.Valid(e.Ref) || !strings.HasPrefix(e.Ref, person.RefPrefix+"-") {
					t.Fatalf("ref %q prefix %q", e.Ref, person.RefPrefix)
				}
				if strings.HasPrefix(e.Ref, person.CandidateRefPrefix+"-") {
					t.Fatalf("minted candidate ref %q", e.Ref)
				}
				if e.Label != "Mother of James" || len(e.ID) != 16 {
					t.Fatalf("%+v", e)
				}
			},
		},
		{
			name: "count by type key excludes other types and merged handles",
			run: func(t *testing.T, c *database.Catalog) {
				person, place := lookupType(t, c, "person"), lookupType(t, c, "place")
				var persons []canonicalentities.Entity
				for i := 0; i < 3; i++ {
					e, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
						return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: person.ID})
					})
					if err != nil {
						t.Fatal(err)
					}
					persons = append(persons, e)
				}
				if _, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: place.ID})
				}); err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, persons[0].ID, persons[2].ID); err != nil {
					t.Fatal(err)
				}
				if n, err := canonicalentities.CountByTypeKey(c, "person"); err != nil || n != 2 {
					t.Fatalf("persons %d %v", n, err)
				}
				if n, err := canonicalentities.CountByTypeKey(c, "event"); err != nil || n != 0 {
					t.Fatalf("events %d %v", n, err)
				}
				if _, err := canonicalentities.CountByTypeKey(c, " "); err == nil {
					t.Fatal("blank key accepted")
				}
			},
		},
		{
			name: "create records one audit change",
			run: func(t *testing.T, c *database.Catalog) {
				place := lookupType(t, c, "place")
				e, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: place.ID})
				})
				if err != nil {
					t.Fatal(err)
				}
				db, err := c.DB()
				if err != nil {
					t.Fatal(err)
				}
				var action, entityType, changes string
				var entityID []byte
				if err := db.QueryRow(`SELECT t.action_type, ch.entity_type, ch.entity_id, ch.changes_json
					FROM audit_changes ch JOIN audit_transactions t ON t.id = ch.audit_transaction_id
					ORDER BY t.revision DESC LIMIT 1`).Scan(&action, &entityType, &entityID, &changes); err != nil {
					t.Fatal(err)
				}
				if action != "create_canonical_entity" || entityType != "canonical_entity" || string(entityID) != string(e.ID) {
					t.Fatalf("%s %s", action, entityType)
				}
				var fields map[string]map[string]any
				if err := json.Unmarshal([]byte(changes), &fields); err != nil {
					t.Fatal(err)
				}
				if fields["ref"]["new"] != e.Ref {
					t.Fatalf("%v", fields)
				}
			},
		},
		{
			name: "unknown subject type rejected",
			run: func(t *testing.T, c *database.Catalog) {
				p6a1 := canonicalentities.CreateInput{SubjectTypeID: make([]byte, 16)}
				_, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, p6a1)
				})
				if !errors.Is(err, canonicalentities.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "missing user rejected",
			run: func(t *testing.T, c *database.Catalog) {
				person := lookupType(t, c, "person")
				_, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: nil}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, nil, canonicalentities.CreateInput{SubjectTypeID: person.ID})
				})
				if !errors.Is(err, canonicalentities.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
		{
			name: "get, get by ref, and list by type round-trip",
			run: func(t *testing.T, c *database.Catalog) {
				person := lookupType(t, c, "person")
				place := lookupType(t, c, "place")
				a, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: person.ID, Argument: "why"})
				})
				if err != nil {
					t.Fatal(err)
				}
				b, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: person.ID})
				})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
					return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: place.ID})
				}); err != nil {
					t.Fatal(err)
				}
				got, err := canonicalentities.Get(c, a.ID)
				if err != nil || got.Ref != a.Ref || got.Argument != "why" || got.MergedIntoID != nil {
					t.Fatalf("%v %+v", err, got)
				}
				byRef, err := canonicalentities.GetByRef(c, b.Ref)
				if err != nil || string(byRef.ID) != string(b.ID) {
					t.Fatalf("%v %+v", err, byRef)
				}
				list, err := canonicalentities.ListByType(c, person.ID)
				if err != nil || len(list) != 2 {
					t.Fatalf("%v len=%d", err, len(list))
				}
				if _, err := canonicalentities.Get(c, make([]byte, 16)); !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("got %v", err)
				}
				if _, err := canonicalentities.GetByRef(c, "nope"); !errors.Is(err, canonicalentities.ErrInvalid) {
					t.Fatalf("got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, testCatalog(t))
		})
	}
}

func lookupType(t *testing.T, c *database.Catalog, key string) subjecttypes.Type {
	t.Helper()
	st, err := subjecttypes.Lookup(c, key, subjecttypes.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	return st
}

func testCatalog(t *testing.T) *database.Catalog {
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
	return c
}
