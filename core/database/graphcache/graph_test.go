package graphcache_test

import (
	"bytes"
	"database/sql"
	"strings"
	"testing"

	"github.com/mendahu/provenencia/core/database"
	"github.com/mendahu/provenencia/core/database/dbtest"
	"github.com/mendahu/provenencia/core/database/effects"
	"github.com/mendahu/provenencia/core/database/graphcache"
	"github.com/mendahu/provenencia/core/database/graphcache/graphcachetest"
	"github.com/mendahu/provenencia/core/database/observations"
	"github.com/mendahu/provenencia/core/database/rowchange"
	"github.com/mendahu/provenencia/core/writes"
)

func TestNodeAndKindAndVerify(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	typeID := graphcachetest.SubjectTypeID(t, c, "person")
	id := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	if _, err := db.Exec(`INSERT INTO canonical_entities (id, subject_type_id, ref) VALUES (?, ?, 'PER-1')`, id, typeID); err != nil {
		t.Fatal(err)
	}
	g := c.Graph()
	if loaded(g, id) {
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
	if loaded(g, id) {
		t.Fatal("Drop left a node")
	}
	if _, err := g.Node(id); err != nil {
		t.Fatal(err)
	}
	if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 1}, Effects: effects.Set{Structure: true}}); err != nil {
		t.Fatal(err)
	}
	if loaded(g, id) {
		t.Fatal("a structure commit left a node")
	}
}

