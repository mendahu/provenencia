# S12-08 — Graph id end to end

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 3 — Graph id end to end
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the same canvas, composer, Promote page and buttons open on a graph id; no view, layout or copy changes)
**Board:** —
**Depends on:** S12-05 (landed [#511](https://github.com/mendahu/provenencia/pull/511)), S12-07 (landed [#520](https://github.com/mendahu/provenencia/pull/520))
**Blocks:** S12-18, S12-09, S12-10, S12-11, S12-14, S12-15
**Branch:** `s12-08-graph-id-end-to-end`, branched from `main` after the plan branch `s12-08-plan` merges (so STEP-10 edits this file in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-07 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md) (its Swift mutation line only)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-queries.mdc`
**Context:** [data model](../source-structure.md#data-model), [`macos-client-patterns.md`](../../../macos-client-patterns.md), [`core/database/effects/README.md`](../../../../core/database/effects/README.md), the S12-07 [Landed note](S12-07-graph-keyed-reads.md#landed); working rules: "A graph is a place", "Effects are graph-keyed, in one step", "A re-keyed symbol is renamed in the PR that re-keys it", "One graph list", "One bridge, one function", "High tolerance for churn", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-10-22 against `main` @ `3c2ff0da`
**Review:** [`S12-08-graph-id-end-to-end.review.md`](S12-08-graph-id-end-to-end.review.md) — Fits / Ready with changes, 2 rounds; every finding answered

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

The Evidence graph canvas, the Citation composer and the Promote page open on an **Evidence graph id** instead of a Source id. The id travels the whole way: in the `WorkspaceLocation` that Back / Forward and restore keep, in the app's cache key for the graph's rows, in every read and write RPC the canvas and composer call, and in the write effects the engine reports. With two graphs on one Source, each opens, loads, refreshes and restores on its own. Today every Source has one graph, so the researcher sees no change.

**Why now.** S12-07 keyed the engine's reads by graph, but each handler still finds "the Source's one graph" through a bridge, and the app still keys everything by Source. Slice 4 (S12-11 graph rows with Open, S12-14 the Sources list accordion, S12-15 omnibar hits) and S12-10 (the composer's linked Artifacts) all need a place that names one graph. The working rule "Effects are graph-keyed, in one step" puts the effects, the wire and the Swift mutations in this one PR.

**What stays.** Promote's RPCs and Go code take a Source id until S12-09. `subjects.CreateInput.GraphID` stays optional until S12-18. Citation counts and citations-by-artifact stay per Source (`citation_sources`). `subjects.source_id`, the Source page, the Sources list layout and every copy string. Breadcrumb labels stay "Evidence graph" / "Evidence graph for …" (the graph's name in breadcrumbs comes with S12-11).

## 2. User stories

**US-1.** As a researcher, when I open a Source's Evidence graph, the composer or Promote, the app shows that graph's Subjects only, and Back / Forward and a relaunch return to the same graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source with one graph; the Sources list row's Open graph, or the Source page's Jump to evidence graph | The canvas opens with `location.graphId` = that graph's id and the same bubbles as on `main` |
| AC-1.2 | A Source whose graph list (sorted by name) is [Default, Jones] (FakeStore) | Open graph and Jump to evidence graph open Default (D-12) |
| AC-1.3 | A Source with no Artifact (its one graph has no linked Artifact) | Open graph and Jump to evidence graph stay disabled, as today |
| AC-1.4 | The canvas on graph G; open a card's property in the composer, then Promote from a card | Both locations carry `graphId` = G and `sourceId` = G's Source; Back returns to the canvas on G |
| AC-1.5 | History holds a canvas location on G, a composer location on G and a Promote location on G; the history file is written and read back | Each entry decodes with `graphId` = G |
| AC-1.6 | A history file written before this PR: a `graph` / `citationComposer` / `promote` entry with a `sourceId` and no `graphId` | It decodes as that Source's page (`sourceSurface` `.page`), as a history entry with no `sourceSurface` does today (D-11) |
| AC-1.7 | A delete-impact row for a Subject, a Citation or an Observation, or a conclusion's reconciliation record; the researcher follows it | The location carries the row's graph id and opens the canvas or composer on it |

**US-2.** As a researcher, a write on one graph refreshes that graph, not every graph of its Source.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Source S split into Default and Jones; a Subject is created on Jones | The write's effects report `evidence_graphs` = [Jones]; the engine drops only Jones's loaded graph; the app reloads only `.evidenceGraph(Jones)` and patches Jones's progress |
| AC-2.2 | Same; S's title is edited | `evidence_graphs` = [Default, Jones] (both cache the Source's title) |
| AC-2.3 | Same; an Artifact is added to S, or a Citation on one of S's Artifacts is edited | `evidence_graphs` = every graph of S, as `sources` = [S] meant on `main` (D-5) |
| AC-2.4 | Same; Jones is renamed | `evidence_graphs` = [Jones]; Default stays loaded |
| AC-2.5 | Same; a link between Jones and an Artifact is added or removed | `evidence_graphs` is empty and `lists` holds `evidence_graphs`, as on `main` |
| AC-2.6 | A Subject is deleted from the canvas on Jones | The write's effects reload `.evidenceGraph(Jones)` and patch Jones's progress; `.deletedSubject(sourceId: S)` reloads S's citation counts, as on `main` |
| AC-2.7 | A Subject is created or relabelled on the canvas or in the composer | The app applies the write's effects alone; `.evidenceGraph` of the Subject's graph reloads once, as `.mutatedSourceGraph` plus effects did on `main` |

**US-3.** As the engine, each graph read and write RPC takes a graph id; none resolves "the Source's graph".

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ListSubjects`, `ListSubjectPositions`, `ListObservationsByEvidenceGraph`, `ListSubjectMemberships`, `ListEvidenceGraphEventTitles` on Jones of a split Source | Only Jones's rows |
| AC-3.2 | The same RPCs with a well-formed id no graph has | Empty lists, no error, as an unknown Source reads today |
| AC-3.3 | `GetEvidenceGraphProgress` on Jones; on an unknown graph id | Jones's counts with `graph_id` = Jones and `source_id` = S; zero counts with `graph_id` = the request's id and `source_id` empty |
| AC-3.4 | `ListEvidenceGraphProgress` with Default and Jones holding Subjects | One row per graph with its `graph_id` and `source_id` |
| AC-3.5 | `CreateSubject` with `graph_id` = Jones | The Subject's `source_id` is S and its graph is Jones; with an unknown graph id the call fails with `subjects.invalid` |
| AC-3.6 | `GetCitation` with the Citation's graph id | The Citation, its notes and its Observations, as on `main` |
| AC-3.7 | `CreateCitedBridge` | Has no `source_id` field; the bridge lands on its endpoints' graph, as on `main` |
| AC-3.8 | `rg -n 'sourceGraphID\|sourceEvidenceGraph\|citations\.SourceID' api core` | No hit |

## 3. Scope

| In | Covered by |
| --- | --- |
| `evidencegraphs.SourceOf` *(new)*; `subjects.InsertTx` reads a graph's Source (D-3) | STEP-1, AC-3.5 |
| Effects keyed by graph: `Effect.EvidenceGraph`, `Set.EvidenceGraphs`, registry paths, `graphcache` drops by graph id (D-4, D-5) | STEP-2, AC-2.1 – AC-2.5 |
| `WriteEffects.evidence_graphs` replaces `sources` (D-1) | STEP-2, AC-2.1 |
| Graph read RPCs take `graph_id`; renamed RPCs and messages; handler helpers and `citations.SourceID` deleted (D-1, D-2, D-6) | STEP-3, AC-3.1 – AC-3.4, AC-3.6, AC-3.8 |
| `CreateSubjectRequest.graph_id`; `CreateCitedBridgeRequest.source_id` removed (D-1, D-3) | STEP-3, AC-3.5, AC-3.7 |
| `WorkspaceLocation.graph_id` on the wire, filled by delete-impact projectors; `ReconcilerOutcomeDetail.graph_id` (D-1, D-10) | STEP-4, AC-1.7 |
| Swift store, models and FakeStore take graph ids (D-7, D-14) | STEP-5 |
| Swift Source-graph types renamed `EvidenceGraph*`; `.evidenceGraph` and `.evidenceGraphProgress` keys; `.mutatedSourceGraph` deleted, `.promotedSubject(graphId:)`, `CatalogMutation.evidenceGraphId`; session progress map, cameras and pending selection by graph (D-8, D-9) | STEP-6, AC-2.1, AC-2.6, AC-2.7 |
| `WorkspaceLocation.graphId`, its normalization and the `evidenceGraph` / `citationComposer` factories; place registry, host, proto mapping, breadcrumb targets (D-10, D-11) | STEP-7, AC-1.4 – AC-1.6 |
| Canvas, composer, Promote entry and model, reconciliation record and the two Source-level entry points open a graph (D-12, D-13) | STEP-8, AC-1.1 – AC-1.4, AC-1.7 |
| Swift tests for each AC | STEP-9 |
| Docs and skills the re-key makes false; map updates (D-15) | STEP-10 |

| Out | Home |
| --- | --- |
| Promote's RPCs, `promotealign`, `promote.SaveBatch`, bridge filing and `Unpromote` taking a graph id; deleting `evidencegraphs.SoleGraph` / `SoleGraphTx` | S12-09 |
| Making `subjects.CreateInput.GraphID` required; deleting `InsertTx`'s `SoleGraphTx` fallback; moving the 66 `subjects.CreateInput{` test literals onto graph ids | S12-18 |
| Deleting the test fixture `evidencegraphstest.SoleGraph` | not planned (D-15: a test fixture, not a bridge) |
| The graph's name in breadcrumbs ("Source › graph name") | S12-11 (scope entry updated in STEP-10) |
| A per-graph Open in the Source page or the Sources list; removing the header jump | S12-11, S12-14 |
| Omnibar hits for graphs | S12-15 |
| The composer offering only linked Artifacts | S12-10 |
| Citation counts and citations-by-artifact per graph | not planned (they are Source-level reads: `citation_sources`) |
| Narrowing artifact and citation effects to the linked graphs (D-5) | not planned until a dogfood note shows the extra reload |
| `graphcache.Querier` onto `database.Querier` | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) (blocked, S12-07) |
| `conclusiondetails.Querier` and `effects.querier` onto `database.Querier` | [`docs/ideas/shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) § "One catalog querier": this PR changes one query in each, not their reads; the next PR that changes their read paths moves them |

## 4. Current state

Verified at `3c2ff0da`.

| Area | Where | Today |
| --- | --- | --- |
| Bridge | `core/database/evidencegraphs/evidencegraphs.go` `SoleGraphTx`, `SoleGraph`, `getTx`, `sqlGet`, `SubjectGraphTx` | `SoleGraph` serves the handlers and `promotealign.loadInputs`; `SoleGraphTx` also serves `subjects.InsertTx`'s empty-`GraphID` fallback; `getTx` reads one graph row (unexported) |
| Handler helpers | `api/ffi/handlers/source_read.go` `sourceGraphID`, `sourceEvidenceGraph` | Map a Source id to its sole graph; an unknown Source reads empty |
| Graph reads | `api/ffi/handlers/subjects.go` `ListSubjects`, `ListSubjectPositions`; `observations.go` `ListObservationsBySource`; `promote.go` `ListSubjectMemberships`; `event_titles.go` `ListSourceEventTitles`; `sourcegraphprogress.go` `ListSourceGraphProgress`, `GetSourceGraphProgress`, `protoProgress`; `citations.go` `GetCitation` (`citations.SourceID` then `sourceEvidenceGraph`) | Parse `source_id`, read the sole graph |
| Graph writes | `api/ffi/handlers/subjects.go` `CreateSubject` (`subjects.CreateInput{SourceID: sourceID}`, no `GraphID`); `connect.go` `CreateCitedBridge` (comment: `source_id` unread until S12-08) | `CreateSubject` lands on the sole graph through `InsertTx`'s fallback |
| `citations.SourceID` | `core/database/citations/citations.go` `SourceID` | Only caller is `GetCitation` |
| Proto | `api/proto/engine.proto`: `Method` `METHOD_LIST_OBSERVATIONS_BY_SOURCE = 69`, `METHOD_LIST_SOURCE_GRAPH_PROGRESS = 79`, `METHOD_GET_SOURCE_GRAPH_PROGRESS = 80`, `METHOD_LIST_SOURCE_EVENT_TITLES = 96`; `source_id = 2` on `ListSubjectsRequest`, `ListSubjectPositionsRequest`, `ListObservationsBySourceRequest`, `ListSubjectMembershipsRequest`, `ListSourceEventTitlesRequest`, `GetSourceGraphProgressRequest`; `CreateSubjectRequest.source_id = 3`; `CreateCitedBridgeRequest.source_id = 3`; `GetCitationRequest` (no graph); `SourceGraphProgress` (`source_id = 1`); `WriteEffects.sources = 2`; `WorkspaceLocation` (fields 1–18, no graph); `ReconcilerOutcomeDetail` (fields 1–19, `source_id = 10`) | Source-keyed |
| Dispatch | `api/ffi/dispatch.go` `MethodListObservationsBySource`, `MethodListSourceGraphProgress`, `MethodGetSourceGraphProgress`, `MethodListSourceEventTitles` and their `case` lines | Named for Sources |
| Effects | `core/database/effects/registry.go` `Set.Sources`, `Effect.Source`, the `Source:` paths on `sources` (`onField("title", self())`), `artifacts`, `citations`, `observations`, `subjects`, `evidence_graphs`, `identity_claims`, `identity_claim_evidence`, `valueSource`/`ownedValue`; `paths()`; `resolve.go` `Resolve`; `README.md` | Every path ends at `sources` |
| Wire effects | `api/ffi/handlers/write_effects.go` `writeEffectsProto` (`Sources: idStrings(set.Sources)`) | Source ids |
| Cache drop | `core/database/graphcache/evidence_graph.go` `dropGraphsOfSources`; `commit.go` `OnCommit` | Scans `g.evidenceGraphs` for each Source id |
| Delete-impact locations | `core/database/deleteimpact/kinds.go` `Location`; `projectors.go` observation projector (lines 72–128), `projectSubject`, `projectCitation`; `api/ffi/handlers/delete_impact.go` `locationFromImpact` | `graph` and `citationComposer` locations carry `SourceID` only |
| Reconciler outcome | `core/database/conclusiondetails/conclusiondetails.go` `Outcome` (`SourceID`), `sqlOutcomes` (selects `s.id` from `subjects s`); `api/ffi/handlers/conclusion_detail.go` `reconcilerOutcomeProto` | No graph |
| Swift key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `.sourceGraph(project:sourceId:)`, `.sourceGraphProgress(project:)`, `Kind` | One key for the canvas, composer and Promote rows; progress map keyed by Source id |
| Swift registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` specs (`.sourceGraph` `invalidateOn: [.mutatedSourceGraph, .deletedSubject, .deletedSource, .promotedSubject]`, `.sourceGraphProgress`), `load(key:store:)` (five `async let` reads by `sourceId`), `invalidations(for effects:)` (`for sourceID in effects.sources`), `Kind.invalidation(project:mutation:)` | Source-keyed |
| Swift mutations and effects | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `.mutatedSourceGraph(sourceId:)`, `.deletedSubject(sourceId:)`, `.promotedSubject(sourceId:)`, `CatalogMutationKind`, `CatalogWriteEffects.sources` | Source ids |
| Swift session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)`, `graphProgressSource(of:)`, `refreshSourceGraphProgress`, `patchSourceGraphProgress`, `graphCameras`, `graphCamera(for:)`, `setGraphCamera`, `pendingGraphSelections`, `setPendingGraphSelection`, `consumePendingGraphSelection`, `warmQuery` | Keyed by Source id |
| Swift rows | `macos/App/Features/Workspace/Session/SourceGraphSnapshot.swift` `SourceGraphRows` (`sourceId`), `SourceGraphSnapshot` (`sourceId`), `SourceGraphPlacedSubject`, `SourceGraphPlacedBridge` | 379 matching lines in 30 app files and 16 test files (`rg -c SourceGraph macos`) |
| Swift places | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `.sourcePromote`, `.sourceCitationComposer`, `.sourceGraph` (keys and `deepId` by `sourceId`), `.sourcesList` (warms `.sourceGraphProgress`), `.sourceDetail`; `WorkspaceDestinationHost.swift` `.sourceGraph` passes `sourceId` to `EvidenceGraphView` | No graph |
| Swift location | `macos/App/Features/Workspace/WorkspaceLocation.swift` `WorkspaceLocation`, `CodingKeys`, `init(from:)`, `==`; `macos/App/Platform/GoStore.swift` `mapWorkspaceLocationToProto`, `mapWorkspaceLocationFromProto` | No `graphId` |
| Breadcrumbs | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `breadcrumbItems` (graph crumb id `graph-\(sourceID)`, target `.graph` by `sourceId`) | Source-keyed |
| Canvas | `macos/App/Features/EvidenceGraph/EvidenceGraphView.swift` `EvidenceGraphView(sourceID:…)`, `EvidenceGraphContent`; `EvidenceGraphModel.swift` `sourceID`, `graphKey`, `composerLocation`, `connectComposerLocation`, `promoteLocation`, `promoteAllLocation`, `createSubject` call, `.mutatedSourceGraph` / `.deletedSubject` applies | Source-keyed |
| Composer | `macos/App/Features/CitationComposer/CitationComposerEntry.swift` (`sourceID` on each case, `identityKey`, `init?(location:)`); `CitationComposerContext.swift` (`sourceID`, `graphKey`, `memoSnapshot`); `CitationComposerModel.swift` (`graphKey`, `graphLocation()`, `fetchOpenCitation` → `store.getCitation`, `createSubject`); `CitationConnections.swift` and `CitationComposerModel+Authoring.swift` (`createCitedBridge(sourceID:)`) | Source-keyed |
| Promote | `macos/App/Features/Promote/PromoteModel.swift` `PromoteEntry` (`sourceID`, `identityKey`, `graphLocation`), `WorkspaceLocation.promote(sourceId:…)`, `graphKey`, `.promotedSubject(sourceId:)`; `PromoteView.swift` `QueryHandle<SourceGraphRows>` | Promote's rows read `.sourceGraph(sourceId)` |
| Entry points | `macos/App/Features/Sources/SourcesListNavigation.swift` `graphLocation(for:)` (needs `source.hasArtifact`); `SourcesListView.swift` (two `onOpenGraph` closures, `SourcesSplitRowProgress` reads `handle.value?[source.id]`); `SourcePageIdentityHeader.swift` `evidenceGraphJump`; `SourcesModel.swift` `graphProgress(for:)`, `progressHandle` | Source-keyed; no graph list read |
| Reconciliation record | `macos/App/Features/Conclusions/ReconciliationRecord.swift` `composerLocation(for:)`; `CatalogReconcilerOutcome` in `macos/App/Platform/GenealogyStore.swift` | No graph |
| Store | `macos/App/Platform/GenealogyStore.swift` (protocol, `SourceGraphProgress`, `CatalogEvidenceGraph`), `GoStore.swift`, `FakeStore.swift`: `listSubjects`, `listSubjectPositions`, `listObservationsBySource`, `listSubjectMemberships`, `listSourceEventTitles`, `listSourceGraphProgress`, `getSourceGraphProgress`, `createSubject`, `createCitedBridge`, `getCitation`, `mapWriteEffects`; FakeStore `subjectsBySource`, `observationsBySource`, `eventTitlesBySource`, `graphProgressBySource`, `bumpSource`, `createSource` (seeds no graph), `copySource` | Source-keyed |
| Go tests | `api/ffi/handlers/subjects_test.go` `TestListSubjects`; `sourcegraphprogress_test.go` `TestGetSourceGraphProgress`; `source_write_effects_test.go`; `citations_test.go` (`GetSources()` at line 358); `promote_test.go` (`GetSources()` at line 161); `connect_test.go`; `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve`; `registry_completeness_test.go` `TestSourceEvidenceIsNotActivity`; `core/database/graphfill/load_test.go` `TestSourceWriteDropsEveryGraph`, `TestRetitleDropsEvidenceGraph`, `TestDeletedGraphIsNotFound`; `core/database/deleteimpact/impact_test.go` (location asserts at lines 279, 748, 1048) | Source-keyed |
| Swift tests | `WorkspaceNavigationTests`, `WorkspaceToolbarBreadcrumbTests`, `PlaceRegistryTests`, `WorkspaceDestinationHostTests`, `SourcesListNavigationTests`, `CatalogQueryRegistryTests`, `WorkspaceSessionTests`, `EvidenceGraphModelTests`, `CitationComposerModelTests`, `PromoteModelTests`, `SourcesModelTests`, `EvidenceGraphStoreTests`, `PersonDetailTests` (reconciliation `composerLocation(for:)`), `SourceGraphSnapshotTests` (line 465 comment "sourceeventtitles") | Seed and assert by Source id |
| Docs | `docs/macos-client-patterns.md` (lines 91, 112); `.agents/skills/add-workspace-place/SKILL.md` (lines 71, 172–173); `.agents/skills/add-catalog-delete/SKILL.md` (line 111); `core/database/effects/README.md`; `docs/ideas/shared-list-helpers.md` (line 7) | Name `sourceGraph`, `SourceGraphRows`, `.mutatedSourceGraph(sourceId:)`, `Set.Sources`, `refreshSourceGraphProgress` |

**Precedent.** `de0b4411` (S12-03, "Report every catalog-wide list a write changed as one declared set on the effects, the wire and Swift") replaced an effects field across Go, the wire and Swift in one commit: it reserved the old proto numbers and added a new field instead of reusing a number. Copy that for `WriteEffects.sources` and every request field below. `7476f964` (S12-07) re-keyed the graph cache and every caller in one commit because deleting the old method breaks every caller at once; STEP-6 to STEP-8 here are the Swift equivalent.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Proto, one edit per message, old numbers reserved as `de0b4411` did. **Methods** (numbers kept): `METHOD_LIST_OBSERVATIONS_BY_EVIDENCE_GRAPH = 69`, `METHOD_LIST_EVIDENCE_GRAPH_PROGRESS = 79`, `METHOD_GET_EVIDENCE_GRAPH_PROGRESS = 80`, `METHOD_LIST_EVIDENCE_GRAPH_EVENT_TITLES = 96`. **Messages renamed:** `ListObservationsBySourceRequest/Response` → `ListObservationsByEvidenceGraphRequest/Response`, `ListSourceEventTitlesRequest/Response` → `ListEvidenceGraphEventTitlesRequest/Response`, `SourceGraphProgress` → `EvidenceGraphProgress`, `ListSourceGraphProgressRequest/Response` → `ListEvidenceGraphProgressRequest/Response`, `GetSourceGraphProgressRequest/Response` → `GetEvidenceGraphProgressRequest/Response`. **Graph id replaces Source id:** `ListSubjectsRequest`, `ListSubjectPositionsRequest`, `ListObservationsByEvidenceGraphRequest`, `ListSubjectMembershipsRequest`, `ListEvidenceGraphEventTitlesRequest`, `GetEvidenceGraphProgressRequest` each get `reserved 2; reserved "source_id"; string graph_id = 3;`. `CreateSubjectRequest`: `reserved 3; reserved "source_id"; string graph_id = 10;`. `CreateCitedBridgeRequest`: its existing statements become `reserved 3, 7, 9, 10;` and `reserved "source_id", "label", "grid_x", "grid_y";`. **Added:** `GetCitationRequest.graph_id = 3`; `EvidenceGraphProgress.graph_id = 7` (keeps `source_id = 1`, "the graph's Source"); `WriteEffects`: its existing `reserved 3, 4, 12;` becomes `reserved 2, 3, 4, 12;`, plus `reserved "sources";` and `repeated string evidence_graphs = 14;` ("Evidence graphs whose loaded rows this write changed: what the graph's rows and graph progress read"); `WorkspaceLocation.graph_id = 19` (comment: "the Evidence graph of a graph, citationComposer or promote surface"); `ReconcilerOutcomeDetail.graph_id = 20`. Regenerate with `./scripts/generate-proto.sh` in each step that edits the proto | Working rules "replaces Source-keyed fields (effects, query keys, wire fields) instead of adding graph ones beside them" and "A re-keyed symbol is renamed": an RPC named for a Source that reads a graph is renamed; `ListSubjects` and friends already name what they list. A reused field number would decode an old Source id as a graph id | Keeping `source_id` beside `graph_id` — the churn rule; renaming the field in place (`source_id = 2` → `graph_id = 2`) — the wire would silently accept a Source id; renaming `ListSubjects` / `ListSubjectPositions` / `ListSubjectMemberships` — their names do not say Source |
| D-2 | Go handler and dispatch names follow D-1: `ListObservationsBySource` → `ListObservationsByEvidenceGraph`, `ListSourceEventTitles` → `ListEvidenceGraphEventTitles`, `ListSourceGraphProgress` → `ListEvidenceGraphProgress`, `GetSourceGraphProgress` → `GetEvidenceGraphProgress`, the four `Method…` constants and `case` lines in `api/ffi/dispatch.go`; `unmarshalErr` tags `list_observations_by_evidence_graph`, `list_evidence_graph_event_titles`, `list_evidence_graph_progress`, `get_evidence_graph_progress`. Files: `api/ffi/handlers/sourcegraphprogress.go` → `evidence_graph_progress.go`, `sourcegraphprogress_test.go` → `evidence_graph_progress_test.go` (`git mv`). Each graph read handler parses `graph_id` with `parseID` and reads `c.Graph().EvidenceGraph(graphID)` (or `subjectpositions.ListByEvidenceGraph`, `graphprogress.Get`). `GetCitation` reads `c.Graph().EvidenceGraph(graphID)` from the request and drops its `citations.SourceID` call. `protoProgress` writes `GraphId` and `SourceId` from the row. `sourceGraphID`, `sourceEvidenceGraph`, `citations.SourceID` and its query `sqlCitationSource` are deleted | One name per RPC end to end. An unknown graph id loads `Found` false with no rows (S12-07 D-1), so AC-3.2 holds with no special case. The composer always knows its graph, so `GetCitation` needs no query to find it | Deriving `GetCitation`'s graph from its first Observation — a query the caller can skip, and no graph at all for a Citation with no Observation |
| D-3 | `evidencegraphs.SourceOf(q database.Querier, graphID []byte) ([]byte, error)` *(new)*, beside `SubjectGraphTx`: `SELECT source_id FROM evidence_graphs WHERE id = ?`; `ErrInvalid` for a nil querier, a malformed id or no row. `graphprogress.Get` calls it (treating `ErrInvalid` as no Source) and drops `sqlGraphSource`. `subjects.InsertTx`: when `GraphID` has 16 bytes and `SourceID` is empty, it reads the Source with `SourceOf(tx, in.GraphID)` (`ErrInvalid` → `subjects.ErrInvalid`); a non-empty `SourceID` that is not the graph's Source is still refused by the composite FK (S12-06 D-5); an empty `GraphID` keeps the `SoleGraphTx` fallback until S12-18. `Create`'s 16-byte `SourceID` check moves into `InsertTx`'s new rule (one of the two ids must be set). `CreateSubject` parses `graph_id` and sends `subjects.CreateInput{GraphID: graphID, …}` with no `SourceID`; an unknown graph reaches Swift as `subjects.invalid` | The rule "a Subject's Source is its graph's Source" lives in `subjects`, so S12-18 only deletes the fallback and the field (review M-5). One graph-to-Source read for `subjects` and `graphprogress` (review M-9); `deleteimpact` keeps its own two because `evidencegraphs` imports it | The handler deriving the Source — written to be moved one PR later; `SourceTx(*sql.Tx)` — `graphprogress.Get` reads on a `*sql.DB` |
| D-4 | `effects.Effect.Source` → `Effect.EvidenceGraph` and `Set.Sources` → `Set.EvidenceGraphs`: a `Path` that ends at `evidence_graphs`. `Resolve`, `paths()`, `writeEffectsProto` (`EvidenceGraphs: idStrings(set.EvidenceGraphs)`) follow. `graphcache.dropGraphsOfSources` → `dropEvidenceGraphs(ids [][]byte)`: `delete(g.evidenceGraphs, string(id))` and `delete(g.memos, string(id))` for each 16-byte id; `OnCommit` calls it with `set.EvidenceGraphs`. Doc comments: `Set.EvidenceGraphs` "the Evidence graphs whose loaded rows changed: graphcache drops each, and graph progress counts them"; `EvidenceGraphMemo` "a commit that names the graph" | Working rule "Effects are graph-keyed, in one step"; S12-07 D-4 named this delete by graph id | Keeping the Source scan beside the graph delete — two drops for one fact |
| D-5 | Registry paths (`core/database/effects/registry.go`), with helpers *(new)* beside `valueSource` and `observationMembers`: `var observationGraph = up("subject_id", "graph_id")`; `graphsOfSource(p Path) Path` = `chain(p, inbound("evidence_graphs", "source_id"))`; `valueGraph(col string) Path` = `chain(inbound("observations", col), observationGraph)`. Entries: `sources` `EvidenceGraph: onField("title", graphsOfSource(self()))`; `artifacts` `graphsOfSource(field("source_id"))`; `citations` `graphsOfSource(up("artifact_id", "source_id"))`; `observations` `observationGraph`; `subjects` `field("graph_id")`; `evidence_graphs` `self()`; `identity_claims` `from("subject_id", field("graph_id"))`; `identity_claim_evidence` `up("identity_claim_id", "subject_id", "graph_id")`; `ownedValue` sets `EvidenceGraph: valueGraph(col)` and keeps `valueSource` for its other fields. `evidence_graph_artifacts` stays without one. `CitationSource`, `SourceScope`, `Handles`, `HandleScope`, `Search`, `Lists` and `Structure` are unchanged | Each entry names the graphs whose cached rows (Subjects, Observations, members, the Source title) the change moves. Artifact and Citation writes keep today's reach — every graph of the Source — because a cached graph holds neither and `main` reloads the whole Source for them; narrowing them is a behavior change no PR asked for. An Observation, its values and a claim live on their Subject's graph (S12-06) | `chain(inbound("evidence_graph_artifacts", "artifact_id"), field("graph_id"))` for artifacts — misses the cascaded link rows on an Artifact delete; `observations` through the Citation's Artifact — every linked graph instead of the one that holds the Observation |
| D-6 | Handlers keep their empty-on-unknown behavior with no helper: `ListSubjectPositions` lists `ListByEvidenceGraph` directly (an unknown graph has no rows); `GetEvidenceGraphProgress` returns `graphprogress.Get`'s result as is (it already sets `GraphID` and leaves `SourceID` empty for an unknown graph) | S12-07 AC-3.2 behavior, now with no bridge | Returning an error for an unknown graph — a behavior change |
| D-7 | Swift store signatures (`GenealogyStore`, `GoStore`, `FakeStore`): `listSubjects(projectDir:graphID:)`, `listSubjectPositions(projectDir:graphID:)`, `listObservationsByEvidenceGraph(projectDir:graphID:)` *(renamed)*, `listSubjectMemberships(projectDir:graphID:)`, `listEvidenceGraphEventTitles(projectDir:graphID:)` *(renamed)*, `listEvidenceGraphProgress(projectDir:)` *(renamed)*, `getEvidenceGraphProgress(projectDir:graphID:)` *(renamed)*, `createSubject(projectDir:userID:graphID:subjectTypeID:label:description:placement:)`, `createCitedBridge(…)` without `sourceID:`, `getCitation(projectDir:graphID:citationID:)`. `SourceGraphProgress` → `EvidenceGraphProgress` with `id` = `graphId`, `graphId`, `sourceId`, the counts, `zeros(graphId:)`. `CatalogWriteEffects.sources` → `evidenceGraphs`; `mapWriteEffects` reads `effects.evidenceGraphs`. `CatalogReconcilerOutcome.graphID` *(new)*. The FakeStore counters and error hooks named for the renamed methods are renamed to match | The Swift side of D-1, one name per RPC | — |
| D-8 | Swift renames with the re-key: `SourceGraphRows` → `EvidenceGraphRows` (`sourceId` → `graphId`), `SourceGraphSnapshot` → `EvidenceGraphSnapshot` (`sourceId` → `graphId`), `SourceGraphPlacedSubject` → `EvidenceGraphPlacedSubject`, `SourceGraphPlacedBridge` → `EvidenceGraphPlacedBridge`, `SourceGraphProgress` → `EvidenceGraphProgress` (D-7); files `SourceGraphSnapshot.swift` → `EvidenceGraphSnapshot.swift` and `macos/ProvenenciaTests/SourceGraphSnapshotTests.swift` → `EvidenceGraphSnapshotTests.swift` (`git mv`, then the four `project.pbxproj` references of each). Keys: `.sourceGraph(project:sourceId:)` → `.evidenceGraph(project:graphId:)`, `.sourceGraphProgress(project:)` → `.evidenceGraphProgress(project:)` (map keyed by graph id); `Kind` cases to match. `PlaceID.sourceGraph` and `WorkspacePresentationID.sourceGraph` keep their names (they follow the section prefix, like `sourcePromote`). **Mutations** (review M-1): `.mutatedSourceGraph` and its `CatalogMutationKind` case are deleted; their three call sites apply `session.apply(effects)` alone, because each write's effects name the Subject's graph (D-5 `subjects`). `.deletedSubject(sourceId:)` keeps its shape and leaves the graph key's `invalidateOn` (it still reloads `.citationCounts(sourceId)`, `.propertiesWorkspace` and `.citationsByArtifact`). `.promotedSubject(sourceId:)` → `.promotedSubject(graphId:)` (Promote's apply returns no effects). `CatalogMutation.evidenceGraphId: String?` *(new)* returns the graph of `.promotedSubject` and nil otherwise; the registry's `.evidenceGraph` case and the session read it. The `.evidenceGraph` spec's `invalidateOn` is `[.deletedSource, .promotedSubject]`: `.promotedSubject` → `.key(.evidenceGraph(graphId:))`, `.deletedSource` → `.allCached(.evidenceGraph)`. `invalidations(for effects:)`: one `.key(.evidenceGraph(graphId:))` per `effects.evidenceGraphs`. Session `apply`: progress graphs are `effects.evidenceGraphs` plus `mutation?.evidenceGraphId`; `graphProgressSource(of:)` is deleted; `refreshSourceGraphProgress` / `patchSourceGraphProgress` → `refreshEvidenceGraphProgress(graphId:)` / `patchEvidenceGraphProgress(graphId:)`; `graphCameras`, `graphCamera(for graphID:)`, `setGraphCamera(_:for graphID:)`, `pendingGraphSelections`, `setPendingGraphSelection(subjectID:graphID:)`, `consumePendingGraphSelection(graphID:)` keyed by graph id. Change the comment at the test file's line 465 from "sourceeventtitles" to "grapheventtitles" | Working rule "A re-keyed symbol is renamed" names `CatalogQueryKey.evidenceGraph(graphId:)`. One mechanism moves the graph cache: the effects (review M-1). One mutation-to-graph read instead of two copies ([`shared-list-helpers.md`](../../../ideas/shared-list-helpers.md) § "Source graph mutations", review M-4) | `.mutatedEvidenceGraph(graphId:)` — repeats what the effects say; `.deletedSubject(sourceId:graphId:)` — a two-id case for a reload the effects already do |
| D-9 | Promote's location, entry and key move here; its RPCs do not. `PromoteEntry` gains `graphID`; `WorkspaceLocation.promote(sourceId:graphId:subjectId:kind:ref:title:sourceTitle:mapAll:)`; `identityKey` becomes `"promote|\(graphID)|…"` (the rest as today); `graphKey` = `.evidenceGraph(graphId: entry.graphID)`; `graphLocation` carries `graphId`; `PromoteModel` applies `.promotedSubject(graphId: entry.graphID)`. `proposePromoteGraphAlignment`, `getPromoteClaimEvidence`, `applyPromoteGraphAlignment` keep `sourceID: entry.sourceID` | User, 2026-10-22: Promote's rows read the canvas key, so the key cannot move without it; the location carries the graph so S12-09 only changes the RPCs and Go | Promote finding its graph in the graph list until S12-09 — an interim adapter |
| D-10 | `WorkspaceLocation.graphId: String?` *(new)*: in `CodingKeys`, the memberwise `init`, `init(from:)` (`decodeIfPresent`, `nilIfEmpty`) and `==`; mapped both ways in `GoStore` (`p.graphID` / `loc.graphID`). Factories *(new)* beside `WorkspaceLocation.personDetail`: `static func evidenceGraph(sourceId: String, graphId: String, ref: String? = nil, title: String? = nil, sourceTitle: String? = nil)` and `static func citationComposer(sourceId: String, graphId: String, subjectId: String? = nil, citationId: String? = nil, artifactId: String? = nil, observationId: String? = nil, connectFromSubjectId: String? = nil, connectToSubjectId: String? = nil, connectBridgeTypeKey: String? = nil, ref: String? = nil, title: String? = nil, sourceTitle: String? = nil)`; every app and FakeStore literal with `sourceSurface: .graph` or `.citationComposer` moves onto them (review M-3: `SourcesListNavigation`, `WorkspaceToolbar`, `CitationComposerModel.graphLocation()`, `PromoteEntry.graphLocation`, `EvidenceGraphModel` composer locations, `ReconciliationRecord`, `PVDeleteImpact` preview, FakeStore fixtures); `mapWorkspaceLocationFromProto` keeps the memberwise `init`. `PlaceRegistry`: `.sourcePromote`, `.sourceCitationComposer` and `.sourceGraph` keys read `.evidenceGraph(graphId:)` (each `guard let graphId`) and keep their Source keys (`.sourceWorkspace(sourceId)`, `.citationCounts(sourceId)`); `.sourceGraph`'s `deepId` is `graphId`; their `matches` do not change (D-11 guarantees a graph). `.sourcesList` warms `.evidenceGraphsList` and `.evidenceGraphProgress`; `.sourceDetail` adds `.evidenceGraphsList`. `WorkspaceDestinationHost` passes `graphID` and `sourceID` to `EvidenceGraphView`. Breadcrumbs: the graph crumb's id is `graph-\(graphID)` and its target is `.evidenceGraph(…)`; the composer and Promote leaf ids use the graph id; labels do not change. Go: `deleteimpact.Location.GraphID` *(new)*: `projectSubject` reads `subjects.graph_id`; the observation projector and `projectCitation`'s first-Observation query select `sub.graph_id`; for a Citation with no Observation, `projectCitation` takes the first graph its Artifact is linked to (`SELECT ega.graph_id FROM evidence_graph_artifacts ega JOIN evidence_graphs g ON g.id = ega.graph_id WHERE ega.artifact_id = ? ORDER BY g.name COLLATE NOCASE, g.id LIMIT 1`), so the link still opens the composer as on `main`; the `sql.ErrNoRows` fallbacks (row already gone) carry no Source today and stay as they are; `projectEvidenceGraph` returns `SourceSurface: surfaceGraph` with `SourceID` and `GraphID` (its comment "not a place of its own until S12-08" goes, review M-7). `locationFromImpact` maps it. `conclusiondetails.Outcome.GraphID` *(new)* from `s.graph_id`; `reconcilerOutcomeProto` maps it | Working rule "A graph is a place"; the location keeps `sourceId` because the Source page, citation counts and Promote's RPCs read it. Factories with a non-optional `graphId` make the compiler find every site, and S12-11, S12-14, S12-15 each call one | A location with only `graphId` — every Source read would first look up the graph's Source; hand-built literals — the next three PRs repeat them |
| D-11 | One normalization in the memberwise `WorkspaceLocation.init`: a `graph`, `citationComposer` or `promote` surface whose `graphId` is nil after `nilIfEmpty` becomes `sourceSurface = .page`, keeping its `sourceId`. a private `static func normalizedSurface(_ surface: SourceSurface, graphId: String?) -> SourceSurface` *(new)*, beside `nilIfEmpty`, holds the rule; the memberwise `init` and `init(from:)` both call it, so history decoding, `mapWorkspaceLocationFromProto` and Swift literals all get it | The same rule that decodes a missing `sourceSurface` as `.page`, written once for every way a location is built (review M-2); the place registry never sees a graph surface without a graph | Encoding it in `init(from:)`, the proto mapper and the place registry separately — three copies of one fact; resolving the Source's first graph on restore — a Source-to-graph lookup in navigation |
| D-12 | The two Source-level entry points open the Source's **first graph in the graph list** (user, 2026-10-22). `SourcesListNavigation.entryGraph(sourceID:in graphs: [CatalogEvidenceGraph]) -> CatalogEvidenceGraph?` *(new)* returns the first `graphs` element whose `sourceID` matches (the list is sorted by name); `graphLocation(for source:graphs:) -> WorkspaceLocation?` returns nil when there is no entry graph or its `artifactIDs` is empty, else a `.graph` location with `sourceId`, `graphId`, `ref` and `title` as today. `SourcesModel` and `SourcePageModel` read `.evidenceGraphsList` (`session.query`) and pass its value; `SourcesSplitRowProgress` reads `handle.value?[entryGraph.id]`; `SourcesModel.graphProgress(for:)` takes the graph id. Accessors *(new)*: `SourcesModel.graphLocation(for source: CatalogSource) -> WorkspaceLocation?` and `SourcePageModel.evidenceGraphLocation: WorkspaceLocation?`, each calling `SourcesListNavigation.graphLocation(for:graphs:)` with the graph list's value or `[]`. `SourcesSplitRow`'s `onOpenGraph: () -> Void` becomes `onOpenGraph: (() -> Void)?`: the zone is still shown only when `source.hasArtifact`, and is disabled when `onOpenGraph` is nil; both `SourcesListView` call sites pass `model.graphLocation(for: source).map { location in { navigation.go(to: location) } }`. `SourcePageIdentityHeader.evidenceGraphJump` reads `model.evidenceGraphLocation`. While the graph list has no value (loading or failed), both are disabled, as for a Source with no Artifact; the list is warm from the sidebar and list places in practice | One function owns the interim choice; S12-11 and S12-14 replace both entry points with per-graph Open. "Open needs a linked Artifact" is the working rule's test, and equals `hasArtifact` while the Default graph links every Artifact | The graph named `Default` — the name is editable; `source.hasArtifact` — a Source fact, not the graph's |
| D-13 | Canvas and composer carry both ids. `EvidenceGraphView(graphID:sourceID:session:store:userID:)`, `EvidenceGraphModel.graphID` *(new)* beside `sourceID`; `graphKey`, `displaySnapshot`, camera, pending selection, `createSubject`, the applied mutations and every location it builds use `graphID`; `sourceRow`, `workspaceKey`, `citationCountsKey`, `sourcePageLocation` keep `sourceID`. `CitationComposerEntry` cases gain `graphID`; `init?(location:)` also needs `location.graphId`; `identityKey` uses the graph id where it uses the Source id today; `CitationComposerContext.graphID` *(new)* serves `graphKey`, `memoSnapshot`, `waitForGraph` and `EvidenceGraphRows`; `getCitation` and `createSubject` send it; `createCitedBridge` sends no Source. `ReconciliationRecord.composerLocation(for:)` sets `graphId: outcome.graphID` | One graph per surface; Source reads stay on the Source | — |
| D-14 | FakeStore mirrors the engine. `createSource` and `copySource` append a `Default` `CatalogEvidenceGraph` (fresh UUID id, `GPH-` ref, the copy's Artifacts linked). `subjectsBySource`, `observationsBySource`, `eventTitlesBySource`, `graphProgressBySource` → `subjectsByGraph`, `observationsByGraph`, `eventTitlesByGraph`, `graphProgressByGraph`. A graph's Source is read through one private helper `sourceOfGraph(_ graphID: String) -> String?` *(new)* that scans every project's list in `evidenceGraphsByProject`. Which graph a write names: a Subject write (`createSubject`, `updateSubject`, `deleteSubject`, `unpromoteSubject`) — the Subject's graph; an Observation write (`addObservationsToCitation`, `updateObservation`, `deleteObservation`, and the Observations of `createCitation` / `createCitedBridge`) — its Subject's graph, found by scanning `subjectsByGraph`; `createCitedBridge` finds its endpoints in `subjectsByGraph` and stores the bridge on the from-Subject's graph; an Artifact or Citation-only write — every graph of the Source. `bumpSource(_:evidence:)` becomes `bumpSource(_ sourceID: String?, graphs: [String])`: `evidenceGraphs` = `graphs`, `citationSources` = [sourceID] when `graphs` is not empty, `sourceScopes` = [sourceID]; writes that passed `evidence: false` pass `[]`. `getCitation` reads `observationsByGraph[graphID]`. FakeStore delete-impact and search fixtures build their locations with the D-10 factories and the referenced Subject's graph. Tests seed a graph with a `CatalogEvidenceGraph` literal in `evidenceGraphsByProject` and use graph ids that differ from Source ids (`"g-1"` for `"src-1"`). Tests whose expectations count graphs after `createSource` or `copySource` (`EvidenceGraphStoreTests`, `SourceCopyModelTests`, `OnboardingModelTests`, `CatalogQueryRegistryTests`) may change those counts by the Default graph and nothing else | The fake must answer graph-keyed calls and report the same graphs the engine reports (AC-2.1), so Swift tests can see a per-graph effect; distinct ids catch a Source-for-graph mix-up, as S12-03 D-2 did in Go | Reusing the Source id as the graph id in fixtures — hides the very bug this PR could introduce; every write reporting every graph of its Source — Swift tests could not tell a per-graph effect from a per-Source one |
| D-15 | `evidencegraphstest.SoleGraph` stays: a test fixture over `List`, used by tests that build one-graph Sources. `evidencegraphs.SoleGraph` and `SoleGraphTx` stay for `promotealign.loadInputs` and `InsertTx`'s empty-`GraphID` fallback (S12-09, S12-18). The map's working rule "One bridge, one function" already says so (reworded when this plan was written) | User, 2026-10-22 (split); a test helper resolves nothing in the product | Deleting the fixture — 50 test call sites move onto `List(...)[0]`, the same code inline |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `evidencegraphs.SourceOf` *(new)* | func | `evidencegraphs.SubjectGraphTx`; `graphprogress.sqlGraphSource`; `deleteimpact` `projectEvidenceGraph` and `register.go` graph reads | Same shape as `SubjectGraphTx`; replaces `graphprogress.sqlGraphSource`; `deleteimpact` keeps its two because `evidencegraphs` imports `deleteimpact` (D-3) |
| `effects.observationGraph`, `graphsOfSource`, `valueGraph` *(new)* | var / func | `observationMembers`, `valueSource`, `ownedValue` | Same idea for the graph path; composes existing combinators (D-5) |
| `Effect.EvidenceGraph`, `Set.EvidenceGraphs` *(renamed)* | field | `Effect.Source`, `Set.Sources` | Replaces them (D-4) |
| `graphcache.dropEvidenceGraphs` *(renamed)* | func | `dropGraphsOfSources` | Replaces it with a delete by key (D-4) |
| `deleteimpact.Location.GraphID`, `conclusiondetails.Outcome.GraphID` *(new)* | field | `Location.SourceID`, `Outcome.SourceID` | Extends the same rows and queries (D-10) |
| Proto `graph_id` fields, `evidence_graphs`, renamed RPCs *(new / renamed)* | FFI | the `source_id` fields they replace | D-1 |
| `WorkspaceLocation.graphId` *(new)* | field | `sourceId` | D-10 |
| `CatalogQueryKey.evidenceGraph`, `.evidenceGraphProgress` *(renamed)* | key | `.sourceGraph`, `.sourceGraphProgress` | Replaces them (D-8) |
| `EvidenceGraphRows`, `EvidenceGraphSnapshot`, `EvidenceGraphPlacedSubject`, `EvidenceGraphPlacedBridge`, `EvidenceGraphProgress` *(renamed)* | type | the `SourceGraph*` types | Renamed only (D-7, D-8) |
| `SourcesListNavigation.entryGraph` *(new)* | func | `SourcesListNavigation.graphLocation` | The one interim "which graph" rule (D-12) |
| `WorkspaceLocation.evidenceGraph(…)`, `.citationComposer(…)` *(new)* | func | `WorkspaceLocation.personDetail`, `WorkspaceLocation.promote(…)` | Same factory pattern; non-optional `graphId` (D-10) |
| `CatalogMutation.evidenceGraphId` *(new)* | property | `CatalogMutation.patchedSource`, `WorkspaceSession.graphProgressSource(of:)` | Replaces the session's switch and the registry's case list (D-8) |
| FakeStore `sourceOfGraph(_:)` *(new)* | helper | FakeStore `owningSource(ofCitation:)` | The fake's one graph-to-Source read (D-14) |
| `EvidenceGraphModel.graphID`, `CitationComposerContext.graphID`, `PromoteEntry.graphID`, `CitationComposerEntry` `graphID`, `CatalogReconcilerOutcome.graphID` *(new)* | field | each type's `sourceID` | The surface's graph (D-9, D-13) |

**Builds on.** `effects` path combinators and their init-time edge check; the S12-07 graph-keyed reads (`c.Graph().EvidenceGraph`, `ListByEvidenceGraph`, `graphprogress.Get`); the S12-05 graph list `.evidenceGraphsList`; the `WorkspaceLocation` legacy-decode pattern (`sourceSurface` → `.page`); `de0b4411`'s reserve-and-add proto pattern.
**Deletes.** `sourceGraphID`, `sourceEvidenceGraph`, `citations.SourceID`, `citations.sqlCitationSource`; `graphprogress.sqlGraphSource`; `dropGraphsOfSources`; `.mutatedSourceGraph` and its kind; `WorkspaceSession.graphProgressSource(of:)`; the `docs/ideas/shared-list-helpers.md` section "Source graph mutations"; `Effect.Source`, `Set.Sources`, `WriteEffects.sources`, `CatalogWriteEffects.sources`; every `source_id` on the six graph read requests, `CreateSubjectRequest` and `CreateCitedBridgeRequest`; the `SourceGraph*` Swift names, `.sourceGraph`, `.sourceGraphProgress`, `.mutatedSourceGraph`; the comment in `api/ffi/handlers/connect.go` about the unread `source_id`; the S12-08 comment on `deleteimpact` `projectEvidenceGraph`.
**Next adds.** S12-09 changes Promote's three RPCs to `entry.graphID` and deletes `SoleGraph` / `SoleGraphTx`. S12-18 deletes `InsertTx`'s empty-`GraphID` fallback and `CreateInput.SourceID`. S12-09 retires `.promotedSubject` the way this PR retires `.mutatedSourceGraph` once Promote's apply returns effects. S12-10 filters the composer by the graph's `artifactIDs` from `.evidenceGraphsList` and `CitationComposerContext.graphID`. S12-11 and S12-14 build per-graph locations with `WorkspaceLocation.evidenceGraph(…)` and delete `entryGraph`. S12-14 reads `.evidenceGraphProgress` per graph. S12-15 maps a graph hit to the same factory's shape.

### Schema

None — no catalog change. `subjects.graph_id` and its index exist (migration 55).

### Go core

| Package | Change |
| --- | --- |
| `core/database/evidencegraphs` | `func SourceOf(q database.Querier, graphID []byte) ([]byte, error)` *(new)*, D-3 |
| `core/database/subjects` | `InsertTx` reads the graph's Source when `SourceID` is empty (D-3) |
| `core/database/graphprogress` | `Get` calls `evidencegraphs.SourceOf`; `sqlGraphSource` deleted (D-3) |
| `core/database/effects` | D-4, D-5 |
| `core/database/graphcache` | D-4 |
| `core/database/citations` | `SourceID` deleted (D-2) |
| `core/database/deleteimpact` | `Location.GraphID`, three projectors (D-10) |
| `core/database/conclusiondetails` | `Outcome.GraphID`, `sqlOutcomes` selects `s.graph_id` after `s.ref` (D-10) |

Errors: none new. `SourceOf` returns `evidencegraphs.ErrInvalid`; `InsertTx` maps it to `subjects.ErrInvalid` (`subjects.invalid`).

### FFI

`api/proto/engine.proto` per D-1; handlers per D-2, D-3, D-6, D-10. Regenerate with `./scripts/generate-proto.sh` (it writes `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`). Sketch of one request:

```proto
message ListSubjectsRequest {
  string project_dir = 1;
  reserved 2;
  reserved "source_id";
  string graph_id = 3;
}
```

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Graph rows key | `CatalogQueryKey.evidenceGraph(project:graphId:)` | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` | D-8 |
| Progress map | `CatalogQueryKey.evidenceGraphProgress(project:)` → `[String: EvidenceGraphProgress]` by graph id | same | D-8 |
| Place | `WorkspaceLocation.graphId` | `macos/App/Features/Workspace/WorkspaceLocation.swift` | D-10, D-11 |
| Entry graph | `SourcesListNavigation.entryGraph`, `graphLocation(for:graphs:)` | `macos/App/Features/Sources/SourcesListNavigation.swift` | D-12 |

No view changes layout. States: a graph id no graph has reads as an empty canvas (as an unknown Source does today); a location without a graph id opens the Source page (D-11). Accessibility: unchanged identifiers.

### Strings

None — no copy.

## 7. Steps

Go builds and its tests pass at the end of STEP-1 to STEP-4. The Xcode build breaks at STEP-2 (the regenerated `engine.pb.swift` drops `sources`) and builds again at the end of STEP-8; STEP-5 to STEP-8 are one Swift change made in order, committed per step. Run the Swift check only from STEP-8.

**STEP-1 — A graph's Source**
Files: `core/database/evidencegraphs/evidencegraphs.go`, `evidencegraphs_test.go`, `core/database/subjects/subjects.go`, `subjects_test.go`, `core/database/graphprogress/graphprogress.go`. Skill: `add-catalog-query`.
Do: D-3: `SourceOf` under `SubjectGraphTx`; `InsertTx`'s new rule; `graphprogress.Get` onto `SourceOf`. Add T-1 and T-22.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/evidencegraphs/ ./core/database/subjects/ ./core/database/graphprogress/` passes.

**STEP-2 — Effects name graphs, on the engine and the wire**
Files: `core/database/effects/{registry.go, resolve.go, README.md, evidence_graphs_test.go, registry_completeness_test.go}`, `core/database/graphcache/{evidence_graph.go, commit.go}`, `core/database/graphfill/load_test.go`, `api/proto/engine.proto` (`WriteEffects` only), `api/ffi/handlers/{write_effects.go, write_effects_test.go, source_write_effects_test.go, citations_test.go, promote_test.go}`, and any other test that reads `Set.Sources` or `GetSources()` on effects (`rg -n 'Set\.Sources|\.Sources\b|GetSources\(\)' core api --type go` lists them; `list.Sources` / `before.Sources` on `ListSourcesResponse` are not effects and stay).
Do: D-4, D-5, D-1's `WriteEffects` edit; regenerate. Rewrite the `Set.Sources` comment at `core/database/graphcache/evidence_graph.go` line 126 with D-4. Rename `TestSourceEvidenceIsNotActivity`'s field check to `EvidenceGraph` and add T-2; rewrite `TestEvidenceGraphsResolve` (T-3); `TestSourceWriteDropsEveryGraph` → `TestGraphWriteDropsOnlyThatGraph` (T-4); `source_write_effects_test.go`'s `wantSources` → `wantGraphs` with the AC-2.3 expectations (T-5). README: `Set.EvidenceGraphs` / `Effect.EvidenceGraph` "the Evidence graphs whose loaded rows a change moves (graphcache drops each; graph progress counts it)", graph rows name themselves, link rows name none, artifact and citation rows name every graph of their Source.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./api/... ./core/...` is clean; `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/ ./core/database/graphcache/ ./core/database/graphfill/ ./api/ffi/handlers/ ./core/writes/` passes; `rg -n 'Set\.Sources|eff\.Source\b|Effect\.Source\b|dropGraphsOfSources|GetEffects\(\)\.GetSources\(\)|eff\.GetSources\(\)' core api --type go` prints nothing (the generated `GetSources()` of `ListSourcesResponse`, `GetWorkspaceNavCountsResponse` and `GetPromoteClaimEvidenceResponse` stay).

**STEP-3 — Graph RPCs take a graph id**
Files: `api/proto/engine.proto`, `api/ffi/dispatch.go`, `api/ffi/handlers/{subjects.go, observations.go, promote.go, event_titles.go, citations.go, connect.go, source_read.go}`, `api/ffi/handlers/sourcegraphprogress.go` → `evidence_graph_progress.go` and its test (`git mv`), `core/database/citations/citations.go` (and the test of `SourceID` if one exists: `rg -n 'citations\.SourceID|SourceID\(c,' core/database/citations`), the handler tests that send these requests (`subjects_test.go`, `citations_test.go`, `connect_test.go`, `promote_test.go`, `promote_batch_test.go`, `promote_evidence_test.go`, `conclusion_headers_test.go`, `evidence_graph_progress_test.go`). Skills: `add-ffi-handler`, `use-catalog-session`.
Do: the rest of D-1 except `WorkspaceLocation` and `ReconcilerOutcomeDetail`; regenerate; D-2, D-3 (the handler half), D-6. Tests take a graph id from `evidencegraphstest.SoleGraph(t, c, sourceID).ID` or `artifactstest.Split`. `TestGetSourceGraphProgress` → `TestGetEvidenceGraphProgress`. Add T-6 to T-9.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./api/... ./core/...` is clean; `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes; `rg -n 'sourceGraphID|sourceEvidenceGraph|citations\.SourceID' api core` prints nothing (T-21); `rg -n 'GetSourceId\(\)' api/ffi/handlers/subjects.go api/ffi/handlers/observations.go api/ffi/handlers/promote.go api/ffi/handlers/event_titles.go api/ffi/handlers/citations.go api/ffi/handlers/evidence_graph_progress.go` prints only `citations.go` `CitationCountsBySource` (a Source-level read that stays).

**STEP-4 — Locations name their graph**
Files: `api/proto/engine.proto` (`WorkspaceLocation`, `ReconcilerOutcomeDetail`), `core/search/search.go`, `api/ffi/handlers/search_catalog.go`, `core/database/deleteimpact/{kinds.go, projectors.go, impact_test.go}`, `api/ffi/handlers/delete_impact.go`, `core/database/conclusiondetails/conclusiondetails.go` and its test, `api/ffi/handlers/conclusion_detail.go`.
Do: those two D-1 lines; regenerate; D-10's Go half (the projectors, `projectEvidenceGraph` included). `core/search/search.go` `WorkspaceLocation` gains `GraphID`, and `api/ffi/handlers/search_catalog.go` copies it both ways (no search hit sets it until S12-15). Add T-10 and T-11.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/deleteimpact/ ./core/database/conclusiondetails/ ./api/ffi/handlers/` passes.

**STEP-5 — The store speaks graph ids**
Files: `macos/App/Platform/{GenealogyStore.swift, GoStore.swift, FakeStore.swift, CoreInvoke.swift}`. Skill: `add-swift-test` (FakeStore).
Do: D-7, D-14, and `mapWorkspaceLocationToProto` / `mapWorkspaceLocationFromProto` for `graphId` (D-10; the D-11 rule lives in `WorkspaceLocation.init`, STEP-7). `CoreInvoke` method names follow D-1's renamed methods.
Check: none on its own (the Xcode build returns at STEP-8); `rg -n 'sourceID: String' macos/App/Platform/GenealogyStore.swift` shows no graph read or write from D-7.

**STEP-6 — Keys, rows, mutations and the session by graph**
Files: every file `rg -l 'SourceGraph|sourceGraph|mutatedSourceGraph|deletedSubject|promotedSubject' macos/App macos/ProvenenciaTests` lists, plus `macos/Provenencia.xcodeproj/project.pbxproj` (file renames). Skill: `add-workspace-place`.
Do: D-8 in full, D-9's `.promotedSubject` and `graphKey`. Rename mechanically first (types, files, keys, and every Swift test function whose name contains `SourceGraph`, with `SourceGraph` → `EvidenceGraph`; `macos/App/Platform/L10n.swift` line 5040's `SourceGraphProgress` type too), then re-key the registry, the session and the mutation call sites.
Check: `rg -n 'SourceGraph|sourceGraphProgress|mutatedSourceGraph' macos/App macos/ProvenenciaTests --glob '!**/Generated/**'` prints nothing; `rg -n '\.sourceGraph\b' macos/App` prints only `PlaceID.sourceGraph` and `WorkspacePresentationID.sourceGraph` uses (D-8).

