# Deployment plan

Working notes for implementation milestones. These are stopping points, not a full product roadmap.

Authoritative domain and stack decisions remain in the sibling docs under [`docs/`](../). Unscheduled product ideas live in [`ideas/`](../ideas/) — a parking lot, not a spike queue.

## Current

_No active spike._ Next milestones will be added here when scheduled.

## Completed

| Spike | Goal |
| --- | --- |
| [Spike 4](archive/spike-4/) | Workspace session + catalog query cache + declarative place registry: fast, consistent navigation; Sources/fields/types on query handles; `add-workspace-place` skill. Design: [`ideas/archive/page-navigation-performance.md`](../ideas/archive/page-navigation-performance.md). Archive: [`archive/spike-4/README.md`](archive/spike-4/README.md), [`completed.md`](archive/spike-4/completed.md). |
| [Spike 3](archive/spike-3/) | Workspace chrome + first-class nav history and catalog search: Back/Forward (persisted, `project.uuid`), toolbar omnibar with Go registry + FTS5 ranking. Archive: [`archive/spike-3/README.md`](archive/spike-3/README.md), [`completed.md`](archive/spike-3/completed.md). Notes: [`navigation-history.md`](archive/spike-3/navigation-history.md), [`omnibar-search.md`](archive/spike-3/omnibar-search.md). |
| [Spike 2](archive/spike-2/) | Validate the Source layer: app workspace chrome (sidebar), audit + schema + Go CRUD/ingest + FFI + macOS Source catalog UI (create Sources, Artifacts, Files, extensible types/metadata). Design steps in Claude Design interleaved with PRs. Dogfood: [`archive/spike-2/dogfood.md`](archive/spike-2/dogfood.md). |
| [Spike 1](archive/spike-1.md) | Scaffold the macOS app, local SQLite project, and first-run onboarding. **Retired the cgo SQLite + Swift dylib risk** (plan A: `mattn/go-sqlite3` inside `libprovenencia.dylib`). |

Spike 4 archive:

- [`archive/spike-4/README.md`](archive/spike-4/README.md) — spike overview (no open steps)
- [`archive/spike-4/completed.md`](archive/spike-4/completed.md) — finished Spike 4 steps (S4-01…S4-09)
- [`archive/spike-4/deployment-plan.md`](archive/spike-4/deployment-plan.md) — PR sequence, cache strategy, descoped S4-10+ notes

Spike 3 archive:

- [`archive/spike-3/README.md`](archive/spike-3/README.md) — spike overview (no open steps)
- [`archive/spike-3/completed.md`](archive/spike-3/completed.md) — finished Spike 3 steps
- [`archive/spike-3/navigation-history.md`](archive/spike-3/navigation-history.md), [`archive/spike-3/omnibar-search.md`](archive/spike-3/omnibar-search.md)
- [`archive/spike-3/design/`](archive/spike-3/design/) — design summaries; archived brief in [`archive/spike-3/design/archive/`](archive/spike-3/design/archive/)

Spike 2 archive:

- [`archive/spike-2/README.md`](archive/spike-2/README.md) — spike overview (no open steps)
- [`archive/spike-2/completed.md`](archive/spike-2/completed.md) — finished Spike 2 steps
- [`archive/spike-2/dogfood.md`](archive/spike-2/dogfood.md) — how to dogfood the Source catalog
- [`archive/spike-2/design/`](archive/spike-2/design/) — Claude Design briefs; completed briefs in [`archive/spike-2/design/archive/`](archive/spike-2/design/archive/)

Older milestone notes live in [`archive/`](archive/).
