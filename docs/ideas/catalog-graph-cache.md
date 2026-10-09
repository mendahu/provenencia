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

`RecomputeTx` and `Rebuild` stay the only ways the canonical graph changes. They already rewrite ARV; they also note which entries are now stale on the write's transaction, and the memory tier drops them when that transaction commits. Writes move onto one transaction wrapper so that commit is the moment the cache hears about them.

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

	display *Display // memo; nil until asked, cleared when a commit makes it stale
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
type Graph struct { /* nodes, kinds, stats memo, last revision heard */ }

// Nodes returns the nodes for ids, loading the missing or stale ones in one
// batched read (InBatch).
func (g *Graph) Nodes(q Querier, ids [][]byte) ([]*Node, error)

// Neighbors is h's links through sig, loading the neighbors it names.
func (g *Graph) Neighbors(q Querier, h *Node, sig graphalign.EdgeSignature) ([]*Node, error)

// Kind is every unmerged handle of a kind, loaded once (lists, candidates);
// later writes refresh only the nodes they touched.
func (g *Graph) Kind(q Querier, kind string) ([]*Node, error)

// Header is the node's display memo, built on first use.
func (g *Graph) Header(q Querier, id []byte) (*Display, error)

// Stats is value frequency and fan-out, kept as counts that commits adjust.
func (g *Graph) Stats(q Querier) (graphalign.Stats, error)
```

Writes update the graph as they commit (see Invalidation), so a read never has to check for staleness first.

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

### Two kinds of stale per write

- **Structure:** the handles `RecomputeTx` rewrote. Their nodes reload.
- **Display:** those handles plus `HeaderDependents`. Their display memos clear.

`Rebuild` (open-time, on a `CacheVersion` change) makes everything stale.

### Writes go through one wrapper that knows how they ended

Today about 24 packages call `db.Begin()` themselves (some 45 call sites) and commit or roll back on their own. That's why `RecomputeTx` can't safely tell the memory tier anything: it only holds the write's `*sql.Tx`, and it can't know whether that transaction will commit.

Instead, every write runs through one wrapper on the catalog. The wrapper owns the transaction, collects what the write changed, and tells the memory tier only after a successful commit:

```go
// package database

// Tx is a write transaction. It is a Querier, and it collects what the write
// changed so listeners hear about it after commit, never before.
type Tx struct {
	*sql.Tx
	changes Changes
}

// Changes is what one committed write touched. It lives for one write and is
// never stored.
type Changes struct {
	Handles    [][]byte // structure: ARV rows rewritten (RecomputeTx)
	Dependents [][]byte // display only: HeaderDependents
	Sources    [][]byte // source scopes (audit.Record)
	Vocabulary bool     // properties or property_terms written
	All        bool     // Rebuild
}

// Write runs fn in a transaction. On success it commits, then hands the
// collected Changes to every listener. On error or panic it rolls back and
// the Changes are dropped.
func (c *Catalog) Write(fn func(tx *Tx) error) error {
	sqlTx, err := c.db.Begin()
	if err != nil {
		return err
	}
	tx := &Tx{Tx: sqlTx}
	if err := fn(tx); err != nil {
		_ = sqlTx.Rollback()
		return err
	}
	if err := sqlTx.Commit(); err != nil {
		return err
	}
	for _, l := range c.listeners {
		l.Committed(tx.changes)
	}
	return nil
}

// OnCommit registers a listener. The graph cache registers itself when the
// catalog session opens.
func (c *Catalog) OnCommit(l CommitListener)
```

The functions that already know what changed note it on the `Tx`:

```go
// package autoreconciler: the signature takes *database.Tx, not a Querier,
// so a write can't call it outside the wrapper.
func RecomputeTx(tx *database.Tx, entityIDs [][]byte) error {
	// … rewrite ARV rows, compute deps, reproject search, as today …
	tx.Touched(ids, deps)
	return nil
}

// package audit: scopes are already resolved here.
func Record(tx *database.Tx, rev Revision) (int64, error) {
	// … as today …
	tx.TouchedSources(sourceIDs)
}

// package propertyterms / properties: any write
tx.TouchedVocabulary()
```

A write path reads the same as today, minus its own begin, commit and rollback:

```go
// before
tx, err := db.Begin()
if err != nil { return err }
defer tx.Rollback()
// … writes, audit.Record, autoreconciler.RecomputeSubjectsTx …
return tx.Commit()

