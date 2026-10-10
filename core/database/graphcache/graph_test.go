package graphcache_test

import (
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/subjectvocab"
	"github.com/mendahu/provenencia/core/database/users"
	"github.com/mendahu/provenencia/core/ref"

	_ "github.com/mendahu/provenencia/core/database/autoreconciler"
)

func TestNodeAndKindAndVerify(t *testing.T) {
	c := openCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var typeID []byte
	if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = 'person' AND origin = 'provenencia'`).Scan(&typeID); err != nil {
		t.Fatal(err)
	}
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, 'PER-1')`, id, typeID); err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	if g.LoadedForTest(id) {
		t.Fatal("a node loaded before anyone read it")
	}
	n, err := g.Node(id)
	if err != nil {
		t.Fatal(err)
	}
	if n == nil || n.Ref != "PER-1" || n.Merged {
		t.Fatalf("node %+v", n)
	}
	nodes, err := g.Kind(typeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Ref != "PER-1" {
		t.Fatalf("kind %d", len(nodes))
	}
	if err := g.Verify(); err != nil {
		t.Fatal(err)
	}
	g.CorruptRefForTest(id)
	if err := g.Verify(); err == nil {
		t.Fatal("Verify accepted a node whose ref does not match the catalog")
	}

	g.Drop()
	if g.LoadedForTest(id) {
		t.Fatal("Drop left a node")
	}
	if _, err := g.Node(id); err != nil {
		t.Fatal(err)
	}
	if err := g.OnCommit(1, effects.Set{Structure: true}); err != nil {
		t.Fatal(err)
	}
	if g.LoadedForTest(id) {
		t.Fatal("a structure commit left a node")
	}
}

func TestCommitInsertsANewHandleIntoALoadedKind(t *testing.T) {
	c := openCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var typeID []byte
	if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = 'person' AND origin = 'provenencia'`).Scan(&typeID); err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	if _, err := g.Kind(typeID); err != nil {
		t.Fatal(err)
	}
	id := []byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, 'PER-2')`, id, typeID); err != nil {
		t.Fatal(err)
	}
	if err := g.OnCommit(1, effects.Set{Handles: [][]byte{id}}); err != nil {
		t.Fatal(err)
	}
	if !g.LoadedForTest(id) {
		t.Fatal("a new handle of a loaded kind was not inserted")
	}
	nodes, err := g.Kind(typeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].Ref != "PER-2" {
		t.Fatalf("kind %+v", nodes)
	}
}

func TestCommitOfAnAssociationLoadsEndpoints(t *testing.T) {
	c := openCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	typeID := func(key string) []byte {
		t.Helper()
		var id []byte
		if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = ? AND origin = 'provenencia'`, key).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	propID := func(key string) []byte {
		t.Helper()
		var id []byte
		if err := db.QueryRow(`SELECT id FROM properties WHERE key = ? AND origin = 'provenencia'`, key).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	personT, eventT, partT := typeID("person"), typeID("event"), typeID("participation")
	person := []byte{3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	event := []byte{3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	assoc := []byte{3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
	if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, 'PER-A'), (?, ?, 'EVT-A'), (?, ?, 'BRG-A')`,
		person, personT, event, eventT, assoc, partT); err != nil {
		t.Fatal(err)
	}
	ins := `INSERT INTO auto_reconciler_values
		(entity_id, property_id, rank, value_entity_id, support, against, reason)
		VALUES (?, ?, 1, ?, 1, 0, 'kept')`
	if _, err := db.Exec(ins, assoc, propID("person"), person); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ins, assoc, propID("event"), event); err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	if err := g.OnCommit(1, effects.Set{Handles: [][]byte{assoc}}); err != nil {
		t.Fatal(err)
	}
	if !g.LoadedForTest(person) || !g.LoadedForTest(event) {
		t.Fatal("an association commit did not load its endpoints")
	}
	n, err := g.Node(person)
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Links) != 1 || string(n.Links[0].Neighbor) != string(event) || !n.Links[0].FromEnd {
		t.Fatalf("links %+v", n.Links)
	}
}

func openCatalog(t *testing.T) *database.Catalog {
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
