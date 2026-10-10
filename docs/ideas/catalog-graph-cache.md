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

Every write goes through one orchestrator (`writes.Run`) that owns the transaction and its three duties: write the data, record audit, bring derived data up to date. An effects registry, built on `catalogmodel` (the same declared model `deleteimpact` reads), says what each kind of change touches. ARV is rewritten inside the transaction as today; the memory tier is told after commit.

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

- **Write functions** write their rows and return what they changed, as `[]rowchange.Change`. No transaction handling, no audit call, no recompute.
- **The orchestrator** (`writes.Run`) owns the transaction and the order of the duties.
- **The effects registry** says, per table, what a change to one of its rows touches. It's built on `catalogmodel`, shared with `deleteimpact` and audit.

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

func Run[T any](c *database.Catalog, op Op, fn func(tx *database.Tx) (T, []rowchange.Change, error)) (T, Result, error) {
	tx := begin(c)
	value, changes, err := fn(tx)                // 1. persistent data
	// on error: rollback, nothing else happens
	rev := audit.Record(tx, op, changes)         // 2. audit (skipped for unaudited types)
	fx := effects.Resolve(tx, changes)           //    what the changes touch
	autoreconciler.RecomputeTx(tx, fx.Handles)   // 3a. ARV + handle search documents
	searchindex.Reproject(tx, fx.Search)         //     Source and vocabulary search documents
	commit(tx)
	tx.runAfterCommit()                          //     disk work the write deferred
	c.notify(rev, fx)                            // 3b. memory tier, only after commit
	return value, Result{Revision: rev, Effects: fx}, nil
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
func DeleteObservation(tx *database.Tx, id []byte) ([]rowchange.Change, error) {
	// … read prev, refuse, release facets, delete — as today …
	return append(released.Changes, rowchange.Change{
		EntityType: "observation", EntityID: id, Action: rowchange.ActionDelete,
		Fields: rowchange.DeletedRow(observationRowMap(prev)),
	}), nil
}

// the FFI handler
_, res, err := writes.Run(c, writes.Op{Action: "delete_observation", UserID: user},
	func(tx *database.Tx) (struct{}, []rowchange.Change, error) {
		changes, err := observations.DeleteObservation(tx, id)
		return struct{}{}, changes, err
	})
```

**Thumbnails.** JPEG generation and the object write stay outside the transaction. The catalog insert returns a `file_derivative` change and leaves the derived `files` row out of the list. `file_derivatives` is `{Entity: "file_derivative", None: true}`: derived bytes, no research data, nothing the graph reads. A change list whose every row is `None` commits with no revision. `AfterCommit` refreshes the parent Source's search document. A unique checksum conflict ends that `Run`, then `Ensure` starts a second `Run` for the link only.

`database.Tx` embeds `*sql.Tx`, and only `writes.Run` creates one. Write functions take `*database.Tx`, so a write can't run outside the orchestrator. A test that fails on any `.Begin()` outside `core/database` and `core/writes` keeps it that way.

### Order inside a write

Derived data is brought up to date once, after the write function returns and before commit. That's equivalent to today: in every write that recomputes (single and batch promote, identity claim create, subject delete, observation write and delete), the recompute is already the last step before commit, and nothing in the write reads ARV after it. Bridge filing in promote reads `identity_claims`, not ARV.

Write functions don't read derived data back. A write returns what it changed and the revision. If the app needs the new state, it reads it with a separate call, which the memory tier serves fresh because it was updated at commit. Writes and reads stay separate calls.

### One catalog model under every registry

Three registries in the codebase answer questions about the same graph of tables:

- **`core/database/catalogmodel`** declares every table (kind and bucket: resource, vocabulary, facet, owned, skip) and every foreign key (with its `ON DELETE`). An honesty test compares it against `PRAGMA foreign_key_list`, so it can't drift from the real schema.
- **`audit/scopes.go`** resolves a change to its Source by hand-written walks up those same foreign keys: observation → citation → artifact → source; a note → its citation → …; a subject or artifact directly by `source_id`.
- **The auto-reconciler's handle lookups** (`sqlHandlesForCitation`, `sqlHandlesForProperty`, `HandlesObservingSubject`, …) are hand-written walks too: observation → subject → `identity_claims` → handle, or inbound on `observations.value_subject_id` for "who points at John".

Only the first is declared and checked against the schema. The other two are SQL strings that happen to follow its edges.

Those declarations live in `catalogmodel`, with the PRAGMA honesty test. `deleteimpact` keeps probes, refusals, and releases, and reads the model from `catalogmodel`. The effects registry is built on the same model.

```go
// package catalogmodel: the declared catalog model, checked against SQLite.

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
- **None:** no audit row and no effects. A change list whose every row is `None` still commits, then runs `AfterCommit`. An empty list rolls back.

Where an effect follows the schema, it's written as a path over declared foreign keys. A path can only name edges the model has, so a renamed or missing column fails at startup, not in a user's catalog. Paths are a handful of combinators, not a query language:

```go
// package effects

up("citation_id", "artifact_id", "source_id")        // follow FKs toward the parent
across("identity_claims", "subject_id", "entity_id") // a join table, one side to the other
inbound("observations", "value_subject_id")          // rows that point at this one
sql(`SELECT …`)                                      // anything a path can't say
```

If the path's column is on the diff, the walk uses old and new, so a move resolves both ids. If the column is absent, the walk reads the live row. Ghosts (the delete change's old fields) apply only when the row is already gone. A polarity edit is the same rule as any other sparse update: `polarity` is on the diff, `citation_id` is not, and the live row still names the citation.

