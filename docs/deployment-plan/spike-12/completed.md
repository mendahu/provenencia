# Spike 12 — Completed

One short paragraph per landed PR, newest first. Written by the PR that lands it.

## S12-13 — Artifact graph checkboxes

**Landed 2026-10-25** in [#562](https://github.com/mendahu/provenencia/pull/562). On a Source with two or more Evidence graphs, Add artifact lists the graphs as checkboxes, all checked, and links the new Artifact only to the checked ones (Add waits until one is checked). An Artifact card in edit lists the graphs checked for its current links; one Save writes the label, description and link changes, links before unlinks. A graph that must keep the Artifact (its Citations use it, or it is the graph's only Artifact) shows locked with the reason, and a refused save shows on the card and re-reads the locks. The canvas's cite tools and Files menu now follow the graph's own links, and its callout and the composer's name the graph. The session indexes the graph list by id and Source (`evidenceGraphsBySource`), which replaces the Sources list's own cache; `PVArtifactChecklistState` is `PVChecklistState`, and `SourceFormLoadArea` gains a compact size. Swift only.

## S12-12 — Graph form

**Landed 2026-10-25** in [#557](https://github.com/mendahu/provenencia/pull/557). A researcher can click Add evidence graph in the Source page's Evidence graphs header, or Edit on a graph's row, and get one form: name, description and the Source's Artifacts as a checklist, all checked for a new graph and the linked ones for an edit. Saving makes or changes the graph in one write; Add toasts "Evidence graph added" and the list stays in name order. An edit reads once which Artifacts the graph's Citations use, shows them checked and locked with the Citations listed, and offers Read again if that read fails. Add is disabled with a reason on a Source with no Artifacts. The last-graph delete notice now names Add evidence graph. The Artifact checklist is promoted to the `PVArtifactChecklist` recipe (shared with Copy Source), `PVCheckbox` gains a lock mark, and the delete notice's Citation list and the forms' load area are shared.

## S12-11 — Source page Graphs section

**Landed 2026-10-25** in [#549](https://github.com/mendahu/provenencia/pull/549). The Source page shows an Evidence graphs section between the overview columns and Artifacts: one card per graph with its name, `GPH-` ref and description, an Open button (disabled with the reason when the graph links no Artifact) and a Delete button that goes through the delete-impact check, with a success toast. The header's Jump to graph is gone. The canvas, Citation composer and Promote breadcrumbs read Sources › Source title › graph name from the session's lists, so a rename shows at once. `WorkspaceLocation.sourcePage` replaces three hand-built copies, and `WorkspaceSession.evidenceGraph(id:)`, `evidenceGraphs(sourceID:)` and `source(id:)` are the one read of each list. Swift only; no engine, FFI or catalog change.

## S12-10 — Composer offers linked Artifacts

**Landed 2026-10-24** in [#540](https://github.com/mendahu/provenencia/pull/540). The Citation composer on one Evidence graph offers only the Artifacts linked to that graph, and on each steps only through the graph's Citations (those with an Observation on one of its Subjects, plus those with none yet), with that graph's counts and names. `ListCitationsByArtifact` takes `graph_id` and `CitationCountsBySource` became `CitationCountsByEvidenceGraph`; the app's citation keys are per graph and link rows now report `CitationSource`, so a link, unlink or graph edit refreshes them. The empty callout reads "This evidence graph has no artifacts". On an existing project nothing visibly changes: every Source has one graph linked to all its Artifacts.

## S12-09 — Promote per graph

**Landed 2026-10-24** in [#535](https://github.com/mendahu/provenencia/pull/535). Promote works on one Evidence graph: the proposal, the claim evidence sheet, Done (claims, pins and the bridges it files) and Unpromote's bridge unfiling each read or write the Subjects of one graph, and the three Promote requests carry `graph_id` in place of `source_id` (old numbers reserved). On a Source with two graphs a Done files only its own graph's Subjects, and a waiting bridge on the other graph stays unfiled until that graph's own Done. `FileSourceBridgesTx` is now `FileEvidenceGraphBridgesTx`, filing and unfiling share one `bridgeHeads` query, `UnfileBridges` takes only the Subject, and `evidencegraphs.SoleGraph` and `SoleGraphTx` are deleted. The Promote page sends the graph id it already held; the researcher sees no change.

**Not this PR.** A Promote entry point per graph on the Source page or Sources list (S12-11, S12-14); graph names in Promote breadcrumbs (S12-11); layer memoization for large Sources (S11-21).

## S12-18 — A Subject needs its graph

**Landed 2026-10-24** in [#532](https://github.com/mendahu/provenencia/pull/532). `subjects.CreateInput` loses `SourceID` and `GraphID` is required: a new Subject names the Evidence graph it joins and the engine reads its Source from that graph, so a second graph on a Source can never send a Subject to the wrong one. `subjects.InsertTx`'s `SoleGraphTx` fallback is deleted, and the cited bridge takes only its endpoints' graph. Every test that built a Subject now passes a graph id, and the copied test helpers that create one Subject share `subjectstest.On` / `OfType`. The researcher sees no change.

**Not this PR.** Promote's `SoleGraph` use and deleting `SoleGraph` / `SoleGraphTx` (S12-09); placing a test Subject in the same write as its create ([`test-subject-fixtures.md`](../../ideas/test-subject-fixtures.md)).

## S12-08 — Graph id end to end

**Landed 2026-10-23** in [#528](https://github.com/mendahu/provenencia/pull/528). The Evidence graph canvas, the Citation composer and the Promote page open on an Evidence graph id, and the id travels the whole way: `WorkspaceLocation.graphId` (kept by Back, Forward and a relaunch; an old history entry for a graph surface with no graph opens the Source page), the app's `.evidenceGraph(graphId:)` cache key, every graph read and write RPC (renamed where they named a Source), and the write effects (`WriteEffects.evidence_graphs` replaces `sources`). A write on one graph reloads only that graph and patches only its progress; a Subject, an Observation and a claim name their own graph, and an Artifact or Citation write still names every graph of its Source. Delete-impact rows and reconciliation records link to the graph that holds them. Until per-graph Open (S12-11, S12-14) the Source page jump and the Sources list Open open the Source's first graph in the graph list; every Source has one graph today, so the researcher sees no change.

**Not this PR.** Promote's RPCs and `SoleGraph` (S12-09); a required `subjects.CreateInput.GraphID` (S12-18); the composer's linked Artifacts (S12-10); graph names in breadcrumbs (S12-11).

## S12-07 — Graph-keyed reads

**Landed 2026-10-22** in [#520](https://github.com/mendahu/provenencia/pull/520). The engine's Evidence graph cache, loader, Subject positions, event titles and graph progress read one Evidence graph instead of one Source, and are named for it (`graphcache.EvidenceGraph`, `graphfill`, `grapheventtitles`, `graphprogress`). Two graphs of one Source load, memoize and count apart. A write that names a Source drops every loaded graph of it, and a graph create, edit or delete now names its Source. The wire and the app do not change: each handler that takes a Source id reads that Source's one graph.

## S12-06 — Graph isolation

**Landed 2026-10-21** in [#515](https://github.com/mendahu/provenencia/pull/515). The engine now refuses an Observation whose Subject, Subject value or Citation Artifact is not in, or linked to, one Evidence graph, with its own plain message ("That subject or artifact is in another evidence graph."); a `source` Subject may still be the value of a non-edge Observation in any graph, so `mentions` crosses Sources. A cited bridge takes its graph from its endpoints, and `subjects.CreateInput.GraphID` lets a Subject join a named graph. Nothing changes for researchers until a second graph can be made.

## S12-05 — Graph RPCs and graph list

**Landed 2026-10-21** in [#511](https://github.com/mendahu/provenencia/pull/511). The app can reach every Evidence graph operation the engine has: list every graph of the project, create, edit and delete a graph, link and unlink an Artifact, and create an Artifact linked to chosen graphs (`graph_ids`; none means every graph of its Source). The graph list is one cached read, `evidenceGraphsList`, that any write reporting the `evidence_graphs` list refreshes. `GetEvidenceGraphUnlinkImpact` names, for each Artifact an edit would unlink, the graph's Citations that use it, through the new `deleteimpact.ListedGroup`. The five `evidencegraphs.*` codes have their own copy, the `last_graph` delete gate reaches Swift, and a graph delete preview reads "This graph has N subjects". No view reads these yet; nothing changes for researchers.

## S12-04 — Graph lifecycle

**Landed 2026-10-20** in [#505](https://github.com/mendahu/provenencia/pull/505). The engine can list, create, edit and delete Evidence graphs. `evidencegraphs.List` is the one read of graph rows (full rows with their linked Artifact ids, by name, for a Source or the project); `Create` links a new graph to chosen Artifacts; `Update` saves name, description and linked Artifacts in one write. `Delete` goes through `deleteimpact`: a graph with Subjects is blocked by the inbound probe and a Source's last graph by a new `last_graph` extra gate, which joins `edge_locked` and `origin_locked` in one per-kind `extraGateList`. No RPC or screen calls these yet; nothing changes for researchers.

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
