# Deployment Plan — Spike 12

A Source can be copied to start a sibling one in a single step, and a big Source can hold several isolated Evidence graphs that are each promoted on their own. The design is in [`source-structure.md`](source-structure.md). It pulls in [`ideas/source-structure.md`](source-structure.md) (moved here) and settles it as two tools; the options weighed and parked are in [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md) and [`ideas/archive/graph-focus-views.md`](../../ideas/archive/graph-focus-views.md) (archived 2026-10-24). One behavior, one PR.

**Slices 5–7 fix what the [app-health audit](app-health-audit.md) of 2026-10-27 found** in the app as slices 1–4 left it: its High and Medium items, grouped into PRs by the files or the behavior they share (user, 2026-10-27). Each scope entry names its audit items (`A-n`); the audit doc holds their evidence and fixes. Slice 5 also makes an Artifact able to stand in no graph, so a new page can start its own graph, and slice 6 gives oversized JPEGs a thumbnail through a small Go decoder of our own.

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
8. **Link rules.** A new graph needs an Artifact; it starts with the Source's Artifacts that are in no graph checked, or with all of them when every Artifact is in a graph. A new Artifact starts linked to the graphs checked in Add artifact (all, by default) or to none. Unlinking an Artifact the graph's Citations use, the graph's last Artifact, or the last graph of an Artifact that has Citations is refused with the reason. A graph with Subjects, or a Source's last graph, can't be deleted. (Amended 2026-10-27 for slice 5; it read: a new graph starts with all Artifacts checked, and a new Artifact starts linked to every graph.)
9. **The canvas stays on its graph.** From the Jones canvas, opening the Smith graph through the omnibar or Back / Forward and placing a Person puts it on Smith. Back to a deleted graph lands on the Sources list, as any missing place does (amended 2026-10-27 at S12-21's plan review; it read: says the graph no longer exists and offers its Source page). The canvas's VoiceOver label and the history menu name the graph.
10. **A page can start its own graph.** On a Source with only its Default graph, Add artifact with the Default box unchecked adds the page in no graph; Add evidence graph then starts with only that page checked. Until it is linked, its card says it is in no evidence graph.
11. **A shared project is safe to open.** A stored file whose type is not allowed, or that is executable, is refused with its name; a metadata link opens only on the web; a crafted position, image or engine fault does not crash the app.
12. **Graph work stays fast on a split census.** A Citation save on a 60-graph census reloads only the graphs its Observations are on; opening a Source page does not scan the audit history; attaching a large scan does not stall the other panes.
13. **One place for each job.** Every audited write returns its effects; cache keys, search kinds and the canvas's tool state each have one declared place, pinned by tests; FakeStore re-implements no engine rule; a new Swift file needs no hand-written project entry; the docs, skills and kit README describe the code as it is.
14. **Keyboard and VoiceOver are safe on graph work.** Return in a delete confirm cancels; a dirty graph form asks before a link inside it closes it; VoiceOver hears a graph save or its failure and can jump to the Evidence graphs section by heading; counts read "1 subject", not "1 subjects".
15. **Every JPEG gets a thumbnail.** An 11,000 × 18,000 scan, over the 64-megapixel decode budget, shows a thumbnail instead of the missing-image glyph and can be used as a cover, through Go code that a Windows build reuses.

**Scale.** A census Source with 60 household graphs and 300 Citations on one page; a project of 3,000 Sources, 37,000 Subjects and 92,000 Observations, with 200,000 audit rows. Added 2026-10-27 for slices 5–7 from the audit's measurements; their plans write fan-out against it (`plan-spike` "Scale").

## Where the notes went

| Idea or note | PR | Brief |
| --- | --- | --- |
| [`source-structure.md`](source-structure.md) (was `ideas/source-structure.md`) — Problem 2, nested Sources and the "obituary or newspaper?" confusion | S12-01, S12-02, S12-17 (Copy Source) | S12-D1, S12-D6 |
| Same — Problem 1, several Evidence graphs per Source | S12-03 – S12-15 | S12-D2 – S12-D5 |
| Same — graph ↔ Artifact association | S12-03, S12-04, S12-05, S12-10, S12-12, S12-13 | S12-D3, S12-D4 |
| [S12-01 plan review](prs/S12-01-copy-source-engine.review.md) M-9 — Add Source and Copy Source refresh the Sources list two ways | S12-16 | — |
| Search hit kinds repeated as Swift literals, not pinned to `core/search`, before graphs add a kind (user, 2026-10-27) | S12-19 | — |
| [`dogfood/ux.md`](../../dogfood/ux.md) — "Dense Evidence graphs (census-scale)" (2026-09-24, descoped from Spike 10): split a census so the household you care about is on its own graph | S12-03 – S12-15 | S12-D2 – S12-D5 |
| [App-health audit](app-health-audit.md) (2026-10-27): the High and Medium items, A-1 – A-35 | S12-20 – S12-38 | — |
| No way to make a new page and a new graph together: Add artifact needs a graph and a new graph needs an Artifact (user, 2026-10-27) | S12-22, S12-38 | — |
| [`ideas/artifact-outside-graph-marker.md`](../../ideas/artifact-outside-graph-marker.md) — the card marker for an Artifact in no graph (user, 2026-10-27) | S12-38 | S12-D7 |
| [`ideas/archive/settled-list-count-publish.md`](../../ideas/archive/settled-list-count-publish.md) — shipped in c6dcd51d, aedf7aa2, 228bd821 and archived; its open question, a stale key read as settled (user, 2026-10-27) | S12-20, S12-26 | — |
| [`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md) — "Fake store keys", "Go reads every Swift mirror" and the done "Query-wait test helper" (user, 2026-10-27); the `effects` querier stays there, blocked by an import cycle (map review K-21) | S12-35, S12-20 | — |
| [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) — "Decode the crop before taking the catalog session" and "One stage-then-write entry point" (map review K-4) | S12-27 | — |
| [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md) — "One media type for the stored image", "Raster byte limit (SEC-3)" and "The PDF cut reopens its file" (user, 2026-10-27) | S12-27, S12-39, S12-28 | — |
| `ideas/large-jpeg-thumbnails.md`, moved here as [`oversized-jpeg-thumbnails.md`](oversized-jpeg-thumbnails.md): a portable, purpose-built decoder instead of macOS ImageIO (user, 2026-10-27) | S12-39 | — |
| [Map review](deployment-plan.review.md) of slices 5–7 (2026-10-27) | S12-20 – S12-39 | — |

**Back in the ideas folder** — [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md): nested Sources, derivation links, named templates, per-graph credibility, layered citation output. [`ideas/archive/graph-focus-views.md`](../../ideas/archive/graph-focus-views.md) (archived 2026-10-24): focus views for one large graph, including the "layers or filters" half of the dense-graphs dogfood note. [`ideas/artifact-outside-graph-marker.md`](../../ideas/artifact-outside-graph-marker.md): naming, in a graph delete's confirm, the Artifacts it leaves in no graph. The audit's Low items are listed, not scheduled, in [`app-health-audit.md`](app-health-audit.md#not-scheduled).

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
  - Slices 5–7 (map review K-13): S12-20 points [`add-catalog-model`](../../../.agents/skills/add-catalog-model/SKILL.md) step 6 at `core/database/effects/README.md` instead of listing `Effect` fields. S12-21 [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) (declared identity). S12-22 `source-layer-data-model.md`, `interpretation-layer-data-model.md` and [`catalog-deletes.md`](../../catalog-deletes.md). S12-25 and S12-26 `core/database/effects/README.md`, [`add-audit-scope`](../../../.agents/skills/add-audit-scope/SKILL.md), `add-workspace-place` and [`macos-client-patterns.md`](../../macos-client-patterns.md) rules 2 and 4. S12-27 [`use-catalog-session`](../../../.agents/skills/use-catalog-session/SKILL.md) and `source-layer-data-model.md`. S12-30 `macos/App/DesignSystem/README.md`. S12-34 `add-workspace-place`. S12-36 [`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md).

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
  - The Default graph may have zero links until the Source's first Artifact, which links automatically. From S12-38 it links when its box is checked, as it is by default.
  - A new Artifact links to the graphs it is given; `artifacts.CreateInput.GraphIDs` empty means every graph of its Source, linked through `evidencegraphs.LinkArtifact` (S12-03), and `CreateArtifactRequest.graph_ids` carries the choice over FFI (S12-05). From S12-22 the wire can also say "no graph" (working rule "An Artifact is the parent").
  - Unlinking is refused when that graph's Citations use the Artifact (the refusal is a plain code; S12-05's preflight lists them) and when it is the graph's last link. From S12-22 also when it is the last graph of an Artifact that has Citations (working rule "An Artifact with Citations keeps at least one graph").
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
- **The graph list has its own effect.** Superseded 2026-10-27 by S12-26, whose `reads` entry for the graph list by Source replaces the catalog-wide `evidence_graphs` list key (map review K-22). S12-03 gives `evidence_graphs` and `evidence_graph_artifacts` an `Effect` that sets the catalog-wide list key `evidence_graphs` (`effects.ListEvidenceGraphs` in `Set.Lists`, beside `vocabulary` and `property_usage`); the wire (`WriteEffects.lists`) and Swift (`CatalogList`) already carry every list key, so S12-05 adds one `CatalogQueryRegistry.invalidations(for: CatalogList)` line for the graph list. Every write that creates, edits, links or deletes a graph (Copy Source, Add Artifact, Source delete included) refreshes the list with no per-feature code. Binds S12-03, S12-05.
- **Single-graph Sources hide link checkboxes.** Linking is automatic while a Source has one graph; the checkboxes appear in the Artifacts section once there are two or more. Binds S12-13. **Superseded 2026-10-27** by "An Artifact is the parent": from S12-38 the boxes show on every Source that has a graph.
- **Source page:** a Graphs section with one row per graph (name, ref, description, Open), create, edit and delete. The header's Evidence graph jump is removed. Each row shows the graph's Subject and Observation counts, the same line as the Sources list (user, 2026-10-25; S12-11's "no progress on this page" is superseded). Binds S12-11, S12-12, S12-14.
- **Sources list:** every row has three columns, Source · Type · Evidence graph: the Source link (cover, title, ref and last update), its type, and one line per graph in name order, each a button carrying its status badge, always shown; nothing expands. A button shows the graph's name and its counts ("12 subjects · 40 observations") and opens the graph when it links at least one Artifact; otherwise it is dashed with "Needs an artifact to open". A Source with no Artifact shows one dashed "No evidence graph yet" block (board S12-D5, redrawn by the user 2026-10-27 from the tokens under each Source's title of 2026-10-26; the accordion first scoped is dropped; the separate Status column of the first redraw was folded into the buttons, corrected here 2026-10-27, map review K-20). Binds S12-14.
- **Status is per graph.** Not started: no Subjects. In progress: at least one Subject and none promoted. Promoted (was "Fully promoted"): at least one Subject promoted, even if others are not. A Source matches a status filter when any of its graphs has it (user, 2026-10-26). Binds S12-14.
- **Omnibar:** graphs are a searchable kind (name, description, ref); a hit opens the graph. Binds S12-15 (through [`add-searchable-kind`](../../../.agents/skills/add-searchable-kind/SKILL.md)).

