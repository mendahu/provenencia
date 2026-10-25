# S12-11 — Source page Graphs section

**Status:** Blocked — on board S12-D2
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D2](../design/S12-D2-source-page-graphs.md), board not agreed yet
**Board:** none yet. Do not start this PR until this line holds the agreed Claude Design share link and Status reads Planned. At agreement the planner compares the frames with §6 and refreshes this plan where they differ (brief §7 Handoff).
**Depends on:** S12-05 (landed [#511](https://github.com/mendahu/provenencia/pull/511)), S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528)), S12-D2
**Blocks:** S12-12
**Branch:** `s12-11-source-page-graphs`, branched from `main` after the plan branch merges (so STEP-7 edits this file in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-10 did; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (place query keys and the session's graph list), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md) (breadcrumbs, `WorkspaceLocation.sourcePage`), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`
**Context:** [where graphs show](../source-structure.md#where-graphs-show), [rules](../source-structure.md#rules); working rules: "Source page", "One graph list", "A graph is a place", "Name and description are editable; graphs list alphabetically by name", "Refusals are of two kinds", "High tolerance for churn"; the S12-05 [Landed note](S12-05-graph-rpcs.md) (graph list Spec), the S12-08 [Landed note](S12-08-graph-id-end-to-end.md)
**Planned:** 2026-10-25 against `main` @ `dff91d9`
**Review:** [`S12-11-source-page-graphs.review.md`](S12-11-source-page-graphs.review.md) — round 1: Fits with changes / Not ready; every finding Fixed; round 2 below

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

The Source page shows every Evidence graph of its Source in a new **Evidence graphs** section: one row per graph, by name, with its name, `GPH-` ref, description, an **Open** button that opens that graph's canvas, and a **Delete** button that goes through the DeleteImpact recipe. The header's **Jump to graph** button is removed. The toolbar breadcrumbs on the canvas, the Citation composer and Promote name the Source and the graph: **Sources › {Source title} › {graph name}**. This is the first PR of the slice that makes several graphs per Source visible ([dogfood bar](../deployment-plan.md#goal-dogfood-bar) outcome 5, Source page half).

**Why now.** S12-05 put the graph list and graph delete on the wire and S12-08 made a graph a place, so the page can list and open graphs without engine work. S12-12 (Graph form) adds create and edit to this section and needs it first.

**What stays.** The engine, FFI and store protocol are unchanged. Every other Source page section is unchanged. On an existing project every Source has one graph named Default, so the page shows one row where the header button was, and its Open goes where Jump to graph went. The Sources list Open keeps opening the Source's first graph (S12-14 replaces it).

## 2. User stories

**US-1.** As a researcher on a Source page, I see the Source's Evidence graphs and open the one I want, so that I can work one household of a census without hunting for it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source with graphs "Smith household", "Default", "Jones household" | The section header reads "Evidence graphs" with meta "3 evidence graphs"; rows read Default, Jones household, Smith household in that order, each with its ref and description |
| AC-1.2 | A graph named "Default" | The row shows "Default" as typed (no localized label) |
| AC-1.3 | Open on "Jones household", which links one Artifact | The canvas of that graph opens: the location's `graphId` is its id, surface `.graph` |
| AC-1.4 | A Source with no Artifacts (its Default graph links none) | Open is disabled; tooltip and VoiceOver hint "Add an artifact to open this evidence graph" |
| AC-1.5 | The page's header | No Jump to graph button; Copy… and Delete remain |
| AC-1.6 | The graph list has no value yet | The section header shows with no meta and one skeleton row |
| AC-1.7 | The graph list read failed with no value | A danger callout in place of the rows; the section header stays; the other sections render |

**US-2.** As a researcher, I delete a graph I no longer need, and I'm told why when I can't.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Delete on an empty "Jones household" (`GPH-JONES`) of a two-graph Source | The DeleteImpact confirm for kind `evidence_graph`, title "Delete evidence graph GPH-JONES?"; confirming calls `deleteEvidenceGraph` once, the row leaves, the toast reads "Evidence graph deleted" / "GPH-JONES was erased from this source." |
| AC-2.2 | Delete on a graph with 3 Subjects | The blocked notice with the group "This graph has 3 subjects"; `deleteEvidenceGraph` is not called |
| AC-2.3 | Delete on a Source's only graph | The `last_graph` notice "This is the source’s only evidence graph"; `deleteEvidenceGraph` is not called |
| AC-2.4 | Delete is visible on every row | It is never disabled (it is disabled only while the Source itself has no value, as the header Delete) |

**US-3.** As a researcher on a graph's canvas, composer or Promote page, the breadcrumbs name the Source and the graph, so that I know which household I'm in and can step back to the Source page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Canvas of "Jones household" on "1851 census" | Crumbs: Sources › 1851 census › Jones household. "1851 census" goes to the Source page; the last crumb has no action |
| AC-3.2 | Composer for scope "John Jones" on that graph | Crumbs: Sources › 1851 census › Jones household › Citation for John Jones; "Jones household" goes to that graph's canvas |
| AC-3.3 | Promote for `SUB-AAAAA` on that graph | Crumbs: Sources › 1851 census › Jones household › Promote SUB-AAAAA |
| AC-3.4 | The Source is renamed "1851 census, Ward 3", or the graph renamed | The crumbs show the new name with no navigation (names come from the lists) |
| AC-3.5 | Neither list has a value and the location carries no Source title | The graph crumb reads "Evidence graph" and the Source crumb is left out: Sources › Evidence graph › … |

## 3. Scope

| In | Covered by |
| --- | --- |
| Evidence graphs section between the overview columns and Artifacts, header with count (GS-1) | STEP-3, AC-1.1 |
| One card row per graph in list order: name, ref, description (GS-2, GS-10) | STEP-3, AC-1.1, AC-1.2 |
| Open per graph, disabled without a linked Artifact (GS-3, GS-4) | STEP-1, STEP-3, AC-1.3, AC-1.4 |
| Delete per graph through the DeleteImpact recipe, toast after (GS-3, GS-5, GS-6) | STEP-2, STEP-3, AC-2.1 – AC-2.4 |
| Loading and error states (GS-8) | STEP-3, AC-1.6, AC-1.7 |
| VoiceOver labels (GS-9) | STEP-3, manual check |
| Header Jump to graph removed with its model properties, helpers and string (GS-7) | STEP-4, AC-1.5 |
| `evidence_graph` delete-impact noun | STEP-2, AC-2.1 |
| Session accessors for one graph and one Source's graphs | STEP-1 |
| `WorkspaceLocation.sourcePage(sourceId:ref:title:)`, with its three hand-built copies moved onto it | STEP-5 |
| Breadcrumbs Sources › Source › graph name on canvas, composer, Promote, names read from the lists | STEP-5, AC-3.1 – AC-3.5 |
| Graph and Promote places warm the graph list | STEP-5, T-9 |
| The stored name `Default` shown verbatim (S12-03 review UX-3, user 2026-10-25) | D-6, AC-1.2 |

| Out | Home |
| --- | --- |
| Add graph and Edit on a row | S12-12 (board S12-D3) |
| Artifact link checkboxes; an id index behind the session's graph accessors | S12-13 (board S12-D4) |
| Per-graph progress; the Sources list per-graph Open (`SourcesListNavigation.entryGraph` stays) | S12-14 (board S12-D5) |
| Graphs in the omnibar | S12-15 |
| A Promote button per graph on the Source page (S12-09's "Not this PR" named S12-11) | not planned — Promote opens from the graph canvas; the working rule "Source page" lists name, ref, description, Open |
| Pruning history entries of a deleted graph | not planned — a deleted graph holds no Subjects; its history entry opens as any missing graph does today |
| A localized label for a graph named `Default` | not planned (D-6) |
| A kit list component with trailing actions | not planned — one call site (brief non-goals) |
| Aligning the Notes section's bare-number count with Artifacts' localized count | dogfood log (`docs/dogfood/ux.md`), one line in STEP-7 |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Header jump | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `evidenceGraphJump` | `PVButton(L10n.Sources.jumpToEvidenceGraph, variant: .secondary, size: .sm, icon: .gitBranch)` before Copy… and Delete in an `HStack(spacing: PVSpacing.space4)`; goes to `model.evidenceGraphLocation`; help / hint from `model.evidenceGraphBlockedReason`; id `sources.page.jumpToEvidenceGraph`. Removed by this PR |
| Page model | `macos/App/Features/Sources/SourcePageModel.swift` `evidenceGraphLocation`, `evidenceGraphBlockedReason` | The entry graph's location and blocked reason, read through `SourcesListNavigation.graphsBySource(in:)` and `graphListLoaded(in:)`. Only the header reads them. Dead after STEP-4 |
| Page delete | `SourcePageModel.swift` `presentImpact(kind:id:ref:title:)` (private), `confirmResourceDelete()`, `deleteImpact: DeleteImpactFlow` | `askDeleteSource` / `askDeleteArtifact` call `presentImpact`. Inside `deleteImpact.confirm { target in … }`: `if target.kind == "source" { …; return }`, then everything after is the Artifact delete (store call, workspace patch, toast, `session.apply(effects)`, `refreshCoverFromStore()`) |
| Page sections | `SourcePageModel` `identity`, `credibility`, `metadata`, `notes`, `artifacts` (each `Source*Section(context:)`) | Each section is a class holding a `SourcePageContext`. No graphs section |
| Page layout | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageContent.body` | `overviewColumns`, then `SourcePageArtifactsView(model:)`, then `SourcePageNotesView(model:)` in one `VStack(spacing: PVSpacing.space9)`; `.pvDeleteImpact(flow: model.deleteImpact, …)` and `.vocabularyToastOverlay($model.toast, …)` already on the page |
| Section shape | `macos/App/Features/Sources/SourcePageNotesView.swift` | `VStack(alignment: .leading, spacing: PVSpacing.space6)` of `PVSectionHeader(title:meta:)` then rows |
| Section count | `SourcePageArtifactsView.swift` `L10n.Sources.artifactsCount(_:)` ("%lld artifacts"); `SourcePageNotesView.swift` bare `"\(count)"` | Two mechanisms on one page |
| Navigation helpers | `macos/App/Features/Sources/SourcesListNavigation.swift` `pageLocation(for:)`, `entryGraph`, `graphsBySource(_:)`, `graphsBySource(in:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`, `graphListLoaded(in:)`, `graphLocation(for:graphs:)` | `graphLocation(for:graphs:)` picks the entry graph and returns `.evidenceGraph(sourceId:graphId:ref:title:)` (title = Source title) when it links an Artifact. `graphsBySource(in:)` and `graphListLoaded(in:)` have one caller each, in `SourcePageModel` (lines 70, 78). `graphsBySource(_:)` and `graphBlockedReason` are used by `SourcesModel` |
| Graph list | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` `evidenceGraphsKey`, `evidenceGraphsHandle`, `linkedArtifacts(graphID:artifacts:)` | "The one place its key and handle are looked up". `linkedArtifacts` does `graphs.first { $0.id == graphID }` inline. No by-id or by-Source accessor |
| Source page locations | `SourcesListNavigation.pageLocation(for:)`; `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `sourcePageLocation(title:ref:)`; `macos/App/Features/CitationComposer/CitationComposerModel.swift` `sourcePageLocation()` | Three hand-built `WorkspaceLocation(section: .sources, sourceId:, sourceSurface: .page, …)`. `WorkspaceLocation` has named constructors for `.evidenceGraph`, `.citationComposer`, `.promote`, not for the page |
| Source title lookup | `EvidenceGraphModel.swift` `resolvedSourceTitle()` | Reads `CatalogQueryKey.sourcesList` and finds the Source by id |
| Graph type | `macos/App/Platform/GenealogyStore.swift` `CatalogEvidenceGraph` | `id`, `ref`, `sourceID`, `name`, `description`, `artifactIDs` |
| Graph delete | `GenealogyStore.deleteEvidenceGraph(projectDir:userID:graphID:) -> CatalogWriteEffects`; `GoStore`, `FakeStore` | Effects carry `lists: [.evidenceGraphs]`, which refreshes the graph list through `CatalogQueryRegistry`. `FakeStore.deleteEvidenceGraph` removes the row, with no call counter |
| Delete impact | `core/database/deleteimpact/register.go` (`fkEdge(catalogmodel.KindEvidenceGraph, "subjects.graph_id", …)`, the `last_graph` count); `core/database/catalogmodel/model.go` `KindEvidenceGraph = "evidence_graph"`; `api/ffi/handlers/delete_impact.go` `GetDeleteImpact` (through `deleteimpact.ParseKind`) | Subjects are an inbound group `subjects.graph_id` (heading "This graph has %lld subjects"); the last graph is gate `last_graph` (`L10n.DeleteImpact.gateLastGraphTitle` / `Body`) |
| Delete copy | `macos/App/Platform/L10n.swift` `L10n.DeleteImpact.noun(_:count:)`, `confirmTitle(noun:ref:)` ("Delete %1$@ %2$@?") | No `evidence_graph` noun: the confirm reads "Delete item GPH-…?" |
| FakeStore impact | `macos/App/Platform/FakeStore.swift` `getDeleteImpact`, `deleteImpactByID`, `listEvidenceGraphsError`, `deleteSourceCalls`, `deleteArtifactCalls` | No `evidence_graph` branch; tests set `deleteImpactByID[id]` |
| Breadcrumbs | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `breadcrumbItems(for:goTo:)` (static) | Graph surface: Sources › "Evidence graph" (id `leaf-sources`). Composer / Promote: Sources › `evidenceGraphFor(sourceTitle:)` ("Evidence graph for {source}", id `graph-<id>`, → canvas) › Citation / Promote. The toolbar has no `WorkspaceSession` environment read; `WorkspaceView` sets `.environment(session)` above it; its `#Preview` sets none |
| Place keys | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` Specs `.sourceGraph`, `.sourcePromote` | Both warm `.sourcesList`; neither warms `.evidenceGraphsList` (`.sourceCitationComposer` and `.sourceDetail` do, the composer as its last key) |
| Strings | `L10n.Sources.jumpToEvidenceGraph`, `L10n.Workspace.evidenceGraphFor(sourceTitle:)`, `L10n.Workspace.evidenceGraphTitle`, `L10n.Sources.needsArtifactTooltip`, `L10n.Sources.toastArtifactDeletedBody(ref:)` ("%@ was erased from this source.") | `evidenceGraphTitle` is also the canvas title (`EvidenceGraphView.swift`) and stays. `needsArtifactTooltip` stays (Sources list Source-level zone) |
| Xcode target | `macos/Provenencia.xcodeproj/project.pbxproj` | Hand-maintained: each Swift file has a `PBXBuildFile`, a `PBXFileReference`, a group child and a Sources-phase entry, with `AA…` ids (`SourcePageArtifactsView.swift` at lines 59, 481, 930, 2250) |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` (`askDeleteSourcePresentsAllowedImpact`, `blockedArtifactDoesNotCallDelete`, private `settle(_:)`), `SourcesListNavigationTests.swift`, `WorkspaceToolbarBreadcrumbTests.swift` (`evidenceGraphLeafUsesGraphTitle`, `citationComposerIncludesNavigableEvidenceGraph`, `promoteIncludesNavigableEvidenceGraph`), `PlaceRegistryTests.swift` (`queryKeysMatchDesignTable`, exact key arrays), `EvidenceGraphModelTests.swift` | The three breadcrumb tests and two `queryKeysMatchDesignTable` cases change in STEP-5 |

**Precedent.** S12-02 ([#485](https://github.com/mendahu/provenencia/pull/485)) added a Source page action (Copy…) with state on `SourcePageModel` and its own view file; copy that split. Artifact delete in `SourcePageModel.askDeleteArtifact` / `confirmResourceDelete` is the delete precedent: `presentImpact`, then the store call, `session.apply(effects)` and a success toast. Do not copy Artifact delete's hand-written `setQueryValue` workspace patch or `refreshCoverFromStore()`: graph delete changes no workspace field and no cover, and the effects refresh the graph list.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A new section class `SourceGraphsSection` *(new)* in `Features/Sources/SourceGraphsSection.swift` *(new)* owns the rows: `rows: [CatalogEvidenceGraph]?` (nil while the graph list has no value), `loadError: Error?` (the handle's error while it has no value), `location(for:)` and `blockedReason(for:)`, both forwarding to `SourcesListNavigation` (D-3, D-4). `SourcePageModel` gains `var graphs: SourceGraphsSection`. | The page's sections are classes on the model; S12-12 adds its form state to this class. | Computed properties on `SourcePageModel` — S12-12 would move them. |
| D-2 | `WorkspaceSession+EvidenceGraphs` gains `evidenceGraph(id:) -> CatalogEvidenceGraph?` *(new)* and `evidenceGraphs(sourceID:) -> [CatalogEvidenceGraph]?` *(new)* (nil while the list has no value; list order otherwise). `linkedArtifacts(graphID:artifacts:)` calls `evidenceGraph(id:)`. `SourceGraphsSection.rows` and the toolbar read these. | The file is the one owner of graph-list reads; S12-13's id index then changes one file (review M-1). | Inline `filter` / `first` at each reader. |
| D-3 | `SourcesListNavigation.graphLocation(for source: CatalogSource, graph: CatalogEvidenceGraph) -> WorkspaceLocation?` *(new)* returns `.evidenceGraph(sourceId:graphId:ref:title:)` with the Source's ref and title, nil when `graph.artifactIDs` is empty. `graphLocation(for:graphs:)` calls it on the entry graph. | One rule for "a graph opens when it links an Artifact", shared with S12-14. | A copy of the rule in the section. |
| D-4 | `SourcesListNavigation.graphBlockedReason(for graph: CatalogEvidenceGraph) -> String?` *(new)* returns `L10n.Sources.graphNeedsArtifactTooltip` *(new)*, key `sources.graph.needsArtifact`, "Add an artifact to open this evidence graph", when `graph.artifactIDs` is empty, nil otherwise. | The rule and its reason sit together for S12-14's rows (review M-2). `needsArtifactTooltip` says "on the source page", wrong on the page and for one of several graphs. | A reason in `SourceGraphsSection`; reusing `needsArtifactTooltip`. |
| D-5 | Delete is always enabled (except while `model.source` is nil). `SourcePageModel.askDeleteGraph(id:)` *(new)* calls `presentImpact(kind: "evidence_graph", id:, ref: graph.ref, title: graph.name)`. The `confirm` closure in `confirmResourceDelete` becomes `switch target.kind` with cases `"source"` (today's body), `"artifact"` (today's fall-through body) and `"evidence_graph"` (`store.deleteEvidenceGraph`, `session.apply(effects)`, toast), and `default: return` (no write). It returns false for a graph (the page stays). | User 2026-10-25: the recipe explains, as Source and Artifact delete. A fourth kind cannot fall into `deleteArtifact` (review M-7). | Disabling Delete on the last graph; a separate `PVConfirm`; an `if` branch before the fall-through. |
| D-6 | A graph's name is shown verbatim everywhere, including `Default`. | User 2026-10-25 (S12-03 review UX-3). Renaming is S12-12. | A localized label for `Default`. |
| D-7 | Section placement: between `overviewColumns` and `SourcePageArtifactsView` in `SourcePageContent`. | User 2026-10-25. | Top of page; below Artifacts. |
| D-8 | Loading is one `PVCard` holding `PVSkeletonBar(width: 160, height: 12)` over `PVSkeletonBar(width: nil, height: 10)`, spacing `PVSpacing.space4`. Error is `PVCallout(tone: .danger, message: L10n.Errors.message(for:))` under the section header. No retry button: the session revalidates the list on its next visible read. | Brief GS-8; the kit. | A spinner; a Read again button. |
| D-9 | Breadcrumbs: `breadcrumbItems(for:names:goTo:)` takes `names: WorkspaceToolbar.BreadcrumbNames` *(new)*, a struct of `sourceTitle: (String) -> String?` and `graphName: (String) -> String?`, default `.none` (both return nil). For the graph, composer and Promote surfaces with a `graphId`: a Source crumb (id `source-<sourceId>`, label `names.sourceTitle(sourceId)` ?? the location's Source title (`title` on `.graph`, `sourceTitle` on composer and Promote), left out when both are empty; action `goTo(.sourcePage(sourceId:, ref: nil, title: label))`); then the graph crumb (id `graph-<graphId>`, label `names.graphName(graphId) ?? L10n.Workspace.evidenceGraphTitle`; on the graph surface it is the leaf with no action; on composer and Promote it goes to `.evidenceGraph(sourceId:graphId:title:)` with the Source crumb's resolved label as `title`); then Citation / Promote as today. `WorkspaceToolbar` reads `@Environment(WorkspaceSession.self)` and builds `BreadcrumbNames` from `session.evidenceGraph(id:)?.name` and the `sourcesList` handle's Source title. | User 2026-10-25 (Sources › Source › graph). One rule: names come from the lists, so a Source or graph rename shows at once (review M-6). The static function stays testable without a session. | A `graphName` field on `WorkspaceLocation` — stale on rename; reading the Source title from per-surface location fields only. |
| D-10 | `PlaceRegistry` `.sourceGraph` and `.sourcePromote` append `.evidenceGraphsList(project: project)` as their last key, as `.sourceCitationComposer` does. | A restored canvas or Promote page shows its graph name without visiting the Source page. | A permanent fallback label. |
| D-11 | Deleted: `SourcePageIdentityHeader.evidenceGraphJump`; `SourcePageModel.evidenceGraphLocation`, `evidenceGraphBlockedReason`; `SourcesListNavigation.graphsBySource(in:)`, `graphListLoaded(in:)`; `L10n.Sources.jumpToEvidenceGraph`; `L10n.Workspace.evidenceGraphFor(sourceTitle:)`; their `Localizable.xcstrings` entries. The `graphBlockedReason(hasArtifact:…)` doc comment drops "/ Jump to evidence graph". | Working rule "High tolerance for churn" (review M-3). | Keeping them for S12-14 — S12-14 reads D-3 and D-4. |
| D-12 | The delete-impact noun for `evidence_graph`: `L10n.DeleteImpact.nounEvidenceGraph` *(new)* `deleteImpact.kind.evidence_graph.one` "evidence graph" and `nounEvidenceGraphs` *(new)* `.other` "evidence graphs", in `noun(_:count:)`. The confirm title then reads "Delete evidence graph GPH-JONES?" through the existing `confirmTitle(noun:ref:)`. | The recipe's titles read the noun. | The "item" fallback; changing the recipe's confirm title. |
| D-13 | The toast: `L10n.Sources.toastGraphDeletedTitle` *(new)* "Evidence graph deleted", `L10n.Sources.toastGraphDeletedBody(ref:)` *(new)* "%@ was erased from this source." | Brief GS-6; the same verb as the Artifact toast on this page (review M-10). | "deleted" — a second verb for the same act on one page. |
| D-14 | Tests set graph-delete impacts through `FakeStore.deleteImpactByID`. `FakeStore` gains `deleteEvidenceGraphCalls` *(new)*, incremented in `deleteEvidenceGraph`. Tests that read the graph list call `settle(session.evidenceGraphsHandle!)` after `warmFromSession()` and after `confirmResourceDelete()`. | Matches `deleteSourceCalls`; `warmFromSession` does not wait on the graph list. | A FakeStore `evidence_graph` impact branch; adding the list to `warmFromSession`'s wait. |
| D-15 | The section count is `L10n.Sources.graphsCount(_:)` *(new)* `sources.page.graphsCount` "%lld evidence graphs", as Artifacts' `artifactsCount`. | The neighbouring section's mechanism, localized digits (review M-8). | A bare number as Notes — the Notes drift is logged, not copied. |
| D-16 | `WorkspaceLocation.sourcePage(sourceId: String, ref: String? = nil, title: String? = nil)` *(new)* builds the Source page location. `SourcesListNavigation.pageLocation(for:)`, `EvidenceGraphModel.sourcePageLocation(title:ref:)` and `CitationComposerModel.sourcePageLocation()` call it; the breadcrumb (D-9) calls it. | A fourth hand-built copy is the third-copy rule's trigger (review M-5); the graph surfaces already have named constructors. | Building it in the toolbar by hand. |
| D-17 | Row layout: `PVCard(elevated: false)` with its default padding; inside, `HStack(alignment: .top, spacing: PVSpacing.space6)` of the text `VStack(alignment: .leading, spacing: PVSpacing.space2)`, `Spacer(minLength: PVSpacing.space6)`, and an `HStack(spacing: PVSpacing.space4)` of Open then Delete with `.fixedSize()`. Text lines: name, then ref under it, then description. Rows stack in `VStack(alignment: .leading, spacing: PVSpacing.space4)`; the section is `VStack(alignment: .leading, spacing: PVSpacing.space6)` like Notes. At a narrow width the text wraps within its line limits and the buttons keep their size. | Removes the layout choices from the implementer (review X-4); header action spacing is `space4`. | Ref beside the name (it competes with long names). |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourceGraphsSection` *(new)* | type | `SourceArtifactsSection` (`Features/Sources/SourceArtifactsSection.swift`) | Same shape (a section class on a `SourcePageContext`); Artifacts' drafts and file staging do not apply. |
| `SourcePageGraphsView` *(new)*, private `SourcePageGraphRow` *(new)* | component (snowflake) | `SourcePageNotesView`, `SourcePageArtifactsView`; `PVList` | `PVList` rows are buttons with no trailing actions ("Rows are buttons, not selection"); the Artifacts card is a preview-well grid card. Brief inventory marks the row a snowflake. |
| `WorkspaceSession.evidenceGraph(id:)`, `evidenceGraphs(sourceID:)` *(new)* | func | `WorkspaceSession.linkedArtifacts(graphID:artifacts:)` | Extends the one owner; `linkedArtifacts` moves onto `evidenceGraph(id:)`. |
| `SourcesListNavigation.graphLocation(for:graph:)`, `graphBlockedReason(for:)` *(new)* | func | `graphLocation(for:graphs:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)` | Extends them: the entry-graph form calls the per-graph one. |
| `WorkspaceLocation.sourcePage(sourceId:ref:title:)` *(new)* | func | `WorkspaceLocation.evidenceGraph(…)`; three hand-built page locations | Replaces the three copies. |
| `WorkspaceToolbar.BreadcrumbNames` *(new)* | type | `EvidenceGraphModel.resolvedSourceTitle()` | A two-closure value so the static builder stays session-free; the model's lookup is private to the canvas. |
| `SourcePageModel.askDeleteGraph(id:)` *(new)* | func | `askDeleteArtifact(id:)` | Same call into `presentImpact`, another kind. |
| `L10n.Sources.graphsHeading`, `graphsCount(_:)`, `openEvidenceGraph`, `openEvidenceGraphLabel(name:)`, `deleteEvidenceGraphLabel(name:)`, `graphNeedsArtifactTooltip`, `toastGraphDeletedTitle`, `toastGraphDeletedBody(ref:)` *(new)*; `L10n.DeleteImpact.nounEvidenceGraph(s)` *(new)* | key | `L10n.Sources.artifactsHeading`, `artifactsCount`, `toastArtifactDeleted*`, `L10n.DeleteImpact.noun*` | New copy in existing namespaces. |
| `FakeStore.deleteEvidenceGraphCalls` *(new)* | helper | `FakeStore.deleteSourceCalls` | Same counter pattern. |

**Builds on.** The page's `DeleteImpactFlow` and `pvDeleteImpact`; the session's one graph list and its accessor file; `PlaceRegistry` keys; `PVBreadcrumbs`; `SourcesListNavigation` as the home of the per-graph open rule; `WorkspaceLocation`'s named constructors.
**Deletes.** D-11, plus the three hand-built Source page locations (D-16) and `linkedArtifacts`' inline lookup (D-2).
**Next adds.** S12-12 adds Add graph to the section header's `actions` slot, Edit to `SourcePageGraphRow` and its form state to `SourceGraphsSection`. S12-13 puts an id index behind `evidenceGraph(id:)` / `evidenceGraphs(sourceID:)` and moves `SourcesModel.graphsBySourceCache` onto it. S12-14 reuses `graphLocation(for:graph:)` and `graphBlockedReason(for:)` per accordion row, then deletes `entryGraph`.

### Schema

None — no catalog change.

### Go core

None — `evidencegraphs.List`, `Delete` and `deleteimpact` already do what the page needs.

### FFI

None — `ListEvidenceGraphs`, `DeleteEvidenceGraph` and `GetDeleteImpact` (kind `evidence_graph`) exist since S12-05.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section title | `PVSectionHeader(title: L10n.Sources.graphsHeading, meta:)` | `Features/Sources/SourcePageGraphsView.swift` *(new)* | `meta` = `L10n.Sources.graphsCount(rows.count)` when `rows` has a value, nil while loading or failed. `actions` empty (S12-12) |
| Graph row | `SourcePageGraphRow` (private snowflake) in `PVCard(elevated: false)` | same file *(new)* | D-17. Name `PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold)`, `lineLimit(2)`; ref `PVFont.mono(size: PVTypeScale.caption)`, `PVColor.textMuted`; description (when non-empty after trimming) `PVFont.body(size: PVTypeScale.bodySmall)`, `PVColor.textMuted`, `lineLimit(3)`. All three `Text(verbatim:)` |
| Open | `PVButton(L10n.Sources.openEvidenceGraph, variant: .secondary, size: .sm, icon: .gitBranch)` | same | `navigation.go(to:)` with `model.graphs.location(for:)`; `.disabled(location == nil)`; `.help` and `.accessibilityHint` from `blockedReason(for:)`; `.accessibilityLabel(Text(L10n.Sources.openEvidenceGraphLabel(name:)))`; id `sources.page.graph.open.<ref>` |
| Delete | `PVIconButton(.trash, label: L10n.Sources.deleteEvidenceGraphLabel(name:), size: .sm, tone: .danger)` | same | `Task { await model.askDeleteGraph(id:) }`; `.disabled(model.source == nil)`; id `sources.page.graph.delete.<ref>` |
| Loading | D-8 | same | id `sources.page.graphs.loading` |
| Error | D-8 | same | id `sources.page.graphs.error` |
| Delete flow | existing `.pvDeleteImpact(flow: model.deleteImpact, …)` | `SourcePageView.swift` | No change to the modifier |
| Section placement | — | `SourcePageView.swift` `SourcePageContent.body` | `SourcePageGraphsView(model: model)` between `overviewColumns` and `SourcePageArtifactsView(model: model)`; the section `VStack` has id `sources.page.graphs` |
| Header | remove `evidenceGraphJump` | `SourcePageIdentityHeader.swift` | The action `HStack` holds `copySourceButton`, `deleteSourceButton` |
| Breadcrumbs | `PVBreadcrumbs` (unchanged) | `Features/Workspace/WorkspaceToolbar.swift` | D-9 |

**Frames** (the brief's, as this plan builds them): frame 1 the page with one Default card under the overview columns and above Artifacts, header with Copy… and Delete only; frame 2 three cards in name order with one-line descriptions; frame 3 the Default card with Open disabled; frames 4 – 6 the existing DeleteImpact confirm and notices; frame 7 two cards and the success toast; frame 8 the header with no count and one skeleton card, then the header with a danger callout; frame 9 a two-line truncated name and a three-line truncated description.

States: loading, error (D-8), one graph, many graphs (no inner scroll), a graph with no linked Artifact (Open disabled), narrow window (D-17). Accessibility: Open and Delete carry the graph's name in their labels; the row's text reads in order (name, ref, description).

### Strings

L10n keys *(new)*; the planner confirms them against the agreed board before Status becomes Planned:

- `L10n.Sources.graphsHeading` `sources.page.graphsHeading` = "Evidence graphs"
- `L10n.Sources.graphsCount(_:)` `sources.page.graphsCount` = "%lld evidence graphs"
- `L10n.Sources.openEvidenceGraph` `sources.page.graph.open` = "Open"
- `L10n.Sources.openEvidenceGraphLabel(name:)` `sources.page.graph.openLabel` = "Open %@"
- `L10n.Sources.deleteEvidenceGraphLabel(name:)` `sources.page.graph.deleteLabel` = "Delete %@"
- `L10n.Sources.graphNeedsArtifactTooltip` `sources.graph.needsArtifact` = "Add an artifact to open this evidence graph"
- `L10n.Sources.toastGraphDeletedTitle` `sources.page.toastGraphDeletedTitle` = "Evidence graph deleted"
- `L10n.Sources.toastGraphDeletedBody(ref:)` `sources.page.toastGraphDeletedBody` = "%@ was erased from this source."
- `L10n.DeleteImpact.nounEvidenceGraph` `deleteImpact.kind.evidence_graph.one` = "evidence graph"; `nounEvidenceGraphs` `deleteImpact.kind.evidence_graph.other` = "evidence graphs"

Removed: `sources.page.jumpToEvidenceGraph`, `workspace.section.evidenceGraph.forSource`.

## 7. Steps

`XCT` below stands for `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`, the command `scripts/check-pr.sh` runs, with an `-only-testing:` class appended. New Swift files go in the Provenencia target by copying `SourcePageArtifactsView.swift`'s four `project.pbxproj` entries (lines 59, 481, 930, 2250 at the planning base) with new ids counting up from the highest `AA…` id in the file, in the same `Sources` group.

**STEP-1 — Graph list accessors and the per-graph open rule**
Files: `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift`, `macos/App/Features/Sources/SourcesListNavigation.swift`, `L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/SourcesListNavigationTests.swift`. Skills: `add-workspace-place`, `add-localized-string`, `add-swift-test`.
Do: D-2 (accessors; `linkedArtifacts` onto `evidenceGraph(id:)`), D-3, D-4. Add T-1, T-2.
Check: `XCT -only-testing:ProvenenciaTests/SourcesListNavigationTests` passes, existing tests unchanged; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — Graphs section model and graph delete**
Files: `macos/App/Features/Sources/SourceGraphsSection.swift` *(new)*, `SourcePageModel.swift`, `L10n.swift` (nouns, toast keys), `Localizable.xcstrings`, `macos/App/Platform/FakeStore.swift`, `project.pbxproj`, `SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: D-1, D-5, D-12, D-13, D-14. `rows` is `context.session.evidenceGraphs(sourceID: context.sourceID)`; `loadError` is the handle's `error` when its `status == .error` and `value == nil`. Add T-3 – T-6.
Check: `XCT -only-testing:ProvenenciaTests/SourcePageModelTests` passes; the xcstrings check passes.

**STEP-3 — Graphs section view**
Files: `macos/App/Features/Sources/SourcePageGraphsView.swift` *(new)*, `SourcePageView.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`. Skills: `add-ui-component`, `add-localized-string`.
Do: §6 macOS, D-7, D-8, D-15, D-17; strings as §6.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO` succeeds; run the app on any project (README "Run the Mac app") and see one Default card above Artifacts.

**STEP-4 — Remove the header jump**
Files: `SourcePageIdentityHeader.swift`, `SourcePageModel.swift`, `SourcesListNavigation.swift`, `L10n.swift`, `Localizable.xcstrings`.
Do: D-11 except `evidenceGraphFor` (STEP-5).
Check: `rg -n 'jumpToEvidenceGraph|evidenceGraphLocation|evidenceGraphBlockedReason|graphsBySource\(in|graphListLoaded\(in' macos` prints nothing; the build succeeds.

**STEP-5 — One Source page location; breadcrumbs name the Source and graph**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `SourcesListNavigation.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `L10n.swift`, `Localizable.xcstrings`, `WorkspaceToolbarBreadcrumbTests.swift`, `PlaceRegistryTests.swift`. Skills: `add-workspace-location`, `add-workspace-place`.
Do: D-16 (each of the three copies returns `.sourcePage(…)` with the fields it sets today); D-9, D-10, and the `evidenceGraphFor` half of D-11. Update the doc comment on `breadcrumbItems` to the new shapes. Give the toolbar `#Preview` `.environment(WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/preview.provenencia"), store: FakeStore()))`. Rewrite the three graph breadcrumb tests as T-7, add T-8, edit the two `queryKeysMatchDesignTable` cases (T-9).
Check: `XCT -only-testing:ProvenenciaTests/WorkspaceToolbarBreadcrumbTests -only-testing:ProvenenciaTests/PlaceRegistryTests -only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes; `rg -n 'evidenceGraphFor' macos` prints nothing; `rg -n 'sourceSurface: \.page' macos/App/Features` prints no Source page location built outside `WorkspaceLocation.swift`.

**STEP-6 — Full check**
Files: none new.
Do: run the whole gate.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md`, this plan, `docs/deployment-plan/spike-12/design/S12-D2-source-page-graphs.md` → `design/archive/`, `docs/deployment-plan/spike-12/design/README.md`, `docs/dogfood/ux.md`.
Do: per the Definition of done, and:
- `SourcesListNavigation.entryGraph`'s doc comment: "(S12-11, S12-14)" → "(S12-14)".
- Working rule "A graph is a place": "Until S12-11 and S12-14 add per-graph Open, the Source page jump and the Sources list Open open" → "Until S12-14 adds per-graph Open, the Sources list Open opens".
- The S12-13 scope entry: "point `SourcePageModel.evidenceGraphLocation`'s `graphsBySource` regroup at it too" → "back `WorkspaceSession.evidenceGraph(id:)` and `evidenceGraphs(sourceID:)` with it, and move `SourcesModel.graphsBySourceCache` onto it".
- `docs/dogfood/ux.md`: one line, "Source page section counts: Notes shows a bare number, Artifacts and Evidence graphs a localized count (S12-11 review M-8)".
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-11-source-page-graphs.md` reports no errors (a warning for the literal "Copy…" is expected).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesListNavigationTests.swift` `graphLocationOpensTheGivenGraph` *(new)* | Source `src-1`, graph `g-2` with one Artifact → `.evidenceGraph(sourceId: "src-1", graphId: "g-2", …)` with the Source's ref and title, and `graphBlockedReason(for:) == nil`; same graph with no Artifact → nil and `graphNeedsArtifactTooltip` | AC-1.3, AC-1.4, D-3, D-4 |
| T-2 | Swift | `SourcesListNavigationTests.swift` `sessionGraphAccessorsReadTheList` *(new)* | Session with no list value → `evidenceGraph(id:)` nil, `evidenceGraphs(sourceID:)` nil; with FakeStore graphs on `src-1` and `src-2`, settled → by-id finds `g-2`; by-Source returns `src-1`'s graphs in list order | D-2 |
| T-3 | Swift | `SourcePageModelTests.swift` `graphRowsAreTheSourcesGraphsByName` *(new)* | FakeStore graphs "Smith household", "Default", "Jones household" on `src-1` and one on `src-2`; warm, settle the list → `model.graphs.rows?.map(\.name)` == ["Default", "Jones household", "Smith household"] | AC-1.1, AC-1.2 |
| T-4 | Swift | `SourcePageModelTests.swift` `graphRowsNilUntilListLoadsAndErrorWhenItFails` *(new)* | Before warm → `rows == nil`; `listEvidenceGraphsError` set, warm, settle → `rows == nil`, `loadError != nil` | AC-1.6, AC-1.7 |
| T-5 | Swift | `SourcePageModelTests.swift` `deleteEmptyGraphCallsStoreAndToasts` *(new)* | Two graphs; `deleteImpactByID["g-jones"]` allowed; `askDeleteGraph(id: "g-jones")` → target kind `evidence_graph`; `confirmResourceDelete()` → false, `deleteEvidenceGraphCalls == 1`, `deleteArtifactCalls == 0`, toast title "Evidence graph deleted"; settle → rows no longer include `g-jones` | AC-2.1, D-5, D-13 |
| T-6 | Swift | `SourcePageModelTests.swift` `blockedGraphDoesNotCallDelete` *(new)*, parameterized over (`.inbound`, group `subjects.graph_id` total 3) and (`.lastGraph`, no groups) | `allowed == false` → after `confirmResourceDelete()`, `deleteEvidenceGraphCalls == 0` and the request's gate is the given one | AC-2.2, AC-2.3 |
| T-7 | Swift | `WorkspaceToolbarBreadcrumbTests.swift` `evidenceGraphCrumbsNameSourceAndGraph`, `citationComposerCrumbsNameSourceAndGraph`, `promoteCrumbsNameSourceAndGraph` (replace the three graph tests) | Names: `src-1` → "1851 census", `g-1` → "Jones household". Canvas → [Sources, 1851 census, Jones household], crumb 1 goes to `.sourcePage(sourceId: "src-1", …)`, last has no action; composer → […, Jones household, Citation for John Jones], crumb 2 goes to `.graph` `g-1`; Promote → […, Promote SUB-AAAAA] | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-9 |
| T-8 | Swift | `WorkspaceToolbarBreadcrumbTests.swift` `graphCrumbsFallBackWithoutNames` *(new)* | `.none` names, canvas with `title: "1851 census"` → Source crumb "1851 census", graph crumb "Evidence graph"; composer with no `sourceTitle` → no Source crumb | AC-3.5 |
| T-9 | Swift | `PlaceRegistryTests.swift` `queryKeysMatchDesignTable` (edit the `.promote` and `.graph` cases) | Expected arrays end with `.evidenceGraphsList(project:)` | D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceToolbarBreadcrumbTests`, `SourcePageModelTests`, `PlaceRegistryTests`, `EvidenceGraphModelTests` (D-16), `check-localizable-xcstrings.py` (removed keys).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (jumpToEvidenceGraph, evidenceGraph.forSource removed)
- [ ] UI matches the brief inventory: the PR description lists each component used per inventory row, with a screenshot of frames 1 and 2
- [ ] Version bump: none (VERSION stays 0.2.0)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status, working rule and S12-13 scope entry reworded (STEP-7); completed.md: one short paragraph of what shipped
- [ ] Brief S12-D2 moved to design/archive/ and design/README.md updated
- [ ] Domain doc: none — no rule in docs/*.md changes; dogfood log line added
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a real project. Open an obituary Source with one Artifact. The header shows Copy… and Delete, no Jump to graph (AC-1.5). Under Description / Credibility / Metadata, **Evidence graphs** shows "1 evidence graph" and one card, "Default", its `GPH-` ref, no description.
2. Click **Open**. The canvas opens. Breadcrumbs read Sources › {the obituary's title} › Default. Click the title crumb: the Source page opens.
3. From the canvas open a Citation composer: breadcrumbs read Sources › {title} › Default › Citation for …. Open Promote on a Subject: … › Default › Promote SUB-….
4. Back on the Source page click **Delete** on Default (it is enabled, AC-2.4): the notice reads "This is the source’s only evidence graph" (AC-2.3). Close it.
5. Rename the Source on its page, then open the canvas: the Source crumb shows the new title (AC-3.4).
6. Add a new Source with no Artifacts. Its Default card's Open is disabled; hovering shows "Add an artifact to open this evidence graph".
7. Relaunch on the canvas from step 2: the breadcrumbs show "Default" without visiting the Source page.
8. VoiceOver on the card: name, ref; Open reads "Open Default"; Delete reads "Delete Default". Narrow the window: the text wraps, the buttons keep their size.

The two-graph cases (AC-1.1, AC-2.1, AC-2.2) are covered by T-3, T-5 and T-6; they become manual once S12-12 can create a graph.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The toolbar reads the session environment where none is set | STEP-5 gives the preview a session; `WorkspaceView` sets it for the app. |
| `PlaceRegistryTests` assert exact key arrays | T-9 edits them in STEP-5. |
| D-16 changes a location an `EvidenceGraphModelTests` case compares | The constructor sets the same fields the copies set; STEP-5 runs that suite. |

**Stop and report if:**

- `GetDeleteImpact` for kind `evidence_graph` returns `not_found` or an `infra` gate for an existing graph (the engine path in §4 is not true).
- `deleteEvidenceGraph`'s effects do not refresh the graph list (the row stays after T-5's settle without hand-written cache writes).
- `WorkspaceToolbar` is presented anywhere without a `WorkspaceSession` in its environment other than its `#Preview`.
- A `.sourcePage(…)` replacement in D-16 would change a field a caller sets today.

## 12. Open questions

None.

## 13. PR

**Title:** S12-11 — Source page Graphs section
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-11-source-page-graphs.md`.
