# Review — S16-35 Layout tests compute their frames instead of waiting on a window

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1 (no blocker, no reshape, so no round 2). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
None.

## Writer's own calls
- D-6 deletes `WorkspaceDestinationHostTests`' three whole-workspace tests instead of keeping one as a smoke test; three wirings lose runtime coverage (`.id(place?.viewIdentity)`, `onChange(initial: true)`, the band layout's calls) and are named in the PR body. Prompted by the user's question "Is this test necessary? Could we solve it with simpler unit and integration tests?" (2026-11-10) and the Batch 3 low-cost rule.
- D-5 leaves `hostInBorderlessWindow` unchanged; tests force layout themselves.
- D-8 measures five runs before and after, one at a time, with load averages.
- The forced-layout claim rests on a scratch run (2026-11-10, not committed): one `layoutSubtreeIfNeeded()` gave `.wide`, then `.regular` after a resize; an `onGeometryChange` inside `SourcePageOverviewLayout` read its frame after one pass. T-4 and a Stop condition hold it.

## Round 1

### M-1 · major · T-3 rebuilds the page-columns widths mapping in the test; the production copy loses its only test
The `stacked: hasSecondary && widthClass != .wide` mapping lives only in `PVPageColumns.columns`' closure; deleting the hosted test leaves it untested while T-3 holds a copy.
**Answer:** Fixed — D-2 adds `PVPageColumnsMetrics.layoutWidths(…)`, which the closure calls and T-3 uses (AC-2.4, AC-2.5 assert `stacked`); Current state gains the Columns widths row; Codebase fit and Deletes list it.

### M-2 · major · `frames` takes a different shape from D-1's `arrange`; `sizeThatFits` keeps its own copy of the stacked rule
**Answer:** Fixed — D-2 is now `PVPageColumnsLayout.arrange(width:widths:fit:) -> (main, side, height)`, used by both `sizeThatFits` and `placeSubviews`; `fit(1, …)` only when `widths.side != nil`; AC-2.6 covers no side.

### M-3 · minor · One shape gets two names; the skill line names no shape or precedent
**Answer:** Fixed — both statics are `arrange`; D-7 names `SourcePageOverviewLayout.arrange`, `PVPageColumnsLayout.arrange` and `PersonTimelinePlacement`; Precedent and Builds on cite `PersonTimelinePlacement`, `PersonFamilyLayout`, `PVSelectPlacement`.

### M-4 · minor · The Overview's Wide branch keeps a second formula for "side column beside main"
**Answer:** Declined — the Overview's main column is three stacked rows placed by its own loop, and reaching `PVPageColumnsLayout.arrange` would need a synthetic main size built from them; the two side formulas agree by `PVPageColumnsMetrics.split`'s definition (`group = main + space9 + side`), which `PVPageColumnsTests` pins, and AC-2.3 pins the Overview's side x. The reviewer allowed declining if the indirection reads worse; it does.

### M-5 · minor · The forced-layout rule sits in the wrong bullet; `HostedWindow.swift`'s doc implies layout just happens
**Answer:** Fixed — D-7 moves the hosted-window text to its own bullet opening with the rule, and adds a line to `hostInBorderlessWindow`'s doc comment; STEP-6 names both files.

### M-6 · nit · `groupX(pane:group:)` is passed a value that is not a pane
**Answer:** Fixed — `groupX(in:group:)`.

### M-7 · nit · Two functions named `arrange` in one type
**Answer:** Fixed — the instance adapter is `placed(width:subviews:)`.

### X-1 · major · AC-1.2's `rg` uses `\|`, a literal pipe
**Answer:** Fixed — AC-1.2 uses `-e hostInBorderlessWindow -e hostedWindows`; STEP-1's check uses `-F -e`.

### X-2 · minor · The full-suite command is a shell function
**Answer:** Fixed — D-8 spells out the `xcodebuild` line, `uptime` before each run, the table at `$TMPDIR/s16-35-runs.md`, and that STEP-7's `check-pr.sh` run is not one of the five.

### X-3 · minor · T-3 and T-5 leave inputs open
**Answer:** Fixed — AC-2.4 – AC-2.6 give the fit heights (100, 100); AC-4.1 gives `.evidenceGraph(sourceId: "src-1", graphId: "g-x", title: "Graph")` and the `projectDir` literal; T-3 is one test with three cases.

### X-4 · minor · D-3's claim that the band header stays covered is not true
**Answer:** Fixed — the claim is gone; D-6 and the Out table name the band layout's calls as wiring no test reaches.

### X-5 · minor · Trait handling is hedged
**Answer:** Fixed — STEP-2, STEP-3, STEP-5 and Deletes drop the trait on all three suites; `WorkspaceDestinationHostTests` keeps `Foundation`, `projectDir` and its plain tests.

### X-6 · minor · Stop condition 4 is not checkable
**Answer:** Fixed — "A second type named `LeavesForSectionRoot` exists in the app target"; D-6 says `private struct` → `struct` and the closure calls `Self.leave(…)`.

### X-7 · minor · Two windows in T-4; frame index order
**Answer:** Fixed — D-4 and AC-3.2 bind and lay out every host; D-1 and AC-2.1 state the index order.
