# S14-03 — Rectangle-only depiction drawing

**Status:** Blocked — on board S14-D6
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D6](../design/S14-D6-composer-drawing-feedback.md), board not drawn yet (user, 2026-10-31: the PR changes how the composer looks, so it waits on a brief)
**Board:** — (when agreed: the *Citation Composer* Claude Design project, turn "S14-D6 · Drawing feedback"; until then § 6 macOS describes each frame in words)
**Depends on:** S14-13 (landed [#632](https://github.com/mendahu/provenencia/pull/632)), S14-D6
**Blocks:** —
**Branch:** `s14-03-rectangle-only-depiction-drawing`
**Version bump:** patch, 0.2.2 → 0.2.3 — fixes to existing drawing feedback and copy, no new feature ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md) (what Go accepts for a depiction region; STEP-9 adds one line to it)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/depiction-values.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "A depiction is drawn as a rectangle", "FakeStore is a seeded fake, not a second engine"; bar outcomes 4 and 9; the 2026-10-18 likeness audit items UX-5, UX-6, UX-7 (composer feedback), UX-12, UX-13 (copy) and the freehand cap feedback, whose text is in `git show f88bef38^:docs/ideas/depictions-and-likenesses.md` lines 19 – 21; S14-13's Landed note (the "can cut" predicate is `ArtifactRegionCrop.snapshot`)
**Planned:** 2026-10-31 against `main` @ `f2dc2a95`
**Review:** [`S14-03-rectangle-only-depiction-drawing.review.md`](S14-03-rectangle-only-depiction-drawing.review.md) — round 1 answered, round 2 pending

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

A researcher drawing a likeness or a picture on a Citation's file is offered the rectangle tool alone, because the saved image is always the rectangle around the region; "Use the citation's region" on a circled entry sets the circle's bounding rectangle. While any region is drawn, the controls that would move the Citation's locator under the drawing are disabled. Draw is blocked on a file the viewer cannot cut, a failed cut shows at once with Save disabled, a PDF Citation saved without a page can save its page from the warning, the captions name the region buttons, and VoiceOver hears the freehand cap. The work comes from the user's rectangle-only call (2026-10-27) and the Spike 10 likeness audit leftovers moved here.

**Why now.** S14-02 shipped Event and Place pictures, so more depictions are drawn; S14-13 left the one "can cut" predicate this PR gates Draw on.

**What stays.** A Citation's own region keeps all seven tools. Go accepts any valid depiction region: a polygon depiction saved earlier opens, shows and saves as it is until it is drawn again. The composer's panes, the depiction row's layout, the drawing band, and the cut itself (`ArtifactRegionCrop.cut`) are unchanged.

## 2. User stories

**US-1.** As a researcher drawing a depiction, I am offered the rectangle tool alone, so that what I draw is what is saved.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Likeness dialog open on an image Citation, Draw pressed | `viewerRegionTools == [.rectangle]`; `viewerArmedRegionTool == .rectangle` |
| AC-1.2 | Freeform armed for the Citation region, then a likeness dialog opened and Draw pressed | `viewerArmedRegionTool == .rectangle` |
| AC-1.3 | Freeform armed, then a likeness dialog opened, Draw not pressed | `viewerRegionTools == [.rectangle]`; `viewerArmedRegionTool == nil`; `isDrawingRegion == false` |
| AC-1.4 | No depiction dialog open | `viewerRegionTools == ArtifactRegionTool.allCases` (seven) |
| AC-1.5 | Drawing a depiction, `viewerArmedRegionTool = nil` (the selected rectangle radio clicked) | drawing ends: `presented == .authoring`, `armedRegionTool == nil` |
| AC-1.6 | Edit of a saved likeness whose region is the 6-point freeform hexagon (0.5, 0.1), (0.8, 0.3), (0.8, 0.7), (0.5, 0.9), (0.2, 0.7), (0.2, 0.3), not redrawn, saved | the commit carries no depiction input, asserted as `editWithoutRedrawKeepsTheImage` asserts it; the region is unchanged |

**US-2.** As a researcher, "Use the citation's region" on a non-rectangle region sets the rectangle around it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Citation region a circle ring centered (0.5, 0.5), radius 0.2 in a square file | `depictionCitationRegion` is `.rectangle` with points (0.3, 0.3), (0.7, 0.3), (0.7, 0.7), (0.3, 0.7) within 1e-6 |
| AC-2.2 | Citation region `.lOpenTopRight` inside box (0.1, 0.2) – (0.6, 0.9) | `.rectangle` over (0.1, 0.2) – (0.6, 0.9) |
| AC-2.3 | Citation region `.rectangle` | returned unchanged (same points, same order) |

**US-3.** As a researcher drawing any region, the controls that would change the Citation's locator under me are disabled.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A Citation region tool armed (no dialog) | `isDrawingRegion`; `setPageFromViewer()` returns false; `removePage()`, `removeRegion()`, `clearRegion()`, `resetToEntireArtifact()` leave `locator` unchanged |
| AC-3.2 | Drawing a depiction | `isDrawingRegion`; the same five calls are refused |
| AC-3.3 | PDF Citation, Find open, a Citation region tool armed through `viewerArmedRegionTool` | `artifactViewer.isFindPresented == false` |
| AC-3.4 | No tool armed, not drawing | `isDrawingRegion == false`; `setPageFromViewer()` sets the page |

**US-4.** As a researcher, Draw is blocked on a file the viewer cannot cut, and a failed cut shows at once with Save disabled.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Image Artifact whose file is missing (viewer `emptyReason == .missingFile`), likeness dialog open | `depictionCanCut == false`; `canStartDepictionDrawing == false`; `useDepictionRegion(.whole)` returns false; `depictionCaptionState == .cannotCut` |
| AC-4.2 | Viewer still loading, a region on the draft | Draw disabled; the preview shows loading; caption is not a warning; Save disabled (manual check 3: the viewer has no loading test seam) |
| AC-4.3 | PDF opened in the viewer, its object file then deleted, a region committed and the preview rendered | `depictionPreviewState == .failed`; `depictionCaptionState == .cutFailed`; `authoringCanConfirm == false`; `canStartDepictionDrawing == true` |
| AC-4.4 | AC-4.3 on an edit with a redrawn region | `editCanConfirm == false` |
| AC-4.5 | AC-4.3, file restored, the same region committed again | the cut runs again: `depictionPreviewState == .cut`; `authoringCanConfirm == true` |

**US-5.** As a researcher on a PDF Citation saved without a page, I save the viewer's page to it from the warning and then draw.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Existing PDF Citation with no page, viewer on page 2, likeness dialog open, `saveDepictionCitationPage()` | FakeStore's `citationsByID["cit-1"]?.locatorJSON` decodes to page 2; `depictionNeedsPage == false`; `canStartDepictionDrawing == true`; returns true |
| AC-5.2 | Same, with `cit-1` removed from FakeStore's `citationsByID` so `updateCitation` throws (no new hook; working rule "FakeStore is a seeded fake") | returns false; `depictionNeedsPage == true`; `fields.error != nil` |
| AC-5.3 | Image Citation, or PDF Citation with a page | returns false; no `updateCitation` in FakeStore's `recordedCalls` |

**US-6.** As a researcher, the depiction copy names what I can do, and VoiceOver hears the freehand cap.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-6.1 | No region drawn | `depictionCaptionState == .empty`, text "Nothing drawn yet. Draw on the file, or use a region button below" |
| AC-6.2 | PDF Citation needs a page | `.needsPage`, text "Draw and the region buttons wait until the citation has a page"; Callout "This citation has no page yet. Save the viewer's page to it to draw a region."; button "Save page 2 to this citation" |
| AC-6.3 | Saved region hidden | Callout "This region no longer matches the citation's page. Draw it again or use a region button below." (manual check 6) |
| AC-6.4 | Edit with the region unchanged | `depictionPreviewState == .saved`, which the view labels "Saved image"; a redrawn edit and a new draft are not `.saved` and read "Image to save" |
| AC-6.5 | Freeform tool with 256 points placed, two more clicks | each click beeps; VoiceOver hears "This freehand region holds its most points, 256. Click the first point to close it." once (manual check 2) |

## 3. Scope

| In | Covered by |
| --- | --- |
| A depiction dialog offers `.rectangle` alone; a Citation region all seven (DF-1, DF-2) | STEP-1, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5 |
| `ArtifactRegionDraft.boundingRectangle` for "Use the citation's region" | STEP-2, AC-2.1, AC-2.2, AC-2.3 |
| Drawing disables Set page, the chip ×, the Clear menu and Find; hides the locator hint (DF-3, UX-5) | STEP-3, AC-3.1, AC-3.2, AC-3.3, AC-3.4 |
| Set page becomes the kit `PVChip` (it gains a disabled state) | STEP-3 |
| The dead embedded tool strip in `ArtifactViewer` goes | STEP-1 |
| Draw and region buttons blocked on a file the viewer cannot cut (DF-4, UX-6) | STEP-4, AC-4.1, AC-4.2 |
| A failed cut shows with Save disabled, and can be retried (DF-5, UX-6) | STEP-5, AC-4.3, AC-4.4, AC-4.5 |
| The needs-page Callout saves the viewer's page; one "can save the Citation" rule on the model (DF-6, UX-7) | STEP-6, AC-5.1, AC-5.2, AC-5.3 |
| Captions and region-hidden Callout name the region buttons; "Saved image" label (DF-7, DF-8, DF-9, DF-10, UX-12, UX-13) | STEP-7, AC-6.1, AC-6.2, AC-6.3, AC-6.4 |
| Freehand cap announced to VoiceOver (DF-11) | STEP-8, AC-6.5 |
| Wide: nothing moves (DF-W1) | Manual check 8 |

| Out | Home |
| --- | --- |
| Removing shapes from Citation regions | not planned (working rule "A depiction is drawn as a rectangle") |
| Go-side rectangle enforcement, rewriting saved polygon depictions | not planned (same rule: Go keeps accepting any valid region) |
| Reshaping a committed polygon depiction's handles into a rectangle | not planned: handle edits keep the saved kind (D-11) |
| The composer's 1400pt pane switch onto the width class | [`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md), the composer's next layout brief |
| A refused image glyph (UX-11), thumbnail work (PRF-2), Go crop work (PRF-4), `encodeJSON` key order | [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) |
| Showing every saved depiction on the photo | same idea file |
| A test seam for VoiceOver announcements or a loading viewer | not planned: the two cases are manual checks (D-12, AC-4.2) |
| `CitationComposerViewerPane.regionSymbol` repeating `ArtifactRegionTool.symbol` | not planned (review M-14, logged) |

## 4. Current state

Verified at `f2dc2a95`.

| Area | Where | Today |
| --- | --- | --- |
| Tool enum | `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift` `ArtifactRegionTool`, `ArtifactRegionKind`, `ArtifactRegionDraft` (`isValid`, `static let whole`) | Seven tools; `ArtifactRegionDraft` has no bounding-rectangle conversion |
| Region geometry | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` `boundingRect(of:)`, `rectanglePoints(from:to:)`, `maxRegionPoints` (256), `appendFreeformVertex` | The pieces `boundingRectangle` composes |
| Tool strip | `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` `ArtifactViewerToolChrome` `regionTools`, `regionRadio`, `setPageButton` | `ForEach(ArtifactRegionTool.allCases)`; a click on the selected radio sets the binding to nil; `setPageButton` is a plain `Button` painting `PVChip`'s accent-selected colors by hand, so `.disabled` would not dim it |
| Embedded strip | same file `ArtifactViewer` `showsToolStrip` (default true), `onSetPage`, `armedRegionTool: Binding<ArtifactRegionTool?>?` | Its only caller, `CitationComposerViewerPane`, passes `showsToolStrip: false`; the branch, `onSetPage` and the binding form exist only for it. The canvas reads only `armedRegionTool?.wrappedValue` |
| Composer strip | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` `toolStrip`, `findToggleButton` (⌘F in its `.background`), `clearMenuButton`, `locatorStrip`, `locatorChip` | Strip binds `$model.armedRegionTool`, the viewer `$model.viewerArmedRegionTool`; Find and ⌘F enabled while drawing; chip × enabled; Clear menu disabled on `isLocatorArtifactOnly \|\| isDrawingDepiction`; `locatorDrawHint` always drawn; the whole strip `.disabled(isTranscribing)` |
| Drawing band | same file `drawingBand`, `claimDrawingFocus` | Hint or wrong-page warning, and Cancel. Unchanged |
| Composer tool state | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `armedRegionTool`, `setPageFromViewer`, `removePage`, `removeRegion`, `clearRegion`, `resetToEntireArtifact`, `disarmRegionTool` | The five locator calls guard `isDrawingDepiction` only, not an armed Citation tool |
| Depiction drawing | `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift` `viewerArmedRegionTool` (getter hides the tool on the wrong page; setter writes `armedRegionTool`; nothing writes through it today), `startDepictionDrawing`, `commitViewerRegion`, `depictionCitationRegion`, `useDepictionRegion`, `depictionDrawBlocked`, `canStartDepictionDrawing`, `depictionCaption`, `depictionPreviewKey`, `depictionPreviewState`, `depictionImage(for:)`, `depictionNeedsPage` | `startDepictionDrawing` keeps an already armed tool (`if armedRegionTool == nil`); `depictionCitationRegion` returns the region as is; `depictionDrawBlocked` is `isTranscribing \|\| depictionNeedsPage`; `depictionImage(for:)` stores a nil render under the current key, so a viewer still loading or showing an empty state makes the preview `.failed`, and a failed key is never cut again |
| Save gates | `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift` `authoringCanConfirm`, `editCanConfirm`; catch of `DepictionRenderFailed` | Save enabled on a failed preview; the failure shows only after Save |
| Citation Save rule | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` private `canSaveCitation` | `!fields.isSaving && !isTranscribing && !isDrawingDepiction && (fields.isDirty \|\| activeCitationID == nil)`, in the view |
| Depiction row | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` (`preview`, `useRegion`) | Needs-page and region-hidden `PVCallout`s with no actions; `.failed` draws the same empty `PVThumbnail` as `.none`; `.saved` labelled `depictionPreviewLabel` ("Image to save"); `useRegion` posts the caption from the view |
| Can-cut predicate | `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` `snapshot(_:page:)` | Nil when the viewer holds nothing `cut` can cut (loading, empty state, no PDF document or source, no page). Read only by `render` |
| Freehand cap | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` `ArtifactRegionOverlayView.handleFreeformClick` (private), `inProgressPoints` | `NSSound.beep()` only |
| Announcements | `macos/App/Platform/AccessibilityAnnouncement.swift` `post(_:priority:)` | The one VoiceOver announcement path; no test seam |
| Citation save | `macos/App/Features/CitationComposer/CitationFieldsDraft.swift` `saveCitation()` | Saves the whole Citation draft, sets `baseline`, reports `error` |
| Copy | `macos/App/Platform/L10n.swift` `L10n.CitationComposer.depictionCaptionEmpty`, `depictionCaptionNeedsPage`, `depictionNeedsPage`, `depictionRegionHidden`, `depictionPreviewLabel`, `depictionRenderFailed`, `locatorDrawHint`; `macos/App/Resources/Localizable.xcstrings` | Copy names Draw only (UX-12) |
| Tests | `macos/ProvenenciaTests/CitationComposerDepictionTests.swift` `existingPDFCitationKeepsItsPage`, `viewerToolIsUnarmedOnTheWrongPage`, `drawingDisablesCitationControlsAndLeaveReturns`, `missingFileCutFails`, `captionFollowsTheDraft`, `citationRegionIsCopied`, `regionButtonsFollowDraw`, `editWithoutRedrawKeepsTheImage`; `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift` (the 256 cap of `appendFreeformVertex` is already tested) | `missingFileCutFails` draws on a missing file and expects a failed cut; after this PR Draw is blocked there (T-6 replaces it). No test deletes a PDF after it opens: T-7 builds that case (`ArtifactRegionCrop.pageBox` returns nil when `ProjectFiles.objectURL` finds no file) |

**Precedent.** S10-63 (keyboard likeness region) added `useDepictionRegion` and the region buttons through `prepareDepictionRegion`, one rule for Draw and the buttons: extend that rule (`depictionDrawBlocked`) rather than adding a second gate. S14-13 ([#632](https://github.com/mendahu/provenencia/pull/632)) left `ArtifactRegionCrop.snapshot` as the one can-cut predicate: read it, do not re-derive from `emptyReason`. `DepictionPreviewState` and the private `DepictionPageCheck` in the same file are the shape for the caption state (D-8).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ArtifactViewerToolChrome` takes `regionTools: [ArtifactRegionTool]` *(new param)* and draws only those. The composer passes `model.viewerRegionTools` *(new)*: `[.rectangle]` when `currentDepictionDraft != nil`, else `ArtifactRegionTool.allCases`. | The viewer stays unaware of depictions (map scope). Keyed on the open depiction draft, not on drawing mode, because `commitViewerRegion` sends any viewer commit to the open depiction draft: a freeform arm before Draw would otherwise save a polygon depiction. | Keying on `isDrawingDepiction` — leaks six tools into the depiction while its dialog is open. A depiction flag on the viewer — the viewer would learn the domain |
| D-2 | One armed-tool property for the strip, the overlay and the lock: `viewerArmedRegionTool`. Getter: nil when `armedRegionTool` is not in `viewerRegionTools`, and nil on the wrong page or a page-less PDF while drawing (as today). Setter: a tool → `artifactViewer.closeFind()`, then `armedRegionTool = tool`; nil → `disarmRegionTool()`. The strip binds `$model.viewerArmedRegionTool` (in place of `$model.armedRegionTool`). `startDepictionDrawing` sets `armedRegionTool = .rectangle` always. | One name: the strip, overlay and lock cannot disagree. Clicking the selected rectangle in drawing mode leaves drawing, as Escape does (DF-1); arming any tool closes Find (DF-3). On the wrong page the strip shows no selected radio while the band shows the warning. | A separate strip binding — a third name for one tool. Clearing `armedRegionTool` when the dialog opens — a second place that writes tool state |
| D-3 | `ArtifactViewer` loses `showsToolStrip`, its embedded `ArtifactViewerToolChrome` branch and `onSetPage`; `armedRegionTool` becomes a plain `ArtifactRegionTool?` value. The pane passes `model.viewerArmedRegionTool`. | The branch is dead (one caller, passing false), and D-1 and D-5 would otherwise widen it. | Keeping it — two strips to update for every chrome change |
| D-4 | `ArtifactRegionDraft.boundingRectangle` *(new)*: `self` when `kind == .rectangle`; else `.rectangle` with `rectanglePoints` of `boundingRect(of: points)`'s corners. `depictionCitationRegion` returns `region.boundingRectangle` (still nil when the region is invalid). | One conversion beside `.whole`; a saved rectangle keeps its vertex order, so `citationRegionIsCopied` holds. Bounding rect area ≥ polygon area, so a valid region gives a valid rectangle. | Converting in `useDepictionRegion` — the button would show one region and set another |
| D-5 | `isDrawingRegion` *(new)* = `isDrawingDepiction \|\| viewerArmedRegionTool != nil`. All five locator calls (`setPageFromViewer`, `removePage`, `removeRegion`, `clearRegion`, `resetToEntireArtifact`) guard it in place of `isDrawingDepiction`. The view disables, on `model.isDrawingRegion`: Set page (`ArtifactViewerToolChrome.isSetPageEnabled` *(new param)*), each chip ×, the Clear menu button (with its existing `isLocatorArtifactOnly`), and `findToggleButton` (the modifier on the toggle also disables the ⌘F button in its `.background`); and it does not draw `locatorDrawHint`. `toolStrip`'s `.disabled(isTranscribing)` stays as it is. | One predicate for "drawing locks the locator"; the model holds the rule, the view mirrors it. A tool hidden by D-2 is not armed on screen, so it does not lock. | Hiding the controls — the layout jumps (brief non-goal) |
| D-6 | `setPageButton` becomes `PVChip(text:isSelected:tone:size:action:)` with `tone: .accent`, `size: .sm`, `isSelected: locatorPage == model.page`, the same two titles, and `.disabled(!isSetPageEnabled)`. | The hand-built button copies `PVChip`'s accent-selected palette and has no disabled dimming (`PVHoverEffect` gives the kit's); DF-3 needs a disabled state. | `.disabled` on the plain button — no visible change |
| D-7 | `depictionCanCut` *(new)* = `ArtifactRegionCrop.snapshot(artifactViewer, page: depictionPage ?? artifactViewer.page) != nil`. `depictionDrawBlocked` = `isTranscribing \|\| depictionNeedsPage \|\| !depictionCanCut`. `depictionPreviewKey` is nil while `!depictionCanCut`, so no cut runs and nothing is stored; `depictionPreviewState` is `.cutting` then if the viewer is loading and the draft needs a cut, else follows today's rules. | One predicate, as S14-13 left it, read at the cut. Through `depictionDrawBlocked`, Draw, both region buttons, `prepareDepictionRegion` and the Tab order all gain it. `.failed` now means only "the cut ran and failed". | A forwarding `ArtifactRegionCrop.canCut` — one caller. Reading `emptyReason` — a second definition of "can cut" |
| D-8 | `DepictionCaptionState` *(new enum, same file)*: `needsPage, cannotCut, cutFailed, hidden, empty, redrawn, box`, with `text: LocalizedStringResource` and `isWarning: Bool` (true for `cutFailed` only). `depictionCaptionState` *(new)* picks the first that holds, in that order: needs page; cannot cut (`!artifactViewer.isLoading && !depictionCanCut`); cut failed (`depictionPreviewState == .failed`); hidden; empty; redrawn; box. `depictionCaption` is removed; callers read `depictionCaptionState.text`. | One place decides the order and the tone; a new caption is one case. | A caption chain plus parallel Bools — must be kept in step by hand |
| D-9 | The failed cut shows in the caption alone (`.cutFailed`, text `depictionRenderFailed`, `PVColor.warningForeground`); the preview stays the empty `PVThumbnail`. `authoringCanConfirm` and `editCanConfirm` return false for a depiction draft whose new image cannot be made: `depictionPreviewKey == nil` with a region needing a cut (cannot cut), or `depictionPreviewState == .failed`. An unchanged edit sends no image and stays savable. `commitViewerRegion` clears `depictionPreviewMemo` when it holds a failed cut, so drawing or pressing a region button again re-runs the cut. | `PVThumbnail`'s contract is "never an error UI"; the caption says it where the researcher looks, before Save (UX-6). The memo reset makes a failed cut retryable. | A warning glyph in the thumbnail — breaks the component's documented contract. A Toast or Callout — brief non-goal |
| D-10 | `CitationComposerModel.canSaveCitation` *(new; moved from the view)* = `!citationSaveBlocked && (fields.isDirty \|\| activeCitationID == nil)`, with `citationSaveBlocked` *(new)* = `fields.isSaving \|\| isTranscribing \|\| isDrawingDepiction`. The evidence column's Save reads `model.canSaveCitation`; its private copy goes. | One rule for when a Citation can be saved; D-11 reads the same rule. | A second guard list in D-11 |
| D-11 | `saveDepictionCitationPage() async -> Bool` *(new)*: false unless `depictionNeedsPage && !citationSaveBlocked && depictionCanCut`; false if `setPage(artifactViewer.page)` is refused; then `await fields.saveCitation()`; returns `fields.error == nil && !depictionNeedsPage`. The Callout's Button reads `L10n.CitationComposer.depictionSavePage(artifactViewer.page)`, `loading: model.fields.isSaving`, disabled unless `!model.citationSaveBlocked && model.depictionCanCut`; on true the view posts `depictionCaptionState.text`, as `useRegion` does. It saves the whole Citation draft, as Save does. | User, 2026-10-31: set the page and save in one click. A failure stays in `fields.error`, shown where the composer shows it today. The view posts, as every composer announcement does. | Setting the page only — the user chose save |
| D-12 | The freehand cap: at the cap, `handleFreeformClick` beeps (as today) and, once per in-progress shape, posts `AccessibilityAnnouncement.post(L10n.ArtifactViewer.freeformCapReached(ArtifactRegionGeometry.maxRegionPoints))`. A private `capAnnounced` flag on the overlay resets wherever `inProgressPoints` is reset. Checked by hand (manual check 2). | One announcement per shape, not per click; the number has one source. The cap itself is already tested. | A closure seam and a test file to assert one line — announcement seams belong on `AccessibilityAnnouncement`, for every caller, if ever |
| D-13 | A committed polygon depiction keeps its handle edits as they are (the overlay edits the committed shape in its own kind); only a new draw is a rectangle. | Working rule: a polygon depiction "saves as it is until redrawn". | Converting on handle edit — rewrites a saved polygon without a redraw (Out) |
| D-14 | Tests live in `CitationComposerDepictionTests.swift` and `ArtifactRegionGeometryTests.swift`; no new test file. The Swift tests and the manual checks run on macOS (`scripts/check-pr.sh` runs `xcodebuild test` there). | — | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionDraft.boundingRectangle` *(new)* | func | `ArtifactRegionDraft.whole`, `ArtifactRegionCrop.boundingBox(of:region:)` | Extends the draft beside `.whole`; `boundingBox` crops pixels, not a region |
| `ArtifactViewerToolChrome.regionTools`, `.isSetPageEnabled` *(new)* | params | `armedRegionTool`, `onSetPage` | Extends the one strip |
| `DepictionCaptionState` *(new)* | enum | `DepictionPreviewState`, `DepictionPageCheck` | Same pattern; replaces `depictionCaption` |
| `viewerRegionTools`, `isDrawingRegion`, `depictionCanCut`, `depictionCaptionState`, `saveDepictionCitationPage()` *(new)* | composer model members | `viewerArmedRegionTool`, `isDrawingDepiction`, `depictionDrawBlocked`, `depictionPreviewState`, `setPageFromViewer` | Each derives from existing state beside its closest member; none stores state |
| `canSaveCitation`, `citationSaveBlocked` *(new on the model)* | members | `CitationComposerEvidenceColumn` private `canSaveCitation` | Moved, not copied |
| L10n keys *(new)*: `CitationComposer.depictionCaptionCannotCut`, `depictionSavePage(_:)`, `depictionPreviewSavedLabel`; `ArtifactViewer.freeformCapReached(_:)` | key | `L10n.CitationComposer.depiction*`, `depictionWrongPage(_:)` | Same namespaces; numbers as arguments |

**Builds on.** `prepareDepictionRegion` / `depictionDrawBlocked` (one rule for Draw, buttons and Tab), `ArtifactRegionCrop.snapshot` (S14-13), `PVChip`, `PVCallout`'s actions row, `AccessibilityAnnouncement`.
**Deletes.** `ArtifactViewer.showsToolStrip`, its embedded strip branch and `ArtifactViewer.onSetPage`; the hand-built `setPageButton` styling; `CitationComposerEvidenceColumn`'s private `canSaveCitation`; `depictionCaption`; the `if armedRegionTool == nil` in `startDepictionDrawing`; `missingFileCutFails` (replaced by T-6 and T-7); the `isDrawingDepiction` guards in the five locator calls (replaced by `isDrawingRegion`).
**Next adds.** A signature key (idea file) is one more depiction Property and gets the rectangle tool with no change. A new caption is one `DepictionCaptionState` case. The composer's layout brief moves its pane switch without touching this.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Draw blocked while the viewer loads or shows an empty state | `CitationDepictionValueArea` buttons, Tab order (`claimTab`) | Tab skips a disabled Draw already (`tabFromPropertySkipsADisabledDraw`); the block lifts when the viewer finishes loading because `depictionCanCut` is derived; an Artifact switch unloads the viewer, so the block holds until the new file loads |
| Citation saved from the observation dialog (D-11) | `CitationFieldsDraft.saveCitation` | Same call the evidence column's Save makes while a dialog is open; baseline moves, so `depictionNeedsPage` clears |
| Armed tool hidden by D-2 | overlay and strip | The strip shows no selected radio, the overlay is unarmed, `isDrawingRegion` is false; Draw arms `.rectangle` |

### Scale

None — no write, list or loop changes beyond one Citation save on a button press.

### Edges

- **Assumptions.** `commitViewerRegion` routes to the depiction draft whenever `currentDepictionDraft != nil`; D-1 and D-2 key on the same condition so they cannot disagree (T-1, T-2). `depictionImage(for:)` stores whatever `render` returns under the current key; D-7 keeps it from running while nothing can be cut, and D-9 clears a failed entry on the next commit (T-7).
- **Neighbors.** A rectangle region → itself (T-4); circle, L and freeform → their box (T-4); a region touching the unit square's edge → `rectanglePoints` clamps (T-4 case at (0, 0) – (1, 1)). An unchanged edit on a file that cannot be cut: caption `.cannotCut` (Draw is blocked, so the caption says why), Save stays enabled because no image is sent (T-6).
- **Removed inputs.** A viewer that unloads (Artifact switch) makes `depictionCanCut` false until the new file loads (T-6).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — no list read from data the app did not write. Depiction Properties are offered on image and PDF Artifacts only (`depictionIsOfferedOnImageAndPDFOnly`), so `.cannotCut` never meets audio, video or a file-less Artifact.

### Schema

None — no catalog change.

### Go core

None — Go keeps accepting any valid depiction region.

### FFI

None — no RPC changes.

### macOS

Until the board is agreed, each frame in words (brief § 5):

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Region radios | `PVIconButton` in `ArtifactViewerToolChrome` | `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` | Frame 1: rectangle alone, selected; frame 2: seven |
| Set page | `PVChip` accent, small, disabled while drawing | same | Frame 1 (PDF variant) |
| Chip ×, Clear menu, Find | `PVIconButton`, `PVButton` link `.disabled(model.isDrawingRegion)` | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` | Frames 1, 2 |
| Locator hint | `Text(locatorDrawHint)` drawn only when `!model.isDrawingRegion` | same | Frames 1, 2: absent |
| Needs-page Callout | `PVCallout(tone: .warning, message:) { PVButton(..., variant: .secondary, size: .sm, loading:) }` | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` | Frames 7, 8 |
| Preview | `PVThumbnail` (empty for `.failed`); `CachedThumbnail` for `.saved`, labelled `depictionPreviewSavedLabel` | same | Frames 5, 6 |
| Caption | existing `Text(model.depictionCaptionState.text)`, `foregroundStyle(state.isWarning ? PVColor.warningForeground : PVColor.textMuted)` | same | Frames 3, 4, 5, 7 |

States: loading (buttons disabled, preview loading, Save disabled), cannot cut, cut failed, needs page, hidden, empty, redrawn, box. Accessibility: the Save page button's label is its title; DF-10 label; DF-11 announcement; `useRegion` and the Save page action post the new caption.

### Strings

| Key | Value |
| --- | --- |
| `L10n.CitationComposer.depictionCaptionEmpty` (changed) | "Nothing drawn yet. Draw on the file, or use a region button below" |
| `L10n.CitationComposer.depictionCaptionNeedsPage` (changed) | "Draw and the region buttons wait until the citation has a page" |
| `L10n.CitationComposer.depictionNeedsPage` (changed) | "This citation has no page yet. Save the viewer's page to it to draw a region." |
| `L10n.CitationComposer.depictionRegionHidden` (changed) | "This region no longer matches the citation's page. Draw it again or use a region button below." |
| `L10n.CitationComposer.depictionCaptionCannotCut` *(new)* | "Draw is unavailable: the file is not showing in the viewer" |
| `L10n.CitationComposer.depictionSavePage(_:)` *(new)* | "Save page %lld to this citation" |
| `L10n.CitationComposer.depictionPreviewSavedLabel` *(new)* | "Saved image" |
| `L10n.ArtifactViewer.freeformCapReached(_:)` *(new)* | "This freehand region holds its most points, %lld. Click the first point to close it." |

Old-scope copy: `depictionDrawingHint` ("Draw the region with the tools above. Escape goes back to the observation.") stays: one tool is still "the tools above" in the strip; the board may reword it (Stop condition).

## 7. Steps

**STEP-1 — One tool for a depiction**
Files: `macos/App/Features/ArtifactViewer/ArtifactViewer.swift`, `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift`, `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift`. Skill: `add-ui-component`.
Do: D-1, D-2, D-3. The pane passes `regionTools: model.viewerRegionTools` and binds `$model.viewerArmedRegionTool`.
Check: T-1, T-2, T-3 pass; `viewerToolIsUnarmedOnTheWrongPage` still passes.

**STEP-2 — Bounding rectangle**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift`, `CitationComposerModel+Depiction.swift`.
Do: D-4.
Check: T-4, T-5 pass; `citationRegionIsCopied`, `citationRegionOnItsPage` still pass.

