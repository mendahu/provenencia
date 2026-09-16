# Spike 4 — Completed steps

Finished Spike 4 work kept for history. Spike overview: [`README.md`](README.md). Descoped notes (S4-10+): [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S4-NN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S4-01](#s4-01--pr-catalog-query-cache-engine) | PR | Workspace catalog query cache engine (no UI wiring) |
| [S4-02](#s4-02--pr-catalog-query-registry) | PR | Declarative query loaders, mutation map, patch API |
| [S4-03](#s4-03--pr-place-registry) | PR | Declarative place registry + resolve(location) |
| [S4-04](#s4-04--pr-wire-workspace-session) | PR | Wire session into navigation; warm cache on commit |
| [S4-05](#s4-05--pr-workspace-destination-host) | PR | Place-based destination host (shim to existing views) |
| [S4-06](#s4-06--pr-migrate-sources) | PR | Sources list on session cache; split list from detail |
| [S4-06b](#s4-06b--source-page-workspace-cache) | PR | Source page on `sourceWorkspace` handle (combined with S4-06) |
| [S4-07](#s4-07--pr-migrate-source-fields) | PR | Source fields on `metadataFieldsList` handle |
| [S4-08](#s4-08--pr-migrate-source-types) | PR | Source types + type suggestions on session cache |
| [S4-09](#s4-09--pr-docs-skill-cleanup) | PR | Skill, docs, cleanup; Spike 4 dogfood sign-off |

---

## Steps

### S4-01 — PR: Catalog query cache engine

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — |
| **Deliverables** | Done. `macos/App/Features/Workspace/Session/`: `ProjectKey`, `CatalogQueryKey`, `QueryStatus`, `QueryHandle`, `WorkspaceSession` with get-or-load, in-flight dedupe, stale-while-revalidate (`isFetching` while prior value visible), and `invalidate(_:)` / `invalidateAll(matching:)`. `@Observable`, `@MainActor`. **No** production wiring; no destination changes. |
| **Tests** | Done. `macos/ProvenenciaTests/WorkspaceSessionTests.swift`: cache hit, concurrent dedupe, invalidate + bulk invalidate, stale-while-revalidate, loader error propagation, FakeStore `sourcesList` integration. |
| **Dogfood** | App unchanged. |
| **Out** | Place registry; view migration; Go changes. |

**Landed:** Session types compile and are unused by production UI. Extension points left for S4-02 (`CatalogQueryRegistry`) and S4-04 (environment injection).

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSessionTests
```

---

### S4-02 — PR: Catalog query registry

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S4-01 |
| **Deliverables** | Done. `CatalogMutation.swift`, `CatalogQueryStalePolicy.swift`, `CatalogQueryRegistry.swift`: declarative loaders for all five initial keys, `sessionFresh` stale policy (reserved `refetchOnRevisit`), and mutation invalidation map. `WorkspaceSession` gains registry init (default `.standard`), `query(_:)`, `setQueryValue(_:value:)`, and `apply(_:)` with **patch** for `.updatedSource` and **bust** for create/delete/CRUD/suggestion mutations. **No** production UI wiring. |
| **Tests** | Done. `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`: registry loads every key, cache hit without refetch, `setQueryValue` skips loader, patch vs invalidate paths, field/type CRUD and suggestion mutation maps. |
| **Dogfood** | App unchanged. |
| **Out** | Place registry; view migration; Go changes. |

**Landed:** Callers can load via `session.query(key)` without ad hoc store switches. Mutation dispatch follows [cache update strategy](deployment-plan.md#cache-update-strategy): identity/cover sync patches list + workspace; bust-only paths invalidate the correct keys.

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/CatalogQueryRegistryTests \
  -only-testing:ProvenenciaTests/WorkspaceSessionTests
```

---

### S4-03 — PR: Place registry

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S4-02 |
| **Deliverables** | Done. `PlaceID.swift`, `WorkspacePresentationID.swift`, `ResolvedPlace.swift`, `PlaceRegistry.swift`: priority-ordered specs mapping `WorkspaceLocation` → place id, presentation id, and `CatalogQueryKey`s. Covers sources list/detail, source fields (one spec for root + row), and source types list/detail (with suggestions). **No** production UI wiring. |
| **Tests** | Done. `macos/ProvenenciaTests/PlaceRegistryTests.swift`: section roots, deep ids, cross-section id ignore, query key table, full `PlaceID` coverage. |
| **Dogfood** | App unchanged. |
| **Out** | View host; `session.apply(location:)` (S4-04). |

**Landed:** `PlaceRegistry.standard.resolve(_:project:)` is the extension point for S4-04 cache warming and S4-05 presentation routing.

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/PlaceRegistryTests \
  -only-testing:ProvenenciaTests/CatalogQueryRegistryTests \
  -only-testing:ProvenenciaTests/WorkspaceSessionTests
```

---

### S4-04 — PR: Wire workspace session

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S4-03 |
| **Deliverables** | Done. `WorkspaceView` owns `@State WorkspaceSession` (projectDir + store) and injects `.environment(session)`. `WorkspaceSession.apply(location:)` resolves via `PlaceRegistry` and non-blocking `warmQuery` for each place key. `WorkspaceNavigation.onLocationCommit` fires synchronously from the private apply path (go, Back/Forward, jump index, fallback, `attachProject` restore). Destination views **unchanged** — legacy `.task` loads overlap briefly while cache warms ahead of S4-05+. Optional `PROVENENCIA_DEBUG_SESSION_APPLY=1` logging in DEBUG. |
| **Tests** | Done. `WorkspaceSessionTests`: apply(location) key table, non-blocking warm, cache hit. `WorkspaceNavigationTests`: `onLocationCommit` on go, history restore, Back/Forward/jump. |
| **Dogfood** | App behaves as today; cache warms in background on navigation. |
| **Out** | Destination host; removing legacy load paths; views reading query handles. |

**Landed:** Every navigation commit warms the place registry query keys. S4-05 can route presentation without changing when cache loads start.

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/WorkspaceSessionTests \
  -only-testing:ProvenenciaTests/WorkspaceNavigationTests \
  -only-testing:ProvenenciaTests/PlaceRegistryTests \
  -only-testing:ProvenenciaTests/CatalogQueryRegistryTests
```

---

### S4-05 — PR: Workspace destination host

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S4-04 |
| **Deliverables** | Done. `WorkspaceDestinationHost.swift`: resolves `navigation.currentLocation` via `PlaceRegistry`, switches on `WorkspacePresentationID`. Host maps presentations to destination views. `WorkspaceContent` uses the host instead of a section switch. `WorkspacePresentationID` is `CaseIterable`. Sidebar + toolbar unchanged. |
| **Tests** | Done. `WorkspaceDestinationHostTests.swift`: presentation routing table, both source presentations → sources destination, registry presentations known. |
| **Dogfood** | No user-visible change; all destinations still work. |
| **Out** | Sources list/detail split; query-backed views; removing legacy loaders. |

**Landed:** Workspace content routing is presentation-driven. S4-06 can split Sources and wire query handles without changing the outer host shell.

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/WorkspaceDestinationHostTests \
  -only-testing:ProvenenciaTests/PlaceRegistryTests \
  -only-testing:ProvenenciaTests/WorkspaceNavigationTests
```

---

### S4-06 — PR: Migrate Sources

| | |
| --- | --- |
| **Kind** | PR ([#113](https://github.com/mendahu/provenencia/pull/113)) |
| **Depends on** | S4-05 |
| **Deliverables** | Done. Deleted `SourcesView`; added `SourcesListView`. `WorkspaceDestinationHost` routes `.sourcesList` / `.sourcePage` from `currentLocation.sourceId`. `SourcesModel` reads `sourcesList` / `sourceTypesList` handles; create → invalidate + list patch; `syncCatalogCounts` on ready. Views warm queries in `.task` and observe `@Bindable QueryHandle` (no `session.query()` in `body`). |
| **Tests** | Done. `SourcesModelTests` rewritten for session warm, filter/sort, create, cache hit, `updatedSource` patch. |
| **Dogfood** | List ↔ detail via click, Back, omnibar: location-driven tree; no `openedSourceID` gate. |
| **Out** | Source page workspace cache (S4-06b — same PR). |

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/SourcesModelTests \
  -only-testing:ProvenenciaTests/WorkspaceDestinationHostTests
```

---

### S4-06b — Source page workspace cache

| | |
| --- | --- |
| **Kind** | PR (combined with S4-06, [#113](https://github.com/mendahu/provenencia/pull/113)) |
| **Depends on** | S4-06 |
| **Deliverables** | Done. `SourcePageView` / `SourcePageModel` sync from `sourceWorkspace` handle; `.task(id: sourceID)` warms query; fingerprint skips redundant section resets. Identity/cover → `session.apply(.updatedSource)`; other edits → `notifyWorkspaceMutated()`. `refreshCoverFromStore` patches cache via `setQueryValue`. |
| **Tests** | Done. `SourcePageModelTests` use `warmFromSession()` (invalidate + await `isFetching` for stale refetch). |
| **Dogfood** | Back to prior source uses cache; switching sources does not remount via `.id(opened)`. |

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/SourcePageModelTests
```

---

### S4-07 — PR: Migrate Source fields

| | |
| --- | --- |
| **Kind** | PR ([#113](https://github.com/mendahu/provenencia/pull/113)) |
| **Depends on** | S4-06 |
| **Deliverables** | Done. `SourceFieldsModel` reads `metadataFieldsList` via `queryHandle`; `warmFieldsQuery()` in view `.task`. Selection via `syncSelection(from:)` on history + handle changes; missing deep id → `fallbackToSectionRoot`. CRUD → `session.apply` + list patch + `syncCatalogCounts`. |
| **Tests** | Done. `SourceFieldsModelTests` use session warm helpers. |
| **Dogfood** | Fields vocabulary: history restore and sidebar return without cold list reload when cache valid. |

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/SourceFieldsModelTests
```

---

### S4-08 — PR: Migrate Source types

| | |
| --- | --- |
| **Kind** | PR ([#113](https://github.com/mendahu/provenencia/pull/113)) |
| **Depends on** | S4-07 |
| **Deliverables** | Done. `SourceTypesModel` on types/fields/suggestions handles; `warmListQueries()` / `warmSuggestions(for:)`. Selection sync from history; suggestions load async into keyed handle. Assign/remove → registry invalidation + cache patch. |
| **Tests** | Done. `SourceTypesModelTests` session warm + suggestion coverage. |
| **Dogfood** | Types selection instant from history; suggestions block shows loading without blocking chrome. |

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS' \
  -only-testing:ProvenenciaTests/SourceTypesModelTests
```

---

### S4-09 — PR: Docs, skill, cleanup

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S4-08 |
| **Deliverables** | Done. [`.cursor/skills/add-workspace-place/SKILL.md`](../../../../.cursor/skills/add-workspace-place/SKILL.md). Updated [`page-navigation-performance.md`](../../../ideas/page-navigation-performance.md) (implemented), [`macos-client-patterns.md`](../../../macos-client-patterns.md) § workspace session, [`add-workspace-location`](../../../../.cursor/skills/add-workspace-location/SKILL.md) (post-migration apply pattern). Optional `PROVENENCIA_DEBUG_NAV_TIMING=1` on navigation commit. Legacy loaders removed in S4-06–08 (`openedSourceID`, `reconcileNavigation`, `load(from:)`). Spike archived; S4-10+ descoped. |
| **Tests** | Full `ProvenenciaTests` green; `PlaceRegistryTests.registryCoversAllPlaceIDs` covers every `PlaceID`. |
| **Dogfood** | Spike 4 checklist signed off (below). |
| **Out** | Go RPC tiering (S4-10+) — descoped, not scheduled. |

**Spike 4 dogfood checklist (signed off):**

- [x] Sources list → detail → Back: no list flash; detail correct on first frame.
- [x] Source A → Source B → Back to A: page instant (cache hit).
- [x] Sources → Fields → Sources: list not refetched (cache hit).
- [x] Omnibar → source → Back → sidebar → Sources: same.
- [x] Create source, edit title on page, return to list: row updated (sync patch on save).
- [x] Fields/types: history restore, sidebar hop, no cold reload when cache valid.
- [x] Relaunch: history restore works; first paint may load (cold cache) then warm.

**Verify:**

```bash
cd macos && xcodebuild test -scheme Provenencia -destination 'platform=macOS'
```
