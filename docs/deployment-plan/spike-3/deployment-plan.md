# Deployment Plan

Spike 3 sequenced PRs: **navigation history** + **omnibar search**, including the small layout moves from the App Layout / Omnibar Results boards.

Authoritative behavior: [`navigation-history.md`](navigation-history.md), [`omnibar-search.md`](omnibar-search.md). Visual summary: [`design/`](design/). Spike overview: [`README.md`](README.md). Finished steps: [`completed.md`](completed.md).

## Status

**In progress.** Open steps below; completed steps live in [`completed.md`](completed.md) (S3-01…S3-04).

## Goal (dogfood bar)

A researcher can:

1. Move around Sources / Source types / Source fields (and Source pages) with **Back / Forward**, jump menus, and clickable toolbar breadcrumbs — stack persisted across relaunch via catalog **`project.uuid`**.
2. Find Sources, source types, and source fields from one toolbar **omnibar** (`⌘K`), pick a hit, and land via the same `go(to:)` history.
3. No longer see per-destination list search (Sources list / vocabulary panes).

## Layout moves (from Design boards)

Implement in the PR steps that own chrome (see [`design/README.md`](design/README.md); boards in Claude Design):

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

## Open PR sequence

Steps are **`S3-NN`**. **Depends on** is the merge gate. Prefer many small PRs; do not fold search engine work into the first toolbar PR.

```text
S3-03 project.uuid (done) ───────────────────┐
        │                                     │
        ▼                                     ▼
S3-04  history (done)                       S3-07  SearchCatalog RPC shell
        │                                     │
        ▼                                     ▼
S3-05  Toolbar chrome (nav + crumbs + field) S3-08  FTS5 projection
        │                                     │
        │                                     ▼
        │                               S3-09  Context ranking + ref path
        │                                     │
        └──────────────┬──────────────────────┘
                       ▼
                 S3-10  Omnibar results + wire search + remove list search
                       │
                       ▼
                 S3-11  Fuzzy / typo
                       │
                       ▼
                 S3-12  Dogfood polish (optional S3-13+ kinds)
```

### S3-05 — PR: Main-column toolbar chrome (nav + breadcrumbs + omnibar field shell)
| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-04 (done); S3-01 board (done) |
| **Deliverables** | Main-column header per App Layout: ~46px, bottom border, content-leading padding. Leading Back/Forward `IconButton`s (disabled at stack ends). Long-press / secondary-click jump menus (nearest **15**, rich rows). Toolbar `PVBreadcrumbs` with navigable ancestors via `go(to:)`. **Strip** Source-page identity-header breadcrumbs and hierarchical list-back. Trailing omnibar **field** always visible (`⌘K` focuses); no results panel required yet (or empty/disabled until S3-10). L10n for new chrome. |
| **Context** | Claude Design App Layout board; [`design/README.md`](design/README.md). |
| **Out** | Search RPC, FTS, results dropdown, deleting list search (S3-10). |
| **Dogfood** | Chrome matches the board; Back/Forward + jump menu + breadcrumb clicks feel browser-like. |

### S3-06 — PR: (optional split) History jump-menu polish

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-05 |
| **Deliverables** | Only if S3-05 ships a minimal menu: refine jump-menu row anatomy (destination icon \| `Destination › title` \| mono ref), hold vs click swallow, Escape/outside dismiss — parity with App Layout Frame 8. Skip this ID if S3-05 lands the full menu. |
| **Out** | Omnibar results (different surface; shared elevation language only). |

### S3-07 — PR: Searchable-kind registry + `SearchCatalog` RPC shell

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — (can parallel S3-04…S3-05 after S3-03 if location payload shape is agreed; otherwise after S3-04) |
| **Deliverables** | Go searchable-kind registry (Sources, source types, source fields). Protobuf `SearchCatalog` + FFI handler + Hit DTO (`kind`, `id`, `ref?`, `title`, `subtitle?`, `match_reason?`, `location`). Request carries query + current `WorkspaceLocation` (at least section). FakeStore. Naïve table scan **behind the same API** only as a bridge. Tests without UI. |
| **Context** | [`omnibar-search.md`](omnibar-search.md) § Incremental delivery S1; [`add-ffi-handler`](../../../.cursor/skills/add-ffi-handler/SKILL.md). |
| **Out** | FTS5 tables; Swift omnibar dropdown; removing list search. |
| **Dogfood** | RPC returns ranked-enough hits for known titles/refs in tests / debug. |

