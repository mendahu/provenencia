# Spike 3 — Design

Claude Design boards for Spike 3 live in the Provenencia Claude Design project — **not** checked into git.

| Board | Implements in |
| --- | --- |
| App Layout (toolbar: Back/Forward, breadcrumbs, omnibar field) | S3-05 (done) |
| Omnibar results (flat ranked rich rows) | S3-10 (done) |
| Omnibar hit row component | S3-10 → `PVOmnibarHitRow` (done) |

Behavior: [`../navigation-history.md`](../navigation-history.md), [`../omnibar-search.md`](../omnibar-search.md). PR sequence: [`../deployment-plan.md`](../deployment-plan.md).

## Layout summary (App Layout)

Main column only (not over the sidebar):

```text
[ ← ] [ → ]   Breadcrumbs …                    [ 🔍  Search …          ⌘K ]
```

- Back/Forward: sm icon buttons; long-press or secondary-click → history jump menu.
- Breadcrumbs: flex-grow; clickable ancestors via `go(to:)`; replace Source-page local trail.
- Omnibar: ~420px / max 60% trailing field; results overlay under the field (S3-10).

## Omnibar results summary

- One shared rich row skeleton; kinds fill slots only (`PVOmnibarHitRow`).
- Flat ranked list (engine order); kind chip disambiguates.
- Dropdown widens left under the field (~640px); empty/short query does not open a panel.
- Leading tile: thumbnail → file-type glyph → kind icon (slot never collapses).
- Surface: same in-hierarchy overlay family as history jump menus (`PVRadius.sm`, content-column overlay) — not an Apple floating window / `NSPanel` / `.popover`.

## Archived briefs

| Brief | Notes |
| --- | --- |
| [`archive/omnibar-results-dropdown-brief.md`](archive/omnibar-results-dropdown-brief.md) | Prompt used for the Omnibar Results board |
