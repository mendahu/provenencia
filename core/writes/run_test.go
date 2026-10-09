package writes_test

import (
	"bytes"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/properties"
	"github.com/mendahu/provenencia/core/database/propertyterms"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/writes"
)

func TestRunRollsBack(t *testing.T) {
	c := newCatalog(t)
	id := mustID(t)
	userRef := mustRef(t)
	ran := false
	_, _, err := writes.Run(c, writes.Op{Action: "test"}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		tx.AfterCommit(func() { ran = true })
		if _, err := tx.Exec(`INSERT INTO users (id, display_name, ref) VALUES (?, ?, ?)`, id, "Ada", userRef); err != nil {
			return struct{}{}, nil, err
		}
		return struct{}{}, nil, errors.New("fail")
	})
	if err == nil || err.Error() != "fail" {
		t.Fatalf("err %v", err)
	}
	if ran {
		t.Fatal("AfterCommit ran")
	}
	if _, err := users.Lookup(c, id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("user %v", err)
	}
	if n := auditCount(t, c); n != 0 {
		t.Fatalf("audit rows %d", n)
	}
}

func TestRunSkipsEmptyChanges(t *testing.T) {
	c := newCatalog(t)
	ran := false
	got, res, err := writes.Run(c, writes.Op{Action: "test"}, func(tx *database.Tx) (string, []rowchange.Change, error) {
		tx.AfterCommit(func() { ran = true })
		return "kept", nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "kept" || res.Revision != 0 || ran {
		t.Fatalf("got %q rev %d after %v", got, res.Revision, ran)
	}
	if n := auditCount(t, c); n != 0 {
		t.Fatalf("audit rows %d", n)
	}
}

func TestRunAfterCommitAndReentry(t *testing.T) {
	c := newCatalog(t)
	var inner error
	got, res, err := writes.Run(c, writes.Op{Action: "test"}, func(tx *database.Tx) (int, []rowchange.Change, error) {
		tx.AfterCommit(func() {
			_, _, inner = writes.Run(c, writes.Op{Action: "test"}, func(tx *database.Tx) (int, []rowchange.Change, error) {
				return 0, oneChange(t), nil
			})
		})
		return 7, oneChange(t), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 || res.Revision == 0 {
		t.Fatalf("got %d rev %d", got, res.Revision)
	}
	if !errors.Is(inner, database.ErrWriteReentry) {
		t.Fatalf("inner %v", inner)
	}
}

func TestRunRefusesReentry(t *testing.T) {
	c := newCatalog(t)
	var inner error
	_, _, err := writes.Run(c, writes.Op{Action: "test"}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		_, _, inner = writes.Run(c, writes.Op{Action: "test"}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
			return struct{}{}, nil, nil
		})
		return struct{}{}, nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(inner, database.ErrWriteReentry) {
		t.Fatalf("inner %v", inner)
	}
}

func TestRunDropsListeners(t *testing.T) {
	c := newCatalog(t)
	ok := &recListener{}
	bad := &recListener{fail: errors.New("nope")}
	c.Listen(ok)
	c.Listen(bad)
	got, res, err := writes.Run(c, writes.Op{Action: "test"}, func(tx *database.Tx) (string, []rowchange.Change, error) {
		return "saved", oneChange(t), nil
	})
	if err == nil || err.Error() != "nope" {
		t.Fatalf("err %v", err)
	}
	if got != "saved" || res.Revision == 0 {
		t.Fatalf("got %q rev %d", got, res.Revision)
	}
	if ok.notices != 1 || bad.notices != 1 || ok.dropped != 1 || bad.dropped != 1 {
		t.Fatalf("ok %+v bad %+v", ok, bad)
	}
	if n := auditCount(t, c); n != 1 {
		t.Fatalf("audit rows %d", n)
	}
}

func TestPropertyTermRun(t *testing.T) {
	c := newCatalog(t)
	userID := mustID(t)
	if err := users.Upsert(c, userID, "Tester", mustRef(t)); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	prop, err := properties.Lookup(c, "event_type", properties.OriginProvenencia)
	if err != nil {
		t.Fatal(err)
	}
	before := searchCount(t, c)

	term, res, err := writes.Run(c, writes.Op{Action: "create_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Create(tx, userID, prop.ID, "Parish Register", "")
		})
	if err != nil {
		t.Fatal(err)
	}
	assertTermEffects(t, res)
	if searchCount(t, c) != before {
		t.Fatalf("search docs %d, was %d", searchCount(t, c), before)
	}
	action, scopes := auditAction(t, c, res.Revision)
	if action != "create_property_term" || scopes != 0 {
		t.Fatalf("action %q scopes %d", action, scopes)
	}

	same, quiet, err := writes.Run(c, writes.Op{Action: "update_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Update(tx, userID, term.ID, "Parish Register", "")
		})
	if err != nil {
		t.Fatal(err)
	}
	if quiet.Revision != 0 || !bytes.Equal(same.ID, term.ID) {
		t.Fatalf("unchanged rev %d", quiet.Revision)
	}

	edited, res, err := writes.Run(c, writes.Op{Action: "update_property_term", UserID: userID},
		func(tx *database.Tx) (propertyterms.Term, []rowchange.Change, error) {
			return propertyterms.Update(tx, userID, term.ID, "Parish Book", "")
		})
	if err != nil {
		t.Fatal(err)
	}
	if edited.Label != "Parish Book" {
		t.Fatalf("label %q", edited.Label)
	}
	assertTermEffects(t, res)
	action, scopes = auditAction(t, c, res.Revision)
	if action != "update_property_term" || scopes != 0 {
		t.Fatalf("action %q scopes %d", action, scopes)
	}
	if searchCount(t, c) != before {
		t.Fatalf("search docs %d, was %d", searchCount(t, c), before)
	}
}

func assertTermEffects(t *testing.T, res writes.Result) {
	t.Helper()
	if res.Revision == 0 || !res.Effects.Vocabulary || res.Effects.Structure || len(res.Effects.Handles) != 0 || len(res.Effects.Search) != 0 {
		t.Fatalf("rev %d effects %+v", res.Revision, res.Effects)
	}
}

type recListener struct {
	fail    error
	notices int
	dropped int
}

func (r *recListener) OnCommit(int64, effects.Set) error {
	r.notices++
	return r.fail
}

func (r *recListener) Drop() { r.dropped++ }

func newCatalog(t *testing.T) *database.Catalog {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func oneChange(t *testing.T) []rowchange.Change {
	t.Helper()
	return []rowchange.Change{{
		EntityType: "property_term",
		EntityID:   mustID(t),
		Action:     rowchange.ActionCreate,
		Fields:     map[string]rowchange.FieldDiff{"label": {New: "x"}},
	}}
}

func mustID(t *testing.T) []byte {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return id[:]
}

func mustRef(t *testing.T) string {
	t.Helper()
	r, err := ref.Mint(ref.PrefixUser)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func auditCount(t *testing.T, c *database.Catalog) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transactions`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func auditAction(t *testing.T, c *database.Catalog, revision int64) (string, int) {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var action string
	var id []byte
	if err := db.QueryRow(`SELECT id, action_type FROM audit_transactions WHERE revision = ?`, revision).Scan(&id, &action); err != nil {
		t.Fatal(err)
	}
	var scopes int
	if err := db.QueryRow(`SELECT COUNT(*) FROM audit_transaction_scopes WHERE audit_transaction_id = ?`, id).Scan(&scopes); err != nil {
		t.Fatal(err)
	}
	return action, scopes
}

func searchCount(t *testing.T, c *database.Catalog) int {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM catalog_search_docs`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
