# Deployment plan

Working notes for implementation milestones. These are stopping points, not a full product roadmap.

Authoritative domain and stack decisions remain in the sibling docs under [`docs/`](../).

## Current

| Spike | Goal |
| --- | --- |
| [Spike 2](spike-2/) | Validate the Source layer: app workspace chrome (sidebar), audit + schema + Go CRUD/ingest + FFI + macOS Source catalog UI (create Sources, Artifacts, Files, extensible types/metadata). Design steps in Claude Design interleaved with PRs. |

Spike 2 contents:

- [`spike-2/README.md`](spike-2/README.md) — task list (S2-01…S2-19)
- [`spike-2/design/`](spike-2/design/) — Claude Design requirement briefs for S2-01…S2-04

## Completed

| Spike | Goal |
| --- | --- |
| [Spike 1](archive/spike-1.md) | Scaffold the macOS app, local SQLite project, and first-run onboarding. **Retired the cgo SQLite + Swift dylib risk** (plan A: `mattn/go-sqlite3` inside `libprovenencia.dylib`). |

Older milestone notes live in [`archive/`](archive/).
