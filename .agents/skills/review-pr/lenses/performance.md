# Performance lens

The question: **will this stay fast on a real project — thousands of Sources,
a census page with dozens of people, a citation with 70 Observations — and
does it keep work off the main thread?** Runs when the diff touches queries,
lists, the Evidence graph, caches, search, derivatives, the catalog session, or
work that runs in a SwiftUI `body` or on the main actor.

Say what grows and how fast. A finding without a size ("N Sources × M
queries") is a guess.

## Read first

- [`docs/macos-client-patterns.md`](../../../../docs/macos-client-patterns.md)
  §1 catalog session and workspace session, §2 cheap `body`.
- [`archive/spike-9/performance-ledger.md`](../../../../docs/deployment-plan/archive/spike-9/performance-ledger.md):
  measured timings and the reserve tiers that were not needed.
- [`use-catalog-session`](../../use-catalog-session/SKILL.md),
  [`add-workspace-place`](../../add-workspace-place/SKILL.md).
- The precedents that fixed real slowness: the combo box builds options only
  while open (#356); the Evidence graph draws cards in AppKit (S10-46…48
  descoped because of it); the grid is one tile (S10-02).

## How to work

1. For each changed path, name the input that grows: rows in a table, cards on
   a graph, observations on a citation, files on a Source, characters in a
   transcription.
2. Count the work per unit of that input: queries, FFI calls, allocations,
   layout passes, decodes.
3. Ask where it runs: the main actor, a SwiftUI `body`, a background task, Go
   under the held catalog session.
4. Compare with the neighbor that already does the same job; it is usually the
   baseline to match.

## Look for

### Go and SQLite

- N+1: a query per row in a loop instead of one grouped query or an `IN` list
  through `sqlbatch`.
- A list query without the index it needs; a full scan on workspace open; an
  unbounded `List` on a table that grows.
- Repeated open or migrate; work done per RPC that belongs to the session.
- Holding the catalog session or a lock across slow I/O (file reads, decodes).
- Rebuilding a derived cache (`graphcache`, `autoreconciler`, `searchindex`)
  in full where an incremental update exists.
- Copying large buffers; decoding a full image where a thumbnail exists.

### FFI and the Mac client

- A synchronous FFI call on the main actor that can take more than a frame
  (a list load, an ingest, a decode).
- An FFI call per row where one call returns the list.
- Work in `body` or in a computed property `body` reads: sorting, filtering,
  formatting dates, building option lists, `session.query()`.
- Observation too wide: a view that observes a whole model and re-renders on
  every field change; missing stable identity in `ForEach`.
- Loading full originals for a grid or a card; recomputing a thumbnail the
  cache should hit.
- A cache bypassed: a destination that loads into local state instead of the
  warmed `QueryHandle`; an invalidate where a patch was enough, so a whole list
  reloads after one edit.
- Search: reprojecting FTS on every keystroke; a query per keystroke without
  the existing debounce.

### Graph and canvas

- Per-card work that scales with every card on the Source (layout, hit
  testing, rotor entries) instead of the cards that changed.
- Drawing that allocates per frame; a bitmap that grows with the document
  instead of tiling.

## Allow

Small, bounded inputs (vocabulary lists, a handful of settings). One-time work
on project open that the ledger already accepts. Speculative micro-optimizations
are not findings.

## Severity in this lens

| Severity | Use when |
| --- | --- |
| **blocker** | Main-thread work that will hang on a real project; N+1 on a list that is already large in dogfood data; a cache bypass that reloads a whole list per edit on a common path |
| **major** | Work that grows faster than the input on a path researchers use daily; a missing index on a new list query |
| **minor** | A measurable cost on a quiet path |
| **nit** | — (performance needs a size, not taste) |

## Good finding shape

The shape to copy. Names in the example are illustrative.

"`SourcesListModel` calls `store.sourceEvidenceStatus(id)` once per row
(N FFI calls, each a query) when the list loads. With 2,000 Sources that is
2,000 round trips on the main actor. The working rule for S10-41 asks for one
grouped query beside `sourcegraphprogress`; return the counts with the list."
