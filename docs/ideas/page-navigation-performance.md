# Page navigation performance

**Status:** **implemented** (Spike 4 archived, S4-01…S4-09). Finished steps: [`docs/deployment-plan/archive/spike-4/completed.md`](../deployment-plan/archive/spike-4/completed.md). S4-10+ Go RPC tiering descoped.

**Agent skill:** [`.cursor/skills/add-workspace-place/SKILL.md`](../../.cursor/skills/add-workspace-place/SKILL.md).

---

## Landed (Spike 4)

- `WorkspaceSession` + `CatalogQueryRegistry` + `QueryHandle` — project-scoped catalog read cache with patch, invalidate, and stale-while-revalidate.
- `PlaceRegistry` + `WorkspaceDestinationHost` — location-driven presentation; Sources list and detail are separate views.
- Sources, Source fields, and Source types read query handles; navigation `apply(location:)` warms keys on every commit.
- macOS patterns: [`macos-client-patterns.md`](../macos-client-patterns.md) § workspace session.

The sections below are the **pre-implementation review** (Sep 2025) kept for rationale and extension guidance.

---

## Problem statement (pre-Spike 4)

Navigation history infrastructure was in place (`WorkspaceNavigation`, persisted stack, `go(to:)` everywhere). The remaining problem was **how destinations load and render** when a place changes. Performance was inconsistent across link clicks, Back/Forward, sidebar switches, and omnibar jumps — and the per-view load pattern would not scale as we add Files, Interpretation, cross-links, and heavier detail pages.

This note captured the target **platform infrastructure** — not band-aids, and not a Sources-only refactor.

---

## Known symptoms

- **Sources Back/Forward → detail:** Brief flash of the source **list** before `SourcePageView` appears. History updates `navigation.currentLocation` immediately, but `SourcesView` gates on `model.openedSourceID`, which is set in `onChange` inside a `Task` — one run-loop late.
- **Vocabulary (fields/types):** Master–detail can flash empty detail before selection applies on restore (partially addressed via `load(from:)`; may still flash while catalog loads or on in-place navigation).
- **Section remounts:** Switching sidebar destinations remounts destination views and reruns `.task { load }` — a full catalog round-trip even when you were on that section thirty seconds ago.
- **Source page cold every time:** Opening (or returning to) a source remounts `SourcePageView` (`.id(opened)`) and always calls `getSourceWorkspace` — a heavy multi-query RPC — even when that source was open in the current session.
- **Unpredictable latency:** Two navigations that “feel the same” can differ by hundreds of ms depending on what else is queued on the catalog session (badge refresh, list load, suggestions, thumbnails, search).

---

## Current architecture (as built)

```text
WorkspaceView
  ├─ WorkspaceNavigation (history + currentLocation)     ← source of truth for *place*
  ├─ CatalogCounts (sidebar badges)
  └─ WorkspaceContent
       switch selectedSection {                         ← destroys inactive sections
         SourcesView      @State SourcesModel           ← list AND detail behind if/else
         SourceFieldsView  @State SourceFieldsModel
         SourceTypesView   @State SourceTypesModel
       }
```

Each destination follows the same template:

1. `.task { await reconcileNavigation(load: true) }` — full FFI fetch on first mount.
2. `.onChange(of: navigation.currentLocation) { Task { reconcileNavigation(load: false) } }` — apply deep state asynchronously.
3. Feature model holds **both** catalog rows (`sources`, `fields`, …) **and** presentation state (`openedSourceID`, `mode`, selection) that mirrors `WorkspaceLocation`.

Sources adds a second layer: list vs `SourcePageView` is gated on `model.openedSourceID`, not on `navigation.currentLocation.sourceId`. The page itself owns another `@State SourcePageModel` with its own `.task { load() }` → `getSourceWorkspace`.

Navigation history is first-class and correct. **Data loading and view mounting are not.**

### Sources list vs detail — not a nested hierarchy

At the workspace level, Sources is **one sidebar section**. List and detail are **not** two top-level `WorkspaceContent` siblings today — they are two distinct view types swapped inside `SourcesView` via `if/else`. They share almost no data (list row sync on edit is a small event, not a shared tree). Detail is a **deep place** (`WorkspaceLocation.sourceId`), not a sub-view of the list.

