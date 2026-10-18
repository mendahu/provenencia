# Spike 12 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S12-02 — Copy Source form

**Landed 2026-10-18** in [#485](https://github.com/mendahu/provenencia/pull/485). A researcher can choose Copy… in a Source page's header (now Jump to graph, Copy, Delete) and get a form with the Add Source title and type fields, prefilled `Copy of <title>` and the original's type. Description, Metadata and Credibility are checked unless the original has none, and an Artifact checklist with All Artifacts lets them drop any Artifact. Saving makes the copy through `copySource` and opens its page, with no toast. A failed copy keeps every choice and a failed read of the original's parts offers Read again. Add Source, Copy Source and the Source page type editor now share one title and type form: errors show as you edit and Create is disabled while the form is invalid. Artifacts share one display title and fileless-icon rule, and `PVSkeletonBar` is lifted from the list skeleton. Copy from a Sources list row is S12-17.

## S12-01 — Copy Source engine

**Landed 2026-10-18** in [#478](https://github.com/mendahu/provenencia/pull/478). The engine can copy a Source in one audited `copy_source` write: a new Source with a chosen title and type and, each only when chosen, the original's description, metadata with its layout, credibility and any subset of its Artifacts. Notes, Subjects, Citations, Observations, graphs and the cover are never copied. `core/database/sourcecopy` reads a draft and creates from one, the slot registry import will reuse; the store gains `copySource`. No screen calls it yet; nothing changes for researchers.