// after
return c.Write(func(tx *database.Tx) error {
	// … writes, audit.Record, autoreconciler.RecomputeSubjectsTx …
	return nil
})
```

What this gets:

- **Exact:** the memory tier hears about a write only once it has committed, and always hears about one that has. A rollback drops the `Changes` with the transaction.
- **No new table, nothing stored.** `Changes` is the outcome of one transaction, held in memory until commit and then handed over. It isn't a feed or a log: audit stays the record of what changed, and ARV stays the record of what the graph is.
- **Enforced by the compiler.** `RecomputeTx` and `audit.Record` take `*database.Tx`, which only `Catalog.Write` makes. A write that bypasses the wrapper doesn't compile against them. A small test that fails on any `.Begin()` outside `core/database` keeps it that way.
- **One mechanism for both stores.** The Source store hears `Sources` from the same `Changes` instead of querying audit scopes.
- **Vocabulary included.** Term and property writes set `Vocabulary`, so the label map drops on exactly those writes.
- **Reads don't check anything** before serving. Everything was applied at commit.

`catalogsession.Do` already serializes every operation, so listeners run before the next read can start, with no lock needed.

The restructure is mechanical but wide: about 45 begin sites in 24 packages move to `Catalog.Write`, and functions that take a `*sql.Tx` today take a `*database.Tx` (it embeds `*sql.Tx`, so their bodies don't change). Paths that write outside a project catalog (onboarding before the session opens, `core/derivatives`) use the same wrapper; with no listeners registered they behave as today.

Alternatives considered:

- **A marks table** written by `RecomputeTx` in the same transaction, drained on read. Exact and needs no restructure, but adds a table and a query before every read.
- **Mark memory directly from `RecomputeTx`.** Can't know whether the transaction commits; a rollback leaves memory marked for data that never landed.
- **Drop everything on any revision change.** Simple and fine for Promote, which makes no writes between decisions. Not fine once lists read the cache: one edit in the Citation Composer would reload every Person.
- **A persisted change feed.** Rejected: it would duplicate audit (what changed) and ARV (what the graph is now).

### What isn't in ARV

These change without `RecomputeTx`, so nodes don't copy them. They are resolved when read:

- **Term and property labels.** A term rename changes every header that shows it. Nodes hold term and property ids; a small vocabulary map (labels, value types, kinship and inverses) is dropped whole on any `properties` or `property_terms` write.
- **Source titles** in the detail page's "Why". Held as Source ids; looked up when read.
- **Today.** Place headers use `TodayDate()` for current parents. The memo keeps the dated periods and settles "today" when read.

### Safety net

The memory tier records the audit revision of the last commit it heard about. If a read finds a newer revision it never heard about (a write that slipped past the wrapper), it drops everything and logs it. That log line is a bug report. With the compiler and the `.Begin()` test in place it should never fire.

### Verifying it

A shadow check makes a missed invalidation fail a test instead of showing stale data:

```go
// Verify rebuilds every loaded node from SQL and compares. Tests run it after
// each write when PROVENENCIA_GRAPH_VERIFY=1; CI sets it for the core suite.
func (g *Graph) Verify(q Querier) error
```

The existing suites exercise every write path, so turning the check on in CI covers them all. A debug build can run it in the app while dogfooding.

## The Source store

Same pattern. `audit.Record` already resolves every transaction to the Sources it touched (the scopes in `audit_transaction_scopes`); it notes them on the `Tx`, and the Source store drops those Sources when the write commits.

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
| Promote stats | recomputed per revision (`computeStats`) | counts on the graph, adjusted at each commit |
| Promote exhibits | members and one-hop bridge neighbors per proposal | node `Members` and `Links` |
| Persons / Events / Places lists | `ListPersons` / `ListEvents` / `ListPlaces` | `Kind` + `Header` |
| Headers by id | `PersonsByIDs` / `EventsByIDs` / `PlacesByIDs` (detail header, Promote rows and alternatives, search hits, membership badges) | `Header` |
| Place chains and place detail | `placeGraph` per request; `ParentsAtDate`, `PartsAtDate`, `SuccessionNames`, `PlaceDetail` | following `part_of` and `succeeded_by` links; replaces #328's per-request index |
| Conclusion detail (`conclusiondetails.ForEntity`) | ARV rows, outcomes and Observations per request | `Values` (all ranks) from the node; per-Observation outcomes stay a SQL read, or a second memo dropped with the display memo |
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

Unknown until measured. A node is an id, a ref, a few dozen values and a handful of links. The graph PR adds a benchmark on a synthetic catalog (on the order of 20k persons, 30k events, a few hub places) that reports heap size and proposal time against today's eager canon.

If it's too much, `Values` can hold only kept rank-1 rows and the detail page reads the rest from SQL.

## Plan

Each step is its own PR, measured against the one before.

1. **The write wrapper.** `Catalog.Write`, `database.Tx`, `Changes` and commit listeners; every begin site moved onto it; `RecomputeTx` and `audit.Record` take `*database.Tx`; the `.Begin()` test. No cache yet, so it lands and settles on its own.
2. **The graph and Promote.** `graphcache` on the catalog session, registered as a commit listener; nodes with structure; lazy fill; the revision safety net; `Verify` in CI. Promote reads through `CanonGraph`; stats and candidates move onto the graph; `canonSteps`, the 5-hop expansion and the #327 caps go. The #327 debounce and narrowed exhibits stay. Benchmark.
3. **Display.** Header memos; lists and `…ByIDs` from the graph; the vocabulary map; "today" at read time.
4. **Place chains** from links, replacing #328's per-request index.
5. **Conclusion detail** from node values, with the "Why" outcomes dropped with the display memo.
6. **The Source store**, fed by `Changes.Sources`; Promote's layer, the Evidence graph and the Composer on it.

Later, and separately: the same `Changes` could ride back to Swift on each write's response ("these handles and Sources changed"), so the client's session cache invalidates by handle instead of the hand-kept `CatalogMutation` map. That reuses the commit's `Changes`; it is not another feed.

## Open questions

1. **Where the graph hangs.** A field on `database.Catalog` (low in the import graph, so an opaque slot), or a map in `graphcache` keyed by `*database.Catalog`, set up in `catalogsession.openResearcher` and dropped on close.
2. **Nested writes.** A few write functions call others that open their own transaction today. With the wrapper, inner functions should take the outer `*database.Tx`; find any that can't.
3. **Stats by counts or recompute.** Adjusting value-frequency and fan-out counts from each commit's `Changes` is exact but fiddly. Recomputing from the graph per revision is simpler and may be fast enough in memory.
4. **Detail "Why" outcomes.** Cache them per handle, or keep the SQL read; decide after measuring.
5. **Concurrency.** `Do` serializes everything today. If reads ever run alongside writes, immutable nodes plus a lock around the maps is enough; nothing here should assume serialization beyond that.