**Target:** two **registered places**, same section, routed by `currentLocation` — not a fat parent wrapping both.

---

## Root causes

### 1. Dual state — place vs presentation

`WorkspaceLocation` commits synchronously in `go(to:)` / Back / Forward. Destination UI reads from feature-model fields that update **one frame later** via `Task { reconcileNavigation }`. Any gate like `if let opened = model.openedSourceID` will render the wrong subtree briefly.

### 2. Section switch = view destruction

`WorkspaceContent` uses `switch section` with no retention. Leaving Sources and returning tears down models and triggers cold `load(from:)` again.

### 3. No project-scoped read cache

Each destination refetches from FFI on mount. No shared session cache for lists, workspaces, or suggestions. `CatalogCounts` deduplicates badge totals only.

### 4. Load waterfalls

List click → frame delay → `SourcePageView` mount → `getSourceWorkspace`. Type restore → `applySelection` → `loadSuggestions`. Heavy RPCs repeat on every remount.

### 5. Catalog session serialization

Concurrent Swift `async` store calls queue on one Go session. Same gesture, different queue depth → inconsistent latency.

### 6. Incidental UI work

`SourcePageView.id(opened)` rebuilds the page when switching sources. List remount reloads thumbnails. `listSources` enriches every row server-side.

---

## Why different navigation methods feel different

| Method | Section view mounted? | Reconcile path | Typical extra work |
| --- | --- | --- | --- |
| List / table click (same section) | Yes | `onChange` → `Task` → `apply` | Sources: + page cold load; Types: + `loadSuggestions` |
| Back / Forward (same section) | Usually yes | Same async `apply` | Sources: list flash + page remount; Types: suggestions RPC |
| Back / Forward (cross-section) | **Remount** target | `.task` → `load(from:)` | Full list fetch + deep reconcile |
| Sidebar switch | **Remount** | `.task` → `load(from:)` | Full list fetch; prior section state discarded |
| Omnibar (same / cross-section) | Same as click / remount | Same | Same |
| Relaunch / restore | Cold | `.task` → `load(from:)` | Everything cold + history restore |

**Fast** when data is still in memory and reconcile is cheap. **Slow** when views remount, pages cold-load, or the catalog queue is busy.

---

## What will not scale

- Per-destination `@State` models tied to view lifetime.
- “Load on appear” as the only hydration strategy.
- Presentation state (`openedSourceID`, `mode`) duplicated alongside `WorkspaceLocation`.
- One monolithic `getSourceWorkspace` per visit with no session cache.
- Copy-paste `reconcileNavigation` in every feature.
- Hard-coded `WorkspaceSession.sourcesStore`-style APIs that require core changes for each new page.

---

## Target architecture — three layers

Build **generic infrastructure first**, migrate Sources as proof, then vocabulary, then future destinations (Files, Interpretation, …). Same declarative spirit as [`core/search/registry.go`](../../core/search/registry.go): search registry says how to **find** a place; a **place registry** says how to **load and show** it.

```text
┌─────────────────────────────────────────────────────────────┐
│  WorkspaceNavigation (generic — already exists)             │
│  Stack of WorkspaceLocation; go(to:) / Back / Forward       │
└──────────────────────────┬──────────────────────────────────┘
                           │ currentLocation
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  PlaceRegistry (declarative)                                │
│  location → ResolvedPlace → queries + view + invalidation   │
└──────────────────────────┬──────────────────────────────────┘
                           │ CatalogQueryKey[]
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  WorkspaceSession + CatalogQueryRegistry                    │
│  Query cache (React Query–like): load, dedupe, stale, bust   │
└──────────────────────────┬──────────────────────────────────┘
                           │ QueryHandle<Value>
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Views (thin subscribers)                                   │
│  WorkspaceDestinationHost picks view from registry            │
└─────────────────────────────────────────────────────────────┘
```

### Layer 1 — Navigation (unchanged)

`WorkspaceNavigation` + persisted stack. Does not know about Sources vs Files. Only requires that `WorkspaceLocation` encodes enough to restore a place (per-kind optional deep ids in JSON — see navigation-history spike).

### Layer 2 — Catalog query cache (the reusable engine)

