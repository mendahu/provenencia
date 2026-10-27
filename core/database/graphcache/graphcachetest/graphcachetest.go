// Package graphcachetest builds catalogs and canonical graphs for tests of the
// graph cache and of the packages that read it.
package graphcachetest

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/autoreconciler"
	"github.com/mendahu/provenencia/core/database/databasetest" // registers graphcache.SetTruth for verify mode
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/sourcetypes"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"
	"github.com/mendahu/provenencia/core/vocabkey"
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

// Canon writes canonical entities and what the graph cache reads of them, and
// commits them the way a write would. Each value is backed by an Observation
// on a Subject that is an accepted member of its entity, and the reconciler
// writes the kept rows from them, so Verify (PROVENENCIA_GRAPH_VERIFY=1) finds
// the cache equal to the catalog. A Subject has the id of its entity. Rows go
// in directly, not through a write, so nothing reaches the graph until Commit.
type Canon struct {
	t        testing.TB
	c        *database.Catalog
	db       *sql.DB
	n        int
	rev      int64
	citation []byte
	graph    []byte
	source   []byte
}

// NewCanon returns a builder over c, with one Source, Artifact, Citation and
// Evidence graph behind every Observation it writes.
func NewCanon(t testing.TB, c *database.Catalog) *Canon {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	k := &Canon{t: t, c: c, db: db}
	typeID, err := sourcetypes.Upsert(c, sourcetypes.Type{Key: "book", Origin: vocabkey.OriginProvenencia, Label: "Book"})
	if err != nil {
		t.Fatal(err)
	}
	k.source, k.graph, k.citation = k.next(5), k.next(6), k.next(7)
	artifact := k.next(8)
	k.exec(`INSERT INTO sources (id, ref, source_type_id, title) VALUES (?, 'SRC-TST', ?, 'Canon')`, k.source, typeID)
	k.exec(`INSERT INTO artifacts (id, ref, source_id, label) VALUES (?, 'ART-TST', ?, 'Scan')`, artifact, k.source)
	k.exec(`INSERT INTO citations (id, ref, artifact_id, locator_json) VALUES (?, 'CIT-TST', ?, '{"version":1,"selectors":[{"type":"page","artifact_page":1}]}')`, k.citation, artifact)
	k.exec(`INSERT INTO evidence_graphs (id, ref, source_id, name) VALUES (?, 'EVG-TST', ?, 'Canon')`, k.graph, k.source)
	return k
}

// next is the next id of one table family: {family, 1, 0, ..., counter}.
func (k *Canon) next(family byte) []byte {
	k.n++
	id := make([]byte, 16)
	id[0], id[1], id[14], id[15] = family, 1, byte(k.n>>8), byte(k.n)
	return id
}

func (k *Canon) exec(query string, args ...any) {
	k.t.Helper()
	if _, err := k.db.Exec(query, args...); err != nil {
		k.t.Fatal(err)
	}
}

// SubjectTypeID returns the id of the provenencia subject type named key.
func SubjectTypeID(t testing.TB, c *database.Catalog, key string) []byte {
	t.Helper()
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var typeID []byte
	if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = ? AND origin = 'provenencia'`, key).Scan(&typeID); err != nil {
		t.Fatal(err)
	}
	return typeID
}

// AssertKindsLoaded fails when loading any of the provenencia subject types
// named by keys onto the graph of c issues a SELECT.
func AssertKindsLoaded(t testing.TB, c *database.Catalog, keys ...string) {
	t.Helper()
	selects := databasetest.CountSelects(t, c)
	for _, key := range keys {
		typeID := SubjectTypeID(t, c, key)
		before := selects.N()
		if _, err := c.Graph().Kind(typeID); err != nil {
			t.Fatal(err)
		}
		if n := selects.N() - before; n != 0 {
			t.Errorf("Kind(%s) issued %d SELECTs, want 0", key, n)
		}
	}
}

// Entity adds a canonical entity of a provenencia subject type, with a Subject
// that is its accepted member, and returns its id. Ids and refs come from a
// counter.
func (k *Canon) Entity(kind string) []byte {
	k.t.Helper()
	typeID := SubjectTypeID(k.t, k.c, kind)
	id := k.next(3)
	ref := fmt.Sprintf("TST-%d", k.n)
	k.exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, ?)`, id, typeID, ref)
	k.exec(`INSERT INTO subjects (id, ref, source_id, graph_id, subject_type_id) VALUES (?, ?, ?, ?, ?)`, id, "SUB-"+ref, k.source, k.graph, typeID)
	k.exec(`INSERT INTO identity_claims (id, subject_id, entity_id, subject_type_id, status) VALUES (?, ?, ?, ?, 'accepted')`,
		k.next(9), id, id, typeID)
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

// observe files one Observation on an entity's Subject and rewrites the
// entity's kept rows from its Observations.
func (k *Canon) observe(entity []byte, propKey, column string, value []byte) {
	k.t.Helper()
	k.exec(`INSERT INTO observations (id, ref, citation_id, subject_id, property_id, `+column+`) VALUES (?, ?, ?, ?, ?, ?)`,
		k.next(4), fmt.Sprintf("OBS-TST-%d", k.n), k.citation, entity, k.propID(propKey), value)
	if err := autoreconciler.RecomputeTx(k.db, [][]byte{entity}); err != nil {
		k.t.Fatal(err)
	}
}

// Term sets the kept term of one property of an entity to a term of the
// catalog: a seeded one, or one the test added.
func (k *Canon) Term(entity []byte, propKey, termKey string) {
	k.t.Helper()
	var termID []byte
	if err := k.db.QueryRow(`SELECT id FROM property_terms WHERE property_id = ? AND key = ? ORDER BY origin = 'provenencia' DESC`,
		k.propID(propKey), termKey).Scan(&termID); err != nil {
		k.t.Fatal(err)
	}
	k.observe(entity, propKey, "value_term_id", termID)
}

// Link adds an association of a bridge subject type. ends maps each endpoint
// property key (person, event, related_to) to the entity it joins; the term is
// the association's disambiguating term (role, relationship_type).
func (k *Canon) Link(kind string, ends map[string][]byte, termProp, termKey string) []byte {
	k.t.Helper()
	id := k.Entity(kind)
	for prop, end := range ends {
		k.observe(id, prop, "value_subject_id", end)
	}
	k.Term(id, termProp, termKey)
	return id
}

// Remove deletes an entity, its Subject and what they hold.
func (k *Canon) Remove(id []byte) {
	k.t.Helper()
	k.exec(`DELETE FROM observations WHERE subject_id = ?`, id)
	k.exec(`DELETE FROM identity_claims WHERE entity_id = ?`, id)
	k.exec(`DELETE FROM subjects WHERE id = ?`, id)
	k.exec(`DELETE FROM canonical_entities WHERE id = ?`, id)
}

// Retype replaces the kept term of one property of an entity.
func (k *Canon) Retype(entity []byte, propKey, termKey string) {
	k.t.Helper()
	k.exec(`DELETE FROM observations WHERE subject_id = ? AND property_id = ?`, entity, k.propID(propKey))
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