### Audit fixes (slices 5–7)

Decisions on the [app-health audit](app-health-audit.md) and its [map review](deployment-plan.review.md) (user, 2026-10-27). They bind S12-20 – S12-39.

- **Fix the High and Medium items, grouped by shared files or behavior.** Each PR names its audit items (`A-n`); `plan-pr` reads them in the audit doc and re-checks each against the code, because its line numbers are from 731e423. The Low items are [not scheduled](app-health-audit.md#not-scheduled); a plan pulls one in only with the user's OK. Idea items pulled in are listed in "Where the notes went".
- **An Artifact is the parent; it may stand in no graph.** A graph still needs at least one Artifact; an Artifact needs no graph. Add artifact lists the Source's graphs as checkboxes on every Source that has a graph, one graph included, all checked by default, and none checked is allowed; card edit does the same. The wire says which graphs with one message, `GraphLinks { repeated string graph_ids }`, on `CreateArtifactRequest` and `UpdateArtifactRequest`: absent means the default (every graph on create, links unchanged on update), present and empty means none. Go carries it as one type (for example `Links{Given bool; IDs [][]byte}`) and never tells nil from empty (map review K-6). No composer offers an Artifact in no graph. Supersedes "Single-graph Sources hide link checkboxes" and the bar's old "a new Artifact starts linked to every graph"; the graph rules' "empty means every graph" becomes the wire default. Binds S12-22, S12-38.
- **An Artifact in no graph says so on its card.** The card reads "Not in an evidence graph" and offers linking it, which opens card edit with the graph boxes (user, 2026-10-27, reversing that day's deferral). Brief S12-D7, the Source page Artifacts section. Binds S12-38.
- **An Artifact with Citations keeps at least one graph.** A Citation with no Observation belongs to no graph and is reached through any graph that links its Artifact, so that Artifact may not leave its last graph. `Unlink` refuses the link with its own user-kind code. A graph `Delete` that would leave such an Artifact in no graph is refused through a `deleteimpact` inbound probe that lists those Citations, and `evidencegraphs.in_use` is reworded to cover Subjects and these Citations; no new extra gate (map review K-7). An Artifact with no Citation may leave every graph. Binds S12-22, S12-38.
- **One unlink check, one link writer.** One helper decides whether a (graph, Artifact) unlink is blocked and why: `Unlink` refuses from it and `UnlinkImpact` reports it. `UnlinkImpact` takes (graph, Artifact) pairs and each block carries its graph and reason; the graph form asks for its graph and each Artifact, the card for each graph and its Artifact, so the preflight and the write cannot disagree. `evidencegraphs.Update`'s links-then-unlinks loop becomes one `applyLinks` that the graph edit and the Artifact update both call (map review K-3). Binds S12-22, S12-38.
- **An Artifact card's Save is one write.** `UpdateArtifactRequest` carries the Artifact's `GraphLinks`, applied through `applyLinks` in the same transaction; the card's locks come from one `UnlinkImpact` call. The Swift copy of the last-link rule, the per-graph loop, the partial-save copy and the link and unlink RPCs the card was their last caller of go (map review K-12). Binds S12-22, S12-38.
- **A new graph starts with the Artifacts in no graph checked** when the Source has any; otherwise with all of them, as today. Changes S12-D3's default. Binds S12-38.
- **A destination declares its identity.** Each resolved place names the id its view resets on, and `WorkspaceDestinationHost` applies `.id(…)` once from it, in place of three hand-written `.id`s and the canvas's missing one. [`add-workspace-place`](../../../.agents/skills/add-workspace-place/SKILL.md) says so and names `SourcePageView`'s `prepare(for:)` as the exception or moves it (map review K-14). Binds S12-21.
- **A write names only the graphs whose loaded rows it moves, and the partial reads it moves by key.** Narrows "Effects are graph-keyed, in one step": a Citation names the graphs of its Observations, an Artifact names none, and a Source title edit names its graphs, which hold the title. Partial reads ride one keyed effect in the shape of `Set.Search`: `Effect.Reads []ReadPath{Read, Path}` → `Set.Reads` → `WriteEffects.reads` (a read name and its ids) → one Swift line per read name. It replaces `citation_sources` with citation reads per graph, carries graph progress only on writes that change counts (Subjects, Observations, Identity Claims), and carries the graph list by Source. Which write moves which read stays in Go's registry (map review K-2). Binds S12-25, S12-26.
- **A hidden graph or Source page goes stale; it does not reload.** `.evidenceGraph` and `.sourceWorkspace` take `whenHidden: .markStale`, as the lists do. A stale key is not settled: `QueryHandle.isStale`, set by `markStale` and cleared by `beginLoading`, keeps `onSettled` from publishing a cached count before the reload ([idea](../../ideas/archive/settled-list-count-publish.md), user 2026-10-27). Binds S12-26.
- **File work and decoding happen before the catalog session.** One staged-object type in `ingest` stages bytes from the project folder outside `catalogsession.Do` (stream, hash and sniff, and for an image decode and encode), inserts its `files` row inside the write, and discards what the write did not use. `ingest.File`, `StageImage` (through `StageDepictions`) and the thumbnail derivative all use it, and handlers stage before `withProjectCatalog` (map review K-4). Binds S12-27.
- **What the app hands to macOS to open is checked first, and Go decides the type.** The `files` read returns an object path only for a type on the ingest allowlist, so Swift checks only that the object has an extension, has no execute bit and is not an executable, application or script type; a refusal shows an error naming the file. Swift keeps no second allowlist. The viewer checks only that the sniffed type matches the media type Go returned, and a pixel budget; `ThumbnailDecodePolicy` gains that budget check as one shared function and keeps PNG and JPEG as the thumbnail cache's own list (map review K-8, K-23). A metadata link opens only through `open_href`, which Go computes with `urlshape.OpenHref` (web links only). Binds S12-23, S12-28.
- **Kit components announce themselves.** `PVToast` posts its own announcement, and `PVCallout` takes an opt-in announcement for an error, which the graph form and the Artifact card use; no call site posts its own (map review K-11). Binds S12-30.
- **A danger confirm's default action is Cancel.** Return cancels; the destructive button has no Return shortcut. The kit README changes with it. Binds S12-30.
- **A dirty form asks before a link inside it navigates away.** A Kept-linked Citation in the graph form asks "Discard changes?" first. Overrides S12-D3 GF-8. Binds S12-30.
- **No graph name in the canvas header.** Refused: the breadcrumb names the graph, and S12-29 adds it to the canvas's VoiceOver label and the Back / Forward history rows through the breadcrumbs' own name lookup (map review K-16). Binds S12-29.
- **Every audited write returns its effects, and a test holds them to it.** A test in `api/ffi` holds every request with a `user_id` to a response with a `WriteEffects` field: S12-32 adds it with an allow-list of S12-33's writes, and S12-33 empties the list (map review K-10). Sidebar counts become an effects list, `counts`, set by `sources` and `canonical_entities` creates and deletes, so the recount follows `WriteEffects.lists` and `navCountTriggers` goes (map review K-5). Binds S12-32, S12-33.
- **FakeStore is a seeded fake, not a second engine.** It returns seeded rows in seeded order and the effects and errors a test scripts (its `…Error` hooks and seeded effects); it re-implements no engine rule: no ordering, graph membership, effects resolution or refusal logic. Engine rules are tested in Go. A store method a PR adds to FakeStore follows this from now on; S12-35 removes the rules the fake carries today. This keeps `.cursor/rules/swift-tests.mdc` as written (map review K-9, user 2026-10-27, option a). Binds S12-22, S12-25, S12-35 and every PR that adds a store method.
- **The cache-key descriptor comes first in slice 6.** S12-34 lands before S12-25 and S12-26, so the progress patch, the graph list by Source and the Sources-list insert are entries in its `refresh` and `patch` slots, not hand-written session functions (map review K-1, user 2026-10-27). Binds S12-25, S12-26, S12-34.
- **Oversized JPEGs get a thumbnail from a reduced decode of our own.** A JPEG over `maxDecodePixels` is decoded at 1/8 scale from its DC coefficients only, then scaled and encoded as today; JPEGs within the budget keep the standard library, and other formats over it stay refused. It is pure Go, with no C library and no platform API, so a Windows build reuses it (user, 2026-10-27). It reads untrusted bytes, so it is bounded and fuzzed ([design](oversized-jpeg-thumbnails.md)). Binds S12-39.
- **The spike closes when every slice has landed.** S12-99 waits on slices 5–7, the refactors included.

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
| S12-14 | Graph counts in the Sources list and on the Source page | 4 | Each Sources list row shows a status badge and a button per graph with its progress, the button opening it; the Source page's graph cards show the count line; the Status filter reads each graph, with In progress redefined; `primary_count` removed | S12-D5 | S12-08, S12-D5 | [Plan](prs/S12-14-sources-list-graph-tokens.md) | Landed #576 |
| S12-19 | Typed omnibar hit kinds | 4 | `CatalogSearchKind` spells the search kinds once in Swift, pinned to `core/search` `Registry`, so a graph kind can't be missed in the app | — | — | [Plan](prs/S12-19-catalog-search-kind.md) | Landed #582 |
| S12-15 | Graphs in the omnibar | 4 | Graphs as a searchable kind; a hit opens the graph; the hit's Source projected on two new search-index display columns (migration) | — | S12-08, S12-19 | [Plan](prs/S12-15-graph-search.md) | Landed #583 |
| S12-20 | Docs and skills match the code | 5 | Audit A-6, A-23, A-33: stale skills and rules, domain docs that contradict the code, six missing DesignSystem README rows and a `check-guidance-names.py` check that fails on a kit folder with no row | — | — | [Plan](prs/S12-20-docs-match-code.md) | Landed #591 |
| S12-21 | The canvas keeps its graph and shows its failures | 5 | A-2, A-11, A-19: `.id(graphID)` on the canvas; a failed-read state and a deleted-graph fallback; the composer shows a failed Citation load; Swift only | — | — | [Plan](prs/S12-21-canvas-keeps-its-graph.md) | Planned |
| S12-22 | Artifacts can stand outside every graph | 5 | A-3, A-8 engine half and the user's fix: `GraphLinks` on Artifact create and update; the update links in one write through `applyLinks`; one unlink check behind `Unlink` and a pair-shaped `UnlinkImpact`; an Artifact with Citations keeps a graph (an inbound probe on graph delete) | — | — | — | Scoped |
| S12-38 | Add an Artifact without a graph; the card saves links in one write | 5 | A-3, A-8 Swift half and the user's fixes: graph boxes on every Source with a graph, none allowed; a card in no graph says so; one card write and preflight; a new graph starts with the Artifacts in no graph checked; the link and unlink RPCs go | S12-D7 | S12-22, S12-D7 | — | Scoped |
| S12-23 | The app opens only safe files and links | 5 | A-1, A-5: stored types filtered on read; executable objects refused in Go and Swift; metadata links open through Go's `open_href` | — | — | — | Scoped |
| S12-24 | Audit history: indexed note authors, audited pointer resets | 5 | A-4, A-22: an index on `audit_changes(entity_id)` (migration); cover and portrait pointer resets released and audited through `deleteimpact` | — | — | — | Scoped |
| S12-25 | Writes name only the graphs they move | 6 | A-7, A-16: Citation and Artifact effects narrowed; one keyed `reads` effect replaces `citation_sources` and carries progress only on writes that change counts, through S12-34's `refresh` slot; `EvidenceGraphProgress.source_id` dropped | — | S12-34 | — | Scoped |
| S12-26 | Reload only what moved, only when it is on screen | 6 | A-9, A-10, A-25: the graph list re-read per Source through a `reads` entry; Add Source inserts its row; hidden graph and Source page keys go stale and are not settled until reloaded; the Sources list sorts once per change | — | S12-25 | — | Scoped |
| S12-27 | File work before the catalog session | 6 | A-24 and map review K-4: one staged-object type stages ingest, depiction images and thumbnails outside `catalogsession.Do`; handlers stage before the session; one constant for the stored-image type; Go only | — | — | — | Scoped |
| S12-28 | A crafted project can't crash the app | 6 | A-26, A-27, A-28: bounded graph positions; a bounded Artifact viewer decode through `ThumbnailDecodePolicy`; the PDF crop re-checks its file through `objectURL`; panic recovery at the FFI boundary | — | S12-23 | — | Scoped |
| S12-29 | Name the Evidence graph; counts take plurals | 6 | A-12, A-14, A-20: the graph's name in the canvas VoiceOver label and history rows; "this Source" copy reworded; plural forms and a catalog check | — | — | — | Scoped |
| S12-30 | Dialogs and VoiceOver behave safely | 6 | A-13, A-17, A-18, A-21: Cancel is a danger confirm's default; a dirty graph form asks before a Citation link closes it; toasts and form errors announced; section headers are headings | — | — | — | Scoped |
| S12-31 | Xcode synchronized folders | 7 | A-34: `project.pbxproj` on synchronized folders; the pbxproj steps leave the skills and rule | — | — | — | Scoped |
| S12-32 | Promote and Source writes return their effects | 7 | A-29, A-35 (Promote): Promote and Source handlers return `WriteEffects`; sidebar counts become a `counts` effects list; a test pins effects on every audited write (allow-list for S12-33); the dead `promoteSubject` path goes | — | — | — | Scoped |
| S12-33 | Vocabulary writes return their effects; dead RPCs go | 7 | A-29, A-35 (the rest): Source type, metadata field, Property, binding and term handlers return `WriteEffects` and the pin test's allow-list empties; six unclaimed RPCs and two test-only store methods deleted | — | S12-32 | — | Scoped |
| S12-34 | One descriptor per cache key | 6 | A-30: each `CatalogQueryKey.Kind` has one typed descriptor (value, loader, spec, invalidation, a `patch` and a `refresh(ids)` slot), with a test that every kind has one; first in slice 6 | — | — | — | Scoped |
| S12-35 | FakeStore is a seeded fake | 6 | A-15 and the user's pulls from `shared-list-helpers`: the fake's engine rules (order, graph membership, effects resolution, refusals) give way to seeded and scripted values; FakeStore keys through `FieldSlug`; `FieldSlug` and `ProjectSlug` fixtures pinned by Go | — | — | — | Scoped |
| S12-39 | Thumbnails for oversized JPEGs | 6 | A pure-Go reduced decoder in `core/raster` reads only DC coefficients, so a JPEG over the 64-megapixel budget gets a 1/8-scale image to thumbnail; bounded and fuzzed; a byte limit inside `core/raster`; Go only | — | — | — | Scoped |
| S12-36 | Search kinds own their columns and locations | 7 | A-31: `KindSpec` names each field's column and carries its location builder; `Reproject` and `RebuildAll` read a projector table; a completeness test; Go only | — | — | — | Scoped |
| S12-37 | Canvas tool and sheet state as enums | 7 | A-32: one tool enum and one sheet enum replace the canvas's flags | — | S12-21 | — | Scoped |
| S12-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S12-D1 | Copy Source form (Source page) | **S12-02** | Sheet: Add Source's title and type fields, part checkboxes, the Artifact checklist. Copy… on the Source page header |
| S12-D6 | Sources list — Copy… row menu | **S12-17** | A row's right-click menu with Copy…, opening the S12-D1 form |
| [S12-D2](design/archive/S12-D2-source-page-graphs.md) | Source page — Graphs section | **S12-11** | One row per graph (name, ref, description, Open, delete); replaces the header jump. Board agreed 2026-10-25 |
| [S12-D3](design/archive/S12-D3-graph-form.md) | Graph form | **S12-12** | Create and edit: name, description, the Artifact checklist from S12-D1 all on; in-use Artifacts locked with their Citations listed; the last-graph notice names Add evidence graph. Board agreed 2026-10-25 |
| [S12-D4](design/archive/S12-D4-artifact-graph-links.md) | Source page — Artifacts section graph links | **S12-13** | Checkboxes in the add dialog and card edit; locked boxes with the reason; a refused card save shows on the card; canvas and composer callouts name the graph. Board agreed 2026-10-25 |
| [S12-D5](design/archive/S12-D5-sources-list-graph-tokens.md) | Sources list — graph buttons | **S12-14** | A three-column row (Source, Type, Evidence graph) with a button per graph carrying its status badge; the count line on the Source page's graph cards; the Status filter's Promoted label. Board agreed 2026-10-26, redrawn 2026-10-27 (first briefed as an accordion, then the tokens under each Source's title) |
| S12-D7 | Source page — Artifacts section: an Artifact in no graph | **S12-38** | The card's "Not in an evidence graph" line and its link action, and the graph boxes on a one-graph Source. Written just before S12-38 |

Slices 5–7 need one brief, S12-D7. The graph name in the canvas header is refused, and the canvas's error and missing-graph states reuse its existing callout (user, 2026-10-27).

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

SLICE 5 — Audit fixes: High
  S12-20  Docs and skills match the code
  S12-21  The canvas keeps its graph and shows its failures
  S12-22  Artifacts can stand outside every graph
  ✎ S12-D7 ──▶ S12-38  Add an Artifact without a graph; the card saves links in one write
  S12-23  The app opens only safe files and links
  S12-24  Audit history: indexed note authors, audited pointer resets
  Check: a Person placed after an omnibar jump lands on the graph on screen; a
         page added in no graph starts its own graph; an Artifact with a
         Citation cannot leave its last graph; a hostile file type or link is
         refused; a Source page opens without scanning history.

SLICE 6 — Audit fixes: Medium
  S12-34  One descriptor per cache key
  S12-35  FakeStore is a seeded fake
  S12-25  Writes name only the graphs they move
  S12-26  Reload only what moved, only when it is on screen
  S12-27  File work before the catalog session
  S12-28  A crafted project can't crash the app
  S12-29  Name the Evidence graph; counts take plurals
  S12-30  Dialogs and VoiceOver behave safely
  S12-39  Thumbnails for oversized JPEGs
  Check: bar outcomes 9, 11, 12, 14 and 15; a Citation save on a 60-graph census
         refreshes only its graphs; Return in a delete confirm cancels; the
         Promote toast reads "Filed 1 subject".

SLICE 7 — Structural debt
  S12-31  Xcode synchronized folders
  S12-32  Promote and Source writes return their effects
  S12-33  Vocabulary writes return their effects; dead RPCs go
  S12-36  Search kinds own their columns and locations
  S12-37  Canvas tool and sheet state as enums
  Check: bar outcome 13; handler tests show every audited write's effects,
         and the pin test passes with an empty allow-list.

S12-99 — Dogfood close / docs
```

Which PRs can run in parallel: slice 1 is independent of slices 2 and 3 and can ship first. After S12-03, S12-04, S12-06 and S12-07 can run in parallel. After S12-08, S12-09, S12-10, S12-11, S12-14 and S12-15 are independent of each other. S12-19 depends on nothing and goes before S12-15. In slices 5–7, every PR with no arrow below can run in parallel. Where two PRs edit the same code, the preferred merge order is: S12-22 before S12-25, so its new writes follow the narrowed effects; S12-34 and S12-35 first in slice 6 (map review K-1 and K-9); S12-34 before S12-32 and S12-33, which shrink its invalidation sets; S12-27 before S12-39, so the reduced decode runs outside the session. S12-38 before S12-30, which then opts the card's and the graph form's rewritten error callouts into announcing (map review K-11). S12-31 needs a Mac and lands best first, before the other slice 5–7 PRs add Swift files (map review K-15).

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

S12-22 ──▶ S12-38   S12-D7 ──▶ S12-38
S12-34 ──▶ S12-25 ──▶ S12-26
S12-23 ──▶ S12-28
S12-32 ──▶ S12-33
S12-21 ──▶ S12-37
slices 1–7 ──▶ S12-99

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

**Ships.** Every Sources list row has three columns (Source · Type · Evidence graph) and lists its graphs alphabetically, one line each: a button with the graph's name and counts carrying its status badge, opening that graph (board S12-D5, redrawn by the user 2026-10-27; replaces the tokens under each Source's title of 2026-10-26 and the accordion first scoped). Each graph card in the Source page's Evidence graphs section shows its full Subject and Observation count line (user, 2026-10-25). Both read the existing project-wide `CatalogQueryKey.evidenceGraphProgress` map, which `WorkspaceSession.apply` already refreshes per graph; the Source page place adds that key. The Status filter reads each graph with In progress redefined, and its last option is Promoted (working rule "Status is per graph"). `EvidenceGraphProgress.primary_count` loses its last reader and is removed from Go, the wire and Swift (user, 2026-10-26).
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

### Slice 5 — Audit fixes: High

#### S12-20 — Docs and skills match the code

**Ships.** The A-6 corrections, with `add-catalog-model` step 6 pointing at `core/database/effects/README.md` instead of listing `Effect` fields (map review K-13); `add-searchable-kind` lists `fieldValuesForKind`; `.cursor/rules/catalog-refs.mdc` reserves `GPH`. (The `swift-tests.mdc` event-titles name and the `design-system-layers.mdc` example were fixed with the map, by the change that added `scripts/check-guidance-names.py`.) The A-23 doc fixes. The A-33 README rows, an `add-ui-component` checklist line for them, and a `scripts/check-guidance-names.py` check that fails on a kit Components or Recipes folder with no README row (user, 2026-10-27). [`ideas/archive/settled-list-count-publish.md`](../../ideas/archive/settled-list-count-publish.md) is archived (it shipped in c6dcd51d, aedf7aa2 and 228bd821; its open question is S12-26's), and the done "Query-wait test helper" item leaves [`ideas/shared-list-helpers.md`](../../ideas/shared-list-helpers.md). Docs, plus that one check.
**Not this PR.** Rewriting `swift-tests.mdc`'s FakeStore paragraph (S12-35); the `add-searchable-kind` reshape (S12-36); the pbxproj steps (S12-31).
**Check.** `check-plan-docs.py` passes; each corrected line matches the code it names (the plan's DoD lists one `rg` per line).
**Context.** Audit A-6, A-23, A-33; spike-wide rule "Each PR updates the docs and skills it makes false"; [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md).

#### S12-21 — The canvas keeps its graph and shows its failures

**Ships.** Each resolved place declares the id its view resets on, replacing `ResolvedPlace.deepId` (which only `PlaceRegistryTests` reads), and `WorkspaceDestinationHost` applies `.id(…)` once from it, the canvas included; a hosting test moves from a graph of one Source to a graph of another; `add-workspace-place` says so. The canvas reads its handle's error into a callout with Read again; a graph missing from the session's graph list (derived from the list, not a new model flag) falls back to the Sources list through `fallbackToSectionRoot()`, as the navigation rule says for a missing deep id (user, 2026-10-27, plan review M-11). The composer shows a failed Citation load on an Artifact switch or a line refresh, as `loadCitationPreservingAuthoring` does. Swift only.
**Not this PR.** The graph's name in VoiceOver and history (S12-29); pruning history entries of deleted graphs; the canvas state enums (S12-37).
**Check.** Bar outcome 9, first two sentences; the A-2 manual steps place "Test" on Smith.
**Context.** Audit A-2, A-11, A-19; working rule "A destination declares its identity"; map review K-14, K-19; `PlaceRegistry` (`deepId`); `add-workspace-place`, `add-swift-test`, `add-localized-string`.

#### S12-22 — Artifacts can stand outside every graph

**Ships.** `GraphLinks` on `CreateArtifactRequest` and `UpdateArtifactRequest` (absent: the default; present and empty: none), carried in Go as one type; `CreateArtifactRequest.graph_ids` (7) is reserved. The Artifact update applies its links through `applyLinks`, the loop lifted from `evidencegraphs.Update`, in one write. One unlink check that `Unlink` refuses from and `UnlinkImpact` reports; `UnlinkImpact` takes (graph, Artifact) pairs and `EvidenceGraphUnlinkBlock` gains `graph_id` and `reason`. `Unlink` refuses the last graph of an Artifact with Citations with a new user-kind code; graph `Delete` gains an inbound probe for the Citations it would strand, with a `viaFormats` heading, and `evidencegraphs.in_use` is reworded to cover both reasons. Store methods with FakeStore stubs that return scripted links and refusals, `L10n.Errors` copy, and `source-layer-data-model.md`, `interpretation-layer-data-model.md`, `catalog-deletes.md`.
**Not this PR.** Any view, and deleting the link and unlink RPCs once the card stops calling them (S12-38); sorting accented graph names (audit L-1).
**Check.** Go tests: an Artifact created in no graph; a card-shaped update that links and unlinks in one revision; the A-3 probes refused with the new code and with the probe's Citations; `UnlinkImpact` reports every reason for a graph's pairs and for an Artifact's pairs, and `Unlink` refuses exactly what it reports. Handler tests for both request changes.
**Context.** Audit A-3, A-8; working rules "An Artifact is the parent", "An Artifact with Citations keeps at least one graph", "One unlink check, one link writer", "An Artifact card's Save is one write", "Refusals are of two kinds"; map review K-3, K-6, K-7; `add-ffi-handler`, `add-catalog-delete`, `add-localized-string`.

#### S12-38 — Add an Artifact without a graph; the card saves links in one write

**Ships.** The Add artifact dialog and the card edit show the graph boxes on every Source with a graph, one graph included, all checked by default; none checked is allowed and sends an empty `GraphLinks`. A card for an Artifact in no graph says "Not in an evidence graph" and offers linking it (brief S12-D7). A card Save is S12-22's one write and its locks come from one `UnlinkImpact` call; the graph form shows the new "keeps its last graph" lock too. A new graph's form starts with the Artifacts in no graph checked when there are any. What this makes dead goes: `isOnlyArtifact`, the per-graph loop, the partial-save strings, `graphListReloaded()`, `L10n.Sources.artifactGraphOnlyArtifact`, and the `LinkEvidenceGraphArtifact` and `UnlinkEvidenceGraphArtifact` RPCs on both sides (handlers, proto methods reserved, store methods, `CoreInvoke`, tests).
**Not this PR.** Naming, in a graph delete's confirm, the Artifacts it leaves in no graph ([idea](../../ideas/artifact-outside-graph-marker.md)); the Sources list and the composer.
**Check.** Bar outcome 10. On a two-graph Source, a card that unlinks one graph and links another saves in one revision; a refused save names the graph and leaves nothing half applied; no app code calls the deleted RPCs.
**Context.** Audit A-3, A-8; working rules "An Artifact is the parent", "An Artifact in no graph says so on its card", "A new graph starts with the Artifacts in no graph checked", "An Artifact card's Save is one write"; map review K-3, K-12; brief S12-D7; boards S12-D3 and S12-D4 for the chrome it reuses; `add-design-brief`, `add-localized-string`, `add-swift-test`, `add-ffi-handler`. After S12-22.

#### S12-23 — The app opens only safe files and links

**Ships.** Go's `files` read keeps only stored types the ingest allowlist accepts, so every object path the FFI returns has an allowed extension. `objectstore.openRegular` and Swift `ProjectFiles.isSafeObject` refuse an object with an execute bit; `openObject` refuses an object with no extension or of an executable, application or script type, and its callers show an error naming the file. Swift keeps no allowlist of its own for this. `MetadataWorkspaceEntry` gains `open_href` from `urlshape.OpenHref`; Swift opens only that, and `openSavedURL` goes.
**Not this PR.** Links inside PDFs (audit L-6); an App Sandbox.
**Check.** Bar outcome 11, files and links. Go tests: a hostile `media_type` read back is dropped; an executable object is refused; `open_href` is empty for `file:` and `smb:`.
**Context.** Audit A-1, A-5; working rule "What the app hands to macOS to open is checked first, and Go decides the type"; map review K-8; [`artifact-file-storage.md`](../../artifact-file-storage.md); `add-ffi-handler`, `add-localized-string`.

#### S12-24 — Audit history: indexed note authors, audited pointer resets

**Ships.** Migration `000057` adds `audit_changes_entity_id_idx`, with a `dbtest` query-plan test on the note queries. The two optional pointers become `facetRelease` entries that return update changes: deleting a cover Artifact records the Source's `cover_mode` and `primary_artifact_id` change with the fields `sources.SetCover` audits, and deleting a portrait's Observation records the `canonical_entities` pointer change. `TestFacetReleaseHonesty` extends to `SET NULL` foreign keys in `BucketOptional`; no second list or test. Updates `catalog-deletes.md` and `add-catalog-delete`, which tie a facet release to `CASCADE` today (map review K-13).
**Not this PR.** `GetSourceWorkspace`'s per-Artifact lookups (audit L-11). Spike 14's S14-01 makes the same pointer an Event's or Place's thumbnail; the release covers it whatever the kind.
**Check.** Bar outcome 12, second half; the A-22 probes find the `source` and `canonical_entity` changes in the delete's revision.
**Context.** Audit A-4, A-22; map review K-17; [`audit-revision-history.md`](../../audit-revision-history.md), [`catalog-deletes.md`](../../catalog-deletes.md); `add-catalog-migration`, `add-catalog-delete`, `add-catalog-model`.

### Slice 6 — Audit fixes: Medium

#### S12-34 — One descriptor per cache key

**Ships.** Each `CatalogQueryKey.Kind` has one descriptor holding its value type, loader, spec and invalidation rule, plus two update slots: `patch` (today's `lastUpdated`) and `refresh(ids)` (reload some rows and merge them into the cached value). Today's hand-written updates move into them with no change in behavior: `lastUpdated` and `patchUpdatedSource` become `patch` entries and `patchEvidenceGraphProgress` a `refresh` entry; S12-25 and S12-26 then add their reads as entries. `load(key:store:) -> Any`, `warmQuery`'s restated types and the separate `specs` table go. `Kind` is `CaseIterable`, and a test holds every kind to one descriptor. `add-workspace-place` teaches the new shape.
**Not this PR.** A behavior change to any key.
**Check.** The new test fails when a kind has no descriptor; `CatalogQueryRegistryTests` and `WorkspaceSessionTests` pass unchanged.
**Context.** Audit A-30; working rule "The cache-key descriptor comes first in slice 6"; map review K-1; maintainability Q2. Before S12-25 and S12-26, which add its entries, and S12-32 and S12-33, which shrink its invalidation sets.

#### S12-35 — FakeStore is a seeded fake

**Ships.** FakeStore loses the engine rules it re-implements: `graphListOrder`, `citationIsInGraph`, the effects resolution in `graphEffects` and `bumpSource`, and the refusal logic in its graph and Copy Source methods. In their place it returns seeded rows in seeded order and the effects and errors a test scripts through its `…Error` hooks. The Swift tests that leaned on those rules script them instead, and the rules stay tested in Go. FakeStore builds keys through `FieldSlug.kebab`, and `swift_vocabulary_test.go` pins the `FieldSlug` and `ProjectSlug` fixtures, decoding their escaped rows ([idea](../../ideas/shared-list-helpers.md)). `.cursor/rules/swift-tests.mdc` gains a line on scripting effects and refusals.
**Not this PR.** Pinning `apperr` codes to `L10n.Errors` (audit L-2).
**Check.** `rg` finds no ordering, membership, effects-resolution or refusal rule in `FakeStore.swift`; the Swift tests pass with scripted values; the fixture pins fail when a Swift fixture drifts from Go.
**Context.** Audit A-15; working rule "FakeStore is a seeded fake, not a second engine"; map review K-9; `add-swift-test`. First in slice 6 with S12-34; S12-22 adds its store methods by the same rule.

#### S12-25 — Writes name only the graphs they move

**Ships.** The keyed reads effect: `Effect.Reads` → `Set.Reads` → `WriteEffects.reads` (`citation_sources` replaced, its number reserved) → one Swift line per read name. Its first entries: a Citation's citation reads per graph (the graphs of its Observations, or every graph that links its Artifact while it has none), and graph progress on Subjects, Observations and Identity Claims. In the `effects` registry a Citation's `EvidenceGraph` is the graphs of its Observations, an Artifact names none, and a Source title keeps its graphs. Swift invalidates `.citationCounts` and `.citationsByArtifact` per graph; the composer's Artifact key becomes a visible key, so `applyWrite` stops querying after `apply`. `EvidenceGraphProgress.source_id` goes (reserved) and `graphprogress` gets an `evidencegraphs` code. Progress and the citation reads refresh through S12-34's `refresh` slot, and FakeStore returns scripted reads. Updates `effects/README.md`, `add-catalog-model`, `add-audit-scope`, `add-workspace-place` and `macos-client-patterns.md` rule 2.
**Not this PR.** The graph list by Source and hidden-key reloads (S12-26); the first-visit progress scan (audit L-12).
**Check.** Bar outcome 12, first half: an effects test on a many-graph Source pins "a Citation names only its Observations' graphs" and "an Artifact names none"; a Swift test shows one Citation save refetching only its graph's progress.
**Context.** Audit A-7, A-16; working rule "A write names only the graphs whose loaded rows it moves, and the partial reads it moves by key"; map review K-2, K-13; `effects.Set.Search` (the shape to copy); `add-ffi-handler`, `add-workspace-place`.

#### S12-26 — Reload only what moved, only when it is on screen

**Ships.** A `reads` entry for the graph list by Source (graph and link rows, Source create, copy and delete) that replaces the catalog-wide `evidence_graphs` list key: `effects.ListEvidenceGraphs`, the `evidence_graphs` wire list key, `CatalogList.evidenceGraphs` with its invalidation case and its `swift_vocabulary_test.go` row go. `ListEvidenceGraphsRequest` takes an optional Source, and the session re-reads only the named Sources' graphs and merges them into the cached list through the cache layer. Add and Copy Source insert the row `CreateSource` returns into the Sources list through the same layer, not a feature write-through. `.evidenceGraph` and `.sourceWorkspace` go stale while hidden, and `QueryHandle.isStale` (set by `markStale`, cleared by `beginLoading`) keeps `isSettled` false until a stale key reloads. `SourcesModel.visibleSources` is memoized against the revisions it reads. Updates `macos-client-patterns.md` rule 4 (list membership patched from a write's own row) and `source-layer-data-model.md` (the graph list's effect).
**Not this PR.** The descriptor itself (S12-34): the graph-list merge and the Sources-list insert are its `refresh` and `patch` entries.
**Check.** A link write re-reads one Source's graphs; an off-screen graph reloads only when revisited; a list that reappears after its key went stale publishes its count only after the reload; typing in Add Source does not re-sort the list (model tests).
**Context.** Audit A-9, A-10, A-25; [`ideas/archive/settled-list-count-publish.md`](../../ideas/archive/settled-list-count-publish.md) (its open question); working rules "One graph list", "A hidden graph or Source page goes stale; it does not reload"; S12-16 (why the hand-written insert went); `add-workspace-place`, `add-ffi-handler`. After S12-25, whose keyed effect it adds a read to.

#### S12-27 — File work before the catalog session

**Ships.** One staged-object type in `ingest`: it stages bytes from the project folder outside `catalogsession.Do` (stream to an `objectstore` temp file, hash and sniff, and for an image decode and encode), inserts its `files` row inside the write (`LookupByChecksum`, `Install`), and discards what the write did not use. `ingest.File`, `StageImage` (through `StageDepictions` in the citation, observation and connect handlers, one stage-then-write entry point) and the thumbnail derivative (`derivatives.Ensure`) all use it, and handlers stage before `withProjectCatalog`; the thumbnail made today inside the session through `thumbnailRelPath` splits into read, decode outside, and write (map review round 2). The stored-image media type becomes one constant in `core/depictionvalue`, used by `stage.go` (three literals today) and `ReadableImage` ([idea](../../ideas/depictions-and-likenesses.md)). Go only. Updates `use-catalog-session` and `source-layer-data-model.md`.
**Not this PR.** Oversized-JPEG thumbnails (S12-39); storing crops below 2,048 px ([idea](../../ideas/depictions-and-likenesses.md)).
**Check.** Bar outcome 12, last clause. A Go test holds a slow reader in the stage step while another session call completes; a depiction save decodes before the session; the ingest, dedupe, purge and depiction tests pass unchanged.
**Context.** Audit A-24; working rule "File work and decoding happen before the catalog session"; map review K-4; `core/ingest/stage.go` (`StagedImage`, the split to generalize); `derivatives.Ensure`; [`source-layer-data-model.md`](../../source-layer-data-model.md) (a purge and an import never interleave); `use-catalog-session`.

#### S12-28 — A crafted project can't crash the app

**Ships.** `subjectpositions` refuses a position outside one cell range on write and clamps it on read; the Swift placement code uses checked arithmetic. The Artifact viewer checks, before it decodes, that the sniffed type matches the media type Go returned and that the declared pixel size is within a budget, through one shared check in `ThumbnailDecodePolicy` (no type list of its own; the policy keeps PNG and JPEG for thumbnails); it shows a screen-sized decode and crops from a bounded one. Whether a file can be cut stays one predicate (`ArtifactRegionCrop.snapshot`), changed in place, which Spike 14's S14-03 gates Draw on. `ffi.Call` recovers a panic into `internal.unknown`, and `catalogsession.Do` defers its unlock. The PDF crop (`ArtifactRegionCrop`) reopens its file through `ProjectFiles.objectURL`: the off-main cut carries the project folder and the object's path instead of a bare URL ([idea](../../ideas/depictions-and-likenesses.md)).
**Not this PR.** Length caps on catalog text (audit L-10); the foreign-catalog open order (audit L-7).
**Check.** Bar outcome 11, last clause. Go tests for the position bounds and for a recovered panic that leaves the session usable; a Swift test feeds the placement extreme positions.
**Context.** Audit A-26, A-27, A-28; working rule "What the app hands to macOS to open is checked first, and Go decides the type"; map review K-8; `core/raster` for the bounds to reuse; `add-ffi-handler`. After S12-23, whose `files` filter the viewer relies on. Spike 14's S14-03 waits on this PR (map review K-25).

#### S12-29 — Name the Evidence graph; counts take plurals

**Ships.** One function gives a location's label from the session's names; the breadcrumbs, the Back / Forward history rows (graph, composer and Promote) and the canvas's VoiceOver label all read it. The A-14 strings name the Evidence graph. The A-20 strings get plural variations, `DeleteImpact.noun` reads plural keys, and `check-localizable-xcstrings.py` fails on a `%lld` noun with no variations.
**Not this PR.** A graph name in the canvas header (refused); tooltips on truncated names (audit L-17).
**Check.** Bar outcome 9, last sentence, and bar outcome 14, counts; the script fails on a planted `%lld subjects` key.
**Context.** Audit A-12, A-14, A-20; working rule "No graph name in the canvas header"; map review K-16; `WorkspaceToolbar.swift` (`BreadcrumbNames`, `HistoryJumpRow`); `add-localized-string`, `.cursor/rules/macos-l10n.mdc`.

#### S12-30 — Dialogs and VoiceOver behave safely

**Ships.** In `PVConfirm` the danger tone makes Cancel the default action and drops Return from the destructive button. The graph form's Kept-linked Citation rows ask "Discard changes?" when the form is dirty and are disabled while it saves. `PVToast` posts its own announcement, and `PVCallout` takes an opt-in announcement for an error, which the graph form's and the Artifact card's save errors use. `PVSectionHeader` titles are headings. Updates `macos/App/DesignSystem/README.md`.
**Not this PR.** Holding toasts longer while VoiceOver runs.
**Check.** Bar outcome 14. Return in a graph delete confirm cancels; a renamed graph survives a click on a Kept-linked Citation until Discard; VoiceOver hears "Evidence graph added", including from a toast drawn outside the shared overlay.
**Context.** Audit A-13, A-17, A-18, A-21; working rules "A danger confirm's default action is Cancel", "A dirty form asks before a link inside it navigates away", "Kit components announce themselves"; map review K-11; board S12-D3 (GF-8 overridden); `add-ui-component`.

#### S12-39 — Thumbnails for oversized JPEGs

**Ships.** A reduced JPEG decoder in `core/raster` that reads only DC coefficients and returns the image at 1/8 scale: baseline and progressive Huffman, 8-bit, one to three components with any sampling factors, restart intervals; arithmetic coding, 12-bit, lossless and four-component JPEGs return `ErrTooLarge`. `EncodeJPEG` uses it when the format is JPEG and the declared size is over `maxDecodePixels`, then scales and encodes as today, so `derivatives` stores the usual thumbnail row and the next Source open shows it. Every count read from the file is checked before use, and the 1/8 image counts against the budget. `core/raster` also takes a byte limit of its own, so every decode is bounded by bytes as well as pixels instead of trusting each caller's cap (likeness audit SEC-3, [idea](../../ideas/depictions-and-likenesses.md)). Go native fuzzing over the decoder. Go only.
**Not this PR.** EXIF orientation; a larger cover derivative; other formats over the budget; the Mac viewer (S12-28).
**Check.** Bar outcome 15. A JPEG just over 64 MP, built in the test from a lazy image with no full buffer, thumbnails to a 256 px long edge; a header-only JPEG declaring a huge frame is refused without allocating it; small baseline and progressive JPEGs decode within tolerance of the standard library's decode scaled by 8; the fuzz seeds pass; PNG and GIF header bombs still return `ErrTooLarge`; an input over the byte limit is refused before any decode.
**Context.** Working rule "Oversized JPEGs get a thumbnail from a reduced decode of our own"; [design](oversized-jpeg-thumbnails.md); `core/raster` (`EncodeJPEG`, `decode`, `maxDecodePixels`; the decoder runs under its caller's `Slot` and takes none itself, map review K-24), `core/derivatives/ensure.go`, `api/ffi/handlers/thumbnails.go`; ITU-T T.81. Preferred after S12-27. If the plan runs past about ten steps, progressive support splits into its own PR.

### Slice 7 — Structural debt

#### S12-31 — Xcode synchronized folders

**Ships.** The App and test groups of `project.pbxproj` become synchronized folders; `add-swift-test`, `add-ui-component` and `.cursor/rules/swift-tests.mdc` drop their pbxproj steps.
**Not this PR.** Any source change.
**Check.** `xcodebuild` build and test pass; a new Swift file builds with no pbxproj edit.
**Context.** Audit A-34. Needs a Mac; lands best before the other slice 5–7 PRs add Swift files.

#### S12-32 — Promote and Source writes return their effects

**Ships.** `ApplyPromoteGraphAlignment`, `UpdateSource`, `DeleteSource` and `SetSourceCover` return `WriteEffects`, and their callers apply them. Sidebar counts become an effects list, `counts`, set by `sources` and `canonical_entities` creates and deletes; the recount follows `WriteEffects.lists` and `navCountTriggers` goes. `.promotedSubject` goes when `conclusionTriggers` no longer needs it, and the other `CatalogMutation` cases the effects now cover go with it. The dead `promoteSubject` path goes (handler, proto method reserved, store methods, FakeStore). A test in `api/ffi` holds every request with a `user_id` to a response with `WriteEffects`, with an allow-list of S12-33's writes.
**Not this PR.** Vocabulary writes (S12-33).
**Check.** Handler tests assert each write's effects; the pin test passes with its allow-list; a Promote Done refreshes the canvas, progress and the sidebar counts through effects alone; a Source rename refreshes Promote's Source titles.
**Context.** Audit A-29, A-35; working rule "Every audited write returns its effects, and a test holds them to it"; map review K-5, K-10; [`macos-client-patterns.md`](../../macos-client-patterns.md) cache contract rule 2 (its exceptions shrink) and `add-workspace-place` (it teaches `navCountTriggers`), both updated; S12-16 (`navCountTriggers`); `add-ffi-handler`.

#### S12-33 — Vocabulary writes return their effects; dead RPCs go

**Ships.** The Source type, metadata field, Property, binding and term writes return `WriteEffects`, their callers apply them, and the `CatalogMutation` cases that go dead are deleted. `EnsureFileThumbnail`, `GetSubjectTypePresentation`, `ListPlaceableSubjectTypes`, `ListProperties`, `ListSubjectTypeProperties` and `RemoveInstallIdentity` go with their handlers, tests and FakeStore code (method numbers reserved), as do the test-only `listSubjectTypes` and `clearSubjectPosition`.
**Not this PR.** Deleting `UpdatePropertyTerm`, `DeletePropertyTerm` or `ReorderSubjectTypeProperties` (named by idea files): they stay and return `WriteEffects` like the rest, so the allow-list empties; the audit's test-net items (L-16).
**Check.** Handler tests assert effects; the pin test's allow-list is empty; no RPC is left without a Swift caller except the three kept; `macos-client-patterns.md` rule 2 lists only exceptions effects cannot express.
**Context.** Audit A-29, A-35; working rule "Every audited write returns its effects, and a test holds them to it"; `add-ffi-handler`. After S12-32, whose pin test it empties. The plan confirms `RemoveInstallIdentity` is not planned before deleting it.

#### S12-36 — Search kinds own their columns and locations

**Ships.** Each `KindSpec` field names its document column, and each spec carries its location builder; `fieldValuesForKind` and `locationFor` read the registry, and their per-kind switches go. `searchindex.Reproject` and `RebuildAll` read a projector table keyed by document kind. A completeness test in `core/search` holds every `Registry` kind to a column per field, a location builder and a projector. `add-searchable-kind` is rewritten to the shape. Go only.
**Not this PR.** Any ranking change; Swift.
**Check.** The `core/search` and `searchindex` tests pass unchanged; the completeness test fails for a field with no column or a kind with no projector.
**Context.** Audit A-31; map review K-18; `add-searchable-kind`; [`omnibar-search.md`](../archive/spike-3/omnibar-search.md).

#### S12-37 — Canvas tool and sheet state as enums

**Ships.** `EvidenceGraphModel` holds one `tool: EvidenceCanvasInputMode` (shift stays a modifier) and one optional sheet enum (create or edit, with its kind) in place of its eight flags; the old reads derive from them.
**Not this PR.** A behavior change.
**Check.** The canvas model tests pass unchanged, plus one per flag combination that can no longer be built.
**Context.** Audit A-32; maintainability Q9; precedent `CitationComposerModel.Phase`. After S12-21 (same files).
