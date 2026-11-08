# Review — S13-28 How the records connect

Plan: [`S13-28-how-the-records-connect.md`](S13-28-how-the-records-connect.md). Brief log: [`../design/S13-D10-how-the-records-connect.review.md`](../design/archive/S13-D10-how-the-records-connect.review.md). Lint clean but for the missing review logs before round 1. The plan is Blocked on the board; "(board)" values were out of scope for both reviewers.

## Round 1 (2026-11-07)

Maintainability (opus): **Fits with changes**. Executability (sonnet): **Not ready**.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | `supportingQueryKeys` is warm-only vocabulary (its doc, `add-workspace-place` line 90); widening it bends a skill and makes a vocabulary reload show Refreshing | **Fixed.** D-2: `companionQueryKey` slot; `supportingQueryKeys` unchanged; STEP-9 updates the skill and `macos-client-patterns.md` |
| M-2 | blocker | A section-private width split and centring beside `PVPageColumnsMetrics.columnWidths`; Event and Place would copy it | **Fixed.** D-3: `PVPageColumnsMetrics.sideColumnWidths` and `ConclusionPinnedSideColumn` (feature-level until a second caller; kit placement Out) |
| M-3 | major | Highlight and focus have two owners; the focus enum cannot reach the diagram | **Fixed.** D-6: `PersonSourcesList` draws both columns and owns focus, hover and the derived highlight, with precedence |
| M-4 | major | Graph locations built from claims a second time | **Fixed.** D-4: `PersonSourcesContent.graphLocations`, built in `init` before the fold |
| M-5 | major | The diagram column re-implements the read-to-state mapping | **Fixed.** D-8: `ConclusionDetailPresentation`, `.gone` as the host shows it |
| M-6 | major | A second hand-built viewport reader | **Fixed.** D-7: `pvPageVisibleRect`; the Timeline moves onto it; `verticalPadding` constant |
| M-7 | major | Docs and skills made false; no § 9 pattern row | **Fixed.** STEP-9 and the DoD sweep |
| M-8 | minor | "Person" in generic geometry names | **Fixed.** `RecordLinksLayout`, `RecordLinksDiagram`; `PersonRecordLinksContent` stays the adapter |
| M-9 | minor | Name the node's kit parts | **Fixed.** D-5: `PVHoverEffect`, `pvFocusRing`, `pvActivatesOnReturn`; no cursor (Out) |
| M-10 | minor | Family is the closer precedent; T-6 in the wrong file | **Fixed.** Precedent paragraph; width tests in `PVPageColumnsTests` (T-3) |
| X-1 | blocker | The host's states and the second column are not reconciled | **Fixed.** D-8, AC-5.3: the diagram column exists only with the rows' content; the host's states stay full width (brief RL-14) |
| X-2 | blocker | The diagram cannot reach the claim groups or the focus state | **Fixed.** D-6 (one owner), D-4 (`graphLocations`), STEP-8 (the section passes the companion read's presentation to `PersonSourcesList`) |
| X-3 | major | How the section learns its width | **Fixed.** D-3: `onGeometryChange` on width only; unmeasured → main takes the width, no side |
| X-4 | major | The brief's 380pt breaks the 200pt citation stop | **Fixed.** D-3 caps `sideRegular` at 364pt; brief RL-W2 recommends 340, at most 364 |
| X-5 | major | Neighbour set, arc states, precedence | **Fixed.** D-4 `neighbors(of:)`; D-6 precedence; a line with no node highlights only itself |
| X-6 | major | Node geometry and edge cases | **Fixed.** D-5 marks mark size, label room, offset and spread as (board); 0, 1, 2 nodes; D-4 duplicates and unknown pairs |
| X-7 | major | STEP order: previews before the states exist | **Fixed.** Section previews moved to STEP-8 |
| X-8 | major | T-7 cited but missing; fixtures unspecified | **Fixed.** Scale no longer cites T-7; D-10 fixtures with counts |
| X-9 | major | T-6 location and the split contract | **Fixed.** D-3 states the contract; T-3 in `PVPageColumnsTests` |
| X-10 | major | Pin: which view reads, clamp order, zero viewport, offset hit-testing | **Fixed.** D-7; Risks row on `offset` (Family precedent) and manual step 4 |
| X-11 | minor | Resize across 1600 not checked | **Fixed.** AC-3.4, manual step 4 |
| X-12 | minor | Refreshing would follow the types list | **Fixed** by M-1 (AC-5.5) |
| X-13 | minor | The `band above` sweep | **Fixed.** Replaced by a sweep of the lines this PR makes false |
| X-14 | minor | Review logs missing | **Fixed.** This log and the brief's |
| X-15 | minor | Temptations: line hover over Unpromote, node help string | **Fixed.** AC-2.2 (whole line); `recordLinksNodeHelp` in D-9; cursor Out |
| X-16 | — | No import or lock issue | — |

Round 2 runs because round 1 had blockers.

## Round 2 (2026-11-07)

Maintainability (opus): **Fits with changes**, two new blockers. Executability (sonnet): **Not ready**, two new blockers.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-11 | blocker | The diagram column copies the host's read-to-state switch (nil handle, starting the read, `.gone` copy) | **Fixed.** D-8: `ConclusionPageSectionHost` gains `key` and `placement`; the diagram column is a second host; STEP-9 rewords `macos-client-patterns.md` line 104 |
| M-12 | blocker | The reusable diagram cannot join the list's focus without the Person enum | **Fixed.** D-5: `focus: (String) -> PVFieldFocus`, built from the list's one `FocusState` |
| X-17 | blocker | No rule for a narrow pane (the 700pt window) | **Board question.** Brief RL-W5 (recommended: stack under the rows below a 900pt pane, unpinned; the user's "same at every width" makes it the user's call on the board); plan D-3, AC-3.5, T-3 row, §12 |
| X-18 | blocker | The graph lines do not fit the narrowed citation column | **Board question.** Brief RL-W2 now asks for the lines at that width (recommended: refs and name on a second caption line); plan D-3, STEP-8, §12 |
| X-19 | major | Content needs strings added later | **Fixed.** Strings are STEP-4, before the content |
| X-20 | major | Who builds the diagram's content and starts its read | **Fixed** by M-11: the second host starts the read; its `make` captures `graphLocations` |
| X-21 | major | Fixtures do not share graph ids; existing previews do not render the host | **Fixed.** D-10 `johannClaims`, `manyClaims`; STEP-8 previews; manual step 6 reworded |
| X-22 | major | `verticalPadding` on one of two literals | **Fixed.** Lines 139 and 153 |
| X-23 | minor | The sweep hits history | **Fixed.** Scoped to the live guides |
| X-24 | minor | `pinOffset` home and height reading | **Fixed.** D-7 |
| X-25 | minor | Style, label cut, line highlight padding, title in the side height, key order, D-4 semantics | **Fixed** but the highlight's padding and the label cut, which are board values (RL-1, RL-4) |
| X-26 | nit | One unmeasured frame | **Fixed.** Stated in D-7 |
| M (minors) | minor | `verticalPadding` both lines; `pinOffset` home; Next adds wording; the visible-rect reader holding its state | **Fixed.** D-7 `PVPageVisibleRect` holds the state; Next adds reworded |

## Refresh after the board (2026-11-07)

The S13-D10 board was agreed (turn 15). The planner checked it against the brief and the code, asked the user four questions (header at Wide: keep it centred; the 1078pt breakpoint: accepted; the narrow rows: an idea file, `docs/ideas/wide-layout.md` § Narrow panes; the board reaches the implementer by its artifact link), and copied the values into D-3, D-5, D-7, D-9, §6 and the tests. Then one round with both reviewers.

Maintainability (opus): **Fits with changes**, three blockers. Executability (sonnet): **Not ready**, one blocker.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-27 | blocker | The graph lines sit in the citation cell (204pt at Regular), not under the whole row as the board's note says | **Fixed.** D-3: the lines move out of the citation cell to span the row, as the board draws them; the row becomes a `VStack` of the cells and the lines |
| M-13 | blocker | `sideColumnWidths` repeats `columnWidths`' Wide split; `widthClass` changes nothing | **Fixed.** D-3: `split(available:sideMax:)` taken out of `columnWidths`; `sideColumnWidths(available:mainMin:)` adds only the stack |
| M-14 | blocker | An `HStack`/`VStack` branch drops state at 1078pt; the kit places columns with a `Layout` | **Fixed.** D-3: `ConclusionSideColumnLayout`, the `PVPageColumnsLayout` shape; stacked gap `space9` (Rules this plan bends). Also answers X-35 |
| M-15 | blocker | Pinning decided by the diagram's layout but applied by the column | **Fixed.** D-7: `pinOffset` returns 0 when the side is taller than the visible area; `diameter` returns only a size |
| X-28 | major | Line hover: width, shape, padding, API | **Fixed.** D-6 |
| X-29 | major | Previews have no session; FakeStore cannot hold a read loading or failed | **Fixed.** D-11 |
| X-30 | major | 1366 × 600 does not reach the unpin by the stated formula | **Fixed.** D-7 names the visible height (the viewport under the pinned header and switcher, about 290pt there); AC-3.3 split into Wide shrink and Regular unpin |
| X-31 | minor | T-5 two-arc spacing | **Fixed.** 1 × spacing apart |
| X-32 | minor | Empty name or title in help and accessibility strings | **Fixed.** D-9 fallbacks |
| X-33 | minor | Focus mapping for the highlight; Unpromote focus | **Fixed.** D-6 `claimGraphs`; Unpromote focus counts |
| X-34 | minor | Second host's empty arguments, ids, `key` name, error title, spinner | **Fixed.** D-8 |
| X-35 | minor | Identity across the stack switch | **Fixed** by M-14 |
| M-16 | minor | Node hover: two trackers; nearest styles not named | **Fixed.** `PVHoverEffect.onHoverChange`; Codebase fit rows |
| M-17 | minor | 720 outside the width tokens | **Fixed.** `PVSpacing.widthSideColumnMax` |
| M-18 | minor | 200pt citation minimum bare | **Fixed.** `PersonSourcesColumns.citationMin`, `minWidth` |
| M-19, X-36, X-37 | nit | "chip" wording; frame id ranges; `graphLocations` loop shape | **Fixed** (first two); the loop is the implementer's |

### Refresh, round 2 (2026-11-07)

Maintainability (opus): **Fits with changes**, two new blockers. Executability (sonnet): **Not ready**, two new blockers. Every refresh round-1 answer was confirmed against the code.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-20 | blocker | `ConclusionSideColumnLayout` copies `PVPageColumnsLayout` | **Fixed.** D-3: the kit's Layout is made internal and takes its widths as an input; the side column is its second caller |
| M-21 | blocker | D-7 still pins inside an `HStack` and reads the section's height twice | **Fixed.** D-7: `PVPageVisibleRect` hands its content its size too; only `sideHeight` is read apart |
| X-38 | blocker | The 50-graph fixture gives 89 arcs, not 91 | **Fixed.** D-10: graphs 48 – 50 apart (97 − 6 = 91) |
| X-39 | blocker | No data for the Source-ref fallback | **Fixed.** D-6: `graphSourceRefs` from `group.sourceRef`; T-6 |
| X-40 | major | `PersonSourcesList` has no `entityId` | **Fixed.** STEP-8 |
| X-41 | minor | Placing the dot on the node point | **Fixed.** D-5: measured offset |
| X-42, M (nits) | nit | T-6 step; `PVPageVisibleRect` case; FakeStore hook naming; "layout unchanged"; 646pt wording; empty defaults on the host | **Fixed**, except the host's empty defaults, kept: only `.column` omits them, and a band section without an empty state is caught in review |

## Verdict

After the refresh (two rounds): no blocker open, no To user item open. The board's values are in the plan; the user answered the header, breakpoint, narrow-rows and board-link questions. Status **Planned**.
