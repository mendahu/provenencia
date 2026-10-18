# Spike 12 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S12-01 — Copy Source engine

**Landed 2026-10-18** in [#478](https://github.com/mendahu/provenencia/pull/478). The engine can copy a Source in one audited `copy_source` write: a new Source with a chosen title and type and, each only when chosen, the original's description, metadata with its layout, credibility and any subset of its Artifacts. Notes, Subjects, Citations, Observations, graphs and the cover are never copied. `core/database/sourcecopy` reads a draft and creates from one, the slot registry import will reuse; the store gains `copySource`. No screen calls it yet; nothing changes for researchers.
