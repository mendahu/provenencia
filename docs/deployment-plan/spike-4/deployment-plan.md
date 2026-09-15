# Deployment Plan — Spike 4

Workspace session, catalog query cache, and declarative place registry. Authoritative design: [`page-navigation-performance.md`](../../ideas/page-navigation-performance.md).

## Status

**In progress.** Completed steps: [`completed.md`](completed.md). Execute **S4-05 → S4-09** in order unless noted. **S4-10+** are optional Go-side follow-ons after the Mac cache exists.

## Goal (dogfood bar)

1. No Sources list flash on Back/Forward or list → detail; presentation follows `currentLocation` on the same frame.
2. Return to Sources list, Source page, or vocabulary after a sidebar detour **without refetch** when session cache is valid.
3. Back to a source visited earlier in the session is **instant** (workspace cache hit).
4. New destinations can register places via registry rows — infrastructure is not Sources-shaped.

## PR sequence

```text
S4-01  Query cache engine (no UI) ✓
  │
  ▼
S4-02  CatalogQueryRegistry (loaders + invalidation map) ✓
  │
  ▼
S4-03  PlaceRegistry + resolve(location) ✓
  │
  ▼
S4-04  Wire session + sync apply(location) on navigation ✓
  │
  ▼
S4-05  WorkspaceDestinationHost (shim → old views)
  │
  ├──────────────────┬──────────────────┐
  ▼                  ▼                  ▼
S4-06              S4-07              S4-08
Sources            Source fields      Source types
(list + detail)    (vocabulary)       (vocabulary + suggestions)
  │                  │                  │
  └──────────────────┴──────────────────┘
                     │
                     ▼
                  S4-09  Docs, skill, cleanup, dogfood metrics
                     │
                     ▼ (optional)
                  S4-10+ Go read-model tiering
```

---

## Checklist

- [x] S4-01 — Catalog query cache engine → [`completed.md`](completed.md)
- [x] S4-02 — Catalog query registry → [`completed.md`](completed.md)
- [x] S4-03 — Place registry → [`completed.md`](completed.md)
- [x] S4-04 — Wire workspace session → [`completed.md`](completed.md)

---

## Cache update strategy

Cross-view sync (detail edit → list row) and navigation comfort (Back / section return) use **different mechanisms**. Do not rely on stale-while-revalidate for mutation sync.

| Mechanism | Use when | Behavior |
| --- | --- | --- |
| **Patch** | Identity/cover/title saves; any mutation whose response includes enough data to update cached rows | Write updated values into affected `QueryHandle`s **synchronously** on MainActor — no refetch, no stale row flash |
| **Invalidate** | Create, delete, or changes too expensive or ambiguous to patch locally | Mark key stale; refetch only when something calls `ensureQuery` (typically `apply(location)` on navigation, or a subscribed view) |
| **Stale-while-revalidate** | **Navigation reads only** — Back, sidebar return, refetch after invalidate when prior value still in handle | Show cached frame immediately; async reload replaces data in place (`isFetching`) |

**S4-01 note:** `invalidate(_:)` marks stale and cancels in-flight work; it does **not** start a background refetch. Optional eager warm-up (e.g. `invalidateAndRefetch` or mutation helper calling `ensureQuery` after bust) is allowed when off-screen prefetch is worth the FFI cost — not required for title/thumbnail sync.

