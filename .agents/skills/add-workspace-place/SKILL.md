---
name: add-workspace-place
description: >-
  Registers a Provenencia macOS workspace place (PlaceRegistry + CatalogQueryKey +
  WorkspaceSession cache + destination view). Use when adding or changing a sidebar
  destination, list/detail split, master–detail vocabulary, CatalogQueryKey,
  PlaceRegistry, WorkspaceDestinationHost, warmListQueries, QueryHandle observation,
  or session.apply(location:) — not for navigation history alone (add-workspace-location)
  or Go SQLite queries alone (add-catalog-query).
---

# Add a workspace place (registry + session cache)

A **place** is a restoreable workspace location plus the catalog reads and SwiftUI
presentation it needs. Spike 4 centralizes this in `Features/Workspace/Session/` and
`WorkspaceDestinationHost` — do **not** copy per-view `.task { load() }` /
`reconcileNavigation` from pre-S4 patterns.

Authoritative design: [`docs/ideas/archive/page-navigation-performance.md`](../../../docs/ideas/archive/page-navigation-performance.md).
Spike decisions: [`docs/deployment-plan/archive/spike-4/README.md`](../../../docs/deployment-plan/archive/spike-4/README.md).

## When this applies

| Change | This skill | Also use |
| --- | --- | --- |
| New sidebar destination / list / detail page | **Yes** | [`add-workspace-location`](../add-workspace-location/SKILL.md) |
| New catalog read cached across navigation | **Yes** | descriptor arm in `CatalogQueryRegistry` (loader, `patches` / `refresh`, invalidateOn, whenHidden) |
| Omnibar-searchable entity | Partial (location + queries) | [`add-searchable-kind`](../add-searchable-kind/SKILL.md) |
| History-only deep id on existing place | No | `add-workspace-location` |
| Go table SQL / FFI handler | No | `add-catalog-query`, `add-ffi-handler` |

## Checklist

```
- [ ] WorkspaceLocation: optional deep id (+ == ignores ref/title fluff)
- [ ] WorkspaceSection + sidebar + L10n (if new destination)
- [ ] PlaceID case + PlaceRegistry Spec (matches, queryKeys, presentation, priority)
- [ ] CatalogQueryKey case + descriptor arm (loader, invalidateOn, whenHidden)
- [ ] WorkspacePresentationID + WorkspaceDestinationHost switch arm
- [ ] View: warm queries in .task; read queryHandle in body; @Bindable QueryHandle child
- [ ] Model: warm*Queries() / syncCatalogCounts; mutations via session.apply / setQueryValue
- [ ] go(to:) at row / link / omnibar call sites
- [ ] PlaceRegistryTests (+ session tests for new keys)
- [ ] add-searchable-kind (if omnibar-searchable)
```

## Architecture (three layers)

```text
WorkspaceNavigation.currentLocation     ← where (history)
PlaceRegistry.resolve(location)       ← place id, presentation, query keys
WorkspaceSession QueryHandles         ← what (catalog read cache)
```

Navigation commit (`go`, Back/Forward, restore) calls `session.apply(location:)`,
which warms each key from the resolved place **non-blocking**. Views subscribe to
handles; they do not own fetch-on-appear lifecycle.

### Cache update strategy

