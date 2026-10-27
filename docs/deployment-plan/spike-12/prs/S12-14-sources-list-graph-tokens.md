# S12-14 — Graph counts in the Sources list and on the Source page

**Status:** Planned
**Revision:** Revision 2 (2026-10-27), on the open [#576](https://github.com/mendahu/provenencia/pull/576). Revision 1 (graph tokens under each Source's title) was built on the branch through `37571892` and reviewed in two rounds ([ship log](S12-14-sources-list-graph-tokens.ship.md)). After testing it, the researcher redrew the list on the board (2026-10-27): a four-column row (Source · Type · Status · Evidence graph) with one stacked button per graph and one status badge per graph. Revision 2 keeps everything below the row (the session reads, the count state, status per graph, the Source page count line, the `primary_count` removal, `PVFlowLayout`, `PVStroke`) and rewrites the list row. Revision 1's text is `git show 37571892:docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-tokens.md`.
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S12-D5](../design/archive/S12-D5-sources-list-graph-tokens.md); board redrawn by the researcher and agreed 2026-10-27
**Board:** [Claude Design](https://claude.ai/design/p/118d2eb7-3b46-4912-a688-09ec7b0070fa?via=share), agreed 2026-10-27: `S2-04 Sources list.dc.html` (frames 1–13 and its Specification panel) and `Sources List Frame.dc.html` (the interactive list it embeds, presets `times`, `help`, `hover`, `blocked`, `loading`, `failed`, `narrow`, `twelve`, `rowmenu`). The planner rendered both and checked each frame against §6. §6 Frames and §5 describe them in words for an implementer who cannot open the link
**Depends on:** S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528)), S12-11 (landed [#549](https://github.com/mendahu/provenencia/pull/549)), S12-13 (landed [#562](https://github.com/mendahu/provenencia/pull/562)), S12-D5 (board agreed 2026-10-27)
**Blocks:** —
**Branch:** `s12-14-sources-list-graph-tokens` (exists; PR #576 open). Continue on it; do not rebase or squash
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut; no schema change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [where graphs show](../source-structure.md#where-graphs-show); working rules: "Sources list", "Status is per graph", "Source page", "One graph list", "Name and description are editable; graphs list alphabetically by name", "High tolerance for churn"
**Planned:** 2026-10-27 against branch `s12-14-sources-list-graph-tokens` @ `37571892` (`main` @ `f22b19a7`)
**Review:** [`S12-14-sources-list-graph-tokens.review.md`](S12-14-sources-list-graph-tokens.review.md) — revision 1: accordion plan and token plan, every blocker Fixed; revision 2: Fits — Ready, 2 rounds, every blocker Fixed (log section "Revision 2")

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

A researcher sees every Evidence graph of every Source in the Sources list. Each row has four columns: the Source link, its type, one status badge per graph, and one button per graph that shows the graph's name and counts and opens its canvas. The Source page's graph cards show the full count line. The Status filter reads each graph: In progress means "has Subjects, none promoted yet", and a graph with at least one Subject promoted counts as Promoted.

**Why now.** S12-11 to S12-13 let a Source hold several graphs, but `main`'s list still opens and counts only the first one, and the Source page shows no progress. Revision 1's token line answered that; the researcher tested it and redrew the row so that status is visible per graph and each graph is a full-size button (board, 2026-10-27).

**What stays.** The list header and count line, the toolbar, the empty and page-level states, the status-filter spinner and callout, Add source and the Copy… form. Search matches Source title and ref only. Counts come from the existing `CatalogQueryKey.evidenceGraphProgress` map; no new query or RPC. Everything revision 1 built below the row stays as built (§4 "Built on the branch").

## 2. User stories

**US-1.** As a researcher, in the Sources list, I see each Source's Evidence graphs, their status and their counts, and can open any of them, so that I can reach the Jones household without opening the Source page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A census Source with graphs "Smith household", "Default", "Jones household", all linked to an Artifact; the list loaded | The census row's graph column stacks three buttons in the order Default, Jones household, Smith household; the Status column stacks three badges, each level with its button |
| AC-1.2 | Jones household has 12 Subjects, 40 Observations, none promoted; Default has none; Smith has 7 and 19 with one promoted | Buttons read "Jones household / 12 subjects · 40 observations", "Default / 0 subjects · 0 observations", "Smith household / 7 subjects · 19 observations"; badges read IN PROGRESS (warning), NOT STARTED (neutral, outlined), PROMOTED (success) |
| AC-1.3 | Click Jones household's button | The canvas opens on Jones household's graph id (`WorkspaceLocation.evidenceGraph(sourceId:graphId:ref:title:)`) |
| AC-1.4 | A fourth graph "Unlinked" links no Artifact | Its button is dashed and faint, reads "Unlinked / *Needs an artifact to open*", has no chevron, does nothing on click, stays focusable, and its help is "Unlinked — Add an artifact to open this evidence graph" |
| AC-1.5 | A Source with no Artifact | One dashed block "No evidence graph yet / *Needs an artifact*" with help "Add an artifact on the source page to build its evidence graph"; the Status column shows a faint "—" |
| AC-1.6 | The graph list has not loaded | One button-sized skeleton (two rules) in the graph column and a faint "—" in the Status column. When the list fails: one block with the alert glyph reading "Evidence graphs / *Couldn’t be read*", help `L10n.Errors.message(for:)` of that failure |
| AC-1.7 | The graph list has loaded and the counts have not | Each button keeps its name, shows a rule where its counts go and still opens; each badge is a rule |
| AC-1.8 | Hover the census title or cover, then click | The title turns the link-hover color and shows its underline, the cover outline darkens, no fill; the Source page opens. The underline does not jump or show twice while it fades in |
| AC-1.9 | Add an Observation on Jones's canvas, Back to the list | Only Jones's button counts change ("12 subjects · 41 observations"), with no list reload |
| AC-1.10 | The list narrower than 880 pt | The Type column and its caption go; the type joins the Source's second line ("Census schedule · SRC-7KQ2D · Yesterday"); Status narrows to 120 pt and the graph column to 240 pt. Twelve graphs stack in full; a 60-character name truncates inside its button |
| AC-1.11 | The census row | The second line under the title reads the ref in mono, then "·", then a history glyph and "Today, 3:15 PM" with help "Last updated 14 Oct 2026 · 3:15 PM MDT"; a Source with no recorded update shows a faint "—" after the glyph |
| AC-1.12 | The column caption | Reads SOURCE · TYPE · STATUS · EVIDENCE GRAPH over the matching columns, and stays pinned while the list scrolls |

**US-2.** As a researcher, when I filter the Sources list by status, a Source shows when any of its graphs has that status. *(Built in revision 1; AC-2.6's first label changes.)*

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A graph with 0 Subjects | Its status is Not started (`EvidenceStatus.empty`) |
| AC-2.2 | A graph with 5 Subjects, 0 filed | In progress |
| AC-2.3 | A graph with 5 Subjects, 1 filed | Promoted |
| AC-2.4 | A graph with 2 Subjects that are not of a primary kind (filed is always 0) | In progress |
| AC-2.5 | Census: Default Not started, Jones In progress, Smith Promoted | The census shows under each of the three, with all three buttons and badges each time |
| AC-2.6 | The Status menu | Options read All statuses, Not started, In progress, Promoted |
| AC-2.7 | A settings file saved with `"status":"filed"` | The list restores with Promoted chosen |

**US-3.** As a researcher, on a Source page, each Evidence graphs card shows that graph's counts. *(Built in revision 1; unchanged.)*

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | The census page; Jones has 12 Subjects and 40 Observations | Jones's card shows "12 subjects · 40 observations" under its ref; Default's shows "0 subjects · not started" |
| AC-3.2 | The page restored straight from a relaunch | The count lines load |
| AC-3.3 | Add an Observation on Jones's canvas, Back to the page | Only Jones's card count moves |

**US-4.** As a contributor, the graph progress row carries only the counts something reads. *(Built in revision 1; unchanged.)*

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `graphprogress.Progress`, `EvidenceGraphProgress` (proto and Swift) | No primary count field; proto field 5 and its name are reserved |

## 3. Scope

| In | Covered by |
| --- | --- |
| Built in revision 1 and kept: session reads, `QueryRead`, count state, status per graph, Source page count line, `primary_count` removal, `PVFlowLayout`, `PVStroke` (D-3, D-5 – D-8, D-13, D-14, D-17 – D-20) | STEP-1 – STEP-5 (done at `37571892`) |
| Status badge: icons, tone, label, "Not started" (D-22) | STEP-6, AC-1.2, AC-2.6 |
| One count fold with the status; one count renderer (D-3a, D-3b, D-23) | STEP-7, AC-1.2, AC-1.7, AC-3.1 |
| The graph button and its states; the four-column row, caption, narrow fold, Source link subline (D-10 – D-12, D-15, D-16, D-21, D-24, D-25) | STEP-8, AC-1.1 – AC-1.12 |
| Map, working rules and docs follow the board (D-26) | STEP-9 |

| Out | Home |
| --- | --- |
| Graph search in the list; omnibar graph hits | S12-15 |
| Create, edit, delete or link graphs from the list | not planned (Source page owns them) |
| A Promote button or filed counts on a button or card | not planned (map scope) |
| A Source-level status or total | not planned (board, out of scope) |
| Right-click on the Type and Status cells, and Shift-F10, opening the row menu (board Copy section) | not planned here: the menu stays on the Source link, as today (D-27); `pvRightClickMenu` has no keyboard trigger (revision 1 review log) |
| The status-filter callout's title and wording on board frame 3f | not planned here: today's callout stays (§1 What stays) |
| The toolbar's leading alignment and search shrink (board "toolbar") | not planned here: today's toolbar stays (§1 What stays) |
| An underline offset of 2 pt (board "Source link") | not planned: SwiftUI's `underline` has no offset on macOS 14 |

## 4. Current state

**Built on the branch (revision 1, keep as is).** Read these; STEP-6 – STEP-8 build on them.

| Area | Where | Today (at `37571892`) |
| --- | --- | --- |
| Query read | `macos/App/Features/Workspace/Session/QueryHandle.swift` `QueryRead`, `QueryHandle.read` | Folds value / error / loading. |
| Session reads | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` | `evidenceGraphsRead(sourceID:)`, `evidenceGraphProgress(graphID:)`, `evidenceGraphCountsLoading`, `evidenceGraphCountsError`. |
| Count state | `macos/App/Features/Sources/SourcesGraphCounts.swift` `SourcesGraphCounts` | `.loading`, `.notStarted`, `.counts(subjects:observations:)`, `.unavailable(message:)`; `init(progress:countsLoading:countsError:)`. |
| Formatter | `macos/App/Platform/L10n.swift` `L10n.Sources.graphCounts(_:style:)`, `GraphCountsStyle { compact, full, spoken }` | `compact` "12 · 40" / "not started" (only the token reads it); `full` "12 subjects · 40 observations" / "0 subjects · not started" / "Counts could not load"; `spoken` "12 subjects, 40 observations" / "0 subjects, not started" / "Counts could not load. <message>"; loading → "counts loading". |
| Status | `macos/App/Features/Sources/SourcesModel.swift` `EvidenceStatus` | `.empty`, `.inProgress`, `.promoted`; `init(_ progress:)`; `label` → `statusEmpty` ("Empty"), `statusInProgress`, `statusPromoted`. `statuses(for:)`, `visibleSources`, `statusCountsError` (with its comment), `isAwaitingFilterData`. |
| Source page | `SourcesGraphCountLine.swift`, `SourceGraphsSection.swift`, `SourcePageGraphsView.swift` | The card's full count line, with the `.unavailable` mark. |
| Row | `macos/App/Features/Sources/SourcesListRow.swift` `SourcesListRow`, private `SourcesLinkLabel`, `SourcesLinkButtonStyle`, `SourcesLinkButtonBody`, the `sourcesLinkHighlighted` environment key, `open(_:)`, `tokens`, `coverSize` (44), `tokenIndent` | Two lines: link + `PVListDetailCell` Updated + ref; then a `PVFlowLayout` of `SourcesGraphToken`. The title is always underlined, with the underline color clear until highlighted (`37571892`, the hover fix). `pvRightClickMenu` with `copySourceEntry` on the link. |
| Token | `macos/App/Features/Sources/SourcesGraphToken.swift` `SourcesGraphToken` (`Kind`, `Open`), private `SourcesGraphTokenFrame`, `SourcesGraphTokenLook { live, blocked, inert }`, `SourcesGraphTokenButtonStyle` | 26 pt pill per state; replaced by STEP-7. |
| List wiring | `macos/App/Features/Sources/SourcesListView.swift` `listBody`, `captionBand` | `LazyVStack` with a pinned `Section` header `captionBand` (one `Text(L10n.Sources.listCaption)`); rows `SourcesListRow(source:model:).id(source.id)`; `ListScrollAnchorModifier`. |
| Strings | `L10n.Sources`: `listCaption`, `graphNeedsArtifact(_:)` ("%@ needs an artifact"), `graphsUnavailable` ("Evidence graphs could not load"), `graphTokenHelp(name:detail:)` ("%1$@ — %2$@"), `statusEmpty` ("Empty"), `needsArtifact` ("Needs an artifact"), `needsArtifactTooltip`, `graphNeedsArtifactTooltip` ("Add an artifact to open this evidence graph"), `openEvidenceGraphLabel(name:)` ("Open graph %@"), `graphCountsUnavailable`, `graphCountsLoading` | — |
| Kit | `DesignSystem/Components/Badge/PVBadge.swift` `PVBadge(_:tone:icon:subtle:)`, `PVBadgeTone` | Uppercase micro pill; `subtle` draws an outline instead of a fill. |
| Kit | `DesignSystem/Components/Icon/PVIcon.swift` `PVSymbol` | Has `circleDashed`, `history`, `chevronForward` ("chevron.right"), `network`, `danger` ("exclamationmark.octagon.fill"). No circle-dot, circle-check or circle-alert. |
| Kit | `DesignSystem/Recipes/LastUpdated/PVLastUpdated.swift` `PVLastUpdated.detail(_:in:)` → `PVListDetail(text:help:icon:faint:)`; `PVListDetailCell` in `DesignSystem/Components/List/PVList.swift` | The cell draws a 14 pt icon at a fixed column width. |
| Kit | `DesignSystem/README.md` FlowLayout row | Says "used by the composer and the Sources list token line". |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` `graphProgressCopyMatchesBoard`, `graphTokenCopyMatchesBoard`, `listCaptionMatchesBoard`, `evidenceStatusFromCounts` | Assert the token copy, `compact` style and the caption. |
| Docs | `../deployment-plan.md` lines for bar outcome 5, working rules "Sources list" and "Status is per graph", the S12-14 map row (says "Landed #576") and scope entry, the S12-D5 design row; `../completed.md` S12-14 paragraph; `../source-structure.md` "Where graphs show"; `../design/README.md` S12-D5 line | Describe tokens; the map row and `completed.md` were written as if #576 had merged. |

**Precedent.** S10-41 drew the Sources list's last-updated column with `PVLastUpdated` (reuse its `detail` for the subline). The board's graph button follows the token's frame and look (revision 1 D-11): same hover fill, focus ring and dashed blocked look, at 52 pt.

## 5. Decisions

Kept from revision 1 (built; do not change): **D-3** count state and formatter, **D-5** status rule, **D-6** `"filed"` → `.promoted`, **D-7** any-graph match, **D-8** Source page place key, **D-9** interim helpers deleted, **D-13** session progress reads, **D-14** `QueryRead`, **D-17** `evidenceGraphsRead`, **D-18** `primary_count` removed, **D-19** `PVFlowLayout` (its README row changes in STEP-9), **D-20** `PVStroke`. Review round 1 of revision 1 added `.unavailable` to D-3 and the name-first help (both kept). Revision 1's D-1, D-4 (two-line row, nothing expands) are replaced by D-16 and D-24; D-2, D-10, D-11, D-12, D-15, D-16 are revised below, and D-3 gains D-3a and D-3b.

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-2 | `SourcesListRow` stays one `ForEach(rows)` element with `.id(source.id)`, one call site in `listBody`, now `SourcesListRow(source:model:columns:)`. It reads the graph list and progress through the session. | `ListScrollAnchorModifier` anchors on Source ids. | Graph buttons as list items. |
| D-3a | `GraphCountsStyle` drops `compact` in STEP-8, when the token that reads it is replaced; `full` and `spoken` stay. `L10n.Sources.graphCounts(_:style:notStarted:)` gains `notStarted: NotStartedReading = .words` *(new enum, nested beside `GraphCountsStyle`: `words`, `zeroCounts`)*. With `.zeroCounts`, `.notStarted` reads as zero counts: full "0 subjects · 0 observations" (`graphCountLine(subjects: 0, observations: 0)`), spoken "0 subjects, 0 observations"; every other case is unchanged. The Source page passes nothing (`.words`: "0 subjects · not started"); the graph button passes `.zeroCounts`. | The board's button always shows both numbers and the badge carries "not started"; the Source page keeps "0 subjects · not started" (frame 12). One formatter, one parameter for the one difference (review M-6). | Two more style cases that each repeat `full` / `spoken` (review M-6); changing frame 12. |
| D-3b | `SourcesGraphCounts.counts` gains the filed count: `case counts(subjects: Int, observations: Int, filed: Int)`, from `progress.filedCount`. `SourcesGraphCounts` gains `var status: EvidenceStatus?`: `.notStarted` → `.empty`; `.counts` with `filed > 0` → `.promoted`, else `.inProgress`; `.loading`, `.unavailable` → nil. `EvidenceStatus.init(_ progress:)` becomes `self = SourcesGraphCounts(progress: progress, countsLoading: false, countsError: nil).status ?? .empty` (the init never yields `.loading` or `.unavailable` with those arguments). Every `.counts(let s, let o)` pattern adds `_`; every construction in a test adds `filed: 0` (T-1 – T-10 change only by that argument). | One fold for a graph's counts and its status: the filter, the badge and the button's VoiceOver label read it, so a badge cannot disagree with its counts (review M-3). T-1 passes unchanged. | A separate `status` on the button's `Kind`, set in the row (review M-3). |
| D-10 | The Source link keeps revision 1's button, `SourcesLinkButtonStyle`, `SourcesLinkButtonBody`, highlight key, identifier, accessibility label and value and `pvRightClickMenu`, and the hover fix at `37571892` (underline always on, its color `.clear` until highlighted; the comment stays). Label changes: cover `SourcesListRow.coverSize` = 40; title as today; the second line is D-25 (the type line goes). | Board "Source link"; AC-1.8. | Toggling `.underline(_:)` (shows two underlines mid-fade); `PVButton(.link)`. |
| D-11 | `SourcesGraphToken.swift` is renamed `SourcesGraphButton.swift` (`git mv`) and its types `SourcesGraphButton`, `SourcesGraphButtonFrame`, `SourcesGraphButtonLook`, `SourcesGraphButtonStyle`. `Kind` and `Open` keep their cases and signatures: `.graph(_, counts:, open:)`, `.needsArtifact`, `.listLoading`, `.listFailed(Error)`; the status is `counts.status` (D-3b). Frame: full column width, height `SourcesGraphButton.height` = 52, padding vertical `PVSpacing.space3` (6), horizontal `PVSpacing.space4 + PVSpacing.space1` (10), `PVRadius.sm`, 1 pt `PVColor.borderDefault` on `PVColor.surfaceCard`. Content `HStack(spacing: PVSpacing.space4 + PVSpacing.space1)`: `PVIcon(.network, size: 15)`; a `VStack(alignment: .leading, spacing: PVSpacing.spacePx)` of the name line and the detail line, `frame(maxWidth: .infinity, alignment: .leading)`; for an opening graph, `PVIcon(.chevronForward, size: 14)`. The style keeps revision 1's look: live → glyphs `textMuted`, hover or press fill `surfaceHover`, border `borderStrong`, glyphs `textPrimary`; blocked → dashed `PVStroke.dashed` `borderDefault`, content `textFaint`, no hover fill; inert → `borderSubtle`, no hover. Focus ring as today. The states that are not buttons (No Artifact, list loading, list failed) draw `SourcesGraphButtonFrame(look:)` directly (`.blocked` for No Artifact, `.inert` for the other two) with `.accessibilityElement(children: .ignore)` and the label and hint from the table. States: the table below D-27. | Board "graph button", "blocked", "loading · failed", frame 9; one view owns every state (revision 1 M-14); the board names it `SourcesGraphButton`. | A new file beside the token; `PVChip` (board frame 13: a link to a place, not a selectable pill). |
| D-12 | `SourcesListRow`'s entries stay as built (`tokens`, renamed `entries`: `!hasArtifact` → `.needsArtifact`; `evidenceGraphsRead` `.loading` → `.listLoading`, `.failed(e)` → `.listFailed(e)`, `.loaded(graphs)` → one `.graph` per graph through `open(_:)`). Each entry draws one line holding its status slot and its button (D-16), so badges and buttons cannot disagree in count, order or height. | One read, one line per graph (review M-4). | A second walk of the graph list for a Status column. |
| D-15 | `captionBand` draws the four column captions on the row's columns (D-24): `L10n.Sources.columnSource`, `columnType` (omitted when compact), `columnStatus`, `columnEvidenceGraph`, each `.pvMicroCaps()` with `textMuted`, vertical padding `PVSpacing.space3 + PVSpacing.spacePx` (7), framed by the same `SourcesListColumns` widths and insets as the row (D-24). Band background, divider, identifier `sources.captionBand` and pinning as today. `listCaption` is deleted. | AC-1.12; board caption row. | One caption text (revision 1). |
| D-16 | Row layout: `HStack(alignment: .top, spacing: 0)` of the Source cell, the Type cell (wide only) and the graph stack; horizontal padding as today (leading `gutterPage - space2`, trailing `gutterPage`); no vertical padding; bottom `PVDivider()`; `.accessibilityElement(children: .contain)`, label the Source title, identifier `sources.row.<id>`. The Source cell: `frame(maxWidth: .infinity, minHeight: SourcesListRow.firstSlot, alignment: .leading)` where `firstSlot` = 68 (= 8 + 52 + 8, so it centres on the first graph), padding trailing `SourcesListColumns.sourceTrailing`. The Type cell: `Text(verbatim: typeLabel)`, `PVFont.body(size: PVTypeScale.caption)` `textSecondary`, one line, tail truncation, `frame(width: columns.type, minHeight: firstSlot, alignment: .leading)` with the text inset by `SourcesListColumns.cellTrailing` on its trailing edge. The graph stack: `VStack(spacing: PVSpacing.space3)` of one line per entry, padding vertical `PVSpacing.space4`; each line is `HStack(spacing: 0)` of the status slot (D-21; `frame(width: columns.status, alignment: .leading)`, trailing inset `SourcesListColumns.cellTrailing`) and the button (`frame(width: columns.graphs)`, then `.padding(.leading, SourcesListColumns.graphLeading)`, so the button itself is 300 or 240 wide). The stack draws the hairline once: `.overlay(alignment: .leading) { PVDivider(axis: .vertical).padding(.leading, columns.status) }`, after the vertical padding so it spans the row. No row hover, no row hit area. Tab order follows layout: the link, then the graph buttons. | Board "row", frame 9 metrics; one line per graph keeps each badge level with its button by construction (review M-4); AC-1.1. | Two column `VStack`s held level by a shared height (review M-4); `PVTable` (a selection table with a row hit area); `Grid`. |
| D-21 | Status slot, one per entry, content leading and vertically centred in its line: `.graph` whose `counts.status` is non-nil → `status.badge` (D-22); `.graph` with `counts == .loading` → `PVSkeletonBar(width: 72, height: 6)`, `accessibilityHidden` (the button says "counts loading"); any other entry (`.unavailable`, `.needsArtifact`, `.listLoading`, `.listFailed`) → `Text(verbatim: "—")` `PVFont.mono(size: PVTypeScale.caption)` `textFaint`, `accessibilityHidden`. | Board "type · updated · status"; a count failure has no status to show (the button carries the error). | A badge for unavailable counts. |
| D-22 | `EvidenceStatus` gains `badgeTone: PVBadgeTone` (`.empty` → `.neutral`, `.inProgress` → `.warning`, `.promoted` → `.success`), `badgeIcon: PVSymbol` (`.circleDashed`, `.circleDot`, `.circleCheck`), `badgeSubtle: Bool` (`.empty` only) and `var badge: some View` = `PVBadge(label, tone: badgeTone, icon: badgeIcon, subtle: badgeSubtle).accessibilityLabel(Text(verbatim: L10n.Sources.statusBadgeLabel(self)))`. `statusBadgeLabel(_:)` is `L10n.format(<resource>, L10n.string(status.label))`, as `openEvidenceGraphLabel` formats its argument. `PVSymbol` gains `circleDot = "smallcircle.filled.circle"` (Lucide `circle-dot`), `circleCheck = "checkmark.circle"` (Lucide `circle-check`), `circleAlert = "exclamationmark.circle"` (Lucide `circle-alert`), each with a one-line doc comment naming the Lucide glyph, beside `circleDashed`. `statusEmpty`'s default becomes "Not started" (key unchanged, comment "Status filter option and badge: a graph with no subjects"). The case stays `.empty`, raw value `"empty"` (stored settings). | Board: the badge labels are the filter's own, so the first option reads Not started; one owner for the status's look. | Renaming the case (breaks stored settings for a label); icons in the row. |
| D-23 | `SourcesGraphCountLine` is the one view that draws a `SourcesGraphCounts` (review M-1). It takes `counts: SourcesGraphCounts` and `style: Style` *(new, `enum Style { case card, button }`)*; the Source page passes `.card` and its existing `progress` / `countsLoading` / `countsError` turn into `counts` at the call site. Text cases render `Text(Self.countsText(_:style:))`, where `countsText` *(new, static)* builds an `AttributedString` from `graphCounts(_, style: .full, notStarted: style == .button ? .zeroCounts : .words)`: `PVFont.body(size: PVTypeScale.micro)` throughout, `PVFont.mono(size: PVTypeScale.micro)` on each run of ASCII digits, and italic on the run equal to `L10n.string(graphNotStarted)` when present; colour `textMuted`, one line, tail truncation. This replaces the card's hand-split `.notStarted` `HStack` and its comment. `.loading` → `PVSkeletonBar` (card: as today; button: `width: 132, height: 6` with vertical padding `PVSpacing.space1`). `.unavailable` → `HStack(spacing: PVSpacing.space2) { PVIcon(.circleAlert, size: 11) in PVColor.danger; text }` with `.help(message)` in both styles (the card's `.danger` glyph becomes `.circleAlert`). The accessibility label stays `graphCounts(_, style: .spoken, notStarted:)` with the same reading. `SourcesGraphButton` puts `SourcesGraphCountLine(counts:style: .button)` in its detail slot. | One renderer and one mono-digit rule for a graph's counts (review M-1); the words come from the one localized line, so a translation keeps its order. | A second renderer in the button (review M-1); `StyledFormat` (its arguments are `%@`; the counts are `%lld` plurals). |
| D-24 | `SourcesListColumns` in `macos/App/Features/Sources/SourcesListColumns.swift` *(new)*: `struct SourcesListColumns: Equatable { let isCompact: Bool; var type: CGFloat? { isCompact ? nil : 150 }; var status: CGFloat { isCompact ? 120 : 132 }; var graphs: CGFloat { isCompact ? 240 : 300 }; static let sourceTrailing = PVSpacing.space6; static let cellTrailing = PVSpacing.space5; static let graphLeading = PVSpacing.space6; static let compactBelow: CGFloat = 880; init(listWidth: CGFloat) }` (`isCompact = listWidth < compactBelow`; the insets are static and read as `SourcesListColumns.sourceTrailing` and so on). The view that owns `listBody` keeps `@State private var columns = SourcesListColumns(listWidth: .infinity)` and sets it from `.onGeometryChange(for: SourcesListColumns.self) { SourcesListColumns(listWidth: $0.size.width) } action: { columns = $0 }` on the `ScrollView`; `captionBand` and every row take `columns`. The first frame lays out wide and corrects on the first geometry report; that is accepted (review X-5). | One definition of the widths and insets the caption and rows share, as `PVTable` columns do (review M-8); board "narrow". | Widths and paddings restated in caption and row (review M-8); `ViewThatFits` (picks by content, not list width); `GeometryReader` in each row. |
| D-25 | The link's second line, `HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3)`, `lineLimit(1)`: when `columns.isCompact`, the type (`PVFont.body(size: PVTypeScale.caption)` `textMuted`, truncates first) and `Text(verbatim: "·")`; then the ref (`PVFont.mono(size: PVTypeScale.micro)` `textMuted`, fixed size); `Text(verbatim: "·")`; then `PVListDetailCell(detail: PVLastUpdated.detail(source.updated.at, in: .current), width: nil, iconSize: 12)`. `PVListDetailCell` (`DesignSystem/Components/List/PVList.swift`) gains `width: CGFloat?` (nil hugs its content: no `frame(width:)`, still `layoutPriority(1)`) and `iconSize: CGFloat = 14`; existing callers do not change. Its doc comment's "(the Sources split row)" becomes "(the Sources list row's second line)", and `PVLastUpdated`'s doc comment says the Sources list draws it on the row's second line. Separators are `accessibilityHidden`; the link's accessibility value stays `PVLastUpdated.label(...)`. | One renderer for a `PVListDetail` (review M-2); board "Source link", "narrow", AC-1.11. | A copy of the cell inline (review M-2). |
| D-26 | Docs follow the board: the map's bar outcome 5, working rule "Sources list" (rewritten), "Status is per graph" (the first status reads "Not started"), the S12-14 row (scope text; status `In review #576`) and scope entry, the S12-D5 design row; `source-structure.md` "Where graphs show"; `completed.md`'s S12-14 paragraph rewritten for what ships; `design/README.md` S12-D5 line ("board redrawn 2026-10-27"); `DesignSystem/README.md` FlowLayout row ("used by the composer") and PVStroke row, and the `PVStroke.dashed` doc comment ("blocked graph buttons" for "blocked tokens"); the Spike 13 map's S13-D9 row (`docs/deployment-plan/spike-13/deployment-plan.md`: "S12-D5's graph buttons"). | The plan, the map and the shipped list must agree (plan-pr §9). | Leaving token wording in the map. |
| D-27 | The row menu stays on the Source link only (`pvRightClickMenu` with `copySourceEntry`, as today) and its Copy item. | A second `pvRightClickMenu` on the Type and Status cells would add a second set of VoiceOver actions to cells that are not elements; the link already carries Copy… for pointer and VoiceOver. | A menu on the whole row (catches graph-button right-clicks, which the board excludes). |

**Graph button states** (D-11, D-23):

| State | Name line | Detail line | Trailing | Help | VoiceOver label / hint | Identifier | Focus · action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Graph, opens | name, `PVFont.body(size: PVTypeScale.bodySmall)` `textPrimary`, one line, tail truncation | `SourcesGraphCountLine(counts:style: .button)` (D-23) | chevron | `graphButtonHelp(name:detail: graphCounts(_, style: .full))` | `graphButtonLabel(open: openEvidenceGraphLabel(name:), counts: graphCounts(_, style: .spoken, notStarted: .zeroCounts), status: counts.status.map { L10n.string($0.label) })` | `sources.row.<sourceID>.graph.<graph.ref>` | focusable · opens the graph |
| Graph, blocked | name, `bodySmall`, faint | `graphNeedsArtifactToOpen`, `PVFont.body(size: PVTypeScale.micro)` italic `textMuted` | — | `graphButtonHelp(name:detail: reason)` | label the name; hint the reason | same | focusable · no-op |
| No Artifact | `noGraphYet`, `bodySmall`, faint | `needsArtifact`, micro italic `textMuted` | — | `needsArtifactTooltip` | label "`noGraphYet`, `needsArtifact`" (one element); hint `needsArtifactTooltip` | `sources.row.<sourceID>.graphBlocked` | not focusable (not a `Button`) |
| Graph list loading | `PVSkeletonBar(width: 72, height: 6)` | `PVSkeletonBar(width: 132, height: 6)` (spacing `PVSpacing.space4 + PVSpacing.spacePx` between them) | — | — | `L10n.Sources.graphsLoading` | `sources.row.<sourceID>.graphsLoading` | not focusable; frame `borderSubtle`, network glyph faint |
| Graph list failed | `graphsFailedTitle`, `bodySmall` `textSecondary` | `graphsFailedDetail`, micro italic `textMuted` | — | `L10n.Errors.message(for:)` | label "`graphsFailedTitle`, `graphsFailedDetail`"; hint the message (not a value) | `sources.row.<sourceID>.graphsError` | not focusable; `PVIcon(.circleAlert, size: 15)` in `PVColor.danger` in place of the network glyph; frame `borderSubtle` |

`graphButtonLabel` with a nil status (counts loading or unavailable) uses the two-part format. The list-loading state's label and the failed state's copy replace revision 1's `graphsUnavailable`.

## 6. Design

### Codebase fit

| New *(this revision adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcesGraphButton` (renamed from `SourcesGraphToken`) | view | `SourcesGraphToken` | Same view, the board's size, layout and name (D-11). |
| `SourcesListColumns` *(new)* | struct | `PVTable` column definitions | `PVTable` is a selection table; the row shares only widths (D-24). |
| `EvidenceStatus.badgeTone`, `badgeIcon` *(new)* | properties | `EvidenceStatus.label` | The status owns its whole look beside its label (D-22). |
| `PVSymbol.circleDot`, `circleCheck`, `circleAlert` *(new)* | kit icons | `circleDashed`, `success`, `danger` | `success` and `danger` are filled glyphs for callouts; the board draws the outlined Lucide glyphs. |
| `SourcesGraphCountLine.Style`, `countsText` *(new)* | view parameter, helper | `SourcesGraphCountLine` (extended, not copied); `StyledFormat.attributed` | D-23. |
| `NotStartedReading` *(new)* | formatter parameter | `GraphCountsStyle` | D-3a. |
| `SourcesGraphCounts.status`, `.counts(…, filed:)` | property, case field | `EvidenceStatus.init(_:)` | D-3b. |
| `PVListDetailCell` `width: nil`, `iconSize` | kit parameters | `PVListDetailCell` (extended) | D-25. |
| `EvidenceStatus.badge`, `badgeSubtle` *(new)* | view, property | `PVBadge` | D-22. |
| L10n keys (§6 Strings) | L10n | `listCaption`, `graphNeedsArtifact`, `graphsUnavailable`, `graphTokenHelp` | Replaced. |

Searched: `pvMicroCaps`, `a11yJoin`, `PVBadge(`, `circle.`, `chevron.right`, `PVLastUpdated`, `PVListDetailCell`, `onGeometryChange`, `GeometryReader`, `ViewThatFits`, `StyledFormat`, `AttributedString`, `PVTable`, `PVDivider(axis: .vertical)`, `graphCounts(`, `.compact`, `SourcesGraphToken`, `listCaption`.

**Builds on.** Revision 1's session reads and count state; `PVBadge`; `PVLastUpdated.detail`; `PVSkeletonBar`; `PVStroke.dashed`; `PVHoverEffect` and `pvFocusRing` through the existing style.
**Deletes.** `SourcesGraphToken` names (renamed), `PVFlowLayout` use in the row, `SourcesListRow.tokenIndent`, the count line's hand-split "not started" `HStack`, `GraphCountsStyle.compact`, L10n `listCaption`, `graphNeedsArtifact(_:)`, `graphsUnavailable`, `graphTokenHelp(name:detail:)` and their catalog entries, tests `listCaptionMatchesBoard` and `graphTokenCopyMatchesBoard` (replaced).
**Next adds.** S12-15's graph hit opens the location a graph button opens and speaks `openEvidenceGraphLabel`. A status badge elsewhere reads `EvidenceStatus.badgeTone` / `badgeIcon`; a reader outside Sources moves `EvidenceStatus` out of `SourcesModel` first.

### Schema

None — no catalog change.

### Go core

None — revision 1 removed `primary_count` (done).

### FFI

None — done in revision 1.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Row | snowflake | `Features/Sources/SourcesListRow.swift` | D-2, D-12, D-16. |
| Source link | snowflake | same file | D-10, D-25. |
| Columns | snowflake (new) | `Features/Sources/SourcesListColumns.swift` *(new)* | D-24. |
| Status badge | `PVBadge` via `EvidenceStatus.badge` | kit | D-21, D-22. |
| Count line | snowflake (extended) | `Features/Sources/SourcesGraphCountLine.swift` | D-23: the card and the button. |
| Graph button | snowflake (renamed) | `Features/Sources/SourcesGraphButton.swift` *(new, renamed)* | D-11, D-23, states table. |
| Caption | — | `SourcesListView.swift` `captionBand` | D-15. |
| Cover | `CachedThumbnail` | `Platform/CachedThumbnail.swift` | 40. |
| Updated | `PVListDetailCell` + `PVLastUpdated.detail` | kit (extended) | D-25. |
| Row menu | `pvRightClickMenu` | kit | D-27. |
| Status menu | `PVSelect` | `SourcesListView.swift` `statusMenu` | Labels from `EvidenceStatus.label` (D-22). |

**Frames** (board `S2-04 Sources list`, list host 1136 wide):

1. *Interactive list.* Header, count line, toolbar unchanged. Caption SOURCE · TYPE · STATUS · EVIDENCE GRAPH. The 1851 census row: cover, title truncated, "SRC-7KQ2D · ⟲ Today, 3:15 PM"; type "Census schedule"; badges NOT STARTED, IN PROGRESS, PROMOTED stacked; buttons Default (0 · 0), Jones household (12 · 40), Smith household (7 · 19), each with a chevron, behind a vertical hairline. One-graph rows show one badge and one button, the Source and Type centred on it. The obituary (no Artifact): Status "—", dashed "No evidence graph yet / Needs an artifact".
2. *Row states.* 2a hover on the census link (link colour, underline, cover outline, no fill). 2b help on "Yesterday". 2c the dashed "Unlinked / Needs an artifact to open" button with its help. 2d graph lists pending (skeleton block, Status "—"); census counts pending (name kept, a rule for counts; a rule for each badge). 2e list failed block. 2f row menu at the pointer. 2g the copy at the top with one Default graph, NOT STARTED.
3. *Toolbar.* Status menu: All statuses, Not started, In progress, Promoted. Filter, search, nothing-matches, spinner and callout states as today.
4. *Status per graph.* The census under each status with all three buttons and badges.
5. *Minimum width (650).* No Type column; "Census schedule · SRC-7KQ2D · Yesterday" on the second line; Status 120, graphs 240; twelve graphs stacked; the Pryke name truncates.
6–8, 10. Empty destination, page-level states, Add source, cover tile: unchanged.
9. *Metrics.* As D-11, D-16, D-24.
11. *Navigation.* Source link → Source page; graph button → that graph's canvas; blocked inert; nothing expands.
12. *Source page.* Unchanged from revision 1 ("0 subjects · not started").
13. *Building blocks.* As the table above.

### Strings

New *(STEP-6)*: `L10n.Sources.statusBadgeLabel(_ status: EvidenceStatus) -> String` = `sources.list.statusBadge` "Status: %@" ("VoiceOver label of a graph's status badge in the Sources list; argument is the status label"). Changed: `sources.list.status.empty` default "Not started", comment per D-22.

New *(STEP-7)*: `graphNeedsArtifactToOpen` = `sources.list.graphNeedsArtifactToOpen` "Needs an artifact to open" ("Second line of a Sources list graph button whose graph links no Artifact"); `noGraphYet` = `sources.list.noGraphYet` "No evidence graph yet" ("Sources list graph column for a Source with no Artifact"); `graphsLoading` = `sources.list.graphsLoading` "Evidence graphs loading" ("VoiceOver label while a Source's graph list loads"); `graphsFailedTitle` = `sources.list.graphsFailedTitle` "Evidence graphs" and `graphsFailedDetail` = `sources.list.graphsFailedDetail` "Couldn’t be read" ("Sources list graph column when the graph list failed to load"); `graphButtonHelp(name:detail:)` = `sources.list.graphButtonHelp` "%1$@ — %2$@" (renamed from `graphTokenHelp`, same comment with "graph button"); `graphButtonLabel(open:counts:status:)` = `sources.list.graphButtonLabel` "%1$@, %2$@, %3$@" and, for a nil status, `sources.list.graphButtonLabelNoStatus` "%1$@, %2$@" ("VoiceOver label of a Sources list graph button: open label, counts, status"). Changed: `needsArtifact` comment "Second line of the Sources list block for a Source with no Artifact". Removed: `graphNeedsArtifact(_:)`, `graphsUnavailable`, `graphTokenHelp`.

New *(STEP-8)*: `columnSource` = `sources.list.columnSource` "Source", `columnType` = `sources.list.columnType` "Type", `columnStatus` = `sources.list.columnStatus` "Status", `columnEvidenceGraph` = `sources.list.columnEvidenceGraph` "Evidence graph" (each "Sources list column caption"). Removed: `listCaption`. The separators "·" and "—" are `Text(verbatim:)` glyphs (`add-localized-string`).

## 7. Steps

STEP-1 – STEP-5 are done at `37571892` (revision 1: session reads and `QueryRead`; count state and status per graph; `primary_count`; `PVFlowLayout` and `PVStroke`; the Source page count line). Do not redo them. Start at STEP-6. Every check needs a Mac with Xcode.

**STEP-6 — Status badge: look and labels.**
Files: `PVIcon.swift`, `SourcesModel.swift`, `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`. Skills: `add-ui-component`, `add-localized-string`, `add-swift-test`.
Do: D-22; `statusBadgeLabel`; strings per §6 Strings for STEP-6.
Check: T-14 passes; `evidenceStatusFromCounts` and `storedFiledStatusRestoresPromoted` pass.

**STEP-7 — One count fold and one count renderer.**
Files: `SourcesGraphCounts.swift`, `SourcesModel.swift`, `L10n.swift`, `SourcesGraphCountLine.swift`, `SourcePageGraphsView.swift`, `SourcesGraphToken.swift` (its `.counts` patterns only), `SourcesModelTests.swift`, `WorkspaceSessionTests.swift` (patterns only). Skills: `add-swift-test`.
Do: D-3a, D-3b, D-23 for the card style. The Source page looks as before except the failure glyph (`.circleAlert`) and the mono digits of "0 subjects".
Check: T-1 – T-10 pass with only the added `filed: 0`; T-11 passes except its `compact` rows, which stay until STEP-8.

**STEP-8 — The graph button and the four-column row.**
Files: `SourcesGraphToken.swift` → `SourcesGraphButton.swift` (`git mv`), `SourcesListColumns.swift` *(new)*, `SourcesListRow.swift`, `SourcesListView.swift`, `DesignSystem/Components/List/PVList.swift`, `DesignSystem/Recipes/LastUpdated/PVLastUpdated.swift` (doc comment), `L10n.swift`, `Localizable.xcstrings`, `SourcesModelTests.swift`, `project.pbxproj` (the renamed file reference and the new file). Skills: `add-ui-component`, `add-localized-string`.
Do: D-2, D-10, D-11, D-12, D-15, D-16, D-21, D-23 (button style), D-24, D-25, D-27 and the states table; strings per §6 Strings for STEP-7 and STEP-8. Rename `tokens` to `entries`; delete `tokenIndent` and the `PVFlowLayout` use from the row. Reword the comments that still say "token": the `SourcesGraphCounts` header, the `openEvidenceGraphLabel` comment (in `L10n.swift` and the catalog: "VoiceOver label for opening an Evidence graph from the Source page or a Sources list graph button; argument is the graph name"), the `SourcesListNavigation` header ("Pure navigation helpers for the Sources list rows and their graph buttons and the Source page graph cards.").
Check: T-11 (its `compact` rows removed with `case compact`), T-12, T-13 pass; `rg -n 'case compact|style: \.compact' macos/App macos/ProvenenciaTests` returns nothing; `rg -n -i 'token' macos/App/Features/Sources macos/App/Platform/L10n.swift macos/App/Resources/Localizable.xcstrings | rg -i 'graph|sources list'` returns nothing; `rg -n 'SourcesGraphToken|graphTokenHelp|graphsUnavailable|graphNeedsArtifact\(|tokenIndent|listCaption|PVFlowLayout' macos/App/Features/Sources macos/ProvenenciaTests` returns nothing; manual check steps 1–5.

**STEP-9 — Docs.**
Files: `../deployment-plan.md`, `../source-structure.md`, `../completed.md`, `../design/README.md`, `macos/App/DesignSystem/README.md`, `macos/App/DesignSystem/Tokens/PVStroke.swift` (doc comment), `docs/deployment-plan/spike-13/deployment-plan.md` (S13-D9 row), this plan (Landed note at merge only).
Do: D-26. The map row status stays `In review #576` until the PR merges; `completed.md` describes the four-column list.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-14-sources-list-graph-tokens.md` reports no errors; `rg -n -i 'graph tokens?\b|token line|a token' docs/deployment-plan/spike-12/deployment-plan.md docs/deployment-plan/spike-12/source-structure.md docs/deployment-plan/spike-12/completed.md` returns nothing (the plan and brief filenames use hyphens and do not match).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

Revision 1's T-1 – T-10 stay and must pass, changed only by the added `filed: 0` in `.counts(…)` constructions (D-3b). T-11 and T-12 are rewritten; T-13 is replaced; T-14 is new.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesModelTests.swift` `evidenceStatusFromCounts` (kept) | nil → empty; subjects 5, filed 0 → inProgress; filed 1 → promoted; subjects 2, filed 0 → inProgress | AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| T-2 | Swift | `SourcesModelTests.swift` `statusMatchesAnyGraphOfTheSource` (kept) | One Source with three graphs shows under each status | AC-2.5 |
| T-3 | Swift | `SourcesListSettingsTests.swift` `storedFiledStatusRestoresPromoted` (kept) | `"filed"` → `.promoted` | AC-2.7 |
| T-4 | Go | `graphprogress_test.go` `TestGraphProgress`, `TestFiledCounts`; `evidence_graph_progress_test.go` `TestListEvidenceGraphProgress` (kept) | No primary count | AC-4.1 |
| T-5 | Swift | `SourcesListNavigationTests.swift` (kept) | `graphLocationOpensTheGivenGraph`, `pageLocationUsesPageSurface` | AC-1.3 |
| T-6 | Swift | `SourcesModelTests.swift` `listReadsEachGraphsProgress` (kept) | A graph write changes only that graph's progress | AC-1.1, AC-1.9 |
| T-7 | Swift | `SourcePageModelTests.swift` `graphRowsReadTheirProgress` (kept) | The place warms progress | AC-3.1, AC-3.2 |
| T-8 | Swift | `SourcePageModelTests.swift` `aGraphWriteMovesOnlyItsRow` (kept) | Only the written graph's progress moves | AC-3.3 |
| T-9 | Swift | `WorkspaceSessionTests.swift` `evidenceGraphsReadFollowsTheHandle` (kept) | loading / failed / loaded | AC-1.6 |
| T-10 | Swift | `WorkspaceSessionTests.swift` `queryReadAndProgressReads`, `progressFailureReadsUnavailable` (kept) | counts loading; unavailable | AC-1.7 |
| T-11 | Swift | `SourcesModelTests.swift` `graphProgressCopyMatchesBoard` (rewritten) | Revision 1's `full` and `spoken` rows unchanged (the `compact` rows go); with `notStarted: .zeroCounts`: full → "0 subjects · 0 observations", spoken → "0 subjects, 0 observations"; `.counts`, `.loading`, `.unavailable` read as with `.words`. `SourcesGraphCounts.status`: notStarted → `.empty`; counts filed 0 → `.inProgress`; filed 1 → `.promoted`; `.loading`, `.unavailable` → nil. `SourcesGraphCountLine.countsText(.notStarted, style: .card)`: the run whose characters are "0" has `font == PVFont.mono(size: PVTypeScale.micro)`, and the run "not started" has `font == PVFont.body(size: PVTypeScale.micro).italic()` (`countsText` builds italics with exactly that expression) | AC-1.2, AC-1.7, AC-3.1, D-3a, D-3b, D-23 |
| T-12 | Swift | `SourcesModelTests.swift` `graphButtonCopyMatchesBoard` (replaces `graphTokenCopyMatchesBoard`) | `graphButtonLabel(open: "Open graph Jones household", counts: "12 subjects, 40 observations", status: In progress label)` == "Open graph Jones household, 12 subjects, 40 observations, In progress"; nil status → "Open graph Default, counts loading"; `graphButtonHelp(name: "Unlinked", detail: <graphNeedsArtifactTooltip>)` == "Unlinked — Add an artifact to open this evidence graph"; `graphNeedsArtifactToOpen`, `noGraphYet`, `graphsFailedTitle`, `graphsFailedDetail`, `graphsLoading` resolve to the §6 strings | AC-1.2, AC-1.4 – AC-1.6, D-23 |
| T-13 | Swift | `SourcesModelTests.swift` `listColumnsMatchBoard` (replaces `listCaptionMatchesBoard`) | `SourcesListColumns(listWidth: 1136)`: type 150, status 132, graphs 300, not compact; `listWidth: 880` not compact; `listWidth: 879.5` compact, type nil, status 120, graphs 240; the four column captions resolve to "Source", "Type", "Status", "Evidence graph" | AC-1.10, AC-1.12, D-15, D-24 |
| T-14 | Swift | `SourcesModelTests.swift` `statusBadgeMatchesBoard` *(new)* | Labels compared with `L10n.string(status.label)`. `.empty`: "Not started", `.neutral`, `.circleDashed`, subtle; `.inProgress`: "In progress", `.warning`, `.circleDot`; `.promoted`: "Promoted", `.success`, `.circleCheck`; only `.empty` is subtle; `statusBadgeLabel(.inProgress)` == "Status: In progress" | AC-1.2, AC-2.6, D-21, D-22 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `SourcesModelTests`, `SourcesListSettingsTests`, `SourcesListNavigationTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes (T-1 – T-10 unchanged)
- [ ] scripts/check-pr.sh prints check-pr: PASS
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; rg -n 'graphTokenHelp|graphsUnavailable|listCaption|SourcesGraphToken' macos returns nothing
- [ ] UI matches the board; kit changes this revision are the three PVSymbol cases and PVListDetailCell's optional width and icon size
- [ ] Version bump: none
- [ ] deployment-plan.md, source-structure.md, completed.md, design/README.md and DesignSystem/README.md per D-26
- [ ] This plan: at merge, Status → Landed in #576 (date) and a Landed note of every difference from revision 2
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source and three graphs: Default (no Subjects), Jones household (Subjects, none promoted), Smith household (one Subject promoted). The census row shows the caption's four columns; under Status, NOT STARTED, IN PROGRESS, PROMOTED, each level with its button; the buttons read "Default / 0 subjects · 0 observations", "Jones household / 12 subjects · 40 observations", "Smith household / 7 subjects · 19 observations", digits in mono (AC-1.1, AC-1.2, AC-1.12). The title's second line reads the census ref, a dot, the history glyph and its update time; hovering the time shows "Last updated" and the full stamp (AC-1.11). A Source with no Artifact shows "—" and the dashed "No evidence graph yet" block (AC-1.5). Relaunch and watch: a skeleton block and "—" first, then buttons with a rule for counts and rules for badges, then counts and badges (AC-1.6, AC-1.7).
2. Hover the census title slowly, in and out: link colour and one underline, no jump, no second line; cover outline darker; no fill. Click it: the Source page opens (AC-1.8). Back. Click Jones household's button: its canvas opens (AC-1.3).
3. On the census Source page, add a page Artifact, add a graph "Unlinked" linked only to it, then delete that Artifact: back in the list, Unlinked's button is dashed with "Needs an artifact to open", no chevron, does nothing on click, and its help names it (AC-1.4).
4. Add an Observation on Jones's canvas; Back to the list. Only Jones's button moved (AC-1.9).
5. Status filter: Not started, In progress and Promoted each show the census with all its buttons and badges (AC-2.5, AC-2.6). Narrow the window below 880 pt of list: the Type column and caption go and the type joins the second line; with a twelve-graph Source the buttons stack and a long name truncates (AC-1.10). Tab from the census link: the focus ring moves to each graph button in name order, the blocked one included. Right-click the census link: Copy… opens.
6. Open the census Source page: each card shows its count line (AC-3.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Source with many graphs makes a tall row | Buttons stack at 58 pt each (52 + 6); the list is lazy per row; the board stacks twelve in full. |
| `onGeometryChange` fires on every resize | It reports a `SourcesListColumns`, which is `Equatable`, so the rows redraw only when the 880 pt line is crossed. |

**Stop and report if:**

- `onGeometryChange(for:of:action:)` does not compile at the macOS 14 deployment target (D-24).
- `pvRightClickMenu` on the Source link swallows the link's primary click or its VoiceOver action.
- The title underline still shows twice or jumps on hover with D-10's fix in place: report what the hover shows.
- Observation does not repaint a row's button or badge when the progress map changes (D-2).
- Any revision 1 test (T-1 – T-10) has to change beyond the added `filed: 0` to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S12-14 — Graph counts in the Sources list and on the Source page
**Body:** objective in one sentence; what shipped against each US; revision 2 (the board redraw) and why; AC evidence; review summary for both revisions; deviations from the plan; manual check; links to this plan, its review log and the ship log.
