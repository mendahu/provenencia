# Deployment Plan

Spike 3 sequenced PRs: **navigation history** + **omnibar search**, including the small layout moves from the App Layout / Omnibar Results boards.

Authoritative behavior: [`navigation-history.md`](navigation-history.md), [`omnibar-search.md`](omnibar-search.md). Visual summary: [`design/`](design/). Spike overview: [`README.md`](README.md). Finished steps: [`completed.md`](completed.md).

## Status

**Spike 3 dogfood complete** (S3-01…S3-12; S3-06 skipped). Optional later slices (**S3-13+**) remain below. History: [`completed.md`](completed.md).

## Goal (dogfood bar)

A researcher can:

1. Move around Sources / Source types / Source fields (and Source pages) with **Back / Forward**, jump menus, and clickable toolbar breadcrumbs — stack persisted across relaunch via catalog **`project.uuid`**.
2. Find Sources, source types, and source fields from one toolbar **omnibar** (`⌘K`), pick a hit, and land via the same `go(to:)` history.
3. No longer see per-destination list search (Sources list / vocabulary panes).

## Layout moves (from Design boards)

Implemented in the PRs that owned chrome (see [`design/README.md`](design/README.md); boards in Claude Design):

| Change | From → To |
| --- | --- |
| Back / Forward | New leading pair on **main-column** toolbar (~46px, content gutter padding) |
| Jump menus | Long-press (~400ms) or secondary-click on Back/Forward; raised surface; rich rows (icon \| trail \| mono ref) |
| Breadcrumbs | Hoist to toolbar flex-grow (`PVBreadcrumbs`); ancestor segments call `go(to:)`; leaf non-navigating |
| Source-page trail | **Remove** `SourcePageIdentityHeader` breadcrumbs / hierarchical `closeSource` back |
| Page title | Stays in **content** (display heading), not between nav buttons |
| Omnibar field | Trailing toolbar `Input` (~420px / max 60%), search icon, placeholder, `⌘K` suffix; always visible |
| Results dropdown | Anchored under field; widens left to ~640px; shared `PVOmnibarHitRow` skeleton; empty query → **no panel** until enough typed (board: second character) |
| List search | Delete Sources + vocabulary search chrome once omnibar covers those kinds |

Keyboard (board): `⌘[` Back, `⌘]` Forward; `⌘K` focuses omnibar.

---

## Optional later slices

```text
S3-01…S3-12 (dogfood done)
        │
        ▼
S3-13+  More kinds / depth (optional)
```

### S3-13+ — Later slices (same spike or next)

Not required for Spike 3 dogfood if Sources + types + fields search well:

| Step | Scope |
| --- | --- |
| **S3-13** | Register Files / Artifacts (own hits or deeper Source rollup) per omnibar S6+ |
| **S3-14+** | Interpretation / Conclusion kinds when those destinations exist |

---

## Suggested PR titles (why-focused)

| Step | Title sketch |
| --- | --- |
| S3-04 | Add first-class workspace navigation history |
| S3-05 | Move Back/Forward and breadcrumbs into the main toolbar |
| S3-08 | Index the catalog with FTS5 for project search |
| S3-09 | Rank search by workspace context and catalog refs |
| S3-10 | Ship the omnibar and retire per-destination search |
| S3-11 | Tolerate typos in catalog omnibar search |
| S3-12 | Harden Spike 3 nav history and omnibar for dogfood |

---

## Parallelism

| Track | Steps |
| --- | --- |
| **Design** | S3-01, S3-02 — done ([`completed.md`](completed.md)) |
| **Core / FFI** | S3-03 → S3-07 → S3-08 → S3-09 → S3-11 — done |
| **Mac workspace** | S3-04 → S3-05 → S3-10 → S3-12 — done |

---

## Explicit non-goals (keep PRs honest)

- [x] No Interpretation / Conclusion destinations or search kinds required for Spike 3 done
- [x] No project Files list (Spike 2 descoped)
- [x] No short human `PRJ-…` project ref
- [x] No SwiftUI `NavigationStack` for session history
- [x] No permanent Swift-side `LIKE` search engine
- [x] No dual list-search chrome after S3-10
- [x] No edit undo/redo; no cross-project search; no AI “ask the catalog”
- [ ] No product `VERSION` bump for docs-only; bump when shipping a release that includes these PRs

---

## Definition of done

Jake can, on his MacBook:

1. [x] Open a project, navigate Sources ↔ Source page ↔ vocabulary, use toolbar Back/Forward and jump menus, relaunch, and land where he left off.
2. [x] Focus omnibar with `⌘K`, find a Source / type / field by title or ref, open it, then Back out.
3. [x] See **no** search fields on the Sources list or vocabulary list panes.
4. [x] Confirm `project.uuid` exists in SQLite and `navigation/{uuid}.json` under Application Support.

When a step finishes, move its write-up to [`completed.md`](completed.md) and leave optional later slices here.