```go
var registry = map[string]Effect{
	"source":      {Source: self(), Search: sourceDoc},
	"source_note": {Source: up("source_id"), Search: sourceDoc},
	"source_type": {Vocabulary: labels, Search: union(selfDoc, sourcesOfType)},
	"citation": {
		Source:  up("artifact_id", "source_id"),
		Handles: onField("transcription_uncertain", sql(handlesUnderCitation)),
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
	"source_credibility_assessment": {Source: up("source_id"), Handles: onField("credibility_grade_id", from("source_id", sql(handlesInSource)))},
	"property":         {Vocabulary: labels, Handles: onField("cardinality", sql(handlesObservingProperty))},
	"property_term": {
		Vocabulary: labels,
		Structure:  onField("directed", "inverse_key"), // relationship signatures change
	},
	"subject_position": {Entity: "subject_position", None: true}, // layout only
	// … every table outside the skip bucket
}

// membersOf: subject → its accepted and provisional handles.
// The status filter is part of this declaration, not a default inside across.
var membersOf = across("identity_claims", "subject_id", "entity_id", "accepted", "provisional")

// observers: Subjects whose Observations have this Subject as their value,
// then their handles.
var observers = chain(inbound("observations", "value_subject_id"), field("subject_id"), membersOf)
```

Two kinds of effect stay hand-written, because they're about meaning, not the schema:

- **Effects tied to one field.** A citation touches handles only when `transcription_uncertain` changes; a credibility assessment only when `credibility_grade_id` changes; a property only on `cardinality` (`onField`).
- **Status-dependent effects.** `onStatus("accepted")` runs when either the old or the new status is `accepted`, so a new accepted claim and a later accepted→rejected edit both recompute observers. A diff that omits `status` did not change it.

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
| **Catalog model** (`catalogmodel`) | tables, columns, foreign keys, buckets, `CHECK` enums (claim `status`, ARV `reason`) | none |
| **Infrastructure** (`effects`, `writes`, `graphcache`, `canonicalgraph.Walk`, `graphalign`'s walk and scoring) | mechanisms over the schema; kinds, property keys, terms and edge signatures as opaque values | none: it may carry keys, never compare against a literal |
| **Registries** | which keys mean what | all of it |
| **Features** (Promote, the lists, the place page) | use registries by name | through registries |

The effects registry is keyed by table and column, so it stays on the catalog-model side. `onField("transcription_uncertain")` and `onField("cardinality")` are columns; `onStatus("accepted")` is a `CHECK` enum on `identity_claims.status`. None of its entries name a term, role or kind.

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

- **The paused stack.** Decided: [mendahu/provenencia#327](https://github.com/mendahu/provenencia/pull/327) to [mendahu/provenencia#340](https://github.com/mendahu/provenencia/pull/340) stay paused until this plan has landed, then each is refactored and rebased onto it in turn. #327 is replaced by PR 8 (its narrowed exhibits carry over; the debounce does not), #328 by PR 10, and #330 is re-judged against the graph-backed readers. #331 to #334 are rebuilt on the `CanonGraph` version of `align.go`.

**Catalog model and effects**

1. **Extract the catalog model** (done). Tables and foreign keys live in `core/database/catalogmodel` (`Tables`, `FKs`, `Kind`, `Bucket`), with the PRAGMA honesty test. `deleteimpact` reads them from there and keeps probes and releases. No behavior change.
2. **Paths and the effects registry** (done). `core/database/effects` has the path combinators and one `Effect` per table outside the skip bucket. `RecomputeTx` does not call it. `audit.Record` stores source scopes from `effects.Sources` (PR 3).
   - **One `Effect` per table**, with a field per job: `Source`, `Handles`, `Search`, `Vocabulary`, `Structure`. An empty field means that job is unaffected. `None: true` is an explicit empty entry (`users`, layout-only `subject_positions`, and the grade tables). An identity claim keeps an empty `Source` (today's `noScope`) and a `Handles` path. A later cache Source scope can be another field. It is not an audit scope.
   - **`Change` lives in `core/database/rowchange`** (`Change`, `FieldDiff`, `FullRow`, `DeletedRow`, and the action constants). `Revision` and `Record` stay in `audit`. Writers say `rowchange.Change`. No aliases. `effects` imports `rowchange` and `catalogmodel`. It does not import `audit`.
   - **Lookup.** If the path's column is on the diff, use old and new. If it is absent, read the live row. Ghosts apply only when the row is already gone. This matches `scopeResolver.via` and `directSource`.
   - **Claim membership** is a filter on the declaration. `membersOf` is `across("identity_claims", "subject_id", "entity_id")` restricted to `accepted` and `provisional`.
   - **`onStatus` fires when either the old or the new status is `accepted`.** A diff that omits `status` does not run the inner path. There is no status-edit writer yet; a unit test covers both sides.
   - **Trigger columns.** Citation handles use `onField("transcription_uncertain")`. A credibility assessment recomputes that Source's handles only on `onField("credibility_grade_id")`.
   - **Grades are seed-stable.** No `Handles` fan-out from `source_credibility_grades` or `claim_confidence_grades`. Source credibility grades are `SeededLocked`, same as properties, so a provenencia grade cannot be deleted. Claim confidence grades have no delete path.
   - **Name and date values** have entries. Source walks inbound `observations` on `value_name_id` / `value_date_id`, then up to the source. Handles walk that same inbound, then `membersOf`. `observations.value_date_id` and `value_name_id` are unique, so one observation owns a value today; the path is still inbound and returns every row that points at it. `sql()` stays for a walk a combinator cannot say.

**Writes**

3. **The orchestrator** (done). `writes.Run`, `database.Tx`, `Op`, `Result`, `AfterCommit`, and commit listeners. Property-term create, update, and delete go through `Run`. Source-layer writes follow in PR 4.
   - **`Run` is generic.** `Run[T any](c, op, fn func(tx *database.Tx) (T, []rowchange.Change, error)) (T, Result, error)`. A delete uses an empty struct as `T`. The closure's `T` is the written value (the term, or nothing). `Result` is the revision plus `effects.Set`.
   - **The FFI handler calls `Run`.** `Create`, `Update`, and `Delete` take `*database.Tx` and return the changes. They do not begin, commit, or call `audit.Record`. `Upsert` stays the un-audited seed path. Tests that call those three call `Run` too.
   - **Re-entry is refused.** A second `Run` on the same catalog, including from `AfterCommit` or a listener, returns `database.ErrWriteReentry`. `Run` does not take the catalog session lock; the handler already holds it.
   - **A failed notify drops every listener.** Commit has already succeeded. Any `OnCommit` error calls `Drop` on every listener. There is no production listener yet. `autoreconciler.Rebuild` does not go through `Run` and is not wired to the hook. `Drop` is what PR 8 will use to publish "drop everything."
   - **`RecomputeTx` keeps its `Querier` parameter.** `*database.Tx` embeds `*sql.Tx`, so `Run` passes the inner transaction. `Run` calls it only when the handle set is non-empty.
   - **Nested writes are out of this PR.** Property terms do not call another write. Later migrations pass the outer `*database.Tx` into the inner function. Source-layer writes move in PR 4. Evidence-layer writes move in PR 5. Conclusion writes still begin their own transaction.
   - **Stored scopes come from `effects.Sources`.** `Record` still rejects an entity type with no scope resolver. The old resolvers stay callable (`audit.LegacySourceIDs`) so the parity test can compare them until PR 7 deletes them. `searchindex.Reproject` dispatches `Effects.Search` to the existing source, source-type, metadata-field, and sources-for-type reprojectors. Property terms have no handles and no search documents. A label edit sets vocabulary and leaves structure unset.
4. **Migrate Source-layer writes** (done). Sources, notes, metadata, source types, metadata fields, artifacts, ingest, and thumbnails go through `writes.Run`.
   - **The FFI handler calls `Run`.** Domain functions take `*database.Tx` and return the changes. They do not begin, commit, audit, or reproject. `ingest.File` and `derivatives.Ensure` call `Run` themselves: the object write stays outside, and a checksum conflict starts a second `Run` after the first returns.
   - **Thumbnails commit without a revision.** The link is a `file_derivative` change and the derived `files` row stays off the list, so the batch is `None`-only: `Run` skips audit, effects, and listeners, commits, then runs `AfterCommit` (the Source search refresh). An empty change list still rolls back. PR 5 retired `Op.Unaudited`; the rule is the change list.
   - **Artifact search is in the registry.** `artifacts` has a Source search document via `field("source_id")`, so create, update, and delete refresh the parent Source. A deleted source still drops its document, because `ReprojectSource` deletes the doc when the row is gone.
   - **Seed `Upsert` stays off `Run`.** Source-type and metadata-field `Create` and `Update` go through `Run` and gain a revision with a null user. `Upsert` remains the un-audited seed path and still reprojects by hand.
   - **`SetCover` does not nest.** The raster check, including `EnsureThumbnail`, finishes before the cover `Run`.
   - **File bytes move to `AfterCommit`.** `artifacts.Delete` unlinks released objects after commit, so a rollback leaves the files in place.
   - **Thumbnails already run inside the session.** `EnsureFileThumbnail` uses `withProjectCatalog` (`catalogsession.Do`). `SetCover`'s handler holds the same session. Dismiss and reorder refresh the Source search document because their layout changes already name one.
5. **Migrate Evidence-layer writes** (done). Citations, observations, subjects, connect, and positions go through `writes.Run`. Name and date inserts stay on the observation's transaction.
   - **The FFI handler calls `Run`.** Domain functions take `*database.Tx` and return the changes. They do not begin, commit, audit, recompute, or reproject. `InsertManyTx` no longer recomputes, so create, add, and connect recompute once, in `Run`. An unchanged citation, observation, or subject returns no changes and `Run` rolls it back.
   - **Handle fixtures.** An in-place name edit, an in-place date edit, a polarity-only observation edit, and a `value_subject_id` move (`subject_id` omitted) resolve the same handles as `RecomputeSubjectsTx`. A citation update that flips `transcription_uncertain` resolves the citation's handles. A locator-only update resolves none. The registry already matched.
   - **Subject delete.** `effects` cannot see the released facet rows, and it cannot import `conclusionheaders`. `Delete` snapshots linked handles and `HeaderDependents` while the link exists, then calls `RecomputeTx` once after the rows are gone. That call also includes the released handles and the handles that observed the subject: header search documents are written at the end of `RecomputeTx`, so they have to see those values already cleared. `Run` recomputes the returned changes again. A subject is not part of the Source search document, so delete does not reproject it.
   - **None-only commits.** When every change names a `None` table, `Run` commits with no revision, skips audit, effects, recompute, and listeners, then runs `AfterCommit`. `subject_positions` and `file_derivatives` carry an entity name so they can appear in that list. `Set` and `Clear` return only a `subject_position` change. PR 6 returns the position from subject create and connect and drops it out of a mixed list before audit.
6. **Migrate conclusion-layer writes** (done). Promote single and batch, identity claims, canonical entities, credibility assessments, and property create, update, and delete go through `writes.Run`.
   - **The FFI handler calls `Run`.** Domain functions take `*database.Tx` and return the changes. They do not begin, commit, audit, recompute, or reproject. A no-op credibility assessment and an empty property diff return no changes, so `Run` rolls them back. `SaveBatch` returns `SeenRevision` on the value; the handler replaces it with `Result.Revision` when that is non-zero, so an all-skip batch still reports the revision the proposal was read at.
   - **Handle fixtures.** An accepted mint, a join, filing the other end of a participation, and a batch that files both ends resolve a set that covers the claim's handle, the associations `AppendFiling` or `FileSourceBridgesTx` returned, and `HandlesObservingSubject`. Those cases were not a strict superset, so `observers` stayed. A provisional claim resolves only its own handle. A grade change resolves the same handles as `HandlesForSource`; an argument-only edit resolves none. A cardinality edit resolves `HandlesForProperty`; a label edit resolves none. The registry already matched, including `onField("credibility_grade_id")` and `onField("cardinality")`. An assessment write also refreshes the Source search document, which `Upsert` did not reproject before.
   - **Mixed lists drop `None` rows.** After the empty-list rollback and the all-`None` commit, `Run` drops every `None` row before `audit.Record` and `effects.Resolve`. The rows are already written, so they commit with the revision and are not recorded. Subject create and connect return the `subject_position` change. Thumbnail links stay all-`None`. The derived `files` row stays off the list.
   - **Left in place.** `GetDeleteImpact` still begins a read snapshot. Subject-type property bindings, and property, subject-type, and grade `Upsert` / `Install`, stay direct unaudited writes. Subject delete's hand `RecomputeTx`, and name and date inserts, stay as PR 5 left them.
7. **Close the old path** (done). The pre-registry scope walks, `LegacySourceIDs`, the unused `Recompute*Tx` wrappers, `released.Handles`, and the filename hand reproject are gone. `Record` accepts an entity type when `effects` has a non-`None` entry. A test fails on `.Begin()` outside `core/database` and `core/writes`. `GetDeleteImpact`'s read snapshot lives in `deleteimpact.Snapshot`.
   - **What still catches a missed effect.** The migration assertion is gone, and `Verify` does not exist until PR 8. Rebuild-equals-upkeep stays pointed at the registry-driven handle set. Stored source scopes stay compared with `effects.Sources`. `Verify` later compares a loaded node with the read that filled it, so a short recompute set (stale ARV on both sides) does not fail it.
   - **The literal test waits for PR 8.** A `type Kind string` ban would freeze a check the graph's node shape has not decided. What that test forbids stays with PR 8's node shape: a scan for `TermKey("…")` misses `k == "person"` and SQL literals (`standard` and `moderate` in `sqlLoadCandidates`). Infrastructure holds ids; registries bind a product key to an id.
   - **Left in place.** Subject delete still calls `RecomputeTx` once after the rows are gone (header dependents are PR 9). Seed `Upsert` on source types and metadata fields still reprojects by hand. Thumbnail `AfterCommit` stays. `EnsureCatalog` and `namevalues.Insert` still begin transactions inside `core/database`.

**The graph**

8. **The graph and Promote** (done). `graphcache` hangs on `Catalog` and listens for commits. A read fills a node (identity, every ARV rank, links, members) or a kind's unmerged ids. A commit replaces what is already loaded, inserts a new id into a loaded kind index, and, for an association, loads the new endpoints. `EnsureCatalog` drops the store after a rebuild. Promote walks that graph with no hop cap: top-5 seeds per unfixed kind, links ordered by association ref then neighbor ref. Stats stay a full SQL scan, rerun when `MAX(audit_transactions.revision)` changes, with the property list taken from the match profiles. The scan lives on the catalog's graph, not in a process-wide cache.
   - **No runtime drop.** A read does not query `audit_transactions` and does not discard the store when the log is ahead. The `.Begin()` test remains the check that a new write does not start its own transaction. `Verify` fails when a loaded node disagrees with a fresh reconcile of the observations, and with the kind index when one is loaded. It does not repair the store. `PROVENENCIA_GRAPH_VERIFY=1` makes `OnCommit` return that error. Production does not set it. A listener error still drops the store.
   - **Ids on the node.** A node stores `subject_type_id`, property ids, and term ids. Direction and `inverse_key` are copied off the term row while the node is loaded. `promotealign` builds `graphalign.EdgeSignature` at the edge. The typed-key literal test stays deferred.
   - **Align stays pure.** The loader records a read error and `Propose` returns it without calling Align. Existing promote fixtures matched after the switch. The #327 caps and the 250ms debounce are not ported. The narrowed one-hop exhibits are.
   - **Ranks stay.** Nodes keep every ARV rank. The benchmark reports heap; this PR does not trim them. A separate prefetch is not added.

9. **Display.** The header registry, with header dependents derived from it (replacing `HeaderDependents`); header memos; lists and `…ByIDs` from the graph; the vocabulary map; "today" at read time.
   - *Open:* handles with no members. No code deletes from `canonical_entities`; a handle that loses its last member just has no ARV rows. The kind index's insert and remove rules are PR 8. This PR decides the lists' membership filter on top of that index.
   - *Open:* the invalidation closure. Dependents are taken on the handles whose structure was reloaded: the recompute set, plus the association's old endpoints (`holders`) and new endpoints (committed ARV). Replacing an endpoint node clears that node's memo and does not clear the person whose birth line embeds the event. A person header reads subject-role events, those events' places, and `ParentChain` to eight rounds (`loadLives`, `placechains.go`). The reverse is transitive to that depth, then continues through location and subject-role hops. `HeaderDependents` today is one `PartsOfPlace` hop (`ChildPlaceIDs`) plus `EventsAtPlace` on the changed ids only, and it does not walk association handles. Porting it leaves grandchild chains, events located in a child place, and life-fact lines stale, in search documents and in memos.
   - *Open:* the header registry names hops. A person header is `EventsOfSubject`, then the birth and death terms, then `PlacesOfEvent`, then `ParentsOfPlace` to depth 8. Dependents fall out of reversing that declaration. A second description of the same hops will drift from `canonicalgraph`, and PR 10 would add a third. Chain and succession reads are added to this declaration, not a new walk.
   - *Open:* where the memo lives. PR 8's nodes are immutable, replaced on reload. A `display` field filled on first read and cleared when a neighbor changes is a write to a node a caller may still hold. Keep memos in a side table keyed by id, dropped when the closure says so. `Verify` compares those memos to a fresh build, not only the node's ARV rows.
   - *Note:* this makes Go reads fast and correct, but Swift's own invalidation gaps (deleting an artifact leaves citations and the conclusion lists stale) stay until `Result.Effects` reaches Swift.
10. **Place chains** from named hops on the graph, replacing #328's per-request index. Named hops move to their own registry file.
    - *Open:* chains share the PR 9 declaration. Containment to depth 8 and succession both ways are header reads on that registry. Their dependents are the reverse of those reads, to the same depth. This PR does not add a separate dependent walk; a grandparent rename has to clear every descendant chain that displays it.
11. **Conclusion detail** from node values, with the "Why" outcomes dropped with the display memo.
    - *Open:* cache the per-Observation outcomes per handle, or keep the SQL read; decide after measuring. A memo joins the PR 9 side table and clears under the same closure; `Verify` covers it either way.
12. **The Source store**, fed by resolved Source scopes; Promote's layer, the Evidence graph and the Composer on it.
    - *Open:* which scopes drop a Source. Audit's `noScope` on `identity_claim`, `identity_claim_evidence`, and `canonical_entity` is Source history, and it is the wrong drop signal. A promote reloads the canonical handle and leaves `Members` on the previous handle unless this store drops every Source the claim's subject belongs to. Use the cache Source scopes from PR 2, and keep audit's `noScope` as it is.
    - *Open:* vocabulary-structure drops. A term's `directed` or `inverse_key` change drops the canonical graph, and it drops this store too. Evidence-graph bridge signatures use the same direction and inverse.
    - *Open:* `Verify` for this store. Compare loaded `Members` and bridge signatures with a fresh read. A canonical-node check does not see a stale membership pointer.

Later, and separately: `Result.Effects` rides back to Swift on each write's response, and the client's session cache invalidates by handle and Source instead of the hand-kept `CatalogMutation` map.