### S3-08 — PR: FTS5 projection

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-07 |
| **Deliverables** | Enable `fts5` on the embedded amalgamation if not already. Migration: FTS search documents for Source / type / field roots. Field weights in registry. Incremental reproject on writes; rebuild/heal on Open/migrate. Swap naïve scanner for FTS retrieval. Roll Source notes / metadata / artifact-filename text into the **Source** document (not separate hits). |
| **Context** | [`omnibar-search.md`](omnibar-search.md) S2; [`application-stack.md`](../../application-stack.md) §10. |
| **Out** | Context boosts / ref fast path polish (S3-09); fuzzy (S3-11); Files/Artifact as own hit kinds. |
| **Dogfood** | Edit a Source title/note → search updates; relevance acceptable on a real catalog. |

### S3-09 — PR: Context ranking + ref fast path

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-08 |
| **Deliverables** | Section/kind context boosts from request location. Exact / prefix **ref** promotion (`SRC-…` etc.). Cheap `match_reason` when match is in rolled-up child text. Bounded hit list (e.g. top 20–50). |
| **Context** | [`omnibar-search.md`](omnibar-search.md) S3; multi-word term-coverage rules in that note. |
| **Out** | Fuzzy; omnibar UI. |
| **Dogfood** | Paste `SRC-…` → obvious top hit; browsing Sources floats Source hits without hiding vocabulary. |

### S3-10 — PR: Omnibar results UI + wire search + remove list search

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-05 (field shell); S3-09 (engine quality bar for dogfood); S3-02 board (done) |
| **Deliverables** | Results dropdown per Omnibar Results board: shared `PVOmnibarHitRow` (lead / title / kind chip / secondary / ref / match context / selected). Anchoring, loading, empty, keyboard selection, Esc/outside dismiss. Debounced `SearchCatalog`; select hit → `go(to: location)` → dismiss. **Delete** Sources list search and `VocabularyListPane` search chrome + query-only filter plumbing. Extend thumbnail/glyph modes as the board requires. L10n. |
| **Context** | Claude Design Omnibar Results board; [`omnibar-search.md`](omnibar-search.md) S4. Board rules: no panel on empty/short query; flat engine order; kind chip disambiguates. |
| **Out** | Fuzzy; Files/Artifact first-class hit kinds; Interpretation kinds. |
| **Dogfood** | End-to-end find → navigate → Back returns; no dual search chrome. |

### S3-11 — PR: Fuzzy / typo matching

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-10 (or S3-09 if tuning without UI) |
| **Deliverables** | FTS5 trigram and/or Go fuzzy on a **shortlist** only; threshold tuning. Never full-catalog fuzzy scan. |
| **Context** | [`omnibar-search.md`](omnibar-search.md) S5. |
| **Out** | Cross-root association search; NL/AI queries. |
| **Dogfood** | Common typos recover without flooding garbage. |

### S3-12 — PR: Spike 3 dogfood polish

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-10; S3-11 preferred |
| **Deliverables** | Copy/a11y/identifiers, restore edge cases, jump-menu + omnibar elevation consistency, regression tests, closeout notes. Fix anything that blocks daily use of history + omnibar. |
| **Out** | New catalog layers; project Files browser. |

### S3-13+ — Later slices (same spike or next)

Not required to close Spike 3 dogfood if Sources + types + fields search well:

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
| S3-07 | Add a catalog SearchCatalog RPC and kind registry |
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
| **Core / FFI** | S3-03 done → S3-07 → S3-08 → S3-09 → (S3-11) |
| **Mac workspace** | S3-04 done → S3-05 (→ S3-06) → S3-10 → S3-12 |

S3-07+ may proceed beside S3-05 once Hit `location` matches `WorkspaceLocation`. **S3-10** is the integration gate.

---

## Explicit non-goals (keep PRs honest)

- [ ] No Interpretation / Conclusion destinations or search kinds required for Spike 3 done
- [ ] No project Files list (Spike 2 descoped)
- [ ] No short human `PRJ-…` project ref
- [ ] No SwiftUI `NavigationStack` for session history
- [ ] No permanent Swift-side `LIKE` search engine
- [ ] No dual list-search chrome after S3-10
- [ ] No edit undo/redo; no cross-project search; no AI “ask the catalog”
- [ ] No product `VERSION` bump for docs-only; bump when shipping a release that includes these PRs

---

## Definition of done

Jake can, on his MacBook:

1. Open a project, navigate Sources ↔ Source page ↔ vocabulary, use toolbar Back/Forward and jump menus, relaunch, and land where he left off.
2. Focus omnibar with `⌘K`, find a Source / type / field by title or ref, open it, then Back out.
3. See **no** search fields on the Sources list or vocabulary list panes.
4. Confirm `project.uuid` exists in SQLite and `navigation/{uuid}.json` under Application Support.

When a step finishes, move its write-up to [`completed.md`](completed.md) and leave the open list here as the checklist.