TanStack Query / React Query analogue. **Not Sources-specific.**

```swift
enum CatalogQueryKey: Hashable, Sendable {
    case sourcesList(project: ProjectKey)
    case sourceTypesList(project: ProjectKey)
    case metadataFieldsList(project: ProjectKey)
    case sourceWorkspace(project: ProjectKey, sourceId: String)
    case typeSuggestions(project: ProjectKey, typeId: String)
    // future: filesList, fileDetail, nodeWorkspace, …
}
```

Each key is registered once in **`CatalogQueryRegistry`** with:

| Field | Purpose |
| --- | --- |
| `loader` | `(store, key) async throws -> Value` |
| `stalePolicy` | session-fresh, refetch-on-revisit, etc. |
| `invalidateOn` | mutation tags that drop this key |
| (engine) | dedupe in-flight fetches per key |

`WorkspaceSession` exposes `query(_ key) -> QueryHandle<Value>` (`status`, `data`, `error`). Views subscribe; they do not `.task { load }` on appear as the primary path.

Typed convenience accessors (e.g. `session.sourcesList`) are fine if they are one-line wrappers over `query(.sourcesList)` — **not** hard-coded store properties that require session changes for every new page.

### Layer 3 — Place registry (declarative extensibility)

One **`PlaceSpec`** per **presentable place** (not per sidebar section). A section can have several places (list root, entity detail, master–detail with optional selection).

```swift
struct PlaceSpec {
    let id: PlaceID                         // .sourcesList, .sourceDetail, .sourceFields, …
    let section: WorkspaceSection
    let matches: (WorkspaceLocation) -> String?   // deep id if any; nil = list root
    let queries: (ProjectKey, String?) -> [CatalogQueryKey]
    let invalidateGroups: [InvalidationGroup]
    // view: registered in presentation table (see below)
}
```

**Examples (initial set):**

| Place | `matches(location)` | Queries |
| --- | --- | --- |
| Sources list | `section == .sources && sourceId == nil` | `.sourcesList`, `.sourceTypesList` (filter/add) |
| Source detail | `section == .sources && sourceId` | `.sourceWorkspace(id)` |
| Source fields | `section == .sourceFields` | `.metadataFieldsList` |
| Source fields (row) | `… && fieldId` | same list (selection only) |
| Source types (row) | `section == .sourceTypes && typeId` | `.sourceTypesList`, `.metadataFieldsList`, `.typeSuggestions(id)` |

**Adding a future File detail page** = new optional on `WorkspaceLocation`, new `CatalogQueryKey` cases, new `PlaceSpec` rows, new view registration — **no changes to the cache engine**.

`registry.resolve(location) -> ResolvedPlace?` hides optional-id sprawl; callers do not switch on five deep-id fields.

### Presentation routing — split Sources list and detail

Replace fat `SourcesView` + internal `if/else` with a **location-driven host**:

```text
WorkspaceDestinationHost(location: currentLocation, session: session)
  → PlaceRegistry.view(for: resolvedPlace)
  → SourcesListView | SourcePageView | SourceFieldsView | …
```

- **Sources list** and **source detail** are two registered places, same section; only the list is reachable from the sidebar (detail via list, omnibar, breadcrumbs, history).
- The router `if` keys on **`resolve(location)`** / **`currentLocation`**, not `openedSourceID` inside a parent.
- List and detail share no parent view lifecycle; coupling is **cache writes** (patch or bust) — not a shared view model.

Vocabulary destinations stay one view (master–detail both visible) but still register as place(s) and use query handles instead of load-on-appear.

### Cache updates — patch vs invalidate vs stale-while-revalidate

| Mechanism | Role |
| --- | --- |
| **Patch** (`setQueryValue`) | Mutation sync: detail save writes updated row into `sourcesList` and `sourceWorkspace(id:)` **synchronously**. Title/thumbnail/list-row sync uses this — same as today's `applyUpdatedSource`. |
| **Invalidate** | Create, delete, or changes too hard to patch: mark key stale; refetch on next `ensureQuery` (navigation `apply` or subscribed view). `invalidate()` alone does not background-fetch. |
| **Stale-while-revalidate** | **Navigation reads only:** Back, section return, or refetch-after-invalidate when old value still in handle. Show cached frame immediately; async reload updates in place. **Not** the primary path for detail → list edits. |

