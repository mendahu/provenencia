# Review — S14-03 Rectangle-only depiction drawing

**Verdict:** Fits with changes — Ready with changes. Every finding answered; the plan stays Blocked on board S14-D6 and is refreshed (and reviewed again) when the board is agreed.
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet). Both reviewed the plan and the S14-D6 brief.

## To user
- None. The two product calls (a brief gates the PR; the needs-page Callout sets the page and saves) were answered by the user on 2026-10-31 before the review.

## Writer's own calls
- A depiction dialog offers one radio whether or not Draw is pressed, because the viewer commits into an open depiction draft (D-1).
- "While drawing" includes an armed Citation region tool: Set page, chip ×, the Clear menu and Find lock (D-5).
- A failed cut shows in the caption alone; the preview stays the empty Thumbnail (D-9, from M-8). The alternative, widening `PVThumbnail`'s contract to show errors, was not taken.
- Set page moves onto `PVChip` (D-6, from M-6); the board shows the look.
- The embedded tool strip in `ArtifactViewer` is deleted (D-3, from M-11).
- The freehand cap announcement and the loading viewer are manual checks; no test seams (D-12, AC-4.2).
- Committed polygon depictions keep their kind on handle edits (D-13).
- Patch bump 0.2.2 → 0.2.3.
- The composer is drawn at Wide as it is today; its pane switch stays in `docs/ideas/wide-layout.md`.

## Round 1

### M-1 · blocker · D-9 writes a second "can the Citation be saved" rule beside one in a view
`CitationComposerEvidenceColumn` private `canSaveCitation` already holds the rule; `saveDepictionCitationPage` would copy it.
**Answer:** Fixed — D-10 moves it onto the model (`canSaveCitation`, `citationSaveBlocked`); the evidence column and D-11 read it.

### M-2 · blocker · `isDrawingRegion` replaces `isDrawingDepiction` in only half the locator guards
`clearRegion`, `resetToEntireArtifact` and `clearMenuButton` kept the old guard.
**Answer:** Fixed — D-5 guards all five locator calls and disables the Clear menu; brief DF-3 and its inventory gain the Clear menu; T-8 covers clear and reset. (Also X-6.)

### M-3 · major · Three names for one armed tool; `isDrawingRegion` read the raw one
**Answer:** Fixed — D-2: `viewerArmedRegionTool` is the one property (filtering getter, policy setter), bound by the strip and the overlay; `stripArmedRegionTool` dropped; `isDrawingRegion` reads `viewerArmedRegionTool`. T-2 asserts a hidden tool does not lock.

### M-4 · major · `.failed` also covered "the viewer cannot cut"
`depictionImage(for:)` stores `render`'s nil while `snapshot` is nil.
**Answer:** Fixed — D-7: `depictionPreviewKey` is nil while `!depictionCanCut`, so no cut runs; `.cutting` while loading. `.failed` now means the cut ran and failed. Current state row updated.

### M-5 · major · Caption chain plus parallel Bools should be one enum
**Answer:** Fixed — D-8 `DepictionCaptionState` with `text` and `isWarning`; `depictionCaption`, `depictionCaptionIsWarning`, `depictionCutFailed` gone.

### M-6 · major · Set page is a hand-rolled copy of `PVChip` with no kit disabled state
**Answer:** Fixed — D-6 moves it onto `PVChip`; brief inventory row is now Component · Rethink. (Also X-7.)

### M-7 · major · Cap copy hard-codes 256
**Answer:** Fixed — `freeformCapReached(_:)` called with `ArtifactRegionGeometry.maxRegionPoints`.

