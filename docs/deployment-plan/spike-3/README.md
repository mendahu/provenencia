# Spike 3 — Workspace layout (nav history + omnibar)

## Status

**Sequenced.** Design boards checked in; PR order lives in [`deployment-plan.md`](deployment-plan.md).

Workspace layout (nav history + omnibar chrome) plus first-class infrastructure: **catalog `project.uuid`**, hand-rolled navigation history, and **Go/SQLite catalog search** (registry + FTS5 + ranking; successive PRs).

Authoritative stack / chrome context: [`macos-client-patterns.md`](../../macos-client-patterns.md), [`application-stack.md`](../../application-stack.md). Spike 2 chrome brief (historical): [`S2-01-workspace-chrome.md`](../archive/spike-2/design/archive/S2-01-workspace-chrome.md).

**Design:** [`design/App Layout.dc.html`](design/App%20Layout.dc.html) — main-column toolbar with Back/Forward, breadcrumbs, and omnibar; [`design/S3-01 Omnibar Results.dc.html`](design/S3-01%20Omnibar%20Results.dc.html) — results dropdown ([`design/README.md`](design/README.md)).

## Goal

First-class **Back/Forward** (persisted, coordinator-driven) and **project search** (toolbar omnibar + engine-side FTS/ranking), not band-aid chrome. Early-dev churn and broad file touch are acceptable; successive PRs should still be dogfoodable.

See [`navigation-history.md`](navigation-history.md) § Implementation posture and [`omnibar-search.md`](omnibar-search.md) § Implementation posture / Incremental delivery.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Sequenced Design + PR steps (S3-01…), layout deltas, dogfood bar |
| [Navigation history](navigation-history.md) | Back/Forward behavior, persistence, `project.uuid` |
| [Omnibar search](omnibar-search.md) | Registry + FTS5 + ranking; remove per-destination search |
| [Design boards](design/README.md) | App Layout + Omnibar Results exports |

## Out of scope (for now)

- Interpretation / Conclusion catalog work (search kinds optional later as S3-13+)
- Files list destination (descoped in Spike 2)
- Catalog access serialization / DB performance ([archived idea](../../ideas/archive/catalog-access-serialization.md))
- Short human project `ref` (unless a later spike needs one)
