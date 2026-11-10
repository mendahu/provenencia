# S16-35 — Layout tests compute their frames instead of waiting on a window

**Status:** Planned
**Spike / slice:** Spike 16 · Slice 3a — Quiet test runs
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: two layouts move their placement arithmetic into static functions with the same results; every view draws as it does today)
**Board:** —
**Depends on:** S16-34 (landed [#772](https://github.com/mendahu/provenencia/pull/772))
**Blocks:** —
**Branch:** `s16-35-layout-tests-without-waits`
**Version bump:** none — no change a researcher can see in the shipped app ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "(Batch 3) A test run shows no window and touches no install"; the S15-04 ship log on [#774](https://github.com/mendahu/provenencia/pull/774), section "F-5 — Not reproduced on this branch alone" (`main` passes 3 of 5 full runs at load average 15–53); S16-34's ship log (STEP-1 and STEP-5 run tables)
**Planned:** 2026-11-10 against `main` @ `db5885df`
**Review:** [`S16-35-layout-tests-without-waits.review.md`](S16-35-layout-tests-without-waits.review.md) — Fits with changes / Ready with changes, 1 round; no blocker

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

The Swift test suite stops failing at random when the researcher's Mac is busy. Eight tests host a SwiftUI view in a test window and then poll for up to 2 or 15 seconds of wall-clock time until SwiftUI lays it out or loads its data. Layout runs on the main actor, which about 1,900 parallel tests share, so under load a wait runs out and the test fails with nothing wrong. This PR tests the same rules without the waits: the two page layouts' placement becomes a static function tested with made-up sizes; the width-class tests force layout and assert at once; the three navigation tests that mount the whole workspace are deleted, because plain tests already cover their logic, and one plain test covers the leave action they reached.

**Why now.** On 2026-11-10 the full suite failed in 2 of 5 runs on `main` (`918c8ad3`) and in about half the runs on the S15-04 branch, every failure a hosted test timing out (`SourcePageOverviewLayoutTests` ×3, `PVPageColumnsTests.wideHostPlacesTheSecondaryBeside`, `PVWidthClassTests` ×2, `WorkspaceDestinationHostTests` ×3), at load average 15–53. Each false failure costs a rerun of a minute and a wrong diagnosis (S15-04's ship log records one). The researcher asked whether simpler unit and integration tests can replace them (user, 2026-11-10).

**What stays.** What every view draws; `SourcePageOverviewLayout`'s and `PVPageColumnsLayout`'s results for every width; `PVMeasuresWidthClass` and `LeavesForSectionRoot` behavior; `waitUntil` and its 2-second default; `hostInBorderlessWindow` and the `.hostedWindows` trait; the model tests that `waitUntil` on async store calls; `ArtifactPDFViewportPageTests`.

## 2. User stories

**US-1.** As a contributor, when I run the full Swift suite while the Mac is busy, it fails only for a real defect, so I do not rerun it or chase a timeout.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `rg -n "waitUntil" macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift macos/ProvenenciaTests/PVPageColumnsTests.swift macos/ProvenenciaTests/PVWidthClassTests.swift macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift` | prints nothing |
| AC-1.2 | `rg -n -e hostInBorderlessWindow -e hostedWindows macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift macos/ProvenenciaTests/PVPageColumnsTests.swift macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift` | prints nothing |
| AC-1.3 | The full suite, five runs one at a time at the head (D-8), each with `uptime` beside it | No run fails in a test this PR changed; any failing run is recorded with its tests and load, not fixed here |

**US-2.** As a contributor, when I change where `SourcePageOverviewLayout` or `PVPageColumnsLayout` puts a column, a plain test with made-up sizes tells me, the same checks the hosted tests made.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `SourcePageOverviewLayout.arrange(width: 1038, widthClass: .regular, fit:)`; `fit` returns height 30 for index 0 (error), 100 for 1 (body), 140 for 2 (side), 80 for 3 (rest); `frames` is in that index order | error `(0, 0, 1038, 30)`; body and side each `(1038 − space11) / 2` wide, at `y = 30 + space9`, both 140 tall; side at `x = half + space11`; rest at `x = 0`, width 1038, `y = body.maxY + space9` |
| AC-2.2 | The same with error height 0 | body `minY == 0` |
| AC-2.3 | `arrange` at the Wide group width (`PVPageColumnsMetrics.split(available: 2296 − 2 × gutterPage, sideMax: widthSecondaryMax).group`), `.wide` | error, body and rest 1180 wide at `x = 0`; side `x = 1180 + space9`, `y = 0`, width 480; `body.minY == 30 + space9`; `rest.minY == body.maxY + space9` |
| AC-2.4 | `PVPageColumnsLayout.arrange(width: 2296, widths: PVPageColumnsMetrics.layoutWidths(proposal: ProposedViewSize(width: 2296, height: nil), widthClass: .wide, hasSecondary: true, primaryFill: false, fillSplitSideMax: nil), fit:)`; `fit` returns height 100 for index 0 (main) and 100 for 1 (side) | `widths.stacked == false`; main 1180 wide at `x = (2296 − 1692) / 2`, `y = 0`; side 480 wide at `x = main.maxX + space9`, `y = 0`; height 100 |
| AC-2.5 | The same at width 1102, `.regular` | `widths.stacked == true`; main and side 1038 wide at `x = 32`; `side.minY == main.maxY + space9`; height `100 + space9 + 100` |
| AC-2.6 | The same at 2296 `.wide` with `hasSecondary: false` | `widths.side == nil`; `side == nil`; `fit` is never called for index 1; height 100 |

**US-3.** As a contributor, the width-class modifier's tests still exercise SwiftUI's own measuring, without a deadline.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `modifierFollowsTheWindow`: host at 2000 wide, `host.layoutSubtreeIfNeeded()` | `box.seen == .wide` at once; after `window.setContentSize(1200 × 400)` and `layoutSubtreeIfNeeded()`, `.regular` at once |
| AC-3.2 | `aFilledHostMeasuresTheSpaceNotTheChild`: both windows' hosts bound, each laid out with its own `layoutSubtreeIfNeeded()` | the filled host reads `.wide`, the bare probe `.regular`, each straight after its own pass |

**US-4.** As a contributor, a page whose entity is gone still has a plain test of what leaving does.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `attachedNavigation()`, then `go(to: .sectionRoot(.sources))` and `go(to: .evidenceGraph(sourceId: "src-1", graphId: "g-x", title: "Graph"))`; a `WorkspaceSession(projectKey: ProjectKey(projectDir: "/tmp/leave-gone-page.provenencia"), store: FakeStore())`; `LeavesForSectionRoot.leave(notice: "Gone", navigation:, session:)` | `currentLocation == .sectionRoot(.sources)`; no Back or Forward item has `graphId == "g-x"`; `session.noticeToast?.title == "Gone"` |
| AC-4.2 | The same with `notice: nil` | location as AC-4.1; `session.noticeToast == nil` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `SourcePageOverviewLayout.arrange` becomes static over a size closure; its tests compute frames | STEP-2, AC-2.1 – AC-2.3 |
| `PVPageColumnsMetrics.groupX(in:group:)` *(new)* replaces three copies of the centering arithmetic | STEP-1, T-1 |
| `PVPageColumnsMetrics.layoutWidths` *(new)* holds `PVPageColumns.columns`' widths mapping (the stacked rule); `PVPageColumnsLayout.arrange` *(new, static)* holds the arithmetic `sizeThatFits` and `placeSubviews` share; the hosted columns test becomes a plain test of both | STEP-3, AC-2.4 – AC-2.6 |
| `PVWidthClassTests`' two hosted tests force layout and assert at once | STEP-4, AC-3.1, AC-3.2 |
| `LeavesForSectionRoot.leave` *(new)* holds the modifier's action and gets a plain test; `WorkspaceDestinationHostTests`' three hosted tests and their fixtures are deleted | STEP-5, AC-4.1, AC-4.2 |
| `add-swift-test` and `hostInBorderlessWindow`'s doc comment teach the rule (D-7) | STEP-6 |
| Before and after run tables (D-8) | STEP-0, STEP-7, AC-1.3 |

| Out | Home |
| --- | --- |
| `waitUntil`'s 2-second default and every model test that waits on async store work | not planned: those wait on the store's tasks, not layout, and did not fail |
| `ArtifactPDFViewportPageTests`' waits on PDFKit's page-changed notification | not planned: PDFKit posts it, not SwiftUI layout; it did not fail |
| Making `hostInBorderlessWindow` lay out by default | not planned (D-5) |
| A test of `PVPageColumnsBandLayout`'s placement (the header's width and centering) | not planned: its width is `columnWidths(…).band` (covered by `fullWidthModeFillsThePrimaryAtWideOnly`) and its centering is `groupX` (T-1); the private layout's call to them is one of the wirings no test reaches (D-6) |
| A runtime test of `WorkspaceDestinationHost`'s `.id(place?.viewIdentity)` line and `LeavesForSectionRoot`'s `onChange(initial: true)` wiring | not planned (D-6) |
| Test parallelism settings, test plans, CI | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Overview layout | `macos/App/Features/Sources/SourcePageView.swift` `SourcePageOverviewLayout` (line ~369), private `arrange(width:subviews:)` | Computes the four frames and the height from `subviews[index].sizeThatFits`; `guard subviews.count == 4`; Wide centers with `let x = (width - split.group) / 2` (line ~389). `sizeThatFits` and `placeSubviews` both call it |
| Columns layout | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` `PVPageColumnsLayout.sizeThatFits` (line ~377), `placeSubviews` (line ~391) | `sizeThatFits` derives the height (`stacked ? main + space9 + side : max(main, side)`); `placeSubviews` separately derives main at `x = bounds.minX + (bounds.width - w.group) / 2` and the side beside (`x + w.main + space9`) or stacked (`y + main.height + space9`) |
| Columns widths | same file, `PVPageColumns.columns` (line ~252) | A closure maps `columnWidths(…)` to `PVPageSideColumnWidths`, with `stacked: hasSecondary && widthClass != .wide`; only the hosted test runs it |
| Band layout | same file, private `PVPageColumnsBandLayout.placeSubviews` (line ~438) | Centers the header at `bounds.minX + (bounds.width - w.group) / 2` |
| Metrics | same file, `enum PVPageColumnsMetrics` `split`, `columnWidths(proposal:widthClass:hasSecondary:primaryFill:fillSplitSideMax:)` | Pure; covered by `PVPageColumnsTests` (about a dozen plain tests) |
| Width class | `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` private `PVMeasuresWidthClass` | `onGeometryChange` → `@State widthClass` → `.environment(\.pvWidthClass, …)` |
| Leave modifier | `macos/App/Features/Workspace/WorkspaceNavigation.swift` private `LeavesForSectionRoot` (line ~328) | `onChange(of: gone, initial: true)`: when gone, sets `session.noticeToast` when `notice` is set, then `navigation.fallbackToSectionRoot()`. Used by `EvidenceGraphView`, `SourcePageView`, `ConclusionDetailPage` |
| Overview tests | `macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift` | `OverviewProbe` in a 1038- or 2296-wide test window; `host(...)` polls `waitUntil` (2 s) for four `onGeometryChange` frames; three tests |
| Columns test | `macos/ProvenenciaTests/PVPageColumnsTests.swift` `wideHostPlacesTheSecondaryBeside`, `ColumnsProbe`, `FrameBox` | The suite's only hosted test; polls `waitUntil` (2 s) |
| Width-class tests | `macos/ProvenenciaTests/PVWidthClassTests.swift` `modifierFollowsTheWindow`, `aFilledHostMeasuresTheSpaceNotTheChild` | Poll `waitUntil` (2 s) for `box.seen` |
| Host tests | `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift` `canvasToCanvasBuildsTheSecondGraphsModel`, `goneGraphFallsBackToTheSourcesList`, `openGraphThatGoesFallsBackToTheSourcesList`; helpers `makeCanvasStore`, `mountHost`, `navigationOnCanvas`, and the `evidenceGraph(_:sourceID:artifactIDs:)` call | Mount the whole `WorkspaceDestinationHost` on a `FakeStore` and poll `waitUntil` (15 s) for query handles and navigation. The file's other tests (`presentation…`, `sourceFamilyPresentationsUseSourcesDestination`, …) are plain |
| Plain coverage the host tests duplicate | `PlaceRegistryTests.viewIdentityResetsOnTheViewsOwnId`; `EvidenceGraphModelTests` (`graphIsGone`, lines ~1000–1045); `WorkspaceNavigationTests.missingEntityPruneRewritesCurrentEntry` | Distinct `viewIdentity` per graph and Subject; `graphIsGone` from the settled list; `fallbackToSectionRoot` rewrites the current entry and persists it |
| Helper | `macos/ProvenenciaTests/HostedWindow.swift` `hostInBorderlessWindow(_:windowSize:hostSize:)` | Sets the host's frame; never forces layout |
| Skill | `.agents/skills/add-swift-test/SKILL.md` line 16 | Teaches `hostInBorderlessWindow` and the trait; says nothing about waiting on layout |

**Precedent.** Pure placement tested without a view: `PersonTimelinePlacement` (`PersonTimelinePlacementTests`), `PersonFamilyLayout`, `PVSelectPlacement.popupFrame`. Forced layout: `PVSelectPlacementTests` (`cb3d9201`) and `ArtifactRegionOverlayTests` call `host.layoutSubtreeIfNeeded()` straight after hosting and assert with no wait; copy that. A scratch run on 2026-11-10 (not committed) confirmed one `layoutSubtreeIfNeeded()` fires `PVMeasuresWidthClass`'s geometry action, its state change and the probe's `onChange` in the same call: `.wide` after the first pass and `.regular` after `setContentSize(1200 × 400)` and one more pass; an `onGeometryChange` inside `SourcePageOverviewLayout` read its frame after one pass. S16-34's plan is the shape for the before-and-after run tables.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `SourcePageOverviewLayout` gets `static func arrange(width: CGFloat, widthClass: PVWidthClass, fit: (Int, CGFloat) -> CGSize) -> (frames: [CGRect], height: CGFloat)` holding today's body unchanged except the fit source and `groupX` (D-3). `frames` is indexed error 0, body 1, side 2, rest 3, as today. Today's private instance `arrange(width:subviews:)` is renamed `placed(width:subviews:)`: it keeps `guard subviews.count == 4 else { return ([], 0) }` and passes `{ subviews[$0].sizeThatFits(ProposedViewSize(width: $1, height: nil)) }` | `Subviews` cannot be built in a test, which is the only reason the tests host a window; a closure keeps the arithmetic in one place | A protocol over `Subviews` — more machinery for one call; keeping the hosted tests with forced layout — they would still test SwiftUI, not the arithmetic, and need a window |
| D-2 | `PVPageColumnsMetrics.layoutWidths(proposal: ProposedViewSize, widthClass: PVWidthClass, hasSecondary: Bool, primaryFill: Bool, fillSplitSideMax: CGFloat?) -> PVPageSideColumnWidths` *(new)* is the body of `PVPageColumns.columns`' `widths` closure (stacked = `hasSecondary && widthClass != .wide`); the closure calls it. `PVPageColumnsLayout` gets `static func arrange(width: CGFloat, widths: PVPageSideColumnWidths, fit: (Int, CGFloat) -> CGSize) -> (main: CGRect, side: CGRect?, height: CGFloat)`, frames relative to the layout's origin; it calls `fit(1, …)` only when `widths.side != nil`. `sizeThatFits` returns `CGSize(width: pane(proposal, w), height: arrange(…).height)` with `w = widths(proposal)`; `placeSubviews` calls it with `bounds.width` and `widths(ProposedViewSize(width: bounds.width, height: nil))` as today and offsets by `bounds.minX` / `bounds.minY`. Both keep `guard !subviews.isEmpty` | One function holds the stacked-or-beside geometry that `sizeThatFits` and `placeSubviews` derive separately today, and the same shape and name as D-1; the stacked rule is tested where it lives | `frames(width:widths:main:side:)` placing only — leaves the height's copy of the stacked rule and admits a `side` size that contradicts `widths.side`; building the widths in the test — a copy of the stacked rule the production code could drift from |
| D-3 | `PVPageColumnsMetrics.groupX(in width: CGFloat, group: CGFloat) -> CGFloat` *(new)* returns `(width - group) / 2`. `PVPageColumnsLayout.frames`, `PVPageColumnsBandLayout.placeSubviews` and `SourcePageOverviewLayout.arrange` use it in place of their three copies | One centering rule for the page band, header and Source page; `groupX` is tested once (T-1) | Leaving the copies — three formulas for one rule; naming the parameter `pane` — the Overview passes its own width, which is not the pane |
| D-4 | `PVWidthClassTests`' two hosted tests stay hosted: they become synchronous (no `async`), bind every host from `hostInBorderlessWindow`'s tuple, call `layoutSubtreeIfNeeded()` on each host after hosting it and after `window.setContentSize`, and `#expect` at once. The suite keeps `.hostedWindows` | They test SwiftUI's own geometry and environment pass, which only a hosted view has; the scratch run shows one forced pass settles it | Deleting them — `PVMeasuresWidthClass` is applied once for the whole app and has no other test |
| D-5 | `hostInBorderlessWindow` does not change. A test that needs layout calls `host.layoutSubtreeIfNeeded()` itself, as `PVSelectPlacementTests` does | Callers that host for `.task` or AppKit behavior keep today's timing; the precedent already reads this way | Laying out inside the helper — changes the timing of every hosted suite at once for no failing test |
| D-6 | `WorkspaceDestinationHostTests`' three hosted tests, `makeCanvasStore`, `mountHost` and `navigationOnCanvas` are deleted; the file keeps `projectDir`, `presentation(for:)` and every plain test. Their logic stays covered by the plain tests in §4 plus T-5. `LeavesForSectionRoot` changes from `private struct` to `struct` and gains `static func leave(notice: String?, navigation: WorkspaceNavigation, session: WorkspaceSession)` (the body of today's `onChange` closure); the closure calls `Self.leave(notice: notice, navigation: navigation, session: session)`. What no test reaches after this PR, each a line or two that review sees: `WorkspaceDestinationHost`'s `.id(place?.viewIdentity)`, the modifier's `onChange(initial: true)` wiring, and `PVPageColumnsBandLayout`'s calls to `columnWidths(…).band` and `groupX` | Each hosted test mounts the whole workspace and waits up to 15 s on async loads; the logic they reach is already tested plainly; the researcher asked for low-cost tests (working rule, Batch 3) | Keeping one hosted canvas test as a smoke test — it carries the same 15 s wait and failed under load on `main`; extracting the `.id` wiring — no seam smaller than the view itself |
| D-7 | In `add-swift-test`, the hosted-window sentences move out of the "Artifact files on disk" bullet (line 16) into a bullet of their own, which opens with the rule: test a `Layout`'s placement through a `static func arrange(width:…fit:)` with made-up sizes (examples: `SourcePageOverviewLayout.arrange`, `PVPageColumnsLayout.arrange`, `PersonTimelinePlacement`), not a hosted window; host a view only for SwiftUI's own behavior (environment, geometry modifiers, view identity), then call `layoutSubtreeIfNeeded()` on the host and assert at once; do not `waitUntil` on layout. The rest of today's hosted-window text (the helpers, the trait, no synthesized events) follows unchanged. `hostInBorderlessWindow`'s doc comment gains: the helper does not force layout; call `layoutSubtreeIfNeeded()` on the returned host, then assert | One place teaches it, and the helper's own doc agrees | Only a comment in `HostedWindow.swift` — the skill is what the implementer loads; leaving the bullet's opening — it tells the next author to host a layout |
| D-8 | STEP-0 runs the full suite five times on the base and STEP-7 five times at the head, one at a time, with `NSUnbufferedIO=YES xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` (the body of `xcode_test` in `scripts/check-pr.sh`), `uptime` printed before each. Each run's result, duration, failing tests and load averages go in a table in `$TMPDIR/s16-35-runs.md`, copied into the PR body. STEP-7's `scripts/check-pr.sh` run is separate and not one of the five. A failing run in a test this PR did not change is recorded, not fixed | The claim is "fewer false failures"; S16-34's tables are the shape; never two runs at once (working rule) | One run each — the failures are load-dependent and one run proves nothing |

**Rules this plan bends.** None. The Batch 3 working rule ("Unit and integration tests stay low-cost") is the reason for D-6; S16-34's D-11 ("changes no test's timeout, wait or assertion") bound only S16-34.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `SourcePageOverviewLayout.arrange(width:widthClass:fit:)` *(new, static)* | func | the private instance `arrange(width:subviews:)` | It is that body made static; the instance method stays as the `Subviews` adapter |
| `PVPageColumnsLayout.arrange(width:widths:fit:)` *(new, static)* | func | `PVPageColumnsLayout.sizeThatFits` and `placeSubviews` | Their two derivations of the stacked-or-beside geometry move into it; both call it |
| `PVPageColumnsMetrics.layoutWidths(proposal:widthClass:hasSecondary:primaryFill:fillSplitSideMax:)` *(new)* | func | the `widths` closure in `PVPageColumns.columns`; `PVPageColumnsMetrics.sideColumnWidths(proposal:mainMin:)` (a different rule, for Person sources) | It is that closure, moved where the pure page arithmetic lives |
| `PVPageColumnsMetrics.groupX(in:group:)` *(new)* | func | three inline `(… - group) / 2` copies (searched `rg "- w\.group\) / 2|- split\.group\) / 2"`) | Replaces them |
| `LeavesForSectionRoot.leave(notice:navigation:session:)` *(new, static)* | func | the modifier's `onChange` closure; `WorkspaceNavigation.fallbackToSectionRoot()` | It is that closure; `fallbackToSectionRoot` has no notice or session |
| `PVPageColumnsTests.columnsLayoutPlacesTheSecondaryBesideAtWideAndBelowAtRegular`, `groupXCentersTheGroup`, `WorkspaceNavigationTests.leavingAGonePageShowsTheNoticeAndRewritesTheEntry` *(new)* | tests | the hosted tests they replace | — |

**Builds on.** `PVPageColumnsMetrics` (pure page arithmetic already lives there); the pure-placement pattern of `PersonTimelinePlacement`, `PersonFamilyLayout` and `PVSelectPlacement`; the `PVSelectPlacementTests` forced-layout pattern; the `.hostedWindows` trait (S16-34).
**Deletes.** `OverviewFrames`, `OverviewProbe` and `host(width:widthClass:errorHeight:)` in `SourcePageOverviewLayoutTests.swift`; `FrameBox`, `ColumnsProbe` and `wideHostPlacesTheSecondaryBeside` in `PVPageColumnsTests.swift`; `canvasToCanvasBuildsTheSecondGraphsModel`, `goneGraphFallsBackToTheSourcesList`, `openGraphThatGoesFallsBackToTheSourcesList`, `makeCanvasStore`, `mountHost`, `navigationOnCanvas` in `WorkspaceDestinationHostTests.swift`; the `.hostedWindows` trait and `async` on those three suites where no test hosts any more; the three inline centering copies; the `widths` closure body in `PVPageColumns.columns`; the separate height derivation in `PVPageColumnsLayout.sizeThatFits`. The `.hostedWindows` trait and `async` on `SourcePageOverviewLayoutTests`, `PVPageColumnsTests` and `WorkspaceDestinationHostTests` (none hosts afterwards). `import AppKit` in each of those three files once nothing in it uses AppKit (`WorkspaceDestinationHostTests` keeps `Foundation`).
**Next adds.** A new page `Layout` adds a static arrange and plain tests, per D-7.

### Reach

None — no new path into existing code, no new state reachable.

### Scale

None — no write, list or loop changes; the layouts do the same arithmetic per layout pass.

### Edges

- **Assumptions.** The static functions must return exactly what the hosted tests saw. They read the same sizes `sizeThatFits` returned, so they do if the extraction moves code without changing it. T-2 – T-4 assert the hosted tests' own numbers; Stop if a number differs.
- **Neighbors.** The Overview's `subviews.count != 4` guard stays in `placed(width:subviews:)` (empty frames, height 0); `PVPageColumnsLayout.arrange` with `widths.side == nil` returns `side: nil` and the main height, as today's `guard let sideWidth = w.side` (AC-2.6).
- **Removed inputs.** Deleting the hosted tests removes runtime coverage of the wirings D-6 names; each is named in the PR body. The stacked rule keeps a test through `layoutWidths` (AC-2.4, AC-2.5).
- **Real data.** None — no table or seed.
- **Untrusted size.** None — no data the app did not write.

### Schema

None — no catalog change.

### Go core

None — Swift only.

### FFI

None —

### macOS

None — no view changes what it draws. Code changes: `SourcePageView.swift` (`SourcePageOverviewLayout`), `PVPageColumns.swift` (`PVPageColumnsMetrics`, `PVPageColumnsLayout`, `PVPageColumnsBandLayout`), `WorkspaceNavigation.swift` (`LeavesForSectionRoot`).

### Strings

None — no copy.

## 7. Steps

**STEP-0 — Before.**
Do: on the base, run the full suite five times one at a time (D-8), writing the table to `$TMPDIR/s16-35-runs.md`.
Check: five rows recorded.

**STEP-1 — One centering rule.** Skill: `add-swift-test`.
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/App/Features/Sources/SourcePageView.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`.
Do: D-3. Add T-1.
Check: `rg -n -e '- w.group) / 2' -e '- split.group) / 2' -F macos/App` prints nothing; T-1 and the existing `PVPageColumnsTests` and `SourcePageOverviewLayoutTests` pass.

**STEP-2 — Overview arithmetic.**
Files: `macos/App/Features/Sources/SourcePageView.swift`, `macos/ProvenenciaTests/SourcePageOverviewLayoutTests.swift`.
Do: D-1. Rewrite the three tests as T-2 (same names, same assertions, synchronous, a `fit` closure returning `CGSize(width: w, height: [errorHeight, 100, 140, 80][index])`); delete the probe, the frames box, `host(...)`, the suite's `.hostedWindows` trait and `async`.
Check: T-2 passes; AC-1.1 and AC-1.2 for this file.

**STEP-3 — Columns arithmetic.**
Files: `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift`, `macos/ProvenenciaTests/PVPageColumnsTests.swift`.
Do: D-2. Replace `wideHostPlacesTheSecondaryBeside`, `ColumnsProbe` and `FrameBox` with T-3; drop the suite's `.hostedWindows` trait.
Check: T-3 and the rest of `PVPageColumnsTests` pass; AC-1.1, AC-1.2 for this file.

**STEP-4 — Width class without waiting.**
Files: `macos/ProvenenciaTests/PVWidthClassTests.swift`.
Do: D-4 (T-4).
Check: T-4 passes; AC-1.1 for this file.

**STEP-5 — Leave action and the host tests.**
Files: `macos/App/Features/Workspace/WorkspaceNavigation.swift`, `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`, `macos/ProvenenciaTests/WorkspaceDestinationHostTests.swift`.
Do: D-6. Add T-5. Delete the three hosted tests, their helpers and the suite's `.hostedWindows` trait; keep every plain test in the file.
Check: T-5 passes; `WorkspaceDestinationHostTests`, `PlaceRegistryTests`, `EvidenceGraphModelTests` and `WorkspaceNavigationTests` pass; AC-1.1, AC-1.2 for this file.

**STEP-6 — Teach it.**
Files: `.agents/skills/add-swift-test/SKILL.md`, `macos/ProvenenciaTests/HostedWindow.swift`.
Do: D-7.
Check: `python3 scripts/check-guidance-names.py` passes.

**STEP-7 — After.**
Do: D-8 at the head; then `scripts/check-pr.sh`.
Check: five rows recorded; AC-1.3; `check-pr: PASS`.

**STEP-8 — Docs**
Files: `../completed.md`, `../deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-35-layout-tests-without-waits.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed, ending in a period.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `PVPageColumnsTests.swift` `groupXCentersTheGroup` *(new)* | `(2296, 1692)` → 302; `(1102, 1038)` → 32; `(1038, 1038)` → 0 | D-3 |
| T-2 | Swift | `SourcePageOverviewLayoutTests.swift` `regularPlacesDescriptionBesideMetadataThenTheRest`, `anEmptyErrorAddsNoGap`, `widePlacesMetadataInASideColumn` *(rewritten)* | AC-2.1, AC-2.2, AC-2.3 | US-2, D-1 |
| T-3 | Swift | `PVPageColumnsTests.swift` `columnsLayoutPlacesTheSecondaryBesideAtWideAndBelowAtRegular` *(new, one test, three cases)* | AC-2.4, AC-2.5, AC-2.6 | US-2, D-2 |
| T-4 | Swift | `PVWidthClassTests.swift` `modifierFollowsTheWindow`, `aFilledHostMeasuresTheSpaceNotTheChild` *(rewritten)* | AC-3.1, AC-3.2 | US-3, D-4 |
| T-5 | Swift | `WorkspaceNavigationTests.swift` `leavingAGonePageShowsTheNoticeAndRewritesTheEntry` *(new)*, on `attachedNavigation()` and a `WorkspaceSession` over a `FakeStore` | AC-4.1, AC-4.2 | US-4, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `PVPageColumnsTests`, `SourcePageOverviewLayoutTests`, `WorkspaceNavigationTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or command output for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] Sweep: each deleted name (`wideHostPlacesTheSecondaryBeside`, `canvasToCanvasBuildsTheSecondGraphsModel`, `goneGraphFallsBackToTheSourcesList`, `openGraphThatGoesFallsBackToTheSourcesList`, `OverviewProbe`, `ColumnsProbe`) `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` with no hit left except ship logs and plans; `scripts/check-guidance-names.py` passes
- [ ] The before and after run tables (D-8) are in the PR body, each run with its load averages
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] PR description: what shipped, AC evidence, the two wirings no test reaches (D-6), deviations
```

## 10. Manual check

1. Run the `rg` commands in AC-1.1 and AC-1.2: both print nothing.
2. Run the full suite five times one at a time with a browser and other apps open (STEP-7, AC-1.3). No run fails in `SourcePageOverviewLayoutTests`, `PVPageColumnsTests`, `PVWidthClassTests` or `WorkspaceDestinationHostTests`.
3. Open a Source page at Regular and at Wide and a Person page at Wide: columns, header and side column sit where they did before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The extraction changes a frame | T-2 and T-3 keep the hosted tests' own numbers; manual check 3 |
| Forced layout does not settle the width class inside the full suite | T-4 runs in all five STEP-7 runs; Stop condition below |
| Deleting the host tests hides a wiring regression | D-6 names the two lines; the PR body lists them for review |

**Stop and report if:**

- A rewritten T-2 or T-3 needs a number different from the hosted test it replaces.
- T-4 fails in any STEP-7 run after one `layoutSubtreeIfNeeded()` per size change.
- A test outside the four suites fails in two or more of the five STEP-7 runs and passed all five STEP-0 runs.
- A second type named `LeavesForSectionRoot` exists in the app target.

## 12. Open questions

None.

## 13. PR

**Title:** S16-35 — Layout tests compute their frames instead of waiting on a window
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; the before and after run tables; the two wirings no test reaches; deviations; `Plan: docs/deployment-plan/spike-16/prs/S16-35-layout-tests-without-waits.md`.
