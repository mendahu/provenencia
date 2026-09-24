# Deployment plan

Working notes for implementation milestones. These are stopping points, not a full product roadmap.

Authoritative domain and stack decisions remain in the sibling docs under [`docs/`](../). Unscheduled product ideas live in [`ideas/`](../ideas/) — a parking lot, not a spike queue. UX friction from dogfooding lives in [`docs/dogfood/ux.md`](../dogfood/ux.md) until something is pulled into a spike.

## Current

| Spike | Goal |
| --- | --- |
| [Spike 8](spike-8/) | **Pause and refine** data entry. Stories: image **Auto Transcribe**; **PDF Artifact thumbnails**; PDF **Find** + **select** + **paste transcription**; graph **visual enhancements**; Source page **Evidence graph jump**; Sources list **graph-progress counts**; **delete paths**. More entry-flow stories will join. Dogfood: [`docs/dogfood/ux.md`](../dogfood/ux.md). Plan: [`spike-8/deployment-plan.md`](spike-8/deployment-plan.md). |

## Completed

| Spike | Goal |
| --- | --- |
| [Spike 7](archive/spike-7/) | Citations / Observations / NameValue / Subject **fields** / citation **composer place** (Option B) / durable connect. Subject types stay seeded (no types editor). Design: [`ideas/archive/interpretation-graph-ui.md`](../ideas/archive/interpretation-graph-ui.md). Archive: [`archive/spike-7/README.md`](archive/spike-7/README.md), [`completed.md`](archive/spike-7/completed.md). |
| [Spike 6](archive/spike-6/) | Evidence graph **canvas prototype**: pan/zoom, click-to-add bubbles, drag/snap, relationship lines, accessibility. **Go** — UI risk retired; provisional connect honesty labeled. Neutral `GraphCanvas` reuse contract. Design: [`ideas/archive/interpretation-graph-ui.md`](../ideas/archive/interpretation-graph-ui.md). Archive: [`archive/spike-6/README.md`](archive/spike-6/README.md), [`completed.md`](archive/spike-6/completed.md). |
| [Spike 5](archive/spike-5/) | Interpretation foundation up to the canvas: candidate refs, `subject_types` + seed, audited `subjects`, layout storage, FFI, nested Sources config (Subject types / fields stubs), Sources list → **Evidence graph** stub, floating-menu unify (S5-10). **No Subject UI** — canvas is Spike 6. Design: [`ideas/archive/interpretation-graph-ui.md`](../ideas/archive/interpretation-graph-ui.md). Archive: [`archive/spike-5/README.md`](archive/spike-5/README.md), [`completed.md`](archive/spike-5/completed.md). |
| [Spike 4](archive/spike-4/) | Workspace session + catalog query cache + declarative place registry: fast, consistent navigation; Sources/fields/types on query handles; `add-workspace-place` skill. Design: [`ideas/archive/page-navigation-performance.md`](../ideas/archive/page-navigation-performance.md). Archive: [`archive/spike-4/README.md`](archive/spike-4/README.md), [`completed.md`](archive/spike-4/completed.md). |
| [Spike 3](archive/spike-3/) | Workspace chrome + first-class nav history and catalog search: Back/Forward (persisted, `project.uuid`), toolbar omnibar with Go registry + FTS5 ranking. Archive: [`archive/spike-3/README.md`](archive/spike-3/README.md), [`completed.md`](archive/spike-3/completed.md). Notes: [`navigation-history.md`](archive/spike-3/navigation-history.md), [`omnibar-search.md`](archive/spike-3/omnibar-search.md). |
| [Spike 2](archive/spike-2/) | Validate the Source layer: app workspace chrome (sidebar), audit + schema + Go CRUD/ingest + FFI + macOS Source catalog UI (create Sources, Artifacts, Files, extensible types/metadata). Design steps in Claude Design interleaved with PRs. Dogfood: [`archive/spike-2/dogfood.md`](archive/spike-2/dogfood.md). |
| [Spike 1](archive/spike-1.md) | Scaffold the macOS app, local SQLite project, and first-run onboarding. **Retired the cgo SQLite + Swift dylib risk** (plan A: `mattn/go-sqlite3` inside `libprovenencia.dylib`). |

Spike 8 (open):

- [`spike-8/README.md`](spike-8/README.md) — spike overview
- [`spike-8/deployment-plan.md`](spike-8/deployment-plan.md) — PR sequence, design gates, dogfood bar
- [`spike-8/completed.md`](spike-8/completed.md) — finished steps
- [`spike-8/design/`](spike-8/design/) — Claude Design briefs (**S8-D1**…**S8-D6** open)
- [`spike-8/artifact-pdf-thumbnails.md`](spike-8/artifact-pdf-thumbnails.md) — scoped **S8-02** note
- [`spike-8/pdf-text-find.md`](spike-8/pdf-text-find.md) — scoped **S8-D2** / **S8-03…S8-05** note

Spike 7 archive:

- [`archive/spike-7/README.md`](archive/spike-7/README.md) — spike overview (no open steps)
- [`archive/spike-7/completed.md`](archive/spike-7/completed.md) — finished Spike 7 steps (S7-01…S7-15 / S7-11)
- [`archive/spike-7/deployment-plan.md`](archive/spike-7/deployment-plan.md) — PR sequence, design gates, dogfood bar
- [`archive/spike-7/design/`](archive/spike-7/design/) — Claude Design briefs (S7-D2…S7-D10); archived briefs in [`archive/spike-7/design/archive/`](archive/spike-7/design/archive/)

Spike 6 archive:

- [`archive/spike-6/README.md`](archive/spike-6/README.md) — spike overview (no open steps); **Go**
- [`archive/spike-6/completed.md`](archive/spike-6/completed.md) — finished Spike 6 steps (S6-D1…S6-05)
- [`archive/spike-6/deployment-plan.md`](archive/spike-6/deployment-plan.md) — PR sequence, design gates, dogfood bar, go/no-go
- [`archive/spike-6/design/`](archive/spike-6/design/) — Claude Design briefs (S6-D1, S6-D2); archived briefs in [`archive/spike-6/design/archive/`](archive/spike-6/design/archive/)

Spike 5 archive:

- [`archive/spike-5/README.md`](archive/spike-5/README.md) — spike overview (no open steps)
- [`archive/spike-5/completed.md`](archive/spike-5/completed.md) — finished Spike 5 steps (S5-01…S5-10)
- [`archive/spike-5/deployment-plan.md`](archive/spike-5/deployment-plan.md) — PR sequence, design gates, definition of done
- [`archive/spike-5/design/`](archive/spike-5/design/) — design summaries; archived briefs in [`archive/spike-5/design/archive/`](archive/spike-5/design/archive/)

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
