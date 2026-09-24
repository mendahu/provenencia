# Deployment plan

Working notes for implementation milestones. These are stopping points, not a full product roadmap.

Authoritative domain and stack decisions remain in the sibling docs under [`docs/`](../). Unscheduled product ideas live in [`ideas/`](../ideas/) — a parking lot, not a spike queue. UX friction from dogfooding lives in [`docs/dogfood/ux.md`](../dogfood/ux.md) until something is pulled into a spike.

Closed spikes under [`archive/`](archive/) keep **themes and decisions**. PR order, checklists, and Claude Design briefs are git history — not duplicated here.

## Current

| Spike | Goal |
| --- | --- |
| [Spike 8](spike-8/) | **Pause and refine** data entry. Composer rethink first; then Auto Transcribe, PDF thumbs / Find / paste, graph chrome, Source-page jump, list counts, delete paths. Dogfood: [`docs/dogfood/ux.md`](../dogfood/ux.md). Plan: [`spike-8/deployment-plan.md`](spike-8/deployment-plan.md). |

## Completed

| Spike | What we decided |
| --- | --- |
| [Spike 7](archive/spike-7/) | Citation → Observation pipeline. Composer is a **place** (Option B). Subject types stay seeded. |
| [Spike 6](archive/spike-6/) | Spatial Evidence graph is **Go**. Neutral `GraphCanvas`. Connect was provisional until Spike 7. |
| [Spike 5](archive/spike-5/) | Interpretation data up to a graph stub. **No Subject UI.** One Sources family. |
| [Spike 4](archive/spike-4/) | Session cache + place registry. Go read-model tiering descoped. |
| [Spike 3](archive/spike-3/) | Persisted Back/Forward + omnibar (FTS5). Contracts: [`navigation-history.md`](archive/spike-3/navigation-history.md), [`omnibar-search.md`](archive/spike-3/omnibar-search.md). |
| [Spike 2](archive/spike-2/) | Source catalog + workspace chrome. Audit on first mutation. Files list descoped. |
| [Spike 1](archive/spike-1.md) | Local `*.provenencia` + install identity. cgo-SQLite-in-dylib risk retired. |

Spike 8 (open):

- [`spike-8/README.md`](spike-8/README.md)
- [`spike-8/deployment-plan.md`](spike-8/deployment-plan.md)
- [`spike-8/design/`](spike-8/design/) — Claude Design briefs. Working rules: [`claude-design-working-rules.md`](spike-8/design/claude-design-working-rules.md)
