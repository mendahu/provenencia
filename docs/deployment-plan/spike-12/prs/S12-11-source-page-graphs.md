# S12-11 — Source page Graphs section

**Status:** Blocked — on board S12-D2
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D2](../design/S12-D2-source-page-graphs.md), board not agreed yet
**Board:** the agreed Claude Design share link, recorded here when the board is agreed (as S12-D1). Until then, §6 macOS describes each frame in words, and the implementer builds from that.
**Depends on:** S12-05 (landed [#511](https://github.com/mendahu/provenencia/pull/511)), S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528)), S12-D2
**Blocks:** S12-12
**Branch:** `s12-11-source-page-graphs`, branched from `main` after the plan branch merges (so STEP-7 edits this file in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-10 did; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (place query keys only), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md) (breadcrumbs), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`
**Context:** [where graphs show](../source-structure.md#where-graphs-show), [rules](../source-structure.md#rules); working rules: "Source page", "One graph list", "A graph is a place", "Name and description are editable; graphs list alphabetically by name", "Refusals are of two kinds", "High tolerance for churn"; the S12-05 [Landed note](S12-05-graph-rpcs.md) (graph list Spec), the S12-08 [Landed note](S12-08-graph-id-end-to-end.md)
**Planned:** 2026-10-25 against `main` @ `dff91d9`
**Review:** [`S12-11-source-page-graphs.review.md`](S12-11-source-page-graphs.review.md)

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

The Source page shows every Evidence graph of its Source in a new **Evidence graphs** section: one row per graph, by name, with its name, `GPH-` ref, description, an **Open** button that opens that graph's canvas, and a **Delete** button that goes through the DeleteImpact recipe. The header's **Jump to graph** button is removed. The toolbar breadcrumbs on the canvas, the Citation composer and Promote name the graph: **Sources › {Source title} › {graph name}**. This is the first PR of the slice that makes several graphs per Source visible ([dogfood bar](../deployment-plan.md#goal-dogfood-bar) outcome 5, Source page half).

**Why now.** S12-05 put the graph list and graph delete on the wire and S12-08 made a graph a place, so the page can list and open graphs without engine work. S12-12 (Graph form) adds create and edit to this section and needs it first.

**What stays.** The engine, FFI and store protocol are unchanged. Every other Source page section is unchanged. On an existing project every Source has one graph named Default, so the page shows one row where the header button was, and its Open goes where Jump to graph went. The Sources list Open keeps opening the Source's first graph (S12-14 replaces it).

## 2. User stories

**US-1.** As a researcher on a Source page, I see the Source's Evidence graphs and open the one I want, so that I can work one household of a census without hunting for it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source with graphs "Smith household", "Default", "Jones household" | The section header reads "Evidence graphs" with meta "3"; rows read Default, Jones household, Smith household in that order, each with its ref and description |
| AC-1.2 | A graph named "Default" | The row shows "Default" as typed (no localized label) |
| AC-1.3 | Open on "Jones household", which links one Artifact | The canvas of that graph opens: `navigation.currentLocation.graphId` is its id, surface `.graph` |
| AC-1.4 | A Source with no Artifacts (its Default graph links none) | Open is disabled; tooltip and VoiceOver hint "Add an artifact to open this evidence graph" |
| AC-1.5 | The page's header | No Jump to graph button; Copy… and Delete remain |
| AC-1.6 | The graph list has no value yet | The section header shows with no meta and one skeleton row |
| AC-1.7 | The graph list read failed with no value | A danger callout in place of the rows; the other sections render |

**US-2.** As a researcher, I delete a graph I no longer need, and I'm told why when I can't.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Delete on an empty "Jones household" of a two-graph Source | The DeleteImpact confirm for kind `evidence_graph`, title "Delete this evidence graph?"; confirming calls `deleteEvidenceGraph` once, the row leaves, the toast reads "Evidence graph deleted" / "GPH-JONES was deleted from this source." |
| AC-2.2 | Delete on a graph with 3 Subjects | The blocked notice with the group "This graph has 3 subjects"; `deleteEvidenceGraph` is not called |
| AC-2.3 | Delete on a Source's only graph | The `last_graph` notice "This is the source's only evidence graph"; `deleteEvidenceGraph` is not called |
| AC-2.4 | Delete is visible on every row | It is never disabled (it is disabled only while the Source itself has no value, as the header Delete) |

**US-3.** As a researcher on a graph's canvas, composer or Promote page, the breadcrumbs name the Source and the graph, so that I know which household I'm in and can step back to the Source page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Canvas of "Jones household" on "1851 census" | Crumbs: Sources › 1851 census › Jones household. "1851 census" goes to the Source page; the last crumb has no action |
| AC-3.2 | Composer for scope "John Jones" on that graph | Crumbs: Sources › 1851 census › Jones household › Citation for John Jones; "Jones household" goes to that graph's canvas |
| AC-3.3 | Promote for `SUB-AAAAA` on that graph | Crumbs: Sources › 1851 census › Jones household › Promote SUB-AAAAA |
| AC-3.4 | The graph list has no row for the graph id (loading, or a deleted graph in history) | The graph crumb reads "Evidence graph" |
| AC-3.5 | Legacy history with no Source title | The Source crumb is left out: Sources › {graph name} › … |

## 3. Scope

| In | Covered by |
| --- | --- |
| Evidence graphs section between the overview columns and Artifacts, header with count (GS-1) | STEP-3, AC-1.1 |
| One card row per graph in list order: name, ref, description (GS-2, GS-10) | STEP-3, AC-1.1, AC-1.2 |
| Open per graph, disabled without a linked Artifact (GS-3, GS-4) | STEP-1, STEP-3, AC-1.3, AC-1.4 |
| Delete per graph through the DeleteImpact recipe, toast after (GS-3, GS-5, GS-6) | STEP-2, STEP-3, AC-2.1 – AC-2.4 |
| Loading and error states (GS-8) | STEP-3, AC-1.6, AC-1.7 |
| VoiceOver labels (GS-9) | STEP-3, manual check |
| Header Jump to graph removed with its model properties and string (GS-7) | STEP-4, AC-1.5 |
| `evidence_graph` delete-impact noun | STEP-2, AC-2.1 |
| Breadcrumbs Sources › Source › graph name on canvas, composer, Promote | STEP-5, AC-3.1 – AC-3.5 |
| Graph and Promote places warm the graph list | STEP-5, T-9 |
| The stored name `Default` shown verbatim (S12-03 review UX-3, user 2026-10-25) | D-6, AC-1.2 |

| Out | Home |
| --- | --- |
| Add graph and Edit on a row | S12-12 (board S12-D3) |
| Artifact link checkboxes | S12-13 (board S12-D4) |
| Per-graph progress; the Sources list per-graph Open (`SourcesListNavigation.entryGraph` stays) | S12-14 (board S12-D5) |
| Graphs in the omnibar | S12-15 |
| A Promote button per graph on the Source page (S12-09's "Not this PR" named S12-11) | not planned — Promote opens from the graph canvas; the working rule "Source page" lists name, ref, description, Open |
| Pruning history entries of a deleted graph | not planned — a deleted graph holds no Subjects; its history entry opens as any missing graph does today |
| A localized label for a graph named `Default` | not planned (D-6) |
| A kit list component with trailing actions | not planned — one call site (brief non-goals) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Header jump | `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `evidenceGraphJump` | `PVButton(L10n.Sources.jumpToEvidenceGraph, .secondary, .sm, icon: .gitBranch)` before Copy… and Delete; goes to `model.evidenceGraphLocation`; help / hint from `model.evidenceGraphBlockedReason`; id `sources.page.jumpToEvidenceGraph`. Removed by this PR |
| Page model | `macos/App/Features/Sources/SourcePageModel.swift` `evidenceGraphLocation`, `evidenceGraphBlockedReason` | The entry graph's location and blocked reason, read through `SourcesListNavigation`. Only the header reads them. Dead after STEP-4 |
| Page delete | `SourcePageModel.swift` `presentImpact(kind:id:ref:title:)` (private), `confirmResourceDelete()`, `deleteImpact: DeleteImpactFlow` | `askDeleteSource` / `askDeleteArtifact` call `presentImpact`; `confirmResourceDelete` branches on `target.kind` (`source`, else artifact). No graph branch |
| Page sections | `SourcePageModel` `identity`, `credibility`, `metadata`, `notes`, `artifacts` (each `Source*Section(context:)`) | Each section is a class holding a `SourcePageContext`. No graphs section |
| Page layout | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageContent.body` | `overviewColumns`, then `SourcePageArtifactsView(model:)`, then `SourcePageNotesView(model:)` in one `VStack(spacing: PVSpacing.space9)`; `.pvDeleteImpact(flow: model.deleteImpact, …)` and `.vocabularyToastOverlay($model.toast, …)` already on the page |
| Section shape | `macos/App/Features/Sources/SourcePageNotesView.swift`, `SourcePageArtifactsView.swift` | `PVSectionHeader(title:meta:)` then rows; Artifacts rows are `PVCard(elevated: false, padding: 0)` |
| Navigation helpers | `macos/App/Features/Sources/SourcesListNavigation.swift` `entryGraph`, `graphsBySource`, `graphBlockedReason`, `graphListLoaded`, `graphLocation(for:graphs:)` | `graphLocation(for:graphs:)` picks the entry graph and returns `.evidenceGraph(sourceId:graphId:ref:title:)` (title = Source title) when it links an Artifact. No per-graph form |
| Graph list | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` `evidenceGraphsKey`, `evidenceGraphsHandle` | The one cached graph list (`[CatalogEvidenceGraph]`, sorted by name from Go and by `FakeStore.graphListOrder`) |
| Graph type | `macos/App/Platform/GenealogyStore.swift` `CatalogEvidenceGraph` | `id`, `ref`, `sourceID`, `name`, `description`, `artifactIDs` |
| Graph delete | `GenealogyStore.deleteEvidenceGraph(projectDir:userID:graphID:) -> CatalogWriteEffects`; `GoStore`, `FakeStore` | Effects carry `lists: [.evidenceGraphs]`, which refreshes the graph list through `CatalogQueryRegistry`. `FakeStore.deleteEvidenceGraph` removes the row, with no call counter |
| Delete impact | `api/ffi/handlers/delete_impact.go` `GetDeleteImpact`; `core/database/catalogmodel/model.go` `KindEvidenceGraph = "evidence_graph"` | Kind `evidence_graph` is parsed; Subjects are an inbound group `subjects.graph_id` (heading "This graph has %lld subjects"); the last graph is gate `last_graph` (`L10n.DeleteImpact.gateLastGraphTitle` / `Body`) |
| Delete nouns | `macos/App/Platform/L10n.swift` `L10n.DeleteImpact.noun(_:count:)` | No `evidence_graph` case: the confirm falls back to "item" |
| FakeStore impact | `macos/App/Platform/FakeStore.swift` `getDeleteImpact`, `deleteImpactByID` | No `evidence_graph` branch; tests set `deleteImpactByID[id]` |
| Breadcrumbs | `macos/App/Features/Workspace/WorkspaceToolbar.swift` `breadcrumbItems(for:goTo:)` (static) | Graph surface: Sources › "Evidence graph". Composer / Promote: Sources › `evidenceGraphFor(sourceTitle:)` ("Evidence graph for {source}", → canvas) › Citation / Promote. No graph name. The toolbar has no `WorkspaceSession` environment read; `WorkspaceView` sets `.environment(session)` above it |
| Place keys | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` Specs `.sourceGraph`, `.sourcePromote` | Neither warms `.evidenceGraphsList`; `.sourceCitationComposer` and `.sourceDetail` do |
| Strings | `L10n.Sources.jumpToEvidenceGraph` (`sources.page.jumpToEvidenceGraph`), `L10n.Workspace.evidenceGraphFor(sourceTitle:)`, `L10n.Workspace.evidenceGraphTitle`, `L10n.Sources.needsArtifactTooltip` | `evidenceGraphTitle` is also the canvas title (`EvidenceGraphView.swift`) and stays. `needsArtifactTooltip` stays (Sources list) |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` (`askDeleteSourcePresentsAllowedImpact`, `blockedArtifactDoesNotCallDelete`), `SourcesListNavigationTests.swift` (`graphLocationOpensTheFirstGraph`, `graphLocationBlockedWithoutArtifact`), `WorkspaceToolbarBreadcrumbTests.swift` (`evidenceGraphLeafUsesGraphTitle`, `citationComposerIncludesNavigableEvidenceGraph`, `promoteIncludesNavigableEvidenceGraph`), `PlaceRegistryTests.swift` | The three breadcrumb tests assert today's labels and change in STEP-5 |

**Precedent.** S12-02 ([#485](https://github.com/mendahu/provenencia/pull/485)) added a Source page action (Copy…) with a model on `SourcePageModel` and its own view file; copy that split (state on the model, view in `SourcePage*View.swift`). Artifact delete in `SourcePageModel.askDeleteArtifact` / `confirmResourceDelete` is the delete precedent: `presentImpact`, then the store call, `session.apply(effects)` and a success toast. Do not copy Artifact delete's hand-written `setQueryValue` workspace patch: graph delete changes no workspace field, and the effects refresh the graph list.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A new section class `SourceGraphsSection` *(new)* in `Features/Sources/SourceGraphsSection.swift` *(new)* owns the rows and their locations: `rows: [CatalogEvidenceGraph]?` (nil while the graph list has no value; the Source's graphs in list order otherwise), `loadError: Error?` (the handle's error while it has no value), `location(for:)` and `blockedReason(for:)`. `SourcePageModel` gains `var graphs: SourceGraphsSection`. | The page's sections are classes on the model; S12-12 adds its form state to this class. | Computed properties on `SourcePageModel` — S12-12 would move them; a struct — the other sections are classes with a context. |
| D-2 | Rows filter the session's list: `evidenceGraphsHandle?.value?.filter { $0.sourceID == sourceID }`. No regroup, no new key. | Working rule "One graph list". The list is already sorted by name. | `SourcesListNavigation.graphsBySource(in:)` — builds every Source's groups to read one. |
| D-3 | `SourcesListNavigation.graphLocation(for source: CatalogSource, graph: CatalogEvidenceGraph) -> WorkspaceLocation?` (new) returns `.evidenceGraph(sourceId:graphId:ref:title:)` with the Source's ref and title, nil when `graph.artifactIDs` is empty. `graphLocation(for:graphs:)` calls it on the entry graph. | One rule for "a graph opens when it links an Artifact", shared with the Sources list now and S12-14 later. | A second copy of the rule in the section. |
| D-4 | The blocked reason on a row is a new string, `L10n.Sources.graphNeedsArtifactTooltip` (new) "Add an artifact to open this evidence graph". | `needsArtifactTooltip` says "on the source page" and "build its evidence graph", wrong on the page and for one of several graphs. | Reusing `needsArtifactTooltip`. |
| D-5 | Delete is always enabled (except while `model.source` is nil). It calls `SourcePageModel.askDeleteGraph(id:)` (new), which calls `presentImpact(kind: "evidence_graph", id:, ref:, title: graph.name)`. `confirmResourceDelete` gains an `evidence_graph` branch: `store.deleteEvidenceGraph`, `session.apply(effects)`, toast. It returns false (the page stays). | User 2026-10-25: the recipe explains, as Source and Artifact delete. One delete flow on the page. | Disabling Delete on the last graph; a separate `PVConfirm`. |
| D-6 | A graph's name is shown verbatim everywhere, including `Default`. | User 2026-10-25 (S12-03 review UX-3). Renaming is S12-12. | A localized label for `Default` — a special case in every place a name shows. |
| D-7 | Section placement: between `overviewColumns` and `SourcePageArtifactsView` in `SourcePageContent`. | User 2026-10-25. | Top of page; below Artifacts. |
| D-8 | Loading is one `PVSkeletonBar` row in a `PVCard`; error is `PVCallout(tone: .danger, message: L10n.Errors.message(for:))`. No retry button: the session revalidates the list on the next visible read, as the Sources list does. | Brief GS-8; the kit. | A spinner; a Read again button (the list Spec is shared, not the page's). |
| D-9 | Breadcrumbs: `breadcrumbItems(for:graphName:goTo:)` takes `graphName: (String) -> String?` (new parameter, default `{ _ in nil }`). For the graph, composer and Promote surfaces with a `graphId`: a Source crumb (label `location.title` for `.graph`, `location.sourceTitle` for composer and Promote; goes to `SourcesListNavigation.pageLocation`-shaped `WorkspaceLocation(section: .sources, sourceId:, sourceSurface: .page, title:)`), left out when that title is empty; then the graph crumb (label `graphName(graphId) ?? L10n.Workspace.evidenceGraphTitle`; on the graph surface it is the leaf with no action, on composer and Promote it goes to `.evidenceGraph(sourceId:graphId:title: sourceTitle)` as today); then Citation / Promote as today. `WorkspaceToolbar` reads `@Environment(WorkspaceSession.self)` and passes `{ id in session.evidenceGraphsHandle?.value?.first { $0.id == id }?.name }`. | User 2026-10-25 (Sources › Source › graph). Working rule "One graph list": breadcrumbs read the list, so a rename (S12-12) shows at once. The static function stays testable without a session. | A `graphName` field denormalized on `WorkspaceLocation` — goes stale on rename and changes history coding. |
| D-10 | `PlaceRegistry` `.sourceGraph` and `.sourcePromote` add `.evidenceGraphsList(project: project)` to their keys. | A restored canvas or Promote page shows its graph name without first visiting the Source page. The composer and Source page already warm it. | Reading the name lazily with a fallback forever. |
| D-11 | `L10n.Workspace.evidenceGraphFor(sourceTitle:)` and `L10n.Sources.jumpToEvidenceGraph` are deleted with their `Localizable.xcstrings` entries. `SourcePageModel.evidenceGraphLocation` and `evidenceGraphBlockedReason` are deleted. | Working rule "High tolerance for churn": no dead paths. | Keeping them for S12-14 — S12-14 reads `SourcesListNavigation`, not the page model. |
| D-12 | The delete-impact noun for `evidence_graph` is `deleteImpact.kind.evidence_graph.one` "evidence graph" / `.other` "evidence graphs", in `L10n.DeleteImpact.noun`. | The recipe's confirm title and notice title read the noun. | The "item" fallback. |
| D-13 | The toast: `L10n.Sources.toastGraphDeletedTitle` (new) "Evidence graph deleted", `L10n.Sources.toastGraphDeletedBody(ref:)` (new) "%@ was deleted from this source." | Brief GS-6; Artifact delete's toast shape. | No toast — the row leaving is quieter than every other page delete. |
| D-14 | Tests set graph-delete impacts through `FakeStore.deleteImpactByID`. `FakeStore` gains `deleteEvidenceGraphCalls` (new, incremented in `deleteEvidenceGraph`). | Matches `deleteSourceCalls`; no fake delete-impact logic for one kind. | A FakeStore `evidence_graph` impact branch. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourceGraphsSection` *(new)* | type | `SourceArtifactsSection` (`Features/Sources/SourceArtifactsSection.swift`) | Same shape (a section class on a `SourcePageContext`); Artifacts' drafts and file staging do not apply. |
| `SourcePageGraphsView` *(new)*, private `SourcePageGraphRow` *(new)* | component (snowflake) | `SourcePageArtifactsView`, `SourcePageNotesView`; `PVList` | `PVList` rows are buttons with no trailing actions (its doc: "Rows are buttons, not selection"); Open and Delete need two controls. Brief inventory marks the row a snowflake. |
| `SourcesListNavigation.graphLocation(for:graph:)` *(new)* | func | `SourcesListNavigation.graphLocation(for:graphs:)` | Extends it: the old one calls the new one. |
| `SourcePageModel.askDeleteGraph(id:)` *(new)* | func | `askDeleteArtifact(id:)` | Same call into `presentImpact`, another kind. |
| `breadcrumbItems` `graphName:` parameter *(new)* | func param | `breadcrumbItems(for:goTo:)` | Extends it. |
| `L10n.Sources.graphsHeading`, `openEvidenceGraph`, `openEvidenceGraphLabel(name:)`, `deleteEvidenceGraphLabel(name:)`, `graphNeedsArtifactTooltip`, `toastGraphDeletedTitle`, `toastGraphDeletedBody(ref:)` *(new)*; `L10n.DeleteImpact.nounEvidenceGraph(s)` *(new)* | key | `L10n.Sources.artifactsHeading`, `toastArtifactDeleted*`, `L10n.DeleteImpact.noun*` | New copy in existing namespaces. |
| `FakeStore.deleteEvidenceGraphCalls` *(new)* | helper | `FakeStore.deleteSourceCalls` | Same counter pattern. |

**Builds on.** The page's `DeleteImpactFlow` and `pvDeleteImpact`; the session's one graph list (`evidenceGraphsHandle`); `PlaceRegistry` keys; `PVBreadcrumbs`; `SourcesListNavigation` as the one home of the "opens when it links an Artifact" rule.
**Deletes.** `SourcePageIdentityHeader.evidenceGraphJump`; `SourcePageModel.evidenceGraphLocation`, `evidenceGraphBlockedReason`; `L10n.Sources.jumpToEvidenceGraph`; `L10n.Workspace.evidenceGraphFor(sourceTitle:)`; their `Localizable.xcstrings` entries; the `sources.page.jumpToEvidenceGraph` identifier.
**Next adds.** S12-12 adds Add graph to the section header's `actions` slot, an Edit button to `SourcePageGraphRow` and its form state to `SourceGraphsSection`. S12-14 reuses `SourcesListNavigation.graphLocation(for:graph:)` for each accordion row and then deletes `entryGraph`.

### Schema

None — no catalog change.

### Go core

None — `evidencegraphs.List`, `Delete` and `deleteimpact` already do what the page needs.

### FFI

None — `ListEvidenceGraphs`, `DeleteEvidenceGraph` and `GetDeleteImpact` (kind `evidence_graph`) exist since S12-05.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Section title | `PVSectionHeader(title: L10n.Sources.graphsHeading, meta:)` | `Features/Sources/SourcePageGraphsView.swift` *(new)* | `meta` = `"\(rows.count)"` when `rows` has a value, nil while loading. `actions` empty (S12-12) |
| Graph row | `SourcePageGraphRow` (private snowflake) in `PVCard(elevated: false)` | same file | `HStack(alignment: .top)`: a `VStack(alignment: .leading)` of name (`PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.semibold)`, `lineLimit(2)`, `Text(verbatim:)`), ref (`PVFont.mono(size: PVTypeScale.caption)`, `PVColor.textMuted`, `Text(verbatim:)`), description when non-empty after trimming (`PVFont.body(size: PVTypeScale.bodySmall)`, `PVColor.textMuted`, `lineLimit(3)`, `Text(verbatim:)`); then Open and Delete |
| Open | `PVButton(L10n.Sources.openEvidenceGraph, variant: .secondary, size: .sm, icon: .gitBranch)` | same | `navigation.go(to:)` with `model.graphs.location(for:)`; `.disabled(location == nil)`; `.help` and `.accessibilityHint` from `blockedReason(for:)`; `.accessibilityLabel(Text(L10n.Sources.openEvidenceGraphLabel(name:)))`; id `sources.page.graph.open.<ref>` |
| Delete | `PVIconButton(.trash, label: L10n.Sources.deleteEvidenceGraphLabel(name:), size: .sm, tone: .danger)` | same | `Task { await model.askDeleteGraph(id:) }`; `.disabled(model.source == nil)`; id `sources.page.graph.delete.<ref>` |
| Loading | `PVSkeletonBar` in one `PVCard` | same | id `sources.page.graphs.loading` |
| Error | `PVCallout(tone: .danger, message: L10n.Errors.message(for: error))` | same | id `sources.page.graphs.error` |
| Delete flow | existing `.pvDeleteImpact(flow: model.deleteImpact, …)` | `SourcePageView.swift` | No change to the modifier |
| Section placement | — | `SourcePageView.swift` `SourcePageContent.body` | `SourcePageGraphsView(model: model)` between `overviewColumns` and `SourcePageArtifactsView(model: model)` |
| Section id | — | — | the section `VStack` has id `sources.page.graphs` |
| Header | remove `evidenceGraphJump` | `SourcePageIdentityHeader.swift` | The action `HStack` holds `copySourceButton`, `deleteSourceButton` |
| Breadcrumbs | `PVBreadcrumbs` (unchanged) | `Features/Workspace/WorkspaceToolbar.swift` | D-9 |

**Frames in words** (until the board link is recorded): frame 1 is the page with one Default card under the overview columns and above Artifacts, header with Copy… and Delete only; frame 2 three cards in name order with one-line descriptions; frame 3 the Default card with Open disabled; frames 4 – 6 the existing DeleteImpact confirm and notices; frame 7 two cards and the success toast; frame 8 the header with no count and one skeleton card, then a danger callout; frame 9 a two-line truncated name and a three-line truncated description. When the board is agreed and differs from this, the board's layout wins for spacing and placement within a card; the components above do not change.

States: loading (D-8), error (D-8), one graph, many graphs (no inner scroll), a graph with no linked Artifact (Open disabled). Accessibility: Open and Delete carry the graph's name in their labels; the row's text is read in order (name, ref, description).

### Strings

L10n keys *(new)*, final copy confirmed against the board in STEP-7:

- `L10n.Sources.graphsHeading` `sources.page.graphsHeading` = "Evidence graphs"
- `L10n.Sources.openEvidenceGraph` `sources.page.graph.open` = "Open"
- `L10n.Sources.openEvidenceGraphLabel(name:)` `sources.page.graph.openLabel` = "Open %@"
- `L10n.Sources.deleteEvidenceGraphLabel(name:)` `sources.page.graph.deleteLabel` = "Delete %@"
- `L10n.Sources.graphNeedsArtifactTooltip` `sources.page.graph.needsArtifact` = "Add an artifact to open this evidence graph"
- `L10n.Sources.toastGraphDeletedTitle` `sources.page.toastGraphDeletedTitle` = "Evidence graph deleted"
- `L10n.Sources.toastGraphDeletedBody(ref:)` `sources.page.toastGraphDeletedBody` = "%@ was deleted from this source."
- `L10n.DeleteImpact.nounEvidenceGraph` `deleteImpact.kind.evidence_graph.one` = "evidence graph"; `nounEvidenceGraphs` `deleteImpact.kind.evidence_graph.other` = "evidence graphs"

Removed: `sources.page.jumpToEvidenceGraph`, `workspace.section.evidenceGraph.forSource`.

## 7. Steps

**STEP-1 — One per-graph open rule**
Files: `macos/App/Features/Sources/SourcesListNavigation.swift`, `macos/ProvenenciaTests/SourcesListNavigationTests.swift`. Skill: `add-swift-test`.
Do: add `graphLocation(for:graph:)` (D-3); make `graphLocation(for:graphs:)` call it on `entryGraph`. Add T-1.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests/SourcesListNavigationTests` passes, the existing tests unchanged.

**STEP-2 — Graphs section model and graph delete**
Files: `macos/App/Features/Sources/SourceGraphsSection.swift` *(new)*, `SourcePageModel.swift`, `macos/App/Platform/L10n.swift` (`DeleteImpact` nouns, toast keys, `graphNeedsArtifactTooltip`), `macos/App/Resources/Localizable.xcstrings`, `macos/App/Platform/FakeStore.swift` (`deleteEvidenceGraphCalls`), `macos/ProvenenciaTests/SourcePageModelTests.swift`. Skills: `add-localized-string`, `add-swift-test`. Add the new file to the Provenencia target in `macos/Provenencia.xcodeproj` as other `Features/Sources` files are.
Do: D-1, D-2, D-5, D-12, D-13, D-14. `rows` reads `context.session.evidenceGraphsHandle`; `location(for:)` uses `context.source` and D-3; `blockedReason(for:)` returns `graphNeedsArtifactTooltip` when `location(for:)` is nil and nil otherwise. Add T-2 – T-6.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests/SourcePageModelTests` passes; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — Graphs section view**
Files: `macos/App/Features/Sources/SourcePageGraphsView.swift` *(new)*, `SourcePageView.swift`, `L10n.swift`, `Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`. Add the file to the target.
Do: the §6 macOS table and D-7, D-8. Strings as §6.
Check: the app builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`); on a FakeStore preview or the dev project the section shows one Default card above Artifacts.

**STEP-4 — Remove the header jump**
Files: `SourcePageIdentityHeader.swift`, `SourcePageModel.swift`, `L10n.swift`, `Localizable.xcstrings`.
Do: D-11 for the header and page model: delete `evidenceGraphJump`, `evidenceGraphLocation`, `evidenceGraphBlockedReason`, `L10n.Sources.jumpToEvidenceGraph` and its catalog entry.
Check: `rg -n 'jumpToEvidenceGraph|evidenceGraphLocation|evidenceGraphBlockedReason' macos` prints nothing; build passes.

**STEP-5 — Breadcrumbs name the graph**
Files: `macos/App/Features/Workspace/WorkspaceToolbar.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `L10n.swift`, `Localizable.xcstrings`, `macos/ProvenenciaTests/WorkspaceToolbarBreadcrumbTests.swift`, `macos/ProvenenciaTests/PlaceRegistryTests.swift`. Skills: `add-workspace-location`, `add-workspace-place`.
Do: D-9, D-10, and the `evidenceGraphFor` half of D-11. Update the doc comment on `breadcrumbItems` to the new shapes. Give the `#Preview` a `WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/preview.provenencia"), store: FakeStore())` environment. Rewrite the three graph breadcrumb tests as T-7 and add T-8, T-9.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO -only-testing:ProvenenciaTests/WorkspaceToolbarBreadcrumbTests -only-testing:ProvenenciaTests/PlaceRegistryTests` passes; `rg -n 'evidenceGraphFor' macos` prints nothing.

**STEP-6 — Full check**
Files: none new.
Do: run the whole gate.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status, Landed note, Strings from the board), `docs/deployment-plan/spike-12/design/S12-D2-source-page-graphs.md` → `design/archive/`, `docs/deployment-plan/spike-12/design/README.md`.
Do: per the Definition of done. In `SourcesListNavigation.entryGraph`'s doc comment, "(S12-11, S12-14)" becomes "(S12-14)". In working rule "A graph is a place", "Until S12-11 and S12-14 add per-graph Open, the Source page jump and the Sources list Open" becomes "Until S12-14 adds per-graph Open, the Sources list Open".
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-11-source-page-graphs.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesListNavigationTests.swift` `graphLocationOpensTheGivenGraph` *(new)* | Source `src-1`, graph `g-2` with one Artifact → `.evidenceGraph(sourceId: "src-1", graphId: "g-2", …)` with the Source's ref and title; same graph with no Artifact → nil | AC-1.3, AC-1.4, D-3 |
| T-2 | Swift | `SourcePageModelTests.swift` `graphRowsAreTheSourcesGraphsByName` *(new)* | FakeStore graphs "Smith household", "Default", "Jones household" on `src-1` and one on `src-2`; warm → `model.graphs.rows?.map(\.name)` == ["Default", "Jones household", "Smith household"] | AC-1.1, AC-1.2, D-2 |
| T-3 | Swift | `SourcePageModelTests.swift` `graphRowsNilUntilListLoadsAndErrorWhenItFails` *(new)* | Before warm → `rows == nil`; `listEvidenceGraphsError` set, warm → `rows == nil`, `loadError != nil` | AC-1.6, AC-1.7 |
| T-4 | Swift | `SourcePageModelTests.swift` `graphWithoutArtifactIsBlocked` *(new)* | Default graph with no Artifacts → `location(for:) == nil`, `blockedReason(for:) == graphNeedsArtifactTooltip` | AC-1.4, D-4 |
| T-5 | Swift | `SourcePageModelTests.swift` `deleteEmptyGraphCallsStoreAndToasts` *(new)* | Two graphs; `deleteImpactByID["g-jones"]` allowed; `askDeleteGraph(id: "g-jones")` → target kind `evidence_graph`; `confirmResourceDelete()` → returns false, `deleteEvidenceGraphCalls == 1`, toast title "Evidence graph deleted", rows no longer include `g-jones` | AC-2.1, D-5, D-13 |
| T-6 | Swift | `SourcePageModelTests.swift` `blockedGraphDoesNotCallDelete` *(new)*, parameterized over (`.inbound` with group `subjects.graph_id` total 3) and (`.lastGraph`, no groups) | `allowed == false` → after `confirmResourceDelete()`, `deleteEvidenceGraphCalls == 0` and the gate is the given one | AC-2.2, AC-2.3 |
| T-7 | Swift | `WorkspaceToolbarBreadcrumbTests.swift` `evidenceGraphCrumbsNameSourceAndGraph`, `citationComposerCrumbsNameSourceAndGraph`, `promoteCrumbsNameSourceAndGraph` (replace the three graph tests) | Lookup `g-1` → "Jones household"; canvas `title: "1851 census"` → labels [Sources, 1851 census, Jones household], crumb 1 goes to the Source page, last has no action; composer → […, Jones household, Citation for John Jones], crumb 2 goes to `.graph` `g-1`; Promote → […, Promote SUB-AAAAA] | AC-3.1, AC-3.2, AC-3.3, D-9 |
| T-8 | Swift | `WorkspaceToolbarBreadcrumbTests.swift` `graphCrumbFallsBackWithoutNameOrSourceTitle` *(new)* | Lookup returns nil → graph crumb "Evidence graph"; composer with no `sourceTitle` → no Source crumb | AC-3.4, AC-3.5 |
| T-9 | Swift | `PlaceRegistryTests.swift` `graphAndPromoteWarmTheGraphList` *(new)* | Resolve a `.graph` and a `.promote` location → query keys contain `.evidenceGraphsList(project:)` | D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `WorkspaceToolbarBreadcrumbTests`, `SourcePageModelTests`, `PlaceRegistryTests` (exact key lists), `check-localizable-xcstrings.py` (removed keys).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (jumpToEvidenceGraph, evidenceGraph.forSource removed)
- [ ] UI matches the brief inventory; no new kit primitive
- [ ] Version bump: none (VERSION stays 0.2.0)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; working rule "A graph is a place" reworded (STEP-7); completed.md: one short paragraph of what shipped
- [ ] Brief S12-D2 moved to design/archive/ and design/README.md updated
- [ ] Domain doc: none — no rule in docs/*.md changes
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a real project. Open an obituary Source with one Artifact. The header shows Copy… and Delete, no Jump to graph (AC-1.5). Under Description / Credibility / Metadata, **Evidence graphs** shows "1" and one card, "Default", its `GPH-` ref, no description.
2. Click **Open**. The canvas opens. Breadcrumbs read Sources › {the obituary's title} › Default. Click the title crumb: the Source page opens.
3. From the canvas open a Citation composer: breadcrumbs read Sources › {title} › Default › Citation for …. Open Promote on a Subject: … › Default › Promote SUB-….
4. Back on the Source page click **Delete** on Default (it is enabled, AC-2.4): the notice reads "This is the source’s only evidence graph" (AC-2.3). Close it.
5. Add a new Source with no Artifacts. Its Default card's Open is disabled; hovering shows "Add an artifact to open this evidence graph".
6. Relaunch on the canvas from step 2: the breadcrumbs show "Default" without visiting the Source page.
7. VoiceOver on the card: name, ref; Open reads "Open Default"; Delete reads "Delete Default".

The two-graph cases (AC-1.1, AC-2.1, AC-2.2) are covered by T-2, T-5 and T-6; they become manual once S12-12 can create a graph.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The toolbar reads the session environment in a context without one (preview, a test host) and crashes | STEP-5 gives the preview a session; `WorkspaceView` sets it for the app. |
| `PlaceRegistryTests` assert exact key arrays | Update their expectations in STEP-5 to include the graph list key. |
| A graph delete leaves the canvas of that graph open in another history entry | Out (not planned); a deleted graph has no Subjects. |

**Stop and report if:**

- `GetDeleteImpact` for kind `evidence_graph` returns `not_found` or an `infra` gate for an existing graph on the dev project (the engine path in §4 is not true).
- `deleteEvidenceGraph`'s effects do not refresh the graph list (the row stays after T-5's confirm without hand-written cache writes).
- `WorkspaceToolbar` is presented anywhere without a `WorkspaceSession` in its environment other than its `#Preview`.
- The agreed board places the section elsewhere than D-7, adds a control not in §6, or uses a component not in the brief's inventory.

## 12. Open questions

None.

## 13. PR

**Title:** S12-11 — Source page Graphs section
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-11-source-page-graphs.md`.