### M-8 · major · A warning glyph in `PVThumbnail` breaks its "never an error UI" contract
**Answer:** Fixed — D-9: the failure shows in the caption alone (writer's call; listed above). Brief DF-5, frame 5, inventory and non-goals updated.

### M-9 · minor · `model.depictionPreviewLabel` shares the L10n key's name
**Answer:** Fixed — the view's `.saved` case passes `depictionPreviewSavedLabel`; no model member. T-11 asserts `.saved`.

### M-10 · minor · One-off announcement seam; D-9 posted from the model
**Answer:** Fixed — D-12 posts inline, checked by hand; no seam, no new test file. D-11's announcement is posted by the view, as `useRegion` does.

### M-11 · minor · The embedded tool strip in `ArtifactViewer` is dead
**Answer:** Fixed — D-3 deletes `showsToolStrip`, its branch and `onSetPage`; `armedRegionTool` becomes a value.

### M-12 · minor · `ArtifactRegionCrop.canCut` only forwards, for one caller
**Answer:** Fixed — dropped; `depictionCanCut` reads `snapshot` directly.

### M-13 · minor · The rectangle-only rule lives only in the spike map
**Answer:** Fixed — STEP-9 adds a line to `add-depiction-value`.

### M-14 · nit · Region kind-to-symbol mapping written twice
**Answer:** Logged — Out row "not planned".

### X-1 · blocker · The plan is not startable: blocked, unreviewed, no board
**Answer:** Declined — that is its intended status: `plan-pr` step 3 sets `Blocked — on board` until the user agrees the board. This review is the review; the plan is refreshed when the board is agreed (brief § 7).

### X-2 · blocker · Swift checks cannot run in this environment
**Answer:** Fixed in part — D-14 and § 8 say the Swift tests and manual checks run on macOS. Declined as a blocker: `scripts/check-pr.sh` is the repo's command and runs `xcodebuild test` on macOS, as for every Swift PR (S14-02, S14-13).

### X-3 · major · `ArtifactRegionCropTests` lines 170 – 190 are a symlink swap, not a deleted PDF
**Answer:** Fixed — the Tests row says no test deletes a PDF after it opens and T-7 builds the case.

### X-4 · minor · Composer strip row omits the Clear menu
**Answer:** Fixed — row lists `clearMenuButton` and its guard.

### X-5 · minor · STEP-8 pbxproj path and IDs
**Answer:** Fixed — moot: no new test file (M-10).

### X-6 · major · Clear menu stays enabled while a Citation tool is armed
**Answer:** Fixed — with M-2.

### X-7 · major · Set page has no disabled style
**Answer:** Fixed — with M-6.

### X-8 · major · A failed cut can never be retried
**Answer:** Fixed — D-9: `commitViewerRegion` clears a failed memo; AC-4.5 and T-7 retry after restoring the file.

### X-9 · major · No way to keep the viewer loading in a test
**Answer:** Fixed — AC-4.2 is manual check 3; no seam (Out row).

### X-10 · major · `StoreBoom` is private, so the message cannot be asserted
**Answer:** Fixed — AC-5.2 asserts `fields.error != nil`.

### X-11 · major · D-9 gaps: refused `setPage`, "success", button while loading
**Answer:** Fixed — D-11 returns false on a refused `setPage`, defines success as `fields.error == nil && !depictionNeedsPage`, and disables the button unless `!citationSaveBlocked && depictionCanCut`; brief DF-6 says so.

### X-12 · minor · Cannot-cut caption on audio or no Artifact; unchanged edit on a cannot-cut file
**Answer:** Fixed — Edges: depiction Properties are offered on image and PDF only (`depictionIsOfferedOnImageAndPDFOnly`); an unchanged edit shows `.cannotCut` and stays savable (T-6).

### X-13 · minor · Repeated announcement; test points for the cap
**Answer:** Fixed — D-12 announces once per shape; the test is gone (M-10).

### X-14 · minor · Where `.disabled` goes for ⌘F; two disable mechanisms; armed vs viewer-armed
**Answer:** Fixed — D-5 puts it on `findToggleButton` (covers its background), keeps `toolStrip`'s transcribing disable as is, and reads `viewerArmedRegionTool`.

### X-15 · minor · STEP file lists; archive link fixes
**Answer:** Fixed — STEP-4, STEP-6 and STEP-8 list `L10n.swift` and `Localizable.xcstrings`; STEP-9 names the `../` link change and every link to update.

### X-16 · nit · No hexagon fixture; how to assert "no new image"
**Answer:** Fixed — AC-1.6 gives the points and points at `editWithoutRedrawKeepsTheImage`'s assertion.

### Brief notes (executability)
Frame 1 needs a Citation with a region; frame 2 and Set page on an image; DF-6 while loading.
**Answer:** Fixed — frames 1 and 2 reworded (the Citation has a circle region; Set page and Find only on the PDF variant); DF-6 disabled while the viewer is opening the file.

## Round 2

Both reviewers confirmed the round-1 answers except where noted below.

### M-15 · blocker · A Citation-armed rectangle still sets the depiction outside the one rule
The getter hid a tool only while drawing, so a rectangle armed before the dialog stayed armed and committed into the depiction without Draw or the can-cut gate.
**Answer:** Fixed — D-2: the getter is nil while a depiction dialog is open and not drawing; the setter's tool click there calls `startDepictionDrawing()`. AC-1.3 covers a rectangle too; AC-1.7 and T-2 cover the radio as Draw. Brief DF-1 updated.

### M-16 · blocker · "Needs a new image and cannot get one" worked out in three places
**Answer:** Fixed — D-7: one private `depictionNeedsCut`; `DepictionPreviewState.waiting` and `.unavailable`; D-9 `depictionImageBlocksSave` reads the state; `DepictionCutKey.isLoading` deleted. (Also X-4, X-5.)

### M-17 · blocker · D-11 set the viewer's page through its own guard
**Answer:** Fixed — D-11 calls `setPageFromViewer()`; `canSaveDepictionCitationPage` includes `!isDrawingRegion` and the button reads it; T-9 adds the armed-tool case.

### M-3 / M-11 / M-14 leftovers (below blocker)
**Answer:** Fixed — D-1 makes the chrome's binding and `onSetPage` non-optional and drops the `!= nil` check; M-14 gets a line in the idea file (STEP-9); D-12's flag clears in a `didSet` whenever the count drops below the cap; the board confirms Find closing on any arm (Stop condition).

### X-1 · blocker · A region button cannot retry a failed cut: the preview task does not re-run
**Answer:** Fixed — D-9: `useRegion` runs `renderDepictionPreview()` after a successful set; Save stays enabled while `.cutting`, as today. AC-4.5 and manual check 5 use a region button.

### X-2 · major · Step order breaks two step checks
**Answer:** Fixed — T-2 no longer asserts `isDrawingRegion` (T-8, STEP-3 does); T-6's Save cases moved to T-7 (STEP-5).

### X-3 · major · `regionTools` parameter collides with the private view builder
**Answer:** Fixed — D-1 renames the builder `regionToolGroup`.

### X-6 · minor · No PDF edit fixture for AC-4.4
**Answer:** Fixed — AC-4.4 gives the locator, the observation region and the steps with the file's existing helpers.

### X-7 · minor · Page after a failed save
**Answer:** Fixed — AC-5.4: the unsaved page stays on the draft.

### X-8 · minor · Home of `canSaveCitation`
**Answer:** Fixed — `CitationComposerModel.swift`, beside `setPageFromViewer`.

### X-9, X-10, X-11 · minor
**Answer:** Fixed — D-12 flag re-arms below the cap; D-5 applies `.disabled` after `.background`; D-6 keeps the accessibility identifier and names `PVChip`'s metrics.
