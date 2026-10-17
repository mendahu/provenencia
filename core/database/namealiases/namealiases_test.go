package namealiases

import (
	"database/sql"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/matchregistry"
	"github.com/mendahu/provenencia/core/namevalue"
	"github.com/mendahu/provenencia/core/propertymeta"
	"github.com/mendahu/provenencia/core/valuecompare"
	"github.com/mendahu/provenencia/core/vocabkey"
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
	rules, err := Rules(c.Graph(), db, matchregistry.Default().Value)
	if err != nil {
		t.Fatal(err)
	}
	assertScore(t, rules, "Jas.", "James", matchregistry.Default().Value.Words.AliasCredit)

	if err := Invalidate(c); err != nil {
		t.Fatal(err)
	}
	rules, err = Rules(c.Graph(), db, matchregistry.Default().Value)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := compareGiven(rules, "Jas.", "James")
	if ok && math.Abs(got-matchregistry.Default().Value.Words.AliasCredit) < 1e-9 {
		t.Fatal("invalidate left the deleted token in the cache")
	}

	insertTable(t, db, "xyzzy", "user", "nickname", [][]string{{"xyzzy", "plugh"}})
	rules, err = Rules(c.Graph(), db, matchregistry.Default().Value)
	if err != nil {
		t.Fatal(err)
	}
	got, ok = compareGiven(rules, "Plugh", "Xyzzy")
	if ok && math.Abs(got-matchregistry.Default().Value.Words.AliasCredit) < 1e-9 {
		t.Fatal("a new table scored before invalidate")
	}

	if err := Invalidate(c); err != nil {
		t.Fatal(err)
	}
	rules, err = Rules(c.Graph(), db, matchregistry.Default().Value)
	if err != nil {
		t.Fatal(err)
	}
	assertScore(t, rules, "Plugh", "Xyzzy", matchregistry.Default().Value.Words.AliasCredit)
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
	rules, err := Rules(c.Graph(), db, matchregistry.Default().Value)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := compareGiven(rules, "Jas.", "James")
	if !ok || math.Abs(got-matchregistry.Default().Value.Words.AliasCredit) < 1e-9 {
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
	rules, err := Rules(c.Graph(), db, matchregistry.Default().Value)
	if err != nil {
		t.Fatal(err)
	}
	assertScore(t, rules, "Plugh", "Xyzzy", matchregistry.Default().Value.Words.AliasCredit)
	assertScore(t, rules, "Pepe", "José", matchregistry.Default().Value.Words.AliasCredit)
	assertScore(t, rules, "Jas.", "James", matchregistry.Default().Value.Words.AliasCredit)
	assertScore(t, rules, "Jim", "James", matchregistry.Default().Value.Words.AliasCredit)
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

func tableByKey(t *testing.T, tables []valuecompare.AliasTable, key string) valuecompare.AliasTable {
	t.Helper()
	for _, table := range tables {
		if table.Key == key {
			return table
		}
	}
	t.Fatalf("no table %q", key)
	return valuecompare.AliasTable{}
}

func hasSet(table valuecompare.AliasTable, a, b string) bool {
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

// compareGiven is the similarity of two given names under rules, and whether
// they are comparable at all.
func compareGiven(rules valuecompare.Rules, a, b string) (float64, bool) {
	res, _ := valuecompare.Compare(given(a), given(b), rules)
	return res.Similarity, res.Relation != valuecompare.Unknown
}

func assertScore(t *testing.T, rules valuecompare.Rules, a, b string, want float64) {
	t.Helper()
	got, ok := compareGiven(rules, a, b)
	if !ok || math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s vs %s: got %.4f %v, want %.4f", a, b, got, ok, want)
	}
}

func TestAddProductNicknamesFillsAnEmptyCatalog(t *testing.T) {
	c := newCatalog(t)
	added, err := AddProductNicknames(c)
	if err != nil {
		t.Fatal(err)
	}
	if added.Sets < 1000 || added.Tokens < added.Sets {
		t.Fatalf("added %+v", added)
	}
	again, err := AddProductNicknames(c)
	if err != nil {
		t.Fatal(err)
	}
	if again.Sets != 0 || again.Tokens != 0 {
		t.Fatalf("second add %+v", again)
	}
	tables, err := Load(mustDB(t, c))
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 1 || tables[0].Key != "nicknames" || tables[0].Origin != vocabkey.OriginProvenencia {
		t.Fatalf("tables %+v", tables)
	}
	if !hasSet(tables[0], "chris", "christine") || hasSet(tables[0], "tina", "christopher") {
		t.Fatal("christine and christopher were joined or chris was missed")
	}
}

func TestAddProductNicknamesGrowsTheShortList(t *testing.T) {
	c := newCatalog(t)
	db := mustDB(t, c)
	insertTable(t, db, "nicknames", vocabkey.OriginProvenencia, "nickname", [][]string{{"james", "jim", "jimmy"}})
	insertTable(t, db, "nicknames", "user", "nickname", [][]string{{"xyzzy", "plugh"}})

	added, err := AddProductNicknames(c)
	if err != nil {
		t.Fatal(err)
	}
	if added.Tokens == 0 {
		t.Fatal("short list gained nothing")
	}
	tables, err := Load(db)
	if err != nil {
		t.Fatal(err)
	}
	product := tableByKey(t, tables, "nicknames")
	// Load orders by key, origin. Both tables share the key. Take provenencia.
	for _, table := range tables {
		if table.Key == "nicknames" && table.Origin == vocabkey.OriginProvenencia {
			product = table
		}
	}
	if setsWithAll(product, "james", "jim", "jimmy") != 1 || !hasSet(product, "james", "jamie") {
		t.Fatalf("james set was duplicated or not grown")
	}
	if !hasSet(product, "chris", "christine") || hasSet(product, "tina", "christopher") {
		t.Fatal("shared nickname joined two formal names")
	}
	user := valuecompare.AliasTable{}
	for _, table := range tables {
		if table.Origin == "user" {
			user = table
		}
	}
	if len(user.Sets) != 1 || !hasSet(user, "xyzzy", "plugh") {
		t.Fatalf("user table changed: %+v", user.Sets)
	}
}

func setsWithAll(table valuecompare.AliasTable, tokens ...string) int {
	n := 0
	for _, set := range table.Sets {
		have := map[string]bool{}
		for _, tok := range set {
			have[tok] = true
		}
		ok := true
		for _, token := range tokens {
			if !have[token] {
				ok = false
				break
			}
		}
		if ok {
			n++
		}
	}
	return n
}

func given(text string) valuecompare.Value {
	return valuecompare.Value{Type: propertymeta.ValueTypeName, Name: &namevalue.Value{Parts: []namevalue.Part{{
		Value: text, Type: namevalue.PartTypeGiven,
	}}}}
}
