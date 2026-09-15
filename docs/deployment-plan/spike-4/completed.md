# Spike 4 — Completed steps

Finished Spike 4 work kept for history. Forward checklist: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S4-NN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S4-01](#s4-01--pr-catalog-query-cache-engine) | PR | Workspace catalog query cache engine (no UI wiring) |

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
