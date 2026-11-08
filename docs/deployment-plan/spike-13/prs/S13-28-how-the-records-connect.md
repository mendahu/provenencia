# S13-28 — How the records connect

**Status:** Blocked — on board S13-D10
**Spike / slice:** Spike 13 · Slice 7 — How a person's records connect
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D10](../design/S13-D10-how-the-records-connect.md), board not drawn yet
**Board:** — (to be recorded when the board is agreed: the Claude Design project and turn, as S13-D6's Board line)
**Depends on:** S13-16 (landed [#722](https://github.com/mendahu/provenencia/pull/722)), S13-17 (landed [#729](https://github.com/mendahu/provenencia/pull/729)), S13-27 ([#743](https://github.com/mendahu/provenencia/pull/743), same branch), S13-D10
**Blocks:** —
**Branch:** `s13-27-plan` (the user's choice, 2026-11-07: S13-27's PR carries this plan; the PR that builds it gets its own branch, `s13-28-how-the-records-connect`, once the board is agreed)
**Runs on:** macOS: every step needs Xcode
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [`evidence-graph-comparison.md`](../evidence-graph-comparison.md) Requirement 2; working rules "The view is a circular link diagram", "Scoring and the view are two concerns", "Sections are places", "A full-width section keeps its states in the band"; dogfood bar 10; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide)
**Planned:** 2026-11-07 against `s13-27-plan` @ `cc03d7f9` (S13-27's code is on this branch, not yet on `main`)
**Review:** [`S13-28-how-the-records-connect.review.md`](S13-28-how-the-records-connect.review.md)

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

The Person page's Sources section shows, beside the list of Sources and pinned in view while the list scrolls, how the person's records connect: each Evidence graph that holds a record of the person sits on a circle labelled with its graph ref, and between two graphs there is a thin or thick arc for each other person both name. Hovering a graph on the circle highlights its line in the list, and hovering a line highlights the graph on the circle, so the researcher knows which record to go and look at. It is dogfood bar 10 and Requirement 2 of [`evidence-graph-comparison.md`](../evidence-graph-comparison.md).

**Why now.** S13-27's read is built; this is its only reader and the last PR of slice 7.

**What stays.** The Source rows' content, columns, folds, Remove and confirm (S13-16, S13-17). The read (S13-27). The other sections. The diagram shows no data about a line (working rule).

## 2. User stories

Sample data for every AC: S13-D10 § 5 (Johann Weber, seven graphs, eight arcs, graphs 3 and 7 apart).

**US-1.** As a researcher on a Person's Sources section, I see the person's records on a circle with an arc per shared person, so that I see which records agree, which stand apart, and which groups do not connect.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Johann's Sources section, Regular and Wide | Seven nodes in the read's order clockwise from the top, labelled EVG-2K7PQ … EVG-9CN2H; eight arcs: 1–2 thick and thin, 1–4 two thick, 2–4 thin, 1–5 thick, 5–6 two thin; nodes 3 and 7 with no arc |
| AC-1.2 | Two arcs between the same two graphs | Two distinct curves, side by side |
| AC-1.3 | A graph with no ref | Labelled with its name |
| AC-1.4 | A person with one graph | The one node and the one-record caption (RL-8) |
| AC-1.5 | Hover or click an arc | Nothing happens; no tooltip |

**US-2.** As a researcher, hovering links the circle and the list, and a node opens its record.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Pointer on node 1 | Node 1 and its arcs highlighted; nodes 2, 4, 5 lightly; the rest receded; the Household 214 line in the list highlighted |
| AC-2.2 | Pointer on the Register 3 line (anywhere on the line, its Unpromote button included) | Node 6 and its arcs highlighted |
| AC-2.3 | Pointer leaves | Everything at rest, unless a node or graph link has keyboard focus (then that one stays highlighted) |
| AC-2.4 | Click node 2 | The Evidence graph EVG-5MD8R opens with J. Weber (CPR-2M8XF) selected, as the Household 388 line's link does; Back returns to Sources |
| AC-2.5 | A graph whose line is folded under "Show N more" | Its node highlights; the list does not unfold |

**US-3.** As a researcher with many records, the diagram stays in view while I scroll the list.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | 50 graphs; scroll the list half way | The diagram is held 24pt below the top of the visible area |
| AC-3.2 | Scroll to the end of the section | The diagram scrolls away with the last rows; it never overlaps content below the section |
| AC-3.3 | A window shorter than the diagram | The diagram fits the visible height (RL-13's answer) |
| AC-3.5 | The window at its 700pt minimum | The board's RL-W5 arrangement; nothing overlaps, no row is cut to nothing |
| AC-3.4 | Resize across 1600 with a node focused | The node keeps focus and highlight; the scroll position holds (RL-W3) |

**US-4.** As a keyboard or VoiceOver user, I reach and open each node.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Tab past the list | Each node in circle order with the kit focus ring; Return opens it; focus highlights as hover does |
| AC-4.2 | VoiceOver on node 3 | "EVG-8TZ4N, Volume 3, Hamburger Adressbuch 1861, shares no one" (RL-12's answer) |

**US-5.** As the app, the diagram loads, fails and refreshes on its own read.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | The record-links read is loading, the claims read loaded | The rows drawn, the diagram column shows the spinner |
| AC-5.2 | The record-links read fails | The rows drawn; the diagram column shows the danger Callout "The connections could not be read" with the error |
| AC-5.3 | The claims read is loading, failed or empty | The section host's state as today, full width; no diagram column |
| AC-5.4 | Either read reloads | The switcher row shows "Refreshing"; both columns stay drawn |
| AC-5.5 | Opening the Sources section | The place warms `.personClaims` and `.personRecordLinks`; a `.sourceTypesList` reload does not show "Refreshing" |

## 3. Scope

| In | Covered by |
| --- | --- |
| RL-1 – RL-12: the diagram, hover both ways, open, a11y, dark | STEP-4 – STEP-8 |
| RL-13: pinned while the rows scroll | STEP-2, STEP-3, STEP-8 |
| RL-W1 – RL-W5: side by side at both widths; full-width mode at Wide; states at Wide; narrow panes (board) | STEP-1, STEP-3, STEP-8 |
| RL-10, RL-14: the diagram's own states; none when the rows show a state | STEP-8 |
| A section's companion read: warmed, refreshing | STEP-1 |
| One reader of the page's visible rectangle (the Timeline moved onto it) | STEP-2 |
| One side-column arrangement: widths in `PVPageColumnsMetrics`, a pinned side column view | STEP-3 |
| The graph line's highlight | STEP-8 |
| Strings | STEP-4 |
| Docs and skills this PR makes false | STEP-9 |

| Out | Home |
| --- | --- |
| Line data: the shared person, facts, conflicts | not planned in Spike 13: [`evidence-graph-comparison.md` § Open](../evidence-graph-comparison.md#open) |
| Event and Place diagrams | [`evidence-graph-comparison.md` § Later](../evidence-graph-comparison.md#later) |
| Pan, zoom, drag; the `GraphCanvas` host | not planned: a fixed circle (S13-D10 non-goals) |
| A kit chord-diagram component; the pinned side column in the kit | not planned until Event or Place is the second caller |
| A pointing-hand cursor on nodes or links | not planned: no kit link shows one (`PVBreadcrumbs` is the only cursor push) |
| Giving S13-27's graph row the person's Subject id, so the diagram needs no claims join | not planned: S13-27 is final for this version; the join's race is handled (D-4) |
| Changes to the record-links read | S13-27 |

## 4. Current state

Verified on `s13-27-plan` @ `cc03d7f9`.

| Area | Where | Today |
| --- | --- | --- |
| Sources section | `macos/App/Features/Conclusions/PersonSourcesSection.swift` `PersonSourcesSection` (one `ConclusionPageSectionHost` over `.personClaims`), `PersonSourcesList` (`@FocusState focus`, `pendingFocus`, `returnFocus`, the header and rows), `PersonSourcesRowView`, `PersonSourcesLineView` (graph link, refs, Unpromote), `PersonSourcesFocus`, `PersonSourcesColumns` (Type 176, Date 120, Ref 104, `spacing` 14, `lineIndent` 20) | One column of rows; focus owned by `PersonSourcesList` |
| Rows model | `macos/App/Features/Conclusions/PersonSourcesContent.swift` `PersonSourcesContent.init` (lines 73 – 107 build every `Line.location` through `SourcesListNavigation.canvasLocation` before `ConclusionFold.shown` folds them) | No per-graph location map |
| Preview data | `macos/App/Features/Conclusions/PersonSourcesPreviewData.swift` `regular`, `folded`, `manySources`, `types` | Claim groups only |
| Section metadata | `macos/App/Features/Workspace/WorkspaceLocation.swift` `ConclusionPageSection.queryKey`, `supportingQueryKeys` (doc: a catalog-wide vocabulary, warm-only), `fillsWidth` | Sources reads `.personClaims`, warms `.sourceTypesList`, does not fill the width |
| Refreshing | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `sectionRefreshing` (line 494) | Follows only `queryKey` |
| Presentation | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPresentation` (lines 6 – 30) | Maps a `QueryRead` to `.loading`, `.gone`, `.failed`, `.content(refreshing:)` |
| Warming | `macos/App/Features/Workspace/Session/PlaceRegistry.swift` lines 199 – 200 | Warms `queryKey` and `supportingQueryKeys` |
| Section host | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` | Spinner, danger Callout (`.gone` → `conclusionDetailsNotFound`), EmptyState; states keep the page band; draws `content` otherwise |
| Page columns | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` `PVPageColumnsMetrics.columnWidths`, `scrollSpace`, `pvPageViewportSize`; vertical padding `PVSpacing.space8` (24) at line 139 | Widths for primary plus side column; the scroll space and the viewport size |
| Viewport reader | `macos/App/Features/Conclusions/PersonTimelineSection.swift` `PersonTimelineTethers` (lines 464 – 500) | The one reader of the visible rectangle over a section, hand-built |
| Node parts | `DesignSystem/Components/HoverEffect/PVHoverEffect.swift`, `DesignSystem/Tokens/PVElevation.swift` `pvFocusRing`, `DesignSystem/Components/InkLink/PVInkLink.swift` `PVInkLinkStyle`, `DesignSystem/Components/ReturnActivation/PVReturnActivation.swift` `pvActivatesOnReturn` | Pressed state, focus ring, Return |
| Tokens | `DesignSystem/Tokens/PVColor.swift` `surfaceSelected`, `accentLine` | Selected row background; accent line |
| Record links read | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonRecordLinks` and parts; `CatalogQueryKey.personRecordLinks` | S13-27: graphs in the Sources order, pairs by graph id; no reader |
| Strings | `macos/App/Platform/L10n.swift` `L10n.Conclusions` | `sources*` keys |
| Guides | `docs/macos-client-patterns.md` line 104 and § 9 "Patterns to reach for"; `.agents/skills/add-workspace-place/SKILL.md` line 90; `docs/ideas/wide-layout.md` line 10 | Describe supporting keys as warm-only, Person's sections as primary-only; no side-column pattern |
| Tests | `macos/ProvenenciaTests/PersonDetailTests.swift` `sectionQueryKeys`; `PersonSourcesContentTests.swift`; `PlaceRegistryTests.swift`; `PVPageColumnsTests.swift` | — |

**Precedent.** S13-12's Family diagram ([#676](https://github.com/mendahu/provenencia/pull/676)): `PersonFamilyLayout`, a pure layout of a section tested without a view, drawn by `PersonFamilySection.diagram` as lines on a `Path` with kit buttons placed by `offset` and micro-caps captions. Copy that split for the diagram. S13-13's Timeline ([#687](https://github.com/mendahu/provenencia/pull/687)) for full-width mode and the viewport reading.

## 5. Decisions

Values marked **(board)** are filled from the agreed board in STEP-0, before this plan is `Planned`.

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Placement (user, 2026-11-07): the diagram beside the rows at both widths, pinned while they scroll. `ConclusionPageSection.fillsWidth` is true for `.sources` | The user's answer; full-width mode is how Timeline and Gallery use Wide | A page secondary column — stacks below at Regular (rule 4) |
| D-2 | A companion read: `ConclusionPageSection.companionQueryKey(project:entityId:) -> CatalogQueryKey?` *(new)*, `.personRecordLinks` for `.sources`, nil otherwise; doc "a second per-entity read the section draws in its own region". `PlaceRegistry`'s `.personDetail` spec warms it right after `queryKey` (before `supportingQueryKeys`); `sectionRefreshing` follows `queryKey` and `companionQueryKey`. `supportingQueryKeys` is unchanged | Each list keeps one meaning: supporting keys stay warm-only vocabulary; a vocabulary reload never shows Refreshing (review M-1) | Widening `supportingQueryKeys` — bends its doc and `add-workspace-place` line 90 |
| D-3 | `PVPageColumnsMetrics.sideColumnWidths(available:widthClass:sideRegular:sideMax:) -> (main: CGFloat, side: CGFloat, group: CGFloat)` *(new)*, pure, beside `columnWidths`: Regular: side = `sideRegular`, main = available − `space9` − side; Wide: main = min(available − `space9` − `sideRegular`, `widthContentMax` − `space9` − `sideRegular`), side = min(available − `space9` − main, `sideMax`); group = main + `space9` + side, centred in `available` by the caller. `ConclusionPinnedSideColumn` *(new)* in `Features/Conclusions/ConclusionPinnedSideColumn.swift`: measures its own width with `onGeometryChange` (width only; until measured, the main column takes the whole width and the side is not drawn), lays out `main` and `side` in an `HStack(alignment: .top, spacing: PVSpacing.space9)` centred, and pins `side` (D-7). `sideRegular` = **(board, ≤ 364pt)**, `sideMax` = **(board)**. Below a pane width of **(board, RL-W5)** (a window down to `PVSpacing.widthWindowMin`, 700pt) the arrangement follows the board's RL-W5 answer, tested in T-3; the graph lines in a narrowed citation column follow the board's RL-W2 line answer, built in `PersonSourcesLineView` (STEP-8) | One home for "a list with a side column at both widths, pinned", reusing the page's width tokens and rule-6 centring; Event and Place reuse it (review M-2). The cap 364 keeps the citation column at least 200pt at Regular (1038 − 32 − 364 − 442) | `PersonSourcesColumns.split` — a second width calculator, and "Columns" with two meanings. A kit `PVPageColumns` option — one caller today |
| D-4 | `PersonRecordLinksContent` *(new, Equatable)* in `Features/Conclusions/PersonRecordLinksContent.swift` *(new)*, the Person adapter: from `CatalogPersonRecordLinks` and `PersonSourcesContent.graphLocations` *(new: `[String: WorkspaceLocation]`, the first line's location per graph id, taken in `init`'s loop before the fold)*. `nodes: [RecordLinksNode]` (`id` graph id, `label` graph ref else graph name, `help` via `L10n.Conclusions.recordLinksNodeHelp(graph:source:)`, `location` optional, `connected`, `accessibilityLabel`); `arcs: [RecordLinksArc]` (`a`, `b` node indexes, `strong`, `slot`, `slots`); `neighbors(of: Int) -> Set<Int>`. `strong` is `.strong`; slots follow the read's line order; `connected` counts distinct other graphs. A duplicate graph id keeps the first; a pair naming a graph not in the list is dropped; a node with no location (the reads raced) is drawn and not clickable | One builder of claim locations (review M-4); the highlight's neighbour set lives in the model, tested | Rebuilding locations from the groups — a second builder |
| D-5 | `RecordLinksLayout` *(new)* in `Features/Conclusions/RecordLinksLayout.swift`, pure, takes node count, arcs and a size: node points on a circle clockwise from the top (0 nodes → none; 1 → centre; 2 → top and bottom); circle diameter = size − 2 × label room **(board)**; label anchors outside at **(board)** offset; each arc a quadratic bowed toward the centre (**board** for the shape), fanned by `slot`/`slots` within a spread of **(board)** points, the spread shared evenly however many arcs. `RecordLinksDiagram` *(new)* in `Features/Conclusions/RecordLinksDiagram.swift` draws arcs on a `Canvas` with hit testing off, and each node as a `Button` placed by `offset` (the Family precedent) with a private `ButtonStyle` combining `PVHoverEffect` and `pvFocusRing` (the `PVInkLinkStyle` shape), `pvActivatesOnReturn`, and `onHover` reporting the node id up. Its inputs are Person-free: `nodes`, `arcs`, `highlightedGraphID: String?`, `onHover: (String?) -> Void`, `open: (String) -> Void`, and `focus: (String) -> PVFieldFocus`, a graph id to the focus handle `PersonSourcesList` builds from its one `FocusState` with `.node(graphID:)` (`PVFieldFocus.init(_:equals:)`, the kit's way to hand one focus binding to a reusable view). The style reads `isFocused` in a nested body view, as `PVInkLinkBody` does; the ring's corner radius is the dot's half width plus 4pt. Node mark: a dot of **(board)** diameter drawn by the view; label mono caption. Arc colour at rest **(board)**; highlighted `PVColor.accentLine`; receded **(board)** opacity | Geometry tested apart from drawing; layout and diagram take nodes and arcs only, so Event and Place add an adapter (review M-8, M-9, M-10) | `PVFocusRingButtonStyle` alone — no pressed state. Drawing nodes on the Canvas — no focus or accessibility |
| D-6 | Interaction state has one owner, `PersonSourcesList`, which now draws both columns through `ConclusionPinnedSideColumn` (main: the header and rows; side: the diagram column). It owns the one `@FocusState focus: PersonSourcesFocus?` (gaining `.node(graphID:)`), `@State hoveredGraphID: String?`, and the derived `highlightedGraphID` = `hoveredGraphID` ?? the graph of the focused node or graph link. A line's hover (the whole line, Unpromote included) and a node's hover set `hoveredGraphID`; a line whose graph has no node highlights only itself. `PersonSourcesRowView` and `PersonSourcesLineView` take `highlightedGraphID` and a hover callback; a highlighted line has the `PVColor.surfaceSelected` background **(board may change)**. `pendingFocus` and `returnFocus` are unchanged | One owner and one derived highlight; the focus enum reaches the diagram because the same view draws it (review M-3; dry run X-2) | A second `@FocusState`; highlight state in the section beside focus in the list |
| D-7 | `PVPageVisibleRect` *(new)* in `PVPageColumns.swift`: a view `PVPageVisibleRect { visible in … }` that holds the reading state itself and hands its content the visible rectangle of the page in its own coordinates (from its frame in `PVPageColumnsMetrics.scrollSpace` and `pvPageViewportSize`). `PersonTimelineTethers` moves onto it and loses its own `@State sectionFrame` (its `visible` unchanged). `ConclusionPinnedSideColumn` uses it: the side column is a frame of the HStack's height; inside it the side view is offset by `pinOffset(visibleMinY:sectionHeight:sideHeight:inset:)` = max(0, min(visibleMinY + inset, sectionHeight − sideHeight)), `pinOffset` is a static on `PVPageColumnsMetrics`, tested in STEP-2; `inset` = `PVPageColumnsMetrics.verticalPadding` *(new constant, 24; lines 139 and 153 both use it)*; `sectionHeight` is the `HStack`'s height read with `onGeometryChange`, `sideHeight` the side view's. The side view's content is the title and caption, then the circle in the height left; the circle stops shrinking at **(board, RL-13)** and the column then scrolls with the page. The side view's height = min(side width + title and caption, visible height − 2 × inset); with no viewport (previews, tests) it is the side width plus title and caption. Until the first width is measured (one frame) the side is not drawn | One reader of the visible rectangle (review M-6); clamp order stated (dry run X-10) | A second hand-built reader |
| D-8 | The diagram column is a second `ConclusionPageSectionHost`. The host gains two parameters: `key: CatalogQueryKey?` (nil, the default, keeps `section.queryKey`; the diagram passes `companionQueryKey`) and `placement: ConclusionPageSectionHost.Placement` *(new: `.band`, the default, today's look; `.column`: the spinner centred in the column, a compact danger `PVCallout` with no `pvKeepsPageBand`, and no empty state)*. The host still gets the handle, starts the read and maps `.gone` to `conclusionDetailsNotFound`. The diagram's host sits inside `PersonSourcesList`'s side column, its `make` capturing `content.graphLocations` to build `PersonRecordLinksContent`. The column exists only when the rows' host draws content; the rows' states stay full width (AC-5.3) | One host for every section read: nil handle, the read's start, the error copy (review M-11; dry run X-20) | Mapping `ConclusionDetailPresentation` in the column — a second copy of the host's switch |
| D-9 | Strings under `L10n.Conclusions`: `recordLinksTitle`, `recordLinksCaption`, `recordLinksErrorTitle`, `recordLinksOneRecord`, `recordLinksNodeHelp(graph:source:)`, `recordLinksGroupAccessibility(count:)`, `recordLinksNodeAccessibility(ref:graph:source:connected:)`, `recordLinksNodeApartAccessibility(ref:graph:source:)`; copy **(board)** | `add-localized-string` | — |
| D-10 | Fixtures in `PersonSourcesPreviewData.swift` (`#if DEBUG`, read by tests as `PersonSourcesContentTests` does): `johannClaims` (S13-D6 § 5's seven records on five Sources, graph ids `g-214`, `g-388`, `g-vol3`, `g-221`, `g-reg2`, `g-reg3`, `g-p7`) and `johannRecordLinks` keyed to them (S13-D10 § 5: seven graphs, the eight lines with their strengths, entity ids `ent-anna`, `ent-karl`, `ent-heinrich`, `ent-maria`) and `manyClaims(graphs: 50)` (graph ids `g-m1` to `g-m50`, fourteen Sources) with `manyRecordLinks(graphs: 50)` keyed to them (graph i shares a person with i + 1 and i + 2, strong when i is even, graphs 47 – 49 apart: 91 lines) | Shared by tests and previews | Fixtures in tests only — previews need them |

**Rules this plan bends.** S13-D6's agreed placement for this view (SS-11, SS-W1: a band above the rows at Regular, the secondary column at Wide) → replaced by the user's answer, 2026-11-07: beside the rows at both widths, pinned while they scroll → the S13-D10 brief's Board line and the map's working rule "The view is a circular link diagram" record it (this planning commit). The archived S13-D6 brief is not edited.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `companionQueryKey` *(new)* | func | `queryKey`, `supportingQueryKeys` | A third meaning would bend `supportingQueryKeys` (D-2) |
| `PVPageColumnsMetrics.sideColumnWidths`, `verticalPadding` *(new)* | func, constant | `PVPageColumnsMetrics.columnWidths` | Beside it, same tokens (D-3) |
| `ConclusionPinnedSideColumn` *(new)* | view | `PVPageColumns` | Feature-level until a second caller (D-3) |
| `PVPageVisibleRect` *(new)* | view | `PersonTimelineTethers`' hand-built reading | Extracted; the Timeline moves onto it (D-7) |
| `PersonSourcesContent.graphLocations` *(new)* | property | `Line.location` | Same loop (D-4) |
| `PersonRecordLinksContent`, `RecordLinksNode`, `RecordLinksArc` *(new)* | types | `PersonSourcesContent` | Adapter for a different read (D-4) |
| `RecordLinksLayout`, `RecordLinksDiagram` *(new)* | type, view | `PersonFamilyLayout`, `PersonFamilySection.diagram`; `GraphCanvas` | Family's split; no pan, zoom or drag (D-5) |
| `PersonSourcesFocus.node` *(new)*; the diagram's `focus: (String) -> PVFieldFocus` input | case, parameter | `PersonSourcesFocus`; `PVFieldFocus` | Extended; the kit's focus handle keeps the diagram Person-free (D-5, D-6) |
| `ConclusionPageSectionHost` `key`, `placement`, `Placement` *(new)* | parameters, enum | `ConclusionPageSectionHost` | Extended so a companion read uses the one host (D-8) |
| `johannClaims`, `manyClaims` *(new)* | fixtures | `PersonSourcesPreviewData.regular` | Matching graph ids for the two columns (D-10) |
| `L10n.Conclusions.recordLinks*` *(new)* | strings | `sources*` | New copy |
| `johannRecordLinks`, `manyRecordLinks` *(new)* | fixtures | `PersonSourcesPreviewData` | Extended (D-10) |
| Tests in §8 *(new)* | tests | `PVPageColumnsTests`, `PersonSourcesContentTests` | — |

**Builds on.** Full-width mode; `PVPageColumnsMetrics`; `ConclusionDetailPresentation`; `SourcesListNavigation.canvasLocation`; `PVHoverEffect`, `pvFocusRing`, `pvActivatesOnReturn`; S13-27's key.
**Deletes.** `PersonTimelineTethers`' own `@State sectionFrame` and its `onGeometryChange` reading (moved into `PVPageVisibleRect`); the literal `PVSpacing.space8` at `PVPageColumns.swift` lines 139 and 153 (now `verticalPadding`).
**Next adds.** Event and Place pages, when they gain sections, add their section cases (each with its `companionQueryKey`), an adapter beside `PersonRecordLinksContent`, and use `ConclusionPinnedSideColumn`; § 9 gains the pattern row (STEP-9).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Node click → Evidence graph with the Subject selected | `SourcesListNavigation.canvasLocation`, as the line | A node with no location is not clickable (D-4) |
| The diagram's read failing or gone while the rows show | Diagram column | Danger Callout (AC-5.2, D-8) |
| The Timeline on `pvPageVisibleRect` | `PersonTimelineTethers` | Same rectangle as today; its tests and manual check unchanged |

### Scale

At the map's Scale (50 graphs, about 120 arcs at most for a household person): one `Canvas` redraw per highlight change; the layout is computed once per value and size. The read is S13-27's (37.9 ms cold at 50 graphs). The 50-graph preview (D-10) is the visual check; T-4 builds content for it.

### Edges

- **Assumptions.** The two reads name the same graphs; a race leaves a node without a location or a line without a node (D-4, D-6). Graph ids unique (duplicates keep the first).
- **Neighbors.** 0, 1, 2 nodes (D-5); a graph sharing with all others; a person whose rows are shorter than the diagram (the side column's height wins, D-7).
- **Removed inputs.** Remove on a line reloads both reads (S13-27's key follows the person's handle and the graph); the node leaves with its last record.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** Arcs per pair from the read: the spread is shared evenly, drawing is linear in arcs, bounded by S13-27's `MaxReadWork`.

### Schema

None — no catalog change.

### Go core

None — the read is S13-27's.

### FFI

None — S13-27's RPC.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Two columns, pinned | `ConclusionPinnedSideColumn` | `Features/Conclusions/` | D-3, D-7 |
| Diagram | `RecordLinksDiagram`, `RecordLinksLayout` | `Features/Conclusions/` | D-5 |
| Node | `Button` + private style (`PVHoverEffect`, `pvFocusRing`), `pvActivatesOnReturn` | in the diagram | D-5, D-6 |
| Title, caption | micro-caps, caption styles | diagram column | RL-9 |
| Loading, error | `ProgressView`, `PVCallout` (danger, compact) | diagram column | D-8 |
| Graph line highlight | `PVColor.surfaceSelected` | `PersonSourcesLineView` | D-6 |

Frames in words: filled from the board in STEP-0.

### Strings

L10n keys *(new)*: D-9. Copy from the board.

## 7. Steps

**STEP-0 — Board answers (planner, before `Planned`)**
Files: this plan, the S13-D10 brief.
Do: copy every **(board)** value and the frames in words into §5 and §6; close §12; review the plan again.
Check: no "(board)" left in this file.

**STEP-1 — A section's companion read; Sources fills the width**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`, `macos/ProvenenciaTests/PlaceRegistryTests.swift`. Skill: `add-workspace-place`.
Do: D-1, D-2. Add T-1, T-2.
Check: `xcodebuild test -scheme Provenencia -only-testing:ProvenenciaTests/PersonDetailTests -only-testing:ProvenenciaTests/PlaceRegistryTests` passes (the scheme and destination `scripts/check-pr.sh` uses).

**STEP-2 — One reader of the visible rectangle**
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Conclusions/PersonTimelineSection.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`. Skill: `add-ui-component`.
Do: D-7's `PVPageVisibleRect`, `verticalPadding` (both padding lines) and `pinOffset`; move `PersonTimelineTethers` onto it. Add T-3's `pinOffset` rows.
Check: `xcodebuild test` passes; the Timeline preview's tethers draw as before.

**STEP-3 — The side-column arrangement**
Files: `PVPageColumns.swift` (`sideColumnWidths`), `macos/App/Features/Conclusions/ConclusionPinnedSideColumn.swift` *(new)*, `PVPageColumnsTests.swift`.
Do: D-3, D-7's pin. Add T-3.
Check: T-3 passes.

**STEP-4 — Strings**
Files: `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-9, with the brief's recommended copy until STEP-0 replaces it.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-5 — The content**
Files: `PersonSourcesContent.swift` (`graphLocations`), `PersonRecordLinksContent.swift` *(new)*, `PersonSourcesPreviewData.swift` (D-10), `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-4, D-10. Add T-4.
Check: T-4 passes; `PersonSourcesContentTests` passes.

**STEP-6 — The layout**
Files: `RecordLinksLayout.swift` *(new)*, `macos/ProvenenciaTests/RecordLinksLayoutTests.swift` *(new)*.
Do: D-5's layout. Add T-5.
Check: T-5 passes.

**STEP-7 — The diagram**
Files: `RecordLinksDiagram.swift` *(new)*. Skill: `add-ui-component`.
Do: D-5's view; previews for Johann, one graph, 50 graphs.
Check: the previews render; `xcodebuild build` with no new warning.

**STEP-8 — Side by side, hover, focus and the diagram's states**
Files: `PersonSourcesSection.swift`.
Do: D-6, D-8: `ConclusionPageSectionHost` gains `key` and `placement`; `PersonSourcesList` draws both columns, the side column a second host over the companion key; `PersonSourcesLineView` takes the highlight and the board's RL-W2 line layout. The four existing `PersonSourcesList` previews pass no record links (the side column shows its spinner); new previews "Johann" (`johannClaims` with `johannRecordLinks`), "Many" (`manyClaims` with `manyRecordLinks`), "Diagram loading" and "Diagram error". Add T-6.
Check: `xcodebuild test`; manual check §10 steps 1 – 6.

**STEP-9 — Docs and skills**
Files: `docs/deployment-plan/spike-13/deployment-plan.md` (map row Status), `docs/macos-client-patterns.md` (line 104: the companion key, the host's `key` and `placement`, and Sources in full-width mode; § 9 "Patterns to reach for": a row "List with a pinned side column at both widths — `ConclusionPinnedSideColumn`"), `.agents/skills/add-workspace-place/SKILL.md` line 90 (the companion key), `docs/ideas/wide-layout.md` line 10 (Sources is not primary-only), `design/S13-D10-…` → `design/archive/`, `design/README.md`, `completed.md`, this plan (status, Landed note).
Do: per the DoD.
Check: `python3 scripts/check-plan-docs.py` on this plan reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PersonDetailTests.swift` `sectionQueryKeys` (updated) | `companionQueryKey`: `.personRecordLinks` for Sources, nil for the rest; `supportingQueryKeys` unchanged; `fillsWidth` true for timeline, gallery, sources | D-1, D-2 |
| T-2 | Swift | `PlaceRegistryTests.swift` `personDetailOnSourcesWarmsTheRecordLinks` *(new)* | Sources section location → keys `[.conclusionDetail, .personClaims, .personRecordLinks, .sourceTypesList]` in the spec's order | AC-5.5 |
| T-3 | Swift | `PVPageColumnsTests.swift` `sideColumnWidths`, `pinOffset` *(new)*, tables | Regular 1038, side 340 → main 666, group 1038; Wide 2232 → main capped, side at `sideMax`, group < 2232; pin: visible top above the section → 0; 300 into a 2,000pt section → 324; past the end → section − side; side taller than the section → 0; the narrow pane (board RL-W5) → the board's arrangement | AC-3.1, AC-3.2, D-3, D-7 |
| T-4 | Swift | `PersonRecordLinksContentTests.swift` *(new)* | `johannRecordLinks` → seven nodes in order, labels, `connected` 3/2/0/2/2/1/0, eight arcs with slots 0/1 of 2 on 1–2, 1–4, 5–6, `neighbors(of: 0)` = {1, 3, 4}; a no-ref graph → its name; a graph with no location → `location` nil; a duplicate graph id → one node; a pair naming an unknown graph → dropped; `manyRecordLinks` → 50 nodes, 91 arcs | AC-1.1 – AC-1.3, D-4, D-10 |
| T-5 | Swift | `RecordLinksLayoutTests.swift` *(new)* | 4 nodes in a 200pt square with no label room → top, right, bottom, left; 0 → none; 1 → centre; 2 → top, bottom; two arcs on one pair → distinct control points | AC-1.1, AC-1.2, AC-1.4, D-5 |
| T-6 | Swift | `PersonSourcesContentTests.swift` `graphLocations` *(new)* | `regular` → one location per graph, the first line's, a folded line's graph included | D-4, AC-2.5 |

Must stay green: everything `scripts/check-pr.sh` runs; `PersonSourcesContentTests`, `PlaceRegistryTests`, `PVPageColumnsTests`, the Timeline's tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: `rg -n 'supportingQueryKeys|primary-only|ConclusionPageSectionHost' docs/macos-client-patterns.md docs/ideas/wide-layout.md .agents/skills/add-workspace-place/SKILL.md` shows each hit true after this PR (landed plans and archived briefs are history and are not edited); `rg -n 'sectionFrame' macos/App/Features/Conclusions/PersonTimelineSection.swift` finds nothing; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (none)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated
- [ ] Domain docs and skills updated: macos-client-patterns.md (line 104, § 9 row), add-workspace-place line 90, ideas/wide-layout.md line 10
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On the product owner's dogfood project, at Regular and Wide, light and dark:

1. Open a person with 5 – 20 records, Sources: the circle beside the rows, an arc per shared person, two arcs between two records side by side, records that share no one apart (bar 10; AC-1.1, AC-1.2).
2. Hover a node, then a graph line (and its Unpromote button): each highlights the other; leave: everything rests; with a Source folded, hover its node: it highlights and the rows stay folded. Hover and click an arc: nothing (AC-2.1, AC-2.2, AC-2.3, AC-2.5, AC-1.5).
3. Click a node: its Evidence graph opens with the person's record selected; Back returns (AC-2.4).
4. On a person with many records, scroll the list: the diagram is held 24pt below the top, then leaves with the last rows; make the window short: the diagram fits; narrow the window to its minimum: the board's RL-W5 arrangement (AC-3.5); with a node focused, resize across 1600: focus, highlight and scroll hold (AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5).
5. Tab to the nodes, Return opens one; VoiceOver reads a node and a record that shares no one (AC-4.1, AC-4.2).
6. The previews "Diagram loading" and "Diagram error" show the spinner and the Callout beside drawn rows; the section's own spinner and error Callout still draw full width with no diagram column (open a person while the claims read is loading); in the app, after a Remove, "Refreshing" shows while both columns stay drawn; a Source type edit shows no Refreshing on the page (AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Labels collide at 50 graphs | The board's RL-1 answer; the 50-graph preview |
| Pinning jitters on scroll | The Timeline reads the same geometry without jitter today |
| `offset` and hit testing | Family places its buttons by `offset` and they hit-test; manual step 4 clicks a node after scrolling |

**Stop and report if:**

- A **(board)** value is still open when you start.
- The board's `sideRegular` is over 364pt (the citation column would fall under 200pt).
- `PersonTimelineTethers` draws differently after moving onto `pvPageVisibleRect`.
- The two reads name different graphs in a fixture the app wrote.

## 12. Open questions

- The **(board)** values in §5 and §6, answered by the S13-D10 board, among them RL-W5 (a narrow pane) and RL-W2's graph-line layout in the narrowed citation column (review X-17, X-18).
- How the agreed board reaches the implementer (the Claude Design project link, as S13-D6; recommended) — asked when the board is agreed.

## 13. PR

**Title:** S13-28 — How the records connect
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-28-how-the-records-connect.md`.
