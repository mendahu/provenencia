# S12-14 — Graph counts in the Sources list and on the Source page

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D5](../design/archive/S12-D5-sources-list-graph-tokens.md), board agreed 2026-10-26
**Board:** [Claude Design](https://claude.ai/design/p/118d2eb7-3b46-4912-a688-09ec7b0070fa?via=share), agreed 2026-10-26 (`S2-04 Sources list`, frames 1–13). The planner rendered the exported board and checked every frame against §6; §6 Frames describes them in words for an implementer who cannot open the link
**Depends on:** S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528)), S12-11 (landed [#549](https://github.com/mendahu/provenencia/pull/549)), S12-13 (landed [#562](https://github.com/mendahu/provenencia/pull/562): `evidenceGraphsBySource`), S12-D5 (board agreed 2026-10-26)
**Blocks:** —
**Branch:** `s12-14-sources-list-graph-tokens`, branched from `main` after the plan branch merges
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-13 did; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (place query keys, session cache reads), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md) (the `primary_count` removal)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [where graphs show](../source-structure.md#where-graphs-show); working rules: "Sources list", "Status is per graph", "Source page", "One graph list", "Name and description are editable; graphs list alphabetically by name", "A graph is a place", "High tolerance for churn"; the S12-11 plan D-2 (session reads of the graph list), the S12-13 [Landed note](S12-13-artifact-graph-checkboxes.md) (`evidenceGraphsBySource`, `EvidenceGraphIndex`)
**Planned:** 2026-10-26 against `main` @ `e3b85f9`
**Review:** [`S12-14-sources-list-graph-tokens.review.md`](S12-14-sources-list-graph-tokens.review.md) — accordion plan: Fits with changes / Ready with changes, 2 rounds; token plan: see the log

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

A researcher sees every Evidence graph of every Source in the Sources list: under each Source's title, a line of graph tokens, one per graph in name order, each with that graph's compact counts and opening it. The Source page's graph cards show the full count line. The Status filter reads each graph, and In progress now means "has Subjects, none promoted yet": a graph that has been through a Promote with at least one Subject filed counts as Promoted.

**Why now.** S12-11 to S12-13 let a Source hold several graphs, but the list still opens and counts only the alphabetically first one (the interim rule in working rule "A graph is a place"), and the Source page shows no progress. The status change is the product owner's call at planning (2026-10-26): today's In progress (`filedCount < primaryCount`) keeps a graph "in progress" forever when one Subject is left unfiled on purpose. That leaves `primary_count` with no reader, so it goes (user, 2026-10-26).

**What stays.** The list's header, count line, toolbar, empty and page-level states, Add source and the Copy… menu. Search matches Source title and ref only. The counts are today's `CatalogQueryKey.evidenceGraphProgress` map; no new query or RPC.

## 2. User stories

**US-1.** As a researcher, in the Sources list, I see each Source's Evidence graphs with their own counts and can open any of them, so that I can reach the Jones household without opening the Source page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A census Source with graphs "Smith household", "Default", "Jones household", all linked to an Artifact; the list loaded | Under the census's title, three tokens in the order Default, Jones household, Smith household |
| AC-1.2 | Jones household has 12 Subjects and 40 Observations; Default has none | Jones's token reads "Jones household 12 · 40" with help "12 subjects · 40 observations"; Default's reads "Default not started" with help "0 subjects · not started" |
| AC-1.3 | Click Jones household's token | The canvas opens on Jones household's graph id (`WorkspaceLocation.evidenceGraph(sourceId:graphId:ref:title:)`) |
| AC-1.4 | A fourth graph "Unlinked" links no Artifact | Its token is dashed and faint, reads "Unlinked needs an artifact", does nothing on click, and has the help "Add an artifact to open this evidence graph" |
| AC-1.5 | A Source with no Artifact | One dashed "Needs an artifact" token with the help "Add an artifact on the source page to build its evidence graph" |
| AC-1.6 | The graph list has not loaded | One skeleton token stands in for the token line. When the list fails: one inert token with the danger glyph whose help is `L10n.Errors.message(for:)` of that failure |
| AC-1.7 | The graph list has loaded and the counts have not | Each token shows a short rule where its count goes and still opens |
| AC-1.8 | Click the census title or cover | The Source page opens; hovering it underlines the title in the link color, with no fill |
| AC-1.9 | Add an Observation on Jones's canvas, Back to the list | Only Jones's token count changes ("12 · 41"), with no list reload |
| AC-1.10 | Twelve graphs on one Source, one with a 60-character name, at the list's minimum width | The tokens wrap onto as many lines as they need; the long name truncates inside its token |

**US-2.** As a researcher, when I filter the Sources list by status, a Source shows when any of its graphs has that status, and the statuses mean Empty / In progress / Promoted as defined below.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A graph with 0 Subjects | Its status is Empty |
| AC-2.2 | A graph with 5 Subjects, 0 filed | In progress |
| AC-2.3 | A graph with 5 Subjects, 1 filed | Promoted (today it reads In progress) |
| AC-2.4 | A graph with 2 Subjects that are not of a primary kind (filed is always 0) | In progress |
| AC-2.5 | Census: Default Empty, Jones In progress, Smith Promoted | The census shows under Empty, under In progress and under Promoted, with all three tokens each time |
| AC-2.6 | The Status menu | Options read All statuses, Empty, In progress, Promoted |
| AC-2.7 | A settings file saved with `"status":"filed"` before this PR | The list restores with Promoted chosen |

**US-3.** As a researcher, on a Source page, each Evidence graphs card shows that graph's counts.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The census page; Jones has 12 Subjects and 40 Observations | Jones's card shows "12 subjects · 40 observations" under its ref, above its description; Default's shows "0 subjects · not started" |
| AC-3.2 | The page restored straight from a relaunch (the list never shown) | The count lines load (the place warms the progress key) |
| AC-3.3 | Add an Observation on Jones's canvas, Back to the page | Only Jones's card count moves, with no page reload |

**US-4.** As a contributor, the graph progress row carries only the counts something reads.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `graphprogress.Progress`, `EvidenceGraphProgress` (proto and Swift) | No primary count field; proto field 5 and its name are reserved; filed still counts only primary-kind Subjects with an accepted claim |

## 3. Scope

| In | Covered by |
| --- | --- |
| One session read of a Source's graphs and of graph progress; one "awaiting first value" check (D-13, D-14, D-17) | STEP-1, T-9, T-10 |
| Status rule per graph, any-graph Source match, Promoted label and stored value (GT-7, D-5 – D-7) | STEP-2, AC-2.1 – AC-2.7 |
| `primary_count` removed from Go, proto, handler and Swift (D-18) | STEP-3, AC-4.1 |
| `PVFlowLayout` promoted from the Citation composer (D-19) | STEP-4 |
| Full count line view and one count summary formatter; the Source page cards draw it; the place warms the progress key (GT-12, D-3, D-8) | STEP-5, AC-3.1 – AC-3.3 |
| Graph token with its states (GT-4 – GT-6, GT-10, D-11, D-12) | STEP-6, AC-1.2, AC-1.4 – AC-1.7 |
| The two-line row, caption, Source link, token line; the split row and interim helpers deleted (GT-1 – GT-3, GT-8, GT-9, GT-11, D-1, D-2, D-9, D-10, D-15, D-16) | STEP-7, AC-1.1, AC-1.3, AC-1.8 – AC-1.10 |

| Out | Home |
| --- | --- |
| Graph search in the list; omnibar graph hits | S12-15 |
| Create, edit, delete or link graphs from the list | not planned (Source page owns them) |
| A Promote button, a status badge, or filed counts on a token or card | not planned (map scope) |
| A Source-level total of Subjects or Observations | not planned (map scope) |
| Shift-F10 opening the row menu (frame 1, GT keyboard note) | not planned here: `pvRightClickMenu` has no keyboard trigger; logged in the review log for a kit PR |
| Moving the status rule into `graphprogress` | when a reader outside `SourcesModel` needs it (review M-5) |
| `SourcesModel.warmListQueries` warming the keys the place already warms | not planned here (review log, "Log it") |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitLayout`, `SourcesSplitRow`, private `SourcesSplitZoneButtonStyle` | Two zones on a `minmax(0,1fr) | 248` track: the page zone (one button: `CachedThumbnail`, title, type, `PVListDetailCell` with `PVLastUpdated`, ref; `pvRightClickMenu` with Copy…; label `L10n.Sources.openSourcePage`, value via `PVLastUpdated.label`) and the graph zone ("Open graph" + private `graphCountLine`, or the sunken "Needs an artifact" zone). Takes one `progress`, `countsLoading`, `onOpenGraph`, `graphBlockedReason`. |
| Count line | same file, private `graphCountLine` | Three states: hairline while `countsLoading && progress == nil`; "0 subjects · not started" when `subjectCount == 0`; else `L10n.Sources.graphCountLine`. |
| List wiring | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListView`, `SourcesListContent` (`progressHandle` parameter), `listBody`, `captionBand`, `SourcesSplitRowProgress` | `listBody` builds `SourcesSplitRow` twice (`SourcesSplitRowProgress` with the handle, inline with `countsLoading: true` without). `captionBand` draws "Source" and "Evidence graph" on the split track. Each row is one `ForEach(rows)` element with `.id(source.id)`. |
| Status menu | same file, `statusMenu` | Options from `SourcesModel.EvidenceStatus.allCases`; identifiers `sources.status.<rawValue>`. |
| Model | `macos/App/Features/Sources/SourcesModel.swift` `EvidenceStatus` | Cases `empty`, `inProgress = "in_progress"`, `filed`; `init(_ progress:)`: empty when `primaryCount == 0`, in progress when `filedCount < primaryCount`, else filed. `.filed` labels `L10n.Sources.statusFullyPromoted`. |
| Model | same file, `evidenceGraphProgressKey(for:)`, private `progressHandle`, private `graphsHandle`, `graphProgress(for:)`, `graphCountsLoading`, `isAwaitingFilterData`, `statusCountsError`, `warmListQueries` | Progress key and handle looked up here; `graphCountsLoading` and `isAwaitingFilterData` hand-write "no value and not settled". |
| Model | same file, `entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)`, `visibleSources` | The interim "first graph" rule; the status filter reads the entry graph only. |
| Settings | `macos/App/Features/Sources/SourcesListSettings.swift` `SourcesListSettings.init(from:)` | Decodes each field on its own; `status` by raw value (`"filed"` stored today). |
| Navigation | `macos/App/Features/Sources/SourcesListNavigation.swift` | `entryGraph(sourceID:in:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`, `graphLocation(for:graphs:)` (interim, list-only); `pageLocation(for:)`, `graphLocation(for:graph:)`, `graphBlockedReason(for:)` (stay). |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` | `refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress` build `CatalogQueryKey.evidenceGraphProgress(project:)` inline; the patch runs only when the map is cached. |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` | "The one place" the graph list's key and handle are looked up: `evidenceGraphsKey`, `evidenceGraphsHandle`, `evidenceGraphs(sourceID:)`, `evidenceGraphsBySource`. |
| Query handle | `macos/App/Features/Workspace/Session/QueryHandle.swift` `QueryHandle` | `status`, `value`, `error`, `isSettled`. `applySuccess` is the only path to `.ready` and always sets a value. |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `.sourcesList`, `.sourceDetail` | The list place warms `.evidenceGraphProgress(project:)`; the Source page place does not. |
| Source page | `macos/App/Features/Sources/SourceGraphsSection.swift` `SourceGraphsSection` | `rows`, `loadError` (reads `evidenceGraphsHandle`), `location(for:)`, `blockedReason(for:)`; no progress read. |
| Source page | `macos/App/Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Card: name, ref, description (combined accessibility element); Open, Edit, Delete. |
| Source page | `macos/App/Features/Sources/SourcePageModel.swift` `warmFromSession()` | Tests and previews warm a hand list of keys; navigation warms through `PlaceRegistry`. |
| Flow layout | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` private `FlowLayout` | Wraps children by `sizeThatFits(.unspecified)`; two call sites in that file. |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Sources` | `columnSource` (`sources.list.columnSource`), `columnEvidenceGraph`, `openGraph`, `graphCountLine`, `graphSubjectCount`, `graphObservationCount`, `graphNotStarted`, `graphCountsLoading`, `graphZoneAccessibility`, `needsArtifact`, `needsArtifactTooltip`, `graphNeedsArtifactTooltip`, `statusFullyPromoted` (`sources.list.status.fullyPromoted`), `openSourcePage`. |
| Go | `core/database/graphprogress/graphprogress.go` `Progress`, `add` | `PrimaryCount` sums primary-kind Subjects; `FiledCount` sums those with an accepted claim. |
| Wire | `api/proto/engine.proto` `EvidenceGraphProgress` | `primary_count = 5`; `reserved 4; reserved "uncited_count";` is the precedent. Handler `api/ffi/handlers/evidence_graph_progress.go` maps it; `macos/App/Platform/GoStore.swift`, `FakeStore.swift` (`primaryCount:` in its progress builder), `GenealogyStore.swift` `EvidenceGraphProgress.primaryCount`. |
| Tests | `core/database/graphprogress/graphprogress_test.go` `TestGraphProgress`, `TestPrimaryAndFiledCounts`; `api/ffi/handlers/evidence_graph_progress_test.go` `TestListEvidenceGraphProgress` | Assert `PrimaryCount`. |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` | `evidenceStatusFromCounts`, `filtersCombineAndClear` (seeds `primary: 2, filed: 1`, expects In progress), `settingsRestoreOnNewModel` (`.filed`), `statusCountsFailureSurfacesAnErrorAndUnfilteredCount` (`.filed`), `statusFilterWaitsForCounts`, `graphProgressCopyMatchesBoard` (`graphZoneAccessibility`), `warmListQueriesLoadsGraphProgress`; helper `progress(_:primary:filed:)`. |
| Tests | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` | Two `EvidenceGraphProgress` literals with `primaryCount:` (about lines 985, 996). |
| Tests | `macos/ProvenenciaTests/SourcesListNavigationTests.swift` | `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`, `graphLocationBlockedWithoutArtifact` and the private helper `graphsBySource(_:)` (interim); `pageLocationUsesPageSurface`, `graphLocationOpensTheGivenGraph` (stay). |
| Tests | `macos/ProvenenciaTests/PlaceRegistryTests.swift` | `resolvesSourceDetail` and the place table assert the exact `.sourceDetail` key list. |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` | Graph section tests; helper `entryGraphLocation(_:)` (reads `model.graphs.location(for:)`, stays); `session.apply(location:)` warms place keys. `FakeStore.graphProgressByGraph` seeds progress. |
| Project | `macos/Provenencia.xcodeproj/project.pbxproj` | Lists each Swift file (no synchronized groups): new files are added, deleted ones removed. |

**Precedent.** S12-11 ([#549](https://github.com/mendahu/provenencia/pull/549)) put the Source page's graph rows on session reads instead of a page copy: progress gets the same session home. S12-13 ([#562](https://github.com/mendahu/provenencia/pull/562)) built `evidenceGraphsBySource` and promoted the checklist frame to the kit at its second caller (the shape for `PVFlowLayout`). S12-09 ([#535](https://github.com/mendahu/provenencia/pull/535)) reserved removed proto fields (the shape for `primary_count`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The row is two lines and nothing expands (board, user 2026-10-26). Every Source with an Artifact shows one token per graph; a Source with no Artifact shows one blocked "Needs an artifact" token. | Agreed board S12-D5. | The accordion first briefed — the user adopted the board. |
| D-2 | `SourcesListRow` *(new)*, `macos/App/Features/Sources/SourcesListRow.swift`, is one `ForEach(rows)` element with `.id(source.id)`; `listBody` has one call site: `SourcesListRow(source:model:)`, which reads progress and the graph list through the session (observation tracks the `@Observable` handles). | `ListScrollAnchorModifier` anchors on Source ids; one call site for the next row parameter (review M-3). | Two row branches; tokens as separate list items. |
| D-3 | `SourcesGraphCountLine` *(new)*, `macos/App/Features/Sources/SourcesGraphCountLine.swift`: `progress: EvidenceGraphProgress?`, `countsLoading: Bool`; body moved from `SourcesSplitRow.graphCountLine`; `.accessibilityElement(children: .ignore)` with label `L10n.Sources.graphCountSummary(progress:countsLoading:separator: ", ")`. `L10n.Sources.graphCountSummary(progress:countsLoading:separator:)` *(new func, no key)* returns "counts loading", "0 subjects{sep}not started" or "N subjects{sep}M observations". The token's help uses separator `" · "`; spoken labels use `", "`. | One count rule for the card, the token help and VoiceOver (review M-2; user asked for help that explains the compact count). | A copy of the branches per caller. |
| D-4 | Removed with the accordion: no expanded state anywhere. | D-1. | — |
| D-5 | Graph status: `.empty` when `subjectCount == 0`; `.promoted` when `filedCount > 0`; else `.inProgress`. A graph missing from the progress map is empty. | User, 2026-10-26: In progress = at least one Subject and none promoted; one filed Subject makes it done. `subjectCount` is what the count shows, so "0 subjects" is Empty. | `primaryCount == 0` as Empty — a graph with only value Subjects would read Empty beside "2 subjects". A new "ever promoted" field — `filedCount > 0` is that fact; unpromoting every Subject returns the graph to In progress, which matches "none promoted". |
| D-6 | `.filed` becomes `.promoted` with raw value `"promoted"`. `SourcesListSettings.init(from:)` decodes the status as a `String`, maps `"filed"` to `.promoted`, then `EvidenceStatus(rawValue:)`; any other unknown value falls back to nil. | One word in code, storage and identifiers (`sources.status.promoted`); the alias lives where stale settings are handled (review M-6). | Keeping raw `"filed"`; dropping saved filters. |
| D-7 | A Source matches a status when any graph of `session.evidenceGraphsRead(sourceID:)`'s `.loaded` value has it; `.loaded([])` matches Empty. While the read is `.loading` or `.failed`, `isAwaitingFilterData` and `statusCountsError` behave as today. `SourcesModel.statuses(for source:) -> Set<EvidenceStatus>` *(new)*; `visibleSources` filters with `statuses(for:).contains(status)`. | User, 2026-10-26: any graph; the same reading for every status; the token line always shows every graph (frame 4). | All graphs must match; hiding tokens that do not match. |
| D-8 | The `.sourceDetail` place adds `.evidenceGraphProgress(project: project)` as the last of its query keys. `SourceGraphsSection.progress(for:)` and `countsLoading` *(new)* forward to D-13; its `rows` and `loadError` forward to D-17. `SourcePageModel.warmFromSession()` is not changed; page tests warm with `session.apply(location: .sourcePage(...))`. | Without the key, `refreshEvidenceGraphProgress` skips and a restored page never loads counts (AC-3.2). Working rule "One graph list". | A per-Source progress RPC; a third hand key list in `warmFromSession`. |
| D-9 | Delete `SourcesListNavigation.entryGraph(sourceID:in:)`, `graphLocation(for:graphs:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`; `SourcesModel.entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)`; the tests `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`, `graphLocationBlockedWithoutArtifact` and the helper `graphsBySource(_:)`. `graphLocation(for:graph:)` and `graphBlockedReason(for:)` stay; tokens use them. | Working rule "High tolerance for churn"; the map's "Until S12-14 adds per-graph Open". | Keeping them. |
| D-10 | The Source link is one `Button` (cover, title, type) with a feature-private `SourcesLinkButtonStyle` *(new, replaces `SourcesSplitZoneButtonStyle`)*: on hover or press the title is `PVColor.textLinkHover` and underlined, the cover gets a 1 pt `PVColor.borderStrong` outline, animated with `PVMotion.fastStandard` unless reduce motion; no fill. Padding `PVSpacing.space2`, corner `PVRadius.sm`, the focus ring as `PVButton`'s. It keeps today's identifier `sources.row.<id>.page`, label `L10n.Sources.openSourcePage`, value `PVLastUpdated.label("\(title), \(ref)", updatedAt:in:)`, and the `pvRightClickMenu` item `L10n.Sources.copySourceEntry`. | GT-3, GT-10; the page zone's existing accessibility carried over. | `PVButton(.link)` — underlines at rest, no cover. |
| D-11 | `SourcesGraphToken` *(new)*, `macos/App/Features/Sources/SourcesGraphToken.swift`, one view with `enum Kind { case graph(CatalogEvidenceGraph, progress: EvidenceGraphProgress?, countsLoading: Bool, onOpen: (() -> Void)?), needsArtifact, listLoading, listFailed(Error) }`. Shape: height 26, padding horizontal `PVSpacing.space4 + PVSpacing.spacePx`, `PVRadius.sm`, 1 pt `PVColor.borderDefault` on `PVColor.surfaceCard`; `PVIcon(.network, size: 13)`, name (`PVFont.body(size: PVTypeScale.caption)`, one line, `.frame(maxWidth: 300)` truncating tail), count (`Text(verbatim: "\(subjects) · \(observations)")` in `PVFont.mono(size: PVTypeScale.micro)` muted, or italic `L10n.Sources.graphNotStarted`, or `PVSkeletonBar(width: 32, height: 6)` while counts load). Hover: fill `PVColor.surfaceHover`, border `PVColor.borderStrong`. Open graph: a `Button` calling `onOpen`, help `graphCountSummary(…, separator: " · ")`, label `L10n.Sources.openGraphNamed(name)` + ", " + `graphCountSummary(…, separator: ", ")`, identifier `sources.row.<graph.sourceID>.graph.<graph.ref>`. Blocked graph (`onOpen == nil`): dashed (`StrokeStyle(lineWidth: 1, dash: [4, 3])`, as `PVCard`'s dashed border), `PVColor.textFaint`, text `L10n.Sources.graphNeedsArtifact(name)`, no count, help and hint `graphNeedsArtifactTooltip`, focusable, does nothing. `needsArtifact`: dashed, italic `L10n.Sources.needsArtifact`, help and hint `needsArtifactTooltip`, identifier `sources.row.<id>.graphBlocked`. `listLoading`: `PVSkeletonBar(width: 96, height: 6)` inside the token frame, label `L10n.DesignSystem.listLoading`. `listFailed`: `PVIcon(.danger, size: 13)` in `PVColor.danger`, help and label `L10n.Errors.message(for:)`, inert. | GT-4 – GT-6, GT-10; one view owns every token state (board inventory). | `PVChip` — selectable pill; per-state views. |
| D-12 | `SourcesModel.tokenLine(for source: CatalogSource) -> EvidenceGraphsRead` *(new)* returns D-17's read, except `.loaded([])` becomes `.loading` (a just-created Source before the list refresh). The row maps: no Artifact (`!source.hasArtifact`) → one `.needsArtifact` token; `.loading` → one `.listLoading`; `.failed(e)` → one `.listFailed(e)`; `.loaded(graphs)` → one `.graph` per graph with `model.graphProgress(for:)`, `model.graphCountsLoading`, and `SourcesListNavigation.graphLocation(for:graph:)` as `onOpen`. | One model member feeds the row (review X-8). | Reading handles in the view. |
| D-13 | `WorkspaceSession+EvidenceGraphs.swift` gains `evidenceGraphProgressKey`, `evidenceGraphProgressHandle`, `evidenceGraphProgress(graphID:)`, `evidenceGraphCountsLoading` (progress or graph-list handle `isAwaitingFirstValue`) *(new)*. `refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress` use the key; `SourcesModel.graphProgress(for:)` and `graphCountsLoading` forward; `SourcesModel.evidenceGraphProgressKey(for:)` and `progressHandle` are deleted (tests move to `session.evidenceGraphProgressKey`). | One read path for the progress map, the graph list's shape (review M-1). | `SourceGraphsSection` copying `SourcesModel`'s reads. |
| D-14 | `QueryHandle.isAwaitingFirstValue: Bool` *(new)* = `value == nil && status != .ready && status != .error`. `graphCountsLoading`, the types branch of `isAwaitingFilterData`, D-13 and D-17 use it. | One meaning of "loading" (review M-4); `.ready` always carries a value, so the types branch's answer is unchanged. | Another hand-written check. |
| D-15 | `SourcesListContent` loses its `progressHandle` parameter; `SourcesSplitRowProgress`, the inline row, `SourcesSplitRow`, `SourcesSplitLayout` and `SourcesSplitRow.swift` are deleted. `captionBand` becomes one text, `L10n.Sources.listCaption` ("Source · evidence graphs"), leading `PVSpacing.gutterPage`, same font, case and band as today. | GT-1, GT-2; review M-3. | Keeping the split track for the caption. |
| D-16 | `SourcesListRow` layout: `VStack(alignment: .leading, spacing: PVSpacing.space4)`; padding vertical `PVSpacing.space5`, horizontal `PVSpacing.gutterPage`; bottom `PVDivider()`. Top line: `HStack(spacing: PVSpacing.space5)` of the Source link (`CachedThumbnail` size `SourcesListRow.coverSize` (44), then title `PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium)` one line and type `PVFont.body(size: PVTypeScale.caption)` muted), `Spacer`, `PVListDetailCell(detail: PVLastUpdated.detail(...), width: PVLastUpdated.columnWidth)`, ref (`PVFont.mono(size: PVTypeScale.micro)` muted, fixed size). Token line: `PVFlowLayout(spacing: PVSpacing.space3, lineSpacing: PVSpacing.space3)`, leading padding `SourcesListRow.coverSize + PVSpacing.space5` (56). The `pvRightClickMenu` stays on the Source link (D-10). `.accessibilityElement(children: .contain)`, label the Source title, identifier `sources.row.<id>`. | GT-1, GT-9 (tab order follows layout order), GT-11; frame 9 metrics mapped to tokens (12 = `space5`, 8 = `space4`, 6 = `space3`, 4 = `space2`). | A menu over the whole row (it would also catch token right-clicks); new spacing literals. |
| D-17 | `WorkspaceSession+EvidenceGraphs.swift` gains `enum EvidenceGraphsRead { case loading, failed(Error), loaded([CatalogEvidenceGraph]) }` and `evidenceGraphsRead(sourceID:)` *(new)*: `failed` when `evidenceGraphsHandle` is `.error` with no value; `loading` while it `isAwaitingFirstValue` or has no handle; else `loaded(evidenceGraphs(sourceID:) ?? [])`. | One rule for "a Source's graphs: loading, failed or rows", shared by the list and the page (review M-10). | Each feature deriving it. |
| D-18 | `primary_count` goes: `graphprogress.Progress.PrimaryCount` is deleted (`add` keeps `vocabkey.IsPrimaryType` for `FiledCount`); `engine.proto` `EvidenceGraphProgress` gets `reserved 5; reserved "primary_count";` and `./scripts/generate-proto.sh` regenerates; `api/ffi/handlers/evidence_graph_progress.go` drops the field; `GoStore`, `FakeStore` and `EvidenceGraphProgress` drop `primaryCount` (FakeStore keeps its primary type set for `filedCount`). The doc comments say filed counts primary-kind Subjects with an accepted Identity Claim. `TestPrimaryAndFiledCounts` becomes `TestFiledCounts`. | User, 2026-10-26; no reader after D-5 (review M-5). | Swift only; keeping the field. |
| D-19 | The Citation composer's private `FlowLayout` moves to `DesignSystem/Components/Flow/PVFlowLayout.swift` *(new)* as `struct PVFlowLayout: Layout` with the same `spacing` / `lineSpacing` initializer and algorithm; the two composer call sites use it and the private copy is deleted. | Second call site (`add-ui-component`: promote on the second caller), as S12-13 did for the checklist frame. | A second private copy in Sources. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcesListRow` *(new)* | view | `SourcesSplitRow`, `SourcesSplitRowProgress` | The agreed board replaces the split row; one wrapper replaces two call sites (D-2, D-15). |
| `SourcesLinkButtonStyle` *(new)* | button style | `SourcesSplitZoneButtonStyle`, `PVButton(.link)` | Replaces the zone style; `.link` underlines at rest (D-10). |
| `SourcesGraphToken` *(new)* | view | `PVChip`, `SourcesSplitRow.graphZone` | A link to a place, not a selectable pill; the zone body is replaced (D-11). |
| `SourcesGraphCountLine` *(new)* | view | `SourcesSplitRow.graphCountLine` (private) | Lifted for the Source page card (D-3). |
| `PVFlowLayout` *(new, promoted)* | kit layout | `CitationComposerAuthoringBox.swift` private `FlowLayout` | Moved, second caller (D-19). |
| `L10n.Sources.graphCountSummary` *(new func)* | L10n func | `graphZoneAccessibility` | Its branches, one formatter for help and speech (D-3). |
| `EvidenceGraphsRead`, `evidenceGraphsRead(sourceID:)` *(new)* | enum / session read | `SourceGraphsSection.rows` / `loadError` | Lifted to the session (D-17). |
| `evidenceGraphProgressKey`, `evidenceGraphProgressHandle`, `evidenceGraphProgress(graphID:)`, `evidenceGraphCountsLoading` *(new)* | session reads | `evidenceGraphsKey`, `evidenceGraphsHandle` | Moved from `SourcesModel` (D-13). |
| `QueryHandle.isAwaitingFirstValue` *(new)* | property | `isSettled` | A different question (D-14). |
| `SourcesModel.statuses(for:)`, `tokenLine(for:)` *(new)* | func | `graphProgress(for:)`, `visibleSources` | Model API the view calls (D-7, D-12). |
| `SourceGraphsSection.progress(for:)`, `countsLoading` *(new)* | forwarders | `SourceGraphsSection.rows` | Forward to D-13. |
| L10n keys `statusPromoted`, `listCaption`, `openGraphNamed`, `graphNeedsArtifact` *(new)* | L10n | `statusFullyPromoted`, `columnSource` / `columnEvidenceGraph`, `openGraph`, `needsArtifact` | Replace or join them (§6 Strings). |

Searched: `FlowLayout`, `: Layout`, `PVChip`, `textLinkHover`, `PVButtonVariant.link`, `pvRightClickMenu`, `PVListDetailCell`, `graphCountLine`, `graphZoneAccessibility`, `evidenceGraphProgress`, `progressHandle`, `entryGraph`, `isSettled`, `primaryCount` / `primary_count` / `PrimaryCount` (all callers listed in §4).

**Builds on.** `evidenceGraphsBySource` (S12-13), `SourcesListNavigation.graphLocation(for:graph:)` / `graphBlockedReason(for:)` (S12-11), `PVLastUpdated` / `PVListDetailCell`, `pvRightClickMenu`, the `PlaceRegistry` query keys ([`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md)), `WorkspaceSession.apply`'s per-graph progress patch, the proto `reserved` precedent.
**Deletes.** `SourcesSplitRow.swift` (`SourcesSplitLayout`, `SourcesSplitRow`, `SourcesSplitZoneButtonStyle`, `graphCountLine`); `SourcesSplitRowProgress` and the inline row; the interim helpers and their three tests and helper (D-9); `SourcesModel.evidenceGraphProgressKey(for:)` and `progressHandle`; the composer's private `FlowLayout`; `PrimaryCount` / `primary_count` / `primaryCount` (D-18); L10n `statusFullyPromoted`, `columnSource` (Sources), `columnEvidenceGraph`, `openGraph`, `graphZoneAccessibility` and their catalog entries.
**Next adds.** S12-15's graph hit opens the location a token opens. A status mark on a token reads `session.evidenceGraphProgress(graphID:)` and `EvidenceStatus(_:)`; a reader outside Sources moves `EvidenceStatus` out of `SourcesModel` first.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphprogress` | `Progress.PrimaryCount` deleted; `add` keeps counting `FiledCount` for primary kinds (D-18). |

### FFI

`api/proto/engine.proto` `EvidenceGraphProgress`: `primary_count = 5` replaced by `reserved 5; reserved "primary_count";`. Handler `api/ffi/handlers/evidence_graph_progress.go` drops `PrimaryCount`. Regenerate with `./scripts/generate-proto.sh` (updates `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Row | snowflake (new) | `Features/Sources/SourcesListRow.swift` *(new)* | D-2, D-16. |
| Source link | snowflake (new) | `Features/Sources/SourcesListRow.swift` | D-10. |
| Graph token | snowflake (new) | `Features/Sources/SourcesGraphToken.swift` *(new)* | D-11, D-12. |
| Token line wrap | `PVFlowLayout` (promoted) | `DesignSystem/Components/Flow/PVFlowLayout.swift` *(new)* | D-19. |
| Full count line | snowflake (lifted) | `Features/Sources/SourcesGraphCountLine.swift` *(new)* | D-3. |
| Cover | `CachedThumbnail` / `PVThumbnail` | `Platform/CachedThumbnail.swift` | 44. |
| Updated | `PVListDetailCell` + `PVLastUpdated` | kit | Unchanged. |
| Row menu | `pvRightClickMenu` | kit | On the Source link. |
| Caption | — | `SourcesListView.swift` `captionBand` | D-15. |
| Status menu | `PVSelect` | `SourcesListView.swift` `statusMenu` | Labels only. |
| Source page card | snowflake, extended | `SourcePageGraphsView.swift` `SourcePageGraphRow` | `SourcesGraphCountLine(progress: model.graphs.progress(for: graph), countsLoading: model.graphs.countsLoading)` on its own line after the ref, before the description, inside the combined element. |

**Frames** (board `S2-04 Sources list`, content host 1136 wide):

1. *Interactive list.* Header "Sources", "8 sources", Add source; toolbar (search "Search title or ref", All types, All statuses, "Sorted by date added"); caption "SOURCE · EVIDENCE GRAPHS". Rows: 1851 census (Census schedule, "Today, 3:15 PM", `SRC-7KQ2D`) with tokens "Default *not started*", "Jones household 12 · 40", "Smith household 7 · 19"; photograph with "Default 12 · 48"; several one-graph rows; an obituary and a 1891 census with a dashed italic "Needs an artifact" token; one row with Updated "—".
2. *Row states.* 2a: hover on the census link — title in link color, underlined, cover outline darker, no fill. 2b: Updated help. 2c: a dashed faint token "Unlinked needs an artifact" with its help. 2d: the token line as one skeleton token (graph list pending); the census tokens with a short rule in place of each count (counts pending). 2e: one token with the danger glyph and the error as help. 2f: the row menu (`copySourceEntry`) at the pointer. 2g: the copy at the top with one "Default *not started*" token. 2h: the Copy Source form (other board).
3. *Toolbar.* Status menu open with All statuses, Empty, In progress, Promoted; "In progress", count line "2 of 8 sources"; search "smith"; nothing matches; status set with counts pending (spinner); status set with counts failed (callout).
4. *Status per graph.* The census under Empty, In progress and Promoted, all three tokens each time.
5. *Minimum width (650).* Title and type truncate; Updated and ref keep their width; twelve tokens wrap; the 60-character Pryke name truncates in its token.
6–8, 10. Empty destination, page-level states, Add source, cover tile: unchanged from today.
9. *Metrics.* As D-16 and D-11.
11. *Navigation.* Source link → Source page; token → that graph's canvas; blocked token inert; nothing expands.
12. *Source page.* Census cards: Default "0 subjects · not started"; Jones "12 subjects · 40 observations" then its description; Smith "7 subjects · 19 observations".
13. *Building blocks.* As §6 macOS.

States: graph list pending / failed, counts pending, blocked graph, no Artifact (D-11, D-12); twelve tokens wrap. Accessibility: row group labelled with the title; Source link label and value as today with the copy action in its actions; tokens "Open graph {name}, {count summary}" with the blocked hint; tab order link then tokens.

### Strings

L10n keys *(new)*: `L10n.Sources.statusPromoted` = `sources.list.status.promoted` "Promoted" ("Status filter option: graphs with at least one promoted subject"); `L10n.Sources.listCaption` = `sources.list.caption` "Source · evidence graphs" ("Caption band above the Sources list"); `L10n.Sources.openGraphNamed(_:)` = `sources.list.openGraphNamed` "Open graph %@" ("VoiceOver label of a Sources list graph token; argument is the graph name"); `L10n.Sources.graphNeedsArtifact(_:)` = `sources.list.graphNeedsArtifact` "%@ needs an artifact" ("Blocked Sources list graph token; argument is the graph name"). New func, no key: `graphCountSummary(progress:countsLoading:separator:)`. Changed comment: `sources.list.status.inProgress` "at least one subject and none promoted". Removed: `statusFullyPromoted`, `sources.list.columnSource`, `columnEvidenceGraph`, `openGraph`, `graphZoneAccessibility`. Kept: `graphSubjectCount`, `graphObservationCount`, `graphCountLine`, `graphNotStarted`, `graphCountsLoading`, `needsArtifact`, `needsArtifactTooltip`, `graphNeedsArtifactTooltip`, `openSourcePage`. The token count is `Text(verbatim:)` digits and a middle dot (`add-localized-string`: verbatim for numbers and glyphs).

## 7. Steps

**STEP-1 — One read of a Source's graphs and of graph progress.**
Files: `QueryHandle.swift`, `WorkspaceSession+EvidenceGraphs.swift`, `WorkspaceSession.swift`, `SourcesModel.swift`, `SourceGraphsSection.swift`, `SourcesListView.swift` (key call sites only), `SourcesModelTests.swift`, `WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Do: D-13, D-14, D-17; `SourceGraphsSection.rows` / `loadError` forward to D-17. No behavior change.
Check: T-9, T-10 pass; `SourcesModelTests` and `SourcePageModelTests` pass; `rg -n 'SourcesModel\.evidenceGraphProgressKey|evidenceGraphProgress\(project: projectKey\)' macos/App` lists only `WorkspaceSession+EvidenceGraphs.swift`.

**STEP-2 — Read status per graph.**
Files: `SourcesModel.swift`, `SourcesListSettings.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`, `SourcesListSettingsTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-5, D-6, D-7. The test helper becomes `progress(_ id:, subjects:, filed:)`. `filtersCombineAndClear` seeds `filed: 0` where it expects In progress (s1, s2, s5) and keeps `filed: 1` where it expects Promoted; `settingsRestoreOnNewModel` and `statusCountsFailureSurfacesAnErrorAndUnfilteredCount` use `.promoted`.
Check: T-1, T-2, T-3 pass with every other `SourcesModelTests` test.

**STEP-3 — Drop `primary_count`.**
Files: `core/database/graphprogress/graphprogress.go`, `graphprogress_test.go`, `api/proto/engine.proto`, generated `engine.pb.go` / `engine.pb.swift`, `api/ffi/handlers/evidence_graph_progress.go`, `evidence_graph_progress_test.go`, `GoStore.swift`, `FakeStore.swift`, `GenealogyStore.swift`, `SourcesModelTests.swift`, `WorkspaceSessionTests.swift`. Skill: `add-ffi-handler`.
Do: D-18.
Check: `./scripts/generate-proto.sh`; `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphprogress/... ./api/ffi/handlers/...` passes; `rg -n -i 'primary_?count' core api macos/App macos/ProvenenciaTests --glob '!*.pb.*'` returns nothing.

**STEP-4 — Promote the flow layout.**
Files: `DesignSystem/Components/Flow/PVFlowLayout.swift` *(new)*, `CitationComposerAuthoringBox.swift`, `project.pbxproj`. Skill: `add-ui-component`.
Do: D-19.
Check: the app builds; the composer's word and term rows look as before (open a Citation composer).

**STEP-5 — One count line; counts on the Source page.**
Files: `SourcesGraphCountLine.swift` *(new)*, `SourcesSplitRow.swift` (draw it until STEP-7), `L10n.swift`, `PlaceRegistry.swift`, `PlaceRegistryTests.swift`, `SourceGraphsSection.swift`, `SourcePageGraphsView.swift`, `SourcePageModelTests.swift`, `SourcesModelTests.swift`, `project.pbxproj`. Skill: `add-workspace-place`, `add-localized-string`.
Do: D-3, D-8; the card line (§6 macOS). `graphZoneAccessibility` now builds on `graphCountSummary` until STEP-7 deletes it. `PlaceRegistryTests` `resolvesSourceDetail` and the place table's `.sourceDetail` row gain `.evidenceGraphProgress(project: project)` at the end.
Check: T-7, T-8, T-11 and `PlaceRegistryTests` pass.

**STEP-6 — The graph token.**
Files: `SourcesGraphToken.swift` *(new)*, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-11 (not yet drawn by the list).
Check: the app builds; T-12 passes.

**STEP-7 — The two-line row.**
Files: `SourcesListRow.swift` *(new)*, `SourcesListView.swift`, `SourcesSplitRow.swift` (deleted), `SourcesModel.swift`, `SourcesListNavigation.swift`, `SourcesListNavigationTests.swift`, `SourcesModelTests.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`. Skill: `add-ui-component`.
Do: D-1, D-2, D-9, D-10, D-12, D-15, D-16.
Check: T-5, T-6 pass; `rg -n 'SourcesSplit|SourcesListNavigation\.entryGraph|model\.entryGraph|graphListLoaded|graphLocation\(for: source\)|graphsBySource\(|graphZoneAccessibility|statusFullyPromoted|columnEvidenceGraph' macos/App macos/ProvenenciaTests` returns nothing (`SourcePageModelTests.entryGraphLocation` is not matched and stays); manual check steps 1–5.

**STEP-8 — Docs.**
Files: `../completed.md`, `../deployment-plan.md` (map row; working rule "A graph is a place" drops its interim first-graph sentence), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-tokens.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesModelTests.swift` `evidenceStatusFromCounts` (rewritten) | nil → empty; subjects 0 → empty; subjects 5, filed 0 → inProgress; subjects 5, filed 1 → promoted; subjects 2, filed 0 → inProgress | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-5 |
| T-2 | Swift | `SourcesModelTests.swift` `statusMatchesAnyGraphOfTheSource` *(new)* | One Source, three graphs (empty, in progress, promoted) → in `visibleSources` under each status; a second Source with one empty graph only under Empty | AC-2.5, D-7 |
| T-3 | Swift | `SourcesListSettingsTests.swift` `storedFiledStatusRestoresPromoted` *(new)* | `{"status":"filed"}` → `.promoted`; `{"status":"promoted"}` → `.promoted`; `{"status":"bogus"}` → nil, other fields kept; `EvidenceStatus.promoted.label` resolves to "Promoted" | AC-2.6, AC-2.7, D-6 |
| T-4 | Go | `core/database/graphprogress/graphprogress_test.go` `TestGraphProgress`, `TestFiledCounts` (renamed) | Today's cases without `PrimaryCount`: a filed primary Subject counts in `FiledCount`; a bridge Subject does not | AC-4.1, D-18 |
| T-5 | Swift | `SourcesListNavigationTests.swift` | The three interim tests and `graphsBySource` deleted; `graphLocationOpensTheGivenGraph` and `pageLocationUsesPageSurface` pass unchanged | D-9 |
| T-6 | Swift | `SourcesModelTests.swift` `tokenLineFollowsTheGraphList` *(new)* | Graph list not loaded → `tokenLine(for:)` `.loading`; failed → `.failed`; loaded without the Source → `.loading`; loaded with three graphs → `.loaded` in name order; with progress on `g-jones` only, `graphProgress(for:)` returns it and nil for the others; `session.apply` naming `g-jones` with new counts changes only `g-jones` | AC-1.1, AC-1.2, AC-1.6, AC-1.9, D-12 |
| T-7 | Swift | `SourcePageModelTests.swift` `graphRowsReadTheirProgress` *(new)* | Two graphs, progress seeded for one; `session.apply(location: .sourcePage(...))` then `waitForQuery` on `session.evidenceGraphProgressHandle` → `model.graphs.progress(for:)` returns it for that graph, nil for the other; `countsLoading` false | AC-3.1, AC-3.2, D-8 |
| T-8 | Swift | `SourcePageModelTests.swift` `aGraphWriteMovesOnlyItsRow` *(new)* | T-7, then the store's count changes and `session.apply` names that graph → its progress updates; the other stays nil | AC-3.3 |
| T-9 | Swift | `WorkspaceSessionTests.swift` `evidenceGraphsReadFollowsTheHandle` *(new)* | No handle → `.loading`; loading → `.loading`; failed without value → `.failed`; loaded → `.loaded` with the Source's graphs, `[]` for an unknown Source | D-17 |
| T-10 | Swift | `WorkspaceSessionTests.swift` `progressReadsComeFromTheSession` *(new)* | Before warm: `evidenceGraphCountsLoading` true, `evidenceGraphProgress(graphID:)` nil; after both handles load: false and the seeded value; `.ready` or `.error` → `isAwaitingFirstValue` false | D-13, D-14 |
| T-11 | Swift | `SourcesModelTests.swift` `graphProgressCopyMatchesBoard` (rewritten) | `graphCountSummary(progress: nil, countsLoading: true, separator: " · ")` == "counts loading"; zeros → "0 subjects · not started"; 12 / 40 → "12 subjects · 40 observations"; with `", "` → "12 subjects, 40 observations"; `graphCountLine(subjects: 12, observations: 48)` unchanged | AC-1.2, AC-3.1, D-3 |
| T-12 | Swift | `SourcesModelTests.swift` `graphTokenCopyMatchesBoard` *(new)* | `openGraphNamed("Jones household")` == "Open graph Jones household"; `graphNeedsArtifact("Unlinked")` == "Unlinked needs an artifact"; `listCaption` resolves to "Source · evidence graphs" | AC-1.4, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests`, `SourcesListNavigationTests`, `SourcesListSettingsTests`, `PlaceRegistryTests`, `SourcePageModelTests`, `WorkspaceSessionTests`, `TestGraphProgress`, `TestListEvidenceGraphProgress`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; rg -n 'fullyPromoted|Fully promoted|columnEvidenceGraph' macos returns nothing
- [ ] UI matches the brief inventory; the only kit addition is PVFlowLayout (promoted)
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; working rule "A graph is a place" drops the interim first-graph sentence; completed.md: one short paragraph of what shipped
- [ ] Brief S12-D5: already archived (board agreed 2026-10-26); design/README.md says "built by S12-14 (#N)"
- [ ] Domain doc: none — the status rule is the map's working rule "Status is per graph"
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source and three graphs: Default (no Subjects), Jones household (Subjects, not promoted), Smith household (promoted once, one Subject filed of several). The census row shows "Default *not started*", "Jones household 12 · 40", "Smith household 7 · 19" under its title; hovering Jones's token shows "12 subjects · 40 observations"; a Source with no Artifact shows "Needs an artifact" (AC-1.1, AC-1.2, AC-1.5). Relaunch and watch the list load: a skeleton token first, then tokens with a short rule for counts, then counts (AC-1.6, AC-1.7).
2. Hover the census title: link color and underline, no fill. Click it: the Source page opens (AC-1.8). Back. Click Jones household's token: its canvas opens on that graph (AC-1.3).
3. On the census Source page, add a new page Artifact, add a graph "Unlinked" linked only to it, then delete that Artifact: back in the list, Unlinked's token is dashed, reads "Unlinked needs an artifact", and does nothing on click (AC-1.4).
4. Add an Observation on Jones's canvas; Back to the list. Only Jones's token moved (AC-1.9).
5. Status filter: Empty, In progress (Jones, AC-2.2) and Promoted (Smith, AC-2.3) each show the census with all its tokens. The last option reads Promoted. Narrow the window to its minimum with a twelve-graph Source: the tokens wrap and a long name truncates (AC-1.10). Tab from the census link: focus moves through its tokens in name order.
6. Open the census Source page: each card shows its count line under the ref (AC-3.1). Quit and relaunch with that page restored: the count lines load (AC-3.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Source with many graphs makes a tall row | Tokens wrap and truncate at 300; the list is lazy per row. |
| The project-wide progress read on every Source page open | One cached map shared with the list; patches are per graph. |
| A graph with only value Subjects can never become Promoted | Accepted (D-5): Promote files only primary kinds; it reads In progress, which is true. |
| `PVFlowLayout` measures with `.unspecified`, so a long token name proposes its ideal width | The token caps the name with `.frame(maxWidth: 300)`; placement proposes the measured size, so the name truncates. |

**Stop and report if:**

- `WorkspaceSession.refreshEvidenceGraphProgress` no longer patches by graph id, or `apply` no longer calls it for every graph in `effects.evidenceGraphs`.
- `ListScrollAnchorModifier` needs ids other than Source ids (D-2 cannot hold).
- A token inside `PVFlowLayout` does not truncate at 300 at the minimum width (frame 5): report with a screenshot.
- `pvRightClickMenu` on the Source link swallows the link's primary click or its VoiceOver action.
- Observation does not repaint `SourcesListRow` when the progress map changes without a `@Bindable` handle (D-2).
- Anything other than the files in §4 reads `primary_count` / `primaryCount` (D-18).

## 12. Open questions

None.

## 13. PR

**Title:** S12-14 — Graph counts in the Sources list and on the Source page
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-tokens.md`.
