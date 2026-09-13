# Spike 3 — Claude Design boards

Visual source of truth for Spike 3 workspace layout and omnibar results. Behavior stays in [`../navigation-history.md`](../navigation-history.md) and [`../omnibar-search.md`](../omnibar-search.md). PR sequence: [`../deployment-plan.md`](../deployment-plan.md).

## Boards (implementation reference)

| Board | Path | Implements in |
| --- | --- | --- |
| App Layout (toolbar) | [`App Layout.dc.html`](App%20Layout.dc.html) | S3-05 (field shell); jump menus / breadcrumbs with S3-04–S3-05 |
| Omnibar results | [`S3-01 Omnibar Results.dc.html`](S3-01%20Omnibar%20Results.dc.html) | S3-10 |
| Omnibar hit row (component) | [`Omnibar Hit Row.dc.html`](Omnibar%20Hit%20Row.dc.html) | S3-10 → `PVOmnibarHitRow` |
| S2-01 chrome (historical) | [`S2-01 Chrome Board.dc.html`](S2-01%20Chrome%20Board.dc.html) | Already shipped as S2-14; included for continuity with the App Layout export |
| Screenshots | [`uploads/`](uploads/) | — |

Open the `.dc.html` files via the Claude Design / local support bundle (`support.js`, `_ds/`, `assets/`). Sync notes from the export: [`claude-design-sync.md`](claude-design-sync.md).

## Layout summary (App Layout)

Main column only (not over the sidebar):

```text
[ ← ] [ → ]   Breadcrumbs …                    [ 🔍  Search …          ⌘K ]
```

- Back/Forward: sm icon buttons; long-press or secondary-click → history jump menu.
- Breadcrumbs: flex-grow; clickable ancestors via `go(to:)`; replace Source-page local trail.
- Omnibar: ~420px / max 60% trailing field; results dropdown designed in S3-01 Omnibar Results.

## Omnibar results summary

- One shared rich row skeleton; kinds fill slots only (`PVOmnibarHitRow`).
- Flat ranked list (engine order); kind chip disambiguates.
- Dropdown widens left under the field (~640px); empty/short query does not open a panel.
- Leading tile: thumbnail → file-type glyph → kind icon (slot never collapses).

## Archived briefs

| Brief | Notes |
| --- | --- |
| [`archive/omnibar-results-dropdown-brief.md`](archive/omnibar-results-dropdown-brief.md) | Prompt used for S3-02; board supersedes it |