**STEP-7 — The place carries the graph**
Files: `macos/App/Features/Workspace/{WorkspaceLocation.swift, WorkspaceToolbar.swift, WorkspaceDestinationHost.swift}`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`. Skills: `add-workspace-location`, `add-workspace-place`.
Do: D-10's Swift half (fields, normalization, factories, registry, host, breadcrumbs), D-11.
Check: none on its own; STEP-8 builds.

**STEP-8 — Every surface opens a graph**
Files: `macos/App/Features/EvidenceGraph/{EvidenceGraphView.swift, EvidenceGraphModel.swift}` (the doc comment at `EvidenceGraphView.swift` line 6 names `WorkspaceLocation.evidenceGraph(…)` instead of `sourceSurface: .graph`), `macos/App/Features/Sources/SourcesSplitRow.swift`, `macos/App/Features/CitationComposer/{CitationComposerEntry.swift, CitationComposerContext.swift, CitationComposerModel.swift, CitationComposerModel+Authoring.swift, CitationConnections.swift, CitationComposerView.swift}`, `macos/App/Features/Promote/{PromoteModel.swift, PromoteView.swift}`, `macos/App/Features/Sources/{SourcesListNavigation.swift, SourcesListView.swift, SourcesModel.swift, SourcePageIdentityHeader.swift, SourcePageModel.swift}`, `macos/App/Features/Conclusions/ReconciliationRecord.swift`, `macos/App/DesignSystem/Recipes/DeleteImpact/PVDeleteImpact.swift` (preview location gets `graphId: "preview-graph"`), and the Swift tests that build these (`EvidenceGraphModelTests`, `CitationComposerModelTests`, `CitationComposerDepictionTests`, `CitationAutoTranscribeTests`, `PromoteModelTests`, `SourcesModelTests`, `SourcesListNavigationTests`, `WorkspaceNavigationTests`, `PlaceRegistryTests`, `WorkspaceDestinationHostTests`, `WorkspaceToolbarBreadcrumbTests`, `WorkspaceSessionTests`, `CatalogQueryRegistryTests`, `PersonDetailTests`, `SearchCatalogStoreTests`, `EvidenceGraphStoreTests`, `SourceCopyModelTests`, `OnboardingModelTests`, `CatalogCountsTests`, `WorkspaceSidebarSectionsTests`, `PropertiesModelTests`, `PVDeleteImpactTests`, and any other test the build names). Skills: `add-workspace-location`, `add-swift-test`.
Do: D-9, D-12, D-13. Existing tests change only their ids, signatures and the mutations D-8 deletes (Source → graph per D-14); their expectations stay, except the Default-graph counts D-14 names.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` passes; `rg -n 'sourceSurface: \.(graph|citationComposer)' macos/App` prints only the two D-10 factories.

