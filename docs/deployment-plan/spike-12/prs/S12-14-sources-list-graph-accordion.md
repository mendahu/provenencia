# S12-14 — Graph counts in the Sources list and on the Source page

**Status:** Blocked — on board S12-D5
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D5](../design/S12-D5-sources-list-graph-accordion.md), board not agreed yet
**Board:** — (to be recorded when the board is agreed: the Claude Design link, and §6 Frames describes each frame in words, as S12-13 did)
**Depends on:** S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528)), S12-11 (landed [#549](https://github.com/mendahu/provenencia/pull/549)), S12-13 (landed [#562](https://github.com/mendahu/provenencia/pull/562): `evidenceGraphsBySource`), S12-D5
**Blocks:** —
**Branch:** `s12-14-sources-list-graph-accordion`, branched from `main` after the plan branch merges
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-13 did; Swift only, no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (place query keys, session cache reads)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [where graphs show](../source-structure.md#where-graphs-show); working rules: "Sources list", "Source page", "One graph list", "Name and description are editable; graphs list alphabetically by name", "A graph is a place", "High tolerance for churn"; the S12-11 plan D-2 (session reads of the graph list), the S12-13 [Landed note](S12-13-artifact-graph-checkboxes.md) (`evidenceGraphsBySource`, `EvidenceGraphIndex`)
**Planned:** 2026-10-26 against `main` @ `e3b85f9`
**Review:** [`S12-14-sources-list-graph-accordion.review.md`](S12-14-sources-list-graph-accordion.review.md) — Fits with changes / Ready with changes, 2 rounds; every blocker Fixed; M-5 To user

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

A researcher sees every Evidence graph of every Source from the Sources list: each row expands to one sub-row per graph, in name order, with that graph's own counts and its own Open. The Source page's graph cards show the same count line. The Status filter reads each graph, and In progress now means "has Subjects, none promoted yet": a graph that has been through a Promote with at least one Subject filed counts as Promoted.

**Why now.** S12-11 to S12-13 let a Source hold several graphs, but the list still opens and counts only the alphabetically first one (the interim rule in working rule "A graph is a place"), and the Source page shows no progress. The status change is the product owner's call at planning (2026-10-26): today's In progress (`filedCount < primaryCount`) keeps a graph "in progress" forever when one Subject is left unfiled on purpose.

**What stays.** The row's page zone, right-click Copy… and click. The graph zone's 248 pt column and hover. A Source with no Artifact keeps its "Needs an artifact" zone. Search matches Source title and ref only. No engine, FFI or catalog change: the counts are today's `CatalogQueryKey.evidenceGraphProgress` map.

## 2. User stories

**US-1.** As a researcher, when I expand a Source in the Sources list, I see each of its Evidence graphs with its own counts and can open any of them, so that I can reach the Jones household without opening the Source page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A census Source with graphs "Smith household", "Default", "Jones household", all linked to an Artifact; the list loaded | The row's graph zone reads "3 evidence graphs" with a down chevron; no sub-rows show |
| AC-1.2 | AC-1.1, click the graph zone | Three sub-rows follow the row in the order Default, Jones household, Smith household; the chevron points up; clicking again removes them |
| AC-1.3 | Jones household has 12 Subjects and 40 Observations; Default has none | Jones's sub-row reads "Open graph" / "12 subjects · 40 observations"; Default's reads "Open graph" / "0 subjects · not started" |
| AC-1.4 | Click Jones household's Open zone | The canvas opens on Jones household's graph id (`WorkspaceLocation.evidenceGraph(sourceId:graphId:ref:title:)` with that id) |
| AC-1.5 | A fourth graph "Unlinked" links no Artifact | Its sub-row's Open zone is disabled with the help and hint "Add an artifact to open this evidence graph"; its count line still shows |
| AC-1.6 | A one-graph obituary | The zone reads "1 evidence graph"; expanding shows one "Default" sub-row with its Open zone |
| AC-1.7 | A Source with no Artifact | The zone is today's "Needs an artifact" with its tooltip; nothing expands |
| AC-1.8 | The graph list has not loaded | The zone shows a skeleton bar in place of its label and does not toggle; when the list fails, the zone shows a danger mark whose help is `L10n.Errors.message(for:)` of that failure |
| AC-1.9 | Expand the census, open a Source page, press Back | The census is still expanded; after a relaunch every row is collapsed |
| AC-1.10 | The census expanded, add an Observation on Jones's canvas, return to the list | Only Jones's count line changes ("12 subjects · 41 observations"), with no list reload |

**US-2.** As a researcher, when I filter the Sources list by status, a Source shows when any of its graphs has that status, and the statuses mean Empty / In progress / Promoted as defined below.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A graph with 0 Subjects | Its status is Empty |
| AC-2.2 | A graph with 5 Subjects, 3 primary, 0 filed | In progress |
| AC-2.3 | A graph with 5 Subjects, 3 primary, 1 filed | Promoted (today it reads In progress) |
| AC-2.4 | A graph with 2 Subjects, 0 primary (only value Subjects) | In progress |
| AC-2.5 | Census: Default Empty, Jones In progress, Smith Promoted | The census shows under Empty, under In progress and under Promoted |
| AC-2.6 | The Status menu | Options read All statuses, Empty, In progress, Promoted |
| AC-2.7 | A settings file saved with `"status":"filed"` before this PR | The list restores with Promoted chosen |

**US-3.** As a researcher, on a Source page, each Evidence graphs card shows that graph's counts.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The census page; Jones has 12 Subjects and 40 Observations | Jones's card shows "12 subjects · 40 observations" under its ref; Default's shows "0 subjects · not started" |
| AC-3.2 | The page opened straight from a relaunch (the list never shown) | The count lines load (the place warms the progress key) |
| AC-3.3 | Add an Observation on Jones's canvas, Back to the page | Only Jones's card count moves, with no page reload |

## 3. Scope

| In | Covered by |
| --- | --- |
| One session read of graph progress, and one "awaiting first value" check on `QueryHandle` (D-13, D-14) | STEP-1, T-10 |
| Status rule per graph, any-graph Source match, Promoted label and stored value (GA-8, D-5 – D-7) | STEP-2, AC-2.1 – AC-2.7 |
| One shared count line view with one spoken form (D-3) | STEP-3, AC-1.3, AC-3.1 |
| Session-held expanded state (GA-7, D-4) | STEP-4, AC-1.9 |
| Graph zone expander on every row with an Artifact (GA-1, GA-6, D-12) | STEP-4, AC-1.1, AC-1.2, AC-1.8 |
| "Needs an artifact" row unchanged (GA-5) | STEP-4, AC-1.7 |
| The interim entry-graph helpers deleted; one row call site (D-9, D-15) | STEP-4 |
| Graph sub-rows with per-graph Open and count line (GA-2 – GA-4, GA-9, GA-10) | STEP-5, AC-1.2 – AC-1.6 |
| Source page cards show the count line; the place warms the progress key (GA-11, D-8) | STEP-6, AC-3.1 – AC-3.3 |

| Out | Home |
| --- | --- |
| Graph search in the list; omnibar graph hits | S12-15 |
| Create, edit, delete or link graphs from the list | not planned (Source page owns them) |
| A Promote button, a status badge, or filed / primary counts on a row, sub-row or card | not planned (map scope, "Not this PR") |
| A Source-level total of Subjects or Observations | not planned (map scope) |
| Persisting expanded rows across relaunch or in history | not planned (user, 2026-10-26) |
| A kit accordion or tree component | not planned (brief non-goals) |
| Moving the status rule into `graphprogress` | when a second reader needs it (review M-5); logged in the review log |
| `SourcesModel.warmListQueries` warming the keys the place already warms | not planned here (review log, "Log it") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow` | Page zone + one graph zone. Takes one `progress`, `countsLoading`, `onOpenGraph`, `graphBlockedReason`. `graphZone` draws "Open graph" + `graphCountLine` for the Source, or the sunken "Needs an artifact" zone when `!source.hasArtifact`. The page zone's `CachedThumbnail` uses its default `size` (44). |
| Count line | same file, private `graphCountLine` | Three states: hairline while `countsLoading && progress == nil`; "0 subjects · not started" when `subjectCount == 0`; else `L10n.Sources.graphCountLine`. Private, one caller. |
| Spoken count | `macos/App/Platform/L10n.swift` `L10n.Sources.graphZoneAccessibility(progress:countsLoading:)` | Repeats the same three branches with "Open graph" in front. |
| Zone style | `SourcesSplitRow.swift` private `SourcesSplitZoneButtonStyle` | Hover fill and muted → primary label lift. |
| List wiring | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListView`, `SourcesListContent` (`progressHandle` parameter), `listBody`, `SourcesSplitRowProgress` | `listBody` has two `SourcesSplitRow(...)` constructions: `SourcesSplitRowProgress` when the progress handle exists, an inline one with `countsLoading: true` when not. Each row is one `ForEach(rows)` element with `.id(source.id)`. |
| Status menu | same file, `statusMenu` | Options from `SourcesModel.EvidenceStatus.allCases`; identifiers `sources.status.<rawValue>`. |
| Model | `macos/App/Features/Sources/SourcesModel.swift` `EvidenceStatus` | Cases `empty`, `inProgress = "in_progress"`, `filed`; `init(_ progress:)`: empty when `primaryCount == 0`, in progress when `filedCount < primaryCount`, else filed. `.filed` labels `L10n.Sources.statusFullyPromoted`. |
| Model | same file, `evidenceGraphProgressKey(for:)`, private `progressHandle`, `graphProgress(for:)`, `graphCountsLoading`, `isAwaitingFilterData`, `statusCountsError`, `warmListQueries` | Progress key and handle looked up here. `graphCountsLoading` and `isAwaitingFilterData` hand-write "no value and not settled" checks. |
| Model | same file, `entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)`, `visibleSources` | The interim "first graph" rule; the status filter reads the entry graph only. |
| Settings | `macos/App/Features/Sources/SourcesListSettings.swift` `SourcesListSettings.init(from:)` | Decodes each field on its own; `status` by raw value (`"filed"` stored today). |
| Navigation | `macos/App/Features/Sources/SourcesListNavigation.swift` | `entryGraph(sourceID:in:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`, `graphLocation(for:graphs:)` (interim, list-only); `graphLocation(for:graph:)` and `graphBlockedReason(for:)` (per graph; used by `SourceGraphsSection`). |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` | `graphCameras` is private UI state, "quit drops it", reached through `graphCamera(for:)` / `setGraphCamera(_:for:)`. `refreshEvidenceGraphProgress` and `patchEvidenceGraphProgress` build `CatalogQueryKey.evidenceGraphProgress(project:)` inline; the patch runs only when the map is already cached. |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` | "The one place" the graph list's key and handle are looked up: `evidenceGraphsKey`, `evidenceGraphsHandle`, `evidenceGraphs(sourceID:)`, `evidenceGraphsBySource`. |
| Query handle | `macos/App/Features/Workspace/Session/QueryHandle.swift` `QueryHandle` | `status`, `value`, `error`, `isSettled`; no "awaiting first value". |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `.sourcesList`, `.sourceDetail` | The list place warms `.evidenceGraphProgress(project:)`; the Source page place does not. |
| Source page | `macos/App/Features/Sources/SourceGraphsSection.swift` `SourceGraphsSection` | `rows`, `loadError`, `location(for:)`, `blockedReason(for:)`; no progress read. |
| Source page | `macos/App/Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Card: name, ref, description (combined accessibility element); Open, Edit, Delete. |
| Source page | `macos/App/Features/Sources/SourcePageModel.swift` `warmFromSession()` | Tests and previews warm a hand list of keys; navigation warms through `PlaceRegistry`. |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Sources` | `graphsCount(_:)` (plural, key `sources.page.graphsCount`), `openGraph`, `graphCountLine`, `graphSubjectCount`, `graphNotStarted`, `graphCountsLoading`, `needsArtifact`, `needsArtifactTooltip`, `graphNeedsArtifactTooltip`, `statusFullyPromoted` (`sources.list.status.fullyPromoted`, "Fully promoted"). |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` | `evidenceStatusFromCounts`, `filtersCombineAndClear` (seeds `primary: 2, filed: 1`, expects In progress), `settingsRestoreOnNewModel` (`.filed`), `statusCountsFailureSurfacesAnErrorAndUnfilteredCount` (`.filed`), `statusFilterWaitsForCounts`, `graphProgressCopyMatchesBoard`, `warmListQueriesLoadsGraphProgress`; helper `progress(_:primary:filed:)` sets `subjectCount = primary`. |
| Tests | `macos/ProvenenciaTests/SourcesListNavigationTests.swift` | `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`, `graphLocationBlockedWithoutArtifact` and the private helper `graphsBySource(_:)` (interim helpers); `graphLocationOpensTheGivenGraph` (stays). |
| Tests | `macos/ProvenenciaTests/PlaceRegistryTests.swift` | `resolvesSourceDetail` and the place table assert the exact `.sourceDetail` key list. |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` | Graph section tests; helper `entryGraphLocation(_:)` (reads `model.graphs.location(for:)`, stays); `session.apply(location:)` warms place keys in some tests. `FakeStore.graphProgressByGraph` seeds progress. |

**Precedent.** S12-11 ([#549](https://github.com/mendahu/provenencia/pull/549)) put the Source page's graph rows on session reads (`SourceGraphsSection.rows` over `evidenceGraphs(sourceID:)`) instead of a page copy: progress gets the same session home. `graphCameras` is the precedent for UI state held by the session until quit, behind methods. S12-13 ([#562](https://github.com/mendahu/provenencia/pull/562)) built `evidenceGraphsBySource`, which the sub-rows read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Every row with an Artifact has the expander; its label is `L10n.Sources.graphsCount(n)`. A Source with no Artifact keeps today's "Needs an artifact" zone with no expander. | User, 2026-10-26: every row expands. A Source with no Artifact has only an unopenable, empty Default graph; the existing zone already says what to do. | One-graph rows keep a direct Open — user chose every row. An expander on the no-Artifact row — one disabled sub-row says less than today's zone. |
| D-2 | Sub-rows are drawn under their row inside the same `ForEach` element: `SourcesListRow` *(new, replaces `SourcesSplitRowProgress`)* is a `VStack(spacing: 0)` of the row and, when expanded, its sub-rows; `listBody` puts `.id(source.id)` on it. | `ListScrollAnchorModifier` anchors on Source ids; a sub-row is not a list row of its own. | Sub-rows as separate `ForEach` items — breaks the scroll anchor's id list. |
| D-3 | `SourcesGraphCountLine` *(new)*, `macos/App/Features/Sources/SourcesGraphCountLine.swift`: `struct SourcesGraphCountLine: View { var progress: EvidenceGraphProgress?; var countsLoading: Bool }`, body moved from `SourcesSplitRow.graphCountLine`, plus `.accessibilityElement(children: .ignore)` and `.accessibilityLabel(Text(verbatim: L10n.Sources.graphCountLineSpoken(progress:countsLoading:)))`. `L10n.Sources.graphCountLineSpoken` *(new func, no new key)* holds the three branches now in `graphZoneAccessibility` minus "Open graph"; `graphZoneAccessibility` becomes `"\(openGraph), \(graphCountLineSpoken(...))"`. | One count rule for sight and VoiceOver (review M-2); the page card speaks "counts loading" too. | Copying the three states into the page card; a second spoken copy. |
| D-4 | Expanded state is private `expandedSourceRows: Set<String>` *(new)* on `WorkspaceSession`, observed (no `@ObservationIgnored`), beside `graphCameras`, reached through `isSourceRowExpanded(_ sourceID:) -> Bool` and `toggleSourceRowExpanded(_ sourceID:)` *(new)*. Only `SourcesListRow` reads it. The toggle does not animate. Deleted Source ids may stay in the set. | User, 2026-10-26: until relaunch. The list's `SourcesModel` is rebuilt with the view; the session lives for the project. Method-shaped like `graphCamera(for:)` (review M-7). | `SourcesListSettings` (persists, user declined); `WorkspaceLocation` (history, user declined); `@State` on the view (lost on leave); a public `Set` property. |
| D-5 | Graph status is read from today's counts: `.empty` when `subjectCount == 0`; `.promoted` when `filedCount > 0`; else `.inProgress`. A graph missing from the progress map is empty. | User, 2026-10-26: In progress = at least one Subject and none promoted; one filed Subject makes it done. `filedCount` is "primary Subjects with an accepted Identity Claim", the only promoted signal the map carries. `subjectCount` is what the count line shows, so a graph reading "0 subjects" is Empty. | `primaryCount == 0` as Empty — a graph with only value Subjects would read Empty beside "2 subjects". A new "ever promoted" engine field — `filedCount > 0` is that fact today; unpromoting every Subject returns the graph to In progress, which matches "none promoted". |
| D-6 | The case `.filed` becomes `.promoted` with raw value `"promoted"`. `SourcesListSettings.init(from:)` decodes the status as a `String` and maps a stored `"filed"` to `.promoted` before `EvidenceStatus(rawValue:)`; any other unknown value falls back to nil as today. | One word in code, storage and UI identifiers (`sources.status.promoted`); the legacy alias lives where stale settings are already handled (review M-6). AC-2.7. | Keeping raw `"filed"` — two names for one status. Dropping saved filters — needless. |
| D-7 | A Source matches a status when any graph of `session.evidenceGraphsRead(sourceID:)`'s `.loaded` value has it. `.loaded([])` (a just-created Source) matches Empty; while the read is `.loading` or `.failed` the filter waits or shows the error as today (`isAwaitingFilterData`, `statusCountsError`). `SourcesModel.statuses(for source: CatalogSource) -> Set<EvidenceStatus>` *(new)* computes it; `visibleSources` filters with `statuses(for:).contains(status)`. | User, 2026-10-26: a Source is in progress when at least one graph is; the same reading for every status. | All graphs must match — user chose any. Hiding non-matching sub-rows — brief GA-8 shows all. |
| D-8 | The `.sourceDetail` place adds `.evidenceGraphProgress(project: project)` as the last of its query keys. `SourceGraphsSection.progress(for graph:)` and `countsLoading` *(new)* forward to the session reads of D-13; its `rows` and `loadError` forward to D-17's read (`.loaded` → rows, `.failed` → error). `SourcePageModel.warmFromSession()` is not changed; page tests warm with `session.apply(location: .sourcePage(...))`. | Without the key, `refreshEvidenceGraphProgress` skips and a page opened first never loads counts (AC-3.2). Working rule "One graph list": progress is its own key, not folded into `GetSourceWorkspace`. | A per-Source progress RPC; a third hand key list in `warmFromSession`. |
| D-9 | Delete the interim helpers: `SourcesListNavigation.entryGraph(sourceID:in:)`, `graphLocation(for:graphs:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`; `SourcesModel.entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)`; the tests `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`, `graphLocationBlockedWithoutArtifact`, and the file-private test helper `graphsBySource(_:)` in `SourcesListNavigationTests.swift` (used only by those three). The per-graph blocked case stays covered by `graphLocationOpensTheGivenGraph`. `SourcesListNavigation.graphLocation(for:graph:)` and `graphBlockedReason(for:)` stay; the sub-row uses them. | Working rule "High tolerance for churn"; the map's "Until S12-14 adds per-graph Open". | Keeping them for the collapsed row — it opens nothing now. |
| D-10 | The expander is a `Button` over the full 248 pt zone with `SourcesSplitZoneButtonStyle(emphasizeOnHover: true)` (made internal, still in `SourcesSplitRow.swift`, so `SourcesGraphSubRow.swift` reuses it): `PVIcon(.network, size: 16)`, the label, `Spacer`, `PVIcon(PVExpandedState.chevron(isExpanded:))`, `.pvExpandedState(isExpanded)`. Padding as today's zone (`space6` horizontal, `space5 + spacePx` vertical). Identifier `sources.row.<sourceId>.graphs`. | Brief inventory: not `PVDisclosureButton` (a `PVButton` trigger, not a full-zone hit area). Chevron and spoken state from the kit. | `PVDisclosureButton`; a local chevron rule. |
| D-11 | `SourcesGraphSubRow` *(new)*, `macos/App/Features/Sources/SourcesGraphSubRow.swift`, takes `graph: CatalogEvidenceGraph`, `progress: EvidenceGraphProgress?`, `countsLoading: Bool`, `onOpen: (() -> Void)?`, `blockedReason: String?`. Columns through `SourcesSplitLayout.columns`. Left: `HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3)` of the name (`PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium)`, one line) and the ref (`PVFont.mono(size: PVTypeScale.micro)`, `textMuted`, fixed size), `accessibilityElement(children: .combine)`, leading padding `SourcesSplitLayout.subRowLeadingInset`, trailing `space6`, vertical `space5 + spacePx`. Right: today's Open zone body moved out of `SourcesSplitRow.graphZone` with `SourcesGraphCountLine` and `graphZoneAccessibility`. Background `PVColor.surfaceSunken`; bottom `PVDivider()`. Identifier `sources.row.<graph.sourceID>.graph.<graph.ref>`. | One view per graph; the Open zone has one home once the row stops drawing it. `graph.sourceID` carries the Source id. | A shared zone view used by both — the row no longer draws an Open zone. |
| D-12 | The row reads D-17's `session.evidenceGraphsRead(sourceID:)` through `SourcesModel.graphZone(for source: CatalogSource) -> EvidenceGraphsRead` *(new)*, which returns it unchanged except that `.loaded([])` becomes `.loading` (a just-created Source before the list refresh). The row draws `.loaded(graphs)` as the expander. It draws `.loading` as `PVSkeletonBar(width: 96, height: 10)` in place of the label, disabled, accessibility label `L10n.DesignSystem.listLoading`; `failed` as `PVIcon(.warning)` in `PVColor.danger` with `.help(L10n.Errors.message(for:))`, disabled, the same text as accessibility label. | Brief GA-6; one model member feeds the row (review X-8), over one session rule (review M-10). | Reading handles in the view; a callout per row — 248 pt is too narrow. |
| D-13 | `WorkspaceSession+EvidenceGraphs.swift` gains the progress reads beside the graph list's: `evidenceGraphProgressKey`, `evidenceGraphProgressHandle`, `evidenceGraphProgress(graphID:) -> EvidenceGraphProgress?`, `evidenceGraphCountsLoading: Bool` (progress handle or graph-list handle awaiting first value) *(new)*. `refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress` use the key; `SourcesModel.graphProgress(for:)` and `graphCountsLoading` forward to it, and `SourcesModel.evidenceGraphProgressKey(for:)` and the private `progressHandle` are deleted (tests move to `session.evidenceGraphProgressKey`). | One read path for the progress map, the shape the graph list already has (review M-1). | `SourceGraphsSection` copying `SourcesModel`'s reads. |
| D-14 | `QueryHandle.isAwaitingFirstValue: Bool` *(new)* = `value == nil && status != .ready && status != .error`. `graphCountsLoading`, the types branch of `isAwaitingFilterData` (today `value == nil && status != .error`, which is the same answer for every reachable state, since `.ready` always carries a value), D-12 and D-13 use it. | One meaning of "loading" (review M-4). | A fourth hand-written check. |
| D-15 | `SourcesListContent` loses its `progressHandle` parameter and `listBody` its inline no-handle row; every row is a `SourcesListRow(source:model:)`, which reads progress through the session (observation tracks the `@Observable` handle). | One row call site; the next row parameter is added once (review M-3, X-9). | Two branches each growing the sub-row loop. |
| D-16 | `SourcesSplitLayout` gains `static let thumbnailSize: CGFloat = 44` and `static var subRowLeadingInset: CGFloat { PVSpacing.gutterPage + thumbnailSize + PVSpacing.space5 }` *(new)*; the row's `CachedThumbnail` passes `size: SourcesSplitLayout.thumbnailSize`. | The sub-row name aligns with the Source title from one constant (review X-10). | A copied literal 44. |
| D-17 | `WorkspaceSession+EvidenceGraphs.swift` gains `enum EvidenceGraphsRead { case loading, failed(Error), loaded([CatalogEvidenceGraph]) }` and `evidenceGraphsRead(sourceID:) -> EvidenceGraphsRead` *(new)*, beside `evidenceGraphs(sourceID:)`: `failed` when `evidenceGraphsHandle` is `.error` with no value; `loading` while it `isAwaitingFirstValue` (D-14) or has no handle; else `loaded(evidenceGraphs(sourceID:) ?? [])`. The Sources list (D-7, D-12) and the Source page (D-8) read it. | One rule for "one Source's graphs: loading, failed or rows" in the session that owns the list (review M-10). | Each feature deriving it from the handle. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcesGraphCountLine` *(new)* | view | `SourcesSplitRow.graphCountLine` (private) | It is that code, lifted for a second caller (D-3). |
| `L10n.Sources.graphCountLineSpoken` *(new func)* | L10n func | `graphZoneAccessibility` | Its three branches, extracted so both read one (D-3). |
| `SourcesGraphSubRow` *(new)* | view | `SourcePageGraphRow`; `SourcesSplitRow.graphZone` | The page card is a `PVCard` with actions; the list needs the split columns. The zone body moves here (D-11). |
| `SourcesListRow` *(new, replaces `SourcesSplitRowProgress`)* | view | `SourcesSplitRowProgress` | Renamed: it now draws the row and its sub-rows (D-2, D-15). |
| `WorkspaceSession` `expandedSourceRows`, `isSourceRowExpanded`, `toggleSourceRowExpanded` *(new)* | session state | `graphCameras`, `graphCamera(for:)` | Same lifetime and shape (D-4). |
| `WorkspaceSession` `evidenceGraphProgressKey`, `evidenceGraphProgressHandle`, `evidenceGraphProgress(graphID:)`, `evidenceGraphCountsLoading` *(new)* | session reads | `evidenceGraphsKey`, `evidenceGraphsHandle`, `evidenceGraphs(sourceID:)` | Moved from `SourcesModel` beside the graph list's reads (D-13). |
| `QueryHandle.isAwaitingFirstValue` *(new)* | property | `isSettled` | Different question: no value yet (D-14). |
| `SourcesModel.statuses(for:)`, `graphZone(for:)` *(new)* | func | `graphProgress(for:)`, `visibleSources` | Model API the view calls (D-7, D-12). |
| `EvidenceGraphsRead`, `WorkspaceSession.evidenceGraphsRead(sourceID:)` *(new)* | enum / session read | `SourceGraphsSection.rows` / `loadError` | Lifted to the session so the list and the page share it (D-17). |
| `SourceGraphsSection.progress(for:)`, `countsLoading` *(new)* | forwarders | `SourceGraphsSection.rows` | Forward to D-13. |
| `SourcesSplitLayout.thumbnailSize`, `subRowLeadingInset` *(new)* | constants | `SourcesSplitLayout.graphZoneWidth` | Same enum (D-16). |
| `L10n.Sources.statusPromoted` *(new, replaces `statusFullyPromoted`)* | L10n key | `statusFullyPromoted` | Renamed with its key (D-6). |

Searched: `DisclosureGroup`, `isExpanded`, `PVDisclosureButton` (kit disclosure, rejected D-10); `graphCountLine`, `graphZoneAccessibility`, `evidenceGraphProgress`, `progressHandle`, `graphsBySource`, `entryGraph`, `isSettled`, `status != .error` (all callers listed in §4).

**Builds on.** `evidenceGraphsBySource` (S12-13), `SourcesListNavigation.graphLocation(for:graph:)` / `graphBlockedReason(for:)` (S12-11), `PVExpandedState` / `pvExpandedState`, the `PlaceRegistry` query keys ([`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md)), `WorkspaceSession.apply`'s per-graph progress patch.
**Deletes.** The interim helpers and their two tests (D-9); `SourcesModel.evidenceGraphProgressKey(for:)` and `progressHandle` (D-13); `SourcesSplitRowProgress` and the inline row (D-15); `SourcesSplitRow`'s `progress`, `countsLoading`, `onOpenGraph`, `graphBlockedReason` parameters and its private `graphCountLine`; `L10n.Sources.statusFullyPromoted` and its catalog entry.
**Next adds.** S12-15's graph hit opens the location a sub-row opens (`graphLocation(for:graph:)`). A later status mark on a sub-row reads `session.evidenceGraphProgress(graphID:)` and `EvidenceStatus(_:)`; a reader outside Sources moves `EvidenceStatus` out of `SourcesModel` first.

### Schema

None — no catalog change.

### Go core

None — the counts are today's `graphprogress.List` / `Get` (pending Open question 2).

### FFI

None — (pending Open question 2).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Split row | snowflake, extended | `SourcesSplitRow.swift` `SourcesSplitRow` | New parameters: `graphZone: EvidenceGraphsRead`, `isExpanded: Bool`, `onToggleGraphs: () -> Void`. `graphZone` draws the expander (D-1, D-10, D-12) or the unchanged "Needs an artifact" branch. |
| Expander | snowflake (new) | `SourcesSplitRow.swift` | D-10, D-12. |
| Graph sub-row | snowflake (new) | `SourcesGraphSubRow.swift` (new) | D-11, D-16. |
| Count line | snowflake (lifted) | `SourcesGraphCountLine.swift` (new) | D-3. |
| List row | snowflake | `SourcesListView.swift` `SourcesListRow` | D-2, D-15: the row with `graphZone: model.graphZone(for:)`, `isExpanded: session.isSourceRowExpanded(source.id)`, then, when expanded and `.loaded(let graphs)`, `ForEach(graphs)` sub-rows with `model.graphProgress(for:)`, `model.graphCountsLoading`, `SourcesListNavigation.graphLocation(for:graph:)` and `graphBlockedReason(for:)`. |
| Status menu | `PVSelect` | `SourcesListView.swift` `statusMenu` | Labels from `EvidenceStatus.label`; no layout change. |
| Source page card | snowflake, extended | `SourcePageGraphsView.swift` `SourcePageGraphRow` | `SourcesGraphCountLine(progress: model.graphs.progress(for: graph), countsLoading: model.graphs.countsLoading)` on its own line after the ref, before the description, inside the combined accessibility element. |

States: loading, failure and not-yet-listed on the expander (D-12); per-sub-row loading hairline from `SourcesGraphCountLine` (reachable when a row was expanded before a reload); 12 graphs list in full (no inner scroll). Accessibility: expander label is the count text with the expanded state; sub-row left reads "{name}, {ref}"; sub-row Open reads `graphZoneAccessibility` with the blocked hint; the count line speaks `graphCountLineSpoken` everywhere.

**Frames.** Filled from the agreed S12-D5 board, frame by frame, before status Planned.

### Strings

L10n keys: `L10n.Sources.statusPromoted` *(new)* = `sources.list.status.promoted` "Promoted" (comment: "Status filter option: graphs with at least one promoted subject"). Removed: `L10n.Sources.statusFullyPromoted` / `sources.list.status.fullyPromoted`. `sources.list.status.inProgress`'s comment says "at least one subject and none promoted". New func, no key: `L10n.Sources.graphCountLineSpoken`. Reused: `L10n.Sources.graphsCount` (comment widened to "Evidence graph count on the Source page section and the Sources list expander"). Final copy from the board replaces this line before Planned.

## 7. Steps

**STEP-1 — One read of graph progress.**
Files: `QueryHandle.swift`, `WorkspaceSession+EvidenceGraphs.swift`, `WorkspaceSession.swift`, `SourcesModel.swift`, `SourceGraphsSection.swift`, `SourcesListView.swift` (key call sites only), `SourcesModelTests.swift`, any test calling `SourcesModel.evidenceGraphProgressKey`. Skill: `add-workspace-place`.
Do: D-13, D-14, D-17; `SourceGraphsSection.rows` / `loadError` forward to D-17. No behavior change.
Check: T-10 passes; existing `SourcesModelTests` and `SourcePageModelTests` pass; `rg -n 'SourcesModel\.evidenceGraphProgressKey|evidenceGraphProgress\(project: projectKey\)' macos/App` lists only `WorkspaceSession+EvidenceGraphs.swift`.

**STEP-2 — Read status per graph.**
Files: `SourcesModel.swift`, `SourcesListSettings.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`, `SourcesListSettingsTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-5, D-6, D-7. The test helper becomes `progress(_ id:, subjects:, filed:)` (primary = subjects). Existing tests: `filtersCombineAndClear` seeds `filed: 0` where it expects In progress (s1, s2, s5) and keeps `filed: 1` where it expects Promoted; `settingsRestoreOnNewModel` and `statusCountsFailureSurfacesAnErrorAndUnfilteredCount` use `.promoted`.
Check: T-1, T-2, T-3 pass with every other `SourcesModelTests` test.

**STEP-3 — One count line, seen and spoken.**
Files: `SourcesGraphCountLine.swift` *(new)*, `SourcesSplitRow.swift`, `L10n.swift`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-3. The row still draws today's Open zone through the new view.
Check: the app builds; `graphProgressCopyMatchesBoard` passes unchanged; T-11 passes.

**STEP-4 — The expander, held per Source; the interim entry graph goes.**
Files: `WorkspaceSession.swift`, `SourcesModel.swift`, `SourcesSplitRow.swift`, `SourcesListView.swift`, `SourcesListNavigation.swift`, `SourcesListNavigationTests.swift`, `SourcesModelTests.swift`. Skill: `add-ui-component`.
Do: D-1, D-4, D-9, D-10, D-12, D-15, D-16. Rows toggle; no sub-rows yet.
Check: T-4, T-5, T-9 pass; `rg -n 'SourcesListNavigation\.entryGraph|model\.entryGraph|graphListLoaded|graphLocation\(for: source\)|graphs: graphsBySource|graphsBySource\(|SourcesSplitRowProgress' macos/App macos/ProvenenciaTests` returns nothing (`SourcePageModelTests.entryGraphLocation` is not matched and stays).

**STEP-5 — Graph sub-rows.**
Files: `SourcesGraphSubRow.swift` *(new)*, `SourcesListView.swift`.
Do: D-2, D-11.
Check: T-6 passes; manual check steps 1–3.

**STEP-6 — Counts on the Source page.**
Files: `PlaceRegistry.swift`, `PlaceRegistryTests.swift`, `SourceGraphsSection.swift`, `SourcePageGraphsView.swift`, `SourcePageModelTests.swift`. Skill: `add-workspace-place`.
Do: D-8; the card line (§6 macOS). `PlaceRegistryTests` `resolvesSourceDetail` and the place table's `.sourceDetail` row gain `.evidenceGraphProgress(project: project)` at the end.
Check: T-7, T-8 and `PlaceRegistryTests` pass.

**STEP-7 — Docs.**
Files: `../completed.md`, `../deployment-plan.md` (map row; working rule "A graph is a place" drops its interim first-graph sentence), this plan (status, Landed note), `../design/S12-D5-sources-list-graph-accordion.md` → `../design/archive/`, `../design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-accordion.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesModelTests.swift` `evidenceStatusFromCounts` (rewritten) | nil → empty; subjects 0 → empty; subjects 5, primary 3, filed 0 → inProgress; subjects 5, primary 3, filed 1 → promoted; subjects 2, primary 0 → inProgress | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-5 |
| T-2 | Swift | `SourcesModelTests.swift` `statusMatchesAnyGraphOfTheSource` *(new)* | One Source, three graphs (empty, in progress, promoted) → `visibleSources` contains it under each of the three statuses; a second Source with one empty graph shows only under Empty | AC-2.5, D-7 |
| T-3 | Swift | `SourcesListSettingsTests.swift` `storedFiledStatusRestoresPromoted` *(new)* | JSON `{"status":"filed"}` → `.promoted`; `{"status":"promoted"}` → `.promoted`; `{"status":"bogus"}` → nil with the other fields kept; `EvidenceStatus.promoted.label` resolves to "Promoted" | AC-2.6, AC-2.7, D-6 |
| T-4 | Swift | `SourcesModelTests.swift` `expandedRowsOutliveTheModel` *(new)* | `session.toggleSourceRowExpanded("s1")`; a new `SourcesModel` on the same session → `session.isSourceRowExpanded("s1")`; toggle again → false | AC-1.9, D-4 |
| T-5 | Swift | `SourcesListNavigationTests.swift` | `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`, `graphLocationBlockedWithoutArtifact` and `graphsBySource(_:)` deleted; `graphLocationOpensTheGivenGraph` passes unchanged | D-9 |
| T-6 | Swift | `SourcesModelTests.swift` `subRowsReadEachGraphsProgress` *(new)* | Three graphs on `s1`, progress on `g-jones` only → `model.graphZone(for: s1)` is `.loaded` in name order; `model.graphProgress(for:)` returns Jones's counts, nil for the others; `session.apply` with effects naming `g-jones` and new store counts changes only `g-jones` | AC-1.2, AC-1.3, AC-1.10 |
| T-7 | Swift | `SourcePageModelTests.swift` `graphRowsReadTheirProgress` *(new)* | Page for a Source with two graphs, progress seeded for one; `session.apply(location: .sourcePage(...))` then `waitForQuery` on `session.evidenceGraphProgressHandle` → `model.graphs.progress(for:)` returns it for that graph, nil for the other; `countsLoading` false | AC-3.1, AC-3.2, D-8 |
| T-8 | Swift | `SourcePageModelTests.swift` `aGraphWriteMovesOnlyItsRow` *(new)* | T-7, then the store's count changes and `session.apply` with effects naming that graph → its progress updates; the other stays nil | AC-3.3 |
| T-9 | Swift | `SourcesModelTests.swift` `graphZoneFollowsTheGraphList` *(new)* | Graph list not loaded → `.loading`; list failed → `.failed`; loaded without the Source → `.loading` (while `session.evidenceGraphsRead` says `.loaded([])`); loaded with it → `.loaded` in name order | AC-1.1, AC-1.8, D-12, D-17 |
| T-10 | Swift | `WorkspaceSessionTests.swift` `progressReadsComeFromTheSession` *(new)* | Before warm: `evidenceGraphCountsLoading` true and `evidenceGraphProgress(graphID:)` nil; after both handles load: false and the seeded value; a handle `.ready` with a value or `.error` → `isAwaitingFirstValue` false | D-13, D-14 |
| T-11 | Swift | `SourcesModelTests.swift` `graphProgressCopyMatchesBoard` (extended) | `graphCountLineSpoken(progress: nil, countsLoading: true)` == "counts loading"; zeros → "0 subjects, not started"; 12 / 48 → "12 subjects, 48 observations"; `graphZoneAccessibility` results unchanged | D-3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests`, `SourcesListNavigationTests`, `SourcesListSettingsTests`, `PlaceRegistryTests`, `SourcePageModelTests`, `WorkspaceSessionTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; rg -n 'fullyPromoted|Fully promoted' macos returns nothing
- [ ] UI matches the brief inventory; no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; working rule "A graph is a place" drops the interim first-graph sentence; completed.md: one short paragraph of what shipped
- [ ] Brief S12-D5 moved to design/archive/ and design/README.md updated
- [ ] Domain doc: none — the status rule is the map's working rule "Status is per graph" and SourcesModel.EvidenceStatus
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source and three graphs: Default (no Subjects), Jones household (Subjects, not promoted), Smith household (promoted once, one Subject filed of several). The Sources list shows "3 evidence graphs" on the census and "1 evidence graph" on the others; a Source with no Artifact reads "Needs an artifact" and does not expand (AC-1.1, AC-1.6, AC-1.7). Relaunch and watch the list load: the graph zones show a skeleton label first (AC-1.8).
2. Expand the census: Default, Jones household, Smith household in that order, each with its count line. Open Jones household: its canvas opens on that graph (AC-1.4). On the census Source page, add a new page Artifact, add a graph "Unlinked" linked only to it, then delete that Artifact: back in the list, Unlinked's sub-row Open is disabled with the hint and its count line shows (AC-1.5).
3. Add an Observation; Back. The census is still expanded; only Jones's line moved.
4. Status filter: In progress shows the census (Jones, AC-2.2); Promoted shows the census (Smith, AC-2.3); Empty shows the census. The last option reads Promoted.
5. Open the census Source page: each graph card shows its count line under the ref.
6. Quit and relaunch with the census Source page as the restored place: the cards' count lines load (AC-3.2). Go to the Sources list: every row is collapsed.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Expanding many rows in a long list costs layout | Sub-rows are built only for expanded rows inside the `LazyVStack` element; only `SourcesListRow` reads the expanded set. |
| The project-wide progress read on every Source page open | It is one cached map shared with the list; patches are per graph. |
| A graph with only value Subjects can never become Promoted | Accepted (D-5): Promote files only primary kinds; it reads In progress, which is true. |

**Stop and report if:**

- `WorkspaceSession.refreshEvidenceGraphProgress` no longer patches by graph id, or `apply` no longer calls it for every graph in `effects.evidenceGraphs`.
- `ListScrollAnchorModifier` needs each sub-row id to restore scroll (D-2 cannot hold).
- The agreed board draws something the inventory does not hold (a status badge, a Source-level total, a kit accordion).
- Observation does not repaint `SourcesListRow` when the progress map changes without a `@Bindable` handle (D-15): report with the test or manual step that shows it.

## 12. Open questions

1. **Board S12-D5.** Not yet agreed. On agreement: record the Board line, fill §6 Frames and Strings from it, then run `review-plan` round 2.
2. **`primaryCount` has no reader after D-5** (review M-5). Recommended: drop it end to end in this PR (`graphprogress.Progress.PrimaryCount`, `engine.proto` field 5 reserved, the handler, `GoStore`, `FakeStore`, `EvidenceGraphProgress`, tests), as one added step, per "High tolerance for churn". Alternative: Swift side only, Go and proto logged.

## 13. PR

**Title:** S12-14 — Graph counts in the Sources list and on the Source page
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-accordion.md`.