func TestCommitInsertsANewHandleIntoALoadedKind(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	typeID := graphcachetest.SubjectTypeID(t, c, "person")
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
	if !loaded(g, id) {
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
	c := graphcachetest.OpenCatalog(t)
	canon := graphcachetest.NewCanon(t, c)
	person, event := canon.Entity("person"), canon.Entity("event")
	assoc := canon.Link("participation", map[string][]byte{"person": person, "event": event}, "role", "subject")
	g := c.Graph()
	canon.Commit(assoc)
	if !loaded(g, person) || !loaded(g, event) {
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
	c := graphcachetest.OpenCatalog(t)
	canon := graphcachetest.NewCanon(t, c)
	person, eventA, eventB := canon.Entity("person"), canon.Entity("event"), canon.Entity("event")
	assocA := canon.Link("participation", map[string][]byte{"person": person, "event": eventA}, "role", "subject")
	assocB := canon.Link("participation", map[string][]byte{"person": person, "event": eventB}, "role", "subject")
	g := c.Graph()
	var n int
	g.SetQueryCounterForTest(func() { n++ })
	canon.Commit(assocA, assocB)
	if !loaded(g, person) || !loaded(g, eventA) || !loaded(g, eventB) {
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

func TestCommitOfAMergedHandle(t *testing.T) {
	tests := []struct {
		name                      string
		mergedLoaded, eventLoaded bool
	}{
		{"both loaded", true, true},
		{"only the event loaded", false, true},
		{"neither loaded", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := graphcachetest.OpenCatalog(t)
			canon := graphcachetest.NewCanon(t, c)
			p, q := canon.Person(), canon.Person()
			e, joins := canon.Birth(p, nil, nil)
			canon.Commit(joins...)
			g := c.Graph()
			g.Drop()
			if tt.eventLoaded {
				if _, err := g.Node(e); err != nil {
					t.Fatal(err)
				}
			}
			if tt.mergedLoaded {
				if _, err := g.Node(p); err != nil {
					t.Fatal(err)
				}
			}
			canon.Merge(p, q)
			canon.Commit(p, q)
			if loaded(g, e) != tt.eventLoaded || loaded(g, p) != tt.mergedLoaded {
				t.Fatalf("loaded: event %v, merged %v", loaded(g, e), loaded(g, p))
			}
			if tt.eventLoaded {
				n, ok := g.Cached(e)
				if !ok {
					t.Fatal("the event is not cached")
				}
				for _, l := range n.Links {
					if bytes.Equal(l.Neighbor, p) {
						t.Fatal("the event still links to the merged Person")
					}
				}
			}
			if tt.mergedLoaded {
				m, ok := g.Cached(p)
				if !ok {
					t.Fatal("the merged Person is not cached")
				}
				if !m.Merged {
					t.Fatal("the merged Person is not marked merged")
				}
				if len(m.Links) != 1 || !bytes.Equal(m.Links[0].Neighbor, e) {
					t.Fatalf("the merged Person's links %+v, want its one link to the event", m.Links)
				}
			}
			if err := g.Verify(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCommitOfAMergedHandleNeitherLoadedQueries(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	canon := graphcachetest.NewCanon(t, c)
	p, q := canon.Person(), canon.Person()
	_, joins := canon.Birth(p, nil, nil)
	canon.Commit(joins...)
	g := c.Graph()
	g.Drop()
	canon.Merge(p, q)
	var n int
	g.SetQueryCounterForTest(func() { n++ })
	canon.Commit(p, q)
	// Two queries: the entity-row read of the committed handles and one link
	// query for the merged handle's neighbors. Nothing is loaded, so nothing
	// reloads, and the scan reads no disambiguation.
	if n != 2 {
		t.Fatalf("queries %d, want 2", n)
	}
}

func loaded(g *graphcache.Graph, id []byte) bool {
	_, ok := g.Cached(id)
	return ok
}

func TestLoadLinksUsesTheEdgeIndex(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
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
			c := graphcachetest.OpenCatalog(t)
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

func TestAcceptedMembers(t *testing.T) {
	var nilNode *graphcache.Node
	cases := []struct {
		name    string
		node    *graphcache.Node
		want    []byte
		wantLen int
	}{
		{"nil node", nilNode, nil, 0},
		{"no members", &graphcache.Node{}, nil, 0},
		{"only provisional", &graphcache.Node{Members: []graphcache.Member{{SubjectID: []byte{1}}}}, nil, 0},
		{"mixed keeps order", &graphcache.Node{Members: []graphcache.Member{
			{SubjectID: []byte{1}, Accepted: true},
			{SubjectID: []byte{2}},
			{SubjectID: []byte{3}, Accepted: true},
		}}, []byte{1, 3}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.node.AcceptedMembers()
			if len(got) != tc.wantLen {
				t.Fatalf("got %d accepted members, want %d", len(got), tc.wantLen)
			}
			for i, m := range got {
				if m.SubjectID[0] != tc.want[i] || !m.Accepted {
					t.Fatalf("member %d is %+v", i, m)
				}
			}
		})
	}
}

func TestNodeIs(t *testing.T) {
	person, place := []byte{1}, []byte{2}
	for _, tc := range []struct {
		name string
		n    *graphcache.Node
		want bool
	}{
		{"a nil node", nil, false},
		{"the type", &graphcache.Node{SubjectTypeID: person}, true},
		{"another type", &graphcache.Node{SubjectTypeID: place}, false},
		{"merged", &graphcache.Node{SubjectTypeID: person, Merged: true}, false},
	} {
		if got := tc.n.Is(person); got != tc.want {
			t.Errorf("%s: Is = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// outcomeRows reads the stored outcome rows of one Property of an entity.
func outcomeRows(t *testing.T, db *sql.DB, entity []byte, propKey string) []graphcache.Outcome {
	t.Helper()
	rows, err := db.Query(`SELECT ao.observation_id, o.subject_id, ao.reason, COALESCE(ao.value_rank, 0), ao.denied_by,
			COALESCE(ao.vote_support, 0), COALESCE(ao.vote_total, 0)
		FROM auto_reconciler_outcomes ao
		JOIN observations o ON o.id = ao.observation_id
		JOIN properties p ON p.id = ao.property_id
		WHERE ao.entity_id = ? AND p.key = ? AND p.origin = 'provenencia'
		ORDER BY ao.observation_id`, entity, propKey)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []graphcache.Outcome
	for rows.Next() {
		var o graphcache.Outcome
		if err := rows.Scan(&o.ObservationID, &o.SubjectID, &o.Reason, &o.ValueRank, &o.DeniedBy, &o.VoteSupport, &o.VoteOf); err != nil {
			t.Fatal(err)
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func sameOutcomeRow(a, b graphcache.Outcome) bool {
	return bytes.Equal(a.ObservationID, b.ObservationID) && bytes.Equal(a.SubjectID, b.SubjectID) &&
		a.Reason == b.Reason && a.ValueRank == b.ValueRank && bytes.Equal(a.DeniedBy, b.DeniedBy) &&
		a.VoteSupport == b.VoteSupport && a.VoteOf == b.VoteOf
}

func propertyID(t *testing.T, db *sql.DB, key string) string {
	t.Helper()
	var id []byte
	if err := db.QueryRow(`SELECT id FROM properties WHERE key = ? AND origin = 'provenencia'`, key).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return string(id)
}

func TestNodeOutcomes(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	k := graphcachetest.NewCanon(t, c)
	p := k.Entity("person")
	k.Term(p, "sex_at_birth", "male")
	k.Commit(p)
	g := c.Graph()
	prop := propertyID(t, db, "sex_at_birth")

	n, err := g.Node(p)
	if err != nil {
		t.Fatal(err)
	}
	rows := n.Outcomes[prop]
	if len(rows) != 1 {
		t.Fatalf("outcome rows %+v", n.Outcomes)
	}
	o := rows[0]
	if !bytes.Equal(o.SubjectID, p) || o.Reason != "kept" || o.ValueRank != 1 || o.DeniedBy != nil || o.VoteSupport != 0 || o.VoteOf != 0 {
		t.Fatalf("outcome %+v", o)
	}
	if want := outcomeRows(t, db, p, "sex_at_birth"); len(want) != 1 || !sameOutcomeRow(o, want[0]) {
		t.Fatalf("node %+v, stored %+v", rows, want)
	}

	k.Term(p, "sex_at_birth", "female")
	k.Commit(p)
	n, err = g.Node(p)
	if err != nil {
		t.Fatal(err)
	}
	rows = n.Outcomes[prop]
	want := outcomeRows(t, db, p, "sex_at_birth")
	if len(rows) != 2 || len(want) != 2 {
		t.Fatalf("node %d rows, stored %d", len(rows), len(want))
	}
	for i := range rows {
		if !sameOutcomeRow(rows[i], want[i]) {
			t.Errorf("row %d: node %+v, stored %+v", i, rows[i], want[i])
		}
	}
	if bytes.Compare(rows[0].ObservationID, rows[1].ObservationID) >= 0 {
		t.Error("rows are not in ascending Observation id")
	}
	if err := g.Verify(); err != nil {
		t.Fatal(err)
	}
}

// likenessObservation returns an Observation id of the entity's Subject.
func likenessObservation(t *testing.T, db *sql.DB, entity []byte) []byte {
	t.Helper()
	var x []byte
	if err := db.QueryRow(`SELECT id FROM observations WHERE subject_id = ? LIMIT 1`, entity).Scan(&x); err != nil {
		t.Fatal(err)
	}
	return x
}

func TestNodePreferredLikeness(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	k := graphcachetest.NewCanon(t, c)
	p := k.Entity("person")
	k.Term(p, "sex_at_birth", "male")
	k.Commit(p)
	g := c.Graph()
	if n, err := g.Node(p); err != nil || n.PreferredLikeness != nil {
		t.Fatalf("node %+v, %v", n, err)
	}
	x := likenessObservation(t, db, p)
	if _, err := db.Exec(`UPDATE canonical_entities SET preferred_likeness_observation_id = ? WHERE id = ?`, x, p); err != nil {
		t.Fatal(err)
	}
	commit := func() {
		t.Helper()
		if err := g.OnCommit(effects.Commit{Activity: effects.Activity{Revision: 100}, Effects: effects.Set{Handles: [][]byte{p}}}); err != nil {
			t.Fatal(err)
		}
	}
	commit()
	n, err := g.Node(p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(n.PreferredLikeness, x) {
		t.Fatalf("PreferredLikeness %x, want %x", n.PreferredLikeness, x)
	}
	if _, err := db.Exec(`UPDATE canonical_entities SET preferred_likeness_observation_id = NULL WHERE id = ?`, p); err != nil {
		t.Fatal(err)
	}
	commit()
	if n, err = g.Node(p); err != nil || n.PreferredLikeness != nil {
		t.Fatalf("PreferredLikeness after clear %x, %v", n.PreferredLikeness, err)
	}
}

func TestVerifyChecksOutcomesAndLikeness(t *testing.T) {
	for _, tc := range []struct {
		name, sql, want string
		useObservation  bool
		moveSubject     bool
	}{
		{"an outcome reason", `UPDATE auto_reconciler_outcomes SET reason = 'outvoted'`, "outcomes", false, false},
		{"the chosen likeness", `UPDATE canonical_entities SET preferred_likeness_observation_id = ?`, "", true, false},
		{"an outcome's Subject", `UPDATE observations SET subject_id = ?`, "outcomes", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := graphcachetest.OpenCatalog(t)
			db, err := c.DB()
			if err != nil {
				t.Fatal(err)
			}
			k := graphcachetest.NewCanon(t, c)
			p := k.Entity("person")
			k.Term(p, "sex_at_birth", "male")
			// A second member of the handle: the Observation moves to it and the
			// handle's values do not change.
			other := []byte{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
			var typeID, source, graph []byte
			if err := db.QueryRow(`SELECT subject_type_id, source_id, graph_id FROM subjects WHERE id = ?`, p).Scan(&typeID, &source, &graph); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO subjects (id, ref, source_id, graph_id, subject_type_id) VALUES (?, 'SUB-OTHER', ?, ?, ?)`, other, source, graph, typeID); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`INSERT INTO identity_claims (id, subject_id, entity_id, subject_type_id, status) VALUES (?, ?, ?, ?, 'accepted')`,
				[]byte{8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8}, other, p, typeID); err != nil {
				t.Fatal(err)
			}
			k.Commit(p)
			g := c.Graph()
			if _, err := g.Node(p); err != nil {
				t.Fatal(err)
			}
			if err := g.Verify(); err != nil {
				t.Fatal(err)
			}
			var args []any
			if tc.useObservation {
				args = append(args, likenessObservation(t, db, p))
			}
			if tc.moveSubject {
				args = append(args, other)
			}
			if _, err := db.Exec(tc.sql, args...); err != nil {
				t.Fatal(err)
			}
			err = g.Verify()
			if err == nil {
				t.Fatal("Verify accepted a node that differs from the catalog")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not name %q", err, tc.want)
			}
		})
	}
}

// A committed write that moves an Observation to another Subject reloads the
// handle's node, so its outcome rows name the new Subject and Verify passes.
func TestOutcomeSubjectFollowsAWrite(t *testing.T) {
	c := graphcachetest.OpenCatalog(t)
	db, err := c.DB()
	if err != nil {
		t.Fatal(err)
	}
	k := graphcachetest.NewCanon(t, c)
	p, q := k.Entity("person"), k.Entity("person")
	k.Term(p, "sex_at_birth", "male")
	k.Commit(p, q)
	g := c.Graph()
	if _, err := g.Node(p); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Node(q); err != nil {
		t.Fatal(err)
	}
	userID := []byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}
	var id, prop, term []byte
	if err := db.QueryRow(`SELECT id, property_id, value_term_id FROM observations WHERE subject_id = ?`, p).Scan(&id, &prop, &term); err != nil {
		t.Fatal(err)
	}
	if _, err := writes.Call(c, writes.Op{Action: "update_observation", UserID: userID}, func(tx *database.Tx) (observations.Listed, []rowchange.Change, error) {
		return observations.Update(tx, userID, observations.Input{ID: id, SubjectID: q, PropertyID: prop, ValueTermID: term})
	}); err != nil {
		t.Fatal(err)
	}
	n, err := g.Node(q)
	if err != nil {
		t.Fatal(err)
	}
	moved := false
	for _, rows := range n.Outcomes {
		for _, o := range rows {
			moved = moved || (bytes.Equal(o.ObservationID, id) && bytes.Equal(o.SubjectID, q))
		}
	}
	if !moved {
		t.Fatalf("outcomes of the new Subject's node: %+v", n.Outcomes)
	}
	if err := g.Verify(); err != nil {
		t.Fatal(err)
	}
}
