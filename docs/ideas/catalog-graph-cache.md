# One write path and a catalog graph in memory

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

Every write goes through one orchestrator (`writes.Run`) that owns the transaction and its three duties: write the data, record audit, bring derived data up to date. An effects registry, built on the same declared schema `deleteimpact` uses, says what each kind of change touches. ARV is rewritten inside the transaction as today; the memory tier is told after commit.

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
- Adjacency: one link per association, both directions, with its edge signature (bridge type, role or term, neighbor kind, directed). Links are indexed by signature, so following a named hop is a map lookup. The graph builds signatures from `connectrules` and treats them as opaque keys (see Vocabulary stays in registries).
- Members: accepted and provisional member Subject ids (exhibits, memberships, the Source store's back-pointers).

**Links are stored on both ends.** A participation is its own handle, but its link sits on the person node and the event node. When the participation changes, `RecomputeTx` names the association handle, not its endpoints, so the graph keeps the endpoints current itself:

- A reverse index, `holders[association] → loaded nodes whose links came from it`, kept current on every node load and drop.
- When a commit marks an association handle as changed, the graph drops the association's node, every node in `holders[association]` (the old endpoints), and the association's new endpoints, read from the now-committed ARV in one query. A moved endpoint can't keep a stale link, and a new endpoint can't miss one.

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
- **The effects registry** says, per table, what a change to one of its rows touches. It's built on a declared schema model shared with `deleteimpact` and audit.

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
	Effects  effects.Set // handles, header dependents, Sources, vocabulary, search
}

func Run(c *database.Catalog, op Op, fn func(tx *database.Tx) ([]audit.Change, error)) (Result, error) {
	tx := begin(c)
	changes, err := fn(tx)                       // 1. persistent data
	// on error: rollback, nothing else happens
	rev := audit.Record(tx, op, changes)         // 2. audit (skipped for unaudited types)
	fx := effects.Resolve(tx, changes)           //    what the changes touch
	autoreconciler.RecomputeTx(tx, fx.Handles)   // 3a. ARV + handle search documents
	searchindex.Reproject(tx, fx.Search)         //     Source and vocabulary search documents
	commit(tx)
	tx.runAfterCommit()                          //     disk work the write deferred
	c.notify(rev, fx)                            // 3b. memory tier, only after commit
	return Result{Revision: rev, Effects: fx}, nil
}
```

**Work after commit.** Some writes have side effects that must happen only once the commit has succeeded: deleting an artifact removes its file from disk afterwards (`artifacts.go:344`), so a rollback never deletes bytes that are still referenced. A write function registers that work on the transaction, and `Run` runs it after a successful commit and drops it on rollback:

```go
tx.AfterCommit(func() { _ = os.Remove(path) })
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

**Thumbnails.** `core/derivatives` writes `file_derivatives` in its own transactions (`ensure.go`). They go through `Run` like every other write, under an unaudited registry entry with no effects: derived bytes, no research data, nothing the graph reads.

`database.Tx` embeds `*sql.Tx`, and only `writes.Run` creates one. Write functions take `*database.Tx`, so a write can't run outside the orchestrator. A test that fails on any `.Begin()` outside `core/database` and `core/writes` keeps it that way.

### Order inside a write

Derived data is brought up to date once, after the write function returns and before commit. That's equivalent to today: in every write that recomputes (single and batch promote, identity claim create, subject delete, observation write and delete), the recompute is already the last step before commit, and nothing in the write reads ARV after it. Bridge filing in promote reads `identity_claims`, not ARV.

Write functions don't read derived data back. A write returns what it changed and the revision. If the app needs the new state, it reads it with a separate call, which the memory tier serves fresh because it was updated at commit. Writes and reads stay separate calls.

### One schema model under every registry

Three registries in the codebase answer questions about the same graph of tables:

- **`deleteimpact/register.go`** declares every table (kind and bucket: resource, vocabulary, facet, owned, skip) and every foreign key (with its `ON DELETE`). An honesty test compares it against `PRAGMA foreign_key_list`, so it can't drift from the real schema.
- **`audit/scopes.go`** resolves a change to its Source by hand-written walks up those same foreign keys: observation → citation → artifact → source; a note → its citation → …; a subject or artifact directly by `source_id`.
- **The auto-reconciler's handle lookups** (`sqlHandlesForCitation`, `sqlHandlesForProperty`, `HandlesObservingSubject`, …) are hand-written walks too: observation → subject → `identity_claims` → handle, or inbound on `observations.value_subject_id` for "who points at John".

