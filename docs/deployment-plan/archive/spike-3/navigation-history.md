# Navigation history (back / forward)

**Status:** archived — Spike 3 complete. Requirements for navigation history and persistence. PR sequence: [`deployment-plan.md`](deployment-plan.md) (S3-03…S3-06).

Visual chrome: Claude Design **App Layout** board ([`design/README.md`](design/README.md)). This note owns **behavior**; the board is the visual source of truth for placement and menu layout.

## Problem

The workspace is becoming a small web of places: sidebar destinations (Sources, Source types, Source fields, …), list → **separate Source page**, master–detail selection inside vocabulary, later Node/entity pages, and cross-links (omnibar → anywhere).

Today, leaving a place is mostly one-way. Open a Source, jump elsewhere, open another Source — there is no **Back** that restores the previous view the way a browser would. Researchers who follow links will get stranded without a trail. Relaunch also forgets where they were.

## Goal

Browser-like **Back** and **Forward** (plus a **recent-entries** jump menu) over a **persisted session history** of views in the open project. On relaunch of that project, restore the stack **and** the location the researcher left on.

```text
… → Sources list → Source SRC-… → Source fields (row X) → Source SRC-… → …
         ← Back                                      Forward →
              long-press / secondary-click: jump menu
```

Not undo for edits. Not document-scroll history. **Navigation** history: which screen / selection was showing.

## Implementation posture (first-class, not a band-aid)

The product is early; **prefer correct infrastructure over minimal diff**. Navigation history is a **platform feature of the Mac workspace**, not a toolbar widget bolted onto today’s ad-hoc switches.

- **Single coordinator** — Every committed navigation (sidebar, Source open/close, vocabulary selection, toolbar breadcrumbs, future omnibar, cross-links) goes through `go(to:)` / Back / Forward / jump. Do not leave parallel paths (`closeSource()`, raw `selectedSection =`, etc.) that bypass the stack.
- **Wide wiring is expected** — Touching `WorkspaceModel`, destination models/views, Source page chrome, tests, and catalog `project.uuid` in one effort is fine. Do not ship a “Back only on Sources” or “history optional if restore fails quietly forever” half-measure.
- **Location is source of truth** — Feature UI applies a `WorkspaceLocation`; it does not own a second notion of “where we are” that drifts from the history store.
- **Persist and restore for real** — Project UUID + Application Support JSON + relaunch restore are part of the feature, not a follow-up. Failures should be explicit enough to fix (prune missing entities; don’t corrupt the stack silently into nonsense).
- **Design for the next destinations** — Per-kind deep fields and a durable project id are intentional so Files / Interpretation / omnibar plug into the same pipe later without a rewrite.
- **Tests** — Cover stack push/truncate/coalesce, jump menu index, persistence round-trip, and restore of at least Sources list ↔ Source page and one vocabulary selection. Prefer model/FakeStore tests over hoping the toolbar looks right.

Churn now is cheaper than a navigation retrofit after more destinations and links exist.

## What counts as a history entry

A stack entry must be enough to **restore** a view. MVP pieces:

| Piece | Example | MVP |
| --- | --- | --- |
| Sidebar destination | Sources, Source fields, Source types, … | **Required** |
| Deep location | Sources list vs Source page `id`/`ref`; selected field/type id; later File / Node / entity id | **Required** |
| UI ephemera | List search query, scroll offset, focus, dirty form drafts | **Omit** |

Push (or replace) when the user **commits** a navigation, including:

- Switching sidebar destination (**always** pushes, even if that destination’s last deep location is unchanged)
- Opening a Source page from the list (or any deep link / omnibar hit)
- Changing the selected row in a master–detail table (Source fields, Source types, …) — treat selection like a browser **query string**: same destination, different deep location, new entry so Back can return with that row preselected
- Any other `go(to:)` from chrome, toolbar breadcrumbs, or cross-links

**Do not** push on every keystroke in a search field, every focus change, or every dirty form edit.

**Same-entry coalescing** — navigating to a location identical to the current entry must not spam the stack (replace or no-op).

## Stack behavior (browser-shaped)

