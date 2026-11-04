package deleteimpact_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/artifacts"
	"github.com/mendahu/provenencia/core/database/artifacts/artifactstest"
	"github.com/mendahu/provenencia/core/database/canonicalentities"
	"github.com/mendahu/provenencia/core/database/deleteimpact"
	"github.com/mendahu/provenencia/core/database/evidencegraphs/evidencegraphstest"
	"github.com/mendahu/provenencia/core/database/identityclaims"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/sources"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjects"
	"github.com/mendahu/provenencia/core/database/subjects/subjectstest"
	"github.com/mendahu/provenencia/core/database/subjecttypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/vocabkey"
	"github.com/mendahu/provenencia/core/writes"
)

func TestReleaseClaimDeletesEmptiedHandle(t *testing.T) {
	t.Run("sole accepted member deletes handle", func(t *testing.T) {
		c, userID, james, per := soleMemberFixture(t)
		cl := mustClaim(t, c, userID, james, per)
		releaseAndAssertHandleGone(t, c, userID, cl.ID, per.ID)
	})

	t.Run("another accepted member keeps handle", func(t *testing.T) {
		c, userID, james, jim, per := twoMemberFixture(t)
		mustClaim(t, c, userID, jim, per)
		cl := mustClaim(t, c, userID, james, per)
		if _, err := releaseClaim(t, c, userID, cl.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := canonicalentities.Get(c, per.ID); err != nil {
			t.Fatalf("handle gone: %v", err)
		}
	})

	t.Run("provisional claim from another subject goes with handle", func(t *testing.T) {
		c, userID, james, jim, per := twoMemberFixture(t)
		clA := mustClaim(t, c, userID, james, per)
		clB, err := writes.Call(c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
			return identityclaims.Create(tx, userID, identityclaims.CreateInput{
				SubjectID: jim.ID, EntityID: per.ID, Status: identityclaims.StatusProvisional,
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		releaseAndAssertHandleGone(t, c, userID, clA.ID, per.ID)
		if _, err := identityclaims.Get(c, clB.ID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("provisional claim survived: %v", err)
		}
	})

	t.Run("provisional on handle with no accepted member stays", func(t *testing.T) {
		c, userID, james, per := soleMemberFixture(t)
		cl, err := writes.Call(c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
			return identityclaims.Create(tx, userID, identityclaims.CreateInput{
				SubjectID: james.ID, EntityID: per.ID, Status: identityclaims.StatusProvisional,
			})
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := releaseClaim(t, c, userID, cl.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := canonicalentities.Get(c, per.ID); err != nil {
			t.Fatalf("handle gone: %v", err)
		}
	})
}

func personFixture(t *testing.T) (*database.Catalog, []byte, []byte, []byte, subjects.Subject, subjects.Subject) {
	t.Helper()
	c, userID := testCatalog(t)
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	src := mustSource(t, c, userID)
	james := mustSubject(t, c, userID, src.ID, personType.ID, "James")
	jim := mustSubject(t, c, userID, src.ID, personType.ID, "Jim")
	return c, userID, src.ID, personType.ID, james, jim
}

func soleMemberFixture(t *testing.T) (*database.Catalog, []byte, subjects.Subject, canonicalentities.Entity) {
	t.Helper()
	c, userID, _, typeID, james, _ := personFixture(t)
	per := mustEntity(t, c, userID, typeID)
	return c, userID, james, per
}

func twoMemberFixture(t *testing.T) (*database.Catalog, []byte, subjects.Subject, subjects.Subject, canonicalentities.Entity) {
	t.Helper()
	c, userID, _, typeID, james, jim := personFixture(t)
	per := mustEntity(t, c, userID, typeID)
	return c, userID, james, jim, per
}

func TestLastMemberGroupMatchesRelease(t *testing.T) {
	c, userID := testCatalog(t)
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	personType, err := subjecttypes.Lookup(c, "person", vocabkey.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	src := mustSource(t, c, userID)
	james := mustSubject(t, c, userID, src.ID, personType.ID, "James")
	per := mustEntity(t, c, userID, personType.ID)
	cl := mustClaim(t, c, userID, james, per)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	group, err := deleteimpact.LastMemberGroup(tx, james.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if group.Via != deleteimpact.ViaLastMember || group.Total != 1 || group.Listed[0].Ref != per.Ref {
		t.Fatalf("%+v", group)
	}
	releaseAndAssertHandleGone(t, c, userID, cl.ID, per.ID)
}

func releaseClaim(t *testing.T, c *database.Catalog, userID, claimID []byte) ([]rowchange.Change, error) {
	t.Helper()
	_ = userID
	db, err := c.DB()
	if err != nil {
		return nil, err
	}
	sqlTx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	changes, err := deleteimpact.ReleaseClaim(&database.Tx{Tx: sqlTx}, claimID)
	if err != nil {
		_ = sqlTx.Rollback()
		return nil, err
	}
	if err := sqlTx.Commit(); err != nil {
		return nil, err
	}
	return changes, nil
}

func releaseAndAssertHandleGone(t *testing.T, c *database.Catalog, userID, claimID, entityID []byte) {
	t.Helper()
	changes, err := releaseClaim(t, c, userID, claimID)
	if err != nil {
		t.Fatal(err)
	}
	var sawEntity bool
	for _, ch := range changes {
		if ch.EntityType == "canonical_entity" && ch.Action == rowchange.ActionDelete {
			sawEntity = true
		}
	}
	if !sawEntity {
		t.Fatalf("changes %v", changes)
	}
	if _, err := canonicalentities.Get(c, entityID); err == nil {
		t.Fatal("handle should be gone")
	}
}

func mustSource(t *testing.T, c *database.Catalog, userID []byte) sources.Source {
	t.Helper()
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
	_, err = artifactstest.Create(c, userID, artifacts.CreateInput{SourceID: src.ID, Label: "Scan"})
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func mustSubject(t *testing.T, c *database.Catalog, userID, sourceID, typeID []byte, label string) subjects.Subject {
	t.Helper()
	graph := evidencegraphstest.SoleGraph(t, c, sourceID).ID
	return subjectstest.OfType(t, c, userID, graph, typeID, label)
}

func mustEntity(t *testing.T, c *database.Catalog, userID, typeID []byte) canonicalentities.Entity {
	t.Helper()
	e, err := writes.Call(c, writes.Op{Action: "create_canonical_entity", UserID: userID}, func(tx *database.Tx) (canonicalentities.Entity, []rowchange.Change, error) {
		return canonicalentities.Create(tx, userID, canonicalentities.CreateInput{SubjectTypeID: typeID})
	})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func mustClaim(t *testing.T, c *database.Catalog, userID []byte, subject subjects.Subject, entity canonicalentities.Entity) identityclaims.Claim {
	t.Helper()
	cl, err := writes.Call(c, writes.Op{Action: "create_identity_claim", UserID: userID}, func(tx *database.Tx) (identityclaims.Claim, []rowchange.Change, error) {
		return identityclaims.Create(tx, userID, identityclaims.CreateInput{
			SubjectID: subject.ID, EntityID: entity.ID, Status: identityclaims.StatusAccepted,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	return cl
}
