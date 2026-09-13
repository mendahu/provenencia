repo: mendahu/provenencia
branch: docs/spike-3-deployment-plan
path: docs/deployment-plan/spike-3/design

## Last sync

date: 2026-09-13
source: `Main Application Layout.zip` (Claude Design export)

### Updated in this project

- App Layout: main-column toolbar with Back/Forward (hold / secondary-click jump menus), breadcrumbs, trailing omnibar field (`⌘K`).
- S3-01 Omnibar Results: flat ranked rich hit rows (`OmnibarHitRow` / Swift `PVOmnibarHitRow`), dropdown chrome, empty/loading/keyboard states.
- Omnibar Hit Row component board + supporting `_ds` / `support.js` / assets for local open.
- S2-01 Chrome Board retained from the same export for continuity with Spike 2 chrome.

## Screen map

| Screen | Built from |
| --- | --- |
| App Layout.dc.html | Spike 3 nav + omnibar field; evolves S2-01 workspace chrome ([`navigation-history.md`](../navigation-history.md), [`omnibar-search.md`](../omnibar-search.md)) |
| S3-01 Omnibar Results.dc.html | [`archive/omnibar-results-dropdown-brief.md`](archive/omnibar-results-dropdown-brief.md) |
| Omnibar Hit Row.dc.html | Shared row skeleton for S3-01 |
| S2-01 Chrome Board.dc.html | Historical Spike 2 chrome inventory (already shipped) |
