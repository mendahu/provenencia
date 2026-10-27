# S12-14 — Graph counts in the Sources list and on the Source page

**Status:** Landed in [#576](https://github.com/mendahu/provenencia/pull/576) (2026-10-26)
**Landed:** Differences from the plan: `PersonDetailTests.presentationFollowsTheHandle` also moved to `QueryRead` values (it called the old `ConclusionDetailPresentation` init); `evidenceGraphsRead(sourceID:)` reads `.loading` for a loaded list with no graph for the Source, empty or not, as D-17 words it; `SourcesGraphToken` composes a private `SourcesGraphTokenFrame` for the border, fill and dashed stroke, and its style takes an `inert` flag for the list-failed token (focusable, no hover fill, no press); the `PVStroke` README line sits in the component table beside the `FlowLayout` row; the new files' project ids are random `5C...` values; `SourcesGraphCountLine` keeps the split row's three looks and speaks `graphCounts(_, style: .spoken)`; `FakeStore` still hard-codes `filedCount: 0` for graphs without a seeded override. The manual check and the stop conditions that need the running app (token truncation at the minimum width, `pvRightClickMenu` against the link's click and VoiceOver action, the link highlight reaching the label, Observation repainting a row without a `@Bindable` handle, tab order and focus rings) were not run in the app by the implementer.
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
**Review:** [`S12-14-sources-list-graph-tokens.review.md`](S12-14-sources-list-graph-tokens.review.md) — accordion plan: Fits with changes / Ready with changes, 2 rounds; token plan: Fits with changes / Ready, 2 rounds; every blocker Fixed

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
| `QueryHandle.read` and one session read of a Source's graphs and of graph progress (D-13, D-14, D-17) | STEP-1, T-9, T-10 |
| One count state and formatter; status per graph, any-graph Source match, Promoted label and stored value (GT-7, D-3, D-5 – D-7) | STEP-2, AC-2.1 – AC-2.7 |
| `primary_count` removed from Go, proto, handler and Swift (D-18) | STEP-3, AC-4.1 |
| `PVFlowLayout` promoted; `PVStroke.dashed` token with the existing dashed strokes moved onto it (D-19, D-20) | STEP-4 |
| Full count line view; the Source page cards draw it; the place warms the progress key (GT-12, D-8) | STEP-5, AC-3.1 – AC-3.3 |
| Graph token with its states (GT-4 – GT-6, GT-10, D-11) | STEP-6, AC-1.2, AC-1.4 – AC-1.7 |
| The two-line row, caption, Source link, token line; the split row and interim helpers deleted (GT-1 – GT-3, GT-8, GT-9, GT-11, D-1, D-2, D-9, D-10, D-12, D-15, D-16) | STEP-7, AC-1.1, AC-1.3, AC-1.8 – AC-1.10 |

| Out | Home |
| --- | --- |
| Graph search in the list; omnibar graph hits | S12-15 |
| Create, edit, delete or link graphs from the list | not planned (Source page owns them) |
| A Promote button, a status badge, or filed counts on a token or card | not planned (map scope) |
| A Source-level total of Subjects or Observations | not planned (map scope) |
| Shift-F10 opening the row menu (board frame 1) | not planned here: `pvRightClickMenu` has no keyboard trigger; logged in the review log for a kit PR |
| Moving the status rule into `graphprogress` | when a reader outside `SourcesModel` needs it (review M-5) |
| `SourcesModel.warmListQueries` warming the keys the place already warms | not planned here (review log, "Log it") |
| The "Sources split row" comment in `DesignSystem/Components/List/PVList.swift` | not planned (a file this PR does not touch; review X-28) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitLayout`, `SourcesSplitRow`, private `SourcesSplitZoneButtonStyle` | Two zones on a `minmax(0,1fr) | 248` track: the page zone (one button: `CachedThumbnail`, `space5` gap, title with its own `.foregroundStyle(PVColor.textPrimary)`, type, `PVListDetailCell` with `PVLastUpdated`, ref; `pvRightClickMenu` with `L10n.Sources.copySourceEntry`; label `L10n.Sources.openSourcePage`, value via `PVLastUpdated.label`) and the graph zone ("Open graph" + private `graphCountLine`, or the sunken "Needs an artifact" zone). The zone style wraps `PVHoverEffect` and adds its own reduce-motion `.animation`. |
| Count line | same file, private `graphCountLine` | Three states: hairline while `countsLoading && progress == nil`; "0 subjects · not started" when `subjectCount == 0`; else `L10n.Sources.graphCountLine`. `L10n.Sources.graphZoneAccessibility` repeats the branches. |
| List wiring | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListView`, `SourcesListContent` (`progressHandle` parameter), `listBody`, `captionBand`, `SourcesSplitRowProgress` | `listBody` builds `SourcesSplitRow` twice (`SourcesSplitRowProgress` with the handle, inline with `countsLoading: true` without). `captionBand` draws "Source" and "Evidence graph" on the split track. Each row is one `ForEach(rows)` element with `.id(source.id)`. |
| Status menu | same file, `statusMenu` | Options from `SourcesModel.EvidenceStatus.allCases`; identifiers `sources.status.<rawValue>`. |
| Model | `macos/App/Features/Sources/SourcesModel.swift` `EvidenceStatus` | Cases `empty`, `inProgress = "in_progress"`, `filed`; `init(_ progress:)`: empty when `primaryCount == 0`, in progress when `filedCount < primaryCount`, else filed. `.filed` labels `L10n.Sources.statusFullyPromoted`. |
| Model | same file, `evidenceGraphProgressKey(for:)`, private `progressHandle`, private `graphsHandle`, `graphProgress(for:)`, `graphCountsLoading`, `isAwaitingFilterData`, `statusCountsError`, `warmListQueries` | Progress key and handle looked up here; `graphCountsLoading` and `isAwaitingFilterData` hand-write "no value and not settled". |
| Model | same file, `entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)`, `visibleSources` | The interim "first graph" rule; the status filter reads the entry graph only. |
| Settings | `macos/App/Features/Sources/SourcesListSettings.swift` `SourcesListSettings.init(from:)` | Decodes each field on its own; `status` by raw value (`"filed"` stored today). |
| Navigation | `macos/App/Features/Sources/SourcesListNavigation.swift` | Header "Pure navigation helpers for the Sources list split-row (S5-08 / S5-D2)". `entryGraph(sourceID:in:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`, `graphLocation(for:graphs:)` (interim, list-only); `pageLocation(for:)`, `graphLocation(for:graph:)`, `graphBlockedReason(for:)` (stay). |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` | `refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress` build `CatalogQueryKey.evidenceGraphProgress(project:)` inline; the patch runs only when the map is cached. |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` | "The one place" the graph list's key and handle are looked up: `evidenceGraphsKey`, `evidenceGraphsHandle`, `evidenceGraphs(sourceID:)`, `evidenceGraphsBySource`. |
| Query handle | `macos/App/Features/Workspace/Session/QueryHandle.swift` `QueryHandle` | `status`, `value`, `error`, `isSettled`. `applySuccess` is the only path to `.ready` and always sets a value; `applyFailure` always sets `error`. `ConclusionListPresentation.init(value:isFetching:status:error:)` (`Features/Conclusions/ConclusionListPage.swift`) and `ConclusionDetailPresentation.init(value:isFetching:status:error:content:)` (`ConclusionDetailPage.swift`) fold these by hand; `macos/ProvenenciaTests/PersonsListTests.swift` calls the list init. |
| Docs | `docs/macos-client-patterns.md` (`QueryHandle` row) | Lists the handle's states and `isSettled`. |
| Place | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` `.sourcesList`, `.sourceDetail` | The list place warms `.evidenceGraphProgress(project:)`; the Source page place does not. |
| Source page | `macos/App/Features/Sources/SourceGraphsSection.swift` `SourceGraphsSection` | `rows`, `loadError` (reads `evidenceGraphsHandle`), `location(for:)`, `blockedReason(for:)`; no progress read. |
| Source page | `macos/App/Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Card: name, ref, description (combined accessibility element); Open (label `L10n.Sources.openEvidenceGraphLabel(name:)`, "Open %@"), Edit, Delete. |
| Source page | `macos/App/Features/Sources/SourcePageModel.swift` `warmFromSession()` | Tests and previews warm a hand list of keys; navigation warms through `PlaceRegistry`. |
| Kit | `macos/App/DesignSystem/Components/HoverEffect/PVHoverEffect.swift` `PVHoverEffect`; `macos/App/DesignSystem/Tokens/PVElevation.swift` `View.pvFocusRing(_:cornerRadius:)` | Hover, press, disabled and reduce motion for every kit `ButtonStyle`; the kit focus ring. |
| Kit | `StrokeStyle(lineWidth: 1, dash: [4, 3])` in `DesignSystem/Components/Card/PVCard.swift`, `DesignSystem/Components/EmptyState/PVEmptyState.swift`, `Features/Onboarding/OnboardingContributorRow.swift`, `Features/Names/NameValueEditorForm.swift`; `dash: isDashed ? [4, 3] : []` in `DesignSystem/Components/Chip/PVChip.swift`; `dash: (placed.isCited || selected) ? [] : [4, 3]` in `Features/EvidenceGraph/EvidenceCardView.swift` | Six hand copies of the dashed stroke; no stroke token in `DesignSystem/Tokens/`. |
| Kit | `macos/App/DesignSystem/README.md` component table | One row per kit component (S12-12 added `ArtifactChecklist`). |
| Flow layout | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` private `FlowLayout` | Wraps children by `sizeThatFits(.unspecified)`; two call sites in that file. |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Sources` | `columnSource` (`sources.list.columnSource`), `columnEvidenceGraph`, `openGraph`, `graphCountLine`, `graphSubjectCount`, `graphObservationCount`, `graphNotStarted`, `graphCountsLoading`, `graphZoneAccessibility`, `needsArtifact`, `needsArtifactTooltip`, `graphNeedsArtifactTooltip`, `statusFullyPromoted` (`sources.list.status.fullyPromoted`), `openSourcePage`, `copySourceEntry`, `openEvidenceGraphLabel(name:)`. |
| Go | `core/database/graphprogress/graphprogress.go` `Progress`, `add` | `PrimaryCount` sums primary-kind Subjects; `FiledCount` sums those with an accepted claim. |
| Wire | `api/proto/engine.proto` `EvidenceGraphProgress` | `primary_count = 5`; `reserved 4; reserved "uncited_count";` is the precedent. Handler `api/ffi/handlers/evidence_graph_progress.go` maps it; `macos/App/Platform/GoStore.swift`, `GenealogyStore.swift` `EvidenceGraphProgress.primaryCount`; `FakeStore.swift` `graphProgress(for:projectDir:)` builds `primaryCount:` from a local `primaryTypeIDs` and hard-codes `filedCount: 0`. `scripts/generate-proto.sh` needs `protoc`, `protoc-gen-go` and `protoc-gen-swift`. |
| Tests | `core/database/graphprogress/graphprogress_test.go` `TestGraphProgress`, `TestPrimaryAndFiledCounts`; `api/ffi/handlers/evidence_graph_progress_test.go` `TestListEvidenceGraphProgress` (asserts `GetPrimaryCount() != 1 || GetFiledCount() != 0`) | Assert `PrimaryCount`. |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` | `evidenceStatusFromCounts`, `filtersCombineAndClear` (seeds `primary: 2, filed: 1`, expects In progress), `settingsRestoreOnNewModel` (`.filed`), `statusCountsFailureSurfacesAnErrorAndUnfilteredCount` (`.filed`), `statusFilterWaitsForCounts`, `graphProgressCopyMatchesBoard` (`graphZoneAccessibility`), `warmListQueriesLoadsGraphProgress`; helper `progress(_:primary:filed:)`. |
| Tests | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` | Two `EvidenceGraphProgress` literals with `primaryCount:`; `session.apply(CatalogWriteEffects(evidenceGraphs: ["g-1"]))` is the effect precedent. |
| Tests | `macos/ProvenenciaTests/SourcesListNavigationTests.swift` | `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`, `graphLocationBlockedWithoutArtifact` and the private helper `graphsBySource(_:)` (interim); `pageLocationUsesPageSurface`, `graphLocationOpensTheGivenGraph` (stay). |
| Tests | `macos/ProvenenciaTests/PlaceRegistryTests.swift` | `resolvesSourceDetail` and the place table assert the exact `.sourceDetail` key list. |
| Tests | `macos/ProvenenciaTests/SourcePageModelTests.swift` | Graph section tests; helper `entryGraphLocation(_:)` (stays); `session.apply(location:)` warms place keys. `FakeStore.graphProgressByGraph` seeds progress. |
| Project | `macos/Provenencia.xcodeproj/project.pbxproj` | Lists each Swift file and group (no synchronized groups): new files and folders are added, deleted ones removed. |

**Precedent.** S12-11 ([#549](https://github.com/mendahu/provenencia/pull/549)) put the Source page's graph rows on session reads: progress gets the same session home. S12-13 ([#562](https://github.com/mendahu/provenencia/pull/562)) built `evidenceGraphsBySource` and promoted the checklist frame to the kit at its second caller, with a README row (the shape for `PVFlowLayout`). S12-09 ([#535](https://github.com/mendahu/provenencia/pull/535)) reserved removed proto fields (the shape for `primary_count`).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The row is two lines and nothing expands (board, user 2026-10-26). Every Source with an Artifact shows one token per graph; a Source with no Artifact shows one blocked "Needs an artifact" token. | Agreed board S12-D5. | The accordion first briefed — the user adopted the board. |
| D-2 | `SourcesListRow` *(new)*, `macos/App/Features/Sources/SourcesListRow.swift`, is one `ForEach(rows)` element with `.id(source.id)`; `listBody` has one call site, `SourcesListRow(source:model:)`, which reads progress and the graph list through the session (observation tracks the `@Observable` handles). | `ListScrollAnchorModifier` anchors on Source ids; one call site (review M-3). | Two row branches; tokens as list items. |
| D-3 | `SourcesGraphCounts` *(new)*, `macos/App/Features/Sources/SourcesGraphCounts.swift`: `enum SourcesGraphCounts: Equatable { case loading, notStarted, counts(subjects: Int, observations: Int) }` with `init(progress: EvidenceGraphProgress?, countsLoading: Bool)` (loading when `countsLoading && progress == nil`; not started when `subjectCount` is 0 or the progress is nil; else counts). `L10n.Sources.graphCounts(_:style:)` *(new func, no key)* with `enum style { compact, full, spoken }`: compact "12 · 40" / `graphNotStarted` / `graphCountsLoading`; full `graphCountLine(subjects:observations:)` / "0 subjects · not started" / `graphCountsLoading`; spoken "12 subjects, 40 observations" / "0 subjects, not started" / `graphCountsLoading`. The token, its help, the Source page line and VoiceOver all switch on or format this one value; views switch only for styling. | One rule for "loading / not started / N · M" (review M-12); help and visible text agree by construction (AC-1.2). | A free `separator:` string; branching per view. |
| D-4 | No expanded state anywhere. | D-1. | — |
| D-5 | Graph status: `.empty` when `SourcesGraphCounts(progress:countsLoading: false) == .notStarted`; `.promoted` when `filedCount > 0`; else `.inProgress`. | User, 2026-10-26: In progress = at least one Subject and none promoted; one filed Subject makes it done. Empty and "not started" are one predicate (D-3). | `primaryCount == 0` as Empty; a new "ever promoted" field — `filedCount > 0` is that fact; unpromoting every Subject returns the graph to In progress, which matches "none promoted". |
| D-6 | `.filed` becomes `.promoted` with raw value `"promoted"`. `SourcesListSettings.init(from:)` decodes the status as a `String`, maps `"filed"` to `.promoted`, then `EvidenceStatus(rawValue:)`; any other unknown value falls back to nil. | One word in code, storage and identifiers; the alias lives where stale settings are handled (review M-6). | Keeping raw `"filed"`; dropping saved filters. |
| D-7 | `SourcesModel.statuses(for source:) -> Set<EvidenceStatus>` *(new)*: for `.loaded(graphs)` of `session.evidenceGraphsRead(sourceID:)`, the statuses of its graphs; for `.loading` or `.failed`, `[]`. `visibleSources` filters with `statuses(for:).contains(status)`. `isAwaitingFilterData` and `statusCountsError` behave as today. | User, 2026-10-26: any graph; the token line always shows every graph (frame 4). | All graphs must match; hiding tokens that do not match. |
| D-8 | The `.sourceDetail` place adds `.evidenceGraphProgress(project: project)` as the last of its query keys. `SourceGraphsSection.progress(for:)` and `countsLoading` *(new)* forward to D-13; `rows` is the `.loaded` value of `evidenceGraphsRead(sourceID:)` (nil otherwise) and `loadError` its `.failed` error. `SourcePageModel.warmFromSession()` is not changed; page tests warm with `session.apply(location: .sourcePage(...))`. | Without the key, `refreshEvidenceGraphProgress` skips and a restored page never loads counts (AC-3.2). Working rule "One graph list". | A per-Source progress RPC; a third hand key list in `warmFromSession`. |
| D-9 | Delete `SourcesListNavigation.entryGraph(sourceID:in:)`, `graphLocation(for:graphs:)`, `graphBlockedReason(hasArtifact:graphListLoaded:location:)`; `SourcesModel.entryGraph(for:)`, `graphLocation(for:)`, `graphBlockedReason(for:)`; the tests `graphLocationOpensTheFirstGraph`, `graphBlockedReasonNeedsTheGraphList`, `graphLocationBlockedWithoutArtifact` and the helper `graphsBySource(_:)`. `graphLocation(for:graph:)` and `graphBlockedReason(for:)` stay; tokens use them. The `SourcesListNavigation` header becomes "Pure navigation helpers for the Sources list rows and tokens and the Source page graph cards." | Working rule "High tolerance for churn"; the map's "Until S12-14 adds per-graph Open"; no comment names a deleted type (review M-18). | Keeping them. |
| D-10 | The Source link is one `Button` whose label is `SourcesLinkLabel` *(new, private in `SourcesListRow.swift`)*: `HStack(spacing: PVSpacing.space4)` of `CachedThumbnail` (size `SourcesListRow.coverSize` = 44) and a `VStack(alignment: .leading, spacing: PVSpacing.space1)` of title (`PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium)`, one line) and type (`PVFont.body(size: PVTypeScale.caption)`, `textMuted`, one line). Style `SourcesLinkButtonStyle` *(new, private, replaces `SourcesSplitZoneButtonStyle`)*: `makeBody` wraps `PVHoverEffect(isPressed:hoverAnimation: PVMotion.fastStandard)` and sets `.environment(\.sourcesLinkHighlighted, showHover \|\| configuration.isPressed)` (a private `EnvironmentKey` *(new)*) on the label; padding `PVSpacing.space2`; `.pvFocusRing(isFocused, cornerRadius: PVRadius.sm)`, where `isFocused` is read by `@Environment(\.isFocused)` in a private body `View` that `makeBody` builds (as `PVButtonBody` does in `PVButton.swift`; a style struct's own environment is outside the button's focus scope); no fill, no hand-written hover state, reduce-motion check or `.animation`. `SourcesLinkLabel` reads the key: highlighted → title `PVColor.textLinkHover` with `.underline()` (SwiftUI's default underline; it has no offset control), and a `RoundedRectangle(cornerRadius: PVRadius.sm).strokeBorder(PVColor.borderStrong, lineWidth: 1)` overlay on the cover; else title `PVColor.textPrimary`, no overlay. The button keeps today's identifier `sources.row.<id>.page`, label `L10n.Sources.openSourcePage`, value `PVLastUpdated.label("\(title), \(ref)", updatedAt:in:)`, and the `pvRightClickMenu` item `L10n.Sources.copySourceEntry`. | GT-3, GT-10; hover, press, focus and reduce motion from the kit (review M-14); a style cannot recolor a label that sets its own color, so the label reads the state (review X-17). | `PVButton(.link)` — underlines at rest, no cover; a hand hover `@State`. |
| D-11 | `SourcesGraphToken` *(new)*, `macos/App/Features/Sources/SourcesGraphToken.swift`, takes `sourceID: String` and `kind: Kind`, `enum Kind { case graph(CatalogEvidenceGraph, counts: SourcesGraphCounts, onOpen: (() -> Void)?, blockedReason: String?), needsArtifact, listLoading, listFailed(Error) }`; a graph token is blocked when `blockedReason` is non-nil and uses that string for help and hint. Frame: height 26, horizontal padding `PVSpacing.space4 + PVSpacing.spacePx`, `PVRadius.sm`, 1 pt `PVColor.borderDefault` on `PVColor.surfaceCard`, `HStack(spacing: PVSpacing.space3)`. Content: `PVIcon(.network, size: 13)` in `textMuted`; name in `PVFont.body(size: PVTypeScale.caption)` `textPrimary`, one line, `.frame(maxWidth: 300, alignment: .leading)`, tail truncation; count: `.counts` → `Text(verbatim: graphCounts(_, style: .compact))` in `PVFont.mono(size: PVTypeScale.micro)` `textMuted`; `.notStarted` → `graphNotStarted` in `PVFont.body(size: PVTypeScale.micro)` italic `textMuted`; `.loading` → `PVSkeletonBar(width: 32, height: 6)`. Style `SourcesGraphTokenButtonStyle` *(new, private)*: wraps `PVHoverEffect`; on hover or press fill `PVColor.surfaceHover` and border `PVColor.borderStrong`; `.pvFocusRing(isFocused, cornerRadius: PVRadius.sm)` with `isFocused` read in a private body `View`, as D-10. Blocked states draw a dashed `PVStroke.dashed` border in `PVColor.borderDefault`, content in `PVColor.textFaint`, and use a `Button` with an empty action (not `.disabled`, which removes focus) and no hover fill. States: see the table below D-20. | GT-4 – GT-6, GT-10; one view owns every token state; kit hover and focus (review M-14); `sourceID` builds identifiers for states with no graph (review X-19). | `PVChip`; per-state views; `.disabled` for blocked tokens. |
| D-12 | `SourcesListRow` maps its Source to tokens: `!source.hasArtifact` → one `.needsArtifact`; else `session.evidenceGraphsRead(sourceID:)`: `.loading` → one `.listLoading`; `.failed(e)` → one `.listFailed(e)`; `.loaded(graphs)` → one `.graph` per graph with `SourcesGraphCounts(progress: session.evidenceGraphProgress(graphID:), countsLoading: session.evidenceGraphCountsLoading)` `onOpen` from `SourcesListNavigation.graphLocation(for:graph:)`, and `blockedReason` from `SourcesListNavigation.graphBlockedReason(for:)` — the one answer to "why can't this graph open" the Source page also reads (review M-19). | One read (review M-17: no list-only mapping). | A `SourcesModel` forwarder. |
| D-13 | `WorkspaceSession+EvidenceGraphs.swift` gains `evidenceGraphProgressKey`, `evidenceGraphProgressHandle: QueryHandle<[String: EvidenceGraphProgress]>?`, `evidenceGraphProgress(graphID:)`, `evidenceGraphCountsLoading` (true when either the progress or the graph-list handle is missing or its `read` is `.loading`) *(new)*. `refreshEvidenceGraphProgress` / `patchEvidenceGraphProgress` use the key; `SourcesModel.graphCountsLoading` forwards to `evidenceGraphCountsLoading`; `statusCountsError` reads `session.evidenceGraphsHandle` and `session.evidenceGraphProgressHandle` (`status == .error` → `error`, the graph list first, as today); `SourcesModel.evidenceGraphProgressKey(for:)`, `progressHandle` and `graphProgress(for:)` are deleted (tests move to the session reads; the row reads the session, D-12). The extension's header comment becomes "The session's cached Evidence graph list and graph progress: the one place their keys and handles are looked up, and the graph's linked Artifacts read from them." | One read path for the progress map (review M-1). | `SourceGraphsSection` copying `SourcesModel`'s reads. |
| D-14 | `QueryHandle.swift` gains `enum QueryRead<Value> { case loading, failed(Error), loaded(Value) }` with `var isLoading: Bool`, and `QueryHandle.read: QueryRead<Value>` *(new)*: `value` non-nil → `.loaded(value)`; else `status == .error` with `error` → `.failed(error)`; else `.loading`. `graphCountsLoading`, the types branch of `isAwaitingFilterData` (`typesHandle.read.isLoading`), D-13 and D-17 use it, and so do the two hand folds it replaces: `ConclusionListPresentation` gets `init(_ read: QueryRead<[Row]>, isFetching: Bool)` (`.loaded(rows)` → `.empty` when empty and not fetching, else `.rows(rows, refreshing: isFetching)`; `.failed(e)` → `.failed(L10n.Errors.message(for: e))`; `.loading` → `.loading`) and `ConclusionDetailPresentation` gets `init(_ read: QueryRead<CatalogConclusionDetail>, isFetching: Bool, content:)`, replacing their `value:isFetching:status:error:` inits; their call sites pass `handle.read, isFetching: handle.isFetching`; `PersonsListTests` builds the list presentation from `QueryRead` values with the same expectations. `docs/macos-client-patterns.md`'s `QueryHandle` row adds "`read` folds value / error / loading". | One definition on the type that owns the state (review M-11); `.ready` always carries a value and `.error` an error, so every answer is unchanged. | A Bool `isAwaitingFirstValue`; a graph-only enum. |
| D-15 | `SourcesListContent` loses its `progressHandle` parameter; `SourcesSplitRowProgress`, the inline row, `SourcesSplitRow`, `SourcesSplitLayout` and `SourcesSplitRow.swift` are deleted. `captionBand` becomes one text, `L10n.Sources.listCaption` ("Source · evidence graphs"), leading `PVSpacing.gutterPage`, same font, case and band as today. | GT-1, GT-2; review M-3. | Keeping the split track for the caption. |
| D-16 | `SourcesListRow` layout: `VStack(alignment: .leading, spacing: PVSpacing.space4)`; padding vertical `PVSpacing.space5`, horizontal `PVSpacing.gutterPage` minus `PVSpacing.space2` on the leading edge (so the link's 4 pt pad sits in the gutter and the cover aligns with the caption); bottom `PVDivider()`. Top line: `HStack(spacing: PVSpacing.space5)` of the Source link, `Spacer(minLength: 0)`, `PVListDetailCell(detail: PVLastUpdated.detail(...), width: PVLastUpdated.columnWidth)`, ref (`PVFont.mono(size: PVTypeScale.micro)` `textMuted`, fixed size). Token line: `PVFlowLayout(spacing: PVSpacing.space3, lineSpacing: PVSpacing.space3)` with leading padding `SourcesListRow.tokenIndent` *(new)* = `PVSpacing.space2 + coverSize + PVSpacing.space4` (4 + 44 + 8 = 56, the title's x). `.accessibilityElement(children: .contain)`, label the Source title, identifier `sources.row.<id>`. | GT-1, GT-9 (tab order follows layout order), GT-11; frame 9 metrics on tokens (12 = `space5`, 8 = `space4`, 6 = `space3`, 4 = `space2`); the indent lines up with the title (review X-18). | A menu over the whole row (it would catch token right-clicks); spacing literals. |
| D-17 | `WorkspaceSession+EvidenceGraphs.swift` gains `evidenceGraphsRead(sourceID:) -> QueryRead<[CatalogEvidenceGraph]>` *(new)*: no handle → `.loading`; else the handle's `read` mapped through `evidenceGraphs(sourceID:)`, and a loaded list with no graph for the Source reads `.loading` (every Source has a Default graph from creation, so it is a list not yet refreshed). | One rule for "a Source's graphs", shared by the list, the filter and the page (reviews M-10, M-17). | Each feature deriving it; two readings of the stale case. |
| D-18 | `primary_count` goes: `graphprogress.Progress.PrimaryCount` deleted (`add` keeps `vocabkey.IsPrimaryType` for `FiledCount`; its doc comment becomes "FiledCount is the subjects of a Promote primary kind (person, event, place) with an accepted Identity Claim"); `engine.proto` `EvidenceGraphProgress` gets `reserved 5; reserved "primary_count";`; `api/ffi/handlers/evidence_graph_progress.go` drops the field; `TestListEvidenceGraphProgress` drops the `GetPrimaryCount() != 1` clause and keeps `GetFiledCount() != 0`; `GoStore`, `GenealogyStore` (`EvidenceGraphProgress` and its doc comment) and `FakeStore` drop `primaryCount` — FakeStore deletes its local `primaryTypeIDs` (keeps `Self.primaryTypeKeys`, used elsewhere). `TestPrimaryAndFiledCounts` becomes `TestFiledCounts`. Regeneration needs `protoc`, `protoc-gen-go` and `protoc-gen-swift` (a Mac set up per the repo README). | User, 2026-10-26; no reader after D-5 (review M-5, X-20, X-25). | Swift only; hand-editing `.pb.*` files. |
| D-19 | The Citation composer's private `FlowLayout` moves to `DesignSystem/Components/FlowLayout/PVFlowLayout.swift` *(new)* as `struct PVFlowLayout: Layout` with the same `spacing` / `lineSpacing` initializer and algorithm, a doc comment ("wraps children by ideal width; a caller caps a long child's width"), and a `#Preview`. The two composer call sites use it; the private copy is deleted. `DesignSystem/README.md` gains a row: "FlowLayout \| `Components/FlowLayout/PVFlowLayout.swift` (S12-14: promoted from the Citation composer; wraps children by ideal width, callers cap long children; used by the composer and the Sources list token line)". | Second call site (`add-ui-component`); folder named without `PV` (`docs/design-system-layers.md` "Folder naming"); the kit inventory lists it (review M-16). | `Components/Flow/`; no README row. |
| D-20 | `DesignSystem/Tokens/PVStroke.swift` *(new)*: `enum PVStroke { static let dashPattern: [CGFloat] = [4, 3]; static let dashed = StrokeStyle(lineWidth: 1, dash: dashPattern) }`. `PVCard`, `PVEmptyState`, `OnboardingContributorRow`, `NameValueEditorForm` use `PVStroke.dashed`; `PVChip` uses `isDashed ? PVStroke.dashPattern : []`; `EvidenceCardView` uses `(placed.isCited || selected) ? [] : PVStroke.dashPattern`; blocked tokens use `PVStroke.dashed`. | One design-system fact instead of seven copies (review M-15). | A sixth literal. |

**Token states** (D-11):

| State | Visible | Help | VoiceOver label / hint | Identifier | Focus · action |
| --- | --- | --- | --- | --- | --- |
| Graph, opens | name + compact count | `graphCounts(_, style: .full)` | `openEvidenceGraphLabel(name:)` + ", " + `graphCounts(_, style: .spoken)` | `sources.row.<sourceID>.graph.<graph.ref>` | focusable · opens the graph |
| Graph, blocked (`blockedReason != nil`) | `graphNeedsArtifact(name)`, dashed, faint, no count | `blockedReason` | label `graphNeedsArtifact(name)`; hint `blockedReason` | same as above | focusable · no-op |
| No Artifact | italic `needsArtifact`, dashed, faint | `needsArtifactTooltip` | label `needsArtifact`; hint `needsArtifactTooltip` | `sources.row.<sourceID>.graphBlocked` | focusable · no-op |
| Graph list loading | `PVSkeletonBar(width: 96, height: 6)` in the token frame | — | `L10n.DesignSystem.listLoading` | `sources.row.<sourceID>.graphsLoading` | not focusable (no `Button`) |
| Graph list failed | `PVIcon(.danger, size: 13)` in `PVColor.danger` | `L10n.Errors.message(for:)` | the same message | `sources.row.<sourceID>.graphsError` | focusable · no-op |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcesListRow`, `SourcesLinkLabel`, `SourcesLinkButtonStyle`, the link highlight `EnvironmentKey` *(new)* | view / style / key | `SourcesSplitRow`, `SourcesSplitRowProgress`, `SourcesSplitZoneButtonStyle`, `PVButton(.link)` | The board replaces the split row; one call site (D-2, D-15); `.link` underlines at rest; the style composes `PVHoverEffect` and `pvFocusRing` (D-10). |
| `SourcesGraphToken`, `SourcesGraphTokenButtonStyle` *(new)* | view / style | `PVChip`, `SourcesSplitRow.graphZone` | A link to a place, not a selectable pill; composes `PVHoverEffect` and `pvFocusRing` (D-11). |
| `SourcesGraphCountLine` *(new)* | view | `SourcesSplitRow.graphCountLine` (private) | Lifted for the Source page card, switching on `SourcesGraphCounts`. |
| `SourcesGraphCounts`, `L10n.Sources.graphCounts(_:style:)` *(new)* | enum / formatter | `graphCountLine` branches, `graphZoneAccessibility` | One state and one formatter (D-3). |
| `QueryRead`, `QueryHandle.read` *(new)* | enum / property | `isSettled`; `ConclusionListPresentation` fold | On the type that owns the state (D-14). |
| `evidenceGraphsRead(sourceID:)`, `evidenceGraphProgressKey`, `evidenceGraphProgressHandle`, `evidenceGraphProgress(graphID:)`, `evidenceGraphCountsLoading` *(new)* | session reads | `evidenceGraphs(sourceID:)`, `evidenceGraphsHandle`; `SourcesModel` progress reads; `SourceGraphsSection.rows` / `loadError` | Lifted to the session (D-13, D-17). |
| `SourcesModel.statuses(for:)` *(new)* | func | `visibleSources` | D-7. |
| `SourceGraphsSection.progress(for:)`, `countsLoading` *(new)* | forwarders | `SourceGraphsSection.rows` | D-8. |
| `PVFlowLayout` *(new, promoted)* | kit layout | composer `FlowLayout` | D-19. |
| `PVStroke` *(new)* | kit token | six `[4, 3]` literals | D-20. |
| L10n keys `statusPromoted`, `listCaption`, `graphNeedsArtifact` *(new)* | L10n | `statusFullyPromoted`, `columnSource` / `columnEvidenceGraph`, `needsArtifact` | Replace or join them (§6 Strings). The token's open label reuses `openEvidenceGraphLabel(name:)` (review M-13). |

Searched: `FlowLayout`, `: Layout`, `PVChip`, `textLinkHover`, `PVButtonVariant.link`, `PVHoverEffect`, `pvFocusRing`, `dash: [4, 3]`, `pvRightClickMenu`, `PVListDetailCell`, `graphCountLine`, `graphZoneAccessibility`, `openEvidenceGraphLabel`, `evidenceGraphProgress`, `progressHandle`, `entryGraph`, `isSettled`, `status == .error`, `primaryCount` / `primary_count` / `PrimaryCount` (all callers listed in §4).

**Builds on.** `evidenceGraphsBySource` (S12-13), `SourcesListNavigation.graphLocation(for:graph:)` / `graphBlockedReason(for:)` (S12-11), `PVHoverEffect`, `pvFocusRing`, `PVLastUpdated` / `PVListDetailCell`, `pvRightClickMenu`, the `PlaceRegistry` query keys ([`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md)), `WorkspaceSession.apply`'s per-graph progress patch, the proto `reserved` precedent.
**Deletes.** `SourcesSplitRow.swift` (`SourcesSplitLayout`, `SourcesSplitRow`, `SourcesSplitZoneButtonStyle`, `graphCountLine`); `SourcesSplitRowProgress` and the inline row; the interim helpers and their three tests and helper (D-9); `SourcesModel.evidenceGraphProgressKey(for:)` and `progressHandle`; the composer's private `FlowLayout`; six dash literals; the Conclusion presentations' raw-field inits; `PrimaryCount` / `primary_count` / `primaryCount` and FakeStore's `primaryTypeIDs` (D-18); L10n `statusFullyPromoted`, `columnSource` (Sources), `columnEvidenceGraph`, `openGraph`, `graphZoneAccessibility` and their catalog entries.
**Next adds.** S12-15's graph hit opens the location a token opens and speaks `openEvidenceGraphLabel`. Any new place reads `handle.read`; `SourceTypesModel` and `EvidenceGraphModel.canCite` still fold by hand (logged). A status mark on a token reads `EvidenceStatus(_:)`; a reader outside Sources moves `EvidenceStatus` out of `SourcesModel` first.

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
| Row | snowflake (new) | `Features/Sources/SourcesListRow.swift` *(new)* | D-2, D-12, D-16. |
| Source link | snowflake (new) | `Features/Sources/SourcesListRow.swift` *(new)* | D-10. |
| Graph token | snowflake (new) | `Features/Sources/SourcesGraphToken.swift` *(new)* | D-11, token states table. |
| Count state | snowflake (new) | `Features/Sources/SourcesGraphCounts.swift` *(new)* | D-3. |
| Token line wrap | `PVFlowLayout` (promoted) | `DesignSystem/Components/FlowLayout/PVFlowLayout.swift` *(new)* | D-19. |
| Dashed stroke | `PVStroke` (new token) | `DesignSystem/Tokens/PVStroke.swift` *(new)* | D-20. |
| Full count line | snowflake (lifted) | `Features/Sources/SourcesGraphCountLine.swift` *(new)* | `progress`, `countsLoading` → `SourcesGraphCounts`; `.loading` hairline as today; `.notStarted` "0 subjects · not started" as today; `.counts` `graphCountLine`; label `graphCounts(_, style: .spoken)`, children ignored. |
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
9. *Metrics.* As D-10, D-11 and D-16.
11. *Navigation.* Source link → Source page; token → that graph's canvas; blocked token inert; nothing expands.
12. *Source page.* Census cards: Default "0 subjects · not started"; Jones "12 subjects · 40 observations" then its description; Smith "7 subjects · 19 observations".
13. *Building blocks.* As §6 macOS.

### Strings

L10n keys *(new)*: `L10n.Sources.statusPromoted` = `sources.list.status.promoted` "Promoted" ("Status filter option: graphs with at least one promoted subject") — STEP-2; `L10n.Sources.graphNeedsArtifact(_:)` = `sources.list.graphNeedsArtifact` "%@ needs an artifact" ("Blocked Sources list graph token; argument is the graph name") — STEP-6; `L10n.Sources.listCaption` = `sources.list.caption` "Source · evidence graphs" ("Caption band above the Sources list") — STEP-7. New func, no key: `graphCounts(_:style:)` — STEP-2. Changed: `sources.page.graph.openLabel` default "Open graph %@" (was "Open %@"), comment "VoiceOver label for opening an Evidence graph from the Source page or a Sources list token" — STEP-6; `sources.list.status.inProgress` comment "at least one subject and none promoted" — STEP-2. Removed: `statusFullyPromoted` (STEP-2); `openGraph`, `graphZoneAccessibility` (STEP-7); `sources.list.columnSource`, `columnEvidenceGraph` (STEP-7). Kept: `graphSubjectCount`, `graphObservationCount`, `graphCountLine`, `graphNotStarted`, `graphCountsLoading`, `needsArtifact`, `needsArtifactTooltip`, `graphNeedsArtifactTooltip`, `openSourcePage`, `copySourceEntry`. The compact count is digits and a middle dot built in the formatter (`add-localized-string`: verbatim for numbers and glyphs).

## 7. Steps

Steps 3 (proto) and every Swift check need a Mac with Xcode, `protoc`, `protoc-gen-go` and `protoc-gen-swift`. If a tool is missing, stop and report; do not hand-edit generated files.

**STEP-1 — One read of a Source's graphs and of graph progress.**
Files: `QueryHandle.swift`, `WorkspaceSession+EvidenceGraphs.swift`, `WorkspaceSession.swift`, `SourcesModel.swift`, `SourceGraphsSection.swift`, `Features/Conclusions/ConclusionListPage.swift`, `Features/Conclusions/ConclusionDetailPage.swift`, `macos/ProvenenciaTests/PersonsListTests.swift`, `SourcesListView.swift` (key call sites only), `SourcesModelTests.swift`, `WorkspaceSessionTests.swift`, `docs/macos-client-patterns.md`. Skill: `add-workspace-place`.
Do: D-13, D-14, D-17; `SourceGraphsSection.rows` / `loadError` read D-17; `warmListQueriesLoadsGraphProgress` reads `session.evidenceGraphProgress(graphID:)`. No behavior change.
Check: T-9, T-10 pass; `SourcesModelTests`, `SourcePageModelTests` and `PersonsListTests` pass; `rg -n 'SourcesModel\.evidenceGraphProgressKey|evidenceGraphProgress\(project: projectKey\)' macos/App` lists only `WorkspaceSession+EvidenceGraphs.swift`.

**STEP-2 — One count state; status per graph.**
Files: `SourcesGraphCounts.swift` *(new)*, `SourcesModel.swift`, `SourcesListSettings.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`, `SourcesListSettingsTests.swift`, `project.pbxproj`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-3, D-5, D-6, D-7; strings per §6 Strings for STEP-2. The test helper becomes `progress(_ id:, subjects:, filed:)`. `filtersCombineAndClear` seeds `filed: 0` where it expects In progress (s1, s2, s5) and keeps `filed: 1` where it expects Promoted; `settingsRestoreOnNewModel` and `statusCountsFailureSurfacesAnErrorAndUnfilteredCount` use `.promoted`.
Check: T-1, T-2, T-3, T-11 pass with every other `SourcesModelTests` test.

**STEP-3 — Drop `primary_count`.**
Files: `core/database/graphprogress/graphprogress.go`, `graphprogress_test.go`, `api/proto/engine.proto`, generated `engine.pb.go` / `engine.pb.swift`, `api/ffi/handlers/evidence_graph_progress.go`, `evidence_graph_progress_test.go`, `GoStore.swift`, `FakeStore.swift`, `GenealogyStore.swift`, `SourcesModelTests.swift`, `WorkspaceSessionTests.swift`. Skill: `add-ffi-handler`.
Do: D-18.
Check: `./scripts/generate-proto.sh`; `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphprogress/... ./api/ffi/handlers/...` passes; `rg -n 'PrimaryCount|primaryCount|primaryTypeIDs|primary_count = ' core api macos/App macos/ProvenenciaTests --glob '!*.pb.*'` returns nothing (the `reserved "primary_count"` line does not match).

**STEP-4 — Promote the flow layout; one dashed stroke.**
Files: `DesignSystem/Components/FlowLayout/PVFlowLayout.swift` *(new)*, `DesignSystem/Tokens/PVStroke.swift` *(new)*, `CitationComposerAuthoringBox.swift`, `PVCard.swift`, `PVEmptyState.swift`, `PVChip.swift`, `OnboardingContributorRow.swift`, `NameValueEditorForm.swift`, `EvidenceCardView.swift`, `DesignSystem/README.md`, `project.pbxproj` (a new `FlowLayout` group and two file entries). Skill: `add-ui-component`.
Do: D-19, D-20.
Check: the app builds; `rg -n '\[4, 3\]|private struct FlowLayout' macos/App` lists only `PVStroke.swift`; the composer's word and term rows and a dashed card look as before.

**STEP-5 — The full count line; counts on the Source page.**
Files: `SourcesGraphCountLine.swift` *(new)*, `SourcesSplitRow.swift` (draws it until STEP-7), `PlaceRegistry.swift`, `PlaceRegistryTests.swift`, `SourceGraphsSection.swift`, `SourcePageGraphsView.swift`, `SourcePageModelTests.swift`, `project.pbxproj`. Skill: `add-workspace-place`.
Do: D-8; the card line (§6 macOS). `PlaceRegistryTests` `resolvesSourceDetail` and the place table's `.sourceDetail` row gain `.evidenceGraphProgress(project: project)` at the end.
Check: T-7, T-8 and `PlaceRegistryTests` pass.

**STEP-6 — The graph token.**
Files: `SourcesGraphToken.swift` *(new)*, `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`, `project.pbxproj`. Skill: `add-ui-component`, `add-localized-string`.
Do: D-11 and the token states table (not yet drawn by the list); strings per §6 Strings for STEP-6.
Check: the app builds; T-12 passes.

**STEP-7 — The two-line row.**
Files: `SourcesListRow.swift` *(new)*, `SourcesListView.swift`, `SourcesSplitRow.swift` (deleted), `SourcesModel.swift`, `SourcesListNavigation.swift`, `SourcesListNavigationTests.swift`, `SourcesModelTests.swift`, `L10n.swift`, `Localizable.xcstrings`, `project.pbxproj`. Skill: `add-ui-component`.
Do: D-1, D-2, D-9, D-10, D-12, D-15, D-16; strings per §6 Strings for STEP-7.
Check: T-5, T-6, T-13 pass; `rg -n 'SourcesSplit|SourcesListNavigation\.entryGraph|model\.entryGraph|graphListLoaded|graphLocation\(for: source\)|graphsBySource\(|graphZoneAccessibility|statusFullyPromoted|columnEvidenceGraph|split-row' macos/App/Features macos/ProvenenciaTests` returns nothing (`SourcePageModelTests.entryGraphLocation` is not matched and stays); manual check steps 1–5.

**STEP-8 — Docs.**
Files: `../completed.md`, `../deployment-plan.md` (map row; working rule "A graph is a place" drops its interim first-graph sentence), `../design/README.md` (S12-D5 "built by S12-14 (#N)"), this plan (status, Landed note).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-tokens.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

Swift handle waits: warm first (`model.warmListQueries()` or `session.apply(location:)`), then `let handle = try #require(session.evidenceGraphProgressHandle)` and `await waitForQuery(handle)`. Writes use `session.apply(CatalogWriteEffects(evidenceGraphs: ["<graph id>"]))`, as `WorkspaceSessionTests` does.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesModelTests.swift` `evidenceStatusFromCounts` (rewritten) | nil → empty; subjects 0 → empty; subjects 5, filed 0 → inProgress; subjects 5, filed 1 → promoted; subjects 2, filed 0 → inProgress | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-5 |
| T-2 | Swift | `SourcesModelTests.swift` `statusMatchesAnyGraphOfTheSource` *(new)* | One Source, three graphs (empty, in progress, promoted) → in `visibleSources` under each status; a second Source with one empty graph only under Empty; `statuses(for:)` is `[]` while the graph list is loading | AC-2.5, D-7 |
| T-3 | Swift | `SourcesListSettingsTests.swift` `storedFiledStatusRestoresPromoted` *(new)* | `{"status":"filed"}` → `.promoted`; `{"status":"promoted"}` → `.promoted`; `{"status":"bogus"}` → nil, other fields kept; `EvidenceStatus.promoted.label` resolves to "Promoted" | AC-2.6, AC-2.7, D-6 |
| T-4 | Go | `core/database/graphprogress/graphprogress_test.go` `TestGraphProgress`, `TestFiledCounts` (renamed); `api/ffi/handlers/evidence_graph_progress_test.go` `TestListEvidenceGraphProgress` | Today's cases without `PrimaryCount`: a filed primary Subject counts in `FiledCount`; a bridge Subject does not; the handler row has `FiledCount` 0 | AC-4.1, D-18 |
| T-5 | Swift | `SourcesListNavigationTests.swift` | The three interim tests and `graphsBySource` deleted; `graphLocationOpensTheGivenGraph` and `pageLocationUsesPageSurface` pass unchanged | D-9 |
| T-6 | Swift | `SourcesModelTests.swift` `listReadsEachGraphsProgress` *(new)* | Three graphs on `s1`, progress on `g-jones` only, after warm → `session.evidenceGraphsRead(sourceID: "s1")` `.loaded` in name order; `session.evidenceGraphProgress(graphID: "g-jones")` the seeded counts, nil for the others; store counts for `g-jones` change and `session.apply(CatalogWriteEffects(evidenceGraphs: ["g-jones"]))` → only `g-jones` changes | AC-1.1, AC-1.2, AC-1.9, D-12 |
| T-7 | Swift | `SourcePageModelTests.swift` `graphRowsReadTheirProgress` *(new)* | Two graphs, progress seeded for one; `session.apply(location: .sourcePage(...))`, then wait on the progress handle → `model.graphs.progress(for:)` returns it for that graph, nil for the other; `countsLoading` false | AC-3.1, AC-3.2, D-8 |
| T-8 | Swift | `SourcePageModelTests.swift` `aGraphWriteMovesOnlyItsRow` *(new)* | T-7, then the store's count changes and `session.apply(CatalogWriteEffects(evidenceGraphs: [id]))` → its progress updates; the other stays nil | AC-3.3 |
| T-9 | Swift | `WorkspaceSessionTests.swift` `evidenceGraphsReadFollowsTheHandle` *(new)* | No handle → `.loading`; loading → `.loading`; failed without value → `.failed`; loaded → `.loaded` with the Source's graphs; loaded without the Source → `.loading` | AC-1.6, D-17 |
| T-10 | Swift | `WorkspaceSessionTests.swift` `queryReadAndProgressReads` *(new)* | `QueryHandle.read`: fresh → `.loading`; after `applySuccess` → `.loaded`; after `applyFailure(clearValue: true)` → `.failed`; after `applyFailure(clearValue: false)` with a value → `.loaded`. Before warm: `evidenceGraphCountsLoading` true, `evidenceGraphProgress(graphID:)` nil; after both load: false and the seeded value | AC-1.7, D-13, D-14 |
| T-11 | Swift | `SourcesModelTests.swift` `graphProgressCopyMatchesBoard` (rewritten) | `SourcesGraphCounts(progress: nil, countsLoading: true)` == `.loading`; zeros → `.notStarted`; 12 / 40 → `.counts`. `graphCounts`: compact "12 · 40" / "not started" / "counts loading"; full "12 subjects · 40 observations" / "0 subjects · not started"; spoken "12 subjects, 40 observations" / "0 subjects, not started" | AC-1.2, AC-3.1, D-3 |
| T-12 | Swift | `SourcesModelTests.swift` `graphTokenCopyMatchesBoard` *(new)* | `openEvidenceGraphLabel(name: "Jones household")` == "Open graph Jones household"; `graphNeedsArtifact("Unlinked")` == "Unlinked needs an artifact" | AC-1.4, D-11 |
| T-13 | Swift | `SourcesModelTests.swift` `listCaptionMatchesBoard` *(new)* | `L10n.string(L10n.Sources.listCaption)` == "Source · evidence graphs" | D-15 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests`, `SourcesListNavigationTests`, `SourcesListSettingsTests`, `PlaceRegistryTests`, `SourcePageModelTests`, `WorkspaceSessionTests`, Conclusions tests (unchanged code, same handle), `TestGraphProgress`, `TestListEvidenceGraphProgress`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; rg -n 'fullyPromoted|Fully promoted|columnEvidenceGraph' macos returns nothing
- [ ] UI matches the brief inventory; kit additions are PVFlowLayout (promoted, README row) and PVStroke only
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; working rule "A graph is a place" drops the interim first-graph sentence; completed.md: one short paragraph of what shipped
- [ ] design/README.md: S12-D5 "built by S12-14 (#N)" (the brief is already archived)
- [ ] docs/macos-client-patterns.md: the QueryHandle row names `read` (D-14)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source and three graphs: Default (no Subjects), Jones household (Subjects, not promoted), Smith household (promoted once, one Subject filed of several). The census row shows "Default *not started*", "Jones household 12 · 40", "Smith household 7 · 19" under its title, the first token's left edge under the title's; hovering Jones's token shows "12 subjects · 40 observations"; a Source with no Artifact shows "Needs an artifact" (AC-1.1, AC-1.2, AC-1.5). Relaunch and watch the list load: a skeleton token first, then tokens with a short rule for counts, then counts (AC-1.6, AC-1.7).
2. Hover the census title: link color and underline, cover outline darker, no fill. Click it: the Source page opens (AC-1.8). Back. Click Jones household's token: its canvas opens on that graph (AC-1.3).
3. On the census Source page, add a new page Artifact, add a graph "Unlinked" linked only to it, then delete that Artifact: back in the list, Unlinked's token is dashed, reads "Unlinked needs an artifact", and does nothing on click (AC-1.4).
4. Add an Observation on Jones's canvas; Back to the list. Only Jones's token moved (AC-1.9).
5. Status filter: Empty, In progress (Jones, AC-2.2) and Promoted (Smith, AC-2.3) each show the census with all its tokens. The last option reads Promoted. Narrow the window to its minimum with a twelve-graph Source: the tokens wrap and a long name truncates (AC-1.10). Tab from the census link: a focus ring shows on the link, then on each token in name order, the blocked one included.
6. Open the census Source page: each card shows its count line under the ref (AC-3.1). Quit and relaunch with that page restored: the count lines load (AC-3.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Source with many graphs makes a tall row | Tokens wrap and truncate at 300; the list is lazy per row. |
| The project-wide progress read on every Source page open | One cached map shared with the list; patches are per graph. |
| A graph with only value Subjects can never become Promoted | Accepted (D-5): Promote files only primary kinds; it reads In progress, which is true. |
| `PVFlowLayout` measures with `.unspecified`, so a long token name proposes its ideal width | The token caps the name with `.frame(maxWidth: 300)`; placement proposes the measured size, so the name truncates. |
| `QueryHandle.read` changes nothing for existing callers | The Conclusion presentations' answers are unchanged (D-14; `PersonsListTests`). |

**Stop and report if:**

- `protoc`, `protoc-gen-go`, `protoc-gen-swift` or Xcode is missing (STEP-3 and every Swift check).
- `WorkspaceSession.refreshEvidenceGraphProgress` no longer patches by graph id, or `apply` no longer calls it for every graph in `effects.evidenceGraphs`.
- `ListScrollAnchorModifier` needs ids other than Source ids (D-2 cannot hold).
- A token inside `PVFlowLayout` does not truncate at 300 at the minimum width (frame 5): report with a screenshot.
- `pvRightClickMenu` on the Source link swallows the link's primary click or its VoiceOver action.
- The link highlight environment value does not reach the label from `SourcesLinkButtonStyle` (D-10): report what the hover shows.
- Observation does not repaint `SourcesListRow` when the progress map changes without a `@Bindable` handle (D-2).
- Anything other than the files in §4 reads `primary_count` / `primaryCount` (D-18).

## 12. Open questions

None.

## 13. PR

**Title:** S12-14 — Graph counts in the Sources list and on the Source page
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-tokens.md`.
