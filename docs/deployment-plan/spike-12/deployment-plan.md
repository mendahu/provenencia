# Deployment Plan — Spike 12

A Source can be copied to start a sibling one in a single step, and a big Source can hold several isolated Evidence graphs that are each promoted on their own. The design is in [`source-structure.md`](source-structure.md). It pulls in [`ideas/source-structure.md`](source-structure.md) (moved here) and settles it as two tools; the options weighed and parked are in [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md) and [`ideas/archive/graph-focus-views.md`](../../ideas/archive/graph-focus-views.md) (archived 2026-10-24). One behavior, one PR.

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
5. **Finding graphs.** Every Sources list row lists its graphs alphabetically, each with its own status and progress, in a button that opens the graph. The Source page has a Graphs section with an Open button per graph; the header's graph jump is gone.
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
| Search hit kinds repeated as Swift literals, not pinned to `core/search`, before graphs add a kind (user, 2026-10-27) | S12-19 | — |
| [`dogfood/ux.md`](../../dogfood/ux.md) — "Dense Evidence graphs (census-scale)" (2026-09-24, descoped from Spike 10): split a census so the household you care about is on its own graph | S12-03 – S12-15 | S12-D2 – S12-D5 |

**Back in the ideas folder** — [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md): nested Sources, derivation links, named templates, per-graph credibility, layered citation output. [`ideas/archive/graph-focus-views.md`](../../ideas/archive/graph-focus-views.md) (archived 2026-10-24): focus views for one large graph, including the "layers or filters" half of the dense-graphs dogfood note.

## Working rules

These are the calls for this spike. PR plans treat them as binding.

### Spike-wide

