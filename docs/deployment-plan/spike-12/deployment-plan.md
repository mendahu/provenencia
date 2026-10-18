# Deployment Plan — Spike 12

A Source can be copied to start a sibling one in a single step, and a big Source can hold several isolated Evidence graphs that are each promoted on their own. The design is in [`source-structure.md`](source-structure.md). It pulls in [`ideas/source-structure.md`](source-structure.md) (moved here) and settles it as two tools; the options weighed and parked are in [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md) and [`ideas/graph-focus-views.md`](../../ideas/graph-focus-views.md). One behavior, one PR.

**A brief covers one view.** A PR that changes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

**Spike 11 runs in parallel.** It is rebuilding Promote's matching and its loaders. The PRs here that touch the same code name the Spike 11 PRs they wait on in **Depends on**. A PR plan written before those land must say so and stop at `Blocked — on S11-NN`.

Map review: [`deployment-plan.review.md`](deployment-plan.review.md).

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a big Source can be split into isolated, separately promoted Evidence graphs, and a Source can be copied to start a sibling one in a single step.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **Copy Source.** From an obituary, **Copy…** (on the Source page or a Sources list row's context menu) opens a form with the title `Copy of …`, a type dropdown preselected to the obituary's type, and checkboxes, all on, for description, metadata with its layout, credibility, and each Artifact. Saving opens the new Source with a Default graph, the copied Artifacts and no cover chosen.
2. **A census with three graphs.** Each graph has its own `GPH-` ref, name, description and linked page scans. In the Jones graph the Citation composer offers only the Jones page.
3. **Isolation holds.** A relationship to a Subject in another graph is refused. Citing an Artifact not linked to the graph is refused. A Subject never moves between graphs.
4. **Promote per graph.** Promoting the Smith graph files only Smith Subjects. Two graphs of one census agreeing on a value count as one Source for the auto-reconciler.
5. **Finding graphs.** Every Sources list row expands to list its graphs alphabetically, each with its own progress and Open link. The Source page has a Graphs section with an Open button per graph; the header's graph jump is gone.
6. **Existing projects look unchanged.** Every Source has one graph named Default holding its Subjects and linked to all its Artifacts; the canvas looks as before.
7. **Navigation.** The omnibar finds a graph by name, description or `GPH-` ref. Back / Forward and restore return to the exact graph, its composer and its Promote page.
8. **Link rules.** A new graph needs an Artifact and starts with all of them checked. A new Artifact starts linked to every graph. Unlinking an Artifact the graph's Citations use, or the graph's last Artifact, is refused with the reason. A graph with Subjects, or a Source's last graph, can't be deleted.

## Where the notes went

| Idea or note | PR | Brief |
| --- | --- | --- |
| [`source-structure.md`](source-structure.md) (was `ideas/source-structure.md`) — Problem 2, nested Sources and the "obituary or newspaper?" confusion | S12-01, S12-02, S12-17 (Copy Source) | S12-D1, S12-D6 |
| Same — Problem 1, several Evidence graphs per Source | S12-03 – S12-15 | S12-D2 – S12-D5 |
| Same — graph ↔ Artifact association | S12-03, S12-04, S12-05, S12-10, S12-12, S12-13 | S12-D3, S12-D4 |
| [S12-01 plan review](prs/S12-01-copy-source-engine.review.md) M-9 — Add Source and Copy Source refresh the Sources list two ways | S12-16 | — |
| [`dogfood/ux.md`](../../dogfood/ux.md) — "Dense Evidence graphs (census-scale)" (2026-09-24, descoped from Spike 10): split a census so the household you care about is on its own graph | S12-03 – S12-15 | S12-D2 – S12-D5 |

**Back in the ideas folder** — [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md): nested Sources, derivation links, named templates, per-graph credibility, layered citation output. [`ideas/graph-focus-views.md`](../../ideas/graph-focus-views.md): focus views for one large graph, including the "layers or filters" half of the dense-graphs dogfood note.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Spike-wide

- **Two tools, one rule for choosing.** A part that can be cited on its own (its own author, title or date) is its own Source, started with Copy Source; a region of one work is another graph ([the two tools](source-structure.md#the-two-tools)). Copy and docs wording follow this rule.
- **High tolerance for churn.** No one but the product owner uses the app yet. No adapters, shims or kept-for-now paths beyond the one bridge below. A PR that moves a caller from Source to graph moves every caller and deletes the Source-keyed version; it **replaces** Source-keyed fields (effects, query keys, wire fields) instead of adding graph ones beside them.
- **One bridge, one function, deleted in S12-08.** Between S12-03 and S12-08 every Source has exactly one graph (creating a second is not exposed until S12-11 / S12-12). `evidencegraphs.SoleGraphTx(sourceID)` is the only place that resolves "the Source's graph"; every interim caller uses it. S12-08 deletes it, and the compiler lists anything left (except Promote's, which S12-09 removes).
- **A re-keyed symbol is renamed in the PR that re-keys it.** One name for the concept: `EvidenceGraph` in Go and Swift (`graphcache` `EvidenceGraph(id)`, `CatalogQueryKey.evidenceGraph(graphId:)`, `graphprogress`), "Evidence graph" in copy. Bare "graph" is already taken by `graphcache.Graph`, `canonicalgraph` and `graphalign`. Binds S12-07, S12-08, S12-09.
- **Each PR updates the docs and skills it makes false.**
  - S12-01: [`source-layer-data-model.md`](../../source-layer-data-model.md) (Copy Source).
  - S12-03: `source-layer-data-model.md` and [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) (graphs, a Subject's graph); [`catalog-refs.md`](../../catalog-refs.md) and the [`add-catalog-ref`](../../../.agents/skills/add-catalog-ref/SKILL.md) prefix table (`GPH`); `core/database/effects/README.md`; [`catalog-deletes.md`](../../catalog-deletes.md).
  - S12-06: `interpretation-layer-data-model.md` (isolation).
  - S12-08: [`macos-client-patterns.md`](../../macos-client-patterns.md), [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md), [`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md) (graph-keyed effects and keys).
  - S12-09: [`promote-graph-alignment.md`](../../promote-graph-alignment.md) ("a layer is one Evidence graph's Subjects"), [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) (Unpromote unfiles the graph's bridges).

### Copy Source

- **The form** ([design](source-structure.md#copy-source)): title field prefilled `Copy of <title>`; required type dropdown preselected to the original's; checkboxes all on for description, metadata (values and layout), credibility (grade and argument), and one per Artifact. Notes are not offered. Graphs, Subjects, Citations and Observations are never copied. The copy's cover is reset (`primary_artifact_id` empty). Binds S12-01, S12-02.
- **Copy lives in its own composing package, `core/database/sourcecopy`.** `sources` can't import `sourcemetadata` (which imports `sources`), so, like `connect.CreateCitedBridge`, a composing package owns the write. It has two steps: read a **draft** from an existing Source (type, title, description, metadata rows and order, credibility, Artifacts), and **create a Source from a draft** through `sources.Create` (so the copy gets its Default graph from S12-03 onward) and `artifacts.Create` (same label, description and `file_id`). Copy is read-draft, apply the form's choices, create. One `writes.Run` transaction, action `copy_source`. The create-from-draft step is the slot [`ideas/registry-source-import.md`](../../ideas/registry-source-import.md) will reuse. Binds S12-01.
- **Entry points:** the Source page (S12-02) and a Sources list row's context menu (S12-17). One brief per view: S12-D1 (Source page and the form), S12-D6 (Sources list). Binds S12-02, S12-17.
- **One title and type form, one Artifact checklist.** S12-02 reuses the Add Source form's title and type fields and validation (`SourcesListView.addForm`, `SourcesModel.Draft`) instead of building a second form, and does not copy `create()`'s `session.setQueryValue` write-through. The per-Artifact checklist is built once in S12-02 and reused by S12-12 (promoted per [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md) when it gets its second caller). Briefs S12-D1 and S12-D3 name it. Binds S12-02, S12-12.

### Graph data model

- **Tables** ([data model](source-structure.md#data-model)): `evidence_graphs (id, ref, source_id, name, description)` with `UNIQUE (id, source_id)`; `evidence_graph_artifacts (graph_id, artifact_id)`; `subjects.graph_id NOT NULL` with a composite FK `(graph_id, source_id)` → `evidence_graphs (id, source_id)`. One migration. Binds S12-03.
- **`subjects.source_id` stays.** Document-level code (auto-reconciler agreement, credibility, the `source` audit scope, delete impact) keeps reading it; graph-level code reads `graph_id`. Binds S12-03, S12-07.
- **Ref prefix `GPH`**, added to `core/ref` beside SRC / ART / CIT / OBS and reserved. Binds S12-03.
- **Refs are minted in Go, never in SQL** ([`add-catalog-ref`](../../../.agents/skills/add-catalog-ref/SKILL.md)). The migration creates each Source's Default graph with `id = source_id` (deterministic, no SQL randomness) and no ref; `evidencegraphs.EnsureRefs` mints the missing refs at open, beside `users.EnsureRefs` in `core/catalogsession`. New graphs mint on insert. Binds S12-03.
- **A graph is part of its Source's record.** It cascades with the Source (which can only be deleted with no Artifacts and no Subjects). A link row is part of both its graph and its Artifact. The schema PR picks the exact `ON DELETE` and `catalogmodel` buckets through `add-catalog-migration` / `add-catalog-model`. Binds S12-03, S12-04.
- **Audit scope stays `source`.** Graph and link rows resolve to their Source's scope through `graph_id → source_id`. No `graph` scope type. Binds S12-03.
- **The migration backfills in SQL.** Every Subject points at its Source's Default graph; every Artifact is linked to it. Test fixtures that insert `sources` or `subjects` rows directly (`subjectpositions_test.go`, `deleteimpact/release_test.go`, `derivatives/thumbnail_test.go`) move onto `sources.Create` / `subjects.Create` in the same PR. Binds S12-03.

### Graph rules

- **`core/database/evidencegraphs` owns graphs and links from S12-03.** S12-03 creates it with the membership reads and the one link writer everything else calls: `SoleGraphTx`, `EnsureRefs`, membership reads (a Subject's graph; whether an Artifact is linked to a graph), `CitationsUsingArtifact(graph, artifact)` (through Observations → Subjects), and `Link` / `Unlink` of (graph, Artifact) pairs that apply the link rules. Artifact create, graph create and edit, and the Artifact card all go through it. S12-04 adds the graph lifecycle and `List`, the one read of graph rows (no second per-Source list). Binds S12-03, S12-04, S12-06, S12-10.
- **Every Source has at least one graph.** `sources.Create` mints one named `Default` in the same transaction. The last graph can't be deleted. Binds S12-03, S12-04.
- **Name and description are editable; graphs list alphabetically by name.** Binds S12-04, S12-11, S12-14.
- **Refusals are of two kinds.** "A graph with Subjects can't be deleted" is an inbound probe through `deleteimpact`, like other in-use deletes. "The last graph", "the graph's last link" and "Citations in this graph use this Artifact" are **domain invariants** in `evidencegraphs`, each with its own user-kind `apperr` code; the Citation refusal lists the Citations from `CitationsUsingArtifact`. They are not a private `sqlInUse`. Binds S12-03, S12-04.
- **Graphs are fully isolated.** A Subject belongs to one graph forever (no move). One check, `requireSameGraph` in `observations` (replacing `requireSubjectSource`), refuses an Observation whose Subject, subject value, or Citation Artifact is not in, or linked to, one graph. `connect.CreateCitedBridge` drops its own scope checks and relies on it; its new bridge Subject takes its graph from its endpoints. Go refuses at write time with user-kind errors. Binds S12-06.
- **Artifact ↔ graph links** ([rules](source-structure.md#rules)):
  - Creating a graph needs at least one Artifact; "Add graph" is unavailable while the Source has none.
  - The Default graph may have zero links until the Source's first Artifact, which links automatically.
  - A new Artifact links to the graphs it is given; `artifacts.CreateInput.GraphIDs` nil means every graph of its Source (S12-03), and `CreateArtifactRequest.graph_ids` carries the choice over FFI (S12-05).
  - Unlinking is refused when that graph's Citations use the Artifact (the error lists them) and when it is the graph's last link.
  - Deleting an Artifact keeps today's rule (refused while Citations use it) and may leave a graph with no links.

  Binds S12-03, S12-04, S12-05.
- **The Citation composer offers only the graph's linked Artifacts.** Binds S12-10.
- **Promote is per graph.** A Promote layer is one graph's Subjects; filing and unfiling bridges is per graph. Binds S12-09.
- **Agreement counts per Source.** The auto-reconciler keeps counting distinct Sources, not graphs. Binds S12-07, S12-09 (nothing in the reconciler changes).
- **Credibility stays per Source.** No per-graph grade this spike.

### Client

- **One graph list.** S12-04's `evidencegraphs.List` (one read: full rows with linked Artifact ids, sorted by name, optional Source filter) is the only graph read; S12-05 wraps it in one project-wide RPC with its own `CatalogQueryKey`. The Source page, Sources list rows, the composer filter, breadcrumbs and the omnibar location all read it. `GetSourceWorkspace` does not gain graphs ([`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md): do not fold a shared list into a page payload). Per-graph progress is its own key, keyed by graph id. Binds S12-05, S12-07, S12-08, S12-10, S12-11, S12-14.
- **A graph is a place.** `WorkspaceLocation` carries the graph id for the graph and composer surfaces (S12-08) and the Promote surface (S12-09), so history and restore land on that graph. Breadcrumbs read Source › graph name. Through [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md).
- **Effects are graph-keyed, in one step.** S12-07 keeps `effects.Set.Sources` keyed by Source; `graphcache` drops every graph of a changed Source. S12-08 re-keys `Set` / `Effect`, `WriteEffects.sources` and the Swift mutations (`.mutatedSourceGraph`, `.deletedSubject`, `.promotedSubject`) to graph ids together, so the wire never carries graph ids that Swift reads as Source ids. The Promote mutation moves in S12-09. Binds S12-07, S12-08, S12-09.
- **The graph list has its own effect.** S12-03 gives `evidence_graphs` and `evidence_graph_artifacts` an `Effect` that sets a catalog-wide `Set.EvidenceGraphs`; S12-05 carries it on the wire with one `CatalogQueryRegistry.invalidations(for:)` line for the graph list key. Every write that creates, edits, links or deletes a graph (Copy Source, Add Artifact, Source delete included) refreshes the list with no per-feature code. Binds S12-03, S12-05.
- **Single-graph Sources hide link checkboxes.** Linking is automatic while a Source has one graph; the checkboxes appear in the Artifacts section once there are two or more. Binds S12-13.
- **Source page:** a Graphs section with one row per graph (name, ref, description, Open), create, edit and delete. The header's Evidence graph jump is removed. No per-graph progress on this page. Binds S12-11, S12-12.
- **Sources list:** every row is an accordion listing its graphs, each with its own progress counts and an Open link. Open is available for a graph with at least one linked Artifact, as today's single button is. Binds S12-14.
- **Omnibar:** graphs are a searchable kind (name, description, ref); a hit opens the graph. Binds S12-15 (through [`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md)).

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S12-01 | Copy Source engine | 1 | `core/database/sourcecopy` (draft read, create from draft) and FFI `CopySource` | — | — | [Plan](prs/S12-01-copy-source-engine.md) | Landed #478 |
| S12-02 | Copy Source form | 1 | The Copy… form, reusing Add Source's title and type fields; Copy… on the Source page | S12-D1 | S12-01, S12-D1 | [Plan](prs/S12-02-copy-source-form.md) | Blocked — on board S12-D1 |
| S12-17 | Copy Source from the Sources list | 1 | Copy… in a Sources list row's context menu opens the S12-02 form | S12-D6 | S12-02, S12-D6 | — | Scoped |
| S12-16 | Add Source reports its effects | 1 | `CreateSourceResponse` gains `WriteEffects`; Add Source applies them with `.createdSource` and drops its Sources list write-through, as Copy Source does | — | S12-02 | — | Scoped |
| S12-03 | Evidence graphs in the catalog | 2 | Tables, `subjects.graph_id`, `GPH` refs (`EnsureRefs`), Default backfill; `evidencegraphs` reads, link writer and `SoleGraphTx`; Source create mints Default; Artifact create links | — | — | — | Scoped |
| S12-04 | Graph lifecycle | 2 | `evidencegraphs` `List` (the one read), create, update, delete with their refusals | — | S12-03 | — | Scoped |
| S12-05 | Graph RPCs and graph list | 2 | FFI and store methods for graphs and links; `CreateArtifactRequest.graph_ids`; one project-wide graph list key | — | S12-04 | — | Scoped |
| S12-06 | Graph isolation | 2 | One `requireSameGraph` for Observations; `connect` relies on it; the canonical walk joins on graph | — | S12-03 | — | Scoped |
| S12-07 | Graph-keyed reads | 3 | `graphcache`, loaders, positions, event titles and progress keyed and named by Evidence graph; effects stay per Source | — | S12-03, S11-14, S11-15 | — | Scoped |
| S12-08 | Graph id end to end | 3 | RPCs, effects, wire and Swift mutations take graph ids together; per-graph query keys; `WorkspaceLocation` carries the graph; `SoleGraphTx` goes | — | S12-05, S12-07 | — | Scoped |
| S12-09 | Promote per graph | 3 | Proposal, Apply, bridge filing, unpromote, the Promote key and location per graph | — | S12-08, S11-17, S11-25, S11-26, S11-27 | — | Scoped |
| S12-10 | Composer offers linked Artifacts | 3 | The Citation composer lists only the graph's linked Artifacts | — | S12-05, S12-08 | — | Scoped |
| S12-11 | Source page Graphs section | 4 | List graphs with Open and Delete; remove the header jump | S12-D2 | S12-05, S12-08, S12-D2 | — | Scoped |
| S12-12 | Graph form | 4 | Create and edit a graph: name, description, the shared Artifact checklist | S12-D3 | S12-02, S12-11, S12-D3 | — | Scoped |
| S12-13 | Artifact graph checkboxes | 4 | Graph checkboxes in the Add Artifact dialog and Artifact card edit, shown with two or more graphs; Swift only | S12-D4 | S12-12, S12-D4 | — | Scoped |
| S12-14 | Sources list graph accordion | 4 | Each row expands to its graphs with per-graph progress and Open | S12-D5 | S12-08, S12-D5 | — | Scoped |
| S12-15 | Graphs in the omnibar | 4 | Graphs as a searchable kind; a hit opens the graph | — | S12-08 | — | Scoped |
| S12-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S12-D1 | Copy Source form (Source page) | **S12-02** | Sheet: Add Source's title and type fields, part checkboxes, the Artifact checklist. Copy… on the Source page header |
| S12-D6 | Sources list — Copy… row menu | **S12-17** | A row's right-click menu with Copy…, opening the S12-D1 form |
| S12-D2 | Source page — Graphs section | **S12-11** | One row per graph (name, ref, description, Open, delete); replaces the header jump |
| S12-D3 | Graph form | **S12-12** | Create and edit: name, description, the Artifact checklist from S12-D1 all on; refusals worded |
| S12-D4 | Source page — Artifacts section graph links | **S12-13** | Checkboxes in the add dialog and card edit; refusal wording for unlink |
| S12-D5 | Sources list — graph accordion | **S12-14** | Expander per row, per-graph progress and Open |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Copy Source
  S12-01  Copy Source engine
  ✎ S12-D1 ──▶ S12-02  Copy Source form
  ✎ S12-D6 ──▶ S12-17  Copy Source from the Sources list
  S12-16  Add Source reports its effects
  Check: Copy an obituary; the new Source has "Copy of …", the same type,
         description, metadata in the same layout, credibility and the checked
         Artifacts, and no cover; unchecked parts are absent.

SLICE 2 — Graphs in the catalog
  S12-03  Evidence graphs in the catalog
  S12-04  Graph lifecycle
  S12-05  Graph RPCs and graph list
  S12-06  Graph isolation
  Check: An existing project opens unchanged. Go tests create a second graph,
         link and unlink Artifacts under the rules, and refuse a cross-graph
         relationship or an unlinked Artifact's Citation.

SLICE 3 — Graph id end to end
  S12-07  Graph-keyed reads
  S12-08  Graph id end to end
  S12-09  Promote per graph
  S12-10  Composer offers linked Artifacts
  Check: The canvas, composer and Promote open on a graph id; Back / Forward
         and restore return to it; Promote files only that graph.

SLICE 4 — Graphs in the UI
  ✎ S12-D2 ──▶ S12-11  Source page Graphs section
  ✎ S12-D3 ──▶ S12-12  Graph form
  ✎ S12-D4 ──▶ S12-13  Artifact graph checkboxes
  ✎ S12-D5 ──▶ S12-14  Sources list graph accordion
  S12-15  Graphs in the omnibar
  Check: the dogfood bar's census case end to end.

S12-99 — Dogfood close / docs
```

Which PRs can run in parallel: slice 1 is independent of slices 2 and 3 and can ship first. After S12-03, S12-04, S12-06 and S12-07 can run in parallel. After S12-08, S12-09, S12-10, S12-11, S12-14 and S12-15 are independent of each other.

### Dependencies at a glance

```text
S12-01 ──▶ S12-02 ──▶ S12-12
S12-02 ──▶ S12-16
S12-02 ──▶ S12-17
S12-D1 ──▶ S12-02   S12-D6 ──▶ S12-17

S12-03 ──▶ S12-04 ──▶ S12-05
S12-03 ──▶ S12-06
S12-03 ──▶ S12-07
S12-05, S12-07 ──▶ S12-08
S12-08 ──▶ S12-09, S12-14, S12-15
S12-05, S12-08 ──▶ S12-10
S12-05, S12-08 ──▶ S12-11 ──▶ S12-12 ──▶ S12-13
S12-D2 ──▶ S12-11   S12-D3 ──▶ S12-12   S12-D4 ──▶ S12-13   S12-D5 ──▶ S12-14

Spike 11:
S11-14, S11-15 ──▶ S12-07
S11-17, S11-25, S11-26, S11-27 ──▶ S12-09
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Two to six lines each. Link, do not restate.

### Slice 1 — Copy Source

#### S12-01 — Copy Source engine

**Ships.** `core/database/sourcecopy`: read a draft from a Source, create a Source from a draft; Copy composes them with the form's choices. An FFI `CopySource` RPC and its store method.
**Not this PR.** The form (S12-02). Graphs: before S12-03 the copy has none; after, it gets Default from `sources.Create`.
**Check.** Go and handler tests: each part copied only when chosen; Artifacts share `file_id`; notes, Subjects and Citations never copied; the copy has no cover; the original is untouched.
**Context.** Working rules "Copy Source"; [Copy Source](source-structure.md#copy-source); `add-catalog-query`, `add-ffi-handler`.

#### S12-02 — Copy Source form

**Ships.** The Copy… form from the Source page header, built on Add Source's title and type fields (`SourcesListView.addForm` is private today and is made reusable here), with the per-Artifact checklist; saving opens the copy's Source page. The form's model and presenter take any host, so S12-17 adds the list entry without changing them.
**Not this PR.** Graph choices in the form (a copy always starts with one Default graph). The Sources list entry point (S12-17).
**Check.** The bar's Copy Source outcome from the Source page.
**Context.** Working rules "Entry points", "One title and type form, one Artifact checklist"; brief S12-D1; `add-localized-string`, `add-ui-component`.

#### S12-17 — Copy Source from the Sources list

**Ships.** A right-click menu with Copy… on a Sources list row's page zone and the row's accessibility action, opening the S12-02 form for that Source (`SourcesModel.copying` / `openCopy`, one `.sourceCopyDialog` line).
**Not this PR.** Other row menu items; any change to the form.
**Check.** The bar's Copy Source outcome from a Sources list row's context menu.
**Context.** Working rule "Entry points"; brief S12-D6; `add-ui-component`, `add-localized-string`. After S12-02, which builds the form.

#### S12-16 — Add Source reports its effects

**Ships.** `CreateSourceResponse` carries `WriteEffects` (`writeEffectsProto` in `handlers.CreateSource`); `GenealogyStore.createSource` returns `(CatalogSource, CatalogWriteEffects)`; `SourcesModel.create()` calls `session.apply(effects, .createdSource)` and drops its `session.setQueryValue` insert into the Sources list, so Add Source and Copy Source refresh the list one way.
**Not this PR.** Any change to the Add Source form or toast.
**Check.** Adding a Source still shows it in the Sources list and updates its type's `usedBy`; a handler test asserts `effects.revision` > 0 and the new id in `source_scopes`; a Swift model test asserts the list is refreshed through the session, not written by hand.
**Context.** Raised by the S12-01 plan review (M-9, [`prs/S12-01-copy-source-engine.review.md`](prs/S12-01-copy-source-engine.review.md)); working rule "One title and type form, one Artifact checklist"; `add-ffi-handler`, `add-workspace-place`, `add-swift-test`. After S12-02 so the two PRs do not both edit `SourcesModel` at once.

### Slice 2 — Graphs in the catalog

#### S12-03 — Evidence graphs in the catalog

**Ships.** The migration (tables, `subjects.graph_id` and its composite FK, Default backfill with `id = source_id`), the `GPH` prefix and `EnsureRefs`, `catalogmodel`, effects (including `Set.EvidenceGraphs`) and audit-scope entries. `evidencegraphs` with its reads, `Link` / `Unlink` and `SoleGraphTx`. `sources.Create` mints Default; Subject creation uses `SoleGraphTx`; `artifacts.Create` takes `GraphIDs` (nil = every graph). Direct-insert test fixtures move onto the create functions.
**Not this PR.** Creating more graphs (S12-04), isolation checks (S12-06), graph-keyed reads (S12-07).
**Check.** An existing project opens unchanged; every Source has one Default graph with a `GPH-` ref, its Subjects and its Artifacts; `TestPragmaHonesty` passes.
**Context.** Working rules "Graph data model", "One bridge", "`core/database/evidencegraphs` owns graphs and links"; [data model](source-structure.md#data-model), [migration](source-structure.md#migration); `add-catalog-migration`, `add-catalog-model`, `add-catalog-ref`, `add-audit-scope`, `add-catalog-query`.

#### S12-04 — Graph lifecycle

**Ships.** `evidencegraphs` `List` (full rows with linked Artifact ids, by name, optional Source filter), `Create` (at least one Artifact, through `Link`), `Update` (name, description), `Delete` (Subjects through `deleteimpact`; last graph as a domain invariant).
**Not this PR.** RPCs (S12-05); UI (slice 4).
**Check.** Go tests for each rule in [Rules](source-structure.md#rules).
**Context.** Working rules "Graph rules", "Refusals are of two kinds"; `add-catalog-query`, `add-catalog-delete`.

#### S12-05 — Graph RPCs and graph list

**Ships.** FFI handlers and `GenealogyStore` / `FakeStore` methods to create, update and delete graphs and to link and unlink Artifacts; `CreateArtifactRequest.graph_ids`; one project-wide graph list RPC over `evidencegraphs.List` with its `CatalogQueryKey`, invalidated by the wire's `EvidenceGraphs` flag. New error codes mapped under `L10n.Errors`.
**Not this PR.** Any view change; `GetSourceWorkspace` is untouched.
**Check.** `runRPC` handler tests and Swift store tests.
**Context.** Working rules "One graph list", "The graph list has its own effect"; `add-ffi-handler`, `add-workspace-place`, `add-swift-test`, `add-localized-string`.

#### S12-06 — Graph isolation

**Ships.** `requireSameGraph` in `observations` insert and update replaces `requireSubjectSource` and also checks the subject value and the Citation's Artifact link; `connect.CreateCitedBridge` drops its duplicate checks and takes the bridge Subject's graph from its endpoints; `canonicalgraph`'s walk joins on `graph_id` (no signature change).
**Not this PR.** Re-keying the walk's callers (S12-07, S12-09).
**Check.** Go tests: each refusal with a user-kind error code; a same-graph write still succeeds.
**Context.** Working rule "Graphs are fully isolated". Today value subjects are not checked (`observations.go`, "Value subjects are not checked").

### Slice 3 — Graph id end to end

#### S12-07 — Graph-keyed reads

**Ships.** `graphcache` caches per Evidence graph; `sourcefill` loaders, `subjectpositions`, `sourceeventtitles` and progress read by graph, renamed to match; `effects.Set` stays keyed by Source and `graphcache` drops every graph of a changed Source. `promotealign.loadLayer` reads its graph through `SoleGraphTx` until S12-09.
**Not this PR.** FFI and Swift (S12-08); Promote (S12-09).
**Check.** Go tests: two graphs on one Source load, count and invalidate independently.
**Context.** Working rules "A re-keyed symbol is renamed", "Effects are graph-keyed, in one step", "`subjects.source_id` stays". Waits on S11-14 and S11-15, which rework `loadLayer` and add a shortlist index on `graphcache`.

#### S12-08 — Graph id end to end

**Ships.** List, create-Subject and cited-bridge RPCs take a graph id; `effects.Set`, `Effect`, `WriteEffects` and the Swift mutations carry graph ids instead of Source ids, together; `CatalogQueryKey.evidenceGraph` and the session cache are per graph; `WorkspaceLocation` carries the graph for graph and composer surfaces; existing entry points open the Source's graph. `SoleGraphTx` is deleted except Promote's use. Updates `macos-client-patterns.md`, `add-workspace-place`, `add-catalog-delete`.
**Not this PR.** Promote (S12-09); new chrome (slice 4).
**Check.** The canvas and composer work as before; Back / Forward and a relaunch return to the same graph.
**Context.** Working rules "A graph is a place", "Effects are graph-keyed, in one step"; `add-workspace-location`, `add-workspace-place`, `add-ffi-handler`. The largest PR; `plan-pr` may split it along FFI and Swift if it runs past ten steps.

#### S12-09 — Promote per graph

**Ships.** `promotealign.Propose`, `promote.SaveBatch`, bridge filing and `Unpromote` take a graph; the Promote RPCs, `PromoteEntry`, Promote's query key and its `WorkspaceLocation` surface carry the graph id. Promote's last `SoleGraphTx` use goes, and the function with it.
**Not this PR.** Any matching change (Spike 11).
**Check.** On a two-graph Source, Promote all on one graph proposes and files only its Subjects.
**Context.** Working rules "Promote is per graph", "Agreement counts per Source"; [`promote-graph-alignment.md`](../../promote-graph-alignment.md). Waits on S11-17, S11-25, S11-26 and S11-27, which rewrite the same Promote code; if S11-21 adds layer memoization, it is keyed per graph here.

#### S12-10 — Composer offers linked Artifacts

**Ships.** The Citation composer's Artifact picker lists only the current graph's linked Artifacts, read from the graph list.
**Not this PR.** Editing links (S12-12, S12-13).
**Check.** With two graphs linked to different Artifacts, each composer offers only its own.
**Context.** Working rules "The Citation composer offers only the graph's linked Artifacts", "One graph list".

### Slice 4 — Graphs in the UI

#### S12-11 — Source page Graphs section

**Ships.** The Graphs section on the Source page: one row per graph, alphabetical, with name, ref, description, Open and Delete (refusals worded). The header's graph jump is removed.
**Not this PR.** Create and edit (S12-12).
**Check.** A Source's graphs show and open; deleting an empty, non-last graph works; the others are refused with the reason.
**Context.** Working rule "Source page"; brief S12-D2.

#### S12-12 — Graph form

**Ships.** Create and edit a graph from the Graphs section: name, description, and the Artifact checklist from S12-02, all on for a new graph. Unavailable while the Source has no Artifact.
**Not this PR.** Link checkboxes in the Artifacts section (S12-13).
**Check.** Create a Jones graph linked to one page; edit its name and links; unlink refusals show the reason.
**Context.** Working rules "Artifact ↔ graph links", "One title and type form, one Artifact checklist"; brief S12-D3.

#### S12-13 — Artifact graph checkboxes

**Ships.** One checkbox per graph in the Add Artifact dialog (all on, sent as `graph_ids`) and in an Artifact card's edit (through link and unlink), shown only when the Source has two or more graphs. Swift only.
**Not this PR.** Graph creation (S12-12).
**Check.** On a one-graph Source nothing changes; on a three-graph Source a new page can be linked to one graph only.
**Context.** Working rule "Single-graph Sources hide link checkboxes"; brief S12-D4.

#### S12-14 — Sources list graph accordion

**Ships.** Every Sources list row expands to list its graphs alphabetically, each with its own progress counts and Open link.
**Not this PR.** Source-level progress totals (not wanted).
**Check.** The bar's Finding graphs outcome, Sources list half.
**Context.** Working rules "Sources list", "One graph list"; brief S12-D5.

#### S12-15 — Graphs in the omnibar

**Ships.** Graphs are a searchable kind (name, description, `GPH-` ref); a hit opens the graph.
**Check.** Searching a graph's ref or a word in its description finds it and opens it.
**Context.** Working rule "Omnibar"; `add-searchable-kind`.