- **Back** — move to the previous entry; enable Forward.
- **Forward** — only after Back (or after jumping into an older entry); **cleared (truncated)** when the user navigates somewhere **new** from the middle of the stack.
- **Jump** — from the recent-entries menu, move the current index to any earlier or later entry without truncating the stack (same as choosing an item in a browser Back/Forward history menu). A **new** navigation after a jump still truncates forward from the new index.
- **Cross-destination** — one history per workspace window, not per sidebar item (otherwise Back cannot leave Sources).
- **Multiple windows (later)** — one history per window.
- **Stack length** — **persist up to 100** entries (drop oldest when exceeded). Jump menus show at most the **nearest 15** in that direction; keyboard Back/Forward still walk the full stored stack.

## Chrome (from App Layout board)

Toolbar sits in the **main column only** (not over the sidebar), ~46px tall, bottom border, content-leading padding aligned with the page gutter:

```text
[ ← ] [ → ]   Breadcrumbs …                    [ 🔍 omnibar          ⌘K ]
```

| Control | Requirement |
| --- | --- |
| **Back / Forward** | Leading pair of small icon buttons (`chevron-left` / `chevron-right`). Disabled when that direction has no entries. |
| **Click** | Single step Back or Forward. |
| **Long-press** | Hold ~400ms on Back or Forward opens that direction’s jump menu (do not also fire a step when the menu was opened via hold). |
| **Secondary-click** | Context-menu on Back or Forward also opens that direction’s jump menu. |
| **Jump menu** | Anchored under the pressed button; raised surface, subtle border, large shadow; width ~290–380px. |
| **Menu rows** | Grid: destination **icon** \| label \| optional **mono ref**. Destination-only entries show the section label as the leaf. Deep entries show `Destination › title` (or equivalent) with the entity ref trailing (e.g. `SRC-…`). |
| **Menu contents** | Back menu: up to the **nearest 15** entries **before** the current index (nearest first). Forward menu: up to the **nearest 15** **after** the current index, in stack order. Choosing a row jumps to that index and closes the menu. Full stack (up to 100) remains available via repeated Back/Forward / keyboard. |
| **Dismiss** | Click outside (or Escape) closes the menu without navigating. |
| **Breadcrumbs** | Flex-grow between nav buttons and omnibar (`PVBreadcrumbs`). Destination alone at the list root; deep location appends the short ref (e.g. `Sources` → `Sources › SRC-3K9M2`). **Not** a second Back control — remove the Source-page identity-header trail; this toolbar slot is the breadcrumb home. **Clickable ancestor segments navigate** via `go(to:)` (e.g. tap `Sources` from a Source page → Sources list location), which **pushes** (or coalesces) on the session stack like any other navigation — not a hierarchical Up that bypasses history, and not display-only. The current (leaf) crumb is non-navigating. |
| **Page title** | Remains in the **content** area (large display heading), not sandwiched between Back/Forward. |
| **Omnibar** | Trailing search field in the same toolbar (placeholder + `⌘K`). Owned by [`omnibar-search.md`](omnibar-search.md); listed here only as layout neighbor. |
| **Keyboard** | `⌘[` Back, `⌘]` Forward (board). Prefer these over arrow-key chords where they would fight text editing. |

## No separate in-page Back

Do **not** keep a second, hierarchical “up to parent list” control on the page that bypasses session history.

Today the Source page still hosts a **Sources → ref** breadcrumb in `SourcePageIdentityHeader` that calls `closeSource()` / pops to the list. **Remove it** from the Source page; the toolbar breadcrumbs above replace that slot. Leaving a deep location uses workspace **Back** / **Forward**, sidebar, omnibar, toolbar breadcrumbs once wired, or other `go(to:)` navigations — all of which update the same stack.

## Persistence (MVP)

History is **not** session-only. For the open project:

1. Persist the **stack**, the **current index**, and enough of each entry to restore (destination + deep location). Optional denormalized title for jump-menu rows is fine (refresh when the location is applied).
2. On relaunch (or reopening that project), restore the stack and **return to exactly where they left off**.
3. Scope: **per project** (and per window when multiple windows exist). Switching projects loads that project’s history, not another’s.
4. **Format / location:** JSON under Application Support — same family as `identity.json` / `active-project.json`. Path: `…/Provenencia/navigation/{projectUUID}.json` (lowercase hex of the catalog project UUID, no braces). Swift-owned `Codable`; atomic write (temp + rename). **Not** `UserDefaults`, **not** inside the `.provenencia` package tree as a loose chrome file, **not** rows in `provenencia.sqlite` for the history stack itself. History I/O needs no FFI once Swift knows the project UUID (from `ProjectInfo` / open).
5. **Deleted / missing entities** — if restore would open a Source (or row) that no longer exists, skip or prune that entry and land on a calm fallback (e.g. destination list) rather than a broken page.

**Stack length:** persist **100** entries max; jump menus show **15** nearest in that direction.

### Project key — catalog project UUID (infrastructure)

Do **not** key history by absolute path, basename, or bookmark. Key by a **durable project UUID** stored in the catalog so rename/move of the `.provenencia` folder keeps the same history on this install, and so other install-local chrome can reuse the same id later.

Today `project` is a singleton bookkeeping row (`id = 1`, label + timestamps) with **no** unique project identity. Spike 3 adds that identity:

| Piece | Requirement |
| --- | --- |
| **Column** | `project.uuid` — `BLOB` UUIDv7, **NOT NULL**, **UNIQUE** (after backfill). Same machine-id pattern as `users.id` ([`catalog-refs.md`](../../../catalog-refs.md) §1). No short `PRJ-…` ref required for this spike. |
| **Mint** | On **Create**, mint UUID when inserting the singleton row. |
| **Backfill** | On **Open** of an older catalog missing `uuid`, mint once and persist (heal path — same spirit as other Open-time backfills). |
| **Immutability** | Never rewrite `uuid` after mint. Copying the whole `.provenencia` folder copies the UUID (two working copies on one Mac share one history file — acceptable; do not mint a new id on copy). |
| **FFI / Swift** | Expose on `ProjectInfo` (or equivalent open payload) so the Mac client can open `navigation/{uuid}.json` without a second round trip. |
| **Migration** | New `core/database/migrations/NNNNNN.sql` via [`add-catalog-migration`](../../../../.cursor/skills/add-catalog-migration/SKILL.md); domain helpers in `core/database/project`. |
| **vs `active-project.json`** | Active project pointer stays **path-based** (which folder to open). History key is **catalog UUID** (which research document’s chrome state). |

Example history document:

```json
{
  "v": 1,
  "projectUuid": "019c0123-4567-7890-abcd-ef0123456789",
  "index": 2,
  "entries": [
    { "section": "sources" },
    {
      "section": "sources",
      "sourceId": "019c…",
      "ref": "SRC-3K9M2",
      "title": "Ilminster parish register, 1841–1852"
    },
    {
      "section": "source-fields",
      "fieldId": "019c…",
      "title": "Folio reference"
    }
  ]
}
```

(`projectUuid` inside the file is optional redundancy for support; the **filename** is authoritative for lookup.)

**Deep-target keys (per kind):** use explicit optional fields on the entry — `sourceId`, `fieldId`, `typeId`, and later peers as destinations gain deep locations. Ids are UUIDs for restore; `ref` / `title` are denormalized jump-menu cache (refresh when the location is applied). Do **not** collapse into a generic `entityKind` + `id`, and do **not** store path strings as the sole representation. Unknown keys from newer app versions should be ignored on decode.

## Implementation sketch

1. **Catalog project UUID** — migration + mint on Create + backfill on Open; surface on `ProjectInfo`.
2. **Location model** — `WorkspaceLocation` (destination + optional deep id / page kind). Single source of truth for “what is on screen.” Serialize as versioned JSON objects with **per-kind** deep fields (`sourceId`, `fieldId`, `typeId`, …), not URL strings alone and not a generic entity envelope.
3. **History store** — array + index on `WorkspaceModel` (or a dedicated type it owns); **load/save** `navigation/{projectUuid}.json` under Application Support.
4. **Navigate API** — `go(to:)`, `goBack()`, `goForward()`, `go(toIndex:)` for menu jumps; all sidebar/deep links call `go(to:)` so the stack stays honest.
5. **Restore** — applying a location sets sidebar selection and feature-model selection/page (`openedSourceID`, vocabulary selection, …); feature models must accept “select this id” from outside (already partly true for vocabulary). Do **not** introduce SwiftUI `NavigationStack` / `NavigationPath` for workspace session history — keep driving the existing section switch + destination swaps.
6. **Relaunch** — after project open (UUID known), apply persisted index location once models are ready (session traffic may overlap; do not reintroduce catalog-ready gating).
7. **Toolbar chrome** — hoist Back/Forward, breadcrumbs, and omnibar into the main-column header per the board; strip Source-page local breadcrumbs.