Optional eager refetch after invalidate (warm cache before user navigates back) is allowed for create/delete; not required for identity/cover saves.

### Invalidation — declarative, like search projection hooks

```swift
enum CatalogMutation {
    case createdSource
    case updatedSource(id: String)
    case deletedMetadataField(id: String)
    // …
}
```

Registry entries declare which mutations **bust** which keys. Handlers call `session.invalidate(.createdSource)` (etc.) for bust-only paths; the session looks up affected `CatalogQueryKey`s from the registry. **`updatedSource` with row payload patches** list + workspace keys instead of invalidate-for-refetch. Features do not manually track every cache entry.

### Navigation commit flow

```text
go(to: location)
  → navigation stack updates (sync)
  → session.apply(location)        // registry resolves place, ensures queries
  → DestinationHost re-renders     // registry picks view
  → views read QueryHandles        // cached, loading, or stale-while-revalidate
```

Back/Forward uses the same path. Cache keys are **location-derived**, not view-lifetime-derived.

---

## Parallel to search registry

| Search (`core/search`) | Session / place layer |
| --- | --- |
| `KindSpec` in `registry.go` | `PlaceSpec` in `PlaceRegistry` |
| Maps hit → `WorkspaceLocation` | Maps `WorkspaceLocation` → place + queries |
| Declares field weights | Declares `CatalogQueryKey`s + invalidation |
| Add kind → registry row + projector | Add place → registry row + query loader + view |
| Omnibar calls `go(to: hit.location)` | Host loads queries for resolved place |

**`add-workspace-place`** skill (authored in S4-09): merged checklist for history + cache + presentation — mirror of [`add-searchable-kind`](../../.cursor/skills/add-searchable-kind/SKILL.md) and [`add-workspace-location`](../../.cursor/skills/add-workspace-location/SKILL.md). See [`.cursor/skills/add-workspace-place/SKILL.md`](../../.cursor/skills/add-workspace-place/SKILL.md).

**Add-a-place checklist (draft):**

```
- [ ] WorkspaceLocation: optional deep id field (if new kind)
- [ ] WorkspaceSection + sidebar (if new destination)
- [ ] PlaceSpec: matches + queries + invalidation
- [ ] CatalogQueryKey + loader in CatalogQueryRegistry
- [ ] View(s) registered in presentation table
- [ ] go(to:) call sites (row, cross-link, omnibar)
- [ ] Search KindSpec (if omnibar-searchable)
- [ ] NavigationHistoryTests + session cache tests
```

---

## Recommended direction

Treat as a **platform spike** before adding more destinations:

1. **`CatalogQueryRegistry` + `WorkspaceSession`** — generic cache engine (query keys, loaders, dedupe, invalidation).
2. **`PlaceRegistry` + `WorkspaceDestinationHost`** — declarative place → queries → view; location-driven routing.
3. **Split Sources list and detail** — two places, two views; remove fat `SourcesView` container and `openedSourceID` gating.
4. **Synchronous place application** — UI and host read `currentLocation`; `session.apply(location)` on navigation commit (no `Task`-deferred reconcile for presentation).
5. **Stale-while-revalidate** — show cached data or skeleton for the **target** place while queries resolve; Back to a warm source is instant.
6. **Migrate Source fields / types** onto query handles; type suggestions as keyed query, not reconcile waterfall.
7. **Go tiering (follow-on)** — lighter list RPCs, tiered workspace payload once Swift cache exists.

Do **not** ship only a synchronous reconcile tweak — it fixes the flash but leaves remount reloads, no extensibility, and source page cost untouched.

### What to avoid

- `WorkspaceSession.sourcesStore` as a permanent hard-coded property.
- Per-view `.task { load }` as primary hydration.
- Copy-paste `reconcileNavigation` per feature.
- One monolithic “fetch everything” RPC instead of composable query keys.
- Burying list/detail `if/else` inside a feature container instead of the place registry / destination host.

---

## Suggested module layout (macOS)

