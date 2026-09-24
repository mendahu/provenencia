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
| New catalog read cached across navigation | **Yes** | registry loader in `CatalogQueryRegistry` |
| Omnibar-searchable entity | Partial (location + queries) | [`add-searchable-kind`](../add-searchable-kind/SKILL.md) |
| History-only deep id on existing place | No | `add-workspace-location` |
| Go table SQL / FFI handler | No | `add-catalog-query`, `add-ffi-handler` |

## Checklist

```
- [ ] WorkspaceLocation: optional deep id (+ == ignores ref/title fluff)
- [ ] WorkspaceSection + sidebar + L10n (if new destination)
- [ ] PlaceID case + PlaceRegistry Spec (matches, queryKeys, presentation, priority)
- [ ] CatalogQueryKey case + CatalogQueryRegistry loader + invalidateOn tags
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
| **Patch** (`setQueryValue`, `.updatedSource`) | Save returns enough data to update list row + detail workspace synchronously |
| **Invalidate** (`session.apply(.created…)`) | Create, delete, or patch is impractical |
| **Stale-while-revalidate** | Navigation reads only — Back, sidebar return; show cached frame, `isFetching` while reloading |

Do **not** call `session.query()` from view `body` or from model computed properties
that `body` reads every frame — that starves the MainActor and breaks observation.

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
- `deepId`: optional id string for history restore / selection.

Extend [`PlaceRegistryTests`](../../../macos/ProvenenciaTests/PlaceRegistryTests.swift):
`registryCoversAllPlaceIDs`, query key table row for the new place.

### 3. `CatalogQueryKey` + registry loader

In `Session/`:

- Add `CatalogQueryKey` case (project-scoped payload as needed).
- Register in [`CatalogQueryRegistry`](../../../macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift): loader calling `GenealogyStore`, `invalidateOn` mutation tags.
- Wire `WorkspaceSession.warmQuery` switch arm if the key kind is new.

**One cache owns each list — do not fold a shared list into a page payload.**
A deep page composes its own key plus the shared list keys it reads (see the
`sourceDetail` spec), so a vocabulary edit invalidates one list and no pages. The
held `catalogsession` makes the extra RPCs cheap, and the lists are usually already
warm from the sidebar.

**Then tag every mutation whose reply changes any *derived* part of a payload, not
just the part the write names.** A `CatalogTypeSuggestion` carries a whole field
row; `workspace.metadata` is a Go-computed join over the type's suggestions; counts
like `usedBy` / `suggestedFieldCount` move when a different table is written. If the
mutation names no single owner for an id-bearing key, `Kind.invalidation` returns
`.allCached(kind)` and every cached key of that kind is busted.

Add [`CatalogQueryRegistryTests`](../../../macos/ProvenenciaTests/CatalogQueryRegistryTests.swift) coverage for load + invalidation.

### 4. Presentation + host

- Add `WorkspacePresentationID` case if the host needs a distinct view family.
- Add arm in [`WorkspaceDestinationHost`](../../../macos/App/Features/Workspace/WorkspaceDestinationHost.swift).
- List vs detail = **two presentations**, two views (Sources pattern). Master–detail
  vocabulary = one presentation, selection from `currentLocation` deep id (Fields/Types).

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
- CRUD: `session.apply(CatalogMutation…)` + `setQueryValue` patch or rely on registry invalidation.
- `syncCatalogCounts()` when list handle reaches `.ready` (sidebar badges).

**Master–detail selection** (fields/types): `reconcileSelection(for:)` on
`onChange(of: navigation.currentLocation)` + `onChange(of: handle.status/value)` —
sync apply from history after list is ready; `fallbackToSectionRoot()` if deep id missing.

**Source-style detail:** `SourcePageView` pattern — `.task(id: sourceID)` warms workspace key;
`@Bindable workspaceHandle` syncs into section models; fingerprint skips redundant section resets.

### 6. Mutations from section models

- Identity/cover/title: `session.apply(.updatedSource(source))` (patches list + workspace).
- Source type moved: `context.applySource(updated, typeChanged: true)` — patches the row, then refetches the page because the engine derives suggested rows from the type.
- Metadata values: `context.notifyMetadataMutated()` (also busts the field list's `usedBy`).
- Other page edits: `context.notifyWorkspaceMutated()` → bust `sourceWorkspace` key.
- Vocabulary CRUD: registry `invalidateOn` + list patch where cheap. A patch is an
  optimization on top of invalidation, never a substitute for it.

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
