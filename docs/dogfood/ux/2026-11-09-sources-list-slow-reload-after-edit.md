# Sources list slow when returning after editing a Source

- **Date:** 2026-11-09
- **Type:** Friction (performance)
- **Where:** Sources list — especially navigating back from a Source page after edits
- **Annoyance (verbatim):** "Loading the sources list is a little slow. I'm going through a process right now where I edit a source and then go back to the list, and the load time is noticeably non-performant. If we could figure out a way to improve that, that would be great."
- **Consequence:** A tight edit-and-return loop on Sources feels sluggish; researchers wait on the list instead of scanning the next row.
- **Wanted:** Returning to the Sources list after a Source edit should feel immediate (stale-while-revalidate at most), with any unavoidable server work kept off the hot path.

## How it works today (code notes)

### macOS: what loads when the list is shown

- **Place warming.** `PlaceRegistry` (`macos/App/Features/Workspace/Session/PlaceRegistry.swift`) registers the Sources list place with four catalog query keys: `sourcesList`, `sourceTypesList`, `evidenceGraphsList`, and `evidenceGraphProgress`. Every `WorkspaceSession.apply(location:)` for that place calls `warm` on each key (`WorkspaceSession.swift`).
- **List view lifecycle.** `WorkspaceDestinationHost` swaps between `SourcesListView` and `SourcePageView` by presentation id (`macos/App/Features/Workspace/WorkspaceDestinationHost.swift`); the list view is torn down on drill-in and constructed again on back. Session-level query caches survive, but `SourcesListView` holds a fresh `@State` `SourcesModel` and re-runs `.task { model.warmListQueries() }` (`SourcesListView.swift`).
- **Empty vs stale-while-revalidate.** The list shows a full-screen `ProgressView` only when `sourcesHandle.status == .loading` **and** `model.sources.isEmpty` (`SourcesListView.swift`). If the session still holds a prior list value, the UI can paint rows while a stale refetch runs in the background (`QueryHandle.beginLoading(keepingValue:)` in `WorkspaceSession.ensureQuery`).
- **Hidden-list policy.** `sourcesList` uses `whenHidden: .markStale` (`CatalogQueryRegistry.swift`). While the Source page is open, the list key is not in `visibleKeys`; registry invalidations (today: `deletedSource`) mark it stale instead of reloading immediately. Revisiting the list triggers `revalidate` → another `listSources` FFI when stale (`WorkspaceSession.applyHiddenPolicy`, test `staleKeyIsNotSettled` in `WorkspaceSessionTests.swift`).
- **Writes from the Source page.** Identity/cover saves call `SourcePageContext.applySource` → `session.apply(..., .updatedSource(source))` or `.changedSourceType(source)`, which patch the cached list row via `CatalogQueryPatch.sourceRow()` and `lastUpdated` without registering a full `sourcesList` invalidation (`CatalogLastUpdated.swift`, `CatalogMutation.patchedSource`). Notes, credibility, metadata, and artifacts use `.mutatedSourceWorkspace` / `.mutatedSourceMetadata`, which invalidate the page `sourceWorkspace` key but not `sourcesList` (test `sourceWriteEffectsPatchSourcesAndLeaveListsAlone` in `WorkspaceSessionTests.swift`). After typical edits, the list **should** often still be warm and not stale — any slowness may be client render work or refetches of sibling keys, not always a full `ListSources`.
- **Per-row UI work (not per-row FFI).** Each `SourcesListRow` (`SourcesListRow.swift`) reads `session.evidenceGraphsRead(sourceID:)` and progress from the project-wide maps warmed by `evidenceGraphsList` / `evidenceGraphProgress` — not one RPC per row. Rows still build one `SourcesGraphButton` per graph, load `CachedThumbnail` from `thumbnailRelPath`, and run filter/sort in `SourcesModel.visibleSources` (client-side sort over the full array, memoized by handle revisions).
- **Source page also warms graph keys.** The Source page place loads the same `evidenceGraphsList` and `evidenceGraphProgress` keys (`PlaceRegistry` `sourceDetail` spec), so those caches are usually already warm when returning — unless a write marked `evidenceGraphsList` stale while hidden (`whenHidden: .markStale` on that key too).

### Go / FFI: what `ListSources` costs

- **Handler pipeline.** `ListSources` (`api/ffi/handlers/sources.go`) opens the catalog session, calls `sources.List`, batches cover thumbnails via `sourceCoverThumbnails` (`api/ffi/handlers/thumbnails.go`), and emits one `engine.Source` proto per row through `sourceProto`.
- **SQL list projection.** `sources.sqlList` (`core/database/sources/sources.go`) selects every Source with a correlated `EXISTS` on `artifacts` for `has_artifact`, then `List` batches `datevalues.LookupManyTx` for dated rows and fills `Updated` from `c.Graph().Activity(effects.ScopeSource, id)` for **each** row. Activity map lookup is cheap after the first read of `ScopeSource` (one grouped audit query in `graphcache.loadActivity`, `core/database/graphcache/activity.go`) — not N separate audit scans, but still O(sources) work per `ListSources` call.
- **Wire size.** `engine.Source` includes full `description` strings for every list row (`api/proto/engine.proto`); protobuf marshal/unmarshal scales with source count and description length. Thumbnail paths are attached in the handler, not stored on the row.
- **Sibling RPCs for the same screen.** The list place also loads `ListEvidenceGraphs` with empty `source_id` (all graphs in the project, each carrying `artifact_ids`, `api/ffi/handlers/evidence_graphs.go`) and `ListEvidenceGraphProgress` (`graphprogress.List`, two `GROUP BY` aggregates over `subjects` and `observations`, `core/database/graphprogress/graphprogress.go`). All three cross the FFI boundary when any of those keys are cold or stale.

### What we cannot confirm without profiling

- Relative time spent in SQLite vs protobuf vs SwiftUI layout vs thumbnail decode on Jake's project size.
- Whether back-navigation slowness is dominated by a stale `listSources`, by `evidenceGraphsList` / `evidenceGraphProgress`, or by row re-render after view re-creation.
- Index impact on `sqlList`'s `EXISTS (SELECT 1 FROM artifacts …)` and on `graphprogress` aggregates (no dedicated investigation of query plans here).

## Fix directions (options, not chosen)

1. **Keep the list model across navigation** — host `SourcesListView` state (or `SourcesModel`) above `WorkspaceDestinationHost` / use a stable view identity so back does not reconstruct the list shell and re-run warm queries from a cold `@State`.
2. **Refetch only the edited row** — extend write patches so returning from a Source page never marks `sourcesList` stale; rely on `sourceRow()` + `lastUpdated` (already partly true) and add richer patched fields when edits change sort keys (date, title, cover, `has_artifact`).
3. **Cheaper list projection** — list-specific proto or DB projection omitting long `description`, optional pagination/virtualization for large catalogs, or moving `has_artifact` to a maintained column/index to avoid per-row `EXISTS` in `sqlList`.
4. **Defer non-critical columns** — load graph buttons / progress lazily per row or per viewport; split "title + ref + updated" from graph zone data so the first paint does not wait on `ListEvidenceGraphProgress`.
5. **Cache policy tuning** — `whenHidden: .revalidate` vs `markStale` for `sourcesList` (trade freshness vs back-navigation latency); or adopt `CatalogQueryStalePolicy.refetchOnRevisit` only for keys where partial patches are insufficient.
6. **Indexes / SQL** — verify plans for `artifacts.source_id`, graph progress aggregates, and audit scope lookups under realistic row counts.