Only the first is declared and checked against the schema. The other two are SQL strings that happen to follow its edges.

So the table and foreign-key declarations move out of `deleteimpact` into a small `schema` package, with the PRAGMA honesty test. `deleteimpact` keeps its own logic (probes, refusals, releases) and reads the model from `schema`. The effects registry is built on the same model.

```go
// package schema: the declared catalog schema, checked against SQLite.

type Table struct {
	Name   string
	PK     string
	Kind   Kind
	Bucket Bucket // resource, vocab, facet, owned, pool, infra, skip
}

type FK struct {
	From, Column string // observations.citation_id
	To           string // citations
	OnDelete     string
	Bucket       Bucket
	Audited      bool // a cascade that official deletes release and audit first
}

var Tables []Table
var FKs []FK
```

### The effects registry

One entry per table, saying what a change to one of its rows touches:

- **Source:** the Sources it belongs to (today's audit scopes).
- **Handles:** the handles whose ARV rows must be rewritten.
- **Vocabulary:** labels (the label map drops) or structure (a term's `directed` or `inverse_key` changed; the graph drops).
- **Search:** the search documents to reproject (a Source's document, a source type's or metadata field's, the Sources of a type). Today 19 calls in write code (`ReprojectSource`, `ReprojectSourceType`, `ReprojectMetadataField`, `ReprojectSourcesForType`) do this by hand; handle documents stay with `RecomputeTx`.
- **Unaudited:** no audit row and no effects; `Run` still owns the transaction.

Where an effect follows the schema, it's written as a path over declared foreign keys. A path can only name edges the model has, so a renamed or missing column fails at startup, not in a user's catalog. Paths are a handful of combinators, not a query language:

```go
// package effects

up("citation_id", "artifact_id", "source_id")        // follow FKs toward the parent
across("identity_claims", "subject_id", "entity_id") // a join table, one side to the other
inbound("observations", "value_subject_id")          // rows that point at this one
sql(`SELECT …`)                                      // anything a path can't say
```

Every path reads the change's fields, old and new. An update that moves an Observation from Subject A to B resolves both. A delete resolves from the old fields, because the row is already gone (audit's scope resolvers do this today with `ghostMap`).

```go
var registry = map[string]Effect{
	"source":      {Source: self(), Search: sourceDoc},
	"source_note": {Source: up("source_id"), Search: sourceDoc},
	"source_type": {Vocabulary: labels, Search: union(selfDoc, sourcesOfType)},
	"citation": {
		Source:  up("artifact_id", "source_id"),
		Handles: onField("certainty", sql(handlesUnderCitation)),
	},
	"observation": {
		Source:  up("citation_id", "artifact_id", "source_id"),
		Handles: from("subject_id", membersOf), // old and new subject_id
	},
	"identity_claim": {
		Handles: union(
			field("entity_id"),
			onStatus("accepted", from("subject_id", observers)), // handles whose members point at the Subject
		),
	},
	"identity_claim_evidence": {Handles: up("identity_claim_id", "entity_id")}, // released pins
	"source_credibility_assessment": {Source: up("source_id"), Handles: from("source_id", sql(handlesInSource))},
	"property":         {Vocabulary: labels, Handles: onField("cardinality", sql(handlesObservingProperty))},
	"property_term": {
		Vocabulary: labels,
		Structure:  onField("directed", "inverse_key"), // relationship signatures change
	},
	"subject_position": {Unaudited: true}, // layout only
	// … every table outside the skip bucket
}

// membersOf: subject → its accepted and provisional handles.
var membersOf = across("identity_claims", "subject_id", "entity_id")

// observers: Subjects whose Observations have this Subject as their value,
// then their handles.
var observers = chain(inbound("observations", "value_subject_id"), field("subject_id"), membersOf)
```

Two kinds of effect stay hand-written, because they're about meaning, not the schema:

- **Effects tied to one field.** A citation touches handles only when `certainty` changes; a property only on `cardinality` (`onField`).
- **Status-dependent effects.** Only an accepted claim moves where subject-valued Observations resolve (`onStatus`).

Header dependents aren't an effects concern. They follow the canonical graph, not foreign keys, and depend on vocabulary (which roles and relationship types a header reads), so they come from the header registry (see Vocabulary stays in registries). `RecomputeTx` computes them from the handles it rewrites, as it does today with `conclusionheaders.HeaderDependents`.

### Completeness, checked against the schema

Audit today fails at runtime when an entity type has no resolver. With the shared model, completeness becomes tests:

- Every table outside the skip bucket has an effect entry, possibly an explicit "none".
- Every path names declared foreign keys.
- Every `ON DELETE CASCADE` into a table whose rows have effects is `Audited`.

A migration that adds a table without deciding its effects fails a test before any code runs.

### Effects the changed rows don't name, and deletes

Most effects come straight from a change's fields: an Observation change names its `subject_id`, so the handles that Subject belongs to are rewritten.

Some don't. A Subject can be the *value* of other Subjects' Observations. A participation says "person: John"; a relationship says "related to: John". ARV stores those values resolved to John's handle. So when John's membership changes (promote, claim accepted, John deleted), every handle with a member whose Observation points at John must be rewritten too, because its value now resolves somewhere else. None of those handles appear in the changed rows: the rows name John and his own handle. That's the `observers` path above (`RecomputeTouchingTx` / `HandlesObservingSubject` today).

Deletes are the case to get right. Effects are resolved after the write, so a resolver can't look up rows that are gone. The rule that makes it work is already enforced by `deleteimpact`:

- Research rows reached by a cascade (notes, identity claims, claim evidence) sit behind `Audited` foreign keys. Official deletes release them explicitly and audit them first (`ReleaseFacets`), and the SQLite cascade is only a backstop. So they come back as changes with their old fields, and the registry resolves them like any other delete.
- The silent cascades (layout, vocabulary joins, name value parts) touch nothing the registry cares about. The cascade test above keeps it that way.
- `ReleaseFacets` returns `released.Handles` today because releasing a pin affects a handle. That's an effect computed in the wrong place; under the registry it falls out of the `identity_claim_evidence` entry, and `released.Handles` goes.

Subject delete collects the Observations pointing at the Subject before it deletes (`inbound`, `ends`). Under the registry those are reported as changes too, and `observers` resolves from their old fields.

This is the part to test hardest. The shadow check (below) catches a missed effect, because ARV would disagree with a full recompute. During the migration, a second check helps: while a write still has its hand-placed recompute calls, a test-only assertion compares the handles it recomputed by hand with the handles the registry resolves for the same changes. A write migrates only when the registry covers everything it did by hand.

### What the memory tier is told

- **Structure:** the handles `RecomputeTx` rewrote. Their nodes reload; for association handles, so do their old and new endpoints (see What a node holds).
- **Display:** those plus `HeaderDependents`. Their display memos clear.
- **Sources:** the resolved Source scopes. The Source store drops those Sources.
- **Vocabulary labels:** drop the label map.
- **Vocabulary structure:** a term's direction or inverse changed. Relationship link signatures include both, so every link using the term is wrong. The graph drops everything; it's rare, and simpler than finding the links.
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

## Vocabulary stays in registries

The rule: **the schema and the infrastructure never name a vocabulary key.** Subjects, Observations, claims and handles are schema, and any layer can work with them. A specific subject type (`person`), role (`subject`), relationship type (`part_of`), event type (`birth`) or Property (`toponym`) is vocabulary, and only registries name it.

### Layers

| Layer | Holds | Vocabulary |
| --- | --- | --- |
| **Schema** (`schema`) | tables, columns, foreign keys, buckets, `CHECK` enums (claim `status`, ARV `reason`) | none |
| **Infrastructure** (`effects`, `writes`, `graphcache`, `canonicalgraph.Walk`, `graphalign`'s walk and scoring) | mechanisms over the schema; kinds, property keys, terms and edge signatures as opaque values | none: it may carry keys, never compare against a literal |
| **Registries** | which keys mean what | all of it |
| **Features** (Promote, the lists, the place page) | use registries by name | through registries |

The effects registry is keyed by table and column, so it stays on the schema side. `onField("certainty")` and `onField("cardinality")` are columns; `onStatus("accepted")` is a `CHECK` enum on `identity_claims.status`. None of its entries name a term, role or kind.

### The registries

Most already exist:

- **`connectrules`:** the bridge kinds (participation, location, relationship, place relationship), which Properties form each association's endpoints, and which Property disambiguates it (role, relationship type, place relationship type). The graph reads this to turn an association handle into links with signatures. It never knows what a participation is.
- **`subjectvocab`:** the seeded subject types, Properties and terms, including direction and inverses (kinship), which the graph uses to sign relationship links.
- **`match` profiles and `graphalign/registry.go`:** which Properties a kind is compared on, and their weights and scales.
- **Named hops** (today in `canonicalgraph`: `EventsOfSubject`, `PlacesOfEvent`, `ParentsOfPlace`, `SuccessorsOfPlace`, …): a vocabulary-bound path (bridge kind + endpoints + term filter), validated against `connectrules` when declared. The generic `Walk` stays in `canonicalgraph`; the declarations move to their own registry file. The graph follows a hop with the same generic call (`g.Follow(h, hops.ParentsOfPlace)`), and place chains name "containment" and "succession" only through it.
- **A header registry** (new): what each kind's header reads. For example: a Person reads its subject-role events of the birth and death types; an Event reads its places through location and their containment chain; a Place reads its containment parents. Header composition reads it, and header dependents are derived from it by reversing each read. That replaces `HeaderDependents`, whose walks hard-code the same knowledge today, and means a new header field can't forget its dependents.

### Leaks today

Places where infrastructure names keys now, to fix along the way:

- **Promote stats** (`promotealign/stats.go`) hard-codes `p.key IN ('name', 'sex_at_birth', 'event_type', 'toponym')`. It should read the Properties from the `match` profiles. Fixed when stats move onto the graph.
- **`HeaderDependents`** hard-codes which hops a header reads. Replaced by the header registry.
- **`canonSteps` and the layer hops** in `promotealign` list hops in code. `canonSteps` goes with the lazy walk; the layer's hops come from `connectrules`.
- **Named hops** sit next to the generic `Walk` in `canonicalgraph.go`. They move to a registry file.

### Enforcing it

**Typed keys, so the compiler does most of it.** Registries export keys as distinct types (`subjectvocab.Kind`, `subjectvocab.PropertyKey`, `subjectvocab.TermKey`, `connectrules.BridgeKind`), and infrastructure APIs accept only those types: `g.Kind(k subjectvocab.Kind)`, signatures built from `TermKey`s. Infrastructure can carry and compare keys it was handed, but writing a key means converting a string literal, which stands out in review and is caught by a narrow test.

**A narrow literal test.** A plain scan for seeded keys would be noisy: `person`, `event`, `place`, `subject` and `name` also appear in table names, column names and JSON. So the test only fails on conversions to the key types (`subjectvocab.TermKey("part_of")`) outside registry files and tests. Those are the only way to name a key, so the test has no common-word false positives.

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
| Place chains and place detail | `placeGraph` per request; `ParentsAtDate`, `PartsAtDate`, `SuccessionNames`, `PlaceDetail` | following the containment and succession hops from the hops registry; replaces #328's per-request index |
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

Each step is its own PR, measured against the one before. PRs that only repoint callers are kept apart from PRs that add behavior, so the churn reviews as churn. Open questions sit under the PR that has to answer them.

**Before starting**

- **The paused stack.** [mendahu/provenencia#331](https://github.com/mendahu/provenencia/pull/331) to [mendahu/provenencia#334](https://github.com/mendahu/provenencia/pull/334) (rescoring, directed duplicates, neutral terms, graphalign cleanup) all edit `align.go`, which PR 8 restructures. Decide whether they land first or are redone after PR 8. [mendahu/provenencia#327](https://github.com/mendahu/provenencia/pull/327) is replaced by PR 8 (its debounce and narrowed exhibits carry over), and [mendahu/provenencia#328](https://github.com/mendahu/provenencia/pull/328) by PR 10. [mendahu/provenencia#330](https://github.com/mendahu/provenencia/pull/330) batches readers that PRs 8–11 replace; decide whether it's still worth landing.

**Schema and effects**

1. **Extract the schema model** (churn). Tables and foreign keys move from `deleteimpact/register.go` into `schema`, with the PRAGMA honesty test; `deleteimpact` reads them from there. No behavior change.
   - *Open:* where `schema` sits. It must stay below `audit`, `deleteimpact`, `effects` and `writes` in the import graph; `database` is the likely parent.
2. **Paths and the effects registry** (logic). The path combinators; an entry for every table outside the skip bucket, including search and vocabulary effects; audit's scope resolvers re-expressed as `Source` paths, with a test that they resolve the same Sources as the old resolvers on the existing fixtures; the completeness tests. Nothing calls the `Handles` side yet.
   - *Open:* claim statuses in paths. The recompute SQL counts `accepted` and `provisional` claims as members; `across("identity_claims", …)` must filter the same way. Decide whether status filters are a path option or part of each declaration.

**Writes**

3. **The orchestrator** (logic). `writes.Run`, `database.Tx`, `Op`, `Result`, `AfterCommit`, commit listeners; `audit.Record` resolves scopes through the registry; `RecomputeTx` takes `*database.Tx`; search reprojection from `Effects.Search`; the migration assertion, covering handles and search documents. Property terms migrated as the pilot.
   - *Open:* `Run`'s return type. Write functions return more than changes (the created entity, `BatchResult.Written`, claims and pins). Either `Run` is generic (`Run[T]`) or write functions return a result struct that carries the changes. Settle before the churn PRs, since every one follows it.
   - *Open:* nested writes. A few write functions call others that open their own transaction today. Under `Run`, inner functions take the outer `*database.Tx`; find any that can't.
4. **Migrate Source-layer writes** (churn). Sources, notes, metadata, source types, metadata fields, artifacts (with `AfterCommit` for file removal), ingest, thumbnails.
   - *Open:* confirm thumbnail generation runs inside `catalogsession.Do`. If it doesn't, that's an existing serialization bug to fix here, not paper over.
5. **Migrate Evidence-layer writes** (churn). Citations, observations, subjects, name values, connect, positions.
6. **Migrate conclusion-layer writes** (churn). Promote single and batch, identity claims, canonical entities, credibility, properties, subject definitions. Includes the transactions in FFI handlers (`api/ffi/handlers/subject_defs.go`, `delete_impact.go`).
7. **Close the old path** (small). Hand-placed recompute and reproject calls, `released.Handles`, the old scope resolvers and the migration assertion go; old entry points become private; the `.Begin()` test; the typed-key literal test over `schema`, `effects` and `writes` (later PRs extend it to `graphcache`).

**The graph**

8. **The graph and Promote.** `graphcache`, registered as a commit listener; nodes with structure; the association reverse index; lazy fill; the revision safety net; `Verify` in CI. Promote reads through `CanonGraph`; stats (Properties from the `match` profiles) and candidates move onto the graph; `canonSteps`, the 5-hop expansion and the #327 caps go. The #327 debounce and narrowed exhibits carry over. Benchmark.
   - *Open:* where the graph hangs. Leaning toward `database.Catalog` (a no-op when unused), because domain-package tests open catalogs directly: a graph on the session would leave `Verify` covering only FFI-level tests.
   - *Open:* cold-cache cost. Align asks for one handle's neighbors at a time, so a cold walk through a hub is hundreds of queries while the session lock blocks every other call. Benchmark cold as well as warm; if cold is slow, give `CanonGraph` a prefetch hint ("load the neighbors of this frontier").
   - *Open:* proposal changes. The lazy walk has no hop limit, `Seeds(kind)` must reproduce today's top-5 property-ranked seeds, and `Links` must keep today's order (association ref, then handle ref) or ties resolve differently. Record proposals on the fixtures and the synthetic catalog before switching, and review every difference.
   - *Open:* memory. ARV holds every candidate value at every rank. Measure real rows per handle; if nodes are too big, keep only kept rank-1 values and let the detail page read the rest from SQL.
   - *Open:* stats by counts or recompute. Adjusting value-frequency and fan-out counts from each commit's effects is exact but fiddly; recomputing from the graph per revision is simpler and may be fast enough in memory.
   - *Open:* catalog-defined vocabulary. Researcher terms and plugin vocabulary live in `property_terms`, not the seed registry. The graph reads direction and inverses from the catalog; the literal test can only cover seeded keys.
   - *Open:* concurrency. `Do` serializes everything today. If reads ever run alongside writes, immutable nodes plus a lock around the maps is enough; nothing should assume serialization beyond that.
9. **Display.** The header registry, with header dependents derived from it (replacing `HeaderDependents`); header memos; lists and `…ByIDs` from the graph; the vocabulary map; "today" at read time.
   - *Open:* handles with no members. No code deletes from `canonical_entities`; a handle that loses its last member just has no ARV rows. `Kind(kind)` must apply the lists' membership filter, and membership changes must add a handle to or remove it from the kind index.
   - *Note:* this makes Go reads fast and correct, but Swift's own invalidation gaps (deleting an artifact leaves citations and the conclusion lists stale) stay until `Result.Effects` reaches Swift.
10. **Place chains** from named hops on the graph, replacing #328's per-request index. Named hops move to their own registry file.
11. **Conclusion detail** from node values, with the "Why" outcomes dropped with the display memo.
    - *Open:* cache the per-Observation outcomes per handle, or keep the SQL read; decide after measuring.
12. **The Source store**, fed by resolved Source scopes; Promote's layer, the Evidence graph and the Composer on it.

Later, and separately: `Result.Effects` rides back to Swift on each write's response, and the client's session cache invalidates by handle and Source instead of the hand-kept `CatalogMutation` map.
