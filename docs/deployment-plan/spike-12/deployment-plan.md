# Deployment Plan — Spike 12

A Source can be copied to start a sibling one in a single step, and a big Source can hold several isolated Evidence graphs that are each promoted on their own. The design is in [`source-structure.md`](source-structure.md). It pulls in [`ideas/source-structure.md`](source-structure.md) (moved here) and settles it as two tools; the options weighed and parked are in [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md) and [`ideas/graph-focus-views.md`](../../ideas/graph-focus-views.md). One behavior, one PR.

**A brief covers one view.** A PR that changes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

**Spike 11 runs in parallel.** It is rebuilding Promote's matching and its loaders. The PRs here that touch the same code name the Spike 11 PRs they wait on in **Depends on**. A PR plan written before those land must say so and stop at `Blocked — on S11-NN`.

## Status

**Open.** Per-PR status is the [PR map](#pr-map).

> **Goal of this spike:** a big Source can be split into isolated, separately promoted Evidence graphs, and a Source can be copied to start a sibling one in a single step.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **Copy Source.** From an obituary, **Copy…** (on the Source page or a Sources list row's context menu) opens a form with the title `Copy of …`, a type dropdown preselected to the obituary's type, and checkboxes, all on, for description, metadata with its layout, credibility, and each Artifact. Saving opens the new Source with a Default graph and the copied Artifacts.
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
| [`source-structure.md`](source-structure.md) (was `ideas/source-structure.md`) — Problem 2, nested Sources and the "obituary or newspaper?" confusion | S12-01, S12-02 (Copy Source) | S12-D1 |
| Same — Problem 1, several Evidence graphs per Source | S12-03 – S12-15 | S12-D2 – S12-D5 |
| Same — graph ↔ Artifact association | S12-04, S12-05, S12-10, S12-12, S12-13 | S12-D3, S12-D4 |

**Back in the ideas folder** — [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md): nested Sources, derivation links, named templates, per-graph credibility, layered citation output. [`ideas/graph-focus-views.md`](../../ideas/graph-focus-views.md): focus views for one large graph.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Spike-wide

- **Two tools, one rule for choosing.** A part that can be cited on its own (its own author, title or date) is its own Source, started with Copy Source; a region of one work is another graph ([the two tools](source-structure.md#the-two-tools)). Copy and docs wording follow this rule.
- **High tolerance for churn.** No one but the product owner uses the app yet. No adapters, shims or kept-for-now paths beyond the one bridge below. A PR that moves a caller from Source to graph moves every caller and deletes the Source-keyed version.
- **One bridge, removed in S12-08.** Between S12-03 and S12-08 every Source has exactly one graph (creating a second is not exposed until S12-11 / S12-12), so Go resolves "the Source's graph" where a request still carries a Source id. S12-08 deletes that resolution.
- **Each PR updates the domain docs it makes false.** [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) (a Subject's home Source and graph, isolation; S12-03, S12-06), [`source-layer-data-model.md`](../../source-layer-data-model.md) (graphs, Artifact links, Copy Source; S12-01, S12-03, S12-04), [`promote-graph-alignment.md`](../../promote-graph-alignment.md) ("a layer is one graph's Subjects"; S12-09).

### Copy Source

- **The form** ([design](source-structure.md#copy-source)): title field prefilled `Copy of <title>`; required type dropdown preselected to the original's; checkboxes all on for description, metadata (values and layout), credibility (grade and argument), and one per Artifact. Notes are not offered. Graphs, Subjects, Citations and Observations are never copied. Binds S12-01, S12-02.
- **A copy is an ordinary create.** `sources.Copy` creates through `sources.Create` (so the copy gets its Default graph from S12-03 onward) and copies each checked Artifact as a new row on the copy with the same label, description and `file_id`, through `artifacts.Create`. One `writes.Run` transaction, action `copy_source`. Binds S12-01.
- **Entry points:** the Source page and a Sources list row's context menu. Binds S12-02.

### Graph data model

- **Tables** ([data model](source-structure.md#data-model)): `evidence_graphs (id, ref, source_id, name, description)` with `UNIQUE (id, source_id)`; `evidence_graph_artifacts (graph_id, artifact_id)`; `subjects.graph_id NOT NULL` with a composite FK `(graph_id, source_id)` → `evidence_graphs (id, source_id)`. One migration. Binds S12-03.
- **`subjects.source_id` stays.** Document-level code (auto-reconciler agreement, credibility, the `source` audit scope, delete impact) keeps reading it; graph-level code reads `graph_id`. Binds S12-03, S12-07.
- **Ref prefix `GPH`**, added to `core/ref` beside SRC / ART / CIT / OBS and reserved. Binds S12-03.
- **A graph is part of its Source's record.** It cascades with the Source (which can only be deleted with no Artifacts and no Subjects). A link row is part of both its graph and its Artifact. The schema PR picks the exact `ON DELETE` and `catalogmodel` buckets through `add-catalog-migration` / `add-catalog-model`. Binds S12-03, S12-04.
- **Audit scope stays `source`.** Graph and link rows resolve to their Source's scope through `graph_id → source_id`. No `graph` scope type. Binds S12-03.
- **The migration backfills in SQL.** One graph named `Default` per existing Source, with a minted `GPH-` ref; every Subject points at it; every Artifact is linked to it. Binds S12-03.

### Graph rules

- **Every Source has at least one graph.** `sources.Create` mints one named `Default` in the same transaction. The last graph can't be deleted. Binds S12-03, S12-04.
- **Name and description are editable; graphs list alphabetically by name.** Binds S12-04, S12-11, S12-14.
- **A graph with Subjects can't be deleted.** Refused through `deleteimpact` like other in-use deletes. Binds S12-04.
- **Graphs are fully isolated.** A Subject belongs to one graph forever (no move). An Observation's Subject, its subject value, and its Citation's Artifact must be in, or linked to, the same graph; `connect.CreateCitedBridge` checks the same. Go refuses at write time with user-kind errors. Binds S12-06.
- **Artifact ↔ graph links** ([rules](source-structure.md#rules)):
  - Creating a graph needs at least one Artifact; "Add graph" is unavailable while the Source has none.
  - The Default graph may have zero links until the Source's first Artifact, which links automatically.
  - A new Artifact links to every graph of its Source.
  - Unlinking is refused when that graph's Citations use the Artifact (the error lists them) and when it is the graph's last link.
  - Deleting an Artifact keeps today's rule (refused while Citations use it) and may leave a graph with no links.
  
  Binds S12-03 (auto-link on create), S12-04, S12-05.
- **The Citation composer offers only the graph's linked Artifacts.** Binds S12-10.
- **Promote is per graph.** A Promote layer is one graph's Subjects; filing and unfiling bridges is per graph. Binds S12-09.
- **Agreement counts per Source.** The auto-reconciler keeps counting distinct Sources, not graphs. Binds S12-07, S12-09 (nothing in the reconciler changes).
- **Credibility stays per Source.** No per-graph grade this spike.

### Client

- **A graph is a place.** `WorkspaceLocation` carries the graph id for the graph, citation composer and Promote surfaces, so history and restore land on that graph. Breadcrumbs read Source › graph name. Binds S12-08 (through [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md)).
- **One query key per graph.** `CatalogQueryKey.sourceGraph` becomes keyed by graph id; `WriteEffects` gains the touched graph ids so only those reload. Binds S12-07, S12-08.
- **Single-graph Sources hide link checkboxes.** Linking is automatic while a Source has one graph; the checkboxes appear in the Artifacts section once there are two or more. Binds S12-13.
- **Source page:** a Graphs section with one row per graph (name, ref, description, Open), create, edit and delete. The header's Evidence graph jump is removed. No per-graph progress on this page. Binds S12-11, S12-12.
- **Sources list:** every row is an accordion listing its graphs, each with its own progress counts and an Open link. Open is available for a graph with at least one linked Artifact, as today's single button is. Binds S12-14.
- **Omnibar:** graphs are a searchable kind (name, description, ref); a hit opens the graph. Binds S12-15 (through [`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md)).

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S12-01 | Copy Source engine | 1 | `sources.Copy` and FFI `CopySource`: title, type, chosen description, metadata and layout, credibility, Artifacts | — | — | — | Scoped |
| S12-02 | Copy Source form | 1 | The Copy… form, from the Source page and the Sources list context menu | S12-D1 | S12-01, S12-D1 | — | Scoped |
| S12-03 | Evidence graphs in the catalog | 2 | Tables, `subjects.graph_id`, `GPH` refs, Default backfill; Source create mints Default; new Artifacts link to every graph | — | — | — | Scoped |
| S12-04 | Graph lifecycle and Artifact links | 2 | `core/database/evidencegraphs`: list, create, update, delete, link and unlink, with their rules and delete impact | — | S12-03 | — | Scoped |
| S12-05 | Graph RPCs | 2 | FFI handlers and store methods for graphs and links; the Source workspace returns graphs and links | — | S12-04 | — | Scoped |
| S12-06 | Graph isolation | 2 | Observations, cited bridges and the canonical walk stay inside one graph | — | S12-03, S11-14 | — | Scoped |
| S12-07 | Graph-keyed reads | 3 | `graphcache`, loaders, positions, event titles and progress keyed by graph; effects carry graph ids | — | S12-03, S11-15 | — | Scoped |
| S12-08 | Graph id end to end | 3 | RPCs take a graph id; per-graph query keys; `WorkspaceLocation` and history carry the graph; the Source-id bridge goes | — | S12-05, S12-07 | — | Scoped |
| S12-09 | Promote per graph | 3 | Proposal, Apply, bridge filing and unpromote per graph; Promote opens on a graph | — | S12-08, S11-17, S11-25, S11-26, S11-27 | — | Scoped |
| S12-10 | Composer offers linked Artifacts | 3 | The Citation composer lists only the graph's linked Artifacts | — | S12-05, S12-08 | — | Scoped |
| S12-11 | Source page Graphs section | 4 | List graphs with Open and Delete; remove the header jump | S12-D2 | S12-05, S12-08, S12-D2 | — | Scoped |
| S12-12 | Graph form | 4 | Create and edit a graph: name, description, Artifact checkboxes | S12-D3 | S12-11, S12-D3 | — | Scoped |
| S12-13 | Artifact graph checkboxes | 4 | Graph checkboxes in the Add Artifact dialog and Artifact card edit, shown with two or more graphs | S12-D4 | S12-12, S12-D4 | — | Scoped |
| S12-14 | Sources list graph accordion | 4 | Each row expands to its graphs with per-graph progress and Open | S12-D5 | S12-08, S12-D5 | — | Scoped |
| S12-15 | Graphs in the omnibar | 4 | Graphs as a searchable kind; a hit opens the graph | — | S12-08 | — | Scoped |
| S12-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S12-D1 | Copy Source form | **S12-02** | Sheet: title, type dropdown, part checkboxes, one per Artifact. Entry points on the Source page and Sources list context menu |
| S12-D2 | Source page — Graphs section | **S12-11** | One row per graph (name, ref, description, Open, delete); replaces the header jump |
| S12-D3 | Graph form | **S12-12** | Create and edit: name, description, Artifact checkboxes all on; refusals worded |
| S12-D4 | Source page — Artifacts section graph links | **S12-13** | Checkboxes in the add dialog and card edit; refusal wording for unlink |
| S12-D5 | Sources list — graph accordion | **S12-14** | Expander per row, per-graph progress and Open |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Copy Source
  S12-01  Copy Source engine
  ✎ S12-D1 ──▶ S12-02  Copy Source form
  Check: Copy an obituary; the new Source has "Copy of …", the same type,
         description, metadata in the same layout, credibility and the checked
         Artifacts; unchecked parts are absent.

SLICE 2 — Graphs in the catalog
  S12-03  Evidence graphs in the catalog
  S12-04  Graph lifecycle and Artifact links
  S12-05  Graph RPCs
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

Which PRs can run in parallel: slice 1 is independent of everything and can ship first. S12-04 and S12-06 run in parallel after S12-03; so do S12-07 and S12-05. After S12-08, S12-09, S12-10, S12-11, S12-14 and S12-15 are independent of each other.

### Dependencies at a glance

```text
S12-01 ──▶ S12-02
S12-D1 ──▶ S12-02

S12-03 ──▶ S12-04 ──▶ S12-05
S12-03 ──▶ S12-06
S12-03 ──▶ S12-07
S12-05, S12-07 ──▶ S12-08
S12-08 ──▶ S12-09, S12-14, S12-15
S12-05, S12-08 ──▶ S12-10
S12-05, S12-08 ──▶ S12-11 ──▶ S12-12 ──▶ S12-13
S12-D2 ──▶ S12-11   S12-D3 ──▶ S12-12   S12-D4 ──▶ S12-13   S12-D5 ──▶ S12-14

Spike 11:
S11-14 ──▶ S12-06
S11-15 ──▶ S12-07
S11-17, S11-25, S11-26, S11-27 ──▶ S12-09
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Two to six lines each. Link, do not restate.

### Slice 1 — Copy Source

#### S12-01 — Copy Source engine

**Ships.** `sources.Copy` (Go) and an FFI `CopySource` RPC with its store method: a new Source with the given title and type, and the chosen parts copied from the original, in one audited transaction.
**Not this PR.** The form (S12-02). Graphs: before S12-03 the copy has none; after, it gets Default from `sources.Create`.
**Check.** Go and handler tests: each part copied only when chosen; Artifacts share `file_id`; notes, Subjects and Citations never copied; the original is untouched.
**Context.** Working rules "Copy Source"; [Copy Source](source-structure.md#copy-source). Open question: does the copy keep the original's cover?

#### S12-02 — Copy Source form

**Ships.** The Copy… form from the Source page and a Sources list row's context menu; saving opens the copy's Source page.
**Not this PR.** Graph choices in the form (a copy always starts with one Default graph).
**Check.** The bar's Copy Source outcome.
**Context.** Working rule "Entry points"; brief S12-D1; [`add-localized-string`](../../../.agents/skills/add-localized-string/SKILL.md), [`add-ui-component`](../../../.agents/skills/add-ui-component/SKILL.md).

### Slice 2 — Graphs in the catalog

#### S12-03 — Evidence graphs in the catalog

**Ships.** The migration (tables, `subjects.graph_id` and its composite FK, Default backfill), the `GPH` prefix, `catalogmodel`, effects and audit-scope entries. `sources.Create` mints Default; Subject creation writes the Source's graph (the one bridge); `artifacts.Create` links the new Artifact to every graph.
**Not this PR.** Creating more graphs (S12-04), isolation checks (S12-06), graph-keyed reads (S12-07).
**Check.** An existing project opens unchanged; every Source has one Default graph with its Subjects and Artifacts; `TestPragmaHonesty` passes.
**Context.** Working rules "Graph data model", "One bridge"; [data model](source-structure.md#data-model), [migration](source-structure.md#migration); `add-catalog-migration`, `add-catalog-model`, `add-catalog-ref`, `add-audit-scope`.

#### S12-04 — Graph lifecycle and Artifact links

**Ships.** `core/database/evidencegraphs`: `ListBySource` (alphabetical), `Create` (at least one Artifact), `Update` (name, description), `Delete` (refused with Subjects or as the last graph), `SetArtifactLinks` (refusals listing the Citations, and for the last link), with delete-impact wiring.
**Not this PR.** RPCs (S12-05); UI (slice 4).
**Check.** Go tests for each rule in [Rules](source-structure.md#rules).
**Context.** Working rules "Graph rules"; `add-catalog-query`, `add-catalog-delete`.

#### S12-05 — Graph RPCs

**Ships.** FFI handlers and `GenealogyStore` / `FakeStore` methods to list, create, update and delete graphs and set their Artifact links; `GetSourceWorkspace` returns graphs and links. New error codes mapped under `L10n.Errors`.
**Not this PR.** Any view change.
**Check.** `runRPC` handler tests and Swift store tests.
**Context.** `add-ffi-handler`, `add-swift-test`, `add-localized-string`.

#### S12-06 — Graph isolation

**Ships.** `observations` insert and update refuse a Subject, subject value or Citation Artifact outside one graph; `connect.CreateCitedBridge` and the canonical walk (`canonicalgraph.WalkSource` → per graph) stay in one graph.
**Not this PR.** Graph-keyed caches (S12-07).
**Check.** Go tests: each refusal, with a user-kind error code; a same-graph write still succeeds.
**Context.** Working rule "Graphs are fully isolated". Today only the Observation's own Subject is checked (`observations.requireSubjectSource`); value subjects are not. Waits on S11-14, which reworks the layer loaders that call `WalkSource`.

### Slice 3 — Graph id end to end

#### S12-07 — Graph-keyed reads

**Ships.** `graphcache` caches per graph; `sourcefill` loaders, `subjectpositions`, `sourceeventtitles` and `sourcegraphprogress` read by graph; `effects.Set` carries graph ids and drops per graph.
**Not this PR.** FFI and Swift (S12-08); Promote (S12-09).
**Check.** Go tests: two graphs on one Source load, count and invalidate independently.
**Context.** Working rules "`subjects.source_id` stays", "One query key per graph". Waits on S11-15 (shortlist index on `graphcache`).

#### S12-08 — Graph id end to end

**Ships.** List, create-Subject and cited-bridge RPCs take a graph id; `WriteEffects` carries graph ids; `CatalogQueryKey` and the session cache are per graph; `WorkspaceLocation` carries the graph for graph, composer and Promote surfaces; existing entry points open the Source's graph. The Source-id bridge from S12-03 is deleted.
**Not this PR.** Promote's own RPCs (S12-09); new chrome (slice 4).
**Check.** The canvas and composer work as before; Back / Forward and a relaunch return to the same graph.
**Context.** Working rule "A graph is a place"; `add-workspace-location`, `add-workspace-place`, `add-ffi-handler`. The largest PR; `plan-pr` may split it along FFI and Swift if it runs past ten steps.

#### S12-09 — Promote per graph

**Ships.** `promotealign.Propose`, `promote.SaveBatch`, bridge filing and `Unpromote` take a graph; the Promote RPCs and `PromoteEntry` carry the graph id.
**Not this PR.** Any matching change (Spike 11).
**Check.** On a two-graph Source, Promote all on one graph proposes and files only its Subjects.
**Context.** Working rules "Promote is per graph", "Agreement counts per Source"; [`promote-graph-alignment.md`](../../promote-graph-alignment.md). Waits on S11-17, S11-25, S11-26 and S11-27, which rewrite the same Promote code.

#### S12-10 — Composer offers linked Artifacts

**Ships.** The Citation composer's Artifact picker lists only the current graph's linked Artifacts.
**Not this PR.** Editing links (S12-12, S12-13).
**Check.** With two graphs linked to different Artifacts, each composer offers only its own.
**Context.** Working rule "The Citation composer offers only the graph's linked Artifacts".

### Slice 4 — Graphs in the UI

#### S12-11 — Source page Graphs section

**Ships.** The Graphs section on the Source page: one row per graph, alphabetical, with name, ref, description, Open and Delete (refusals worded). The header's graph jump is removed.
**Not this PR.** Create and edit (S12-12).
**Check.** A Source's graphs show and open; deleting an empty, non-last graph works; the others are refused with the reason.
**Context.** Working rule "Source page"; brief S12-D2.

#### S12-12 — Graph form

**Ships.** Create and edit a graph from the Graphs section: name, description, Artifact checkboxes all on for a new graph. Unavailable while the Source has no Artifact.
**Not this PR.** Link checkboxes in the Artifacts section (S12-13).
**Check.** Create a Jones graph linked to one page; edit its name and links; unlink refusals show the reason.
**Context.** Working rule "Artifact ↔ graph links"; brief S12-D3.

#### S12-13 — Artifact graph checkboxes

**Ships.** One checkbox per graph in the Add Artifact dialog (all on) and in an Artifact card's edit, shown only when the Source has two or more graphs.
**Not this PR.** Graph creation (S12-12).
**Check.** On a one-graph Source nothing changes; on a three-graph Source a new page can be linked to one graph only.
**Context.** Working rule "Single-graph Sources hide link checkboxes"; brief S12-D4.

#### S12-14 — Sources list graph accordion

**Ships.** Every Sources list row expands to list its graphs alphabetically, each with its own progress counts and Open link.
**Not this PR.** Source-level progress totals (not wanted).
**Check.** The bar's Finding graphs outcome, Sources list half.
**Context.** Working rule "Sources list"; brief S12-D5.

#### S12-15 — Graphs in the omnibar

**Ships.** Graphs are a searchable kind (name, description, `GPH-` ref); a hit opens the graph.
**Check.** Searching a graph's ref or a word in its description finds it and opens it.
**Context.** Working rule "Omnibar"; `add-searchable-kind`.