| Mechanism | Use when |
| --- | --- |
| **Patch** (`patchCached`, `setQueryValue`, descriptor `patches`) | Overlay after a write that returned enough data. A write that reports its activity patches a list's last-updated cell through a `CatalogQueryPatch.lastUpdated` on that list's descriptor (`patches`, with its own `follows`); no refetch. An enriched Source uses `.updatedSourceRow` and `.updatedSourceShell` on the same slots. `patchCached` and `setQueryValue` stay the primitives those slots call. For a change the write does not describe (a value, title, or membership), pair the patch with `apply` so the descriptor still refetches. |
| **Refresh** (descriptor `refreshes`, a list) | A key follows the ids a write reports, one entry per slot (`CatalogQueryFollows`: `.reads(…)`, `.touchedHandles`, `.evidenceGraphs`). An entry either reloads some rows and merges them into a warmed value (`CatalogQueryRefresh(follows:run:)`: evidence-graph progress, a Get-one and map merge skipped when the map was never warmed) or revalidates the key. A key that reloads when a write names an id declares `applyHiddenPolicy(follows:ids:)` there, or `applyHiddenPolicyOnAny(follows:)` when any id will do, never a block in `invalidations(for:project:)`, which keeps only `lists` and `structure`. `ids` returns what the key depends on, read from the key (the citation kinds' graph, a detail page's entity id) or from its cached value (person claims: the graphs they hold a claim on). `WorkspaceSession.revalidatedKeys(for:_:)` names the warmed keys a write revalidates, and `apply` revalidates each once alongside the registry's invalidations. |
| **Invalidate + revalidate** (`session.apply`) | Default after a catalog write. `apply` marks keys stale and reloads any that are already warmed. Features do **not** `session.query` after `apply`. A write returns `WriteEffects` and the app calls `apply(effects)`; when a mutation still names keys the effects cannot (the Source page's own workspace, the metadata field list), call `apply(effects, mutation)`, which applies both once and never invalidates a key twice. |
| **Mark stale** (`whenHidden: .markStale`) | A write moves a value the current screen does not show, and refetching on every write is too costly. `apply` flags the warmed key without a refetch; the next `query` / place visit reloads it, showing the cached frame meanwhile. `whenHidden` (`.revalidate`, `.markStale`, `.evict`) sets what an invalidated key does when its place is not on screen. |
| **Stale-while-revalidate** | Navigation reads — Back, sidebar return; show cached frame, `isFetching` while reloading. Not the only refresh path after a write. |

Do **not** call `session.query()` from view `body` or from model computed properties
that `body` reads every frame — that starves the MainActor and breaks observation.

Do **not** fold a shared list into a page payload (`evidenceGraph` is subjects +
positions + observations; types live on `propertiesWorkspace`). Do **not**
copy catalog lists onto a feature model — observe the handle.

## Steps

### 1. History + location ([`add-workspace-location`](../add-workspace-location/SKILL.md))

Every committed place still flows through `navigation.go(to:)`. This skill assumes
history wiring exists or is added in parallel.

### 2. `PlaceID` + `PlaceRegistry` spec

In [`PlaceRegistry.swift`](../../../macos/App/Features/Workspace/Session/PlaceRegistry.swift):

- Add `PlaceID` case (if new place kind).
- Add `Spec` with **priority** (more specific places beat section roots).
- `matches`: `WorkspaceLocation` predicate (ignore cross-section stray ids).
- `queryKeys`: `[CatalogQueryKey]` for this place (include shared pool keys, e.g. types list on source list).
- `presentation`: `WorkspacePresentationID` for the host router.
- `viewIdentity`: the id the place's view resets on. `WorkspaceDestinationHost` applies `.id(place.viewIdentity)` once for every place; a different value remounts the destination and its `@State` model. Declare the deep id a view is built from (graph, Source, entity, composer or Promote entry key); declare `nil` for list roots and for master–detail places that follow `currentLocation` themselves. Do not add `.id` in a host arm and do not rebind a model in place on an id change.

Extend [`PlaceRegistryTests`](../../../macos/ProvenenciaTests/PlaceRegistryTests.swift):
`registryCoversAllPlaceIDs`, query key table row for the new place.

### 3. `CatalogQueryKey` + descriptor arm

In `Session/`:

- Add `CatalogQueryKey` case (project-scoped payload as needed) and a `Kind` case.
- Add one arm on `CatalogQueryKey.Kind.descriptor` in [`CatalogQueryRegistry.swift`](../../../macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift): the loader (`guard case` on the key, calling `GenealogyStore`), `invalidateOn`, `whenHidden`, and `invalidation`. The switch is exhaustive, so a new kind does not compile until that arm exists.

**One cache owns each list — do not fold a shared list into a page payload.**
A deep page composes its own key plus the shared list keys it reads (see the
`sourceDetail` spec), so a vocabulary edit invalidates one list and no pages. The
held `catalogsession` makes the extra RPCs cheap, and the lists are usually already
warm from the sidebar.

**Then tag every mutation whose reply changes any *derived* part of a payload, not
just the part the write names.** A `CatalogTypeSuggestion` carries a whole field
row; `workspace.metadata` is a Go-computed join over the type's suggestions; counts
like `usedBy` / `suggestedFieldCount` move when a different table is written. If the
mutation names no single owner for an id-bearing key, the descriptor's `invalidation` returns
`.allCached(kind)` and every cached key of that kind is busted.

Add [`CatalogQueryRegistryTests`](../../../macos/ProvenenciaTests/CatalogQueryRegistryTests.swift) coverage for load + invalidation (`registry.invalidations`) and the hidden policy (`registry.hiddenPolicy`).

### 4. Presentation + host

- Add `WorkspacePresentationID` case if the host needs a distinct view family.
- Add arm in [`WorkspaceDestinationHost`](../../../macos/App/Features/Workspace/WorkspaceDestinationHost.swift). The host applies identity from the spec; an arm adds no `.id`.
- List vs detail = **two presentations**, two views (Sources pattern). Master–detail
  vocabulary = one presentation, selection from `currentLocation`'s row id (Fields/Types); their `viewIdentity` is `nil`.

### 5. View + model (observation pattern)

**Shell view** (example: list destination):

```swift
@Environment(WorkspaceSession.self) private var session

var body: some View {
    Group {
        if let handle: QueryHandle<[Row]> = session.queryHandle(listKey) {
            ListContent(handle: handle, model: model)
        } else {
            ProgressView()
        }
    }
    .task { model.warmListQueries() }  // idempotent; navigation may have warmed already
}

private struct ListContent: View {
    @Bindable var handle: QueryHandle<[Row]>
    @Bindable var model: MyModel
    // read handle.status / handle.value so Observation repaints when load completes
}
```

**Model rules:**

- `warmListQueries()` calls `session.query(key)` — **only** from `.task`, actions, or tests.
- Display arrays read `session.queryHandle(key)?.value ?? []`.
- CRUD: `session.apply(CatalogMutation…)` only. Optional `setQueryValue` overlay; never `query` after `apply`.
- Sidebar badges: the list view publishes with `.onSettled(handle) { model.syncCatalogCounts() }`, which runs on appear and on each `revision` bump while `handle.isSettled` (`.ready`, not refetching). Never on `onChange(of: status)`: a stale refetch stays `.ready`, so its reload is `.ready → .ready`.
- A mutation that changes a sidebar count is listed in `CatalogQueryRegistry.navCountTriggers`; `WorkspaceSession.apply` recounts after it. Features never call `CatalogCounts.refreshAll()`. Publish from a list only on load, never after its own write: the list may be hidden or still reloading when a write returns.

**Master–detail selection** (fields/types): `reconcileSelection(for:)` on
`onChange(of: navigation.currentLocation)` + `onChange(of: handle.status/value)` —
sync apply from history after list is ready; `fallbackToSectionRoot()` if deep id missing.

**Source-style detail:** `SourcePageView` pattern — the host remounts it per Source through `viewIdentity`; a plain `.task` warms the workspace key;
`@Bindable workspaceHandle` syncs into section models; fingerprint skips redundant section resets.

### 6. Mutations from section models

- Identity/cover/title: `session.apply(.updatedSource(source))` (patches list + workspace).
- Source type moved: `context.applySource(updated, typeChanged: true)` — patches the row, then refetches the page because the engine derives suggested rows from the type.
- Metadata values: `context.applyMetadataWrite(effects)` with the write's effects (also busts the field list's `usedBy`).
- Other page edits (notes, artifacts, credibility): `context.applyWrite(effects)` → bust the `sourceWorkspace` key, and the effects move the lists.
- Graph and composer writes (`createSubject`, `updateSubject`, `deleteSubject`) and Artifact ↔ graph link writes (`updateArtifact` with `graphIDs`, `updateEvidenceGraph`): `session.apply(effects)`, plus `.deletedSubject` for a delete (it reloads Observation `usedBy`; the citation reads follow the effects).
- Any write under a Source moves its `updated` activity (engine audit scopes). Every write under a Source returns `WriteEffects` with its activity, and `apply(effects)` patches the Sources row in place; the mutations that ride along (`mutatedSourceWorkspace`, `mutatedSourceMetadata`, `deletedSubject`) never invalidate `sourcesList` or the Conclusion lists — the Sources row follows the effects' `sourceScopes`, each graph's rows follow `evidenceGraphs`, the keyed reads (citation counts and citations by Artifact, graph progress) follow `reads` for only the graphs named, and the Conclusion lists follow `handles`. A new Source child write returns `WriteEffects` (`writeEffectsProto(wrote)` in its handler); the writes that still report none (`updateSource`, `setSourceCover`, `deleteSource`, promote, Property writes) keep their mutation kinds. `createSource` returns `WriteEffects` and is applied with `apply(effects, .createdSource)`, as Copy Source is.
- A new list with a last-updated cell: conform its row to `CatalogLastUpdatedRow` and add `.lastUpdated(Row.self, follows: …)` to its descriptor `patches`.
- Vocabulary CRUD: registry `invalidateOn` + list patch where cheap. A patch is an
  optimization on top of invalidation unless the write reported everything the patch needs (last-updated activity).

