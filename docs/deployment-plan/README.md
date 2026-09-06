# Deployment plan

Working notes for implementation milestones. These are stopping points, not a full product roadmap.

Authoritative domain and stack decisions remain in the sibling docs under [`docs/`](../).

## Current

| Spike | Goal |
| --- | --- |
| [Spike 2](spike-2.md) | Validate the Source layer: app workspace chrome (sidebar), audit + schema + Go CRUD/ingest + FFI + macOS Source catalog UI (create Sources, Artifacts, Files, extensible types/metadata). Design steps in Claude Design interleaved with PRs. |

## Completed

| Spike | Goal |
| --- | --- |
| [Spike 1](archive/spike-1.md) | Scaffold the macOS app, local SQLite project, and first-run onboarding. **Retired the cgo SQLite + Swift dylib risk** (plan A: `mattn/go-sqlite3` inside `libprovenencia.dylib`). |

Older milestone notes live in [`archive/`](archive/).