**Detail → list (Sources):** On save, **patch** both `sourceWorkspace(sourceId:)` and the matching row inside `sourcesList` (same role as today's `SourcesModel.applyUpdatedSource`). Invalidate is for create source, delete, and other paths where local patch is impractical.

---

## S4-05 — PR: WorkspaceDestinationHost

| | |
| --- | --- |
| **Depends on** | S4-04 |
| **Title sketch** | Add workspace destination host for place-based routing |
| **Deliverables** | `WorkspaceDestinationHost.swift`: reads `navigation.currentLocation`, `registry.resolve`, switches on presentation id. **Shim phase:** maps each presentation id to existing views (`SourcesView`, `SourceFieldsView`, `SourceTypesView`). Replace `WorkspaceContent` inner `switch section` with host. Sidebar + toolbar unchanged. |
| **Tests** | Host mounts correct view type per location (list vs detail not split yet — still `SourcesView` for both source places). |
| **Dogfood** | No user-visible change; all destinations still work. |
| **Out** | Sources split; query-backed views. |

---

## S4-06 — PR: Migrate Sources (split list + detail)

| | |
| --- | --- |
| **Depends on** | S4-05 |
| **Title sketch** | Migrate Sources to session cache and split list from detail |
| **Deliverables** | Split `SourcesView` → `SourcesListView` + keep `SourcePageView`. Host routes `sourcesList` → list, `sourceDetail` → page from **`currentLocation.sourceId`** (remove `openedSourceID` view gate). List reads `session.query(.sourcesList)` / `.sourceTypesList`; drop list `.task { load }` and async `reconcileNavigation`. Row click / create still `go(to:)`. **Mutations:** create source → invalidate list keys (+ optional eager refetch); identity/cover save → **patch** `sourcesList` row + `sourceWorkspace(id:)` synchronously (replace `applyUpdatedSource`; do not invalidate-for-refetch on title/thumbnail). Delete fat `SourcesView` or reduce to thin re-export. Update/add `SourcesModel` tests → store/query tests. |
| **Tests** | No list flash (location-driven tree); list data from cache; invalidate on create; **edit title/cover on page patches list row without second list FFI**; navigation tests still pass. |
| **Dogfood** | Sources list ↔ detail via click, Back, omnibar: correct frame immediately; list loads once per session. |
| **Out** | Source page workspace cache (S4-06b if split, or same PR below). |

**Note:** If this PR is too large, split into **S4-06a** (split views + location routing + list on cache) and **S4-06b** (source page on workspace cache — below).

---

## S4-06b — PR: Source page workspace cache (optional split from S4-06)

| | |
| --- | --- |
| **Depends on** | S4-06 (or combined with it) |
| **Title sketch** | Load source detail from workspace session cache |
| **Deliverables** | `SourcePageView` / `SourcePageModel` read `session.query(.sourceWorkspace(id))` instead of `.task { load() }`. Stale cache paints immediately; miss shows skeleton for **target** source id. Remove `.id(opened)` remount forcing full reload when switching sources (identity from location + cache key). Saves **patch** workspace + list keys (see [cache update strategy](#cache-update-strategy)); bust-only paths use invalidate. `refreshCoverFromStore` patches cache, not ad hoc bypass. |
| **Tests** | Back to prior source = cache hit, no second `getSourceWorkspace`; switch A → B → A; edit patches both caches without refetch. |
| **Dogfood** | Source page Back/Forward feels instant on revisits. |

*If combined with S4-06, omit S4-06b as a separate PR.*

---

## S4-07 — PR: Migrate Source fields

| | |
| --- | --- |
| **Depends on** | S4-06 (or S4-06b) |
| **Title sketch** | Migrate Source fields to workspace session cache |
| **Deliverables** | `SourceFieldsView` reads `metadataFieldsList` query handle. Remove `.task` / `reconcileNavigation` / `load(from:)` / `apply(from:)`. Selection from `currentLocation.fieldId` (sync); missing id → `fallbackToSectionRoot` after list ready. CRUD invalidates list query + `CatalogCounts.publish*`. Master–detail layout unchanged. |
| **Tests** | Row select via history; sidebar return without refetch; delete → fallback. |
| **Dogfood** | Fields vocabulary navigation matches Sources snappiness. |

---

## S4-08 — PR: Migrate Source types

| | |
| --- | --- |
| **Depends on** | S4-07 |
| **Title sketch** | Migrate Source types to session cache including suggestions |
| **Deliverables** | `SourceTypesView` on query handles: types list, fields pool, `typeSuggestions(typeId)` keyed separately. Remove async reconcile waterfall. Selection UI instant from `currentLocation.typeId`; suggestions load async into query handle (detail shows loading state for suggestions block only). Assign/remove invalidates suggestions key (+ types list counts if needed). |
| **Tests** | History restore with type selected; suggestions deduped; no empty-detail flash. |
| **Dogfood** | Types navigation stable; suggestion panel loads without blocking selection chrome. |

---

## S4-09 — PR: Docs, skill, cleanup, and dogfood metrics

| | |
| --- | --- |
| **Depends on** | S4-08 |
| **Title sketch** | Document workspace place registry and remove legacy navigation loaders |
| **Deliverables** | Remove dead code: `openedSourceID`, `reconcileNavigation` helpers, unused `load(from:)` / `apply(from:)` on migrated models. Author [`.cursor/skills/add-workspace-place/SKILL.md`](../../../.cursor/skills/add-workspace-place/SKILL.md) (checklist from idea doc). Update [`page-navigation-performance.md`](../../ideas/page-navigation-performance.md) status → implemented / archive pointer. Update [`macos-client-patterns.md`](../../macos-client-patterns.md) § workspace session. Optional `#if DEBUG` navigation timing logs behind flag. Brief [`spike-4/completed.md`](completed.md) template. |
| **Tests** | Full `ProvenenciaTests` green; registry coverage test enumerates every `PlaceID` has presentation + queries. |
| **Dogfood** | Full spike-4 dogfood checklist (below). |
| **Out** | Go RPC changes. |

### Spike 4 dogfood checklist

- [ ] Sources list → detail → Back: no list flash; detail correct on first frame.
- [ ] Source A → Source B → Back to A: page instant (cache hit).
- [ ] Sources → Fields → Sources: list not refetched (cache hit).
- [ ] Omnibar → source → Back → sidebar → Sources: same.
- [ ] Create source, edit title on page, return to list: row updated (sync patch on save, not SWR refetch).
- [ ] Fields/types: history restore, sidebar hop, no cold reload when cache valid.
- [ ] Relaunch: history restore still works; first paint may load (cold cache) then warm.

---

## Optional — Go read-model tiering (S4-10+)

Execute **after S4-09** when Swift cache proves which RPCs hurt. Can parallelize sub-steps.

| Step | Scope |
| --- | --- |
| **S4-10** | Lighter `ListSources` (defer per-row cover enrichment; client uses `ThumbnailCache` / type icon). |
| **S4-11** | Tiered `GetSourceWorkspace` or split RPC (identity shell first, metadata/artifacts lazy) — Swift cache loads tiers as separate query keys. |
| **S4-12** | Prefetch hint RPC (optional): `go(to:)` target warms cache from Go side. |

**Out:** Required for Spike 4 dogfood if Mac cache hits are acceptable on local SQLite.

---

## Suggested PR titles (why-focused)

| Step | Title sketch |
| --- | --- |
| S4-01 | Add workspace catalog query cache engine |
| S4-02 | Register catalog query loaders for workspace navigation |
| S4-03 | Add declarative workspace place registry |
| S4-04 | Wire workspace session and sync navigation apply |
| S4-05 | Add workspace destination host for place-based routing |
| S4-06 | Migrate Sources to session cache and split list from detail |
| S4-06b | Load source detail from workspace session cache |
| S4-07 | Migrate Source fields to workspace session cache |
| S4-08 | Migrate Source types to session cache including suggestions |
| S4-09 | Document workspace place registry and remove legacy navigation loaders |
| S4-10 | Lighten Sources list RPC for navigation cache |

---

## Parallelism

| Track | Steps |
| --- | --- |
| **Mac session (critical path)** | S4-01 → S4-09 |
| **Go tiering (optional)** | S4-10+ after S4-06 or S4-09 |

No design-board dependency. Spike is infrastructure + migration.

---

## Definition of done

Jake can, on his MacBook:

1. Navigate the workspace without Sources list flash and with session-warm returns (checklist above).
2. Point to `PlaceRegistry` + `CatalogQueryRegistry` as the extension point for Files / Interpretation later.
3. Follow `add-workspace-place` skill to add a new cached place without copying `.task` load boilerplate.

---

## Metrics (capture in S4-09 or during migration)

Record before S4-06 and after S4-08:

- Time from `go(to:)` to first correct frame.
- Time to interactive (list populated, source page header populated).
- FFI call count: same-section deep nav, cross-section return, Back to cached source.