**STEP-9 — Swift tests for each AC**
Files: the test files named in §8.
Do: add T-12 to T-20 and T-23.
Check: the same `xcodebuild test` passes.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row; working rule "One bridge, one function"), this plan (status, Landed note), `docs/macos-client-patterns.md`, `.agents/skills/add-workspace-place/SKILL.md`, `.agents/skills/add-catalog-delete/SKILL.md`, `.agents/skills/add-workspace-location/SKILL.md`, `.agents/skills/add-audit-scope/SKILL.md`, `docs/ideas/shared-list-helpers.md`.
Do: per the Definition of done, and:
- `macos-client-patterns.md` line 91: `evidenceGraph` (`EvidenceGraphRows`); line 112: the naming example becomes `.deletedSource`, and add "A write that returns `WriteEffects` needs no mutation for what its effects name: the Evidence graph follows `evidenceGraphs`."
- `add-workspace-place`: line 71 `evidenceGraph`; line 172 "Graph and composer writes (`createSubject`, `updateSubject`, `deleteSubject`): `session.apply(effects)`, plus `.deletedSubject(sourceId:)` for a delete"; line 173 "the Source's graph and citation reads follow `sources`" → "each graph's rows follow `evidenceGraphs`, and the Source's citation reads follow `citationSources`".
- `add-catalog-delete` line 111: "source erase also busts `sourceWorkspace` / `evidenceGraph`".
- `add-audit-scope` line 24: `Effect.Source` → `Effect.EvidenceGraph`, "the Evidence graphs whose loaded rows a change moves: the graph's rows and graph progress" (citation counts are `Effect.CitationSource`).
- `add-workspace-location`: one line under its location rules: "A graph, citationComposer or promote location carries `graphId` beside `sourceId` and is built with `WorkspaceLocation.evidenceGraph(…)`, `.citationComposer(…)` or `.promote(…)`; one without `graphId` becomes the Source page in `init`."
- `shared-list-helpers.md`: delete the section "Source graph mutations" (`CatalogMutation.evidenceGraphId` does it); in "One catalog querier", drop `graphprogress`'s graph-Source read if listed and keep `conclusiondetails` and `effects` as open.
- Working rule "One bridge, one function": verify it reads as D-15 says; it was reworded when this plan was written. Verify the S12-18 scope entry names `evidencegraphs.SourceOf` and says `InsertTx` already reads the graph's Source (both written with this plan's review).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-08-graph-id-end-to-end.md` reports no errors; `rg -n 'mutatedSourceGraph|SourceGraphRows|Set\.Sources|Effect\.Source\b|refreshSourceGraphProgress' docs .agents .cursor core api macos --glob '!docs/deployment-plan/**'` prints nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Go fixtures: `artifactstest.Split(t, c, userID, sourceID, "Jones")` returns Default, Jones and the Jones page; `artifactstest.SubjectOn(t, c, userID, sourceID, graphID, typeKey, label)` puts a Subject on a graph. Swift fixtures: a `CatalogEvidenceGraph` literal per graph in `FakeStore.evidenceGraphsByProject`, graph ids distinct from Source ids (D-14).

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/evidencegraphs/evidencegraphs_test.go` `TestSourceOf` *(new)* | Default and Jones of S, read on a `*sql.Tx` and on `c.DB()` → S; an unknown 16-byte id, a 3-byte id, a nil querier → `ErrInvalid` | D-3 |
| T-2 | Go | `core/database/effects/registry_completeness_test.go` `TestEvidenceGraphPathsEndAtGraphs` *(new)* | Every registry entry with a non-zero `EvidenceGraph` → its `Path.check(table)` returns `"evidence_graphs"`; the set of such tables is `sources`, `artifacts`, `citations`, `observations`, `subjects`, `evidence_graphs`, `identity_claims`, `identity_claim_evidence`, `date_values`, `name_values`, `depiction_values` | D-5 |
| T-3 | Go | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve` *(changed)* | Source create → `EvidenceGraphs` = [Default]; graph create → [new graph]; rename → [that graph]; delete → [that graph]; link-only change → empty; Subject on Jones → [Jones]; Observation on Jones → [Jones]; Source retitle → [Default, Jones]; Artifact create → [Default, Jones] | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, D-5 |
| T-4 | Go | `core/database/graphfill/load_test.go` `TestGraphWriteDropsOnlyThatGraph` *(renamed from `TestSourceWriteDropsEveryGraph`)* | Default, Jones and a second Source's graph loaded with memos; Subject created on Jones through `writes.Run` → Jones and its memo dropped; Default, its memo and the other Source's graph kept | AC-2.1, D-4 |
| T-5 | Go | `api/ffi/handlers/source_write_effects_test.go` *(changed)* | Each existing row's `wantSources` becomes `wantGraphs`: evidence writes report the Source's graph id in `GetEvidenceGraphs()`, others report none | AC-2.3 |
| T-6 | Go | `api/ffi/handlers/subjects_test.go` `TestListSubjects` *(cases added)* | Split S, Ada on Default, Bea on Jones → `ListSubjects(graph_id = Jones)` = [Bea]; `ListSubjectPositions(Jones)` holds only Bea's cell; an unknown graph id → empty `Subjects` and `Positions`, no error | AC-3.1, AC-3.2, D-6 |
| T-7 | Go | `api/ffi/handlers/evidence_graph_progress_test.go` *(new)* (moved from `sourcegraphprogress_test.go`) `TestGetEvidenceGraphProgress` *(renamed, cases added)* | Jones with one person → `graph_id` Jones, `source_id` S, counts; unknown id → zeros, `graph_id` = request's, `source_id` empty; `ListEvidenceGraphProgress` → one row per graph with both ids | AC-3.3, AC-3.4, D-6 |
| T-8 | Go | `api/ffi/handlers/subjects_test.go` `TestCreateSubjectOnGraph` *(new)* | `graph_id` = Jones → Subject `source_id` S, `evidencegraphs.SubjectGraphTx` = Jones, `effects.evidence_graphs` = [Jones]; unknown graph id → error code `subjects.invalid` | AC-3.5, AC-2.1, D-3 |
| T-9 | Go | `api/ffi/handlers/citations_test.go` `TestGetCitation` *(new or cases added)* | A Citation with two Observations on the Source's graph, `GetCitation(graph_id, citation_id)` → the Citation, its notes and both Observations; `connect_test.go` builds `CreateCitedBridgeRequest` with no `SourceId` and passes | AC-3.6, AC-3.7, D-2 |
| T-10 | Go | `core/database/deleteimpact/impact_test.go` *(cases added beside lines 279, 748, 1048)* | A Subject on Jones → `Location.GraphID` = Jones; a Citation and an Observation on Jones → `GraphID` = Jones; a Citation with no Observation → `GraphID` = the first graph its Artifact is linked to; a graph row (Jones, in a Source delete's report) → `SourceSurface` `graph`, `GraphID` Jones | AC-1.7, D-10 |
| T-11 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` *(case added)* | An outcome whose Observation's Subject is on the Source's graph → `Outcome.GraphID` = that graph | AC-1.7, D-10 |
| T-12 | Swift | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift` `graphLocationsRoundTripTheGraph` *(new)*, `legacyGraphSurfaceWithoutGraphDecodesAsPage` *(new)* | Canvas, composer and Promote locations on `"g-1"` persist and restore with `graphId` `"g-1"`; JSON with `sourceSurface: "graph"` / `"citationComposer"` / `"promote"` and no `graphId` → `.page`, same `sourceId` | AC-1.5, AC-1.6, D-11 |
| T-13 | Swift | `macos/ProvenenciaTests/PlaceRegistryTests.swift` `resolvesSourceGraphDistinctFromPage` → `resolvesEvidenceGraphDistinctFromPage` *(renamed, changed)* | `.graph` on `"g-1"` → keys include `.evidenceGraph(graphId: "g-1")` and `.sourceWorkspace(sourceId: "src-1")`, `deepId` `"g-1"`; composer and Promote places key `.evidenceGraph("g-1")` | D-10 |
| T-14 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs` *(changed)*, `evidenceGraphEffectsReloadOnlyThatGraph` *(new)* | `CatalogWriteEffects(evidenceGraphs: ["g-2"])` → `.key(.evidenceGraph(graphId: "g-2"))` only; `.deletedSubject(sourceId: "src-1")` → `.citationCounts("src-1")` and no `.evidenceGraph` key; `.promotedSubject(graphId: "g-1")` → `.evidenceGraph("g-1")`; `.deletedSource` → `.allCached(.evidenceGraph)` | AC-2.1, AC-2.6, AC-2.7, D-8 |
| T-15 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `applyMutatedSourceGraphPatchesProgressOnly` → `applyGraphEffectsPatchProgressOnly`, `graphCameraIsPerSource` → `graphCameraIsPerGraph` *(renamed, changed)* | Two graphs of one Source with progress warm; `apply(CatalogWriteEffects(evidenceGraphs: ["g-2"]))` → only `"g-2"`'s progress entry is fetched and patched; `apply(.promotedSubject(graphId: "g-1"))` → `"g-1"`'s; cameras for `"g-1"` and `"g-2"` are kept apart | AC-2.1, D-8 |
| T-16 | Swift | `macos/ProvenenciaTests/SourcesListNavigationTests.swift` `graphLocationOpensTheFirstGraph` *(new)*, `graphLocationBlockedWithoutArtifact` *(changed)* | Graphs [Default `"g-1"`, Jones `"g-2"`] of `"src-1"` → location `graphId` `"g-1"`, `sourceId` `"src-1"`, `.graph`; entry graph with no `artifactIDs`, or no graph for the Source → nil | AC-1.1, AC-1.2, AC-1.3, D-12 |
| T-17 | Swift | `macos/ProvenenciaTests/EvidenceGraphModelTests.swift` `graphLocationsCarryTheGraph` *(new)*, `mutatedSourceGraphInvalidatesSourceGraph` → `subjectWriteEffectsReloadTheGraph` *(renamed, changed)* | A model on `"g-1"` of `"src-1"` → `composerLocation`, `connectComposerLocation`, `promoteLocation`, `promoteAllLocation` have `graphId` `"g-1"` and `sourceId` `"src-1"`; `confirmCreate` calls FakeStore `createSubject` with `graphID` `"g-1"` and the FakeStore effects name `["g-1"]`, which reload `.evidenceGraph("g-1")` | AC-1.4, AC-2.7, D-13, D-14 |
| T-18 | Swift | `macos/ProvenenciaTests/CitationComposerModelTests.swift` `entryNeedsTheGraph` *(new)* | `CitationComposerEntry(location:)` on a composer location with `graphId` `"g-1"` → entry `graphID` `"g-1"`; without `graphId` → nil; opening an existing Citation not in the rows calls `getCitation` with `graphID` `"g-1"` | AC-1.4, D-13 |
| T-19 | Swift | `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift` `citationComposerIncludesNavigableEvidenceGraph`, `promoteIncludesNavigableEvidenceGraph` *(changed)* | The graph crumb's target has `graphId` `"g-1"` and `.graph`; its label is unchanged | AC-1.4, D-10 |
| T-20 | Swift | `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` `createSourceMakesADefaultGraph` *(new)*; `macos/ProvenenciaTests/PersonDetailTests.swift` *(case added)* | FakeStore `createSource` → one graph named `Default` for it, and `createArtifact` links to it; a reconciler outcome with `graphID` `"g-1"` → `ReconciliationRecord.composerLocation(for:)` has `graphId` `"g-1"` | AC-1.7, D-13, D-14 |
| T-21 | Go | STEP-3 check `rg -n 'sourceGraphID|sourceEvidenceGraph|citations\.SourceID' api core` *(command, written in STEP-3)* | Prints nothing | AC-3.8, D-2 |
| T-22 | Go | `core/database/subjects/subjects_test.go` `TestInsertReadsTheGraphsSource` *(new)* | `CreateInput{GraphID: Jones}` with no `SourceID` → Subject `SourceID` S, `GraphID` Jones; `GraphID` Jones with `SourceID` = another Source → `ErrInvalid`; neither id → `ErrInvalid`; unknown 16-byte `GraphID` → `ErrInvalid` | AC-3.5, D-3 |
| T-23 | Swift | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift` `graphSurfaceWithoutGraphBecomesThePage` *(new)* | The memberwise `init` with `.graph`, `.citationComposer` or `.promote` and `graphId` nil or `""` → `.page`, same `sourceId`; `GoStore.mapWorkspaceLocationFromProto` of a proto with `source_surface` `graph` and empty `graph_id` → `.page` | AC-1.6, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `api/ffi/handlers` (every graph RPC test), `core/database/effects` and `graphfill` (drop tests), `EvidenceGraphModelTests`, `CitationComposerModelTests`, `WorkspaceSessionTests`, `CatalogQueryRegistryTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no copy
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none: no UI change
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — none: no brief
- [ ] Domain doc updated if a rule in docs/*.md changed: macos-client-patterns.md, add-workspace-place, add-catalog-delete, add-workspace-location, add-audit-scope, core/database/effects/README.md, docs/ideas/shared-list-helpers.md, the "One bridge" working rule (STEP-2, STEP-10)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. (AC-1.1) Open an existing project. From the Sources list, Open graph on a Source with Artifacts: the canvas shows the same bubbles in the same places as on `main`. From the Source page, Jump to evidence graph opens the same canvas.
2. (AC-1.3) A Source with no Artifact: Open graph and Jump to evidence graph are disabled.
3. (AC-1.4, AC-2.1) On the canvas, add a Subject: it appears at once and the Sources list progress for that Source moves by one. Open a card's property in the composer, save an Observation, go Back: the canvas shows it. Promote a card: the Promote page opens on that card. Back returns to the canvas on the same graph; Back again returns to the composer.
4. (AC-1.5) Quit and relaunch: the app restores the last place (canvas, composer or Promote) on the same graph; Back / Forward walk the same entries.
5. (AC-1.7) Open a Person whose reconciliation record lists an Observation; follow it: the composer opens on that Observation.
6. (AC-1.6) Not checkable by hand without a pre-PR history file; covered by T-12.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Swift caller keeps sending a Source id where a graph id is now expected; with one graph per Source nothing visible fails | D-14's distinct fixture ids; T-14 to T-18 assert graph ids |
| A history file on a tester's machine holds graph surfaces with no graph | D-11 decodes them as the Source page; T-12 |
| Artifact and Citation writes still reload every graph of the Source | Same reach as `main` (D-5); narrowing is Out |
| Swift does not build from STEP-2 to STEP-8 | Stated in §7; Go checks gate those steps |
| The FakeStore reports a different graph set than the engine for one write | D-14 names each write's graphs; T-17 checks the Subject case end to end |

**Stop and report if:**

- `./scripts/generate-proto.sh` cannot run (no `protoc`, `protoc-gen-go` or `protoc-gen-swift`), or `xcodebuild` is not available: stop before STEP-2 and report; the generated files must be regenerated, not hand-edited.

- An `effects` path in D-5 panics at package init (an edge `catalogmodel` does not declare), or T-2 finds a path ending elsewhere than `evidence_graphs`.
- `inbound("evidence_graphs", "source_id")` on a Source delete leaves a deleted graph loaded in `graphcache` (a test in `graphfill` or `deleteimpact` reads a found graph after its Source is deleted).
- `TestEvidenceWarmIssuesNoQuery`, `TestProposeWarmIssuesNoQuery` or `TestEvidenceGraphSecondReadIssuesNoQueries` fails after STEP-2 while passing on `main`.
- A Swift test needs a changed expectation (not just ids, signatures or the deleted mutations) in STEP-8, other than the D-8 / D-10 renames and the Default-graph counts D-14 names.
- A Swift write that D-8 moves onto `session.apply(effects)` alone returns no effects, or its effects do not name the Subject's graph (the graph would stop reloading).
- Something other than Promote's three RPCs, `promotealign`, `InsertTx` and tests still calls `evidencegraphs.SoleGraph` or `SoleGraphTx` after STEP-3.
- The STEP-10 `rg` finds a hit outside the files STEP-10 lists.

## 12. Open questions

None.

## 13. PR

**Title:** S12-08 — Graph id end to end
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-08-graph-id-end-to-end.md`.
