# Review — S14-35 Wider history jump menu

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

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
**Answer:** Fixed. Directions 2–4 are kept and re-homed; round 2 (X-8) changed where they go:
- In round 1 they went to a new file, `docs/ideas/history-jump-menu-labels.md`.
- In round 2 they moved to a new section in `docs/ideas/shared-list-helpers.md`, "History menu row labels". That file already holds "The history menu on the shared rows", and its README line already says "history menu rows".

STEP-5 writes the section, with its code pointers corrected, and then deletes the ux note. The three Out rows link that file, and the DoD checks that no link to the ux note is left.

### M-2 · minor · Codebase fit does not justify a second pane measurement
**Answer:** Fixed. The Codebase fit row for `contentWidth` names `pvMeasuresWidthClass()` and says why it does not fit: it publishes a class, not points, and sits below the toolbar overlays. The width carries a doc comment saying it is for overlay placement only. Since M-7, that comment is on `WorkspaceContent`'s `@State`.

### M-3 · minor · The omnibar results panel already overflows the column
**Answer:** Fixed. A row is added to `docs/ideas/structural-debt.md` ("S14-35 plan") in this planning commit, and the Out line's home is that row. The PR does not widen.

### M-4 · minor · One name for the constant and the function
**Answer:** Fixed. `panelWidth` stays the preferred width (the `OmnibarResultsModel` precedent name). The function is `fittedPanelWidth(contentWidth:anchorMinX:)`, and the §6 row says which is which.

### M-5 · minor · The header links a file STEP-5 removes
**Answer:** Fixed. STEP-5 repoints this plan's **Context:** header and the scope entry's Context at the idea file.

### M-6 · nit · Test file name
**Answer:** Fixed. The file is renamed `HistoryJumpMenuModelTests.swift`, after the type under test, like `OmnibarResultsModelTests.swift`.

## Round 2

Both reviewers checked the round-1 answers. All hold except M-3's debt-row text, which led to M-7.

### M-7 · blocker · The plan's own debt row says `contentWidth` is in the wrong place
The structural-debt row said the omnibar cap would have to move `HistoryJumpMenuModel.contentWidth` to a column-level value. `WorkspaceContent` already owns both overlay models as `@State`.
**Answer:** Fixed. The width now lives on `WorkspaceContent` and the menu's model stays free of it:
- D-3: `WorkspaceContent` owns `@State contentWidth` and passes it as `HistoryJumpMenuHost(jumpMenu:contentWidth:)`.
- `HistoryJumpMenuModel` keeps only `panelWidth` and `fittedPanelWidth`, so T-1 is unchanged.
- The §3 In row, §4 Host row, §6 Codebase fit and macOS rows, and STEP-1 to STEP-3 are updated to match.
- The debt row now says the omnibar reads the same value as one more argument.

### X-8 · minor · STEP-5 does not choose the ideas README heading
**Answer:** Fixed. The leftovers go into a section of `docs/ideas/shared-list-helpers.md`, next to its history-menu section, so no new file or index line is needed (see M-1).

### X-9 · minor · AC-1.2's 61-character sample is borderline at 600pt
**Answer:** Fixed. The sample Citation is now "Smith, p. 4", so the leaf is 51 characters. D-1 says about 55 characters fit. The Stop condition says there is no fallback width.

### X-10 · nit · The scope entry's **Ships.** text omits the cap
**Answer:** Fixed. STEP-5 adds "capped to the content column on a small window" to the scope entry.

### X-11 · nit · STEP-2's check proves only that the code compiles
**Answer:** Logged. The real check is §10 step 3, which §10 already gives to the user or orchestrator.

Note: the executability reviewer could not resolve `f48846c` in its view of the shallow clone. It is the parent of the first plan commit on this branch and the merge commit of #727 on `main`.
