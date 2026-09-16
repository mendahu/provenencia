# Deployment Plan — Spike 4

Workspace session, catalog query cache, and declarative place registry. Authoritative design: [`page-navigation-performance.md`](../../../ideas/archive/page-navigation-performance.md).

## Status

**Done.** Spike archived after dogfood (S4-01…S4-09). History: [`completed.md`](completed.md). **S4-10+** Go read-model tiering below is **descoped** (not scheduled open spike work).

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
S4-05  WorkspaceDestinationHost (shim → old views) ✓
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
                     ▼ (descoped)
                  S4-10+ Go read-model tiering
```

---

## Checklist

- [x] S4-01 — Catalog query cache engine → [`completed.md`](completed.md)
- [x] S4-02 — Catalog query registry → [`completed.md`](completed.md)
- [x] S4-03 — Place registry → [`completed.md`](completed.md)
- [x] S4-04 — Wire workspace session → [`completed.md`](completed.md)
- [x] S4-05 — Workspace destination host → [`completed.md`](completed.md)
- [x] S4-06 — Migrate Sources (list + detail split) → [`completed.md`](completed.md)
- [x] S4-07 — Migrate Source fields → [`completed.md`](completed.md)
- [x] S4-08 — Migrate Source types → [`completed.md`](completed.md)
- [x] S4-09 — Docs, skill, cleanup → [`completed.md`](completed.md)

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

## Completed migration (S4-06…S4-09)

Landings and dogfood sign-off: [`completed.md`](completed.md) (S4-06…S4-06b in PR [#113](https://github.com/mendahu/provenencia/pull/113); S4-09 closes docs/skill).

Extension point for new destinations: [`.cursor/skills/add-workspace-place/SKILL.md`](../../../../.cursor/skills/add-workspace-place/SKILL.md).

---

## Descoped — Go read-model tiering (S4-10+)

**Not scheduled.** Was optional follow-on after S4-09 if Swift cache profiling showed RPC cost; descoped after dogfood — local SQLite + session cache hits are acceptable. Kept for historical reference.

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
