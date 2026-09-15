# Spike 4 — Completed steps

Finished Spike 4 work kept for history. Forward checklist: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S4-NN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S4-01](#s4-01--pr-catalog-query-cache-engine) | PR | Workspace catalog query cache engine (no UI wiring) |
| [S4-02](#s4-02--pr-catalog-query-registry) | PR | Declarative query loaders, mutation map, patch API |
| [S4-03](#s4-03--pr-place-registry) | PR | Declarative place registry + resolve(location) |

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
| **Deliverables** | Done. `PlaceID.swift`, `WorkspacePresentationID.swift`, `ResolvedPlace.swift`, `PlaceRegistry.swift`: priority-ordered specs mapping `WorkspaceLocation` → place id, presentation id, and `CatalogQueryKey`s. Covers sources list/detail, source fields (one spec for root + row), source types list/detail (with suggestions), and files placeholder. **No** production UI wiring. |
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
