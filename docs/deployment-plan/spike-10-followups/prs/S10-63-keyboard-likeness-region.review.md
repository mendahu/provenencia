# Review — S10-63 Keyboard region for a likeness

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus, pr-reviewer), executability (sonnet, Explore)

## To user

None.

## Writer's own calls

- D-3: two whole-file labels by viewer kind ("Use the whole page" on a PDF, "Use the whole photo" otherwise), exposed as `depictionWholeFileTitle`.
- D-4: Draw and both buttons share `prepareDepictionRegion()`; the buttons then call `commitViewerRegion`.
- D-5: the whole-file region is `ArtifactRegionDraft.whole`, built from `CGRect.unitSquare`.
- D-6: the page and the Citation region come from one `depictionCitationLocator` (saved locator on an existing Citation, unsaved on a new one).
- D-7: **Use the citation's region** is hidden, not disabled, when the Citation has no region; both buttons are disabled exactly when Draw is.
- D-8: a button replaces the region; the derived `isDepictionRedrawn` decides the redraw; no new flag, no confirm.
- D-9: `FlowLayout(spacing: space4, lineSpacing: space3)`, Draw first, new buttons `.secondary` `.sm` with no icons.
- D-10: no new focus cases; Tab reaches the buttons after Draw.
- D-11: Cancel takes focus through `claimTab`'s double hop; the existing drawing hint is announced.
- D-12: straight apostrophe, matching the namespace.
- D-14: `canStartDepictionDrawing` keeps its name; its doc comment names every reader.
- D-15: a successful press announces the new caption; a refused press adds nothing (UX-6 is Out).
- D-16: one Platform helper, `AccessibilityAnnouncement.post`; `NameValueEditorForm` moves onto it.
- Version bump none, as every Spike 10 and Spike 11 PR since #402.

## Round 1

### M-1 · major · `useDepictionRegion` copies Draw's guard and page move
`startDepictionDrawing()` already guards `canStartDepictionDrawing` and moves the viewer to `depictionPage`; share one preamble.
**Answer:** Fixed — D-4 now names `prepareDepictionRegion() -> ComposerBase?` (guard, base check, page move); `startDepictionDrawing` and `useDepictionRegion` both call it. Codebase fit, STEP-2 and Deletes updated.

### M-2 · major · `depictionCitationRegion` repeats `depictionPage`'s ternary
**Answer:** Fixed — D-6 adds one private `depictionCitationLocator`; `depictionPage` and `depictionCitationRegion` both read it.

### M-3 · major · A second hand-copied VoiceOver announcement
`NameValueEditorForm.moveAndAnnounce` is the only post; the plan copied it.
**Answer:** Fixed — D-16 adds `macos/App/Platform/AccessibilityAnnouncement.swift` (`post(_:priority:)`), `moveAndAnnounce` moves onto it, Scope In gains the row, STEP-2 registers the file in `project.pbxproj`, AC-3.4 and manual check 8 cover the moved call. The `identityAnnouncement`-as-hint shape is logged here and listed Out as not planned.

### M-4 · minor · The draw rule's name and comment go wrong with three readers
**Answer:** Fixed (comment) — D-14 rewrites the doc comment to name Draw, both buttons and `prepareDepictionRegion`; the rename is Out (eight call sites and tests for no behavior).

### M-5 · minor · The whole-file region belongs on the region type
**Answer:** Fixed — D-5 is now `ArtifactRegionDraft.whole` in `ArtifactRegionTool.swift`, built from `CGRect.unitSquare`. Moving the three test-file rectangle builders onto a shared builder is logged here and listed Out as not planned.

### M-6 · nit · `FlowLayout` line spacing
**Answer:** Fixed — `lineSpacing: PVSpacing.space3`, matching `CitationComposerAuthoringBox.swift`'s other `FlowLayout`.

### X-1 · major · Swift checks and `check-pr.sh` cannot pass off macOS
**Answer:** Declined — `.agents/skills/ship-pr/SKILL.md` ("You are on the user's Mac: `scripts/check-pr.sh` needs `xcodebuild`. On another machine the gate cannot pass; stop and say so.") already binds the run, and no Spike 11 plan restates it (`docs/deployment-plan/spike-11/prs/S11-0*.md`). The Xcode commands now carry `CODE_SIGNING_ALLOWED=NO`, as `scripts/check-pr.sh` `xcode_test` does (X-8).

### X-2 · major · Focus on Cancel has no fallback
**Answer:** Fixed — D-11 binds `claimTab`'s double hop (write on the next turn, write again a turn later if it did not land) and clears the flag in `.onDisappear`; the Risk row cites it and a new Stop line covers manual check 6 failing after the double hop.

### X-3 · major · What a press does on failure or success
**Answer:** Fixed — D-15: success announces `depictionCaption` through the helper (no new string, AC-1.9); a refused press adds nothing beyond the disabled state, Callout and caption, with an Out line homed at audit UX-6.

### X-4 · minor · Conditional `import AppKit`
**Answer:** Fixed — STEP-4 now says the viewer pane needs no new import: the AppKit call lives in the helper, which imports AppKit.

### X-5 · minor · Wrong symbol `ArtifactRegionGeometry.unitSquare`
**Answer:** Fixed — Current state gains the `CGRect.unitSquare` row, and Codebase fit names it correctly.

### X-6 · minor · Test fixture details; no test for D-3's label
**Answer:** Fixed — STEP-5 names `regionJSON(_:page:)` with `seedCitation(locatorJSON:)`, the edit entry `makeModel(citationID: "cit-1")`, and `setPage` after `prepare()`; D-3 exposes `depictionWholeFileTitle` and T-10 tests it (AC-1.8).

### X-7 · minor · Commit boundaries
**Answer:** Fixed — §7 lists the commits in order with their subjects; STEP-3 and STEP-5 share one commit.

### X-8 · minor · Xcode command differs from CI
**Answer:** Fixed — every `xcodebuild` command in §7 carries `CODE_SIGNING_ALLOWED=NO`.

### X-9 · minor · The empty caption sits beside buttons it does not mention
**Answer:** Fixed (Out line) — the copy still describes the main path, so it stays; rewording is Out, homed at the audit's copy nits (UX-12/UX-13).

### X-10 · nit · Two Go rows repeat existing ones
**Answer:** Logged — the rows exist to pin the whole-file payload (AC-1.7), not to fail first.

### X-11 · nit · D-6 copies the saved region on an existing Citation
**Answer:** Fixed — AC-2.7 and the second half of T-6 pin it.

### Also aligned
STEP-6 now holds only the decisions-note edit; the map row, Status and Landed note go in `implement-pr` step 8's commit "Record S10-63 as landed in pull request N." after the draft PR has a number (§7 note and DoD).