**STEP-3 — Drawing locks the locator**
Files: `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `ArtifactViewer.swift`, `CitationComposerViewerPane.swift`. Skill: `add-ui-component`.
Do: D-5, D-6.
Check: T-8 passes; `drawingDisablesCitationControlsAndLeaveReturns` and `CitationAutoTranscribeTests` still pass.

**STEP-4 — Draw needs a file the viewer can cut**
Files: `CitationComposerModel+Depiction.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-7, D-8 (with the `cannotCut` key at its final value).
Check: T-6 passes; `missingFileCutFails` removed; `tabFromPropertySkipsADisabledDraw`, `regionButtonsFollowDraw`, `captionFollowsTheDraft` (reading `depictionCaptionState`) still pass.

**STEP-5 — A failed cut shows with Save disabled**
Files: `CitationComposerModel+Depiction.swift`, `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`.
Do: D-9.
Check: T-7 passes; `refusedSaveKeepsTheRegion`, `staleCutIsNotStored` still pass.

**STEP-6 — Save the page from the Callout**
Files: `CitationComposerModel+Depiction.swift`, `CitationComposerModel.swift` (or the extension that holds Citation fields), `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift`, `CitationComposerAuthoringBox.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-10, D-11.
Check: T-9 passes; `existingPDFCitationKeepsItsPage` still passes.

**STEP-7 — Copy**
Files: `L10n.swift`, `Localizable.xcstrings`, `CitationComposerAuthoringBox.swift`. Skill: `add-localized-string`.
Do: § 6 Strings; the `.saved` preview case passes `depictionPreviewSavedLabel`.
Check: T-10, T-11 pass; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-8 — Freehand cap announcement**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-12.
Check: builds; manual check 2.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`; `deployment-plan.md` (map row status, Brief link to `design/archive/`); this plan (status, Brief link); `design/S14-D6-composer-drawing-feedback.md` → `design/archive/` (its relative links gain one `../`, as archived S14-D2's did); `design/README.md` (Scheduled → Archived); `docs/ideas/depictions-and-likenesses.md` (the audit paragraph says S14-03 took its items); `.agents/skills/add-depiction-value/SKILL.md` (one line: the client draws a depiction with the rectangle tool alone, through `CitationComposerModel.viewerRegionTools`, and the saved image is the region's bounding box; Go accepts any valid region).
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-03-rectangle-only-depiction-drawing.md` passes; every moved link resolves.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

All in `macos/ProvenenciaTests/`, Swift Testing, run on macOS.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `CitationComposerDepictionTests` `depictionOffersTheRectangleAlone` *(new)* | no dialog → seven; likeness dialog → `[.rectangle]`; Draw → `.rectangle`; freeform armed first then Draw → `.rectangle` | AC-1.1, AC-1.2, AC-1.4, D-1, D-2 |
| T-2 | Swift | same file `citationToolDoesNotReachTheDepiction` *(new)* | freeform armed, likeness dialog opened → `viewerArmedRegionTool == nil`, `isDrawingRegion == false` | AC-1.3, D-2, D-5 |
| T-3 | Swift | same file `deselectingTheRectangleEndsDrawing` *(new)* | drawing, `viewerArmedRegionTool = nil` → `presented == .authoring`, `armedRegionTool == nil` | AC-1.5, D-2 |
| T-4 | Swift | `ArtifactRegionGeometryTests` `boundingRectangleOfEachKind` *(new)* | circle, L, freeform → box; rectangle → identical; full-square freeform → (0,0)–(1,1) | AC-2.1, AC-2.2, AC-2.3, D-4 |
| T-5 | Swift | `CitationComposerDepictionTests` `circledCitationRegionGivesItsBox` *(new)* | Citation circle region → `depictionCitationRegion` is its box; `useDepictionRegion` sets it | AC-2.1 |
| T-6 | Swift | same file `drawWaitsForACuttableFile` *(new; replaces `missingFileCutFails`)* | missing image file → `depictionCanCut == false`, `.cannotCut`, `useDepictionRegion(.whole)` false, `depictionPreviewKey == nil`, `authoringCanConfirm == false` with a region set; an unchanged edit there stays confirmable | AC-4.1, D-7, D-8, D-9 |
| T-7 | Swift | same file `failedCutDisablesSaveAndRetries` *(new)* | PDF opened, then its object deleted, region committed, preview rendered → `.failed`, `.cutFailed` (`isWarning`), `authoringCanConfirm == false`, Draw enabled; the same on a redrawn edit → `editCanConfirm == false`; PDF rewritten with `ArtifactFileFixtures.writeBlankPDF`, same region committed, preview rendered → `.cut`, confirmable | AC-4.3, AC-4.4, AC-4.5, D-9 |
| T-8 | Swift | same file `drawingLocksTheLocator` *(new)* | Citation tool armed → `isDrawingRegion`, the five locator calls refused; arming on a PDF with Find open closes Find; drawing a depiction → same refusals; nothing armed → `setPageFromViewer()` sets the page | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-5 |
| T-9 | Swift | same file `calloutSavesTheCitationsPage` *(new)* | PDF Citation with no page, viewer page 2 → true, stored locator page 2, needs-page clears, Draw enabled; `cit-1` removed from the fake → false, still needs page, `fields.error != nil`; image Citation → false, no `updateCitation` in `recordedCalls` | AC-5.1, AC-5.2, AC-5.3, D-10, D-11 |
| T-10 | Swift | same file `captionFollowsTheDraft` (extended) | empty and needs-page states read the new values | AC-6.1, AC-6.2, D-8 |
| T-11 | Swift | same file `previewLabelFollowsTheSave` *(new)* | unchanged edit → `.saved`; redrawn edit and new draft → not `.saved`; the AC-1.6 hexagon edit, not redrawn, saves with no depiction input | AC-6.4, AC-1.6, D-13 |

AC-4.2, AC-6.3 and AC-6.5 are manual checks (3, 6, 2). Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerDepictionTests`, `CitationAutoTranscribeTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left (`showsToolStrip`, `depictionCaption`, `CitationComposerEvidenceColumn.canSaveCitation`); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none in docs/*.md; add-depiction-value and docs/ideas/depictions-and-likenesses.md updated (STEP-9)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. New project, an image Source with a family photo, a Citation with a circle region round a face. Add Likeness for a Person: the strip shows one radio. Draw: rectangle armed, the region chip's × and the Clear menu disabled, no locator hint. Escape, then "Use the citation's region": the preview shows the circle's box. (Bar 4, bar 9.)
2. Close the dialog: seven radios. Arm freehand: chip × and Clear disabled, no hint. Place 256 points with VoiceOver on: the 257th and 258th clicks beep, and VoiceOver reads the cap once. (Bar 9.)
3. Open a Citation on a large scan and open a likeness dialog while the viewer is still loading: Draw disabled, the preview loading, Save disabled; they enable when the scan shows.
4. Open a Citation whose image file was removed from the project's objects folder: the viewer's empty state, the likeness row's buttons disabled, the cannot-cut caption.
5. A PDF Citation: draw a likeness, delete the PDF from the project's objects folder, then redraw: the warning caption, Save disabled; put the file back and press Draw again on the same region: the preview, Save enabled.
6. A PDF Citation saved without a page, viewer on page 2: the Callout and "Save page 2 to this citation"; press it: the page chip appears, the Callout goes, Draw enabled. A saved likeness whose region no longer matches the Citation's page shows the new region-hidden Callout. (Bar 9.)
7. Edit a saved likeness without redrawing: VoiceOver reads the preview as "Saved image". A saved polygon likeness (from an older build) opens and saves unchanged. On a PDF, Set page dims while drawing.
8. Widen the window past 1600pt of pane: frames 1 and 5 look the same, the canvas wider. Check light and dark.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Disabling the chip × on an armed Citation tool surprises a researcher who arms a tool by accident | Clicking the selected radio, or Escape, disarms (D-2); manual check 2 |
| `depictionCanCut` false while the viewer loads blocks Draw briefly | The preview shows loading (D-7); the block lifts on load |
| D-11 saves other unsaved Citation edits | User accepted, 2026-10-31 |
| `PVChip` for Set page looks different from today's hand-built button | The board (S14-D6 frame 1) shows it; Stop condition 1 |

**Stop and report if:**

- The agreed S14-D6 board differs from § 6 or a DF requirement (copy, which controls disable, Set page's look, where the Save page button sits, the drawing hint's wording).
- `ArtifactRegionCrop.snapshot` is not nil for a viewer showing an empty state, or is nil for a loaded image or PDF page (D-7 would block or allow the wrong files).
- `commitViewerRegion` no longer routes to the depiction draft whenever `currentDepictionDraft != nil` (D-1, D-2 key on that).
- A test in `CitationAutoTranscribeTests` expects the chip ×, Clear menu or Set page enabled with a tool armed.
- `fields.saveCitation()` cannot run while the observation dialog is open (D-11).
- `PVChip` cannot take `.disabled` and dim (D-6).

## 12. Open questions

None. The plan is blocked on the board, not on a question.

## 13. PR

**Title:** S14-03 — Rectangle-only depiction drawing
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-03-rectangle-only-depiction-drawing.md`.
