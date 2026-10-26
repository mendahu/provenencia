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
**Review:** — (not reviewed yet)

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
| Graph zone expander on every row with an Artifact (GA-1, GA-6) | STEP-4, AC-1.1, AC-1.2, AC-1.8 |
| Graph sub-rows with per-graph Open and count line (GA-2 – GA-4, GA-9, GA-10) | STEP-5, AC-1.2 – AC-1.6 |
| "Needs an artifact" row unchanged (GA-5) | STEP-4, AC-1.7 |
| Session-held expanded state (GA-7, D-4) | STEP-3, AC-1.9 |
| Status rule per graph, any-graph Source match, Promoted label (GA-8, D-5 – D-7) | STEP-1, AC-2.1 – AC-2.7 |
| One shared count line view (D-3) | STEP-2, AC-1.3, AC-3.1 |
| Source page cards show the count line; the place warms the progress key (GA-11, D-8) | STEP-6, AC-3.1 – AC-3.3 |
| The interim entry-graph helpers deleted (D-9) | STEP-3 |

| Out | Home |
| --- | --- |
| Graph search in the list; omnibar graph hits | S12-15 |
| Create, edit, delete or link graphs from the list | not planned (Source page owns them) |
| A Promote button, a status badge, or filed / primary counts on a row, sub-row or card | not planned (map scope, "Not this PR") |
| A Source-level total of Subjects or Observations | not planned (map scope) |
| Persisting expanded rows across relaunch or in history | not planned (user, 2026-10-26) |
| A kit accordion or tree component | not planned (brief non-goals) |
| Any change to `graphprogress` or the progress RPCs | not planned (D-5: the rule is a Swift reading of today's counts) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow` | Page zone + one graph zone. Takes one `progress`, `countsLoading`, `onOpenGraph`, `graphBlockedReason`. `graphZone` draws "Open graph" + `graphCountLine` for the Source, or the sunken "Needs an artifact" zone when `!source.hasArtifact`. |
| Count line | same file, private `graphCountLine` | Three states: hairline while `countsLoading && progress == nil`; "0 subjects · not started" when `subjectCount == 0`; else `L10n.Sources.graphCountLine`. Private, one caller. |
| Zone style | same file, private `SourcesSplitZoneButtonStyle` | Hover fill and muted → primary label lift. |
| List wiring | `macos/App/Features/Sources/SourcesListView.swift` `listBody`, `SourcesSplitRowProgress` | `ForEach(rows)` draws one `SourcesSplitRow` per Source, `.id(source.id)`; `SourcesSplitRowProgress` observes the progress handle and passes the entry graph's progress. |
| Status menu | same file, `statusMenu` | `SourcesModel.EvidenceStatus.allCases` labels. |
| Model | `macos/App/Features/Sources/SourcesModel.swift` `EvidenceStatus` | Cases `empty`, `inProgress = "in_progress"`, `filed`; `init(_ progress:)`: empty when `primaryCount == 0`, in progress when `filedCount < primaryCount`, else filed. `.filed` labels `L10n.Sources.statusFullyPromoted`. |
| Model | same file, `entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)` | The interim "first graph" rule, used by the list wiring. |
| Model | same file, `visibleSources` | The status filter reads the entry graph's progress only. |
| Model | same file, `graphProgress(for:)`, `graphCountsLoading`, `warmListQueries`, `statusCountsError`, `isAwaitingFilterData` | Progress map read by graph id; loading of the progress and graph-list handles. Stay. |
| Settings | `macos/App/Features/Sources/SourcesListSettings.swift` | `status: SourcesModel.EvidenceStatus?`, decoded by raw value (`"filed"` stored today). |
| Navigation | `macos/App/Features/Sources/SourcesListNavigation.swift` | `entryGraph(sourceID:in:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`, `graphLocation(for:graphs:)` (interim, list-only); `graphLocation(for:graph:)` and `graphBlockedReason(for:)` (per graph; used by `SourceGraphsSection`). |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` | `graphCameras` holds per-graph UI state "not catalog data; quit drops it" (`@ObservationIgnored`). `apply(_:_:)` calls `refreshEvidenceGraphProgress(graphId:)` for each graph a write names; it patches the progress map only when the key is already cached. |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` `evidenceGraphsBySource`, `evidenceGraphs(sourceID:)`, `evidenceGraphsHandle` | The graph list by Source, in name order. |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `.sourcesList`, `.sourceDetail` | The list place warms `.evidenceGraphProgress(project:)`; the Source page place does not. |
| Source page | `macos/App/Features/Sources/SourceGraphsSection.swift` `SourceGraphsSection` | `rows`, `loadError`, `location(for:)`, `blockedReason(for:)`; no progress read. |
| Source page | `macos/App/Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Card: name, ref, description; Open, Edit, Delete. |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Sources` | `graphsCount(_:)` ("%lld evidence graph(s)", plural, key `sources.page.graphsCount`), `openGraph`, `graphCountLine`, `graphSubjectCount`, `graphNotStarted`, `graphZoneAccessibility`, `needsArtifact`, `needsArtifactTooltip`, `graphNeedsArtifactTooltip`, `statusFullyPromoted` (`sources.list.status.fullyPromoted`, "Fully promoted"). |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` | `evidenceStatusFromCounts`, `filtersCombineAndClear`, `statusFilterWaitsForCounts`, `graphProgressCopyMatchesBoard`, `warmListQueriesLoadsGraphProgress`; helper `progress(_:primary:filed:)` sets `subjectCount = primary`. |
| Tests | `macos/ProvenenciaTests/SourcesListNavigationTests.swift` | `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList` (interim helpers); `graphLocationOpensTheGivenGraph` (stays). |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` | Graph section tests; `FakeStore.graphProgressByGraph` seeds progress. |

**Precedent.** S12-11 ([#549](https://github.com/mendahu/provenencia/pull/549)) put the Source page's graph rows on session reads (`SourceGraphsSection.rows` over `evidenceGraphs(sourceID:)`) instead of a page copy: read progress the same way. `graphCameras` is the precedent for UI state held by the session until quit. S12-13 ([#562](https://github.com/mendahu/provenencia/pull/562)) built `evidenceGraphsBySource`, which the sub-rows read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Every row with an Artifact has the expander; its label is `L10n.Sources.graphsCount(n)`. A Source with no Artifact keeps today's "Needs an artifact" zone with no expander. | User, 2026-10-26: every row expands. A Source with no Artifact has only an unopenable, empty Default graph; the existing zone already says what to do. | One-graph rows keep a direct Open — user chose every row. An expander on the no-Artifact row — one disabled sub-row says less than today's zone. |
| D-2 | Sub-rows are drawn by the list, under the row, inside the same `ForEach` element (a `VStack(spacing: 0)` holding the row and, when expanded, its sub-rows), with `.id(source.id)` on that `VStack`. | `ListScrollAnchorModifier` anchors on Source ids; a sub-row is not a list row of its own. | Sub-rows as separate `ForEach` items — breaks the scroll anchor's id list. |
| D-3 | `SourcesGraphCountLine` *(new)*, `macos/App/Features/Sources/SourcesGraphCountLine.swift`: `struct SourcesGraphCountLine: View { var progress: EvidenceGraphProgress?; var countsLoading: Bool }`, body moved verbatim from `SourcesSplitRow.graphCountLine`. The sub-row and `SourcePageGraphRow` draw it. | One count rule, two views (map scope). | Copying the three states into the page card. |
| D-4 | Expanded state is `WorkspaceSession.expandedSourceRows: Set<String>` *(new)*, observed (no `@ObservationIgnored`), beside `graphCameras`. `SourcesModel.isExpanded(_ sourceID:) -> Bool` and `toggleExpanded(_ sourceID:)` *(new)* read and write it. Deleted Source ids may stay in the set. | User, 2026-10-26: until relaunch. The list's `SourcesModel` is rebuilt with the view; the session lives for the project. It must be observed so a toggle repaints. | `SourcesListSettings` (persists, user declined); `WorkspaceLocation` (history, user declined); `@State` on the view (lost on leave). |
| D-5 | Graph status is read from today's counts: `.empty` when `subjectCount == 0`; `.promoted` when `filedCount > 0`; else `.inProgress`. A graph missing from the progress map is empty. | User, 2026-10-26: In progress = at least one Subject and none promoted; one filed Subject makes it done. `filedCount` is "primary Subjects with an accepted Identity Claim", the only promoted signal the map carries. `subjectCount` is what the count line shows, so a graph reading "0 subjects" is Empty. | Keeping `primaryCount == 0` as Empty — a graph with only value Subjects would read Empty while its line says "2 subjects". A new "has ever been promoted" engine field — `filedCount > 0` is that fact today; Unpromote of every Subject returns the graph to In progress, which matches "none promoted". |
| D-6 | The case `.filed` is renamed `.promoted` with raw value `"filed"` kept; `inProgress` keeps `"in_progress"`. | Stored settings decode by raw value (AC-2.7). The Swift name follows the new label. | A new raw value — drops every saved status filter. |
| D-7 | A Source matches a status when any graph in `session.evidenceGraphsBySource[source.id]` has it. A Source with no graph in the list matches Empty. `SourcesModel.statuses(for source: CatalogSource) -> Set<EvidenceStatus>` *(new)* computes it; `visibleSources` filters with `statuses(for:).contains(status)`. | User, 2026-10-26: a Source is in progress when at least one graph is; the same reading for every status. | All graphs must match — user chose any. Hiding non-matching sub-rows — user did not ask; brief GA-8 shows all. |
| D-8 | The `.sourceDetail` place adds `.evidenceGraphProgress(project: project)` to its query keys. `SourceGraphsSection.progress(for graph: CatalogEvidenceGraph) -> EvidenceGraphProgress?` and `countsLoading: Bool` *(new)* read the session handle the way `SourcesModel.graphProgress(for:)` and `graphCountsLoading` do. | Without the key, `refreshEvidenceGraphProgress` skips (it patches only a cached map) and a page opened first never loads counts (AC-3.2). Working rule "One graph list": per-graph progress is its own key, not folded into `GetSourceWorkspace`. | A per-Source progress RPC — the map already covers it, no engine change. |
| D-9 | Delete the interim helpers: `SourcesListNavigation.entryGraph(sourceID:in:)`, `graphLocation(for:graphs:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`; `SourcesModel.entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)`; the tests `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`. `SourcesListNavigation.graphLocation(for:graph:)` and `graphBlockedReason(for:)` stay; the sub-row uses them. | Working rule "High tolerance for churn"; the map's "Until S12-14 adds per-graph Open". | Keeping them for the collapsed row — the collapsed row opens nothing now. |
| D-10 | The expander is a `Button` over the full 248 pt zone with `SourcesSplitZoneButtonStyle(emphasizeOnHover: true)` (made internal, still in `SourcesSplitRow.swift`, so `SourcesGraphSubRow.swift` reuses it), trailing `PVIcon(PVExpandedState.chevron(isExpanded:))`, and `.pvExpandedState(isExpanded)`. | Brief inventory: not `PVDisclosureButton` (a `PVButton` trigger, not a full-zone hit area). The chevron and spoken state come from the kit so they match every other disclosure. | `PVDisclosureButton` — wrong hit area and look. A local chevron rule. |
| D-11 | `SourcesGraphSubRow` *(new)*, `macos/App/Features/Sources/SourcesGraphSubRow.swift`, takes `graph: CatalogEvidenceGraph`, `progress: EvidenceGraphProgress?`, `countsLoading: Bool`, `onOpen: (() -> Void)?`, `blockedReason: String?`. Left column: name and ref, `accessibilityElement(children: .combine)`; right column: today's Open zone body moved out of `SourcesSplitRow.graphZone` into this view, with `graphZoneAccessibility(progress:countsLoading:)`. Identifiers `sources.row.<sourceId>.graph.<graphRef>` (zone) and `sources.row.<sourceId>.graphs` (expander). | One view per graph; the Open zone has one home after the row stops drawing it. | A shared "zone" view used by both — the row no longer draws an Open zone. |
| D-12 | While the graph list has no value and no error, the expander shows `PVSkeletonBar(width: 96, height: 10)` in place of the label and is disabled. On a graph-list error it shows `PVIcon(.warning)` in `PVColor.danger` with `.help(L10n.Errors.message(for:))` and is disabled. | Brief GA-6. `statusCountsError` already surfaces the error list-wide only when a status filter is set. | A callout per row — 248 pt is too narrow. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcesGraphCountLine` *(new)* | view | `SourcesSplitRow.graphCountLine` (private) | It is that code, lifted for a second caller (D-3). |
| `SourcesGraphSubRow` *(new)* | view | `SourcePageGraphRow`; `SourcesSplitRow.graphZone` | The page card is a `PVCard` with actions; the list needs the split columns. The zone body moves here (D-11). |
| `WorkspaceSession.expandedSourceRows` *(new)* | session state | `graphCameras` | Same lifetime; must be observed (D-4). |
| `SourcesModel.isExpanded(_:)`, `toggleExpanded(_:)`, `statuses(for:)` *(new)* | func | `graphProgress(for:)`, `visibleSources` | Model API the view calls; `statuses(for:)` replaces the inline entry-graph read. |
| `SourceGraphsSection.progress(for:)`, `countsLoading` *(new)* | func | `SourcesModel.graphProgress(for:)`, `graphCountsLoading` | The page model reads the same key; a page cannot hold a `SourcesModel`. |
| `L10n.Sources.statusPromoted` *(new, replaces `statusFullyPromoted`)* | L10n key | `statusFullyPromoted` | Renamed with its key (D-6). |

Searched: `DisclosureGroup`, `isExpanded`, `PVDisclosureButton` (kit disclosure, rejected D-10); `graphCountLine`, `evidenceGraphProgress`, `graphsBySource`, `entryGraph` (all callers listed in §4).

**Builds on.** `evidenceGraphsBySource` (S12-13), `SourcesListNavigation.graphLocation(for:graph:)` / `graphBlockedReason(for:)` (S12-11), `PVExpandedState` / `pvExpandedState`, the `PlaceRegistry` query keys ([`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md)), `WorkspaceSession.apply`'s per-graph progress patch.
**Deletes.** The interim helpers and their two tests (D-9); `SourcesSplitRow`'s `progress`, `countsLoading`, `onOpenGraph`, `graphBlockedReason` parameters and its private `graphCountLine`; `L10n.Sources.statusFullyPromoted` and its catalog entry.
**Next adds.** S12-15's graph hit opens the same location a sub-row opens (`graphLocation(for:graph:)`). A later per-graph status mark on a sub-row reads `EvidenceStatus(progress)` with no new rule.

### Schema

None — no catalog change.

### Go core

None — the counts are today's `graphprogress.List` / `Get`.

### FFI

None —

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Split row | snowflake, extended | `SourcesSplitRow.swift` `SourcesSplitRow` | New parameters: `graphCount: Int?` (nil while the graph list has no value), `graphListError: Error?`, `isExpanded: Bool`, `onToggleGraphs: () -> Void`. `graphZone` draws the expander (D-1, D-10, D-12) or the unchanged "Needs an artifact" branch. |
| Expander | snowflake (new) | `SourcesSplitRow.swift` | Network icon, `Text(verbatim: L10n.Sources.graphsCount(n))`, `Spacer`, chevron; label = the count text; `.pvExpandedState`. Identifier `sources.row.<id>.graphs`. |
| Graph sub-row | snowflake (new) | `SourcesGraphSubRow.swift` (new) | `SourcesSplitLayout.columns`; left padding `PVSpacing.gutterPage` + 44 (the `CachedThumbnail` default `size`) + `PVSpacing.space5` so the name aligns with the Source title; background `PVColor.surfaceSunken`; bottom `PVDivider()`. |
| Count line | snowflake (lifted) | `SourcesGraphCountLine.swift` (new) | D-3. |
| List wiring | — | `SourcesListView.swift` | `SourcesSplitRowProgress` passes the row's new parameters and, when `model.isExpanded(source.id)`, draws `ForEach(session.evidenceGraphs(sourceID:) ?? [])` sub-rows with `handle.value?[graph.id]`; the no-handle branch passes `countsLoading: true` to its sub-rows (D-2). |
| Status menu | `PVSelect` | `SourcesListView.swift` `statusMenu` | Labels from `EvidenceStatus.label`, no layout change. |
| Source page card | snowflake, extended | `SourcePageGraphsView.swift` `SourcePageGraphRow` | `SourcesGraphCountLine(progress: model.graphs.progress(for: graph), countsLoading: model.graphs.countsLoading)` on its own line after the ref, before the description, inside the combined accessibility element. |

States: loading and graph-list failure on the expander (D-12); per-sub-row loading hairline from `SourcesGraphCountLine`; 12 graphs list in full (no inner scroll). Accessibility: expander label is the count text with the expanded state; sub-row left reads "{name}, {ref}"; sub-row Open reads `graphZoneAccessibility` with the blocked hint.

**Frames.** Filled from the agreed S12-D5 board, frame by frame, before status Planned.

### Strings

L10n keys: `L10n.Sources.statusPromoted` *(new)* = `sources.list.status.promoted` "Promoted" (comment: "Status filter option: graphs with at least one promoted subject"). Removed: `L10n.Sources.statusFullyPromoted` / `sources.list.status.fullyPromoted`. `sources.list.status.inProgress`'s comment says "at least one subject and none promoted". Reused: `L10n.Sources.graphsCount` (comment widened to "Evidence graph count on the Source page section and the Sources list expander"). Final copy from the board replaces this line before Planned.

## 7. Steps

**STEP-1 — Read status per graph.**
Files: `SourcesModel.swift`, `SourcesListSettings.swift` (no change expected; verify decoding), `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-5, D-6, D-7. Rename the label key. Update `progress(_:primary:filed:)` in the tests to take `subjects:` separately where a test needs `subjectCount != primaryCount`.
Check: T-1, T-2, T-3 pass; `xcodebuild test` for `SourcesModelTests`.

**STEP-2 — Lift the count line.**
Files: `SourcesGraphCountLine.swift` *(new)*, `SourcesSplitRow.swift`. Skill: `add-ui-component`.
Do: D-3. The row still draws today's Open zone through the new view.
Check: the app builds; `graphProgressCopyMatchesBoard` passes.

**STEP-3 — Hold expanded rows; drop the interim entry graph.**
Files: `WorkspaceSession.swift`, `SourcesModel.swift`, `SourcesListNavigation.swift`, `SourcesListNavigationTests.swift`, `SourcesModelTests.swift`.
Do: D-4, D-9 (the list still compiles against the row's old parameters until STEP-4: keep the row's call sites passing `nil` / `true` for one step, or do STEP-3 and STEP-4 in one commit).
Check: T-4, T-5 pass; `rg -n 'entryGraph|graphLocation\(for: source\)|graphListLoaded' macos` returns nothing.

**STEP-4 — The expander.**
Files: `SourcesSplitRow.swift`, `SourcesListView.swift`. Skill: `add-ui-component`.
Do: D-1, D-10, D-12; the row's new parameters (§6 macOS).
Check: the app builds; in the app, rows read "N evidence graphs" and toggle the chevron.

**STEP-5 — Graph sub-rows.**
Files: `SourcesGraphSubRow.swift` *(new)*, `SourcesListView.swift`.
Do: D-2, D-11.
Check: T-6 passes; manual check steps 1–4.

**STEP-6 — Counts on the Source page.**
Files: `PlaceRegistry.swift`, `SourceGraphsSection.swift`, `SourcePageGraphsView.swift`, `SourcePageModelTests.swift`. Skill: `add-workspace-place`.
Do: D-8; the card line (§6 macOS).
Check: T-7, T-8 pass.

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
| T-3 | Swift | `SourcesListSettingsTests.swift` `storedFiledStatusRestoresPromoted` *(new)* | Settings JSON `{"status":"filed"}` → `status == .promoted`; `EvidenceStatus.promoted.label` resolves to "Promoted" | AC-2.6, AC-2.7, D-6 |
| T-4 | Swift | `SourcesModelTests.swift` `expandedRowsOutliveTheModel` *(new)* | Toggle `s1` on one `SourcesModel`; a new `SourcesModel` on the same session → `isExpanded("s1")`; toggle again → false | AC-1.9, D-4 |
| T-5 | Swift | `SourcesListNavigationTests.swift` | `graphLocationOpensTheFirstGraph` and `graphBlockedReasonNeedsTheGraphList` deleted; `graphLocationOpensTheGivenGraph` passes unchanged | D-9 |
| T-6 | Swift | `SourcesModelTests.swift` `subRowsReadEachGraphsProgress` *(new)* | Three graphs on `s1` with progress on `g-jones` only → `session.evidenceGraphs(sourceID: "s1")` in name order and `model.graphProgress(for:)` per graph returns Jones's counts, nil for the others; applying effects naming `g-jones` with new counts changes only `g-jones` | AC-1.2, AC-1.3, AC-1.10 |
| T-7 | Swift | `SourcePageModelTests.swift` `graphRowsReadTheirProgress` *(new)* | Page for a Source with two graphs, progress seeded for one → `model.graphs.progress(for:)` returns it for that graph and nil for the other after the place's keys warm; `countsLoading` false | AC-3.1, AC-3.2, D-8 |
| T-8 | Swift | `SourcePageModelTests.swift` `aGraphWriteMovesOnlyItsRow` *(new)* | T-7, then `session.apply` with effects naming one graph and the store's count changed → that graph's progress updates; the other stays nil | AC-3.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests` (status helper, filter tests), `SourcesListNavigationTests`, `SourcePageModelTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (statusFullyPromoted gone)
- [ ] UI matches the brief inventory; no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; working rule "A graph is a place" drops the interim first-graph sentence; completed.md: one short paragraph of what shipped
- [ ] Brief S12-D5 moved to design/archive/ and design/README.md updated
- [ ] Domain doc: none — the status rule lives in SourcesModel.EvidenceStatus only (rg "Fully promoted" docs returns nothing)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source and three graphs: Default (no Subjects), Jones household (Subjects, not promoted), Smith household (promoted once). The Sources list shows "3 evidence graphs" on the census and "1 evidence graph" on the others; a Source with no Artifact reads "Needs an artifact" and does not expand (AC-1.1, AC-1.6, AC-1.7). Relaunch and watch the list load: the graph zones show a skeleton label first (AC-1.8).
2. Expand the census: Default, Jones household, Smith household in that order, each with its count line. Open Jones household: its canvas opens on that graph (AC-1.4). Add a fourth graph "Unlinked" with no Artifact (Source page, then unlink): its sub-row's Open is disabled with the hint (AC-1.5).
3. Add an Observation; Back. The census is still expanded; only Jones's line moved.
4. Status filter: In progress shows the census (Jones, AC-2.2); Promoted shows the census (Smith, one Subject filed of several, AC-2.3); Empty shows the census. The last option reads Promoted.
5. Open the census Source page: each graph card shows its count line under the ref.
6. Quit and relaunch: every row is collapsed.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Expanding many rows in a long list costs layout | Sub-rows are only built for expanded rows inside the `LazyVStack` element. |
| The project-wide progress read on every Source page open | It is one cached map shared with the list; patches are per graph. |
| A graph with only value Subjects can never become Promoted | Accepted (D-5): Promote files only primary kinds; it reads In progress, which is true. |

**Stop and report if:**

- `WorkspaceSession.refreshEvidenceGraphProgress` no longer patches by graph id, or `apply` no longer calls it for every graph in `effects.evidenceGraphs`.
- `ListScrollAnchorModifier` needs each sub-row id to restore scroll (D-2 cannot hold).
- The agreed board draws something the inventory does not hold (a status badge, a Source-level total, a kit accordion).
- `SourcesListSettings` decodes status by something other than raw value (D-6).

## 12. Open questions

1. **Board S12-D5.** Not yet agreed. On agreement: record the Board line, fill §6 Frames and Strings from it, then run `review-plan`.

## 13. PR

**Title:** S12-14 — Graph counts in the Sources list and on the Source page
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-accordion.md`.
