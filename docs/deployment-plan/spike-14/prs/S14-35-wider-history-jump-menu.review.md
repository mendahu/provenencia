# Review — S14-35 Wider history jump menu

**Verdict:** Fits with changes — Ready with changes (round 1)
**Rounds:** 1 (round 2 pending). **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None.

## Writer's own calls
- **D-1:** the preferred width is 600pt. No source names a number. It is wide enough for a graph name plus its "Citation for" label, and narrower than the omnibar results panel (640pt).
- **D-2 / D-3:** the panel is capped to the content column (column − anchor x − 16pt), using a width the content view measures. The map row says only "raise the width"; the cap stops the panel from running off a small window.
- **Brief:** none. The map scopes Batch 7's other PRs as ungated. Only the panel's width changes; the rows, components and copy stay the same.

## Round 1

### X-1 · blocker · STEP-2: `.onGeometryChange` placement is ambiguous
On the bare `VStack` it could measure less than the column; after `.frame(maxWidth: .infinity)` it measures the full column.
**Answer:** Fixed. §6 macOS "Measure" row and STEP-2 now place it after `.frame(maxWidth:maxHeight:)` and `.background`, immediately before `.coordinateSpace(name:)`.

### X-2 · major · T-1 has no stated test form
**Answer:** Fixed. T-1 names `@Suite struct` with no `@MainActor`, `@Test(arguments:)` over `(CGFloat, CGFloat, CGFloat)` tuples, and `#expect(… == expected)`.

### X-3 · major · The manual checks cannot be run by the implementing agent
**Answer:** Fixed. §10 says the user or the `ship-pr` orchestrator runs it, and the PR carries screenshots at 1366pt and at the minimum width. The DoD line says the same.

### X-4 · minor · Version bump wording
**Answer:** Fixed. STEP-4 names all four `MARKETING_VERSION` lines (the reviewer counted three; the pbxproj has four) and says to read `VERSION` on `main`.

### X-5 · minor · STEP-5 docs edits underspecified
**Answer:** Fixed. STEP-5 now gives:
- the map status `Landed #N`;
- the `completed.md` heading and Landed line, in the shape of S14-34's entry;
- that the docs commit comes after the PR exists;
- that the stale **Ships.** clause in the scope entry is rewritten.

### X-6 · minor · `panelWidth` is both a `static let` and a `static func`
**Answer:** Fixed. The function is renamed `fittedPanelWidth` (also M-4), and the sketch uses `Self.panelWidth`.

### X-7 · minor · Non-positive width at a tiny column
**Answer:** Fixed. D-5 now forbids `max(0, …)` explicitly. The window minimum keeps the column at 436pt or more, so the width is at least about 376pt.

### M-1 · major · Deleting the dogfood note drops the deferred fix directions
**Answer:** Fixed. STEP-5 moves the note to `docs/ideas/history-jump-menu-labels.md` instead of deleting it:
- directions 2–4 kept, and direction 1 (shipped) dropped;
- "How it works today" corrected;
- the note indexed in `docs/ideas/README.md`.

The three Out rows point at that file. The map's "Where the notes went" row names it. The DoD checks no ux link is left.

### M-2 · minor · Codebase fit does not justify a second pane measurement
**Answer:** Fixed. The Codebase fit row for `contentWidth` names `pvMeasuresWidthClass()` and says why it does not fit: it publishes a class, not points, and sits below the toolbar overlays. The property gets a doc comment saying it is for overlay placement only.

### M-3 · minor · The omnibar results panel already overflows the column
**Answer:** Fixed. A row is added to `docs/ideas/structural-debt.md` ("S14-35 plan") in this planning commit, and the Out line's home is that row. The PR does not widen.

### M-4 · minor · One name for the constant and the function
**Answer:** Fixed. `panelWidth` stays the preferred width (the `OmnibarResultsModel` precedent name). The function is `fittedPanelWidth(contentWidth:anchorMinX:)`, and the §6 row says which is which.

### M-5 · minor · The header links a file STEP-5 removes
**Answer:** Fixed. STEP-5 repoints this plan's **Context:** header and the scope entry's Context at the idea file.

### M-6 · nit · Test file name
**Answer:** Fixed. The file is renamed `HistoryJumpMenuModelTests.swift`, after the type under test, like `OmnibarResultsModelTests.swift`.
