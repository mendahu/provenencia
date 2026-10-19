# Spike 12 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S12-03 — Evidence graphs in the catalog

**Landed 2026-10-19** in [#498](https://github.com/mendahu/provenencia/pull/498). The catalog (format 55) holds Evidence graphs: every Source has a Default graph with a `GPH-` ref, every Subject belongs to one graph of its Source, and Artifacts and graphs are linked many-to-many. Existing projects migrate in place, so each Source's Default graph holds its Subjects and is linked to all its Artifacts; the refs are minted when a project next opens, through one `catalogsession.ReconcileOnOpen` that create, `OpenCatalog` and `Do` share. `core/database/evidencegraphs` owns graphs and links (`Link`, `Unlink` with their refusals, `SoleGraphTx` until S12-08), and the graph list has its own `evidence_graphs` list key in `effects.Set.Lists` (with `vocabulary` and `property_usage`, one declared list set on the effects, the wire and Swift). No screen or RPC changes; nothing changes for researchers.

## S12-16 — Add Source reports its effects

**Landed 2026-10-19** in [#492](https://github.com/mendahu/provenencia/pull/492). `CreateSource` now returns `WriteEffects`, and Add Source applies them with `.createdSource` as Copy Source does, so the Sources list refreshes through the session and the hand-written list insert is gone. The session also recounts the sidebar after every write in `CatalogQueryRegistry.navCountTriggers` (Add, Copy, Source delete, Promote), which fixes the Sources count staying stale after Copy and Source delete; Promote's own recount went. Nothing else changes for researchers.

## S12-17 — Copy Source from the Sources list

**Landed 2026-10-18** in [#488](https://github.com/mendahu/provenencia/pull/488). A researcher can right-click a Sources list row's page zone, choose Copy…, and get the Copy Source form for that Source; saving opens the copy's page. The kit gains action menu items (`PVActionMenuItem`, `PVActionMenuRows`) and `pvRightClickMenu`, which adds a right-click menu's state, trigger and one VoiceOver action per item. The Source cover and Person portrait menus moved onto `pvRightClickMenu` (the cover now offers Revert to default to VoiceOver), and the Citation composer Clear menu onto the items layer.

## S12-02 — Copy Source form

**Landed 2026-10-18** in [#485](https://github.com/mendahu/provenencia/pull/485). A researcher can choose Copy… in a Source page's header (now Jump to graph, Copy, Delete) and get a form with the Add Source title and type fields, prefilled `Copy of <title>` and the original's type. Description, Metadata and Credibility are checked unless the original has none, and an Artifact checklist with All Artifacts lets them drop any Artifact. Saving makes the copy through `copySource` and opens its page, with no toast. A failed copy keeps every choice and a failed read of the original's parts offers Read again. Add Source, Copy Source and the Source page type editor now share one title and type form: errors show as you edit and Create is disabled while the form is invalid. Artifacts share one display title and fileless-icon rule, and `PVSkeletonBar` is lifted from the list skeleton. Copy from a Sources list row is S12-17.

## S12-01 — Copy Source engine

**Landed 2026-10-18** in [#478](https://github.com/mendahu/provenencia/pull/478). The engine can copy a Source in one audited `copy_source` write: a new Source with a chosen title and type and, each only when chosen, the original's description, metadata with its layout, credibility and any subset of its Artifacts. Notes, Subjects, Citations, Observations, graphs and the cover are never copied. `core/database/sourcecopy` reads a draft and creates from one, the slot registry import will reuse; the store gains `copySource`. No screen calls it yet; nothing changes for researchers.
