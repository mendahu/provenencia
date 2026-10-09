# A catalog graph in memory

**Status:** idea, 2026-10-09. Direction agreed in review of the Promote cost fix (mendahu/provenencia#327, to be redone on this). Not scheduled.

Related: [`promote-graph-alignment.md`](../promote-graph-alignment.md) (the walk), [`conclusion-layer-data-model.md`](../conclusion-layer-data-model.md) (handles and the auto-reconciler), [`archive/catalog-access-serialization.md`](archive/catalog-access-serialization.md) (the held catalog session), [`archive/page-navigation-performance.md`](archive/page-navigation-performance.md) (the Swift session cache, which this does not replace).

## The problem

The canonical graph already has a maintained, persistent form: `auto_reconciler_values` (ARV), which every write rewrites inside its own transaction (`autoreconciler.RecomputeTx`). Every reader then rebuilds pieces of it from SQL on every request and throws them away:

| Reader | Rebuilds | Lifetime |
| --- | --- | --- |
| Promote canon (`promotealign.loadCanon`) | handles, values and edges up to 5 hops out | one proposal |
| Promote stats (`promotealign/stats.go`) | value frequency, fan-out | process, keyed by file + revision |
| Promote candidates (`matching.CandidatesOfType`) | every handle of a kind | one proposal (cached per revision in #327) |
| Place chains (`conclusionheaders.placeGraph`) | place links, up to 8 rounds | one request (indexed per request in #328) |
| Conclusion headers (`ListPersons`/`ListEvents`/`ListPlaces`, `…ByIDs`) | handles, values, walks, chains | one request |

Promote is where it hurts first. It proposes again on every decision, and the eager 5-hop expansion pulls whole neighborhoods around hubs (a town with hundreds of events). #327 bounded that with caps (`maxEdgesPerStep`, `maxCanonHandles`), which silently drop real candidates past the cap. We don't want caps. We want reads that are cheap enough not to need them.

## The idea

Two tiers of the same data:

- **Persistent tier:** ARV, as today. Survives restarts; SQL and the search index read it.
- **Memory tier:** a graph of nodes held on the open catalog session, built from ARV (plus a few small lookups), shaped for walking: each node holds its values and its adjacency, and points at its neighbors.

`RecomputeTx` and `Rebuild` stay the only ways the canonical graph changes. They already rewrite ARV; they also tell the memory tier which entries are now stale. Nothing new is asked of write code.

Two stores live in the memory tier:

1. **Canonical store:** one node per handle (Person, Event, Place, and the association handles that join them).
2. **Source store:** per Source, its Subjects and their connections (the Evidence graph), the layer Promote aligns.

Memberships (`identity_claims`) are the pointers between them: a Source Subject points at the handle it belongs to.

Everything lives in Go. The data comes from SQLite and every consumer (Align, scoring, exhibits, headers, place chains, details) is Go. The FFI is request/response protobuf; Swift keeps caching view-shaped responses as it does today.

## What a node holds

Two kinds of data on one node.

**Structure** (the walk, matching, stats):

- Identity: id, ref, kind, merged.
- Values: every ARV row for the handle (all ranks, with reason), by property id. Matching reads kept rank-1; the detail page reads all of them.
- Adjacency: one link per association, both directions, with its edge signature (bridge type, role or term, neighbor kind, directed). Links are indexed by signature, so "neighbors of h through `participation/subject`" is a map lookup.
- Members: accepted and provisional member Subject ids (exhibits, memberships, the Source store's back-pointers).

**Display** (lists, headers, detail):

- A memo built from the node and its neighbors the first time a header or detail is asked for, and dropped when the node or a neighbor it reads changes (see Invalidation). It is derived from structure, never filled independently, so it cannot drift from it.
- It holds ids, not labels: term and property labels, Source titles, and "today" are resolved when read (see What isn't in ARV).

```go
// Package graphcache is the memory tier of the canonical graph: ARV shaped
// for walking, held on the open catalog session.
package graphcache

type Node struct {
	ID     []byte
	Ref    string
	Kind   string // person, event, place, or an association kind
	Merged bool

	Values  map[PropertyKey][]Value // every ARV row, rank order; Value carries rank and reason
	Links   map[SigKey][]Link       // adjacency by edge signature, both directions
	Members [][]byte                // accepted and provisional member Subjects

	display *Display // memo; nil until asked, cleared by a display mark
}

type Link struct {
	Neighbor    []byte // the handle at the other end
	Association []byte // the association handle (participation, location, …)
	Sig         graphalign.EdgeSignature
	FromEnd     bool // true when this node is the association's first endpoint
}
```

Nodes are immutable once built. A reload replaces the node in the map, so a reader holding the old pointer sees a consistent old value. `catalogsession.Do` already serializes every operation on a project, so the store needs no lock of its own today, and immutability keeps that true if reads ever run concurrently.

## Reading: lazy fill

Nothing is preloaded. A node is loaded the first time something asks for it, in batches:

```go
type Graph struct { /* nodes, kinds, stats memo, marks cursor */ }

// Nodes returns the nodes for ids, loading the missing or stale ones in one
// batched read (InBatch), and drains pending marks first.
func (g *Graph) Nodes(q Querier, ids [][]byte) ([]*Node, error)

// Neighbors is h's links through sig, loading the neighbors it names.
func (g *Graph) Neighbors(q Querier, h *Node, sig graphalign.EdgeSignature) ([]*Node, error)

// Kind is every unmerged handle of a kind, loaded once (lists, candidates);
// later marks refresh only the marked nodes.
func (g *Graph) Kind(q Querier, kind string) ([]*Node, error)

// Header is the node's display memo, built on first use.
func (g *Graph) Header(q Querier, id []byte) (*Display, error)

// Stats is value frequency and fan-out, kept as counts the marks adjust.
func (g *Graph) Stats(q Querier) (graphalign.Stats, error)
```

Every read starts by draining marks (below), so a read after a write always sees committed data.

## Invalidation

### One function, already called everywhere

Every write that changes the canonical graph ends in `autoreconciler.RecomputeTx(ids)`, through one of six entry points:

| Entry point | Called by | Resolves to |
| --- | --- | --- |
| `RecomputeSubjectsTx` | observation create, update and delete | handles the Subjects are members of |
| `RecomputeTouchingTx` | identity claim create, single promote | the handles plus handles whose members point at the Subject |
| `RecomputeCitationTx` | citation certainty | handles with Observations under the citation |
| `RecomputeSourceTx` | source credibility | handles with Observations in the Source |
| `RecomputePropertyTx` | property cardinality | handles with Observations of the property |
| `RecomputeTx` | all of the above; also directly by promote batch, subject delete, observation delete (released handles), identity claim create | the handles themselves |

`RecomputeTx` then computes `conclusionheaders.HeaderDependents(ids)`, the handles whose headers embed one of these (a Place reaches its child places, its events and their subject persons; an Event its subject persons; a Person the events of their subject-role participations), and reprojects search for both sets.

So the question "I changed observation X, what is stale?" is already answered on every write, inside the write's transaction, and tests already depend on it: a write path that skipped it would leave ARV and search stale.

### Two marks per write

- **Structure mark:** the handles `RecomputeTx` rewrote. Their nodes reload.
- **Display mark:** those handles plus `HeaderDependents`. Their display memos clear.

`Rebuild` (open-time, on a `CacheVersion` change) marks everything.

### Marks are rows in the write's transaction

`RecomputeTx` gets only a `Querier` (the write's `*sql.Tx`), not the session, so it can't reach the memory tier directly. It writes its marks into a small table in the same transaction:

```sql
CREATE TABLE graph_cache_marks (
	seq       INTEGER PRIMARY KEY AUTOINCREMENT,
	entity_id BLOB,          -- NULL: everything (Rebuild)
	scope     TEXT NOT NULL  -- 'structure' or 'display'
);
```

```go
// In RecomputeTx, after the ARV rows and before the search reprojection:
if err := graphcache.MarkTx(q, ids, deps); err != nil {
	return err
}

// On every read, before serving:
func (g *Graph) drain(q Querier) error {
	// SELECT seq, entity_id, scope FROM graph_cache_marks WHERE seq > g.cursor
	// structure → drop node; display → clear memo; NULL → drop all.
	// g.cursor = max seq seen.
}
```

This gets the transaction semantics for free:

- A rolled-back write leaves no marks, and a committed one always does. The memory tier never sees uncommitted data and never misses a committed change.
- Write code doesn't change. The only new call is inside `RecomputeTx` and `Rebuild`.
- It works for any writer that goes through `RecomputeTx`, including paths outside the FFI (onboarding, tests).

The table is pruned at open (the memory tier starts empty, so old marks mean nothing) and can be trimmed below the cursor at any time.

Alternatives considered:

- **Mark memory directly from `RecomputeTx`.** Needs the session reachable from a `*sql.Tx` (a wrapper `Querier` or a package-level "current graph"), and a rollback after the mark costs a reload but is otherwise harmless. More plumbing, less exact.
- **Drop everything on any revision change.** Simple and correct, and fine for Promote, which makes no writes between decisions. Not fine once lists read the cache: one edit in the Citation Composer would reload every Person.
- **A separate change feed.** Rejected: it would duplicate audit (what changed) and ARV (what the graph is now).

### What isn't in ARV

These change without `RecomputeTx`, so nodes don't copy them. They are resolved when read:

- **Term and property labels.** A term rename changes every header that shows it. Nodes hold term and property ids; a small vocabulary map (labels, value types, kinship and inverses) is dropped whole on any `properties` or `property_terms` write.
- **Source titles** in the detail page's "Why". Held as Source ids; looked up when read.
- **Today.** Place headers use `TodayDate()` for current parents. The memo keeps the dated periods and settles "today" when read.

### Safety net

The memory tier records the audit revision it last drained against. If a read finds a revision it didn't expect with no marks to explain it (a path that changes ARV without `RecomputeTx`), it drops everything and logs it. That log line is a bug report.

### Verifying it

A shadow check makes a missed invalidation fail a test instead of showing stale data:

```go
// Verify rebuilds every loaded node from SQL and compares. Tests run it after
// each write when PROVENENCIA_GRAPH_VERIFY=1; CI sets it for the core suite.
func (g *Graph) Verify(q Querier) error
```

The existing suites exercise every write path, so turning the check on in CI covers them all. A debug build can run it in the app while dogfooding.

## The Source store

Same pattern, different source of marks. Audit already resolves every transaction to the Sources it touched (`audit_transaction_scopes`, `scope_type = 'source'`), so the Source store doesn't need its own marks table: it drains "Sources scoped by a revision above my cursor" and reloads those.

```go
type SourceGraph struct {
	SourceID  []byte
	Subjects  map[string]*Subject // values, kind, ref
	Bridges   []Bridge            // the Evidence graph's connections, with signatures
	Members   map[string][]byte   // Subject → handle (identity claims); points into the canonical store
	Titles    map[string]string   // event titles (sourceeventtitles)
}

func (g *Graph) Source(q Querier, sourceID []byte) (*SourceGraph, error)
```

Subject positions (Evidence graph layout) don't bump the revision and aren't in the store; the Evidence graph keeps reading them directly.

## Where it's used

### Canonical store

| Consumer | Today | With the graph |
| --- | --- | --- |
| Promote walk (`graphalign.Align`) | eager 5-hop canon (`loadCanon`, `canonSteps`), with caps in #327 | lazy: Align asks for a handle and its neighbors by signature; no expansion, no caps |
| Promote candidates | kind scan per proposal (`matching.CandidatesOfType`) | `Kind(kind)` |
| Promote stats | recomputed per revision (`computeStats`) | counts on the graph, adjusted by marks |
| Promote exhibits | members and one-hop bridge neighbors per proposal | node `Members` and `Links` |
| Persons / Events / Places lists | `ListPersons` / `ListEvents` / `ListPlaces` | `Kind` + `Header` |
| Headers by id | `PersonsByIDs` / `EventsByIDs` / `PlacesByIDs` (detail header, Promote rows and alternatives, search hits, membership badges) | `Header` |
| Place chains and place detail | `placeGraph` per request; `ParentsAtDate`, `PartsAtDate`, `SuccessionNames`, `PlaceDetail` | following `part_of` and `succeeded_by` links; replaces #328's per-request index |
| Conclusion detail (`conclusiondetails.ForEntity`) | ARV rows, outcomes and Observations per request | `Values` (all ranks) from the node; per-Observation outcomes stay a SQL read, or a second memo with the same display mark |
| Future: pedigree, timelines, "what happened here" | — | walks over `Links` |

### Source store

| Consumer | Today | With the graph |
| --- | --- | --- |
| Promote layer (`promotealign.loadLayer`) | rebuilt per proposal | `Source(id)` |
| Evidence graph, Citation Composer | five FFI calls (`listSubjects`, `listObservationsBySource`, `listSubjectMemberships`, `listSourceEventTitles`, positions) | four from `Source(id)`; positions as today |

### Stays on SQL

- **Writes.** `RecomputeTx` and search reprojection read the write's own uncommitted rows; the memory tier only knows committed state.
- **Full-text search.** FTS stays the query; only hit headers come from `Header`.
- **Counts** (`workspaceNavCounts`), already a cheap `COUNT(*)`.

## Promote on the graph

Align reads the canon in three ways after setup: a handle by id (`st.handles`), one handle's links (`st.canonAdj`), and handles of a kind for the unreachable fallback (`candidatesOfKind`). The walk only visits handles reachable from its anchors, and only crosses a link that corresponds to a bridge in the Source (`corresponds()`). So the scope of the walk is already set by the Source being promoted: no `canonSteps`, no diameter, no per-kind hop rules. A hub's 400 events are read only when a layer event points at that hub, and then all 400 are candidates.

`Canon` becomes an interface over the graph:

```go
// CanonGraph is what Align reads of the canonical graph.
type CanonGraph interface {
	Handle(id []byte) (*Handle, bool)
	Links(id []byte) []CanonLink
	Seeds(kind string) []match.Candidate
}
```

Align is pure and has no error path; the adapter records the first load error, returns empty results after it, and `Propose` returns that error once Align finishes. (Or Align grows error returns; decide in the PR.)

Weighting, not hopping, is where per-signature judgment belongs: fan-out already makes a connection with many neighbors count for less (`EdgeSupportHigh`, `NeighborCreditHigh` in `graphalign/registry.go`).

## Memory

Unknown until measured. A node is an id, a ref, a few dozen values and a handful of links. The first PR adds a benchmark on a synthetic catalog (on the order of 20k persons, 30k events, a few hub places) that reports heap size and proposal time against today's eager canon.

If it's too much, `Values` can hold only kept rank-1 rows and the detail page reads the rest from SQL.

## Plan

Each step is its own PR, measured against the one before.

1. **The graph and Promote.** `graphcache` on the catalog session; nodes with structure; lazy fill; `graph_cache_marks` written by `RecomputeTx` and `Rebuild`; the revision safety net; `Verify` in CI. Promote reads through `CanonGraph`; stats and candidates move onto the graph; `canonSteps`, the 5-hop expansion and the #327 caps go. The #327 debounce and narrowed exhibits stay. Benchmark.
2. **Display.** Header memos; lists and `…ByIDs` from the graph; the vocabulary map; "today" at read time.
3. **Place chains** from links, replacing #328's per-request index.
4. **Conclusion detail** from node values, with the "Why" outcomes on the same display mark.
5. **The Source store**, drained from audit scopes; Promote's layer, the Evidence graph and the Composer on it.

Later, and separately: the same marks could ride back to Swift on each write's response ("these handles changed"), so the client's session cache invalidates by handle instead of the hand-kept `CatalogMutation` map. That reuses the marks; it is not another feed.

## Open questions

1. **Where the graph hangs.** A field on `database.Catalog` (low in the import graph, so an opaque slot), or a map in `graphcache` keyed by `*database.Catalog`, set up in `catalogsession.openResearcher` and dropped on close.
2. **Marks table vs. marking memory directly** (see Invalidation). The table is the recommendation; confirm it in PR 1.
3. **Stats by counts or recompute.** Adjusting value-frequency and fan-out counts from marks is exact but fiddly. Recomputing from the graph per revision is simpler and may be fast enough in memory.
4. **Detail "Why" outcomes.** Cache them per handle, or keep the SQL read; decide after measuring.
5. **Concurrency.** `Do` serializes everything today. If reads ever run alongside writes, immutable nodes plus a lock around the maps is enough; nothing here should assume serialization beyond that.