```text
Features/Workspace/
  WorkspaceNavigation.swift
  WorkspaceLocation.swift
  WorkspaceDestinationHost.swift     ← single router (replaces fat section switch logic)
  Session/
    CatalogQueryKey.swift
    CatalogQueryRegistry.swift       ← declarative loaders + invalidation
    PlaceRegistry.swift              ← location → place → queries → view id
    WorkspaceSession.swift           ← generic cache engine
    QueryHandle.swift
Features/Sources/
  SourcesListView.swift              ← split from SourcesView
  SourcePageView.swift               ← unchanged role; reads query handle
  …
```

Feature folders contain views + mutation actions that call `session.invalidate(…)`. They do not own fetch-on-appear lifecycle.

---

## Migration sketch

1. Add `Session/` types: `CatalogQueryKey`, `CatalogQueryRegistry`, `QueryHandle`, `WorkspaceSession` (engine only; FakeStore-friendly tests).
2. Add `PlaceRegistry` with initial places; wire `WorkspaceDestinationHost`.
3. Register query loaders for sources list, source types list, metadata fields list, source workspace, type suggestions.
4. **Sources spike:** split `SourcesListView` / keep `SourcePageView`; register `.sourcesList` and `.sourceDetail`; host routes from `currentLocation`; drop `openedSourceID` view gate; mutations invalidate list/workspace keys.
5. Wire `session.apply(location)` into navigation path (`WorkspaceView` or navigation coordinator).
6. Migrate Source fields / types to query handles + place specs.
7. Replace `WorkspaceContent` section `switch` with `WorkspaceDestinationHost` (optional view retention later if needed).
8. Author **`add-workspace-place`** skill from the checklist above.
9. Tests: sync first frame (no list flash); cache hit on Back; section switch no refetch when valid; invalidation on mutate; registry coverage for each place.

---

## Metrics to capture before/after

- Time from `go(to:)` to first **correct** frame (no wrong subtree).
- Time to interactive per place (list usable, detail usable).
- FFI call count per navigation type.
- Cache hit rate for repeated places in-session.
- Queue wait: Swift dispatch → Go handler start (optional debug logging).

---

## Explicitly out of scope (for this spike)

- Scroll/search restoration in history entries.
- Cross-project cache.
- Multi-window session sharing.
- Replacing hand-rolled history with `NavigationStack`.
- Generic persisted deep-id envelope on `WorkspaceLocation` (JSON stays per-kind optional fields; registry abstracts in Swift).

---

## Deployment

Sequenced PRs: [`docs/deployment-plan/archive/spike-4/deployment-plan.md`](../deployment-plan/archive/spike-4/deployment-plan.md) (S4-01…S4-09; S4-10+ descoped).

| Step | Summary |
| --- | --- |
| S4-01 | Query cache engine (tests only) |
| S4-02 | `CatalogQueryRegistry` + loaders + invalidation |
| S4-03 | `PlaceRegistry` + `resolve(location)` |
| S4-04 | Wire session + sync `apply(location)` |
| S4-05 | `WorkspaceDestinationHost` (shim to existing views) |
| S4-06 | Sources: split list/detail, cache-backed list |
| S4-06b | Source page workspace cache (optional split from S4-06) |
| S4-07 | Source fields migration |
| S4-08 | Source types migration (+ suggestions query) |
| S4-09 | Skill, docs, cleanup, dogfood metrics |
| S4-10+ | Optional Go RPC tiering |

---

## Related

- [`docs/deployment-plan/archive/spike-4/deployment-plan.md`](../deployment-plan/archive/spike-4/deployment-plan.md) — PR sequence (archived).
- [`docs/deployment-plan/archive/spike-3/navigation-history.md`](../deployment-plan/archive/spike-3/navigation-history.md) — history behavior (done).
- [`docs/ideas/archive/catalog-access-serialization.md`](archive/catalog-access-serialization.md) — catalog session (done).
- [`core/search/registry.go`](../../core/search/registry.go) — declarative registry pattern to mirror.
- [`.cursor/skills/add-workspace-location/SKILL.md`](../../.cursor/skills/add-workspace-location/SKILL.md) — wiring new places (to extend).
- [`.cursor/skills/add-searchable-kind/SKILL.md`](../../.cursor/skills/add-searchable-kind/SKILL.md) — search registry (parallel track).
