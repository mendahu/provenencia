# A catalog graph in memory, and one write path

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

Every write goes through one orchestrator (`writes.Run`) that owns the transaction and its three duties: write the data, record audit, bring derived data up to date. An effects registry says what each kind of change touches. ARV is rewritten inside the transaction as today; the memory tier is told after commit.

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

- A memo built from the node and its neighbors the first time a header or detail is asked for, and dropped when the node or a neighbor it reads changes (see Writes). It is derived from structure, never filled independently, so it cannot drift from it.
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

Writes update the graph as they commit (see Writes), so a read never has to check for staleness first.

## Writes: one orchestrator

Every write the app makes has three duties:

1. **Write the persistent data** (the rows the operation is about).
2. **Record audit** (`audit.Record`: the revision, and each row's old and new fields).
3. **Bring derived data up to date:** the persistent tier (ARV and the search index, inside the transaction) and the memory tier (after commit).

Today each write function does all three by hand: about 45 transaction sites, 40 `audit.Record` calls and 13 hand-placed `Recompute…Tx` calls, each write choosing which derived data it affects. Duty 3 is the fragile one.

The split:

- **Write functions** write their rows and return what they changed, as `[]audit.Change`. No transaction handling, no audit call, no recompute.
- **The orchestrator** (`writes.Run`) owns the transaction and the order of the duties.
- **The effects registry** says, per entity type, what a change to it touches.

The frontend doesn't change. It already names the operation it wants (`UpdateObservation`, `DeleteSubject`, …). It does not also declare what it's updating: that would be a second description of the write that could disagree with what the write did, which is what the Swift `CatalogMutation` map is today. What changed comes from the write function's own changes, with old and new values, the same data audit already records.

### The orchestrator

```go
// package writes: the one write path between the FFI and domain code.

type Op struct {
	Action      string // audit action_type, e.g. "delete_observation"
	Description string
	UserID      []byte
}

type Result struct {
	Revision int64
	Effects  effects.Set // handles, header dependents, Sources, vocabulary
}

func Run(c *database.Catalog, op Op, fn func(tx *database.Tx) ([]audit.Change, error)) (Result, error) {
	tx := begin(c)
	changes, err := fn(tx)                       // 1. persistent data
	// on error: rollback, nothing else happens
	rev := audit.Record(tx, op, changes)         // 2. audit (skipped for unaudited types)
	fx := effects.Resolve(tx, changes)           //    what the changes touch
	autoreconciler.RecomputeTx(tx, fx.Handles)   // 3a. ARV + search, inside the transaction
	commit(tx)
	c.notify(rev, fx)                            // 3b. memory tier, only after commit
	return Result{Revision: rev, Effects: fx}, nil
}
```

A write function, before and after:

```go
// before: owns its transaction, records audit, picks its recomputes
func DeleteObservation(c *database.Catalog, userID, id []byte) error {
	tx, err := db.Begin()
	// … read prev, refuse, release facets, delete …
	autoreconciler.RecomputeSubjectsTx(tx, [][]byte{prev.SubjectID})
	autoreconciler.RecomputeTx(tx, released.Handles)
	audit.Record(tx, audit.Revision{ActionType: "delete_observation", Changes: changes})
	return tx.Commit()
}

// after: writes and reports
func DeleteObservation(tx *database.Tx, id []byte) ([]audit.Change, error) {
	// … read prev, refuse, release facets, delete — as today …
	return append(released.Changes, audit.Change{
		EntityType: "observation", EntityID: id, Action: audit.ActionDelete,
		Fields: audit.DeletedRow(observationRowMap(prev)),
	}), nil
}

// the FFI handler
res, err := writes.Run(c, writes.Op{Action: "delete_observation", UserID: user},
	func(tx *database.Tx) ([]audit.Change, error) {
		return observations.DeleteObservation(tx, id)
	})
```

`database.Tx` embeds `*sql.Tx`, and only `writes.Run` creates one. Write functions take `*database.Tx`, so a write can't run outside the orchestrator. A test that fails on any `.Begin()` outside `core/database` and `core/writes` keeps it that way.

### Order inside a write

Derived data is brought up to date once, after the write function returns and before commit. That's equivalent to today: in every write that recomputes (single and batch promote, identity claim create, subject delete, observation write and delete), the recompute is already the last step before commit, and nothing in the write reads ARV after it. Bridge filing in promote reads `identity_claims`, not ARV.

Write functions don't read derived data back. A write returns what it changed and the revision. If the app needs the new state, it reads it with a separate call, which the memory tier serves fresh because it was updated at commit. Writes and reads stay separate calls.

### The effects registry

Audit already has half of this: `audit/scopes.go` maps every audited entity type to the Sources a change belongs to, and `Record` rejects a type without a resolver. The effects registry extends each entry with the handles it touches and whether it's vocabulary, so one table keyed by entity type answers every "what does this affect" question. A new audited table has to decide all of its effects in one place.

```go
// package effects

type Effect struct {
	Source     sourceResolver // today's audit scope resolver, moved here
	Handles    handleResolver // handles whose ARV rows must be rewritten
	Vocabulary bool           // drop the label map
	Unaudited  bool           // no audit row; Run still owns the transaction
}

var registry = map[string]Effect{
	"observation": {
		Source:  viaCitation,
		Handles: membersOf("subject_id"), // old and new subject_id from the change's fields
	},
	"citation": {
		Source:  viaArtifact,
		Handles: onField("certainty", handlesUnderCitation),
	},
	"source_credibility_assessment": {Source: direct, Handles: handlesInSource},
	"identity_claim":          {Handles: claimEntityAndObservers},
	"identity_claim_evidence": {Handles: claimEntity}, // released pins
	"property":                {Vocabulary: true, Handles: onField("cardinality", handlesObservingProperty)},
	"property_term":           {Vocabulary: true},
	"subject_position":        {Unaudited: true}, // layout only
	// … every audited type
}
```

For an unaudited type, `Run` skips `audit.Record` and resolves no effects. It still owns the transaction, so every write goes through the same path: subject positions, type–property bindings and onboarding seeds included.

Header dependents aren't a registry concern. `RecomputeTx` already computes them from the handles it rewrites (`conclusionheaders.HeaderDependents`), and returns them for the memory tier.

### Effects the changed rows don't name

Most effects come straight from a change's fields: an Observation change names its `subject_id`, so the handles that Subject belongs to are rewritten.

Some don't. A Subject can be the *value* of other Subjects' Observations. A participation says "person: John"; a relationship says "related to: John". ARV stores those values resolved to John's handle. So when John's membership changes (promote, claim accepted, John deleted), every handle with a member whose Observation points at John must be rewritten too, because its value now resolves somewhere else. None of those handles appear in the changed rows. The rows name John and his own handle. This is `RecomputeTouchingTx` / `HandlesObservingSubject` today.

The registry entry for `identity_claim` expresses it as a lookup: "Observations whose value is this Subject → their Subjects → those Subjects' handles". That's an ordinary SQL resolver, like the scope resolvers.

The hard case is deletes. Effects are resolved after the write, and by then a deleted Subject's inbound Observations may be gone too, so the lookup finds nothing. Subject delete handles this today by collecting `inbound` and `ends` before it deletes. With the registry, every row a delete removes must come back as an `audit.Change` with its old fields (the deleted Observation, with its old `subject_id` and value). The resolver then reads the old fields, as audit's scope resolvers already do for deleted rows (`ghostMap`). Rows removed by `ON DELETE CASCADE` without being reported would be invisible, so a delete reports what it cascades, or its effects are resolved from the change before the row goes.

This is the part to test hardest. The shadow check (below) catches a missed effect: ARV would disagree with a full recompute.

During the migration, a second check helps: while a write still has its hand-placed recompute calls, a test-only assertion compares the handles it recomputed by hand with the handles the registry resolves for the same changes. A write migrates only when the registry covers everything it did by hand.

### What the memory tier is told

- **Structure:** the handles `RecomputeTx` rewrote. Their nodes reload.
- **Display:** those plus `HeaderDependents`. Their display memos clear.
- **Sources:** the resolved Source scopes. The Source store drops those Sources.
- **Vocabulary:** drop the label map.
- **All:** `Rebuild` (open-time, on a `CacheVersion` change).

Only after commit. A rollback drops it with the transaction, so the memory tier never sees uncommitted data and never misses a committed change. `catalogsession.Do` serializes every operation, so the next read can't start before the cache is updated, with no lock needed.

`Result.Effects` is also what a write's FFI response could carry back to Swift later ("these handles and Sources changed"), so the client's session cache invalidates by handle instead of the hand-kept `CatalogMutation` map. It is the commit's outcome, not a stored feed: audit stays the record of what changed, and ARV the record of what the graph is.

### Alternatives considered

- **A marks table** written by `RecomputeTx` in the same transaction, drained on read. Exact and needs no restructure, but adds a table and a query before every read.
- **Mark memory directly from `RecomputeTx`.** Can't know whether the transaction commits.
- **A transaction wrapper only** (`Catalog.Write`), with write functions still calling audit and recompute themselves. Exact, but leaves duty 3 scattered across every write.
- **Drop everything on any revision change.** Fine for Promote, not once lists read the cache.
- **A persisted change feed.** Rejected: it would duplicate audit and ARV.

### What isn't in ARV

These change without `RecomputeTx`, so nodes don't copy them. They are resolved when read:

- **Term and property labels.** A term rename changes every header that shows it. Nodes hold term and property ids; a small vocabulary map (labels, value types, kinship and inverses) is dropped whole on any `properties` or `property_terms` write.
- **Source titles** in the detail page's "Why". Held as Source ids; looked up when read.
- **Today.** Place headers use `TodayDate()` for current parents. The memo keeps the dated periods and settles "today" when read.

### Safety net

The memory tier records the audit revision of the last commit it heard about. If a read finds a newer revision it never heard about (a write that slipped past `writes.Run`), it drops everything and logs it. That log line is a bug report. With the compiler and the `.Begin()` test in place it should never fire.

### Verifying it

A shadow check makes a missed invalidation fail a test instead of showing stale data:

```go
// Verify rebuilds every loaded node from SQL and compares. Tests run it after
// each write when PROVENENCIA_GRAPH_VERIFY=1; CI sets it for the core suite.
func (g *Graph) Verify(q Querier) error
```

The existing suites exercise every write path, so turning the check on in CI covers them all. A debug build can run it in the app while dogfooding.

## The Source store

Same pattern. The effects registry resolves every write's Source scopes (the resolvers audit uses today), and the Source store drops those Sources when the write commits.

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

Each step is its own PR, measured against the one before. PRs that only repoint callers are kept apart from PRs that add behavior, so the churn reviews as churn.

**Writes**

1. **The orchestrator and registry** (logic). `writes.Run`, `database.Tx`, `Op`, `Result`, commit listeners; the effects registry for every audited type, with audit's scope resolvers moved into it; `audit.Record` and `RecomputeTx` take `*database.Tx`; the migration assertion. One small write path (property terms) migrated as the pilot.
2. **Migrate Source-layer writes** (churn). Sources, notes, metadata, source types, metadata fields, artifacts, ingest.
3. **Migrate Evidence-layer writes** (churn). Citations, observations, subjects, name values, connect, positions.
4. **Migrate conclusion-layer writes** (churn). Promote single and batch, identity claims, canonical entities, credibility, properties, subject definitions.
5. **Close the old path** (small). Hand-placed recompute calls and the migration assertion go; old entry points become private; the `.Begin()` test.

**The graph**

6. **The graph and Promote.** `graphcache` on the catalog session, registered as a commit listener; nodes with structure; lazy fill; the revision safety net; `Verify` in CI. Promote reads through `CanonGraph`; stats and candidates move onto the graph; `canonSteps`, the 5-hop expansion and the #327 caps go. The #327 debounce and narrowed exhibits stay. Benchmark.
7. **Display.** Header memos; lists and `…ByIDs` from the graph; the vocabulary map; "today" at read time.
8. **Place chains** from links, replacing #328's per-request index.
9. **Conclusion detail** from node values, with the "Why" outcomes dropped with the display memo.
10. **The Source store**, fed by resolved Source scopes; Promote's layer, the Evidence graph and the Composer on it.

Later, and separately: `Result.Effects` rides back to Swift on each write's response, and the client's session cache invalidates by handle and Source instead of the hand-kept `CatalogMutation` map.

## Open questions

1. **Where the graph hangs.** A field on `database.Catalog` (low in the import graph, so an opaque slot), or a map in `graphcache` keyed by `*database.Catalog`, set up in `catalogsession.openResearcher` and dropped on close.
2. **Nested writes.** A few write functions call others that open their own transaction today. Under `writes.Run`, inner functions take the outer `*database.Tx`; find any that can't.
3. **Cascaded deletes.** List every `ON DELETE CASCADE` that removes a row the registry needs to see (Observations under a deleted Subject or citation), and make the delete report it.
4. **Stats by counts or recompute.** Adjusting value-frequency and fan-out counts from each commit's effects is exact but fiddly. Recomputing from the graph per revision is simpler and may be fast enough in memory.
5. **Detail "Why" outcomes.** Cache them per handle, or keep the SQL read; decide after measuring.
6. **Concurrency.** `Do` serializes everything today. If reads ever run alongside writes, immutable nodes plus a lock around the maps is enough; nothing here should assume serialization beyond that.
