# Performance lens

The question: **will this stay fast on a real project — thousands of Sources,
a census page with dozens of people, a citation with 70 Observations — and
does it keep work off the main thread?** Runs when the diff touches queries,
lists, the Evidence graph, caches, search, derivatives, the catalog session, or
work that runs in a SwiftUI `body` or on the main actor.

Say what grows and how fast. A finding without a size ("N Sources × M
queries") is a guess. Two things always grow: the catalog (Sources,
Subjects, handles, Observations) and its history (audit transactions and
scope rows, revisions). Every project only gets bigger, so a cost that
scales with either one on a path researchers use daily is a defect now,
not when someone notices. A layout pass that invalidates itself is the exception:
the input can be one control, and the work does not stop.

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
   baseline to match. If the neighbor has the same growth problem, the
   neighbor is not a baseline: report both. Do not lower a severity because
   the change follows, copies, or widens an existing pattern.
5. For every write path, ask what the client does after it: patch the rows
   the write reported, or reload. For every read path, ask whether it reads
   rows it returns or scans everything to answer.

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
- A read that aggregates history (audit transactions, scope rows, revisions)
  on every list load or navigation, where a cache patched on commit would
  answer from memory. One load per session is the most it should cost.
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
- A write that reports what it changed, followed by a reload instead of a
  patch. The write's own result (ids, revision, time, new value) is the
  cheapest update there is.
- Reloads of data that is not on screen. An invalidation should mark a
  hidden value stale and let the next visit reload it, not rebuild it now.
- Search: reprojecting FTS on every keystroke; a query per keystroke without
  the existing debounce.

### SwiftUI layout

Layout is a fixed point, and SwiftUI runs it on the main thread until the
sizes agree. The parent proposes a size, the child reports one, and if that
report changes the next proposal, layout runs again in the same turn. Nothing
in the code looks like a loop. There is no `onChange`, timer, or growing
collection. When the sizes never agree, pointer events do not get a turn, so
pan, drag, and unrelated updates all hitch.

- A choice that changes the space it is offered. `ViewThatFits`, or a custom
  `Layout` that picks among children, inside a stack, spacer, or
  `layoutPriority` that hands the chosen child the leftover width. The wide
  form does not fit, so the narrow form is chosen; the narrow form leaves
  room, so the wide form is offered again. `.fixedSize()` on the wide form
  makes the disagreement permanent: the child reports its ideal width and
  ignores the proposal.
- A measurement taken under a different proposal than the one used to place
  the view, then written back as the view's size. `Text`, and any view that
  wraps, measured at width 0 reports the height of a full wrap. Using that
  height as the placed height changes what the parent proposes next. An
  unbounded ideal width does the same when the leftover width depends on the
  width just reported.

The choice has to come from a width the choice does not change: the
container's own width, or an ideal width measured where it cannot affect the
proposal.

### Graph and canvas

- Per-card work that scales with every card on the Source (layout, hit
  testing, rotor entries) instead of the cards that changed.
- Drawing that allocates per frame; a bitmap that grows with the document
  instead of tiling.

## Allow

Small, bounded inputs (vocabulary lists, a handful of settings). One-time work
on project open that the ledger already accepts. Speculative micro-optimizations
are not findings. A layout cycle is not a bounded input: it does not finish.

## Severity in this lens

| Severity | Use when |
| --- | --- |
| **blocker** | Main-thread work that will hang on a real project; a layout cycle that does not settle, even on one control; N+1 on a list that is already large in dogfood data; a cache bypass that reloads a whole list per edit on a common path; any per-edit, per-load, or per-navigation cost on a common path that grows with the catalog or its history, including one this change only adds a caller to or widens |
| **major** | Work that grows faster than the input on a path researchers use daily; a missing index on a new list query |
| **minor** | A measurable cost on a quiet path that does not grow with the catalog or its history |
| **nit** | — (performance needs a size, not taste) |

## Known traps

Classes that reached `main` past this lens and were found later, by an
app-health audit or a landing. Each line names the class, the shape it took,
and what to check. Add a line when an audit traces an item back to this
lens ([`review-pr`](../SKILL.md) "Close the loop").

- **Rate at the spike's scale.** Use the map's **Scale** line (60 household
  graphs on one Source, 300 Citations on one page), not a one-graph project.
  Fan-outs and per-child RPC loops look free at one.
- **A write that names every sibling.** An effect that names every graph of a
  Source when the write moved one reloads every graph per save, on the case
  the spike was built for. Name what the write's rows load.
- **A project-wide list reloaded for one parent.** The write knows which
  Source moved; re-read that one.
- **A key with no hidden policy.** A `CatalogQueryKey` without `whenHidden`
  reloads off screen after every write that names it.
- **A history table without its lookup index.** Two correlated subqueries on
  `audit_changes` by `entity_id` scanned all history per note on every Source
  page load.
- **File work under the session.** Copying, hashing or decoding inside
  `catalogsession.Do` stalls every pane until it ends.
- **Work in `body` over the whole list.** A filter and sort of every row runs
  on each keystroke of any field the same view reads.

## Good finding shape

The shape to copy. Names in the example are illustrative.

"`SourcesListModel` calls `store.sourceEvidenceStatus(id)` once per row
(N FFI calls, each a query) when the list loads. With 2,000 Sources that is
2,000 round trips on the main actor. The working rule for S10-41 asks for one
grouped query beside `graphprogress`; return the counts with the list."