Omnibar and other cross-links become history-aware automatically once they use `go(to:)`.

## Why it fits Provenencia

Evidence work is link-shaped: Source ↔ (later) File ↔ Citation ↔ Person. Without history, every cross-link is a trap. Back/Forward is the cheap complement to omnibar search: search jumps you in; history gets you out — and persistence makes the workspace feel continuous across days.

## Resolved decisions

| Question | Decision |
| --- | --- |
| Sidebar destination switch always push? | **Yes.** |
| Master–detail row change pushes? | **Yes** (query-string style deep location). |
| Persist across relaunch? | **Yes** — stack + leave-off location; per project. |
| Persistence format? | **JSON in Application Support** (Option A); not UserDefaults / project package / history-in-SQLite. |
| Project key? | **Catalog `project.uuid` (UUIDv7)** — mint on Create, backfill on Open; history file `navigation/{uuid}.json`. Not path/basename/bookmark. |
| Stack size? | **Store 100**, jump menu shows **15** nearest; keyboard walks the full stored stack. |
| Coordinator? | **Hand-rolled** `WorkspaceLocation` + history store driving existing section/deep switches — **not** SwiftUI `NavigationStack` / `NavigationPath`. |
| Recent-entries long-press menu? | **Yes — MVP** (also secondary-click); see Chrome. |
| In-page / hierarchical Back? | **No** — remove Source-page trail; toolbar breadcrumbs + history. |
| Title between Back/Forward? | **No** — breadcrumbs + content page title (per board). |
| Toolbar breadcrumb clicks? | **Navigate via `go(to:)`** — ancestor segments push (or coalesce) on the session stack; leaf is display-only. |
| Deep-target JSON fields? | **Per-kind keys** (`sourceId`, `fieldId`, `typeId`, …) plus optional `ref` / `title` for menu chrome — not a generic `entityKind`+`id`, not path strings alone. |
| Short project `ref`? | **No** for this spike (and not required for history). Machine `project.uuid` only; no `PRJ-…`. |
| Implementation depth? | **First-class coordinator + catalog UUID + full wiring** — not a narrow toolbar band-aid; wide file touch OK. |
| Multiple windows | One history per window (when we have them). |

## Still open

_None for navigation history behavior._ Implementation order is locked in [`deployment-plan.md`](deployment-plan.md) (catalog UUID → history store → toolbar chrome; omnibar on the parallel search track).

## Explicitly out of scope

- Edit undo/redo, audit “revert,” or time-travel through catalog versions
- Browser-style tab strip / multiple concurrent location stacks per window
- Cross-project merged history
- Restoring list search queries or scroll offsets
- Short human project `ref` (`PRJ-…` or similar)

## Related docs

- [`deployment-plan.md`](deployment-plan.md) — sequenced PRs
- [`design/README.md`](design/README.md) — App Layout board + uploads
- [`omnibar-search.md`](omnibar-search.md) (jumps that should push history; toolbar neighbor)
- [`macos-client-patterns.md`](../../../macos-client-patterns.md)
- [`S2-01-workspace-chrome.md`](../spike-2/design/archive/S2-01-workspace-chrome.md)
- [`S2-04-sources-list.md`](../spike-2/design/archive/S2-04-sources-list.md) / [`S2-23-source-detail.md`](../spike-2/design/archive/S2-23-source-detail.md) (list ↔ Source page; historical local-back notes — superseded by this spike)
- [`S2-20-files-list.md`](../spike-2/design/archive/S2-20-files-list.md) (Source deep link — brief descoped with S2-21; still useful as a Files→Source jump sketch)
