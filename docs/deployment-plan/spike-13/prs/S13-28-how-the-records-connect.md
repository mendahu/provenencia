# S13-28 — How the records connect

**Status:** Landed in [#744](https://github.com/mendahu/provenencia/pull/744) (2026-11-08)
**Spike / slice:** Spike 13 · Slice 7 — How a person's records connect
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S13-D10](../design/archive/S13-D10-how-the-records-connect.md), board agreed 2026-11-07
**Board:** [Person Detail canvas](https://claude.ai/artifact/N9V52ewQjxeNvh3Az3YkXQ), page "Person Detail - Sources", turn 15 "S13-D10 · How the records connect" (frames 15a – 15z; the turn's decision list RL-1 – RL-13 and RL-W). The frames are also in words in §6 → Frames in words; where this plan and the board differ, this plan wins (see **Rules this plan bends**)
**Depends on:** S13-16 (landed [#722](https://github.com/mendahu/provenencia/pull/722)), S13-17 (landed [#729](https://github.com/mendahu/provenencia/pull/729)), S13-27 ([#743](https://github.com/mendahu/provenencia/pull/743), same branch), S13-D10
**Blocks:** —
**Branch:** `s13-27-plan` (the user's choice, 2026-11-07: S13-27's PR carries this plan; the PR that builds it gets its own branch, `s13-28-how-the-records-connect`)
**Runs on:** macOS: every step needs Xcode
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [`evidence-graph-comparison.md`](../evidence-graph-comparison.md) Requirement 2; working rules "The view is a circular link diagram", "Scoring and the view are two concerns", "Sections are places", "A full-width section keeps its states in the band"; dogfood bar 10; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide)
**Planned:** 2026-11-07 against `s13-27-plan` @ `cc03d7f9` (S13-27's code is on this branch, not yet on `main`)
**Review:** [`S13-28-how-the-records-connect.review.md`](S13-28-how-the-records-connect.review.md)
**Landed:** differences from the plan: `PersonSourcesPreviewData.previewSession` takes `entityId:recordLinks:failure:` (nil links answer an hour late) instead of a `delay:`; the diagram's side column is `PersonRecordLinksColumn` in its own file; `ConclusionPageSectionHost.emptyTitle` is optional (nil by default) rather than an empty string, and `emptyMessage` defaults to ""; nodes are placed with `.position`, so the one unmeasured frame is off by the dot's inset; the 700pt check was a hosted-window render, the manual check in the app is the product owner's. Round 3 (the user's request, 2026-11-08): the diagram is a generic `CircularLinkDiagram` with a `CircularLinkDiagramStyle`, `CircularLinkLayout` and `CircularLinkNode`, `CircularLinkArc` and `CircularLinkHover` data types under `Features/Conclusions/CircularLinkDiagram/`, and `PersonRecordLinksContent` is the only Person-aware piece; D-4, D-5 and D-6 name the old `RecordLinksNode`, `RecordLinksArc`, `RecordLinksLayout`, `RecordLinksDiagram` and `RecordLinksHover`, now these types. Round 1: the arcs are two layers (the rest on an `Equatable` `Canvas` dimmed by opacity, the highlighted node's arcs above), and the hover is `RecordLinksHover`, an observable the rows, nodes and arcs read, so a hover does not run the content's init or the layout; each `Curve` carries its arc; the 50-graph household preview is `householdRecordLinks(graphs: 50)`; `maxArcs` and `maxNodes` cap the content; `RecordLinksArc.neighbors(of:in:)` is the one neighbour function; `RecordLinksLayout` owns the one-node geometry; `PVPageColumnsMetrics.sideColumnWidths(proposal:mainMin:)` and `fallbackPane` hold the width fallback; `placement` is `.column(loadingHeight:)` and the diagram takes `idRoot`; a node with no location is disabled and a focused node shows its ref.

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
| AC-1.4 | A person with one graph | The one node centred, no circle and no caption line, and under it "One record — nothing to connect yet" |
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
| AC-3.3 | A window shorter than the diagram | At Wide 2560 × 900 the circle shrinks below 552pt so the column fits the visible height, labels keeping their size. At Regular the circle is already the 172pt floor; at 1366 × 600 (about 290pt visible under the pinned header and switcher, less than the column's height) the column stops pinning and scrolls with the page |
| AC-3.5 | A pane narrower than 1078pt (window under 1342pt), down to the 700pt window | The diagram stacks under the rows, 40pt below them, at the column width up to 720pt, not pinned; the rows are drawn as today (their own narrow layout is S13-D6's, [`docs/ideas/wide-layout.md` § Narrow panes](../../../ideas/wide-layout.md#narrow-panes)) |
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
| AC-5.2 | The record-links read fails | The rows drawn; the diagram column shows the danger Callout "The connections could not be read" with the error's own message as its body (the host's mapping, as every section read) |
| AC-5.3 | The claims read is loading, failed or empty | The section host's state as today, full width; no diagram column |
| AC-5.4 | Either read reloads | The switcher row shows "Refreshing"; both columns stay drawn |
| AC-5.5 | Opening the Sources section | The place warms `.personClaims` and `.personRecordLinks`; a `.sourceTypesList` reload does not show "Refreshing" |

## 3. Scope

| In | Covered by |
| --- | --- |
| RL-1 – RL-12: the diagram, hover both ways, open, a11y, dark | STEP-4 – STEP-8 |
| RL-13: pinned while the rows scroll | STEP-2, STEP-3, STEP-8 |
| RL-W1 – RL-W5: side by side at both widths; full-width mode at Wide; states at Wide; stacked below a 1078pt pane | STEP-1, STEP-3, STEP-8 |
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

Values come from the agreed board (turn 15) unless a Decision says otherwise.

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Placement (user, 2026-11-07): the diagram beside the rows at both widths, pinned while they scroll. `ConclusionPageSection.fillsWidth` is true for `.sources` | The user's answer; full-width mode is how Timeline and Gallery use Wide | A page secondary column — stacks below at Regular (rule 4) |
| D-2 | A companion read: `ConclusionPageSection.companionQueryKey(project:entityId:) -> CatalogQueryKey?` *(new)*, `.personRecordLinks` for `.sources`, nil otherwise; doc "a second per-entity read the section draws in its own region". `PlaceRegistry`'s `.personDetail` spec warms it right after `queryKey` (before `supportingQueryKeys`); `sectionRefreshing` follows `queryKey` and `companionQueryKey`. `supportingQueryKeys` is unchanged | Each list keeps one meaning: supporting keys stay warm-only vocabulary; a vocabulary reload never shows Refreshing (review M-1) | Widening `supportingQueryKeys` — bends its doc and `add-workspace-place` line 90 |
| D-3 | One split for a column beside a side column. `PVPageColumnsMetrics.split(available:sideMax:) -> (main: CGFloat, side: CGFloat, group: CGFloat)` *(new)*, pure, taken out of `columnWidths`' Wide-with-secondary branch, which now calls it with `PVSpacing.widthSecondaryMax` (its results unchanged): with A = available − `space9`, main = max(0, min(`widthContentMax`, A − `widthInspector`)), side = min(`sideMax`, max(`widthInspector`, A − main)), group = main + `space9` + side. `PVPageColumnsMetrics.sideColumnWidths(available:mainMin:) -> PVPageSideColumnWidths` *(new; struct `main`, `side`, `group`, `stacked`)*: **stacked** when available < `mainMin` + `space9` + `widthInspector` (main = available, side = min(available, `PVSpacing.widthSideColumnMax`), group = available); otherwise `split(available:sideMax: PVSpacing.widthSideColumnMax)`. `PVSpacing.widthSideColumnMax` *(new token, 720, beside `widthSecondaryMax`; doc: a side column inside a full-width section, not the page's secondary column)*. `available` is the section's width between the gutters. Person passes `mainMin` = `PersonSourcesColumns.minWidth` *(new: `fixedWidth` (type + date + ref + 3 × spacing = 442) + `citationMin` (200), both new)* = 642, so the breakpoint is an available width of 1014 (a 1078pt pane). Examples: 1038 → main 666, side 340, group 1038; 1014 → main 642; 1013 → stacked, side 720; 372 → stacked, side 372; 2232 → main 1180, side 720, group 1932; 1536 → main 1164, side 340. `ConclusionPinnedSideColumn` *(new)* in `Features/Conclusions/ConclusionPinnedSideColumn.swift` takes `mainMin`, `main` and `side` views and places them with the kit's `PVPageColumnsLayout`, made internal and given its widths as an input: a `(ProposedViewSize) -> PVPageSideColumnWidths` closure (`side` optional for a page with no secondary), built from `columnWidths` for `PVPageColumns` and from `sideColumnWidths` for the side column. Its placement is unchanged: sized from the proposed width alone; side by side, main then side `space9` apart, the group centred, height the taller; stacked, side under main `space9` apart (the kit's stacked gap; the board drew 40), leading because the group is the whole width, height the sum. One `Layout` over the same two subviews, so crossing 1078pt keeps identity, focus, hover and the diagram's read, as crossing 1600 does. The pin is a modifier on the side subview (D-7). The graph lines move out of the citation cell so they span the whole row (board RL-W: the whole row width less the lines' indent); today they sit in `PersonSourcesRowView.citation`, inside `.personSourcesCitationCell()` (204pt after the indent at Regular). `PersonSourcesRowView.body` becomes a `VStack(alignment: .leading, spacing: PVSpacing.space2)`: the existing `HStack` of cells, its citation cell holding only the title link, then the lines' `VStack` (and Show more) with its leading indent. The row's accessibility element, padding and order (title, then lines) are unchanged | One formula for a column beside a side column, one placement mechanism; Event and Place reuse the column with their own `mainMin` (review M-2, M-13, M-14) | `PersonSourcesColumns.split` — a second width calculator. A kit `PVPageColumns` option — one caller today. An `HStack`/`VStack` branch — resets state at 1078pt |
| D-4 | `PersonRecordLinksContent` *(new, Equatable)* in `Features/Conclusions/PersonRecordLinksContent.swift` *(new)*, the Person adapter: from `CatalogPersonRecordLinks` and `PersonSourcesContent.graphLocations` *(new: `[String: WorkspaceLocation]`, the first line's location per graph id, taken in `init`'s loop before the fold)*. `nodes: [RecordLinksNode]` (`id` graph id, `label` graph ref else graph name, `help` via `L10n.Conclusions.recordLinksNodeHelp(graph:source:)`, `location` optional, `connected`, `accessibilityLabel`); `arcs: [RecordLinksArc]` (`a`, `b` node indexes, `strong`, `slot`, `slots`); `neighbors(of: Int) -> Set<Int>`. `strong` is `.strong`; slots follow the read's line order; `connected` counts distinct other graphs. A duplicate graph id keeps the first; a pair naming a graph not in the list is dropped; a node with no location (the reads raced) is drawn and not clickable | One builder of claim locations (review M-4); the highlight's neighbour set lives in the model, tested | Rebuilding locations from the groups — a second builder |
| D-5 | `RecordLinksLayout` *(new)* in `Features/Conclusions/RecordLinksLayout.swift`, pure, takes node count, arcs and a size: node points on a circle clockwise from the top (0 nodes → none; 1 → centre; 2 → top and bottom). The label room is 84pt on the left and right (one ref plus 14pt) and 30pt above and below, so the circle's diameter is the column width − 168 (172pt at 340, 552pt at 720) and the drawing's height is the diameter + 60. Each node has a label side: right when cos θ > 0.35, left when cos θ < −0.35, else top above the centre and bottom below it (θ measured from the centre). Each arc is a quadratic from node to node whose control point is 30% of the way from the centre to the chord's midpoint, moved along the chord's perpendicular by (slot − (slots − 1) / 2) × spacing, where spacing = clamp(0.06 × radius, 7, 14) and, so a pair with many shared people stays inside the circle, at most radius / (slots − 1) (this plan's cap; the board drew at most three arcs on a pair). `RecordLinksDiagram` *(new)* in `Features/Conclusions/RecordLinksDiagram.swift` draws arcs on a `Canvas` with hit testing off, and each node as a `Button` placed by `offset` (the Family precedent) with a private `ButtonStyle` combining `PVHoverEffect` and `pvFocusRing` (the `PVInkLinkStyle` shape, the `SourcesGraphButtonBody` fill), `pvActivatesOnReturn`; the node's hover reaches the diagram through `PVHoverEffect`'s new optional `onHoverChange: ((Bool) -> Void)?` *(new, default nil, called from its one `onHover`)*, so one tracker serves the fill and the report. Its inputs are Person-free: `nodes`, `arcs`, `highlightedGraphID: String?`, `onHover: (String?) -> Void`, `open: (String) -> Void`, and `focus: (String) -> PVFieldFocus`, a graph id to the focus handle `PersonSourcesList` builds from its one `FocusState` with `.node(graphID:)` (`PVFieldFocus.init(_:equals:)`, the kit's way to hand one focus binding to a reusable view). The style reads `isFocused` in a nested body view, as `PVInkLinkBody` does. A node button is the dot and the ref in one row (the ref on the node's label side, `PVSpacing.space3` apart; top labels above the dot, bottom below), padding 2pt × 4pt, corner radius `PVRadius.md`, the hover background `PVColor.surfaceHover` and the focus ring around the whole button; placed so the dot's centre is the node point: each node button reads its own size with `onGeometryChange` and is offset by the dot's known place inside it (padding + 4pt from the end the dot sits on: leading for a right label, trailing for a left one, bottom for a top label, top for a bottom one); until measured it is drawn centred on the point. Dot: 8pt, a 1.5pt ring. Ref: `PVFont.mono` caption. Node states: at rest dot and ring `textSecondary`, ref `textSecondary`; highlighted dot and ring `accent`, ref `accent` semibold; a neighbour (the other end of a highlighted arc) dot `accentSoft` with an `accentLine` ring, ref `textPrimary`; every other node at 35% opacity while something is highlighted. Past 30 nodes only the highlighted or focused node shows its ref, on a `surfacePage` background (radius `PVRadius.sm`); the dots stay. The circle: 1pt `borderSubtle`. Arcs: round caps, thin 1.25pt, thick 3pt, `borderStrong` at rest; a highlighted node's arcs `accentLine`, drawn above the rest; thick drawn above thin; every other arc at 18% opacity while something is highlighted. One node: drawn centred near the top of a 64pt region with no circle, and under it the one-record line (italic caption, `textMuted`, centred) | Geometry tested apart from drawing; layout and diagram take nodes and arcs only, so Event and Place add an adapter (review M-8, M-9, M-10) | `PVFocusRingButtonStyle` alone — no pressed state. Drawing nodes on the Canvas — no focus or accessibility |
| D-6 | Interaction state has one owner, `PersonSourcesList`, which now draws both columns through `ConclusionPinnedSideColumn` (main: the header and rows; side: the diagram column). It owns the one `@FocusState focus: PersonSourcesFocus?` (gaining `.node(graphID:)`), `@State hoveredGraphID: String?`, and the derived `highlightedGraphID` = `hoveredGraphID` ?? the graph of the focused node or graph link. A line's hover (the whole line, Unpromote included) and a node's hover set `hoveredGraphID`; a line whose graph has no node highlights only itself. `PersonSourcesRowView` and `PersonSourcesLineView` take `highlightedGraphID` and a hover callback. Each line is `.frame(maxWidth: .infinity, alignment: .leading)`, padded 2pt vertically and 4pt horizontally (the lines' leading indent drops from 20 to 16 so the text stays put), `.contentShape(Rectangle())`, and reports hover with `.onHover` (which leaves the link's and Unpromote's own hover alone). A highlighted line has the `PVColor.surfaceSelected` background in a `PVRadius.sm` rounded rectangle across the whole line, and every line naming that graph takes it. Focus on a graph link or its Unpromote button (`.graph(claimID:)`, `.unpromote(claimID:)`) or on a node counts for the highlight; a claim's graph comes from `PersonSourcesContent.claimGraphs` *(new: claim id → graph id, built in the same pass as `graphLocations`)*. The same pass builds `graphSourceRefs` *(new: graph id → its Source's ref, from `group.sourceRef`)*, which D-9's fallback reads, since `CatalogRecordLinksGraph` carries no Source ref. `pendingFocus` and `returnFocus` are unchanged | One owner and one derived highlight; the focus enum reaches the diagram because the same view draws it (review M-3; dry run X-2) | A second `@FocusState`; highlight state in the section beside focus in the list |
| D-7 | `PVPageVisibleRect` *(new)* in `PVPageColumns.swift`: a view `PVPageVisibleRect { visible, size in … }` that holds the reading state itself and hands its content the visible rectangle of the page in its own coordinates and its own size (one reader: the section's height comes from it too) (from its frame in `PVPageColumnsMetrics.scrollSpace` and `pvPageViewportSize`). `PersonTimelineTethers` moves onto it and loses its own `@State sectionFrame` (its `visible` unchanged). `ConclusionPinnedSideColumn` wraps its Layout (D-3) in it and offsets the side subview by `pinOffset(visibleMinY:visibleHeight:sectionHeight:sideHeight:inset:)` = 0 when the side is taller than visibleHeight − 2 × inset (it then scrolls with the page), else max(0, min(visibleMinY + inset, sectionHeight − sideHeight)); stacked, the column applies no offset, `pinOffset` is a static on `PVPageColumnsMetrics`, tested in STEP-2; `inset` = `PVPageColumnsMetrics.verticalPadding` *(new constant, 24; lines 139 and 153 both use it)*; `visibleMinY`, `visibleHeight` and `sectionHeight` come from `PVPageVisibleRect`; only `sideHeight` is read with `onGeometryChange` on the side view. The offset does not change the Layout's size. The side view's content is the title and caption (4pt apart), then, 8pt below, the drawing (D-5) at the column's width. Its diameter comes from `RecordLinksLayout.diameter(width:fitting:) -> CGFloat` *(new, pure)*: max(172, min(width − 168, fitting − 60)) and never more than width − 168, where 172 is `RecordLinksLayout.minimumDiameter` *(new)* and `fitting` is the height left for the drawing: the viewport height (`pvPageViewportSize`, the scroll area under the pinned header and switcher) − 2 × inset − the title-and-caption block's height (read with `onGeometryChange`) − 8; with no viewport, width − 168. At the floor the side is taller than the visible area, so `pinOffset` unpins it; the diagram decides nothing about pinning (review M-15). Labels keep their size. Until the first width is measured (one frame) the side is not drawn | One reader of the visible rectangle (review M-6); clamp order stated (dry run X-10) | A second hand-built reader |
| D-8 | The diagram column is a second `ConclusionPageSectionHost`. The host gains two parameters: `queryKey: CatalogQueryKey?` (nil, the default, keeps `section.queryKey`; the diagram passes `companionQueryKey`; the host's private computed `key` uses it) and `placement: ConclusionPageSectionHost.Placement` *(new: `.band`, the default, today's look; `.column`: the spinner centred in the column, a compact danger `PVCallout` with no `pvKeepsPageBand`, and no empty state)*. `.column`: the spinner centred in a 232pt-high frame (a 172pt drawing) with no top padding; id root `persons.detail.<section>.column`, so its ids do not collide with the rows'; the diagram passes `errorTitle: L10n.Conclusions.recordLinksErrorTitle`. `emptyTitle`, `emptyMessage` and `isEmpty` gain defaults (empty strings; never empty), so the diagram's host passes none. A content with 0 nodes (the reads raced) draws nothing in the column. The host still gets the handle, starts the read and maps `.gone` to `conclusionDetailsNotFound`. The diagram's host sits inside `PersonSourcesList`'s side column, its `make` capturing `content.graphLocations` to build `PersonRecordLinksContent`. The column exists only when the rows' host draws content; the rows' states stay full width (AC-5.3) | One host for every section read: nil handle, the read's start, the error copy (review M-11; dry run X-20) | Mapping `ConclusionDetailPresentation` in the column — a second copy of the host's switch |
| D-9 | Strings under `L10n.Conclusions`: `recordLinksTitle`, `recordLinksCaption`, `recordLinksErrorTitle`, `recordLinksOneRecord`, `recordLinksNodeHelp(graph:source:)`, `recordLinksGroupAccessibility(count:)`, `recordLinksNodeAccessibility(ref:graph:source:connected:)`, `recordLinksNodeApartAccessibility(ref:graph:source:)`. Copy: title "How the records connect" (micro caps, semibold, `textMuted`); caption "An arc per shared person · thick when 2+ facts agree" (caption, `textMuted`, one line at 340pt); error title "The connections could not be read" (the body is the error's message, from the host); one record "One record — nothing to connect yet"; node help "%@ — %@" (graph name, else graph ref — Source title, else the Source's ref, as `PersonSourcesContent` falls back; the native help tag `.help`, not the kit Tooltip, so 50 nodes do not mount 50 tooltips); group "How the records connect, %lld records" (one: "1 record"); node "%@, %@, %@, connected to %lld records" (one: "1 record"); apart "%@, %@, %@, shares no one". The accessibility strings take the node's label (ref, else name), the graph name (else ref) and the Source title (else ref). No arc counts or weights are read aloud | `add-localized-string`; the board's copy | The brief's two-line caption — the board cut it to one line at 340pt |
| D-11 | Previews: `PersonSourcesList` now holds a host, so every preview gets `.environment(PersonSourcesPreviewData.previewSession(recordLinks:failure:delay:))` *(new, `#if DEBUG`)*: a `WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/preview.provenencia"), store:)` over a `FakeStore` with `seededPersonRecordLinks` set for the preview's entity id, as `ConclusionPictureViewer`'s preview builds its session. `FakeStore` gains `getPersonRecordLinksError: Error?` (thrown by `getPersonRecordLinks` when set) and `personRecordLinksDelayNanoseconds: UInt64` (slept before answering), copying `createCitationDelayNanoseconds`. "Diagram error" sets a `CoreInvokeError.coded` failure with code `catalog.already_open`; "Diagram loading" a one-hour delay; the four existing previews the same delay | Previews reach every column state through the one host | A diagram-column view fed a state by previews — a second copy of the host's switch (D-8) |
| D-10 | Fixtures in `PersonSourcesPreviewData.swift` (`#if DEBUG`, read by tests as `PersonSourcesContentTests` does): `johannClaims` (S13-D6 § 5's seven records on five Sources, graph ids `g-214`, `g-388`, `g-vol3`, `g-221`, `g-reg2`, `g-reg3`, `g-p7`) and `johannRecordLinks` keyed to them (S13-D10 § 5: seven graphs, the eight lines with their strengths, entity ids `ent-anna`, `ent-karl`, `ent-heinrich`, `ent-maria`) and `manyClaims(graphs: 50)` (graph ids `g-m1` to `g-m50`, fourteen Sources) with `manyRecordLinks(graphs: 50)` keyed to them (graph i shares a person with i + 1 and i + 2, strong when i is even, graphs 48 – 50 apart (97 pairs less the 6 touching them): 91 lines) | Shared by tests and previews | Fixtures in tests only — previews need them |

**Rules this plan bends.** S13-D6's agreed placement for this view (SS-11, SS-W1: a band above the rows at Regular, the secondary column at Wide) → replaced by the user's answer, 2026-11-07: beside the rows at both widths, pinned while they scroll → the S13-D10 brief's Board line and the map's working rule "The view is a circular link diagram" record it (this planning commit). The archived S13-D6 brief is not edited.
The S13-D10 board's Wide header (frame 15b: the 1180pt band left-aligned with the rows) → the header keeps full-width mode's centred band, as on the Timeline and Gallery, so it does not move between sections (user, 2026-11-07) → no kit change; the PR's description names the difference from frame 15b.
The board's stacked gap (frame 15t: 40pt under the rows) → `space9` (32pt), the gap the kit's stacked secondary column uses, so a stacked side column reads as the page's own (planner, review M-14) → the PR's description names it.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `companionQueryKey` *(new)* | func | `queryKey`, `supportingQueryKeys` | A third meaning would bend `supportingQueryKeys` (D-2) |
| `PVPageColumnsMetrics.split`, `sideColumnWidths`, `PVPageSideColumnWidths`, `verticalPadding`; `PVSpacing.widthSideColumnMax` *(new)* | funcs, struct, constant, token | `PVPageColumnsMetrics.columnWidths` | `split` is taken out of it; one formula for both (D-3) |
| `PVPageColumnsLayout` made internal, widths as an input | Layout | itself | `ConclusionPinnedSideColumn` is its second caller; one Layout places every column pair (D-3, review M-20) |
| `PersonSourcesColumns.fixedWidth`, `citationMin`, `minWidth` *(new)* | constants | `PersonSourcesColumns` widths | The breakpoint's reason lives with the columns (D-3) |
| `PVHoverEffect.onHoverChange` *(new)* | parameter | `PVMenuRowButtonStyle`'s `onHoverChange` | One hover tracker per node; menu rows have no focus ring and `SourcesGraphButtonBody` has its own frame (D-5) |
| `FakeStore.getPersonRecordLinksError`, `personRecordLinksDelayNanoseconds`; `PersonSourcesPreviewData.previewSession` *(new)* | test hooks | `FakeStore.createCitationDelayNanoseconds` | Copied (D-11) |
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
**Deletes.** The inline Wide-with-secondary arithmetic in `columnWidths` (moved into `split`); `PersonTimelineTethers`' own `@State sectionFrame` and its `onGeometryChange` reading (moved into `PVPageVisibleRect`); the literal `PVSpacing.space8` at `PVPageColumns.swift` lines 139 and 153 (now `verticalPadding`).
**Next adds.** Event and Place pages, when they gain sections, add their section cases (each with its `companionQueryKey`), an adapter beside `PersonRecordLinksContent`, and use `ConclusionPinnedSideColumn`; § 9 gains the pattern row (STEP-9).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Node click → Evidence graph with the Subject selected | `SourcesListNavigation.canvasLocation`, as the line | A node with no location is not clickable (D-4) |
| The diagram's read failing or gone while the rows show | Diagram column | Danger Callout (AC-5.2, D-8) |
| The Timeline on `PVPageVisibleRect` | `PersonTimelineTethers` | Same rectangle as today; its tests and manual check unchanged |

### Scale

At the map's Scale (50 graphs; a household person is S13-27's fixture, 1,225 pairs and 3,675 arcs): the arcs are drawn once for the layout and a highlight change redraws only the highlighted node's arcs; the layout is computed once per value and size. The read is S13-27's (37.9 ms cold at 50 graphs). The 50-graph preview (D-10) is the visual check; T-4 builds content for it.

### Edges

- **Assumptions.** The two reads name the same graphs; a race leaves a node without a location or a line without a node (D-4, D-6). Graph ids unique (duplicates keep the first).
- **Neighbors.** 0, 1, 2 nodes (D-5); a graph sharing with all others; a person whose rows are shorter than the diagram (the side column's height wins, D-7).
- **Removed inputs.** Remove on a line reloads both reads (S13-27's key follows the person's handle and the graph); the node leaves with its last record.
- **Real data.** None — no table-driven rule.
- **Untrusted size.** Arcs per pair from the read: the fan never grows wider than the radius (D-5's cap), drawing is linear in arcs, bounded by S13-27's `MaxReadWork`; the content keeps at most `maxArcs` (75,000, strong first) and `maxNodes` (1,000), 20 times the household Scale, since a read past `MaxReadWork` still lists every pair's lines.
- **Short and narrow windows.** A window under the 172pt floor unpins the column (D-7); a pane under 1078pt stacks it (D-3). Both are layout only: one `Layout` keeps view identity (D-3), so hover, focus, the scroll position and the read hold across them, as across 1600 (RL-W3).

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

**Frames in words** (board turn 15; sample person Johann Weber, seven records on seven graphs):

| Frame | What it shows |
| --- | --- |
| 15a Regular | Pane 1102: the Source rows (666pt, citation 224) and, 32pt beside them, the 340pt diagram column: title, caption, the 172pt circle. Seven nodes clockwise from the top in the rows' order, refs outside; Household 214 and Household 388 joined by two arcs side by side; Volume 3 and Page 7 with no arc |
| 15p, 15q Pinned | 50 graphs: scrolling the rows, the column holds 24pt under the pinned header and switcher; at the section's end it leaves with the last rows, its bottom never below the section's |
| 15c, 15d Hover | Pointer on node 1: its five arcs `accentLine`, nodes 2, 4, 5 the neighbour state, the rest receded, the Household 214 line `surfaceSelected`, help tag "Household 214 — Hamburg census 1867". Pointer on the Register 3 line: node 6 and its two arcs to Register 2 |
| 15e, 15f Large | 50 graphs, about 120 arcs: refs hidden past 30 nodes; on hover the node's ref on a page-coloured chip, neighbours' refs still hidden |
| 15g – 15k States | One graph (D-5's one-node region); two graphs with no arc (top and bottom, nothing between); the diagram loading (spinner centred in the column, rows drawn); its read failed (danger Callout in the column); Refreshing (diagram stays drawn) |
| 15l Keyboard | After the last Unpromote, Tab reaches the nodes in circle order, kit focus ring; focus highlights as hover; VoiceOver per D-9 |
| 15m, 15n Short window | 2560 × 900 Wide: the circle shrinks to fit. 1366 × 600: below the 172pt floor the column scrolls with the page, title and caption included |
| 15w, 15pd, 15x, 15y, 15z Dark | Frames 15a, 15p, 15f, 15g, 15j in dark |
| 15b, 15bh, 15be, 15pw Wide | Pane 2296: rows 1180 (citation 738), 32pt, the diagram 720pt (552pt circle), the 1932pt group centred; the header keeps its centred band (Rules this plan bends) |
| 15bo, 15bl, 15bf States at Wide | One graph, spinner and Callout centred in the 720pt column |
| 15s, 15t, 15r Narrow | Pane 1078: side by side, citation 200. Pane 1077 and the 436pt pane (700pt window): the diagram stacked under the rows, unpinned |
| 15u, 15v, 15o Section states | No records, loading, failed: the section's EmptyState, spinner or Callout full width, no diagram column |

### Strings

L10n keys *(new)* and copy: D-9.

## 7. Steps

**STEP-1 — A section's companion read; Sources fills the width**
Files: `macos/App/Features/Workspace/WorkspaceLocation.swift`, `macos/App/Features/Workspace/Session/PlaceRegistry.swift`, `macos/App/Features/Conclusions/ConclusionDetailPage.swift`, `macos/ProvenenciaTests/PersonDetailTests.swift`, `macos/ProvenenciaTests/PlaceRegistryTests.swift`. Skill: `add-workspace-place`.
Do: D-1, D-2. Add T-1, T-2.
Check: `xcodebuild test -scheme Provenencia -only-testing:ProvenenciaTests/PersonDetailTests -only-testing:ProvenenciaTests/PlaceRegistryTests` passes (the scheme and destination `scripts/check-pr.sh` uses).

**STEP-2 — One reader of the visible rectangle**
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Conclusions/PersonTimelineSection.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`. Skill: `add-ui-component`.
Do: D-7's `PVPageVisibleRect`, `verticalPadding` (both padding lines) and `pinOffset`; move `PersonTimelineTethers` onto it. Add T-3's `pinOffset` rows.
Check: `xcodebuild test` passes; the Timeline preview's tethers draw as before.

**STEP-3 — The side-column arrangement**
Files: `PVPageColumns.swift` (`split`, `columnWidths` onto it, `sideColumnWidths`, `PVPageSideColumnWidths`, `PVPageColumnsLayout` internal with its widths as an input), `macos/App/DesignSystem/Tokens/PVSpacing.swift` (`widthSideColumnMax`), `macos/App/Features/Conclusions/ConclusionPinnedSideColumn.swift` *(new)*, `PersonSourcesSection.swift` (`PersonSourcesColumns.fixedWidth`, `citationMin`, `minWidth` only), `PVPageColumnsTests.swift`.
Do: D-3, D-7's pin and its unpin. Add T-3.
Check: T-3 passes.

**STEP-4 — Strings**
Files: `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-9, with its copy.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-5 — The content**
Files: `PersonSourcesContent.swift` (`graphLocations`), `PersonRecordLinksContent.swift` *(new)*, `PersonSourcesPreviewData.swift` (D-10), `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift` *(new)*. Skill: `add-swift-test`.
Do: D-4, D-10. Add T-4, T-6.
Check: T-4 passes; `PersonSourcesContentTests` passes.

**STEP-6 — The layout**
Files: `RecordLinksLayout.swift` *(new)*, `macos/ProvenenciaTests/RecordLinksLayoutTests.swift` *(new)*.
Do: D-5's layout. Add T-5.
Check: T-5 passes.

**STEP-7 — The diagram**
Files: `RecordLinksDiagram.swift` *(new)*, `PVHoverEffect.swift` (`onHoverChange`). Skill: `add-ui-component`.
Do: D-5's view; previews for Johann, one graph, 50 graphs.
Check: the previews render; `xcodebuild build` with no new warning.

**STEP-8 — Side by side, hover, focus and the diagram's states**
Files: `PersonSourcesSection.swift`, `ConclusionPageSectionHost.swift`, `PersonSourcesPreviewData.swift` (`previewSession`), `macos/App/Platform/FakeStore.swift` (D-11), `macos/App/DesignSystem/Components/HoverEffect/PVHoverEffect.swift` (`onHoverChange` lands in STEP-7).
Do: D-3's row layout, D-6, D-8, D-11: `ConclusionPageSectionHost` gains `queryKey` and `placement`; `PersonSourcesList` gains an `entityId` input (its construction in `PersonSourcesSection` and its previews pass it), for the second host; `PersonSourcesList` draws both columns, the side column a second host over the companion key; `PersonSourcesLineView` takes the highlight, the full-width frame and the padding (D-6). The four existing `PersonSourcesList` previews use D-11's session with the delay (the side column shows its spinner); new previews "Johann" (`johannClaims` with `johannRecordLinks`), "Many" (`manyClaims` with `manyRecordLinks`), "Diagram loading" and "Diagram error". Check: `xcodebuild test`; manual check §10 steps 1 – 6.

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
| T-3 | Swift | `PVPageColumnsTests.swift` `sideColumnWidths`, `pinOffset` *(new)*, tables | `columnWidths`' Wide-with-secondary examples unchanged; `sideColumnWidths`, `mainMin` 642: Regular 1038 → main 666, side 340, group 1038, not stacked; Regular 1014 → main 642, not stacked; Regular 1013 → stacked, main 1013, side 720; Regular 372 (the 700pt window) → stacked, side 372; Wide 2232 → main 1180, side 720, group 1932; Wide 1536 → main 1164, side 340, group 1536. `pinOffset`: visible top above the section → 0; 300 into a 2,000pt section → 324; past the end → section − side; side taller than the section → 0; side taller than visible − 48 → 0 | AC-3.1, AC-3.2, AC-3.5, D-3, D-7 |
| T-4 | Swift | `PersonRecordLinksContentTests.swift` *(new)* | `johannRecordLinks` → seven nodes in order, labels, `connected` 3/2/0/2/2/1/0, eight arcs with slots 0/1 of 2 on 1–2, 1–4, 5–6, `neighbors(of: 0)` = {1, 3, 4}; a no-ref graph → its name; a graph with no location → `location` nil; a duplicate graph id → one node; a pair naming an unknown graph → dropped; `manyRecordLinks` → 50 nodes, 91 arcs | AC-1.1 – AC-1.3, D-4, D-10 |
| T-5 | Swift | `RecordLinksLayoutTests.swift` *(new)* | 4 nodes at width 340 → diameter 172, height 232, points top, right, bottom, left with label sides top, right, bottom, left; 0 → none; 1 → centre; 2 → top, bottom; two arcs on one pair → control points 1 × spacing apart (offsets ±0.5 × spacing), spacing 7 at radius 86; 30 arcs on one pair → the fan no wider than the radius; `diameter`: width 340, fitting nil → 172; 340, fitting 300 → 172; 340, fitting 200 → 172; 720, fitting 500 → 440; 720, fitting nil → 552; 720, fitting 100 → 172 (D-7) | AC-1.1, AC-1.2, AC-1.4, AC-3.3, D-5, D-7 |
| T-6 | Swift | `PersonSourcesContentTests.swift` `graphLocations` *(new)* | `regular` → one location per graph, the first line's, a folded line's graph included; `claimGraphs` maps every claim to its graph; `graphSourceRefs` gives each graph its Source's ref | D-4, D-6, AC-2.5 |

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
4. On a person with many records, with a node focused resize across 1342pt: focus, highlight and the diagram hold; scroll the list: the diagram is held 24pt below the top, then leaves with the last rows; make the window short: the circle shrinks, and at 600pt high the column scrolls with the page; narrow the window below 1342pt: the diagram stacks under the rows, unpinned (AC-3.5); with a node focused, resize across 1600: focus, highlight and scroll hold (AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5).
5. Tab to the nodes, Return opens one; VoiceOver reads a node and a record that shares no one (AC-4.1, AC-4.2).
6. The previews "Diagram loading" and "Diagram error" show the spinner and the Callout beside drawn rows; the section's own spinner and error Callout still draw full width with no diagram column (open a person while the claims read is loading); in the app, after a Remove, "Refreshing" shows while both columns stay drawn; a Source type edit shows no Refreshing on the page (AC-5.1, AC-5.2, AC-5.3, AC-5.4, AC-5.5).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Labels collide at 50 graphs | Refs hidden past 30 nodes (D-5); the 50-graph preview |
| Pinning jitters on scroll | The Timeline reads the same geometry without jitter today |
| `offset` and hit testing | Family places its buttons by `offset` and they hit-test; manual step 4 clicks a node after scrolling |

**Stop and report if:**

- `PersonSourcesColumns` widths differ from 176 / 120 / 104 with spacing 14 (then `minWidth` and the 1078pt breakpoint are wrong).
- `columnWidths` returns different numbers after it moves onto `split` (an existing `PVPageColumnsTests` row fails).
- `PersonTimelineTethers` draws differently after moving onto `PVPageVisibleRect`.
- The two reads name different graphs in a fixture the app wrote.

## 12. Open questions

None.

## 13. PR

**Title:** S13-28 — How the records connect
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-28-how-the-records-connect.md`.
