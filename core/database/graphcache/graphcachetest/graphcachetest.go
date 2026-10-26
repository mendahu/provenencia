// Package graphcachetest builds catalogs and canonical graphs for tests of the
// graph cache and of the packages that read it.
package graphcachetest

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	_ "github.com/mendahu/provenencia/core/database/autoreconciler" // registers graphcache.SetTruth for verify mode
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
)

// OpenCatalog creates a catalog in a temporary directory with one user and the
// seeded subject vocabulary. It closes with the test.
func OpenCatalog(t testing.TB) *database.Catalog {
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
	if err := users.Upsert(c, []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}, "Tester", r); err != nil {
		t.Fatal(err)
	}
	if err := subjectvocab.Install(c); err != nil {
		t.Fatal(err)
	}
	return c
}

// Canon writes canonical entities and the rows the graph cache reads, and
// commits them the way a write would.
type Canon struct {
	t   testing.TB
	c   *database.Catalog
	db  *sql.DB
	n   int
	rev int64
}

// NewCanon returns a builder over c.
func NewCanon(t testing.TB, c *database.Catalog) *Canon {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	return &Canon{t: t, c: c, db: db}
}

// Entity adds a canonical entity of a provenencia subject type and returns its
// id. Ids and refs come from a counter.
func (k *Canon) Entity(kind string) []byte {
	k.t.Helper()
	var typeID []byte
	if err := k.db.QueryRow(`SELECT id FROM subject_types WHERE key = ? AND origin = 'provenencia'`, kind).Scan(&typeID); err != nil {
		k.t.Fatal(err)
	}
	k.n++
	id := make([]byte, 16)
	id[0], id[1], id[14], id[15] = 3, 1, byte(k.n>>8), byte(k.n)
	if _, err := k.db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, ?)`,
		id, typeID, fmt.Sprintf("TST-%d", k.n)); err != nil {
		k.t.Fatal(err)
	}
	return id
}

func (k *Canon) propID(key string) []byte {
	k.t.Helper()
	var id []byte
	if err := k.db.QueryRow(`SELECT id FROM properties WHERE key = ? AND origin = 'provenencia'`, key).Scan(&id); err != nil {
		k.t.Fatal(err)
	}
	return id
}

const insertValue = `INSERT INTO auto_reconciler_values
	(entity_id, property_id, rank, %s, support, against, reason)
	VALUES (?, ?, 1, ?, 1, 0, 'kept')`

// Term sets the rank-1 kept term of one property of an entity to a seeded term.
func (k *Canon) Term(entity []byte, propKey, termKey string) {
	k.t.Helper()
	prop := k.propID(propKey)
	var termID []byte
	if err := k.db.QueryRow(`SELECT id FROM property_terms WHERE property_id = ? AND key = ? AND origin = 'provenencia'`,
		prop, termKey).Scan(&termID); err != nil {
		k.t.Fatal(err)
	}
	if _, err := k.db.Exec(fmt.Sprintf(insertValue, "value_term_id"), entity, prop, termID); err != nil {
		k.t.Fatal(err)
	}
}

// Link adds an association of a bridge subject type. ends maps each endpoint
// property key (person, event, related_to) to the entity it joins; the term is
// the association's disambiguating term (role, relationship_type).
func (k *Canon) Link(kind string, ends map[string][]byte, termProp, termKey string) []byte {
	k.t.Helper()
	id := k.Entity(kind)
	for prop, end := range ends {
		if _, err := k.db.Exec(fmt.Sprintf(insertValue, "value_entity_id"), id, k.propID(prop), end); err != nil {
			k.t.Fatal(err)
		}
	}
	k.Term(id, termProp, termKey)
	return id
}

// Remove deletes an entity's values and row.
func (k *Canon) Remove(id []byte) {
	k.t.Helper()
	if _, err := k.db.Exec(`DELETE FROM auto_reconciler_values WHERE entity_id = ?`, id); err != nil {
		k.t.Fatal(err)
	}
	if _, err := k.db.Exec(`DELETE FROM canonical_entities WHERE id = ?`, id); err != nil {
		k.t.Fatal(err)
	}
}

// Retype replaces the kept term of one property of an entity.
func (k *Canon) Retype(entity []byte, propKey, termKey string) {
	k.t.Helper()
	if _, err := k.db.Exec(`DELETE FROM auto_reconciler_values WHERE entity_id = ? AND property_id = ?`, entity, k.propID(propKey)); err != nil {
		k.t.Fatal(err)
	}
	k.Term(entity, propKey, termKey)
}

// Commit tells the graph that ids were written, as a commit's effects do.
func (k *Canon) Commit(ids ...[]byte) {
	k.t.Helper()
	k.commit(effects.Set{Handles: ids})
}

// CommitVocabulary tells the graph that the vocabulary was written.
func (k *Canon) CommitVocabulary() {
	k.t.Helper()
	k.commit(effects.Set{Lists: []effects.List{effects.ListVocabulary}})
}

func (k *Canon) commit(set effects.Set) {
	k.t.Helper()
	k.rev++
	if err := k.c.Graph().OnCommit(effects.Commit{Activity: effects.Activity{Revision: k.rev}, Effects: set}); err != nil {
		k.t.Fatal(err)
	}
}
