package namealiases

import (
	"database/sql"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/namevalues"
	"github.com/mendahu/provenencia/core/match"
)

func TestInstallAndLoad(t *testing.T) {
	c := newCatalog(t)
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	db := mustDB(t, c)
	tables, err := Load(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 2 {
		t.Fatalf("tables %d", len(tables))
	}
	abbrev := tableByKey(t, tables, "abbreviations")
	if !hasSet(abbrev, "jas", "james") {
		t.Fatalf("abbreviations missing jas/james: %+v", abbrev.Sets)
	}
	nicks := tableByKey(t, tables, "nicknames")
	if !hasSet(nicks, "jim", "james") {
		t.Fatalf("nicknames missing jim/james: %+v", nicks.Sets)
	}

	if _, err := db.Exec(`DELETE FROM name_alias_members WHERE token = 'jas'`); err != nil {
		t.Fatal(err)
	}
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	tables, err = Load(db)
	if err != nil {
		t.Fatal(err)
	}
	if hasSet(tableByKey(t, tables, "abbreviations"), "jas", "james") {
		t.Fatal("second install restored a deleted token")
	}
}

func TestWarmHoldsAliasTablesUntilInvalidated(t *testing.T) {
	c := newCatalog(t)
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	if err := Warm(c); err != nil {
		t.Fatal(err)
	}
	db := mustDB(t, c)
	if _, err := db.Exec(`DELETE FROM name_alias_members WHERE token = 'jas'`); err != nil {
		t.Fatal(err)
	}
	cmp, err := Comparer(c.Graph(), db, match.DefaultNames)
	if err != nil {
		t.Fatal(err)
	}
	assertScore(t, cmp, "Jas.", "James", match.AliasCredit)

	if err := Invalidate(c); err != nil {
		t.Fatal(err)
	}
	cmp, err = Comparer(c.Graph(), db, match.DefaultNames)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := cmp.Compare(given("Jas."), given("James"))
	if ok && math.Abs(got-match.AliasCredit) < 1e-9 {
		t.Fatal("invalidate left the deleted token in the cache")
	}

	insertTable(t, db, "xyzzy", "user", "nickname", [][]string{{"xyzzy", "plugh"}})
	cmp, err = Comparer(c.Graph(), db, match.DefaultNames)
	if err != nil {
		t.Fatal(err)
	}
	got, ok = cmp.Compare(given("Plugh"), given("Xyzzy"))
	if ok && math.Abs(got-match.AliasCredit) < 1e-9 {
		t.Fatal("a new table scored before invalidate")
	}

	if err := Invalidate(c); err != nil {
		t.Fatal(err)
	}
	cmp, err = Comparer(c.Graph(), db, match.DefaultNames)
	if err != nil {
		t.Fatal(err)
	}
	assertScore(t, cmp, "Plugh", "Xyzzy", match.AliasCredit)
}

func TestEmptyCatalogScoresNoAliases(t *testing.T) {
	c := newCatalog(t)
	db := mustDB(t, c)
	tables, err := Load(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 0 {
		t.Fatalf("empty catalog loaded %d tables", len(tables))
	}
	cmp, err := Comparer(c.Graph(), db, match.DefaultNames)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := cmp.Compare(given("Jas."), given("James"))
	if !ok || math.Abs(got-match.AliasCredit) < 1e-9 {
		t.Fatalf("empty catalog still aliased: %.4f %v", got, ok)
	}
}

func TestUserAndPluginRowsScore(t *testing.T) {
	c := newCatalog(t)
	if err := Install(c); err != nil {
		t.Fatal(err)
	}
	db := mustDB(t, c)
	insertTable(t, db, "xyzzy", "user", "nickname", [][]string{{"xyzzy", "plugh"}})
	insertTable(t, db, "pepe", "plugin:example", "nickname", [][]string{{"josé", "pepe"}})
	cmp, err := Comparer(c.Graph(), db, match.DefaultNames)
	if err != nil {
		t.Fatal(err)
	}
	assertScore(t, cmp, "Plugh", "Xyzzy", match.AliasCredit)
	assertScore(t, cmp, "Pepe", "José", match.AliasCredit)
	assertScore(t, cmp, "Jas.", "James", match.AliasCredit)
	assertScore(t, cmp, "Jim", "James", match.AliasCredit)
}

func newCatalog(t *testing.T) *database.Catalog {
	t.Helper()
	c, err := database.Create(t.TempDir(), "t.provenencia")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func mustDB(t *testing.T, c *database.Catalog) *sql.DB {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func insertTable(t *testing.T, db *sql.DB, key, origin, kind string, sets [][]string) {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO name_alias_tables (id, key, origin, label, kind) VALUES (?, ?, ?, ?, ?)`,
		id[:], key, origin, key, kind); err != nil {
		t.Fatal(err)
	}
	for i, set := range sets {
		for _, token := range set {
			if _, err := db.Exec(`INSERT INTO name_alias_members (table_id, set_idx, token) VALUES (?, ?, ?)`,
				id[:], i, token); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, role := range []string{"given", "nick", "untyped"} {
		if _, err := db.Exec(`INSERT INTO name_alias_roles (table_id, role) VALUES (?, ?)`, id[:], role); err != nil {
			t.Fatal(err)
		}
	}
}

func tableByKey(t *testing.T, tables []match.AliasTable, key string) match.AliasTable {
	t.Helper()
	for _, table := range tables {
		if table.Key == key {
			return table
		}
	}
	t.Fatalf("no table %q", key)
	return match.AliasTable{}
}

func hasSet(table match.AliasTable, a, b string) bool {
	for _, set := range table.Sets {
		foundA, foundB := false, false
		for _, token := range set {
			if token == a {
				foundA = true
			}
			if token == b {
				foundB = true
			}
		}
		if foundA && foundB {
			return true
		}
	}
	return false
}

func assertScore(t *testing.T, cmp match.NameComparer, a, b string, want float64) {
	t.Helper()
	got, ok := cmp.Compare(given(a), given(b))
	if !ok || math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s vs %s: got %.4f %v, want %.4f", a, b, got, ok, want)
	}
}

func given(text string) match.Value {
	return match.Value{Name: &namevalues.Value{Parts: []namevalues.Part{{
		Value: text, Type: namevalues.PartTypeGiven,
	}}}}
}