- **Two tools, one rule for choosing.** A part that can be cited on its own (its own author, title or date) is its own Source, started with Copy Source; a region of one work is another graph ([the two tools](source-structure.md#the-two-tools)). Copy and docs wording follow this rule.
- **High tolerance for churn.** No one but the product owner uses the app yet. No adapters, shims or kept-for-now paths beyond the one bridge below. A PR that moves a caller from Source to graph moves every caller and deletes the Source-keyed version; it **replaces** Source-keyed fields (effects, query keys, wire fields) instead of adding graph ones beside them.
- **One bridge, one function, retired by S12-08, S12-18 and S12-09.** Between S12-03 and S12-08 every Source has exactly one graph (creating a second is not exposed until S12-11 / S12-12). `evidencegraphs.SoleGraphTx(sourceID)` is the only place that resolves "the Source's graph"; every interim caller uses it. S12-08 removes the handlers' use, S12-18 removes `subjects.InsertTx`'s fallback, and S12-09 removes Promote's and deletes `SoleGraph` and `SoleGraphTx` (planning S12-08, user 2026-10-22). The test fixture `evidencegraphstest.SoleGraph` stays.
- **A re-keyed symbol is renamed in the PR that re-keys it.** One name for the concept: `EvidenceGraph` in Go and Swift (`graphcache` `EvidenceGraph(id)`, `CatalogQueryKey.evidenceGraph(graphId:)`, `graphprogress`), "Evidence graph" in copy. Bare "graph" is already taken by `graphcache.Graph`, `canonicalgraph` and `graphalign`. Binds S12-07, S12-08, S12-09.
- **Each PR updates the docs and skills it makes false.**
  - S12-01: [`source-layer-data-model.md`](../../source-layer-data-model.md) (Copy Source).
  - S12-03: `source-layer-data-model.md` and [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md) (graphs, a Subject's graph); [`catalog-refs.md`](../../catalog-refs.md) and the [`add-catalog-ref`](../../../.agents/skills/add-catalog-ref/SKILL.md) prefix table (`GPH`); `core/database/effects/README.md`; [`catalog-deletes.md`](../../catalog-deletes.md).
  - S12-06: `interpretation-layer-data-model.md` and [`source-layer-data-model.md`](../../source-layer-data-model.md) § 7.1 (isolation).
  - S12-08: [`macos-client-patterns.md`](../../macos-client-patterns.md), [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md), [`add-catalog-delete`](../../../.agents/skills/add-catalog-delete/SKILL.md), [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md), `core/database/effects/README.md` (graph-keyed effects, keys and locations).
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
- **Refs are minted in Go, never in SQL** ([`add-catalog-ref`](../../../.agents/skills/add-catalog-ref/SKILL.md)). The migration creates each Source's Default graph with a fresh UUIDv7 id from the `uuid7()` SQL function (`core/database/driver.go`) and no ref; `evidencegraphs.EnsureRefs` mints the missing refs at open, beside `users.EnsureRefs` in `core/catalogsession`. New graphs mint on insert. Binds S12-03.
- **A graph is part of its Source's record.** It cascades with the Source (which can only be deleted with no Artifacts and no Subjects). A link row is part of both its graph and its Artifact. The schema PR picks the exact `ON DELETE` and `catalogmodel` buckets through `add-catalog-migration` / `add-catalog-model`. Binds S12-03, S12-04.
- **Audit scope stays `source`.** Graph and link rows resolve to their Source's scope through `graph_id → source_id`. No `graph` scope type. Binds S12-03.
- **The migration backfills in SQL.** Every Subject points at its Source's Default graph; every Artifact is linked to it. Test fixtures that insert `subjects` rows directly (`subjectpositions_test.go`) move onto `subjects.Create` in the same PR. Fixtures that insert only `sources` or `artifacts` rows from a package that cannot import the create functions (`deleteimpact/release_test.go`, `derivatives/thumbnail_test.go`) stay: the schema still accepts them. Binds S12-03.

### Graph rules

- **`core/database/evidencegraphs` owns graphs and links from S12-03.** S12-03 creates it with the membership reads and the one link writer everything else calls: `SoleGraphTx`, `EnsureRefs`, membership reads (a Subject's graph; whether an Artifact is linked to a graph), `CitationsUsingArtifact(graph, artifact)` (through Observations → Subjects), and `Link` / `Unlink` of (graph, Artifact) pairs that apply the link rules. Artifact create, graph create and edit, and the Artifact card all go through it. S12-04 adds the graph lifecycle and `List`, the one read of graph rows (no second per-Source list). `citations` owns the reads of a graph's Citations (S12-10: those the graph uses, the same rule as `CitationsUsingArtifact`, plus those with no Observation). Binds S12-03, S12-04, S12-06, S12-10.
- **Every Source has at least one graph.** `sources.Create` mints one named `Default` in the same transaction. The last graph can't be deleted. Binds S12-03, S12-04.
- **Name and description are editable; graphs list alphabetically by name.** Binds S12-04, S12-11, S12-14.
- **Refusals are of two kinds.** Graph delete refusals go through `deleteimpact`, so `GetDeleteImpact` and `Delete` agree: "A graph with Subjects can't be deleted" is an inbound probe, like other in-use deletes, and "the last graph" is a `last_graph` extra gate (beside `edge_locked`) that `Delete` maps to its own user-kind code `evidencegraphs.last_graph` (user, 2026-10-19, S12-04 review M-1). "The graph's last link" and "Citations in this graph use this Artifact" are **domain invariants** in `evidencegraphs`, each with its own user-kind `apperr` code. A refusal is a plain code, as for deletes; the Citations behind "Citations in this graph use this Artifact" come from `CitationsUsingArtifactTx` through S12-05's preflight read, shown with the DeleteImpact recipe. They are not a private `sqlInUse`. Binds S12-03, S12-04.
- **Graphs are fully isolated.** A Subject belongs to one graph forever (no move). One check, `requireSameGraph` in `observations` (replacing `requireSubjectSource`), refuses an Observation whose Subject, subject value, or Citation Artifact is not in, or linked to, one graph. `connect.CreateCitedBridge` drops its own scope checks and relies on it; its new bridge Subject takes its graph from its endpoints. Go refuses at write time with user-kind errors (`observations.other_graph`). A product reification Subject (type `source`) may be the value of a non-edge Observation in any graph, so `mentions` still crosses Sources (user, 2026-10-21). Binds S12-06.
- **Artifact ↔ graph links** ([rules](source-structure.md#rules)):
  - Creating a graph needs at least one Artifact; "Add graph" is unavailable while the Source has none.
  - The Default graph may have zero links until the Source's first Artifact, which links automatically.
  - A new Artifact links to the graphs it is given; `artifacts.CreateInput.GraphIDs` empty means every graph of its Source, linked through `evidencegraphs.LinkArtifact` (S12-03), and `CreateArtifactRequest.graph_ids` carries the choice over FFI (S12-05).
  - Unlinking is refused when that graph's Citations use the Artifact (the refusal is a plain code; S12-05's preflight lists them) and when it is the graph's last link.
  - Deleting an Artifact keeps today's rule (refused while Citations use it) and may leave a graph with no links.

  Binds S12-03, S12-04, S12-05.
- **The Citation composer offers only the graph's linked Artifacts**, and on each only the graph's Citations: those with an Observation on one of its Subjects, plus those with none yet (Citations added at S12-10 planning, user 2026-10-24). Binds S12-10.
- **Promote is per graph.** A Promote layer is one graph's Subjects; filing and unfiling bridges is per graph. Binds S12-09.
- **Agreement counts per Source.** The auto-reconciler keeps counting distinct Sources, not graphs. Binds S12-07, S12-09 (nothing in the reconciler changes).
- **Credibility stays per Source.** No per-graph grade this spike.

### Client

- **One graph list.** S12-04's `evidencegraphs.List` (one read: full rows with linked Artifact ids, sorted by name, optional Source filter) is the only graph read; S12-05 wraps it in one project-wide RPC with its own `CatalogQueryKey`. The Source page, Sources list rows, the composer filter, and breadcrumbs all read it; an omnibar graph hit carries its Source on its search document (S12-15, user 2026-10-27). `GetSourceWorkspace` does not gain graphs ([`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md): do not fold a shared list into a page payload). Per-graph progress is its own key, keyed by graph id. Binds S12-05, S12-07, S12-08, S12-10, S12-11, S12-14.
- **A graph is a place.** `WorkspaceLocation` carries the graph id for the graph, composer and Promote surfaces (S12-08; Promote moved from S12-09 because its rows read the canvas's key, user 2026-10-22), so history and restore land on that graph. Breadcrumbs read Source › graph name. Through [`add-workspace-location`](../../../.agents/skills/add-workspace-location/SKILL.md).
- **Effects are graph-keyed, in one step.** S12-07 keeps `effects.Set.Sources` keyed by Source; `graphcache` drops every graph of a changed Source. S12-08 re-keys `Set` / `Effect`, `WriteEffects.sources` and the Swift mutations (`.mutatedSourceGraph`, `.deletedSubject`, `.promotedSubject`) to graph ids together, so the wire never carries graph ids that Swift reads as Source ids. `.promotedSubject` moves in S12-08 too, because Promote reads the canvas's key (user, 2026-10-22). Binds S12-07, S12-08, S12-09.
- **The graph list has its own effect.** S12-03 gives `evidence_graphs` and `evidence_graph_artifacts` an `Effect` that sets the catalog-wide list key `evidence_graphs` (`effects.ListEvidenceGraphs` in `Set.Lists`, beside `vocabulary` and `property_usage`); the wire (`WriteEffects.lists`) and Swift (`CatalogList`) already carry every list key, so S12-05 adds one `CatalogQueryRegistry.invalidations(for: CatalogList)` line for the graph list. Every write that creates, edits, links or deletes a graph (Copy Source, Add Artifact, Source delete included) refreshes the list with no per-feature code. Binds S12-03, S12-05.
- **Single-graph Sources hide link checkboxes.** Linking is automatic while a Source has one graph; the checkboxes appear in the Artifacts section once there are two or more. Binds S12-13.
- **Source page:** a Graphs section with one row per graph (name, ref, description, Open), create, edit and delete. The header's Evidence graph jump is removed. Each row shows the graph's Subject and Observation counts, the same line as the Sources list (user, 2026-10-25; S12-11's "no progress on this page" is superseded). Binds S12-11, S12-12, S12-14.
- **Sources list:** every row has four columns, Source · Type · Status · Evidence graph: the Source link (cover, title, ref and last update), its type, and one line per graph in name order holding a status badge and a button, always shown; nothing expands. A button shows the graph's name and its counts ("12 subjects · 40 observations") and opens the graph when it links at least one Artifact; otherwise it is dashed with "Needs an artifact to open". A Source with no Artifact shows one dashed "No evidence graph yet" block and a dash for Status (board S12-D5, redrawn by the user 2026-10-27 from the tokens under each Source's title of 2026-10-26; the accordion first scoped is dropped). Binds S12-14.
- **Status is per graph.** Not started: no Subjects. In progress: at least one Subject and none promoted. Promoted (was "Fully promoted"): at least one Subject promoted, even if others are not. A Source matches a status filter when any of its graphs has it (user, 2026-10-26). Binds S12-14.
- **Omnibar:** graphs are a searchable kind (name, description, ref); a hit opens the graph. Binds S12-15 (through [`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md)).

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S12-01 | Copy Source engine | 1 | `core/database/sourcecopy` (draft read, create from draft) and FFI `CopySource` | — | — | [Plan](prs/S12-01-copy-source-engine.md) | Landed #478 |
| S12-02 | Copy Source form | 1 | The Copy… form, reusing Add Source's title and type fields; Copy… on the Source page | S12-D1 | S12-01, S12-D1 | [Plan](prs/S12-02-copy-source-form.md) | Landed #485 |
| S12-17 | Copy Source from the Sources list | 1 | Copy… in a Sources list row's context menu opens the S12-02 form; kit action menu items and `pvRightClickMenu`, with the cover, portrait and Clear menus moved onto them | S12-D6 | S12-02, S12-D6 | [Plan](prs/S12-17-sources-list-copy.md) | Landed #488 |
| S12-16 | Add Source reports its effects | 1 | `CreateSourceResponse` gains `WriteEffects`; Add Source applies them with `.createdSource` and drops its Sources list write-through, as Copy Source does; the session recounts the sidebar after Add, Copy, Source delete and Promote | — | S12-02 | [Plan](prs/S12-16-add-source-effects.md) | Landed #492 |
| S12-03 | Evidence graphs in the catalog | 2 | Tables, `subjects.graph_id`, `GPH` refs (`EnsureRefs`), Default backfill; `evidencegraphs` reads, link writer and `SoleGraphTx`; Source create mints Default; Artifact create links | — | — | [Plan](prs/S12-03-evidence-graphs-catalog.md) | Landed #498 |
| S12-04 | Graph lifecycle | 2 | `evidencegraphs` `List` (the one read), create, update (name, description and links in one write), delete with their refusals; the `last_graph` delete gate | — | S12-03 | [Plan](prs/S12-04-graph-lifecycle.md) | Landed #505 |
| S12-05 | Graph RPCs and graph list | 2 | FFI and store methods for graphs and links; `CreateArtifactRequest.graph_ids`; one project-wide graph list key | — | S12-04 | [Plan](prs/S12-05-graph-rpcs.md) | Landed #511 |
| S12-06 | Graph isolation | 2 | One `requireSameGraph` for Observations; `connect` relies on it; the canonical walk joins on graph | — | S12-03 | [Plan](prs/S12-06-graph-isolation.md) | Landed #515 |
| S12-07 | Graph-keyed reads | 3 | `graphcache`, loaders, positions, event titles and progress keyed and named by Evidence graph; effects stay per Source | — | S12-03, S11-14, S11-15 | [Plan](prs/S12-07-graph-keyed-reads.md) | Landed #520 |
| S12-08 | Graph id end to end | 3 | RPCs, effects, wire and Swift mutations take graph ids together; per-graph query keys; `WorkspaceLocation` carries the graph, Promote's included; the handlers' bridge goes | — | S12-05, S12-07 | [Plan](prs/S12-08-graph-id-end-to-end.md) | Landed #528 |
| S12-18 | A Subject needs its graph | 3 | `subjects.CreateInput.GraphID` required (Source read from the graph), `InsertTx`'s `SoleGraphTx` fallback deleted; test literals move onto graph ids | — | S12-08 | [Plan](prs/S12-18-subject-needs-graph.md) | Landed #532 |
| S12-09 | Promote per graph | 3 | Proposal, Apply, bridge filing, unpromote and the Promote RPCs per graph; `SoleGraph` / `SoleGraphTx` deleted | — | S12-08, S12-18, S11-17, S11-25, S11-26, S11-27 | [Plan](prs/S12-09-promote-per-graph.md) | Landed #535 |
| S12-10 | Composer offers linked Artifacts | 3 | The Citation composer lists only the graph's linked Artifacts, and on each only the graph's Citations (citation list and counts re-keyed by graph) | — | S12-05, S12-08 | [Plan](prs/S12-10-composer-linked-artifacts.md) | Landed #540 |
| S12-11 | Source page Graphs section | 4 | List graphs with Open and Delete; remove the header jump; breadcrumbs name the graph | S12-D2 | S12-05, S12-08, S12-D2 | [Plan](prs/S12-11-source-page-graphs.md) | Landed #549 |
| S12-12 | Graph form | 4 | Create and edit a graph: name, description, the shared Artifact checklist | S12-D3 | S12-02, S12-11, S12-D3 | [Plan](prs/S12-12-graph-form.md) | Landed #557 |
| S12-13 | Artifact graph checkboxes | 4 | Graph checkboxes in the Add Artifact dialog and Artifact card edit, shown with two or more graphs; Swift only | S12-D4 | S12-12, S12-D4 | [Plan](prs/S12-13-artifact-graph-checkboxes.md) | Landed #562 |
| S12-14 | Graph counts in the Sources list and on the Source page | 4 | Each Sources list row shows a status badge and a button per graph with its progress, the button opening it; the Source page's graph cards show the count line; the Status filter reads each graph, with In progress redefined; `primary_count` removed | S12-D5 | S12-08, S12-D5 | [Plan](prs/S12-14-sources-list-graph-tokens.md) | In review #576  |
| S12-19 | Typed omnibar hit kinds | 4 | `CatalogSearchKind` spells the search kinds once in Swift, pinned to `core/search` `Registry`, so a graph kind can't be missed in the app | — | — | [Plan](prs/S12-19-catalog-search-kind.md) | In review #582 |
| S12-15 | Graphs in the omnibar | 4 | Graphs as a searchable kind; a hit opens the graph; the hit's Source projected on two new search-index display columns (migration) | — | S12-08, S12-19 | [Plan](prs/S12-15-graph-search.md) | Planned |
| S12-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S12-D1 | Copy Source form (Source page) | **S12-02** | Sheet: Add Source's title and type fields, part checkboxes, the Artifact checklist. Copy… on the Source page header |
| S12-D6 | Sources list — Copy… row menu | **S12-17** | A row's right-click menu with Copy…, opening the S12-D1 form |
| [S12-D2](design/archive/S12-D2-source-page-graphs.md) | Source page — Graphs section | **S12-11** | One row per graph (name, ref, description, Open, delete); replaces the header jump. Board agreed 2026-10-25 |
| [S12-D3](design/archive/S12-D3-graph-form.md) | Graph form | **S12-12** | Create and edit: name, description, the Artifact checklist from S12-D1 all on; in-use Artifacts locked with their Citations listed; the last-graph notice names Add evidence graph. Board agreed 2026-10-25 |
| [S12-D4](design/archive/S12-D4-artifact-graph-links.md) | Source page — Artifacts section graph links | **S12-13** | Checkboxes in the add dialog and card edit; locked boxes with the reason; a refused card save shows on the card; canvas and composer callouts name the graph. Board agreed 2026-10-25 |
| [S12-D5](design/archive/S12-D5-sources-list-graph-tokens.md) | Sources list — graph buttons | **S12-14** | A four-column row (Source, Type, Status, Evidence graph) with a status badge and a button per graph; the count line on the Source page's graph cards; the Status filter's Promoted label. Board agreed 2026-10-26, redrawn 2026-10-27 (first briefed as an accordion, then the tokens under each Source's title) |

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
  S12-18  A Subject needs its graph
  S12-09  Promote per graph
  S12-10  Composer offers linked Artifacts
  Check: The canvas, composer and Promote open on a graph id; Back / Forward
         and restore return to it; Promote files only that graph.

SLICE 4 — Graphs in the UI
  ✎ S12-D2 ──▶ S12-11  Source page Graphs section
  ✎ S12-D3 ──▶ S12-12  Graph form
  ✎ S12-D4 ──▶ S12-13  Artifact graph checkboxes
  ✎ S12-D5 ──▶ S12-14  Graph counts in the Sources list and on the Source page
  S12-19  Typed omnibar hit kinds
  S12-15  Graphs in the omnibar
  Check: the dogfood bar's census case end to end.

S12-99 — Dogfood close / docs
```

Which PRs can run in parallel: slice 1 is independent of slices 2 and 3 and can ship first. After S12-03, S12-04, S12-06 and S12-07 can run in parallel. After S12-08, S12-09, S12-10, S12-11, S12-14 and S12-15 are independent of each other. S12-19 depends on nothing and goes before S12-15.

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
S12-08 ──▶ S12-18 ──▶ S12-09
S12-08 ──▶ S12-09, S12-14, S12-15
S12-19 ──▶ S12-15
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

**Ships.** A right-click menu with Copy… on a Sources list row's page zone and the row's accessibility action, opening the S12-02 form for that Source (`SourcesModel.copying` / `openCopy`, one `.sourceCopyDialog` line), and the kit's `PVActionMenuItem` / `pvRightClickMenu`, with the Source cover, Person portrait and Citation composer Clear menus moved onto them.
**Not this PR.** Other row menu items; any change to the form.
**Check.** The bar's Copy Source outcome from a Sources list row's context menu.
**Context.** Working rule "Entry points"; brief S12-D6; `add-ui-component`, `add-localized-string`. After S12-02, which builds the form.

#### S12-16 — Add Source reports its effects

**Ships.** `CreateSourceResponse` carries `WriteEffects` (`writeEffectsProto` in `handlers.CreateSource`); `GenealogyStore.createSource` returns `(CatalogSource, CatalogWriteEffects)`; `SourcesModel.create()` calls `session.apply(effects, .createdSource)`, drops its `session.setQueryValue` insert into the Sources list, so Add Source and Copy Source refresh the list one way. The session owns the sidebar recount: `CatalogQueryRegistry.navCountTriggers` (`.createdSource`, `.deletedSource`, `.promotedSubject`) and a `WorkspaceSession` hook that `WorkspaceView` wires to `CatalogCounts.refreshAll()`. That fixes the count lag after Copy Source and Source delete, and Promote's hand-written recount goes (user, 2026-10-19).
**Not this PR.** Any change to the Add Source form or toast, the Copy form, the delete flow, or Promote's write.
**Check.** Adding a Source still shows it in the Sources list and updates its type's `usedBy`; a handler test asserts `effects.revision` > 0 and the new id in `source_scopes`; a Swift model test asserts the list is refreshed through the session, not written by hand. The sidebar Sources count is right at once after Add Source, Copy Source (from the page or a list row) and a Source delete.
**Context.** Raised by the S12-01 plan review (M-9, [`prs/S12-01-copy-source-engine.review.md`](prs/S12-01-copy-source-engine.review.md)); working rule "One title and type form, one Artifact checklist"; `add-ffi-handler`, `add-workspace-place`, `add-swift-test`. After S12-02 so the two PRs do not both edit `SourcesModel` at once.

### Slice 2 — Graphs in the catalog

#### S12-03 — Evidence graphs in the catalog

**Ships.** The migration (tables, `subjects.graph_id` and its composite FK, Default backfill with `uuid7()` ids), the `GPH` prefix and `EnsureRefs`, `catalogmodel`, effects (the `evidence_graphs` list key; `Set.Lists`, `WriteEffects.lists` and Swift `CatalogList` replace the per-list bools) and audit-scope entries. `evidencegraphs` with its reads, `Link` / `Unlink` and `SoleGraphTx`. `sources.Create` mints Default; Subject creation uses `SoleGraphTx`; `artifacts.Create` takes `GraphIDs` (empty = every graph, through `evidencegraphs.LinkArtifact`). The direct `subjects` insert fixture moves onto `subjects.Create`. One open-time repair list, `catalogsession.ReconcileOnOpen`, mints the backfilled refs.
**Not this PR.** Creating more graphs (S12-04), isolation checks (S12-06), graph-keyed reads (S12-07).
**Check.** An existing project opens unchanged; every Source has one Default graph with a `GPH-` ref, its Subjects and its Artifacts; `TestPragmaHonesty` passes.
**Context.** Working rules "Graph data model", "One bridge", "`core/database/evidencegraphs` owns graphs and links"; [data model](source-structure.md#data-model), [migration](source-structure.md#migration); `add-catalog-migration`, `add-catalog-model`, `add-catalog-ref`, `add-audit-scope`, `add-catalog-query`.

#### S12-04 — Graph lifecycle

**Ships.** `evidencegraphs` `List` (full rows with linked Artifact ids, by name, optional Source filter), `Create` (at least one Artifact, through `Link`), `Update` (name, description and the full link set in one write), `Delete` (Subjects and the last graph both through `deleteimpact`: an inbound probe and a `last_graph` extra gate). The test-only `defaultGraphID` copies S12-03 left in the `artifacts`, `subjects`, `evidencegraphs` and `connect` tests move onto `List` (S12-03 review MNT-8).
**Not this PR.** RPCs (S12-05); UI (slice 4).
**Check.** Go tests for each rule in [Rules](source-structure.md#rules).
**Context.** Working rules "Graph rules", "Refusals are of two kinds"; `add-catalog-query`, `add-catalog-delete`.

#### S12-05 — Graph RPCs and graph list

**Ships.** FFI handlers and `GenealogyStore` / `FakeStore` methods to create, update and delete graphs and to link and unlink Artifacts; `CreateArtifactRequest.graph_ids`; one project-wide graph list RPC over `evidencegraphs.List` with its `CatalogQueryKey`, invalidated by the wire's `evidence_graphs` list key (one `CatalogQueryRegistry` list-key mapping line; S12-03 puts the key on the wire). `L10n.Errors` copy for each of `evidencegraphs.invalid`, `evidencegraphs.artifact_in_use` and `evidencegraphs.last_artifact` (S12-03 review UX-2), and for S12-04's `evidencegraphs.in_use` and `evidencegraphs.last_graph`. `DELETE_IMPACT_GATE_LAST_GRAPH` in `engine.proto`, its `gateToProto` case and the DeleteImpact recipe's extra-gate title and body for `last_graph`. A `viaFormats` delete-confirm heading for the `subjects.graph_id` edge, plural like `subjects.source_id` ("This graph has %lld subjects"), so a graph delete with Subjects does not fall back to "… reference this" (S12-03 review UX-1). The `artifact_in_use` preflight read (over `CitationsUsingArtifactTx`) takes the set of Artifact ids a graph edit removes, not one (graph, Artifact) pair, so the S12-12 form can name which removed Artifact blocks `evidencegraphs.Update` and its Citations (S12-04 review UX-1).
**Not this PR.** Any view change; `GetSourceWorkspace` is untouched.
**Check.** `runRPC` handler tests and Swift store tests. The DoD has one line per new code showing its `L10n.Errors` copy (not the unknown-error default) and one showing the `subjects.graph_id` heading.
**Context.** Working rules "One graph list", "The graph list has its own effect"; `add-ffi-handler`, `add-workspace-place`, `add-swift-test`, `add-localized-string`.

#### S12-06 — Graph isolation

**Ships.** `requireSameGraph` in `observations` insert and update replaces `requireSubjectSource` and also checks the subject value and the Citation's Artifact link; `connect.CreateCitedBridge` drops its duplicate checks and takes the bridge Subject's graph from its endpoints. Planning added: `subjects.CreateInput.GraphID` (optional until S12-08), the user-kind code `observations.other_graph`, and the reification exemption in the working rule. No walk changes: S11-19 moved one Source's walk in memory (`graphcache.Graph.WalkSource` over a loaded `SourceGraph`), and `canonicalgraph.Walk` reads only canonical rows.
**Not this PR.** The in-memory Source walk becoming per graph (S12-07, when `SourceGraph` is keyed by Evidence graph); re-keying Promote's layer (S12-09); removing `CreateCitedBridgeRequest.source_id` (S12-08).
**Check.** Go tests: each refusal with a user-kind error code; a same-graph write still succeeds.
**Context.** Working rule "Graphs are fully isolated". Today value subjects are not checked (`observations.go`, "Value subjects are not checked").

### Slice 3 — Graph id end to end

#### S12-07 — Graph-keyed reads

**Ships.** `graphcache` caches per Evidence graph; `sourcefill` loaders, `subjectpositions`, `sourceeventtitles` and progress read by graph, renamed to match; `effects.Set` stays keyed by Source and `graphcache` drops every graph of a changed Source. `promotealign.loadLayer` reads its graph through `SoleGraphTx` until S12-09. Planning added: claim evidence reads each member's own graph (`graphcache.Member.GraphID`), and `evidence_graphs` rows set `Effect.Source` so a deleted graph leaves the cache.
**Not this PR.** FFI and Swift (S12-08); Promote (S12-09).
**Check.** Go tests: two graphs on one Source load and count independently; a write under the Source drops both, and a write under another Source drops neither (per-graph drops come with S12-08's graph-keyed effects). Every handler test passes unchanged.
**Context.** Working rules "A re-keyed symbol is renamed", "Effects are graph-keyed, in one step", "`subjects.source_id` stays". Waits on S11-14 and S11-15, which rework `loadLayer` and add a shortlist index on `graphcache`.

#### S12-08 — Graph id end to end

**Ships.** List and create-Subject RPCs (and `GetCitation`) take a graph id, and the RPCs named for a Source are renamed for the Evidence graph; `CreateCitedBridgeRequest` drops `source_id` and takes no graph id (the bridge's graph comes from its endpoints, S12-06); `effects.Set`, `Effect`, `WriteEffects` and the Swift mutations carry graph ids instead of Source ids, together; `CatalogQueryKey.evidenceGraph` and the session cache are per graph, and the Swift `SourceGraph*` types are renamed; `WorkspaceLocation` carries the graph for graph, composer and Promote surfaces, and delete-impact and reconciliation locations fill it; existing entry points open the Source's first graph in the graph list. The handlers' `sourceGraphID` / `sourceEvidenceGraph` bridge goes. Updates `macos-client-patterns.md`, `add-workspace-place`, `add-catalog-delete`, `add-workspace-location`. Planning split off S12-18 and moved Promote's location and key here (user, 2026-10-22).
**Not this PR.** Promote's RPCs and Go code (S12-09); `subjects.CreateInput.GraphID` required (S12-18); the graph's name in breadcrumbs (S12-11); new chrome (slice 4).
**Check.** The canvas and composer work as before; Back / Forward and a relaunch return to the same graph.
**Context.** Working rules "A graph is a place", "Effects are graph-keyed, in one step"; `add-workspace-location`, `add-workspace-place`, `add-ffi-handler`. The largest PR; `plan-pr` may split it along FFI and Swift if it runs past ten steps.

#### S12-18 — A Subject needs its graph

**Ships.** `subjects.CreateInput.GraphID` is required and `CreateInput.SourceID` goes: `InsertTx` already reads the graph's Source with `evidencegraphs.SourceOf` when `SourceID` is empty (S12-08), so this PR deletes its `SoleGraphTx` fallback and the field. The `subjects.CreateInput{` test literals (66 in about 40 files) take a graph id, through `evidencegraphstest.SoleGraph` or `artifactstest.Split`. Go only.
**Not this PR.** Promote (S12-09).
**Check.** `go test` passes; `rg -n 'SoleGraphTx' core api` lists only `evidencegraphs` itself, `promotealign` (through `SoleGraph`) and tests.
**Context.** Split from S12-08 at planning (user, 2026-10-22); working rule "One bridge, one function"; `add-catalog-query`.

#### S12-09 — Promote per graph

**Ships.** `promotealign.Propose`, `promote.SaveBatch`, bridge filing and `Unpromote` take a graph; the Promote RPCs take `PromoteEntry.graphID` (S12-08 put the graph on `PromoteEntry`, its location and its query key). Promote's last `SoleGraph` use goes, and `SoleGraph` and `SoleGraphTx` with it (`core/database/audit/scopes_test.go` too).
**Not this PR.** Any matching change (Spike 11).
**Check.** On a two-graph Source, Promote all on one graph proposes and files only its Subjects.
**Context.** Working rules "Promote is per graph", "Agreement counts per Source"; [`promote-graph-alignment.md`](../../promote-graph-alignment.md). Waits on S11-17, S11-25, S11-26 and S11-27, which rewrite the same Promote code; if S11-21 adds layer memoization, it is keyed per graph here.

#### S12-10 — Composer offers linked Artifacts

**Ships.** The Citation composer's Artifact picker lists only the current graph's linked Artifacts, read from the graph list. On a shared Artifact the Citation stepper lists only the graph's Citations: `ListCitationsByArtifact` takes the graph and `CitationCountsBySource` becomes `CitationCountsByEvidenceGraph` (added at planning, user 2026-10-24). The empty callout names the graph.
**Not this PR.** Editing links (S12-12, S12-13); the canvas's cite gate and Files list (S12-13).
**Check.** With two graphs linked to different Artifacts, each composer offers only its own.
**Context.** Working rules "The Citation composer offers only the graph's linked Artifacts", "One graph list".

### Slice 4 — Graphs in the UI

#### S12-11 — Source page Graphs section

**Ships.** The Graphs section on the Source page: one row per graph, alphabetical, with name, ref, description, Open and Delete (refusals worded). The header's graph jump is removed. Breadcrumbs read Source › graph name (moved from S12-08 at its planning). The stored name `Default` (`evidencegraphs.DefaultName`, English catalog data) is shown as the researcher's own text, not localized; decide here whether a graph still named `Default` shows a localized label (S12-03 review UX-3).
**Not this PR.** Create and edit (S12-12).
**Check.** A Source's graphs show and open; deleting an empty, non-last graph works; the others are refused with the reason.
**Context.** Working rule "Source page"; brief S12-D2.

#### S12-12 — Graph form

**Ships.** Create and edit a graph from the Graphs section: name, description, and the Artifact checklist from S12-02, all on for a new graph. Saving an edit is one `Update` call carrying the checklist (S12-04 D-7), so a refused unlink keeps the whole edit unsaved. Unavailable while the Source has no Artifact. The `last_graph` delete notice says "Add another graph before you delete this one"; S12-11 made it reachable before any Add graph existed, so this PR checks the body names the control it adds, and words the case where Add graph is unavailable because the Source has no Artifact (S12-11 review UX-3).
**Not this PR.** Link checkboxes in the Artifacts section (S12-13).
**Check.** Create a Jones graph linked to one page; edit its name and links; unlink refusals show the reason.
**Context.** Working rules "Artifact ↔ graph links", "One title and type form, one Artifact checklist"; brief S12-D3.

#### S12-13 — Artifact graph checkboxes

**Ships.** One checkbox per graph in the Add Artifact dialog (all on, sent as `graph_ids`) and in an Artifact card's edit (through link and unlink), shown only when the Source has two or more graphs. Swift only. The canvas's cite gate and Files list read the graph's linked Artifacts (`CatalogEvidenceGraph.linkedArtifacts`; moved from S12-10 at its planning, user 2026-10-24), through the session accessor `WorkspaceSession.linkedArtifacts(graphID:artifacts:)` (S12-10). If the canvas reads it per row, give the session one id-indexed view of the graph list built when its value lands, and back `WorkspaceSession.evidenceGraph(id:)` and `evidenceGraphs(sourceID:)` with it, and move `SourcesModel.graphsBySourceCache` onto it (S12-10 review PRF-1). With the checkboxes in, the composer's `noArtifactsCallout` ("Add one from the Source page") should also offer linking an existing Artifact (S12-10 review UX-1). It adds no dialog: Add Artifact stays on `pvFormDialog(isPresented:)`, because its draft lives on `SourceArtifactsSection` and there is no model object for `pvFormDialog(item:)` (planning, 2026-10-25). Links show only in the dialog and in card edit, never on a resting card; a locked box shows its reason as a hint with no Citation list; card link changes save with the card's Save (user, 2026-10-25).
**Not this PR.** Graph creation (S12-12).
**Check.** On a one-graph Source nothing changes; on a three-graph Source a new page can be linked to one graph only.
**Context.** Working rule "Single-graph Sources hide link checkboxes"; brief S12-D4.

#### S12-14 — Graph counts in the Sources list and on the Source page

**Ships.** Every Sources list row has four columns (Source · Type · Status · Evidence graph) and lists its graphs alphabetically, one line each: a status badge and a button with the graph's name and counts, opening that graph (board S12-D5, redrawn by the user 2026-10-27; replaces the tokens under each Source's title of 2026-10-26 and the accordion first scoped). Each graph card in the Source page's Evidence graphs section shows its full Subject and Observation count line (user, 2026-10-25). Both read the existing project-wide `CatalogQueryKey.evidenceGraphProgress` map, which `WorkspaceSession.apply` already refreshes per graph; the Source page place adds that key. The Status filter reads each graph with In progress redefined, and its last option is Promoted (working rule "Status is per graph"). `EvidenceGraphProgress.primary_count` loses its last reader and is removed from Go, the wire and Swift (user, 2026-10-26).
**Not this PR.** Source-level progress totals (not wanted). Filed / primary counts or a Promote button on a button or card. Shift-F10 for the row menu (the kit's right-click menu has no keyboard trigger).
**Check.** The bar's Finding graphs outcome, Sources list half. On a census Source with three graphs, the row shows three badges and three buttons and each Source page card shows that graph's counts; adding an Observation in one graph updates only its button and card, with no reload.
**Context.** Working rules "Sources list", "Status is per graph", "Source page", "One graph list"; brief S12-D5.

#### S12-19 — Typed omnibar hit kinds

**Ships.** `CatalogSearchKind` in `macos/App/Platform/` with one case per `core/search` `Registry` kind; `CatalogSearchHitPresentation`, `CatalogSearchHitRow` and FakeStore's search switch on it; a `TestSwiftVocabularyMatchesGo` row pins it to the registry. `CatalogSearchHit.kind` stays a string, so an unknown kind still decodes.
**Not this PR.** The `evidence_graph` case (S12-15 adds it, and the pin then asks for it); typing the request's `kinds`.
**Check.** The Go pin passes and fails when a case is removed; the omnibar's labels and leads are unchanged; FakeStore search tests are unchanged.
**Context.** Working rule "Omnibar"; `.cursor/rules/macos-client.mdc` (wire vocabulary spelled once); [`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md); `api/ffi/swift_vocabulary_test.go` `catalog-wide lists` row.

#### S12-15 — Graphs in the omnibar

**Ships.** Graphs are a searchable kind (name, description, `GPH-` ref); a hit opens the graph. Each graph's search document carries its Source's id and ref in two new display columns (migration `000056`), and a Source rename reprojects its graphs (planning, user 2026-10-27).
**Check.** Searching a graph's ref or a word in its description finds it and opens it.
**Context.** Working rule "Omnibar"; `add-searchable-kind`; S12-19's `CatalogSearchKind` (add `evidenceGraph`, its label, lead and FakeStore priority).
