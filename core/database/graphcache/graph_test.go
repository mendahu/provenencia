package graphcache_test

import (
	"bytes"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
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
	if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1}, Effects: effects.Set{Structure: true}}); err != nil {
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
	if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1}, Effects: effects.Set{Handles: [][]byte{id}}}); err != nil {
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
	if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1}, Effects: effects.Set{Handles: [][]byte{assoc}}}); err != nil {
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

func TestCommitReloadsASharedEndpointOnce(t *testing.T) {
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
		return append([]byte(nil), id...)
	}
	propID := func(key string) []byte {
		t.Helper()
		var id []byte
		if err := db.QueryRow(`SELECT id FROM properties WHERE key = ? AND origin = 'provenencia'`, key).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return append([]byte(nil), id...)
	}
	personT, eventT, partT := typeID("person"), typeID("event"), typeID("participation")
	person := []byte{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	eventA := []byte{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	eventB := []byte{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
	assocA := []byte{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}
	assocB := []byte{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5}
	if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES
		(?, ?, 'PER-S'), (?, ?, 'EVT-A'), (?, ?, 'EVT-B'), (?, ?, 'BRG-A'), (?, ?, 'BRG-B')`,
		person, personT, eventA, eventT, eventB, eventT, assocA, partT, assocB, partT); err != nil {
		t.Fatal(err)
	}
	ins := `INSERT INTO auto_reconciler_values
		(entity_id, property_id, rank, value_entity_id, support, against, reason)
		VALUES (?, ?, 1, ?, 1, 0, 'kept')`
	for _, row := range [][2][]byte{{assocA, eventA}, {assocB, eventB}} {
		if _, err := db.Exec(ins, row[0], propID("event"), row[1]); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(ins, row[0], propID("person"), person); err != nil {
			t.Fatal(err)
		}
	}
	g := c.Graph()
	var n int
	g.SetQueryCounterForTest(func() { n++ })
	if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1}, Effects: effects.Set{Handles: [][]byte{assocA, assocB}}}); err != nil {
		t.Fatal(err)
	}
	if !g.LoadedForTest(person) || !g.LoadedForTest(eventA) || !g.LoadedForTest(eventB) {
		t.Fatal("shared endpoint was not loaded")
	}
	// One entity-row read, two endpoint lookups, one chunk (entity, values,
	// members, links, disambiguation). A reload per association is well above this.
	// Eight queries reload the endpoints once. The rest is one dependent walk
	// over the committed handles. A reload per association adds another chunk.
	if n > 16 {
		t.Fatalf("queries %d, want one reload of the shared endpoint", n)
	}
}

func TestLoadLinksUsesTheEdgeIndex(t *testing.T) {
	c := openCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	q, args := graphcache.LinkLoadSQL([][]byte{make([]byte, 16)})
	plans := dbtest.QueryPlan(t, db, q, args...)
	for _, detail := range plans {
		if strings.HasPrefix(detail, "SCAN ") && !strings.Contains(detail, "CONSTANT") {
			t.Fatalf("full scan: %s", detail)
		}
	}
	joined := strings.Join(plans, "\n")
	if !strings.Contains(joined, "auto_reconciler_values_edge_idx") {
		t.Fatalf("edge index unused:\n%s", joined)
	}
}

func TestDropLeavesAliasIndex(t *testing.T) {
	g := graphcache.New(nil)
	g.SetAliasIndex([]string{"jas"})
	g.Drop()
	raw, ok := g.AliasIndex()
	if !ok {
		t.Fatal("drop cleared alias index")
	}
	tables, ok := raw.([]string)
	if !ok || len(tables) != 1 || tables[0] != "jas" {
		t.Fatalf("%#v", raw)
	}
}

func TestDropLeavesNameFormat(t *testing.T) {
	g := graphcache.New(nil)
	if v, ok := (*graphcache.Graph)(nil).NameFormat(); ok || v != nil {
		t.Errorf("a nil graph holds name format %v, %v", v, ok)
	}
	if _, ok := g.NameFormat(); ok {
		t.Fatal("a new graph holds a name format")
	}
	g.SetNameFormat("western")
	g.Drop()
	if v, ok := g.NameFormat(); !ok || v != "western" {
		t.Fatalf("after Drop NameFormat = %v, %v, want western, true", v, ok)
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

// seedActivity writes one audit transaction scoped to ids of one scope type.
func seedActivity(t *testing.T, db *sql.DB, revision int64, at, scopeType string, ids ...[]byte) {
	t.Helper()
	tx := make([]byte, 16)
	tx[0], tx[15] = byte(revision), 0xA7
	if _, err := db.Exec(`INSERT INTO audit_transactions (id, revision, created_at) VALUES (?, ?, ?)`, tx, revision, at); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if _, err := db.Exec(`INSERT INTO audit_transaction_scopes (audit_transaction_id, scope_type, scope_id) VALUES (?, ?, ?)`, tx, scopeType, id); err != nil {
			t.Fatal(err)
		}
	}
}

func TestActivity(t *testing.T) {
	const (
		t3 = "2026-10-15T13:00:00Z"
		t5 = "2026-10-15T15:00:00Z"
		t7 = "2026-10-15T17:00:00Z"
	)
	h := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	other := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	src := []byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}
	tests := []struct {
		name string
		run  func(t *testing.T, g *graphcache.Graph, db *sql.DB, queries *int)
	}{
		{
			name: "first read loads once and takes the newest revision with its time",
			run: func(t *testing.T, g *graphcache.Graph, db *sql.DB, queries *int) {
				seedActivity(t, db, 3, t3, effects.ScopeCanonicalEntity, h)
				seedActivity(t, db, 5, t5, effects.ScopeCanonicalEntity, h, other)
				seedActivity(t, db, 6, t5, effects.ScopeSource, src)
				got, err := g.Activity(effects.ScopeCanonicalEntity, h)
				if err != nil || got != (effects.Activity{Revision: 5, At: t5}) {
					t.Fatalf("h %+v %v", got, err)
				}
				if *queries != 1 {
					t.Fatalf("queries %d, want 1", *queries)
				}
				if got, _ := g.Activity(effects.ScopeCanonicalEntity, other); got.Revision != 5 {
					t.Fatalf("other %+v", got)
				}
				if got, _ := g.Activity(effects.ScopeCanonicalEntity, src); got != (effects.Activity{}) {
					t.Fatalf("a Source id under canonical_entity: %+v", got)
				}
				if *queries != 1 {
					t.Fatalf("queries %d after memory reads, want 1", *queries)
				}
			},
		},
		{
			name: "an id with no scope row, a bad type, and a nil graph",
			run: func(t *testing.T, g *graphcache.Graph, db *sql.DB, queries *int) {
				if got, err := g.Activity(effects.ScopeSource, src); err != nil || got != (effects.Activity{}) {
					t.Fatalf("no rows: %+v %v", got, err)
				}
				for _, bad := range []string{"", "mystery"} {
					if _, err := g.Activity(bad, h); err == nil {
						t.Fatalf("type %q accepted", bad)
					}
				}
				var none *graphcache.Graph
				if got, err := none.Activity(effects.ScopeSource, src); err != nil || got != (effects.Activity{}) {
					t.Fatalf("nil graph: %+v %v", got, err)
				}
			},
		},
		{
			name: "a commit moves a loaded map forward without a query; an older one does not",
			run: func(t *testing.T, g *graphcache.Graph, db *sql.DB, queries *int) {
				seedActivity(t, db, 5, t5, effects.ScopeCanonicalEntity, h)
				if _, err := g.Activity(effects.ScopeCanonicalEntity, h); err != nil {
					t.Fatal(err)
				}
				seedActivity(t, db, 7, t7, effects.ScopeCanonicalEntity, h)
				before := *queries
				commit := effects.Commit{Activity: effects.Activity{Revision: 7, At: t7}, Effects: effects.Set{HandleScopes: [][]byte{h}, SourceScopes: [][]byte{src}}}
				if err := g.OnCommit(commit); err != nil {
					t.Fatal(err)
				}
				got, _ := g.Activity(effects.ScopeCanonicalEntity, h)
				if got != (effects.Activity{Revision: 7, At: t7}) {
					t.Fatalf("after commit %+v", got)
				}
				if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 6, At: t5}, Effects: effects.Set{HandleScopes: [][]byte{h}}}); err != nil {
					t.Fatal(err)
				}
				if got, _ := g.Activity(effects.ScopeCanonicalEntity, h); got.Revision != 7 {
					t.Fatalf("an older commit moved it: %+v", got)
				}
				// The Source map was never loaded, so its first read still loads.
				if *queries != before {
					t.Fatalf("queries %d, want %d: a commit or a memory read queried", *queries, before)
				}
			},
		},
		{
			name: "Drop clears the maps and the next read loads with one query",
			run: func(t *testing.T, g *graphcache.Graph, db *sql.DB, queries *int) {
				seedActivity(t, db, 5, t5, effects.ScopeCanonicalEntity, h)
				if _, err := g.Activity(effects.ScopeCanonicalEntity, h); err != nil {
					t.Fatal(err)
				}
				seedActivity(t, db, 8, t7, effects.ScopeCanonicalEntity, h)
				g.Drop()
				before := *queries
				got, err := g.Activity(effects.ScopeCanonicalEntity, h)
				if err != nil || got.Revision != 8 || *queries != before+1 {
					t.Fatalf("after Drop %+v %v, queries %d, want %d", got, err, *queries, before+1)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := openCatalog(t)
			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			g := c.Graph()
			var queries int
			g.SetQueryCounterForTest(func() { queries++ })
			tt.run(t, g, db, &queries)
		})
	}
}

func TestTermsByID(t *testing.T) {
	v := graphcache.Vocabulary{Terms: []graphcache.VocabTerm{
		{ID: []byte("id-a"), PropertyKey: "sex_at_birth", Key: "male", Label: "Male"},
		{ID: []byte("id-b"), PropertyKey: "sex_at_birth", Key: "unknown", Label: "Unknown", Neutral: true},
	}}
	got := v.TermsByID()
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if a := got["id-a"]; a.Key != "male" || a.Label != "Male" || a.Neutral {
		t.Errorf("id-a = %+v", a)
	}
	if b := got["id-b"]; b.Key != "unknown" || b.Label != "Unknown" || !b.Neutral {
		t.Errorf("id-b = %+v", b)
	}
}

func TestMatchIndexDrops(t *testing.T) {
	c := openCatalog(t)
	g := c.Graph()
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	stored := func() bool {
		_, ok := g.MatchIndex()
		return ok
	}
	g.SetMatchIndex("index")
	if v, ok := g.MatchIndex(); !ok || v != "index" {
		t.Fatalf("MatchIndex = %v, %v after SetMatchIndex", v, ok)
	}
	g.SetAliasIndex(nil)
	if stored() {
		t.Error("SetAliasIndex left the match index")
	}
	g.SetMatchIndex("index")
	g.SetNameFormat("western")
	if stored() {
		t.Error("SetNameFormat left the match index")
	}
	g.SetMatchIndex("index")
	if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1}, Effects: effects.Set{Handles: [][]byte{id}}}); err != nil {
		t.Fatal(err)
	}
	if stored() {
		t.Error("a commit that names a handle left the match index")
	}
	g.SetMatchIndex("index")
	g.Drop()
	if stored() {
		t.Error("Drop left the match index")
	}
	t.Run("vocabulary commit", func(t *testing.T) {
		g.SetMatchIndex("index")
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 2}, Effects: effects.Set{Lists: []effects.List{effects.ListVocabulary}}}); err != nil {
			t.Fatal(err)
		}
		if stored() {
			t.Error("a vocabulary-only commit left the match index")
		}
	})
	g.ResetMatchIndexForTest()
	if stored() {
		t.Error("ResetMatchIndexForTest left the match index")
	}
}

// S11-39 F-8: LoadKinds reads the links of every handle of the kinds through
// one catalog-wide association scan, and each node it installs equals the one
// a per-id load installs.
func TestLoadKindsMatchesNodeLoads(t *testing.T) {
	c := openCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	one := func(q string, args ...any) []byte {
		t.Helper()
		var id []byte
		if err := db.QueryRow(q, args...).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	typeID := func(key string) []byte {
		return one(`SELECT id FROM subject_types WHERE key = ? AND origin = 'provenencia'`, key)
	}
	propID := func(key string) []byte {
		return one(`SELECT id FROM properties WHERE key = ? AND origin = 'provenencia'`, key)
	}
	termID := func(prop, term string) []byte {
		return one(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
			WHERE p.key = ? AND p.origin = 'provenencia' AND t.key = ?`, prop, term)
	}
	var n byte
	entity := func(typ, ref string) []byte {
		n++
		id := []byte{4, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, n}
		if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, ?)`, id, typeID(typ), ref); err != nil {
			t.Fatal(err)
		}
		return id
	}
	value := func(assoc, prop, to []byte, rank int, reason string) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO auto_reconciler_values
			(entity_id, property_id, rank, value_entity_id, support, against, reason) VALUES (?, ?, ?, ?, 1, 0, ?)`,
			assoc, prop, rank, to, reason); err != nil {
			t.Fatal(err)
		}
	}
	term := func(assoc, prop, term []byte) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO auto_reconciler_values
			(entity_id, property_id, rank, value_term_id, support, against, reason) VALUES (?, ?, 1, ?, 1, 0, 'kept')`,
			assoc, prop, term); err != nil {
			t.Fatal(err)
		}
	}
	firstTerm := func(prop string) []byte {
		return one(`SELECT t.id FROM property_terms t JOIN properties p ON p.id = t.property_id
			WHERE p.key = ? AND p.origin = 'provenencia' ORDER BY t.key LIMIT 1`, prop)
	}

	p1, p2, p3, p4 := entity("person", "PER-1"), entity("person", "PER-2"), entity("person", "PER-3"), entity("person", "PER-4")
	merged := entity("person", "PER-M")
	e1, e2 := entity("event", "EVT-1"), entity("event", "EVT-2")
	l1, l2 := entity("place", "PLC-1"), entity("place", "PLC-2")
	if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, p1, merged); err != nil {
		t.Fatal(err)
	}
	person, event, place := propID("person"), propID("event"), propID("place")
	related, from, to := propID("related_to"), propID("from"), propID("to")

	// A participation with a role term.
	a := entity("participation", "BRG-1")
	value(a, person, p1, 1, "kept")
	value(a, event, e1, 1, "kept")
	term(a, propID("role"), termID("role", "subject"))
	// A relationship with a parent term.
	a = entity("relationship", "BRG-2")
	value(a, person, p1, 1, "kept")
	value(a, related, p2, 1, "kept")
	term(a, propID("relationship_type"), termID("relationship_type", "parent"))
	relAssoc := a
	// A location with no disambiguation.
	a = entity("location", "BRG-3")
	value(a, event, e1, 1, "kept")
	value(a, place, l1, 1, "kept")
	// A place relationship with a term.
	a = entity("place_relationship", "BRG-4")
	value(a, from, l1, 1, "kept")
	value(a, to, l2, 1, "kept")
	term(a, propID("place_relationship_type"), firstTerm("place_relationship_type"))
	// A merged association.
	a = entity("relationship", "BRG-5")
	value(a, person, p1, 1, "kept")
	value(a, related, p3, 1, "kept")
	if _, err := db.Exec(`UPDATE canonical_entities SET merged_into_id = ? WHERE id = ?`, relAssoc, a); err != nil {
		t.Fatal(err)
	}
	// An association whose endpoint value is rank 2, and one not kept.
	a = entity("relationship", "BRG-6")
	value(a, person, p2, 1, "kept")
	value(a, related, p3, 2, "kept")
	a = entity("participation", "BRG-7")
	value(a, person, p3, 1, "kept")
	value(a, event, e2, 1, "weak")
	// An association to a merged neighbor.
	a = entity("relationship", "BRG-8")
	value(a, person, p2, 1, "kept")
	value(a, related, merged, 1, "kept")
	term(a, propID("relationship_type"), termID("relationship_type", "parent"))
	// A self-loop relationship.
	a = entity("relationship", "BRG-9")
	value(a, person, p4, 1, "kept")
	value(a, related, p4, 1, "kept")
	term(a, propID("relationship_type"), termID("relationship_type", "parent"))
	// A person in two associations with the same event.
	a = entity("participation", "BRG-10")
	value(a, person, p2, 1, "kept")
	value(a, event, e1, 1, "kept")
	term(a, propID("role"), termID("role", "witness"))

	kinds := [][]byte{typeID("person"), typeID("event"), typeID("place")}
	scanned := graphcache.New(db)
	if err := scanned.LoadKinds(kinds); err != nil {
		t.Fatal(err)
	}
	perID := graphcache.New(db)
	total := 0
	for _, k := range kinds {
		want, err := scanned.Kind(k)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range want {
			got, err := perID.Node(w.ID)
			if err != nil {
				t.Fatal(err)
			}
			total += len(w.Links)
			if !reflect.DeepEqual(w.Links, got.Links) {
				t.Errorf("%s links differ:\nLoadKinds %+v\nNode      %+v", w.Ref, w.Links, got.Links)
			}
			if w.Ref != got.Ref || !reflect.DeepEqual(w.Values, got.Values) || !reflect.DeepEqual(w.Members, got.Members) {
				t.Errorf("%s differs beyond its links", w.Ref)
			}
		}
		again, err := perID.Kind(k)
		if err != nil {
			t.Fatal(err)
		}
		if len(again) != len(want) {
			t.Errorf("kind lists differ: %d vs %d", len(want), len(again))
		}
		for i := range again {
			if i < len(want) && !bytes.Equal(again[i].ID, want[i].ID) {
				t.Errorf("kind order differs at %d", i)
			}
		}
	}
	if total == 0 {
		t.Fatal("the fixture produced no links")
	}
}

// S11-39 F-8: after LoadKinds, reading the kinds issues no query.
func TestLoadKindsThenKindIssuesNoQuery(t *testing.T) {
	c := openCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	var typeID []byte
	if err := db.QueryRow(`SELECT id FROM subject_types WHERE key = 'person' AND origin = 'provenencia'`).Scan(&typeID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, 'PER-1')`,
		[]byte{5, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, typeID); err != nil {
		t.Fatal(err)
	}
	g := graphcache.New(db)
	if err := g.LoadKinds([][]byte{typeID}); err != nil {
		t.Fatal(err)
	}
	queries := 0
	g.SetQueryCounterForTest(func() { queries++ })
	nodes, err := g.Kind(typeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || queries != 0 {
		t.Fatalf("Kind after LoadKinds returned %d nodes with %d queries", len(nodes), queries)
	}
}

// S11-39 F-8: the association scan never reads a table end to end: the
// associations come from the type index, the endpoints from the edge index.
func TestLinkScanUsesIndexes(t *testing.T) {
	c := openCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	assocQ, endsQ, termsQ, assocArgs, endsArgs, termsArgs := graphcache.LinkScanSQL()
	for name, q := range map[string]struct {
		sql  string
		args []any
	}{
		"associations": {assocQ, assocArgs},
		"endpoints":    {endsQ, endsArgs},
		"terms":        {termsQ, termsArgs},
	} {
		plans := dbtest.QueryPlan(t, db, q.sql, q.args...)
		for _, detail := range plans {
			if strings.HasPrefix(detail, "SCAN ") && (strings.Contains(detail, "auto_reconciler_values") || strings.Contains(detail, "canonical_entities")) {
				t.Errorf("%s: full scan: %s", name, detail)
			}
		}
		if name == "endpoints" && !strings.Contains(strings.Join(plans, "\n"), "auto_reconciler_values_edge_idx") {
			t.Errorf("endpoints: edge index unused:\n%s", strings.Join(plans, "\n"))
		}
	}
}