### 7. Tests

| Layer | Tests |
| --- | --- |
| Registry | `PlaceRegistryTests`, `CatalogQueryRegistryTests` |
| Session | `WorkspaceSessionTests` (apply location warms keys, cache hit) |
| Host | `WorkspaceDestinationHostTests` (presentation routing) |
| Feature model | FakeStore + `warmListQueries()` / `warmFromSession()` helpers |

[`add-swift-test`](../add-swift-test/SKILL.md) for wiring new Swift test files.

## Do not

- Gate list vs detail on feature-model `openedSourceID` — use `PlaceRegistry` + host
- Put `session.query()` in `body` or computed properties hit every render
- Use empty `onChange` hacks instead of `@Bindable QueryHandle` observation
- Invalidate list + workspace on every title keystroke — patch synchronously
- Skip `PlaceRegistryTests.registryCoversAllPlaceIDs` when adding `PlaceID` cases

## Related

- Navigation history: [`add-workspace-location`](../add-workspace-location/SKILL.md)
- Catalog session / FFI: [`use-catalog-session`](../use-catalog-session/SKILL.md)
- Search hits → location: [`add-searchable-kind`](../add-searchable-kind/SKILL.md)
- macOS UI patterns: [`docs/macos-client-patterns.md`](../../../docs/macos-client-patterns.md) § workspace session

## Debug

- `PROVENENCIA_DEBUG_SESSION_APPLY=1` — log place + query keys on `session.apply(location:)`
- `PROVENENCIA_DEBUG_NAV_TIMING=1` — log timestamp on each `WorkspaceNavigation` commit (DEBUG builds)
